// The client's texture atlases, the way Forever's frames name their art ("UI-Frame-Metal-
// CornerTopLeft", "spellbook-Page-Left-C60"). The practice server reads them from the client's
// UiTextureAtlas tables (cmd/practice/atlas.go). We ask for the names we draw in batches and cut
// each region out of its file once.

import { tex } from './assets';
import { Box, div, place } from './panels';

export type AtlasInfo = {
	file: number;
	// The size the frames lay it out at (half the pixels for art made at twice the density).
	width: number;
	height: number;
	// The region of the file, in its pixels.
	left: number;
	right: number;
	top: number;
	bottom: number;
	fileWidth: number;
	fileHeight: number;
	// 1 tiles across, 2 tiles down.
	tile?: number;
};

const infos = new Map<string, Promise<AtlasInfo | undefined>>();
let pending: { names: Set<string>; done: Promise<Record<string, AtlasInfo>> } | undefined;

// atlasInfo looks an atlas up. Calls in the same moment go to the server as one request.
export function atlasInfo(name: string): Promise<AtlasInfo | undefined> {
	const key = name.toLowerCase();
	let p = infos.get(key);
	if (!p) {
		if (!pending) {
			const names = new Set<string>();
			const done = Promise.resolve().then(async () => {
				pending = undefined;
				const r = await fetch(`/api/atlas?names=${encodeURIComponent([...names].join(','))}`);
				return r.ok ? ((await r.json()) as Record<string, AtlasInfo>) : {};
			});
			pending = { names, done };
		}
		pending.names.add(name);
		p = pending.done.then(all => all[name]).catch(() => undefined);
		p.then(a => !a && console.warn(`no atlas ${name}`));
		infos.set(key, p);
	}
	return p;
}

// atlasUrl gives an atlas's region as a data URL.
export async function atlasUrl(name: string): Promise<string | undefined> {
	const a = await atlasInfo(name);
	if (!a) return undefined;
	return tex(a.file, [a.left, a.top, a.right - a.left, a.bottom - a.top]);
}

export type AtlasOptions = {
	// Size the element to the atlas's own size (useAtlasSize="true").
	useAtlasSize?: boolean;
	// Mirror it (SetTexCoord flips), across or down.
	flipX?: boolean;
	flipY?: boolean;
	cls?: string;
};

// setAtlas draws an atlas into an element: stretched to the element, or tiled at its own size
// when the atlas tiles.
export async function setAtlas(e: HTMLElement, name: string, opts: AtlasOptions = {}): Promise<AtlasInfo | undefined> {
	const [a, u] = await Promise.all([atlasInfo(name), atlasUrl(name)]);
	if (!a || !u) return undefined;
	if (opts.useAtlasSize) place(e, { w: a.width, h: a.height });
	e.style.backgroundImage = `url(${u})`;
	e.style.backgroundRepeat = 'no-repeat';
	if (a.tile) {
		e.style.backgroundRepeat = a.tile === 3 ? 'repeat' : a.tile === 1 ? 'repeat-x' : 'repeat-y';
		e.style.backgroundSize = `${a.tile & 1 ? `${a.width}px` : '100%'} ${a.tile & 2 ? `${a.height}px` : '100%'}`;
	} else e.style.backgroundSize = '100% 100%';
	if (opts.flipX || opts.flipY) e.style.transform = `scale(${opts.flipX ? -1 : 1}, ${opts.flipY ? -1 : 1})`;
	return a;
}

// atlasTexture is a FrameXML <Texture atlas="...">: an element at a spot drawing the atlas.
export function atlasTexture(parent: HTMLElement, name: string, box: Box = {}, opts: AtlasOptions = {}): HTMLDivElement {
	const e = div(parent, `bz-tex ${opts.cls ?? ''}`, box);
	void setAtlas(e, name, opts);
	return e;
}

// preloadAtlases asks for many atlases at once, ahead of drawing them.
export function preloadAtlases(names: string[]): Promise<unknown> {
	return Promise.all(names.map(n => atlasUrl(n)));
}

// ---- NineSlice ----

type Piece = { atlas: string; x?: number; y?: number; x1?: number; y1?: number; mirrorH?: boolean; mirrorV?: boolean };
export type NineSliceLayout = {
	TopLeftCorner?: Piece;
	TopRightCorner?: Piece;
	BottomLeftCorner?: Piece;
	BottomRightCorner?: Piece;
	TopEdge?: Piece;
	BottomEdge?: Piece;
	LeftEdge?: Piece;
	RightEdge?: Piece;
	Center?: Piece;
};

