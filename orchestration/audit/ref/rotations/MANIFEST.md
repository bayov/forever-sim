# Rotation audit: reference files

- `baseline_c23ef04_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  c23ef0440 (pass 44). Since dd053d2 (sim only; no req .json changed): hunter pets lose Pet Aggression and
  get the new swing (hunters -38 to -80), two-hand white crits double the whole swing's Rage (2H warriors
  +10 to +15), the October 30 hotfix values for Frost and Arcane mage (Frostfire +30, Frost +21, Arcane -17).
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at c23ef0440 (`/tmp/audit-cache/cli/c23ef04408a0/wowsimcli`).
  - Next pass: build the CLI on the new master and run `baseline.py` with the same arguments. Diff the
    output against this file. If no req file changed, any row that moved more than about 1 DPS is a sim
    change.
  - When the req files change too, use `regress.py` (old CLI vs new CLI on the new req files) to tell sim
    changes apart from setting changes.

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
