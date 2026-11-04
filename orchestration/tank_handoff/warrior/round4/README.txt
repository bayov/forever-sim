Round 4 of the Protection warrior (2026-11-04, branch tank-fixes5): consumable search (tank_fairness-20261104-10), UI vs bench (-11), stale doc line (-9).
Binaries: go build -tags=with_db -o $BIN ./tools/tankbench ./tools/audit/tank_rotations/abtest (proto Go code first, protoc -I=./proto --go_out=./sim/core ./proto/*.proto).
gen.py <preset>  req/<variant>_<preset>.json: tools/tankbench/req/warrior_protection_<preset>.json with one consumable changed (run from this dir; reads the current request, so run it on the round-3 requests in req_before/).
ab.sh <preset> <iters> <seed> <variants...>   A = the request, B = req/<variant>; ab2.sh the same on base2/ with req2/ (Precision + Elemental stone).
Screen: all variants, 3 presets, 3,000 iterations seed 1. Decide: winners at 10,000 on seeds 1 and 2 (disjoint). Winners: Natural Precision flask and Elemental Sharpening Stone (all three), Dirge's chops and Greater Stoneshield Potion (Max survival).
bench/<request>__<boss>.txt: tankbench -iters 1000 on the final requests; bench_summary.py (run from the repo root) -> summary_rows.tsv (rows for orchestration/audit/ref/tank_bench/summary.tsv).
tsc.txt: npx tsc --noEmit (empty: clean).
