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
	SunderRamp,
	TotemWeaponBuff,
	TristateEffect,
	UnitReference,
	UnitReference_Type,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { FeralTankDruid_Options as DruidOptions, FeralTankDruid_Rotation as DruidRotation } from '../core/proto/druid.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import BalancedAPLJSON from './apls/balanced.apl.json';
import MaxSurvivalAPLJSON from './apls/max_survival.apl.json';
import MaxThreatAPLJSON from './apls/max_threat.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import Level60Phase1SyntheticMaxSurvivalGearJSON from './gear_sets/level60_phase1_synthetic_max_survival.gear.json';
import Level60Phase1SyntheticMaxThreatGearJSON from './gear_sets/level60_phase1_synthetic_max_threat.gear.json';

// The Forever bear (docs/tank_druid.md has the audit, the searches and the numbers). Three presets along
// the tank sim's objective, 2 x SafeDPS - w x 900 x healing needed / one healer's HPS (docs/tank_sim.md,
// the user's ruling of 2026-11-02): Max threat (w = 1/3), Balanced (1, the default) and Max survival (3),
// each searched (tools/rotopt -tank, template druid_tank) on the Anniversary encounters without Ragnaros
// (tools/encounters), scored as their mean.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

// The level 60 Phase 1 tank set, synthetic items not in the game (tools/database/
// forever_synthetic_items_druid_tank.go has the budget math): standards 36's bear row (physical
// archetype, about 730 weighted points with the feral two-hander), 254 Stamina, 164 Agility, 125
// Strength, 105 Defense and 3% hit (8% with the flask). The trinkets are the tank stat trinket (23 Defense, 1%
// dodge) and the cat's on-use Strength trinket, which the tank objective scores far over the defensive
// armor one (its opening threat). Idol of the Ursine Twins (Lacerate resets Primal Bite 10% of the
// time) is ahead of Idol of Brutality and Enraged Idol. Enchants (round 3, each A/B'd at the preset's weight,
// docs/tank_druid.md): Agility on the head and legs (Lesser Arcanums of Voracity), the neck, back (ahead of
// Superior Defense's 70 armor), boots and weapon, Greater Stats, Superior Deflection on the wrist (ahead of
// Superior Stamina at Balanced and Max survival) and Superior Strength on the gloves. Balanced wears this set.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Bear', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS for a bear under Forever, made of synthetic items that are not in the game. The Balanced enchants.',
	group: 'Synthetic',
});

// The same set with Gloves Greater Agility in place of Superior Strength: Max survival's (tank_fairness-20261112-2:
// +14.3 / +23.7 at weight 3, 10,000 iterations, seeds 1 / 2).
export const GearLevel60Phase1SyntheticMaxSurvival = PresetUtils.makePresetGear('Bear (Max survival)', Level60Phase1SyntheticMaxSurvivalGearJSON, {
	tooltip: "The synthetic bear set with Max survival's Greater Agility on the gloves.",
	group: 'Synthetic',
});

// The same items with Max threat's enchants: Strength on the head and legs (Lesser Arcanum of Voracity, 8
// Strength), the wrist (Superior Strength) and the weapon (2H Strength, 25), each ahead of the Agility or
// Stamina one at weight 1/3 on two seeds; the neck stays Agility (Strength there is a tie).
export const GearLevel60Phase1SyntheticMaxThreat = PresetUtils.makePresetGear('Bear (Max threat)', Level60Phase1SyntheticMaxThreatGearJSON, {
	tooltip: "The synthetic bear set with Max threat's Strength enchants on the head, legs, wrist and weapon.",
	group: 'Synthetic',
});

export const GearPresets = [GearLevel60Phase1Synthetic, GearLevel60Phase1SyntheticMaxThreat, GearLevel60Phase1SyntheticMaxSurvival];

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The three presets' rotations (tools/rotopt -template druid_tank, each searched at its weight; round 3 on the
// unmultiplied mana-gain threat moved only Max threat's Enrage). All three put Primal Bite on cooldown, keep Lacerate at 5 stacks, keep Demoralizing Roar
// up just below that line (refreshed under 3 sec), use Barkskin on its timer from 15 sec in, in the filler's
// place, shift into Dire Bear Form 1.5 sec before the pull and Enrage in the fight. Balanced: Lacerate as the
// filler, Maul at 20 Rage, Enrage under 40 Rage and Frenzied Regeneration at 100 Rage. Max threat: Balanced
// without Frenzied Regeneration, with Enrage under 70 Rage and Barkskin from 20 sec in (re-tuned on the Windfury
// Totem rule of 2026-11-12, docs/tank_druid.md). Max survival: Balanced with Lacerate kept up but not the filler,
// Maul at 40 Rage, Frenzied Regeneration at 80 Rage and Swipe as the filler at 80 Rage (Windfury re-tune), so FR
// fires every cooldown. No preset enrages before the pull any more
// (each weight's best build enrages in the fight).
export const APLMaxThreat = PresetUtils.makePresetAPLRotation('Max threat', MaxThreatAPLJSON);
export const APLBalanced = PresetUtils.makePresetAPLRotation('Balanced', BalancedAPLJSON);
export const APLMaxSurvival = PresetUtils.makePresetAPLRotation('Max survival', MaxSurvivalAPLJSON);

