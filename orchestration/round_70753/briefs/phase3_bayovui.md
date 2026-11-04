# Round 70753, phase 3b: bayovui's data from the new forever-sim

You work in the wow repo (github.com/bayov/wow, cloned in this session) and need forever-sim too: attach
`bayov/forever-sim` with the add_repo tool (claude-code-remote MCP) and clone it next to wow as `forever-sim`. The nine
class workers of forever-sim's round 70753 (Blizzard's November 2 development notes, client build 1.60.1.70753)
re-optimized every class, and forever-sim's master has their work (`orchestration/round_70753/`). Bring BayovCore's
data built from forever-sim up to date.

Read the wow repo's `CLAUDE.md`, `bayovui/AGENTS.md` and README, and the `update-data` skill
(`.agents/skills/update-data/SKILL.md`), section 4 "bayovui from forever-sim". This container is a cloud clone, not the
user's machine: there is no jj (use git), no game client files beyond what's tracked, and `deploy.py` can't reach the
game. Steps that need the user's machine (classspells.py's QuestieDB, a Wowhead scrape from there, deploy, in-game
checks) go in your report as to-dos, not workarounds.

Do, in this order, each with its tests:
1. Built-in leveling stat weights (`UpgradesData.lua`): the `update-leveling-weights` skill, for every class it covers.
   The shaman and rogue trees changed the most (forever-sim's `tools/forever_talents/notes/<class>.json`). Run its sims
   with a background job and short polls.
2. BiS lists (`UpgradesBiSData.lua`): `uv run bayovui/build/bis.py`.
3. Revealed items (`UpgradesRevealsData.lua`): `uv run bayovui/build/reveals.py` (forever-sim's hidden.py ran at 70753).
4. Leveling rotation (`LevelingRotationData.lua`): `levelrotation.py` needs `classspells.py`, which needs QuestieDB on
   the user's machine. If it can run here on the tracked data, do it; if not, list what's stale (the shaman's Lava Burst
   out, Lightning Mastery, Stormstrike 6 s, Improved Ghost Wolf row 4, Totemic Call; Shank and Poised for rogues).
5. The user's own characters' stat weights (the `update-stat-weights` skill) for every character whose class changed a
   lot in this round, if the tracked character history allows it here. The user plays a shaman (Bayov Oom), and
   Enhancement moved the most of any spec.

Commit with git in topical commits with real messages, scoped to your paths, each ending with the trailer lines your
session prompt specifies; no model names in files; never force-push; don't touch WTF/ or third-party addons. Push to
`main` (fetch and rebase first; the repo's practice is that main is ready to ship). Write a short summary in your last
commit's message and send one message to `@parent` with `send_message` (claude-code-remote MCP): the commits, what
changed, and the user's to-dos (deploy `uv run bayovui/build/deploy.py beta forever`, in-game checks, anything that
needs their machine). Then stop.
