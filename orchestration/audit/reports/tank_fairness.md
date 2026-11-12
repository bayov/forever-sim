# tank_fairness pass 9: 2026-11-12, master cb0d928

Due: 2 watched files changed since 33db32b (sim/druid/talents.go, sim/shaman/talents.go), both in cb0d928
("Talent tooltips: bring stale text in line with what the sim already does").

## What changed

- sim/druid/talents.go: a comment on Rend and Tear (where the 3% shows in the tree). No code.
- sim/shaman/talents.go: `ImprovedFireNova` renamed `TrialByFire` (proto/shaman.proto field 9, same number), so
  saved talent strings and requests decode the same. No tank request names the field. No behaviour change.
- Also in range: 934c698a2 (10 new items seen in game, db.bin). No tank preset picks them up (scores identical).

## Re-sim

`tools/audit/tank_fairness/run.sh <dir> 3000 1` (new helper: builds tankbench with the database, runs the 12
presets on broodlord, chromaggus, golemagg, magmadar, onyxia, vaelastrasz, and writes summary.tsv through
`tools/audit/tank_fairness/summary.py`; pass 8 named summary.py but it was never committed).

All 84 rows (72 boss rows plus 12 six-boss means) match `ref/tank_fairness/summary.tsv` from pass 8 exactly
(score, TPS, DTPS, healing needed within 1.5; in fact equal). The baseline stands unchanged.

## Ranking (unchanged from pass 8, seed 1, 3000 iterations, six-boss mean)

| weight | order | scores |
|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 3970 / 3166 / 2663 / 2498 |
| Balanced | Warrior, Paladin, Bear, Warden | 3236 / 2751 / 1717 / 1134 |
| Max survival | Warrior = Paladin, Bear, Warden | 1564 / 1559 / -579 / -2741 |

The explanations in pass 8's report and orchestration/fairness_tanks.md still hold.

## Same terms

Requests unchanged since pass 8 (flask, race, potion re-checked): the Natural flasks on all Max survival presets
(20261111-1), Night Elf / Undead Warden (20261105-2), paladin Major Mana Potion with no potion action (20261105-1),
legacy warrior_protection.json still present (20261104-5). Paladin damage taken (20261104-2) still waits on the user.

## Findings

No new findings. All five open findings unchanged on master; noted under each.
