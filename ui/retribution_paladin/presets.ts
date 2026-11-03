import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	FirePowerBuff,
	Flask,
	Food,
	IndividualBuffs,
	IntellectElixir,
	ManaRegenElixir,
	MobType,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	ShadowPowerBuff,
	SpellPowerBuff,
	SpiritElixir,
	Stat,
	StrengthBuff,
	SunderRamp,
	TotemWeaponBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { PaladinAura, PaladinOptions as RetributionPaladinOptions, PaladinSeal } from '../core/proto/paladin.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import Level60APLJSON from './apls/level60.apl.json';
import Level60CrusaderDutyAPLJSON from './apls/level60_crusader_duty.apl.json';
import Level60ShockRetAPLJSON from './apls/level60_shock_ret.apl.json';
import Level60ShockRetCrusaderDutyAPLJSON from './apls/level60_shock_ret_crusader_duty.apl.json';
import Level60ShockRetWisdomDutyAPLJSON from './apls/level60_shock_ret_wisdom_duty.apl.json';
import Level60SubHolyAPLJSON from './apls/level60_sub_holy.apl.json';
import Level60UndeadAPLJSON from './apls/level60_undead.apl.json';
import Level60WisdomDutyAPLJSON from './apls/level60_wisdom_duty.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import Level60Phase1SyntheticCrusaderDutyGearJSON from './gear_sets/level60_phase1_synthetic_crusader_duty.gear.json';
import Level60Phase1SyntheticUndeadGearJSON from './gear_sets/level60_phase1_synthetic_undead.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

// Synthetic Phase 1 retribution gear, from tools/database/forever_synthetic_items_paladin.go (ids
// from 992000, not in the game).
//
// Forever's level 60 client gear is not tuned yet and Molten Core and Onyxia have no Forever data,
// so a gear search on real items ends up in Classic gear. These items rebuild the Era Phase 1
// retribution set (Lionheart Helm, Onyxia Tooth Pendant, Truestrike Shoulders, Onslaught Girdle,
// Band of Accuria, Quick Strike Ring, Spinal Reaper and the rest) under the orchestrator's
// standards 1b, the Forever budget for a hybrid: every armor piece at slot mod x (1.125 ilvl +
// 2.0) x 1.093 / 1.103 / 1.048 in weighted linear stats (spell power at 0.86), within 6% of the
// 1.7-norm shape line (4 or 5 stat lines a piece), on the shared item level table (66 for head,
// shoulders, chest, hands, legs and feet, 71 for neck, back, wrist, waist and rings). Armor totals
// 750.6, the two-hander 91.1 (target 90.6). The mix is Ret's hybrid one at the hit-adjusted
// targets: Stamina 17.6%, Strength 13.2, Agility 2.5, Intellect 10.0, Spirit 10.0, attack power
// 1.5, spell power 28.3, crit 7.5, hit 9.5 (80 rating, 8%, the melee cap from gear since Forever has
// no hit suppression; the 2026-10-23 regear moved the shoulders' 10 hit rating to 5 Strength and 5
// crit rating, +2.7 / +2.3 / +2.3 at 180 sec on twist / sub Holy / Shock). The file has every
// piece's numbers. The
// head and chest carry Justice Battlegear, Forever's Tier 1 retribution set, so only its 2-piece
// bonus (1% attack and cast speed) is live, the same shape as every class's preset. The 4-piece
// (36 attack power against Undead) adds nothing on the raid boss, and the 5-piece (0.5 sec off
// Judgement) would add 5.0 at 180 sec on the Level 60 build, 5.4 on sub Holy and 9.5 on Shock.
//
// The two-hander is an axe with Spinal Reaper's 203 - 305 at 3.4, the Era Phase 1 retribution
// weapon, so no paladin race gets a weapon racial from it. The relic is Libram of Law (Pix
// Xizzix, 4% Judgement of Righteousness damage), 0.8 ahead of no libram on the twist build and
// 3.8 on Shock (measured on an earlier set). After the 2026-10-23 regear (paired, 20000
// iterations, seeds 1 and 1000003, 120 / 180 / 300 sec), Libram of Invocation (5% off the seals) is
// -2.0 / -2.1 / +3.0 against Law on twist and -2.5 / -2.3 / +2.1 on sub Holy: the mana it saves only
// pays in long fights, so Law stays (talents and relics are judged at 180 sec). On Shock Invocation
// is about 4 behind and Libram of Infusion (6% crit on Holy Shock) 1.3. The Crusader duty builds
// take Libram of Fervor (GearLevel60Phase1SyntheticCrusaderDuty). The
// trinkets are synthetic (Forever's launch trinkets are not simulated, standards 8a): a stat
// trinket of 32 spell power and 18 crit rating (45.5 linear) and an on-use trinket of 140 spell
// power for 20 sec every 2 min. The enchants are the
// best Phase 1 ones, Forever's Enchanting 300 enchants included: Lesser Arcanum of Voracity (+8
// Strength) on head and legs, Necklace Spell Power, Cloak Agility, Greater Stats, Bracer Greater
// Spellpower, Gloves Holy Power, a Mystic Rugged Armor Kit on the boots and Spellblasting on the
// two-hander. Spellblasting is 10 DPS over Crusader on the twist build and 27 on Shock, Holy Power
// 2 to 7 over Gloves Superior Strength, and the Rugged kits lose on chest and legs and tie on the
// boots. The gear stays fixed for the talent, rotation and race searches.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Ret sub-Holy', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

