# Fairness pass 46 (2026-11-18, master 2f7e255af)

Result: quiet (no new finding). Range 5f20dfdf4..2f7e255af. No re-sweep: no report build moved.

## What changed in the watched paths

| Change | Commit | Effect on the comparison |
|---|---|---|
| Beast Mastery weaves for every main hand swing (APL, presets) | 084813024, 5533c0701 | BM 1115.5 (5k baseline) vs 1115.2 (20k shared run): no move |
| Shieldsworn Holy Strike no Windfury Totem (core flag comment, paladin holy_strike.go) | 4e81669a9 | tank only; DPS paladins unchanged (Ret 1024.2 vs 1023.3) |
| Vampiric Embrace heals 10% / 20% | 4e81669a9 | healing only; Shadow 948.6 vs 948.7 |
| Druid Idol of Rejuvenation, Shaman Totem of Tides | 192cb2e06, 6a3cadbfc | healers only; no DPS request uses them |
| Item hotfixes (Darkmoon Cards, Bad Hat, Militant's Band, Onyxia Blood Talisman, ...), item_effects.go | fa03a8689, 52596be22, 3ec610c40, 16b197c5b, 056dccb4e | synthetic gear untouched; no DPS build moved |
| Re-sim of 32 builds; hunter / warrior / Subtlety metas and builds.json refreshed | f42ee71bd, 2c6bd10e5 | published lag of pass 45 closed |
| report_hashes.txt | 5533c0701 | test data |

## Checks

- Shared 20k run of 2f7e255af vs `ref/fairness/sweep_baseline.log` (5k) at each report race: all 43 base builds
  within 1.4 except Enhancement Water Shield (1027.8 -> 1025.7). Paired 20k seed 1 on the 5f20dfdf4 CLI for all 8
  Enhancement / Frostfire / Demonology requests: 0 of 8 moved more than 0.3. The -2.1 is 5k baseline noise. No leak.
- Race: report race still best or within 0.6 (MM Orc 1201.7 vs Night Elf 1202.3; Fire Human 1053.8 vs Gnome 1054.2;
  Combat Daggers Undead 936.0 vs Gnome 936.6). Unchanged from pass 45.
- Gear: `audit1b.py` on 2f7e255af, 19 of 19 "5a PASS | 1b PASS".
- Published vs current: builds.json within 1.5 of the shared run on 68 of 68 (results auditor, update 2026-11-18 under
  results-20261104-2). The pass-45 ranking flips (Survival under Ret, Arms over Affliction in the published order) are
  gone. Meta-only lags remain: Affliction +31.9 to +36.7, Lone Wolf -22.4, rogues -10.5 to -20.1, Fire -4.9
  (results-20261104-1 / -2; confirmed independently, not refiled). They do not reach builds.json.

## Ranking

Unchanged from pass 45 (top Lone Wolf 1291.8, bottom Smite 781.0, 1.65x). Full table: `ref/fairness/table_baseline.md`.

## Baseline

Kept at 5f20dfdf4 (no build moved). `orchestration/fairness_report.md` §1: published column refreshed for the 13 hunter /
warrior / Subtlety rows and a pass-46 note added; the DPS columns stand.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, FTT + WF imbue stacking, boss partial resist log, Warrior Rage k and
demon/pet scaling ranges (see passes/fairness.md).
