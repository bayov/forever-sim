# Shaman: round 70753 (the November 2 development notes)

Branch `round-70753-shaman`, landed on master at 1e51d24. Phase A (mechanics) and Phase B (tree, re-optimization, races, audit)
are done.

## Summary

**What changed** (all from Blizzard's notes, post 9: client 1.60.1.70753 has none of the shaman lines and no server
hotfix has them, so standards 37 applies; the per-line ledger, the DPS of each change and the notes-only list are
below):

- Lightning Bolt ranks 4–10 at 0.78 of Era's base (rank 10 +76%; the ×1.12 assumption is now 1.0): Elemental +79,
  Enhancement +11 at 180 s.
- Every shock at Era's base (the ×1.08 on Earth Shock 7 is now 1.0): Elemental +28, Enhancement +45 to +58.
- Lightning Overload 10/20/30%: Elemental +48.
- Stormstrike 150% and 6 s: Enhancement +105 to +117. Maelstrom Weapon ×2 (4 PPM a point): +30 to +36. Mental
  Dexterity 150%: +26. Flurry 30%: +12 to +14.
- Chain Lightning 2.5 s, 1.45 × the comparable bolt, 2.25 × its mana: 0 (still out of the rotations).
- Lava Burst removed and Call of Thunder (+5% crit) in its square; Elemental Mastery → Lightning Mastery; Improved
  Ghost Wolf to row 4 (`tools/forever_talents/notes/shaman.json`, every Enhancement string remapped); Totemic
  Projection → Totemic Call (30 s, totems to our spot).
- Re-optimized talents and rotations (below), races unchanged.

**Per build** (20,000 iterations, seed 1, master's request on master's binary → this branch, both with A2's base
stats; race unchanged):