export const APLPresets = [APLBalanced, APLMaxThreat, APLMaxSurvival];

export const DefaultAPL = APLBalanced;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent search's build at all three weights (tools/rotopt -talent-search on the scoring encounters):
// Ferocity, Feral Swiftness, Thick Hide, Shredding Attacks (Lacerate 12 Rage), Savage Fury, Sharpened
// Claws, Predatory Strikes, Lacerate, Blood Frenzy, Leader of the Pack, Natural Instinct, Rend and Tear,
// Natural Reaction and Primal Bite, with Heart of the Wild and Naturalist in Restoration. Furor scores
// nothing now (a hair under, at every weight), so its 2 points went to Brutal Impact (Bash, which the
// scoring fights don't use); Feral Instinct's 2 are Swipe's, for packs.
export const TalentsBear = PresetUtils.makePresetTalents('Bear', SavedTalents.create({ talentsString: '-5022233202310212551-055' }));
// Max survival (Windfury re-tune, 2026-11-12): Swipe is its filler at 80 Rage, so Feral Instinct takes Brutal
// Impact's second point (+2.4 at 10,000 on two seeds).
export const TalentsBearMaxSurvival = PresetUtils.makePresetTalents(
	'Bear (Max survival)',
	SavedTalents.create({ talentsString: '-5023133202310212551-055' }),
);

export const TalentPresets = [TalentsBear, TalentsBearMaxSurvival];

export const DefaultTalents = TalentsBear;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// The bear starts the pull with no Rage (the requests' startingRage 0); its rotation shifts into Dire
// Bear Form 1.5 sec before the pull (the presets take no Furor, which scores nothing for it).
export const DefaultOptions = DruidOptions.create({
	innervateTarget: UnitReference.create(),
	startingRage: 0,
});

// The rotation's old simple settings, which the APL presets replace.
export const DefaultRotation = DruidRotation.create({
	maulRageThreshold: 30,
	maintainDemoralizingRoar: true,
	lacerateTime: 4.0,
});

// Forever launch Phase 1 consumables, raid buffs included (standards 11), each the best of its slot for the tank
// objective at each preset's weight (docs/tank_druid.md): the Flask of Natural Precision (+5% expertise in Hyjal and
// the Barrow Deeps, standards 11; round 5: 205 to 290 over the Flask of Natural Accuracy, which beat the Titans by 125
// to 300), Elixir of the Phalanx (400 health, 500 armor), Elixir of Greater Fortitude, Elixir of the Mongoose (45
// over Greater Agility), Juju Power and Might (ahead of Giants and Winterfall Firewater), Smoked Desert Dumpling
// (ahead of Dirge's Stamina chops at Max threat and Balanced, and of Grilled Squid), Dragonbreath Chili, a Dense
// Weightstone on the mace, and the Mighty Rage Potion at the pull. Round 4: the Goblin Sapper Charge (Engineering;
// usable in Bear Form, tank_mechanics-15) gains 180 to 300; every filler explosive (Goblin Land Mine, EZ-Thro Dark
// Bomb or Thorium Grenade) loses 4 to 18 after it, so there is none. Round 5 (Consumables, round 5): Rumsey Rum
// Black Label (+6 to +76) and Spirit of Zanza at Balanced and Max survival (R.O.I.D.S. at Max threat; the Blasted
// Lands buffs share the Zanza slot).
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	alcohol: Alcohol.AlcoholRumseyRumBlackLabel,
	armorElixir: ArmorElixir.ElixirOfThePhalanx,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MightyRagePotion,
	dragonBreathChili: true,
	flask: Flask.FlaskOfNaturalPrecision,
	food: Food.FoodSmokedDesertDumpling,
	healthElixir: HealthElixir.ElixirOfGreaterFortitude,
	mainHandImbue: WeaponImbue.DenseWeightstone,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.SpiritOfZanza,
});

// Max threat takes R.O.I.D.S. for Spirit of Zanza (+41).
export const MaxThreatConsumes = Consumes.create({
	...DefaultConsumes,
	zanzaBuff: ZanzaBuff.ROIDS,
});

