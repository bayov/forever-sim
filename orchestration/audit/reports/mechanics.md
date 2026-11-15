# Mechanics pass 84 (2026-11-15, master 25ce81de7)

Range 70ea588bb..25ce81de7. Client 1.60.1.70938 (probe 300: nothing newer). No new findings; three open findings resolved.

## Get current

| Source | Newest | New since pass 83 |
|---|---|---|
| Client (wago probe 300) | 1.60.1.70938 | none |
| Dev notes t/2360696 | post 10, 2026-11-04 (post 1 edited 2026-11-03) | none |
| Hotfix thread t/2403040 | post 10, 2026-11-14T01:54Z (post 1 edited 01:54:25Z) | none |
| Blue tracker | 2026-11-15T01:18Z dungeon XP post | none |
| TASKS / daily update | Windfury Totem strike type and Execute double roll: user ruling 2026-11-15 keeps the sim (one roll per ability; white extra swing) | settled, not findings |

## Commits triaged

| Commit | Change | Check | Verdict |
|---|---|---|---|
| 9ec69044f | Windfury Totem under Forever: every landed MH melee hit, minus SpellFlagNoWindfuryTotem (seal procs); allow list deleted | matches Nov 13 hotfix text (tier 1) | ok; resolves mechanics-20261112-1 |
| 9ec69044f, 6ce77e5ef | Judgement of Command melee special on MH table (no rangedJudgement); JoR / JoF stay ranged | Nov 13 hotfix (tier 1); JoC has no NoWindfuryTotem flag, so it rolls the totem as the hotfix says (6ce77e5ef's message saying otherwise is stale, code is right) | ok; resolves mechanics-20261114-1 |
| baa363615 | 2H Weapon Spec Holy part by defense type | user ruling + log (EV-0064) | settled |
| 78c395984 | Imp Firebolt r2-4 | = 70938 SpellEffect | ok; resolves mechanics-20261113-1 |
| 78c395984 | Seal of the Dawn 78 AP, Rune of the Dawn 46 SD (Forever) | sp.py 23930 bp 78 aura 102/131; 24198 bp 46 aura 180 | ok |
| 78c395984 | Wild Growth 0.066 | healer auditor's row | not checked here |
| 9ec69044f | Wildstalker 5pc Multi-Shot mask (hotfix push 114065) | class change, hotfix-only | left to hunter/class audits |
| 9f8a71d2d | Windfury Weapon ICD moved onto its aura (APL readable) | no value change | ok |

Tests: `go test -tags=with_db ./sim/paladin/...`, `./sim/warlock`, `./sim/core -run Windfury` pass on 25ce81de7.

## Open

mechanics-20261113-2 (Weakening Disease), mechanics-20261104-1 (druid 70790 scales): untouched in this range.
