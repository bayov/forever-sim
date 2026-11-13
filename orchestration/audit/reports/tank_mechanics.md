# tank_mechanics pass 15

- master: 3a86d4a6bf60814bb28d3624c92072d9200337a7 (from cad23ae1a9)
- client: 1.60.1.70938 (unchanged since pass 14)
- result: quiet

## What changed

| Commit | File | Change | Tank impact |
|---|---|---|---|
| 4c4790e3e | sim/shaman/flametongue_weapon.go | FlametongueWeaponCut rank 6: 0 -> 5.5 (log fit, Nov 12-13, levels 55-60) | none |

No Warden preset imbues Flametongue Weapon: the Warden uses Rockbiter (docs/tank_shaman.md:43,
"Windfury and Flametongue Weapon for Rockbiter lose 1,500 to 1,700 at every weight"). Lowering
Flametongue's hit only widens that gap, so no preset or rotation choice moves. Tank setups
(ui/dps_report/data/tank_setup.json) use stones and Wizard Oil for the other tanks. The cut is a
DPS-side log fit (tier: own combat logs), owned by the mechanics pass.

## Not done

- Blizzard posts since dev_notes_seen not re-read this pass (no new client build).

## Open findings (unchanged)

- tank_mechanics-20261108-1 (low): no test pins Warden shock threat under the Stormstrike mark.
- tank_mechanics-20261107-1 (low): client 70862 threat labels in comments and docs.
