import { Phase } from '../core/constants/other';
import * as PresetUtils from '../core/preset_utils';
import {
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	FirePowerBuff,
	Flask,
	Food,
	FrostPowerBuff,
	IndividualBuffs,
	IntellectElixir,
	ManaRegenElixir,
	MobType,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	SpellPowerBuff,
	SpiritElixir,
	Stat,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common';
import { Mage_Options as MageOptions, Mage_Options_ArmorType as ArmorType, Mage_Options_StaffImbue as StaffImbue } from '../core/proto/mage';
import { SavedTalents } from '../core/proto/ui';
import { Stats } from '../core/proto_utils/stats';
import Level60ArcaneAPL from './apls/level60_arcane.apl.json';
import Level60FireAPL from './apls/level60_fire.apl.json';
import Level60FrostAPL from './apls/level60_frost.apl.json';
import Level60FrostfireAPL from './apls/level60_frostfire.apl.json';
import P1APL from './apls/p1.apl.json';
import Level60Phase1SyntheticGear from './gear_sets/level60_phase1_synthetic.gear.json';
import Level60Phase1SyntheticArcaneGear from './gear_sets/level60_phase1_synthetic_arcane.gear.json';
import P0BISGear from './gear_sets/p0.bis.gear.json';
import P1BISGear from './gear_sets/p1.bis.gear.json';

// The Level 60 presets were searched on 2026-10-12, and again on 2026-10-13 on the Forever
// budget gear, with tools/rotopt (the mage_arcane, mage_fire, mage_frost and mage_frostfire
// templates, mage_presets.sh) against the level 60 raid in tools/rotopt/mksettings
// (spec_mage.go): the Level 60 boss (level 63, 3731 armor, a generic boss of type Unknown),
// Blessings of Kings and Wisdom, Arcane Brilliance, Mana Spring Totem, Moonkin Aura, Curse of the
// Elements and the other shared debuffs. Improved Scorch and Winter's Chill are the mage's own
// under Forever, so the raid gives none. The DPS below is at 120 / 180 / 300 sec with 20000
// iterations, on the best race for the build.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearP0BIS = PresetUtils.makePresetGear('Pre-BiS', P0BISGear);
export const GearP1BIS = PresetUtils.makePresetGear('P1 BiS', P1BISGear);

// A guess at Phase 1 BiS under Forever, made of the synthetic items in
// tools/database/forever_synthetic_items_mage.go (ids from 996001, not in the game), on the
// cross-class standards every class's Level 60 preset follows. Each piece is priced on the
// Forever budget (standards 1b): spell power at 0.86, each armor piece at its slot's share of
// L = 1.125 x item level + 2.0 times the caster factor, with four stat lines like a real caster
// epic, at one item level table for everyone (66 for head, shoulders, chest, hands, legs and
// feet, 71 for the rest), and Forever's caster weapon rate for the staff (193.4) and wand (21.3).
// Armor totals 708, weapons 215. The stats are a fair mix in the proportions of Season of
// Discovery's level 60 BiS mage sets (standards 5a), with hit at the cap. Frostfire, Fire and
// Frost share one set: 421 spell power, 151 Intellect, 119 Stamina, 88 Spirit, 10% hit and 103
// crit rating (7.4%). Arcane has its own set with more spell power and less crit: 460 spell
// power, 147 Intellect, 70 crit rating (5%), the rest about the same. With Elemental Precision or
// Arcane Focus and Manaflare Regalia's 2 piece bonus (1% hit, on the head and chest only) the hit
// is the 16% cap against a level 63 boss; 2% less hit for the other stats loses 7 to 11. The 4
// and 5 piece bonuses would add 56.8 (6%) to Frostfire at 180 sec and nothing to the others (the
// 5 piece is all Frostfire Bolt, the 4 piece only against Elementals).
//
// The trinkets are synthetic too (Forever's real Phase 1 trinkets are gone or not simulated):
// 41 spell power and 11 crit rating, and an on-use +140 spell power for 20 sec every 2 min.
//
// The weapon is a staff with Staff of Dominance's damage. With Forever's 2H Spellblasting (+65
// spell power) it is 32 to 43 ahead of the sword and off hand, and Human with the sword is still
// behind Undead with the staff. The enchants are the best Phase 1 ones without a raid source:
// Lesser Arcanum of Voracity (Intellect), +6 spell power on the neck, Mystic Rugged Armor Kits
// (+6 spell power) on chest, hands, legs and feet, Greater Spellpower (+16) on the wrist and
// Spellblasting on the staff. Arcane wears Gloves - Arcane Power (+20 Arcane) instead of the kit,
// 7.7 ahead.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
export const GearLevel60Phase1SyntheticArcane = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic (Arcane)', Level60Phase1SyntheticArcaneGear, {
	tooltip: 'The Level 60 synthetic Arcane set (more spell power, less crit) with Gloves - Arcane Power.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1Synthetic, GearLevel60Phase1SyntheticArcane, GearP0BIS, GearP1BIS],
};

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The Classic rotation the fork shipped. It still runs, but it was written for the Classic tree.
export const APLP1DPS = PresetUtils.makePresetAPLRotation('DPS', P1APL);

