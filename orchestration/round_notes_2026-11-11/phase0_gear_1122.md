# Phase 0 gear, step 1: tooling and sets, 2026-11-22

Brief: `briefs/phase0_gear_1122.md`. No search was run (step 3).

## What changed

- `tools/rotopt/phase0_specs.json`: one entry per DPS spec of the report (23): the report build, talents, start gear,
  weapon setup (two-hand, dual-wield, one-hand, any), dagger limits and output paths. Enhancement keeps its talents and
  timesaver start gear. Every other spec starts empty. Fury has a two-hand entry (`fury`) and a dual wield one (`fury_dw`).
  Mutilate holds daggers in both hands, Subtlety a main hand dagger. Casters, hunters and Feral are "any".
- `bis_phase0.sh` takes `SPEC` (default `enh`). New lists go to `ui/<dir>/bis/level60_phase0_<spec>.bis.json`.
  Only Enhancement writes a gear preset.
- `gear.go`: weapon types for hunter, druid, mage, priest and warlock. Ranged types: bows, guns, crossbows and thrown
  for hunters, wands, idols. Caster suffixes for mage, warlock, priest and druid, each by the schools it casts.
  Shaman and paladin suffixes are as before. New `-gear-weapons`, `-gear-mh-types`, `-gear-oh-types`
  (`applyWeaponRules`). Tests: `gear_tables_test.go`.
- Unlisted items: a pick or a top 5 item or enchant that `phase0_content.json` / `phase0_enchants.json` doesn't list
  now fails the run, after the list is written. It prints them all. `ALLOW_UNLISTED=1` goes on. The Enhancement
  list's 187 items that the 2026-11-12 check left open by not listing them are now listed by name, with a `checked`
  field that says so. Every enchant was already listed.
- The 8 Dungeon Set 1 sets (below).
- Dry run: `DRY_RUN=1 tools/rotopt/bis_phase0.sh` (`-gear-pool-report`, `phase0_pool.go`). It runs in 20s for all
  specs. Output: `phase0_pool/<spec>.json`.

## Enhancement check

Old and new scripts were run with a stub `go` that captures the rotopt call. The arguments, settings JSON and APL
are byte for byte the same in four modes: the full run, `RANK_ONLY=1`, `RANK_ONLY=1 SLOTS=trinket` and `EFFECTS_ONLY=1`.
The "two-hand" setup still passes only `-gear-rank-two-hand-only` and leaves the pool whole, as before. The shaman
pool tables didn't change. The post-run annotation on today's list gives the same list, and the new unlisted check
finds nothing.

## The 8 sets

From client 1.60.1.71037, `sim/common/item_sets/forever_dungeon_set_1_reveal.go`, with a test for each set. 2pc is +8
all resistances in all 8 (modelled). Every 5pc is a "when struck" proc, not modelled because a boss doesn't attack the
player in the sim.

| Set | Class | 3pc | 4pc | 6pc |
|---|---|---|---|---|
| Feralheart Raiment 1667 | druid | 8 mp5 | 200 mana on 5% of casts (the Energy and Rage parts aren't modelled) | +40 AP, +23 SP |
| Beastmaster Armor 1669 | hunter | 8 mp5 | 5% on auto attacks: 200 mana | +40 AP |
| Sorcerer's Regalia 1671 | mage | 8 mp5 | 5% on casts: 200 mana | +23 SP |
| Soulforge Armor 1673 | paladin | 8 mp5 | Crusader's Wrath, +65 SP for 10s | +23 SP |
| Vestments of the Virtuous 1675 | priest | 8 mp5 | 5% on casts: 200 mana | +23 SP |
| Darkmantle Armor 1676 | rogue | 15 health per 5s, not modelled | 20 Energy on a swing | +40 AP |
| Deathmist Raiment 1681 | warlock | 8 mp5 | 300 self-heal, not modelled | +23 SP |
| Battlegear of Heroism 1778 | warrior | 15 health per 5s, not modelled | heals 88 to 132, 10 Rage | +20 Strength |

Procs with no rate in the data use assumed rates: 3 per minute on swings, 4% of casts for Crusader's Wrath, and 1 per
minute for the energize and Warrior's Resolve. When a new set has the same name as a Classic set, the two share helpers
or an aura so it isn't registered twice. The higher set ID gets the `-ID` golden label. The golden `.results` files only
gain an entry per set. TASKS.md has the unmodelled parts.

Stale check: `stale.mjs` already flags most builds because of master's own changes from the comparison session.
I didn't record. No preset wears these sets, so recording after that session deploys is safe.

## Dry run per spec

Options are the pool's entries across slots, counting each random suffix roll. Unmodelled means a Classic tooltip with
"Use:" or "Chance on hit:" and no registered item or weapon effect. Classic and partial mean the stats are hidden
(Classic stats, or only part of Forever's). The last column counts Forever items with no tooltip to check.

| Spec | Items | Options | Unmodelled | Classic stats | Partial stats | Forever, no tooltip |
|---|---|---|---|---|---|---|
| enh | 2384 | 22127 | 73 | 36 | 18 | 609 |
| ele | 2286 | 22025 | 70 | 32 | 18 | 565 |
| ret | 2691 | 26790 | 80 | 30 | 18 | 633 |
| balance | 1720 | 15413 | 60 | 16 | 6 | 462 |
| feral | 1720 | 15413 | 60 | 16 | 6 | 462 |
| hunter_bm | 2195 | 16677 | 71 | 25 | 6 | 504 |
| hunter_mm | 2195 | 16677 | 71 | 25 | 6 | 504 |
| hunter_sv | 2195 | 16677 | 71 | 25 | 6 | 504 |
| mage_arcane | 1275 | 7647 | 55 | 14 | 6 | 337 |
| mage_fire | 1275 | 7647 | 55 | 14 | 6 | 337 |
| mage_frost | 1275 | 7647 | 55 | 14 | 6 | 337 |
| mage_frostfire | 1275 | 7647 | 55 | 14 | 6 | 337 |
| shadow | 1323 | 7275 | 62 | 11 | 12 | 369 |
| smite | 1323 | 7275 | 62 | 11 | 12 | 369 |
| rogue_combat_swords | 1732 | 12096 | 59 | 25 | 6 | 432 |
| rogue_assassination | 1569 | 11071 | 42 | 18 | 6 | 402 |
| rogue_subtlety | 1619 | 11483 | 47 | 22 | 6 | 408 |
| affliction | 1294 | 7385 | 57 | 14 | 6 | 332 |
| demonology | 1294 | 7385 | 57 | 14 | 6 | 332 |
| destruction | 1294 | 7385 | 57 | 14 | 6 | 332 |
| arms | 2662 | 21196 | 84 | 31 | 12 | 602 |
| fury | 2662 | 21196 | 84 | 31 | 12 | 602 |
| fury_dw | 2496 | 20311 | 60 | 29 | 12 | 561 |

98 distinct items have unmodelled effects across all pools. The item lists are in `phase0_pool/<spec>.json`.

## Notes for step 2

- `TestItemSwitches` and `TestWarriorWfGate` (tools/rotopt) fail on clean master too.
  `TestRaidForeverReportHashesStable` fails on HEAD too (paladin__level_60_sub_holy_undead).
- `phase0_prices.py` adds any priced item it hasn't seen to the content check. That item then counts as checked.
