Warrior round 5 (2026-11-04, branch tank-fixes6 on master 9495b20): talents and rotation knobs of the three Protection
presets re-searched on round 4's full raid consumable set, the way round 3 did. Results: docs/tank_warrior.md, "Round 5".

Binaries: go build -tags=with_db -o /tmp/claude-0/warrior_r5/ ./tools/rotopt ./tools/tankbench ./tools/audit/tank_rotations/abtest
The rotopt template (tools/rotopt/warrior_tank.go) now writes round 4's Goblin Sapper Charge line after the trinkets, so
its APL for knobs/<preset> is byte for byte the UI's ui/tank_warrior/apls/level60_<preset>.apl.json (apl/gen_*.apl.json).

Scripts (run from anywhere; BIN / ABTEST / TANKBENCH override the binaries):
  opt.sh <preset> <out> <request> [flags]   rotopt -template warrior_tank -tank <preset> on the six scoring bosses, 1,000 / 3,000 (round 3's)
  search.sh <preset> <n>                    t_<n>_<preset>.txt talent search with knobs/knobs_<preset>, then k_<n>_<preset>.txt the knob search on the talents it found
  kv.sh <preset> <name> <knobs> [talents]   req/<name>_<preset>.json: the request with the APL rotopt writes for knobs/<preset> plus <knobs> (and the talents)
  ab.sh <out> <iters> <seed> <preset> A B   paired abtest -preset <preset> (own weight), B - A -> ab/<out>.txt
  batch1.sh                                 the candidates at 10,000 on seeds 1 and 2 (tankscore.Seed: disjoint), A = req/cur_<preset>.json
  bench.sh; python3 <this dir>/bench_summary.py (from the repo root)   tankbench -iters 1000, default seed -> bench/, summary_rows.tsv
Requests: req/warrior_protection_* = the round 4 requests (A of every A/B, as req/cur_*); final/ = the new tools/tankbench/req.
The variant requests req/<name>_<preset>.json were deleted after the runs: rebuild them with kv.sh (batch1.sh needs them).
knobs/<preset>: the final knob strings (later duplicates win; Balanced gained sunderRage=25).

Stages (own weight, six scoring bosses):
  t_1, k_1     talent then knob search, all three presets (1,000 a candidate, 3,000 to confirm)
  t_2, k_2     Balanced again with sunderRage=25 (nothing moved); Max threat's k_1 already ran on its new talents (nothing moved)
  ab/o10_*     10,000 iterations, seeds 1 and 2, B - A against the round 4 request; ab_summary.txt lists them
    tal       = talents 05-05050005-250330020330001351 (Deflection 5, no Blood Craze; Max threat's new talents, also tried at Balanced, Max survival)
    talwall12 = tal + Shield Wall in the last 12 s (Max threat)
    su25      = Sunder Armor filler from 25 Rage instead of 30 (Balanced, Max survival); talsu25 both (Balanced)
  bench/       the final requests, tankbench 1,000, default seed; summary_rows.tsv = their 24 summary.tsv rows
