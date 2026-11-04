# Findings: Results and spec strength

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### results-20261104-2: 33 metas / builds.json lag the phase 5 hit-trade regears (feral +26, Ret +11, Fury builds.json -16 to -25)
- severity: low (no sim error; the requests moved, the published numbers have not been re-simmed yet)
- affects: druid feral-cat (+2), paladin level_60 (+3), mage arcane, rogue (6), shaman (6), warlock destruction_fire (3), warrior arms/fury (builds.json only)
- found: 2026-11-04, master 0facd9428
- evidence: shared run (68 builds, 20k, seed 1) vs meta published.dps: feral-cat 1164.9 vs 1139.0 and duty-ff 1160.1 vs 1134.1 (74cc2b8e3 regear, hit-4); paladin level_60 1103.2 vs 1092.1, duty-crusader +11.8, duty-wisdom +10.7 (3f4d44d4a, Ret twist regear onto hit-3); mage arcane +7.6 (74cc2b8e3); shaman elemental/enh +6.5 to +7.4 (87922bb77 hotfix 113380 LB/CL/shock, 3e4796fe7 regear); rogues -12 to +7.9 (31c10b61c, aa4551a04 regear onto hit steps, on top of results-20261104-1); warlock destruction_fire +0.8 to +1.1 (c8a7b4ac3, talents 242, the meta note says +1.1 at 180 s, matches). Warrior: metas match the run (cd3f44cf2, <=1.6), builds.json is 6.6 to 25 higher than the run (arms, fury, fury_2h families), waiting on the republish. Shock Ret duties are within 0.3. Drift report equals the reference; no new drift.
- proposed fix: re-sim and republish the report (metas and builds.json) for these builds.
- needs: orchestrator fix

### results-20261104-1: five rogue metas' published.dps lag the hotfix re-sim (up to 17.3 DPS)
- severity: low (builds.json, from e9185b1aa, matches my 20k seed-1 run on all five; only the metas lag)
- affects: tools/dps_report/req/rogue__{seal-fate-daggers,seal-fate-swords,mutilate,combat-swords,combat-daggers}.meta.json; runreq.sh and verify.py flag them OFF
- found: 2026-11-04, master 6b480b060
- evidence: shared run vs meta: seal-fate-daggers 903.4 vs 920.7, seal-fate-swords 940.4 vs 957.0, mutilate 858.4 vs 866.1, combat-swords 958.6 vs 963.5, combat-daggers 949.9 vs 946.6. The moves come from 316a600ac (hotfix 113483: Mutilate 44/63/88, Shank, Venom, tree); rogue__subtlety's meta was set (881.7 vs 881.3), the other five were not.
- proposed fix: set the five metas' published.dps and source from builds.json.
- needs: orchestrator fix

## Resolved

### results-20261102-1: shaman__elemental-level60-mana-tide meta published.dps is 0.4 off (753.2 vs 752.8)
- severity: low (builds.json 752.8 matches my 20k seed-1 run 752.83; only the meta lags)
- affects: tools/dps_report/req/shaman__elemental-level60-mana-tide.meta.json; runreq.sh and verify.py flag it OFF
- found: 2026-11-02, master 796495d6f
- evidence: shared run /tmp/audit-work/runs/master/summary.tsv: 752.83 vs meta 753.2; the other 72 metas are within 0.15. Likely the meta was set before the last mana-tide re-search; sure of the numbers, not of the cause.
- proposed fix: set the meta's published.dps to 752.8 and its source from builds.json.
- update 2026-11-03, master f5c99ad018: the shaman worker left all six shaman metas at the 2026-11-02 values on purpose ("meta.published is left for the report run", round_70753/shaman.md:172). Run 970.1, 952.7, 1149.8, 1159.2, 1055.1, 1111.3 match the worker's finals; the metas still say 768.5, 753.2, 879.3, 883.5, 795.8, 834.8. Supersedes the 0.4 gap above; expect it to clear with the regeneration.
- update 2026-11-04, master f331f0cff: unchanged. Shared run (68 builds, 20k, seed 1) matches builds.json on all 68 within 0.15; only the six shaman metas still carry the 2026-11-02 values (+199 to +277 DPS gap in meta vs run).
- update 2026-11-04, master fb584eafa: unchanged for the shaman metas. The warrior Fury follow-up (7b4dc5d97) set its six metas to the run (all within 0.15); builds.json still has the pre-follow-up warrior numbers (fury_2h 1169.9 vs run 1173.3, fury_2h-duty-bs 1165.4 vs 1168.8, fury_2h-duty-sunder 1132.6 vs 1134.6), waiting on the republish.
- needs: orchestrator fix
- resolved: 2026-11-04, 6b480b060 (shaman metas now match the run; builds.json caught up with the Fury follow-up)

