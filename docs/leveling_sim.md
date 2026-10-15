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

1. **Search.** It takes a random time to find the next mob. Out-of-combat regen (Spirit,
   mana regen outside the five-second rule) ticks during it.
2. **Pull.** The APL picks how to pull a mob that isn't in combat yet: any spell at any rank
   (Lightning Bolt rank 1 to save mana, Earth Shock, Flame Shock, …) or a face pull. The pull
   range sets the run-in time, which starts when the pull spell lands. Spells cast while the mob
   runs in are free damage, and the mob doesn't attack until it arrives.
3. **Chain pull.** While mobs are running in or already hitting us, the APL can pull another
   one, up to the pack cap. It decides from values like mobs in combat, mobs still running in,
   HP and mana percent, and cooldowns. Pushback makes a cast pull worse once a mob is hitting us,
   so the APL may chain-pull with an instant instead.
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
| Finding the next mob, out of combat | Random 5–11 s (8 s on average) |
| Finding a mob to chain-pull, in combat | Random 2–4 s (3 s on average) |
| Mob run speed | The level's average from the mob table: 6 yd/s at levels 1–2, about 8 yd/s from level 11 |
| Run-in time | (pull range − 5 yd melee) ÷ run speed, from when the pull spell lands: about 3.1 s for Lightning Bolt from 30 yd and 1.9 s for a shock from 20 yd at 8 yd/s |
| Face pull | No free damage, 1.5 s to walk up |
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

1. **The Grind encounter:** the session loop above, with run-in, the not-yet-in-combat mob
   state, state carried across pulls, eating, drinking and bandaging, the death penalty and the
   per-level XP table.
2. **APL additions:** pull and chain-pull actions; out-of-combat actions (eat and drink until a
   threshold, bandage, heal); values for mobs in combat, mobs running in and in-combat state.
3. **Enhancement shaman end to end at level 30**, checked against the user's beta logs.
4. **Optimizers:** the pull and recovery policy, leveling stat weights with a Pawn export, and
   the best mob level.
5. **Synthetic gear curves** from 10 to 60 and the talent route search.
6. **UI:** "Leveling" as a third mode next to DPS and PvP.

## From the logs, later

A small tool run on the user's machine against `~/forever-logs` measures, for each kill:

- the time from a kill to the next pull, minus time spent eating or drinking
- the run-in time, from the pull to the mob's first swing
- the gap between chain-pulled mobs
- the pack sizes actually taken

The logs have AFK and unfocused stretches, so it drops gaps over about 60 s and reports the
median and the fast quartile. The fast quartile becomes the default, since we're simulating
focused play. Those numbers replace the search and run-in defaults above.
