// The chat frame (ChatFrame1 of Classic Era's FloatingChatFrame.xml): the General tab with the
// room's system lines and what you type, and the Combat Log tab with your hits.
//
// It looks and fades like the game's. The background, tabs and scroll buttons fade in while the
// mouse is over the frame and fade out a second after it leaves (CSS transitions). Lines fade
// out two minutes after they came and come back when you scroll or hover. The Combat Log keeps
// its hits in a list and only makes lines while its tab is shown.

import './chat.css';
import type { HitEvent } from '../protocol';
import { glow, loadFonts, SND, sound, tex, texVars } from './assets';
import type { RoomLink } from './link';
import { clickable, div } from './panels';

// ChatLayout is where the user's EllesmereUI keeps chat (cmd/practice/ui.go), in UI units.
export type ChatLayout = { position?: { point?: string; relPoint?: string; x?: number; y?: number }; size?: { w?: number; h?: number } };

// The chat textures (interface/chatframe/...).
const CT = {
	tabLeft: 374162,
	tabMid: 374164,
	tabRight: 374166,
	tabSelLeft: 374174,
	tabSelMid: 374176,
	tabSelRight: 374178,
	tabHlLeft: 374168,
	tabHlMid: 374170,
	tabHlRight: 374172,
	upUp: 130958,
	upDown: 130957,
	downUp: 130952,
	downDown: 130951,
	endUp: 130955,
	endDown: 130954,
	blink: 130945,
	mouseHilight: 130757,
	inputLeft: 375507,
	inputMid: 389190,
	inputRight: 375508,
};

// The game's chat colors (ChatTypeInfo) and combat log colors (COMBATLOG_DEFAULT_COLORS: your
// lines are 0.7 gray, and the spell, amount and school are that times 1.5).
const SYSTEM = '#ffff00';
const SAY = '#ffffff';

// The sim's school bits (sim/core/spell_school.go) and the game's names for them.
const schools: [number, string][] = [
	[1 << 1, 'Physical'],
	[1 << 2, 'Arcane'],
	[1 << 3, 'Fire'],
	[1 << 4, 'Frost'],
	[1 << 5, 'Holy'],
	[1 << 6, 'Nature'],
	[1 << 7, 'Shadow'],
];

// How many lines each tab keeps: the game's chat keeps 128, its combat log 500 or so.
const generalCap = 128;
const combatCap = 500;

type Line = { text: string; color: string };

export class ChatFrame {
	readonly el = document.createElement('div');
	private readonly msgs: HTMLDivElement;
	private readonly lines: HTMLDivElement;
	private readonly quickBar: HTMLDivElement;
	private readonly tabs: HTMLDivElement[] = [];
	private readonly bottomButton: HTMLDivElement;
	private readonly editBox: HTMLDivElement;
	private readonly input: HTMLInputElement;
	private readonly header: HTMLDivElement;
	private tab = 0;
	private editOpen = false;
	private general: Line[] = [];
	// The combat log's hits, newest last. Their text is made when the tab shows them.
	private hits: HitEvent[] = [];
	private atBottom = true;
	private wheel = 0;
	private names = new Map<number, string>();
	private asked = new Set<number>();
	private pending = new Set<number>();
	private w = 430;
	private h = 120;
	private layoutX = 32;
	private layoutY = 95;

