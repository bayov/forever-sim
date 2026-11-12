# Tank and healer audit, 2026-11-11/12

User's ask: "fix findings and audit tanking and healing again to ensure we use fair, and good presets for each so we judge each spec at its best".

Published: master 29a729ee3 (data and Changes lines), then 5a06b7964 (state). Deployed to https://forever-dps.pages.dev (deploy 24284d79). Full suite green; B8 passed at phone and desktop width in both themes.

## Open findings at the start, and their fixes

| Finding | Fix |
|---|---|
| healer_rotations-20261111-1, healer_fairness-20261111-1 | Restoration druid holds Nature's Swiftness for an instant Regrowth below 45% (raid) or an instant Healing Touch 11 below 45% (tank, was 60%). Heart of the Wild 2 replaces Subtlety 2 (`05302001--5253005153113051`). |
| tank_fairness-20261111-1 | The best flask per preset on its own objective (standards 11). Titans at Max survival for the warrior, bear and Warden, and at Warden Balanced. Supreme Power for paladin Max survival. Paladin takes Wizard Oil; warrior takes Dense Sharpening Stone. |
| tank_rotations-20261111-1 | Warden re-tuned after the Nov 10 shaman notes. Max threat: talents `05350002-050032313251012201-05`, Searing Totem and Fire Nova all fight. Balanced and Max survival: Healing Stream Totem. All three: Greater Mageblood Elixir instead of Mageblood Potion, plus a new Balanced gear set. |
| tank_mechanics-20261111-1 | Stale values in docs/tank_shaman.md corrected. |
| healer_mechanics-20261111-13 | Ironweave golden regenerated after the seen-health fix (heal_targets.go: APL health values now read the health the healer has seen). |

## Audit passes

| Pass | Commit | What it found, and the fix |
|---|---|---|
| tank_mechanics 12 | bb67d825d | 2 findings. Judgements are ranged under Forever: they now use the ranged hit table and ranged special proc mask (sim/paladin, Ret -0.47 to -0.70%). The second was the DPS-side mechanics question below. |
| tank_gear 7 | 3123733be | 3 enchant findings. Warrior gets Greater Agility boots and Superior Strength wrist; Warden gets an Agility cloak and Superior Deflection wrist at Max threat. |
| tank_rotations 10 | c9d959115 | 4 findings. Bear Max survival: Frenzied Regeneration and Swipe filler at 80 Rage, Maul at 40, new talents. Bear Max threat: Barkskin at 20 s. Warrior Max survival: Shield Wall in the last 13 s. |
| tank_fairness 10 | 0a3f60000 | 1 finding (flask and consumable parity), folded into the flask fix above. |
| tank_fairness 11 | a2d1dceaa | 2 findings. -2: better head, leg, neck, glove, wrist and chest enchants for every tank, each confirmed on two seeds at 10k iterations. -3: Tauren beats Night Elf for bear Max survival. |
| healer_mechanics 11 | 3fc5898c9 | 1 finding: the seen-health fix and the Ironweave golden. |
| healer_gear 6 | fa4ef6bcc | 3 findings. Consumable parity: every healer takes Greater Cleric's Elixir. Food is Wicked Smoothie (druid) or Runn Tum Tuber Surprise. Oil is Wizard Oil, or Lesser Mana Oil for the paladin raid and Holy tank presets (Brilliant oils are raid-only). Table in docs/healing_sim.md. |
| healer_rotations 9 | 8f3f74f3f | 2 findings. Shaman raid holds Nature's Swiftness for a Healing Wave 10 below 65%, and the tank preset drops its Lesser Healing Wave 5 line. The Discipline priest running out of mana on Chromaggus stays open (no variant helps). |
| healer_fairness 9 | 25f65da9c | 1 new finding plus an older one. Holy priest raid preset takes Twilight Focus 3 instead of Inspiration (+0.021). The Holy paladin raid preset becomes a Flash of Light healer (+0.03). A head, leg, glove and chest enchant screen moved no healer more than 0.006, so nothing was adopted. |

A second full round was not run: the parent asked to keep the remaining passes short because of cost. The last fairness passes on both sides came back with only the findings fixed above.

## Windfury Totem

