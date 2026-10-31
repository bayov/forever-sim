import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	Explosive,
	FirePowerBuff,
	Flask,
	Food,
	IndividualBuffs,
	IntellectElixir,
	LevelingOptions,
	ManaRegenElixir,
	MobType,
	NaturePowerBuff,
	Potions,
	Profession,
	PseudoStat,
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
	UnitReference,
	UnitReference_Type,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { EnhancementShaman_Options as EnhancementShamanOptions, ShamanSyncType } from '../core/proto/shaman.js';
import {
	LevelingPlanSettings,
	LevelingRoute,
	LevelingRotation,
	LevelingRotationPreset,
	LevelingTalentBuild,
	LevelingTalentPick,
	LevelingTalentRequirement,
	LevelingTrainingSettings,
	SavedTalents,
} from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import Level30LevelingAPLJSON from './apls/level30_leveling.apl.json';
import LevelingPlan10APLJSON from './apls/leveling_plan_10.apl.json';
import LevelingPlan15APLJSON from './apls/leveling_plan_15.apl.json';
import LevelingPlan20APLJSON from './apls/leveling_plan_20.apl.json';
import LevelingPlan25APLJSON from './apls/leveling_plan_25.apl.json';
import LevelingPlan30APLJSON from './apls/leveling_plan_30.apl.json';
import LevelingPlanAuditedJSON from './apls/leveling_plan_audited.json';
import LevelingRouteJSON from './leveling_route.json';
import LevelingTrainingJSON from './leveling_training.json';
import MagmaAPLJSON from './apls/magma.apl.json';
import ManaTideAPLJSON from './apls/mana_tide.apl.json';
import OptimizedAPLJSON from './apls/optimized.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import Level60Phase1SyntheticBurningTotemGearJSON from './gear_sets/level60_phase1_synthetic_burning_totem.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

