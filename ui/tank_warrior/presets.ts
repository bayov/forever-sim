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
import ProtectionAPLJSON from './apls/protection.apl.json';
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

export const GearPresets = {
	[Phase.Phase1]: [GearP0BISSwordBoard, GearP0BIS, GearP1BIS],
};

export const DefaultGear = GearP0BISSwordBoard;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const AplReck = PresetUtils.makePresetAPLRotation('DPS (With Reck)', APLReckJSON);
export const APLNoReck = PresetUtils.makePresetAPLRotation('DPS (No Reck)', APLNoReckJSON);
// A plain Protection rotation in Defensive Stance: Shield Block whenever it is down, Shield Slam on
// cooldown, Revenge, Sunder Armor to 5 stacks, Heroic Strike from 50 Rage and Sunder Armor as the
// filler from 40, Bloodrage under 30 Rage.
export const APLProtection = PresetUtils.makePresetAPLRotation('Protection', ProtectionAPLJSON);

export const APLPresets = {
	[Phase.Phase1]: [APLProtection, APLNoReck, AplReck],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/classic/talent-calc and copy the numbers in the url.

export const TalentsP1Prot = PresetUtils.makePresetTalents('Protection', SavedTalents.create({ talentsString: '35--251533120331221351' }));

export const TalentPresets = {
	[Phase.Phase1]: [TalentsP1Prot],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

export const PresetBuildTanky = PresetUtils.makePresetBuild('Tanky', { gear: DefaultGear, talents: TalentsP1Prot, rotation: DefaultAPL });

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
