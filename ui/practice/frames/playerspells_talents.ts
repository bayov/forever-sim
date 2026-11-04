// The Talents tab of the spells frame: Forever's ClassTalentsFrame (Blizzard_PlayerSpells
// ClassTalents with the Camelot files and Blizzard_SharedTalentUI). Forever draws the classic
// three trees side by side in Mainline's talent frame: the class's art behind them, a header
// per tree with its icon and the points spent, square talent buttons where the client's trait
// tree puts them, and arrows from each prerequisite.
//
// Read only, with the character's points from the request's talents string. Left out: the
// search box, Apply Changes and the reset and undo buttons (all for spending points), and the
// spec activation, since the room has one talent build. The Secondary tab shows locked.

import { iconUrl } from '../icons';
import { atlasInfo, setAtlas } from './atlas';
import type { RoomLink } from './link';
import { div, text } from './panels';
import { atlasVars } from './playerspells_book';
import { description, loadTrees, nodeCenter, parsePoints, TalentData, talentsString, treeIcons, TreeData } from './talent_trees';

// TalentButtonUtil.BaseVisualState, the ones a read only tree can show.
type VisualState = 'maxed' | 'selectable' | 'normal' | 'gated' | 'disabled';

// A tooltip's empty line (GameTooltip_AddBlankLineToTooltip). A plain space would fold away.
const blank = '\u00a0';

// The talent frame sits 4 above the bottom of the 1218 by 708 spells frame (1212 by 681).
const frameX = 3;
const frameY = 708 - 4 - 681;

export class TalentsView {
	private trees: TreeData[] = [];
	private points: number[][] = [];

	constructor(
		readonly link: RoomLink,
		readonly root: HTMLElement,
	) {
		root.classList.add('bz-tal');
		void this.build();
	}

	private async build() {
		const cls = this.link.loaded.class;
		const classFile = cls.replace(/^Class/, '').toLowerCase();
		void atlasVars(this.root, {
			'--tal-shadow': 'talents-node-square-shadow',
			'--tal-maxed': 'talents-node-square-yellow',
			'--tal-selectable': 'talents-node-square-green',
			'--tal-disabled': 'talents-node-square-gray',
			'--tal-locked': 'talents-node-square-locked',
			'--tal-line-yellow': 'talents-arrow-line-yellow',
			'--tal-line-gray': 'talents-arrow-line-gray',
			'--tal-line-locked': 'talents-arrow-line-locked',
			'--tal-head-yellow': 'talents-arrow-head-yellow',
			'--tal-head-gray': 'talents-arrow-head-gray',
			'--tal-head-locked': 'talents-arrow-head-locked',
			'--tal-checkmark': 'Talents-Checkmark-c60',
			'--tal-lock': 'Talents-lock-c60',
		});
		const f = div(this.root, 'bz-tal-frame', { x: frameX, y: frameY, w: 1212, h: 681 });

		// Background fills the frame and 20 above it, ClassBackground is the class's art from 70
		// below that to 36 above the bottom, and BackgroundBorder frames it, 2 and 4 outside.
		atlasBox(f, 'Talents-Background-c60', 0, -20, 1212, 701);
		atlasBox(f, `talent-background-${classFile}`, 0, 50, 1212, 595);
		const border = div(f, 'bz-tal-border', { x: -2, y: 46, w: 1216, h: 603 });
		void nineSliceAtlas(border, 'Talents-inner-frame-c60', 40);
		atlasBox(f, 'Talents-divider-left-c60', 606 - 591, 50, 591, 56);
		atlasBox(f, 'Talents-divider-right-c60', 606, 50, 595, 56);
		atlasBox(f, 'Talents-divider-vertical-c60', 404.5, 138, 2, 468);
		atlasBox(f, 'Talents-divider-vertical-c60', 807.5, 138, 2, 468);

		this.specTabs(f);

		this.trees = await loadTrees(cls);
		this.points = parsePoints(talentsString(this.link.loaded.request), this.trees);
		const spent = this.points.map(p => p.reduce((a, b) => a + b, 0));
		const unspent = Math.max(0, this.link.loaded.level - 9 - spent.reduce((a, b) => a + b, 0));
		this.currency(f, unspent);

		const edges = div(f, 'bz-tal-layer');
		const buttons = div(f, 'bz-tal-layer');
		this.trees.forEach((tree, i) => {
			this.treeHeader(f, i, tree.name, treeIcons[cls]?.[i], spent[i]);
			const points = this.points[i];
			const at = (r: number, c: number) => tree.talents.findIndex(t => t.location.rowIdx === r && t.location.colIdx === c);
			tree.talents.forEach((t, k) => {
				const rank = points[k] ?? 0;
				const pre = t.prereqLocation ? at(t.prereqLocation.rowIdx, t.prereqLocation.colIdx) : -1;
				const preTalent = pre >= 0 ? tree.talents[pre] : undefined;
				const preMet = !preTalent || (points[pre] ?? 0) >= preTalent.maxPoints;
				const need = t.location.rowIdx * 5;
				const state = visualState(t, rank, unspent, spent[i] >= need, preMet);
				this.button(buttons, i, t, rank, state, spent[i], need, preMet);
				if (preTalent)
					this.edge(edges, nodeCenter(i, preTalent.location), nodeCenter(i, t.location), preMet ? 'yellow' : state === 'gated' ? 'locked' : 'gray');
			});
		});
	}

