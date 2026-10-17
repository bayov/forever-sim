# Pass: rotations and resources

Follow `orchestration/audit/PASS.md` around this doc. It started as the long-running rotation auditor's handoff (2026-10-17); keep it current when the method changes.

## Role

- **What:** audit every class's Level 60 rotations for DPS gains and resource handling (mana, rage,
  energy).
- **Don't:**
  - Don't change rotations or class branches, and don't commit. You audit and file findings; the orchestrator applies them (B13).
  - Don't use tools/rotopt searches, or the cli flags `-eval`, `-timeline` or `-breakdown`.
- **Do:** use paired A/B runs with hand-edited APL copies (`tools/audit/rotations/ab.py`), and
  debug logs (`prepull_log.py`, or `simOptions.debug`).
- **Cadence:** the audit loop's gate starts a pass when a watched file changes on master (checked every 6 h).

## What to watch

1. **Master and class-branch commits** (`git fetch origin`):
   - `ui/**/apls/**`: the Level 60 files. Note that several aren't named `level60*`:
     - hunter `bm/mm/lw/sv(_mark/_scorpid)`
     - enhancement `optimized/magma/mana_tide`
     - the `*_duty*` and `*_battle_shout` duty variants
   - `sim/**/*.go`, excluding `_test`: rage, cast and pushback, racials, buffs and debuffs, spell
     resistances, and class talents.
   - `tools/dps_report/req/*.json`: the canonical preset settings, rotation embedded at
     `raid.parties[0].players[0].rotation`.
   - `orchestration/standards.md` and `orchestration/update_ledger.md`: rulings and client builds.
     The ledger says whether a client update is cosmetic or not.
2. **External:** client updates and Blizzard notes reach the repo through the update-ledger / Forever-update
   PRs. Watch the ledger rather than wago.tools directly.

### Deciding whether a pass is quiet