// The same set with Tenets of the Silver Hand (1% more damage against Undead, spell 1302545) in the
// relic slot, for the vs Undead builds. Paired at 180 sec with 20000 iterations against an Undead
// boss it beats Libram of Law by 6.2 on Level 60 (vs Undead), 6.3 on sub Holy vs Undead and 3.4 on
// Shock vs Undead (2026-10-23).
export const GearLevel60Phase1SyntheticUndead = PresetUtils.makePresetGear('Ret sub-Holy · vs Undead', Level60Phase1SyntheticUndeadGearJSON, {
	tooltip: 'The synthetic Phase 1 set with Tenets of the Silver Hand, for an Undead boss.',
	group: 'Synthetic',
});

// The same set with Libram of Fervor (+33 to Judgement of the Crusader's Holy damage taken, a Phase 1
// world drop) in the relic slot, for the Crusader duty builds, which judge the Crusader themselves.
// Paired at 120 / 180 / 300 sec (20000 iterations, seeds 1 and 1000003, 2026-10-23 regear) it beats
// Libram of Law by +5.9 / +5.9 / +5.8 on Level 60 (Crusader duty) and +8.0 / +7.7 / +7.1 on Shock
// (Crusader duty). On the base builds it does nothing: the raid's Judgement of the Crusader is
// someone else's.
export const GearLevel60Phase1SyntheticCrusaderDuty = PresetUtils.makePresetGear(
	'Ret sub-Holy · Judgement of the Crusader duty',
	Level60Phase1SyntheticCrusaderDutyGearJSON,
	{
		tooltip: 'The synthetic Phase 1 set with Libram of Fervor, for the Crusader duty builds.',
		group: 'Synthetic',
	},
);

