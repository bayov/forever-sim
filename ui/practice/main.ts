// The practice room page (docs/practice_room.md): pick the settings and the character, then play
// the fight in 3D.

import {
	barRows,
	baseKeyName,
	commandFor,
	commandForSlot,
	keysForSlot,
	MacroAction,
	MacroContext,
	macroIconSpell,
	Modifiers,
	mouseButtonName,
	parseMacro,
	runMacro,
	shortKey,
	slotForCommand,
	SlotContent,
	SpellNames,
	Toolkit,
} from './actions';
import { CameraOptions, Mover } from './controls';
import { iconName, iconUrl, setIcon } from './icons';
import { Batch, ClockState, FromWorker, Input, PlayerState, Settings, simTimeAt, Summary, TargetState, ToolkitSpell, ToWorker, wallNow } from './protocol';
import { Room } from './scene';

const app = document.getElementById('app')!;

type Loaded = { request: unknown; rotationSpells: number[] | null; class: string; race: string; level: number; name: string };
type RecordSummary = { guid: string; name: string; realm: string; class: string; race: string; level: number };
type SavedSettings = {
	version: 1;
	slots?: Record<string, SlotContent | null>;
	binds?: Record<string, string | null>;
	fight?: Partial<Settings>;
};
type CharacterData = {
	record?: { guid: string; name: string; slots?: Record<string, Record<string, unknown> | false> };
	bindings: Record<string, string>;
	cvars: Record<string, string>;
	saved?: SavedSettings;
};

const defaultSettings: Settings = { mode: 'free', lengthMs: 180000, countdownMs: 5000, pingMs: 0, executeFromStart: false, spellQueueWindowMs: 400 };

function el<K extends keyof HTMLElementTagNameMap>(tag: K, attrs: Record<string, string> = {}, ...children: (Node | string)[]): HTMLElementTagNameMap[K] {
	const e = document.createElement(tag);
	for (const [k, v] of Object.entries(attrs)) {
		if (k === 'class') e.className = v;
		else e.setAttribute(k, v);
	}
	for (const c of children) e.append(c);
	return e;
}

async function getJSON<T>(url: string, init?: RequestInit): Promise<T> {
	const r = await fetch(url, init);
	const json = await r.json();
	if (!r.ok || (json && json.error)) throw new Error(json?.error ?? r.statusText);
	return json as T;
}

// ---- Setup ----

async function showSetup(error?: string) {
	app.replaceChildren();
	const panel = el('div', { class: 'setup' });
	app.append(panel);
	panel.append(el('h1', {}, 'Practice Room'));
	if (error) panel.append(el('div', { class: 'setup-error' }, error));

	const presets = await getJSON<string[]>('/api/presets').catch(() => [] as string[]);
	const records = await getJSON<RecordSummary[]>('/api/records').catch(() => [] as RecordSummary[]);
	const last = JSON.parse(localStorage.getItem('practice.setup') ?? '{}');

	const presetSelect = el('select', { id: 'preset' });
	presetSelect.append(el('option', { value: '' }, 'Paste settings below'));
	for (const p of presets) presetSelect.append(el('option', { value: p }, p));
	presetSelect.value = last.preset ?? (presets.includes('shaman__enhancement-level60') ? 'shaman__enhancement-level60' : '');
	const textarea = el('textarea', { id: 'settings-text', placeholder: 'A class sim export link or JSON', rows: '4' });
	textarea.value = last.text ?? '';

	const recordSelect = el('select', { id: 'record' });
	recordSelect.append(el('option', { value: 'default' }, 'No record (default keys)'));
	for (const r of records) recordSelect.append(el('option', { value: r.guid }, `${r.name} · ${r.realm} · ${r.race} ${r.class} ${r.level}`));
	recordSelect.value = last.record ?? (records[0]?.guid || 'default');

	panel.append(
		el('label', {}, 'Build', presetSelect),
		el('label', {}, 'Or settings', textarea),
		el('label', {}, 'Keybinds and bars', recordSelect),
	);
	const start = el('button', { id: 'enter', class: 'primary' }, 'Enter room');
	panel.append(start);
	start.onclick = async () => {
		start.disabled = true;
		try {
			const text = presetSelect.value ? await (await fetch(`/api/presets/${presetSelect.value}`)).text() : textarea.value;
			localStorage.setItem('practice.setup', JSON.stringify({ preset: presetSelect.value, text: textarea.value, record: recordSelect.value }));
			const loaded = await getJSON<Loaded>('/api/request', { method: 'POST', body: text });
			const character = await getJSON<CharacterData>(`/api/character/${recordSelect.value}`);
			await enterRoom(loaded, recordSelect.value, character);
		} catch (e) {
			showSetup(String(e));
		}
	};
}

// ---- The room ----

const settingKey = 'practice.fight';

class PracticeRoom {
	readonly root = el('div', { class: 'room' });
	readonly canvas = el('canvas', { class: 'scene' });
	readonly hud = el('div', { class: 'hud' });
	room!: Room;
	worker: Worker;
	toolkit!: Toolkit;
	mover!: Mover;
	clock: ClockState = { wallAtStart: 0, startMs: 0, shiftMs: 0, pausedMs: 0 };
	player?: PlayerState;
	target?: TargetState;
	fight: Batch['fight'] = { mode: 'free', phase: 'prepull', lengthMs: 0 };
	hasTarget = true;
	mouseover: 'enemy' | 'none' = 'none';
	nextInputId = 1;
	slots = new Map<number, SlotContent>();
	bindings: Record<string, string>;
	settings: Settings;
	itemsByName = new Map<string, number>();
	running = false;
	editing = false;
	binding = false;
	cursor?: SlotContent;
	hoverSlot?: number;
	summary?: Summary;
	pulledAt?: number;
	dummyPos: [number, number] = [0, 0];
	private buttons = new Map<number, HTMLElement>();
	// Timings for spike 4: the WASM load, each call into Go, the swaps, and the time from a key
	// press to its answer on the page.
	readonly timings = { loadMs: 0, startCallMs: 0, calls: [] as number[], swaps: [] as number[], pressToAnswer: [] as number[] };
	private pressedAt = new Map<number, number>();
	// sessions counts the sessions started, for tests waiting on a restart.
	sessions = 0;
	private lastFrame = 0;
	private dragging: { button: number; x: number; y: number; moved: boolean; picked: 'enemy' | 'ground' | undefined } | undefined;

