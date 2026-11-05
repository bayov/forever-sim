# tank_fairness pass 5: 2026-11-05, master bba052f

Due: 1 watched file changed since a079415: `tools/tankbench/main.go`. Nothing else under the tank presets, the
encounters, tankscore or the four tank classes' sim or UI code changed (`git diff --stat a0794153 bba052f7`), so pass
4's numbers (`ref/tank_fairness/summary.tsv`, 3,000 iterations, seeds 1 and 7) still stand and were not re-simmed.

## The tankbench change

Three opt-in flags for the DPS Report's paired race sims (tools/dps_report/bench_races.py, tankheal.py):

| flag | effect | default |
|---|---|---|
| `-first-seed N` | raw seed of fight 0 in place of `tankscore.Seed(-seed)` | 0 = unchanged |
| `-labeled` | `SimOptions.UseLabeledRands` (one random stream per roll kind) | off |
| `-vectors FILE` | `SaveAllValues`, appends per-fight SafeDPS / survive HPS (-1 for none) / DTPS / TPS as JSON | off |

With none set the request, seed and scoring are byte for byte what they were, so no bench number, ranking or preset
comparison moves. The flags apply to every request alike; nothing is spec-specific. Not a fairness issue.

## Same terms, re-checked where the change reaches: races

The change exists to re-sim the race tables (ui/dps_report/data/tank_races.json, c46062a). Each preset's race against
its build's table, on its own weight's score:

| tank | preset races (threat / balanced / survival) | top of its race table | preset row in table |
|---|---|---|---|
| warrior | Undead ×3 | Undead at all three | yes |
| paladin | Undead ×3 | Undead at all three | yes |
| bear | Night Elf ×3 | Night Elf at all three | yes |
| Warden | Night Elf / Night Elf / Undead | Orc / Orc (Troll within error) / Tauren | **no** |

The Warden's races are not playable: Forever's client (CharBaseInfo, via racials.json) has no Night Elf or Undead
shaman, and run_build.CLASS_RACES lists Dwarf, Orc, Tauren, Troll and Skyborne Windshaper. Round 3 picked them from a
race A/B that included every race (docs/tank_shaman.md calls (f), (g)). Estimated cost of moving to the best legal
race, from round 3's paired A/B: about -76 / -109 / -43 (2.4% / 6% / 2%). The Warden is last at every weight already,
so the ranking does not move; the gap to the bear grows. New finding **tank_fairness-20261105-2** (high, worker fix).

## Unchanged from pass 4

- Own-weight six-boss means (seed 1): Max threat warrior 4650 / bear 3491 / paladin 3249 / Warden 3169; Balanced
  3867 / paladin 2826 / bear 2500 / 1794; Max survival 2151 / paladin 1922 / bear 233 / -1994.
- Open: -20261105-1 (paladin potion action, high), -20261104-2 (paladin damage taken, user calls), -20261104-5
  (legacy warrior_protection.json). None fixed on this master.

## Not done this pass

No re-sim (no preset input changed). The Anniversary-log comparison is pass 4's.
