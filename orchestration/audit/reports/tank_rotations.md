# tank_rotations pass 16

Master 5f20dfdf4 (2026-11-17), against pass 15's baseline (master 25ce81de7 / last audited 1a85df1d7). 1,000
iterations, default seed, every tank request on all seven boss files. Quiet: no new finding.

## What changed

Since 1a85df1d7 (gate: sim/shaman/item_sets_pve.go, sim/warrior/execute.go, reckless_swings.go, talents.go):

- **Server hotfix push 114097 (Nov 17)**: Unbridled Wrath's ProcTypeMask 4 -> 20, so under Forever a landed damaging
  melee ability procs it too (Whirlwind excepted, from wowsims/forever's EV-0138 log) (sim/warrior/talents.go). The
  three scoring Prot warrior presets take 5/5 Unbridled Wrath (`05-05050005-...`); the legacy `warrior_protection`
  request doesn't.
- Reckless Swings: the same push's mask; a Cleave uses one charge a cast (no tank uses it).
- Execute: under Forever a missed/avoided Execute keeps its extra Rage (no tank rotation casts Execute).
- Shaman item set bonuses (sim/shaman/item_sets_pve.go, Five Thunders etc.): no Warden request wears them; Warden rows
  bit-identical.
- rotopt gear-set search (tools/rotopt/gear*.go), core helpers `ItemSetOf`, `Pet.InheritedStats`: no sim effect.

No tank APL, tankbench request, rotopt tank template, tools/encounters, tools/tankbench/main.go or tools/tankscore change.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for 5f20dfdf4. 80 of 104 rows bit-identical: bear, paladin, Warden and the
legacy warrior request. Scoring mean (six bosses without Ragnaros) at the preset's weight:

| request | score old -> new | SafeDPS old -> new | TPS old -> new | healing needed old -> new |
|---|---|---|---|---|
| warrior_protection_balanced | 3443.2 -> 3470.2 | 2211.8 -> 2221.3 | 2065.0 -> 2074.5 | 344.3 -> 340.7 |
| warrior_protection_max_survival | 1795.2 -> 1847.5 | 1805.8 -> 1822.3 | 1757.8 -> 1772.7 | 217.0 -> 215.3 |
| warrior_protection_max_threat | 4235.0 -> 4221.3 | 2346.0 -> 2342.7 | 2172.7 -> 2175.3 | 463.5 -> 472.0 |

Cause: Unbridled Wrath on abilities (more Rage, more Heroic Strike). Max threat's dip is Broodlord alone (4064 ->
3956 ± 29, healing needed 638 -> 709, an unpaired 1k row). Checked with old (1a85df1d7) and new tankbench builds,
seed 2, 3,000 iterations, Max threat score: Broodlord 4006 -> 4021, Chromaggus 3756 -> 3763, Golemagg 4236 -> 4269,
Magmadar 4329 -> 4340, Onyxia 4318 -> 4321, Vaelastrasz 4672 -> 4722 (mean about +20). The move is a ruled
hotfix's, upward everywhere; no finding.

## Rotation lines

Only the warrior's Rage income moved, so this pass re-tested the Rage-gated lines. Paired `abtest`, own weight,
six scoring bosses, scoring-mean diff (B minus preset):

| request | change | 1k seed 1 | 3k seed 1 / seed 2 |
|---|---|---|---|
| max_threat | Heroic Strike at 20 Rage (was 30) | +4.9 ± 3.1 | +1.5 ± 1.8 / +1.2 ± 1.7 |
| max_threat | Heroic Strike at 25 | | +0.9 ± 1.5 / +0.7 ± 1.4 |
| max_threat | Heroic Strike at 40 | -8.8 ± 3.5 | |
| max_threat | Bloodrage under 20 Rage (was 30) | -4.9 ± 1.6 | |
| balanced | Heroic Strike at 30 (was 40) | -1.2 ± 10.5 | |
| balanced | Heroic Strike at 50 | -10.5 ± 11.2 | |
| balanced | Bloodrage under 20 | -8.0 ± 6.7 | |
| max_survival | Heroic Strike at 30 (was 40) | +3.4 ± 25.0 | |
| max_survival | Heroic Strike at 50 | -43.1 ± 22.5 | |
| max_survival | Bloodrage under 20 | -5.7 ± 4.4 | |

Nothing passes the bar: the shipped thresholds still hold under the new Rage income (Max threat's lower HS gate is
+0.05%, within noise). Variants: `tools/audit/tank_rotations/mk_warrior.py <req> <out> setcond <line> <cond>`.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call, unchanged.
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling, unchanged.

## Notes for the orchestrator

- The pass doc's open item (main-hand timing for Windfury) was done for all four tanks in pass 15; it can be removed
  from passes/tank_rotations.md.
- The first `go build ./tools/tankbench` failed on a missing `sim/core/proto` package; a `make sim/core/proto/api.pb.go`
  then reported it up to date and the build worked (likely another auditor generating protos at the same time).