	constructor(
		readonly link: RoomLink,
		readonly layout: ChatLayout | undefined,
	) {
		const el = this.el;
		el.className = 'bz-chat';
		if (layout?.size?.w && layout.size.h) {
			this.w = layout.size.w;
			this.h = layout.size.h;
		}
		if (layout?.position && layout.position.x !== undefined && layout.position.y !== undefined) {
			this.layoutX = layout.position.x;
			this.layoutY = layout.position.y;
		}
		el.style.width = `${this.w}px`;
		el.style.height = `${this.h}px`;

		div(el, 'bz-chat-bg');
		this.quickBar = div(el, 'bz-chat-quick');
		div(this.quickBar, 'bz-chat-quick-btn', undefined, 'My actions');
		this.msgs = div(el, 'bz-chat-msgs');
		this.lines = div(this.msgs, 'bz-chat-lines');

		// The dock: the tabs sit on the frame's top edge.
		const dock = div(el, 'bz-chat-dock');
		for (const [i, label] of ['General', 'Combat Log'].entries()) {
			const t = div(dock, 'bz-chat-tab');
			for (const part of ['l', 'm', 'r']) {
				div(t, `bz-chat-tab-bg ${part}`);
				div(t, `bz-chat-tab-sel ${part}`);
				div(t, `bz-chat-tab-hl ${part}`);
			}
			div(t, 'bz-chat-tab-text', undefined, label);
			clickable(t, () => {
				sound(SND.chatScroll);
				this.selectTab(i);
			});
			this.tabs.push(t);
		}

		// The scroll buttons on the left, bottom up: to the end, down, up.
		const buttons = div(el, 'bz-chat-buttons');
		const up = div(buttons, 'bz-chat-btn up');
		const down = div(buttons, 'bz-chat-btn down');
		this.bottomButton = div(buttons, 'bz-chat-btn end');
		clickable(up, () => this.scrollLines(-1));
		clickable(down, () => this.scrollLines(1));
		clickable(this.bottomButton, () => this.scrollToBottom());

		// The edit box under the frame.
		this.editBox = div(el, 'bz-chat-edit');
		div(this.editBox, 'bz-chat-edit-l');
		div(this.editBox, 'bz-chat-edit-m');
		div(this.editBox, 'bz-chat-edit-r');
		this.header = div(this.editBox, 'bz-chat-edit-header', undefined, 'Say: ');
		this.input = document.createElement('input');
		this.input.className = 'bz-chat-input';
		this.input.maxLength = 255;
		this.input.spellcheck = false;
		this.input.autocomplete = 'off';
		this.editBox.append(this.input);
		this.bindInput();

		this.msgs.addEventListener(
			'wheel',
			e => {
				e.preventDefault();
				e.stopPropagation();
				// A mouse notch is one line, like the game. A touchpad's small steps add up.
				this.wheel += e.deltaMode === 0 ? e.deltaY : e.deltaY * 100;
				const steps = Math.trunc(this.wheel / 100) || (Math.abs(e.deltaY) >= 100 ? Math.sign(e.deltaY) : 0);
				if (!steps) return;
				this.wheel = 0;
				this.scrollLines(steps);
			},
			{ passive: false },
		);
		// The game's chat lets your clicks through to the world: you can target or turn the camera
		// with the mouse over it. We hand the press to the room's canvas.
		el.addEventListener('mousedown', e => {
			const t = e.target as Element;
			if (t !== el && !t.closest('.bz-chat-msgs')) return;
			const canvas = this.link.three()?.renderer.domElement ?? document.querySelector('canvas');
			if (!canvas) return;
			e.stopPropagation();
			canvas.dispatchEvent(new MouseEvent('mousedown', e));
		});
		el.addEventListener('contextmenu', e => e.preventDefault());
		this.selectTab(0);
		void this.load();
		// We place the frame once it's in the layer, and again when the layer's size changes (the
		// window or the UI scale).
		requestAnimationFrame(() => {
			this.place();
			const layer = this.el.parentElement;
			if (layer) new ResizeObserver(() => this.place()).observe(layer);
		});
	}

	// load gets the textures and the chat font. The frame shows from the start, so it doesn't
	// wait for a first open.
	private async load() {
		const t = (fdid: number) => () => tex(fdid);
		const g = (fdid: number) => () => glow(fdid);
		await Promise.all([
			loadFonts(),
			texVars({
				'bz-chat-tab-l': t(CT.tabLeft),
				'bz-chat-tab-m': t(CT.tabMid),
				'bz-chat-tab-r': t(CT.tabRight),
				'bz-chat-tabsel-l': g(CT.tabSelLeft),
				'bz-chat-tabsel-m': g(CT.tabSelMid),
				'bz-chat-tabsel-r': g(CT.tabSelRight),
				'bz-chat-tabhl-l': g(CT.tabHlLeft),
				'bz-chat-tabhl-m': g(CT.tabHlMid),
				'bz-chat-tabhl-r': g(CT.tabHlRight),
				'bz-chat-up': t(CT.upUp),
				'bz-chat-up-down': t(CT.upDown),
				'bz-chat-down': t(CT.downUp),
				'bz-chat-down-down': t(CT.downDown),
				'bz-chat-end': t(CT.endUp),
				'bz-chat-end-down': t(CT.endDown),
				'bz-chat-blink': g(CT.blink),
				'bz-chat-btn-hl': g(CT.mouseHilight),
				'bz-chat-input-l': t(CT.inputLeft),
				'bz-chat-input-m': t(CT.inputMid),
				'bz-chat-input-r': t(CT.inputRight),
			}),
		]).catch(e => console.warn('chat', e));
		this.layoutTabs();
	}

