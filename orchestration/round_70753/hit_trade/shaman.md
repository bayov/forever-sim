# Round 70753 phase 5, hit trade: shaman

All six builds searched (talent search, mean of 120 / 180 / 300 s, 3000 iterations, pins kept) and confirmed at 20,000
iterations, seed from the request, per length. DPS is mean (stderr). "old" = current talents, no -hit-steps; "new" = picked
talents with `-hit-steps <set> -hit-cap <cap> -hit-step <step>`. Gear step files are `<set>_hit<N>.gear.json` next to the set.

Significant = the difference of old and new exceeds 2 combined standard errors at that length.

| build | old talents | new talents | step (gear set) | hit before -> after (cap) |
|---|---|---|---|---|
| enhancement-level60 | 353003102-055030110051012251-02 | 353003102-055030110051012151-03 | -3 (level60_phase1_synthetic_hit-3) | 10 -> 11 with Tidal Focus 3/5; step -3 brings it to 8 (cap 8) |
| enhancement-imp-gw | 353003102-055030110051212151-01 | unchanged | -1 (level60_phase1_synthetic_hit-1) | 9 -> 8 (8) |
| enhancement-water-shield | 052002-054030210051012151-053020001 | 052002-055030110051012151-053020001 | -4 (level60_phase1_synthetic_hit-4) | 13 -> 9 at -4 (cap 8, cap step -5) |
| enhancement-mana-tide | 03-055030310051012051-053051001001 | 03-055030210051012151-053111031001 | -4 (level60_phase1_synthetic_burning_totem_hit-4) | 13 -> 9 at -4 (cap 8, cap step -5) |
| elemental-level60 | 5501331003001351-05500101-053 | 5501331003001351-05500001-05301 | +0 (level60_phase1_synthetic) | 16 -> 16 (16) |
| elemental-level60-mana-tide | 5501331003001351-04-053150001001 | unchanged | +0 | 16 -> 16 (16) |

## Talent changes in words

- Enhancement Level 60: Improved Stormstrike 2 -> 1, Tidal Focus 2 -> 3 (hit 10 -> 11 on the gear, so a deeper step).
- Imp GW: none (search found nothing; the step -1 alone).
- Water Shield: Enhancement tree 054030210051012151 -> 055030110051012151 (one point from one talent to another; a mana tie-break move). The step is most of the gain.
- Mana Tide: Enhancement tree 055030310051012051 -> 055030210051012151 and Restoration 053051001001 -> 053111031001 (points moved off Improved Healing Wave / Totemic Focus, partly mana tie-break moves among equal DPS); manaTideTotem, waterShield and spiritWeapons kept.
- Elemental Level 60: Enhancement 05500101 -> 05500001 and Restoration 053 -> 05301. Search start 958.0 and end 957.7: a tie, not a gain.
- Elemental Mana Tide: none.

## DPS (20,000 iterations)

| build | 120 s old -> new | 180 s old -> new | 300 s old -> new | significant? |
|---|---|---|---|---|
| enhancement-level60 | 1169.1(0.6) -> 1176.9(0.6) | 1159.4(0.5) -> 1165.3(0.5) | 1140.5(0.4) -> 1144.1(0.4) | yes at all three (+7.8, +5.9, +3.6) |
| enhancement-imp-gw | 1161.8(0.6) -> 1162.6(0.6) | 1149.8(0.5) -> 1151.0(0.5) | 1128.7(0.4) -> 1130.0(0.4) | marginal (+0.8, +1.2, +1.3: about 1 to 2 SE); step only |
| enhancement-water-shield | 1118.9(0.6) -> 1125.3(0.6) | 1111.3(0.4) -> 1117.4(0.5) | 1108.1(0.3) -> 1113.4(0.3) | yes at all three (+6.4, +6.1, +5.3) |
| enhancement-mana-tide | 1065.1(0.5) -> 1068.6(0.5) | 1055.1(0.4) -> 1060.3(0.4) | 1046.7(0.3) -> 1055.2(0.3) | yes at all three (+3.5, +5.2, +8.5) |
| elemental-level60 | 979.7(0.3) -> 979.0(0.3) | 970.1(0.2) -> 969.5(0.2) | 924.6(0.2) -> 924.5(0.2) | no change (-0.7, -0.6, -0.1): keep the old talents |
| elemental-level60-mana-tide | 962.7(0.3) | 952.7(0.2) | 935.1(0.2) | no change (identical talents and step; one run) |

