// The talent trees the talents frame draws: the sim's Forever trees (ui/core/talents/trees), the
// character's points from the request's talents string, and where the client puts each talent.

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

// Where the client puts the talents. Forever's three trees of a class are one trait tree, and
// each talent is a TraitNode at PosX, PosY (the same grid for every class at build 70425). The
// frame draws a node's center at PosX / 10, PosY / 10 from the talent frame's top-left, less
// its pan offset (ClassTalentsFrame's basePanOffset 49, 24 less the Camelot ClassVisuals
// panOffset 60, 31).
const treeLeft = [102, 502, 908];
const rowTop = 213;
const step = 60;
const pan = { x: 49 - 60, y: 24 - 31 };

// nodeCenter gives a talent's center in the talent frame, from its tree and grid spot.
export function nodeCenter(tree: number, loc: TalentLoc): { x: number; y: number } {
	return { x: treeLeft[tree] + loc.colIdx * step - pan.x, y: rowTop + loc.rowIdx * step - pan.y };
}

// The trees' icons (TalentTab's icon), in the order of the class's trees.
export const treeIcons: Record<string, string[]> = {
	ClassDruid: ['spell_nature_starfall', 'ability_racial_bearform', 'spell_nature_healingtouch'],
	ClassHunter: ['ability_hunter_beasttaming', 'ability_marksmanship', 'ability_hunter_swiftstrike'],
	ClassMage: ['spell_holy_magicalsentry', 'spell_fire_firebolt02', 'spell_frost_frostbolt02'],
	ClassPaladin: ['spell_holy_holybolt', 'spell_holy_devotionaura', 'spell_holy_auraoflight'],
	ClassPriest: ['spell_holy_wordfortitude', 'spell_holy_guardianspirit', 'spell_shadow_shadowwordpain'],
	ClassRogue: ['ability_rogue_eviscerate', 'ability_backstab', 'ability_stealth'],
	ClassShaman: ['spell_nature_lightning', 'spell_nature_lightningshield', 'spell_nature_magicimmunity'],
	ClassWarlock: ['spell_shadow_deathcoil', 'spell_shadow_metamorphosis', 'spell_shadow_rainoffire'],
	ClassWarrior: ['ability_rogue_eviscerate', 'ability_warrior_innerrage', 'inv_shield_06'],
};
