# Rotation audit: reference files

- `baseline_0404484_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  0404484fb (pass 79), all 68 rows, 3000 it, seed 424242 (CLI `/tmp/audit-cache/cli/0404484fb68f/wowsimcli`). Moves since
  eceb09019 come mostly from world buffs doing nothing in raids (rogue Combat Swords -11%, hunters -10%); warriors +0.4 to
  +0.8% from the stance-modified Rage from hits taken.

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
