# Forever class changes audit

Class-level and cross-class mechanics that WoW Forever (client 1.60.1.70338) changed from Classic Era (1.15.9.70003), checked against the sim on each class branch. Kept up to date by the Forever class changes auditor.

Source tiers: 1 client data on wago.tools (db2 tables); 2 Wowhead Forever tooltips (dataEnv=17); 3 Blizzard posts; 4 foreverchanges.pro; 5 the repo's wiki and notes.

Measurements are paired sims on each class branch's own build: the same mksettings settings run with and without the change, through a local measurement patch that is never committed. 20,000 iterations, 180 sec, the Level 60 generic boss. The error is about +-0.2 to 0.5 DPS.

## Open findings (pass 1, 2026-10-13)

| # | Finding | Tier | Classes | Sim status | DPS effect (180 s) | Sent to |
|---|---|---|---|---|---|---|
| R1 | Poisons take the class-imbue slot (SpellEffect 360). Sharpening stones and oils keep the temporary-enchant slot (54). So a rogue weapon carries a poison and a stone together. | 1 | Rogue | FIXED (rogue, head 1e012bb16): poisons in RogueOptions main_hand_poison / off_hand_poison beside the stone; Dense Sharpening Stone on both blades; pinned by TestPoisonBesideSharpeningStone. Finals at 180 s, Undead: Combat Daggers 708.1, Combat Swords 679.1, Seal Fate Daggers 674.4, Mutilate 669.8, Seal Fate Swords 644.7, Subtlety 643.6. Axes: an Orc with axes is +6.2 on Combat Swords (1.5 over Undead), but the preset stays on swords (standard 16). | Combat Daggers +13.0 (+1.9%), Combat Swords +9.5 (+1.4%), re-measured by the rogue session | Rogue |
| M1 | Mage-only Scrolls of Imbue (277500 Greater Flame +20 Fire, 277501 Greater Frost +20 Frost, 277502 Precision +1% spell crit; level 46, staff) take the class-imbue slot (360), so they stack with Wizard Oil. Source is probably the new Comprehension skill; launch availability is unconfirmed. | 1, 2 | Mage | FIXED (mage 93ac238): new Mage option staff_imbue, which stacks with Wizard Oil; presets take Greater Flame (Fire, Frostfire), Greater Frost (Frost), Precision (Arcane). Launch availability is labelled as an assumption (source: the mage Comprehension skill, SkillLine 3012). Cryoblast scroll not added. | Frostfire +11.4, Fire +10.8, Frost +10.9, Arcane +6.2 (re-measured by the mage session; race order unchanged) | Mage |
| P1 | Hammer of the Righteous (407632) is a level 40 paladin trainer spell. CLOSED (paladin session): it needs a one-hand axe, mace or sword (SpellEquippedItems subclass mask 145), and it shares cooldown category 2404 with Holy Strike. | 1, 4 | Paladin (Ret) | Correctly left out | 0 (every Ret preset uses a two-hander) | Paladin |
| H1 | Hunter's Mark III (98) is stronger than IV (71), and Trueshot Aura IV (75) stronger than V (50). Standards 8a (strongest learned rank) applies. | 1, 2 | Hunter | Wrong: sim/core uses the top rank (debuffs.go:270-271, buffs.go:1624-1625). The hunter session agreed and has the fix ready; the shared sim/core edit waits on the user's approval. | BM +10.6, MM +12.3, Lone Wolf +14.3 (+1.1 to 1.4%), SV 0 | Hunter |
| H2 | Pet Bite rank 8 does 81-99 (sim: 81-91). | 1, 2 | Hunter | FIXED (hunter 2669b4d, finals re-run 302fe1f) | MM +0.4, BM +0.7, SV +0.4 | Hunter |
| D1 | Every shapeshift form can use items (SpellShapeshiftForm flag 0x80). Weapon enchants and procs, Windfury Totem included, work while shapeshifted (Blizzard deep dive, per the druid session). Tiger's Fury is removed (deep dive; nothing in client 70338 grants it). | 1, 3 | Druid (Feral) | FIXED on the druid branch (bd9d11e, 63e7c37, 28da841e9, 43cf0dd38): potions, runes and Thistle Tea usable in form, cat swings the weapon's DPS, imbues work in form, Windfury Totem replaces Grace of Air in Feral's melee group, Dense Weightstone replaces Elemental Sharpening Stone (Molten Core plans) | Windfury Totem +13.3 over Grace of Air; Dense Weightstone -4.6; Tiger's Fury removed -24.9; Feral 717.2 at 180 s | Druid |
| S1 | Shadow Oil's proc (350026) does 63-71 and Frost Oil's (350025) 39-43. Forever scales them from level 0. | 1, 2 | Ret, Enhancement (Shadow Oil) | Wrong: sim uses Era 52-61 / 33-38 (shared consumes.go) | About +7 for Ret (Shadow Oil is 37.3 of 981), similar for Enhancement | Orchestrator: being implemented on the shared consumes branch |
| S2 | Dragonbreath Chili rolls 61-69 (base 65, variance 0.123). | 1 | Every spec using it | Wrong: flat 60 (shared consumes.go:521) | About +1.5 | Orchestrator: being implemented on the shared consumes branch |
| S3 | Faerie Fire and Curse of Recklessness are exclusive under Forever. CLOSED: the user re-confirmed they stay exclusive (standards 20). The client data shows no exclusivity either way. | user ruling | Every physical spec | Right by ruling | 0 | Orchestrator |
| S4 | Forever enchants missing from the database: Bracer Superior Agility (1248599, +9 Agility) and 2H Weapon Strength (1248668, +25 Strength). Launch availability is unconfirmed. | 1, 2 | Rogue, feral, hunter; Arms, Ret and 2H Enhancement | Ignored | Not measured | Orchestrator: added to the database on the consumes branch, unused until launch availability is confirmed |
| G1 | Gnome Eureka! only covers the spells in each class's spell mask. For example, the warlock mask leaves out Corruption, Curses, Siphon Life and Incinerate, the priest's SW:Pain, the rogue's Rupture and Garrote, the warrior's Rend. | 1 | Gnome warlocks (presets) | Wrong (sim/core/racials_forever.go:206-215) | Warlock -0.3 to -0.8, under the error; Gnome stays first | Warlock (agrees; the shared racials file is for the orchestrator to route) |
| W2 | Siphon Life is the only class DoT without the Forever "periodic can crit" attribute. CLOSED: forever-bugs #303 ("Siphon Life ticks could never crit") was fixed for the next build, so the missing flag is a client bug. | 1, 3 | Warlock | Right: sim crits, pinned by TestWarlockSiphonLifeCrits (warlock 55f3d3f) | 0 | Warlock |

