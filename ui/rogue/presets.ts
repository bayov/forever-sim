import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	Flask,
	Food,
	IndividualBuffs,
	MobType,
	Profession,
	Race,
	RaidBuffs,
	SpellPowerBuff,
	StrengthBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
	SapperExplosive,
	Stat,
} from '../core/proto/common.js';
import { RogueOptions } from '../core/proto/rogue.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import BackstabAPL from './apls/combat_backstab.apl.json';
import BackstabRuptureAPL from './apls/combat_backstab_rupture.apl.json';
import BackstabOnCooldownAPL from './apls/combat_backstab_on_cooldown.apl.json';
import BackstabSndWindowAPL from './apls/combat_backstab_snd_window.apl.json';
import BackstabSweatyAPL from './apls/combat_backstab_sweaty.apl.json';
import SinisterStrikeAPL from './apls/combat_sinister_strike.apl.json';
import SinisterStrikeOnCooldownAPL from './apls/combat_sinister_strike_on_cooldown.apl.json';
import SinisterStrikeHoldForARAPL from './apls/combat_sinister_strike_hold_for_ar.apl.json';
import SinisterStrikeSndOpenerAPL from './apls/combat_sinister_strike_snd_opener.apl.json';
import SinisterStrikeSndWindowAPL from './apls/combat_sinister_strike_snd_window.apl.json';
import SinisterStrikeSweatyAPL from './apls/combat_sinister_strike_sweaty.apl.json';
import SinisterStrikeIEAAPL from './apls/combat_sinister_strike_iea.apl.json';
import MutilateAPL from './apls/forever_mutilate.apl.json';
import Level20BackstabAPL from './apls/level20_backstab.apl.json';
import Level20SinisterStrikeAPL from './apls/level20_sinister_strike.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import BackstabGearPreBiS from './gear_sets/combat_backstab_prebis.gear.json';
import SinisterStrikeGearPreBiS from './gear_sets/combat_sinister_strike_prebis.gear.json';
import BackstabGearP1BiS from './gear_sets/combat_backstab_p1_bis.gear.json';
import BackstabGearP2BiS from './gear_sets/combat_backstab_p2_bis.gear.json';
import SinisterStrikeGearP1BiS from './gear_sets/combat_sinister_strike_p1_bis.gear.json';
import SinisterStrikeGearP2BiS from './gear_sets/combat_sinister_strike_p2_bis.gear.json';
import Level20BackstabGear from './gear_sets/level20_backstab.gear.json';
import Level20SinisterStrikeGear from './gear_sets/level20_sinister_strike.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);
export const GearBackstabPreBiS = PresetUtils.makePresetGear('Backstab Pre-BiS', BackstabGearPreBiS);
export const GearSinisterStrikePreBiS = PresetUtils.makePresetGear('Sinister Strike Pre-BiS', SinisterStrikeGearPreBiS);
export const GearBackstabP1BiS = PresetUtils.makePresetGear('Backstab P1 BiS', BackstabGearP1BiS);
export const GearBackstabP2BiS = PresetUtils.makePresetGear('Backstab P2 BiS', BackstabGearP2BiS);
export const GearSinisterStrikeP1BiS = PresetUtils.makePresetGear('Sinister Strike P1 BiS', SinisterStrikeGearP1BiS);
export const GearSinisterStrikeP2BiS = PresetUtils.makePresetGear('Sinister Strike P2 BiS', SinisterStrikeGearP2BiS);
// Level 20 sets found by tools/rotopt -level 20 -gear-search over everything an Undead
// rogue can equip at 20, on the wowhead Forever item data: dungeon blues (Serpent's
// Shoulders and Gloves of the Fang from Wailing Caverns, Cultist's Armguards from
// Blackfathom Deeps, Blackened Defias Belt, Feet of the Lynx), a Scouting Tunic of Power,
// Band of the Fist from the Blackfathom quest and the Warsong Gulch neck and ring. Both
// sets assume Engineering and Leatherworking for Gnomish Goggles, Parachute Cloak and
// Brawler's Leather Pants. Only the main hand differs: Cruel Barb off Edwin VanCleef for
// Sinister Strike, an Assassin's Blade for Backstab, and both hold Butcher's Cleaver in
// the off hand, which Forever lets a rogue wield (its Hack and Slash reads "Axe/Sword").
//
// The rogue used to wear Field Researcher's Loop, which is gone: it is marked rogue gear
// but the only quest that hands it over is the mage's Greater Friend of the Library, so
// nobody can have it. Losing 7 Strength, 7 Agility and 7 Stamina costs more than the axe
// wins back, and both sets land 0.6 below where they were.
//
// Lil Timmy's Peashooter is in the ranged slot for 0.7 DPS. It is a 1 in several thousand
// drop off the Defias in Westfall, so take the numbers with that in mind.
//
// A search with Blessing of Kings on keeps both sets. The trinket slots stay empty
// because nothing a rogue can wear there at 20 does anything for damage. The search did
// put Rune of Duty in for 0.1, but it is 4 Stamina and resistances, so that was noise.
export const GearLevel20Backstab = PresetUtils.makePresetGear('Level 20 Backstab', Level20BackstabGear);
export const GearLevel20SinisterStrike = PresetUtils.makePresetGear('Level 20 Sinister Strike', Level20SinisterStrikeGear);

