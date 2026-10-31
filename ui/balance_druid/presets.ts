import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	Flask,
	Food,
	IndividualBuffs,
	IntellectElixir,
	ManaRegenElixir,
	MobType,
	NaturePowerBuff,
	PartyBuffs,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	SpellPowerBuff,
	SpiritElixir,
	Stat,
	TristateEffect,
	UnitReference,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { BalanceDruid_Options as BalanceDruidOptions } from '../core/proto/druid.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import Balance from './apls/balance.apl.json';
import Level60APLJSON from './apls/level60.apl.json';
import Level60FaerieFireAPLJSON from './apls/level60_ff.apl.json';
import Level60WrathAPLJSON from './apls/level60_wrath.apl.json';
import Level60WrathFaerieFireAPLJSON from './apls/level60_wrath_ff.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import Level60Phase1SyntheticWrathGearJSON from './gear_sets/level60_phase1_synthetic_wrath.gear.json';
import P0BISGear from './gear_sets/p0.bis.gear.json';
import P1BISGear from './gear_sets/p1.bis.gear.json';
import P2BISGear from './gear_sets/p2.bis.gear.json';
import P3BISGear from './gear_sets/p3.bis.gear.json';
import P4BISGear from './gear_sets/p4.bis.gear.json';
import P5BISGear from './gear_sets/p5.bis.gear.json';
// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearP0BIS = PresetUtils.makePresetGear('Pre-BiS', P0BISGear);
export const GearP1BIS = PresetUtils.makePresetGear('P1 BiS', P1BISGear);
export const GearP2BIS = PresetUtils.makePresetGear('P2 BiS', P2BISGear);
export const GearP3BIS = PresetUtils.makePresetGear('P3 BiS', P3BISGear);
export const GearP4BIS = PresetUtils.makePresetGear('P4 BiS', P4BISGear);
export const GearP5BIS = PresetUtils.makePresetGear('P5 BiS', P5BISGear);

// The level 60 Phase 1 set for Forever, made of synthetic items that are not in the game
// (tools/database/forever_synthetic_items_druid.go, which has the budget math), priced by
// standards 1b (Forever's caster budget per piece, 4 stat lines a piece, the staff at Forever's
// caster weapon rate) on the shared item level table, in the standards 5a stat mix (the SoD level
// 60 BiS proportions, orchestration/gear_targets.md): armor 709.5 and staff 193.0 weighted points,
// 375 spell power, 147 Intellect, 136 Stamina (15%), 89 Spirit (10%), 7% crit and 11% hit, which
// with Nature's Reach and the 2 piece is the 16% spell hit cap against a level 63 boss. Hit to the
// cap still sims best: a set 50 hit rating under it, the rest in the same proportions, is 8.7
// behind. Two synthetic item level 71 trinkets (standards 1b): 36 spell power and 15 crit rating,
// and 140 spell power for 20 sec every 2 min, used on the pull. Only the head and chest carry
// Grovekeeper Eclipse, Forever's crafted Balance set, so only its 2 piece (1% hit) is live, like
// every class's set; the 5 piece (3 sec more Insect Swarm) would add 12.6 at 180 sec. Swarming Idol
// (Insect Swarm 2 sec longer, a real Forever Phase 1 idol) is worth 3.3 (those two and the
// enchant comparisons below measured on the set before the 5a rebudget). Enchants are the best
// Phase 1 ones a moonkin can get, nothing from a raid: Lesser Arcanum of Voracity on the head,
// Forever's +6 spell power on the neck, Mystic Rugged Armor Kit on the chest, legs and feet,
// Greater Spellpower on the wrist, Arcane Power on the gloves (4.7 ahead of Natural Power,
// Starfire is Arcane) and Spellblasting on the staff (6.9 ahead of Mighty Spell Power).
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Balance', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

