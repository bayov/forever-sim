# Fairness between tanks

The tank_fairness auditor keeps this current (orchestration/audit/passes/tank_fairness.md): the tank ranking at each
preset weight, its reasons, and the open questions. Last pass: 2026-11-04, master f62bfc7
(orchestration/audit/reports/tank_fairness.md; table in orchestration/audit/ref/tank_fairness/summary.tsv).

## Ranking (tankbench, the six boss files, each spec's matching preset)

| weight | order (best first) | note |
|---|---|---|
| Max threat (w 1/3) | Warrior, then Paladin / Bear, Warden last | warrior 1.3 to 2.3x the others' score on every boss |
| Balanced (w 1) | Warrior, Paladin, Bear, Warden | paladin ahead of bear on all 6 bosses (bear ahead with no encounter) |
| Max survival (w 3) | Warrior or Paladin (paladin on Onyxia and Ragnaros), Bear, Warden | |

TPS: warrior 2150 to 2360, bear 1650 to 1740, Warden 1390 to 1650, paladin 1230 to 1490.
Healing needed (Onyxia): paladin 49, warrior 237, bear 388, Warden 646.

## Why

- **Warrior threat**: Forever's Revenge (97 to 145 + 28% AP, 2.25 threat a point; measured in game) with Heroic
  Strike and Shield Slam make about 80% of its TPS. Sourced; taken as real.
- **Paladin survival**: same armor, avoidance and block share as the warrior, but about half the damage taken.
  Likely its block value multipliers (Shield Specialization, Steadfast) and flat reductions (Improved Righteous
  Fury, Iron Creed). Under review (finding tank_fairness-20261104-2).
- **Warden**: Defense 346, little block, 15% crushing blows, lower armor: 2 to 3x the healing needed
  (finding -3).
- **Bear**: most armor and health, no block so 15% crushes; middle of the pack.

## Unequal terms (open)

- Only the bear takes Elixir of the Phalanx and Elixir of Greater Fortitude; worth +200 to +1700 Balanced score to
  the others (finding -1). Fixing it narrows nothing between warrior and paladin but helps the Warden most.
- Flasks are right per spec at the Balanced weight (Titans for warrior, paladin, Warden; Natural Accuracy for bear).

## Open questions

1. Is the paladin's low damage taken right (finding -2)? The Anniversary logs show no paladin main tanks.
2. Is the Warden meant to rank last, or is its set's Defense too low (finding -3)?
3. Why is the paladin's TPS 22 to 30% under the others? Not yet explained.
4. Do the paladin and warrior have a Max survival build distinct from Balanced (finding -4)?
5. The req files hard-code the tank's own debuffs; tankbench strips them through `tankscore.OwnDebuffs`. Do rotopt,
   healbench and the tests too?
