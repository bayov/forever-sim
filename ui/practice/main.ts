// The practice room page (docs/practice_room.md): pick the settings and the character, then play
// the fight in 3D.

import * as THREE from 'three';
import {
	baseKeyName,
	commandFor,
	commandForSlot,
	keysForSlot,
	lockableCodes,
	MacroAction,
	matchNoteLabel,
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
import { iconName, iconUrl, meleeIcon, setIcon } from './icons';
import { css, HudLayout, pack, place, readLayout, Rect } from './layout';
import { AuraState, Batch, ClockState, FromWorker, HitEvent, Input, PlayerState, Settings, simTimeAt, Summary, TargetState, ToolkitSpell, ToWorker, wallNow } from './protocol';
import { Perf } from './perf';
import { Graphics, graphicsPresets, Room } from './scene';
import { loadCursors } from './cursors';
import { Pointer } from './pointer';
import { RaidTest, raidSpells } from './raid';
import { auraColor, avoidColor, CombatText, schoolColor } from './sct';
import { Tooltip, TooltipContent, tooltipInfo } from './tooltip';
import { Effects, MeleeSounds } from './wow/effects';
import { decodeHooks, fileProgress } from './wow/files';
import { find, setStyle, setText } from './dom';
import { Model } from './wow/model';
import { Terrain } from './wow/terrain';
import { SoundRef, Sounds } from './wow/sound';
import { Appearance, UnitView } from './wow/unit';

const app = document.getElementById('app')!;

type Loaded = {
	request: unknown;
	rotationSpells: number[] | null;
	class: string;
	race: string;
	level: number;
	name: string;
	trinkets?: number[];
	// The client's race and class ids and the gear by inventory slot, for the model.
	look?: { race: number; class: number; items: Record<string, number>; weaponSubclass?: number; weaponMaterial?: number };
};
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
	// The EllesmereUI profile and the class keybind notes (cmd/practice/ui.go).
	ui?: Record<string, unknown>;
	noteBinds?: Record<string, string>;
};

// The Bayov class pack's bars (bayovui/auras/forever/shaman/group.py): the main bar's
// cooldowns, the utility bar's, and what the maintain cells hold.
const packMain = ['Stormstrike', 'Earth Shock', 'Fire Nova', 'Chain Lightning', 'Lava Burst', 'Rage of the Farseer', "Nature's Swiftness", 'Blood Fury', 'Berserking'];
const packMainItems = ['Sapper', 'Grenade', 'Dynamite', 'Mana Potion', 'Rune'];
const packUtility = ['Grounding Totem', 'Earthbind Totem', 'Stoneclaw Totem', 'Mana Tide Totem', 'Totemic Projection', 'Purge'];
const packUtilityItems = ['Healthstone', 'Healing Potion', 'Bandage'];
const totemElements = ['earth', 'fire', 'water', 'air'];
// shownSpells are the spells the room shows that aren't in your toolkit: imbue procs, the totems'
// attacks (Searing bolts, Magma pulses, Fire Nova) and the totems' buffs.
const shownSpells = [
	439440, 29469, 8034, 8037, 10458, 16352, 16353, 408423, 408424, 408426, 408427, 408428, 3606, 6350, 6351, 6352, 10435, 10436, 8187, 10579, 10580, 10581,
	25362, 25360, 10494, 10610,
];
// Class colors for the unit frames and the meter.
const classColors: Record<string, string> = {
	ClassShaman: '#0070dd',
	ClassWarrior: '#c69b6d',
	ClassRogue: '#fff468',
	ClassPaladin: '#f48cba',
	ClassHunter: '#aad372',
	ClassMage: '#3fc7eb',
	ClassWarlock: '#8788ee',
	ClassPriest: '#ffffff',
	ClassDruid: '#ff7c0a',
};