// The same set for the Wrath build, with Natural Power (+20 Nature damage) on the gloves in place
// of Arcane Power, 8.1 ahead for Wrath (762.2 against 754.1 at 180 sec, with Starfire rank 7).
export const GearLevel60Phase1SyntheticWrath = PresetUtils.makePresetGear('Balance Wrath', Level60Phase1SyntheticWrathGearJSON, {
	tooltip: 'The synthetic Phase 1 set with Natural Power on the gloves, for the Wrath build.',
	group: 'Synthetic',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1Synthetic, GearLevel60Phase1SyntheticWrath],
	[Phase.Phase5]: [GearP0BIS, GearP1BIS, GearP2BIS, GearP3BIS, GearP4BIS, GearP5BIS],
};

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The tools/rotopt druid_balance template's best (druid_presets.sh) on the synthetic set and
// the Level 60 talents: a Starfire cast 3 sec before the pull (it lands at 0, worth 9.9),
// Insect Swarm and Moonfire whenever they are down, a free Starfire on Clearcasting (Wrath does
// not use it up), then Wrath when Eclipse is down to 2 charges or fewer and no Wrath is in
// flight, Starfire otherwise. That spends every Eclipse charge (each makes a Starfire 0.5 sec
// shorter) without throwing Wraths at a full stack. The knob search kept it at 120, 180 and 300
// sec (827.0 at 180 sec, 40,000 iterations). On the set before the 5a rebudget, dropping
// Moonfire cost 56 DPS at 180 sec, Insect Swarm 62, and casting only Starfire 74. Innervate is
// left to its own trigger (at 500 mana), 2 ahead of casting it at 40% mana. The synthetic spell
// power trinket goes 0.1 sec before the pull, so it is up when the Starfire lands: 1.8 / 1.1 /
// 0.7 ahead of using it at 0 (rotation auditor). Major Spellblasting Potion and Elune's Light
// (the Night Elf's racial) went at the same moment, and Elune's later uses waited for the trinket
// when waiting loses no use: 2.2 / 1.5 / 1.6 ahead at 120 / 180 / 300 sec (rotation auditor's
// cooldown pass; the Wrath build gains 1.2 / 0.8 / 1.0 from the same lines). In the last seconds,
// once a Starfire would no longer land before the end, it casts Moonfire instead: 4.4 / 2.9 / 1.4
// ahead (rotation auditor), 835.9 / 831.6 / 820.7 with Wrath rank 8 at x1.085 (Night Elf, before
// the raid's damage hits and Nature's Focus).
// The presets were Tauren from client 1.60.1.70425's racials to 2026-10-27, with the Elune's Light
// lines out. The Level 60 preset is Night Elf again since the race-tuning round (2026-10-27, standards
// 29 and 16), and its rotation has them back (elune=1): +1.1 at 180 sec on seed 1 (+1.0 on seed
// 1000003) over leaving Elune's Light to the autocast line, 0.0 for a Tauren. The Wrath build is
// Windshaper Skyborne and leaves them out.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Balance', Level60APLJSON, {
	tooltip: "The optimizer's best level 60 rotation: both dots up, then Wrath and Starfire in turn for Eclipse.",
	group: 'Synthetic',
});
// export const APLP1Balance = PresetUtils.makePresetAPLRotation('Balance', P1APL);
export const DefaultBalance = PresetUtils.makePresetAPLRotation('Default Balance', Balance);
// Wrath as the only filler (1.5 sec with Improved Wrath, half the mana) after a Wrath 1.5 sec
// before the pull, for a moonkin that would rather not weave: 774.7 / 776.5 / 773.7 at 120 /
// 180 / 300 sec, 6.1% behind the Level 60 rotation, on its own talents (TalentsLevel60Wrath)
// and gear (GearLevel60Phase1SyntheticWrath). Moonfire once a Wrath plus 1.5 sec would no longer
// land before the end is 5.9 / 3.8 / 2.2 ahead (rotation auditor), 787.0 / 786.6 / 782.3 with Wrath
// rank 8 at x1.085.
export const APLLevel60Wrath = PresetUtils.makePresetAPLRotation('Balance Wrath', Level60WrathAPLJSON, {
	tooltip: 'Both dots up, a free Starfire on Clearcasting, and Wrath otherwise.',
	group: 'Synthetic',
});
// Faerie Fire duty (standards 23): the moonkin that keeps Faerie Fire on the boss for the raid.
// The Level 60 rotations leave it to the raid's debuff; these cast it first after the pull and
// again once it has under 1.5 sec left. Forever's Faerie Fire and Curse of Recklessness share the
// minor armor slot (standards 20), so the line waits while a curse holds it, and the duty's cost
// shows with the raid's Faerie Fire and Curse of Recklessness both off. The moonkin gets nothing
// from the armor itself, so the duty costs its GCDs: 32.4 / 30.7 / 28.7 DPS at 120 / 180 / 300
// sec for Level 60 (797.6 / 797.5 / 789.9), 30.5 / 28.6 / 27.1 for Level 60 Wrath (752.1 /
// 755.8 / 753.4), Tauren, 20,000 iterations. With Subtlety (2026-10-24): 33.5 / 31.5 / 29.1 (824.4 /
// 824.8 / 817.8) and 31.7 / 29.3 / 27.6 (779.9 / 784.0 / 781.4), seed 1. The knob search kept the rest of both rotations.
export const APLLevel60FaerieFire = PresetUtils.makePresetAPLRotation('Balance · Faerie Fire duty', Level60FaerieFireAPLJSON, {
	tooltip:
		"The Balance rotation, keeping Faerie Fire on the target for the raid. Pick the duty build to switch off the raid's Faerie Fire and Curse of Recklessness.",
	group: 'Synthetic',
});
export const APLLevel60WrathFaerieFire = PresetUtils.makePresetAPLRotation('Balance Wrath · Faerie Fire duty', Level60WrathFaerieFireAPLJSON, {
	tooltip:
		"The Balance Wrath rotation, keeping Faerie Fire on the target for the raid. Pick the duty build to switch off the raid's Faerie Fire and Curse of Recklessness.",
	group: 'Synthetic',
});
export const APLPresets = {
	[Phase.Phase1]: [APLLevel60, APLLevel60FaerieFire, APLLevel60Wrath, APLLevel60WrathFaerieFire, DefaultBalance],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent search's best (tools/rotopt -talent-search) on the synthetic set with the Level 60
// rotation at 180 sec: 38 Balance (everything that does damage, Genesis 2 and Moonglow 3), and
// Nature's Focus 5 (pushback from the raid's hits), Naturalist 5 (1% damage a point) and Subtlety 3
// from Restoration. Subtlety 3 (Forever's version cuts Nature and Arcane threat) is a must-take
// (standards 26), so the search pins it (-require subtlety) from 5532220115501351--5023, and it
// moved Genesis 3 into Naturalist: 857.9 / 856.3 / 846.9 at 120 / 180 / 300 sec against 864.0 /
// 862.4 / 853.1 for the old 41 Balance build (5532220115501351--505), 6.1 / 6.1 / 6.2 for the
// threat cut (Tauren, 20,000 iterations, seed 1; the same on seed 1000003). Before that, a search
// on the final set only traded Genesis, Moonglow and Reflection points within the noise. The other
// archetypes were well behind on the first set at 180 sec: 31 / 0 / 20 (Reflection and Tranquil
// Spirit, no Moonfury) 10% and 41 / 10 / 0 4%.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Balance', SavedTalents.create({ talentsString: '5232220115501351--5053' }), {
	tooltip: "The talent search best on the synthetic Phase 1 set with Subtlety: 38 Balance, Nature's Focus, Naturalist and Subtlety from Restoration.",
	group: 'Synthetic',
});
// The Wrath rotation's own search: Genesis 5 in place of Moonglow 3 (Wrath's mana cost is already
// low). It took Heart of the Wild 3 (Intellect) for Moonglow before Subtlety 3 became a must-take
// (standards 26); pinned (-require subtlety), the search gives Heart of the Wild's 3 points to
// Naturalist and Subtlety: 811.6 / 813.3 / 809.0 at 120 / 180 / 300 sec against 814.5 / 816.2 /
// 812.0 (5502220115501351-03-505), 2.9 / 2.9 / 3.0 for the threat cut (Tauren, 20,000 iterations,
// seed 1; the same on seed 1000003).
export const TalentsLevel60Wrath = PresetUtils.makePresetTalents('Balance Wrath', SavedTalents.create({ talentsString: '5502220115501351--5053' }), {
	tooltip: 'The Wrath rotation talent search best: Genesis in place of Moonglow, Subtlety from Restoration.',
	group: 'Synthetic',
});

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60Wrath],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = BalanceDruidOptions.create({
	innervateTarget: UnitReference.create(),
});

