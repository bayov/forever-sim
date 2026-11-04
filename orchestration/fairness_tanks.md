# Fairness between tanks

The tank_fairness auditor keeps this current (orchestration/audit/passes/tank_fairness.md): the tank ranking at each
preset weight, its reasons, and the open questions. Last pass: 2026-11-04, pass 3, master dc303b0 (after round 3)
(orchestration/audit/reports/tank_fairness.md; table in orchestration/audit/ref/tank_fairness/summary.tsv).

## Ranking (tankbench 3000 iterations, seed 1, six-boss mean, each spec's matching preset)

| weight | order (best first) | scores |
|---|---|---|
| Max threat (w 1/3) | Warrior, Bear, Paladin, Warden | 3700 / 2845 / 2801 / 2422 |
| Balanced (w 1) | Warrior, Paladin, Bear, Warden | 3012 / 2418 / 1849 / 748 |
| Max survival (w 3) | Paladin, Warrior, Bear, Warden | 1449 / 1204 / -656 / -3996 |

TPS (Balanced): warrior 1895, bear 1610, Warden 1441, paladin 1306. Healing needed (Balanced): paladin 197, warrior 323, bear 488, Warden 853.
Each preset wins at its own weight, except within noise: the warrior's and Warden's Max survival over Balanced at w=3, and at w=1 the paladin's Max threat edges its Balanced (+12 ± 8). The paladin's Balanced and Max survival are one build.

## Why

- **Warrior threat**: Forever's Revenge with Heroic Strike and Shield Slam. Sourced; taken as real.
- **Paladin survival**: about the warrior's armor, avoidance and block share, but much less damage taken. Every effect is checked
  against Forever's sources (block value ×1.6 from Shield Specialization and Steadfast, Templar's Bulwark, Improved Righteous Fury,
  Iron Creed, Steadfast Libram). Three calls are left for the user (finding -2): block value priced at 0.4 for a ×1.6 class, Bulwark
  on a timer in a scored preset, and the Libram's +30 in game.
- **Warden**: last everywhere, a mail class with little block. A real class gap by user ruling (standards 40); not compensated.
- **Bear**: most armor and health, no block; middle of the pack. Flask of Natural Accuracy is worth about 300 to it at Max threat.

## Unequal terms (open)

- Consumables were searched for the bear only. Dirge's chops gain for the paladin's Max survival (+62), the Warden's Balanced (+42)
  and Max survival (+157), so those presets are under-built (finding -10). Each tank's flask choice is confirmed.
- The warrior UI runs queueDelay 250 (bench 0; -75 at Max threat) and keeps the debuffs the bench strips (finding -11). Bench unaffected.
- The DPS Report's Tanking data was still round 2's at dc303b0; a regen was deploying (check it next pass).

## Open questions

1. The paladin calls in finding -2 (user).
2. Why is the paladin's TPS about 21% under the others'? Not yet explained.
3. Is the warrior to be scored at queueDelay 0 (bench) or 250 (UI)? A ruling if 250 is the intended realism guard (finding -11).

Answered: rotopt strips the tank's own debuffs through `tankscore.MakeTank` → `OwnDebuffs` (tools/rotopt/run.go:211), as tankbench
does. The Warden keeps no round-1 picks unverified (round 3 re-ran races and trinkets).