// Frostfire Bolt as the filler. Scorch puts up 5 stacks of Fire Vulnerability and keeps them,
// Pyroblast waits for 3 Heating Up stacks (a 1.5 sec cast), Ice Lance goes on every Fingers of
// Frost charge (with SoD's 42.9% spell power coefficient, +66 at 180 sec over none and +7 over
// only the last charge), Arcane Missiles takes a Missile Barrage proc. Fire Blast only goes with
// mana the fight can't use (above 150 a second of the fight left, +1.5 to +3.7); on cooldown its
// mana is worth more as Frostfire Bolts. Evocation below 30% mana with at least 60 sec of the
// fight left, then rank 2 Frostfire Bolt when the mana won't last. The knobs were searched over
// 120, 180 and 300 sec together; with Ice Lance (2026-10-15) the search at 180 sec moved
// Evocation up from below 10% with 30 sec left, for Ice Lance's mana: +6.2 at 180 sec, level at
// 120 and -3.9 at 300 (below 20% with 60 sec left would be -0.4 at 180 and +3.0 at 300). The
// on-use trinket waits for the 5th Fire Vulnerability stack (+3.7 at 180 sec). The Fingers of
// Frost Ice Lance goes ahead of Scorch and Pyroblast (+18.9 / +17.9 / +19.4 at 120 / 180 / 300
// sec), and when a Frostfire Bolt would not land before the end (its cast plus 1.5 sec) Ice Lance
// goes instead (+26.5 / +22.8 / +22.2 with both, the rotation auditor's pass 19).
export const APLLevel60Frostfire = PresetUtils.makePresetAPLRotation('Level 60 Frostfire', Level60FrostfireAPL, {
	tooltip: 'Frostfire Bolt, Pyroblast on 3 Heating Up stacks, Ice Lance on Fingers of Frost, Scorch for Fire Vulnerability.',
	group: 'Level 60',
});
// Frostbolt, with Ice Lance on every Fingers of Frost charge (4 times the damage on a Frozen
// target and Shatter's crit; with SoD's 42.9% spell power coefficient it is +175 at 180 sec over
// no Ice Lance and +18 over only the last charge). Frost rarely runs low on mana, so Evocation
// almost never goes. When a Frostbolt would not land before the end (its cast plus 1 sec), Ice
// Lance instead (+11.0 / +6.8 / +4.0 at 120 / 180 / 300 sec, the rotation auditor's pass 19).
export const APLLevel60Frost = PresetUtils.makePresetAPLRotation('Level 60 Frost', Level60FrostAPL, {
	tooltip: 'Frostbolt, with Ice Lance on Fingers of Frost.',
	group: 'Level 60',
});
// Scorch to 5 Fire Vulnerability stacks and before they run out, Pyroblast at 3 Heating Up
// stacks, Arcane Missiles on Missile Barrage, Fire Blast above 20% mana, and Fireball. Pyroblast
// at 1 or 2 stacks, Frostfire Bolt or Scorch as the filler and holding Fire Blast while Heating Up
// is full all lost. Combustion and the on-use trinket wait for the 5th Fire Vulnerability stack
// (the trinket +1.6 at 180 sec). When a Fireball would not land before the end (its cast plus
// 1.04 sec), Fire Blast and Scorch instead (+6.3 / +4.6 / +2.8, the rotation auditor's pass 19).
export const APLLevel60Fire = PresetUtils.makePresetAPLRotation('Level 60 Fire', Level60FireAPL, {
	tooltip: 'Fireball, Pyroblast on 3 Heating Up stacks, Scorch for Fire Vulnerability, Fire Blast.',
	group: 'Level 60',
});
// Arcane Blast to 2 stacks, then Arcane Missiles on a Missile Barrage proc. Without one, Arcane
// Blast goes on to 4 stacks and the missiles go anyway, ending the stacks and taking their
// bonus. Below 20% mana one Arcane Blast and the missiles (30% was 1 behind on the fair mix gear, and the Forever budget gear kept it).
// That is ahead of three Arcane Blasts and the missiles whatever the proc (the template's
// default). No Evocation. Arcane Power and Presence of Mind wait for the on-use trinket unless it
// won't be back before the last 15 sec (+1.6 at 300 sec, nothing lost shorter). When an Arcane
// Blast would not land before the end, Fire Blast and Arcane Explosion instead (+10.4 / +7.1 /
// +4.0 at 120 / 180 / 300 sec, the rotation auditor's pass 19).
export const APLLevel60Arcane = PresetUtils.makePresetAPLRotation('Level 60 Arcane', Level60ArcaneAPL, {
	tooltip: 'Arcane Blast stacks spent by Arcane Missiles, on Missile Barrage where it can.',
	group: 'Level 60',
});

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60Frostfire, APLLevel60Frost, APLLevel60Fire, APLLevel60Arcane, APLP1DPS],
};

