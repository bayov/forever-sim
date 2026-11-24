# Consumables pass 104 (master e29cfe00f, client 1.60.1.71037)

Result: quiet, no findings.

Moved since e7062a6a4: ui/enhancement_shaman/gear_sets/level60_phase0.gear.json only (Phase 0 BiS rerun, commits f3c71bce0, be03bcf67, 5e556ff1e). No newer 1.60 build on wago.

| Check | Result |
|---|---|
| Enchants | 2543 Arcanum of Rapidity now on head and legs (item 18329, +1% haste, head/legs kit in enchant_overrides.go); 1506 dropped from legs. Source doubtful, but standards 42 says use it. |
| Trinkets | Chained Essence of Eranikus 10455 and Adaptive Combat Assistant 272437 (ilvl 65 trinket, in update_ledger SHR-U-21). Blackhand's Breadth 13965 removed. Real items are intended in Phase 0 files (standards 42), so scan.sh NONSYN is expected for level60_phase0. |
| scan.sh | banned, Revelation and crit-elixir sections empty; counts unchanged (minor-armor count prints 3, not the expected 2, pre-existing). |
| Branches | Warrior branch shows ahead only with stale copies of unrelated files; not audited. |

Note: scan.sh's non-synthetic-trinket check globs level60* and so flags phase0 files; a pass tweak could exclude phase0.
No open items beyond the two low findings already in the findings file.