// The level 60 consumables, Forever launch Phase 1 only (nothing from Molten Core, Zul'Gurub,
// Blackwing Lair, Ahn'Qiraj or Naxxramas, so Wizard Oil and not Brilliant Wizard Oil). The five
// +2% crit elixirs (Mongoose, Grizzly, Owl, Sages, Whale) are one elixir in the client, so the
// moonkin takes one: Elixir of Sages with Elixir of Greater Intellect, level with the Owl and
// Greater Spirit at 180 sec and 1.3 ahead at 300. If both crit elixirs stacked, the Owl would
// add 15.1 at 180 sec. Other elixirs in different slots stack in the sim (notes.md). Taking
// each one out costs, at 180 sec on the first set: Flask of Supreme Power the most (Distilled
// Wisdom is 6 behind, the moonkin does not run dry in a raid fight), Elixir of Sages 14, Elixir
// of Nature Power 9, Major Spellblasting Potion 9 (over a Major Mana Potion), Lobster Stew 4,
// Goblin Sapper Charge 4.3 (6.5 / 4.3 / 3.6 at 120 / 180 / 300 sec, 3.9 to 7.1 on the other three
// builds, so all four take it, standards 11), Cerebral Cortex Compound 3.5 and Demonic Rune 1.4. Greater
// Mageblood Elixir and Greater Intellect are within the noise and stay for longer fights.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorSpellblastingPotion,
	// The fight is in a natural flask zone (standards 11), but the best natural flask, Aggression, is 85
	// behind the Flask of Supreme Power at 180 sec on both builds (Swiftness 101, Accuracy and Precision
	// 112 behind, the hit is capped).
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.WizardOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	naturePowerBuff: NaturePowerBuff.ElixirOfNaturePower,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfSages,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The raid the Level 60 numbers were measured under (tools/rotopt/mksettings -spec balance):
// the shared level 60 raid with the moonkin in a caster group, so a shaman's Mana Spring
// Totem, Blessing of Wisdom and Kings but not Might, and its own Moonkin aura. Under Forever
// Curse of the Elements covers every school and Curse of Shadow is gone, so only the first is
// on. The melee debuffs do nothing for it.
export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	bloodPact: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	frostResistanceAura: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
	salvation: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
	warchiefsBlessing: false,
});

