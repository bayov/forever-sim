# Results audit, pass 50 (master 54878fa3d, client build 1.60.1.70532)

Due: 402 watched files changed since daba9d1a5 (alignment round, class rounds, report v36).

## Checked
- Shared run (all 64 requests, 20k, seed 1): 0 of 64 differ from published by more than 0.15 (max 0.0). Requests went 66 to 64: Ret `level_60_undead` and `level_60-duty-crusader` removed (state.json, report v36).
- builds.json vs the new run: 64 of 64 match. This run is now the reference (`ref/results/req_summary.txt`).
- Unread fields: only `notes`.
- Client build now 70532 in gate; the tables were not re-diffed this pass (no sim port cites a new build).
- Drift: regen.py failed on "Night Elf" (race with a space); fixed in the helper. New drift vs the old reference, measured in finding results-20261025-1: consumes (potion, flask, filler explosive) that mksettings writes and the requests do not. Old standing items gone: warlock sanctityAura/curseOfShadow leftovers, Ret undead and crusader seal differences, rogue/feral/priest duty items partly changed. Reference drift_report.txt replaced.

## Ranking (top and bottom, new base)
Hunters now lead (BM 1172, MM 1168, Lone Wolf 1159), then Mage Fire 1101 and Frostfire 1096; rogues are lowest (Subtlety 716, Mutilate 719, Seal Fate 729-732). The whole table rose against v35 from the regear and rotation round; the ranking order changed (hunters above warlocks). Not re-explained per spec this pass; no paired before/after sims beyond the vs-published comparison.

## Open
- results-20261025-1 (consumes drift).
- Per-spec ranking explanation for the v36 order is not done.
