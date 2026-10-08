import * as InputHelpers from '../core/components/input_helpers.js';
import { makeShamanImbueInput } from '../core/components/inputs/shaman_imbues';
import { Ruleset } from '../core/proto/api.js';
import { Spec } from '../core/proto/common.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const RaidDamageHitsInput = InputHelpers.makeSpecOptionsNumberInput<Spec.SpecEnhancementShaman>({
	fieldName: 'raidDamageHitsPerMinute',
	label: 'Raid hits per minute',
	labelTooltip:
		'Spell hits the shaman takes from raid damage per minute. Each spends a Water Shield globe (2% mana, one every 3.5 sec at most). ' +
		'The boss is on the tank, so this is the only thing that feeds Water Shield. Lightning Shield needs melee hits and ignores it.',
	float: true,
	positive: true,
	showWhen: player => player.sim.getRuleset() === Ruleset.RulesetForever,
});

export const RaidDamageHitsVariationInput = InputHelpers.makeSpecOptionsNumberInput<Spec.SpecEnhancementShaman>({
	fieldName: 'raidDamageHitsPerMinuteVariation',
	label: 'Raid hits +/-',
	labelTooltip:
		'Each sim iteration takes a random rate between [value, -1 * value] hits per minute away from Raid hits per minute. ' +
		'For example, 6 hits per minute with +/- 2 gives each iteration a rate between 4 and 8.',
	float: true,
	positive: true,
	showWhen: player => player.sim.getRuleset() === Ruleset.RulesetForever,
});

export const ShamanImbueInput = makeShamanImbueInput<Spec.SpecEnhancementShaman>();
