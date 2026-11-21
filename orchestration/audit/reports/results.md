# Results pass 96 (master 93e568c50, 2026-11-21)

Trigger: Wowhead scrape commit changed assets/database (db.bin, db.json, leftover_db.*).
- Scrape adds items (new Five Thunders set pieces, Idol of the Huntress source) and refreshes recipes and class spells. No sim, proto, request or preset change.
- Shared run (68 builds, 20k, seed 1) vs metas: 45 of 68 moved more than 0.5, the same families and sizes as results-20261104-2 (Affliction +38 to +43, Lone Wolf -22.4, rogues -10 to -20, rest <= 6.6). Nothing new.
- Only test AllItems results changed (shaman Five Thunders set, not used by any request).
- No finding opened or resolved. Open findings unchanged.
