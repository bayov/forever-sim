# Round 70753, phase 3: client 70755 and the audit's findings

The round 70753 audit suite (`orchestration/round_70753/audit.md`) found four low findings, and client build
1.60.1.70755 appeared on wago.tools after it ran. The release launches 2026-11-04. Read
`orchestration/round_70753/README.md` (the round, standards 37 in `orchestration/standards.md`: notes-only values),
`orchestration/round_70753/audit.md`, and the findings it names in `orchestration/audit/findings/<id>.md`.

The DPS report is being regenerated at the same time by session `session_01TXnjyRkxieT1vttRnuMQBY` (phase 3 report).
Whenever something you land on master moves a Level 60 report build's DPS, send that session one message with
`send_message` (claude-code-remote MCP) naming the master commit and the build ids, so it re-sims them before it
publishes.

Do, in this order, one topic per commit, each landed on master (the repo's CLAUDE.md) with its tests:

1. **Client 70755.** The `update-sources` skill for the new build: pull its wago.tools tables, diff them against
   70753 (`tools/sources/stale.py`, the skill's diff step), and port what changed. For each notes-only value
   (standards 37; every class round file lists them under "Notes-only values") that 70755 now carries, take the
   client's number, change its source to the client and take it off the notes-only list (gating stays, since
   Forever-only). A client value that differs from the notes' number wins (source tiers). Talent trees: if 70755 has
   the rogue, warrior or shaman rework, compare it with `tools/forever_talents/notes/<class>.json` and switch the
   class to the client tree when it's there (check every preset's talent string stays legal; a re-pick is a worker's
   call, so if a preset breaks, say so in your report instead of re-optimizing). Write what 70755 changed and what you
   did with each line to `orchestration/round_70753/client_70755.md`. Rewrite `sim/testdata/report_hashes.txt` if
   builds moved, and send the report session the moved build ids.
2. **Gating** (client_update-20261103-2 and results-20261103-2): every notes-only or hotfix-only value still applied
   outside Forever gets gated under Forever, as standards 37 says, so the Classic sim gets Classic's values back. The
   list is in client_update-20261103-2 (rogue talents.go values, Mutilate 0.9, Flawless Execution 15, Shank and
   Ghostly Strike 7 s; hunter Arcane Shot and Serpent Sting AP shares and the hotfix shot tables; paladin Vengeance at
   5 stacks; shaman talents and spell tables), plus anything else you find of the kind. No Forever DPS may move: the
   Forever rows of report_hashes.txt stay unchanged, and say which Classic tests and goldens moved.
3. **Tank threat** (mechanics-20261103-2): the notes' Bear Form 115%, Defensive Stance 80% and Defiance 7/13/20%
   (`bear.go:52`, `stances.go:110-112`), gated under Forever, cited, and listed in the druid and warrior round files'
   "Notes-only values" (unless 70755 carries them). 0 DPS; check the tank presets' threat tests and say what moved.
4. **Enhancement leveling XP/h** (README "Phase 3 to-do"): measure `ui/enhancement_shaman/leveling_weights.json` again
   (the `leveling-opt` skill's measuring step; leave the template's knobs alone unless the rotation no longer works on
   the new tree, and then say so). bayovui's `weights.py` is rerun from the wow repo later; list it as a to-do.

Mark each finding resolved the way `orchestration/audit/README.md` says. Rules: never force-push; end every commit message
with the trailer lines your session prompt specifies; no model names in files; long runs (tests, sims) in the background
with short polls or in subagents, keeping your own context small (the repo's CLAUDE.md). When done, write a short summary
at `orchestration/round_70753/fix.md` (per item: what changed, which builds moved, tests), push it, and send one message to
`@parent` with `send_message` giving the master commits and anything that needs a user ruling. Then stop.
