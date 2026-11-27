import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	Alcohol,
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
	LevelingOptions,
	MobType,
	Potions,
	Profession,
	PseudoStat,
	Race,
	RaidBuffs,
	SapperExplosive,
	Stat,
	StrengthBuff,
	TotemWeaponBuff,
	TristateEffect,
	UnitReference,
	UnitReference_Type,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { LevelingPlanSettings, LevelingRoute, SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import { TankWarrior_Options as TankWarriorOptions, WarriorShout, WarriorStance } from '../core/proto/warrior.js';
import APLNoReckJSON from './apls/dps_no_reck.apl.json';
import APLReckJSON from './apls/dps_reck.apl.json';
import LevelingPlanJSON from './apls/leveling_plan.json';
import LevelingPlanArmsJSON from './apls/leveling_plan_arms.json';
import LevelingPlanArmsProtDualWieldJSON from './apls/leveling_plan_arms_prot_dw.json';
import LevelingPlanDualWieldJSON from './apls/leveling_plan_dual_wield.json';
import LevelingPlanFuryProtJSON from './apls/leveling_plan_fury_prot.json';
import LevelingPlanProtJSON from './apls/leveling_plan_prot.json';
import Level60BalancedAPLJSON from './apls/level60_balanced.apl.json';
import Level60MaxSurvivalAPLJSON from './apls/level60_max_survival.apl.json';
import Level60MaxThreatAPLJSON from './apls/level60_max_threat.apl.json';
import ProtectionAPLJSON from './apls/protection.apl.json';
import Level60Phase1SyntheticGear from './gear_sets/level60_phase1_synthetic.gear.json';
import Level60Phase1SyntheticArmorTrinketGear from './gear_sets/level60_phase1_synthetic_armor_trinket.gear.json';
import Level60Phase1SyntheticBalancedGear from './gear_sets/level60_phase1_synthetic_balanced.gear.json';
import P0BISGear from './gear_sets/p0.bis.gear.json';
import P0BISSwordBoardGear from './gear_sets/p0.bis_sword_board.gear.json';
import P1BISGear from './gear_sets/p1.bis.gear.json';
import LevelingWeightsDualWieldJSON from './leveling_weights_dual_wield.json';
import LevelingWeightsShieldJSON from './leveling_weights_shield.json';
import LevelingWeightsShieldFuryProtJSON from './leveling_weights_shield_fury_prot.json';
import LevelingWeightsTwoHandJSON from './leveling_weights_two_hand.json';
import LevelingRouteJSON from '../warrior/leveling_route.json';

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearP0BIS = PresetUtils.makePresetGear('Pre-BiS', P0BISGear);
export const GearP1BIS = PresetUtils.makePresetGear('P1 BiS', P1BISGear);
// The Pre-BiS armor with a sword and a shield (Mirah's Song, Draconian Deflector), so Shield Slam,
// Shield Block and Revenge have a shield to use. The golden suite (sim/warrior/tank_warrior) uses it.
export const GearP0BISSwordBoard = PresetUtils.makePresetGear('Pre-BiS (Sword and Board)', P0BISSwordBoardGear);

// The synthetic Level 60 Phase 1 Protection set (standards 36, tools/database/forever_synthetic_items_warrior_tank.go):
// 106 Defense from the items, 110 block value, 239 Stamina, 8% hit (the cap), the tank stat trinket and the attack power on-use,
// Superior Deflection (Defense +9) on the wrists and Agility +5 on the cloak: they beat Superior Stamina and Superior Defense
// at every weight or tie (docs/tank_warrior.md, Round 3 gear A/Bs). Greater Agility on the boots at all three weights, and
// Superior Strength on the wrists at Max threat and Balanced (tank_gear-20261112-1, -2; the armor trinket set keeps Deflection).
// Max threat wears it, with Lesser Arcanum of Voracity Strength on the head and legs (tank_fairness-20261112-2: +20.5 / +20.7
// over Constitution at 10,000 iterations, seeds 1 / 2). The attack power on-use beats the armor one by 210 and 180 (round 3).
export const GearLevel60Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic)', Level60Phase1SyntheticGear);
// The same set with Lesser Arcanum of Voracity Agility on the head and legs: Balanced's (+8.2 / +10.5 over Constitution,
// ahead of Strength's +6.0 / +5.8; tank_fairness-20261112-2).
export const GearLevel60SyntheticBalanced = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic, Balanced)', Level60Phase1SyntheticBalancedGear);
// The same set with the armor on-use trinket (990617) in place of the attack power one: Max survival's, where it
// wins by 13 to 16 ± 5 when held for the last 35 sec (round 3, two seeds): all of it on Broodlord; without Broodlord the attack
// power trinket would win by about 70 (tank_rotations-20261104-11, a call for the user). Necklace Agility in place of
// Strength (+13.8 / +13.2, tank_fairness-20261112-2).
export const GearLevel60SyntheticArmorTrinket = PresetUtils.makePresetGear(
	'Level 60 Phase 1 (Synthetic, armor trinket)',
	Level60Phase1SyntheticArmorTrinketGear,
);

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Synthetic, GearLevel60SyntheticBalanced, GearLevel60SyntheticArmorTrinket, GearP0BISSwordBoard, GearP0BIS, GearP1BIS],
};

