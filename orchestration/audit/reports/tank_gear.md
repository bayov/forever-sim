# tank_gear pass 2: 2026-11-04, master b38ff16

Forced, after tank round 2 landed (b38ff16). I re-ran `python3 tools/audit/tank_gear/{warrior,paladin,druid,shaman}.py`, checked each pass-1 fix, and compared enchants across the six tank gear sets and their tankbench requests. The Tanks rows of gear_targets.md and standards 36 did not change since f62bfc74b, so step 4 stands as pass 1 left it. Standards 40 (the Warden is a mail class and need not be raid-viable) is not re-raised. No sims this pass.

## Pieces (linear % / norm % of target, recomputed)

| slot | Prot warrior | Prot paladin | Feral bear | Warden |
|---|---|---|---|---|
| Head | 99.4 / 100.4 | 99.6 / 103.4 | 100.1 / 99.3 | 99.6 / 100.8 |
| Neck | 100.1 / 96.7 | 100.5 / 100.4 | 100.1 / 95.7 | 100.5 / 100.7 |
| Shoulder | 100.3 / 102.0 | 99.6 / 103.9 | 100.3 / 101.1 | 100.4 / 102.6 |
| Back | 100.1 / 96.7 | 99.5 / 99.9 | 100.1 / 95.7 | 100.5 / 98.0 |
| Chest | 99.4 / 100.4 | 99.6 / 103.4 | 100.1 / 99.8 | 99.6 / 100.8 |
| Wrist | 100.1 / 97.6 | 99.5 / 98.1 | 100.1 / 96.1 | 100.5 / 97.9 |
| Hands | 100.3 / 101.4 | 100.6 / 96.6 | 100.3 / 100.9 | 99.9 / 103.3 |
| Waist | 99.5 / 101.5 | 99.1 / 96.1 | 100.2 / 100.6 | 100.5 / 102.9 |
| Legs | 99.4 / 103.2 | 99.6 / 102.5 | 100.1 / 99.3 | 99.6 / 100.8 |
| Feet | 100.3 / 102.8 | 99.1 / 100.0 | 100.3 / 100.7 | 100.4 / 102.6 |
| Finger | 100.1 / 95.4 | 99.5 / 104.0 | 100.1 / 95.7 | 100.5 / 100.5 |
| Finger II | 100.1 / 105.3 | 100.5 / 99.0 | 100.1 / 95.4 | 100.1 / 101.0 |
| Weapon | 100.2 | 100.4 | 100.3 (2H) | 100.4 |
| Shield | 99.9 | 99.9 | none | 100.1 |
| Tank trinket | 100.2 / 104.2 | 100.3 / 104.2 | 101.4 / 102.6 (rounding, pass 1) | 100.2 / 97.0 |
| Set total | 760.0 / 761.0 | 841.8 / 844.1 | 730.5 / 729.3 | 845.1 / 844.1 |

Every piece is within 1% linear and 6% norm; armor types, item levels and slots are unchanged and right.

## Set shares (hit-adjusted, trinkets excluded)

| spec | Stamina (row, tol 1.5) | others out of 3 | top non-Sta | hit |
|---|---|---|---|---|
| Prot warrior | 31.4 (31.4) | none (hit +1.1, crit -1.3) | 30.8% Def | 80 rating, 8% |
| Prot paladin | 32.6 (32.4) | none | 27.2% | 50 + Precision 3%, 8% |
| Feral bear | 34.8 (33.4) | none (Agi +2.8) | 34.4% Agi | 3% + flask |
| Warden | 23.9 (23.9) | none | 20.4% Int | 55 + Tidal Focus |

## Pass-1 findings

| finding | verdict |
|---|---|
| -1 paladin Stamina share | closed: 32.6 |
| -2 warrior trinket A/B | closed: A/B in docs/tank_warrior.md; Max survival wears 990617, the others 991018, as it says |
| -3 enchant values | **reopened**: paladin and Warden shields still on 929; Superior Deflection simmed for the warrior only |
| -4 bear chest and Stamina | closed: chest 72.5 / 72.44, Stamina 34.8 |
| -5 stale comments | closed: bear names 998115; warrior header 649.6 / 760.0 / 107 Def |
| -6 Warden boots | closed: simmed, Greater Stamina kept |

## Enchants across the tanks

| slot | warrior | paladin | bear | Warden |
|---|---|---|---|---|
| Shield | 1886 (+9 Sta) | **929** | n/a | **929** |
| Wrist | 1248665 Superior Deflection | 1886 Sup Sta | 1886 | 1886 |
| Back | 1219587 Agi +5 | 1219587 | 1219587 | 1889 +70 armor |
| Feet | 929 Sta +7 | 929 | 1887 Agi +7 | 929 |
| Head / legs | 1503 (+100 health) | 1507 (Sta 8) | 1508 (Agi 8) | 1507 |

929 is no longer a shield enchant in the DB. The CLI and tankbench still apply +7 Stamina by effect id, but the web UI matches enchants per slot and loads the paladin and Warden shields bare. So in the UI those presets lose 7 Stamina, and against the warrior they are 2 Stamina short (finding -3).

Druid (did not re-check this round): its enchants and consumables are unchanged since pass 1, which accepted them (Flask of Natural Accuracy, Mongoose, Phalanx, Greater Fortitude, Agility arcanum on head and legs, Agility on boots and 2H, Strength on gloves; the preset comment records its consumable A/Bs at Balanced). They were not re-simmed on the round-2 engine; the two open enchant questions (wrist -3, back -7) cover the bear too.

## New findings

- tank_gear-20261104-7 (low): cloak enchant differs across tanks with no A/B (Agility +5 on three, Superior Defense +70 armor on the Warden).

## Left open

- No sims this pass; the wrist and back A/Bs belong to the class workers.
