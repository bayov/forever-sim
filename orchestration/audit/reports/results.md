# Results pass 91 (master cfc7f9382; shared run on 25ce81de7, later commits are notes only)

Shared run (68 builds, 20k, seed 1) vs published:
- builds.json: 68 of 68 within 1.5 of the run (2694dc2ed re-sim; 6a21e3bcd healing mode not in the DPS set).
- Metas: 28 of 68 off by more than 0.5 (was 41): Affliction +31.9 to +36.7, hunters (lone-wolf, marksmanship, BM) -9 to -22, rogues combat/seal-fate -10 to -20, mage fire -4.9. Same families as results-20261104-1/-2. No new family.
- Moves since pass 90 (builds.json 70ea588bb -> now, 30 builds, all explained by the Nov 15 round):
  - Windfury Totem on every main-hand melee hit (2694dc2ed, 6d27e1591): fury_2h +6.7 to +6.8% (+67 to +70), fury +3.8 to +4.0%, arms +2.3 to +2.5%, Enhancement +2.5 to +2.7%, feral +2.5 to +2.6% (with Bite at 65 Energy), Survival +2.0%, subtlety +0.4%.
  - Paladins +3.8 to +4.4% (6ce77e5ef Judgement of Command on melee defense, baa363615 Two-Handed Weapon Specialization's Holy reach, c26878d48 Ret twist and Shock re-tune; 6158c9602, 8b12a8a8c re-sims).
  - Execute's second totem roll and Blood Fury weapon damage are listed as probable-bug assumptions (user rulings, not re-flagged).
- Drift (mksettings regen vs committed): only the known feral sapper line on both feral requests (results-20261105-1). Reference drift_report.txt is older than that finding.
- Unread fields: none (no output).
- Reference req_summary.txt not replaced (still the pass 40 copy); builds.json is the working reference.
