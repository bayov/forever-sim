# Consumables pass 78, master da272241e5

Moved: sim/common/item_effects.go (new The Chief's Enforcer 9477 weapon proc, 1.14 PPM, 294 physical, from the user's log),
sim/common/itemhelpers/weaponprocs.go (WeaponProc.NoSelfProc). Neither is a consumable, buff or enchant.
No preset, gear set or mksettings file references item 9477 (git grep).

Client 70862 -> 70888: builddiff.py reports 0 consumable spell changes, 0 of 56 Forever enchants changed.
scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty; totem counts as ruled.
Count note: minor-armor exclusive effects in debuffs.go is now 3 (expect 2): ForeverDustCloudAura (debuffs.go:1167)
shares the slot with Faerie Fire and CoR by design (client 70532). Not a finding; the expected count is stale.
Class branches show no useful ahead state (stale copies).

Open items unchanged: 4 open findings in findings/consumables.md (Natural flask halving is the high one, awaiting a ruling).
