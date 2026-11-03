# Tank workers' brief (T1, 2026-11-03)

The user, 2026-11-03, to the tank session: "keep going and implement all tank specs, then run similar auditors
to the ones we run for dps specs (see the audit routine for example). Ensure every tank ability and talent works
(we checked dps, but also need to ensure tanking ones). Keep going without waiting for my calls until we have all
tanking specs working with optimized presets, and save all calls you need [me] to make to the very end." The
healer specs are the healing session's (`orchestration/brief_healers.md`), not ours.

The tank session (session_01TXnjyP5nUQs8zySrhUvYoi) orchestrates. One worker owns each tank, and one the gear
standard:

| Worker | Spec (proto) | Package | UI | Bench request today | Synthetic item ids |
|---|---|---|---|---|---|
| warrior | Protection warrior (`TankWarrior`) | `sim/warrior/tank_warrior` | `ui/tank_warrior` | `warrior_protection.json` | 990600-990699 |
| paladin | Protection paladin (`ProtectionPaladin`) | `sim/paladin/protection` | `ui/protection_paladin` | `paladin_protection_valor.json` | 990700-990799 |
| druid | Feral tank (`FeralTankDruid`) | `sim/druid/tank` (new, `bear.go`, `bear_attacks.go`) | `ui/feral_tank_druid` | `druid_bear.json` | 990800-990899 |
| shaman | Warden shaman (`WardenShaman`) | `sim/shaman/warden` | `ui/warden_shaman` | none yet | 999000-999099 |
| gear | the tank gear standard (standards 36, the Tanks rows of `orchestration/gear_targets.md`) | - | - | - | - |

## Read first

- `docs/tank_sim.md`: the user's rulings (threat against survival as weights, the three presets, healing needed
  as the survival score, Forever's boss crit, a tank applies its own debuffs), the tank sheet, the bench, the
  objective and the paired noise test, the encounters, and the rules audit so far (warrior, paladin, bear).
- The tools: `tools/tankbench` (`-encounter`, `-actions`, `-boss-stats`), `tools/tankscore` (`MakeTank`,
  `OwnDebuffs`, `Objective`, `PairedDiff`), `tools/rotopt` (`-tank max-threat|balanced|max-survival`,
  `-tank-encounter a.json,b.json` scores the mean over several encounters), `tools/encounters/README.md`.
- Your class's DPS brief (`orchestration/brief_<class>.md`), its forever-wiki page
  (`forever-wiki/classes/<class>.md`), `orchestration/standards.md` (it overrides older docs; 1b, 2, 4, 5, 5a,
  6, 7, 8a are the gear rules, and 36 will be the tank one), `orchestration/forever_class_changes.md`,
  `docs/data_sources.md` (source order), `notes.md`.
- wowsims/forever (github.com/wowsims/forever) is a sim of Forever with in-game measurements this repo cites
  (their sim/<class>/, commit messages with logs). Clone it read-only into your scratch space if you need it;
  never push, comment or open anything there.

## The scoring setup (the orchestrator's defaults, calls for the user at the end)

- **Encounters:** the Anniversary set (`tools/encounters/*.json`) without Ragnaros: Magmadar, Golemagg, Onyxia,
  Broodlord, Vaelastrasz, Chromaggus, scored as their mean
  (`-tank-encounter tools/encounters/magmadar.json,...`). Ragnaros is benched, not scored: its tanks wore fire
  resistance, which no set models, and without it Ragnaros alone outweighs the rest.