export const DefaultGear = GearLevel60SyntheticBalanced;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const AplReck = PresetUtils.makePresetAPLRotation('DPS (With Reck)', APLReckJSON);
export const APLNoReck = PresetUtils.makePresetAPLRotation('DPS (No Reck)', APLNoReckJSON);
// A plain Protection rotation in Defensive Stance: Shield Block whenever it is down, Shield Slam on
// cooldown, Revenge, Sunder Armor to 5 stacks, Heroic Strike from 50 Rage and Sunder Armor as the
// filler from 40, Bloodrage under 30 Rage.
export const APLProtection = PresetUtils.makePresetAPLRotation('Protection', ProtectionAPLJSON);
// The three tank presets (docs/tank_warrior.md, Presets): the rotopt template warrior_tank searched on the six
// Anniversary encounters at the survival weights 1/3, 1 and 3. None reads the tank's health; Shield Wall goes on a
// timer and Last Stand is left to the player.
export const APLLevel60MaxThreat = PresetUtils.makePresetAPLRotation('Level 60 Max threat', Level60MaxThreatAPLJSON);
export const APLLevel60Balanced = PresetUtils.makePresetAPLRotation('Level 60 Balanced', Level60BalancedAPLJSON);
export const APLLevel60MaxSurvival = PresetUtils.makePresetAPLRotation('Level 60 Max survival', Level60MaxSurvivalAPLJSON);

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60Balanced, APLLevel60MaxThreat, APLLevel60MaxSurvival, APLProtection, APLNoReck, AplReck],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/classic/talent-calc and copy the numbers in the url.

export const TalentsP1Prot = PresetUtils.makePresetTalents('Protection', SavedTalents.create({ talentsString: '35--251533120331221351' }));

// The talent searches' builds (docs/tank_warrior.md, Presets), on the six scoring encounters.
// Balanced and Max survival (both searches end on the same talents): Deflection 5; Cruelty, Unbridled Wrath and Enrage 5/5; Protection 31 with Shield Slam, no Last Stand or Vanguard.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60 Protection', SavedTalents.create({ talentsString: '05-05050005-250530020310001351' }));
// Max threat: Deflection 5, no Blood Craze (round 5, on the full raid consumables), Anticipation 3, Improved Sunder
// Armor 3 (Improved Heroic Strike went once Unbridled Wrath stopped proccing on Heroic Strike).
export const TalentsLevel60MaxThreat = PresetUtils.makePresetTalents(
	'Level 60 Protection (Max threat)',
	SavedTalents.create({ talentsString: '05-05050005-250330020330001351' }),
);
export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60MaxThreat, TalentsP1Prot],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options Presets
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = TankWarriorOptions.create({
	// The bench's: none (the requests carry no queueDelay). 250 cost 2% at Max threat (tank_fairness-20261104-11).
	queueDelay: 0,
	startingRage: 0,
	shout: WarriorShout.WarriorShoutBattle,
	stance: WarriorStance.WarriorStanceDefensive,
});

