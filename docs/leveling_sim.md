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
   mana regen outside the five-second rule) ticks during it. Positions are on a flat 2D plane:
   each search we run off in a random direction, and the mob we find stands a random 10 to 40
   yards away in a random direction from us.
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
   or searching for the next pull right away. A healer with spare mana can heal instead of
   eating, since healing turns mana into HP. Only the APL rests: no setting sits us down.
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
table.

Health is Forever's, not 1.12's. Our beta logs (`logcheck -mobs`, 2026-10-16) show that Forever
gives every normal melee mob of a level the same health. It's the 1.12 warrior class mob's
health at that level, 2.5% more a level from level 9, and 30% more from level 20 on. Needles
Cougar has 896 health at level 28 in 1.12 and 1164 in Forever, and so does every other level 28
mob we fought, the paladin class ones too. The rule fits all 40 creatures we saw from level 2
to 35 to within a point, except at level 35: the table has 1824 there, and the one mob we saw
had 1745. Above 35 it's the same rule, not yet checked. At level 30 a mob has 1365 health where
the 1.12 average was 1019, so every kill takes about a third longer.

A few rows:

| Level | Health | Armor | Damage | Swing | Run speed |
|---|---|---|---|---|---|
| 10 | 208 | 531 | 14.1–18.4 | 1.90 s | 7.5 yd/s |
| 20 | 642 | 924 | 32.0–39.2 | 1.99 s | 8.1 yd/s |
| 30 | 1365 | 1288 | 45.8–55.2 | 1.94 s | 8.1 yd/s |
| 40 | 2378 | 2189 | 66.7–82.0 | 1.96 s | 8.3 yd/s |
| 50 | 3600 | 3123 | 89.9–110.3 | 1.93 s | 8.2 yd/s |
| 60 | 5356 | 3728 | 120.2–148.6 | 1.92 s | 8.0 yd/s |

Damage is a white swing before our armor, with the mob's attack power in it (AP / 14 per second
of swing). `go run --tags=with_db ./tools/logcheck -mobs` measures the same things from our
logs, so we can see where Forever's mobs differ. Besides health, the logs agree with the table on
armor and on the 2 s swing. The damage of a landed swing is within the table's range at most
levels, and under it at 27 and 28 (the Thousand Needles mobs hit softer).

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

The settings (`LevelingOptions`) describe the world: the session's length, the mobs (level,
spread, type, armor, the most that can be on us at once), how long searches take and how far
the mob we find stands, how much of a chain search we run, and the death penalty. Every choice
the character makes is the rotation's: how many mobs to pull (Look For Mob, with Mobs This
Pull to stop at a pack size), how to pull them, and when to eat, drink, bandage or heal. Eat
Below, Drink Below, Bandages and Pack Size were settings until 2026-10-15. Old saved settings
with them still load, and those values are ignored.

### Session

| Setting | Default |
|---|---|
| Session length | 20 min |
| Start state | Full HP and mana, all cooldowns ready |
| Death | 3 min corpse run, back at 50% HP and mana. Cooldowns keep ticking during the run. |

### Pulling

| Setting | Default |
|---|---|
| Finding the next mob, out of combat | Random 2–7 s (4.5 s on average), running. From the logs (below), 5–11 s until 2026-10-16. |
| Finding a mob to chain-pull, in combat | Random 2–4 s (3 s on average), running 3 times in 4 |
| Where the mob we find stands | Random 10–40 yd in whole yards, in a random direction |
| Searching | Running in a random direction at our run speed |
| Our run speed | 7 yd/s |
| Mob run speed | The level's average from the mob table: 6 yd/s at levels 1–2, about 8 yd/s from level 11 |
| Run-in time | (distance − 5 yd melee) ÷ run speed, from when the pull spell lands: about 3.1 s for Lightning Bolt from 30 yd (after 1.5 s of flight) and 1.9 s for a shock from 20 yd at 8 yd/s |
| Aggro radius | 20 yd for a mob of our level, ±1 yd a level of difference, at least 5 (cmangos `Creature::GetAttackDistance`) |
| Face pull | No free damage. We walk up until the mob notices us. Out of combat, when the APL hasn't pulled a mob, isn't casting and hasn't moved for 1 s, we walk up to it. |
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
| HP, armor, damage, swing and run speed by level | The mob table (above): health from the logs, the rest from 1.12 |
| XP per kill | Classic's formula, until we see whether Forever changed it |

