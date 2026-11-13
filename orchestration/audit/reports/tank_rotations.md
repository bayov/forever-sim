# tank_rotations pass 13

Master 3a86d4a6b (2026-11-13), against pass 12's 97da1cca5 baseline (state last.master cad23ae1a). 1,000
iterations, default seed, every tank request on all seven boss files.

## What changed

Since cad23ae1a (gate: 1 watched file, sim/shaman/flametongue_weapon.go):

- 4c4790e3e: Flametongue Weapon rank 6 hits 5.5 less under Forever. The three Warden requests use
  `shamanImbue: RockbiterWeapon` (tools/tankbench/req/shaman_warden_*.json:97-98), so no tank preset casts
  Flametongue.
- Also in range, outside the gate: c7a99fa49 (Windfury Weapon attacks roll Windfury Totem; Wardens wear Rockbiter),
  d22fe3e23 / 041bb0a90 (Frightskull Shaft PPM and disease), 1743980e6 (Shadow / Frost Oil spell IDs), and
  tools/rotopt BiS work (bis.go, gear.go, run.go, search.go). None touches a tank APL, request, rotopt tank
  template, tools/encounters, tools/tankbench/main.go or tools/tankscore.

## Baseline

All 104 rows of ref/tank_bench/summary.tsv (13 requests x 7 bosses + scoring means) are bit-identical to pass 12;
only the header line was rewritten for this master. No finding.

## Rotation lines

No rotation changed and no engine input to the tanks moved, so pass 12's A/Bs stand; nothing re-tried.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call, unchanged.
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling, unchanged.

## Notes for the next pass

- The first `go build` of tankbench in this session failed with "no required module provides package
  .../sim/core/proto" and succeeded on retry a minute later, likely a concurrent pass regenerating the protos.
  Retry once before treating it as an error.
