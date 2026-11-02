# Mechanics pass 59 (2026-11-02, master 796495d6f)

Range audited: b647a21ff..796495d6f. About 20 non-merge commits on watched paths, mostly the healing sim, the tank sim and the practice room. TASKS.md commits are healing / tank sim design and rulings. No `forever-update/*` merge in range.

## Get current

- Client: 1.60.1.70707, unchanged (probe 300).
- Dev notes thread 2360696: still 8 posts, edit times unchanged (post 8 edited 2026-10-27T16:46:55Z).
- Blue tracker: newest post 2026-10-31 02:17 (Midnight hotfixes); nothing new for Forever since 10-29 (live Nov 4 announcement).

## Code range

| Commit | What | Verdict |
|---|---|---|
| 76ee911a7 | 75% armor cap in `GetArmorDamageModifier` | Classic cap (tier 3); 3 x (400 + 85 x 63) = 17,265 checks out. Boss armor is far below the cap, so no DPS preset moves. Verdict row added. |
| d6f1333b7, 16df0b71c | Boss crit: `classic_crit` target option keeps Classic's rate | Off by default; Forever's no-crit-under-level rule unchanged for every preset. |
| a517e7e01 | `OutcomeEnemyMeleeSpecial` (boss abilities on players: one roll for miss / dodge / parry / block, own crit roll, no crush) | Enemy-side only, tank sim. Boss off-hand shares the main-hand table. No DPS effect. |
| 9c038d2f3, 6ca43223b, 27350b33a, 8dd2b8409 | Healing sim: health tracking, deaths, overhealing, raid damage | Gated on `Encounter.TrackHealth`; dead-player guards in spell / mana / auto attack only trigger when a player dies. `OnPeriodicDamageTaken` now skips auras ended mid-loop (safe). Healing threat uses the same effective-heal cap as before. |
| 6265a9ab7 | Lightning Overload copies are procs | 408477 Attributes_3 = 0 on 70707 (no Not a Proc), tier 1. Matches. Verdict row added. |
| e32660fe2 | Totem buff hold that ended before the pull drops at once | Correct: `latestHold` now starts at the earliest sim time, not 0. |
| 5a3a75309, 9a8051fdc, 5075b45e0, 035c27633, 437c6fda9, 7c13620ee | Practice room hooks (manual control, `SkipPullStart`, `FireTotemInReach`, `PlaceTotem`) | Nil / false in normal sims; no mechanic change. Healing Stream Totem now has metrics (bookkeeping only). |
| fa2d49997, 6c357819c | Caster and BM hunter distances (standards 33) | Positioning ruling, not a client value. |
| 80fec473d | APL condition memoisation | Engine plumbing; energy thresholds unwrap the memo. No mechanic. |

## Open findings

mechanics-20261031-1 (Efficiency / Lacerate) and mechanics-20261027-2 (Amplify Curse / Bane of Agony) remain open; no druid, hunter or warlock code touched them in range.

## Result

No new findings.
