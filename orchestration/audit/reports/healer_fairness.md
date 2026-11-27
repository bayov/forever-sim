# Healer fairness pass 36 (2026-11-27, master f4c45c3f9)

Result: quiet (no new finding, none resolved). Still open: -20261124-1 (raid filler tracks the DPS report; ruled),
-20261115-1 (raid druid not re-tuned for the 0.066 Wild Growth), -20261104-4 (no healing by class in the boss files).

Due for watched files in 5ab589d85..f4c45c3f9:

| File | Change | Reaches a healer? |
|---|---|---|
| `orchestration/standards.md` | standards 11 rogue filler talents ruling; standards 40 Shield Slam stays 2x block value | no (rogue and warrior tank rules) |
| `sim/common/item_effects.go` | Talisman of Ascendance comment only (6 casts from the tooltip, 71139 tables) | no (comment; item 237283 in no healer request, pass 35) |
| `sim/common/talisman_of_ascendance_test.go` | test text | no |

No file under `tools/healbench/`, the healer `ui/` presets, the healer class packages or the core heal/mana/raid-damage
files changed, and the raid filler `tools/dps_report/req/mage__frost.json` is byte-identical (git diff --quiet). So no
re-bench: pass 35's 63 cells (`ref/healer_fairness/bench_5ab589d.tsv`) stand on this master.

Same terms (step 1): unchanged from pass 35; every request wears its preset's gear set, consumables, buffs, race,
professions, level, heal assignment and reaction settings as before.

Ranking (pass 35 numbers): tank Disc 1.234, shaman 1.206, Holy 1.200, druid 1.188, paladin 1.106; raid Holy 1.440,
shaman 1.346, paladin 1.343, druid 1.340. Spread: tank 11.5% first to last, raid Holy 7% ahead of a three-way tie inside
noise. Nothing past 15%.

Roles (step 3): unchanged. Logs (step 2): still blocked by -20261104-4.
