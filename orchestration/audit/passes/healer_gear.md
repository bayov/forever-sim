# Pass: healer synthetic gear

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03, as `gear.md` is for the DPS specs.

You check the healers' synthetic Phase 1 sets (`tools/database/forever_synthetic_items_shaman_restoration.go` and
the other healers' `forever_synthetic_items_<class>_<spec>.go`, `ui/<spec>/gear_sets/level60*`) against the
standards: 1b (the pure caster budget, the shape check, the weapon rate, trinkets), 5 (Spirit), 5a (the stat mix)
and 35 (healers' sets, the healer rows of `orchestration/gear_targets.md`, Healers).

## The pass

1. **What changed** since `last.master` on the watched paths.
2. **Each piece**: weighted linear total within 1% of slot mod x L(ilvl) x f (pure caster f), the 1.7-norm within
   6% of its shape target, the armor type the class wears, item level and slot as the set's comment says. Weights:
   Sta / Int / Spi / crit rating 1, spell power 0.86, healing on top of spell power 0.45, MP5 2.4, hit 1.
   `orchestration/tools/healers/` has the scripts the sets were made with; recompute, don't trust the comments.
3. **Each set**: every stat's share within its tolerance of the spec's healer row (Stamina +-1.5, the rest
   +-3), Spirit about 10%, no hit, guardrail (no stat over 55% of the non-Stamina share), the trinkets on 1b's
   trinket rules, the relic's effect modelled (`sim/<class>/items.go`), the enchants the best a Phase 1 character
   gets outside the raids (`assets/db_inputs/forever_wowhead_recipes.json` for where a recipe comes from).
4. **Each healer row**: does it follow from its sources as `gear_targets.md`, Healers, says (recompute from the
   cached Wowhead pages with `orchestration/tools/healers/`; never fetch Wowhead in a pass)?
5. Report: a table per set (piece, linear, norm, verdict) and per row; new findings.
