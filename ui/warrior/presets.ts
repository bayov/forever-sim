import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Consumes,
	Debuffs,
	Encounter,
	Explosive,
	Flask,
	Food,
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
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Warrior_Options as WarriorOptions, WarriorShout, WarriorStance } from '../core/proto/warrior.js';
import { Stats } from '../core/proto_utils/stats.js';
import APLNoReckJSON from './apls/dps_no_reck.apl.json';
import APLReckJSON from './apls/dps_reck.apl.json';
import Level60ArmsAPLJSON from './apls/level60_arms.apl.json';
import Level60ArmsDragonkinAPLJSON from './apls/level60_arms_dragonkin.apl.json';
import Level60FuryAPLJSON from './apls/level60_fury.apl.json';
import Level60Fury2HAPLJSON from './apls/level60_fury_2h.apl.json';
import Level60Fury2HShoutAPLJSON from './apls/level60_fury_2h_battle_shout.apl.json';
import Level60FuryShoutAPLJSON from './apls/level60_fury_battle_shout.apl.json';
import Level60ArmsShoutAPLJSON from './apls/level60_arms_battle_shout.apl.json';
import Level60Phase1SyntheticArmsGearJSON from './gear_sets/level60_phase1_synthetic_arms.gear.json';
import Level60Phase1SyntheticFuryGearJSON from './gear_sets/level60_phase1_synthetic_fury.gear.json';
import P0BISGear from './gear_sets/p0.bis.gear.json';
import Phase1Gear from './gear_sets/phase_1.gear.json';
import Phase2Gear from './gear_sets/phase_2.gear.json';
import Phase3Gear from './gear_sets/phase_3.gear.json';
import Phase4Gear from './gear_sets/phase_4.gear.json';
import Phase5Gear from './gear_sets/phase_5.gear.json';
import Phase6Gear from './gear_sets/phase_6.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearP0BIS = PresetUtils.makePresetGear('Pre-BiS', P0BISGear);
export const GearPhase1 = PresetUtils.makePresetGear('P1 BiS', Phase1Gear);
export const GearPhase2 = PresetUtils.makePresetGear('P2 BiS', Phase2Gear);
export const GearPhase3 = PresetUtils.makePresetGear('P3 BiS', Phase3Gear);
export const GearPhase4 = PresetUtils.makePresetGear('P4 BiS', Phase4Gear);
export const GearPhase5 = PresetUtils.makePresetGear('P5 BiS', Phase5Gear);
export const GearPhase6 = PresetUtils.makePresetGear('P6 BiS', Phase6Gear);