export const DefaultConsumes = Consumes.create({
	// The tank bench's consumables (tools/tankbench/req/warrior_protection_*.json): every tank takes the same
	// armor and health elixirs (tank_fairness-20261104-1), the flask per spec.
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	armorElixir: ArmorElixir.ElixirOfThePhalanx,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MightyRagePotion,
	dragonBreathChili: true,
	food: Food.FoodSmokedDesertDumpling,
	// Round 4 (docs/tank_warrior.md, Round 4 consumables): the Flask of Natural Precision, +5% expertise in the
	// Natural flasks' zones (the encounter's naturalFlaskZone), beats the Titans by 380 to 660.
	flask: Flask.FlaskOfNaturalPrecision,
	healthElixir: HealthElixir.ElixirOfGreaterFortitude,
	// The stone goes on the main hand: a shield can't hold one. The Dense Sharpening Stone is the best stone a phase-1 player can get (the Elemental is raid-only, tank_fairness-20261112-1); the Dense Weightstone is -12.
	mainHandImbue: WeaponImbue.DenseSharpeningStone,
	strengthBuff: StrengthBuff.JujuPower,
	// Round 4's raid set (docs/tank_warrior.md, Round 4 consumables, the full raid set): Spirit of Zanza (+78 to +161 over
	// none; Balanced and Max survival), Rumsey Rum Black Label (+11 to +48) and the Goblin Sapper Charge (Engineering;
	// +181 to +212: its threat at the pull; the rotations throw it after the trinkets). Thrown explosives lose (their
	// 1 sec cast at the pull costs 270 to 430), so none is taken (standards 11).
	zanzaBuff: ZanzaBuff.SpiritOfZanza,
	alcohol: Alcohol.AlcoholRumseyRumBlackLabel,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
});

// Max threat takes R.O.I.D.S. in the Zanza slot (+30 over Spirit of Zanza, +19 to +23 over Ground Scorpok Assay).
export const MaxThreatConsumes = Consumes.create({
	...DefaultConsumes,
	zanzaBuff: ZanzaBuff.ROIDS,
});

// Max survival eats Dirge's Kickin' Chimaerok Chops (+41 to +46 over the Dumpling) and drinks the Greater
// Stoneshield Potion for the Mighty Rage Potion (+83 to +92) at weight 3. Since client 70862 halved the Natural
// flasks (30 Stamina), the Flask of the Titans wins here (+46.5 / +41.8, 10,000 iterations, seeds 1 / 2;
// tank_fairness-20261111-1); Natural Precision still wins at Max threat and Balanced.
export const MaxSurvivalConsumes = Consumes.create({
	...DefaultConsumes,
	flask: Flask.FlaskOfTheTitans,
	food: Food.FoodDirgesKickChimaerokChops,
	defaultPotion: Potions.GreaterStoneshieldPotion,
});

// The raid's buffs and the boss's debuffs are the tank bench's (tools/tankbench/req/warrior_protection_*.json), so the
// UI presets load what the bench scores (tank_fairness-20261104-6).
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
});

// Not Thunder Clap, Demoralizing Shout/Roar or the tank Sunder ramp: tankscore.OwnDebuffs strips them for the warrior
// because its own rotation keeps them up (tools/tankscore/tankscore.go), so the bench never has them.
export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	curseOfRecklessness: true,
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	faerieFire: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	scorpidSting: true,
});

export const OtherDefaults = {
	// A tank faces the boss: the boss parries and blocks its attacks, and its parries haste the boss's swing.
	inFrontOfTarget: true,
	// The bench's professions and race (tank_fairness-20261104-6).
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceUndead,
};

