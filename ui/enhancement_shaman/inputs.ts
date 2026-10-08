import * as InputHelpers from '../core/components/input_helpers.js';
import { shamanImbueTooltip } from '../core/forever/shaman_imbues';
import { Player } from '../core/player.js';
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

// The game's tooltip of the imbue, at the rank and the level of the character.
const imbueTooltip = (imbue: WeaponImbue) => (player: Player<Spec.SpecEnhancementShaman>) => shamanImbueTooltip(imbue, player.getEffectiveLevel());

export const ShamanImbueInput = InputHelpers.makeSpecOptionsEnumInput<Spec.SpecEnhancementShaman>({
	fieldName: 'shamanImbue',
	label: 'Shaman weapon imbue',
	labelTooltip:
		'The main hand imbue. In Forever it no longer takes the weapon enchant slot, ' +
		'so the main hand imbue under consumables takes an oil or a stone on top of it.',
	values: [
		{ name: 'None', value: WeaponImbue.WeaponImbueUnknown },
		{ name: 'Windfury Weapon', value: WeaponImbue.WindfuryWeapon, icon: 'spell_nature_cyclone', richTooltip: imbueTooltip(WeaponImbue.WindfuryWeapon) },
		{
			name: 'Rockbiter Weapon',
			value: WeaponImbue.RockbiterWeapon,
			icon: 'spell_nature_rockbiter',
			richTooltip: imbueTooltip(WeaponImbue.RockbiterWeapon),
		},
		// The game's own icon name has the typo.
		{
			name: 'Flametongue Weapon',
			value: WeaponImbue.FlametongueWeapon,
			icon: 'spell_fire_flametounge',
			richTooltip: imbueTooltip(WeaponImbue.FlametongueWeapon),
		},
		{
			name: 'Frostbrand Weapon',
			value: WeaponImbue.FrostbrandWeapon,
			icon: 'spell_frost_frostbrand',
			richTooltip: imbueTooltip(WeaponImbue.FrostbrandWeapon),
		},
	],
	showWhen: player => player.sim.getRuleset() === Ruleset.RulesetForever,
});
