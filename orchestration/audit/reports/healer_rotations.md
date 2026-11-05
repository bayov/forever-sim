# Healer rotations pass 4 (quiet: no new findings; -2 and -3 still open)

- Master: bba052f75 (`last.master` was a07941531). Audited 2026-11-05.
- Bench: `tools/healbench` built from this checkout. Healbench's default seed (101), `-carry`, 1,000 iterations,
  each request on each of the 7 boss files at the file's own length in its role. Script:
  `tools/audit/healer_rotations/baseline.sh [out.tsv] [iters]` (reads `$AUDIT_WORK/healbench`, 2 benches at once,
  about 14 minutes). It was missing again (the pass 3 copy was never committed); rewritten this pass. The run's
  commit must include `tools/audit/healer_rotations/`, or the next pass rewrites it a third time.

## What changed

- `tools/healbench/main.go` (the watched file): new flags `-seed` (default 101, the old hard-coded seed),
  `-labeled` (default off), `-scales`/`-vectors` (per-fight values for `tools/dps_report/tankheal.py`'s race
  pairing). With the defaults the request is the same as before: `RandomSeed: 101`, `UseLabeledRands: false`;
  `SaveAllValues` only with `-vectors`. The table and carry paths are untouched.
- Engine: `sim/core/raid.go` appends per-fight death flags to the healing-fight metrics,
  `sim/core/sim_concurrent.go` merges them. Bookkeeping only; nothing reads it in the sim.
- No healer APL, healbench request or `tools/encounters/*.json` changed.

## The baseline

All 63 rows (9 requests x 7 bosses) are byte-identical to `ref/healer_bench/summary.tsv` (diff empty): carry,
EHPS, overheal, EH per mana, mana left, OOM time, deaths and death floor. So the new flags' defaults keep the
bench exactly as it was. `ref/healer_bench/summary.tsv` left as is (no change). The mana side at 180 / 300 s was
not re-run: with no APL, heal or mana change and an identical boss-length baseline, it cannot have moved.

## Rotation lines

No line changed since pass 2, so its line review holds. Open: -2 (Discipline tank healer OOM on Chromaggus at
101 s, carry 0.866 / 365 mana, unchanged) and -3 (shaman tank healer's Lesser Healing Wave 5 dead line). No new
variants tried; pass 3's (shield gate 0.9, Renew mana gate) stand.

## Open items for the next pass

- Emergency Flash Heal at 35% for Disc (not run). The Holy priest's Chromaggus mana (677 left).
- `-labeled` could tighten paired A/B comparisons (rolls stay aligned past where two rotations differ); worth
  trying for sub-0.02 carry questions instead of 4,000 iterations, once checked against the default seed path.
