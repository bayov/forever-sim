# forever-sim

- Merge finished work into `master` and push it without asking, with no PR (user ruling, 2026-10-28). A session told to
  develop on another branch still pushes there too, then fast-forwards or merges `master` to it.
- When a change touches the DPS Report (`ui/dps_report/`, `ui/scss/dps_report/`, or the sim UI code it uses, such as the
  events view and the replay), deploy it after the merge: `make dps-report-deploy`.

## Keep the context small

Every call reads the whole conversation again, and those reads were most of our spend in
the token audit of 2026-11-03. A session at 250k costs about five times as much per call
as one at 50k. So:

- Send long output to a file, not into the conversation: sim runs, test runs, logs, builds
  and scrapes. Redirect it (`> out.txt 2>&1`), then read only what you need with `tail`,
  `grep` or `sed -n`. The same goes for big JSON from a tool.
- Hand bulk reading to a Sonnet subagent (the `Agent` tool with `model: sonnet`, or the
  `Explore` agent): reading across many files, combing a long log, comparing two codebases.
  Ask it for the conclusion and the paths and lines that back it, not the file contents.
  Keep Opus for design, hard debugging and the final edit.
- Around 150k tokens, or when the task turns into a different one, stop and write a
  handoff: what's done and pushed, what's left, the decisions made and why, and the files
  and commands that matter. Put it in your reply so the user can paste it into a new
  session, and suggest they do. Don't let one session run on for many hours through
  several tasks.