	// treeHeader is ClassTalentTreeHeaderTemplate, centered at the BackgroundBorder's top-left
	// plus 140 + 400 per tree, 100 down: the tree's icon in its ring, the points spent in the box
	// under it, the name to its right and the divider below.
	private treeHeader(f: HTMLElement, i: number, name: string, icon: string | undefined, spent: number) {
		const h = div(f, 'bz-tal-header', { x: -2 + 140 + i * 400 - 24, y: 46 + 100 - 24, w: 48, h: 48 });
		const ic = div(h, 'bz-tal-header-icon', { x: 6 - 3, y: 6 - 3, w: 36, h: 36 });
		if (icon) ic.style.backgroundImage = `url(${iconUrl(icon)})`;
		div(h, 'bz-text bz-tal-header-name', { x: 54, y: 24 - 2 - 10 }, name);
		atlasBox(h, 'Talents-Main-Ring-c60', 21 - 27.5, 21 - 27.5, 55, 55);
		atlasBox(h, 'talents-main-ring-box-c60', 24 + 14 - 12, 24 + 12 - 10, 24, 20);
		text(h, String(spent), 'GameFontNormal', { x: 24 + 14 - 12, y: 24 + 12 - 10 + 3, w: 24 }, 'center');
		atlasBox(h, 'Talents-small-divider-c60', 24 + 60 - 177.5, 48 + 20 - 64, 355, 64);
	}

	// currency is ClassTalentCurrencyDisplayTemplate at the border's top-right: Unspent Talents and
	// the count in its box.
	private currency(f: HTMLElement, unspent: number) {
		const right = -2 + 1216 - 20;
		const top = 46 + 6;
		atlasBox(f, 'Talents-Square-Box-c60', right - 125, top, 125, 45);
		const label = text(f, 'Unspent Talents', 'GameFontHighlight', { y: top + 22.5 - 7 }, 'right');
		label.classList.add('bz-tal-unspent-label');
		label.style.right = `${1212 - (right - 125 + 60)}px`;
		const n = text(f, String(unspent), 'GameFontHighlight', { x: right - 4 - 48, y: top + 22.5 - 18, w: 48 }, 'center');
		n.classList.add('bz-tal-unspent');
	}

	// specTabs are the TabSystem's Primary and Secondary tabs over the border's top-left, 150
	// wide (minTabWidth). Primary is the active one. Secondary is locked like it is until you
	// learn Dual Specialization, since the room has no second build.
	private specTabs(f: HTMLElement) {
		const row = div(f, 'bz-tal-tabs', { x: -2 + 70, y: 46 + 4 - 32 });
		const primary = topTab(row, 'Primary', 0, true, false);
		primary.querySelector('.bz-tal-tab-text')!.append(markup('bz-tal-check'));
		const secondary = topTab(row, 'Secondary', 151, false, true);
		secondary.querySelector('.bz-tal-tab-text')!.append(markup('bz-tal-lock'));
		this.link.tooltip.attach(secondary, () => ({ kind: 'lines', title: 'Unlock at your class trainer', titleClass: 'bz-tal-error' }));
	}

	// button is ClassTalentButtonSquareTemplate (40 by 40) centered on its node: the icon, the
	// border for its state, its shadow and the points spent in the corner.
	private button(parent: HTMLElement, tree: number, t: TalentData, rank: number, state: VisualState, spent: number, need: number, preMet: boolean) {
		const c = nodeCenter(tree, t.location);
		const b = div(parent, `bz-tal-btn ${state}`, { x: c.x - 20, y: c.y - 20, w: 40, h: 40 });
		div(b, 'bz-tal-shadow');
		const icon = div(b, 'bz-tal-icon');
		icon.style.backgroundImage = `url(${iconUrl(t.icon || 'inv_misc_questionmark')})`;
		div(b, 'bz-tal-dim');
		div(b, 'bz-tal-state');
		div(b, 'bz-tal-state hover');
		const showRank = rank > 0 || state === 'selectable';
		if (showRank) div(b, 'bz-tal-spend', undefined, String(rank));

		this.link.tooltip.attach(b, () => {
			// The game's talent tooltip: the name, the rank, the rank's text, the next rank's text
			// once you have a point, then what stops you, in red.
			const lines: string[] = [`Rank ${rank}/${t.maxPoints}`, blank, description(t, Math.max(rank, 1))];
			const marks = ['bz-tt-talent', 'w2'];
			if (rank > 0 && rank < t.maxPoints) {
				lines.push(blank, 'Next Rank:');
				marks.push(`w${lines.length + 1}`);
				lines.push(description(t, rank + 1));
			}
			if (spent < need) {
				const more = need - spent;
				lines.push(blank, `Spend ${more} more point${more === 1 ? '' : 's'} to unlock this row`);
				marks.push(`r${lines.length + 1}`);
			} else if (!preMet) {
				lines.push(blank, 'Requires all preceding talents');
				marks.push(`r${lines.length + 1}`);
			}
			return { kind: 'lines', title: t.name ?? '', lines, titleClass: marks.join(' ') };
		});
	}

