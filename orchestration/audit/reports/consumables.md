# Consumables pass 80, master a0b4d3be9

Moved since 33db32b36 (pass 79): sim/core/buffs.go, assumptions_def.py, standards.md 17, hunter/rogue/warlock/warrior presets.ts.

- buffs.go: Windfury Totem now rolls on listed main hand abilities under Forever (rollsOn func). Rotation/mechanics, not a buff value; no consumable effect.
- assumptions_def.py: wf_totem_nodamage text only.
- presets.ts: class default switch (user ruling 2026-11-12): defaults, races, warlock summon. No Consumes, RaidBuffs or Debuffs object changed; no gear-set, mksettings, consumes.go, debuffs.go or enchant change.
- scan.sh: all sections empty. Counts as before except minor-armor exclusive effects = 3 (ref 2); debuffs.go unchanged in this range (third is a pet debuff, Dust Cloud; not re-verified this pass).
- Wago latest 1.60 build: 70888 (already audited). Class branches carry stale copies; not audited.

No open items beyond findings file.
