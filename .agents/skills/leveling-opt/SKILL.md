---
name: leveling-opt
description: Optimize a leveling rotation for the leveling sim (sim/core/leveling.go) - tune a rotopt *_leveling template's knobs for mobs killed in a 20 min session, have an auditor agent read sample sessions for what the knobs can't catch, test and fold its findings into the template, and repeat. Use it when the user asks to optimize, tune, audit or improve a leveling rotation or APL, find the best mob level, or rerun the leveling optimizer after a sim change.
---

# Optimize a leveling rotation

The leveling sim scores a rotation by the mobs it kills in a 20 min solo session
(`docs/leveling_sim.md`). Two halves work together:

- **The optimizer** (`tools/rotopt` with a `*_leveling` template) tunes numbers: rest and
  chain-pull thresholds, pull spells, which totems. It's a coordinate descent over the knobs,
  with the random seed fixed so candidates see the same sessions.
- **The auditor** (an agent with `orchestration/auditor_leveling.md`) reads sample sessions
  and finds what the rules miss: decisions that are hard to write down as fixed code. Its
  findings become new rules or knobs in the template, and the optimizer tunes them next round.

Run everything from the repo root.

## 1. Set up

```sh
go build -tags=with_db -o /tmp/rotopt ./tools/rotopt
```

Pick the template and settings. Today: `shaman_enh_leveling` (`tools/rotopt/aplkit/shaman_enh_leveling.go`)
with `tools/rotopt/leveling/level30_enh_solo.json` (the Level 30 Enhancement preset's gear and
talents, solo: no party or raid buffs, no starting totems, no explosives or potions; level 29 to
31 mobs). For another level or class, make a settings file the same way and say so in the docs.

## 2. Optimize the knobs

```sh
/tmp/rotopt -template shaman_enh_leveling -settings tools/rotopt/leveling/level30_enh_solo.json \
  -iters 400 -confirm-iters 2000 -out /tmp/lvl -name pass1 > /tmp/lvl/pass1.log
```

About 2 s a candidate at 400 sessions on 4 cores; a pass is minutes. The log has every
candidate; the last lines have the score (mobs killed with its standard error, deaths, mobs a
pull, seconds in combat, resting, searching and dead, food, drinks, bandages) and the best
knobs. `-knobs name=value,...` starts from earlier results.

## 3. Audit

Start the auditor as a background agent (Agent tool, `general-purpose`), so it works while you
do other things:

> Read orchestration/auditor_leveling.md and follow it. Settings: <file>. Template: <name>,
> knobs: <the best knobs>. Rotation: <path to pass1.apl.json>, score <x (se)>. Already tried:
> <the "tried" list from orchestration/leveling_audit_log.md for this template>. Report your
> ranked findings.

While it runs you can sample sessions yourself (`-session -session-seed N`) or sweep the mob
level (section 6).

## 4. Check what it found

For every finding it says helped, rerun its A/B yourself at 4000 iterations, on `-seed 1` and
`-seed 1000001` (session i runs on seed + i, so nearby seeds repeat the same sessions). Keep it only if both agree and the gain is at least two standard errors of the
difference. Note down everything it tried, kept or not, in `orchestration/leveling_audit_log.md`
(template, date, idea, numbers, verdict), so later audits skip the dead ends.

## 5. Fold it into the template

Write each kept finding into the template as a rule, with a knob when its threshold or
on/off isn't obvious (0 for off keeps the old behaviour reachable). If it needs an APL value or
action the sim doesn't have, add it (`sim/core/apl_values_leveling.go`, the leveling session,
the UI's `apl_values.ts` / `apl_actions.ts`), with a test. Then go back to step 2 starting
from the last best knobs. Stop when an audit finds nothing worth half a mob, or after three
rounds, and tell the user where it ended.

## 6. The best mob level

With the final rotation, score XP an hour at each mob level offset (the session's mobs are
our level plus the offset, ±1):

```sh
for o in -3 -2 -1 0 1 2 3; do /tmp/rotopt -template shaman_enh_leveling -settings <file> \
  -knobs <best> -objective xp -mob-offset $o -eval -confirm-iters 2000 | head -1; done
```

Higher mobs give more XP each but take longer and hurt more, lower ones the reverse. If the
best offset isn't 0, rerun step 2 at that offset (`-mob-offset`), since the best knobs move
with it.

## 7. Record

- `docs/leveling_sim.md`: the results table (rotation, mobs killed, XP an hour, deaths, time
  split) and what the audit found, in the doc's plain style.
- Ship the final rotation JSON where the UI can load it once there's a leveling mode (for now
  `tools/rotopt/leveling/<template>_<level>.apl.json`).
- `orchestration/leveling_audit_log.md`: the tried list.
- Commit and push (the repo's usual way).
