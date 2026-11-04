import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	ArmorElixir,
	AttackPowerBuff,
	Consumes,
	Debuffs,
	Encounter,
	Flask,
	Food,
	HealingModel,
	HealthElixir,
	IndividualBuffs,
	MobType,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SpellPowerBuff,
	Stat,
	StrengthBuff,
	SunderRamp,
	TotemWeaponBuff,
	TristateEffect,
	UnitReference,
	UnitReference_Type,
	WeaponImbue,
} from '../core/proto/common.js';
import { PaladinAura, PaladinOptions as ProtectionPaladinOptions, PaladinSeal } from '../core/proto/paladin.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
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

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = ProtectionPaladinOptions.create({
	// The tank presets' options (the bench's): Seal of Fury, Righteous Fury, Devotion Aura.
	aura: PaladinAura.DevotionAura,
	primarySeal: PaladinSeal.Fury,
	righteousFury: true,
});

// Round 4 (2026-11-04, orchestration/tank_handoff/paladin/round4, docs/tank_paladin.md "Round 4 consumables"): the
// standard consumable set (the bear's and the warrior's: no Zanza, Rumsey Rum, Blasted Lands buff or Goblin Sapper,
// docs/tank_druid.md call 10), searched per preset. Balanced is the default.
export const BalancedConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	armorElixir: ArmorElixir.ElixirOfThePhalanx,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MajorManaPotion,
	dragonBreathChili: true,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodNightfinSoup,
	healthElixir: HealthElixir.ElixirOfGreaterFortitude,
	mainHandImbue: WeaponImbue.ElementalSharpeningStone,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
});

// Max threat takes Balanced's consumables.
export const MaxThreatConsumes = BalancedConsumes;

// Max survival takes the Flask of Natural Precision (+5% expertise in the natural-flask zone) and Dirge's Kickin'
// Chimaerok Chops (25 Stamina).
export const MaxSurvivalConsumes = Consumes.create({
	...BalancedConsumes,
	flask: Flask.FlaskOfNaturalPrecision,
	food: Food.FoodDirgesKickChimaerokChops,
});

export const DefaultConsumes = BalancedConsumes;

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

///////////////////////////////////////////////////////////////////////////
//                                Encounter
///////////////////////////////////////////////////////////////////////////

// The tank bench's boss (docs/tank_sim.md, The bench): the shared Level 60 boss (level 63, 805 attack power, a 2.0 sec
// swing) on the paladin, with the stand-in healer at 600 HPS every 2 sec (1 to 3), a 3 minute fight. The presets were
// searched on the Anniversary encounters (tools/encounters), which the UI doesn't load; this is the bench's own default
// fight. The player's reaction time is the UI default, 200 ms, as the requests have.
const bench_encounter = (name: string, consumes: Consumes) =>
	PresetUtils.makePresetEncounter(
		name,
		Encounter.create({
			duration: 180,
			durationVariation: 20,
			...PresetUtils.LEVEL_60_ZONE,
			targets: [
				{
					id: 213336,
					name: 'Level 60',
					level: 63,
					mobType: MobType.MobTypeUnknown,
					stats: new Stats().withStat(Stat.StatAttackPower, 805).withStat(Stat.StatArmor, 3731).withStat(Stat.StatHealth, 127393).asArray(),
					minBaseDamage: 3000,
					damageSpread: 0.3333,
					swingSpeed: 2,
					parryHaste: true,
				},
			],
		}),
		{
			tanks: [UnitReference.create({ type: UnitReference_Type.Player, index: 0 })],
			healingModel: HealingModel.create({ hps: 600, cadenceSeconds: 2, cadenceVariation: 1 }),
			raidBuffs: DefaultRaidBuffs,
			debuffs: DefaultDebuffs,
			buffs: DefaultIndividualBuffs,
			consumes,
			inFrontOfTarget: true,
		},
	);
export const EncounterLevel60Tank = bench_encounter('Generic boss, tanking', BalancedConsumes);
export const EncounterLevel60TankMaxThreat = bench_encounter('Generic boss, tanking (Max threat)', MaxThreatConsumes);
export const EncounterLevel60TankMaxSurvival = bench_encounter('Generic boss, tanking (Max survival)', MaxSurvivalConsumes);

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The three Level 60 tank presets (the user's ruling: threat against survival as weights, Balanced the default;
// docs/tank_sim.md). Undead sims best at every weight: Touch of the Grave's drain heals and makes threat.
const Level60TankOptions = { primarySeal: PaladinSeal.Fury, righteousFury: true, aura: PaladinAura.DevotionAura };

export const PresetBuildLevel60MaxThreat = PresetUtils.makePresetBuild('Level 60 Max threat', {
	tooltip: 'Scored with survival weighed at 1/3 (docs/tank_sim.md). Consumables: Flask of Supreme Power, Nightfin Soup, Elemental Sharpening Stone (+319 over round 3).',
	gear: GearLevel60SyntheticMaxThreat,
	talents: TalentsLevel60MaxThreat,
	rotation: APLLevel60MaxThreat,
	race: Race.RaceUndead,
	professions: [Profession.Engineering, Profession.Blacksmithing],
	options: Level60TankOptions,
	encounter: EncounterLevel60TankMaxThreat,
});
export const PresetBuildLevel60Balanced = PresetUtils.makePresetBuild('Level 60 Balanced', {
	tooltip: 'Scored with survival weighed at 1 (docs/tank_sim.md). Consumables: Flask of Supreme Power, Nightfin Soup, Elemental Sharpening Stone (+198 over round 3).',
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60Balanced,
	race: Race.RaceUndead,
	professions: [Profession.Engineering, Profession.Blacksmithing],
	options: Level60TankOptions,
	encounter: EncounterLevel60Tank,
});
export const PresetBuildLevel60MaxSurvival = PresetUtils.makePresetBuild('Level 60 Max survival', {
	tooltip: "Scored with survival weighed at 3 (docs/tank_sim.md). Balanced's gear, talents and rotation (the search at this weight lands on the same ones); the consumables differ: Flask of Natural Precision, Dirge's Kickin' Chimaerok Chops, Elemental Sharpening Stone (+142 over round 3).",
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60MaxSurvival,
	race: Race.RaceUndead,
	professions: [Profession.Engineering, Profession.Blacksmithing],
	options: Level60TankOptions,
	encounter: EncounterLevel60TankMaxSurvival,
});
