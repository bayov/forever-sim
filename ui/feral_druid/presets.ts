import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
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
	Stat,
	StrengthBuff,
	SunderRamp,
	TotemWeaponBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { FeralDruid_Options as FeralDruidOptions, FeralDruid_Rotation as FeralDruidRotation } from '../core/proto/druid.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import Level60APLJSON from './apls/level60.apl.json';
import Level60FaerieFireAPLJSON from './apls/level60_ff.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

// The level 60 Phase 1 set for Forever, made of synthetic items that are not in the game
// (tools/database/forever_synthetic_items_druid.go, which has the budget math), priced by
// standards 1b (Forever's physical budget per piece, 3 stat lines a piece, the two-hander at 0.90
// of the item level line with no attack power) on the shared item level table, in the standards
// 5a stat mix (the SoD level 60 BiS proportions, orchestration/gear_targets.md): armor 650.0 and
// two-hander 79.0 weighted points, 194 Agility, 182 Strength, 143 Stamina (19.6%), 8% hit (the
// hit cap against a level 63 boss since 2026-10-23, no hit suppression; the 10 hit rating the set
// carried for the old 9% cap went into 5 Strength and 5 Agility on the legs, 5.5 ahead at 180 sec,
// where all 10 as crit is 1.4 behind), 4.9% crit, 45 Intellect for Shifting Power's mana, 32 attack power and no
// Spirit (the cross-class rule for Energy users). Two synthetic item level 71 trinkets
// (standards 1b): 29 Agility and 17 Strength, and 165 Strength for 20 sec every 2 min (2.9 ahead
// of the same on-use as 165 Agility, 3.0 ahead of 335 attack power). Only the head and chest carry
// Grovekeeper Ferocity, Forever's crafted Feral set, so only its 2 piece (1% attack speed) is
// live, like every class's set; the 5 piece (1 sec off Shifting Power) would add 5.9 at 180 sec.
// Idol of the Dream (Rip 2 sec longer, a seventh tick; Sunken Temple or crafted, a real Forever
// Phase 1 idol) is 3.0 ahead of Howling Idol (0.5 sec off Shifting Power) at 180 sec on both seed
// sets (the 5 piece and the enchant comparisons below measured on the set before the 5a rebudget). Enchants are the
// best Phase 1 ones a cat can get, nothing from a raid: Lesser Arcanum of Voracity (Agility) on
// the head and legs, Forever's +5 Agility on the neck, Agility on the back, Greater Stats, +9
// Strength on the wrist, Superior Strength on the gloves (3.2 ahead of Greater Agility), +7
// Agility on the feet and +25 Agility on the weapon. Forceful Rugged Armor Kit (+10 attack power)
// is 1.5 to 2.2 behind each of them.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Feral', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

export const GearPresets = [GearLevel60Phase1Synthetic];

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The tools/rotopt druid_feral template's best (druid_presets.sh) on the synthetic set and the
// Level 60 talents, re-searched on 2026-10-23's sim (Rake's attack power, Rend and Tear on its
// ticks, Omen at the swing speed, the live Rip): Rip at 5 combo points whenever it is down (2.0 to
// 5.6 ahead of 4 with Idol of the Dream), Ferocious Bite at 5 while Rip has 6 sec left, Shadowmeld
// into Ravage on Clearcasting or in the last 10 sec (the Night Elf; ahead of Rake, 5 ahead of after
// it), Rake whenever it is down with more than 6 sec left, Shred, Shifting Power at 50 Energy or
// less but not while a Bite is due (its Energy goes into Shreds), Berserk held for the Strength
// trinket when waiting loses no use, Elune's Light outside Berserk, no Tiger's Fury (Forever
// removed it, sim/druid/tigers_fury.go), no new Rip in the last 8 sec and Ferocious Bite at any
// combo points in the last 2. Keeping Rake up is now 38 to 42 ahead of no Rake at 120 / 180 / 300
// sec on both seed sets (836.8 / 823.8 / 813.3 to 875.5 / 864.7 / 855.5, seed 1). The knob search
// kept the rest at 120, 180 and 300 sec.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Feral', Level60APLJSON, {
	tooltip: "The optimizer's best level 60 rotation: Rip at 5 combo points, Rake kept up, Shred, Shadowmeld into Ravage, Shifting Power and Berserk.",
	group: 'Synthetic',
});

