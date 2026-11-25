# Fairness pass 58 (2026-11-25, master 5536db4bc)

Quiet, no new finding. Range 226ada86b..5536db4bc: the only watched file is sim/compsearch/drops.json (86fbcf456,
regenerated after the trinket-line fix). It is the raid sim composition search's cache-normalization table (which
effects a build cannot read, from bit-identical sims, guarded by TestDropsAreUnread); no DPS report request, preset,
APL, gear set or sim/core code reads it, so no Level 60 DPS number or ranking can move. The other commits touch the
report's assumption data, a brief and a healer ruling, all outside this pass's paths. Not re-swept; baseline stays
63129ce35. Open findings unchanged, fairness-20261125-1 (MM report race Orc trails Skyborne High Order by 6.1) among them.
