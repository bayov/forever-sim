# How Forever (and SoD) price item stats

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
