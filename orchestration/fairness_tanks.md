# Fairness between tanks

The tank_fairness auditor keeps this current (orchestration/audit/passes/tank_fairness.md): the tank ranking at each
preset weight, its reasons, and the open questions. Last pass: 2026-11-12, pass 10, master 67ad7c6a4, a forced re-audit after the day's tank fixes (Warden re-tune,
flask re-picks, paladin Wizard Oil and ranged Judgement, tank enchants, bear Frenzied Regeneration, warrior Shield Wall,
Windfury Totem no longer procced by Heroic Strike). The report has the details (orchestration/audit/reports/tank_fairness.md);
the table is in orchestration/audit/ref/tank_fairness/summary.tsv.

## Ranking (tankbench 3000 iterations, seed 1, six-boss mean, each spec's matching preset)

| weight | order (best first) | scores |
|---|---|---|
| Max threat (w 1/3) | Warrior, Paladin, Bear, Warden | 3810 / 3187 / 2663 / 2535 |
| Balanced (w 1) | Warrior, Paladin, Bear, Warden | 3106 / 2774 / 1717 / 1184 |
| Max survival (w 3) | Paladin, Warrior, Bear, Warden | 1711 / 1573 / 14 / -2385 |

The warrior lost 130 to 160 at Max threat and Balanced since pass 9: Windfury Totem procs from white swings only
(user ruling 2026-11-12). The bear's Max survival rose 590 (Titans, Frenzied Regeneration at 70 Rage); the paladin now
leads Max survival by 138, and its Stoneshield potion action (finding -20261105-1, +402) would widen that.

TPS (Balanced): warrior 1893, paladin 1506, bear 1425, Warden 1343. Healing needed (Balanced): paladin 241, warrior 344,
bear 474, Warden 665.

Each preset wins its own weight (seed 1). The paladin's Max threat and Balanced share one build since round 5 and tie
at Max threat (its call 14).

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
- **Bear:** most armor and health, no block. Third since the Natural flask cut, which took its Precision expertise and
  Stamina; the paladin, on Supreme Power at Max threat and Balanced, lost little.
- **Warden's drop:** besides the flask, Blizzard's Nov 10 tuning (Stormstrike 100% and 8 sec, Mental Dexterity 33/66/100%)
  cost it about 15% on Magmadar. That is real; the 2026-11-12 re-tune of its talents and APLs won back +20 / +13 / +342
at its three weights (most of it the Titans at Max survival).

## Unequal terms (open)

- The paladin never drinks its potion. `autocastOtherCooldowns` drinks no defensive potion in a tank score, and the
  paladin's requests have no explicit potion action as the other three tanks' do. Greater Stoneshield with the action
  gains +58 / +172 / +402 at the paladin's three weights (pass 10, two seeds; finding -20261105-1).
- The Warden plays races a Forever shaman can't be: Night Elf at Max threat and Balanced, Undead at Max survival
  (client CharBaseInfo: shaman is Dwarf, Orc, Tauren, Troll or Skyborne Windshaper). Going legal costs about -60 (Troll
  or Orc) / -80 (Troll) / -158 (Tauren) (finding -20261105-2). The Warden is last at every weight either way.
- Raid-only consumables (finding -20261112-1, user ruling): the warrior keeps the Elemental Sharpening Stone (+20 / +20
  / +13 over the Dense Sharpening Stone) that the paladin swapped for the Wizard Oil and the bear declined; the
  Warden's Mageblood Potion, also raid-only, is worth nothing measurable. Either the rule binds all four tanks or none
  (then the bear takes the stone, +21 at Max survival, and the paladin the Brilliant Wizard Oil, +52).
- Checked and equal: the paladin has no Blessing of Wisdom (every other mana spec does), but Wisdom gains it nothing
  (-1 to -8, noise). Everything else matches: all four tanks have searched every slot of the full raid set
  (standards 40); the Warden's other mana consumables (Demonic Rune, Greater Intellect) are noise for the paladin.

## Open questions

1. The paladin calls in finding -2 (user).
2. The paladin's TPS was about 18% under the others'. Since the flask cut it is second (1500, against the bear's 1425), so
   the question is mostly moot.
3. Answered 2026-11-12: the Warden was re-tuned for the Nov 10 shaman notes (talents, Searing / Fire Nova, Healing
   Stream at Max survival and Balanced); +37 / +50 / +356 since pass 9.
4. Raid-only consumables for tanks (finding -20261112-1, user).

Answered:
- rotopt strips the tank's own debuffs through `tankscore.MakeTank` → `OwnDebuffs` (tools/rotopt/run.go:211), as tankbench does.
- The warrior is scored at queueDelay 0 on the bench and in the UI (88d044d).
- The DPS Report's Tanking data matches the requests' talents, races and scores (re-sim at 877bcec8c).
