import { APLRotation } from './proto/apl';
import { LevelingPlanSettings, LevelingRotation, LevelingRotationPreset, LevelingTalentBuild, LevelingTalentPick } from './proto/ui';
import type { TalentsConfig } from './talents/talents_picker';

// The leveling plan: what we wear, which talents we have and which rotation we play at each
// level on the way up. The Leveling mode sims it level by level and adds up the hours.
//
// The gear comes from the leveling gear plan (leveling_gear.ts). The talents are a build and
// the order we put its points in, so at each level we have the first points of that order.
// A later build is a respec at its level. It puts every point back by its own order, and the
// results show how many hours it saves over keeping the build we had, so we can weigh that
// against the price of the respec. Each rotation is played from its level until the next
// one's, because a rotation tuned at level 30 rests and heals at the wrong times at level 12.
//
// This file must not import the UI (like leveling_gear.ts), so we can run it in node.

// XP from each level to the next, in Classic 1.12. Forever may have changed it (Need to Verify
// in notes.md).
const XP_TO_NEXT_LEVEL: Record<number, number> = {
	1: 400,
	2: 900,
	3: 1400,
	4: 2100,
	5: 2800,
	6: 3600,
	7: 4500,
	8: 5400,
	9: 6500,
	10: 7600,
	11: 8800,
	12: 10100,
	13: 11400,
	14: 12900,
	15: 14400,
	16: 16000,
	17: 17700,
	18: 19400,
	19: 21300,
	20: 23200,
	21: 25200,
	22: 27300,
	23: 29400,
	24: 31700,
	25: 34000,
	26: 36400,
	27: 38900,
	28: 41400,
	29: 44300,
	30: 47400,
	31: 50800,
	32: 54500,
	33: 58600,
	34: 62800,
	35: 67100,
	36: 71600,
	37: 76100,
	38: 80800,
	39: 85700,
	40: 90700,
	41: 95800,
	42: 101000,
	43: 106300,
	44: 111800,
	45: 117500,
	46: 123200,
	47: 129100,
	48: 135100,
	49: 141200,
	50: 147500,
	51: 153900,
	52: 160400,
	53: 167100,
	54: 173900,
	55: 180800,
	56: 187900,
	57: 195000,
	58: 202300,
	59: 209800,
};

export function xpToNextLevel(level: number): number {
	return XP_TO_NEXT_LEVEL[level] ?? 0;
}

export const LEVELING_PLAN_MAX_LEVEL = 60;

// Forever gives its extra talent points from level 25 on.
export const BONUS_TALENT_POINTS_LEVEL = 25;
const MAX_TALENT_POINTS = 51;

export function talentPointsAt(level: number, bonusPoints: number): number {
	const bonus = level >= BONUS_TALENT_POINTS_LEVEL ? bonusPoints : 0;
	return Math.min(MAX_TALENT_POINTS, Math.max(0, level - 9) + bonus);
}

// The level we put in a build's point number index (0 for the first point), when the build
// starts at startLevel. A respec puts all the points it has at its level back at once.
export function pickLevel(index: number, startLevel: number, bonusPoints: number): number {
	for (let level = Math.max(startLevel, 10); level <= LEVELING_PLAN_MAX_LEVEL; level++) {
		if (talentPointsAt(level, bonusPoints) > index) return level;
	}
	return LEVELING_PLAN_MAX_LEVEL + 1;
}

// The levels we sim, from the plan's first level up to, but not including, its last.
export function planLevels(plan: LevelingPlanSettings): number[] {
	const levels: number[] = [];
	for (let level = plan.fromLevel; level < plan.toLevel; level++) levels.push(level);
	return levels;
}

// The builds by level, lowest first. The first one starts at the plan's first level whatever
// its own level says.
export function sortedBuilds(plan: LevelingPlanSettings): LevelingTalentBuild[] {
	return [...plan.builds].sort((a, b) => a.level - b.level);
}

export function buildIndexAt(plan: LevelingPlanSettings, level: number): number {
	const builds = sortedBuilds(plan);
	let index = 0;
	builds.forEach((build, i) => {
		if (i > 0 && build.level <= level) index = i;
	});
	return index;
}

// Where a build is played: from its level up to the next build's, or the plan's end.
export function buildLevels(plan: LevelingPlanSettings, index: number): number[] {
	const builds = sortedBuilds(plan);
	const from = index == 0 ? plan.fromLevel : builds[index].level;
	const to = index + 1 < builds.length ? builds[index + 1].level : plan.toLevel;
	return planLevels(plan).filter(level => level >= from && level < to);
}

// The level ranges by level, lowest first. The first one starts at the plan's first level
// whatever its own level says.
export function sortedRanges(plan: LevelingPlanSettings): LevelingRotation[] {
	return [...plan.rotations].sort((a, b) => a.level - b.level);
}

// The level range we play at a level.
export function rangeAt(plan: LevelingPlanSettings, level: number): LevelingRotation | undefined {
	const ranges = sortedRanges(plan);
	let found = ranges[0];
	ranges.forEach(range => {
		if (range.level <= level) found = range;
	});
	return found;
}

