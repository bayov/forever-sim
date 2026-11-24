# Results pass 100, master 566067a394 (2026-11-24)

Shared run (68 builds, 20k, seed 1) vs builds.json: 61 within 1.5, 7 off, all explained by sim commits since the last re-sim.

| Build | builds.json | run | move | cause (paired 20k seed 1) |
|---|---|---|---|---|
| druid feral-cat | 1185.9 | 1171.2 | -14.8 | be773f5eb Rip/Rake tick AP shares: 1186.2 -> 1171.2 |
| druid feral-cat-duty-ff | 1179.5 | 1164.9 | -14.6 | same, 1179.8 -> 1164.9 |
| warrior fury, duty-bs, duty-sunder | | | -5.5, -4.6, -4.6 | cd71e78c2 Flurry/Reckless Swings/Unbridled Wrath: -6.2, -4.5, -5.1 |
| warrior fury_2h family | | | -1.0 to -1.5 | same commit |
| hunters (9) | | | +0.7 to +1.4 | b82ad9a1e pet summoned with Aspect of the Hawk, within noise |

Arms builds unchanged. Metas still lag (see results-20261104-2). Drift check not re-run: mksettings and requests unchanged.
Result: no new finding; update added to results-20261104-2.
