# Consumables pass 94, master 4be96efef

Result: quiet, no new findings.

- Moved watched file: sim/common/item_effects.go (Adaptive Combat Assistant gets CooldownPriorityLow and a comment; a trinket cooldown ordering change). No enchant, oil, consumable or buff value changed. No preset or level60 gear set changes in the range.
- scan.sh on master: banned, Revelation, non-synthetic trinket and crit-elixir sections empty. Counts as before (minor-armor exclusive count 3 because of Dust Cloud; the scan's expectation of 2 is stale, see pass 93).
- wago latest: 1.60.1.70959, same as last pass; no new build.
- Class branches are stale copies, not audited.
- Open items: consumables-20261110-1, consumables-20261105-1 (unchanged).
