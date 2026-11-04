# Round 70753 phase 5, hit trade: Priest

Brief `briefs/phase5_hit_trade_class.md`. Builds: `priest__shadow` and `priest__shadow-holy`. `priest__smite` is left
alone because it has no gear hit to trade. Gear set: `ui/shadow_priest/gear_sets/level60_phase1_synthetic` (steps
`_hit-1` and `_hit+1` to `_hit+5`, trading into SpellPower and crit). Hit flags: `-hit-cap spell:16 -hit-school shadowFocus:1`.

**Result: neither build changes talents or gear.** Both stay on step +0, the current gear set itself. Their total Shadow
hit there is 16% against the 16% cap: 10 from gear, 1 from the Raiments 2-piece and 5 from Shadow Focus 5/5. Trading
Shadow Focus points for gear hit loses about 1.5 DPS (Shadow) or 2.2 DPS (Shadow + Holy) for each point moved (table
below). So the old and new numbers come from the same build and the same gear, and they match exactly.

## priest__shadow (Undead)

- Old talents: `00530400102-3-50052101231302251`. New talents: the same, with no changes.
- Gear step: +0 (`level60_phase1_synthetic.gear.json`). Shadow hit is 16 / 16 before and after.
- The search starts at 946.8 combined (3,000 iterations) and ends at 946.5 (20,000 to confirm). No move gains more
  than its noise. The best swaps are Dark Inspiration for Blackout, Spirit Tap, Silence or Shadow Reach, at +0.0 after
  refinement.

| | 120 s | 180 s | 300 s |
|---|---|---|---|
| old (current talents, current gear) | 944.7 ± 0.2 | 948.7 ± 0.2 | 946.0 ± 0.1 |
| new (same talents, `-hit-step +0`) | 944.7 ± 0.2 | 948.7 ± 0.2 | 946.0 ± 0.1 |

Shadow Focus against gear hit: scored on the 120/180/300 mean at 20,000 iterations. Each Shadow Focus point moves to
Blackout and is replaced by a gear hit step.

| Shadow tree | Shadow Focus | step | combined |
|---|---|---|---|
| `50052101231302251` | 5 | +0 | **946.5** |
| `41052101231302251` | 4 | +1 | 945.0 |
| `32052101231302251` | 3 | +2 | 943.7 |
| `05052101231302251` | 0 | +5 | 939.0 |

## priest__shadow-holy (Undead)

- Old talents: `-3025011303-50052101231302251`. New talents: the same, with no changes.
- Gear step: +0 (`level60_phase1_synthetic.gear.json`). Shadow hit is 16 / 16 before and after.
- The search ran with Twilight Focus pinned (`-require twilightFocus`). Like round 70753's search, it moves the Holy
  points into Discipline: it ends at `01500500103-3-54042001211302251` (947.3 combined), a Discipline and Shadow build
  with 3 Holy. That is the Level 60 Shadow build's idea, so the variant keeps its Holy talents. The search never moves
  Shadow Focus, and its step stays +0 all the way. The Shadow Focus check on the variant's own Holy points (table below)
  also keeps Shadow Focus at 5.

| | 120 s | 180 s | 300 s |
|---|---|---|---|
| old (current talents, current gear) | 902.3 ± 0.2 | 906.6 ± 0.2 | 873.9 ± 0.2 |
| new (same talents, `-hit-step +0`) | 902.3 ± 0.2 | 906.6 ± 0.2 | 873.9 ± 0.2 |

| Shadow tree | Shadow Focus | step | combined |
|---|---|---|---|
| `50052101231302251` | 5 | +0 | **894.3** |
| `41052101231302251` | 4 | +1 | 892.0 |
| `32052101231302251` | 3 | +2 | 889.8 |
| `05052101231302251` | 0 | +5 | 883.3 |

## Commands

Run from the repo root, with `G=ui/shadow_priest/gear_sets/level60_phase1_synthetic` and
`H="-hit-steps $G -hit-cap spell:16 -hit-school shadowFocus:1"`. The binary is `go build --tags=with_db ./tools/rotopt`,
built after `protoc -I=./proto --go_out=./sim/core ./proto/*.proto`.

    # searches
    rotopt -template priest_shadow -settings tools/dps_report/req/priest__shadow.json -combined -durs 120,180,300 \
        -talent-search -require silentResolve -iters 3000 -confirm-iters 20000 $H
    rotopt -template priest_shadow -settings tools/dps_report/req/priest__shadow-holy.json -combined -durs 120,180,300 \
        -talent-search -require twilightFocus -iters 3000 -confirm-iters 20000 $H
    # Shadow Focus check (one talent string per line, as in the tables)
    rotopt -template priest_shadow -settings tools/dps_report/req/priest__<id>.json -combined -durs 120,180,300 \
        -talents-file sf.txt -iters 20000 $H
    # confirms, d in 120 180 300
    rotopt -template priest_shadow -settings tools/dps_report/req/priest__<id>.json -eval -confirm-iters 20000 -durs $d
    rotopt -template priest_shadow -settings tools/dps_report/req/priest__<id>.json -eval -confirm-iters 20000 -durs $d \
        $H -hit-step +0

## Notes

- The `_hit-1` step is never needed: no priest build is over the cap. It exists for a Shadow Focus 5/5 build that has
  racial hit, such as a Dwarf (`hit_steps.json`).
- Nothing for the phase 5 session to regear on the priest side.
