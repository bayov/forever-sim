# mechanics: branch pass, phase0-cleanup 782ad6089 (2026-11-27)

Scoped by hand to the branch's one mechanic change (README "Running a pass by hand"); state/mechanics.json not advanced
(it tracks master). Full pass still due on master.

## Checked
- 413f06cb, Talisman of Ascendance (22678, use 28200): every non-auto cast of ours spends a charge
  (`sim/common/item_effects.go:4790`, SpellFlagAPL and not white melee). Matches ProcTypeMask 87376 (melee and ranged
  class spells, ranged autos, helpful and harmful none/magic spells, no melee autos) in client 1.60.1.71139 and
  1.15.7.61186. The use's own spell has SpellFlagNoOnCastComplete, so it doesn't eat a charge. Cooldown 60 s and the
  20 s shared trinket category match ItemEffect 104401. `TestTalismanOfAscendanceCountsEveryCast` passes.
- Left as is: potions, runes and racials carry SpellFlagNoOnCastComplete and so spend no charge. The mask's "helpful
  spell, no damage class" would count item spells in game. Rarely inside the 20 s, unmeasured, not filed.

## New finding
- mechanics-20261127-1 (high, needs user ruling): 28200 has ProcCharges 0 in both clients. The 6-cast end rests on the
  tooltip, and the code comment (:4758) misquotes the tables. Running the full 20 s at 5 stacks is +13 to +22 DPS on the
  19 wearers (paired 20k).

## Context
- The wearers' requests still carry 22678 although the Phase 0 lists took it out (gear-20261127-1, fairness-20261127-1, open).
