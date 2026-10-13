You are one of nine parallel workers. Each worker owns one class of WoW Forever in the forever-sim repo (github.com/bayov/forever-sim, a fork of wowsims/classic modelling World of Warcraft: Forever). The orchestrating session (session id: session_01AGmMcVxGoubbDodiQDQwhz) hands out the classes and reports to the user. You own **Mage**. You end with **one PR into `master`** for your class.

This is a long task (many hours). Work through it on your own. Don't stop to ask what to do next. Only stop for a real blocker you can't get around (then message the orchestrator, see "Reporting").

# Your branch

You start on branch `claude/wow-forever-dps-presets-m9u15h-mage`. It sits on commit `e188da0` ("Prepare for per-class level 60 preset work"), which all nine class branches share. Keep `e188da0` as an ancestor: never rebase, amend or squash it away, and never force-push. Commit and push only to `claude/wow-forever-dps-presets-m9u15h-mage` (`git push -u origin claude/wow-forever-dps-presets-m9u15h-mage`, retry network failures up to 4 times with 2/4/8/16 s backoff). Never push to `master` or another worker's branch.

# The goal (the user's words, condensed)

For every DPS spec of your class:

1. Make sure all its spells and talents are coded correctly and are up to date with the latest WoW Forever changes. Go over wago.tools (client DB2 data), wowhead's Forever database, foreverchanges.pro, the official patch notes / development notes, and other information about the spec.
2. Synthesize level 60 Phase 1 gear for the spec, like we did for the enhancement shaman, with proper item level budgeting and stat selection for that spec. Mana users spend some of the budget (about 10%) on Spirit.
3. With that synthetic gear held fixed, optimize talents and rotation.
4. AFTERWARDS, simulate every race the class can be, to find the best race overall, the best Horde race and the best Alliance race.
5. Build the spec's default **Level 60** preset: the best race, the synthetic gear, the best talents and rotation, and the rest of the settings like the enhancement shaman's Level 60 preset (party and raid buffs and debuffs that fit the spec, fitting consumes, professions, the Level 60 encounter, etc.).
6. If several talent builds for a spec land within 10% DPS of each other, ship more than one talent preset (and a matching preset build), but only for meaningfully different builds (enhancement going sub Restoration or sub Elemental, say). Never for a couple of minor points moved around.

# Read first

