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
	Potions,
	Profession,
	Race,
	RaidBuffs,
	Spec,
	Stat,
	StrengthBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { FeralDruid_Options as FeralDruidOptions, FeralDruid_Rotation as FeralDruidRotation } from '../core/proto/druid.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import FeralAPL from './apls/feral.apl.json';
import Level60APLJSON from './apls/level60.apl.json';
import SimpleVaelAPL from './apls/simple_vael.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import P2BISGear from './gear_sets/p2.bis.gear.json';
import P2PreBISGear from './gear_sets/p2.pre-bis.gear.json';
import P3BISGear from './gear_sets/p3.bis.gear.json';
import P4BISGear from './gear_sets/p4.bis.gear.json';
import P5BISGear from './gear_sets/p5.bis.gear.json';
import P6BISGear from './gear_sets/p6.bis.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearP2PreBIS = PresetUtils.makePresetGear('P2 Pre-BiS', P2PreBISGear);
export const GearP2BIS = PresetUtils.makePresetGear('P2 BiS', P2BISGear);
export const GearP3BIS = PresetUtils.makePresetGear('P3 BiS', P3BISGear);
export const GearP4BIS = PresetUtils.makePresetGear('P4 BiS', P4BISGear);
export const GearP5BIS = PresetUtils.makePresetGear('P5 BiS', P5BISGear);
export const GearP6BIS = PresetUtils.makePresetGear('P6 BiS', P6BISGear);

// The level 60 Phase 1 set for Forever, made of synthetic items that are not in the game
// (tools/database/forever_synthetic_items_druid.go, which has the budget math), priced by
// standards 1b (Forever's physical budget per piece, 3 stat lines a piece, the two-hander at 0.90
// of the item level line with no attack power) on the shared item level table, in the standards
// 5a stat mix (the SoD level 60 BiS proportions, orchestration/gear_targets.md): armor 650.0 and
// two-hander 79.0 weighted points, 189 Agility, 177 Strength, 143 Stamina (19.6%), 9% hit (the
// special attack hit cap against a level 63 boss; moving 10 hit rating into Agility ties, into
// crit is 1.2 behind), 4.9% crit, 45 Intellect for Shifting Power's mana, 32 attack power and no
// Spirit (the cross-class rule for Energy users). Two synthetic item level 71 trinkets
// (standards 1b): 29 Agility and 17 Strength, and 165 Strength for 20 sec every 2 min (2.9 ahead
// of the same on-use as 165 Agility, 3.0 ahead of 335 attack power). Only the head and chest carry
// Grovekeeper Ferocity, Forever's crafted Feral set, so only its 2 piece (1% attack speed) is
// live, like every class's set; the 5 piece (1 sec off Shifting Power) would add 5.9 at 180 sec.
// Howling Idol (1 sec off Shifting Power, a real Forever Phase 1 idol) is worth 12.5 (those two
// and the enchant comparisons below measured on the set before the 5a rebudget). Enchants are the
// best Phase 1 ones a cat can get, nothing from a raid: Lesser Arcanum of Voracity (Agility) on
// the head and legs, Forever's +5 Agility on the neck, Agility on the back, Greater Stats, +9
// Strength on the wrist, Superior Strength on the gloves (3.2 ahead of Greater Agility), +7
// Agility on the feet and +25 Agility on the weapon. Forceful Rugged Armor Kit (+10 attack power)
// is 1.5 to 2.2 behind each of them.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1Synthetic],
	[Phase.Phase4]: [GearP2PreBIS, GearP2BIS, GearP3BIS, GearP4BIS, GearP5BIS, GearP6BIS],
};

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLFeral = PresetUtils.makePresetAPLRotation('Feral', FeralAPL);
export const APLSimpleVael = PresetUtils.makePresetAPLRotation('Simple Vaelastrasz', SimpleVaelAPL);

// The tools/rotopt druid_feral template's best (druid_presets.sh) on the synthetic set and the
// Level 60 talents: Rip at 4 combo points whenever it is down, Ferocious Bite at 5 while Rip
// has 6 sec left (1 to 3 ahead of 4), Shred, Shifting Power at 50 Energy or less but not while a Bite is due (its
// Energy goes into Shreds, 0.6 to 0.9 ahead), Berserk on cooldown, no Tiger's Fury (Forever
// removed it, sim/druid/tigers_fury.go; with it the cat did 24.9 more), no new Rip
// in the last 8 sec and Ferocious Bite at any combo points in the last 2 (0.6 to 1.4). No
// Rake: 40 Energy for about 160 damage where a Shred does 740, and Rip keeps Rend and Tear's
// bleed up on its own (keeping Rake up costs 33 at every length). Rip at 5 combo
// points is 1 to 2 behind. The search on the 1b gear kept every knob at 120 and 180 sec (at 300
// sec Shifting Power at 40 Energy is 0.3 ahead). Berserk's later uses wait for the Strength
// trinket when waiting loses no use, and Elune's Light (the Night Elf's 10% crit) stays out of
// Berserk, where every builder crits anyway: 2.5 / 0.9 / 2.7 ahead at 120 / 180 / 300 sec
// (rotation auditor's cooldown pass).
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Level 60', Level60APLJSON, {
	tooltip: "The optimizer's best level 60 rotation: Rip at 4 combo points, Shred, Shifting Power and Berserk on cooldown.",
	group: 'Level 60',
});

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60],
	[Phase.Phase4]: [APLFeral, APLSimpleVael],
};

