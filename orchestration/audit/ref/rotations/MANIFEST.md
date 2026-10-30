# Rotation audit: reference files

- `baseline_fb14858_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  fb1485888 (pass 40, client 1.60.1.70707 mage talents, extra attacks keep the swing timer, Windfury Totem
  100 ms ICD, Eviscerate 4% AP: warriors +6 to +10%, survival +4%, paladins/rogues/feral +2 to +3%,
  fire -2.4%, frostfire -3.6%, arcane +1.4%).
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at fb1485888 (`/tmp/audit-cache/cli/fb1485888563/wowsimcli`).
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
