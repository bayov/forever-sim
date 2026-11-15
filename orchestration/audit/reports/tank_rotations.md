# tank_rotations pass 14

Master 70ea588bb (2026-11-15), against pass 13's baseline (state last.master d630df2cb). 1,000 iterations,
default seed, every tank request on all seven boss files.

## What changed

Since d630df2cb (gate: sim/druid/ferocious_bite.go, sim/druid/forms.go), one commit touches a tank input:

- 0cc2c3c8b: an avoided Ferocious Bite refunds 80% of its 35 Energy under Forever (Cat Form only; no tank APL
  casts Bite), and Sharpened Claws now gives spell crit as well as melee crit in the feral forms
  (`GetFormShiftStats`, sim/druid/forms.go:105-110). That reaches the bear requests.

No tank APL, request, rotopt tank template, tools/encounters, tools/tankbench/main.go or tools/tankscore changed.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for this master. Only the three bear requests moved, and all within error:

| request | score mean (old -> new) | SafeDPS mean | per-boss move vs score err |
|---|---|---|---|
| druid_bear_max_threat | 2782.2 -> 2787.2 | 1677.8 -> 1680.7 | +3 to +10, err 14-50 |
| druid_bear_balanced | 1856.2 -> 1862.8 | 1569.3 -> 1572.5 | +3 to +14, err 19-41 |
| druid_bear_max_survival | 261.8 -> 261.2 | 1226.3 -> 1231.3 | -74 to +53, err 46-134 |

The +2 to +5 SafeDPS is consistent with the extra spell crit (spell-school damage the bear does). The other 10
requests are bit-identical. No finding.

## Open item: main-hand ability timing for Windfury Totem (done this pass)

Totem proc ICD is 100 ms (sim/core/buffs.go, WindfuryTotemBuffAura). Helper `tools/audit/tank_rotations/mk_wfgate.py`
gates each priority entry that casts a main-hand weapon ability (Shield Slam, Revenge, Sunder; the bear's 1238070
and 1235827 and Swipe; Holy Strike; Stormstrike) with an extra condition:

- **after**: not within 100 ms after a main-hand swing (`autoTimeToNext(MH) <= autoSwingTime(MH) - 0.1s`);
- **spread**: also not within 100 ms before the next swing.

Paired `abtest` (tankscore.PairedDiff), 1,000 iterations, scoring mean of the six bosses, at the preset's weight:

| request | after | spread |
|---|---|---|
| druid_bear_max_threat | -28.1 ± 6.9 | -139.3 ± 11.9 |
| druid_bear_balanced | -27.4 ± 15.0 | -91.0 ± 20.3 |
| druid_bear_max_survival | -83.6 ± 42.2 | -89.9 ± 45.9 |
| paladin_protection_max_threat | -95.8 ± 6.8 | -98.1 ± 6.9 |
| paladin_protection_balanced | -106.7 ± 12.7 | -107.5 ± 12.8 |
| paladin_protection_max_survival | -121.3 ± 29.8 | -134.1 ± 30.0 |
| shaman_warden_max_threat | -3.4 ± 4.0 | -6.4 ± 5.4 |
| shaman_warden_balanced | -5.8 ± 8.3 | -1.9 ± 10.6 |
| shaman_warden_max_survival | +17.0 ± 17.7 | +2.2 ± 22.5 |
| warrior_protection_max_threat | -6.1 ± 5.0 | -27.2 ± 11.7 |
| warrior_protection_balanced | +15.6 ± 9.2 | -64.1 ± 17.1 |
| warrior_protection_max_survival | -7.2 ± 20.1 | -61.7 ± 27.5 |
| warrior_protection (legacy, -preset balanced) | -18.6 ± 11.8 | -14.6 ± 15.5 |

The two positive rows, at 3,000 iterations on seeds 1 and 2: warrior balanced after +8.9 ± 5.5 and -0.5 ± 6.4;
Warden max survival after -6.7 ± 9.2 and -14.9 ± 9.9. Neither holds. Holding a GCD ability off the swing costs more
than the totem procs it saves (a 20% proc only collides inside a 100 ms window). The Enhancement-style "hold
Stormstrike for Windfury Weapon's cooldown" try doesn't apply: all three Warden requests use Rockbiter Weapon.
Maul and Heroic Strike are the swing itself and weren't gated. No finding. The orchestrator can remove the open item
from passes/tank_rotations.md (this pass may not edit that doc).

## Rotation lines

No rotation changed; pass 13's A/Bs stand.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call, unchanged.
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling, unchanged.
