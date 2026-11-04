# Protection paladin, round 4: consumables (tank_fairness-20261104-10)

Branch tank-fixes5. Method and winners: docs/tank_paladin.md "Round 4 consumables". `ab_summary.txt` is every
paired A/B (`ab/`), `sum_s1_3k*.txt` the 3,000 screens, `before_10k_seed1.txt` / `after_10k_seed1.txt` the bench
at 10,000 iterations (seed 1), `req_before/` round 3's requests, `var*/` the variant requests, `variants.txt`,
`mkcons.py`, `mkcombo.sh`, `run*.sh` the scripts (build abtest and tankbench with `-tags=with_db`; the proto Go
files are generated with `npx protoc -I=./proto --go_out=./sim/core ./proto/*.proto`), `summary_paladin_rows.tsv`
the replacement rows for orchestration/audit/ref/tank_bench/summary.tsv (1,000 iterations, default seed; now the
second pass's requests, from `extras/rows_raw/` via `extras/bench_summary.py`), `tsc.txt`.
First pass (tank-fixes5, standard set without Zanza, Rum, Blasted Lands buffs and Sapper): Max threat and Balanced
Supreme Power / Nightfin / Elemental stone; Max survival Natural Precision / Dirge chops / Elemental. `var/` and
`after_10k_seed1_with_extras.txt` hold an earlier search that included the extras.

## extras/: second pass, the full raid set (tank-fixes6)

The exclusion was a worker's choice, not a ruling; standards item 11 applies. Start = master's requests (`extras/var/base_*`,
= `extras/before/`). `variants.txt` single swaps (every Zanza / Blasted Lands buff, alcohol, explosive) ->
`screen_s1_3k.txt`; `variants2.txt` stacks (`f_<zanza>` = Zanza + Rumsey Rum Black Label + Goblin Sapper, plus
leave-outs and the Holy Hand Grenade); `variants3.txt` / `variants3q.txt` the flask/food/stone/elixir re-check on the
Spirit (`r_*`) and Scorpok (`q_*`) stacks -> `reconfirm_s1_3k.txt`. `mkvar.sh` writes `var/` (`BASE=` picks the start),
`run.sh` the paired abtest (`ab/s<seed>_<A>_vs_<B>_<preset><suffix>.txt`; summarize with `../summ.py`), `final.sh` the
finals (paired before/after and `../bench.py` on `before/` and `after/`, 10,000, seeds 1 and 2:
`{before,after}_10k_seed{1,2}.txt`). Binaries built with `go build -tags=with_db` into /tmp/claude-0/paladin_tank.
Winners: Max threat Ground Scorpok Assay, Balanced and Max survival Spirit of Zanza; all three Rumsey Rum Black Label
and the Goblin Sapper Charge (Engineering was already their profession); flask/food/stone/elixirs unchanged.
`after/` = tools/tankbench/req/paladin_protection_*.json. Numbers: docs/tank_paladin.md "Round 4 consumables".
