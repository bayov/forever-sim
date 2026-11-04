# tank_gear pass 1: 2026-11-04, master f62bfc74b

This is the first pass since the four tank specs landed with synthetic gear, realigned to client 70753 (standards 4: tank hit sized to the cap counting talents). It was forced. Each spec's script recomputes every number below from the Go items and the gear JSON. Run it from the repo root: `python3 tools/audit/tank_gear/<warrior|paladin|druid|shaman>.py`.

Prices: standards 1b plus 36. Defense is 1.5, block value 0.4, block 5 per 1%, bonus armor 0.1, hit and crit 1 per rating. Targets are slot mod x L(ilvl) x f, and the norm is the 1.7-norm shape target.

## Pieces (linear % / norm % of target)

| slot | Prot warrior | Prot paladin | Feral bear | Warden |
|---|---|---|---|---|
| Head | 99.4 / 100.4 | 99.6 / 102.0 | 100.1 / 99.3 | 99.6 / 100.8 |
| Neck | 100.1 / 96.7 | 100.5 / 100.4 | 100.1 / 95.7 | 100.5 / 100.7 |
| Shoulder | 100.3 / 102.0 | 99.6 / 101.9 | 100.3 / 100.9 | 100.4 / 102.6 |
| Back | 100.1 / 96.7 | 99.5 / 98.6 | 100.1 / 95.7 | 100.5 / 98.0 |
| Chest | 99.4 / 100.4 | 99.6 / 102.0 | **98.7** / 98.4 | 99.6 / 100.8 |
| Wrist | 100.1 / 97.6 | 99.5 / 98.1 | 100.1 / 95.7 | 100.5 / 97.9 |
| Hands | 100.3 / 101.4 | 100.6 / 95.2 | 100.3 / 100.7 | 99.9 / 103.3 |
| Waist | 99.5 / 101.5 | 99.1 / 96.1 | 100.2 / 100.6 | 100.5 / 102.9 |
| Legs | 99.4 / 103.2 | 99.6 / 101.4 | 100.1 / 99.3 | 99.6 / 100.8 |
| Feet | 100.3 / 102.8 | 99.1 / 100.0 | 100.3 / 100.7 | 100.4 / 102.6 |
| Finger | 100.1 / 95.4 | 99.5 / 102.3 | 100.1 / 95.7 | 100.5 / 100.5 |
| Finger II | 100.1 / 105.3 | 100.5 / 99.0 | 100.1 / 95.4 | 100.1 / 101.0 |
| Weapon | 100.2 | 100.4 | 100.3 (2H) | 100.4 |
| Shield | 99.9 | 99.9 | none | 100.1 |
| Ranged | 99.2 | n/a | n/a | n/a |
| Tank trinket | 100.2 / 104 | 100.3 / 104.2 | **101.4** / 102.6 (rounding) | 100.2 / 97 |
| Set total | 760.0 / 761.0 | 841.8 / 844.1 | 729.5 / 729.3 | 845.1 / 844.1 |

The armor types are right everywhere: plate, plate, leather and mail, with cloth backs. Item levels and slots match standards 2 and the comments. Shields are where the class tanks with one, and the bear has none.

## Set shares against the Tanks rows (hit-adjusted, trinkets excluded)

| spec | out of tolerance | near the limit | guardrail (top non-Stamina share) |
|---|---|---|---|
| Prot warrior | none | hit +1.1, crit -1.3 | Defense 30.8% |
| Prot paladin | **Stamina 30.2 against 32.4 ±1.5** | none | Defense 27.8% |
| Feral bear | **Stamina 35.1 against 33.4 ±1.5 (+0.2 over)** | Agility +2.8 | Agility 34.6% |
| Warden | none | none (all within 0.1) | Intellect 20.4% |

## Hit (standards 4)

- **Warrior:** 80 rating, 8.0%. No talent gives hit. At the cap.
- **Paladin:** 50 rating plus Precision 3/3, 8.0%. At the cap.
- **Bear:** 30 rating, 3.0%, with no talent hit. It reaches 8% only with the Flask of Natural Accuracy. The author sized it to the flask on a sim (ahead of the Titans flask) and says so, which standards 4's "mix that sims best" allows. This is open for a re-check, not a finding.
- **Warden:** 55 rating plus Tidal Focus 5%, 10.5%, which is 2.5% over the melee cap. docs/tank_shaman.md:50-53 and :137 show the excess beats crit at every weight, and it also counts for Earth Shock. Accepted under standards 4.

## Trinkets

- All four stat trinkets are Defense with block value or dodge, carry no Stamina and are within 1b and 36.
- The on-use trinkets average 37:
  - **Paladin:** Strength on-use. The armor alternative was simmed lower.
  - **Bear:** Strength on-use, deliberate per presets.ts:49-52.
  - **Warden:** AP on-use, and the armor on-use in the Max survival set.
  - **Warrior:** AP on-use with no recorded A/B (finding 2).

## Enchants

- All are non-raid picks. Gloves Threat, Cloak Dodge and Arcanum of Protection are AQ or later, and correctly excluded.
- Open points (finding 3):
  - Shield Greater Stamina is +7 where the vendor formula gives +9.
  - Bracer Deflection is understated in enchant_overrides.go, and Superior Deflection has not been evaluated.
- Warden boots: Mystic Rugged Armor Kit is unevaluated (finding 6).
- Not yet compared, judged unlikely to change: Crusader against Strength +15 on paladin and Warden.

## Tanks rows (gear_targets.md)

All four follow from orchestration/tools/gear_targets_tanks.json the way gear_targets.md, Tanks, says:
- Stamina is held, resistances (and the bear's Spirit) are dropped, and the rest is re-scaled.
- The paladin and Warden add Spirit 10%.

The budgets (761.0, 844.1, 729.3, 844.1) match standards 36. This checks only the arithmetic from the committed SoD shares. No wowsims/sod checkout was available to re-derive them, and no pass fetches Wowhead.

The paladin's row Stamina of 32.4 is measured with resistances still in the denominator, which would be about 36.6% without them. It is a documented choice and is not raised.

## New findings

| id | severity | title |
|---|---|---|
| tank_gear-20261104-1 | medium | Paladin Stamina share 30.2 outside 32.4 ±1.5 |
| tank_gear-20261104-2 | medium | Warrior wears AP on-use trinket with no A/B against armor trinket |
| tank_gear-20261104-3 | low | Stale Classic enchant values (shield +7 not +9, Deflection) |
| tank_gear-20261104-4 | low | Bear chest -1.3% linear; Stamina share +0.2 over tolerance |
| tank_gear-20261104-5 | low | Stale comments (bear trinket, warrior header numbers) |
| tank_gear-20261104-6 | low | Warden boots: Mystic Rugged Armor Kit unevaluated |
