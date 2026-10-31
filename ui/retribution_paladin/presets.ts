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
import Level60ShockAPLJSON from './apls/level60_shock.apl.json';
import Level60ShockCrusaderDutyAPLJSON from './apls/level60_shock_crusader_duty.apl.json';
import Level60ShockRetAPLJSON from './apls/level60_shock_ret.apl.json';
import Level60ShockRetCrusaderDutyAPLJSON from './apls/level60_shock_ret_crusader_duty.apl.json';
import Level60ShockRetWisdomDutyAPLJSON from './apls/level60_shock_ret_wisdom_duty.apl.json';
import Level60ShockUndeadAPLJSON from './apls/level60_shock_undead.apl.json';
import Level60ShockWisdomDutyAPLJSON from './apls/level60_shock_wisdom_duty.apl.json';
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

// The paladin_ret template with the knobs the level 60 search settled on (see
// tools/rotopt/paladin_presets.sh), searched over 120, 180 and 300 sec together on the Level 60
// talents and the synthetic set. Judgement of the Crusader comes from the raid (our own Seal of
// the Crusader opener is 11 behind). Consecration goes first down to 20% mana. Seal of Command
// stays up and Twist of Light twists to Seal of Righteousness and back whenever no Echo is waiting
// for a swing, so every swing gets both seals (126 DPS over no twisting; in the last 1.5 sec before
// each swing until 2026-10-30, when extra attacks kept the swing timer and the free window won 13.9). Judgement and
// Holy Strike on cooldown, Seal of Righteousness judged whenever Judgement is ready (no swing
// window, 3 DPS over waiting for one), Hammer of Wrath below 20% health. Seal of Righteousness as the seal
// that stays up is 1.5 behind. The spell power on-use trinket and the Goblin Sapper Charge go at
// the top on cooldown: on the autocast line at the bottom they waited for a free GCD and went
// out 3 to 4.5 sec late (+6.1 / +3.3 / +2.1 at 120 / 180 / 300 sec, Shock +4.7 / +1.6 / +1.7).
// The trinket goes at the pull and is then held for the 20% execute phase whenever that loses no
// use of it (+1.3 at 180 sec, Shock +1.1, nothing at 120 sec). All four Level 60 rotations do this.
// Rotation audit pass 23 (2026-10-16): Hammer of Wrath goes above Consecration (at the very top
// in the Shock duty list), so in the execute phase it isn't held behind a Consecration that ticks
// for less. Paired over two seed sets at 20000 iterations, 120 / 180 / 300 sec: twist +2.1 /
// +2.2 / +2.2, sub Holy +2.0 / +2.5 / +1.6, their duty +2.2 to 2.7 / +2.4 to 2.5 / +1.7 to 2.2,
// Shock +0.8 / +1.4 / +0.5, Shock duty +1.9 / +1.1 / +0.8. Shock vs Undead keeps its order;
// pass 24 put Hammer of Wrath above Holy Strike in the vs Undead twist rotation, where Consecration
// is already last (see APLLevel60Undead).
// Regear and rotation round (2026-10-23, the search over 120, 180 and 300 sec together, then paired
// at 20000 iterations on seeds 1 and 1000003): the twist goes on at any mana (it stopped below 20%),
// +0.0 / +0.1 / +5.7 at 120 / 180 / 300 sec, the Crusader duty copy +0.0 / +0.1 / +6.2. The 20%
// floor never binds in a 120 sec fight.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Ret twist', Level60APLJSON, {
	tooltip: 'Level 60 rotation: Consecration, Seal of Command twisted with Seal of Righteousness before every swing, Judgement and Holy Strike on cooldown.',
	group: 'Synthetic',
});
// The rotation for the Level 60 Shock talents (paladin_presets.sh, searched the same way): no
// Twist of Light, so no twisting. Consecration first down to the last point of mana, Judgement on
// cooldown on Seal of Righteousness, Holy Strike on cooldown (90 DPS), Light's Vigil above 40%
// mana just before Holy Shock is ready (64), so the Holy Shock after it detonates it and starts no
// cooldown, then Holy Shock (148), and Hammer of Wrath below 20% health (23).
// Regear and rotation round (2026-10-23): Light's Vigil only when the next swing is more than 2 sec
// off. Its 1.5 sec cast resets the swing when it completes, so a Vigil started just before a swing
// cost a white hit: +45.6 / +35.4 / +26.3 at 120 / 180 / 300 sec (paired, two seed sets; 1.5 sec
// +47.6 / +34.4 / +24.3, 2.5 sec +39.8 / +32.8 / +25.3; a 10, 50 or 70% Vigil floor is 8 to 14
// behind 30% at 180 sec with it). The Shock vs Undead and Shock duty rotations take it too.
// User ruling 2026-10-27: Consecration after the seal, Judgement, Holy Strike, Light's Vigil and Holy
// Shock lines instead of first, the Vigil only when the next swing is more than 2.5 sec off, and above
// 60% mana in fights over 210 sec (with it Holy Shock goes out on cooldown whenever the Vigil is held).
// rotopt -eval, 20000 iterations, Human: 875.8 / 863.6 / 838.2 / 819.2 -> 894.1 / 881.2 / 840.2 / 827.2
// at 120 / 180 / 300 / 480 sec (Undead 879.5 / 866.6 / 840.1 / 820.7 -> 897.0 / 883.2 / 841.9 /
// 828.9). The Shock vs Undead and Shock duty rotations lose with the same three changes (vs Undead at
// every length, the duty at 300 sec), so they keep theirs.
export const APLLevel60Shock = PresetUtils.makePresetAPLRotation('Ret Shock (deep Holy)', Level60ShockAPLJSON, {
	tooltip: "Level 60 Holy Shock rotation: Judgement and Holy Strike on cooldown, Light's Vigil and Holy Shock, then Consecration.",
	group: 'Synthetic',
});
// The rotation for the Level 60 Shock talents (paladin_presets.sh): no Twist of Light and
// no Light's Vigil. Seal of Command stays up with no twisting (+6 over Seal of Righteousness),
// Consecration first down to the last point of mana, then Judgement, Holy Strike and Holy Shock on
// cooldown, and Hammer of Wrath below 20% health above Holy Strike. Searched over 120, 180 and 300 sec
// together, before and after the talent search.
export const APLLevel60ShockRet = PresetUtils.makePresetAPLRotation('Ret Shock', Level60ShockRetAPLJSON, {
	tooltip:
		'Level 60 Holy Shock rotation for the Retribution side talents: Seal of Command, Consecration, then Judgement, Holy Strike and Holy Shock on cooldown.',
	group: 'Synthetic',
});
// Level 60 (sub Holy): the twist rotation with Seal of Righteousness as the seal that stays up,
// twisting to Seal of Command before each swing (regear and rotation round, 2026-10-23): +1.1 /
// +0.5 / +0.9 at 120 / 180 / 300 sec on the sub Holy talents (paired, two seed sets). The twist
// talents keep Seal of Command, and the sub Holy Crusader duty is 0.4 behind with it, so that one
// keeps the twist duty rotation.
export const APLLevel60SubHoly = PresetUtils.makePresetAPLRotation('Ret sub-Holy', Level60SubHolyAPLJSON, {
	tooltip: 'Level 60 sub Holy rotation: the twist rotation with Seal of Righteousness up, twisted to Seal of Command before every swing.',
	group: 'Synthetic',
});
// Level 60 (vs Undead), for an Undead or Demon boss, where Exorcism and Holy Wrath land
// (standards 19; the default boss is type Unknown, where they do nothing). Searched on an Undead
// boss over 120, 180 and 300 sec together, with talent searches in between, from the twist and
// sub Holy builds. Seal of Righteousness stays up and Twist of Light twists to Seal of Command
// before each swing at any mana, Judgement on whichever seal is up, Holy Strike, Hammer of Wrath
// and Exorcism on cooldown, Holy Wrath above 80% mana (2.4 DPS), Consecration last above 20%.
// Rotation audit pass 24 (2026-10-16): Hammer of Wrath above Holy Strike, paired over two seed sets
// at 20000 iterations: +2.0 / +2.4 / +1.4 at 120 / 180 / 300 sec, sub Holy vs Undead +1.4 / +2.2 / +1.9.
// Regear and rotation round (2026-10-23): no Holy Wrath. Its 2 sec cast resets the swing when it
// completes and costs more than it adds: +9.6 / +6.8 / +8.3 at 120 / 180 / 300 sec on the new vs
// Undead talents (paired, two seed sets).
export const APLLevel60Undead = PresetUtils.makePresetAPLRotation('Ret sub-Holy · vs Undead', Level60UndeadAPLJSON, {
	tooltip: 'Level 60 rotation against an Undead or Demon boss: the twist rotation with Exorcism.',
	group: 'Synthetic',
});
// Level 60 Shock (vs Undead): the Shock rotation with Exorcism on cooldown, and Light's Vigil
// above 30% mana in fights up to 240 sec and above 60% in longer ones, which leaves the mana for
// Exorcism in long fights (a flat 60% was 9.7 / 10.5 behind at 120 / 180 sec). Holy Wrath is 4 to
// 11 behind: its 2 sec cast pushes back Holy Shocks and Holy Strikes. Since the regear round
// (2026-10-23) Light's Vigil waits until the next swing is more than 2 sec off, as in
// APLLevel60Shock: +52.8 / +40.4 / +17.5 at 120 / 180 / 300 sec (a flat 50 or 70% Vigil floor with
// it is 4.7 and 13.5 behind at 180 sec).
export const APLLevel60ShockUndead = PresetUtils.makePresetAPLRotation('Ret Shock (deep Holy) · vs Undead', Level60ShockUndeadAPLJSON, {
	tooltip: 'Level 60 Holy Shock rotation against an Undead or Demon boss, with Exorcism and more mana kept for it in long fights.',
	group: 'Synthetic',
});
// Judgement of the Crusader duty (standards 23): one paladin in the raid keeps Judgement of the
// Crusader on the boss, the others judge only their damage seals and get it from the raid. Under
// Forever it is 161 Holy damage taken for 40 sec, refreshed by the paladin's own melee hits, so the
// duty is one Seal of the Crusader after the base pre-pull and its Judgement at the pull (then
// again only if it ever drops). The twist duty keeps the twist knobs. The Shock duty puts Seal of
// Righteousness back up ahead of Consecration, since the opener's Judgement leaves it down (rotation
// audit pass 20: +6.0 / +3.0 / +0.5 at 120 / 180 / 300 sec), and Holy Strike ahead of Holy Shock.
// Since the regear round (2026-10-23) the twist duty twists at any mana (as APLLevel60) and the
// Shock duty holds Light's Vigil until the next swing is more than 2 sec off (as APLLevel60Shock:
// +37.0 / +32.2 / +24.3 at 120 / 180 / 300 sec).
export const APLLevel60CrusaderDuty = PresetUtils.makePresetAPLRotation('Ret sub-Holy · Judgement of the Crusader duty', Level60CrusaderDutyAPLJSON, {
	tooltip: 'Level 60 twist rotation that also keeps Judgement of the Crusader on the boss for the raid.',
	group: 'Synthetic',
});
export const APLLevel60ShockCrusaderDuty = PresetUtils.makePresetAPLRotation(
	'Ret Shock (deep Holy) · Judgement of the Crusader duty',
	Level60ShockCrusaderDutyAPLJSON,
	{
		tooltip: 'Level 60 Holy Shock rotation that also keeps Judgement of the Crusader on the boss for the raid.',
		group: 'Synthetic',
	},
);
export const APLLevel60ShockRetCrusaderDuty = PresetUtils.makePresetAPLRotation(
	'Ret Shock · Judgement of the Crusader duty',
	Level60ShockRetCrusaderDutyAPLJSON,
	{
		tooltip: 'The Ret Shock rotation that also keeps Judgement of the Crusader on the boss for the raid.',
		group: 'Synthetic',
	},
);
// Judgement of Wisdom duty (standards 23): the same as the Crusader duty, with Seal of Wisdom and its
// Judgement in place of the Crusader's. The twist duty copy serves the twist and sub Holy builds.
export const APLLevel60WisdomDuty = PresetUtils.makePresetAPLRotation('Ret sub-Holy · Judgement of Wisdom duty', Level60WisdomDutyAPLJSON, {
	tooltip: 'Level 60 twist rotation that also keeps Judgement of Wisdom on the boss for the raid.',
	group: 'Synthetic',
});
export const APLLevel60ShockWisdomDuty = PresetUtils.makePresetAPLRotation('Ret Shock (deep Holy) · Judgement of Wisdom duty', Level60ShockWisdomDutyAPLJSON, {
	tooltip: 'Level 60 Holy Shock rotation that also keeps Judgement of Wisdom on the boss for the raid.',
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
	APLLevel60Shock,
	APLLevel60ShockCrusaderDuty,
	APLLevel60ShockWisdomDuty,
	APLLevel60ShockRet,
	APLLevel60ShockRetCrusaderDuty,
	APLLevel60ShockRetWisdomDuty,
	APLLevel60Undead,
	APLLevel60ShockUndead,
];

// Since the regear round (2026-10-23) the class default is Level 60 (sub Holy), the build highest at
// 180 sec (944.5 against twist's 943.3).
export const DefaultAPL = APLLevel60SubHoly;

///////////////////////////////////////////////////////////////////////////
//                                 Talent presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/forever/talent-calc and copy the numbers in the url.

// The level 60 talent search on the synthetic Phase 1 set (2026-10-12, tools/rotopt
// -talent-search at 180 sec with the Level 60 rotation, then every finalist at 120, 180 and 300
// sec with 50000 iterations, Undead). We started it from three archetypes: Retribution to Twist
// of Light with the side points in Holy, the same with them in Protection, and a Holy Power
// build that leaves Twist of Light out (30 Holy, 21 Retribution).
//
// Client 1.60.1.70425 (2026-10-16) cut Judgement of Righteousness to 42.9% of spell power,
// Champion of the Light to 10 / 20 / 30% and Consecration to 4.2% a tick, and Blizzard's 15
// October beta notes halved Seal of Righteousness's spell power (12% a hit with a two-hander,
// once). The numbers in the comments below are from before. On the 70425 tree (raid hits,
// Concentration Aura, Human), at 120 / 180 / 300 sec with 20000 iterations:
// - Level 60: 831.2 / 826.9 / 819.2 (833.4 / 828.9 / 820.6 on the re-searched talents below).
// - Level 60 (sub Holy): 828.2 / 822.9 / 813.6.
// - Level 60 Shock: 867.1 / 851.0 / 812.3, ahead of the twist builds up to 180 sec.
// - Against an Undead boss: vs Undead 923.1 / 918.2 / 901.1, sub Holy vs Undead
//   925.9 / 922.0 / 913.4, Shock vs Undead 952.3 / 924.1 / 880.9.
// The talent searches (180 sec, Undead, before the raid hits) kept every build but two. Both
// moves trade one fight length against another, so the presets stay:
// - vs Undead: three Divine Intellect points into Divine Strength and Reverence into Precision 3
//   (`520030002-503-05225231001330301`) was +8 / +5 / -5.
// - sub Holy: a Purifying Power point into Holy Conduit 2 (`550030031--05225331001330321`) was a
//   tie up to 180 sec and +3.5 at 300.
//
// Level 60: the Retribution tree down to Twist of Light (Benediction, Improved Judgement, Holy
// Conduit, Conviction, Vindication, Sanctified Judgement, Seal of Command, Sacred Arbiter,
// Two-Handed Weapon Specialization, Vengeance, Champion of the Light, Instrument of Law 1),
// Toughness 5 to reach Precision 3, and Divine Strength 5, Divine Intellect 2 and Improved Seals
// 3 in Holy. 986.6 / 981.9 / 972.7 at 120 / 180 / 300 sec on the final set. The search's last move (a Vindication
// point into Divine Intellect for mana) was a tie. Twist of Light is the build: without it the
// seals cannot be twisted, and the Holy Power build was 24% behind before any search.
// The re-search on the 70425 tree with Seal of Righteousness's spell power counted once
// (2026-10-16, Human) moved a Divine Intellect point into Instrument of Law 2
// (`51003-503-05225331001330321`): +2.2 / +2.0 / +1.4 at 120 / 180 / 300 sec (paired 20000
// iterations, two seed sets), and the same on the Crusader duty copy.
//
// Regear and rotation round (2026-10-23: no melee hit suppression, so the cap is 8% and the set
// carries 8% hit; talent searches at 180 sec from each preset, then paired at 120 / 180 / 300 sec,
// 20000 iterations, seeds 1 and 1000003). Precision is now over the cap for melee, so the vs Undead
// search dropped Protection for the sub Holy side, and every sub Holy and vs Undead build takes
// `550030022--05225331001330321` (Divine Strength 5, Divine Intellect 5, Improved Seals 3,
// Reverence 2, Purifying Power 2, and Holy Conduit 2 for a Reverence point): sub Holy +0.1 / +1.0 /
// +12.6, vs Undead +28.0 / +29.0 / +32.5, sub Holy vs Undead +21.1 / +20.7 / +8.9 (seed 1). Twist
// keeps its talents (the same string is +0.5 / +0.8 / +10.2 on seed 1 and +0.2 / -0.0 / +10.1 on
// seed 1000003), and so does Shock (no move). Finals, Human, 120 / 180 / 300 sec, seed 1 (seed
// 1000003 within 1):
// - Level 60: 949.3 / 943.3 / 929.8. Level 60 (sub Holy): 950.9 / 944.5 / 935.2, the class default.
// - Level 60 Shock: 879.0 / 865.5 / 838.9.
// - Crusader duty: twist 936.8 / 936.8 / 928.0, sub Holy 936.9 / 936.6 / 932.1, Shock 876.4 / 868.6
//   / 843.7 (ahead of its base from 180 sec: Libram of Fervor's +33 Holy damage taken).
// - Against an Undead boss: vs Undead and sub Holy vs Undead 1053.2 / 1047.8 / 1029.3 (the same
//   build now), Shock vs Undead 982.3 / 965.2 / 934.4.
//
// Must-take talents (standards 26, 2026-10-23): every Level 60 build takes Pursuit of Justice 2
// (movement speed) whatever it costs. We put it in each string by hand and re-ran the talent
// searches at 180 sec with it pinned (tools/rotopt -require pursuitOfJustice), then paired the
// results at 120 / 180 / 300 sec (20000 iterations, seeds 1 and 1000003). What it costs, against
// the old builds with the Goblin Land Mine:
// - Level 60 (twist): Precision 3 to 2 and Vindication 3 to 2 (`51003-502-05225231201330321`),
//   -3.7 / -3.4 / -3.5.
// - Level 60 (sub Holy): Purifying Power 2, which does nothing on the generic boss
//   (`55003002--05225331201330321`), 0.0 at every length.
// - Level 60 Shock (and its vs Undead and duty copies): Vindication 2 to 0
//   (`55003003201121251--05225030201`, no search move beat it), -2.3 / -2.3 / -2.1 (vs Undead
//   -2.2 / -2.8 / -2.5).
// - Level 60 (sub Holy, vs Undead): Vindication 3 to 1 (`550030022--05225131201330321`), -4.1 /
//   -4.1 / -4.1. The search moved Reverence and Vindication a point each instead, 2.4 / 1.9 ahead
//   at 120 / 180 sec but 8.8 behind at 300, so the hand build stays. The search from the twist
//   talents ended 8 behind on the Prot side, so the vs Undead builds stay one preset.
// - Crusader duty: each keeps its base build's talents (twist -3.8 / -4.1 / -3.9, sub Holy 0.0,
//   Shock -1.6 / -1.9 / -2.1). The twist duty search moved to the sub Holy side
//   (`55003001-1-05225331201330321`, 4.0 / 4.6 / 7.9 over the twist talents), the same DPS as the
//   sub Holy duty build.
// Finals with Pursuit of Justice and no filler explosive, Human, 120 / 180 / 300 sec, seed 1 (seed
// 1000003 within 0.7):
// - Level 60: 951.5 / 945.1 / 932.7. Level 60 (sub Holy): 956.4 / 949.1 / 941.0, the class default.
// - Level 60 Shock: 881.5 / 869.0 / 843.0.
// - Crusader duty: twist 939.2 / 938.6 / 930.8, sub Holy 942.9 / 942.8 / 939.5, Shock 877.9 / 871.5
//   / 847.8.
// - Against an Undead boss: sub Holy vs Undead 1055.6 / 1050.8 / 1032.8, Shock vs Undead 985.0 /
//   968.4 / 938.0.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Ret twist', SavedTalents.create({ talentsString: '51003-502-05225231201330321' }), {
	tooltip: 'Level 60 talents from the talent search: Retribution to Twist of Light, Precision, Divine Strength and Improved Seals.',
	group: 'Synthetic',
});
// Level 60 (sub Holy): the side points in Holy instead of Protection (Divine Strength 5, Divine
// Intellect 5, Improved Seals 3, Reverence 3, Purifying Power 2) and Instrument of Law 2 for a
// Holy Conduit point. It was the search's best from the Holy seed. 981.9 / 977.2 / 965.1 at 120 /
// 180 / 300 sec, 0.5% behind Level 60 at 120 sec and 0.8% at 300 because it has no Precision.
export const TalentsLevel60SubHoly = PresetUtils.makePresetTalents('Ret sub-Holy', SavedTalents.create({ talentsString: '55003002--05225331201330321' }), {
	tooltip: 'Level 60 talents with the side points in Holy instead of Protection.',
	group: 'Synthetic',
});
// Level 60 Shock (deep Holy), called Level 60 Shock until 2026-10-27 (older comments use that
// name): the Holy Shock playstyle, its own talent and rotation search from a 31 Holy seed (Holy Shock, Divine Precision, Consecrated Ground, Holy Power, Light's Vigil). Divine
// Strength 5, Divine Intellect 5, Improved Seals 3, Reverence 3, Purifying Power 2, Illumination 1
// (to reach row 5), Divine Favor, Holy Shock, Divine Precision 2, Consecrated Ground 2, Holy
// Power 5 and Light's Vigil, with Benediction, Improved Judgement, Holy Conduit, Conviction,
// Vindication 2, Sanctified Judgement 3 and Sacred Arbiter in Retribution. 987.3 / 970.6 / 930.0
// at 120 / 180 / 300 sec on the final set (Spellblasting and Holy Power favour it), 1.2% behind
// Level 60 at 180 sec and 4.4% at 300 (Divine Precision does not cover its judgements): without Twist of Light it has nothing for the free GCDs
// between its cooldowns, and it runs out of mana in long fights. Light's Vigil is worth 64 and Holy Shock 148.
export const TalentsLevel60Shock = PresetUtils.makePresetTalents(
	'Ret Shock (deep Holy)',
	SavedTalents.create({ talentsString: '55003003201121251--05225030201' }),
	{
		tooltip: "Level 60 Holy Shock talents: deep Holy to Light's Vigil with Holy Power, and Retribution to Sacred Arbiter.",
		group: 'Synthetic',
	},
);
// Level 60 Shock (Level 60 Shock (Ret side) on 2026-10-26): Holy Shock with most points in
// Retribution. Holy Shock needs 20 Holy points, so Twist of Light (31 Retribution) is out of reach, and so is Light's Vigil. The talent
// search at 180 sec, with Holy Shock, Champion of the Light and Pursuit of Justice pinned, from a
// 21/0/30 seed: Divine Strength 5, Divine Intellect 5, Improved Seals 3, Reverence 3, Purifying
// Power 2, Illumination 1, Divine Favor, Holy Shock and Consecrated Ground 2 in Holy; Benediction 5,
// Improved Judgement 2, Holy Conduit 1, Conviction 5, Sanctified Judgement 3, Seal of Command,
// Pursuit of Justice 2, Sacred Arbiter, Two-Handed Weapon Specialization 2, Vengeance 3 and Champion
// of the Light 3 in Retribution. Human, 20000 iterations, seed 1: 939.5 / 932.4 / 910.7 at 120 / 180
// / 300 sec, 1.8% behind Level 60 (sub Holy) at 180 sec and 6.7% ahead of Level 60 Shock (deep Holy).
export const TalentsLevel60ShockRet = PresetUtils.makePresetTalents('Ret Shock', SavedTalents.create({ talentsString: '550030032011012--052150312012303' }), {
	tooltip: 'Level 60 Holy Shock talents with most points in Retribution: Holy Shock, Consecrated Ground and Champion of the Light, no Twist of Light.',
	group: 'Synthetic',
});
// Level 60 (sub Holy, vs Undead): the talent search from Level 60 (sub Holy) on an Undead boss
// moved Instrument of Law 2 to 1 for Holy Conduit 2 (Exorcism and Holy Wrath cost 40% less). It
// already has Purifying Power 2. 1074.6 / 1071.5 / 1061.7 at 120 / 180 / 300 sec against an
// Undead boss, the best build there at every length (9.6% over Level 60 (sub Holy) on its own
// boss at 180 sec): its Divine Intellect 5 and Reverence 3 pay for the extra spells.
export const TalentsLevel60SubHolyUndead = PresetUtils.makePresetTalents(
	'Ret sub-Holy · vs Undead',
	SavedTalents.create({ talentsString: '550030022--05225131201330321' }),
	{ tooltip: 'Level 60 sub Holy talents for an Undead or Demon boss, with Holy Conduit 2.', group: 'Synthetic' },
);

