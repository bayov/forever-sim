# Paladin: round 70753 (the November 2 development notes)

Worker: Paladin (Retribution DPS specs). Brief: `briefs/paladin.md`. Status: **landed** (Phase A and B) on master at 6d678dc.

## Final summary

**What changed** (each gated under Forever, with a test):

- **Vengeance stacks up to 5 times** (was 3). Notes-only value (standards 37); client 70753 still has 3 (20050).
  About +45 at 180 sec on every build (+4.4%).
- **Two-Handed Weapon Specialization raises all Holy damage**, not just Physical: Consecration, Exorcism, Hammer of
  Wrath, Holy Wrath and Holy Shock join the seals, judgements and Holy Strike (counted once). Notes-only value; client
  20111 is still Physical. +8 at 180 sec on twist and sub-Holy, +15 on Ret Shock.
- **A2's race base stats** (shared): Undead -2.1.
- **Ret sub-Holy vs Undead rotation**: Consecration first and Hammer of Wrath at the top of the execute phase
  (`consecrationFirst` 0 -> 1, `hammerOfWrathFirst` 3 -> 1): +8.4 / +7.0 / +4.0 at 120 / 180 / 300 sec (seed 1;
  +8.5 / +7.2 / +5.0 on seed 1000003).
- Everything else in the paladin notes was already in the sim from the 10-30 hotfix (Phase A ledger below). The paladin
  trees are unchanged in 70753, so no talent string moved.

**Per build** (Undead, 20,000 iterations, seed 1, 120 / 180 / 300 sec; before = the 2026-11-01 numbers):

| Build | Talents | Before | After |
|---|---|---|---|
| Ret twist (`paladin__level_60`) | `050050002-503-05233101201301351` (kept) | 1043.4 / 1039.3 / 1028.9 | 1095.5 / 1092.1 / 1080.3 |
| Ret sub-Holy (class default) | `0512500022-3-05233101201301351` (kept) | 1042.3 / 1038.1 / 1031.7 | 1093.3 / 1091.2 / 1085.8 |
| Ret Shock | `05125210220021--0523310120130133` (kept) | 983.0 / 974.9 / 968.5 | 1037.5 / 1030.2 / 1025.5 |
| Ret sub-Holy vs Undead | `0502520022-2-05233101201301351` (kept) | 1153.9 / 1151.4 / 1145.1 | 1222.9 / 1221.4 / 1214.0 |
| Ret twist, Crusader duty | twist | – / 1033.3 / – | 1078.0 / 1082.9 / 1075.9 |
| Ret twist, Wisdom duty | twist | – / 1024.6 / – | 1069.5 / 1075.3 / 1069.1 |
| Ret sub-Holy, Crusader duty | sub-Holy | – / 1032.1 / – | 1076.9 / 1082.8 / 1083.3 |
| Ret sub-Holy, Wisdom duty | sub-Holy | – / 1023.7 / – | 1069.2 / 1075.0 / 1075.6 |
| Ret Shock, Crusader duty | Shock | – / 974.4 / – | 1024.5 / 1029.2 / 1027.9 |
| Ret Shock, Wisdom duty | Shock | – / 965.2 / – | 1015.5 / 1020.0 / 1018.7 |

Twist and sub-Holy include the exploration's Holy Strike first (below). The default stays Ret sub-Holy: it ties twist
at 180 sec (1091.2 against 1092.1).