- `forever-wiki/sim-status.md` (what the sim models and assumes; your class's bullets under Talents; Comparing races; the new `## Level 60 presets` section), `forever-wiki/classes/<your class>.md`, `forever-wiki/talent-system.md`, `forever-wiki/stats-and-gear.md`, `forever-wiki/racials.md` (race and class combinations, which races are Alliance or Horde, incl. the two Skyborne races), `forever-wiki/references.md`.
- `notes.md` (open questions, "Need to Verify", incl. your class's new section), `TASKS.md`.
- `tools/forever_talents/README.md` (how the talent trees are made, `import_client.py` and `import_talents.py`).
- The enhancement shaman as the model to copy: `ui/enhancement_shaman/presets.ts` (the Level 60 parts: `GearLevel60Phase1Synthetic`, `TalentsLevel60*`, `DefaultConsumes`, `DefaultRaidBuffs`, `DefaultIndividualBuffs`, `DefaultDebuffs`, `OtherDefaults`, `EncounterLevel60`, `PresetBuildLevel60*`), `ui/enhancement_shaman/sim.ts` / `index.ts` (how builds and defaults are wired in), `ui/enhancement_shaman/gear_sets/level60_phase1_synthetic.gear.json`, and `tools/database/forever_synthetic_items.go` (the synthetic item method, read its whole comment).
- `shaman_audit.md` as an example of how carefully a class was audited (sources, numbers, what changed and by how much).
- `tools/rotopt/` (the optimizer: `main.go` flags, `talents.go`, `search.go`, `shaman_enh.go`, `rogue_*.go` and `paladin_ret.go` as template examples, `*_presets.sh`) and `tools/rotopt/mksettings/` (builds the RaidSimRequest JSON; each spec registers itself in its own `spec_<name>.go` via `init()`, see `specs.go`).

# Setup (the container doesn't have protoc)

```sh
apt-get install -y protobuf-compiler || (apt-get update && apt-get install -y protobuf-compiler)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
export PATH=$PATH:$HOME/go/bin
make sim/core/proto/api.pb.go           # Go protos (re-run after any proto/*.proto change)
npm ci                                   # for the TS side: `make ui/core/proto/api.ts` (or `make proto`) and `npx tsc --noEmit`
go test --tags=with_db ./sim/<your class>/...
```

Regenerate the item DB after changing items: `go run tools/database/gen_db/*.go -outDir=./assets -gen=db` (offline, deterministic, ~20 s). The generated `.pb.go` / `ui/core/proto/*.ts` files are gitignored, don't commit them.

The container has 4 cores. Sims are CPU bound, so don't run several heavy searches at once; run long ones in the background and keep working.

# The user's standing rules

- Sources, best first: the Forever beta / Forever client data, then Season of Discovery (the Forever client reuses a lot of its code), then Anniversary and 2019 Classic. Avoid citing vmangos (private server). https://wago.tools/db2 has the latest Forever client data (builds 70291 and up; check for the newest build). foreverchanges.pro tracks changes between builds and has a talent calculator. wowhead's Forever database (https://www.wowhead.com/forever/...) has spells and items; plain curl gets a 403 there, so try WebFetch or WebSearch, or other mirrors.
- Don't take tooltips at face value; the beta's tooltips can be wrong. Prefer the client's spell effect data (wago DB2 tables like SpellEffect, SpellMisc, SpellCooldowns, SpellPower, SpellDuration, SpellClassOptions, Talent, TalentTab) over tooltip text. When the only source is a tooltip, say so in the code comment and in notes.md.
- Where nothing settles a mechanic, keep the sim's current assumption (or pick the most plausible one), mark it in code with a comment, and add it under your class in `notes.md` "# Need to Verify". Say what would settle it.
- Write code that reads like the surrounding code (comment density, naming, idiom). Comments and docs in this repo are plain, short sentences with the numbers and sources; match that.
- Don't put any model name or model id in commits, code, comments or the PR.

# Step by step

## 1. Research and audit

For each DPS spec of your class, list every spell and talent that can move DPS (incl. buffs the class gives itself, procs, set bonuses, the racials it interacts with). For each one compare the sim (`sim/<class>/...`) against the newest Forever data: damage per rank at 60, coefficients, cost, cast time, cooldown, duration, proc rates, talent per-rank values, prerequisites and rows. Look for every Forever change to the class since the client build the wiki was written from (1.60.1.69876 / 70009): foreverchanges.pro diffs, development notes, patch notes, wago builds. Fix what's wrong, implement what's missing (talents the picker marks as not simulated that do move DPS, new Forever spells), and add focused tests where a mechanic is subtle. Keep a short audit log of what you checked and what you changed (DPS before and after where it moved); it goes into the PR body and the wiki.

If your class's talent tree is still the BlizzCon transcription (`tools/forever_talents/data/<class>.json` has `ranksSource` observed / classic-prior / extrapolated / manual rather than `client`), convert it: first check whether a newer client tree exists (foreverchanges.pro's talent calculator, hyjal.cc, wago's Talent tables) and update `tools/forever_talents/client/<class>.json` if so, then `tools/forever_talents/import_client.py <class>`, `tools/forever_talents/import_talents.py <class> --write`, which regenerates both `ui/core/talents/trees/<class>.json` and the talents message in `proto/<class>.proto` (they must stay in the same order). Then go through the sim code talent by talent against the real ranks (rank 2+ guesses are usually wrong somewhere) and fix every saved talent string in presets, tests and docs that the new order breaks. Update the "not simulated" talent list.

## 2. Synthetic gear

Put your class's items in a new file `tools/database/forever_synthetic_items_mage.go` that appends to `ForeverSyntheticItems` from an `init()` function, with ids from your class's block (996000-996999). Don't edit `forever_synthetic_items.go` itself. Follow the method in that file's comment:

- One set per DPS spec (share a set between specs only when their stat priorities really are the same, and say why). Name them like "<Spec> Synthetic <Slot> (Phase 1)" with a `SetName` for the pieces that form a set.
- Each piece takes the place of an Era Phase 1 BiS item the spec would wear (Molten Core, Onyxia, the pre-raid dungeon sets, world drops, crafted gear) and gets the full epic budget of its slot at the Phase 1 item level (Molten Core tier at 66, or the Era Phase 1 BiS item's level where that is higher). Use the Classic item budget formula with slot modifiers, spell power at Forever's tuned price (0.55 of a Strength point), hit at 10 rating per 1% and crit at 14 rating per 1% (Forever's ratios). Write down the budget math in the file comment so a reader can check it.
- Pick a stat mix per slot that fits the spec, not the same mix on every piece. Mana users put about 10% of the set's budget on Spirit (the Forever devs want Spirit to matter again); rage and energy users take none. Give the set the hit the Era Phase 1 BiS for that spec carries (melee about 9 to 10% for a +3 boss, casters what their Era BiS had, remembering hit rating is unified under Forever), and enough to matter but not every stat on every piece.
- Weapons keep the damage range and speed of the real Era Phase 1 weapon they replace. Armor values follow the slot and armor type.
- If Forever has a crafted or tier set for the spec, give the right pieces its set bonuses the way the shaman's pieces carry The Spiritcaller's Rage.
- Add `ui/<spec>/gear_sets/level60_phase1_synthetic.gear.json` with fitting Phase 1 enchants, regenerate the DB, and make the gear preset in `presets.ts` with a comment like the shaman's.

The gear stays fixed from here on (don't gear-search it into real items for the preset).

## 3. Settings, talents and rotation

- Add `tools/rotopt/mksettings/spec_<name>.go` for each spec (register via `init()`; don't edit the other spec files). Set the spec's raid: start from the shared level 60 raid in `main.go` (it already has both blessings, Arcane Brilliance, every debuff) and change what the spec would really get: Blessing of Wisdom vs Might, the totems a shaman in its group would drop, caster debuffs (Curse of Shadow / Elements, Shadow Weaving, Improved Scorch, Winter's Chill) for casters, Sunder / Faerie Fire / Curse of Recklessness etc. for melee. Pick its level 60 consumes and professions the same way the enhancement shaman did (try the candidates in each slot, keep what helps). Default encounter: the shared Level 60 boss (level 63, 3731 armor, Humanoid), 180 sec like the enhancement preset.
- Write a rotopt template per spec in its own file `tools/rotopt/<class>_<spec>.go` and register it from `init()` (`templates["..."] = ...`), not by editing the map in `template.go`. Make the rotation's real decisions knobs (thresholds, priorities, cooldown alignment, mana floors).
- Search talents with the gear fixed. The talent search is a hill climb, so seed it from every plausible archetype (deep in each tree, the hybrid splits players actually use) and compare where they end up. Alternate talent search and knob search until neither moves. Use enough iterations for the differences you're claiming (rotopt reports noise; confirm finalists at 20,000+ iterations), and check 120 / 180 / 300 sec.
- Keep each meaningfully different build within 10% of the best as its own talent preset (and its own rotation if it needs one).

## 4. Races

AFTER talents and rotation are settled, sim every race your class can be under Forever (see racials.md; include both Skyborne races where allowed) on the final build with the same seed, at 20,000+ iterations (standard error well under 1 DPS), at 180 sec and also 120 and 300 sec. Report the best race overall, the best Horde race and the best Alliance race, with the full table. Weapon-type racials (Human swords and maces, Orc axes, Dwarf guns/maces, Troll bows, etc.) depend on the synthetic weapon's type; say how the weapon type you chose affects the ranking, and why you chose it (the Era Phase 1 BiS weapon type), rather than changing the gear to favour a race.

## 5. Presets

In `ui/<spec>/presets.ts` (and `sim.ts` / `index.ts` as needed, like the enhancement shaman): a `Level 60` group with the synthetic gear preset, the talent preset(s), the rotation preset(s), `EncounterLevel60`, consumes, raid buffs, individual buffs, debuffs, professions, and `PresetBuildLevel60` (plus one build per extra talent preset), using the best overall race. Make the Level 60 build the spec's default (gear, talents, rotation, race, consumes, buffs) the way enhancement does. Keep the spec's existing presets working; fix ones your talent changes broke. Put comments above each preset with the numbers (DPS at 120 / 180 / 300 sec, how it was searched, what lost and by how much), in the enhancement presets' style.

## 6. Tests and checks

- `go test --tags=with_db ./sim/<your class>/...` passes. When your change intentionally moves a golden, regenerate it (`go test` writes `*.results.tmp`; copy over the `.results` for your class only) and say in the commit why it moved.
- Keep edits to shared code (`sim/core`, `sim/common`, `tools/rotopt/*.go` outside your template, `proto/common.proto`, `ui/core`) to what your class truly needs, because eight other PRs touch the same repo. If a shared change moves other classes' goldens, regenerate those too and call it out in the PR.
- `npx tsc --noEmit` passes, and run eslint / prettier on the TS files you changed (`npx eslint <files>`, `npx prettier --check <files>`).
- Try `make dist/classic/.dirstamp` once near the end if it fits in the disk and time budget (it's what CI builds), at least confirm the UI for your spec compiles.

## 7. Docs

- `forever-wiki/sim-status.md`: fill in only your class's subsection under `## Level 60 presets` (gear, talents, rotation, the race table, the numbers) and update your class's existing bullets (Talents, Comparing races) where they're now wrong. Don't reflow other classes' text.
- `forever-wiki/classes/<your class>.md`: the Forever changes you found.
- `notes.md`: open assumptions under your class's section in "# Need to Verify" (and `## At 60` if they need level 60).

## 8. Commit, push and open the PR

Commit in topic-sized commits with clear messages, ending each message with the attribution lines your own session's system reminder gives. The DB regen (`assets/database/*`) goes in the same commit as the items that caused it. Push to `claude/wow-forever-dps-presets-m9u15h-mage`.

When everything is done, open a PR from `claude/wow-forever-dps-presets-m9u15h-mage` into `master` with the GitHub MCP tools (load them with ToolSearch, e.g. `select:mcp__github__create_pull_request`). Title: "Mage: Forever audit and level 60 presets". Body: what you audited and changed (with sources), the synthetic gear (budget and stat mix), the talent and rotation results, the race table with best overall / Horde / Alliance, the presets added, and the assumptions left open. Add a note that all nine class PRs share base commit e188da0, and that `assets/database/*` conflicts between them are resolved by re-running gen_db after merging. End the body with the attribution lines your session's system reminder gives for PRs. If you can't create the PR yourself, push the branch and tell the orchestrator.

# Reporting

Send short status messages to the orchestrator with `send_message` (claude-code-remote MCP server; load it with ToolSearch `select:mcp__claude-code-remote__send_message` if needed) to session `session_01AGmMcVxGoubbDodiQDQwhz`, at these points: after research (what's out of date), after talents and spells are fixed, after the synthetic gear, after talent/rotation optimization (best builds and DPS), after the race sweep (overall / Horde / Alliance best), and when the PR is open (with its URL). Two to six lines each, numbers included. Message it right away if you hit a blocker.

# Class notes

DPS specs: Arcane, Fire and Frost, all in `ui/mage`. The talent tree is still the BlizzCon transcription, so convert it to the client tree (step 1) and check every talent and spell (Fireball, Frostbolt, Arcane Missiles, Scorch / Improved Scorch, Ignite, Combustion, Arcane Power, Presence of Mind, Winter's Chill, Icy Veins-like Forever additions, etc.). Mana user: Spirit about 10%. Gear: cloth, spell power (school spell power only if the set is school specific), Intellect, Spirit, spell hit, spell crit. Fire and Frost may share one set if their stat priorities really match; say why. Races: every race a mage can be under Forever (Gnome, Human, Troll Berserking, Undead, Skyborne, etc.).

# Addendum (from the user, 2026-10-12): every spec, and a rotation audit

1. Cover every DPS spec of the class, not just the best one. If a spec plays very differently (different core abilities, resource loop, weapon style, school), it is its own spec: give it its own synthetic gear where its stats differ, its own talent and rotation search, and its own Level 60 talent preset, rotation preset and preset build, even when it is more than 10% behind the class's best spec. Consider more specs than the obvious ones; the per-class list below is a floor, not a ceiling. The 10% rule still applies within a spec: keep meaningfully different variants of the same spec within 10% of that spec's best (twist vs shock Retribution, Enhancement sub Elemental vs sub Restoration).
2. Audit each spec's rotation by hand. For each spec's final build, look at a few iterations with rotopt -timeline (and -breakdown, cast counts, uptimes): every spell is cast when and how it's meant to be, DoTs and debuffs stay up, cooldowns and procs are used, nothing is cast that shouldn't be, no idle GCDs or resource capping that the rotation could avoid, channels and cast times behave. Fix what's wrong and re-run the search if the fix moves the result. Write a short audit note per spec (what you checked and found) in the PR body.

Specs to cover for your class (at least): Arcane (Arcane Blast / Missile Barrage), Fire (Fireball / Heating Up / Pyroblast / Combustion), Frost (Frostbolt / Fingers of Frost / Ice Lance); consider Frostfire Bolt as its own spec.
