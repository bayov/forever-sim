# client_update reference data

What the client-update auditor keeps outside the repo and how to re-create it, plus the small reference copies committed here. The scripts are in `tools/audit/client_update/`. Set three variables first:

```sh
export AUDIT_TREE=/path/to/forever-sim   # the checkout
export AUDIT_WORK=/path/to/scratch       # downloads and outputs, not committed
export AUDIT_REF=$AUDIT_TREE/orchestration/audit/ref/client_update
```

## Not committed (re-create on demand)

### DB2 tables: `$AUDIT_WORK/db2/<Table>.<build>.csv`

- **Source:** wago.tools CSV dumps, `https://wago.tools/db2/<Table>/csv?build=1.60.1.<build>`.
- **Tables:** `tools/audit/client_update/tables.txt` lists 39. That's everything in the ledger's "How it was checked" list, plus ItemXItemEffect, SpellXSpellVisual, SpellRadius and SpellRange.
  - SpellProcsPerMinuteMod and SpellScaling are not served for 1.60 builds, so the list leaves them out.
  - The trait curve tables (Curve, CurvePoint) were never downloaded. A talent changed only through its curve shows as a `TraitDefinitionEffectPoints ... CurveID: a -> b` line. Read the rank spells' own SpellEffect rows instead.
- **Builds held:**
  - 69893 (beta open), 70009, 70170, 70291, 70334, 70338, 70374, 70425 and 70450.
  - ItemXItemEffect only for 70425 and 70450.
  - About 20 MB a build, 169 MB in all.
- **To re-create:**

  ```sh
  tools/audit/client_update/wago_fetch.sh 70425 70450   # any builds; skips files already present
  ```

  A routine pass needs only the previous baseline and the new build, e.g. `wago_fetch.sh 70450 <new>`. Re-create older builds only for a question about history.

### Build diffs: `$AUDIT_WORK/diffs/<old>-<new>.md`

```sh
mkdir -p $AUDIT_WORK/diffs
python3 -I tools/audit/client_update/db2diff.py $AUDIT_WORK/db2 70425 70450 $AUDIT_WORK/diffs/70425-70450.md
```

The diffs for every consecutive pair from 69893 to 70450 are committed in `diffs/` here, so the history doesn't need re-downloading. `trait_tree.py` reads them from `$AUDIT_WORK/diffs`, so copy them there (`cp $AUDIT_REF/diffs/*.md $AUDIT_WORK/diffs/`).

### Simulator binaries, for before/after DPS

```sh
cd $AUDIT_TREE
PATH=$PATH:$(go env GOPATH)/bin npx protoc -I=./proto --go_out=./sim/core ./proto/*.proto
go build --tags=with_db -o $AUDIT_WORK/bin/after ./cmd/wowsimcli
# the same on a detached origin/master worktree for $AUDIT_WORK/bin/before
```

### Raw web pages

The raw forum JSON and the foreverchanges and Blizzard news HTML can be fetched again. The commands are in `orchestration/audit/passes/client_update.md`, step 1.

## Committed here

| Path | What | Size |
|---|---|---|
| `diffs/69893-70009.md` … `diffs/70425-70450.md` | `db2diff.py` output for every consecutive beta build pair, grouped by class then spell | 1.2 MB |
| `sources/devnotes.txt` | The development notes thread (topic 2360696) as text, as of 70425 (posts 1, 4, 5, 6) | 79 KB |
| `sources/devnotes_post6_edited_2026-10-16.txt` | The same thread after post 6's wording-only edit of 2026-10-16 20:50. Diff against the next fetch to spot quiet edits | 79 KB |
| `sources/fc_patch-notes.txt`, `fc_beta.txt`, `fc_changes.txt` | foreverchanges.pro patch notes, beta and changes pages as text (2026-10-16; 70450 not yet listed) | 99 KB |
| `sources/news243015*.txt`, `news24302097.txt`, `news24302099.txt`, `news24310968.txt` | The five class deep-dive articles (Hunter+Druid, Priest+Warrior, Mage+Shaman, Paladin, Rogue+Warlock) | 105 KB |
| `class_audit_brief.md` | The brief given to each per-class agent for the full 70009–70425 audit; reusable for a large build | 9 KB |

## Baseline

- The newest audited build is **1.60.1.70450** (2026-10-16). It is cosmetic only against 70425; see `orchestration/update_ledger.md`, "Build 1.60.1.70450".
- `.agents/skills/dps-report/state.json` `foreverClientBuild` says the same.