### Recovery

The APL decides when to sit down to eat, drink, both or neither, and to what share of health
and mana to rest (Rest (Leveling)), and when to bandage (Bandage (Leveling), with Bandage
Ready). A rotation with no Rest or Bandage line, like one written for a single fight, gets two
default lines in front of its own: bandage below 50% health, then eat and drink to full below
50% health or 40% mana, both only while we stand (not while searching). A rotation that should
never rest has a Rest line switched off, which keeps the defaults out.

Spirit regen goes on top of food and water, for health and mana alike:

| Regen | Model |
|---|---|
| Health from Spirit, out of combat | Wrath's game tables: a 2 s tick of the first 50 Spirit times a class and level ratio, plus the rest times a lower ratio. At level 30 a warrior gets 0.56 a Spirit a tick, a rogue 0.33, a shaman 0.17, a mage 0.10. A Forever beta measurement fits the shape but not every value (notes.md, Need to Verify). |
| Sitting (eating, drinking or just resting) | A third more health from Spirit, for every class (the beta measurement saw about 30%). Mana regen isn't changed by sitting. |
| Mana from Spirit | The sim's Forever formula, outside the five-second rule (notes.md). |
| Troll Regeneration | 10% more health regen, and 10% of it in combat. |
| In combat | No health regen from Spirit (Troll aside). |

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

A bandage is the best cloth for our level, assuming First Aid is kept up. The 8 s channel and
the 60 s Recently Bandaged debuff are modelled.

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
     - Done (2026-10-16): `logcheck -mobs` on the user's logs. Forever's mobs have up to 30%
       more health than the table had, which the table now uses (Mobs above). Armor and swing
       time match. Mobs also crit us far less than Classic's 5%: 4 crits in about 2,200
       swings at us from mobs 6 levels under us to 5 over, where crushing blows came as
       expected (13% at 3 levels over, 26% at 4). The sim still gives mobs Classic's crit
       (notes.md, Need to Verify).
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
   - Where things are (on a line at first, a 2D plane since step 3): each mob has a distance
     from us, and we have a position along the way we run. The mob we find stands 10 to 40 yards away; a pulled one runs in from where it
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
   - Positions are on a flat 2D plane (`sim/core/leveling_world.go`): ours, each mob's and
     each totem's. Each search runs off in a random direction, and the mob we find stands in a
     random direction from us. A pulled mob runs straight at us, and while we move it turns
     toward us every 0.1 s, so it takes longer to reach us when we walk away. A mob notices us
     the moment our path comes within its aggro radius, and a totem's reach and a Magma
     Totem's pulse are measured from where it stands. (The first version of step 3 had
     everything on one line, with every mob ahead of us; the table below is from the plane.)
   - A mob notices us within its aggro radius (above). A face pull stops there, and the mob
     runs the rest of the way.
   - The APL decides when to look for another mob while fighting (Look For Mob (Leveling)),
     and how to pull it. `pack_size` was left as a fixed policy for a rotation without that
     (default 1, none), until step 10 took it out. Walk To Pull Target (Leveling) walks up to the mob, all the way or into
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
     | Preset: a totem whenever its buff is missing | 57.7 | 0.00 | 33,400 | 556 s | 285 s | 360 s | 1.00 |
     | Buff totems only with a mob pulled | 56.9 | 0.00 | 32,900 | 549 s | 261 s | 390 s | 1.00 |
     | + Totemic Recall once the fight is over | 58.1 | 0.00 | 33,600 | 561 s | 239 s | 400 s | 1.00 |
     | + Totemic Projection instead of Recall, when a totem is out of reach | 57.3 | 0.00 | 33,200 | 553 s | 265 s | 382 s | 1.00 |
     | No buff totems | 65.0 | 0.00 | 37,600 | 627 s | 166 s | 407 s | 1.00 |
     | Recall, + look for another mob above 60% health and 40% mana while fewer than 2 are pulled | 77.9 | 0.18 | 45,000 | 216 s | 369 s | 584 s | 2.67 |
     | No buff totems, + the same chain pulls | 79.7 | 0.31 | 46,100 | 216 s | 322 s | 608 s | 2.79 |

     At level 30, single pulls die in about 25 s, and the totems' mana (Strength of Earth 65,
     Mana Spring 40, Flametongue 90 at these ranks) costs more rest than their buffs save.
     Recalling them when the fight is over gets a quarter back and is better than leaving them
     behind. The rotation-driven chain pulls are the big gain: searching drops from about 600 s
     to 220 s a session, since we find the next mob while fighting. A pull averages 2.7 mobs,
     because a mob found when the first dies is pulled too, as long as health and mana allow.
     The preset rotation was built for a boss, so none of this is tuned yet.
   - Still options: when to eat, drink and bandage. Step 4 moves them into the APL.
