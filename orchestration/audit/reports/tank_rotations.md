# tank_rotations pass 36

master 4a6dad5c2f, 2026-11-27. Previous pass on f4c45c3f9f. Result: quiet (no new findings).

## What changed

Only `dc7eb8c8f` touches watched files, and only the Leveling mode:

- `ui/tank_warrior/apls/leveling_plan_prot.json`: the leveling plan's talent pick order (5/15/31 without Defiance).
- `ui/tank_warrior/presets.ts`: the comment above `DefaultLevelingPlan` and the `Leveling: Prot` build's tooltip.

No tankbench request embeds the leveling plan (`grep leveling_plan_prot tools/tankbench/req tools/rotopt` finds
nothing), and no scoring APL, rotopt tank template, `sim/**` file, encounter file, `tools/tankbench` or
`tools/tankscore` changed since the last pass.

## Baseline

Not rerun: with no input to any tank request changed and the default seed, every row would match pass 35 exactly.
`ref/tank_bench/summary.tsv` (pass 34/35) stands.

## Rotation lines

No scoring rotation changed, so no new one-line A/B was due. Leveling rotations are out of this pass's scope
(they belong to the leveling-opt audit).

## Open

- Findings 20261104-11 and 20261104-3 still wait on user rulings.
- The Windfury main-hand ability timing item (pass doc, Open items) is still not done.
