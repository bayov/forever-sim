# Mechanics pass 95 (2026-11-20, master 813f2b1e9)

Range 4be96efef..813f2b1e9: one sim commit (4c3ac0026, Hand of Justice against Dwarves) and one TASKS.md
item (Enhancement Phase 0 BiS in-game checks, nothing mechanical). Client unchanged at 1.60.1.70959. No new
Blizzard posts. Quiet: no new findings. Open findings mechanics-20261120-1, 20261118-2, 20261118-1, 20261113-2
and 20261104-1 stay open (nothing in the range touches them).

## Get current

| Source | Seen | Newer? |
|---|---|---|
| Client (wago.py latest --probe 300) | 1.60.1.70959 | no |
| Hotfix thread t/2403040 | post 13, 2026-11-20 04:00:55Z (post 1 edited 04:01:18Z) | no |
| Dev notes t/2360696 | post 10, 2026-11-04 19:35Z (post 1 edited 2026-11-03 03:20Z) | no |
| Daily update (TASKS.md / update_ledger.md, forever-update merges) | b47c1d55e only (BiS in-game checks) | nothing mechanical |

## Commit triage

| Commit | Change | Check | Verdict |
|---|---|---|---|
| 4c3ac0026 | Hand of Justice procs 3% against targets with the new Dwarf mob family, 1% otherwise under Forever (2% Era); new `ConditionVsDwarves` item condition for the BiS list | sp.py 15600 on 70959: ProcChance 3, ICD 2000 ms, #0 aura 42 -> 15601 (extra attack), #1 dummy aura bp 3 | Matches tier 1 plus the 2026-10-10 user ruling (a third of the procs off Dwarves). Default targets are MobFamilyUnknown and no preset sets Dwarf, so no report DPS moves. Dark Iron Dwarves counting is unconfirmed, as the code comment says. |

Knock-on: `MobFamily_MobFamilyDwarf` is used only by the HoJ proc chance and the item condition (git grep);
`result.Target` is the hit's target, so a multi-target fight rolls each target's own chance.

## Updated

- `ref/mechanics/verdicts.md`: Hand of Justice row, "as of" moved to master 813f2b1e9.
