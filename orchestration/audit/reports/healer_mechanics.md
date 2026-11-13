# Healer mechanics, pass 13 (2026-11-13)

Master 97da1cca5c against last pass 71562a8b03. Client 1.60.1.70938. Result: findings (1 new, high). Findings 10 and 11 still open.

## Get current

| Source | Seen | Result |
|---|---|---|
| Newest client (`wago.py latest`) | 1.60.1.70938 (was 70888) | `wago.py diff 1.60.1.70888 1.60.1.70938 --set mechanics`: Wild Growth coefficient doubled (finding 15); Nature's Swiftness 17116 mask gains Wild Growth |
| Blizzard notes | thread t/2403040 | not re-read: forum returned "performed this action too many times" (rate limit) twice. Next pass: look for a note on Wild Growth's coefficient |
| `orchestration/update_ledger.md` | no 70938 entry yet | nothing to take |

## Client 70888 -> 70938, healer rows

| Row | Change | Sim | Verdict | Tier |
|---|---|---|---|---|
| Wild Growth 408120 / 1238214 / 1238215 (SpellEffect 1072065, 1235327, 1235329) | EffectBonusCoefficient 0.033 -> 0.066 a tick; base/rppl/ticks unchanged | `sim/druid/healing.go:174` 0.033 | **mismatch** (finding 15) | 1 |
| Nature's Swiftness 17116 (SpellEffect 695239) | mask_1 0x02000000 -> 0x06000000; 0x04000000 is Wild Growth's class bit (SpellClassOptions 408120 mask_1 67108864) | NS takes Wild Growth since 9628e85b6 | match (client now agrees with the Nov 10 hotfix) | 1 |
| Divine Spirit / Prayer of Spirit class mask_1 +0x00400000 | new class bit (Reagent Economy 1273052 now reaches it) | reagent cost only | n/a | 1 |
| Drink 1330734 (new, ItemEffect 238817) | new drink spell, 489 mana per tick aura 85 | consumables (healer_gear's domain) | n/a | 1 |
| Holy Power 1310077 (renamed from Greater Firepower), Gloves of Holy Might, Poised, Call of Water | Holy damage, DPS or dummy rows | no healer path | n/a | 1 |

No other heal rank (Healing Touch, Regrowth, Rejuvenation, Tranquility, Swiftmend, Bloom, Healing Wave, Lesser
Healing Wave, Chain Heal, Riptide, Holy Light, Flash of Light, Holy Shock, Heal, Greater Heal, Flash Heal, Renew,
Prayer of Healing, Power Word: Shield, Penance), healing talent row, cast time, cost, cooldown or duration changed.

## Measured (finding 15)

`healbench -role raid -carry tools/healbench/req/druid_restoration_raid.json`, master vs patch (coef 0.066), paired seeds, 1000 iterations a step:

| Build | HPS | EH per mana | Overheal | Carry |
|---|---|---|---|---|
| master (0.033) | 2583 | 3.65 | 60.3% | 0.875 |
| patch (0.066) | 2696 | 3.91 | 61.8% | 0.892 (+1.9%) |

The druid tank preset casts no Wild Growth.

## What changed (watched paths, 71562a8b03..97da1cca5c)

| File / commit | Change | Healer effect | Verdict |
|---|---|---|---|
| `sim/paladin/unsealed_wrath.go` (33b13e310) | Windfury Totem swings store no Echo (Nov 11 hotfix) | Retribution talent; Holy presets don't take it | n/a (mechanics auditor) |

## Tests

`sim/druid/healing_ranks_test.go:87` pins the coefficient at 0.033, so the fix must change it; with the patch `go test ./sim/druid/` passes. Finding 10 unchanged.

## Open items

- Blizzard notes not read this pass (rate limit): next pass, check the thread for the Wild Growth change.
- Next pass: re-check finding 15 when `sim/druid/healing.go` changes; finding 11 when `sim/shaman/healing.go` changes.
