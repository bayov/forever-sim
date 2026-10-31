# Mechanics pass 53 (2026-10-31, master d96adcce1)

Range audited: e05fd732e..d96adcce1. One non-merge code commit on watched paths (fedb73949, mksettings `-sunder ramp|full|none`, default ramp); report_hashes.txt re-recorded (d32d6210f). TASKS.md: d96adcce1 adds the October 30 hotfix items (paladin / druid trees, Frost mage, priest / warlock). No `forever-update/*` merge in range.

## Get current

- Client: 1.60.1.70707, unchanged (probe 300).
- Dev notes thread 2360696: still 8 posts, edit times unchanged (post 8 10-27 16:46:55, post 7 10-27 16:47:14).
- Blue tracker: no posts since 2026-10-31 12:54 UTC.
- Server hotfixes: read from wago.tools/hotfixes (`?search=<spellID>`, JSON in the page's data-page attribute) for 24907, 24932, 11597.

## Code range

| Commit | What | Verdict |
|---|---|---|
| fedb73949 | mksettings: raid Sunder Armor ramps from the pull at 60 (tank 1/GCD, 1 pull Sunder from a DPS warrior, none for warrior presets) | Raid assumption (standard 32). Engine `scheduleSunderRamp` (sim/core/debuffs.go:945) unchanged since pass 52; 11597 on 70707 = 450 armor/stack, 30 s, GCD 1.5 s, matching `SunderArmorAura`. No finding. |
| d32d6210f | report hashes | Data only. |

## October 30 hotfix (shared part)

| Item | Tier 1 data | Sim | Result |
|---|---|---|---|
| Moonkin Aura 24907 | push 113283: aura 290 (all crit) 3, effect 65 | 3% spell crit | mechanics-20261031-2 |
| Leader of the Pack 24932 | push 113283: aura 290 (all crit) 3, effect 65 | 3% melee crit | mechanics-20261031-2 |

Measured with +3% crit (seed 1, 20k iterations): Fury +3.26%, Combat Swords +2.49%, Fire +2.77%, Elemental +2.29%. That is the effect if the two stack. If they don't, presets with both buffs stay as they are.

Class items in the hotfix (paladin and druid trees, Frost mage, priest, warlock) belong to the class workers and are already in TASKS.md.

## Open items

- tools/audit/mechanics/hotfixes.py is still missing from the repo. The wago.tools page search above works in its place.