	constructor(
		readonly loaded: Loaded,
		readonly guid: string,
		readonly character: CharacterData,
	) {
		this.bindings = { ...character.bindings };
		for (const [key, command] of Object.entries(character.saved?.binds ?? {})) {
			if (command) this.bindings[key] = command;
			else delete this.bindings[key];
		}
		const stored = JSON.parse(localStorage.getItem(settingKey) ?? '{}');
		this.settings = { ...defaultSettings, ...stored, ...(character.saved?.fight ?? {}) };
		const queueWindow = parseFloat(character.cvars.SpellQueueWindow ?? '');
		if (queueWindow > 0) this.settings.spellQueueWindowMs = queueWindow;
		this.worker = new Worker(new URL('./worker.ts', import.meta.url), { type: 'module' });
		this.worker.onmessage = (e: MessageEvent<FromWorker>) => this.onWorker(e.data);
	}

	post(msg: ToWorker) {
		this.worker.postMessage(msg);
	}

	now(): number {
		return simTimeAt(this.clock, wallNow());
	}

	async mount() {
		app.replaceChildren(this.root);
		this.root.append(this.canvas, this.hud);
		this.room = new Room(this.canvas, 1.5, 2.0, 1.5, 2.3);
		const cvars = this.character.cvars;
		const options: CameraOptions = {
			yawSpeed: parseFloat(cvars.cameraYawMoveSpeed ?? '180'),
			pitchSpeed: parseFloat(cvars.cameraPitchMoveSpeed ?? '90'),
			smoothStyle: parseInt(cvars.cameraSmoothStyle ?? '4'),
			yawSmoothSpeed: parseFloat(cvars.cameraYawSmoothSpeed ?? '180'),
			maxDistance: 15 * Math.min(4, parseFloat(cvars.cameraDistanceMaxZoomFactor ?? '1')),
			zoomStep: 1,
		};
		this.mover = new Mover(
			[0, -15],
			Math.PI / 2,
			options,
			m => this.sendInput({ kind: 'move', pos: m.pos, velocity: m.velocity, facing: m.facing, turnRate: m.turnRate }),
			() => this.sendInput({ kind: 'jump' }),
		);
		this.buildHud();
		window.addEventListener('resize', () => this.resize());
		this.resize();
		this.room.setFov(parseFloat(cvars.cameraFov ?? '90'));
		this.bindEvents();
		this.start();
		requestAnimationFrame(t => this.frame(t));
	}

	resize() {
		this.room.resize();
		this.room.setFov(parseFloat(this.character.cvars.cameraFov ?? '90'));
	}

	start() {
		this.summary = undefined;
		this.pulledAt = undefined;
		this.hud.querySelector('.summary')?.remove();
		this.mover.move.pos = [0, -15];
		this.mover.move.facing = Math.PI / 2;
		this.mover.cam.yaw = Math.PI / 2;
		this.hasTarget = true;
		const seed = Math.floor(Math.random() * 2 ** 31) + 1;
		this.post({ type: 'start', request: this.loaded.request, settings: this.settings, seed, wallNow: wallNow() });
	}

	onWorker(msg: FromWorker) {
		switch (msg.type) {
			case 'ready':
				this.timings.loadMs = msg.loadMs;
				break;
			case 'started': {
				this.timings.startCallMs = msg.callMs;
				this.sessions++;
				const first = !this.toolkit;
				this.running = true;
				if (first) this.setupToolkit(msg.toolkit);
				break;
			}
			case 'batch':
				this.clock = msg.clock;
				this.timings.calls.push(msg.callMs);
				if (msg.batch.rebase) this.timings.swaps.push(msg.callMs);
				for (const a of msg.batch.answers ?? []) {
					const at = this.pressedAt.get(a.inputId);
					if (at !== undefined) {
						this.timings.pressToAnswer.push(wallNow() - at);
						this.pressedAt.delete(a.inputId);
					}
				}
				this.onBatch(msg.batch).catch(e => console.error(e));
				break;
			case 'attempt':
				if (msg.attempt) {
					fetch('/api/attempts', { method: 'POST', body: JSON.stringify(msg.attempt) })
						.then(r => r.json())
						.then(r => this.showMessage(`Saved attempt ${r.id}`));
				}
				break;
			case 'error':
				this.showError(msg.message.split('\n')[0]);
				console.error(msg.message);
				break;
		}
	}

	async setupToolkit(spells: ToolkitSpell[]) {
		const ids = spells.filter(s => s.spellId).map(s => s.spellId);
		const extra: number[] = [];
		// Spell slots the client knows by other ranks, and spells in the rotation, need names too.
		for (const v of Object.values(this.character.record?.slots ?? {})) {
			if (v && v.kind === 'spell' && typeof v.id === 'number') extra.push(v.id);
		}
		extra.push(...(this.loaded.rotationSpells ?? []));
		const names = await getJSON<SpellNames>(`/api/spells?ids=${[...ids, ...extra].join(',')}`).catch(() => ({}) as SpellNames);
		this.toolkit = new Toolkit(spells, names);
		for (const s of spells) {
			if (s.itemId) {
				iconName('item', s.itemId); // warm the cache
			}
		}
		this.loadSlots();
		this.renderBars();
		this.renderSpellbook();
	}

