# Pass: Forever client updates

Follow `orchestration/audit/PASS.md` around this doc. It started as the long-running game-update auditor's handoff (2026-10-17); keep it current when the method changes.

This auditor makes sure every change in a new WoW Forever client build, announced or not, is recorded in `orchestration/update_ledger.md` (this auditor `owns` it), and that every change that matters to level 60 DPS is filed as a finding, sized, for the orchestrator to land. The audit loop starts a pass only when its gate sees a 1.60 build on wago.tools newer than `client_build` in `state/client_update.json`.

**What changed with the audit loop:** a pass no longer opens PRs or messages workers. It writes the ledger, files one finding per DPS-relevant change (with the before/after numbers from step 4 when it can size it cheaply), and stops. Landing the fixes is the orchestrator's job (`.agents/skills/dps-report`, B13); steps 4 and 5 below are the method it follows.

**Baseline:** 1.60.1.70450, 2026-10-16. It is cosmetic only against 70425. The ledger covers 70009 to 70450.

**Files:**
- Scripts: `tools/audit/client_update/`.
- Reference data and the commands to re-create it: `orchestration/audit/ref/client_update/MANIFEST.md`. Read that first.

## Rules

- In a pass, write only the ledger and your audit-loop files (`PASS.md`); the run commits them. No branches, PRs or messages. What used to become a `notes.md` "Need to Verify" item becomes a low-severity finding instead (`notes.md` belongs to other sessions).
- When the orchestrator lands a fix (B13): never merge without the user's go-ahead, never force-push. Land sim changes as small PRs on `claude/wow-forever-dps-presets-m9u15h-updates-<topic>`.
- Gate every Forever change behind the Forever ruleset (`sim.Env.IsForever()` or the ruleset check the surrounding code uses). Classic behaviour must not change.
- Read `orchestration/standards.md` before touching a preset.
- Class changes, shared mechanics, racials, items, enchants and consumables all become findings: `needs: orchestrator fix`, with the class or system in `affects`.
- Never change the sim on a tooltip alone. The client data (DB2) or Blizzard's notes must back the change. A server-side change the client doesn't show becomes a "needs verify" finding.

## Step by step

### 0. Setup

```sh
source tools/audit/setup.sh                 # AUDIT_WORK and the toolchain
export AUDIT_TREE=$PWD AUDIT_REF=$AUDIT_TREE/orchestration/audit/ref/client_update
T=tools/audit/client_update
git fetch origin master
```

### 1. Detect a new build and its notes

1. `python3 tools/sources/wago.py latest --probe 300` prints the newest 1.60 build on wago.tools. The gate runs the same command, so its `client_build` is the same answer.
   - It takes the maximum across all products (Forever appears under `wow_classic_beta`, `wow_cn_beta` and `wowf`, and `wowf` lags by days). Then it probes the db2 exports past that, because the build list has lagged behind them before.
   - If the result is the baseline (`client_build` in your state, or what the ledger last records), the pass is done: report "no new build".
2. **Blizzard development notes** (topic 2360696). Fetch the thread slowly, since the forum rate-limits:

   ```sh
   curl -sL --max-time 30 'https://us.forums.blizzard.com/en/wow/t/2360696.json?print=true' -o $AUDIT_WORK/devnotes.json
   python3 -I $T/devnotes.py $AUDIT_WORK/devnotes.json > $AUDIT_WORK/devnotes.txt
   diff $AUDIT_REF/sources/devnotes_post6_edited_2026-10-16.txt $AUDIT_WORK/devnotes.txt
   ```

   - Each new post usually covers one build. So far: post /1 is 70009, /4 is 70170, /5 is 70291 and /6 is 70425.
   - An edited post keeps its date, so compare the `upd` times.
   - The thread slug changes with each update; the numeric URL redirects to it.
   - If the JSON comes back empty or as HTML, wait a minute and retry with the slug URL from the redirect.
   - Also look for new blue posts and threads in the WoW Forever forum category, from the posters Kaivax, Aggrend, Gnedi, Linxy and Brenoch. One build (70170) also had its own "Warrior Updates in Today's Beta Build" thread.
3. **foreverchanges.pro/patch-notes.**

   ```sh
   curl -sL --max-time 30 https://foreverchanges.pro/patch-notes -o $AUDIT_WORK/fc.html
   ```

   - It's a Next.js page: the text sits in `self.__next_f.push` chunks. `tools/database/foreverchanges/scrape.py` shows how to decode them.
   - It lists each build. "Found in the game" or "Not in Blizzard's notes" lines are unannounced client changes it spotted, which makes it a good cross-check.
   - It lags a day or more behind wago.
4. **Wowhead.** Its Forever news returns HTTP 403 from cloud sessions, so don't spend time on it. Its build articles retell the Blizzard posts. The repo's `scrape-wowhead` skill and `tools/database/forever_wowhead/wowhead.py` can reach single pages if you really need one.
5. **Blizzard news** (worldofwarcraft.blizzard.com/en-us/news). Class deep dives are design articles that describe the design as of their build.

### 2. Download and diff

