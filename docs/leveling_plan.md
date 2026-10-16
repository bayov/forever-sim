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
- **Gear.** The leveling gear plan (docs/leveling_gear.md), as before. The Gear part shows the selected level's gear in the Gear tab's gear pane. A click
  on a slot opens the Gear tab's item list with the items that level can wear, sorted by the
  stat weights of that level, and the item we pick becomes one of ours from that level on.
- **Talents.** A build is the order we put its points in, and at each level we have as many of
  them as the level gives (one a level from 10, plus any extra points from 25, none by
  default).
  The talent trees edit the order: a click puts the next point in a talent, and a right click
  takes its last point out. Each talent in the trees shows the level of its first point. The
  Order list has a row for each talent we put points in one after the other, with a pip for
  each rank, and the arrows move the row earlier or later. A point the tree doesn't allow yet
  (its row isn't open, or the talent it needs isn't full) shows red. Points after the plan's
  last level are dimmed, because we don't sim them.
- **Respecs.** A build after the first is a respec at its level. There we put all the points
  back by its own order. A run also sims the respec's levels with the build we had, on the
  same sessions, and says how many hours the respec saves. We weigh that against the price
  of the respec ourselves.
- **Rotations.** The plan has its own rotations (`LevelingPlanSettings.rotation_library`). They
  start as copies of the spec's presets, and we add more from the spec's presets or the
  Rotation tab, rename them and change them in the editor. A rotation with the name of a
  spec's preset that we changed is marked "edited", and "Spec's version" undoes our changes.
  The level ranges on top pick one of them for each range of levels, from the range's level
  until the next range's. A shaman's range also picks the weapon imbue, because Windfury
  Weapon only comes at 30. Plans saved before the library held each range's rotation in the
  range, and we move those into the library when we load them.
- **The plan's character.** The rotation editor's spell lists come from the character's level
  and talents, and the gear pane's item list from its level and stat weights. So the Gear and
  Rotations parts each have a character of their own (`LevelingPlanCharacter`), made from the
  request of a level as the run sims it. The character of the settings never changes, so a
  reload or another mode always finds it as it was.
- **Solo.** We level without the Settings tab's buffs, debuffs and consumables, unless
  "Settings tab's buffs" is on. The Settings tab usually has a raid's, and Goblin Sapper
  Charges hurt a level 10 more than the mobs do.

## The run

Simulate runs one grinding session request per level, each with the main iterations, then the
respec comparisons. A level's hours are the XP to the next level over the XP an hour of its
session, with Classic 1.12's XP table (notes.md, Need to Verify). Quests, dungeons and travel
are left out, so the total is the time of a plan that only grinds. It's a fair way to compare
plans, not a forecast of a real leveling time.

The Overview has the run's totals and a row per level, with the talents and the items we get
there, the rotation we play, and bars for the XP an hour and the hours. The grind button on a
row (and Grind in the Gear part) takes that level of the plan to the
Grinding mode, with its level, gear, talents, rotation and imbue, so we can look at its
session in the Results tab.

The last run stays in the browser's storage. The Leveling mode has no sim history yet.

## The enhancement shaman's plan

`DefaultLevelingPlan` in `ui/enhancement_shaman/presets.ts`: the Level 30 Leveling build in
the order the leveling stat weights were measured with (Thundering Strikes, Mental Dexterity,
Improved Ghost Wolf, Shamanistic Focus, Ancestral Knowledge, then Stormstrike at 25 and
Flurry), and the "Level N Leveling Plan" rotations, the leveling optimizer's best at levels 10,
15, 20, 25 and 30 on the plan's gear, with Rockbiter Weapon until 30.

A first run (2026-10-16, 300 sessions a level) takes about 19 hours from 10 to 30. XP an
hour goes from 19,000 at 10 to 31,000 at 25, and falls inside each five levels as the gear
and the rotation fall behind the level.

## Open

- Searching the talent order and the respecs, level by level.
- A sim history for the Leveling mode.
- Rotations searched at more levels, and for the levels past 30 once the beta opens them.
- Forever's XP table.
