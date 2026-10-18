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

- **Levels.** We sim each level from From level up to, but not including, To level (10 to 60
  by default). We grind up to 59, and 60 is where we stop. The gear plan, the XP table and the
  mob table go to 60, but the stat weights and the rotations stop at 30 for now (Open).
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
- **Ranks.** A rotation casts the highest rank we know of a spell when it names the spell's
  top rank, which is the "Highest rank" pick in the editor's rank menu. A rank picked by its
  number stays that rank, to downrank on purpose. So a range's rotation learns the new ranks
  as we level through the range: the Level 10 rotation casts Healing Wave rank 3 from 12 and
  Earth Shock rank 3 from 14. The spec's rotations name top ranks. Plans saved before
  2026-10-17 named the ranks of their level, and the tab moves them over once: each
  rotation's spells at the highest rank of its first range's level name the top rank.
- **The plan's character.** The rotation editor's spell lists come from the character's level
  and talents, and the gear pane's item list from its level and stat weights. So the Gear and
  Rotations parts each have a character of their own (`LevelingPlanCharacter`), made from the
  request of a level as the run sims it. The character of the settings never changes, so a
  reload or another mode always finds it as it was.
- **Solo.** We level without the Settings tab's buffs, debuffs and consumables, unless
  "Settings tab's buffs" is on. The Settings tab usually has a raid's, and Goblin Sapper
  Charges hurt a level 10 more than the mobs do.

## Talent search

The Talents part has a Talent search. It finds the order
of talent points that levels us the fastest over all the plan's levels, with the plan's gear
and rotations, and the best build to respec into at each respec level (20, 30, 40 and 50 by
default, the ones inside the plan count). Nothing in the plan changes until we click "Use this
route" or "Add this respec to the plan". We decide on a respec by the hours it saves against
its gold.

- **It runs in Go.** The search is `tools/levelingplan`, natively on all the cores, because the
  browser's wasm sims were too slow for a plan from 10 to 60. Search sends the plan to its local
  server, which we start with `go run --tags=with_db ./tools/levelingplan -serve localhost:3001`.
  It answers with progress lines and then the result, and Stop ends the search. The server makes
  each level's request the way the tab does, and the gear plan the way the Gear part does
  (`tools/levelingplan/gear.go`, a port of `ui/core/leveling_gear.ts`). The UI sends its own
  gear at each level too, and the search stops with an error when the two don't agree, so a
  change to one gear plan needs the same change in the other. The route search is a port of the
  browser's, and `route_test.go` checks it finds the same routes with the same sims on a made up
  problem. A 10 to 30 search takes 17s on 16 cores, where the browser took 3.5 minutes.
- **Whole routes, never one level at a time.** A route is the order of all our points, and its
  score is the hours of all the levels added up. Taking the best point at each level would be
  a greedy search, and it can be slower overall: the point that does the most at level 10 can
  keep us out of a row we want at 20. So a change to a route is kept only when the total goes
  down.
- **Starts.** A change at a time can't get from one kind of route to another, because the
  points in between are worse. So the search scores many starts: the plan's own build, one that
  fills each tree from the top, and one for each deeper talent that goes straight for it (like
  Stormstrike at 25). It searches from the best three.
