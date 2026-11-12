# Fairness between tanks

The tank_fairness auditor keeps this current (orchestration/audit/passes/tank_fairness.md): the tank ranking at each
preset weight, its reasons, and the open questions. Last pass: 2026-11-12, pass 11, master 4fb90c298, after the main-hand Windfury Totem rule (standards 17) and
the presets re-tuned on it (bear Frenzied Regeneration 80 Rage, Swipe filler, Barkskin 20 s; warrior Dense stone; Warden
Greater Mageblood Elixir). The report has the details (orchestration/audit/reports/tank_fairness.md); the table is in
orchestration/audit/ref/tank_fairness/summary.tsv.

## Ranking (tankbench 3000 iterations, seed 1, six-boss mean, each spec's matching preset)

| weight | order (best first) | scores |
|---|---|---|
| Max threat (w 1/3) | Warrior, Paladin, Bear, Warden | 4202 / 3187 / 2789 / 2536 |
| Balanced (w 1) | Warrior, Paladin, Bear, Warden | 3426 / 2774 / 1865 / 1189 |
| Max survival (w 3) | Warrior, Paladin, Bear, Warden | 1762 / 1711 / 191 / -2385 |

The main-hand Windfury rule gave the warrior +392 / +320 / +189 (Sunder Armor and Revenge roll the totem) and the bear
+126 / +148 / +177 (Maul, plus its re-tune); the paladin and Warden are unchanged. The warrior now leads Max survival by
51, but the paladin's potion action (finding -20261105-1, +402) and its agility enchants (-20261112-2, +78) would each
put the paladin back ahead.

TPS (Balanced): warrior 2052, paladin 1506, bear 1483, Warden 1345. Healing needed (Balanced): paladin 241, warrior 340,
bear 464, Warden 665.

Each preset wins its own weight (seed 1). The paladin's Max threat and Balanced share one build since round 5 and tie
at Max threat (its call 14). Every flask still wins at its preset after the Windfury rule (Titans stays behind Natural
Precision at the warrior's and bear's Balanced, by 125 and 82).

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
- Enchants (finding -20261112-2): head / legs Lesser Arcanum of Voracity (Agility, or Strength for the warrior's
  threat), Necklace Agility and Gloves Greater Agility beat the current picks on the paladin (Max survival +78 / +62,
  Balanced +26), the Warden (+16 to +52), the warrior (+6 to +21) and the bear's Max survival (+14 to +24).
- Race (finding -20261112-3): the bear's Max survival plays Night Elf; Tauren is +32 / +12 since its re-tune.
- Raid-only consumables: settled 2026-11-12 (user ruling, option a): no tank uses a raid-only stone, oil or elixir.
- Checked and equal: the paladin has no Blessing of Wisdom (every other mana spec does), but Wisdom gains it nothing
  (-1 to -8, noise). Everything else matches: all four tanks have searched every slot of the full raid set
  (standards 40); the Warden's other mana consumables (Demonic Rune, Greater Intellect) are noise for the paladin.

## Open questions

1. The paladin calls in finding -2 (user).
2. The paladin's TPS was about 18% under the others'. Since the flask cut it is second (1500, against the bear's 1425), so
   the question is mostly moot.
3. Answered 2026-11-12: the Warden was re-tuned for the Nov 10 shaman notes (talents, Searing / Fire Nova, Healing
   Stream at Max survival and Balanced); +37 / +50 / +356 since pass 9.
4. Enchants and the bear's Max survival race (findings -20261112-2, -3; worker fixes).

Answered:
- rotopt strips the tank's own debuffs through `tankscore.MakeTank` → `OwnDebuffs` (tools/rotopt/run.go:211), as tankbench does.
- The warrior is scored at queueDelay 0 on the bench and in the UI (88d044d).
- The DPS Report's Tanking data matches the requests' talents, races and scores (re-sim at 877bcec8c).
