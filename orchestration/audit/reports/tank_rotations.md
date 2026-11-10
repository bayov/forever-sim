# tank_rotations: pass 7 (2026-11-10, master ca5e35c3f)

Due on sim/shaman/flametongue_weapon.go and talents.go (bcc5d6cd4: Maelstrom Weapon stacks on every melee hit under
Forever, Flametongue rank 5 cut by 7). No tank APL, request, rotopt template, encounter file, tankbench or tankscore
changed. The Warden requests use Rockbiter and cast no Lightning Bolt; a tankbench built at bcc5d6cd4's parent
(d7bb9ab80) gives all 21 Warden rows identical to master, so that commit reaches no tank.

Pass 6 benched 9ba00c2dc but recorded 91ad19535, so it never benched 0ce9dbf40 (Windfury Totem procs only from white
main hand swings under Forever, user ruling 2026-11-08 "for now"). That change moves every tank row here: see
finding tank_rotations-20261110-1. The orc shaman base stats (3b0983bc6) are levels 2-43 only.

## Baseline (ref/tank_bench/summary.tsv, header now ca5e35c3f)
Every request on every boss (13 x 7 = 91 rows), 1,000 iterations, default seed. Scoring means, pass 6 in brackets:

| request | w | score (pass 6) | SafeDPS (pass 6) | DTPS | healing needed | death floor |
|---|---|---|---|---|---|---|
| druid_bear_balanced | 1 | 2006 (2498) | 1650 (1883) | 584 | 462 | 0.0 |
| druid_bear_max_survival | 3 | -298 (238) | 1460 (1649) | 544 | 387 | 0.0 |
| druid_bear_max_threat | 0.333 | 2954 (3492) | 1758 (2024) | 612 | 624 | 0.0 |
| paladin_protection_balanced | 1 | 2753 (2835) | 1688 (1724) | 314 | 245 | 0.0 |
| paladin_protection_max_survival | 3 | 1763 (1910) | 1572 (1611) | 298 | 186 | 0.0 |
| paladin_protection_max_threat | 0.333 | 3172 (3249) | 1694 (1731) | 306 | 256 | 0.0 |
| shaman_warden_balanced | 1 | 1682 (1786) | 1805 (1857) | 675 | 691 | 0.0 |
| shaman_warden_max_survival | 3 | -2031 (-1988) | 1741 (1788) | 686 | 666 | 0.0 |
| shaman_warden_max_threat | 0.333 | 3055 (3170) | 1894 (1949) | 718 | 796 | 0.0 |
| warrior_protection | 1* | -47 (-14) | 1032 (1048) | 636 | 756 | 0.0 |
| warrior_protection_balanced | 1 | 3637 (3880) | 2272 (2386) | 377 | 319 | 0.0 |
| warrior_protection_max_survival | 3 | 1966 (2138) | 1975 (2059) | 336 | 241 | 0.0 |
| warrior_protection_max_threat | 0.333 | 4375 (4652) | 2396 (2535) | 415 | 426 | 0.0 |

Ragnaros moves the same way (bear Balanced SafeDPS 1796 -> 1559, warrior Balanced 2310 -> 2203).

## Why they moved
Under Forever the totem's proc mask is `ProcMaskMeleeMHAuto` (`sim/core/buffs.go`). Maul is `MHSpecial` only
(`sim/druid/bear_attacks.go:43`), so it lost its totem procs: bear SafeDPS -12% at every weight. Heroic Strike is
`MHSpecial | MHAuto` (`sim/warrior/heroic_strike_cleave.go:25`) and keeps them; warriors lose only the yellow-hit
procs (Shield Slam, Revenge, Sunder): -5%. Warden loses Stormstrike's procs (-3%), paladin its special hits (-2%).
The bear/warrior asymmetry between two on-next-swing attacks is the finding; the commit says these weren't checked.

## Rotation lines
No rotation, request or template changed since pass 4's 36 one-line A/Bs. The totem rule cuts every tank's melee
threat about evenly within a class, so no line's order should flip; with Maul now worth less, a bear's Rage
might go further on Swipe or Savage Bite, but that waits for the ruling. Not re-probed.

## Open findings
- 20261110-1 (Windfury Totem and on-next-swing attacks): new, user ruling.
- 3 (paladin Bulwark 13 s) and 11 (warrior Max survival armor trinket 35 s): user calls, unchanged; not re-tested.

## Notes for the next pass
- `tools/audit/tank_rotations/bench_summary.py` now takes the header as its first argument and still reads
  `/tmp/audit-work/tank_rotations/base/`. Export `AUDIT_WORK` in every shell. For an older tankbench, add a
  worktree under `$AUDIT_WORK`, copy `sim/core/proto/*.go` from the main tree (when `proto/` is unchanged), build there.
- Bench the commit you record as `last.master`.

## Conclusion
One new finding (high): Windfury Totem's white-swing-only rule takes 12% off the bear's SafeDPS through Maul while
Heroic Strike keeps its procs.
