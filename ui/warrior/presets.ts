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
import Level60FuryAPLJSON from './apls/level60_fury.apl.json';
import Level60Fury2HAPLJSON from './apls/level60_fury_2h.apl.json';
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
// neck, back, wrist, waist, rings, one-handers and bow, 76 for a two-hander): Strength, Agility,
// crit, 9% hit (the special attack cap) and some attack power, no Spirit. Only the head and legs
// carry Battlegear of Glory, Forever's Molten Core set for Arms and Fury, so only its 2-piece is
// live. The 4-piece (+36 attack power against Humanoids) would add 11.7 on Fury, 9.3 on two-hand
// Fury and 9.4 on Arms. The 5-piece adds nothing, since Recklessness is used once a fight. Arms and
// Fury share the armor, rings and bow, the file has the details.
//
// Enchants: Lesser Arcanum of Voracity (8 Strength) on the head and legs, Forever's +5 Strength on
// the neck (+3.7 against +2.9 for Agility on Fury), Forever's Cloak - Agility (5), Greater Stats,
// Superior Strength on the bracers, Forever's Gloves - Superior Strength (15), Greater Agility on the
// boots (a Forceful Rugged Armor Kit there is within 0.3) and Crusader. Nothing for the shoulders
// outside raids. The trinkets are Blackhand's Breadth and Forever's Adaptive Combat Assistant (2
// expertise, a world drop), 10.0 ahead of Hand of Justice on Fury and 4.6 on two-hand Fury, and 0.7
// behind on Arms (inside the noise). Raid trinkets are out.
//
// On 2026-10-12 the Level 60 Fury build (Human, 180 sec) does 995.2 on the Fury set, where the Era
// Phase 1 set (phase_1.gear.json) does 1010.8 with the same talents and rotation. The synthetic set
// has no weapon skill (Edgemaster's Handguards' +7) and the shared budget prices hit and crit
// higher than the Era items did.
export const GearLevel60Phase1SyntheticFury = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic (Fury)', Level60Phase1SyntheticFuryGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever for a dual wield warrior, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
// The same armor with a two-hand sword in place of Bonereaver's Edge, for Arms and two-hand Fury.
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
// talents and set, Human, 180 sec: Bloodthirst, then Whirlwind with 25 Rage, Execute first in the
// execute phase, Heroic Strike queued from 20 Rage, and Hamstring as a filler from 10 Rage (its hits
// proc Windfury Totem's extra attacks and Flurry). Death Wish waits for the last 33 sec,
// Recklessness for the last 25, both inside the execute phase. Bloodrage under 40 Rage, no
// Berserker Rage (one point of the talent is not worth the GCD), the Mighty Rage Potion at the pull.
// It is 995.2 where the old DPS (With Reck) rotation is 978.4. Slam on its 18 sec cooldown and
// holding the fillers for Bloodthirst and Whirlwind both lose.
export const APLLevel60Fury = PresetUtils.makePresetAPLRotation('Level 60 Fury', Level60FuryAPLJSON, { group: 'Level 60' });
// The same template on the two-hand Fury talents and the two-hand set: the Fury rotation with
// Heroic Strike from 70 Rage instead of 20 (a slow two-hander makes less Rage), Bloodrage under 70
// and Death Wish for the last 30 sec.
export const APLLevel60Fury2H = PresetUtils.makePresetAPLRotation('Level 60 Fury (Two-Hand)', Level60Fury2HAPLJSON, { group: 'Level 60' });
// The tools/rotopt search (template warrior_arms) on the Level 60 Arms talents and the two-hand
// set, Human, 180 sec: Battle Stance, Mortal Strike, Slam above 30 Rage (15 sec cooldown with
// Improved Slam), Overpower on dodges and Bloodthrill procs, Rend kept up for Bloodthrill, Execute
// after them in the execute phase, Heroic Strike from 70 Rage and Hamstring from 10 Rage as a filler.
// With 20 sec left and Recklessness ready it moves to Berserker Stance for Recklessness (no more
// Rage than Tactical Mastery keeps) and stays there. The rotation audit found the stance swap and
// the Hamstring filler missing: without them Arms did 811.9, with them 868.7 (the auditor's
// numbers, before the Grizzly elixir). Berserker Stance with Whirlwind in place of Overpower and
// Rend for the whole fight is 790.0 with its own talents, 90 behind. Spearing Strike is worth
// nothing at 15 Rage for 40% weapon damage.
export const APLLevel60Arms = PresetUtils.makePresetAPLRotation('Level 60 Arms', Level60ArmsAPLJSON, { group: 'Level 60' });

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60Fury, APLLevel60Fury2H, APLLevel60Arms, APLNoReck, AplReck],
};

