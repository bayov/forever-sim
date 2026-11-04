# tank_mechanics pass 2 (2026-11-04, master b38ff167d, client 1.60.1.70755+hf)

Forced by hand after tank round 2 landed (b38ff16): the four class workers fixed pass 1's findings and re-optimized every tank's three presets. Five read-only subagents (warrior, paladin, bear, Warden, engine) checked each fix, then what round 2 touched.

**Getting current:** no newer client build than 70755. Blizzard's posts: `dev_notes_seen` unchanged (mechanics pass 65's). Hotfix push 113380 (shaman shocks, Stormstrike) landed on master in 87922bb77 and matches the hotfix rows; it doesn't move the Warden's threat. The local `+hf` tables turned out to be the bare build's (see verdicts.md's setup note); the Warden checks used wago.tools/hotfixes directly.

## Pass-1 findings
| Finding | Verdict | Evidence |
|---|---|---|
| -1 tank behind the boss | resolved | MakeTank, tankbench, all 13 requests in front; TestMakeTankInFront; bear DTPS 386 parry haste on vs 350 off |
| -2 Primal Bite ×2 | resolved | threat ×2 only, form 2.15 separate, test on literals |
| -3 double parry haste | resolved | 7 encounter files parryHaste false; SoD copies left (-14) |
| -4 Revenge 0.25 | resolved (ruled) | constant + to-verify note in code and doc |
| -5 Rockbiter | resolved (ruled) | note in code and doc, value unchanged |
| -6 Unbridled Wrath | resolved | auto attacks only; test fails on old proc |
| -7 bear white Rage | resolved | Feral tank only, 11.25 / 22.5, 0 on avoided |
| -8 health reads | resolved | Onyxia 300 vs 5000 HPS identical for the warrior; latent APL holes in -14 |
| -9 Reckoning | resolved | test fails with the old gate restored |
| -10 docs | resolved | leftovers in -13 |
| -11 tests | open (engine part) | class rows done; parry haste, warrior in TestScoringRunReadsNoHealth, crush/block/tankscore helpers, Reckoning periodic case missing |

## What round 2 changed
| Spec | Item | Client | Sim | Verdict | Tier |
|---|---|---|---|---|---|
| Warrior | Bloodrage / Reckless Swings health costs | 20% / 25% base health | × class base (1689 Orc) | convention, to verify | 1 / convention |
| Warrior | PayHealthCost | — | once into the replay's healing needed; not DTPS, not TMI; health floored at 0 | ok | — |
| Warrior | Shield Greater Stamina 1886, Deflection 923, Superior Deflection 8214 | +9 Stam, +7 / +9 Defense | same | ok | 1 |
| Paladin | Reckoning crits | mask 0x202A8, 8%/rank, 4 stacks | any non-periodic crit | ok; 1.5 s recovery on crits open | 1 |
| Paladin | Divine Favor / Oath of Light no Benediction | empty class masks | same; no tank takes them | ok | 1 |
| Paladin | Templar's Bulwark 1311015, trinket line off GCD | 110 mana, 300 s, no GCD | same | ok | 1 |
| Paladin | no bubble | — | no 642/1020/498/5573/10278 anywhere | holds | ruling |
| Bear | Enrage in fight, Frenzied Regeneration ≥100 Rage, Roar 204 AP / 48.4 threat, Maul 10 Rage | rows | same | ok | 1 |
| Bear | Blood Frenzy +5 on white crits | 16958 "any critical strike" | 27.5 on a white crit | ok | 1 |
| Warden | Earth Shock r7 517-545, shared 6 s shock cooldown, Reverberation | push 113380, category 19 | same | ok | 1 |
| Warden | Stormstrike 150%, 6 s | push 113380 | same | ok (labels stale, -13) | 1 |
| Warden | Ancestral Fortitude aura 101 on all armor | 16177 | ArmorMultiplier once | ok; no Warden takes it | 1 |
| Engine | Elixir of the Phalanx 1250920, Greater Fortitude 1250931 | +400 health, +500 armor; +400 health | applied once (+800 / +500 measured) | ok | 1 |
| Engine | raid damage share-under-one (b2c16f671) | — | NotTanks only; tank events still roll avoidance | ok | — |
| Engine | health follows max health down (0aed93989) | — | live run only; replay doesn't (-14) | latent | — |
| Engine | mana gain threat | Classic: 0.5/mana unmultiplied | 0.5 × threat multiplier | **no source (-12)** | 3 vs none |

Presets: the 12 requests use the top ranks, face the boss, and match their UI APLs. Warrior Max threat and Balanced list 990617 without wearing it (a no-op). The Warden is still over the melee hit cap (10.5% vs 8%), already tied in docs/tank_shaman.md. Paladin Balanced and Max survival are byte-identical (tank_fairness-4 territory, not mechanics). The bench summary's rows reproduce at 1,000 iterations (Chromaggus Balanced, all four classes); its header names 23b2ec2 though rows were re-benched later (tank_rotations' file).

## New findings
- -12 (high): mana gains make threat × the threat multiplier, extrapolated from the Rage rule with no source. Bear Balanced -2.9% with JoW off; roughly half of that rides on the multiplier.
- -13 (low): stale source labels and comments (Warden push 113380 rows, bear 1.5 comment, warrior shield enchant name and forever_enchants comment, paladin round file).
- -14 (low): latent scoring-run holes (replay ignores max-health drops; APL health values unguarded) and the SoD encounter copies still double parry haste.
