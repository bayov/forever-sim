# tank_rotations pass 8

Master 707ae92cc (2026-11-11), against pass 7's ca5e35c3f / state 27d0d59d6. 1,000 iterations, default seed.

## What changed

No tank APL, request, rotopt template, encounter or tankbench/tankscore change. Engine changes that reach the tanks:
- 032692409 (client 1.60.1.70862): Natural flasks halved, Precision 60 Sta + 5% expertise -> 30 + 2%. Every request
  on FlaskOfNaturalPrecision moved (all but paladin Balanced and Max threat, which use Supreme Power and are unchanged).
- cb651085c (Blizzard Nov 10 shaman tuning): Stormstrike 150% / 6 s -> 100% / 8 s, Mental Dexterity 1/3 per point,
  Maelstrom Weapon PPM, imbue and weapon totem stack. Moves the Warden.
- Windfury Totem's white-swing-only mask (finding 20261110-1) is unchanged.

Attribution: a worktree build of master with sim/core/consumes.go reverted to 032692409^ reproduces pass 7's
warrior, bear and paladin rows exactly (e.g. warrior Balanced Onyxia 3737 / SafeDPS 2299; bear 1812 / 1700;
paladin Max survival Magmadar 3021 / 1510), so the flask nerf is their whole move. The Warden with the old flask is
still about 10% down (Onyxia SafeDPS 1899 -> 1716, Magmadar 1734 -> 1560): the shaman tuning.

## Baseline (scoring mean, without Ragnaros)

| request | weight | score (pass 7 -> 8) | SafeDPS | TPS | DTPS | healing needed |
|---|---|---|---|---|---|---|
| druid_bear_balanced | 1 | 2006.5 -> 1717.0 | 1649.5 -> 1519.0 | 1428.2 | 586.8 | 471.7 |
| druid_bear_max_survival | 3 | -297.7 -> -578.3 | 1460.3 -> 1343.2 | 1262.3 | 545.2 | 393.3 |
| druid_bear_max_threat | 0.333 | 2953.7 -> 2654.8 | 1758.0 -> 1616.3 | 1505.0 | 612.2 | 638.7 |
| paladin_protection_balanced | 1 | 2753.2 -> 2753.2 | 1688.5 -> 1688.5 | 1501.3 | 314.3 | 245.3 |
| paladin_protection_max_survival | 3 | 1763.3 -> 1554.0 | 1571.7 -> 1504.2 | 1337.5 | 298.3 | 197.3 |
| paladin_protection_max_threat | 0.333 | 3171.8 -> 3171.8 | 1693.5 -> 1693.5 | 1506.5 | 306.2 | 256.5 |
| shaman_warden_balanced | 1 | 1681.7 -> 1124.7 | 1805.2 -> 1556.7 | 1379.0 | 675.0 | 716.8 |
| shaman_warden_max_survival | 3 | -2030.7 -> -2736.0 | 1741.2 -> 1502.8 | 1348.8 | 684.7 | 691.2 |
| shaman_warden_max_threat | 0.333 | 3054.8 -> 2495.7 | 1893.5 -> 1626.2 | 1437.7 | 718.0 | 823.3 |
| warrior_protection | 1* | -46.8 -> -46.8 | 1031.5 -> 1031.5 | 1263.3 | 636.5 | 756.3 |
| warrior_protection_balanced | 1 | 3637.2 -> 3226.5 | 2271.7 -> 2100.5 | 1952.3 | 379.3 | 342.8 |
| warrior_protection_max_survival | 3 | 1965.5 -> 1561.8 | 1975.0 -> 1822.3 | 1778.2 | 336.7 | 251.8 |
| warrior_protection_max_threat | 0.333 | 4375.2 -> 3957.7 | 2396.5 -> 2196.0 | 2028.5 | 415.2 | 443.0 |

Full rows: ref/tank_bench/summary.tsv (rewritten this pass).

## Changes tried (paired, B minus preset, scoring mean at the preset's weight)

| change | iters | diff |
|---|---|---|
| Warden Bal: Stormstrike above Earth Shock | 1k | -39.9 ± 16.0 |
| Warden Bal: no Stormstrike | 1k | -527.1 ± 16.5 |
| Warden Bal: fire totem + Fire Nova window 60 s -> 90 s | 1k | -1.8 ± 5.5 |
| Warden Bal: window -> whole fight | 1k | -12.3 ± 6.3 |
| Warden Bal: window -> 40 s | 1k | -8.5 ± 12.9 |
| Warden Bal: Fire Nova mana floor 40% -> 20% | 1k | -4.4 ± 5.6 |
| Warden Bal: Fire Nova mana floor 40% -> 60% | 1k | -13.5 ± 8.5 |
| Warrior Bal: Titans for Natural Precision | 1k | -86.5 ± 10.9 |
| Bear Bal: Titans | 1k | -66.8 ± 13.0 |
| Warden Bal: Titans | 1k | +12.1 ± 14.3 |
| Paladin Max surv: Titans | 3k | **+133.2 ± 15.5** |
| Paladin Max surv: Supreme Power | 1k | +139.0 ± 27.4 |
| Warden Max surv: Titans | 3k | **+243.1 ± 23.6** |
| Bear Max surv: Titans | 3k | **+108.5 ± 12.9** |
| Warrior Max surv: Titans | 3k | +39.5 ± 16.2 |

The Warden's rotation lines hold after the tuning. Warrior, paladin and bear rotations are unchanged since pass 7
and their engine inputs moved only by the flask's expertise and Stamina, so their lines weren't re-searched.

## Findings

- Flask: the Titans-over-Natural-Precision result at Max survival is tank_fairness-20261111-1, filed the same day;
  the table above confirms it independently (default seed), so it isn't filed again here.
- New: 20261111-1 (medium) Warden talents and timers not re-optimized after the Nov 10 shaman tuning.
- Still open: 20261110-1 (Windfury/Maul, user ruling), 20261104-11, 20261104-3 (Broodlord timer calls).

## Repro

`go build -tags=with_db -o $AUDIT_WORK/abtest ./tools/audit/tank_rotations/abtest`; variants with
`tools/audit/tank_rotations/mk_warden.py` (the line's fields are under `x["action"]`); baseline with tankbench
then `tools/audit/tank_rotations/bench_summary.py`.