Run `apl_logic_diff.py <last> origin/master` and check for sim `.go` changes since the last audited
master. The pass is quiet when both of these hold:
- No Level 60 APL logic changed (notes-only edits don't count).
- No non-test sim `.go` changed, or only leveling code that the raid presets don't reach.

The gate already filters on the paths; if `apl_logic_diff.py` shows only notes-only edits and no sim change the presets reach, the pass is quiet: update `state/rotations.json` and stop.

## Running a pass

```
source tools/audit/setup.sh                       # AUDIT_WORK and the toolchain
export AUDIT_TREE=$(dirname $(tools/audit/build_cli.sh origin/master))/tree   # a detached checkout of origin/master; never edit it
# The loop's shared run ($AUDIT_WORK/runs/master/summary.tsv, 20k, seed 1) already sims every request on the new master.
# Old CLI for regress.py: tools/audit/build_cli.sh <last.master from state/rotations.json>
tools/audit/rotations/apl_logic_diff.py <last_sha> origin/master
tools/audit/rotations/build_cli.sh cli_new
tools/audit/rotations/baseline.py $AUDIT_WORK/cli_new > $AUDIT_WORK/base.tsv    # diff against ref/rotations/baseline_*.tsv
tools/audit/rotations/regress.py <old_cli> $AUDIT_WORK/cli_new                   # if the sim changed
tools/audit/rotations/ab.py $AUDIT_WORK/cli_new $AUDIT_TREE/tools/dps_report/req/<preset>.json - variant.json --s0 1000000
tools/audit/rotations/ab.py ... --s0 200000000                                   # second seed set, mandatory
```

### Verifying merged findings

- First check that each open finding in `findings/rotations.md` that has since landed on master landed as written. Use the logic diff on the APL file, and move it to `## Resolved`.
- Workers sometimes re-measure and apply a variant instead. That's fine when their re-measurement backs it.
  Spot-check one with `ab.py` if in doubt.

### Large changes

When a pass touches many classes (a client data update, a sim mechanic change, new rulings), fan out to
about 4 subagents (Sonnet is enough for the A/B runs; keep the judgment in your own context):
- warrior + paladin
- rogue + hunter + feral
- mage + priest
- shaman + warlock + balance

Give each the brief in `tools/audit/rotations/brief.txt`, with its "What changed" section updated.

## Method and tolerances

- **Paired runs:** 20,000 iterations as 8 blocks of 2,500. Block k uses seed S0 + k·3e6, with the same
  seeds for A and B. Report B−A ± SE at 120, 180 and 300 s.
- **Two seed sets:** confirm every candidate on two independent sets (S0 = 1e6 and S0 = 2e8) and report
  the pooled mean. The 8-block SE understates the spread between seed sets by about 40%: one Enhancement
  finding read +2.5 on one set, and its true value over 10 sets was +1.4.
- **Bar for a finding:** a pooled gain of at least about 1 DPS at 180 s that doesn't lose at 120 or
  300 s. Correctness bugs count regardless of size: a dead line in the wrong place, a wrong spell id or
  rank, or a duty that lapses.
- **Report borderline results as optional**, not as findings: under 1 DPS, or positive at 180 s only.
- **Fight-length trade-offs** (+X at 180 s but negative at 120 or 300 s) are not findings.

## Where findings go

- Every finding goes to `orchestration/audit/findings/rotations.md` (format in `PASS.md`), `needs: orchestrator fix`. Add to its evidence:
  - the exact JSON line or lines, and the file:line in master's APL;
  - paired numbers for both seed sets and pooled, at 120/180/300 s;
  - the baseline at 180 s.
- Optional results under the bar go in `reports/rotations.md` under "Optional", not in the findings file.
- When a regression baseline moves (sim changed), write the new `baseline.py` output as `orchestration/audit/ref/rotations/baseline_<sha>_3000it_seed424242.tsv`, delete the old one, and record its name in `state/rotations.json` as `baseline`.

## Current baseline (pass 26, 2026-10-17)

- **Last audited master:** d2633f4a4.
- **Client:** 1.60.1.70425 data. 70450 is cosmetic per the ledger.
- **Sim state:**
  - No level-based partial resists (#102).
  - Off-hand Rage ×0.5.
  - Raid hits in core: 2 a minute at 1500 damage. Warriors get rage from them, and Enrage procs on them.
  - Racials rework: Blood Fury is on a 3-minute cooldown; Windshaper, Human, Undead and High Order are
    passive-only in raid.
  - Curse of the Elements and Mana Spring last 300 s.
- **Reference DPS:** `orchestration/audit/ref/rotations/baseline_d2633f4_3000it_seed424242.tsv` (the newest is named in `state/rotations.json` as `baseline`).

## Open items and pending rulings

- **Balance:** Night Elf with Elune's Light vs Tauren was never run. Settings were drafted in the old
  container; rebuild from the req by setting race Night Elf and re-adding the Elune's Light lines from git
  history before e64360398.
- **Elemental Mastery:** holding it was never tested; it's on autocast, fired with the trinket.
- **Warlock duty-curse presets:** Life Tap, Doom and DoT variants weren't screened separately; the base
  results should carry over.
- **Optional (under the bar), offered to the class workers before they were retired:**
  - Rogue Mutilate window 0.5 s (about +0.7).
  - Warrior Fury-BS Heroic Strike 55 (inconsistent between seed sets).
  - Hunter end-of-fight Aimed or Sniper gates (+0.5–0.8).
  - Shadow Mind Blast time-left gate (+0.5).
  - Destruction Fire DoT gates (+0.7).
  - Water Shield Magma gate (+0.9).

## Settled: don't re-flag

- **Race-specific lines in shared APLs are intentional.** Examples: Dark Sacrifice for Undead,
  Chastise for Dwarf, Berserking for Troll, Elune's Light, Blood Fury on non-Orc presets.
- **Talent-less lines are inert by design.** Adrenaline Rush in Mutilate and Subtlety, Missile Barrage in
  Frostfire, Cold Blood in Combat Daggers.
- **Multi-Shot never fires in BM, MM or LW,** because Aimed Shot takes the shared cooldown first.
  Multi-Shot first loses.
- **Curse of the Elements duty:** keep it on the first GCD, even though casting it after your own DoTs
  gains about 3 for the warlock. The raid loses more.
- **Shot travel time:** shots apply damage after travel, so hunter end-of-fight rules gain little.
- **Lightning Shield and Water Shield:** Enhancement keeps Lightning Shield (Water Shield loses 8–11).
  Elemental's recast is gated on ≥50% mana and ≥60 s left.
- **Enhancement Mana Tide:** the worker's "Fire Nova after shocks" beats my "Flame Shock above Stormstrike"
  under Flametongue (verified in pass 25).
- **Berserker Rage:** loses in every placement.
- **Smite:** the rank 7 fallback and 35 mana a second hold.
- **Paladin twist:** 1.5 s before the swing is a sharp optimum.
- **Stealth and Prowl openers:** none (user ruling).
- **Pet consumables:** none.
- **Demon buffs:** the demon gets no melee party buffs.

## Gotchas

- `auraRemainingTime` on an expired aura returns −currentTime, not 0. Guard with `auraIsActive`.
- APL `Duration × fractional constant` truncates to 0. Use integer arithmetic (×4 then ÷5).
- Seeds iterate as seed+i inside a run, so spread blocks out (3e6 apart).
- Warrior Land Mine: its 1 s cast restarts swing timers, so throw it right after a main-hand swing.
- Shaman totems use a 1 s GCD, so pre-pull totems 1 s apart are fine.
- Mana gems and runes share a cooldown. Potions have their own 2-minute cooldown. Sapper and Land Mine
  share a 1-minute cooldown.
- Pre-casts must land at ≥ 0: damage counts from t = 0.
- When an old CLI runs new req files after a client data update, unknown spell ids give nonsense numbers.
  Attribute moves via `git log`, not that table.
- The shared `$AUDIT_TREE` checkout may be read by subagents. Don't move it while they run. Give
  each subagent its own worktree if master may move.
- The machine can get heavily loaded with parallel agents. Keep to about 4, and bound each one's work: a pass
  shouldn't run for hours (`PASS.md`).
- Workers merge fast. Before reporting a branch as "not merged", check `git log origin/master` for the
  merge.
