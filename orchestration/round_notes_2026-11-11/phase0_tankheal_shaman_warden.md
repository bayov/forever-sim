# Phase 0 tank gear: Warden Shaman (2026-11-25)

Branch `phase0-tankheal-shaman_warden`, merged with `origin/phase0-tankheal` at d91a2c7 (Windfury Totem extra
attacks, throws, stun grenades, Premier PvP sets). Lists in `ui/warden_shaman/bis/`, gear sets in
`ui/warden_shaman/gear_sets/`, requests `tools/tankbench/req/shaman_warden_*.json` wear them.

## Scores

tankbench, 1000 iterations, six scored bosses, mean of the preset's own column. Synthetic is
`git show origin/master:<req>`.

| Preset | Phase 0 | Synthetic | Gap |
|---|---|---|---|
| balanced | 2026.2 | 1245.7 | +780.5 |
| max_threat | 3418.0 | 2608.2 | +809.8 |
| max_survival | -1337.0 | -2234.0 | +897.0 |

Phase 0 beats the synthetic set on all three, the opposite of what a placeholder set should do. The synthetic
sets are worth a look by the parent (the shaman sets may lack the defense or shield the real ones have).

## Picks

Same in all three unless noted.

| Slot | balanced | max_threat | max_survival |
|---|---|---|---|
| Head | Face of The Five Thunders (doubtful, end of the Valthalak chain) | same | Helm of Narv (world drop) |
| Neck | Amulet of the Darkmoon (1200 tickets) | same | same |
| Shoulder | Spaulders of The Five Thunders (Mokvar quest) | same | Darkspear Epaulets (Darkspear Revered, Horde) |
| Back | Shifting Cloak (LW 300, Dire Maul cache pattern) | Stalwart Cloak (75 Reals) | Shifting Cloak |
| Chest | Onyxia Scale Breastplate (doubtful, LW, Classic recipe needs a raid scale) | same | same |
| Wrist | Bands of The Five Thunders | same | Pyremail Wristguards of Fire Protection (Incendius, BRD) |
| Hands | Storm Gauntlets (LW 270) | same | Mourning Handguards (open by the 11-12 check, no source noted) |
| Waist | Girdle of The Five Thunders | same | Blackmaw Chain Belt (Blackmaw Hold) |
| Legs | Legguards of the Chromatic Defier | Legplates of The Five Thunders | Legguards of the Chromatic Defier |
| Feet | Windreaver Greaves | Treads of The Five Thunders | Defiler's Chain Greaves (doubtful, AB unconfirmed, Horde) |
| Rings | Watcher's Signet + Tracker's Signet of Vigilance (both doubtful, no Watchers rep source) | same | same |
| Trinkets | Chained Essence of Eranikus (ST quest), Adaptive Combat Assistant (125 Reals) | same | Defender's Grip Stabilizer (125 Reals), Adaptive Combat Assistant |
| Main hand | Iceblade Hacker (Ras Frostwhisper) | same | same |
| Off hand | Evergreen Shield (doubtful, plans have no source) | same | same |
| Ranged | Totem of Rage (Magmus, BRD) | same | same |

Enchants include Arcanum of Protection, Superior Deflection and Law of Nature on the shield (doubtful, the
formula left Wowhead Forever on 11-04). Max threat's neck was moved by hand from the doubtful Amulet of the
Dawn to Amulet of the Darkmoon (+17 and +10 at 10000 iterations, two seeds).

## Flags

- Doubtful picks: Face of The Five Thunders, Onyxia Scale Breastplate, both Watchers rings, Evergreen Shield,
  Law of Nature, Defiler's Chain Greaves, Highlander's Mail Greaves (Alliance row).
- "infeasible" on Windreaver Greaves, Totem of Rage and Bloodmail Boots comes from DPS gold per DPS. Tanks use
  their own price, the gear sets wear them.
- Theramore reqrep 7 is tagged unevenly: 271950 open, 271945 and 271949 infeasible.
- Wowhead Forever renamed Major Health (20026) to Major Stamina +10, the sim still models +100 Health.
- Several formula drops are tagged open without an AH price check.
- Balanced's APL names the synthetic trinket 999216, which does nothing now. Switching it to 10455 lost 75, so
  it stays.
- Merge conflict on Force Reactive Disk (18168): I kept "out" (schematic drops in Molten Core, AtlasLoot), the
  other side had "doubtful", which would let it be picked. Predator's Shroud took the other side's "doubtful".
- Tagged this round: Rattlecage Buckler 14528 (Rattlegore, Scholo, max threat top 5), Beaststalker's Mantle
  16679 (Wyrmthalak, LBRS, Alliance row top 5).
- Five unmodelled or hidden stat pool items could matter to a tank. None is a pick.

## Set trap

score.py, 3000 iterations, confirmed at 10000 on seeds 1 and 2. Five Thunders, all variants together.