// The layouts of Blizzard_SharedXML/Mainline/NineSliceLayouts.lua the frames use, with Forever's
// offset fixes from Camelot/NineSliceLayoutOverrides.lua already applied (the top right corner
// 2 to the left, the bottom corners at y -8).
export const nineSliceLayouts: Record<string, NineSliceLayout> = {
	PortraitFrameTemplate: {
		TopLeftCorner: { atlas: 'UI-Frame-PortraitMetal-CornerTopLeft', x: -13, y: 16 },
		TopRightCorner: { atlas: 'UI-Frame-Metal-CornerTopRight', x: 2, y: 16 },
		BottomLeftCorner: { atlas: 'UI-Frame-Metal-CornerBottomLeft', x: -13, y: -8 },
		BottomRightCorner: { atlas: 'UI-Frame-Metal-CornerBottomRight', x: 2, y: -8 },
		TopEdge: { atlas: '_UI-Frame-Metal-EdgeTop' },
		BottomEdge: { atlas: '_UI-Frame-Metal-EdgeBottom' },
		LeftEdge: { atlas: '!UI-Frame-Metal-EdgeLeft' },
		RightEdge: { atlas: '!UI-Frame-Metal-EdgeRight' },
	},
	ButtonFrameTemplateNoPortrait: {
		TopLeftCorner: { atlas: 'UI-Frame-Metal-CornerTopLeft', x: -8, y: 16 },
		TopRightCorner: { atlas: 'UI-Frame-Metal-CornerTopRight', x: 2, y: 16 },
		BottomLeftCorner: { atlas: 'UI-Frame-Metal-CornerBottomLeft', x: -8, y: -8 },
		BottomRightCorner: { atlas: 'UI-Frame-Metal-CornerBottomRight', x: 2, y: -8 },
		TopEdge: { atlas: '_UI-Frame-Metal-EdgeTop' },
		BottomEdge: { atlas: '_UI-Frame-Metal-EdgeBottom' },
		LeftEdge: { atlas: '!UI-Frame-Metal-EdgeLeft' },
		RightEdge: { atlas: '!UI-Frame-Metal-EdgeRight' },
	},
};

// nineSlice draws a NineSlice layout around an element, like NineSliceUtil.ApplyLayout: the
// corners at their atlas sizes, offset from the frame's corners (x right and y UP, like the
// game), and the edges between the corners' inner sides, tiled. An edge spans the corners'
// height (top and bottom) or width (left and right), as the game anchors it.
export function nineSlice(root: HTMLElement, layout: NineSliceLayout | string): HTMLDivElement {
	const l = typeof layout === 'string' ? nineSliceLayouts[layout] : layout;
	const box = div(root, 'bz-nineslice');
	void (async () => {
		const pieces = [l.TopLeftCorner, l.TopRightCorner, l.BottomLeftCorner, l.BottomRightCorner, l.TopEdge, l.BottomEdge, l.LeftEdge, l.RightEdge];
		const [tl, tr, bl, br] = await Promise.all(pieces.slice(0, 4).map(p => (p ? atlasInfo(p.atlas) : undefined)));
		const piece = (p: Piece | undefined, css: Partial<Record<'left' | 'right' | 'top' | 'bottom' | 'width' | 'height', number>>) => {
			if (!p) return;
			const e = div(box, 'bz-tex');
			e.style.position = 'absolute';
			for (const [k, v] of Object.entries(css)) e.style.setProperty(k, `${v}px`);
			void setAtlas(e, p.atlas, { flipX: p.mirrorH, flipY: p.mirrorV });
		};
		const off = (p?: Piece) => ({ x: p?.x ?? 0, y: p?.y ?? 0 });
		const [o1, o2, o3, o4] = [off(l.TopLeftCorner), off(l.TopRightCorner), off(l.BottomLeftCorner), off(l.BottomRightCorner)];
		const size = (a?: AtlasInfo) => ({ w: a?.width ?? 0, h: a?.height ?? 0 });
		const [s1, s2, s3, s4] = [size(tl), size(tr), size(bl), size(br)];
		piece(l.TopLeftCorner, { left: o1.x, top: -o1.y, width: s1.w, height: s1.h });
		piece(l.TopRightCorner, { right: -o2.x, top: -o2.y, width: s2.w, height: s2.h });
		piece(l.BottomLeftCorner, { left: o3.x, bottom: o3.y, width: s3.w, height: s3.h });
		piece(l.BottomRightCorner, { right: -o4.x, bottom: o4.y, width: s4.w, height: s4.h });
		// The edges reach 1 into the corners, so no seam shows at a fractional UI scale.
		piece(l.TopEdge, { left: o1.x + s1.w - 1, right: -o2.x + s2.w - 1, top: -o1.y, height: s1.h });
		piece(l.BottomEdge, { left: o3.x + s3.w - 1, right: -o4.x + s4.w - 1, bottom: o3.y, height: s3.h });
		piece(l.LeftEdge, { left: o1.x, width: s1.w, top: -o1.y + s1.h - 1, bottom: o3.y + s3.h - 1 });
		piece(l.RightEdge, { right: -o2.x, width: s2.w, top: -o2.y + s2.h - 1, bottom: o4.y + s4.h - 1 });
	})();
	return box;
}
