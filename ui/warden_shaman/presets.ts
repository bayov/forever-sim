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
	Encounter,
	Flask,
	Food,
	HealingModel,
	HealthElixir,
	IndividualBuffs,
	IntellectElixir,
	ManaRegenElixir,
	MobType,
	NaturePowerBuff,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	SpellPowerBuff,
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
import { EnhancementShaman_Options as EnhancementShamanOptions, ShamanSyncType } from '../core/proto/shaman.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import DefaultAPLJSON from './apls/default.apl.json';
import Level60BalancedAPLJSON from './apls/level60_balanced.apl.json';
import Level60MaxSurvivalAPLJSON from './apls/level60_max_survival.apl.json';
import Level60MaxThreatAPLJSON from './apls/level60_max_threat.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import Phase1SyntheticTankGear from './gear_sets/phase1_synthetic_tank.gear.json';
import Phase1SyntheticTankArmorTrinketGear from './gear_sets/phase1_synthetic_tank_armor_trinket.gear.json';
import Phase1SyntheticTankArmorTrinketBalancedGear from './gear_sets/phase1_synthetic_tank_armor_trinket_balanced.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);
// The synthetic Phase 1 tank set (docs/tank_shaman.md): Max threat's with the attack power on-use trinket, Balanced's and
// Max survival's with the armor one. Max threat and Balanced wear Agility on the cloak and Greater Agility on the boots, Max
// survival Superior Stamina and Greater Stamina (tank_gear-20261112-1, -3); the wrists are Superior Deflection.
// Head, legs, neck and gloves (tank_fairness-20261112-2, 10,000 iterations, seeds 1 / 2, against Voracity Stamina, Necklace
// Strength and Gloves Superior Strength): every set wears Necklace Agility. Max threat: Voracity Agility, Gloves Natural
// Power, +21.7 / +18.5. Balanced: Voracity Agility, Gloves Greater Agility, +31.8 / +23.5. Max survival: Lesser Arcanum of
// Tenacity (+125 armor), Gloves Greater Agility, +82.5 / +91.2.
export const GearLevel60Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic)', Phase1SyntheticTankGear);
export const GearLevel60SyntheticArmorTrinket = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic, armor trinket)', Phase1SyntheticTankArmorTrinketGear);

export const GearLevel60SyntheticArmorTrinketBalanced = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic, armor trinket, Balanced)', Phase1SyntheticTankArmorTrinketBalancedGear);

export const GearPresets = {
	[Phase.Phase1]: [GearBlank, GearLevel60Synthetic, GearLevel60SyntheticArmorTrinket, GearLevel60SyntheticArmorTrinketBalanced],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

// The default is the Balanced tank build's (docs/tank_shaman.md, Presets), as the tank warrior's.
export const DefaultGear = GearLevel60SyntheticArmorTrinketBalanced;

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

export const DefaultAPL = APLLevel60Balanced;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '35050013-0502333135510102' }));

// The optimized tank builds (docs/tank_shaman.md, Presets), one a weight (round 3): Balanced takes Elemental Warding 3
// and Toughness 4; Max survival Guardian Totems 2 for Improved Lightning Shield 2. Max threat was re-searched after the
// Nov 10 shaman notes (2026-11-12): Elemental Devastation +2, Elemental Warding +2 and Shamanistic Focus for 3 points
// of Toughness and 2 of Improved Reincarnation.
export const TalentsLevel60Tank = PresetUtils.makePresetTalents('Level 60 Tank', SavedTalents.create({ talentsString: '05050003-050032203451012201-050002' }));
export const TalentsLevel60TankMaxSurvival = PresetUtils.makePresetTalents(
	'Level 60 Tank Max survival',
	SavedTalents.create({ talentsString: '05050003-050230113541012201-050002' }),
);
export const TalentsLevel60TankMaxThreat = PresetUtils.makePresetTalents(
	'Level 60 Tank Max threat',
	SavedTalents.create({ talentsString: '05350002-050032313251012201-05' }),
);

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60Tank, TalentsLevel60TankMaxThreat, TalentsLevel60TankMaxSurvival],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultTalents = TalentsLevel60Tank;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = EnhancementShamanOptions.create({
	shamanImbue: WeaponImbue.RockbiterWeapon,
	syncType: ShamanSyncType.Auto,
});

