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
