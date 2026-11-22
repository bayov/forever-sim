# Real Phase 0 gear, step 2: item effects, stats and availability, 2026-11-22

Read first: `briefs/phase0_gear_1122.md` (the plan) and `phase0_gear_1122.md` (step 1's notes, with the dry run table).
DPS specs only (user, 2026-11-22). The synthetic gear stays the report's main gear.

Step 2 makes the Phase 0 search's inputs trustworthy, so step 3 can run every spec's search. Four workers split it by
item kind. Each one owns its slots for every DPS spec in `tools/rotopt/phase0_specs.json`:

| Part | Slots |
|---|---|
| weapons | main hand, off hand, two-hand, ranged, wands, relics (idols, librams, totems), thrown |
| jewelry | trinkets, rings, necks, cloaks |
| cloth-leather | head, shoulders, chest, wrists, hands, waist, legs, feet in cloth and leather |
| mail-plate | the same armor slots in mail and plate |

## What to do for your slots

1. **Find the candidates.** Run `DRY_RUN=1` and `RANK_ONLY=1` for each spec (step 1's notes say how). Ranking on
   stats alone gives each slot's top items. Effects can't lift an item past the top 10 of its slot unless its effect is
   big, so judge the rest by the effect's size.
2. **Item effects:** model the "Use:" and "Chance on hit:" effects of candidates that could change a pick. The dry run
   lists the unmodelled ones. Read the effect from client data first: wago.tools db2 ItemEffect, SpellEffect and
   SpellName for build 1.60.1.71037, or `assets/db_inputs`. Then `docs/data_sources.md`'s other sources. Don't go past
   Wowhead's bot check. Model in `sim/common` (or the class package for class items), with a test each, following how
   existing Forever item effects are written. Use the Forever value where it differs from Classic. Where an effect
   can't be modelled (defensive, out of combat), say so in a comment and leave it.
3. **Hidden and Classic stats:** for candidates with Classic or partial stats, find Forever's stats. Use the client data
   (ItemSparse), `assets/db_inputs/forever_seen_in_game.json` (the user's in-game record) or foreverchanges.pro. Where
   no source has them, mark the item doubtful in `phase0_content.json` so the search ranks it but says so.
4. **Availability:** give every candidate a `phase0_content.json` entry: its source (dungeon and boss, quest, crafted
   with profession and recipe source, reputation and level, world drop, PvP rank) and whether it's obtainable now, by
   the rules already in the file. Exalted reputation and 300g or more are infeasible (`-gear-no-pick`). Molten Core
   and other unopened raids are out. Do the same for enchants your slots take, in `phase0_enchants.json`.
5. **Check:** each spec's `RANK_ONLY=1` run for your slots finishes with no unlisted-item failure, and the effects you
   modelled show in `EFFECTS_ONLY=1`.

Don't run the full searches (step 3), and don't write gear presets.

## Shared files

`phase0_content.json`, `phase0_enchants.json`, `sim/common/item_effects.go` (or wherever effects go) and TASKS.md are
shared by all four workers. Commit small and often. Merge origin/master before each push. Resolve conflicts by keeping
both sides' entries. Never drop another worker's entries.

## Coordination

- The wowsims/forever comparison session (`session_01TXnjyQoHMJg2CjHPYTGUMX`) lands class fixes and may still re-sim
  and deploy the report. It doesn't touch item effects or gear. Don't publish or deploy the report.
- Master currently fails TestItemSwitches, TestWarriorWfGate and TestRaidForeverReportHashesStable on a clean checkout.
  That isn't yours. Your packages' other tests must pass.
- Item effect code may flag report builds stale. No preset wears these items, so don't record or re-sim. The orchestrator
  does that after the comparison session's deploy.

## Rules

- Never force-push. End every commit message with your session prompt's trailer lines.
- Keep model names out of files. Keep long runs in subagents or in the background.
- Leave bayovui, the leveling weights and the leveling rotations alone.
- Merge to master and push without a PR, then push to your own development branch too.
- User style in anything they read: concise, no em dash or semicolon, "180s" not "180 s".

## When done

Write `orchestration/round_notes_2026-11-11/phase0_step2_<part>.md` with:
- the candidates per slot and spec, briefly;
- each effect modelled, with its test;
- the stats you found and their source;
- the items left doubtful or unmodelled, and why;
- anything that needs a user call.

Push it, send one message to `@parent` with the master commit and a summary under 200 words, then stop.