export const GearPresets = {
	[Phase.Phase1]: [GearBackstabPreBiS, GearSinisterStrikePreBiS, GearBackstabP1BiS, GearSinisterStrikeP1BiS, GearLevel20Backstab, GearLevel20SinisterStrike],
	[Phase.Phase2]: [GearBackstabPreBiS, GearSinisterStrikePreBiS, GearBackstabP2BiS, GearSinisterStrikeP2BiS, GearLevel20Backstab, GearLevel20SinisterStrike],
};

export const DefaultGear = GearSinisterStrikePreBiS;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets[]
///////////////////////////////////////////////////////////////////////////

export const ROTATION_PRESET_BACKSTAB = PresetUtils.makePresetAPLRotation('Backstab', BackstabAPL, {});
export const ROTATION_PRESET_SINISTER_STRIKE = PresetUtils.makePresetAPLRotation('Sinister Strike', SinisterStrikeAPL, {});
// Variants of the Backstab and Sinister Strike rotations, generated by
// tools/rotopt/rogue_presets.sh. The plain ones are the optimizer's best, these are the
// alternatives to compare against.
export const ROTATION_PRESET_BACKSTAB_RUPTURE = PresetUtils.makePresetAPLRotation('Backstab (Rupture at 5 points)', BackstabRuptureAPL, {});
export const ROTATION_PRESET_BACKSTAB_ON_COOLDOWN = PresetUtils.makePresetAPLRotation('Backstab (cooldowns on cooldown)', BackstabOnCooldownAPL, {});
export const ROTATION_PRESET_BACKSTAB_SND_WINDOW = PresetUtils.makePresetAPLRotation(
	'Backstab (cooldowns in a 20s Slice and Dice window)',
	BackstabSndWindowAPL,
	{},
);
export const ROTATION_PRESET_SINISTER_STRIKE_ON_COOLDOWN = PresetUtils.makePresetAPLRotation('SS (cooldowns on cooldown)', SinisterStrikeOnCooldownAPL, {});
export const ROTATION_PRESET_SINISTER_STRIKE_HOLD_FOR_AR = PresetUtils.makePresetAPLRotation(
	'SS (everything with Adrenaline Rush)',
	SinisterStrikeHoldForARAPL,
	{},
);
export const ROTATION_PRESET_SINISTER_STRIKE_SND_OPENER = PresetUtils.makePresetAPLRotation(
	'SS (5 point Slice and Dice before cooldowns)',
	SinisterStrikeSndOpenerAPL,
	{},
);
export const ROTATION_PRESET_SINISTER_STRIKE_SND_WINDOW = PresetUtils.makePresetAPLRotation(
	'SS (cooldowns in a 20s Slice and Dice window)',
	SinisterStrikeSndWindowAPL,
	{},
);
export const ROTATION_PRESET_BACKSTAB_SWEATY = PresetUtils.makePresetAPLRotation('Backstab (Sweaty)', BackstabSweatyAPL, {});
export const ROTATION_PRESET_SINISTER_STRIKE_SWEATY = PresetUtils.makePresetAPLRotation('Sinister Strike (Sweaty)', SinisterStrikeSweatyAPL, {});
export const ROTATION_PRESET_SINISTER_STRIKE_IEA = PresetUtils.makePresetAPLRotation('Improved Expose Armor (SS)', SinisterStrikeIEAAPL, {});
export const ROTATION_PRESET_MUTILATE = PresetUtils.makePresetAPLRotation('Mutilate', MutilateAPL, {});
// Level 20 rotations from the same templates: Slice and Dice, Rupture at 3 points (rank 1
// Rupture beats rank 3 Eviscerate at this level), builder whenever the energy is there,
// Thistle Tea when low. The lines name the level 60 spell ranks and the sim resolves them to the
// rank the level knows, so the icons show a higher rank than what is cast.
export const ROTATION_PRESET_LEVEL20_BACKSTAB = PresetUtils.makePresetAPLRotation('Level 20 Backstab', Level20BackstabAPL, {});
export const ROTATION_PRESET_LEVEL20_SINISTER_STRIKE = PresetUtils.makePresetAPLRotation('Level 20 Sinister Strike', Level20SinisterStrikeAPL, {});

