import * as PresetUtils from '../core/preset_utils.js';
import { Consumes, Debuffs, Flask, Food, IndividualBuffs, PartyBuffs, RaidBuffs, TristateEffect, UnitReference } from '../core/proto/common.js';
import { RestorationDruid_Options as RestorationDruidOptions } from '../core/proto/druid.js';
import { SavedTalents } from '../core/proto/ui.js';
import BlankGear from './gear_sets/blank.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const DefaultGear = PresetUtils.makePresetGear('Blank', BlankGear);

// Default talents, in the sim's per-tree digit format, on the trees of the launch realms' 2026-10-30
// hotfix (tools/forever_talents/client/druid.json). The hotfix swapped Furor (now Feral) and Heart of
// the Wild (now Restoration row 1), and Reflection (now row 2) and Subtlety (now row 3), and put Bloom
// in Gift of the Earthmother's place. Thicc Resto is the raid sim's Restoration druid preset
// (docs/raid_sim.md): Genesis (10% more healing over time since the hotfix), Moonglow, Nature's Majesty
// and Nature's Splendor in Balance, and every core Restoration talent. The second build, which took
// Furor while it was a Restoration talent (Celestial Focus is gone from Forever), takes Heart of the
// Wild in its place (10% Intellect).
export const HeartOfTheWildRestoTalents = {
	name: 'Heart of the Wild Resto',
	data: SavedTalents.create({
		talentsString: '05--5553035153113051',
	}),
};
export const ThiccRestoTalents = {
	name: 'Thicc Resto',
	data: SavedTalents.create({
		talentsString: '05302001--5053025153113051',
	}),
};

export const DefaultOptions = RestorationDruidOptions.create({
	innervateTarget: UnitReference.create(),
});

export const DefaultConsumes = Consumes.create({
	flask: Flask.FlaskUnknown,
	food: Food.FoodUnknown,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	divineSpirit: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
});

export const DefaultPartyBuffs = PartyBuffs.create({});

export const DefaultDebuffs = Debuffs.create({
	faerieFire: true,
	sunderArmor: true,
});

export const OtherDefaults = {
	distanceFromTarget: 18,
};
