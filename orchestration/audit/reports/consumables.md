# Consumables pass 38

Master 1ab3ec28a0, previous audited 753d81da74. Result: quiet, no findings.

- Moved file: ui/enhancement_shaman/presets.ts (20+, 44-). The diff removes the solo Level 30 leveling
  encounter and build (EncounterLeveling, PresetBuildLevel30Leveling) and adds EncounterLevelingPlan and
  PresetBuildLeveling10to60. Leveling presets only; no Level 60 Consumes, buffs, debuffs, enchants or gear touched.
- scan.sh on master: every section above "== counts" empty; counts as expected (2 minor-armor effects,
  MobTypeUnknown default, Salvation and Sunder on, totem set per spec as ruled).
- wago latest 1.60 build: 1.60.1.70450, unchanged.
- Class branches show only stale copies of other classes' files (gotcha 1); not audited.
- No open items.
