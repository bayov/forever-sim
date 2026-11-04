# Mage: hit-trade talent re-search (round 70753, phase 5)

Branch `phase5-hit-mage` from `phase5-hit-trade` (aed188a4b). Brief: `briefs/phase5_hit_trade_class.md`.

## Summary

| Build | Talents (old -> new) | Gear step | Old DPS 120 / 180 / 300 | New DPS 120 / 180 / 300 |
|---|---|---|---|---|
| mage__arcane (Gnome) | `050205203100311531-5551-003` -> `050205203100311531-5550000000001-003` | `level60_phase1_synthetic_arcane_hit-1` (-1) | 986.1(0.4) / 962.3(0.3) / 944.5(0.3) | 993.5(0.4) / 969.9(0.3) / 952.1(0.3) |
| mage__fire (Human) | `0502052-55500201230133031-003` unchanged | `level60_phase1_synthetic` (+0) | 1066.8(0.4) / 1057.9(0.4) / 1042.0(0.3) | unchanged |
| mage__frost (High Order) | `0502050030001-4-253500230111301211` unchanged | `level60_phase1_synthetic` (+0) | 944.5(0.6) / 944.7(0.5) / 937.7(0.4) | unchanged |
| mage__frostfire (Human, default) | `-2550030103013303-0235000311203002` unchanged | `level60_phase1_synthetic` (+0) | 1052.4(0.6) / 1064.3(0.6) / 1041.8(0.4) | unchanged |

20,000 iterations, seed 1, the report request at each length (±20 s), knobs from `tools/rotopt/mage_presets.sh`. The old
numbers reproduce the published ones exactly. Every mage build keeps Precision 3/3: the gear sets already counted it,
and dropping points of it for the hit back in gear loses at every rank (table below).

## Per build

### mage__arcane

- Talents: old `050205203100311531-5551-003`, new `050205203100311531-5550000000001-003`.
- Change in words: Flame Throwing 1 -> Master of Elements 1 (Fire). The search's first round took Burning Soul 1 for Flame
  Throwing (+0.6, under its noise), the mana tie-break then moved it to Master of Elements (more mana left at the same DPS).
- Gear step: -1, `ui/mage/gear_sets/level60_phase1_synthetic_arcane_hit-1.gear.json`. Total hit 17% (gear 10 as a Gnome's
  11, set 2pc +1, Precision +6) against the 16% cap before; 16/16 after. This step alone is the gain: the current talents
  on step -1 sim 993.2 / 969.3 / 951.3 (+7.1 / +7.0 / +6.8).
- Talent swap on step -1: +0.3 / +0.6 / +0.8 at 120 / 180 / 300 (each about 1 to 2 standard errors of the difference).
  Small; the gear step is what matters. The phase 5 session can take either talent string.
- DPS: old 986.1(0.4) / 962.3(0.3) / 944.5(0.3); new 993.5(0.4) / 969.9(0.3) / 952.1(0.3) (+7.4 / +7.6 / +7.6).
- Pins: `-require arcaneBlast,arcanePower`.

### mage__fire

- Talents unchanged `0502052-55500201230133031-003`; step +0 (`level60_phase1_synthetic.gear.json`), hit 16/16 before and
  after (gear 9, set +1, Precision +6). 996 sim runs, no move out of the noise.
- DPS 1066.8(0.4) / 1057.9(0.4) / 1042.0(0.3) both old and new. Pin `-require combustion`.

### mage__frost

- Talents unchanged `0502050030001-4-253500230111301211`; step +0, hit 16/16. 618 sim runs.
- DPS 944.5(0.6) / 944.7(0.5) / 937.7(0.4) both. Pin `-require iceBarrier`.

### mage__frostfire

- Talents unchanged `-2550030103013303-0235000311203002`; step +0, hit 16/16. 996 sim runs, no pins.
- DPS 1052.4(0.6) / 1064.3(0.6) / 1041.8(0.4) both.

## Precision checked directly

