---
name: compare-wowsims-forever
description: Compare our sim with wowsims/forever on one of our builds, matching the setup and tracing every DPS gap to a rule the two sims model differently. Use when the user asks to compare a class, spec, preset or mechanic with wowsims/forever, asks why theirs differs, or after they changed a class we sim.
---

# Compare with wowsims/forever

wowsims/forever is the other WoW Forever sim built on the game files. Run one of our requests
in both and explain the gap. The comparison raises questions but never settles them: a rule
becomes ours only on its evidence (game data, a log, a post).

**Never push to wowsims/forever, comment on it or open anything there.** Our tools in its
checkout stay out of its git (`.git/info/exclude`). If a finding looks like their bug, tell the
user; reporting it upstream is the user's call.

Tools are in `tools/wsf_compare/`. `$WSF` is their checkout (default `~/wowsims-forever`),
`$WSF_WORK` the work folder (default `~/wsf_compare`), binaries in `$WSF_WORK/bin`.

| File | What it does |
|---|---|
| `setup.sh` | Clones or updates wowsims/forever to latest master, adds our tools, dumps their builds and APLs, builds every binary. |
| `convert.py` | Our request to `req_ours.json`, `req_theirs.json` and `items.json` (our synthetic gear as their items). |
| `run.py` | Converts, runs both, prints `ours`, `theirs`, `theirs - ours` with its 2 sigma error. |
| `align.py` | Finds the bonus stats that make their sheet at the pull match ours; prints the `ALIGN` value. |
| `breakdown.py` | Per ability, for the player and each pet: DPS, casts, hits, crits, misses, dodges per iteration. |
| `layers.py` | Both sheets built up layer by layer (naked, talents, gear, consumables, buffs). |
| `ours_stats/` | Our sheet at the pull. Their counterpart and their runner `cmpsim` are in `theirs/` as `.txt`, so our build skips them. |

Reference files here:

- `known_differences.md`: differences already measured. Read your class's rows before step 3,
  so you can tell what is new. Update it when one is settled or a new one found.
- `converter.md`: everything the converter and `cmpsim` translate. Read it when a result looks
  off for a new class, rotation or synthetic item, and when adding a class.

## 1. Set up

```bash
tools/wsf_compare/setup.sh
```

About 3 min the first time. Rerun it whenever either sim changes, since every binary is built
from a commit. Needs protoc (`/root/tools/protoc/bin`). Note both commits it prints. If you
patched their checkout (step 4), save and revert the patch first, or the update stops:
`git -C $WSF diff > $WSF_WORK/<name>.patch && git -C $WSF checkout -- .`

## 2. Run the builds aligned

Use the DPS Report's requests (`tools/dps_report/req/<id>.json`), plus every variant that
differs in the part compared (pets, weapons, talents). Screen at 3,000 iterations; run the final
table at 20,000.

```bash
cd tools/wsf_compare
A=$(./align.py hunter__survival ~/wsf_compare/out/sv)
ALIGN=$A ./run.py hunter__survival ~/wsf_compare/out/sv --iters 3000
```

- **Align the sheets first.** `align.py` closes the stat gaps at the pull (Strength and Agility
  before Kings, then attack power; Intellect, Spirit and spell power for casters). Each gap it
  closes is a difference too: find its cause with `layers.py` (race bases, buff ranks, rating
  conversion) and list it.
- **The rotation is ours in both** (`--apl ours`, the default), so the run compares mechanics.
- **Ranged hunters** run standing still (`--turret`, drops our `Weave:` lines): their weave hides
  everything else. Melee specs run as they are.
- `--apl theirs:<name>` runs their rotation (`$WSF_WORK/theirs_builds`,
  `<class>_<spec>_<build>` or `apl_<class>_<spec>_<apl>`). On our gear it runs 72 to 174 DPS
  behind ours for reasons that aren't mechanics, so use it only for a question about their
  rotation.

## 3. Read the gap

