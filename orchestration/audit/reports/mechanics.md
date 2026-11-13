# Mechanics pass 78 (2026-11-13, master 71562a8b0)

Quiet: no new findings. mechanics-20261112-3 (Judgement ranged) was resolved on master by f1c07e13a (the
commit moved it to Resolved itself). mechanics-20261112-1 (Holy Strike and Windfury Totem), -2 (Unsealed Wrath
echoes from Windfury Totem attacks) and 20261104-1 stay open: no Unsealed Wrath change on master
(`git grep IsWindfuryTotemAttack sim/paladin` is empty; the daily update's PR #128 still carries it, brief
`round_notes_2026-11-11/briefs/update_1112.md`).

## Commits 8306f1d52..71562a8b0 (sim code, net)

| Commit | What | Verdict |
|---|---|---|
| f1c07e13a | Paladin: judgement damage spells (Command, Righteousness, Fury) DefenseTypeRanged, ProcMaskRangedSpecial, CastTypeRanged under Forever (`rangedJudgement`, sim/paladin/judgement.go) | Matches the Nov 11 hotfix (t/2403040 post 8, tier 1): no melee procs, ranged hit. Keeps No Active Defense; hit chance still reads the shared MeleeHit stat, which is the paladin's ranged hit too. OK |
| 2a0208136 | Windfury Totem by client flag (ClientMainHand) | Dropped in the merge (c1502631c / b92a6040f) in favour of master's a0b4d3be9, audited in pass 76: `git diff 8306f1d52 origin/master -- sim/core/buffs.go attack.go spell.go` is empty. No net change |
| 67ad7c6a4 | Tank rotations; buffs.go hunk | Also superseded by the merge (no net core diff); only APLs, tests and goldens remain |
| 65b5d49cf | Healer APL health values read the healer's seen health (`seenCurrentHealth`, heal_targets.go) | APL information model, not a game mechanic. No finding |
| e05135ad3 | Resto druid APLs; healing.go comment now names Wild Growth for Nature's Swiftness | Comment matches the code (`SpellCode_DruidWildGrowth` in the NS switch, hotfix push 113885). OK |

## Getting current

- Client: newest 1.60 build 70888 (`wago.py latest --probe 300`), unchanged.
- Dev notes t/2360696: still post 10 (2026-11-04).
- Hotfix thread t/2403040: new post 9 (Kaivax, 2026-11-12 19:10 UTC) says the Nov 10 shaman changes went live at
  10:45 PST ("Weapon imbues and Weapon totems can now be used at the same time", Stormstrike 100% / 8 s, Mental
  Dexterity 33/66/100%, Maelstrom Weapon 20% per stack and 6-10 PPM). All already in the sim: cb651085c (shaman
  values, tank_mechanics / tank_rotations findings) and `sim/core/totem_weapon_buffs.go:19`, `buffs.go:2146`
  (imbue plus totem). Posts 1 and 6 unedited since pass 77.
- Blue tracker: newest posts are the above, a podcast (2412856) and nothing else mechanical.
- Daily update: 49ea90d64 (brief to land PR #128: ranged Judgement, Unsealed Wrath WF echoes, hotfix wands);
  the ranged Judgement part landed through f1c07e13a, the rest is not on master.