export const DefaultAPL = APLLevel60Frostfire;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Each build is the talent search's best, seeded from every archetype players use and searched
// again on the final rotation over 120, 180 and 300 sec.
//
// Frostfire (0/29/22): Wake of Fire, Incineration, Improved Fireball, Ignite, Pyroblast, Improved
// Scorch, Heating Up, Master of Elements, Critical Mass and Fire Power 4 in Fire, Elemental
// Precision, Ice Shards, Piercing Ice, Frost Channeling, Ice Lance, Shatter and Fingers of Frost in
// Frost (Ice Lance for Fingers of Frost, and the rotation casts it on the charges). Frostfire
// Bolt is Fire and Frost at once, so it gets Fire Power, Ignite and Critical Mass and also Ice
// Shards, Piercing Ice and Shatter. foreverchanges.pro's popular Fire build is the same split. It
// is 1040.8 / 1011.5 / 996.4 at 120 / 180 / 300 sec (Undead, with Ice Lance at SoD's 42.9% and the
// Greater Flame scroll; the talent search on the Ice Lance rotation kept the build, 2026-10-15). Frost Channeling 3 does nothing for this build (the client
// leaves Frostfire Bolt out of it, the talent verification shows 0.0 either way): the points are
// the 20 in Frost that Fingers of Frost needs.
export const TalentsLevel60Frostfire = PresetUtils.makePresetTalents(
	'Level 60 Frostfire',
	SavedTalents.create({ talentsString: '-1355000013013304-00550003310003002' }),
	{ tooltip: 'Fire and Frost for Frostfire Bolt: Heating Up, Fingers of Frost, Shatter.', group: 'Level 60' },
);
// Frost (10/3/38): every Frost damage talent to Winter's Chill and Ice Barrier, and Arcane
// Concentration for Clearcasting. Incineration 3 is for Ice Lance, which the rotation casts on
// every Fingers of Frost charge. Four seeds (deep Frost, Frost with Arcane Meditation, Frost with
// Incineration and a 51 point Frost build) all ended within 0.2, and the talent search on the Ice
// Lance rotation (2026-10-15) kept the build. It is 948.6 / 948.3 / 940.9 (High Order, with the
// Major Spellblasting Potion and the Greater Frost scroll), 6% behind Frostfire and second of the
// four now that Ice Lance takes SoD's 42.9% (with the client's 0 it was last, at 745.4).
export const TalentsLevel60Frost = PresetUtils.makePresetTalents('Level 60 Frost', SavedTalents.create({ talentsString: '050005-03-0555300331001301251' }), {
	tooltip: "Deep Frost: Frostbolt, Fingers of Frost, Shatter, Winter's Chill.",
	group: 'Level 60',
});
// Fire (16/31/4): every Fire damage talent to Combustion, with Arcane Concentration and Missile
// Barrage (Fireball has 20% to make Arcane Missiles free and fast) and Elemental Precision 4.
// Arcane Impact 3 over Improved Channeling and Wake of Fire 2 over Master of Elements 3 are 1 and
// 3 ahead. It is 904.4 / 889.1 / 872.6 (Undead, with the Greater Flame scroll), 12%
// behind Frostfire. Fire with Frost (Elemental Precision, Ice Shards) was 740 at 180 sec on the
// old rotation, 3% behind, and a Fire mage that goes into Frost is better off as Frostfire.
export const TalentsLevel60Fire = PresetUtils.makePresetTalents(
	'Level 60 Fire',
	SavedTalents.create({ talentsString: '05020500300001-23550000130123051-004' }),
	{ tooltip: 'Deep Fire to Combustion, with Missile Barrage from Arcane.', group: 'Level 60' },
);
// Arcane (33/3/15): Arcane to Arcane Power with Arcane Blast and Missile Barrage, and Incineration
// 3 (it covers Arcane Blast). The 15 Frost points do nothing for Arcane Blast and Arcane Missiles:
// the trees have nothing else they can buy, so they sit where a Frost filler would use them. Three
// seeds (Arcane with Frost, Arcane with Fire, 47 point Arcane) ended on the same build. It is 856.0 /
// 843.5 / 829.4 (Undead, on the Arcane gear, with the Precision scroll), 17% behind Frostfire.
export const TalentsLevel60Arcane = PresetUtils.makePresetTalents('Level 60 Arcane', SavedTalents.create({ talentsString: '051225003100311531-03-02550003' }), {
	tooltip: 'Arcane to Arcane Power: Arcane Blast and Missile Barrage.',
	group: 'Level 60',
});

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60Frostfire, TalentsLevel60Frost, TalentsLevel60Fire, TalentsLevel60Arcane],
};

