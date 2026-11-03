# Round 70753, phase 3: audit suite

Run by hand on 2026-11-03: `gate.py --force results,rotations,gear,consumables,mechanics,fairness,client_update`
on master f5c99ad018, client 1.60.1.70753. The shared sims covered all 68 report requests at 20k iterations, seed 1,
with 0 errors. Healer and tank auditors stayed off. Each pass's report is in `orchestration/audit/reports/<id>.md` and
its findings are in `orchestration/audit/findings/<id>.md`.

## Per auditor

| Auditor | Pass | Result |
|---|---|---|
| results | 73 | 2 new (low). All 68 shared-sim DPS values match the class round files' finals to rounding. The six shaman metas are stale by design; their new numbers went under results-20261102-1. |
| rotations | 49 | Quiet. 28 APLs were reworked. Checked clean: dead lines, talent legality (the rogue, warrior and shaman trees match `tools/forever_talents/notes`), duty uptimes, overcap and OOM, and paired probes. |
| gear | 53 | Quiet. The Restoration shaman set is within budget. No preset enchant changed. |
| consumables | 67 | Quiet. Trueshot Aura is at the notes' values. Dragonfire and Caustic Smog can't crit. No preset consume changed. |
| mechanics | 61 | 1 new (low). mechanics-20261031-1 (Efficiency and Lacerate) is resolved. |
| fairness | 25 | Quiet. Every build's move matches its round file. Top over bottom is 1.73x (it was 1.49x). The report races still hold. |
| client_update | 8 | 1 new (low). client_update-20261103-1 (Poised) is resolved. The 8 hotfix lines are all in the sim. Ledger: 765 rows, 380 landed, 31 missing. |

## New findings, by class

**Hunter**
- results-20261103-2 (low, needs: worker fix). The notes-only Arcane Shot 0.15 and Serpent Sting 5% AP shares (`arcane_shot.go:19`,
  `serpent_sting.go:22`) and the hotfix shot tables are not gated under Forever. Only the Classic sim is affected.
  `hunter.md:126` calls the hunter files Forever-only by design, so this may be accepted.
- client_update-20261103-2 also covers this, below.

**Rogue, shaman, hunter, paladin**
- client_update-20261103-2 (low, needs: user ruling). Several notes-only values are not gated under Forever, though
  standards 37 and `rogue.md` say they are:
  - rogue: the talents.go values, Mutilate 0.9, Flawless Execution 15, Shank and Ghostly Strike 7 s;
  - hunter: the Arcane Shot and Serpent Sting shares;
  - paladin: Vengeance at 5 stacks;
  - shaman: its talents and spell tables.

  This moves no DPS under Forever. It does change Classic results: rogue Relentless Strikes and Seal Fate were Classic's
  values before this round. The mechanics and fairness passes spot-checked the gating and found it in place for the
  values they looked at, so the client_update list is the fuller one.

**Rogue**
- results-20261103-1 (low, needs: orchestrator fix). `rogue__subtlety.meta.json` has race "Night Elf", where every other
  meta has "NightElf". `regen.py` fails on it with "unknown race", and `build_data.py`'s isPreset compare is probably wrong
  too. The request itself regenerates with 0 diffs. The fairness pass made `tools/audit/fairness/table.py` tolerate the
  space so it could run.

**Druid (Feral tank), warrior (Protection)**
- mechanics-20261103-2 (low, needs: worker fix). The notes' tank threat values (Bear Form 115%, Defensive Stance 80%,
  Defiance 7/13/20%) are not in the sim: `bear.go:52` is still 1.5 and `stances.go:110-112` still uses 1.3 and 0.05 per
  point. This is threat only, 0 DPS. No to-do list picks it up yet.

**Mage, priest, warlock:** no new findings.

## Checked and holding
- Standards 37: every notes-only value in the nine round files matches the post 9 text (client_update checked all of
  them; mechanics, results and fairness spot-checked). Gating is the one gap, in the findings above.
- The rogue, warrior and shaman talent trees match the notes' rows, ranks and prerequisites. The generated UI trees match
  the notes JSON.

## Left for the next run
- Client build 1.60.1.70755 appeared on wago.tools after this run, so the gate now shows mechanics and client_update due
  for it.
- These landed on master after f5c99ad018 and were not audited here: the warlock Affliction rework (dc0a0fcfa, b17875a3a)
  and the shaman exploration (5f88ddea1). Fairness is due for them.
- The rotations pass didn't re-run a full per-class threshold scan, because the class workers ran knob searches this
  round.
