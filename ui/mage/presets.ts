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
// Frost share one set: 429 spell power, 154 Intellect, 119 Stamina, 88 Spirit, 9% hit and 103
// crit rating (7.4%). Arcane has its own set with more spell power and less crit: 460 spell
// power, 147 Intellect, 10% hit, 70 crit rating (5%), the rest about the same. With Elemental
// Precision 3 (6% since the 2026-10-20 hotfix) or Arcane Focus 5 and Manaflare Regalia's 2 piece
// bonus (1% hit, on the head and chest only) the hit is the 16% cap against a level 63 boss. The
// main set had 10% until 2026-10-23, one over the cap with the hotfix; its head moved those 10 hit
// rating to 3 Intellect and 8 spell power (+5.3 Fire, +7.3 Frostfire, +2.1 Frost at 180 sec). The 4
// and 5 piece bonuses would add 56.8 (6%) to Frostfire at 180 sec and nothing to the others (the
// 5 piece is all Frostfire Bolt, the 4 piece only against Elementals).
//
// The trinkets are synthetic too (Forever's real Phase 1 trinkets are gone or not simulated):
// 41 spell power and 11 crit rating, and an on-use +140 spell power for 20 sec every 2 min.
//
// The weapon is a staff with Staff of Dominance's damage. With Forever's 2H Spellblasting (+65
// spell power) it is 32 to 43 ahead of the sword and off hand. Since client 70425 no racial
// depends on the weapon, so the sword no longer helps a Human. The enchants are the best Phase 1 ones without a raid source:
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

