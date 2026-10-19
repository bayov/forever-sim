# Mechanics pass 34 (2026-10-19 18:50 UTC, master 753d81da7)

Range audited: f73aab787..753d81da7. Result: quiet (no new findings).

## Get current

| Source | State |
|---|---|
| Client build | 1.60.1.70450, no build answered up to +300 |
| Dev notes thread t/2360696 | newest post /6 (10-16 00:15, edited 10-16 20:50), unchanged |
| Blue posters | Aggrend last 10-08 (podcast); Linxy 10-17 posts are retail / Midnight; nothing new on Forever mechanics |
| Daily update | no TASKS.md / update_ledger.md changes and no forever-update merges in range |

## Commits in range (sim code, non-test)

| Commit | What | Verdict |
|---|---|---|
| ad5e98ae6 | Stoneskin flat cut (before armor) now leaves at least 1 damage after armor, only on hits it cut | Matches tier 1 beta logs (rank 1: raw 3 and 4 swings dealt 1; no 0 hit in any log), user agreed. Leveling-only in effect (raid hits are 1500). Added to verdicts.md |
| a5aba7447 | Talent Points APL value; leveling Maelstrom line gated on 4+ points | APL only, no mechanic |
| 7a6c78bed | DoT APL values resolve the DoT on the current target each time; health values can read the Pull Target | APL only; single-target raid sims see the same target |
| e20c5b3cf | Leveling idle-check skipping | Performance, results stated identical; no formula |

## Conclusions

- No shared formula, racial, raid buff, debuff or consumable value moved except the Stoneskin floor, which is tier 1 backed.
- Minor note, not filed: the 1-damage floor is applied right after armor, before any later damage-taken multipliers; with Forever's small multipliers this can only matter on 1-2 damage hits.
