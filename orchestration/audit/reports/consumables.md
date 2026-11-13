# Consumables pass 84 (master 3a86d4a6b)

Moved since cad23ae1a: sim/common/item_effects.go, sim/core/buffs.go, sim/core/consumes.go (4 commits). No preset, mksettings, gear set or enchant file moved. wago latest 1.60 build is still 70938 (already audited).

| Commit | Change | Verdict |
|---|---|---|
| 1743980e6 | Forever Shadow Oil / Frost Oil proc spells keyed 350026 / 350025 (ID only, damage ranges untouched) | no value change |
| d22fe3e23, 041bb0a90 | Frightskull Shaft 3.9 PPM, Weakening Disease 15 x 7 ticks under Forever (log-fitted, client base 15, 14 s) | no preset or gear set uses it; nothing to flag |
| c7a99fa49 | Windfury Weapon attacks (8232..16362) roll Windfury Totem under Forever; MH special only | consistent with the main-hand-only ruling; DPS rulings for shaman/Enh belong to class audits |

- scan.sh: banned, Revelation, trinket and crit-elixir sections empty. Totem counts match the rulings. Minor-armor count 3 vs 2 is ForeverDustCloudAura (known stale expectation).
- Class branches: stale copies, none audited (gotcha 1).
- No open items beyond the two low findings already in findings/consumables.md (not re-checked: both files unchanged).
