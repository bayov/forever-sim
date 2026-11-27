# Findings: Rotations and resources

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### rotations-20261126-3: Might of Shahram stuns the stun-immune raid boss, and since 8f5b59325 a stunned target can't dodge, parry or block
- severity: low (+0.8 to +1.6 DPS, 0.1-0.2%, on the 7 paladin builds that wear Blackblade of Shahram; no rank change)
- affects: paladin__level_60, -duty-crusader, -duty-wisdom, paladin__level_60_sub_holy, -duty-crusader, -duty-wisdom, paladin__level_60_sub_holy_undead (all wear 12592)
- found: 2026-11-26, master ab3fff7007
- evidence: `regress.py` old CLI bfbee78fcca3 vs new ab3fff700700, same new reqs, 3000 it, seed 424242: only these 7 presets moved more than 0.05 DPS (+1.2, +1.6, +1.6, +1.1, +0.8, +1.2, +1.4); every other preset, all warriors included, is unchanged. Cause: 8f5b59325 (warrior leveling) added `if result.Target.PseudoStats.Stunned { return false }` to the dodge, parry and block rolls (sim/core/spell_outcome.go:724, 755, 775). Might of Shahram (sim/common/item_effects.go:541-566) sets `Stunned = true` on every active target with no `StunImmune` check, while the raid boss is `StunImmune` (sim/core/target.go:192, EV-0088) and Forever explosives already skip it (sim/core/consumes.go:1438). So for 5 s per proc the boss stops dodging and parrying the paladin, which can't happen in game.
- proposed fix: in Might of Shahram's ApplyEffects, skip `aoeTarget.StunImmune` targets under `sim.IsForever()` (as consumes.go:1438 does). Expected effect: the 7 builds go back by the amounts above.
- needs: orchestrator fix

### rotations-20261122-1: Shock Ret: Hammer of Wrath and Holy Strike ahead of Judgement and Seal of Command
- severity: low (+2.2 at 180 s, 0.22%)
- affects: paladin__level_60_shock_ret (its duty-crusader and duty-wisdom variants share the order; not measured)
- found: 2026-11-22, master 8b5137c6d
- evidence: `ui/retribution_paladin/apls/level60_shock_ret.apl.json` runs Judgement (line 10), Seal of Command, Holy Shock,
  then `{"action":{"castSpell":{"spellId":{"spellId":24239}}}` (Hammer of Wrath, line 13) and
  `{"action":{"castSpell":{"spellId":{"spellId":10333}}}` (Holy Strike, line 14). After the Nov 19 Two-Handed Weapon
  Specialization hotfix (1c80dafff) Holy Strike gained the most; the sister Ret builds already run it ahead of the seal
  lines (1b4a74f33 did so for Undead). B moves those two lines, in that order, to just after Consecration.
  Paired `ab.py` (20k it, 8 blocks, CLI 8b5137c6d), B-A at 120 / 180 / 300 s:
  | variant | S0 1e6 | S0 2e8 | pooled |
  |---|---|---|---|
  | HoW, HS ahead of Judgement | +1.28 / +2.13 / -0.53 | +1.63 / +2.36 / +0.07 | **+1.46 / +2.25 / -0.23** |
  | HS alone ahead of Judgement (HoW stays) | +2.28 / +3.39 / -2.54 | +2.32 / +3.73 / -1.97 | +2.30 / +3.56 / -2.26 (trade-off) |
  | same, HS gated on mana >= 30% / 50% | | | +3.6 / -4.5 and +1.8 / -12.4 at 180 / 300 (worse) |
  SE 0.25-0.6 a set. A at 180 s: 1017.0. The 300 s cell of the filed variant is flat within noise; medium confidence.
- proposed fix: in level60_shock_ret.apl.json (and the two duty copies, after an A/B on each) move Hammer of Wrath then
  Holy Strike to right after Consecration. Fits the open "re-time the paladin rotations" item, which 1b4a74f33 triggered.
- needs: orchestrator fix

### rotations-20261112-1: Dagger and Seal Fate rogues: Instant Poison on the main hand, Deadly Poison on the off hand
- severity: high (Seal Fate Daggers +16.3 at 180 s, 1.8%; Mutilate +14.3, 1.65%; Seal Fate Swords +9.8, 1.05%; Subtlety +2.4, 0.28%)
- affects: rogue__mutilate, rogue__seal-fate-daggers, rogue__seal-fate-swords, rogue__subtlety. Not the two Combat builds (swap loses 2.3-3.2).
- found: 2026-11-12, master a0b4d3be9
- evidence: every rogue req carries `"rogue":{"options":{"mainHandPoison":"DeadlyPoison","offHandPoison":"InstantPoison"}}`
  (e.g. `tools/dps_report/req/rogue__mutilate.json:1101`), from the shared `ui/rogue/presets.ts:345-346`. B swaps the two.
  Paired `ab.py --req-b` (20k it, 8 blocks, CLI a0b4d3be9859), B-A at 120 / 180 / 300 s:
  | preset | S0 1e6 | S0 2e8 | pooled | A at 180 |
  |---|---|---|---|---|
  | mutilate | +14.91 / +14.47 / +14.13 | +15.02 / +14.18 / +13.77 | **+14.97 / +14.33 / +13.95** | 865.0 |
  | seal-fate-daggers | +16.41 / +16.26 / +15.70 | +16.63 / +16.28 / +15.78 | **+16.52 / +16.27 / +15.74** | 903.6 |
  | seal-fate-swords | +10.34 / +9.60 / +9.23 | +9.98 / +10.04 / +9.68 | **+10.16 / +9.82 / +9.46** | 939.0 |
  | subtlety | +2.08 / +2.27 / +2.04 | +2.44 / +2.61 / +2.65 | **+2.26 / +2.44 / +2.35** | 862.6 |
  | combat-daggers | -1.96 / -2.72 / -3.28 | -2.57 / -2.73 / -3.19 | -2.27 / -2.73 / -3.24 | 936.2 |
  | combat-swords | -2.63 / -2.81 / -3.17 | - | - | - |
  SE 0.2-0.7 a set. Instant Poison on both hands loses on every build (-13 to -20; Seal Fate Daggers -1.5).
  Part of this predates the change: the old CLI (71994fca6) gives the swap +7.4 (Mutilate) and +6.3 (Seal Fate
  Daggers) at 180 s, set 1. This master's Windfury Totem rolls on main hand abilities (a0b4d3be9) and Venom's
  +20% reaching Instant Poison (e2322baa1) roughly double it. High confidence on the sign; numbers as measured.
