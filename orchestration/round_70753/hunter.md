# Hunter: round 70753 (the November 2 development notes)

Worker: the Hunter class session of the round 70753 full pass. Branch `round-70753-hunter`.

## Status

Phase A (research and mechanics) done. Phase B (re-optimize, races, audit, land) waits for "sources landed".

## Sources read

- Blizzard's development notes post 9 (2026-11-02), `orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt`, "- Hunter", Races and Bug Fixes.
- Client 1.60.1.70753 against 70707 (`wago.py fetch 1.60.1.70753 --set client_update --set talents`, `db2diff.py 70707 70753`). The Hunter section changes no damage, cost or talent value: only `Attributes_14 0 -> 64` on every spell, Freezing Trap / Frost Trap's category (2183 -> 411, and category 411 renamed "Fire Traps" -> "Hunter Traps"), Lone Wolf's Attributes_3 (+0x100000) and the racial Command (20576) added. `import_db2.py 1.60.1.70753 hunter`: "0 structure, 0 numbers, 0 wording", so the Hunter tree is unchanged from the one on master.
- **Server hotfix push 113534 (tier 1)**, 2026-11-03 10:15 UTC, region 93, client build 70707, from wago.tools' hotfix log (`hotfixes.py fetch 1.60.1.70707 SpellEffect CurvePoint` and `--search 113534`, 70 of its 73 rows). It carries every Hunter number in the notes, and foreverchanges.pro's "November 3 hotfix" entry reads the same rows. Wago has no hotfix rows for build 70753 yet.
- Wowhead's Forever tooltips (dataEnv 16) still show the old values (Aimed Shot 166, Arcane Shot 217, Hawk 55, Trueshot Aura 50): below the hotfix.

The tiers agree with each other here. The client's 70753 tables carry the 10-28 hotfix values (Hawk 15 / 22 / 30 / 38 / 47 / 55, Trueshot Aura 15 / 22 / 30 / 40 / 50), and the 11-03 hotfix (tier 1, newer) replaces them with the notes' values. So the brief's example disagreement (Trueshot Aura) resolves in favour of the notes' numbers, by the hotfix.

## Ledger

