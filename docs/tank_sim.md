# Tank sim

Tank presets built like the DPS Report's, scored on threat and on survival. Tank optimization trades
TPS against damage taken, so the sim has to sim the boss hitting the tank and say how close the tank came
to dying. This comes before the raid sim's simmed tanks ([raid_sim.md](raid_sim.md), Threat), which
replace its 700 TPS stand-in.

**Where it stands (2026-11-02):** the engine reports a tank sheet with the thresholds tanks gear for, and
`tools/tankbench` sims tank builds against seven Classic raid bosses (`tools/encounters`) with the healing a
raid puts on its tank, and finds each build's healing needed and death floor. The bosses' melee, the fight lengths
and the healing are Warcraft Logs numbers from the Anniversary realms, which score tanks, with Season of Discovery's
kept to compare ([Encounters](#encounters)). Two builds run (Protection
warrior and paladin, both on Classic pre-raid plate). Four audits checked the work and their findings are
fixed ([Audits](#audits)). The optimizer's objective is ruled and built: the bench prints it, and rotopt's
searches maximize it with `-tank`. Next comes the rules audit, then tank gear and rotations. Real healers come from the healing sim ([healing_sim.md](healing_sim.md)).

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
- **The optimizer's objective** as proposed (see [The optimizer's objective](#the-optimizers-objective)):
  SafeDPS against healing needed in raid DPS, healing needed by replay, presets at w = 1/3, 1 and 3.
- **Tanks keep 5% deaths** for healing needed. The healing sim judges healers at under 2% tank deaths, a
  separate line for a separate score.

## Encounters

**One set of boss encounters for the tank and healing sims** (agreed with the healing sim session,
2026-11-02), in `tools/encounters/<boss>.json` (its README has the format). The tank bench reads them
(`-encounter`), the heal bench not yet (it has its own `-profile`). Each file is a proto Encounter
(the boss's melee, its abilities as raid damage events, tank swaps) and the healing a raid puts on its main
tank, with the sources and the Warcraft Logs reports behind the numbers. The tank sim writes the tank side
(the boss's melee, its abilities on the tank, the healing on the tank), the healing sim the raid damage.

**The healing on the tank** is set the way a raid sets it: healers on the tank, each one's healing per second
on it, and their heals' cast time. Until healers heal, the bench turns that into the stand-in stream: their
HPS added up, a heal every cast time over the healers, each one healer's heal. `-healers`, `-healer-hps` and
`-heal-cast` change it for a run.

**A boss's abilities on the tank** are raid damage events aimed at the boss's target or the tanks. A melee
one (Mortal Strike, Cleave, Hateful Strike) rolls the tank's miss, dodge, parry and block like a white swing,
crits on its own roll and never crushes (`melee`, `OutcomeEnemyMeleeSpecial`), with no dual wield miss, and
one can cut the healing taken for a while (`healing_taken_multiplier`, `debuff_seconds`, Mortal Strike's 50%
for 5 s). A ticking event is never melee, and its debuff comes when it's applied. Not modelled yet: threat
drops (Knock Away and the like), armor cuts (Corrosive Acid, Magma Splash), stacking DoTs (each stack is
its own event), and abilities that scale with the boss's attack power (the files take Demoralizing Shout's
cut off the weapon damage ones).

**The seven** (2026-11-02): Magmadar, Golemagg, Ragnaros, Onyxia, Broodlord Lashlayer, Vaelastrasz and
Chromaggus, their tank side only. Forever's own raids (Barrow Deeps, Hyjal Summit, Onyxia) have no logs yet, so
these are stand-ins.

**Their numbers are Warcraft Logs'** (2026-11-02). **Tanks are scored on the Anniversary realms' set**,
`tools/encounters/` (the user, 2026-11-02). `tools/encounters/sod/` is the same bosses from Season of
Discovery's Normal raids, first in the source order but tuned far above a pre-raid tank, kept to compare. From the logs: the boss's melee, his weapon damage abilities (Cleave,
Mortal Strike, Knock Away), the fight's length, Elemental Fire's rate and the healing on the main tank. The other abilities
keep the Classic Era client's spell data and the vmangos scripts' timers. `tools/encounters/from_logs.py`
writes them from what `wclcli.py tank` mined (`tools/encounters/logs/`, with each sample of fights), and
each file's `source`, `reports` and `warcraftLogs` say which kills.

- **What was checked first:** `wclcli.py tank` on real reports. The boss's damage comes back only with
  `hostilityType: Enemies` (133 events against none without, SoD Onyxia), and the main tank's events have
  `heal` events with `overheal` (130 of 136) and `absorbed` events.
- **The fights** are the newest public reports of each zone, not rankings: ranked fights are the fastest
  kills, and ranked reports from 2019 to 2024 are archived (their events need a paid account). Season of
  Discovery: 20 to 47 Normal kills a boss (2025 to 2026; Heroic and Mythic, the heat levels, hit up to 40%
  harder and are left out). Anniversary: 21 to 30 kills a boss from the realms' first months (Molten Core
  and Onyxia December 2024 to February 2025, Blackwing Lair March to July 2025), all 40 players.
- **The Anniversary logs are its Hardcore realms'** (Doomhowl, Soulseeker, the same raids). Its Normal
  and PvP realms weren't found through the API: `fresh` serves the Classic progression realms' reports
  (`tools/warcraftlogs/README.md`). Era wasn't needed, both sources having all seven, and TBC has none of them.
- **Everything is measured while the boss is on the main tank**: his swings on it no more than 6 s apart.
  Tank swaps, Onyxia's air phase and Ragnaros's knockbacks fall outside.
- **The melee** is the median normal hit's `unmitigatedAmount` (before armor and before the crit or
  crushing multiplier), with the spread of each hit over its own kill's median (about 0.8 to 1.2 of it, close
  to Classic's ±14%: across kills the raids' debuffs on the boss differ). The swing gap is the mean while he's on
  the tank, parry haste and Frenzy in. Both are taken back through the debuffs the bench puts on him
  (`tankbench -boss-stats`: attack power 805 to 601, −37 a swing, attack speed 0.833), so under them he
  swings as in the logs. That assumes the logged raids had about the bench's debuffs.
- **The length** is the kills' median, except Onyxia and Ragnaros (the main tank's time on him a kill) and
  Vaelastrasz (his longest stretch on one tank, one tank's stint).
- **The healing** is the healers with 5% or more of the main tank's healing in the median kill, the
  healing it took from them with overheal and absorbs, and their heals and HoT ticks a second. The tank's own
  heals are left out (Season of Discovery's main tanks were mostly druids, warlocks and paladins, the
  Anniversary's warriors).
- **Golemagg's Double Attack** goes: its swings log as white swings, so they're in the swing rate.
  **Elemental Fire** comes at the logs' rate: 18.9 a minute on the Anniversary realms (3 of 4 swings, the
  client's proc on landed hits), 5.5 in Season of Discovery's Normal.

| Encounter | Melee before armor | Swing | Length | Healers × HPS, heal | SoD | Anniversary |
|---|---|---|---|---|---|---|
| Magmadar | 1,902 to 2,699 / 1,541 to 2,441 | 1.37 / 1.38 s | 29 / 53 s | 3 × 374, 1.8 s / 8 × 258, 3.1 s | 20 players | 40 |
| Golemagg | 1,441 to 2,147 / 1,744 to 2,914 | 1.70 / 1.45 s | 27 / 56 s | 3 × 445, 1.5 s / 7 × 258, 3.2 s | | |
| Ragnaros | 1,918 to 2,879 / 1,626 to 2,671 | 2.70 / 2.40 s | 64 / 71 s | 3 × 383, 2.1 s / 6 × 285, 2.9 s | | |
| Onyxia | 3,029 to 4,428 / 1,309 to 2,166 | 2.22 / 1.92 s | 124 / 35 s | 3 × 307, 2.3 s / 8 × 174, 4.5 s | 40 | 40 |
| Broodlord | 3,096 to 4,652 / 2,042 to 3,195 | 1.74 / 1.78 s | 97 / 41 s | 4 × 485, 1.5 s / 8 × 300, 3.0 s | 20 and 40 | 40 |
| Vaelastrasz | 4,417 to 6,681 / 3,325 to 4,705 | 2.56 / 2.64 s | 25 / 17 s | 3 × 1,098, 1.1 s / 6 × 714, 1.8 s | 20 | 40 |
| Chromaggus | 4,237 to 6,071 / 2,513 to 4,161 | 1.16 / 1.37 s | 218 / 115 s | 4 × 424, 1.7 s / 8 × 303, 3.1 s | 20 and 40 | 40 |

(Season of Discovery / Anniversary.) Broodlord's Mortal Strike is 16,110 to 26,548 before armor in Season
of Discovery and 10,784 to 15,905 on the Anniversary realms. Raids heal their main tank at 2 to 4 times its
damage taken, with 60 to 80% overheal (Season of Discovery up to 90%).

**What it shows:**

- **Season of Discovery's Molten Core is near Classic's; its Onyxia and Blackwing Lair aren't.** Its Normal Molten
  Core hits like the Anniversary's (and vmangos's). Its Onyxia hits twice as hard, and its Blackwing Lair 1.4 to 1.6
  times, tuned for its own players: at Onyxia, Broodlord and Chromaggus its main tanks took 460 to 800 damage a second
  where ours take 920 to 1,380.
- **vmangos's melee was about right**, Broodlord's too: the 0.6 of vmangos the file took from players' reports
  was low (the Anniversary's 2,042 to 3,195 against vmangos's 2,746 to 3,640 before debuffs).
- **Classic raids' healing is many healers on slow heals**: 6 to 8 healers with 5% or more, a heal every 0.3 to 0.6 s
  in all. Season of Discovery's is 3 or 4 healers on fast heals.

**The bench on them** (1,000 iterations, the two Valor-geared builds, healing needed at 5% deaths, the death
chance at the file's healing):

| Encounter | Warrior TPS | DTPS | Healing needed | Death | Paladin TPS | DTPS | Healing needed | Death |
|---|---|---|---|---|---|---|---|---|
| Magmadar (SoD) | 1199 | 470 | 385 | 0% | 627 | 520 | 448 | 0% |
| Golemagg (SoD) | 1111 | 294 | 155 | 0% | 583 | 335 | 203 | 0% |
| Ragnaros (SoD) | 1142 | 695 | 1013 | 1.9% | 498 | 744 | 1150 | 5.1% |
| Onyxia (SoD) | 1271 | 917 | 1471 | 88% | 620 | 984 | 1605 | 98% |
| Broodlord (SoD) | 1255 | 932 | none works (floor 87%) | 92% | 605 | 1022 | none works (floor 92%) | 96% |
| Vaelastrasz (SoD) | 1237 | 1584 | 2747 | 3.2% | 568 | 1683 | 4324 | 7.1% |
| Chromaggus (SoD) | 1283 | 1271 | 2124 | 47% | 639 | 1381 | 2387 | 73% |
| Magmadar (Anniversary) | 1176 | 402 | 372 | 0% | 614 | 450 | 421 | 0% |
| Golemagg (Anniversary) | 1190 | 441 | 427 | 0% | 601 | 494 | 484 | 0% |
| Ragnaros (Anniversary) | 1223 | 1219 | 2761 | 31% | 520 | 1265 | 2961 | 40% |
| Onyxia (Anniversary) | 1209 | 538 | 556 | 0% | 663 | 569 | 587 | 0% |
| Broodlord (Anniversary) | 1196 | 606 | 3519 (floor 4.5%) | 6.5% | 595 | 666 | 2541 (floor 3.8%) | 5.5% |
| Vaelastrasz (Anniversary) | 1219 | 1171 | 1528 | 0% | 563 | 1244 | 1635 | 0% |
| Chromaggus (Anniversary) | 1261 | 718 | 905 | 0% | 611 | 791 | 1006 | 0% |

- **Season of Discovery's Broodlord is a one-shot** for a pre-raid tank (7,000 health): his Mortal Strike lands for
  7,000 to 11,500 through 56% armor. Its Onyxia and Chromaggus kill at the healing their raids gave their own,
  sturdier tanks. Scoring Forever's pre-raid tanks on them mostly measures that gap, so **tanks are scored on the
  Anniversary realms' set** (the user, 2026-11-02).
- **Healing needed under the damage taken** (Golemagg, Magmadar) is the short fights: a 27 s kill takes about a
  quarter of a tank's health from its own pool.
- **Ragnaros on the Anniversary realms** needs 2.3 times the damage taken, and its tanks took 725 a second where
  ours take 1,219: theirs wore fire resistance, which resists most of Elemental Fire (its ticks landed for 135 of 600).
- **Broodlord's Mortal Strike** puts the death floor near 5% on the Anniversary realms (11,000 to 16,000 before
  armor), so healing needed there is unstable (its error is unbounded): a 7,000 health tank is near the line.

**From the logs:** `tools/warcraftlogs/wclcli.py tank` reads, per encounter, the boss's swings and abilities
on the main tank, the tank's damage taken a second, and its healing taken (effective, overheal, absorbs,
heals a second, healers, top heal spells). It needs a Warcraft Logs API client, on the user's machine or as
`WCL_CLIENT_ID` and `WCL_CLIENT_SECRET` in a cloud environment. `tools/warcraftlogs/README.md` has where each
realm's logs are.

