# Dashboard job templates

These are the actual handoffs used on 2026-10-13 to build and change the dashboard. Reuse them as templates:
- `$SCR` was the orchestrator's scratchpad, where the pipeline lived then. The pipeline is now in `tools/dps_report/` (see its README). Point new handoffs there, or copy it into a scratch directory.
- `<DASHBOARD_ARTIFACT_ID>` is the dashboard artifact. Its URL is in `../../state.json`.
- Replace `<the attribution trailer lines your session prompt specifies>` with the trailer from the running session's prompt.

| Template | Use for |
|---|---|
| HANDOFF_dash_refresh.md | rebuild every dataset on new class heads (block B7) |
| HANDOFF_assumptions.md | the assumption scenarios knobs, sims and panel (B9) |
| HANDOFF_specpage.md, HANDOFF_races.md, HANDOFF_setup_active.md, HANDOFF_forever_tooltips.md, HANDOFF_redesign.md, HANDOFF_tooltips.md | page features (B8) |
| HANDOFF_budget.md, HANDOFF_bis_realism.md | research analyses that ended up as orchestration/*.md (B10) |
