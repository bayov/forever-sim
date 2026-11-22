# healer_mechanics pass 25 (2026-11-22, master 8b5137c6d)

Range b4b68c52e9..8b5137c6d6. Gate: 9 watched files (paladin consecration, judgement, seal_of_fury, soc, sor, talents;
shaman item_sets_pve, stormstrike, talents). Client still 1.60.1.71037 (wago latest). Update ledger: no new healer row
(PAL-HF-07 is Retribution). Blizzard t/2403040 post 14 (2026-11-21, "November 20") is rogue, warlock, dungeons,
Stormshroud Armor, quests and UI: no healer line.

## Changed files

| File | Change | Healer effect | Verdict |
|---|---|---|---|
| sim/paladin/talents.go | 2H Weapon Specialization Holy part school-wide under Forever (Nov 19 hotfix), item-swap aware | `SchoolDamageDealtMultiplier[Holy]` only; heals read `HealingDealtMultiplier` (`sim/core/spell_result.go` CasterHealingMultiplier), so Holy Shock's heal, Holy Light, Flash of Light untouched; no Holy preset wields a 2H with the talent | no healer effect |
| consecration.go, judgement.go, seal_of_fury.go, soc.go, sor.go | spell-side 2H multiplier dropped under Forever | damage spells only | none |
| sim/shaman/talents.go, stormstrike.go | Maelstrom Weapon / Stormstrike comments now cite client 70959 | Enhancement only | none |
| sim/shaman/item_sets_pve.go | comment: Dungeon Set 1 pieces in set 1679 | no healer preset wears 1679 | none |
| sim/core/buffs.go (unwatched) | Windfury Totem ICD comment only | none | none |

## Tests

`go test -tags=with_db ./sim/paladin/... ./sim/shaman/...` green (holy, protection, retribution, elemental,
enhancement, warden).

## Findings

None new. Open: 10 (untested paths, low), 11 (Riptide cost 147/180/231 vs client 150/180/230, low), unchanged (no
restoration code in range).
