// The talent trees the talent frame draws: the sim's Forever trees (ui/core/talents/trees), the
// character's points from the request's talents string, and the game's art for each tree.

import { talentBackgrounds } from './assets';

export type TalentLoc = { rowIdx: number; colIdx: number };

export type TalentData = {
	name?: string;
	location: TalentLoc;
	maxPoints: number;
	description?: string;
	ranks?: (number | string)[][];
	icon?: string;
	spellIds?: number[];
	prereqLocation?: TalentLoc;
};

export type TreeData = { name: string; talents: TalentData[] };

// Each class's trees load on the talent frame's first open, so the room doesn't carry all nine.
const loaders: Record<string, () => Promise<{ default: unknown }>> = {
	ClassDruid: () => import('../../core/talents/trees/druid.json'),
	ClassHunter: () => import('../../core/talents/trees/hunter.json'),
	ClassMage: () => import('../../core/talents/trees/mage.json'),
	ClassPaladin: () => import('../../core/talents/trees/paladin.json'),
	ClassPriest: () => import('../../core/talents/trees/priest.json'),
	ClassRogue: () => import('../../core/talents/trees/rogue.json'),
	ClassShaman: () => import('../../core/talents/trees/shaman.json'),
	ClassWarlock: () => import('../../core/talents/trees/warlock.json'),
	ClassWarrior: () => import('../../core/talents/trees/warrior.json'),
};

export async function loadTrees(cls: string): Promise<TreeData[]> {
	const load = loaders[cls];
	if (!load) return [];
	const m = await load();
	return m.default as TreeData[];
}

// parsePoints reads a talents string ("2550031-05...-...", one digit per talent in tree order,
// trees split by '-') into points per talent of each tree.
export function parsePoints(talents: string, trees: TreeData[]): number[][] {
	const parts = talents.split('-');
	return trees.map((tree, i) => tree.talents.map((_, j) => parseInt(parts[i]?.[j] ?? '0', 10) || 0));
}

// talentsString finds the talents string in the RaidSimRequest (protojson).
export function talentsString(request: unknown): string {
	const r = request as { raid?: { parties?: { players?: { talentsString?: string }[] }[] } } | undefined;
	return r?.raid?.parties?.[0]?.players?.[0]?.talentsString ?? '';
}

// The client's file names for the trees whose name differs from the sim's tree name.
const bgNames: Record<string, string> = {
	Elemental: 'ElementalCombat',
	Retribution: 'Combat',
	'Shadow Magic': 'Shadow',
	Affliction: 'Curses',
	Demonology: 'Summoning',
};

// background gives a tree's background textures (top-left, top-right, bottom-left,
// bottom-right). The listfile has each tree's four files in the order bottomleft, bottomright,
// topleft, topright, so they sit around the top-left's id. Like the game, Mage Fire is the
// fallback.
export function background(cls: string, tree: string): [number, number, number, number] {
	const name = cls.replace(/^Class/, '') + (bgNames[tree] ?? tree.replace(/\s+/g, ''));
	const tl = talentBackgrounds[name] ?? talentBackgrounds.MageFire;
	return [tl, tl + 1, tl - 2, tl - 1];
}

// description fills a rank's values into the talent's {n} placeholders.
export function description(t: TalentData, rank: number): string {
	const values = t.ranks?.[Math.max(0, rank - 1)] ?? t.ranks?.[0] ?? [];
	return (t.description ?? '').replace(/\{(\d+)\}/g, (_, i) => {
		const v = values[Number(i)];
		return v === undefined ? '?' : String(v);
	});
}

// The class icons in interface/targetingframe/ui-classes-circles (CLASS_ICON_TCOORDS).
export const classCircleCoords: Record<string, [number, number, number, number]> = {
	ClassWarrior: [0, 0.25, 0, 0.25],
	ClassMage: [0.25, 0.49609375, 0, 0.25],
	ClassRogue: [0.49609375, 0.7421875, 0, 0.25],
	ClassDruid: [0.7421875, 0.98828125, 0, 0.25],
	ClassHunter: [0, 0.25, 0.25, 0.5],
	ClassShaman: [0.25, 0.49609375, 0.25, 0.5],
	ClassPriest: [0.49609375, 0.7421875, 0.25, 0.5],
	ClassWarlock: [0.7421875, 0.98828125, 0.25, 0.5],
	ClassPaladin: [0, 0.25, 0.5, 0.75],
};

// The prerequisite lines between talents, as Blizzard_TalentUI's TALENT_BRANCH_ARRAY: per tier
// and column, which branch pieces and arrows go there. 1 is lit (the prerequisite is met), -1
// gray, 0 none.
export type BranchNode = {
	id?: number;
	up: number;
	down: number;
	left: number;
	right: number;
	leftArrow: number;
	rightArrow: number;
	topArrow: number;
};

export const MAX_TIERS = 7;
export const COLUMNS = 4;

export function newBranches(): BranchNode[][] {
	const a: BranchNode[][] = [];
	for (let i = 1; i <= MAX_TIERS + 1; i++) {
		a[i] = [];
		for (let j = 1; j <= COLUMNS + 1; j++) a[i][j] = { up: 0, down: 0, left: 0, right: 0, leftArrow: 0, rightArrow: 0, topArrow: 0 };
	}
	return a;
}