// A guess at Phase 1 BiS under Forever, made of the synthetic items in
// tools/database/forever_synthetic_items_warrior.go (ids from 991000, not in the game).
//
// Forever's level 60 gear in the client is not tuned yet, and Molten Core and Onyxia have no
// Forever data. The synthetic items take the place of the Era Phase 1 DPS warrior BiS
// (phase_1.gear.json) at the item levels every class shares (66 for the big armor slots, 71 for the
// neck, back, wrist, waist, rings, one-handers, bow and trinkets, 76 for a two-hander), priced on
// Forever's physical melee budget (standards 1b: L = 1.125 ilvl + 2, armor slot mod x L x 0.950 /
// 0.959 / 0.911, three stat lines a piece) in the fair stat mix (SoD level 60 BiS warrior shares,
// Stamina at about 19.7%): Strength first, then Stamina, Agility, crit and attack power, 9% hit (the
// special attack cap), no Spirit. Totals: armor 650.5, bow 19.5, swords 64.0, two-hander 79.0; the
// Fury set 733.5, the Arms / Fury 2H set 748.5. Only the head and legs carry Battlegear of Glory,
// Forever's Arms and Fury tier set, so only its 2-piece is live. The 4-piece (+36 attack
// power against Humanoids) would add 10.8 on Fury, 8.8 on two-hand Fury and 8.4 on Arms against a
// Humanoid boss (Undead, 180 sec). The 5-piece adds nothing, since Recklessness is used once a fight. Arms and Fury share the
// armor, rings, bow and trinkets, the file has the details.
//
// Enchants: Lesser Arcanum of Voracity (8 Strength) on the head and legs, Forever's +5 Strength on
// the neck (+3.7 against +2.9 for Agility on Fury), Forever's Cloak - Agility (5), Greater Stats,
// Superior Strength on the bracers, Forever's Gloves - Superior Strength (15), Greater Agility on the
// boots (a Forceful Rugged Armor Kit there is within 0.3) and Crusader. Nothing for the shoulders
// outside raids.
//
// The trinkets are synthetic too (Forever removed the old dungeon trinkets and the sim does not
// model its launch trinkets yet): a stat trinket with 30 Strength and 16 crit (45.9 weighted), and an
// on-use trinket with +335 attack power for 20 sec every 2 min (37 weighted on average, Earthstrike's
// size; sim/warrior/forever_synthetic_trinkets.go).
//
// On 2026-10-13 the Level 60 Fury build (Human, 180 sec) does 1031.1 on the Fury set, and the Era
// Phase 1 set (phase_1.gear.json) 984.2 with the same talents and rotation.
export const GearLevel60Phase1SyntheticFury = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic (Fury)', Level60Phase1SyntheticFuryGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever for a dual wield warrior, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
// The same armor with a two-hander in place of Bonereaver's Edge, for Arms and two-hand Fury. It is a
// 3.7 sec axe, the slowest Era Phase 1 BiS two-hand speed (standards 22).
export const GearLevel60Phase1SyntheticArms = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic (Two-Hand)', Level60Phase1SyntheticArmsGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever for a two-hand warrior, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1SyntheticFury, GearLevel60Phase1SyntheticArms, GearPhase1, GearP0BIS],
	[Phase.Phase2]: [GearPhase2],
	[Phase.Phase3]: [GearPhase3],
	[Phase.Phase4]: [GearPhase4],
	[Phase.Phase5]: [GearPhase5],
	[Phase.Phase6]: [GearPhase6],
};

export const DefaultGear = GearLevel60Phase1SyntheticFury;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const AplReck = PresetUtils.makePresetAPLRotation('DPS (With Reck)', APLReckJSON);
export const APLNoReck = PresetUtils.makePresetAPLRotation('DPS (No Reck)', APLNoReckJSON);