- balanced: 6 or 8 pieces lose (2029 against 1810 to 1946). Kept 4 pieces (Face, Spaulders, Bands, Girdle).
- max_threat: 6 pieces with Bands and Treads wins (+23.6, +23.5). Kept.
- max_survival: 2, 3, 4, 6 and 8 pieces all lose. No Five Thunders.
- Face of The Five Thunders beats Helm of Narv by 6 (balanced) and 17 (max threat). Narv is the fallback if
  the Face is ruled out.

## Talents and rotation

Paired abtest, 10000 iterations, two seeds. In this sim the Warden is crit immune at 315 Defense, which every
set meets.

- max_threat: talents `0535-051031303351012201-050002` (was `05350002-050032313251012201-05`), +43.4 and +50.8.
  Fire totem only in the last 80s (its mana goes to Earth Shock before), +5.4 and +6.5. List ranked again: 3418.1,
  no pick changed.
- max_survival: the APL's trinket id 999216 became 272437 (the worn one), +58.6 and +70.1. Trinkets ranked
  again, no change. Its talent search lost.
- balanced: nothing won.

## Faction rows

- balanced and max_threat (Night Elf, Alliance): no faction picks, skipped.
- max_survival (Undead, Horde): Alliance row as RaceDwarf on shoulder and feet. Theramore Epaulets and
  Highlander's Mail Greaves (doubtful). Files `..._max_survival_alliance.bis.json` and `.gear.json`. List score
  -1668.2. The drop is the race: the Horde set on a Dwarf scores -1683 at 3000 iterations, and the Alliance set
  is about 11 better. Flame Walkers ranked above the greaves but was a wash on the bench (-10, +8.6).

## Public list

Warcraft Tavern "Shaman Tanking Progressive Pre-Raid BiS" (Classic Era, undated): Annihilator, Draconian
Deflector, Helm of Narv, Mark of Tyranny and others. It agrees on Narv for max survival. Its other picks are Classic items, and
Forever items outrank them in our lists.

## Re-rank on Phase 0 prices (fc39e50)

phase0_check.py flagged open items above four picks. RANK_ONLY on those slots kept every pick, and a swap on the
bench (3000 iterations, seeds 1 and 2, preset column) clears no flag, so all four stand:

- balanced head, Helm of Narv for Face of The Five Thunders: -0.3, -5.4. The rank misses the 4-piece set proc.
- max_survival neck, Verek's Collar: -3.9, -37.4. Evil Eye Pendant: +12.0, -38.4.
- max_survival feet, Flame Walkers for Defiler's Chain Greaves: +8.5, -8.4.

The two seeds differ by about 55 on max survival, so these are within noise and the cheaper fallbacks are near
equal. Scores vs synthetic are unchanged (list scores 1962.8, 3418.1, -1336.9).

## Legal races (tank_fairness-20261105-2, 2026-11-26)

Night Elf and Undead can't be shamans (run_build.py CLASS_RACES). Each preset was measured on its Phase 0 set at all
five legal races, 3000 iterations, seeds 1 and 2, own column:

| Preset | Troll | Tauren | Orc | Windshaper | Dwarf |
|---|---|---|---|---|---|
| balanced | **1996.2** | 1945.0 | 1940.5 | 1936.7 | 1904.8 |
| max_threat | **3396.2** | 3316.9 | 3350.2 | 3322.7 | 3301.5 |
| max_survival | -1589.4 | **-1535.8** | -1666.7 | -1692.0 | -1683.8 |

Troll at balanced and max threat, Tauren at max survival, all Horde. Set in the requests, the lists, the UI presets and
sim.ts (Alliance default Dwarf, Horde Troll), and docs/tank_shaman.md. An explicit Berserking cast changes nothing
(identical to the decimal), so no APL change. Guard: tools/audit/gear/test_bench_races.py checks every tank and healer
bench request's race against CLASS_RACES.

Every list ranked again on the new race (FROM=current, all slots):

- balanced 2022.6. No pick changed.
- max_threat 3420.2. Legs Kilt of The Five Thunders for the Legplates (still 6 pieces), gloves enchant Natural Power.
- max_survival -1213.1. Enchants Arcanum of Protection on legs, Rugged Armor Kit on feet, Enchant Weapon - Recovery
  (45 Merchant's Favor, heals 5% of max health once per 10 sec, tagged this round).

Tagged: Enchant Weapon - Recovery, Tristam Legguards (The Beast, UBRS), Raptorscale Bulwark (doubtful, reported
only), Golem Fitted Pauldrons (Phalanx, BRD).

Trinkets: balanced's request lost its dead 999216 line (the bench is identical without it, and rotopt no longer swaps
it for Chained Essence, which loses 75). Bench and list now agree: 2022.7, 3420.2, -1212.8 against 2022.6, 3420.2,
-1213.1. The UI APLs still name 999216, which is correct while the UI presets wear the synthetic gear.

Scores, tankbench 1000 iterations, Phase 0 vs synthetic (master): balanced 2022.7 vs 1245.7 (+777.0), max_threat
3420.2 vs 2608.2 (+812.0), max_survival -1212.8 vs -2234.0 (+1021.2).

Faction rows: balanced and max threat have no faction picks. Max survival's Alliance row is Dwarf: Theramore Epaulets
and Flame Walkers, -1446.8.
