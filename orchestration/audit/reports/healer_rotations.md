# Healer rotations pass 16 (quiet; healer_rotations-20261115-1 and -20261104-2 still open)

- Master: 2f7e255af (`last.master` was 5f20dfdf4). Audited 2026-11-18.
- Bench: `tools/healbench` built from this checkout (Go protos generated with `make sim/core/proto/api.pb.go`
  first), default seed, `-carry`, 1,000 iterations, via `tools/audit/healer_rotations/baseline.sh`.
- The run was cut short by the coordinator after 23 of 63 rows (all `druid_restoration`, all
  `druid_restoration_raid`, and the first 9 `paladin_holy` rows). The box was shared with other auditors and the
  full sweep was on course for about an hour.

## What changed since 5f20dfdf4

- `sim/druid/healing.go`, `sim/druid/items_restoration.go`: Forever's Idol of Rejuvenation (22398), Regrowth's
  direct heal +5% on a target with our Rejuvenation. Gated on the idol; neither druid request wears it (relic
  279250).
- `sim/shaman/items.go`: Totem of Tides (22345), -5% Riptide cost. Neither shaman request wears it (relic 249398).
- `sim/common/item_effects.go`: Darkmoon Card: Blue Dragon (3% of casts) and Heroism (158-238). Worn by no
  healer request (grep of `tools/healbench/req/*.json` for 19288/19287 ids: none).
- `sim/priest/vampiric_embrace.go`: heals 10% (caster 20%) of the priest's Shadow damage. No healer priest
  APL casts Shadow spells or Vampiric Embrace.
- `sim/paladin/holy_strike.go`, `sim/core/flags.go`: Shieldsworn Holy Strike no longer rolls Windfury Totem
  (a melee spec's matter; healbench has no Windfury Totem on the holy paladin).
- No healer APL, healbench request, encounter file or `tools/healbench/main.go` changed.

## Baseline

The 23 rows re-simmed are byte-identical to `ref/healer_bench/summary.tsv` (e.g. `druid_restoration_raid` Onyxia
0.944, 7.1% other deaths, as in the reference). The other 40 rows were not re-run; none of this window's changes
reaches their kits, so the reference stays as it is.

## Changes tried

None: no rotation or bench input moved, and both open findings stand as measured on 5f20dfdf4.

## Open items

- Next pass: run the full 63-row sweep (start it early; about 60 minutes on a shared 4-core box).
