# Mechanics pass 46 (2026-10-28, master 1274e15b3)

Range audited: 224b8a4e0..1274e15b3. One code commit: 73cedc9a9 (Hunter: launch realms' 2026-10-28 hotfix for Trueshot Aura and Aspect of the Hawk), touching `sim/core/buffs.go` and `sim/hunter/aspects.go`. Ledger change in the same commit (HUN-HF-01/-02 settled); no `forever-update/*` merge.

## Get current

- Client: newest 1.60 build is still 1.60.1.70658 (`wago.py latest --probe 300`). No re-check needed.
- Dev notes thread 2360696: still 8 posts; post 7 (edited 2026-10-27 16:47:14) and post 8 (edited 16:46:55) unchanged.
- Blue posts since 2026-10-28 01:28:41 (Kaivax, Aggrend, Gnedi, Linxy, Brenoch): none.
- Server hotfixes: `tools/audit/mechanics/hotfixes.py` is still missing from the tree; the wago.tools hotfix API route is gone (404). Not re-read.

## 73cedc9a9: Trueshot Aura and Aspect of the Hawk

| Item | Client 70658 (tier 1) | Sim now | Verdict |
|---|---|---|---|
| Trueshot Aura r1-r4 (20905 = r4) | 30 / 40 / 50 / 75 | 15 / 22 / 30 / 40 | Launch-realm hotfix (foreverchanges.pro), adopted by user ruling 2026-10-28. Not a finding. |
| Trueshot Aura r5 (20906) | 50 | 50 | Matches client. Ranks now monotonic, so `bestLearnedRank` picks the top rank. |
| Aspect of the Hawk r1-r5 (14321 = r5, bp 90) | 20 / 35 / 50 / 70 / 90 | 15 / 22 / 30 / 38 / 47 | Hotfix, user ruling. Not a finding. |
| Aspect of the Hawk r6 (14322) | 55, 110 mana, level 58 | 55, 110 mana; L60 casts r6 | Matches client. |
| Hawk r6 outside Forever | n/a | 110 (was 55 for every env) | Classic value restored; no Forever effect. |

Knock-on: `ForeverTrueshotAuraRank5RAP` / `foreverTrueshotAuraRAP` are read only by the raid Trueshot Aura in buffs.go (ranged AP only, so melee unaffected); every other suite's golden moves only its sheet's ranged AP by -50. Hunter's Mark (debuffs.go, 143) untouched. No DPS sim run: the commit's goldens (BM/MM/LW -1.5% to -1.9%) are the intended effect of the ruling.

## Findings

- No new findings. mechanics-20261027-2 (Amplify Curse vs Bane of Agony) stays open; no warlock commit in range.
- verdicts.md: hunter assumption row updated to the hotfixed curve.