// Faerie Fire duty (standards 23): the cat that keeps Faerie Fire on the boss for the raid. Forever
// has no Faerie Fire (Feral) (16857, 17390 to 17392 have no client rows): in Cat Form the druid casts
// the caster Faerie Fire, free there, on a 1.5 sec GCD with no cooldown. The Level 60 rotation leaves
// it to the raid's debuff; this one casts it first after the cooldowns and again once it has under 7
// sec left (2, 3, 5 and 7 sec all keep it up 99.9% of the fight, within 1 DPS). Forever's Faerie Fire
// and Curse of Recklessness share the minor armor slot (standards 20), so the line waits while a curse
// holds it, and the duty's cost shows with both off: 10.2 / 8.3 / 4.3 DPS at 120 / 180 / 300 sec
// (869.7 / 858.5 / 846.9, Night Elf, 20,000 iterations, seed 1, no bomb; Faerie Fire no longer resets
// the swing), the GCDs and the moment before it lands, since the cat's own hits get the armor.
export const APLLevel60FaerieFire = PresetUtils.makePresetAPLRotation('Feral · Faerie Fire duty', Level60FaerieFireAPLJSON, {
	tooltip:
		"The Feral rotation, keeping Faerie Fire on the target for the raid (free in Cat Form). Pick the duty build to switch off the raid's Faerie Fire and Curse of Recklessness.",
	group: 'Synthetic',
});

export const APLPresets = [APLLevel60, APLLevel60FaerieFire];

export const DefaultAPL = APLLevel60;

export const DefaultRotation = FeralDruidRotation.create({
	maintainFaerieFire: false,
	minCombosForRip: 3,
	maxWaitTime: 2.0,
	precastTigersFury: false,
	useShredTrick: false,
});

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent search's best (tools/rotopt -talent-search) on the synthetic set with the Level 60
// rotation at 180 sec: Genesis 5 (Rip's ticks) and Nature's Majesty 2 (crit) from Balance, 34
// Feral with Ferocity 5, Feral Swiftness 2, Shredding Attacks, Savage Fury, Sharpened Claws,
// Shifting Power and Improved Shifting Power, Predatory Strikes, Leader of the Pack, Natural
// Instinct, Rend and Tear 5 and Berserk, and Furor 5 and Naturalist 5 in Restoration. With Rake
// back in the rotation (2026-10-23), the search moved Blood Frenzy 1, Thick Hide 3 and a point of
// Feral Swiftness into Ferocity 5 (Rake and Shred 5 Energy cheaper): 852.7 to 864.4 at 180 sec.
// Feral Swiftness 2 (movement speed) is a must-take (standards 26), so its second point comes back
// from Natural Shapeshifter, and the search pins it (-require feralSwiftness) and keeps the rest:
// 881.1 / 868.1 / 852.5 at 120 / 180 / 300 sec against 881.1 / 868.2 / 857.7 (Night Elf,
// 20,000 iterations, seed 1), the long fight losing Natural Shapeshifter's cheaper shifts. Earlier searches started from pure Feral builds (0 / 35 / 16 with Ferocity
// and Reflection, 0 / 33 / 18 with Heart of the Wild and Thick Hide) and from the Balance split, and
// all of them moved their spare points into Genesis and Nature's Majesty.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Feral', SavedTalents.create({ talentsString: '05002-55200032021030212051-055' }), {
	tooltip: 'The talent search best on the synthetic Phase 1 set.',
	group: 'Synthetic',
});

export const TalentPresets = [TalentsLevel60];