	// place puts the frame where the layout says, kept on screen like the game's clamp: the scroll
	// buttons fit on its left (35) and the edit box under it. We also keep the edit box above the
	// room's Edit bars, Bind keys and Spellbook buttons in the bottom left corner.
	private place() {
		const layer = this.el.parentElement;
		const zoom = parseFloat(layer?.style.getPropertyValue('zoom') ?? '') || 1;
		const hudButton = document.querySelector('.hud .edit-bars, .hud .bind-keys');
		const hudTop = hudButton ? (innerHeight - hudButton.getBoundingClientRect().top) / zoom : 0;
		const x = Math.max(35, this.layoutX);
		const y = Math.max(50, this.layoutY, hudTop + 36);
		this.el.style.left = `${Math.round(x)}px`;
		this.el.style.bottom = `${Math.round(y)}px`;
	}

	// layoutTabs sizes each tab to its text, like PanelTemplates_TabResize with the chat's padding.
	private layoutTabs() {
		if (!this.el.isConnected) {
			requestAnimationFrame(() => this.layoutTabs());
			return;
		}
		let x = 0;
		for (const t of this.tabs) {
			const text = t.querySelector('.bz-chat-tab-text') as HTMLElement;
			const w = Math.ceil(text.offsetWidth) + 32 + 10;
			t.style.left = `${x}px`;
			t.style.width = `${w}px`;
			x += w;
		}
	}

	// The room calls these.

	// typing says whether the edit box is open, so keys go to it and not to the bars.
	get typing(): boolean {
		return this.editOpen;
	}

	// openEditBox opens the edit box, like Enter in the game.
	openEditBox() {
		if (!this.el.parentElement) return;
		// The room's UI scale may have changed since we placed the frame.
		this.place();
		this.editOpen = true;
		this.editBox.classList.add('open');
		// The text starts after the header, like the game's text insets.
		this.input.style.left = `${15 + this.header.offsetWidth}px`;
		this.input.focus();
	}

	// closeEditBox closes it, like Esc. It says whether it was open.
	closeEditBox(): boolean {
		if (!this.editOpen) return false;
		this.editOpen = false;
		this.input.value = '';
		this.editBox.classList.remove('open');
		this.input.blur();
		return true;
	}

	// system adds a yellow system line to the General tab.
	system(text: string) {
		this.addGeneral({ text, color: SYSTEM });
	}

	// combat adds the session's hits to the Combat Log tab. While the tab is hidden we only keep
	// them.
	combat(hits: HitEvent[]) {
		for (const h of hits) this.hits.push(h);
		if (this.hits.length > combatCap * 1.5) this.hits.splice(0, this.hits.length - combatCap);
		if (this.tab !== 1) return;
		this.appendLines(hits.map(h => this.combatLine(h)));
	}

	// ---- Edit box ----

	private bindInput() {
		const input = this.input;
		input.addEventListener('keydown', e => {
			// The room's keys skip an input's keys. Nothing else should see them either.
			e.stopPropagation();
			if (e.key === 'Enter') {
				e.preventDefault();
				const text = input.value;
				this.closeEditBox();
				this.send(text);
			} else if (e.key === 'Escape') {
				e.preventDefault();
				this.closeEditBox();
			}
		});
		// A click in the world doesn't take the keyboard away from the edit box, like the game.
		input.addEventListener('blur', e => {
			if (!this.editOpen) return;
			const next = e.relatedTarget as HTMLElement | null;
			if (next && /^(INPUT|SELECT|TEXTAREA)$/.test(next.tagName)) {
				this.closeEditBox();
				return;
			}
			setTimeout(() => this.editOpen && document.hasFocus() && input.focus());
		});
		addEventListener('focus', () => this.editOpen && input.focus());
	}

