# Mechanics pass 55 (2026-11-01, master c23ef0440)

Range audited: dd053d24c..c23ef0440. Eight non-merge commits on watched paths; TASKS.md / ledger commits 98e819d39, 86741656a, d463df97c (Kang and druid crit auras settled by user ruling). No `forever-update/*` merge in range.

## Get current

- Client: 1.60.1.70707, unchanged (probe 300).
- Dev notes thread 2360696: still 8 posts, edit times unchanged.
- Blue tracker: newest post 2026-10-31 02:17 (Midnight hotfixes); nothing new for Forever.
- foreverchanges.pro: newest entry is still "A hotfix to the October 30 build" (racial pass already landed in 500e3a1; Moonkin / LotP, frost mage, Arcane Blast, Righteous Fury, Improved Drains now in the sim).
- wago.tools hotfixes: push 113486 (70707, region 91) confirms Arcane Blast 400573 at 12% Int (e0) / 100% mana (e1), as the sim now has.

## Code range

| Commit | What | Verdict |
|---|---|---|
| 5b2eec6de | Two-hand white crit doubles the swing's whole Rage (was one-hand-rate bonus) | Consistent with the Forever logs cited (16.2 -> 32.4, 14.8 -> 29.7). Only `foreverWhiteHitRage` changes; one-hand crit unchanged (already x2). Verdict row updated. No finding. |
| f962f6c8f | Moonkin Aura / LotP: 3% melee + 3% spell crit once under Forever | Implements mechanics-20261031-2 per the 2026-11-01 ruling. Druid self-auras go through the same raidBuffs path (druid.go:138-142), so no double count. Camp Chair still excluded only by Moonkin (tooltip rule, unverified; unchanged). No finding. |
| 3607e01f3 | Kang the Decapitator Rend proc | 17153 on 70707: 15 s, aura 3, bp 100 every 3000 ms. Matches. 1 PPM assumed. No preset wears it. |
| b79e033bc | Righteous Fury +33% all threat; Arcane Blast 12% / 100% / 18 s | Class; Arcane Blast values match hotfix push 113486. |
| 98e819d39 | Frost mage Oct 30 hotfix values (push 113486) | Class; matches foreverchanges' Oct 30 entry text (Ice Lance 300%, Ice Barrier 15% / 5%, Winter's Chill 20%). |
| aea78721b | Improved Drains off Wrack | Class; matches the Oct 30 hotfix text. |
| 560579d78, e9344df43 | Hunter pets: no Pet Aggression; swing from creature dps less pet AP | Class (pet auditor's area); hunters are -4% to -8% vs published in the shared run, expected from these commits. |

## Shared run vs published

Two-hand warriors +0.9 to +1.3% (Rage crit). Casters -0.4 to -0.9% (explosives at 10 yd and caster ranges from earlier commits, audited pass 54). Frost mages +2%, Arcane -2.4% (hotfixes). No unexplained shared shift.

## Open items

- mechanics-20261101-1, mechanics-20261031-1, mechanics-20261027-2 still open.
