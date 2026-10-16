# Leveling plan

The Leveling mode sims a character from one level to another, with the gear, talents and
rotation it has at each level, and scores the plan by the hours it takes. The plan is
`ui/core/leveling_plan.ts`, and the Leveling Plan tab edits it
(`ui/core/components/individual_sim_ui/leveling_plan_tab.tsx`).

The mode we had before, one grinding session at one level, is now called Grinding. It's the
same session (docs/leveling_sim.md). Its id is still `leveling` in the code, so the runs we
saved before stay in its history.

## The modes and the tabs

In the Leveling mode the Leveling Plan tab takes the place of the Gear, Talents, Rotation,
Batch, Results and Compare tabs, because those edit or show one character at one level. The
Settings tab stays, for the race, the professions and the world we grind in (the Encounter
part has the session's options). The mode comes from the encounter like the others:
`LevelingOptions.plan` is on in the Leveling mode, and the Go sim never reads it.

## The plan

- **Levels.** We sim each level from From level up to, but not including, To level (10 to 30
  by default, the beta's cap).
- **Gear.** The leveling gear plan (docs/leveling_gear.md), as before.
- **Talents.** A build is the order we put its points in, and at each level we have as many of
  them as the level gives (one a level from 10, plus the extra points Forever gives from 25).
  The talent trees edit the order: a click puts the next point in a talent, a right click
  takes its last point out, and the Order list moves a point earlier or later. A point the
  tree doesn't allow yet (its row isn't open, or the talent it needs isn't full) shows red.
- **Respecs.** A build after the first is a respec at its level. There we put all the points
  back by its own order. A run also sims the respec's levels with the build we had, on the
  same sessions, and says how many hours the respec saves. We weigh that against the price
  of the respec ourselves.
- **Rotations.** Each rotation is played from its level until the next one's. It's a preset or
  a copy of the Rotation tab's. A shaman's also picks the weapon imbue, because Windfury
  Weapon only comes at 30.
- **Solo.** We level without the Settings tab's buffs, debuffs and consumables, unless
  "Settings tab's buffs" is on. The Settings tab usually has a raid's, and Goblin Sapper
  Charges hurt a level 10 more than the mobs do.

## The run

Simulate runs one grinding session request per level, each with the main iterations, then the
respec comparisons. A level's hours are the XP to the next level over the XP an hour of its
session, with Classic 1.12's XP table (notes.md, Need to Verify). Quests, dungeons and travel
are left out, so the total is the time of a plan that only grinds. It's a fair way to compare
plans, not a forecast of a real leveling time.

The grind button on a row (and Grind in the Gear part) takes that level of the plan to the
Grinding mode, with its level, gear, talents, rotation and imbue, so we can look at its
session in the Results tab.

The last run stays in the browser's storage. The Leveling mode has no sim history yet.

## The enhancement shaman's plan

`DefaultLevelingPlan` in `ui/enhancement_shaman/presets.ts`: the Level 30 Leveling build in
the order the leveling stat weights were measured with (Thundering Strikes, Mental Dexterity,
Improved Ghost Wolf, Shamanistic Focus, Ancestral Knowledge, then Flurry and Stormstrike from
25), and the "Level N Leveling Plan" rotations, the leveling optimizer's best at levels 10,
15, 20, 25 and 30 on the plan's gear, with Rockbiter Weapon until 30.

A first run (2026-10-16, 300 sessions a level) takes about 19 hours from 10 to 30. XP an
hour goes from 19,000 at 10 to 31,000 at 25, and falls inside each five levels as the gear
and the rotation fall behind the level.

## Open

- Searching the talent order and the respecs, level by level.
- A sim history for the Leveling mode.
- Rotations searched at more levels, and for the levels past 30 once the beta opens them.
- Forever's XP table.
