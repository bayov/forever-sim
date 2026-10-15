# Leveling sim

A third sim mode next to DPS and PvP. It scores how fast a character grinds mobs, counting the
downtime between pulls, instead of how much damage it does in one fight. These are the decisions so far
(the user, 2026-10-15). Enhancement shaman comes first, but nothing in the design is specific to
it.

## What it answers

- **The pull and recovery policy:** how many mobs to pull, when to chain-pull another one, which
  spell and rank to pull with, and when to sit down to eat and drink or heal.
- **Leveling stat weights** for a level band, measured on the leveling score, not DPS. Int,
  Spirit, Stamina and armor are worth something here, unlike in the Level 60 weights. They
  can be exported as a Pawn scale per band (10–20, 20–30, …) to judge upgrades while playing.
- **Talents:** the best no-respec route from 10 to 60 (which point to take at each level),
  plus the best build for each of 30–40, 40–50 and 50–60 with its gain over the no-respec route
  in that band. The sim doesn't decide whether a respec is worth its gold. It reports the gains
  and the user decides.
- **Mob level:** which mob level to grind relative to ours (−2, −1, 0, +1, +2, …).

Quest XP is left out. We assume we grind the whole way.

## The score

**Mobs killed in a 20-minute grinding session.** Bigger is better. The last mob counts as the
fraction of its HP we took off, so the score doesn't jump in whole mobs.

The session also reports **XP gained**, from the XP per kill by mob level. At a fixed mob level
both rank everything the same way. Only the mob level search needs XP, because higher-level
mobs give more XP each.

It also reports deaths per session, so a fast but fragile policy is visible as one.

Across levels, minutes to finish level L = mobs needed for L ÷ mobs per session × 20. Adding
that up over the levels gives the total time to level, which is what the talent route
minimizes. Later levels need more mobs and take longer per mob, so they weigh more in the total
without any hand-picked weights.

## The session loop

Each iteration is one session. HP, mana, cooldowns, buffs and debuffs carry over from pull to
pull, so a 3-minute cooldown is ready for some pulls and not others, as in real play.

1. **Search.** It takes a random time to find the next mob. We run around while we search, so
   only instants go out (totems, shields, buffs), no hard casts. Out-of-combat regen (Spirit,
   mana regen outside the five-second rule) ticks during it. The mob we find stands a random
   10 to 40 yards away.
2. **Pull.** The APL picks how to pull it: any harmful spell at any rank that reaches it
   (Earth Shock within 20 yards, Lightning Bolt rank 1 within 30 to save mana, …) or a face
   pull. A mob notices us when the pull spell lands, so a bolt's flight counts, and then runs to
   us from where it stood, following us if we move. Spells cast while it runs in are free
   damage, and it doesn't attack until it arrives. A mob also notices us once we're within its
   aggro radius: 20 yards at our level, a yard more for each level it has over us, a yard less
   for each under. So a face pull is walking up until it notices us and runs the rest of the
   way, and a mob we find that close notices us at once. The APL can walk up to it, all the way
   or into a spell's range; when it hasn't pulled it, isn't casting and hasn't moved for a
   second, we walk up to it anyway. An instant can still pull it on the way.
3. **Chain pull.** While fighting, the APL decides whether to look for another mob, from our
   health, mana, cooldowns, mobs on us and mobs still running in, so a pull can take one mob,
   two, three or more. Most of the time we run around to find it with the mobs on us
   following, so again only instants and melee. The APL pulls it from where it stands, walks up
   to it with the mobs in tow, or leaves it as the next pull. Pushback makes a cast pull worse
   once a mob is hitting us, so the APL may chain-pull with an instant instead.
4. **Fight.** Every mob has its own HP and dies on its own when its HP runs out, and then it
   stops attacking. We're out of combat only once all of them are dead.
5. **Recover.** Out of combat the APL picks among eating and drinking, bandaging, self-heals
   (once healing exists) or searching for the next pull right away. A healer with spare mana
   can heal instead of eating, since healing turns mana into HP.
