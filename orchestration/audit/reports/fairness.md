# Fairness pass 27 (2026-11-04, master f331f0c)

Result: quiet (no new finding). This was a full pass: 126 watched files had changed since f23ea25, among them class code
(shaman, hunter, rogue, paladin, warlock, warrior, druid), every request (Salvation dropped), the item database and the
standards 37 gating. I re-swept all 43 builds (5k, seed 1, race overrides) on the loop's f331f0c CLI and tree, and the
new baseline is f331f0c.

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Shared 20k run (`runs/vs_published.txt`): 62 of 68 builds are within 0.5 of their published number. The six shamans
  (+199 to +277) are the stale metas that results-20261102-1 already tracks; builds.json matches all 68.
- Salvation (90e0bb01a): gone from every request, race file and mksettings, so the raid stays symmetric. It was threat
  only, so no DPS moved.
- Standards 37 gates (13c34f20f): shaman, hunter, warlock, warrior and rogue values for runs outside Forever. No
  Forever number moved.
- Sweep against the f5c99ad baseline (`table.py` flags 7 rows):

| Build | Moved | Attributed to |
|---|---|---|
| Smite | +19.1 (761.9 -> 781.0) | Spiritual Guidance build, da726bbe1; published 781.0 |
| Affliction | +18.4 | the Wrack preset (pass 26) |
| Feral | +8.7 | druid rotation finds, d560d30b8; published 1139.0 |
| Balance | +1.2 | same; published 952.2 |
| Ret / sub-Holy | +0.6 / +0.1 | Holy Strike ahead of Judgement, e799dccfa |
| Enhancement / imp-gw | +0.4 / -0.8 | 5k noise (no shaman Forever change) |
| Elemental | 0.0 | flagged only for the stale-meta gap |

Nothing moved in a class whose code didn't change, so nothing shared leaked.

- Published against current: every base build and variant is within 1.4 at 5k, the shamans aside.
- Best race against the report race: the best race everywhere within noise (Combat Swords Undead 0.5 under Orc, MM
  tallstrider / wolf Orc 0.8 / 0.5 under Night Elf).

## Results auditor overlap

- results-20261102-1 (shaman metas): confirmed independently by the 5k sweep (Enhancement +276.4, Elemental +201.9,
  the shaman variants +200 to +272). Not filed again.

## Ranking

The order is unchanged from pass 26. Smite rises to 781.0 and stays last. Top over bottom is 1.68x (was 1.73x). Full table:
`orchestration/fairness_report.md` §1 and `ref/fairness/table_baseline.md`.