// Frostfire Bolt as the filler. Scorch puts up 5 stacks of Fire Vulnerability and keeps them
// (again with under 7 sec left on them), Pyroblast waits for 3 Heating Up stacks (a 1.5 sec
// cast), Ice Lance goes on every Fingers of Frost charge, ahead of Scorch and Pyroblast (3 times
// the damage on a Frozen target since client 70425, with SoD's 42.9% spell power coefficient), and
// Arcane Missiles takes a Missile Barrage proc. Since 70425 every Fire Blast adds a Heating Up
// stack, so Fire Blast goes on cooldown above 80% mana and below that with the mana the fight
// can't use (above 70 a second of the fight left). Evocation below 20% mana with at least 45 sec
// of the fight left, then rank 2 Frostfire Bolt when the mana won't last. The on-use trinket waits
// for the 5th Fire Vulnerability stack. When a Frostfire Bolt would not land before the end (its
// cast plus 1.5 sec) Ice Lance goes instead. The knobs were searched over 120, 180 and 300 sec
// together, with the raid's hits (pushback); the rotation auditor's pass 23 moved the Scorch
// refresh, Evocation and spare-mana Fire Blast thresholds (+6.0 at 180 sec). It wands when
// nothing else can be cast (2026-10-23): +2.2 at 480 sec, nothing at 120 to 300 sec. The
// 2026-10-23 search, after Fingers of Frost's charges came to be spent a spell batch after each
// cast, put Ice Lance on the last charge only (the first goes on a Frostfire Bolt, which reads the
// target as Frozen for Shatter), Evocation below 10% with 90 sec left and rank 2 below 30 mana a
// second of the fight left: +15 / +44 / +39 at 120 / 180 / 300 sec on both seed sets.
export const APLLevel60Frostfire = PresetUtils.makePresetAPLRotation('Level 60 Frostfire', Level60FrostfireAPL, {
	tooltip: 'Frostfire Bolt, Pyroblast on 3 Heating Up stacks, Ice Lance on the last Fingers of Frost charge, Scorch for Fire Vulnerability, Fire Blast.',
	group: 'Level 60',
});
// Frostbolt, with Ice Lance on every Fingers of Frost charge (3 times the damage on a Frozen
// target and Shatter's crit), and two Arcane Blasts every 15 sec for 30% of Intellect as spell
// damage (with Arcane Mind on the 2026-10-23 build; one stack is 16 behind). A charge granted
// during a Frostbolt's cast is spent a spell batch after that Frostbolt completes, so it and the
// Ice Lance cast right after both read Frozen; holding Ice Lance for a Frostbolt every time lost
// 11 to 23 at 180 sec. Frost rarely runs low on mana, so Evocation almost never goes. When a
// Frostbolt would not land before the end (its cast plus 1 sec), Ice Lance instead.
export const APLLevel60Frost = PresetUtils.makePresetAPLRotation('Level 60 Frost', Level60FrostAPL, {
	tooltip: 'Frostbolt, with Ice Lance on Fingers of Frost and two Arcane Blasts for spell damage.',
	group: 'Level 60',
});
// Scorch to 5 Fire Vulnerability stacks and before they run out, Pyroblast at 3 Heating Up
// stacks, Arcane Missiles on Missile Barrage, Fire Blast above 10% mana (every Fire Blast adds a
// Heating Up stack since client 70425, two on a crit; 20% before the rotation auditor's pass 23),
// and Fireball. Combustion and the on-use
// trinket wait for the 5th Fire Vulnerability stack. When a Fireball would not land before the
// end (its cast plus 1.04 sec), Fire Blast and Scorch instead.
export const APLLevel60Fire = PresetUtils.makePresetAPLRotation('Level 60 Fire', Level60FireAPL, {
	tooltip: 'Fireball, Pyroblast on 3 Heating Up stacks, Scorch for Fire Vulnerability, Fire Blast.',
	group: 'Level 60',
});
// Client 70425 made Arcane Blast's stacks spell damage: 15% of Intellect each, up to 4, for 15
// sec from the first, and nothing ends them early. So Arcane builds four stacks at the start of
// every 15 sec (the fourth costs 6.25 times the first), and below 20% mana only one. Arcane
// Missiles on a Missile Barrage proc goes first, ahead of the Arcane Blasts: each Arcane Blast has
// 40% to proc it, so a proc during the ramp is used while the stacks boost it instead of being
// rolled over (+40.5 at 180 sec, the rotation auditor's pass 23). Frostbolt is the filler: it
// beats Arcane Missiles that aren't free. No Evocation. Arcane Power and Presence of Mind wait for
// the on-use trinket unless it won't be back before the last 15 sec. When a Frostbolt would not
// land before the end (its cast plus 0.75 sec), Fire Blast and Arcane Explosion instead.
export const APLLevel60Arcane = PresetUtils.makePresetAPLRotation('Level 60 Arcane', Level60ArcaneAPL, {
	tooltip: 'Four Arcane Blast stacks every 15 sec for spell damage, Arcane Missiles on Missile Barrage, Frostbolt.',
	group: 'Level 60',
});

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60Frostfire, APLLevel60Frost, APLLevel60Fire, APLLevel60Arcane, APLP1DPS],
};

