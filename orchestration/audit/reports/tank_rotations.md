# tank_rotations: pass 1 (2026-11-04, master f62bfc74)

This is the first pass since the four tank specs landed (Protection warrior, Protection paladin, Feral bear, Warden
shaman), realigned to client 70753 and the November notes. It was forced (gate.py --force) and run by hand.

## What changed
Everything since the placeholder last.master 3e90308 (468 commits): all four tank presets, their APLs and requests,
the rotopt tank templates, tools/tankbench, tools/tankscore and tools/encounters. There was no earlier baseline,
so no row could have moved against one.

## Baseline (new: ref/tank_bench/summary.tsv)
Every tools/tankbench/req/*.json request on every tools/encounters/*.json boss, at 1,000 iterations with the
default seed: `tankbench -iters 1000 -encounter <boss> <req>` (about 0.2 s a run). tankbench has no TSV mode, so
tools/audit/tank_rotations/bench_summary.py parses its markdown rows. The score is the column for the request's
preset. Below is the scoring mean over the six bosses without Ragnaros:

| request | w | score | SafeDPS | TPS | DTPS | healing needed | death floor |
|---|---|---|---|---|---|---|---|
| druid_bear_balanced | 1 | 1855.0 | 1714.0 | 1699.7 | 569.5 | 582.5 | 0.0 |
| druid_bear_max_survival | 3 | -843.7 | 1598.3 | 1627.0 | 568.2 | 487.3 | 0.0 |
| druid_bear_max_threat | 0.333 | 2905.3 | 1714.8 | 1701.8 | 568.8 | 582.7 | 0.0 |
| paladin_protection_balanced | 1 | 2555.0 | 1601.8 | 1429.8 | 303.2 | 257.2 | 0.0 |
| paladin_protection_max_survival | 3 | 1257.0 | 1601.8 | 1429.8 | 303.2 | 257.2 | 0.0 |
| paladin_protection_max_threat | 0.333 | 3036.0 | 1643.3 | 1465.7 | 322.0 | 288.3 | 0.0 |
| shaman_warden_balanced | 1 | 501.3 | 1744.7 | 1589.3 | 815.5 | 1054.8 | 0.1 |
| shaman_warden_max_survival | 3 | -5325.7 | 1688.2 | 1547.8 | 782.0 | 1017.3 | 0.1 |
| shaman_warden_max_threat | 0.333 | 2552.2 | 1777.3 | 1618.7 | 822.7 | 1066.3 | 0.1 |
| warrior_protection | 1* | -13.2 | 1327.3 | 1589.3 | 657.2 | 946.8 | 0.4 |
| warrior_protection_balanced | 1 | 3795.0 | 2474.0 | 2312.7 | 378.2 | 394.5 | 0.0 |
| warrior_protection_max_survival | 3 | 1487.8 | 2474.0 | 2312.7 | 378.2 | 394.5 | 0.0 |
| warrior_protection_max_threat | 0.333 | 4708.0 | 2613.7 | 2444.2 | 459.2 | 524.2 | 0.0 |

- warrior_protection.json is the legacy request (Orc, old gear, no preset). It is kept for tests and healbench,
  marked 1* in the table and not judged.
- Balanced and Max survival are the same build for the paladin (docs/tank_paladin.md:229-237) and the warrior
  (Max survival keeps Balanced's build). The two bear presets Max threat and Balanced differ only in Maul at 20/30
  Rage (docs/tank_druid.md call 8). All three are deliberate.
- On Ragnaros (not scored) every score is negative and death runs 5–39%. The Warden's healing needed (~1,050)
  is about twice the others'. That comes from mechanics, not the rotation: Stoneskin and Lightning Shield are
  up all fight. Ruled in the docs as call (c); it belongs to tank_mechanics.
- Broodlord's death floor sits near 5% for the Warden, so paired errors there run 300–900.

## Method
Paired A/B runs with tools/audit/tank_rotations/abtest (new): two requests, the same seed and iterations, and
tankscore Objective.PairedDiff per boss plus the scoring mean. Run it from a built tree, for example
`tools/audit/build_cli.sh`'s worktree. Each B changes one line, made with tools/audit/tank_rotations/mk_<spec>.py.
1,000 iterations a row, and 3,000 for every gain under twice its error.

## Changes tried (scoring mean, paired)
| spec | preset | change | diff ± se | iters |
|---|---|---|---|---|
| bear | Bal / MT / MS | Roar below the Lacerate keep-up line | +197.5±16.7 / +207.8±11.7 / +176.9±30.0 | 1k/1k/3k |
| bear | Bal | Roar removed (breaks the 11-02 ruling) / after Primal Bite | +223.8 / +39.8 | 1k |
| bear | MS | shift -2.5 s, Enrage -1 s | +45.6±16.6 (on the Roar fix) | 3k |
| bear | Bal | in-fight Enrage; Barkskin removed; Primal Bite below Lacerate; Lacerate refresh 3 s; Maul 20 | -5 / -124 / -235 / 0 / +3±4 | 1k |
| bear | MS | Frenzied Regeneration at 60 / removed; Barkskin to the top | -253 / -432 / -785 | 1k |
| warrior | Bal / MS | Shield Wall in the last 18 s (12 s now) | +46.0±7.1 / +175±20 (Vael -107/-283) | 3k |
| warrior | Bal / MS | the same at 14 / 16 / 20 / 22 s | +22/+74, +37/+137, +20/+110, -36/-32 | 3k |
| warrior | MT | Shield Wall 18 s; Sunder at 15 Rage; Bloodrage on cooldown | -2.4±3.9; +3.9±1.1; +1.4±0.5 | 3k |
| warrior | all | Heroic Strike thresholds, Bloodrage before the pull, Shield Block on cooldown or removed | none better | 1k |
| paladin | Bal / MS / MT | Bulwark at 13 s (10 s now) | +73.3±5.0 / +219.9±14.9 / +23.1±1.5 (mostly Broodlord) | 3k |
| paladin | Bal | Bulwark at 11 / 12 / 14 / 15 / 8 / 0 s | +15 / +20 / +48 / -53 / -39 / -235 | 1–3k |
| paladin | Bal / MT / MS | autocast (trinket) line to the top | +13.9±0.6 / +14.1±0.2 / +13.4±1.8 | 3k |
| paladin | Bal | Consecration floor; Hammer of Wrath under 20%; pre-pull Judgement; reseal under 1.5 s; Holy Strike first | 0 / -99 / -115 / 0 / -23 | 1k |
| warden | MT / Bal / MS | Earth Shock above Stormstrike | +62.0±16.2 / +51.2±48.6 / -18.7±148 | 3k |
| warden | Bal | no Lightning Shield refresh; ES rank 6; LS rank 6; add MT's Mana Spring | -259 / -215 / -41 / +5±5 | 1k |

## Lines that check out
- Warrior: Shield Block keeps its buff up (~10 casts in 56 s). Shield Slam runs on cooldown, GCDs are full,
  Heroic Strike takes ~93% of swings at MT without starving a GCD, ranks are max, and the opener (Defensive
  Stance, Bloodrage, Mighty Rage Potion, trinket) is right. Charge is gone with Vanguard (settled).
- Paladin: seal, both Judgements, Holy Shield and Holy Strike all run at their cooldowns, and mana never binds.
  The opener is right.
- Bear: Primal Bite and Lacerate fill every GCD, Maul takes ~96% of swings, and Feral Charge and Faerie Fire
  aren't in Forever's sim.
- Warden: Lightning Shield and Stoneskin stay up, and Mana Spring at MT earns its place (+23).
- No scoring preset has a line that reads health.

## Findings
8 new: tank_rotations-20261104-1 to -8 (findings/tank_rotations.md). Four are high (bear Roar order, warrior
Shield Wall timer, paladin Bulwark timer (needs a user ruling over Broodlord's timers), Warden shock order), two
are medium (bear MS Enrage timing, paladin trinket line) and two are low (warrior MT Sunder threshold, notes).
Three of them come from coarse or hard-coded rotopt template knobs: warrior wallAt step 12, paladin bulwarkAt
step 10, and the fixed line order in the druid and shaman templates.
