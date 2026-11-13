# tank_rotations pass 11

Master 71562a8b0 (2026-11-13), against pass 10's 8306f1d52 baseline. 1,000 iterations, default seed; A/Bs paired
(abtest) at the preset's own weight on the six scoring bosses, Ragnaros reported apart.

## What changed

Since 8306f1d52 (tank-relevant, via the tankheal merges):

- 67ad7c6a4: bear Max survival Frenzied Regeneration at 70 / Maul 40, Warden Balanced Healing Stream, warrior Max
  survival Shield Wall 13 s (closes 20261112-1..4 and 20261110-1).
- 2a0208136 / c1502631c: Windfury Totem rolls on main-hand abilities by their client flag (standards 17).
- ca7c42400: re-tune on that rule: bear Max threat Barkskin from 20 s; bear Max survival FR 80, Swipe filler at 80,
  Feral Instinct 3 talents. Warrior and bear Balanced knobs held.
- 09a2f47ef: raid-only consumables out (warrior Dense stone, Warden Greater Mageblood Elixir).
- 814decce0, a6a8305da: enchants per slot at each weight (all four tanks), new paladin / warrior / bear gear sets,
  bear Max survival Tauren.
- f1c07e13a: Judgement is a ranged attack. Healer-side commits (65b5d49cf, 8fd5ddcb9, ...) don't touch tank rows.
- No change to tools/encounters, tools/tankbench/main.go or tools/tankscore.

## Baseline (scoring mean, own weight; pass 10 ref -> now; full grid in ref/tank_bench/summary.tsv, rewritten)

| request | score | SafeDPS | DTPS | healing needed |
|---|---|---|---|---|
| bear Balanced | 1856.2, identical | 1569.3 | 586.0 | 462.0 |
| bear Max survival | -296.5 -> 261.8 | 1394.2 -> 1226.3 | 543.7 -> 552.7 | 375.7 -> 277.5 |
| bear Max threat | 2771.5 -> 2782.2 | 1672.8 -> 1677.8 | 611.7 -> 611.8 | 633.3 -> 633.5 |
| paladin Max threat | 3171.8 -> 3266.7 | 1693.5 -> 1744.8 | 306.2 -> 308.7 | 256.5 -> 264.8 |
| paladin Balanced | 2753.2 -> 2837.5 | 1688.5 -> 1716.5 | 314.3 -> 308.0 | 245.3 -> 238.0 |
| paladin Max survival | 1554.0 -> 1810.7 | 1504.2 -> 1637.2 | 298.3 -> 287.8 | 197.3 -> 198.3 |
| Warden Max threat | 2495.7 -> 2562.0 | 1626.2 -> 1671.8 | 718.0 -> 680.7 | 823.3 -> 842.8 |
| Warden Balanced | 1124.7 -> 1224.2 | 1556.7 -> 1523.3 | 675.0 -> 662.0 | 716.8 -> 658.7 |
| Warden Max survival | -2736.0 -> -2337.0 | 1502.8 -> 1464.0 | 684.7 -> 671.3 | 691.2 -> 638.5 |
| warrior_protection (legacy) | -14.2, identical | 1048.2 | 637.3 | 757.0 |
| warrior Max threat | 4216.3 -> 4235.0 | 2327.2 -> 2346.0 | 415.0 -> 418.3 | 445.2 -> 463.5 |
| warrior Balanced | 3434.3 -> 3420.8 | 2191.2 -> 2201.2 | 377.3 -> 375.5 | 333.3 -> 344.8 |
| warrior Max survival | 1771.3 -> 1776.0 | 1897.5 -> 1796.2 | 334.8 -> 321.8 | 245.7 -> 217.3 |

Every move lines up with an intended request change (enchants, gear sets, Tauren, FR/Swipe/talents, HST, Shield
Wall 13 s, consumables out, ranged Judgement, the Windfury rule). The two requests no commit touched (bear Balanced,
legacy warrior_protection) are bit-identical, so the engine moved nothing else. Warrior Balanced -13.5 is inside its
error (per-boss score_err 15-65) and follows its enchant / stone change. No new finding.

## Changes tried (bear Max survival, w = 3, paired 1,000, 7 bosses)

The shipped FR 80 / Swipe 80 is the newest knob set (Tauren, new talents), so it was re-checked:

| change | scoring mean diff | paired err | verdict |
|---|---|---|---|
| FR at 70 Rage (from 80) | -119.4 | 37.4 | loses |
| FR at 90 Rage | -461.6 | 38.4 | loses |
| Swipe filler at 70 Rage (from 80) | -278.7 | 36.5 | loses |

FR 80 / Swipe 80 holds on the Tauren set. Other presets' knobs and talents were searched on the new rule by
ca7c42400 / a6a8305da (10k paired, two seeds) and are not re-run here (the host was loaded by other sims).

## Open findings

New: none. Open, unchanged (user calls): 20261104-11 (warrior Max survival armor trinket held 35 s, Broodlord),
20261104-3 (paladin Bulwark 13 s, Broodlord).

Raw output: `/tmp/audit-work/tank_rotations/` (base/, ab_fr70.txt, ab_fr90.txt, ab_sw70.txt).

## Note for the next pass

Building needs Go protos: `protoc -I=./proto --go_out=./sim/core ./proto/*.proto` (PATH must include /root/go/bin).
`$AUDIT_WORK` doesn't persist between Bash calls; use the absolute /tmp/audit-work path for build outputs.
