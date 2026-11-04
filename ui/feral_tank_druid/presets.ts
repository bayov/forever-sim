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
	Stat,
	StrengthBuff,
	SunderRamp,
	TotemWeaponBuff,
	TristateEffect,
	UnitReference,
	UnitReference_Type,
	WeaponImbue,
} from '../core/proto/common.js';
import { FeralTankDruid_Options as DruidOptions, FeralTankDruid_Rotation as DruidRotation } from '../core/proto/druid.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import BalancedAPLJSON from './apls/balanced.apl.json';
import MaxSurvivalAPLJSON from './apls/max_survival.apl.json';
import MaxThreatAPLJSON from './apls/max_threat.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';

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
// archetype, 729 weighted points with the feral two-hander), 256 Stamina, 164 Agility, 122 Strength,
// 105 Defense and 3% hit (8% with the flask). The trinkets are the tank stat trinket (23 Defense, 1%
// dodge) and the cat's on-use Strength trinket, which the tank objective scores far over the defensive
// armor one (its opening threat). Idol of the Ursine Twins (Lacerate resets Primal Bite 10% of the
// time) is ahead of Idol of Brutality and Enraged Idol. Enchants: Agility on the head and legs (Lesser
// Arcanums of Voracity), the neck, back, boots and weapon, Greater Stats, Superior Stamina on the wrist
// and Superior Strength on the gloves, each ahead of or tied with the Stamina one at Balanced.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Bear', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS for a bear under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

export const GearPresets = [GearLevel60Phase1Synthetic];

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The three presets' rotations (tools/rotopt -template druid_tank, each searched at its weight, round 2 on the
// flat white-hit Rage). All three put Primal Bite on cooldown, keep Lacerate at 5 stacks, keep Demoralizing Roar
// up just below that line (refreshed under 3 sec), use Barkskin on its timer from 15 sec in, in the filler's
// place, shift into Dire Bear Form 1.5 sec before the pull and Enrage in the fight. Balanced: Lacerate as the
// filler, Maul at 20 Rage, Enrage under 40 Rage and Frenzied Regeneration at 100 Rage. Max threat: Balanced
// without Frenzied Regeneration and with Enrage under 60 Rage. Max survival: Balanced with Lacerate kept up but
// not the filler and Maul at 10 Rage, so its Rage goes to Maul. No preset enrages before the pull any more
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

export const TalentPresets = [TalentsBear];

export const DefaultTalents = TalentsBear;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// The bear starts the pull with no Rage; its rotation shifts before the pull for Furor's.
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

// Forever launch Phase 1 consumables (nothing from a raid), each the best of its slot for the tank
// objective at Balanced: the Flask of Natural Accuracy (5% hit and 60 Stamina in Hyjal and the Barrow
// Deeps, standards 11; 105 ahead of the Flask of the Titans), Elixir of the Phalanx (400 health, 500
// armor), Elixir of Greater Fortitude, Elixir of the Mongoose (taking it out costs 124), Juju Power and
// Might, Smoked Desert Dumpling (ahead of Dirge's Stamina chops and Grilled Squid), Dragonbreath Chili,
// a Dense Weightstone on the mace, and the Mighty Rage Potion at the pull (Greater Stoneshield, a
// defensive potion, loses it all).
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	armorElixir: ArmorElixir.ElixirOfThePhalanx,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MightyRagePotion,
	dragonBreathChili: true,
	flask: Flask.FlaskOfNaturalAccuracy,
	food: Food.FoodSmokedDesertDumpling,
	healthElixir: HealthElixir.ElixirOfGreaterFortitude,
	mainHandImbue: WeaponImbue.DenseWeightstone,
	strengthBuff: StrengthBuff.JujuPower,
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

// The three presets (the user's ruling: threat against survival as weights, Balanced the default), each a
// Night Elf (the best race at Max threat and Balanced: Elune's Light; tied with the Tauren at Max survival), on
// the synthetic set.
export const PresetBuildMaxThreat = PresetUtils.makePresetBuild('Max threat', {
	tooltip: 'Scored with survival weighed at 1/3 (docs/tank_sim.md).',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsBear,
	rotation: APLMaxThreat,
	encounter: EncounterLevel60Tank,
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
	tooltip: 'Scored with survival weighed at 3 (docs/tank_sim.md).',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsBear,
	rotation: APLMaxSurvival,
	encounter: EncounterLevel60Tank,
	race: Race.RaceNightElf,
});
