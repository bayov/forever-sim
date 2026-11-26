# healer_mechanics pass 32

Branch phase0-tankheal 3e9748641 (last: master 85691655a, client 71091). Quiet.

## Get current

- Client: newest build 1.60.1.71139. `wago.py diff --set mechanics 1.60.1.71091 1.60.1.71139`: SpellEffect ~23 (Sayge's
  Dark Fortunes, Fengus' Ferocity 230 -> 115, Enveloping Winds and Fist of Shahram targets, Reinforce, Reagent Economy,
  Weight of Duty, Explosive Runeshard), SpellMisc ~67 (rogue poisons and Vanish Attributes_5, world buffs Attributes_6,
  Holy Fire DurationIndex 21 -> 0), ItemEffect (Nutritious Food), ItemSparse (bonding and sell prices), CurvePoint 5.
  No heal, healing talent, Mana Spring / Mana Tide / Innervate or healer item is touched. None of the changed spells is
  in sim/ for a healer. Holy Fire is a damage spell (mechanics' scope). (`--by-spell` crashes on a table without
  SpellID: wago.py diff_by_spell KeyError 'SpellID'; the plain diff works.)
- update_ledger.md: no 71139 entry yet; nothing healing-related since the last pass.

## What changed on the watched paths

- eac7ba8b5: healbench now builds its raid through tools/healscore (shared with rotopt's healer objective). The
  druid_restoration tank rows of the new baseline are identical to the last one row by row, so the refactor kept the
  bench's results. docs/healing_sim.md gains "The rotopt objective" (search-side, no game rule).
- tools/healbench/req/*: equipment only (Phase 0 sets); healer_gear's scope.
- sim/druid/bear_attacks.go: bear threat, not healing.

## Spells, engine, tests

Nothing changed, so the per-spec tables of pass 31 stand. Open findings unchanged (Innervate on Mana Well,
Riptide cost, untested ranks, Vampiric Embrace 10%, Wild Growth coefficient, seen vs true health, priest golden,
NS and Wild Growth).

No new findings.
