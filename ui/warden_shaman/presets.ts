import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	ArmorElixir,
	AttackPowerBuff,
	Consumes,
	Debuffs,
	Flask,
	Food,
	HealthElixir,
	IndividualBuffs,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	StrengthBuff,
	SunderRamp,
	TotemWeaponBuff,
	TristateEffect,
	WeaponImbue,
} from '../core/proto/common.js';
import { EnhancementShaman_Options as EnhancementShamanOptions, ShamanSyncType } from '../core/proto/shaman.js';
import { SavedTalents } from '../core/proto/ui.js';
import DefaultAPLJSON from './apls/default.apl.json';
import Level60BalancedAPLJSON from './apls/level60_balanced.apl.json';
import Level60MaxSurvivalAPLJSON from './apls/level60_max_survival.apl.json';
import Level60MaxThreatAPLJSON from './apls/level60_max_threat.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import Phase1SyntheticTankGear from './gear_sets/phase1_synthetic_tank.gear.json';
import Phase1SyntheticTankArmorTrinketGear from './gear_sets/phase1_synthetic_tank_armor_trinket.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);
// The synthetic Phase 1 tank set (docs/tank_shaman.md), with the attack power on-use trinket or the armor one.
export const GearLevel60Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic)', Phase1SyntheticTankGear);
export const GearLevel60SyntheticArmorTrinket = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic, armor trinket)', Phase1SyntheticTankArmorTrinketGear);

export const GearPresets = {
	[Phase.Phase1]: [GearBlank, GearLevel60Synthetic, GearLevel60SyntheticArmorTrinket],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultGear = GearBlank;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLDefault = PresetUtils.makePresetAPLRotation('Default', DefaultAPLJSON);

export const APLLevel60MaxThreat = PresetUtils.makePresetAPLRotation('Level 60 Max threat', Level60MaxThreatAPLJSON);
export const APLLevel60Balanced = PresetUtils.makePresetAPLRotation('Level 60 Balanced', Level60BalancedAPLJSON);
export const APLLevel60MaxSurvival = PresetUtils.makePresetAPLRotation('Level 60 Max survival', Level60MaxSurvivalAPLJSON);

export const APLPresets = {
	[Phase.Phase1]: [APLDefault, APLLevel60MaxThreat, APLLevel60Balanced, APLLevel60MaxSurvival],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '35050013-0502333135510102' }));

// The optimized tank builds (docs/tank_shaman.md, Presets), one a weight: Balanced from the earlier shared build
// takes Elemental Weapons 3 for Shamanistic Focus; Max survival a point of Elemental Warding for a point of Flurry;
// Max threat Elemental Devastation 3 for Elemental Warding and Toughness 2.
export const TalentsLevel60Tank = PresetUtils.makePresetTalents('Level 60 Tank', SavedTalents.create({ talentsString: '05050002-050032203551012201-050002' }));
export const TalentsLevel60TankMaxSurvival = PresetUtils.makePresetTalents(
	'Level 60 Tank Max survival',
	SavedTalents.create({ talentsString: '05050003-050032113541012201-050002' }),
);
export const TalentsLevel60TankMaxThreat = PresetUtils.makePresetTalents(
	'Level 60 Tank Max threat',
	SavedTalents.create({ talentsString: '0535-050032303351012201-050002' }),
);

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60Tank, TalentsLevel60TankMaxThreat, TalentsLevel60TankMaxSurvival],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

// The three Level 60 tank presets (threat against survival as weights, Balanced the default; docs/tank_sim.md).
export const PresetBuildLevel60MaxThreat = PresetUtils.makePresetBuild('Level 60 Max threat', {
	tooltip: 'Scored with survival weighed at 1/3 (docs/tank_sim.md). Elemental Devastation talents, Mana Spring Totem, the attack power trinket.',
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60TankMaxThreat,
	rotation: APLLevel60MaxThreat,
	race: Race.RaceTauren,
	professions: [Profession.Engineering, Profession.Blacksmithing],
});
export const PresetBuildLevel60Balanced = PresetUtils.makePresetBuild('Level 60 Balanced', {
	tooltip: "Scored with survival weighed at 1 (docs/tank_sim.md). Max survival's talents score within the paired error here: either build serves both.",
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60Tank,
	rotation: APLLevel60Balanced,
	race: Race.RaceTauren,
	professions: [Profession.Engineering, Profession.Blacksmithing],
});
export const PresetBuildLevel60MaxSurvival = PresetUtils.makePresetBuild('Level 60 Max survival', {
	tooltip:
		"Scored with survival weighed at 3 (docs/tank_sim.md). The armor trinket. Balanced's talents score within the paired error here: either build serves both.",
	gear: GearLevel60SyntheticArmorTrinket,
	talents: TalentsLevel60TankMaxSurvival,
	rotation: APLLevel60MaxSurvival,
	race: Race.RaceTauren,
	professions: [Profession.Engineering, Profession.Blacksmithing],
});

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = EnhancementShamanOptions.create({
	shamanImbue: WeaponImbue.RockbiterWeapon,
	syncType: ShamanSyncType.Auto,
});

export const DefaultConsumes = Consumes.create({
	// The tank bench's consumables (tools/tankbench/req/shaman_warden_*.json): every tank takes the same
	// armor and health elixirs (tank_fairness-20261104-1), the flask per spec.
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	armorElixir: ArmorElixir.ElixirOfThePhalanx,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MajorManaPotion,
	dragonBreathChili: true,
	flask: Flask.FlaskOfTheTitans,
	food: Food.FoodSmokedDesertDumpling,
	healthElixir: HealthElixir.ElixirOfGreaterFortitude,
	strengthBuff: StrengthBuff.JujuPower,
});

// The raid, blessings and debuffs of the tank bench's requests (tools/tankbench/req/shaman_warden_*.json, the same
// in all three), so a preset loads the setup it was scored with (tank_fairness-20261104-6).
export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	bloodPact: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	frostResistanceAura: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	graceOfAirTotem: TristateEffect.TristateEffectRegular,
	leaderOfThePack: true,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
	thorns: TristateEffect.TristateEffectImproved,
	totemWeaponBuff: TotemWeaponBuff.TotemWeaponBuffWindfury,
	trueshotAura: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
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
	sunderRamp: SunderRamp.create({ tank: true }),
	thunderClap: TristateEffect.TristateEffectImproved,
});

export const OtherDefaults = {
	// A tank faces the boss: the boss parries and blocks its attacks, and its parries haste the boss's swing.
	inFrontOfTarget: true,
	distanceFromTarget: 5,
	// The bench's professions (Engineering + Blacksmithing, as every tank's request).
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
};