export const APLPresets = {
	[Phase.Phase1]: [
		ROTATION_PRESET_BACKSTAB,
		ROTATION_PRESET_BACKSTAB_RUPTURE,
		ROTATION_PRESET_BACKSTAB_ON_COOLDOWN,
		ROTATION_PRESET_BACKSTAB_SND_WINDOW,
		ROTATION_PRESET_SINISTER_STRIKE,
		ROTATION_PRESET_SINISTER_STRIKE_ON_COOLDOWN,
		ROTATION_PRESET_SINISTER_STRIKE_HOLD_FOR_AR,
		ROTATION_PRESET_SINISTER_STRIKE_SND_OPENER,
		ROTATION_PRESET_SINISTER_STRIKE_SND_WINDOW,
		ROTATION_PRESET_BACKSTAB_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_IEA,
		ROTATION_PRESET_MUTILATE,
		ROTATION_PRESET_LEVEL20_BACKSTAB,
		ROTATION_PRESET_LEVEL20_SINISTER_STRIKE,
	],
	[Phase.Phase2]: [
		ROTATION_PRESET_BACKSTAB,
		ROTATION_PRESET_BACKSTAB_RUPTURE,
		ROTATION_PRESET_BACKSTAB_ON_COOLDOWN,
		ROTATION_PRESET_BACKSTAB_SND_WINDOW,
		ROTATION_PRESET_SINISTER_STRIKE,
		ROTATION_PRESET_SINISTER_STRIKE_ON_COOLDOWN,
		ROTATION_PRESET_SINISTER_STRIKE_HOLD_FOR_AR,
		ROTATION_PRESET_SINISTER_STRIKE_SND_OPENER,
		ROTATION_PRESET_SINISTER_STRIKE_SND_WINDOW,
		ROTATION_PRESET_BACKSTAB_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_IEA,
		ROTATION_PRESET_MUTILATE,
		ROTATION_PRESET_LEVEL20_BACKSTAB,
		ROTATION_PRESET_LEVEL20_SINISTER_STRIKE,
	],
};

//Need to add main hand equip logic or talent/rotation logic to map to Auto APL
export const DefaultAPLs: Record<number, PresetUtils.PresetRotation> = {
	[0]: ROTATION_PRESET_SINISTER_STRIKE,
	[1]: ROTATION_PRESET_BACKSTAB,
};

export const DefaultAPLBackstab = ROTATION_PRESET_BACKSTAB;
export const DefaultAPLSinisterStrike = ROTATION_PRESET_SINISTER_STRIKE;
export const DefaultAPLIEA = ROTATION_PRESET_SINISTER_STRIKE_IEA;
export const DefaultAPLMutilate = ROTATION_PRESET_MUTILATE;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/forever/talent-calc and copy the numbers in the url.

// Preset name must be unique. Ex: 'Backstab DPS' cannot be used as a name more than once