6. **Death.** Dying costs a corpse run, and then we come back at partial HP and mana. It's
   a time penalty, so the optimizer can weigh a risky pull against a safe one.

## Mobs

Simple melee mobs only. They run to us and swing. No fleeing, no casters and no special
abilities: those are a small share of leveling, and casters would need interrupts and silences
to be modelled.

Per session, mobs are drawn from configurable mixes of level, armor tier and creature type
(creature type matters for Slaying-style talents and item bonuses). HP, armor, damage, swing and
run speed come from a per-level table.

**The mob table** (`sim/core/mob_stats_auto_gen.go`, `core.NormalMobStatsAt`,
`core.NormalMobTarget`) is the average normal mob at each level from 1 to 60. It's taken from
the 1.12 world database (cmangos classic-db), weighted by how many of each mob are spawned in the
open world. It covers hostile and neutral mobs we can attack, rank 0 only, and leaves out casters
(mage class mobs, 3% of spawns). `tools/export_normal_mobs.py` exports the mobs from the dump to
`assets/db_inputs/mobs/normal_mobs.csv`, and `tools/gen_mob_stats.py` averages them into the
table. A few rows:

| Level | Health | Armor | Damage | Swing | Run speed |
|---|---|---|---|---|---|
| 10 | 202 | 531 | 14.1–18.4 | 1.90 s | 7.5 yd/s |
| 20 | 486 | 924 | 32.0–39.2 | 1.99 s | 8.1 yd/s |
| 30 | 1019 | 1288 | 45.8–55.2 | 1.94 s | 8.1 yd/s |
| 40 | 1792 | 2189 | 66.7–82.0 | 1.96 s | 8.3 yd/s |
| 50 | 2772 | 3123 | 89.9–110.3 | 1.93 s | 8.2 yd/s |
| 60 | 3993 | 3728 | 120.2–148.6 | 1.92 s | 8.0 yd/s |

Damage is a white swing before our armor, with the mob's attack power in it (AP / 14 per second
of swing). `go run --tags=with_db ./tools/logcheck -mobs` measures the same things from our
logs, so we can see where Forever's mobs differ.

We fight from in front, so mobs can parry and block our swings, and we can parry, dodge and block
theirs (the existing "in front of target" option).

## Gear

Synthetic and deliberately not optimized, like real leveling gear. Each slot is something we
picked up a few levels ago, mostly greens, with stats the spec only partly uses. The weapon has
its own curve, since weapon DPS by level drives most of a melee spec's damage. The lag and stat
efficiency should be checked against the BayovCore snapshots of the user's own characters.

## Talents

- **Checkpoints.** We sim at a set of levels, just after key spell ranks and milestones like
  the mount at 40, and interpolate between them.
- **No-respec route.** Given a final build, the points are ordered greedily: at each level we
  take the available point with the biggest gain under the tier rules. A hill-climb over the
  final build, scored by total time to level, sits around that. rotopt's talent hill-climb
  covers most of it.
