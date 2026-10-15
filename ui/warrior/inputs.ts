import * as InputHelpers from '../core/components/input_helpers.js';
import { Ruleset } from '../core/proto/api.js';
import { Spec } from '../core/proto/common.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const RaidDamageHitsInput = InputHelpers.makeSpecOptionsNumberInput<Spec.SpecWarrior>({
	fieldName: 'raidDamageHitsPerMinute',
	label: 'Raid hits per minute',
	labelTooltip:
		'Direct hits the warrior takes from raid damage per minute, on average. They land at random times, and each gives Rage like any damage taken ' +
		'(about 15 Rage for 1500 damage from a level 63 boss). The boss is on the tank, so these are the only hits a DPS warrior takes.',
	float: true,
	positive: true,
	showWhen: player => player.sim.getRuleset() === Ruleset.RulesetForever,
});

export const RaidDamagePerHitInput = InputHelpers.makeSpecOptionsNumberInput<Spec.SpecWarrior>({
	fieldName: 'raidDamagePerHit',
	label: 'Raid damage per hit',
	labelTooltip: 'The damage of each raid hit, before your damage taken modifiers (Recklessness takes 20% more).',
	positive: true,
	showWhen: player => player.sim.getRuleset() === Ruleset.RulesetForever,
});