## Downranking penalty (orchestrator request, 2026-10-13)

- **Rule.** Blizzard's 2026-10-01 development notes (tier 3): ranks far below your level get less spell power, and trigger class abilities less often. Foreverchanges (tier 4) gives the formula: a rank keeps 1 - 0.05 x (level - 17 - learn level) of its coefficient. It was measured on heals up to level 30, so its size at 60 is assumed. At 60 it hits ranks learned before level 43.
- **Which presets it touches.** I checked every Level 60 preset's casts (result JSON) and every Level 60 APL's spell IDs against SpellLevels (client 70338). Only two qualify:
  - Priest Smite fallback: done on the priest branch (6ee9d02eb, ea9eb5c22, 333cbeb04). It moved to rank 7. 180 s 708.8 -> 706.5, 300 s 687.0 -> 643.8.
  - Elemental shaman Lightning Bolt VIII fallback (level 38, below 20% mana). It keeps 75% of its coefficient. Estimate: 0 at 180 s (0.2 casts), about -5 to -7 DPS (-1%) at 300 s (15.6 casts, 58 DPS), before moving the fallback to rank IX (level 44).
  - Elemental: DONE on the shaman branch (651dcb6ce). It uses a local downrankCoefficient (same formula) on Lightning Bolt and Chain Lightning, and will switch to the core helper once it lands. The fallback moved up one rank, to one learned at 44 with the full coefficient. Net change against before the penalty: Level 60 -0.4 and Mana Tide +0.2 at 300 s, 0 at 120 and 180 s. Orc stays best. Enhancement is unaffected.
  - Every other low-level ID in the APLs is a single-rank Forever spell, an attack power ability, a cooldown or a stance.
- **Proposal.** One shared helper in sim/core, owned by the priest session, which already has it: `(*Character).ForeverDownrankCoefficient`. The shaman applies it to Lightning Bolt and Chain Lightning. It is in orchestration/spec_assumptions.md as `downrank_penalty` (stronger / FC formula / none).

## Confirmed already right in the sim

- Rogues can use one-hand axes (SkillLineAbility 196 adds the Rogue class). Hack and Slash covers axes and the UI allows them. Tier 1.
- Weapon slots: a totem buff, a class imbue (SpellEffect 360) and an oil or stone (54) all stack. Rogue, warrior, Ret and Survival presets get Windfury Totem next to their poisons, stones or oils. Windfury Weapon turns off the shaman's own Windfury Totem. Tier 1.
- Warlock Firestone and Spellstone are effect 360 weapon imbues that stack with an oil (on the warlock branch). Tier 1.
- Weapon and armor proficiencies are otherwise unchanged. Dual wield is warrior, hunter and rogue only. Thrown, wands and relics are unchanged. Tier 1.
- Stances and forms: warrior stance requirements are unchanged. Ambush, Garrote and Cheap Shot still need Stealth; Premeditation no longer does. The sim allows the new casts in Moonkin and the caster Faerie Fire in any form. Tier 1.
- GCD: only totems changed (1.0 sec). Nothing shows haste lowering the GCD. Cast times and cooldowns for DPS spells match each branch. Tier 1.
- Talents made baseline or removed: all handled on the branches (Omen of Clarity, Tactical Mastery, Improved Battle Shout/MotW/Fortitude built into the buffs, Consecration, Holy Strike, Blessing of Kings, Aimed Shot, Curse of the Elements, ...). Tier 1.
- Periodic damage can crit: Forever sets "periodic can crit" (Attributes_8 0x200) on 39 class DoTs, which matches sim/core/ruleset.go. The wiki still calls this unconfirmed. Tier 1.
- Resources: energy, rage and focus are unchanged. Spell cost changes are on the branches. Tier 1.
- Racials: all values match sim/core/racials_forever.go except G1. Every preset race/class pair is allowed. (Wiki error: Windshaper Skyborne has no Mage, only High Order.) Tier 1.
- Rank levels: unchanged for every vanilla rank, so no top rank at 60 moved. Tier 1.
- Proc chances and talent curves match for the warrior, rogue, shaman, mage, priest, warlock, paladin, druid and hunter procs checked, and for Crusader, Fiery and Lifestealing. PPM rates are server side, so client data can't check them. Tier 1.
- Personal and per-caster debuffs (Winter's Chill, Fire Vulnerability, Shadow Weaving, Improved Shadow Bolt, Stormstrike, Deadly Poison) match. Leader of the Pack and Moonkin Aura are exclusive. Windfury Totem and Grace of Air don't stack. Tier 1 and 4.

## Not decidable from client data

- PPM values, the spirit regen and crit-per-stat tables, glancing blows, and pet haste and crit inheritance (the pet scaling auras list them with server-side values). These need beta logs.
