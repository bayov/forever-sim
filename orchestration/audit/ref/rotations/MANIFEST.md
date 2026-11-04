# Rotation audit: reference files

- `baseline_0facd94_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  0facd9428 (pass 54). Against the previous 6b480b0 baseline: Feral cat +26 (Nature's Reach 2, hit-4 step), Ret
  +10 to +12 (hit-3 step), rogues +5 to +8 (hit+2/+3 steps), Arcane +7 (hit-1, Master of Elements 1),
  Elemental +7 (hotfix 113380), Enhancement +7.5 to +9 (talents, hit steps, hotfix), warriors -2.9 to +1.3
  (Unbridled Wrath on autos only, Reckless Swings on landed swings, Fury hit+2), Destruction fire +0.8.
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at 0facd9428 (`/tmp/audit-cache/cli/0facd9428f42/wowsimcli`).
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