	loadSlots() {
		this.slots.clear();
		const record = this.character.record;
		if (record?.slots) {
			for (const [slot, v] of Object.entries(record.slots)) {
				if (!v) continue;
				const kind = v.kind as string;
				if (kind === 'spell' && typeof v.id === 'number') this.slots.set(+slot, { kind: 'spell', id: v.id, name: v.name as string });
				else if (kind === 'item' && typeof v.id === 'number') this.slots.set(+slot, { kind: 'item', id: v.id, name: v.name as string });
				else if (kind === 'macro') this.slots.set(+slot, { kind: 'macro', name: (v.name as string) ?? '', body: (v.body as string) ?? '', icon: v.icon as number });
			}
		} else {
			// No record: the spec's main spells from the rotation on bar 1 and the bars after it.
			let slot = 1;
			for (const id of this.loaded.rotationSpells ?? []) {
				const resolved = this.toolkit.resolveSpell(id);
				if (resolved === undefined) continue;
				this.slots.set(slot, { kind: 'spell', id: resolved });
				slot = slot === 12 ? 61 : slot + 1;
			}
			for (const s of this.toolkit.spells) {
				if (s.kind === 'item' && s.itemId) {
					this.slots.set(slot, { kind: 'item', id: s.itemId });
					slot = slot === 12 ? 61 : slot + 1;
				}
			}
		}
		for (const [slot, v] of Object.entries(this.character.saved?.slots ?? {})) {
			if (v) this.slots.set(+slot, v);
			else this.slots.delete(+slot);
		}
	}

	async onBatch(b: Batch) {
		if (b.player) this.player = b.player;
		if (b.target) this.target = b.target;
		this.fight = b.fight;
		for (const e of b.errors ?? []) this.showError(e);
		for (const h of b.hits ?? []) this.combatText(h.damage, h.outcome, h.periodic);
		if (b.pulled) {
			this.pulledAt = 0;
			if (b.pulled.early) this.showMessage('Early pull!');
		}
		if (b.units) {
			const enemy = b.units.find(u => u.ref.kind === 'enemy');
			if (enemy) this.dummyPos = enemy.pos;
		}
		if (this.player?.totems !== undefined) this.room.setTotems(this.player.totems ?? []);
		if (b.ended) {
			this.running = false;
			if (b.ended.summary) {
				this.summary = b.ended.summary;
				// Procs and other spells outside the toolkit need their names too.
				const missing = (b.ended.summary.spells ?? []).map(s => s.spellId).filter(id => id && !this.toolkit.names[id]);
				if (missing.length) Object.assign(this.toolkit.names, await getJSON<SpellNames>(`/api/spells?ids=${missing.join(',')}`).catch(() => ({})));
				this.showSummary(b.ended.summary, b.ended.stoppedEarly);
				this.post({ type: 'attempt' });
			} else {
				this.start();
			}
		}
	}

	// ---- inputs ----

	sendInput(input: Omit<Input, 'id'>, pressedAtWall = wallNow()): number {
		const id = this.nextInputId++;
		if (this.running) {
			if (input.kind !== 'move') this.pressedAt.set(id, pressedAtWall);
			this.post({ type: 'input', input: { ...input, id } as Input, pressedAtWall });
		}
		return id;
	}

	mods(e: { altKey: boolean; ctrlKey: boolean; shiftKey: boolean }): Modifiers {
		return { alt: e.altKey, ctrl: e.ctrlKey, shift: e.shiftKey };
	}

	bindEvents() {
		const wallOf = (e: Event) => performance.timeOrigin + e.timeStamp;
		window.addEventListener('keydown', e => {
			if ((e.target as HTMLElement)?.tagName === 'INPUT' || (e.target as HTMLElement)?.tagName === 'SELECT') return;
			const base = baseKeyName(e.code);
			if (!base) return;
			if (this.binding && this.hoverSlot !== undefined && !['SHIFT', 'CTRL', 'ALT'].includes(base)) {
				e.preventDefault();
				if (base === 'ESCAPE') {
					this.binding = false;
					this.renderBars();
					return;
				}
				this.bindKey(withModsName(base, this.mods(e)), this.hoverSlot);
				return;
			}
			if (base === 'ESCAPE') {
				e.preventDefault();
				this.escape(wallOf(e));
				return;
			}
			const command = commandFor(this.bindings, base, this.mods(e));
			if (!command) return;
			e.preventDefault();
			if (e.repeat) return;
			this.runCommand(command, true, e, wallOf(e));
		});
		window.addEventListener('keyup', e => {
			const base = baseKeyName(e.code);
			if (!base) return;
			// Releasing a key stops its movement whatever modifiers are held now.
			for (const mods of [this.mods(e), { alt: false, ctrl: false, shift: false }]) {
				const command = commandFor(this.bindings, base, mods);
				if (command) this.runCommand(command, false, e, wallOf(e));
			}
		});
		window.addEventListener('blur', () => {
			this.mover.held = { forward: false, back: false, strafeLeft: false, strafeRight: false, turnLeft: false, turnRight: false };
			this.mover.leftDown = this.mover.rightDown = false;
			this.post({ type: 'pause', wallNow: wallNow() });
		});
		window.addEventListener('focus', () => this.post({ type: 'resume', wallNow: wallNow() }));
		document.addEventListener('visibilitychange', () => this.post({ type: document.hidden ? 'pause' : 'resume', wallNow: wallNow() }));

		this.canvas.addEventListener('contextmenu', e => e.preventDefault());
		this.canvas.addEventListener('mousedown', e => {
			const extra = mouseButtonName(e.button);
			if (extra) {
				e.preventDefault();
				const command = commandFor(this.bindings, extra, this.mods(e));
				if (command) this.runCommand(command, true, e, wallOf(e));
				return;
			}
			const picked = this.pickAt(e.clientX, e.clientY);
			this.dragging = { button: e.button, x: e.clientX, y: e.clientY, moved: false, picked };
			if (e.button === 0) this.mover.leftDown = true;
			if (e.button === 2) {
				this.mover.rightDown = true;
				this.mover.move.facing = this.mover.cam.yaw;
			}
		});
		window.addEventListener('mousemove', e => {
			const d = this.dragging;
			if (d) {
				if (!d.moved && Math.hypot(e.clientX - d.x, e.clientY - d.y) > 3) {
					d.moved = true;
					this.canvas.requestPointerLock?.();
				}
				if (d.moved) this.mover.look(e.movementX, e.movementY);
			} else if (e.target === this.canvas) {
				this.mouseover = this.pickAt(e.clientX, e.clientY) === 'enemy' ? 'enemy' : 'none';
				this.canvas.style.cursor = this.mouseover === 'enemy' ? 'crosshair' : 'default';
			}
		});
		window.addEventListener('mouseup', e => {
			const extra = mouseButtonName(e.button);
			if (extra) {
				e.preventDefault();
				const command = commandFor(this.bindings, extra, this.mods(e));
				if (command) this.runCommand(command, false, e, wallOf(e));
				return;
			}
			if (e.button === 0) this.mover.leftDown = false;
			if (e.button === 2) this.mover.rightDown = false;
			const d = this.dragging;
			if (d && d.button === e.button) {
				this.dragging = undefined;
				if (document.pointerLockElement) document.exitPointerLock();
				if (!d.moved) this.click(d.button, d.picked, wallOf(e));
			}
		});
		this.canvas.addEventListener(
			'wheel',
			e => {
				e.preventDefault();
				const base = e.deltaY < 0 ? 'MOUSEWHEELUP' : 'MOUSEWHEELDOWN';
				const command = commandFor(this.bindings, base, this.mods(e));
				if (command && command !== 'CAMERAZOOMIN' && command !== 'CAMERAZOOMOUT') {
					this.runCommand(command, true, e, wallOf(e));
					this.runCommand(command, false, e, wallOf(e));
					return;
				}
				this.mover.zoom(e.deltaY < 0 ? -1 : 1);
			},
			{ passive: false },
		);
		// Mouse buttons 4 and 5 go back and forward in history otherwise.
		window.addEventListener('mouseup', e => {
			if (e.button === 3 || e.button === 4) e.preventDefault();
		});
	}

