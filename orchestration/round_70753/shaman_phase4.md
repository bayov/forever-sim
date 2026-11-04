# Round 70753, phase 4: shaman

## Chain Lightning 0.714 (standards 39)

`sim/shaman/chain_lightning.go`: every rank's spell power coefficient is 0.714 (2.5 / 3.5, the standard one for the
notes' 2.5 s cast; the client's rows give 0.571, rank 3 0.517). Listed in `shaman.md`'s "Notes-only values".
Enhancement casts no Chain Lightning (no request, no template knob). Elemental search (tools/rotopt shaman_ele,
`chainLightning` 1 = on cooldown / 2 = only under Lightning Mastery, mana floors 0 / 40 / 60 / 80, 20,000 iterations,
seeds 1 and 1000003; the two seed sets agree within about 1 DPS), delta against the shipped rotation at 120 / 180 / 300 s:

- Level 60: on cooldown -14 / -76 / -140; best per length -2 (40% floor) / -11 (80%) / -24 (80%); only under
  Lightning Mastery -14 to -75.
- Mana Tide: on cooldown +2.5 to +3.0 / -12 / -70; best otherwise 0 (40%) / -2 (80%) / -10 (80%); under Lightning
  Mastery -8 to -32.

Nothing wins at all three lengths, so no rotation changed. Mana keeps it out, not damage per cast.
`assumptions_def.py` has no row naming the 0.571 coefficient, so nothing changed there. The Elemental test goldens
(`TestElemental.results`; the test's default APL casts Chain Lightning) moved; `report_hashes.txt` did not.

## The six shaman metas and the report

The metas' `published.dps` (DPS report 2026-11-02) are now the 2026-11-04 finals. The re-sim at the new coefficient
reproduces the report to 0.1 everywhere:

| Build | Meta before (11-02) | 180 s before | 180 s now | 120 / 300 s now |
|---|---|---|---|---|
| shaman__elemental-level60 | 768.5 | 970.1 | 970.1 | 979.7 / 925.0 |
| shaman__elemental-level60-mana-tide | 753.2 | 952.7 | 952.7 | 963.2 / 935.6 |
| shaman__enhancement-level60 | 883.5 | 1159.4 | 1159.4 | 1170.3 / 1140.6 |
| shaman__enhancement-imp-gw | 879.3 | 1149.8 | 1149.8 | 1162.5 / 1129.6 |
| shaman__enhancement-water-shield | 834.8 | 1111.3 | 1111.3 | 1120.4 / 1109.1 |
| shaman__enhancement-mana-tide | 795.8 | 1055.1 | 1055.1 | 1065.4 / 1046.6 |

## The imp-gw raid gap (compsearch TestReportRaid)

Noise; nothing changed. On non-overlapping seed blocks (the sim seeds iteration i with RandomSeed + i, so seeds 1 to 4
share almost every iteration), the raid minus alone gap is -1.56 at the test's seed 7 / 4000 iterations (SE 1.46),
+0.34 / -0.56 / +3.05 / -2.10 on four more 4000 blocks, and -0.42 ± 0.61 pooled over 23,000 iterations (z -0.7).
Time out of mana is the same in and out of the raid, no spell moves consistently, and the trees before 7df92bc and the
paladin branch give the same. The -6.0 was a 4.1 sigma draw. No other shaman shares the imp-gw shaman's party (the
Elemental one is in party 3), and `totem_reach.go` runs only in leveling sessions. The tolerance stays at 4 sigma.

The builds that do sit off in the raid, at 1000 iterations on 4 independent seeds (80 member checks): hunter
marksmanship +2.9 (z +3.0), survival +2.8 (+3.0), beast mastery +2.0 (+2.6), warlock demonology -3.2 (-2.85).
Those, not the shaman, are the likely flakes: a hunter / warlock look.

## Tests

Full suite (`go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`): green except
`compsearch TestDropsAreUnread` (warlock__affliction and its duty variants change when Battle Shout / Trueshot Aura
are dropped), which fails the same way at ca22004, before this phase.

## Left open

- `changes.py check`: every commit of the live update is covered (the last publish run added the missing lines).
- `stale.mjs` lists 62 of 68 builds as stale ("tools/dps_report/patches changed", warlock also sim/warlock) from
  13c34f2, after the data commit; a probe would clear them.
- Master has two failures this phase didn't cause and didn't fix (a follow-up session has them):
  TestRaidForeverReportHashesStable "No item with id: 998001" (druid Balance synthetic gear missing from db.bin since
  d57aa24; it passed on this branch before the merge of 9ff134b0b) and compsearch TestDropsAreUnread (warlock__affliction).
- Hydrate notes that the four warrior Fury packs were simmed from another request than the current one (the warrior
  follow-up's requests); a warrior re-sim, not this phase's.
- The shaman builds were re-simmed twice, at f3c17a2 and on master's downranking refactor plus the rogue hotfix
  (d12e98cd9): identical to 0.1 both times.
