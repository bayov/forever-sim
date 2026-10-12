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
import { Mage_Options as MageOptions, Mage_Options_ArmorType as ArmorType } from '../core/proto/mage';
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

// The Level 60 presets were searched on 2026-10-12 with tools/rotopt (the mage_arcane, mage_fire,
// mage_frost and mage_frostfire templates, mage_presets.sh) against the level 60 raid in
// tools/rotopt/mksettings (spec_mage.go): the Level 60 boss (level 63, 3731 armor, Humanoid),
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
// cross-class standards every class's Level 60 preset follows. Each piece has the full epic
// budget of its slot at Forever's cheaper spell power, at one item level table for everyone (66
// for head, shoulders, chest, hands, legs and feet, 71 for the rest): 526 spell power, 137
// Intellect, 80 Spirit (10% of the budget), 10% hit and 6% crit. With Elemental Precision or
// Arcane Focus and Manaflare Regalia's 2 piece bonus (1% hit, on the head and chest only) that is
// the 16% cap against a level 63 boss. The 4 and 5 piece bonuses would add 62.9 (7%) to Frostfire
// at 180 sec and nothing to the others (the 5 piece is all Frostfire Bolt, the 4 piece only
// against Elementals).
//
// The weapon is a staff with Staff of Dominance's damage and the budget of Azuresong Mageblade
// and an off hand together. With Forever's 2H Spellblasting (+65 spell power) it is 16 to 40
// ahead of the sword and off hand, so Human's sword crit no longer counts. No trinkets: Forever
// removed the old dungeon ones and its launch trinkets aren't simulated yet. The enchants are the
// best Phase 1 ones without a raid source: Lesser Arcanum of Voracity (Intellect), +6 spell power
// on the neck, Mystic Rugged Armor Kits (+6 spell power) on chest, hands, legs and feet, Greater
// Spellpower (+16) on the wrist and Spellblasting on the staff. Arcane wears Gloves - Arcane Power
// (+20 Arcane) instead of the kit, 7.6 ahead.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
export const GearLevel60Phase1SyntheticArcane = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic (Arcane)', Level60Phase1SyntheticArcaneGear, {
	tooltip: 'The Level 60 synthetic set with Gloves - Arcane Power, for Arcane.',
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
// Pyroblast waits for 3 Heating Up stacks (a 1.5 sec cast), Fingers of Frost's Shatter crit goes
// to the next bolts, Arcane Missiles takes a Missile Barrage proc. Ice Lance takes no spell power
// (the client's coefficient is 0), so casting it on the charges costs 55 to 65. Fire Blast only
// goes with mana the fight can't use (above 150 a second of the fight left, +1.5 to +3.7); on
// cooldown its mana is worth more as Frostfire Bolts. Evocation below 10% mana with at least 30
// sec of the fight left, then rank 2 Frostfire Bolt when the mana won't last. The knobs were
// searched over 120, 180 and 300 sec together, so one rotation serves every length. The searches
// for one length at a time (on the first gear) wanted Fire Blast with Evocation at 30 to 40% and
// no time limit, which was behind this one at every length, because Evocation's 8 sec channel late
// in a short fight costs more than the mana is worth.
export const APLLevel60Frostfire = PresetUtils.makePresetAPLRotation('Level 60 Frostfire', Level60FrostfireAPL, {
	tooltip: 'Frostfire Bolt, Pyroblast on 3 Heating Up stacks, Scorch for Fire Vulnerability.',
	group: 'Level 60',
});
// Frostbolt, with Fingers of Frost's Shatter crit on the next bolts. Ice Lance takes no spell power
// (the client's coefficient is 0), so casting it on the charges costs 20. Frost rarely runs low on
// mana, so Evocation almost never goes.
export const APLLevel60Frost = PresetUtils.makePresetAPLRotation('Level 60 Frost', Level60FrostAPL, {
	tooltip: 'Frostbolt, with Shatter on Fingers of Frost.',
	group: 'Level 60',
});
// Scorch to 5 Fire Vulnerability stacks and before they run out, Pyroblast at 3 Heating Up
// stacks, Arcane Missiles on Missile Barrage, Fire Blast above 20% mana, and Fireball. Pyroblast
// at 1 or 2 stacks, Frostfire Bolt or Scorch as the filler and holding Fire Blast while Heating Up
// is full all lost.
export const APLLevel60Fire = PresetUtils.makePresetAPLRotation('Level 60 Fire', Level60FireAPL, {
	tooltip: 'Fireball, Pyroblast on 3 Heating Up stacks, Scorch for Fire Vulnerability, Fire Blast.',
	group: 'Level 60',
});
// Arcane Blast to 2 stacks, then Arcane Missiles on a Missile Barrage proc. Without one, Arcane
// Blast goes on to 4 stacks and the missiles go anyway, ending the stacks and taking their
// bonus. Below 30% mana one Arcane Blast and the missiles. That is 16 DPS ahead of three Arcane
// Blasts and the missiles whatever the proc (the template's default). No Evocation.
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
// Scorch, Heating Up, Master of Elements, Critical Mass and Fire Power 3 in Fire, Elemental
// Precision, Ice Shards, Piercing Ice, Frost Channeling, Ice Lance, Shatter and Fingers of Frost in
// Frost (Ice Lance only because Fingers of Frost needs it). Frostfire Bolt is Fire and Frost at
// once, so it gets Fire Power, Ignite and Critical Mass and also Ice Shards, Piercing Ice and
// Shatter. foreverchanges.pro's popular Fire build is the same split. It is 923.2 / 903.4 / 888.6
// at 120 / 180 / 300 sec (Undead). Master of Elements 3
// and Frost Channeling 3 are for the mana: on the first gear they cost 2.4 at 180 sec and were worth
// 30 at 300.
export const TalentsLevel60Frostfire = PresetUtils.makePresetTalents(
	'Level 60 Frostfire',
	SavedTalents.create({ talentsString: '-1355000013013304-00550003310003002' }),
	{ tooltip: 'Fire and Frost for Frostfire Bolt: Heating Up, Fingers of Frost, Shatter.', group: 'Level 60' },
);
// Frost (10/3/38): every Frost damage talent to Winter's Chill and Ice Barrier, and Arcane
// Concentration for Clearcasting. Incineration 3 was for Ice Lance, which the rotation no longer
// casts; no other place for the 3 points gains anything either (the search found no move above
// +0.0). Four seeds (deep Frost, Frost with
// Arcane Meditation, Frost with Incineration and a 51 point Frost build) all ended within 0.2.
// It is 794.8 / 793.6 / 793.4 (Undead), 12% behind Frostfire, and last of the four now that Ice
// Lance takes no spell power (with SoD's 42.9% it was second).
export const TalentsLevel60Frost = PresetUtils.makePresetTalents('Level 60 Frost', SavedTalents.create({ talentsString: '050005-03-0555300331001301251' }), {
	tooltip: "Deep Frost: Frostbolt, Fingers of Frost, Shatter, Winter's Chill.",
	group: 'Level 60',
});
// Fire (16/31/4): every Fire damage talent to Combustion, with Arcane Concentration and Missile
// Barrage (Fireball has 20% to make Arcane Missiles free and fast) and Elemental Precision 4.
// Arcane Impact 3 over Improved Channeling is 1 ahead. It is 900.1 / 884.2 / 873.7 (Undead), 2%
// behind Frostfire. Fire with Frost (Elemental Precision, Ice Shards) was 740 at 180 sec on the
// old rotation, 3% behind, and a Fire mage that goes into Frost is better off as Frostfire.
export const TalentsLevel60Fire = PresetUtils.makePresetTalents(
	'Level 60 Fire',
	SavedTalents.create({ talentsString: '05020500300001-13550000130133051-004' }),
	{ tooltip: 'Deep Fire to Combustion, with Missile Barrage from Arcane.', group: 'Level 60' },
);
// Arcane (33/3/15): Arcane to Arcane Power with Arcane Blast and Missile Barrage, and Incineration
// 3 (it covers Arcane Blast). The 15 Frost points do nothing for Arcane Blast and Arcane Missiles:
// the trees have nothing else they can buy, so they sit where a Frost filler would use them. Three
// seeds (Arcane with Frost, Arcane with Fire, 47 point Arcane) ended on the same build. It is 864.0 /
// 852.9 / 842.9 (Undead, on the Arcane gear), 5% behind Frostfire.
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
export const DefaultOptions = MageOptions.create({
	armor: ArmorType.MageArmor,
});

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