// The tank bench's boss (docs/tank_sim.md, The bench): the shared Level 60 boss (level 63, 805 attack power, a 2.0
// sec swing) on the warrior, with the stand-in healer at 600 HPS every 2 sec (1 to 3), a 3 minute fight, in a
// zone where the Natural flasks work. The presets were searched on the Anniversary encounters (tools/encounters),
// which the UI doesn't load; this is the bench's own default fight.
export const EncounterLevel60Tank = PresetUtils.makePresetEncounter(
	'Generic boss, tanking',
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
		consumes: DefaultConsumes,
		inFrontOfTarget: true,
	},
);

// The same fight with Max threat's consumables.
export const EncounterLevel60TankMaxThreat = PresetUtils.makePresetEncounter('Generic boss, tanking (Max threat)', EncounterLevel60Tank.encounter, {
	tanks: EncounterLevel60Tank.tanks,
	healingModel: EncounterLevel60Tank.healingModel,
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: MaxThreatConsumes,
	inFrontOfTarget: true,
});

// The same fight with Max survival's consumables.
export const EncounterLevel60TankMaxSurvival = PresetUtils.makePresetEncounter('Generic boss, tanking (Max survival)', EncounterLevel60Tank.encounter, {
	tanks: EncounterLevel60Tank.tanks,
	healingModel: EncounterLevel60Tank.healingModel,
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: MaxSurvivalConsumes,
	inFrontOfTarget: true,
});

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

export const PresetBuildTanky = PresetUtils.makePresetBuild('Tanky', { gear: GearP0BISSwordBoard, talents: TalentsP1Prot, rotation: APLProtection });

// The three Level 60 tank presets (the user's ruling: threat against survival as weights, Balanced the default).
export const PresetBuildLevel60MaxThreat = PresetUtils.makePresetBuild('Level 60 Max threat', {
	tooltip: 'Scored with survival weighed at 1/3 (docs/tank_sim.md). Its own talents, no Thunder Clap, Shield Wall in the last 13 sec.',
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60MaxThreat,
	rotation: APLLevel60MaxThreat,
	encounter: EncounterLevel60TankMaxThreat,
	race: Race.RaceUndead,
});
export const PresetBuildLevel60Balanced = PresetUtils.makePresetBuild('Level 60 Balanced', {
	tooltip: 'Scored with survival weighed at 1 (docs/tank_sim.md). Shield Wall in the last 17 sec.',
	gear: GearLevel60SyntheticBalanced,
	talents: TalentsLevel60,
	rotation: APLLevel60Balanced,
	encounter: EncounterLevel60Tank,
	race: Race.RaceUndead,
});
export const PresetBuildLevel60MaxSurvival = PresetUtils.makePresetBuild('Level 60 Max survival', {
	tooltip: 'Scored with survival weighed at 3 (docs/tank_sim.md). Balanced\'s talents and rotation (but Sunder Armor as filler from 30 Rage, not 25, and Shield Wall in the last 13 sec, not 17), Dirge\'s chops, the Greater Stoneshield Potion and the Flask of the Titans, with the armor trinket held for the last 35 sec (it wins only on Broodlord, docs/tank_warrior.md).',
	gear: GearLevel60SyntheticArmorTrinket,
	talents: TalentsLevel60,
	rotation: APLLevel60MaxSurvival,
	encounter: EncounterLevel60TankMaxSurvival,
	race: Race.RaceUndead,
});

///////////////////////////////////////////////////////////////////////////
//                                 Leveling
///////////////////////////////////////////////////////////////////////////

// The leveling sessions play this spec for every warrior, two-hand and dual wield too: it's the
// DPS spec's warrior with Shield Block and Shield Wall on top (docs/leveling_plan.md, A plan
// without the browser). So the 1 to 60 plan lives here.