// The tank bench's consumables (tools/tankbench/req/shaman_warden_*.json), searched per preset with the full raid
// treatment (orchestration/standards.md item 11; docs/tank_shaman.md, Round 5). Every tank takes the same armor and
// health elixirs (tank_fairness-20261104-1). Rockbiter Weapon stays on the main hand: a stone or oil in its place loses
// about 1,900. The Greater Stoneshield Potion goes with the APLs' potion action (autocast never drinks a defensive
// potion). Balanced is the default.
export const BalancedConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	alcohol: Alcohol.AlcoholRumseyRumBlackLabel,
	armorElixir: ArmorElixir.ElixirOfThePhalanx,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.GreaterStoneshieldPotion,
	dragonBreathChili: true,
	flask: Flask.FlaskOfTheTitans,
	food: Food.FoodDirgesKickChimaerokChops,
	healthElixir: HealthElixir.ElixirOfGreaterFortitude,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	naturePowerBuff: NaturePowerBuff.ElixirOfNaturePower,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.SpiritOfZanza,
});
// Max threat takes the Flask of Natural Precision and the Ground Scorpok Assay for Spirit of Zanza. Balanced and Max
// survival take the Flask of the Titans since the Nov 10 shaman notes (2026-11-12, 10,000 iterations, seeds 1 / 2:
// Balanced +13.5 / +19.0, Max survival +254.6 / +230.3 over Natural Precision; Supreme Power loses at Max threat).
export const MaxThreatConsumes = Consumes.create({
	...BalancedConsumes,
	flask: Flask.FlaskOfNaturalPrecision,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
});
// Max survival leaves out the Greater Mageblood Elixir (a tie there). The Mageblood Potion is raid-only (tank_fairness-20261112-1).
export const MaxSurvivalConsumes = Consumes.create({ ...BalancedConsumes, manaRegenElixir: ManaRegenElixir.ManaRegenElixirUnknown });

export const DefaultConsumes = BalancedConsumes;

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
	// Balanced's race (the default build).
	race: Race.RaceNightElf,
};

// The tank bench's boss (docs/tank_sim.md, The bench; tools/tankbench/req/shaman_warden_*.json): the shared Level 60
// boss (level 63, 805 attack power, a 2.0 sec swing) on the Warden, with the stand-in healer at 600 HPS every 2 sec
// (1 to 3), a 3 minute fight, in a zone where the Natural flasks work. The presets were searched on the Anniversary
// encounters (tools/encounters), which the UI doesn't load; this is the bench's own default fight, as the tank
// warrior's and the paladin's.
const benchEncounter = (name: string, consumes: Consumes) =>
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
export const EncounterLevel60Tank = benchEncounter('Generic boss, tanking', BalancedConsumes);
export const EncounterLevel60TankMaxThreat = benchEncounter('Generic boss, tanking (Max threat)', MaxThreatConsumes);
export const EncounterLevel60TankMaxSurvival = benchEncounter('Generic boss, tanking (Max survival)', MaxSurvivalConsumes);

// The three Level 60 tank presets (threat against survival as weights, Balanced the default; docs/tank_sim.md).
export const PresetBuildLevel60MaxThreat = PresetUtils.makePresetBuild('Level 60 Max threat', {
	tooltip:
		"Scored with survival weighed at 1/3 (docs/tank_sim.md). Night Elf (Elune's Light), Mana Spring Totem, Searing Totem and Fire Nova all fight, the attack power trinket on cooldown. Round 5 consumables: Natural Precision, Ground Scorpok Assay, Greater Stoneshield Potion on cooldown.",
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60TankMaxThreat,
	rotation: APLLevel60MaxThreat,
	encounter: EncounterLevel60TankMaxThreat,
	race: Race.RaceNightElf,
	professions: [Profession.Engineering, Profession.Blacksmithing],
});
export const PresetBuildLevel60Balanced = PresetUtils.makePresetBuild('Level 60 Balanced', {
	tooltip:
		"Scored with survival weighed at 1 (docs/tank_sim.md). Night Elf (Elune's Light), Healing Stream Totem, Searing Totem and Fire Nova in the last 60 sec, the armor trinket held for the last 25 sec, Superior Deflection on the wrists. Consumables: Flask of the Titans, Spirit of Zanza, Greater Stoneshield Potion on cooldown.",
	gear: GearLevel60SyntheticArmorTrinketBalanced,
	talents: TalentsLevel60Tank,
	rotation: APLLevel60Balanced,
	encounter: EncounterLevel60Tank,
	race: Race.RaceNightElf,
	professions: [Profession.Engineering, Profession.Blacksmithing],
});
export const PresetBuildLevel60MaxSurvival = PresetUtils.makePresetBuild('Level 60 Max survival', {
	tooltip:
		'Scored with survival weighed at 3 (docs/tank_sim.md). Undead (Touch of the Grave), Healing Stream Totem, Searing Totem and Fire Nova in the last 60 sec, the armor trinket held for the last 25 sec, Superior Deflection on the wrists. Consumables: Flask of the Titans, Spirit of Zanza, Greater Stoneshield Potion on cooldown, no Mageblood.',
	gear: GearLevel60SyntheticArmorTrinket,
	talents: TalentsLevel60TankMaxSurvival,
	rotation: APLLevel60MaxSurvival,
	encounter: EncounterLevel60TankMaxSurvival,
	race: Race.RaceUndead,
	professions: [Profession.Engineering, Profession.Blacksmithing],
});
