// The shared Forever templates the character frame uses: sliced atlases, the MinimalScrollBar
// with its scroll box, the ColoredProgressBarTemplate (skill and reputation bars), the
// CharacterFrameSidePaneTemplate (the detail pane on the right) and the LargeSideTabButtonTemplate
// tabs. From the Gethe/wow-ui-source `forever` branch: Blizzard_SharedXML/Shared/Scroll/
// MinimalScrollBar.xml, Blizzard_SharedXML/Camelot/ProgressBars/ColoredProgressBar.xml,
// Camelot/CharacterFrame.xml and Mainline/SharedUIPanelTemplates.xml.

import { atlasInfo, atlasTexture, atlasUrl, setAtlas } from './atlas';
import { clickable, div, text, type Box } from './panels';

// sliced draws an atlas with its slice margins (UiTextureAtlasElementSliceData), the way the
// client draws a texture bigger than its atlas: the corners at their size and the edges stretched
// or tiled. Margins are in UI units.
export async function sliced(e: HTMLElement, name: string, margins: [l: number, t: number, r: number, b: number], tile = false) {
	const [a, u] = await Promise.all([atlasInfo(name), atlasUrl(name)]);
	if (!a || !u) return;
	const s = (a.right - a.left) / a.width;
	const [l, t, r, b] = margins;
	e.style.boxSizing = 'border-box';
	e.style.borderStyle = 'solid';
	e.style.borderColor = 'transparent';
	e.style.borderWidth = `${t}px ${r}px ${b}px ${l}px`;
	e.style.borderImage = `url(${u}) ${t * s} ${r * s} ${b * s} ${l * s} fill / ${t}px ${r}px ${b}px ${l}px ${tile ? 'round' : 'stretch'}`;
}

// ---- MinimalScrollBar ----

export type ScrollBox = {
	el: HTMLDivElement;
	content: HTMLDivElement;
	update(): void;
	scrollTo(y: number): void;
};

// scrollBox is a WowScrollBoxList at box with a MinimalScrollBar at barBox (both in the
// parent's units). fade is the ScrollBox's edge fade length: the list fades out at an edge that
// has more past it.
export function scrollBox(parent: HTMLElement, box: Box, barBox: Box, opts: { fade?: number; step?: number } = {}): ScrollBox {
	const el = div(parent, 'bz-cs-box', box);
	const content = div(el, 'bz-cs-content');
	const bar = div(parent, 'bz-msb', { w: 8, ...barBox });
	const track = div(bar, 'bz-msb-track');
	atlasTexture(track, 'minimal-scrollbar-track-top', { x: 0, y: 0, w: 8, h: 8 });
	const mid = atlasTexture(track, '!minimal-scrollbar-track-middle', { x: 0, y: 8, w: 8 });
	mid.style.bottom = '8px';
	const end = atlasTexture(track, 'minimal-scrollbar-track-bottom', { x: 0, w: 8, h: 8 });
	end.style.bottom = '0';
	const thumb = div(track, 'bz-msb-thumb', { x: 0, w: 8 });
	atlasTexture(thumb, 'minimal-scrollbar-small-thumb-top', { x: 0, y: 0, w: 8, h: 8 });
	const tmid = atlasTexture(thumb, 'minimal-scrollbar-small-thumb-middle', { x: 0, y: 8, w: 8 });
	tmid.style.bottom = '8px';
	const tend = atlasTexture(thumb, 'minimal-scrollbar-small-thumb-bottom', { x: 0, w: 8, h: 8 });
	tend.style.bottom = '0';
	const arrow = (dir: 'top' | 'bottom') => {
		const b = div(bar, `bz-msb-arrow ${dir}`, { x: -4.5, w: 17, h: 11 });
		for (const state of ['', '-over', '-down']) {
			const t = div(b, `bz-msb-arrow-tex${state || '-normal'}`);
			void setAtlas(t, `minimal-scrollbar-arrow-${dir}${state}`);
		}
		return b;
	};
	const back = arrow('top');
	back.style.top = '0';
	const forward = arrow('bottom');
	forward.style.bottom = '0';

	let y = 0;
	const max = () => Math.max(0, content.scrollHeight - el.clientHeight);
	const update = () => {
		const m = max();
		y = Math.max(0, Math.min(m, y));
		content.style.transform = `translateY(${-y}px)`;
		back.classList.toggle('disabled', y <= 0);
		forward.classList.toggle('disabled', y >= m);
		thumb.style.display = m > 0 ? '' : 'none';
		const th = track.clientHeight;
		const h = Math.max(20, Math.round((th * el.clientHeight) / Math.max(1, content.scrollHeight)));
		thumb.style.height = `${h}px`;
		thumb.style.top = `${m > 0 ? ((th - h) * y) / m : 0}px`;
		if (opts.fade) {
			const f = opts.fade;
			const top = y > 0 ? `transparent 0, #000 ${f}px` : '#000 0';
			const bottom = y < m ? `#000 calc(100% - ${f}px), transparent 100%` : '#000 100%';
			el.style.maskImage = `linear-gradient(${top}, ${bottom})`;
		}
	};
	const scrollTo = (v: number) => {
		y = v;
		update();
	};
	const step = opts.step ?? 60;
	el.addEventListener(
		'wheel',
		e => {
			e.preventDefault();
			e.stopPropagation();
			scrollTo(y + (e.deltaY > 0 ? step : -step));
		},
		{ passive: false },
	);
	clickable(back, () => scrollTo(y - step / 2));
	clickable(forward, () => scrollTo(y + step / 2));
	thumb.addEventListener('mousedown', e => {
		if (e.button !== 0) return;
		e.stopPropagation();
		e.preventDefault();
		const startY = e.clientY;
		const start = y;
		const zoom = thumb.getBoundingClientRect().height / (thumb.offsetHeight || 1);
		const move = (m: MouseEvent) => {
			const t = track.clientHeight - thumb.offsetHeight;
			if (t > 0) scrollTo(start + ((m.clientY - startY) / zoom / t) * max());
		};
		const up = () => {
			removeEventListener('mousemove', move);
			removeEventListener('mouseup', up);
		};
		addEventListener('mousemove', move);
		addEventListener('mouseup', up);
	});
	return { el, content, update, scrollTo };
}