- **Band builds.** For 30–40, 40–50 and 50–60 we find the best build (a fresh respec at the
  band's start, then a point at each level) and report its gain over the no-respec route in
  that band, in mobs per session and minutes saved over the band.

## Defaults

All of these are configurable. The timing defaults are placeholders until we measure them
from the user's logs.

### Session

| Setting | Default |
|---|---|
| Session length | 20 min |
| Start state | Full HP and mana, all cooldowns ready |
| Death | 3 min corpse run, back at 50% HP and mana. Cooldowns keep ticking during the run. |

### Pulling

| Setting | Default |
|---|---|
| Finding the next mob, out of combat | Random 5–11 s (8 s on average), running |
| Finding a mob to chain-pull, in combat | Random 2–4 s (3 s on average), running 3 times in 4 |
| Where the mob we find stands | Random 10–40 yd, in whole yards |
| Our run speed | 7 yd/s |
| Mob run speed | The level's average from the mob table: 6 yd/s at levels 1–2, about 8 yd/s from level 11 |
| Run-in time | (distance − 5 yd melee) ÷ run speed, from when the pull spell lands: about 3.1 s for Lightning Bolt from 30 yd (after 1.5 s of flight) and 1.9 s for a shock from 20 yd at 8 yd/s |
| Aggro radius | 20 yd for a mob of our level, ±1 yd a level of difference, at least 5 (cmangos `Creature::GetAttackDistance`) |
| Face pull | No free damage. We walk up until the mob notices us. When the APL hasn't pulled a mob, isn't casting and hasn't moved for 1 s, we walk up to it. |
| Buff totem reach | 30 yd (client SpellRadius), checked once a second |
| A mob's first swing | Random within its swing timer after it arrives |
| Max pack size | 5 |

### Mobs

| Setting | Default |
|---|---|
| Target mob level | Our level |
| Spread | Uniform ±1 around the target level (target 20 gives mobs of level 19 to 21) |
| Armor | The level's average (an armor multiplier, 1 by default). Within a level armor barely varies: it goes by the mob's class, and the melee mobs (warrior class, 85% of spawns) all have about the same. |
| Creature type | The level's mix from the mob table, e.g. at 30: 56% beast, 34% humanoid, 7% demon, 3% undead |
| HP, armor, damage, swing and run speed by level | The mob table (below), until the logs check it |
| XP per kill | Classic's formula, until we see whether Forever changed it |

### Recovery

Eating and drinking happen together. We stand up when both bars are full or both buffs run out.
The sit thresholds are APL knobs the optimizer tunes. The defaults are mana below 40% or HP
below 50%.

The best vendor food and water for our level (amounts and durations from
`assets/db_inputs/wowhead_item_tooltips.csv`):

| Level | Water | Food |
|---|---|---|
| 1 | Refreshing Spring Water, 151.2 mana over 18 s | Tough Jerky, 61.2 health over 18 s |
| 5 | Ice Cold Milk, 436.8 over 21 s | Haunch of Meat, 243.6 over 21 s |
| 15 | Melon Juice, 835.2 over 24 s | Mutton Chop, 552 over 24 s |
| 25 | Sweet Nectar, 1344.6 over 27 s | Wild Hog Shank, 874.8 over 27 s |
| 35 | Moonberry Juice, 1992 over 30 s | Cured Ham Steak, 1392 over 30 s |
| 45 | Morning Glory Dew, 2934 over 30 s | Roasted Quail, 2148 over 30 s |

Bandages are on by default: the best cloth for our level, assuming First Aid is kept up. The
8 s channel and the 60 s Recently Bandaged debuff are modelled.

### Gear

| Setting | Default |
|---|---|
| Item level | Each slot lags our level by a random 0–10 levels (5 on average) |
| Quality | 80% green, 10% white, 10% dungeon blue |
| Stat efficiency | 65% of the budget on stats the spec uses |
| Weapon | Its own weapon DPS curve by level |

## What the engine has today

- **Mob auto-attacks against the player.** A target with `tank_index` set to us swings at us
  with its base damage, spread, speed and school (`sim/core/target_ai.go`). Several targets
  can be on us at once.
- **The full attack table for a mob hitting us:** miss, dodge, parry, block, crit and crushing
  by level difference, plus armor (`sim/core/target.go`).
- **A health bar.** Damage taken comes off it and death is recorded, but the fight goes on
  (`sim/core/health.go`). Every iteration starts at full health. Death is recorded whenever a
  mob is on us, with or without a healing model (since groundwork 1).
- **Spell pushback**, including talents that reduce it (`sim/core/cast.go`): 1.0 s for the
  first hit, then 0.8, 0.6, 0.4 and 0.2 s, as on the cmangos Classic server. Since groundwork 1
  it can't take the cast bar back past its start, as there. Forever may differ, which logs of
  casts under mob hits would show.
- **No dodge, parry or block while casting**, as on the cmangos Classic server.
- **"In front of target"** as a player option (`proto/api.proto`, `in_front_of_target`).
- **A chosen spell rank** in action IDs. The Elemental rotation already casts Lightning Bolt
  rank 7 at low mana.
- **Healing math:** coefficients, crits, HoT snapshots and metrics
  (`sim/core/spell_result.go`, `CalcHealing`). There are APL values for current health and
  health percent.
- **Self-heals that exist for DPS reasons:** Drain Life, Siphon Life, Death Coil, Vampiric
  Embrace and Lay on Hands. Life Tap costs health.
- **A health-based fight end** (`Encounter.use_health`). Since groundwork 2 each target with
  health dies on its own (`sim/core/target_death.go`), and the fight ends when the last one
  does. Until groundwork 1 such a fight reported a 600 s length (the placeholder before the
  presim), so its DPS and damage taken per second were wrong. It now reports the time the
  kill took.
- **The normal mob table** and a target built from it (`core.NormalMobTarget`), since
  groundwork 1.
- **Everything below level 60:** per-level base stats, ranks by level and level-difference
  hit tables.
- **rotopt:** knob search, talent hill-climb, gear search and stat weights.

## Groundwork

These come first. Each one stands on its own in today's DPS mode, so it can be checked before
anything leveling-specific exists.

1. **Damage taken.**
   - Done (2026-10-15):
     - The normal mob table by level, from the 1.12 world database (above).
     - Death is recorded whenever a mob is on us, with no healing model needed.
     - A health fight reports the time the kill took, not 600 s.
     - Pushback can't take a cast back past its start, as on the cmangos Classic server.
     - Tests of a mob's swings on us: outcomes by level difference, hit damage after armor,
       a full solo fight and a pack that kills us (`sim/shaman/enhancement/normal_mob_test.go`),
       and pushback (`pushback_test.go`).
     - `logcheck -mobs`, to measure the same things from the logs.
   - Open:
     - Run `logcheck -mobs` on the user's machine, and compare Forever's mobs with the table.
     - Check the player's defenses at low level (dodge per Agility for every class, block,
       armor) against the beta's character sheet. notes.md has dodge per Agility checked at
       level 1 for some classes only.
     - Check the pushback rule in Forever's logs (casts under mob hits).
     - A "solo mob" encounter preset in the UI. Until the leveling mode has its own UI, a
       target from the table with "use health" on does it by hand, and the results show
       damage taken per second and the chance of death.