// A guess at Phase 1 BiS under Forever, made of the synthetic items in
// tools/database/forever_synthetic_items.go (ids from 990000, not in the game).
//
// Forever's level 60 gear in the client is not tuned yet, and Molten Core and Onyxia have no
// Forever data. So the real-item Phase 1 search ends up in Classic gear and can't show Forever's
// hybrid Strength, Intellect and spell power items. The synthetic items are Era Phase 1 hybrids
// (Crown of Destruction, the Bloodmail set, Death's Clutch, Wristguards of True Flight) rebuilt
// at Phase 1 epic budgets with Forever's cheaper spell power, and six of them carry The
// Spiritcaller's Rage set bonuses. The file has the details.
//
// On 2026-10-05 a gear search over these and every real item up to item level 78 kept the
// synthetic items everywhere but the chest. Timbermaw Tunic (Forever, Leatherworking 300, binds
// on equip) beats the synthetic chest by 7.2. The relic is Burning Totem (Flame Shock 3 sec
// longer), 5.9 ahead of no relic. Forever's Totem of the Storm is worth 1.0, because the
// rotation casts few Lightning Bolts. The set is 938.3 at 180 sec on the Level 60 talents,
// against 848.5 for the real-item Phase 1 set. The talent search and the rotation search found
// nothing better on it. The enchants are the Phase 1 set's, not searched again.
//
// The set has 100 hit rating, the 10% melee hit the Era Phase 1 BiS carries. Without it the
// synthetic set misses 7% of its white swings and sims at 841.7.
//
// On 2026-10-06 we gave the synthetic items Spirit (about 10% of their budget) and a
// different stat mix per slot, and split the ring into two (a Strength and spell power one,
// and a caster one with Intellect and Spirit). The Forever devs want Spirit to matter
// again, so we expect it on endgame gear. The set dropped from 938.6 to 907.8 at 180 sec,
// because Spirit only gives mana while casting through Improved Stormstrike (50%) and is
// worth about 0.08 DPS a point. The DPS numbers in the comments of this file from before
// that date are on the old set. A gear search over the new set swaps the neck, cloak and
// both rings for real items (+23 at 180 sec), and keeps every other synthetic piece. The
// preset stays all synthetic.
//
// On 2026-10-12 every class's synthetic set moved to one cross-class standard
// (orchestration/standards.md): item level 66 for the head, shoulders, chest, hands, legs and
// feet, 71 for the neck, back, wrist, waist and rings, 76 for the two-hander, hit to the 9% melee
// cap against a level 63 boss (90 rating), and The Spiritcaller's Rage on only the hands and legs,
// so its 2 piece bonus (1% attack and cast speed) is the only one live. The 4 and 5 piece bonuses
// (36 attack power against Elementals, Stormstrike 0.5 sec sooner) would add more. The preset
// wears the synthetic chest instead of Timbermaw Tunic. The enchants are the best outside the
// raids: Crusader on the two-hander (+9.0 over Enchant Weapon - Strength, ahead of 2H Agility and
// Revelation), Gloves - Superior Strength (+2.8), Cloak - Agility (+0.5), Enchant Necklace -
// Strength (+2.1), Boots - Greater Agility (+2.2). Forceful Rugged Armor Kits lose to the chest,
// leg and boot enchants.
//
// On 2026-10-13 (the gear round, standards 1b) the set was rebuilt on Forever's budget for a hybrid:
// each armor piece carries slot mod * (1.125 * item level + 2.0) * 1.093 / 1.103 / 1.048 in weighted
// stats (spell power at 0.86, attack power 0.5), 4 or 5 stat lines a piece by the 1.7-norm shape
// check, and the two-hander 90.6. The mix is the Enhancement row of gear_targets.md, hit-adjusted:
// Stamina 19.5%, Strength 9.7%, Agility 2.1%, Intellect 10.2%, Spirit 9.9%, attack power 2.9%,
// spell power 30.5%, crit 4.4% and hit 10.7% (the cap). The armor is 748.5 and the two-hander 90.8
// weighted points; the file has the shares and the shape check per piece. It also carries two
// synthetic item level 71 trinkets (Forever removed the old dungeon trinkets and the sim doesn't
// model its launch ones): a stat trinket (32 Strength, 14 crit rating) and an on-use one (335 attack
// power for 20 sec every 2 min, the weight of Earthstrike). A spell power on-use (140) sims 3.6
// behind at 180 sec. Both trinkets are synthetic, not in the game.
//
// On 2026-10-23 the melee hit cap fell to 8% (standards 4: Forever has no hit suppression), so the
// neck's 10 hit rating became 10 crit rating: 80 hit rating, crit 5.6% of the set. Above 8% only the
// spells gain from hit (about 4 DPS a 1% at 180 sec), and 10 crit rating is worth about 5.5. Level 60
// went from 893.4 / 875.5 / 854.7 to 894.7 / 877.0 / 856.1 at 120 / 180 / 300 sec. The relic is now
// Forever's Totem of the Storm (Lightning Bolt can trigger Maelstrom Weapon at half the chance): with
// Fire Nova's hits a proc it is +0.9 to +2.1 over Burning Totem on Level 60 and Imp GW at every length
// on two seed sets, and within 0.4 on Water Shield. Mana Tide keeps Burning Totem (the Burning Totem
// set below), where Totem of the Storm is 0.1 to 1.5 behind.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Enhancement', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});
// The same set with Burning Totem, for the Mana Tide build.
export const GearLevel60Phase1SyntheticBurningTotem = PresetUtils.makePresetGear('Enh Mana Tide', Level60Phase1SyntheticBurningTotemGearJSON, {
	tooltip: 'The synthetic Phase 1 set with Burning Totem in place of Totem of the Storm, for the Mana Tide build.',
	group: 'Synthetic',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1Synthetic, GearLevel60Phase1SyntheticBurningTotem],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// Generated by tools/rotopt/shaman_presets.sh (the shaman_enh template). Optimized is the
// knob search's best at 60: Stormstrike on its 8 sec cooldown, Flame Shock when its DoT
// is down, Lightning Bolt only at 5 Maelstrom Weapon stacks, and Earth Shock otherwise
// (Earth Shock and the bolt spend the Stormstrike mark for +20%). Rage of the Farseer goes
// on cooldown, with Blood Fury and Berserking waiting for it when that costs them no use.
//
// The air totem is Windfury Totem, because in a raid the shaman puts it down for the
// melee in the group. The shaman gets nothing from it, because Windfury Weapon on the main
// hand turns off the totem's buff for the shaman (see sim/core/totem_weapon_buffs.go). The
// fire totem is Flametongue Totem, for the shaman alone. A weapon holds only one totem
// buff, so the group keeps Windfury Totem's, but the shaman's weapon has room for
// Flametongue Totem's. Fire Nova goes on cooldown after the shocks while the shaman has
// 20% mana, with Mana Spring Totem for the mana. Fire Nova is a spell under Forever that
// goes off around the fire totem already down, with its own 10 sec cooldown.
//
// That is 850.5 / 839.2 / 825.3 at 120 / 180 / 300 sec (2026-10-07). Searing Totem in its
// place was 801.7 / 792.0 / 780.9. A shaman alone does better with Grace of Air in place
// of Windfury Totem (866.1 at 180 sec).
//
// With Windfury Totem and Searing Totem, Rockbiter Weapon was even with Windfury Weapon
// (791.9 against 792.0 at 180 sec). We keep Windfury Weapon because Spirit Weapons turns
// Rockbiter into 30% more threat instead of 30% less.
//
// Fire Nova is the most expensive spell, so it gets the free cast from Clearcasting
// (Elemental Focus). While Clearcasting is up, Fire Nova goes right after Stormstrike,
// ahead of the shocks, and it goes below the 20% floor too. On the synthetic set with
// potions that is 909.1 / 900.1 / 889.9 at 180 / 300 / 600 sec, +0.9 / +4.6 / +6.8 over
// Fire Nova after the shocks alone. Holding the bolt or Earth Shock for the next
// Stormstrike mark lost 2 to 11 at every length, because the mark already lands on about
// 20 of the 22 Stormstrikes in 180 sec.
//
// That was 893 DPS at 300 sec before Clearcasting, on the Phase 2 gear. Grace of Air with
// Flametongue Totem was 828. All the numbers in this paragraph and the one above are from
// before 2026-10-07, when Windfury Weapon still got the level 68 attack power, Windfury
// Totem still stacked with it for the shaman, and Enhancing Totems was still in the sim.
//
// With the raid's Flametongue Totem on our weapon (the user's ruling 2026-10-16, see the raid
// buffs) Magma Totem beats Searing Totem on the Level 60 build, recast with as little as 2 sec
// left: +28.5 / +19.0 / +12.7 at 120 / 180 / 300 sec (Trial by Fire's 20 sec on Magma Totem since
// client 70425 helps), and Earth Shock waits for the Stormstrike mark (+0.3 to +1.5 more), paired at
// 20000 iterations on two seed sets. Level 60 is 906.6 / 886.4 / 863.0 and Imp GW 900.3 / 886.9 /
// 866.9. Windfury Weapon is still the imbue: Rockbiter 28 and Frostbrand 45 behind at 180 sec.
//
// Since the rotation pass of 2026-10-16 Lightning Bolt at 5 Maelstrom Weapon stacks waits for the
// shocks: Level 60 907.2 / 887.0 / 863.2 at 120 / 180 / 300 sec (+0.6 / +0.6 / +0.2, and +1.2 / +0.7
// / +1.1 on the second seed set), Imp GW 901.3 / 887.8 / 867.9 (+1.0 / +0.9 / +1.0, and +0.1 / +0.3
// / +0.6), paired at 20000 iterations. A talent search on it found no build ahead at every length:
// Elemental Focus for Tidal Focus is +4.5 to +8.6 at 180 and 300 sec but 2.2 to 3.4 behind at 120.
export const APLOptimized = PresetUtils.makePresetAPLRotation('Enhancement', OptimizedAPLJSON, {
	tooltip: "The optimizer's best level 60 rotation, with Windfury Totem for the group, Magma Totem and Fire Nova.",
	group: 'Synthetic',
});
// The rotation for the Level 60 + Water Shield and Level 60 + Mana Tide talents. It is
// Optimized with Magma Totem in place of Searing Totem, and a new Magma Totem goes down
// with as little as 10 sec left in the fight. Those builds have no Elemental Focus, so
// there is no Clearcasting line, and Flame Shock now goes first (see below).
//
// Those builds have no points for Improved Fire Nova, and Magma Totem makes up for some of
// it. On the Water Shield build it is 912.0 / 899.7 / 883.4 at 120 / 180 / 300 sec, where
// Searing Totem is 888.8 / 879.5 / 877.0. On the Level 60 build Searing Totem is still
// ahead (938.4 against 927.1 at 180 sec).
//
// Since 2026-10-15 it keeps Lightning Shield up in place of Water Shield, like Optimized (see the
// raid damage note under Options).
//
// Since the rotation auditor's pass 23 (2026-10-16) a new Magma Totem goes down with as little as
// 2 sec left (it was 10): on the Water Shield build 828.9 / 819.3 / 815.7 at 120 / 180 / 300 sec
// (+1.2 / +0.8 / +0.5), on Mana Tide 790.6 / 781.7 / 777.1 (+1.0 / +0.6 / +0.4), paired at 20000
// iterations on two seed sets.
//
// Since the rotation pass of 2026-10-16 the Water Shield build puts Flame Shock up first, then
// Stormstrike, Earth Shock and the bolt, with Fire Nova last: 851.9 / 842.2 / 838.5 at 120 / 180 / 300
// sec (+1.4 / +1.8 / +1.7, and +1.3 / +1.9 / +1.8 on the second seed set). It was within noise on Mana
// Tide, which keeps its own order.
export const APLMagma = PresetUtils.makePresetAPLRotation('Enh Water Shield', MagmaAPLJSON, {
	tooltip: 'The rotation for the Water Shield talents, with Magma Totem and no Clearcasting line.',
	group: 'Synthetic',
});
// Magma Totem with Mana Tide Totem once the shaman is down to 20% mana. Mana Spring Totem
// waits for the tide to end, because putting it down would knock the tide down.
//
// The shaman rarely runs that low, so it is worth little. On the Mana Tide build it is
// +2.5 at 300 sec and even at 120 and 180 sec.
export const APLManaTide = PresetUtils.makePresetAPLRotation('Enh Mana Tide', ManaTideAPLJSON, {
	tooltip: 'The Enh Water Shield rotation that also drops Mana Tide Totem at 20% mana.',
	group: 'Synthetic',
});

// Level 30 Leveling is the leveling optimizer's best rotation for a solo level 30 grinding
// mobs of its level (tools/rotopt, the shaman_enh_leveling template, see
// docs/leveling_sim.md). It needs the Leveling mode: it rests, looks for the next mob and
// pulls it with the leveling actions, which do nothing in a single fight. A better rotation
// from a later optimizer run can replace apls/level30_leveling.apl.json under the same name.
//
// What it does is in docs/leveling_sim.md (step 5, pass 9): it rests to 90% health and 80%
// mana and pulls with Flame Shock. While still fighting it stands and looks for the next mob,
// from 70% health (50% when its target is nearly dead). It keeps Mana Spring and Strength of
// Earth Totem down, and with two mobs puts down Searing Totem and casts Fire Nova. It heals
// with Lesser Healing Wave below 30% health.
export const APLLevel30Leveling = PresetUtils.makePresetAPLRotation('Level 30 Leveling', Level30LevelingAPLJSON, {
	tooltip: "The leveling optimizer's best rotation for a solo level 30. It rests, finds the next mob and pulls it, for the Leveling mode.",
	group: 'Leveling',
});

// The Leveling Plan rotations are the leveling optimizer's best rotations at levels 10, 15, 20,
// 25 and 30, each searched on the leveling gear plan's gear at that level with the plan's
// talents there (the stat weights work of 2026-10-16, see EPLevelingGearPlan). Below 30 they
// use Rockbiter Weapon. They are all the same rotation, with Flame Shock pulls, Searing Totem
// from two mobs and Fire Nova, but they rest, heal and chain pull at other health and mana.
// The level 10 one pulls with the rank of Healing Wave a level 10 has.
export const APLLevelingPlan = (
	[
		[10, LevelingPlan10APLJSON],
		[15, LevelingPlan15APLJSON],
		[20, LevelingPlan20APLJSON],
		[25, LevelingPlan25APLJSON],
		[30, LevelingPlan30APLJSON],
	] as Array<[number, any]>
).map(([level, json]) =>
	PresetUtils.makePresetAPLRotation(`Level ${level} Leveling Plan`, json, {
		tooltip: `The leveling optimizer's best rotation for a level ${level} in the leveling plan's gear and talents.`,
		group: 'Leveling Plan',
	}),
);

export const APLPresets = {
	[Phase.Phase1]: [APLOptimized, APLMagma, APLManaTide, APLLevel30Leveling, ...APLLevelingPlan],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent search's best on the Phase 2 gear (tools/rotopt -talent-search) with Spirit
// Weapons pinned (-require spiritWeapons): the sim scores no threat, but the 30% cut is a
// must in a raid. Shamanistic Focus, Rage of the Farseer and Mental Quickness are in.
// Improved Stormstrike is the mana talent: with no Mana Spring the shaman is dry at five
// minutes. The dodge/parry reset half never fires on a boss the shaman is behind. Rank 2
// is worth 17 DPS at 300 sec, measured after the 2026-10-02 fix that took its regen from
// 100% down to the client's 50%.
//
// The 2026-09-24 beta build moved Elemental Fury to row 6, out of an enhancement
// shaman's reach, and took the cast speed off Rage of the Farseer. Its 5 points went to
// Reverberation 5 and Ancestral Knowledge 5. That is 821 DPS at 300 sec where the old
// build was 834.
//
// That leaves one point with no DPS in it. The search breaks such ties on mana left at
// the end of the fight, and puts it in Elemental Focus: 0.7 DPS behind (inside the noise)
// with 420 more mana at 300 sec. Convection is 250 more and Elemental Alacrity nothing,
// since the rotation only casts Lightning Bolt at 5 Maelstrom Weapon stacks.
//
// Since Fire Nova became a spell with its own cooldown, Improved Fire Nova is worth 2
// points. An earlier hand edit took them out of Concussion (030533102-...), but that left
// 3 points in the first row, and Reverberation needs 5 above it. That build cannot be
// trained in game.
//
// On 2026-10-05 we reran the full talent search on the Phase 2 gear and the Optimized
// rotation, from four starting builds. Three of them ended on this one: Convection 5 in
// place of Concussion, Improved Fire Nova 2, and Ancestral Knowledge 3 to pay for it.
// Convection beats Concussion because the shaman runs low on mana and the cheaper shocks
// and Fire Novas are worth more than 5% more damage on them. It is 929.9 / 897.4 / 886.0
// at 120 / 300 / 480 sec. The last legal build (0505331-055030031005112251) was 898.1 /
// 881.9 / 873.1, and Concussion 5 with the same points elsewhere was 1 to 7 behind.
//
// On 2026-10-12 we ran the search again on the synthetic Phase 1 set and the current sim, as the
// mean of 120, 180 and 300 sec, and it swapped Convection 5 for Concussion 5 (+3.0, 795.7 at
// 20000 iterations). The mana the sim gives now (Spirit on the synthetic set, Mana Spring Totem)
// is enough without the cheaper shocks. A tie-break on mana that moves 2 points from Elemental
// Weapons to Ancestral Knowledge is 0.8 behind. The knob search kept the Optimized rotation. On
// the standard gear and consumes the search only made such tie-breaks on mana (795.9 against
// 796.1), so the build stays: 803.2 / 791.9 / 771.1 at 120 / 180 / 300 sec. On the 2026-10-13
// gear round's set the search again only made mana tie-breaks inside the noise (792.0 against
// 792.6), so the build stays: 813.5 / 801.8 / 781.6, and 813.8 / 801.6 / 782.3 once Rage of the Farseer
// waits for the on-use trinket, and 824.5 / 813.9 / 794.0 with Elixir of Cunning and Ferocity and Forever's oil and chili damage.
//
// Client 1.60.1.70425 (2026-10-16) cut Fire Nova (coefficient 0.143 from 0.214, rank 5 232 to 259
// base from 397 to 443) and moved the Elemental tree around. The search (Spirit Weapons pinned, as
// the mean of 120, 180 and 300 sec, 2 raid hits a minute) dropped Elemental Focus and a point of
// Improved Stormstrike for Tidal Focus 2 (Clearcasting Fire Novas are worth little now), and moved
// Elemental Weapons' points to Ancestral Knowledge: +7.9 / +6.3 / -7.2 at 120 / 180 / 300 sec
// paired against the old build (it runs lower on mana in long fights). With the raid's second
// shaman's Grace of Air (standards 25) it is 835.5 / 825.3 / 809.1, from 865.7 / 852.9 / 834.8.
//
// With the raid's Flametongue Totem and Lightning Bolt after the shocks (2026-10-16) the talent search
// on two seed sets put Elemental Focus back in place of a Tidal Focus point. Presets are tuned at 180
// sec, so it ships though it costs a little at 120: 905.0 / 892.2 / 871.8 at 120 / 180 / 300 sec
// (-2.2 / +5.2 / +8.6, and -3.4 / +4.5 / +7.9 on the second seed set), paired at 20000 iterations.
// Improved Stormstrike 2 as well was further ahead at 300 sec but behind at 180. The knob search on
// the new talents kept the Optimized rotation.
//
// The regear round (2026-10-23: the 8% melee cap and the neck's crit, Totem of the Storm, Fire Nova's
// hits a proc, Earth Shock without the Stormstrike hold) moved the Tidal Focus point to Improved
// Stormstrike 2 (Spirit Weapons pinned). Judged at 180 sec, paired at 20000 iterations on seeds 1 and
// 1000003: +1.2 / +1.8 at 180 sec and +5.0 / +4.7 at 300, -3.2 / -2.3 at 120 (the mana from Improved
// Stormstrike pays off in longer fights). The knob search on the new talents kept the rotation. Level
// 60 is 894.1 / 880.1 / 863.1 at 120 / 180 / 300 sec. Unpinned, the search drops Spirit Weapons (a
// threat talent the sim doesn't score) for that second Improved Stormstrike point instead: +5.3 at 180.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Enhancement', SavedTalents.create({ talentsString: '053503102-055030011005112251' }), {
	tooltip: 'Level 60 talents from the talent search, with Spirit Weapons and Concussion.',
	group: 'Synthetic',
});
// The same search with Improved Ghost Wolf 2 pinned as well (-require
// spiritWeapons,improvedGhostWolf). The two points come out of Ancestral Knowledge, and
// the spare point goes back into it (4) for 1.3 DPS over Elemental Focus. The build is
// 1.2 DPS behind Level 60 at 300 sec.
//
// After the 2026-10-05 search it is the Level 60 build with the 2 points for Improved
// Ghost Wolf taken out of Ancestral Knowledge (1). The search with both talents pinned
// found nothing better. It is 894.0 at 300 sec, 3.4 behind Level 60.
//
// On 2026-10-12 it takes Concussion 5 in place of Convection 5 like Level 60 (+2.8). It is
// 810.4 / 797.7 / 778.7 at 120 / 180 / 300 sec on the 2026-10-13 gear round's set with Rage of the
// Farseer waiting for the on-use trinket (820.9 / 809.4 / 790.0 with Elixir of Cunning and Ferocity and Forever's oil and chili damage),
// 4.1 behind Level 60 at 180 (800.0 / 788.3 / 767.3 before).
//
// Under client 1.60.1.70425 the search with Improved Ghost Wolf and Spirit Weapons pinned (from the
// Level 60 changes above) kept Elemental Focus and Ancestral Knowledge 4: 827.4 / 818.4 / 811.5 at
// 120 / 180 / 300 sec (862.0 / 848.7 / 830.7 before), 6.9 behind Level 60 at 180.
//
// In the regear round (2026-10-23) the search with both pinned made the same move as Level 60, Tidal
// Focus 1 to Improved Stormstrike 2: +1.0 / +1.9 at 180 sec, +4.9 / +3.9 at 300 and -2.8 / -2.9 at 120 on
// two seed sets. 890.4 / 876.2 / 858.6 at 120 / 180 / 300 sec with the Optimized rotation, 3.9 behind
// Level 60 at 180.
export const TalentsLevel60ImpGW = PresetUtils.makePresetTalents('Enh Imp GW', SavedTalents.create({ talentsString: '053503102-054032001005112251' }), {
	tooltip: 'The Enhancement talents with Improved Ghost Wolf 2, taken out of Ancestral Knowledge.',
	group: 'Synthetic',
});
// The best build that takes Water Shield (-require spiritWeapons,waterShield), with the
// Magma Totem rotation. Water Shield needs 10 Restoration points above it, and those go
// to Tidal Focus 5 (+5% hit) and Totemic Focus 5 (cheaper totems). That leaves 9 points
// for Elemental, which is not enough for Improved Fire Nova. So we take Convection 5,
// Call of Flame 1 and Elemental Devastation 3, and Improved Stormstrike drops to 1.
//
// It is 912.0 / 899.7 / 883.4 at 120 / 180 / 300 sec, where Level 60 is 950.1 / 938.4 /
// 923.1. Water Shield itself adds nothing in the Level 60 encounter, because it only
// gives mana when something hits the shaman.
//
// The first search ran with Searing Totem and ended on Concussion 5 and Reverberation 1.
// Magma Totem was 5 ahead on that build, and the search with Magma Totem moved the points
// to Convection and Call of Flame.
//
// On 2026-10-12 the search ran again on the current sim, with the Water Shield requirement and the
// Magma Totem rotation, as the mean of 120, 180 and 300 sec. It moved Elemental Devastation's
// points to Call of Flame 3 and Convection to Concussion, and one point from Tidal Focus to
// Mindfulness: 758.6 from 749.9. With Fire Nova ahead of the shocks and the 1 sec Earth Shock
// hold in the Magma Totem rotation it was 773.5 / 764.5 / 748.9 at 120 / 180 / 300 sec. On the
// standard gear and consumes it is 758.3 / 750.8 / 738.9, 5.2% behind Level 60 at 180. The search
// there only moved 2 points from Elemental Weapons to Ancestral Knowledge on a mana tie-break
// (757.3 against 758.6), so the build stays. On the 2026-10-13 gear round's set the same tie-break
// was 752.8 against 751.2, inside the noise, so it stays: 763.1 / 755.6 / 741.4, and 762.9 / 755.4 /
// 742.3 with Rage of the Farseer waiting for the on-use trinket (772.0 / 765.5 / 753.2 with Elixir
// of Cunning and Ferocity), 5.9% behind Level 60 at 180.
//
// Client 1.60.1.70425 moved Tidal Focus up to the first row and Totemic Focus down to the second,
// so the old Restoration points (Tidal Focus 3, Mindfulness 2, Totemic Focus 5) no longer reach
// Water Shield. The search (Water Shield and Spirit Weapons pinned) settled on Tidal Focus 5,
// Mindfulness 3 and Totemic Focus 2, and one point from Elemental Weapons into Ancestral
// Knowledge: 803.7 / 795.3 / 793.3 at 120 / 180 / 300 sec (813.8 / 804.2 / 792.4 before), 3.6%
// behind Level 60 at 180. Casting Water Shield, whose globes now also come from our spell crits,
// is still 3 to 11 behind Lightning Shield on this build and on Mana Tide.
export const TalentsLevel60WaterShield = PresetUtils.makePresetTalents(
	'Enh Water Shield',
	SavedTalents.create({ talentsString: '051003-054030021005112151-053020001' }),
	{ tooltip: 'Level 60 talents with 10 Restoration points for Water Shield.', group: 'Synthetic' },
);
// The best build that takes Water Shield and Mana Tide Totem (-require
// spiritWeapons,waterShield,manaTideTotem), with the Mana Tide rotation. Mana Tide needs
// 15 Restoration points above it. Next to Tidal Focus 5 and Totemic Focus 5 we put
// Mindfulness 3 and Improved Reincarnation 1 there, and they do nothing in the sim. With
// 16 points in Restoration and 31 to reach Rage of the Farseer, only 4 are left. They go
// to Concussion 2 and Ancestral Knowledge 5.
//
// It is 883.8 / 873.7 / 866.0 at 120 / 180 / 300 sec. Elemental Devastation is worth a
// lot (27.8 on the Water Shield build), but its 8 points cost Rage of the Farseer and
// Maelstrom Weapon points, and every such build we tried was 25 to 35 behind.
//
// On 2026-10-12 the search on the current sim (Mana Tide Totem and Water Shield pinned) found
// nothing better: 750.4 / 743.9 / 737.9 at 120 / 180 / 300 sec with the Mana Tide rotation. On
// the standard gear and consumes it was 735.5 / 728.8 / 724.6. On the 2026-10-13 gear round's set
// the search found nothing better: 739.7 / 733.0 / 727.4, and 739.9 / 733.4 / 728.4 with Rage of the
// Farseer waiting for the on-use trinket (749.3 / 742.7 / 737.9 with Elixir of Cunning and Ferocity and Forever's oil and chili damage),
// 8.7% behind Level 60 at 180.
//
// Under client 1.60.1.70425 the search (Mana Tide Totem and Spirit Weapons pinned) moved a point
// from Improved Stormstrike to Concussion: 769.0 / 760.5 / 756.7 at 120 / 180 / 300 sec (790.3 /
// 782.1 / 778.0 before, mostly the Fire Nova cut), 7.9% behind Level 60 at 180.
export const TalentsLevel60ManaTide = PresetUtils.makePresetTalents(
	'Enh Mana Tide',
	SavedTalents.create({ talentsString: '03-055030031005112051-053051001001' }),
	{ tooltip: 'Level 60 talents with Water Shield and Mana Tide Totem.', group: 'Synthetic' },
);

