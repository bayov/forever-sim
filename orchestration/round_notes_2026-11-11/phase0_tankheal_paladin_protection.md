# Phase 0 tank: protection paladin (2026-11-25)

Balanced, Max threat and Max survival, on the final sim code (origin/phase0-tankheal with the five sim batches,
Windfury 1a1c5f0ba and the Premier drop a1904ac61). Scored by the tank score, six bosses, as bis_phase0.sh does.
Lists in `ui/protection_paladin/bis/level60_phase0_paladin_protection_<preset>.bis.json`, gear sets in
`gear_sets/`, both in the preset's `tools/tankbench/req/paladin_protection_<preset>.json`. Race Undead (Horde).

## How the lists were made

- All three: full search, then RANK_ONLY over all slots on the final sim batches.
- Max threat: feet re-ranked after the Premier drop, then RANK_ONLY over all slots again on the new talents.
- No Premier item was a pick in Balanced or Max survival.
- Balanced, Max survival and the Alliance row: off hand re-ranked after Force Reactive Disk went out
  (Molten Core schematic). Evergreen Shield takes it in all three. In Max survival The Immovable Object is
  8.3 ahead in the list ranking, inside the noise of about 11, and the confirm kept Evergreen Shield.

## Picks

Flags: D doubtful. Everything else is open.

| Slot | Balanced | Max threat | Max survival |
|---|---|---|---|
| Head | Stalwart Helm (BS, UBRS plans) | Stalwart Helm | Enriched Thorium Helm (BS, Gor'mak plans) |
| Neck | Orb of the Darkmoon (Faire) | Nacreous Shell Necklace (see below) | Amulet of the Darkmoon (Faire) |
| Shoulder | Soulforge Spaulders | Soulforge Spaulders | Darkspear Mantle (Darkspear Revered, Horde) |
| Back | Stalwart Cloak (Pix, 75 Reals) | Regal Direhorn Drape (Forever rare) | Shifting Cloak (LW, Dire Maul) |
| Chest | Breastplate of the Chromatic Flight (UBRS quest) | same | same |
| Wrist | Soulforge Bracers | Soulforge Bracers | Soulforge Wristguards |
| Hands | Soulforge Handguards | Storm Gauntlets (LW 270) | Heavy Thorium Gauntlets (D, BS plans with no source) |
| Waist | Soulforge Belt | Soulforge Belt | Redeemer's Waistcord (Pix, 100 Reals) |
| Legs | Soulforge Legguards | Soulforge Legguards | Soulforge Legguards |
| Feet | Soulforge Sabatons | Depleted Thorium Sabatons (D, BS plans with no source) | Depleted Thorium Sabatons (D) |
| Rings | Tracker's and Watcher's Signets of Vigilance (D, no rep source) | Watcher's and Tracker's Signets of Vigilance (D) | same as Max threat |
| Trinkets | Chained Essence of Eranikus, Adaptive Combat Assistant | same | same |
| Main hand | Hand of Edward the Odd (world drop) | same | same |
| Off hand | Evergreen Shield (D, BS plans with no source, may need raid essence) | Evergreen Shield | Evergreen Shield |
| Libram | Steadfast Libram (D, formula has no source) | same | same |

- Nacreous Shell Necklace 22403 shows "infeasible" in the list from its DPS price (the price restored on
  2026-11-24), but the tank search allowed it. Next best is Orb of the Darkmoon, 11.5 behind.
- Soulforge (1673) is the protection, retribution and holy D1 set. Mixing variants counts for the set.

## Price re-rank (prices 23469c224)

phase0_check.py flagged open items above the picks once the prices landed. RANK_ONLY over the flagged slots
(Balanced neck, wrist, finger; Max survival neck, hands, waist, finger; Max threat neck), swaps must clear
2 stderr at 5000 iterations:
- Balanced: Tracker's Signet of Vigilance for Watcher's Signet of Brilliance, +17.4 (6.3), confirmed. The
  wrist enchant moved too. List 2965.3.
- Flags that stand, all below 2 stderr in the confirm: Balanced neck (Amulet of the Darkmoon -1.5 (5.0),
  Horn of Ve'han -4.5 (6.5)) and wrist (Soulforge Wristguards +5.7 (3.6)). Max survival neck (Horn of Ve'han
  +6.4 (13.1)), hands (Force Imbued Gauntlets +19.8 (16.9)), waist (Brigam Girdle -7.4 (13.6), the rest
  lower). Max threat neck (Orb of the Darkmoon -6.1 (2.9)). The list ranking runs fewer iterations, and
  Max survival's spread is wide, so those rows rank above the pick there and not in the confirm.
- Alliance row re-ranked on the shoulder from the new Max survival list: unchanged.
- Max survival confirms after the master merge, bench at 5000 iterations against the list's 775.5 (stderr
  about 15): Warder's Waistguard +6.3, Soulforge Belt -26.3, Talisman of Evasion +0.7. None clears 2 stderr,
  list kept. Balanced wrist and neck were confirmed above.

