# tank_fairness pass 6: 2026-11-08, master 9ba00c2

Due: 12 watched files changed since 73d5bfa: `orchestration/standards.md` (25 and 40: no raid shaman has
Flametongue Totem, since Windfury and Flametongue Totem replace each other) and eleven `sim/shaman/*.go` files
(totems, imbue swap, shocks, Lightning Bolt, Stormstrike). The same range also changed `sim/core` (spell damage rolled
in whole points, 2a60326; Forever's downranking penalty on heals only, 0193721), which reaches every tank.

## Same terms

- Totems: all twelve tank requests carry `totemWeaponBuff: TotemWeaponBuffWindfury`, the melee default standards 25
  keeps. The Flametongue change touches only the Enhancement presets; no tank request used Flametongue Totem.
- Nothing else under the tank requests, encounters, tankscore or the warrior, paladin and druid code changed. Races,
  consumables, buffs and the synthetic sets are as in pass 5; its two open asymmetries stand (below).

## The numbers: master 9ba00c2 against 73d5bfa, paired

tankbench `-iters 3000 -seed 1`, every `*_{max_threat,balanced,max_survival}` request on the six scoring bosses, built
from both commits (tools/audit/build_cli.sh trees), summarized with `tools/audit/tank_fairness/summarize.py NEW OLD`.

| request | threat | balanced | survival | TPS | DTPS | heal needed | largest delta |
|---|---|---|---|---|---|---|---|
| warrior max threat | 4650 | 3813 | 1303 | 2316 | 414 | 427 | 0 |
| warrior balanced | 4466 | 3867 | 2072 | 2200 | 376 | 317 | ±0.2 |
| warrior max survival | 3903 | 3465 | 2150 | 1998 | 334 | 239 | -0.2 |
| paladin max threat | 3248 | 2819 | 1531 | 1542 | 306 | 256 | 0 |
| paladin balanced | 3240 | 2826 | 1582 | 1534 | 314 | 245 | 0 |
| paladin max survival | 3077 | 2789 | 1922 | 1430 | 296 | 179 | 0 |
| bear max threat | 3491 | 2365 | -1012 | 1874 | 612 | 622 | 0 |
| bear balanced | 3348 | 2500 | -44 | 1760 | 584 | 457 | -0.2 TPS |
| bear max survival | 2964 | 2281 | 232 | 1529 | 545 | 374 | 0 |
| Warden max threat | 3169 | 1704 | -2690 | 1724 | 719 | 797 | 0 |
| Warden balanced | 3077 | 1794 | -2055 | 1648 | 676 | 692 | 0 |
| Warden max survival | 2956 | 1718 | -1994 | 1601 | 685 | 672 | 0 |

Every six-boss mean moves by at most 0.2 (the warrior's whole-point spell rolls); the Warden is bit-identical, so the
shaman changes do not reach its rotation (no Flametongue, no imbue swap, no downranked heals in its presets). These
equal pass 4's numbers at a079415 (`ref/tank_fairness/summary.tsv`), which were not refreshed.

Ranking at each preset's own weight, unchanged: Max threat warrior 4650 / bear 3491 / paladin 3248 / Warden 3169;
Balanced warrior 3867 / paladin 2826 / bear 2500 / Warden 1794; Max survival warrior 2150 / paladin 1922 / bear 232 /
Warden -1994. Each preset still leads its own weight within its spec, and the three are distinct builds.

## Open findings, unchanged on this master

- -20261105-2 (high): Warden presets on Night Elf / Undead, races a Forever shaman can't be.
- -20261105-1 (high): paladin has no potion action; Greater Stoneshield worth +58 / +172 / +309.
- -20261104-2 (medium): paladin damage taken, waiting on user calls (a), (b), (c).
- -20261104-5 (low): legacy warrior_protection.json still read by tests and tools.

## Not done

No Anniversary-log comparison (pass 4's stands; no encounter changed). No new findings.
