// The Blizzard panels' shared base: the layer they sit in (scaled like the game's UIParent), the
// panel manager that opens, stacks and closes them like UIParent's UIPanelWindows, and the
// shared widgets (textures, buttons, tabs, checkboxes, sliders, scroll frames, item buttons).
//
// Everything is in UI units. The layer has CSS zoom set to the room's UI scale, so a frame's
// numbers are the FrameXML's numbers. A closed panel is display: none and does no work.

import './frames.css';
import { setIcon } from '../icons';
import { glow, loadFonts, sharedVars, SND, sound, tex } from './assets';

// The game's left panel spot (UIParent's TOP_OFFSET and LEFT_OFFSET in Classic).
const topOffset = 104;
const leftOffset = 0;

export type PanelArea = 'left' | 'dialog' | 'free';

export abstract class Panel {
	// The frame's size, as in its FrameXML.
	abstract readonly width: number;
	abstract readonly height: number;
	// How much room the panel takes in the left row: its art's width, without the see-through
	// margin on its right.
	stackWidth = 0;
	area: PanelArea = 'left';
	openSound?: number;
	closeSound?: number;
	el?: HTMLElement;
	manager?: UIPanels;
	isOpen = false;
	private built?: Promise<void>;

	// build makes the frame's content, on its first open.
	protected abstract build(root: HTMLElement): Promise<void> | void;
	// shown runs on each open, after build, to bring the content up to date.
	protected shown(): void {}
	protected hidden(): void {}

	ensureBuilt(layer: HTMLElement): Promise<void> {
		if (!this.built) {
			const el = document.createElement('div');
			el.className = 'bz-frame';
			el.style.width = `${this.width}px`;
			el.style.height = `${this.height}px`;
			el.style.display = 'none';
			layer.append(el);
			this.el = el;
			this.built = Promise.all([sharedVars(), loadFonts()])
				.then(() => this.build(el))
				.catch(e => console.warn('frame', e));
		}
		return this.built;
	}

	onShown() {
		this.shown();
	}

	onHidden() {
		this.hidden();
	}

	close() {
		this.manager?.close(this);
	}
}

export class UIPanels {
	readonly layer = document.createElement('div');
	private panels: Panel[] = [];
	// The open left panels, left to right.
	private row: Panel[] = [];
	// Listeners for when a panel opens or closes (the HUD hides its tooltips and the like).
	onChange?: () => void;

	constructor() {
		this.layer.className = 'bz-layer';
	}

	// setScale sets the room's UI scale: screen pixels per UI unit.
	setScale(u: number) {
		this.layer.style.setProperty('zoom', String(u));
		this.layer.style.width = `${innerWidth / u}px`;
		this.layer.style.height = `${innerHeight / u}px`;
		this.layout();
	}

	get uiWidth(): number {
		return parseFloat(this.layer.style.width) || innerWidth;
	}

	get uiHeight(): number {
		return parseFloat(this.layer.style.height) || innerHeight;
	}

	add<P extends Panel>(p: P): P {
		p.manager = this;
		this.panels.push(p);
		return p;
	}

	anyOpen(area?: PanelArea): boolean {
		return this.panels.some(p => p.isOpen && (!area || p.area === area));
	}

	toggle(p: Panel) {
		if (p.isOpen) this.close(p);
		else this.open(p);
	}

	async open(p: Panel) {
		if (p.isOpen) return;
		p.isOpen = true;
		if (p.openSound) sound(p.openSound);
		if (p.area === 'left') {
			this.row.push(p);
			// The game keeps as many side by side as fit, and closes the oldest for a new one.
			while (this.row.length > 1 && this.rowWidth() > this.uiWidth) this.close(this.row[0]);
		}
		await p.ensureBuilt(this.layer);
		if (!p.isOpen || !p.el) return;
		p.el.style.display = '';
		this.layout();
		p.onShown();
		this.onChange?.();
	}

	close(p: Panel, quiet = false) {
		if (!p.isOpen) return;
		p.isOpen = false;
		if (p.closeSound && !quiet) sound(p.closeSound);
		this.row = this.row.filter(x => x !== p);
		if (p.el) p.el.style.display = 'none';
		p.onHidden();
		this.layout();
		this.onChange?.();
	}

	// closeAll closes every open panel, like the game's CloseAllWindows, and says whether there
	// was one.
	closeAll(area?: PanelArea): boolean {
		let any = false;
		for (const p of this.panels) {
			if (p.isOpen && (!area || p.area === area)) {
				this.close(p);
				any = true;
			}
		}
		return any;
	}

	private rowWidth(): number {
		return this.row.reduce((w, p) => w + (p.stackWidth || p.width), leftOffset);
	}