	pickAt(clientX: number, clientY: number) {
		const rect = this.canvas.getBoundingClientRect();
		return this.room.pick(((clientX - rect.left) / rect.width) * 2 - 1, -((clientY - rect.top) / rect.height) * 2 + 1);
	}

	// A click on the dummy targets it, a right click also attacks. A click on the ground keeps
	// your target (Sticky Targeting).
	click(button: number, picked: 'enemy' | 'ground' | undefined, wall: number) {
		if (picked !== 'enemy') return;
		this.hasTarget = true;
		this.sendInput({ kind: 'target', target: { kind: 'enemy', index: 0 } }, wall);
		if (button === 2) this.sendInput({ kind: 'startAttack' }, wall);
	}

	// Esc closes menus first, then stops your cast, and only then clears your target.
	escape(wall: number) {
		if (this.editing || this.binding) {
			this.editing = this.binding = false;
			this.cursor = undefined;
			this.renderBars();
			this.hud.querySelector('.spellbook')?.classList.add('hidden');
			return;
		}
		const book = this.hud.querySelector('.spellbook');
		if (book && !book.classList.contains('hidden')) {
			book.classList.add('hidden');
			return;
		}
		if (this.player?.cast && this.player.cast.end > this.now()) {
			this.sendInput({ kind: 'cancelCast' }, wall);
			return;
		}
		if (this.hasTarget) {
			this.hasTarget = false;
			this.sendInput({ kind: 'clearTarget' }, wall);
		}
	}

	runCommand(command: string, down: boolean, e: { altKey: boolean; ctrlKey: boolean; shiftKey: boolean }, wall: number) {
		const h = this.mover.held;
		switch (command) {
			case 'MOVEFORWARD':
				h.forward = down;
				if (down) this.mover.autorun = false;
				return;
			case 'MOVEBACKWARD':
				h.back = down;
				if (down) this.mover.autorun = false;
				return;
			case 'STRAFELEFT':
				h.strafeLeft = down;
				return;
			case 'STRAFERIGHT':
				h.strafeRight = down;
				return;
			case 'TURNLEFT':
				h.turnLeft = down;
				return;
			case 'TURNRIGHT':
				h.turnRight = down;
				return;
		}
		if (!down) return;
		switch (command) {
			case 'JUMP':
				this.mover.jump(wallNow());
				return;
			case 'TOGGLEAUTORUN':
				this.mover.autorun = !this.mover.autorun;
				return;
			case 'TARGETNEARESTENEMY':
			case 'TARGETPREVIOUSENEMY':
				this.hasTarget = true;
				this.sendInput({ kind: 'target', target: { kind: 'enemy', index: 0 } }, wall);
				return;
			case 'ATTACKTARGET':
				this.sendInput({ kind: 'startAttack' }, wall);
				return;
			case 'TOGGLESPELLBOOK':
				this.hud.querySelector('.spellbook')?.classList.toggle('hidden');
				return;
		}
		const slot = slotForCommand(command);
		if (slot !== undefined) this.pressSlot(slot, this.mods(e), wall);
	}

	macroContext(mods: Modifiers): MacroContext {
		return { mods, hasTarget: this.hasTarget, targetHarm: true, targetDead: false, mouseover: this.mouseover };
	}

	pressSlot(slot: number, mods: Modifiers, wall: number) {
		const content = this.slots.get(slot);
		if (!content || !this.toolkit) return;
		this.flashButton(slot);
		const actions: MacroAction[] = [];
		if (content.kind === 'spell') {
			const id = this.toolkit.resolveSpell(content.id);
			if (id !== undefined) actions.push({ kind: 'cast', spellId: id, target: 'target' });
			else this.showError('Unknown spell');
		} else if (content.kind === 'item') {
			if (this.toolkit.byItem.has(content.id)) actions.push({ kind: 'useItem', itemId: content.id });
			else this.showError("You don't have that item");
		} else {
			actions.push(...runMacro(parseMacro(content.body), this.macroContext(mods), this.toolkit, this.itemsByName));
		}
		for (const a of actions) {
			switch (a.kind) {
				case 'cast': {
					const spell = this.toolkit.bySpell.get(a.spellId);
					// @mouseover on the dummy targets the dummy for this cast only.
					const target = spell?.harmful ? { kind: 'enemy' as const, index: 0 } : { kind: 'player' as const, index: 0 };
					this.sendInput({ kind: 'cast', spellId: a.spellId, target }, wall);
					break;
				}
				case 'useItem':
					this.sendInput({ kind: 'useItem', itemId: a.itemId, target: { kind: 'player', index: 0 } }, wall);
					break;
				case 'startAttack':
					if (!this.hasTarget) this.showError('There is nothing to attack.');
					else this.sendInput({ kind: 'startAttack' }, wall);
					break;
				case 'stopAttack':
					this.sendInput({ kind: 'stopAttack' }, wall);
					break;
				case 'clearTarget':
					this.hasTarget = false;
					this.sendInput({ kind: 'clearTarget' }, wall);
					break;
			}
		}
	}

