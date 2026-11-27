# Fairness pass 66 (2026-11-27, master 5ab589d85)

Quiet pass (no new finding). Range 34994816e..5ab589d85 (15 watched files): Talisman of Ascendance loses a charge
to every APL spell, not only damage and healing spells (413f06cb0, sim/common/item_effects.go, from the classic
SpellAuraOptions proc mask); priest tank-healer APLs and lists, restoration druid list, rogue effects files
refreshed, Phase 0 content and enchant rows for healer items. No sim/core, class, request, database or standards
change, so no race sweep: the 2af970d74 sweep stays the baseline.

## Published against current

The loop's 20k run (/tmp/audit-work/runs/vs_published.txt) against meta `published.dps`: the warlock Talisman
wearers are 1.2 to 5.6 under (affliction-duty-recklessness -5.6, -duty-curse -5.2, demonology duties -4.5 / -4.2,
destruction duties -4.4 / -4.0, demonology -2.4). That is the Talisman fix: bc5f5d61c re-simmed those builds into
builds.json but left the metas. Already filed as results-20261127-1; confirmed independently here. Balance and
Elemental (also Talisman wearers) move 0.7 or less. Every other build is within 1.5, as at pass 65.

The fix is applied evenly: one rule (an APL spell that isn't a melee white hit) for every wearer. Warlocks pay
most because curses and Life Tap now eat charges.

## Gear sets against requests (tools/audit/fairness/gearset.py, now in the repo)

Paired 10k, seeds 1 / 1000003, 5ab589d85 CLI, request with only its equipment set to the current gear set:

| Build | Request | Gear set | Delta | Pass 65 |
|---|---|---|---|---|
| Affliction | 962.8 | 945.6 | -17.2 | -17.2 |
| Affliction duty-curse | 948.1 | 935.7 | -12.4 | -17.7 |
| Affliction Shadow Bolt | 818.4 | 810.4 | -8.0 | -10.2 |
| Demonology | 873.8 | 867.2 | -6.6 | -9.6 |
| Destruction | 866.6 | 866.8 | +0.2 | -1.5 |
| Balance | 801.8 | 796.2 | -5.6 | -5.7 |
| Elemental | 856.2 | 857.9 | +1.7 | +1.6 |
| Enhancement | 966.4 | 972.5 | +6.1 | +6.1 |

Added to fairness-20261127-1 (still open, high): Affliction still drops below Survival and Enhancement imp-gw;
Demonology drops below Enhancement mana-tide and ties Destruction.

## Open findings re-checked

- fairness-20261126-4 (Ret Maelstrom): phase0_check still flags it (1.87 above the pick). Open.
- fairness-20261125-2 (explosives): Fury, Mutilate, Ret still carry no fillerExplosive. Open.
- Race picks (Demonology Orc, MM Orc, Arcane Gnome, Fire, Shadow): unchanged requests. Open.
- fairness-20261126-3 (Shahram stun): item_effects.go unchanged there. Open.
