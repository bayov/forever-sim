# Rank tuning pass 1: master 0a204d4 (2026-10-17 14:58 UTC)

Result: **quiet** (no findings).

## Checked

| Check | Result |
|---|---|
| New client build (`wago.py latest`) | none; newest is 1.60.1.70450 (baseline) |
| Sensitivity spell files, d2633f4..0a204d4 | unchanged |
| `sim/` and `tools/dps_report/req`, d2633f4..0a204d4 | b803e88: shaman heal rank families (`healing.go`, `shaman.go`, heal tests); b1b0763: item db regenerated from the Wowhead scrape (golden results only). No req file changed. |
| 66 bases, shared 20k run vs `baseline_d2633f4.csv` | 66/66 matched, largest move 0.001%: level within noise |
| Sensitivities | not re-run: bases and patched spell code unchanged, so percentages and place shifts carry over |

## Conclusions

- Registering Healing Wave, Lesser Healing Wave and Chain Heal as rank families doesn't affect any DPS preset.
- The item db regeneration moved only golden test gear, not the preset gear.
- Every verdict and sensitivity in `orchestration/rank_tuning.md` stands. The baseline stays `baseline_d2633f4.csv`.
- Added a dated check section to `orchestration/rank_tuning.md`.
