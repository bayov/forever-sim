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
	SpellPowerBuff,
	StrengthBuff,
	SunderRamp,
	TotemWeaponBuff,
	TristateEffect,
	WeaponImbue,
} from '../core/proto/common.js';
import { PaladinAura, PaladinOptions as ProtectionPaladinOptions, PaladinSeal } from '../core/proto/paladin.js';
import { SavedTalents } from '../core/proto/ui.js';
import APLBasicProtectionJson from './apls/basic_prot.apl.json';
import APLP5ProtJson from './apls/p5prot.apl.json';
import Level60BalancedAPLJson from './apls/level60_balanced.apl.json';
import Level60MaxSurvivalAPLJson from './apls/level60_max_survival.apl.json';
import Level60MaxThreatAPLJson from './apls/level60_max_threat.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import Level60Phase1SyntheticGear from './gear_sets/level60_phase1_synthetic.gear.json';
import Level60Phase1SyntheticMaxThreatGear from './gear_sets/level60_phase1_synthetic_max_threat.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);

export const GearLevel60Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic)', Level60Phase1SyntheticGear);
// Max threat's set: the same items, Greater Agility (+7) on the boots for Greater Stamina (+7). It won at w = 1/3 on
// two disjoint seeds and tied at the other weights (docs/tank_paladin.md, round 3).
export const GearLevel60SyntheticMaxThreat = PresetUtils.makePresetGear(
	'Level 60 Phase 1 (Synthetic, Max threat)',
	Level60Phase1SyntheticMaxThreatGear,
);

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Synthetic, GearLevel60SyntheticMaxThreat, GearBlank],
};

export const DefaultGear = GearBlank;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLP5Prot = PresetUtils.makePresetAPLRotation('P5 Prot', APLP5ProtJson);
export const APLBasicProt = PresetUtils.makePresetAPLRotation('Basic Prot', APLBasicProtectionJson);

// The three Level 60 tank rotations (tools/rotopt -template paladin_tank, docs/tank_paladin.md). Every weight kept
// the template's defaults, so the three are the same rotation today; each keeps its own file for the next search.
export const APLLevel60MaxThreat = PresetUtils.makePresetAPLRotation('Level 60 Max threat', Level60MaxThreatAPLJson);
export const APLLevel60Balanced = PresetUtils.makePresetAPLRotation('Level 60 Balanced', Level60BalancedAPLJson);
export const APLLevel60MaxSurvival = PresetUtils.makePresetAPLRotation('Level 60 Max survival', Level60MaxSurvivalAPLJson);

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60Balanced, APLLevel60MaxThreat, APLLevel60MaxSurvival],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [APLBasicProt, APLP5Prot],
	[Phase.Phase5]: [APLP5Prot, APLP5Prot],
};

export const DefaultAPL = APLPresets[Phase.Phase5][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/classic/talent-calc and copy the numbers in the url.

export const P4ProtTalents = {
	name: 'P4 Prot',
	data: SavedTalents.create({
		talentsString: '-55302102513521531-052',
	}),
};

export const P5ProtTalents = {
	name: 'P5 Prot',
	data: SavedTalents.create({
		talentsString: '-55302102513521531-052',
	}),
};

export const TalentsLevel60 = {
	name: 'Level 60 Protection',
	data: SavedTalents.create({
		talentsString: '04-25300102513501531-5023',
	}),
};

export const TalentsLevel60MaxThreat = {
	name: 'Level 60 Protection (Max threat)',
	data: SavedTalents.create({
		talentsString: '05005-25300102413001531-5023',
	}),
};

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60MaxThreat],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [P4ProtTalents],
	[Phase.Phase5]: [P5ProtTalents],
};

export const DefaultTalents = TalentPresets[Phase.Phase5][0];

// The three Level 60 tank presets (the user's ruling: threat against survival as weights, Balanced the default;
// docs/tank_sim.md). Undead sims best at every weight: Touch of the Grave's drain heals and makes threat.
const Level60TankOptions = { primarySeal: PaladinSeal.Fury, righteousFury: true, aura: PaladinAura.DevotionAura };

export const PresetBuildLevel60MaxThreat = PresetUtils.makePresetBuild('Level 60 Max threat', {
	tooltip: 'Scored with survival weighed at 1/3 (docs/tank_sim.md).',
	gear: GearLevel60SyntheticMaxThreat,
	talents: TalentsLevel60MaxThreat,
	rotation: APLLevel60MaxThreat,
	race: Race.RaceUndead,
	professions: [Profession.Engineering, Profession.Blacksmithing],
	options: Level60TankOptions,
});
export const PresetBuildLevel60Balanced = PresetUtils.makePresetBuild('Level 60 Balanced', {
	tooltip: 'Scored with survival weighed at 1 (docs/tank_sim.md).',
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60Balanced,
	race: Race.RaceUndead,
	professions: [Profession.Engineering, Profession.Blacksmithing],
	options: Level60TankOptions,
});
export const PresetBuildLevel60MaxSurvival = PresetUtils.makePresetBuild('Level 60 Max survival', {
	tooltip: "Scored with survival weighed at 3 (docs/tank_sim.md). Balanced's build: the search at this weight lands on the same one.",
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60MaxSurvival,
	race: Race.RaceUndead,
	professions: [Profession.Engineering, Profession.Blacksmithing],
	options: Level60TankOptions,
});

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = ProtectionPaladinOptions.create({
	// The tank presets' options (the bench's): Seal of Fury, Righteous Fury, Devotion Aura.
	aura: PaladinAura.DevotionAura,
	primarySeal: PaladinSeal.Fury,
	righteousFury: true,
});

export const DefaultConsumes = Consumes.create({
	// The tank bench's consumables (tools/tankbench/req/paladin_protection_*.json): every tank takes the same
	// armor and health elixirs (tank_fairness-20261104-1), the flask per spec.
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	armorElixir: ArmorElixir.ElixirOfThePhalanx,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MajorManaPotion,
	dragonBreathChili: true,
	food: Food.FoodSmokedDesertDumpling,
	flask: Flask.FlaskOfTheTitans,
	healthElixir: HealthElixir.ElixirOfGreaterFortitude,
	mainHandImbue: WeaponImbue.DenseSharpeningStone,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
});

// The tank bench's buffs and debuffs (tools/tankbench/req/paladin_protection_*.json; tank_fairness-20261104-6):
// loading a preset gives the build the bench ranks.
export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	fengusFerocity: false,
	moldarsMoxie: false,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	bloodPact: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
	graceOfAirTotem: TristateEffect.TristateEffectRegular,
	arcaneBrilliance: true,
	divineSpirit: true,
	battleShout: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
	leaderOfThePack: true,
	moonkinAura: true,
	shadowProtection: true,
	frostResistanceAura: true,
	fireResistanceAura: true,
	thorns: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	totemWeaponBuff: TotemWeaponBuff.TotemWeaponBuffWindfury,
});

export const DefaultDebuffs = Debuffs.create({
	judgementOfWisdom: true,
	judgementOfLight: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	faerieFire: true,
	curseOfElements: true,
	sunderRamp: SunderRamp.create({ tank: true }),
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	demoralizingRoar: TristateEffect.TristateEffectImproved,
	demoralizingShout: TristateEffect.TristateEffectImproved,
	thunderClap: TristateEffect.TristateEffectImproved,
	insectSwarm: true,
	scorpidSting: true,
});

export const OtherDefaults = {
	// A tank faces the boss: the boss parries and blocks its attacks, and its parries haste the boss's swing.
	inFrontOfTarget: true,
	distanceFromTarget: 5, // Max melee range
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
};
