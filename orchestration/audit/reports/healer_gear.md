# healer_gear pass 11 (master a6677362fc)

Trigger: orchestration/standards.md changed (1 line) since 8219f1083f.

- Diff: rule 27 only (haste now shortens the spell GCD, 1.5/(1+haste), 1s floor; ruling 2026-11-18). No change to rules 1b, 5, 5a, 35.
- No change on master since the last pass to tools/database synthetic healer items, ui/*/gear_sets, or orchestration/gear_targets.md.
- Haste is not in the healer weights (Sta/Int/Spi/crit/SP/healing/MP5/hit), so no piece's linear total, norm or set share moves.
- Per-piece and per-set tables from pass 10 stand (not recomputed: inputs and standards unchanged).
- Findings: none open; none new.
