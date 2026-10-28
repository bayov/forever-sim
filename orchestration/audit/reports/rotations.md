# Rotations pass 38 (2026-10-28)

- **Master:** dd663e652 (last audited c9d93dd66). Client 1.60.1.70630 data, unchanged.
- **Result:** no new findings. rotations-20261028-1 (hunter Trueshot/Hawk caveats) still open: the phrasing
  pass reworded the lines but kept the old numbers.

## What changed

- No non-test sim `.go` changed. `apl_logic_diff.py c9d93dd dd663e6`: 23 Level 60 APLs, all from the
  standards 31 "rotation lines state their intent" round plus one Smite fix:
  - dd4ff5d91 Warrior: Recklessness with the last Death Wish; Elune's Light (1259799) not during Recklessness.
  - 8624ba126 Mage: Frostfire Evocation cutoff 90 s -> 75 s, matching Mana Gem gate; Fire racials after Scorch.
  - d387da420 Warlock (8 APLs): Troll Berserking just before a filler, not on the DoT opener.
  - 3ad93a86b Elemental (base and Mana Tide): low-rank Lightning Bolt when mana won't last the fight (mana < 30 x time left) instead of < 20 % mana.
  - d8185e98f Ret Holy Shock (3 APLs): Light's Vigil floor from mana, not fight length.
  - e419e7f50 Smite: Chastise (1277335) removed.
- The 23 reqs differ from c9d93dd only in the embedded rotation.

## Checks

Paired A/B, old rotation (c9d93dd req) vs new (master req), on master's req and the shared CLI, 20k it,
8 blocks, both seed sets, 180 +/- 20 s variation kept. B-A pooled (DPS):

| preset | 120 s | 180 s | 300 s |
|---|---:|---:|---:|
| warrior__fury | 0.00 | 0.00 | +6.3 |
| warrior__fury-duty-battle-shout | 0.00 | 0.00 | +5.7 |
| warrior arms (3), fury_2h (2) | 0.00 | 0.00 | 0.00 |
| shaman__elemental-level60 | +0.01 | +0.52 | +9.6 |
| shaman__elemental-level60-mana-tide | 0.00 | +0.16 | +1.9 |
| mage__frostfire | 0.00 | +0.49 | +0.31 |
| mage__fire (Human) | 0.00 | 0.00 | 0.00 |
| paladin__level_60_shock | 0.00 | 0.00 | +11.4 |
| paladin__level_60_shock-duty-crusader | 0.00 | 0.00 | +10.4 |
| paladin__level_60_shock_undead | -0.08 | -0.22 | -0.01 |
| priest__smite (Undead) | 0.00 | 0.00 | 0.00 |
| warlock, 8 presets (Human) | 0.00 | 0.00 | 0.00 |

Race checks (racial lines inert on the base presets), RaceTroll copies of the reqs:

| preset (Troll) | 120 s | 180 s | 300 s |
|---|---:|---:|---:|
| warlock__destruction | +1.21 | +0.31 | +0.90 |
| warlock__affliction | +1.79 | +0.79 | +0.96 |
| mage__fire | +0.69 | +0.14 | +0.46 |

- Every change holds or gains at 120, 180 and 300 s; the longer-fight gains are the intent of standards 31.
  shock_undead's -0.2 at 180 s is inside the noise (seed sets +0.04 and -0.47 at SE 0.14-0.25).
- Shared run vs published (`runs/vs_published.txt`): only mage__frostfire moved (+0.9), as expected from the
  Evocation change.
- Baseline not refreshed (no sim change): `ref/rotations/baseline_1274e15_3000it_seed424242.tsv`.
- Open finding rotations-20261028-1: `tools/dps_report/req/hunter__{beast-mastery,marksmanship,lone-wolf}.meta.json:65-66`
  and `ui/dps_report/data/notes.json` still say Trueshot 100 RAP and Hawk rank 5 (90).

## Not covered

- Night Elf Fury Elune's Light (needs the Elune's Light req lines) and Dwarf Smite were not A/B'd; the
  Smite removal is a fix, not a tuning change.

## Optional

- None new. Earlier optional items in `passes/rotations.md` stand.
