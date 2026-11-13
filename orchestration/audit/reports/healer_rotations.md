# Healer rotations pass 11 (quiet: no new findings; -20261104-2 still open)

- Master: 97da1cca5 (`last.master` was 71562a8b0). Audited 2026-11-13.
- Bench: `tools/healbench` built from this checkout, default seed, `-carry`, 1,000 iterations, each request on each
  of the 7 boss files at its own length in its role (`tools/audit/healer_rotations/baseline.sh`).

## What changed since 71562a8b0

- Gate: one watched file, `sim/paladin/unsealed_wrath.go` (33b13e310): Windfury Totem extra swings no longer store
  Unsealed Wrath echoes (Nov 11 hotfix). It touches only the paladin's own landed white swings with the Retribution
  talent; neither Holy paladin request (talents 53320015002101251-5-052330002) melees in the bench.
- Also on master, not in the gate: `sim/core/flags.go` plus a Windfury Totem extra-attack test. Melee only.
- No healer APL, healbench request, encounter file or `tools/healbench/main.go` changed.

## Baseline

All 63 rows (9 requests x 7 boss files) are identical to `ref/healer_bench/summary.tsv` in every column (carry,
EHPS, overheal, EH/mana, mana left, OOM time, deaths). Max carry delta 0. The reference is unchanged, so it was not
rewritten.

## Rotation lines

No rotation changed and no input to them changed, so the pass 10 analysis stands (see git history of this file
for its tables and A/B screens). No new one-line changes were tried.

## Findings

- New: none.
- Open: healer_rotations-20261104-2 (Discipline tank healer OOM on Chromaggus at its own length and Golemagg at
  300 s): unchanged, as nothing in the Disc request or mana engine moved.
