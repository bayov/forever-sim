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
	ManaRegenElixir,
	MobType,
	Potions,
	Profession,
	PseudoStat,
	Race,
	RaidBuffs,
	SapperExplosive,
	ShadowPowerBuff,
	SpellPowerBuff,
	Stat,
	StrengthBuff,
	TristateEffect,
	UnitReference,
	UnitReference_Type,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import {
	AirTotem,
	EarthTotem,
	EnhancementShaman_Options as EnhancementShamanOptions,
	FireTotem,
	ShamanSyncType,
	StartingTotems,
	WaterTotem,
} from '../core/proto/shaman.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import GraceOfAirAPLJSON from './apls/grace_of_air.apl.json';
import Level30p5APLJSON from './apls/level30p5.apl.json';
import Level30PvPAPLJSON from './apls/level30_pvp.apl.json';
import MagmaAPLJSON from './apls/magma.apl.json';
import ManaTideAPLJSON from './apls/mana_tide.apl.json';
import OptimizedAPLJSON from './apls/optimized.apl.json';
import WindfuryAPLJSON from './apls/windfury.apl.json';
import Level30p5GearJSON from './gear_sets/level30p5.gear.json';
import Level30PvPGearJSON from './gear_sets/level30_pvp.gear.json';
import Level30PvPHighHPGearJSON from './gear_sets/level30_pvp_high_hp.gear.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import Phase1GearJSON from './gear_sets/phase_1.gear.json';
import Phase2GearJSON from './gear_sets/phase_2.gear.json';
import Phase3GearJSON from './gear_sets/phase_3.gear.json';
import Phase5GearJSON from './gear_sets/phase_5.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearPhase1 = PresetUtils.makePresetGear('Phase 1', Phase1GearJSON, { tooltip: 'Level 60 Phase 1 gear made of real items.', group: 'Level 60' });
export const GearPhase2 = PresetUtils.makePresetGear('Phase 2', Phase2GearJSON, { tooltip: 'Level 60 Phase 2 gear.', group: 'Level 60' });
export const GearPhase3 = PresetUtils.makePresetGear('Phase 3', Phase3GearJSON, { tooltip: 'Level 60 Phase 3 gear.', group: 'Level 60' });
export const GearPhase5 = PresetUtils.makePresetGear('Phase 5', Phase5GearJSON, { tooltip: 'Level 60 Phase 5 gear.', group: 'Level 60' });
// The gear search at level 30 with 26 talent points, from the old level 20 + 5 set, with
// Engineering and Alchemy and the items no source is known for left out. The rest is
// Razorfen Kraul, Scarlet Monastery and Uldaman pieces plus "of the Tiger" and "of the
// Gorilla" greens.
//
// The two hander is Rage of the Storm (a Forever mace, Stormstrike does 10% more). wowhead
// has no source for it yet, and we assume it comes from a level 30 shaman quest chain. It
// is 6.4 DPS ahead of Corpsemaker (Razorfen Kraul) at 30, 60 and 300 sec, even with the
// Orc's Axe Specialization lost. With it the search swaps Pathfinder Hat for Enduring Cap,
// because the extra Stormstrike damage needs more mana (+1.5 at 60 sec).
//
// Manual Crowd Pummeler (Gnomeregan) is left out. The sim lets it use its 50% haste every
// 30 sec for the whole fight, but its 3 charges never come back in 1.12.
//
// On 2026-10-04 we searched again for the 120 sec Level 30 encounter, with its consumes
// and rotation. Three slots changed, and all three bring mana: Green Silk Armor
// (Tailoring, binds on equip) for Spirewind Fetter, Unearthed Bands of Intellect for
// Unearthed Bands of Power, and Defiler's Lizardhide Girdle (Friendly with The Defilers,
// from the Arathi Basin vendor in Hammerfall) for Archer's Belt of the Tiger. The set is
// 237.9 / 216.1 / 206.6 / 178.5 at 30 / 60 / 120 / 300 sec, from 236.5 / 215.4 / 201.5 /
// 176.0 on the old set. The girdle alone is worth 1.8 at 120 sec.
//
// The Darkspear shop does not work on the beta, so its items are left out. On 2026-10-06 we
// searched again with them allowed, and none of the eight a level 30 can wear wins a slot.
// Darkspear Raider's Cloak is the closest at 0.3 behind Knight's Cloak of the Gorilla.
// Darkspear Voodoo Seal needs level 33. Enduring Cap is 2.7
// ahead of Spellpower Goggles Xtreme at 120 sec. The goggles only win by 0.2 at 30 sec and
// 0.9 at 60, because the shaman does not run out of mana there.
//
// On 2026-10-05 the weapon went from Lesser Strength (+15 Strength) to Impact (+6 weapon
// damage), because the +15 Strength enchant is not in the beta. That is 206.5 against 207.7
// at 120 sec. Revelation (Enchanting 140) is the other choice. Its proc chance is unknown,
// so we tried three, and it lands at 204.9 / 204.3 / 204.1 for one proc every 1 / 2 / 3
// minutes. That is about 1 DPS for each proc a minute over no enchant (203.9), because a
// proc only turns one shock into a crit. It would need about three procs a minute to catch
// Impact.
//
// On 2026-10-06 we added Polished Driftwood Icon (Enchanting, 8% of mana regen goes on
// while casting) in the relic slot, and we took out a double count of healing gear. The sim
// used to add a third of every item's healing as spell damage, but Forever's items already
// carry that part ("healing by up to 21 and damage by up to 7" is 7 spell power and 14
// healing). That took the old set from 206.6 to 205.3 at 120 sec, and the icon puts back
// 0.4 (2.3 at 300 sec). A new search then swapped in Zealot's Robe (An Unholy Alliance)
// for Green Silk Armor and Traveled Gloves (No Honor Among Thieves) for the old gloves, +0.9
// each. The set is 237.8 / 215.7 / 207.5 / 181.7 at 30 / 60 / 120 / 300 sec.
//
// Later on 2026-10-06 the shaman took up Leatherworking in place of Alchemy, and we searched
// again with its gear. Skirmisher's Leather Belt (Leatherworking, binds on pickup) replaces
// Defiler's Lizardhide Girdle for +1.5 at 120 sec, and Prowler's Leather Belt is 0.4 behind
// it. Nothing else changed, enchants included. Defender's Leather Helm is 3.7 behind Enduring
// Cap, and Totemic Leather Leggings are 0.9 behind Ferine Leggings. The set is 239.6 / 217.7
// / 209.0 / 182.6 at 30 / 60 / 120 / 300 sec.
//
// On 2026-10-09 we searched again on the items of the 8 October beta build. Rage of the
// Storm got weaker (94-141 at 3.3 speed and 12 Intellect, was 102-154 at 3.6 and 16), and
// several quest rewards lost stats, so the old set fell to 195.6 at 120 sec. It is still
// our best weapon. Three slots changed: Kaleidoscope Chain (world drop) for Mark of the
// Pack Leader, Batwing Mantle (Blind Hunter, Razorfen Kraul, now with 6 spell power) for
// Bloodmage Mantle, and Garb of Florid Feathers (Excavation Site, now 9 Agility and 15
// Intellect) for Zealot's Robe, which went from 10 / 10 / 15 Intellect, Spirit and spell
// power to 7 / 7 / 11. The Gelkis and Magram rewards in Desolace are left out, because their quests
// need level 35 now. Enchants did not change. The set is 228.9 / 207.2 / 199.0 / 172.2 at
// 30 / 60 / 120 / 300 sec.
export const GearLevel30p5 = PresetUtils.makePresetGear('Level 30 + 5', Level30p5GearJSON, {
	tooltip: 'Level 30 gear with Rage of the Storm, searched for the 120 sec Level 30 encounter.',
	group: 'Level 30',
});
// Level 30 PvP is our gear for the Level 30 PvP encounter (70% of the fight out of melee
// range), with the PvP talents, Engineering and Leatherworking. It sims at 148.8 with 2434
// health and 790 armor without buffs. Three pieces are leatherworking crafts (Totemic Leather
// Helm, Skirmisher's Leather Belt and Totemic Leather Leggings).
//
// We chose it slot by slot (2026-10-07, rotopt -slot-tradeoffs). For each slot we swapped in
// every item and enchant against the rest of the set and kept the swaps no other swap beats
// on both damage and survival. Then we took the survival swap wherever it cost less than 3.5
// DPS for 100 health (an armor point counts a quarter of a health point). Each slot's swap
// and the price where it starts to pay:
//
//   head      Totemic Leather Helm, more damage for the same health
//   chest     Superior Stamina in place of Lesser Stats, 1.6
//   wrist     Unearthed Bands of Stamina for Technician's Bracers, 3.4, and Stamina +5
//             for Lesser Healing Power, 2.4
//   hands     Razzeric's Racing Grips (Safety First) with a Thick kit for Traveled Gloves, 3.0
//   rings     Nogg's Gold Ring and Determined Band, 3.1 to 3.2
//   feet      Gnomebot Operating Boots (Mekgineer Thermaplugg) with a Thick kit, 3.4
//   shoulder  Watchman Pauldrons for Berylline Pads, 3.8
//   back      Grimsteel Cape for Brilliant Cloak (Strahnbrad Mystery), 4.3
//   legs      Supple Bellyskin Leggings for Totemic Leather Leggings, 4.4
//   waist     Warden's Leather Belt for Skirmisher's Leather Belt, about 5, in neither set
//
// The last three are in Level 30 PvP (High HP). A weighted gear search (rotopt
// -health-weight) undershoots these. It only takes a swap that beats the noise at its
// iteration count by two standard errors, and most of these swaps gain less than that. It
// also leaves the enchants to a later pass.
//
// The 2026-10-06 version priced health at 3.5 too, but with 30% out of melee range, and
// did 160.5 with 1894 health here. 100 health costs far less damage at 70%, because the
// shaman is mostly shocking. In a one on one fight, a set wins when our damage times our
// health is higher, so 1% of damage is worth 1% of health, about 6 DPS for 100 health here.
// By that measure (health with buffs) this set is 17% ahead of the old one.
export const GearLevel30PvP = PresetUtils.makePresetGear('Level 30 PvP', Level30PvPGearJSON, {
	tooltip: 'Level 30 PvP gear that gives up some damage for more health, at up to 3.5 DPS per 100 health.',
	group: 'Level 30 PvP',
});
// Level 30 PvP (High HP) takes the same slot by slot swaps up to 5 DPS for 100 health. That
// adds Watchman Pauldrons (Scarlet Monastery trash), Grimsteel Cape (Vorrel's Revenge) and
// Supple Bellyskin Leggings to Level 30 PvP: 140.4 with 2617 health and 838 armor. Damage
// times health is about the same as Level 30 PvP, so pick by the fight. Warden's Leather
// Belt costs 2.6 DPS for 32 health and 50 armor here, just above the price.
//
// No two pieces in either set come from the same quest. Traveled Gloves and Determined Band
// both come from No Honor Among Thieves, and the gloves are out now.
export const GearLevel30PvPHighHP = PresetUtils.makePresetGear('Level 30 PvP (High HP)', Level30PvPHighHPGearJSON, {
	tooltip: 'Level 30 PvP gear with more health and armor than Level 30 PvP, for about 8 DPS less.',
	group: 'Level 30 PvP',
});

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
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [GearPhase1, GearLevel60Phase1Synthetic, GearLevel30p5, GearLevel30PvP, GearLevel30PvPHighHP],
	[Phase.Phase2]: [GearPhase2],
	[Phase.Phase3]: [GearPhase3],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [GearPhase5],
	[Phase.Phase6]: [],
};

