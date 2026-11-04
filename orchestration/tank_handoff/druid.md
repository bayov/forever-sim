# Feral tank (bear) handoff, 2026-11-04

Branch `worktree-agent-afa0905bceb681edc` (merged `claude/clever-galileo-qim1ee` at 659cad198).

## Done

- Task A, the 70753 realignment. The tests in `sim/druid/tank/bear_rules_test.go` use the notes' Bear Form
  threat x2.15 and Rend and Tear 3%. The TestP1Balance and TestP1Feral tank-set goldens moved because the
  merge kept their pre-round values: they now move by the same factor as every other synthetic set's lines.
  This was committed as 3013df6ed. The audit is redone in `docs/tank_druid.md` ("After the 70753 update").
  Balanced moved from 1,736.2 to 1,833.6 (six-encounter mean, 3,000 iterations, old against new code, same
  build).
- Step 6. The talent and knob searches at 600 iterations are confirmed at 3,000, and the race check is done.
  - Talents: `-5022233202310212551-055` at every weight (Furor 5 moved to Shredding Attacks 3 and Brutal
    Impact 2).
  - Knobs, Balanced: `roarRefresh=3,primalBite=1,lacerate=2,lacerateRefresh=4,maulRage=30,enrage=1,enragePrepull=1,prepullShift=1,barkskin=2`.
  - Knobs, Max threat: the same with maulRage=20 (2,910.6 at 3,000 iterations).
  - Knobs, Max survival: the same with frRage=100 and enragePrepull=3 (-857.1 at 3,000 iterations).
  - Race: Night Elf is best at every weight.
  - Outputs: `tools/tankbench/req/druid_bear_{max_threat,balanced,max_survival}.json` (`druid_bear.json` is
    removed and `tank_test.go` now reads the balanced one), `ui/feral_tank_druid/apls/*.apl.json`,
    `presets.ts` and `sim.ts`. `npx tsc --noEmit -p .` is clean for the bear; it fails only in
    `ui/tank_warrior/presets.ts`.
- Step 7. Tankbench ran on all seven encounters at 1,000 iterations. The table and the calls for the user are
  in `docs/tank_druid.md` ("Presets", "Calls for the user").
- `go test -tags=with_db ./sim/druid/...` passes.

## Left

- The full `go test -tags=with_db -timeout 40m` over `./sim/...` (without sim/practice and sim/web) was
  stopped early, and only its first packages ran. Run it, then check that `sim/testdata/report_hashes.txt`
  did not move.
- Known failure that isn't the bear's: sim/core TestForeverConsumeProcDamage ("Shadow Oil: no proc spell
  1382").
- Still name `druid_bear.json`: `docs/tank_sim.md:514` (shared), `TASKS.md` and
  `orchestration/brief_tanks.md`.

## Commands (run from the worktree with a rotopt built from `tools/rotopt`)

```
rotopt -template druid_tank -tank <1/3|1|3> \
  -tank-encounter tools/encounters/{magmadar,golemagg,onyxia,broodlord,vaelastrasz,chromaggus}.json (comma-separated) \
  -talent-search | -knobs <k=v,...> [-eval] [-race NightElf] -iters 600 -confirm-iters 3000 \
  -out <dir> -name <preset>
tankbench -encounter tools/encounters/<e>.json -iters 1000 tools/tankbench/req/druid_bear_*.json
```
