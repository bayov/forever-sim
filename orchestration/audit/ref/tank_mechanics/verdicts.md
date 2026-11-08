# Tank mechanics verdicts

What the tank_mechanics pass verified, settled and left open. docs/tank_sim.md (its Rules audit) and docs/tank_<class>.md carry the class workers' own audits.

Setup note (pass 2): the local `wago/<build>+hf` tables can be the bare build's (wago.py logs "no cached search covers SpellEffect ... so SpellEffect is the build's own table"). Check hotfixed rows on wago.tools/hotfixes (region 91) or run `hotfixes.py fetch` first; push 113380 (shocks, Stormstrike) is not in them.

## Verified (pass 1, 2026-11-04, client 1.60.1.70755+hf; pass 2 re-checked what round 2 touched)
- Every tank ability and talent the presets use matches the client rows (tier 1). Tables: reports/tank_mechanics.md.
- Pass 2: Revenge 0.25 AP (ruled, to verify); Unbridled Wrath auto attacks only; Reckoning on any non-periodic crit; Primal Bite 2× threat (notes); bear white-hit Rage 11.25/22.5, auto swings only; tank enchants 1886 (+9 Stam), Deflection +7, Superior Deflection +9 Defense (client rows); Elixir of the Phalanx and Greater Fortitude (1250920, 1250931) applied once; Ancestral Fortitude aura 101 on all armor (no Warden takes it); hotfix 113380 shocks and Stormstrike match; Templar's Bulwark 1311015; Enrage, Frenzied Regeneration, Demoralizing Roar rows.
- Engine (tier 3 Classic rules): boss white table and melee ability, crush, Forever 0.6% boss crit and `classic_crit`, armor and its cap, block, raid damage events rolling avoidance, Rage gain threat (tier 2 measurements). Tanks face the boss; parry haste fires; encounters don't double it; a scoring run reads no health (warrior measured HPS-invariant). Self health costs (PayHealthCost) go once into healing needed, not DTPS or TMI.
- Pass 3 (client 1.60.1.70790+hf): mana-gain threat unmultiplied (0.5 a mana, Classic tier 3); Enrage 5229 rows; Stormstrike 150%/6 s; items in Bear/Dire Bear Form follow SpellShapeshiftForm flags 0xd8 (CanUseItems) and each spell's Attributes_0 0x10000 (tier 1): potions, runes, healthstones, EZ-Thro, Goblin Sapper Charge and Goblin Land Mine usable; Dense Dynamite, Thorium/Iron Grenade, Solid Dynamite and Holy Hand Grenade not (sim wrong on sapper and land mine, -15).
- The Warden gets no Salvation compensation in the Nov 2 notes, and the sim gives none (tier 1 notes).

- Pass 3 (master 0facd9428, client 1.60.1.70790): 70790 is a game-data build (SpellEffect +53/-10/~171 rows vs 70755). For tanks it only bakes in hotfix values the sim already uses (tier 1): Earth Shock r1-r7 (r7 531, 3.4/level), Flame Shock, Stormstrike 150% (1368308) and 6 s, Vengeance 5 stacks, Two-Handed Weapon Specialization Physical+Holy, Reckless Swings 1329554/1329555 (5 Rage x4 charges, 25% base health, 15 s), warrior ranks (Improved Tactical Mastery 3, Flurry 3, Improved Whirlwind 2, Bloodthrill 2; tools/forever_talents/data/warrior.json already has them). No bear, Protection paladin or Protection warrior ability row changed. Booming Voice's mask adds Howls (no tank effect). Mana-gain threat now unmultiplied (49f0218), replay caps on max health drops (0950b22): verified, tests pass.
- Pass 4 (master a07941531, client 1.60.1.70790): the round 4/5 consumables match their client rows (tier 1): Goblin Sapper Charge 600/500 with Variance 0.5 (450-750 out, 375-625 self), Greater Stoneshield +2000 BonusArmor (not scaled by bear form), Flask of Natural Precision +60 Stam (zone expertise only), Rumsey Rum Black Label +15 Stam, Spirit of Zanza +50/+50, Ground Scorpok +25 Agi. Improved Seals 20224 masks reach the Seal of Fury hit (20418) and Judgement of Fury (20414), 3%/rank. Fire Nova r5 hit 408428 and Searing Totem r6 rows match.
- Pass 6 (master 12aa2796f, client 1.60.1.70862): the client now carries the November 2 threat values the sim already used (tier 1, was notes-only): Defensive Stance Passive 7376 #2 A_MOD_THREAT 30 -> 80 (x1.8), Bear Form (Passive2) 21178 #0 50 -> 115 (x2.15). Defiance 12792 is still 15 at rank 3 in the client, so 7/13/20% stays notes-only. No other player tank spell changed (Mangle! 57596 is a creature spell; the rest of the diff is new creature and flight-path spells).
- Pass 7 (master 9ba00c2dc, client 1.60.1.70862): Stormstrike's mark on spells now scales the result after the damage calc and redoes threat with ThreatFromDamage (same threat as before; Nov 5-8 live logs, tier 1 logs). Damage rolls in whole points (Nov 5-8 logs); no tank proc or probability roll uses sim.Roll (cast.go pushback moved to RandomFloat). Downranking penalty on heals only (user ruling 2026-11-08). Windfury/Flametongue Totem replace each other (beta, user): the Warden presets use Rockbiter + raid Windfury Totem and place no Flametongue Totem, so unaffected.

## Open
- Sapper self-hit rolls the caster's spell crit, so a self-crit can proc Reckoning, Redoubt or Enrage (in-game behaviour unknown; once a fight).
- Mana gain threat: unmultiplied since 49f0218 (Classic tier 3, to verify with a log).
- Explosives in form: sapper and land mine (-15).
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
