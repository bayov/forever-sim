# Dashboard: next round (queued by the user 2026-10-13, after the redesign ships)

## Assumption scenarios per spec
User request (paraphrased): add a section about each assumption we made for each spec. Examples: warlock pet scaling, Maelstrom Weapon PPM. For each assumption, offer several reasonable settings for what the real effect will be on release (e.g. Maelstrom Weapon low / moderate / high PPM, with predicted values suggested by a subagent). The viewer can then select configurations and see their effect. In the Ranking tab these are top-level settings, grouped by class, above the DPS ranking widget. Remember each selection per viewer.

Plan:
1. Inventory. For each class, list the spec assumptions the sim makes where Forever's real value is unknown or estimated. Sources: each PR's caveats and sensitivity lines, forever-wiki/sim-status.md, notes.md, the class audits, and the code comments marked as guesses or placeholders.
   Examples:
   - warlock pet scaling (SoD coefficients as placeholder);
   - Maelstrom Weapon PPM;
   - Windfury Totem ICD (1.5 s vs 100 ms);
   - Bane of Doom coefficient;
   - Rage of the Farseer cooldown;
   - Trueshot Aura / Hunter's Mark ranks;
   - hunter pet scaling;
   - Lava Burst's Flame Shock bonus.
   Each entry gets the current value, its source, why it is uncertain, and 3 settings (low / expected / high) with a one-line rationale. Class sessions or a research subagent propose the values; the orchestrator reviews them.
2. Sim support. Each assumption becomes a knob the dashboard pipeline can set per run, without changing the shipped presets. Prefer a dashboard-only patch in the local merged tree (env var or debug option read where the constant lives), so the class PRs stay clean. If a knob is useful in the real sim UI, the class session adds it in its PR instead.
3. Runs. For every build that an assumption touches, sim each setting (180 s plus the length set, 20k iterations, preset race; race sweep only if a setting could change the best race).
   Combinations: precompute the full grid when a class has 3 or fewer assumptions. Otherwise store per-assumption deltas and combine them multiplicatively, with a note that combined values are approximate. Spot-check a few combinations against real sims.
4. Data. A new `assumptions` dataset (class, spec(s), key, label, description, source, settings [{id, label, value, rationale}], default) and a `scenario_dps` dataset (build, assumption, setting, dps, se, plus lengths).
5. Page. Ranking tab, above the ranking: a compact "Assumptions" panel grouped by class (collapsible per class). Each assumption is a segmented control (Low / Expected / High) with an info tooltip. The ranking, the spec detail headline and Compare recompute from the selections, a visible "custom assumptions" badge and reset appear when anything differs from the defaults, and the Method tab lists every assumption with its sources.
   Persistence: the user asked for a cookie. The artifact runs in its own sandboxed origin, so use localStorage (try/catch; cookies may be blocked in the sandbox), falling back to the URL hash.
