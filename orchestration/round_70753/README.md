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

## Source rule for this round (standards 37, user ruling 2026-11-03)

- Server data (top tier): the 2026-11-03 hotfix to the 70753 build carries the hunter damage lines of the notes
  (`orchestration/audit/ref/client_update/sources/foreverchanges_2026-11-03_hotfix_and_70753.txt`).
- Client 70753 carries the notes' talent reworks but none of the other classes' damage lines (Shred, Starfire,
  Lightning Bolt and so on stay at 70707 values), and no server hotfix has them yet.
- So for a note line that neither the client nor a hotfix carries, the sim models the **notes' value**, gated under
  Forever, citing the notes and "client 70753 unchanged". Each one is listed in the class's round file under
  "Notes-only values", so they can be checked against the next hotfix or launch logs. Where a hotfix or the client
  does carry a value, it wins.

## bayovui (wow repo)

Hand-written parts landed on wow main 3b7870df (2026-11-03): the shaman pack (Lava Burst out, Totemic Call, Lightning
Mastery), BayovCore Training, CVars and the Threat comments, and a comment in Buff Coordinator (wow 68a4962 had
already removed Salvation, Mass Salvation and Blessing of Light). Left for phase 3b and the user:
- Generated, stale: ClassSpellsData (classspells.py needs QuestieDB on the user's machine and a Wowhead scrape; still
  Binding Heal, Lava Burst, Elemental Mastery, Totemic Projection, no Poised or Shank, old trainer prices),
  LevelingRotationData (Enhancement plan and talent route), the stat and leveling weights, the BiS lists, MacroIconsData.
- Threat modifiers (Bear 115%, Defensive Stance 80%, Righteous Fury 33% + 90% Holy, Defiance, Salvation gone) live
  in forever-sim's tools/combat_log/threat.py and the tank sim, not BayovCore.
- The user: deploy BayovCore and Buff Coordinator to beta and forever, re-export or install the forever shaman and
  bayov auras, delete a leftover "Bayov - Shaman Lava Burst" aura if one stays, check Totemic Call's 30 s cooldown
  in game, re-check the RXP route's training-gold steps against the new trainer costs.
