# Healer fairness pass 8 (2026-11-12, master cb0d928fd6)

Result: quiet. No new finding. Still open: -20261111-1 (high, druid raid NS on Wild Growth), -5 (high, paladin raid
preset), -4 (low). Due because 3 watched files changed (33db32b..cb0d928fd6, 3 commits). What they do:
- sim/druid/talents.go: comment only (Rend and Tear tooltip note).
- sim/priest/priest.go: TalentTreeSizes Shadow 18 -> 17. The Holy and Disc talent strings have no Shadow tree part
  (005005031330011531-0350510303, 00500503113-03505103030121531), so their parsing is unchanged.
- sim/shaman/talents.go + proto: improved_fire_nova renamed trial_by_fire (same field 9). Both shaman healer strings
  start with an empty Elemental tree, so nothing reaches them.
- tools/healbench, req and encounters unchanged, so section 1 stands.

Spot check, healbench built from cb0d928fd6, 1,000 iterations, -carry: identical to bench_9ba00c2.tsv in HPS and carry
for priest_holy (tank Broodlord 800, 0.930), priest_discipline (tank Broodlord 656, 0.971), shaman_restoration (tank
Broodlord 662, 0.974), priest_holy_raid (raid Onyxia 280, 0.978), shaman_restoration_raid (raid Onyxia 267, 0.920).
Rankings as in pass 7: raid Holy 1.437, paladin 1.413, shaman 1.395, druid 1.391 (druid column from
`ref/healer_fairness/druid_raid_707ae92.tsv`). The tables below are pass 2's.

## 1. Same terms (the 9 requests)

Identical in all 9 (md5 of each block): consumables, raid buffs, party buffs, debuffs, encounter; level 60,
Engineering + Alchemy. Differences, all as in pass 1 and fine: race (Human priest/paladin, Skyborne Windshaper
shaman/druid), paladin's off-hand 990314 and Libram, weapon enchants, mana cooldowns by class. Slot 16 is now
stat-less for all four (990514 has no stats). healbench overrides reaction (0.3 s ±25%), stopcast 0.5, assignment.

## 2. The numbers (healbench -carry, 1,000 iterations, each file at its own length; 0facd9428 and a079415316 give the same rows)

Tank role:

| Boss | Shaman | Holy priest | Disc priest | Paladin | Druid |
|---|---|---|---|---|---|
| Onyxia | 1.346 | 1.379 | 1.378 | 1.257 | 1.327 |
| Broodlord | 0.974 | 0.930 | 0.971 | 0.895 | 0.875 |
| Chromaggus | 0.969 | 0.955 | 0.866 (OOM 101 s) | 0.911 | 1.029 |
| Golemagg | 1.915 | 1.969 | 1.991 | 1.911 | 1.904 |
| Magmadar | 2.019 | 2.104 | 2.101 | 1.875 | 2.030 |
| Ragnaros | 0.577 | 0.592 | 0.596 | 0.572 | 0.627 |
| Vaelastrasz | 0.894 | 0.906 | 0.927 | 0.812 | 0.852 |
| **Mean** | **1.242** | **1.262** | **1.261** | **1.176** | **1.235** |

Pass 1 (c8f5858bd): shaman 1.103, Holy 1.220, Disc 1.149, paladin 1.028, druid 1.040. Every column rose 0.09-0.20:
the tanks now face the boss (87f28803b: parries and blocks cut the tank's damage) and the gates were re-tuned. The
Holy priest's lead over the shaman fell from +0.117 to +0.020, as pass 1 predicted (+0.03 to +0.05) once the bench
and gates were fixed. The paladin is last, 7% under the priests; no cell is 15% off its file's mean (largest: Disc
Chromaggus -8%, where it runs out of mana at 101 s, already healer_rotations-20261104-2, still open; Holy ends that
file with 677 mana). All tank healers end at carry with 3,000-7,200 mana on 6 of 7 files: cast-time bound, not mana.

Raid role:

| Boss | Shaman | Holy priest | Paladin | Druid |
|---|---|---|---|---|
| Onyxia | 0.919 | 0.978 | 0.984 | 0.997 |
| Broodlord | 0.902 | 0.901 | 0.905 | 0.920 |
| Chromaggus | 1.090 | 1.070 | 1.081 | 1.109 |
| Golemagg | 2.662 | 2.597 | 2.699 | 2.660 |
| Magmadar | 1.809 | 2.054 | 1.918 | 2.002 |
| Ragnaros | 1.235 | 1.308 | 1.301 | 1.270 |
| Vaelastrasz | 1.152 | 1.152 | 1.002 | 1.102 |
| **Mean** | **1.396** (1.395 on 9ba00c2dc) | **1.437** | **1.413** | **1.437** |

Pass 1: 1.373 / 1.414 / 1.413 / 1.413. The shaman is still about 3% behind (last on 3 files, -12% to the priest
on Magmadar). Nobody OOMs. Overheal: druid 45-78%, priest 42-72%, shaman 22-73%, paladin 17-23%.

### The paladin's raid preset (new finding -5)

Pass 1's open question 4 (the tank preset beat the raid preset in the raid role on 3 files) re-measured on all 7:

| Variant, raid role, paired seeds | Mean | Note |
|---|---|---|
| paladin_holy_raid (shipped) | 1.413 | |
| + the tank preset's Oath of Light line | 1.413 | no change; Oath targets Tank 0 |
| − Holy Light 9/8 and Divine Favor | 1.426 | +0.013 |
| paladin_holy (tank preset) | 1.438 | +0.025, up on 6 of 7 |
| tank preset − Holy Light lines | 1.438 | identical: they never fire without a tank |
| tank preset, Tank targets → MostHurt | 1.417 | Holy Light at the most hurt costs again |

So a Flash of Light raid paladin (LoH, Holy Shock, FoL under 35%, FoL 6 at the most hurt at 0.7x) carries 1.438,
level with the priest and druid; the shipped preset's Holy Light lines cost 0.025. Finding -5, high.

### Against the logs

Unchanged: the boss files keep healer counts by class (5.1 priests, 3.0 shamans, 1.7 paladins, 1.2 druids of 11)
but no HPS by class (finding -4). The sim's raid role has all four within 3%, which fits raids picking by numbers.

## 3. What a role rewards

Each column is the preset for its role at its file's own length. The tank bench now keeps the 7 fillers idle, so a
jumping heal has a raid. The tank-presets-in-raid cross-check was re-run only for the paladin (above); the other
classes' presets didn't change role logic. Not re-run: raid presets in the tank role.

## Reproduce

- Build: master 9ba00c2dc; `PATH=$PATH:~/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto &&
  go build --tags=with_db -o healbench ./tools/healbench`
- Per cell, from the worktree: `healbench -encounter tools/encounters/<boss>.json -role <tank|raid> -carry -iters 1000
  tools/healbench/req/<req>.json`; tank role for the 5 non-raid requests, raid role for the 4 `_raid` ones.
- Rows: `ref/healer_fairness/bench_9ba00c2.tsv` (the 63 shipped rows, this pass); the paladin variants of finding
  -5 stay in `ref/healer_fairness/bench_0facd94.tsv`. The bench takes about 40 min on 4 shared cores: run it in the
  background.

## Open items for the next pass

- Re-bench the druid raid column after finding 20261111-1's fix (expect about 1.437).

- Re-bench the raid column after finding -5's re-tune.
- Disc tank Chromaggus OOM at carry (healer_rotations-20261104-2): check after its fix.
- Raid presets in the tank role were not re-run this pass.
