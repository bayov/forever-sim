# Fairness between tanks

The tank_fairness auditor keeps this current (orchestration/audit/passes/tank_fairness.md): the tank ranking at each
preset weight, its reasons, and the open questions. Last pass: 2026-11-26, pass 34, branch phase0-tankheal 3e9748641 (the tanks wear their Phase 0 lists, standards 43: bear first at Max threat and Balanced, 21% over the warrior at Balanced; warrior first at Max survival; the paladin falls to third there, 8271 health, and Titans would give it +120, finding -20261126-1). Master's own pass 34 (same day, before the merge): master ab3fff7007 (warrior leveling batch: spell ranks by level, leveling-only stuns, swings from behind and Cannibalize; all 12 presets identical to pass 33, ranking unchanged). Pass 33: 2026-11-25, master 85691655a (wowsims/forever's logged tank threat batch, 2f42f4a44: bear Lacerate 10/3 and Primal Bite 3x threat lift bear TPS 46% and its scores +1408 / +1191 / +846, now first at Max threat and Balanced; warrior Shield Slam 2 x block value / 255 threat, Thunder Clap 1x and rage-gain threat cost it TPS 50 to 77 and -132 / -208 / -197; flat physical reduction capped at half the hit gives the paladin +43 / +65 / +97; Warden noise). Pass 32: 2026-11-25, master 0404484fb (Rage from hits taken now carries the warrior stance's 0.9 / 1.1 and twice the Rage from spells, ticks and Physical specials, b4cab8498: warrior -2.5 / -10.5 / -1.2, bear +1 / +21 / +108 at Max threat / Balanced / Max survival, the bear's gain on Onyxia, Broodlord, Magmadar and Vaelastrasz; Arcane Brilliance now overwrites the Warden's Elixir of Greater Intellect, f1942793d: Warden -28 / -38 / -54, TPS -1%; Windfury / Flametongue Totem raid scope and world-buff rules reach no tank score; paladin identical; ranking unchanged). Pass 29: 2026-11-24, master 566067a39 (warrior Unbridled Wrath off Shield Slam, Revenge and Shield Bash, and Heroic Strike / Cleave spending a Flurry charge, both from wowsims/forever's logs; druid Rip / Rake shares reach no bear preset: warrior -3.5 / -5.7 / -22.0 at Max threat / Balanced / Max survival, TPS -2 to -6, every other preset identical; ranking unchanged). Pass 27: 2026-11-23, master 5518f4873 (a guard in `sim/core/attack.go` so an empty weapon slot never swings; every tank wields a weapon, all 84 preset rows identical to pass 26). Pass 26: 2026-11-22, master 7a6d1beb9 (Judgement of Wisdom pays no mana off an avoided white swing under Forever: paladin TPS about -0.3%, Warden -0.2%, warrior identical; warrior Deep Wounds, Execute and Windfury talent rules reach no tank preset). Pass 25: master 8b5137c6d (paladin Two-Handed Weapon Specialization on all Holy damage, Nov 19 hotfix; standards 41; warrior push 114097 talents; no tank preset moves: all 12 presets identical to pass 24). Pass 24: 2026-11-20, master e07e8fdb8 (item conditions and target mob family, Tenets of the Silver Hand's
Undead condition, Totem of Rage's Forever 2%, a crit-against-target mark; no tank preset wears an affected item; all 12
presets identical to pass 23). Pass 23: 2026-11-19, master e671e881e (the least-HPS
search in `sim/core/tank_score.go` restarts replays from the last capped point; all 12 presets identical to pass 22).
Pass 22: master 8f388c7f3 (a Swiftmend refactor in `sim/druid/healing.go`; all 12 presets identical to pass 21). Pass 21: 2026-11-18, master a6677362f (spell GCD
haste, standards 27, and the healer reaction-time wake fix: the Warden -16 / +1 / +30, within about 2 sigma, the others
identical; ranking unchanged). Pass 19, master 2f7e255af: a Shieldsworn Holy Strike no longer rolls Windfury Totem
(paladin -29 / -38 / -22, TPS -1%). Last
full change: pass 16, master 25ce81de7, after the Nov 13 hotfixes (PR #130: Windfury Totem rolls on every landed main hand melee hit, standards 17;
Judgement of Command melee again) and the warrior's Balanced and Max survival re-tune (f2a5dfe25). The report has the details
(orchestration/audit/reports/tank_fairness.md); the table is in orchestration/audit/ref/tank_fairness/summary.tsv.

## Ranking (tankbench 3000 iterations, seeds 1 and 2 mean, six-boss mean, each spec's matching preset, Phase 0 gear)

| weight | order (best first) | scores |
|---|---|---|
| Max threat (w 1/3) | Bear, Warrior, Paladin, Warden | 5603 / 4455 / 3913 / 3400 |
| Balanced (w 1) | Bear, Warrior, Paladin, Warden | 3696 / 3058 / 2952 / 2024 |
| Max survival (w 3) | Warrior, Bear, Paladin, Warden | 1079 / 1047 / 780 / -1401 |