Our client-flag implementation (88bf7f732) was dropped for the user's own a0b4d3be9 on master; see windfury_rule.md. The two disagreements:
- Mutilate MH 1310706/1310707: no effect. The sim's level 60 Mutilate is 1241584.
- Extra swings rolling the totem: master lets them roll. Blocking them would cost Combat Swords 2.75 to 2.92 DPS and Subtlety 0.1. Our old block never actually worked, so the published numbers already match master.

Every re-tune was re-checked under master's rule and still wins: Fury duty +5 to +7, Survival Wing Clip +1 to +4, Combat Swords over Daggers +7 to +15, and every tank pick.

## Scores and ranks

Tanks: six-boss mean on each preset's own weight. The "old" column is from the start of the round (2026-11-11).

| Preset | Old | New | Old rank | New rank |
|---|---|---|---|---|
| Warrior Max threat | 4202 (by pass 11) | 4235 | 1 | 1 |
| Warrior Balanced | 3426 (by pass 11) | 3421 | 1 | 1 |
| Warrior Max survival | 1561.5 | 1776 | 2 | 2 |
| Paladin Max threat | 3168.8 | 3267 | 2 | 2 |
| Paladin Balanced | 2755.3 | 2838 | 2 | 2 |
| Paladin Max survival | 1564.8 | 1811 | 1 | 1 |
| Bear Max threat | 2789 (by pass 11) | 2782 | 3 | 3 |
| Bear Balanced | 1865 (by pass 11) | 1856 | 3 | 3 |
| Bear Max survival | -594.3 | 262 (Tauren) | 3 | 3 |
| Warden Max threat | 2495.2 | 2562 | 4 | 4 |
| Warden Balanced | 1133.3 | 1224 | 4 | 4 |
| Warden Max survival | -2738.0 | -2337 | 4 | 4 |

The warrior briefly passed the paladin at Max survival mid-round. The paladin's new enchants put it back in front: 1811 to 1776.

Healers: mean carry. The tank healers rank against each other, and the raid healers against each other.

| Preset | Old | New | Old rank | New rank |
|---|---|---|---|---|
| Holy priest, tank | 1.278 | 1.278 | 1 | 1 |
| Discipline, tank | 1.272 | 1.272 | 2 | 2 |
| Druid, tank | 1.235 | 1.269 | 4 | 3 |
| Shaman, tank | 1.257 | 1.257 | 3 | 4 |
| Paladin, tank | 1.152 | 1.152 | 5 | 5 |
| Holy priest, raid | 1.440 | 1.461 | 1 | 1 |
| Paladin, raid | 1.400 | 1.430 | 3 | 2 |
| Druid, raid | 1.391 | 1.414 | 4 | 3 |
| Shaman, raid | 1.412 | 1.412 | 2 | 4 |

The druid, Holy priest and paladin raid rows were re-measured after the changes. Shaman, Discipline and the Holy priest tank row come from the pass-9 bench.

## DPS builds re-simmed

- Survival: 1013.6 to 1014.5. The Tallstrider and Wolf variants each gain 0.9.
- Fury on Battle Shout duty: 1043.7 to 1048.8.
- Ret paladin, all 10 builds: -0.47 to -0.70% from ranged Judgement (level_60: 1009.3 to 1004.2).
- Rogue: the default is back to Combat Swords. The DPS numbers didn't change.

## Needs the user

1. Paladin tank potion (tank_fairness-20261105-1): +58 / +172 / +402 if allowed.
2. Warden races (tank_fairness-20261105-2): the current race is illegal for shamans. A legal race costs -63 / -79 / -158.
3. Paladin tank damage taken (-20261104-2): waits on calls (a) to (c).
4. The legacy warrior_protection.json (-20261104-5).
5. Discipline priest goes out of mana on Chromaggus (healer_rotations -2). No variant improves the objective.
6. Paladin Bulwark at 10 s or 13 s (older ruling).
7. Healer gloves have no enchant, because a kit gained nothing in the sim. Add one for completeness?
8. Windfury extra swings: master lets them roll. Blocking them would cost Combat Swords about 2.8 DPS. Which is right in game?
9. DPS-side mechanics questions Mechanics-20261112-1/-2: Holy Strike and Unsealed Wrath.
10. The Blizzard news site returned errors, so posts after Nov 10 weren't checked.
11. Auditors created stray root directories (`/dummy`, `/tank_gear`, `/tank_fairness`), and their deletion was blocked here. Please remove them.