- proposed fix: give Mutilate, both Seal Fate builds and Subtlety their own options with `mainHandPoison: InstantPoison`,
  `offHandPoison: DeadlyPoison` (presets.ts and the four req files); Combat keeps the current pairing.
- needs: orchestrator fix

## Resolved

### rotations-20261127-1: Duty-curse warlocks: Talisman of Ascendance fires before the duty curse, which eats a charge
- severity: medium (+4.6 at 180 s, 0.49%, on the Affliction duty builds; +2.2 to +2.6 Demonology; +1.4 Destruction Recklessness)
- affects: warlock__affliction-duty-curse, -duty-recklessness, warlock__demonology-duty-curse, -duty-recklessness, warlock__destruction-duty-recklessness. Destruction CoE and both Destruction Fire duty builds gain under the bar (+0.3 to +0.6, optional).
- found: 2026-11-27, master 5ab589d858
- evidence: since 413f06cb0 every APL cast eats one of Talisman's 6 charges (sim/common/item_effects.go:4756-4790). The duty
  APLs fire the Talisman from `autocastOtherCooldowns` (priority line 0) at t=0, so the first-GCD duty curse (CoE 1311680 /
  CoR, `ui/warlock/apls/level60_affliction_duty_curse.apl.json:9`, demonology :11, destruction :9) takes a charge and every
  later spell gets one stack less. B inserts `{"action":{"castSpell":{"spellId":{"itemId":22678}}}}` right after the duty-curse
  line (the curse stays on the first GCD, so the settled CoE ruling holds). Paired `ab.py` (20k it, 8 blocks, CLI 5ab589d85898),
  B-A at 120 / 180 / 300 s, pooled over S0 1e6 and 2e8 (SE 0.03-0.6 a set):
  | preset | S0 1e6 | S0 2e8 | pooled | A at 180 |
  |---|---|---|---|---|
  | affliction-duty-curse | +4.73 / +4.70 / +3.43 | +4.57 / +4.56 / +3.36 | **+4.65 / +4.63 / +3.40** | 948.0 |
  | affliction-duty-recklessness | +4.37 / +4.59 / +3.56 | +4.24 / +4.44 / +3.42 | **+4.31 / +4.52 / +3.49** | 942.5 |
  | demonology-duty-curse | +3.85 / +1.96 / +1.05 | +2.70 / +2.46 / +1.52 | **+3.28 / +2.21 / +1.29** | 864.7 |
  | demonology-duty-recklessness | +3.59 / +2.69 / +1.79 | +2.81 / +2.51 / +1.34 | **+3.20 / +2.60 / +1.57** | 859.2 |
  | destruction-duty-recklessness | +1.83 / +1.60 / -0.07 | +1.71 / +1.23 / +0.31 | **+1.77 / +1.42 / +0.12** | 851.1 |
  | destruction-duty-curse | | | +1.76 / +0.56 / +0.37 (optional) | 856.9 |
  | destruction_fire duty curse / reck | | | +0.15 / +0.38 / +0.49 and -0.18 / +0.41 / +0.37 (optional) | 816 / 811 |
  Not for the non-duty builds: Talisman held for Shadow Bolt loses 3.4 on warlock__affliction, and on Balance (held for Starfire
  -3.1, after Faerie Fire on duty-ff -1.25). Caveat: the warlock Phase 0 lists on master no longer wear 22678 (pass 85: the APL
  holds moved to 13968 / 10455) but these reqs still do; if the next regeneration drops the Talisman, this finding is moot.
- proposed fix: while the warlock duty reqs wear 22678, add the Talisman line right after the duty-curse line in the 5 builds above
  (req rotation and APL file). Otherwise close it when the reqs drop the Talisman.
- needs: orchestrator fix
- resolved: 2026-11-27, 94d6cc79f (moot: the Phase 0 re-gear took Talisman of Ascendance 22678 off every report request; no req or APL names it now)

