# Consumables pass 89 (master 5f20dfdf4, client 70959)

No new findings. Open items: the two low nits already in findings (consumables-20261110-1, consumables-20261105-1), unchanged.

## Moved files
| File | Change | Verdict |
|---|---|---|
| sim/common/item_effects.go | Manual Crowd Pummeler haste CD 30 s -> 3 min (Forever tooltip, 3 charges uncounted) | Item is in no preset or level60 gear set (rotopt excludes it); no effect |
| tools/dps_report/assumptions_def.py | Paladin SoC Windfury, Unsealed Wrath, Warrior UW and Reckless Swings wording | Rage and proc-rule text, not a buff/consumable value |

## Client 70938 -> 70959 (builddiff.py)
- Consumable spell changes: 1, Maelstrom Weapon (408498) effect 1 value 50 -> 2. A shaman class aura, not a consumable; no preset consumes touched. Class auditors' area.
- Forever enchants: 58 checked, 0 changed.

## scan.sh on master
Banned, Revelation, trinket and crit-elixir sections empty. Totem counts as the ruling. Minor-armor exclusive count is 3 (expect 2): the third is the Dust Cloud pet aura (debuffs.go:1167, client 70532), 505 armor, same slot; benign, scan expectation is stale.

Class branches listed "ahead" carry stale copies (gotcha 1); not audited.
