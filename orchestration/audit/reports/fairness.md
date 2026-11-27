# Fairness pass 68 (2026-11-27, master f4c45c3f9)

Range 313069ce..f4c45c3f9: the rogue filler-points ruling (standards 43 list, 447e6e8f1), rogue presets and report
requests moved off Deflection / Riposte / Setup into utility talents (04631ff49) and republished (f4c45c3f9); the
re-geared metas' published and duty verify refresh (87364fe2a, metas only). No sim/, proto or database change.
Scoped pass (rogue race sweep plus the shared report run), not a full re-sweep; the 2af970d74 sweep stays the
baseline. Quiet: no new finding.

## Published against current (shared run, /tmp/audit-work/runs/vs_published.txt)

28 of 68 builds moved more than 0.5 against meta.published; the largest is Lone Wolf -1.5, every other within 1.3.
All inside the 5k noise (+/-1.5). No class is stale against another.

## Rogue talents (the only DPS-side change)

- Every rogue request's talentsString equals its preset in ui/rogue/presets.ts (5 changed, Mutilate unchanged and
  already free of Deflection / Riposte / Setup). phase0_specs.json rogue rows point at the requests, so the lists
  follow the new talents.
- The ruling only moves points that add no DPS, so it cannot change the cross-spec ranking. It is rogue-only, but
  other classes' DPS-neutral points are not an asymmetry (they score nothing either way).

Race sweep, 5k seed 1, f4c45c3f9 CLI, against the 2af970d74 baseline:

| Build | Report race | Best | DPS now | Baseline | Published |
|---|---|---|---|---|---|
| Combat Swords | Undead | Undead | 840.8 | 841.3 | 840.4 |
| Combat Daggers | Undead | Undead | 829.4 | 829.8 | 829.2 |
| Seal Fate Swords | Undead | Undead | 849.3 | 849.3 | 848.4 |
| Seal Fate Daggers | Undead | Undead | 819.8 | 820.3 | 820.0 |
| Mutilate | Undead | Undead | 786.6 | 787.1 | 787.1 |
| Subtlety | NightElf | NightElf | 786.7 | 787.5 | 787.3 |

Every rogue moves -0.5 to -0.8 uniformly across races (Seal Fate Swords 0.0), within noise and in the only class
whose inputs changed; the report race is still the best race for all six. Log: /tmp/audit-work/fairness/rogue_sweep.log.

## Other checks

- phase0_check.py: no DPS-request flag. "Open above pick" lines for DPS lists are unchanged from pass 67
  (Destruction main hand 1.75, Enhancement main hand 1.77, Ret Maelstrom 1.85, the last settled by pass 67).
- Standards diff is the rogue ruling alone; nothing cross-class to re-check.

## Open items

- None new. Open findings unchanged: fairness-20261126-3 (Shahram stun, off the report) and -20261126-2 (Arcane
  Undead +1.9) plus the older race findings.
- fairness_report.md section 1 not rewritten: no full re-sweep this pass, and the rogue moves are below its
  table's resolution. Next full sweep should take a new baseline.
