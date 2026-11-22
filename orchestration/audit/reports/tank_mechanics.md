# Tank mechanics, pass 26

- Master: 8b5137c6d (from b4b68c52e). Client: 1.60.1.71037 (unchanged since pass 25).
- Result: quiet. Open findings unchanged (20261117-1, 20261108-1, 20261107-1; the last gets new evidence).

## What changed on master (10 watched files)

| file / change | source | tank effect |
|---|---|---|
| paladin talents.go: Two-Handed Weapon Specialization's Holy part now on the Holy school (all Holy damage) under Forever; spells' own multiplier 1; item swap reapplies damage and threat cut | Blizzard Nov 19 hotfix, t/2403040 post 13 (tier 1 notes); retires the 2026-11-15 ruling (standards 41) | none: prot paladin presets hold a shield; `TestProtection.results` unchanged |
| consecration.go, judgement.go, seal_of_fury.go, soc.go, sor.go | follow the above (comments, Consecration multiplier 1) | none (no 2H) |
| warrior talents.go: Flurry not spent by Heroic Strike / Cleave under Forever | Nov 16 hotfix, t/2403040 post 11 (tier 1 notes) | none: tank warrior talents `35--251533120331221351` take no Fury, no Flurry |
| shaman stormstrike.go, talents.go: comments cite client 70959 for Stormstrike 100% / 8 s and Maelstrom PPM | client 70959 (tier 1) | values unchanged (100%, 8 s, as before); Warden takes no Maelstrom |
| shaman item_sets_pve.go: Dungeon Set 1 in set 1679 (comment) | reveal 2026-11-17 | no Warden preset equips it |

Read in full: `applyWeaponSpecialization` (reset effect restores the start multipliers before the swap-back, so no drift) and the Flurry consumption trigger.

## Tests

`go test -tags=with_db ./sim/paladin/... ./sim/warrior/... ./sim/shaman/warden/...`: all pass (protection, tank_warrior, warden included). Without `-tags=with_db` they panic on missing items, which is expected.

## Findings

No new finding. 20261107-1 gets more evidence: docs/tank_shaman.md:430 and :464 still call Stormstrike 100% / 8 s and Maelstrom PPM notes-only, though client 70959 carries them (the code comments were relabelled this round).
