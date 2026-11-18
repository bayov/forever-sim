# Haste shortens a spell's GCD (sim fix and report rerun), 2026-11-18

A combat log settled notes.md's open item "Beta: does haste shorten the GCD?". The answer is yes, for spells.

## The log evidence (user's Enhancement shaman, Forever beta)

- Warchief's Blessing is spell 16609, +15% haste. It speeds casts too: Lightning Bolt casts in 1.505s without it, 1.30s
  with it.
- With no haste, back-to-back instants sit a median 1.52s apart.
- With the buff up, after a magic spell, 384 of 1,000 pairs sit at 1.27–1.33s (1.5 / 1.15 = 1.304s). The shortest gap
  is 1.23s.
- After Stormstrike the shortest gap is 1.40s, so physical abilities keep their 1.5s GCD.
- Not tested: a floor, the totems' 1s GCD, other haste sources, a rogue.

## What to change

1. `ApplyWarchiefsBlessing` (`sim/core/buffs.go` ~2300) only multiplies melee speed by 1.15. Add 15% spell haste
   (cast speed) too.
2. A spell's GCD becomes 1.5 / (1 + spell haste). Physical abilities (melee and ranged, Stormstrike's kind) keep 1.5s.
   Decide "spell" the way the client does as far as the sim can tell (magic defense type, not melee or ranged), and say
   in notes.md which edge cases the log didn't cover (a hunter's magic-school shots, Judgement, Seals). Use a 1.0s
   floor, like wowsims/forever, unless the code shows a reason not to, and list the floor under "# Need to Verify" in
   notes.md.
3. Check the class GCD rules still behave and test each:
   - rogue and cat form's 1.0s;
   - the totems' 1s GCD (the floor should hold it at 1s);
   - Nature's Grace's 10% GCD cut (spell 16886), alongside haste.

   Add a core test for the formula, the floor and a physical ability.
4. Patch `tools/dps_report/patches/0004` has a `haste_gcd` knob for the report's assumption. The sim now does this
   natively, so remove the knob from the patch the way b77ffa52a removed AssumeWindfuryPreAttack. Patch changes flag
   every build, so record them with `stale.mjs --record` where the patch change alone moves nothing.
5. Update notes.md:59 ("Beta: does haste shorten the GCD?"). Close it with the log evidence, keeping only what's still
   open (the floor, other haste sources, the edge cases). Update the `haste_gcd` assumption in
   `tools/dps_report/assumptions_def.py`: drop it if no preset's number depends on anything the log left open (as with
   warr_exec_wf2), or reword it to the open part. Update the wowsims comparison notes that list haste_gcd as a gap.
6. Rerun the DPS Report. Casters, Enhancement and healers move wherever Warchief's Blessing is on (a default buff).

## Coordination

- **publish_1118** (`session_01TXnjyR5482fZFFMyQWVPAQ`) is re-simming Healing and will publish and deploy. Do the code,
  tests and notes now, but don't run `pipeline.sh run`, `publish` or a deploy until it has pushed master and deployed
  (its deployId lands in `.agents/skills/dps-report/state.json` on master). Then merge origin/master and go. If it
  hasn't landed by 12:00 UTC, send it a message asking where it is.
- The wowsims comparison session (`session_01TXnjyQoHMJg2CjHPYTGUMX`) owns `sim/warrior` and some hunter fixes. Stay off
  those files. Merge whatever it lands on master before you publish.

## Steps

Read first: `orchestration/round_notes_2026-11-11/briefs/update_1117.md` (the same landing steps), `update_1117.md`,
`tools/dps_report/README.md` and the `regenerate-dps-report` skill.

A fresh clone needs:
- `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`;
- `make ui/core/proto/api.ts`;
- `DASH_TREE=<your checkout>` for `pipeline.sh`.

Rebuild `tools/dps_report/bin` from HEAD before any probe.

1. The code fix, with tests. Run the full suite in a subagent
   (`go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`). Regenerate goldens only for the
   intended moves, and say which failures were there before.
2. Scope check: `pipeline.sh hydrate`, `node tools/dps_report/stale.mjs`, `pipeline.sh probe`, and the Tanking and
   Healing staleness check. Name each moved build's cause, and report anything that moves but shouldn't.
3. Re-optimize only a build that moves more than 1% from this rule: rotopt, two seed sets at 120, 180 and 300s, the way
   update_1115 did. Paladins and Enhancement carry GCD_ROT_WARN, so look at them first.
4. Land and publish as update_1117's step 6 says: report_hashes, `pipeline.sh run`, `publish`, weights for the moved
   builds, Changes-tab lines with `changes.py check` clean, metas, presets.ts and the wiki where rotations changed,
   `tankheal.py run heal` (and tank if it moves). Merge origin/master before you publish and again before you push.
   Push master, then `make dps-report-deploy` (on Cloudflare error 8000009, re-run
   `python3 tools/dps_report/deploy_pages.py`), update `.agents/skills/dps-report/state.json`, and push the final master
   to your own development branch too.

Rules: never force-push, end every commit message with your session prompt's trailer lines, keep model names out of
files, and keep long runs in subagents or in the background. Leave bayovui, the leveling weights and the leveling
rotations alone. User style in anything they read: concise, no em dash or semicolon, "180s" not "180 s".

## When done

Write `orchestration/round_notes_2026-11-11/haste_gcd_1118.md` with:
- the rule as coded, its tests and the edge cases left open;
- each moved build's old and new 180s number;
- the Tanking and Healing moves;
- the deploy id;
- anything that needs a user ruling.

Push it, send one message to `@parent` with the master commit, the deploy, the numbers and the rulings needed, then stop.
