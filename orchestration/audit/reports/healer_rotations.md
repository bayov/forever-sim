# Healer rotations pass 2 (findings: 1 new low; 1 resolved; 1 medium still open)

- Master: 0facd9428 (`last.master` was 1e9eb71f7). Audited 2026-11-04.
- Bench: `tools/healbench` built from this checkout. Healbench's default seed, `-carry`, 1,000 iterations, each
  request on each of the 7 boss files at the file's own length in its role. Script:
  `tools/audit/healer_rotations/baseline.sh <healbench> <out.tsv>` (new; it writes the summary.tsv format).
  Mana side: Golemagg and Ragnaros at `-dur 180` and `-dur 300`, 500 iterations.
- New baseline: `ref/healer_bench/summary.tsv` (63 rows).

## What changed

- Every healer APL and request was re-tuned: shaman 324d64834; paladin 0b15937a8 (tank heals at 70% of the
  heal's expected healing missing); druid bc8e5473b; priest dda14db65 and a7d9b57ef. All 9 requests still
  embed exactly their UI APL (checked by comparing the priority lists).
- The bench's tank side changed: 87f28803b (the tank faces the boss, so it parries and blocks, and a parry
  hastes the swing); 7b6623f27 (the tank role keeps the raid's fillers, idle, which is healer_fairness-20261104-1);
  Onyxia's `parryHaste` was turned off; and the tank_mechanics work.
- Attribution: I ran the old APLs from 1e9eb71f7 on today's bench. They give nearly the same carry as the new
  APLs: Holy priest Chromaggus/Golemagg/Magmadar/Onyxia 0.955/1.969/2.094/1.365 (new APL 0.955/1.969/2.104/1.379);
  Disc Chromaggus 0.854 (new 0.866); shaman Golemagg 1.914 (new 1.915); paladin Golemagg 1.846 (new 1.911,
  so its re-tune gives +0.065). So the bench, not the rotations, moved the tank rows, by +0.04 to +0.34. Those
  are intended tank-side fixes, so I filed no finding. The one row that fell, Holy priest Chromaggus 1.002 -> 0.955,
  comes from the idle-filler fix: Prayer of Mending no longer bounces only between the tank and the priest, and
  the commit itself measured that fall. The raid rows moved by at most 0.04, except druid raid Magmadar (+0.133,
  its own re-tune).

## The baseline (carry, 1,000 it, 0facd9428)

| Boss (tank / raid length) | Tank: shaman | Holy priest | Disc priest | Paladin | Druid | Raid: shaman | Holy priest | Paladin | Druid |
|---|---|---|---|---|---|---|---|---|---|
| Onyxia (35 s / 137 s) | 1.346 | 1.379 | 1.378 | 1.257 | 1.327 | 0.919 | 0.978 | 0.984 | 0.997 |
| Broodlord (41 s / 44 s) | 0.974 | 0.930 | 0.971 | 0.895 | 0.875 | 0.902 | 0.901 | 0.905 | 0.920 |
| Chromaggus (115 s / 115 s) | 0.969 | 0.955 | 0.866 | 0.911 | 1.029 | 1.090 | 1.070 | 1.081 | 1.109 |
| Golemagg (56 s / 56 s) | 1.915 | 1.969 | 1.991 | 1.911 | 1.904 | 2.662 | 2.597 | 2.699 | 2.660 |
| Magmadar (53 s / 54 s) | 2.019 | 2.104 | 2.101 | 1.875 | 2.030 | 1.809 | 2.054 | 1.918 | 2.002 |
| Ragnaros (71 s / 97 s) | 0.577 | 0.592 | 0.596 | 0.572 | 0.627 | 1.235 | 1.308 | 1.301 | 1.270 |
| Vaelastrasz (17 s / 31 s) | 0.894 | 0.906 | 0.927 | 0.812 | 0.852 | 1.152 | 1.152 | 1.002 | 1.102 |

## Mana side (500 it, carry / mana left, OOM time when the healer runs dry)

| Request | Golemagg 180 s | Golemagg 300 s | Ragnaros 180 s | Ragnaros 300 s |
|---|---|---|---|---|
| shaman_restoration | 1.481 / 4089 | 1.203 / 3840 | 0.502 / 6085 | 0.442 / 6085 |
| priest_holy | 1.195 / 3125 | 0.957 / 1043 | 0.418 / 5165 | 0.327 / 5165 |
| priest_discipline | 1.112 / 2089 | **0.920 / 885, OOM 210 s** | 0.402 / 4860 | 0.331 / 4860 |
| paladin_holy | 1.473 / 6249 | 1.344 / 5543 | 0.501 / 5815 | 0.480 / 5815 |
| druid_restoration | 1.549 / 5301 | 1.212 / 5071 | 0.572 / 5061 | 0.476 / 5061 |
| shaman_restoration_raid | 2.083 / 7602 | 1.897 / 7588 | 1.141 / 7322 | 1.055 / 7325 |
| priest_holy_raid | 2.181 / 7441 | 2.089 / 7412 | 1.238 / 6425 | 1.178 / 6103 |
| paladin_holy_raid | 2.183 / 7969 | 2.016 / 7963 | 1.199 / 7820 | 1.170 / 7813 |
| druid_restoration_raid | 2.215 / 6965 | 1.990 / 6924 | 1.197 / 5391 | 1.119 / 4491 |

On Ragnaros the tank-role rows are the same at 180 and 300 s, because the tank dies before 180 s at the carry
scale. No raid healer runs low. At the boss lengths, the Disc priest (365 mana, OOM at 101 s) and the Holy priest
(677 mana) end Chromaggus nearly dry. The others end with 2,000 or more.

## Rotation lines checked

- The Holy priest's dead lines (-1) are gone: no Power Infusion, and the raid healer is down to 6 lines. Resolved.
- Disc (-2): the a7d9b57ef mana plan doesn't stop the OOM on today's bench. New evidence is under the finding.
  The one-line variants (shield gate 0.9, emergency Flash Heal at 35%) were started but cut off at wrap-up and
  are not measured. They are the next pass's first item.
- Shaman tank healer: Lesser Healing Wave 5 (line 9) has the same mana-only condition as rank 6 on line 8, and
  logs show 0 casts. Filed as -3 (low).
- Shaman Mana Spring Totem `remainingTime >= 20s` is a "don't drop a totem as the boss dies" gate, not timing
  from a coincidence of the fight length. Druid Innervate at 35% and shaman Mana Tide at 35% mana are reasonable.
  Nature's Swiftness fires on the tank under 50% (2 casts in a 300 s log).

## Open items for the next pass

- Measure the Disc variants on Chromaggus (4,000 it) and Golemagg at 300 s. Look at the Holy priest's
  Chromaggus mana (677 left).
