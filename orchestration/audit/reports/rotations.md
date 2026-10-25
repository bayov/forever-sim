# Rotations and resources: pass 31 (2026-10-25T18:55:28Z)

Master 016daa654 (last audited 54878fa3d). Result: quiet, 0 new findings.

## What changed

- Hunter Mark duty is now assumed for every hunter (user ruling 2026-10-25, standards 23). The
  `*_mark` and `*_scorpid` APLs and their `-duty-mark` / `-duty-scorpid` presets were removed. The base
  `bm`, `mm`, `lw`, `sv` APLs and the `_turret` rotations gained the two Hunter's Mark lines that the
  `_mark` files had: a pre-pull cast of 14325 at −1.5 s, and a refresh when `auraRemainingTime` on the target is under 3 s,
  placed right after `autocastOtherCooldowns`.
- `sim/hunter/hunters_mark.go`: the change is to the comment only. Apart from that, only `_test` files changed. The `tools/rotopt` hunter files are search tooling, and the presets don't reach them.
- Client 1.60.1.70532 is client-identical to 70450 (ledger, build 70532).

## Checks

| Check | Result |
|---|---|
| `apl_logic_diff.py 54878fa3d 016daa654` | 14 files changed. The only lines added are the two Mark lines in 7 base and turret APLs. The 7 `_mark` and `_scorpid` files were deleted. |
| Embedded req rotation vs `ui/hunter/apls/*.apl.json` (bm, lw, mm, sv) | All 4 are identical. |
| Req raid `huntersMark` debuff | Off in all 4 hunter reqs, so the mark isn't counted twice. |
| `baseline.py` (3k iterations, seed 424242) vs `baseline_54878fa` | The new hunter base rows equal the old `-duty-mark` rows exactly (BM 1171.20, LW 1156.89, MM 1165.69, SV 1055.72). The 7 removed rows are gone, and every other preset is unchanged. |
| Shared run vs published | The hunter rows match the published numbers. |
| Open findings to verify | None open. |

## Optional (under the bar)

- Hunter's Mark end-of-fight gate: adding `remainingTime > 6s` to the refresh line, on MM, 20k iterations.
  The gain shows only at 120 s, so it is a fight-length edge.

| s0 | 120 s | 180 s | 300 s |
|---|---|---|---|
| 1e6 | +0.23 ±0.02 | +0.00 | +0.00 |
| 2e8 | +0.21 ±0.02 | +0.00 | +0.00 |

## Baseline

The baseline is now `ref/rotations/baseline_016daa6_3000it_seed424242.tsv`. It replaces `baseline_54878fa`, because the presets changed. The sim did not change.

## Still open (carried over)

These are the open items in `passes/rotations.md`: Balance Night Elf with Elune's Light, holding Elemental Mastery, and screening the Warlock duty-curse variants.
