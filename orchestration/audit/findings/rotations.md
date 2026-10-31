# Findings: Rotations and resources

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

## Resolved

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
- resolved: 2026-10-31, e1fcd57b7: rotopt paladin_ret twistWindow=0 (no swing gate) on the five twisting APLs and their requests; the report re-simmed the Ret builds.