export const DefaultGear = GearPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// Two ways to run the air totem, each holding one totem. Twisting Windfury Totem with
// Grace of Air is gone: since the 2026-09-24 beta build the air totem buffs no longer
// stack in a party, even from different shamans.
export const APLGraceOfAir = PresetUtils.makePresetAPLRotation('Grace of Air', GraceOfAirAPLJSON, {
	tooltip: 'Keeps Grace of Air Totem in the air slot.',
	group: 'Level 60',
});
export const APLWindfury = PresetUtils.makePresetAPLRotation('Windfury Totem', WindfuryAPLJSON, {
	tooltip: 'Keeps Windfury Totem in the air slot.',
	group: 'Level 60',
});
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
export const APLOptimized = PresetUtils.makePresetAPLRotation('Optimized', OptimizedAPLJSON, {
	tooltip: "The optimizer's best level 60 rotation, with Windfury Totem for the group, Flametongue Totem for the shaman and Fire Nova.",
	group: 'Level 60',
});
// The rotation for the Level 60 + Water Shield and Level 60 + Mana Tide talents. It is
// Optimized with Magma Totem in place of Searing Totem, and a new Magma Totem goes down
// with as little as 10 sec left in the fight. Those builds have no Elemental Focus, so
// there is no Clearcasting line, and Flame Shock goes after the bolt (within noise there).
//
// Those builds have no points for Improved Fire Nova, and Magma Totem makes up for some of
// it. On the Water Shield build it is 912.0 / 899.7 / 883.4 at 120 / 180 / 300 sec, where
// Searing Totem is 888.8 / 879.5 / 877.0. On the Level 60 build Searing Totem is still
// ahead (938.4 against 927.1 at 180 sec).
export const APLMagma = PresetUtils.makePresetAPLRotation('Magma Totem', MagmaAPLJSON, {
	tooltip: 'Optimized with Magma Totem in place of Searing Totem, for the Water Shield talents.',
	group: 'Level 60',
});
// Magma Totem with Mana Tide Totem once the shaman is down to 20% mana. Mana Spring Totem
// waits for the tide to end, because putting it down would knock the tide down.
//
// The shaman rarely runs that low, so it is worth little. On the Mana Tide build it is
// +2.5 at 300 sec and even at 120 and 180 sec.
export const APLManaTide = PresetUtils.makePresetAPLRotation('Mana Tide', ManaTideAPLJSON, {
	tooltip: 'The Magma Totem rotation that also drops Mana Tide Totem at 20% mana.',
	group: 'Level 60',
});
// Level 30 + 5 is Stormstrike on cooldown, Flame Shock when its DoT is down and Earth Shock
// otherwise (it spends the Stormstrike mark for +20%), with Mana Spring Totem up (level 26)
// and the shocks only above 10% mana. Flametongue Totem (level 28) takes the fire slot from
// Searing Totem (+6.4 at 60 sec and +8.6 at 300), and Fire Nova goes after the shocks.
//
// The rotation is tuned for the 120 sec Level 30 encounter. Fire Nova only goes out while
// the shaman is above 80% mana, so it spends the mana from the opening and leaves the rest
// for the shocks. On the 120 sec gear (see GearLevel30p5), Fire Nova at any mana was 9.6
// behind at 120 sec, because the shocks then run dry, and no Fire Nova at all was 0.5
// behind. At any mana Fire Nova would still be 2.4 / 6.6 ahead at 30 / 60 sec, where the
// pool lasts.
//
// The shocks stop below 10% mana. Down to 0% was even at 120 sec, but it ran dry and was
// 8.3 behind at 300 sec. A 20% floor was 1.0 behind at 120 (without Fire Nova).
export const APLLevel30p5 = PresetUtils.makePresetAPLRotation('Level 30 + 5', Level30p5APLJSON, {
	tooltip: 'Level 30 rotation with Flametongue Totem. Fire Nova only goes out above 80% mana.',
	group: 'Level 30',
});
// Level 30 PvP is the Level 30 + 5 rotation with Fire Nova at any mana, because the 60 sec
// fight is over before the pool runs out (+6.3 at 60 sec). Out of melee range the shaman
// still shocks, but has no Stormstrike, white hits or Fire Nova.
//
// A Lightning Bolt hard cast while out of range was 28 behind. The 3.5 sec cast runs on
// after the shaman is back in range and holds the swing.
export const APLLevel30PvP = PresetUtils.makePresetAPLRotation('Level 30 PvP', Level30PvPAPLJSON, {
	tooltip: 'The Level 30 + 5 rotation with Fire Nova at any mana. Out of melee range it only shocks.',
	group: 'Level 30 PvP',
});