- **Changes.** From a start we keep making the change that saves the most hours over all the
  levels, until none saves more than the noise of the levels it changes (or 0.05% of the hours
  when that's less, because two routes play the same sessions and most of the noise cancels).
  A change swaps two picks up to three points apart, moves a pick up to three places (the ones
  in between shift over), or takes the last point out of a talent and puts it in another one.
  There are too many of the last kind to sim each at every level, so we first guess each one
  from the build at the middle and the last level without the point and with the new one, and
  sim the best six at every level.
- **The same sessions.** Each level is simmed on its own seed, so two routes with the same
  talents at a level get the same hours there, and we sim each level and talents once. A search
  of the 10 to 30 plan takes 600 to 1,800 sims of a level at 50 sessions each.
- **Respecs.** At a respec level all our points are free, and we search the levels from it on.
  A respec's build is also a route from the first level, and sometimes it's the better route
  (it only moves a point the route could have put there too). So we search the route again from
  the respec builds, and a respec then only shows what a respec saves.
- **The numbers we show** come from other sessions, at the run's iterations, so the search
  can't have picked a route by the luck of its own sessions. The search's own best is a few
  minutes lower than that.
- **Required talents.** Some talents we take whatever the sim says, like Improved Ghost Wolf
  for walking to quests and grind spots, which a grinding session doesn't sim. The plan's
  Required talents (`LevelingPlanSettings.required_talents`) say a talent, its points and the
  level we want them by, like `-require` in the DPS talent search (tools/rotopt). Every route the
  search looks at has them by then, and each respec has them from its level on. A new one
  starts at the earliest level its row allows. The enhancement shaman's plan requires Improved
  Ghost Wolf 2/2 by 16, as early as its row opens.
- **Talents the sim gives nothing for** (`notSimulated`) get no points unless they're
  required. Threat isn't simmed either, and solo leveling doesn't need Spirit Weapons.
- **Rotations stay as they are.** A respec into another tree plays the plan's rotation for
  those levels, so it can look worse than it is.

The first search (2026-10-17, enhancement shaman, 10 to 30, the spec's plan) found a route
25 minutes faster than the plan's talents (19h 22m at 3000 sessions a level). It puts Improved
Lightning Shield and Elemental Weapons where the plan had Improved Ghost Wolf, and after
Stormstrike it takes Tidal Focus, Concussion and Elemental Weapons over Flurry. A respec at 20
saves nothing. With Improved Ghost Wolf 2/2 required by 20 (2026-10-18) the route takes
19h 30m, 8 minutes more, with Improved Ghost Wolf at 17 and 19. The spec's plan now requires it
by 16, as early as its row opens, and its own build takes it at 15 and 16. A search of that plan
finds 19h 11m, 13 minutes faster than its talents.

## Rotation search

With Rotations too (`rotation_search`), the search also tunes a rotation for each level range,
in turns with the route: the rotations with the plan's talents, then the route with them, then
the rotations of the ranges whose talents the route changed, until neither changes (at most 3
times). It's `tools/levelingplan/rotation.go`, on the leveling rotation template that the
leveling optimizer uses too (`tools/rotopt/aplkit/shaman_enh_leveling.go`, only the
enhancement shaman's for now).

- **Ranges.** A range starts at each level that brings an ability the template can use (a
  spell like Chain Lightning at 32, a talent like Stormstrike, an imbue like Windfury Weapon)
  or a new rank of a spell it casts. For the shaman that's every even level and the talent
  levels. A range's rotation names the top ranks, so it casts each new rank as we learn it.
- **Knobs.** A range tunes only the knobs its levels can change. The first range, and a new
  ability 15 levels after we last did, tune every knob. A new ability tunes the rest and mana
  thresholds, the imbue and that ability's knobs (`Knob.Needs`). A new rank only tunes the thresholds,
  since the rank costs more mana. The other knobs carry over from the range before.
- **Elimination.** Each round tries every one-knob change of the range's knobs (a threshold
  only within 2 steps of where it is). All start at 50 sessions on up to 3 of the range's
  levels (only its first for a new rank), and the best quarter goes on to 200, then 800. A
  change wins when it saves more than its noise at 800. We take the best winner, or the best
  of each knob together when that saves more. A range makes up to 3 rounds (1 for a new rank).
  A knob whose worth depends on the imbue (`Knob.WithImbue`, the air totem) is also tried
  together with each other imbue. Windfury Totem does nothing next to Windfury Weapon, so
  from Windfury Weapon neither the totem alone nor Rockbiter alone wins, but the two together
  do.
- **Later passes** keep the knobs of a range whose levels and talents stayed the same, and
  start the others from the pass before.
- **The numbers we show** compare the route with its rotations against the plan's talents
  with the plan's rotations, and the plan's talents with the rotations found, on the confirm
  sessions. "Use this route and rotations" puts the rotations in the plan's library as
  "Searched 10" and so on, in place of an earlier search's, with the level ranges and imbues.

Template rules past 30 (2026-10-18), off by default so the Level 30 rotations stay as they
were: Lightning Bolt at a number of Maelstrom Weapon stacks, Chain Lightning with a number of
mobs on us, Grace of Air or Windfury Totem, Healing Stream Totem in place of Mana Spring Totem,
and Lava Burst for a route that takes the Elemental talent. Frost Shock has no rule, because
its slow does nothing in the sim: our mobs don't flee and we never kite.

The first run (2026-10-18, enhancement shaman, the spec's plan, 10 to 60, 3 passes, 20 min):
the plan's talents and rotations take 123h 42m. The rotations found take 112h 27m with the
same talents, and with the route 107h 48m, 15h 54m faster (±3m). Past 30 they take Windfury
Weapon at 32, then at 42 Flametongue Weapon with Windfury Totem in place of
Strength of Earth Totem, and Lightning Bolt at 5 Maelstrom Weapon stacks. The respecs save
nothing. A 10 to 30 run takes 6 min and saves 37 minutes, 31 of them the rotations'.

That run missed Rockbiter with Windfury Totem at 32, because it tried the imbue and the air
totem only one at a time, and it didn't tune the imbue at 32. With both fixed and the new rules
(2026-10-18, 28 min): the rotations found take 111h 2m with the plan's talents, and 105h 48m
with the route, 17h 55m faster than the plan (±3m). They take Healing Stream Totem from 20 to
24, Rockbiter with Windfury Totem from 32 (about 5% faster a level than Windfury Weapon there),
and Flametongue Weapon with Windfury Totem from 42, without Strength of Earth Totem. The
respecs still save nothing.

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

`DefaultLevelingPlan` in `ui/enhancement_shaman/presets.ts`, since 2026-10-19 the route and
rotations of the 10 to 60 rotation search (Rotation search, the second run): 105h 48m. The
route and the ranges are in `ui/enhancement_shaman/apls/leveling_plan_searched.json`, with a
range only where the rotation or the imbue changes. Thundering Strikes, Improved Ghost Wolf at
15 and 16 (required), Mental Dexterity, Shamanistic Focus, Elemental Weapons, Stormstrike at 25,
then Flurry, Improved Stormstrike, Spirit Weapons, Tidal Focus, Maelstrom Weapon from 38 and
Rage of the Farseer at 43, and Elemental from 44. Rockbiter Weapon, with Windfury Totem from 32,
then Flametongue Weapon with Windfury Totem from 42.

Before that it was the Level 30 Leveling build into Elemental toward the Level 60 + Imp GW
build, and the "Level N Leveling Plan" rotations (the leveling optimizer's best at 10, 15, 20,
25 and 30, still in the rotation presets), with the Level 30 one up to 59. A run of 10 to 60
(2026-10-18, 200 sessions a level) took about 120 hours, 19.4 of them from 10 to 30.

## Open

- A sim history for the Leveling mode.
- Rotation templates for the other specs.
- Fleeing mobs in the leveling sim, which Frost Shock would be for.
- Leveling stat weights above 30, for the gear plan.
- Forever's XP table.