// The tools/rotopt search (template warrior_fury, see warrior_presets.sh) on the Level 60 Fury
// talents and set, 180 sec, with Forever's normalized Rage (standards 21: a white hit gives 3.46 Rage
// a second of weapon speed at level 30, ×1.5 at 60 with s = 0.5): Whirlwind with 25 Rage, then
// Bloodthirst, Execute after them in the execute phase, Heroic Strike queued from 55 Rage (40 before
// the off hand's Rage was halved, 50 before pass 23, 2026-10-16), and Hamstring as a filler from 10 Rage (its hits proc
// Windfury Totem's extra attacks and Flurry).
// Death Wish goes on cooldown and Recklessness at the pull with the other cooldowns: holding them
// for the execute phase, where Execute spends all the Rage, lost 14 at 180 sec with flat Rage. Bloodrage under 40
// Rage, no Berserker Rage. The Mighty Rage Potion goes at the pull and the second is held for the
// last 20 sec when that costs no use, and the on-use trinket goes on cooldown (+20.7 at 120 sec,
// where the hold alone skipped the pull potion; the rotation auditor's pass 14). The Goblin Land Mine waits for the moment after
// a main hand swing, since a swing that comes due during its 1 sec cast waits for the cast to end,
// and the Sapper goes on cooldown ahead of it. Blood Fury, Berserking and Elune's Light wait for Death Wish when that costs
// no use. It is 1031.1 (Human) where the old DPS (With Reck) rotation is 981.2. Slam on its 18 sec
// cooldown and holding the fillers for Bloodthirst and Whirlwind both lose.
export const APLLevel60Fury = PresetUtils.makePresetAPLRotation('Level 60 Fury', Level60FuryAPLJSON, { group: 'Level 60' });
// The same template on the two-hand Fury talents and the two-hand set: the Fury rotation with
// Heroic Strike from 85 Rage instead of 40, Hamstring from 20, Bloodrage under 20, Slam from 30 Rage
// right after a main hand swing, ahead of Whirlwind (+8.4 at 180 sec; Slam at any time loses), and
// Death Wish held for the last 36 sec with Recklessness for the last 30. The second potion and the
// trinket's last use are held for the last 20 and 34 sec, and Blood Fury and Berserking for the last
// 30 (Elune's Light keeps waiting for Death Wish, held it lost 7). The 34 and 30 sec holds, from 25,
// are +4.4 / +2.1 / +1.2 on the Orc at 120 / 180 / 300 sec (trinket +1.1 and Blood Fury +0.5 of it
// at 180 sec), Troll +0.9 and Undead +1.3 at 180 sec. With the raid damage hits' Rage (the
// rotation auditor's pass 21, 2026-10-16), Bloodrage moved from under 70 Rage to under 20 and
// Heroic Strike from 90 to 85: +0.7 / +2.5 / +1.9 at 120 / 180 / 300 sec (944.5 at 180).
export const APLLevel60Fury2H = PresetUtils.makePresetAPLRotation('Level 60 Fury (Two-Hand)', Level60Fury2HAPLJSON, { group: 'Level 60' });
// The tools/rotopt search (template warrior_arms) on the Level 60 Arms talents and the two-hand
// set, Human, 180 sec: Battle Stance, Mortal Strike, Slam above 20 Rage (15 sec cooldown with
// Improved Slam), Overpower on dodges and Bloodthrill procs, Rend kept up for Bloodthrill, Execute
// after them in the execute phase, Heroic Strike from 90 Rage (70 before Forever's normalized Rage)
// and Hamstring from 20 Rage as a filler.
// With 25 sec left and Recklessness ready it moves to Berserker Stance for Recklessness (no more
// Rage than Tactical Mastery keeps), then back to Battle Stance once it is out, since Recklessness
// stays up (+6.1 at 180 sec; the rotation auditor's pass 14). The second potion is held for the last 15 sec
// (with Slam and Hamstring at 20, +6.9 at 180 sec, the rotation auditor's pass 2). The rotation audit found the stance swap and
// the Hamstring filler missing: without them Arms did 811.9, with them 868.7 (the auditor's
// numbers, on the earlier gear). Berserker Stance with Whirlwind in place of Overpower and
// Rend for the whole fight was 790.0 with its own talents on the earlier gear, 90 behind. Spearing Strike is worth
// nothing at 15 Rage for 40% weapon damage, except against Giants and Dragonkin (the preset below).
export const APLLevel60Arms = PresetUtils.makePresetAPLRotation('Level 60 Arms', Level60ArmsAPLJSON, { group: 'Level 60' });
// Arms against a Giant or Dragonkin boss (Onyxia is Dragonkin in Phase 1), where Spearing Strike does
// 120% of normalized weapon damage instead of 40%: the Arms rotation with Spearing Strike on cooldown
// from 15 Rage.
export const APLLevel60ArmsDragonkin = PresetUtils.makePresetAPLRotation('Level 60 Arms (vs Dragonkin)', Level60ArmsDragonkinAPLJSON, {
	group: 'Level 60',
});

// The Battle Shout duty (standards 23): the base rotation, plus Battle Shout before the pull (with a
// Bloodrage for its Rage, 3 sec before the pull so its GCD is over by then) and again when 35 sec of it
// are left, unless it lasts the fight. Forever's Battle Shout lasts 3 min, so only a fight over about
// 177 sec refreshes it, at about 142 sec, before the execute phase.
export const APLLevel60FuryShout = PresetUtils.makePresetAPLRotation('Level 60 Fury (Battle Shout duty)', Level60FuryShoutAPLJSON, { group: 'Level 60' });
export const APLLevel60Fury2HShout = PresetUtils.makePresetAPLRotation('Level 60 Fury (Two-Hand) (Battle Shout duty)', Level60Fury2HShoutAPLJSON, {
	group: 'Level 60',
});
export const APLLevel60ArmsShout = PresetUtils.makePresetAPLRotation('Level 60 Arms (Battle Shout duty)', Level60ArmsShoutAPLJSON, { group: 'Level 60' });

export const APLPresets = {
	[Phase.Phase1]: [
		APLLevel60Fury,
		APLLevel60FuryShout,
		APLLevel60Fury2H,
		APLLevel60Fury2HShout,
		APLLevel60Arms,
		APLLevel60ArmsShout,
		APLLevel60ArmsDragonkin,
		APLNoReck,
		AplReck,
	],
};