// ---- ColoredProgressBarTemplate ----

export type ProgressBar = { el: HTMLDivElement; set(fraction: number, label: string | Node[]): void };

// progressBar is a ColoredProgressBarTemplate with the blue fill (SkillsBarTemplate): the
// stat-bar background, the fill inside its 6 unit margins, masked to the bar's round ends, and the
// text over it.
export function progressBar(parent: HTMLElement, box: Box): ProgressBar {
	const el = div(parent, 'bz-cpb', { h: 29, ...box });
	const bg = div(el, 'bz-cpb-bg');
	void sliced(bg, 'common-stat-bar-BG', [29, 0, 14, 0]);
	const masked = div(el, 'bz-cpb-masked');
	void Promise.all([atlasInfo('common-stat-bar-Mask'), atlasUrl('common-stat-bar-Mask')]).then(([a, u]) => {
		if (!a || !u) return;
		const s = (a.right - a.left) / a.width;
		masked.style.setProperty('-webkit-mask-box-image', `url(${u}) 0 ${10 * s} 0 ${10 * s} / 0 10px 0 10px stretch`);
	});
	const area = div(masked, 'bz-cpb-area');
	const fill = div(area, 'bz-cpb-fill');
	void setAtlas(fill, 'common-stat-bar-blue');
	const label = text(el, '', 'GameFontHighlight');
	label.classList.add('bz-cpb-text');
	return {
		el,
		set(fraction, l) {
			const p = Math.min(1, Math.max(0, fraction));
			fill.style.display = p > 0 ? '' : 'none';
			fill.style.width = `${p * 100}%`;
			// The fill shows the left part of the bar art (SetTexCoord(0, p, 1, 0)), at the art's
			// full width.
			fill.style.backgroundSize = `${100 / Math.max(p, 0.001)}% 100%`;
			label.textContent = '';
			if (typeof l === 'string') label.textContent = l;
			else label.append(...l);
		},
	};
}

// ---- CharacterFrameSidePaneTemplate ----

// SidePane is the detail pane on the frame's right (the skill and the faction details): a title,
// the divider line, a description and rows (categories and wrapped text), or the gray empty text.
export class SidePane {
	readonly el: HTMLDivElement;
	readonly title: HTMLDivElement;
	readonly extra: HTMLDivElement;
	private readonly divider: HTMLDivElement;
	private readonly description: HTMLDivElement;
	private readonly rows: HTMLDivElement;
	private readonly empty: HTMLDivElement;

