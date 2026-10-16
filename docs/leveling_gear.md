# Leveling gear plan

What a character wears at each level from 1 to 60 while speed leveling, for the leveling sim
and the talent route. The Leveling Gear tab shows it (Leveling mode only). The plan is
`ui/core/leveling_gear.ts`.

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
- **Wearing it.** We put a find on only when it beats what we wear in that slot, by the stat
  weights at that level. Most finds don't, so a slot can keep a lucky blue for ten or twenty
  levels. A ring or a trinket replaces the weaker of the two, and a two-hander also takes the
  off hand.
- **Start.** Level 1 starts in a white chest, legs and boots, with a white weapon.

The stat weights are the spec's leveling stat weights preset closest to the level (the
enhancement shaman has Level 30 Leveling only, so far), or the Current EP for a spec without
one.

A plan is one random history from its seed, and Reroll plays out another. Each level and kind
of slot rolls its own random numbers, so the same seed gives the same plan, and an item we add
doesn't change any other find.

## Our items

Add Item puts an item we have into the plan from a level on: a level, a slot and the item (a
random suffix item is one entry per suffix). The plan offers it at that level like a find, and
wears it until something better comes along. The grid gives our items a gold edge.

## Using it

Equip puts the selected level's plan gear on in the Gear tab and makes us that level, so the
leveling sim runs on it. The seed and our items are saved with the gear settings.

## Open

- The finds and the quality mix are guesses (notes.md, Need to Verify). Bayov-Oom's BayovCore
  history as it levels will show the real pace.
- Leveling stat weights at more levels, so the plan judges low level items by their own
  weights.
- The leveling sim can take the plan's gear at its own level directly, and average over many
  seeds, for the talent route.