export const GearPresets = [GearLevel60Phase1Synthetic, GearLevel60Phase1SyntheticUndead, GearLevel60Phase1SyntheticCrusaderDuty];

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The paladin_ret template (tools/rotopt/paladin_ret.go) with the knobs tools/rotopt/paladin_presets.sh
// gives each build, re-searched on the launch realms' October 30 hotfix trees (1.60.1.70707+hf) (2026-11-01): talent searches at 180
// sec, knob searches over 120, 180 and 300 sec together, each choice checked at 20000 iterations at
// 120 / 180 / 300 sec. The script's notes have the numbers. Every rotation puts the on-use trinket
// (at the pull, then held for the execute phase when that loses no use) and the Goblin Sapper Charge
// at the top, Hammer of Wrath in the execute phase at any mana, and Judgement on cooldown.
//
// Ret twist: Unsealed Wrath. Seal of Command stays up; once it has stored 2 copies, Seal of
// Righteousness goes up for the swing that spends them all (and stores one of its own), and Seal of
// Command goes back once it has. When the mana won't cover 30 a second of the fight left, the swap
// waits for all 3 copies, since each swap costs a seal. Consecration first down to 20% mana, Holy
// Strike on cooldown.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Ret twist', Level60APLJSON, {
	tooltip:
		'Level 60 rotation: Seal of Command, swapped to Seal of Righteousness for the swing that spends its Unsealed Wrath copies, Consecration, Judgement and Holy Strike on cooldown.',
	group: 'Synthetic',
});
// Ret Shock: no Unsealed Wrath (Holy Shock takes 20 Holy points). Seal of Command up, Consecration
// first at any mana, Judgement, Holy Strike and Holy Shock on cooldown, Hammer of Wrath above Holy
// Strike.
export const APLLevel60ShockRet = PresetUtils.makePresetAPLRotation('Ret Shock', Level60ShockRetAPLJSON, {
	tooltip:
		'Level 60 Holy Shock rotation for the Retribution side talents: Seal of Command, Consecration, then Judgement, Holy Strike and Holy Shock on cooldown.',
	group: 'Synthetic',
});
// Ret sub-Holy: the twist rotation, swapping on 2 copies at any mana (its Reverence and Divine
// Intellect keep it in mana; the twist's mana rule loses 2.3 at 300 sec here).
export const APLLevel60SubHoly = PresetUtils.makePresetAPLRotation('Ret sub-Holy', Level60SubHolyAPLJSON, {
	tooltip: 'Level 60 sub Holy rotation: the twist rotation, swapping seals once Seal of Command has stored 2 Unsealed Wrath copies.',
	group: 'Synthetic',
});
// Ret sub-Holy vs Undead (standards 19): Exorcism lands, so the swap waits for all 3 copies to keep
// the mana for it, no Holy Wrath (its 2 sec cast resets the swing). Since round 70753 (Two-Handed
// Weapon Specialization on all Holy damage) Consecration goes first and Hammer of Wrath above it in
// the execute phase: +8.4 / +7.0 / +4.0 at 120 / 180 / 300 sec (was Consecration last, Hammer of
// Wrath above Holy Strike).
export const APLLevel60Undead = PresetUtils.makePresetAPLRotation('Ret sub-Holy · vs Undead', Level60UndeadAPLJSON, {
	tooltip: 'Level 60 rotation against an Undead or Demon boss: the twist rotation on 3 copies, with Exorcism.',
	group: 'Synthetic',
});
// Judgement of the Crusader duty (standards 23): one paladin in the raid keeps Judgement of the
// Crusader on the boss, the others get it from the raid. Under Forever it is 161 Holy damage taken for
// 40 sec, refreshed by the paladin's own melee hits, so the duty is Seal of the Crusader at -1.5 sec
// and its Judgement at the pull (then again only if it ever drops). The twist duty (shared by twist and
// sub Holy) keeps the sub Holy knobs; the Ret Shock duties put the damage seal back up ahead of
// Consecration, since the opener's judgement leaves it down (+0.7 / +1.8 / +2.1 at 120 / 180 / 300 sec).
export const APLLevel60CrusaderDuty = PresetUtils.makePresetAPLRotation('Ret sub-Holy · Judgement of the Crusader duty', Level60CrusaderDutyAPLJSON, {
	tooltip: 'Level 60 twist rotation that also keeps Judgement of the Crusader on the boss for the raid.',
	group: 'Synthetic',
});
export const APLLevel60ShockRetCrusaderDuty = PresetUtils.makePresetAPLRotation(
	'Ret Shock · Judgement of the Crusader duty',
	Level60ShockRetCrusaderDutyAPLJSON,
	{
		tooltip: 'The Ret Shock rotation that also keeps Judgement of the Crusader on the boss for the raid.',
		group: 'Synthetic',
	},
);
// Judgement of Wisdom duty (standards 23): the same as the Crusader duty, with Seal of Wisdom and its
// Judgement in place of the Crusader's (the template's jotc=2). The twist duty copy serves the twist
// and sub Holy builds.
export const APLLevel60WisdomDuty = PresetUtils.makePresetAPLRotation('Ret sub-Holy · Judgement of Wisdom duty', Level60WisdomDutyAPLJSON, {
	tooltip: 'Level 60 twist rotation that also keeps Judgement of Wisdom on the boss for the raid.',
	group: 'Synthetic',
});
export const APLLevel60ShockRetWisdomDuty = PresetUtils.makePresetAPLRotation('Ret Shock · Judgement of Wisdom duty', Level60ShockRetWisdomDutyAPLJSON, {
	tooltip: 'The Ret Shock rotation that also keeps Judgement of Wisdom on the boss for the raid.',
	group: 'Synthetic',
});

