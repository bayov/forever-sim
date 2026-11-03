# Paladin: round 70753 (the November 2 development notes)

Worker: Paladin (Retribution DPS specs). Brief: `briefs/paladin.md`. Status: **Phase A done** (with the notes-only values), waiting for "sources landed"
(A1's talent remap and the 70753 trees, A2's shared changes) for Phase B.

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
  the three lines the client doesn't have: Templar's Bulwark magic, Two-Handed Weapon Specialization on Holy, Vengeance 5
  stacks. Everything else in the paladin section was already in the 10-30 launch-realm hotfix (1.60.1.70707+hf), which
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
notes-only; the orchestrator's provisional rule of 2026-11-03 (`README.md`: model a notes value that neither the client
nor a hotfix carries, gated under Forever) puts them in.

## Notes-only values

Modelled from the notes, gated under Forever; client 1.60.1.70753 unchanged and no hotfix row (re-check wago.tools'
hotfix log for 20050 and 20111 before the Phase B finals; if a row lands, it wins).

1. **Vengeance stacks up to 5 times** (was 3). Notes: "Vengeance changed to stacking up to 5 times (Was stacking up to 3
   times)". Client 20050 SpellAuraOptions CumulativeAura 3. `sim/paladin/talents.go` `applyVengeance` (MaxStacks 5).
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

1. **Vengeance 5 stacks and Two-Handed Weapon Specialization on Holy (notes vs client).** Answered by the orchestrator's
   provisional rule (2026-11-03): modelled from the notes, see "Notes-only values". If the user rules the client in,
   revert both (one commit) and re-run Phase B's searches.
2. **Salvation in the shared raid.** Salvation is gone from the game, but mksettings `main.go:74` and every Level 60
   request still set `raid.buffs.salvation: true` (sim/core/buffs.go:374 gives it under Forever). Threat only, 0 DPS.
   Recommendation: drop it in mksettings and the requests in one cross-class change (the orchestrator's), and leave
   `SalvationAura` for old requests or gate it off. I'll leave my requests' flag as mksettings writes it.