// The first and last level of a range of the plan. The last one is below the first when the
// range starts at or past the plan's last level.
export function rangeLevels(plan: LevelingPlanSettings, index: number): { from: number; to: number } {
	const ranges = sortedRanges(plan);
	const from = index == 0 ? plan.fromLevel : ranges[index].level;
	const to = index + 1 < ranges.length ? ranges[index + 1].level - 1 : plan.toLevel - 1;
	return { from, to };
}

export function libraryRotation(plan: LevelingPlanSettings, name: string): LevelingRotationPreset | undefined {
	return plan.rotationLibrary.find(preset => preset.name == name);
}

// The rotation we play at a level.
export function rotationAt(plan: LevelingPlanSettings, level: number): APLRotation | undefined {
	const range = rangeAt(plan, level);
	return range ? libraryRotation(plan, range.name)?.rotation : undefined;
}

// A name for a new rotation of the library, the name itself when it's free, or else with a
// number after it, like "Level 20 Leveling Plan 2".
export function freeRotationName(plan: LevelingPlanSettings, name: string): string {
	const base = name.trim() || 'Rotation';
	if (!libraryRotation(plan, base)) return base;
	for (let n = 2; ; n++) {
		if (!libraryRotation(plan, `${base} ${n}`)) return `${base} ${n}`;
	}
}

// Moves the rotations of a plan saved before the library into it.
//
// Such a plan held each level range's rotation in the range itself, named after the preset it
// came from, or with no name for a copy of the Rotation tab. Each one goes into the library
// under that name, and two ranges with the same rotation share it.
export function withRotationLibrary(plan: LevelingPlanSettings): LevelingPlanSettings {
	if (!plan.rotations.some(range => range.rotation)) return plan;
	const result = LevelingPlanSettings.clone(plan);
	for (const range of result.rotations) {
		const rotation = range.rotation;
		range.rotation = undefined;
		if (!rotation) continue;
		const same = result.rotationLibrary.find(preset => preset.rotation && APLRotation.equals(preset.rotation, rotation));
		if (same) {
			range.name = same.name;
			continue;
		}
		const name = freeRotationName(result, range.name || 'Rotation tab copy');
		result.rotationLibrary.push(LevelingRotationPreset.create({ name, rotation }));
		range.name = name;
	}
	return result;
}

// The shape of a class's talent trees, as the leveling plan needs it: each talent's row and
// column, its ranks and the talent it needs first.
export interface TalentShape {
	pointsPerRow: number;
	trees: Array<{
		name: string;
		// notSimulated: the sim gives nothing for the talent, like Improved Ghost Wolf.
		talents: Array<{ name: string; row: number; col: number; maxPoints: number; prereq?: number; notSimulated?: boolean }>;
	}>;
}

export function talentShape(config: TalentsConfig<any>, pointsPerRow = 5): TalentShape {
	return {
		pointsPerRow,
		trees: config.map(tree => ({
			name: tree.name,
			talents: tree.talents.map(talent => ({
				name: talent.name ?? String(talent.fieldName ?? 'Talent'),
				row: talent.location.rowIdx,
				col: talent.location.colIdx,
				maxPoints: talent.maxPoints,
				prereq: talent.prereqLocation
					? tree.talents.findIndex(t => t.location.rowIdx == talent.prereqLocation!.rowIdx && t.location.colIdx == talent.prereqLocation!.colIdx)
					: undefined,
				notSimulated: talent.notSimulated,
			})),
		})),
	};
}

export function emptyCounts(shape: TalentShape): number[][] {
	return shape.trees.map(tree => tree.talents.map(() => 0));
}

export function countsOf(picks: LevelingTalentPick[], shape: TalentShape): number[][] {
	const counts = emptyCounts(shape);
	picks.forEach(pick => {
		if (counts[pick.tree]?.[pick.talent] !== undefined) counts[pick.tree][pick.talent]++;
	});
	return counts;
}

// The talents string of some counts, like "-0540320010051122".
export function countsToTalentsString(counts: number[][]): string {
	return counts
		.map(tree => tree.join('').replace(/0+$/g, ''))
		.join('-')
		.replace(/-+$/g, '');
}

export function talentsStringToCounts(talents: string, shape: TalentShape): number[][] {
	const parts = talents.split('-');
	return shape.trees.map((tree, t) => tree.talents.map((_, i) => Number(parts[t]?.charAt(i) || 0)));
}

// The talents we have after the first count picks of a build.
export function buildTalentsString(build: LevelingTalentBuild | undefined, count: number, shape: TalentShape): string {
	return countsToTalentsString(countsOf(build?.picks.slice(0, Math.max(count, 0)) ?? [], shape));
}

// The plan's talents string at a level, with the build it plays there. With buildIndex we
// take that build instead, like the one we had before a respec.
export function talentsAt(plan: LevelingPlanSettings, level: number, shape: TalentShape, buildIndex?: number): string {
	const build = sortedBuilds(plan)[buildIndex ?? buildIndexAt(plan, level)];
	return buildTalentsString(build, talentPointsAt(level, plan.bonusTalentPoints), shape);
}