export const CombatBackstabTalents = PresetUtils.makePresetTalents('Backstab', SavedTalents.create({ talentsString: '005323102-30230320201515231-002' }));
export const CombatSinisterStrikeTalents = PresetUtils.makePresetTalents(
	'Sinister Strike',
	SavedTalents.create({ talentsString: '00532310301-32003311201515231' }),
);
export const CombatSinisterStrikeIEATalents = PresetUtils.makePresetTalents(
	'Improved Expose Armor (SS)',
	SavedTalents.create({ talentsString: '005323123-32003311201515131' }),
);
export const AssassinationMutilateTalents = PresetUtils.makePresetTalents('Mutilate', SavedTalents.create({ talentsString: '00530310551021051-302303202004' }));
// The 11 points a level 20 has. Sinister Strike: Malice 5, Imp SnD 2, Murder 2, Imp SS 2,
// and without Murder (anything but Humanoids and Giants) the two points go to Ruthlessness
// and Relentless Strikes instead. Backstab: Malice 1, Imp Evis 3, Lightning Reflexes 2,
// Puncturing Wounds 3, Opportunity 2 (the best Backstab build has no Murder, five Combat
// points are needed to reach Puncturing Wounds and nothing is left for row 2 of Assassination).
export const Level20SinisterStrikeTalents = PresetUtils.makePresetTalents('Level 20 Sinister Strike', SavedTalents.create({ talentsString: '005022-02' }));
export const Level20SinisterStrikeNoMurderTalents = PresetUtils.makePresetTalents(
	'Level 20 Sinister Strike (no Murder)',
	SavedTalents.create({ talentsString: '0053021' }),
);
export const Level20BackstabTalents = PresetUtils.makePresetTalents('Level 20 Backstab', SavedTalents.create({ talentsString: '001-3023-002' }));

export const TalentPresets = {
	[Phase.Phase1]: [
		CombatBackstabTalents,
		CombatSinisterStrikeTalents,
		CombatSinisterStrikeIEATalents,
		AssassinationMutilateTalents,
		Level20BackstabTalents,
		Level20SinisterStrikeTalents,
		Level20SinisterStrikeNoMurderTalents,
	],
	[Phase.Phase2]: [
		CombatBackstabTalents,
		CombatSinisterStrikeTalents,
		CombatSinisterStrikeIEATalents,
		AssassinationMutilateTalents,
		Level20BackstabTalents,
		Level20SinisterStrikeTalents,
		Level20SinisterStrikeNoMurderTalents,
	],
};

export const DefaultTalentsAssassin = TalentPresets[Phase.Phase2][0];
export const DefaultTalentsCombat = TalentPresets[Phase.Phase2][0];
export const DefaultTalentsSubtlety = TalentPresets[Phase.Phase2][0];

export const DefaultTalentsBackstab = TalentPresets[Phase.Phase2][0];
export const DefaultTalentsSinisterStrike = TalentPresets[Phase.Phase2][1];
export const DefaultTalentsIEA = TalentPresets[Phase.Phase2][2];
export const DefaultTalentsMutilate = TalentPresets[Phase.Phase2][3];

export const DefaultTalents = DefaultTalentsSinisterStrike;

///////////////////////////////////////////////////////////////////////////
//                                Build Presets
///////////////////////////////////////////////////////////////////////////
export const PresetBuildBackstab = PresetUtils.makePresetBuild('Backstab', {
	gear: GearBackstabP2BiS,
	talents: DefaultTalentsBackstab,
	rotation: DefaultAPLBackstab,
});
export const PresetBuildSinisterStrike = PresetUtils.makePresetBuild('Sinister Strike', {
	gear: GearSinisterStrikeP2BiS,
	talents: DefaultTalentsSinisterStrike,
	rotation: DefaultAPLSinisterStrike,
});
export const PresetBuildIEA = PresetUtils.makePresetBuild('IEA', {
	gear: GearSinisterStrikeP2BiS,
	talents: DefaultTalentsIEA,
	rotation: DefaultAPLIEA,
});
export const PresetBuildMutilate = PresetUtils.makePresetBuild('Mutilate', {
	gear: GearBackstabP2BiS,
	talents: DefaultTalentsMutilate,
	rotation: DefaultAPLMutilate,
});

