# Findings: healer_fairness

## Open

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

### healer_fairness-20261112-1: the Holy priest raid preset keeps Inspiration, dead in the raid role, instead of Twilight Focus; every other raid preset takes its pushback talent
- severity: high (+0.021 mean raid carry, 1.5%; widens the Holy priest's raid lead from +0.026 to +0.047)
- affects: Holy priest raid healer (`tools/healbench/req/priest_holy_raid.json`, the ui preset it comes from); the tank preset is right
- found: 2026-11-12, master 4fb90c298
- evidence: both Holy presets use `00500503113-03505103030121531` (Inspiration 3/3, Twilight Focus 0/3). Inspiration
  (+25% armor after a heal crit) only matters on a tank, and the raid role has none; Twilight Focus (70% chance not to
  lose cast time when damaged) is the priest's pushback talent, the one the shaman raid preset takes (Healing Focus 3
  for Ancestral Healing 3), the paladin (Spiritual Focus 5) and the druid (Nature's Focus 5). One-point swap
  `00500503113-33505100030121531` (Twilight Focus 3, Inspiration 0; row gates unchanged), healbench -role raid -carry,
  1,000 it, paired seeds: seed 101 1.440 -> 1.461 (Vaelastrasz 1.125 -> 1.262, the rest -0.017 to +0.012), seed 202
  1.449 -> 1.469 (Vaelastrasz 1.123 -> 1.263). The same swap in the tank role costs 0.073 (1.278 -> 1.205), so only the
  raid preset should change. Counter-check: the shaman raid preset with Ancestral Healing back for Healing Focus loses
  0.039 (Vaelastrasz -0.156), so pushback is what Vaelastrasz's raid-wide fire damage rewards. Rows in
  `ref/healer_fairness/screen_4fb90c2.tsv` (holy_raid_TF, sham_raid_AH, holy_tank_TF); helper
  `tools/audit/healer_fairness/screen.sh`. Sure (two seeds, one file carries it, the mechanism is clear).
- proposed fix: give the Holy priest raid preset `00500503113-33505100030121531` (ui preset and req), keep the tank
  preset's string; re-bench the raid column.
- needs: worker fix
- resolved: 2026-11-12 (the Holy priest raid preset takes `00500503113-33505100030121531`: TalentsHolyRaid in
  ui/healing_priest/presets.ts on the Holy raid healer build, tools/healbench/req/priest_holy_raid.json, the
  priest_tuning `holy_raid` talents; the tank preset keeps Inspiration). Re-benched, 1,000 it, seed 101: raid mean
  1.440 -> 1.461 (Onyxia 0.988, Broodlord 0.906, Chromaggus 1.065, Golemagg 2.607, Magmadar 2.062, Ragnaros 1.340,
  Vaelastrasz 1.262), first; seed 202 1.449 -> 1.469 (the finding's pair). Rows in
  `ref/healer_fairness/retune_25f65da_s101.tsv`.

### healer_fairness-20261104-5: the Holy paladin raid preset's Holy Light lines cost it 0.025 carry; a Flash of Light raid paladin is first, not third
- severity: high (+0.025 mean raid carry, 1.8%; the paladin goes from 1.413, third, to 1.438, level with the Holy priest and druid at 1.437)
- affects: Holy paladin raid healer (tools/healbench/req/paladin_holy_raid.json and the ui preset it comes from)
- found: 2026-11-04, master 0facd9428
- evidence: healbench -role raid -carry, 1,000 iterations, all 7 files, paired seeds (same seed for every variant):
  the tank preset (paladin_holy.json) in the raid role 1.034 .904 1.121 2.695 1.954 1.331 1.027 (mean 1.438) against
  the raid preset 0.984 .905 1.081 2.699 1.918 1.301 1.002 (1.413); up on 6 of 7, Golemagg -0.004. In the raid role
  the bench has no tank, so the tank preset's Holy Light and Oath of Light lines (target Tank 0) never fire: removing
  its Holy Light lines leaves every cell unchanged (1.438), and adding its Oath of Light line to the raid preset
  changes nothing (1.413). What runs is Lay on Hands, Holy Shock, Flash of Light under 35% and Flash of Light 6 at the
  most hurt at its 0.7x gate. The raid preset with its Holy Light 9/8 and Divine Favor lines removed: 1.426 (+0.013;
  the rest is the Flash of Light lines' gate). Mana is no limit either way (about 7,900 of the pool left at carry, no
  OOM). Answers open question 4 of orchestration/fairness_healers.md (pass 1 saw +0.017 on 3 files at 500 it).
  Variant requests were built with jq/python from the two presets; rows in `ref/healer_fairness/bench_0facd94.tsv`
  (healers paladin_holy in role raid, paladin_holy_raid_nohl, paladin_holy_tank_nohl, paladin_holy_tank_mosthurt).
  Sure of the direction (paired seeds, 6 of 7 files); the best raid APL itself is not searched.
- proposed fix: re-tune the raid preset with Holy Light 9/8 and Divine Favor dropped or gated far higher, and Flash
  of Light 6 at the most hurt at 0.7x (as the tank preset's); the healer_rotations auditor can search the best. Then
  re-bench H3's raid column.
- needs: worker fix
- 2026-11-12, master 4fb90c298, still open on the new consumables: the tank preset's APL in the raid role with the raid
  preset's Lesser Mana Oil 1.424 against the raid preset's 1.400 (+0.024, 1,000 it, seed 101; Onyxia +0.033,
  Broodlord +0.011, Chromaggus +0.044, Golemagg +0.016, Magmadar +0.019, Ragnaros -0.031, Vaelastrasz +0.076); the
  raid preset with its Holy Light 9/8 and Divine Favor lines dropped 1.403 (+0.003). A Flash of Light raid paladin
  would be second (1.424), ahead of the druid's 1.414.
- resolved: 2026-11-12 (the raid preset is a Flash of Light healer: no Holy Light or Divine Favor line; Lay on
  Hands under 10%, Holy Shock under 50%, Flash of Light 6 under 35%, Flash of Light 6 and 4 at the most hurt at
  0.7x of their Expected Healing; ROLE_DEFAULTS raid in orchestration/tools/healers/paladin_holy_requests.py,
  which now also reproduces the shipped requests: the tank's 0.7x gate and the raid's Lesser Mana Oil were
  knobs it had lost). Healbench -role raid -carry, 1,000 it, 7 files: 1.400 -> 1.430 (seed 101), 1.402 -> 1.429
  (seed 202); Flash of Light 6 alone at 0.7x 1.424 / 1.430, 6, 5 and 4 1.430 / 1.430 (a tie, the preset binds
  two ranks); 13 other variants (gates 0.6/0.8x, ranks 6+5, 6+3, 4 alone, Holy Shock at 60/70% or as a filler,
  Divine Favor under 35/50%, Concentration Aura, no Lay on Hands) at or below 1.425 (docs/healing_paladin.md,
  Raid healer). Seed 101 by file: Onyxia +0.041, Broodlord +0.012, Chromaggus +0.059, Golemagg +0.018, Magmadar
  +0.011, Ragnaros -0.007, Vaelastrasz +0.078. Onyxia at 300 s: 7,559 mana left at carry, no OOM. The raid
  paladin is now second (1.430), behind the Holy priest (1.461), ahead of the druid (1.414) and shaman (1.412).
  Rows in `ref/healer_fairness/retune_25f65da_s101.tsv` and `_s202.tsv`.

### healer_fairness-20261111-1: Nature's Swiftness now goes on Wild Growth in the druid raid preset; raid druid falls from first (1.437) to last (1.391)
- severity: high (mean raid carry -0.046, -3.2%; the druid goes from level first to fourth, behind the shaman's 1.395)
- affects: Restoration druid raid preset (`tools/healbench/req/druid_restoration_raid.json`); the tank preset is unchanged (identical rows on Broodlord and Vaelastrasz)
- found: 2026-11-11, master 707ae92cc5
- evidence: 9628e85b6 adds Wild Growth to Nature's Swiftness's spells (hotfix 113885, sim/druid/healing.go:756). The raid
  preset fires NS from `autocastOtherCooldowns` (line 1 of its APL, "...and Nature's Swiftness"), and its Wild Growth line
  (APL notes "Wild Growth at the most hurt") sits above the Healing Touch lines, so NS is now spent on a Wild Growth cast
  instead of a Healing Touch. healbench -role raid -carry, 1,000 it, carry old (bench_9ba00c2.tsv) -> new: Onyxia 0.997 -> 0.943,
  Broodlord 0.920 -> 0.906, Chromaggus 1.109 -> 1.092, Golemagg 2.660 -> 2.658, Magmadar 2.002 -> 1.818, Ragnaros 1.270 -> 1.240,
  Vaelastrasz 1.102 -> 1.079; mean 1.437 -> 1.391. HPS rises (Magmadar 800 -> 903) but overheal too (74.7% -> 78.1%).
  Attribution sure: healbench built from 707ae92cc5 with only that healing.go hunk reverted gives the baseline exactly
  (Magmadar 2.002, Onyxia 0.997). Holy priest, shaman and the druid tank preset match the baseline (dot.go's change reaches no HoT here).
- proposed fix: in the druid raid APL, take NS out of the autocast and cast it explicitly right before a Healing Touch line
  (e.g. castSpell NS gated on the HT 11 condition), or gate Wild Growth on NS not being active. Re-bench the raid column
  after; expect the druid back near 1.437. Whether NS on Wild Growth ever pays (instant party HoT on a spike) is the
  healer_rotations pass's to measure.
- needs: worker fix (healer_rotations)
- resolved: 2026-11-12, e05135ad3 (NS on its own line in both druid presets: an instant Regrowth under 45% in the raid,
  an instant Healing Touch 11 under 60% on the tank; Subtlety 2 -> Heart of the Wild 2). 1,000 it, own lengths: raid
  1.391 -> 1.438 (first, level with the Holy priest's 1.437), tank 1.235 -> 1.273 (first, Holy priest 1.262). Other
  healers' rows unchanged (no sim change outside a comment). The shaman raid preset still leaves its Nature's
  Swiftness (16188) to the autocast, as the druid's did; whether a held NS pays there too is the healer_rotations
  pass's to measure.

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
- resolved: 2026-11-04, 7b6623f27 (healbench: the tank role keeps the raid, idle). Re-benched on 0facd9428: the Holy priest tank lead over the shaman is +0.020 (1.262 against 1.242).

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
- resolved: 2026-11-04, 324d64834 (shaman 0.6x), 0b15937a8 (paladin 0.7x), a7d9b57ef (priest 0.8x/0.5x), bc8e5473b (druid 0.8x/0.7x); the workers measured each against the old gate on all 7 files.

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
- resolved: 2026-11-04, dda14db65 (990514 carries no stats; Power Infusion line dropped).
