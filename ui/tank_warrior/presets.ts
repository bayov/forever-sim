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
	MobType,
	Potions,
	Profession,
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
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import { TankWarrior_Options as TankWarriorOptions, WarriorShout, WarriorStance } from '../core/proto/warrior.js';
import APLNoReckJSON from './apls/dps_no_reck.apl.json';
import APLReckJSON from './apls/dps_reck.apl.json';
import Level60BalancedAPLJSON from './apls/level60_balanced.apl.json';
import Level60MaxSurvivalAPLJSON from './apls/level60_max_survival.apl.json';
import Level60MaxThreatAPLJSON from './apls/level60_max_threat.apl.json';
import ProtectionAPLJSON from './apls/protection.apl.json';
import Level60Phase1SyntheticGear from './gear_sets/level60_phase1_synthetic.gear.json';
import Level60Phase1SyntheticArmorTrinketGear from './gear_sets/level60_phase1_synthetic_armor_trinket.gear.json';
import P0BISGear from './gear_sets/p0.bis.gear.json';
import P0BISSwordBoardGear from './gear_sets/p0.bis_sword_board.gear.json';
import P1BISGear from './gear_sets/p1.bis.gear.json';

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
// Max threat and Balanced wear it: the attack power on-use beats the armor one by 210 and 180 (round 3).
export const GearLevel60Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic)', Level60Phase1SyntheticGear);
// The same set with the armor on-use trinket (990617) in place of the attack power one: Max survival's, where it
// wins by 13 to 16 ± 5 when held for the last 35 sec (round 3, two seeds): all of it on Broodlord; without Broodlord the attack
// power trinket would win by about 70 (tank_rotations-20261104-11, a call for the user).
export const GearLevel60SyntheticArmorTrinket = PresetUtils.makePresetGear(
	'Level 60 Phase 1 (Synthetic, armor trinket)',
	Level60Phase1SyntheticArmorTrinketGear,
);

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Synthetic, GearLevel60SyntheticArmorTrinket, GearP0BISSwordBoard, GearP0BIS, GearP1BIS],
};

export const DefaultGear = GearLevel60Synthetic;

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
	// The stone goes on the main hand: a shield can't hold one. The Elemental stone beats the Dense by 14 to 21.
	mainHandImbue: WeaponImbue.ElementalSharpeningStone,
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
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60Balanced,
	encounter: EncounterLevel60Tank,
	race: Race.RaceUndead,
});
export const PresetBuildLevel60MaxSurvival = PresetUtils.makePresetBuild('Level 60 Max survival', {
	tooltip: 'Scored with survival weighed at 3 (docs/tank_sim.md). Balanced\'s talents and rotation (but Sunder Armor as filler from 30 Rage, not 25), Dirge\'s chops, the Greater Stoneshield Potion and the Flask of the Titans, with the armor trinket held for the last 35 sec (it wins only on Broodlord, docs/tank_warrior.md).',
	gear: GearLevel60SyntheticArmorTrinket,
	talents: TalentsLevel60,
	rotation: APLLevel60MaxSurvival,
	encounter: EncounterLevel60TankMaxSurvival,
	race: Race.RaceUndead,
});