| Build | Talents | Race | 120 s | 180 s | 300 s |
|---|---|---|---|---|---|
| Elemental Level 60 | `5501331003001351-05500101-053` (Call of Thunder in Lava Burst's square) | Windshaper Skyborne | 775.3 → 979.7 | 767.6 → 970.1 | 738.8 → 924.6 |
| Elemental Mana Tide | `5501331003001351-04-053150001001` | Windshaper Skyborne | 760.2 → 962.7 | 752.8 → 952.7 | 737.0 → 935.1 |
| Enhancement Level 60 | `053503102-055030110051012251` → `353003102-055030110051012251-02` | Orc | 892.8 → 1169.0 | 883.0 → 1159.2 | 866.7 → 1140.5 |
| Enhancement Imp GW | `053503102-054030010051212251` → `353003102-055030110051212151-01` | Orc | 889.0 → 1162.2 | 878.8 → 1149.7 | 862.1 → 1128.1 |
| Enhancement Water Shield | `051003-054030210051012151-053020001` → `052002-054030210051012151-053020001` | Orc | 842.9 → 1118.9 | 834.2 → 1111.3 | 830.4 → 1108.1 |
| Enhancement Mana Tide | `03-055030310051012051-053051001001` | Orc | 803.4 → 1065.1 | 795.4 → 1055.1 | 789.4 → 1046.7 |

(The "old" strings are in the new tree's order; master's had Improved Ghost Wolf earlier.)

Rotation changes (each positive or even at 120 / 180 / 300 s on seeds 1 and 1000003, paired at 20,000):
Enhancement Optimized `earthShockBeforeBolt=0` (Level 60 +0.9 to +3.1, Imp GW +0.8 to +4.2); Magma (Water Shield)
`flameShockFirst=0, prepullFireLead=1.5` (0 to +1.3); Elemental Level 60 `lightningBoltDownrank=1` (0 / 0 / +5.0 to
+5.3); the Lava Burst lines and knob are gone. Mana Tide builds keep their rotations. Talents judged at 180 s: the
Level 60 search's `443003102-055030110051012151-03` ties the shipped build (+2.5 at 120, −2 to −3 at 300, equal mean).

**Best race** (180 s, 20,000 iterations, two seed sets, each race's own rotation and gear): Elemental — overall and
Horde Windshaper Skyborne (970.1; Orc 965.0, Troll 962.4, Tauren 955.5), Alliance Dwarf (961.4). Enhancement —
overall and Horde Orc (1159.2; Windshaper Skyborne 1154.7, Troll 1150.8, Tauren 1144.0), Alliance Dwarf (1150.1).
The other builds rank the same. The Elemental Troll overrides were remade (their rotation still had Lava Burst):
Berserking after Lightning Mastery is +6.3 / +6.2.

**Audit** (my pass over the DPS auditors' checks, details under "Audit" below): rotations clean (casts, cooldowns,
shock and Stormstrike usage checked on breakdowns); one open mechanics question (Chain Lightning's coefficient);
gear and consumables unchanged and still fitting; client_update covered by the ledger.

**Open questions and shared changes:** under "Questions for the orchestrator" and "Shared files touched".

## Sources

- **Client 1.60.1.70753 has none of the shaman lines.** `db2diff.py 70707 70753`, Shaman section: 173 spells changed,
  all but two only by `SpellMisc.Attributes_14` 0 → 64. The two others are Skyborne's Charged (1270893) and one
  Attributes_2 bit on 326658, neither DPS. Lightning Bolt (15208 still 196 ± 5.4% + 1.2 a level), Chain Lightning, the
  shocks, Lightning Overload (408438), Mental Dexterity (415140), Flurry, Stormstrike (17364), Lava Burst, Elemental
  Mastery (16166) and Totemic Projection (437009) are all 70707's rows.
- **No server hotfix for the shaman.** `tools/sources/hotfixes.py fetch 1.60.1.70753 --set client_update` finds no
  rows on 70753. foreverchanges.pro's "November 3 hotfix" (2026-11-03) changes only the hunter. Its shaman lines
  restate Blizzard's notes.
- **Wowhead Forever** (tooltip API, dataEnv 16): still the beta's values (Lightning Bolt rank 10 "191 to 213" at 56).
- So for every shaman line the **development notes (post 9) are the highest tier there is**. Where the notes give a
  rule rather than a number, the numbers come from Classic Era's client (1.15.9.69722 on wago.tools) as the notes say.

## Ledger

DPS: paired 10,000 iterations, 180 s ± 20, seed 1. Every change was added alone to master b648ac704 (the Phase A base),
on the Phase A presets: Lava Burst and the old talent strings are still in, and so is the 1.12 / 1.08 rank tuning
for "before". Columns: Elemental Level 60 / Elemental Mana Tide / Enhancement Level 60 / Imp GW / Mana Tide /
Water Shield.

| Note line | Tier and evidence | Sim status before | Fix | DPS at 180 s |
|---|---|---|---|---|
| Windfury Weapon animates an extra swing | notes; cosmetic | n/a | none | 0 |
| Windfury Totem animates an extra swing | notes; cosmetic | n/a | none | 0 |
| Totemic Projection → Totemic Call: teleports totems to the shaman's own location, 30 s cooldown | notes; 437009 unchanged in 70753 | wrong (1 min; each totem kept its offset) | `totem_reach.go`: 30 s, every totem to our spot. Spell id 437009 kept; `tools/rotopt/spell_names.go` names it Totemic Call. Test `TestTotemicCall` | 0 (leveling only) |
| Lightning Bolt: every rank up, max rank +80%, each rank about 78% of Classic Era's base | notes + Era client (15208: 419–467 at 56, 2.4 a level) | wrong (190–212 at 60, ×1.12 by the `sham_lb_31plus` ruling) | `lightning_bolt.go`: ranks 4–10 at 0.78 × Era (base points, die and growth per level). Rank 10 is 334–372 at 60 (+76% over the client, the notes' "80%"). Ranks 1–3 already deal 1.07 / 0.96 / 0.92 of Era, more than 0.78, and the notes give no number, so they keep the client's. `LightningBoltRank10DamageScale` 1.12 → 1.0 | +79 / +78 / +11 / +11 / +11 / +12 (net of dropping the ×1.12) |
| Chain Lightning: 2.5 s cast, first target about 145% of comparable-rank Lightning Bolt, mana about 225% | notes | wrong (2 s, 119–133, 315 mana) | `chain_lightning.go`: "comparable rank" taken as the newest Lightning Bolt at Chain Lightning's learn level (ranks 6, 7, 8, 10 at 32 / 40 / 48 / 56). 1.45 × its range and growth: rank 4 485–539 at 60. Mana 2.25 ×: 248 / 304 / 360 / 495. Cast 2.5 s, cooldown still 6 s. Tests `TestOrcShamanRankTuning`, `TestOrcShamanChainLightningJumps` | 0 (no preset casts it on one target; re-checked in Phase B: still −25 to −80 with any mana floor) |
| Shocks: base damage reverted to Classic Era | notes + Era client (10414 517–545, 29228 292 + 4 × 80, 10473 492–520 at 60) | wrong (Forever's 0.55–0.57 of Era; Earth Shock 7 ×1.08 by `sham_es7_tuning`) | `earth_shock.go`, `flame_shock.go`, `frost_shock.go`: Era base points, die and growth per level for every rank, Flame Shock's DoT too. Coefficients, costs and cooldowns stay Forever's. `EarthShockRank7DamageScale` 1.08 → 1.0. Tests `TestOrcShamanRankTuning`, `TestOrcShamanSpellPowerBeta`, `TestOrcShamanFlameShockTicks` | +28 / +28 / +58 / +58 / +45 / +53 |
| Lightning Overload 10 / 20 / 30% (was 3 / 7 / 10%) | notes (408438 still 10% in 70753) | wrong | `lightning_overload.go`. Test `TestOrcShamanLightningOverload` | +48 / +48 / 0 / 0 / 0 / 0 |
| Lava Burst removed, Call of Thunder (+5% crit Lightning Bolt and Chain Lightning) in its place | notes; 16120 in the client already says 5% | missing | landed (Phase B): `notes/shaman.json`, `talents.go` `applyCallOfThunder`, Lava Burst removed from the sim and the rotations; test `TestOrcShamanCallOfThunder` | Elemental +39 at 180 s on the old talents (969.8 against 931.1 with Lava Burst) |
| Elemental Mastery renamed Lightning Mastery | notes | landed (effect unchanged: Lightning Bolt 0.5 s faster, Chain Lightning no cooldown) | landed (Phase B): renamed, needs Call of Thunder, Lava Burst out of it; test `TestOrcShamanLightningMastery` | 0 |
| Elemental Reach no longer reduces Totemic Projection's cooldown | notes | landed (the sim never modelled the cut) | none | 0 |
| Improved Ghost Wolf moved from row 2 to row 4 | notes | tree | landed (Phase B): row 4, column 4 assumed; every Enhancement string remapped, Imp GW re-searched, the leveling route's points to 25–26 | Imp GW now 9.5 behind Level 60 at 180 s (4.2 before) |
| Maelstrom Weapon: chance to trigger doubled | notes; 408498 ProcChance 100, no PPM row, so the rate is server side | wrong (2 PPM a point, report assumption `sham_mw_ppm` 10 at 5/5) | `talents.go`: 4 PPM a point, 20 at 5/5. Totem of the Storm's bolt chance doubles with it. Test `TestOrcShamanRelics` | 0 / 0 / +36 / +36 / +30 / +36 |
| Stormstrike 150% normalized weapon damage, 6 s cooldown | notes | wrong (100%, 8 s) | `stormstrike.go`: 1.5 × (weapon roll + normalized AP part); a flat "+N damage" stays 1 ×. Tests `TestOrcShamanWeaponDamage`, `TestOrcShamanStormstrikeOutcomes` | 0 / 0 / +105 / +105 / +117 / +117 |
| Mental Dexterity 50 / 100 / 150% of Intellect as AP | notes (415140 still 33/67/100 curve) | wrong | `talents.go`. Test `TestOrcShamanMentalDexterity` | 0 / 0 / +27 / +26 / +26 / +26 |
| Flurry 6 / 12 / 18 / 24 / 30% | notes | wrong (5–25%) | `talents.go`. Test `TestOrcShamanHaste` | 0 / 0 / +13 / +14 / +12 / +12 |
| Riptide | notes | n/a | healing session | – |
| Talents no longer have a low-rank proc penalty | notes | n/a for the shaman (no shaman talent applies one in the sim) | A2 (shared part) | 0 |
| Races: Dwarf Artisan's Precision, Orc Savage removed and Blood Fury 2 min, Troll Berserking 15 s, Skyborne Wind-Blessed Haste 2%, Tauren Endurance, base stats fix | notes | mostly landed in the 10-26 racial pass; base stats are A2's | A2 | – |
| Threat and healing lines | – | n/a | tank / healer sessions | – |

All changes together (the same runs): Elemental Level 60 767.5 → 931.1 (+163.6), Elemental Mana Tide 752.6 → 915.0
(+162.4), Enhancement Level 60 883.4 → 1144.3 (+260.9), Imp GW 879.2 → 1139.0 (+259.8), Mana Tide 795.3 → 1056.4
(+261.1), Water Shield 835.1 → 1109.7 (+274.6). These are before the tree changes and before any re-optimization.

## Notes-only values (standards 37)

Every value below comes from the notes alone; client 1.60.1.70753 unchanged (its rows are 70707's), no server hotfix.
The next hotfix check and the launch logs compare against this list first.

| Value | Sim | Client 1.60.1.70753 |
|---|---|---|
| Lightning Bolt ranks 4–10 base | 0.78 × Era: rank 10 334.31–371.75 at 60, +1.87 a level | 15208 196 (± 5.4%) + 1.2 a level at 56 |
| Chain Lightning ranks 1–4 first target | 1.45 × the comparable Lightning Bolt: rank 4 484.75–539.04 at 60 | 10605 unchanged (119–133) |
| Chain Lightning mana | 248 / 304 / 360 / 495 (2.25 ×, rounded to whole mana) | 145 / 200 / 250 / 315 |
| Chain Lightning cast | 2.5 s | 2 s |
| Earth / Flame / Frost Shock base | Era's: 517–545, 292 + 4 × 80, 492–520 at 60 (every rank) | 301 / 166 + 4 × 44 / 283 centres |
| Lightning Overload | 10 / 20 / 30% | 408438 3 / 7 / 10% |
| Maelstrom Weapon | 4 PPM a point (doubled from the sim's 2; the rate is server side either way) | 408498 ProcChance 100, no PPM row |
| Stormstrike | 150% normalized weapon damage, 6 s | 17364 100%, 8 s |
| Mental Dexterity | 50 / 100 / 150% | 415140 33 / 67 / 100% |
| Flurry | 6 / 12 / 18 / 24 / 30% | 5 to 25% |
| Totemic Call | 30 s, totems to the shaman's spot | 437009 Totemic Projection, 1 min |

Gating (finding client_update-20261103-2): every value above is under `Env.IsForever()`. Outside Forever, Lightning Bolt
ranks 4–10 take Era's (Classic's) full damage, Chain Lightning Classic's damage and mana, Flurry Classic's 10–30%; the
shocks' Era bases are Classic's already. Lightning Overload, Maelstrom Weapon, Mental Dexterity and Totemic Call don't exist
in Classic, so outside Forever they keep the client's (or the sim's pre-notes) value. Stormstrike's cooldown and 150% sit
inside its `IsForever()` branch.

## Shared files touched

- `tools/rotopt/spell_names.go`: 437009's display name.
- `tools/dps_report/patches/0004-...patch`: its `sham_mw_ppm` hunk's removed and default lines now read
  `4*float64(...)`, so the patch still applies to `sim/shaman/talents.go` (`git apply --check --include='sim/shaman/*'`
  passes). The `sham_lb_31plus` and `sham_es7_tuning` hunks apply unchanged; their constants are now 1.0.
- `sim/testdata/report_hashes.txt`: rewritten after the last merge of master (only the six shaman lines moved).
- `tools/rotopt/talents_test.go`, `tools/rotopt/bis_lists.sh`, `tools/rotopt/leveling/level30_enh_solo.json`: the
  new tree's strings, no Lava Burst knob.
- `ui/warden_shaman/presets.ts`, `ui/restoration_shaman/presets.ts`, `sim/shaman/warden/warden_test.go`: talent
  strings remapped to the new tree (question 5).
- `notes.md`: the shaman Need to Verify rows the notes settle or open.
- `sim/shaman/warden/TestWardenShaman.results` (the tank session's golden): moved by the shared Stormstrike, Flurry,
  shock and Mental Dexterity changes, regenerated.

## Phase B notes

- Talent tree: `tools/forever_talents/notes/shaman.json` (remove Lava Burst, add Call of Thunder on its square, rename
  Elemental Mastery and point its prerequisite at Call of Thunder, move Improved Ghost Wolf to row 4 column 4, and the
  new tooltips), then `import_db2.py`, `import_client.py`, `import_talents.py`; the proto message keeps the sim's
  `improved_fire_nova` name. Every Enhancement string was remapped mechanically (Improved Ghost Wolf from position 6 to
  13): tests, presets (the Warden's and the Restoration healer's too), the report requests, the Level 30 leveling
  settings and BiS list.
- Sim: Lava Burst removed (spell, Call of Flame, Lightning Mastery, the Spiritcaller's Storm 5 piece now a no-op),
  Call of Thunder and Lightning Mastery added, with tests (`TestOrcShamanCallOfThunder`, `TestOrcShamanLightningMastery`).
- Leveling (the user's route, `tools/rotopt/aplkit/shaman_enh_leveling.go`): the template builds and runs; its Lava
  Burst knob is gone. What changed for a later re-optimization: Improved Ghost Wolf can't come before 15 Enhancement
  points, so the route's two points moved from levels 15–16 to 25–26 (`leveling_plan_audited.json`, the plan's
  required-by level 16 → 26); Stormstrike every 6 s at 150%, Maelstrom Weapon twice as often, Flurry and Mental
  Dexterity up, the shocks much stronger at every rank, Chain Lightning slower and dearer, Totemic Call every 30 s to
  our spot. `tools/rotopt/talents_test.go` uses a legal 26-point build now (the Level 30 + 5 string has 13 points
  above row 4 under the new tree).

## Audit

- **Rotations** (paired 20,000 on two seed sets, 120/180/300 s). Elemental: Flame Shock up, overloads on about 30%
  of bolts, Lightning Mastery 1.5 times a 180 s fight, Lightning Shield 85% uptime, the end-of-fight Earth Shock, no
  idle time; Chain Lightning with every mana floor (40/60/80%, and only under Lightning Mastery) loses 25 to 80 at
  every length, because its hit is only 1.09 × a bolt's at the client coefficient (question 4). Enhancement:
  Stormstrike 29.6 casts in 180 s (30 possible), the shared shock cooldown 28.4 of 30, Maelstrom Weapon up 89% with a
  bolt at 5 stacks, Flurry 78%, Rage of the Farseer with the trinket. No findings past the shipped knob changes.
- **Mechanics / client_update**: every note line has a ledger row; 70753 and its (empty) hotfix rows checked.
  Finding: Chain Lightning's coefficient is unannounced (question 4).
- **Gear**: the synthetic sets didn't change; the Elemental Dwarf hit override still applies (Artisan's Precision 1%
  over the 16% cap), Enhancement's hit isn't over its cap for any race.
- **Consumables**: unchanged; Elemental Level 60 still ends a 180 s fight with mana to spare (2192), so the damage
  potion stays.
- **Results**: the six requests' 180 s numbers match the finals above (the same requests and binary); `meta.published`
  is left for the report run.

## Exploration after landing (2026-11-03)

A wider search for builds and variants the round's searches (all seeded from the presets) could have missed. Nothing
new made the cut; the presets stand.

- **Talents** (talent search at 180 s, knob search at 120 / 180 / 300 s, talent search again; Spirit Weapons pinned on
  Enhancement; 3,000 iterations):
  - Enhancement seeded with 24 Elemental points (Lightning Overload 3, Call of Thunder, no Rage of the Farseer,
    `55303310000013-05503011005101212`) and with Reverberation 5 for the Era shocks (`55050310-055030110051012251`):
    both climb back to `443003102-055030110051012151-03` (1160.1), the build that ties the shipped Level 60.
  - Enhancement seeded with 19 Restoration points (`-055030110051012251-0530510014`): stuck at
    `0530031-055030310051012151-051`, 1133.2 (26 behind Level 60, pinning no talent, so no preset).
  - Elemental seeded with more Enhancement points (Improved Lightning Shield 3, Shamanistic Focus,
    `5501331003001351-05500301-051`) and with Reverberation 5 for Eye of the Storm (`5505331000001351-0550010-053`):
    both climb back to the shipped `5501331003001351-05500101-053` (970.2).
- **Rotation pairs the one-knob search can't reach** (5,000 iterations, the same seed as the preset, 120 / 180 / 300 s):
  - Enhancement: Lightning Bolt at 4 Maelstrom Weapon stacks -29 (-27 with the Stormstrike hold); hard-cast bolts at
    40 / 60 / 80% mana -8 to -54; Earth Shock held up to 2 s for the mark -4 to -5; a 20% shock mana floor -1 to -32;
    Searing Totem -15 to -30; no Fire Nova -50 to -69, Fire Nova ahead of the shocks -1 to -2; Earth Shock only -37;
    Flame and Frost Shock -15 to -17; Earth Shock only with 3 s of Flame Shock left -9 to -14. The one gain,
    `earthShockHold=1` (Earth Shock waits up to 1 s for the Stormstrike mark), is +0.3 / +0.4 / +0.3: under the 1 DPS
    bar. Shipped afterwards on the user's call (the `optimized` APL, Level 60 and Imp GW): paired at 20,000
    iterations on seeds 1 / 1000003, Level 60 +0.1 / -0.7, +0.2 / -0.7, 0.0 / 0.0 and Imp GW -0.4 / +0.2,
    +0.1 / +0.5, +0.6 / +1.3 at 120 / 180 / 300 s (a wash on Level 60, about +0.4 on Imp GW).
  - The user's follow-ups on the Stormstrike mark (5,000 iterations, seed 1, paired with the shipped APL, Level 60 and Imp GW,
    120 / 180 / 300 s). The Maelstrom bolt waiting for the mark on top of the Earth Shock hold (`maelstromHold=1`): -0.6 to
    +0.6 on Level 60, -1.0 to -1.8 on Imp GW; `maelstromHold=2` -2.5 to -5.6. No Earth Shock while the mark is up and
    Maelstrom Weapon has 3+ stacks (the new `earthShockSaveMark` knob) -25 to -34, at 4+ stacks -12 to -17 (-12 to -17 with
    the bolt hold too, -10 to -15 without the Earth Shock hold), at 5 stacks no change (the instant bolt is ahead of Earth
    Shock anyway). Forever's mark is used by the next spell, so the "last charge" and "whenever up" variants are the same.
    Nothing shipped.
  - Elemental: Earth Shock on the shock cooldown -10 to -23 (-10 to -37 with 40 / 60% mana floors); Earth Shock in place
    of Flame Shock -66 to -93 (Era's Flame Shock DoT carries 112 DPS); no Lightning Shield -3 to -9; no Searing Totem
    -20 to -28; no prepull bolt -5 to -12; no end-of-fight Earth Shock -4 to -11; the Flame Shock end gate at 0 or
    12 s -0.1 to -3.4.
- **Chain Lightning at a 0.714 coefficient** (question 4's other reading, a throwaway build): it still loses on one
  target, on cooldown -13 / -77 / -141, under Lightning Mastery only -15 / -19 / -75, with 40 to 60% mana floors -5
  to -42 at 120 / 180 / 300 s. Mana, not damage per cast, keeps it out, so the coefficient question can't change the
  Elemental rotation.

## Questions for the orchestrator

1. **Three report assumptions are now settled by the notes**: `sham_lb_31plus` (×1.12, the notes retuned Lightning
   Bolt's top ranks), `sham_es7_tuning` (×1.08, the shocks are Era's now) and `sham_mw_ppm` (10 PPM Expected; the
   notes double it, so the sim's value is 20). Recommendation: drop the first two from `assumptions_def.py` and the
   0004 patch, and move `sham_mw_ppm`'s settings to 10 / 20 / 30 (Low / Expected / High) or make it one value at 20.
2. **Lightning Bolt ranks 1–3** keep the client's values (already above 0.78 of Era) though the notes say "all ranks
   increased". Leveling only. Recommendation: keep until a client or hotfix has numbers; a notes.md "Need to Verify"
   row is in `notes.md`.
3. **Chain Lightning's "comparable rank"** is our reading (the newest Lightning Bolt at Chain Lightning's learn level).
   Only rank 4 matters at 60, and there it's unambiguous (both learned at 56).
4. **Chain Lightning's spell power coefficient** (open, `notes.md`). The notes raise the cast to 2.5 s but name no
   coefficient, so the sim keeps the client's 0.571 (2 / 3.5). If Blizzard moved it with the cast to 0.714 (2.5 / 3.5,
   as Lightning Bolt), Chain Lightning would hit for about 1.27 × a bolt and might enter Elemental's rotation.
   Recommendation: keep the client's value (standards 37) and check the first launch log. Low stakes: at 0.714 it
   still loses on one target (see the exploration above).
5. **Other sessions' files** I had to touch because the tree moved: the Warden and Restoration presets' strings, the
   Warden golden, and the Restoration healer preset, which now has Improved Ghost Wolf 2 on row 4 with only 8
   Enhancement points (illegal in game). The healing session should pick new points.
6. **Report assumptions** (question 1): `sham_lb_31plus` and `sham_es7_tuning` are 1.0 in the sim now; the meta
   caveats that described them are replaced.
