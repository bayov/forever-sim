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

## Conditional ranks

Some items do more in some fights. Blackhand's Breadth's mark doubles on Orcs and Ogres, Seal of
the Dawn gives attack power only against Undead, and Golden Banana doubles in forest or
grassland. The list is ranked on a plain boss, where they get none of that. So with
`-gear-rank-conditions` (`tools/rotopt/bis_conditions.go`) we rank each such item a second time
with its condition met, and write where it would go on the same list in its `conditional` field.
The normal ranks never change.

The item effects declare their conditions in the sim (`core.RegisterItemConditions`, in
`sim/core/item_conditions.go`), so we don't screen the whole pool under every condition. The
creature type helpers (`core.NewMobTypeAttackPowerEffect` and its spell damage twin) declare
theirs, and a custom effect adds one line next to it. A creature condition goes on every target
of the encounter. A condition the fight already meets is skipped.

For each condition we score the set with the condition met, once, and then each of its items in
the set with the condition met too. The item's conditional `loss` is what it loses against that
set. A ring or trinket goes in place of either pick, and the better of the two counts. These runs
are at `-gear-rank-confirm-iters` and share their seed.

Its conditional `rank` is its place among the list's other entries by that loss. A negative loss
beats a pick, so it is rank 1. Otherwise it comes after every other entry whose normal loss is the
same or less. The picks lose 0, so a rank no higher than the slot's number of picks always means
BiS under the condition, and anything else is past the picks. A pick with a condition keeps its
place: its rank is its own position and its loss 0.

An item can make the list only with its condition met, like Seal of the Dawn on a trinket list. So
we look at every conditional item of the slot's pool (the same filters as the search), not only
the listed ones. When one's conditional rank is within `-gear-rank-keep`, it joins the list as a
normal entry, with its normal `dps` and `loss` at `-iters`, at the place its normal DPS gives it.
The list can then be longer than `-gear-rank-keep`.

Item sets and enchants with a condition (the Undead slaying sets, Beastslaying, Elemental
Slayer) aren't ranked this way. A set bonus needs the other pieces of the set, and the enchants
have their own list.

## On cooldown ranks

The list is ranked with every use pressed at the pull (the APL's `autocastOtherCooldowns`). The
user's addon swaps trinkets between fights, so it also needs to know what an item is worth while
its use is on cooldown: a strong use can make a trinket BiS for the pull, and filler for the next
fight. So with `-gear-rank-cooldowns` (`tools/rotopt/bis_cooldowns.go`) every listed item with a
use gets a second rank, with its use never pressed, so only its stats, equip effects and procs
count. The picks get it too. It is written in the entry's `cooldown` field.

An item has a use when its effect registers a major cooldown in the sim, which covers every
trinket use and other usable gear like Manual Crowd Pummeler. The sim switch is
`Player.item_uses_off`: the use isn't registered as a cooldown, so the autocast never presses it,
and an APL that names it finds it never castable.

These runs are at `-gear-rank-confirm-iters` and share their seed. The base is the set itself,
with its uses pressed, scored the same way once. An item that isn't a pick goes in the places its
list scored it in, a ring or trinket in place of either pick, and the better of the two counts.
The rank is the conditional ranks' rule: a negative loss is rank 1, and otherwise it comes after
every other entry that loses as little or less.

These runs (and the effect values below) give each kind of roll its own random source
(`SimOptions.use_labeled_rands`). Without it, one Hand of Justice proc shifts every roll after it
in the iteration, and the two runs of a pair are about as noisy as two unrelated runs. On
2026-11-20 it took the paired noise of the effect values from 2.5 to 0.6 DPS at 2000 iterations.

A trinket with a use can lose less with the use off than with it pressed. On the 2026-11-20 set,
Adaptive Combat Assistant's use goes first at the pull and puts the Forever burst trinkets' shared
category on cooldown for 15s, so Chained Essence of Eranikus goes off 15s later. Its use is worth
-4 DPS there.

## Effect values

BayovCore scores an item by stat weights, so it can't score an item's effects (a proc, a use, a
mark) and says so in the tooltip. With `-gear-effects-out` (`tools/rotopt/bis_effects.go`) we
measure what each item's effects add on the BiS set, and BayovCore adds that to the item's stat
score.

- `dps` is the set with the item in its slot, less the same set with all of the item's effects
  left out (`Player.item_effects_off`), so only its stats count there. A ring or trinket goes in
  place of the weaker pick, the one the set loses less without. A pick is the set as it is.
- `cooldown`, for an item with a use, is what its effects add while the use is on cooldown: the set
  with only the use never pressed, less the set with all its effects off. It is 0 for an item whose
  only effect is its use.
- `conditional` is `dps` with each of the item's conditions met.

The items are every item of the gear pool (the search's own filters: phase, no raids, faction,
class, minimum item level, the includes and excludes) that has an item effect in the sim, and the
set's picks. A one hander goes in the main hand with the off hand empty. An off hand only item
(Force Reactive Disk) is left out next to a two hand pick, since nothing places it. Each run is
at `-gear-effects-iters` (40000 at 45s in `bis_phase0.sh`), where the paired noise is about 0.15
DPS for most items and 0.4 at most.

```json
{
  "dur": 45,
  "dps": 1120.18,
  "iters": 40000,
  "made": "2026-11-20",
  "sim": "e07e8fdb812d+",
  "items": {
    "10455": {"name":"Chained Essence of Eranikus","dps":29.5,"cooldown":0,"conditional":[{"label":"vs Trolls","dps":59.3}]},
    "11815": {"name":"Hand of Justice","dps":8.2}
  }
}
```

`dur` is the fight length, `dps` the set's DPS there, and `iters` the iterations of each run.
`made` and `sim` are as in the list file.

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
- `conditional` (only with `-gear-rank-conditions`, and only on an item with a condition) says
  where the item would go on this list with its condition met, one line for each condition:
  `{"label":"vs Orcs and Ogres","rank":2,"loss":-0.42}`. `label` is what the player reads,
  `rank` is its place (no higher than the number of picks means BiS), and `loss` is what it
  loses against the set with the condition met. See Conditional ranks.
- `cooldown` (only with `-gear-rank-cooldowns`, and only on an item with a use) says where the
  item would go on this list with its use never pressed, as while it is on cooldown:
  `{"rank":9,"loss":13.7}`, with the same rule as `conditional`. See On cooldown ranks.
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
(`ui/enhancement_shaman/bis/level60_phase0.bis.json`) and its effect values
(`level60_phase0.effects.json` next to it). It runs the DPS Report's default
Enhancement fight (Orc, Engineering and Leatherworking, level 63 boss) at `DUR` seconds with the talents
the user plays, from a real pre-raid set. `TALENTS` sets the talents, and the default is
`053303102-054030110051212141-03` (2026-11-14). The search uses `-gear-no-raids` and
`-gear-phase 2`, so only what Forever has before its raids open can come up
(`tools/rotopt/phase0.go`). It also leaves out the items `tools/rotopt/phase0_content.json`
marks `out`, and keeps 25 items a slot. Our item data puts Classic patch 1.10's Dungeon Set 2
chain in phase 1 or 2, so its summoned bosses' drops (Sothos and Jarien, Lord Valthalak, Kormok,
Isalien, Mor Grayhoof, Theldren) are `out` there by hand. Forever made the chain Mokvar's Five
Thunders upgrades, so Beads of Ogre Might (Falrin's Vendetta, a step without a summon) is open. The other way round, the items marked `include` there join whatever their phase
(`-gear-include`). Our data takes its phases from Classic, where these came later, but Forever has
them now: the Darkmoon cards and ticket rewards (the Faire is open), Totem of Rage (Magmus) and
Totem of Sustaining (Scholomance), the Defilers' Arathi Basin rewards (doubtful until we know
Arathi Basin is open), and every other item the hand check keeps that our data puts after phase
2, so a doubtful item is marked on the list rather than left out. The list keeps the off hand empty next to its two hander
(`-gear-rank-two-hand-only`). Then it ranks the enchants on the set (see Enchants), with
`phase0_enchants.json` as the hand check, and the items on the set with its best enchants, and
the conditional items with their condition met (see Conditional ranks), and the items with a use
with their use on cooldown (see On cooldown ranks). Last it writes the effect values.