## What tanks gear for

From the Classic, TBC and Season of Discovery tanking guides and wowsims' code (sources at the end), and
what our engine does under Forever. Chances are against a level 63 boss.

| Threshold | Classic | What it means | Forever in our sim |
|---|---|---|---|
| Crit immune | 440 Defense (5.6% crit, 0.04% a point) | crits do double damage, the main spike | **315 Defense** (the boss crits 0.6%, see below), 440 with `classic_crit` |
| Uncrushable | miss + dodge + parry + block ≥ 102.4% on the sheet | a crushing blow does 150%, 15% of swings | the same: the sheet's coverage plus the boss's crit at 100% (our chances already take off the 0.6% a stat for the level gap) |
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
- **Boss school:** magic damage ignores armor and avoidance. Under Forever resistance takes R / (R + 9 × 63)
  on average with no cap (35.7% at 315), and a DoT's tick meets a tenth of it (`foreverAverageResist`).
  Classic's was R / (5 × 63) × 0.75, capped at 75% at 315. Ragnaros' main tank wore fire resistance.
- **More attackers:** each swing rolls on its own, and they use up Shield Block's and Holy Shield's
  charges faster.
- **A dual wielding boss** misses its white swings 19% more, not its abilities.
- **Swing speed and damage** don't move the caps. They move the spikes, the TMI and the chance of death.

