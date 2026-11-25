# Consumables pass 106 (master 63129ce35, client 1.60.1.71091)

Result: quiet. No new findings. Open: consumables-20261110-1 and consumables-20261105-1 (both low, still open).

## Moved since eceb09019
| File | Change | Verdict |
|---|---|---|
| ui/shadow_priest/presets.ts | Level 60 talent string and tooltip only | no Consumes, RaidBuffs or Debuffs change |
| ui/rogue/gear_sets/level60_phase0_* | new Combat Daggers, Seal Fate Daggers/Swords, Alliance sets; weapon swaps | enchants checked below |
| ui/hunter/gear_sets/level60_phase0_* | new Lone Wolf sets, BM/MM weapon changes | enchants checked below |

No mksettings, consumes.go, buffs.go, debuffs.go, forever_enchants.go or assumptions file moved. Latest wago probe is still 71091.

## Enchants in the new sets
All ids resolve against the 71091 db2 or tools/database/forever_enchants.go: 1248599 bracer Superior Agility, 1249059 neck Agility, 1219587 cloak Agility, 1248640 gloves, 1887, 1891, 2543, 1508, 1900, 2564, 664, 34, 931, 803.
Each has a source in tools/rotopt/phase0_enchants.json, except gloves Superior Agility (spell 25080) behind 2564. The sets use 2564 on an off-hand weapon, which is the Enchant Weapon - Agility (spell 23800, Timbermaw rep, tier not in data) variant and is listed. Not new: it was already on the Subtlety and Survival sets.
The pass doc's note that bracer Superior Agility (1248599) is unused is stale: master's Phase 0 sets and req files use it, and the db and test carry it. Cosmetic doc nit, not raised.

## No open items beyond the two lows above.
