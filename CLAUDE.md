# forever-sim

- Merge finished work into `master` and push it without asking, with no PR (user ruling, 2026-10-28). A session told to
  develop on another branch still pushes there too, then fast-forwards or merges `master` to it.
- When a change touches the DPS Report (`ui/dps_report/`, `ui/scss/dps_report/`, or the sim UI code it uses, such as the
  events view and the replay), deploy it after the merge: `make dps-report-deploy`.