### Forever's boss crit

**We go with Forever's rule: a level 63 boss crits a player 0.6% of the time, not Classic's 5.6%**
(`sim/core/target.go`, `newWeaponAttackTable`). So crit immunity costs 15 Defense instead of 140, which
takes most of Defense's value away and changes every tank gear set. Audited 2026-11-02, more likely than
Classic's rule but not confirmed:

**For Forever's rule:**

- **Forever's server doesn't use Classic's rule for the mobs we've seen**, below or above us. In the
  user's level 30 beta logs, about 2,900 swings from mobs at or under our level never crit, where Classic
  gives about 85. Mobs 1 to 5 levels above crit 4 times in 745 swings, where Classic gives about 42 and
  Forever's rule 4.6 (a flat 0.5% above us fits the 4 crits as well). A level 1 to 3 character took 136 swings with none (Classic about 7). The chance of
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
  blows off the table (1 less the crit and the coverage, since crit sits before crush on the table)
- **Armor** and its reduction, **block value**, **health**, **effective health**
- **The biggest hit**: the boss's biggest white swing on us (the top of its damage range with its flat weapon
  damage, a crit if it can crit us, else a crushing blow, through armor and our flat and multiplied damage
  taken). Its abilities aren't in it: a one-shot from one shows in the bench's death floor.
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

