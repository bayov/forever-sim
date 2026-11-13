# Healer rotations pass 10 (quiet: no new findings; -20261104-2 still open)

- Master: 71562a8b0 (`last.master` was 8306f1d52). Audited 2026-11-13.
- Bench: `tools/healbench` built from this checkout, default seed, `-carry`, 1,000 iterations, each request on each
  of the 7 boss files at its own length in its role (`tools/audit/healer_rotations/baseline.sh`). New helper:
  `tools/audit/healer_rotations/ab.sh <healbench> <role> <iters> <req>... --` prints one carry row (7 files + mean)
  per request for paired screens.
- The by-hand pass 9 on tankheal (8f3f74f3f) is now on master; this pass follows both.

## What changed since 8306f1d52

- Merge of tankheal: the seen-health health values (`sim/core/apl_values_resources.go`, `heal_targets.go`), the
  druid NS comment (`sim/druid/healing.go`), the healer consumables, e05135ad3 / 8fd5ddcb9 / e0e351f2f (the fixes
  for -20261111-1, -20261112-1, -20261104-3, -20261112-2), and 7b8a41fcc (Holy priest raid Twilight Focus; Holy
  paladin raid Flash of Light 6 / 4 in place of Holy Light 9 / 8 and Divine Favor).
- Every fix above is on master (druid tank NS at 45%, shaman tank LHW 5 gone, shaman raid NS under 65% with an
  instant Healing Wave 10). Their findings, which carried `resolved:` lines under Open, are moved to Resolved.

## Baseline (carry: Onyxia / Broodlord / Chromaggus / Golemagg / Magmadar / Ragnaros / Vaelastrasz, mean)

| request | carry | mean |
|---|---|---|
| druid_restoration | 1.403 0.893 1.023 1.951 2.045 0.634 0.932 | 1.269 |
| druid_restoration_raid | 0.991 0.924 1.094 2.630 1.829 1.278 1.152 | 1.414 |
| paladin_holy | 1.241 0.883 0.908 1.806 1.867 0.568 0.789 | 1.152 |
| paladin_holy_raid | 1.018 0.912 1.146 2.702 1.914 1.283 1.035 | 1.430 |
| priest_discipline | 1.363 0.958 0.871 2.027 2.143 0.602 0.938 | 1.272 |
| priest_holy | 1.378 0.931 0.970 2.000 2.154 0.600 0.910 | 1.278 |
| priest_holy_raid | 0.988 0.906 1.065 2.607 2.062 1.340 1.262 | 1.461 |
| shaman_restoration | 1.340 0.978 0.984 1.938 2.062 0.587 0.911 | 1.257 |
| shaman_restoration_raid | 0.946 0.885 1.115 2.674 1.887 1.236 1.138 | 1.412 |

Against the ref (written by 8f3f74f3f / a2d1dceaa): 49 of 63 rows byte-identical. Only the two raid presets that
7b8a41fcc changed moved: paladin raid mean 1.400 -> 1.430 (Vaelastrasz +0.078, Chromaggus +0.059, Ragnaros -0.007),
Holy priest raid 1.440 -> 1.461 (Vaelastrasz +0.137, Broodlord -0.017). Both match the commit's numbers and are
rotation / talent changes, so no finding. Disc tank still OOM on Chromaggus at 101 s (341 mana; -20261104-2).

## Changes tried (paladin raid, the one rotation rewritten this round)

Divine Favor now has no line, so `autocastOtherCooldowns` raises it on cooldown and the next Flash of Light or Holy
Shock takes the crit. Holding it (raid role, own lengths, 1,000 it, default seed):

| variant | carry | mean |
|---|---|---|
| master (DF on cooldown) | 1.018 0.912 1.146 2.702 1.914 1.283 1.035 | 1.430 |
| DF when the lowest is under 50% (before Holy Shock) | 1.013 0.906 1.131 2.701 1.913 1.267 1.029 | 1.423 |
| DF when the lowest is under 35% (before the emergency FoL) | 1.011 0.910 1.139 2.684 1.906 1.268 1.031 | 1.421 |

Both lose; on cooldown stays.

Mana at 300 s (500 it): paladin raid Magmadar 1.672 carry / 8,173 mana left, Ragnaros 1.139 / 8,071; Holy priest
raid Magmadar 1.766 / 7,229, Ragnaros 1.188 / 6,057 (overheal 52-54%). Neither runs dry. The paladin ends near full
mana, but the bigger heal it dropped (Holy Light 9 / 8 with Divine Favor) lost 0.03 carry in 7b8a41fcc's screen, so
this is cast-time bound, not a finding.

## Open items

- -20261104-2 (Disc tank OOM on Chromaggus) unchanged; needs a worker.