export const TalentPresets = [TalentsLevel60, TalentsLevel60SubHoly, TalentsLevel60Shock, TalentsLevel60ShockRet, TalentsLevel60SubHolyUndead];

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
	// (2026-10-23 regear; the Major Spellblasting Potion is within 1). The Shock builds run short of
	// mana and take the Major Mana Potion (ShockConsumes).
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

// The Shock builds' consumes (mksettings spec ret_shock): the Major Mana Potion, +12.8 / +18.4 / +12.8
// over the Dragonfire Potion on Shock / Shock vs Undead / Shock Crusader duty at 180 sec (+14.1 /
// +28.7 / +15.0 at 300; the Major Spellblasting Potion is 11 to 27 behind it). Every other slot
// keeps DefaultConsumes: the Flask of Supreme Power is 45 to 52 over the best natural flask
// (Aggression), and the Spirit and mana regen elixirs are within 0.6 of the next best.
export const ShockConsumes = Consumes.create({
	...DefaultConsumes,
	defaultPotion: Potions.MajorManaPotion,
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
	sunderArmor: true,
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
// The Level 60 encounter with the Shock consumes.
export const EncounterLevel60Shock = PresetUtils.makePresetEncounter('Ret Shock (deep Holy)', EncounterLevel60.encounter!, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: ShockConsumes,
});
// Races after the regear round (2026-10-23; 20000 iterations, seeds 1 and 1000003, 180 sec, seed 1):
// Human is the best race and the best Alliance one on every build, Undead the best (and only) Horde
// paladin, and Dwarf last. Human leads Undead by 5.1 / 5.8 / 2.4 on Level 60 / sub Holy / Shock, 6.0
// vs Undead, 3.7 Shock vs Undead and 5.4 / 5.6 / 2.9 on the three Crusader duty builds, with the same
// lead on the second seed set; Dwarf is 4.7 to 9.0 behind Human.
// Level 60: the Human on the synthetic set with the Level 60 talents and rotation. Client 70425
// made every racial weapon-free, and Human's Ruthless (2% crit) now beats Undead's Touch of the
// Grave on every Level 60 build at every length (+0.7 to +4.0 on two seed sets). Seed 1, 20000
// iterations, 120 / 180 / 300 sec, client 70425, Seal of Righteousness's spell power once,
// Concentration Aura:
// - Level 60: Human 831.2 / 826.9 / 819.2, Undead 829.5 / 824.8 / 816.6.
// - sub Holy: Human 828.2 / 822.9 / 813.6, Undead 825.6 / 820.4 / 810.6.
// - Shock: Human 867.1 / 851.0 / 812.3, Undead 866.4 / 849.4 / 810.4.
// Dwarf was about 5 behind Undead (with the double count). Human is the best race and the best
// Alliance one. Undead is the best (and only) Horde paladin. The earlier Undead pick (986.6 /
// 981.9 / 972.7 against Human's 971.0 / 966.4 / 957.6) was before client 70425, when Human's
// crit only counted with a sword.
// Race re-pick (2026-10-27, standards 29 and 16, on the 10-26 racial pass with each race's own rotation
// and gear): every Level 60 build is Undead again. At 180 sec (verify.py, 20000 iterations, seed 1),
// Undead against Human: Level 60 939.4 / 937.8, sub Holy 943.7 / 941.8, Shock (deep Holy) 866.6 / 863.6
// on the old Shock rotation (883.2 / 881.2 on the 2026-10-27 one), Shock 927.9 / 925.5; the Crusader
// duty and vs Undead builds follow their base builds.
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
// Level 60 (sub Holy): the same with the sub Holy talents. Undead 981.9 / 977.2 / 965.1 at 120 /
// 180 / 300 sec, Human 966.9 / 962.0 / 951.5, Dwarf 966.2 / 961.4 / 949.9.
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

// Level 60 Shock: the Holy Shock talents and rotation on Seal of Righteousness. Races at 120 /
// 180 / 300 sec (50000 iterations): Undead 987.3 / 970.6 / 930.0, Human 971.7 / 955.7 / 915.8,
// Dwarf 970.8 / 954.4 / 914.1.
export const PresetBuildLevel60Shock = PresetUtils.makePresetBuild('Ret Shock (deep Holy)', {
	group: 'Synthetic',
	tooltip: "Level 60 Undead Holy Shock build with Light's Vigil, on the synthetic Phase 1 gear.",
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Shock,
	rotation: APLLevel60Shock,
	encounter: EncounterLevel60Shock,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.ConcentrationAura,
		primarySeal: PaladinSeal.Righteousness,
	}),
});