// The level 20 comparison setup: Mutanus the Devourer's level and armor (a level 22 dungeon
// boss, VanCleef is 888), 60 second fights, no raid buffs or debuffs since the buff table
// holds level 60 values, Instant Poison on both weapons and Thistle Tea.
//
// Blessing of Kings is the exception. It is 10% of every attribute rather than a number
// off the level 60 table, so it is right at any level, and a group at 20 almost always
// has a paladin in it. The gear search ran with it.
const level20Encounter = (name: string, targetName: string, mobType: MobType) =>
	PresetUtils.makePresetEncounter(
		name,
		Encounter.create({
			duration: 60,
			durationVariation: 5,
			executeProportion20: 0.2,
			executeProportion25: 0.25,
			executeProportion35: 0.35,
			targets: [
				{
					id: 3654,
					name: targetName,
					level: 22,
					mobType,
					stats: new Stats().withStat(Stat.StatArmor, 922).withStat(Stat.StatHealth, 20000).asArray(),
					minBaseDamage: 40,
					damageSpread: 0.3333,
					swingSpeed: 2,
					parryHaste: true,
				},
			],
		}),
		{
			raidBuffs: RaidBuffs.create({}),
			debuffs: Debuffs.create({}),
			buffs: IndividualBuffs.create({ blessingOfKings: true }),
			consumes: Consumes.create({
				defaultConjured: Conjured.ConjuredRogueThistleTea,
				mainHandImbue: WeaponImbue.InstantPoison,
				offHandImbue: WeaponImbue.InstantPoison,
			}),
		},
	);
export const EncounterLevel20 = level20Encounter('Level 20', 'Mutanus the Devourer', MobType.MobTypeHumanoid);
// The same boss as an Undead, for the builds that skip Murder (it only works on Humanoids
// and Giants).
export const EncounterLevel20Undead = level20Encounter('Level 20 (Undead)', 'Level 22 undead boss', MobType.MobTypeUndead);
// Undead, the rogue we play, and the best of the three Horde races now that Touch of the
// Grave has no internal cooldown. At 30 / 60 / 300 sec on the shipped sets with Blessing
// of Kings, Sinister Strike does 77.7 / 70.5 / 65.0 as Undead, 77.5 / 70.1 / 64.4 as a
// Troll and 76.3 / 69.1 / 63.7 as an Orc. Backstab does 77.2 / 69.6 / 63.6 as Undead,
// 76.5 / 68.8 / 62.9 as a Troll and 75.5 / 67.8 / 62.3 as an Orc.
//
// The Orc gets Axe Specialization on Butcher's Cleaver in the off hand already. Putting
// Razor's Edge in the main hand as well costs it 2.7 at 60 sec, because Cruel Barb hits
// much harder (30-57 against 25-48). Nothing else moves with the race here. The Undead
// and the Troll have no weapon specialization, and the rotations carry every race's
// cooldown line for the others to ignore.
export const PresetBuildLevel20Backstab = PresetUtils.makePresetBuild('Level 20 Backstab', {
	gear: GearLevel20Backstab,
	talents: Level20BackstabTalents,
	rotation: ROTATION_PRESET_LEVEL20_BACKSTAB,
	encounter: EncounterLevel20,
	race: Race.RaceUndead,
	level: 20,
});
export const PresetBuildLevel20SinisterStrike = PresetUtils.makePresetBuild('Level 20 Sinister Strike', {
	gear: GearLevel20SinisterStrike,
	talents: Level20SinisterStrikeTalents,
	rotation: ROTATION_PRESET_LEVEL20_SINISTER_STRIKE,
	encounter: EncounterLevel20,
	race: Race.RaceUndead,
	level: 20,
});

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = RogueOptions.create({});

///////////////////////////////////////////////////////////////////////////
//                         Consumes/Buffs/Debuffs
///////////////////////////////////////////////////////////////////////////

export const P1Consumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredRogueThistleTea,
	dragonBreathChili: true,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodGrilledSquid,
	mainHandImbue: WeaponImbue.InstantPoison,
	offHandImbue: WeaponImbue.DeadlyPoison,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
});

export const DefaultConsumes = {
	[Phase.Phase1]: P1Consumes,
};

export const P1RaidBuffs = RaidBuffs.create({
	battleShout: TristateEffect.TristateEffectImproved,
	fireResistanceAura: true,
	fireResistanceTotem: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	leaderOfThePack: true,
	trueshotAura: true,
});

export const DefaultRaidBuffs = {
	[Phase.Phase1]: P1RaidBuffs,
};

export const P1IndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	fengusFerocity: false,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
	warchiefsBlessing: false,
});

export const DefaultIndividualBuffs = {
	[Phase.Phase1]: P1IndividualBuffs,
};

export const P1DefaultDebuffs = Debuffs.create({
	curseOfRecklessness: true,
	faerieFire: true,
	sunderArmor: true,
});

export const DefaultDebuffs = {
	[Phase.Phase1]: P1DefaultDebuffs,
};

export const P1OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
};

export const OtherDefaults = {
	[Phase.Phase1]: P1OtherDefaults,
};
