# forever-sim

- Merge finished work into `master` and push it without asking, with no PR (user ruling, 2026-10-28). A session told to
  develop on another branch still pushes there too, then fast-forwards or merges `master` to it.
- When a change touches the DPS Report (`ui/dps_report/`, `ui/scss/dps_report/`, or the sim UI code it uses, such as the
  events view and the replay), deploy it after the merge: `make dps-report-deploy`.

## Keep the context small

Every call reads the whole conversation again, and those reads were most of our spend in
the token audit of 2026-11-03. A session at 250k costs about five times as much per call
as one at 50k. So by default, hand work that brings in a lot of text to a subagent, and
keep only its result in your own context:

- Anything that reads or produces a lot goes to a subagent (the `Agent` tool): reading
  across many files, combing a log, comparing two codebases, sim runs, test runs, builds,
  scrapes, big JSON from a tool. Ask it for the conclusion and the paths and lines that
  back it, not the raw output. Use `model: sonnet` (or the `Explore` agent) for reading and
  running, and Opus for the design, the hard debugging and the final edit.
- Do it yourself only when it's small: a file or two you're about to edit, a command whose
  output is a few lines. When you run something with long output yourself, send it to a
  file (`> out.txt 2>&1`) and read only what you need with `tail`, `grep` or `sed -n`.
- A subagent may start subagents of its own when its work really splits: independent parts
  that can run at once, or a big read it doesn't need to keep. Otherwise it does the work
  itself. Never pass a whole task down to one agent, since that adds a level that only
  waits, and never go more than two levels below the main session. Each agent's brief
  says what it must hand back.

## Waiting on workers

A session that runs workers (cloud sessions or background agents) waits for their messages:
a hand-back or task notification wakes it. Don't poll them, since every wake re-reads the
whole context: a wake every 15 minutes for two days is about 200 turns. Schedule at most
one fallback check-in, 60 minutes or more ahead, in case a worker hangs, and have it read
`get_session` before any transcript.

## Cloud sessions

Most sessions run in the cloud on claude.ai/code. You're in one when `~/projects/forever-sim`
doesn't exist. Push `master` as usual. A watcher on the user's PC brings their local checkout
(`~/projects/forever-sim`) onto master within a minute, and the dev server reloads on its own. It
skips that while the checkout has unpushed commits or uncommitted edits to the same files.

For anything else only the user's PC can do (their combat logs, the log audit, the World of Warcraft
folder), message the local play session (`play-wow`) with your `send_message` tool. Send one message
with everything the job needs and say what you want back. A local agent does it and answers you with
a message, so wait for it and don't poll. Your first message to the user's PC fails with
`elevated_session_untrusted_device`: call `confirm_repository_trust` with the digest the error
gives, and tell the user to approve the card in your session (once per session, the user agreed to
it). The play session runs only while the user plays. When nobody answers, tell the user what has to
run on their PC.

Skills were written for the user's PC, so read their steps this way in the cloud:

- `~/projects/forever-sim` or a local clone: use your own checkout instead.
- `~/.cache/wow-data` starts empty here, and the scripts download what they need into it. Allowed
  hosts so far: wago.tools, www.wowhead.com and the Wowhead bot-check host.
- A step in the World of Warcraft folder or `bayovui/` (like rebuilding BayovCore's data after a
  scrape): push to bayov/wow when your session can reach it, and its main deploys on its own.
  Otherwise ask the play session to land that part.
- Combat logs and the log audit live only on the PC: ask the play session.

On the user's PC, start heavy work (sims, scrapes, big audits) as a cloud session with `claude
--cloud "<task>"` rather than a local agent.