	// ---- bars and binds ----

	bindKey(key: string, slot: number) {
		const command = commandForSlot(slot);
		if (!command) return;
		const saved = this.saved();
		saved.binds = saved.binds ?? {};
		// A key binds one command: drop the slot's old keys and the key's old command.
		for (const k of keysForSlot(this.bindings, slot)) {
			delete this.bindings[k];
			saved.binds[k] = null;
		}
		this.bindings[key] = command;
		saved.binds[key] = command;
		this.save(saved);
		this.renderBars();
	}

	saved(): SavedSettings {
		this.character.saved = this.character.saved ?? { version: 1 };
		return this.character.saved;
	}

	save(saved: SavedSettings) {
		fetch(`/api/settings/${this.guid}`, { method: 'PUT', body: JSON.stringify(saved) });
	}

	setSlot(slot: number, content: SlotContent | undefined) {
		const saved = this.saved();
		saved.slots = saved.slots ?? {};
		if (content) this.slots.set(slot, content);
		else this.slots.delete(slot);
		saved.slots[String(slot)] = content ?? null;
		this.save(saved);
	}

	buildHud() {
		this.hud.innerHTML = `
			<div class="unit-frame player-frame"><div class="name"></div><div class="bar health"><span></span></div><div class="bar mana"><span></span><em></em></div></div>
			<div class="unit-frame target-frame"><div class="name"></div><div class="bar health"><span></span></div><div class="debuffs"></div></div>
			<div class="buffs"></div>
			<div class="errors"></div>
			<div class="messages"></div>
			<div class="fight-bar">
				<span class="phase"></span><span class="timer"></span>
				<button class="start-encounter">Start Encounter</button>
				<button class="stop">Stop</button>
				<button class="restart">Restart</button>
				<button class="leave">Leave</button>
				<button class="settings-toggle">Settings</button>
			</div>
			<div class="fight-settings hidden"></div>
			<div class="meter"><div class="dps"></div><div class="total"></div></div>
			<div class="countdown"></div>
			<div class="castbar hidden"><div class="fill"></div><div class="label"></div></div>
			<div class="swingbar hidden"><div class="fill"></div></div>
			<div class="bars"></div>
			<div class="side-bars"></div>
			<div class="bar-tools"><button class="edit-bars">Edit bars</button><button class="bind-keys">Bind keys</button><button class="book">Spellbook</button></div>
			<div class="spellbook hidden"></div>
			<div class="cursor-icon hidden"></div>
			<div class="combat-text"></div>`;
		const q = (s: string) => this.hud.querySelector(s) as HTMLElement;
		q('.player-frame .name').textContent = `${this.loaded.name || 'You'} · ${this.loaded.level}`;
		q('.start-encounter').onclick = () => this.post({ type: 'startEncounter', wallNow: wallNow() });
		q('.stop').onclick = () => this.post({ type: 'stop', wallNow: wallNow() });
		q('.restart').onclick = () => {
			this.post({ type: 'stop', wallNow: wallNow() });
			this.running = false;
			setTimeout(() => this.start(), 50);
		};
		q('.leave').onclick = () => {
			this.worker.terminate();
			showSetup();
		};
		q('.settings-toggle').onclick = () => q('.fight-settings').classList.toggle('hidden');
		q('.edit-bars').onclick = () => {
			this.editing = !this.editing;
			this.binding = false;
			this.cursor = undefined;
			q('.spellbook').classList.toggle('hidden', !this.editing);
			this.renderBars();
		};
		q('.bind-keys').onclick = () => {
			this.binding = !this.binding;
			this.editing = false;
			this.renderBars();
		};
		q('.book').onclick = () => q('.spellbook').classList.toggle('hidden');
		this.renderFightSettings();
		this.hud.addEventListener('mousemove', e => {
			const c = q('.cursor-icon');
			c.style.left = `${e.clientX + 8}px`;
			c.style.top = `${e.clientY + 8}px`;
		});
	}

	renderFightSettings() {
		const box = this.hud.querySelector('.fight-settings') as HTMLElement;
		box.replaceChildren();
		const s = this.settings;
		const mode = el('select', {}, el('option', { value: 'free' }, 'Free'), el('option', { value: 'encounter' }, 'Encounter'));
		mode.value = s.mode;
		const length = el('input', { type: 'number', min: '10', step: '10', value: String(s.lengthMs / 1000) });
		const countdown = el('input', { type: 'number', min: '0', step: '1', value: String(s.countdownMs / 1000) });
		const ping = el('input', { type: 'number', min: '0', step: '10', value: String(s.pingMs) });
		const execute = el('input', { type: 'checkbox' });
		execute.checked = s.executeFromStart;
		const apply = el('button', {}, 'Apply · Restart');
		apply.onclick = () => {
			this.settings = {
				...this.settings,
				mode: mode.value as Settings['mode'],
				lengthMs: Math.max(10, +length.value) * 1000,
				countdownMs: Math.max(0, +countdown.value) * 1000,
				pingMs: Math.max(0, +ping.value),
				executeFromStart: execute.checked,
			};
			localStorage.setItem(settingKey, JSON.stringify(this.settings));
			const saved = this.saved();
			saved.fight = { mode: this.settings.mode, lengthMs: this.settings.lengthMs, countdownMs: this.settings.countdownMs, pingMs: this.settings.pingMs, executeFromStart: this.settings.executeFromStart };
			this.save(saved);
			box.classList.add('hidden');
			this.post({ type: 'stop', wallNow: wallNow() });
			this.running = false;
			setTimeout(() => this.start(), 50);
		};
		box.append(
			el('label', {}, 'Mode', mode),
			el('label', {}, 'Length (s)', length),
			el('label', {}, 'Countdown (s)', countdown),
			el('label', {}, 'Ping (ms)', ping),
			el('label', { title: 'The dummy is in execute range from the start' }, 'Execute range', execute),
			apply,
		);
	}