- **Presets:** three a spec, Max threat (w = 1/3), Balanced (1) and Max survival (3) (the user's ruling).
  Balanced is the spec's default.
- **Objective settings:** rotopt's defaults (2 threat members, a 900 DPS slot, 5% deaths), as `docs/tank_sim.md`.
- **Gear:** every tank wears a synthetic set by the tank gear standard (below), not real items.

## The work, in order (class workers)

1. **Rules audit of every tank ability and talent** against Forever's newest client (`tools/sources/wago.py`:
   SpellEffect, SpellMisc, SpellLevels, SpellPower, SpellDuration, SpellCooldowns, SpellCastTimes, the talent
   rows and their trait curves; `latest`, then `fetch <build>+hf` for the hotfixed rows) and the patch notes,
   including the October 30 launch-realm hotfix (`orchestration/update_ledger.md`). Every ability a tank of
   your class uses or could use (threat, mitigation, avoidance, self-heals, taunts, cooldowns, its own debuffs),
   every talent a tank build can reach, and the class's defensive passives (stances, forms, auras, shields,
   imbues). Ranks and levels, values, costs, cooldowns, durations, threat (flat threat, multipliers: the client
   rarely states them, so tier your sources: wowsims/forever's measurements, logs, then Classic's and say so).
   Write it as a table per ability or talent (client value, sim value, source and tier, verdict) in your own
   `docs/tank_<class>.md`. Fix what's wrong. Never guess a value: list what no source confirms for the end
   (Calls for the user). The rules audit's first pass is in `docs/tank_sim.md` (Rules audit): build on it.
2. **Implement what's missing** that a tank uses or that its score can see. Health-based buttons (Shield Wall,
   Last Stand, Templar's Bulwark, Frenzied Regeneration) can't be scored by the replay when they react to
   health (a scoring run reads none): implement them for the UI and the healing sim, keep them out of scoring
   rotations unless they're used on a timer that doesn't read health, and say what that costs the class.
