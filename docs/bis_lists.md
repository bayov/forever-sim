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

- Rings and trinkets are one list each, with both picks on it. Every other item is scored
  in both slots and listed with the better of the two, since we'd put it in place of the pick
  it loses less to. The two slots never hold the same item.
- When the main hand pick is a two hander, a one hander needs an off hand to be fair. We find
  the best one hander with nothing in the off hand, then the best off hand next to it, and then
  score every one hander next to that off hand. Each one hander names that off hand in `with`.
  Two handers and one handers share the main hand list.
- The off hand list then says what we'd hold if we went one hand. It is scored next to that
  one hander, its `with` names it, and none of its entries is a pick. When the pick is a one
  hander, the off hand list is scored next to the main hand pick instead, and the off hand pick
  is on the list with the others.
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

### The EP trim for tanks and healers

A tank set's score is six boss sims and a healer's is fourteen (seven bosses in two roles), and
a slot's pool holds 500 to 2500 options, most of them random suffix rolls. A DPS search can
afford to screen them all, a tank or healer search can't. So `-gear-ep-keep N` trims the pool
before the search and the ranking (`tools/rotopt/ep.go`). It measures linear stat weights of the
search's own objective on a reference set (`-gear-ep-settings`, the starting set by default):
each stat the pool has goes on the player as a bonus stat, about an item's worth of it, and
its weight is the score it adds over the score without, per point. The runs share their seed
and go through the same scoring as the search, so a tank's weights are in its tank score over
all its bosses and a healer's in damage carried. A healer's reference carries are still measured
on the starting set. Hit and crit count as one stat each, since Forever pays an item's hit and
crit to melee and spells alike. Each slot then keeps its best N options by the weights, from the
item's and the suffix's stats. Set bonuses and enchants don't count there. What the weights
can't judge is always kept: every item with a use or proc effect, modelled or not (so the hand
check still sees the unmodelled ones), every set piece, items with weapon skill, the starting
items and the `-gear-include` items. A weapon slot also keeps the best N of each weapon class
(one hander, two hander, shield, held in off hand) by the weights and by weapon DPS. The run
prints the weights and each slot's count before and after.

### Items we rank but never pick

`-gear-no-pick` takes item IDs the search never picks, but every list still ranks. The Phase 0
list uses it for the items that need Exalted reputation (`avail` `infeasible` in
`phase0_content.json`, user 2026-11-20), like The Unstoppable Force from Frostwolf. That grind
is very long, so the set doesn't count on it, and each slot's pick is the best item we can get.

It also leaves out items that cost too much gold for what they add: more than 30g for each DPS
(`MAX_GOLD_PER_DPS`, user 2026-11-22). `tools/rotopt/phase0_prices.py --write` prices the items
of every spec's list in `phase0_specs.json` in one run (since 2026-11-25, Enhancement's alone
before) and divides by the DPS each adds over the best cheaper item of its slot on that list (the
second best for rings and trinkets). It writes `gold` and `goldPerDps` into `phase0_content.json`,
`goldPerDps` as one number per spec (`{"mage_fire": 160.0, "mage_frost": 12.4}`), since an item
adds different DPS on each list. The script makes every item over the line on the spec's list
infeasible there, with "133g per DPS" in its note.

