# Results pass 87, master 97da1cca5 (2026-11-13)
Shared run (68 builds, 20k, seed 1) vs builds.json: 68 of 68 within 1.5 (builds.json last touched by f6b04b9e7, the Ret re-tune and Windfury Totem re-sim).
vs metas: 39 of 68 off by more than 0.5 (was 45); largest: fury_2h families -117..-119, fury -75..-85, Ret Shock crusader duty -63, arms -41..-47, Affliction +32..+37, Survival -23..-26. Known finding results-20261104-2 (metas lag the re-sims); the paladin metas were refreshed by f6b04b9e7.
Since the last pass (71562a8b0): Windfury Totem rules, Ret Judgement re-tunes with report hashes, Volley procs (hunter hotfix 2026-11-11), Unsealed Wrath, item DB regen (8 items seen in game, 38 reveals). No sim error found; the moves are the already-explained ones (f1c07e13a, 2a0208136, hotfix 113766). Not paired-simmed this pass.
Drift (regen.py with a master mksettings): one difference against the reference, the feral Goblin Sapper line on feral-cat and feral-cat-duty-ff (results-20261105-1, still open). Unread fields: none printed besides the expected.
The reference req_summary.txt and drift_report.txt were not replaced (baseline stays 0facd9428).
