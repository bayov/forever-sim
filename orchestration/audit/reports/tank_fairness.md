# tank_fairness pass 18: 2026-11-17, master 5f20dfdf4

Due: 4 watched files changed since 1a85df1d7: `sim/warrior/{execute,reckless_swings,talents}.go` (hotfix push 114097:
Unbridled Wrath now procs on landed melee abilities but Whirlwind; a Cleave spends one Reckless Swings charge;
ad1d7f469) and `sim/shaman/item_sets_pve.go` (Forever's Five Thunders set 1679; no Warden preset wears it). Quiet: no
new finding.

## Method

No change to `tools/tankbench`, `tools/encounters` or the tank requests. Built tankbench at 5f20dfdf4 (`--tags=with_db`,
after `make sim/core/proto/api.pb.go`), ran `tools/audit/tank_fairness/bench.sh` on all 13 requests x six scoring
bosses, 3,000 iterations, seed 1, and compared with `ref/tank_fairness/summary.tsv`. Bear, paladin and Warden rows are
identical to the baseline. Only the warrior moved; Max survival warrior and paladin re-run on seed 2. Baseline refreshed
(the 21 warrior rows).

## Same terms

Requests byte-identical to pass 17. The open unequal terms stand: paladin Major Mana Potion with no potion action
(-20261105-1), Warden Night Elf / Night Elf / Undead (-20261105-2), legacy `warrior_protection.json` (-20261104-5).
Push 114097 is a Forever hotfix to the warrior's own talents, so its gain is fair.

## Ranking (six-boss mean, own weight, seed 1)

| weight | order | pass 18 | pass 17 |
|---|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4235 / 3275 / 3123 / 2644 | 4223 / 3275 / 3123 / 2644 |
| Balanced | Warrior, Paladin, Bear, Warden | 3475 / 2843 / 2131 / 1280 | 3452 / 2843 / 2131 / 1280 |
| Max survival | **Warrior = Paladin**, Bear, Warden | 1837 / 1798 / 364 / -2242 | 1794 / 1798 / 364 / -2242 |

Warrior deltas (seed 1, same seed both builds): +12.0 / +23.0 / +42.9 score, TPS +6.4 / +12.0 / +12.8 (the extra Rage
goes to Heroic Strike), DTPS and healing needed within 0.5 / -4 / -2.

Max survival, seed 2: paladin 1832.5 ± 15.9, warrior 1822.3 ± 12.9. Two-seed means 1815 (paladin) and 1830 (warrior):
a tie within error, as in pass 17. The paladin's open potion action (+374 at Max survival, pass 13) still decides it.

Spread unchanged in kind: Warden 38% behind the warrior at Max threat (class gap, standards 40), bear 5% behind the
paladin at Max threat and 25% at Balanced. The logs disagree as before (no paladin main tank on any boss; warriors on
all six, druids on four).

## What a preset rewards

Not re-checked by variant (no request changed). The warrior's Max threat build differs from Balanced / Max survival
only in Protection (`250330020330001351` against `250530020310001351`); each still leads its own weight on this run.

## Open item

The warrior's talents were tuned before Unbridled Wrath procced on Heroic Strike; a re-tune may find a few more points
(tank_rotations' or the warrior worker's job, not a fairness finding).
