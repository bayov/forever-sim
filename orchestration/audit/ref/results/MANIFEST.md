# Results audit: reference files

What the results audit compares against, and what it kept only locally (and how to re-create it).

## In this directory

| File | What it is | Made by |
|---|---|---|
| `req_summary.txt` | One line per DPS report request (66): `<id> mine <dps> pub <published> diff <d>`. The DPS of each committed `tools/dps_report/req/<id>.json` run as is, 20,000 iterations, seed 1, 180 s, on a wowsimcli built from master 1735382cd. All 66 are within 0.15 of `published.dps`. | Pass 40 (2026-10-17 00:34 UTC), `runreq.sh` |
| `drift_report.txt` | The drift check's output for all 66 requests on master 1735382cd: what mksettings rebuilds from each meta against the committed request. Every difference listed is expected (see the handoff, "Drift check"). | Pass 40, `regen.py` |

Both still hold at master b389688f7 (2026-10-17 14:33 +0300): nothing the binaries read changed after 1735382cd, and
`ui/dps_report/data/builds.json` (report v35) matches `req_summary.txt` on all 66 builds within 0.15 (pass 41).

## Kept locally, not committed

Re-create these only when a check needs them; a normal pass needs only the binaries.

| Local item | Size | What it was for | Re-create |
|---|---|---|---|
| `bin/wowsimcli`, `bin/mksettings` | 23 MB, 15 MB | Every pass | `tools/audit/results/build.sh` (needs protoc and protoc-gen-go, see the handoff's Setup) |
| `bin/wowsimcli_nochanpb` | 23 MB | One-off: master with channel pushback (ab517bf) reverted, to measure its effect | Not needed again. If it is: a worktree at the commit, `git revert --no-commit ab517bf`, then `go build --tags=with_db -o ... ./cmd/wowsimcli` |
| `runs/req34` to `runs/req40` | about 6.5 MB each | Each pass's inputs and outputs (`<id>.in.json`, `<id>.out.json`, `summary.txt`) | `tools/audit/results/pass.sh <tag>` on the commit |
| `drift/<id>.regen.json` | 66 files, 2.6 MB | mksettings output for the drift check | `regen.py` writes them to `$AUDIT_WORK/drift/` |
| `dl/sao/SpellAuraOptions.csv`, `dl/sao/Spell.csv` | full tables | Proc chance and proc type checks (for example Enrage 12317: ProcChance 30, ProcTypeMask 139944) | `python3 tools/sources/wago.py path 1.60.1.70338 SpellAuraOptions` and the same for `Spell` |
| `db/CurvePoint.csv`, `db/TraitDefinition.csv`, `db/TraitDefinitionEffectPoints.csv` | 1 MB | `curve.py`: talent-rank values that the client keeps in curves | `curve.py <build> <spell>` downloads them through `tools/sources/wago.py` |
| `talent.csv` | 31 KB | Talent table (spell ids per rank) | `python3 tools/sources/wago.py path 1.60.1.70338 Talent` |
| Single spells | | Base points, coefficients, ranks | `https://wago.tools/db2/SpellEffect/csv?build=1.60.1.70338&filter%5BSpellID%5D=<id>`, likewise `Spell` (`filter%5BID%5D`), `SpellName`, `SpellMisc` |
| `spellname.csv` | 0.8 MB | Spell id to name lookup | `python3 tools/sources/wago.py path 1.60.1.70338 SpellName` |
| `jobs*.txt`, `st_*.md`, `status.txt`, `builds40.json`, `builds41.json` | | Superseded: job lists from the per-branch method (passes 1 to 33), sim-status snapshots from 2026-10-12, and two builds.json copies for the pass-41 check (both are in git history) | Not needed |

Client builds: 1.60.1.70338 was the reference until the 70425 ports; the repo now names 1.60.1.70450 (spells, talents, racials and enchants
identical to 70425, `.agents/skills/dps-report/state.json`). Use the newest build in the URLs above. wago.tools lists builds at
https://wago.tools/builds.
