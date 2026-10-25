# Gear pass 39 (master 54878fa3d, client 1.60.1.70532)

Baseline compared: daba9d1a5 (pass 38). Quiet: no findings.

- Changed gear files: 9 synthetic item files, 10 level60/p0 sets (shaman Enh x2, Feral, Ret x2, rogue x3, Fury, tank warrior), sim/druid and sim/paladin items.go (Idol of the Dream, Talons of Wrath: relic effects, not budget).
- Rulings: standards 4 (physical hit cap 8% yellow / 27% dual-wield white, 2026-10-23), 8a/10/11 (flask, elixir text), 26 (must-take talents). Only 4 touches gear: hit shares in lin2 fell 1 to 1.5 points on every physical set and the other stats rescaled up, as expected.
- lin2.py vs the d2633f4a4 reference: only hit/rescale shifts and three new set lines (Enh burning totem, Ret crusader duty/undead, rogue subtlety). No stat more than 2 points from target (largest: Enh Crit 5.6/4.6, Rogue AP 5.9/7.1, a caster SP 36.4/34.5). Weapons 1.037 and 1.041 x L(76) for Enh/Ret 2H, inside tolerance.
- wepcheck vs pass 38: no CHANGED / NEW / DPS-DRIFT. trinkets.py: no '!' (on-use averages 26.8 / 37.2).
- Not re-run: ref/gear baselines (left at d2633f4a4; hit shares now differ by the ruling above).