	// send says what you typed: plain text and /say (/s) say it, other slash commands get the
	// game's help line.
	private send(raw: string) {
		let text = raw.trim();
		if (!text) return;
		if (text.startsWith('/')) {
			const m = /^\/(\S+)\s*(.*)$/.exec(text);
			const cmd = m?.[1].toLowerCase() ?? '';
			if (cmd !== 'say' && cmd !== 's') {
				this.system("Type '/help' for a listing of a few commands.");
				return;
			}
			text = m?.[2].trim() ?? '';
			if (!text) return;
		}
		this.addGeneral({ text: `[${this.link.playerName()}] says: ${text}`, color: SAY });
	}

	// ---- Lines and scrolling ----

	private selectTab(i: number) {
		this.tab = i;
		this.tabs.forEach((t, j) => t.classList.toggle('selected', i === j));
		this.el.classList.toggle('combat', i === 1);
		this.lines.textContent = '';
		this.atBottom = true;
		if (i === 0) this.appendLines(this.general.map(l => lineEl(escape(l.text), l.color)));
		else this.appendLines(this.hits.slice(-combatCap).map(h => this.combatLine(h)));
		this.scrollToBottom();
	}

	private addGeneral(line: Line) {
		this.general.push(line);
		if (this.general.length > generalCap) this.general.shift();
		if (this.tab === 0) this.appendLines([lineEl(escape(line.text), line.color)]);
	}

	// appendLines adds lines at the bottom, drops the oldest past the cap and keeps the view at
	// the bottom when it was there.
	private appendLines(els: HTMLElement[]) {
		if (!els.length) return;
		for (const e of els) e.classList.add('new');
		this.lines.append(...els);
		const cap = this.tab === 0 ? generalCap : combatCap;
		while (this.lines.childElementCount > cap) this.lines.firstElementChild!.remove();
		// Only the newest lines run their fade. The older ones are above the view and stay hidden
		// until you scroll up to them, so we don't keep hundreds of fades going.
		const kids = this.lines.children;
		const keep = Math.ceil(this.h / 15) + 8;
		for (let i = kids.length - keep - 1; i >= 0 && kids[i].classList.contains('new'); i--) kids[i].classList.remove('new');
		if (this.atBottom) this.msgs.scrollTop = this.msgs.scrollHeight;
		else this.bottomButton.classList.add('flash');
	}

	private lineHeight(): number {
		return (this.lines.lastElementChild as HTMLElement | null)?.offsetHeight || 15;
	}

	private scrollLines(n: number) {
		this.msgs.scrollTop += n * this.lineHeight();
		this.afterScroll();
	}

	private scrollToBottom() {
		this.msgs.scrollTop = this.msgs.scrollHeight;
		this.afterScroll();
	}

	private afterScroll() {
		const m = this.msgs;
		this.atBottom = m.scrollTop >= m.scrollHeight - m.clientHeight - 2;
		this.bottomButton.classList.toggle('flash', !this.atBottom);
		this.lines.classList.toggle('scrolled', !this.atBottom);
		this.restartFades();
	}

	// restartFades shows the faded lines again and starts their two minutes over.
	private restartFades() {
		this.lines.classList.add('reset');
		void this.lines.offsetWidth;
		this.lines.classList.remove('reset');
	}

	// ---- Combat log text ----

