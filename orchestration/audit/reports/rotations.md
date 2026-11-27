# Rotations pass 86 (2026-11-27, master 5ab589d858)

Result: findings (1 new). Range 34994816e..5ab589d858: Talisman of Ascendance charges on every APL cast (413f06cb0),
priest tank healers pace Flash Heal (ff87e66b3), healer re-sims, Phase 0 rogue effects refresh, rotopt balance no trinket hold.

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py 34994816e 5ab589d858` | 2 files, both healer (discipline / holy tank_healer Flash Heal pacing). No Level 60 DPS APL logic changed |
| Non-test sim `.go` changed | `sim/common/item_effects.go`: Talisman of Ascendance (22678) now eats a charge on every APL spell except white melee. Reaches 19 DPS reqs (warlock, Balance, Elemental) |
| Shared run vs published | warlocks -1.2 to -5.6 (duty-curse builds lose most); Balance and Elemental within 0.4 |
| Talisman placement (new mechanic) | see below; one finding |
| Open findings landed? | 20261126-3 (Might of Shahram), 20261122-1 (Shock Ret), 20261112-1 (rogue poisons): still open, untouched |
| Baseline | not refreshed (`baseline_ab3fff7_3000it_seed424242.tsv`); Talisman wearers moved 1-6 DPS. Refresh on the next pass that runs `baseline.py` |

## Talisman of Ascendance placement (paired ab.py, 20k it, both seed sets, pooled B-A at 120 / 180 / 300 s)

| preset | variant | pooled |
|---|---|---|
| affliction-duty-curse | Talisman before Shadow Bolt | +0.71 / +1.07 / -1.07 (trade-off) |
| affliction-duty-curse | Talisman right after duty curse | **+4.65 / +4.63 / +3.40** |
| affliction-duty-recklessness | after duty curse | **+4.31 / +4.52 / +3.49** |
| demonology-duty-curse / -reck | after duty curse | **+2.21 / +2.60 at 180** |
| destruction-duty-recklessness | after duty curse | **+1.42 at 180**, +0.12 at 300 |
| destruction-duty-curse | after duty curse | +0.56 at 180 (optional) |
| destruction_fire duty curse / reck | after duty curse | +0.38 / +0.41 at 180 (optional) |
| affliction | Talisman before Shadow Bolt | -5.9 / -3.4 / -3.1 (autocast holds) |
| balance | Talisman before Starfire | -3.9 / -3.1 / -1.9 |
| balance-duty-ff | before Starfire / after Faerie Fire | -1.7 / -1.25 at 180 |

Filed as rotations-20261127-1. Caveat: the warlock Phase 0 lists on master no longer wear 22678 (pass 85) but the reqs
still do; the finding applies only while they do. Elemental (moved <=0.1) not tested.

## Optional

- Destruction CoE and Destruction Fire duty builds: Talisman after the duty curse, +0.4 to +0.6 at 180 s.
