# Healer fairness pass 19 (2026-11-18, master c198d81e3)

Result: quiet. No new findings. Still open: -20261115-1 (high, raid druid not re-tuned for the 0.066 Wild Growth),
-20261104-4 (low).

Due because master changed 2 watched files since 2f7e255af: `sim/priest/priest.go`, `sim/priest/talents.go`
(ac0217aed). The change: `Priest.AddPartyBuffs` now clears `PartyBuffs.dark_inspiration` (the given form,
`core.DarkInspirationUptime`) when the priest has the Dark Inspiration talent, which already gives the party the real aura;
`talents.go` only gains a comment. What could reach a healer:
- Dark Inspiration is a Shadow talent (`proto/priest.proto:58`). The healer priest presets' talent strings have no Shadow
  points (Disc `005005031330011531-0350510303`, Holy `00500503113-03505103030121531`, Holy raid
  `00500503113-33505100030121531`), so `AddPartyBuffs` is a no-op for them.
- No healbench request sets the party buff (`grep -i darkInspiration tools/healbench/req/*.json`: none).
- `tools/healbench/`, `tools/encounters/`, healer requests: unchanged.

## 1. Same terms

Requests unchanged, so pass 11's check holds: everything but the healer row is identical in all 9 requests.

## 2. The numbers (healbench -carry, 1,000 iterations, seed 101)

Healbench built from c198d81e3 (`go build -tags with_db`). 35 cells re-run with `tools/audit/healer_fairness/screen.sh`:
the priest tank presets (Disc, Holy) and the raid priest, paladin and shaman (the 21 raid cells pass 18 left).
Carry and EHPS identical to `ref/healer_fairness/bench_25ce81d.tsv` in all 35; with pass 18's 39 cells, every one of the
63 is now confirmed on this code. Table unchanged:

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
