# Rotations pass 59 (2026-11-11, master 707ae92cc)

Last audited master 27d0d59d6. 59 commits; those that reach the raid presets:

| commit | change | reaches raid presets |
|---|---|---|
| cb651085c, 54a15ba24, 5dffd2206 | Shaman Nov 10 tuning (Stormstrike 100% / 8 s, Mental Dexterity 33/66/100%, Maelstrom 5+rank PPM, imbue + weapon totem both work); Enhancement talents and rotations re-optimized | yes (Enhancement) |
| 5607180c7, 546d0dcaa, 99c8ce2b1 | Affliction rotation after hotfix 113766 (Bane of Agony, Corruption recast 2.5 s early, Wrack unconditional), Wrack -30%, Imp Firebolt | yes (Affliction) |
| f20d5938a | no dot tick after the aura ran out when expiry is processed late | yes (Affliction, small) |
| 032692409 | natural flasks 30 Stamina, 2% zone bonus (70862) | yes (warrior, rogue, hunter, Feral) |
| 30e67a529, 6feec7094 | Arcane Shot / Summon Hawk 11% base mana; hawk melee +80% | yes (hunters) |
| 5bfae1163 | Arcane Concentration chance split across an Arcane Missiles channel | yes (Arcane) |
| a34d22451 | Sweeping Strikes keeps charges with no second target | no damage change single-target |
| 333643570 | Idol of Ferocity -2 energy Claw/Rake | no (Feral presets wear 220606) |
| 9628e85b6 | Nature's Swiftness makes Wild Growth instant | no (healing) |

`apl_logic_diff.py 27d0d59d6 707ae92cc`: 6 APL files, all from the two re-optimizations above
(Enhancement optimized / mana_tide / magma; the three Affliction APLs).

## Baseline move (3000 it, seed 424242)

New baseline `ref/rotations/baseline_707ae92_3000it_seed424242.tsv` (replaces `baseline_ca5e35c_...`).

| group | old -> new at 180 s | cause |
|---|---|---|
| Enhancement (4 builds) | 1041.6-1144.7 -> 928.9-1030.8 (-113 to -123) | Nov 10 tuning (notes-only values, standards 37) |
| Affliction (3) | 1018.4-1035.2 -> 1052.2-1071.0 (+34 to +37) | new rotation after 113766 |
| Fury DW / 2H (6) | -35 to -38 | flasks + items |
| Arms (4), rogue (6), Feral (2), BM/MM/LW (7) | -16 to -23 | natural flasks |
| Survival (3) | -26 to -28 | flasks, hunter mana |
| Arcane / Fire | -8.3 / -4.1 | Arcane Concentration split |

Everything else moved under 0.5.

## Checks this pass (paired `ab.py`, 20k it, 8 blocks, shared CLI 707ae92cc50b)

**Enhancement imbue with the group's Windfury Totem (pass-doc open item).** Standard 25 now says imbue and
weapon totem both work. `ab.py --req_b` (new option: A/B of two settings files) on the req vs the req
plus `raid.buffs.totemWeaponBuff: Windfury`, and vs Flametongue Weapon + Windfury Totem. B-A at 120/180/300 s:

| preset | variant | S0 1e6 | S0 2e8 | pooled at 180 |
|---|---|---|---|---|
| enhancement-level60 | + raid WF Totem | +1.08 / +0.44 / +2.28 | +0.65 / +0.04 / +2.42 | +0.24 |
| enhancement-level60 | FT Weapon + WF Totem | -103.1 / -106.2 / -106.0 | -103.4 / -106.3 / -106.0 | -106.2 |
| enhancement-mana-tide | + raid WF Totem | -0.21 / -0.09 / +1.35 | +0.90 / -0.08 / +1.76 | -0.09 |
| enhancement-mana-tide | FT Weapon + WF Totem | -94.0 / -93.0 / -92.2 | -92.2 / -92.6 / -91.4 | -92.8 |

The Enhancement APLs drop their own Windfury Totem (10614) once a fight, and since cb651085c it procs next to
Windfury Weapon (16362 attacks in the metrics), so the group's totem adds nothing; the +2 at 300 s is the
gap before the own totem goes down. Windfury Weapon beats Flametongue Weapon by ~100 DPS at both. The open
item is settled for the current gear; standards 40 still asks for both pairings when endgame gear changes.

**Affliction (new rotation, A 1071.1 at 180 s), S0 1e6:**

| variant | 120 / 180 / 300 s |
|---|---|
| Wrack only with Corruption and Agony up (old gate) | -3.48 / -2.71 / -1.88 |
| Wrack only with Corruption >= 6 s left | -7.45 / -11.05 / -19.58 |
| Corruption recast at 1.5 s instead of 2.5 s | -8.11 / -8.17 / -4.56 |

All lose; the knob-searched rotation holds.

**Hunter mana after Arcane Shot's 11% cost (MM 2.5 s, LW 3.1 s OOM a fight in the shared run).** Gate Arcane
Shot (14287) on `currentMana >= remainingTime x 10`, S0 1e6: MM 0.00 / -0.57 / -4.48, LW 0.00 / -0.47 / -1.88.
Loses; unconditional Arcane Shot holds.

## Open findings

- rotations-20261109-1 (Arms / Fury 2H Heroic Strike 90-rage line) and rotations-20261105-1 (Feral Ferocious
  Bite gate 65): not landed (no warrior or Feral APL change). Not re-measured; the flask change moves both
  builds' baselines uniformly. Idol of Ferocity (333643570) isn't worn by the Feral presets.

## Optional

None new.

## Left for a later pass

- Balance Night Elf / Elune's Light vs Tauren, Elemental Mastery hold, warlock duty-curse screens (unchanged).
- Enhancement water-shield and imp-gw weren't run with the raid WF Totem; they share the APL's own totem,
  so the null result should carry over (estimate).
