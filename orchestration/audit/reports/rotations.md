# Rotations pass 69 (2026-11-18, master 2f7e255af)

Last audited master: 5f20dfdf4. Result: quiet (no new finding). One finding still open
(rotations-20261112-1, rogue poison swap: every rogue req still has Dense Sharpening Stones on both hands).

## What changed (5f20dfdf4..2f7e255af, 26 commits)

`apl_logic_diff.py`: 1 APL file, `ui/hunter/apls/bm.apl.json` (3 weave lines rewritten).

| commit | change | reaches presets? |
|---|---|---|
| 084813024 | BM weaves Raptor Strike on every main-hand swing, queued when ready (weaveRaptor=0) | BM, BM-wolf, BM-tallstrider |
| 4e81669a9 | Shieldsworn Holy Strike gets SpellFlagNoWindfuryTotem; Vampiric Embrace heals 10%/20% | tanks/healers only (no DPS preset takes Shieldsworn) |
| 3ec610c40, 52596be22, fa03a8689, 16b197c5b | Darkmoon Blue Dragon 3%, Heroism 158-238, Bad Hat, Militant's Band, Nov 13 item hotfixes | no DPS preset moved |
| 056dccb4e | Cannibal's Cleaver +25 AP vs Humanoids, Draconian Aegis +9 SP vs Dragonkin | no preset wears them |
| 192cb2e06 | Idol of Rejuvenation, Regrowth bonus | healers only |

## Baseline (3000 it, seed 424242)

`baseline.py` on 2f7e255af vs `baseline_5f20dfd`: only the BM presets moved past 0.05 DPS.

| preset | 5f20dfd | 2f7e255 | delta |
|---|---|---|---|
| hunter__beast-mastery | 1113.87 | 1115.41 | +1.54 |
| hunter__beast-mastery-wolf | 1065.33 | 1066.69 | +1.36 |
| hunter__beast-mastery-tallstrider | 1056.00 | 1056.59 | +0.59 |

New baseline: `ref/rotations/baseline_2f7e255_3000it_seed424242.tsv` (old one removed).
No `regress.py` needed: the sim changes reach no DPS preset, confirmed by the baseline.

The shared run's `vs_published.txt` shows Affliction +3.5%, Lone Wolf -1.7% and the sword/dagger
rogues -1 to -2% against the published report. None of it moved in this window (my baseline is flat
there), so it's the published report lagging master; that's the results/fairness auditors' ground.

## Verified the merged BM weave

Spot-check on an independent seed set (the worker used seeds 1 and 1000003), new rotation (A) vs
the 5f20dfd rotation (B), `ab.py`, 20k iterations, s0 = 2e8:

| | 120 s | 180 s | 300 s |
|---|---|---|---|
| old - new | -0.98 ± 0.21 | -0.85 ± 0.21 | -2.21 ± 0.21 |

The new weave wins everywhere; it stands.

## Windfury Totem ability timing (open item, partly done)

New helper `tools/audit/rotations/wf_gate.py` ANDs "main-hand swing speed - time to next swing
>= 0.1 s" onto every cast of the given spells (keeps the ability out of the totem's 100 ms after a
swing). Paired `ab.py`, 20k it, both seed sets, B - A:

| preset, ability | set | 120 s | 180 s | 300 s |
|---|---|---|---|---|
| warrior__fury, Bloodthirst 23894 | 1e6 | -0.40 ± 0.32 | -0.00 ± 0.29 | +0.24 ± 0.38 |
| | 2e8 | -0.62 ± 0.50 | +0.46 ± 0.30 | -0.43 ± 0.37 |
| rogue__combat-swords, Sinister Strike 11294 | 1e6 | -0.60 ± 0.39 | -0.37 ± 0.25 | -0.41 ± 0.25 |
| | 2e8 | +0.17 ± 0.41 | +0.08 ± 0.25 | -0.14 ± 0.20 |
| warrior__arms, Mortal Strike 21553 | both | +0.00 | +0.00 | +0.00 |

Nothing near the bar. Arms reads exactly 0: the gated MS line never fires inside the window
(or MS is cast from another action); not dug into. Still to do: Enhancement Stormstrike held for
Windfury Weapon's cooldown, Heroic Strike, Backstab, Paladin Holy/Crusader Strike and Judgement of
Command, feral Claw/Shred, Raptor Strike, and the "spread abilities between swings" try.

## Open items carried

- Rogue poison swap (rotations-20261112-1): not landed.
- Warrior HS/Cleave thresholds: re-search on the next warrior rotation search (pass 68 already
  tried Fury DW 50/70, 2H 75, Arms 75: all lose).
- Others as in `passes/rotations.md` (Balance Night Elf, Elemental Mastery hold, Enhancement
  Flametongue Weapon with Windfury Totem, warlock duty-curse screens).

## Optional

None this pass.
