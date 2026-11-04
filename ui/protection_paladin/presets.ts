import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	Alcohol,
	ArmorElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Explosive,
	FirePowerBuff,
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
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { Blessings, PaladinAura, PaladinOptions as ProtectionPaladinOptions, PaladinSeal } from '../core/proto/paladin.js';
import { SavedTalents } from '../core/proto/ui.js';
import APLBasicProtectionJson from './apls/basic_prot.apl.json';
import APLP5ProtJson from './apls/p5prot.apl.json';
import Level60BalancedAPLJson from './apls/level60_balanced.apl.json';
import Level60MaxSurvivalAPLJson from './apls/level60_max_survival.apl.json';
import Level60MaxThreatAPLJson from './apls/level60_max_threat.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import Level60Phase1SyntheticGear from './gear_sets/level60_phase1_synthetic.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);

export const GearLevel60Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic)', Level60Phase1SyntheticGear);

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Synthetic, GearBlank],
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
	gear: GearLevel60Synthetic,
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
	tooltip: 'Scored with survival weighed at 3 (docs/tank_sim.md).',
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
	aura: PaladinAura.RetributionAura,
	primarySeal: PaladinSeal.Righteousness,
	personalBlessing: Blessings.BlessingOfSanctuary,
	righteousFury: true,
});

export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	healthElixir: HealthElixir.ElixirOfFortitude,
	armorElixir: ArmorElixir.ElixirOfSuperiorDefense,
	defaultPotion: Potions.GreaterStoneshieldPotion,
	dragonBreathChili: true,
	food: Food.FoodTenderWolfSteak,
	flask: Flask.FlaskOfTheTitans,
	firePowerBuff: FirePowerBuff.ElixirOfGreaterFirepower,
	fillerExplosive: Explosive.ExplosiveDenseDynamite,
	//mainHandImbue: WeaponImbue.WildStrikes,
	//offHandImbue: WeaponImbue.MagnificentTrollshine,

	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.ROIDS,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredDemonicRune,
	alcohol: Alcohol.AlcoholRumseyRumBlackLabel,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	fengusFerocity: false,
	moldarsMoxie: false,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfRecklessness: true,
	faerieFire: true,
	giftOfArthas: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	improvedScorch: true,
});

export const OtherDefaults = {
	distanceFromTarget: 5, // Max melee range
	profession1: Profession.Blacksmithing,
	profession2: Profession.Engineering,
};