**Races** (each race's own rotation and gear, standards 29; 180 sec, seed 1 / seed 1000003):

| Build | Undead | Human | Dwarf |
|---|---|---|---|
| Ret twist | 1090.9 / 1090.1 | 1087.0 / 1086.6 | 1082.3 / 1081.9 |
| Ret sub-Holy | 1089.8 / 1089.3 | 1087.0 / 1087.0 | 1082.4 / 1081.7 |
| Ret Shock | 1030.2 / 1031.3 | 1026.5 / 1026.2 | 1022.4 / 1022.2 |
| Ret sub-Holy vs Undead | 1221.4 / 1221.5 | 1218.7 / 1218.5 | 1209.4 / 1209.3 |

Best overall and best Horde: **Undead** on every build (2.3 to 5.1 ahead of Human on both seed sets, so no re-pick).
Best Alliance: **Human** (Dwarf 4.0 to 9.3 behind it). Paladins can be Human, Dwarf or Undead only.
The 120 and 300 sec rows are in the scratch results and give the same order.

**Search.** Talent searches at 180 sec (3,000 iterations, refined at 10,000; Pursuit of Justice pinned, Holy Shock too
on Ret Shock) from each build's own string: no move on any build. Knob searches over 120, 180 and 300 sec together on all
eight rotations: only vs Undead moved (above); the rest kept every knob.

**Audit** (the DPS passes applied to the paladin builds):

- results: every meta's published and verify number is this round's 20,000 iteration seed 1 run on this request; the
  mksettings drift (`tools/audit/results/regen.py`) equals the reference (only the expected duty debuffs and
  `primarySeal` Command); no unread fields added.
- rotations (standards 31): every rotation was knob-searched over 120 / 180 / 300 sec with the requests' 180 ± 20 sec
  variation; the one change holds at all three lengths on two seed sets. The vs Undead meta's rotation lines now follow
  its new priority. Vengeance is up 179 of 180 sec on twist and Shock (a 1,000 iteration look), and every rotation's
  spells cast.
- gear: no gear change this round (nothing to re-audit).
- consumables: the potion slot re-checked under the new Holy multiplier: the Dragonfire Potion still beats the Major Mana
  Potion on twist (+4.8 / +5.3 / +2.5) and Ret Shock (+4.8 / +5.2 / +4.7). Nothing else moved.
- mechanics and client_update: Phase A below; both notes-only values have `TestNotesOnlyValues70753`, and
  `TestJudgementOfRighteousnessMelee` reads the specialization through the Holy school. wago.tools' hotfix log was
  re-checked before the finals: no row for 20050 or 20111 (the 70753 rows logged since are region 95 copies of push
  113139 and push 113557's creature spells).

**Exploration after landing** (2026-11-03, the user's request): talent searches at 180 sec (Pursuit of Justice pinned)
from five other starting builds, and the knob search's near misses re-tested at 20,000 iterations on two seed sets.

| Starting build | Ended at | 180 s |
|---|---|---|
| 20 Holy / 31 Ret (`05125010222--05233101201301351`) | the Ret sub-Holy talents | 1089.4 |
| Ret Shock with Champion of the Light 5 for Consecrated Ground (`05125210220001--0523310120130135`) | the Ret Shock talents | 1030.0 |
| Twist with Swift Judgement (`04-5032212001-05233101201301351`) | itself | 1028.7 (-64; out of mana at 167 sec) |
| 0 / 20 / 31 with Swift Judgement (`-5032212041-05233101201301351`) | the Swift Judgement build above | 1028.7 |
| Consecrated Ground, no Holy Shock or Unsealed Wrath (`05125210220020--0523310120130134`) | `0512521022001--0523110122130135` | 942.0 |

No new build. Holy Strike ahead of Judgement and the seal lines (`holyStrikeFirst=1`), 120 / 180 / 300 sec, seed 1 and
seed 1000003: twist +1.2 / +1.2 / +0.4 and +2.4 / +1.3 / +1.9, sub-Holy +1.0 / +1.4 / +1.5 and +2.5 / +2.2 / +2.5:
**shipped** on those two. Crusader duty -5.4 / -4.0 / -1.5 and -5.5 / -3.7 / -0.9; Wisdom duty +0.3 to +1.1, Ret Shock
+0.3 to +0.5, vs Undead +0.5 to +1.3 (under 1 at 180 sec): kept off. The race table above predates it; a rotation
change of about 1 DPS doesn't touch the 2.3+ race margins.

**Open questions and shared changes:**

- No shared code changed (only `sim/paladin/` and the paladin files). `sim/testdata/report_hashes.txt` is rewritten for the
  10 paladin lines.
- `pal_2h_spec_holy` in the report's assumptions (`tools/dps_report/assumptions_def.py`, `orchestration/spec_assumptions.md`)
  describes the old reach (seals, judgements and Holy Strike only); under the notes' value it now covers all Holy damage.
  The orchestrator owns that row; I'd reword it in phase 3.
- Salvation in the shared raid: the orchestrator's phase 3 change (question 2 below).
- If a hotfix or client build later sets Vengeance or Two-Handed Weapon Specialization to the client's values, revert the
  two gated lines in `sim/paladin/talents.go` and re-run the vs Undead knob search (its change came with the Holy reach).

## Phase A: the notes against the sim

Sources checked (2026-11-03):

- Dev notes post 9 (`orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt`), "- Paladin", "Races"
  and "Bug Fixes".
- Client 1.60.1.70753 on wago.tools (`python3 tools/sources/wago.py fetch 1.60.1.70753 --set client_update`): Spell,
  SpellEffect, SpellAuraOptions, SpellMisc, SpellLevels, SkillLineAbility.
- Server hotfixes: wago.tools has no Forever (region 91) rows logged for build 70753 yet (`hotfixes.py fetch 1.60.1.70753
  --set client_update`: 0 rows in every table; searches for 20049, 20050, 20111 and "Vengeance" find only the 10-30 push
  113139 on 70707). foreverchanges.pro/patch-notes has a 2026-11-03 hotfix, but its rows are hunter only (Aspect of the
  Hawk, Trueshot Aura, Aimed Shot, Arcane Shot, Serpent Sting, Sniper Shot, Lightning Reflexes).
- foreverchanges.pro's own list of the notes' lines (its 2026-11-03 "Blizzard's notes" entry) lists for the paladin only
  the three lines it took for notes-only: Templar's Bulwark magic, Two-Handed Weapon Specialization on Holy, Vengeance 5
  stacks. (Corrected 2026-11-04, tank_mechanics-20261104-10: Templar's Bulwark is not notes-only. 1311015 carries
  DispelType 1, Magic, in clients 70753 and 70755+hf, so it is client data.) Everything else in the paladin section was already in the 10-30 launch-realm hotfix (1.60.1.70707+hf), which
  70753 writes into the client tables (update_ledger.md, the 70753 paragraph).

Tiers: H = server hotfix, C = client 70753, N = dev notes, FC = foreverchanges.pro.

| Note line | Tier and evidence | Sim status | Fix |
|---|---|---|---|
| Righteous Fury at level 10; 33% all threat, +90% Holy; free with The Quick and the Dead | N, C | n/a (tank session) | tank/healer sessions |
| Salvation, Mass Salvation and Revoke removed | N; C (1038, 25895 no longer trained, FC 10-30) | partial: the raid buff `salvation: true` is still on in every Level 60 request (mksettings `main.go:74`, all 10 paladin requests). Threat only, 0 DPS | cross-class, for the orchestrator (see Questions) |
| Blessing of Light removed | N; FC 10-30 "no longer taught by trainers" | n/a: no request or preset uses it (`proto/paladin.proto:72` keeps the enum) | none |
| Judgement of Fury 4 → 3 sec | N, C | n/a: Seal of Fury isn't in the sim | none |
| Holy: Spiritual Focus 5 ranks, row 1 | N, C | n/a (pushback; no DPS build) | A1's tree |
| Holy: Improved Seals 5 ranks, same 15% | N, C (MaxRanks 5) | landed (`talents.go`, 3% a rank) | A1's tree remap |
| Holy: Reverence row 2, 2 ranks of 15% | N, C | landed (`talents.go:52`) | A1's tree |
| Holy: Unyielding Faith to Retribution row 3 | N, C | n/a (fear/disorient duration) | A1's tree |
| Holy: Purifying Power to row 2 | N, C | landed (`talents.go:133`, Exorcism and Holy Wrath cooldowns) | A1's tree |
| Holy: Voice of Truth removed (a trainer spell at 20, FC) | N, C (1310897 in SkillLineAbility, level 20: immune to Silence and Interrupt 6 sec) | n/a: not simmed (`concentration_aura.go:12`); only pushback from raid hits, and Concentration Aura already covers 35% | none |
| Holy: Divine Favor, Illumination to row 3; Holy Conduit moved to Holy row 3 | N, C | landed (`talents.go:118`) | A1's tree |
| Holy: Divine Precision row 4, 4 / 8% Holy hit, Silence duration | N, C | landed (`talents.go:90-96`) | A1's tree |
| Holy: Oath of Light, Boundless Light new | N, C | n/a (healing) | tank/healer sessions |
| Holy: Holy Power crit healing only | N, C | landed (no damage crit) | none |
| Holy: Light's Vigil friendly only | N, C | landed (heals only, the DPS rotations no longer cast it) | none |
| Protection: Reckoning stacks fix; Divine Intellect to Protection row 1; Sacred Duty; Improved Seal of Fury removed; Templar's Bulwark row 3 and dispellable; Improved Righteous Fury 3 / 6 / 9%; Shield Specialization 2 ranks; Toughness row 3; Anticipation row 4; Stoicism; Holy Shield row 5; Iron Creed 3 ranks; Shieldsworn | N, C (Templar's Bulwark magic: N only) | landed for the 10-30 values (update_ledger 70753 paragraph); the rest n/a for DPS | tank/healer sessions; A1's tree (Divine Intellect's move matters to the DPS builds' strings: 2% Intellect a rank, `talents.go:38`) |
| Protection: Swift Judgement unleashes the seal, own cooldown | N, C | landed (`swift_judgement.go`) | none |
| Protection: One-Handed Weapon Specialization removed | N, C | landed | none |
| Retribution: Holy Conduit, Sanctified Judgement, Conviction to row 2 | N, C | landed (rows: A1's tree) | A1's tree |
| Retribution: Conviction 2 / 4 / 6% crit, all spells and attacks | N, C (20117 aura 290, 3 ranks) | landed (`talents.go:28-29`) | none |
| Retribution: Vindication no attack power | N, C | landed | none |
| Retribution: Two-Handed Weapon Specialization row 4 | N, C | landed (row: A1's tree) | A1's tree |
| Retribution: **Two-Handed Weapon Specialization raises Holy and Physical damage by 2 / 4 / 6%** | **N and FC only. C disagrees**: 20111 effect 0 is still aura 79 misc 1 (Physical), the tooltip "Increases the physical damage you deal by $s1%"; no hotfix row since push 113139 | was as the client (Physical plus the melee-rolled Holy spells, `pal_2h_spec_holy`) | **fixed: notes-only value** (see below), `talents.go` `applyWeaponSpecialization`, `TestNotesOnlyValues70753` |
| Retribution: Two-Handed Weapon Specialization 7 / 14 / 20% less threat | N, C | landed (`talents.go:284-286`) | none |
| Retribution: Instrument of Law removed | N, C | landed | none |
| Retribution: Champion of the Light 12–60% of Intellect, 0.2–1.0 sec off Hammer of Wrath | N, C (1311084: aura 174 100 (curve), effect 1 aura 107 -1000 ms on Hammer of Wrath's mask) | landed (`talents.go:57-58`, `hammer_of_wrath.go:64`) | none |
| Retribution: **Vengeance stacks up to 5 times (was 3)** | **N and FC only. C disagrees**: 20050 SpellAuraOptions CumulativeAura 3, duration index 9 (30 sec); tooltip "$20050u times" reads it; no hotfix row | was as the client, 3 stacks | **fixed: notes-only value** (see below), `talents.go` `applyVengeance`, `TestNotesOnlyValues70753` |
| Retribution: Vengeance needs Seal of Command | N, C | n/a (tree prerequisite) | A1's tree |
| Retribution: Turn Evil new | N, C | n/a (Turn Undead) | A1's tree |
| Retribution: Twist of Light removed; Unsealed Wrath the 31-point talent, needs Repentance | N, C | landed (`unsealed_wrath.go`, `pal_unsealed_wrath`) | A1's tree (Repentance prerequisite) |
| Classes: talents lose the low-rank proc penalty | N | n/a for the paladin: the sim has no low-rank proc penalty on any paladin talent (Exorcism's downrank cut is spell power, `exorcism.go:23`, and the presets cast top ranks) | A2's shared part |
| Races: base stats fixed (Tauren +5 Strength over Human, High Order and Windshaper) | N | A2 (shared) | after A2: re-run the race sweep in Phase B |
| Bug fixes: Throw windup, root DR, quitter debuff | N | n/a | none |

The 70753 client tables against the sim's 70707+hf reading: nothing else differs for the paladin (update_ledger.md, the 70753
paragraph, checked the trees' MaxRanks, Conviction, Holy Power, Divine Precision, Champion of the Light, Swift Judgement,
Unsealed Wrath, the Echo auras and Two-Handed Weapon Specialization's tooltip and threat; I re-read 20049, 20050, 20111,
20117 and 1311084 in 70753 for this file).

Every line the client or a hotfix carries was already in the sim from the 10-30 hotfix. The two lines that move DPS are
notes-only; the user's ruling of 2026-11-03 (standards 37: an announced value that neither the client nor a hotfix
carries takes the notes' number, gated under Forever) puts them in.

## Notes-only values

**2026-11-04:** both values below are now hotfix values: push 113542 (logged 2026-11-03) sets 20050's CumulativeAura to 5
and 20111 effect 0's school mask to Physical and Holy (`hotfix_113483_paladin.md`). No DPS moves; the text below is the
original record.

Modelled from the notes, gated under Forever; client 1.60.1.70753 unchanged and no hotfix row (re-check wago.tools'
hotfix log for 20050 and 20111 before the Phase B finals; if a row lands, it wins).

1. **Vengeance stacks up to 5 times** (was 3). Notes: "Vengeance changed to stacking up to 5 times (Was stacking up to 3
   times)". Client 20050 SpellAuraOptions CumulativeAura 3. `sim/paladin/talents.go` `applyVengeance` (MaxStacks 5),
   Forever only; outside Forever `applyClassicVengeance` gives Classic's 3% a point for 8 sec, no stacks.
2. **Two-Handed Weapon Specialization raises Holy and Physical damage by 2 / 4 / 6%.** Notes: "changed to increase both
   Holy and Physical damage by 2/4/6% (Was just Physical damage)". Client 20111 effect 0 aura 79 misc 1 (Physical).
   Under Forever the Holy school takes the multiplier (Consecration, Exorcism, Hammer of Wrath, Holy Wrath and Holy Shock
   join the seals, judgements and Holy Strike), and the melee-rolled Holy spells drop their own factor so it counts
   once. `sim/paladin/talents.go` `applyWeaponSpecialization`; `TestJudgementOfRighteousnessMelee` now reads it through
   the school.

Both: `TestNotesOnlyValues70753` (`sim/paladin/retribution/hotfix_test.go`); the TestRetribution golden moved (its
builds take both talents). On the current requests (old talents and rotations, 5,000 iterations, seed 1):

| Build | Before (120 / 180 / 300 s) | After | 180 s |
|---|---|---|---|
| Ret twist | 1044.3 / 1038.8 / 1027.2 | 1096.9 / 1092.9 / 1082.0 | +54.1 (+5.2%) |
| Ret sub-Holy | 1042.6 / 1037.8 / 1031.1 | 1095.4 / 1092.3 / 1086.7 | +54.5 (+5.3%) |
| Ret Shock | 982.6 / 973.8 / 967.7 | 1039.5 / 1032.4 / 1027.5 | +58.6 (+6.0%) |

### Each value alone (throwaway builds, before the change)

Throwaway builds (not committed), current requests, 5,000 iterations, seed 1, 120 / 180 / 300 sec:

| Build | Sim now | Vengeance 5 stacks | 2H Spec on all Holy damage |
|---|---|---|---|
| Ret twist (`paladin__level_60`) | 1044.3 / 1038.8 / 1027.2 | 1088.1 / 1084.2 / 1073.5 (+4.4%) | 1052.6 / 1047.1 / 1035.3 (+0.8%) |
| Ret sub-Holy | 1042.6 / 1037.8 / 1031.1 | 1086.4 / 1083.3 / 1077.8 (+4.4%) | 1051.1 / 1046.3 / 1039.5 (+0.8%) |
| Ret Shock | 982.6 / 973.8 / 967.7 | 1023.8 / 1016.7 / 1011.8 (+4.4%) | 997.6 / 988.8 / 982.5 (+1.5%) |

The 2H variant multiplies the Holy school by 2 / 4 / 6% and drops the per-spell factor on the seals, judgements and Holy
Strike, so they still get it once.

## Questions for the orchestrator

1. **Vengeance 5 stacks and Two-Handed Weapon Specialization on Holy (notes vs client).** Settled by the user's ruling
   (standards 37, 2026-11-03): modelled from the notes, see "Notes-only values". Templar's Bulwark as a Magic effect,
   which this list called notes-only, is in the client (1311015 DispelType 1 in 70753 and 70755+hf; tank_mechanics-20261104-10).
2. **Salvation in the shared raid.** Salvation is gone from the game, but mksettings `main.go:74` and every Level 60
   request still set `raid.buffs.salvation: true` (sim/core/buffs.go:374 gives it under Forever). Threat only, 0 DPS.
   The orchestrator drops it from mksettings and every request in one shared change in phase 3; left alone here.
