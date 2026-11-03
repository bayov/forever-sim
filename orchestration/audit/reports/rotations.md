# Rotations pass 50 (no findings)

- Master: f23ea25d2 (from f5c99ad01), audited 2026-11-03. Client 1.60.1.70753, unchanged.
- Not quiet by the diff rule: 4 Level 60 APLs changed in logic. No non-test sim `.go` changed (only
  `sim/compsearch/pool.go`, a preset name in the left-out list, and `dps_warlock_test.go`), so the
  regression baseline stays `ref/rotations/baseline_f5c99ad_3000it_seed424242.tsv`; `regress.py` not run.

## What changed (`apl_logic_diff.py f5c99ad01 f23ea25d2`)

| Commit | Change |
|---|---|
| 734fe0e93 | Warlock Affliction becomes the Wrack build with the Imp sacrificed (Undead, `sacrificedPet: Imp`). `level60_affliction` and both duty-curse APLs gain Siphon Life (18881) and Wrack (1316697) lines; the new `level60_affliction.apl.json` is line for line the old `level60_affliction_wrack` APL audited in pass 49. The old Shadow Bolt rotation lives on as `level60_affliction_shadow_bolt` (logic unchanged; the diff shows it as new because of the rename). |
| 203b686f6 | Enhancement `optimized`: Earth Shock (10414) waits up to 1 s for the Stormstrike mark (user's call). |

Shared run (20k, seed 1): Affliction 1034.1, duty-curse 1025.1, duty-recklessness 1019.9, Shadow Bolt
Affliction 1015.6. Per iteration on Affliction: Wrack 23, Immolate 9.5, Corruption 8.7, Siphon Life 5.7,
Shadow Bolt 2.3, Life Tap ~0 (mana is not the limit, as pass 49 noted).

## A/B screens (paired, 20k as 8 x 2,500, B-A at 120 / 180 / 300 s, `ab.py`, shared CLI f23ea25d20fc)

Warlock Affliction (`warlock__affliction`, base 1034.4 at 180 s), seed set 1e6:

| Variant | 120 | 180 | 300 |
|---|---|---|---|
| Corruption refreshed below 6 s (so it doesn't lapse under the 6 s Wrack channel) | -5.5 | -6.0 | -12.1 |
| Corruption and Siphon Life refreshed below 6 s | -12.0 | -12.8 | -22.4 |
| Corruption refreshed below 3 s | -2.4 | -2.4 | -5.6 |
| Wrack above Immolate | -12.6 | -13.5 | -8.2 |
| Immolate dropped | -12.6 | -13.6 | -8.2 |

All lose clearly on one set; no second set needed. A Shadow Trance (Nightfall) Shadow Bolt line is moot:
the build has no Nightfall, so the aura doesn't exist (the line ran unconditionally in the test, -190).

Enhancement Earth Shock hold, B = old unconditional Earth Shock, B-A:

| Preset | s0 | 120 | 180 | 300 |
|---|---|---|---|---|
| imp-gw | 1e6 | -0.10 ±0.37 | -0.40 ±0.20 | -0.82 ±0.18 |
| imp-gw | 2e8 | +0.18 ±0.20 | -0.18 ±0.18 | -0.52 ±0.28 |
| level60 | 1e6 | +0.07 ±0.21 | +0.18 ±0.27 | -0.56 ±0.20 |
| level60 | 2e8 | +0.20 ±0.20 | -0.13 ±0.20 | -0.45 ±0.34 |

The hold is neutral to +0.5 for the new rotation, matching the commit ("a wash on Level 60, about +0.4 on Imp GW").

## Conclusions

- The Wrack rotation holds against refresh-early and Immolate variants. Nothing to file.
- The Enhancement change landed as described and doesn't lose. Nothing to file.

## Optional

None.

## Left for a later pass

- Wrack duty-curse and duty-recklessness variants weren't screened separately. They share the base lines,
  so the base results should carry over.
- The open items in `passes/rotations.md` are unchanged.