	// layout puts the left panels side by side from the left edge, at the game's height.
	layout() {
		let x = leftOffset;
		for (const p of this.row) {
			if (!p.el) continue;
			p.el.style.left = `${x}px`;
			p.el.style.top = `${topOffset}px`;
			x += p.stackWidth || p.width;
		}
		for (const p of this.panels) {
			if (p.area !== 'dialog' || !p.el) continue;
			p.el.style.left = `${Math.round((this.uiWidth - p.width) / 2)}px`;
			p.el.style.top = `${Math.round((this.uiHeight - p.height) / 2)}px`;
		}
	}
}

// ---- Widgets ----

type Box = { x?: number; y?: number; w?: number; h?: number; right?: number; bottom?: number };

// div makes an element with classes, placed at a spot of its parent (top-left based, in UI
// units, y going down).
export function div(parent: HTMLElement | null, cls: string, box?: Box, text?: string): HTMLDivElement {
	const e = document.createElement('div');
	e.className = cls;
	if (box) place(e, box);
	if (text !== undefined) e.textContent = text;
	parent?.append(e);
	return e;
}

export function place(e: HTMLElement, box: Box) {
	e.style.position = 'absolute';
	if (box.x !== undefined) e.style.left = `${box.x}px`;
	if (box.y !== undefined) e.style.top = `${box.y}px`;
	if (box.right !== undefined) e.style.right = `${box.right}px`;
	if (box.bottom !== undefined) e.style.bottom = `${box.bottom}px`;
	if (box.w !== undefined) e.style.width = `${box.w}px`;
	if (box.h !== undefined) e.style.height = `${box.h}px`;
}

export type TexOptions = {
	// Texture coordinates as the FrameXML gives them: left, right, top, bottom from 0 to 1.
	coords?: [number, number, number, number];
	// Drawn with additive blending in the game.
	add?: boolean;
	// The texture's own size, when coords are given. Without it we read it from the file.
	size?: [number, number];
	cls?: string;
};

// texture draws a texture at a spot, like a FrameXML <Texture>.
export function texture(parent: HTMLElement, fdid: number, box: Box, opts: TexOptions = {}): HTMLDivElement {
	const e = div(parent, `bz-tex ${opts.cls ?? ''}`, box);
	setTexture(e, fdid, opts);
	return e;
}

// setTexture fills an element with a texture, stretched to its size, or cut by coords.
export function setTexture(e: HTMLElement, fdid: number, opts: TexOptions = {}) {
	const url = opts.add ? glow(fdid) : tex(fdid);
	url.then(async u => {
		e.style.backgroundImage = `url(${u})`;
		if (opts.coords) {
			const [l, r, t, b] = opts.coords;
			let size = opts.size;
			if (!size) {
				const img = new Image();
				img.src = u;
				await img.decode();
				size = [img.width, img.height];
			}
			const w = e.offsetWidth || parseFloat(e.style.width);
			const h = e.offsetHeight || parseFloat(e.style.height);
			const bw = w / (r - l);
			const bh = h / (b - t);
			e.style.backgroundSize = `${bw}px ${bh}px`;
			e.style.backgroundPosition = `${-l * bw}px ${-t * bh}px`;
		} else e.style.backgroundSize = '100% 100%';
	});
}

// fourPiece lays a panel's art: a 256 wide left part and a 128 wide right part, top and bottom,
// each 256 high, from the frame's top-left plus an offset (most panels draw them at 2, -1).
export function fourPiece(parent: HTMLElement, files: [number, number, number, number], dx = 0, dy = 0, heights: [number, number] = [256, 256]) {
	const [tl, tr, bl, br] = files;
	texture(parent, tl, { x: dx, y: dy, w: 256, h: heights[0] });
	texture(parent, tr, { x: dx + 256, y: dy, w: 128, h: heights[0] });
	texture(parent, bl, { x: dx, y: dy + heights[0], w: 256, h: heights[1] });
	texture(parent, br, { x: dx + 256, y: dy + heights[0], w: 128, h: heights[1] });
}

export type FontName =
	| 'GameFontNormal'
	| 'GameFontHighlight'
	| 'GameFontNormalSmall'
	| 'GameFontHighlightSmall'
	| 'GameFontNormalLarge'
	| 'GameFontDisable'
	| 'GameFontDisableSmall'
	| 'GameFontBlack'
	| 'GameFontGreen'
	| 'GameFontRed'
	| 'NumberFontNormal'
	| 'NumberFontNormalSmall'
	| 'SubSpellFont'
	| 'ChatFontNormal';

// text is a FontString: one of the game's fonts at a spot.
export function text(parent: HTMLElement, s: string, font: FontName, box?: Box, align: 'left' | 'center' | 'right' = 'left'): HTMLDivElement {
	const e = div(parent, `bz-text ${font}`, box, s);
	if (align !== 'left') e.style.textAlign = align;
	return e;
}

