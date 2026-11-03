# Results and spec strength: pass 73, master f5c99ad018

Round 70753 phase 3. Client 1.60.1.70753. Shared run: 68 requests, 20k, seed 1 (/tmp/audit-work/runs/master). The six rogue "(vs Humanoid)" builds are gone (user ruling); warlock Affliction Wrack is new.

## Numbers
- Every request's DPS equals its class's round file final (checked by string match of each 180 s value; 3 differ by 0.1 from rounding: shaman imp-gw 1149.75, rogue combat-daggers 946.65, warrior arms 1097.75).
- 62 of 68 metas agree with the run within 0.15 (class workers set them). The six shaman metas still carry the 2026-11-02 values on purpose (shaman.md:172): run 970.1 / 952.7 / 1149.8 / 1159.2 / 1055.1 / 1111.3 against meta 768.5 / 753.2 / 879.3 / 883.5 / 795.8 / 834.8. Waiting on the report regeneration (open finding results-20261102-1, updated).
- Drift (regen.py): the reference was stale. Gone: the five rogue Humanoid requests. New: affliction_wrack (0 diffs). One new failure: rogue__subtlety, meta race "Night Elf" is not a proto name (finding results-20261103-1); with NightElf the request regenerates with 0 diffs. Reference files replaced with this run.
- unread_fields.py: nothing printed (even `notes` is no longer reported), so no field is dropped.

## Ranking at 180 s (20k, seed 1)
| Tier | Builds |
|---|---|
| 1300+ | Hunter Lone Wolf 1313.7 |
| 1200-1260 | MM 1254.3, Ret sub-Holy vs Undead 1221.4 (encounter preset) |
| 1100-1200 | BM 1195.6, Fury 2H 1169.9, Enhancement 1159.2, Fury 1152.6, Arms Dragonkin 1142.2, Feral 1131.1, Arms 1097.8, Ret twist 1090.9 |
| 1000-1100 | Ret sub-Holy 1089.8, Destruction 1063.4, Frostfire 1064.3, Fire 1057.9, Demonology 1053.9, Survival 1041.2, Ret Shock 1030.2, Affliction 1015.6 |
| under 1000 | Destruction fire 996.7, Arcane 962.4, Balance 951.5, Shadow 948.7, Elemental 970.1, Frost 944.7, Rogues 866-963, Wrath 861.5, Smite 762 |

Why: hunters rise 14-20% from the 2026-11-03 hotfix (Aimed/Arcane/Serpent/Sniper bases, Hawk 110, Trueshot 100, Barrage 15%; tier 1) plus the notes' Arcane 0.15 and Serpent 5% shares (tier 4); Survival only +1.6% (no Barrage, shots matter less). Enhancement +31% and Elemental +26% are the notes' Stormstrike 150%/6 s, Maelstrom x2, shocks at Era base, Lightning Bolt at 0.78 of Era and Overload 10-30% (all notes only, ledger in shaman.md). Feral +20% is Shred 200%, Claw 140%, Ravage 435%, bleeds and Bite x1.25, Rend and Tear 3% (notes only). Paladins +5% from Vengeance x5 and Two-Hander on Holy. Rogues +25-30% from the reworked tree (Opportunity 5 points, Aggression, Coup de Grace, Hemorrhage on Garrote, poison coefficients doubled, Mutilate 90%). Warlock only Destruction moved (+66, Shadow and Flame 4%/point). Priest, mage flat.

## Notes-only values
Checked against the notes text (sources/devnotes_post9_2026-11-02.txt): druid (Shred/Claw/Ravage/bleeds/Bite/Starfire/Rend and Tear), hunter shares, paladin (Vengeance 5, Two-Hander Holy), rogue (Shank, poisons 0.009/0.01, Mutilate, Flawless 15, Aggression), shaman, warlock (Wrack 0.286 and 86 base, S&F 4%), warrior: each is listed in its class file, cites the notes, and matches the text. Gated: druid, rogue poisons, warrior, paladin, shaman yes; hunter shares not (finding results-20261103-2). Mage has none.
Trees: rogue.json matches every row, rank and prerequisite of the notes. Warrior and shaman match the notes; the warrior columns (Enrage, Reckless Swings, Gore Drinker, Improved Whirlwind) are guesses, recorded in the entries' sources. Hemorrhage/Ghostly Strike dagger nerf has no numbers; sim keeps 145%/180% (documented).
Not checked: gear, Phase 3 re-sim effects (Salvation drop planned), Skyborne stats (open question for the user).
