# Rotations pass 63 (2026-11-13, master 97da1cca5)

Last audited master 71562a8b0. What reaches the DPS raid presets:

| commit | change | reaches DPS presets |
|---|---|---|
| a1a1a81bb | Windfury Totem: no extra attack (Windfury, HoJ, Sword Spec, Thrash, Reckoning) rolls it; stored extra attacks keep their source | yes (Combat Swords -3; warriors under 1) |
| 33b13e310 / unsealed_wrath.go | Unsealed Wrath stores no echo from Windfury Totem swings (Nov 11 hotfix) | yes (Retribution -12 to -19 at fixed APL) |
| ec8ab9b3d | Ret APLs re-tuned after the hotfixes: Seal of Righteousness held, twist back at 3 stored copies; Ret Shock on SoR, Holy Shock ahead of Holy Strike; Undead talents | yes (worker-measured, two seeds) |
| 2fcb6fe2a | Volley activates no procs | hunters: no move over 1 |
| 02cf6c40e | item effect 18656 retimed to 28 every 3 s under Forever | no preset move over 1 |

`apl_logic_diff.py 71562a8b0 origin/master`: 7 APL files, all Retribution, all from ec8ab9b3d.

## Sim regression (`regress.py`, old CLI 71562a8 vs new CLI 97da1cc on the new req files, 3000 it, seed 424242)

| build | old | new | diff |
|---|---|---|---|
| paladin__level_60 / -duty-crusader / -duty-wisdom | 1000.5 / 991.5 / 985.6 | 981.4 / 978.2 / 971.0 | -19.2 / -13.4 / -14.7 |
| paladin__level_60_sub_holy / -crusader / -wisdom | 992.7 / 981.9 / 975.3 | 973.4 / 969.7 / 962.4 | -19.3 / -12.2 / -12.9 |
| paladin__level_60_sub_holy_undead | 1127.5 | 1113.4 | -14.2 |
| rogue__combat-swords | 946.6 | 943.6 | -3.0 |

Everything else within 1 DPS (warriors included: the extra-attack gate barely touches them).

## Baseline move (3000 it, seed 424242)

New baseline `ref/rotations/baseline_97da1cc_3000it_seed424242.tsv`. Against 71562a8: Ret twist and sub-Holy
builds -18 to -22, vs Undead -12.5, Ret Shock +2.4 / Wisdom duty +3.5, Combat Swords -3.0; the rest under 1.
This matches the sim regression plus the re-tune's gain.

## Re-tune spot check

`ab.py` on paladin__level_60, new CLI, A = 71562a8's req rotation, B = master's, S0 = 2e8, 20k it:
+2.09 ±0.50 / +1.89 ±0.22 / +3.98 ±0.24 at 120 / 180 / 300 s. The re-tune wins in every cell, as its
commit says.

## Conclusions

- No new findings. The sim moves are the Nov 11 hotfixes, and the worker re-tuned Retribution on top of
  them (confirmed above).
- Open findings, all still unapplied on master: 20261112-1 (rogue poisons), 20261109-1 (Fury 2H 90-rage
  Heroic Strike line still at `level60_fury_2h.apl.json`), 20261105-1 (feral FB gate still 45). Their
  presets moved under 1 DPS this pass (Combat Swords -3 is not one of 20261112-1's dagger builds), so
  their measurements stand.

## Open items

- Retribution sub-Holy and duty builds: only the twist build's re-tune was spot-checked.
- Carried from the pass doc: Balance Night Elf vs Tauren, Elemental Mastery hold, Enhancement Flametongue
  Weapon with Windfury Totem, warlock duty-curse screens.