// The talents the leveling optimizer ran with: level 30 with 5 extra talent points, with
// Stormstrike, Spirit Weapons and Improved Ghost Wolf 2.
export const TalentsLevel30Leveling = PresetUtils.makePresetTalents('Level 30 Leveling', SavedTalents.create({ talentsString: '-0540320010051122' }), {
	tooltip: 'Level 30 with 5 extra talent points, which the leveling optimizer ran with.',
	group: 'Leveling',
});

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60ImpGW, TalentsLevel60WaterShield, TalentsLevel60ManaTide, TalentsLevel30Leveling],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// The raid damage in the Level 60 encounter (PresetUtils.LEVEL_60_RAID_DAMAGE, the same for every
// spec since 2026-10-16): 2 direct hits a minute, at random times. Nothing we can read measures
// it (Warcraft Logs was out of reach, 2026-10-15), so it comes from Season of Discovery boss
// timers in the Deadly Boss Mods source. Most bosses land one or two hits on the raid every 20 to
// 40 sec (Ragnaros's Wrath every 26 sec, Broodlord's Blast Wave, Kel'Thuzad's Frostbolt Volley,
// Chromaggus's breaths), Patchwerk and Loatheb none, and Firemaw, Vaelastrasz and Huhuran's last
// 30% about 10. The DPS report's assumptions run 0 and 10 too.
//
// Each hit fires a Lightning Shield orb, so every Level 60 rotation keeps the shield up: before
// the pull, and in the fight when nothing else is ready (+2.8 at 6 a minute over putting it
// first). At 180 sec Level 60 went from 826.0 to 833.8, Imp GW from 822.0 to 829.5 (about +26 at
// 10 a minute), and with no hits it costs nothing measurable. The Water Shield and Mana Tide builds
// cast Lightning Shield too: Water Shield's 2% mana a globe was worth less than Lightning Shield's
// orbs at every rate we tried (0, 2, 6 and 10 a minute), whether we kept Water Shield for low mana
// or not. Water Shield went from 776.1 to 785.3, Mana Tide from 753.4 to 763.3 (+32 and +34 at 10
// a minute).
//
// Since the rotation auditor's pass 21 (2026-10-16) Optimized puts Lightning Shield up at -3 sec
// and Searing Totem at -1.5 sec, so the totem no longer takes the first GCD of the fight, and
// none of the rotations recasts Strength of Earth, Windfury or Mana Spring with less than 20 sec
// left (the starting totems run out a few seconds before a 5 min fight ends). Level 60 is 845.7 /
// 834.6 / 815.5 at 120 / 180 / 300 sec (+1.2 / +0.8 / +1.9), Imp GW 842.1 / 830.3 / 811.0, Water
// Shield 792.6 / 785.4 / 772.5 and Mana Tide 770.0 / 763.3 / 758.0 (the last two only gain at 300).
export const DefaultOptions = EnhancementShamanOptions.create({
	shamanImbue: WeaponImbue.WindfuryWeapon,
	syncType: ShamanSyncType.Auto,
});

