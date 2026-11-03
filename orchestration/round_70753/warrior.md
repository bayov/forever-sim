# Warrior: round 70753 (November 2 development notes)

Worker: the Warrior class worker, branch `round-70753-warrior`. Brief: `briefs/warrior.md`.

## Final summary (Phase B landed, 2026-11-03)

### What changed

None of the warrior lines are in client 70753 or a hotfix, so every value below is a notes-only value (standards 37,
listed under "Notes-only values"), gated under Forever.

| Change | Source | DPS effect |
|---|---|---|
| Slam 15 sec (18); Improved Slam 2.5 / 5 sec off it (3 / 3) | notes | Arms and two-hand Fury Slam more often |
| Two-Handed Weapon Specialization 2 / 4 / 6% (1 / 2 / 3%) | notes | +3% on every two-hand build |
| Improved Tactical Mastery 3 points, 5 / 10 / 15 Rage | notes | frees 2 points (Arms Recklessness swap keeps 25 Rage) |
| Bloodthrill 2 points, 10 / 20% | notes | frees 3 Arms points for the same 20% |
| Flurry 3 points, 10 / 20 / 30%, no Death Wish prerequisite | notes | +5% attack speed on Fury, 2 points freed |
| Reckless Swings (new, Fury row 4) | notes; assumptions: off the GCD, landed swings incl. Heroic Strike, flat 5 Rage, no health cost | every build takes it; never casting it costs Fury 29 DPS at 180 sec |
| Improved Whirlwind (new, Fury row 6) | notes | both Fury builds take 2/2 |
| Enrage row 3, Gore Drinker row 5 under Blood Craze, Bloodthirst needs Death Wish | notes | tree only |
| Race base stats (A2), the 10-26 racial pass already in | notes | Arms re-picks Orc |

Tree: `tools/forever_talents/notes/warrior.json` (the columns of Enrage, Reckless Swings, Gore Drinker and Improved
Whirlwind are guessed; the notes give rows). Fury's tree is 19 talents (proto fields renumbered, `TalentTreeSizes`
17/19/18). Every warrior talent string in the repo is remapped, the tank ones too (their Fury points only move position).

### Per build (Orc, 20,000 iterations, seed 1; 180 is 180 +- 20 sec)

| Build | Old 120 / 180 / 300 | New 120 / 180 / 300 | Talents | Race |
|---|---|---|---|---|
| Fury (dual wield) | 1079.5 / 1041.5 / 1017.9 | 1194.2 / 1152.6 / 1124.8 | `30313113002-0505300551120100321` | Orc (Orc) |
| Fury Sunder duty | 1044.6 / 1012.5 / 991.4 | 1156.5 / 1117.7 / 1095.4 | same | Orc |
| Fury Battle Shout duty | 1076.0 / 1039.5 / 1015.7 | 1193.2 / 1150.5 / 1123.2 | same | Orc |
| **Fury 2H (new class default)** | 1085.2 / 1044.4 / 1009.4 | 1211.7 / 1169.9 / 1133.4 | `30313113032-2503300501122100321` | Orc (Orc) |
| Fury 2H Sunder duty | 1044.9 / 1009.0 / 980.5 | 1173.5 / 1132.6 / 1101.9 | same | Orc |
| Fury 2H Battle Shout duty | 1083.0 / 1041.3 / 1007.7 | 1210.5 / 1165.4 / 1131.5 | same | Orc |
| Arms | 1022.9 / 1008.7 / 987.0 | 1111.7 / 1097.7 / 1070.6 | `30303213032215201-050500050012-2` | Orc (was Undead) |
| Arms Sunder duty | 978.5 / 970.4 / 955.2 | 1055.0 / 1047.1 / 1027.5 | same | Orc (was Undead) |
| Arms Battle Shout duty | 1022.2 / 1004.9 / 984.1 | 1111.0 / 1094.0 / 1068.3 | same | Orc (was Undead) |
| Arms vs Dragonkin | 1066.0 / 1050.3 / 1028.8 | 1157.5 / 1142.2 / 1114.2 | `30303213132215201-050500050012-1` | Orc (was Undead) |