export const DefaultTalents = TalentsLevel60Frostfire;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// Mage Armor gives 50% casting regen under Forever (Classic 30%), 8 to 17 DPS over Ice Armor.
// Forever has no Molten Armor.
//
// Forever's mage-only staff imbue scrolls (the Comprehension skill, client 70338 items 277500 to
// 277502) go in the class imbue slot, so they stack with Wizard Oil. Each build takes its best:
// Greater Flame (+20 Fire) for Frostfire and Fire, Greater Frost (+20 Frost) for Frost, Precision
// (+1% spell crit) for Arcane. Report assumption mage_imbue_scrolls, user ruling 2026-10-15: the
// scrolls are in the game, though foreverchanges.pro lists them as hidden. At 180 sec they add
// 10.8 to 11.4 (Greater Flame, Greater Frost) and 6.2 (Precision).
export const DefaultOptions = MageOptions.create({
	armor: ArmorType.MageArmor,
	staffImbue: StaffImbue.ImbueGreaterFlame,
});
export const FrostOptions = MageOptions.create({ ...DefaultOptions, staffImbue: StaffImbue.ImbueGreaterFrost });
export const ArcaneOptions = MageOptions.create({ ...DefaultOptions, staffImbue: StaffImbue.ImbuePrecision });

// The level 60 consumables. On 2026-10-12 we tried every consumable the sim applies in each slot
// on all four builds at 180 sec. The presets model Forever's launch, with no raid open, so nothing
// that only a raid sells: Wizard Oil in place of Brilliant Wizard Oil (Zandalar Tribe, Wizard Oil
// is 12 to 16 over none). Cerebral Cortex Compound is a Blasted Lands quest. A character takes one
// of Forever's +2% crit elixirs (Mongoose, Grizzly, Owl, Sages, Whale: the client makes them role
// variants of one elixir). Elixir of Sages (25 Spirit, 2% crit) with Elixir of Greater Intellect
// is 0.3 to 1 ahead of the Owl with Greater Spirit, and Elixir of the Owl on top would add 10 to
// 21 if they stacked (open). Greater Mageblood Elixir (20 mp5) is 0 to 6, Lobster Stew
// (28 spell power) 3 to 4 over Nightfin Soup. Major Spellblasting Potion is 7 ahead of the Major
// Mana Potion for Frost and 15 behind for Frostfire, so the mana potion stays. Elixir of Greater
// Firepower is the Elixir of Holy Power under Forever (+40 Holy), so the Fire elixir is the plain
// Elixir of Firepower. Flask of Supreme Power is 75 to 90 ahead of Distilled Wisdom, and the
// Goblin Land Mine costs 2 to 11, because throwing it takes a GCD.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	frostPowerBuff: FrostPowerBuff.ElixirOfFrostPower,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.WizardOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	spiritElixir: SpiritElixir.ElixirOfSages,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The shared level 60 raid of tools/rotopt/mksettings (every class's presets use it), with the mage's
// changes: a shaman's Mana Spring Totem, and Blessing of Wisdom in place of Might. Only Arcane
// Brilliance, Divine Spirit, Gift of the Wild, Fortitude, Mana Spring and Moonkin Aura move a
// mage's damage; the rest are there so the raid matches the other classes'. Both factions have
// paladins and shamans under Forever.
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
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	sanctityAura: true,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
});