// The class options every preset build sets.
//
// Since 2026-10-16 the totems go down in each rotation's prepull actions, one global cooldown
// apart, with their mana. Before that the builds started the fight with them already down for
// free (the old StartingTotems option), which the game doesn't allow.
const PresetBuildOptions = { shamanImbue: WeaponImbue.WindfuryWeapon };

// The level 60 consumables. On 2026-10-08 we tried every consumable the sim applies in each
// slot on the Level 60 build at 180 sec. Under Forever the shaman imbue stacks with an oil or a
// stone, and Shadow Oil was best there (+19.6, Elemental Sharpening Stone was +10). Goblin Land
// Mine (+4.1, dropped 2026-10-23, below) and Elixir of Shadow Power (+1.2, for Shadow Oil) were
// also worth it. Grilled Squid
// (1% crit under Forever) is +3.4 over Blessed Sunfruit.
//
// On 2026-10-12 Forever's Phase 1 consumables came in (the shared consumables branch), tried the
// same way as the mean of 120, 180 and 300 sec: Elixir of Nature Power +4.5 (Lightning Bolt and
// the shocks), Greater Mageblood Elixir +1.6, Elixir of Greater Intellect and Greater Agility in
// place of Wisdom and the Mongoose. A character takes one of the five 2% crit elixirs (the
// standard): Elixir of the Sages is best (815.8 on the earlier gear, the Owl 815.6, the Grizzly
// 814.0, the Mongoose 811.1). If Forever lets them stack, the Mongoose, the Grizzly and the Owl on
// top would add about 36. Major Frenzy Potion, Mightfish Steak and Flank au Poivre lose.
//
// On 2026-10-13 the consumables auditor's per-slot pass moved two slots (paired, 20000 iterations):
// Elixir of Cunning (25 Agility, 25 Intellect, no crit) in place of Greater Agility, +10.2 at 180
// sec, and Elixir of Ferocity (18 Strength, 18 Agility) in place of Juju Power, +1.2 more. Both gain at
// 120, 180 and 300 sec on all four Level 60 builds.
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfCunning,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MajorManaPotion,
	defaultConjured: Conjured.ConjuredDemonicRune,
	dragonBreathChili: true,
	// No filler explosive (standards 11, 2026-10-23: explosives only when they gain). The Goblin Land
	// Mine's 1 sec cast resets the swing timer, so dropping it is +8.0 / +8.7 at 180 sec on seeds 1 and
	// 1000003 for Level 60 (Imp GW +8.1 / +8.7, Mana Tide +7.9 / +7.6, Water Shield +8.4 / +8.8, and
	// +6 to +10 at 120 and 300 sec on all four). The EZ-Thro Dark Bomb ties the Land Mine (the same 450
	// average and the same cast) and the EZ-Thro Thorium Grenade is 0.8 behind it. The Goblin Sapper
	// Charge needs no cast and stays (+3.6 at 180 sec on Level 60).
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	// The fight is in a natural flask zone (standards 11), but the best natural flask, Precision, is 6
	// behind the Flask of Supreme Power at 180 sec (2 to 6 with Mana Tide or Water Shield; Aggression 13
	// to 16, Accuracy and Swiftness 20 to 36 behind; rechecked 2026-10-23). The Dragonfire Potion is 20
	// behind the Major Mana Potion, 3 to 4 ahead at 180 sec with Mana Tide or Water Shield but 6 to 8
	// behind at 300.
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodGrilledSquid,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.ShadowOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	naturePowerBuff: NaturePowerBuff.ElixirOfNaturePower,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	shadowPowerBuff: ShadowPowerBuff.ElixirOfShadowPower,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfSages,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.ROIDS,
});

