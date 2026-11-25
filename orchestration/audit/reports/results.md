# Results pass 103 (2026-11-25, phase0-step4-publish 226ada86b)

Findings: 1 found and resolved in the pass (results-20261125-1). Range 63129ce35..226ada86b: every DPS request on
its Phase 0 set, 78802e6e (live on-use trinket lines), 1fed43cbf (Auto Time To Next), the trinket-line fix
67fe28cf0 and the publish of all 68 DPS builds.

- Shared run (68 requests, 20k, seed 1, CLI 01cc70b80) vs metas: 32 over 0.5 (up to 4.8), from metas written by
  the class workers' verify.py before 78802e6e and 1fed43cbf (results-20261125-1).
- After 67fe28cf0: 16 builds re-simmed and published (Affliction 960.8 -> 965.7, Destruction 877.6 -> 880.5,
  Demonology 891.9 -> 893.7, Fire 833.3 -> 835.0, Shadow Bolt 807.8 -> 810.4, Fury DW 1005.0 -> 1008.5, Sunder
  971.3 -> 973.9, Battle Shout 999.7 -> 1002.1). Ranks move by one place at most (affliction-duty-curse 12 -> 11,
  demonology 18 -> 17).
- 49 metas written from 20k seed-1 runs of their requests; builds.json within 1.3 of the run for every build.
- stale.mjs: 0 of 68 DPS builds stale (52 recorded by hand: sim/core/bulksim.go only touches the bulk sim's
  progress reporter, which the report's single sims never run). The 12 tank and 9 heal benches read stale on the
  same file and cannot be re-recorded: their packs have no probe (`benchprobe`: "no probe to compare with").
