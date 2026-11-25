# Consumables pass 109, master 0404484fb

Findings: consumables-20261125-2 (medium): Intellect/Spirit elixirs overwritten by AB/DS under Forever (estimate, unsimmed).

Verified: item_effects.go (Puncture Armor 100, Armor Shatter into the minor armor slot; no preset impact), buffs.go (world buffs off in raid instances; no preset or mksettings enables any world buff), consumes.go (Elixir of Cunning Intellect moved to the elixir pass; overwrite rule), assumptions_def.py (warrior rage text only), warrior presets.ts (comment only). scan.sh: banned list unchanged (phase0 gear sets, gear pass owns); minor-armor count 3 by design. wago latest still 1.60.1.71091. Class branches were not ahead in a meaningful way (stale copies).
Open: consumables-20261125-1, 20261110-1, 20261105-1, 20261125-2.
