# Mechanics pass 63 (2026-11-04, master f331f0cff)

Quiet: no new findings. mechanics-20261103-2 (tank threat) was already moved to Resolved (0dcb0d87a).

## Range

f23ea25d2..f331f0cff. Three non-test code commits under the watched paths:

| Commit | What | Mechanics check |
|---|---|---|
| 90e0bb01a | mksettings drops Blessing of Salvation; every Level 60 request loses `raid.buffs.salvation` | Post 9 (tier 1): "Salvation, Mass Salvation, and Revoke have been removed." Salvation was threat only (aura 10, −30%), so no DPS moves; commit measured fury / fire identical. Correct. verdicts row updated. |
| 0dcb0d87a | Bear Form x2.15, Defensive Stance x1.8 times Defiance 1.07/1.13/1.20, under Forever | Matches post 9 wording exactly ("from 50% to 115%", "from 30% to 80%", "7/13/20% from 5/10/15%"); client 70753/70755 still bp 50 / 30, so notes-only per standards 37. Classic keeps old values. Threat only. Resolves mechanics-20261103-2. |
| 13c34f20f | Gates round 70753's notes-only / hotfix-only values (rogue, hunter, paladin, shaman, warlock pet health) under Forever | Read the code hunks: every Forever branch keeps its prior value; only the non-Forever path changes (report_hashes.txt unchanged). No shared mechanic touched. |

`sim/testdata/report_hashes.txt` is test data. No sim/core, racial, raid buff/debuff or consume code moved.

## Getting current

| Source | Result |
|---|---|
| Client build | 1.60.1.70755 still newest (gate). |
| Hotfixes | `hotfixes.py pushes --set mechanics 70755`: no rows. |
| Dev notes post 9 | Edited 2026-11-03T20:47:06Z. Diff against the saved copy: only "Gamepad Support has entered a functional Beta state" (was "a complete functional state"). No mechanic. |
| Blue tracker | Since 11-03 19:05: "Explore WoW: Forever with a Gamepad" article (no values), Linxy "Bunmover needs work!" (category 17, not Forever mechanics). |
| Daily update | No TASKS.md / update_ledger.md changes in range besides the audit-loop commit. |
