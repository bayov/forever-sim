Round 3 of the Protection warrior presets (2026-11-04), on tank-warrior-3 merged with origin/master fd9538c.
Binaries built with: go build -tags=with_db -o $BIN/{rotopt,tankbench,abtest} ./tools/rotopt ./tools/tankbench ./tools/audit/tank_rotations/abtest
(the proto Go code first: protoc -I=./proto --go_out=./sim/core ./proto/*.proto with protoc-gen-go v1.36.6).

Scripts (run from the repo, BIN=<binary dir>):
  opt.sh <preset> <out> <request> [flags]  rotopt -template warrior_tank -tank <preset> on the six scoring encounters,
                                           -iters 1000 -confirm-iters 3000 (flags override; -eval for one build, -seed N).
  ab.sh <out> <iters> <seed> <preset> A B  abtest -preset <preset> -iters <iters> -seed <seed> A B (B - A).
  kv.sh <preset> <name> <knobs>            req/<name>_<preset>.json (variants deleted after the runs; rebuild with kv.sh/mkvar.py): the request with the APL rotopt writes for knobs/<preset> plus <knobs>.
  mkvar.py <req> <out> 3=:1889 ...         a request variant (slot=item:enchant, apl=, race=, talents=).
  races.sh <preset> <knobs> <iters> <out> <request>   rotopt -eval for every race.
  bench_summary.py                         bench/*.txt -> summary_rows.tsv (tools/audit/tank_rotations/bench_summary.py, paths changed).
knobs/<preset>: the final knob strings (later duplicates win). req/cur_* = round-2 presets, req/final/F_* = final requests.

Files:
  before_<preset>_s{1,2}.txt  opt.sh <preset> ... -knobs <round-2 knobs> -eval -iters 10000 -confirm-iters 10000 -seed N
  after_<preset>_s{1,2}.txt   the same on the final requests
  t_<n>_<preset>.txt          opt.sh ... -talent-search (n = alternation round); k_<n>_<preset>.txt the knob search
  races_<preset>.txt          races.sh at 1,000; races3k_<preset>.txt Undead and Tauren at 3,000
  o3_*/o10_*.txt              ab.sh at 3,000/10,000 iterations, _sN = seed N; ab_summary.txt indexes them
  matrix3k.txt                each final request scored at each weight (opt.sh -eval, 3,000)
  bench/<request>__<boss>.txt tankbench -iters 1000 -encounter tools/encounters/<boss>.json tools/tankbench/req/<request>.json
  summary_rows.tsv            those as orchestration/audit/ref/tank_bench/summary.tsv rows (no header lines)
