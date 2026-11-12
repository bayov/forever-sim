# Healer mechanics: pass 11 (2026-11-12 UTC, master 55cd92b09 [tankheal], client 1.60.1.70888)

Forced re-audit after the Restoration druid re-tune (`git diff cb0d928fd 55cd92b09`). One new low engine finding (14).
Findings 10 and 11 still open, unchanged.

## Get current

- wago.tools newest build still 1.60.1.70888; no new client rows.
- Blizzard posts: news.blizzard.com returned 500/307 to the fetcher, so nothing newer than the Nov 10 hotfixes
  (`dev_notes_seen`) could be read this pass. `orchestration/update_ledger.md`: no entry since pass 10.
# Healer mechanics, pass 11 (2026-11-12)

Master 8306f1d52 against last pass a0b4d3be9. Result: quiet.

## Get current

| Source | Seen | Result |
|---|---|---|
| Newest client (wago.py latest) | 1.60.1.70888 | Same as pass 10, no verdict reopened |
| Blizzard notes | Nov 10 hotfixes (already seen) | Nothing newer |
| orchestration/update_ledger.md | no commits since a0b4d3be9 | Nothing new |

## What changed (sim/, a0b4d3be9..8306f1d52)

| File | Change | Healer effect | Verdict |
|---|---|---|---|
| `sim/druid/healing.go:740-743` | Nature's Swiftness comment names Wild Growth | none (code since 9628e85b6, pass 9) | OK |
| `orchestration/tools/healers/druid_restoration_apl.py` | NS on its own line: `ns_on` rg/ht/wg, `ns_at` | rotation only | OK |
| `ui/restoration_druid/apls/*.apl.json`, healbench req | raid: NS at Lowest Health < 45% then instant Regrowth 9; tank: < 60% then instant HT 11 | see below | OK |
| talents `05302001--5253005153113051` | Heart of the Wild 2 for Subtlety 2 | +4% Intellect (`talents.go:51-54`, 2%/pt, client tier 1, verdict row kept) | OK |

## Nature's Swiftness as the presets now use it

| Check | Client / design | Sim | Verdict | Tier |
|---|---|---|---|---|
| Spells it reaches | 17116 mask HT, Regrowth, Wrath (+Wild Growth, hotfix 113885) | `healing.go:754-760` | match | 1 |
| Autocast leaves it once the APL casts it | design | `apl.go:165-174` removes castSpell MCDs | OK | - |
| Off GCD, consumed by the next reached cast, 3 min from use | 17116 GCD 0, CD 180 s | log: NS and the instant heal at the same timestamp | match | 1 |
| Raid: NS -> Regrowth 9858 instant, 525 mana, 1.5 s GCD | - | Ragnaros x2.0 log, 20.22 s | OK | - |
| Tank: NS -> HT 25297 instant, 756 mana | - | Onyxia x1.3 log, 16.50 s | OK | - |
| NS fires when the Lowest Health line has no target | - | true health read vs seen-health pick | finding 14 | - |

## Engine

- Finding 14: Current Health % / Current Health of Most Hurt or Lowest Health read the stand-in's true health
  (`apl_values_resources.go:40-47,78-86`) while the pick and Health Missing use seen health (reaction time, heals in
  flight). Emergency lines can fire before the healer sees the hit and then find no target; with the assignment all
  dead they read 0. Patch tested in a scratch copy: carry identical on all 63 req x boss pairs (200 iterations,
  paired seeds); deaths and scares identical at x1.3 (400 iterations, 4 presets x Onyxia, Ragnaros). Low.

## Tests

`go test -tags=with_db ./sim -run TestDruid` (incl. `TestDruidNaturesSwiftnessOnRegrowth` with Wild Growth) and
`./sim/druid/...`: all pass. No test drives the new NS line through the APL (emergency line -> instant heal); add one
with finding 14's engine test (finding 10's list).
| Commit | Files | Healer impact |
|---|---|---|
| c380bbaef Druid: Rake's direct hit adds 6.5% AP | sim/druid/rake.go, feral tests/results | None. Cat form melee ability; the Restoration druid presets never cast Rake, and the change is limited to Rake's hit base damage |
| 368b2e181 Hunter pet melee speed | sim/core/pet.go, sim/core/unit.go, sim/hunter/* | None. Only pets with FollowsOwnerMeleeSpeed (the hunter pet); healers' pets (none) and health, heals, regen untouched |

Steps 3 to 5 (spells, engine, tests): no heal, healing talent, set bonus or healing engine file changed, so
the pass 10 tables and verdicts in ref/healer_mechanics/verdicts.md stand unchanged.

## Findings

No new findings. Open: healer_mechanics-20261104-11 (Riptide mana), healer_mechanics-20261104-10 (untested
heal paths); neither touched by these commits.
