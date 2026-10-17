# Rotations and resources: pass 28 (2026-10-17T18:50:34Z)

Master 1e2f7ec5b (last audited dc37245e5). Result: quiet, 0 new findings.

## Checked

| Change | Reaches raid presets? |
|---|---|
| Level 60 APLs (apl_logic_diff.py) | 0 APL files touched; only leveling_plan_* / level30_leveling changed |
| sim/core/spell_ranks.go: TopRankID registry | No: used by rotation-editor metadata and leveling rotations |
| sim/core/unit.go: HighestRankId in unit metadata | No: metadata only, no combat logic |
| sim/shaman/healing.go: TopHealRankID | No: only tools/rotopt/shaman_enh_leveling.go calls it (ranks[1] is right; index 0 is a placeholder) |
| ui/core/forever/racials.json | No: UI text only (Go doesn't read it); Blood Fury 3 min already in the sim |
| ui/enhancement_shaman/presets.ts | Leveling plan flag only |
| update_ledger.md | Tool-path wording only; no new client build |

Shared run (20k, seed 1) vs published: 0 of 66 requests moved more than 0.5 DPS.

## Open items

Unchanged from passes/rotations.md (Balance Night Elf, Elemental Mastery hold, warlock duty-curse screens).