// button is a UIPanelButtonTemplate: the red three part button with yellow text.
export function button(parent: HTMLElement, label: string, box: Box, onClick: () => void): HTMLDivElement {
	const e = div(parent, 'bz-button', box);
	div(e, 'bz-button-text', undefined, label);
	clickable(e, () => {
		if (!e.classList.contains('disabled')) onClick();
	});
	return e;
}

// closeButton is UIPanelCloseButton, the red X.
export function closeButton(parent: HTMLElement, box: Box, onClick: () => void): HTMLDivElement {
	const e = div(parent, 'bz-close', { w: 32, h: 32, ...box });
	clickable(e, onClick);
	return e;
}

// tab is a CharacterFrameTabButtonTemplate, a tab under a panel.
export function tab(parent: HTMLElement, label: string, box: Box, onClick: () => void): HTMLDivElement {
	const e = div(parent, 'bz-tab', box);
	div(e, 'bz-tab-l');
	div(e, 'bz-tab-m');
	div(e, 'bz-tab-r');
	div(e, 'bz-tab-text', undefined, label);
	clickable(e, () => {
		if (e.classList.contains('selected')) return;
		onClick();
	});
	return e;
}

// selectTab marks one tab of a row as the open one.
export function selectTab(tabs: HTMLElement[], index: number) {
	tabs.forEach((t, i) => t.classList.toggle('selected', i === index));
}

// checkbox is an OptionsCheckButtonTemplate with its label on the right.
export function checkbox(parent: HTMLElement, label: string, box: Box, checked: boolean, onChange: (on: boolean) => void): HTMLDivElement {
	const e = div(parent, 'bz-check', box);
	div(e, 'bz-check-box');
	text(e, label, 'GameFontHighlightSmall', undefined).classList.add('bz-check-label');
	e.classList.toggle('checked', checked);
	clickable(e, () => {
		const on = !e.classList.contains('checked');
		e.classList.toggle('checked', on);
		sound(SND.menuOption);
		onChange(on);
	});
	return e;
}

// slider is an OptionsSliderTemplate: a title over it, the low and high values under its ends,
// and the value shown in the title.
export function slider(
	parent: HTMLElement,
	label: string,
	box: Box,
	o: { min: number; max: number; step: number; value: number; format: (v: number) => string; low?: string; high?: string },
	onChange: (v: number) => void,
): HTMLDivElement {
	const e = div(parent, 'bz-slider', { w: 144, h: 17, ...box });
	const title = text(e, '', 'GameFontNormalSmall', undefined, 'center');
	title.classList.add('bz-slider-title');
	text(e, o.low ?? o.format(o.min), 'GameFontHighlightSmall').classList.add('bz-slider-low');
	text(e, o.high ?? o.format(o.max), 'GameFontHighlightSmall').classList.add('bz-slider-high');
	const thumb = div(e, 'bz-slider-thumb');
	let value = o.value;
	const show = () => {
		const f = (value - o.min) / (o.max - o.min || 1);
		thumb.style.left = `${f * 100}%`;
		title.textContent = `${label}: ${o.format(value)}`;
	};
	show();
	const set = (clientX: number) => {
		const r = e.getBoundingClientRect();
		const f = Math.max(0, Math.min(1, (clientX - r.left) / r.width));
		const v = Math.round((o.min + f * (o.max - o.min)) / o.step) * o.step;
		const next = Math.max(o.min, Math.min(o.max, +v.toFixed(6)));
		if (next === value) return;
		value = next;
		show();
		onChange(value);
	};
	e.addEventListener('mousedown', ev => {
		ev.stopPropagation();
		set(ev.clientX);
		const move = (m: MouseEvent) => set(m.clientX);
		const up = () => {
			removeEventListener('mousemove', move);
			removeEventListener('mouseup', up);
		};
		addEventListener('mousemove', move);
		addEventListener('mouseup', up);
	});
	return e;
}

// dropdown is the game's dropdown menu as a select (pointer.ts draws its list while the mouse is
// locked).
export function dropdown(parent: HTMLElement, box: Box, options: [value: string, label: string][], value: string, onChange: (v: string) => void): HTMLSelectElement {
	const e = document.createElement('select');
	e.className = 'bz-select';
	place(e, box);
	for (const [v, l] of options) {
		const o = document.createElement('option');
		o.value = v;
		o.textContent = l;
		e.append(o);
	}
	e.value = value;
	e.addEventListener('change', () => {
		sound(SND.menuOption);
		onChange(e.value);
	});
	parent.append(e);
	return e;
}

