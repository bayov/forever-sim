# tank_gear pass 3: 2026-11-04, master dc303b0

Forced, after tank round 3 landed (dc303b0). Watched changes since b38ff16: six tank gear sets (two new Max threat sets, paladin and bear) and standards.md (items 4 and 40). The synthetic tank items (`tools/database/forever_synthetic_items_*_tank.go`), the Tanks rows of gear_targets.md and standards 36 did not change, so steps 2 and 4 stand. No sims this pass (A/B evidence is the workers', in docs/tank_<class>.md and orchestration/tank_handoff/<class>/round3*/).

## Pieces and shares (recomputed with tools/audit/tank_gear/*.py)

| spec | set total (target) | Stamina share (row ±1.5) | top non-Sta | verdict |
|---|---|---|---|---|
| Prot warrior | 760.0 (761.0) | 31.4 (31.4) | 30.8% Def | ok |
| Prot paladin | 841.8 (844.1) | 32.6 (32.4) | 27.2% | ok |
| Feral bear | 730.5 (729.3) | 34.8 (33.4) | 34.4% Agi | ok |
| Warden | 845.1 (844.1) | 23.9 (23.9) | 20.4% Int | ok |

Per-piece linear and norm unchanged from pass 2's table (all within 1% / 6%).

## Gear sets against tankbench requests (items and enchants, all 17 slots)

| class | Max threat set | Balanced set | Max survival set | req match |
|---|---|---|---|---|
| warrior | level60_phase1_synthetic | same | ..._armor_trinket | 3/3 |
| paladin | ..._max_threat (new) | level60_phase1_synthetic | same | 3/3 |
| bear | ..._max_threat (new) | level60_phase1_synthetic | same | 3/3 |
| Warden | phase1_synthetic_tank | ..._armor_trinket | same | 3/3 |

Consumables also match presets.ts: bear Max survival Dirge's chops, Warden Max threat Natural Accuracy (+5% melee and spell hit; spell hit makes it useful past the yellow cap; A/B in docs/tank_shaman.md).

## Enchants (round 3)

| slot | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| Shield | 1886 | 863 Lesser Block | n/a | 1886 |
| Wrist | 1248665 | 1248665 | 1248665 (MT 1885 Sup Str) | 1248665 (MT 1886) |
| Back | 1219587 | 1219587 | 1219587 | 1889 |
| Feet | 929 | 929 (MT 1887 Agi) | 1887 | 929 |
| Head / legs | 1503 | 1507 | 1508 (MT 1506 Str) | 1507 |
| Weapon | 1900 | 1900 | 2646 (MT 1248668 Str 25) | 1900 |

Every id is live for its slot in db.json; every new pick is a world-drop formula (863: 11168; 1885: 16246; 1887 boots: 16245) and is backed by a recorded A/B. Lesser Block was A/B'd on all three shield tanks (wins only for the paladin).

## Pass-2 findings

| finding | verdict |
|---|---|
| -3 enchant values (shield 929, wrist Deflection) | closed: no 929 on any shield; wrists A/B'd for all four and worn as docs say |
| -7 cloak A/B | closed: all four A/B'd and recorded; sets match |

## DPS Report

The tank data (ui/dps_report/data/tank_*.json, last 4feee54c2 at 17:13) predates round 3 (18:01-18:50), so its numbers and enchants were not judged against round 3. One generator problem holds regardless of the regen: enchant names are looked up by effect id only (new finding -8).

## Nits (not findings)

- docs/tank_paladin.md:454 still says the warrior and Warden were not tested with 863; both were (docs/tank_warrior.md:230, docs/tank_shaman.md:74).

## New findings

- tank_gear-20261104-8 (low): DPS Report names enchants by effect id alone (wrist 1886 shows as "Enchant Shield - Greater Stamina", etc.).