## Consumables on the Phase 0 sets (tank_fairness-20261126-1, 2026-11-26)

- Max survival takes the Flask of the Titans for Supreme Power, per the audit (+108.5 / +131.2, 3000
  iterations, seeds 1 / 2). Request and `MaxSurvivalConsumes` changed.
- Screen, one swap at a time against each preset's set, 1000 iterations seed 1: flask, food, Agility, armor,
  health and mana elixirs, oil and stone, Greater Arcane Elixir, potion, Zanza, Juju Power and Might.
  Brilliant Wizard Oil leads at every preset (+26 to +45) but stays out (Zandalar vendor, raid-only, as in
  docs/tank_paladin.md), as does the Mageblood Potion. The potion swaps score exactly the base: the default
  potion is never drunk.
- Confirmed at 3000 on seeds 1 and 2: Max threat Spirit of Zanza for the Ground Scorpok Assay +21.5 (4.2) /
  +12.3 (4.3), taken (request and `MaxThreatConsumes`). Balanced Dirge's Chops +5.5 / +3.8, not taken.
- No gear re-search: Spirit of Zanza swaps 25 Agility for 50 Spirit and 50 Stamina.

## Tags added

- phase0_content.json: the searches' unlisted picks and top 5s, mostly dungeon drops and crafts (open).
  Doubtful: Steadfast Libram, Evergreen Shield, Heavy Thorium Gauntlets, Depleted
  Thorium Sabatons, Warder's Sabatons, Stronghold Gauntlets, Frostwolf Advisor's Cloak (Horde AV, like the
  other AV items). Out: Quel'Serrar, Lord Blackwood's Buckler, Core Armor Kit. Legplates of Vigilance is
  included with the Dungeon Set 2 note, as the file does for that chain. Energized Chestplate added open
  (Alzzin, Dire Maul), and Spaulders of Valor and Barrier Shield from the Alliance rows. Force Reactive
  Disk is out on the base branch (Molten Core schematic).
- phase0_enchants.json: the enchants the searches ranked, open (trainer, vendor, Merchant's Favor, Dire Maul
  librams). Core Armor Kit is out (Core Leather).

## Set trap

bench at 1000 iterations, sets swapped into the list, then any close call at 10000:
- Balanced (list 2795.7 on the old code): Imperial Plate 2pc 2533.7, Imperial 8pc 2030.7, Blessed 8pc 2120.3.
  List kept (Soulforge 6pc).
- Max threat (list 3686.7): Imperial 2pc 3458.8, Blessed 8pc 3091.3, Valor 8pc 3255.0. List kept. After the
  re-rank Storm Gauntlets (3858.7) beat Soulforge hands for 6pc (3791.5 and 3783.5), and the feet then moved
  to Depleted Thorium Sabatons, so it is Soulforge 4pc.
