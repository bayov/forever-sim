# Consumables pass 90 (master 2f7e255af82b)

Result: quiet, no new findings.

## Moved watched files (5f20dfdf4..2f7e255af)
| File | Change | Verdict |
|---|---|---|
| sim/common/item_effects.go | Cannibal's Cleaver and Draconian Aegis mob-type effects; Darkmoon Blue Dragon 3%, Heroism 158-238, Maelstrom comment (hotfix 114044) | Trinket/weapon item effects only; no preset or mksettings uses these items (git grep empty) |
| tools/dps_report/assumptions_def.py | warr_exec_wf2 reduced to single setting (Execute extra Windfury roll fixed Nov 16) | Not a consumable or buff value; consistent with the Windfury Totem main-hand ruling |
| ui/hunter/presets.ts | Comment and tooltip text only (BM weave) | No Consumes or buff change |

## Checks
- scan.sh origin/master: banned, Revelation, non-synthetic trinkets, multi-crit-elixir sections all empty.
- No change to consumes.go, buffs.go, debuffs.go, forever_enchants.go or mksettings in the range.
- wago latest build 1.60.1.70959, same as state; no new client.
- Class branches: only stale copies of other classes' files (gotcha 1); not audited further.

No open items beyond the two low findings already in findings/consumables.md.
