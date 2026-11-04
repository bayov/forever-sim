# Fairness between tanks

The tank_fairness auditor keeps this current (orchestration/audit/passes/tank_fairness.md): the tank ranking at each
preset weight, its reasons, and the open questions. Last pass: 2026-11-04, pass 2, master b38ff16 (after round 2)
(orchestration/audit/reports/tank_fairness.md; table in orchestration/audit/ref/tank_fairness/summary.tsv).

## Ranking (tankbench 3000 iterations, six-boss mean, each spec's matching preset)

| weight | order (best first) | scores |
|---|---|---|
| Max threat (w 1/3) | Warrior, Bear, Paladin, Warden | 3697 / 2837 / 2782 / 2289 |
| Balanced (w 1) | Warrior, Paladin, Bear, Warden | 3012 / 2387 / 1874 / 537 |
| Max survival (w 3) | Paladin, Warrior, Bear, Warden | 1337 / 1204 / -736 / -4392 |

TPS (Balanced): warrior 1895, bear 1618, Warden 1428, paladin 1311. Healing needed (Balanced): paladin 211, warrior 323, bear 488, Warden 912.
Each preset wins at its own weight. The paladin's Balanced and Max survival are one build, and the warrior's Max survival leads its Balanced at w=3 only within noise.

## Why

- **Warrior threat**: Forever's Revenge with Heroic Strike and Shield Slam. Sourced; taken as real.
- **Paladin survival**: same armor, avoidance and block share as the warrior, but half the damage taken. Every effect was checked
  against Forever's sources (block value ×1.6 from Shield Specialization and Steadfast, Templar's Bulwark, Improved Righteous Fury,
  Iron Creed, Steadfast Libram). Three calls are left for the user (finding -2): block value priced at 0.4 for a ×1.6 class, Bulwark
  on a timer in a scored preset, and the Libram's +30 in game.
- **Warden**: last everywhere, a mail class with Defense 346 and little block. A real class gap by user ruling (standards 40); not
  compensated.
- **Bear**: most armor and health, no block; middle of the pack.

## Unequal terms (open)

- The Warden's Max threat flask should be Natural Accuracy (+30 at w=1/3); the warrior's and paladin's Titans is right (finding -7).
- UI presets: paladin and Warden lack Kings and Might, and the warrior and Warden UIs use other professions than the bench (finding -6). Bench unaffected.
- Round 2's search depth differed by class: the Warden's race and trinket, the paladin's gear and enchants, the bear's enchants (finding -8).
- The DPS Report's Tanking data predates round 2; it is being regenerated.

## Open questions

1. The paladin calls in finding -2 (user).
2. Does the Warden keep Tauren and its trinkets on the round-2 engine (finding -8)?
3. Why is the paladin's TPS about 30% under the warrior's? Not yet explained.
4. Do rotopt, healbench and the tests strip the tank's own debuffs through `tankscore.OwnDebuffs` as tankbench does? Not yet checked.
