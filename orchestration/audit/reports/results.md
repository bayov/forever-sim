# Results pass 52, master 74b3a5f7c (client 1.60.1.70532)

Trigger: 37 watched files (hunter and warrior sim, APLs, presets, requests, builds.json). Used the shared 20k seed 1 run
(/tmp/audit-work/runs/master, CLI built at 74b3a5f7c); 57 requests.

- Published vs mine: 52 of 57 within 0.15. Over: druid feral-cat -1.4, feral-cat-duty-ff -1.5, rogue subtlety -1.0, subtlety-humanoid -1.1,
  mage frost +0.28. builds.json already carries my numbers for all five (866.7, 862.1, 714.9, 740.3, 946.9): the 117d18af4 full regeneration
  updated the dashboard, but the metas' published.dps (feral and subtlety: the 2026-10-24 rounds) were not refreshed. Meta lag, not a sim gap.
- Moves against the reference (previous req_summary), all explained:
  | Build | was | now | Cause |
  |---|---|---|---|
  | Beast Mastery / Lone Wolf / Marksmanship / Survival | 1171.5 / 1156.8 / 1166.2 / 1055.0 | 1141.6 / 1127.0 / 1139.4 / 1025.2 | crit per Agility at Classic's value (01afcb1d3, 0c9e30c27), MM and LW cast Multi-Shot, MM plays a Human (cdceb56c4) |
  | Arms / Arms Dragonkin / Fury / Fury 2H | 957.4 / 1000.1 / 1068.7 / 1033.7 | 906.6 / 948.4 / 986.5 / 965.6 | white-hit Rage flat with level (107874b75), rotations and talents re-searched (44fdd8599, 41665489a) |
  Battle Shout duty builds move with their bases (-54, -81, -69). Hunter -2.6% and warrior -5% to -7.7%: ruling-driven, sized by
  paired runs in the commits' own notes; I did not re-pair.
- Ranking: Demonology still leads, then Fire, Frostfire, Affliction; hunters now ~1025-1142 (BM, MM, LW still top-5 range), Fury 986.5 and
  Arms 906.6 sit mid-table. Warriors lost the most; no spec looks like a new outlier.
- Drift: identical to the reference. regen.py now maps "Night Elf" (feral meta) to NightElf; without it the feral entries failed.
- Open finding results-20261025-1 (consumes drift) unchanged.
- Unread fields: only notes.
- Reference req_summary.txt replaced with this run.
