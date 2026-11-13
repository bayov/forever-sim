# Rotations pass 62 (2026-11-13, master 71562a8b0)

Last audited master 8306f1d52. 55 commits; what reaches the DPS raid presets:

| commit | change | reaches DPS presets |
|---|---|---|
| f1c07e13a | Paladin Judgement (Command, Righteousness, Fury) is a ranged special on the ranged table under Forever (Nov 11 hotfix; `sim/paladin/judgement.go` `rangedJudgement`) | yes (every Retribution build, about -5) |
| 8f5de1c87 | Survival: Wing Clip mana floor (50%) dropped (`ui/hunter/apls/sv.apl.json`, 3 reqs); Fury on Battle Shout duty: fillerGap 1, Heroic Strike from 60 Rage in Reckless Swings; Combat Swords is the rogue default again | yes (worker-measured on two seeds) |
| others | tank and healer APLs (feral tank, prot warrior, Warden, holy paladin, resto druid/shaman), healer targeting (`apl_values_resources.go`, `heal_targets.go`, `druid/healing.go`) | no |

`apl_logic_diff.py 8306f1d52 71562a8b0`: 13 APL files; the only DPS ones are `sv.apl.json` and
`level60_fury_battle_shout.apl.json`, both from 8f5de1c87 and backed by its paired runs.

## Baseline move (3000 it, seed 424242, shared CLI 71562a8b0358)

New baseline `ref/rotations/baseline_71562a8_3000it_seed424242.tsv`; every other preset within 0.5.

| build | 8306f1d | 71562a8 | move |
|---|---|---|---|
| paladin (all 10 Retribution builds) | 963-1131 | 958-1126 | -3.0 to -6.0 |
| hunter__survival / -tallstrider / -wolf | 1015.1 / 981.0 / 990.0 | 1015.6 / 981.7 / 991.0 | +0.6 / +0.7 / +0.9 |
| warrior__fury-duty-battle-shout | 1044.2 | 1047.8 | +3.6 |

The shared run's large drops against the published report (Fury 2H -10%, Retribution -6 to -9%) come from
the Windfury Totem rule already audited in passes 60-61; this pass adds only the ~5 above for Retribution.

## Conclusions

- No new findings. Judgement as a ranged attack costs Retribution about 5 DPS evenly across builds; it changes
  no APL choice in an obvious way (Judgement lines are seal-twist gated, not proc-driven). Not A/B'd this pass.
- Open findings re-checked on master, all still unapplied: 20261112-1 (rogue poisons: req still Deadly MH /
  Instant OH), 20261109-1 (Fury 2H 90-rage Heroic Strike line still present), 20261105-1 (feral FB gate still 45).
  None of their inputs moved this pass, so their measurements stand.

## Open items

- Retribution: whether Judgement priority or the twist timing should move now that Judgement is ranged
  (expected small; untested).
- Carried from the pass doc: Balance Night Elf vs Tauren, Elemental Mastery hold, Enhancement Flametongue
  Weapon with Windfury Totem, warlock duty-curse screens.