	// combatLine writes a hit the way the game's combat log does by default (Blizzard_CombatLog
	// in full text mode, the ACTION_*_FULL_TEXT strings). The line is gray, the spell, amount and
	// school are white.
	private combatLine(h: HitEvent): HTMLElement {
		const target = escape(this.link.enemyName(h.target));
		const white = h.otherId === 7;
		const spell = white ? '' : this.spellName(h);
		let source = escape(this.link.playerName());
		if (h.source === 'pet') source = 'Pet';
		else if (h.source === 'totem' && spell && !/^Fire Nova/.test(spell)) source = escape(this.totemName(spell));
		const own = `${source}'s`;
		const s = `<b>${escape(spell)}</b>`;
		const amount = `<b>${Math.round(h.damage).toLocaleString('en-US')}</b> <b>${schoolName(h.school)}</b>`;
		const landed = h.outcome !== 'miss' && h.outcome !== 'dodge' && h.outcome !== 'parry';
		let text: string;
		if (!landed) {
			const how = h.outcome === 'dodge' ? 'was dodged by' : h.outcome === 'parry' ? 'was parried by' : '';
			// A spell that isn't a weapon blow misses as a full resist.
			const resisted = h.outcome === 'miss' && !white && !(h.school & 2);
			if (white) text = how ? `${own} attack ${how} ${target}.` : `${own} attack misses ${target}.`;
			else if (resisted) text = `${own} ${s} was fully resisted by ${target}.`;
			else text = how ? `${own} ${s} ${how} ${target}.` : `${own} ${s} misses ${target}.`;
		} else {
			const result = h.outcome === 'crit' ? ' (Critical)' : h.outcome === 'glance' ? ' (Glancing)' : '';
			if (white) text = `${own} melee swing hits ${target} for ${amount}.${result}`;
			else if (h.periodic) text = `${target} suffers ${amount} damage from ${own} ${s}.${result}`;
			else text = `${own} ${s} hits ${target} for ${amount}.${result}`;
		}
		return lineEl(text, '#b3b3b3', 'cl');
	}

	// spellName names a hit's spell: the client's spell when it shows another one (imbue procs),
	// the item for an item's hit. Names the toolkit lacks come from the server once.
	private spellName(h: HitEvent): string {
		if (!h.spellId && h.itemId) return this.link.itemName(h.itemId) ?? `Item ${h.itemId}`;
		const id = h.visualSpellId ?? h.spellId;
		if (!id) return 'Unknown';
		const name = this.link.toolkit()?.names[id]?.name ?? this.names.get(id);
		if (name) return name;
		this.fetchName(id);
		return `Spell ${id}`;
	}

	// fetchName asks the server for a spell's name once. The names missing from one update go in
	// one request, and the log redraws when they come.
	private fetchName(id: number) {
		if (this.asked.has(id)) return;
		this.asked.add(id);
		this.pending.add(id);
		if (this.pending.size === 1) queueMicrotask(() => void this.fetchNames());
	}

	private async fetchNames() {
		const ids = [...this.pending];
		this.pending.clear();
		try {
			const res = await fetch(`/api/spells?ids=${ids.join(',')}`);
			const got = (await res.json()) as Record<string, { name: string }>;
			for (const [k, v] of Object.entries(got)) if (v?.name) this.names.set(Number(k), v.name);
		} catch (e) {
			console.warn('chat names', e);
		}
		if (this.tab === 1) this.selectTab(1);
	}

	// totemName is the totem that cast a spell, as the game names the totem (Searing Totem IV).
	private totemName(spell: string): string {
		const name = /Magma/.test(spell) ? 'Magma Totem' : /Flametongue/.test(spell) ? 'Flametongue Totem' : 'Searing Totem';
		const names = this.link.toolkit()?.names ?? {};
		let rank = 0;
		for (const n of Object.values(names)) {
			if (n.name !== name) continue;
			rank = Math.max(rank, parseInt(/\d+/.exec(n.rank ?? '')?.[0] ?? '1'));
		}
		return rank > 1 ? `${name} ${roman(rank)}` : name;
	}
}

function lineEl(html: string, color: string, cls = ''): HTMLDivElement {
	const e = document.createElement('div');
	e.className = `bz-chat-line ${cls}`;
	e.style.color = color;
	e.innerHTML = html;
	return e;
}

function schoolName(school: number): string {
	return schools.find(([bit]) => school & bit)?.[1] ?? 'Unknown';
}

function roman(n: number): string {
	return ['', 'I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII', 'IX', 'X'][n] ?? String(n);
}

function escape(s: string): string {
	return s.replace(/[&<>"]/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c]!);
}
