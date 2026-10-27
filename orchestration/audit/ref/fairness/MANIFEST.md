# Fairness audit reference files

| File | What it is |
|---|---|
| `sweep_baseline.log` | `tools/audit/fairness/sweep.py` output on master 06a8987 (2026-10-27, pass 10): every base report request at every race, 5000 iterations, seed 1. The next pass compares against it. |
| `table_baseline.md` | `table.py` output built from that log: the cross-spec ranking. |
| `fairness_report.md` | The full fairness report: section 1 is the current table, with every earlier table kept in `<details>` blocks. It also holds the sensitivities (§3), the asymmetries (§4), the stat-mix work (§7) and the history (§6). |

`gear_targets_v3.json` is not copied, because it is identical to `orchestration/tools/gear_targets_v3.json`, which `audit1b.py` reads directly.

Nothing large is kept. To re-create what a pass needs:

```sh
git worktree add --detach $AUDIT_TREE origin/master          # or: git -C $AUDIT_TREE checkout --detach origin/master
AUDIT_TREE=... AUDIT_WORK=... tools/audit/fairness/build.sh  # one wowsimcli for every class, about 2 min
```

Client tables (wago.tools DB2 CSV), downloaded only when a question needs them, each into its own new empty directory:

```sh
curl -sS -o SpellInterrupts.csv "https://wago.tools/db2/SpellInterrupts/csv?build=1.60.1.70425"
# same pattern for SpellEffect, SpellMisc, SpellCooldowns, TraitDefinitionEffectPoints, CurvePoint ...
```

The pass uses no other local state. The old per-class worktrees, binaries and mksettings-built requests are retired: since 2026-10-16 the pass sims the report's own requests.