// drawLines marks the path from a prerequisite (tier, column) to the talent at (buttonTier,
// buttonColumn), 1-based like the game. It's TalentFrame_DrawLines.
export function drawLines(a: BranchNode[][], buttonTier: number, buttonColumn: number, tier: number, column: number, met: boolean) {
	const m = met ? 1 : -1;
	if (buttonColumn === column) {
		for (let i = tier + 1; i <= buttonTier - 1; i++) if (a[i][buttonColumn].id !== undefined) return;
		for (let i = tier; i <= buttonTier - 1; i++) {
			a[i][buttonColumn].down = m;
			if (i + 1 <= buttonTier - 1) a[i + 1][buttonColumn].up = m;
		}
		a[buttonTier][buttonColumn].topArrow = m;
		return;
	}
	if (buttonTier === tier) {
		const left = Math.min(buttonColumn, column);
		const right = Math.max(buttonColumn, column);
		for (let i = left + 1; i <= right - 1; i++) if (a[tier][i].id !== undefined) return;
		for (let i = left; i <= right - 1; i++) {
			a[tier][i].right = m;
			a[tier][i + 1].left = m;
		}
		if (buttonColumn < column) a[buttonTier][buttonColumn].rightArrow = m;
		else a[buttonTier][buttonColumn].leftArrow = m;
		return;
	}
	// The prerequisite is diagonal from us. We go across its tier first, then down.
	let left = Math.min(buttonColumn, column);
	let right = Math.max(buttonColumn, column);
	if (left === column) left++;
	else right--;
	let blocked = false;
	for (let i = left; i <= right; i++) if (a[tier][i].id !== undefined) blocked = true;
	left = Math.min(buttonColumn, column);
	right = Math.max(buttonColumn, column);
	if (!blocked) {
		a[tier][buttonColumn].down = m;
		a[buttonTier][buttonColumn].up = m;
		for (let i = tier; i <= buttonTier - 1; i++) {
			a[i][buttonColumn].down = m;
			a[i + 1][buttonColumn].up = m;
		}
		for (let i = left; i <= right - 1; i++) {
			a[tier][i].right = m;
			a[tier][i + 1].left = m;
		}
		a[buttonTier][buttonColumn].topArrow = m;
		return;
	}
	// Blocked going across first, so we go down the prerequisite's column, then across.
	if (left === buttonColumn) left++;
	else right--;
	for (let i = left; i <= right; i++) if (a[buttonTier][i].id !== undefined) return;
	for (let i = tier; i <= buttonTier - 1; i++) {
		a[i][column].up = m;
		a[i + 1][column].down = m;
	}
	if (buttonColumn < column) a[buttonTier][buttonColumn].rightArrow = m;
	else a[buttonTier][buttonColumn].leftArrow = m;
}

type Coords = [number, number, number, number];

// TALENT_BRANCH_TEXTURECOORDS and TALENT_ARROW_TEXTURECOORDS: [lit, gray]. A left above its
// right means the piece is drawn mirrored.
export const branchCoords: Record<string, [Coords, Coords]> = {
	up: [
		[0.12890625, 0.25390625, 0, 0.484375],
		[0.12890625, 0.25390625, 0.515625, 1.0],
	],
	down: [
		[0, 0.125, 0, 0.484375],
		[0, 0.125, 0.515625, 1.0],
	],
	left: [
		[0.2578125, 0.3828125, 0, 0.5],
		[0.2578125, 0.3828125, 0.5, 1.0],
	],
	right: [
		[0.2578125, 0.3828125, 0, 0.5],
		[0.2578125, 0.3828125, 0.5, 1.0],
	],
	topright: [
		[0.515625, 0.640625, 0, 0.5],
		[0.515625, 0.640625, 0.5, 1.0],
	],
	topleft: [
		[0.640625, 0.515625, 0, 0.5],
		[0.640625, 0.515625, 0.5, 1.0],
	],
	bottomright: [
		[0.38671875, 0.51171875, 0, 0.5],
		[0.38671875, 0.51171875, 0.5, 1.0],
	],
	bottomleft: [
		[0.51171875, 0.38671875, 0, 0.5],
		[0.51171875, 0.38671875, 0.5, 1.0],
	],
	tdown: [
		[0.64453125, 0.76953125, 0, 0.5],
		[0.64453125, 0.76953125, 0.5, 1.0],
	],
	tup: [
		[0.7734375, 0.8984375, 0, 0.5],
		[0.7734375, 0.8984375, 0.5, 1.0],
	],
};

export const arrowCoords: Record<string, [Coords, Coords]> = {
	top: [
		[0, 0.5, 0, 0.5],
		[0, 0.5, 0.5, 1.0],
	],
	right: [
		[1.0, 0.5, 0, 0.5],
		[1.0, 0.5, 0.5, 1.0],
	],
	left: [
		[0.5, 1.0, 0, 0.5],
		[0.5, 1.0, 0.5, 1.0],
	],
};

// pick gives the lit or gray coords for a node's value (1 or -1).
export function pick(c: [Coords, Coords], v: number): Coords {
	return v > 0 ? c[0] : c[1];
}