// Level 60 Shock: the Retribution side Holy Shock talents and rotation on Seal of
// Command, with the twist builds' consumes: the Dragonfire Potion is +5.5 / +5.7 over the Major
// Mana Potion at 120 / 180 sec and 10.3 behind at 300 (paired, 20000 iterations).
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
export const EncounterLevel60ShockCrusaderDuty = PresetUtils.makePresetEncounter(
	'Ret Shock (deep Holy) · Judgement of the Crusader duty',
	EncounterLevel60.encounter!,
	{
		tanks: [],
		raidBuffs: DefaultRaidBuffs,
		debuffs: Debuffs.create({ ...DefaultDebuffs, judgementOfTheCrusader: TristateEffect.TristateEffectMissing }),
		buffs: DefaultIndividualBuffs,
		consumes: ShockConsumes,
	},
);
// The Judgement of the Crusader duty presets: each base build with its duty rotation and the duty
// encounter. Undead, 20000 iterations, 120 / 180 / 300 sec, against the base build on the
// base encounter: Level 60 973.6 / 973.1 / 967.4 (-13.7 / -9.0 / -5.9), sub Holy
// 969.4 / 969.0 / 959.9 (-12.8 / -8.6 / -5.5), Shock 974.4 / 960.9 / 923.5 (-13.7 / -10.5 /
// -7.0). Shock pays more: its pull Judgement goes on the Crusader instead of Righteousness and it
// has to put Seal of Righteousness back up.
// The twist duty build's talent search with Pursuit of Justice pinned moved to the sub Holy side and
// tied the sub Holy duty build, so the twist duty build keeps its base build's talents. The sub Holy
// duty build keeps the twist duty rotation.
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
export const PresetBuildLevel60ShockCrusaderDuty = PresetUtils.makePresetBuild('Ret Shock (deep Holy) · Judgement of the Crusader duty', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead Holy Shock build that keeps Judgement of the Crusader on the boss for the raid.',
	gear: GearLevel60Phase1SyntheticCrusaderDuty,
	talents: TalentsLevel60Shock,
	rotation: APLLevel60ShockCrusaderDuty,
	encounter: EncounterLevel60ShockCrusaderDuty,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.ConcentrationAura,
		primarySeal: PaladinSeal.Righteousness,
	}),
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
export const EncounterLevel60ShockWisdomDuty = PresetUtils.makePresetEncounter(
	'Ret Shock (deep Holy) · Judgement of Wisdom duty',
	EncounterLevel60.encounter!,
	{
		tanks: [],
		raidBuffs: DefaultRaidBuffs,
		debuffs: Debuffs.create({ ...DefaultDebuffs, judgementOfWisdom: false }),
		buffs: DefaultIndividualBuffs,
		consumes: ShockConsumes,
	},
);
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
export const PresetBuildLevel60ShockWisdomDuty = PresetUtils.makePresetBuild('Ret Shock (deep Holy) · Judgement of Wisdom duty', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead Holy Shock build that keeps Judgement of Wisdom on the boss for the raid.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Shock,
	rotation: APLLevel60ShockWisdomDuty,
	encounter: EncounterLevel60ShockWisdomDuty,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.ConcentrationAura,
		primarySeal: PaladinSeal.Righteousness,
	}),
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
export const EncounterLevel60ShockUndead = PresetUtils.makePresetEncounter('Ret Shock (deep Holy) · vs Undead', EncounterLevel60Undead.encounter!, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: ShockConsumes,
});
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
// Level 60 Shock (vs Undead): the Shock talents (a talent search on the Undead boss found no
// better) with the vs Undead Shock rotation. Undead 1070.9 / 1041.3 / 997.2, Human 1055.4 /
// 1026.6 / 982.5, Dwarf 1053.9 / 1024.7 / 980.6, 7.3% over Level 60 Shock on its own boss at
// 180 sec.
export const PresetBuildLevel60ShockUndead = PresetUtils.makePresetBuild('Ret Shock (deep Holy) · vs Undead', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead Holy Shock build against an Undead or Demon boss, with Exorcism.',
	gear: GearLevel60Phase1SyntheticUndead,
	talents: TalentsLevel60Shock,
	rotation: APLLevel60ShockUndead,
	encounter: EncounterLevel60ShockUndead,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: OptionsRighteousness,
});