- **The boss** is the request's, the report's Level 60 boss by default (level 63, 2.0 s swing, about 3,975
  to 4,975 before armor with its 805 attack power), or an encounter's (`-encounter`). `-boss-level` and
  `-boss-damage` change it, `-dur` the fight length.
- **The stand-in healer** is the engine's healing model (below), at `-hps` (default 600) every `-cadence`
  seconds (2 ± `-cadence-variation` 1), or the encounter's healers (`-healers`, `-healer-hps`, `-heal-cast`),
  whose variation is half their cadence. It's the same trick as the raid sim's 700 TPS tank.
- **It prints** the objective at the three presets, SafeDPS, healing needed at `-death` (5%) and the death
  floor, all from the one sim, then TPS, DPS, DTPS, TMI (`-tmi-window` 6 s), the chance of death, the sheet,
  and how the boss's swings landed over the fight. `-actions` adds threat by ability, `-bisect` checks the
  replay's healing needed by bisection over whole sims (about 10 more a build), `-classic-crit` gives the
  boss Classic's crit, and `-dps-start`, `-dps-threat`, `-threat-members`, `-dps-slot` and `-healer-each` set
  the objective.

### The stand-in healer

**One stream of heals, not healers** (`applyHealingModel`, `sim/core/health.go`, from wowsims):

