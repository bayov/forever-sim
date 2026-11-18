# tank_rotations pass 19

Master a6677362f (2026-11-18), against pass 18's baseline (master 8219f1083). 1,000 iterations, default seed, every
tank request on all seven boss files. Quiet: no new finding.

## What changed

Gate: 11 watched files. Their diffs are score-neutral by reading:

- `core/rage.go`, `warrior/rend.go`, `warrior/sweeping_strikes.go`, `paladin/{consecration,holy_strike,soc}.go`,
  `druid/wrath.go`, `core/debuffs.go`: DPS Report assumption knobs (`core.Assume*`, `reportSwitches`), each
  defaulting to the current constant unless a `DASH_ASSUME` switch is set. Tankbench sets none.
- `core/health.go`: health-history lookup rewritten as a binary search (same result); `target_death.go`,
  `water_totems.go`: `sightEpoch` and `SpellMetrics()` accessor plumbing.

Outside the watched list, ee35da473 ("Haste shortens a spell's GCD under Forever", `core/cast.go` hastedGCD, and
Warchief's Blessing +15% cast speed) is an engine change backed by a Forever combat log (2026-11-18). It is what moved
the Warden rows below.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for a6677362f. Warrior, Prot paladin and bear rows are bit-identical to pass
18's. Only the 24 Warden rows moved, mixed sign and mostly inside the rows' error:

| request | scoring mean before | after | largest row moves |
|---|---|---|---|
| shaman_warden_balanced | 1282.8 | 1288.0 | Onyxia +76 (±~48), Broodlord -63 (±~55) |
| shaman_warden_max_survival | -2221.2 | -2192.8 | all rows within ±30 |
| shaman_warden_max_threat | 2645.7 | 2621.0 | Broodlord -68 (±~43), Vaelastrasz -50 (±~33), Ragnaros +197 (±~145) |

SafeDPS moved by -2 to -24 per row, healing needed by -6 to +59. Same seed but a different code path, so these are
not paired; the pattern (no consistent direction across bosses) reads as the GCD rule reshuffling the rolls plus a
small real effect. It comes from a ruled, log-backed engine change, so it is not a finding here (any question on the
rule itself belongs to tank_mechanics).

## Rotation lines

The one Warden line the shorter spell GCD could bear on is Fire Nova's mana gate (Max threat, `currentManaPercent >= 0.4`):
more spell casts per minute could drain mana sooner. Paired A/B (`abtest`, 1,000 iterations, default seed, all seven
bosses):

| change | scoring mean diff | error |
|---|---|---|
| Fire Nova gate 0.4 -> 0.3 | +1.6 | 1.4 |
| Fire Nova gate 0.4 -> 0.6 | -1.1 | 2.2 |

Neutral: every scoring boss 0.0 diff; only Ragnaros moves. No other tank APL, request, rotopt template, tank
ability, talent or encounter changed, so no further one-line A/B was warranted. Pass 17's checks still hold.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call, unchanged (warrior rows identical).
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling, unchanged (paladin rows identical).

## Notes for the orchestrator

- The pass doc's open item (main-hand timing for Windfury) was done for all four tanks in pass 15; it can still be
  removed from passes/tank_rotations.md.
- tankbench and abtest built in the shared CLI worktree (/tmp/audit-cache/cli/a6677362fc06/tree), same commit.
- Housekeeping: an earlier command in this pass ran with `$AUDIT_WORK` unset, so the raw tankbench outputs also sit in
  `/tank_rotations/` (and `/tr_diff.txt`, `/tank_rotations/old.tsv`) at the filesystem root. A safety check refused
  their removal; they are scratch and can be deleted by hand.
