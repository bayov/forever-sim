# healer_rotations pass 9 (2026-11-12, master 3065ba744 on tankheal, by hand)

Forced re-audit after the Restoration druid re-tune (e05135ad3), the seen-health fix for Most Hurt / Lowest Health
(65b5d49cf) and the healer consumables (ea25e50df). Bench: `tools/healbench`, 1,000 iterations, seed 101, own lengths.
# Healer rotations pass 9 (quiet: no new findings; -20261111-1, -20261104-2, -20261104-3 still open)

- Master: 8306f1d52 (`last.master` was a0b4d3be9). Audited 2026-11-12.
- Bench: `tools/healbench` built from this checkout, healbench's default seed, `-carry`, 1,000 iterations, each
  request on each of the 7 boss files at the file's own length in its role. Script:
  `tools/audit/healer_rotations/baseline.sh <healbench> <out.tsv> [iters] [extra flags]` (run from the repo root).
  It was missing from master again. This copy is restored from commit 8f3f74f3f, which is only on
  `origin/tankheal`. The run's commit must include it.
- Note: `origin/tankheal` holds a by-hand "healer_rotations pass 9" (8f3f74f3f, 2 new findings, a rewritten
  summary.tsv) that isn't merged into master. This pass is numbered from master's state (pass 8 -> 9). Whoever
  merges tankheal must reconcile the two.

## What changed since cb0d928fd

- e05135ad3: both druid presets cast Nature's Swiftness on its own line (tank: Healing Touch 11 under 60%; raid:
  Regrowth under 45%); Heart of the Wild 2.
- 65b5d49cf: Current Health / Current Health % of Most Hurt and Lowest Health read the healer's seen health. Not
  "no effect" as healer_mechanics-20261112-14 measured at 200 it: bisect on the old requests, paladin tank Golemagg
  1.911 -> 1.789, druid tank Magmadar 2.102 -> 2.003 (3fc5898c9 -> 65b5d49cf; Windfury 2a0208136 and the tank
  re-tunes move nothing here). A deliberate correctness fix, so not a finding; the thresholds were re-checked below.
- ea25e50df: Greater Cleric's Elixir, Runn Tum / Wicked Smoothie, Wizard Oil / Lesser Mana Oil on every healer.

## Baseline (mean carry of the 7 files; full rows in ref/healer_bench/summary.tsv)
The gate fired on `sim/druid/rake.go`: Forever Rake's direct hit now takes 0.065 x attack power (the user's ruling,
2026-11-12). Rake is a cat-form ability that no healer request or APL casts. Other files touched since a0b4d3be9:
`sim/core/pet.go` and `sim/core/unit.go` (pets inherit their owner's melee speed, used by the hunter pet only), the
hunter files and test results. No healer APL, healbench request, `tools/encounters/*.json` or
`tools/healbench/main.go` changed.

| request | new | ref | change |
|---|---|---|---|
| druid_restoration (tank) | 1.242 | 1.273 | -0.031 (seen health; Golemagg -0.092, Magmadar -0.098) |
| druid_restoration_raid | 1.414 | 1.438 | -0.024 (Magmadar -0.224) |
| paladin_holy (tank) | 1.152 | 1.176 | -0.024 (Golemagg -0.105, seen health) |
| paladin_holy_raid | 1.400 | 1.413 | -0.013 |
| priest_discipline | 1.272 | 1.261 | +0.010 |
| priest_holy | 1.278 | 1.262 | +0.015 |
| priest_holy_raid | 1.440 | 1.437 | +0.004 |
| shaman_restoration (tank) | 1.257 | 1.242 | +0.015 |
| shaman_restoration_raid | 1.392 | 1.395 | -0.003 |

Mana: only Discipline still runs dry (Chromaggus, OOM 101 s, -2 open). Every other row ends with 2,100+ mana.
Raid rows' death floor not recorded this pass (the first baseline.sh run read the wrong column; fixed).
All 63 rows (9 requests x 7 boss files) are byte-identical to `ref/healer_bench/summary.tsv` (`cmp` identical), so
the ref stays as it is.

## Changes tried (paired, 1,000 it, mean carry of 7 files; base in brackets)

| preset | change | mean | vs base |
|---|---|---|---|
| druid tank [1.242] | NS at Lowest < 35 / 40 / **45** / 50 / 70% (60) | 1.268 / 1.267 / **1.269** / 1.258 / 1.224 | **+0.027** -> finding -1 |
| druid tank | Swiftmend + Regrowth line at 75 / 95% (85) | 1.232 / 1.236 | -0.010 / -0.006 |
| druid raid [1.414] | NS at 35 / 55% (45) | 1.400 / 1.406 | -0.014 / -0.008 |
| shaman tank [1.257] | NS at Tank < 35 / 40 / 60% (50) | 1.249 / 1.251 / 1.259 | -0.009 / -0.006 / +0.002 |
| shaman tank | NS held for Healing Wave 10 on Lowest < 35 / 45 / 55% | 1.239 / 1.253 / 1.248 | all lose (Onyxia -0.08) |
| shaman tank | emergency LHW at Lowest < 30 / 45% (35) | 1.252 / 1.258 | -0.005 / +0.000 |
| shaman raid [1.392; seed 202 1.389] | NS held for HW 10 on Lowest < 35 / 45 / **55%** | 1.392 / 1.402 / **1.403** (seed 202: - / 1.395 / 1.400) | **+0.010** -> finding -2 |
| shaman raid | same at 45% with LHW 6 | 1.392 | 0 |
| paladin tank [1.152] | Holy Shock at 40 / 60% (50) | 1.154 / 1.141 | +0.002 / -0.011 |
| paladin tank | Flash of Light emergency at 25 / 45% (35) | 1.150 / 1.147 | -0.002 / -0.005 |
| Discipline [1.272] | Flash Heal emergency at 40 / 60% (50) | 1.264 / 1.260 | -0.008 / -0.012 |
| Holy tank [1.278] | Flash Heal emergency at 40 / 60% (50) | 1.269 / 1.282 | -0.009 / +0.004 (mixed by boss) |

## Conclusions
No rotation changed, so there were no new one-line A/B runs. The baseline re-checked the open findings, and they
are unchanged:

- -20261111-1 (Resto druid raid: Wild Growth uses up Nature's Swiftness): still open. Its rows are unchanged.
- -20261104-2 (Disc tank healer runs out of mana on Chromaggus): unchanged.
- -20261104-3 (Resto shaman tank healer's LHW 5 line never fires): unchanged.

1. Druid tank: the 60% NS line was tuned on true health; on seen health 45% wins every file (+0.027, finding -1).
   Druid raid 45% and shaman tank 50% are still best.
2. Shaman: the tank healer already holds NS (Tank under 50%) and holding it for a Healing Wave on Lowest Health
   loses. The raid healer autocasts NS; holding it for an emergency Healing Wave 10 at 55% gains +0.010 on two
   seeds (finding -2).
3. Consumables: no threshold or line moved with them; the other presets' emergency thresholds are at their best
   (all one-step moves within +-0.004 or losing). Consumable picks themselves are healer_gear's (measured in
   ea25e50df). Talents were not re-searched this pass.
4. Open: -2 (Discipline OOM, unchanged), -3 (shaman tank LHW 5 dead line, unchanged).

Helper: `tools/audit/healer_rotations/baseline.sh <healbench> <out.tsv> [iters]` writes summary.tsv.