### rotations-20261126-1: Balance APLs still pre-pull and gate Elune's Light on Chained Essence of Eranikus (10455) after the re-rank moved Balance to Talisman of Ascendance
- severity: low (correctness: dead lines naming a trinket the build no longer wears; 0.0 DPS to clean up)
- affects: druid__balance, druid__balance-wrath and their -duty-ff builds (ui/balance_druid/apls/level60{,_ff,_wrath,_wrath_ff}.apl.json lines 5 and 11, and the four req rotations)
- found: 2026-11-26, master bfbee78fc
- evidence: a60eb2ca1 swaps trinket 10455 for 22678 (Talisman of Ascendance) in the four Balance reqs, but the APLs keep the pre-pull
  `{"action":{"castSpell":{"spellId":{"itemId":10455}}},"doAtValue":{"const":{"val":"-0.1s"}}}` and Elune's Light's (1259799) condition on
  `auraIsActive/spellTimeToReady/spellUsesLostByDelay` of itemId 10455. Paired ab.py, 20k it, 8 blocks, CLI bfbee78fcca3, A = the req as is (802.6 at 180 s), B-A at 120 / 180 / 300 s:
  - B3, drop the 10455 pre-pull and the Elune's Light condition (cast on cooldown): +0.00 / +0.00 / +0.00 on balance and balance-wrath (set 1): identical to A, so the lines are dead.
  - B2, drop the pre-pull, gate Elune's Light on 22678: +0.00 / -0.03 / +0.07 (set 1).
  - B1, 10455 -> 22678 everywhere (pre-pull the Talisman): set 1 -1.76 / -1.21 / -0.63, set 2 -1.73 / -1.17 / -0.73, pooled -1.75 / -1.19 / -0.68. Don't pre-pull it; autocast already fires it on cooldown (as rotations-20261125-1 on the warlock Talisman found).
