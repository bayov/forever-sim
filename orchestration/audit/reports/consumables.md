# Consumables pass 79, master 707ae92cc5

Moved since 27d0d59d6 (pass 78): sim/core/consumes.go, buffs.go, totem_weapon_buffs.go, sim/common/item_effects.go,
forever_enchants.go (+ json), assumptions_def.py, spec_enh.go, standards.md 25, elemental/enhancement/warlock presets.ts.

Result: no new findings.

| Change | Check | Verdict |
|---|---|---|
| Natural flasks 30 Stamina, 2% zone bonus (032692409) | code and test match client 70862 values | finding 20261107-1 resolved |
| Windfury Weapon + Totem and Flametongue Weapon + Totem now stack (Nov 10 hotfix) | imbue-blocks-totem code removed; standards 25 amended; Enh presets keep no totemWeaponBuff, other melee keep Windfury | consistent, ruling unchanged |
| Lesser Spirit bracer 7859 / boots 13687 now +4 Spirit; new enchant 1314331 | no preset or level60 gear set or mksettings uses them | nothing to do |
| Bladewhisper's Namesake 287072 (+8 AP vs Humanoids) | not in any preset or gear set | nothing to do |
| Enh, Ele, Warlock presets.ts | talents, rotations and comments only; consumes and raid objects untouched | nothing to do |
| spec_enh.go | comment only | nothing to do |

scan.sh on master: banned, Revelation, non-synthetic trinket and crit-elixir sections empty; totem counts as ruled.
Minor-armor exclusive count is 3 (ForeverDustCloudAura shares the slot by design), the expected count of 2 is stale.
Client: latest wago build is 1.60.1.70888, already audited in pass 78.

Open: consumables-20261110-1 (enchant 41 display name, low), consumables-20261105-1 (feral mksettings lacks Goblin Sapper, low).
Both re-checked still present on master.
