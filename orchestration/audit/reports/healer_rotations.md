# Healer rotations pass 6 (quiet: no new findings; -2 and -3 still open)

- Master: ca5e35c3f (`last.master` was 91ad19535). Audited 2026-11-10.
- Bench: `tools/healbench` built from this checkout, healbench's default seed, `-carry`, 1,000 iterations, each
  request on each of the 7 boss files at the file's own length in its role. Script:
  `tools/audit/healer_rotations/baseline.sh [out.tsv] [iters]` (reads `$AUDIT_WORK/healbench`, 2 benches at once,
  about 25 minutes on 4 cores). It was missing a fourth time (never committed); rewritten again. The run's commit
  must include `tools/audit/healer_rotations/`.

## What changed

The gate fired on `sim/shaman/flametongue_weapon.go` and `sim/shaman/talents.go` (bcc5d6cd4):

- Maelstrom Weapon stacks on every melee hit in Forever, and Flametongue Weapon rank 5 cut. Both are Enhancement
  only: neither shaman healer request takes Maelstrom Weapon or imbues Flametongue (both use Brilliant Mana Oil).
- `tools/database/forever_enchants.go` (04bdcdb9e): enchants 41 and 66 take Forever's Stamina values. No healer
  request uses either.
- No healer APL, healbench request, heal, mana or engine code, `tools/encounters/*.json` or
  `tools/healbench/main.go` changed.

## The baseline

All 63 rows (9 requests x 7 bosses) re-run: byte-identical to `ref/healer_bench/summary.tsv` (max change 0 carry,
0 EHPS, 0 mana). Ref unchanged.

Rows behind the open findings: Discipline tank, Chromaggus 0.866 carry, 364 mana left, OOM at 101 s (unchanged).

## Rotation lines

No line changed since pass 2, so its line review holds; the 180 / 300 s mana side was not re-run (no APL, heal or
mana change, baseline identical). Open: -2 (Discipline tank healer OOM on Chromaggus) and -3 (shaman tank healer's
Lesser Healing Wave 5 dead line, still lines 8-9 on 10468 / 10467), both unchanged.

## Open items for the next pass

- Emergency Flash Heal at 35% for Disc (still not run). The Holy priest's Chromaggus mana (676 left).
- `-labeled` for tighter paired A/B comparisons, once checked against the default seed path.