// Stat weights for the leveling gear plan, by mobs killed in a session, in attack power, every 5
// levels from 5 to 59 (tools/levelingplan -weights, 2026-11-26). One set a weapon style: two-hand
// on the Arms plan, dual wield on the Gore Drinker build, and a shield on Protection. The Leveling
// Gear tab judges items by the set of the plan's style at that level (docs/leveling_warrior.md,
// Leveling 1-60).
type LevelingWeightsFile = { levels: Array<{ level: number; weights: Record<string, number | undefined> }> };

const levelingGearWeights = (file: LevelingWeightsFile, style: string, group: string, tooltip: string) =>
	file.levels.map(({ level, weights }) => {
		const w = (name: string) => weights[name] ?? 0;
		return PresetUtils.makePresetEpWeights(
			`Level ${level} Leveling Gear ${style}`,
			Stats.fromMap(
				{
					[Stat.StatStrength]: w('Strength'),
					[Stat.StatAgility]: w('Agility'),
					[Stat.StatStamina]: w('Stamina'),
					[Stat.StatSpirit]: w('Spirit'),
					[Stat.StatAttackPower]: 1,
					[Stat.StatMeleeHit]: w('MeleeHit'),
					[Stat.StatMeleeCrit]: w('MeleeCrit'),
					[Stat.StatArmor]: w('Armor'),
					[Stat.StatBonusArmor]: w('Armor'),
					[Stat.StatDefense]: w('Defense'),
					[Stat.StatDodge]: w('Dodge'),
					[Stat.StatParry]: w('Parry'),
					[Stat.StatBlock]: w('Block'),
					[Stat.StatBlockValue]: w('BlockValue'),
				},
				{
					[PseudoStat.PseudoStatMainHandDps]: w('MainHandDps'),
					[PseudoStat.PseudoStatOffHandDps]: w('OffHandDps'),
				},
			),
			{ tooltip: `Stat weights for a solo level ${level} ${tooltip}. The Leveling Gear tab judges items by them.`, group },
		);
	});

export const EPLevelingGearPlanTwoHand = levelingGearWeights(
	LevelingWeightsTwoHandJSON,
	'Two-Hand',
	'Leveling Gear Plan Two-Hand',
	'Arms warrior with a two-hander',
);
export const EPLevelingGearPlanDualWield = levelingGearWeights(
	LevelingWeightsDualWieldJSON,
	'Dual Wield',
	'Leveling Gear Plan Dual Wield',
	'Gore Drinker warrior with two weapons',
);
export const EPLevelingGearPlanShield = levelingGearWeights(
	LevelingWeightsShieldJSON,
	'Shield',
	'Leveling Gear Plan Shield',
	'Protection warrior with a one-hander and shield',
);

// The shield weights measured on the Fury Prot build, which its plan takes in place of the ones
// above (LevelingPlanSettings.weights_build).
export const EPLevelingGearPlanShieldFuryProt = levelingGearWeights(
	LevelingWeightsShieldFuryProtJSON,
	'Shield (Fury Prot)',
	'Leveling Gear Plan Shield (Fury Prot)',
	'Fury Prot warrior with a one-hander and shield',
);

export const EPPresets = [...EPLevelingGearPlanTwoHand, ...EPLevelingGearPlanDualWield, ...EPLevelingGearPlanShield, ...EPLevelingGearPlanShieldFuryProt];

// The leveling plan of the Leveling mode (ui/core/leveling_plan.ts), from level 1 to 60, with
// Talented's first talent at 5 (docs/leveling_warrior.md, Leveling 1-60). Protection all the way,
// with a two-hander until Shield Block at 16 and a shield from then on. Its talent order and its
// rotations ("Searched 1" and so on) are the talent and rotation search's. Its talent order is the
// one we chose to play (docs/leveling_warrior.md, The plan).
// tools/levelingplan/configs/undead_warrior_leveling_prot.json makes the same plan for the
// command line.
export const DefaultLevelingPlan = LevelingPlanSettings.fromJson(LevelingPlanProtJSON as any);

// Our RestedXP route for an Undead Warrior (tools/leveling_route): the quest rewards we take on it
// and the dungeons we run, at the levels we're at then. The leveling gear plan takes them in place
// of the ones it makes up.
export const LevelingRouteUndeadWarrior = LevelingRoute.fromJson(LevelingRouteJSON);

