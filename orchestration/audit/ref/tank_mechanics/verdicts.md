# Tank mechanics verdicts

What the tank_mechanics pass verified, settled and left open. docs/tank_sim.md (its Rules audit) and docs/tank_<class>.md carry the class workers' own audits.

Setup note (pass 2): the local `wago/<build>+hf` tables can be the bare build's (wago.py logs "no cached search covers SpellEffect ... so SpellEffect is the build's own table"). Check hotfixed rows on wago.tools/hotfixes (region 91) or run `hotfixes.py fetch` first; push 113380 (shocks, Stormstrike) is not in them.

## Verified (pass 1, 2026-11-04, client 1.60.1.70755+hf; pass 2 re-checked what round 2 touched)
- Every tank ability and talent the presets use matches the client rows (tier 1). Tables: reports/tank_mechanics.md.
- Pass 2: Revenge 0.25 AP (ruled, to verify); Unbridled Wrath auto attacks only; Reckoning on any non-periodic crit; Primal Bite 2× threat (notes); bear white-hit Rage 11.25/22.5, auto swings only; tank enchants 1886 (+9 Stam), Deflection +7, Superior Deflection +9 Defense (client rows); Elixir of the Phalanx and Greater Fortitude (1250920, 1250931) applied once; Ancestral Fortitude aura 101 on all armor (no Warden takes it); hotfix 113380 shocks and Stormstrike match; Templar's Bulwark 1311015; Enrage, Frenzied Regeneration, Demoralizing Roar rows.
- Engine (tier 3 Classic rules): boss white table and melee ability, crush, Forever 0.6% boss crit and `classic_crit`, armor and its cap, block, raid damage events rolling avoidance, Rage gain threat (tier 2 measurements). Tanks face the boss; parry haste fires; encounters don't double it; a scoring run reads no health (warrior measured HPS-invariant). Self health costs (PayHealthCost) go once into healing needed, not DTPS or TMI.
- The Warden gets no Salvation compensation in the Nov 2 notes, and the sim gives none (tier 1 notes).

## Open
- Mana gain threat × threat multiplier (-12): no source; Classic is unmultiplied.
- Revenge 0.25 (ruled, to verify): a level 60 Revenge log after 70707.
- Rockbiter on Stormstrike, ×speed (ruled, to verify): a log.
- Bloodrage / Reckless Swings base health: class base 1689 (Orc 60) vs wowsims' 2629; Bloodrage only in the presets, ~+0.9% healing needed at stake. A level 60 Bloodrage log.
- Reckoning's 1.5 s ProcCategoryRecovery: the sim (tooltip, wowsims/forever) applies it to blocks only; the client puts it on the whole aura. Only matters with two crits in 1.5 s at 0.6% boss crit.
- Shield Slam's +508 threat and Thunder Clap's 2.5×: open calls in docs/tank_warrior.md.
- Earth Shock ×2 threat: user call (a), tier 2 (wowsims/sod).
- Maul and Lacerate threat beyond the form; Enrage's 16% armor cut; Faerie Fire (Feral) threat in bear form (value is set server-side).
- Holy Shield's ×1.2 threat (tooltip + tier 3); how Righteous Fury's two multipliers stack.
- Redoubt's proc chance below 5/5 (the presets take 5/5).
- No avoidance while casting (spell_outcome.go:673): tier 3 only.