2. **Separate mob death.** Done (2026-10-15).
   - In a health fight each mob has its own health. When it runs out, after the killing
     blow's procs, the mob dies (`sim/core/target_death.go`): it stops swinging and casting,
     its auras and our debuffs and dots on it end, and it leaves
     `Encounter.ActiveTargetUnits`. Whoever had it as their target moves on to the next living
     one, and the fight ends with the last kill. Overkill doesn't count toward the damage the
     mobs took.
   - Nothing lands on a dead mob: the core drops damage to it, and every AoE in the sim takes
     its targets from the living ones.
   - AoE kinds (`sim/core/aoe.go`): hard capped (`Encounter.AoETargetCount(N)`), soft capped
     (`AoESoftCapMultiplier`, split or square root past N) and unlimited (every living
     target). Which kind each ability is in Forever is the Forever mechanics auditor's AoE
     scaling check (orchestration/auditor_forever.md, orchestration/aoe_scaling.md). Until
     then the sim keeps what it had: the hard caps in each spell, the 10 target split on
     Explosive Trap and one explosive, and full damage for the rest.
   - The shaman's fire totems run on the shaman, not on the first mob, so they keep going
     when it dies, and Searing Totem shoots our current target.
   - Timed fights never kill a mob, and with the totems as they were every golden test result
     came out the same to the last digit. Moving the totems onto the shaman reorders events
     that happen at the same moment, which moves the shaman results by up to 0.025 DPS (under
     0.01%, within noise). The shaman golden results were updated for it.
   - Tests: `sim/shaman/enhancement/mob_death_test.go` (a pack dying one by one, Chain
     Lightning and Fire Nova on the one mob left, Searing Totem moving on) and
     `sim/core/aoe_test.go`.
   - A cast at a mob that dies during it stops there, as in game: no mana spent, no effect,
     its cooldown given back, and we can act again when the GCD from its start runs out.
