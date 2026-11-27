# Findings: Synthetic gear

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### gear-20261127-1: 29 DPS requests still wear pre-6bd2d60a2 gear, with Armaments of Battle items now marked out
- severity: medium (the report sims gear the lists no longer pick; size not measured, items are trinkets, wrists, rings and heads)
- affects: Balance (4 builds), Feral cat (2), Elemental (2), Enhancement (4), Smite, Affliction/Demonology/Destruction warlock (12), Fury 2H (3)
- found: 2026-11-27, master 34994816e7
- still open at master 5ab589d858 (pass 85): same 29 requests flagged.
- evidence: `python3 tools/audit/gear/phase0_check.py` prints 29 "its equipment is not level60_phase0_<spec>" lines. Commit 6bd2d60a2 (Nov 26) re-ran the lists without the Armaments of Battle rewards (Talisman of Ascendance 22678, Verimonde's Last Resort 22688, Bracers of Subterfuge 22668) and rewrote the gear sets (`grep -c 22678 ui/warlock/gear_sets/level60_phase0_affliction*.json` is 0), but tools/dps_report/req/*.json were not regenerated; the race overrides (Gnome, Human, Tauren, Dwarf, Skyborne) flag the same items as out. Shared run builds sim the old gear. List dps vs shared run is ~1% lower for every spec incl. unchanged ones (fight difference, not flagged).
- proposed fix: regenerate the report requests and race overrides from the new gear sets (regenerate-dps-report).
- needs: orchestrator fix


## Resolved
