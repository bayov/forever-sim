# Mechanics pass 110 (2026-11-26)

- Master: f9fa79917 (last audited 2af970d74). Client 1.60.1.71139 (unchanged; probe of 300 builds found none newer).
- Result: findings (1 medium).

## Get current

- Blizzard: blue tracker and the hotfix thread (2403040, last post 16, 2026-11-25 23:25, not edited since) have nothing newer than the last pass read. Development notes thread 2360696 unchanged (8 posts, 2026-11-04).
- Daily update f9fa79917 (TASKS.md): build 71139, Revelation per wave, Chain Lightning Clearcasting, Blackhand's Breadth and Sayge notes. Blackhand's Breadth (mechanics-20261125-2) and Sayge (mechanics-20261124-1) are already open here; Chain Lightning Clearcasting is a shaman class item. Revelation was not filed anywhere: re-checked against the sim, now mechanics-20261126-5.

## Repo since 2af970d74

| Commit | Change | Check | Verdict |
|---|---|---|---|
| 7de8438a6 | Water Shield 4 globes with Totem of Sustaining (23200) | 71139 spell 28856: effect 6, aura 107 (flat modifier) misc 4 (charges) bp 1, mask word 1 0x20 (Water Shield); tier 1 | Verified |
| a589203b7 | Lacerate hit 10% weapon x min(stacks after cast, 5), was x stacks before | wowsims/forever EV-0040 in-game test (15 hits, within 0.5% of the 1:2:3:4:5 model); tooltip says "per existing application" | Log-sourced, accepted; bear only, no DPS preset |
| a589203b7 | Rest of the Phase 0 tags merge | gate lists only bear_attacks.go and the two shaman files as changed sim code | No change |

## Revelation (mechanics-20261126-5)

Probe: the next hit of the channel holding the charge spends it ($AUDIT_WORK/mechanics/revelation_channel.patch). 20,000 iterations, seed 1, paired, tools/dps_report/req requests.

| Build | Master | Probe | Delta |
|---|---|---|---|
| mage__arcane | 833.89 | 828.42 | -5.47 (-0.66%) |
| priest__shadow | 861.81 | 861.81 | 0 |
| druid__balance | 801.90 | 801.90 | 0 |

The per-wave half (area spells roll once a wave) moves no single-target preset; not measured.
