# Leveling sim: hand-off

Read this to pick up the leveling sim work. Start with `docs/leveling_sim.md`, the main
write-up: what the sim models, its defaults, the history of each step, and the optimizer
results. This page is the short version: where things stand, where the code is, how to run
it, and what's next.

## The goal

forever-sim is a wowsims Classic fork for WoW Forever. The leveling sim answers "how fast do
I level" rather than "how much DPS". A run is a 20 min solo grinding session, scored by mobs
killed (and XP an hour). The character searches for mobs, pulls one or more, fights, rests
(eats, drinks, bandages), dies now and then, and repeats. Everything the character chooses is
in its rotation (APL); the settings only describe the world. Enhancement Shaman at level 30
is the first class done end to end.

## Where it stands (2026-10-16, all on `master`)

- **Engine.** The session states are search, pull, combat, rest, bandage, dead and ready. The
  world is a 2D plane: a mob is found at a random bearing 10 to 40 yd away and notices us
  inside its aggro radius. We have face pulls, ranged pulls and chain pulls. Shaman totems
  have positions and reach (buff totems 30 yd, Searing 20 yd, Magma 8 yd, Fire Nova 10 yd),
  and Totemic Recall and Totemic Projection work. Out of combat health regen follows the
  Wrath per-Spirit tables, with the sitting bonus. Food, drink and bandages are modelled.
- **APL.** Leveling actions: Look For Mob, Walk To Pull Target, Walk To Totem (shaman), Rest
  (eat/drink until health/mana targets), Bandage. Leveling values: distance to a unit,
  mobs on us, mobs running in, mobs this pull, is moving, bandage ready, totem distance.
  The unit `PullTarget` is the mob we found and haven't pulled.
- **Settings.** `LevelingOptions` in `proto/common.proto`: session length, mob level offset
  and spread, mob type, armor multiplier, max pack size, search times, found distances,
  chain-search running share, death penalty, revive fraction. Eat Below, Drink Below,
  Bandages and Pack Size were removed (fields reserved): the rotation decides those now.
- **Fallbacks in the session**, which only keep it going:
  - A mob we found and haven't touched gets walked up to after 1 s idle out of combat
    (`levelingFacePullAfter`). It waits while we cast or walk, and stops once a spell is on
    the way.
  - Between pulls we search after 10 s even if the rotation keeps acting
    (`levelingReadyAtMost`).
  - A rotation with no Rest or Bandage line gets defaults in front (`levelingDefaultRests`):
    bandage below 50% health, eat and drink below 50% health or 40% mana, both while not
    moving.
- **Optimizer.** `tools/rotopt` with the `shaman_enh_leveling` template. Level 30 solo
  Enhancement scores **93.0 mobs a session** (about 53,800 XP an hour, 0.03 deaths, 2.3 mobs
  a pull), the same on two seeds. Pass and audit history: `orchestration/leveling_audit_log.md`.
  Mobs one level under us are best by about 1% in XP an hour.
- **UI.**
  - "Leveling" is an Encounter Type next to PvP, in the individual sims only.
    `ui/core/encounter.ts` has `getLeveling`/`setLeveling`/`isLeveling`; the inputs are in
    `ui/core/components/encounter_picker.ts`.
  - Results lead with mobs killed, XP an hour, deaths, mobs a pull and the time split
    (`ui/core/components/leveling_results.tsx`, `raid_sim_action.tsx`, `sim_history.tsx`,
    `compare_results.tsx`).
  - Stat weights are off in Leveling mode.
  - Enhancement Shaman has a "Leveling" preset group: build "Level 30 Leveling (solo)",
    rotation `ui/enhancement_shaman/apls/level30_leveling.apl.json`.
  - DPS and PvP modes must look and behave as before: keep every new UI piece behind
    `encounter.isLeveling()` or the presence of `encounterMetrics.leveling`.

## Where the code is

| Area | Files |
|---|---|
| Session engine | `sim/core/leveling.go` (states, mob pool, pulls, deaths, metrics, `prepareLevelingEncounter`) |
| 2D world | `sim/core/leveling_world.go` (`Vec2`, movement, chase, aggro radius, `pollPull`, `onCast`/`onHit`) |
| Rest, regen | `sim/core/leveling_rest.go`, `sim/core/leveling_tables.go` |
| APL | `sim/core/apl_values_leveling.go`, `sim/core/apl_action.go`, `proto/apl.proto`; UI `ui/core/components/individual_sim_ui/apl_values.ts`, `apl_actions.ts` |
| Shaman totems | `sim/shaman/totems.go`, `totem_reach.go`, `earth_totems.go` |
| Tests | `sim/core/leveling_test.go`, `leveling_world_test.go`, `sim/shaman/enhancement/leveling_*_test.go`, `tools/rotopt/leveling_test.go` |
| Optimizer | `tools/rotopt/leveling.go` (scoring, `-session` summary), `apl_leveling.go` (APL helpers), `shaman_enh_leveling.go` (template and knobs), `leveling/level30_enh_solo.json` (settings), `leveling/shaman_enh_leveling_30.apl.json` (best rotation) |
| Auditor | `orchestration/auditor_leveling.md` (brief), `orchestration/leveling_audit_log.md` (what was tried) |
| Loop skill | `.agents/skills/leveling-opt/SKILL.md` (`.claude/skills` links to it) |

## How to run

