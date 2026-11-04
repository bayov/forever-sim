// The talent frame (SHIFT+K): the class's three trees with your talents, read only. It's the
// Classic Era TalentFrame (Blizzard_TalentUI): the frame art, the tree's background in a scroll
// frame, the talent buttons on the game's grid with their prerequisite lines, and a tab per
// tree.

import './talents.css';
import { iconUrl } from '../icons';
import { SND, sound, tex, texVars, TX } from './assets';
import type { RoomLink } from './link';
import { portrait } from './model_view';
import { closeButton, div, Panel, scrollFrame, selectTab, setTexture, tab, text, texture } from './panels';
import {
	arrowCoords,
	background,
	branchCoords,
	BranchNode,
	classCircleCoords,
	COLUMNS,
	description,
	drawLines,
	loadTrees,
	MAX_TIERS,
	newBranches,
	parsePoints,
	pick,
	talentsString,
	TreeData,
} from './talent_trees';

// Blizzard_TalentUI's layout numbers.
const OFFSET_X = 35;
const OFFSET_Y = 20;
const STEP = 63;
const BUTTON = 37;
// interface/paperdollinfoframe/ui-character-scrollbar.
const characterScrollBar = 136569;
// TALENT_BUTTON_SIZE, the branch and arrow pieces' size.
const PIECE = 32;

type Coords = [number, number, number, number];

export class TalentFrame extends Panel {
	readonly width = 384;
	readonly height = 512;
	stackWidth = 353;
	openSound = SND.talentOpen;
	closeSound = SND.talentClose;

	private trees: TreeData[] = [];
	private points: number[][] = [];
	private selected = 0;
	private tabs: HTMLElement[] = [];
	private bg: HTMLElement[] = [];
	private spent?: HTMLElement;
	private unspent?: HTMLElement;
	private scroll?: ReturnType<typeof scrollFrame>;

	constructor(readonly link: RoomLink) {
		super();
	}

	protected async build(root: HTMLElement) {
		root.classList.add('bz-talent');
		const cls = this.link.loaded.class;
		const [trees] = await Promise.all([loadTrees(cls), this.loadArt()]);
		this.trees = trees;
		this.points = parsePoints(talentsString(this.link.loaded.request), trees);

		// The frame takes the mouse only over its art, like the game's HitRectInsets.
		div(root, 'bz-talent-hit', { x: 0, y: 0, w: 354, h: 437 });

		// The portrait: the class icon until the room draws the character's face there.
		const portrait = div(root, 'bz-portrait', { x: 7, y: 6, w: 60, h: 60 });
		const circle = classCircleCoords[cls];
		if (circle) texture(portrait, TX.classCircles, { x: 0, y: 0, w: 60, h: 60 }, { coords: circle, size: [256, 256] });

		// The tree's background sits over the frame art, which is dark where the tree shows.
		texture(root, TX.generalTopLeft, { x: 2, y: 1, w: 256, h: 256 });
		texture(root, TX.generalTopRight, { x: 258, y: 1, w: 128, h: 256 });
		texture(root, TX.talentBotLeft, { x: 2, y: 257, w: 256, h: 256 });
		texture(root, TX.talentBotRight, { x: 258, y: 257, w: 128, h: 256 });
		this.bg = [
			div(root, 'bz-tex', { x: 23, y: 77, w: 256, h: 256 }),
			div(root, 'bz-tex', { x: 279, y: 77, w: 64, h: 256 }),
			div(root, 'bz-tex', { x: 23, y: 333, w: 256, h: 128 }),
			div(root, 'bz-tex', { x: 279, y: 333, w: 64, h: 128 }),
		];

		text(root, 'Talents', 'GameFontNormal', { x: 70, y: 17, w: 256 }, 'center');
		this.spent = text(root, '', 'GameFontNormalSmall', { x: 70, y: 47, w: 256 }, 'center');

		// The scroll bar's dark track, from UI-Character-ScrollBar.
		texture(root, characterScrollBar, { x: 317, y: 72, w: 31, h: 256 }, { coords: [0, 0.484375, 0, 1], size: [64, 256] });
		texture(root, characterScrollBar, { x: 317, y: 305, w: 31, h: 106 }, { coords: [0.515625, 1, 0, 0.4140625], size: [64, 256] });
		this.scroll = scrollFrame(root, { x: 23, y: 77, w: 296, h: 332 });
		this.scroll.el.classList.add('bz-talent-scroll');

		this.unspent = text(root, '', 'GameFontNormalSmall', { x: 30, y: 415 });

		closeButton(root, { x: 324, y: 9 }, () => this.close());

		let x = 15;
		this.tabs = trees.map((tree, i) => {
			const w = Math.ceil(measure(tree.name)) + 40;
			const t = tab(root, tree.name, { x, y: 434, w }, () => {
				sound(SND.characterTab);
				this.select(i);
			});
			x += w - 15;
			return t;
		});
		this.select(this.selected);
	}