export const DefaultTalents = TalentsLevel60;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// Clearcasting fished for before the pull (standards 13 allows pre-pull casts; wowsims/forever
// b324dba843), so the fight opens with it and Omen's cooldown running: 4.6 ahead for the Night Elf
// (its Shadowmeld into Ravage spends it) and 1.2 for the Tauren at 180 sec, before the Rake change.
export const DefaultOptions = FeralDruidOptions.create({
	latencyMs: 100,
	prepullClearcasting: true,
});

// The level 60 consumables, Forever launch Phase 1 only (nothing from Molten Core, Zul'Gurub,
// Blackwing Lair, Ahn'Qiraj or Naxxramas). The five +2% crit elixirs (Mongoose, Grizzly, Owl,
// Sages, Whale) are one elixir in the client, so the cat takes one: Elixir of the Mongoose,
// with Forever's Elixir of Ferocity (18 Strength and 18 Agility, no crit) in the Strength slot,
// 2.6 ahead of Juju Power and 4.7 ahead of the Grizzly with Greater Agility at 180 sec. If two
// crit elixirs stacked, the Grizzly in place of Ferocity would add 4.9. Other elixirs in
// different slots stack in the sim (notes.md). We tried every consumable the sim applies in
// each slot, first on an early Level 60 build at 180 sec (608 DPS): taking each one out costs
// Elixir of the Mongoose 19.7 (Greater Agility is 9.5 behind), Dragonbreath Chili 9.8 and Juju
// Might 6.1, and Ground Scorpok Assay is 1.3 ahead of R.O.I.D.S. Flank au Poivre is 4 ahead of
// Grilled Squid. Flask of Distilled Wisdom keeps Shifting Power in mana; Flask of Supreme Power
// is 6 ahead only through the chili's fire damage. Forever lets a druid use potions, Thistle Tea
// and Rage potions in form, and makes weapon imbues work on a cat (Blizzard's World of Warcraft:
// Forever class deep dive, Druid): on the cat with its weapon's damage, at 180 sec, Dense
// Weightstone (the mace's stone) is 4.9 ahead of no stone (Elemental Sharpening Stone would be
// 9.5, but its plans only drop in Molten Core), Thistle Tea 8.0 ahead of the
// Demonic Rune (12.2 drunk at 25 Energy or less, sim/druid/thistle_tea.go), and the Mighty Rage
// Potion's 60 Strength 6.3 ahead of the Major Mana Potion (Major Frenzy Potion 5.0), also at 300
// sec. The Engineering explosives can't be used in form, but EZ-Thro Dark Bomb can: 4.1 / 3.2 / 1.9
// ahead of none at 120 / 180 / 300 sec (standards 11 takes an explosive only when it gains; the Faerie
// Fire duty build goes without, below). The natural flask (standards 11, the user's decision 2026-10-23: the fight is in Hyjal
// Summit or the Barrow Deeps) is Aggression, +23.6 over the Flask of Distilled Wisdom (Precision
// +22.2, Swiftness +13.8, Accuracy +0.2 with the hit capped). Rechecked on the 2026-10-23 rotation:
// Precision 2.9 and Swiftness 10.7 behind Aggression at 180 sec.
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredRogueThistleTea,
	defaultPotion: Potions.MightyRagePotion,
	dragonBreathChili: true,
	fillerExplosive: Explosive.ExplosiveEZThroDarkBomb,
	flask: Flask.FlaskOfNaturalAggression,
	food: Food.FoodFlankAuPoivre,
	mainHandImbue: WeaponImbue.DenseWeightstone,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
});

// The Faerie Fire duty build throws no bomb: its GCDs lose to the Faerie Fire casts, none is 1.6 /
// 0.6 / 0.0 ahead at 120 / 180 / 300 sec (seed 1; 1.9 / 1.2 / 0.4 on seed 1000003).
export const DefaultConsumesFaerieFireDuty = Consumes.create({ ...DefaultConsumes, fillerExplosive: Explosive.ExplosiveUnknown });