| Note line | Tier and evidence | Sim status (before) | Fix |
|---|---|---|---|
| Hunter's Mark can no longer be applied multiple times with different ranks | notes; sim's mark is one exclusive effect | landed | none: `core.huntersMarkAura` registers one "HuntersMark" exclusive effect, the strongest wins |
| Rank 4 Hunter's Mark less RAP than rank 3 fixed | client 70753 14325 e1 bp 143 (rank 3 98); 10-21 hotfix | landed | none (`ForeverHuntersMarkRank4` 143 since 2026-10-23) |
| Explosive Trap debuff wording | notes | n/a | text only |
| All traps share one 30 s cooldown | client 70753 SpellCategories: Freezing Trap 1499 / 14310 / 14311 and Frost Trap 13809 2183 -> 411 ("Hunter Traps") | wrong (Fire and Frost had a timer each) | one timer for every trap (`sim/hunter/hunter.go`); `TestForeverTrapsShareCooldown`. No preset casts a Frost trap, so no DPS moves |
| Aimed Shot bonus +~110% | hotfix 113534: 40 / 70 / 115 / 185 / 260 / 350 (was 20 / 34 / 55 / 89 / 125 / 166) | missing | `AimedShotBonus` (`aimed_shot.go`); `TestForeverHotfix20261103Damage` |
| Arcane Shot base +~30%, AP 0.10 -> 0.15 | hotfix 113534: 25 / 35 / 50 / 85 / 120 / 175 / 220 / 280; the AP share is server side (hotfix row's BonusCoefficientFromAP still 0), so it is tier 4 (notes), the same 10% the sim had by report assumption `hunter_shot_ap_shares` | missing | `ArcaneShotBaseDamage`, `arcaneShotRAPCoefficient` 0.15 |
| Serpent Sting AP 3% -> 5% a tick, base +~20% | hotfix 113534 ticks 3 / 7 / 14 / 26 / 40 / 58 / 75 / 100 / 135 (x5: 15 ... 500, rank 9 675); the AP share is tier 4 (notes) as above | missing | `SerpentStingDamage`, `serpentStingRAPPerTick` 0.05 |
| Aspect of the Hawk 20 / 35 / 50 / 70 / 90 / 110 | hotfix 113534 (client 70753 has the 10-28 hotfix's 15 / 22 / 30 / 38 / 47 / 55) | wrong (55) | Forever takes Classic's table (`aspects.go`); `TestForeverHotfix20261103Aspects` |
| Pets keep their innate attack speed passives | notes (a bug on untraining) | landed | none: the sim always applies the beast's Faster / Slower Attack passive (`foreverPetAttackSpeedPercent`) |
| Pathfinding moved row 2 -> 3 | 10-30 hotfix, already in the tree (`hunter.go` header) | landed | none (movement, not simmed) |
| Barrage 3 / 7 / 10 -> 5 / 10 / 15% | hotfix 113534 CurvePoint 82803 = 5 / 10 / 15 (client 70753 3 / 7 / 10) | missing | `barrageMultiplier` (`talents.go`); `TestForeverHotfix20261103Talents` |
| Sniper Shot bonus +100% | hotfix 113534: 320 / 450 / 590 (was 160 / 225 / 295) | missing | `SniperShotBonus` (`sniper_shot.go`) |
| Trueshot Aura 20 / 35 / 50 / 75 / 100 | hotfix 113534 (client 70753 15 / 22 / 30 / 40 / 50) | wrong (50) | `foreverTrueshotAuraRAP` in `sim/core/buffs.go` (shared file, Forever only, the hunter's own buff); `TestForeverBestLearnedRank` |
| Lacerating Strikes always overwrites | notes; hotfix 113534 changes its SpellMisc 857128 (attributes) | landed | none: every landed Mongoose Bite recasts the bleed with the new damage (`procLaceratingStrikes`) |
| Lightning Reflexes 2 -> 3% Agility a rank | hotfix 113534 CurvePoint 109548 = 3 / 6 / 9 / 12 / 15 (client 70753 2 ... 10) | missing | `talents.go` 0.03 a rank; `TestForeverHotfix20261103Talents` |
| Finding mechanics-20261031-1: Efficiency discounts Lacerate | client 70425 to 70753 19416 mask word 1 0x80081000 holds Lacerate's 0x1000 | wrong | Lacerate in Efficiency's list; `TestEfficiencyClassMask` wants 85. No report build takes both, so no DPS moves |
| Low-rank proc penalty removed | notes | n/a for Hunter (A2 has the shared part) | none |
| Roots from procs get diminishing returns (Entrapment, Improved Wing Clip, Counterattack) | notes | n/a | PvP / control |
| Throw's 0.5 s windup | notes, client 2764 CastingTimeIndex 3 -> 1 | n/a (no Hunter preset throws; A2's) | none |
| Races: Dwarf Artisan's Precision, Human Ruthless 1%, Orc Command / Blood Fury 2 min, Troll Berserking 15 s, Tauren Endurance, Skyborne 2% haste, Undead Touch of the Grave, base stats | notes | A2 (already in `sim/core/racials_forever.go` on master: Command for hunter and warlock pets) | Phase B race sweep |
| Threat (Disengage etc.) and healing rows | | n/a | tank/healer sessions |

The Arcane Shot and Serpent Sting attack power shares are the only notes values no client or hotfix row shows (they are server side, as they were at 10% / 3%). They stay report assumption `hunter_shot_ap_shares`, now at the notes' 15% / 5%.

## Notes-only values (standards 37)

Values neither client 70753 nor a hotfix row carries; the sim takes the notes' value. Check them against the next hotfix or launch logs.

| Value | Notes | Before | Where |
|---|---|---|---|
| Arcane Shot's attack power share | 0.15 ("raised from 0.1") | 0.10 (report assumption `hunter_shot_ap_shares`) | `arcaneShotRAPCoefficient`, `sim/hunter/arcane_shot.go` |
| Serpent Sting's attack power share a tick | 5% ("increased to 5% from 3%") | 3% (same assumption) | `serpentStingRAPPerTick`, `sim/hunter/serpent_sting.go` |

Both shares were server side before too (the hotfix rows' BonusCoefficientFromAP is 0). The Hunter shot files hold Forever's values with no Classic branch (Aimed Shot's 166 was already ungated), so these are Forever-only like the rest of the file.

## Phase A DPS effect

The report's 10 Hunter requests as they are on master (old talents, rotations and race), 180 ± 20 s, 5,000 iterations, seed 1, master 1839885 against this branch:

| Build | Before | After | Change |
|---|---|---|---|
| beast-mastery | 1045.9 | 1191.8 | +13.9% |
| beast-mastery-wolf | 1002.0 | 1147.8 | +14.5% |
| beast-mastery-tallstrider | 988.1 | 1133.9 | +14.7% |
| marksmanship | 1065.8 | 1244.9 | +16.8% |
| marksmanship-wolf | 1041.5 | 1220.8 | +17.2% |
| marksmanship-tallstrider | 1033.7 | 1212.9 | +17.3% |
| lone-wolf | 1090.4 | 1300.9 | +19.3% |
| survival | 1024.8 | 1041.9 | +1.7% |
| survival-wolf | 998.9 | 1015.9 | +1.7% |
| survival-tallstrider | 990.1 | 1007.3 | +1.7% |

Survival gains only Lightning Reflexes and Serpent Sting: its melee abilities, the bulk of its damage, have no line in the notes. Phase B re-optimizes the talents and rotations against the new numbers (Aimed Shot is now worth far more than its cast, Barrage and Lightning Reflexes more points).

Goldens: the five Hunter goldens move (these changes). The Balance and Feral CharacterStats goldens and the other classes' lines in `sim/testdata/report_hashes.txt` move only by the raid's Trueshot Aura on the stat sheet (ranged attack power +50); their DPS is the same to the last digit (rogue combat-swords, warrior fury, mage fire, paladin level_60, druid feral-cat checked at the same seed).

## Open questions and shared changes

- Shared change: `sim/core/buffs.go` Trueshot Aura ranks (Forever only; only hunters use ranged attack power) and its test `sim/core/forever_ranks_test.go`.
- `orchestration/spec_assumptions.md` row `hunter_tsa_hm_rank` still says Trueshot Aura 50; the orchestrator owns that file (to update: 100 by hotfix 113534).