```sh
for b in <baseline> <new>; do python3 tools/sources/wago.py fetch $b --set client_update; done   # e.g. 70450 70471
python3 -I $T/table_summary.py <baseline> <new>
mkdir -p $AUDIT_WORK/diffs
python3 -I tools/sources/db2diff.py <baseline> <new> $AUDIT_WORK/diffs/<baseline>-<new>.md
python3 -I $T/items_diff.py <baseline> <new>
python3 -I $T/enchant_diff.py <baseline> <new>
```

**`table_summary.py`** is the first look: every table either identical or with added, removed and changed row counts.
- If SpellEffect, SpellMisc, SpellPower, SpellCooldowns, SpellAuraOptions, SpellClassOptions, the Trait* tables, ItemSparse (changed), SpellItemEnchantment and ChrRaces are all identical, and the added rows are cosmetic, the build is cosmetic. 70450 looked like that: only added rows, for pets, mounts and ensembles.

**`db2diff.py`** groups every change by class (SpellClassOptions, SkillLineAbility, talent trees, triggered spells, same names), then:
- Racial, Item, No class, and the lookup and talent-tree tables.
- Its class tags are heuristic, so also grep the "No class" section for spell names you care about.
- A heading "N spells whose only change is EffectBonusCoefficient 1 -> 0" means a batch of item procs lost spell power scaling. 70009 did that to many item spells.

**`items_diff.py`** lists added and removed items and each changed item with the changed fields. It flags:
- items a DPS-report request wears (`tools/dps_report/req/*.json`);
- items at ilvl ≥ 55 or required level ≥ 50.

**For a single spell,** run `spell.py <build> <id or exact name>` on both builds and compare. `rank.py <build> <effectIndex> <rank ids>` gives a rank's damage at 60.

**For talents,** `trait_tree.py <TraitTreeID> <builds>` pulls every diff line for a class tree, after you copy the committed diffs into `$AUDIT_WORK/diffs`.
- The curve tables aren't downloaded. A `TraitDefinitionEffectPoints CurveID a -> b` line means the talent's values changed, and you read the new numbers from the rank spells' SpellEffect rows.

**Useful client fields:**
- **SpellEffect:**
  - EffectBasePointsF is the base value; Variance is the spread, so the range is avg·(1 ± Variance/2).
  - EffectRealPointsPerLevel is growth per level up to SpellLevels.MaxLevel.
  - EffectBonusCoefficient is the spell power coefficient; BonusCoefficientFromAP is the attack power coefficient.
  - EffectAuraPeriod is the tick period in ms.
  - EffectSpellClassMask_0..3 set which spells a modifier touches; match them against SpellClassOptions.SpellClassMask of the target spells.
- **Party versus raid:**
  - Effect 35 is a party aura and 65 a raid aura.
  - ImplicitTarget 20 is party and 56 is raid.
  - Several 70425 buffs became raid-wide this way.
- **Other tables:**
  - SpellMisc: CastingTimeIndex → SpellCastTimes.Base, DurationIndex → SpellDuration, and the Attributes bit flags. For example Attributes_6 bit 0x8000 means no pushback (Slam in 70425).
  - SpellInterrupts.InterruptFlags 8 is the auto-attack reset (Faerie Fire lost it in 70170).
  - SpellCooldowns: RecoveryTime and CategoryRecoveryTime, in ms.
  - SpellAuraOptions: ProcChance, ProcCharges and ProcCategoryRecovery (the internal cooldown).

### 3. Classify each change

| Class | Examples | Action |
|---|---|---|
| Cosmetic | pets, mounts, toys, appearances, ensembles, tents, visuals, tooltip text, labels, NPC or quest spells, new low-level items | One summary line in the build's ledger section. No rows. |
| Class (spell, talent, pet) | coefficients, base points, costs, cooldowns, cast times, procs, talent values, new or removed talents | A ledger row per change. Send it to the class worker through the orchestrator. |
| Racial | ChrRaces, racial spells (Eureka!, Touch of the Grave, Blood Fury…) | A ledger row. Fix it yourself in `sim/core/racials_forever*.go`. |
| Item | ItemSparse stats on any level 60 or preset item, ItemEffect or proc spell changes | A ledger row. Fix it in `tools/database/forever_items.go` and regenerate the db, or in `sim/common/item_effects.go` and the weapon procs. |
| Enchant | SpellItemEnchantment | A ledger row. Fix it in `tools/database/forever_enchants.go` and regenerate. |
| Shared mechanic, buff or debuff | raid buffs, totems, shouts, consumables | A ledger row. Fix it in `sim/core/buffs.go`, `debuffs.go` and `consumes`. |

**Does it matter to level 60 DPS?** Each row gets one value:
- `yes`: it changes a preset or a modelled mechanic.
- `minor`: it is modelled, but no preset uses it.
- `n/a`: healing, tanking, threat, PvP, leveling or UI.

**Announced but absent from the client:** the change is server side (a hotfix), or it lands in a later client. The 10-02 warrior pass showed up only in 70291. Mark it `ambiguous (needs verify)` and file a low-severity finding.

