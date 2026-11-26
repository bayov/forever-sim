# Consumables pass 111 (master bfbee78fc, client 1.60.1.71139)

No new findings. Open items from pass 110 (20261125-1, -2, -3, 20261110-1, 20261105-1) stand unchanged.

## What moved
- Master since 85691655a: only Phase 0 gear sets (44 files, the re-rank). No change to presets.ts, mksettings, sim/core, sim/common, forever_enchants.go, assumptions_def.py or the mage imbue.
- Client 71091 -> 71139 (builddiff.py): 2 consumable spells changed, both world buffs no preset uses: Fengus' Ferocity (22817, effects 2/3 value 230 -> 115) and Sayge's Dark Fortune of Damage (23768, effect 0 now a flat aura 4 instead of 79). Forever enchants: 58 checked, 0 changed.

## Verified
- scan.sh: banned list only lists Phase 0 gear-set enchants 2544 and Revelation 1248805 (gear pass owns legality; 1248805 is in phase0_enchants.json). Revelation-outside-shamans, trinket, crit-elixir lists empty. Minor-armor count 3 by design (see 20261125-2). Air/weapon totems and Sunder/MobType lines unchanged.
- New Phase 0 copies (Horde/Alliance sets, Ret Shock, SV hunter, rogue Seal Fate) use enchants all present in phase0_enchants.json, except 2544, 2543, 2564, 1891, 1887, which the gear pass already carried.

No open items beyond the above.