// The Leveling mode with the options the plan was measured with: mobs a level below ours, packs of
// up to 12, a camp of mobs with their own places, 1 point of Thrill of Adventure and Enchant Weapon
// - Recovery from 20. No stand-in healer, which the tanking fight has. The fight's duration and
// target only count when we go back to the DPS mode.
export const EncounterLevelingPlan = PresetUtils.makePresetEncounter(
	'Leveling',
	Encounter.create({
		...EncounterLevel60Tank.encounter,
		leveling: LevelingOptions.create({
			plan: true,
			mobLevelOffset: -1,
			maxPackSize: 12,
			campMobs: -1,
			thrillOfAdventure: 1,
			recoveryFromLevel: 20,
		}),
	}),
	{ group: 'Leveling', healingModel: HealingModel.create({}) },
);
const levelingOptions = TankWarriorOptions.create({
	queueDelay: 250,
	startingRage: 0,
	shout: WarriorShout.WarriorShoutBattle,
	stance: WarriorStance.WarriorStanceBattle,
});
const makeLevelingBuild = (name: string, tooltip: string, levelingPlan: LevelingPlanSettings) =>
	PresetUtils.makePresetBuild(name, {
		group: 'Leveling',
		tooltip,
		encounter: EncounterLevelingPlan,
		race: Race.RaceUndead,
		options: levelingOptions,
		levelingPlan,
	});

// The hours are 1 to 60 with Recovery from 20, 8 gear seeds and 400 sessions a level on master of
// 2026-11-26 (docs/leveling_warrior.md, Leveling 1-60). They compare builds, not real time.
export const PresetBuildLevelingProt = makeLevelingBuild(
	'Leveling: Prot',
	'Our 1 to 60 plan: Protection 5/15/31, a two-hander until 16, then a shield. Sim: 96.4h, 26 deaths.',
	DefaultLevelingPlan,
);

// The other builds we compared it with. Arms then DW had the talent and rotation search to 39 and
// hand-tuned rotations from 40. The others have hand-tuned rotations.
// tools/levelingplan/configs/undead_warrior_leveling_*.json make the same plans for the command line.
export const PresetBuildLeveling1to60 = makeLevelingBuild(
	'Leveling: Arms then DW',
	'Arms with a two-hander, then Gore Drinker with two weapons from 44. Sim: 99.5h, 21 deaths.',
	LevelingPlanSettings.fromJson(LevelingPlanJSON as any),
);
// A well-known endgame build (4/31/16), not one people level with.
export const PresetBuildLevelingFuryProt = makeLevelingBuild(
	'Leveling: Fury Prot',
	'The endgame Fury Prot build (4/31/16) with a shield from level 5. Sim: 112.7h, 30 deaths.',
	LevelingPlanSettings.create({ ...LevelingPlanSettings.fromJson(LevelingPlanFuryProtJSON as any), weightsBuild: 'Fury Prot' }),
);
export const PresetBuildLevelingArms = makeLevelingBuild(
	'Leveling: Arms',
	'Arms with a two-hander all the way. Sim: 103.0h, 11 deaths.',
	LevelingPlanSettings.fromJson(LevelingPlanArmsJSON as any),
);
export const PresetBuildLevelingDualWield = makeLevelingBuild(
	'Leveling: Dual Wield',
	'Gore Drinker Fury with two weapons all the way. Sim: 102.5h, 10 deaths.',
	LevelingPlanSettings.fromJson(LevelingPlanDualWieldJSON as any),
);
export const PresetBuildLevelingArmsProtDualWield = makeLevelingBuild(
	'Leveling: Arms, Prot, DW',
	'Arms with a two-hander, Protection with a shield at 30, Gore Drinker with two weapons at 40. Sim: 101.2h, 20 deaths.',
	LevelingPlanSettings.fromJson(LevelingPlanArmsProtDualWieldJSON as any),
);
