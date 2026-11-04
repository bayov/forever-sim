# Phase 5 hit trade: Warrior (on the hotfix 113539 tree)

Branch `phase5-hit-warrior` at `origin/phase5-hit-trade` 1e8385a (master with the Warrior hotfix 113539 merged, the
hit-trade tool, the rebuilt db.bin). Brief: `briefs/phase5_hit_trade_class.md`, and the phase 5 coordinator's
follow-up: redo on the new Fury tree from the hotfix's picks with the reqs' current knobs. (The first result, on the
old tree, is in this file's history: no change, Fury DW +1/+2 about 2 DPS better than +0.)

## Result

- **Talents: no change anywhere.** Each base build's talent search (hit flags, mean of 120/180/300 s, from the
  hotfix picks and the knobs in `tools/rotopt/warrior_presets.sh`) ends on its starting talents. The duty builds
  keep their base build's talents.
- **Arms, Fury 2H, Dragonkin and their duties: step +0**, the set as it is. A Warrior has no talent that adds
  yellow hit (Furious Precision is off-hand only), so these sit at 8.00% on the 8% cap; no step to trade.
- **Fury Dual Wield and both its duties: step +2, `ui/warrior/gear_sets/level60_phase1_synthetic_fury_hit+2`**
  (12% gear hit, was 10%). The search, with `-hit-span 2` around the cap step -2, can only try -4..0 (-3..0 exist),
  and keeps +0 there. Steps +1 and +2 are outside its window; the sweep below finds +2 best, and the paired test
  keeps it.

Fury DW step sweep, current talents, 20,000 iterations, seed 1 (standard error in brackets):

| step (gear hit) | 120 s | 180 s | 300 s |
|---|---|---|---|
| -2 (8%) | 1176.1 (0.6) | 1133.8 (0.4) | 1109.8 (0.3) |
| -1 (9%) | 1177.1 (0.5) | 1135.6 (0.4) | 1110.6 (0.3) |
| +0 (10%, current) | 1179.5 (0.5) | 1136.0 (0.4) | 1111.6 (0.3) |
| +1 (11%) | 1179.6 (0.5) | 1137.6 (0.4) | 1111.9 (0.3) |
| +2 (12%, the top step) | 1181.7 (0.5) | 1137.7 (0.4) | 1111.9 (0.3) |

Paired test, +2 against +0: 10 seeds 1,000,003 apart (1, 1000004, ..., 9000028), 2,000 iterations each (20,000 per
step and length), both steps on the same seed; mean of the 10 per-seed differences, standard error from their spread:

| | 120 s | 180 s | 300 s | mean of the three |
|---|---|---|---|---|
| +2 minus +0 | +3.48 ± 0.59 (5.9 SE) | +2.68 ± 0.82 (3.3 SE) | +0.88 ± 0.53 (1.7 SE) | +2.35 ± 0.38 (6.2 SE) |

+2 clears 2 SE at 120 and 180 s and on the three-length mean the search scores on; at 300 s it gains but by 1.7 SE,
so not proven there (never a loss). I keep +2. If the rule is meant per length, 300 s alone does not pass: the
coordinator's call. Per-seed differences: 120 s +2.0 +6.9 +3.7 +3.9 +5.1 +1.8 +1.9 +5.6 +2.5 +1.4; 180 s +0.6 +3.8 +0.4
+2.5 +3.1 -2.4 +5.1 +3.8 +6.8 +3.1; 300 s -0.8 -2.5 +3.1 +2.0 +1.0 -0.5 +2.1 +1.3 +1.4 +1.7.

Odd, for the tool: (1) the span is centred on the cap step, so a dual wielder above its cap can't reach the steps
past the set (+1, +2); `-hit-span 4` would cover -6..+2. (2) +2 is the top step that exists and still gains at the
short fights, so `_hit+3` and up might gain more. (3) `-seed` with neighbouring values (1, 2, 3...) gives nearly the
same fights (the means moved 0.2 against a standard error of 1.4): seeds for a paired test must be far apart.

## Per build

Old = the hotfix picks on step +0, no `-hit-steps`. New = the pick with `-hit-steps <set> -hit-cap melee:8 -hit-step
<step>`. 20,000 iterations, seed 1. Old 180 s = each meta's published `verify` DPS. On step +0 new = old exactly
(the same gear file; run and checked for the seven Arms-set builds).

