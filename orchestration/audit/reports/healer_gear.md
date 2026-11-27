# healer_gear: branch pass, phase0-cleanup 782ad6089 (2026-11-27)

Scoped by hand to the priest lists the branch re-scored (bis.json only; gear sets unchanged). State not advanced.

| spec | list score | faction | bis picks vs gear set | requests |
|---|---|---|---|---|
| priest_holy | 133.59 -> 132.29 | Human Alliance | identical (17 slots) | priest_holy, priest_holy_raid wear it |
| priest_discipline | 122.23 -> 123.67 | Human Alliance | identical (2H staff 812, no offhand) | priest_discipline wears it |

- `phase0_check.py`: no flags on either priest healer row.
- The score moves come from the paced Flash Heal APLs. No pick in the sets changed.
- No new findings. healer_gear-20271127-1 (paladin and shaman requests) looks fixed on master by d74a5326; the master pass
  will move it.