	renderBars() {
		const bars = this.hud.querySelector('.bars') as HTMLElement;
		const side = this.hud.querySelector('.side-bars') as HTMLElement;
		bars.replaceChildren();
		side.replaceChildren();
		let shown = 0;
		this.buttons.clear();
		bars.classList.toggle('editing', this.editing);
		bars.classList.toggle('binding', this.binding);
		const rows = barRows.filter((row, i) => i < 5 || [...Array(12).keys()].some(k => this.slots.has(row.first + k)));
		for (const row of rows) {
			const used = [...Array(12).keys()].some(k => this.slots.has(row.first + k));
			if (!used && !this.editing && !this.binding) continue;
			const rowEl = el('div', { class: 'bar-row' });
			for (let k = 0; k < 12; k++) {
				const slot = row.first + k;
				const b = el('div', { class: 'button', 'data-slot': String(slot) });
				b.append(el('div', { class: 'icon' }), el('div', { class: 'sweep' }), el('div', { class: 'key' }), el('div', { class: 'name' }));
				const keys = keysForSlot(this.bindings, slot);
				(b.querySelector('.key') as HTMLElement).textContent = keys.length ? shortKey(keys[0]) : '';
				this.fillButton(b, this.slots.get(slot));
				b.onmouseenter = () => (this.hoverSlot = slot);
				b.onmouseleave = () => (this.hoverSlot = undefined);
				b.onmousedown = e => {
					e.stopPropagation();
					if (this.editing) {
						const old = this.slots.get(slot);
						this.setSlot(slot, this.cursor);
						this.cursor = old;
						this.renderBars();
						this.renderCursor();
						return;
					}
					if (!this.binding) this.pressSlot(slot, this.mods(e), performance.timeOrigin + e.timeStamp);
				};
				rowEl.append(b);
				this.buttons.set(slot, b);
			}
			// Three bars along the bottom, the rest as vertical bars on the right.
			if (shown++ < 3) bars.append(rowEl);
			else side.append(rowEl);
		}
	}

	renderCursor() {
		const c = this.hud.querySelector('.cursor-icon') as HTMLElement;
		c.classList.toggle('hidden', !this.cursor);
		if (this.cursor) this.fillButton(c, this.cursor);
	}

	fillButton(b: HTMLElement, content: SlotContent | undefined) {
		const icon = (b.querySelector('.icon') as HTMLElement) ?? b;
		const name = b.querySelector('.name') as HTMLElement | null;
		icon.style.backgroundImage = '';
		if (name) name.textContent = '';
		b.title = '';
		if (!content) return;
		if (content.kind === 'spell') {
			const id = this.toolkit?.resolveSpell(content.id);
			setIcon(icon, 'spell', content.id);
			b.title = this.toolkit?.nameOf(id ?? content.id) ?? '';
			b.classList.toggle('unknown', id === undefined);
		} else if (content.kind === 'item') {
			setIcon(icon, 'item', content.id);
			b.title = content.name ?? `Item ${content.id}`;
		} else {
			const macro = parseMacro(content.body);
			const shown = this.toolkit ? macroIconSpell(macro, this.toolkit, this.macroContext({ alt: false, ctrl: false, shift: false })) : undefined;
			if (shown) setIcon(icon, 'spell', shown);
			else icon.style.backgroundImage = `url(${iconUrl('inv_misc_questionmark')})`;
			if (name) name.textContent = content.name;
			b.title = `${content.name}\n${content.body}`;
			b.classList.toggle('unsupported', macro.unsupported.length > 0);
			b.dataset.spell = shown ? String(shown) : '';
		}
	}

	renderSpellbook() {
		const book = this.hud.querySelector('.spellbook') as HTMLElement;
		book.replaceChildren(el('div', { class: 'book-title' }, 'Spellbook · click to pick up, then click a slot'));
		const grid = el('div', { class: 'book-grid' });
		const entries = [...this.toolkit.spells].sort((a, b) => (a.kind === b.kind ? this.labelOf(a).localeCompare(this.labelOf(b)) : a.kind.localeCompare(b.kind)));
		for (const s of entries) {
			const content: SlotContent = s.itemId ? { kind: 'item', id: s.itemId } : { kind: 'spell', id: s.spellId! };
			const item = el('div', { class: 'book-entry' }, el('div', { class: 'icon' }), el('span', {}, this.labelOf(s)));
			this.fillButton(item, content);
			item.onmousedown = e => {
				e.stopPropagation();
				this.cursor = content;
				this.editing = true;
				this.renderBars();
				this.renderCursor();
			};
			grid.append(item);
		}
		book.append(grid);
	}

	labelOf(s: ToolkitSpell): string {
		if (s.itemId) return `Item ${s.itemId}`;
		const n = this.toolkit.names[s.spellId!];
		return n ? (n.rank ? `${n.name} (${n.rank})` : n.name) : `Spell ${s.spellId}`;
	}

	flashButton(slot: number) {
		const b = this.buttons.get(slot);
		if (!b) return;
		b.classList.add('pressed');
		setTimeout(() => b.classList.remove('pressed'), 120);
	}

	// ---- feedback ----

	showError(text: string) {
		const box = this.hud.querySelector('.errors') as HTMLElement;
		// Errors already on screen only flash (Forever's UIErrorsFrame rules).
		const existing = [...box.children].find(c => c.textContent === text) as HTMLElement | undefined;
		if (existing) {
			existing.classList.remove('flash');
			void existing.offsetWidth;
			existing.classList.add('flash');
			return;
		}
		const line = el('div', { class: 'error-line' }, text);
		box.append(line);
		while (box.children.length > 3) box.firstChild?.remove();
		setTimeout(() => line.classList.add('fade'), 2000);
		setTimeout(() => line.remove(), 2500);
	}

	showMessage(text: string) {
		const box = this.hud.querySelector('.messages') as HTMLElement;
		const line = el('div', {}, text);
		box.append(line);
		setTimeout(() => line.remove(), 3000);
	}

