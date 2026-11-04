# tank_fairness pass 1: 2026-11-04, master f62bfc7

First pass since the four tank specs landed on master, realigned to client 70753 and the November notes.
Raw table: `ref/tank_fairness/summary.tsv` (91 rows: 13 requests x 6 boss files + no encounter, every tankbench column).

## How

- Same terms: every field of the 13 `tools/tankbench/req/*.json` flattened and diffed.
- Numbers: `go build -tags=with_db -o tankbench ./tools/tankbench`, then
  `tankbench -encounter tools/encounters/<boss>.json tools/tankbench/req/*.json` for broodlord, chromaggus, golemagg,
  magmadar, onyxia, ragnaros, and once with no encounter (600 HPS stand-in). 3000 iterations, fixed seed (a rerun is
  byte-identical); ± on score about 8 to 50, on healing needed 2 to 3. (A fresh checkout needs
  `protoc -I=./proto --go_out=./sim/core ./proto/*.proto` before the build.)
- Consumable variants and `-actions` on the Balanced requests, Onyxia and Broodlord.

## Same terms

Equal in all 13: level 60, reaction 200 ms, Engineering + Blacksmithing, food, Dragonbreath Chili, Juju Power and
Might, Mongoose, the whole buff and debuff blocks, Salvation absent, the encounter (180 s, level 63, 3731 armor),
seed. The tank's own debuffs come out through `tankscore.OwnDebuffs` (tools/tankscore/tankscore.go:103), which
strips Thunder Clap, Demo Shout, Demo Roar and the Sunder ramp for the warrior and the two shouts/roars for the bear;
the files hard-code them, so any reader that skips OwnDebuffs (rotopt, healbench, tests: not checked) would be unfair.
Races are each spec's best by the objective (Undead warrior and paladin, Night Elf bear, Tauren Warden).

Differences: the bear alone has Phalanx + Greater Fortitude and Natural Accuracy (finding -1); the paladin alone has
Greater Arcane Elixir and no stone (noise, under -1); mana users take Major Mana Potion and the Warden Wisdom (fair);
the Warden's second trinket differs in Max survival (standard 36 allows it). Gear is each spec's synthetic set.

## Numbers (Balanced request: TPS / healing needed / death chance at the encounter's healing)

| boss | Bear | Paladin | Warden | Warrior |
|---|---|---|---|---|
| none (600 HPS) | 1551/426/0% | 1228/254/0% | 1391/748/66% | 2147/364/0% |
| broodlord | 1709/531/0% | 1413/422/0% | 1593/1818/2.8% | 2310/566/0% |
| chromaggus | 1660/709/0% | 1380/475/0% | 1513/1155/0% | 2359/649/0% |
| golemagg | 1654/341/0% | 1394/83/0% | 1541/576/0% | 2289/232/0% |
| magmadar | 1651/291/0% | 1414/31/0% | 1560/499/0% | 2277/185/0% |
| onyxia | 1738/388/0% | 1490/49/0% | 1653/646/0% | 2341/237/0% |
| ragnaros | 1655/2071/17% | 1265/1772/5.9% | 1456/2341/30% | 2314/2001/10.7% |

Score at each spec's matching preset (Max threat / Balanced / Max survival):

| boss | Bear | Paladin | Warden | Warrior |
|---|---|---|---|---|
| onyxia | 2814 / 1476 / -1549 | 3310 / 3083 / 2575 | 2582 / 288 / -6076 | 4760 / 3799 / 1349 |
| ragnaros | 1131 / -3230 / -16070 | 967 / -2828 / -14017 | 598 / -4248 / -18999 | 2903 / -1455 / -14093 |

The warrior leads Max threat and Balanced on every boss and Max survival on all but Onyxia and Ragnaros (paladin).

Outliers and their reasons:
- Warrior TPS +44% to +59% over the others' mean: Revenge (727 TPS on Onyxia from 5.9 casts), Heroic Strike (646),
  Shield Slam (482). Forever's Revenge is 97 to 145 + 28% AP at 2.25 threat a point (sim/warrior/revenge.go:24,
  measured in game, 913a5bcf69). Explained by a sourced mechanic; no finding.
- Paladin TPS -22% to -30%: no single cause found; open question.
- Paladin healing needed -43% to -90% (finding -2), Warden +89% to +259% (finding -3).
- Bear within 6% on TPS, middle on healing.

Sheets (Magmadar, Balanced): armor 61.7 / 60.4 / 54.2 / 63.7% (paladin / warrior / Warden / bear), Defense
460 / 458 / 346 / 428, avoided 47 / 48 / 33 / 50%, blocked 45 / 45 / 6 / 0%, crushed 6 / 3 / 15 / 15%.

## Against the logs

Main tanks in tools/encounters `healing.notes`: warriors on every boss, druids on Broodlord (6), Onyxia (5),
Ragnaros (4), Chromaggus and Golemagg (1); no paladin or shaman. The sim's warrior-first threat ranking agrees;
its paladin-best survival does not (finding -2). Logged tank damage taken a second against the sim's warrior/bear:
close on Broodlord, Chromaggus, Onyxia; the sim is 25 to 45% low on Golemagg and Magmadar (Magma Splash's hit and
armor cut not modelled) and 1.4 to 1.9x high on Ragnaros (Elemental Fire). These are encounter-file matters for
tank_mechanics, not fairness, since every tank meets the same file.

## Presets

Max threat tops its spec's TPS in 26 of 28 cells. Paladin and warrior Balanced = Max survival (same build), bear
Balanced ~ Max threat (finding -4).

## Consumable deltas (Balanced, TPS / healing needed / Balanced score)

| spec | variant | onyxia | broodlord |
|---|---|---|---|
| Warrior | + Phalanx + Greater Fortitude | +5 / -46 / +247 | +4 / -146 / +447 |
| Paladin | same | +3 / -37 / +199 | +4 / -106 / +327 |
| Warden | same | 0 / -62 / +319 | 0 / -557 / +1672 |
| Bear | Titans instead of Natural Accuracy | -117 / -15 / -220 | -108 / -18 / -221 |

## New findings

-1 high (elixirs), -2 medium (paladin damage taken), -3 medium (Warden), -4 medium (presets not separated),
-5 low (legacy warrior request).

## Not done

Readers of the req files other than tankbench (rotopt, healbench, tests) not checked for OwnDebuffs; paladin
block-value share not isolated; paladin's low TPS not explained.