// Max survival eats Dirge's Kickin' Chimaerok Chops (25 Stamina): +80 to +90 over the Dumpling at weight 3. Round 5: the
// Greater Stoneshield Potion with its APL's potion action (+200 over the Mighty Rage Potion; autocast never drinks a
// defensive potion, so round 3's -470 to -620 was the Rage potion's loss alone). Since client 70862 halved the
// Natural flasks (30 Stamina), the Flask of the Titans wins here (+115.9 / +111.6, 10,000 iterations, seeds 1 / 2;
// tank_fairness-20261111-1); Natural Precision still wins at Max threat and Balanced.
export const MaxSurvivalConsumes = Consumes.create({
	...DefaultConsumes,
	flask: Flask.FlaskOfTheTitans,
	defaultPotion: Potions.GreaterStoneshieldPotion,
	food: Food.FoodDirgesKickChimaerokChops,
});

// The tank bench's raid (tools/tankbench/req/druid_bear*.json): the Level 60 raid without Salvation (a
// tank doesn't take it), without the raid's Demoralizing Roar and Shout (the bear keeps its own Roar
// up, tankscore.OwnDebuffs), the bear in a melee group.
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
	sunderRamp: SunderRamp.create({ tank: true }),
	thunderClap: TristateEffect.TristateEffectImproved,
});

export const OtherDefaults = {
	// A tank faces the boss: the boss parries and blocks its attacks, and its parries haste the boss's swing.
	inFrontOfTarget: true,
	distanceFromTarget: 5,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceNightElf,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The tank bench's boss (docs/tank_sim.md, The bench): the shared Level 60 boss (level 63, 805 attack
// power, a 2.0 sec swing) on the bear, with the stand-in healer at 600 HPS every 2 sec (1 to 3), a 3
// minute fight. The presets were searched on the Anniversary encounters (tools/encounters), which the
// UI doesn't load; this is the bench's own default fight.
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

// The same fight with Max threat's consumables (R.O.I.D.S. for Spirit of Zanza).
export const EncounterLevel60TankMaxThreat = PresetUtils.makePresetEncounter('Generic boss, tanking (Max threat)', EncounterLevel60Tank.encounter, {
	tanks: EncounterLevel60Tank.tanks,
	healingModel: EncounterLevel60Tank.healingModel,
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: MaxThreatConsumes,
	inFrontOfTarget: true,
});

// The same fight with Max survival's consumables (Dirge's chops for the Dumpling, the Greater Stoneshield Potion).
export const EncounterLevel60TankMaxSurvival = PresetUtils.makePresetEncounter('Generic boss, tanking (Max survival)', EncounterLevel60Tank.encounter, {
	tanks: EncounterLevel60Tank.tanks,
	healingModel: EncounterLevel60Tank.healingModel,
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: MaxSurvivalConsumes,
	inFrontOfTarget: true,
});

// The three presets (the user's ruling: threat against survival as weights, Balanced the default), each a
// Night Elf at Max threat and Balanced (Elune's Light), Tauren at Max survival (Endurance's 5% health: +25.5 / +0.1 /
// +17.8 / +42.0 over Night Elf on seeds 1 to 4 at 10,000 iterations, a mean of +21; tank_fairness-20261112-3), on the
// synthetic set (Max threat and Max survival with their own enchants).
export const PresetBuildMaxThreat = PresetUtils.makePresetBuild('Max threat', {
	tooltip: 'Scored with survival weighed at 1/3 (docs/tank_sim.md). Strength enchants on the head, legs, wrist and weapon; Enrage under 70 Rage; R.O.I.D.S. for Spirit of Zanza.',
	gear: GearLevel60Phase1SyntheticMaxThreat,
	talents: TalentsBear,
	rotation: APLMaxThreat,
	encounter: EncounterLevel60TankMaxThreat,
	race: Race.RaceNightElf,
});
export const PresetBuildBalanced = PresetUtils.makePresetBuild('Balanced', {
	tooltip: 'Scored with survival weighed at 1 (docs/tank_sim.md). The default.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsBear,
	rotation: APLBalanced,
	encounter: EncounterLevel60Tank,
	race: Race.RaceNightElf,
});
export const PresetBuildMaxSurvival = PresetUtils.makePresetBuild('Max survival', {
	tooltip: "Scored with survival weighed at 3 (docs/tank_sim.md). Balanced's gear with Greater Agility on the gloves, Feral Instinct 3 and Brutal Impact 1 (Swipe is its filler at 80 Rage), Dirge's chops for the Dumpling, the Greater Stoneshield Potion on cooldown, the Flask of the Titans.",
	gear: GearLevel60Phase1SyntheticMaxSurvival,
	talents: TalentsBearMaxSurvival,
	rotation: APLMaxSurvival,
	encounter: EncounterLevel60TankMaxSurvival,
	race: Race.RaceTauren,
});
