# Findings: tank_mechanics

## Open


## Resolved

### tank_mechanics-20261104-12: Mana gains make threat times the tank's threat multiplier, with no source
- severity: high (Balanced bear -2.9% with Judgement of Wisdom off; how much of it is wrong is unknown)
- affects: engine; Feral tank most (2.15× form), Protection paladin and Warden too
- found: 2026-11-04, master b38ff167d
- evidence: sim/core/mana.go:121-127 `manaGainThreat` books 0.5 threat a mana × `PseudoStats.ThreatMultiplier`, "the way Rage gains do". The Rage rule (rage.go:476-492) rests on wowsims/forever 913a5bcf69 and c48ea6cea3, the owner's in-game threat measurements (tier 2). Nothing measures mana; Classic gives resource-gain threat unmultiplied (tier 3). Added in 42f982678 (healing sim, 2026-11-02); pass 1 wrongly listed "Rage and mana gain threat" as right. The bear lives off it: the pre-pull shift's mana comes back from Judgement of Wisdom at the pull, and docs/tank_druid.md:163 credits "+62 at Balanced" of the shift to it. Measured (bear auditor, paired, 3,000 iterations, Balanced preset with vs without the JoW debuff): Balanced score 1844 → 1791 (-53), SafeDPS 1439 → 1409; Max threat 2533 → 2476. Unmultiplied mana threat would take roughly half of the -53 off (estimate, not run: needs a code edit), still about 1% of Balanced.
- proposed fix: either find a source for the multiplier on mana (a wowsims/forever measurement or a Forever log: a bear or paladin's threat on a JoW tick), or use the unmultiplied 0.5 a mana under Forever, as Classic, until one turns up. Then rebench the bear, paladin and Warden presets. If the user would rather keep the extrapolation, record it as a to-verify value like Revenge's.
- How sure: sure the multiplier on mana has no source; unsure what the game does.
- needs: orchestrator fix (or user ruling, to keep it as to-verify)
- fixed: 2026-11-04, 49f0218 (+ re-bench c5200fc). Forever mana-gain threat is 0.5 a mana with no threat multiplier, as Classic and wowsims/forever (its mana.go books gain*ThreatPerManaGained unmultiplied); to-verify note in mana.go and standards 40 (keep unmultiplied unless a log shows otherwise). TestBearManaGainThreat fails with the old multiplier (107.5 vs 50 for 100 mana in Bear Form). The Rage rule's multiplier keeps its source: wowsims/forever 913a5bcf6's message, "Owner's in-game threat measurements: ... Rage gained is 5 threat per rage, times the unit's threat multiplier" (no log attached; c48ea6cea3 only moves the booking to the gain). Re-bench, 1k, own weight six-boss mean: bear 2837.3/1867.5/-750.2 -> 2808.5/1839.0/-775.0, paladin 2785.0/2397.7/1363.3 -> 2761.0/2373.7/1339.5, Warden 2292.3/547.8/-4356.7 -> 2272.2/529.5/-4367.5 (Max threat/Balanced/Max survival).

### tank_mechanics-20261104-13: Stale source labels and comments after round 2
- severity: low (docs and comments; no number moves)
- affects: Warden, bear, warrior, paladin docs
- found: 2026-11-04, master b38ff167d
- evidence:
  - Warden: hotfix push 113380 (wago.tools/hotfixes, region 91) puts Stormstrike's 150% (SpellEffect 1368308), its 6 s cooldown (SpellCooldowns 55241) and Earth Shock r7's base 531, 517-545 at 60 (SpellEffect 688559) in client data, yet sim/shaman/stormstrike.go:17-18 says they are "not in client 1.60.1.70753" and docs/tank_shaman.md rows 238 and 240 call them notes-only. The sim's values match the hotfix (tier 1).
  - Bear: sim/druid/bear_attacks.go:9-11 says each attack threatens for "the form's 1.5"; under Forever it is 2.15 and Primal Bite also ×2.
  - Warrior: docs/tank_warrior.md:213 calls the shield enchant "Superior Stamina (+9, enchant 1886)"; it is Enchant Shield - Greater Stamina (20017, enchant 1886, the id the bracer's Superior Stamina shares). tools/database/forever_enchants.go:94-96 says Superior Deflection is in "no preset yet"; 5d9a3aa18 put it on all three warrior presets.
  - Paladin: orchestration/round_70753/paladin.md:138 still says "(Templar's Bulwark magic: N only)", against :115, :210 and docs/tank_paladin.md:43 (N, C; 1311015 DispelType 1).
- proposed fix: cite push 113380 as tier 1 in the shaman rows and comment; reword the bear comment; "Greater Stamina (+9, enchant 1886)" and drop "no preset yet"; "N, C" at paladin.md:138.
- needs: worker fix
- fixed: 2026-11-04, 49f0218 (stormstrike.go and tank_shaman.md rows cite hotfix push 113380 as tier 1; bear_attacks.go comment; tank_warrior.md shield enchant; forever_enchants.go; round_70753/paladin.md:138).
- fixed (Warden part): 2026-11-04, 71a107e on tank-shaman-3 (the same stormstrike.go and doc citations), merged with 49f0218 in 6c73945.

### tank_mechanics-20261104-14: Scoring-run guards with latent holes, and SoD encounters still double parry haste
- severity: low (latent; no preset or scored encounter hits them today)
- affects: engine, the tank score
- found: 2026-11-04, master b38ff167d
- evidence:
  - The replay doesn't follow max health down: damage events (sim/core/tank_score.go:47-52) carry no max health and `lives()` (:151) caps only on heals, while the live run now cuts health at unit.go:341 (0aed93989). Only a timed Last Stand would hit it today, and no preset casts one.
  - APL values `CurrentHealth`, `CurrentHealthPercent`, `HealthMissing` (apl_values_resources.go:40, 76, 145) read live health in a scoring run; `readsNoHealth` guards only major_cooldown.go:153 and racials.go:172. No preset or rotopt tank template uses them.
  - tools/encounters/sod/*.json (all 7) keep `parryHaste: true` on log-fitted swing gaps, so comparing Anniversary with SoD counts SoD's parry haste twice (about +5% DTPS). Kept for comparison only (docs/tank_sim.md:68), not scored.
- proposed fix: record maxHealth on replay damage events and cap there; have MakeTank reject (or warn on) health-reading APL values in a scoring run; set parryHaste false in the SoD copies as 87f28803b did for the top-level files.
- needs: orchestrator fix
- fixed: 2026-11-04, 0950b22. Replay damage events carry max health and cap on it (TestTankScoreReplayCapsAtMaxHealthDrop); CurrentHealth, CurrentHealthPercent and HealthMissing read full health in a scoring run (TestScoringRunAPLReadsNoHealth, all four tanks); tools/encounters/sod/*.json parryHaste false (fitted by the same from_logs.py from swing gaps that include parry haste).

### tank_mechanics-20261104-11: Tank mechanics with no test that fails when they break
- severity: low (none known wrong except -6, -7, -8)
- affects: all four tanks, engine
- found: 2026-11-04, master f62bfc74b
- evidence:
  - Warrior: Unbridled Wrath (its rate, its Rage, which hits proc it); Blood Craze; Enrage's +2% a point; Heroic Strike's +138; a real block spending Shield Block's charges.
  - Paladin: Deflection; Improved Judgement and Judgement's 10 s and 6%; Divine Strength; Improved Seals on 20418 and 20414; Consecration's cost and cooldown; Holy Strike's 20 mana.
  - Bear: white-hit Rage (-7); Primal Bite's notes threat (-2); Thick Hide after a mid-fight Defense change; Rage-gain threat in Bear Form.
  - Warden: Lightning Shield not firing on an avoided swing; Stoneskin on a crit or crushing blow; Rockbiter on a dodged Stormstrike.
  - Engine: tank in front (-1); parry haste in a tank run; rolled crush vs the sheet; the block value formula; health invariance of a scoring run (-8); the cast-time no-avoidance rule (spell_outcome.go:673); tools/tankscore has no tests (OwnDebuffs, MakeTank, Score, PairedDiff); OutcomeEnemyMeleeSpecial under Forever's crit (only ClassicCrit is tested).
- proposed fix: add the tests, with the ones for -1, -2 and -7 first.
- needs: worker fix
- fixed (warrior list): 2026-11-04, 9173d087a: sim/warrior/tank_warrior/talents_test.go (Unbridled Wrath, Enrage, Blood Craze, Heroic Strike +138, Shield Block charges on real blocks). Other classes and engine: their workers'.
- fixed (paladin row only): 2026-11-04, 639590a. Deflection, Divine Strength (TestProtectionStatTalents), TestJudgementCooldownAndCost, TestImprovedSealsOnSealOfFury, TestConsecrationAndHolyStrikeCosts; each seen failing on a broken value.
- fixed (bear rows): 2026-11-04, bde6912 (TestBearWhiteHitRage, TestPrimalBite's 2×, TestThickHideDefenseChange, TestBearRageGainThreat); each seen failing on a broken value.
- fixed (Warden part): 2026-11-04, 952a7a4 (`TestForeverWardenAvoidedHits`: no orb on an avoided swing, no Rockbiter threat on an avoided Stormstrike, Stoneskin before the crit and crush multipliers).
- pass 2 (2026-11-04, master b38ff167d): the class rows hold (all named tests pass). Still open: engine part. TestMakeTankInFront and TestScoringRunReadsNoHealth now exist, but no test fails if parry haste stops firing for a MakeTank tank (proposed: DTPS with parryHaste on > off), TestScoringRunReadsNoHealth leaves out the warrior (the only tank with self health costs and a Shield Wall timer), and rolled crush vs the sheet, the block value formula, OwnDebuffs/Score/PairedDiff and Forever-crit OutcomeEnemyMeleeSpecial stay untested. Also TestReckoning has no periodic-crit case (holds today only through the core routing ticks to OnPeriodicDamageTaken, spell_result.go:607-613).
- needs (now): orchestrator fix (engine tests); paladin worker (Reckoning periodic case)
- fixed (engine part): 2026-11-04, 0950b22. TestMakeTankParryHaste, TestScoringRunReadsNoHealth with the warrior (its APL needed autocast and no pull potion for a cooldown to ever read health), TestForeverRolledOutcomesMatchSheet, TestBlockValueFormula, TestForeverBossMeleeAbilityForeverCrit, TestOwnDebuffs/TestMakeTankSetup/TestScore/TestPairedDiff, and TestReckoning's periodic case (20177 says non-periodic, so a periodic crit gives no charge); each seen failing on a broken value.

### tank_mechanics-20261104-1: Every scored tank stands behind the boss: no boss parries or blocks, no parry haste
- severity: high
- affects: engine; all four tanks (tankbench, the tank score, rotopt tank fights)
- found: 2026-11-04, master f62bfc74b
- evidence: none of the 13 `tools/tankbench/req/*.json` set `inFrontOfTarget`, and `tankscore.MakeTank` (tools/tankscore/tankscore.go:90) doesn't either. The boss only parries or blocks the attacker `if InFrontOfTarget` (sim/core/spell_outcome.go:284, 314, 339), so it never parries a tank and parry haste (attack.go:1176) never fires. Encounter runs with `parryHaste` set true and false are identical to the last digit. Yet the tank sheet and docs/tank_sim.md:178 count the boss's dodge and parry "from the front". Measured with tankbench on the default boss, 3,000 iterations, behind → in front:

  | Preset | TPS | SafeDPS | DTPS | Balanced score |
  |---|---|---|---|---|
  | Warrior | 2147 → 1851 | 2205 → 1817 | 308 → 338 | 3590 → 2709 |
  | Paladin | 1228 → 1115 | 1319 → 1183 | 237 → 255 | 2066 → 1733 |
  | Bear | 1551 → 1314 | 1540 → 1220 | 349 → 381 | 2121 → 1366 |
  | Warden | 1391 → 1238 | 1478 → 1280 | 592 → 632 | 1274 → 768 |

  The paladin passes the bear on Balanced. The bear loses most, because all its threat is parryable melee.
  To reproduce: set `players[0].inFrontOfTarget=true` in a copy of a request, then run `tankbench -iters 3000 <req>`. No test checks that a scored tank is in front.
- proposed fix: in MakeTank, set `tank.InFrontOfTarget = true` (`ref/tank_mechanics/patches/engine_01_tank_in_front.patch`). Add a test asserting the boss parries the tank in a scoring run. Then rerun the tank presets, the tank knobs and the docs' numbers. Lands together with -3.
- How sure: high.
- needs: orchestrator fix
- fixed: 2026-11-04, 87f2880
- resolved: 2026-11-04, 87f28803b (pass 2: MakeTank and tankbench set InFrontOfTarget, all 13 requests too; TestMakeTankInFront; parry haste measured firing, bear Balanced DTPS 386 on vs 350 off)

### tank_mechanics-20261104-2: Primal Bite misses Forever's "threat roughly doubled"
- severity: high
- affects: druid (Feral tank, bear)
- found: 2026-11-04, master f62bfc74b
- evidence: Blizzard's development notes, post 4 (client 70170), say "Primal Bite: Threat generated by this ability has been roughly doubled." The saved copy is orchestration/audit/ref/client_update/sources/devnotes_post7_2026-10-26.txt:248, and the ledger logs it as DRU-70170-06 at orchestration/update_ledger.md:1530 ("text only"). The 70755 tooltip says "a high amount of threat", and no client row carries the value (tier 1, notes). The sim sets `ThreatMultiplier: 1` (sim/druid/bear_attacks.go:244). docs/tank_druid.md and docs/tank_sim.md:511 say "no row states" any extra threat; they missed the note. Primal Bite makes 246.5 of 1,550 TPS. Patched to 2× (tankbench, 1,000 iterations):

  | Preset | Balanced score | Max threat | SafeDPS |
  |---|---|---|---|
  | Balanced | 2134 → 2628 (+23%) | 2773 → 3268 | 1546 → 1794 |
  | Max survival | 2012 → 2459 | | |

- proposed fix: 2× under Forever only, as a notes-only value (standards 37). The patch is `ref/tank_mechanics/patches/bear_fix_primal_bite_threat.patch`. Then change bear_rules_test.go:517 from 2.15×damage to 2×2.15. Rerun the Lacerate and filler knobs.
- How sure: sure that the note applies and the sim omits it; moderately sure of exactly 2.0, since the note says "roughly".
- needs: worker fix
- fixed: 2026-11-04, bde6912 (2× under Forever, notes-only; TestPrimalBite asserts 2 × 2.15)
- resolved: 2026-11-04, bde691295 (pass 2: ThreatMultiplier 2.0 under Forever at bear_attacks.go:244, damage ×1, form 2.15 unchanged; the test uses literals)

### tank_mechanics-20261104-3: Anniversary encounter files count parry haste twice
- severity: high (once -1 lands; today parry haste never fires)
- affects: engine; all tanks on tools/encounters/*.json
- found: 2026-11-04, master f62bfc74b
- evidence: `from_logs.py:97-100` takes swingSpeed from the logs' mean swing gap, with parry haste already in it (docs/tank_sim.md:91). All 7 files in tools/encounters/ also set `parryHaste: true`. Chromaggus with the tank in front, parry haste on vs off:

  | Preset | DTPS | Healing needed |
  |---|---|---|
  | Warrior | 536 vs 512 (+4.7%) | 689 vs 649 (+6%) |
  | Paladin | 417 vs 401 | 506 vs 480 |
  | Bear | 629 vs 605 | — |
  | Warden | 970 vs 935 | — |

- proposed fix: set `parryHaste: false` in the 7 files (`ref/tank_mechanics/patches/engine_02_encounters_no_double_parry_haste.patch`). A better fix keeps the sim's parry haste and takes the logged tanks' parry haste out of the swing gap, but that needs the logs' parry counts.
- How sure: high.
- needs: orchestrator fix
- fixed: 2026-11-04, 87f2880
- resolved: 2026-11-04, 87f28803b (pass 2: all 7 tools/encounters/*.json parryHaste false, from_logs.py writes false; the SoD copies are in -14)

### tank_mechanics-20261104-4: Revenge's attack power ratio, 0.28 (log fit) vs 0.25 (70707 tooltip)
- severity: medium (high if 0.25 is confirmed)
- affects: warrior (Protection)
- found: 2026-11-04, master f62bfc74b
- evidence: the sim uses 0.28 (sim/warrior/revenge.go:183). Call 3 in docs/tank_warrior.md compares this with the tooltip's 22.5%. That is the Defensive Stance figure: 25% before the stance's 10% damage cut. wowsims/forever d7a38ae now uses 0.25 on every rank (tools/database/overrides/spell_overrides.go, about line 508). Their source is the tooltip, still to be measured. They dropped the 0.28 fit because its logs predate 70707. Under our tiers a tooltip alone doesn't change a value. At 0.25: Balanced 3590 → 3482 (−3.0%), Max survival −5.6%, TPS −2.7% (tankbench). Revenge is 32% of Balanced's TPS.
- proposed fix: a level 60 Revenge log taken after 70707 settles it. Until then, fix Call 3's framing (`ref/tank_mechanics/patches/warrior_revenge_ap_call_doc.patch`).
- How sure: unsure which value is right.
- needs: user ruling
- ruling: 2026-11-04, user: use 0.25 for now, with a to-verify note (a level 60 Revenge log after 70707 settles it). Worker fix: warrior.
- fixed: 2026-11-04, 53709995e: 0.25 on every rank, to verify in code and docs/tank_warrior.md.
- resolved: 2026-11-04, 53709995e (pass 2: RevengeAPCoefficientForever 0.25, revenge.go:15-33, to-verify note in code and docs/tank_warrior.md row 34 and Call 3; only rank 5 under Forever)

### tank_mechanics-20261104-5: Rockbiter's flat threat on Stormstrike: 72 × weapon speed or 72
- severity: medium
- affects: shaman (Warden)
- found: 2026-11-04, master f62bfc74b
- evidence: sim/shaman/rockbiter_weapon.go:62-71 gives every Stormstrike 72 × 2.9 = 209 threat, ×1.302 with Spirit Weapons. The designer note on 20871 says the points are "desired threat per second" (tier 1). That fits ×speed on white swings, but the passive also procs on specials (ProcTypeMask 20). wowsims/forever (weapon_imbues.go:277-286) multiplies specials by speed too; wowsims/sod gives them a flat 72. Both are tier 2 readings. A flat 72 would take about 35 TPS, 2.5% of Balanced's 1,387, off the Warden (tankbench, 200 iterations).
- proposed fix: none until a combat log shows one Stormstrike's threat. The patch for the flat case is `ref/tank_mechanics/patches/shaman_02_rockbiter_special_flat_CONDITIONAL.patch`.
- How sure: 50/50 on the mechanic; sure of the size.
- needs: user ruling
- ruling: 2026-11-04, the user left it to the orchestrator: keep 72 × weapon speed on specials too (wowsims/forever's reading, and the sim closest to Forever), with a to-verify note in docs/tank_shaman.md and the code (one Stormstrike's threat in a combat log settles it). Worker fix: shaman, the note only.
- fixed: 2026-11-04, 952a7a4 (to-verify note in sim/shaman/rockbiter_weapon.go) and dd4b82a (docs/tank_shaman.md, call (c)); no value change.
- resolved: 2026-11-04, 952a7a4d4 and dd4b82ad1 (pass 2: note at rockbiter_weapon.go:62-67 and docs/tank_shaman.md:149, value unchanged)

### tank_mechanics-20261104-6: Unbridled Wrath procs on Heroic Strike
- severity: medium
- affects: warrior (Protection; Fury DPS builds that use Heroic Strike too)
- found: 2026-11-04, master f62bfc74b
- evidence: 12322's proc mask is 4, melee auto attacks only (tier 1). The sim's Heroic Strike and Cleave carry the main-hand auto-attack proc flag (sim/warrior/heroic_strike_cleave.go:25, 79). The Unbridled Wrath check (sim/warrior/talents.go:239) looks only at that flag, so every landed Heroic Strike rolls for 1 Rage. wowsims/forever marks Heroic Strike special only. Patched, tankbench: Balanced 3590 → 3570 (−0.6%), Max survival 1953 → 1915 (−1.9%, about 2σ), Max threat 4377 → 4359.
- proposed fix: skip Unbridled Wrath on melee specials under Forever (`ref/tank_mechanics/patches/warrior_unbridled_wrath_heroic_strike.patch`). Rerun the DPS Report's Fury rows after.
- How sure: fairly sure.
- needs: worker fix
- fixed: 2026-11-04, 9173d087a: auto attacks only; test TestForeverTankUnbridledWrath. DPS presets (20k it.): Fury 1136.0 → 1134.9 (-0.10%), Fury 2H 1151.0 → 1150.8, Arms 1088.0 → 1088.3 (noise); TestP1DPSWarrior .results 1557.45 → 1553.30.
- resolved: 2026-11-04, 9173d087a (pass 2: talents.go:243 skips ProcMaskMeleeSpecial under Forever; TestForeverTankUnbridledWrath would fail on the old proc)

### tank_mechanics-20261104-7: The bear's white hits use Classic's damage Rage, and crits don't double it
- severity: medium (moves nothing today: the bear's Rage isn't binding)
- affects: druid (Feral tank)
- found: 2026-11-04, master f62bfc74b
- evidence: sim/druid/tank/tank.go:42-46 enables the Rage bar but never calls `EnableForeverNormalizedRage` or `AddCritDealtRageMultiplier(2)`. Only warrior.go:184-185 does. So rage.go:286 falls back to damage×7.5/c. Probe (go test -overlay): two swings give 13.18 Rage at +0 AP and 48.03 at +3000 AP. Standards 21 (settled) says 11.25 a swing, flat, and the Oct 8 notes double it on a crit. Scores today, base → fixed: Balanced 2134 → 2128, Max survival 2012 → 2006, Max threat 2101 → 2102 (noise). The "Rage from the swing ... right" row in docs/tank_druid.md is wrong.
- proposed fix: `ref/tank_mechanics/patches/bear_fix_bear_white_rage.patch` (in Initialize, under Forever). Add a test asserting 11.25 a swing, 22.5 on a crit, independent of AP.
- How sure: sure.
- needs: worker fix
- fixed: 2026-11-04, bde6912 (TestBearWhiteHitRage: 11.25 a landed swing, 22.5 a crit, at +0 and +3,000 AP)
- resolved: 2026-11-04, bde691295 (pass 2: only the Feral tank enables it, tank.go:74-80; auto swings only, 11.25 landed incl. glancing/blocked, 22.5 crit, 0 on miss/dodge/parry; Blood Frenzy +5 on top per 16958)

### tank_mechanics-20261104-8: Survival cooldowns and Berserking read health in a scoring run
- severity: low (latent: no preset hits it)
- affects: engine; tank score (paladin, bear and Warden APLs use autocastOtherCooldowns)
- found: 2026-11-04, master f62bfc74b
- evidence: two places hold survival cooldowns while `CurrentHealthPercent() > HpPercentForDefensives` (default 0): major_cooldown.go:140-141 and apl_idle.go:348. In a scoring run those cooldowns fire only once the tank's own health hits 0, which depends on the bench's HPS. Troll Berserking reads health too (racials.go:172). That breaks the rule that nothing in a scoring run reads the tank's health. Measured: Warden and warrior Balanced at 300, 600 and 5,000 HPS, and a bear with the 990815 trinket at 200, 400 and 3,000, all give identical SafeDPS, DTPS and healing needed. A gear search with an armor, dodge or health on-use item, a health potion, or Shield Wall on autocast could hit it.
- proposed fix: when `ts.replay`, skip survival cooldowns or force `HpPercentForDefensives = -1`, and give Berserking a fixed percentage. Or have tankscore reject requests that carry them. Add a test that SafeDPS and DTPS are equal at two HPS.
- needs: orchestrator fix
- fixed: 2026-11-04, 87f2880
- resolved: 2026-11-04, 87f28803b (pass 2: survivalCooldownDue false in a replay, Berserking fixed 10%; Onyxia at 300 and 5000 HPS gives identical warrior SafeDPS/DTPS/TMI; latent APL holes in -14)

### tank_mechanics-20261104-9: Reckoning ignores spell crits
- severity: low
- affects: paladin (Protection)
- found: 2026-11-04, master f62bfc74b
- evidence: 20177's proc mask is 0x202A8, which includes magic damage taken (tier 1). The sim reacts only to melee or ranged crits (sim/paladin/talents.go:233). Far under 1% at the settled 0.6% boss crit.
- proposed fix: drop the ProcMaskMeleeOrRanged gate for the crit branch only.
- needs: worker fix
- fixed: 2026-11-04, e9751cc. 20177's tooltip: "being the victim of a non-periodic critical strike"; TestReckoning fails on the old gate.
- resolved: 2026-11-04, e9751cc (pass 2: crits of any non-periodic kind charge, ticks route to OnPeriodicDamageTaken; TestReckoning fails with the old gate restored)

### tank_mechanics-20261104-10: Tank class docs: wrong tier labels and stale rows
- severity: low
- affects: shaman, druid, paladin, warrior docs (no sim change)
- found: 2026-11-04, master f62bfc74b
- evidence: in docs/tank_shaman.md, Earth Shock's ×2 threat rests on "wowsims/classic, vmangos". vmangos is never a source; wowsims/sod earth_shock.go:54 is a tier 2 source for the same 2. Rockbiter's white-hit threat is labeled tier 2, but 20871's designer note makes it tier 1. In docs/tank_druid.md, the white-hit Rage row is wrong (-7), and Swipe's 3% AP is tier 1 (post 4 notes, line 241), not tier 2. In orchestration/round_70753/paladin.md:114, Templar's Bulwark's dispel is "notes only", but 1311015 has DispelType 1 (Magic) in 70755+hf. In docs/tank_warrior.md, Call 3's framing is off (-4).
- proposed fix: `ref/tank_mechanics/patches/shaman_01_docs_tiers.patch` and `warrior_revenge_ap_call_doc.patch`; the druid and paladin lines by hand.
- needs: worker fix
- fixed (druid rows): 2026-11-04, bde6912 (white-hit Rage row, Swipe's 3% as tier 1, Primal Bite's notes threat in the table and calls)
- fixed (warrior row, Call 3): 2026-11-04, 53709995e. Shaman, druid and paladin rows are their workers'.
- fixed (paladin line only): 2026-11-04, 3f9c055. round_70753/paladin.md now says Templar's Bulwark's Magic is client data; docs/tank_paladin.md already had it as N, C.
- fixed (shaman part): 2026-11-04, 952a7a4 (shaman_01_docs_tiers.patch) and dd4b82a (the tier key no longer names vmangos).
- resolved: 2026-11-04, bde691295, 53709995e, 3f9c055, 952a7a4d4, dd4b82ad1 (pass 2: the rows hold; leftover stale labels found in pass 2 are -13)
