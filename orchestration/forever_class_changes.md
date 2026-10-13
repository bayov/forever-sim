# Forever class changes audit

Class-level and cross-class mechanics that WoW Forever (client 1.60.1.70338) changed from Classic Era (1.15.9.70003), checked against the sim on each class branch. Kept up to date by the Forever class changes auditor.

Source tiers: 1 client data on wago.tools (db2 tables); 2 Wowhead Forever tooltips (dataEnv=17); 3 Blizzard posts; 4 foreverchanges.pro; 5 the repo's wiki and notes.

Measurements are paired sims on each class branch's own build: the same mksettings settings run with and without the change, through a local measurement patch that is never committed. 20,000 iterations, 180 sec, the Level 60 generic boss. The error is about +-0.2 to 0.5 DPS.

## Open findings (pass 1, 2026-10-13)

| # | Finding | Tier | Classes | Sim status | DPS effect (180 s) | Sent to |
|---|---|---|---|---|---|---|
| R1 | Poisons take the class-imbue slot (SpellEffect 360). Sharpening stones and oils keep the temporary-enchant slot (54). So a rogue weapon carries a poison and a stone together. | 1 | Rogue | Wrong: poisons share Consumes.MainHandImbue / OffHandImbue with stones (sim/rogue/rogue.go:220-229), and no preset has a stone. | Dense Sharpening Stone on both hands: Combat Daggers +13.0 (+1.9%), Combat Swords +9.5 (+1.4%) | Rogue |
| M1 | Mage-only Scrolls of Imbue (277500 Greater Flame +20 Fire, 277501 Greater Frost +20 Frost, 277502 Precision +1% spell crit; level 46, staff) take the class-imbue slot (360), so they stack with Wizard Oil. Source is probably the new Comprehension skill; launch availability is unconfirmed. | 1, 2 | Mage | Ignored: the UI lists them as not simulated, in the oil's picker | Frostfire +11.4, Fire +10.8, Frost +10.9 (+1.2 to 1.5%), Arcane +6.2 (Precision) | Mage |
| P1 | Hammer of the Righteous (407632) is a level 40 paladin trainer spell, not a talent: 3x main hand DPS as Holy, 6 sec cooldown, 6% base mana. | 1, 2 | Paladin (Ret) | Ignored: "Not modelled" (sim-status.md:266) | Not measured; estimate +30 to 40 (about 4%) if GCDs and mana allow | Paladin |
| H1 | Hunter's Mark III (98) is stronger than IV (71), and Trueshot Aura IV (75) stronger than V (50). Standards 8a (strongest learned rank) applies. | 1, 2 | Hunter | Wrong: sim/core uses the top rank (debuffs.go:270-271, buffs.go:1624-1625) | BM +10.6, MM +12.3, Lone Wolf +14.3 (+1.1 to 1.4%), SV 0 | Hunter |
| H2 | Pet Bite rank 8 does 81-99 (sim: 81-91). | 1, 2 | Hunter | Wrong | Small, not measured | Hunter |
| D1 | Every shapeshift form can use items (SpellShapeshiftForm flag 0x80: Cat 112->216, Bear 80->216, Moonkin 65->201). Cat and Bear keep the "don't use weapon" flag 0x10. The Windfury Totem buff is now a player aura, not a weapon enchant. | 1 | Druid (Feral) | Wrong: potions and runes are refused in form; feral gets no Windfury Totem, stones or oils | Depends on the druid session's form model | Druid (already on forms) |
| S1 | Shadow Oil's proc (350026) does 63-71 and Frost Oil's (350025) 39-43. Forever scales them from level 0. | 1, 2 | Ret, Enhancement (Shadow Oil) | Wrong: sim uses Era 52-61 / 33-38 (shared consumes.go) | About +7 for Ret (Shadow Oil is 37.3 of 981), similar for Enhancement | Orchestrator (shared) |
| S2 | Dragonbreath Chili rolls 61-69 (base 65, variance 0.123). | 1 | Every spec using it | Wrong: flat 60 (shared consumes.go:521) | About +1.5 | Orchestrator (shared) |
| S3 | Faerie Fire and Curse of Recklessness are exclusive under Forever in the sim, based only on a user call (shaman_audit.md:503). The client shows no exclusivity: no wording; the same aura 674 as Sunder and Expose; they stacked in Classic and TBC. | 1 (no evidence for exclusivity) | Every physical spec | Possibly wrong | Both stacking: boss armor after 5 Sunders 976 -> 471, about +8.5% physical damage | Orchestrator (user ruling to confirm) |
| S4 | Forever enchants missing from the database: Bracer Superior Agility (1248599, +9 Agility) and 2H Weapon Strength (1248668, +25 Strength). Launch availability is unconfirmed. | 1, 2 | Rogue, feral, hunter; Arms, Ret and 2H Enhancement | Ignored | Not measured | Orchestrator (shared database) |
| G1 | Gnome Eureka! only covers the spells in each class's spell mask. For example, the warlock mask leaves out Corruption, Curses, Siphon Life and Incinerate, the priest's SW:Pain, the rogue's Rupture and Garrote, the warrior's Rend. | 1 | Gnome warlocks (presets) | Wrong (sim/core/racials_forever.go:206-215) | Warlock -0.3 to -0.8, under the error; Gnome stays first | Warlock (FYI) |
| W2 | Siphon Life is the only class DoT without the Forever "periodic can crit" attribute. | 1 (weak; Pandemic's text says otherwise) | Warlock | Sim lets it crit | Small | Warlock (FYI, leave unless a log shows otherwise) |

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