4. **Resting in the APL.** Done (2026-10-15), `sim/core/leveling_rest.go`.
   - When the last mob dies we're ready: the APL can sit down (Rest (Leveling): eat, drink,
     both or neither, until health and mana are at a share of max, with another food or
     drink when one runs out short), bandage (Bandage (Leveling), Bandage Ready), heal or buff.
     Once it isn't casting and its global cooldown is up, we search. It can also rest while we
     search or with a mob found. A rotation with neither action got the fixed policy, until
     step 10 put default lines in its place.
   - Health regen from Spirit is now Wrath's game tables, a third more sitting, with Troll
     Regeneration (Recovery above). The 1.12 formulas the sim had gave a level 30 shaman with
     50 Spirit 10.4 health a second; the tables give 4.2 (5.6 sitting).
   - Food, water and bandages used are in the results, for the gold they cost later.
   - The solo demo with the "no buff totems, look for another mob" rotation, different rest
     policies (500 sessions):

     | Rest policy | Mobs | Deaths | XP / hour | Searching and pulling | Resting | Combat | Mobs a pull |
     |---|---|---|---|---|---|---|---|
     | Fixed: eat below 50% health, drink below 40% mana | 78.9 | 0.35 | 45,600 | 220 s | 320 s | 601 s | 2.74 |
     | Rotation: the same thresholds, until both are full | 78.5 | 0.31 | 45,400 | 231 s | 322 s | 595 s | 2.77 |
     | Rotation: the same thresholds, until 80% | 80.5 | 0.34 | 46,500 | 270 s | 265 s | 608 s | 2.48 |
     | Rotation: rest earlier, below 70% health or 60% mana | 79.1 | 0.29 | 45,800 | 216 s | 332 s | 603 s | 2.90 |
     | Rotation: bandage below 70% when ready, eat and drink as above | 77.9 | 0.32 | 45,100 | 222 s | 330 s | 594 s | 2.85 |
     | Rotation: Lesser Healing Wave below 60% out of combat, drink below 40% | 72.1 | 0.39 | 41,700 | 307 s | 287 s | 542 s | 2.30 |

     The rest policies land within 2 mobs of each other. Healing between pulls costs mana
     that the next pull needs, and the drinking it takes is slower than eating.