// Blessings of Kings and Wisdom.
export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
	warchiefsBlessing: false,
});

// The shared raid's debuffs. Curse of the Elements covers Arcane, Fire and Frost under Forever
// (Curse of Shadow is gone, so it is left out here; the sim ignores it under Forever), and
// Judgement of Wisdom gives mana back on spell hits too. The armor and attack power debuffs
// don't touch spells. Improved Scorch and Winter's Chill are
// the mage's own debuffs under Forever and the rotations put them up.
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
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
});

// Engineering is for the Goblin Sapper. Tailoring makes the Manaflare Regalia, the set the
// synthetic pieces stand in for. Alchemy and Enchanting changed nothing.
export const OtherDefaults = {
	distanceFromTarget: 25,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
	race: Race.RaceUndead,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor), a 3 minute fight, and
// the default buffs, debuffs and consumes.
const level60Encounter = Encounter.create({
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
			mobType: MobType.MobTypeUnknown,
			stats: new Stats().withStat(Stat.StatAttackPower, 805).withStat(Stat.StatArmor, 3731).withStat(Stat.StatHealth, 127393).asArray(),
			minBaseDamage: 3000,
			damageSpread: 0.3333,
			swingSpeed: 2,
			parryHaste: true,
		},
	],
});
export const EncounterLevel60 = PresetUtils.makePresetEncounter('Level 60', level60Encounter, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumes,
});
// Frost has the mana to spare for a Major Spellblasting Potion in place of the Major Mana Potion:
// +7.9 at 180 sec and +4.6 at 300 (the consumables auditor), and +10.5 / +7.6 with Ice Lance on
// the charges (2026-10-15). Frostfire runs dry with it at 300 sec
// (-77), Fire loses 24 / 18 and Arcane is +2.7 / -2.1, so they keep the Major Mana Potion.
export const EncounterLevel60Frost = PresetUtils.makePresetEncounter('Level 60 (Frost)', level60Encounter, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: Consumes.create({ ...DefaultConsumes, defaultPotion: Potions.MajorSpellblastingPotion }),
});

