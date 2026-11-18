# Healer fairness pass 18 (2026-11-18, master 2f7e255af)

Result: quiet. No new findings. Still open: -20261115-1 (high, raid druid not re-tuned for the 0.066 Wild Growth),
-20261104-4 (low).

Due because master changed 5 watched files since 5f20dfdf4: `sim/druid/healing.go`, `sim/druid/items_restoration.go`,
`sim/paladin/holy_strike.go`, `sim/priest/vampiric_embrace.go`, `sim/shaman/items.go` (update_1117 and the Nov 13/16
hotfixes). What could reach a healer:
- Idol of Rejuvenation (22398): Regrowth's direct heal +5% on a target with our Rejuvenation (new `regrowthOnRejuvBonus`
  in `sim/druid/healing.go`). The druid presets wear Idol of Swiftness (279250), so inert for them.
- Totem of Tides (22345): Riptide 5% cheaper. The shaman presets wear Polished Driftwood Icon (249398).
- Darkmoon Card: Blue Dragon 3% on Forever, Heroism heals 158-238; Cannibal's Cleaver, Draconian Aegis: no healer wears them.
- Holy Strike no longer rolls Windfury Totem with Shieldsworn; Vampiric Embrace 10% (caster 20%) of the priest's own Shadow
  damage only: no healer preset casts either.
- `tools/healbench/`, `tools/encounters/`, healer requests: unchanged.

## 1. Same terms

Requests unchanged, so pass 11's check holds: everything but the healer row is identical in all 9 requests.

## 2. The numbers (healbench -carry, 1,000 iterations, seed 101)

Healbench built from 2f7e255af (`go build -tags with_db`). 39 of the 63 cells re-run (all 35 tank cells and 4 of
the raid druid's 7): carry and EHPS identical to
`ref/healer_fairness/bench_25ce81d.tsv` in all 39. The remaining 24 raid cells (paladin, priest, shaman raid) were cut
short when the coordinator asked the pass to finish; no code they run changed. Table unchanged from pass 17:

| Role | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Tank | Holy priest 1.278 | Disc priest 1.272 | Druid 1.269 | Shaman 1.257 | Paladin 1.152 |
| Raid | Holy priest 1.461 | Paladin 1.430 | Shaman 1.412 | Druid 1.392 | |

Spread: tank 1.152-1.278, raid 1.392-1.461 (4.7%). Nothing past the 15% line.

### Relic screen (the new relic effects, swapped into the shipped presets; seed 101, 1,000 it, one seed)

| Variant | Mean carry | vs shipped |
|---|---|---|
| druid tank, Idol of Rejuvenation for Idol of Swiftness | 1.275 | +0.006 |
| druid raid, Idol of Rejuvenation for Idol of Swiftness | 1.392 | 0.000 |
| shaman tank, Totem of Tides for Polished Driftwood Icon | 1.253 | -0.004 |

Neither beats the shipped relic past single-seed noise; not a finding. The druid tank +0.006 could be re-checked on
paired seeds if the druid presets are re-tuned (-20261115-1). Raid shaman with Totem of Tides not finished.

## 3. What a role rewards

Each preset benched only in its own role, each file at its own length. Unchanged.

## Against the logs

Still can't be checked: the boss files keep no healing by class (-20261104-4).

## Open items

- -20261115-1 (high): re-tune the raid druid for the 0.066 Wild Growth; try Wild Growth in the tank preset.
- -20261104-4 (low): HPS by healer class in the boss files.
- Next pass: the 24 raid cells not re-run here, and the raid shaman Totem of Tides screen.
