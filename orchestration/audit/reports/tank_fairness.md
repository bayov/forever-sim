# tank_fairness pass 7: 2026-11-10, master ca5e35c

Due: 2 watched files changed since 91ad195: `sim/shaman/talents.go` and `sim/shaman/flametongue_weapon.go`, both in
bcc5d6c (Maelstrom Weapon stacks on every landed melee hit under Forever instead of 4 PPM a point; Flametongue Weapon
rank 5 takes 7 off its points).

## Does it reach a tank?

| change | Warden presets | reaches a tank |
|---|---|---|
| Maelstrom Weapon every hit | Enhancement talent 17 (Maelstrom Weapon) is 0 in all three talent strings (`050032203451012201`, `050032303551012201`, `050230113541012201`); `applyMaelstromWeapon` returns at 0 points (sim/shaman/talents.go:546) | no |
| Flametongue Weapon rank 5 cut | all three imbue Rockbiter (main hand) and run Windfury Totem; no Flametongue Weapon | no |

Nothing else in the range touched `tools/tankbench`, `sim/core`, the warrior, paladin or druid code, `ui/warden_shaman`
or the tank docs. So no re-sim: the pass 6 numbers (master 9ba00c2, `ref/tank_fairness/summary.tsv`) stand, and the
ranking in `orchestration/fairness_tanks.md` is unchanged.

## Same terms

Unchanged from pass 6. The open asymmetries stand, none fixed on master:

- tank_fairness-20261105-2 (high): Warden presets still Night Elf / Undead / Night Elf, races a Forever shaman can't be.
- tank_fairness-20261105-1 (high): paladin presets still `defaultPotion: MajorManaPotion` with no potion action.
- tank_fairness-20261104-2 (medium): paladin's low damage taken, waits on the user's calls (a), (b), (c).
- tank_fairness-20261104-5 (low): legacy `warrior_protection.json` still in `tools/tankbench/req/`.

## Conclusion

Quiet. No new findings.
