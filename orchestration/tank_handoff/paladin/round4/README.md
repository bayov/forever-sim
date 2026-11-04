# Protection paladin, round 4: consumables (tank_fairness-20261104-10)

Branch tank-fixes5. Method and winners: docs/tank_paladin.md "Round 4 consumables". `ab_summary.txt` is every
paired A/B (`ab/`), `sum_s1_3k*.txt` the 3,000 screens, `before_10k_seed1.txt` / `after_10k_seed1.txt` the bench
at 10,000 iterations (seed 1), `req_before/` round 3's requests, `var*/` the variant requests, `variants.txt`,
`mkcons.py`, `mkcombo.sh`, `run*.sh` the scripts (build abtest and tankbench with `-tags=with_db`; the proto Go
files are generated with `npx protoc -I=./proto --go_out=./sim/core ./proto/*.proto`), `summary_paladin_rows.tsv`
the replacement rows for orchestration/audit/ref/tank_bench/summary.tsv (1,000 iterations, default seed), `tsc.txt`.
Winners (standard set; Zanza, Rum, Scorpok and Sapper left out per docs/tank_druid.md call 10): Max threat and Balanced Supreme Power / Nightfin / Elemental stone; Max survival Natural Precision / Dirge chops / Elemental. `var/` and `after_10k_seed1_with_extras.txt` hold the first search that included the extras.
