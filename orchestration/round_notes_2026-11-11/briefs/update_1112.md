# Daily update 2026-11-12: land its findings

The daily update routine (session `session_01TXnjyQKw3sjm6WDB91xdDg`) found these on 2026-11-12. The user wants them
landed. Its branch `forever-update/2026-11-12` (PR #128, two commits) is not merged yet:
- 33b13e310: Judgement rolls as a ranged attack (Blizzard's Nov 11 hotfix, https://us.forums.blizzard.com/en/wow/t/2403040
  post 8). Judgements use the ranged table and no longer trigger melee procs. Unsealed Wrath stores no echo from
  Windfury Totem swings. It adds an extra-attack source to `sim/core/attack.go` and `AutoAttacks.IsWindfuryTotemAttack()`
  in `sim/core/buffs.go`. On its own sims, Ret level_60 goes from 1005.90 to 982.05.
- d787a98ed: the Nov 11 hotfix wands (Touch of the Depths 287004, Satyr's Touch 287036) and Band of the Damned 276073.

Its follow-ups are the three TASKS.md items dated 2026-11-12 (commit 8306f1d52). Read them.

Read first:
- `orchestration/round_70753/README.md` (how a round lands and publishes);
- `orchestration/standards.md` 17 (the Windfury Totem rule), 37 and 38;
- the `regenerate-dps-report` skill and `tools/dps_report/README.md`.

A fresh clone first needs `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`.

1. **Merge the branch** into a branch of your own off origin/master. Master changed Windfury Totem after the branch
   was cut: a0b4d3be9 makes it roll on main-hand abilities (standards 17). Check that the two fit together:
   - `IsWindfuryTotemAttack()` must still recognize every totem extra swing, including the swings from abilities;
   - the Unsealed Wrath gate must still hold;
   - a judgement must not roll the totem. Since the hotfix it's a ranged attack, so it shouldn't.
   Add a test for each point the merge could break.
2. **TASKS item 2026-11-12 "other Nov 11 class hotfixes".**
   - Hunter Volley no longer activates procs: gate weapon and item procs out of Volley's ticks if they can see them now.
   - Rend and Tear on Pounce, Frost Nova, Confounding Flash and Twilight Focus in Shadowform: confirm there's no sim
     effect, and say so in the item.
3. **TASKS item 2026-11-12 "items revealed Nov 11-12".**
   - Rerun `tools/sources/hidden.py`.
   - Compare the listed items' Forever tooltips with db.json. Fix what differs, through the same inputs the update
     routine used (`forever_seen_in_game.json` with a `by` naming the source, or `forever_items.go`).
   - Run the gear search only for builds where a changed item could enter the set. Say which ones you checked.
4. **TASKS item 2026-11-12 "Judgement ranged attack"** leaves points open: weapon specialization, whether JotC and JoW
   roll hit, crit, and whether Windfury swings spend echoes. Keep the branch's choices, and record each as an assumption
   (`tools/dps_report/assumptions_def.py`) if it isn't one already. These are the user's calls once a log settles them.
5. **Re-optimize** the paladin builds the judgement change moves: Ret, its variants, and Prot paladin in Tanking.
   - Seal and judgement choice, talents (Precision and hit talents now matter for ranged hit), and the rotation.
   - Use rotopt, the way earlier rounds did. Keep what wins on two seed sets at 120, 180 and 300 s.
6. **Land and publish.**
   - A tank/heal worker (`session_01TXnjyS9M7qtkpmQ5AysBKq`, branch `tankheal`) is publishing to master right now.
     Its changes include paladin healer and tank presets. **Don't publish until it has landed**: watch for its merge
     on origin/master, and merge origin/master right before you publish.
   - Then follow the README:
     - rewrite `sim/testdata/report_hashes.txt` for what moved;
     - run the full suite in a subagent;
     - hydrate, `pipeline.sh run <ids>` for every build that moved (DPS, Tanking and Healing), then `publish <ids>`;
     - add the Changes-tab lines (`changes.py check` must list nothing uncovered);
     - update the metas, presets.ts and the wiki.
   - Merge into master and push, deploy with `make dps-report-deploy`, and update both state.json files (B11).
   - Tick the TASKS items you finished, and leave the ones a log must settle open.
   - Close PR #128 with a comment that says it landed through master. End the comment with the footer from your system
     prompt.

Rules:
- Never force-push.
- End every commit message with the trailer lines your session prompt specifies.
- Keep model names out of files.
- Keep long runs in the background or in subagents, so your own context stays small (the repo's CLAUDE.md).
- Leave the leveling weights, bayovui's stat weights and the leveling rotations alone (user ruling).

When done:
- Write `orchestration/round_notes_2026-11-11/update_1112.md` with:
  - each change and how it was confirmed;
  - each moved build's old (live) and new 180 s number, with its cause;
  - the Tanking and Healing moves;
  - the deploy id;
  - anything that needs a user ruling.
- Push it.
- Send one message to `@parent` with `send_message` (claude-code-remote MCP), with the master commit, the deploy, the
  numbers and the rulings needed. Then stop.
