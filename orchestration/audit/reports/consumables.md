# Consumables report, pass 47 (master 06a898751)

Result: quiet, no findings.

## Moved since d3df9247f
- assets/db_inputs/forever_enchants.json: meta only (rebuilt from client 1.60.1.70630); no enchant entry changed.
- tools/dps_report/assumptions_def.py: Missile Barrage / Arcane Missiles assumption now "No proc" (client 70630 tooltip); Seal of Command wording. Text only, no buff value.
- ui/retribution_paladin/presets.ts: preset renames (Level 60 Shock -> "Shock (deep Holy)", "Shock (Ret side)" -> "Shock") and comments. No Consumes, buff or debuff object changed.
- ui/warlock/presets.ts: comment only (Spellstone is Shadow-only since 70630; Destruction Spellstone vs Firestone even at 1009.1 / 1009.5, so the Spellstone choice stays within noise).
- Class and consumes branches: only the usual stale diverged diffs (gotcha 1).

## Verified
- scan.sh on master: banned items, Revelation, non-synthetic trinkets, double crit elixirs all empty.
- Counts unchanged: raid totems as before; minor-armor count 3 is the known unused Faerie Fire (Feral) / dust cloud effect (scan's expected 2 is stale).
- wago latest probe: 1.60.1.70630 (already audited).

No open items.
