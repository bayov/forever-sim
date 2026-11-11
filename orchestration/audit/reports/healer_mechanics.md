# Healer mechanics: pass 9 (2026-11-11 UTC, master 707ae92cc5, client 1.60.1.70888)

Due for 9 watched files changed on master (27d0d59d6..707ae92cc5). Client build unchanged (70888). One finding resolved,
one new (low).

## Get current

- wago.tools newest build still 1.60.1.70888 (gate.json); no new client rows to check.
- Blizzard's hotfix post (news 24315281, now "November 10"), healer lines, tier 1:
  | Line | Sim | Verdict |
  |---|---|---|
  | Nov 10 Druid: Nature's Swiftness also makes Wild Growth instant | `sim/druid/healing.go:752-760` takes Wild Growth (9628e85b6) | match; finding 12 resolved |
  | Nov 10 Priest: Revelation no longer activated by Vampiric Embrace healing | `sim/common/enchant_effects.go:289` procs only on ProcMaskSpellDamage direct casts; heals never reach it | match |
  | Nov 10 Holy: Holy Nova unlearned when untalented | talents gate the spell | n/a |
  | Nov 10 Enchanting: Lesser Spirit (boots, bracer) give their stated Spirit | server fix; sim takes the stated values | match (gear auditors' area) |
  | Nov 6 Druid: Bloom heals; Swiftmend tooltip | already covered in pass 8 | match |
- `orchestration/update_ledger.md`: no healer row since pass 8 (DRU-HF-01 is the Nature's Swiftness line).
- Shaman Nov 10 tuning (Stormstrike, Mental Dexterity, Maelstrom Weapon, imbue + weapon totem): no Restoration heal, healing
  talent or mana effect; the totem comment changes (`air_totems.go`, `fire_totems.go`, `shaman.go`, `totem_receivers.go`) are
  weapon buffs only.

## Changed code read in full

| File | Change | Healer effect | Verdict |
|---|---|---|---|
| `sim/druid/healing.go` registerNaturesSwiftnessCD | Wild Growth joins `affected` | instant Wild Growth, spends the buff | match, tier 1 |
| `sim/core/dot.go` + `aura.go` endedAt | CleanUp ticks only if the aura is active or the tick is at or before its end; new `SetTickPeriod` (Wrack only) | HoT last tick at expiry still lands (NextActionAt == endedAt); Swiftmend/cancel ends early as before; refreshes keep the aura active | match; druid, paladin, priest and raid healing tests pass |
| `sim/druid/items.go` | Idol of Ferocity -2 energy | feral only | n/a |
| `sim/shaman/talents.go` | Mental Dexterity, Maelstrom Weapon | enhancement only | n/a (mechanics) |

## Tests

`go test -tags=with_db` on `./sim/druid/... ./sim/paladin/... ./sim/core/ ./sim/` (healing filters) and
`./sim/shaman/enhancement/` healing tests: all pass. `TestDruidNaturesSwiftnessOnRegrowth` now fails if Wild Growth
leaves Nature's Swiftness. `./sim/priest/healing/` fails one row, `Phase1-AllItems-IronweaveBattlesuit` (Holy HPS 249.986 ->
252.680, Disc 109.002 -> 109.557), as do Shadow priest and Mage on the same row: a stale golden whose cause wasn't bisected
(finding 13). Note: without `-tags=with_db` these packages panic on missing items; always pass it.

## Findings

- New: healer_mechanics-20261111-13 (low): priest suites red on the Ironweave AllItems golden.
- Resolved: 12 (Nature's Swiftness on Wild Growth), 9628e85b6.
- Still open: 11 (Riptide cost 147/180/231 vs client 150/180/230), 10 (untested paths).
