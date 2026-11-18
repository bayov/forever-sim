# tank_rotations pass 17

Master 2f7e255af (2026-11-18), against pass 16's baseline (master 5f20dfdf4). 1,000 iterations, default seed, every
tank request on all seven boss files. Quiet: no new finding.

## What changed

Since 5f20dfdf4 (gate: sim/druid/healing.go, sim/druid/items_restoration.go, sim/paladin/holy_strike.go,
sim/shaman/items.go):

- **Blizzard's Nov 16 hotfixes** (4e81669a9): a Shieldsworn Holy Strike no longer rolls Windfury Totem
  (`core.SpellFlagNoWindfuryTotem` when the paladin has Shieldsworn; sim/paladin/holy_strike.go). All three paladin
  presets take Shieldsworn and the totem.
- Idol of Rejuvenation (Regrowth +5% on our Rejuvenation) and Totem of Tides (Riptide -5% mana): healer items, no
  tank request wears them; bear and Warden rows bit-identical.

No tank APL, tankbench request, rotopt tank template, tools/encounters, tools/tankbench/main.go or tools/tankscore change.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for 2f7e255af. 80 of 104 rows bit-identical: warrior, bear and Warden. Paladin
scoring mean (six bosses without Ragnaros) at the preset's weight:

| request | score old -> new | TPS old -> new | healing needed old -> new |
|---|---|---|---|
| paladin_protection_balanced | 2836.5 -> 2809.3 | 1530.8 -> 1520.5 | 241.3 -> 243.8 |
| paladin_protection_max_survival | 1833.5 -> 1780.2 | 1463.8 -> 1449.0 | 199.0 -> 201.5 |
| paladin_protection_max_threat | 3272.8 -> 3248.8 | 1557.3 -> 1546.0 | 273.2 -> 272.3 |

Cause: Holy Strike's lost totem procs, TPS -0.7 to -1.0% on every boss (unpaired 1k rows; Max survival's larger
score move is mostly Broodlord 484 -> 344 ± 104, noise at w = 3). A ruled hotfix, in the expected direction; no finding.

## Rotation lines

Only Holy Strike's value moved, so this pass tested whether it should still sit above Consecration. Paired `abtest`,
1k seed 1, own weight, scoring-mean diff (B minus preset), `mk_paladin.py ... cons_before_hstrike`:

| request | change | scoring mean | Ragnaros |
|---|---|---|---|
| balanced | Consecration before Holy Strike | -109.0 ± 12.6 | -716 |
| max_threat | Consecration before Holy Strike | -101.1 ± 6.9 | -281 |
| max_survival | Consecration before Holy Strike | -136.7 ± 30.5 | -377 |

Holy Strike on cooldown stays clearly right (Iron Creed and Shieldsworn carry it, not the totem). The Windfury
main-hand timing item is moot for Holy Strike now that a Shieldsworn Holy Strike can't roll the totem.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call, unchanged.
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling, unchanged.

## Notes for the orchestrator

- The pass doc's open item (main-hand timing for Windfury) was done for all four tanks in pass 15; it can still be
  removed from passes/tank_rotations.md.
- As in pass 16, the first `go build ./tools/tankbench` failed on a missing `sim/core/proto` package and worked a
  moment later (another auditor generating protos at the same time).
