# Findings: Fairness between specs

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### fairness-20261025-1: Marksmanship's report race (Skyborne Windshaper) loses 2.5 DPS to the Human
- severity: low (0.2%; Marksmanship stays #2 either way, 1170.2 Human vs 1167.6 Windshaper, Lone Wolf 1159)
- affects: hunter Marksmanship (base and duty-mark builds); Lone Wolf checked and stays Windshaper
- found: 2026-10-25, master 54878fa3d04f
- evidence: tools/dps_report/req/hunter__marksmanship.json with only the race changed, 20000 iterations: Human 1170.2 / 1170.1, Windshaper 1167.6 / 1167.7 (seeds 1 / 1000003), +2.5. hunter__marksmanship-duty-mark: Human 1168.8 / 1168.0, Windshaper 1166.1 / 1165.6, +2.6. Lone Wolf: Windshaper 1159.3 / 1158.8 vs Human 1157.8 / 1158.3 (Windshaper +1.0, keep). The fairness sweep at 5000 iterations agrees (Human 1169.4, Windshaper 1168.0). The race table in forever-wiki/sim-status.md:318 (2026-10-23) had Windshaper ahead by 2.4 at 180 sec (1168.9 vs 1166.5), so a later hunter commit in the round (the mana-gated weave rotations or the finals) moved Human up about 3.7 and the race was not re-checked. Reproduce: the race sweep (`tools/audit/fairness/sweep.py hunter__marksmanship`) or edit the race in a scratch copy of the request.
- proposed fix: re-run the hunter race sweep at 120/180/300 sec on the current Marksmanship presets; if Human still leads, switch the Marksmanship presets (and their duty builds) to Human, regenerate the requests and re-publish (+2.5 DPS).
- needs: orchestrator fix

## Resolved
