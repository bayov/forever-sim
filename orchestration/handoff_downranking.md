# Downranking in the engine (handoff, 2026-11-04)

From the shaman round 70753 session (session_01TXnjyRtqzkPgGAGjRWBDxS). The task is to move Forever's
downranking penalty out of the class packages and into `sim/core`, so every class gets it from one place
under one rule.

## The rule, and what's open

- **The notes:** Blizzard's beta development notes (2026-10-01, us.forums.blizzard.com/en/wow/t/2360696/4)
  say ranks "vastly below your current level receive less benefit from your Spell Damage and Spell Healing
  and have a reduced chance to activate your Class abilities and Talents". They give no numbers.
- **The size:** foreverchanges.pro/downranking measured it from beta heals at level 30 and below. A rank
  keeps `1 − 0.05 × (caster level − learn level − 17)` of its spell power coefficient, clamped to [0, 1].
- **Open, the user's ruling:** the 2026-10-10 beta damage log in
  `sim/shaman/enhancement/spell_power_test.go` (`TestOrcShamanSpellPowerBeta`) shows the penalty does not
  hit the shaman shocks.
  - **The log:** level 30, 91 spell power. Earth Shock rank 1 (learned at 4) and Flame Shock rank 1
    (learned at 10) hit at their full coefficient.
  - **What the formula would give:** 55% and 85%. Earth Shock would hit for 39–41 against the log's 55–57.
  - **The choice:** heals only, or keep as is. "Keep as is" means Lightning Bolt, Chain Lightning and the
    priest nukes keep the assumed penalty, and the shocks don't take it. See `notes.md`, "Forever's
    downranking penalty".
  - **What to do with it:** make the engine take the scope as a setting (which spell kinds get the penalty)
    so the ruling is a one-line change. Default to today's behaviour, so the refactor moves no number. Ask
    the user for the ruling before changing the default.

## Where it lives today

- **`downrankCoefficient`, defined twice:** `sim/priest/priest.go:160` and `sim/shaman/talents.go:185`.
  The formula is the same in both, and both are a no-op outside Forever (`Env.IsForever()`).
- **Priest:** Smite, Holy Fire, Mind Blast, Shadow Word: Pain and Devouring Plague.
- **Shaman:** Lightning Bolt, Chain Lightning, and the heals in `sim/shaman/healing.go:135`.
- **Without it:**
  - shaman shocks, totems (Searing, Magma, Healing Stream), Fire Nova and Lightning Shield;
  - every other class's spells (mage, warlock, druid, paladin, hunter), including druid and paladin heals.
- **Not modelled anywhere:** the "reduced chance to activate Class abilities and Talents" half of the notes.

## Design sketch

- **Inputs:** `core.SpellConfig` already carries `Rank` and `RequiredLevel` (`sim/core/spell.go:29-30`,
  copied at line 313), and the classes set `RequiredLevel` to each rank's learn level.
- **Hook:** apply the factor at spell registration on Forever, for ranked spells (`Rank > 0`). It scales
  `BonusCoefficient`, the DoT's `BonusCoefficient` and the HoT's.
- **Order:** the engine has to read `RequiredLevel` before a class edits it, and must not double-apply.
  Delete the two class functions in the same change, or the priest and shaman spells get the factor twice.
- **What's left out:**
  - **Opt-out per spell:** for spells the scope leaves out, or whose `RequiredLevel` isn't the learn level.
    Check the weapon imbues (Flametongue and Frostbrand use a 0.1 coefficient), procs, item effects and
    talent-granted spells.
  - **Rank-less spells:** they keep their coefficient.
- **Learn level vs required level:** confirm the classes set `RequiredLevel` to the learn level. A spell
  with an empty or different `RequiredLevel` would get the wrong factor.

## Done when

- **No number moves at the default scope:** `TestRaidForeverReportHashes` and every class golden stay
  byte-identical. If a hash moves, that's a spell gaining or losing the penalty; find out why before going
  on.
- **One table-driven test** in `sim/core`:
  - the formula at the edges (17 levels past the learn level keeps 1, 18 keeps 0.95, the clamp at 0);
  - Classic unchanged;
  - DoTs and HoTs scaled;
  - the opt-out.
- **Tests moved, not lost:** the per-class downrank tests (`sim/shaman/enhancement/downrank_test.go`,
  `sim/priest/shadow/forever_test.go`) still pass, or move into core.
- **Updated text:** the comments that describe the rule in `lightning_bolt.go`, the shock files and
  `priest.go`, plus the downranking lines in `notes.md` and `forever-wiki/classes/shaman.md`.

## If the user rules "heals only"

- **What changes:** the default flips, and the Elemental, Shadow and Smite numbers move.
- **What to redo:**
  - the low-mana fallback ranks in `tools/rotopt/shaman_presets.sh` (Elemental `lightningBoltDownrank`)
    and the priest rotations, with rotopt;
  - the report requests (`mksettings`) and goldens;
  - the report hashes (`RAID_FOREVER_HASHES=$PWD/sim/testdata/report_hashes.txt`, absolute path, run in
    `sim/`).
- **Who owns what:** the orchestrator owns `ui/dps_report/data`, `tools/dps_report/raw` and the deploy.

## Gotchas

- **Stale generated code:** regenerate the Go protos after a merge
  (`protoc -I=./proto --go_out=./sim/core ./proto/*.proto`), or other rounds' talent changes break the build.
- **`sim/web` fails to set up** without `binary_dist`. That's unrelated, so ignore it.
- **No level 60 evidence:** no log yet shows a downranked damage spell at 60. `/log-audit` on a downranked
  Lightning Bolt or Smite would settle the scope.
