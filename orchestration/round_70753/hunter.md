# Hunter: round 70753 (the November 2 development notes)

Worker: the Hunter class session of the round 70753 full pass. Branch `round-70753-hunter`.

## Final summary (2026-11-03)

Landed on master at 7711ade16 (2026-11-03). Phase A and B done. Full suite (`-timeout 40m`) green on the merged tree before the push.

### What changed

| Change | Source | DPS effect (180 s) |
|---|---|---|
| Aimed Shot +40 ... +350, Arcane Shot 25 ... 280, Serpent Sting 15 ... 675, Sniper Shot +320 / 450 / 590 | server hotfix push 113534, 2026-11-03 (tier 1) | the bulk of the ranged gain |
| Aspect of the Hawk 110 (rank 6), Trueshot Aura 100 (rank 5) | same hotfix (client 70753 still 55 / 50) | ranged builds |
| Barrage 5 / 10 / 15%, Lightning Reflexes 3% Agility a rank | same hotfix (CurvePoint 82803, 109548) | MM / LW (Barrage), Survival (LR) |
| Arcane Shot 15% of RAP, Serpent Sting 5% a tick | Blizzard's 11-02 notes only (notes-only, standards 37) | all four |
| One 30 s cooldown for every trap | client 70753 SpellCategories | 0 (no preset casts a Frost trap) |
| Efficiency discounts Lacerate (mechanics-20261031-1) | client 70425 to 70753 mask | 0 (no build takes both) |
| Rotations: MM and LW Aimed Shot at any mana, MM Serpent Sting first, BM weave with no shot slack | rotopt knob searches, 20k on two seed sets | MM +10.0, LW +10.0, BM +1.6 |
| Beast Mastery: Improved Stings 2 -> 1, Improved Arcane Shot 3 -> 4 | rotopt talent search | BM +2.2 |
| Classic race base stats, racials (A2, master) | shared | a few DPS a race |

Measured all together, the master requests before the round against this branch's (20,000 iterations, seed 1, each length ± 20 s):

| Build | Talents | Race | Old 120 / 180 / 300 | New 120 / 180 / 300 | 180 s change |
|---|---|---|---|---|---|
| beast-mastery | 5021201505021241-00510504005 | Orc | 1049.7 / 1046.5 / 1033.3 | 1197.3 / 1195.6 / 1176.3 | +14.2% |
| beast-mastery-wolf | same | Orc | 1004.4 / 1001.8 / 990.0 | 1151.2 / 1150.5 / 1132.7 | +14.8% |
| beast-mastery-tallstrider | same | Orc | 991.2 / 988.5 / 976.6 | 1138.3 / 1136.9 / 1119.2 | +15.0% |
| marksmanship | 502-0053551501503051-5 | Orc | 1081.1 / 1066.3 / 1039.1 | 1268.9 / 1254.3 / 1217.6 | +17.6% |
| marksmanship-wolf | same | Orc | 1055.7 / 1040.9 / 1014.7 | 1243.1 / 1228.3 / 1192.4 | +18.0% |
| marksmanship-tallstrider | same | Orc | 1047.7 / 1033.2 / 1006.9 | 1236.0 / 1220.9 / 1185.8 | +18.2% |
| lone-wolf | 5-0053552511503051-5 | Night Elf | 1106.7 / 1089.9 / 1055.5 | 1332.4 / 1313.7 / 1264.7 | +20.5% |
| survival | 502-005005-500230230250122151 | Orc | 1031.9 / 1024.6 / 1016.3 | 1048.6 / 1041.2 / 1032.7 | +1.6% |
| survival-wolf | same | Orc | 1004.6 / 998.6 / 990.8 | 1020.7 / 1014.5 / 1006.9 | +1.6% |
| survival-tallstrider | same | Orc | 997.2 / 990.7 / 983.1 | 1013.8 / 1007.4 / 999.5 | +1.7% |

The old 180 s column reproduces every published number. Lone Wolf is the class's best at 180 s; the UI default stays Beast Mastery, as it was before (Lone Wolf already led then).

### Talents

Searched at 180 s (2,000 iterations a candidate, hill climb, gear fixed) from each current build and from a Marksmanship / Survival hybrid (0/21/30) that reaches Lightning Reflexes' 15%. Marksmanship, Lone Wolf and Survival hold. Beast Mastery moves a point from Improved Stings to Improved Arcane Shot: +2.2 / +1.9 at 180 s on seeds 1 / 1000003 (+0.9 / +0.8 at 120 s, -2.3 / -2.6 at 300 s; talents are judged at 180 s). The hybrid ends at 1163.4 against deep Marksmanship's 1244.9 (-6.5%); it isn't added as a preset (the report's build list is the orchestrator's; question below). The Hunter tree is unchanged in 70753 (A1's remap: none).

### Rotations

Knob searches over 120 / 180 / 300 s combined (3,000 iterations, frozen: reaction times, markPrepull, leylineMana, hawk, intimidation), confirmed at 20,000 on seeds 1 and 1000003, positive at every length on both (`tools/rotopt/hunter_presets.sh`):