	protected shown() {
		this.scroll?.update();
		// The character's face over the class icon, once the models are in.
		const face = this.el?.querySelector<HTMLElement>('.bz-portrait');
		if (face) portrait(face, this.link);
	}

	// loadArt makes the textures the talent buttons and lines use: the slot tinted the game's
	// three ways, the rank border, the branches and the arrows.
	private loadArt(): Promise<void> {
		return texVars({
			'bz-tal-slot-green': () => tinted(TX.emptySlotWhite, [0.1, 1, 0.1]),
			'bz-tal-slot-gold': () => tinted(TX.emptySlotWhite, [1, 0.82, 0]),
			'bz-tal-slot-gray': () => tinted(TX.emptySlotWhite, [0.5, 0.5, 0.5]),
			'bz-tal-rank': () => tex(TX.talentRankBorder),
			'bz-tal-rank-gray': () => tinted(TX.talentRankBorder, [0.5, 0.5, 0.5]),
			'bz-tal-branches': () => tex(TX.talentBranches),
			'bz-tal-arrows': () => tex(TX.talentArrows),
		});
	}

	private select(i: number) {
		this.selected = i;
		selectTab(this.tabs, i);
		const tree = this.trees[i];
		if (!tree) return;
		const files = background(this.link.loaded.class, tree.name);
		this.bg.forEach((e, k) => setTexture(e, files[k]));
		const points = this.points[i];
		const treeSpent = points.reduce((a, b) => a + b, 0);
		const total = this.points.flat().reduce((a, b) => a + b, 0);
		const unspent = Math.max(0, this.link.loaded.level - 9 - total);

		this.spent!.replaceChildren(`Points spent in ${tree.name} Talents: `, white(String(treeSpent)));
		this.unspent!.replaceChildren('Talent Points: ', white(String(unspent)));

		const content = this.scroll!.content;
		content.replaceChildren();
		content.style.height = `${OFFSET_Y + (MAX_TIERS - 1) * STEP + BUTTON + OFFSET_Y}px`;
		const branchLayer = div(content, 'bz-tal-layer');
		const buttonLayer = div(content, 'bz-tal-layer');
		const arrowLayer = div(content, 'bz-tal-layer');

		// TalentFrame_Update: each talent's state, and the lines to it from its prerequisite.
		const branches = newBranches();
		const at = (r: number, c: number) => tree.talents.findIndex(t => t.location.rowIdx === r && t.location.colIdx === c);
		tree.talents.forEach((t, k) => {
			const tier = t.location.rowIdx + 1;
			const column = t.location.colIdx + 1;
			const rank = points[k] ?? 0;
			branches[tier][column].id = k;
			const tierUnlocked = t.location.rowIdx * 5 <= treeSpent;
			const forceDesaturated = unspent <= 0 && rank === 0;
			let met = tierUnlocked && !forceDesaturated;
			let prereq: { name: string; max: number; met: boolean } | undefined;
			if (t.prereqLocation) {
				const p = at(t.prereqLocation.rowIdx, t.prereqLocation.colIdx);
				const pt = tree.talents[p];
				const learnable = pt !== undefined && (points[p] ?? 0) >= pt.maxPoints;
				if (!learnable || forceDesaturated) met = false;
				drawLines(branches, tier, column, t.prereqLocation.rowIdx + 1, t.prereqLocation.colIdx + 1, met);
				if (pt) prereq = { name: pt.name ?? '', max: pt.maxPoints, met: learnable };
			}

			const b = div(buttonLayer, 'bz-tal-btn', { x: (column - 1) * STEP + OFFSET_X, y: (tier - 1) * STEP + OFFSET_Y, w: BUTTON, h: BUTTON });
			div(b, 'bz-tal-slot');
			const icon = div(b, 'bz-tal-icon');
			icon.style.backgroundImage = `url(${iconUrl(t.icon || 'inv_misc_questionmark')})`;
			div(b, 'bz-tal-rank-border');
			text(b, String(rank), 'GameFontNormalSmall').classList.add('bz-tal-rank');
			const state = met ? (rank < t.maxPoints ? 'green' : 'gold') : rank === 0 ? 'off' : 'gray';
			b.classList.add(state);

			this.link.tooltip.attach(b, () => {
				const lines = [`Rank ${rank}/${t.maxPoints}`];
				const marks = ['bz-tal-tt', 'w2'];
				if (!tierUnlocked) {
					lines.push(`Requires ${t.location.rowIdx * 5} points in ${tree.name} Talents`);
					marks.push(`r${lines.length + 1}`);
				}
				if (prereq && !prereq.met) {
					lines.push(`Requires ${prereq.max} point${prereq.max === 1 ? '' : 's'} in ${prereq.name}`);
					marks.push(`r${lines.length + 1}`);
				}
				lines.push(description(t, rank));
				if (rank > 0 && rank < t.maxPoints) {
					lines.push(' ', 'Next rank:');
					marks.push(`w${lines.length + 1}`);
					lines.push(description(t, rank + 1));
				}
				return { kind: 'lines', title: t.name ?? '', lines, titleClass: marks.join(' ') };
			});
		});
		this.drawBranches(branches, branchLayer, arrowLayer);
		this.scroll!.scrollTo(0);
	}

