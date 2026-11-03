# Tank work handoff (paused 2026-11-03 13:05 UTC)

The tank workers (`orchestration/brief_tanks.md`) were stopped by the account's weekly usage limit, and then the
user paused the tank work while the wow-update session brings the sim to a big game update. The user: "You will
need to pick up where u left off afterwards, and fix what you already did to align to newest game update. But,
you'll wait for wow-update to complete."

## Where it stands

This branch (`claude/clever-galileo-qim1ee`, not master) holds the four class workers' work so far, merged, with
the item database regenerated from all four synthetic sets (`go run ./tools/database/gen_db -outDir=assets -gen=db`,
without `with_db`: the half-merged database can't load the new item sets). It builds. Each worker's branch is kept
too (`worktree-agent-*`, local to the container).

| Worker | Done | Left (brief's steps) |
|---|---|---|
| warrior | 1-4: Forever threat for Sunder, Heroic Strike, Cleave and Demoralizing Shout, Bloodrage's health cost, the tank keeps its own Sunders, a test per tank ability and talent, the rotopt template; 5: synthetic set; 6 started (two preset APLs) | finish 6 (searches, Max survival), 7, `docs/tank_warrior.md` |
| paladin | 1-3: Seal of Fury, Holy Shield rank 4, Templar's Bulwark, the immunities, tests; 4: template started; 5: synthetic set, tank librams | finish 4 and 6, 7, `docs/tank_paladin.md` |
| druid | 1-3: audit fixes and a test per ability and talent; 4: `druid_tank` template; 5: synthetic set; UI moved to Forever (APLs, presets) | 6, 7, `docs/tank_druid.md` |
| shaman | 1-3: Rockbiter's threat on Stormstrike, Spirit Weapons, tests, an end-to-end tank test, bench request; 4: `shaman_tank` template; 5: synthetic set | 6, 7, `docs/tank_shaman.md` |
| gear | standards 36 and the Tanks rows (on master) | - |

The tank auditors (`orchestration/audit/passes/tank_*.md`) are on master, off until the tank work merges.

## When the update is in

1. Merge `origin/master` (the update) into this branch, regenerate the protos and the database, and run every
   test (`go test -tags=with_db -timeout 40m` over `./sim/...` without `sim/practice` and `sim/web`).
2. Realign every tank change to the new client build: re-check each worker's audit tables (`docs/tank_<class>.md`,
   the rules audit in `docs/tank_sim.md`) against the new build's rows (`tools/sources/wago.py latest`, the
   update's ledger entries in `orchestration/update_ledger.md`), and the bear (`sim/druid/bear.go`), the encounter
   files' calibration (`tankbench -boss-stats`) and the Rage rules against what the update changed.
3. Restart the four class workers on this branch's state with the brief, steps 6 and 7 plus the realignment, then
   merge, run the tank auditors once, fix their findings, and give the user every call at the end (docs' "Calls for
   the user" sections; the orchestrator's: Ragnaros out of the scoring set, the scoring set itself, standards 36's
   five calls, the bear's threat assumptions).