- Marksmanship: `aimedMana` 40 -> 0, `serpentLate` 1 -> 0. +1.3 / +10.0 / +12.8 (seed 1000003 +0.9 / +9.0 / +11.9).
- Lone Wolf: `aimedMana` 40 -> 0. +1.8 / +10.0 / +9.8 (+2.1 / +10.2 / +10.7).
- Beast Mastery: `weaveShotSlack` 0.75 -> 0 (walk in only when no shot comes off cooldown before it's back). +1.3 / +1.6 / +2.0 over its new talents (+0.7 / +1.5 / +2.2).
- Survival: nothing moves.

### Races (verify.py with each race's own gear and rotation files, 180 s, 20,000 iterations, seeds 1 / 1000003)

| Build | Best overall | Best Horde | Best Alliance | Preset race |
|---|---|---|---|---|
| Marksmanship | Orc 1254.3 / 1253.7 (Night Elf 1254.3 / 1253.6, a tie) | Orc | Night Elf | Orc (kept: no lead over 0.5) |
| Lone Wolf | Night Elf 1313.7 / 1313.4 | Windshaper 1309.2 / 1308.6 | Night Elf | Night Elf |
| Beast Mastery | Orc 1195.6 / 1195.2 | Orc | Night Elf 1189.6 / 1189.3 | Orc |
| Survival | Orc 1041.2 / 1041.3 | Orc | Night Elf 1034.0 / 1034.5 | Orc |

Full orders (seed 1): Marksmanship Orc, Night Elf, Human 1247.1, both Skyborne 1246.3, Troll 1242.9, Dwarf 1240.4, Tauren 1229.7. Lone Wolf Night Elf, both Skyborne, Orc 1307.0, Human 1305.9, Troll 1303.2, Dwarf 1298.9, Tauren 1287.1. Beast Mastery Orc, Night Elf, both Skyborne 1182.8, Human 1182.7, Troll 1179.8, Dwarf 1176.9, Tauren 1168.2. Survival Orc, Night Elf, both Skyborne 1031.8, Human 1030.4, Troll 1027.6, Dwarf 1024.0, Tauren 1017.5. No race-tuned rotation was re-searched (the only race files are the Dwarf's hit swaps, which still apply).

### Audit (passes results, rotations, gear, consumables, mechanics, client_update, applied to the Hunter builds)

- results: each of the 10 requests at 20,000 iterations, seed 1, equals its new `meta.published` / `meta.verify` (written from the same run). Every move is explained by the commits above. Old numbers reproduce the published ones exactly.
- rotations: breakdowns of Marksmanship and Beast Mastery at 180 s: Serpent Sting up all fight (11.4 / 11.9 casts), Arcane Shot and Aimed Shot on their cooldowns, Multi-Shot nearly never (expected), Sniper Shot 9.1 of 12 under its mana gate, Rapid Fire and racials on cooldown, no capped mana. Every change is positive at 120, 180 and 300 s on two seed sets (standards 31). Rotation lines keep their template intent text. No new finding.
- gear: no gear change (synthetic sets fixed by the rules). Not re-swept for the new stat values.
- consumables: no change; not re-swept (Major Mana Potion vs Major Frenzy Potion could move with the bigger shots; under 1 DPS last time on the old numbers, so left).
- mechanics and client_update: the ledger below; every Hunter line of the notes, client 70753 and hotfix 113534 checked. Finding mechanics-20261031-1 fixed.

### Open questions and shared changes

- Shared change: `sim/core/buffs.go` Trueshot Aura ranks (Forever only; ranged attack power, so only hunters' DPS moves) and `sim/core/forever_ranks_test.go`. It moves every report build's stat sheet by +50 ranged attack power, hence `sim/testdata/report_hashes.txt` and the Balance / Feral CharacterStats goldens; their DPS is unchanged.
- Orchestrator files to update: `orchestration/spec_assumptions.md` (`hunter_tsa_hm_rank` now 100 by hotfix; `hunter_shot_ap_shares` now 15% / 5% by the notes) and the report's assumption text for them.
- Stale caveat numbers in the metas I didn't re-measure: "Improved Tracking ... worth 43.8 / 51.8 DPS", "Wildstalker 4 / 5 piece would add 13 / 17", "Major Frenzy Potion 25.7 behind", the Mark-duty costs. They're on the old numbers.
- Question (recommendation: no): add a Marksmanship / Survival hybrid preset (-6.5%)? It's within the 10% rule but plays like Marksmanship with fewer shots; I'd leave it out.
- Commit attribution: the brief names another session's Claude-Session line; my commits carry this session's own, per its system attribution rule.

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

Both shares were server side before too (the hotfix rows' BonusCoefficientFromAP is 0). Both shares and the hotfix shot tables (Aimed Shot, Arcane Shot, Serpent Sting, Sniper Shot) are gated under `Env.IsForever()` (finding results-20261103-2): outside Forever Arcane Shot and Serpent Sting take Classic's ranks and no attack power share, Aimed Shot Classic's bonus, and Sniper Shot, which Classic doesn't have, the client's 160 / 225 / 295. The shot files' older beta values (casts, coefficients) are still ungated.

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
