# Tank sim

Tank presets built like the DPS Report's, scored on threat and on survival. Tank optimization trades
TPS against damage taken, so the sim has to sim the boss hitting the tank and say how close the tank came
to dying. This comes before the raid sim's simmed tanks ([raid_sim.md](raid_sim.md), Threat), which
replace its 700 TPS stand-in.

**Where it stands (2026-11-02):** the engine reports a tank sheet with the thresholds tanks gear for, and
`tools/tankbench` sims tank builds against the report's boss with a stand-in healer. Two builds run
(Protection warrior and paladin, both on Classic pre-raid plate). Next comes the rules audit, then tank
gear and rotations. The open decision is how a build trades threat for survival (see [Open
decisions](#open-decisions)).

## What tanks gear for

From the Classic, TBC and Season of Discovery tanking guides and wowsims' code (sources at the end), and
what our engine does under Forever. Chances are against a level 63 boss.

| Threshold | Classic | What it means | Forever in our sim |
|---|---|---|---|
| Crit immune | 440 Defense (5.6% crit, 0.04% a point) | crits do double damage, the main spike | **315 Defense** (the boss crits 0.6%, see below) |
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

**Our sim gives a level 63 boss 0.6% crit against a player, not Classic's 5.6%** (`sim/core/target.go`,
`newWeaponAttackTable`). The rule comes from the user's leveling logs. In them, about 2,900 swings from mobs
at or under our level never crit, and mobs 1 to 5 levels above crit 4 times in 745 swings. Raid bosses
aren't in Forever yet (no boss data before 2026-12-09).

So crit immunity costs 15 Defense instead of 140, which takes most of Defense's value away and changes
every tank gear set. **Need to Verify**: an elite or a boss's crit rate in the logs before tank gear is
tuned on it. Until then the bench can be run both ways.

## The tank sheet

`UnitMetrics.tank_sheet`, for a player the boss attacks (`sim/core/tank_sheet.go`). It's taken once, as the
first iteration's prepull starts: permanent buffs and consumables on, no cooldowns or procs. It uses the
same attack table the boss's swings roll on.

- **The boss's swing at us**, on its one roll table in order: miss, dodge, parry, block, crit, crush, hit
- **Defense**, and the Defense at which crits stop
- **Coverage**: miss, dodge, parry and block added up, and the **crush gap** still needed to push crushing
  blows off the table
- **Armor** and its reduction, **block value**, **health**, **effective health**
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
- **The stand-in healer** heals a steady `-hps` (default 600) every `-cadence` seconds (2 ± 1). It's the
  engine's healing model (`sim/core/health.go`), the same trick as the raid sim's 700 TPS tank.
- **It prints** TPS, DPS, DTPS, TMI (6 s windows), the chance of death, the sheet, and how the boss's swings
  landed over the fight. `-actions` adds threat by ability.

**First numbers** (2026-11-02, 2,000 iterations, 180 s, 600 HPS). Both builds wear the Classic Valor set,
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

**Proposed: "healing needed"**, the stand-in healer's HPS at which the tank dies in 5% of fights, found by
bisection. It's one number per build that doesn't depend on picking a healer, and it counts both damage
taken and spikes (a spiky build needs more healing to stay at 5%). TMI and DTPS stay in the report next to
it. It costs about 6 sims a build instead of 1.

## Open decisions

1. **How a build trades threat for survival.** (a) Weights, with 2 or 3 presets along the curve ("Max
   threat", "Balanced", "Max survival"), like the PvP High HP preset's `-health-weight`. (b) A death chance
   floor. (c) Survival only. Recommended: (a). The user ruled against floors for PvP.
2. **The survival score**: healing needed (recommended), or the chance of death at a fixed healer.
3. **Forever's boss crit**: keep 0.6% until a log shows otherwise, or tune tank gear on Classic's 5.6%.

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
   Bring the Feral tank into the build. Check Forever's boss crit.
2. **Rotations**: the paladin's and the Warden's first, then rotopt templates for all four.
3. **The survival score** in the bench and the search tooling.
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
