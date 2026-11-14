# BiS lists

A BiS list is the gear search's best set for a spec, with a ranking for every slot of what
else could go there. A player who can't get the pick for a slot reads down its list to the
next item they can get, and sees what it costs them in DPS.

The lists live with their spec, as `ui/<spec dir>/bis/<name>.bis.json`, for example
`ui/enhancement_shaman/bis/level60.bis.json`. BayovCore reads them: `bayovui/build/bis.py`
in the user's WoW folder turns them into the addon's data.

## How a list is made

`tools/rotopt -gear-search` finds the set the usual way (see `tools/rotopt/gear.go`), slot by
slot until a full round changes nothing. With `-gear-rank-out` it then ranks each slot
(`tools/rotopt/bis.go`). Every item the pool offers for the slot goes into the set in place of
the pick, with the rest of the set as it is and the slot's enchant carried over, and we record
its DPS and what it loses against the set. We keep the best `-gear-rank-keep` (10 by default)
per slot, the picks included.

A slot can have hundreds of items, so we race them instead of scoring them all at `-iters`.
First we drop the weaker rolls of each random suffix. Two rolls of an item with the same suffix
name only differ in how much of each stat they give, so a roll with no more of any stat than
another one can't do better. Then we score everything at `-gear-screen-iters`, then what is
left at about the geometric mean of the two (447 for 100 and 2000), then the rest at `-iters`.
Between two steps an item is dropped only when it is behind by more than `-confidence`
standard errors of the gap: behind an empty slot, or behind the item one and a half times
`-gear-rank-keep` down the list (counting each item and suffix name once). The runs share their
seed, so the error of a gap is the paired one, from each iteration's DPS: a few points of a
stat change little from one iteration to the next, while 2% crit moves every roll after it and
pairs about as badly as two unrelated runs. An item that scores
exactly what the empty slot scores is dropped too, since the runs share their seed and it
changes nothing the sim sees. Before 2026-11-13 we kept a plain top 50 of the 100 iteration
screen, which is several DPS off. Trinkets are within a few DPS of each other and of an empty
slot, so the Phase 0 trinket list lost Blackhand's Breadth (7.7 DPS better than an empty slot)
and came out with 19 of its 25 entries.

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
- With `-gear-rank-two-hand-only` we skip the one hand setup. When the pick is a two hander, the
  main hand list has only two handers and the off hand list is empty, since the off hand stays
  empty next to it. BayovCore shows that off hand as empty.

