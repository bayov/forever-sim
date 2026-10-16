# BiS lists

A BiS list is the gear search's best set for a spec, with a ranking for every slot of what
else could go there. A player who can't get the pick for a slot reads down its list to the
next item they can get, and sees what it costs them in DPS.

The lists live with their spec, as `ui/<spec dir>/bis/<name>.bis.json`, for example
`ui/enhancement_shaman/bis/level30.bis.json`. BayovCore reads them: `bayovui/build/bis.py`
in the user's WoW folder turns them into the addon's data.

## How a list is made

`tools/rotopt -gear-search` finds the set the usual way (see `tools/rotopt/gear.go`), slot by
slot until a full round changes nothing. With `-gear-rank-out` it then ranks each slot
(`tools/rotopt/bis.go`). Every item the pool offers for the slot goes into the set in place of
the pick, with the rest of the set as it is and the slot's enchant carried over, and we record
its DPS and what it loses against the set. Each slot is screened at `-gear-screen-iters` first
and the best `-gear-screen-keep` are scored again at `-iters`, the same as the search does. We
keep the best `-gear-rank-keep` (10 by default) per slot, the picks included.

A few slots work differently:

- Rings and trinkets are one list each. The two picks come first. Every other item is scored
  in both slots and listed with the better of the two, since we'd put it in place of the pick
  it loses less to. The two slots never hold the same item.
- When the main hand pick is a two hander, a one hander needs an off hand to be fair. We find
  the best one hander with nothing in the off hand, then the best off hand next to it, and then
  score every one hander next to that off hand. Each one hander names that off hand in `with`.
  Two handers and one handers share the main hand list.
- The off hand list then says what we'd hold if we went one hand. It is scored next to that
  one hander, its `with` names it, and none of its entries is a pick. When the pick is a one
  hander, the off hand list is scored next to the main hand pick instead, and the off hand pick
  comes first.

Two things are left out. A random suffix item comes in several rolls of the same name ("of the
Tiger" with more or less of each stat), so only the best roll of each name is listed. And an
item that scores no better than an empty slot does nothing in the sim (a Hook of the Master
Angler in a trinket slot), so it isn't listed at all.

Enchants aren't part of the list. A loss is the DPS difference at `-iters` with the same random
seed, so two entries closer than a few tenths are a tie. An entry can even show a small
negative loss when it is within the search's noise of the pick.

## The file

```json
{
  "name": "Level 30",
  "spec": "ele",
  "class": "SHAMAN",
  "race": "Orc",
  "faction": "Horde",
  "level": 30,
  "talents": "550430130010301",
  "durations": [120],
  "dps": 89.12,
  "made": "2026-10-16",
  "sim": "abd7cd344ed1+",
  "slots": {
    "head": [
      {"id":10502,"name":"Spellpower Goggles Xtreme","dps":89.12,"loss":0,"bis":true},
      {"id":15339,"suffix":2002,"name":"Pathfinder Hat of Nature's Wrath","dps":87.97,"loss":1.15}
    ],
    "chest": [
      {"id":273044,"name":"Arcanowisp Robes","dps":89.12,"loss":0,"unsourced":true,"bis":true},
      {"id":7065,"name":"Green Silk Armor","dps":88.65,"loss":0.47}
    ],
    "mainhand": [
      {"id":271804,"name":"Soulsplatter Mace","dps":89.12,"loss":0,"with":7749,"bis":true},
      {"id":271803,"name":"Greatstaff of the Necrokhans","dps":86.71,"loss":2.41}
    ],
    "offhand": [
      {"id":7749,"name":"Omega Orb","dps":89.12,"loss":0,"with":271804,"bis":true},
      {"id":273027,"name":"Raptor's Gaze","dps":88.56,"loss":0.56,"with":271804}
    ]
  }
}
```

- `name` is the list's label and `spec` the spec's name in mksettings (`enh`, `ele`, ...).
  `class`, `race`, `faction`, `level`, `talents` and `durations` are the fight it was simmed
  on, and `dps` is the set's DPS there.
- `made` is the date and `sim` the forever-sim commit it ran on, with a `+` when the working
  copy had changes on top.
- `slots` has `head`, `neck`, `shoulder`, `back`, `chest`, `wrist`, `hands`, `waist`, `legs`,
  `feet`, `finger`, `trinket`, `mainhand`, `offhand` and `ranged`, picks first and then best
  first. A slot the set leaves empty has no pick, only the items that do something there.
- An entry has the item's `id`, its random `suffix` when it has one, its `name` (with the
  suffix), `dps` and `loss`. `bis` marks the picks. `unsourced` marks an item wowhead knows no
  way of getting yet, worth a look by hand. `with` is the item in the other hand: on a one hand
  main hand the off hand it was scored with, on an off hand the main hand.

## Regenerating

`tools/rotopt/bis_lists.sh` writes the shaman lists, from the repo root. It builds each spec's
fight with mksettings and runs the search with `-gear-rank-out`. It takes a while (the script
says how long), because the level 30 pools are big. A new spec is a few more lines there.

## Level 30 for now

Level 30 is the beta's level cap, so the lists use it as endgame until Forever releases. The
fight is mksettings' level 30 boss for 120 sec, and the pool has every item a level 30 can
wear. Once Forever is out, the lists move to level 60 with Forever's real item data, and the
level 30 ones can go.
