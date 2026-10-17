# Rotations and resources: pass 27

- Master audited: 0a204d43d (previous: d2633f4a4), 2026-10-17.
- Result: quiet.

## What changed

| Area | Change | Reaches raid presets? |
|---|---|---|
| Level 60 APLs | `apl_logic_diff.py d2633f4a4 0a204d43d`: 0 APL files touched | n/a |
| `sim/shaman/healing.go`, `shaman.go` | Register Healing Wave, Lesser Healing Wave and Chain Heal rank families (`core.RegisterSpellRanks`) | No: no Level 60 shaman APL casts a heal |
| `tools/dps_report/req/*.json` | unchanged | n/a |
| Item database (`leftover_db`, Wowhead scrape) | regenerated; test `.results` moved | Checked with the shared run |

## Check

Shared run (20k iterations, seed 1) against the published numbers: `compare_runs`: 0 of 66 presets moved more than 0.5 DPS.
The baseline stays `baseline_d2633f4_3000it_seed424242.tsv`, because no sim change reaches the presets.

## Optional

None new. The optional list in `passes/rotations.md` still stands.
