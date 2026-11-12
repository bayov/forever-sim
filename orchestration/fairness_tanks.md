# Fairness between tanks

The tank_fairness auditor keeps this current (orchestration/audit/passes/tank_fairness.md): the tank ranking at each
preset weight, its reasons, and the open questions. Last pass: 2026-11-11, pass 8, master 707ae92. All 12 presets were re-simmed. Since pass 6 the Natural flasks were
halved (client 70862) and Blizzard's Nov 10 shaman tuning cut Stormstrike and Mental Dexterity, and every score fell.
The report has the details (orchestration/audit/reports/tank_fairness.md); the table is in
orchestration/audit/ref/tank_fairness/summary.tsv.

## Ranking (tankbench 3000 iterations, seed 1, six-boss mean, each spec's matching preset)

| weight | order (best first) | scores |
|---|---|---|
| Max threat (w 1/3) | Warrior, Paladin, Bear, Warden | 3970 / 3166 / 2663 / 2498 |
| Balanced (w 1) | Warrior, Paladin, Bear, Warden | 3236 / 2751 / 1717 / 1134 |
| Max survival (w 3) | Warrior = Paladin, Bear, Warden | 1564 / 1559 / -579 / -2741 |

Re-picked flasks and the Warden re-tune (2026-11-12, finding -20261111-1 fixed; 10,000 iterations, seed 1): Max
threat 3969 / 3188 / 2661 / 2515, Balanced 3241 / 2775 / 1719 / 1147, Max survival paladin 1731, warrior 1608, bear
-478, Warden -2396 (Titans at Max survival for the warrior, bear and Warden and at the Warden's Balanced, Supreme Power
for the paladin, Wizard Oil on the paladin). The order is the same; the paladin now leads Max survival by 122. The
paladin's Stoneshield potion action (finding -20261105-1) would widen that.

TPS (Balanced): warrior 1950, paladin 1500, bear 1425, Warden 1380. Healing needed (Balanced): paladin 245, warrior 337,
bear 474, Warden 713.

Each preset wins its own weight (seed 1). The paladin's Max threat and Balanced share one build since round 5 and tie
at their weights (its call 14).

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
  paladin's round 4 A/B ran the Stoneshield without the explicit potion action the other three tanks use. The Stoneshield
  with the action gains +58 / +172 / +309 at the paladin's three weights (finding -20261105-1).
- The Warden plays races a Forever shaman can't be: Night Elf at Max threat and Balanced, Undead at Max survival
  (client CharBaseInfo: shaman is Dwarf, Orc, Tauren, Troll or Skyborne Windshaper). Night Elf's Elune's Light is
  worth about +76 / +109 over the best legal race (Orc), Undead about +43 over Tauren (finding -20261105-2). The
  Warden is last at every weight either way; its legal scores are lower by about 2.4% / 6% / 2%.
- Raid-only consumables: the warrior's Elemental Sharpening Stone and the Warden's Mageblood Potion come from raid
  sources the consumables pass rules out for phase 1 DPS specs (the paladin dropped its stone for the Wizard Oil on
  2026-11-12; the bear's stone, +21 at Max survival, and the paladin's Brilliant Wizard Oil, +52, were not adopted).
  A ruling on whether that rule binds the tanks too.
- Everything else matches: all four tanks have searched every slot of the full raid set (standards 40). The mana
  consumables the Warden takes (Demonic Rune, Mageblood, Greater Intellect) are noise for the paladin.

## Open questions

1. The paladin calls in finding -2 (user).
2. The paladin's TPS was about 18% under the others'. Since the flask cut it is second (1500, against the bear's 1425), so
   the question is mostly moot.
3. The Warden after the Nov 10 shaman tuning: do its talents and APL still fit?

Answered:
- rotopt strips the tank's own debuffs through `tankscore.MakeTank` → `OwnDebuffs` (tools/rotopt/run.go:211), as tankbench does.
- The warrior is scored at queueDelay 0 on the bench and in the UI (88d044d).
- The DPS Report's Tanking data matches the requests' talents, races and scores (re-sim at 877bcec8c).
