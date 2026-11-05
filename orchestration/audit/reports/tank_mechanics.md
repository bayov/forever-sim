# tank_mechanics pass 5 (2026-11-05, master bba052f75, client 1.60.1.70790+hf)

Due on one watched file: tools/tankbench/main.go (cb4065412, "Tank and heal races: paired per-fight sims"). Result: quiet.

**Getting current:** wago's latest is still 1.60.1.70790 (no new build). The update_ledger change in range is the audit loop's own commit. Blizzard's development notes thread was not re-read this pass (pass 4 hit the forum's 422 rate limit); `dev_notes_seen` stays at post 9 (2026-11-04T02:12:27Z), and the next full pass should check it.

**What changed:** tankbench gained three flags for tools/dps_report/bench_races.py, all tooling:

| Flag | What it does | Mechanic touched |
|---|---|---|
| `-first-seed` | sets SimOptions.RandomSeed raw, in place of `tankscore.Seed(-seed)` | none |
| `-labeled` | SimOptions.UseLabeledRands (one random stream per roll kind) | none: same rolls, different stream layout |
| `-vectors` | SaveAllValues; appends one JSON line a build with per-fight SafeDPS, survive HPS (+Inf as -1), DTPS, TPS | none |

The other engine edits in range (sim/core/raid.go, sim_concurrent.go, proto FightDeaths) are the healing fight's per-fight death flags; no tank path reads them. No tank ability, talent or engine file changed.

**Checks (warrior_protection request, 20 fights):**

| Check | Result |
|---|---|
| `-vectors` lengths equal fights for every series | 20/20/20/20 |
| `-vectors` means match the printed table (TPS) | 1329.8 vs 1330 |
| `-first-seed 1000 -iters 3` fight 3 equals `-first-seed 1002 -iters 1` | DTPS 499.08, TPS 1357.77, survive 486.23 in both |
| `-labeled` keeps the outcome distribution (different stream) | DTPS 485.5 vs 467.2, TPS 1335.3 vs 1329.8 at 20 fights, within noise |

No new findings. Verdicts in `ref/tank_mechanics/verdicts.md` unchanged.
