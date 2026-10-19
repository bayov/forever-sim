# Consumables pass 36 (master e6d1a5ad9)

Moved: sim/core/consumes.go, ui/enhancement_shaman/presets.ts. No findings.

- consumes.go: `ShouldActivateNeedsEvent: true` on makeManaConsumableMCD. Flag exists (major_cooldown.go:63, apl_idle.go:253); no value change.
- enhancement_shaman presets.ts: leveling plan swapped for the searched route/rotations (imports, DefaultLevelingPlan only). Level 60 Consumes, raid buffs and gear untouched.
- scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty; counts as before (minor-armor effects 2).
- wago latest: still 1.60.1.70450. No class branch audited (own-commit review not needed beyond master).

No open items.
