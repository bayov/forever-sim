# Mechanics pass 54 (2026-11-01, master dd053d24c)

Range audited: d1935640f..dd053d24c. Three non-merge commits on watched paths. No TASKS.md / update ledger change, no `forever-update/*` merge in range.

## Get current

- Client: 1.60.1.70707, unchanged (probe 300).
- Dev notes thread 2360696: still 8 posts, edit times unchanged (post 8 10-27 16:46:55, post 7 10-27 16:47:14).
- Blue tracker: no posts since 2026-10-31 19:10 UTC.

## Code range

| Commit | What | Verdict |
|---|---|---|
| bb21e3b32 | Goblin Sapper Charge and Goblin Land Mine usable only within 10 yd (`explosiveBlastRadius`, ExtraCastCondition); hunters / mages / Balance / Elemental drop or swap explosives | Verified on 70707: 13241 e0 and 4043 e0 ImplicitTarget 22/15, EffectRadiusIndex 13 (10 yd). Hunter weaves move the hunter in through movement.go, so the Sapper in a weave is consistent. No finding. |
| e563383a3 | gofmt | No change. |
| 905787afd | Warlock and priest presets at 25 yd (were mksettings' default 5 yd); goldens, report hashes | Warlocks: every APL spell is 30 yd, fine. Shadow: Mind Flay is 20 yd (24 with Shadow Reach), the sim casts it anyway: mechanics-20261101-1 (low, 0 DPS measured). |

## Measurement

priest__shadow at 25 / 24 / 20 yd (20k iterations, same seed): 948.73 / 948.73 / 948.73 DPS. Nothing the Shadow APL casts travels, so the distance only matters for whether Mind Flay is legal.

## Open items

- mechanics-20261031-2 (Moonkin Aura / LotP all crit stacking) still needs a user ruling.
- mechanics-20261031-1, mechanics-20261027-2 still open.
