import * as PresetUtils from '../core/preset_utils.js';
import {
	Conjured,
	Consumes,
	Debuffs,
	Flask,
	Food,
	IndividualBuffs,
	IntellectElixir,
	ManaRegenElixir,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SpellPowerBuff,
	SpiritElixir,
	SunderRamp,
	TristateEffect,
	UnitReference,
	UnitReference_Type as UnitType,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { RestorationDruid_Options as RestorationDruidOptions } from '../core/proto/druid.js';
import { SavedTalents } from '../core/proto/ui.js';
import RaidHealerAPLJson from './apls/raid_healer.apl.json';
import TankHealerAPLJson from './apls/tank_healer.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';

// The Restoration druid, a healer of the healing sim (docs/healing_druid.md, docs/healing_sim.md H3). Its
// numbers come from tools/healbench, which scores a healer by the damage it carries in a role: the tank
// healer on the boss's melee and abilities on the tanks, the raid healer on the boss's damage on the rest.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

// The Restoration druid synthetic Phase 1 set (tools/database/forever_synthetic_items_druid_restoration.go, ids
// 990401 to 990413), on the Restoration druid row of orchestration/gear_targets.md: the two Wowhead druid lists'
// average, healing on top of spell power, Intellect, Spirit 10%, a little crit and MP5, no hit. The shaman's
// trinkets and enchants, the staff path, and the Idol of Swiftness (3 sec off Swiftmend).
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Restoration', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS for a healer under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

export const GearPresets = [GearLevel60Phase1Synthetic];

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// Both rotations cast Innervate on ourselves at 35% mana, Swiftmend and then Regrowth at whoever in the
// assignment is under the emergency line, and bind at most 4 ranks of a heal, each cast when what's missing
// covers its Expected Healing (Forever's downranking from the engine takes the low ranks' coefficients down).
//
// The tank healer: Rejuvenation on the tank whenever it's off, Regrowth on it when its HoT is off and the hole
// covers the heal, Swiftmend at it, Healing Touch 11, 8 and 6 at 90% of the heal, then Rejuvenation at the most
// hurt of the rest. The emergency line is 40%. 1,000 iterations, carry: Onyxia 1.165, Broodlord 0.731,
// Vaelastrasz 0.688, Chromaggus 0.870, Onyxia at 180 / 300 sec 1.087 / 1.063, against the Restoration shaman's
// 1.141, 0.837, 0.775, 0.884, 1.010 / 0.969. Healing Touch 11, 10, 9 instead (500 iterations): 1.094, 0.701,
// 0.617, 0.790.
export const APLTankHealer = PresetUtils.makePresetAPLRotation('Tank healer', TankHealerAPLJson, {
	tooltip: 'Rejuvenation and Regrowth on the tank, Swiftmend and bound ranks of Healing Touch at it, with Innervate on yourself.',
	group: 'Healing sim',
});
// The raid healer: Wild Growth at the most hurt, Rejuvenation 11 at the most hurt once it misses 20% of the HoT,
// Swiftmend, Regrowth 9, then Healing Touch 11 and 9; the emergency line is 50%. 1,000 iterations, carry:
// Onyxia 0.711, Broodlord 0.828, Vaelastrasz 1.092, Chromaggus 0.953, Ragnaros 1.240, against the shaman's 0.678,
// 0.817, 1.110, 0.906, 1.171. Rejuvenation at 30% and a 35% line: 0.679, 0.815, 1.086, 0.887, 1.181.
export const APLRaidHealer = PresetUtils.makePresetAPLRotation('Raid healer', RaidHealerAPLJson, {
	tooltip: 'Rejuvenation, Wild Growth, Swiftmend and Regrowth at the most hurt, Healing Touch for a big hole, with Innervate on yourself.',
	group: 'Healing sim',
});

export const APLPresets = [APLTankHealer, APLRaidHealer];

export const DefaultAPL = APLTankHealer;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Thicc Resto (51 points, 40 of them Restoration and Genesis, Moonglow, Nature's Majesty and Nature's Splendor
// from Balance). On the tank healer at 500 iterations: Heart of the Wild for Nature's Focus is level (Onyxia
// 1.177 against 1.172), 46 Restoration with 5 Genesis loses 0.06 on Onyxia, all 51 in Restoration 0.08.
export const TalentsRestoration = PresetUtils.makePresetTalents('Restoration', SavedTalents.create({ talentsString: '05302001--5053025153113051' }), {
	tooltip: 'Wild Growth, Swiftmend, Bloom and Nature\'s Swiftness, with Genesis and Nature\'s Splendor from Balance.',
	group: 'Healing sim',
});

export const TalentPresets = [TalentsRestoration];

export const DefaultTalents = TalentsRestoration;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// Innervate on ourselves.
export const DefaultOptions = RestorationDruidOptions.create({
	innervateTarget: UnitReference.create({ type: UnitType.Self }),
});

// The Restoration shaman's level 60 consumables (the Balance druid's with a Brilliant Mana Oil and the Major
// Mana Potion).
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.BrilliantManaOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfSages,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The shared level 60 raid buffs and debuffs, as the Restoration shaman's.
export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	bloodPact: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	frostResistanceAura: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	leaderOfThePack: true,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
	salvation: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	curseOfRecklessness: true,
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	demoralizingRoar: TristateEffect.TristateEffectImproved,
	demoralizingShout: TristateEffect.TristateEffectImproved,
	faerieFire: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	scorpidSting: true,
	sunderRamp: SunderRamp.create({ tank: true, pullSunders: 1 }),
	thunderClap: TristateEffect.TristateEffectImproved,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	channelClipDelay: 50,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceSkyborneWindshaper,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

export const PresetBuildTankHealer = PresetUtils.makePresetBuild('Tank healer', {
	group: 'Healing sim',
	tooltip: 'Level 60 Windshaper Skyborne on the Restoration synthetic set with the tank healer rotation.',
	race: Race.RaceSkyborneWindshaper,
	level: 60,
	gear: DefaultGear,
	talents: TalentsRestoration,
	rotation: APLTankHealer,
});

export const PresetBuildRaidHealer = PresetUtils.makePresetBuild('Raid healer', {
	group: 'Healing sim',
	tooltip: 'Level 60 Windshaper Skyborne on the Restoration synthetic set with the raid healer rotation.',
	race: Race.RaceSkyborneWindshaper,
	level: 60,
	gear: DefaultGear,
	talents: TalentsRestoration,
	rotation: APLRaidHealer,
});
