# tank_fairness pass 4: 2026-11-05, master a079415

Due: 33 watched files changed since dc303b0: tank rounds 4 and 5 (full raid consumables under standards 40's new
sentence, talents and knobs re-searched on them), the Warden UI's encounter, the DPS Report re-sim at 877bcec8c.
Raw table: `ref/tank_fairness/summary.tsv` (72 rows: 12 preset requests × 6 scoring bosses, seed 1, scores at all three weights).

## How

- `go build -tags=with_db -o $AUDIT_WORK/tank_fairness/tankbench ./tools/tankbench`, then `tools/audit/tank_fairness/run.sh`
  (every boss file but Ragnaros, 3,000 iterations, seeds 1 and 7) and `summ.py s1 x`. Six-boss mean error = sqrt(Σerr²)/6.
- Same terms: `tools/audit/tank_fairness/terms.py` flattens all 12 requests (gear, talents, rotation set aside) and prints what differs.
- Variants (scratch, `$AUDIT_WORK/tank_fairness/var/`): each paladin preset with the Greater Stoneshield Potion and an explicit
  potion action, with the Demonic Rune, with the Mageblood Potion and with the Elixir of Greater Intellect; seeds 1 and 2.
- UI: consumable enums per presets.ts against the requests; the Warden's sim.ts defaults. DPS Report: tank_builds.json's
  talents, races and scores against the requests and this run.

## Own-weight six-boss means (3,000 iterations, seed 1 ± error [seed 7])

| preset | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| Max threat (w 1/3) | 4649.8 ± 5.0 [4647.5] | 3248.5 ± 3.2 [3240.7] | 3490.8 ± 5.1 [3469.7] | 3168.7 ± 4.2 [3164.8] |
| Balanced (w 1) | 3867.3 ± 7.0 [3841.5] | 2825.7 ± 5.2 [2824.5] | 2500.3 ± 7.5 [2472.0] | 1794.0 ± 6.2 [1776.8] |
| Max survival (w 3) | 2150.7 ± 13.6 [2144.5] | 1921.8 ± 14.3 [1888.7] | 232.5 ± 11.1 [214.3] | -1993.7 ± 17.5 [-2000.3] |

All within about 2× error of the DPS Report's tank_builds.json (re-simmed at 877bcec8c, 1,000 iterations); its talents
and races match all 12 requests. Pass 3's open item (the report still on round 2) is closed.

Since pass 3, own weight: warrior +950 / +855 / +947, paladin +448 / +408 / +473, bear +646 / +651 / +888,
Warden +747 / +1046 / +2003. Max survival's order flips: warrior first, paladin second (finding -20261105-1 flips it back).

## Does each preset win its own weight (rows: preset; columns: w = 1/3, 1, 3)

| | 1/3 | 1 | 3 |
|---|---|---|---|
| warrior MT / Bal / MS | **4650** / 4466 / 3903 | 3813 / **3867** / 3465 | 1303 / 2073 / **2151** |
| paladin MT / Bal / MS | **3249** / 3240 / 3077 | 2819 / **2826** / 2789 | 1531 / 1582 / **1922** |
| bear MT / Bal / MS | **3491** / 3349 / 2964 | 2365 / **2500** / 2281 | -1012 / -44 / **233** |
| Warden MT / Bal / MS | **3169** / 3077 / 2956 | 1704 / **1794** / 1718 | -2690 / -2055 / **-1994** |

Every preset wins its own weight on both seeds. The paladin's Max threat and Balanced are one build since round 5
(talents and gear the same; only Spirit of Zanza against Scorpok and the rotation knobs differ). They tie at their
weights: +8.3 ± 4.5 at 1/3 (seed 7 +3.0) and +6.7 ± 7.8 at 1. This is documented as the paladin's call 14
(docs/tank_paladin.md:582-584), a measured result, so it isn't raised. The paladin's Max survival is its own build again.
The warrior's and Warden's Max survival now beat Balanced at w=3 beyond the noise (+78 ± 22, +61 ± 24).

## Same terms

All 12 requests are equal on these:
- level, reaction time, distance, position;
- professions;
- all raid buffs and debuffs, with the tank's own stripped;
- encounter and options, no Salvation;
- Mongoose, Phalanx, Greater Fortitude, both Jujus;
- Dragonbreath Chili, Rumsey Rum, the Goblin Sapper.

What differs is per-class picks, each A/B'd in round 4 or 5:
- **Flask:** Natural Precision everywhere except paladin Max threat and Balanced, which take Supreme Power.
- **Food:**
  - Dirge's chops on every Max survival and on all three Warden presets.
  - Nightfin Soup on paladin Max threat and Balanced.
  - Dumpling elsewhere.
