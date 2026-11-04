# tank_fairness pass 2: 2026-11-04, master b38ff16

Forced by hand after round 2 (four class workers fixed pass 1's findings and re-optimized every preset).
Raw table: `ref/tank_fairness/summary.tsv` (84 rows: 12 preset requests × 6 scoring boss files, scores at all three weights).

## How

- `go build -tags=with_db -o tankbench ./tools/tankbench` (a fresh checkout needs
  `protoc -I=./proto --go_out=./sim/core ./proto/*.proto` first), then per boss
  `tankbench -iters 3000 -encounter tools/encounters/<boss>.json tools/tankbench/req/*_{max_threat,balanced,max_survival}.json`
  (seed 1; about 1 s a boss). Six-boss mean error = sqrt(Σerr²)/6; paired error taken as the quadrature of two means' (conservative).
- Same terms: all 12 requests flattened with jq, gear/talents/rotation set aside.
- Equal care: the round-2 sections of docs/tank_*.md and the UI presets (ui/{tank_warrior,protection_paladin,feral_tank_druid,warden_shaman}/presets.ts) against the requests.
- DPS Report Tanking mode: ui/dps_report (Tanking code and data), tools/dps_report/tankheal.py.
- Variants: Max threat with Flask of Natural Accuracy (warrior, paladin, Warden); paladin with improved Blessing of Wisdom.

## Own-weight six-boss means (3000 iterations) against round 2's claims

| preset | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| Max threat (w 1/3) | 3696.8 ± 5.8 (3694.8) | 2781.8 ± 3.0 (2785.0) | 2836.7 ± 5.8 (2837.3) | 2288.8 ± 4.5 (2292) |
| Balanced (w 1) | 3012.3 ± 7.7 (3008.5) | 2387.3 ± 5.1 (2397.7) | 1873.5 ± 8.6 (1867.6) | 536.8 ± 8.7 (548) |
| Max survival (w 3) | 1204.2 ± 15.3 (1205.0) | 1336.7 ± 13.7 (1363.3) | -736.0 ± 22.3 (-750.2) | -4391.5 ± 24.1 (-4357) |

Every claim reproduces within 2× error (the workers' 1000-iteration means). The paladin Max survival figure is the weakest (-1.95%), and it's the Balanced file at w=3.
Mean TPS (Balanced): warrior 1895, bear 1618, Warden 1428, paladin 1311. Healing needed (Balanced, six-boss mean): paladin 211, warrior 323, bear 488, Warden 912.
The workers' rows in `ref/tank_bench/summary.tsv` (1000 iterations) all agree with this run within noise; its header still names 23b2ec2 (finding -9).

## Does each preset win its own weight (rows preset, columns w = 1/3, 1, 3)

| | 1/3 | 1 | 3 |
|---|---|---|---|
| warrior MT / Bal / MS | **3696.8** / 3622.2 / 3423.2 | 2886.8 / **3012.3** / 2868.3 | 456.3 / 1182.3 / **1204.2** |
| paladin MT / Bal=MS | **2781.8** / 2737.5 | 2374.8 / **2387.3** | 1153.2 / **1336.7** |
| bear MT / Bal / MS | **2836.7** / 2769.2 / 2625.3 | 1763.2 / **1873.5** / 1785.0 | -1456.5 / -813.5 / **-736.0** |
| Warden MT / Bal / MS | **2288.8** / 2246.7 / 2145.2 | 498.5 / **536.8** / 511.0 | -4873.3 / -4592.8 / **-4391.5** |

Yes, everywhere. These ties are within 2× paired error: warrior MS over Bal at w=3 (+21.8 against 48.6), paladin Bal over MT at w=1 (+12.5 against 15.0), and the paladin's Bal and MS (byte-identical files).
The docs say the weights don't separate those, which the brief allows (finding -4 closed). Builds that differ:
- Warrior: Bal and MS differ only in a trinket (991018 against 990617) and the APL.
- Paladin: MT and Bal differ only in talents.
- Bear: one talent string; only the APL differs.
- Warden: Bal and MS differ in talents and the second trinket.

## Same terms

The 12 requests are equal on:
- level 60, reaction 200 ms, inFrontOfTarget;
- Engineering + Blacksmithing;
- food, Mongoose, Juju Power and Might, Dragonbreath Chili;
- Kings and improved Might;
- every raid buff and debuff, the encounter and simOptions;
- no Salvation.

Phalanx + Greater Fortitude are on all 12, and the paladin and warrior both have Dense Sharpening Stone (finding -1 holds). The fair differences:
- races are each spec's best;
- mana users take Major Mana Potion;
- Wisdom is on the Warden only (worth nothing on the paladin: -1.5 ± 7 Balanced);
- the bear's Weightstone, the Warden's Rockbiter;
- the paladin's Greater Arcane Elixir.

Open: the flask was never re-picked at Max threat. Natural Accuracy is +30.4 for the Warden at w=1/3, -23 for the warrior and 0 for the paladin (finding -7).

## Spread at Balanced, per boss (vs the mean of the other three)

- Warrior ahead on all 6 bosses: score +58% to +233%, TPS +27% to +41% on 5. Revenge, Heroic Strike and Shield Slam, all sourced; accepted in pass 1.
- Paladin ahead on all 6: score +15% to +73%, healing needed -43% to -96%. Finding -2: the mechanics are checked; three calls wait on the user.
- Warden behind on all 6: score -48% to -130%, healing needed +82% to +263%. Ruled a class gap (standards 40).
- Bear: +46% score on Broodlord, -19% on Onyxia; otherwise in line.

Logs: the Anniversary logs' main tanks are warriors (druids on four), with no paladin or shaman main tank. The sim ranks the paladin second at Balanced and first in damage taken, the disagreement noted in pass 1 and parked as finding -2's calls.

## Equal care (round 2)

All four classes searched at 1,000 iterations, confirmed at 3,000, and scored the six-boss mean at the preset's weight. What differed:
- The warrior also re-searched races, trinkets and the wrist enchant, and committed its raw output.
- The Warden kept round 1's race and trinket.
- The paladin ran no gear or enchant A/B.
- The bear skipped enchants and consumables.

All of this is in finding -8 (low; nothing measured lost).
UI presets match their requests on gear, talents, race and consumables. They don't match on individual buffs (paladin: Wisdom only; Warden: none) or on professions (warrior, Warden): finding -6.

## DPS Report, Tanking mode

- Code: no per-class special case. tankheal.py runs every tank with the same 7 boss files, the same 1000 iterations and seed, and 6 scored bosses (Ragnaros shown, scored for none). Only the legacy warrior_protection.json is excluded. report.js uses one path for all four tanks: the same weights, sort, ± display and class colours. The notes are uneven in count, but none is unfair, and the Warden's "needs the most healing" fits standards 40.
- Data: stale. ui/dps_report/data/tank_*.json was committed in e9e1a637d from a bench at 51e2f5ea (13:47Z). Every round-2 preset commit is later: paladin 14:52, Warden 15:00, warrior 15:09, bear 15:30. The warrior Max threat and all three Warden talent strings differ from the requests. A separate session is regenerating it, and the new data hasn't landed on master, so this pass doesn't judge the numbers. No finding: the next pass checks the regenerated data.

## Findings

- Closed: -1 (elixirs and stone), -3 (ruled and checked), -4 (each preset wins its weight; ties documented).
- Still open: -2 (paladin damage taken; user calls a, b, c), -5 (legacy warrior file; the healing session's call).
- New: -6 medium UI buffs and professions; -7 medium Warden Max threat flask; -8 low uneven search; -9 low stale notes and docs.
