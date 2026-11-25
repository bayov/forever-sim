# Rotations pass 76 (2026-11-25, master 63129ce35)

Quiet. Range 9470bae6a..63129ce35 (Phase 0 step 4: hunter and rogue gear lists, race rows, hunter_bm talents,
`tools/rotopt/bis_phase0.sh` and `phase0_specs.json`).

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py 9470bae6a 63129ce35` | 0 APL files touched |
| non-test `sim/**/*.go` | no change |
| embedded rotation in changed `tools/dps_report/req/*.json` (`jq -S`) | identical in every existing request; the raw diff on the hunter requests is key order of `notes` only |
| new race rows (hunter BM/MM/LW/SV, rogue Combat Swords, Seal Fate Swords) | gear/race copies; Seal Fate Daggers Dwarf row removed |
| `tools/rotopt/*` changes | tooling only, not read by the sim or the presets |

No sims run: nothing in the range reaches rotation logic, so the shared run needs no A/B.

## Open findings

Unchanged in this range (no paladin or rogue APL/req rotation edits): rotations-20261122-1, rotations-20261112-1.

## Open items

- Carried from pass 75: rotations of mage, paladin, shaman, feral, arms and shadow not re-searched on Phase 0 gear.
- Hunter BM on its new talents and the rogue Phase 0 sets (new off-hand stones) were not re-searched; worth a
  rotation check the next time a hunter or rogue rotation search runs.
- The standing items in `passes/rotations.md` (Windfury timing, warrior HS thresholds, paladin hasted GCD, etc.).
