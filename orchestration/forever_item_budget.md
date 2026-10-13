# How Forever (and SoD) price item stats

User ruling 2026-10-13: only Forever items of item level 30 and under count (the beta caps at level 30; items above may still carry stale, untuned data). The section "ilvl <= 30 only" at the end supersedes sections 1-3 and the proposal below, which used all items.

Measured 2026-10-13 from the Forever beta item list (assets/db_inputs/forever_wowhead_items.json: Wowhead's Forever listing, 12,114 items) and wowsims/sod's item DB. Scripts in orchestration/tools/ (primary.py, sod_primary.py, fit.py, single.py, epic.py, linear_total.py).

## 1. Not linear: stats combine with an exponent of about 1.5 to 2

Items that carry only primary stats (Str/Agi/Sta/Int/Spi, all weight 1) need no weights to test. Grouping them by quality, slot and item level, the right formula gives every item in a group the same budget whether it has one stat or four. RMS spread within groups:

| Exponent p | Forever ilvl 1-35 (beta-tuned) | Forever ilvl 36-60 | Forever ilvl 61+ | SoD's new items | Classic items |
|---|---|---|---|---|---|
| 1.0 (linear) | 12.3% | 12.0% | 9.8% | 13.5% | 13.6% |
| 1.5 | 7.5% | 7.8% | 5.8% | 7.6% | 8.4% |
| 1.75 | 7.5% | 7.6% | 5.8% | 6.9% | 8.0% |
| 2.0 | 8.1% | 7.8% | 6.3% | 6.9% | 8.2% |

Linear is the worst fit everywhere. The classic 1.5-power norm, cost = (sum (w x v)^1.5)^(2/3), or something slightly steeper, is what Forever and SoD use. The rest of the spread is integer rounding on small items.

## 2. Weights and slot modifiers (fit over 3,310 armor items with primaries, crit, hit, spell power, AP, MP5)

- Slot modifiers: head / chest / legs 1.0; shoulder / hands / waist / feet 0.74-0.77; neck / back / wrist / finger 0.55-0.58. Matches standards item 1 (0.77, 0.5625).
- Crit rating about 1.0 per point and hit rating about 1.1 per point (standards: 1.0), AP 0.55 (standards 0.5), MP5 about 2.3-2.4.
- Spell power about 0.86-0.97 per point, NOT 0.55. Check: at 0.55, Forever's caster epics come out 2% under budget and physical epics 5% over; at 0.86 (the classic spell damage and healing price) both land on budget (1.000 and 1.006). The few single-stat spell power items agree (0.8-1.0).

## 3. The budget in practice (what our linear rule should match)

Real Forever epics (ilvl 55-80, median 3-4 stats per item) carry a weighted linear total, per slot modifier, of about 1.28 x ilvl - 10.9 (with spell power at 0.86):

| ilvl | Real Forever epics | Ours (standards 1a: 1.73 x (1.2 ilvl - 36.4)) |
|---|---|---|
| 66 (head, shoulders, chest, hands, legs, feet) | 73.5 | 74.0 (+1%) |
| 71 (neck, back, wrist, waist, rings, one-handers) | 79.9 | 84.4 (+6%) |
| 76 (two-handers) | 86.3 | 94.8 (+10%) |

Real items with the 1.5-norm carry 1.42-1.47x their 1.5-norm value in linear stats (3-4 stats per item), not 1.73x.

## Proposed change (pending the user's go-ahead)

Keep linear spend (it keeps sets comparable), but calibrate it to Forever: each piece's weighted linear total = slot mod x (1.28 x ilvl - 10.9), and spell power weighs 0.86 per point.

## ilvl <= 30 only (user ruling 2026-10-13)

Data: the Forever listing's green and blue items of item level 5-30 (ilvl 1-4 rows are placeholders and scaling items; the Season of Discovery records with ids 200000-249999 are left out). That is 1,555 clean items, meaning items with only stats the model prices and no proc or on-use effect in their Classic tooltip: 1,104 armor (681 with Classic ids, 423 new to Forever), 450 weapons, shields and off-hands, and 1 trinket. There is no epic with stats at ilvl 30 or under, so every number below comes from greens (q2) and blues (q3). The Era baseline is Wowhead's Classic gear planner dump (assets/db_inputs/wowhead_gearplannerdb.txt, versionNum 11300), which uses the listing's stat keys.

- wago_db2_items.csv is not the Forever client. It is ItemSparse of build 1.15.3.55646, the 2024 Era/SoD client. Its primaries match Era on 948 of 948 matched items and Forever on only 715. It confirms the Era baseline but cannot check Forever.
- Spell power keys in the Forever listing: `splpwr` is damage and healing. A `splheal` beside it is extra healing on top: Naga Battle Gloves have splpwr 5 and splheal 10, and the tooltip says "healing by up to 15 and damage by up to 5". The all-items scripts dropped that extra healing and skipped the spldmg/splheal pairs. Here both are counted: `sp` covers damage and healing, `heal` is the extra healing.
- Scripts are in orchestration/tools/: b30_data.py (loader), b30_exponent.py (Q1), b30_fit.py (Q2-Q4), b30_era.py (Q5) and b30_level60.py (Q6). Run each one with python3 from any directory.

### 1. Exponent: linear is ruled out, p = 1.5 to 1.75

The test uses items with primaries only, grouped by quality, slot and ilvl. The table shows the RMS spread within groups that mix stat counts, and the spread whole-point rounding alone would cause:

| p | Forever armor ilvl 5-30 (426 items, 93 groups) | Forever armor ilvl 20-30 (249) | Forever weapons (130) | Era armor ilvl <= 30 (416) |
|---|---|---|---|---|
| 1.0 (linear) | 12.2% | 12.0% | 16.6% | 12.1% |
| 1.25 | 8.5% | 7.5% | 11.8% | 8.3% |
| 1.5 | 7.4% | 5.3% | 9.7% | 6.9% |
| 1.75 | 7.6% | 4.9% | 9.2% | 6.9% |
| 2.0 | 8.3% | 5.4% | 9.3% | 7.4% |
| rounding alone | ~8.1% | ~4.9% | ~11.1% | ~7.4% |
| best p (0.05 grid) | 1.55 | 1.70 | 1.80 | 1.65 |

- Multi-stat premium: in the same group, 2-stat items carry 1.25x the linear total of the single-stat items (median, n=175) and 3-stat items carry 1.50x (n=12). Linear pricing predicts 1.00. An even split at p=1.5 predicts 1.26 and 1.44.
- A fit over all 1,104 armor items, with every stat and fitted weights, trimmed at 3 sigma, gives an RMS log error of 17.4% at p=1, 14.6% at 1.25, 12.5% at 1.5, 11.9% at 1.75 and 11.8% at 2.
- Verdict: linear is ruled out. From p=1.5 up, the spread is already at the rounding floor, so 1.5 and 1.75 cannot be told apart. Era's ilvl <= 30 items give the same answer. The 33 primary-only items new to Forever lean steeper (best p 2.35), but there are too few to matter.

### 2. Stat weights: spell power 0.86 confirmed, 0.55 refuted

Fit at p=1.5 over armor ilvl 5-30, with 1,029 of 1,104 items kept after the 3-sigma trim. Ranges are 90% bootstrap intervals; n counts the items that carry the stat.

| stat | all armor | Classic ids | Forever-new ids | ilvl 15-30 | Era level 60 (blues 45-70 + epics 55-80, 1,083 items) |
|---|---|---|---|---|---|
| spell power (dmg + heal) | **0.86** (0.82-0.90), n=176 | 0.93 (0.88-0.97), n=77 | 0.80 (0.73-0.85), n=98 | 0.86 (0.82-0.89), n=158 | 0.84 (0.83-0.85), n=277 |
| extra healing | 0.40 (0.33-0.47), n=50 | 0.46, n=10 | 0.39, n=38 | 0.36 | 0.45, n=97 |
| school spell power | 0.60 (0.56-0.62), n=57 | 0.70, n=9 | 0.60, n=48 | 0.64 | 0.71, n=28 |
| AP | 0.49 (0.44-0.53), n=17 | 0.52, n=8 | 0.39, n=9 | 0.52 | 0.52, n=72 |
| crit rating (14 = 1%) | 0.79, n=1: not measurable | - | n=1 | n=1 | 1.01 (0.99-1.02), n=171 |
| hit rating (10 = 1%) | 0.74, n=1: not measurable (weapons: n=3, 1.02) | - | n=1 | n=1 | 1.09 (1.03-1.15), n=79 |
| RAP | n=1 (a gun): not measurable | | | | n=2 |
| MP5 | 1.8, n=5: weak (1.0-2.5 across subsets) | | | | 2.39 (2.33-2.44), n=113 |

- Swaps, a direct check: 35 armor items swapped exactly one primary for spell power at the same ilvl and quality and changed nothing else. Since value is kept for any p exactly when w_sp x SP = the old amount, rounding to whole points bounds w_sp. A weight of 0.55 fits 0 of the 35, 0.86 fits 24 and 0.93-1.0 fits 29. Examples: Minor Channeling Ring went from 4 Spirit to 5 SP, High Robe of the Adjudicator from 7 to 8, Warsong Sash from 9 to 11.
- Weapons are priced differently. A fit over weapons only gives sp 0.32 (0.19-0.42, n=45): Forever's caster weapons carry about 2.7x the spell power the armor price would give (Rod of the Sleepwalker has 32 SP at ilvl 29). Mixing weapons in pulls sp to 0.66, so the armor value applies to armor and jewelry.
- Verdict: spell power weighs 0.86 (0.80-0.93 depending on the subset). The ilvl 30 and under data cannot price crit, hit or RAP, because only one to three items carry each. Era level 60 prices crit at 1.0 per point (14 = 1%), hit at 1.1 per point (10 = 1%), AP at 0.52 and MP5 at 2.4.

### 3. Slot modifiers

| slot | Forever armor ilvl 5-30 (n) | Forever ilvl 15-30 | Era level 60 | standards |
|---|---|---|---|---|
| chest | 1 | 1 | 1 | 1 |
| head | 1.07 (1.05-1.10, 44) | 1.02 | 1.00 | 1 |
| legs | 1.00 (156) | 1.00 | 1.00 | 1 |
| shoulder | 0.74 (41) | 0.74 | 0.745 | 0.77 |
| hands | 0.72 (127) | 0.74 | 0.745 | 0.77 |
| waist | 0.75 (117) | 0.76 | 0.745 | 0.77 |
| feet | 0.74 (125) | 0.74 | 0.75 | 0.77 |
| neck | 0.56 (10) | 0.53 | 0.55 | 0.5625 |
| back | 0.555 (85) | 0.54 | 0.55 | 0.5625 |
| wrist | 0.58 (90) | 0.55 | 0.55 | 0.5625 |
| finger | 0.565 (45) | 0.55 | 0.55 | 0.5625 |

- Weapons, from a fit over armor plus weapons (chest = 1): two-hand 1.10 (n=123), one-hand 0.42 (87), main-hand 0.42 (56), held-in-off-hand 0.56 (26), ranged 0.31 (61), thrown 0.33 (5). The standards use 1.0, 0.42, 0.42, 0.5625 and 0.32.
- Trinkets have no data at ilvl 30 and under: the only stat-only trinket is Worgenbane Talisman (2 MP5). Era's 7 stat-only blue and epic trinkets (ilvl 60-83) sit at 0.70-0.75 of the chest budget on the 1.5-norm. Because they carry only one or two stats, their linear totals are 0.36-0.62 of the multi-stat epic trend (the 3 epics: 0.48, 0.60, 0.62).
- Verdict: the three tiers hold. The middle tier is 0.72-0.76, so the standards' 0.77 is about 3% high. Jewelry, back and wrist sit at 0.55-0.58.

### 4. Budget per item level at ilvl <= 30

| | fit (chest equivalent) | ilvl 20 | ilvl 25 | ilvl 30 |
|---|---|---|---|---|
| green, 1.5-norm V (n=728) | 0.458 ilvl - 2.98 | 6.2 | 8.5 | 10.8 |
| blue, 1.5-norm V (n=301) | 0.509 ilvl + 0.75 | 10.9 | 13.5 | 16.0 |
| green, linear total / slot mod (1.74 stats per item, rms 14%) | 0.619 ilvl - 4.95 | 7.4 | 10.5 | 13.6 |
| blue, linear total / slot mod (2.76 stats per item, rms 19%) | 0.783 ilvl - 1.01 | 14.6 | 18.6 | 22.5 |

The linear totals use the section 2 weights and the standards' slot mods. Linear total divided by 1.5-norm value, by stats per item: 1.00 for 1 stat, 1.25 for 2, 1.42 for 3, 1.57 for 4 and 1.70 for 5. Level 60 epics average 3.9 stats per item and come out at 1.51 (section 6), against the 1.73 in standards 1a.

### 5. Forever vs Era for the same items: the low-level budget is unchanged

- There are 948 Classic-id items, clean on both sides. Of these, 647 (68%) are identical. 277 have changed stats, 123 changed ilvl (mostly +2), 66 went from green to blue, and 58 that had no stats in Era gained some (mostly weapons gaining spell power). No item lost all its stats.
- Same id, same ilvl and quality: the median value ratio Forever / Era is 1.000 in every ilvl band (5-15, 16-20, 21-25, 26-30) for both qualities. The mean is 1.03 for greens and 1.06 for blues, driven by the added stats.
- Items Forever moved in ilvl or quality were rebudgeted to match: on Era's curves they sit at 0.99 of budget before and 0.98 after (n=96 armor).
- Most changes are stat swaps: 19 Spirit to spell power, 8 Int to spell power, 5 Sta to spell power, roughly point for point (section 2).
- Curves fitted with the same weights and slot mods, Forever / Era:

| | ilvl 15 | ilvl 20 | ilvl 25 | ilvl 30 |
|---|---|---|---|---|
| green, 1.5-norm V | 1.05 | 1.01 | 0.99 | 0.98 |
| blue, 1.5-norm V | 1.03 | 1.02 | 1.02 | 1.01 |
| green, linear total | 1.00 | 1.01 | 1.02 | 1.02 |
| blue, linear total (Era has only 71 blue armor items at ilvl 30 or under) | 1.17 | 1.09 | 1.05 | 1.03 |

- Can a line fitted at low levels reach level 60? Era's own line fitted on ilvl 15-30 predicts its ilvl 60 items within 3%: greens 32.1 against an actual 33.2, blues 46.5 against 48.1.
- Verdict: Forever did not change the low-level budget (within about 3%). It reworded stats (mostly Spirit to spell power), raised some items' ilvl or quality with a matching budget, and added stats to stat-less items. Taking the level 60 budget from Era's level 60 items is justified to about +/-3%, plus the scatter of real items.

### 6. Level 60 epic budget for the synthetic gear

The source is Era's epics at ilvl 55-80: 331 clean armor items with 3.9 stats each. The linear total per slot mod is fitted against ilvl:

| weights | trend | ilvl 66 | ilvl 71 | ilvl 76 |
|---|---|---|---|---|
| **proposal: sp 0.86, AP 0.5, crit 1.0, hit 1.0, MP5 2.4, healing 0.45** | **1.125 ilvl + 2.0** | **76.3** | **81.9** | **87.5** |
| ilvl <= 30 weights (sp 0.86, MP5 2.0, healing 0.40) | 1.042 ilvl + 6.6 | 75.4 | 80.6 | 85.8 |
| Era level 60 fitted weights (sp 0.84, crit 1.01, hit 1.09) | 1.131 ilvl + 1.8 | 76.4 | 82.1 | 87.7 |
| sp 0.55 (standards weights) | 1.042 ilvl + 4.6 | 73.4 | 78.6 | 83.8 |
| now: standards 1a, 1.73 x (1.2 ilvl - 36.4) | | 74.0 (-3% vs proposal) | 84.4 (+3%) | 94.8 (+8%) |
| all-items proposal, 1.28 ilvl - 10.9 | | 73.6 (-4%) | 80.0 (-2%) | 86.4 (-1%) |

- Uncertainty:
  - Item scatter is 12% RMS around the trend.
  - The ilvl range matters most at 76. Fitting ilvl 60-80 or 58-76 changes nothing (75.3-75.4 / 80.3-80.6 / 85.3-85.9 at the ilvl <= 30 weights). Fitting the Phase 1 range 55-71 flattens the line to 74.0 / 76.9 / 79.9.
  - Forever / Era at ilvl 30 and under adds +/-3%.
  - Spell power weight 0.80-0.93 moves the trend by +/-1%.
  - Overall: -5% to +3% at ilvl 66, -7% to +3% at 71 and -10% to +3% at 76. Each low end is the Phase 1 range fit; two-handers sit at the edge of Phase 1 data.
- On the 1.5-norm, Era epics have V = 0.731 ilvl + 3.07 (51.3 / 55.0 / 58.6, against the standards' V of 42.8 / 48.8 / 54.8) and carry 1.51x that in linear stats. The standards' lower V times 1.73 lands near the right total at 66 but climbs too steeply.
- Caveat, casters vs physical: real caster epics carry 4.4 stats per piece and physical ones 3.4. On the 1.5-norm that gives casters more linear stats: at sp 0.86 they sit at 1.04 of the common linear trend and physical pieces at 0.93 (caster / physical 1.12). A linear sp weight of 0.50-0.55 makes them equal (1.00-1.02). That is the stats-per-piece effect standards 1a removes on purpose, not Forever's price for spell power. So with one linear total for everyone and sp 0.86, caster sets get about 4% fewer linear points than real Era caster epics of the same ilvl, and physical sets about 7% more.
- Trinkets at 71: Era's stat-only epic trinkets carry 0.48-0.62 of the trend in linear stats (n=3). Jewelry's 0.5625 x 81.9 = 46 linear points per full-budget stat trinket sits in the middle of that range.

**Recommendation:** spend each piece linearly at slot mod x (1.125 x ilvl + 2.0), with spell power at 0.86 per point (healing on top 0.45, AP 0.5, RAP 0.4 unmeasured, crit and hit rating 1.0, MP5 2.4) and slot mods 1 / 0.75 / 0.5625. Per piece that gives:

| piece | linear total |
|---|---|
| ilvl 66, head / chest / legs | 76.3 |
| ilvl 66, shoulder / hands / feet | 57.2 |
| ilvl 71, waist | 61.4 |
| ilvl 71, neck / back / wrist / ring | 46.1 |
| ilvl 71, physical one-hander (x 0.42) | 34.4 |
| ilvl 71, trinket (x 0.5625) | 46.1 |
| ilvl 76, two-hander | 87.5 |

This is within 4% of the all-items proposal at every ilvl, so the earlier proposal stands in substance. Against standards 1a, it is +3% at 66, -3% at 71 and -8% at 76.
