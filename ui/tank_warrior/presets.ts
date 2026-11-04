import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	Alcohol,
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
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { SavedTalents } from '../core/proto/ui.js';
import { TankWarrior_Options as TankWarriorOptions, WarriorShout, WarriorStance } from '../core/proto/warrior.js';
import APLNoReckJSON from './apls/dps_no_reck.apl.json';
import APLReckJSON from './apls/dps_reck.apl.json';
import Level60BalancedAPLJSON from './apls/level60_balanced.apl.json';
import Level60MaxSurvivalAPLJSON from './apls/level60_max_survival.apl.json';
import Level60MaxThreatAPLJSON from './apls/level60_max_threat.apl.json';
import ProtectionAPLJSON from './apls/protection.apl.json';
import Level60Phase1SyntheticGear from './gear_sets/level60_phase1_synthetic.gear.json';
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
// 106 Defense, 110 block value, 239 Stamina, 8% hit (the cap), the tank stat trinket and the attack power on-use.
export const GearLevel60Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 (Synthetic)', Level60Phase1SyntheticGear);

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Synthetic, GearP0BISSwordBoard, GearP0BIS, GearP1BIS],
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
// Balanced and Max survival (whose search kept Balanced's build): Deflection 5; Cruelty, Unbridled Wrath and Enrage 5/5; Protection 31 with Shield Slam, no Last Stand or Vanguard.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60 Protection', SavedTalents.create({ talentsString: '05-05050005-250530020310001351' }));
// Max threat: Improved Heroic Strike 3 in place of Deflection, Blood Craze 1, Anticipation 4, Improved Sunder Armor 3.
export const TalentsLevel60MaxThreat = PresetUtils.makePresetTalents(
	'Level 60 Protection (Max threat)',
	SavedTalents.create({ talentsString: '3-05050015-250430020330001351' }),
);
export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60MaxThreat, TalentsP1Prot],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

export const PresetBuildTanky = PresetUtils.makePresetBuild('Tanky', { gear: GearP0BISSwordBoard, talents: TalentsP1Prot, rotation: APLProtection });

// The three Level 60 tank presets (the user's ruling: threat against survival as weights, Balanced the default).
export const PresetBuildLevel60MaxThreat = PresetUtils.makePresetBuild('Level 60 Max threat', {
	tooltip: 'Scored with survival weighed at 1/3 (docs/tank_sim.md).',
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60MaxThreat,
	rotation: APLLevel60MaxThreat,
	race: Race.RaceUndead,
});
export const PresetBuildLevel60Balanced = PresetUtils.makePresetBuild('Level 60 Balanced', {
	tooltip: 'Scored with survival weighed at 1 (docs/tank_sim.md).',
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60Balanced,
	race: Race.RaceUndead,
});
export const PresetBuildLevel60MaxSurvival = PresetUtils.makePresetBuild('Level 60 Max survival', {
	tooltip: 'Scored with survival weighed at 3 (docs/tank_sim.md).',
	gear: GearLevel60Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60MaxSurvival,
	race: Race.RaceUndead,
});

///////////////////////////////////////////////////////////////////////////
//                                 Options Presets
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = TankWarriorOptions.create({
	queueDelay: 250,
	startingRage: 0,
	shout: WarriorShout.WarriorShoutBattle,
	stance: WarriorStance.WarriorStanceDefensive,
});

export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	alcohol: Alcohol.AlcoholRumseyRumBlackLabel,
	armorElixir: ArmorElixir.ElixirOfSuperiorDefense,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MightyRagePotion,
	dragonBreathChili: true,
	food: Food.FoodSmokedDesertDumpling,
	flask: Flask.FlaskOfTheTitans,
	healthElixir: HealthElixir.ElixirOfFortitude,
	// The stone goes on the main hand: a shield can't hold one, and under Forever Elemental Sharpening
	// Stone crits only the hand it is on (da4f2018a7).
	mainHandImbue: WeaponImbue.ElementalSharpeningStone,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.ROIDS,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	battleShout: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectRegular,
	fireResistanceAura: true,
	fireResistanceTotem: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	leaderOfThePack: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	stoneskinTotem: TristateEffect.TristateEffectRegular,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	fengusFerocity: false,
	moldarsMoxie: false,
	rallyingCryOfTheDragonslayer: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
	warchiefsBlessing: false,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfRecklessness: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	faerieFire: true,
	giftOfArthas: true,
	insectSwarm: true,
	sunderArmor: true,
});

export const OtherDefaults = {
	profession1: Profession.Blacksmithing,
	profession2: Profession.Enchanting,
	race: Race.RaceHuman,
};
