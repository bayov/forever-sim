# Consumables pass 92 (master a6677362f, client 1.60.1.70959)

Result: quiet. No new findings. Two open low findings (consumables-20261110-1, -20261105-1) unchanged.

## What moved since 8219f1083 (4 watched files, no presets.ts / mksettings / gear / consumes.go change)
| file | change | verdict |
|---|---|---|
| sim/core/buffs.go | `Assume("hunter_tsa60")` knob re-added at level 60 (default unchanged); Warchief's Blessing now also gives 15% cast speed under Forever (combat log 2026-11-18) | no preset/mksettings effect; value is a log-backed ruling |
| sim/core/debuffs.go | `Assume("hunter_hm60")` knob on Hunter's Mark (raid debuff and own mark); default unchanged | none |
| tools/dps_report/assumptions_def.py | `haste_gcd` now a single-value "Yes, for spells" row | assumption, not consumables |
| orchestration/standards.md | rule 27 rewritten (haste shortens spell GCD) | doc |

## Verified
- scan.sh: banned items, Revelation, trinkets, +2% crit elixirs all empty. Totem counts as baseline.
- Minor-armor exclusive effects count is 3, not the doc's 2: the third is Dust Cloud (ForeverDustCloudAura, client 70532), unchanged since the last pass. Pass doc's expected count is stale (nit).
- wago latest probe: 1.60.1.70959, no newer build.

No open items beyond the two low findings.