- proposed fix: B3 in the four APLs and four req rotations: delete the 10455 pre-pull line and the Elune's Light line's condition (rewrite its note: on cooldown). No re-sim needed (bit-identical).
- needs: orchestrator fix
- resolved: 2026-11-26, c1ebe2d79 (pre-pull 10455 line dropped, Elune's Light ungated, in all four APLs and the four req rotations)

### rotations-20261126-2: Fury DW APLs still use Blackhand's Breadth (13965) after the re-rank moved Fury DW to Heart of Wyrmthalak and Darkmoon Card: Maelstrom
- severity: low (correctness: dead line, 0.0 DPS; the build now wears no on-use trinket)
- affects: warrior__fury, warrior__fury-duty-battle-shout, warrior__fury-duty-sunder and their race rows (ui/warrior/apls/level60_fury.apl.json:15, level60_fury_battle_shout.apl.json:18, level60_fury_sunder.apl.json:16, and the three req rotations)
- found: 2026-11-26, master ab3fff7007
- evidence: f79e810e1 changed Fury DW trinkets 13965+19289 -> 22321+19289 (req equipment items[12..13]); the APL line `{"action":{"castSpell":{"spellId":{"itemId":13965}}},"notes":"Blackhand's Breadth on cooldown: its mark beats holding it for the last 30 s."}` now names an item not worn. Arms and Fury 2H still wear 13965 with no line (autocast), which is fine. Per the Phase 0 trinkets open item, a build with no on-use trinket shouldn't keep a trinket line.
- proposed fix: delete that line from the three APLs and their req rotations.
- needs: orchestrator fix
- resolved: 2026-11-26, 832070749 (Phase 0 prices round 2 put Blackhand's Breadth 13965 back on Fury DW, req trinkets 19289+13965, so the line names a worn on-use trinket again)

### rotations-20261125-1: Feral cat APLs still cast the EZ-Thro Dark Bomb (260817) after the presets moved to the Goblin Land Mine
- severity: low (correctness: a dead line naming an item the build no longer carries; +0.5 DPS at 180 s once fixed)
- affects: druid__feral-cat, druid__feral-cat-duty-ff (ui/feral_druid/apls/level60.apl.json:4, level60_ff.apl.json:4, and the embedded req rotations)
- found: 2026-11-25, master 85691655a
- evidence: 22d2cf952 (EV-0088) moved `ui/feral_druid/presets.ts` to `fillerExplosive: ExplosiveGoblinLandMine`, but line 4 of both APLs is still
  `{"action":{"castSpell":{"spellId":{"itemId":260817}},"condition":{"and":{"vals":[{"cmp":{"lhs":{"currentEnergy":{}},"op":"OpLe","rhs":{"const":{"val":"60"}}}},{"cmp":{"lhs":{"autoTimeToNext":{"autoType":"MainHand"}},"op":"OpGe",...autoSwingTime - 0.4s}}]}}}`
  with the note "its 1 sec cast restarts the swing timer", which EV-0088 retired. With the Land Mine equipped that line is dead and the mine
  goes out as an ungated major cooldown. Paired ab.py on the req with fillerExplosive set to ExplosiveGoblinLandMine (consumables-20261125-3's fix),
  20k it, 8 blocks, A = the APL as is; B-A at 120 / 180 / 300 s:
  - feral-cat, itemId 4395 + energy gate only (swing term dropped): 1e6 +0.47 / +0.53 / +0.01, 2e8 +0.51 / +0.62 / -0.03, pooled +0.49 / +0.58 / -0.01
  - feral-cat, itemId 4395 keeping both gates: pooled +0.41 / +0.52 / -0.02
  - feral-cat-duty-ff, 4395 + energy gate: 1e6 +0.49 / +0.39 / +0.05, 2e8 +0.75 / +0.65 / -0.12, pooled +0.62 / +0.52 / -0.04
  - baseline at 180 s (A): feral-cat 1061.5, duty-ff 1056.4
- proposed fix: in both APLs (and the req rotations when they are regenerated with the Land Mine), set the line's itemId to 4395, drop the
  `autoTimeToNext >= autoSwingTime - 0.4s` term, keep `currentEnergy <= 60`, and rewrite the note. Land with consumables-20261125-3.
- needs: orchestrator fix
- resolved: 2026-11-26, 35a2e9dca (variant): both APLs cast the Land Mine (4395) and keep the `autoTimeToNext >= autoSwingTime - 0.4s` term with the energy gate; this pass measured that form at pooled +0.52 at 180 s against +0.58 without the term, within noise. The req rotations carry it too.

### rotations-20261125-1: Two of 78802e6e's live on-use trinket lines lose DPS: the warlock Talisman of Ascendance pre-pull and Fury DW's held Blackhand's Breadth
- severity: medium (Affliction -0.5%, Fury DW -0.3%; no place in the ranking changes)
- affects: all 16 warlock builds, warrior__fury and its two duty variants
- found: 2026-11-25, phase0-step4-publish 01cc70b80
- evidence: 20k iterations, seed 1, 180 s, CLI at 01cc70b80, one line changed per variant. Warlocks, Talisman pre-pull at -0.1s against none (autocast still casts it on cooldown, 3.33 a fight either way; the pre-pull holds the shared burst-trinket category so Weakness Analyzer goes later): Affliction 960.87 / 965.79, Destruction 877.59 / 881.25, Demonology 892.02 / 894.19, Fire 833.57 / 835.28, Shadow Bolt 808.17 / 810.11; the pre-pull at 0s is the same 960.87. Swapping the pre-pull to the other trinket: 964.92 (Affliction). Fury DW, Blackhand's Breadth (5 min) held for the last 30 s against on cooldown: 1005.64 / 1008.99, Sunder 971.23 / 974.55, Battle Shout 1000.87 / 1002.87. Other trinket lines checked made dead (pre-78802e6e behaviour): Arms, Enhancement imp-gw and water-shield unchanged to 0.01.
- proposed fix: drop the warlocks' Talisman pre-pull line; cast Blackhand's Breadth on cooldown in the three Fury DW APLs.
- needs: orchestrator fix
- resolved: 2026-11-25, 67fe28cf0 (the fix above in requests, race rows and ui APLs; re-simmed and published)

### rotations-20261115-1: Arms: hold the main-hand abilities while Windfury Totem's 100 ms cooldown runs
- severity: low (+2.5 DPS at 180 s, 0.24%; Dragonkin +1.6)
- affects: warrior__arms, warrior__arms_dragonkin (+ the two Arms duty builds, same lines, not run). Not Fury DW (0.0) or Fury 2H (+0.4, under the bar).
- found: 2026-11-15, master 70ea588bb (pass 66, the user's Windfury timing request)
- evidence: Mortal Strike, Slam, Rend, Hamstring, Execute and Spearing Strike roll Windfury Totem under Forever
  (`sim/core/windfury_totem_procs_forever.go`), and the totem's ICD is 100 ms (`sim/core/buffs.go`, standards 17), so an
  ability that lands just after a procced swing loses its roll. Variant B appends to each of those castSpell lines
  (21553, 11605, 11574, 7373, 20662, 1310222) the gate
  `{"and":{"vals":[{"auraIcdIsReadyWithReactionTime":{"auraId":{"spellId":10610}}},{"or":{"vals":[{"not":{"val":{"auraIsActive":{"auraId":{"spellId":10610}}}}},{"cmp":{"lhs":{"auraRemainingTime":{"auraId":{"spellId":10610}}},"op":"OpLe","rhs":{"const":{"val":"900ms"}}}}]}}]}}`
  (the ICD value alone reads "ready" while the buff is younger than the reaction time, hence the remaining-time clause; the
  buff alone drops early once its charges are used). Made by `tools/audit/rotations/wf_gate.py REQ OUT`.
  Lines: `ui/warrior/apls/level60_arms.apl.json:18,19,21,22,25`, `_battle_shout:21,22,24,25,28`, `_dragonkin:18,19,21,22,23,26`,
  `_sunder:19,20,22,23,26`. Paired `ab.py` (20k it, 8 blocks, CLI 70ea588bb116), B-A at 120 / 180 / 300 s:
  | preset | S0 1e6 | S0 2e8 | pooled | A at 180 |
  |---|---|---|---|---|
  | arms | +1.39 / +2.42 / +2.30 | +1.83 / +2.57 / +2.38 | **+1.61 / +2.50 / +2.34** | 1042.9 |
  | arms_dragonkin | - / +1.60 / - | - | - | 1100.8 |
  | fury_2h | +0.35 / +0.30 / +0.28 | +0.08 / +0.42 / +0.33 | +0.22 / +0.36 / +0.31 | 1033.0 |
  SE 0.14-0.27 a set. Gating 100 ms after every main-hand swing (not only procs) gains about the same on Arms (+2.27 at 180,
  set 1) and loses on Fury DW (-0.58). Which of the six lines carries the gain was not split out. Rests on the unverified
  100 ms ICD (report assumption wf_totem_icd); with the Era 1.5 s it would be worth more.
- proposed fix: add the gate above to those lines in the four `level60_arms*.apl.json` files and to the two Arms req rotations
  (warrior__arms, warrior__arms_dragonkin; the duty reqs too if they embed them).
- needs: orchestrator fix
- resolved: 2026-11-15, a8ea43990 (variant): the landed gate is the swing form, `autoTimeToNext(MH) <= autoSwingTime(MH) - 0.1s`,
  on Mortal Strike, Overpower, Rend, Execute and Hamstring in the three `level60_arms*.apl.json` and the three Arms reqs (pass 66
  measured that form at +2.27 on Arms). Dragonkin was left out; pass 67 re-measured the same gate there under the Nov 13 rule
  (every main-hand melee hit rolls the totem): -0.15 / -0.38 / -0.84 at 120 / 180 / 300 s (set 1, A 1126.1), so leaving it out is right.

### rotations-20261105-1: Feral cat: Ferocious Bite energy gate 45 -> 65, now that extra Energy hits harder (hotfix 113403)
- severity: low (+1.9 DPS at 180 s, 0.16%; +1.1 on the FF duty)
- affects: druid__feral-cat, druid__feral-cat-duty-ff (ui/feral_druid/apls/level60.apl.json:10, level60_ff.apl.json:11)
- found: 2026-11-05, master a07941531
- evidence: ff56f8ddc (`sim/druid/ferocious_bite.go`, hotfix push 113403) raised rank 6's damage per extra
  Energy 2.7 -> 3.3. The 5-CP Ferocious Bite line still fires at `{"cmp":{"lhs":{"currentEnergy":{}},"op":"OpGe","rhs":{"const":{"val":"45"}}}}`.
  Paired `ab.py` (20k it, 8 blocks, shared CLI a0794153163a), B-A at 120/180/300 s:
  | gate | S0 1e6 | S0 2e8 | pooled |
  |---|---|---|---|
  | 35 | -2.61 / -2.92 / -3.41 | - | - |
  | 55 | +1.25 / +1.18 / +2.10 | +1.40 / +1.22 / +1.92 | +1.33 / +1.20 / +2.01 |
  | 65 | +1.63 / +1.86 / +3.28 | +1.80 / +1.96 / +3.12 | +1.72 / +1.91 / +3.20 |
  | 75 | -0.42 / +0.19 / +1.85 | - | - |
  | 65, FF duty | +0.70 / +0.89 / +2.06 | +1.18 / +1.33 / +2.06 | +0.94 / +1.11 / +2.06 |
  Baseline at 180 s: 1169.1 (feral-cat), 1164.5 (duty-ff). Gains at every length on both sets; 65 is the peak of
  the measured grid (SE about 0.2-0.4 a set).
- update 2026-11-12 (pass 61, master 8306f1d52): c380bbaef gives Rake's direct hit 6.5% of attack power (feral +29.3
  at 180 s in the baseline). The gate is still 45 on master and 65 now gains more. Same variant, 20k it, 8 blocks,
  CLI 8306f1d52875, B-A at 120 / 180 / 300 s:
  feral-cat: set 1 +2.07 / +2.97 / +4.81, set 2 +2.17 / +2.50 / +4.36, **pooled +2.12 / +2.74 / +4.59** (A 1153.1 at 180).
  duty-ff: set 1 +1.83 / +2.19 / +3.99, set 2 +1.77 / +1.79 / +3.76, **pooled +1.80 / +1.99 / +3.88** (A 1148.3).
  Gate 75 (set 1): +0.28 / +1.28 / +4.16, so 65 stays the pick. Severity stays low (0.24%).
- update 2026-11-15 (pass 66, master 70ea588bb): 0cc2c3c8b refunds 80% of an avoided Bite's 35 Energy. Gate 65 still gains,
  same variant, 20k it, CLI 70ea588bb116, B-A at 120 / 180 / 300 s: feral-cat set 1 +1.80 / +2.51 / +4.17, set 2 +2.05 / +2.47 / +4.53,
  **pooled +1.93 / +2.49 / +4.35** (A 1155.8); duty-ff set 1 +1.29 / +1.71 / +3.54, set 2 +1.50 / +1.63 / +3.74, **pooled +1.40 /
  +1.67 / +3.64** (A 1150.7). Gate 75 (set 1): -0.05 / +0.44 / +3.28, so 65 stays the pick.
- proposed fix: in both APLs (and the two req rotations), the 5-CP Ferocious Bite line's `"val":"45"` -> `"val":"65"`.
- needs: orchestrator fix
- resolved: 2026-11-15, f9990ec4e (Bite at 65 Energy in both Feral APLs, plus a totem cooldown gate).

### rotations-20261109-1: Fury 2H: drop the 90-rage Heroic Strike line (Arms no longer: Windfury Totem now rolls on Heroic Strike)
- severity: low (now +2.0 DPS at 180 s on Fury 2H, 0.19%; was +3.5 on Arms and Fury 2H)
- affects: warrior__arms (+ -duty-battle-shout, -duty-sunder, _dragonkin), warrior__fury_2h (+ its two duty builds). Not Fury DW.
- found: 2026-11-09, master 66a243cf8
- evidence: 0ce9dbf40 (`sim/core/buffs.go:2166`, user ruling 2026-11-08) makes Windfury Totem proc from
  `ProcMaskMeleeMHAuto` only under Forever, so a queued Heroic Strike now also costs that swing's 20% totem
  chance. Arms WF Totem attacks fell 27.6 -> 9.2 a fight (2k it, 180 s) and Arms lost 199 DPS, Fury 2H 189
  (baseline diff, all from this commit). The line is `{"castSpell":{"spellId":{"rank":8,"spellId":11567,"tag":1}},
  "condition":{... currentRage OpGe "90", not isExecutePhase E20}}` at `ui/warrior/apls/level60_arms.apl.json:24`
  (also `_battle_shout:27`, `_dragonkin:25`, `_sunder:25`, `level60_fury_2h.apl.json:25`, `_battle_shout:28`,
  `_sunder:26`); req rotation index 16 in warrior__arms.json and warrior__fury_2h.json. Variant B deletes it.
  Paired `ab.py` (20k it, 8 blocks, CLI 66a243cf8a64), B-A at 120 / 180 / 300 s:
  | preset | S0 1e6 | S0 2e8 | pooled | A at 180 |
  |---|---|---|---|---|
  | arms | +2.36 / +3.48 / +2.83 | +2.18 / +3.49 / +2.09 | **+2.27 / +3.49 / +2.46** | 887.5 |
  | fury_2h | +2.19 / +2.88 / +1.58 | +1.43 / +4.13 / +2.12 | **+1.81 / +3.51 / +1.85** | 962.5 |
  SE 0.25-0.66 a set. Raising the gate to 100 gains less (Arms pooled +1.0 / +1.4 / +0.9, Fury 2H +0.9 / +1.3 / +1.5).
  Fury DW keeps its 60 gate: 80 loses 2.8 and 50 loses 1.7 at 180 s (set 1). Duty and Dragonkin builds were not
  run separately; they share the line, so the gain should carry over (estimate).
- update 2026-11-12 (pass 60, master a0b4d3be9): a0b4d3be9 reverses the premise; Heroic Strike now rolls Windfury
  Totem again. Re-measured, same variant (delete req rotation index 16), 20k it, both seed sets, B-A at 120 / 180 / 300 s:
  Arms pooled -1.24 / -4.11 / -5.92 (A 1043.2 at 180): keep the line on Arms, Arms_dragonkin and the Arms duties.
  Fury 2H pooled +0.86 / +1.98 / +0.40 (set 1 +0.78 / +1.78 / +0.15, set 2 +0.94 / +2.17 / +0.65; A 1033.2): still
  just over the bar, no loss at any length. Lowering the gate loses on both (Arms 75 / 60 / 50: -5.1 / -19.4 / -31.5
  at 180; Fury 2H -6.8 / -17.8 / -31.5). Fury DW's 60 gate holds (40 -7.6, 50 -3.2, 75 -0.4 at 180, pooled).
- proposed fix (narrowed 2026-11-12): delete the line from the three Fury 2H APLs (`level60_fury_2h*.apl.json`) and
  warrior__fury_2h's req rotation only; leave Arms as is.
- original proposed fix: delete the 90-rage Heroic Strike line from the seven Arms and Fury 2H APLs and their req rotations.
  If 0ce9dbf40's ruling is reversed (the commit says Heroic Strike wasn't checked in the logs), restore it.
- needs: orchestrator fix
- resolved (withdrawn): 2026-11-15, pass 67, master 25ce81de7: under the Nov 13 rule (Windfury Totem on every main-hand
  melee hit) deleting the line now loses on Fury 2H: -2.23 / -2.99 / -4.09 at 120 / 180 / 300 s (set 1, SE 0.3-0.6, A 1104.5).
  Keep it on every build.

### rotations-20261108-1: Enhancement Mana Tide and Water Shield: Flametongue Weapon beats Windfury Weapon now that our own Windfury Totem is the group's weapon totem
- severity: medium (+5.6 DPS at 180 s on Mana Tide, 0.54%; +3.4 on Water Shield, 0.31%)
- affects: shaman__enhancement-mana-tide, shaman__enhancement-water-shield. Not Level 60 or Imp GW (see evidence).
- found: 2026-11-08, master 9ba00c2dc
- evidence: since 9ba00c2dc the presets have no raid Flametongue Totem, and our Windfury Weapon still turns off
  the Windfury Totem we drop ourselves. Variant B sets only
  `"enhancementShaman":{"options":{"shamanImbue":"FlametongueWeapon"}}` (req line
  `"shamanImbue": "WindfuryWeapon"`, mana-tide.json:748, water-shield.json:714; ui/enhancement_shaman/presets.ts:565
  `PresetBuildOptions`, shared by all four builds). A quick 2k-it check confirms B gets Flametongue Weapon hits
  (16342) and Windfury Totem attacks (tag 3). Paired `tools/audit/rotations/ab_req.py` (20k it, 8 blocks, shared CLI
  9ba00c2dc473), B-A at 120 / 180 / 300 s:
  | preset | S0 1e6 | S0 2e8 | pooled | A at 180 |
  |---|---|---|---|---|
  | mana-tide | +5.69 / +5.60 / +3.05 | +6.05 / +5.61 / +3.17 | **+5.87 / +5.61 / +3.11** | 1041.9 |
  | water-shield | +3.73 / +3.45 / +1.22 | +3.84 / +3.40 / +1.77 | **+3.79 / +3.43 / +1.50** | 1098.3 |
  | level60 | -0.45 / -1.23 / -4.15 | +1.80 / -0.88 / -3.71 | +0.68 / -1.06 / -3.93 | 1147.1 |
  | imp-gw | +2.32 / +1.52 / -2.70 | +2.06 / -0.35 / -3.13 | +2.19 / +0.59 / -2.92 | 1131.4 |
  SE 0.3-1.0 a set. Mana Tide and Water Shield gain at every length on both sets; Level 60 loses, Imp GW is a
  fight-length trade-off (not a finding). Rotations unchanged (not re-searched under Flametongue Weapon).
  Reproduce: `jq '.raid.parties[0].players[0].enhancementShaman.options.shamanImbue="FlametongueWeapon"' req > B.json;
  ab_req.py CLI req B.json --s0 1000000` (and 200000000).
- proposed fix: Mana Tide and Water Shield builds take `shamanImbue: WeaponImbue.FlametongueWeapon` (presets.ts:
  a per-build options object in place of `PresetBuildOptions` for those two; their two req files; mksettings
  `spec_enh.go:61` if it builds them). Level 60 and Imp GW keep Windfury Weapon. Also worth a rotopt re-search of
  the two under Flametongue Weapon (orchestrator's call; this pass doesn't search).
- needs: orchestrator fix

- resolved: 2026-11-09, 0ce9dbf40 (withdrawn, not applied: with Windfury Totem procs from white swings only, Flametongue Weapon now loses 28.4 on Mana Tide and 32.2 on Water Shield at 180 s, pooled over both seed sets; the presets keep Windfury Weapon)

### rotations-20261101-1: Frost mage: a fourth Arcane Blast stack above 70% mana, now that a stack costs 100% (was 175%)
- severity: low (+1.8 DPS at 180 s, 0.2%)
- affects: mage__frost (ui/mage/apls/level60_frost.apl.json)
- found: 2026-11-01, master c23ef0440
- evidence: the October 30 hotfix (b79e033bc, `sim/mage/arcane_blast.go`) made each Arcane Blast stack
  12% of Intellect, +100% mana per stack (was 15%, 175%) and 18 sec. The Frost APL still caps the
  stacks at 3 (line 9, `"auraNumStacks" ... "OpLt" ... "3"`). Plain cap 4 is a trade-off (+4.6 / +2.7 /
  -15.7 at 120/180/300 s, S0 = 1e6: it runs out of mana in long fights). Gating the fourth stack on mana
  holds at every length. Variant: insert a copy of line 9 just above it, with the stack cap `"4"` and one
  more condition `{"cmp":{"lhs":{"currentManaPercent":{}},"op":"OpGt","rhs":{"const":{"val":"0.7"}}}}`.
  Paired, 20k it in 8 blocks, `tools/audit/rotations/ab.py`, baseline 929.7 at 180 s:
  | variant | S0 | 120 s | 180 s | 300 s |
  |---|---|---|---|---|
  | mana > 0.7 | 1e6 | +2.20 ±0.42 | +1.53 ±0.53 | +0.32 ±0.56 |
  | mana > 0.7 | 2e8 | +3.18 ±0.36 | +2.12 ±0.44 | +0.74 ±0.52 |
  | mana > 0.7 | pooled | +2.69 | +1.83 | +0.53 |
  | mana > 0.6 | pooled | +3.07 | +2.18 | +0.22 (-0.16 on S0 = 1e6) |
  Fairly sure of the gain at 120 and 180 s; 300 s is flat.
- proposed fix: the 0.7 line above (0.6 is +0.35 more at 180 s but flat to negative at 300 s). Also update
  that line's note, which still says "15% of Intellect ... for 15 sec" (now 12%, 18 sec).
- needs: orchestrator fix
- resolved: 2026-11-02, 5a6e9c6a6 (obsolete, not applied): the re-searched Frost preset `0502050030001-4-253500230111301211` drops the Arcane Blast and Missile Barrage talents for Incineration 4, so the Arcane Blast line (now line 11 of level60_frost.apl.json) is inert: 0 casts in 300 iterations. Frost 929.3 -> 944.9 at 180 s. Its note still says 15% / 15 sec, inert.

### rotations-20261028-1: BM, MM and LW caveats still say Trueshot Aura 100 and Hawk rank 5 (90), on the published report
- severity: low (wording only; the numbers are right)
- affects: hunter__beast-mastery, hunter__marksmanship, hunter__lone-wolf (report notes)
- found: 2026-10-28, master 1274e15b3
- evidence: 73cedc9a9 moved the sim to the 2026-10-28 hotfix: Trueshot Aura rank 5 = 50 (`sim/core/buffs.go:1748`), and Hawk rank 6 (14322) at 55 RAP, ranks 1-5 lowered to 15/22/30/38/47 (`sim/hunter/aspects.go:48-52`); the APLs and reqs cast 14322. But `tools/dps_report/req/hunter__{beast-mastery,marksmanship,lone-wolf}.meta.json:65-66`, and the deployed `ui/dps_report/data/notes.json`, still say "Assumed: Trueshot Aura's top rank gives 100 ranged attack power (the client has 50)" and "Aspect of the Hawk stays at rank 5 (90 RAP). Forever's rank 6 gives only 55." Sure: `git grep 'stays at rank 5' origin/master`.
- update 2026-10-28, master dd663e652: still open. The phrasing pass reworded both lines but kept the numbers: line 65 "Assumed: Trueshot Aura's top rank gives 100 ranged attack power (the client says 50)...", line 66 "Uses Aspect of the Hawk rank 5 (90 ranged attack power). Forever's rank 6 gives only 55." Same in notes.json.
- update 2026-10-30, master d50909a37: still open, and copied into the four new pet variants: `git grep -l 'rank 5 (90' origin/master -- tools/dps_report/req` lists hunter__{beast-mastery,marksmanship}{,-tallstrider,-wolf} and hunter__lone-wolf (lines 68-69 in the variants). The fix below applies to all seven.
- update 2026-10-30, master fb1485888: still open, same seven meta files and notes.json.
- proposed fix: line 65 becomes "Trueshot Aura's top rank gives 50 ranged attack power, the launch realms' 2026-10-28 hotfix (ranks 1-4 lowered to 15/22/30/40). Hunter's Mark gives 143, the server hotfix (the client has 71)."; line 66 becomes "Aspect of the Hawk is rank 6 (55 RAP): the 2026-10-28 hotfix lowers ranks 1-5 to 15/22/30/38/47." Then regenerate notes.json and run `make dps-report-deploy`.
- needs: orchestrator fix
- resolved: 2026-10-31, e1fcd57b7: the seven hunter metas say Trueshot Aura 50 and Aspect of the Hawk rank 6 (55); notes.json regenerated and redeployed.

### rotations-20261030-2: Arms and Fury 2H Hamstring rage gate is too high since extra attacks keep the swing timer
- severity: medium (+4.3 arms, +3.1 fury_2h at 180 s)
- affects: warrior__arms, arms_dragonkin, arms-duty-sunder, arms-duty-battle-shout, fury_2h and its duty variants
- found: 2026-10-30, master fb1485888
- evidence: Hamstring 7373 r3 "For the procs." line, `ui/warrior/apls/level60_arms.apl.json:24` (`{"cmp":{"lhs":{"currentRage":{}},"op":"OpGe","rhs":{"const":{"val":"30"}}}}`; same 30 in level60_arms_battle_shout:27, arms_dragonkin:25, arms_sunder:25), and `level60_fury_2h.apl.json:25` (val 20). Variant: the rage const -> `10`.
  - arms (baseline 995.9 at 180 s): s0=1e6 120 +5.83 ±0.80, 180 +4.29 ±0.53, 300 +4.67 ±0.49; s0=2e8 120 +5.61 ±0.84, 180 +4.28 ±0.37, 300 +5.04 ±0.38; pooled 120 +5.7, 180 +4.3, 300 +4.9. Threshold 15 is equal (pooled 180 +4.3); 20 +3.5, 25 +1.9, 40 -5.6 (s0=1e6, 180 s).
  - fury_2h (baseline 1027.2): s0=1e6 120 +2.04 ±0.45, 180 +3.16 ±0.52, 300 +1.96 ±0.37; s0=2e8 120 +1.88 ±0.70, 180 +2.98 ±0.60, 300 +2.32 ±0.27; pooled 120 +2.0, 180 +3.1, 300 +2.1. Raising it to 30 loses 4.2.
  - arms-duty-sunder screen (s0=1e6, 180 s): +2.90 ±0.82.
  - Heroic Strike gates hold: arms 90 -> 80 -6.2, -> 70 -17.5; fury 80 -> 70 -1.1, -> 90 -2.3; fury Hamstring 10 -> 25 -0.1.
- proposed fix: Hamstring rage const 30 -> 10 in the four arms APLs, 20 -> 10 in level60_fury_2h (and its battle_shout/sunder variants if they carry the line), and in the embedded req rotations.
- needs: orchestrator fix
- resolved: 2026-10-31, e1fcd57b7: Hamstring from 10 Rage in the four Arms APLs and the three two-hand Fury ones, and their requests; the report re-simmed them.

### rotations-20261030-1: Ret paladin twist window of 1.5 s now loses 14 DPS; twisting whenever the Echo is down wins
- severity: high (+1.4% at 180 s on paladin__level_60 and sub_holy)
- affects: paladin__level_60, level_60_sub_holy, sub_holy_undead and their -duty-wisdom / -duty-crusader variants (the shock presets have no twist lines)
- found: 2026-10-30, master fb1485888
- evidence: the two Twist of Light lines, `ui/retribution_paladin/apls/level60.apl.json:14-15` (also level60_sub_holy.apl.json:14-15, level60_undead.apl.json:12-13, level60_crusader_duty/level60_wisdom_duty.apl.json:17-18), gate on `{"cmp":{"op":"OpLt","lhs":{"autoTimeToNext":{"autoType":"MainHand"}},"rhs":{"const":{"val":"1.5s"}}}}`. Variant: that const `1.5s` -> `10s` on both lines (in effect: no swing gate, swap whenever Echo 1311703/1311704 is down). Paired ab.py, 8x2500 it, paladin__level_60 (baseline 960.8 at 180 s):
  - s0=1e6: 120 +11.35 ±0.54, 180 +13.86 ±0.43, 300 +4.85 ±0.51
  - s0=2e8: 120 +12.35 ±0.65, 180 +14.02 ±0.37, 300 +4.65 ±0.28
  - pooled: 120 +11.9, 180 +13.9, 300 +4.75. sub_holy screen (s0=1e6, 180 s): +14.91 ±0.39.
  - window sweep at 180 s (s0=1e6): 1.4 -0.2, 1.6 +2.2, 1.7 +2.7, 1.8 +3.6, 2.0 +4.9, 2.5 +8.3, 3.0 +6.6, 10 +13.9.
  - Where it comes from (3000 it, seed 7): ~4 more casts each of Seal of Righteousness 20293 and Seal of Command 20920 a fight, SoC procs 20947 +10 DPS, 25713 +8, Consecration/Exorcism slightly fewer.
  - Regression: on the bbbaf2e CLI the same 10s variant lost (-4.0 ±0.27 at 180 s), so this pass's sim changes (most likely f3017897f, extra attacks keep the swing timer) flipped the "1.5 s is a sharp optimum" ruling. Paladin presets rose +1.8 to +3.2% in the baseline here with no paladin APL change. Fairly sure of the gain; the mechanics auditor may want to check that the Echo/extra-attack interaction is right before the window is widened.
- proposed fix: on both twist lines in the five APLs above (and the embedded req rotations), drop the `autoTimeToNext < 1.5s` term, or set it to `10s`. Re-check the duty and undead variants after (only level_60 and sub_holy were measured). Update the "1.5 s sharp optimum" settled item in passes/rotations.md.
- needs: orchestrator fix
- resolved: 2026-10-31, e1fcd57b7: rotopt paladin_ret twistWindow=0 (no swing gate) on the four twisting APLs against the generic boss and their requests; the report re-simmed the Ret builds (+12.5 to +16.0). level60_undead (sub_holy_undead) keeps the 1.5 sec window: there the free window lost 1.5 in the re-sim, and paired ab.py on two seed sets gives the 1.5 sec window +1.1 to +3.5 at 120 / 180 / 300 sec.

