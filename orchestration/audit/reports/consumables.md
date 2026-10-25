# Consumables pass 40 (master 54878fa3d, client 1.60.1.70532)

Result: quiet. No new findings, nothing open.

## Checked
- scan.sh on master: banned items, Revelation, non-synthetic trinkets, two-crit-elixir presets all empty. Elemental Sharpening Stone appears only in tank_warrior (out of scope).
- Minor-armor exclusive effects in debuffs.go: 3, not the 2 the pass doc expects. The third is the new ForeverDustCloudAura (hunter pet Dust Cloud, rank 5 1265904 = -505 armor, matches client 70532). Expected; the pass doc's count should become 3.
- New consumables on master (user ruling 2026-10-23, standards 10/11): Flasks of Natural Accuracy/Aggression/Precision (zone bonus on via LEVEL_60_ZONE and mksettings NaturalFlaskZone), Dragonfire and Caustic Smog Potions, EZ-Thro Dark Bomb. presets.ts and mksettings agree for every spec (flask, potion, explosive, imbue). Tooltips (tt cache) match the sim comments.
- Hunter's Mark rank 4 is 143 (user took the 2026-10-21 hotfix on 2026-10-23); debuffs.go, assumptions_def.py and the pass doc's settled list (+140) now differ only in the doc.
- New gear sets (enh burning totem, ret crusader duty/undead, rogue subtlety) carry the same enchants as their base sets; trinkets synthetic.
- Duty presets: Ret crusader duty switches off only Judgement of the Crusader; warrior duty only Battle Shout.

## Could not check
- Sources of the Natural flasks, Dragonfire, Caustic Smog and EZ-Thro Dark Bomb: wowhead returned nothing from this sandbox (empty body). Not verified as launch-obtainable; user rulings cover their use.

## Open items
None. Pass doc housekeeping for the next editor: Hunter's Mark +143, minor-armor count 3.
