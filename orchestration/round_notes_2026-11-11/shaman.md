# Notes round 2026-11-11: shaman (Blizzard's Nov 10 class tuning)

Landed on master at b44dfdbab (report data e23157a04), deployed to forever-dps.pages.dev (deploy f48be298).

## Changes and how each was confirmed

No combat logs dated after 2026-11-10 16:51 UTC are in this environment (none of any date: `~/forever-logs` is absent),
so nothing could be checked against a log. Every value below is **notes only** (standards 37), cites Blizzard's
hotfix article and Kaivax's "Class Tuning Incoming - November 10", and is listed under "Notes-only values" in
`orchestration/round_70753/shaman.md` with the row or log that would confirm it.

| # | Change | Sim | Confirmed by | Test |
|---|---|---|---|---|
| S1 | Stormstrike 100% weapon damage (was 150%) | `sim/shaman/stormstrike.go` | notes only (SpellEffect 1368308 still 150 at push 113380) | TestOrcShamanWeaponDamage |
| S2 | Stormstrike 8 s cooldown (was 6) | same | notes only (SpellCooldowns 55241 still 6000 ms) | TestOrcShamanStormstrikeOutcomes, TestForeverWardenAbilities |
| S3 | Mental Dexterity 33/66/100% Int to AP | `sim/shaman/talents.go` | notes only (no 415140 row) | TestOrcShamanMentalDexterity, TestForeverWardenTalents |
| S4 | Maelstrom Weapon 20% a stack at every rank | `talents.go` | notes only | TestOrcShamanMaelstromWeaponRanks |
| S5 | Maelstrom Weapon (5 + rank) PPM on main hand speed, replacing every-hit (bcc5d6cd4) | `talents.go` | notes only; the tooltip's example (60% at 3.6, 33% at 2.0, 5/5) checked | TestOrcShamanMaelstromWeaponProcRate; Totem of the Storm now 0.5 x the swing chance (TestOrcShamanRelics) |
| S6 | Weapon imbue and weapon totem work together (WF+WF, FT+FT) | blocks removed in `sim/core/buffs.go`, `sim/core/totem_weapon_buffs.go`; APL notes removed | "live" in the post, no row; no log | TestOrcShamanImbueWithWeaponTotem, TestOrcShamanImbueProcs, TestOrcShamanWeaponTotemSlot |
| S7 | Rage of the Storm: 16 Str, 15 Sta, 102-154 at 3.6 | hand entry in `assets/db_inputs/forever_seen_in_game.json` | hotfix 113886 | Flametongue/Frostbrand weapon tests |

S6 cooldowns: the post does not say whether the two Windfury procs share a cooldown. Each keeps its own (weapon 20% /
1.5 s; totem 20% / 100 ms, white main hand swings only): report assumption `wf_weapon_totem_icd`.

## 180 s DPS, published, old -> new

| Build | Old | New |
|---|---|---|
| Enhancement Level 60 | 1159.4 | 1030.6 |
| Enhancement Imp GW | 1149.8 | 1016.2 |
| Enhancement Mana Tide | 1055.1 | 930.0 |
| Enhancement Water Shield | 1111.3 | 976.3 |
| Elemental Level 60 | 970.1 | 976.9 |
| Elemental Mana Tide | 952.7 | 959.6 |

Elemental's move comes from hotfix 113380, which was already on master but not yet re-simmed. S1-S6 leave Elemental
identical, and S6 doesn't reach it, since no group member in its builds carries an imbue. No raid preset or bench
build pairs an imbue with the same weapon totem, so no raid build was re-simmed for S6. Enhancement's overall ranks fell
8->37, 10->43, 18->51, 31->60. The other classes are still the Nov 4 sims, so these ranks compare against stale numbers.

## New talents and rotation (rotopt, 20k iterations, seeds 1 and 1000003, kept only where it won at 120/180/300 s)

- Imbue and totems, all four builds: Windfury Weapon + own Windfury Totem + Magma Totem still wins (Flametongue Weapon
  93-100 behind, Flametongue Totem 110-120 behind, Searing 25 behind).
- Level 60 and Imp GW: Convection 3 -> Reverberation 3 (`053303102-...`), +15 to +21 DPS over the old talents on the
  new sim. Rotation: the bolt waits for Earth Shock again, Earth Shock no longer holds for the Stormstrike mark, and
  Fire Nova's mana budget goes from 12 to 16 per second.
- Water Shield: Elemental Devastation 2 -> 1, Call of Flame 2 -> 3 (+1 to +3). Rotation knobs unchanged.
- Mana Tide: talents unchanged. Flame Shock goes ahead of Stormstrike, the bolt comes after Earth Shock, and the Earth
  Shock hold for Flame Shock's end is dropped (+2 to +3).
- Gear: no change. Hit steps stay at hit-3 / hit-1 / hit-4 / hit-4. Intellect falls to last in the Level 60 weights,
  but moving 10 Int to Str gains only 0.3-0.4 DPS, under the 1 DPS bar. Dwarf race gear: Imp GW re-spent (+1.6). The
  other three Dwarf overrides are dropped.
- compsearch: Enhancement no longer takes Flametongue Totem; the Windfury gift is Windfury Totem.

## Left alone (user ruling), and what it would need

- Leveling weights (`ui/enhancement_shaman/leveling_weights*.json`): Intellect is worth less (S3), Stormstrike less
  (S1/S2), Maelstrom Weapon procs less at 3.6 speed. Ranks 1-4 of Maelstrom Weapon are now worth far more (S4), and
  Rage of the Storm changed (S7). They need a re-fit.
- Bayovui stat weights for Enhancement: re-derive. Int has dropped to last at level 60.
- Leveling rotations (`tools/rotopt/leveling/*enh*`): re-search. Stormstrike now has an 8 s cooldown, there are fewer
  Maelstrom Weapon stacks, and Windfury Weapon now works with Windfury Totem.

## Needs a user ruling

1. `wf_weapon_totem_icd`: separate cooldowns for the weapon's and the totem's Windfury procs is an assumption.
2. `orchestration/standards.md` 25 still records the old rule, that Windfury Weapon turns the totem off. Amend it?
3. Other classes' 62 DPS builds are still the Nov 4 sims, though all have moved since, mostly from the Windfury Totem
   white-swing change (0ce9dbf; Arms 1088 -> 888). Tank and heal modes are stale too (about 15 min and 3 h). Re-sim them?
4. compsearch m1_10_class_buffs, m1_20, m1_20_class_buffs and m1_40_class_buffs have never been stamped, so it can't
   tell whether their builds changed.
5. Raising the Enhancement base gear's Intellect share is held by standards 5a/5b. It would gain under 1 DPS.
6. Rage of the Storm's override is a hand entry in `forever_seen_in_game.json`. A BayovCore sync that doesn't keep hand
   entries would drop it.
7. Five unused rotation explanation notes remain, one of them Enhancement's.
8. These packages fail identically on origin/master, from older expected results: hunter, mage, priest/healing,
   priest/shadow, rogue/dps_rogue, warlock/dps, warrior/dps_warrior.
