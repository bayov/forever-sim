# Change ledger since round 70753 (2026-11-11)

Every Forever class, spell and item change announced or shipped after the round 70753 sources
(2026-11-02/03), checked against master at 5a216e724. Sources are in
`orchestration/audit/ref/client_update/sources/`:

| Short name | File | What it holds |
|---|---|---|
| BN | `blizzard_hotfixes_2026-11-10.txt` | Blizzard's hotfix article (Nov 5, 6, 9, 10) |
| BF | `blizzard_forum_blue_2026-11-11.txt` | Kaivax's blue posts, including "Class Tuning Incoming - November 10" |
| FC | `foreverchanges_2026-11-11_patch-notes.txt`, `foreverchanges_2026-11-11_changes_since_70753.txt` | foreverchanges.pro |
| W | `wago_hotfixes_70888_2026-11-11.txt` | wago.tools server hotfix rows, read 2026-11-11 10:30-11:01Z |
| DN10 | `devnotes_post10_2026-11-04.txt` | Dev notes post 10. UI only, no rows here. |

The newest client build is still 1.60.1.70888, and no client build carries any change below.

The **Shipped** column uses these values:
- **announced**: in Blizzard's notes only; no server row yet.
- **hotfix N**: a server hotfix row at push N.
- **live (no row)**: Blizzard lists it as live, but it is server-side behaviour the hotfix tables don't show.

The **Status** column uses these values:
- **done**: already on master, with the commit.
- **open**: the sim needs a change.
- **none**: no sim effect.

## Shaman

The report builds named below are in `tools/dps_report/req/`:
- **Enh4** is the four Enhancement builds: `shaman__enhancement-level60`, `-imp-gw`, `-mana-tide` and `-water-shield`. All four take Mental Dexterity 3/3, Maelstrom Weapon 5/5, Stormstrike, Windfury Weapon and their own Windfury Totem.
- **Ele2** is the two Elemental builds: `shaman__elemental-level60` and `-mana-tide`.

