# Cross-class standards (orchestrator rulings)

The rules every class follows so the specs' Level 60 numbers compare fairly. Newest rulings win; this file is the reference. Each class PR states how it applies them.

## Gear

1. Budget formula: epic item value V(ilvl) = 1.2 * ilvl - 36.4; a piece costs (sum (weight * amount)^1.5)^(2/3); weights Str/Agi/Sta/Int/Spi 1.0, spell power 0.55 (school-only lower), AP 0.5, ranged AP 0.4, hit rating 1.0/pt (10 = 1%), crit rating 1.0/pt (14 = 1%); armor free per slot and armor type. Slot mods: 1.0 head / chest / legs / two-hand, 0.77 shoulder / hands / waist / feet, 0.5625 neck / back / wrist / finger / off-hand / shield, 0.42 one-hand, 0.32 ranged / wand / relic.
2. One item-level table for every class (fairness P5, 12:00 ruling): head, shoulder, chest, hands, legs, feet at item level 66; neck, back, wrist, waist and both rings at 71. That makes the twelve armor slots cost the same for every spec (about 402.5).
3. One weapon budget per hand style (fairness P4): two-handers at item level 76 (budget 54.8). Physical one-handers at item level 71 each. A caster's weapon slots total about 73 whether it's a main hand (the Azuresong rule: the real Era weapon's stats at Era prices, 46.2) plus an off-hand at 71 (26.8), or a staff (give the staff the same ~73). Ranged / wand / relic at item level 71 times 0.32, and every spec fills its ranged / wand / relic slot with something that does something for it (no dead libram, no empty idol slot). Weapon damage and speed come from the real Era Phase 1 weapon and are free; the weapon type stays the Era Phase 1 BiS type.
4. Hit (fairness P2): every set optimises hit against its other stats at budget prices: usually that means hit up to the spec's cap against a level 63 boss (yellow cap for melee, 16% spell miss for casters), counting talents and the 2-piece bonus, never over the cap, but if the same budget points do more as another stat (e.g. spell power for Balance), use the mix that sims best and say so.
5. Spirit about 10% of the set for mana users, none for rage and energy users.
6. Set bonuses (fairness P3): the Level 60 presets carry 2-piece bonuses only. Each class keeps its real Forever Tier 1 / crafted set on exactly two synthetic pieces so only the 2-piece is live (same shape for everyone: about 1% haste or 1% hit). Report what the 4- and 5-piece bonuses would add in a separate line in the presets comment, wiki and PR.
7. Every slot carries the best Phase 1 enchant that exists for it, the Forever necklace enchants included; no item or enchant whose only source is an unopened raid (Arcanum of Focus is Zandalar: use the best non-raid head enchant, e.g. Voracity; Enchant Weapon - Spell Power's formula 18259 is Molten Core only: the non-raid caster stat option is Mighty Intellect, Thorium Brotherhood). Mighty Intellect (Thorium Brotherhood vendor, Revered) is allowed, noted as borderline. Forever's Enchant Weapon - Revelation is NOT used by casters other than the enhancement shaman (its sim trigger list was written for shaman shocks and misfires on other casters); a PR may report it as an unverified alternative. Casters may use a staff (Forever's Enchant 2H Weapon - Spellblasting, +65 SP, vendor) or a main hand + off-hand, both at the ~73 caster weapon budget, whichever sims best. Forever's high-level enchants with non-raid sources (2H Spellblasting / Mighty Spell Power, Bracer Greater Spellpower +16 SP, Gloves school power +20 / Superior Strength +15, Cloak Agility +5, Off-Hand Wisdom, Mystic / Forceful Rugged Armor Kits) are added on the shared branch; re-pick enchants after merging it. Non-raid caster picks found so far: Mystic Thick Armor Kit (+5 SP) on chest / hands / legs / feet, Lesser Healing Power (6 spell damage) on wrist; no Phase 1 caster shoulder or back enchant.

## Consumables

8. Forever launch Phase 1 only: nothing whose only source (item, recipe, plans or quest) is an unopened vanilla raid (MC, ZG, BWL, AQ, Naxx). World, dungeon, vendor, quest, Onyxia and Blasted Lands quest items are fine. Elixir of Frost Power (Winter Veil recipe, tradeable) stays. No holiday soulbound items.
9. Forever's new consumables are simulated on the shared branch claude/wow-forever-dps-presets-m9u15h-consumes (merge its latest head).
10. At most ONE of the five +2% crit elixirs per character: Mongoose, Grizzly, Owl, Sages, Whale. Other slots take elixirs without that crit. State what a second crit elixir would add.
11. Every spec gets the full raid treatment: flask, the best elixir per slot, food, weapon oil / stone / imbue, potion, rune or equivalent, and explosives when its professions allow them.

## Raid, encounter and settings

12. The shared mksettings raid and encounter (level 63, 3731 armor, Humanoid; 120 / 180 / 300 sec); each spec in the group it would really be in. Curse of the Elements covers all magic schools; no Curse of Shadow; Improved Scorch / Winter's Chill only for the mage; Shadow Weaving only for the priest.
13. Pre-pull is allowed for everyone (buffs, potions, totems, imbues, pets, pre-casts), but damage only counts from time 0, so pre-casts are timed to land at 0 or later.
14. Pets: hunter pets don't get player buffs (Forever hunter rework); warlock demons do.
15. Final numbers at 20,000+ iterations from binaries rebuilt after the last gear or DB change.

## Races

16. Presets use the best race on the spec's Era BiS weapon type. Race tables give best overall, Horde and Alliance. The cross-spec ranking also gets a column with weapon-type racials off (fairness P6), so the weapon type doesn't decide the spec order.
