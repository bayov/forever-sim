# Healer mechanics: pass 8 (2026-11-10 UTC, master da272241e5, client 1.60.1.70888)

Due for the client build 1.60.1.70862 -> 1.60.1.70888. No watched code changed on master
(ca5e35c3f..da272241e5 touches item procs, logcheck and audit docs only). One new finding.

## Get current

- wago.tools newest build 1.60.1.70888; `wago.py diff --set mechanics 70862 70888` (the `--by-spell` mode crashes:
  KeyError 'SpellID' on a table without it, `tools/sources/wago.py:297`).
- `orchestration/update_ledger.md`: unchanged since 2026-11-08.
- Blizzard Nov 9 hotfix post (news 24315281): druid lines are Feral Charge and (Nov 6) Bloom, Swiftmend tooltip; nothing
  healer-numeric. foreverchanges.pro Nov 10 client hotfix: Nature's Swiftness now reaches Wild Growth (finding 12).

## Client 70862 -> 70888, healer rows

| Row | Change | Sim | Verdict | Tier |
|---|---|---|---|---|
| Swiftmend 18562 tooltip | "of the consumed spell" dropped | keeps the HoT under Forever (bc8e5473b) | match, now client agrees too | 1 |
| Natural Instinct 1223242 eff 0 (1211036) | crit-damage mask word 3 0 -> 0x88 (Pounce) | healing part 1361117 (12/25% Int) unchanged | healer part match; feral part is mechanics' | 1 |
| Fear Ward 39554 | castable in Shadowform | not a heal | n/a | 1 |
| Serenity Field 272439 (trinket, ilvl 65) | stat 41 15000 -> 7500, use CD 90 -> 360 s | not in the sim or any healer gear set | n/a (gear auditors) | 1 |
| Nature's Swiftness 17116 | wago mask unchanged; Nov 10 hotfix adds Wild Growth | HT, Regrowth, Wrath only | mismatch, finding 12 | 1 (hotfix cache, via foreverchanges.pro) |

No other heal, healing talent, HoT, cost, cast time or cooldown row of the four healers changed
(SpellPower, SpellCastTimes, SpellDuration, SpellLevels: 0 changes).

## Measured (finding 12)

`tools/healbench`, 1000 iterations, seed 101, patched `affected` in `registerNaturesSwiftnessCD` to take Wild Growth:

| Request | Carry before | after | HPS | Mana left |
|---|---|---|---|---|
| druid_restoration_raid, -role raid | 0.866 | 0.875 (+1.0%) | 2518 -> 2527 | 948 -> 865 |
| druid_restoration, -role tank | 1.354 | 1.354 | 845 -> 845 | same |

## Open findings

- 12 (new, medium): Nature's Swiftness misses Wild Growth since the Nov 10 hotfix.
- 11 (Riptide cost): still open, `sim/shaman/healing.go:117,132` scales by 0.6 (147/180/231 vs 150/180/230).
- 10 (untested heal paths): unchanged.

## Notes for the next pass

- healbench needs `protoc --go_out=./sim/core ./proto/*.proto` in a fresh tree; setup.sh installs protoc but doesn't run it.
- Confirm finding 12 against wago once a build carries the 17116 mask change.
