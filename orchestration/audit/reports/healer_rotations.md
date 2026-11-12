# Healer rotations pass 9 (quiet: no new findings; -20261111-1, -20261104-2, -20261104-3 still open)

- Master: 8306f1d52 (`last.master` was a0b4d3be9). Audited 2026-11-12.
- Bench: `tools/healbench` built from this checkout, healbench's default seed, `-carry`, 1,000 iterations, each
  request on each of the 7 boss files at the file's own length in its role. Script:
  `tools/audit/healer_rotations/baseline.sh <healbench> <out.tsv> [iters] [extra flags]` (run from the repo root).
  It was missing from master again. This copy is restored from commit 8f3f74f3f, which is only on
  `origin/tankheal`. The run's commit must include it.
- Note: `origin/tankheal` holds a by-hand "healer_rotations pass 9" (8f3f74f3f, 2 new findings, a rewritten
  summary.tsv) that isn't merged into master. This pass is numbered from master's state (pass 8 -> 9). Whoever
  merges tankheal must reconcile the two.

## What changed

The gate fired on `sim/druid/rake.go`: Forever Rake's direct hit now takes 0.065 x attack power (the user's ruling,
2026-11-12). Rake is a cat-form ability that no healer request or APL casts. Other files touched since a0b4d3be9:
`sim/core/pet.go` and `sim/core/unit.go` (pets inherit their owner's melee speed, used by the hunter pet only), the
hunter files and test results. No healer APL, healbench request, `tools/encounters/*.json` or
`tools/healbench/main.go` changed.

## The baseline

All 63 rows (9 requests x 7 boss files) are byte-identical to `ref/healer_bench/summary.tsv` (`cmp` identical), so
the ref stays as it is.

## Rotation lines

No rotation changed, so there were no new one-line A/B runs. The baseline re-checked the open findings, and they
are unchanged:

- -20261111-1 (Resto druid raid: Wild Growth uses up Nature's Swiftness): still open. Its rows are unchanged.
- -20261104-2 (Disc tank healer runs out of mana on Chromaggus): unchanged.
- -20261104-3 (Resto shaman tank healer's LHW 5 line never fires): unchanged.

## Open items for the next pass

- Emergency Flash Heal at 35% for Disc (still not run). `-labeled` for tighter paired comparisons.
