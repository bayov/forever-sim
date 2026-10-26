# Audit loop log

One line per pass, newest last: `<UTC time> | <auditor> | pass <n> | master <sha> | <result> | <summary>`.

2026-10-17 13:30 | all | - | master d2633f4 | moved | The seven long-running auditor sessions were retired; their last passes were quiet on d2633f4 (fairness on 1735382). State seeded from them.
2026-10-17 15:04 | results | pass 43 | master 0a204d4 | quiet | All 66 requests within 0.15 of published; shaman heal rank families don't touch DPS.
2026-10-17 15:04 | rotations | pass 27 | master 0a204d4 | quiet | No APL/req change; 0 of 66 presets moved over 0.5 DPS.
2026-10-17 15:04 | mechanics | pass 30 | master 0a204d4 | quiet | Only change is shaman heal rank families; no new client values.
2026-10-17 15:04 | rank_tuning | pass 1 | master 0a204d4 | quiet | 66 presets match baseline_d2633f4 within 0.001%.
2026-10-17 15:04 | fairness | pass 1 | master 0a204d4 | quiet | Full race sweep identical to baseline; 19 gear sets pass; ranking unchanged.
2026-10-17 18:56 | results | pass 44 | master 1e2f7ec5b | quiet | no new findings
2026-10-17 18:56 | rotations | pass 28 | master 1e2f7ec5b | quiet | no new findings
2026-10-17 18:56 | gear | pass 37 | master 1e2f7ec5b | quiet | no new findings
2026-10-17 18:56 | consumables | pass 33 | master 1e2f7ec5b | quiet | no new findings
2026-10-17 18:56 | mechanics | pass 31 | master 1e2f7ec5b | quiet | no new findings
2026-10-17 18:56 | fairness | pass 2 | master 1e2f7ec5b | quiet | no new findings
2026-10-18 06:49 | results | pass 45 | master d70ef8ca3 | quiet | 0 of 66 moved; leveling plan, hidden items only
2026-10-18 06:49 | consumables | pass 34 | master d70ef8ca3 | quiet | only leveling plan changed
2026-10-18 12:50 | results | pass 46 | master 209208949 | quiet | 0 of 66 moved; leveling plan, hidden items, db generation only
2026-10-18 12:50 | gear | pass 38 | master 209208949 | quiet | ForeverSeenInGame moved to json; no preset gear affected
2026-10-18 12:50 | consumables | pass 35 | master 209208949 | quiet | only leveling plan changed
2026-10-18 18:53 | mechanics | pass 32 | master b2eac4867 | quiet | database.go AddToDatabase export only; client unchanged
2026-10-19 00:51 | results | pass 47 | master e6d1a5ad9 | quiet | 66 requests within 0.15 of published; no number moved
2026-10-19 00:51 | consumables | pass 36 | master e6d1a5ad9 | quiet | mana consumable NeedsEvent flag only; leveling plan changed
2026-10-19 00:51 | mechanics | pass 33 | master e6d1a5ad9 | quiet | idle-check skipping commits change no DPS; client still 70450
2026-10-19 18:51 | results | pass 48 | master 753d81da7 | quiet | 66 requests within 0.15 of published; no number moved
2026-10-19 18:51 | consumables | pass 37 | master 753d81da7 | quiet | only leveling plan source changed in presets.ts; client still 70450
2026-10-19 18:51 | mechanics | pass 34 | master 753d81da7 | quiet | Stoneskin min-1-damage row verified; other commits are rotation logic; client still 70450
2026-10-20 00:48 | results | pass 49 | master 1ab3ec28a0 | quiet | Only presets.ts leveling swap; 0/66 builds moved
2026-10-20 00:48 | consumables | pass 38 | master 1ab3ec28a0 | quiet | Leveling preset swap only; scan clean
2026-10-21 18:48 | consumables | pass 39 | master 599f9de74 | quiet | Client 70450 -> 70532: 0 consumable changes; scan clean
2026-10-21 18:48 | mechanics | pass 35 | master 599f9de74 | quiet | 70532 tables byte-identical to 70450; verdicts hold
2026-10-21 18:48 | client_update | pass 1 | master 599f9de74 | 2 new findings | 70532 identical to 70450; Champion of the Light and mage Frost hotfixes unconfirmed
2026-10-25 00:49 | rotations | pass 29 | master 15f953ddef | quiet | no new findings
2026-10-25 00:49 | mechanics | pass 36 | master 15f953ddef | quiet | no new findings
2026-10-25 00:49 | fairness | pass 3 | master 15f953ddef | quiet | no new findings
2026-10-25 13:18 | results | pass 50 | master 54878fa3d04f | 1 new finding | mksettings and committed requests disagree on consumes (medium); published numbers unaffected
2026-10-25 13:18 | rotations | pass 30 | master 54878fa3d04f | quiet | new rotations win or tie everywhere; baseline refreshed
2026-10-25 13:18 | gear | pass 39 | master 54878fa3d04f | quiet | new hit cap moved hit shares 1 to 1.5 points; no stat more than 2 points from target
2026-10-25 13:18 | consumables | pass 40 | master 54878fa3d04f | quiet | new flasks, potions and explosives consistent between presets and mksettings
2026-10-25 13:18 | mechanics | pass 37 | master 54878fa3d04f | 1 new finding | Major Spellstone is Shadow only after a server hotfix (low)
2026-10-25 13:18 | fairness | pass 4 | master 54878fa3d04f | 1 new finding | Marksmanship report race Windshaper loses 2.5 DPS to Human (low)
2026-10-25 19:03 | results | pass 51 | master 016daa654 | quiet | no new findings; hunters moved -0.5 to -2.5 from own Hunter's Mark; results-20261025-1 still open
2026-10-25 19:03 | rotations | pass 31 | master 016daa654 | quiet | hunter duty presets removed, APLs match presets
2026-10-25 19:03 | gear | pass 40 | master 016daa654 | quiet | only a new low-level axe; no gear rule change
2026-10-25 19:03 | consumables | pass 41 | master 016daa654 | quiet | hunter duty presets removed, presets and mksettings agree
2026-10-25 19:03 | mechanics | pass 38 | master 016daa654 | quiet | two hunter ruling commits, no shared mechanic moved
2026-10-25 19:03 | fairness | pass 5 | master 016daa654 | quiet | no new findings; fairness-20261025-1 has new evidence (Human +2.6 on MM)
2026-10-26 01:06 | gear | pass 41 | master 74b3a5f7c | quiet | only standards.md (Rage ruling) changed, no gear files
2026-10-26 01:06 | consumables | pass 42 | master 74b3a5f7c | quiet | preset/assumption changes carry no consumes change
2026-10-26 01:06 | mechanics | pass 39 | master 74b3a5f7c | quiet | three 2026-10-25 ruling commits verified against client 70532
2026-10-26 01:06 | results | pass 52 | master 74b3a5f7c | quiet | no new findings; results-20261025-1 still open; moves explained by commits
2026-10-26 01:06 | rotations | pass 32 | master 74b3a5f7c | quiet | warrior/hunter moves explained by rulings; paired A/B holds
2026-10-26 01:06 | fairness | pass 6 | master 74b3a5f7c | quiet | fairness-20261025-1 resolved (MM Human best race)
