# Fairness pass 10 (2026-10-27, master 06a8987)

Result: findings (1 new, high). Full pass: sim/core racials, mage and warlock code, presets and requests changed.

## Why it ran

Since d3df924: the 2026-10-26 racial pass (500e3a147, sim/core/racials_forever.go), Missile Barrage off
Arcane Missiles' missiles (a52cdcf53), Ignite off Arcane Instability / Arcane Power (db6e43f98), Shadow-only
Spellstone (d6fcdd3ef), the Ret Shock preset names (723f57838), and 4273ee51d's re-sim of every build.

## What I checked

- Gear: `audit1b.py`, all 19 sets "5a PASS | 1b PASS".
- Published against current: the shared 20k run (`$AUDIT_WORK/runs/master`) matches the deployed builds.json within
  0.05 on all 58 builds. The metas' published.dps are stale on all 58 (results-20261027-1, confirmed here; not
  re-filed). `table.py` now compares against builds.json and accepts "Night Elf" with a space in a meta.
- Race sweep (5000 it, seed 1), table in `ref/fairness/table_baseline.md`; 20k paired seeds on the 14 largest gaps.
- Shared code: the racial changes are race-gated only. Orc Command reaches hunter and warlock pets
  (BM pet 333 vs 317 as Undead, Demonology 85 vs 81). Ignite / Spellstone changes are class-local.

## Movement (vs baseline 99bd53a, best race)

| Build | Before | Now | Cause |
|---|---|---|---|
| Arcane | 983.2 | 902.0 | Missile Barrage ruling a52cdcf53 |
| Frost | 947.1 | 912.1 | same (Frost's APL casts Arcane Missiles) |
| Warlocks | | -0.6% to -1.2% | racials; Spellstone Fire half for Destruction Fire |
| everything else | | -0.1% to -1.0% | racial pass 500e3a147 |

No unexplained move.

## Finding

fairness-20261027-1 (high): report races were kept through the racial pass. Best race vs report race, 20k
paired seeds: BM Orc +12.2, Balance Wrath High Order +10.2, Balance Night Elf +8.5, Survival Orc +8.4 (passes
Affliction, #8 to #7), MM Orc +6.1, Seal Fate Daggers / Swords Undead +4.2 / +4.0, Mutilate +2.5, Arcane Gnome
+2.4, Arms Undead +2.3, Combat Daggers +2.1, Ret Shock (Ret side) Undead +2.1, Enhancement Windshaper +1.3.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, FTT + Windfury, boss partial resist, Rage k and pet scaling.