Old numbers: the deployed report (`ui/dps_report/data/builds.json`, 2026-11-02; its 120 / 300 at 5,000 iterations).
The request metas carry the new 180 as `published` and `verify`.

Talents (180 sec talent search on the synthetic sets, then the knob search, alternated until neither moved):
- Fury: Booming Voice and a point of Improved Berserker Rage out; Reckless Swings, Improved Whirlwind 2/2, Enrage 5/5 in.
- Fury 2H: Flurry's 3 points left Bloodthirst short of its row, so the start takes Reckless Swings and Improved
  Whirlwind; the search moves 2 points of Unbridled Wrath to Enrage 5/5. Furious Precision 3 is dead with a
  two-hander, but no move of its points scored better.
- Arms: the 5 freed points and a third point of Improved Heroic Strike (to reopen Two-Handed Weapon Specialization's row)
  go to Reckless Swings, Enrage 5/5 and Improved Bloodrage 2/2, for Booming Voice. Dragonkin: Spearing Strike for a
  point of Improved Bloodrage.

Rotations (`tools/rotopt/warrior_presets.sh`, dated block 2026-11-03): Reckless Swings on cooldown everywhere; Arms and
Dragonkin Slam ahead of Mortal Strike, any time (+3.2 to +3.8 at 180); Arms Sunder duty Slam right after a swing;
Arms Battle Shout duty Slam first; Fury Sunder duty Heroic Strike from 70 and Bloodthirst first (+6.5 / +7.3); Fury 2H
Sunder duty fillerGap=1 (+2.6 / +3.4). Each kept only where it won at 120, 160, 180, 200 and 300 sec on seeds 1 and
1000003 at 20,000 iterations (standards 31 and the handoff rule).

### Races (180 +- 20 sec, 20,000 iterations, seeds 1 / 1000003; before the last Arms rotation change)

| Build | Best overall | Best Horde | Best Alliance | Last |
|---|---|---|---|---|
| Fury | Orc 1152.6 / 1151.9 | Orc | Night Elf 1146.8 / 1146.0 | Tauren 1132.6 |
| Fury 2H | Orc 1169.9 / 1170.8 | Orc | Night Elf 1162.8 / 1162.9 | Tauren 1148.1 |
| Arms | Orc 1094.0 / 1094.1 | Orc | Night Elf 1088.5 / 1089.2 | Tauren 1076.6 |

All 10 races (seeds 1 / 1000003; Dwarf with its race gear override). Undead is second on Fury and Arms (3.1 to 4.9 behind), the two Skyborne races tie. Arms switches to Orc (+3.8 / +3.1 over Undead, past the 0.5 bar on both seed sets).

| Race | Faction | Fury | Fury 2H | Arms |
|---|---|---|---|---|
| Orc | Horde | 1152.6 / 1151.9 | 1169.9 / 1170.8 | 1094.0 / 1094.1 |
| Undead | Horde | 1148.5 / 1147.0 | 1161.9 / 1162.4 | 1090.2 / 1091.0 |
| NightElf | Alliance | 1146.8 / 1146.0 | 1162.8 / 1162.9 | 1088.5 / 1089.2 |
| SkyborneHighOrder | Alliance | 1145.8 / 1146.3 | 1157.3 / 1157.2 | 1085.7 / 1087.6 |
| SkyborneWindshaper | Horde | 1145.8 / 1146.3 | 1157.3 / 1157.2 | 1085.7 / 1087.6 |
| Troll | Horde | 1144.5 / 1144.1 | 1160.5 / 1161.1 | 1083.0 / 1083.8 |
| Human | Alliance | 1143.7 / 1143.4 | 1158.6 / 1158.9 | 1085.4 / 1085.6 |
| Gnome | Alliance | 1140.2 / 1140.1 | 1155.9 / 1157.2 | 1084.9 / 1085.3 |
| Dwarf | Alliance | 1139.2 / 1139.2 | 1153.8 / 1154.5 | 1081.4 / 1081.7 |
| Tauren | Horde | 1132.6 / 1131.5 | 1148.1 / 1149.1 | 1076.6 / 1076.8 |

### Audit (the DPS passes applied to the warrior builds)

- results: each request's 180 sec number is its own final run (20,000 iterations, seed 1); `report_hashes.txt` rewritten
  after the last merge, only the 10 warrior rows moved.
- rotations: breakdowns of Fury and Arms at 180 sec. Reckless Swings 12.5 casts a fight (the 15 sec cooldown allows 12
  plus the pull), each spent within about 2 sec on Fury (24.8% uptime) and 49% uptime on Arms' slower swings;
  Bloodthirst 27.2 of 30 possible casts, Whirlwind 19.6 with its 8 sec cooldown, Flurry 92% uptime; Arms Mortal Strike
  25.8, Overpower 17.2 (12 Bloodthrill windows), Rend 7.1, Slam 11.2. Every line states its intent (the Reckless Swings
  line: "for the Rage of the next 4 swings"). Found and fixed: the knob search's one-knob steps missed Slam-first on
  Arms and Dragonkin (a two-knob move), found from the Battle Shout duty's search.
- gear: sets unchanged; Orc has no hit racial, so no set goes over the cap. The Dwarf race overrides
  (`req/races/warrior__arms.Dwarf.json`, `warrior__fury_2h.Dwarf.json`) keep their gear swap; their `gain` fields are
  from before this round (the pipeline's race sweep re-measures them).
- consumables: unchanged. Two-hand Fury still carries the class DefaultConsumes' off-hand stone, which does nothing.
- mechanics / client_update: the ledger and the notes-only list below; every notes value has a test.
- Nothing open for the warrior in `orchestration/audit/findings/`.

### Open questions and shared changes

- Shared files touched: `proto/warrior.proto` (Fury's message regenerated: two new fields, field numbers from 25 on
  renumbered), `ui/core/talents/trees/warrior.json`, the tank warrior's talent strings (`ui/tank_warrior/presets.ts`,
  `tools/tankbench/req/warrior_protection.json`, `sim/warrior/tank_warrior/forever_test.go`: position only, values
  unchanged, tank goldens unmoved), `tools/audit/results/runclass.sh` (an example string), `sim/testdata/report_hashes.txt`.
- Not touched (others own them): `forever-wiki/sim-status.md` still quotes the old warrior talent strings and numbers;
  `ui/dps_report/data`, `ui/raid_sim/data/builds.json`.
- Question: the class default moves to two-hand Fury (the handoff rule: highest at 180 sec). It is the first time
  two-hand beats dual wield; the report's default badge follows `isDefault`. Recommendation: keep the rule.
- Question: Reckless Swings' GCD and which attacks count are not in the notes. Off the GCD and landed swings
  (Heroic Strike included) are my reading; with a GCD it would still be cast on cooldown. A client row settles it.
- Commit trailers use this session's own Claude-Session line, not the brief's (which names the orchestrator's session).

## Phase A: research and mechanics

### Sources checked (2026-11-03)

- Dev notes post 9 (2026-11-02), `orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt`, "- Warrior"
  section, Races and Bug Fixes.
- Client 1.60.1.70753 (wago.tools, `wago.py fetch 1.60.1.70753 --set client_update`). **None of the warrior changes in the
  notes are in it:** Slam's category cooldown is still 18000 on every rank (1464, 8820, 11604, 11605, 1240193, 1310196-200),
  Improved Slam 12862 effect 2 is still -3000 with no rank curve, there is no Reckless Swings spell, Improved Whirlwind
  24431 is the old Burning Crusade row (aura 107 misc 14, -30), Flurry 12966 still 30 on the buff with the 5 point tree.
  The ledger's 70753 section agrees: its Trait changes are in the paladin, druid and mage trees only.
- Server hotfixes for 70753 (`hotfixes.py fetch 1.60.1.70753 --set talents --set spelltext`): no rows at all yet.
- foreverchanges.pro /patch-notes: lists the warrior lines under "November 3, 2026 — Announced for launch / Blizzard's
  notes" (Defensive Stance 80%, Reckless Swings, Improved Whirlwind, Slam 15 sec, ...). Nothing in its data yet.
- So the notes are the highest tier that has these changes. They postdate the client and hotfix data, so there is no
  disagreement to resolve: the client simply hasn't caught up. Each code comment says "client 70753 still has the old
  value". When a hotfix or a newer build carries them, re-check the values (the notes give no Reckless Swings GCD,
  no spell id, and no ranks for Improved Whirlwind's rows).

### Ledger

| Note line | Tier and evidence | Sim status before | Fix |
|---|---|---|---|
| Victory Rush tooltip fixed | notes | n/a (not in a DPS rotation; needs a killing blow) | none |
| Defensive Stance threat 30% → 80% | notes | n/a | tank/healer sessions (tanking) |
| Slam cooldown 15 sec (was 18) | notes; client 70753 still 18000 | wrong (18 sec) | `sim/warrior/slam.go`: 15 sec under Forever. Test `TestForeverWarriorSpellValues` |
| Arms: Improved Tactical Mastery 3 points, 5/10/15 Rage (was 5 points of 3) | notes; client 70753 still 3 a point (12295) | wrong (3 a point) | `sim/warrior/stances.go`: 5 a point, points capped at 3 (old 5 point strings keep 15). Test `TestForeverWarriorImprovedTacticalMastery`. Tree max rank: Phase B (A1's remap) |
| Arms: Two-Handed Weapon Specialization 2/4/6% (was 1/2/3%) | notes; client 70753 still 1% curve | wrong (1% a point) | `sim/warrior/talents.go`: 2% a point under Forever. Test `TestForeverWarriorTwoHandedSpecialization` |
| Arms: Bloodthrill 2 points, 10/20% (was 5 points of 4%) | notes; client 70753 still 4% a point | wrong | `talents.go`: 10% a point, capped at 2 (old 5 point strings get 20%). Test `TestForeverWarriorBloodthrillChance`. Tree: Phase B |
| Arms: Improved Slam takes 2.5/5 sec off Slam's cooldown (was 3/3) | notes; client 70753 still -3000 flat | wrong (3 sec at either rank) | `slam.go`: 2.5 sec a point. Tests in `TestForeverWarriorSpellValues` (15 / 12.5 / 10 sec) |
| Fury: Blood Craze and Gore Drinker don't generate threat | notes | n/a for DPS (Gore Drinker isn't simulated: healing only) | tank/healer sessions |
| Fury: Blood Craze won't activate when sitting | notes | n/a | none |
| Fury: Booming Voice also cuts Piercing Howl's cost 5-25% | notes | n/a (Piercing Howl isn't in a DPS rotation or the sim) | none |
| Fury: Enrage to row 3 (was 4) | notes | tree only | Phase B (tree) |
| Fury: new talent Reckless Swings (row 4, 1 point, 15 sec cd, 25% of base health, next 4 melee attacks +5 Rage each, 8 sec) | notes (no client spell yet) | missing | Phase B: needs the talent field. Plan: an off-GCD self buff with 4 charges, each landed white swing (main or off hand) adds 5 Rage and uses a charge; the health cost is modelled as nothing for DPS (brief), and the 5 Rage is flat, not doubled by a crit (standards 21 doubles the swing's own Rage). Both are assumptions to note |
| Fury: Gore Drinker to row 5, under Blood Craze | notes | tree only (no DPS) | Phase B (tree) |
| Fury: Death Wish no longer required for Flurry | notes | tree only | Phase B (tree) |
| Fury: Flurry 3 points, 10/20/30% for 3 swings (was 5 points of 5%) | notes; client 70753 still the 5 point tree | wrong | `talents.go` `makeFlurryAura`: 10% a point, capped at 3, under Forever. Tests `TestForeverWarriorFlurrySpeed`, `TestForeverWarriorDeepWoundsAndFlurry`. Tree: Phase B |
| Fury: new talent Improved Whirlwind (row 6, 2 points, -1/2 sec Whirlwind cd, +15/30% to the closest enemy) | notes (24431 in the client is the old row) | missing | Phase B: needs the talent field. Plan: cooldown 10 − points sec; +15% a point on the main and off hand hits on the first target only |
| Fury: Bloodthirst needs Death Wish | notes | tree only | Phase B (tree) |
| Protection: Defiance 7/13/20% | notes | n/a | tank/healer sessions (tanking) |
| Talents lose the low-rank proc penalty | notes | n/a (no warrior talent has one) | A2 (shared part) |
| Races: Dwarf Artisan's Precision, Human Ruthless 1% / Spirit 10%, Orc Savage removed and Blood Fury 2 min, Skyborne Wind-Blessed 2%, Tauren Endurance, Troll Berserking 15 sec, Undead Touch of the Grave 15 sec | notes (10-26 racial pass), client 70753 | landed (ledger 70753 section, `sim/core/racials_forever.go`) | none; Phase B race sweep |
| Bug fix: race base stats (Tauren warrior +5 Str over Human) | notes | A2 | A2; Phase B race sweep |
| Bug fix: Throw's 0.5 sec windup | notes | A2 | A2 (no warrior rotation throws) |
| Bug fix: roots from procs get diminishing returns (Improved Hamstring) | notes | n/a (bosses) | none |

### Notes-only values

Per the orchestrator's provisional source rule (round README, a32a20024): modelled from the notes under Forever,
citing the notes and "client 70753 unchanged". For the warrior, client 70753 carries none of the class's lines, talent
reworks included (no Trait rows for tree 1117, no Reckless Swings spell), and the 2026-11-03 hotfix has hunter lines only.

- Slam cooldown 15 sec (client 18000).
- Improved Slam 2.5 / 5 sec off Slam's cooldown (client -3000 flat).
- Two-Handed Weapon Specialization 2/4/6% (client 1/2/3%).
- Improved Tactical Mastery 3 points, 5/10/15 Rage (client 5 points of 3).
- Bloodthrill 2 points, 10/20% (client 5 points of 4%).
- Flurry 3 points, 10/20/30%, no Death Wish prerequisite (client 5 points of 5%).
- Reckless Swings (15 sec, next 4 swings +5 Rage, 8 sec; no client spell, placeholder id 1399001) and Improved
  Whirlwind (1 / 2 sec, +15 / 30% on the closest enemy). Tests `TestForeverWarriorRecklessSwings`,
  `TestForeverWarriorImprovedWhirlwind`.
- The tree moves: Enrage row 3, Gore Drinker row 5 under Blood Craze, Bloodthirst under Death Wish
  (`tools/forever_talents/notes/warrior.json`).

### DPS effect of Phase A (goldens)

`TestP1DPSWarrior` (Forever, talents `30305013-25053005152110501`: 5 points of Flurry, now read as 3 at 30%, and
Improved Tactical Mastery 5, now 15 Rage): Average-Default 1518.5 → 1562.9 (+2.9%). Tank goldens didn't move.

### Phase A notes

- The warrior tree was built from the notes in Phase B (see the summary). The rank caps in the code (`min(points, N)`)
  stay as a guard for old strings.