export const APLPresets = [
	APLLevel60,
	APLLevel60SubHoly,
	APLLevel60CrusaderDuty,
	APLLevel60WisdomDuty,
	APLLevel60ShockRet,
	APLLevel60ShockRetCrusaderDuty,
	APLLevel60ShockRetWisdomDuty,
	APLLevel60Undead,
];

// The class default is Level 60 (sub Holy), the build highest at 180 sec since the regear round
// (2026-10-23). On the hotfix trees it ties Ret twist (1038.1 against 1039.3 at 180 sec, within the
// noise), so it stays; in round 70753 too (1089.8 against 1090.9, seed 1000003 1089.3 against
// 1090.1).
export const DefaultAPL = APLLevel60SubHoly;

///////////////////////////////////////////////////////////////////////////
//                                 Talent presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/forever/talent-calc and copy the numbers in the url.

// Re-searched on the launch realms' October 30 hotfix trees (1.60.1.70707+hf) (2026-11-01), which reworked the whole tree: every saved
// string reads differently. Each build is a talent search at 180 sec (tools/rotopt -talent-search,
// Pursuit of Justice pinned, standards 26) from a seed of its archetype, on its own rotation, then
// the knobs again (tools/rotopt/paladin_presets.sh). Undead, 20000 iterations, seed 1, 120 / 180 / 300
// sec on the synthetic Phase 1 set, before the hotfix -> after:
// - Ret twist: 979.1 / 974.5 / 955.3 -> 1043.4 / 1039.3 / 1028.9.
// - Ret sub-Holy: 983.5 / 977.9 / 970.6 -> 1042.3 / 1038.1 / 1031.7.
// - Ret Shock: 968.5 / 960.6 / 940.2 -> 983.0 / 974.9 / 968.5.
// - Ret Shock (deep Holy), the 31 point Holy Shock build: 920.7 / 908.4 / 880.4 -> 863.3 / 855.6 /
//   849.3, 17.7% behind Ret twist at 180 sec (it was 6.8%). Dropped by the user's ruling (2026-11-02,
//   "drop deep holy shock"), with its Crusader and Wisdom duty runs and its vs Undead build (979.9 /
//   973.5 / 952.0).
// - Against an Undead boss: sub-Holy vs Undead 1074.0 / 1069.6 / 1052.7 -> 1153.9 / 1151.4 / 1145.1.
// What moved them: Conviction became 6% crit with spells and attacks (was 5% melee), Unsealed Wrath's
// stored Seal of Command copies, Champion of the Light's free Hammer of Wrath cast, and Holy Power and
// Light's Vigil leaving damage (the deep Holy build lost both). Swift Judgement, now a free Judgement
// of its own, is worth 26 on the twist build, but its 16 Protection points cost 60.
//
// Round 70753 (2026-11-03): the November 2 development notes' Vengeance 5 stacks (was 3) and
// Two-Handed Weapon Specialization on all Holy damage (notes-only values, standards 37), and the
// race base stats fix. The talent searches at 180 sec (Pursuit of Justice pinned, Holy Shock too on
// Ret Shock) kept every build's talents, and the knob searches every rotation but vs Undead's.
// Undead, 20000 iterations, seed 1, 120 / 180 / 300 sec, before -> after:
// - Ret twist: 1043.4 / 1039.3 / 1028.9 -> 1094.3 / 1090.9 / 1079.9.
// - Ret sub-Holy: 1042.3 / 1038.1 / 1031.7 -> 1092.3 / 1089.8 / 1084.3.
// - Ret Shock: 983.0 / 974.9 / 968.5 -> 1037.5 / 1030.2 / 1025.5.
// - Ret sub-Holy vs Undead: 1153.9 / 1151.4 / 1145.1 -> 1222.9 / 1221.4 / 1214.0.
// Undead stays the best race on each (Human 2.3 to 5.1 behind at 180 sec on both seed sets, Dwarf
// 8.4 to 12.2).
//
// Ret twist: Retribution to Unsealed Wrath (Benediction, Improved Judgement, Conviction, Sanctified
// Judgement, Seal of Command, Pursuit of Justice, Sacred Arbiter, Two-Handed Weapon Specialization,
// Repentance, Vengeance, Champion of the Light 5, a Vindication point to reach row 4), Divine Strength
// 5, Improved Seals 5 and Holy Conduit 2 in Holy, Divine Intellect 5 and Precision 3 in Protection.
// The search kept the seed: Champion of the Light is 80 a point pair, Improved Seals 10 a point.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Ret twist', SavedTalents.create({ talentsString: '050050002-503-05233101201301351' }), {
	tooltip: 'Level 60 talents: Retribution to Unsealed Wrath, Divine Strength, Improved Seals and Holy Conduit in Holy, Divine Intellect and Precision.',
	group: 'Synthetic',
});
// Ret sub-Holy: the same Retribution 31, with Divine Precision 2 (8% hit on Holy Strike and
// Consecration) and Reverence 2 in Holy instead of Precision (Divine Strength 5, Healing Light 1 to
// reach row 4, Reverence 2, Improved Seals 5, Holy Conduit 2), and Divine Intellect 3.
export const TalentsLevel60SubHoly = PresetUtils.makePresetTalents('Ret sub-Holy', SavedTalents.create({ talentsString: '0512500022-3-05233101201301351' }), {
	tooltip: 'Level 60 talents with the side points in Holy: Divine Precision and Reverence instead of Precision.',
	group: 'Synthetic',
});
// Ret Shock: Holy Shock with most points in Retribution. Holy 21 (Divine Strength 5, Healing Light 1,
// Reverence 2, Improved Seals 5, Purifying Power 2, Divine Favor, Holy Conduit 2, Divine Precision 2,
// Holy Shock) plus Consecrated Ground 2, Retribution 28 to Champion of the Light 3 (Benediction,
// Improved Judgement, Conviction, Sanctified Judgement, Vindication 1, Seal of Command, Pursuit of
// Justice, Sacred Arbiter, Two-Handed Weapon Specialization, Repentance, Vengeance). The search kept
// the seed.
export const TalentsLevel60ShockRet = PresetUtils.makePresetTalents('Ret Shock', SavedTalents.create({ talentsString: '05125210220021--0523310120130133' }), {
	tooltip: 'Level 60 Holy Shock talents with most points in Retribution: Holy Shock, Consecrated Ground and Champion of the Light 3, no Unsealed Wrath.',
	group: 'Synthetic',
});
// Ret sub-Holy vs Undead: the talent search from Ret sub-Holy on an Undead boss took Purifying Power 2
// (Exorcism and Holy Wrath a third faster, +38) for Healing Light and Divine Intellect points.
export const TalentsLevel60SubHolyUndead = PresetUtils.makePresetTalents(
	'Ret sub-Holy · vs Undead',
	SavedTalents.create({ talentsString: '0502520022-2-05233101201301351' }),
	{ tooltip: 'Level 60 sub Holy talents for an Undead or Demon boss, with Purifying Power 2.', group: 'Synthetic' },
);