	combatText(damage: number, outcome: string, periodic?: boolean) {
		const box = this.hud.querySelector('.combat-text') as HTMLElement;
		const p = this.room.screenPoint(this.dummyPos, 2.6);
		if (!p.visible) return;
		const text = outcome === 'hit' || outcome === 'crit' || outcome === 'glance' || outcome === 'block' || outcome === 'resist' ? String(Math.round(damage)) : outcome[0].toUpperCase() + outcome.slice(1);
		const t = el('div', { class: `ct ${outcome}${periodic ? ' periodic' : ''}` }, text + (outcome === 'crit' ? '!' : ''));
		t.style.left = `${p.x + (Math.random() - 0.5) * 60}px`;
		t.style.top = `${p.y}px`;
		box.append(t);
		setTimeout(() => t.remove(), 1500);
	}

	showSummary(sum: Summary, stoppedEarly: boolean) {
		const box = el('div', { class: 'summary' });
		box.append(
			el('h2', {}, `${sum.dps.toFixed(1)} DPS`),
			el('div', { class: 'summary-line' }, `${Math.round(sum.damage)} damage · ${(sum.durationMs / 1000).toFixed(1)}s · ${sum.casts} casts${stoppedEarly ? ' · stopped early' : ''}`),
			el('div', { class: 'summary-line' }, `Idle GCD ${(sum.idleGcdMs / 1000).toFixed(1)}s · out of range ${(sum.outOfRangeMs / 1000).toFixed(1)}s`),
		);
		const table = el('table', {}, el('tr', {}, el('th', {}, 'Ability'), el('th', {}, 'Damage'), el('th', {}, '%'), el('th', {}, 'Casts'), el('th', {}, 'Hits'), el('th', {}, 'Crit'), el('th', {}, 'Miss')));
		for (const s of sum.spells ?? []) {
			const name = s.otherId ? (s.tag === 2 ? 'Melee (off hand)' : 'Melee') : s.spellId ? this.toolkit.nameOf(s.spellId) : `Item ${s.itemId}`;
			const icon = el('span', { class: 'mini-icon' });
			if (s.spellId) setIcon(icon, 'spell', s.spellId);
			else if (s.itemId) setIcon(icon, 'item', s.itemId);
			else icon.style.backgroundImage = `url(${iconUrl('inv_sword_04')})`;
			const hits = s.hits + s.misses;
			table.append(
				el(
					'tr',
					{},
					el('td', {}, icon, name),
					el('td', {}, String(Math.round(s.damage))),
					el('td', {}, sum.damage ? ((100 * s.damage) / sum.damage).toFixed(1) : '0'),
					el('td', {}, String(s.casts)),
					el('td', {}, String(hits)),
					el('td', {}, hits ? `${((100 * s.crits) / hits).toFixed(0)}%` : ''),
					el('td', {}, hits ? `${((100 * s.misses) / hits).toFixed(0)}%` : ''),
				),
			);
		}
		box.append(table);
		const auras = (sum.auras ?? []).slice(0, 12);
		if (auras.length) {
			const list = el('div', { class: 'uptimes' });
			for (const a of auras) {
				const icon = el('span', { class: 'mini-icon' });
				if (a.spellId) setIcon(icon, 'spell', a.spellId);
				else if (a.itemId) setIcon(icon, 'item', a.itemId);
				list.append(el('span', { class: 'uptime' }, icon, `${a.uptimePct.toFixed(0)}%`));
			}
			box.append(el('div', { class: 'summary-line' }, 'Uptimes'), list);
		}
		const again = el('button', { class: 'primary' }, 'Restart');
		again.onclick = () => this.start();
		box.append(again);
		this.hud.append(box);
	}

	// ---- per frame ----

	frame(t: number) {
		requestAnimationFrame(tt => this.frame(tt));
		const dt = this.lastFrame ? Math.min(0.1, (t - this.lastFrame) / 1000) : 0;
		this.lastFrame = t;
		const wall = wallNow();
		// Ghost Wolf runs 40% faster.
		this.mover.speedFactor = this.player?.auras?.some(a => a.spellId === 2645) ? 1.4 : 1;
		if (this.running) this.mover.step(dt, wall);
		const m = this.mover.move;
		const h = this.mover.height(wall);
		this.room.setUnit(this.room.player, m.pos, m.facing, h);
		this.room.setUnit(this.room.dummy, this.dummyPos, (3 * Math.PI) / 2);
		this.room.setTargeted(this.hasTarget);
		this.room.placeCamera(m.pos, h, this.mover.cam);
		this.room.render();
		this.updateHud();
	}

