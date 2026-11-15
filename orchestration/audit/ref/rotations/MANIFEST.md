# Rotation audit: reference files

- `baseline_25ce81d_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  25ce81de7 (pass 67), all 68 rows re-simmed (shared CLI `/tmp/audit-cache/cli/25ce81de7755/wowsimcli`). Moves since
  70ea588 come from the client 70938 / Nov 13 hotfix round (Windfury Totem on every main-hand melee hit, melee
  Judgement of Command, Two-Handed Weapon Specialization's Holy reach) and that round's re-tuned rotations.

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
