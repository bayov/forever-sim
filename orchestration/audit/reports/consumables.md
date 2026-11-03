# Consumables pass 66 (master a0b306984)

Trigger: 7 watched files moved since 796495d6f (standards.md, assumptions_def.py, spec_hunter.go, presets.ts of hunter, Enhancement shaman, Restoration shaman and warlock) and client 1.60.1.70707 -> 1.60.1.70753.

## Verified
- Hunter presets.ts and spec_hunter.go: the ranged hunters (BM, MM, Lone Wolf) lose the Goblin Sapper and keep the EZ-Thro Dark Bomb; Survival keeps the Sapper (presets.ts:333, mksettings line 113 conditional). This is the fix for consumables-20261101-1; both files agree. The finding was already resolved.
- Warlock presets.ts: Affliction race Human -> Orc and comments only. No Consumes change.
- Enhancement shaman presets.ts: leveling weights moved to leveling_weights.json. No consumes, buff or imbue change.
- Restoration shaman presets.ts: new healing spec (out of scope). It carries a Brilliant Mana Oil mainhand imbue, Flask of Supreme Power and other consumes; not a Level 60 DPS preset, so not audited.
- standards.md: one ruling note (Balance Subtlety 3). assumptions_def.py: Ice Lance text only.
- scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty. Counts as before (minor-armor count 3, known: Dust Cloud; ref expects 2).
- Class branches: scan lists them as "ahead" only through stale copies (gotcha 1); no own commits audited.
- Client 70707 -> 70753 (builddiff.py): forever_enchants.go 53 enchants checked, 0 changed. Consumable spell changes: Minor Speed 13889 8 -> 4, Improved Drains 403511 gains a second effect, new Speed 1328947 and Enchant Boots - Speed (item 289559, spell 1328948, enchant 8737). None used by a preset, mksettings or the sim (git grep).

## Open
- None.
