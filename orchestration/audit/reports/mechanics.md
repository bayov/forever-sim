# Mechanics pass 76 (2026-11-12, master a0b4d3be98)

Quiet: no new findings. One open finding updated (mechanics-20261112-1: the warrior / rogue / Survival part
is settled by a0b4d3be9; Holy Strike alone is left, now low).

## Range
cb0d928fd6..a0b4d3be98: two sim commits, one TASKS log note (d6ffc5601), the 2026-11-12 rulings (e89962738).

| commit | what | verdict |
|---|---|---|
| e2322baa1 | Rogue: Venom's +20% poison chance reaches Instant Poison | Correct. Hotfix 113483 row 1310703 eff2 mask0 268558336 = 0x10000000 Wound + 0x10000 Deadly + 0x8000 + 0x4000 + 0x2000; client 70888 SpellClassOptions gives Instant Poison VI (11340) and 11335 mask0 8192 = 0x2000. Tier 1. |
| a0b4d3be9 | Windfury Totem under Forever rolls on white MH swings plus MH specials whose spell has Attributes_3 0x400 or Attributes_0 0x4 (generated list from 70425) | Correct for every ability with log data (tools/logcheck/findings.md: HS 21/109, Cleave 64/308, SS 178/1610, Sunder, Revenge, Execute, Raptor, Mongoose, Wing Clip roll; Stormstrike 0/4467, BT, WW, Claw, Swipe, Rake never). See below. |

## Windfury Totem check
- Regenerated the flag set from 70888 SpellMisc (`$AUDIT_WORK/mechanics/wfset.py`): 317 spells vs the
  committed 313. Added: NPC Strike 1283076, Cleave 1283413/1327248/1327716/1328013, Head Smash 1291712.
  Removed: 1310706/1310707 (deleted Mutilate rank). No player ability changes; regeneration not needed.
- Cross-checked every action in the 69 shared runs (`$AUDIT_WORK/mechanics/acts.py`): flagged and used:
  Hamstring, Ambush, Rupture, Backstab, Sinister Strike, Eviscerate, Heroic Strike, Rend, Sunder, Slam,
  Raptor Strike, Wing Clip, Mongoose Bite, Hemorrhage, Execute (cast 20662), Mortal Strike, Mutilate (MH half;
  OH half is ProcMaskMeleeOHSpecial), Spearing Strike. Not flagged: Whirlwind, Bloodthirst, Overpower,
  Stormstrike, Holy Strike, Shred, Rake, Ravage, Ferocious Bite, Lacerate, Shank, Judgements.
- No log sample yet for Rupture, Backstab, Ambush, Hemorrhage, Mutilate, Slam, Mortal Strike, Overpower, Shred,
  Ferocious Bite: they follow the client-flag rule, which fits every sampled ability. Not a finding.
- Holy Strike: still excluded; Nov 11 hotfix reads as it now activating the totem. Kept under mechanics-20261112-1.
- Note: Heroic Strike and Raptor Strike carry ProcMaskMeleeMHAuto too, so they matched the old white-only gate
  already; the change mainly adds Cleave, Sinister Strike, Eviscerate, Sunder, Execute and the other specials.
- `tools/sources/wf_totem_procs.py` defaults to ~/forever-logs/wago (absent on cloud machines); pass the wago
  cache folder explicitly. Nit, not filed.

## Get current
- Client: newest 1.60 build is still 1.60.1.70888 (`wago.py latest --probe 300`).
- Dev notes t/2360696: unchanged, post 10 (2026-11-04) newest.
- Hotfix thread t/2403040: unchanged since post 8 (2026-11-12 00:48); post 1 edit 00:47:36 as before.
- Blue tracker: newest is the Nov 11 hotfix post 8; nothing newer.

## Open findings
mechanics-20261112-3 (Judgement ranged), -2 (Unsealed Wrath echoes), -1 (Holy Strike, now low), 20261104-1
(druid scales): no commit touched them.
