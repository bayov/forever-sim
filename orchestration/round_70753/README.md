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

## Phase 3 to-do (shared, the orchestrator's)

- Drop Blessing of Salvation from the raid buffs: tools/rotopt/mksettings/main.go:74 and every Level 60 request's
  raid.buffs.salvation (threat only, 0 DPS; removed in game since the 10-30 hotfix). Do it with the phase 3 re-sim so
  the request hashes move once.
- forever-wiki/sim-status.md still has the old warrior strings (warrior worker); check every class's rows there.
- ui/enhancement_shaman/leveling_weights.json (the XP/h parts for Enhancement levels 5-59 and the no-spec shaman) was
  measured on 11-02, before this round: measure it again (leveling-opt), then bayovui's weights.py (phase 3b's report).
- Two-hand Fury is now the highest warrior build at 180 s, so the warrior worker moved the class default (isDefault and
  the ui defaults) to it. Listed as a question for the user in round_70753/warrior.md.

## Open questions for the user (from phase 1)

- Skyborne base stats: the notes say both Skyborne races got the base stat fix but give no numbers. The sim keeps the
  old offset (-1 Str, +1 Agi, -1 Sta, +1 Int, 0 Spi). A naked Skyborne character sheet next to a Human of the same
  class at launch settles it.
- Gnome Intellect offset +3 (level 1 offset; the 1.12 level 60 rows' +5 to +9 read as Expansive Mind baked in).
- Human Spirit below 60: the 1.12 rows may already include the old 5%, and the sim applies the new 10% on top
  (leveling only, older than this round).
- Assumptions (B9): shaman: drop `sham_lb_31plus` and `sham_es7_tuning` (now 1.0), move `sham_mw_ppm` to 20, Chain Lightning coefficient unannounced (kept 0.571); reword `pal_2h_spec_holy` (Two-Handed Weapon Specialization now covers all Holy damage; paladin worker); collapse `mage_ice_lance_cd` to one value (client 70753 gives every Ice Lance rank the cooldown;
  mage worker); `hunter_tsa_hm_rank` says Trueshot Aura 50, now 100 by hotfix 113534 (hunter worker).
- `tools/audit/gear/lin2.py` failed on the healer and tank tables of gear_targets.md: fixed (it skips rows that are not stat targets).