export const DefaultPartyBuffs = PartyBuffs.create({});

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

// An engineer for the Goblin Sapper Charge (+4.3 at 180 sec) and an alchemist, like the other
// level 60 presets. The moonkin casts from range, so it does not melee between casts.
export const OtherDefaults = {
	distanceFromTarget: 30,
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
export const EncounterLevel60FaerieFireDuty = PresetUtils.makePresetEncounter('Balance · Faerie Fire duty', EncounterLevel60.encounter, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: Debuffs.create({ ...DefaultDebuffs, faerieFire: false, curseOfRecklessness: false }),
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumes,
});
// The Level 60 build, Tauren: 857.9 / 856.3 / 846.9 at 120 / 180 / 300 sec (seed 1, 20,000
// iterations; 858.4 / 856.6 / 847.4 on seed 1000003) with Subtlety 3 (2026-10-24); the Wrath build
// 811.6 / 813.3 / 809.0 (811.6 / 813.3 / 809.1). Every race a druid can be under Forever on the
// builds before Subtlety (864.0 / 862.4 / 853.1 for the Tauren on the 2026-10-23 sim), seed 1 (the
// Night Elf with its Elune's Light lines):
//
//	                     120 sec  180 sec  300 sec
//	Tauren                864.0    862.4    853.1   best overall and Horde
//	Night Elf             865.2    861.5    851.2   best Alliance (ahead at 120 sec)
//	Windshaper Skyborne   852.4    855.9    849.6
//	High Order Skyborne   852.3    855.8    849.3
//
// The Wrath build: Tauren 814.5 / 816.2 / 812.0, Windshaper Skyborne 815.1 / 816.6 / 812.5 (best
// Alliance; 0.5 and 0.6 ahead at 180 sec on the two seed sets, under the 0.5 lead on both a race
// switch needs), High Order Skyborne 815.0 / 816.5 / 812.4, Night Elf 815.1 / 814.4 / 810.1. Its
// knob search would weave Starfire (filler 3, +32), which is the Level 60 build, so the Wrath
// build keeps Wrath as its filler. No racial goes by weapon type.
// Race re-pick (2026-10-27, standards 29 and 16: each race with its own rotation and gear, a tie
// within noise keeps the race): the Level 60 build is Night Elf with Elune's Light held for the
// trinket, 851.2 at 180 sec against the Tauren's 841.5 (20,000 iterations, seed 1); the Wrath build
// is Windshaper Skyborne, 811.1 against the Tauren's 800.7. The Faerie Fire duty builds follow them.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Balance', {
	group: 'Synthetic',
	tooltip: 'Level 60 Night Elf on the synthetic Phase 1 gear with the Balance rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceNightElf,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
export const PresetBuildLevel60FaerieFire = PresetUtils.makePresetBuild('Balance · Faerie Fire duty', {
	group: 'Synthetic',
	tooltip:
		"The Balance build keeping Faerie Fire on the target for the raid. It turns off the raid's Faerie Fire and Curse of Recklessness, which share its armor slot.",
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60FaerieFire,
	encounter: EncounterLevel60FaerieFireDuty,
	race: Race.RaceNightElf,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
export const PresetBuildLevel60Wrath = PresetUtils.makePresetBuild('Balance Wrath', {
	group: 'Synthetic',
	tooltip: 'The Balance build with Wrath as the only filler.',
	gear: GearLevel60Phase1SyntheticWrath,
	talents: TalentsLevel60Wrath,
	rotation: APLLevel60Wrath,
	encounter: EncounterLevel60,
	race: Race.RaceSkyborneWindshaper,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
export const PresetBuildLevel60WrathFaerieFire = PresetUtils.makePresetBuild('Balance Wrath · Faerie Fire duty', {
	group: 'Synthetic',
	tooltip:
		"The Balance Wrath build keeping Faerie Fire on the target for the raid. It turns off the raid's Faerie Fire and Curse of Recklessness, which share its armor slot.",
	gear: GearLevel60Phase1SyntheticWrath,
	talents: TalentsLevel60Wrath,
	rotation: APLLevel60WrathFaerieFire,
	encounter: EncounterLevel60FaerieFireDuty,
	race: Race.RaceSkyborneWindshaper,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
