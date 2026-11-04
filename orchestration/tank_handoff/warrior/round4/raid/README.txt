Warrior round 4, the full raid consumable set (2026-11-04, branch tank-fixes6): standards 11 for the three Protection presets.
Docs/tank_druid.md call 10 (no Zanza / Blasted Lands, Rum or Sapper on tanks) was a worker's choice, not a ruling, and is dropped.
Method and results: docs/tank_warrior.md, "Round 4 consumables, the full raid set".

Binaries: PATH=$PATH:~/go/bin npx protoc -I=./proto --go_out=./sim/core ./proto/*.proto; go build -tags=with_db -o /tmp/claude-0/warrior_tank/ ./tools/tankbench ./tools/audit/tank_rotations/abtest
Scripts (run from this dir unless noted):
  mkvar.py <base> <out> key=Value... apl=<itemId>@<index>  one variant (consumes field, = - removes; apl inserts a castSpell item action)
  mk.sh <variants file> <base dir> <out dir>                writes <out>/<name>_<preset>.json for the three presets
  run.sh <base dir> <var dir> <seed> <iters> <tag> <variant>[:<preset>]...   paired abtest -preset <p> (own weight), A = base, B = variant -> ab/<tag>_s<seed>_<variant>_<preset>.txt
  summ.sh '<glob>'   scoring mean diff ± paired error (B - A) of the ab files; ab_summary.txt = all of them
  bench.sh; bench_summary.py (from the repo root)  tankbench -iters 1000, default seed, final requests -> bench/, ../summary_warrior_rows.tsv
  ba.sh <seed>   before (base/) vs after (final/) at 10,000 -> ab/ba10k_s<seed>_<preset>.txt
Request sets: base/ = the requests before this pass (round 4's standard set); base3/ = the stacked candidate; final/ = the new tools/tankbench/req (base3 with the UI APL's note text).
Stages (six scoring bosses at the preset's own weight):
  v1  variants1.txt on base/, 3,000, seed 1: every Zanza-slot buff, every alcohol, the sapper (three APL placements) and every explosive the professions allow (Engineering is already the warrior's: no profession change).
  v2  variants2.txt on base/, 3,000, seed 1: Zanza choice + Rum + Sapper stacked, and leave-outs.
  v3  variants3.txt on base3/, 3,000, seed 1: Zanza swaps, Grog, flask, food, elixirs, stone, potion and chili re-checked on the stacked set.
  d10k  the close v3 calls at 10,000 on seeds 1 and 2; lo10k  variants4.txt leave-outs on base3/ at 10,000, seeds 1 and 2 (lo_sapper also drops the APL action).
  ba10k  before -> after, 10,000, seeds 1 and 2.
