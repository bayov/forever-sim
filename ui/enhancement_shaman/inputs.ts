import * as InputHelpers from '../core/components/input_helpers.js';
import { Ruleset } from '../core/proto/api.js';
import { Spec, WeaponImbue } from '../core/proto/common.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const RaidDamageHitsInput = InputHelpers.makeSpecOptionsNumberInput<Spec.SpecEnhancementShaman>({
	fieldName: 'raidDamageHitsPerMinute',
	label: 'Raid damage hits per minute',
	labelTooltip:
		'Spell hits the shaman takes from raid damage per minute. Each spends a Water Shield globe (2% mana, one every 3.5 sec at most). ' +
		'The boss is on the tank, so this is the only thing that feeds Water Shield. Lightning Shield needs melee hits and ignores it.',
	float: true,
	positive: true,
	showWhen: player => player.sim.getRuleset() === Ruleset.RulesetForever,
});

export const ShamanImbueInput = InputHelpers.makeSpecOptionsEnumInput<Spec.SpecEnhancementShaman>({
	fieldName: 'shamanImbue',
	label: 'Shaman weapon imbue',
	labelTooltip:
		'The main hand imbue. In Forever it no longer takes the weapon enchant slot, ' +
		'so the main hand imbue under consumables takes an oil or a stone on top of it.',
	values: [
		{ name: 'None', value: WeaponImbue.WeaponImbueUnknown },
		{ name: 'Windfury Weapon', value: WeaponImbue.WindfuryWeapon, icon: 'spell_nature_cyclone', spellId: 16362 },
		{ name: 'Rockbiter Weapon', value: WeaponImbue.RockbiterWeapon, icon: 'spell_nature_rockbiter', spellId: 16316 },
		// The game's own icon name has the typo.
		{ name: 'Flametongue Weapon', value: WeaponImbue.FlametongueWeapon, icon: 'spell_fire_flametounge', spellId: 16342 },
		{ name: 'Frostbrand Weapon', value: WeaponImbue.FrostbrandWeapon, icon: 'spell_frost_frostbrand', spellId: 16356 },
	],
	showWhen: player => player.sim.getRuleset() === Ruleset.RulesetForever,
});