// Whether a point in a talent is allowed when we have the counts: the tree has the points its
// row needs, the talent it needs is full, and it has a rank left.
function canPick(counts: number[][], shape: TalentShape, tree: number, talent: number): boolean {
	const config = shape.trees[tree]?.talents[talent];
	if (!config) return false;
	if (counts[tree][talent] >= config.maxPoints) return false;
	const treePoints = counts[tree].reduce((sum, points) => sum + points, 0);
	if (treePoints < config.row * shape.pointsPerRow) return false;
	if (config.prereq !== undefined && counts[tree][config.prereq] < shape.trees[tree].talents[config.prereq].maxPoints) return false;
	return true;
}

// Why each pick of a build can't be taken when we come to it, or null when it can.
export function pickProblems(picks: LevelingTalentPick[], shape: TalentShape): Array<string | null> {
	const counts = emptyCounts(shape);
	return picks.map(pick => {
		const config = shape.trees[pick.tree]?.talents[pick.talent];
		if (!config) return 'This talent is gone from the tree';
		let problem: string | null = null;
		if (counts[pick.tree][pick.talent] >= config.maxPoints) problem = 'It already has all its ranks';
		else if (counts[pick.tree].reduce((sum, points) => sum + points, 0) < config.row * shape.pointsPerRow)
			problem = `Its row needs ${config.row * shape.pointsPerRow} points in ${shape.trees[pick.tree].name} first`;
		else if (config.prereq !== undefined && counts[pick.tree][config.prereq] < shape.trees[pick.tree].talents[config.prereq].maxPoints)
			problem = `It needs ${shape.trees[pick.tree].talents[config.prereq].name} full first`;
		counts[pick.tree][pick.talent]++;
		return problem;
	});
}

// The picks that take us from the counts we have to the target, in an order we can take them.
// We take the talents of the lowest row first, and in a row the tree and the talent that come
// first. When nothing left can be taken (a target the tree doesn't allow), we put the rest at
// the end anyway, and pickProblems marks them.
export function picksToward(from: number[][], target: number[][], shape: TalentShape): LevelingTalentPick[] {
	const counts = from.map(tree => [...tree]);
	const picks: LevelingTalentPick[] = [];
	for (;;) {
		let best: { tree: number; talent: number; row: number } | null = null;
		shape.trees.forEach((tree, t) =>
			tree.talents.forEach((config, i) => {
				if (counts[t][i] >= (target[t]?.[i] ?? 0) || !canPick(counts, shape, t, i)) return;
				if (!best || config.row < best.row) best = { tree: t, talent: i, row: config.row };
			}),
		);
		if (!best) break;
		const { tree, talent } = best as { tree: number; talent: number };
		counts[tree][talent]++;
		picks.push(LevelingTalentPick.create({ tree, talent }));
	}
	shape.trees.forEach((tree, t) =>
		tree.talents.forEach((_, i) => {
			for (let n = counts[t][i]; n < (target[t]?.[i] ?? 0); n++) picks.push(LevelingTalentPick.create({ tree: t, talent: i }));
		}),
	);
	return picks;
}

// A build whose talents at some levels are the given talents strings, like the Level 10 to
// Level 30 builds of a spec's presets. Between two of them we take the new points lowest row
// first (picksToward).
export function buildFromMilestones(level: number, milestones: string[], shape: TalentShape): LevelingTalentBuild {
	let counts = emptyCounts(shape);
	const picks: LevelingTalentPick[] = [];
	for (const talents of milestones) {
		const target = talentsStringToCounts(talents, shape);
		picks.push(...picksToward(counts, target, shape));
		counts = countsOf(picks, shape);
	}
	return LevelingTalentBuild.create({ level, picks });
}

// The build's picks after the talents picker changed its talents to newTalents. A talent with
// a point more gets a pick at the end, and one with fewer points loses its last picks, so the
// order we click the talents in is the order we put the points in.
export function picksAfterEdit(picks: LevelingTalentPick[], newTalents: string, shape: TalentShape): LevelingTalentPick[] {
	const before = countsOf(picks, shape);
	const after = talentsStringToCounts(newTalents, shape);
	let result = [...picks];
	shape.trees.forEach((tree, t) =>
		tree.talents.forEach((_, i) => {
			for (let n = before[t][i]; n > after[t][i]; n--) {
				const last = result.map(pick => pick.tree == t && pick.talent == i).lastIndexOf(true);
				if (last >= 0) result = [...result.slice(0, last), ...result.slice(last + 1)];
			}
		}),
	);
	shape.trees.forEach((tree, t) =>
		tree.talents.forEach((_, i) => {
			for (let n = before[t][i]; n < after[t][i]; n++) result.push(LevelingTalentPick.create({ tree: t, talent: i }));
		}),
	);
	return result;
}

export function talentName(pick: LevelingTalentPick, shape: TalentShape): string {
	return shape.trees[pick.tree]?.talents[pick.talent]?.name ?? 'Unknown talent';
}

// The hours a level takes us, when we get xpPerHour.
export function hoursAtLevel(level: number, xpPerHour: number): number {
	return xpPerHour > 0 ? xpToNextLevel(level) / xpPerHour : Infinity;
}
