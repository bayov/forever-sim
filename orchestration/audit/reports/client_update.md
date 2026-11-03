# Client update pass: 1.60.1.70707 -> 1.60.1.70753

- Pass 7, 2026-11-03, master a0b306984d. `wago.py latest --probe 300` returns 1.60.1.70753.
- For `.agents/skills/dps-report/state.json`: `foreverClientBuild` -> 1.60.1.70753 (the orchestrator updates it).
- Result: findings (1 new, low). 70753 ("the November 2 build" on foreverchanges.pro) writes the launch realms' October 30 hotfix into the client: the druid and paladin tree rework, the 10-26 racial pass, and the mage, priest, warlock and hunter changes. The sim had landed all of them from the hotfix cache, and every DPS value checked matches the client now. The one new DPS mechanic is Poised.

## Client diff

`table_summary.py 70707 70753`: 1587 spells changed, 1374 of them only `SpellMisc.Attributes_14` 0 -> 64. Trait* changes in the paladin (1100), druid (1089) and mage (1112) trees. ItemSparse +40 (recipes, Pouch of Violet Dust, Enchant Boots - Speed) / ~4 (Toughened Leather Armor, low level). SpellItemEnchantment +1 (boots Speed 8%), ~1 (Defense text). ChrRaces: FactionID only. `items_diff.py`: no level 60 or request item changed. Blizzard dev notes: no new post since /8 (2026-10-27).

## DPS-relevant changes and sim status

| Change (70753 client) | Sim status |
|---|---|
| Leader of the Pack, Moonkin Aura: aura 290, crit with everything | landed (`sim/core/buffs.go:292`, from the hotfix) |
| Paladin: Conviction all crit, 3 ranks; Holy Power no crit; Divine Precision 2 ranks; Champion of the Light 5 ranks + Hammer of Wrath cast; Swift Judgement; Unsealed Wrath cost cut 0; Echo auras 3 stacks / 15 s | landed (`sim/paladin/`) |
| Paladin, druid and mage tree MaxRanks (14 node entries) | match `ui/core/talents/trees/*.json` |
| Druid: Shifting Power 8 s, 1 rank; Predatory Discipline; Heart of the Wild / Furor swap | landed (`sim/druid/`) |
| Mage: Arcane Blast 12% Int / 100% mana; Ice Lance 300% frozen, 8 s cd on every rank; Winter's Chill 20%; Frigid Power 15% | landed (`sim/mage/`) |
| Racials: Ruthless 1%, Artisan's Precision 1% hit, Endurance no crit damage, Human Spirit 10%, Command, Berserking 15 s, Touch of the Grave 15 s ICD, Wind Blessed 2% | landed (`sim/core/racials_forever.go`) |
| Poised (rogue 18, druid 32): 10 Energy off the first ability at full Energy out of combat | **missing**, finding client_update-20261103-1 (ROG-U-01, DRU-U-15) |
| Camp Chair: one 2% all-crit aura 290, the same as Moonkin Aura and Leader of the Pack | minor: the sim drops it under Moonkin Aura only; no request uses the chair; notes.md already asks |

n/a and the full list: the ledger's "Build 1.60.1.70753" section. Open findings 20261030-2 (Missile Barrage), 20261030-4 (Fengus' Ferocity) and 20261027-3 (Amplify Curse): 70753 changes none of those spells beyond Attributes_14, so they stay open.

## Left for later

- `sim/priest/shadow_word_death.go:15` and `sim/core/buffs.go:292`'s comments describe the pre-70753 client (10% backlash; "the client's own tables still have the Era split"). Comment-only, no finding.
