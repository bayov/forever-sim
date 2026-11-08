# Fairness pass 32 (2026-11-08, master 9ba00c2)

Result: quiet (no new finding; the Enhancement staleness is already open as results-20261104-2, the imbue pairing as rotations-20261108-1). 27 watched files changed since 73d5bfa: `sim/core` (whole-point damage
rolls 2a60326a1, downranking on heals only 0193721bc, Windfury / Flametongue Totem replace each other 9ba00c2dc),
the shaman Stormstrike mark (5a870eccc), standards 25 and 40, the four Enhancement requests and presets. Full pass:
`audit1b.py`, a full 5k race sweep on the loop's 9ba00c2 CLI and tree, a 20k bisect of the Enhancement move and the
other imbue pairing. New baseline: 9ba00c2.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Stormstrike mark after the damage calc | 5a870eccc | none on the report requests (Enhancement 1166.7 before and after, 20k) |
| Spell damage rolls in whole points (`sim.Roll` rounds); probability rolls moved to `RandomFloat` | 2a60326a1 | variance only; no build's average moves. I checked every non-damage `sim.Roll` use: none is a 0-to-1 probability any more |
| Forever downranking penalty on heals only (user ruling 2026-11-08) | 0193721bc | no DPS build moves (no report rotation casts a damage rank learned 18+ levels below 60) |
| Windfury and Flametongue Totem replace each other; Enhancement loses the raid's Flametongue Totem (standards 25) | 9ba00c2dc | Enhancement -20.0 (20k), -20.7 (5k sweep); the other melee keep Windfury Totem and never had Flametongue Totem, so the raid stays symmetric |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Movement against the a079415 baseline: only the four Enhancement builds moved (-20.7, -19.4, -19.8, -18.0 at 5k).
  Everything else identical to 0.1. Bisect at 20k on the old level60 request: 73d5bfa / 5a870ec / 0193721 all 1166.7,
  9ba00c2 1146.7. Nothing shared leaked.
- Best race: unchanged everywhere; Orc stays best for Enhancement (Dwarf 1138.3 best Alliance). Seal Fate Swords'
  Undead is 0.2 under Night Elf (noise).
- Published against current: Enhancement now -11.6 to -16.7 under its published numbers and 3.0 under Fury (two-hand)
  at 5k (4.1 at 20k), so the published ranking has the order wrong (Enhancement 5th published, 6th now). Independent confirmation of the 2026-11-08 update under results-20261104-2, with the rank flip and the bisect above; not filed again. The other ten rows off by
  more than 1.5 are the same as pass 31 (Feral +30.6, Ret +11.0, SF Swords -11.9, SF Daggers -11.6, Arcane +7.9,
  Combat Daggers +7.9, Elemental +7.3, Subtlety +4.0, Mutilate -2.7): still open as results-20261104-1/-2 and
  results-20261105-1; independent confirmation, not filed again. A 20k run of those seven on the 73d5bfa CLI gives the
  same numbers to 0.01, so none of this range's commits touches them.
- Standards 40, the other pairing: Flametongue Weapon with the group's Windfury Totem, unregeared, 20k: level60 1146.1
  (-0.6), imp-gw 1131.7 (-0.5). A tie on Windfury-weighted gear for these two; rotations-20261108-1 finds Flametongue Weapon ahead on Mana Tide (+5.6) and Water Shield (+3.4). Re-sim the four after that is settled.

## Pass doc notes for the next pass

- The "Settled" bullet on the melee raid ("The second shaman's fire slot is Flametongue Totem") and the open item
  `sham_ftt_stack` are superseded by standards 25 as of 2026-11-08: no raid shaman has Flametongue Totem.
- `table.py` needs `AUDIT_TREE` exported in the same shell (it raises KeyError otherwise).

Top over bottom: 1.68x (Lone Wolf 1314.4 / Smite 781.0).
