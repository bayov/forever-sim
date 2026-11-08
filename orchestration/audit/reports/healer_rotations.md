# Healer rotations pass 5 (quiet: no new findings; -2 and -3 still open)

- Master: 9ba00c2dc (`last.master` was 73d5bfa9e). Audited 2026-11-08.
- Bench: `tools/healbench` built from this checkout (needs `make proto` first in a fresh session). Healbench's
  default seed (101), `-carry`, 1,000 iterations, each request on each of the 7 boss files at the file's own length
  in its role. Script: `tools/audit/healer_rotations/baseline.sh [out.tsv] [iters]` (reads `$AUDIT_WORK/healbench`,
  2 benches at once, about 25 minutes on 4 cores). It was missing a third time (pass 4's copy was never committed);
  rewritten. The run's commit must include `tools/audit/healer_rotations/`.

## What changed

The gate fired on 11 `sim/shaman/*.go` files. For the healers:

- `sim/shaman/*` totems (2a60326a1 .. 9ba00c2dc): Windfury and Flametongue Totem now replace each other in the
  weapon's totem slot. Neither the shaman tank nor the raid request places either totem: no effect.
- `sim/core/downrank.go` (0193721bc): the downranking penalty is heals only (user ruling 2026-11-08). Heals keep the
  same factor; only healer damage spells change, and no healer request casts one. No effect on healing.
- `sim/core/sim.go` (2a60326a1): `Roll` rounds to whole points, for heals and boss hits. Averages stay the same
  for whole-number ranges; only the spread of each hit changes. `cast.go`'s pushback rolls now call
  `RandomFloat("Damage Roll")` directly (same stream as before).
- No healer APL, healbench request, `tools/encounters/*.json` or `tools/healbench/main.go` changed.

## The baseline

All 63 rows (9 requests x 7 bosses) re-run: 24 byte-identical to the old ref, the rest moved by at most 0.002
carry, 1 EHPS and 5 mana (mean carry change -0.0001). That is the whole-point roll's noise, well under 0.02, so
not a finding. `ref/healer_bench/summary.tsv` refreshed with the new rows.

Rows behind the open findings: Discipline tank, Chromaggus 0.866 carry, 364 mana left, OOM at 101 s (unchanged);
Holy priest tank, Chromaggus 0.955, 676 mana left.

## Rotation lines

No line changed since pass 2, so its line review holds. The mana side at 180 / 300 s was not re-run: no APL,
heal or mana change, and the boss-length rows moved by noise only. Open: -2 (Discipline tank healer OOM on
Chromaggus) and -3 (shaman tank healer's Lesser Healing Wave 5 dead line), both unchanged.

## Open items for the next pass

- Emergency Flash Heal at 35% for Disc (still not run). The Holy priest's Chromaggus mana (676 left).
- `-labeled` for tighter paired A/B comparisons, once checked against the default seed path.