export const APLPresets = {
	[Phase.Phase1]: [APLOptimized, APLMagma, APLManaTide, APLGraceOfAir, APLWindfury, APLLevel30p5, APLLevel30PvP],
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
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '500533102-053030031005112251' }), {
	tooltip: 'Level 60 talents from the talent search, with Spirit Weapons and Convection.',
	group: 'Level 60',
});
// The same search with Improved Ghost Wolf 2 pinned as well (-require
// spiritWeapons,improvedGhostWolf). The two points come out of Ancestral Knowledge, and
// the spare point goes back into it (4) for 1.3 DPS over Elemental Focus. The build is
// 1.2 DPS behind Level 60 at 300 sec.
//
// After the 2026-10-05 search it is the Level 60 build with the 2 points for Improved
// Ghost Wolf taken out of Ancestral Knowledge (1). The search with both talents pinned
// found nothing better. It is 894.0 at 300 sec, 3.4 behind Level 60.
export const TalentsLevel60ImpGW = PresetUtils.makePresetTalents('Level 60 + Imp GW', SavedTalents.create({ talentsString: '500533102-051032031005112251' }), {
	tooltip: 'The Level 60 talents with Improved Ghost Wolf 2, taken out of Ancestral Knowledge.',
	group: 'Level 60',
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
export const TalentsLevel60WaterShield = PresetUtils.makePresetTalents(
	'Level 60 + Water Shield',
	SavedTalents.create({ talentsString: '500013-053030031005112151-050050001' }),
	{ tooltip: 'Level 60 talents with 10 Restoration points for Water Shield.', group: 'Level 60' },
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
export const TalentsLevel60ManaTide = PresetUtils.makePresetTalents(
	'Level 60 + Mana Tide',
	SavedTalents.create({ talentsString: '02-055030031005112151-053051001001' }),
	{ tooltip: 'Level 60 talents with Water Shield and Mana Tide Totem.', group: 'Level 60' },
);
// Level 30 with 5 extra talent points (26 in all), with Stormstrike, Spirit Weapons and
// Improved Ghost Wolf 2 pinned. We scored every build of the other Enhancement talents,
// with up to 5 points in Concussion, Convection or Totemic Focus. Flurry 5, Mental
// Quickness 2 and Improved Stormstrike 2 are in every top build. On Corpsemaker the top
// dozen were within 0.1 DPS at 60 sec. They only move points between Ancestral
// Knowledge, Improved Lightning Shield, Elemental Weapons and Shamanistic Focus.
//
// With Rage of the Storm we scored them all again, and Ancestral Knowledge 4 in place of
// Improved Lightning Shield 3 is now on top (+0.5 / +0.9 / +0.7 at 30 / 60 / 300 sec),
// because the mana goes to more Stormstrikes. This one is Thundering Strikes 5, Ancestral
// Knowledge 4, Mental Dexterity 3 and Shamanistic Focus 1.
//
// The sim used to give Improved Stormstrike 2 double the mana regen (100% instead of 50%).
// After the fix we scored every build again and this one is still on top at 30, 60 and
// 300 sec. Improved Stormstrike 2 is still worth 11 DPS at 60 sec.
export const TalentsLevel30p5 = PresetUtils.makePresetTalents('Level 30 + 5', SavedTalents.create({ talentsString: '-0540320010051122' }), {
	tooltip: 'Level 30 with 5 extra talent points, with Stormstrike, Spirit Weapons and Improved Ghost Wolf 2.',
	group: 'Level 30',
});
// The talents we play in PvP. They are picked by hand for survival (Toughness gives 8%
// more Stamina, about 90 health), not by a search.
export const TalentsLevel30PvP = PresetUtils.makePresetTalents('Level 30 PvP', SavedTalents.create({ talentsString: '-0502320013401122' }), {
	tooltip: 'Level 30 PvP talents picked by hand for survival, with Toughness for more Stamina.',
	group: 'Level 30 PvP',
});

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60ImpGW, TalentsLevel60WaterShield, TalentsLevel60ManaTide, TalentsLevel30p5, TalentsLevel30PvP],
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

export const DefaultOptions = EnhancementShamanOptions.create({
	shamanImbue: WeaponImbue.WindfuryWeapon,
	syncType: ShamanSyncType.Auto,
});

// The class options every preset build sets. They always set the starting totems, so loading a
// build puts its own in place and changing them marks the build as modified.
//
// The Level 30 PvP builds start the fight with no totems and put them down in the rotation's
// prepull actions.
const PresetBuildOptions = { shamanImbue: WeaponImbue.WindfuryWeapon, startingTotems: StartingTotems.create() };

// The raid and dungeon builds start the fight with their totems already down, each at the
// default seconds before the pull, one totem GCD apart (see starting_totems.tsx). They cost no
// mana, so the rotations leave them out of the prepull actions. Totems last 5 min under
// Forever, so a 3 min fight never puts them down again.
//
// On 2026-10-08 that was +0.3 at 180 sec on Level 60 over the old prepull totems, and +0.7
// at 120 sec on Level 30 + 5. At 300 sec it can be 1 to 2 behind, because the totems run out
// 30 sec sooner and the rotation puts them down again.
const startingTotemsLevel60 = StartingTotems.create({
	earth: EarthTotem.StrengthOfEarthTotem,
	earthSecondsBeforePull: 28,
	air: AirTotem.WindfuryTotem,
	airSecondsBeforePull: 27,
	fire: FireTotem.FlametongueTotem,
	fireSecondsBeforePull: 29,
	water: WaterTotem.ManaSpringTotem,
	waterSecondsBeforePull: 30,
});
const PresetBuildOptionsLevel60 = { ...PresetBuildOptions, startingTotems: startingTotemsLevel60 };
// Magma Totem lasts only 20 sec, so it goes down 1 sec before the pull.
const PresetBuildOptionsLevel60Magma = {
	...PresetBuildOptions,
	startingTotems: StartingTotems.create({ ...startingTotemsLevel60, fire: FireTotem.MagmaTotem, fireSecondsBeforePull: 1 }),
};
// Windfury Totem comes at level 32.
const PresetBuildOptionsLevel30 = {
	...PresetBuildOptions,
	startingTotems: StartingTotems.create({ ...startingTotemsLevel60, air: AirTotem.NoAirTotem, airSecondsBeforePull: 0 }),
};

// The level 60 consumables. On 2026-10-08 we tried every consumable the sim applies in each
// slot on the Level 60 build at 180 sec. Under Forever the shaman imbue stacks with an oil or a
// stone, and Shadow Oil was best there (+19.6, Brilliant Wizard Oil and Elemental Sharpening
// Stone were +10). Goblin Land Mine (+4.1), Elixir of Wisdom (+2.1) and Elixir of Shadow Power
// (+1.2, for Shadow Oil) were also worth it. Grilled Squid (1% crit under Forever) is +3.4 over
// Blessed Sunfruit, and even with Smoked Desert Dumpling, which is not out yet.
//
// All together with the starting totems the Level 60 build went from 788.4 to 819.6 at 180 sec.
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MajorManaPotion,
	defaultConjured: Conjured.ConjuredDemonicRune,
	dragonBreathChili: true,
	fillerExplosive: Explosive.ExplosiveGoblinLandMine,
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodGrilledSquid,
	intellectElixir: IntellectElixir.ElixirOfWisdom,
	mainHandImbue: WeaponImbue.ShadowOil,
	manaRegenElixir: ManaRegenElixir.MagebloodPotion,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	shadowPowerBuff: ShadowPowerBuff.ElixirOfShadowPower,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.ROIDS,
});