const defaultSettings: Settings = { mode: 'free', lengthMs: 180000, countdownMs: 5000, pingMs: 60, executeFromStart: false, spellQueueWindowMs: 400 };

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
	// The real models, once they load. Until then the units are capsules.
	playerView?: UnitView;
	dummyView?: UnitView;
	// The cast whose precast shows, by its start: it goes off when it runs its full time.
	private castShown?: { spellId: number; start: number; end: number };
	private toolkitReady!: Promise<void>;
	// Every spell of yours the sim may show a hit of (procs and the like), beside the toolkit.
	private shownBySim: number[] = [];
	private toolkitResolve!: () => void;
	private loadingModels = false;
	private lastFacing?: number;
	private listenerPos = new THREE.Vector3();
	private vocal: Record<string, SoundRef[]> = {};
	private lastVocal = 0;
	private totemCastAt = new Map<number, number>();
	private totemPrecastAt?: { spellId: number; at: number; element: string };
	private totemsShown = new Map<string, number>();
	// While Ghost Wolf is up, playerView is the wolf and orcView your character.
	private wolfView?: UnitView;
	private orcView?: UnitView;
	sounds?: Sounds;
	effects?: Effects;
	private aurasShown = '';
	private buttons = new Map<number, HTMLElement>();
	layout: HudLayout;
	tooltip!: Tooltip;
	// u is screen pixels to a UI unit.
	u = 1;
	// Slots filled from the keybind notes, not the character's bars.
	noteSlots = new Set<number>();
	itemNames = new Map<number, string>();
	private packIcons: { el: HTMLElement; spell: ToolkitSpell }[] = [];
	private maintainCells = new Map<string, HTMLElement>();
	// The cast history timeline: our casts with when we made them, and where each sits.
	private castLog: { at: number; spellId?: number; itemId?: number; el: HTMLElement; row: number; shift: number }[] = [];
	private axisLabels = new Map<number, HTMLElement>();
	// Timings for spike 4: the WASM load, each call into Go, the swaps, and the time from a key
	// press to its answer on the page.
	readonly timings = { loadMs: 0, startCallMs: 0, calls: [] as number[], swaps: [] as number[], pressToAnswer: [] as number[] };
	private pressedAt = new Map<number, number>();
	// sessions counts the sessions started, for tests waiting on a restart.
	sessions = 0;
	private lastFrame = 0;
	// The session's updates wait here for half a ping, the time the server's reply takes to reach
	// the game (docs/practice_room.md, "Ping").
	private inbox: { at: number; msg: Extract<FromWorker, { type: 'batch' }> }[] = [];
	// Frame times of the last second, for the FPS readout.
	private frameTimes: number[] = [];
	private sct?: CombatText;
	perf?: Perf;
	graphics: Graphics = readGraphics();
	// Your auras in the last update, so the combat text shows the ones you gain.
	private aurasSeen?: Set<string>;
	// The buffs you start with that never run out (the sim's raid buffs and consumables). They get
	// their own rows, above the buffs of the fight.
	private settingAuras = new Set<string>();
	private statsShownAt = 0;
	private historyWidth = 0;
	private pointer?: Pointer;
	private raid?: RaidTest;
	private drainTimer?: ReturnType<typeof setInterval>;
	// Set when you leave the room, so nothing of it keeps running.
	private left = false;
	private dragging: { button: number; buttons: Set<number>; both?: boolean; x: number; y: number; moved: boolean; picked: 'enemy' | 'ground' | undefined } | undefined;

	constructor(
		readonly loaded: Loaded,
		readonly guid: string,
		readonly character: CharacterData,
	) {
		this.layout = readLayout(character.ui as Record<string, unknown> | undefined);
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

	// now is the sim time the page shows. It's half a ping behind the session, like the game shows
	// the server's state when its reply arrives.
	now(): number {
		return simTimeAt(this.clock, wallNow() - this.settings.pingMs / 2);
	}

	async mount() {
		app.replaceChildren(this.root);
		this.root.append(this.canvas, this.hud);
		this.room = new Room(this.canvas, 1.5, 2.0, 1.5, 2.3);
		this.room.setGraphics(this.graphics);
		this.perf = new Perf(this.room.renderer);
		this.pointer = new Pointer(this.root);
		// A model built mid fight (a spell's effect) can cost a frame.
		Model.onLoad = (fdid, ms) => !this.loadingModels && this.perf?.note('model', performance.now(), ms, String(fdid));
		decodeHooks.onDecode = (fdid, ms) => !this.loadingModels && this.perf?.note('texture', performance.now(), ms, String(fdid));
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
		this.toolkitReady = new Promise(r => (this.toolkitResolve = r));
		this.start();
		requestAnimationFrame(t => this.frame(t));
		await this.loadModels();
		// What you did behind the loading screen doesn't count: the session starts over.
		this.start();
	}

	// loadModels loads every model, texture, animation and sound the room uses behind a loading
	// screen, so nothing pops in mid fight. A model that fails to load leaves its capsule.
	async loadModels() {
		this.loadingModels = true;
		const screen = document.createElement('div');
		screen.className = 'loading';
		screen.innerHTML = '<div class="loading-box"><div class="loading-text">Loading</div><div class="loading-bar"><div></div></div></div>';
		this.root.append(screen);
		const text = screen.querySelector('.loading-text') as HTMLElement;
		const fill = screen.querySelector('.loading-bar > div') as HTMLElement;
		const startAsked = fileProgress.asked;
		const startDone = fileProgress.done;
		let shown = 0;
		const tick = () => {
			if (!screen.isConnected) return;
			const asked = fileProgress.asked - startAsked;
			const done = fileProgress.done - startDone;
			// The total grows as models name their textures, so the bar never runs backwards.
			shown = Math.max(shown, asked > 0 ? done / asked : 0);
			fill.style.width = `${(shown * 100).toFixed(1)}%`;
			text.textContent = `Loading · ${done}/${asked} files`;
			requestAnimationFrame(tick);
		};
		tick();

		Model.camera = this.room.camera;
		this.sounds = new Sounds(this.room.scene);
		this.sounds.self = this.room.player;
		const effects = new Effects(this.room.scene, this.sounds);
		this.effects = effects;
		const jobs: Promise<unknown>[] = [loadCursors(this.root)];
		// Models we draw once while loading, so their shaders and textures are ready for the fight.
		const warm: Model[] = [];
		const raidSize = Number(new URLSearchParams(location.search).get('raid') ?? 0);

		// The ground is a flat spot in the northern Barrens (Kalimdor's WDT), with trees around it and
		// no buildings near.
		jobs.push(
			Terrain.load({ wdt: 782779, x: 750, y: -2650, radius: 240, flat: 45, blend: 110, clear: 25 })
				.then(async t => {
					this.room.setTerrain(t.group);
					await t.loadDoodads();
				})
				.catch(e => console.warn('terrain', e)),
		);
		// The Barrens' daytime ambience and wind, which play from the start (SoundAmbience 3049).
		jobs.push(
			Promise.all([292584, 292700].map(id => getJSON<SoundRef[]>(`/api/sounds/kit/${id}`)))
				.then(async kits => {
					await Promise.all(kits.map(k => this.sounds!.preload(k)));
					this.sounds!.ambience(kits);
				})
				.catch(e => console.warn('ambience', e)),
		);

		// Totems show their element's model (CreatureDisplayInfo 4587 to 4590).
		const totemDisplays: Record<string, number> = { fire: 4589, earth: 4588, water: 4587, air: 4590 };
		const totemLooks = new Map<string, Promise<Appearance | undefined>>();
		for (const [element, display] of Object.entries(totemDisplays)) {
			const look = getJSON<Appearance>(`/api/appearance/creature/${display}`).catch(() => undefined);
			totemLooks.set(element, look);
			jobs.push(
				look.then(async a => {
					if (!a) return;
					const view = new UnitView(new THREE.Group(), []);
					await view.load(a).catch(() => undefined);
					const m = view.model as Model | undefined;
					if (!m) return;
					await m.loadEffectTextures();
					warm.push(m);
				}),
			);
		}
		this.room.totemModel = async element => {
			const a = await totemLooks.get(element);
			if (!a) return undefined;
			const view = new UnitView(new THREE.Group(), []);
			await view.load(a).catch(() => undefined);
			return view.model as Model | undefined;
		};

		const look = this.loaded.look;
		jobs.push(
			getJSON<MeleeSounds>(`/api/sounds/melee?weaponSubclass=${look?.weaponSubclass ?? -1}`)
				.then(async m => {
					effects.melee = m;
					await Promise.all([m.swing, m.swingCrit, m.impact, m.impactCrit].map(k => this.sounds!.preload(k)));
				})
				.catch(() => undefined),
		);
		if (look?.race) {
			const items = Object.entries(look.items)
				.map(([slot, id]) => `${slot}:${id}`)
				.join(',');
			// The client's files don't say a character's sex, so it's a choice in the room (0 male).
			const sex = Number(localStorage.getItem('practice.modelSex') ?? 0);
			// The raid test (?raid=40): others in your gear around a boss, to measure the drawing.
			if (raidSize > 1) {
				const raid = new RaidTest(this.room.scene, this.room.camera, effects);
				jobs.push(
					raid
						.load(raidSize, items, url => getJSON(url))
						.then(() => (this.raid = raid))
						.catch(e => console.warn('raid', e)),
				);
			}
			jobs.push(
				getJSON<Record<string, SoundRef[]>>(`/api/sounds/vocal?race=${look.race}&sex=${sex}`)
					.then(async v => {
						this.vocal = v;
						await Promise.all(Object.values(v).map(k => this.sounds!.preload(k)));
					})
					.catch(() => undefined),
			);
			// Drawing and sheathing the weapon: metal (1) or wood (2).
			const wood = look.weaponMaterial === 2;
			jobs.push(
				Promise.all([wood ? 699 : 700, wood ? 697 : 698].map(id => getJSON<SoundRef[]>(`/api/sounds/kit/${id}`)))
					.then(async ([draw, sheathe]) => {
						effects.draw = draw;
						effects.sheathe = sheathe;
						await Promise.all([this.sounds!.preload(draw), this.sounds!.preload(sheathe)]);
					})
					.catch(() => undefined),
			);
			jobs.push(
				getJSON<Appearance>(`/api/appearance/character?race=${look.race}&sex=${sex}&class=${look.class}&items=${items}`)
					.then(async a => {
						const view = new UnitView(this.room.player, this.room.placeholders(this.room.player), this.room.scene);
						await view.load(a);
						view.onSound = (kind, at) => effects.unitSound(kind, at);
						this.playerView = view;
						// Footsteps on plain ground, by the body's creature sounds.
						const us = a.creatureDisplayId
							? await getJSON<{ footsteps?: Record<string, SoundRef[]>; attack?: SoundRef[] }>(`/api/sounds/unit/${a.creatureDisplayId}`).catch(() => undefined)
							: undefined;
						effects.steps = us?.footsteps?.Grass ?? us?.footsteps?.Dirt ?? us?.footsteps?.Default;
						effects.grunts = us?.attack;
						await Promise.all([this.sounds!.preload(effects.steps), this.sounds!.preload(effects.grunts)]);
					})
					.catch(e => console.warn('player model', e)),
			);
		}
		// Ghost Wolf's wolf (SpellShapeshiftForm 16, creature display 4613).
		jobs.push(
			getJSON<Appearance>('/api/appearance/creature/4613')
				.then(async a => {
					const wolf = new UnitView(new THREE.Group(), [], this.room.scene);
					await wolf.load(a);
					wolf.onSound = (kind, at) => effects.unitSound(kind, at);
					this.wolfView = wolf;
				})
				.catch(e => console.warn('ghost wolf model', e)),
		);
		jobs.push(
			getJSON<Appearance>('/api/appearance/creature/dummy')
				.then(async a => {
					const view = new UnitView(this.room.dummy, this.room.placeholders(this.room.dummy), this.room.scene);
					await view.load(a);
					this.dummyView = view;
				})
				.catch(e => console.warn('dummy model', e)),
		);
		jobs.push(
			this.toolkitReady.then(() =>
				effects
					.preload([...this.toolkit.spells.filter(s => s.kind !== 'item').map(s => s.spellId ?? 0).filter(id => id > 0), ...shownSpells, ...this.shownBySim, ...(raidSize > 1 ? raidSpells : [])])
					.then(models => warm.push(...models))
					.catch(e => console.warn('effects', e)),
			),
		);
		await Promise.all(jobs);
		await this.room.warmUp(warm).catch(e => console.warn('warm up', e));
		screen.remove();
		this.loadingModels = false;
	}

	resize() {
		this.room.resize();
		this.room.setFov(parseFloat(this.character.cvars.cameraFov ?? '90'));
		this.layoutHud();
	}

	start() {
		this.summary = undefined;
		this.pulledAt = undefined;
		this.hud.querySelector('.summary')?.remove();
		this.clearHistory();
		this.sct?.clear();
		this.perf?.reset();
		this.aurasSeen = undefined;
		this.settingAuras.clear();
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
				// The last session's updates still on their way show first, as they did before.
				this.drainInbox(Infinity);
				this.timings.startCallMs = msg.callMs;
				this.sessions++;
				const first = !this.toolkit;
				this.running = true;
				if (first) {
					this.shownBySim = msg.shown ?? [];
					this.setupToolkit(msg.toolkit);
				}
				break;
			}
			case 'batch':
				this.inbox.push({ at: wallNow() + this.settings.pingMs / 2, msg });
				this.drainInbox();
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

	// drainInbox shows the session's updates whose half ping has passed, in order.
	drainInbox(wall = wallNow()) {
		while (this.inbox.length && this.inbox[0].at <= wall) {
			const msg = this.inbox.shift()!.msg;
			this.clock = msg.clock;
			this.timings.calls.push(msg.callMs);
			if (msg.callMs > 3) this.perf?.note('sim', performance.now(), msg.callMs, `${msg.batch.hits?.length ?? 0} hits`);
			if (msg.batch.rebase) this.timings.swaps.push(msg.callMs);
			for (const a of msg.batch.answers ?? []) {
				const at = this.pressedAt.get(a.inputId);
				if (at !== undefined) {
					this.timings.pressToAnswer.push(wallNow() - at);
					this.pressedAt.delete(a.inputId);
				}
			}
			this.onBatch(msg.batch).catch(e => console.error(e));
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
		this.toolkitResolve();
		// Item names, for macros' /use lines, the keybind notes and the spellbook.
		await Promise.all(
			spells
				.filter(s => s.itemId)
				.map(async s => {
					iconName('item', s.itemId!); // warm the cache
					const info = await tooltipInfo('item', s.itemId!);
					if (info.name) {
						this.itemNames.set(s.itemId!, info.name);
						this.itemsByName.set(info.name.toLowerCase(), s.itemId!);
					}
				}),
		);
		this.loadSlots();
		this.renderBars();
		this.renderSpellbook();
		this.renderPack();
	}

	loadSlots() {
		this.slots.clear();
		this.noteSlots.clear();
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
		this.placeNoteBinds();
	}

	// placeNoteBinds puts toolkit spells the bars don't have on the slot of the key the class
	// keybind notes plan for them, when that slot is empty.
	placeNoteBinds() {
		const notes = this.character.noteBinds;
		if (!notes) return;
		const onBars = new Set<string>();
		for (const content of this.slots.values()) {
			const key = this.contentKey(content);
			if (key) onBars.add(key);
		}
		for (const [key, label] of Object.entries(notes)) {
			const content = matchNoteLabel(label, this.toolkit, this.itemNames, this.loaded.trinkets ?? []);
			if (!content) continue;
			const ck = this.contentKey(content);
			if (!ck || onBars.has(ck)) continue;
			const command = this.bindings[key];
			const slot = command ? slotForCommand(command) : undefined;
			if (slot === undefined || this.slots.has(slot)) continue;
			this.slots.set(slot, content);
			this.noteSlots.add(slot);
			onBars.add(ck);
		}
	}

	// contentKey names what a slot casts, to tell whether a spell is on the bars already.
	contentKey(content: SlotContent): string | undefined {
		if (content.kind === 'spell') {
			const id = this.toolkit.resolveSpell(content.id);
			return id !== undefined ? `s${id}` : undefined;
		}
		if (content.kind === 'item') return `i${content.id}`;
		const shown = macroIconSpell(parseMacro(content.body), this.toolkit, this.macroContext({ alt: false, ctrl: false, shift: false }));
		return shown !== undefined ? `s${shown}` : undefined;
	}

	async onBatch(b: Batch) {
		if (b.rebase) {
			// The session moved its clock (the pull is 0): the history's casts move with it.
			for (const c of this.castLog) c.at += b.rebase.shiftMs;
			for (const l of this.axisLabels.values()) l.remove();
			this.axisLabels.clear();
		}
		if (b.player) {
			this.showGains(b.player.auras ?? []);
			this.player = b.player;
		}
		if (b.target) this.target = b.target;
		this.fight = b.fight;
		for (const e of b.errors ?? []) this.showError(e);
		this.combatText(b.hits ?? []);
		for (const h of b.hits ?? []) this.animateHit(h);
		for (const c of b.casts ?? []) {
			this.addHistory(c.at, c.spellId, c.itemId);
			this.animateCast(c.spellId);
			// Fire Nova goes off around the fire totem. Its hits in this update already showed it.
			if (c.totem && c.visualSpellId && this.now() - (this.totemCastAt.get(c.visualSpellId) ?? -1e9) > 400) {
				this.totemCastAt.set(c.visualSpellId, this.now());
				this.effects?.totemCast(c.visualSpellId, this.room.totemModels.get(c.totem), undefined, () => undefined);
			}
		}
		if (b.pulled) {
			this.pulledAt = 0;
			if (b.pulled.early) this.showMessage('Early pull!');
		}
		if (b.units) {
			const enemy = b.units.find(u => u.ref.kind === 'enemy');
			if (enemy) this.dummyPos = enemy.pos;
		}
		// The session leaves the totems out when there are none (Totemic Recall, or they ran out).
		if (b.player) {
			this.room.setTotems(b.player.totems ?? []);
			this.showTotemBuffs(b.player.totems ?? []);
		}
		this.showAuras();
		if (b.ended) {
			this.running = false;
			if (b.ended.summary) {
				this.summary = b.ended.summary;
				// Procs and other spells outside the toolkit need their names too.
				const missing = (b.ended.summary.spells ?? []).map(s => s.spellId).filter(id => id && !this.toolkit.names[id]);
				if (missing.length) Object.assign(this.toolkit.names, await getJSON<SpellNames>(`/api/spells?ids=${missing.join(',')}`).catch(() => ({})));
				this.showSummary(b.ended.summary, b.ended.stoppedEarly);
				this.savePerf(b.ended.summary);
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
		// The session keeps running in the background, like the game does. You stop moving, because
		// the keys you held won't tell us when they come up.
		const letGo = () => {
			this.mover.held = { forward: false, back: false, strafeLeft: false, strafeRight: false, turnLeft: false, turnRight: false };
			this.mover.leftDown = this.mover.rightDown = false;
			if (this.running && !this.loadingModels) this.mover.step(0, wallNow());
		};
		window.addEventListener('blur', letGo);
		document.addEventListener('visibilitychange', () => document.hidden && letGo());
		// A hidden tab gets no frames, and its timers run at most once a second, so the session's
		// updates still get shown (Searing Totem's bolts, the fight's end).
		this.drainTimer = setInterval(() => this.drainInbox(), 100);

		this.canvas.addEventListener('contextmenu', e => e.preventDefault());
		this.canvas.addEventListener('mousedown', e => {
			const extra = mouseButtonName(e.button);
			if (extra) {
				e.preventDefault();
				const command = commandFor(this.bindings, extra, this.mods(e));
				if (command) this.runCommand(command, true, e, wallOf(e));
				return;
			}
			// Both buttons held are one drag: the second button joins the first one's drag, so the
			// mouse keeps steering while both run you forward.
			if (this.dragging) {
				this.dragging.buttons.add(e.button);
				this.dragging.both = true;
			}
			else this.dragging = { button: e.button, buttons: new Set([e.button]), x: e.clientX, y: e.clientY, moved: false, picked: this.pickAt(e.clientX, e.clientY) };
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
					if (this.pointer?.active) this.pointer.freeze();
					else this.canvas.requestPointerLock?.();
				}
				if (d.moved) this.mover.look(e.movementX, e.movementY);
			} else if (e.target === this.canvas) {
				this.mouseover = this.pickAt(e.clientX, e.clientY) === 'enemy' ? 'enemy' : 'none';
				this.canvas.classList.toggle('attack', this.mouseover === 'enemy');
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
			if (d && d.buttons.delete(e.button) && d.buttons.size === 0) {
				this.dragging = undefined;
				if (this.pointer?.active) {
					if (d.moved) this.pointer.thaw(d.x, d.y);
				} else if (document.pointerLockElement) document.exitPointerLock();
				// A click is a press and release of one button that didn't move.
				if (!d.moved && e.button === d.button && !d.both) this.click(d.button, d.picked, wallOf(e));
			}
		});
		// The wheel zooms over the HUD too, like the game's, but not over a panel you read or scroll.
		this.root.addEventListener(
			'wheel',
			e => {
				if ((e.target as Element).closest?.('.spellbook, .fight-settings, .summary, .v-select')) return;
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
			<div class="uf player-frame"><div class="hp"><span class="fill"></span><em class="right"></em></div></div>
			<div class="uf target-frame"><div class="hp"><span class="fill"></span><em class="left"></em><em class="right name"></em></div><div class="debuffs"></div></div>
			<div class="uf tot-frame"><div class="hp"><span class="fill"></span><em class="center"></em></div></div>
			<div class="pack"></div>
			<div class="castbar hidden"><div class="cb-icon"></div><div class="cb-bar"><div class="fill"></div><div class="latency"></div><span class="cb-name"></span><span class="cb-time"></span></div></div>
			<div class="buffs"><div class="buff-row fixed"></div><div class="buff-row timed"></div></div>
			<div class="nameplate hidden"><div class="np-debuffs"></div><div class="np-name"></div><div class="np-hp"><span class="fill"></span></div><div class="np-arrow left"></div><div class="np-arrow right"></div></div>
			<div class="dmeter"><div class="dm-header"></div><div class="dm-bar hidden"><span class="fill"></span><em class="l"></em><em class="r"></em></div></div>
			<div class="history"><div class="hist-now"></div><div class="hist-casts"></div><div class="hist-axis"></div></div>
			<div class="action-bars"></div>
			<div class="errors"></div>
			<div class="messages"></div>
			<div class="fight-bar">
				<span class="phase"></span><span class="timer"></span><span class="net"></span>
				<button class="start-encounter">Start Encounter</button>
				<button class="stop">Stop</button>
				<button class="restart">Restart</button>
				<button class="settings-toggle">Settings</button>
				<button class="fullscreen">Full screen</button>
				<button class="leave">Leave</button>
			</div>
			<div class="fight-settings hidden"></div>
			<div class="countdown"></div>
			<div class="bar-tools"><button class="edit-bars">Edit bars</button><button class="bind-keys">Bind keys</button><button class="book">Spellbook</button></div>
			<div class="spellbook hidden"></div>
			<div class="cursor-icon hidden"></div>
			<div class="combat-text"></div>`;
		const q = (s: string) => this.hud.querySelector(s) as HTMLElement;
		const L = this.layout;
		this.tooltip = new Tooltip(this.hud, () => ({ left: innerWidth / 2 + L.tooltip.x * this.u, top: innerHeight / 2 - L.tooltip.y * this.u }));
		const tip = (sel: string, content: () => TooltipContent | undefined) => this.tooltip.attach(q(sel), content);
		q('.player-frame .hp .fill').style.background = css(L.playerColor);
		q('.target-frame .hp .fill').style.background = css(L.targetColor);
		q('.tot-frame .hp .fill').style.background = classColors[this.loaded.class] ?? '#0070dd';
		q('.cb-bar .fill').style.background = css(L.castColor);
		q('.cb-bar .latency').style.background = css(L.castLatency, L.castLatency.a);
		q('.dm-bar .fill').style.background = classColors[this.loaded.class] ?? '#0070dd';
		q('.tot-frame .center').textContent = this.playerName();
		tip('.player-frame', () => ({ kind: 'lines', title: this.playerName(), lines: [`Level ${this.loaded.level} ${this.loaded.race.replace(/^Race/, '')} ${this.loaded.class.replace(/^Class/, '')}`] }));
		tip('.target-frame', () => this.targetTooltip());
		tip('.nameplate', () => this.targetTooltip());
		tip('.tot-frame', () => ({ kind: 'lines', title: this.playerName(), lines: ["Your target's target"] }));
		tip('.dmeter', () => ({ kind: 'lines', title: 'Damage Done', lines: ['Damage | DPS since the pull'] }));
		tip('.start-encounter', () => ({ kind: 'lines', title: 'Start Encounter', lines: [`${fmtTime(this.settings.lengthMs)} fight after a ${this.settings.countdownMs / 1000}s countdown`] }));
		tip('.stop', () => ({ kind: 'lines', title: 'Stop', lines: ['End the fight and show the summary'] }));
		tip('.restart', () => ({ kind: 'lines', title: 'Restart', lines: ['A new fight from the start spot'] }));
		tip('.net', () => ({ kind: 'lines', title: 'Latency · FPS', lines: ['Simulated ping, set in Settings'] }));
		tip('.fullscreen', () => ({ kind: 'lines', title: 'Full screen', lines: ['Ctrl+T, Ctrl+W and Esc reach your binds'] }));
		tip('.leave', () => ({ kind: 'lines', title: 'Leave', lines: ['Back to the build and character pick'] }));
		tip('.settings-toggle', () => ({ kind: 'lines', title: 'Settings', lines: ['Graphics, fight mode, length and ping'] }));
		tip('.edit-bars', () => ({ kind: 'lines', title: 'Edit bars', lines: ['Pick up a spell, then click a slot'] }));
		tip('.bind-keys', () => ({ kind: 'lines', title: 'Bind keys', lines: ['Hover a button and press a key'] }));
		tip('.book', () => ({ kind: 'lines', title: 'Spellbook', lines: ['Every spell and item you can use'] }));
		q('.start-encounter').onclick = () => this.post({ type: 'startEncounter', wallNow: wallNow() });
		q('.stop').onclick = () => this.post({ type: 'stop', wallNow: wallNow() });
		q('.restart').onclick = () => {
			this.post({ type: 'stop', wallNow: wallNow() });
			this.running = false;
			setTimeout(() => this.start(), 50);
		};
		q('.leave').onclick = () => {
			this.leave();
			showSetup();
		};
		q('.settings-toggle').onclick = () => q('.fight-settings').classList.toggle('hidden');
		q('.fullscreen').onclick = () => this.toggleFullscreen();
		document.addEventListener('fullscreenchange', () => {
			q('.fullscreen').textContent = document.fullscreenElement ? 'Exit full screen' : 'Full screen';
		});
		this.sct = new CombatText(q('.combat-text'));
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

	// toggleFullscreen goes full screen and locks your bound keys and Esc. Only then does Chrome
	// hand the page its own shortcuts, like Ctrl+T (new tab) and Ctrl+W (close tab). Hold Esc to
	// leave full screen.
	//
	// We don't lock every key, because then Chrome would also catch Alt+Tab and the Win key, and
	// you couldn't get back to WoW or Discord (docs/practice_room.md, "Keys").
	async toggleFullscreen() {
		const kb = (navigator as Navigator & { keyboard?: { lock(codes?: string[]): Promise<void>; unlock(): void } }).keyboard;
		if (document.fullscreenElement) {
			kb?.unlock();
			await document.exitFullscreen();
			return;
		}
		await document.documentElement.requestFullscreen();
		const bound = new Set(Object.keys(this.bindings).map(k => k.replace(/^(ALT-|CTRL-|SHIFT-)+/, '')));
		const codes = lockableCodes().filter(code => code === 'Escape' || bound.has(baseKeyName(code) ?? ''));
		await kb?.lock(codes).catch(e => console.warn('keyboard lock', e));
	}

	// leave closes the room. Its frames, timers, sounds, session and GPU memory all stop with it.
	leave() {
		this.left = true;
		this.running = false;
		clearInterval(this.drainTimer);
		this.worker.terminate();
		this.pointer?.dispose();
		if (document.pointerLockElement) document.exitPointerLock();
		void this.sounds?.listener.context.close().catch(() => {});
		this.room.renderer.dispose();
	}

	// savePerf sends the fight's performance report to the practice server
	// (~/forever-practice/perf/), and starts a new one.
	savePerf(summary: Summary) {
		const perf = this.perf;
		if (!perf) return;
		const report = perf.report({ fight: { durationMs: summary.durationMs, dps: summary.dps }, settings: this.settings, graphics: this.graphics, raid: this.raid?.count, timings: { swaps: this.timings.swaps } });
		perf.reset();
		fetch('/api/perf', { method: 'POST', body: JSON.stringify(report) })
			.then(r => r.json())
			.then(r => r.id && this.showMessage(`Saved performance ${r.id}`))
			.catch(e => console.warn('perf', e));
	}

	// drawStats shows the simulated latency and the frame rate, a few times a second.
	drawStats(t: number) {
		if (t - this.statsShownAt < 250) return;
		this.statsShownAt = t;
		const fps = this.frameTimes.length;
		(this.hud.querySelector('.net') as HTMLElement).textContent = `${this.settings.pingMs}ms · ${fps} FPS`;
	}

	playerName(): string {
		return this.loaded.name || this.character.record?.name || 'You';
	}

	targetTooltip(): TooltipContent | undefined {
		if (!this.target) return undefined;
		return { kind: 'lines', title: this.target.name, titleClass: 'hostile', lines: [`Level ${this.target.level}`, `${Math.round(this.target.damageTaken)} damage taken`] };
	}

	// layoutHud places the frames for the window's size, like the game scales its UI.
	layoutHud() {
		const W = innerWidth;
		const H = innerHeight;
		const L = this.layout;
		this.u = H / L.uiHeight;
		const u = this.u;
		this.hud.style.setProperty('--u', `${u}px`);
		const q = (s: string) => this.hud.querySelector(s) as HTMLElement;
		place(q('.player-frame'), L.player, u, W, H);
		place(q('.target-frame'), L.target, u, W, H);
		place(q('.tot-frame'), L.targetOfTarget, u, W, H);
		place(q('.dmeter'), L.meter, u, W, H);
		const cb = L.castbar;
		place(q('.castbar'), cb, u, W, H);
		q('.player-frame em').style.fontSize = `${L.playerTextSize * u}px`;
		for (const e of this.hud.querySelectorAll<HTMLElement>('.target-frame em')) e.style.fontSize = `${L.targetTextSize * u}px`;
		q('.cb-name').style.fontSize = `${L.castTextSize * u}px`;
		q('.cb-time').style.fontSize = `${L.castTimerSize * u}px`;
		q('.dm-bar').style.height = `${L.meterBarHeight * u}px`;
		this.renderBars();
		this.renderPack();
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
		// Graphics apply at once. A preset sets the others, and changing one makes it Custom.
		const g = this.graphics;
		const preset = el('select', {}, el('option', { value: 'low' }, 'Low'), el('option', { value: 'medium' }, 'Medium'), el('option', { value: 'high' }, 'High'), el('option', { value: 'custom' }, 'Custom'));
		preset.value = g.preset;
		const scale = el('select', {}, ...[0.5, 0.75, 1].map(v => el('option', { value: String(v) }, `${v * 100}%`)));
		scale.value = String(g.renderScale);
		const shadows = el('select', {}, el('option', { value: 'off' }, 'Off'), el('option', { value: 'low' }, 'Low'), el('option', { value: 'medium' }, 'Medium'), el('option', { value: 'high' }, 'High'));
		shadows.value = g.shadows;
		// The screen's own height in pixels (the first choice), and the common heights under it.
		const nativeH = Math.round(screen.height * devicePixelRatio);
		const heights = [2160, 1440, 1080, 720].filter(v => v < nativeH);
		const resolution = el('select', {}, el('option', { value: '0' }, `${nativeH}p`), ...heights.map(v => el('option', { value: String(v) }, `${v}p`)));
		const resolutionValue = (h: number) => (heights.includes(h) ? String(h) : '0');
		resolution.value = resolutionValue(g.maxHeight);
		const setGraphics = (next: Graphics) => {
			this.graphics = next;
			try {
				localStorage.setItem(graphicsKey, JSON.stringify(next));
			} catch {
				// Only this visit keeps it.
			}
			this.room.setGraphics(next);
			preset.value = next.preset;
			scale.value = String(next.renderScale);
			shadows.value = next.shadows;
			resolution.value = resolutionValue(next.maxHeight);
		};
		preset.onchange = () => preset.value !== 'custom' && setGraphics({ ...graphicsPresets[preset.value as 'low' | 'medium' | 'high'], maxHeight: this.graphics.maxHeight });
		resolution.onchange = () => setGraphics({ ...this.graphics, maxHeight: +resolution.value });
		scale.onchange = () => setGraphics({ ...this.graphics, preset: 'custom', renderScale: +scale.value });
		shadows.onchange = () => setGraphics({ ...this.graphics, preset: 'custom', shadows: shadows.value as Graphics['shadows'] });
		box.append(
			el('label', { class: 'resolution-label' }, 'Resolution', resolution),
			el('label', {}, 'Graphics', preset),
			el('label', { class: 'scale-label' }, 'Render scale', scale),
			el('label', {}, 'Shadows', shadows),
			el('div', { class: 'settings-gap' }),
			el('div', { class: 'settings-gap' }),
			el('label', {}, 'Mode', mode),
			el('label', {}, 'Length (s)', length),
			el('label', {}, 'Countdown (s)', countdown),
			el('label', {}, 'Ping (ms)', ping),
			el('label', { class: 'execute-label' }, 'Execute range', execute),
			apply,
		);
		this.tooltip.attach(box.querySelector('.resolution-label') as HTMLElement, () => ({ kind: 'lines', title: 'Resolution', lines: ['Draws at this height, stretched to fit the screen', 'Lower is faster on a 4K screen'] }));
		this.tooltip.attach(box.querySelector('.scale-label') as HTMLElement, () => ({ kind: 'lines', title: 'Render scale', lines: ['Lower draws fewer pixels, for more FPS'] }));
		this.tooltip.attach(box.querySelector('.execute-label') as HTMLElement, () => ({ kind: 'lines', title: 'Execute range', lines: ['The dummy is under 20% health from the start'] }));
	}

	// renderBars draws the action bars where EllesmereUI has them, each a block of buttons.
	renderBars() {
		const box = this.hud.querySelector('.action-bars') as HTMLElement;
		box.replaceChildren();
		this.buttons.clear();
		box.classList.toggle('editing', this.editing);
		box.classList.toggle('binding', this.binding);
		const u = this.u;
		for (const bar of this.layout.bars) {
			const count = bar.cols * bar.rows;
			const used = [...Array(count).keys()].some(k => this.slots.has(bar.first + k));
			const showAll = bar.showEmpty || this.editing || this.binding;
			if (!used && !showAll) continue;
			const block = el('div', { class: 'bar-block', 'data-bar': bar.key });
			place(block, bar.rect, u, innerWidth, innerHeight);
			for (let k = 0; k < count; k++) {
				const slot = bar.first + k;
				if (!showAll && !this.slots.has(slot)) continue;
				const b = el('div', { class: 'button', 'data-slot': String(slot) });
				b.append(el('div', { class: 'icon' }), el('div', { class: 'sweep' }), el('div', { class: 'cd' }), el('div', { class: 'key' }), el('div', { class: 'name' }));
				const col = k % bar.cols;
				const row = Math.floor(k / bar.cols);
				b.style.left = `${Math.round(col * (bar.size + bar.padding) * u)}px`;
				b.style.top = `${Math.round(row * (bar.size + bar.padding) * u)}px`;
				b.style.width = b.style.height = `${Math.round(bar.size * u)}px`;
				const key = b.querySelector('.key') as HTMLElement;
				const keys = keysForSlot(this.bindings, slot);
				key.textContent = keys.length ? shortKey(keys[0]) : '';
				key.style.fontSize = `${bar.keyFont * u}px`;
				const name = b.querySelector('.name') as HTMLElement;
				name.style.fontSize = `${bar.macroFont * u}px`;
				name.style.color = css(bar.macroColor);
				this.fillButton(b, this.slots.get(slot));
				b.classList.toggle('from-notes', this.noteSlots.has(slot));
				b.onmouseenter = () => {
					this.hoverSlot = slot;
					const c = this.slotTooltip(slot);
					if (c) this.tooltip.show(c);
				};
				b.onmouseleave = () => {
					this.hoverSlot = undefined;
					this.tooltip.hide();
				};
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
				block.append(b);
				this.buttons.set(slot, b);
			}
			box.append(block);
		}
	}

	// slotTooltip is the game tooltip of what a slot holds, with its key under it.
	slotTooltip(slot: number): TooltipContent | undefined {
		const content = this.slots.get(slot);
		const keys = keysForSlot(this.bindings, slot);
		const extra: string[] = [];
		if (this.binding) extra.push(keys.length ? `Key: ${keys.join(', ')} · press a key to bind` : 'Press a key to bind');
		if (this.noteSlots.has(slot)) extra.push('From your keybind notes');
		if (!content) return this.binding ? { kind: 'lines', title: 'Empty slot', lines: extra } : undefined;
		if (content.kind === 'spell') {
			const id = this.toolkit?.resolveSpell(content.id);
			if (id === undefined) extra.push('Not in this build');
			return { kind: 'spell', id: id ?? content.id, extra };
		}
		if (content.kind === 'item') return { kind: 'item', id: content.id, extra };
		const shown = this.toolkit ? macroIconSpell(parseMacro(content.body), this.toolkit, this.macroContext({ alt: false, ctrl: false, shift: false })) : undefined;
		const macro = parseMacro(content.body);
		if (macro.unsupported.length) extra.push(`Not supported here: ${macro.unsupported.join(' ')}`);
		if (shown) return { kind: 'spell', id: shown, extra: [`Macro: ${content.name}`, ...extra] };
		return { kind: 'lines', title: content.name || 'Macro', lines: [...content.body.split('\n'), ...extra] };
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
		if (!content) return;
		if (content.kind === 'spell') {
			const id = this.toolkit?.resolveSpell(content.id);
			setIcon(icon, 'spell', content.id);
			b.classList.toggle('unknown', id === undefined);
		} else if (content.kind === 'item') {
			setIcon(icon, 'item', content.id);
		} else {
			const macro = parseMacro(content.body);
			const shown = this.toolkit ? macroIconSpell(macro, this.toolkit, this.macroContext({ alt: false, ctrl: false, shift: false })) : undefined;
			if (shown) setIcon(icon, 'spell', shown);
			else icon.style.backgroundImage = `url(${iconUrl('inv_misc_questionmark')})`;
			if (name) name.textContent = content.name;
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
			this.tooltip.attach(item, () => (s.itemId ? { kind: 'item', id: s.itemId } : { kind: 'spell', id: s.spellId! }));
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
		if (s.itemId) return this.itemNames.get(s.itemId) ?? `Item ${s.itemId}`;
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
		// Your character says it too ("Not enough mana!"), at most once a second.
		const voice = this.vocal[text];
		if (voice && this.sounds && performance.now() - this.lastVocal > 1000) {
			this.lastVocal = performance.now();
			this.sounds.play(voice, undefined, 'vocal');
		}
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

	// combatText shows what we did on the scrolling combat text. Hits of one spell at the same
	// moment (a Fire Nova on several enemies) show as one line, like MSBT merges them.
	combatText(hits: HitEvent[]) {
		const groups = new Map<string, HitEvent[]>();
		for (const h of hits) {
			if (h.source !== 'player' && h.source !== 'totem') continue;
			const key = `${h.at} ${h.spellId ?? ''} ${h.itemId ?? ''} ${h.otherId ?? ''} ${h.tag ?? ''}`;
			const g = groups.get(key);
			if (g) g.push(h);
			else groups.set(key, [h]);
		}
		for (const g of groups.values()) {
			const h = g[0];
			const icon = (el: HTMLElement) => {
				if (h.spellId) setIcon(el, 'spell', h.visualSpellId ?? h.spellId);
				else if (h.itemId) setIcon(el, 'item', h.itemId);
				else el.style.backgroundImage = `url(${iconUrl(meleeIcon)})`;
			};
			const landed = g.filter(x => x.outcome !== 'miss' && x.outcome !== 'dodge' && x.outcome !== 'parry');
			const count = g.length > 1 ? `×${g.length}` : '';
			if (!landed.length) {
				const word = h.outcome[0].toUpperCase() + h.outcome.slice(1);
				this.sct?.add({ side: 'out', kind: 'avoid', text: word, note: count, color: avoidColor, icon });
				continue;
			}
			const amount = Math.round(landed.reduce((a, x) => a + x.damage, 0));
			const crit = landed.some(x => x.outcome === 'crit');
			const notes = [count];
			if (landed.some(x => x.outcome === 'glance')) notes.push('glancing');
			if (landed.some(x => x.outcome === 'block')) notes.push('blocked');
			if (landed.some(x => x.outcome === 'resist')) notes.push('resisted');
			this.sct?.add({
				side: 'out',
				kind: crit ? 'crit' : h.periodic ? 'tick' : 'damage',
				text: crit ? `${amount}!` : String(amount),
				note: notes.filter(Boolean).join(', '),
				color: schoolColor(h.school),
				icon,
			});
		}
	}

	// showGains shows the buffs you gain on the combat text ("+Flurry"). The first update of a
	// session only learns what you have.
	showGains(auras: AuraState[]) {
		const key = auraKey;
		const now = new Set(auras.map(key));
		const seen = this.aurasSeen;
		this.aurasSeen = now;
		if (!seen) {
			for (const a of auras) if (a.expiresAt === undefined) this.settingAuras.add(key(a));
			return;
		}
		for (const a of auras) {
			// The same auras the buff frame shows.
			if (seen.has(key(a)) || !a.label || (a.expiresAt === undefined && !((a.spellId ?? 0) > 0))) continue;
			this.sct?.add({
				side: 'in',
				kind: 'aura',
				text: `+${a.label}`,
				color: auraColor,
				icon: el => {
					if (a.spellId) setIcon(el, 'spell', a.spellId);
					else if (a.itemId) setIcon(el, 'item', a.itemId);
				},
			});
		}
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
		if (this.left) return;
		requestAnimationFrame(tt => this.frame(tt));
		const perf = this.loadingModels ? undefined : this.perf;
		perf?.begin(t);
		// The frames' rects for the combat text, read before this frame changes the page, so the
		// browser doesn't lay it out twice.
		const rect = (s: string) => find(this.hud, s).getBoundingClientRect();
		const sctRects = this.sct && { zone: rect('.combat-text'), player: rect('.player-frame'), target: find(this.hud, '.target-frame').classList.contains('hidden') ? undefined : rect('.target-frame'), history: rect('.history').bottom, buffs: rect('.buffs').bottom };
		// The cast history's width, read here for the same reason.
		this.historyWidth = find(this.hud, '.history').clientWidth;
		this.drainInbox();
		perf?.part('inbox');
		this.frameTimes.push(t);
		while (this.frameTimes[0] < t - 1000) this.frameTimes.shift();
		const dt = this.lastFrame ? Math.min(0.1, (t - this.lastFrame) / 1000) : 0;
		this.lastFrame = t;
		const wall = wallNow();
		// Ghost Wolf runs 40% faster.
		this.mover.speedFactor = this.player?.auras?.some(a => a.spellId === 2645) ? 1.4 : 1;
		if (this.running && !this.loadingModels) this.mover.step(dt, wall);
		const m = this.mover.move;
		const h = this.mover.height(wall);
		this.room.setUnit(this.room.player, m.pos, m.facing, h);
		this.room.setUnit(this.room.dummy, this.dummyPos, (3 * Math.PI) / 2);
		this.room.setTargeted(this.hasTarget);
		this.mover.updateZoom(dt);
		this.room.placeCamera(m.pos, h, this.mover.cam);
		if (this.sounds) {
			this.room.player.getWorldPosition(this.listenerPos);
			this.listenerPos.y += 2;
			this.sounds.place(this.listenerPos, this.room.camera);
		}
		perf?.part('move');
		this.animateUnits(dt, wall);
		perf?.part('units');
		perf?.gpuBegin();
		this.room.render();
		perf?.gpuEnd();
		perf?.part('render');
		this.updateHud();
		this.drawStats(t);
		perf?.part('hud');
		if (this.sct && sctRects) this.sct.draw(this.u, sctRects.zone, sctRects.player, sctRects.target, sctRects.history, sctRects.buffs);
		perf?.part('sct');
		perf?.end();
	}

	// ---- the models' animations ----

	animateUnits(dt: number, wall: number) {
		const m = this.mover.move;
		// Velocity relative to where you face: forward and to the left.
		const cos = Math.cos(m.facing);
		const sin = Math.sin(m.facing);
		const forward = m.velocity[0] * cos + m.velocity[1] * sin;
		const left = -m.velocity[0] * sin + m.velocity[1] * cos;
		const cast = this.player?.cast;
		const now = this.now();
		const shown = this.castShown;
		if (shown && (!cast || cast.spellId !== shown.spellId || cast.start !== shown.start)) {
			this.castShown = undefined;
			// A cast that ran its full time goes off. One that stopped early was cancelled.
			if (now >= shown.end - 150) this.releaseCast(shown.spellId);
			else {
				this.playerView?.castStop();
				this.effects?.castStop();
			}
		}
		if (cast && !this.castShown && cast.end > now) {
			this.castShown = { spellId: cast.spellId, start: cast.start, end: cast.end };
			const view = this.playerView;
			view?.castStart();
			this.effects?.castStart(cast.spellId, view).then(anim => {
				if (anim && this.castShown?.start === cast.start) view?.castStart(anim);
			});
		} else if (cast && this.castShown) {
			// Pushback moves the end.
			this.castShown.end = cast.end;
		}
		const tp = this.totemPrecastAt;
		if (tp && now >= tp.at) {
			this.totemPrecastAt = undefined;
			this.effects?.totemPrecast(tp.spellId, this.room.totemModels.get(tp.element));
		}
		this.raid?.update(dt);
		this.effects?.update(dt);
		this.room.animate(dt);
		// The facing's turn rate, for the shuffle when turning in place.
		const turn = this.lastFacing === undefined || dt <= 0 ? 0 : Math.atan2(Math.sin(m.facing - this.lastFacing), Math.cos(m.facing - this.lastFacing)) / dt;
		this.lastFacing = m.facing;
		this.playerView?.update(dt, { forward, left, turn, airborne: this.mover.airborne(wall), walking: false }, !!this.player?.attackOn);
		this.dummyView?.update(dt, { forward: 0, left: 0, airborne: false, walking: false }, false);
	}

	animateHit(h: HitEvent) {
		if (h.source === 'totem') {
			this.animateTotemHit(h);
			return;
		}
		if (h.source !== 'player') return;
		const moving = Math.hypot(...this.mover.move.velocity) > 0.1;
		const landed = h.outcome !== 'miss' && h.outcome !== 'dodge' && h.outcome !== 'parry';
		const crit = h.outcome === 'crit';
		const white = h.otherId === 7;
		const spell = h.spellId ? this.toolkit?.spells.find(s => s.spellId === h.spellId) : undefined;
		// What happens when the blow lands: the impact sound and effect and the dummy's flinch.
		const connect = () => {
			// A melee ability hits with the weapon too (Stormstrike's two blows).
			if (white || spell?.melee) this.effects?.impact(this.dummyView?.model?.root, crit, landed);
			if (white) {
				// Nothing more: the white swing has no spell effect.
			} else if (h.spellId && !h.periodic && landed) this.effects?.spellImpact(h.visualSpellId ?? h.spellId, this.dummyView, this.playerView);
			if (landed && !h.periodic) this.dummyView?.wound(crit);
		};
		const view = this.playerView;
		if (view && landed && (crit || spell?.melee) && !h.periodic) this.effects?.grunt(view.model?.root);
		// A melee blow lands when the swing gets there (the animation's $HIT), not when it starts.
		// White swings are "other action 7", tag 1 for the main hand and 2 for the off hand.
		if (white && view) {
			view.swing(h.tag === 2, moving);
			view.onHit(connect);
		} else if (spell?.melee && !h.periodic && view) view.onHit(connect);
		else connect();
	}

	// animateTotemHit shows a totem's attack: the Searing Totem's bolt flying from the totem, the
	// Magma Totem's pulse and Fire Nova's ring around the totem, and the hit on the dummy.
	animateTotemHit(h: HitEvent) {
		const spellId = h.visualSpellId ?? h.spellId;
		if (!spellId || !this.effects) return;
		const totem = this.room.totemModels.get(h.totem ?? 'fire');
		const landed = h.outcome !== 'miss' && h.outcome !== 'dodge' && h.outcome !== 'parry' && h.outcome !== 'resist';
		const land = () => {
			if (!landed) return;
			this.effects?.spellImpact(spellId, this.dummyView, undefined, totem);
			if (!h.periodic) this.dummyView?.wound(h.outcome === 'crit');
		};
		// One cast on the totem per pulse, however many targets it hits.
		const now = this.now();
		if (now - (this.totemCastAt.get(spellId) ?? -1e9) > 400) {
			this.totemCastAt.set(spellId, now);
			this.effects.totemCast(spellId, totem, this.dummyView?.model, land);
		} else land();
		// The Searing Totem casts its next bolt for 2.2s before it lands.
		if (h.nextAt) this.totemPrecastAt = { spellId, at: h.nextAt - 2200, element: h.totem ?? 'fire' };
	}

	// showTotemBuffs plays a new totem's buff on you (the glow at your feet), by the buff spell's
	// visual. The sim's totem auras carry the summon spell, so we pick the buff by the totem's name.
	showTotemBuffs(totems: { element: string; spellId: number }[]) {
		const buffs: Record<string, number> = {
			'Strength of Earth Totem': 25362,
			'Stoneskin Totem': 25362,
			'Grace of Air Totem': 25360,
			'Healing Stream Totem': 25362,
			'Mana Spring Totem': 10494,
			'Windfury Totem': 10610,
		};
		const seen = new Set<string>();
		for (const t of totems) {
			seen.add(t.element);
			if (this.totemsShown.get(t.element) === t.spellId) continue;
			this.totemsShown.set(t.element, t.spellId);
			const buff = buffs[this.toolkit?.names[t.spellId]?.name ?? ''];
			if (buff) this.effects?.buffBurst(buff, this.playerView);
		}
		for (const e of [...this.totemsShown.keys()]) if (!seen.has(e)) this.totemsShown.delete(e);
	}

	// showAuras keeps your auras' effects and your imbue's glow in step with the session.
	showAuras() {
		const p = this.player;
		if (!p || !this.effects) return;
		const ids = (p.auras ?? []).map(a => a.spellId ?? 0).filter(id => id > 0);
		this.setGhostWolf(ids.includes(2645));
		const key = ids.join(',');
		if (key !== this.aurasShown) {
			this.aurasShown = key;
			this.effects.setAuras(ids, this.playerView);
		}
		this.effects.setImbue('main', p.mainHandImbue, this.playerView?.mainHand());
		this.effects.setImbue('off', p.offHandImbue, this.playerView?.offHand());
	}

	// animateCast plays a spell going off. A spell with a cast time reports at its start, so it
	// goes off later, when its cast runs out (animateUnits).
	// setGhostWolf swaps your character for the wolf while Ghost Wolf is up.
	setGhostWolf(on: boolean) {
		const wolf = this.wolfView;
		if (!wolf || on === !!this.orcView) return;
		if (on) {
			this.orcView = this.playerView;
			if (this.orcView?.model) this.orcView.model.root.visible = false;
			this.room.player.add(wolf.group);
			this.playerView = wolf;
		} else {
			wolf.group.removeFromParent();
			this.playerView = this.orcView;
			if (this.playerView?.model) this.playerView.model.root.visible = true;
			this.orcView = undefined;
		}
	}

	animateCast(spellId?: number) {
		if (!spellId) return;
		const cast = this.player?.cast;
		if (cast && cast.spellId === spellId && cast.end > this.now()) return;
		this.releaseCast(spellId);
	}

	releaseCast(spellId: number) {
		const view = this.playerView;
		if (!view) return;
		const spell = this.toolkit?.spells.find(s => s.spellId === spellId);
		if (!spell || spell.kind === 'item') return;
		const moving = Math.hypot(...this.mover.move.velocity) > 0.1;
		this.effects?.cast(spellId, view, this.dummyView).then(v => {
			// A spell that does no damage (a shield, an imbue, a heal) has no hit to play its
			// impact on, so it plays on you as it goes off.
			if (v && !spell.harmful) this.effects?.spellImpact(spellId, view, view);
			const anim = v?.anims.cast ?? undefined;
			if (spell.melee) view.swing(false, moving, true, anim);
			else view.castEnd(anim, moving);
		});
	}

	// ---- the class pack ----

	// renderPack draws the Bayov class pack around the resource bars: maintain cells over them,
	// the main and utility cooldown bars under them.
	renderPack() {
		const box = this.hud.querySelector('.pack') as HTMLElement;
		box.replaceChildren();
		this.packIcons = [];
		this.maintainCells.clear();
		const u = this.u;
		const W = innerWidth;
		const H = innerHeight;
		const L = this.layout;
		const width = pack.width;
		const put = (e: HTMLElement, r: Rect) => {
			place(e, r, u, W, H);
			box.append(e);
			return e;
		};
		const bottom = L.player.y - L.player.h / 2;
		const power = put(el('div', { class: 'pk-power' }, el('span', { class: 'fill' }), el('em')), { x: 0, y: bottom, w: width, h: pack.powerH, point: 'BOTTOM' });
		(power.querySelector('.fill') as HTMLElement).style.background = css(L.manaColor);
		(power.querySelector('em') as HTMLElement).style.fontSize = `${11 * u}px`;
		this.tooltip.attach(power, () => ({ kind: 'lines', title: 'Mana', lines: this.player ? [`${Math.round(this.player.mana)} / ${Math.round(this.player.maxMana)}`] : [] }));
		let top = bottom + pack.powerH + pack.gap;
		const swing = put(el('div', { class: 'pk-swing' }, el('span', { class: 'fill' })), { x: 0, y: top, w: width, h: pack.swingH, point: 'BOTTOM' });
		this.tooltip.attach(swing, () => ({ kind: 'lines', title: 'Swing timer', lines: ['Red while auto attack is on but not swinging'] }));
		top += pack.swingH;
		if (this.loaded.class === 'ClassShaman') {
			top += pack.gap;
			const pips = put(el('div', { class: 'pk-pips' }), { x: 0, y: top, w: width, h: pack.pipH, point: 'BOTTOM' });
			for (let i = 0; i < 5; i++) pips.append(el('span'));
			this.tooltip.attach(pips, () => ({ kind: 'lines', title: 'Maelstrom Weapon', lines: ['At 5, your next Lightning Bolt is instant'] }));
			top += pack.pipH;
		}
		if (!this.toolkit) return;

		// The maintain cells: the shield in cell 1, the four totems on cells 3 to 6 shifted half a
		// cell right, and Flame Shock in cell 9.
		const step = pack.iconW + pack.gap;
		const left = -width / 2 + pack.iconW / 2;
		const cellY = top + pack.gap + pack.iconH / 2;
		const cell = (name: string, x: number) => {
			const c = put(this.packIcon('maintain'), { x, y: cellY, w: pack.iconW, h: pack.iconH, point: 'CENTER' });
			c.classList.add('hidden');
			this.maintainCells.set(name, c);
			this.tooltip.attach(c, () => {
				const id = c.dataset.spell ? +c.dataset.spell : undefined;
				const expiresAt = c.dataset.expires ? +c.dataset.expires : undefined;
				return id ? { kind: 'aura', spellId: id, label: name, remaining: () => (expiresAt !== undefined ? fmtRemaining(expiresAt - this.now()) : '') } : undefined;
			});
		};
		cell('shield', left);
		totemElements.forEach((e, i) => cell(e, left + (2 + i + 0.5) * step));
		cell('dot', left + 8 * step);

		// The main bar and the utility bar, nine to a row, centred.
		const pick = (spellNames: string[], itemNames: string[]) => {
			const out: ToolkitSpell[] = [];
			for (const name of spellNames) {
				const id = this.toolkit.spellByName(name);
				const spell = id !== undefined ? this.toolkit.bySpell.get(id) : undefined;
				if (spell && spell.cdMs > 0) out.push(spell);
			}
			for (const part of itemNames) {
				for (const [id, name] of this.itemNames) {
					const spell = this.toolkit.byItem.get(id);
					if (spell && name.includes(part) && !out.includes(spell)) out.push(spell);
				}
			}
			return out;
		};
		const row = (spells: ToolkitSpell[], centreY: number) => {
			spells.forEach((spell, i) => {
				const inRow = Math.min(pack.perRow, spells.length - Math.floor(i / pack.perRow) * pack.perRow);
				const x = (-(inRow - 1) / 2 + (i % pack.perRow)) * step;
				const y = centreY - Math.floor(i / pack.perRow) * (pack.iconH + pack.gap);
				const icon = put(this.packIcon('cooldown'), { x, y, w: pack.iconW, h: pack.iconH, point: 'CENTER' });
				setIcon(icon.querySelector('.img') as HTMLElement, spell.itemId ? 'item' : 'spell', (spell.itemId ?? spell.spellId)!);
				this.tooltip.attach(icon, () => (spell.itemId ? { kind: 'item', id: spell.itemId } : { kind: 'spell', id: spell.spellId! }));
				this.packIcons.push({ el: icon, spell });
			});
			return Math.ceil(spells.length / pack.perRow);
		};
		const mainY = bottom - pack.gap - pack.iconH / 2;
		const mainRows = row(pick(packMain, packMainItems), mainY);
		row(pick(packUtility, packUtilityItems), mainY - Math.max(1, mainRows) * (pack.iconH + pack.gap));
	}

	packIcon(kind: string): HTMLElement {
		return el('div', { class: `pk-icon ${kind}` }, el('div', { class: 'img' }), el('div', { class: 'swipe' }), el('div', { class: 'time' }), el('div', { class: 'stacks' }));
	}

	// updatePack fills the pack for this frame.
	updatePack(p: PlayerState, now: number, cds: Map<string, { readyAt: number; length: number }>) {
		const q = (s: string) => find(this.hud, s) as HTMLElement | null;
		const pct = (a: number, b: number) => `${b > 0 ? Math.max(0, Math.min(100, (100 * a) / b)) : 0}%`;
		const power = q('.pk-power');
		if (power) {
			setStyle(find(power, '.fill'), 'width', pct(p.mana, p.maxMana));
			setText(find(power, 'em'), String(Math.round(p.mana)));
		}
		const swing = q('.pk-swing');
		if (swing) {
			const next = p.nextSwingAt ?? now;
			const frac = p.attackOn && p.swingMs > 0 ? (p.swingMs - (next - now)) / p.swingMs : 0;
			setStyle(find(swing, '.fill'), 'width', `${Math.max(0, Math.min(1, frac)) * 100}%`);
			swing.classList.toggle('alert', p.attackOn && !p.swinging && this.fight.phase === 'fight');
		}
		const maelstrom = (p.auras ?? []).find(a => a.label.startsWith('Maelstrom Weapon'));
		const pips = q('.pk-pips');
		if (pips) for (let i = 0; i < pips.children.length; i++) pips.children[i].classList.toggle('on', (maelstrom?.stacks ?? 0) > i);

		const inFight = this.fight.phase === 'fight';
		const show = (name: string, id: number | undefined, expiresAt: number | undefined, opts: { stacks?: number; missingId?: number }) => {
			const c = this.maintainCells.get(name);
			if (!c) return;
			const shownId = id ?? (inFight ? opts.missingId : undefined);
			c.classList.toggle('hidden', !shownId);
			c.classList.toggle('missing', !id && !!shownId);
			if (!shownId) return;
			if (c.dataset.spell !== String(shownId)) {
				c.dataset.spell = String(shownId);
				setIcon(c.querySelector('.img') as HTMLElement, 'spell', shownId);
			}
			const left = id && expiresAt !== undefined ? expiresAt - now : undefined;
			c.dataset.expires = left !== undefined ? String(expiresAt) : '';
			setText(find(c, '.time'), left !== undefined ? fmtLeft(left) : '');
			c.classList.toggle('ending', left !== undefined && left < 5000);
			c.classList.toggle('soon', left !== undefined && left >= 5000 && left < 10000);
			setText(find(c, '.stacks'), opts.stacks && opts.stacks > 1 ? String(opts.stacks) : '');
		};
		const shield = (p.auras ?? []).find(a => a.label.startsWith('Lightning Shield') || a.label.startsWith('Water Shield'));
		show('shield', shield?.spellId, shield?.expiresAt, { stacks: shield?.stacks, missingId: this.toolkit.spellByName('Lightning Shield') });
		for (const element of totemElements) {
			const t = (p.totems ?? []).find(t => t.element === element);
			const c = this.maintainCells.get(element);
			if (t && c) c.dataset.last = String(t.spellId);
			show(element, t?.spellId, t?.expiresAt, { missingId: c?.dataset.last ? +c.dataset.last : undefined });
		}
		const flameShock = this.toolkit.spellByName('Flame Shock');
		const dot = (this.target?.debuffs ?? []).find(d => (d.spellId && this.toolkit.names[d.spellId]?.name === 'Flame Shock') || /^Flame ?Shock/.test(d.label));
		show('dot', dot?.spellId, dot?.expiresAt, { missingId: flameShock });
		this.maintainCells.get('dot')?.classList.toggle('faded', !dot);

		for (const { el: icon, spell } of this.packIcons) {
			const cd = cds.get(spell.itemId ? `i${spell.itemId}` : `s${spell.spellId}`);
			const left = cd && cd.readyAt > now ? cd.readyAt - now : 0;
			icon.classList.toggle('on-cd', left > 0);
			setStyle(find(icon, '.swipe'), '--elapsed', String(left > 0 && cd!.length > 0 ? 1 - left / cd!.length : 1));
			setText(find(icon, '.time'), left > 0 ? fmtLeft(left) : '');
			icon.classList.toggle('oom', spell.cost > p.mana);
		}
	}

	// ---- per frame ----

	updateHud() {
		const now = this.now();
		const q = (s: string) => find(this.hud, s);
		const p = this.player;
		const pct = (a: number, b: number) => `${b > 0 ? Math.max(0, Math.min(100, (100 * a) / b)) : 0}%`;
		const hpText = (cur: number, max: number) => `${fmtNumber(cur)} | ${max > 0 ? ((100 * cur) / max).toFixed(1) : '0.0'}%`;
		if (p) {
			setStyle(q('.player-frame .fill'), 'width', pct(p.health, p.maxHealth));
			setText(q('.player-frame .right'), hpText(p.health, p.maxHealth));
		}
		const hasTarget = this.hasTarget && !!this.target;
		q('.target-frame').classList.toggle('hidden', !hasTarget);
		q('.tot-frame').classList.toggle('hidden', !hasTarget || !(p?.attackOn || this.fight.phase === 'fight'));
		if (this.target) {
			// The dummy can't die, so its frame shows the damage it took instead of its health.
			setStyle(q('.target-frame .fill'), 'width', this.settings.executeFromStart ? '19%' : '100%');
			setText(q('.target-frame .left'), fmtNumber(this.target.damageTaken));
			setText(q('.target-frame .name'), this.target.name);
			this.renderAuras(q('.target-frame .debuffs'), this.target.debuffs ?? [], now, 'debuff');
		}
		if (p) {
			const buffs = (p.auras ?? []).filter(a => a.expiresAt !== undefined || (a.spellId ?? 0) > 0);
			const fixed = (a: AuraState) => a.expiresAt === undefined && this.settingAuras.has(auraKey(a));
			this.renderAuras(q('.buffs .fixed'), buffs.filter(fixed), now, 'buff');
			this.renderAuras(q('.buffs .timed'), buffs.filter(a => !fixed(a)), now, 'buff');
		}
		this.updateNameplate(now);
		this.drawHistory(now);

		// The cast bar, with the ping's share at its end like EllesmereUI's latency.
		const cast = p?.cast;
		const castbar = q('.castbar');
		if (cast && cast.end > now) {
			castbar.classList.remove('hidden');
			if (castbar.dataset.spell !== String(cast.spellId)) {
				castbar.dataset.spell = String(cast.spellId);
				setIcon(q('.cb-icon'), 'spell', cast.spellId);
			}
			setStyle(q('.cb-bar .fill'), 'width', pct(now - cast.start, cast.end - cast.start));
			setStyle(q('.cb-bar .latency'), 'width', pct(this.settings.pingMs, cast.end - cast.start));
			setText(q('.cb-name'), this.toolkit?.nameOf(cast.spellId) ?? '');
			setText(q('.cb-time'), Math.max(0, (cast.end - now) / 1000).toFixed(1));
		} else {
			castbar.classList.add('hidden');
			castbar.dataset.spell = '';
		}

		// Buttons: GCD and cooldown sweeps, out of mana, out of range.
		const cds = new Map<string, { readyAt: number; length: number }>();
		for (const c of p?.cooldowns ?? []) cds.set(c.spellId ? `s${c.spellId}` : `i${c.itemId}`, c);
		if (p && this.toolkit) {
			this.updatePack(p, now, cds);
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
				let cdLeft = 0;
				const cd = cds.get(key);
				if (cd && cd.readyAt > now && cd.length > 0) {
					frac = Math.max(frac, (cd.readyAt - now) / cd.length);
					cdLeft = cd.readyAt - now;
				}
				if (spell && spell.gcdMs > 0 && p.gcdEnd > now && p.gcdEnd > p.gcdStart) frac = Math.max(frac, (p.gcdEnd - now) / (p.gcdEnd - p.gcdStart));
				setStyle(find(b, '.sweep'), '--sweep', `${Math.min(1, frac) * 360}deg`);
				setText(find(b, '.cd'), cdLeft > 1500 ? fmtLeft(cdLeft) : '');
				b.classList.toggle('oom', !!spell && spell.cost > p.mana);
				const range = spell?.melee ? 5 : spell?.maxRange ? spell.maxRange + 3 : 0;
				b.classList.toggle('oor', !!spell?.harmful && range > 0 && (dist > range || !this.hasTarget));
				b.classList.toggle('queued', !!spell && spell.spellId !== undefined && spell.spellId === p.queuedSpellId);
			}
		}

		// The fight bar, the countdown and the damage meter.
		const phase = this.fight.phase;
		setText(q('.phase'), { prepull: 'Prepull', countdown: 'Countdown', fight: this.fight.mode === 'encounter' ? 'Encounter' : 'Free', ended: 'Ended' }[phase]);
		const shown = phase === 'ended' ? (this.fight.endedAtMs ?? 0) : now;
		setText(q('.timer'), phase === 'fight' || phase === 'ended' ? `${fmtTime(shown)}${this.fight.mode === 'encounter' ? ' / ' + fmtTime(this.fight.lengthMs) : ''}` : phase === 'countdown' ? fmtTime(now) : '');
		q('.start-encounter').classList.toggle('hidden', phase !== 'prepull');
		q('.stop').classList.toggle('hidden', phase !== 'fight');
		const countdown = q('.countdown');
		setText(countdown, phase === 'countdown' && now < 0 ? String(Math.ceil(-now / 1000)) : '');
		const damage = p?.damageDone ?? 0;
		const elapsed = phase === 'fight' ? Math.max(now, 1) : phase === 'ended' ? Math.max(this.fight.endedAtMs ?? 1, 1) : 0;
		setText(q('.dm-header'), `Damage Done${elapsed ? ` (${fmtTime(elapsed)})` : ''}`);
		q('.dm-bar').classList.toggle('hidden', !elapsed);
		if (elapsed) {
			setText(q('.dm-bar .l'), this.playerName());
			setText(q('.dm-bar .r'), `${fmtNumber(damage)} | ${((damage * 1000) / elapsed).toFixed(1)}`);
		}
	}

	// updateNameplate keeps the dummy's EllesmereUI nameplate over its head.
	updateNameplate(now: number) {
		const plate = find(this.hud, '.nameplate');
		const p = this.room.screenPoint(this.dummyPos, 2.75);
		plate.classList.toggle('hidden', !p.visible || !this.target);
		if (!p.visible || !this.target) return;
		const L = this.layout;
		const w = L.nameplate.w * this.u;
		setStyle(plate, 'left', `${Math.round(p.x - w / 2)}px`);
		setStyle(plate, 'top', `${Math.round(p.y)}px`);
		setStyle(plate, 'width', `${Math.round(w)}px`);
		setStyle(find(plate, '.np-hp'), 'height', `${Math.round(L.nameplate.h * this.u)}px`);
		setText(find(plate, '.np-name'), this.target.name);
		plate.classList.toggle('targeted', this.hasTarget);
		plate.classList.toggle('in-combat', this.fight.phase === 'fight');
		setStyle(find(plate, '.np-hp .fill'), 'width', this.settings.executeFromStart ? '19%' : '100%');
		this.renderAuras(plate.querySelector('.np-debuffs') as HTMLElement, this.target.debuffs ?? [], now, 'np');
	}

	renderAuras(box: HTMLElement, auras: AuraState[], now: number, kind: 'buff' | 'debuff' | 'np') {
		const shown = auras.filter(a => a.spellId || a.itemId);
		const key = shown.map(a => `${a.spellId}:${a.itemId}:${a.stacks}`).join('|');
		if (box.dataset.key !== key) {
			box.dataset.key = key;
			box.replaceChildren();
			for (const a of shown) {
				const icon = el('div', { class: `aura ${kind}` }, el('div', { class: 'icon' }), el('div', { class: 'time' }), el('div', { class: 'stacks' }, a.stacks && a.stacks > 1 ? String(a.stacks) : ''));
				setIcon(icon.querySelector('.icon') as HTMLElement, a.spellId ? 'spell' : 'item', (a.spellId || a.itemId)!);
				const expiresAt = a.expiresAt;
				this.tooltip.attach(icon, () => ({ kind: 'aura', spellId: a.spellId, itemId: a.itemId, label: a.label, remaining: () => (expiresAt !== undefined ? fmtRemaining(expiresAt - this.now()) : '') }));
				box.append(icon);
			}
		}
		const times = box.querySelectorAll('.time');
		shown.forEach((a, i) => {
			const t = times[i] as HTMLElement | undefined;
			if (t) setText(t, a.expiresAt !== undefined ? fmtLeft(a.expiresAt - now) : '');
		});
	}

	// The cast history timeline, like the Replay view's: our casts where we made them, the newest
	// at the current time on the right, sliding left as time goes on. Casts too close together
	// for their icons to fit side by side go up a row. Sizes are in 4K pixels like the Bayov
	// pack, scaled with the UI.
	historyScale() {
		return {
			k: this.u / 1.8,
			iconW: 51,
			iconH: 36,
			rows: 2,
			rowH: 40,
			pxPerSec: 57,
			nowPx: 37,
			labelEvery: 5,
		};
	}

	addHistory(at: number, spellId: number | undefined, itemId: number | undefined) {
		if (!spellId && !itemId) return;
		const icon = el('div', { class: 'h-icon' });
		setIcon(icon, spellId ? 'spell' : 'item', (spellId || itemId)!);
		this.tooltip.attach(icon, () => (spellId ? { kind: 'spell', id: spellId, extra: [`Cast at ${fmtClock(at)}`] } : { kind: 'item', id: itemId!, extra: [`Used at ${fmtClock(at)}`] }));
		(this.hud.querySelector('.hist-casts') as HTMLElement).append(icon);
		this.castLog.push({ at, spellId, itemId, el: icon, row: 0, shift: 0 });
		// We place the casts from the newest back. A cast goes in the lowest row where it doesn't
		// cover the icon after it. When no row has room (a burst of cooldowns at the pull), it
		// moves left to the first free place in a row, so each icon still shows.
		const h = this.historyScale();
		const step = h.iconW + 3;
		const rowStarts = new Array<number>(h.rows).fill(Infinity);
		for (let i = this.castLog.length - 1; i >= 0; i--) {
			const c = this.castLog[i];
			const x = (c.at / 1000) * h.pxPerSec;
			let row = rowStarts.findIndex(start => x + step <= start);
			let placed = x;
			if (row < 0) {
				row = rowStarts.indexOf(Math.max(...rowStarts));
				placed = rowStarts[row] - step;
			}
			rowStarts[row] = placed;
			c.row = row;
			c.shift = placed - x;
		}
	}

	clearHistory() {
		for (const c of this.castLog) c.el.remove();
		this.castLog = [];
		for (const l of this.axisLabels.values()) l.remove();
		this.axisLabels.clear();
	}

	drawHistory(now: number) {
		const box = find(this.hud, '.history');
		const h = this.historyScale();
		const k = h.k;
		const width = this.historyWidth || box.clientWidth;
		const nowX = width - h.nowPx * k;
		const xAt = (at: number) => nowX - ((now - at) / 1000) * h.pxPerSec * k;
		const oldest = now - ((nowX + h.iconW * k) / (h.pxPerSec * k)) * 1000;
		while (this.castLog.length && this.castLog[0].at + 0 < oldest - 2000) this.castLog.shift()!.el.remove();
		for (const c of this.castLog) {
			const pop = Math.min(1, Math.max(0, (now - c.at) / 150));
			const eased = 1 - (1 - pop) * (1 - pop);
			const x = xAt(c.at) + c.shift * k - (h.iconW * k) / 2;
			setStyle(c.el, 'width', `${h.iconW * k}px`);
			setStyle(c.el, 'height', `${h.iconH * k}px`);
			setStyle(c.el, 'top', `${(h.rows - 1 - c.row) * h.rowH * k}px`);
			setStyle(c.el, 'transform', `translateX(${x.toFixed(1)}px) scale(${(1.12 - 0.12 * eased).toFixed(3)})`);
			setStyle(c.el, 'opacity', eased.toFixed(3));
		}
		// The axis: ticks every second slide with the time, a label every few seconds.
		const axis = find(this.hud, '.hist-axis');
		setStyle(axis, '--offset', `${xAt(0).toFixed(1)}px`);
		setStyle(axis, '--sec', `${h.pxPerSec * k}px`);
		const every = h.labelEvery * 1000;
		const labels = new Set<number>();
		for (let at = Math.ceil(oldest / every) * every; at <= now; at += every) {
			labels.add(at);
			let l = this.axisLabels.get(at);
			if (!l) {
				l = el('span', { class: 'hist-label' }, fmtTime(at));
				this.axisLabels.set(at, l);
				axis.append(l);
			}
			setStyle(l, 'transform', `translateX(${xAt(at).toFixed(1)}px) translateX(-50%)`);
		}
		for (const [at, l] of this.axisLabels) {
			if (!labels.has(at)) {
				l.remove();
				this.axisLabels.delete(at);
			}
		}
	}
}

const graphicsKey = 'practice.graphics';

// readGraphics gives the graphics settings saved in this browser, or High.
function readGraphics(): Graphics {
	try {
		const saved = JSON.parse(localStorage.getItem(graphicsKey) ?? 'null');
		if (saved && typeof saved.renderScale === 'number') return { ...graphicsPresets.high, ...saved };
	} catch {
		// Nothing saved.
	}
	return { ...graphicsPresets.high };
}

function auraKey(a: AuraState): string {
	return `${a.spellId ?? ''}:${a.itemId ?? ''}:${a.label}`;
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

// fmtNumber writes big numbers short, like the game's unit frames: 12.3K, 1.2M.
function fmtNumber(n: number): string {
	const v = Math.round(n);
	if (Math.abs(v) >= 1e6) return `${(v / 1e6).toFixed(1)}M`;
	if (Math.abs(v) >= 1e4) return `${(v / 1e3).toFixed(1)}K`;
	return String(v);
}

// fmtClock is a moment of the fight like the Replay view's clock: "1:05.3", "-0:02.0".
function fmtClock(ms: number): string {
	const neg = ms < 0;
	const t = Math.round(Math.abs(ms) / 100) / 10;
	const m = Math.floor(t / 60);
	return `${neg ? '-' : ''}${m}:${(t - m * 60).toFixed(1).padStart(4, '0')}`;
}

// fmtRemaining is an aura tooltip's last line, like the game's.
function fmtRemaining(ms: number): string {
	if (ms <= 0) return '';
	const s = ms / 1000;
	if (s >= 3600) return `${Math.ceil(s / 3600)} hours remaining`;
	if (s >= 60) return `${Math.ceil(s / 60)} minutes remaining`;
	return `${Math.ceil(s)} seconds remaining`;
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
