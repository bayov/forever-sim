# Healer workers' brief (H3, 2026-11-03)

The user, 2026-11-03, to the healing session: implement every healer spec, make sure every healing ability
and talent works, ship optimized presets, and run auditors like the DPS specs' (the audit loop,
`orchestration/audit/`). "Keep going without waiting for my calls ... save all calls you need [me] to make to
the very end." The tank specs are the tank session's, not ours.

The healing session (session_01TXnjyPVS2mPceSNTempQD9) orchestrates. One worker owns each class:

| Worker | Spec (proto) | Presets | Package |
|---|---|---|---|
| paladin | Holy paladin (`HolyPaladin`, `PaladinOptions`) | tank healer, raid healer | `sim/paladin/holy` (new) |
| druid | Restoration druid (`RestorationDruid`) | tank healer (HoTs), raid healer | `sim/druid/restoration` (the old `_restoration` is wowsims', never built) |
| priest | Holy and Discipline priest (`HealingPriest`) | Holy raid, Holy tank, Discipline tank (shields) | `sim/priest/healing` (old, unregistered) |

The Restoration shaman (H2) is done and is the model. Copy what it did.

## Read first

- `docs/healing_sim.md`: the user's rulings (what a healer is scored on, assignments, bound ranks, the
  score), the engine (health, heal targets, heals in flight, Expected Healing, cancelling), H2, and the raid
  damage from the logs. `orchestration/handoff_healing.md` for the gotchas.
