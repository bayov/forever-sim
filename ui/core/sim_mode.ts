import { Encounter } from './encounter';
import { Encounter as EncounterProto } from './proto/common';

// The ways we sim a character. They score different things and don't compare, so each has
// its own sim history and its own last run in the Results tab.
//
// DPS is a fight against a boss. PvP is a fight against an enemy player, where white hits never
// glance. Grinding is a session against normal mobs, scored by the mobs we kill. Leveling sims
// the leveling plan (leveling_plan.ts) level by level, with a grinding session at each, and
// scores it by the hours it takes.
//
// Grinding was called Leveling before we had the plan, so its id is still 'leveling'. That
// keeps the runs and the encounters we saved under it.
//
// The mode isn't a setting of its own. It comes from the encounter: PvP when it's a PvP fight,
// Grinding when it has leveling options, and Leveling when those options say plan. So loading a
// preset or the settings of a run also picks its mode.
export type SimMode = 'dps' | 'pvp' | 'leveling' | 'levelingPlan';

export const SIM_MODES: SimMode[] = ['dps', 'pvp', 'leveling', 'levelingPlan'];

export const simModeNames: Record<SimMode, string> = {
	dps: 'DPS',
	pvp: 'PvP',
	leveling: 'Grinding',
	levelingPlan: 'Leveling',
};

export function encounterMode(encounter: Encounter): SimMode {
	if (encounter.getLeveling()?.plan) return 'levelingPlan';
	if (encounter.isLeveling()) return 'leveling';
	if (encounter.getPvp()) return 'pvp';
	return 'dps';
}

export function encounterProtoMode(encounter: EncounterProto | undefined): SimMode {
	if (encounter?.leveling?.plan) return 'levelingPlan';
	if (encounter?.leveling) return 'leveling';
	if (encounter?.pvp) return 'pvp';
	return 'dps';
}
