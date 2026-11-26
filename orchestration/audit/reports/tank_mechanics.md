# tank_mechanics pass 37 (2026-11-26, master f9fa79917)

Gate: 15 watched files changed since pass 36 (2af970d74). Client build still 1.60.1.71139 (wago.py latest);
Blizzard posts not re-read, the state's check was 13:01 today.

## What changed

| File | Commit | Change | Verdict | Tier |
|---|---|---|---|---|
| sim/druid/bear_attacks.go | 95962a7d3 | Lacerate hit 10% weapon x min(stacks before + 1, 5) x Rend and Tear | right (finding 20261125-1, already resolved); the hit rolls first, a miss refunds 80% and adds no stack; TestLacerate pins it | 1.5 (wowsims/forever EV-0040, in-game) |
| sim/shaman/water_shield.go, items.go | 7de8438a6 | Totem of Sustaining (23200) gives Water Shield 4 globes | right: 28856 effect 6 aura 107 misc 4 (SPELLMOD charges) points 1, mask word 1 0x20 = 408510's class mask; 408510 ProcCharges 3 | 1 (client 71139) |
| tools/tankbench/req/*.json (12) | many | Phase 0 gear, Warden races (Troll / Tauren), paladin consumables, trinket casts, bear APL knobs | no mechanic change; bear Balanced and Max threat drop Demoralizing Roar (paired +46 / +144, TestForeverBear pins Balanced) as a scored choice under the own-debuff ruling | - |

## Notes

- Totem of Sustaining: no tank request or tank gear set wears it, and the Warden casts no Water Shield, so no
  tank score moves. No test pins the 4th globe (Enhancement's shields_test.go covers own-crit globes only); that is
  the DPS mechanics pass's scope.
- Open findings unchanged: 20261124-4 (Stoneform autocast), 20261124-3 (racial tests), 20261117-1, 20261108-1,
  20261107-1.

## Result

Quiet: no new findings.