| # | Change (exact numbers) | Source, date | Shipped | Sim now | Status | Moves |
|---|---|---|---|---|---|---|
| S1 | Stormstrike 100% weapon damage (was 150%) | BN Nov 10, BF 2404092 | announced. W: SpellEffect 1368308 still 150 (push 113380), checked 11:01Z | `sim/shaman/stormstrike.go` `StormstrikeWeaponDamagePercent = 1.5` | open | Enh4; Enhancement stat and leveling weights (`ui/enhancement_shaman/leveling_weights*.json`) |
| S2 | Stormstrike cooldown 8 sec (was 6) | BN Nov 10 | announced. W: SpellCooldowns 55241 still 6000 ms (113380) | `stormstrike.go` `registerStormstrikeSpell`: under Forever `cooldown = time.Second * 6` (line 72). The header comment at line 9 still says 8 sec. | open | Enh4; APLs that wait on the cooldown |
| S3 | Mental Dexterity converts 33/66/100% of Intellect to AP (was 50/100/150%) | BN Nov 10 | announced. W: no row for 415140 since push 93879 | `sim/shaman/talents.go:38` `ratio := []float64{0, .5, 1, 1.5}` under Forever | open | Enh4 (all 3/3); Int's value in the Enhancement weights; gear and enchant picks |
| S4 | Maelstrom Weapon 20% less cast time and mana a stack at every rank (was 4/8/12/16/20%) | BN Nov 10 | announced. W: no row | `talents.go:572-573` `.04 * points` and `4 * points`. At 5/5 that is already 20%, but ranks 1-4 are too low. | open (ranks 1-4 only) | No report build (all 5/5); leveling talents at ranks 1-4 |
| S5 | Maelstrom Weapon procs 6/7/8/9/10 PPM by rank (was 20 PPM at any rank). Tooltip example: 60% a swing at 3.6 speed, 33% at 2.0, at 5/5. | BN Nov 10 | announced. W: SpellAuraOptions 223898 still ProcChance 100, no PPM row (113380) | `talents.go:545` `applyMaelstromWeapon`: under Forever (`everyHit`, line 562) every landed melee hit gives a stack (fitted to the Nov 8-9 logs, bcc5d6cd4). The non-Forever path is 2 PPM a point. | open. Needs `ppm = 5 + points` under Forever. | Enh4: about 40% fewer stacks with a 3.6 weapon, so fewer instant Lightning Bolts and Chain Lightnings |
| S6 | Weapon imbues and weapon totems work at the same time (Windfury Weapon and Windfury Totem can both proc) | BN Nov 10 | live (no row) | `sim/core/buffs.go:2165` `blocked := mainHandHasImbue(character, windfuryWeaponEnchantIds)` turns off the WF Totem buff. `sim/core/totem_weapon_buffs.go:111` blocks FT Totem under FT Weapon. Both came in with 9ba00c2dc. | open | Enh4 (all cast WF Totem 10614 over WF Weapon; the APL note "Windfury Weapon turns its buff off for the shaman" must go). Group members with an imbue in the raid builds. |
| S7 | Rage of the Storm: +16 Strength (was Intellect: Blizzard says 16 Int, FC says 12 Int), 102-154 at 3.60 (was 91-137 at 3.20), DPS 35.6 unchanged | BN and FC Nov 10 | hotfix 113886 (ItemSparse: stats Str/Sta, speed 3600) | Item db: Sta 15, Int 12, 91-137 at 3.2, ilvl 40. The 10% Stormstrike effect is in `sim/shaman/items.go:203`. Used in `tools/rotopt/leveling/level30_enh_solo.json` and in enhancement tests. | open (data pull) | Level 30 Enhancement leveling only, not the report |
| S8 | Lightning Overload deals the right damage for Lightning Bolt | BN Nov 9 | hotfix 113805 (SpellEffect 1072626-1072630; 408477 is base 176.5 + 1.05 a level, half of LB rank 10's 353 + 2.1) | `sim/shaman/lightning_overload.go:25` `config.DamageMultiplier *= .5` on the LB ranks | none (already matches at rank 10; lower ranks not checked) | Ele2 unchanged |
| S9 | Maelstrom Weapon can be tracked in the Cooldown Manager | BN Nov 10 | hotfix 113871 (CooldownSetSpell) | n/a | none | none |
| - | Already on master since Nov 3, not open: bcc5d6cd4 (MW on every melee hit, Flametongue rank 5 cut), 0ce9dbf40 (WF Totem procs from white MH swings only), 9ba00c2dc (WF and FT Totem replace each other), 5a870eccc (Stormstrike mark) | | | | done | |

S1, S2, S3, S5 and S6 together are a large Enhancement nerf, offset in part by S6. Redo the APLs and the talent search after landing them.

## Druid

| # | Change | Source, date | Shipped | Sim now | Status |
|---|---|---|---|---|---|
| D1 | Nature's Swiftness also makes Wild Growth instant | BN Nov 10 | hotfix 113885 (SpellAuraOptions 129480, SpellEffect 695239 class mask) | `sim/druid/healing.go` `registerNaturesSwiftnessCD`: Healing Touch, Regrowth and Wrath only | open (healing sims only) |
| D2 | Natural Instinct applies to Lacerate and Pounce | BN Nov 9 | hotfix 113854 (Pounce bleed 9007/9824/9826) | `sim/druid/talents.go` `applyNaturalInstinct` includes Lacerate. Pounce is not modeled. | done for Lacerate, none for Pounce |
| D3 | Feral Charge (Cat) row | W | hotfix 113853 (SpellMisc 705261) | not modeled | none |
| D4 | Claw/Rake cost spell 27851 | W | hotfix 113860 (SpellEffect 704563) | Claw and Rake costs | check: the number wasn't read |
| D5 | Omen of Clarity after a spec swap, Bloom heal, Pounce debuff display, Shadowmeld Ravage/Pounce, tooltips | BN Nov 5-9 | live | n/a | none |

## Hunter

| # | Change | Source, date | Shipped | Sim now | Status |
|---|---|---|---|---|---|
| H1 | Arcane Shot costs 11% of base mana at all ranks | BN Nov 10 | hotfix 113892 (SpellPower 3044, 14281-14287) | `sim/hunter/arcane_shot.go:36`: flat cost table `{0,25,35,50,80,105,135,160,190}` | open (7 BM/MM builds) |
| H2 | Summon Hawk costs 11% of base mana at all ranks | BN Nov 10 | hotfix 113892 (SpellPower 313567 and the other ranks) | `sim/hunter/summon_hawk.go:131`: cost `{0,80,105,135,190}` | open (same 7 builds) |
| H3 | Hawk melee damage about 80% higher | BN Nov 10, FC | live (no row) | `summon_hawk.go` `newHawks`: log-fitted swing, `AttackPowerPerDPS: 112*hawkLogUnleashedFury` | open; refit from new logs |
| H4 | Pet health 150% to 115%; Great Stamina gives 30 to 400 pet health | BN/FC Nov 9-10 | live | n/a | none (DPS) |
| H5 | Frost Trap Aura 13810 effect row | W | hotfix 113864 (2 sec periodic) | n/a | none |

## Mage

| # | Change | Source, date | Shipped | Sim now | Status |
|---|---|---|---|---|---|
| M1 | Arcane Concentration's chance "now correctly splits across the cast and every bolt (was granting 10% on cast and per bolt)" | BN Nov 9 | live (no row) | `sim/mage/talents.go:238` `applyArcaneConcentration`: 2% a point, never procs from Arcane Missiles, 1 sec ICD | open; the wording is ambiguous, so confirm in logs (mage__arcane/fire/frost meta) |
| M2 | Flamestrike ground ticks no longer consume Clearcasting | BN Nov 9 | live | Flamestrike ticks and Clearcasting | check; likely no effect |
| M3 | Arcane Blast lasts 18 sec | earlier | - | `sim/mage/arcane_blast.go:31` `ArcaneBlastDuration = 18s` | done |

## Paladin, Priest, Rogue

| # | Change | Source, date | Status |
|---|---|---|---|
| P1 | Consecration bug with despawned targets; Seals and Redoubt in the Cooldown Manager | BN Nov 9-10 | none |
| PR1 | Revelation vs Vampiric Embrace healing; PI persisting on allies; Fear Ward in Shadowform; Holy Nova unlearn | BN Nov 5-9 | none. `sim/common/enchant_effects.go` Revelation already procs from direct damage casts only. |
| R1 | Quick and the Dead legacy talent "sometimes did not work as intended" | BN Nov 6 | none. Not in the sim, no numbers. |
| R2 | Shank tooltip | BN | none |

## Warlock

| # | Change | Source, date | Shipped | Sim now | Status |
|---|---|---|---|---|---|
| WL1 | Wrack 30% less damage | BN Nov 10 | hotfix 113766. SpellEffect 1349556: base 36 to 25.2, coefficient 0.286 to 0.2002, 2.5 to 1.75 a level, so 86 to 60.2 at 60. | `sim/warlock/wrack.go:19-21`: 36 + 2.5 a level, coefficient 0.286 | open (3 Affliction builds take Wrack) |
| WL2 | Wrack no longer raises the periodic damage; Corruption, Siphon Life and Bane of Agony tick 100% faster while channelled (same total duration) | BN Nov 10 | hotfix 113766 (SpellEffect 1349557: aura 4, value 100) | `sim/warlock/talents.go:299` `applyWrack`: ×1.1 on Corruption and Bane of Agony while channelled | open |
| WL3 | Soul Siphon: Drain Life, Drain Soul and Wrack get 4/8/12% for each of Corruption, Siphon Life and Bane of Agony on the target, up to 36% (was any Affliction spell). Wrack no longer raises its own damage. | BN Nov 6 and 10 | live (no row) | `talents.go:317` `soulSiphonMultiplier`: counts Corruption, any bane, Siphon Life and the curse, cap 3, 4% a point | open (minor: drop the curse and the other banes) |
| WL4 | Imp Firebolt ranks 1-4 base damage up (rank 1 6-8, was 4-5; rank 4 18-22, was 18-19); Imp health +15% | BN Nov 10 | hotfix 113766 (SpellEffect 681913 base 6, var 0.333, 0.3 a level) | `sim/warlock/imp.go` `impFireboltRanks` | open (low ranks only; level 60 unaffected) |
| WL5 | Voidwalker, Succubus and Incubus health +15%; Voidwalker +40 health | BN Nov 6 and 10 | live | n/a | none |
| WL6 | Bane of Havoc now on the GCD | BN Nov 10 | live | not modeled | none |
| WL7 | Incinerate limited to Destruction; Orc Command | earlier | - | Orc Command per TASKS | none / done |

## Warrior

| # | Change | Source, date | Shipped | Sim now | Status |
|---|---|---|---|---|---|
| WR1 | Sweeping Strikes no longer loses charges if no second target is hit | BN Nov 10 | live | `sim/warrior/sweeping_strikes.go:90` removes a stack on every melee hit. The copy is only cast when `AoETargetCount > 1` (line 84). | none for single target. Open, minor, for AoE with fewer targets than hits. |
| WR2 | Thunderclap in the Cooldown Manager; Reckless Swings tooltip | BN | live | n/a | none |

## Items

| # | Change | Source, date | Shipped | Sim now | Status |
|---|---|---|---|---|---|
| I1 | Rage of the Storm, see S7 | BN, FC Nov 10 | hotfix 113886 | db Int 12, 3.2 speed | open (data pull) |
| I2 | Earthen Silk Slippers +17 Healing instead of +18 Spell Power | FC Nov 10 | hotfix 113944 (ItemSparse 254013: stat types 41/42/6) | db still 18 Spell Power | open (data pull; healer gear only) |
| I3 | Sparklance Wand 280031 deals Arcane (was Nature) | BN Nov 10 | live (item row not checked) | wand school | open, minor (not in any report build) |
| I4 | Searing Dagger 273003 only procs from damaging spells | BN Nov 6 | live | proc mask | check (not in any report build) |
| I5 | Gloves of Holy Might "work as intended" | BN Nov 9 | hotfix 113833 (Spell 1325229, 1330855, SpellEquippedItems, SpellPower 315227) | item 867 effect | check what the new spells do |
| I6 | Enchant Lesser Spirit (boots, bracer) gives Spirit, not resist all | BN Nov 10 | hotfix 113844 (SpellItemEnchantment 8203) | enchant data | check `forever_enchants.json`; leveling only |
| I7 | Berdun's Trusty Rifle has no random stats and is not BoE; Swiftstrike Shot and Swiftfeather Arrow no longer BoP; Goliath's Wristguards changed | BN, FC | live | item db | open (data pull) |
| I8 | Engineering trinkets, Flamebreak Talisman | W | hotfix 113769 | already open in TASKS.md | open (tracked) |
| I9 | SpellAuraOptions 248634 (spell 1324323: 100%, 15 sec ICD) | W | hotfix 113846 | unknown spell | check |

## General

| # | Change | Source, date | Status |
|---|---|---|---|
| G1 | Dungeon enemy XP +50% | BF 2406595, Nov 8 | none (the leveling sim has no dungeons) |
| G2 | Quests 48+ give gold; AH and mail changes; vendor prices; Shatterspear Trolls give no XP | BN Nov 9-10 | none |
| G3 | Four new unreleased spells 1331270-1331273 (status 3) | W hotfix 113928 | none yet |
| G4 | No new client build after 70888 | wago `latest --probe 300` | - |

## Sources not reached

- The Blizzard forum JSON API returned 422 (rate limited). The crawler HTML of the same threads worked and is saved.
- No Blizzard hotfix section for Nov 11 existed when this was read.
- Wowhead was not checked.
- foreverchanges.pro item pages were not tried. TASKS notes they return 403 from the cloud.
- Soul Siphon, the imbue/totem stacking and the Hawk damage have no hotfix rows. They are likely server scripts or creature data that wago doesn't show.