export const DefaultAPLs = [APLLevel60Fury];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Dual wield Fury, from the tools/rotopt talent search (alternated with the rotation search until
// neither moved) on the synthetic Fury set, Human, 180 sec: 32 Fury to Bloodthirst with Furious
// Precision, Dual Wield Specialization and Flurry, 17 Arms to Impale, and Improved Bloodrage. Enrage
// took a second point from Improved Berserker Rage in the 2026-10-16 search (+2.8 at 180 sec), now
// that the raid damage hits trigger it. Weaponmaster
// hybrids (25 Arms for the swords' extra attacks, no Bloodthirst) were 823.5 on the earlier gear,
// 17% behind.
export const TalentsLevel60Fury = PresetUtils.makePresetTalents('Level 60 Fury', SavedTalents.create({ talentsString: '30305013002-25053005122010501-2' }), {
	tooltip: 'Level 60 dual wield Fury talents from the talent search.',
	group: 'Level 60',
});
// Two-hand Fury from the same search on the two-hand set: Two-Handed Weapon Specialization in place
// of Dual Wield Specialization and Furious Precision. 899.4 at 180 sec (Human), 13% behind dual wield
// Fury but its own way to play, so it gets its own preset.
export const TalentsLevel60Fury2H = PresetUtils.makePresetTalents(
	'Level 60 Fury (Two-Hand)',
	SavedTalents.create({ talentsString: '30305013032-25053000152110501' }),
	{ tooltip: 'Level 60 two-hand Fury talents from the talent search.', group: 'Level 60' },
);
// Arms from the same search on the two-hand set: 35 Arms to Mortal Strike with Bloodthrill,
// Weaponmaster (1% extra attacks a point with a sword), Improved Slam and Impale, and Cruelty,
// Unbridled Wrath and Booming Voice in Fury. Arms cannot reach Death Wish. Booming Voice does
// nothing here, since the raid has Battle Shout, but no other talent in reach does better with those
// points. 912.5 at 180 sec (Human), 12% behind dual wield Fury and just ahead of two-hand Fury, its
// own preset as the Mortal Strike spec.
export const TalentsLevel60Arms = PresetUtils.makePresetTalents('Level 60 Arms', SavedTalents.create({ talentsString: '20305213032515201-55050000001' }), {
	tooltip: 'Level 60 Arms talents from the talent search.',
	group: 'Level 60',
});
// The talent and rotation searches on a Dragonkin boss take Spearing Strike and Improved Bloodrage
// for two points of Booming Voice and one of Improved Execute.
export const TalentsLevel60ArmsDragonkin = PresetUtils.makePresetTalents(
	'Level 60 Arms (vs Dragonkin)',
	SavedTalents.create({ talentsString: '20305213132515201-3505-2' }),
	{
		tooltip: 'Level 60 Arms talents against a Giant or Dragonkin boss, with Spearing Strike.',
		group: 'Level 60',
	},
);

// The old Classic DPS talents, rebuilt for the Forever tree on 2026-10-12 (the points in talents
// Forever removed went to Furious Precision, Improved Execute and Improved Berserker Rage).
export const TalentsP1DPS = PresetUtils.makePresetTalents('DPS', SavedTalents.create({ talentsString: '30305013-25053005152110501' }));

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60Fury, TalentsLevel60Fury2H, TalentsLevel60Arms, TalentsLevel60ArmsDragonkin, TalentsP1DPS],
};

export const DefaultTalents = TalentsLevel60Fury;

///////////////////////////////////////////////////////////////////////////
//                                 Options Presets
///////////////////////////////////////////////////////////////////////////

