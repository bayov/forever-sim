# tank_rotations: pass 3 (2026-11-04, master dc303b0)

Forced by hand after tank round 3 landed (dc303b0 and earlier): engine fd9538c (mana-gain threat unmultiplied,
disjoint tank seeds via tankscore.Seed, six-boss error x1.25), then the four class workers fixed pass 2's findings
and re-optimized every preset (orchestration/tank_round3/*.md; raw output in orchestration/tank_handoff/<class>/round3/).

## What changed (b38ff16..dc303b0)
32 watched files: all 12 preset requests' rows, the Warden's Balanced and Max survival APLs (armor trinket hold),
bear Max threat (Enrage under 70, a Max threat gear set), warrior Max threat (Shield Wall 13 s), paladin gear,
rotopt's tank templates (shaman trinketAt, druid enrageRage max 100), tools/tankscore/seed.go, the class sim files.

## Baseline (ref/tank_bench/summary.tsv, header now dc303b0)
Every request on every boss, 1,000 iterations, default seed. Every row reproduces the ref the class workers had
already committed (589208463), so nothing moved without a rotation, gear or engine change. Scoring mean (six
bosses without Ragnaros) at own weight, b38ff16 -> dc303b0:

| request | w | score | SafeDPS | DTPS | healing needed |
|---|---|---|---|---|---|
| druid_bear_max_threat | 1/3 | 2837 -> 2842 | 1685 -> 1699 | 578 -> 595 | 593 -> 618 |
| druid_bear_balanced | 1 | 1868 -> 1852 | 1603 -> 1589 | 576 -> 569 | 484 -> 483 |
| druid_bear_max_survival | 3 | -750 -> -634 | 1520 -> 1491 | 577 -> 570 | 463 -> 442 |
| paladin_protection_max_threat | 1/3 | 2785 -> 2804 | 1494 -> 1493 | 320 -> 304 | 238 -> 219 |
| paladin_protection_balanced | 1 | 2398 -> 2437 | 1457 -> 1453 | 301 -> 288 | 208 -> 191 |
| paladin_protection_max_survival | 3 | 1363 -> 1500 | 1457 -> 1453 | 301 -> 288 | 208 -> 191 |
| shaman_warden_max_threat | 1/3 | 2292 -> 2430 | 1593 -> 1684 | 807 -> 792 | 954 -> 1003 |
| shaman_warden_balanced | 1 | 548 -> 767 | 1551 -> 1577 | 792 -> 730 | 905 -> 846 |
| shaman_warden_max_survival | 3 | -4357 -> -3971 | 1483 -> 1517 | 754 -> 741 | 858 -> 833 |
| warrior_protection_max_threat | 1/3 | 3695 -> 3700 | 2052 -> 2051 | 417 -> 412 | 417 -> 411 |
| warrior_protection_balanced | 1 | 3008 (same) | 1964 | 367 | 324 |
| warrior_protection_max_survival | 3 | 1205 (same) | 1851 | 354 | 296 |
| warrior_protection (legacy) | 1* | -14 (same) | | | |

Death floor 0 everywhere. Moved beyond error: the three paladin rows, the three Warden rows and bear Max survival;
each has a round-3 preset commit and the engine change (not split). Every row matches the workers' own round-3
tables in docs/tank_{druid,shaman,warrior}.md. The paladin doc's "Next to the other tanks" table still quotes
round-2 bear and Warden numbers (1,839 / -839, 530 / -4,579): a stale doc, not a rotation issue.
Drift: in all 12 preset requests, priorityList and prepullActions equal the ui APL, and talents, race and gear
equal presets.ts and its gear sets.

## Pass 2's findings (paired, B minus preset, disjoint seeds 11-13 at 3k, 14 at 10k)
| finding | state | check |
|---|---|---|
| 9 Warden MS armor trinket held 25 s | resolved | 30 s -29 / -41 / -15, 10k -34.4 ± 10.0; 20 s -74 / -74 / -48; autocast -360 / -398 / -354 |
| 10 Warden Bal armor trinket held 25 s | resolved | 30 s +2 / -5 / -13, 10k -17.5 ± 4.1; 20 s -21 / -15 / -20; autocast -126 / -149 / -144 |
| 12 bear MT Enrage under 70 | resolved | 60 -3.4 / -5.3 / -3.9; 80 +0.5 / +0.3 / +0.5, 10k +0.4 ± 0.1 (0.01%, a plateau); none +0.3 / -0.1 / +0.2 |
| 13 seeds | resolved | disjoint seeds in use; every round-3 call keeps its sign on fresh seeds |
| warrior MT Shield Wall 13 s (round 3) | holds | 12 s -2.7 / -3.9 / -7.0, 10k -1.8 ± 1.4; 14 s +0.5 / -2.8 / +0.2, 10k -4.8 ± 1.4 (Vaelastrasz -34) |
| warrior Bal / MS Shield Wall 17 s | holds | 16 s -7.3 ± 4.4 / 0.0 ± 10.1; 18 s -3.8 ± 4.2 / -6.8 ± 9.6 (seed 11) |
| 3 paladin Bulwark 13 s | open, user call | not re-tested |
| 11 warrior MS armor trinket | open, user call | hold 30 s vs 35 s -29.6 ± 5.7 (seed 11) |
Warden Max threat keeps 990015 on autocast, as round 3 chose.

## Line-by-line probe (45 A/Bs, seeds 21 and 22, wins confirmed on 23)
New: finding 14 (Warden Searing Totem and Fire Nova in the last 60 s: Max threat +43.2 / +46.4 / +48.4 ± 9,
Balanced with Mana Spring +70.5 / +48.8 / +76.0 ± 13 to 15), finding 15 (Warden Balanced Mana Spring before the
pull, +12.7 / +10.6 / +10.8 ± 4 to 4.6).

Tried, no win (1k unless said): paladin Balanced Hammer of Wrath under 20% (-67.4), Consecration above Holy Strike
(-90.6), Holy Strike above Holy Shield (-20.0), Judgement before the pull (-100.3), a lower Consecration rank
(-62.7); round 3's knob output shows the Consecration mana floor flat 0-60 on the new engine. Warden Earth Shock
rank 6 (Bal -196.9, MT -247.7), rank 5 under 30% mana (-2.6 / -1.1, ties); Flametongue + Fire Nova (-103.2);
Searing + Fire Nova without a time gate (MT +34 but Chromaggus -100; Bal unconfirmed); Healing Stream before the
pull (Bal +24.2 / +0.5 ± 14.4, Broodlord swing, unsettled); Max survival fire totem / Mana Spring (unconfirmed,
± 30 to 36). Bear Balanced Enrage under 30 / 50 (-20.2, -0.5 ± 11.8), Max survival 30 / 50 (-34.0, -118.6).
Warrior Balanced Heroic Strike floor 40 / 60 (+11.5 ± 13.4, -13.1 ± 12.9: noise).

Chromaggus GCDs (pass 2's open item): structural. At Balanced the Warden casts Stormstrike 29, Earth Shock 16-17,
Lightning Shield 8, about 54 of 77 GCDs; both shocks are on cooldown and every GCD-only filler loses, except Fire
Nova inside the 60 s gate (finding 14). Mana runs out around 95 s and costs about 3 Earth Shocks.

## Open items for the next pass
- Finding 14's Vaelastrasz gain partly comes from mana spent under his regen being given back as mana-gain threat
  (fd9538c, still "to verify"); whether Essence of the Red's regen makes threat in game is tank_mechanics'.
- The paladin's pre-pull Judgement losing 100 for one GCD is large; possibly its cooldown pushing Swift Judgement
  out of line. Not chased.
- Scratch (gone with the session): $AUDIT_WORK/tank_rotations/{base,verify,probe}/.
