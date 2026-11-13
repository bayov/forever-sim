# tank_fairness pass 12: 2026-11-13, master 71562a8b0

Due: 38 watched files changed since 4fb90c298. For the tanks, that is a6a8305da: the best legal enchant per slot on all
12 presets at each weight, and the bear's Max survival race Night Elf to Tauren (closing -20261112-2, -3). Also
changes to the Rake/feral tests and a Windfury test (no tank request reads them). Quiet: no new finding.

## Method

`tools/audit/tank_fairness/bench.sh` on the 12 presets × six scoring bosses, 3,000 iterations, seeds 1 and 2
(tankbench built with `--tags=with_db`; without it the synthetic items panic, now noted in bench.sh). Paired variants
(`variants.py`, `paired.py`, seed 1): paladin Greater Stoneshield plus the potion action, and the Warden's legal races.
New baseline: `ref/tank_fairness/summary.tsv` (seed 1, 3,000).

## Same terms

Only the enchants and the bear's Max survival race changed in the requests. Consumables, buffs, debuffs, reaction time,
professions and the other races are unchanged from pass 11. The enchant table now differs by preset. That is by design:
each preset took the winner of a screen of every legal Phase 1 enchant at its own weight (finding -20261112-2, settled).
The bear's Tauren is legal and won by its objective. The two old unequal terms are still open: the paladin's potion
and the Warden's races.

## Ranking (six-boss mean, own weight, 3,000 iterations; seed 1 / seed 2)

| weight | order | scores seed 1 | seed 2 | pass 11 |
|---|---|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4223 / 3263 / 2789 / 2557 | 4220 / 3259 / 2769 / 2550 | 4202 / 3187 / 2789 / 2536 |
| Balanced | Warrior, Paladin, Bear, Warden | 3428 / 2832 / 1865 / 1219 | 3419 / 2827 / 1850 / 1211 | 3426 / 2774 / 1865 / 1189 |
| Max survival | **Paladin**, Warrior, Bear, Warden | 1803 / 1776 / 230 / -2312 | 1816 / 1774 / 199 / -2317 | 1711 / 1762 / 191 / -2385 |

Change since pass 11 (seed 1): paladin +76 / +58 / +92, Warden +21 / +30 / +74, warrior +21 / +3 / +14, bear 0 / 0 /
+39. These match the worker's 10,000-iteration figures in a6a8305da. The paladin retakes Max survival by 27 (seed 1) / 42
(seed 2).

TPS / DTPS / healing needed at Balanced: warrior 2057 / 374 / 344, paladin 1527 / 309 / 239, bear 1482 / 587 / 464,
Warden 1351 / 660 / 660. The spreads are the ones explained in passes 4 to 11:
- the warrior's threat comes from Revenge and the Windfury rule;
- the paladin's low damage comes from its block value multipliers (finding -20261104-2, waiting on the user);
- the Warden is a real class gap (standards 40).

The logs disagree as before: no paladin main tank on any boss (warriors on all six, druids on four). The sim puts the
paladin second at Max threat and Balanced and first at Max survival.

## What each preset rewards (seed 1; seed 2 has the same winners)

| preset | at Max threat | at Balanced | at Max survival |
|---|---|---|---|
| warrior MT / Bal / MS | **4223** / 4081 / 3382 | 3299 / **3428** / 2980 | 524 / 1470 / **1776** |
| paladin MT / Bal / MS | **3263** / 3233 / 3110 | 2810 / **2832** / 2783 | 1449 / 1628 / **1803** |
| bear MT / Bal / MS | **2789** / 2718 / 2207 | 1633 / **1865** / 1713 | -1833 / -695 / **230** |
| Warden MT / Bal / MS | **2557** / 2434 / 2347 | 990 / **1219** / 1182 | -3707 / -2426 / **-2312** |

Each preset wins its own weight, and the three builds are distinct for every spec. The paladin's Max threat and
Balanced are still close at Max threat (30), since they shared one build until the enchant split.

## Open findings re-checked

| finding | status | this pass |
|---|---|---|
| -20261105-1 paladin never drinks its potion (high) | not fixed | Stoneshield + action, paired seed 1: +60.5 / +168.3 / +367.8; DTPS -29 to -32. No longer flips Max survival (the paladin already leads) |
| -20261105-2 Warden plays illegal races (high) | not fixed | best legal, paired seed 1: Orc -49.8 (MT), Troll -76.8 (Bal), Tauren -123.5 (MS) |
| -20261104-2 paladin takes about half the warrior's damage (medium) | waits on user calls (a), (b), (c) | Balanced healing needed 239 against 344 / 464 / 660 |
| -20261104-5 legacy warrior_protection.json (low) | not fixed | file still present |

## Left for later

Nothing beyond the open findings.
