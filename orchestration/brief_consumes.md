# Shared Forever consumables worker

You add Forever's new Phase 1 consumables to the sim's shared code, on branch `claude/wow-forever-dps-presets-m9u15h-consumes` (starts at base commit e188da0, which all nine class branches share). Nine class workers will merge your branch into theirs, so the change must be self-contained, small and correct. Commit and push only to that branch, never force-push, and don't open a PR (it rides along in the class PRs).

## What to add

The consumables auditor found that Forever's best consumables exist in the enums and the UI but aren't simulated (the UI marks them "not simulated"), so every class picks from a short list and casters are understated by about 2 to 3%. Values from client 1.60.1.70338 (verify each against wago.tools db2 SpellEffect for the item's spell before you code it; the client wins over this list):

- Elixir of the Owl: 25 Intellect + 2% spell crit? (check the client: the auditor read "25 Int + 2% crit"; find out whether the crit is spell, melee or both under Forever's unified crit) — Intellect elixir slot.
- Elixir of Sages: 25 Spirit + 2% crit — its slot per the enum.
- Elixir of the Grizzly: 25 Strength + 2% crit — Strength slot (replaces Juju Power).
- Elixir of Nature Power: +40 Nature spell power.
- Greater Mageblood Elixir: 20 mp5 (Forever's, Alchemy 175).
- Elixir of Greater Intellect: 25 Intellect (not in the IntellectElixir enum yet: add it at the end of the enum).
- Major Spellblasting Potion: +47 spell power for 30 sec (shares the potion cooldown; makeStatBuffPotionMCD already exists).
- Major Frenzy Potion: +80 attack power for 30 sec.
- Foods: Lobster Stew and Baked Salmon 28 spell power, Mightfish Steak 40 AP, Flank au Poivre 20 Agility; plus any Forever food that replaces Smoked Desert Dumplings' 20 Strength (Bear Bruscitti / Steaming Stag Steak per the auditor), if it's in the enum.
- Also check whether Elixir of the Owl and Elixir of Sages stack with each other (notes.md lists it as open): follow the client's elixir category flags (guardian/battle) if they're there; otherwise keep the existing assumption and say so.

## How

- Look at proto/common.proto (the enums: Potions, the elixir enums, Food), sim/core/consumes.go (how each enum value is applied, makeStatBuffPotionMCD) and the UI consumables picker (ui/core/components, search for "not simulated" / the list that marks them) so they stop showing as not simulated.
- Gate Forever-only items on the Forever ruleset only if the existing code does that for Forever items; follow the file's existing patterns.
- Only append to enums (never renumber).
- Add a small Go test that applies each new consumable under Forever and checks the stat or aura.
- Make sure nothing else changes: `go test --tags=with_db ./sim/...` must leave every class's golden results unchanged (none of the goldens use these items). `npx tsc --noEmit` passes.
- Commit in one or two clear commits (attribution lines from your session's system reminder), push, and message the orchestrator (session_01AGmMcVxGoubbDodiQDQwhz) the commit hash and the final list with values and sources. Setup notes for protoc etc. are in orchestration/auditors_common.md on the coordination branch.
