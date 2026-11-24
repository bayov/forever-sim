# Phase 0 step 4, publish (2026-11-24)

All seven class workers are done: the level 60 DPS builds now have Phase 0 gear sets, requests and race
rows on master (6fdbcc951 or later). Their notes are `orchestration/round_notes_2026-11-11/phase0_step4_<group>.md`
(mage, warlock, druid_priest, rogue, shaman_ret, hunter, warrior). Tanks and healers stay on synthetic.
You finish the step: fixes, publish, full audit, fix round, deploy. Follow the `dps-report` skill and
`orchestration/audit/standards.md` for how a round runs.

## 1. Fixes before publish

1. **race_gear.py faction sets skip the hit trade.** When a race row uses the other faction's set
   (`<gearSet>_horde` / `_alliance`) the script returns early, so Dwarf rows never trade excess hit:
   Arms Dwarf (9.3% hit, about 5 dps), and Dwarf BM, MM, SV hunters (0.6% unspent, SV 1%). Apply the
   trade on top of the faction set, re-sim those rows.
2. **Synthetic item ids in APLs.** Balance and Smite prepull cast trinket 995034. Sweep every DPS APL for
   synthetic ids (99xxxx and anything from `tools/database/forever_synthetic_items*.go`) and replace them
   with the Phase 0 set's real on-use trinket, or drop the cast when the set has none.
3. **Deathmaw's Guise (282049)** is missing from `tools/rotopt/phase0_content.json` and blocked rogue
   gearOut. Add it with its right tags (one entry per line, the file's format).
4. **Wording branch `phase0-step4-wording`** (head 541cb9205): merge it. It changes layout.html,
   report.js, build_data.py, README.md, rotation_text.json (`{trinket}` keys), rotation.py, standards
   clause 43, the gear and fairness passes, and adds `tools/audit/gear/phase0_check.py`. Make
   phase0_check.py not flag items that rank but are tagged `avail: "infeasible"` (Nacreous Shell
   Necklace ranks above the neck pick at 217g per dps, that's by design). Then run it and fix what it finds.
5. **Checks:**
   - Frost mage is 23% under synthetic (726.4 vs 944.7), the most of any build. Confirm it's real
     (gear, not a broken set or rotation).
   - Mage lists read 9 to 14 dps above their requests (list scoring includes 90s use values?). Find
     out, and make sure every request sims to its list's final measure within noise.
   - Seal Fate Swords Gnome rotation gain (+4.7) was measured on synthetic gear. Re-check it.
   - Hunter MM and SV re-ranked only head and feet. Run a full RANK_ONLY for both if it fits in time.

## 2. Publish

1. `stale.mjs`: re-run every stale row. Expect tank 12/12 and heal 9/9 stale (built on 8950440 bins)
   and Troll DPS hashes. Tank and healer rows get re-simmed as is, not moved to Phase 0.
2. `pipeline.sh hydrate`, `run`, `publish` for the DPS ids (and the stale tank/heal ids).
3. `build_data.py`: fix every rotation_text `{trinket}` warning.
4. `changes.py`, and one Changes tab entry: the DPS report now uses real Phase 0 gear (gear you can get
   at level 60 in the first phase), not synthetic gear, so numbers dropped about 5 to 20%.

## 3. Audit and deploy

1. The full auditor suite (gate.py / B5 per `orchestration/audit/auditors.json`), with the hit-dependent
   checks. Fix the flaky simsignals test if it fails. Add a test that pins the tank racial lines (prot
   Blood Fury and Berserking casts) if tank_mechanics still asks for one.
2. Fix round, then re-run the auditors that found something.
3. Merge to master (fetch and merge first, never force-push), push branch too, `make dps-report-deploy`,
   and check the live page loads with the new numbers.

## Hand back

Send the parent one message: the master commit, the deploy URL check, per build the published Phase 0
dps (one line per class), audit findings and fixes, and anything left open. Under 300 words. Then stop.
