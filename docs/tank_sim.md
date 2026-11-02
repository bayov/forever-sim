# Tank sim

Tank presets built like the DPS Report's, scored on threat and on survival. Tank optimization trades
TPS against damage taken, so the sim has to sim the boss hitting the tank and say how close the tank came
to dying. This comes before the raid sim's simmed tanks ([raid_sim.md](raid_sim.md), Threat), which
replace its 700 TPS stand-in.

**Where it stands (2026-11-02):** the engine reports a tank sheet with the thresholds tanks gear for, and
`tools/tankbench` sims tank builds against the report's boss with a stand-in healer and finds each build's
healing needed. Two builds run (Protection warrior and paladin, both on Classic pre-raid plate). Next comes
the rules audit, then tank gear and rotations. Real healers come from the healing sim
([healing_sim.md](healing_sim.md)).

## The user's rulings (2026-11-02)

- **Threat against survival:** weights, with 2 or 3 presets along the curve ("Max threat", "Balanced", "Max
  survival"), like the PvP High HP preset. No death chance floor.
- **The survival score:** healing needed (see [Scoring survival](#scoring-survival)). The user: "sometimes no
  amount of healing will work, if tank can be one-shotted". So the bench also reports the boss's biggest hit
  and the death floor, and a build that dies too often whatever the healing says so.
- **Forever's boss crit:** audit which rule is more likely and go with it until a log confirms it. The audit
  below picks Forever's 0.6%.
- **The stand-in healer stays** until the healing sim's healers heal, then tanks switch to them.
- **Tanks are scored on realistic raid encounters**, not a generic boss: each one the boss's real damage on
  its tank and the healing such a raid puts on it, both tunable, the healing most of all, since it's our own
  group's (the user: "healing ... depends on our group which we control rather than the boss itself which
  we dont"). See [Encounters](#encounters).
- **Sources for boss and healing numbers, best first:** Warcraft Logs from Season of Discovery, then the
  Anniversary realms (Vanilla, then TBC), then Classic Era, then private server cores like vmangos
  (docs/data_sources.md).

## Encounters

**One set of boss encounters for the tank and healing sims** (agreed with the healing sim session,
2026-11-02), in `tools/encounters/<boss>.json` (its README has the format). Each file is a proto Encounter
(the boss's melee, its abilities as raid damage events, tank swaps) and the healing a raid puts on its main
tank, with the sources and the Warcraft Logs reports behind the numbers. The tank sim writes the tank side
(the boss's melee, its abilities on the tank, the healing on the tank), the healing sim the raid damage.

**The healing on the tank** is set the way a raid sets it: healers on the tank, each one's healing per second
on it, and their heals' cast time. Until healers heal, the bench turns that into the stand-in stream: their
HPS added up, a heal every cast time over the healers, each one healer's heal. `-healers`, `-healer-hps` and
`-heal-cast` change it for a run.

**A boss's abilities on the tank** are raid damage events aimed at the boss's target or the tanks. A melee
one (Mortal Strike, Cleave, Hateful Strike) rolls the tank's miss, dodge, parry and block like a white swing,
crits on its own roll and never crushes (`melee`, `OutcomeEnemyMeleeSpecial`), and one can cut the healing
taken for a while (`healing_taken_multiplier`, `debuff_seconds`, Mortal Strike's 50% for 5 s). Not modelled
yet: threat drops (Knock Away and the like).

**The first seven** (2026-11-02): Magmadar, Golemagg, Ragnaros, Onyxia, Broodlord Lashlayer, Vaelastrasz and
Chromaggus, their tank side only. No Warcraft Logs numbers were found on the web, so they're from the vmangos
core, the last source in the order, with each file's `source` saying which numbers and `notes` what's left out.
A 2006 Molten Core full clear from Blizzard's servers (WWS, hits after armor) backs vmangos's boss melee: the
cmangos core's is about a third of it and doesn't fit the log. Patchwerk is waiting: neither core's melee is
confirmed and the two differ, and Hateful Strike is an off-tank's problem. Each file has the boss's Warcraft
Logs encounter ID for when logs replace the numbers.

The raid's healing on the tank in them is a guess until the logs: 4 healers at 400 HPS in 2 s heals in Molten
Core and Onyxia (about 2.4 times a pre-raid warrior's damage taken, the Classic Era logs' ratio), 4 at 500 in
Blackwing Lair.

**First numbers** (1,000 iterations, the two Valor-geared builds, healing needed at 5% deaths, the death
chance at the file's healing):

| Encounter | Warrior TPS | DTPS | Healing needed | Death | Paladin TPS | DTPS | Healing needed | Death |
|---|---|---|---|---|---|---|---|---|
| Magmadar | 1051 | 253 | 255 | 0% | 504 | 290 | 299 | 0% |
| Golemagg | 1185 | 448 | 571 | 0% | 588 | 492 | 615 | 0% |
| Ragnaros | 1159 | 916 | 1976 | 16% | 461 | 961 | 2336 | 22% |
| Onyxia | 1209 | 527 | 597 | 0% | 611 | 563 | 633 | 0% |
| Broodlord | 1190 | 625 | none works (floor 43%) | 81% | 555 | 682 | none works (floor 46%) | 85% |
| Vaelastrasz (45 s) | 1251 | 1389 | 1802 | 2% | 566 | 1463 | 2251 | 7% |
| Chromaggus | 1147 | 480 | 588 | 0% | 506 | 539 | 678 | 0% |

- **Broodlord is the one-shot case.** Mortal Strike's 13,730 to 18,200 before armor comes to about 6,000 to
  7,900 on these tanks' 56% armor, against 7,000 health (the Valor set has no neck, back, rings or trinkets),
  and halves the healing for 5 s. No healing saves them, so the build isn't viable there. A real Blackwing
  Lair tank has more health and armor, which is exactly what the score should push the search toward.
- **Ragnaros** needs about 2.2 times the warrior's damage taken: Elemental Fire's 600 a second for 8 s on
  most hits, with only the raid's fire resistance buffs against it. A fire resistance set is the next test.
- **Vaelastrasz** is one tank's 45 s stint.

**From the logs:** `tools/warcraftlogs/wclcli.py tank` reads, per encounter, the boss's swings and abilities
on the main tank, the tank's damage taken a second, and its healing taken (effective, overheal, absorbs,
heals a second, healers, top heal spells). It needs a Warcraft Logs API client, on the user's machine or as
`WCL_CLIENT_ID` and `WCL_CLIENT_SECRET` in a cloud environment.

## What tanks gear for

From the Classic, TBC and Season of Discovery tanking guides and wowsims' code (sources at the end), and
what our engine does under Forever. Chances are against a level 63 boss.

| Threshold | Classic | What it means | Forever in our sim |
|---|---|---|---|
| Crit immune | 440 Defense (5.6% crit, 0.04% a point) | crits do double damage, the main spike | **315 Defense** (the boss crits 0.6%, see below), 440 with `classic_crit` |
| Uncrushable | miss + dodge + parry + block ≥ 102.4% on the sheet | a crushing blow does 150%, 15% of swings | the same: the sheet's coverage at 100% (our chances already take off the 0.6% a stat for the level gap) |
| Crush chance | 15% from a boss 3 levels up, none from 60 to 62 | | 15% at +3, none below |
| Armor cap | 75% at 17,265 armor | A / (A + 400 + 85 × 63) | the same, capped since 2026-11-02 |
| Hit cap (threat) | 9% special at 300 skill | | **8%**, Forever has no hit suppression |
| Boss dodge, parry | 6.5% and 14% from the front at 300 skill | lost threat, parry haste | from our attack table, on the sheet |
| Effective health | health / (1 − armor reduction) | how much raw damage from full | the same, with damage taken modifiers |

Tanks wear a threat set, a mitigation set or a mix, by fight. The "440 Defense is mandatory" view is
considered disproven for Classic: most raid tanks wore hybrids. A bear can't block or parry, so it can never
be uncrushable and leans on armor and health. TBC's own caps (490 Defense, expertise) don't apply here.

**They shift by encounter.** The sheet reads the boss from the encounter, so all of them move with it:

- **Boss level:** crit and each avoidance move 0.2% a level, and crushing blows need +3.
- **Boss school:** magic damage ignores armor and avoidance. Resistance takes R / (5 × 63) × 0.75 on
  average, capped at 75% (315 resistance). Ragnaros' main tank wore fire resistance.
- **More attackers:** each swing rolls on its own, and they use up Shield Block's and Holy Shield's
  charges faster.
- **A dual wielding boss** misses 19% more.
- **Swing speed and damage** don't move the caps. They move the spikes, the TMI and the chance of death.

### Forever's boss crit

**We go with Forever's rule: a level 63 boss crits a player 0.6% of the time, not Classic's 5.6%**
(`sim/core/target.go`, `newWeaponAttackTable`). So crit immunity costs 15 Defense instead of 140, which
takes most of Defense's value away and changes every tank gear set. Audited 2026-11-02, more likely than
Classic's rule but not confirmed:

**For Forever's rule:**

- **Forever's server doesn't use Classic's rule for the mobs we've seen**, below or above us. In the
  user's level 30 beta logs, about 2,900 swings from mobs at or under our level never crit, where Classic
  gives about 85. Mobs 1 to 5 levels above crit 4 times in 745 swings, where Classic gives about 40 and
  Forever's rule 4.6. A level 1 to 3 character took 136 swings with none (Classic about 7). The chance of
  so few crits under Classic's rule is effectively nil (notes.md, how often mobs crit us).
- **The rule that fits keeps Classic's level term and drops its 5% base.** Applied by level, a level 63
  boss is 0.6%. Nothing in the client data or the dev notes says bosses use another formula.

**For Classic's rule:**

- **No elite or boss has been seen.** The logs are beta leveling zones (Durotar to level 30). A boss could
  carry its own crit, or the server could give elites and bosses another rate.
- **The guides say 440**, Warcraft Tavern's Forever stats guide among them, but that's Classic's text, not
  a measurement.
- **wowsims/forever uses 5.6%**: the TBC code they started from, with no Forever measurement in their
  history.

**Settles it:** a few hundred boss or elite swings on a tank 2 or 3 levels below it, in a dungeon on the
release realms (a level 58 to 60 tank in Blackrock Spire or Stratholme), counting the crits. At Classic's
5.4% for +2, 300 swings give about 16 crits, Forever's rule about 1.

**Both can be run.** A target's `classic_crit` (`proto/common.proto`, Target) gives it Classic's rate, and
the bench's `-classic-crit` sets it.

## The tank sheet

`UnitMetrics.tank_sheet`, for a player the boss attacks (`sim/core/tank_sheet.go`). It's taken once, as the
first iteration's prepull starts: permanent buffs and consumables on, no cooldowns or procs. It uses the
same attack table the boss's swings roll on.

- **The boss's swing at us**, on its one roll table in order: miss, dodge, parry, block, crit, crush, hit
- **Defense**, and the Defense at which crits stop
- **Coverage**: miss, dodge, parry and block added up, and the **crush gap** still needed to push crushing
  blows off the table
- **Armor** and its reduction, **block value**, **health**, **effective health**
- **The biggest hit**: the boss's biggest white swing on us (the top of its damage range, a crit if it can
  crit us, else a crushing blow, through armor and damage taken)
- **Our side**, which sets threat: our special attacks' miss, the boss's dodge and parry from the front,
  our spells' miss

What the fight did with Shield Block and Holy Shield shows in the boss's swing outcomes instead, which the
bench reads from the boss's action metrics ("Took").

## The bench

`tools/tankbench` sims tank requests (`tools/tankbench/req/*.json`, the DPS Report's request format with
the tank as the first player) against a boss attacking the tank, with a stand-in healer:

```bash
go run -tags=with_db ./tools/tankbench -iters 3000 tools/tankbench/req/*.json
```

- **The boss** is the request's, the report's Level 60 boss by default (level 63, 2.0 s swing, 3,000 to
  4,000 before armor, 805 attack power). `-boss-level` and `-boss-damage` change it.
- **The stand-in healer** is the engine's healing model (below), at `-hps` (default 600) every `-cadence`
  seconds (2 ± 1). It's the same trick as the raid sim's 700 TPS tank.
- **It prints** TPS, DPS, DTPS, TMI (6 s windows), the chance of death, the sheet, and how the boss's swings
  landed over the fight. `-actions` adds threat by ability, `-needed` the healing needed and the death floor
  (about 10 more sims a build), `-classic-crit` gives the boss Classic's crit.

### The stand-in healer

**One stream of heals, not healers** (`applyHealingModel`, `sim/core/health.go`, from wowsims):

- A heal lands every cadence seconds on average. With a variation, each gap is drawn from two even spreads
  either side of the cadence (2 ± 1 s: anywhere from 1 to 3 s).
- **Each heal is the HPS times the time since the last one.** A long gap gets a big heal, so the healing per
  second is exact and only its timing is lumpy.
- Healing over the tank's max health is lost. Healing taken modifiers apply.
- **Not modelled:** separate healers, cast times and heal sizes, crits, shields, reacting to the tank's
  health (a healer who sees a big hit casts a fast heal), healers going out of mana. The heal comes whatever
  the tank's health, and stops for nothing.

**What real raids look like:** wowsims/forever's Classic Era Molten Core logs put the main tank's incoming
healing at 2.4 times its damage taken, overhealing included, in heals about 0.3 s apart, 1.5% of it absorbs
(their Protection warrior's `OtherDefaults`). So a real main tank is healed by several healers at once, far
above its DTPS, and dies to a burst between heals rather than to running short on healing.

**The real healers come from the healing sim** ([healing_sim.md](healing_sim.md)): each healer's own spells,
cast times, mana, shields and targets. Until then the stream is what scores tanks, and healing needed reads
as "how much a steady healer would need", not as a healer comp.

**First numbers** (2026-11-02, 2,000 iterations, 180 s, 600 HPS, Forever's boss crit). Both builds wear the Classic Valor set,
Mirah's Song and Draconian Deflector (`ui/tank_warrior/gear_sets/p0.bis_sword_board`, no neck, back, rings
or trinkets), so these test the sim, not the classes:

| Build | TPS | DTPS | TMI | Death | EH | Coverage | Crush gap | Took: blocked | crushed |
|---|---|---|---|---|---|---|---|---|---|
| Protection warrior | 1137 | 431 | 51 | 0% | 18,048 | 46.6% | 53.4% | 59% | 1.4% |
| Protection paladin | 501 | 497 | 61 | 8.6% | 18,389 | 36.7% | 63.3% | 23% | 12.3% |

Shield Block keeps the warrior nearly uncrushable (2 blocks over 7 s on a 5 s cooldown). The paladin's
Holy Shield is down most of the fight because its rotation is stale (below).

**The chance of death is a cliff at the tank's DTPS.** The warrior dies in 98% of fights at 350 HPS, 65% at
425, 12% at 500 and none at 600. The paladin goes from 100% at 350 to 9% at 600. So one healer number mostly
says where the cliff falls, and a small change in DTPS swings the result.

## Scoring survival

**Healing needed** (the user's ruling): the stand-in healer's HPS at which the tank dies in 5% of fights,
found by bisection to 1% of DTPS (`healingNeeded`, `tools/tankbench`). It's one number per build that
doesn't depend on picking a healer, and it counts both damage taken and spikes (a spiky build needs more
healing to stay at 5%). TMI and DTPS stay next to it. Every run uses the request's seed, so the chance of
death moves only with the HPS.

**Sometimes no healing is enough.** A tank one swing can kill from full, or one a burst kills between two
heals, dies whatever the healing. So the bench also reports:

- **The biggest hit** (the sheet's) and how many of them the tank's health takes. Under 1 is a one-shot.
- **The death floor**: the chance of death at 20 times the DTPS, where every heal tops the tank off, so
  only damage between two heals kills it. At 5% or over, healing needed is "none works", and the build
  isn't viable against that boss with that healer cadence.

The floor depends on the cadence. With 0.3 s heals it's lower than with 2 s heals, so the cadence is part of
the scoring setup and stays fixed across builds.

**The first numbers** (1,000 iterations, the builds above):

| Build | Boss crit | DTPS | Healing needed | Death floor | Biggest hit | Hits to kill |
|---|---|---|---|---|---|---|
| Protection warrior | Forever's | 431 | 515 (1.20× DTPS) | 0% | 2,777 (a crushing blow) | 2.5 |
| Protection paladin | Forever's | 497 | 617 (1.24×) | 0% | 2,697 | 2.6 |
| Protection warrior | Classic's | 433 | 527 (1.22×) | 0% | 3,703 (a crit) | 1.9 |
| Protection paladin | Classic's | 518 | 696 (1.34×) | 0% | 3,596 | 1.9 |

Classic's crit costs the paladin far more than the warrior, since Shield Block pushes most crits off the
warrior's table and the paladin's Holy Shield is down most of the fight.

## The optimizer's objective (proposed)

From a metrics review (2026-11-02, a subagent that read the bench and measured its noise). Not ruled on yet.

**Score, in raid DPS: threat members × SafeDPS − w × DPS slot × healing needed / one healer's HPS.**

- **Healing needed stays the survival number.** DTPS gives Stamina no value, effective health ignores
  avoidance, TMI is a stand-in for what healing needed measures directly, and the chance of death at the
  encounter's own healing is a cliff (0% or 100%) with no slope for a search to follow.
- **Computed by replay, not bisection.** Without `track_health` the damage and the heal times don't depend
  on the HPS, only the heal sizes do. So one sim can record each iteration's damage and heal gaps and find
  the least HPS that iteration lives through. Healing needed is then a percentile of those, with no rounding
  step and with a standard error for the search's noise test, at a tenth of the cost. Measured: the
  bisection's step (1% of DTPS, 4.3 HPS for the warrior) is bigger than the sims' own noise (about 2.4 HPS
  over 6 seeds of 1,000 iterations). It's exact only while nothing reads the tank's health (health-based
  defensives and potions), which a scoring run turns off.
- **SafeDPS for threat:** the steady DPS a reference DPS member can do without passing the tank at any
  moment of the fight, the mean over iterations of each one's lowest ratio. Passes come in the first
  seconds ([raid_sim.md](raid_sim.md), Threat), which fight TPS hides, and a threat drop shows as a dip.
  Fight TPS stays as a column.
- **Both in raid DPS:** a healer on the tank costs one DPS slot, and SafeDPS frees the DPS of the members
  close to the tank's threat (2 or 3). So w = 1 means one more healer costs one DPS player, which holds
  across bosses and gear. Presets: Max threat w = 1/3, Balanced 1, Max survival 3.
- **One hard reject:** a death floor over 5% (no healing saves it).
- **The sheet's thresholds report and explain, they aren't scored.** Their effect is already in the
  outcomes. Final presets are re-checked with `-classic-crit`, where a 440 Defense build can flip.

Engine work it needs: the replay data and percentiles in the metrics, the tank's threat over time against a
reference curve, threat drops on boss abilities, and a tank objective in rotopt's search.

## Open decisions

1. **The objective above**, and the weights of the three presets.

## Findings so far

- **The DPS Report's requests give the raid Salvation** (`raid.buffs.salvation`, −30% threat). Copied into a
  tank's request it cut the warrior from 1137 to 794 TPS. The bench's requests drop it.
- **The raid sim's 700 TPS stand-in** came from the warrior's tank suite. On the report's boss with no
  Salvation the same build does 1137. Check which buffs the 700 had before the raid sim's simmed tanks
  replace it.
- **The Protection paladin's rotation is stale.** `basic_prot` casts Holy Shield only while aura 20134 is
  up, a Season of Discovery rune, so Holy Shield is up about a third of the time. Its options need Righteous
  Fury on (`righteousFury`), which the test suite has and the UI preset should check.
- **The Feral tank isn't simmed at all.** `sim/druid/_tank` and `_maul.go`, `_frenzied_regeneration.go` are
  left out of the build, and `register_all.go` has it commented out.
- **The paladin and the Warden shaman have no gear** (`blank` sets only), and no tank has Forever synthetic
  gear or a tuned rotation.
- **Armor had no cap.** Armor now takes off at most 75% (`GetArmorDamageModifier`). Nothing simmed before
  got near it.

## Next steps

1. **Rules audit** of each tank's defensive talents, stats and abilities against Forever's client data and
   foreverchanges (Protection warrior, Feral tank, Protection paladin, Warden shaman), like the class audits.
   Bring the Feral tank into the build. Ask for a dungeon boss log for Forever's boss crit.
2. **Rotations**: the paladin's and the Warden's first, then rotopt templates for all four.
3. **The objective** (above) in the engine and rotopt's gear and talent searches.
4. **Tank gear**: a tank row in `orchestration/gear_targets.md` and synthetic tank items, then gear and talent
   searches on the score.
5. **The report**: a tank section with TPS next to damage taken, healing needed and the sheet's thresholds.
6. **The raid sim**: the simmed tank's threat in place of the 700 TPS stand-in, then the comp search.

## Sources

- Attack table and Defense: https://github.com/magey/classic-warrior/wiki/Attack-table,
  https://warcraft.wiki.gg/wiki/Attack_table, https://warcraft.wiki.gg/wiki/Defense,
  https://warcraft.wiki.gg/wiki/Crushing_blow
- Armor and resistance: https://warcraft.wiki.gg/wiki/Armor, https://warcraft.wiki.gg/wiki/Resistance
- Shield Block and Holy Shield: https://warcraft.wiki.gg/wiki/Shield_Block,
  https://www.wowhead.com/classic/spell=20925
- Uncrushable and TBC caps: https://warcrafttavern.com/tbc/guides/pve-protection-paladin-tank-stat-priority,
  https://www.curseforge.com/wow/addons/tankcaps
- Classic practice: https://almarsguides.com/wow/gettingstarted/Classes/Classic/Warrior/Tanking
- Season of Discovery: https://www.warcrafttavern.com/wow-classic-sod/guides/protection-warrior-tank/
- wowsims/sod's tank metrics: `sim/core/health.go` and `calculateTMI` in `sim/core/metrics_aggregator.go`
  (the same code is in ours)
