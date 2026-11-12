# tank_mechanics pass 12 (2026-11-12, master 55cd92b09, client 1.60.1.70888+hf)

Forced re-audit after the tank fixes (`git diff --stat cb0d928fd HEAD`; local branch `tankheal` at 55cd92b09 treated
as master). Result: 2 new findings.

**Getting current:** wago's latest is still 1.60.1.70888 (no client diff since pass 11). Blizzard's Nov 11 hotfixes
(t/2403040 post 8) are recorded by mechanics as 20261112-1/-2/-3; their reach on the tanks is below.

**Changed code since 71994fca6:** only `sim/druid/healing.go` (Restoration healer, no tank spell). The tank changes are
presets, APLs and `tools/rotopt/shaman_tank.go` (a potion knob). So the pass checked the mechanics the presets now use
that they did not before.

## New picks against the client (tier 1, 70888)

| Spec / preset | Pick | Client | Sim | Verdict |
|---|---|---|---|---|
| Warden Max threat | Shamanistic Focus 1/1 | 1223030 A_ADD_PCT_MODIFIER cost -45%, mask 0x90100000 (shocks, Lightning Shield) | x0.55 on shocks.go:23 and lightning_shield.go:126 | match; tested (mana_costs_test, warden forever_test:201) |
| Warden Max threat | Elemental Devastation 3/3 | trait ranks 3/6/9% melee crit, 10 s; buff 30165 | 3% x rank, 30165 under Forever, 10 s; spell-damage crits only (Searing bolts ProcMaskEmpty, Fire Nova hits procs) | match; tested (talent_procs_test, overload_procs_test) |
| Warden Max threat | Elemental Warding 2/3 | trait ranks 3/7/10 (fire, frost, nature; aura 87 misc 28) | 0.03/0.07/0.10 | match; tested (elemental_warding_test) |
| Warden (Balanced, Max survival) | Improved Reincarnation 2/2 | +2/+4% max health | x(1+0.02 rank) | match |
| Warden Max survival | Healing Stream Totem r5 10463 | 80 mana; 10461 heals 11 per 2 s, coefficient 0.022 | water_totems.go: 11, 0.022, 80, 5% fixed crit | match; fixed-size heals in the replay (tank_score.go:114-121), no health read |
| Warden Max threat | Searing Totem and Fire Nova all fight | rows as pass 4 | unchanged | match |
| warrior, bear, Warden | Flask of the Titans | 17626 aura 34 +1200 health | +1200 Health | match |
| paladin | Flask of Supreme Power | 17628 +150 spell damage | +150 SpellPower | match |
| paladin | Wizard Oil | 24 spell power | +24 (and 24 healing, Forever) | match |
| Warden Max threat, warrior/bear MT/Bal | Flask of Natural Precision | 1293742: 30 Stamina, 2% expertise in the zone | 30 Stamina; 2% in NaturalFlaskZone | match (supersedes pass 4's 60 / 5%) |

## Nov 11 hotfixes (mechanics-20261112-1/-2/-3) against the tank presets

| Finding | Touches a tank preset? | Measured (tankbench 3,000 it, seed 1, paired) |
|---|---|---|
| -1 Holy Strike / Windfury Totem | Option (a) no: no tank casts Holy Strike. Option (b), yellow attacks proc it: yes, warrior, bear, Warden (all take the raid totem) | warrior +6.3 to +9.5%, bear +16 to +21% (Max survival -128 -> 69), Warden +2.6% at Max threat; new 20261112-2 |
| -2 Unsealed Wrath echoes | No: no paladin tank preset takes Unsealed Wrath | - |
| -3 Judgement ranged | Yes, though it names only Command and Righteousness: every Protection paladin preset judges Seal of Fury, and Judgement of Fury (seal_of_fury.go:97-114) is a melee special | proc mask alone: Max threat -8, Balanced -9, Max survival -1 objective; TPS -3; new 20261112-1 |

Objective at each preset's own weight, stock -> option (b):

| Preset | Stock | WF on yellow | TPS |
|---|---|---|---|
| warrior Max threat | 3568 | 3794 | 1898 -> 2015 |
| warrior Balanced | 2893 | 3081 | 1806 -> 1905 |
| warrior Max survival | 1271 | 1392 | 1645 -> 1734 |
| bear Max threat | 2411 | 2802 | 1360 -> 1569 |
| bear Balanced | 1650 | 2005 | 1295 -> 1498 |
| bear Max survival | -128 | 69 | 1078 -> 1197 |
| Warden Max threat | 2177 | 2234 | 1226 -> 1258 |
| Warden Balanced | 924 | 983 | 1014 -> 1044 |
| Warden Max survival | -1310 | -1270 | 1129 -> 1154 |

The Warden rows include Stormstrike procs, which our Nov 5-8 logs rule out; the warrior and bear rows are mostly
Heroic Strike / Maul (on-next-swing, untested in logs) and Revenge / Shield Slam / Sunder.

Patches (not applied): `$AUDIT_WORK/tank_mechanics/wf_yellow.patch`, `$AUDIT_WORK/tank_mechanics/jof_ranged_procmask.patch`.

## Engine and tests

No sim/core tank file, tank_score, tank_sheet, stances.go or bear.go change in range; pass 7's engine checks stand.
`go test -tags=with_db` on sim/shaman/..., sim/paladin/..., sim/druid/..., sim/warrior/...: all 12 packages pass
(TestP1DPSWarrior's goldens were refreshed in range).

## Findings

New: 20261112-2 (high, user ruling: WF-on-yellow size on the tanks), 20261112-1 (low: Judgement of Fury still melee).
Still open: 20261108-1 (no shock-under-mark threat test), 20261107-1 (stale "client disagrees" labels in
docs/tank_warrior.md:31, docs/tank_druid.md:22/55/425, sim/druid/bear.go:57).
# Tank mechanics, pass 12

Master 8306f1d52 (from a0b4d3be9). Client 1.60.1.70888, unchanged since pass 11.

## What changed

| Commit | File | Tank relevance |
|---|---|---|
| c380bbaef | sim/druid/rake.go | None. Rake's direct hit now adds 6.5% AP under Forever (user ruling 2026-11-12). Rake is a Cat Form ability; no bear preset, APL (`sim/druid/tank`, `ui/*tank*`) or tank test casts it. Bear code mentions Rake only in a Lacerate comment (Rend and Tear) and Savage Fury's spell list, neither of which this change touches. |

## Engine and specs

No engine, warrior, paladin, shaman or bear file changed. Verdicts in `ref/tank_mechanics/verdicts.md` stand.

## Findings

None new. Still open: 20261111-1, 20261108-1, 20261107-1 (their files unchanged on master).
