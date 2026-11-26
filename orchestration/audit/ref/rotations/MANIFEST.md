# Rotation audit: reference files

- `baseline_8569165_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  85691655a (pass 80), all 68 rows, 3000 it, seed 424242 (CLI `/tmp/audit-cache/cli/85691655ab99/wowsimcli`). Moves since
  0404484fb come from the wowsims/forever evidence batches (EV-*): hunters +3 to +4.9% (pet white swing takes the
  hunter's crit, EV-0134), rogues -1.4 to -3.0% (weapon PPMs and set procs, EV-0144), warriors and casters about 1-2%.

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
