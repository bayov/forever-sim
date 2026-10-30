# Results audit, pass 64 (master fb1485888, client build 1.60.1.70707)

Why due: 62 watched files changed since bbbaf2e3c (extra attacks keep the swing timer, Windfury Totem 100 ms cooldown,
Eviscerate 4% AP, stones only on fitting weapons, mage 70707 talents and Arcane Blast stacks, leveling and raid-sim work).

## Checked
- Shared run (20k, seed 1, 77 requests): builds.json matches the run on all 77 within 0.15 (the report v-re-sim fb1485888 is consistent).
- Metas' published.dps: 46 of 77 differ from the run (finding results-20261030-1, low, same pattern as 20261027-1).
- Moves against the previous req_summary.txt (45 over 0.5), by cause, from commits f3017897f and the mage commits:
  - Warrior Arms +9.6..10.5%, Fury / Fury 2H +6.3..7.1%: extra attacks (Windfury Totem, Hack and Slash, Sword Spec) leave
    the swing timer alone, WF cooldown 100 ms (assumed until a log settles it, per notes.md).
  - Hunter Survival +4.3% (melee, Raptor Strike under WF); BM and MM unchanged.
  - Paladin +1.6..3.3% (Hand of Justice / extra attacks, WF); Feral +2.6..3.1%.
  - Rogues +1.8..3.0%: Eviscerate 4% AP per combo point, plus WF timer rule; Combat presets lost the off-hand stone mismatch.
  - Mage: Frostfire -3.4%, Fire -2.6%, Arcane +1.5% (70707 talent re-pick, 4 Arcane Blast stacks at any mana), Frost -0.3%.
  Sizes not re-verified with a paired run on the old binary (no old CLI in the shared cache); directions fit the commits.
- Drift report (regen.py): identical to the reference (155 lines). Standing consumes drift (results-20261025-1) unchanged.
- Unread fields: none.

## Housekeeping
- ref/results/req_summary.txt replaced with this run (77 lines); drift_report.txt unchanged. regen.py: hunter ids with -tallstrider/-wolf now map (it crashed with KeyError before).

## Conclusion
One new low finding (stale metas). No modelling errors seen.
