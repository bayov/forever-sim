import { Encounter } from './encounter';
import { Encounter as EncounterProto } from './proto/common';

// The three ways we sim a character. They score different things and don't compare, so each
// has its own sim history and its own last run in the Results tab.
//
// DPS is a fight against a boss. PvP is a fight against an enemy player, where white hits never
// glance. Leveling is a grinding session against normal mobs, scored by the mobs we kill.
//
// The mode isn't a setting of its own. It comes from the encounter: PvP when it's a PvP fight,
// and Leveling when it has leveling options. So loading a preset or the settings of a run also
// picks its mode.
export type SimMode = 'dps' | 'pvp' | 'leveling';

export const SIM_MODES: SimMode[] = ['dps', 'pvp', 'leveling'];

export const simModeNames: Record<SimMode, string> = {
	dps: 'DPS',
	pvp: 'PvP',
	leveling: 'Leveling',
};

export function encounterMode(encounter: Encounter): SimMode {
	if (encounter.isLeveling()) return 'leveling';
	if (encounter.getPvp()) return 'pvp';
	return 'dps';
}

export function encounterProtoMode(encounter: EncounterProto | undefined): SimMode {
	if (encounter?.leveling) return 'leveling';
	if (encounter?.pvp) return 'pvp';
	return 'dps';
}