// The raid the level 60 numbers were measured under (the rogue search's export minus
// the totems): both paladin blessings, Judgement of Wisdom and every debuff, but no
// second shaman in the group, so the player's own Strength of Earth and Grace of Air
// are what the rotation casts and there is no Mana Spring. Judgement of Wisdom alone is
// 60 DPS here because without it the shaman is out of mana two minutes in.
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
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
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

// The shaman is an engineer and a leatherworker (an alchemist until 2026-10-06). Engineering
// is for the Goblin Sapper in the consumes (worth about 1 DPS) and the goggles. Gear another
// profession makes is only in the presets when it binds on equip, so it can be bought. The
// gear searches before that date ran without the leatherworking gear that binds on pickup.
export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
	race: Race.RaceOrc,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor), a 3 minute
// fight, and the default buffs, debuffs and consumes. It is here so switching back from
// the Level 20 build restores everything that build changed. The synthetic Phase 1 set and
// its talents were searched at 180 sec.
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
// On 2026-10-08, with the starting totems and the new consumables, Level 60 is 826.5 / 819.6 /
// 807.7 at 120 / 180 / 300 sec, Level 60 + Water Shield 776.6 / 767.0 / 751.2 and Level 60 +
// Mana Tide 756.6 / 750.3 / 745.3. A knob search kept every rotation knob but the Mana Tide
// build's Fire Nova order (see shaman_presets.sh). It wanted Grace of Air in place of Windfury
// Totem (+23 at 180 sec), but the builds keep Windfury Totem for the group. The talents were
// not searched again.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Level 60', {
	group: 'Level 60',
	tooltip: 'Level 60 Orc on the synthetic Phase 1 gear with the Optimized rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLOptimized,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsLevel60,
});
export const PresetBuildLevel60WaterShield = PresetUtils.makePresetBuild('Level 60 + Water Shield', {
	group: 'Level 60',
	tooltip: 'Level 60 with the Water Shield talents and the Magma Totem rotation.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60WaterShield,
	rotation: APLMagma,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsLevel60Magma,
});
export const PresetBuildLevel60ManaTide = PresetUtils.makePresetBuild('Level 60 + Mana Tide', {
	group: 'Level 60',
	tooltip: 'Level 60 with the Water Shield and Mana Tide talents and the Mana Tide rotation.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60ManaTide,
	rotation: APLManaTide,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsLevel60Magma,
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
// Stat weights for the level 30 sets, in attack power (2026-10-07). Each one is measured
// on its own set and encounter: rotopt -weights at 100k iterations, then direct probes
// for the rest.
//
// Hit and crit come from probes because under Forever an item's hit or crit counts for
// melee and spells both. The weights run moves the melee and the spell stat one at a time,
// so it reads about half. The sets also have no hit, and the weights run can't probe below
// zero. The hit value is for the first 2%. Each point above that is worth less, because
// the spell hit cap comes first.
export const EPLevel30 = PresetUtils.makePresetEpWeights(
	'Level 30',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.84,
			[Stat.StatIntellect]: 2.93,
			[Stat.StatSpirit]: 0.4,
			[Stat.StatAttackPower]: 1,
			[Stat.StatMeleeHit]: 48.3,
			[Stat.StatSpellHit]: 48.3,
			[Stat.StatMeleeCrit]: 37,
			[Stat.StatSpellCrit]: 37,
			[Stat.StatSpellPower]: 2.09,
			[Stat.StatSpellDamage]: 2.09,
			[Stat.StatFirePower]: 1.07,
			[Stat.StatNaturePower]: 0.78,
			[Stat.StatMP5]: 1.28,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 13.83,
		},
	),
	{ tooltip: 'Stat weights for the Level 30 + 5 gear on the 120 sec Level 30 encounter.', group: 'Level 30' },
);
// Level 30 PvP (High HP) weights against each kind of enemy player (2026-10-07). Each one
// sims the High HP set in a 60 sec duel from in front of the enemy, so a parry or a block
// can stop our hits. A melee enemy keeps us in range, so we are out of melee range 30% of
// the fight. A caster or a hunter kites us, 70%. The enemy at level 30:
//
//   cloth      mage, priest, warlock: 320 armor, 4.5% dodge, never hits us in melee
//              (so Lightning Shield never fires)
//   hunter     leather, 770 armor, 7% dodge, 5% parry, the pet hits us, kites
//   rogue      leather, 790 armor, 14% dodge, 5% parry, fast hits
//   shaman     leather until 40, 700 armor, 6.5% dodge, no parry, two hander
//   warrior    and paladin with a two hander: mail, 1320 armor, 4.5% dodge, 5% parry
//   prot       warrior or paladin with a shield: 2000 armor, 5.5% dodge, 6% parry, 12%
//              block for 18
//
// The armor is the median level 25 to 30 green or better per slot, plus 2 per Agility. The
// dodge is the class base plus its Agility at level 30.
//
// Health is priced so that 1% of our damage is worth 1% of our health (about 2790 with
// buffs). That is what decides a duel, and it is where the High HP set's 5 DPS for 100
// health comes from. So the price follows our damage in each fight, from 4.5 against
// cloth to 6.0 against a shaman. Armor is priced by how much health it saves us against
// the enemy's physical damage: one armor is 0.74 health against a fully physical enemy, at
// our 838 armor. A caster's damage is all spells, so there armor is worth nothing.
//
// In the melee fights we run out of mana before the minute is up, so Intellect and MP5 are
// worth a lot more there, most of all against a rogue.
const makeLevel30PvPEP = (
	name: string,
	tooltip: string,
	ep: {
		agi: number;
		sta: number;
		int: number;
		spi: number;
		mp5: number;
		hit: number;
		crit: number;
		sp: number;
		fire: number;
		nature: number;
		armor: number;
	},
) =>
	PresetUtils.makePresetEpWeights(
		name,
		Stats.fromMap(
			{
				[Stat.StatStrength]: 2.2,
				[Stat.StatAgility]: ep.agi,
				[Stat.StatStamina]: ep.sta,
				[Stat.StatIntellect]: ep.int,
				[Stat.StatSpirit]: ep.spi,
				[Stat.StatAttackPower]: 1,
				[Stat.StatMeleeHit]: ep.hit,
				[Stat.StatSpellHit]: ep.hit,
				[Stat.StatMeleeCrit]: ep.crit,
				[Stat.StatSpellCrit]: ep.crit,
				[Stat.StatSpellPower]: ep.sp,
				[Stat.StatSpellDamage]: ep.sp,
				[Stat.StatFirePower]: ep.fire,
				[Stat.StatNaturePower]: ep.nature,
				[Stat.StatMP5]: ep.mp5,
				[Stat.StatArmor]: ep.armor,
				[Stat.StatBonusArmor]: ep.armor,
			},
			{
				[PseudoStat.PseudoStatMainHandDps]: 13.66,
			},
		),
		{ tooltip, group: 'Level 30 PvP' },
	);