export const DefaultAPL = APLLevel60Frostfire;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Each build is the talent search's best on the client 1.60.1.70425 tree, seeded from the last
// builds and searched again on the final rotation over 120, 180 and 300 sec together, with the
// raid's 2 hits a minute (pushback). 70425 moved Missile Barrage to Arcane row 1 (2 points) and
// added Arcane Wrath, so every Arcane string reads differently from the 70338 ones. Fire Power
// is 3 points and Critical Mass 5, and Winter's Chill only helps Cone of Cold now.
//
// The 2026-10-20 hotfix to 70450 took Piercing Ice out of the Frost tree, made Elemental Precision
// 3 points of 2%, and swapped Frostbite with Permafrost and Frost Channeling with Improved Blizzard
// (the Frost strings read differently from here on). The freed points were placed by sims at 180
// sec, 10,000 to 40,000 iterations; the other points are as the search left them.
//
// Frostfire (0/29/22): Incineration, Improved Fireball, Ignite, Burning Soul 3 (pushback),
// Pyroblast, Improved Scorch, Heating Up, Master of Elements, Critical Mass 5 and Fire Power 3 in
// Fire, Elemental Precision, Ice Shards, Frost Channeling, Ice Lance, Shatter and Fingers of Frost
// in Frost. Frostfire Bolt is Fire and Frost at once, so it gets Fire Power, Ignite and Critical
// Mass and also Ice Shards and Shatter. A second seed with Arcane Blast (10 Arcane points, without
// Fingers of Frost) ended 65 behind on a Combustion build. It is 1047.6 / 1041.5 / 997.8 at 120 /
// 180 / 300 sec (Human; 1081.9 at 180 before the hotfix). Frost Channeling 3 does nothing for this
// build (the client leaves Frostfire Bolt out of it): the points are the 20 in Frost that Fingers
// of Frost needs, and since the hotfix so are Improved Frostbolt 2, Improved Blizzard 1 and Arctic
// Reach 2, the five Piercing Ice and Elemental Precision gave back. Leaving Fingers of Frost for 12
// Arcane points (Arcane Subtlety, Arcane Concentration) is 1009.3.
export const TalentsLevel60Frostfire = PresetUtils.makePresetTalents(
	'Level 60 Frostfire',
	SavedTalents.create({ talentsString: '-0155003013012503-0235000311203002' }),
	{ tooltip: 'Fire and Frost for Frostfire Bolt: Heating Up, Fingers of Frost, Shatter.', group: 'Level 60' },
);
// Frost (17/3/31): every Frost damage talent to Ice Barrier but Winter's Chill (a Cone of Cold
// debuff since 70425), with Arcane Concentration, Missile Barrage 2, Arcane Impact and Arcane
// Blast from Arcane: one Arcane Blast every 15 sec is 15% of Intellect as spell damage. Incineration
// 3 is for Ice Lance. The old Winter's Chill points went to Frost Channeling, Arcane Meditation and
// Permafrost (a point the 30 for Ice Barrier needs). Ice Barrier gives 3% Frost damage for a minute
// and 3% more while the shield holds since the hotfix, and the five points Piercing Ice and
// Elemental Precision gave back are Frost Warding 2 (a fifth more absorb) and Improved Blizzard 3,
// which do nothing for the damage but are what the 30 for Ice Barrier needs. It is 823.8 / 820.8 /
// 814.3, 21% behind Frostfire (High Order, with the Major Spellblasting Potion and the Greater Frost
// scroll; 843.9 at 180 before the hotfix). Stopping at Fingers of Frost (22 Frost) for Arcane
// Subtlety, Arcane Mind and Arcane Instability is 885.1 at 180 sec, and Arcane Subtlety in place
// of two Incineration points 839.2. 70425 cost it Winter's Chill's crit and Ice Lance's fourth
// times on a Frozen target.
export const TalentsLevel60Frost = PresetUtils.makePresetTalents(
	'Level 60 Frost',
	SavedTalents.create({ talentsString: '05020050031001-03-253500133101301201' }),
	{
		tooltip: 'Deep Frost: Frostbolt, Fingers of Frost, Shatter, with Arcane Blast for spell damage.',
		group: 'Level 60',
	},
);
// Fire (13/35/3): every Fire damage talent to Combustion with Critical Mass 5, Burning Soul 3
// (pushback) and Master of Elements 3, with Arcane Concentration and Missile Barrage 2 (Fireball
// has 20% to make Arcane Missiles free and fast) and Elemental Precision 3. The two points the
// hotfix took off Elemental Precision are Arcane Subtlety 2 (30 less resistance on the boss), worth
// more than Improved Flamestrike and Blast Wave (1067.2) or Improved Channeling 2 (1067.1). It is
// 1100.8 / 1094.3 / 1080.9 (Human, with the Greater Flame scroll; 1067.6 at 180 before the
// hotfix). Every Fire Blast adds a Heating Up stack since 70425, which puts Fire ahead of Frostfire.
export const TalentsLevel60Fire = PresetUtils.makePresetTalents('Level 60 Fire', SavedTalents.create({ talentsString: '0402205-23550030130135031-003' }), {
	tooltip: 'Deep Fire to Combustion, with Missile Barrage from Arcane.',
	group: 'Level 60',
});
// Arcane (36/2/13): Arcane to Arcane Power with Arcane Blast, Missile Barrage 2 and Improved
// Channeling 5 (pushback on Arcane Blast), Arcane Subtlety 2, Incineration 2 (it covers Arcane
// Blast) and Frost for the Frostbolt filler (Improved Frostbolt, Elemental Precision, Ice Shards).
// Arcane Wrath lost to Missile Barrage's second point: a free Arcane Missiles at twice the speed is
// worth more than the 3.5 sec cast. The five points Piercing Ice and Elemental Precision gave back
// are Arcane Subtlety 2 and Improved Channeling 3 (982.0 at 180 sec, 40,000 iterations), ahead of
// Arcane Subtlety with Frost Channeling 3 (977.3) and of Frost Channeling and Ice Lance (951.1). It
// is 989.5 / 981.9 / 971.4, 6% behind Frostfire (Human, on the Arcane gear, with the Precision
// scroll; 964.7 at 180 before the hotfix).
export const TalentsLevel60Arcane = PresetUtils.makePresetTalents(
	'Level 60 Arcane',
	SavedTalents.create({ talentsString: '0552205003100301531-02-0535' }),
	{
		tooltip: 'Arcane to Arcane Power: four Arcane Blast stacks for spell damage, Missile Barrage, Frostbolt.',
		group: 'Level 60',
	},
);

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
	// The fight is in a natural flask zone (standards 11), but the best natural flask is 61 to 80 behind
	// the Flask of Supreme Power at 180 sec (Aggression on Arcane, Fire and Frostfire, Swiftness on
	// Frost).
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
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
	salvation: true,
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
	race: Race.RaceHuman,
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

