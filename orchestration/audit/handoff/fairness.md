# Fairness auditor: handoff

## The role

Make sure the Level 60 DPS presets compare fairly across classes. That means:
- equal-strength synthetic gear (standards 1b, 5a, 5b);
- the same representative consumables, buffs, debuffs and raid assumptions for every spec;
- fair set bonuses, weapons and races;
- a published cross-spec ranking that matches what the current requests actually sim.

Rules:
- Don't fix class code yourself.
- Class-specific defects go to the orchestrator (or the class worker, when the orchestrator allows it).
- Cross-class questions go to the orchestrator as a proposal with measured options, then wait for its ruling or the user's.
- No PR comments. Don't use `tools/rotopt`.
- Never change master or class branches.

## What to watch

On each pass:

```sh
git fetch origin master
git log --oneline <last audited sha>..origin/master
git diff --stat <last audited sha> origin/master -- sim proto tools/dps_report/req assets/database \
    orchestration/standards.md 'ui/*/presets.ts' 'ui/*/apls' 'ui/*/gear_sets' tools/rotopt/mksettings
```

- **The pass is quiet** when that diff is empty. Changes that only touch the leveling mode, the report UI (`ui/dps_report`, `ui/core`), `datasets.json` / `provenance.json` or tools don't count.
- **Do a full pass** when it touches:
  - `sim/core` (every class);
  - a class's `sim/<class>`, presets or APLs;
  - `tools/dps_report/req`;
  - the item database;
  - `orchestration/standards.md`.

External sources the rulings rest on:
- client builds on wago.tools (`https://wago.tools/builds`; the current client is 1.60.1.70425, and 70450 is cosmetic only per the ledger);
- Blizzard's Forever development notes (`us.forums.blizzard.com/en/wow/t/2360696`).

A new client build usually brings shared changes such as racials and raid buffs. Once the class ports land, re-sweep everything.

## Commands

```sh
export AUDIT_TREE=/path/to/checkout-at-origin-master AUDIT_WORK=/path/to/scratch
tools/audit/fairness/build.sh                         # $AUDIT_WORK/wowsimcli
python3 tools/audit/fairness/audit1b.py               # 19 lines, each "5a PASS | 1b PASS"
python3 tools/audit/fairness/sweep.py > $AUDIT_WORK/sweep.log      # about 20 min, 4 threads
python3 tools/audit/fairness/table.py $AUDIT_WORK/sweep.log \
    orchestration/audit/ref/fairness/sweep_baseline.log <baseline sha> > $AUDIT_WORK/table.md
```

`table.py` prints `<--` lines on stderr for every row that needs a look.

## What to compare, and the tolerances

- **Published against current:** `meta.published.dps` against the report race's DPS in the sweep. At 5000 iterations the noise is about ±1.5. A larger gap means the report is stale for that build. The usual cause is a shared change (raid buff, ruling) that was re-simmed for only some classes, which makes the ranking unfair. That's a finding.
- **Best race against the report race:** the report should use the best race. A loss of 1 or more is worth a note to the class owner. Ties within 0.5 are noise; Arcane and Survival are exact Human/Orc ties today.
- **Movement against the baseline:** for every build that moved, attribute the change to a commit. An unexplained move, or a move in a class whose code didn't change, is a finding: it means a shared change leaked.
- **Gear:** every `audit1b.py` line must pass. A FAIL goes to the class owner with the slot and the ratio.
- **Shared assumptions applied unevenly** are the classic finding. Examples:
  - raid hits that push back one class's casts but not another's;
  - a pushback-avoid talent taken but not modelled;
  - one class's numbers re-simmed after a ruling and the others not.
  
  Measure the size (a local hook or a request edit, never a commit) and propose options to the orchestrator.

## Where findings go

- Cross-class findings and the ranking summary: the orchestrator session, `session_01AGmMcVxGoubbDodiQDQwhz` (as of 2026-10-17).
- The results auditor (`session_01X9dJY1kQbnkLdLwRbqh6GT`) re-sims the report at 20k iterations and often reports the same gaps. Don't duplicate a point it has already raised; confirm it independently instead.
- Keep `orchestration/audit/ref/fairness/fairness_report.md` §1 current: the new table, the old one moved into a `<details>` block, and a short "what moved / findings" list.

## Baseline (2026-10-17 00:50 UTC)

- Master 1735382, client 1.60.1.70425.
- Report version 35; its re-sim on 787da848 changed only report data.
- Every base build and variant is within 1.2 DPS of its published number. The report race is the best race everywhere.
- Top over bottom is 1.79×.

