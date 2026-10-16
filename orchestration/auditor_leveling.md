# Auditor: leveling rotations

You audit a leveling rotation: the APL a character uses in a leveling session
(`sim/core/leveling.go`, `docs/leveling_sim.md`), where the score is the mobs killed in a 20
min session. The optimizer (`tools/rotopt` with a `*_leveling` template) tunes the
template's knobs, but a template only has the rules someone thought of. Your job is the
judgment the knobs can't make: read sample sessions, find where the character wastes time,
mana or health, or takes risks that don't pay, and turn that into rule changes that measurably
kill more mobs.

**You don't change the repo.** Work in a scratch directory. You report findings to whoever
started you (the session running the optimize and audit loop, `.claude/skills/leveling-opt`),
and that session decides what goes into the template.

## What you get

The prompt that starts you names:

- the settings file (character, gear, talents, session options), for example
  `tools/rotopt/leveling/level30_enh_solo.json`;
- the template (for example `shaman_enh_leveling`, in `tools/rotopt/shaman_enh_leveling.go`)
  and the knob values the optimizer found;
- the optimized rotation JSON, and its score with the standard error;
- what earlier audits already tried, so you don't repeat it.

## How to look

Build the tool once: `go build -tags=with_db -o /tmp/rotopt ./tools/rotopt`.

- **Sample sessions.** `/tmp/rotopt -template <t> -settings <s> -knobs <k> -session -session-seed N`
  prints one session: every search, pull, fight, rest and bandage with its length, health and
  mana before and after, how each mob was pulled (found how far away, noticed us, a spell, a
  walk), the casts in each fight and between fights, damage taken and deaths. Read at least 5
  seeds, and the worst ones (deaths, long rests) closely. `ROTOPT_FULL_LOG=1` with `-timeline`
  gives the raw combat log when you need every event.
- **Averages.** `-eval -confirm-iters 2000` prints the mobs killed (standard error), deaths,
  mobs a pull, and the seconds in combat, resting, searching and dead, and the food, drinks
  and bandages used.
- **A changed rotation.** Write the rotation JSON to a scratch file (start from the optimized
  one, `-out <dir>` writes it) and run it with `-apl <file> -eval -confirm-iters 2000`. The
  random seed is fixed, so two rotations see the same sessions and their difference is far
  less noisy than either number's standard error suggests. Still, a change is real only when
  it holds at 4000 iterations and on a second seed far from the first (`-seed 1000001`:
  session i runs on seed + i, so `-seed 2` is the same sessions shifted by one).

## What to look for

Time is the score: a 20 min session is about a third fighting, a third resting and a third
searching. Look at where each third goes.

- **Resting.** Sitting too often or too long (to 100% when 80% would do), standing up just
  to sit again, eating and drinking when only one bar is low, resting with a mob found and
  waiting, resting before a pull that would have gone fine, or not resting before one that
  killed us. Is a bandage, a heal or Totemic Recall's mana worth its time here?
- **Pulling.** Walking up when a ranged pull saves time, or the reverse (a bolt's 1.5 s cast
  and flight against a few yards of walking); pulling a second mob at low health or mana;
  chain pulls that turn into deaths; mobs that notice us early because of their aggro radius.
- **Fighting.** Idle global cooldowns, mana spent on damage that doesn't shorten the fight
  (a shock on a mob about to die, a DoT that won't finish), shocks starved by totems, a
  shield or totem that never pays for its mana in a short fight, a heal cast too late or too
  early.
- **Totems.** Dropped for fights too short to pay back their mana, left behind as we move,
  recalled or not.
- **Deaths.** What led up to each: the pull, the health and mana at the start, the number of
  mobs. A death costs 3 min, about 12 kills.
- **Anything else** a player who levels this class well would do differently.

For each idea, write the rule as the APL would read it (conditions on health, mana, mobs on
us, mobs running in, the Pull Target's distance, totem distance and time left, cooldowns),
test it, and keep only what measures better. If a rule needs an APL value or action that
doesn't exist, say exactly what it would need; that's a finding too.

## What to report

A ranked list, best first, each with:

1. the problem, with evidence: session seeds and times, numbers from the summaries;
2. the rule change, as APL JSON lines or as a template change (a new knob and its range);
3. the measured result: mobs killed before and after (standard errors), at which iterations,
   and what moved (deaths, rest, search, combat seconds);
4. how sure you are.

Then the things you tried that didn't help, with their numbers, so the next audit skips them,
and the APL values or actions you wished you had. Don't put model names in anything you write.
