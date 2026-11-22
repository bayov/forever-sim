# Report regeneration, 2026-11-22

The paladin Two-Handed Weapon Specialization hotfix (1c80dafff, standards 41), PRs #133 to #136 (merged at 7249392f2) and the
regeneration (master 62e0e98c, deploy 0bb7fb55). The raid sim's data was refreshed in 6ce47850c (deploy 36e77585).

## Moves at 180s

- Ret: +0.5% (Level 60, Sub Holy) and +0.9% (Shock Ret), duty variants alike. Ret vs Undead 1137.4 to 1151.0, re-optimized.
- Rogues: +0.4 to +0.6% (Shank procs).
- Fury and Fury 2H: +0.1 to +0.2% (Flurry with Heroic Strike and Cleave).
- Tanking and Healing: no moves. #134 and #136 move nothing.

## Rulings (user, 2026-11-22)

1. Ret vs Undead keeps Holy Strike ahead of the seal and twist lines (`holyStrikeFirst=1`), though it gains only 0.2 to 0.6
   DPS, under the 1 DPS bar. It won every cell at 120, 180 and 300s on both seed sets.
2. Seal Fate Swords stays Undead, although Gnome ties it (943.0 against 942.9).

## Settled

- The BM cat pet share in `ui/raid_sim/data/builds.json` is 17%. The 28% predates b69e037f5, the haste GCD report data.
