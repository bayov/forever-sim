# Results pass 51, master 016daa654 (client 1.60.1.70532)

Trigger: 43 watched files (hunter Mark duty presets removed, Scorpid duties removed, hunter builds regenerated, db, builds.json).
Used the shared 20k seed 1 run (/tmp/audit-work/runs/master); 57 requests.

- Published vs mine: 57 of 57 within 0.15.
- builds.json vs summary: 4 hunter rows differ by 0.0 to 0.5 (builds rounding of re-sim; BM 1171.5, LW 1156.8, MM 1166.1, SV 1055.0). 7 duty builds removed.
- Moves against the reference (explained by e71b90f51, 5429f4185, c20bbc3df: every hunter keeps up its own Hunter's Mark):
  BM 1172.0 -> 1171.5, Lone Wolf 1159.3 -> 1156.8, MM 1167.6 -> 1166.2, SV 1056.6 -> 1055.0. Removed: 7 hunter duty-mark/duty-scorpid requests (214bc117b).
- Drift: identical to the reference except the removed hunter duty entries. regen.py fixed for races with a space ("Night Elf" -> NightElf), which broke the feral entries.
- Open finding results-20261025-1 (consumes drift) still reproduces: Ret Shock potion, Elemental mana-tide potion, rogue flask/filler.
- Unread fields: only notes.
- Ranking order of the top is unchanged: Demonology and the hunters lead; no new outliers.
