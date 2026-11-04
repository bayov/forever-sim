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
2026-10-26 12:56 | consumables | pass 43 | master f796d9953 | quiet | assumptions_def.py wording-only change; no consumes values moved
2026-10-26 12:56 | results | pass 53 | master f796d9953 | quiet | no new findings; 23 meta.json notes-wording only; 0 movement
2026-10-26 12:56 | fairness | pass 7 | master f796d9953 | quiet | caveats text only; no sweep needed
2026-10-26 19:20 | gear | pass 42 | master f62cb7971 | quiet | only standards.md changed (no gear rules)
2026-10-26 19:20 | consumables | pass 44 | master f62cb7971 | quiet | 7 new single-value assumptions; no consumable values moved
2026-10-26 19:20 | mechanics | pass 40 | master f62cb7971 | quiet | 5 mechanics (standard 28) read, code correct; mechanics-20261025-1 still open
2026-10-26 19:20 | results | pass 54 | master f62cb7971 | quiet | builds.json matches 20k run within 0.15; baseline moved
2026-10-26 19:20 | rotations | pass 33 | master f62cb7971 | quiet | moves explained by 32a6848f9; baseline refreshed
2026-10-26 19:20 | fairness | pass 8 | master f62cb7971 | quiet | all moves from 32a6848f9; baseline moved
2026-10-27 00:55 | consumables | pass 45 | master 99bd53ada | quiet | new Ret Shock build adds no consumables; scan clean
2026-10-27 00:55 | results | pass 55 | master 99bd53ada | quiet | new paladin__level_60_shock_ret 932.38 matches published 932.4
2026-10-27 00:55 | rotations | pass 34 | master 99bd53ada | quiet | 4 reorderings of the new Shock (Ret side) APL all lost
2026-10-27 00:55 | fairness | pass 9 | master 99bd53ada | quiet | Shock (Ret side) Human best race; paladin builds unchanged
2026-10-27 06:57 | consumables | pass 46 | master d3df9247f | quiet | client 70630: 0 consumable spell or enchant changes; scan clean
2026-10-27 06:57 | mechanics | pass 41 | master d3df9247f | 2 new findings | racial rework (dev notes post 7) not in sim; Amplify Curse no longer boosts Bane of Agony
2026-10-27 06:57 | client_update | pass 2 | master d3df9247f | 5 new findings | 70630 diff: Missile Barrage tooltip, Spellstone Shadow-only, Amplify Curse, Touch of the Grave, racial pass
2026-10-27 13:21 | results | pass 56 | master 06a898751 | 1 new finding | shared run matches builds.json; metas' published.dps stale (low)
2026-10-27 13:21 | rotations | pass 35 | master 06a898751 | quiet | no APL changes; Frost/Arcane A/B tests hold
2026-10-27 13:21 | gear | pass 43 | master 06a898751 | quiet | enchants JSON meta-only change
2026-10-27 13:21 | consumables | pass 47 | master 06a898751 | quiet | scan clean; no consumable changes
2026-10-27 13:21 | mechanics | pass 42 | master 06a898751 | quiet | racials, Spellstone, Ignite verified; Amplify Curse still open
2026-10-27 13:21 | fairness | pass 10 | master 06a898751 | 1 new finding | report races stale after racial pass: 13 builds off best race (high)
2026-10-27 18:48 | consumables | pass 48 | master 37a8a8ccd | quiet | build 70634: no consumable or enchant changes; scan clean
2026-10-27 18:48 | mechanics | pass 43 | master 37a8a8ccd | 1 new finding | Blood Fury/Berserking/Touch of the Grave dev-note changes not on master (PR #117 fixes)
2026-10-27 18:48 | client_update | pass 3 | master 37a8a8ccd | quiet | 70634 client-identical to 70630; devnotes unreadable (rate limit)
2026-10-28 01:10 | results | pass 57 | master 5d27b0015 | quiet | 58/58 sims match builds.json; metas' published.dps still stale (results-20261027-1, noted)
2026-10-28 01:10 | rotations | pass 36 | master 5d27b0015 | quiet | post-8 racial changes + race-tuned APLs: all presets up 3.5-7.4 DPS; Fury Death Wish sync still right; baseline refreshed
2026-10-28 01:10 | gear | pass 44 | master 5d27b0015 | quiet | only standards.md changed; no gear/item changes
2026-10-28 01:10 | consumables | pass 49 | master 5d27b0015 | quiet | race re-picks/comments only; scan clean
2026-10-28 01:10 | mechanics | pass 44 | master 5d27b0015 | quiet | Blood Fury/Berserking/Touch of the Grave match dev notes post 8; mechanics-20261027-3 resolved
2026-10-28 01:10 | fairness | pass 11 | master 5d27b0015 | quiet | all presets on best race; fairness-20261027-1 resolved; sweep.py/table.py helper fixes
2026-10-28 07:32 | results | pass 58 | master 224b8a4e0 | quiet | shared run matches builds.json; 9 race overrides reproduce their gains; results-20261027-1 still open
2026-10-28 07:32 | consumables | pass 50 | master 224b8a4e0 | quiet | client 70658 has 0 consumable or enchant changes; scan clean
2026-10-28 07:32 | mechanics | pass 45 | master 224b8a4e0 | quiet | 25 mechanics tables identical in 70658; mechanics-20261027-2 still open
2026-10-28 07:32 | fairness | pass 12 | master 224b8a4e0 | quiet | no tuned race beats its preset's race; Frostfire Gnome vs Human is a tie
2026-10-28 07:32 | client_update | pass 4 | master 224b8a4e0 | 2 new findings | 70658 hotfix: Aspect of the Hawk ranks 1-5 and Trueshot Aura ranks 1-4 lowered (hunter -0.7 to -1.3%)
2026-10-28 13:06 | results | pass 59 | master 1274e15b3 | quiet | 58 builds within 0.15 of published; hunters -19 to -25 DPS from the 73cedc9a9 Hawk/Trueshot hotfix
2026-10-28 13:06 | rotations | pass 37 | master 1274e15b3 | 1 new finding | hunter APLs use Hawk rank 6; rotations-20261028-1: BM/MM/LW meta caveats still cite old TSA 100 / Hawk rank 5 (low)
2026-10-28 13:06 | gear | pass 45 | master 1274e15b3 | quiet | only standards.md changed, no gear impact
2026-10-28 13:06 | consumables | pass 51 | master 1274e15b3 | quiet | Trueshot Aura ranks match client 70658
2026-10-28 13:06 | mechanics | pass 46 | master 1274e15b3 | quiet | hunter hotfix matches client 70658; mechanics-20261027-2 still open
2026-10-28 13:06 | fairness | pass 13 | master 1274e15b3 | quiet | only hunters moved (-0.1 to -2.2%), still rank 1-3; sweep.py race_override fix left uncommitted
2026-10-28 19:14 | results | pass 60 | master dd663e652 | quiet | 56 of 58 requests within 0.15 of published; frostfire and elemental metas slightly stale
2026-10-28 19:14 | rotations | pass 38 | master dd663e652 | quiet | standards 31 rotation rewrites lose nothing beyond noise; rotations-20261028-1 gets new evidence
2026-10-28 19:14 | gear | pass 46 | master dd663e652 | quiet | standards.md change is not a gear rule
2026-10-28 19:14 | consumables | pass 52 | master dd663e652 | quiet | assumptions_def.py wording only; scan.sh minor-armor expectation stale (3, not 2)
2026-10-28 19:14 | fairness | pass 14 | master dd663e652 | quiet | all race rows +0.0 vs baseline; top/bottom 1.58x; sweep.py race_override fix left uncommitted
2026-10-29 00:51 | consumables | pass 53 | master dcd927490 | quiet | Windfury Totem refactor only; client 70685 no consumable changes
2026-10-29 00:51 | mechanics | pass 47 | master dcd927490 | quiet | Windfury/Execute changes are metrics only; client 70685 only Adventurer's Tent
2026-10-29 00:51 | client_update | pass 5 | master dcd927490 | quiet | 70685 touches no level 60 DPS data (Adventurer's Tent, Camping only)
2026-10-30 01:00 | results | pass 61 | master d50909a37 | quiet | 75 of 77 within 0.15 of published; feral duty-ff -3.7 from caster Faerie Fire
2026-10-30 01:00 | rotations | pass 39 | master d50909a37 | quiet | new duty APLs keep debuffs up; baseline refreshed
2026-10-30 01:00 | consumables | pass 54 | master d50909a37 | quiet | only feral Faerie Fire comments changed in presets
2026-10-30 01:00 | mechanics | pass 48 | master d50909a37 | quiet | raid sim values match client 70685; Seal of Wisdom 50% proc assumed
2026-10-30 01:00 | fairness | pass 15 | master d50909a37 | quiet | 37 builds match baseline at every race; top/bottom 1.58x
2026-10-30 07:01 | results | pass 62 | master 36a772fd1 | quiet | 76 of 77 builds within 0.15 of published; mage frostfire meta stale (known)
2026-10-30 07:01 | consumables | pass 55 | master 36a772fd1 | 1 new finding | Elixir of Wisdom +3 Int in 70707, sim has 6 (low)
2026-10-30 07:01 | mechanics | pass 49 | master 36a772fd1 | 1 new finding | Tauren Endurance 2% crit damage now applied in 70707 (needs user ruling)
2026-10-30 07:01 | client_update | pass 6 | master 36a772fd1 | 4 new findings | 70707 carries the mage talent rework (high); 3 low
2026-10-30 12:58 | results | pass 63 | master bbbaf2e3c | quiet | 77 sims within tolerance of published; two over-tolerance builds match the reference run
2026-10-30 12:58 | gear | pass 47 | master bbbaf2e3c | quiet | forever_enchants.json meta stamp only
2026-10-30 12:58 | consumables | pass 56 | master bbbaf2e3c | quiet | Elixir of Wisdom 6 -> 3 matches client, resolves consumables-20261030-1
2026-10-30 12:58 | mechanics | pass 50 | master bbbaf2e3c | quiet | 8 code commits read, none change game values
2026-10-30 19:12 | results | pass 64 | master fb1485888 | 1 new finding | metas' published.dps stale for 46 of 77 builds (low)
2026-10-30 19:12 | rotations | pass 40 | master fb1485888 | 2 new findings | Ret twist window and Hamstring rage threshold now beat current APLs
2026-10-30 19:12 | consumables | pass 57 | master fb1485888 | quiet | imbue-fit and presets agree, no findings
2026-10-30 19:12 | mechanics | pass 51 | master fb1485888 | 1 new finding | wf_totem_icd/extra_attack_swing cards and standards.md contradict published numbers (medium)
2026-10-30 19:12 | fairness | pass 16 | master fb1485888 | 1 new finding | Frost race: Undead now beats High Order (low)
2026-10-31 13:01 | results | pass 65 | master e05fd732e | quiet | 77 builds match published; 14 moves all explained by e1fcd57b7 fixes
2026-10-31 13:01 | rotations | pass 41 | master e05fd732e | quiet | APL changes are last pass's three fixes; new 3000it baseline
2026-10-31 13:01 | gear | pass 48 | master e05fd732e | quiet | no gear content changed; demonology warlock set deleted
2026-10-31 13:01 | consumables | pass 58 | master e05fd732e | quiet | no consume/buff/enchant values changed
2026-10-31 13:01 | mechanics | pass 52 | master e05fd732e | 1 new finding | Efficiency should also cut Lacerate mana cost (low)
2026-10-31 13:01 | fairness | pass 17 | master e05fd732e | quiet | full race sweep; report race best everywhere, top/bottom 1.54x
2026-10-31 19:02 | results | pass 66 | master d96adcce1 | quiet | Sunder ramp (std 32) moved 36 physical builds -0.5 to -4.1 vs published; expected lag
2026-10-31 19:02 | rotations | pass 42 | master d96adcce1 | quiet | rogue APL note rewording only
2026-10-31 19:02 | gear | pass 49 | master d96adcce1 | quiet | only standards.md changed, no gear rule
2026-10-31 19:02 | consumables | pass 59 | master d96adcce1 | quiet | presets swap sunderArmor for sunderRamp, consistent with standards
2026-10-31 19:02 | mechanics | pass 53 | master d96adcce1 | 1 new finding | Moonkin/LotP hotfix to 3% all crit: stacking ruling needed (high if stack)
2026-10-31 19:02 | fairness | pass 18 | master d96adcce1 | 1 new finding | Demonology wears shared Destro gear but has its own target (medium, needs user ruling)
2026-11-01 13:18 | results | pass 67 | master dd053d24c | 1 new finding | 74 of 77 metas' published.dps stale vs builds.json after re-sim (low)
2026-11-01 13:18 | rotations | pass 43 | master dd053d24c | quiet | no APL change; only ranged hunters moved from Sapper reach; baseline refreshed
2026-11-01 13:18 | consumables | pass 60 | master dd053d24c | 1 new finding | ranged hunters keep Sapper worth +0.16 DPS at 25 yd, comments say +4.8 (low)
2026-11-01 13:18 | mechanics | pass 54 | master dd053d24c | 1 new finding | Shadow presets at 25 yd but Mind Flay range is 20 yd (low, no DPS change)
2026-11-01 13:18 | fairness | pass 19 | master dd053d24c | 2 new findings | caster distances have no standard (medium, needs user ruling); Affliction Human 1.4 under Orc (low)
2026-11-01 19:08 | results | pass 68 | master c23ef0440 | quiet | 66 of 77 builds match published; 11 off are stale report data (mage hotfixes, warrior Rage re-sims)
2026-11-01 19:08 | rotations | pass 44 | master c23ef0440 | 1 new finding | Frost mage rotation caps Arcane Blast at 3 stacks, a 4th above 70% mana gains up to +2.7 DPS (low)
2026-11-01 19:08 | gear | pass 50 | master c23ef0440 | quiet | only standards.md changed, no gear rules touched
2026-11-01 19:08 | consumables | pass 61 | master c23ef0440 | quiet | Moonkin/Leader of the Pack crit ruling applied; Sapper finding still open
2026-11-01 19:08 | mechanics | pass 55 | master c23ef0440 | quiet | Rage, crit auras, Kang proc and hotfix commits match sources
2026-11-01 19:08 | fairness | pass 20 | master c23ef0440 | 1 new finding | Arcane report race Gnome now 2.8 DPS under Undead after Arcane Blast hotfix (low)
2026-11-02 01:11 | results | pass 69 | master 5a6e9c6a6 | quiet | Shared run matches builds.json within 0.15 on 47; 30 stale ones traced to commits; results-20261101-1 updated
2026-11-02 01:11 | rotations | pass 45 | master 5a6e9c6a6 | quiet | 17 Level 60 APLs re-searched on Oct 30 hotfix trees; new baseline; rotations-20261101-1 closed as obsolete
2026-11-02 01:11 | consumables | pass 62 | master 5a6e9c6a6 | quiet | Only Ret Shock potion change; Sapper finding still open
2026-11-02 01:11 | mechanics | pass 56 | master 5a6e9c6a6 | quiet | 5 class hotfix-tree commits match sources; client 70707 still newest
2026-11-02 01:11 | fairness | pass 21 | master 5a6e9c6a6 | quiet | All moves traced to commits; fairness-20261101-3 resolved; new race-sweep baseline
2026-11-02 07:01 | results | pass 70 | master fa865a3b8b7a | quiet | All 73 builds within 0.15 of published; results-20261101-1 resolved
2026-11-02 07:01 | rotations | pass 46 | master fa865a3b8b7a | quiet | Only deep Holy Shock Ret APLs deleted; no live APL logic changed
2026-11-02 07:01 | consumables | pass 63 | master fa865a3b8b7a | quiet | ret_shock/ShockConsumes removal and Frostfire default; Sapper finding still open
2026-11-02 07:01 | mechanics | pass 57 | master fa865a3b8b7a | quiet | Only 231d5492a (preset removals); client 70707 still newest
2026-11-02 07:01 | fairness | pass 22 | master fa865a3b8b7a | quiet | Published within 1.0 on all builds; new sweep baseline; Orc Affliction already open
2026-11-02 12:58 | consumables | pass 64 | master 67084d0ec2 | quiet | Only assumptions_def.py moved; no consumable assumption changed; Sapper finding still open
2026-11-02 12:58 | mechanics | pass 58 | master 67084d0ec2 | quiet | sim.go reseed speedup (same rolls), drops.json test data; client 70707 still newest
2026-11-02 19:05 | results | pass 71 | master 796495d6fd | 1 new finding | Shared run matches builds.json on all 73; Elemental Mana Tide meta says 753.2 vs 752.8 (low)
2026-11-02 19:05 | rotations | pass 47 | master 796495d6fd | quiet | 19 changed APLs fine; new rotations win paired checks; new baseline 796495d
2026-11-02 19:05 | gear | pass 51 | master 796495d6fd | quiet | New Demonology set hits its gear_targets row
2026-11-02 19:05 | consumables | pass 65 | master 796495d6fd | quiet | No buff/consume changes; Sapper finding still open
2026-11-02 19:05 | mechanics | pass 59 | master 796495d6fd | quiet | Client 70707 still newest; Lightning Overload proc change matches client
2026-11-02 19:05 | fairness | pass 23 | master 796495d6fd | quiet | 0/73 off published; Orc Affliction evidence added to open fairness-20261101-2
2026-11-03 01:10 | results | pass 72 | master a0b306984d | quiet | 72/73 within 0.15 of published; mana-tide meta finding still open
2026-11-03 01:10 | rotations | pass 48 | master a0b306984d | quiet | Frostfire -10.9 from Ice Lance/Fire Blast shared CD, rotation order still wins; warlock Doom hold confirmed
2026-11-03 01:10 | gear | pass 52 | master a0b306984d | quiet | Only an empty resto gear set deleted; full audit skipped
2026-11-03 01:10 | consumables | pass 66 | master a0b306984d | quiet | 0 of 53 enchants changed on client 70753
2026-11-03 01:10 | mechanics | pass 60 | master a0b306984d | 1 new finding | Read Leyline cast is 0.5s on 70753, sim has 2s (low)
2026-11-03 01:10 | fairness | pass 24 | master a0b306984d | quiet | 73 builds within 0.5 DPS; top/bottom 1.49x
2026-11-03 01:10 | client_update | pass 7 | master a0b306984d | 1 new finding | Poised (rogue/druid energy) missing from sim (low)
2026-11-03 18:29 | results | pass 73 | master f5c99ad018 | 2 new findings | round 70753 phase 3 (forced): 68 shared sims match the class round files; Subtlety meta race "Night Elf"; hunter notes-only shares ungated (low)
2026-11-03 18:29 | rotations | pass 49 | master f5c99ad018 | quiet | round 70753 phase 3 (forced): 28 APLs reworked; dead lines, talent legality, duties, resources and paired probes clean
2026-11-03 18:29 | gear | pass 53 | master f5c99ad018 | quiet | round 70753 phase 3 (forced): Restoration shaman set within budget; enchants file unchanged for presets
2026-11-03 18:29 | consumables | pass 67 | master f5c99ad018 | quiet | round 70753 phase 3 (forced): Trueshot Aura notes values, potions can't crit; no preset consume changes
2026-11-03 18:29 | mechanics | pass 61 | master f5c99ad018 | 1 new finding | round 70753 phase 3 (forced): notes' tank threat values (Bear 115%, Def Stance 80%, Defiance) not in the sim (low, 0 DPS)
2026-11-03 18:29 | fairness | pass 25 | master f5c99ad018 | quiet | round 70753 phase 3 (forced): every move matches its round file; top/bottom 1.73x; race picks hold
2026-11-03 18:29 | client_update | pass 8 | master f5c99ad018 | 1 new finding | round 70753 phase 3 (forced): no build after 70753; several notes-only values ungated under Forever (low, needs ruling)
2026-11-03 19:03 | consumables | pass 68 | master f23ea25d2 | quiet | Warlock Affliction/Imp-sacrifice diff touches no consumes; client 70755 has no consumable changes
2026-11-03 19:03 | mechanics | pass 62 | master f23ea25d2 | quiet | Imp sacrifice (Burning Shadow 15%) matches client; 70755 tables identical to 70753
2026-11-03 19:03 | client_update | pass 9 | master f23ea25d2 | 1 new finding | build 70755 technical only; server hotfix push 113483 changes rogue talent tree (high), rest of push unverified
2026-11-03 19:03 | results | pass 74 | master f23ea25d2 | quiet | all 68 requests within 0.15 of published except known stale shaman metas; results-20261103-1 resolved
2026-11-03 19:03 | rotations | pass 50 | master f23ea25d2 | quiet | Wrack Affliction and Earth Shock wait hold; 5 warlock variants all lose 2-22 DPS
2026-11-03 19:03 | fairness | pass 26 | master f23ea25d2 | quiet | every report race still best; Subtlety Night Elf fix confirmed
2026-11-04 01:07 | results | pass 75 | master f331f0cff | 0 new findings | 68 builds match published within 0.15; six shaman metas stale (update on results-20261102-1)
2026-11-04 01:07 | rotations | pass 51 | master f331f0cff | quiet | Feral/Balance/Ret rotation changes confirmed (+7.7..8.1 feral, +0.4..0.6 balance); no Forever DPS moved by sim code
2026-11-04 01:07 | gear | pass 54 | master f331f0cff | quiet | only enchants meta stamp changed (build 70755)
2026-11-04 01:07 | consumables | pass 69 | master f331f0cff | 1 new findings | presets.ts still has salvation:true after mksettings dropped it (consumables-20261104-1, low)
2026-11-04 01:07 | mechanics | pass 63 | master f331f0cff | quiet | Salvation removal and tank threat values match dev notes post 9
2026-11-04 01:07 | fairness | pass 27 | master f331f0cff | quiet | all 43 builds re-swept, report race best everywhere, top/bottom 1.68x
2026-11-04 UTC | results | pass 76 | master fb584eafa | quiet | 68 sims match metas within 0.15 except six stale shaman metas (results-20261102-1)
2026-11-04 UTC | rotations | pass 52 | master fb584eafa | quiet | 7 one-line Fury APL variants, none beat the new level60_fury
2026-11-04 UTC | gear | pass 55 | master fb584eafa | quiet | only standards.md 38/39 changed, no gear rule
2026-11-04 UTC | consumables | pass 70 | master fb584eafa | quiet | Salvation dropped from all 12 presets, resolved consumables-20261104-1
2026-11-04 UTC | mechanics | pass 64 | master fb584eafa | quiet | no sim code changed; dev notes post 9 Swiftmend edit needs no sim change
2026-11-04 UTC | fairness | pass 28 | master fb584eafa | quiet | Fury rebuilt per race, Orc best, gear check 19/19
2026-11-04 13:12 | results | pass 77 | master 6b480b0604 | 1 new finding | results-20261104-1: five rogue metas' published DPS lag hotfix 113483
2026-11-04 13:12 | rotations | pass 53 | master 6b480b0604 | quiet | rogue/warrior hotfix rotations hold; baseline refreshed
2026-11-04 13:12 | gear | pass 56 | master 6b480b0604 | quiet | new healer sets pass
2026-11-04 13:12 | consumables | pass 71 | master 6b480b0604 | quiet | no consumable value changes
2026-11-04 13:12 | mechanics | pass 65 | master 6b480b0604 | quiet | downrank ruling settled; Power Infusion healing verified
2026-11-04 13:12 | fairness | pass 29 | master 6b480b0604 | quiet | only warrior/rogue moved, from hotfixes
2026-11-04 UTC | healer_gear | pass 1 | master a1b1163d2 | quiet | forced first pass: 4 healer sets within 1b budget/shape and row shares, relics modelled; Wowhead cache empty so rows not re-derived from pages
