# Rotations pass 47 (quiet: no findings)

- Master: 796495d6f (from b647a21ff), audited 2026-11-02. Client 1.60.1.70707, unchanged.
- New baseline: `ref/rotations/baseline_796495d_3000it_seed424242.tsv` (replaces `baseline_5a6e9c6_...`).

## What changed

| Change | Effect on Level 60 rotations |
|---|---|
| sim/core APL, cast, aura, mana, energy, raid damage edits (healing sim, tank sim, practice-room hooks, APL condition memo) | `regress.py` (old vs new CLI, same reqs, 3k it, seed 424242): no preset moved more than 1 DPS. Elemental +0.5, Elemental Mana Tide -0.9 (Lightning Overload copies are procs, 6265a9ab7). Demonology not comparable (new set unknown to the old CLI) |
| Rogue re-search (da5be5321): swing-timer gates dropped from builders, Combat Daggers drops Vanish/Ambush/Preparation, Seal Fate Daggers reordered | Spot-check: old Combat Daggers APL vs new, paired 20k: -1.73/-1.51/-1.74 (s0 1e6), -1.87/-1.69/-1.96 (s0 2e8) at 120/180/300 s. New wins |
| Warrior knob re-search (f62b4a582, effd5cd42): Fury potion drops its "first 3 s" branch, 2H Bloodrage < 10 rage, Slam 60, Arms-Sunder Slam above Mortal Strike | Spot-check: re-adding `currentTime <= 3s` to Fury's potion: -2.6/-1.6 at 120 s, 0 at 180/300 s. New wins |
| Hunter MM/LW Aimed Shot above 40% mana (bb8edc2fe) | +1.7 to +3.6 in baseline |
| Mage Fire/Arcane end-of-fight gates, Balance Night Elf at 36 yd with Elune's Light (32c825cff, fa2d49997) | Setting changes (standards 33). Balance Wrath -3.0, Frostfire -3.2 in baseline from the caster distances |
| Demonology synthetic set (be5126ccd) | +7.5 to +9.8 |

## Checks

- `apl_logic_diff.py b647a21ff 796495d6f`: 19 files with logic changes, all from the re-searches above. No wrong spell ids
  or dead lines spotted. Elune's Light (1259799) is the Forever Night Elf racial (`sim/core/racials_forever.go:254`);
  its -0.1 s pre-pull cast is a buff, not damage.
- Shared run vs published report: 0 of 73 moved more than 0.5 DPS.
- Open findings: none.

## Optional (under the bar)

Carried over, not re-measured: Mutilate window 0.5 s (the swing gates it tuned are gone now; probably moot), Fury-BS HS 55,
hunter end-of-fight gates, Shadow Mind Blast time-left gate, Destruction Fire DoT gates, Water Shield Magma gate.
