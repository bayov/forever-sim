# Rotation audit: reference files

- `baseline_6b480b0_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  6b480b060 (pass 53). Against the previous f331f0c baseline only rogue and warrior rows moved, both from server
  hotfixes (rogue req files only remapped, warrior ones re-tuned): rogue 113483 (316a600ac: Shank no longer procs poisons or enchants, Mutilate
  flat bonus; Seal Fate Daggers -17.8, Seal Fate Swords -16.0, Mutilate -8.2, Combat Swords -4.1, Combat Daggers
  +2.2) and warrior 113539 (7cd399e88 + e07adf11b: Raging Blows out, Reckless Swings on a 1 s GCD; -6.6 to -26.7,
  Dragonkin +6.0). The downranking (b57f5cc0a) and Chain Lightning 0.714 (280fd26e2) changes moved no row.
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at 6b480b060 (`/tmp/audit-cache/cli/6b480b06046c/wowsimcli`).
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
