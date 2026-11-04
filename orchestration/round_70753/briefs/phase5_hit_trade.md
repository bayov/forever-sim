# Round 70753, phase 5: value hit talents against gear that trades hit away

User request (2026-11-04): "For every spec that can take hit%, have alternative synthetic gear pieces that can trade
hit for other stats so we gauge the value of hit% talents appropriately." Also: sweep every report build for hit over
or under its cap, and fix Enhancement's.

Read `orchestration/round_70753/README.md`, `orchestration/standards.md` (items 1b, 4 and 5b, and source tiers 37 to
39), the synthetic gear files (`tools/database/forever_synthetic_items*.go`), `tools/dps_report/race_gear.py` (it already
re-spends a Dwarf's excess hit), `tools/rotopt` (the talent search; `reprice.go`, `gear.go`) and the `dps-report` and
`regenerate-dps-report` skills.

## The problem

Every set carries gear hit to the spec's cap (standards 4), but the talent search (`rotopt`) scores talents on fixed
gear. A hit talent taken on top of capped gear only shows what the hit does past the cap, and nothing ever scores
"talent + gear with that much less hit". Example: Enhancement Level 60 (`353003102-055030110051012251-02`) has 80 gear
hit rating (8%, the melee cap) plus Tidal Focus 2/5 (+2% melee and spell hit, `sim/shaman/talents.go:82-90`), so 10%,
with 20 rating doing nothing for melee and about 4 DPS per 1% for spells, where 10 crit rating is about 5.5
(`forever-wiki/sim-status.md:588`).

Hit talents in the sim, from a grep (check for more, including school-only hit such as Shadow Focus, and set bonuses):
shaman Tidal Focus, hunter Surefooted, paladin Precision, rogue Precision, mage Precision, warlock Suppression, druid
Nature's Reach, warrior Furious Precision (off hand only), and the Dwarf racial (`sim/core/racials_forever.go:47`).
Caps against a level 63 boss: 8% yellow melee and ranged, 27% dual-wield white, 16% spell (17% miss, 1% floor).

## What to do

1. **Sweep.** For every build in `ui/dps_report/data/builds.json` (and its meta), compute total hit by kind: gear +
   enchants + talents + set bonuses + racial + buffs and consumes in the request, against each cap that matters to it.
   Use the sim's own stat sheet (the character's final stats) rather than adding by hand. Table: build, gear hit, other
   hit, total, cap, over/under. Write it to `orchestration/round_70753/hit_sweep.md`.
2. **Hit-trade gear.** For every spec that can take a hit talent (or a race with one), add alternative synthetic
   pieces in that class's file and id range that move hit rating, in 1% steps (10 rating), into the spec's best other
   stat at budget prices (standards 1b: hit and crit rating cost the same; keep each piece within budget and the shape
   check). Step down as far as the spec's talents and racial can add. Where a spec sits below its cap by design (for
   example spell power over hit for Balance, standards 4), steps up as well. Which stat takes the freed budget should
   come from the sim (the stat weights or a short search), not a guess; say what you chose per spec. Give each step a
   gear set (for example `<set>_hit-1` ... `_hit-N`) and put the pieces in db.bin.
3. **Search with it.** Make the talent search pick, for each candidate talent build, the gear step that its talent
   (and racial) hit calls for: the step that brings total hit to the cap, or whichever nearby step sims best for specs
   where hit past the yellow cap still pays (dual wielders' white swings, hybrids' spells; standards 4 "hit vs other
   stats by sim"). Fold `race_gear.py`'s Dwarf rule into the same mechanism if that is simple; if not, leave it and say
   why. Do this in the tool, not by hand per build, so later searches keep it.
4. **Re-search and regear.** Re-run the talent search (the mean of 120, 180 and 300 s, as the class workers did; keep
   the pins each build is named for) for every spec where a hit talent is in play or the sweep found hit off its cap.
   Put each preset on its matching step. Re-sim the moved builds at 120/180/300 s and the race tables where the step
   changes per race. Update standards item 4 to describe the steps. Regenerate the stat weights of the moved builds.

Then rewrite `sim/testdata/report_hashes.txt` for the rows that moved, run the full test suite
(`go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`; a fresh clone first needs
`PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`), regenerate the moved builds in the
report, add Changes-tab lines (`tools/dps_report/changes.py check` clean), land on master and deploy
(`make dps-report-deploy`).

## Another session is active

The phase 4 follow-ups session (`briefs/phase4_followups.md`, branch `phase4-followups`) is rebuilding db.bin (the
missing synthetic item 998001) and may publish to the report. You can do items 1 to 3 now, but rebuild db.bin and
publish only on top of its db.bin fix: merge origin/master (or its branch, if it hasn't landed) before you rebuild
db.bin, and again right before you publish. If the report data moved, take master's, hydrate, re-check stale and
publish again.

Rules: never force-push; end every commit message with the trailer lines your session prompt specifies; no model names in
files; long runs in the background or in subagents, and keep your own context small (the repo's CLAUDE.md). The class
re-searches split well: run them in parallel subagents, one per class, each handing back the picked build, its step and
its 120/180/300 s numbers. When done, send one message to `@parent` with `send_message` (claude-code-remote MCP): the
master commit, the deploy, the sweep's over/under builds, every talent pick that changed, the old and new 180 s numbers
of every build that moved, and anything that needs a user ruling. Then stop.
