# Healer rotations pass 21 (quiet)

- Master: e07e8fdb8 (`last.master` was e671e881e). Audited 2026-11-20.
- Bench: `tools/healbench` built in the audit checkout (protos generated with `make sim/core/proto/api.pb.go`;
  setup.sh still doesn't do it), default seed, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh`.

## What changed since e671e881e

- c04e51c13 / 617a63635 / 449bff490 (Forever trinket uses, item conditions registry, Forever BiS items):
  - `sim/core/spell_result.go`: `critChanceAgainst(target)` added to physical and spell crit chance, a per-target
    bonus from a mark of the caster's on that target. Zero unless something calls `AddCritChanceAgainst`, which
    only the new DPS trinket effects do on enemy targets; heals on friendly units get 0.
  - `sim/paladin/items.go`: Tenets of the Silver Hand registers an Undead mob-type condition (metadata only).
  - `sim/shaman/items.go`: Totem of Rage on Forever is +2% shock damage instead of +30 bonus damage. Not in any
    healer request.
- No healer APL, request, encounter file, `sim/druid`/`sim/priest` or healbench main.go change.

## Baseline (1,000 it) vs ref

All 63 rows (9 requests x 7 boss files, each in its role) byte-identical to `ref/healer_bench/summary.tsv`, which
confirms none of the changes reaches the healers. Ref kept as is.

Rows behind the open findings:

| request | boss | carry | mana left | OOM | other died |
|---|---|---|---|---|---|
| druid_restoration_raid | onyxia | 0.946 | 6,303 | (1,403 s) | 5.8% |
| priest_discipline (tank) | chromaggus | 0.871 | 341 | 101 s | - |
| shaman_restoration_raid | magmadar | 1.778 | 7,638 | (2,322 s) | 0.0% |

## Rotation lines

No line changed and the baseline is identical, so the line-by-line findings of earlier passes stand; no new
one-line variants were tried this pass.

## Findings

No new findings. Open, unchanged: -20261118-1 (baseline moved by 7199c8f9c, paired confirmation still pending),
-20261115-1 (Wild Growth gate x0.75), -20261104-2 (Discipline tank OOM on Chromaggus).
