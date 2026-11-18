# Healer fairness pass 20 (2026-11-18, master 8219f1083)

Result: quiet. No new findings. Still open: -20261115-1 (high, raid druid not re-tuned for the 0.066 Wild Growth),
-20261104-4 (low).

Due because master changed 2 watched files since c198d81e3: `sim/core/dot.go`, `sim/druid/hurricane.go` (8219f1083,
"cast speed hastes Hurricane and Volley ticks"). The change: under Forever, `Dot.hastedTicks` gives a dot with
`AffectedByCastSpeed` a hasted, ms-rounded tick period, and either keeps its tick count (`HasteReducesDuration`, Volley)
or keeps its base duration with the extra ticks that fit (Hurricane). What could reach a healer:
- Only `sim/druid/hurricane.go:76` and `sim/hunter/volley.go:68` set `AffectedByCastSpeed` (grep of `sim/`, non-test).
  No HoT (Renew, Rejuvenation, Regrowth, Wild Growth, Lifebloom-likes) and no healer channel (Tranquility) sets it.
- For every other dot `hastedTicks` returns `(TickLength, 0)`, the old `RecomputeAuraDuration` result, and
  `MaxTicksRemaining` adds `hasteTicks` = 0. No healer preset casts Hurricane.
- `tools/healbench/`, `tools/encounters/`, healer requests: unchanged.

## 1. Same terms

Requests unchanged, so pass 11's check holds: everything but the healer row is identical in all 9 requests.

## 2. The numbers (healbench -carry, 1,000 iterations, seed 101)

Healbench built from 8219f1083 (`go build -tags with_db`, after `make sim/core/proto/api.pb.go`). 28 cells re-run with
`tools/audit/healer_fairness/screen.sh`: the druid in both roles (the class whose code changed), the raid shaman and the
tank Holy priest. Carry and EHPS identical to `ref/healer_fairness/bench_25ce81d.tsv` in all 28. Table unchanged:

| Role | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Tank | Holy priest 1.278 | Disc priest 1.272 | Druid 1.269 | Shaman 1.257 | Paladin 1.152 |
| Raid | Holy priest 1.461 | Paladin 1.430 | Shaman 1.412 | Druid 1.392 | |

Spread: tank 1.152-1.278, raid 1.392-1.461 (4.7%). Nothing past the 15% line.

## 3. What a role rewards

Each preset benched only in its own role, each file at its own length. Unchanged.

## Against the logs

Still can't be checked: the boss files keep no healing by class (-20261104-4).

## Open items

- -20261115-1 (high): re-tune the raid druid for the 0.066 Wild Growth; try Wild Growth in the tank preset.
- -20261104-4 (low): HPS by healer class in the boss files.
- Carried from pass 18: the raid shaman Totem of Tides relic screen (tank variant was -0.004; low priority).