An item's gold is the cheapest of its auction house price, its recipe's materials and the effort
of farming it. The first two come from our auction house scans (BayovCore's BayovAuctionData).
The effort is the hours it takes at 30g an hour (user, 2026-11-22). Without it a rare drop cost
nothing, so the list picked Treant's Bane, a 0.38% drop from the Dire Maul tribute, over Brain
Hacker at 255g. The hours are the runs we expect times the time of a run. The runs are 1 / (drop
chance x roll share), and the share is 1: we assume only we roll on a dungeon drop (user,
2026-11-25; it was 0.5, one other player, until then). A run takes the dungeon's typical clear
time plus 20 minutes, for getting there and finding a group. So a bind on pickup drop costs only
its effort, and a bind on equip drop costs the cheaper of its effort and the auction house.
Treant's Bane comes to 263 runs, 219 hours and about 6600g, so Brain Hacker is still the better buy. Quest,
vendor and reputation items get no effort price yet, and neither do world drops outside
dungeons, since their chance is too small to farm.

The chances come from the BayovUI wiki (Naowh's Dungeon Journal), and from AtlasLoot Classic's
`droprate.lua` where the wiki has none. The clear times are estimates in the script's
`RUN_MINUTES`. `--gold-per-hour`, `--roll-share` and `--run-overhead` change the three numbers,
and the table says how it priced each item. Prices move, so run it again before a rerun, and once
more after one, since new picks change what each item is measured against.
But a player who has the reputation still sees where the item goes on the set. Such an item is
ranked like any other, on the set with its real loss, but it never takes a pick's place
(see above). When it beats the pick, it sits above the pick with a negative loss, since each list
is in order of loss. A starting item that is on it comes off before the search, like an
excluded one.

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

Its conditional `rank` is its place among the list's other entries by that loss: after every other
entry whose normal loss is the same or less. The picks lose 0, so an item that loses 0 or more
comes after them. The rank doesn't say whether the item is BiS under the condition, since an item
we never pick can sit above the picks. Its `bis` flag does: it is set when the item beats the set
with the condition met (a negative loss) and isn't one we never pick. A pick with a condition
keeps its place: its rank is its own position, its loss 0, and it is `bis`.

An item can make the list only with its condition met, like Seal of the Dawn on a trinket list. So
we look at every conditional item of the slot's pool (the same filters as the search), not only
the listed ones. When one's conditional rank is within `-gear-rank-keep`, it joins the list as a
normal entry, with its normal `dps` and `loss` at `-iters`, at the place its normal loss gives it.
The list can then be longer than `-gear-rank-keep`.

Item sets and enchants with a condition (the Undead slaying sets, Beastslaying, Elemental
Slayer) aren't ranked this way. A set bonus needs the other pieces of the set, and the enchants
have their own list.

## How uses count

With `-use-dur` (`tools/rotopt/use_values.go`) a set is scored in two parts (user, 2026-11-20):

    score(S) = DPS at -durs of S with every use off + the sum of useValue(i, S) over S's uses
    useValue(i, S) = DPS at -use-dur of S - the same with only i's use off

The Phase 0 list scores at 180s with a 90s `-use-dur`. Everything but the uses counts on 180s,
because there's little or no time to regen mana between the user's fights. The uses count on 90s,
pressed at the pull as the APL's `autocastOtherCooldowns` does. The user's level 60 dungeon fights
are short, an addon swaps trinkets between them, and swapping a usable item in puts it on a 30s
equip cooldown, so sometimes its use never gets pressed. 90s takes that into account roughly.

Every score of the gear search works this way: the search, the list's ranks, picks and losses, the
conditional ranks (with the condition met in both parts) and the enchants. An item has a use when
its effect registers a major cooldown in the sim, which covers every trinket use and other usable
gear like Manual Crowd Pummeler. The sim switch is `Player.item_uses_off`: the use isn't
registered as a cooldown, so the autocast never presses it, and an APL that names it finds it
never castable.

The use values of a set depend mostly on which uses it holds, so we measure them once for each
group of uses and reuse them for every set with the same uses. An option without a use then costs
only its run at 180s, and an option with a use adds the 90s runs of its group (one with every use
pressed, and one for each use with it off). Each ranking measures its base set's values on the
base itself. The 90s runs are at `-use-iters` (20000) with labeled random numbers (below).

A use can be worth little on a set. On the 2026-11-20 set, Adaptive Combat Assistant's use goes
first at the pull and puts the Forever burst trinkets' shared category on cooldown for 15s, so
Chained Essence of Eranikus goes off 15s later. Its use is worth +0.4 DPS there.

## On cooldown ranks

The user's addon swaps trinkets between fights, so it also needs to know what an item is worth
while its use is on cooldown: a strong use can make a trinket BiS for the pull, and filler for the
next fight. So with `-gear-rank-cooldowns` (`tools/rotopt/bis_cooldowns.go`) every listed item with
a use gets a second rank, from its score without its own use value. The other items' uses still
count. Its cooldown loss is its loss plus its use value, so this needs no runs of its own. The
picks get it too. It is written in the entry's `cooldown` field, with the rank rule of the
conditional ranks: it comes after every other entry that loses as little or less, and `bis` says
it would still be the pick.

A trinket that adds nothing is never ranked (user, 2026-11-24), since its rank would only say where
an empty slot goes. `zeroTrinkets` scores the set with each listed trinket, its use off, against the
set with that slot empty. Those that add less than 0.05 DPS add nothing. That includes the ones whose
stats do nothing for the spec, like Flamebreak Talisman's fire resistance:

- With a use (Barov Peasant Caller), its `cooldown` has only the loss, and no rank.
- Without one, but with a conditional rank (Rune of the Dawn, vs Undead), the entry is marked
  `zero`. It keeps its conditional ranks, and its place on the list says nothing.
- With neither, it isn't listed.

The use values and the effect values below give each kind of roll its own random source
(`SimOptions.use_labeled_rands`). Without it, one Hand of Justice proc shifts every roll after it
in the iteration, and the two runs of a pair are about as noisy as two unrelated runs. On
2026-11-20 it took the paired noise of the effect values from 2.5 to 0.6 DPS at 2000 iterations.

## Uses against several mobs

A dungeon pull is often several mobs, and some uses hit them all: Smokey's Lighter breathes fire
in a cone, and Chained Essence of Eranikus leaves a poison cloud. The list is ranked on one boss, where they hit only the boss. So with
`-use-aoe-targets 4` (`tools/rotopt/bis_aoe.go`, user, 2026-11-20) an item with a use gets one more
conditional rank, `vs 4 mobs`, with the rank rule of the other conditionals.

Only the use counts against the mobs. Its use value is measured on the `-use-dur` fight against 4
copies of the boss, all alive for the whole fight and all in reach of a cone or a cloud. The rest of
the score stays on the one boss at `-durs`. That is what the use adds on a pull, without switching
the rotation or the gear to an AoE build. The rotation is the same against 4 mobs (it has no target
count conditions), and its Magma Totem and Fire Nova hit all of them, but a use value is the run
with the use less the run without it, both against the 4 mobs, so that part cancels out. What stays
is what the use adds to the AoE damage. Manual Crowd Pummeler's haste gains a few DPS that way
(+67.1 to +72.1 on a test set), because more swings give more Dragonbreath Chili procs, and the
chili's fire hits all 4 mobs.

Most uses gain little or nothing on the mobs, and a line for them would only repeat the normal
rank. So an item gets the line only when its use is worth more than 1 DPS more against the 4 mobs
than against the boss. The use values are 0.1 to 0.3 DPS apart from one run to the next at 20000
iterations, so that gain is real. The effects file gets the same line in its `conditional`.

On 2026-11-20 the use values went from the boss to 4 mobs like this: Chained Essence of Eranikus
+17.8 to +71.8, Smokey's Lighter +2.7 to +10.9, Eye of the Beast +7.6 to +15.2 (its spell hit
helps the Magma Totem, Fire Nova and chili damage on all 4), Adaptive Combat Assistant +0.5 to +1.7
and Manual Crowd Pummeler +78.3 to +83.3. Everlook Pathcarver hits up to 4 mobs too, but it splits
the same damage among them, so it gains nothing. Eye of the Beast then goes ahead of the picks on
4 mobs, and Smokey's Lighter goes from 22nd to rank 4. The pass adds about a minute to the ranking
of the trinkets and to the effects pass.

## Effect values

BayovCore scores an item by stat weights, so it can't score an item's effects (a proc, a use, a
mark) and says so in the tooltip. With `-gear-effects-out` (`tools/rotopt/bis_effects.go`) we
measure what each item's effects add on the BiS set, and BayovCore adds that to the item's stat
score.

- `dps` is what the item's effects add to the set's score: the set with the item in its slot, less
  the same set with all of the item's effects left out (`Player.item_effects_off`), both at `dur`
  with every use off, plus the item's use value at `use_dur`. A ring or trinket goes in place of
  the weaker pick, the one the set loses less without. A pick is the set as it is.
- `cooldown`, for an item with a use, is what its effects add while the use is on cooldown: `dps`
  without the use value. It is 0 for an item whose only effect is its use.
- `conditional` is `dps` with each of the item's conditions met. For `vs 4 mobs` (see Uses against
  several mobs) that is the part at `dur` as in `dps`, plus the use value against the 4 mobs.

