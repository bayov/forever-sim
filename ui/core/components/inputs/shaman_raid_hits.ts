import { Ruleset } from '../../proto/api.js';
import { Spec } from '../../proto/common.js';
import { SpecOptions } from '../../proto_utils/utils.js';
import * as InputHelpers from '../input_helpers.js';

// The raid damage that hits the shaman, in the encounter settings of Elemental and Enhancement.
// It is what sets off Lightning Shield and Water Shield, because the boss is on the tank.

type RaidHitsSpec = Spec.SpecElementalShaman | Spec.SpecEnhancementShaman;

export function makeRaidDamageHitsInput<SpecType extends RaidHitsSpec>() {
	return InputHelpers.makeSpecOptionsNumberInput<SpecType>({
		fieldName: 'raidDamageHitsPerMinute' as keyof SpecOptions<SpecType>,
		label: 'Raid hits per minute',
		labelTooltip:
			'Hits the shaman takes from raid damage per minute, on average. They land at random times. Each one fires a Lightning Shield orb at the target or spends a Water Shield globe (2% mana). ' +
			'Both shields go off at most once every 3.5 sec. The boss is on the tank, so this is the only thing that sets off your shield.',
		float: true,
		positive: true,
		showWhen: player => player.sim.getRuleset() === Ruleset.RulesetForever,
	});
}

export function makeRaidDamageHitsVariationInput<SpecType extends RaidHitsSpec>() {
	return InputHelpers.makeSpecOptionsNumberInput<SpecType>({
		fieldName: 'raidDamageHitsPerMinuteVariation' as keyof SpecOptions<SpecType>,
		label: 'Raid hits +/-',
		labelTooltip:
			'Each sim iteration takes a random rate between [value, -1 * value] hits per minute away from Raid hits per minute. ' +
			'For example, 6 hits per minute with +/- 2 gives each iteration a rate between 4 and 8.',
		float: true,
		positive: true,
		showWhen: player => player.sim.getRuleset() === Ruleset.RulesetForever,
	});
}