export const DefaultAPL = APLLevel60;

export const DefaultRotation = FeralDruidRotation.create({
	maintainFaerieFire: false,
	minCombosForRip: 3,
	maxWaitTime: 2.0,
	precastTigersFury: false,
	useShredTrick: false,
});

export const SIMPLE_ROTATION_DEFAULT = PresetUtils.makePresetSimpleRotation('Simple Default', Spec.SpecFeralDruid, DefaultRotation);

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent search's best (tools/rotopt -talent-search) on the synthetic set with the Level 60
// rotation at 180 sec: Genesis 5 (Rip's ticks) and Nature's Majesty 2 (crit) from Balance, 34
// Feral with Shredding Attacks, Savage Fury, Sharpened Claws, Shifting Power and Improved
// Shifting Power, Predatory Strikes, Blood Frenzy 1, Leader of the Pack, Natural Instinct, Rend
// and Tear 5 and Berserk, and Furor 5, Naturalist 5 and Natural Shapeshifter 1 in Restoration.
// Searches started from pure Feral builds (0 / 35 / 16 with Ferocity and Reflection, 0 / 33 / 18
// with Heart of the Wild and Thick Hide) and from the Balance split, and all of them moved
// their spare points into Genesis and Nature's Majesty. On the final set the fifth point of
// Rend and Tear is 1.5 ahead of the second point of Blood Frenzy (659.3 against 657.8).
// Ferocity buys nothing once the rotation drops Rake, so its row 1 points sit in Feral
// Swiftness and Thick Hide for the same DPS.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '05002-05200332021031212051-05501' }), {
	tooltip: 'The talent search best on the synthetic Phase 1 set.',
	group: 'Level 60',
});

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60],
};

export const DefaultTalents = TalentsLevel60;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = FeralDruidOptions.create({
	latencyMs: 100,
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
// Forever class deep dive, Druid): on the cat with its weapon's damage, at 180 sec, Elemental
// Sharpening Stone is 9.5 ahead of no stone (Dense Weightstone 4.9), Thistle Tea 8.0 ahead of the
// Demonic Rune (12.2 drunk at 25 Energy or less, sim/druid/thistle_tea.go), and the Mighty Rage
// Potion's 60 Strength 6.3 ahead of the Major Mana Potion (Major Frenzy Potion 5.0), also at 300
// sec. Explosives still can't be used in form.
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredRogueThistleTea,
	defaultPotion: Potions.MightyRagePotion,
	dragonBreathChili: true,
	flask: Flask.FlaskOfDistilledWisdom,
	food: Food.FoodFlankAuPoivre,
	mainHandImbue: WeaponImbue.ElementalSharpeningStone,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
});

// The raid the Level 60 numbers were measured under (tools/rotopt/mksettings -spec feral): the
// shared level 60 raid with the cat in a melee group, so a shaman's Strength of Earth and Grace
// of Air (Windfury Totem now works in Cat Form under Forever and would add 12.5 at 180 sec, but the
// shared raid gives no melee spec one), every blessing, and its own Leader of the
// Pack. Under Forever the totems and Battle Shout are capped below their Ahn'Qiraj book ranks
// (Strength of Earth 42, Grace of Air 77, Battle Shout 115), which costs the cat 14.8 at 180 sec.
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
	sanctityAura: true,
	shadowProtection: true,
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
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
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
});

// An engineer and an alchemist like the other level 60 presets. Nothing either makes works in
// Cat Form, so the pair is free to change.
export const OtherDefaults = {
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
// The Level 60 build, 721.0 / 709.2 / 696.9 at 120 / 180 / 300 sec. Every race a druid can be
// under Forever on it with the same seed, 40,000 iterations (standard error 0.1):
//
//	                     120 sec  180 sec  300 sec
//	Night Elf             721.0    709.2    696.9   best overall and Alliance
//	Windshaper Skyborne   717.0    707.1    694.5   best Horde
//	High Order Skyborne   716.9    707.1    694.4
//	Tauren                714.0    704.1    690.9
//
// Night Elf's lead is its Agility. No racial goes by weapon type, so the weapon changes
// nothing, and the ranking with weapon type racials off is the same.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Level 60', {
	group: 'Level 60',
	tooltip: 'Level 60 Night Elf on the synthetic Phase 1 gear with the Level 60 rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceNightElf,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
