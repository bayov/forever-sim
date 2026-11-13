# Healer mechanics, pass 12 (2026-11-13)

Master 71562a8b03 against last pass 8306f1d52. Client 1.60.1.70888. Result: quiet. Findings 10 and 11 still open.

## Get current

| Source | Seen | Result |
|---|---|---|
| Newest client (`wago.py latest`) | 1.60.1.70888 | Same as pass 11, no verdict reopened |
| Blizzard notes | Nov 11 hotfixes / blue post (`ref/client_update/sources/blizzard_forum_blue_2026-11-11.txt`) | Healer lines: Nature's Swiftness takes Wild Growth (in sim since 9628e85b6), Revelation off Vampiric Embrace (pass 9), Power Infusion persists on allies, Shadowform/Fear Ward, Lesser Spirit enchants (pass 9). None changes a heal, healing talent or engine path the presets use |
| `orchestration/update_ledger.md` | last entry build 70888 | Nothing new |

## What changed (watched paths, 8306f1d52..71562a8b03)

| File / commit | Change | Healer effect | Verdict |
|---|---|---|---|
| `sim/core/heal_targets.go:80-88`, `apl_values_resources.go` (65b5d49cf) | Current Health / Current Health % of Most Hurt and Lowest Health read the healer's seen health; dead assignment reads full | closes finding 14; test `sim/healing_health_values_test.go` | OK (already resolved in findings) |
| `sim/druid/healing.go:740` (e05135ad3) | NS comment names Wild Growth | comment only; mask code from 9628e85b6 matches the Nov 10 hotfix (tier 1) | OK |
| `sim/paladin/judgement.go`, `seal_of_fury.go`, `soc.go`, `sor.go` (f1c07e13a) | Judgement damage spells are ranged specials under Forever (Nov 11 hotfix, tier 1) | none for Holy: healer presets don't judge for damage; raid Judgement of Wisdom / Light debuffs still proc from the raid's melee | OK (DPS/tank auditors own it) |
| `tools/healbench/req/*.json` (ea25e50df, 8fd5ddcb9, e0e351f2f, 7b8a41fcc) | consumables (Greater Cleric's Elixir, food, Lesser Mana Oil), APL re-tunes (shaman NS -> instant Healing Wave, Holy priest Twilight Focus, paladin Flash of Light, druid NS at 45%) | rotation and consumables only; consumables belong to healer_gear (finding healer_gear-20261112-1 fixed) | OK |

No heal rank, healing talent row or engine path (health, targets, absorbs, HoTs, regen, raid damage) changed
beyond the above. `sim/shaman/healing.go` unchanged, so finding 11 (Riptide 147/180/231 vs client 150/180/230)
stands; finding 10 (untested paths) unchanged.

## Open items

- None blocking. Next pass: re-check finding 11 when `sim/shaman/healing.go` changes, and any new client build.