On the synthetic sets (pass 33) it was 4537 / 4097 / 3282 / 2606, 3342 / 3251 / 2863 / 1244 and paladin 1875, warrior
1617, bear 1301, Warden -2277. The history below is from the synthetic sets.

The Nov 13 Windfury rule (every landed main hand hit rolls the totem) gave the bear +328 / +260 / +129 (12% at Max
threat, TPS +163) and the Warden +88 / +61 at Max threat / Balanced; the paladin +38 at Balanced; the warrior nothing
(its rolling strikes already rolled; Shield Slam is a shield hit). Measured by building tankbench at b02b98c72 and
d4325b775 (PR #130's merge), same requests. The bear is now 5% behind the paladin at Max threat (was 14%). Max survival
was a tie within the error until the Nov 16 hotfix took Windfury Totem off the Shieldsworn Holy Strike (the game's fix,
so fair): two-seed means now warrior 1830 (1836.7 / 1822.3), paladin 1781 (1775.8 ± 14.0 / 1785.7 ± 14.0), the warrior
ahead by 49 (2.7%), outside the error. The paladin's open potion action (finding -20261105-1, +374 at Max survival in
pass 13) would still put the paladin first. Pass 29 (master 566067a39): warrior 1814.7 ± 12.2 against paladin 1778.0 ± 15.9 (seed 1), the lead down to 37 (2%) after Unbridled Wrath left Shield Slam and Revenge. Pass 32 (master 0404484fb): 1813.5 against 1778.0, unchanged by the stance Rage rule. Pass 33 (master 85691655a): the warrior threat batch cost the warrior 197 there, so the paladin leads, 1874.5 against 1617.0 (16%), with or without its open potion.

TPS (Balanced, Phase 0): bear 2612, warrior 2137, paladin 1858, Warden 1738. Healing needed (Balanced): paladin 449, warrior 524,
Warden 700, bear 711.

Each preset wins its own weight except the bear's Balanced: its Max threat build beats it at w 1, 3889 against 3696 on Phase 0 gear (a tie on the synthetic sets since 2f42f4a44; finding -20261125-1). The three builds are distinct for every spec. Every flask still wins
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
- **Bear threat (pass 33):** wowsims/forever's ThreatLog readings (EV-0034, EV-0037) put Lacerate at 10/3 of its damage and Primal Bite at 3x, both before the form's 2.15 (their rank 5 ticks read 10/3 x 2.15 a stack, so the stacking is right). Lacerate went 248 -> 825 TPS on Magmadar; the bear now out-threatens the warrior by 19%. Sourced; taken as real. The Anniversary logs give damage taken, not threat, so they neither confirm nor contradict it.
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
- Racial actives: the warrior's APLs name every racial (Blood Fury, Berserking, Elune's Light, Eureka!, and Stoneform
  from 20 s at Balanced and Max survival, 0f567e196 to dfd015925); the other three tanks rely on autocast, which fires
  the damage racials but never Stoneform (a survival cooldown) in a scoring run. So the paladin and Warden Dwarf race
  rows miss Stoneform, about +10 Balanced / +40 Max survival by the warrior's sweep (finding -20261124-1). No preset
  race moves: every preset is Undead, Night Elf or Tauren, and the warrior presets' new lines are Undead no-ops.
- Raid-only consumables: settled 2026-11-12 (user ruling, option a): no tank uses a raid-only stone, oil or elixir.
- Enchants and the bear's Max survival race: settled 2026-11-12 (findings -20261112-2, -3). Every preset carries the
  best legal Phase 1 enchant per slot for its own weight (10,000 iterations, seeds 1 and 2, paired): paladin +75 / +53 /
  +76 (Holy Power gloves, Voracity Agility, a spell power neck at Max threat and Balanced), Warden +20 / +28 / +87,
  warrior +21 / +9 / +14, bear Max survival +19 (gloves) and +21 (Tauren).
- Checked and equal: the paladin has no Blessing of Wisdom (every other mana spec does), but Wisdom gains it nothing
  (-1 to -8, noise). Everything else matches: all four tanks have searched every slot of the full raid set
  (standards 40); the Warden's other mana consumables (Demonic Rune, Greater Intellect) are noise for the paladin.
  Since f1942793d the Warden's Elixir of Greater Intellect adds nothing under Arcane Brilliance (removing it: Balanced
  1234.2 both ways, seed 1); a dead slot, not an unequal term (consumables-20261125-2 owns the elixir picks).
- Rage from hits taken (b4cab8498): only the warrior's stance enters it (Defensive 0.9, measured in our logs); Bear
  Form has no such modifier in the sim. Spells, ticks and Physical specials give both rage tanks twice the Rage.

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