export const TalentPresets = [TalentsLevel60, TalentsLevel60SubHoly, TalentsLevel60ShockRet, TalentsLevel60SubHolyUndead];

export const DefaultTalents = TalentsLevel60SubHoly;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// The Level 60 builds run Concentration Aura: the raid's paladins already give Retribution
// Aura, which only fires when something hits the paladin in melee, while Concentration Aura's 35%
// chance to ignore pushback from the raid hits is +0.3 on twist, +2.7 on Shock, +2.5 on Shock vs
// Undead and +1.0 on sub Holy vs Undead at 180 sec (20000 iterations, two seed sets).
export const DefaultOptions = RetributionPaladinOptions.create({
	aura: PaladinAura.ConcentrationAura,
	primarySeal: PaladinSeal.Command,
});

// The level 60 consumables. We tried every consumable the sim applies in each slot on the
// synthetic set at 180 sec (and 300 for the mana ones), each against the full set. Flask of
// Supreme Power is worth 65, Dragonbreath Chili 25, Shadow Oil 30 over no oil (18 over an
// Elemental Sharpening Stone), Elixir of Greater Firepower 16 (Forever made it the Elixir of
// Holy Power, +40 Holy, so it is for the seals, judgements and Consecration), Elixir of the
// Mongoose 21 (before the one crit elixir rule), R.O.I.D.S. 9, Greater Arcane Elixir 15, Goblin Land Mine 4, Elixir of Shadow Power
// 1.5, Goblin Sapper 1.6. The Major Mana Potion and Demonic Rune are 39 and 28 at 300 sec, and
// the Major Spellblasting or Frenzy Potion in the potion's place lose 33 to 36 there. Flask of the
// Titans or Distilled Wisdom give nothing here.
//
// With Forever's Phase 1 consumables in the sim (the shared consumables commit): a character
// takes at most one of the five 2% crit elixirs (Mongoose, Grizzly, Owl, Sages, Whale, the
// orchestrator's ruling, see notes.md). With Greater Agility, Juju Power, Greater Intellect and
// Greater Spirit in the slots, the Level 60 build does 965.1 at 180 sec, and one crit elixir
// adds 13 (Mongoose, Grizzly) to 15 (Owl, Sages, Whale, all the same since it is the crit that
// counts). We take the Elixir of the Owl in the Intellect slot (980.5), on the Shock build too
// (937.1, Sages is 0.5 ahead there, inside the noise). These were on the earlier set with
// trinkets. On the final set, a second crit elixir (Mongoose in place of Cunning) would add
// 6.2 on the Level 60 build at 180 sec and 4.3 on Shock (Forever's Mongoose gives spell crit too). Elixir of
// Giants is 1 to 2 behind Juju Power, and Elixir of Ferocity (18 Strength and Agility) beats Juju
// Power on the final set by 1.6 / 1.5 / 0.8 at 180 sec (twist / sub Holy / Shock). Greater Spirit is worth 1.3 on Shock and nothing on the
// twist build, Greater Agility 9 and 5. Lobster Stew beats Nightfin Soup by 3 (Mightfish Steak
// and Flank au Poivre are 3 to 4 behind), and Greater Mageblood Elixir takes the Mageblood
// Potion's place (the same at 180 sec, 3.3 at 300). From the consumables audit (2026-10-13):
// Forever's Elixir of Cunning (25 Agility and 25 Intellect, no crit, so it goes with the Owl)
// takes Greater Agility's place, +8.9 / +9.4 / +7.3 at 180 sec on twist / sub Holy / Shock and
// +10.5 on sub Holy vs Undead. The Major Frenzy Potion loses to the Major Mana Potion and
// Mightfish Steak to Lobster Stew.
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfCunning,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredDemonicRune,
	// The Dragonfire Potion is +5.2 / +5.2 over the Major Mana Potion on Level 60 / sub Holy at 180 sec
	// (2026-10-23 regear; the Major Spellblasting Potion is within 1).
	defaultPotion: Potions.DragonfirePotion,
	dragonBreathChili: true,
	// No filler explosive (standards 11, 2026-10-23: only when it gains). At 180 sec, 20000
	// iterations, seeds 1 and 1000003, no explosive is ahead of the Goblin Land Mine and the EZ-Thro
	// Dark Bomb (the same 450 average Fire damage) by 4.9 to 7.0 on every Level 60 build, and of the
	// EZ-Thro Thorium Grenade by 5.7 to 7.7: each throw's 1 sec cast resets the swing. The Goblin
	// Sapper Charge (no cast) stays, +4.4 to 5.1.
	firePowerBuff: FirePowerBuff.ElixirOfGreaterFirepower,
	// The fight is in a natural flask zone (standards 11), but the best natural flask, Precision, is 19
	// to 26 behind the Flask of Supreme Power at 180 sec (Aggression 48 to 52 behind on Shock).
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfTheOwl,
	mainHandImbue: WeaponImbue.ShadowOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	shadowPowerBuff: ShadowPowerBuff.ElixirOfShadowPower,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfGreaterSpirit,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.ROIDS,
});