// A mage stands in a caster group: Arcane Brilliance, Divine Spirit, Gift of the Wild, Power Word:
// Fortitude, a shaman's Mana Spring Totem and Moonkin Aura. Both factions have paladins and
// shamans under Forever.
export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	divineSpirit: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
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

// Curse of the Elements covers Arcane, Fire and Frost under Forever (Curse of Shadow is gone), and
// Judgement of Wisdom gives mana back on spell hits too. Improved Scorch and Winter's Chill are
// the mage's own debuffs under Forever and the rotations put them up.
export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	judgementOfWisdom: true,
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

// Races, on each build with its rotation, 20000 iterations, 180 sec (120 and 300 sec in
// forever-wiki/sim-status.md). The staff has no racial, so the table is the same with weapon
// racials off. Undead (Touch of the Grave at 10% for mages) is best on every build and the best
// Horde race. The best Alliance race is High Order on Frostfire and Frost and Gnome (5% more mana)
// on Fire and Arcane. Windshaper can't be a mage. On Frostfire Gnome is ahead at 120 and 300 sec
// (927.0 and 890.4 against Undead's 923.2 and 888.6).
//   Frostfire: Undead 903.4, High Order 901.6, Orc 901.6, Troll 898.1, Gnome 896.5, Human 893.5.
//   Fire:      Undead 884.2, Gnome 881.8, Orc 878.9, High Order 875.9, Human 873.8, Troll 872.0.
//   Arcane:    Undead 852.9, Orc 843.7, Gnome 842.9, Troll 840.1, Human 836.9, High Order 833.9.
//   Frost:     Undead 793.6, High Order 793.2, Troll 792.2, Gnome 792.1, Orc 791.2, Human 785.3.
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
	tooltip: 'Level 60 Undead Frost on the synthetic Phase 1 gear, in a 3 minute raid fight.',
	talents: TalentsLevel60Frost,
	rotation: APLLevel60Frost,
	race: Race.RaceUndead,
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
});
