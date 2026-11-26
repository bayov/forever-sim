// Runs the UI's leveling gear plan (ui/core/leveling_gear.ts) on LevelingPlanSearchRequests, for
// TestGearPlanMatchesTS.
//
//	node gear_twin.mjs db.bin requests.json
//
// requests.json is a list of requests in proto JSON. It prints a list of gear plans, each a list
// with a line for each level: the level, then slot:item/suffix for each slot.
import { build } from 'esbuild';
import { readFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const core = join(dirname(fileURLToPath(import.meta.url)), '../../ui/core');
const outfile = join(tmpdir(), `leveling_gear_twin_${process.pid}.mjs`);
await build({
	stdin: {
		contents: `
			export * from './leveling_gear';
			export { UIDatabase, LevelingPlanSearchRequest, LevelingWeaponStyle } from './proto/ui';
			export { Stats } from './proto_utils/stats';
		`,
		resolveDir: core,
		loader: 'ts',
	},
	bundle: true,
	format: 'esm',
	platform: 'node',
	outfile,
	logLevel: 'error',
	// stats.ts takes a few helpers from files that need a browser.
	plugins: [
		{
			name: 'no-browser',
			setup(build) {
				build.onResolve({ filter: /\/(names|utils)(\.js)?$/ }, () => ({ path: 'stub', namespace: 'stub' }));
				build.onLoad({ filter: /.*/, namespace: 'stub' }, () => ({
					contents: `
						export const getClassStatName = () => '';
						export const pseudoStatNames = new Map();
						export const getEnumValues = e => Object.values(e).filter(value => typeof value == 'number');
					`,
				}));
			},
		},
	],
});
const { planLevelingGear, LEVELING_GEAR_SLOTS, UIDatabase, LevelingPlanSearchRequest, LevelingWeaponStyle, Stats } = await import(
	pathToFileURL(outfile).href
);

const [dbFile, requestsFile] = process.argv.slice(2);
const db = UIDatabase.fromBinary(readFileSync(dbFile));
const items = new Map(db.items.map(item => [item.id, item]));
const suffixes = new Map(db.randomSuffixes.map(suffix => [suffix.id, suffix]));

// weightsAt in leveling_plan_gear.tsx.
function weightsAt(presets, level) {
	const above = presets.findIndex(preset => preset.level >= level);
	if (above == -1) return presets[presets.length - 1].weights;
	if (above == 0 || presets[above].level == level) return presets[above].weights;
	const [lo, hi] = [presets[above - 1], presets[above]];
	const t = (level - lo.level) / (hi.level - lo.level);
	return lo.weights.scale(1 - t).add(hi.weights.scale(t));
}

const out = JSON.parse(readFileSync(requestsFile, 'utf8')).map(json => {
	const req = LevelingPlanSearchRequest.fromJson(json);
	const presets = req.gearWeights.map(p => ({ level: p.level, style: p.style, weights: Stats.fromProto(p.weights) }));
	const presetsOf = style => {
		const own = presets.filter(p => p.style == style);
		return own.length ? own : presets.filter(p => p.style == LevelingWeaponStyle.LevelingWeaponStyleFree);
	};
	const rules = req.classRules;
	const lookup = ids => ids.flatMap(id => items.get(id) ?? []);
	const plan = planLevelingGear({
		items: db.items,
		randomSuffix: id => suffixes.get(id),
		playerClass: req.base.raid.parties[0].players[0].class,
		classRules: {
			maxArmorTypeAt: level => rules.maxArmorType[Math.min(Math.max(level, 1), rules.maxArmorType.length) - 1],
			weaponTypes: rules.weaponTypes.map(weaponType => ({ weaponType, canUseTwoHand: rules.twoHandWeaponTypes.includes(weaponType) })),
			rangedWeaponTypes: rules.rangedWeaponTypes,
			canDualWield: rules.canDualWield,
			dualWieldLevel: rules.dualWieldLevel,
		},
		faction: req.faction,
		seed: req.gearPlan.seed,
		pins: [],
		route: req.route && {
			quests: req.route.quests.map(q => ({ name: q.name, level: q.level, choices: lookup(q.choiceIds), rewards: lookup(q.rewardIds) })),
			dungeons: req.route.dungeons,
		},
		oneHandAndShield: req.plan.oneHandAndShield,
		weaponStyles: req.plan.weaponStyles,
		weightsAt: (level, style) => weightsAt(presetsOf(style), level),
	});
	return plan.levels.map(
		l =>
			`${l.level}` +
			LEVELING_GEAR_SLOTS.filter(slot => l.gear[slot])
				.map(slot => ` ${slot}:${l.gear[slot].item.id}/${l.gear[slot].randomSuffix?.id ?? 0}`)
				.join(''),
	);
});
console.log(JSON.stringify(out));
