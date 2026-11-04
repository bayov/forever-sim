// Forever's frame chrome (its "Camelot" UI, a Mainline style UI with its own art): the
// PortraitFrameTemplate (metal NineSlice, rock background, round portrait, title, red close
// button) and the PanelTabButtonTemplate tabs under a frame. From the Gethe/wow-ui-source
// `forever` branch: Blizzard_SharedXML/Mainline/SharedUIPanelTemplates.xml with the Camelot
// overrides (close button 28 by 28 at TOPRIGHT 0, 3).

import { atlasTexture, nineSlice, setAtlas } from './atlas';
import { Box, clickable, div, text, texture } from './panels';

// Interface\FrameGeneral\UI-Background-Rock, tiled behind the frame's content.
const backgroundRock = 374155;

export type PortraitFrame = { title: HTMLDivElement; portrait: HTMLDivElement; close: HTMLDivElement; bg: HTMLDivElement };

// portraitFrame draws a PortraitFrameTemplate on root (the frame, sized already).
export function portraitFrame(root: HTMLElement, title: string, onClose: () => void, opts: { layout?: string; noPortrait?: boolean } = {}): PortraitFrame {
	root.classList.add('bz-portrait-frame');
	const bg = div(root, 'bz-tex bz-pf-bg', { x: 2, y: 21 });
	bg.style.right = '2px';
	bg.style.bottom = '2px';
	texture(bg, backgroundRock, { x: 0, y: 0 }, { cls: 'bz-pf-rock' });
	const streaks = atlasTexture(root, '_UI-Frame-TopTileStreaks', { x: 6, y: 21 });
	streaks.style.right = '2px';
	streaks.style.height = '43px';
	const portrait = div(root, 'bz-pf-portrait', { x: -5, y: -7, w: 62, h: 62 });
	if (opts.noPortrait) portrait.style.display = 'none';
	nineSlice(root, opts.layout ?? (opts.noPortrait ? 'ButtonFrameTemplateNoPortrait' : 'PortraitFrameTemplate'));
	const titleBox = div(root, 'bz-pf-title', { x: 58, y: 1, h: 20 });
	titleBox.style.right = '24px';
	const t = text(titleBox, title, 'GameFontNormal', { x: 0, y: 5 }, 'center');
	t.style.right = '0';
	const close = closeButton(root, { right: 0, y: -3 }, onClose);
	return { title: t, portrait, close, bg };
}

// closeButton is UIPanelCloseButton: the red X (RedButton-Exit), 28 by 28 in Forever.
export function closeButton(parent: HTMLElement, box: Box, onClick: () => void): HTMLDivElement {
	const e = div(parent, 'bz-redx', { w: 28, h: 28, ...box });
	const up = div(e, 'bz-redx-up');
	const down = div(e, 'bz-redx-down');
	const hi = div(e, 'bz-redx-hi');
	void setAtlas(up, 'RedButton-Exit');
	void setAtlas(down, 'RedButton-exit-pressed');
	void setAtlas(hi, 'RedButton-Highlight');
	clickable(e, onClick);
	return e;
}

// panelTab is a PanelTabButtonTemplate: a tab hanging under a frame, sized to its text (20 on
// each side, like PanelTemplates_TabResize). selectTab marks the open one.
export function panelTab(parent: HTMLElement, label: string, box: Box, onClick: () => void): HTMLDivElement {
	const e = div(parent, 'bz-ptab', { h: 32, ...box });
	// The ends hang out past the button (Left at x -3, Right at x 7, the active ones at -1 and 8),
	// and the middle runs between them.
	const set = (kind: '' | 'a', left: string, mid: string, right: string, dl: number, dr: number) => {
		const l = div(e, `bz-ptab-l${kind}`);
		const m = div(e, `bz-ptab-m${kind}`);
		const r = div(e, `bz-ptab-r${kind}`);
		void Promise.all([setAtlas(l, left), setAtlas(m, mid), setAtlas(r, right)]).then(([al, , ar]) => {
			if (!al || !ar) return;
			for (const [x, a] of [
				[l, al],
				[m, al],
				[r, ar],
			] as const) {
				x.style.top = '0px';
				x.style.height = `${a.height}px`;
			}
			l.style.left = `${dl}px`;
			l.style.width = `${al.width}px`;
			r.style.right = `${-dr}px`;
			r.style.width = `${ar.width}px`;
			m.style.left = `${dl + al.width}px`;
			m.style.right = `${-dr + ar.width}px`;
		});
	};
	set('', 'uiframe-tab-left', '_uiframe-tab-center', 'uiframe-tab-right', -3, 7);
	set('a', 'uiframe-activetab-left', '_uiframe-activetab-center', 'uiframe-activetab-right', -1, 8);
	const t = text(e, label, 'GameFontNormalSmall', undefined, 'center');
	t.classList.add('bz-ptab-text');
	if (box.w === undefined) {
		// Sized to the text once the font is in.
		requestAnimationFrame(() => (e.style.width = `${Math.max(t.scrollWidth + 40, 10)}px`));
	}
	clickable(e, () => {
		if (!e.classList.contains('selected')) onClick();
	});
	return e;
}
