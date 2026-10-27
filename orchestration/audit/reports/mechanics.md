# Mechanics pass 43 (2026-10-27, master 37a8a8ccd)

Range audited: 06a898751..37a8a8ccd. No sim or mksettings commit. One TASKS.md commit (37a8a8ccd): client 70634 and the post 8 racial additions. No `forever-update/*` merge; PR #117 (forever-update/2026-10-27-70634) is open.

## Get current

- Client: 1.60.1.70630 -> 1.60.1.70634. `wago.py diff 70630 70634 --by-spell` is empty and all 25 mechanics-set tables are byte-identical (both fetched fresh; 70634 shows 70630's Amplify Curse mask 32768, so it is a real fetch). No verdict moves.
- Dev notes thread: post 8 (Kaivax, 2026-10-27 15:37, edited 16:46) is new, and post 7 was edited 2026-10-27 16:47 to add the same three lines. Quoted:
  - "Blood Fury's cooldown reduced to 2 minutes (was 3 minutes)."
  - "Berserking duration increased to 15 seconds (was 10 seconds)."
  - "Touch of the Grave internal cooldown reduced to 15 seconds (was 25 seconds)."
  - Skyborne clarification: Charged and Attuned persist through death and are removed upon entering a raid (the sim's 50% Read Leyline stays right for raid presets).
- Server hotfix tool `tools/audit/mechanics/hotfixes.py` named in state does not exist in the tree; hotfixes not re-read this pass.

## Checks

| Item | Master | Tier 1 now | Verdict |
|---|---|---|---|
| Blood Fury CD | 3 min (racials_forever.go) | 2 min (post 8) | Finding mechanics-20261027-3 |
| Berserking duration | 10 sec (racials.go) | 15 sec (post 8) | Same finding |
| Touch of the Grave ICD | 25 sec (racials_forever.go:546) | 15 sec (post 8) | Same finding |

PR #117's sim diff makes exactly these three changes (Berserking 15 sec only under Forever, comments cite post 8 and client 70634). Reviewed: matches the notes.

## Open

- mechanics-20261027-3 (new, medium): fixed by PR #117 once merged; re-sim Orc, Troll, Undead builds.
- mechanics-20261027-2 (Amplify Curse x1.5 still on Bane of Agony): still open.
- Recheck the post 7 / 8 racials when a client build ships the rows.
