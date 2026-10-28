# Consumables report, pass 50 (master 224b8a4e0)

Quiet: no findings.

- Master 5d27b0015 -> 224b8a4e0: no change to presets.ts, level60 gear sets, mksettings, sim/core consumes/buffs/debuffs, sim/common, forever_enchants.go, assumptions_def.py or the mage staff imbue (git diff checked).
- Client 70634 -> 70658 (consumables set): builddiff.py shows 0 consumable spell changes, 53 Forever enchants checked, 0 changed.
- scan.sh on master: banned items, Revelation outside shamans, non-synthetic trinkets and multi-crit-elixir lists all empty. Raid totems unchanged (melee: Grace of Air + Windfury; Enhancement: Grace of Air + Flametongue).
- MinorArmorReduction count is 3 (third is the hunter pet Dust Cloud, as noted last pass); not a finding.
- Class branches: changes.sh reports them with thousands of commits "ahead" (unrelated history); not audited, same as earlier passes.

No open items.