## Step check after the coordinator's rotopt fix (a6c243fd5)

The searches ran on a rotopt built before the fix (span centred on the cap step only). For the four Enhancement builds I
ran, on the picked talents with the fixed binary, paired runs (same seed, 20,000 iterations, 120 / 180 / 300 s) of every
step from the pick up to +1. Mean of the three lengths (stderr of the mean about 0.3 per length):

| build | pick | steps tried | result |
|---|---|---|---|
| enhancement-level60 | -3: 1162.1 | -2: 1161.2, -1: 1160.0, +0: 1158.3, +1: 1159.6 | pick stays; -2 is 1.2 behind (under 2 SE gain needed) |
| enhancement-imp-gw | -1: 1147.9 | +0: 1146.8, +1: 1148.4 | pick stays; +1 is +0.5, under the 2 SE bar (about 0.9) |
| enhancement-water-shield | -4: 1118.7 | -3: 1119.0, -2: 1116.4, -1: 1115.1, +0: 1113.3, +1: 1114.3 | pick stays; -3 is +0.3, a tie |
| enhancement-mana-tide | -4: 1061.4 | -3: 1060.4, -2: 1058.3, -1: 1056.4, +0: 1055.4, +1: 1055.3 | pick stays |

No step past the pick towards +1 beats it by 2 SE. The elemental builds sit on the cap with no span, so they were not rechecked.

## Odd things

- Water Shield and Mana Tide: the cap step is -5 (Tidal Focus 5/5 brings 13% to 8%), the search wore -4, i.e. 9% total hit,
  and -3 ties it. The set has -5 as its last step; -6 does not exist. A step past the cap bought nothing.
- Imp GW and Level 60 steps -2 .. +1 are all within a few DPS, so the hit trade is worth about 1 to 8 DPS per Enhancement build. The gain on Level 60 comes
  mostly with the extra Tidal Focus point and Improved Stormstrike moved out.
- The Mana Tide search ran 6 rounds, many of them mana tie-break moves among equal DPS (round 6 returned to round 4's 1063.0); the
  talent string should be taken as one of several equal ones.
- Elemental Level 60: the search's "new" talents are a tie within noise (-0.6 at 180 s); the old talents are the ones to keep.
- Old DPS at 180 s match the published numbers (Elemental 970.1 / 952.7, Enhancement 1159.4 / 1111.3 / 1055.1 and Imp GW 1149.8).

## Commands

Run from the repo root with GOMAXPROCS=2. Knobs are the `gen`/`gen_ele` knobs in `tools/rotopt/shaman_presets.sh`
(`optimized` for Level 60 and Imp GW, `magma` for Water Shield, `mana_tide` for Mana Tide, `ele60` plus
`lightningBoltDownrank=1,lightningBoltDownrankManaPerSec=30` for Elemental and `...=2,manaTideMana=10,...ManaPerSec=20` for Elemental Mana Tide).
`E` = `ui/enhancement_shaman/gear_sets/level60_phase1_synthetic` (Mana Tide: `..._burning_totem`), `L` = `ui/elemental_shaman/gear_sets/level60_phase1_synthetic`.

- Search: `rotopt -template shaman_enh|shaman_ele -settings tools/dps_report/req/<id>.json -level 60 -knobs "<knobs>" -durs 120,180,300 -combined -talent-search -require <pins> -iters 3000 -hit-steps <set> <-hit-cap melee:8 -hit-span 1 | -hit-cap spell:16>`
  Pins: Level 60 spiritWeapons; Imp GW spiritWeapons,improvedGhostWolf; Water Shield spiritWeapons,waterShield; Mana Tide spiritWeapons,waterShield,manaTideTotem; Elemental none; Elemental Mana Tide manaTideTotem.
- Old: `rotopt -template ... -settings ... -level 60 -knobs "<knobs>" -eval -confirm-iters 20000 -durs <d>`
- New: the same plus `-talents <picked> -hit-steps <set> <cap flags> -hit-step <step>` (steps: -3, -1, -4, -4, +0).
- Step check: the New command with `-hit-step` set to each of the steps listed above.
