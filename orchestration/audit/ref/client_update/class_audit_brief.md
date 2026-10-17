<!-- The brief this auditor gave each per-class agent for the full 70009-70425 audit (2026-10-16).
Reuse it for a large build: change the build list, the in-flight PR list and the checkout sha.
The 'already done or in flight' section is a snapshot of 2026-10-16 and is stale. -->

# Game-update audit brief (per class or system)

You are checking that **every change announced in a WoW Forever beta update has landed in the sim** for one class or system (named in your prompt). Repo: $AUDIT_TREE (Go sim in sim/, TS UI in ui/, level 60 DPS report requests in tools/dps_report/req/*.json). The checkout is origin/master (6a87c9c). **Do not edit, commit or push anything in the repo.** Write your output only to the file named in your prompt.

S = $AUDIT_WORK (scratch). The text sources and diffs below are also committed under orchestration/audit/ref/client_update/ (sources/ and diffs/).

## Sources (already downloaded; Wowhead is blocked, 403)

- `S/src/devnotes.txt`: the Blizzard thread "WoW Forever Beta Development Notes" (https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696). Each `=====` header is one post. Post to build mapping:
  - post 1, 2026-09-24 → build **70009** (link .../2360696/1)
  - post 4, 2026-10-01 → build **70170** (link .../2360696/4). The 10-02 "Warrior Updates in Today's Beta Build" (https://us.forums.blizzard.com/en/wow/t/warrior-updates-in-todays-beta-build/2369360) is the same build; its lines are already in post 4.
  - post 5, 2026-10-08 → build **70291** (link .../2360696/5)
  - post 6, 2026-10-15 → build **70425** (link .../2360696/6)
  - 69893 is the build the beta opened on (baseline). 69977, 70058, 70124, 70205, 70235, 70245, 70334, 70338 and 70374 are client-only builds per foreverchanges (no spell changes). The DB2 diffs confirm or refute that.
- Class deep dives (design articles, 2026-09-30 to 10-15; they describe the design at that build): `S/src/news24301515.txt` Hunter+Druid (https://worldofwarcraft.blizzard.com/en-us/news/24301515), `news24301514.txt` Priest+Warrior (/news/24301514), `news24302097.txt` Mage+Shaman (/news/24302097), `news24310968.txt` Rogue+Warlock (/news/24310968), `news24302099.txt` Paladin (/news/24302099).
- `S/src/fc_patch-notes.txt`: foreverchanges.pro/patch-notes, every build newest first, class by class. Lines under "Found in the game" or marked "Not in Blizzard's notes" are **unannounced changes** they found in the client.
- Client DB2 diffs between consecutive builds: `S/diffs/<old>-<new>.md`, grouped by class (then Racial, Item, No class, lookup tables). The pairs are 69893-70009, 70009-70170, 70170-70291, 70291-70334, 70334-70338, 70338-70374, 70374-70425. Class tagging is heuristic (SpellClassOptions, SkillLineAbility, talent trees, triggered spells, same name), so also grep the "No class" section for your spells.
- Raw CSVs: `S/db2/<Table>.<build>.csv` for builds 69893 70009 70170 70291 70334 70338 70374 70425 (Spell, SpellName, SpellEffect, SpellAuraOptions, SpellCooldowns, SpellCategories, SpellEquippedItems, SpellLevels, SpellMisc, SpellPower, SpellClassOptions, SpellDuration, SpellCastTimes, SpellProcsPerMinute, TraitDefinition, TraitDefinitionEffectPoints, TraitNode*, ItemSparse, ItemEffect, SpellItemEnchantment, SkillLineAbility, ...). Read them with `python3 -I` and the csv module. Useful columns: SpellEffect.EffectBasePointsF (base points; damage = base, Variance gives the spread), EffectBonusCoefficient (spell power coefficient), BonusCoefficientFromAP, EffectRealPointsPerLevel, EffectAuraPeriod; SpellMisc.CastingTimeIndex → SpellCastTimes.Base, DurationIndex → SpellDuration; SpellCooldowns.RecoveryTime/CategoryRecoveryTime (ms); SpellAuraOptions.ProcChance/ProcCharges/CumulativeAura/ProcCategoryRecovery (ICD ms). Talent rank values come from TraitDefinitionEffectPoints.CurveID (the curve table is not downloaded; use the Spell's own EffectBasePointsF per rank spell, or SpellEffect rows of each rank's spell).
- Prior work in the repo: `orchestration/forever_class_changes.md` (an earlier class-change audit against client 70338), `orchestration/brief_<class>.md`, `forever-wiki/classes/<class>.md`, `notes.md` (Need to Verify list), `orchestration/spec_assumptions.md`.

## What is already done or in flight (check before calling anything missing)

- Merged to master: #50 shared 70425 port (racials: Eureka 8 s, Blood Fury 3 min, Touch of the Grave; Moonkin Aura / Leader of the Pack split; raid-wide shouts and totems), #51 wording, #52 druid 70425 port, #53 Orc Savage, #55 generic encounter raid hits plus Forever damage-taken Rage.
- In flight: paladin 70425 port, PR #54 (branch origin/claude/wow-forever-dps-presets-m9u15h-paladin). Class workers are porting 70425 for mage, shaman, priest, warlock and hunter on origin/claude/wow-forever-dps-presets-m9u15h-<class> (inspect with `git -C $AUDIT_TREE log origin/master..origin/claude/wow-forever-dps-presets-m9u15h-<class>` and `git show <branch>:<path>`). For a 70425 change in those classes: if the branch has it, status is **in flight (worker PR)**; if the branch lacks it, say so explicitly ("missed by worker").
- Warrior and rogue have no 70425 worker.

## What to produce

Write `S/ledger/<name>.md` with:

1. **Announced changes**, one table row per announced line in your scope (every build, 70009 through 70425, plus deep-dive statements that describe a mechanic), newest build first:
   `| ID | Source (link, date, build) | Change | L60 DPS? | Client check | Sim status | Code location / note |`
   - ID: a short code like WAR-70425-03.
   - L60 DPS?: `yes` (changes a level 60 DPS preset or a mechanic the sim models), `minor` (sim-modelled but no preset uses it), or `n/a` (healing, tanking/threat, PvP, leveling-only, UI, movement, quests). n/a rows may be terse, and several n/a lines of one build may share one row.
   - Client check: the DB2 table + row/spell ID with old → new values, and the number the sim should use. If the client does not hold it (server side), say "server side".
   - Sim status: one of `landed`, `missing`, `partial`, `wrong`, `not simulated on purpose`, `in flight (worker PR)`, `n/a`, `ambiguous (needs verify)`.
   - Code location: file:line on master (or branch:file) that implements it, or where the fix would go.
2. **Unannounced client changes**: the reverse check. Every change in the DB2 diffs for your class/system that the notes did not mention and that could matter at level 60 DPS (base points, coefficients, cooldowns, cast times, durations, proc chances, costs, talent values, new/removed spells or talents). Same columns. Skip pure tooltip text, visuals, labels, and NPC/quest spells. Also confirm or refute that the client-only builds changed nothing for you.
3. **Proposed fixes**: for each missing/partial/wrong row with DPS impact, the concrete change (file, function, old value → new value, which test to add), a rough DPS-impact guess, and which presets in tools/dps_report/req it affects.
4. **Need to Verify candidates**: where the client and the notes disagree, or the server decides. Give a one-paragraph item each in the style of notes.md.
5. A one-line count of rows by status.

Be accurate over fast: verify each "landed" by reading the code (value and mechanic), not by the existence of a function name. Quote the exact sim value you found. Keep the file under about 600 lines.
