# tank_fairness pass 16: 2026-11-15, master 25ce81de7

Due: 16 watched files changed since 70ea588bb: PR #130 (Nov 13 hotfixes: Windfury Totem rolls on every landed main
hand melee hit, standards 17; Judgement of Command melee again; paladin seal and talent changes), the warrior's Balanced
and Max survival re-tune (f2a5dfe25: Revenge before Shield Slam, an 8 s refresh window, a 40 rage gate), rotopt's
Windfury gate knobs, and a Wild Growth cooldown fix in `sim/druid/healing.go`. Quiet: no new finding.

## Method

`tools/audit/tank_fairness/bench.sh` on the 12 presets x six scoring bosses, 3,000 iterations, seed 1, tankbench built
at 25ce81de7 with `--tags=with_db`. Attribution: tankbench built at b02b98c72 (before PR #130) and d4325b775 (its
merge), the bear's three presets, the Warden's Max threat and Balanced and the paladin's Balanced, same HEAD requests.
New baseline `ref/tank_fairness/summary.tsv` (seed 1, 3,000, all 12 presets).

## Same terms

Only the two warrior requests changed, and only in their APL (priority order and two thresholds); level, set,
consumables, buffs, race, professions and reaction time are byte-identical to pass 13. The open unequal terms stand
(paladin potion -20261105-1, Warden races -20261105-2). The standards change (17) applies to every tank alike.

## Ranking (six-boss mean, own weight, seed 1)

| weight | order | pass 16 | pass 15 |
|---|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4223 / 3275 / 3123 / 2644 | 4223 / 3246 / 2794 / 2557 |
| Balanced | Warrior, Paladin, Bear, Warden | 3452 / 2843 / 2131 / 1280 | 3428 / 2804 / 1870 / 1219 |
| Max survival | **Paladin = Warrior**, Bear, Warden | 1798 / 1794 / 364 / -2242 | 1776 / 1776 / 231 / -2312 |

Errors: warrior ± 5.4 / 8.4 / 12.6, paladin ± 3.4 / 5.3 / 13.1, bear ± 5.3 / 7.9 / 20.4, Warden ± 4.3 / 7.4 / 16.8.

### Where the moves come from (seed 1, before -> after PR #130)

| preset | b02b98c72 | d4325b775 | delta | TPS delta |
|---|---|---|---|---|
| bear Max threat | 2788.5 | 3116.8 | +328 (+11.8%) | +163 |
| bear Balanced | 1864.5 | 2124.5 | +260 | +144 |
| bear Max survival | 229.8 | 358.5 | +129 | +94 |
| Warden Max threat | 2556.5 | 2644.2 | +88 | +42 |
| Warden Balanced | 1218.8 | 1279.8 | +61 | +36 |
| paladin Balanced | 2804.5 | 2842.5 | +38 | +14 |

DTPS and healing needed barely move (bear +9 healing needed at Balanced; the rest within
noise): the gain is threat from Windfury Totem now rolling on the bear's main hand specials, the Warden's
Stormstrike and the paladin's Holy Strike / Judgement of Command. The later commits add only the bear +6 (rotopt, no
request change; within error) and the warrior's re-tune, +24 at Balanced and +18 at Max survival (its own APL work).
The warrior's Max threat is unchanged to the point: Sunder, Revenge and Heroic Strike already rolled, and Shield Slam is
a shield hit. Whether every bear ability that now rolls is a main hand melee hit under Forever is tank_mechanics' to
check; under standards 17 as written the gain is fair.

Spread: the bear closed from 14% to 5% behind the paladin at Max threat and from 33% to 25% at Balanced. Warden stays
37% behind the warrior at Max threat (class gap, standards 40). TPS / healing needed at Balanced: warrior 2063 / 342,
paladin 1533 / 241, bear 1628 / 473, Warden 1387 / 666. The bear now out-threats the paladin at Balanced (TPS) while
scoring below it on healing needed.

The logs disagree as before: no paladin main tank on any boss (warriors on all six, druids on four). The bear's rise
moves the sim toward the logs; the paladin still sits second at Max threat and Balanced and level first at Max survival.

## What each preset rewards (seed 1, six-boss mean)

| preset | at Max threat | at Balanced | at Max survival |
|---|---|---|---|
| warrior MT / Bal / MS | **4223** / 4100 / 3404 | 3299 / **3452** / 3001 | 524 / 1508 / **1794** |
| paladin MT / Bal / MS | **3275** / 3246 / 3123 | 2818 / **2843** / 2791 | 1448 / 1632 / **1798** |
| bear MT / Bal / MS | **3123** / 3002 / 2399 | 1961 / **2131** / 1890 | -1525 / -484 / **364** |
| Warden MT / Bal / MS | **2644** / 2506 / 2414 | 1076 / **1280** / 1250 | -3629 / -2397 / **-2242** |

Every preset wins its own weight and the three builds stay distinct for every spec. The paladin's spread is still the
narrowest (Max threat 3275 against Balanced's 3246, Balanced 2843 against 2818).

## Findings

No new finding. Open: -20261105-2 (Warden races), -20261105-1 (paladin potion; still decides Max survival),
-20261104-2 (paladin damage taken), -20261104-5 (legacy warrior_protection.json, still present). None fixed on master.
