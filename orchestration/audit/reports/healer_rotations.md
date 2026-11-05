# Healer rotations pass 3 (quiet: no new findings; -2 and -3 still open)

- Master: a07941531 (`last.master` was dc303b026). Audited 2026-11-05.
- Bench: `tools/healbench` built from this checkout. Healbench's default seed, `-carry`, 1,000 iterations, each
  request on each of the 7 boss files at the file's own length in its role. Script:
  `tools/audit/healer_rotations/baseline.sh <healbench> <out.tsv> [iters]` (pass 2 described it but it was not in
  the repo; rewritten this pass; a full run takes about 30 minutes).

## What changed

- No healer APL, healbench request, `tools/encounters/*.json` or `tools/healbench/main.go` changed.
- Engine: `sim/core/spell_result.go` (a heal on a tank in a score's replay makes threat from the whole heal,
  tank_mechanics-20261104-16): threat only, not read by the healer score. `sim/druid/ferocious_bite.go`
  (hotfix 113403): cat damage only. `sim/core/consumes.go` (explosives usable in form), `major_cooldown.go`
  (BonusArmor on-use trinkets become survival cooldowns), `character.go` (comment): none touches a healer's
  heals, mana or cooldowns. TestDisc/TestHoly hps moved by under 0.3%, from the shared test raid, not the bench.

## The baseline

All 63 rows (9 requests x 7 bosses) are identical to `ref/healer_bench/summary.tsv` from pass 2: carry, EHPS,
overheal, EH per mana, mana left and OOM time match to the printed digit. The table in pass 2's report still
stands (for example Disc Chromaggus 0.866 / 365 mana / OOM 101 s; Holy priest Chromaggus 0.955 / 677).
`ref/healer_bench/summary.tsv` rewritten from this run (no diff). The mana side at 180 / 300 s was not re-run:
with no APL or heal change and an identical boss-length baseline, it cannot have moved.

## Changes tried (paired seeds, 1,000 it, tank role)

| Disc variant | Chromaggus (115 s) carry / mana / OOM | Golemagg (56 s) carry / mana |
|---|---|---|
| master | 0.866 / 365 / 101 s | 1.991 / 6234 |
| Power Word: Shield mana gate 0.9 (line 3, was 0.7) | 0.889 / 428 / 107 s | 1.984 / 6139 |
| Renew only at 40% mana or more (line 5) | 0.865 / 435 / 103 s | 1.991 / 6234 |

The shield gate 0.9 again gives about +0.02 on Chromaggus and -0.007 on Golemagg, and still runs dry; the
Renew gate does nothing. Both added under -2. Neither fixes the OOM, so -2 still needs a worker.

## Rotation lines

No line changed since pass 2, so pass 2's line review holds: -3 (shaman Lesser Healing Wave 5 dead line) is
still open as filed; -2 is still open.

## Open items for the next pass

- Emergency Flash Heal at 35% for Disc (not run). The Holy priest's Chromaggus mana (677 left).
