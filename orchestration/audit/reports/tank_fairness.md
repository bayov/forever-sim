# tank_fairness pass 3: 2026-11-04, master dc303b0

Forced by hand after round 3: the engine fix fd9538c (mana-gain threat unmultiplied, disjoint tank seeds) and the four
class workers' fixes for pass 2's findings, with re-optimized presets.
Raw table: `ref/tank_fairness/summary.tsv` (72 rows: 12 preset requests × 6 scoring bosses, seed 1, scores at all three weights).

## How

- `go build -tags=with_db -o tankbench ./tools/tankbench` (protoc-gen-go needs `PATH=$PATH:/root/go/bin`), then per boss
  `tankbench -iters 3000 -encounter tools/encounters/<boss>.json tools/tankbench/req/*_{max_threat,balanced,max_survival}.json`,
  seed 1 and again on seed 7. Six-boss mean error = sqrt(Σerr²)/6.
- Same terms: all 12 requests flattened and compared by script, with gear, talents and rotation set aside. UI presets
  compared slot by slot (gear, enchants, talents, APL JSON, race, buffs, debuffs, consumables, professions, options).
- Equal care: orchestration/tank_round3/*.md, orchestration/tank_handoff/<class>/round3/, and the round-3 sections of docs/tank_*.md.
- Variants (scratch requests): the flask swapped on all 12 presets, Dirge's Kickin' Chimaerok Chops on the warrior, paladin and
  Warden, and warrior queueDelay 250.

## Own-weight six-boss means (3000 iterations, seed 1 ± error [seed 7]) against round 3's claims

| preset | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| Max threat (w 1/3) | 3700.0 ± 5.8 [3694.7] (3695.6) | 2800.7 ± 3.1 [2796.8] (2793.7) | 2844.5 ± 5.9 [2821.3] (2842.5) | 2421.8 ± 6.5 [2409.7] (2430.0) |
| Balanced (w 1) | 3012.3 ± 7.7 [2991.3] (3010.6) | 2418.3 ± 5.5 [2405.3] (2415.0) | 1849.0 ± 9.1 [1829.0] (1851.8) | 748.2 ± 9.7 [723.3] (767.2) |
| Max survival (w 3) | 1204.2 ± 15.3 [1166.7] (1202.1) | 1449.0 ± 14.9 [1409.7] (1445.2) | -655.5 ± 22.0 [-718.8] (-633.8) | -3996.3 ± 22.1 [-3986.0] (-3970.8) |

- Every claim reproduces within 2× error on seed 1.
- The Warden's Balanced is soft: it was a 1,000-iteration figure, and seed 7 is 44 under it. This is noted under finding -9.
- Seed-to-seed spread at w=3 reaches about 40, more than the quoted errors.

## Does each preset win its own weight (rows: preset; columns: w = 1/3, 1, 3)

| | 1/3 | 1 | 3 |
|---|---|---|---|
| warrior MT / Bal / MS | **3700** / 3622 / 3423 | 2905 / **3012** / 2868 | 519 / 1182 / **1204** |
| paladin MT / Bal=MS | **2801** / 2742 | **2430** / 2418 | 1320 / **1449** |
| bear MT / Bal / MS | **2845** / 2742 / 2581 | 1729 / **1849** / 1772 | -1617 / -828 / **-656** |
| Warden MT / Bal / MS | **2422** / 2350 / 2247 | 531 / **748** / 686 | -5144 / -4057 / **-3996** |

Within noise:
- the warrior's Max survival over Balanced at w=3 (+22 ± 24);
- the Warden's Max survival over Balanced at w=3 (+60 ± 35);
- at w=1 the paladin's Max threat edges out Balanced (+11.8 ± 7.9, 1.5×).

The paladin's Max threat and Balanced differ only in talents and one enchant; its Balanced and Max survival are byte-identical. Finding -4 allowed this kind of tie and it was documented, so it isn't re-raised.

## Same terms

All 12 requests are equal on:
- level 60, 200 ms reaction, 5 yd, in front;
- Kings and improved Might;
- Engineering and Blacksmithing;
- all 18 raid buffs and 13 debuffs (the tank's own debuffs stripped by `tankscore.OwnDebuffs`);
- the encounter and sim options;
- no Salvation.

The accepted differences hold:
- races are each spec's best (the Warden's now differ by preset: Night Elf, Night Elf, Undead);
- mana users take Major Mana Potion;
- Wisdom is on the Warden only;
- the bear's weightstone and the Warden's Rockbiter;
- the paladin's Greater Arcane Elixir.

New since pass 2, and A/B'd:
- The bear takes Flask of Natural Accuracy on all three presets (Titans -309 at Max threat). The Warden's Max threat takes it too.
- Every other preset keeps Titans, and each swap measured here confirms its current flask (-24 to -270 on two seeds, paladin Max threat -6.5 ± 4.3).
- The bear's Max survival eats Dirge's chops. No other tank was tried on it, and it gains for the paladin and Warden: finding -10.
- Enchants differ by class, each A/B'd: the paladin takes Lesser Block on the shield (its call 11) and the Warden Superior Defense on the cloak.

Open question 4 from pass 2 is answered: rotopt goes through `tankscore.MakeTank` (tools/rotopt/run.go:211, main.go:274), which calls `OwnDebuffs`.

## Spread at Balanced (per boss, vs the mean of the other three)

| boss | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| Magmadar | 3438 (+56%) | 2813 (+17%) | 2239 (-14%) | 1560 (-45%) |
| Golemagg | 3263 (+64%) | 2590 (+17%) | 2111 (-11%) | 1274 (-52%) |
| Onyxia | 3155 (+81%) | 2978 (+65%) | 1706 (-24%) | 558 (-79%) |
| Broodlord | 2678 (+127%) | 1928 (+35%) | 1872 (+29%) | -262 (-112%) |
| Vaelastrasz | 3435 (+53%) | 2764 (+12%) | 2220 (-16%) | 1734 (-38%) |
| Chromaggus | 2105 (+214%) | 1437 (+61%) | 946 (-10%) | -375 (-125%) |

| Balanced, six-boss mean | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| TPS | 1895 | 1306 | 1610 | 1441 |
| DTPS | 366 | 288 | 570 | 730 |
| Healing needed | 323 | 197 | 488 | 853 |

Explanations, all unchanged from pass 2:
- **Warrior threat (+22% to +41% TPS):** Revenge, Heroic Strike and Shield Slam, all sourced.
- **Paladin damage taken (healing needed -45% to -99%):** finding -2. The mechanics are checked; three user calls remain.
- **Warden healing need (+83% to +236%):** a class gap by ruling (standards 40).
- **Paladin TPS (about 21% under the others):** still unexplained (open question below).

The logs disagree as before: the Anniversary main tanks are warriors (druids on four bosses), with no paladin or shaman.

## Equal care (round 3)

| | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| search / confirm | 1k / 3k | 1k / 3k | 1k / 3k | 3k A/Bs on the round-2 search |
| final numbers | 10k, seeds 1 and 2 | 10k, seed 1 | 3k (UI: 1k) | 1k |
| races × 3 weights | yes | yes | yes | yes |
| talents, APL knobs | yes | yes (Bulwark frozen) | yes | yes |
| enchants A/B'd | cloak, wrist, shield | wrist, cloak, boots, shield | 8+ slots | wrist, cloak, shield |
| consumables A/B'd | none | none | all, each weight | Max threat flask only |
| raw output committed | yes | yes | yes | yes |

The uneven part that matters is consumables, and it is measured in finding -10.

No class's choices rest only on 1,000-iteration runs. The Warden's headline scores do, and they are slightly high (finding -9).

## UI = bench

Script-checked for all 12 presets: gear, enchants, talents, APL JSON, race, professions, individual and raid buffs, and consumables match their requests.

Finding -6 is closed. One exception remains, the warrior:
- queueDelay is 250 in the UI and 0 on the bench, which is -75 at Max threat;
- its UI debuffs keep Thunder Clap, the Shouts and the tank Sunder ramp that the bench strips.

Both are finding -11.

Two small things, not findings:
- The bear's Max survival food comes from its encounter preset, not its build.
- The paladin and Warden UIs default to a non-tank build until a build is chosen.

## DPS Report, Tanking mode

- **Data:** the regen had not landed on master at dc303b0. The last tank data, 4feee54c2 (17:13Z), predates every round-3 preset commit (17:54 to 18:20). tank_builds.json still holds:
  - the Warden's round-2 races and talents, at 2292 / 548 / -4357 against round 3's 2430 / 767 / -3971;
  - the paladin's old Balanced wrist and 1363.3 Max survival;
  - the Warden's Max threat on Titans.

  This is recorded here, not as a finding, because a regen session is deploying it now. The next pass checks tank_builds.json against the requests.
- **Code:** tankheal.py and report.js treat all four tanks alike: the same 7 boss files, 1,000 iterations, weights 1/3, 1 and 3, six scored bosses, and no per-class case.

## Findings

- **Closed:** -6 (UI buffs and professions), -7 (the Warden's Max threat flask holds, Titans -24.7 / -32.7), -8 (round 3 searched all four with raw output; what's left is -10).
- **Still open:** -2 (paladin damage taken; user calls), -5 (legacy warrior file; the healing session's call), -9 (partly fixed; the tank_bench/summary.tsv header, docs/tank_sim.md status, paladin call 11, warrior doc line 199 and the Warden's 1k Balanced figure remain).
- **New:**
  - -10 (high): consumables were searched for the bear only; Dirge's chops gain +62 on paladin Max survival, +42 on Warden Balanced and +157 on Warden Max survival.
  - -11 (medium): the warrior UI's queueDelay 250 and bench-stripped debuffs.