// The raid damage in the Level 60 encounters (PresetUtils.LEVEL_60_RAID_DAMAGE, the same for every
// spec): 2 direct hits a minute at random times, 1500 damage each. Under Forever each gives 3.4
// Rage in any stance (core.ForeverDamageTakenRage: 20 Rage for a level 60 mob's 5356 health, as if
// 40% armor took its share, before any of our own damage taken modifiers), and has Enrage's 30%
// chance, because Forever's Enrage fires on any damage taken. Since our stance and Recklessness no
// longer change it (2026-10-16), Fury moved by -0.2 to -0.6.
// At 180 sec (10,000 iterations, the report requests) that is +3.4 on Fury (1040.4 to 1043.8),
// +11.0 on Fury two-hand (921.0 to 932.0), -1.1 on Arms (945.6 to 944.5: Slam's cast is pushed
// back) and -0.1 on Arms vs Dragonkin; at 10 hits a minute +16.0, +47.5, -4.1 and -3.3. With
// Classic's damage-taken Rage (15.4 a hit) each build was 4 to 5 higher at 2 a minute.

export const DefaultOptions = WarriorOptions.create({
	queueDelay: 250,
	startingRage: 0,
	shout: WarriorShout.WarriorShoutBattle,
	stance: WarriorStance.WarriorStanceBerserker,
});

// The class options the preset builds set. Fury starts in Berserker Stance, Arms in Battle Stance
// (Overpower, Rend and Bloodthrill need it).
const PresetBuildOptionsFury = {
	queueDelay: 250,
	startingRage: 0,
	shout: WarriorShout.WarriorShoutBattle,
	stance: WarriorStance.WarriorStanceBerserker,
};
const PresetBuildOptionsArms = { ...PresetBuildOptionsFury, stance: WarriorStance.WarriorStanceBattle };

// The level 60 consumables. On 2026-10-12 we tried every consumable the sim applies in each slot on
// the Level 60 Fury build (Undead) at 180 sec. A character takes at most one of Forever's five +2%
// crit elixirs (Mongoose, Grizzly, Owl, Sages, Whale; the client lists them as one elixir in role
// variants). Mongoose with Forever's Elixir of Ferocity (18 Strength and 18 Agility) is the best pair
// on all three builds: Fury 992.7, against 991.5 with Juju Power, 987.3 for Grizzly with Greater
// Agility and 987.2 with Elixir of Giants (on the earlier gear). Taking Grizzly as well would add
// 15.1 on Fury, 12.0 on Fury two-hand and 10.9 on Arms on the final gear (both crit elixirs together
// were +73.2 over neither). The rest of the set we started from won its slot: Juju Might
// (+13.9 over none), Smoked Desert Dumplings (+0.8 over Mightfish Steak, +7.2 over Flank au Poivre),
// the Mighty Rage Potion (+9.5 over the Major Frenzy Potion), R.O.I.D.S. and Dragonbreath Chili.
// Goblin Land Mine between Sappers is +1.2 on Fury, +3.4 on Fury two-hand and +0.9 on Arms, a little
// ahead of Thorium Grenades and Dense Dynamite (EZ-Thro Dark Bomb in its place is 6 to 15 behind:
// the rotations time the Land Mine). The flask was Flask of the Titans, which does nothing for the
// DPS, until the natural flasks (2026-10-23, below).
//
// The presets only take what Phase 1 can get without a raid that is not open yet. Elemental
// Sharpening Stone's plans drop only in Molten Core, so the swords get Dense Sharpening Stones (10
// and 13 behind). The Smoked Desert Dumplings come from an AQ quest. They stand in for Forever's own
// +20 Strength foods (Bear Bruscitti and Steaming Stag Steak in client 70338), which the sim does not
// have yet.
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MightyRagePotion,
	dragonBreathChili: true,
	fillerExplosive: Explosive.ExplosiveGoblinLandMine,
	// The natural flask (standards 11, 2026-10-23): Precision, +66.4 / +68.1 / +10.0 over the Flask of
	// the Titans on Fury / Fury two-hand / Arms at 180 sec. Arms would do better with Aggression (+31.8),
	// but the builds share these consumes. Accuracy is +34.3 on dual wield Fury and nothing with a
	// two-hander (the yellow hit is capped).
	flask: Flask.FlaskOfNaturalPrecision,
	food: Food.FoodSmokedDesertDumpling,
	mainHandImbue: WeaponImbue.DenseSharpeningStone,
	offHandImbue: WeaponImbue.DenseSharpeningStone,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.ROIDS,
});