`DUR` is 45 by default (user, 2026-11-20). It comes from the user's level 60 dungeon logs of
2026-11-20: trash fights over 20s had a median of 32 to 38s, bosses 42s, and 7 to 13% of fights
went over 60s. Every use is pressed at the pull and an addon swaps trinkets between fights, so a
strong use can make an item BiS at that length. When the raids open (2026-12-09) their bosses and
trash may need a longer fight, measured again from raid logs. The list was ranked at 180s before.
A short fight's DPS varies more from one iteration to the next, so the script scales every
iteration count by 180/`DUR`, for the same precision and run time as the 180s list had.

`RANK_ONLY=1 tools/rotopt/bis_phase0.sh` skips the gear search (`-gear-rank-only`) and ranks the
list's own picks again, enchants included. It is for a rerun after
a sim or item data change. The ranking still puts in an item that beats its pick. `FROM` is the
list it takes the picks from (the list itself by default), `OUT` writes somewhere else, and any
arguments go on to rotopt after the script's, so `-iters 300` makes a quick check.

`EFFECTS_ONLY=1 tools/rotopt/bis_phase0.sh` only writes the effect values, for the picks of
`FROM` (about 1.5 min). `EFFECTS_OUT` writes them somewhere else, and `EFFECTS_OUT=none` leaves
them out of a full or `RANK_ONLY` run.

`RANK_ONLY=1 SLOTS=trinket,chest tools/rotopt/bis_phase0.sh` ranks only those lists, their
conditional entries and their enchants (`-gear-rank-slots`), and copies every other list and
enchant list from `FROM` as it is (`-gear-rank-from`). The mainhand and offhand lists are scored
together, so naming either one scores both, but only the named one is written anew. The copied lists keep their numbers, so the list's
top-level `dps` stays the old one. The ranked lists' losses are against the set's DPS as measured
now, and the output prints both. A `SLOTS` run writes the effect values only when `EFFECTS_OUT`
is set.

## Before release

Level 30 was the beta's level cap, so the first lists (Enhancement and Elemental) used it as
endgame: mksettings' level 30 boss for 120 sec, with every item a level 30 can wear. We removed
them on 2026-11-05 (user), because BayovCore showed their picks to leveling characters at every
level. Forever came out on 2026-11-04, and the first level 60 list is the Phase 0 Enhancement one
(2026-11-11).
