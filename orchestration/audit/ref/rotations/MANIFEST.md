# Rotation audit: reference files

- `baseline_8b5137c_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  8b5137c6d (pass 71), all 68 rows, 3000 it, seed 424242 (CLI `/tmp/audit-cache/cli/8b5137c6d6a2/wowsimcli`). Moves since
  a6677362f: paladins +0.5 to +1.1% (Nov 19 Two-Handed Weapon Specialization hotfix; Undead also its Holy Strike reorder),
  rogues +0.2 to +0.6% (Shank procs poisons), warriors +0.05 to +0.3% (Nov 16 Flurry hotfix).
## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
