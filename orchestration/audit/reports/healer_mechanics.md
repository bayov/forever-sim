# Healer mechanics: pass 6 (2026-11-08 UTC, master 9ba00c2dc4, client 1.60.1.70862)

Quiet pass. Due because 11 watched files changed on master (all `sim/shaman/*`), 73d5bfa9e1..9ba00c2dc4.

## Get current

| Source | Seen | Healer content |
|---|---|---|
| wago.tools latest build | 1.60.1.70862 (unchanged) | none new |
| Dev notes thread 2360696 | re-fetched 2026-11-08; still 10 posts, last 2026-11-04T19:35Z | none new |
| update_ledger.md | no change since pass 5 | none |

## What changed (4 commits on the watched paths)

| Commit | Change | Healer effect |
|---|---|---|
| 0193721bc | `ForeverDownrankScope` = heals only (user ruling 2026-11-08); damage ranks keep full coefficient | none: every heal config (priest, shaman incl. Healing Stream Totem, druid, paladin incl. Holy Shock heal and Penance heal bolts) has `ProcMaskSpellHealing`, so `DownrankAuto` still reads them as heals and they keep the penalty. Holy Shock damage and Penance damage bolts left it, as ruled. Docs `healing_*.md` updated in the same commit. |
| 2a60326a1 | `sim.Roll` rounds to the nearest whole point | heals roll whole points too (all heal base rolls use `sim.Roll`); the average holds for whole ends. Evidence is damage logs only; heals assumed alike (verdict row added). Probability rolls moved to `RandomFloat`; no heal or raid-damage path used `Roll(0,1)` (`raid_damage.go` uses `RollWithLabel`, unrounded, fine for boss damage). |
| 5a870eccc | Stormstrike mark after the damage calc | none (Enhancement) |
| 9ba00c2dc | Windfury and Flametongue Totem replace each other in the weapon totem slot | none: weapon totems only; water totems and `totem_receivers.go` party reach for Mana Spring/Healing Stream unchanged |

## Tests

`go test -tags=with_db` on `sim/core` (Downrank, Heal, Absorb, RaidDamage), `sim/paladin/holy`, `sim/priest/...`,
`sim/druid/...`, `sim/shaman/...` and `sim/` (Heal, Raid, Totem, Mana, Rank, Swiftmend, Riptide, Bloom, Shield): all pass.
`TestDownrank` covers "auto reads a heal from its ProcMask" and "auto damage is out of a heals-only scope";
`TestDruidDownrankedHeals` and `TestHealingPriestRanks` hold the heal side. (Without `-tags=with_db` several heal tests
panic on missing items; that's the build tag, not a bug.)

## Open findings

- healer_mechanics-20261104-11 (Riptide's mana 147/180/231 vs client 150/180/230): still open, `sim/shaman/healing.go` unchanged.
- healer_mechanics-20261104-10 (untested heals, talents, engine paths): still open; no new gap from this change.

## New findings

None.