// The level 60 raid (tools/rotopt/mksettings): both paladin blessings, Kings, Judgement of the
// Crusader and of Wisdom from another paladin, every debuff, Curse of the Elements (10% Holy
// damage under Forever, worth 62 here). The paladin is in a melee group with a shaman, who puts
// down Strength of Earth, Windfury Totem and Mana Spring, and Grace of Air from a second shaman in the raid as well (the user's ruling 2026-10-16, "both totems": client 1.60.1.70425 lets Grace of Air stack with Windfury Totem and makes both raid-wide).
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
	manaSpringTotem: TristateEffect.TristateEffectRegular,
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

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	fengusFerocity: false,
	moldarsMoxie: false,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
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

// At level 60 the paladin is an engineer (the Goblin Sapper and the Land Mine) and a
// blacksmith. Gear another profession makes is only in the presets when it binds on equip, so it
// can be bought.
//
// The distance and the channel clip delay are the ones the DPS Report sims with.
export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceUndead,
	distanceFromTarget: 5,
	channelClipDelay: 50,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor) as type Unknown, the
// generic boss of standards 19, so no creature-type bonus applies (Exorcism and Holy Wrath do
// nothing on it), a 3 minute fight, and the default buffs, debuffs and consumes.
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
// Races on the hotfix trees (2026-11-01, standards 29 and 16: verify.py, 20000 iterations, seed 1, 180
// sec, each race with its own gear where req/races has it): Undead stays the best race on every build.
// Undead / Human / Dwarf: Ret twist 1039.0 / 1032.8 / 1029.2, sub-Holy 1038.1 / 1033.0 / 1029.5, Ret
// Shock 974.9 / 968.0 / 964.8, sub-Holy vs Undead 1151.4 / 1145.8 / 1138.1. Human is the best Alliance
// paladin. The duty builds follow their base builds.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Ret twist', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead on the synthetic Phase 1 gear with the Ret twist talents and rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});
// Level 60 (sub Holy): the same with the sub Holy talents.
export const PresetBuildLevel60SubHoly = PresetUtils.makePresetBuild('Ret sub-Holy', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead with the side talent points in Holy instead of Protection.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60SubHoly,
	rotation: APLLevel60SubHoly,
	encounter: EncounterLevel60,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});

