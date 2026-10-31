# Results audit, pass 65 (master e05fd732e, client build 1.60.1.70707)

Why due: 172 watched files changed since fb1485888/a2a3e66 (audit-loop fixes e1fcd57b7, re-sims 3d4cdda4a / dfe8c5a16 / a88c140af,
mksettings consumes 8d0a22be9, raid-sim threat count e05fd732e, UI and leveling work).

## Checked
- Shared run (20k, seed 1, 77 requests, master binary): builds.json matches the run on all 77 within 0.15; every meta's published.dps
  matches too (results-20261030-1 stays resolved, no stale metas).
- Moves against the previous req_summary.txt (14 over 0.15), all explained by 3d4cdda4a (sim change e1fcd57b7):
  - Ret / Ret sub Holy and their duties +12.5 to +16.0 (free twist window).
  - Arms (4 builds) +2.7 to +4.8, Fury 2H (3) +1.5 to +2.9 (Hamstring from 10 Rage).
  - Mage Frost 908.3 -> 911.6 (Undead). Hunters unmoved (Efficiency on Strider Kick is mana only).
  Sizes equal the commit message's; builds.json carries them.
- Drift (regen.py vs ref): identical to the reference. Unread fields: none (not even notes).
- No new finding; none open.

## Reference
req_summary.txt replaced with this run (77 requests); baseline e05fd732e. Ranking explanation unchanged apart from the Ret, Arms and
Fury 2H lifts above.