export const DefaultAPLs = [APLLevel60Fury];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Dual wield Fury, from the tools/rotopt talent search (alternated with the rotation search until
// neither moved) on the synthetic Fury set, Human, 180 sec: 32 Fury to Bloodthirst with Furious
// Precision, Dual Wield Specialization and Flurry, 17 Arms to Impale, and Improved Bloodrage. Enrage
// is worth nothing in a raid (the boss does not hit the warrior), so it keeps one point. Weaponmaster
// hybrids (25 Arms for the swords' extra attacks, no Bloodthirst) are 823.5, 17% behind.
export const TalentsLevel60Fury = PresetUtils.makePresetTalents('Level 60 Fury', SavedTalents.create({ talentsString: '30305013002-25053005112110501-2' }), {
	tooltip: 'Level 60 dual wield Fury talents from the talent search.',
	group: 'Level 60',
});
// Two-hand Fury from the same search on the two-hand set: Two-Handed Weapon Specialization in place
// of Dual Wield Specialization and Furious Precision. 873.8 at 180 sec (Human), 12% behind dual wield
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
// points. 879.7 at 180 sec (Human), 12% behind dual wield Fury, its own preset as the Mortal Strike
// spec.
export const TalentsLevel60Arms = PresetUtils.makePresetTalents('Level 60 Arms', SavedTalents.create({ talentsString: '20305213032515201-55050000001' }), {
	tooltip: 'Level 60 Arms talents from the talent search.',
	group: 'Level 60',
});

// The old Classic DPS talents, rebuilt for the Forever tree on 2026-10-12 (the points in talents
// Forever removed went to Furious Precision, Improved Execute and Improved Berserker Rage).
export const TalentsP1DPS = PresetUtils.makePresetTalents('DPS', SavedTalents.create({ talentsString: '30305013-25053005152110501' }));

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60Fury, TalentsLevel60Fury2H, TalentsLevel60Arms, TalentsP1DPS],
};

export const DefaultTalents = TalentsLevel60Fury;

///////////////////////////////////////////////////////////////////////////
//                                 Options Presets
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = WarriorOptions.create({
	queueDelay: 250,
	startingRage: 0,
	shout: WarriorShout.WarriorShoutBattle,
	stance: WarriorStance.WarriorStanceBerserker,
});

// The class options the preset builds set. Fury starts in Berserker Stance, Arms in Battle Stance
// (Overpower, Rend and Bloodthrill need it).
const PresetBuildOptionsFury = { queueDelay: 250, startingRage: 0, shout: WarriorShout.WarriorShoutBattle, stance: WarriorStance.WarriorStanceBerserker };
const PresetBuildOptionsArms = { ...PresetBuildOptionsFury, stance: WarriorStance.WarriorStanceBattle };

