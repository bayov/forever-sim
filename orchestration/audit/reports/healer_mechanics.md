# Healer mechanics: pass 10 (2026-11-12 UTC, master cb0d928fd6, client 1.60.1.70888)

Due for 3 watched files changed on master (33db32b36..cb0d928fd6), all in one commit (cb0d928fd, talent tooltips).
Client build unchanged (70888). Quiet: no new finding; one resolved, one updated.

## Get current

- wago.tools newest build still 1.60.1.70888; no new client rows.
- Blizzard posts: none newer than the Nov 10 hotfixes already covered in pass 9 (`dev_notes_seen`); not re-fetched this pass.
- `orchestration/update_ledger.md`: unchanged since pass 9.

## What changed

| File | Change | Healer effect | Verdict |
|---|---|---|---|
| `sim/priest/priest.go:9` | `TalentTreeSizes` Shadow 18 -> 17 | none: `FillTalentsProto` (`sim/core/character.go:879`) adds each size as an offset, so the last is never read | OK |
| `sim/shaman/talents.go:206-216` | `ImprovedFireNova` -> `TrialByFire` (proto field 9 renamed, same number) | Fire Nova only, not a heal; talent strings unchanged | OK |
| `sim/druid/talents.go:362` | Rend and Tear comment | none (feral) | OK |

Healer text in the same commit (`tools/forever_talents/notes/shaman.json`), checked against the sim and verdicts:

| Tooltip | Text | Sim / client | Verdict | Tier |
|---|---|---|---|---|
| Riptide r1 | heals 445 over 15 sec, Chain Heal +25% | 89 x 5 ticks, 3 s; bonus row | match | 1 |
| Riptide r1 cost | 147 Mana, 12 sec cooldown | sim 147, client 70790+ 150; 12 s match | finding 11 (update added) | 1 |
| Mana Tide Totem r1 | 13 sec, 88 every 3 s, party in 30 yd | `core.ManaTideTotemDuration` 13 s; 88 | match | 1 |

## Tests

`go test -tags=with_db` of ./sim/shaman/..., ./sim/priest/..., ./sim/druid/..., ./sim/paladin/... and ./sim/mage/: all pass.
The stale Ironweave golden of finding 13 was regenerated in 3335001a5, so finding 13 is resolved. (Without `-tags=with_db`
the suites panic on missing items; that is the setup, not a regression.)

## Findings

- Resolved: healer_mechanics-20261111-13 (3335001a5).
- Updated: healer_mechanics-20261104-11, the new Riptide tooltip prints the sim's 147 instead of the client's 150.
- Still open: 10 (untested mechanics), 11.
