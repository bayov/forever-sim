# Results pass 106 (master 85691655a)

Range 0404484fb..85691655a. Watched paths changed: sim/core, sim/hunter, items, procs, warrior/tank, casters (efeb06b95, 60fde111f,
295206ddf, b663145ad, 2f42f4a44 via merges, 22d2cf952), and tools/rotopt/mksettings (spec_hunter.go, spec_feral.go). No request change.

Shared run (68 builds, 20k, seed 1) vs the previous binary (0404484fb, same requests): 64 of 68 moved by more than 1 DPS.

| Build | 0404484fb | efeb06b95 | merged batches | 85691655a |
|---|---|---|---|---|
| hunter beast-mastery | 1002.9 | 1002.9 | 1054.8 | 1044.1 |
| hunter survival | 919.9 | 919.9 | 949.1 | 949.1 |
| hunter lone-wolf | 1164.9 | 1164.9 | 1164.9 | 1152.0 |
| priest shadow | 854.2 | 854.2 | 869.2 | 869.2 |
| shaman elemental-level60 | 852.4 | 854.6 | 867.3 | 867.3 |
| rogue mutilate | 821.1 | 821.1 | 795.3 | 795.3 |
| warrior fury | 1014.9 | 1014.9 | 997.3 | 1001.8 |
| warlock affliction_shadow_bolt | 800.0 | 800.0 | 800.0 | 794.2 |

The casters batch moves none of these; the hunter, items, procs and tank batches (merge becbcd30d) move most; 22d2cf952 moves
hunter Dark Bomb, demon crit, Windfury Totem and Shield Slam rules. Other rogue builds -11 to -17, Enhancement -22 to -28 vs metas.
Moves under 1.5% were not individually bisected.

Drift: new. Hunter and feral requests carry EZ-Thro Dark Bomb; mksettings writes Dense Dynamite (hunters) and Goblin Land Mine
(feral). Regenerated: BM +9.6, MM +9.3, Lone Wolf +11.5, feral +3.8 (finding results-20261125-1). Feral sapper line unchanged.
Unread fields not run (no request change). Metas and builds.json need the full re-sim (results-20261104-2).