5. **The leveling optimizer.** Done (2026-10-15).
   - `tools/rotopt` scores leveling settings by the mobs killed in a session (`-objective xp`
     for XP an hour, to compare mob levels), with `-mob-offset`, and `-session` boils one
     session down to its searches, pulls, fights and rests with health, mana, casts and times.
     The `shaman_enh_leveling` template has the rules and their knobs: rest, bandage and heal
     thresholds, the pull spell, chain pulls and look ahead, totems and Totemic Recall, the
     fight's shocks, Fire Nova, Lightning Shield and heal. Settings for the solo Level 30
     Enhancement shaman are in `tools/rotopt/leveling/level30_enh_solo.json`.
   - Some choices are hard to put into fixed rules, so an auditor agent
     (`orchestration/auditor_leveling.md`) reads sample sessions, tests rule changes, and the
     ones that hold up go into the template for the optimizer to tune. The skill
     `.agents/skills/leveling-opt` runs the loop, and `orchestration/leveling_audit_log.md`
     keeps what each round tried.
   - Level 30 solo, against level 29 to 31 mobs (4000 sessions):

     | Rotation | Mobs | Deaths | XP / hour | Searching and pulling | Resting | Combat | Mobs a pull |
     |---|---|---|---|---|---|---|---|
     | Template defaults before tuning | 79.5 | 0.31 | about 46,000 | 237 s | 318 s | 596 s | 2.42 |
     | Optimizer pass 1 | 90.2 | 0.01 | 52,200 | 235 s | 300 s | 664 s | 2.54 |
     | Audit round 1, look ahead (pass 2 kept every knob) | 97.8 | 0.01 | 56,600 | 144 s | 344 s | 711 s | 2.19 |
     | Audit round 2: Mana Spring, rank 1 Searing, rest mana to 80%, chain as the target dies | 99.5 | 0.01 | 57,600 | 141 s | 338 s | 719 s | 2.32 |
     | Optimizer pass 3: pull with rank 1 Earth Shock | 100.0 | 0.01 | 57,900 | 140 s | 327 s | 732 s | 2.13 |
     | Pass 3's rotation with one imbue (Windfury Weapon) | 91.9 | 0.02 | - | 106 s | 340 s | 750 s | 2.21 |
     | Optimizer pass 4: Fire Nova with two mobs on us | 93.0 | 0.03 | 53,800 | 110 s | 378 s | 707 s | 2.30 |
     | Pass 4's rotation with Forever's mob health (2026-10-16) | 71.8 | 0.22 | - | 72 s | 381 s | 709 s | 2.37 |
     | + the 2 to 7 s search from the logs | 72.4 | 0.22 | 41,900 | 63 s | 386 s | 714 s | 2.37 |
     | Optimizer pass 5: look ahead above 60% health, rest health to full | 73.8 | 0.00 | 42,700 | 97 s | 390 s | 712 s | 2.33 |

     The rows down to pass 3 ran with Rockbiter Weapon and Windfury Weapon both on the main
     hand: the settings had Rockbiter in the consumes and Windfury in the class options, and
     the sim put both on. The game takes one shaman imbue, so now the class options' imbue
     wins, and the settings have Windfury alone (it kills more than Rockbiter alone, 89.6 to
     86.3 mobs with the round 2 rotation). The rules the earlier rows found still stand; pass 4
     re-tuned the knobs with one imbue and added Fire Nova with two mobs on us (+1.1 mobs on
     both seeds), which the weaker melee now makes worth its mana.

     The last three rows have Forever's mob health and search time (From the logs, below).
     The health costs 21 mobs a session (93.0 to 71.8), since each kill takes a third longer.
     The shorter search gives back only 0.5, because the rotation mostly finds its next mob
     while still fighting. Pass 5 kept every rule and moved two knobs: look for the next mob
     above 60% health instead of 40%, and rest health to full instead of 90%. With mobs that
     take longer to kill, looking ahead at 40% to 60% health led into pulls that killed us
     (0.22 deaths a session). It's +1.3 mobs on seed 1 and on seed 1000001 (72.55 to 73.84),
     almost all from the look ahead.

     Pass 1: sit down below 80% health and stand at 90% instead of 50% and 100%, no Flame
     Shock DoT in the fight, Lesser Healing Wave below 30% in combat, no buff totems (their
     mana costs more rest than they save in 20 to 30 sec fights).

     Audit round 1 found an engine bug (a swing could pull the mob we'd found from 40 yards
     when the last mob died) and the biggest rule so far: look for the next mob while still
     fighting, at any health over 40%, but pull it only when nearly full (over 90% health and
     50% mana); otherwise rest with it waiting, and it's the next pull with no search.
     Searching drops from 235 to 144 sec a session.

     Audit round 2 saw that rests wait on mana, not health (drinking is slower than eating),
     so its gains make mana cheaper: Mana Spring Totem in combat with no Totemic Recall (the
     recall's global cooldown costs more than its refund gives back), rank 1 Searing Totem and
     only with two mobs pulled, resting mana to 80%, and pulling the next mob as the current
     one dies when we're healthy. It also found that our own lower-rank Strength of Earth and
     Stoneskin totems gave the top rank's buff, now fixed. Pass 3 then pulled with rank 1
     Earth Shock instead of walking in (+0.5 mobs on seed 1, +0.4 on a second seed).
   - The best mob level, XP an hour with the pass 4 rotation (2000 sessions): 49,600 at 3
     levels under us, 52,800 at 2 under, 54,300 at 1 under, 53,700 at our level, 51,300 at 1
     over, 41,900 at 2 over (deaths climb to 0.7 a session). Mobs one level under us are best by
     about 1%, with mobs of our level next; the knobs were tuned at our level.
   - Again with Forever's mob health and the pass 5 rotation (2000 sessions, 2026-10-16):
     43,200 at 3 levels under us, 44,200 at 2 under, 43,900 at 1 under, 42,700 at our level and
     40,500 at 1 over. Mobs two levels under us are now best, by 3% over our own level. That's
     also what the user fought in the 2026-10-10 log (level 27 to 29 at 30). Tuning the knobs
     at 2 under only adds half a mob (90.5 to 91.0: rest mana to 60% instead of 80%), so the
     preset keeps the knobs tuned at our level, and the default mob level stays ours.
6. **Enhancement shaman end to end at level 30**, checked against the user's beta logs.
7. **More optimizing:** leveling stat weights with a Pawn export, the talent route, other
   classes' templates.
8. **Synthetic gear curves** from 10 to 60 and the talent route search.
9. **UI**, done: "Leveling" is an Encounter Type next to PvP in the individual sims, with the
   session's options (the fight-only inputs hide). Results lead with mobs killed, XP an hour,
   deaths and mobs a pull, and the time split; the history and Compare tab compare mobs
   killed; stat weights are off for now. Enhancement shaman has a "Leveling" preset group
   with the solo Level 30 build and this rotation (`apls/level30_leveling.apl.json`).

10. **Every choice in the rotation.** Done (2026-10-15).
    - Eat Below, Drink Below, Bandages and Pack Size are gone from the settings (their proto
      fields are reserved, so old settings still load). A session only eats, drinks,
      bandages and chain pulls when the rotation says so.
    - New APL value: Mobs This Pull (Leveling), the mobs the current pull has taken, the dead
      ones too, 0 out of combat. Look For Mob while it's under 3 pulls three at a time, which
      is what Pack Size 3 did. Mobs On You plus Mobs Running In can't say that, since they
      drop as mobs die.
    - A rotation with no Rest or Bandage line gets default ones in front (Recovery above): the
      old settings' 50% and 40%, eating and drinking until full. The old policy stood up when
      both bars were full or the first food and drink ran out; the default lines eat another
      when one runs out short, a small difference (the naked level 30 test shaman kills 23.6
      mobs a session with them, 23.2 before).
    - Two fallbacks stay in the session, because they keep it going rather than choose
      anything: out of combat, a mob we found that the rotation hasn't pulled after a second
      of standing idle gets walked up to (a rotation pulls it, walks up to it or rests within
      that second); and when we're ready between pulls, we search after 10 sec even if the
      rotation keeps acting (buffing every GCD, say).
    - The Level 30 Leveling preset and the optimizer template already rested, bandaged and
      chain pulled through the rotation, so their score is unchanged: 92.98 mobs (4000
      sessions) before and after.

## From the logs

`go run --tags=with_db ./tools/logcheck -pacing` (2026-10-16) reads `~/forever-logs` and
measures our grinding, for mobs within 3 levels of ours that we killed:

- the time from a kill to the next pull, less the time spent eating, drinking, bandaging or
  casting a heal
- how we pulled, from how far, and how long the mob took to reach us
- how many mobs a fight took, and how the later ones joined
- how far we traveled between fights, and whether we cast Ghost Wolf on the way

It drops gaps over 60 s (AFK) and gaps with a death in them. `-from` and `-to` pick a stretch.

Most of what the archive holds isn't plain grinding. The level 30 shaman's log of 2026-10-10
has about 200 kills of level 27 to 31 mobs, but from 20:00 on it's beta tests (white swings
and Flametongue Totem, about 55 damage a second). From 18:00 to 20:00 is normal play (110 to
125 damage a second, Stormstrike and shocks). The 2026-10-11 log is that shaman killing mobs 12
or more levels under it, and the PvE2 logs from 2026-10-14 have few lines of ours (BayovCore's
history has that character at level 1). So the
numbers come from one stretch of 52 fights, with the whole archive next to them:

| | 18:00 to 20:00 | All of the archive |
|---|---|---|
| Fights | 52 | 202 |
| Kill to next pull, fast quartile | 5.7 s | 6.5 s |
| Kill to next pull, median | 7.7 s | 10.5 s |
| Gaps with a rest | 6 of 40 (a heal of about 2 s, never food or water) | 29 of 140 |
| Travel in the gap, median | 39 yd | 49 yd |
| Pulled with Flame Shock | 46% | 31% |
| Walked up and swung first | 37% | 38% |
| The mob attacked first | 10% | 24% |
| Flame Shock pull distance, median | 18 yd | 19 yd |
| Run-in after a Flame Shock pull, median | 1.2 s | 1.6 s |
| Fights with one mob | 47 of 52 | 189 of 202 |

What it changed:

- **Search.** The sim's gap after a kill, when it had to search, was 8.5 s at the fast
  quartile and 9.6 s at the median with the 5 to 11 s search. The default search is now 2 to
  7 s, which gives 5.0 s and 6.4 s, about the 5.7 s fast quartile of focused play. The gap in
  the log has looting in it, which the sim doesn't model, so the search time stands for both.

What it didn't settle:

- **Run-in.** A Flame Shock pull from 18 yd reaches us in 1.2 s in the log. The sim gives 1.6 s
  at 8.1 yd/s, but we stand still in the sim while in the game we walk toward the mob. That fits
  the sim's mob speed, so it stays.
- **Chain search and pack sizes.** We took one mob a fight almost every time, and only 4 times
  in the normal stretch pulled a second mob while fighting (5 to 11 s after the one before).
  That's too few to measure the 2 to 4 s chain search. The sim's best rotation takes 2.3 mobs a
  pull, so how fast we find a mob while fighting matters a lot, and it's still a guess.
- **Found distance.** The log shows where we pulled from, not where we first saw the mob.
- **Rests.** We never ate or drank, and healed with Healing Wave instead. The sim's rotation
  rests about 390 s a session (eating and drinking). A grinding log with the sim's kind of
  pulls would show whether that holds.

A longer stretch of focused grinding, ideally with some chain pulls, would settle the chain
search and the rests.
