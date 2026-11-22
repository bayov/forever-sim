# Real Phase 0 gear for every DPS spec, step 1: tooling and sets, 2026-11-22

The user wants real, non-synthetic Phase 0 gear (what is obtainable in WoW Forever before the raids open on 2026-12-09) for
every DPS spec, starting from the Enhancement-only Phase 0 BiS pipeline. DPS only: no healers, no tanks (user, 2026-11-22).
The synthetic gear stays the report's main gear (standards 1b and 8a). Phase 0 becomes an extra variant later (step 4).

The readiness check found (origin/master 783636b06):
- `tools/rotopt/bis_phase0.sh` is hardwired to Enhancement: the request (:163), talents (:150), start gear (:181),
  `-gear-rank-spec enh` and `-two-hand-only` (:250-251), output paths (:158, :165, :273).
- `tools/rotopt/gear.go`:
  - `weaponTypesFor` (:105) has no hunter, druid, mage, priest or warlock entry;
  - `rangedTypesFor` (:113) lacks bows, guns, crossbows, wands and idols;
  - `usefulSuffix` (:189) keeps caster suffixes only for shaman and paladin.
- `phase0_content.json` holds only the 374 items Enhancement's search met (1 plate item), and an unlisted item counts as
  open (:89). `phase0_enchants.json` (59) is mostly melee.
- The 8 new Dungeon Set 1 sets from #134 aren't wired: 1667, 1669, 1671, 1673, 1675, 1676, 1681, 1778 (TASKS.md:479-483).
- 87 Classic items in the pools have unmodelled use or proc effects, mostly weapons. 770 Forever-only items have no
  Forever tooltip to check.

## The whole plan

1. Tooling and sets (this brief).
2. Forever tooltips for the pool items, model the effects of likely picks, and hand-check availability by armor type
   (cloth, leather, mail, plate in parallel).
3. Run every DPS spec's search in parallel.
4. A "Phase 0" variant of each report build behind a Ranking toggle, like the duty presets (`build_data.py:394`).

## Step 1: what to do

1. Make `bis_phase0.sh` take the spec. Each DPS spec in the report gets its request, talents, start gear (empty slots
   are fine), rank spec and weapon rules (two-hander only, dual wield, one-hand plus off-hand, ranged slot) from one table
   in the repo, not from flags typed by hand. Enhancement must produce the same list as today: check it.
2. Fill the `gear.go` tables for every class: weapon types, ranged types (bows, guns, crossbows, thrown, wands, idols,
   librams, totems) and caster suffixes for mage, warlock, priest and druid. Add tests.
3. Make an unlisted item fail loudly instead of counting as open. Have the script print every picked or top-ranked item
   that `phase0_content.json` doesn't list, so step 2 knows what to check. Do the same for enchants.
4. Wire the 8 new Dungeon Set 1 sets' bonuses from the client data (wago ItemSetSpell / SpellEffect for the newest
   build, assets/db_inputs). Add a test per set. Where a bonus can't be modelled, say so in a comment and in TASKS.md.
5. Add a dry run that lists, per spec, the pool size, the items with unmodelled effects (the readiness scan's method:
   `core.HasItemEffect` / `HasWeaponEffect` against "Use:" and "Chance on hit:" tooltips) and the items with Classic or
   partial stats. That list is step 2's input. Commit its output under `orchestration/round_notes_2026-11-11/`.

Don't run the full searches yet (step 3).

## Coordination

- The wowsims/forever comparison session (`session_01TXnjyQoHMJg2CjHPYTGUMX`) is re-simming 16 warrior and shaman
  builds and deploying the report, then landing a docs commit. It doesn't touch item effects, set bonuses, gear sets or
  presets. Merge origin/master often. Don't publish or deploy the report.
- Set bonus code may flag report builds stale. Run `node tools/dps_report/stale.mjs`. Where a change moves nothing
  (no preset wears these sets), record it with `stale.mjs --record`, after the comparison session's deploy has landed.

## Rules

- Never force-push. End every commit message with your session prompt's trailer lines.
- Keep model names out of files. Keep long runs in subagents or in the background.
- Leave bayovui, the leveling weights and the leveling rotations alone.
- Merge to master and push without a PR, then push to your own development branch too.
- User style in anything they read: concise, no em dash or semicolon, "180s" not "180 s".

## When done

Write `orchestration/round_notes_2026-11-11/phase0_gear_1122.md` with:
- what changed;
- the Enhancement check;
- the 8 sets and what each bonus does;
- the dry run's per-spec numbers.

Push it, send one message to `@parent` with the master commit and a summary under 200 words, then stop.
