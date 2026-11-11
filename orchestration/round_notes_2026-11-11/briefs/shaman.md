# Notes round 2026-11-11: the shaman changes (Blizzard's Nov 10 class tuning)

The user wants the shaman changes landed first, before any other class. Read first:
- `orchestration/round_notes_2026-11-11/ledger.md`, its Shaman section (S1 to S9) and the sources it names;
- `orchestration/round_70753/README.md` and `orchestration/round_70753/shaman.md`, including its "Notes-only values";
- `orchestration/standards.md` 37 and 38;
- the `regenerate-dps-report` skill and `tools/dps_report/README.md`.

A fresh clone first needs `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`. The `.pb.go`
files are gitignored, and stale ones silently sim old field numbers.

**Source rule.** S1 to S6 are in Blizzard's Nov 10 hotfix post, but no server row carries them yet. Standards 37 (a user
ruling) says to model an announced value at its announced number. Do that for each one, cite the post, and add each to
shaman.md's "Notes-only values" with the row or log that would confirm it. If the user's combat logs dated after
2026-11-10 16:51 UTC show a Stormstrike hit, a Maelstrom Weapon stack rate or Windfury procs, check the new values
against them (`tools/combat_log`, the `log-audit` skill). A log that disagrees wins: report it.

1. **Sim changes**, Forever only (standards 38), each with a test:
   - S1: Stormstrike deals 100% weapon damage (`sim/shaman/stormstrike.go`).
   - S2: Stormstrike has an 8 sec cooldown. Fix the header comment too.
   - S3: Mental Dexterity gives 33/66/100% of Intellect as attack power (`sim/shaman/talents.go`).
   - S4: Maelstrom Weapon takes 20% off cast time and mana per stack at every rank.
   - S5: Maelstrom Weapon procs at 6/7/8/9/10 PPM by rank, replacing the every-hit rule from bcc5d6cd4. Check the
     tooltip's example: about 60% a swing at 3.6 speed at 5/5.
   - S6: a weapon imbue and a weapon totem work together. Windfury Weapon and Windfury Totem can both proc, and so can
     Flametongue Weapon and Flametongue Totem. Remove the blocks in `sim/core/buffs.go` and
     `sim/core/totem_weapon_buffs.go`, and the APL notes that cite them.
     Check how the two Windfury procs share or don't share a cooldown. Blizzard's post may say. If it doesn't, keep each
     proc's own rule and list it as an assumption (`tools/dps_report/assumptions_def.py`).
   - S7 (Rage of the Storm, hotfix 113886): update the item if the db pull hasn't. It affects only leveling.
2. **Re-optimize Enhancement.** Taken together these are a large nerf, and they change what's best.
   - Re-search the talents, rotation and imbue/totem choice for all four Enhancement builds with rotopt, the way
     shaman.md did. The imbue/totem choice now includes Windfury Weapon together with Windfury Totem.
   - Mental Dexterity's cut makes Intellect worth less. Check whether the Enhancement synthetic gear's stat mix and the
     hit-trade steps (`orchestration/round_70753/hit_trade/`) need re-pricing.
   - Keep what wins on two seed sets at 120, 180 and 300 s.
   - Elemental moves only through S6 (a group member's imbue) or not at all. Confirm that with a sim and leave it otherwise.
3. **Report.**
   - Rewrite the shaman rows of `sim/testdata/report_hashes.txt`.
   - Run the full suite in a subagent: `go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`.
   - Re-sim every shaman build: hydrate, `pipeline.sh run <ids>`, `publish <ids>`.
   - Update the metas' `published.dps`, presets.ts and the wiki.
   - Add the Changes-tab lines. `changes.py check` must list nothing uncovered.
   - Raid builds whose group members carry an imbue move through S6, so re-sim the ones that do.
   - Merge origin/master right before you publish, since other sessions push to master. Then land on master, deploy
     (`make dps-report-deploy`; the proxy authenticates api.cloudflare.com) and update both state.json files (B11).
   - Leave the leveling weights, the stat weights in bayovui and the leveling rotations alone (user ruling). List what
     they would need.

Rules:
- Never force-push.
- End every commit message with the trailer lines your session prompt specifies.
- Keep model names out of files.
- Keep long runs in the background or in subagents, so your own context stays small (the repo's CLAUDE.md).

When done:
- Write `orchestration/round_notes_2026-11-11/shaman.md` with:
  - each change and how it was confirmed (notes only, row or log);
  - every shaman build's old and new 180 s number;
  - the new talents and rotation, and anything that needs a user ruling.
- Push it.
- Send one message to `@parent` with `send_message` (claude-code-remote MCP), with the master commit, the deploy, the
  numbers and the rulings needed. Then stop.
