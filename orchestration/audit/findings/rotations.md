# Findings: Rotations and resources

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### rotations-20261028-1: BM, MM and LW caveats still say Trueshot Aura 100 and Hawk rank 5 (90), on the published report
- severity: low (wording only; the numbers are right)
- affects: hunter__beast-mastery, hunter__marksmanship, hunter__lone-wolf (report notes)
- found: 2026-10-28, master 1274e15b3
- evidence: 73cedc9a9 moved the sim to the 2026-10-28 hotfix: Trueshot Aura rank 5 = 50 (`sim/core/buffs.go:1748`), and Hawk rank 6 (14322) at 55 RAP, ranks 1-5 lowered to 15/22/30/38/47 (`sim/hunter/aspects.go:48-52`); the APLs and reqs cast 14322. But `tools/dps_report/req/hunter__{beast-mastery,marksmanship,lone-wolf}.meta.json:65-66`, and the deployed `ui/dps_report/data/notes.json`, still say "Assumed: Trueshot Aura's top rank gives 100 ranged attack power (the client has 50)" and "Aspect of the Hawk stays at rank 5 (90 RAP). Forever's rank 6 gives only 55." Sure: `git grep 'stays at rank 5' origin/master`.
- update 2026-10-28, master dd663e652: still open. The phrasing pass reworded both lines but kept the numbers: line 65 "Assumed: Trueshot Aura's top rank gives 100 ranged attack power (the client says 50)...", line 66 "Uses Aspect of the Hawk rank 5 (90 ranged attack power). Forever's rank 6 gives only 55." Same in notes.json.
- proposed fix: line 65 becomes "Trueshot Aura's top rank gives 50 ranged attack power, the launch realms' 2026-10-28 hotfix (ranks 1-4 lowered to 15/22/30/40). Hunter's Mark gives 143, the server hotfix (the client has 71)."; line 66 becomes "Aspect of the Hawk is rank 6 (55 RAP): the 2026-10-28 hotfix lowers ranks 1-5 to 15/22/30/38/47." Then regenerate notes.json and run `make dps-report-deploy`.
- needs: orchestrator fix

## Resolved
