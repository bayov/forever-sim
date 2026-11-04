# Findings: healer_fairness

## Open

### healer_fairness-20261104-1: the tank-role bench has no raid, so Prayer of Mending ping-pongs between the tank and the priest
- severity: high (Holy priest tank carry -0.085 mean, -0.177 on Magmadar, once fixed; the priest's lead over the shaman goes from +0.135 to +0.05)
- affects: Holy priest tank healer (the Discipline preset casts no PoM); every tank-role ranking in docs/healing_sim.md H3
- found: 2026-11-04, master a1b1163d2 (tank role unchanged on c8f5858bd: priest_holy Magmadar re-run there)
- evidence: `tools/healbench/main.go:226-228` sets fillers to 0 in `-role tank`, so the raid holds only the tanks and
  the one healer. PoM's jump (`sim/priest/healing.go:594-613`) picks any other living player without PoM, so it can
  only go to the tank or the priest; Holy Binding's self-heal (`sim/priest/healing_talents.go:168-169`) sets the charge
  off on the priest at once through `OnHealTaken` (`healing.go:570-574`) and it jumps straight back. At carried damage
  on Magmadar the tank gets about 3 PoM heals a cast (all 5 charges: tank, priest, tank, priest, tank); with 7 idle
  fillers in the raid, 1. PoM's own model matches client and tooltip (5 charges, 10 s cd, 0.429, 390 mana).
  With 7 idle fillers in tank role (healbench overlay, 1,000 iterations, all 7 files, paired seeds):
  Holy priest -.016 -.061 -.088 -.118 -.177 -.043 -.092 (Ony, Brood, Chrom, Golem, Mag, Rag, Vael), mean 1.231 to 1.146;
  shaman +0.003, paladin +0.010, druid -0.004 (noise). With fillers, dropping PoM from the rotation moves the priest
  +0.007: PoM is worth about nothing on a tank healer once it can't bounce back. Sure (reproduced on every file).
  The overlay is `ref/healer_fairness/tank_fillers_overlay.diff` (env `HB_TANK_FILLERS=1`); the deltas are in
  `ref/healer_fairness/attribution_a1b1163.tsv`. Tank columns on c8f5858bd moved by at most 0.05 against a1b1163d2
  (`ref/healer_fairness/tank_c8f5858.tsv`), so the measurement stands.
- proposed fix: in `-role tank`, keep the fillers (or a few idle raid members) in the raid, taking no raid damage and
  not counted in the score; or keep PoM from jumping to healers in a tank bench. Then re-bench the tank columns of H3.
- needs: orchestrator fix

### healer_fairness-20261104-2: the shared "missing at least the heal's Expected Healing" gate holds back the shaman, paladin and druid tank presets
- severity: medium (shaman +0.026 mean carry, paladin +0.018, druid +0.011, priest +0.005; no rank changes)
- affects: tank-healer presets of all four classes (tools/healbench/req/*.json, the ui presets they come from)
- found: 2026-11-04, master a1b1163d2
- evidence: every tank APL casts its main heal on the tank only when it misses at least 100% (druid 90%) of the
  heal's Expected Healing. Lowering it, 1,000 iterations, all 7 files: shaman 0.8x plus Lesser Healing Wave gate 0
  +.034 +.022 +.026 +.021 +.046 +.013 +.023 (mean +0.026; 0.8x alone +0.020, 0.6x +0.022); paladin 0.6x +0.018,
  0.8x +0.013; druid 0.8x +0.011; priest 0.8x +0.003, 0.6x +0.005. None of them is mana-bound at its carry (39-89%
  mana left at the end; mana gates at 0 change nothing), so a heal held for a bigger hole is throughput lost.
  The gate costs the slow-heal classes more than the priest, which is part of the priest's lead (about 0.02).
- proposed fix: re-tune each tank preset's gate (0.6-0.8x) in its class's preset and request; the healer_rotations
  auditor can confirm the best per class.
- needs: worker fix

### healer_fairness-20261104-3: the Holy priest preset's slot 16 is a budgeted synthetic wand, the others' a stat-less relic; its Power Infusion line never fires
- severity: low (wand -0.003 carry mean when removed; Power Infusion 0)
- affects: Holy and Discipline priest presets
- found: 2026-11-04, master a1b1163d2
- evidence: 990514 (+6 Int, +10 SP, +15 healing, about 21 weighted points, +2.4% of the 902 budget) sits on top of the
  set, while the shaman, paladin and druid carry real relics with no budgeted stats (249398, 272436, 279250); the
  four armor budgets are 901-903 without slot 16 (standards 1b, 35). `priest_holy.json`'s APL has a Power Infusion
  action, but the talents `00500503113-03505103030121531` don't take Power Infusion, so it never casts.
- proposed fix: count the wand inside the 902 budget (or give it no stats), and drop the dead Power Infusion line
  (or take the talent if it gains).
- needs: worker fix

### healer_fairness-20261104-4: the boss files' raidHealing has no healing by class, so the sim's healer ranking can't be checked against the logs
- severity: low (blocks a check; moves no number)
- affects: the pass's step 2 for every healer
- found: 2026-11-04, master c8f5858bd
- evidence: `jq '.raidHealing' tools/encounters/<boss>.json` holds `healerClasses` (mean count a kill), `hpsEachHealer`
  (one median for all healers, with overheal), `raidHps`, `hpsOnTanks`, `hpsOnOthers`, nothing by class
  (`tools/encounters/raid_from_logs.py:144-158`); the mined `tools/encounters/logs/anniversary_*_raid.json` keep
  the same. Composition over the 7 files: priest 5.11, shaman 3.00, paladin 1.69, druid 1.21 of 11.
- proposed fix: have `wclcli.py raid` (tools/warcraftlogs/wclcli.py) keep effective HPS and tank share by healer
  class (median per player), re-mine, and store it in `raidHealing.byClass`.
- needs: orchestrator fix

## Resolved
