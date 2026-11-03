# Shaman: round 70753 (the November 2 development notes)

Branch `round-70753-shaman`. Phase A (mechanics that don't need the new talent tree) is done. Phase B waits for A1's
talent remap and A2's shared changes.

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
on the current presets: Lava Burst and the old talent strings are still in, and so is the 1.12 / 1.08 rank tuning
for "before". Columns: Elemental Level 60 / Elemental Mana Tide / Enhancement Level 60 / Imp GW / Mana Tide /
Water Shield.

| Note line | Tier and evidence | Sim status before | Fix | DPS at 180 s |
|---|---|---|---|---|
| Windfury Weapon animates an extra swing | notes; cosmetic | n/a | none | 0 |
| Windfury Totem animates an extra swing | notes; cosmetic | n/a | none | 0 |
| Totemic Projection → Totemic Call: teleports totems to the shaman's own location, 30 s cooldown | notes; 437009 unchanged in 70753 | wrong (1 min; each totem kept its offset) | `totem_reach.go`: 30 s, every totem to our spot. Spell id 437009 kept; `tools/rotopt/spell_names.go` names it Totemic Call. Test `TestTotemicCall` | 0 (leveling only) |
| Lightning Bolt: every rank up, max rank +80%, each rank about 78% of Classic Era's base | notes + Era client (15208: 419–467 at 56, 2.4 a level) | wrong (190–212 at 60, ×1.12 by the `sham_lb_31plus` ruling) | `lightning_bolt.go`: ranks 4–10 at 0.78 × Era (base points, die and growth per level). Rank 10 is 334–372 at 60 (+76% over the client, the notes' "80%"). Ranks 1–3 already deal 1.07 / 0.96 / 0.92 of Era, more than 0.78, and the notes give no number, so they keep the client's. `LightningBoltRank10DamageScale` 1.12 → 1.0 | +79 / +78 / +11 / +11 / +11 / +12 (net of dropping the ×1.12) |
| Chain Lightning: 2.5 s cast, first target about 145% of comparable-rank Lightning Bolt, mana about 225% | notes | wrong (2 s, 119–133, 315 mana) | `chain_lightning.go`: "comparable rank" taken as the newest Lightning Bolt at Chain Lightning's learn level (ranks 6, 7, 8, 10 at 32 / 40 / 48 / 56). 1.45 × its range and growth: rank 4 485–539 at 60. Mana 2.25 ×: 248 / 304 / 360 / 495. Cast 2.5 s, cooldown still 6 s. Tests `TestOrcShamanRankTuning`, `TestOrcShamanChainLightningJumps` | 0 (no preset casts it on one target; to re-check in Phase B) |
| Shocks: base damage reverted to Classic Era | notes + Era client (10414 517–545, 29228 292 + 4 × 80, 10473 492–520 at 60) | wrong (Forever's 0.55–0.57 of Era; Earth Shock 7 ×1.08 by `sham_es7_tuning`) | `earth_shock.go`, `flame_shock.go`, `frost_shock.go`: Era base points, die and growth per level for every rank, Flame Shock's DoT too. Coefficients, costs and cooldowns stay Forever's. `EarthShockRank7DamageScale` 1.08 → 1.0. Tests `TestOrcShamanRankTuning`, `TestOrcShamanSpellPowerBeta`, `TestOrcShamanFlameShockTicks` | +28 / +28 / +58 / +58 / +45 / +53 |
| Lightning Overload 10 / 20 / 30% (was 3 / 7 / 10%) | notes (408438 still 10% in 70753) | wrong | `lightning_overload.go`. Test `TestOrcShamanLightningOverload` | +48 / +48 / 0 / 0 / 0 / 0 |
| Lava Burst removed, Call of Thunder (+5% crit Lightning Bolt and Chain Lightning) in its place | notes; 16120 in the client already says 5% | missing | **Phase B** (needs A1's tree: the `lava_burst` talent field and the presets' strings) | – |
| Elemental Mastery renamed Lightning Mastery | notes | landed (effect unchanged: Lightning Bolt 0.5 s faster, Chain Lightning no cooldown) | Phase B: the label and the Lava Burst half of it go with Lava Burst | 0 |
| Elemental Reach no longer reduces Totemic Projection's cooldown | notes | landed (the sim never modelled the cut) | none | 0 |
| Improved Ghost Wolf moved from row 2 to row 4 | notes | tree | **Phase B** (A1's tree; the Imp GW preset's string) | – |
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

Gating: the shaman's spell tables and these talents were already Forever-only values (the fork keeps no Classic copy of
them); Stormstrike's cooldown and 150% sit inside its `IsForever()` branch.

## Shared files touched

- `tools/rotopt/spell_names.go`: 437009's display name.
- `tools/dps_report/patches/0004-...patch`: its `sham_mw_ppm` hunk's removed and default lines now read
  `4*float64(...)`, so the patch still applies to `sim/shaman/talents.go` (`git apply --check --include='sim/shaman/*'`
  passes). The `sham_lb_31plus` and `sham_es7_tuning` hunks apply unchanged; their constants are now 1.0.
- `sim/testdata/report_hashes.txt`: the six shaman lines only. `TestRaidForeverReportHashesStable` already fails on
  master b648ac704 for 4 druid, 10 paladin and 1 shaman request (not from this branch); those lines are left alone.
- `sim/shaman/warden/TestWardenShaman.results` (the tank session's golden): moved by the shared Stormstrike, Flurry,
  shock and Mental Dexterity changes, regenerated.

## Questions for the orchestrator

1. **Three report assumptions are now settled by the notes**: `sham_lb_31plus` (×1.12, the notes retuned Lightning
   Bolt's top ranks), `sham_es7_tuning` (×1.08, the shocks are Era's now) and `sham_mw_ppm` (10 PPM Expected; the
   notes double it, so the sim's value is 20). Recommendation: drop the first two from `assumptions_def.py` and the
   0004 patch, and move `sham_mw_ppm`'s settings to 10 / 20 / 30 (Low / Expected / High) or make it one value at 20.
2. **Lightning Bolt ranks 1–3** keep the client's values (already above 0.78 of Era) though the notes say "all ranks
   increased". Leveling only. Recommendation: keep until a client or hotfix has numbers; a notes.md "Need to Verify"
   row goes in with Phase B.
3. **Chain Lightning's "comparable rank"** is our reading (the newest Lightning Bolt at Chain Lightning's learn level).
   Only rank 4 matters at 60, and there it's unambiguous (both learned at 56).
