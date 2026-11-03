# Pass: tank rotations and resources

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03, as `rotations.md` is for the DPS specs.

You audit every tank preset's rotations (`ui/tank_warrior`, `ui/protection_paladin`, `ui/feral_tank_druid`,
`ui/warden_shaman` apls, and `tools/tankbench/req/*.json`, which embed them) for score they leave on the table:
threat (SafeDPS) and survival (healing needed), on `tools/tankbench` and the objective (`docs/tank_sim.md`, The
optimizer's objective).

## Rules

- Don't change rotations; file findings with the A/B numbers. Paired runs: same seed (tankbench's default), the
  same requests apart from the one line, 1,000 iterations a row, 3,000 to settle a gain under twice its paired
  error (`tankscore.PairedDiff`, or rotopt's search output).
- Judge at the preset's own weight (Max threat 1/3, Balanced 1, Max survival 3) on the scoring encounters
  (the Anniversary set without Ragnaros, `orchestration/brief_tanks.md`, scored as their mean), and look at
  Ragnaros and each boss alone for a change that helps one and hurts the rest.
- A rotation line may not read the tank's health in a scoring preset (the replay can't follow it); a health-based
  button is a finding only if its timer-based use would score.

## Setup

```sh
source tools/audit/setup.sh
go build -tags=with_db -o $AUDIT_WORK/tankbench ./tools/tankbench
T=$AUDIT_WORK/tankbench
for b in tools/encounters/*.json; do $T -iters 1000 -encounter $b tools/tankbench/req/<request>.json; done
```

## The pass

1. **What changed** since `last.master`: the tank APLs and requests, the rotopt tank templates, `sim/**` tank
   abilities, talents and the engine (attack table, threat, Rage), `tools/encounters/*.json`,
   `tools/tankbench/main.go`, `tools/tankscore`.
2. **The baseline**: every tank request on every boss file, 1,000 iterations: the score at its weight, SafeDPS,
   TPS, DTPS, healing needed, death floor. Compare it with `ref/tank_bench/summary.tsv`. A row that moved more than
   its error without a rotation change is a finding for the change that moved it (or for `tank_mechanics`). Then
   write the new `ref/tank_bench/summary.tsv`.
3. **Each rotation line**: does it do what its note says, in game terms (standards 31)? Is a cooldown or a
   resource wasted (Rage capped, Shield Block or Holy Shield down, a debuff dropped, Maul or Heroic Strike starving
   the next GCD)? Does the opener make the pull's threat (Charge with Vanguard, Bloodrage, the first GCDs)? Try the
   one-line changes that test these and report the ones that win.
4. Report: the baseline table, the changes tried with their numbers, new findings.