// Races, on each build with its rotation, 20000 iterations, 180 sec (120 and 300 sec in
// forever-wiki/sim-status.md). The staff has no racial, so the table is the same with weapon
// racials off. Undead (Touch of the Grave at 10% for mages) is best on Fire and Arcane. On Frostfire,
// Orc, Undead and Gnome are within 1 (one standard error), with Undead ahead at 120 sec and Gnome at
// 300, so the build keeps Undead. On Frost, Undead, High Order and Orc are within 0.7, so the build
// keeps High Order. The best Alliance race is High Order on Frost and Gnome (5% more mana) on
// Frostfire, Fire and Arcane. Windshaper can't be a mage. With Ice Lance, the staff imbue scrolls
// and the rotation auditor's pass 19 (Frostfire's Ice Lance ahead of Scorch, and instants when the
// filler would not land before the end, 2026-10-15):
//   Frostfire: Orc 1012.1, Undead 1011.5, Gnome 1011.1, High Order 1010.1, Troll 1004.7, Human 1003.7.
//   Fire:      Undead 889.1, Gnome 885.0, Orc 883.0, Human 877.1, High Order 876.8, Troll 876.1.
//   Arcane:    Undead 843.5, Orc 834.5, Gnome 832.8, Troll 829.3, Human 826.2, High Order 824.1.
//   Frost:     Undead 949.0, High Order 948.3, Orc 948.3, Gnome 947.1, Troll 946.0, Human 939.1.
const buildCommon = {
	group: 'Level 60',
	gear: GearLevel60Phase1Synthetic,
	encounter: EncounterLevel60,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2] as [Profession, Profession],
	options: DefaultOptions,
};

export const PresetBuildLevel60Frostfire = PresetUtils.makePresetBuild('Level 60 Frostfire', {
	...buildCommon,
	tooltip: 'Level 60 Undead Frostfire on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	talents: TalentsLevel60Frostfire,
	rotation: APLLevel60Frostfire,
	race: Race.RaceUndead,
});
export const PresetBuildLevel60Frost = PresetUtils.makePresetBuild('Level 60 Frost', {
	...buildCommon,
	tooltip: 'Level 60 High Order Frost on the synthetic Phase 1 gear, in a 3 minute raid fight, with a Major Spellblasting Potion.',
	encounter: EncounterLevel60Frost,
	talents: TalentsLevel60Frost,
	rotation: APLLevel60Frost,
	race: Race.RaceSkyborneHighOrder,
	options: FrostOptions,
});
export const PresetBuildLevel60Fire = PresetUtils.makePresetBuild('Level 60 Fire', {
	...buildCommon,
	tooltip: 'Level 60 Undead Fire on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	talents: TalentsLevel60Fire,
	rotation: APLLevel60Fire,
	race: Race.RaceUndead,
});
export const PresetBuildLevel60Arcane = PresetUtils.makePresetBuild('Level 60 Arcane', {
	...buildCommon,
	tooltip: 'Level 60 Undead Arcane on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	gear: GearLevel60Phase1SyntheticArcane,
	talents: TalentsLevel60Arcane,
	rotation: APLLevel60Arcane,
	race: Race.RaceUndead,
	options: ArcaneOptions,
});