// scrollFrame is a UIPanelScrollFrameTemplate: a clipped area with the game's scroll bar on its
// right. The content scrolls by the wheel, the arrows and dragging the knob.
export function scrollFrame(parent: HTMLElement, box: Box): { el: HTMLDivElement; content: HTMLDivElement; update: () => void; scrollTo: (y: number) => void } {
	const el = div(parent, 'bz-scroll', box);
	const view = div(el, 'bz-scroll-view');
	const content = div(view, 'bz-scroll-content');
	const bar = div(el, 'bz-scrollbar');
	const up = div(bar, 'bz-scroll-btn up');
	const track = div(bar, 'bz-scroll-track');
	const knob = div(track, 'bz-scroll-knob');
	const down = div(bar, 'bz-scroll-btn down');
	let y = 0;
	const max = () => Math.max(0, content.scrollHeight - view.clientHeight);
	const update = () => {
		const m = max();
		y = Math.max(0, Math.min(m, y));
		content.style.transform = `translateY(${-y}px)`;
		bar.style.display = m > 0 ? '' : 'none';
		up.classList.toggle('disabled', y <= 0);
		down.classList.toggle('disabled', y >= m);
		const t = track.clientHeight - knob.offsetHeight;
		knob.style.top = `${m > 0 ? (y / m) * t : 0}px`;
	};
	const scrollTo = (v: number) => {
		y = v;
		update();
	};
	el.addEventListener(
		'wheel',
		e => {
			e.preventDefault();
			e.stopPropagation();
			scrollTo(y + (e.deltaY > 0 ? 1 : -1) * (view.clientHeight / 2));
		},
		{ passive: false },
	);
	clickable(up, () => scrollTo(y - view.clientHeight / 2));
	clickable(down, () => scrollTo(y + view.clientHeight / 2));
	knob.addEventListener('mousedown', e => {
		e.stopPropagation();
		const startY = e.clientY;
		const start = y;
		const zoom = knob.getBoundingClientRect().height / (knob.offsetHeight || 1);
		const move = (m: MouseEvent) => {
			const t = track.clientHeight - knob.offsetHeight;
			if (t > 0) scrollTo(start + ((m.clientY - startY) / zoom / t) * max());
		};
		const up2 = () => {
			removeEventListener('mousemove', move);
			removeEventListener('mouseup', up2);
		};
		addEventListener('mousemove', move);
		addEventListener('mouseup', up2);
	});
	return { el, content, update, scrollTo };
}

export const qualityColors = ['#9d9d9d', '#ffffff', '#1eff00', '#0070dd', '#a335ee', '#ff8000', '#e6cc80', '#e6cc80'];

// itemButton is an ItemButtonTemplate (37 by 37): an icon, the quality border, a count and the
// hover highlight. An empty one shows its slot's background texture.
export function itemButton(parent: HTMLElement, box: Box, emptyTexture?: number, framed = true): ItemButton {
	const e = div(parent, `bz-item${framed ? ' framed' : ''}`, { w: 37, h: 37, ...box });
	const empty = emptyTexture ? texture(e, emptyTexture, { x: 0, y: 0, w: 37, h: 37 }, { cls: 'bz-item-empty' }) : undefined;
	const icon = div(e, 'bz-item-icon');
	const border = div(e, 'bz-item-border');
	const count = text(e, '', 'NumberFontNormal');
	count.classList.add('bz-item-count');
	return {
		el: e,
		set(content) {
			if (!content) {
				e.classList.remove('filled');
				icon.style.backgroundImage = '';
				delete icon.dataset.icon;
				border.style.display = 'none';
				count.textContent = '';
				if (empty) empty.style.display = '';
				return;
			}
			e.classList.add('filled');
			if (empty) empty.style.display = 'none';
			setIcon(icon, content.kind, content.id);
			const q = content.quality;
			if (q !== undefined && q >= 2) {
				border.style.display = '';
				border.style.setProperty('--q', qualityColors[q] ?? '#fff');
			} else border.style.display = 'none';
			count.textContent = content.count && content.count > 1 ? String(content.count) : '';
		},
	};
}

export type ItemButton = { el: HTMLDivElement; set(content: { kind: 'item' | 'spell'; id: number; quality?: number; count?: number } | undefined): void };

// clickable makes an element act like a button: pressed while the left button is down, a click
// on release. Its mousedown doesn't reach the world behind.
export function clickable(e: HTMLElement, onClick: (ev: MouseEvent) => void) {
	e.addEventListener('mousedown', ev => {
		if (ev.button !== 0) return;
		ev.stopPropagation();
		e.classList.add('pushed');
		const up = () => {
			e.classList.remove('pushed');
			removeEventListener('mouseup', up, true);
		};
		addEventListener('mouseup', up, true);
	});
	e.addEventListener('click', ev => {
		if (ev.button !== 0) return;
		ev.stopPropagation();
		onClick(ev);
	});
}
