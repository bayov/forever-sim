# Phase 5 hit trade: Warrior

Branch `phase5-hit-warrior` from `phase5-hit-trade` (aed188a4b). Brief: `briefs/phase5_hit_trade_class.md`.

## Result

**No Warrior build changes talents or gear step.** Each of the four base builds' talent searches (with the hit flags,
mean of 120/180/300 s, from the current talents and knobs) ends on its current talents, on step +0, the set as it is.
The duty builds keep their base build's talents, so they are unchanged too. A Warrior has no talent that adds
yellow melee hit: Furious Precision 3/3 adds off-hand hit only, and so Arms, Fury 2H and Dragonkin sit at 8.00% on 8%.
Fury Dual Wield's set carries 10% (cap step -2). The search, with `-hit-span 2`, tried -3..0 (no -4 exists) and kept +0:
white swings still pay for hit past the yellow cap.

**Odd: Fury Dual Wield wants more hit than the span can reach.** At 20,000 iterations with its current talents,
step +1 and +2 beat +0 at 120 and 180 s, by more than 2 standard errors of the difference, and +2 is the top step
there is. The span is centered on the cap step (-2), so +1 and +2 lie outside -4..0 and the search never tried them.

| Fury DW step (gear hit) | 120 s | 180 s | 300 s |
|---|---|---|---|
| -2 (8%) | 1192.5 (0.6) | 1150.5 (0.4) | 1126.5 (0.3) |
| -1 (9%) | 1195.2 (0.6) | 1152.1 (0.4) | 1127.0 (0.3) |
| +0 (10%, current) | 1196.2 (0.5) | 1153.0 (0.4) | 1128.7 (0.3) |
| +1 (11%) | 1197.9 (0.5) | 1154.3 (0.4) | 1129.3 (0.3) |
| +2 (12%) | 1198.6 (0.5) | 1155.1 (0.4) | 1129.4 (0.3) |

+2 against +0: +2.4 / +2.1 / +0.7. I kept +0 as the pick, because this is what the brief's flags give, and left the
call to the phase 5 session: wear `level60_phase1_synthetic_fury_hit+2` (and its Sunder and Battle Shout duties,
not simmed on +2 here), or add `_hit+3` and up to see where it stops. Simmed with the same talents, the talents
do not depend on the step (Furious Precision is in the build already, and nothing else in the trees moves hit).

## Per build

Talents are the same before and after; the step's gear set is the set itself (step +0), so the new runs wear the
same gear as the old and give the same numbers. Old = no `-hit-steps`; new = `-hit-steps ... -hit-step 0`.
DPS at 20,000 iterations, seed 1, standard error in brackets, old = new.

| Build | Talents (old = new) | Gear set (step +0) | Hit total / cap | 120 s | 180 s | 300 s |
|---|---|---|---|---|---|---|
| warrior__fury | `30313113002-0505300551120100321` | `level60_phase1_synthetic_fury` | 10 / 8 yellow (white 10 MH, 20 OH / 27) | 1196.2 (0.5) | 1153.0 (0.4) | 1128.7 (0.3) |
| warrior__fury-duty-sunder | same as fury | same | same | 1156.5 (0.6) | 1117.7 (0.4) | 1095.4 (0.3) |
| warrior__fury-duty-battle-shout | same as fury | same | same | 1193.2 (0.5) | 1150.5 (0.4) | 1123.2 (0.3) |
| warrior__fury_2h | `30313113032-2505001501122100321` | `level60_phase1_synthetic_arms` | 8 / 8 | 1214.5 (0.6) | 1173.3 (0.5) | 1135.3 (0.4) |
| warrior__fury_2h-duty-sunder | same as fury_2h | same | 8 / 8 | 1175.0 (0.6) | 1134.6 (0.5) | 1104.4 (0.4) |
| warrior__fury_2h-duty-battle-shout | same as fury_2h | same | 8 / 8 | 1213.9 (0.6) | 1168.8 (0.5) | 1133.8 (0.4) |
| warrior__arms | `30303213032215201-050500050012-2` | `level60_phase1_synthetic_arms` | 8 / 8 | 1111.7 (0.6) | 1097.7 (0.5) | 1070.6 (0.4) |
| warrior__arms-duty-sunder | same as arms | same | 8 / 8 | 1055.0 (0.6) | 1047.1 (0.5) | 1027.5 (0.4) |
| warrior__arms-duty-battle-shout | same as arms | same | 8 / 8 | 1111.0 (0.6) | 1094.0 (0.5) | 1068.3 (0.4) |
| warrior__arms_dragonkin | `30303213132215201-050500050012-1` | `level60_phase1_synthetic_arms` | 8 / 8 | 1157.5 (0.6) | 1142.2 (0.5) | 1114.2 (0.4) |

The 180 s numbers equal each build's published `verify` DPS in its meta. Search ends (mean of 120/180/300 s, the
search refines at 10,000 iterations): Fury 1159.1 (0.4), Fury 2H 1173.9 (0.4), Arms 1092.7 (0.4), Dragonkin 1137.7
(0.4). In each, no one- or two-point move gained beyond its noise.

## Commands

`rotopt` = `go build --tags=with_db ./tools/rotopt` (needs `protoc` and `protoc-gen-go` for `sim/core/proto`).
`$K` = the knobs of that build's `gen` line in `tools/rotopt/warrior_presets.sh` (the line for the APL named in its meta).
`G=ui/warrior/gear_sets`.

Talent searches (pins: Bloodthirst for Fury, Mortal Strike for Arms, plus Spearing Strike for Dragonkin):

    rotopt -template warrior_fury -settings tools/dps_report/req/warrior__fury.json -talents 30313113002-0505300551120100321 -knobs $K \
      -hit-steps $G/level60_phase1_synthetic_fury -hit-cap melee:8 -hit-span 2 -require bloodthirst -talent-search -combined -durs 120,180,300
    rotopt -template warrior_fury -settings tools/dps_report/req/warrior__fury_2h.json -talents 30313113032-2505001501122100321 -knobs $K \
      -hit-steps $G/level60_phase1_synthetic_arms -hit-cap melee:8 -require bloodthirst -talent-search -combined -durs 120,180,300
    rotopt -template warrior_arms -settings tools/dps_report/req/warrior__arms.json -talents 30303213032215201-050500050012-2 -knobs $K \
      -hit-steps $G/level60_phase1_synthetic_arms -hit-cap melee:8 -require mortalStrike -talent-search -combined -durs 120,180,300
    rotopt -template warrior_arms -settings tools/dps_report/req/warrior__arms_dragonkin.json -talents 30303213132215201-050500050012-1 -knobs $K \
      -hit-steps $G/level60_phase1_synthetic_arms -hit-cap melee:8 -require mortalStrike,spearingStrike -talent-search -combined -durs 120,180,300

Confirms, per build `<id>` (template, talents, knobs as above; Fury DW and its duties take the fury set, the rest
the arms set), for d in 120, 180, 300:

    old: rotopt -template <t> -settings tools/dps_report/req/<id>.json -talents <talents> -knobs $K -eval -confirm-iters 20000 -durs $d
    new: the same plus -hit-steps $G/<set> -hit-cap melee:8 -hit-step 0
    Fury DW step sweep: the same as new with -hit-step -2, -1, +1, +2
