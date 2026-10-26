# Mechanics pass 40 (2026-10-26T13:30:00Z, master f62cb7971)

Range audited: f796d9953..f62cb7971 (3 non-merge sim commits). Result: quiet, no new findings.

## Get current

| Source | State |
|---|---|
| Client build | 1.60.1.70532, no build answered up to +300 |
| Dev notes thread t/2360696 | newest post /6 (10-16 00:15, edited 10-16 20:50), unchanged |
| Daily update | no forever-update merges, no TASKS.md / ledger changes in range |
| Server hotfixes | not re-read (hotfixes.py still not in the repo) |

## Commits

| Commit | Change | Check |
|---|---|---|
| 32a6848f9 | Five wowsims/forever mechanics as report assumptions | all ruled in standards 28 (user, 2026-10-26); not findings. Code reviewed below |
| ded834911 | Replay: channel log lines | log-only (dot.go, cast.go messages); wand Shoot excluded |
| 76db8c9c7 | Replay: pushback log lines | log-only |

## Code review of 32a6848f9

- **Live DoT ticks** (`spell_result.go` liveTick): Forever only, non-physical, non-helpful dots that went through Snapshot. Base moves by coef x (bonus now - bonus at snapshot), multiplier by (attacker mult x dot mult now) / at snapshot, crit by the difference, so post-Snapshot spell-specific edits to the snapshot survive. Target-side multipliers already apply at tick. Rollovers keep the old snap values. Correct.
- **Channel FSR** (`mana.go`): hold in channel OnGain only for mana-paid channels; release in OnExpire uses min(now, ExpiresAt) for lazy expiry. Since IsChanneling is ChanneledDot != nil, the next spell cannot spend its cost before release, so release can't roll back a newer FSR refresh. Correct.
- **Multi-Shot**: client 70532 SpellMisc 2643 CastingTimeIndex 1 (instant) and Attributes_2 0x20000 (Do Not Reset Combat Timers) confirmed; the 0.5 s cast is the ruled `hunter_multishot_cast` assumption; KeepsSwingTimer matches Aimed / Sniper Shot.
- **Hunter base RAP** 2L - 10 (110 at 60), `hunter_base_rap`; Bite miss keeps extra energy and CP, base 35 not refunded (no Refund on the cost), `druid_bite_miss`.

## Open findings

- mechanics-20261025-1 (Spellstone Fire power, low): still open, `sim/warlock/talents.go:139,147` unchanged.