// The raid the level 60 numbers were measured under (the rogue search's export minus
// the totems): both paladin blessings, Judgement of Wisdom and every debuff, but no
// second shaman in the group, so the player's own totems are what the rotation casts.
// Judgement of Wisdom alone is 60 DPS here because without it the shaman is out of mana
// two minutes in. Grace of Air comes from the raid's second shaman, the one whose Grace of
// Air every melee group gets (the user's ruling 2026-10-16, "both totems": client
// 1.60.1.70425 lets it stack with Windfury Totem and makes both raid-wide).
//
// That second shaman's fire slot holds Flametongue Totem (the user's ruling 2026-10-16: "Let's
// assume it works, that it doesn't interfere with other melee getting Windfury Totem, and that
// shamans with their WF imbue get FTT on top"). Our own Windfury Weapon turns our own Windfury
// Totem's buff off for us, so the weapon's totem slot holds the raid's Flametongue Totem; the
// other melee keep Windfury Totem.
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
	totemWeaponBuff: TotemWeaponBuff.TotemWeaponBuffFlametongue,
	leaderOfThePack: true,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
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

// The shaman is an engineer and a leatherworker (an alchemist until 2026-10-06). Engineering
// is for the Goblin Sapper in the consumes (+3.6 at 180 sec on Level 60) and the goggles. Gear
// another profession makes is only in the presets when it binds on equip, so it can be bought.
// The gear searches before that date ran without the leatherworking gear that binds on pickup.
//
// The distance and the channel clip delay are the ones the DPS Report sims with.
export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
	race: Race.RaceOrc,
	distanceFromTarget: 5,
	channelClipDelay: 50,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor), a 3 minute
