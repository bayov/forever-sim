# Daily updates 2026-11-13 and 2026-11-14: land PRs #129 and #130

The user wants the daily update routine's last two PRs landed. Neither is merged yet:
- **PR #129**, branch `forever-update/2026-11-13-70938`, one commit (78c395984), a draft:
  - Wild Growth's coefficient goes from 0.033 to 0.066;
  - Imp Firebolt ranks 2-4 take client 70938's values;
  - Seal and Rune of the Dawn change against Undead;
  - two new items (282027, 286590) and five seen-in-game items;
  - db regenerated.
- **PR #130**, branch `forever-update/2026-11-14`, one commit (9ec69044f), from Blizzard's Nov 13 hotfixes
  (https://us.forums.blizzard.com/en/wow/t/2403040/10):
  - Judgement of Command is a melee attack again, while Righteousness and Fury stay ranged;
  - Windfury Totem rolls on every landed main-hand melee hit, white or special, never on off-hand strikes, extra attacks
    or seal procs (`SpellFlagNoWindfuryTotem`), and Windfury Weapon's main-hand attacks still roll it (the user's Nov 13
    log finding, c7a99fa49);
  - Wildstalker 5pc takes 1 s off Multi-Shot;
  - Diabolic Mantle and Legplates of Vigilance get their Forever stats.

  The routine counted 30 report requests that move: paladin, warrior, shaman enhancement, hunter survival, druid feral cat
  and rogue subtlety.

User ruling, 2026-11-15: **keep Windfury Totem's attack power** (+95/+179/+246). Wago push 114091 seemed to remove it,
but the in-game tooltip after the hotfix still reads "1 extra attack with 246 extra melee attack power". Don't change
`WindfuryBuffBonusAP`. The TASKS item stays open for a log to confirm it.

Read first:
- the TASKS.md items dated 2026-11-13 and 2026-11-14;
- `orchestration/round_70753/README.md` (how a round lands and publishes);
- `orchestration/standards.md` 17 (the Windfury Totem rule), 37 and 38;
- the `regenerate-dps-report` skill and `tools/dps_report/README.md`;
- `orchestration/round_notes_2026-11-11/update_1112.md` and `enh_1113.md`. They are the last two landings, and this one
  follows the same steps.

A fresh clone needs:
- `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`;
- `make ui/core/proto/api.ts`;
- `DASH_TREE=<your checkout>` for `pipeline.sh`.

1. **Merge both branches** into a branch of your own off origin/master. Master moved a lot since they were cut:
   - the Enhancement round, enh_1113;
   - the attribute stat-weight step (2745da6ce);
   - Phase 0 BiS and item commits.

   Check these after the merge:
   - Windfury Weapon's main-hand attacks still roll the totem (c7a99fa49's test);
   - a totem extra swing still doesn't roll it;
   - the Unsealed Wrath gate still holds for Righteousness and Fury;
   - Judgement of Command now sets off Windfury Totem and on-hit procs.

   Update standards 17 to the new rule.
2. **Wild Growth** (TASKS 2026-11-14): make sure the three ranks share one 6 s cooldown (category 1237). Fix it if they
   don't, with a test.
3. **Scope check.** Run `pipeline.sh hydrate`, `node tools/dps_report/stale.mjs` and `pipeline.sh probe`. Re-sim every
   build that moves, in DPS, Tanking and Healing (Wild Growth moves the druid healers). Name each one's cause, and report
   anything that moves but shouldn't.
4. **Re-optimize what the rules change.** Use rotopt, the way earlier rounds did, and keep what wins on two seed sets at
   120, 180 and 300 s.
   - **Ret and its variants, and Prot paladin in Tanking:** the seal and judgement choice (Command is melee again), the
     talents and the rotation.
   - **The four Enhancement builds:** the rotation and the totem twisting. Every main-hand ability now rolls Windfury.
   - **The warrior DPS builds and Prot warrior:** the rotation. Bloodthirst, Whirlwind's main-hand strike and the other
     specials now roll it.
   - **Any other build that moves more than 1%.**
   - Re-run the hit-trade step for a build only if its talents or its hit needs changed.
5. **Land and publish** per the README:
   - rewrite `sim/testdata/report_hashes.txt`;
   - run the full suite in a subagent: `go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`.
     Regenerate goldens only where a move is one of these changes' intended effect, and say which failures were there
     before;
   - `pipeline.sh run <ids>`, then `publish <ids>`, then weights for the moved builds;
   - Changes-tab lines (`changes.py check` must list nothing uncovered);
   - update the metas, presets.ts and the wiki for each build whose talents, rotation or gear changed;
   - merge origin/master right before you publish and again before you push;
   - push to master and deploy (`make dps-report-deploy`; the proxy authenticates api.cloudflare.com);
   - update `.agents/skills/dps-report/state.json`;
   - push the final master commit to your session's own development branch too;
   - tick the TASKS items you finished, and leave open the ones a log must settle;
   - close PRs #129 and #130, each with a comment that says it landed through master. End each comment with the footer
     from your system prompt.

Rules:
- Never force-push.
- End every commit message with the trailer lines your session prompt specifies.
- Keep model names out of files.
- Keep long runs in the background or in subagents, so your own context stays small (the repo's CLAUDE.md).
- Leave the leveling weights, the leveling rotations and bayovui alone (user ruling).

When done:
- Write `orchestration/round_notes_2026-11-11/update_1115.md` with:
  - each change and how it was confirmed;
  - each moved build's old (live) and new 180 s number, with its cause;
  - the new picks and what each beat;
  - the Tanking and Healing moves;
  - the deploy id;
  - anything that needs a user ruling.
- Push it.
- Send one message to `@parent` with `send_message` (claude-code-remote MCP), with the master commit, the deploy, the
  numbers, the picks and the rulings needed. Then stop.