```sh
# Tests (sim/web fails to build on its own; ignore it)
go test --tags=with_db ./sim/... ./tools/...

# Protos after a .proto change (TS protos are gitignored, regenerate locally)
PATH=$PATH:$HOME/go/bin npx protoc -I=./proto --go_out=./sim/core ./proto/*.proto
for f in api test ui; do npx protoc --ts_opt generate_dependencies --ts_out ui/core/proto --proto_path proto proto/$f.proto; done

# UI type check and format
npx tsc --noEmit
npx prettier --check <changed files>

# Optimizer: score, tune, read one session
go build -tags=with_db -o /tmp/rotopt ./tools/rotopt
/tmp/rotopt -template shaman_enh_leveling -settings tools/rotopt/leveling/level30_enh_solo.json -eval -confirm-iters 4000
/tmp/rotopt -template shaman_enh_leveling -settings tools/rotopt/leveling/level30_enh_solo.json -iters 400 -confirm-iters 4000 -out /tmp/opt -name passN
/tmp/rotopt -template shaman_enh_leveling -settings tools/rotopt/leveling/level30_enh_solo.json -session -session-seed 7
```

Optimizer notes:
- `-objective xp` scores XP an hour, and `-mob-offset N` sets the mob level. `-knobs a=1,b=2`
  starts from given knobs; `-apl file` runs a hand-written rotation.
- Session i runs on seed + i. To check a gain on other sessions, use `-seed 1000001`, not
  `-seed 2`.
- Keep a change only if it holds at 4000 sessions on both seeds.
- A pass is about 30 to 60 min. Run it in the background with `nice`.

## Conventions

- Work on the branch you're given, merge `origin/master` in (don't rebase), and push to
  master only when the user asks. Other sessions push to master often, so fetch and merge
  just before pushing.
- When two sides both changed a `*.results` file, regenerate it from the `.results.tmp` the
  test writes, after checking that the change is expected.
- Commits end with the attribution lines the session gives you. No model names in commits,
  code or docs.
- Docs (`docs/leveling_sim.md`, the audit log) use plain, user-centric prose. Record each
  result with its numbers, and note what was tried and didn't help.
- A local `make` build may regenerate `ui/dps_report/index.html`. Restore it with
  `git checkout` before committing.
- `vite.build-workers.ts` finds `wasm_exec.js` in Go 1.24's `lib/wasm`, and falls back to
  the old `misc/wasm`.

## Things learned the hard way

- **Weapon imbues.** The game takes one shaman imbue on the main hand.
  `shaman.OneMainHandImbue` keeps the class options' imbue and drops a different one from
  the consumes. Results before that fix ran with Rockbiter and Windfury both on, which
  inflated them to about 100 mobs.
- **Execute phases** go by time over the whole encounter, so `prepareLevelingEncounter`
  turns them off, and PvP too. A per-mob execute (Execute, Hammer of Wrath on each mob's
  health) isn't modelled yet.
- **Leveling option values.** A value of 0 or less falls back to its default, so "0" can't
  be set for an option like the death penalty.
- **Retargeting.** When the last mob dies, our target moves to the found mob at its real
  distance. Before that fix, a melee swing pulled it from 40 yd.
- **Totem ranks.** Our own lower-rank totems give their own rank's buff
  (`StrengthOfEarthTotemRankAura`, `StoneskinTotemRankAura`).
- **Session pacing.** Rests are paced by mana, not health: drinking is slower than eating.
  The optimizer's gains came from cheaper mana (Mana Spring Totem, rank 1 Searing Totem,
  Fire Nova only with two mobs on us), not from casting less.

## What's next (recommended order)

1. **Check the Level 30 Enhancement session against the user's play logs** (roadmap step 6,
   "From the logs" in the docs). The search, chain-search, run-in and found-distance
   defaults are guesses, and they drive the best rest and chain-pull choices.
   - The logs are in `~/forever-logs` on the user's machine; the `log-audit` skill reads
     them.
   - Measure, for each kill: the time from the kill to the next pull (less eating and
     drinking), the run-in, the gaps between chain pulls, and pack sizes.
   - Drop gaps over about 60 s (AFK) and use the fast quartile as the default.
   - Then re-tune with the optimizer.
2. **Leveling stat weights and a Pawn export** (mobs or XP an hour per stat), and turn the
   UI's stat weights back on for Leveling mode.
3. **The talent route:** the talent to take at each level, scored by leveling speed.
4. **Other classes.**
   - Each needs a `*_leveling` rotopt template (rest, pull, fight rules with knobs), a solo
     settings file, a preset in its UI's "Leveling" group, and an optimize and audit round
     (the `leveling-opt` skill).
   - Warriors and rogues have no mana. Check the user's belief that they get a bigger
     sitting health bonus; it's in Need to Verify.
5. **Gear curves from level 10 to 60,** to sim a whole leveling journey.

Open small items (also in `TASKS.md`):
- **Shaman regen:** measure out-of-combat health ticks on the beta (two Spirit values under
  50 and one over, at two levels, standing and sitting).
- **Chain-pull bolt:** a bolt's flight time to a chain-pull target that isn't our current
  target uses our current target's distance.
- **Stoneskin Totem:** its client data also has a 20 yd effect; we use 30 yd for all of it.
- **The 1 s face-pull fallback** could become a setting, so a rotation can wait longer for a
  cooldown before pulling.
- **UI:**
  - The live readout during a leveling run shows DPS until the run finishes.
  - The Compare tab's "Biggest changes" list and distribution chart still use DPS.
  - Leveling options can't be set to 0.
- **The auditor's wish list,** from the audit log: the session's state (ready, searching,
  pulling) as an APL value, the Pull Target's aggro radius, the target's time to die, and
  how long a rest will take on each bar.
