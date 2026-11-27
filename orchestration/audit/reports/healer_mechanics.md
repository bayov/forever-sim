# healer_mechanics pass 35

Master 5ab589d858 (last: 34994816e, client 1.60.1.71139). Quiet: no new findings.

## Get current

- Client: newest build still 1.60.1.71139 (`wago.py latest`).
- Blizzard hotfix thread t/2403040: last post still 2026-11-25 23:25 UTC = `dev_notes_seen`. No newer hotfix thread in the
  forum search.
- No heal, talent or engine commits since last pass (`git log 34994816e..5ab589d858 -- sim/core sim/priest sim/paladin
  sim/druid sim/shaman` shows only the merge 6ba031e40, which touches none of them).

## What changed on the watched paths (all tools/healbench/req/)

| Commit | Files | Healing mechanic? | Verdict |
|---|---|---|---|
| ff87e66b3 | priest_holy.json, priest_discipline.json | APL only: Flash Heal first lines gain `currentManaPercent >= remainingTimePercent` | rotation, not mechanics (healer_rotations) |
| d74a53261 | paladin_holy(_raid), shaman_restoration(_raid) | gear/enchant swaps (1505 -> 1503 etc.) | gear, not mechanics |
| eef393de0, 11d2cc9c9, 62f3a9768 | druid_restoration(_raid) | Phase 0 gear set, enchants 1888->2463, 1219580->2567, +1255074 | gear, not mechanics |

No new spell, talent, set bonus or engine path enters through these requests; the heals they cast are the ones verified in
`ref/healer_mechanics/verdicts.md`.

## Open findings (unchanged on this master)

- -19 Oath of Light 60 min (holy_talents.go still 5 min); -18 Genesis on Tranquility; -17 Mana Well energize under
  Innervate; -11 Riptide cost 147/231 vs client 150/230; -10 untested paths. None fixed on 5ab589d858.
