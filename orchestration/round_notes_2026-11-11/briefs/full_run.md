# Notes round 2026-11-11: full DPS, Tanking and Healing re-sim and deploy

Since the last full deploy (5aa05cc0, master 8dc49856), master has landed:
- the shaman Nov 10 tuning, already published for Enhancement only (`orchestration/round_notes_2026-11-11/shaman.md`);
- the other classes' changes, not published (`classes.md` here): Wrack and Soul Siphon, Bane of Agony in Affliction,
  hunter mana costs and hawk, Arcane Concentration, Nature's Swiftness on Wild Growth, Sweeping Strikes, item data;
- the dot-after-expiry fix f20d593 in sim/core;
- the log audits' sim changes since Nov 4, among them Windfury Totem procing from white main-hand swings only
  (0ce9dbf40), Maelstrom Weapon, Flametongue rank 5, Orc base stats, enchants 41 and 66, Farraki Enforcement, and
  items seen in game.

You re-sim every build and both modes, publish, land and deploy. The steps are the same as last time's brief,
`orchestration/round_70753/briefs/phase5_full_run.md`. Read it and do its steps 1, 2, 5 and 6 with these changes:
- **Compare against the live report**, every build's 180 s number. A move beyond the noise needs a named cause: a
  commit, and the classes or shaman round's own number where there is one. Survival hunter is known to drop from about
  1041 to about 945 (the classes worker's sims), from an earlier change that nobody has named yet. Find which commit
  did it and whether it's intended. If it isn't, report it rather than ship a wrong number.
- **Changes tab.** Write lines for whatever moved that no round has covered yet: the log-audit changes, and the dot fix
  if the classes worker's line doesn't cover it. `changes.py check` must list nothing uncovered.
- **Tests.** These packages fail on master from expected results that older commits didn't update: mage,
  priest/healing, priest/shadow, rogue/dps_rogue, warrior/dps_warrior, and tools/rotopt TestWarriorDpsHoldUnderBuff.
  For each, find which commit moved it. If the move is the intended effect of that commit, regenerate the expected
  results. If it isn't, report it. The full suite must be green before you land.
- **state.json** `foreverClientBuild` still says 70755. Set it to the client build the data now comes from (70888 if
  the classes worker's pull used it; check).
- **Timing.** Send the timing summary to `session_01TXnjyPE5JfDBjP3yqKyxXE` only if that session is still active
  (`get_session`). Otherwise put it in your round file.
- Several other sessions push to master (the log audit loop, the daily update). Merge origin/master right before you
  publish, and again before you push. If a merge moves a build, re-sim just that build.
- The user's open rulings from this round (Arcane Concentration's reading, Bane of Agony's ramp, the warlock default,
  the Windfury cooldowns) may change picks after you land. Leave every pick as master has it.

When deployed:
- Write `orchestration/round_notes_2026-11-11/full_run.md` with:
  - the 180 s ranking;
  - every build's old (live) and new 180 s number, each move with its cause;
  - the race-table changes;
  - the Tanking and Healing moves;
  - the tests you regenerated, and why;
  - the deploy id.
- Push it.
- Send one message to `@parent` with `send_message` (claude-code-remote MCP), with the master commit, the deploy, the
  numbers and anything that needs a ruling. Then stop.

Rules:
- Never force-push.
- End every commit message with the trailer lines your session prompt specifies.
- Keep model names out of files.
- Keep long runs in the background or in subagents, so your own context stays small (the repo's CLAUDE.md).
