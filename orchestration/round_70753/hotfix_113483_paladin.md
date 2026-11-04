# Hotfix of 2026-11-03 18:29Z: the paladin rows (push 113542)

Source: wago.tools/hotfixes, product wowf, region 91, build 70707, logged 2026-11-03 18:29:17Z. The rogue rework is push
113483 (`hotfix_113483.md`). It has no paladin row. The paladin rows are in **push 113542**, logged in the same second
as 113483 and the rest of that day's pushes. Region 93 has identical copies. Old values come from the 70707 base tables
and client 70753/70755 (`paladin.md`, Phase A). No later push overrides any of these records.

## How it was read

The container had no hotfix cache. The first `import_db2.py 1.60.1.70707+hf paladin` built `+hf` from no rows, which
is just the plain 70707 tables. That gave a false diff: 66 structure and 51 number changes, the whole October 30 rework
undone. The fix was to fetch the hotfix rows and apply them:

    python3 -I tools/sources/hotfixes.py fetch 1.60.1.70707 --set talents --set spelltext   # about 75 min
    python3 -I tools/sources/hotfixes.py apply 1.60.1.70707 --set talents --set spelltext
    python3 tools/forever_talents/import_db2.py 1.60.1.70707+hf <class>

The "70707 spell" search got 14,090 of 14,118 rows (28 lost to the site's paging). Every Trait and CurvePoint search
was complete. After the fetch, paladin shows **0 structure, 3 numbers, 3 wording**:

- numbers: Vengeance ranks 1-3, `1/30/3 -> 1/30/5`, `2/30/3 -> 2/30/5`, `3/30/3 -> 3/30/5` (the stack count)
- wording: Two-Handed Weapon Specialization ranks 1-3, "physical damage" becomes "Physical and Holy damage"

The brief named Oath of Light (1328548). No hotfix after push 113139 (10-30) has a row for it, so its numbers are the
10-30 values the client already has. The brief's "4 numbers" was probably counted against the empty `+hf`.

## Push 113542, every row

| Table | Record | Field | Old | New |
|---|---|---|---|---|
| SpellAuraOptions | 247421 (Vengeance proc 20050) | CumulativeAura | 3 | **5** (DifficultyID 0, ProcChance 101, ProcCharges 0 unchanged) |
| SpellEffect | 697263 (20111 Two-Handed Weapon Specialization, effect 0, aura 79) | EffectMiscValue_0 (school mask) | 1 (Physical) | **3** (Physical + Holy); base points 2 unchanged |
| Spell | 20111 | Description | "Increases the physical damage you deal by $s1%, ..." | "Increases the Physical and Holy damage you deal by $s1%, ..." |
| SpellName | 20111 | Name | Two-Handed Weapon Specialization | same (re-sent) |
| CurvePoint | 331369, 331370 (curve 110350) | | 1 -> 12, 2 -> 24 | same (re-sent) |
| CurvePoint | 371197-371199 (curve 110350) | | 3 -> 36, 4 -> 48, 5 -> 60 | deleted, re-added as 372118-372120 with the same values |
| SpellScript | 130401 | | | status 3 (invalid) |

Curve 110350 belongs to TraitDefinition 142634, Champion of the Light (1311084). Its values stay 12/24/36/48/60.

## Ledger (update_ledger.md, PAL-HF-01 to 04)

| ID | Spell | Old (client 70755) | New (hotfix 113542) | Sim before | Sim now | Citation in code |
|---|---|---|---|---|---|---|
| PAL-HF-01 | Vengeance 20049 / 20050 | stacks up to 3 | stacks up to 5 | 5, a notes-only value (round 70753) | 5, now from the hotfix | `sim/paladin/talents.go` `applyVengeance`; `TestHotfix113542` |
| PAL-HF-02 | Two-Handed Weapon Specialization 20111 | Physical only (misc 1) | Physical and Holy (misc 3), 2/4/6% | Holy school too, a notes-only value | the same, now from the hotfix | `talents.go` `applyWeaponSpecialization`; `TestHotfix113542`, `TestJudgementOfRighteousnessMelee` |
| PAL-HF-03 | 20111 tooltip | "physical damage" | "Physical and Holy damage" | old text | hotfix text | `tools/forever_talents/client/paladin.json` (`import_db2.py 1.60.1.70707+hf paladin --write`), `data/paladin.json` (`import_client.py`), `ui/core/talents/trees/paladin.json` (`import_talents.py --write`) |
| PAL-HF-04 | Champion of the Light curve 110350 | 12-60% | 12-60% (re-keyed) | 12-60% | unchanged | none |

Both values were already in the sim under Forever, gated and tested, from round 70753's notes-only values (standards
37). The hotfix now carries them, so it is the source (tier: a server hotfix beats the notes and the client). Only the
citations change: the comments in `talents.go` and the test renamed from `TestNotesOnlyValues70753` to
`TestHotfix113542`. The paladin.md warning ("if a hotfix later sets Vengeance or 2H Spec to the client's values,
revert") no longer applies: the hotfix set the notes' values.

The talent import also regenerates `notSimulated` in the UI tree from the sim code. Holy Power gets it, since the sim
doesn't read it. Stoicism and Unsealed Wrath lose it, since the code reads them. Those flags are UI only and have
nothing to do with the hotfix.

## Other classes (import_db2.py 1.60.1.70707+hf, the client files as they are now)

| Class | Diffs | Which 11-03 push | In the sim? |
|---|---|---|---|
| Druid | none | none | n/a |
| Priest | none | none | n/a |
| Shaman | none | none | n/a |
| Rogue | none (the tree was written from +hf in 316a600) | 113483; also 113351 Poised (re-sent, already in client 70753) | yes, nothing left unwritten |
| Warlock | none; notes entries Wrack and Shadow and Flame "can go" (the client+hf now reads the same) | 113527 (Wrack), 113537 / 113351 (Spellstone tooltips) | yes (notes values = hotfix) |
| Mage | 2 wording: Missile Barrage drops Frost Blast from its procs | 113455 | yes: `sim/mage/talents.go` `applyMissileBarrage` has no Frost Blast case. 0 DPS |
| Hunter | 10 numbers: Trueshot Aura 15 -> 20, Barrage 3/7/10 -> 5/10/15, Sniper Shot 160 -> 320, Lightning Reflexes 2 -> 3% a rank | 113534 | yes: HUN-HF-04 to 10, landed from foreverchanges.pro's copy of the same hotfix. `client/hunter.json` isn't rewritten from +hf |
| Warrior | 6 structure, 1 number, 7 wording: Raging Blows removed, Reckless Swings (node 113636) and Improved Whirlwind (105977) now in the tree, Reckless Swings / Flurry / Improved Whirlwind columns moved, Spearing Strike 40/80 -> 75/75, Booming Voice now covers Howls; push 113539 also changes Slam, Sweeping Strikes, Deep Wounds, Concussion Blow and Improved Slam | 113539 | **no**: no ledger row names push 113539; `sim/warrior/spearing_strike.go` is still 40% weapon damage plus 80%; the notes entries for Reckless Swings and Improved Whirlwind "can go" |

Other pushes of the same second are not class talents: 113548 (Death Pact, a Baron Rivendare spell), 113574 (Holy
Strike tooltips, a trailing space only), 113537 (Counterspell - Silenced, Winter's Might text) and 113498 (curves).

## DPS

The values didn't change, so no number can move. Checked: the CLI built before and after the edit
(`go build -tags=with_db ./cmd/wowsimcli`). Each request in `tools/dps_report/req/paladin__*.json` ran at 120 / 180 /
300 s, 20,000 iterations, seed 1. The two builds give identical results:

| Build | 120 / 180 / 300 s (before = after) |
|---|---|
| paladin__level_60 | 1095.5 / 1092.1 / 1080.3 |
| paladin__level_60-duty-crusader | 1078.0 / 1082.9 / 1075.9 |
| paladin__level_60-duty-wisdom | 1069.6 / 1075.3 / 1069.1 |
| paladin__level_60_shock_ret | 1037.5 / 1030.2 / 1025.5 |
| paladin__level_60_shock_ret-duty-crusader | 1024.5 / 1029.3 / 1027.9 |
| paladin__level_60_shock_ret-duty-wisdom | 1015.5 / 1020.0 / 1018.7 |
| paladin__level_60_sub_holy | 1093.3 / 1091.2 / 1085.8 |
| paladin__level_60_sub_holy-duty-crusader | 1076.9 / 1082.8 / 1083.3 |
| paladin__level_60_sub_holy-duty-wisdom | 1069.2 / 1075.0 / 1075.6 |
| paladin__level_60_sub_holy_undead | 1222.9 / 1221.4 / 1214.1 |

The 180 s numbers equal the published ones (paladin.md). No talent value changed, so no talent pick needs another
look. The round 70753 searches already ran with Vengeance at 5 and 2H Spec on Holy.

Tests: `go test --tags=with_db ./sim/paladin/...` passes. `go test --tags=with_db ./sim/... -run
TestRaidForeverReportHashesStable` passes in `sim` (only `sim/web` fails to build, for its missing `binary_dist`
package). No paladin hash moves.

## Left for others

- Warrior push 113539 has not been written into the sim (see the table). It needs a warrior session like this one.
- `pal_2h_spec_holy` (`tools/dps_report/assumptions_def.py`, `orchestration/spec_assumptions.md:251`) and
  `forever-wiki/classes/paladin.md` still call the two values notes-only. They can now cite hotfix 113542. The
  orchestrator owns those files.