### results-20261103-2: hunter notes-only shares (Arcane Shot 0.15, Serpent Sting 5% a tick) and the hotfix shot tables are not gated under Forever
- severity: low (no report number affected; only a non-Forever hunter sim would see them)
- affects: hunter (all builds, only when Env is not Forever)
- found: 2026-11-03, master f5c99ad018
- evidence: sim/hunter/arcane_shot.go:19 `arcaneShotRAPCoefficient = 0.15` and sim/hunter/serpent_sting.go:22,77 `serpentStingRAPPerTick = 0.05` are plain constants; the dot tick even has an `if sim.IsForever()` right after (line 78), so the file does have Forever/Classic branches. The druid (sim/druid/tuning_70753.go, Forever-gated), rogue poisons (poisons.go:167,247 `TernaryFloat64(IsForever...)`) and paladin notes-only values are gated as standards 37 asks. orchestration/round_70753/hunter.md:126 says the hunter files are Forever-only (Aimed Shot's 166 was already ungated), so this may be accepted; sure of the code, not of the intent.
- proposed fix: either gate the two shares (and say so in hunter.md) or record in hunter.md that sim/hunter is Forever-only by design and drop the Classic branches.
- needs: worker fix
- resolved: 2026-11-03, 13c34f2: both shares gated under `Env.IsForever()` (outside Forever Classic's 0), and the hotfix shot tables too (Classic's Arcane Shot, Serpent Sting and Aimed Shot ranks; Sniper Shot, not in Classic, the client's 160 / 225 / 295); hunter.md says so. No Forever number moves.

### results-20261101-1: 62 of 77 metas' published.dps are stale after the 2026-11-01 re-sims (38fb06ea9, 11de3318a)
- severity: low (no published number is wrong: builds.json matches my 20k seed-1 run on all 77 within 0.1)
- affects: tools/dps_report/req/*.meta.json (74); runreq.sh and verify.py report them as OFF published by 0.5 to 8.9 DPS
- found: 2026-11-01, master dd053d24c
- evidence: 38fb06ea9 re-simmed 32 builds into builds.json but the metas still carry earlier values (druid__balance meta 851.2, builds.json and run 846.6; warlock__destruction_fire 986.4 vs 977.7; hunter__lone-wolf 1094.3 vs 1087.4). Same pattern as results-20261030-1 (resolved).
- update, 2026-11-02, master 5a6e9c6a6: now 62 of 77 off by more than 0.15 (the 15 paladin metas were set from the hotfix re-search, so they match my run, but builds.json still has the old paladin numbers). The 9 pet hunters' metas still carry the pre-11de3318a values (BM 1131.8 vs builds.json and run 1045.4).
- proposed fix: set each meta's published.dps and source from builds.json, per tools/dps_report/README.md "update meta.published".
- needs: orchestrator fix
- resolved: 2026-11-02, 3fd9b46d2 (metas set from the 2026-11-02 report; all 73 metas, builds.json and my 20k seed-1 run agree within 0.15)

### results-20261025-1: mksettings and the committed requests disagree on consumes (potion, flask, filler explosive)
- severity: medium (Ret Shock -13.7 DPS = -1.6% if regenerated; others under 2.5)
- affects: paladin__level_60_shock* (defaultPotion MajorManaPotion vs DragonfirePotion), shaman__elemental-level60-mana-tide (vs CausticSmogPotion), all rogues (flask FlaskOfNaturalAccuracy vs FlaskOfNaturalPrecision on combat-daggers; fillerExplosive ExplosiveGoblinLandMine absent in requests), priest__smite and feral duty-ff (fillerExplosive)
- found: 2026-10-25, master 54878fa3d
- evidence: `python3 tools/audit/results/regen.py` vs `ref/results/drift_report.txt` (previous reference). Paired 20k seed 1, regen JSON vs committed: Ret Shock 855.3 vs 869.0; Elemental mana-tide 759.9 vs 762.1; Smite filler explosive part of an already-standing 4-field drift; rogue combat-daggers 768.7 vs 770.3, combat-swords 762.0 vs 764.0, mutilate 718.7 vs 718.9, seal-fate-swords 728.8 vs 728.9. Published numbers are unaffected (all 64 requests within 0.15 of published).
- proposed fix: make mksettings defaults match the standards 26 consumes the requests carry (or regenerate the requests); the Ret/Ele potion and rogue flask are the ones that move DPS.
- needs: orchestrator fix
- resolved: 2026-10-31 (user approved the fix): the potion and flask lines were regen.py's spec mapping, not mksettings (it ran every paladin as `ret` and every Combat rogue as `rogue_combat`; mksettings has `ret_shock` with the Major Mana Potion and `rogue_combat_daggers` with Natural Accuracy). regen.py now takes a variant's spec from its base build, then `-spec` from the meta's notes, then the id. The real drift is gone: mksettings has `rogue_combat_swords` and `rogue_combat_daggers` without the Goblin Land Mine, `rogue_assassination` without it, `ele_mana_tide` with the Major Mana Potion, and `shadow_holy` and `smite` with their own consumes, so a regenerated request equals the committed one. Left by design: the feral Faerie Fire duty preset drops the base's EZ-Thro Dark Bomb, a duty edit its notes record, like every duty preset's debuff toggles. ref/results/drift_report.txt is regenerated. No request changed, so no published number moves.

### results-20261030-1: 46 of 77 metas' published.dps are stale after the 2026-10-30 re-sim (cd1f6b2d0, fb1485888)
- severity: low (no published number is wrong: builds.json matches my 20k seed-1 run on all 77 within 0.15)
- affects: tools/dps_report/req/*.meta.json (46); runreq.sh and verify.py report them as OFF published
- found: 2026-10-30, master fb1485888
- evidence: the re-sim commits update ui/dps_report/data/builds.json but not the metas (warrior__arms meta 905.6, builds.json 994.9, run 994.9; hunter__survival 1028.9 vs 1072.6; mage__fire 1089.0 vs 1061.0). Same pattern as results-20261027-1.
- proposed fix: set each meta's published.dps and source from builds.json, per tools/dps_report/README.md "update meta.published".
- needs: orchestrator fix
- resolved: 2026-10-31: every meta's published.dps and source set from builds.json after the re-sim, in the same commit as the data.

### results-20261027-1: all 58 metas' published.dps are stale after the 2026-10-26 racial-pass re-sim
- severity: low (no DPS number is wrong: builds.json matches my 20k seed-1 run on all 58 within 0.1)
- affects: tools/dps_report/req/*.meta.json (all 58); runreq.sh / verify.py compare against them
- found: 2026-10-27, master 06a898751
- evidence: 4273ee51d re-simmed every build into ui/dps_report/data/builds.json but left `published.dps` in the metas at the 10-25 values (warrior__fury meta 986.6, builds.json 975.2, run 975.19; mage__arcane 982.5 vs 899.1, -8.5%; Ret Shock metas were updated only by 3d9aa585b). Every request is "OFF published" by 0.15 to 83 DPS until they are updated.
- proposed fix: set each meta's `published.dps` (and `source`) from builds.json, as tools/dps_report/README.md "update meta.published" says.
- update 2026-10-28, master 5d27b0015: still open after post 8's re-sim (73e31ec4c, 9857d17e3). 34 of 58 metas differ from builds.json by more than 0.15 (mage__frost 946.6 vs 911.1, warlock__affliction 1035.3 vs 1018.1); my 20k seed-1 run matches builds.json on all 58 within 0.1.
- update 2026-10-28, master 224b8a4e0: still 34 of 58 after the racial-audit deploy.
- needs: orchestrator fix
- resolved: 2026-10-28, every req/*.meta.json published.dps set from the 2026-10-28 full re-sim's builds.json.

### results-20261103-1: rogue__subtlety meta has race "Night Elf" (every other meta uses the enum name, "NightElf")
- severity: low (request and published number are right; the DPS report race table and regen.py are affected)
- affects: tools/dps_report/req/rogue__subtlety.meta.json:8 (rogue Subtlety)
- found: 2026-11-03, master f5c99ad018
- evidence: `python3 tools/audit/results/regen.py` prints `rogue__subtlety MKSETTINGS FAIL: unknown race "Night Elf"` (mksettings takes the proto name). With `-race NightElf` and the meta's other arguments the regenerated request equals the committed one (0 diffs), so no drift. tools/dps_report/build_data.py:402,412 compare `r['race'] == m['race']` (race-sweep rows use "NightElf"), so the Subtlety race table's `isPreset` flag is likely never set for Night Elf; faction falls to Alliance, which is right. Introduced in 25fc79c83.
- proposed fix: set the meta's `race` to `NightElf`.
- needs: orchestrator fix
- resolved: 2026-11-03, 5b766ca2b (meta race NightElf; regen.py now reports 0 diffs for rogue__subtlety)