	updateHud() {
		const now = this.now();
		const q = (s: string) => this.hud.querySelector(s) as HTMLElement;
		const p = this.player;
		const pct = (a: number, b: number) => `${b > 0 ? Math.max(0, Math.min(100, (100 * a) / b)) : 0}%`;
		if (p) {
			(q('.player-frame .health span') as HTMLElement).style.width = pct(p.health, p.maxHealth);
			(q('.player-frame .mana span') as HTMLElement).style.width = pct(p.mana, p.maxMana);
			q('.player-frame .mana em').textContent = `${Math.round(p.mana)} / ${Math.round(p.maxMana)}`;
		}
		const tf = q('.target-frame');
		tf.classList.toggle('hidden', !this.hasTarget);
		if (this.target) {
			q('.target-frame .name').textContent = `${this.target.name} · ${this.target.level}`;
			this.renderAuras(q('.target-frame .debuffs'), this.target.debuffs ?? [], now);
		}
		if (p) this.renderAuras(q('.buffs'), (p.auras ?? []).filter(a => a.expiresAt !== undefined || (a.spellId ?? 0) > 0), now);

		// The cast bar and the swing timer.
		const cast = p?.cast;
		const castbar = q('.castbar');
		if (cast && cast.end > now) {
			castbar.classList.remove('hidden');
			(q('.castbar .fill') as HTMLElement).style.width = pct(now - cast.start, cast.end - cast.start);
			q('.castbar .label').textContent = `${this.toolkit?.nameOf(cast.spellId) ?? ''} ${Math.max(0, (cast.end - now) / 1000).toFixed(1)}`;
		} else {
			castbar.classList.add('hidden');
		}
		const swing = q('.swingbar');
		if (p?.attackOn && p.swingMs > 0) {
			swing.classList.remove('hidden');
			swing.classList.toggle('dim', !p.swinging);
			const next = p.nextSwingAt ?? now;
			(q('.swingbar .fill') as HTMLElement).style.width = pct(p.swingMs - (next - now), p.swingMs);
		} else {
			swing.classList.add('hidden');
		}

		// Buttons: GCD and cooldown sweeps, out of mana, out of range.
		if (p && this.toolkit) {
			const cds = new Map<string, { readyAt: number; length: number }>();
			for (const c of p.cooldowns ?? []) cds.set(c.spellId ? `s${c.spellId}` : `i${c.itemId}`, c);
			const dist = Math.hypot(this.mover.move.pos[0] - this.dummyPos[0], this.mover.move.pos[1] - this.dummyPos[1]);
			for (const [slot, b] of this.buttons) {
				const content = this.slots.get(slot);
				let spell: ToolkitSpell | undefined;
				let key = '';
				if (content?.kind === 'spell') {
					const id = this.toolkit.resolveSpell(content.id);
					spell = id !== undefined ? this.toolkit.bySpell.get(id) : undefined;
					key = `s${id}`;
				} else if (content?.kind === 'item') {
					spell = this.toolkit.byItem.get(content.id);
					key = `i${content.id}`;
				} else if (content?.kind === 'macro' && b.dataset.spell) {
					spell = this.toolkit.bySpell.get(+b.dataset.spell);
					key = `s${b.dataset.spell}`;
				}
				let frac = 0;
				const cd = cds.get(key);
				if (cd && cd.readyAt > now && cd.length > 0) frac = Math.max(frac, (cd.readyAt - now) / cd.length);
				if (spell && spell.gcdMs > 0 && p.gcdEnd > now && p.gcdEnd > p.gcdStart) frac = Math.max(frac, (p.gcdEnd - now) / (p.gcdEnd - p.gcdStart));
				(b.querySelector('.sweep') as HTMLElement).style.setProperty('--sweep', `${Math.min(1, frac) * 360}deg`);
				b.classList.toggle('oom', !!spell && spell.cost > p.mana);
				const range = spell?.melee ? 5 : spell?.maxRange ? spell.maxRange + 3 : 0;
				b.classList.toggle('oor', !!spell?.harmful && range > 0 && (dist > range || !this.hasTarget));
				b.classList.toggle('queued', !!spell && spell.spellId !== undefined && spell.spellId === p.queuedSpellId);
			}
		}

		// The fight bar, the countdown and the meter.
		const phase = this.fight.phase;
		q('.phase').textContent = { prepull: 'Prepull', countdown: 'Countdown', fight: this.fight.mode === 'encounter' ? 'Encounter' : 'Free', ended: 'Ended' }[phase];
		const shown = phase === 'ended' ? (this.fight.endedAtMs ?? 0) : now;
		q('.timer').textContent = phase === 'fight' || phase === 'ended' ? `${fmtTime(shown)}${this.fight.mode === 'encounter' ? ' / ' + fmtTime(this.fight.lengthMs) : ''}` : phase === 'countdown' ? fmtTime(now) : '';
		q('.start-encounter').classList.toggle('hidden', phase !== 'prepull');
		q('.stop').classList.toggle('hidden', phase !== 'fight');
		const countdown = q('.countdown');
		countdown.textContent = phase === 'countdown' && now < 0 ? String(Math.ceil(-now / 1000)) : '';
		const damage = p?.damageDone ?? 0;
		const elapsed = phase === 'fight' ? Math.max(now, 1) : phase === 'ended' ? Math.max(this.fight.endedAtMs ?? 1, 1) : 0;
		q('.meter .dps').textContent = elapsed ? `${((damage * 1000) / elapsed).toFixed(1)} DPS` : '';
		q('.meter .total').textContent = elapsed ? `${Math.round(damage)} damage` : '';
	}

	renderAuras(box: HTMLElement, auras: { spellId?: number; itemId?: number; label: string; stacks?: number; expiresAt?: number }[], now: number) {
		const key = auras.map(a => `${a.spellId}:${a.itemId}:${a.stacks}`).join('|');
		if (box.dataset.key !== key) {
			box.dataset.key = key;
			box.replaceChildren();
			for (const a of auras) {
				if (!a.spellId && !a.itemId) continue;
				const icon = el('div', { class: 'aura', title: a.label }, el('div', { class: 'icon' }), el('div', { class: 'time' }), el('div', { class: 'stacks' }, a.stacks && a.stacks > 1 ? String(a.stacks) : ''));
				setIcon(icon.querySelector('.icon') as HTMLElement, a.spellId ? 'spell' : 'item', (a.spellId || a.itemId)!);
				box.append(icon);
			}
		}
		const times = box.querySelectorAll('.time');
		let i = 0;
		for (const a of auras) {
			if (!a.spellId && !a.itemId) continue;
			const t = times[i++] as HTMLElement | undefined;
			if (t) t.textContent = a.expiresAt !== undefined ? fmtLeft(a.expiresAt - now) : '';
		}
	}
}

function withModsName(base: string, mods: Modifiers): string {
	let name = base;
	if (mods.shift) name = 'SHIFT-' + name;
	if (mods.ctrl) name = 'CTRL-' + name;
	if (mods.alt) name = 'ALT-' + name;
	return name;
}

function fmtTime(ms: number): string {
	const neg = ms < 0;
	const s = Math.abs(ms) / 1000;
	const m = Math.floor(s / 60);
	const r = Math.floor(s % 60);
	return `${neg ? '-' : ''}${m}:${String(r).padStart(2, '0')}`;
}

function fmtLeft(ms: number): string {
	if (ms <= 0) return '';
	const s = ms / 1000;
	if (s >= 3600) return `${Math.floor(s / 3600)}h`;
	if (s >= 60) return `${Math.floor(s / 60)}m`;
	return s < 10 ? s.toFixed(1) : String(Math.floor(s));
}

async function enterRoom(loaded: Loaded, guid: string, character: CharacterData) {
	const room = new PracticeRoom(loaded, guid, character);
	await room.mount();
	(window as unknown as { practice: PracticeRoom }).practice = room;
}

showSetup();
