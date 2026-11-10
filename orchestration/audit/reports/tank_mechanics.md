# tank_mechanics pass 9 (2026-11-10, master da272241e, client 1.60.1.70888+hf)

Due on the client build 1.60.1.70862 -> 1.60.1.70888; no watched file changed (ca5e35c3f..da272241e). Result: quiet.

**Getting current:** wago's latest is 1.60.1.70888. Blizzard's Nov 9 and Nov 10 hotfixes (read from TASKS.md's daily update, da272241e): hunter Arcane Shot / Summon Hawk / Great Stamina, mage Arcane Concentration, Lightning Overload, Nature's Swiftness, Natural Instinct on Pounce, Deathless Marrow NPC nerfs. None touches a tank ability, talent or the tank engine. No ledger entries in range.

**Client diff** (`wago.py diff 1.60.1.70862+hf 1.60.1.70888+hf --by-spell`, 22 spells):

| Row | Tank effect | Verdict |
|---|---|---|
| 1223242 Natural Instinct #0 mask word 3 0 -> 136 (Pounce Bleed 9007 0x80, Lacerate 414647 0x8) | 414647 is a Lacerate with only a weapon-damage-percent effect (31), triggered by no client spell: likely the server's Lacerate hit. The sim gives NI's +10%/point crit damage to the whole Lacerate spell already (sim/druid/talents.go `applyNaturalInstinct`, TestBearTalentDamage); ticks use OutcomeTick. Client now agrees with the sim. | right (tier 1) |
| 9005/9823/9827 Pounce, 414924 Feral Charge (Cat) attributes | cat only; no bear preset uses them | n/a |
| Frost Nova radius, Frost Trap, Cultivation, Escape Rope, Teach Summon Felsteed, Flamebreak Talisman, NPC spells (Bloodbolt, Charge 1283073, Fireball, Poison Shock) | none for tanks (Charge 1283073 is an NPC spell) | n/a |

**What changed in code:** 2e320521f (Chief's Enforcer proc, sim/common), c2baee417 (logcheck), da272241e (TASKS.md). Nothing in sim/core tank files, sim/warrior, sim/paladin, sim/druid, sim/shaman or tools/tankbench/tankscore.

**Engine and tests:** untouched this range; pass 7's checks stand.

**Open findings:** 20261107-1 (threat labels: docs/tank_druid.md:22 and :418 still say 21178 reads 50) and 20261108-1 (no Stormstrike-mark shock threat test in sim/shaman/warden/forever_test.go) remain open.