- Max survival (list 822.7, with Force Reactive Disk, before the off hand re-rank): Soulforge 4pc 785.8, Soulforge 6pc 836.3, Imperial 8pc -608.8, Blessed 8pc
  -653.5. At 10000 iterations Soulforge 6pc is 775.0 against the list's 788.5. List kept (Soulforge 2pc).

## Hand check

The picks were read against their sources and tags. Two subagent tags were fixed: Stronghold Gauntlets from
infeasible to doubtful (infeasible is for Exalted rep only), Legplates of Vigilance from out to the Dungeon
Set 2 include. Breastplate of the Chromatic Flight is a UBRS quest chain with a heavy farm (open).

## Talents and rotation

Knob search on the Phase 0 gear: no change at any preset. Talent search, then paired checks at 10000
iterations, two seeds:
- Balanced: no change.
- Max threat: Anticipation 2 for Shield Specialization 2 (`05005-25300100413201531-5023`), 3906.7 vs 3886.8
  and 3905.3 vs 3886.6, noise about 3. Adopted. Time to OOM drops to about 204s from about 1016s.
- Max survival: `03-15301102513501531-50231` was 789.6 vs 788.5 and 782.1 vs 766.1, noise about 11. Not
  separated, talents kept.

## Faction row

Only Max survival has a faction pick (Darkspear Mantle, Horde).
Alliance row (RANK_ONLY on the shoulder and off hand): Theramore Mantle (Theramore rep, Alliance) and
Evergreen Shield. Dwarf 591.5 against Human 548.1 on the list's measure (534.4 against 523.5 on the final
measure), so the row is Dwarf. Horde (Undead) is 800.2 on the same measure. Files:
`..._max_survival_alliance.bis.json` and `.gear.json`.

No faction enchant was picked. Balanced and Max threat have no row.

## Score, Phase 0 against synthetic

`tankbench -iters 1000`, mean of the preset's column over magmadar, golemagg, onyxia, broodlord,
vaelastrasz and chromaggus. Synthetic is the request on origin/master.

| Preset | Phase 0 | Synthetic | Gap |
|---|---|---|---|
| Max threat | 3936.3 | 3280.3 | +656.0 |
| Balanced | 2965.3 | 2857.7 | +107.6 |
| Max survival | 906.5 | 1858.0 | -951.5 |

Max survival falls short on Broodlord (-628 against 336) and Chromaggus (-2145 against -583). The cause is
traced under Max survival vs synthetic below (the gear).


## Max survival vs synthetic

`tankbench -iters 1000`, mean of the six scoring bosses. Synthetic is the origin/master request. Stats are tankbench's sheet against a level 63 boss. Block value is (sheet value + 0.05 Str) x the block value multiplier. Coverage is miss + dodge + parry + block (102.4% on the sheet is uncrushable). Both sets clear the 315 Defense that stops Forever's 0.6% boss crit, so crit plays no part. Max survival is 2 x SafeDPS less 3 x 900 / 300 x the healing that holds deaths to 5%, so healing needed is the survival side.

| Stat | Phase 0 | Synthetic |
|---|---|---|
| Health | 9471 | 8691 |
| Stamina | 627 | 669 |
| Armor | 9391 | 9816 |
| Defense | 382 | 463 |
| Miss / dodge / parry / block | 9.7 / 18.4 / 12.7 / 7.7 | 12.9 / 20.3 / 15.9 / 17.9 |
| Block value | 135 | 282 |
| Coverage | 48.5% | 67.0% |
| Crushing blows taken | 9.8% | 2.5% |
| DTPS | 380 | 284 |
| Healing needed | 293 | 197 |
| Max survival | 906.5 | 1858.0 |

The only non-gear field that differs is the flask (Titans on Phase 0, Supreme Power on synthetic, deliberate). The synthetic request on Titans scores 1825.3 (healing needed 160), so the gap is the gear. Holy Shield's 30% lifts synthetic to about 97% coverage, close to uncrushable, but Phase 0 only to about 78%. Lost Defense (-81), block chance (-10.2) and block value (-147) put crushing blows back on the table.