// Races, on each build with its rotation, 20000 iterations, 180 sec, with the 70425 racials (no
// racial depends on the weapon any more), before the rotation auditor's pass 23. Human (Ruthless, 2% crit) is best on Frostfire, Fire and
// Arcane, with Orc (Savage, 1% crit, and Blood Fury) 0.3 behind on Arcane, within the noise. On
// Frost, High Order's 3% haste is ahead. The best Horde race is Orc on Arcane, Fire and Frost and
// Orc on Frostfire too (High Order, 3.7 behind Human there, is Alliance). Windshaper can't be a mage.
//   Frostfire: Human 1010.8, High Order 1007.1, Orc 1004.8, Gnome 996.2, Undead 995.5, Troll 989.1.
//   Fire:      Human 993.2, Orc 989.0, Gnome 985.8, Undead 982.7, Troll 974.9, High Order 970.9.
//   Arcane:    Human 884.9, Orc 884.6, Gnome 880.5, High Order 877.6, Undead 877.6, Troll 872.7.
//   Frost:     High Order 827.3, Human 820.2, Orc 819.8, Gnome 817.0, Undead 816.9, Troll 814.2.
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
	tooltip: 'Level 60 Human Frostfire on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	talents: TalentsLevel60Frostfire,
	rotation: APLLevel60Frostfire,
	race: Race.RaceHuman,
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
	tooltip: 'Level 60 Human Fire on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	talents: TalentsLevel60Fire,
	rotation: APLLevel60Fire,
	race: Race.RaceHuman,
});
export const PresetBuildLevel60Arcane = PresetUtils.makePresetBuild('Level 60 Arcane', {
	...buildCommon,
	tooltip: 'Level 60 Human Arcane on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	gear: GearLevel60Phase1SyntheticArcane,
	talents: TalentsLevel60Arcane,
	rotation: APLLevel60Arcane,
	race: Race.RaceHuman,
	options: ArcaneOptions,
});
