# Phase 0 gear for the tank and healer reports (2026-11-25)

The user wants the tank and healer reports on real Phase 0 gear, like the DPS report (master 86fbcf456 and
later): per spec Phase 0 BiS lists, gear sets, re-sims, the full tank and healer audit, deploy. Today they
wear the synthetic sets (standards 35). No toggle.

Read first: `briefs/phase0_step3_1122.md` (how the Phase 0 search runs), `briefs/phase0_step4_1124.md`
(how a spec moves onto its set), `tools/rotopt/bis_phase0.sh`, `tools/rotopt/phase0_specs.json`,
`tools/rotopt/phase0_prices.py`, `orchestration/brief_tanks.md`, the tankbench and healbench READMEs and
requests (`tools/tankbench/req/`, `tools/healbench/req/`), and `orchestration/audit/standards.md` (clause 43
is Phase 0 gear).

## In parallel with the price worker

Session phase0_prices (brief `briefs/phase0_prices_1125.md`) is changing `phase0_prices.py` (price every
list, roll share 1.0 for dungeon drops), `phase0_content.json`, and some DPS lists, and will merge and
deploy later today. Don't edit `phase0_prices.py` or the DPS files. Build your scoring changes in
`bis_phase0.sh` as small, separate additions, merge master often, and run your final searches and
pricing only after that worker's merge lands on master (look for its Changes entry or "phase0-prices"
merge). Until then, design, build, and test on small runs.

## 1. Scoring

The Phase 0 search scores a set by DPS. Add tank and healer scoring that reads the same score the reports
publish: the tank score (mean over the six scored bosses, per preset: Balanced, Max threat, Max survival)
via tankbench, and the healer score via healbench. Keep the search's other rules: one content check, the
gold line, Exalted rewards ranked but not picked, unsure items in (standards 42), faction by race.

**Gold line for score units:** 30g per DPS on a ~1000 DPS build is about 3000g per 1% of score. Use that
per percent of the spec's score: an item over 3000g for each 1% it adds is infeasible. Make it a setting
next to MAX_GOLD_PER_DPS and note it in standards 43.

If a preset's gear would differ (Max threat against Max survival), decide from how the synthetic sets are
split today: one list per preset if they are split, else one list on Balanced. Say which in the notes.

## 2. Lists, sets, requests

For every tank and healer spec in the reports: a Phase 0 list (full search, then the set-trap check from
step 4: a set bonus that needs several pieces at once), `gearOut`, the gear set, requests on it, race rows
with the other faction's swaps, consumables and enchants from the Phase 0 tables. Re-check rotations and
talents that depend on hit, defense or mana at the new gear. Delegate per spec to subagents (at most two
levels below the main session).

The raid healer bench fills the raid with the DPS report's Frost mage request. User ruling 2026-11-25:
keep that, so the filler wears Phase 0 gear (healer_fairness-20261124-1).

## 3. Publish, audit, deploy

`stale.mjs`, the tank and healer pipelines (hydrate, run, publish), their tables and notes, wording that
still says synthetic for tanks and healers (layout, report.js, build_data.py, README, standards), a Changes
tab entry. Then the full tank and healer auditors in `orchestration/audit/auditors.json` (tank_*, healer_*),
plus the fairness pass checking that every list item that could be picked has a price, fix round, merge to
master (fetch and merge first, never force-push), push your branch too, `make dps-report-deploy`, check the
live page.

## Hand back

Send the parent one message under 300 words: master commit, deploy check, per spec the Phase 0 score
against synthetic, the scoring and gold-line choices, audit findings and fixes, anything open. Then stop.
