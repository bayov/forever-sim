# Mechanics pass 69 (2026-11-07, master 12aa2796f)

One new finding: mechanics-20261107-1 (high), the natural flasks on client 70862.

## Range

c67d9e014..12aa2796f: two commits (an audit-state commit and an item-database fold-in). No commit on sim or mksettings.

## Getting current

| Source | Result |
|---|---|
| Client build | 1.60.1.70790 -> 1.60.1.70862. `wago.py diff --by-spell`: 441 lines, 24 changed rows apart from new or NPC spells. |
| Dev notes thread | Still 10 posts; post 10 2026-11-04T19:35:16Z, post 9 edited 19:35:27Z, post 1 03:20:46Z. Nothing new. |
| Blue tracker | Newest: Linxy 2026-11-07T01:07Z (Midnight hotfixes) and 11-06 (Venomous Abyss raid tuning): retail, not Forever. Aggrend 11-05 (zeppelins). |
| Daily update | No TASKS.md, ledger or forever-update merge since the last pass. |

## 70862 shared rows

| Spell | Change | Sim |
|---|---|---|
| Flasks of Natural Accuracy / Aggression / Precision / Swiftness (1293740-1293743) | Stamina 60 -> 30; zone bonus 5 / 4 / 5 / 5 -> 2 / 2 / 2 / 2 | Still 60 and 5/4/5/5: mechanics-20261107-1 |
| Defensive Stance Passive (7376) #2 | 30 -> 80 (80% threat) | Matches (x1.8) |
| Bear Form (Passive2) (21178) #0 | 50 -> 115 | Matches (x2.15) |
| Command, warlock (20575) #1 aura 108 mask 0x3000 | 5 -> 0 | Not modelled; now 0. #0 pet damage 5% holds |
| SpellItemEnchantment | adds 8726 Luminarian's Binding (dummy 1314330) | n/a |
| ItemEffect | 231823 Makeshift Stickin' Spear cooldown 5 -> 15 s; 232058 5 s | Not in the sim |

Everything else: NPC spells (Hateful Strike, Cleave, Shadow Bolt Volley, Meathook, Righteous Fire, Mangle!), new spells, Use Soulstone aura options removed, The Quick and the Dead attributes, NPC Kidney Shot / Gouge levels. CurvePoint: 36 diff lines, left to the talents auditor.

## Natural flask effect (measured)

Shared CLI 12aa2796fcc6, 20k iterations, seed 1, paired against `runs/master`; requests given bonusStats of −30 Stamina and −3 hit / −2 crit / −3 expertise (`$AUDIT_WORK/mechanics/flask/mk.py`).

| Preset (flask) | Master | 70862 values | Change |
|---|---|---|---|
| rogue combat-daggers (Accuracy) | 954.5 | 911.4 | −4.5% |
| warrior fury_2h (Precision) | 1150.8 | 1103.6 | −4.1% |
| warrior fury (Precision) | 1137.1 | 1092.5 | −3.9% |
| warrior arms (Aggression) | 1088.3 | 1049.6 | −3.6% |
| hunter marksmanship (Aggression) | 1254.3 | 1210.3 | −3.5% |
| hunter beast-mastery (Aggression) | 1195.6 | 1153.8 | −3.5% |
| hunter lone-wolf | 1313.7 | 1268.2 | −3.5% |
| hunter survival (Precision) | 1040.6 | 1007.0 | −3.2% |
| druid feral-cat (Aggression) | 1169.5 | 1133.5 | −3.1% |
| rogue subtlety / seal-fate / mutilate / combat-swords (Precision) | | | −1.9% to −2.2% |

Casters, Balance, Elemental, Retribution and Enhancement drink Flask of Supreme Power and don't move. The physical specs fall 2 to 4.5% against the casters, so the ranking moves.

## Conclusions

Fix the natural flasks (mechanics-20261107-1), then re-search each preset's flask. Every other shared verdict holds on 70862. Open item from pass 67 (the 70790 hotfix rows fetch) is still open; 70862 bakes in whatever of them landed.