- **Zanza slot:**
  - R.O.I.D.S. on warrior and bear Max threat.
  - Scorpok on paladin and Warden Max threat.
  - Spirit of Zanza elsewhere.
- **Potion:**
  - Mighty Rage on warrior and bear Max threat and Balanced.
  - Greater Stoneshield, with a potion action, on warrior and bear Max survival and all three Warden presets.
  - **Major Mana Potion on all three paladin presets**, with no potion action (below).
- **Mana consumables:** Greater Arcane on the paladin and Warden; the Warden alone takes the Demonic Rune, Greater Intellect, Nature Power and Mageblood (not at Max survival).
- **Weapon:** Elemental stone on the warrior and paladin, Dense weightstone on the bear, Rockbiter on the Warden.
- **Race:** Undead for the warrior and paladin, Night Elf for the bear, Night Elf / Night Elf / Undead for the Warden.
- **Wisdom:** the Warden only, as before.

The one unequal term: `autocastOtherCooldowns` never drinks a defensive potion in a tank score. The other three tanks
found this and added an explicit potion action. The paladin's search tried the Stoneshield without one, and read 0.0 ± 0.0 as "never drinks one".

Paladin variants, own weight, seed 1 / seed 2:

| variant | Max threat | Balanced | Max survival |
|---|---|---|---|
| Greater Stoneshield + potion action | +57.8 / +58.8 | +171.6 / +174.6 | +309.2 / +318.2 |
| Demonic Rune | +1.7 / +1.6 | +3.0 / +3.3 | +2.2 / -1.2 |
| Mageblood | -0.8 / -0.7 | -0.9 / +0.1 | -1.3 / -0.7 |
| Greater Intellect | +0.5 / +1.0 | +2.0 / +2.0 | +2.5 / -1.0 |

The Stoneshield result is finding -20261105-1 (high). It takes the paladin's Max survival to 2,231 against the warrior's
2,151. The mana consumables are noise: the paladin isn't mana-bound, so leaving them out is fair.

## Spread at Balanced (seed 1, per boss, against the mean of the other three)

| boss | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| Magmadar | 4200 (+51%) | 3208 (+3%) | 2808 (-14%) | 2342 (-31%) |
| Golemagg | 4041 (+58%) | 2961 (+1%) | 2666 (-12%) | 2070 (-36%) |
| Onyxia | 3990 (+65%) | 3369 (+28%) | 2289 (-23%) | 1616 (-50%) |
| Broodlord | 3588 (+75%) | 2221 (-11%) | 2557 (+7%) | 1370 (-51%) |
| Vaelastrasz | 4463 (+42%) | 3433 (-1%) | 3195 (-10%) | 2796 (-24%) |
| Chromaggus | 2922 (+130%) | 1762 (+6%) | 1487 (-15%) | 570 (-72%) |

| Balanced, six-boss mean | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| TPS | 2200 | 1534 | 1760 | 1648 |
| DTPS | 376 | 314 | 584 | 676 |
| Healing needed | 317 | 245 | 457 | 692 |

The full raid set narrowed the gaps (the Warden's healing need was +236% over the others at worst in pass 3; now +100% on the mean).

Explanations:
- **Warrior score (+42% to +130%) and TPS (+34% over the others' mean):** Revenge, Heroic Strike and Shield Slam, all sourced. Taken as real.
- **Paladin healing needed (lowest; -50% against the others' mean):** finding -2. The mechanics are checked; three user calls remain.
  - With the Stoneshield, it would drop to 188.
- **Paladin TPS (-18% under the others' mean):** still unexplained; open question in fairness_tanks.md.
- **Warden (last everywhere):** a class gap by ruling (standards 40).

The Anniversary logs are unchanged: warrior main tanks, druids on four bosses, no paladin or shaman. The sim ranks the
paladin second at Balanced and (once fixed) first at Max survival, which the logs don't show. The likely reasons are
finding -2's calls and the population.

## UI = bench

The consumable enums in each presets.ts are the requests' set. Finding -11 is closed:
- the Warden's sim.ts now loads the bench encounter and Night Elf;
- its default gear, APL and talents are the tank build.

## Findings

- **New:** -20261105-1 (high): the paladin never drinks its potion. Greater Stoneshield with a potion action gains
  +58 / +172 / +309 at Max threat / Balanced / Max survival.
- **Resolved:**
  - -11: the Warden UI's default encounter.
  - -10: every tank searched every consumable slot.
  - -9: the last stale figure is now only a round-3 column.
- **Still open:**
  - -2: the paladin's damage taken; waits on the user's calls.
  - -5: the legacy warrior file; the healing session's call.
