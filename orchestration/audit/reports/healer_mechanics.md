# Healer mechanics: pass 3 (2026-11-05 UTC, master a079415316, client 1.60.1.70790)

Quiet pass. Due because two watched files changed on master since pass 2 (dc303b026).

## Get current

| Source | Seen | Healer content |
|---|---|---|
| wago.tools latest build | 1.60.1.70790 (same as pass 2) | none new |
| Dev notes thread 2360696 | post 10 (2026-11-04 19:35Z) and post 9 edit (19:35Z) | post 10 says it added a "User Interface and Accessibility" section to post 9 (swing timer, bags, nameplates, spellbook, PvP scoreboard, keybinds): no healing mechanic |
| update_ledger.md, tools/sources | no change in range | none |

## What changed on master (watched paths)

| File | Commit | Healer effect | Verdict |
|---|---|---|---|
| sim/core/spell_result.go (healingThreat) | 464deda7b | a heal on a tank in a tank-score replay makes threat from the whole heal, not effective healing, since the replay's health is the stand-in healer's (tank_mechanics-20261104-16). `readsNoHealth()` is `tankScore != nil && replay` (tank_score.go:93), so healing fights and healer presets are untouched | OK, tank-sim design; row added to verdicts Engine table |
| sim/druid/ferocious_bite.go | ff56f8ddc | feral damage per extra Energy (hotfix 113403); no heal, HoT, mana or Restoration path | out of scope |

Priest goldens (TestDisc hps 111.86 -> 112.17, TestHoly 205.71 -> 206.32) moved with the goldens' item-set change
(aa7cd24dc, c9973fb7f: hit-trade step items left out of AllItems), not a healing rule. No healer spell file changed.

## Open findings re-checked

- healer_mechanics-20261104-11 (Riptide mana 147/180/231 vs client 150/180/230): still open, `sim/shaman/healing.go:117,130-132` unchanged.
- healer_mechanics-20261104-10 (untested mechanics): no new healer or engine healing tests in range; still open.

## Findings

None new.
