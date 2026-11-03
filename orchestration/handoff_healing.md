# Healing sim handoff (2026-11-02)

The healing sim session (session_01TXnjyQ4KhQuoE7tNiuXLrW) handed off here. Read `docs/healing_sim.md` first:
it holds the user's rulings, the design, what's built, the bench's numbers and the milestones. This file is
what that session knew and the doc doesn't, and the order of the work left.

## State at handoff

- **Master:** 7efa6d02a (the bound ranks commit e41f082eb, merged). Pushed to `claude/quirky-ptolemy-tdezqg`
  too. Nothing uncommitted.
- **Done, in order:** H1 (groundwork: overheal, health and death, raid damage events, heal targets and
  assignments, Chain Heal jumps, `tools/healbench`), the score (damage carried per role, death floor, the
  per-fight death metrics), what healers see (heals in flight, reaction time, cancelling), bound ranks
  (at most 4 ranks of a heal, picked by Expected Healing).
- **Tests:** `go test -tags=with_db ./sim/ ./sim/core/... ./sim/shaman/...` passes, `npx tsc --noEmit -p .`
  is clean. `tools/raid_sim_data`'s `TestDataFileIsFresh` was already failing on master from another session's
  change (a stale `builds.json`). Not ours.
- **Warcraft Logs credentials are in the environment** (`WCL_CLIENT_ID`, `WCL_CLIENT_SECRET`), and
  `tools/warcraftlogs/wclcli.py` reads them. Nothing has been mined for healing yet.

## The work left, in order

The user approved this plan's first steps one at a time ("go"). Bound ranks were done to the user's own
ruling. Report each step's results before starting the next unless the user says to keep going.

1. **Done 2026-11-02** (session_01TXnjyPVS2mPceSNTempQD9, `docs/healing_sim.md`, Raid damage from the logs).
   **Raid damage from Warcraft Logs** for the 7 shared boss files (`tools/encounters/*.json`: Onyxia,
   Broodlord, Chromaggus, Golemagg, Magmadar, Ragnaros, Vaelastrasz). They have the tank side only, from
   vmangos (the lowest source tier). Ours to add: the boss's damage on the rest of the raid as `raid_damage`
   events (`who` NotTanks or Everyone, `count`, `every_seconds`, `spread_seconds`, ticks), the real fight
   lengths, and the healer counts, healing and deaths real raids had (the user's "check against
   warcraftlogs" for the raid sim's 20% margin). Sources best first: SoD logs, then Anniversary (Vanilla,
   then TBC), then Era, vmangos last (`docs/data_sources.md`). Extend `wclcli.py` with a raid command next to
   its `tank` command (damage taken by non-tanks per ability, per fight; fight duration; healer count;
   deaths), keep within its point budget (3600 an hour, the cache in `~/.cache/forever-sim/wcl/` dies with the
   container). Record every number's source and report codes in the file's `source` and `reports`. Then the
   healbench numbers again on `-encounter` with `-role raid` and `both`.
2. **Done 2026-11-02**, its synthetic gear 2026-11-03 (`docs/healing_sim.md`, H2).
   **H2, the Restoration shaman as a real spec:** register the agent (today it's only the bench's request
   `tools/healbench/req/shaman_restoration.json` on the elemental agent), gear, rotations and presets per role
   (tank healer, raid healer), Water Shield, Mana Spring by group, Mana Tide. Its spells and talents are in.
3. **H3, the other healers:** Holy paladin, then Restoration druid, then priest (needs an absorb pipeline).
   Each with a rules audit against Forever's client data, like the DPS class audits. Each heal with a cast
   time needs `SpellConfig.ExpectedHealing` for heals in flight and Expected Healing to work (the shaman's
   three heals have it, `sim/shaman/healing.go`).
4. **Healer-aware Innervate, Mana Tide and Power Infusion** in the raid sim.
5. **H5, the DPS Report's healer section.** Deploy after (`make dps-report-deploy`, CLAUDE.md).
6. **H6, the raid sim:** simmed tanks and healers, the comp search counting healing, then choosing how many
   healers (the 20% margin).

Open user decisions, with the defaults proposed and not yet answered: the scare line (20% of max health,
reported only, not scored) and Power Word: Shield absorbs (effective while absorbing, the unused part
counted as overheal). Open audit gaps (`docs/healing_sim.md`, The audit): aura values on Most Hurt and Lowest
Health, range and groups, Hateful Strike's targeting, players' self-saves, Judgement of Light a no-op,
Forever's Spirit regen unverified, per-healer deaths within its assignment.

## Coordination with the tank session

The tank sim is session_01TXnjyP5nUQs8zySrhUvYoi since 17:54 (session_01TXnjyPCpeSGJhYroNvpQiF before it, archived; `docs/tank_sim.md`), reached with the remote session tools'
`send_message`. Agreed with it:

- The boss files in `tools/encounters/` are shared: the tank side is theirs, the raid damage ours.
- `tools/tankscore` gives `LoadEncounter`, `UseEncounter`, `MakeTank` and `TankScoreSummary`; healbench uses
  them.
- The tank bench keeps a 5% death line for its "healing needed"; healers score at 2% for tanks, 5% for the
  others (the user's ruling).
- Proto field numbers collided once (both sessions add to the same messages). Check `origin/master` before
  picking a number, and renumber yours on a clash.

## Gotchas

- **Protos:** the Go files are untracked. After a `proto/*.proto` change:
  `PATH=$PATH:~/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto` and
  `make ui/core/proto/api.ts`. protoc and protoc-gen-go v1.36.6 may need installing in a fresh container.
- **New result fields:** add them to the multi-threaded combiner (`sim/core/sim_concurrent.go`), or they
  vanish with more than one thread. Then the report hashes (`sim/testdata/report_hashes.txt`) move. First
  check the old hashes still match with the new fields cleared in `canonicalResult`, then rewrite them with
  `RAID_FOREVER_HASHES=testdata/report_hashes.txt go test -tags=with_db -run TestRaidForeverReportHashes ./sim/`.
  A merge conflict on that file is resolved the same way.
- **healbench:** `go run -tags=with_db ./tools/healbench -role tank|raid|both -encounter
  tools/encounters/<boss>.json -carry -iters 2000 <healer request>...` (a file path for `-encounter`, not a
  name). A carry at 500 iterations moves about ±0.01. Several requests give one row each, for comparing
  rotations.
- `go test` of the whole `./sim/...` runs past the Bash tool's 2 minute limit: run it in the background, and
  `tools/compsearch` needs `-timeout 90m`.
- Commits end with the attribution lines from the session's system reminder, and never name a model.
