# Healer fairness pass 34 (2026-11-27, master 34994816e)

Result: findings (1 new: healer_fairness-20261127-1, medium). Still open: -20261126-1 (druid on its synthetic set, ranked
first tank healer without a note), -20261124-1 (raid filler tracks the DPS report; ruled), -20261115-1 (raid druid not
re-tuned for the 0.066 Wild Growth), -20261104-4 (no healing by class in the boss files).

Due for 8 watched files in f9fa79917..34994816e, all from 6bd2d60a2 ("Phase 0 lists without the Armaments of Battle
rewards"): the priest, paladin and shaman bis lists and the paladin and shaman (Horde and Alliance) gear sets. The only
sim change in the range (`sim/common/item_effects.go`, Cloud of Chained Essence's tick crit) reaches no healer preset.

Same terms (step 1): no healbench request changed. The paladin and shaman requests' equipment still equals the
f9fa79917 gear sets item for item, so they wear what the gear sets no longer do:

| Request | Slot | Request wears (removed) | Gear set now |
|---|---|---|---|
| paladin_holy, paladin_holy_raid | head | 22689, enchant 1505 | 14140, enchant 1503 |
| paladin_holy, paladin_holy_raid | trinket | 22678 | 272438 (the other trinket 11832 stays) |
| shaman_restoration, shaman_restoration_raid | finger | 22681 (Band of Piety) | 16058 |

The priest changes are in the bis lists only (their gear sets and requests unchanged). Hence -20261127-1. The same stale
ids are in 15 DPS report requests (Balance, Elemental, Affliction, Demonology, Destruction); not this pass's scope.

Numbers (step 2): all 63 cells re-benched (healbench `-tags with_db` in the build_cli.sh worktree of 34994816e,
`tools/audit/healer_fairness/screen.sh`, seed 101, 1,000 it; `ref/healer_fairness/bench_3499481.tsv`): identical to
pass 33's `bench_f9fa799.tsv` in all 63 cells. The four stale requests with their items swapped for the current gear
sets (`ref/healer_fairness/reqsync_3499481.tsv`):

| Request (role) | Mean as shipped | Mean on current gear set | Change | Largest cell |
|---|---|---|---|---|
| paladin_holy (tank) | 1.100 | 1.106 | +0.006 | golemagg +0.028 |
| shaman_restoration (tank) | 1.201 | 1.206 | +0.005 | golemagg +0.019 |
| paladin_holy_raid (raid) | 1.343 | 1.343 | 0.000 | golemagg -0.007 / onyxia +0.008 |
| shaman_restoration_raid (raid) | 1.348 | 1.346 | -0.002 | golemagg +0.013 |

No rank changes: tank druid 1.255 (synthetic), Disc 1.228, shaman 1.206, Holy 1.194, paladin 1.106; raid Holy 1.440,
druid 1.351 (synthetic), shaman 1.346, paladin 1.343 (the bottom three still a tie inside noise). Roles (step 3):
unchanged from pass 33. The logs comparison (step 2) stands as in pass 33.
