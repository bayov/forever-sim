# Phase 0 gear for one tank or healer spec (2026-11-25)

The parent session (brief `briefs/phase0_tankheal_1125.md`) moves the tank and healer reports onto Phase 0
gear. You do the specs your session prompt names (keys of `tools/rotopt/phase0_specs.json`: a tank class's
three presets, or one healer). The parent publishes, audits and deploys.

Work on your own branch from `origin/phase0-tankheal` (named in your prompt). Don't merge to master. Merge
`origin/phase0-tankheal` often (the parent adds prices and fixes there), keep both sides on a conflict.

Read first: the "Tanks and healers" block and the rest of the header of `tools/rotopt/bis_phase0.sh`,
`briefs/phase0_step3_1122.md` (how a search runs and is checked), `briefs/phase0_step4_1124.md` (set trap,
faction rows). Use subagents for big reads and keep your own context small (CLAUDE.md).

## Setup

The cloud container has no Go protos: `apt-get install -y protobuf-compiler`, then
`go install google.golang.org/protobuf/cmd/protoc-gen-go@<version in go.mod>`, PATH with `$(go env GOPATH)/bin`,
`make sim/core/proto/api.pb.go`. `nproc` is likely 4: run one search at a time, in the background, output in
a file, and wait for its notification (no polling).

## Per spec

1. `DRY_RUN=1 SPEC=<spec> tools/rotopt/bis_phase0.sh`, look at the new pool file's unmodelled and hidden
   stat items in your slots.
2. Full search: `SPEC=<spec> tools/rotopt/bis_phase0.sh > <file> 2>&1` (background). A tank search takes
   about 1 to 2 hours, a healer search several. It scores the report's own score (tank: the preset on six
   bosses, healer: damage carried in both roles, docs/healing_sim.md, "The rotopt objective").
3. Unlisted picks or top 5s fail the run: check each (source, cost, open before the raids, by
   `tools/rotopt/phase0.go` and the step 2 notes `orchestration/round_notes_2026-11-11/phase0_step2_*.md`)
   and tag it in `phase0_content.json` or `phase0_enchants.json` (shared files, other workers edit them:
   small additions, keep both sides). Rerun with `RANK_ONLY=1 SLOTS=...` when a tag changes a pick. Never
   ship with `ALLOW_UNLISTED=1`.
4. Set trap (standards 43): a set bonus that needs several pieces at once. Search again from a start at the
   full set and at its bonus points (TALENTS/start via a temporary spec row or FROM list) for the spec's
   dungeon sets (e.g. Valor, Lightforge, Wildheart, Beaststalker or The Elements for Warden, Devout, the Five
   Thunders chain where open). Keep what wins.
5. Hand check: each pick's source and availability, flags (unmodelled effect, hidden stats, doubtful tag),
   and a comparison with a public pre-raid tank or healer list if the web-fetch agent reaches one (never
   bypass Wowhead's bot check).
6. Gear set: the full run writes `gearOut`. Put it into the requests: every request of the spec gets
   `.raid.parties[0].players[0].equipment = {items: <the set's items>}` (tanks: the preset's own
   `tools/tankbench/req/<preset>.json`; healers: both `tools/healbench/req/<spec>.json` and `<spec>_raid.json`).
   Don't touch `tools/tankbench/req/warrior_protection.json` (the healers' bench tank) or the mage filler.
7. Re-check what depends on the gear: tanks' talents and rotation at the new defense (crit immunity at 440
   with talents), hit, rage or mana; healers' rotation and mana (they rarely go OOM on this bench). Use the
   class's tools (`tools/rotopt/<class>_tank.go`, `docs/tank_*.md`, `docs/healing_sim.md`). Change talents or
   APLs only where the sim says so, and RANK_ONLY the list again if they change.
8. The other faction's row: the race whose faction differs from the preset's (`tools/dps_report/run_build.py`
   CLASS_RACES). `RACE=<proto race> RANK_ONLY=1 SLOTS=<the slots with faction rewards> FROM=<list> OUT=<other list>`,
   then `GEAR_ONLY=1 FROM=<other list> GEAR_OUT=<gearOut minus .gear.json>_<horde|alliance>.gear.json`. The
   parent wires bench_races to wear it. If no pick is a faction reward, say so and skip.
9. Score: the request's score with the report's bench at its settings (tank: `go run -tags=with_db ./tools/tankbench
   -iters 1000 -encounter tools/encounters/<boss>.json <req>` on the six scored bosses, mean of the preset's
   column; healer: `tools/healbench -carry -role tank|raid -iters 1000 -encounter ...` on the seven bosses), on
   the Phase 0 set and on the synthetic set (git show origin/master:<req>), with the gap.

Don't touch `bis_phase0.sh`, `tools/rotopt/*.go`, `phase0_prices.py`, `phase0_specs.json` (ask the parent),
the report (`tools/dps_report`, `ui/dps_report`) or other specs' files. Don't run `phase0_prices.py` (the
parent prices every list once the price worker's merge is on master).

## Hand back

Write `orchestration/round_notes_2026-11-11/phase0_tankheal_<group>.md`: per spec the picks per slot with
source, flags, set trap result, talent or rotation changes, the faction row, and the score on Phase 0 against
synthetic. Commit to your branch (never force-push), push it. Send the parent (`@parent`) one message under
250 words: branch and commit, per spec the score Phase 0 vs synthetic, notable picks, anything only the user
can rule on. Then stop. Style in anything the user reads: concise, no em dashes or semicolons, "180s".