export const EPLevel30PvPVsCloth = makeLevel30PvPEP('PvP vs Cloth', 'Level 30 PvP (High HP) against a mage, priest or warlock, 70% out of melee range.', {
	agi: 0.8,
	sta: 6.39,
	int: 1.96,
	spi: 0,
	mp5: 0,
	hit: 38.5,
	crit: 23.5,
	sp: 2.09,
	fire: 1.3,
	nature: 0.48,
	armor: 0,
});
export const EPLevel30PvPVsHunter = makeLevel30PvPEP('PvP vs Hunter', 'Level 30 PvP (High HP) against a hunter, 70% out of melee range.', {
	agi: 1.79,
	sta: 7.35,
	int: 2.49,
	spi: 0.2,
	mp5: 0.44,
	hit: 41.7,
	crit: 24.9,
	sp: 3.12,
	fire: 1.46,
	nature: 1.32,
	armor: 0.4,
});
export const EPLevel30PvPVsRogue = makeLevel30PvPEP('PvP vs Rogue', 'Level 30 PvP (High HP) against a rogue, 30% out of melee range.', {
	agi: 1.85,
	sta: 6.4,
	int: 4.54,
	spi: 0.82,
	mp5: 2.68,
	hit: 40.7,
	crit: 23.6,
	sp: 2.36,
	fire: 1.18,
	nature: 0.93,
	armor: 0.37,
});
export const EPLevel30PvPVsShaman = makeLevel30PvPEP('PvP vs Shaman', 'Level 30 PvP (High HP) against an enhancement shaman, 30% out of melee range.', {
	agi: 1.29,
	sta: 6.19,
	int: 2.92,
	spi: 0.62,
	mp5: 1.62,
	hit: 36.7,
	crit: 22.4,
	sp: 2.25,
	fire: 1.14,
	nature: 0.83,
	armor: 0.21,
});
export const EPLevel30PvPVsWarrior = makeLevel30PvPEP(
	'PvP vs Warrior/Paladin',
	'Level 30 PvP (High HP) against a warrior or paladin with a two hander, 30% out of melee range.',
	{
		agi: 1.84,
		sta: 6.58,
		int: 2.69,
		spi: 0.47,
		mp5: 1.15,
		hit: 41,
		crit: 23.7,
		sp: 2.46,
		fire: 1.39,
		nature: 0.77,
		armor: 0.4,
	},
);
export const EPLevel30PvPVsProt = makeLevel30PvPEP(
	'PvP vs Prot',
	'Level 30 PvP (High HP) against a warrior or paladin with a shield, 30% out of melee range.',
	{
		agi: 1.84,
		sta: 7.19,
		int: 3.59,
		spi: 0.77,
		mp5: 1.91,
		hit: 42.3,
		crit: 23.9,
		sp: 3,
		fire: 1.59,
		nature: 1,
		armor: 0.44,
	},
);