// fight, and the default buffs, debuffs and consumes. It is here so switching back from
// the Level 20 build restores everything that build changed. The synthetic Phase 1 set and
// its talents were searched at 180 sec.
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
// On 2026-10-08, with the starting totems and the new consumables, Level 60 is 826.5 / 819.6 /
// 807.7 at 120 / 180 / 300 sec, Level 60 + Water Shield 776.6 / 767.0 / 751.2 and Level 60 +
// Mana Tide 756.6 / 750.3 / 745.3. A knob search kept every rotation knob but the Mana Tide
// build's Fire Nova order (see shaman_presets.sh). It wanted Grace of Air in place of Windfury
// Totem (+23 at 180 sec), but the builds keep Windfury Totem for the group. Standards 23 (user ruling,
// 2026-10-15): there is no totem duty for shamans, every shaman drops its own party's totems (Strength
// of Earth, Windfury Totem and Mana Spring for this melee party), like a warrior's Sunder. The talents were
// not searched again.
//
// On 2026-10-12, after the shaman audit fixes since 2026-10-08 (the sim had moved Level 60 to
// 805.6 / 795.7 / 778.7), the talent and knob searches ran again (see the talent presets and
// shaman_presets.sh). They ran once more on the standard gear and Phase 1 consumables and kept
// every build. There (20000 iterations) Level 60 is 803.2 / 791.9 / 771.1 at 120 / 180 / 300
// sec, Level 60 + Imp GW 800.0 / 788.3 / 767.3, Level 60 + Water Shield 758.3 / 750.8 / 738.9
// and Level 60 + Mana Tide 735.5 / 728.8 / 724.6.
//
// Those are after the shared change that leaves out the Ahn'Qiraj book ranks under Forever:
// Strength of Earth Totem IV (42 Strength, was 53) and the raid's Battle Shout VI. It costs every
// build about 7.7 DPS (Level 60 was 811.0 / 799.6 / 778.6). The talent and knob searches ran again
// on it and kept every build, with only mana tie-breaks inside the noise.
//
// On 2026-10-13 the gear round (standards 1b, with the synthetic trinkets) moved them: Level 60 is
// 813.5 / 801.8 / 781.6, Level 60 + Imp GW 810.3 / 797.9 / 778.0, Level 60 + Water Shield 763.1 /
// 755.6 / 741.4 and Level 60 + Mana Tide 739.7 / 733.0 / 727.4 (20000 iterations). The knob search
// moved the Optimized rotation's fire totem from Flametongue Totem to Searing Totem (+5.9 at 180
// sec; Flametongue Totem holds the weapon buff slot Windfury Totem needs), and kept every other knob.
//
// Since then (the rotation auditor's cooldown pass) Rage of the Farseer waits for the on-use trinket
// when the wait costs it no use, in all three Level 60 rotations. Paired at 20000 iterations it is
// +0.7 to +1.3 at 300 sec and within ±0.2 at 120 and 180 (standard error 0.4): Level 60 is now
// 813.8 / 801.6 / 782.3, Level 60 + Imp GW 810.4 / 797.7 / 778.7, Level 60 + Water Shield 762.9 /
// 755.4 / 742.3 and Level 60 + Mana Tide 739.9 / 733.4 / 728.4.
//
// Then the consumables auditor's pass put in Elixir of Cunning and Elixir of Ferocity (see
// DefaultConsumes): Level 60 was 823.8 / 813.0 / 793.4. With the shared consumables branch's Forever
// damage for Shadow Oil (63 to 71) and Dragonbreath Chili (61 to 69) Level 60 is 824.5 / 813.9 /
// 794.0, Level 60 + Imp GW 820.9 / 809.4 / 790.0, Level 60 + Water Shield 772.0 / 765.5 / 753.2 and
// Level 60 + Mana Tide 749.3 / 742.7 / 737.9.
//
// Then the two-hander took the slowest Era Phase 1 BiS two-hand speed (standards 22): Spinal Reaper's
// damage per second at Dreadforge Retaliator's 3.7 sec, still an axe. On 2026-10-15 (20000 iterations)
// Level 60 is 834.8 / 823.8 / 803.7, Level 60 + Imp GW 831.3 / 819.8 / 799.9, Level 60 + Water Shield
// 783.2 / 776.1 / 763.2 and Level 60 + Mana Tide 760.0 / 753.4 / 748.2.
//
// The Optimized rotation's Fire Nova then took a mana rule in place of its 20% floor: outside
// Clearcasting it needs 12 mana for every second left in the fight, so the pool is spent by the end
// (the rotation auditor's pass 19). Paired at 20000 iterations on the DPS report requests (with the
// raid's Trueshot Aura): Level 60 835.2 / 826.0 / 805.4 (+0.4 / +2.2 / +1.7) and Level 60 + Imp GW
// 831.8 / 822.0 / 800.2 (+0.5 / +2.2 / +0.3). 14 and 16 a second lose on Imp GW at 300 sec, 20 and
// up lose 6 to 10. The Magma Totem rotation keeps its floor (at most +0.35, -0.86 at 300 sec).
//
// The races on Level 60 at 180 sec (20000 iterations, standard error 0.4): Orc 823.8, Windshaper
// Skyborne 810.0, Tauren 809.4, Troll 809.1, Dwarf (the only Alliance shaman) 804.8. Best overall and
// best Horde: Orc. Best Alliance: Dwarf. The Orc's lead is Axe Specialization. A 3.7 sec mace of the
// same damage per second would give the Dwarf its 1% instead (812.5). Orc wins the Imp GW, Water
// Shield and Mana Tide builds too (sim-status.md has the older tables).
//
// The regear round (2026-10-23, 20000 iterations, seeds 1 and 1000003, the natural flask zone on): the
// neck's 10 hit rating over the new 8% cap went to crit, Totem of the Storm took the relic slot (Mana Tide
// keeps Burning Totem), Earth Shock no longer waits for the Stormstrike mark, and Level 60 and Imp GW took
// Improved Stormstrike 2 for Tidal Focus 1. The Water Shield and Mana Tide searches kept their talents and
// rotations. At 120 / 180 / 300 sec, from the old presets under the new sim: Level 60 893.4 / 875.5 / 854.7
// to 894.1 / 880.1 / 863.1, Imp GW 889.4 / 871.2 / 850.8 to 890.4 / 876.2 / 858.6, Water Shield 841.2 /
// 829.9 / 824.9 to 843.3 / 831.6 / 826.7, Mana Tide 802.4 / 790.5 / 784.4 to 803.6 / 792.2 / 785.6. The
// Major Mana Potion stays: the Dragonfire and Caustic Smog Potions are +3 to +6 at 120 and 180 sec on Water
// Shield and Mana Tide but 6 to 8 behind at 300, and 20 behind on Level 60. The best natural flask,
// Precision, is 2 to 6 behind the Flask of Supreme Power.
//
// The races at 180 sec (seed 1; seed 1000003 agrees within 0.8): Level 60 Orc 880.1, Windshaper Skyborne
// 878.1, Tauren 872.0, Dwarf 871.6, Troll 868.7; Imp GW 876.2, 874.1, 867.7, 867.6, 864.6; Water Shield
// 831.6, 828.0, 823.1, 823.0, 820.3; Mana Tide 792.2, 790.8, 784.1, 783.7, 782.2. Best overall and best
// Horde: Orc on every build (1.3 to 3.5 ahead of Windshaper Skyborne on both seed sets). Best Alliance:
// Dwarf, the only Alliance shaman.
// Race re-pick (2026-10-27, standards 29 and 16, on the 10-26 racial pass with each race's own rotation
// and gear): Level 60 and Level 60 + Imp GW move to the Windshaper Skyborne, 882.0 and 878.0 at 180 sec
// against the Orc's 880.5 and 876.4 (verify.py, 20000 iterations, seed 1). Water Shield and Mana Tide stay
// Orc.
// Second race re-pick (2026-10-27, standards 29 and 16, after dev notes post 8: Orc Blood Fury 2 min, Troll
// Berserking 15 sec, Undead Touch of the Grave 15 sec): Level 60 and Level 60 + Imp GW move back to the Orc,
// 884.2 and 880.1 at 180 sec against the Windshaper's 882.0 and 878.0 (verify.py, 20000 iterations, seed 1).
// Every Enhancement preset is Orc again.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Enhancement', {
	group: 'Synthetic',
	tooltip: 'Level 60 Orc on the synthetic Phase 1 gear with the Enhancement rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLOptimized,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptions,
});
export const PresetBuildLevel60ImpGW = PresetUtils.makePresetBuild('Enh Imp GW', {
	group: 'Synthetic',
	tooltip: 'The Enhancement build with Improved Ghost Wolf 2, taken out of Ancestral Knowledge.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60ImpGW,
	rotation: APLOptimized,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptions,
});
export const PresetBuildLevel60WaterShield = PresetUtils.makePresetBuild('Enh Water Shield', {
	group: 'Synthetic',
	tooltip: 'Level 60 with the Water Shield talents and rotation.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60WaterShield,
	rotation: APLMagma,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptions,
});
export const PresetBuildLevel60ManaTide = PresetUtils.makePresetBuild('Enh Mana Tide', {
	group: 'Synthetic',
	tooltip: 'Level 60 with the Water Shield and Mana Tide talents and rotation.',
	gear: GearLevel60Phase1SyntheticBurningTotem,
	talents: TalentsLevel60ManaTide,
	rotation: APLManaTide,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptions,
});

