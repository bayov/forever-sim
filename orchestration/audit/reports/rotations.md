# Rotations pass 36 (2026-10-28)

- **Master:** 5d27b0015 (last audited 37a8a8ccd). Client 1.60.1.70634 data; post 8 changes are server side.
- **Result:** quiet, 0 new findings. No open findings carried.

## What changed

- Non-test sim `.go` (`sim/core/racials.go`, `racials_forever.go`), from dev notes post 8:
  Blood Fury 3 → 2 min, Berserking 10 → 15 s (Forever only), Touch of the Grave ICD 25 → 15 s.
- `apl_logic_diff.py 37a8a8c 5d27b00`: 6 APL files changed, all worker-measured race tuning:
  - Balance `level60`, `level60_ff`: Elune's Light (1259799) pre-pull at −0.1 s and synced to the trinket (Night Elf).
  - Ret `level60_shock*` (3 files): Seal/Judgement mana gates rewritten, Hammer of Wrath / Consecration order.
  - Rogue `level60_combat_daggers`: Eureka! (1259812) gated on Slice and Dice ≥ 8 s left (inert on Undead).
- Req files: race re-picks (now 28 Undead, 15 Orc, 5 Night Elf, 4 Human, 4 Windshaper, 1 Gnome,
  1 High Order). The req rotations carry each race's lines.

## Checks

- **Blood Fury use on the Orc presets** (shared run, 180 s, 20k): every Orc preset casts it 2 times per
  fight (hunters BM/MM/SV and warlocks on MCD autocast, Fury and Enhancement through their sync lines).
- **Fury and Enhancement sync lines still right at 2 min?** Paired A/B, sync line vs Blood Fury on cooldown
  (condition removed), 20k it, B−A:

| preset | seed set | 120 s | 180 s | 300 s |
|---|---|---:|---:|---:|
| warrior__fury | 1e6 | −2.61 ± 0.04 | −3.86 ± 0.02 | −1.18 ± 0.01 |
| warrior__fury | 2e8 | −2.65 ± 0.04 | −3.85 ± 0.02 | −1.20 ± 0.01 |
| shaman__enhancement-level60 | 1e6 | +0.00 | −0.04 | +0.33 |
| shaman__enhancement-level60 | 2e8 | +0.00 | −0.04 | +0.33 |

  The Death Wish sync keeps its value (+3.9 at 180 s, pooled); Enhancement's line is neutral. No change.
- **Berserking 15 s:** no preset is Troll; nothing to check.
- **regress.py** (37a8a8c CLI vs 5d27b00 CLI, master req files, 3k it, seed 424242): every preset moved up,
  +3.5 to +7.4 (Undead Touch of the Grave, Orc Blood Fury). No row lost, so no rotation is hurt by the change.

## Baseline moves (vs `baseline_06a8987`)

Every row moved up or by under 1; the biggest (race re-pick plus post 8):

| preset | old | new | Δ |
|---|---:|---:|---:|
| paladin__level_60_shock-duty-crusader | 865.4 | 891.7 | +26.2 |
| paladin__level_60_shock | 863.6 | 887.8 | +24.2 |
| hunter__beast-mastery | 1134.1 | 1150.4 | +16.3 |
| hunter__survival | 1016.1 | 1029.2 | +13.0 |
| druid__balance-wrath | 800.6 | 811.9 | +11.3 |
| hunter__marksmanship | 1126.0 | 1137.2 | +11.2 |

New reference: `ref/rotations/baseline_5d27b00_3000it_seed424242.tsv` (06a8987 file removed).

## Note for the results auditor

`runs/vs_published.txt`: Fury (−4.4), Enhancement Mana Tide/Water Shield (−3.5) and Frost (−35.5) sit below
their published numbers, but regress.py shows they gained in this window, so their published values are
stale (set before the 2026-10-26 racial pass), not a rotation regression.

## Optional (under the bar)

None new. Carried: see `passes/rotations.md`.
