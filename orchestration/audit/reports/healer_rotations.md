# healer_rotations pass 31

master 34994816e (last: f9fa79917). Quiet: no new findings.

## What changed

The 8 watched files are Phase 0 healer lists: `bis/*.bis.json` for Discipline, Holy priest, Holy paladin and both
Restoration shaman lists, and the `gear_sets/*.gear.json` of Holy paladin and both shaman sets (commits 287da91da,
6bd2d60a2: Armaments of Battle rewards are out, so e.g. 22689/22678/22681/20717 swap to 14140/11832/272438/16058/18527).
The healbench requests embed their own gear and did not change, so these lists don't reach the bench (healer_gear
covers whether the requests should follow them). The only sim change in the range that a bench row could see is
sim/common/item_effects.go (5591bd019: Cloud of Chained Essence ticks crit at a fixed 2.7%). The Frost mage filler
(tools/dps_report/req/mage__frost.json) wears that trinket, but every raid row stayed byte-identical. No healer APL,
healbench request, encounter file, healbench code or filler request changed.

## Baseline (tools/audit/healer_rotations/baseline.sh, 1,000 iterations, -carry, own role, boss lengths)

All 63 rows are byte-identical to ref/healer_bench/summary.tsv (cmp), so the means of pass 29 stand:

| request | mean carry |
|---|---|
| priest_holy (tank) | 1.194 |
| priest_holy_raid | 1.440 |
| priest_discipline (tank) | 1.228 |
| paladin_holy (tank) | 1.100 |
| paladin_holy_raid | 1.343 |
| shaman_restoration (tank) | 1.201 |
| shaman_restoration_raid | 1.348 |
| druid_restoration (tank) | 1.255 |
| druid_restoration_raid | 1.351 |

Mana: unchanged (Holy and Discipline tank still OOM on Chromaggus, open -20261125-1, -20261104-2).
The ref was not rewritten, since it is identical.

## Changes tried

None this pass: no rotation or bench input changed, and the open findings' one-line variants were measured in
earlier passes (see findings/healer_rotations.md).

## Setup note

A fresh session needs `make sim/core/proto/api.pb.go` (after `go install .../protoc-gen-go`) before the healbench
build; tools/audit/setup.sh installs the tools but doesn't generate the Go protos.

No new findings. No open finding was fixed on master.