	// edge is ClassTalentEdgeArrowTemplate: a 6 thick line from the prerequisite's center to 1.2
	// half widths short of the talent's center, where the arrow head points at it.
	private edge(parent: HTMLElement, from: { x: number; y: number }, to: { x: number; y: number }, kind: 'yellow' | 'gray' | 'locked') {
		const dx = from.x - to.x;
		const dy = from.y - to.y;
		const len = Math.hypot(dx, dy);
		const end = { x: to.x + (dx / len) * 24, y: to.y + (dy / len) * 24 };
		const lineLen = Math.hypot(end.x - from.x, end.y - from.y);
		const angle = Math.atan2(end.y - from.y, end.x - from.x);
		const line = div(parent, `bz-tal-line ${kind}`, { x: from.x, y: from.y - 3, w: lineLen, h: 6 });
		line.style.transform = `rotate(${angle}rad)`;
		const head = div(parent, `bz-tal-head ${kind}`, { x: end.x - 7, y: end.y - 6, w: 14, h: 12 });
		// The head's art points down, and turns with the line.
		head.style.transform = `rotate(${angle - Math.PI / 2}rad)`;
	}
}

// visualState is TalentButtonBaseMixin:CalculateVisualState for a talent we can only look at.
function visualState(t: TalentData, rank: number, unspent: number, rowOpen: boolean, preMet: boolean): VisualState {
	if (rank >= t.maxPoints) return 'maxed';
	if (unspent > 0 && rowOpen && preMet) return 'selectable';
	if (rank > 0) return 'normal';
	if (!rowOpen) return 'gated';
	return 'disabled';
}

// topTab is a ClassTalentsFrameTabTemplate: a TabSystem tab hanging over the frame (its art turned
// upside down), GameFontNormal or GameFontHighlight when selected, 4 under the middle.
function topTab(parent: HTMLElement, label: string, x: number, selected: boolean, disabled: boolean): HTMLElement {
	const e = div(parent, `bz-tal-tab${selected ? ' selected' : ''}${disabled ? ' disabled' : ''}`, { x, y: 0, w: 150, h: 32 });
	const art = div(e, 'bz-tal-tab-art');
	const pieces = selected
		? (['uiframe-activetab-left', '_uiframe-activetab-center', 'uiframe-activetab-right', 7] as const)
		: (['uiframe-tab-left', '_uiframe-tab-center', 'uiframe-tab-right', 6] as const);
	const [left, mid, right, overhang] = pieces;
	// Drawn the bottom tab's way, then the whole art turned half around (SetRotation(math.pi)).
	void Promise.all([atlasInfo(left), atlasInfo(right)]).then(([al, ar]) => {
		if (!al || !ar) return;
		const l = div(art, 'bz-tex', { x: 0, y: 0, w: al.width, h: 32 });
		const r = div(art, 'bz-tex', { x: 150 + overhang - ar.width, y: 0, w: ar.width, h: 32 });
		const m = div(art, 'bz-tex', { x: al.width, y: 0, w: 150 + overhang - ar.width - al.width, h: 32 });
		void setAtlas(l, left);
		void setAtlas(m, mid);
		void setAtlas(r, right);
	});
	const t = div(e, `bz-tal-tab-text ${selected ? 'GameFontHighlight' : disabled ? 'GameFontDisable' : 'GameFontNormal'}`, undefined, label);
	t.classList.add('bz-text');
	return e;
}

function markup(cls: string): HTMLElement {
	const m = document.createElement('span');
	m.className = cls;
	return m;
}

// atlasBox draws an atlas stretched over a box.
function atlasBox(parent: HTMLElement, name: string, x: number, y: number, w: number, h: number): HTMLElement {
	const e = div(parent, 'bz-tex', { x, y, w, h });
	void setAtlas(e, name);
	return e;
}

// nineSliceAtlas draws an atlas the client slices (UiTextureAtlasElementSliceData): the corners
// stay their size, `margin` units each, and the rest stretches.
async function nineSliceAtlas(e: HTMLElement, name: string, margin: number) {
	const a = await atlasInfo(name);
	if (!a) return;
	await setAtlas(e, name);
	const u = e.style.backgroundImage;
	e.style.backgroundImage = '';
	const px = Math.round((margin * (a.right - a.left)) / a.width);
	e.style.borderStyle = 'solid';
	e.style.borderWidth = `${margin}px`;
	e.style.borderImage = `${u} ${px} fill / ${margin}px stretch`;
}