3. **Tests: every tank ability and talent works.** A test per ability (it does its damage or effect, costs its
   resource, its cooldown and duration are right, its threat is what the source says) and per talent (on
   against off moves exactly what it should: 5% dodge is 5% more dodge on the tank sheet, a threat talent's
   threat, a cost talent's cost, a proc procs at its rate). In your package (`sim/<class>/<spec>/..._test.go`),
   on small requests like `TestForeverTankScore` (`sim/warrior/tank_warrior/forever_test.go`) and
   `TestForeverBear`. A test that would pass with the mechanic broken is no test. The DPS specs' tests stay
   green.
4. **Rotation**: a rotopt template for your tank (`tools/rotopt/<class>_tank.go`, knobs for what a tank
   decides: what to spend Rage or mana on, Heroic Strike / Maul thresholds, which debuffs and when, openers
   such as Charge with Vanguard, the stand-in healer's timing doesn't read health), searched with `-tank` on
   the scoring encounters. Your own debuffs come from you (`tankscore.OwnDebuffs`: extend it for your class if
   you keep a raid debuff up yourself, and say which).
5. **Synthetic gear** once the gear worker's standard is on master (`git fetch origin` and look for standards
   36 and the Tanks rows in `orchestration/gear_targets.md`; do steps 1-4 first): a set at your row
   (`tools/database/forever_synthetic_items_<class>_tank.go`, your id range, armor of your class, a shield where
   your class tanks with one, weapon as standards 1b and 22, two trinkets by 1b's trinket rules with the
   standard's tank trinket), enchants as standards 7, then
   `go run -tags=with_db ./tools/database/gen_db -outDir=assets -gen=db`. Only your items may change in
   `assets/database/db.*`: if anything else changes, stop and tell the orchestrator in your report.
6. **Presets, optimized**: a talent search and a knob search at each of the three weights on the scoring
   encounters (start from your class's best known build; rotopt `-talent-search`, `-knobs`; 600 iterations a
   candidate and 3,000 to confirm, more where the paired noise says so), races checked (standards: the race the
   spec sims best with, by the same objective). Then `tools/tankbench/req/<class>_<spec>_<preset>.json` (the
   balanced one replaces today's request), `ui/<spec>/apls/*.apl.json`, `gear_sets`, `presets.ts` and `sim.ts`
   with the three presets, and `npx tsc --noEmit -p .`. The UI spec page's launch state stays as it is.
7. **Numbers**: tankbench on all seven Anniversary encounters (Ragnaros too) for your three presets, 1,000
   iterations: TPS, SafeDPS, DTPS, healing needed, death floor, the score at each weight. Write them into your
   doc, next to the other tanks' current numbers.

## The gear worker

1. **Tank stat prices** (standards 36, a default the user rules on at the end): Defense, dodge, parry, block
   chance, block value, bonus armor and the resistances aren't priced in 1b. Price them the way 1b priced the
   rest: from Era level 60 epics (`orchestration/bis_budget_rules.md`'s method and data;
   `assets/database/db.json` has the Era items, and Classic's item budget costs are a cross-check), so that a
   tank piece carries the same budget as a DPS piece of its slot and item level. Say what each stat costs, the
   fit and its error, and the alternatives' effect on a tank set.
2. **Tank stat mixes** (the Tanks section of `orchestration/gear_targets.md`): a row per tank (Protection
   warrior, Protection paladin, Feral tank, Warden shaman) from real level 60 tank gear, the way the DPS rows
   came from wowsims/sod's BiS sets (`orchestration/tools/gear_targets_v3.py`): wowsims/sod's tank presets
   where it has them (warrior, bear, paladin, shaman tank), Wowhead's Forever and Classic pre-raid tank BiS
   lists otherwise. Wowhead: only through `tools/database/forever_wowhead/wowhead.py` (the scrape-wowhead
   skill), a handful of pages, slowly; you're the only worker who may fetch them.
3. Commit on your branch with the standard and the rows; the orchestrator merges it to master first, and
   tells the class workers.

## Rules

- **Your worktree, your branch.** Commit there with clear messages ending with the attribution lines below.
  Never push, never touch master; the orchestrator merges. `git fetch origin` and merge `origin/master` into your
  branch now and then (the gear standard lands there).
- **Shared files.** Keep to your class's files. Shared core (`sim/core/**`, `proto/**`, `tools/tankscore`,
  `tools/tankbench`, `tools/rotopt/*.go` other than your template, `docs/tank_sim.md`, `orchestration/standards.md`)
  only where your work can't be done otherwise, as small as you can, and say so in your report. New proto
  fields: check `origin/master` first and take the next free number in your class's own message.
- **Shared class files.** The healing session's paladin and druid workers (and its shaman, done) edit the same
  class packages (`sim/paladin/*.go`, `sim/druid/*.go`, `sim/shaman/*.go`, the class talent trees). Keep your edits
  there small and to the point, merge `origin/master` often, and resolve conflicts keeping both sides' work.
- **Docs**: your own `docs/tank_<class>.md` (audit tables, what you fixed and built, presets and their numbers,
  your calls). The orchestrator folds a summary into `docs/tank_sim.md`.
- **Tests and checks before every commit**: `go build -tags=with_db ./...` (skip `sim/web`),
  `go test -tags=with_db ./sim/<class>/... ./sim/core/...` and your own tests. Before you report done, the whole
  `go test -tags=with_db -timeout 40m` over `./sim/...` without `sim/practice` and `sim/web`.
  `npx tsc --noEmit -p .` after UI changes. `sim/testdata/report_hashes.txt` must not move unless a DPS number
  really should (then say why).
- **Protos**: after a `.proto` change, `PATH=$PATH:~/go/bin npx protoc -I=./proto --go_out=./sim/core ./proto/*.proto`
  and `make ui/core/proto/api.ts`.
- **Calls for the user**: anything only the user can decide (a value no source confirms, a mechanic with two
  readings, a rule to change) goes in a "Calls for the user" section of your doc with your default and the
  alternatives' measured effect. Use your default and keep going; never stop to ask.
- Commits end with:

      Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
      Claude-Session: https://claude.ai/code/session_01TXnjyP5nUQs8zySrhUvYoi

  and never name a model anywhere else.

## Report

When you're done (or truly blocked), reply with: your branch and last commit, what you fixed and built (a line
each), the presets' numbers, the shared files you touched, and your calls for the user. Under 40 lines; the
details are in your doc.
