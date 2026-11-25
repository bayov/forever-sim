# Tank mechanics pass 34 (2026-11-25, master 85691655a)

No new finding. Client still 1.60.1.71091 (wago latest); no new Blizzard post (hotfix thread t/2403040 last post
2026-11-24, dev notes t/2360696 2026-11-04). Range c3bbe33..85691655a (21 commits); tank reach:

| Commit | Tank reach | Verdict |
|---|---|---|
| 2f42f4a44 Tank batch (threat, damage taken) | Prot warrior, bear, engine | Read in full; matches its EV citations (tier 1.5) |
| 22d2cf952 six rulings replaced by logs | Prot warrior (Shield Slam MH special) | Correct; WF Totem and weapon procs suppressed, UW excludes it by spell |
| b663145ad proc rules | Prot paladin (Redoubt EV-0022, Vengeance EV-0165) | Correct; Redoubt any outcome under Forever only |
| 295206ddf items (Force of Will 11810) | any tank wearing it | -51 flat before armor, half-hit cap; no request wears it |
| others (hunter, casters, report, TASKS) | none | - |

## Warrior (2f42f4a44, 22d2cf952)

| Item | Sim | Source / tier | Verdict |
|---|---|---|---|
| Rage-gain threat | 5/Rage, unscaled for Bloodrage instant, UW, Shield Spec, Master of Defense; 0 before pull | EV-0043/44/45/174, 1.5 | ok, TestForeverTankRageThreat |
| Bloodrage ticks 29131 | 5 x stance x Defiance | EV-0044, 1.5 | ok |
| Heroic Strike flat | 2.5 x 56 = 140 x stance | EV-0004/48/174, 1.5 | ok, abilities_test |
| Cleave flat | 96 split over landed targets | EV-0050 (r4 extrapolated), 1.5 | ok |
| Slam / Thunder Clap / Battle Shout | 1.5 x level / 1 x dmg / 0 | EV-0004, EV-0048, EV-0043 | ok |
| Shield Slam | 625-685 + 2 x BV, 255 flat once, MH special, no WF/weapon procs | client 70425 + EV-0004/50, R-0010/41 | ok, TestForeverShieldSlamIsMainHandSpecial |
| Shield Block last charge | aura up one 10 ms batch, cancelled on refresh/expiry | EV-0004/47 (0-2 ms) | ok, TestForeverTankShieldBlockBatch |

## Bear (2f42f4a44)

| Item | Sim | Source / tier | Verdict |
|---|---|---|---|
| Lacerate | 10/3 x dmg, hit and ticks | EV-0034/37, 1.5 | ok, bear_rules_test |
| Primal Bite | 3 x dmg (notes' "roughly doubled" retired) | EV-0034, 1.5 | ok |
| Faerie Fire | 2 x level flat, Bear/Dire Bear only (x2.15 = 258) | EV-0034/37 | ok, TestBearFaerieFireThreat |
| Blood Frenzy Rage | unscaled 5/Rage | EV-0035 | ok, TestBearBloodFrenzyRageThreat |
| Enrage armor cut | also 2 x Agility and bonus armor share | EV-0109/32/34 | ok, TestEnrage |

## Engine

- Damage-taken Rage armor rows re-fit (38..59); level 63 still 40%, so level 63 scores don't move. rage_forever_test.
- Flat physical reductions together cap at half the hit (spell_resistances.go:26); positive bonus (Gift of Arthas)
  x BonusCoefficient (white swings 1). Stoneskin test expects the cap.
- Tank sheet BiggestHit (tank_sheet.go:107-117) leaves out BonusPhysicalDamageTakenBeforeArmor though its comment
  says "our flat"; display only, no request carries Stoneskin or Force of Will. Nit, not filed.
- Tank score replay untouched in range.

## Checks run

- `go test -tags=with_db ./sim/warrior/tank_warrior/ ./sim/druid/tank/ ./sim/paladin/protection/ ./sim/shaman/warden/ ./sim/core/`: all pass.
- tankbench -labeled, balanced requests, 3000 iters (objective Max threat / Balanced / Max survival):
  bear 3979 / 3320 / 1343, paladin 2583 / 2196 / 1036, Warden 1778 / 929 / -1616, warrior 3478 / 2929 / 1279.

## Open findings

tank_mechanics-20261124-4, -20261124-3, -20261117-1, -20261108-1, -20261107-1 (unchanged).
