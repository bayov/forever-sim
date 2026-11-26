# Consumables pass 110, master 85691655a

Findings: consumables-20261125-3 (medium): report requests keep the Dark Bomb that EV-0088 makes useless on bosses.

Verified: client 71091 unchanged (no newer build). scan.sh: banned list and NONSYN lines unchanged in kind (gear-set enchants 2544/Revelation are Phase 0 sets owned by the gear pass); minor-armor count 3 by design (see 20261125-2). consumes.go stunExplosives and swing-keep, buffs.go Windfury extra-attack rolls, debuffs.go JoW/Gift of Arthas, sim/common proc changes: no consumable or enchant value change. Hunter (Dense Dynamite) and feral (Land Mine) presets.ts match their mksettings; enhancement keeps Goblin Land Mine, marked due a re-check by its owner.
Open: 20261125-3, -2, -1, 20261110-1, 20261105-1.