3. **Healing.**
   - Shaman, done (2026-10-15, `sim/shaman/healing.go`, shaman_audit.md 11.5): Healing Wave,
     Lesser Healing Wave, Chain Heal and Riptide from the Forever client, every heal landing on
     us, and the healing talents (Improved Healing Wave, Healing Way, Purification, Healing
     Focus, Ancestral Healing, Tidal Focus, Tidal Mastery, Nature's Swiftness, Water Shield,
     Mental Quickness). A rotation can heal on `currentHealthPercent`. A naked level 30 dies
     to one level 30 mob 46% of the time with white swings only, and never with Lesser Healing
     Wave below 40% health.
   - Still to do, other classes:
     - Priest: Lesser Heal, Heal, Flash Heal, Greater Heal and Renew (today commented out or
       missing).
     - Paladin: Holy Light and Flash of Light.
     - Druid: Healing Touch, Rejuvenation and Regrowth.
   - Absorbs in the core, for Power Word: Shield, Mana Shield and Sacrifice.

## The leveling mode

After the groundwork:

1. **The Grind encounter.** Done (2026-10-15), `sim/core/leveling.go` and
   `leveling_tables.go`, set by `Encounter.leveling` (`LevelingOptions` in
   `proto/common.proto`), results in `EncounterMetrics.leveling` (`LevelingMetrics`).
   - The mobs: a pool of max pack size normal mobs at each level of the spread, all of one
     creature type (default Beast), reused as they die. A mob's level and type are fixed when
     the sim is built, because its attack tables and the talents and items that read its type
     are.
   - The loop: search (random 5 to 11 sec), pull, run-in from the pull range at the mob's
     level's run speed, chain pulls every 2 to 4 sec up to the pack size, fight, then a
     bandage, eating and drinking when health or mana is low, and search again. While a mob
     runs in we can cast at it but not reach it in melee, and it can't hit us. Its first
     swing comes at a random point of its swing timer. Death: the mobs leave, a 3 min corpse
     run, back at 50% health and mana.
   - Out of combat we regen health from Spirit by the 1.12 server's class formulas, half
     again while sitting (notes.md, Need to Verify). Food, water and bandages come from the
     item data by level. XP per kill is the 1.12 formula by level difference.
   - The pull and rest decisions were options (pack size, eat_below, drink_below). Step 2
     moved how we pull into the rotation; the rest is step 3.
   - Results: mobs killed (the one we're on at the end by the share of its health we took,
     with its standard deviation), XP and XP per hour, deaths, the time in combat, resting,
     searching and dead, and mobs per pull.
   - Found on the way: an AoE that killed a mob changed the living list while looping over
     it, and could crash. Adding and removing a living target now copies the list.
   - First numbers, the Level 30 Enhancement preset (gear, talents and rotation, 200 sessions
     of 20 min against level 29 to 31 mobs):

     | Pack size | Mobs | Deaths | XP / hour | Combat | Resting | Searching | Dead |
     |---|---|---|---|---|---|---|---|
     | 1 | 75.5 | 0.0 | 43,700 | 471 s | 120 s | 608 s | 1 s |
     | 2 | 89.6 | 0.3 | 51,900 | 567 s | 204 s | 382 s | 48 s |
     | 3 | 82.5 | 1.4 | 47,700 | 513 s | 209 s | 244 s | 235 s |

     Two at a time comes out best. At three, about 1.4 deaths a session eat the gain. The
     preset's rotation was built for Vishas, so it doesn't pull, heal or chain pull on
     purpose yet.
   - Not modelled yet: a mob's type is one per session.
