# Fairness between tanks

The tank_fairness auditor keeps this current (orchestration/audit/passes/tank_fairness.md): the tank ranking at each
preset weight, its reasons, and the open questions. Last pass and full run: 2026-11-15, pass 16, master
25ce81de7, after the Nov 13 hotfixes (PR #130: Windfury Totem rolls on every landed main hand melee hit, standards 17;
Judgement of Command melee again) and the warrior's Balanced and Max survival re-tune (f2a5dfe25). The report has the details
(orchestration/audit/reports/tank_fairness.md); the table is in orchestration/audit/ref/tank_fairness/summary.tsv.

## Ranking (tankbench 3000 iterations, seed 1, six-boss mean, each spec's matching preset)

| weight | order (best first) | scores |
|---|---|---|
| Max threat (w 1/3) | Warrior, Paladin, Bear, Warden | 4223 / 3275 / 3123 / 2644 |
| Balanced (w 1) | Warrior, Paladin, Bear, Warden | 3452 / 2843 / 2131 / 1280 |
| Max survival (w 3) | Paladin = Warrior (tie), Bear, Warden | 1798 / 1794 / 364 / -2242 |

The Nov 13 Windfury rule (every landed main hand hit rolls the totem) gave the bear +328 / +260 / +129 (12% at Max
threat, TPS +163) and the Warden +88 / +61 at Max threat / Balanced; the paladin +38 at Balanced; the warrior nothing
(its rolling strikes already rolled; Shield Slam is a shield hit). Measured by building tankbench at b02b98c72 and
d4325b775 (PR #130's merge), same requests. The bear is now 5% behind the paladin at Max threat (was 14%). Max survival
stays a tie within the error (paladin 1797.7 ± 13.1, warrior 1793.8 ± 12.6); the paladin's potion action (finding
-20261105-1) is still open and would decide it.

TPS (Balanced): warrior 2063, bear 1628, paladin 1533, Warden 1387. Healing needed (Balanced): paladin 241, warrior 342,
bear 473, Warden 666.

Each preset wins its own weight on both seeds, and the three builds are distinct for every spec. Every flask still wins
at its preset (pass 11).

## Why

- **Warrior threat:** Forever's Revenge with Heroic Strike and Shield Slam, and since the main-hand Windfury rule
  Sunder Armor and Revenge rolling the totem (logged: standards 17). Sourced; taken as real.
- **Paladin survival:** it has about the warrior's armor, avoidance and block share, but takes much less damage. Every
  effect is checked against Forever's sources: block value ×1.6 from Shield Specialization and Steadfast, Templar's
  Bulwark, Improved Righteous Fury, Iron Creed and the Steadfast Libram. Three calls are left for the user (finding -2):
  - block value priced at 0.4 for a ×1.6 class;
  - Bulwark on a timer in a scored preset;
  - the Libram's +30 in game.
- **Warden:** last everywhere, a mail class with little block. A real class gap by user ruling (standards 40); not
  compensated. The full raid set narrowed it (healing needed about twice the others' mean, from up to 3.4×).
- **Bear:** most armor and health, no block. Since the Nov 13 Windfury rule its main hand specials roll the totem too,
  which is most of its +12% at Max threat (whether each bear ability counts as a main hand hit is tank_mechanics'). Third since the Natural flask cut, which took its Precision expertise and
  Stamina; the paladin, on Supreme Power at Max threat and Balanced, lost little.
- **Warden's drop:** besides the flask, Blizzard's Nov 10 tuning (Stormstrike 100% and 8 sec, Mental Dexterity 33/66/100%)
  cost it about 15% on Magmadar. That is real; the 2026-11-12 re-tune of its talents and APLs won back +20 / +13 / +342
at its three weights (most of it the Titans at Max survival).

## Unequal terms (open)

- The paladin never drinks its potion. `autocastOtherCooldowns` drinks no defensive potion in a tank score, and the
  paladin's requests have no explicit potion action as the other three tanks' do. Greater Stoneshield with the action
  gains +61 / +168 / +368 at the paladin's three weights (pass 12, on the new enchants; finding -20261105-1).
- The Warden plays races a Forever shaman can't be: Night Elf at Max threat and Balanced, Undead at Max survival
  (client CharBaseInfo: shaman is Dwarf, Orc, Tauren, Troll or Skyborne Windshaper). Going legal costs about -50 (Orc)
  / -77 (Troll) / -124 (Tauren) (pass 12; finding -20261105-2). The Warden is last at every weight either way.
- Raid-only consumables: settled 2026-11-12 (user ruling, option a): no tank uses a raid-only stone, oil or elixir.
- Enchants and the bear's Max survival race: settled 2026-11-12 (findings -20261112-2, -3). Every preset carries the
  best legal Phase 1 enchant per slot for its own weight (10,000 iterations, seeds 1 and 2, paired): paladin +75 / +53 /
  +76 (Holy Power gloves, Voracity Agility, a spell power neck at Max threat and Balanced), Warden +20 / +28 / +87,
  warrior +21 / +9 / +14, bear Max survival +19 (gloves) and +21 (Tauren).
- Checked and equal: the paladin has no Blessing of Wisdom (every other mana spec does), but Wisdom gains it nothing
  (-1 to -8, noise). Everything else matches: all four tanks have searched every slot of the full raid set
  (standards 40); the Warden's other mana consumables (Demonic Rune, Greater Intellect) are noise for the paladin.

## Open questions

1. The paladin calls in finding -2 (user).
2. The paladin's TPS was about 18% under the others'. Since the flask cut it is second (1500, against the bear's 1425), so
   the question is mostly moot.
3. Answered 2026-11-12: the Warden was re-tuned for the Nov 10 shaman notes (talents, Searing / Fire Nova, Healing
   Stream at Max survival and Balanced); +37 / +50 / +356 since pass 9.
4. Answered 2026-11-12: enchants and the bear's Max survival race (findings -20261112-2, -3) fixed by the tank worker.

Answered:
- rotopt strips the tank's own debuffs through `tankscore.MakeTank` → `OwnDebuffs` (tools/rotopt/run.go:211), as tankbench does.
- The warrior is scored at queueDelay 0 on the bench and in the UI (88d044d).
- The DPS Report's Tanking data matches the requests' talents, races and scores (re-sim at 877bcec8c).
