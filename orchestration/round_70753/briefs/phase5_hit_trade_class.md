# Round 70753 phase 5, class worker: re-search the talents with hit-trade gear steps

You are one class worker of the phase 5 session (brief `briefs/phase5_hit_trade.md`, read its "The problem"). Branch:
start from `phase5-hit-trade`, work on `phase5-hit-<class>`, and push only there. Do NOT touch master, db.bin, the gear
sets, the synthetic items, tools/rotopt, req/, presets or the report: the phase 5 session regears and publishes.

## What exists

- Every spec with a hit talent or racial has gear steps next to its Level 60 set: `<set>_hit-N.gear.json` carries N%
  (10N rating) less gear hit, re-spent on the spec's best other stat at budget prices; `<set>_hit+N` carries N% more.
  The step pieces are in db.bin already (`orchestration/tools/hit_steps.py` and `hit_steps.json` say which stat each
  spec trades into and why; `orchestration/round_70753/hit_sweep.md` has every build's hit today).
- `tools/rotopt` flags (tools/rotopt/hitsteps.go): `-hit-steps <set path without .gear.json> -hit-cap <kind:cap>`
  make every run wear the step that brings the build's total hit (talents, race, buffs, set bonus) back to the cap.
  `-hit-span N` makes the talent search also try N steps either side and keep the best (dual wielders, hybrids).
  `-hit-school field:pct` adds school-only hit talents. `-hit-step ±N` forces a step. `-print-stats` and the talent
  search's last lines say which step a build wears.

Per gear set (gear sets under `ui/<dir>/gear_sets/`):

| gear set | flags |
|---|---|
| enhancement_shaman level60_phase1_synthetic, _burning_totem | -hit-cap melee:8 -hit-span 1 |
| elemental_shaman level60_phase1_synthetic | -hit-cap spell:16 |
| retribution_paladin level60_phase1_synthetic, _undead, _crusader_duty | -hit-cap melee:8 -hit-span 1 |
| warrior level60_phase1_synthetic_arms | -hit-cap melee:8 |
| warrior level60_phase1_synthetic_fury (dual wield) | -hit-cap melee:8 -hit-span 2 |
| rogue every level60 set | -hit-cap melee:8 -hit-span 2 |
| hunter level60_phase1_synthetic_ranged, _survival | -hit-cap ranged:8 (survival: melee:8, the same stat) |
| balance_druid level60_phase1_synthetic, _wrath | -hit-cap spell:16 |
| feral_druid level60_phase1_synthetic | -hit-cap melee:8 |
| mage level60_phase1_synthetic, _arcane | -hit-cap spell:16 |
| warlock level60_phase1_synthetic, _demonology | -hit-cap spell:16 |
| shadow_priest level60_phase1_synthetic | -hit-cap spell:16 -hit-school shadowFocus:1 |

(Smite has no hit gear to trade; leave priest__smite.)

## What to do

1. For each Level 60 build of your class in `ui/dps_report/data/builds.json` (its request is
   `tools/dps_report/req/<id>.json`, its meta `<id>.meta.json`), re-run the talent search the way your class's round 70753
   worker did (`orchestration/round_70753/<class>.md`, `tools/rotopt/<class>_presets.sh`: template, knobs, pins with
   `-require`, the mean of 120, 180 and 300 s via `-combined -durs 120,180,300`), from the build's current talents,
   with the hit flags above. Keep the pins each build is named for. Duty and encounter variants (meta `baseBuild`)
   keep their base build's talents unless they were searched on their own before.
2. Confirm each build at 20,000 iterations, 120, 180 and 300 s each (`-eval -confirm-iters 20000 -durs <d>`), both the
   old (current talents, current gear: no -hit-steps) and the new (picked talents with `-hit-steps ... -hit-step <its step>`).
3. Write `orchestration/round_70753/hit_trade/<class>.md`: per build: old and new talent string, the talent changes in
   words, the gear step (and gear set file), total hit before and after against the cap, old and new DPS at
   120/180/300 with errors, and the exact commands. Note anything odd (a step the tool can't reach, a search that
   wanted more steps than exist). Commit it on `phase5-hit-<class>` (end the message with the trailer lines your
   session prompt specifies; no model names in files) and push there. Then stop.

Machine: long runs in the background, outputs to files, read only what you need (CLAUDE.md "Keep the context small").
Searches on several builds can run side by side if the machine has the cores (`nproc`).