- The shaman: `sim/shaman/healing.go` (ranks, coefficients, `ExpectedHealing`, Chain Heal's jumps, Riptide),
  `sim/shaman/restoration/restoration.go` (the agent), `ui/restoration_shaman/` (presets, two role APLs, the
  synthetic gear set), `tools/healbench/req/shaman_restoration*.json`, `sim/raid_healing_test.go` (the healing
  fight tests and their helpers), `tools/database/forever_synthetic_items_shaman_restoration.go` and
  `orchestration/tools/healers/` (the gear scripts).
- Your class's DPS brief (`orchestration/brief_<class>.md`) and audit notes for how the DPS classes were
  audited against Forever's client: sources, tiers, the numbers and what changed. `orchestration/standards.md`
  (it overrides older docs; 1b, 5, 5a and 35 are the gear rules), `orchestration/forever_class_changes.md`,
  `docs/data_sources.md` (source order), `notes.md`.

## The work, in order

1. **Rules audit of every healing spell and talent** against Forever's newest client
   (`tools/sources/wago.py`: SpellEffect, SpellMisc, SpellLevels, SpellPower, SpellDuration, SpellCooldowns,
   SpellCastTimes, Talent and friends) and the patch notes, including the October 30 launch-realm hotfix
   (paladin and druid reworks, `orchestration/update_ledger.md`). Ranks and their levels, base heal, cost,
   cast time, coefficient (and the downranking rule the class already uses), cooldowns, HoT ticks, talents'
   values. Write it down as the DPS classes did (a table per spell: client value, sim value, source). Fix
   what's wrong. Never guess a value: tier your sources, and list anything you can't confirm for the end
   (see "Calls for the user").
2. **Implement the heals and healing talents that are missing**, and register the spec's agent
   (`sim/register_all.go`; a caster that doesn't melee, like the shaman's). Every heal with a cast time
   needs `SpellConfig.ExpectedHealing` (heals in flight and Expected Healing read it). Heals go through
   `CalcAndDealHealing` / HoTs through the dot pipeline as the shaman's do, crit with `OutcomeHealingCrit`.
   APL targets: `Tank`, `MostHurt`, `LowestHealth` (`sim/core/heal_targets.go`).
3. **Tests: every healing ability and talent works.** A test per heal (it heals its target, costs its
   mana, its cast time and cooldown are right, its HoT ticks the right number of times) and per talent (on
   against off moves exactly what it should: a 10% heal talent heals 10% more, a cost talent costs less, a
   proc procs). Put them in your package (`sim/<class>/<spec>/..._test.go`) on small healing fights like
   `sim/raid_healing_test.go`'s, or in `sim/raid_healing_test.go` style in your own file under `sim/`
   (`sim/raid_healing_<class>_test.go`). A test that would pass with the mechanic broken is no test.
4. **Synthetic gear**: a healer row for your spec in `orchestration/gear_targets.md` (Healers section)
   from your class's two Wowhead lists, priced the shaman's way, and a set on it
   (`tools/database/forever_synthetic_items_<class>_<spec>.go`, a free id range: paladin 990300-990399,
   druid 990400-990499, priest 990500-990599, armor of your class, the trinkets your class's DPS sets use
   if they're spell power, or two healer trinkets on the same 1b trinket rules). The guide pages are already
   in the shared Wowhead cache (`~/.cache/wow-data/wowhead/pages`, read with
   `tools/database/forever_wowhead/wowhead.py`'s `Wowhead().page(path)`, which won't fetch them again):
   - paladin: `forever/guide/classes/paladin/holy/bis-gear`, `classic/guide/paladin-healing-pre-raid-best-in-slot-bis-gear-wow-classic`
   - druid: `forever/guide/classes/druid/restoration/bis-gear`, `classic/guide/druid-healing-pre-raid-best-in-slot-bis-gear-wow-classic`
   - priest: `forever/guide/classes/priest/holy/bis-gear`, `forever/guide/classes/priest/discipline/bis-gear`, `classic/guide/priest-healing-pre-raid-best-in-slot-bis-gear-wow-classic` (one priest set unless Holy and Discipline really differ)

   **Never fetch other Wowhead pages from a script**: three workers share one address and Wowhead blocks
   bursts. If you need one more page, ask the orchestrator. Then
   `go run -tags=with_db ./tools/database/gen_db -outDir=assets -gen=db`. Only your items may change in
   `assets/database/db.*`: if anything else changes, stop and tell the orchestrator.
5. **Presets, optimized**: a tank healer and a raid healer rotation (the priest a Discipline one too),
   tuned on `tools/healbench` as the shaman's were: bound ranks (at most 4 a heal, the user's ruling),
   mana tools (Divine Favor and Illumination, Innervate on oneself, Inner Focus and Shadowfiend if Forever has
   it, mana potions and runes through `autocastOtherCooldowns`), thresholds compared at the files' own lengths
   and at 180 and 300 s. Talents: compare the builds that matter (deep Holy against hybrids, the 31-point
   talents) on the bench; `tools/rotopt` scores DPS only. Then `tools/healbench/req/<class>_<spec>.json` (and
   `_raid`), `ui/<spec dir>/apls/*.apl.json`, `presets.ts` and `sim.ts` like the shaman's, UI type checks.
   The UI spec page stays unlaunched (`ui/core/launched_sims.ts`): launching is the user's call.
6. **Numbers**: healbench on every boss file (`tools/encounters/*.json`, the Anniversary set) in the tank
   and raid roles, 1,000 iterations, against the shaman's presets as the reference. Write them into your
   class's doc.

## Rules

- **Your worktree, your branch.** Commit there with clear messages ending with the session's attribution
  lines (below). Never push, never touch master; the orchestrator merges. Run `git fetch origin` and merge
  `origin/master` into your branch now and then.
- **Shared files.** Keep to your class's files. Shared core (`sim/core/**`, `proto/api.proto`,
  `proto/common.proto`, `proto/apl.proto`, `tools/healbench/main.go`, `docs/healing_sim.md`) only where
  your work can't be done otherwise, as small as you can, and say so in your report. The priest owns the
  absorb pipeline in core (Power Word: Shield, Divine Aegis; the open default: an absorb counts as
  effective healing while it absorbs and its unused part as overheal). New proto fields: check
  `origin/master` first and take the next free number in your class's own message.
- **Docs**: your own `docs/healing_<class>.md` (audit tables, what you fixed, presets and their numbers).
  The orchestrator folds a summary into `docs/healing_sim.md`.
- **Tests and checks before every commit**: `go build -tags=with_db ./...` (skip `sim/web`),
  `go test -tags=with_db ./sim/<class>/... ./sim/core/...`, `go test -tags=with_db -run TestHealingFight ./sim/`
  and your own tests. The full `go test ./sim/` takes about 2.5 minutes: run it before you report done.
  `npx tsc --noEmit -p .` after UI changes. `sim/testdata/report_hashes.txt` must not move unless a DPS
  number really should (then say why).
- **Protos**: after a `.proto` change, `PATH=$PATH:~/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`
  and `make ui/core/proto/api.ts`.
- **Calls for the user**: anything only the user can decide (a value no source confirms, a mechanic with
  two readings, a rule to change) goes in a "Calls for the user" section of your doc with your default and
  the alternatives' measured effect. Use your default and keep going; never stop to ask.
- Commits end with:

      Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
      Claude-Session: https://claude.ai/code/session_01TXnjyPVS2mPceSNTempQD9

  and never name a model anywhere else.

## Report

When you're done (or truly blocked), reply with: your branch and last commit, what you fixed and built (a
line each), the presets' numbers against the shaman's, the shared files you touched, and your calls for the
user. Keep it under 40 lines; the details are in your doc.
