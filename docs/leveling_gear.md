# Leveling gear plan

What a character wears at each level from 1 to 60 while speed leveling, for the leveling sim
and the talent route. The Gear part of the Leveling Plan tab shows it (Leveling mode only, see
docs/leveling_plan.md). The plan is `ui/core/leveling_gear.ts`.

## The model

We don't take the best item for each level. While leveling we gear up as we go:

- **Finds.** Each level, each slot comes across a few items from quest rewards and drops: on
  average 0.3 a level for most armor, 0.25 for head, shoulders and back, 0.15 for the neck, 0.3
  for the two rings together, 0.06 for the two trinkets, 0.4 for weapons and 0.1 for the
  ranged slot (`FINDS_PER_LEVEL`). That's a find every three or four levels in most slots.
- **Quality.** A find's quality goes by our level (`QUALITY_MIX`): 85% white at level 1, 60%
  at 8, and greens take over around 12. Blues start at 2% at 12, then 5% at 18, 10% at 25, 14%
  at 35, 18% at 45 and 25% at 60.
- **The item.** A green or blue find is a real item from the database for that slot: one we
  can wear at that level (armor type, weapon kind, faction, required level), of item level 3
  under ours to 5 over, from a drop or a quest, or a world drop with no source. It's picked at
  random, so it often has stats we don't want. A random suffix item rolls one of its suffixes.
  Crafted, vendor and reputation items aren't finds.
- **White items.** The database has none, so a white find is a plain piece we make up: armor
  at 90% of a green piece's of that item level, or a weapon at 85% of a green weapon's DPS
  (the green lines are fit to the database). The database makes the same ones at startup, so
  gear with them in it still loads.
- **Quest rewards and drops.** 60% of finds are quest rewards (`QUEST_SHARE`), and there we
  take the best of 3 for us (`QUEST_CHOICES`). That's where the plan picks what our spec wants.
  The rest are drops, with whatever stats they have.
- **Dungeons.** We run the dungeons of our faction (`LEVELING_DUNGEONS`) with an 80% chance
  each, once: 1 level after the lowest level their loot needs. A second run is rare while
  leveling. Each boss drops one item from its own loot, or two half the time. We get an item
  we'd wear 75% of the time (`DUNGEON_WIN_CHANCE`): half the time one other player wants it
  too, and we win half of those rolls. So Ragefire Chasm's blues come at 14, and Ruins of
  Lordaeron's (a Forever dungeon) at 16. Boss loot isn't a find out in the world.
- **Wearing it.** We put a find on only when it beats what we wear in that slot, by the stat
  weights at that level. Most finds don't, so a slot can keep a lucky blue for ten or twenty
  levels. A ring or a trinket replaces the weaker of the two, and a two-hander also takes the
  off hand.
- **Start.** Level 1 starts in a white chest, legs and boots, with a white weapon.

The stat weights come from the leveling sim, by mobs killed in a session. The enhancement shaman
has them at levels 10, 15, 20, 25 and 30 (the "Leveling Gear Plan" presets), measured on the
plan's own gear at that level with that level's best rotation. Between two of those levels we
blend them by level, and above 30 we keep the level 30 ones. A spec without them takes its
"Leveling" presets, or the Current EP.

Stamina, Intellect and MP5 move with the gear, because they decide how often we die. At level
20, +10 Stamina was worth 2.29 a point on gear with little Intellect (0.19 deaths a session)
and nothing on the gear those weights picked (more Intellect, 0.08 deaths). So each level's
weights are the mean of three rounds of plan gear and probes. The final plan's gear kills as
many mobs as the best round's or more at 15 to 30 (62.7 against 59.7 at 25), and 0.3 fewer at 10.

The weights rate Intellect near Strength from level 15 (2.83 against 2 at 20), so the plan
often wears cloth caster pieces like Robe of Evocation and Darkweave Breeches. Intellect
shortens our rests and keeps us alive through the heals we cast.

## The route

A spec can have a leveling route, the quests and dungeons of a RestedXP guide. The enhancement
shaman has ours, the "Bayov On-Release Guide" walked as an Orc Shaman
(`ui/enhancement_shaman/leveling_route.json`, from `tools/leveling_route/route.py`). Then the
plan's quest rewards and dungeon runs are the route's, and only the drops out in the world and
the boss loot stay random:

- **Quest rewards.** Each quest we turn in gives its reward at the level we're at then. Of a
  choice we take the item that helps us most, and we get every fixed reward. We wear one when
  it beats what we have, like a find. A random find that would be a quest reward doesn't
  happen, and a quest reward is never a drop. 217 of the route's 953 quests give an item in
  the database, which starts at item level 10.
- **Dungeons.** We run each of the route's dungeons once, at the route's level: Ragefire Chasm
  at 14, Ruins of Lordaeron 16, Wailing Caverns 21, Shadowfang Keep 24, Blackfathom Deeps 28,
  Razorfen Kraul 30, Gnomeregan 31, Razorfen Downs 37, Scarlet Monastery 38, Uldaman 42,
  Zul'Farrak 47, Maraudon 49, The Temple of Atal'Hakkar 55 and Blackrock Depths 57. Excavation
  Site (32) and City of Dalaran (33) are Forever's, and the database has no loot for them yet.
- **The level.** The tool counts each quest's XP (wowhead's quest rows) and the guide's "grind
  to" steps, and each section starts at least at the level its name says. It doesn't count
  the mobs we kill on the way, so it can be up to a level behind the guide.

The Gear part's Quest route switch makes the plan make them up again. On the 10 to 60 plan, four
seeds took 97.1 to 98.6 hours with the route (98.0 on average) and 98.5 to 101.7 without it
(100.4), with the plan's own talents and rotations.

A plan is one random history from its seed, and Reroll plays out another. Each level and kind
of slot rolls its own random numbers, so the same seed gives the same plan, and an item we add
doesn't change any other find.

## Our items

A click on a slot of the selected level's gear pane opens the Gear tab's item list, with the
items that level can wear sorted by the stat weights of that level. The item we pick there
(with its random suffix) goes into the plan as ours from that level on. The plan offers it at
that level like a find, and wears it until something better comes along. The grid gives our
items a gold edge, and the pane says "ours" under them.

## Using it

Grind takes the selected level of the plan to the Grinding mode, with its gear, talents and
rotation. The seed and our items are saved with the gear settings.

## Open

- The finds, the quality mix and the dungeon runs are guesses (notes.md, Need to Verify).
  Bayov-Oom's BayovCore history as it levels will show the real pace. Which faction The Hall
  of Thanes is for is a guess too.
- Without a route, the dungeons' quest rewards aren't in: the item database doesn't say which
  quest belongs to which dungeon.
- How often another player in the group rolls on an item we want (half the time) is a guess.
- Leveling stat weights above level 30, once the beta goes past it.
- The leveling sim can take the plan's gear at its own level directly, and average over many
  seeds, for the talent route.