### 4. Check the sim and size the impact

1. For every `yes` or `minor` row, find where the sim models it (grep the spell ID, name or constant in `sim/`), read the value, and set the status:
   - `landed`, `missing`, `partial` or `wrong`;
   - `not simulated on purpose`;
   - `in flight (worker PR)`;
   - `ambiguous (needs verify)`;
   - `n/a`.

   Check "landed" by reading the value, not by finding a function name. Check the open worker PRs and branches (`git log origin/master..origin/claude/wow-forever-dps-presets-m9u15h-<class>`) before calling something missing.
2. To size a fix, build a before and an after `wowsimcli` (MANIFEST.md shows how), then run:

   ```sh
   $T/dps_beforeafter.sh $AUDIT_WORK/bin/before $AUDIT_WORK/bin/after <request ids...>
   ```

   That runs 180 s, 20k iterations, seeds 1 and 1000003. Request IDs are the file names in `tools/dps_report/req/`, without `.json`. Pick the requests whose class, race, gear or buffs touch the change.
   - For a preset or request-only change, use `req_beforeafter.sh <cli> <ids>` instead.
   - Report each result as DPS and percent.

### 5. Land fixes

- Make one small PR per topic, with:
  - a Forever-gated code change;
  - a test like `sim/warrior/dps_warrior/items_forever_test.go` or `sim/warlock/dps/eureka_forever_test.go` that checks Forever and Classic;
  - goldens updated (`go test --tags=with_db ./sim/...`, then copy the `.results.tmp` files over the `.results` files after reading each diff);
  - the before/after table in the body.
- If you touched the db: `go run tools/database/gen_db/*.go -outDir=./assets -gen=db`.
- **Gotchas when building and testing:**
  - Regenerate the Go protos before building (`npx protoc -I=./proto --go_out=./sim/core ./proto/*.proto` with protoc-gen-go on PATH).
  - Regenerate the TS protos before `npx tsc --noEmit`, because master moves them often.
  - gen_db once crashed in mage `applyArcaneTalents` on a stale hard-coded talent string. The fix is the string from `mage.json` (`gen_db/main.go`).
  - On merge conflicts in goldens, take master's, rerun the tests and copy the `.results.tmp` files again.
- Report each PR and its DPS table to the orchestrator.

### 6. Update the ledger

`orchestration/update_ledger.md`:

**New build section.** Add a `## Build 1.60.1.<N> (<date>)` section just above `## Per-class ledgers`, newest above older ones.
- A cosmetic build gets one paragraph, like the 70450 one: what changed, "no class, racial, talent, enchant or existing-item change", the dev-notes status, and "No rows added."
- A build with real changes gets a table with the per-class column layout:

  ```
  | ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
  ```

  - ID: `<CLASS>-<build>-NN` for an announced line, `<CLASS>-U-NN` for an unannounced client change, `SHR-...` for shared systems.
  - Source: `P<post>` (dev-notes post), `DD` (deep dive) or `FC` (foreverchanges), each with its date and build.
  - Client check: the table, row or spell ID, old → new, and the number the sim should use.

**Other updates:**
- Update the Summary table counts.
- Update "This audit's PRs" for any PR you open.
- Update "Rulings needed" for anything only the user can decide.
- Note the new build for `.agents/skills/dps-report/state.json` `foreverClientBuild` in your report; the orchestrator updates that file.
- Leave `notes.md` alone; its "Need to Verify" items come from the findings, when the orchestrator handles them.

**Large builds.** For a build with many class changes (like 70425), split the work with one agent per class, using `orchestration/audit/ref/client_update/class_audit_brief.md` adapted to the new build, then review and merge their sections.

### 7. Report

- One finding per DPS-relevant change (`yes` or `minor` with status `missing`, `partial` or `wrong`) in `findings/client_update.md`, with the ledger row ID, the client check, the sim location and the sized effect. Things only the user can decide are `needs: user ruling` (and go in the ledger's "Rulings needed").
- `reports/client_update.md`: the build, one line per DPS-relevant change with its sim status, and the ledger section you added.
- Set `client_build` in your state to the build you processed.

If nothing DPS-relevant changed, the report says so in two or three lines and there are no findings.

## Current state (2026-10-17)

- **Ledger:** 718 rows for 70009–70425, a "settles" section, a 70450 section (cosmetic) and row SHR-U-14 (Lesser Intellect bracer, #93).
- **Merged audit PRs:** #61 (druid), #69 (racials), #71 (warrior), #73 (rogue), #77 (items and Grace of Air), #88 (presets), #93 (enchant and gen_db), #95 (raid buff UI), #113 (70450 ledger line).
- **Open items routed to others:** see the ledger's "Missed by the class workers" and "Rulings needed" sections. Their status there is as of 2026-10-16. Check master before repeating any of them.
- **Sources at the baseline:**
  - The dev notes' latest post is /6 (2026-10-15, 70425), with a wording-only edit at 2026-10-16 20:50.
  - foreverchanges did not yet list 70450.
  - Wowhead is blocked.
