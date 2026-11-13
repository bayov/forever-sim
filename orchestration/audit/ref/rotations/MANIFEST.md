# Rotation audit: reference files

- `baseline_3a86d4a_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  3a86d4a6b (pass 64). Against the previous 97da1cc baseline: Enhancement +20.0 to +24.6 (c7a99fa49: Windfury
  Weapon's attacks roll the shaman's own Windfury Totem under Forever); every other preset unchanged.
  (The shared CLI was `/tmp/audit-cache/cli/3a86d4a6bf60/wowsimcli`; made the same way as before.)

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