2. **Pulling in the APL.** Done (2026-10-15), positions in `sim/core/leveling.go`.
   - Where things are: each mob has a distance from us, and we have a position along the way
     we run. The mob we find stands 10 to 40 yards away; a pulled one runs in from where it
     stood; one on us is in melee range. We run while we search, and walk up to a mob the APL
     doesn't pull, and while we run no hard cast goes out (`Moving`).
   - Spell ranges from the client's SpellRange (`Spell.MaxRange`, held to only in a leveling
     session): shocks 20 yards, Lightning Bolt, Chain Lightning and Lava Burst 30.
   - Totems stay where we drop them. Searing Totem shoots a mob within 20 yards of it, Magma
     Totem's pulse reaches 8 yards around it, and Fire Nova 10 yards around it, castable only
     within 30 yards of it. A totem dropped while we search is left behind. Outside a leveling
     session nothing moves, so other fights are unchanged.
   - The APL: `PullTarget` (a `UnitReference`) is the mob we found and haven't pulled, none
     once a spell on its way pulls it. Values: Distance (Leveling), Mobs Running In, Mobs On
     You, Is Moving. A spell that attacks the mob pulls it when it lands; a totem or a shield
     cast with the mob as our target doesn't.
   - First numbers, the Level 30 Enhancement preset with pull lines added in front of its
     rotation (500 sessions of 20 min against level 29 to 31 mobs). Its own Flame Shock and
     Earth Shock pull a mob within 20 yards, and we walk up to one further away:

     | Rotation | Pack size | Mobs | Deaths | XP / hour | Searching and pulling | Resting | Combat | Mobs a pull |
     |---|---|---|---|---|---|---|---|---|
     | Preset | 1 | 73.8 | 0.03 | 42,700 | 717 s | 65 s | 412 s | 1.00 |
     | + Lightning Bolt rank 1 at 20–30 yd | 1 | 71.9 | 0.02 | 41,600 | 720 s | 64 s | 414 s | 1.00 |
     | + Earth Shock rank 1 within 20 yd, Lightning Bolt rank 1 at 20–30 yd | 1 | 73.3 | 0.02 | 42,400 | 736 s | 24 s | 436 s | 1.00 |
     | Preset | 2 | 88.7 | 0.60 | 51,300 | 373 s | 185 s | 535 s | 1.33 |
     | + Lightning Bolt rank 1 at 20–30 yd, no mob on us | 2 | 86.0 | 0.61 | 49,700 | 380 s | 181 s | 530 s | 1.36 |
     | + Lightning Bolt rank 1 at 20–30 yd, chain pulls too | 2 | 81.9 | 0.61 | 47,300 | 387 s | 176 s | 530 s | 1.47 |

     A bolt pull loses time: we stand still for the 1.5 s cast and the bolt's flight, where
     walking 5 to 10 yards closer and shocking gets the mob to us sooner. A rank 1 shock pull
     saves mana and so rest time, but its cooldown is shared with the preset's Flame Shock.
     With a pack size of 2 the second mob often isn't found before the first dies, so a pull
     takes 1.3 to 1.5 mobs on average. The mob found then is the next pull, without a search.
   - Since then (step 3) the APL decides on chain pulls and the mob notices us from its aggro
     radius, so these numbers are out of date.
