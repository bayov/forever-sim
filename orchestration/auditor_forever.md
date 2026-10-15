# Forever mechanics auditor

Read auditors_common.md first: the read-only rule, the worker table, how to run the sim, and how passes work all apply to you. Two updates to that file: there are now six auditors, and every class has a PR (Rogue #9, Druid #8).

## Your job

Make sure every mechanic the sim uses for these presets is **WoW Forever's**, not one inherited from Classic Era, Season of Discovery, Anniversary or the wowsims defaults, unless Forever actually kept it. The class workers audited their own spells and talents. You cover what falls between them: the shared combat engine, the shared raid, and each class's assumptions that nobody checked against Forever.

### Source order (the user's ruling)

1. **The Forever beta.** That means the client data for build 1.60.1.70338 on wago.tools (db2 tables, including TraitDefinitionEffectPoints / CurvePoint for talents), beta patch notes, blue posts, Blizzard's Forever articles and deep dives, beta tester logs and reports, and the Forever pages on Wowhead and foreverchanges.
2. **Season of Discovery**, as the backup when the Forever beta is silent: SoD client data, patch notes and blue posts.
3. **Classic Era, Anniversary and 2019 Classic**, only if SoD is also silent.

Label every verdict with its tier. Never cite vmangos. Never decide anything from a tooltip alone. If wago.tools refuses WebFetch, don't reach it any other way; note what you couldn't check.

### Checklist (shared engine first, then per class)

Shared combat engine (sim/core and sim/common):
- Rage from damage dealt and taken (the conversion constant at level 60/63, the hit-factor table, offhand factor). This was already fixed once: `^` was XOR, so 230.6 is the right value at 60. Re-check the rest of the formula.
- Melee hit table against a level 63 boss: miss (dual wield too), dodge, parry (bosses attacking from the front; do presets attack from behind?), glancing blows (rate and damage reduction by weapon skill), the crit cap and crit suppression, block. Weapon skill effects, including racials.
- Spell hit against level 63 (base 83%, the 99% cap), partial resists and binary spells, spell crit multiplier, school-specific modifiers.
- Armor mitigation formula; Curse of Recklessness, Sunder Armor, Expose Armor and Faerie Fire values in Forever.
- Haste: how melee, ranged and spell haste stack (Forever's Skyborne +1% haste racial and Troll Berserking at 10%). GCD, and whether haste lowers it in Forever.
- Mana: the five-second rule, Spirit regen formula, the Mp5 and Int-to-crit and Agi-to-crit/AP conversions per class at 60. Energy regen and focus.
- Proc rules: PPM against fixed chance, Windfury (Forever totem and weapon values, ICD), Crusader, the Forever Enchanting 300 enchants on the shared branch, and Forever consumable values (consumes.go on the shared branch).
- Raid buffs and debuffs in the shared raid (tools/rotopt/mksettings/main.go): check each against Forever. Values, stacking, debuff slot limits if Forever has any, totem changes, Blessing changes, Leader of the Pack, Trueshot Aura, Moonkin Aura, Curse of the Elements covering every school at 10%, and the mage/priest personal debuffs (Improved Scorch, Winter's Chill, Shadow Weaving).
- AoE scaling (the user, 2026-10-15, for the leveling sim in docs/leveling_sim.md, where we pull up to 5 mobs). For every ability that hits more than one target (AoE spells and totems, cleaves, chains, splash procs, explosives), find which of three kinds it is in Forever, with how many targets and the formula:
  - **Hard capped:** hits at most N targets (Cleave, Swipe, Multi-Shot, Sweeping Strikes, Chain Lightning's jumps). Client data: SpellTargetRestrictions.MaxTargets, SpellEffect.ChainTargets.
  - **Soft capped:** hits every target, but damage drops past N targets (say what it drops by: split evenly past N, a square root rule like modern WoW's, or something else).
  - **Unlimited:** every target at full damage.
  Also note chain falloff (each jump's damage cut) where there is one. The sim today has the hard caps written into each spell, a 10-target split (`Encounter.AOECapMultiplier`) that only Explosive Trap and one explosive use, and full damage on every target for everything else. Write the table (ability, kind, N, formula, source tier) to orchestration/aoe_scaling.md, Enhancement shaman's abilities first, then the other leveling specs.
- Racials: compare forever-wiki/racials.md with the client spells. That includes the Forever reworks (Human Sword Specialization +2% crit, Orc axe crit, Undead Touch of the Grave, Gnome Eureka!, Troll Berserking 10%, Skyborne Wind Blessed / Elemental Insight) and that each race's class list matches Blizzard's Forever race-class table.

Per class (a lighter pass; the workers own these): look for mechanics the class code inherited from Era or SoD that Forever changed or removed. Examples: SoD runes or set bonuses that leaked into Forever presets, Era coefficients where the Forever client has new ones, Forever talents modelled with Era behaviour, abilities Forever removed or added. Spot-check coefficients and base values in SpellEffect against the sim for the spells that make up most of each preset's damage (use the damage breakdown in the result JSON).

### Where findings go

- **Shared engine, shared raid or racials:** to the orchestrator (session_01AGmMcVxGoubbDodiQDQwhz) right away, with the evidence and the proposed fix. These touch every PR, so the orchestrator decides and lands them on the shared branch.
- **Class-specific findings:** to that class's worker, copying a one-line summary to the orchestrator.
- **After each full pass:** send the orchestrator a summary, then schedule the next pass in about 45 minutes.

Prioritise anything that moves DPS by 1% or more, or that changes relative standing between specs. Say how sure you are and which source tier settles it. Don't repeat a finding a worker has already answered unless you have new evidence.

## Pet buffs

A separate temporary research session (session_01FXVP4wddQ4UAWbse2E1uAG) owns the question of which buffs apply to hunter pets and warlock demons. Leave that topic to it.
