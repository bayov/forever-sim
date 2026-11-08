# Fairness between tanks

The tank_fairness auditor keeps this current (orchestration/audit/passes/tank_fairness.md): the tank ranking at each
preset weight, its reasons, and the open questions. Last pass: 2026-11-08, pass 6, master 9ba00c2 (shaman totem/imbue, whole-point spell rolls, heal-only downrank penalty; all 12 presets re-simmed against 73d5bfa, every six-boss mean within 0.2, so the numbers below stand) (orchestration/audit/reports/tank_fairness.md; table in orchestration/audit/ref/tank_fairness/summary.tsv).

## Ranking (tankbench 3000 iterations, seed 1, six-boss mean, each spec's matching preset)

| weight | order (best first) | scores |
|---|---|---|
| Max threat (w 1/3) | Warrior, Bear, Paladin, Warden | 4650 / 3491 / 3249 / 3169 |
| Balanced (w 1) | Warrior, Paladin, Bear, Warden | 3867 / 2826 / 2500 / 1794 |
| Max survival (w 3) | Warrior, Paladin, Bear, Warden | 2151 / 1922 / 233 / -1994 |

The Max survival order is provisional. With the Greater Stoneshield Potion and a potion action, the paladin's Max
survival is 2231 and it is first again (finding tank_fairness-20261105-1).

TPS (Balanced): warrior 2200, bear 1760, Warden 1648, paladin 1534. Healing needed (Balanced): paladin 245, warrior 317,
bear 457, Warden 692.

Each preset wins its own weight on two seeds. The paladin's Max threat and Balanced share one build since round 5 and
tie at their weights (its call 14).

## Why

- **Warrior threat:** Forever's Revenge with Heroic Strike and Shield Slam. Sourced; taken as real.
- **Paladin survival:** it has about the warrior's armor, avoidance and block share, but takes much less damage. Every
  effect is checked against Forever's sources: block value ×1.6 from Shield Specialization and Steadfast, Templar's
  Bulwark, Improved Righteous Fury, Iron Creed and the Steadfast Libram. Three calls are left for the user (finding -2):
  - block value priced at 0.4 for a ×1.6 class;
  - Bulwark on a timer in a scored preset;
  - the Libram's +30 in game.
- **Warden:** last everywhere, a mail class with little block. A real class gap by user ruling (standards 40); not
  compensated. The full raid set narrowed it (healing needed about twice the others' mean, from up to 3.4×).
- **Bear:** most armor and health, no block; middle of the pack.

## Unequal terms (open)

- The paladin never drinks its potion. `autocastOtherCooldowns` drinks no defensive potion in a tank score, and the
  paladin's round 4 A/B ran the Stoneshield without the explicit potion action the other three tanks use. The Stoneshield
  with the action gains +58 / +172 / +309 at the paladin's three weights (finding -20261105-1).
- The Warden plays races a Forever shaman can't be: Night Elf at Max threat and Balanced, Undead at Max survival
  (client CharBaseInfo: shaman is Dwarf, Orc, Tauren, Troll or Skyborne Windshaper). Night Elf's Elune's Light is
  worth about +76 / +109 over the best legal race (Orc), Undead about +43 over Tauren (finding -20261105-2). The
  Warden is last at every weight either way; its legal scores are lower by about 2.4% / 6% / 2%.
- Everything else matches: all four tanks have searched every slot of the full raid set (standards 40). The mana
  consumables the Warden takes (Demonic Rune, Mageblood, Greater Intellect) are noise for the paladin.

## Open questions

1. The paladin calls in finding -2 (user).
2. Why is the paladin's TPS about 18% under the others'? Not yet explained.

Answered:
- rotopt strips the tank's own debuffs through `tankscore.MakeTank` → `OwnDebuffs` (tools/rotopt/run.go:211), as tankbench does.
- The warrior is scored at queueDelay 0 on the bench and in the UI (88d044d).
- The DPS Report's Tanking data matches the requests' talents, races and scores (re-sim at 877bcec8c).