3. **Chain pulls, aggro radius and totems in the APL.** Done (2026-10-15).
   - Positions are on a line (`sim/core/leveling_world.go`): ours, each mob's and each totem's.
     A pulled mob chases us at its speed and changes course when we move, so it can take
     longer to reach us when we walk away.
   - A mob notices us within its aggro radius (above). A face pull stops there, and the mob
     runs the rest of the way.
   - The APL decides when to look for another mob while fighting (Look For Mob (Leveling)),
     and how to pull it. `pack_size` is now only a fixed policy for a rotation without that
     (default 1, none). Walk To Pull Target (Leveling) walks up to the mob, all the way or into
     a spell's range.
   - Buff totems reach 30 yards (Strength of Earth, Stoneskin, Grace of Air, Windfury,
     Flametongue and Mana Spring Totem): further away we lose the buff, and get it back when
     we come back while the totem stands. Healing Stream and Mana Tide give nothing on a pulse
     out of reach. Our totems vanish when we die.
   - Totemic Recall (36936, level 20): takes every totem back for a quarter of its mana cost,
     a 1.5 s global cooldown, no cost. Totemic Projection (437009, level 22): moves every totem
     to where we stand, 25% of base mana (Totemic Focus discounts it), 1 min cooldown. Both
     from the client data, build 70338.
   - APL: Walk To Totem (Leveling) walks back toward a totem with the mobs following, and
     Totem Distance (Leveling) is how far one is.
   - The leveling demo is solo now: the Level 30 Enhancement preset's gear, talents and
     rotation without party or raid buffs, Blessing of Kings, starting totems, explosives or
     potions. 500 sessions of 20 min against level 29 to 31 mobs:

     | Rotation | Mobs | Deaths | XP / hour | Searching and pulling | Resting | Combat | Mobs a pull |
     |---|---|---|---|---|---|---|---|
     | Preset: a totem whenever its buff is missing | 55.7 | 0.00 | 32,200 | 536 s | 317 s | 347 s | 1.00 |
     | Buff totems only with a mob pulled | 56.8 | 0.00 | 32,900 | 549 s | 261 s | 391 s | 1.00 |
     | + Totemic Recall once the fight is over | 58.2 | 0.00 | 33,700 | 561 s | 239 s | 400 s | 1.00 |
     | + Totemic Projection instead of Recall, when a totem is out of reach | 57.1 | 0.00 | 33,100 | 550 s | 266 s | 384 s | 1.00 |
     | No buff totems | 65.1 | 0.00 | 37,700 | 628 s | 166 s | 406 s | 1.00 |
     | Recall, + look for another mob above 60% health and 40% mana while fewer than 2 are pulled | 77.6 | 0.15 | 44,900 | 226 s | 386 s | 563 s | 2.50 |
     | No buff totems, + the same chain pulls | 82.2 | 0.31 | 47,600 | 238 s | 323 s | 587 s | 2.51 |

     At level 30, single pulls die in about 25 s, and the totems' mana (Strength of Earth 65,
     Mana Spring 40, Flametongue 90 at these ranks) costs more rest than their buffs save.
     Recalling them when the fight is over gets a quarter back and is better than leaving them
     behind. The rotation-driven chain pulls are the big gain: searching drops from about 600 s
     to 230 s a session, since we find the next mob while fighting. A pull averages 2.5 mobs,
     because a mob found when the first dies is pulled too, as long as health and mana allow.
     The preset rotation was built for a boss, so none of this is tuned yet.
   - Still options: when to eat, drink and bandage. Step 4 moves them into the APL.
4. **APL additions for the rest of the loop:** out-of-combat actions (eat and drink until a
   threshold, bandage, heal).
5. **Enhancement shaman end to end at level 30**, checked against the user's beta logs.
6. **Optimizers:** the pull and recovery policy, leveling stat weights with a Pawn export, and
   the best mob level.
7. **Synthetic gear curves** from 10 to 60 and the talent route search.
8. **UI:** "Leveling" as a third mode next to DPS and PvP.

## From the logs, later

A small tool run on the user's machine against `~/forever-logs` measures, for each kill:

- the time from a kill to the next pull, minus time spent eating or drinking
- the run-in time, from the pull to the mob's first swing
- the gap between chain-pulled mobs
- the pack sizes actually taken

The logs have AFK and unfocused stretches, so it drops gaps over about 60 s and reports the
median and the fast quartile. The fast quartile becomes the default, since we're simulating
focused play. Those numbers replace the search and run-in defaults above.
