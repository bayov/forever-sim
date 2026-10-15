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
(creature type matters for Slaying-style talents and item bonuses). HP, armor, damage and swing
speed come from a per-level table.

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
| Mob run speed | 7 yd/s |
| Run-in time | (pull range − 5 yd melee) ÷ 7 yd/s, from when the pull spell lands: about 3.6 s for Lightning Bolt from 30 yd, 2.1 s for a shock from 20 yd |
| Face pull | No free damage, 1.5 s to walk up |
| A mob's first swing | Random within its swing timer after it arrives |
| Max pack size | 5 |

### Mobs

| Setting | Default |
|---|---|
| Target mob level | Our level |
| Spread | Uniform ±1 around the target level (target 20 gives mobs of level 19 to 21) |
| Armor tier | 25% low, 50% medium, 25% high |
| Creature type | 40% humanoid, 40% beast, 10% undead, 10% other |
| HP, damage and swing speed by level | Classic's averages for normal mobs, until the logs check them |
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
  (`sim/core/health.go`). Every iteration starts at full health.
- **Spell pushback**, including talents that reduce it (`sim/core/cast.go`). It has never fired
  in our sims because nothing hit us. The rule (1.0 s, then 0.2 s less per hit) needs checking
  against logs.
- **"In front of target"** as a player option (`proto/api.proto`, `in_front_of_target`).
- **A chosen spell rank** in action IDs. The Elemental rotation already casts Lightning Bolt
  rank 7 at low mana.
- **Healing math:** coefficients, crits, HoT snapshots and metrics
  (`sim/core/spell_result.go`, `CalcHealing`). There are APL values for current health and
  health percent.
- **Self-heals that exist for DPS reasons:** Drain Life, Siphon Life, Death Coil, Vampiric
  Embrace and Lay on Hands. Life Tap costs health.
- **A health-based fight end** (`Encounter.use_health`). It adds up the targets' HP and ends
  the fight when that total runs out, so mobs don't die separately.
- **Everything below level 60:** per-level base stats, ranks by level and level-difference
  hit tables.
- **rotopt:** knob search, talent hill-climb, gear search and stat weights.

## Groundwork

These come first. Each one stands on its own in today's DPS mode, so it can be checked before
anything leveling-specific exists.

1. **Damage taken.**
   - A table of normal-mob HP, armor, damage and swing speed by level.
   - Check the player's defenses at low level (dodge per Agility for every class, block,
     armor).
   - Check the pushback rule and front-facing parry and block.
   - Calibrate against the user's logs: damage taken per mob, and how often a mob's swings miss,
     get dodged or crit.
   - The DPS sim gets a "solo mob" encounter that reports HP lost per kill and deaths.
2. **Separate mob death.** Each mob has its own HP, dies when it runs out and stops attacking.
   The fight ends when all are dead.
3. **Healing.**
   - Self-heals by rank and level, checked against the client data:
     - Shaman: Healing Wave and Lesser Healing Wave.
     - Priest: Lesser Heal, Heal, Flash Heal, Greater Heal and Renew (today commented out or
       missing).
     - Paladin: Holy Light and Flash of Light.
     - Druid: Healing Touch, Rejuvenation and Regrowth.
   - Absorbs in the core, for Power Word: Shield, Mana Shield and Sacrifice.
   - Shaman goes first.

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