1. **Player or pet.** `run.py` prints pets apart. A gap that is all pet is a pet rule.
2. **Which ability.** `breakdown.py <out dir>`. Look at casts first (rotation or resource), then
   hits and crits per cast (hit table), then damage per hit (scaling, coefficients, ranks). Our
   main hand, off hand and extra attacks are `OtherActionAttack` tags 1, 2 and 3; theirs are
   tags 14 and 15 and the extra attack's own spell (10610 Windfury, 13960 Sword
   Specialization).
3. **Which stat.** `layers.py <out dir>` when the sheets disagree before alignment.
4. **Read both codes** for that rule: ours in `sim/`, theirs in `$WSF/sim/` (spell data generated
   from the game files in `sim/core/spelldata` and `sim/<class>/spell_data_auto_gen.go`).
   `git -C $WSF log -S <name>` often gives their evidence.

## 4. Isolate each cause

Flip one rule on one side and rerun; the DPS it moves is that rule's worth. Use 3,000
iterations, 20,000 for anything under 3 DPS.

- **Ours:** edit `req_ours.json` (options, consumables, buffs), or flip the rule in a build of
  our sim from a scratch worktree, never the checkout's working tree.
- **Theirs:** edit `req_theirs.json`, or patch their checkout (uncommitted, reverted as in
  step 1).
- `run.py` takes other binaries: `OURS=<path>` (a scratch build, or
  `tools/dps_report/bin/wowsimcli_assume` with `DASH_ASSUME=<key>=<value>` for a report
  assumption) and `THEIRS=<name in $WSF_WORK/bin>` (e.g.
  `go build -tags=with_db -o $WSF_WORK/bin/theirs_cmpsim_x ./cmd/cmpsim` in `$WSF`). Measure an
  assumption against `wowsimcli_assume` with no knob, not against `ours_cli`.
- The causes should add up to the gap within the error. If they don't, there is another cause.

## 5. Sort the causes

Say which sim is right for each cause, on what evidence:

- **Our bug** (theirs has game data or a log, ours a Classic or Era carryover or a plain
  mistake): fix it with its test, then re-sim the report builds it moves
  (`regenerate-dps-report`).
- **Their bug:** note it for the user with the evidence. Their checkout doesn't change.
- **Open question** (no evidence, or it conflicts): name the report assumption
  (`orchestration/spec_assumptions.md`) or ruling ours follows, and what would settle it (a log,
  a sheet). Add it to `notes.md`'s To Verify list if missing. The user decides.
- **A choice, not a game fact** (engine timing, the generic boss): keep ours and say so.

Our side holds the user's rulings (`orchestration/`, `notes.md`, code comments citing them).
Don't propose reversing one without new evidence; name the ruling when theirs differs.

## 6. Report

- A table per build: ours, theirs, `theirs - ours` with its error, at 20,000 iterations, aligned.
- The causes with their DPS, sorted as in step 5.
- Both commits, the iterations, and any alignment used (`ALIGN`, `--their-mob-type`).
- Ask before acting on a judgement call. Our bugs with clear evidence follow `CLAUDE.md`: fix,
  merge, re-sim, deploy.

## Adding a class

`convert.py` maps every DPS Report class and stops on any other. For a new one, add:

1. Its options block (`class_opts`), from their `$WSF/proto/<class>.proto` against ours.
2. Its consumables in `C`: weapon oils, poisons, imbues, class-only items.
3. Its spell ids in `remap()`: the ranks theirs registers (`spell_data_auto_gen.go`) against
   what our rotation names, and any tag that differs.
4. Its cooldowns in `CD`, for `spellUsesLostByDelay`.
5. A talent tree whose order differs from ours (compare the tree files).
6. Synthetic gear effects in `cmpsim`, and the class's caster or melee stats in `align.py`.

Then run one build, check the sheets match (`layers.py`), check every ability casts in both
(0 casts in theirs is an id that didn't map), and add what bit to `converter.md`.