// A level 30 shaman soloing Interrogator Vishas (Scarlet Monastery Graveyard, level 32)
// for 120 sec, with no raid around. The boss hits the shaman, so Lightning Shield gets to fire, and
// Blessing of Kings is on because a group at that level almost always has a paladin in it.
// Windfury Weapon (level 30) is the shaman imbue in
// the build's options, 25 DPS ahead of Rockbiter and 24 ahead of Flametongue Weapon (with
// Flametongue Totem's buff on top of either, see core/totem_weapon_buffs.go).
//
// The consumes are the best a level 30 Engineer and Alchemist can use, every profession
// at 225. A stat elixir and the Scroll of the same stat do not stack, so we take the
// higher one. Each line below is what the item is worth at 60 / 300 sec, measured by
// taking it out of the full set:
// - Elixir of Agility +3.8 / +3.0. Scroll of Agility II is only +1.5.
// - Scroll of Strength II +2.4 / +2.3. Elixir of Ogre's Strength and Elixir of Giant
//   Growth (8 Strength) are 0.2 behind.
// - Lesser Arcane Elixir +3.7 / +2.6 and Elixir of Fire Power +1.3 / +0.9.
// - Lesser Mageblood Elixir +0.0 / +2.4.
// - Scroll of Intellect II +2.0 / +2.6 and Scroll of Spirit II +0.0 / +1.4. Elixir of
//   Wisdom gives 6 Intellect and loses to the scroll's 8.
// - Briny Seafood Stew (14 spell damage). Barbecued Buzzard Wing (10 Intellect) is 1.0
//   behind at 60 sec and 1.0 ahead at 300.
// - Dragonbreath Chili +4.0 / +3.6. It stacks with the Well Fed food.
// - Shadow Oil on the weapon, +2.0 / +2.7 over Lesser Wizard Oil.
// - Goblin Sapper Charge +3.1 / +0.8. It shares the 1 minute explosive cooldown, so the
//   Goblin Land Mine only adds on fights longer than a minute (+3.9 at 300 sec).
// - Mana Potion. Spellblasting Potion (+16 spell damage for 30 sec) is 0.6 behind at 60
//   sec next to Briny Seafood Stew, because the stew already leaves the shaman short on
//   mana by the end, and 8 behind at 300.
// The set takes the build from 194.1 / 182.4 / 133.9 to 240.1 / 222.7 / 169.7 at 30 / 60
// / 300 sec. These were measured on the 60 sec gear and rotation. At 120 sec the same
// picks hold: Spellblasting Potion is 4.2 behind the Mana Potion, Lesser Wizard Oil is 2.4
// behind Shadow Oil, the Goblin Land Mine is worth 3.3, and Barbecued Buzzard Wing ties
// Briny Seafood Stew.
export const EncounterLevel30 = PresetUtils.makePresetEncounter(
	'Level 30',
	Encounter.create({
		duration: 120,
		durationVariation: 5,
		executeProportion20: 0.2,
		executeProportion25: 0.25,
		executeProportion35: 0.35,
		targets: [
			{
				id: 3983,
				name: 'Interrogator Vishas',
				level: 32,
				mobType: MobType.MobTypeHumanoid,
				stats: new Stats().withStat(Stat.StatArmor, 1063).withStat(Stat.StatHealth, 30000).asArray(),
				minBaseDamage: 52,
				damageSpread: 0.3333,
				swingSpeed: 2,
				parryHaste: true,
				tankIndex: 0,
			},
		],
	}),
	{
		tanks: [UnitReference.create({ type: UnitReference_Type.Player, index: 0 })],
		raidBuffs: RaidBuffs.create({ scrollOfIntellect: true, scrollOfSpirit: true }),
		debuffs: Debuffs.create({}),
		buffs: IndividualBuffs.create({ blessingOfKings: true }),
		consumes: Consumes.create({
			agilityElixir: AgilityElixir.ElixirOfAgility,
			defaultPotion: Potions.ManaPotion,
			dragonBreathChili: true,
			fillerExplosive: Explosive.ExplosiveGoblinLandMine,
			firePowerBuff: FirePowerBuff.ElixirOfFirepower,
			food: Food.FoodBrinySeafoodStew,
			mainHandImbue: WeaponImbue.ShadowOil,
			manaRegenElixir: ManaRegenElixir.LesserMagebloodElixir,
			sapperExplosive: SapperExplosive.SapperGoblinSapper,
			spellPowerBuff: SpellPowerBuff.LesserArcaneElixir,
			strengthBuff: StrengthBuff.ScrollOfStrength,
		}),
	},
);
// Stat weights for leveling (2026-10-16): what each stat adds to the mobs a solo level 30
// kills in a 20 min session, in attack power, on the Level 30 Leveling (solo) build. One
// attack power adds 0.042 mobs a session.
//
// Each stat is a direct probe at 20,000 sessions, +10 of it (+5 MP5, +20 spell power, +200
// armor, +1% hit or crit) against none, with the same random rolls. Under Forever a bonus to
// melee hit or crit moves the spell one too, so +1% is one bonus, not two. We take the step up,
// because that's what an upgrade does. Most stats come out the same stepping down, but
// Stamina doesn't. -10 Stamina costs 1.45 attack power a point and +10 adds 0.86, and +30 adds
// no more than +20, because Stamina only shortens our rests and we rest to a fixed health.
// The weapon DPS and the fire and nature split are from the stat weights run (±10).
//
// Next to a level 30 boss fight, Spirit is worth more than four times as much and MP5 three times, while
// Intellect is worth less. Over a session the rests matter as much as the fight, and Spirit
// and MP5 shorten them. Hit is for the first 2% (20.9 and then 22.1). The set has none, and 1% below
// none costs only a quarter of what the first 1% adds, so the stat weights run, which
// averages the step down and the step up, reads it low (13.3). Armor is worth little, since
// mobs at our level barely hurt us between rests.
export const EPLevel30Leveling = PresetUtils.makePresetEpWeights(
	'Level 30 Leveling',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.98,
			[Stat.StatAgility]: 1.81,
			[Stat.StatStamina]: 0.85,
			[Stat.StatIntellect]: 1.79,
			[Stat.StatSpirit]: 1.84,
			[Stat.StatAttackPower]: 1,
			[Stat.StatMeleeHit]: 21.5,
			[Stat.StatSpellHit]: 21.5,
			[Stat.StatMeleeCrit]: 14.1,
			[Stat.StatSpellCrit]: 14.1,
			[Stat.StatSpellPower]: 1.46,
			[Stat.StatSpellDamage]: 1.46,
			[Stat.StatFirePower]: 0.29,
			[Stat.StatNaturePower]: 1.12,
			[Stat.StatMP5]: 3.76,
			[Stat.StatArmor]: 0.06,
			[Stat.StatBonusArmor]: 0.06,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 13.62,
		},
	),
	{ tooltip: 'Stat weights for a solo level 30 grinding mobs of its level, by mobs killed in a session.', group: 'Leveling' },
);

