# Consumables pass 72, master 0facd9428 (since 6b480b060)

Findings: none new. No open items.

Verified:
- changes.sh: 164 files, nearly all new synthetic hit-step level60 gear sets; no enchant or trinket outside the rules (scan.sh trinkets section empty).
- presets.ts diffs (enhancement_shaman, feral_druid, mage, retribution_paladin, rogue, warlock, warrior): gear-set wiring, talent strings and comments only; no Consumes, RaidBuffs or Debuffs change. Tank, restoration, warden and holy presets out of scope.
- mksettings, sim/core, sim/common, sim/mage: no change.
- tools/database/forever_enchants.go: adds Bracer - Superior Deflection 1248665 (tank warrior only, out of scope; no DPS preset uses it).
- assumptions_def.py: only a Swiftmend text; no buff or consumable assumption.
- scan.sh on master: banned, Revelation, trinkets and crit-elixir sections empty; counts as before (minor-armor effects 3, Salvation absent).
- New client build 1.60.1.70790 (was 70755). builddiff.py 70755 to 70790: 5 consumable spells changed, all poisons (Instant VI, Deadly IV and V, Mind-numbing III, Wound IV): effect 1 (the "$m2 charges" count) rose, e.g. Instant VI 175 to 250, Deadly V 180 to 240. Sim does not model poison charges (poisons.go:30 TODO), so no DPS effect. Forever enchants: 54 checked, 0 changed.
- Class branches stale, not audited.
