# Pass: Phase 0 gear

Follow `orchestration/audit/PASS.md` around this doc. Keep it current when the method changes.

Since 2026-11-24 every DPS preset in the report wears its spec's Phase 0 list: real items obtainable before Forever's first raids open on 2026-12-09 (`orchestration/standards.md` 43). This pass checks that each list is legal and obtainable, that the search behind it is sound, and that the report sims the list it says it does. Tank and healer synthetic sets have their own passes (`tank_gear`, `healer_gear`). The synthetic DPS method this pass used until then (budget lines, `lin2.py`, `wepcheck.py`) is in this file's git history and in `tools/audit/gear/`, for synthetic pieces that come back.

The auditor never changes code, presets, lists or docs, and doesn't commit (the audit run commits its files). No PR comments. Don't run the gear search (`tools/rotopt/bis_phase0.sh`) except its dry run (`DRY_RUN=1`), and don't put model names in anything you write. Findings go to `orchestration/audit/findings/gear.md`.

## What is audited

- **The spec table:** `tools/rotopt/phase0_specs.json`, one row per DPS spec (23): the report build the search fights as (`req`), talents, start gear, weapon setup (`weapons`, `mhTypes`, `ohTypes`), and where the list and gear set go.
- **The lists:** `ui/<dir>/bis/level60_phase0_<spec>.bis.json` (or the row's `out`), with its `.effects.json`. Each slot ranks up to 25 items on the final set, in order of loss. `bis: true` marks the picks. `dps` is the final set's, `sim` the sim commit it ran on, `faction` and `race` the build's.
- **The gear sets:** `ui/<dir>/gear_sets/level60_phase0_<spec>.gear.json` (or `gearOut`), the picks with their enchants.
- **The content checks:** `tools/rotopt/phase0_content.json` (items: `avail` out, infeasible or doubtful, `faction`, `include`, `note`, `goldPerDps`) and `tools/rotopt/phase0_enchants.json` (enchants by spell ID, the same tags).
- **The pool rules:** `tools/rotopt/phase0.go` (raid-only items), `faction.go` (the faction check), `bis_phase0.sh` (the rules in its header comment, the gold line `MAX_GOLD_PER_DPS`, 30 by default).
- **The report's gear:** each DPS request's `gearSet` and equipment in `tools/dps_report/req/`, and the race overrides' `equipment` in `req/races/`.
- **The step notes:** `orchestration/round_notes_2026-11-11/phase0_step3_<group>.md` and `phase0_step4_<group>.md`, where each class records its picks, flags, the comparison with public lists and its calibration.

## The rules you check against (standards 42, 43, 7, 8a, 29, 40)

- **Pool:** dungeon drops, quest rewards, crafting, world drops, the auction house, reputation rewards up to Revered, PvP reputations included.
- **Out:** raid and world boss drops, Onyxia's head rewards, legendaries, crafts that need a raid material, Classic phase 3 and later, and what the content check marks "out" (known not open before the raids).
- **Ranked, never picked:** "infeasible" items (Exalted reputation) and items over the gold line. They may sit above the pick in a slot's ranking, with a negative loss. That is how a list shows them.
- **Unsure is in:** a doubtful tag, an unknown source or an unconfirmed zone is no reason to leave a good item out (42). It needs a note saying what is unsure. Don't raise it as a finding unless the source is known to be closed.
- **Faction:** no pick only the other faction can get, by the build's race. A race override of the other faction swaps those slots for its own faction's best (29).
- **Enchants:** the best open enchant per slot, from `phase0_enchants.json`. No raid formula, no other faction's quest enchant.
- **Hit:** no target. The set carries the hit the search finds best. The class re-checks its hit-dependent talents and rotations at that hit (4).
- **Set bonuses:** every bonus the pieces reach counts. A set whose bonus needs several pieces at once must have been searched from a start at the full set and at its bonus points, since a one-slot-at-a-time search can stop short of it (the "set trap", step 4).
- **Effects:** a use or proc counts where the sim models it. A pick with an unmodelled effect, hidden stats or a doubtful tag must be flagged in the class's step notes.
- **The report sims the list:** the request wears the list's gear set and sims to the list's `dps` within noise.
- **Enhancement** sims both imbue pairings on its list (40).

## A pass, step by step

```
source tools/audit/setup.sh                # AUDIT_WORK and the toolchain
export AUDIT_TREE=$PWD                     # repo checkout (master)
export AUDIT_WORK=$AUDIT_WORK/gear; mkdir -p $AUDIT_WORK
export AUDIT_GITREF=origin/master
git fetch -q origin
F='phase0|/bis/level60|gear_sets/level60|faction.go|forever_enchants|forever_items|tools/dps_report/req'
```

1. **Rulings.** Read new commits to `orchestration/standards.md` since `last.master` in `state/gear.json`:
   `git log --oneline <last>..origin/master -- orchestration/standards.md`. Gear rules are 4, 6, 7, 8a, 22, 29, 40, 42 and 43 for the DPS presets.
2. **Changes.** `git diff --name-only <last>..origin/master | grep -E "$F" | grep -v -E 'tank|heal|dps_report/data'`.
   **Quiet pass:** nothing filtered changed and there's no new gear ruling. Update `state/gear.json` and stop.
3. **Mechanical checks.** `python3 tools/audit/gear/phase0_check.py > $AUDIT_WORK/check.txt` and read the `!` lines and the last line. It flags:
   - a pick or enchant the content check doesn't list, marks out or infeasible, or puts over the gold line;
   - a pick or enchant of the other faction (the content check's `faction` against the list's, or a race override's race);
   - a gear set that isn't the list's picks and enchants;
   - an open item more than 1 DPS (`--noise`) above the slot's pick: the search stopped short there;
   - a request whose `gearSet` isn't a Phase 0 set or keeps a `_hit` step, or whose equipment isn't that set.
   Each is a finding unless the class's step notes explain it (a re-rank in progress, a noise confirm at 20000 iterations).
4. **Each changed list, by hand.** For every pick that is new since the last pass:
   - **Source:** read its content note. Check it against the item's sources in `assets/database/db.json` (or `python3 tools/audit/gear/tt.py ID` for the Wowhead tooltip). A source that is known to be closed before the raids (raid loot, a raid material, an NPC the user saw missing) is a finding. An unconfirmed one isn't (42).
   - **Faction:** an item our data puts on one side (its restriction, a one-faction reputation, quest or vendor) must match the build's race. The faction check catches most. A Classic quest reward with no quest side in our data can slip through: look at the quest.
   - **Effects and stats:** an item with a "Use:" or "Chance on hit:" line, or with Classic or partial stats, must be in the class's flags. `DRY_RUN=1 SPEC=<spec> tools/rotopt/bis_phase0.sh` lists the pool's unmodelled effects and hidden stats under `orchestration/round_notes_2026-11-11/phase0_pool/`.
   - **Weapons:** the setup matches the spec row (two-hand, dual-wield, one-hand or any; dagger-only hands where the rotation needs them), and the class can use the type.
5. **Set bonuses and the set trap.** Count each set's pieces in every gear set. For a set one or two pieces short of its next bonus, check the class's step 4 notes for a search from a start at that bonus. Check the bonus is wired in the sim (`sim/<class>/item_sets*.go`, `sim/common/`): a live bonus with no code is a finding.
6. **Hit.** For each build, add the gear's hit to its talents, race and buffs (the report's Character stats, or a sim's final stats). Note any build over its cap (8% yellow, 16% spells against a level 63 boss, more for dual wielders by sim) with a hit talent still taken. Check that the class's step 4 notes re-checked its hit talents at the Phase 0 set's hit.
7. **The list's dps against the report.** When the loop gave you a shared run (`$AUDIT_WORK/../runs/master/summary.tsv`), compare each build's 180s DPS with its spec list's `dps`. For a build that wears its spec's list with the list's talents, a gap over about 2 DPS (20k iterations) means the list or the request is stale. A list whose `sim` commit predates a sim change that moved its class (`git log <sim>..origin/master -- sim/<class> sim/core`) is stale too. Builds with other talents than the list (duty, encounter and other-talent presets) may differ: note, don't flag.
8. **Report.**
   - A break is a finding in `findings/gear.md` (format in `PASS.md`), with the spec, slot, item id, the rule (the standards item) and `needs: worker fix` (the class) or `orchestrator fix`.
   - A clean pass with something new (a new list, new picks, a ruling that touches gear) gets a short `reports/gear.md`: per spec the list's dps, faction, picks that changed, doubtful picks and unmodelled effects.
   - A quiet pass only updates `state/gear.json`.

## Settled; don't re-flag

- **Infeasible items above the pick.** Exalted rewards (Warsong Gulch, Frostwolf, the Defilers, the Watchers' Warden's Signets, Timbermaw's Stormcarver plans) are ranked but never picked, so their negative loss is by design. The Darkspear Raiders' gloves and shoulders need only Revered and are open. Read the tag in today's `phase0_content.json`, not the list's copy: a list keeps the tags of the day it ran (Arms still shows the Raider Gauntlets as infeasible), and `phase0_check.py` does the same.
- **Doubtful picks** with a note (standards 42): Darkmoon cards, Alterac Valley and Arathi Basin rewards, Dungeon Set 2 steps after the Valthalak summons, items whose quest or vendor we haven't found.
- **The search's fight:** 180s with uses off, plus each use on a 90s fight pressed at the pull (`USE_DUR`). The report sims 180s, so an on-use trinket's worth there differs from the list's. That isn't a finding.
- **Hit over the cap** on a set where the search found nothing better for that slot, as Fury two-hand with Precision (step 3 warrior notes).

## Gotchas

- Enchants in `phase0_enchants.json` are keyed by spell ID. A list's enchant entry has `id` (what the gear set stores), `game` and `spell`: look up the tag by `spell`.
- Rings and trinkets are one list each (`finger`, `trinket`) with two picks.
- Enhancement's files have no spec suffix: `level60_phase0.bis.json` and `level60_phase0.gear.json`.
- The gear set's slot order is head, neck, shoulder, back, chest, wrist, hands, waist, legs, feet, finger 1, finger 2, trinket 1, trinket 2, main hand, off hand, ranged.
- Race overrides that still wear synthetic `_hit` steps (ids 99xxxx) predate the Phase 0 requests. Flag them once the class's Phase 0 requests are on master.
- Cadence: the audit loop's gate starts a pass when a watched list, gear set, content check or request changes on master.