// Stat weights for the leveling gear plan (2026-10-16), by mobs killed in a 20 min session, in
// attack power, at levels 10 to 30. The Leveling Gear tab judges the items it finds by them, and
// blends two of them for the levels in between.
//
// Each level is simmed on the plan's own gear at that level (seed 1, no enchants), with the
// leveling optimizer's best rotation for that level and gear. The talents follow the Level 30
// Leveling build in the order we'd take them: Thundering Strikes, Mental Dexterity, Improved
// Ghost Wolf, Shamanistic Focus, Ancestral Knowledge, then Flurry and Stormstrike at 25, where
// the 5 extra points start. Below 30 we use Rockbiter, which beats Flametongue there.
//
// Each stat is a direct probe at 20,000 sessions (+10 of it, +20 attack power or spell power,
// +5 MP5, +200 armor, +1% hit or crit, +2 weapon DPS). Strength is always 2, because it only
// gives us attack power. The weights move with the gear, because Stamina, Intellect and MP5
// decide how often we die (0.05 to 0.25 deaths a session). The gear the plan picks by one set
// of weights changes the next. So each level is the mean of three rounds: the plan's gear by the
// Level 30 Leveling weights, then by those weights, then by the mean of the first two. Stamina at
// level 20, for example, came out at 2.29, -0.04 and 2.33.
//
// [level, [Str, Agi, Sta, Int, Spi, hit, crit, SP, Fire, Nature, MP5, Armor, MH DPS]]
const levelingGearPlanWeights: Array<[number, number[]]> = [
	[10, [2.0, 0.74, 2.6, 0.94, 1.67, 7.07, 3.29, 2.75, 0.34, 2.33, 4.12, 0.09, 13.34]],
	[15, [2.0, 0.8, 1.42, 1.59, 1.01, 8.13, 5.11, 2.95, 0.6, 2.26, 2.69, 0.08, 14.05]],
	[20, [2.0, 1.24, 1.53, 2.83, 0.68, 13.28, 8.03, 3.48, 0.66, 2.54, 3.15, 0.13, 13.56]],
	[25, [2.0, 1.11, 1.77, 2.22, 0.51, 13.91, 10.0, 2.08, 0.37, 1.6, 3.36, 0.07, 13.14]],
	[30, [2.0, 1.29, 0.56, 2.0, 0.95, 14.19, 9.73, 1.63, 0.28, 1.27, 1.82, 0.08, 13.56]],
];
export const EPLevelingGearPlan = levelingGearPlanWeights.map(([level, [str, agi, sta, int, spi, hit, crit, sp, fire, nature, mp5, armor, mhDps]]) =>
	PresetUtils.makePresetEpWeights(
		`Level ${level} Leveling Gear`,
		Stats.fromMap(
			{
				[Stat.StatStrength]: str,
				[Stat.StatAgility]: agi,
				[Stat.StatStamina]: sta,
				[Stat.StatIntellect]: int,
				[Stat.StatSpirit]: spi,
				[Stat.StatAttackPower]: 1,
				[Stat.StatMeleeHit]: hit,
				[Stat.StatSpellHit]: hit,
				[Stat.StatMeleeCrit]: crit,
				[Stat.StatSpellCrit]: crit,
				[Stat.StatSpellPower]: sp,
				[Stat.StatSpellDamage]: sp,
				[Stat.StatFirePower]: fire,
				[Stat.StatNaturePower]: nature,
				[Stat.StatMP5]: mp5,
				[Stat.StatArmor]: armor,
				[Stat.StatBonusArmor]: armor,
			},
			{ [PseudoStat.PseudoStatMainHandDps]: mhDps },
		),
		{
			tooltip: `Stat weights for a solo level ${level} on the leveling gear plan's gear, by mobs killed in a session. The Leveling Gear tab judges items by them.`,
			group: 'Leveling Gear Plan',
		},
	),
);

export const EPPresets = [EPLevel30Leveling, ...EPLevelingGearPlan];

///////////////////////////////////////////////////////////////////////////
//                                 Leveling Plan
///////////////////////////////////////////////////////////////////////////

// The leveling plan of the Leveling mode (ui/core/leveling_plan.ts), from level 1 to 60. We
// grind up to 59, and 60 is where we stop.
//
// The talents and rotations come from the 2026-10-19 audit of the searched plan
// (docs/leveling_plan.md, The audit): agents read sessions at each level band, wrote new
// rules into the rotations and tried other talent routes, and we kept what held on two seeds.
// It takes 102h 8m, against 107h 5m for the search's plan. The route is the search's to 37
// (Thundering Strikes, Improved Ghost Wolf at 15 and 16, required, for walking, Mental
// Dexterity, Shamanistic Focus, Elemental Weapons, Stormstrike at 25, Flurry, Improved
// Stormstrike, Spirit Weapons and Tidal Focus), then Tidal Focus 4 and 5 and Mindfulness at
// 38 to 42, Maelstrom Weapon at 43 to 47, Rage of the Farseer at 48, and Convection and Call
// of Flame after. The rotations are "Audited 10" and so on, one range for each level that
// changed something, all on Rockbiter Weapon.
export const DefaultLevelingPlan = LevelingPlanSettings.create({
	// From level 1, where levels 1 to 9 play the first range's rotation with the spells we have
	// by then, and no talents.
	fromLevel: 1,
	toLevel: 60,
	bonusTalentPoints: 0,
	builds: [LevelingTalentBuild.create({ level: 10, picks: LevelingPlanAuditedJSON.route.map(pick => LevelingTalentPick.fromJson(pick as any)) })],
	rotationLibrary: LevelingPlanAuditedJSON.rotationLibrary.map(preset => LevelingRotationPreset.fromJson(preset as any)),
	rotations: LevelingPlanAuditedJSON.rotations.map(range => LevelingRotation.fromJson(range)),
	rotationsNameTopRanks: true,
	// Improved Ghost Wolf as early as its row opens (15 and 16), for walking to quests and grind
	// spots, which the grinding sessions don't sim.
	requiredTalents: [LevelingTalentRequirement.create({ tree: 1, talent: 5, points: 2, level: 16 })],
	// When we train each rank, from tools/levelingplan -training: 115g earned by 40 and riding
	// (90g after the Honored discount) at 43.
	training: LevelingTrainingSettings.fromJson(LevelingTrainingJSON as any),
});

// Our RestedXP route for an Orc Shaman (tools/leveling_route): the quest rewards we take on it
// and the dungeons we run, at the levels we're at then. The leveling gear plan takes them in
// place of the ones it makes up.
export const LevelingRouteOrcShaman = LevelingRoute.fromJson(LevelingRouteJSON);

// The Leveling mode with the plan above. The plan levels alone without the Settings tab's buffs,
// so we leave those as they are. The fight's duration and target are the Level 30 encounter's,
// and they only count when we go back to the DPS mode.
export const EncounterLevelingPlan = PresetUtils.makePresetEncounter(
	'Leveling',
	Encounter.create({
		...EncounterLevel30.encounter,
		leveling: LevelingOptions.create({ plan: true }),
	}),
	{ group: 'Leveling' },
);
export const PresetBuildLeveling1to60 = PresetUtils.makePresetBuild('Leveling 1-60', {
	group: 'Leveling',
	tooltip: 'The Leveling mode with our 1 to 60 plan: an Orc grinding alone, with its talent route and rotations.',
	encounter: EncounterLevelingPlan,
	race: Race.RaceOrc,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	levelingPlan: DefaultLevelingPlan,
});