	constructor(parent: HTMLElement, box: Box) {
		this.el = div(parent, 'bz-sidepane', box);
		this.title = text(this.el, '', 'GameFontNormal');
		this.title.classList.add('bz-sp-title', 'GameFontNormalMed3');
		this.divider = atlasTexture(this.el, 'UI-Character-Info-ScrollLine', { w: 210, h: 7 }, { cls: 'bz-sp-divider' });
		// The pane's own widgets go here (the skill's rank bar), between the divider and the text.
		this.extra = div(this.el, 'bz-sp-extra');
		this.description = text(this.el, '', 'GameFontHighlight');
		this.description.classList.add('bz-sp-desc');
		this.rows = div(this.el, 'bz-sp-rows');
		this.empty = text(this.el, '', 'GameFontDisable', undefined, 'center');
		this.empty.classList.add('bz-sp-empty');
	}

	setTitle(s: string) {
		this.title.textContent = s;
	}

	setDescription(s: string | undefined, height: number) {
		this.description.style.display = s ? '' : 'none';
		this.description.textContent = s ?? '';
		this.description.style.height = `${height}px`;
	}

	reset() {
		this.rows.textContent = '';
	}

	// addCategory is a CharacterFrameSidePaneCategoryTemplate: the title bar art with a label.
	addCategory(label: string) {
		const r = div(this.rows, 'bz-sp-cat');
		atlasTexture(r, 'UI-Character-Info-Title', { x: 0, y: 0, w: 195, h: 40 });
		text(r, label, 'GameFontHighlight', undefined, 'center').classList.add('bz-sp-cat-label');
	}

	// addWrapped is a CharacterFrameSidePaneWrappedRowTemplate: gold text, with white parts.
	addWrapped(parts: (string | { white: string })[]) {
		const r = text(this.rows, '', 'GameFontNormal');
		r.classList.add('bz-sp-wrapped');
		for (const p of parts) {
			if (typeof p === 'string') r.append(p);
			else {
				const w = document.createElement('span');
				w.className = 'bz-white';
				w.textContent = p.white;
				r.append(w);
			}
		}
	}

	addSpacer(h: number) {
		div(this.rows, 'bz-sp-spacer').style.height = `${h}px`;
	}

	setEmpty(s: string) {
		this.reset();
		this.setTitle('');
		this.setDescription(undefined, 1);
		this.el.classList.add('empty');
		this.empty.textContent = s;
	}

	clearEmpty() {
		this.el.classList.remove('empty');
	}
}

// ---- LargeSideTabButtonTemplate ----

export type SideTab = { el: HTMLDivElement; icon: HTMLDivElement; select(on: boolean): void };

// sideTab is a CharacterFrameModeSideTabTemplate: 55 by 55 (the art's height less its 5 see
// through units), the icon at 50 by 50 inside the tab's mask, 4 units left of center in Camelot.
export function sideTab(parent: HTMLElement, box: Box, tooltip: string, attach: (el: HTMLElement, title: string) => void, onClick: () => void): SideTab {
	const el = div(parent, 'bz-sidetab', { w: 55, h: 55, ...box });
	atlasTexture(el, 'common-sidetab', { x: 0, y: -2.5, w: 55, h: 60 });
	const masked = div(el, 'bz-sidetab-masked', { x: 0, y: -2.5, w: 55, h: 60 });
	void atlasUrl('common-sidetab-mask').then(u => {
		if (!u) return;
		masked.style.maskImage = `url(${u})`;
		masked.style.maskSize = '100% 100%';
	});
	const icon = div(masked, 'bz-sidetab-icon', { x: 2.5 - 4, y: 5, w: 50, h: 50 });
	atlasTexture(el, 'common-sidetab-selected', { x: 0, y: -2.5, w: 55, h: 60 }, { cls: 'bz-sidetab-selected' });
	atlasTexture(el, 'common-sidetab-hover', { x: 0, y: -2.5, w: 55, h: 60 }, { cls: 'bz-sidetab-hover' });
	attach(el, tooltip);
	clickable(el, onClick);
	return { el, icon, select: on => el.classList.toggle('selected', on) };
}