// The raid the Level 60 numbers were measured under (tools/rotopt/mksettings -spec feral): the
// shared level 60 raid with the cat in a melee group, so a shaman's Strength of Earth and
// Windfury Totem, as the warrior and rogue presets have (it works on a cat under Forever), and
// Grace of Air from a second shaman in the raid as well (the user's ruling 2026-10-16, "both totems": client 1.60.1.70425 lets Grace of Air stack with Windfury Totem and makes both raid-wide), every
// blessing, and its own Leader of the Pack. Under Forever the totems and Battle Shout are capped
// below their Ahn'Qiraj book ranks (Strength of Earth 42, Battle Shout 115).
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
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
	thorns: TristateEffect.TristateEffectImproved,
	totemWeaponBuff: TotemWeaponBuff.TotemWeaponBuffWindfury,
	trueshotAura: true,
	salvation: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	fengusFerocity: false,
	rallyingCryOfTheDragonslayer: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
	warchiefsBlessing: false,
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
	// The raid's Sunder Armor builds up from the pull (standard 32): a warrior tank stacks it and a DPS warrior casts one on its first GCD.
	sunderRamp: SunderRamp.create({ tank: true, pullSunders: 1 }),
	thunderClap: TristateEffect.TristateEffectImproved,
});

// An engineer and an alchemist like the other level 60 presets. Nothing either makes works in
// Cat Form, so the pair is free to change.
export const OtherDefaults = {
	channelClipDelay: 50,
	distanceFromTarget: 5,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceNightElf,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the shared Level 60 boss (level 63, 3731 armor, type Unknown, standards 19), a 3 minute
// fight, and the buffs, debuffs and consumes above.
export const EncounterLevel60 = PresetUtils.makePresetEncounter(
	'Generic boss',
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
// The duty presets' raid (standards 23): the Level 60 raid without its Faerie Fire, and without its
// Curse of Recklessness, which shares Faerie Fire's minor armor slot under Forever (standards 20),
// so the armor comes only from our own Faerie Fire. The base builds set the full raid back.
export const EncounterLevel60FaerieFireDuty = PresetUtils.makePresetEncounter('Faerie Fire duty', EncounterLevel60.encounter, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: Debuffs.create({ ...DefaultDebuffs, faerieFire: false, curseOfRecklessness: false }),
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumesFaerieFireDuty,
});
// The Level 60 build, Night Elf: 881.1 / 868.1 / 852.5 at 120 / 180 / 300 sec (seed 1, 20,000
// iterations; 881.0 / 867.9 / 852.6 on seed 1000003) with Feral Swiftness 2 (2026-10-24); the duty
// build 875.6 / 863.6 / 850.8 (875.6 / 863.7 / 850.9). Every race a druid can be under Forever on the
// build before that (Natural Shapeshifter 1 for the second Feral Swiftness point), seed 1 (seed
// 1000003 within 0.3):
//
//	                     120 sec  180 sec  300 sec
//	Night Elf             881.1    868.2    857.7   best overall and Alliance
//	Tauren                872.9    862.9    854.7   best Horde
//	Windshaper Skyborne   866.0    856.5    848.5
//	High Order Skyborne   865.9    856.5    848.5
//
// The Night Elf's Shadowmeld into Ravage (on Clearcasting, which the pre-pull fishing and Rake's
// extra Omen rolls make more of) and Elune's Light put it 5.2 ahead of the Tauren at 180 sec on both
// seed sets. No racial goes by weapon type.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Feral', {
	group: 'Synthetic',
	tooltip: 'Level 60 Night Elf on the synthetic Phase 1 gear with the Feral rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceNightElf,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});
export const PresetBuildLevel60FaerieFire = PresetUtils.makePresetBuild('Feral · Faerie Fire duty', {
	group: 'Synthetic',
	tooltip:
		"The Feral build keeping Faerie Fire on the target for the raid (free in Cat Form). It turns off the raid's Faerie Fire and Curse of Recklessness, which share its armor slot.",
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60FaerieFire,
	encounter: EncounterLevel60FaerieFireDuty,
	race: Race.RaceNightElf,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});