// Level 60 Shock: the Retribution side Holy Shock talents and rotation on Seal of Command, with the
// twist builds' consumes.
export const PresetBuildLevel60ShockRet = PresetUtils.makePresetBuild('Ret Shock', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead Holy Shock build with most points in Retribution, on the synthetic Phase 1 gear.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60ShockRet,
	rotation: APLLevel60ShockRet,
	encounter: EncounterLevel60,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.ConcentrationAura,
		primarySeal: PaladinSeal.Command,
	}),
});

// The Level 60 encounter for the Judgement of the Crusader duty presets: the same boss and raid
// without the raid's Judgement of the Crusader, so the debuff comes only from the paladin's own
// judgement.
export const EncounterLevel60CrusaderDuty = PresetUtils.makePresetEncounter('Judgement of the Crusader duty', EncounterLevel60.encounter!, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: Debuffs.create({ ...DefaultDebuffs, judgementOfTheCrusader: TristateEffect.TristateEffectMissing }),
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumes,
});
// The Judgement of the Crusader duty presets: each base build with its duty rotation and talents and
// the duty encounter. Undead, 20000 iterations, seed 1, 120 / 180 / 300 sec (against the base build on
// the base encounter at 180): twist 1029.9 / 1033.3 / 1026.1 (-6.0), sub Holy 1028.4 / 1032.1 / 1031.6
// (-6.0), Ret Shock 972.0 / 974.4 / 971.5 (-0.5).
export const PresetBuildLevel60CrusaderDuty = PresetUtils.makePresetBuild('Ret twist · Judgement of the Crusader duty', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead with the twist talents that keeps Judgement of the Crusader on the boss for the raid.',
	gear: GearLevel60Phase1SyntheticCrusaderDuty,
	talents: TalentsLevel60,
	rotation: APLLevel60CrusaderDuty,
	encounter: EncounterLevel60CrusaderDuty,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});
export const PresetBuildLevel60SubHolyCrusaderDuty = PresetUtils.makePresetBuild('Ret sub-Holy · Judgement of the Crusader duty', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead with the sub Holy talents that keeps Judgement of the Crusader on the boss for the raid.',
	gear: GearLevel60Phase1SyntheticCrusaderDuty,
	talents: TalentsLevel60SubHoly,
	rotation: APLLevel60CrusaderDuty,
	encounter: EncounterLevel60CrusaderDuty,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});
export const PresetBuildLevel60ShockRetCrusaderDuty = PresetUtils.makePresetBuild('Ret Shock · Judgement of the Crusader duty', {
	group: 'Synthetic',
	tooltip: 'The Ret Shock build that keeps Judgement of the Crusader on the boss for the raid.',
	gear: GearLevel60Phase1SyntheticCrusaderDuty,
	talents: TalentsLevel60ShockRet,
	rotation: APLLevel60ShockRetCrusaderDuty,
	encounter: EncounterLevel60CrusaderDuty,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.ConcentrationAura,
		primarySeal: PaladinSeal.Command,
	}),
});