// The level 60 consumables. On 2026-10-12 we tried every consumable the sim applies in each slot on
// the Level 60 Fury build (Undead) at 180 sec. A character takes at most one of Forever's five +2%
// crit elixirs (Mongoose, Grizzly, Owl, Sages, Whale; the client lists them as one elixir in role
// variants). Mongoose with Forever's Elixir of Ferocity (18 Strength and 18 Agility) is the best pair
// on all three builds: Fury 992.7, against 991.5 with Juju Power, 987.3 for Grizzly with Greater
// Agility and 987.2 with Elixir of Giants. Taking Grizzly as well would add 15.3 on Fury and 11 on
// Fury two-hand and Arms (both crit elixirs together are +73.2 over neither). The rest of the set we started from won its slot: Juju Might
// (+13.9 over none), Smoked Desert Dumplings (+0.8 over Mightfish Steak, +7.2 over Flank au Poivre),
// the Mighty Rage Potion (+9.5 over the Major Frenzy Potion), R.O.I.D.S. and Dragonbreath Chili.
// Goblin Land Mine between Sappers is +1.2 on Fury, +3.4 on Fury two-hand and +0.9 on Arms, a little
// ahead of Thorium Grenades and Dense Dynamite. Flask of Supreme Power is +7, all of it Dragonbreath
// Chili scaling with spell power, so it stays out.
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
	flask: Flask.FlaskOfTheTitans,
	food: Food.FoodSmokedDesertDumpling,
	mainHandImbue: WeaponImbue.DenseSharpeningStone,
	offHandImbue: WeaponImbue.DenseSharpeningStone,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.ROIDS,
});

// The raid the level 60 numbers were measured under (tools/rotopt/mksettings, spec fury and arms):
// the shared level 60 raid, Battle Shout from another warrior, and a shaman in the warrior's group
// with Strength of Earth and Windfury Totem. Forever's air totems no longer stack in a party, so the
// shaman picks Windfury over Grace of Air for a warrior.
export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	bloodPact: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	frostResistanceAura: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	leaderOfThePack: true,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	sanctityAura: true,
	shadowProtection: true,
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
	thorns: TristateEffect.TristateEffectImproved,
	totemWeaponBuff: TotemWeaponBuff.TotemWeaponBuffWindfury,
	trueshotAura: true,
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
	curseOfShadow: true,
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	demoralizingRoar: TristateEffect.TristateEffectImproved,
	demoralizingShout: TristateEffect.TristateEffectImproved,
	exposeArmor: TristateEffect.TristateEffectImproved,
	faerieFire: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	scorpidSting: true,
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
});

// Undead is the best race for all three builds (Touch of the Grave rolls on every ability press,
// Hamstring included), Human next (2% crit with the sword). See sim-status.md for the race table. Engineering is for the
// Goblin Sapper Charge.
export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceUndead,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor, Humanoid), a 3 minute
// fight, and the default buffs, debuffs and consumes. The synthetic Phase 1 sets, their talents and
// rotations were searched at 180 sec.
export const EncounterLevel60 = PresetUtils.makePresetEncounter(
	'Level 60',
	Encounter.create({
		duration: 180,
		durationVariation: 20,
		executeProportion20: 0.2,
		executeProportion25: 0.25,
		executeProportion35: 0.35,
		targets: [
			{
				id: 213336,
				name: 'Level 60',
				level: 63,
				mobType: MobType.MobTypeHumanoid,
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

// On 2026-10-12 at 20,000 iterations, 120 / 180 / 300 sec: Level 60 Fury (Undead) 1060.2 / 1004.3 /
// 981.8, Level 60 Fury (Two-Hand) (Undead) 926.5 / 876.3 / 855.9 and Level 60 Arms (Undead) 906.3 /
// 882.7 / 866.4.
export const PresetBuildLevel60Fury = PresetUtils.makePresetBuild('Level 60 Fury', {
	group: 'Level 60',
	tooltip: 'Level 60 dual wield Fury on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	gear: GearLevel60Phase1SyntheticFury,
	talents: TalentsLevel60Fury,
	rotation: APLLevel60Fury,
	encounter: EncounterLevel60,
	race: Race.RaceUndead,
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
	race: Race.RaceUndead,
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
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsArms,
});
