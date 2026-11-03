# Round 70753: the November 2 development notes

The full pass for client build 1.60.1.70753 and Blizzard's development notes post 9 (2026-11-02,
`orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt`), started 2026-11-03.

- Phase 1 (local agents): A1 ports the client tables, talent trees (all classes, `orchestration/talent_remap_70753.md`)
  and DB; A2 the shared sim/core changes (racials, race base stats, Throw, the low-rank proc penalty); C the
  hand-written bayovui parts (BuffCoordinator, BayovCore Threat and Training, the shaman pack, BayovRotation).
- Phase 2 (one cloud session per class): `briefs/<class>.md`. Each writes its ledger and final report to `<class>.md`
  here and lands on master.
- Phase 3: the whole audit suite on master (dps-report B5), the fixes, the report's regeneration (B7, B9), deploy and
  state (B11); then bayovui's generated data (update-data: weights, leveling weights, BiS lists).
