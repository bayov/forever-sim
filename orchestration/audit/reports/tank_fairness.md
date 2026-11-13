# tank_fairness pass 14: 2026-11-13, master 3a86d4a6b

Due: 1 watched file changed since cad23ae1a: `sim/shaman/flametongue_weapon.go` (4c4790e3e, Flametongue Weapon rank 6
cut of 5.5 per hit from the Nov 12-13 logs). No tank request, encounter or tankbench file changed. Quiet: no new finding.

## Method

All three Warden presets (`tools/tankbench/req/shaman_warden_*.json`) imbue Rockbiter Weapon and none swaps imbues;
`FlametongueWeaponCut` is read only by Flametongue itself (`flametongueRank` also feeds `imbue_swap.go`, which the
Warden presets don't use). To confirm, `tools/audit/tank_fairness/bench.sh` on the three Warden presets × six scoring
bosses, 3,000 iterations, seed 1, tankbench built at 3a86d4a6b with `--tags=with_db`: byte-identical to the
`ref/tank_fairness/summary.tsv` baseline rows. The warrior, paladin and bear presets have no shaman code path and were
not re-run. Baseline unchanged.

## Pass 13's findings (unchanged, carried below)

`orchestration/standards.md`. They are the Nov 11 paladin hotfixes (33b13e310: Judgement rolls as a ranged attack,
Unsealed Wrath stores no echo from Windfury Totem swings, seal changes) and a1a1a81bb (no extra attack rolls Windfury
Totem; standards 17 now says so). No tank request, encounter or tankbench file changed. Quiet: no new finding.

## Method

`tools/audit/tank_fairness/bench.sh` on the 12 presets × six scoring bosses, 3,000 iterations, seeds 1 and 2, tankbench
built at 97da1cca5 with `--tags=with_db`. Paired variant (`variants.py`, `paired.py`, seed 1): paladin Greater
Stoneshield plus the potion action. New baseline: `ref/tank_fairness/summary.tsv` (seed 1, 3,000).

## Same terms

The requests are byte-identical to pass 12, so the terms are as pass 12 found them: the two open unequal terms (paladin
potion, Warden races) and the settled enchant and consumable rulings. The standards change (17) applies to every tank
alike.

## Ranking (six-boss mean, own weight, 3,000 iterations; seed 1 / seed 2)

| weight | order | scores seed 1 | seed 2 | pass 12 (seed 1) |
|---|---|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4223 / 3246 / 2789 / 2557 | 4220 / 3238 / 2769 / 2550 | 4223 / 3263 / 2789 / 2557 |
| Balanced | Warrior, Paladin, Bear, Warden | 3428 / 2804 / 1865 / 1219 | 3419 / 2801 / 1850 / 1211 | 3428 / 2832 / 1865 / 1219 |
| Max survival | **Paladin = Warrior**, Bear, Warden | 1776 / 1776 / 230 / -2312 | 1786 / 1774 / 199 / -2317 | 1803 / 1776 / 230 / -2312 |

Only the paladin moved: -17.5 / -27.3 / -27.6 (seed 1), -21 / -26 / -30 (seed 2). Warrior, bear and Warden are
unchanged to the point (none of their presets has an extra-attack source that rolled Windfury Totem in a scored way).
The paladin's loss is the hotfixes, which are Forever's own (Blue post t/2403040; whether they are modelled right is
tank_mechanics' and results' to check). Max survival is now a tie: seed 1 paladin 1775.8 ± 13.6 against the warrior's
1775.5 ± 12.8, seed 2 1785.7 ± 14.0 against 1773.8 ± 10.9. The open potion finding (-20261105-1) decides that place.

TPS / DTPS / healing needed at Balanced: warrior 2057 / 374 / 344, paladin 1519 / 309 / 245, bear 1482 / 587 / 464,
Warden 1351 / 660 / 660. The spreads are the ones explained in passes 4 to 12 (warrior Revenge and the Windfury rule;
paladin block value multipliers, finding -20261104-2; Warden a class gap, standards 40).

The logs disagree as before: no paladin main tank on any boss (warriors on all six, druids on four). The sim puts the
paladin second at Max threat and Balanced and level first at Max survival.

## What each preset rewards (seed 1; seed 2 has the same winners)

| preset | at Max threat | at Balanced | at Max survival |
|---|---|---|---|
| warrior MT / Bal / MS | **4223** / 4081 / 3382 | 3298 / **3428** / 2980 | 524 / 1470 / **1776** |
| paladin MT / Bal / MS | **3246** / 3214 / 3092 | 2788 / **2804** / 2763 | 1416 / 1576 / **1776** |
| bear MT / Bal / MS | **2788** / 2718 / 2207 | 1633 / **1864** / 1712 | -1833 / -695 / **230** |
| Warden MT / Bal / MS | **2556** / 2434 / 2347 | 990 / **1219** / 1182 | -3707 / -2426 / **-2312** |

Each preset wins its own weight, and the three builds are distinct for every spec. The paladin's Max threat leads its
Balanced at Max threat by 32, and its Balanced leads Max threat at Balanced by 16 (seed 2: 31, 17): close but separate.

## Open findings re-checked

| finding | status | this pass |
|---|---|---|
| -20261105-1 paladin never drinks its potion (high) | not fixed | Stoneshield + action, paired seed 1: +59.3 / +169.3 / +374.5; DTPS -29 to -32. Now decides Max survival (tie without it) |
| -20261105-2 Warden plays illegal races (high) | not fixed | requests unchanged; pass 12's paired cost stands (Orc -50, Troll -77, Tauren -124) |
| -20261104-2 paladin takes about half the warrior's damage (medium) | waits on user calls (a), (b), (c) | Balanced healing needed 245 against 344 / 464 / 660 |
| -20261104-5 legacy warrior_protection.json (low) | not fixed | file still present |

## Left for later

Nothing beyond the open findings.