Two things are left out. A random suffix item comes in several rolls of the same name ("of the
Tiger" with more or less of each stat), so only the best roll of each name is listed. And an
item that scores no better than an empty slot does nothing in the sim (a Hook of the Master
Angler in a trinket slot), so it isn't listed at all.

A loss is the DPS difference at `-iters` with the same random
seed, so two entries closer than a few tenths are a tie. An entry can even show a small
negative loss when it is within the noise of the pick.

The picks follow the ranking. When a slot's best item beats the set at all at `-iters`, we
score it again next to the set at `-gear-rank-confirm-iters` (10 times `-iters` by default). If
it beats the set by more than `-confidence` paired standard errors there, it takes the pick's place and we rank everything again on the new set, at most three
rankings in all. Confirmed items in different slots go in together, the biggest gain first,
unless together they score below the biggest one alone. Every slot works this way, a one hand
setup in place of a two hander pick too. The gear search can keep an item the ranking then puts
behind, because the ranking runs on the set's best enchants and races every item of the pool,
where the search screened them at a few hundred iterations. On 2026-11-14 the Phase 0 list
kept Wildheart Bracers while its own wrist ranking had Windtalker's Wristguards 2.5 DPS ahead.
We used to ask for the `-confidence` margin at `-iters` too, but then Don Julio's Band stayed
2.3 DPS ahead of the rings on the next run, and it was 2.4 ahead at 20000 iterations.

## Enchants

With `-gear-rank-enchants` the list also ranks the enchants (`tools/rotopt/bis_enchants.go`).
Before the items are ranked, each slot of the set tries every enchant the database has for its
item: the enchant's item type, a two hand enchant only on a two hander, the kits' item and
character level, and the class or profession an enchant may need. `-gear-phase` holds for them
too, so the Zul'Gurub and Naxxramas shoulder and head enchants never come up. The best one
stays when it beats the slot's enchant by more than the noise, and then the next slot goes on
the set as it now is. We go round until a round changes nothing, so that round's lists are all
scored on the final set. Then the items are ranked on the set with those enchants, so the
list's `dps` has them.

Anything that goes in the enchant slot is an option of it, not only the enchanter's: the
Dire Maul arcanums on the head and legs, the leatherworking kits, and Iron Counterweight on a
two hander. One slot holds one of them, so Iron Counterweight and Crusader never stack (user).

The enchants are a few DPS apart, and haste and crit pair badly from one run to the next, so
`-gear-rank-enchant-iters` (10000 in `bis_phase0.sh`) is the iterations of their ranking. A
first try at 1000 iterations had Arcanum of Rapidity 4 DPS behind on the head and 6 ahead on
the legs, where it is level on both. An enchant that scores exactly what no enchant scores
does nothing in the sim (Stamina, resistances, Minor Speed), so it isn't listed. Each slot
keeps `-gear-rank-enchant-keep` (10) enchants, the pick included.

The value of `-gear-rank-enchants` is the hand check of what the game has open, keyed by the
enchant's spell ID (`tools/rotopt/phase0_enchants.json`, 2026-11-14). `avail` is `now`,
`grind`, `doubtful` or `out`, as for items, `note` says why for the player, and `source` is
where the formula comes from, for us. `out` leaves the enchant out of the search (Molten Core's
Weapon Spell Power, the Alliance quest's Fiery Blaze). A formula with no known source or a drop
under about 0.5% is `doubtful`, and Honored reputation or a 0.5% to 5% drop is `grind`. An
enchant the search lists with no line in the file gets a warning in the output, so the check
can be filled in.

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
      {"id":10502,"name":"Spellpower Goggles Xtreme","dps":89.12,"loss":0,"bis":true,"enchant":1506},
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
  },
  "enchants": {
    "head": [
      {"id":1506,"spell":15397,"item":11645,"name":"Lesser Arcanum of Voracity","effect":"+8 Strength","dps":89.12,"loss":0,"bis":true},
      {"id":2543,"spell":22840,"item":18329,"name":"Arcanum of Rapidity","effect":"+1% attack speed","dps":89.01,"loss":0.11,"avail":"doubtful","note":"Needs Pristine Black Diamond, source unknown"}
    ],
    "hands": [
      {"id":1248640,"game":2563,"spell":1248640,"name":"Enchant Gloves - Superior Strength","effect":"+15 Strength","dps":89.12,"loss":0,"bis":true,"avail":"grind","note":"Merchant's Favor Honored"}
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
- `enchant` on a pick is the enchant the set has on it, by its `id` in the slot's enchant list.
- `enchants` (only with `-gear-rank-enchants`) has a list for each slot that takes an enchant,
  keyed like `slots` (`head`, `neck`, `back`, `chest`, `wrist`, `hands`, `legs`, `feet`,
  `mainhand`, ...), the pick first (`bis`) and then best first. `id` is the enchant's effect ID
  in the sim, what a set's `enchant` holds. `game` is the enchant ID the game puts on an item
  link, when it isn't `id`: Forever's own enchants have their spell ID as effect ID in the sim
  (`tools/database/forever_enchants.go`), and the client's (`assets/db_inputs/forever_enchants.json`)
  says what the game uses. `spell` is the spell its formula teaches, and `item` the item that
  applies it, for an arcanum, a kit or Iron Counterweight. `effect` says what it gives
  ("+8 Strength", "+4 all stats"), empty for a proc or weapon damage enchant (Crusader, Fiery
  Weapon). `dps` and `loss` are at `-gear-rank-enchant-iters`, so the pick's `dps` can differ
  a little from the list's. `avail` and `note` are from `phase0_enchants.json` (see Enchants).
- `avail` and `note` come from checking by hand what the game has open
  (`tools/rotopt/phase0_content.json`, 2026-11-12). `avail` is `grind` (open but slow) or
  `doubtful` (we don't know of a way yet), and an entry we can get now has none. `note` says
  why, or what's still in the way. `bis_phase0.sh` leaves out every item the check marks `out`
  (not before the raids) and copies `avail` and `note` onto the entries after the search.

## Regenerating

`tools/rotopt/bis_lists.sh` writes the shaman lists, from the repo root. It builds each spec's
fight with mksettings and runs the search with `-gear-rank-out`. It takes a while (the script
says how long), because the level 30 pools are big. A new spec is a few more lines there.

`tools/rotopt/bis_phase0.sh` writes the level 60 Phase 0 Enhancement list
(`ui/enhancement_shaman/bis/level60_phase0.bis.json`). It runs the DPS Report's default
Enhancement fight (Orc, Engineering and Leatherworking, level 63 boss, 180s) with the talents
the user plays, from a real pre-raid set. `TALENTS` sets the talents, and the default is
`053303102-054030110051212141-03` (2026-11-14). The search uses `-gear-no-raids` and
`-gear-phase 2`, so only what Forever has before its raids open can come up
(`tools/rotopt/phase0.go`). It also leaves out the items `tools/rotopt/phase0_content.json`
marks `out`, and keeps 25 items a slot. Our item data puts Classic patch 1.10's Dungeon Set 2
chain in phase 1 or 2, so its summoned bosses' drops (Sothos and Jarien, Lord Valthalak, Kormok,
Isalien, Mor Grayhoof, Theldren) and its quest rewards (Beads of Ogre Might) are `out` there by
hand. The other way round, the items marked `include` there join whatever their phase
(`-gear-include`). Our data takes its phases from Classic, where these came later, but Forever has
them now: the Darkmoon cards and ticket rewards (the Faire is open), Totem of Rage (Magmus) and
Totem of Sustaining (Scholomance), the Defilers' Arathi Basin rewards (doubtful until we know
Arathi Basin is open), and every other item the hand check keeps that our data puts after phase
2, so a doubtful item is marked on the list rather than left out. The list keeps the off hand empty next to its two hander
(`-gear-rank-two-hand-only`). Then it ranks the enchants on the set (see Enchants), with
`phase0_enchants.json` as the hand check, and the items on the set with its best enchants.

## Before release

Level 30 was the beta's level cap, so the first lists (Enhancement and Elemental) used it as
endgame: mksettings' level 30 boss for 120 sec, with every item a level 30 can wear. We removed
them on 2026-11-05 (user), because BayovCore showed their picks to leveling characters at every
level. Forever came out on 2026-11-04, and the first level 60 list is the Phase 0 Enhancement one
(2026-11-11).