// The two Level 30 PvP weights are the mean of the six enemies above (2026-10-07), each
// enemy counted the same. They are for gearing when we don't know who we will fight.
//
// Level 30 PvP (High HP) is the plain mean of the six presets above.
//
// For Level 30 PvP we ran the same six fights on the Level 30 PvP set (153.7 on average,
// where the High HP set does 143.7). That set was picked with health at 3.5 DPS for 100,
// not 5, so we scale the health price of every fight down to average 3.5. Each fight
// keeps its share, so it goes from 3.0 against cloth to 4.1 against a shaman. The other
// stats are close to High HP, but Intellect and MP5 are worth less. The set has more
// mana, so we run out later in the melee fights.
//
// Before 2026-10-07 both came from the single Level 30 PvP encounter, with health at a
// flat 3.5 and 5 and an armor point at a quarter of a health point.
export const EPLevel30PvP = makeLevel30PvPEP('Level 30 PvP', 'The mean of the six enemy weights on Level 30 PvP gear, with 100 health worth 3.5 DPS.', {
	agi: 1.42,
	sta: 4.51,
	int: 2.3,
	spi: 0.14,
	mp5: 0.37,
	hit: 42.6,
	crit: 24.9,
	sp: 2.59,
	fire: 1.37,
	nature: 0.91,
	armor: 0.19,
});
export const EPLevel30PvPHighHP = makeLevel30PvPEP(
	'Level 30 PvP (High HP)',
	'The mean of the six enemy weights on Level 30 PvP (High HP) gear, with 100 health worth 5.1 DPS.',
	{
		agi: 1.57,
		sta: 6.68,
		int: 3.03,
		spi: 0.48,
		mp5: 1.3,
		hit: 40.1,
		crit: 23.6,
		sp: 2.55,
		fire: 1.34,
		nature: 0.89,
		armor: 0.31,
	},
);
export const EPPresets = [
	EPLevel30,
	EPLevel30PvP,
	EPLevel30PvPHighHP,
	EPLevel30PvPVsCloth,
	EPLevel30PvPVsHunter,
	EPLevel30PvPVsRogue,
	EPLevel30PvPVsShaman,
	EPLevel30PvPVsWarrior,
	EPLevel30PvPVsProt,
];