// The raid the level 60 numbers were measured under (tools/rotopt/mksettings, spec fury and arms):
// the shared level 60 raid, Battle Shout from another warrior, and a shaman in the warrior's group
// with Strength of Earth and Windfury Totem, and Grace of Air from a second shaman in the raid as well (the user's ruling 2026-10-16, "both totems": client 1.60.1.70425 lets Grace of Air stack with Windfury Totem and makes both raid-wide).
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
	salvation: true,
});

// Kings and Might. A Rage user has no use for Wisdom.
export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	fengusFerocity: false,
	rallyingCryOfTheDragonslayer: false,
	songflowerSerenade: false,
	warchiefsBlessing: false,
});

// The shared level 60 raid's debuffs. The ones that matter to a warrior are Sunder Armor, Faerie
// Fire, Curse of Recklessness and Expose Armor.
export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	curseOfRecklessness: true,
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	demoralizingRoar: TristateEffect.TristateEffectImproved,
	demoralizingShout: TristateEffect.TristateEffectImproved,
	exposeArmor: TristateEffect.TristateEffectMissing,
	faerieFire: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	scorpidSting: true,
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
});

// Orc is the best race for every warrior build since client 70425, whose Savage gives 1% crit with
// any weapon (it was Axe Specialization), and Human, whose Ruthless gives 2% crit with any weapon,
// is the best Alliance race. No racial depends on the weapon any more. See sim-status.md for the
// race table. Engineering is for the Goblin Sapper Charge.
export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceOrc,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor, type Unknown so that no
// creature-type bonus applies, standards 19), a 3 minute fight, and the default buffs, debuffs and
// consumes. The synthetic Phase 1 sets, their talents and rotations were searched at 180 sec.
export const EncounterLevel60 = PresetUtils.makePresetEncounter(
	'Level 60',
	Encounter.create({
		duration: 180,
		durationVariation: 20,
		executeProportion20: 0.2,
		executeProportion25: 0.25,
		executeProportion35: 0.35,
		...PresetUtils.LEVEL_60_RAID_DAMAGE,
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
		tanks: [],
		raidBuffs: DefaultRaidBuffs,
		debuffs: DefaultDebuffs,
		buffs: DefaultIndividualBuffs,
		consumes: DefaultConsumes,
	},
);

// The Level 60 raid without its Battle Shout, for the Battle Shout duty builds, whose warrior
// shouts for the party (standards 23).
export const EncounterLevel60BattleShoutDuty = PresetUtils.makePresetEncounter('Level 60 (Battle Shout duty)', EncounterLevel60.encounter, {
	tanks: [],
	raidBuffs: RaidBuffs.create({ ...DefaultRaidBuffs, battleShout: TristateEffect.TristateEffectMissing }),
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumes,
});

// The same boss as a Dragonkin, for Spearing Strike (standards 19).
export const EncounterLevel60Dragonkin = PresetUtils.makePresetEncounter(
	'Level 60 (vs Dragonkin)',
	Encounter.create({
		duration: 180,
		durationVariation: 20,
		executeProportion20: 0.2,
		executeProportion25: 0.25,
		executeProportion35: 0.35,
		...PresetUtils.LEVEL_60_RAID_DAMAGE,
		...PresetUtils.LEVEL_60_ZONE,
		targets: [
			{
				id: 213336,
				name: 'Level 60',
				level: 63,
				mobType: MobType.MobTypeDragonkin,
				stats: new Stats().withStat(Stat.StatAttackPower, 805).withStat(Stat.StatArmor, 3731).withStat(Stat.StatHealth, 127393).asArray(),
				minBaseDamage: 3000,
				damageSpread: 0.3333,
				swingSpeed: 2,
				parryHaste: true,
			},
		],
	}),
	{
		tanks: [],
		raidBuffs: DefaultRaidBuffs,
		debuffs: DefaultDebuffs,
		buffs: DefaultIndividualBuffs,
		consumes: DefaultConsumes,
	},
);

// On 2026-10-15 at 20,000 iterations, 120 / 180 / 300 sec: Level 60 Fury (Undead) 1091.9 / 1057.9 /
// 1035.9, Level 60 Fury (Two-Hand) (Orc) 962.5 / 925.7 / 895.4 and Level 60 Arms (Orc) 963.5 /
// 951.0 / 925.4, with Forever's normalized Rage at s = 0.5 (standards 21). Arms keeps Rend up, which
// adds 2% of attack power a tick under Forever (report assumption warr_rend_ap, user ruling
// 2026-10-15): +19.5 at 180 sec over Rend with no attack power (933.1); the Fury builds don't cast it. The weapons have the type
// and speed each build sims best with among Era Phase 1 BiS speeds (standards 22): a 2.8 sec sword
// and a 1.7 sec axe for Fury (a 2.9 sec axe since 2026-10-16, below), a 3.7 sec axe for the two-hand builds. On the Era BiS swords the
// builds did 1038.1, 901.6 and 915.1 at 180 sec (Undead). The numbers below were measured on those swords. How k grows
// from level 30 to 60 is an assumption: flat (s = 0) is 950.6 / 835.0 / 861.4 at 180 sec, Classic-like (s = 1)
// 1076.1 / 924.2 / 935.0, and the off hand at half Rage (s = 0.5) 945.1 on Fury (-8.9%; the
// two-hand builds have no off hand). The Ahn'Qiraj book ranks being out (Heroic Strike VIII, the
// raid's Battle Shout VI, Strength of Earth and Grace of Air) cost 22.6 / 14.8 / 14.2 at 180 sec, and
// Deep Wounds' ticks not critting 1.6 / 6.7 / 7.2 (both measured before normalized Rage). With weapon
// racials off, Fury stays at its Undead numbers, and the Undead does 920.4 on two-hand Fury and 929.4
// on Arms at 180 sec (before Rend's attack power share). Every build spends its first GCD after the
// pull on one Sunder Armor (standards 23), which cost 4.8 / 3.5 / 2.3 on Fury, 2.1 / 1.8 / 2.3 on
// two-hand Fury and 5.4 / 1.6 / 1.6 on Arms; the numbers above include it.
//
// The Battle Shout duty builds (standards 23) shout for the party themselves, on the raid without
// its Battle Shout.
//
// On 2026-10-16 (client 70425's racials, the raid damage hits; 20,000 iterations, 120 / 180 / 300
// sec): Fury (Orc) 1093.5 / 1055.1 / 1031.4 (Undead, its race before, 1078.0 / 1043.8 / 1021.5),
// two-hand Fury (Orc) 973.5 / 932.3 / 903.6, Arms (Orc) 959.1 / 944.2 / 920.4 and Arms vs Dragonkin
// (Orc) 1000.0 / 982.4 / 958.3. On Battle Shout duty: Fury 1092.6 / 1053.2 / 1029.8, two-hand Fury
// 972.2 / 929.4 / 901.9 and Arms 959.4 / 942.5 / 918.2, so the duty costs 0.9 / 1.9 / 1.6,
// 1.3 / 2.9 / 1.7 and -0.3 / 1.7 / 2.2. The numbers above this paragraph are from before.
// On 2026-10-16, after the user's rulings that the off hand makes half the Rage (#75) and that
// Windfury Totem procs only from the main hand (#86), with Grace of Air in the melee raid (#88),
// Fury was re-searched (Orc, 20,000 iterations, 120 / 180 / 300 sec). The off hand became a 2.9
// sec axe (Deathbringer's speed; slow off hands from 2.4 to 2.9 sec were 5 to 7 ahead of the 1.7
// sec axe, the 2.8 sec main hand stayed best), Enrage took a point from Improved Berserker Rage,
// and Heroic Strike moved to 50 Rage: 1002.3 / 972.0 / 953.8 to 1019.9 / 987.9 / 968.7, the Battle
// Shout duty 998.2 / 968.1 / 951.4 to 1016.2 / 984.0 / 965.6. Two-hand Fury, unchanged at 1007.8 /
// 969.6 / 944.9, stays behind. Orc is still the best race (5.7 ahead of High Order Skyborne, the
// best Alliance race now, at 180 sec on two seed sets).
// The rotation auditor's pass 23 (2026-10-16): Recklessness gets its own line while Death Wish is
// up, just ahead of the other cooldowns line, so the trinket goes one GCD after it instead of in the
// same pass (+1.2 / +1.0 / +0.6 on Fury, +1.1 / +1.0 / +0.7 on the Battle Shout duty at 120 / 180 /
// 300 sec, on two seed sets). With it, Heroic Strike from 55 Rage beats 50 on both seed sets at
// every length (+1.4 to +2.1 on Fury; the duty +0.7 / +1.1 at 180 sec). Fury 1018.1 / 986.5 /
// 967.4 to 1020.8 / 989.0 / 970.0, the duty 1015.5 / 983.5 / 964.6 to 1018.1 / 985.2 / 967.3.
export const PresetBuildLevel60Fury = PresetUtils.makePresetBuild('Level 60 Fury', {
	group: 'Level 60',
	tooltip: 'Level 60 dual wield Fury on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	gear: GearLevel60Phase1SyntheticFury,
	talents: TalentsLevel60Fury,
	rotation: APLLevel60Fury,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsFury,
});
export const PresetBuildLevel60FuryShout = PresetUtils.makePresetBuild('Level 60 Fury (Battle Shout duty)', {
	group: 'Level 60',
	tooltip: 'Level 60 dual wield Fury keeping Battle Shout up. The warrior shouts for the party, so the raid has no other Battle Shout.',
	gear: GearLevel60Phase1SyntheticFury,
	talents: TalentsLevel60Fury,
	rotation: APLLevel60FuryShout,
	encounter: EncounterLevel60BattleShoutDuty,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsFury,
});
export const PresetBuildLevel60Fury2H = PresetUtils.makePresetBuild('Level 60 Fury (Two-Hand)', {
	group: 'Level 60',
	tooltip: 'Level 60 two-hand Fury on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	gear: GearLevel60Phase1SyntheticArms,
	talents: TalentsLevel60Fury2H,
	rotation: APLLevel60Fury2H,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsFury,
});
export const PresetBuildLevel60Fury2HShout = PresetUtils.makePresetBuild('Level 60 Fury (Two-Hand) (Battle Shout duty)', {
	group: 'Level 60',
	tooltip: 'Level 60 two-hand Fury keeping Battle Shout up. The warrior shouts for the party, so the raid has no other Battle Shout.',
	gear: GearLevel60Phase1SyntheticArms,
	talents: TalentsLevel60Fury2H,
	rotation: APLLevel60Fury2HShout,
	encounter: EncounterLevel60BattleShoutDuty,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsFury,
});
export const PresetBuildLevel60Arms = PresetUtils.makePresetBuild('Level 60 Arms', {
	group: 'Level 60',
	tooltip: 'Level 60 Arms on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	gear: GearLevel60Phase1SyntheticArms,
	talents: TalentsLevel60Arms,
	rotation: APLLevel60Arms,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsArms,
});
export const PresetBuildLevel60ArmsShout = PresetUtils.makePresetBuild('Level 60 Arms (Battle Shout duty)', {
	group: 'Level 60',
	tooltip: 'Level 60 Arms keeping Battle Shout up. The warrior shouts for the party, so the raid has no other Battle Shout.',
	gear: GearLevel60Phase1SyntheticArms,
	talents: TalentsLevel60Arms,
	rotation: APLLevel60ArmsShout,
	encounter: EncounterLevel60BattleShoutDuty,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsArms,
});
// Against a Dragonkin boss (Orc, 20,000 iterations, 120 / 180 / 300 sec): 1002.5 / 988.6 / 962.8
// with the Sunder opener (1008.2 / 991.0 / 964.5 before), where the generic Arms build did 968.9 /
// 952.6 / 927.0 before the Sunder opener (+4.0% at 180 sec). Fury doesn't get one:
// Spearing Strike needs Battle Stance, and a stance dance for it, with the talent searches, is +0.5%
// on dual wield Fury and +0.8% on two-hand Fury at 180 sec.
export const PresetBuildLevel60ArmsDragonkin = PresetUtils.makePresetBuild('Level 60 Arms (vs Dragonkin)', {
	group: 'Level 60',
	tooltip: 'Level 60 Arms with Spearing Strike, against a Giant or Dragonkin boss such as Onyxia.',
	gear: GearLevel60Phase1SyntheticArms,
	talents: TalentsLevel60ArmsDragonkin,
	rotation: APLLevel60ArmsDragonkin,
	encounter: EncounterLevel60Dragonkin,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsArms,
});
