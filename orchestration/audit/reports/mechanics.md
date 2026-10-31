# Mechanics pass 52 (2026-10-31, master e05fd732e)

Range audited: a2a3e66e1..e05fd732e. 4 code commits touching watched files. TASKS.md / update ledger: only the audit-closing commit 35ea909bd (Tauren health-only user ruling) and e05fd732e; no `forever-update/*` merge.

## Get current

- Client: 1.60.1.70707, unchanged (gate probe).
- Dev notes thread 2360696: still 8 posts; post 7 (upd 10-27 16:47:14) and post 8 (upd 10-27 16:46:55) unchanged.
- Blue posts: Linxy's newest is Midnight hotfixes (10-31 02:17); no Forever blue post found since 10-30 18:50. Kaivax's activity feed is not public (404).
- tools/audit/mechanics/hotfixes.py, named in the state, is not in the repo; server hotfixes were not re-read this pass.

## Code range

| Commit | Change | Verdict |
|---|---|---|
| e1fcd57b7 | Efficiency discounts Strider Kick | tier 1 match: 70707 19416 mask word 2 0x202000 holds Strider Kick's 0x200000 (1317257). The same mask's word 1 (0x80081000) also holds Lacerate's 0x1000, which the sim leaves out: new finding mechanics-20261031-1 (low, no report build affected) |
| e05fd732e | Raid sim threat check (watchThreat hooks in rage.go, spell_result.go) | observer only: no RNG, no events, nil unless the raid request sets ThreatCheck; no DPS effect. 110% melee / 130% ranged pull ratios are the Classic rule |
| 8d0a22be9 | mksettings per-build consumes (rogue Land Mine off, Ele Mana Tide potion, Shadow+Holy Mageblood, Smite) | consumes only, matching the ui presets; no request changed. Consumables auditor's ground |
| 159443615 | mksettings doc comment gear name | comment only |

## Conclusions

- One new low finding: Efficiency should cover Lacerate (client 70425 to 70707 all agree); the existing test reads the mask with the 0x1000 bit dropped.
- Verdicts: Efficiency row added; "as of" moved to 2026-10-31 / e05fd732e.
