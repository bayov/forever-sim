# Consumables pass 57 (master fb1485888)

No findings. Last audited master bbbaf2e3c; client 1.60.1.70707 is still the newest build. Class branches are not ahead in any audited way (stale copies only).

Moved watched files, all checked:
- sim/core/consumes.go: weapon imbues now need a fitting weapon (ImbueFitsWeapon, from the stones' SpellEquippedItems). A misfit imbue does nothing and warns.
- tools/rotopt/mksettings/main.go: fitStonesToWeapons swaps a Dense Sharpening Stone for a Dense Weightstone (same +8) on a blunt weapon and clears empty hands. Consistent with the rogue presets.
- ui/rogue/presets.ts: Combat Daggers and new Combat Swords consumes put a Dense Weightstone on the off-hand mace (994014 is a mace; Seal Fate's off hand 994017 is a dagger, so keeps the sharpening stone). Matches mksettings.
- sim/core/buffs.go: Windfury Totem proc ICD 100 ms under Forever (sim note says to verify in game); not a consumable value, outside this audit's number checks.
- tools/dps_report/assumptions_def.py: new mage_ice_lance_cd assumption; no buff content.
- ui/mage/presets.ts, ui/enhancement_shaman/presets.ts: talents, rotation notes, leveling plan; Consumes, buffs and gear untouched.

scan.sh on master: every ban section empty. Minor-armor exclusive-effect count is 3 (known stale expectation, Dust Cloud). Totem counts as the ruling.

No open items.
