# Phase 0 re-gear of stale report requests (2026-11-27)

Commit 6bd2d60a2 took the Armaments of Battle rewards out of the Phase 0 lists and gear sets (Talisman of
Ascendance 22678, Verimonde's Last Resort 22688, Bracers of Subterfuge 22668), and 3698a73d7 moved
Enhancement from Chained Essence to Darkmoon Card: Maelstrom. The report requests in `tools/dps_report/req/`
and their race overrides and metas were never regenerated, so about 37 published builds sim gear their lists
no longer pick. Read `orchestration/audit/findings/gear.md` (gear-20261127-1) and `fairness.md`
(fairness-20261127-1, fairness-20261126-4) for the builds, the evidence and the measured sizes.

## Do

1. Regenerate every request whose gear differs from its current `level60_phase0_<spec>` gear set, with its
   race rows (`tools/dps_report/race_gear.py`) and metas. Check each request sims to its list.
   `tools/audit/gear/phase0_check.py` must print no "its equipment is not level60_phase0_<spec>" line.
2. Ret, Sub Holy and Shock Ret (fairness-20261126-4): Darkmoon Card: Maelstrom sits above Chained Essence
   of Eranikus. Re-rank the trinkets (`SLOTS=`), and if Maelstrom wins, put it in the list, gear set,
   requests and race rows.
3. Correct the Talisman comment at `sim/common/item_effects.go:4758` (and `talisman_of_ascendance_test.go:16`): the 6 casts come from the
   tooltip, the client tables hold no charges (mechanics-20261127-1). The behavior stays.
4. Shield Slam stays at 2x block value (user, 2026-11-27). Note it in `standards.md` or the rulings where the
   open question sits, so it stops being flagged.

## Publish

`stale.mjs`, `pipeline.sh hydrate/run/publish` for the moved rows, `build_data.py`, `changes.py`, one Changes
tab entry with the per-build DPS before and after. Run the gear, fairness and results auditors on the changed
builds and resolve the findings this round fixes. Merge to master (fetch and merge first, never force-push),
`make dps-report-deploy`, check the live page.

## Hand back

One message to the parent, under 250 words: master commit, deploy check, per build what moved (DPS before
and after, rank changes within a spec), what's left open. Then stop.
