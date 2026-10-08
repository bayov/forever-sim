import { shamanImbueTooltip } from '../../forever/shaman_imbues';
import { Player } from '../../player';
import { Spec, WeaponImbue } from '../../proto/common.js';
import { ElementalShaman_Options, EnhancementShaman_Options, WardenShaman_Options } from '../../proto/shaman.js';
import { IndividualSimSettings } from '../../proto/ui.js';
import { SpecOptions } from '../../proto_utils/utils.js';
import * as InputHelpers from '../input_helpers.js';

// The shaman's own weapon imbue, in the class settings of every shaman spec.
//
// Under Forever a weapon imbue acts as a buff on the character. It no longer takes the weapon's
// temporary enchant, so it stacks with an oil or a stone in the consumables and with another
// shaman's Windfury or Flametongue Totem.

type ImbueSpec = Spec.SpecElementalShaman | Spec.SpecEnhancementShaman | Spec.SpecWardenShaman;

const SHAMAN_IMBUES = [WeaponImbue.RockbiterWeapon, WeaponImbue.FlametongueWeapon, WeaponImbue.FrostbrandWeapon, WeaponImbue.WindfuryWeapon];

export function makeShamanImbueInput<SpecType extends ImbueSpec>() {
	// The game's tooltip of the imbue, at the rank and the level of the character.
	const imbueTooltip = (imbue: WeaponImbue) => (player: Player<SpecType>) => shamanImbueTooltip(imbue, player.getEffectiveLevel());
	return InputHelpers.makeSpecOptionsEnumInput<SpecType>({
		fieldName: 'shamanImbue' as keyof SpecOptions<SpecType>,
		label: 'Shaman weapon imbue',
		labelTooltip:
			'The main hand imbue. In Forever it acts as a buff on the character and no longer takes the weapon enchant slot, ' +
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
	});
}

// The shaman imbue used to be a consumable. We move one saved in the consumables, from before,
// to the class settings, unless the class settings already have one. Then the consumables keep
// only an oil or a stone.
export function migrateShamanImbue(settings: IndividualSimSettings) {
	const player = settings.player!;
	const consumes = player.consumes;
	const spec = player.spec;
	let options: { shamanImbue: WeaponImbue };
	switch (spec.oneofKind) {
		case 'elementalShaman':
			options = spec.elementalShaman.options ??= ElementalShaman_Options.create();
			break;
		case 'enhancementShaman':
			options = spec.enhancementShaman.options ??= EnhancementShaman_Options.create();
			break;
		case 'wardenShaman':
			options = spec.wardenShaman.options ??= WardenShaman_Options.create();
			break;
		default:
			return;
	}
	if (!consumes) return;
	if (SHAMAN_IMBUES.includes(consumes.mainHandImbue) && !options.shamanImbue) options.shamanImbue = consumes.mainHandImbue;
	if (SHAMAN_IMBUES.includes(consumes.mainHandImbue)) consumes.mainHandImbue = WeaponImbue.WeaponImbueUnknown;
	if (SHAMAN_IMBUES.includes(consumes.offHandImbue)) consumes.offHandImbue = WeaponImbue.WeaponImbueUnknown;
}