| Rank | Spec | DPS |
|---|---|---|
| 1 | Demonology | 1177.8 |
| 2 | Lone Wolf | 1094.6 |
| 3 | Frostfire | 1083.7 |
| 4 | Fire | 1067.5 |
| 5 | Affliction | 1066.4 |
| 6 | Destruction | 1048.1 |
| 7 | MM | 1044.8 |
| 8 | Destruction Fire | 1019.7 |
| 9 | Fury | 989.7 |
| 10 | Arms | 976.8 |
| 11 | BM | 971.0 |
| 12 | Fury 2H | 971.0 |
| 13 | Arcane | 964.8 |
| 14 | Shadow | 960.6 |
| 15 | Survival | 934.9 |
| 16 | Enhancement | 892.0 |
| 17 | Ret | 889.1 |
| 18 | Balance | 873.1 |
| 19 | Frost | 844.4 |
| 20 | Elemental | 772.0 |
| 21 | Smite | 765.8 |
| 22 | Feral | 752.9 |
| 23 | Combat Swords | 726.5 |
| 24 | Combat Daggers | 725.5 |
| 25 | Seal Fate Daggers | 687.2 |
| 26 | Mutilate | 683.9 |
| 27 | Seal Fate Swords | 677.8 |
| 28 | Subtlety | 656.7 |

The full table, with races, is `orchestration/audit/ref/fairness/table_baseline.md`.

## Open items

- **Hunter end-of-fight test:** Aimed and Sniper Shot past the fight's end. The rotation auditor was asked to test it; the result hasn't been seen yet.
- **Ice Lance Low range (user call pending):** the sim uses 0.429 (Season of Discovery server value). The client says 0, and the dashboard Low is 0.2145. The recommendation is that Low include 0, the only value with a Forever source; at 0, Frost loses about 18.5%.
- **Flametongue Totem stacking with the Windfury imbue (`sham_ftt_stack`):** user ruling "assume it works". Still To Verify in game; worth +38 to Enhancement.
- **Partial resist against bosses:** "no level resist" is ruled (#102). A boss log is still on notes.md's To Verify list.
- **Warrior Rage k (standards 21) and demon / pet scaling:** see the report's §3. These are the largest assumption-driven ranges.

## Settled: don't re-flag

- Rogues never start in stealth (user ruling).
- Every DPS takes 2 raid hits a minute as real spells. Pushback is correct: Slam and the casts carry the pushback flag (InterruptFlags 15) in SpellInterrupts. Nature's Focus, Improved Channeling and Burning Soul are modelled.
- No racial depends on the weapon any more (client 70425). The weapons-racial-off column is retired. Human Ruthless gives 2% crit to spells too, per the client text.
- Salvation is threat only. Ret runs its own Concentration Aura, like shamans drop their own totems.
- Melee raid: Grace of Air plus Windfury Totem (standards 25). The second shaman's fire slot is Flametongue Totem, which only Enhancement benefits from. The raid stays symmetric: each melee spec gets its best weapon totem.
- Warlock demons and hunter pets take no player buffs (standards 14), so the Moonkin / LotP split changes no warlock number.
- Creature-type talents on the generic boss: Improved Tracking on; Murder, Ret's type bonuses and Beast Slaying off (standards 19).
- Trinkets: two budget-equal synthetic trinkets per spec (1b).

## Gotchas

- **Sim the report's own requests, not mksettings with a hand-kept talent copy.** The copies went stale silently after the class ports. The old existence check passed because the stale strings survive as vs-type variants.
- **Cache on the binary too.** Item stats live in the compiled database, so a database-only change needs a rebuild and a new cache key. `sweep.py` keys on the binary's mtime.
- **SoD items carry both melee and spell hit/crit.** Take the max, not the sum. `audit1b.py` does.
- **5b:** Ret and Enhancement hold their SP and Int, and the 5a check must not rescale them.
- **Never use `pgrep`/`pkill -f <script name>`.** It matches your own shell and kills it. Wait on the PID or run the sweep in the background with a completion notice.
- **Shaman and warrior raid-hit options look alike but differ.** Read the code before assuming an effect: the shamans' hits once only fired the shield and dealt no damage.
- **`tools/audit/fairness/hacks_optional.py`** holds three env-gated local switches used for one-off measurements: no weapon racials, skipping set-bonus tiers, and a Rage-only raid hit. Apply it to a scratch checkout only, never commit it. The sweep doesn't need it.