The items are every item of the gear pool (the search's own filters: phase, no raids, faction,
class, minimum item level, the includes and excludes) that has an item effect in the sim, and the
set's picks. A one hander goes in the main hand with the off hand empty. An off hand only item
(Force Reactive Disk) is left out next to a two hand pick, since nothing places it. The runs at
`dur` are at `-gear-effects-iters` (10000 in `bis_phase0.sh`), and the use values at `-use-iters`.
On 2026-11-20 the paired noise was 0.1 DPS for most items and 0.3 at most, at both lengths. The
whole effects pass takes about 2.5 min, the uses against several mobs included.

```json
{
  "dur": 180,
  "use_dur": 90,
  "dps": 1003.28,
  "iters": 10000,
  "made": "2026-11-20",
  "sim": "3c9401f7b436+",
  "items": {
    "9449": {"name":"Manual Crowd Pummeler","dps":76.1,"cooldown":0},
    "10455": {"name":"Chained Essence of Eranikus","dps":16.8,"cooldown":0,"conditional":[{"label":"vs Trolls","dps":35.1}]},
    "11815": {"name":"Hand of Justice","dps":7.5}
  }
}
```

`dur` is the fight length of everything but the uses, `use_dur` the uses' fight length, `dps` the
set's score, and `iters` the iterations of each run at `dur`. `made` and `sim` are as in the list
file.

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
enchant's spell ID (`tools/rotopt/phase0_enchants.json`, 2026-11-14). `avail` is
`doubtful` or `out`, as for items, `note` says why for the player, and `source` is
where the formula comes from, for us. `out` leaves the enchant out of the search (Molten Core's
Weapon Spell Power, the Alliance quest's Fiery Blaze). A formula with no known source or a drop
under about 0.5% is `doubtful`. Anything else is open, and its `note` says what it takes
(Honored reputation, a formula drop rate). An
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
      {"id":1248640,"game":2563,"spell":1248640,"name":"Enchant Gloves - Superior Strength","effect":"+15 Strength","dps":89.12,"loss":0,"bis":true,"note":"Merchant's Favor Honored"}
    ]
  }
}
```

- `name` is the list's label and `spec` the spec's name in mksettings (`enh`, `ele`, ...).
  `class`, `race`, `faction`, `level`, `talents` and `durations` are the fight it was simmed
  on, and `dps` is the set's DPS there. `use_dur`, when there is one, is the fight length the
  uses count on (see How uses count), and `dps` is then the set's score.
- `made` is the date and `sim` the forever-sim commit it ran on, with a `+` when the working
  copy had changes on top.
- `slots` has `head`, `neck`, `shoulder`, `back`, `chest`, `wrist`, `hands`, `waist`, `legs`,
  `feet`, `finger`, `trinket`, `mainhand`, `offhand` and `ranged`. Each list is in order of
  `loss`, best first, and on a tie a pick goes first (user, 2026-11-20). An item we never pick
  (`infeasible`) can sit above the picks, so only `bis` says which entries are the picks, never
  their place. A slot the set leaves empty has no pick, only the items that do something there.
- An entry has the item's `id`, its random `suffix` when it has one, its `name` (with the
  suffix), `dps` and `loss`. `loss` is against the set, so a negative loss beats the slot's
  pick. `bis` marks the picks. `unsourced` marks an item wowhead knows no
  way of getting yet, worth a look by hand. `with` is the item in the other hand: on a one hand
  main hand the off hand it was scored with, on an off hand the main hand.
- `enchant` on a pick is the enchant the set has on it, by its `id` in the slot's enchant list.
- `conditional` (only with `-gear-rank-conditions`, and only on an item with a condition) says
  where the item would go on this list with its condition met, one line for each condition:
  `{"label":"vs Orcs and Ogres","rank":2,"loss":-0.42,"bis":true}`. `label` is what the player
  reads, `rank` is its place on the list, `loss` is what it loses against the set with the
  condition met, and `bis` says it would be the pick then. See Conditional ranks. `vs 4 mobs` is the same
  for an item whose use gains on several mobs, see Uses against several mobs.
- `cooldown` (only with `-gear-rank-cooldowns`, and only on an item with a use) says where the
  item would go on this list without its use value, as while its use is on cooldown:
  `{"rank":9,"loss":13.7}`, with the same rule as `conditional` (`bis` when it would still be
  the pick). A trinket that adds nothing then has only the loss. See On cooldown ranks.
- `zero` marks a trinket that adds nothing on the list's own fight, and is only listed for its
  conditional ranks. See On cooldown ranks.
- `enchants` (only with `-gear-rank-enchants`) has a list for each slot that takes an enchant,
  keyed like `slots` (`head`, `neck`, `back`, `chest`, `wrist`, `hands`, `legs`, `feet`,
  `mainhand`, ...), in order of `loss` like the item lists, with `bis` on the pick. An enchant
  within the noise of the pick can sit above it. `id` is the enchant's effect ID
  in the sim, what a set's `enchant` holds. `game` is the enchant ID the game puts on an item
  link, when it isn't `id`: Forever's own enchants have their spell ID as effect ID in the sim
  (`tools/database/forever_enchants.go`), and the client's (`assets/db_inputs/forever_enchants.json`)
  says what the game uses. `spell` is the spell its formula teaches, and `item` the item that
  applies it, for an arcanum, a kit or Iron Counterweight. `effect` says what it gives
  ("+8 Strength", "+4 all stats"), empty for a proc or weapon damage enchant (Crusader, Fiery
  Weapon). `dps` and `loss` are at `-gear-rank-enchant-iters`, so the pick's `dps` can differ
  a little from the list's. `avail` and `note` are from `phase0_enchants.json` (see Enchants).
- `avail` and `note` come from checking by hand what the game has open
  (`tools/rotopt/phase0_content.json`, 2026-11-12). `avail` is `doubtful` (we don't know of
  a way yet) or `infeasible` (it needs Exalted reputation, user 2026-11-20, or costs over 30g
  per DPS, user 2026-11-22), and an open entry
  has none (user, 2026-11-20: the note tells the effort, so open but slow has no tag of its
  own). `note` says what it takes, or what's still in the way. `bis_phase0.sh` leaves out every
  item the check marks `out` (not before the raids) and copies `avail` and `note` onto the
  entries after the search. An `infeasible` item is never a pick, but it is still ranked, so
  its `loss` can be negative (see Items we rank but never pick).

## Regenerating

`tools/rotopt/bis_lists.sh` writes the shaman lists, from the repo root. It builds each spec's
fight with mksettings and runs the search with `-gear-rank-out`. It takes a while (the script
says how long), because the level 30 pools are big. A new spec is a few more lines there.

`tools/rotopt/bis_phase0.sh` writes the level 60 Phase 0 Enhancement list
(`ui/enhancement_shaman/bis/level60_phase0.bis.json`) and its effect values
(`level60_phase0.effects.json` next to it). It also writes the picks with their enchants as
`ui/enhancement_shaman/gear_sets/level60_phase0.gear.json`, the gear of the sim UI's "Phase 0"
build preset (with the list's talents, under the Phase 0 group). It skips that file when `OUT`
is set, since then the list isn't the real one.

It uses two fight lengths (user, 2026-11-20, see How uses count). Everything but the uses counts
on `DUR` (180s), because there's little or no time to regen mana between fights. Each use counts
for what it adds on `USE_DUR` (90s), for the short dungeon fights and the 30s equip cooldown a
swapped in usable item gets. When the raids open (2026-12-09) both may need measuring again from
raid logs. The list records them as `durations` and `use_dur`, and the effects file as `dur` and
`use_dur`. From 2026-11-20 to the rework the same day, the list was ranked on 45s with the uses in.

It runs the DPS Report's default
Enhancement fight (Orc, Engineering and Leatherworking, level 63 boss) with the talents
the user plays, from a real pre-raid set. `TALENTS` sets the talents, and the default is
`053303102-055030110051202141-03` (2026-11-14). The search uses `-gear-no-raids` and
`-gear-phase 2`, so only what Forever has before its raids open can come up
(`tools/rotopt/phase0.go`). It also leaves out the items `tools/rotopt/phase0_content.json`
marks `out`, and ranks the ones marked `infeasible` without ever picking them (`-gear-no-pick`,
see Items we rank but never pick). It keeps 25 items a slot. Our item data puts Classic patch 1.10's Dungeon Set 2
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
with their use on cooldown (see On cooldown ranks) and against several mobs. Last it writes the
effect values.

`AOE_TARGETS` is how many mobs the uses are also valued against (4 by default, 0 leaves it out),
for the `vs 4 mobs` lines (see Uses against several mobs). The user's dungeon pulls are often
several mobs, and 4 is a typical pack.

`RANK_ONLY=1 tools/rotopt/bis_phase0.sh` skips the gear search (`-gear-rank-only`) and ranks the
list's own picks again, enchants included. It is for a rerun after
a sim or item data change. The ranking still puts in an item that beats its pick. `FROM` is the
list it takes the picks from (the list itself by default), `OUT` writes somewhere else, and any
arguments go on to rotopt after the script's, so `-iters 300` makes a quick check.

A pick that is now `out` or a no-pick item gives its place to the best entry of its list that
isn't one, so the rerun starts from that item and the ranking takes it from there. On 2026-11-26
the Armaments of Battle rewards went out (the Argent Dawn quests aren't in Forever), and
`RANK_ONLY SLOTS=trinket` replaced Talisman of Ascendance this way. An `out` item never joins the
pool, and the script drops it from the lists a `SLOTS` run copies from `FROM` too. A copied
list whose pick went out keeps it, and the run warns, so name that slot in `SLOTS`.

A rerun (`RANK_ONLY`, `SLOTS`, `EFFECTS_ONLY`) reads `FROM`, so it stops when `FROM` was ranked on
other lengths than `DUR` and `USE_DUR`: a few slots ranked on 180s next to lists copied from a 45s
list would read as one list. `MIX_DURS=1` goes on anyway.

`EFFECTS_ONLY=1 tools/rotopt/bis_phase0.sh` only writes the effect values, for the picks of
`FROM` (about 2.5 min). `EFFECTS_OUT` writes them somewhere else, and `EFFECTS_OUT=none` leaves
them out of a full or `RANK_ONLY` run.

`GEAR_ONLY=1 SPEC=<spec> tools/rotopt/bis_phase0.sh` sims nothing and writes the spec's gear set
(`gearOut` in `phase0_specs.json`, `ui/<dir>/gear_sets/level60_phase0_<spec>.gear.json`) from the
picks of its list (`FROM`, the list itself by default) with their enchants. Without `SPEC` it writes
every spec's. A full or `RANK_ONLY` run that writes the real list writes the gear set too. Each spec
has one (2026-11-24), for the sim UI and the DPS Report's level 60 DPS specs.

`RANK_ONLY=1 SLOTS=trinket,chest tools/rotopt/bis_phase0.sh` ranks only those lists, their
conditional entries and their enchants (`-gear-rank-slots`), and copies every other list and
enchant list from `FROM` (`-gear-rank-from`). The mainhand and offhand lists are scored
together, so naming either one scores both, but only the named one is written anew. The list's
top-level `dps` is the final set's, measured in this run as in a full one, and each copied entry
keeps its `loss` but gets its `dps` again from that measure (the set's DPS less the loss), so no
figure is left from the old set (2026-11-24). The output prints how far the set moved from the old
list's `dps`. A `SLOTS` run writes the effect values only when `EFFECTS_OUT` is set.

## Before release

Level 30 was the beta's level cap, so the first lists (Enhancement and Elemental) used it as
endgame: mksettings' level 30 boss for 120 sec, with every item a level 30 can wear. We removed
them on 2026-11-05 (user), because BayovCore showed their picks to leveling characters at every
level. Forever came out on 2026-11-04, and the first level 60 list is the Phase 0 Enhancement one
(2026-11-11).