- A heal lands every cadence seconds on average. With a variation, each gap is drawn from two even spreads
  either side of the cadence (2 ± 1 s: anywhere from 1 to 3 s).
- **Each heal is the HPS times the time since the last one.** A long gap gets a big heal, so the healing per
  second is exact and only its timing is lumpy.
- Healing over the tank's max health is lost. Healing taken modifiers apply.
- **Approximated:** separate healers, cast times and heal sizes, as one stream whose heals are one healer's
  heal (an encounter's healers).
- **Not modelled:** crits, shields, reacting to the tank's health (a healer who sees a big hit casts a fast
  heal), healers going out of mana. The heal comes whatever the tank's health, and stops for nothing.

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

**Healing needed** (the user's ruling): the stand-in healer's HPS at which the tank dies in 5% of fights.
It's one number per build that doesn't depend on picking a healer, and it counts both damage taken and
spikes (a spiky build needs more healing to stay at 5%). TMI and DTPS stay next to it. It comes from a
replay of each iteration (see [The optimizer's objective](#the-optimizers-objective)), checked against the
first version's bisection over whole sims: 611 against 611 for the paladin and 578 against 581 for the
warrior at Onyxia, within the bisection's step.

**Sometimes no healing is enough.** A tank one swing can kill from full, or one a burst kills between two
heals, dies whatever the healing. So the bench also reports:

- **The biggest hit** (the sheet's) and how many of them the tank's health takes. Under 1 is a one-shot.
- **The death floor**: the share of iterations the tank dies in even when every heal tops it off, so only
  damage between two heals kills it. Over 5%, healing needed is "none works", and the build isn't viable
  against that boss with that healer cadence.

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

## The optimizer's objective

From a metrics review (2026-11-02, a subagent that read the bench and measured its noise), adopted by the
user as proposed the same day, and built.

**Score, in raid DPS: threat members × SafeDPS − w × DPS slot × healing needed / one healer's HPS.**

- **Healing needed stays the survival number.** DTPS gives Stamina no value, effective health ignores
  avoidance, TMI is a stand-in for what healing needed measures directly, and the chance of death at the
  encounter's own healing is a cliff (mostly 0% or near 100%) with little slope for a search to follow.
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

**How it's built:**

- **The engine** (`sim/core/tank_score.go`): a player with `tank_scoring` (`proto/common.proto`) and the
  stand-in healer records each iteration's damage taken and heals, and at its end replays them to find the
  least stand-in HPS the iteration lives through (+Inf when none does). It also follows the tank's threat on
  the boss for SafeDPS: the reference member starts 2 s after the pull, makes 0.7 threat a point of damage
  (Salvation) and pulls at 110%, all settable. Both land in `UnitMetrics.tank_score`, and
  `core.TankScoreSummary` turns the list into healing needed at a death share, its standard error (from the
  order statistics) and the death floor. A test checks that the replay agrees with the run itself: the
  iterations the run died in are exactly those whose least HPS is over the healer's.
- **The objective** (`tools/tankscore`): the encounter files, setting a request up as a tank, and the score
  with its standard error, shared by the bench and rotopt. A build no healing saves at 5% scores far below
  any other (−1,000,000, lower the more often it dies), so a search still climbs out of it.
- **rotopt** `-tank max-threat|balanced|max-survival` (or a weight) scores the player as a tank in place of
  DPS, so every search (rotation knobs, gear, talents) maximizes it, against `-tank-encounter` (whose fight
  length is the default `-durs`) or a 600 HPS healer (`-tank-hps`). `-tank-threat-members`,
  `-tank-dps-slot` and `-tank-death` set the rest.

**First search** (2026-11-02): rotopt's talent search on the warrior at Broodlord, Balanced, 600 iterations a
candidate and 3,000 to confirm (248 sims). It kept the build. Its best swap, Improved Thunder Clap's 3 points
into Cruelty for more threat (SafeDPS 1,487 to 1,517), gained 73 ± 43, short of the search's two standard
errors. **Healing needed is noisy where one ability makes the bad fights**: ±15 HPS at 3,000 iterations at
Broodlord (Mortal Strike's healing cut), against ±3 at Onyxia. Two candidates share the random seed, so their
difference is far less noisy than the two errors added up, which is what the search's test assumed.

**The paired test** (2026-11-02): every rotopt search (knobs, talents, gear, enchants) now tests a tank
candidate's gain against the error of its difference from the best, iteration by iteration
(`tankscore.PairedDiff`). Two runs share their seed and iterations, and the engine keeps each iteration's
SafeDPS and least HPS in seed order (`TankScore.safe_dps_each`, `survive_hps`), so iteration k of two runs had
the same rolls up to where the builds differ. The error is a bootstrap: 400 resamples of the iterations, the same
ones for both runs. It falls back to the two errors added up when the runs don't pair, over several fight
lengths, or near the death floor (more than a tenth of the resamples with no healing that saves the tank).
Checked against the real spread of the difference over 10 seeds (1,000 iterations, the warrior, the Anniversary
files):

| Encounter | Builds | Spread over seeds | Paired error | Errors added up |
|---|---|---|---|---|
| Onyxia | without Strength of Earth Totem | 0.7 | 1.2 | 57.8 |
| Chromaggus | without Strength of Earth Totem | 0.5 | 0.8 | 34.9 |
| Onyxia | without Power Word: Fortitude | 12.9 | 20.4 | 57.1 |
| Chromaggus | without Power Word: Fortitude | 22.5 | 19.4 | 39.9 |

So the search's noise drops 2 to 50 times, least where survival moves. At Broodlord the pre-raid warrior sits on
the death floor (4.5%), where healing needed has no finite error and nothing pairs: a search there needs a
sturdier tank or more healing.

**First search with it** (2026-11-02): the warrior's talent search at Chromaggus (Anniversary), Balanced, 600
iterations a candidate and 3,000 to confirm (789 sims, 73 s). It took Improved Thunder Clap's 3 points into Cruelty
(+66, noise 8: the swap the Broodlord search couldn't tell from noise), then Improved Shield Wall's 2 (+22, noise
7.7), to `35-05-251530120331201351`. **Both moves are the sim's gaps, not findings yet**: the bench's requests give
the boss Improved Thunder Clap as a raid debuff whatever the tank takes, and a scoring run has no health-based
defensives, so Shield Wall and Last Stand can't pay. Many talents score exactly the same (Iron Will, Lingering
Rage, Booming Voice), the sim not reading them. So the rules audit (Next steps, 1) comes before the presets'
searches: the debuffs a tank brings itself have to come from its own talents.

Not modelled yet, which the score would reward once it is: threat drops from boss abilities (Knock Away),
and a tank's own health-based effects, which the replay can't follow (health potions, defensives at a health
share, APL conditions on health). A scoring run leaves those out.

## Open decisions

None for the objective. The weights of the presets get checked once the gear search runs on them (two
presets landing on the same build means a weight moves).

## Findings so far

- **The DPS Report's requests give the raid Salvation** (`raid.buffs.salvation`, −30% threat). Copied into a
  tank's request it cut the warrior from 1137 to 794 TPS. The bench's requests drop it.
- **The raid sim's 700 TPS stand-in** came from the warrior's tank suite. On the report's boss with no
  Salvation the same build does 1137. Check which buffs the 700 had before the raid sim's simmed tanks
  replace it.
- **The Protection paladin's rotation is stale.** `basic_prot` casts Holy Shield only while aura 20134 is
  up, Redoubt (a talent proc on 10% of melee hits taken under Forever, `sim/paladin/talents.go`), so Holy
  Shield is up about a third of the time. Its options need Righteous Fury on (`righteousFury`), which the
  UI preset and the test suite have and the bench's first request missed.
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
2. **Encounters**: Patchwerk, threat drops, armor cuts and stacking DoTs in the engine.
3. **Rotations**: the paladin's and the Warden's first, then rotopt templates for all four.
4. **The three presets' searches** (the paired test is in, above).
5. **Tank gear**: a tank row in `orchestration/gear_targets.md` and synthetic tank items, then gear and talent
   searches on the score.
6. **The report**: a tank section with TPS next to damage taken, healing needed and the sheet's thresholds.
7. **The raid sim**: the simmed tank's threat in place of the 700 TPS stand-in, then the comp search.

## Audits

Four auditors checked the work on 2026-11-02, each starting cold, and their findings are fixed:

- **Engine:** the tank sheet matches the rolled outcomes within noise over 60,000 swings a case, the armor
  cap moves no existing result, and `classic_crit` works end to end. Fixed: a boss ability's blocked crit was
  booked twice, its miss took the white swing's dual wield 19%, a dual wielding boss crashed the sim (no off
  hand attack table for enemies, older than this work), a ticking event dropped its debuff, and the biggest
  hit left out flat damage and damage taken. Open: the cores may roll a melee ability's block after the
  avoidance, not on the same roll.
- **Bench and log miner:** the bisection is valid (the chance of death moves only with the HPS, checked over
  sweeps). Fixed: the death floor at 20 times DTPS didn't always top the tank off (now 1,000), "none works"
  at a reachable target, an HPS of 0 starting the engine's presim healer, the encounter dropping the
  request's execute phases, half-set healer flags, and in the miner the boss's events needing
  `hostilityType: Enemies`, results lost when the budget ran out mid-sample, and absorbs left out of the
  healing ratio.
- **Encounter data:** every spell value matches the Classic Era client's spell tables, and the timers the
  vmangos scripts. Fixed: Broodlord's weapon damage (above), Vaelastrasz's stacking breath, Double Attack's
  and Elemental Fire's rates, Demoralizing Shout on weapon damage abilities, first casts, Onyxia's notes.
- **This doc** against the code: every number matched. Fixed: the resistance formula, the swing range of the
  report's boss, the biggest hit's scope, the uncrushable test's crit, aura 20134, stale lines.

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