// On 2026-10-08, with the starting totems, Level 30 + 5 was 209.1 / 201.0 / 174.7 at 60 / 120 /
// 300 sec. After the 2026-10-09 regear on the 8 October build's items it is 207.2 / 199.0 /
// 172.2. Every consumable the sim applies that a level 30 can use was tried, and the ones
// here stayed on top. A knob search kept every rotation knob.
export const PresetBuildLevel30p5 = PresetUtils.makePresetBuild('Level 30 + 5', {
	group: 'Level 30',
	tooltip: 'Level 30 Orc with 5 extra talent points, soloing Interrogator Vishas for 120 sec with level 30 consumes.',
	gear: GearLevel30p5,
	talents: TalentsLevel30p5,
	rotation: APLLevel30p5,
	epWeights: EPLevel30,
	encounter: EncounterLevel30,
	race: Race.RaceOrc,
	level: 30,
	bonusTalentPoints: 5,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsLevel30,
});

// Level 30 PvP is a 60 sec fight against a level 30 player in PvP mode. The enemy keeps us
// out of melee range for 70% of the fight, in random stretches of 1 to 10 sec, and our
// white hits on a player never glance. The enemy has 600 armor (a level 30 in mail or
// leather) and hits us, so Lightning Shield fires. Buffs and consumes are the same as for
// the Level 30 encounter.
//
// We attack from in front, as in any duel, so the enemy can parry. It dodges and parries
// 5% each and has no shield. Since 2026-10-07, before that we attacked from behind. From
// the front, Level 30 PvP does 143.1 (148.8 from behind) and Level 30 PvP (High HP) 135.0
// (140.4). The gear numbers above are from behind.
export const EncounterLevel30PvP = PresetUtils.makePresetEncounter(
	'Level 30 PvP',
	Encounter.create({
		duration: 60,
		durationVariation: 5,
		executeProportion20: 0.2,
		executeProportion25: 0.25,
		executeProportion35: 0.35,
		pvp: true,
		pvpMeleeDowntime: 0.7,
		targets: [
			{
				name: 'Level 30 Player',
				level: 30,
				mobType: MobType.MobTypeHumanoid,
				stats: new Stats()
					.withStat(Stat.StatArmor, 600)
					.withStat(Stat.StatHealth, 30000)
					.withStat(Stat.StatDodge, 5)
					.withStat(Stat.StatParry, 5)
					.asArray(),
				minBaseDamage: 52,
				damageSpread: 0.3333,
				swingSpeed: 2,
				tankIndex: 0,
			},
		],
	}),
	{
		tanks: EncounterLevel30.tanks,
		raidBuffs: EncounterLevel30.raidBuffs,
		debuffs: EncounterLevel30.debuffs,
		buffs: EncounterLevel30.buffs,
		consumes: EncounterLevel30.consumes,
		inFrontOfTarget: true,
	},
);
export const PresetBuildLevel30PvP = PresetUtils.makePresetBuild('Level 30 PvP', {
	group: 'Level 30 PvP',
	tooltip: 'A 60 sec duel against a level 30 player who keeps us out of melee range 70% of the time.',
	// We start from the High HP set and its stat weights. It does about 8 DPS less than
	// Level 30 PvP for 180 more health, and in a duel we'd rather stay alive longer.
	gear: GearLevel30PvPHighHP,
	talents: TalentsLevel30PvP,
	rotation: APLLevel30PvP,
	epWeights: EPLevel30PvPHighHP,
	encounter: EncounterLevel30PvP,
	race: Race.RaceOrc,
	level: 30,
	bonusTalentPoints: 5,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptions,
});

