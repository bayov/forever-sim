# Gear pass 56 (master 6b480b060, client 1.60.1.70755)

Trigger: two new healer sets (Holy paladin, healing priest) and standards 40 / gear_targets healer rows.
lin2.py and std.py do not know HealingPower (lin2 prints total 0, std.py raises KeyError), so I added
`tools/audit/gear/healers.py` (weights 1b: SP 0.86, healing 0.45, MP5 2.4, crit 1; healer caster f).

| Set | Pieces (lin / 1.7-norm) | Weapons | Stat shares vs gear_targets row | Result |
|---|---|---|---|---|
| Holy paladin 990301-314 | all within +-0.4% lin, norm -2.8..+1.7% | MH 1.819x L(71), OH 0.528x (1.82 / 0.53) | Sta 14.9/14.7, Int 18.8/18.7, Spi 10.0, SP 21.7/21.9, Heal 26.9/26.8, Crit 2.8, MP5 5.1 | clean |
| Healing priest 990501-514 | all within +-0.5% lin, norm -3.2..-0.7% | staff 2.217x L(76), wand 0.261x | Sta 12.1, Int 15.3, Spi 10.1, SP 27.1/27.5, Heal 21.3/21.1, Crit 5.8/5.9, MP5 8.3/8.1 | clean |

- Item levels 66 / 71 / 76 correct. Trinkets are the Ret (992014/5) and Shadow (995033/4) ones; trinkets.py: no `!`.
- Enchants: all Forever/Phase 1 allowed (Voracity arcanum, Healing Power neck/bracer, Spellblasting, Mighty Intellect, Off-Hand Wisdom, Rugged kits). Cloak 1888 matches the Resto/Elemental shaman sets.
- Paladin carries Justice Armor 2-piece on head+chest (standards 6). Priest and Resto shaman healer sets carry no set bonus (Vestments of Prophecy starts at 3 pieces, so no 2-piece exists); consistent.
- Real items: Libram of Economy 272436 (phase 1, Stranglethorn drop) fine.
- Standards 40 (Enhancement imbues) is not a gear rule. No other gear files changed (blank.gear.json is empty).
- Not checked: wepcheck / std.py on the other 23 sets (no changes to them since pass 55).
