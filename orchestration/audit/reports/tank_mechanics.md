# Tank mechanics, pass 25

- Master: e6eaa9470 (from 813f2b1e9). Client: 1.60.1.70959 -> 1.60.1.71037 (wago latest, 2026-11-20).
- Result: quiet. Open findings unchanged (20261117-1 gets new evidence, 20261108-1, 20261107-1).

## What changed on master

No sim, tank doc or tankbench file changed: the 9 commits are rotopt/BiS (Enhancement phase 0), TASKS.md, log-audit findings and the audit loop's own commit. TASKS.md's new items (Shank push 114268, unplaced Nov 20 pushes, Onyxia Blood Talisman 18406 test) touch no tank preset: no tank gear set equips 18406.

## Client 71037 vs 70959 (+hf, `wago.py diff --set mechanics` and default `--by-spell`)

| spell | change | tank effect |
|---|---|---|
| 12322 Unbridled Wrath | SpellAuraOptions ProcTypeMask_0 4 -> 20 | push 114097 now in the bare build; sim already mask 20 (talents.go). Sunder question stays open (20261117-1) |
| 12966 Flurry, 1329555 Reckless Swings | ProcTypeMask_0 4 -> 20 | same push; Flurry consumption is client_update-20261117-2's (DPS); Reckless Swings test in tank_warrior unchanged |
| 8516/10608/10610 Windfury Totem | SpellCategories 0 -> 2619, category recovery 100 ms | totem cast only; the raid Windfury buff the tanks take is unchanged |
| Execute 5308/20658-20662 | Attributes_2 +0x100, Attributes_3 -> 0x70400 | no tank preset casts Execute; the DPS mechanics pass's |
| 1324298 Krol'dok Resolve | ProcCategoryRecovery 0 -> 1000 | "chance on block to heal", not in the sim or any tank preset |
| Heroism 23682/23689, Aura of the Blue Dragon, Vampiric Embrace, Sap, Mount Speed, Traveler's Sign, NPC spells | various | not tank mechanics |

No change to Defensive Stance, Bear/Dire Bear Form, Righteous Fury, Shield Block, Holy Shield, Revenge, Shield Slam, Sunder, Maul, Swipe, Stormstrike, Rockbiter, Earth Shock, any tank talent row, SpellShapeshiftForm or threat aura.

## Tests

Not rerun: no sim code changed since pass 24 (813f2b1e9), when all tank tests passed.

## Note for the next pass

`wago.py diff --set mechanics --by-spell` fails (KeyError 'SpellID' on a table without SpellID); use `--set mechanics` plain plus default `--by-spell`. `python3 -I` breaks wago.py's sibling `hotfixes` import, so run it without `-I`.
