# Rotation audit: reference files

- `baseline_796495d_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  796495d6f (pass 47). Since 5a6e9c6: the 2026-11-02 rotation re-searches (rogue builders whenever affordable, warrior
  knobs after the two-hand Rage crit fix, hunter Aimed Shot above 40% mana, mage/balance end-of-fight gates), caster
  distances (standards 33), Balance Wrath as Night Elf, the Demonology synthetic set (+7.5 to +9.8), and the deep Holy
  Shock presets removed. Sim-only moves (regress.py, old vs new CLI on the same reqs) were all under 1 DPS.
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at 796495d6f (`/tmp/audit-cache/cli/796495d6fdae/wowsimcli`).
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
