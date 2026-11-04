# Tank mechanics verdicts

What the tank_mechanics pass verified, settled and left open. docs/tank_sim.md (its Rules audit) and docs/tank_<class>.md carry the class workers' own audits.

## Verified (pass 1, 2026-11-04, client 1.60.1.70755+hf)
- Every tank ability and talent the presets use matches the client rows, except the findings below (tier 1). The tables are in reports/tank_mechanics.md, pass 1.
- Engine (all Classic rules, tier 3): the boss's white table and its melee ability, crush by level, Forever's 0.6% boss crit and `classic_crit`, armor and its 75% cap, block, raid damage events rolling avoidance, Rage and mana gain threat. The sheet's boss side matches the rolled outcomes.
- The Warden gets no Salvation compensation in the Nov 2 notes, and the sim gives none (tier 1 notes).

## Open
- Primal Bite threat ×2 (finding -2; tier 1 notes, value "roughly").
- Revenge's AP ratio, 0.28 vs 0.25 (-4): needs a level 60 log after 70707.
- Rockbiter's threat on Stormstrike, ×speed vs flat (-5): needs a log.
- Shield Slam's +508 threat and Thunder Clap's 2.5×: open calls in docs/tank_warrior.md.
- Earth Shock ×2 threat: user call (a), tier 2 (wowsims/sod).
- Maul and Lacerate threat beyond the form; Enrage's 16% armor cut; Faerie Fire (Feral) threat in bear form (value is set server-side).
- Holy Shield's ×1.2 threat (tooltip + tier 3); how Righteous Fury's two multipliers stack.
- Redoubt's proc chance below 5/5 (the presets take 5/5).
- No avoidance while casting (spell_outcome.go:673): tier 3 only.
