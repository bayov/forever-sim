# Consumables pass 93, master e07e8fdb8

Result: quiet, no new findings.

- Moved watched file: sim/common/item_effects.go (weapon proc rulesets: Brain Hacker, Gravestone War Axe, Seeping Willow, Argent Avenger undead condition, Treant's Bane etc.; new item ids). No enchant, oil, consumable or buff value changed.
- No preset or level60 gear set references the touched items (git grep on presets.ts and level60 gear: none).
- scan.sh origin/master: all banned/Revelation/trinket/crit-elixir sections empty. Counts as before; the minor-armor exclusive count reads 3 because Dust Cloud (ForeverDustCloudAura) registers the category; debuffs.go unchanged in this range, FF and CoR still share it. scan.sh expectation (2) is stale.
- wago latest: 1.60.1.70959, same as last pass; no new build.
- Class branches are behind master (stale copies), not audited.
- Open items: consumables-20261110-1, consumables-20261105-1 (unchanged).
