# Mechanics pass 64 (2026-11-04, master fb584eafa)

Quiet: no new findings.

## Range

f331f0cff..fb584eafa. No non-test commits under `sim/**/*.go` or `tools/rotopt/mksettings/**`.

| Commit | What | Mechanics check |
|---|---|---|
| fb584eafa | `sim/testdata/report_hashes.txt` (the gate's trigger) | Test data only. |
| 7b4dc5d97 | Warrior Fury presets / APL / requests | Preset choices, no mechanic code. |
| fdb14acd0 | UI presets drop Blessing of Salvation | Matches post 9's removal (pass 63). |
| d9ba4a55c | Standards 38 (Forever only) and 39 (cast time / 3.5 s for announced cast time changes) | Rulings; noted for future passes. |
| bf4bdbeaf, d3dc95920 | Downranking handoff and notes | Open user ruling on scope (heals only vs today's split). Max ranks at 60, so no report DPS moves. Added to verdicts' open items, not filed. |

## Getting current

| Source | Result |
|---|---|
| Client build | 1.60.1.70755 still newest (gate probe). |
| Hotfixes | `hotfixes.py pushes --set mechanics 70755`: no rows. |
| Dev notes post 9 | Edited 2026-11-04T02:12:27Z (version 3). Word diff against `ref/client_update/sources/devnotes_post9_2026-11-02.txt`: adds a Swiftmend developer note ("This ability does not consume the Rejuvenation or Regrowth effect on the target"), plus the Gamepad wording from pass 63. The sim has no Swiftmend; no change needed. The saved copy predates this edit. |
| Blue tracker | Nothing since 11-03 21:41 (Linxy, Bunmover). |
| Daily update | No TASKS.md / update_ledger.md changes and no forever-update merges in range. |

## Conclusions

Every verdict in `ref/mechanics/verdicts.md` holds on 70755. Open findings unchanged.
