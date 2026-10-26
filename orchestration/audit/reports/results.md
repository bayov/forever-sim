# Results audit, pass 54 (master f62cb7971, quiet)

- Why due: 9 watched files changed (sim/core x6, druid ferocious_bite, hunter multi_shot, builds.json). Source: 32a6848 "assume wowsims/forever's Multi-Shot cast, live DoT ticks, hunter RAP, Bite energy and channel regen" (standards 27 and 28), then the full re-sim 758605874 (v36, all 57 builds, on 32a6848's sim). No sim change after that.
- Shared run (57 requests, 20k, seed 1): 23 of 57 moved more than 0.5 DPS against the old reference; all are the seven new mechanics. Largest: Affliction -5.9 (-0.57%), Shadow -4.9, Shadow+Holy -4.8, Demonology -4.8, Balance -4.2, Lone Wolf -4.1, Smite -2.6, Marksmanship -3.2, Beast Mastery +2.2. Matches the 758605874 message (its list, rounded) and no rank change beyond Smite 30th to 31st.
- builds.json vs my run: all 57 within 0.15 (largest 0.1). The dashboard is exact.
- Metas' published.dps: not updated by the v36 re-sim (still the 2026-10-23/24 rounds), so runreq flags 26 requests vs the metas by up to 5.9. Known meta lag, not a sim gap; builds.json is the real published number.
- Drift: identical to the reference. regen.py fixed: it passed "Night Elf" (two feral metas) to mksettings, which wants "NightElf"; spaces are now stripped.
- Unread fields: only `notes`.
- Standing finding results-20261025-1 stays open (unchanged).
- Reference req_summary.txt replaced with this run (baseline f62cb7971).