// Level 30 PvP against each kind of enemy player, with the High HP set and its stat
// weights for that enemy (see EPLevel30PvPVsCloth for how we built each enemy). A melee
// enemy keeps us in range, so we are out of melee range 30% of the fight. A caster or a
// hunter kites us, 70%. We attack from in front.
const makeLevel30PvPEncounter = (
	name: string,
	meleeDowntime: number,
	enemy: { name: string; armor: number; dodge: number; parry: number; block?: number; blockValue?: number; swingSpeed: number },
) =>
	PresetUtils.makePresetEncounter(
		name,
		Encounter.create({
			duration: 60,
			durationVariation: 5,
			executeProportion20: 0.2,
			executeProportion25: 0.25,
			executeProportion35: 0.35,
			pvp: true,
			pvpMeleeDowntime: meleeDowntime,
			targets: [
				{
					name: enemy.name,
					level: 30,
					mobType: MobType.MobTypeHumanoid,
					stats: new Stats()
						.withStat(Stat.StatArmor, enemy.armor)
						.withStat(Stat.StatHealth, 30000)
						.withStat(Stat.StatDodge, enemy.dodge)
						.withStat(Stat.StatParry, enemy.parry)
						.withStat(Stat.StatBlock, enemy.block ?? 0)
						.withStat(Stat.StatBlockValue, enemy.blockValue ?? 0)
						.asArray(),
					// A caster never hits us in melee, so it has no swing at all.
					minBaseDamage: enemy.swingSpeed ? 52 : 0,
					damageSpread: 0.3333,
					swingSpeed: enemy.swingSpeed,
					tankIndex: 0,
				},
			],
		}),
		{
			tanks: EncounterLevel30.tanks,
			raidBuffs: EncounterLevel30.raidBuffs,
			debuffs: EncounterLevel30.debuffs,
			buffs: EncounterLevel30.buffs,
			consumes: EncounterLevel30.consumes,
			inFrontOfTarget: true,
		},
	);
const makeLevel30PvPBuild = (name: string, tooltip: string, encounter: PresetUtils.PresetEncounter, epWeights: PresetUtils.PresetEpWeights) =>
	PresetUtils.makePresetBuild(name, {
		group: 'Level 30 PvP',
		tooltip,
		gear: GearLevel30PvPHighHP,
		talents: TalentsLevel30PvP,
		rotation: APLLevel30PvP,
		epWeights,
		encounter,
		race: Race.RaceOrc,
		level: 30,
		bonusTalentPoints: 5,
		professions: [OtherDefaults.profession1, OtherDefaults.profession2],
		options: PresetBuildOptions,
	});

// 126.5 DPS. Lightning Shield never fires, because a caster doesn't hit us in melee.
export const EncounterLevel30PvPVsCloth = makeLevel30PvPEncounter('Level 30 PvP vs Cloth', 0.7, {
	name: 'Level 30 Mage, Priest or Warlock',
	armor: 320,
	dodge: 4.5,
	parry: 0,
	swingSpeed: 0,
});
// 129.8 DPS. The pet's hits fire Lightning Shield.
export const EncounterLevel30PvPVsHunter = makeLevel30PvPEncounter('Level 30 PvP vs Hunter', 0.7, {
	name: 'Level 30 Hunter',
	armor: 770,
	dodge: 7,
	parry: 5,
	swingSpeed: 2,
});
// 154.3 DPS.
export const EncounterLevel30PvPVsRogue = makeLevel30PvPEncounter('Level 30 PvP vs Rogue', 0.3, {
	name: 'Level 30 Rogue',
	armor: 790,
	dodge: 14,
	parry: 5,
	swingSpeed: 1.3,
});
// 167.9 DPS. Shamans wear leather until level 40 and can't parry.
export const EncounterLevel30PvPVsShaman = makeLevel30PvPEncounter('Level 30 PvP vs Shaman', 0.3, {
	name: 'Level 30 Enhancement Shaman',
	armor: 700,
	dodge: 6.5,
	parry: 0,
	swingSpeed: 3.5,
});
// 147.3 DPS. Warriors and paladins wear mail until level 40.
export const EncounterLevel30PvPVsWarrior = makeLevel30PvPEncounter('Level 30 PvP vs Warrior/Paladin', 0.3, {
	name: 'Level 30 Warrior or Paladin',
	armor: 1320,
	dodge: 4.5,
	parry: 5,
	swingSpeed: 3.4,
});
// 136.3 DPS. A level 30 shield has about 650 armor and blocks 14, plus Strength / 20. The
// block chance is the base 5% with Shield Specialization and some defense.
export const EncounterLevel30PvPVsProt = makeLevel30PvPEncounter('Level 30 PvP vs Prot', 0.3, {
	name: 'Level 30 Prot Warrior or Paladin',
	armor: 2000,
	dodge: 5.5,
	parry: 6,
	block: 12,
	blockValue: 18,
	swingSpeed: 2.5,
});

export const PresetBuildLevel30PvPVsCloth = makeLevel30PvPBuild(
	'Level 30 PvP vs Cloth',
	'The High HP set against a mage, priest or warlock who kites us and never hits us in melee.',
	EncounterLevel30PvPVsCloth,
	EPLevel30PvPVsCloth,
);
export const PresetBuildLevel30PvPVsHunter = makeLevel30PvPBuild(
	'Level 30 PvP vs Hunter',
	'The High HP set against a hunter who kites us while the pet hits us.',
	EncounterLevel30PvPVsHunter,
	EPLevel30PvPVsHunter,
);
export const PresetBuildLevel30PvPVsRogue = makeLevel30PvPBuild(
	'Level 30 PvP vs Rogue',
	'The High HP set against a rogue who keeps us in melee range most of the fight.',
	EncounterLevel30PvPVsRogue,
	EPLevel30PvPVsRogue,
);
export const PresetBuildLevel30PvPVsShaman = makeLevel30PvPBuild(
	'Level 30 PvP vs Shaman',
	'The High HP set against an enhancement shaman who keeps us in melee range most of the fight.',
	EncounterLevel30PvPVsShaman,
	EPLevel30PvPVsShaman,
);
export const PresetBuildLevel30PvPVsWarrior = makeLevel30PvPBuild(
	'Level 30 PvP vs Warrior/Paladin',
	'The High HP set against a warrior or paladin with a two hander.',
	EncounterLevel30PvPVsWarrior,
	EPLevel30PvPVsWarrior,
);
export const PresetBuildLevel30PvPVsProt = makeLevel30PvPBuild(
	'Level 30 PvP vs Prot',
	'The High HP set against a warrior or paladin with a shield, who can block our hits.',
	EncounterLevel30PvPVsProt,
	EPLevel30PvPVsProt,
);
