# Level 31+ rank tuning auditor

Read auditors_common.md first: the read-only rule, the worker table, how to run the sim, and how passes work all apply to you. Two updates to that file: every class now has a PR (Rogue #9, Druid #8), and the source order and Forever-only rules are in standards.md and auditor_forever.md.

## Why this audit exists (the user's words, paraphrased)

WoW Forever is in beta with a level cap of 30. Spell ranks up to level 30 have been actively balanced through the beta, and that balancing is still going on. Ranks learned at level 31 and above can't be played yet, so they may be stale: old entries in the client that Blizzard tuned long ago and hasn't updated for Forever's many changes, and that will get proper tuning at launch. The level 60 presets lean entirely on those ranks. Your job is to give a verdict, for each one, on whether it looks tuned for Forever or stale.

## Scope

For every spell the level 60 presets cast (the damage breakdown in each preset's result JSON shows which ones matter), plus raid buffs and debuffs from the shared raid, compare its ranks learned at level 31 and above against:

1. **Its own ranks up to level 30 in Forever.** Build the per-rank curve from client 1.60.1.70338 (wago.tools: SpellEffect base points and per-level scaling, coefficients / bonus coefficient, SpellPower cost, SpellCastTimes, SpellCooldowns, SpellLevels). Note where Forever changed the up-to-30 ranks compared with Era: the ratio, the coefficient, the cost and the cast time.
2. **The same spell's curve in Classic Era (1.15.x) and Season of Discovery.** Use the client data where you can get it, and Wowhead Classic / SoD pages or wowsims Era/SoD source as a fallback.

Then give a verdict per spell:
- **Tuned:** the 31+ ranks follow the same Forever change as the up-to-30 ranks. For example, if ranks 1–5 were raised 20% over Era, ranks 6+ are raised too, or the 31+ values differ from Era in a way consistent with Forever's design.
- **Stale:** the up-to-30 ranks were changed but the 31+ ranks are still Era's or SoD's exact values. Or the 31+ curve breaks the up-to-30 trend in a way that looks like an untouched old entry: a jump or flattening, a coefficient that ignores a Forever rework, a cost out of line.
- **Unchanged everywhere:** Forever didn't touch the spell at any rank, so there's nothing to suspect.
- **Unclear:** say what evidence would settle it.

Also cover:
- **New Forever spells and talents that only exist at 31+.** Examples: 31-point talents, Lava Burst, Frostfire Bolt, Bane of Doom, Penance, Light's Vigil, Demonic Pact, Shadow Word: Death. Compare them with the Forever up-to-30 power level of the same class, and with SoD's versions where they came from SoD runes.
- **Talents in the deeper tiers** (points 21+, which need level 31+): check that their trait curves look consistent with the shallower tiers' Forever changes.

For every stale or unclear verdict, estimate the DPS sensitivity. Run the preset with that spell's 31+ ranks re-scaled the way Forever changed its up-to-30 ranks, and report the DPS change and how it moves the spec in the cross-spec ranking. Do this with a local patch you never commit.

## Sources, best first

The Forever beta (client 70338, beta patch notes and blue posts, Blizzard's Forever articles and class deep dives, beta logs); then Season of Discovery; then Classic Era. Never vmangos. Never decide from a tooltip alone. If wago.tools refuses WebFetch, don't reach it any other way; note what you couldn't check.

## Deliverables

- **Report.** A report at `orchestration/rank_tuning.md` on the coordination branch (claude/wow-forever-dps-presets-m9u15h). This is the only file in the repo you may change. Commit and push it (fetch and rebase if the push races, never force), and update it after each pass. Contents:
  - one table per class: spell, ranks checked, verdict, evidence, DPS sensitivity;
  - a cross-class summary: which specs' standings rest on likely-stale data, and how much the ranking could move.
- **Messages.** Send the orchestrator (session_01AGmMcVxGoubbDodiQDQwhz) a summary after each pass. Send each class worker a short note on what affects its spec, so it can mention it in its PR's open-questions section. Workers must not change their sims on your verdicts alone; the orchestrator rules on any change.
- **Passes.** Work in passes, starting with the specs at the top and bottom of the ranking (Demonology, Destruction, Retribution, Fury; Rogue, Feral, Elemental, Smite), then the rest. Schedule the next pass about 45 minutes after each one. Keep going until the orchestrator tells you to stop.

Commit attribution lines for your report commits:

    Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
    Claude-Session: (your own session URL)

Don't put model names in the report.
