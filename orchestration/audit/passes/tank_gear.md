# Pass: tank synthetic gear

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03, as `gear.md` is for the DPS specs.

You check the tanks' synthetic Phase 1 sets (`tools/database/forever_synthetic_items_<class>_tank.go`,
`ui/<spec>/gear_sets/level60*`) against the standards: 1b (the budget, the shape check, the weapon and shield,
trinkets), 2 (item levels), 4 (hit), 5a (the stat mix), 6 (set bonuses), 7 (enchants), 8a (real items) and 36 (tank
stat prices, the Tanks rows of `orchestration/gear_targets.md`).

## The pass

1. **What changed** since `last.master` on the watched paths.
2. **Each piece**: weighted linear total within 1% of slot mod x L(ilvl) x f at standards 36's prices (Defense,
   dodge, parry, block, block value, bonus armor), the 1.7-norm within 6% of its shape target, the armor type the
   class wears, a shield where the class tanks with one, item level and slot as the set's comment says.
   Recompute, don't trust the comments.
3. **Each set**: every stat's share within its tolerance of the spec's Tanks row, hit as standards 4 for a tank's
   threat, the guardrail, the trinkets on 1b's rules and 36's tank trinket, the enchants the best a Phase 1
   character gets outside the raids (`assets/db_inputs/forever_wowhead_recipes.json`).
4. **Each Tanks row**: does it follow from its sources as `gear_targets.md`, Tanks, says (never fetch Wowhead in a
   pass)?
5. Report: a table per set (piece, linear, norm, verdict) and per row; new findings.