The search log prints only the best moves, so each build also scored every legal string with 1, 2 or 3 Precision points
moved to one other talent, wearing the cap step (`-talents-file`, 5000 iterations, mean of 120 / 180 / 300). Best string
per Precision rank:

| Build | Precision 3 | 2 | 1 | 0 |
|---|---|---|---|---|
| Arcane (steps -1 / +1 / +3 / +5) | 971.6 | 959.7 | 949.1 | 938.6 |
| Fire (steps 0 / +2 / +4 / +5*) | 1055.8 | 1054.4 | 1045.1 | 1024.6* |
| Frost (0 / +2 / +4 / +5*) | 941.9 | 939.5 | 937.0 | 916.8* |
| Frostfire (0 / +2 / +4 / +5*) | 1054.1 | 1047.7 | 1041.9 | 1017.5* |

\* Odd: Precision 0 on the shared set needs step +6 and `level60_phase1_synthetic_hit+6` doesn't exist (hit_steps.json
says up 6, the files stop at +5), so those builds sit at 15/16 hit. Even rank 1 (reachable, +4) loses 9 to 14 DPS, so a
+6 step wouldn't change the answer.

## Odd

- `-seed 2` gave the same numbers as seed 1 to the decimal in `-eval` (Arcane, all three lengths), so the second seed set
  isn't a real second check; the eval seems to use the request's own seed.
- The build needed `protoc` and `protoc-gen-go` (sim/core/proto isn't checked in); installed in the container.

## Commands

From the repo root, `rotopt` = `go run --tags=with_db ./tools/rotopt` (built once), `G=ui/mage/gear_sets`, knobs `$K_<spec>`
the `gen` lines of `tools/rotopt/mage_presets.sh`.

```sh
# talent search (each build)
rotopt -template mage_arcane -settings tools/dps_report/req/mage__arcane.json -knobs "$K_arcane" -talent-search \
  -combined -durs 120,180,300 -hit-steps $G/level60_phase1_synthetic_arcane -hit-cap spell:16 -require arcaneBlast,arcanePower
rotopt -template mage_fire -settings tools/dps_report/req/mage__fire.json -knobs "$K_fire" -talent-search \
  -combined -durs 120,180,300 -hit-steps $G/level60_phase1_synthetic -hit-cap spell:16 -require combustion
rotopt -template mage_frost -settings tools/dps_report/req/mage__frost.json -knobs "$K_frost" -talent-search \
  -combined -durs 120,180,300 -hit-steps $G/level60_phase1_synthetic -hit-cap spell:16 -require iceBarrier
rotopt -template mage_frostfire -settings tools/dps_report/req/mage__frostfire.json -knobs "$K_frostfire" -talent-search \
  -combined -durs 120,180,300 -hit-steps $G/level60_phase1_synthetic -hit-cap spell:16

# old (each build, d in 120 180 300)
rotopt -template mage_<spec> -settings tools/dps_report/req/mage__<spec>.json -knobs "$K_<spec>" -eval -confirm-iters 20000 -durs $d

# new: Arcane
rotopt -template mage_arcane -settings tools/dps_report/req/mage__arcane.json -knobs "$K_arcane" \
  -talents 050205203100311531-5550000000001-003 -eval -confirm-iters 20000 -durs $d \
  -hit-steps $G/level60_phase1_synthetic_arcane -hit-cap spell:16 -hit-step -1
# new: Fire, Frost, Frostfire (current talents)
rotopt -template mage_<spec> -settings tools/dps_report/req/mage__<spec>.json -knobs "$K_<spec>" -eval -confirm-iters 20000 \
  -durs $d -hit-steps $G/level60_phase1_synthetic -hit-cap spell:16 -hit-step +0

# Precision check: strings from a one-off script (each legal move of 1-3 Precision points to one talent)
rotopt -template mage_<spec> -settings tools/dps_report/req/mage__<spec>.json -knobs "$K_<spec>" -talents-file <file> \
  -combined -durs 120,180,300 -iters 5000 -hit-steps $G/<set> -hit-cap spell:16
```
