# Consumables pass 95 (master 813f2b1e9)

Trigger: sim/common/item_effects.go changed (4be96efef..813f2b1e9).

- Diff: Hand of Justice proc chance is now 3% against Dwarf-family targets, 1% otherwise (Forever), 2% on Era;
  adds ConditionVsDwarves (defined in sim/core/item_conditions.go). No consumable, buff, debuff or enchant value changed.
- Hand of Justice (11815) appears in no Level 60 presets.ts or level60 gear set, so the standing trinket check is unaffected.
- No other watched file moved on master. Class branches not re-scanned beyond changes.sh output (all stale copies, gotcha 1).

Result: no new findings. Two low findings stay open (enchant 41 name, feral mksettings Sapper).
