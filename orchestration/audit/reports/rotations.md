# Rotations pass 52 (no findings)

- Master: fb584eafa (from f331f0cff), audited 2026-11-04. Client 1.60.1.70753 (no ledger change).
- No non-test sim `.go` changed. One Level 60 APL changed in logic (`ui/warrior/apls/level60_fury.apl.json`), so not quiet by the diff rule; screened it.
- Regression baseline kept (`ref/rotations/baseline_f331f0c_3000it_seed424242.tsv`): no sim change, only Fury's rows move by the rotation.

## What changed

| Commit | Change |
|---|---|
| 7b4dc5d97 | Dual wield Fury (`level60_fury`): Bloodthirst ahead of Whirlwind; Heroic Strike from 60 Rage (was 80) and Hamstring from 20 (was 10), both only when Bloodthirst and Whirlwind are >= 1 s from ready; Bloodrage under 50 (was 20). Duty presets keep their rotations. Fury 2H talents: Unbridled Wrath 5/5 (talents only, 2H APL unchanged). |
| fdb14acd0 | Salvation dropped from the UI raid buffs (req files already had it out). |
| d9ba4a55c | Standards 38 (Forever only) and 39 (cast time coefficient). No rotation effect. |

Checks: `tools/dps_report/req/warrior__fury.json` embeds exactly the new APL (priority list and prepull equal). Shared run: warrior__fury 1153.0 at 180 s.

## Screen of the new Fury rotation (`ab.py`, 20k it, 8 blocks, S0 = 1e6, shared CLI fb584eafaf39)

B-A at 120 / 180 / 300 s, each variant one line from master:

| Variant | 120 | 180 | 300 |
|---|---|---|---|
| Heroic Strike 50 | -2.58 ±0.67 | -2.55 ±0.59 | -2.84 ±0.25 |
| Heroic Strike 70 | -1.06 ±0.81 | +0.52 ±0.48 | -0.74 ±0.44 |
| Hamstring 15 | -0.30 ±0.39 | +0.04 ±0.37 | -0.36 ±0.35 |
| Hamstring 30 | -0.26 ±0.43 | -1.28 ±0.46 | -1.93 ±0.43 |
| Bloodrage < 40 | -0.02 ±0.18 | +0.37 ±0.29 | -0.72 ±0.29 |
| Bloodrage < 70 | -0.02 ±0.30 | -0.22 ±0.69 | -0.61 ±0.48 |
| Whirlwind before Bloodthirst | -1.17 ±0.57 | -0.64 ±0.47 | +0.10 ±0.26 |

No variant reaches the bar (none gains at 120 and 300 s), so no second seed set was needed. The new thresholds sit at a local optimum.

## Optional

None new. Earlier optional items stand (see `passes/rotations.md`).

## Open

- Fury duty presets (`fury-duty-battle-shout`, `fury-duty-sunder`) still run the old thresholds by design; whether the new ones carry over to them was not screened.