	// drawBranches lays the branch pieces and arrows, as TalentFrame_DrawLines.
	private drawBranches(a: BranchNode[][], branchLayer: HTMLElement, arrowLayer: HTMLElement) {
		const branch = (c: Coords, x: number, y: number) => piece(branchLayer, 'bz-tal-branch', c, x, -y);
		const arrow = (c: Coords, x: number, y: number) => piece(arrowLayer, 'bz-tal-arrow', c, x, -y);
		const C = branchCoords;
		let ignoreUp = false;
		for (let i = 1; i <= MAX_TIERS; i++) {
			for (let j = 1; j <= COLUMNS; j++) {
				const node = a[i][j];
				// Offsets in the game's terms: y goes up, so we flip it when placing.
				const x = (j - 1) * STEP + OFFSET_X + 2;
				const y = -((i - 1) * STEP) - OFFSET_Y - 2;
				if (node.id !== undefined) {
					if (node.up) {
						if (!ignoreUp) branch(pick(C.up, node.up), x, y + PIECE);
						else ignoreUp = false;
					}
					if (node.down) branch(pick(C.down, node.down), x, y - PIECE + 1);
					if (node.left) branch(pick(C.left, node.left), x - PIECE, y);
					if (node.right) {
						const next = a[i][j + 1];
						if (next.left && next.down < 0) branch(pick(C.right, next.down), x + PIECE, y);
						else branch(pick(C.right, node.right), x + PIECE + 1, y);
					}
					if (node.rightArrow) arrow(pick(arrowCoords.right, node.rightArrow), x + PIECE / 2 + 5, y);
					if (node.leftArrow) arrow(pick(arrowCoords.left, node.leftArrow), x - PIECE / 2 - 5, y);
					if (node.topArrow) arrow(pick(arrowCoords.top, node.topArrow), x, y + PIECE / 2 + 5);
				} else if (node.up && node.left && node.right) {
					branch(pick(C.tup, node.up), x, y);
				} else if (node.down && node.left && node.right) {
					branch(pick(C.tdown, node.down), x, y);
				} else if (node.left && node.down) {
					branch(pick(C.topright, node.left), x, y);
					branch(pick(C.down, node.down), x, y - 32);
				} else if (node.left && node.up) {
					branch(pick(C.bottomright, node.left), x, y);
				} else if (node.left && node.right) {
					branch(pick(C.right, node.right), x + PIECE, y);
					branch(pick(C.left, node.left), x + 1, y);
				} else if (node.right && node.down) {
					branch(pick(C.topleft, node.right), x, y);
					branch(pick(C.down, node.down), x, y - 32);
				} else if (node.right && node.up) {
					branch(pick(C.bottomleft, node.right), x, y);
				} else if (node.up && node.down) {
					branch(pick(C.up, node.up), x, y);
					branch(pick(C.down, node.down), x, y - 32);
					ignoreUp = true;
				}
			}
		}
	}
}

// piece draws a 32 by 32 cut of a sprite sheet (the branches or the arrows), mirrored when the
// coords run right to left like the game's.
function piece(parent: HTMLElement, cls: string, c: Coords, x: number, y: number) {
	const [l, r, t, b] = c;
	const e = div(parent, cls, { x, y, w: PIECE, h: PIECE });
	const bw = PIECE / Math.abs(r - l);
	const bh = PIECE / Math.abs(b - t);
	e.style.backgroundSize = `${bw}px ${bh}px`;
	e.style.backgroundPosition = `${-Math.min(l, r) * bw}px ${-Math.min(t, b) * bh}px`;
	if (l > r) e.style.transform = 'scaleX(-1)';
}

function white(s: string): HTMLElement {
	const e = document.createElement('span');
	e.className = 'bz-white';
	e.textContent = s;
	return e;
}

// measure gives a tab label's width in the tab font.
function measure(s: string): number {
	const c = document.createElement('canvas').getContext('2d')!;
	c.font = "10px 'Friz Quadrata'";
	return c.measureText(s).width;
}

// tinted is a texture multiplied by a color, like a FrameXML SetVertexColor.
async function tinted(fdid: number, rgb: [number, number, number]): Promise<string> {
	const img = new Image();
	img.src = await tex(fdid);
	await img.decode();
	const c = document.createElement('canvas');
	c.width = img.width;
	c.height = img.height;
	const g = c.getContext('2d')!;
	g.drawImage(img, 0, 0);
	g.globalCompositeOperation = 'multiply';
	g.fillStyle = `rgb(${rgb.map(v => Math.round(v * 255)).join(',')})`;
	g.fillRect(0, 0, c.width, c.height);
	g.globalCompositeOperation = 'destination-in';
	g.drawImage(img, 0, 0);
	return c.toDataURL();
}