// The Level 60 encounter for the Judgement of Wisdom duty presets: the same boss and raid without the
// raid's Judgement of Wisdom, so the debuff comes only from the paladin's own judgement.
export const EncounterLevel60WisdomDuty = PresetUtils.makePresetEncounter('Judgement of Wisdom duty', EncounterLevel60.encounter!, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: Debuffs.create({ ...DefaultDebuffs, judgementOfWisdom: false }),
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumes,
});
// The Judgement of Wisdom duty presets: each base build with its Wisdom duty rotation and the duty
// encounter, on the base build's gear.
export const PresetBuildLevel60WisdomDuty = PresetUtils.makePresetBuild('Ret twist · Judgement of Wisdom duty', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead with the twist talents that keeps Judgement of Wisdom on the boss for the raid.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60WisdomDuty,
	encounter: EncounterLevel60WisdomDuty,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});
export const PresetBuildLevel60SubHolyWisdomDuty = PresetUtils.makePresetBuild('Ret sub-Holy · Judgement of Wisdom duty', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead with the sub Holy talents that keeps Judgement of Wisdom on the boss for the raid.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60SubHoly,
	rotation: APLLevel60WisdomDuty,
	encounter: EncounterLevel60WisdomDuty,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});
export const PresetBuildLevel60ShockRetWisdomDuty = PresetUtils.makePresetBuild('Ret Shock · Judgement of Wisdom duty', {
	group: 'Synthetic',
	tooltip: 'The Ret Shock build that keeps Judgement of Wisdom on the boss for the raid.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60ShockRet,
	rotation: APLLevel60ShockRetWisdomDuty,
	encounter: EncounterLevel60WisdomDuty,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.ConcentrationAura,
		primarySeal: PaladinSeal.Command,
	}),
});

// The Level 60 boss as an Undead, for the vs Undead builds (Exorcism and Holy Wrath land on
// Undead and Demons only).
export const EncounterLevel60Undead = PresetUtils.makePresetEncounter(
	'Undead boss',
	Encounter.create({
		...EncounterLevel60.encounter!,
		targets: EncounterLevel60.encounter!.targets.map(target => ({ ...target, mobType: MobType.MobTypeUndead })),
	}),
	{
		tanks: [],
		raidBuffs: DefaultRaidBuffs,
		debuffs: DefaultDebuffs,
		buffs: DefaultIndividualBuffs,
		consumes: DefaultConsumes,
	},
);
const OptionsRighteousness = RetributionPaladinOptions.create({
	aura: PaladinAura.ConcentrationAura,
	primarySeal: PaladinSeal.Righteousness,
});
// The vs Undead builds, on an Undead boss (50000 iterations, 120 / 180 / 300 sec). Undead was
// the best race on each before client 70425. Now Human is (seed 1, 20000 iterations): vs Undead
// Human 923.1 / 918.2 / 901.1 against Undead 920.0 / 915.2 / 897.6, sub Holy vs Undead
// 925.9 / 922.0 / 913.4 against 923.5 / 919.0 / 910.0, Shock vs Undead 952.3 / 924.1 / 880.9
// against 950.7 / 921.5 / 878.8. The older numbers:
// Level 60 (vs Undead) was its own build until 2026-10-24. It had the same talents, gear and
// rotation as the sub Holy one below since the regear round, so it went.
// Level 60 (sub Holy, vs Undead): Undead 1074.6 / 1071.5 / 1061.7, Human 1058.5 / 1055.7 /
// 1046.3, Dwarf 1058.1 / 1055.2 / 1045.2. The best build against an Undead boss.
export const PresetBuildLevel60SubHolyUndead = PresetUtils.makePresetBuild('Ret sub-Holy · vs Undead', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead with the sub Holy talents against an Undead or Demon boss, with Exorcism and Holy Wrath.',
	gear: GearLevel60Phase1SyntheticUndead,
	talents: TalentsLevel60SubHolyUndead,
	rotation: APLLevel60Undead,
	encounter: EncounterLevel60Undead,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: OptionsRighteousness,
});
