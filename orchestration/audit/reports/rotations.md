# Rotations pass 75 (2026-11-24, master eceb09019)

Quiet: the Level 60 APL changes in this range all beat the APLs they replaced, on both seed sets, on the new Phase 0
gear. No new findings. The Phase 0 regear moved most presets by 6 to 20%, and only some classes had their rotations
re-searched on it (open item below).

## What changed (e29cfe00f..eceb09019, 84 commits)

- APL logic (`apl_logic_diff.py`, 13 files):
  - `shadow_priest/level60_smite` (36ac6f31d): spell 1277328 below 60% mana (was 40%), rank 8 Smite gated on
    mana/20 >= time left (was /35), fallback rank 3 Smite (598) in place of rank 7 (10933). This retires the settled
    "Smite rank 7 fallback and 35 mana a second" line in the pass doc for the Phase 0 gear.
  - `warlock/level60_affliction*`, `level60_demonology*` (e47cbeea4): Immolate (11668) removed.
  - `warrior/level60_fury*` (841b2a21a): Execute moved ahead of Bloodthirst.
  - `tank_warrior/level60_*` (8d5fdd664, dfd015925): racials added, Stoneform from 20 s. Tank presets, not DPS.
- `tools/dps_report/req`: 26 commits of Phase 0 step 4 regear (gear, talents for Affliction, race sets).
- Sim `.go` (non-test): racials (Troll Regeneration, Stoneform, Blood Fury adds SpellDamage), Innervate/Energized scale only
  the MP5 share (EV-0114), shaman class row and Skyborne base offsets, Brain Hacker debuff 30 s. None needs a rotation change.

## New APLs vs the ones they replaced (paired `ab.py`, 20k it, 8 blocks; A = new, B = old APL; same new req)

| preset (A at 180 s) | S0 1e6: B-A 120 / 180 / 300 | S0 2e8: B-A 120 / 180 / 300 | verdict |
|---|---|---|---|
| priest smite (695.4) | -11.5 / -7.5 / -25.4 | -11.4 / -7.6 / -25.2 | new wins |
| warlock affliction (966.1) | -30.4 / -34.7 / -51.2 | -30.3 / -34.5 / -51.3 | new wins |
| warlock demonology (894.0) | -15.0 / -14.3 / -14.7 | -14.8 / -13.7 / -13.8 | new wins |
| warrior fury (1007.2) | -5.1 / -3.7 / -4.6 | -4.2 / -3.8 / -4.3 | new wins |

All SE under 0.6. The old warlock APLs also carry the old line order, so the Immolate share of the gain is not isolated.

## Baseline (3000 it, seed 424242), largest moves vs `baseline_566067a`

| group | move |
|---|---|
| warlock (all 13) | -11 to -20% (affliction_shadow_bolt -208, destruction -186) |
| mage, balance, elemental, Fury 2H | -12 to -16% |
| feral, Fury, paladin, shadow, smite | -8 to -11% |
| enhancement, arms | -6 to -8% |
| rogue, hunter | under 0.5 DPS |

These follow the req regear, not the sim code. New file: `ref/rotations/baseline_eceb090_3000it_seed424242.tsv`.

## Dead lines and resources (`dead_lines.py` on `runs/master`)

- No OOM over 3.1 s (Lone Wolf 3.1, MM 2.5, affliction_shadow_bolt 2.8); warlock Demo/Destruction 1.4-1.7 s.
- Item lines now inert because the Phase 0 sets dropped the item: 991018 (all Fury), 996032 (Fire, Frostfire),
  997022 (all warlock), 995034 (priests), 992015 (paladins), 998015 (balance), 21180 (rogues). Harmless; can be cleaned up.
- `dead_lines.py` reports `OtherActionPotion` as dead for every warrior, but the result shows Mighty Rage Potion (13442)
  cast twice a fight: the tool doesn't map the potion action to its item id. Tool limitation, not a rotation bug.

## Open items (carried and new)

- **New: Phase 0 rotations not re-searched.** Re-searched on the new sets: Smite, Balance Skyborne, warlock, Fury. Not
  re-searched: mage, paladin, shaman, feral, arms, shadow, Fury 2H. Lower gear can move mana and rage thresholds; screen them
  in the next pass that has time (fan-out per the pass doc).
- Carried from pass 74: warrior HS/Cleave thresholds, paladin hasted GCD, Windfury main-hand timing, Balance Night Elf,
  Elemental Mastery hold, Enhancement Flametongue imbue with Windfury Totem, warlock duty-curse screens.
- Open findings rotations-20261122-1 (Shock Ret) and rotations-20261112-1 (rogue poisons) not landed: their APL and req
  files are unchanged in this range.