| Build | Talents (old = new) | Step, gear set | Hit total / cap, old -> new | Old 120 / 180 / 300 | New 120 / 180 / 300 |
|---|---|---|---|---|---|
| warrior__fury | `30313113002-050530055120100231-1` | +2, `level60_phase1_synthetic_fury_hit+2` | yellow 10 -> 12 / 8; white MH 10 -> 12, OH 20 -> 22 / 27 | 1179.5 (0.5) / 1136.0 (0.4) / 1111.6 (0.3) | 1181.7 (0.5) / 1137.7 (0.4) / 1111.9 (0.3) |
| warrior__fury-duty-sunder | same | +2, same | same | 1137.7 (0.5) / 1097.6 (0.4) / 1076.0 (0.3) | 1141.5 (0.5) / 1100.4 (0.4) / 1077.8 (0.3) |
| warrior__fury-duty-battle-shout | same | +2, same | same | 1169.0 (0.5) / 1124.2 (0.4) / 1096.7 (0.3) | 1171.8 (0.5) / 1126.0 (0.4) / 1097.7 (0.3) |
| warrior__fury_2h | `30313113032-350500150122100231` | +0, `level60_phase1_synthetic_arms` | 8 / 8 | 1192.5 (0.6) / 1151.0 (0.5) / 1117.1 (0.4) | same |
| warrior__fury_2h-duty-sunder | same | +0 | 8 / 8 | 1153.2 (0.6) / 1115.2 (0.5) / 1086.5 (0.4) | same |
| warrior__fury_2h-duty-battle-shout | same | +0 | 8 / 8 | 1190.6 (0.6) / 1147.6 (0.5) / 1114.6 (0.4) | same |
| warrior__arms | `30303213032215201-05050005012-2` | +0, `level60_phase1_synthetic_arms` | 8 / 8 | 1101.0 (0.6) / 1088.0 (0.5) / 1062.3 (0.4) | same |
| warrior__arms-duty-sunder | same | +0 | 8 / 8 | 1046.7 (0.6) / 1040.0 (0.5) / 1021.5 (0.4) | same |
| warrior__arms-duty-battle-shout | same | +0 | 8 / 8 | 1100.8 (0.6) / 1084.4 (0.5) / 1060.1 (0.4) | same |
| warrior__arms_dragonkin | `30303213132215201-05050005012-1` | +0, `level60_phase1_synthetic_arms` | 8 / 8 | 1160.8 (0.6) / 1147.0 (0.5) / 1120.6 (0.4) | same |

The duties' +2 numbers are single seed-1 runs, not paired. Search ends (mean of 120/180/300 s at 10,000 iterations):
Fury DW 1141.9 (0.4), Fury 2H 1153.0 (0.4), Arms 1083.5 (0.4), Dragonkin 1142.9 (0.4); no move gained beyond noise.

## Commands

`rotopt` = `go build --tags=with_db ./tools/rotopt` (after `protoc -I=./proto --go_out=./sim/core ./proto/*.proto`).
`$K` = the knobs on that build's `gen` line in `tools/rotopt/warrior_presets.sh` (its meta's `apl`).
`G=ui/warrior/gear_sets`.

Talent searches (pins: Bloodthirst for Fury, Mortal Strike for Arms, plus Spearing Strike for Dragonkin):

    rotopt -template warrior_fury -settings tools/dps_report/req/warrior__fury.json -talents 30313113002-050530055120100231-1 -knobs $K \
      -hit-steps $G/level60_phase1_synthetic_fury -hit-cap melee:8 -hit-span 2 -require bloodthirst -talent-search -combined -durs 120,180,300
    rotopt -template warrior_fury -settings tools/dps_report/req/warrior__fury_2h.json -talents 30313113032-350500150122100231 -knobs $K \
      -hit-steps $G/level60_phase1_synthetic_arms -hit-cap melee:8 -require bloodthirst -talent-search -combined -durs 120,180,300
    rotopt -template warrior_arms -settings tools/dps_report/req/warrior__arms.json -talents 30303213032215201-05050005012-2 -knobs $K \
      -hit-steps $G/level60_phase1_synthetic_arms -hit-cap melee:8 -require mortalStrike -talent-search -combined -durs 120,180,300
    rotopt -template warrior_arms -settings tools/dps_report/req/warrior__arms_dragonkin.json -talents 30303213132215201-05050005012-1 -knobs $K \
      -hit-steps $G/level60_phase1_synthetic_arms -hit-cap melee:8 -require mortalStrike,spearingStrike -talent-search -combined -durs 120,180,300

Confirms, per build `<id>` with its template, talents and knobs, for d in 120, 180, 300:

    old: rotopt -template <t> -settings tools/dps_report/req/<id>.json -talents <talents> -knobs $K -eval -confirm-iters 20000 -durs $d
    new: the same plus -hit-steps $G/<set> -hit-cap melee:8 -hit-step <step>   (fury set and +2 for Fury DW and its duties; arms set and 0 for the rest)
    Fury DW sweep: new with -hit-step -2, -1, +1, +2
    Paired: new for warrior__fury with -hit-step 0 and +2, -confirm-iters 2000 -seed $((1 + k*1000003)) for k = 0..9
