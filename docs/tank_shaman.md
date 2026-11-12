# The Warden shaman (shaman tank)

Forever's shaman tank (`WardenShaman`, `sim/shaman/warden`, `ui/warden_shaman`): a one-hander with Rockbiter
Weapon and a shield, Spirit Weapons for the parry and the threat, Stormstrike, Earth Shock and Lightning
Shield for threat, Stoneskin Totem and Guardian Totems for mitigation. The tank workers' brief
(`orchestration/brief_tanks.md`, T1) and `docs/tank_sim.md` say how tanks are scored.

**Where it stands (2026-11-03, realigned 2026-11-04 to client 70753 and the November 2 notes, below):** the Warden tanks in the sim (the boss attacks it, it blocks with its shield,
parries with Spirit Weapons, and Rockbiter's and Spirit Weapons' threat are on), every tank ability and talent
is audited against client 70753 (the October 30 hotfix's build) and tested, it wears a synthetic tank set by
standards 36, and it has a rotopt template (`tools/rotopt/shaman_tank.go`) and three presets
(`tools/tankbench/req/shaman_warden_*.json`, `ui/warden_shaman`), re-optimized in tank round 3 and extended in round 4 (Presets),
and re-tuned on 2026-11-12 after the Nov 10 shaman notes and the Natural flask cut (next section).

## Re-tuned after the Nov 10 shaman notes and the flask cut (2026-11-12)

Findings tank_rotations-20261111-1 and tank_fairness-20261111-1. The Nov 10 notes (Stormstrike 100% and 8 s, Mental
Dexterity 33 / 66 / 100%, Maelstrom Weapon at 5 + rank PPM, imbues and totems stacking) cut the Warden's SafeDPS about
10%, and client 70862 halved the Natural flasks. The same search as round 3 on the round 5 requests: `rotopt -template
shaman_tank -tank <preset>` on the six scoring bosses, talents (`-talent-search -require spiritWeapons`) then knobs,
1,000 a candidate and 3,000 to confirm, flasks and the weapon imbue screened at 3,000; every adopted change confirmed
paired at 10,000 on seeds 1 and 2 (`abtest`). The template gained a `potion` knob (the Greater Stoneshield line the
presets already had), so its APLs are the presets' byte for byte.

| Preset | Change | 10,000, seed 1 | 10,000, seed 2 |
|---|---|---|---|
| Max threat | Talents `05350002-050032313251012201-05` (Elemental Devastation +2, Elemental Warding +2, Shamanistic Focus; Toughness -3, Improved Reincarnation -2) | +15.1 ± 2.6 | +18.6 ± 2.6 |
| Max threat | Searing Totem and Fire Nova all fight (`fireLast=0`), on the new talents | +4.6 ± 0.8 | +3.9 ± 0.7 |
| Balanced | Flask of the Titans for Natural Precision | +13.5 ± 4.4 | +19.0 ± 4.3 |
| Max survival | Flask of the Titans for Natural Precision | +254.6 ± 12.1 | +230.3 ± 12.6 |
| Max survival | Healing Stream Totem (`waterTotem=2`), with the Titans | +87.1 ± 12.0 | +106.7 ± 13.8 |

Whole request, old against new (10,000, own weight, six-boss mean, seed 1 / seed 2): Max threat 2,495.2 → **2,514.7**
/ 2,496.5 → **2,519.0** (+19.7 ± 2.6 / +22.5 ± 2.6); Balanced 1,133.3 → **1,146.8** / 1,127.3 → **1,146.3**; Max
survival -2,738.0 → **-2,396.0** / -2,736.7 → **-2,399.7** (+341.7 ± 12.3 / +337.0 ± 13.3). The Warden stays fourth
at every weight.

Not adopted: Balanced's talent search found Elemental Devastation +1 for Reverberation (`05140003-...`, +20 at 3,000
under the Titans), which ties at 10,000 (-1.8 / -4.3 under Precision, -2.0 / 0.0 under the Titans); Max survival's two
talent searches (Precision and Titans) found nothing; Balanced's knob search moved nothing. The Flask of Supreme Power
ties or loses at Max threat (+1.4 ± 4.2 at 3,000 on the old build, -3.8 / -3.5 ± 2.6 at 10,000 on the new one) and
loses at Balanced (-23) and Max survival (-165); the Titans lose at Max threat (-65); Distilled Wisdom, the Chromatic
Resistance flask and the other Natural flasks lose 59 to 300. Windfury and Flametongue Weapon for Rockbiter lose 1,500
to 1,700 at every weight. Re-screened on the final builds (3,000, seed 1), dropping each of Greater Arcane, Nature
Power, Greater Intellect, the Demonic Rune and the rum, adding or dropping the Mageblood, other Zanza buffs and foods:
every current pick holds (the closest, Mageblood at Max survival +7.9 ± 11.2, and Nightfin Soup at Max threat -0.3 ±
1.4).

**Balanced takes Healing Stream Totem** (tank_rotations-20261112-2, later the same day). With the Agility cloak and
boots (tank_gear-20261112-1, -3) the knob search from Balanced's knobs moves `waterTotem` from 1 (Mana Spring) to 2
(Healing Stream), the move Max survival made with the Titans. Paired at 10,000: +17.2 ± 4.9 (seed 1) / +28.8 ± 4.6
(seed 2), 1,159.2 → **1,176.5** (seed 1, abtest's six-boss mean); every scoring boss gains but Chromaggus (-18 / -18;
Magmadar +39 / +37, Golemagg +24 / +37, Onyxia +30 / +54, Broodlord +31 / +33, Vaelastrasz -2 / +31). Balanced's talent
search on it (`-require spiritWeapons`, 1,000 a candidate) wandered on mana tie-breaks to `05140003-051031213351012201-050002`,
which loses at 10,000 (-13.4 ± 4.8 / -15.0 ± 4.5), so the talents stay.

## Presets (re-optimized 2026-11-04, tank round 3; rounds 4 and 5 the same day)

Three presets on the six scoring encounters (Magmadar, Golemagg, Onyxia, Broodlord, Vaelastrasz, Chromaggus), by
the tank objective at w = 1/3, 1 and 3 (`docs/tank_sim.md`): `tools/tankbench/req/shaman_warden_<preset>.json`,
`ui/warden_shaman/apls/level60_<preset>.apl.json`, the synthetic set (`gear_sets/phase1_synthetic_tank.gear.json`
for Max threat, `phase1_synthetic_tank_armor_trinket.gear.json` for Balanced and Max survival) and three builds in
`presets.ts`. The spec page's launch state is unchanged (blank gear, the default APL, the Level 60 talents).
Ragnaros is benched, not scored. Re-searched on the round 3 engine (master fd9538c: mana gains make threat without
the tank's threat multiplier, tank_mechanics-20261104-12; disjoint iteration seeds, tank_rotations-20261104-13) with
round 2's search (`rotopt` talent and knob searches, paired `abtest` at 3,000 iterations), plus the round 3 items
below. Every choice resting on a small paired difference was re-run on another seed (3,000 at seeds 2 and 3, or
10,000 at a disjoint `tankscore.Seed`). Raw output: `orchestration/tank_handoff/shaman/round3/` (`ab_summary.txt`
lists every A/B; tank_fairness-20261104-8).

The spec page's raid buffs, blessings (Kings, improved Might and Wisdom), debuffs (the tank's Sunder ramp, no Expose
Armor), consumables, professions (Engineering + Blacksmithing) and distance are the requests' (tank_fairness-20261104-6;
they had no blessings and Alchemy + Enchanting before), and each build loads its own flask.

| Preset | Race | Talents | On-use trinket (knobs) | Flask | Wrist / cloak / shield | Six-boss mean, own weight |
|---|---|---|---|---|---|---|
| Max threat | Night Elf | `05350002-050032313251012201-05` | attack power (990015) on cooldown (`armorTrinket=0`, `waterTotem=1`: Mana Spring Totem; `fireTotem=1`, `fireNova=1`, `fireLast=0`) | Natural Precision | 1886 / 1889 / 1886 | 2,514.7 |
| Balanced | Night Elf | `05050003-050032203451012201-050002` | armor (999216), held for the last 25 s (`armorTrinket=1`, `trinketAt=25`; `waterTotem=2`: Healing Stream Totem; `fireTotem=1`, `fireNova=1`, `fireLast=60`) | the Titans | 1248665 / 1219587 / 1886 | 1,176.5 |
| Max survival | Undead | `05050003-050230113541012201-050002` | armor (999216), held for the last 25 s (`armorTrinket=1`, `trinketAt=25`; `waterTotem=2`: Healing Stream Totem; `fireTotem=1`, `fireNova=1`, `fireLast=60`) | the Titans | 1248665 / 1889 / 1886 | -2,396.0 |

The table is 2026-11-12's (talents, knobs, flasks and the six-boss means at 10,000 iterations, seed 1; previous
section); the rest of this section is round 5's. The six-boss means were round 5's (tankbench, 1,000 iterations, seed 1; round 4's were 2,588.2 +/- 10.5, 944.3 +/-
17.5 and -3,718.0 +/- 54.0). Round 5's consumables (Round 5 below), at every preset: the Flask of Natural Precision,
Greater Arcane Elixir, Elixir of Nature Power, Elixir of Greater Intellect, the Demonic Rune, Rumsey Rum Black Label
and the Greater Stoneshield Potion, drunk by an explicit potion action just before `autocastOtherCooldowns`; Ground
Scorpok Assay at Max threat and Spirit of Zanza at Balanced and Max survival; the Mageblood Potion at Max threat and
Balanced. Rockbiter stays on the weapon (a stone or oil is the same temporary enchant). The UI's three builds load
the same consumables and the bench encounter.

Enchants: 1886 on the wrists is Superior Stamina (+9), 1248665 Superior Deflection (Defense +9); 1889 on the back is
Superior Defense (+70 armor), 1219587 Agility; 1887 on the feet is Greater Agility, 929 Greater Stamina. Since 2026-11-12
(tank_gear-20261112-1, -3; paired, 10,000, seeds 1 and 2) Max threat wears Superior Deflection wrists (+4.4 / +5.8, ± 1.5),
Agility on the back (+3.9 / +3.3) and Greater Agility boots (+5.4 / +5.0); Balanced Agility on the back (+4.2 / +4.2) and
Greater Agility boots (+9.0 / +10.0); Max survival keeps Superior Defense, Greater Stamina boots (-8.8 ± 10.3 at 3,000) and
Superior Deflection wrists. Balanced's UI set is its own (`phase1_synthetic_tank_armor_trinket_balanced.gear.json`);
1886 on the shield is Enchant Shield - Greater Stamina (+9). Head, legs, neck and gloves since 2026-11-12
(tank_fairness-20261112-2; every legal Phase 1 enchant screened slot by slot at 3,000 on seeds 1 and 2, the winners
paired at 10,000 on seeds 1 and 2 against Voracity Stamina 1507, Necklace Strength 1249019 and Gloves Superior Strength
1248640): all three wear Enchant Necklace - Agility (1249059). Max threat: Voracity Agility (1508) on the head and legs and
Gloves Natural Power (1248645, +20 nature damage), +21.7 / +18.5. Balanced: Voracity Agility and Gloves Greater Agility
(1887), +31.8 / +23.5 (Lesser Arcanum of Tenacity on the head and legs instead: +30.0 / +22.7). Max survival: Lesser
Arcanum of Tenacity (1504, +125 armor) on the head and legs and Gloves Greater Agility, +82.5 / +91.2 (Voracity Agility
instead: +64.7 / +57.7; adding Major Health on the chest: -1.8 / -3.7). On that set, Greater Agility boots and the Agility
cloak screened at +14.7 / +14.0 (3,000, seed 1) but did not hold at 10,000 (+12.5 / +1.8 and +5.8 / 0.0; both together
+1.5 / -15.3). All three keep the template's
other defaults (Stoneskin, Lightning Shield, Earth Shock on cooldown above Stormstrike with `shockFirst=1`, no self
heal). Since round 4 all three put Searing Totem down before the pull and recast it, with Fire Nova at 40% mana or
more, only in the fight's last 60 s (`fireLast=60`; Round 4 below), and all three eat Dirge's Kickin' Chimaerok Chops
and use the Goblin Sapper Charge (autocast with the other cooldowns, no filler explosive). `trinketAt` holds the
trinket for the fight's last N seconds unless the wait costs a use (`tools/rotopt/shaman_tank.go`, as the warrior's;
tank_rotations-20261104-9). Re-checked after 464deda made values under 20 s work and `armorTrinket=0` really mean
unused (tank_mechanics-16): 0, 5, 10, 15, 20 and 30 against 25, and no trinket line, at 10,000 iterations on seeds 1
and 2, all lose at Balanced and Max survival, so 25 stays (Round 4 below).

**Scores** (mean of the six at the preset's own weight; tankbench, 1,000 iterations, seed 1,
`orchestration/audit/ref/tank_bench/summary.tsv`; "before" is round 2's build on the round 3 engine, "round 3" the
round 3 build before round 4's fire totem, Mana Spring, sapper and food):

| Preset | Before | Round 3 | Round 4 | After (round 5) |
|---|---|---|---|---|
| Max threat | 2,276.0 +/- 7.9 | 2,430.0 +/- 10.4 | 2,588.2 +/- 10.5 | 3,169.8 |
| Balanced | 534.3 +/- 15.2 | 767.2 +/- 18.0 | 944.3 +/- 17.5 | 1,786.0 |
| Max survival | -4,356.7 +/- 46.9 | -3,970.8 +/- 47.7 | -3,718.0 +/- 54.0 | -1,988.2 |

Round 3's build against round 4's at 10,000 iterations (errors x1.25 for the shared seed, `docs/tank_sim.md` Seeds),
seed 1 / seed 2: Max threat 2,415.8 +/- 4.2 / 2,420.0 +/- 4.4 to 2,583.5 +/- 3.2 / 2,579.8 +/- 3.6; Balanced 732.2
+/- 6.5 / 737.5 +/- 5.8 to 943.8 +/- 6.2 / 937.7 +/- 5.8; Max survival -4,016.0 +/- 15.8 / -4,020.8 +/- 15.7 to
-3,719.8 +/- 15.4 / -3,682.7 +/- 16.3.

Paired, round 3's final request against round 2's (3,000 iterations, seed 1): Max threat +151.6 +/- 6.4, Balanced
+225.7 +/- 11.9, Max survival +388.7 +/- 30.0.

### Round 5: consumables (2026-11-04)

The standards item 11 ruling of 2026-11-04: natural flasks are legitimate (`naturalFlaskZone` is on in every request's
encounter and in the UI's), and the tanks get the full raid treatment, the Zanza slot, the Blasted Lands buffs and
the rum included, with the sapper and professions kept (tank_fairness-20261104-10). Paired `abtest`, the six-boss
mean at the preset's own weight, B minus A, 10,000 iterations at seeds 1 / 2, Max threat / Balanced / Max survival;
adopted only what wins on both seeds. Raw output: `orchestration/tank_handoff/shaman/round5/` (`out/` one slot at a
time against round 4's request, `out3/` and `out4/` the combinations, `out5/` before and after; `summ.py` lists them).

- **Flask** (round 4: Natural Accuracy at Max threat, Titans at the other two). The Flask of Natural Precision: +255.7
  +/- 3.4 / +203.8 +/- 7.0 / +32.6 +/- 19.3 (seed 2 +261.3 / +218.4 / -17.3 +/- 20.3, a tie alone at Max survival
  that wins in the combination, +68 +/- 15 / +43 +/- 15). Every other flask loses: Titans -38 at Max threat, Natural
  Accuracy -22 / -174 at Balanced and Max survival (seed 2 -21 / -222), Supreme Power, Distilled Wisdom, Chromatic
  Resistance, Natural Aggression and Natural Swiftness all below.
- **Spell elixirs**, each +26 to +35 on both seeds at every weight: Greater Arcane Elixir, Elixir of Nature Power,
  Elixir of Greater Intellect. Greater Firepower 0.
- **Demonic Rune**: +43.4 +/- 3.1 / +50.6 +/- 6.8 (seed 2 +53.8 / +54.3); at Max survival +33 / -13 alone, +46 / +55
  in the combination. Adopted at all three.
- **Zanza slot**, alone over none: Spirit of Zanza +50.6 / +105.6 / +344.1 (seed 2 +55 / +114.5 / +320.7), Ground
  Scorpok Assay +77.5 / +166 at Balanced and Max survival, Lung Juice Cocktail +54 / +172, R.O.I.D.S. +24, Cerebral
  Cortex Compound +26 to +35, Gizzard Gum, Sheen and Swiftness of Zanza 0. In the combination Scorpok beats Spirit at
  Max threat (+15.5 / +15.6); Spirit stays at the other two.
- **Rumsey Rum Black Label** +15 / +33.7 / +102.4 (seed 2 +17 / +33.4 / +101.6); Kreeg's Stout Beatdown -5.
- **Greater Stoneshield Potion.** `autocastOtherCooldowns` never drinks a defensive potion in a tank score
  (`HpPercentForDefensives` is 0), so a potion swap alone changes nothing past losing the old one (-104 / -84 at
  Balanced). With an explicit potion action ("Greater Stoneshield Potion on cooldown") before the autocast line, on
  the combination: +95 / +329 / +1,056 (seed 2 +93 / +323 / +1,050). The three APLs carry the line.
- **Mageblood Potion** (the mana-regen slot; raid-only, so since 2026-11-12 the Greater Mageblood Elixir, which measures +3.0 at Balanced and -0.7 at Max threat over it, tank_fairness-20261112-1) after the rest: +3.4 +/- 1.2 / +5.3 (seed 2 +3.0 +/- 1.2 / +6.0) at Max
  threat and Balanced, adopted; not at Max survival (the rune is the pick there).
- **Losers**: a Dense Weightstone, an elemental stone or a wizard or mana oil in place of Rockbiter -1,800 to -2,000;
  Greater Agility -50; Superior Defense -55 to -290; Fortitude Forever -21 to -120; Giants -5; Winterfall Firewater
  -2; no chili -9 to -41; Major Troll's Blood 0.
- **Before and after** (round 4's request against round 5's, paired, 10,000): Max threat 2,583.5 -> 3,167.7 (+584.3
  +/- 3.1; seed 2 2,579.8 -> 3,170.2, +590.2 +/- 3.4), Balanced 943.8 -> 1,790.7 (+846.8 +/- 6.0; seed 2 937.7 ->
  1,790.0, +852.3 +/- 5.7), Max survival -3,719.8 -> -1,999.5 (+1,720.2 +/- 15.4; seed 2 -3,682.7 -> -1,992.8,
  +1,689.9 +/- 15.1).

### Round 4 A/Bs (2026-11-04)

Paired `abtest`, scoring mean at the preset's weight, B minus A (A the preset as it stood before the item, B the
variant), 10,000 iterations at seeds 1 / 2 (disjoint, `tankscore.Seed`) unless marked "3k s3" (a 3,000-iteration
screen at seed 3). Adopted only what wins on both seeds beyond its error. Raw output:
`orchestration/tank_handoff/shaman/round4/` (`ab_summary.txt` lists every A/B).

- **`trinketAt` re-check** (tank_mechanics-16; against 25 s). Balanced: 0 (on cooldown) -134.4 +/- 4.6 / -135.6
  +/- 4.4, 5 s -178.8 / -181.3, 10 s -125.4 / -128.7, 15 s -77.2 / -80.0, 20 s -24.8 +/- 3.7 / -23.9 +/- 3.0, 30 s
  -6.4 +/- 4.0 / -13.0 +/- 3.9, no trinket line (`armorTrinket=0`) -219.3 / -215.2. Max survival: 0 -389.7 / -384.7,
  5 s -536.8 / -519.7, 10 s -400.0 / -360.6, 15 s -247.8 / -204.8, 20 s -76.1 +/- 11.0 / -59.7 +/- 9.6, 30 s -20.2
  +/- 10.1 / -36.0 +/- 10.6, no trinket line -642.4 / -630.2 (errors 3.8 to 4.6 at Balanced, 10 to 14 at Max
  survival). 25 s stays.
- **Fire totem and Fire Nova** (tank_rotations-20261104-14). The template's new `fireLast` knob gates the fire totem
  and Fire Nova to the fight's last N s (the totem still goes down before the pull). 3k s3 screens, Searing Totem with
  Fire Nova: gate 0 (whole fight) / 30 / 40 / 50 / 60 / 75 / 90 s, Max threat +34.0 / +22.6 / +33.8 / +43.0 / +44.5 /
  +42.4 / +40.3 (+/- 9), Balanced (with Mana Spring) +34.8 / +42.6 / +48.0 / +44.1 / +42.5 / +42.4 / +38.1 (+/- 13 to
  15), Max survival (with Mana Spring) +41.6 / +42.5 / +50.9 / +27.4 / +23.1 / +31.9 / +30.0 (+/- 33 to 40). At 60 s,
  Searing without Fire Nova -9.8 / +0.2 / +20.8, Magma Totem +24.1 / +22.6 / +76.9, Flametongue Totem -162.7 / -144.4
  / -64.6; Fire Nova's mana floor 0 / 20 / 60% within 8 of 40%. Confirmed: Max threat 60 s +53.9 +/- 4.7 / +47.7 +/-
  5.0 (40 s +39.0 / +36.0); Balanced with Mana Spring 60 s +67.7 +/- 7.1 / +59.2 +/- 6.4 (40 s +53.4 / +42.6; without
  Mana Spring +55.4 / +43.8); Max survival 60 s +59.3 +/- 17.7 / +55.7 +/- 18.2 (with Mana Spring +56.5 / +60.0, 40 s
  +43.7 / +54.8, Magma +46.0 / +85.2 +/- 18). All three take Searing Totem and Fire Nova in the last 60 s.
- **Mana Spring before the pull** (tank_rotations-20261104-15; `waterTotem=1`, Max threat's). Balanced +9.3 +/- 2.3 /
  +10.8 +/- 1.8: taken. Max survival +5.3 +/- 5.8 / +4.0 +/- 5.1: inside its error, not taken.
- **Explosives** (standards rule 11; the Warden has Engineering). On the fire totem builds, the Goblin Sapper Charge
  (`sapperExplosive`, a DPS cooldown the rotation's `autocastOtherCooldowns` line fires): Max threat +101.0 +/- 4.5 /
  +99.0 +/- 4.7, Balanced +102.7 +/- 6.9 / +101.2 +/- 6.9, Max survival +75.1 +/- 18.2 / +115.1 +/- 17.5. A filler on
  top (sapper and filler against the preset): Goblin Land Mine +94.8 / +94.4, +94.9 / +94.1, +68.7 / +100.0; Dense
  Dynamite and Thorium Grenade +94.4 / +93.9, +94.5 / +93.7, +68.4 / +99.7; EZ-Thro Dark Bomb as the Land Mine within
  0.4. Every filler costs 4 to 15 against the sapper alone (the GCDs, as for the bear). All three take the sapper, no
  filler.
- **Food** (tank_fairness-20261104-10; against the Smoked Desert Dumpling, with the sapper). Dirge's Kickin' Chimaerok
  Chops: Max threat +12.7 +/- 1.6 / +13.2 +/- 1.6, Balanced +41.1 +/- 1.8 / +40.1 +/- 1.9, Max survival +161.5 +/- 4.5
  / +167.1 +/- 5.1. Grilled Squid +0.4 / +2.0, +5.4 / +4.8, +5.0 / +1.1 (+/- 1.5 to 7). Blessed Sunfruit -9.9, -9.8,
  -9.2 on both seeds. All three take Dirge's chops.

### Round 3 A/Bs

Paired `abtest` (`tools/audit/tank_rotations/abtest`), scoring mean at the preset's weight, B minus A, B the variant;
3,000 iterations at seed 1 unless a seed is named (s2, s3: 3,000 at seed 2 or 3; 10k sN: 10,000 at disjoint seed N).
Listed as Max threat / Balanced / Max survival where all three were run.

- **Armor trinket hold** (tank_rotations-20261104-9, -10). 999216 held for the last N s against the round 2 request
  (autocast on cooldown at Max survival, 990015 on cooldown at Balanced). Max survival: 20 s +265.6 +/- 19.7, 25 s
  +344.3 +/- 15.3, 30 s +321.9 +/- 16.4, 35 s +252.1 +/- 18.0, 40 s +159.8 +/- 16.6, 45 s +83.1 +/- 12.8. Balanced:
  20 s +78.6 +/- 7.0, 25 s +98.5 +/- 7.3, 30 s +97.2 +/- 7.3, 35 s +67.3 +/- 6.5, worn on cooldown -6.5 +/- 3.7. Max
  threat keeps 990015: 999216 held 30 s -26.0 +/- 2.2, on cooldown -65.1 +/- 1.3 (on the final build: -12.2 +/- 4.2 and
  -68.2 +/- 1.5). On the final builds, 990015 on cooldown in place of the held armor trinket: Balanced -112.3 +/- 8.1,
  Max survival -544.6 +/- 25.0.
- **25 s against 30 s** (final builds): Balanced +6.9 +/- 5.8, s2 +6.4 +/- 5.4, 10k s4 +16.6 +/- 4.2, 10k s5 +8.9 +/- 3.2;
  Max survival +20.6 +/- 15.5, s2 +23.4 +/- 17.7, 10k s3 +7.6 +/- 9.2, 10k s4 +29.5 +/- 10.2, 10k s5 +33.2 +/- 11.0.
  35 s against 30 s loses (-29.4 +/- 5.9, -100.8 +/- 15.8). So 25 s, confirmed only at 10,000 (call (i)).
- **Flask** (tank_fairness-20261104-7). Natural Accuracy against Titans at Max threat: +31.1 +/- 3.8 on round 2's
  build; on the final build Titans against Natural Accuracy -24.5 +/- 5.0. Max threat takes Natural Accuracy.
- **Shield** (tank_gear-20261104-3). Lesser Block (863) against Greater Stamina (1886): -8.6 +/- 3.4 / -12.6 +/- 4.9 /
  -40.6 +/- 12.7; s2 -10.9 +/- 3.6 / -16.6 +/- 4.3 / -71.6 +/- 14.9. 1886 stays.
- **Wrists** (tank_gear-20261104-3). Superior Deflection (1248665) against Superior Stamina (1886). Max threat +1.2 +/-
  4.1, s2 -1.4 +/- 5.2, final build -2.1 +/- 6.7: keeps 1886. Balanced +2.3 +/- 7.8, s2 +1.0 +/- 7.0, final +4.7 +/- 7.6,
  10k s5 +6.7 +/- 4.3, 10k s6 (Night Elf, 25 s) +15.8 +/- 4.3. Max survival +6.7 +/- 20.4, s2 +26.8 +/- 19.9, 10k s3
  +15.4 +/- 10.6, final +7.9 +/- 22.0, 10k s4 +26.1 +/- 12.2, 10k s5 +34.5 +/- 12.2, 10k s6 (Undead, 25 s) -2.7 +/- 10.5.
  Balanced and Max survival take 1248665 on the pooled gain (call (h)).
- **Cloak** (tank_gear-20261104-7). Agility +5 (1219587) against Superior Defense (1889): +1.0 +/- 2.6 / -5.1 +/- 5.2 /
  -15.2 +/- 13.6; s2 +3.0 +/- 3.3 / -5.8 +/- 4.8 / -21.2 +/- 12.1; final builds +1.8 +/- 3.0 / -1.8 +/- 4.5 / -7.2 +/-
  15.3. 1889 stays everywhere (Max threat's gain is inside its error).
- **Race** (tank_fairness-20261104-8; against Tauren, on round 3's Tauren talents). Night Elf +104.2 +/- 8.7 / +111.3
  +/- 10.4 / -83.6 +/- 31.8, s2 +99.9 +/- 9.2 / +102.2 +/- 11.7 / -31.0 +/- 34.4 (Elune's Light). Undead +45.5 +/- 11.5 /
  +76.9 +/- 12.4 / +14.0 +/- 33.7; at Max survival s2 +100.1 +/- 30.9, 10k s3 +38.9 +/- 18.4, and on the final build
  Tauren against Undead 10k s7 -27.5 +/- 17.5: about +43 +/- 11 pooled over the four (call (f)). Undead against Night
  Elf: Max threat -58.7 +/- 13.7, Balanced -34.4 +/- 11.0. Troll +38.2 / +2.3 / -188.6, Gnome +33.9 / -5.9 / -159.9,
  Orc +27.8 / -3.5 / -176.2, Human, Dwarf and the Skyborne lower at every weight. So Night Elf at Max threat and
  Balanced (call (g)), Undead at Max survival.
- **Talents** (rotopt talent searches on Tauren, then again after the race change; paired against the round 2 build).
  Max threat, Elemental Devastation 1 and Toughness 5 (Night Elf): +16.4 +/- 6.4, s2 +17.2 +/- 6.3. Balanced, Elemental
  Warding 3 and Toughness 4: +5.0 +/- 1.7, s2 +8.1 +/- 1.9. Max survival, Guardian Totems 2 for Improved Lightning
  Shield 2: +3.3 +/- 1.2, s2 +2.5 +/- 1.5, s3 +3.1 +/- 3.4 (call (j)). The knob searches moved nothing else (Balanced's
  `shockMana=10`: +0.7 +/- 1.2, not taken).

### Round 2 (for the record)

**Earth Shock above Stormstrike** (tank_rotations-20261104-4). Earth Shock makes about 2,225 threat a cast against
Stormstrike's 1,500, and with Stormstrike first the shock waited whenever both were ready. Swapping the two lines in
the round 1 requests (paired, `tools/audit/tank_rotations/abtest`, scoring mean): Max threat +91.6 +/- 10.6 (1,000
iterations), Balanced +105.0 +/- 12.5 (3,000), Max survival +129.3 +/- 33.1 (3,000); every scoring boss gains at
every weight (Broodlord +141 / +104 / +107). The template's `shockFirst` knob (default 1) emits it, and the knob
searches kept it at all three weights.

**Talents, round 2** (`rotopt -talent-search -require spiritWeapons`, 1,000 iterations a candidate, 3,000 to confirm; paired
diffs at 3,000 with `abtest`):
- Max threat: from the round 1 build, Elemental Devastation 3 for a point of Elemental Warding and Toughness 2:
  +26.7 +/- 2.8. Balanced's talents score worse here (2,228 against 2,289).
- Balanced: Elemental Weapons 3 for Shamanistic Focus: +27.5 +/- 11.5 against the round 1 build (B). Max
  survival's build scores 16.3 +/- 11.1 lower here, within the noise; the search's winner stays.
- Max survival: a point of Elemental Warding for a point of Flurry: +71.5 +/- 33.1 against B. Balanced's build
  scores 44.2 +/- 32.5 lower, within the noise (the Broodlord death floor near 5% makes this weight noisy); the
  search's winner stays.
- So the three builds differ, but Balanced's and Max survival's are not separable at their own weights by more than
  the paired error (tank_fairness-20261104-4): either build serves both.

**Knobs, round 2** (600 to 1,000 iterations, 3,000 to confirm): the defaults everywhere (Stoneskin, Lightning Shield, Earth
Shock on cooldown above Stormstrike, no fire totem, no self heal), plus Mana Spring Totem at Max threat. Max
survival's dedicated armor trinket line lost 3.3 +/- 1.2 to the autocast line: since round 2's engine a scoring run
never reads health, so the autocast line fires the defensive trinket on cooldown too. Round 3's `trinketAt` hold beats both (above).

**Hit** (tank_fairness-20261104-3; the gear auditor found the set over the hit cap). The set's 5.5% gear hit plus
Tidal Focus's 5% is 10.5% against the 8% melee cap, 2.5% over. Gear hit counts for spells in Forever
(`unifyEquipHitAndCrit`), so it still helps Earth Shock, and the boss's parry and block, which hit doesn't remove,
are the round 2 engine's. Moving 25 hit rating to crit (neck and a ring, budget neutral, to 8.0% with Tidal Focus) on
the Balanced request, 3,000 iterations: 415 against 404, a tie (unpaired: two databases). Round 1 measured a loss
(359 against 463) on the older engine. The set keeps the row's hit (standards 4: no mix sims better).

**Boots enchant** (tank_gear-20261104-6). Greater Stamina (929) against the vendor kits, paired, 3,000 iterations:
Mystic Rugged Armor Kit (1254966: 6 spell power, 40 armor) Max threat +2.7 +/- 0.5, Balanced -6.6 +/- 2.7, Max
survival -26.1 +/- 6.7; Forceful Rugged Armor Kit (1254965: 10 attack power, 40 armor) +1.6 +/- 0.5, -7.7 +/- 2.7,
-27.2 +/- 6.7. Greater Stamina wins at Balanced and Max survival. Max threat shares Balanced's set, so it keeps
Greater Stamina too (call (e)).

**Gear and mechanics checked for correctness, not for the gap** (tank_fairness-20261104-3, user ruling 2026-11-04,
standards 40). The Warden's 2 to 3 times higher healing need than the plate tanks is a real class gap: a mail class
with much less defense built in, which won't necessarily be a viable raid tank. Nothing here compensates for it. What
was checked: the set sits at its Tanks row (standard 36, `tools/database/forever_synthetic_items_shaman_tank.go`);
the hit over the cap ties (above); the shield block, Spirit Weapons' parry, Lightning Shield, Stoneskin and
Rockbiter have tests that fail when they break (Tests); the abilities and talents audit (Rules audit) has no wrong
row.

### Numbers (1,000 iterations, all seven encounters)

`tankbench -iters 1000` on each encounter, seed 1 (`orchestration/audit/ref/tank_bench/summary.tsv`). Healing needed is
at 5% deaths, in HPS; the score is the objective at the preset's own weight, +/- its error. The mean is over the six
scoring encounters (Ragnaros is reported, not scored). The other tanks' rows are in the same file.

**Max threat** (`shaman_warden_max_threat.json`, score at w = 1/3)

| Encounter | TPS | SafeDPS | DTPS | Healing needed | Death floor | Score |
|---|---|---|---|---|---|---|
| Magmadar | 1,496 | 1,630 | 512 | 479 | 0.0% | 2,702 +/- 13 |
| Golemagg | 1,488 | 1,614 | 558 | 551 | 0.0% | 2,586 +/- 13 |
| Onyxia | 1,621 | 1,787 | 646 | 637 | 0.0% | 2,476 +/- 17 |
| Broodlord | 1,558 | 1,708 | 744 | 1,526 | 0.5% | 1,890 +/- 50 |
| Vaelastrasz | 1,671 | 1,912 | 1,399 | 1,731 | 0.0% | 3,097 +/- 25 |
| Chromaggus | 1,335 | 1,455 | 891 | 1,092 | 0.0% | 1,829 +/- 12 |
| Ragnaros | 1,392 | 1,515 | 1,421 | 2,754 | 0.3% | 130 +/- 61 |
| Mean of the six | 1,528 | 1,684 | 792 | 1,003 | 0.1% | 2,430.0 +/- 10.4 |

**Balanced** (`shaman_warden_balanced.json`, score at w = 1)

| Encounter | TPS | SafeDPS | DTPS | Healing needed | Death floor | Score |
|---|---|---|---|---|---|---|
| Magmadar | 1,423 | 1,536 | 484 | 436 | 0.0% | 1,551 +/- 18 |
| Golemagg | 1,412 | 1,511 | 524 | 496 | 0.0% | 1,293 +/- 21 |
| Onyxia | 1,530 | 1,671 | 591 | 542 | 0.0% | 538 +/- 36 |
| Broodlord | 1,470 | 1,595 | 684 | 1,110 | 0.2% | -140 +/- 87 |
| Vaelastrasz | 1,555 | 1,775 | 1,250 | 1,445 | 0.0% | 1,729 +/- 36 |
| Chromaggus | 1,266 | 1,376 | 848 | 1,050 | 0.0% | -368 +/- 28 |
| Ragnaros | 1,317 | 1,417 | 1,293 | 2,133 | 0.0% | -3,902 +/- 152 |
| Mean of the six | 1,443 | 1,577 | 730 | 846 | 0.0% | 767.2 +/- 18.0 |

**Max survival** (`shaman_warden_max_survival.json`, score at w = 3)

| Encounter | TPS | SafeDPS | DTPS | Healing needed | Death floor | Score |
|---|---|---|---|---|---|---|
| Magmadar | 1,375 | 1,468 | 492 | 416 | 0.0% | -1,418 +/- 44 |
| Golemagg | 1,368 | 1,446 | 534 | 482 | 0.0% | -2,151 +/- 51 |
| Onyxia | 1,448 | 1,580 | 599 | 517 | 0.0% | -4,868 +/- 96 |
| Broodlord | 1,396 | 1,507 | 698 | 1,106 | 0.2% | -6,940 +/- 245 |
| Vaelastrasz | 1,461 | 1,691 | 1,260 | 1,447 | 0.0% | -2,092 +/- 70 |
| Chromaggus | 1,342 | 1,410 | 865 | 1,030 | 0.0% | -6,356 +/- 58 |
| Ragnaros | 1,286 | 1,358 | 1,291 | 2,069 | 0.0% | -16,886 +/- 619 |
| Mean of the six | 1,398 | 1,517 | 741 | 833 | 0.0% | -3,970.8 +/- 47.7 |

## Calls for the user

- **(a) Earth Shock's threat multiplier.** Default 2 (Season of Discovery's, wowsims/sod, tier 2, and Classic's;
  the tooltip says "causes a high amount of threat"; no Forever source gives a number). wowsims/forever uses 1.
  Measured in round 1 on the then Balanced preset (1,000 iterations, means of the six, older engine): at 1, TPS
  1,589 to 1,399, SafeDPS 1,745 to 1,526, scores 2,493 / 501 / -5,475 to 2,056 / 63 / -5,913. With the shock now
  above Stormstrike more of the Warden's threat rests on it. A /combatlog threat check in game would settle it.
- **(b) The defensive on-use trinket.** Default: the synthetic 1,665 armor for 20 s every 2 min (999216), held for
  the last 25 s, in Balanced and Max survival, as the defensive counterpart of the Enhancement set's attack power
  trinket at the same item level. The alternative, the attack power trinket, costs Balanced 112.3 +/- 8.1 and Max
  survival 544.6 +/- 25.0 on round 3's builds.
- **(c) Rockbiter's threat on Stormstrike** (to verify; ruled 2026-11-04, standards 40). The sim keeps 72 x weapon
  speed on specials, as on white swings (wowsims/forever's reading); wowsims/sod gives a special a flat 72, which
  would take about 35 TPS (2.5%) off the Warden. One Stormstrike's threat in a combat log settles it
  (`sim/shaman/rockbiter_weapon.go` says so; the flat case is
  `orchestration/audit/ref/tank_mechanics/patches/shaman_02_rockbiter_special_flat_CONDITIONAL.patch`).
- **(d) Hit above the melee cap.** Default: the set keeps the row's 5.5% gear hit (10.5% with Tidal Focus); moving
  the 2.5% over the cap into crit ties at Balanced on the round 2 engine (415 against 404, unpaired).
- **(e) Boots for Max threat.** Max threat would gain 2.7 +/- 0.5 (0.1%) from Mystic Rugged Armor Kit on the boots,
  but the sets share one boots enchant, and Greater Stamina wins by 6.6 at Balanced. Default: Greater Stamina
  everywhere; Mystic Rugged boots on Max threat's set alone is the alternative.
- **(f) Undead at Max survival** (tank round 3). Undead beats Tauren by about +43 +/- 11 pooled over four seeds
  (+14.0 +/- 33.7, s2 +100.1 +/- 30.9, 10k s3 +38.9 +/- 18.4, 10k s7 +27.5 +/- 17.5), but no single run clears 3
  sigma. The gain rests on how the sim models Touch of the Grave's drain, and Undead loses Tauren's Broodlord margin
  (the 5% health that kept his death floor lowest). Default: Undead; Tauren is the alternative.
- **(g) Faction.** Night Elf (Alliance) now leads Max threat and Balanced by about 100 (+104.2 +/- 8.7 and +111.3 +/-
  10.4 against Tauren, seed 2 alike) through Elune's Light, which reverses round 1's Tauren-first conclusion; Undead
  is the best Horde race there (-58.7 +/- 13.7 and -34.4 +/- 11.0 against Night Elf). The faction is the user's choice.
  Default: Night Elf at Max threat and Balanced, Undead at Max survival (a mixed-faction set of presets).
- **(h) Superior Deflection wrists at Balanced and Max survival.** Taken on small pooled gains: Balanced +1.0 to +15.8
  over five runs, Max survival +6.7 to +34.5 over six runs but -2.7 +/- 10.5 on the final build at 10k seed 6.
  Default: Superior Deflection (1248665); Superior Stamina (1886) is the alternative.
- **(i) The armor trinket held 25 s against 30 s.** 25 s wins at both weights, but only the 10,000-iteration runs
  separate them (Balanced +16.6 +/- 4.2 and +8.9 +/- 3.2, Max survival +29.5 +/- 10.2 and +33.2 +/- 11.0; the
  3,000-iteration runs and 10k s3 at Max survival, +7.6 +/- 9.2, are inside their errors). Default: 25 s.
- **(j) Max survival's talents.** Guardian Totems 2 for Improved Lightning Shield 2 gains +3.3 +/- 1.2 at seed 1,
  but under twice the error at seeds 2 and 3 (+2.5 +/- 1.5, +3.1 +/- 3.4). Default: the new build; round 2's
  `05050003-050032113541012201-050002` is the alternative.
- The Warden's higher healing need is not a call any more: the user ruled it a real class gap (standards 40, above).

## Why the Warden takes more damage than the other tanks (2026-11-04)

The preliminary bench before the update put the Warden far behind the other tanks on survival. On the realigned sim,
the balanced request (before the searches) against the other tanks' current requests, 1,000 iterations:

| Healing needed (HPS) | Magmadar | Golemagg | Onyxia | Broodlord | Vaelastrasz | Chromaggus | Mean | Mean without Broodlord |
|---|---|---|---|---|---|---|---|---|
| Warden | 509 | 574 | 654 | 3,007 (floor 4.3%) | 1,706 | 1,152 | 1,267 | 919 |
| Protection warrior | 383 | 435 | 555 | 1,882 (2.1%) | 1,504 | 922 | 947 | 760 |
| Protection paladin | 376 | 439 | 522 | 1,341 (0.6%) | 1,494 | 913 | 848 | 749 |
| Feral tank | 447 | 496 | 572 | 1,237 (0.8%) | 1,624 | 979 | 893 | 824 |

Two different things, both a **real class gap**, not the gear or the rotation:

1. **Damage taken is about 30% higher** (Magmadar: Warden 537 DTPS, warrior 409, paladin 416, bear 455). Taking
   the gap apart with bonus stats on the Warden (Magmadar, 1,000 iterations): the paladin's armor (+1,200, 54.2% to
   58.2% reduction) takes its DTPS to 494 (-8%); a block chance like Holy Shield's or Shield Block's (+45%) to 500
   (-7%); and the warrior's Defensive Stance takes 10% more off. 0.92 x 0.93 x 0.90 of 537 is 413, the warrior's
   409. So the Warden lacks exactly the three things the other shield tanks have: plate (or the bear's armor
   multiplier), an active block that also pushes crushing blows off the table (the Warden takes 15% crushing blows,
   the warrior 4%), and a stance or talent that cuts damage taken. Forever's shaman tree has none of them (no
   Shield Specialization, no armor talent: Toughness is Stamina, Shield Block is a warrior's), and the notes add
   none. Its own mitigation is Stoneskin Totem (36 off a hit before armor, about 3% at these bosses), 5% passive
   block and Spirit Weapons' parry; its avoidance (33%) is in fact the best of the shield tanks.
2. **Broodlord is a health threshold.** Mortal Strike (11,000 to 16,000 before armor) between two heals kills a
   tank whose health is close to it. The Tauren Warden (8,536 health) has a 4.3% death floor, so healing needed at
   5% deaths is unbounded; an Orc, Troll or Dwarf Warden (8,120 to 8,140) is at 6.5 to 8.1%, past the 5% line
   ("none works"). +1,500 health takes the floor to 0.1% (healing needed 1,315), +1,200 armor to 0.5% (1,440). The
   warrior and paladin get there with less health through Shield Block, Holy Shield and their reductions.
   Broodlord therefore decides the Warden's score and its race: Tauren is the only race that survives him. (On round
   3's engine and builds the death floor is 0.2 to 0.5% for every preset, and Night Elf and Undead score higher:
   Presets.)

The gear isn't the cause: the set's Stamina share is the row's (23.9%) and the Warden has the most health of the
shield tanks (8,536 against 7,079 and 6,994); what a hybrid set spends on Intellect, Spirit and spell power the other
tanks spend on threat stats, not on mitigation. The rotation can only add Stoneskin (on), Healing Stream Totem or a
health-blind Lesser Healing Wave (searched below).

## Realigned to client 70753 and the November 2 notes (2026-11-04)

Master's shaman round (`orchestration/round_70753/shaman.md`) brought the notes' shaman lines, standards 37 (the
notes' value where client 70753 and the hotfixes don't carry it). Client 70753 changed no tank row of the shaman
(`db2diff.py 70707 70753`: 173 shaman spells, all only by `SpellMisc.Attributes_14`): Rockbiter Weapon and its threat
proc 20871, Spirit Weapons (16268), Lightning Shield, Stoneskin Totem, Anticipation, Toughness, Guardian Totems,
Elemental Warding and the heals are as audited below. What moved for the Warden:

- **Notes-only values the Warden uses** (the round's list, all under `IsForever()`): Earth Shock at Era's base
  (517-545 at 60, threat 2 x damage: the biggest threat change), Flurry 6 to 30%, Lightning Bolt rank 10 at
  334-372; then Stormstrike 150% and 6 s, Mental Dexterity 50 / 100 / 150% and Maelstrom Weapon 4 PPM a point,
  which Blizzard's November 10 class tuning replaced (below): Stormstrike 100% and 8 s, Mental Dexterity
  33 / 66 / 100%, Maelstrom Weapon (5 + rank) PPM (notes-only; client 70888 still has 150% and 6 s).
- **The tree:** Improved Ghost Wolf moved to row 4 (after Stormstrike), so every Enhancement string moved: the Level
  60 build `35050013-0502303313551102` is `35050013-0502333135510102` (the same points). Remapped in the tests'
  base build and talent indices, the bench request and the UI preset (the DPS worker had done the preset).
  Lava Burst's square is Call of Thunder, which no Warden build takes.
- **Salvation is gone:** SafeDPS's reference member makes 1 threat a point of damage (was 0.7); no tank request had
  Salvation on the tank.
- **Races:** Tauren Endurance is 5% health and nothing else (A2), still the Warden's best by the objective (below).
- **Merge check:** `git diff f9cc1ecaf HEAD -- sim/shaman/` (my last commit before the update against today): no
  mechanic applied twice or lost. Rockbiter's threat on Stormstrike (72 x speed) and Spirit Weapons' 1.302 are
  intact, master's round only scaled Stormstrike's damage. The Warden golden (`TestWardenShaman.results`) had been
  regenerated on master without the 1.302 (TPS 0.15% low) and without the synthetic set's two rows: regenerated.
- **Tests:** `TestForeverWardenTalents` reads Mental Dexterity at 33 / 66 / 100% (50 / 100 / 150% until the
  November 10 tuning) and the new talent indices, `TestForeverWardenAbilities` Stormstrike's 8 s (6 s until then).

## Rules audit

Client 70753 (`tools/sources/wago.py`, no hotfix rows logged for it: the October 30 hotfix is in its tables,
`orchestration/update_ledger.md`) and hotfix push 113380 (build 70755, 2026-11-04, wago.tools/hotfixes region 91:
the shocks and Stormstrike), forever-wiki/classes/shaman.md, `shaman_audit.md` (the shaman's DPS audit,
whose sections are cited as 6.4, 7.4 and so on). Tiers for what the client doesn't state: (1) a Forever log or
beta measurement, (2) wowsims/forever's reading, (3) Classic's (wowsims/classic).

### Abilities

| Ability | Client (70753) | Sim | Source, tier | Verdict |
|---|---|---|---|---|
| Rockbiter Weapon rank 7 (16316, passive 16313) | 653 attack power at 60 (554 + 16.5 a level to 62), 150 mana, 60 min | 653, x1.07 / 1.13 / 1.2 with Elemental Weapons | client; beta sheet at 30 (7.4), tier 1 | right |
| Rockbiter's threat (proc 20871) | effect 77 (script), 72 points; cast off the passive on done melee auto and special hits (ProcTypeMask 20) | 72 x weapon speed on each landed white hit; **was a flat 72 on Stormstrike, now 72 x speed** | client mask and 20871's designer note ("Effect 1 points = desired threat per second, before misses"), tier 1 for white hits; on Stormstrike the x speed is wowsims/forever's reading, tier 2 (wowsims/sod gives a special the flat 72); **to verify** (user ruling 2026-11-04, standards 40: keep x speed until a combat log shows one Stormstrike's threat; call (c)) | **fixed** |
| Spirit Weapons' threat (16268) | effect 1 -30% threat, all schools; effect 2 +86 on Rockbiter's passive threat effect (0 alone) | **0.7 x 1.86 = 1.302 under Rockbiter (was 1.3)**, 0.7 otherwise; picked by the imbue in the options or consumes | client, tier 1 | **fixed** |
| Spirit Weapons' parry | not a client effect ("Gives a chance to parry") | can parry: 5% base, plus Defense, less 0.6% for the boss's level | beta (6.4: 18 of 380 below us, low above), tier 1 | right |
| Shield block | Classic's: 5% base with a shield, Defense on top | 5% (less 0.6% vs a level 63), block value = the shield's + Strength / 20 - 1 | Classic, tier 3; no shaman shield on the beta (6.4) | right |
| Earth Shock rank 7 (10414) | client 70753: 301 centre; **hotfix push 113380 (build 70755, 2026-11-04; SpellEffect 688559): Era's base, 531, 517-545 at 60**, as the November 2 notes said; 38.6% of spell power, 450 mana, 6 s shared shock cooldown (category 19), 1.5 s GCD, "Causes a high amount of threat" | 517-545 (master's shaman round, `earth_shock.go`; the x1.08 is 1.0 now); threat 2 x damage | damage: hotfix 113380, client data, tier 1; threat multiplier: 2, Season of Discovery's (wowsims/sod earth_shock.go `ThreatMultiplier = 2`), tier 2, and Classic's (wowsims/classic), tier 3; no vmangos; wowsims/forever gives 1 (no source); the tooltip says "high" | right, **call (a)** |
| Flame Shock, Frost Shock | no threat text | threat 1 x damage | client text, tier 3 | right |
| Stormstrike (17364) | client 70753: 100% weapon damage, 8 s; **hotfix push 113380 (build 70755; SpellEffect 1368308, SpellCooldowns 55241): 150% normalized weapon damage, 6 s**, as the notes said; 8% of base mana, +20% to the next Lightning Bolt, Chain Lightning or Earth Shock; **Blizzard's November 10 class tuning (hotfix article 24315281): 100% weapon damage, 8 s** (no client row yet: client 70888 still 150 / 6000 ms) | 100%, 8 s (Nov 10 tuning, notes-only; `stormstrike.go`; a flat "+N damage" stays 1 x); threat 1 x damage plus Rockbiter's 72 x speed (not scaled by the weapon damage percent) | the Nov 10 notes, notes-only (standards 37); before them hotfix 113380 (client data, tier 1), client, 7.7 | right |
| Lightning Shield rank 7 (10432, orb 26363) | 3 orbs of 178 + 26.7% of spell power, 370 mana, 10 min, one orb every 3.5 s (ProcCategoryRecovery) at whoever hits us | the same; orbs roll spell hit, never crit (beta, 7.15); threat 1 x damage | client and beta, tier 1 | right |
| Stoneskin Totem rank 6 (10408, buff 10405) | -30 physical damage taken (aura 14), 210 mana, 5 min | -30 before armor, x1.1 / 1.2 with Guardian Totems; a crit or crushing blow doubles what's left | client; beta fit before armor (6.3, 7.29), tier 1 | right |
| Strength of Earth Totem rank 4 (10442) | Strength (block value at Strength / 20) | the raid's (another shaman's, raid-wide since 70425) in the bench; ours a knob | the DPS audit, 7.21 | right |
| Healing Stream Totem rank 5 (10463) | 11 every 2 s (tooltip, before talents) | the same; crits 5% (7.28); a periodic heal, so no Ancestral Healing | client, beta | right |
| Lesser Healing Wave, Healing Wave | 775-865 in 1.5 s; 1508-1722 in 3 s | the same; heals make 0.5 threat a point, split over the enemies | client; Classic threat, tier 3 | right |
| Casting while tanking | | no dodge, parry or block while casting, only miss (`OutcomeEnemyMeleeSpecial`, the white table) | beta (6.4: too few casts to tell), tier 3 | right, as the engine's rule |
| Totems' own damage | | makes no threat for us (Searing and Magma Totem's attacks have threat 0); Fire Nova (a shaman spell under Forever) makes ours | Classic: a totem holds its own threat, tier 3 | right |

### Talents a tank build can reach

| Talent | Client (70753, trait curves of 70630, unchanged since) | Sim | Verdict |
|---|---|---|---|
| Anticipation (16254) | 2 / 4 / 6% dodge | the same | right (beta export, 8.26) |
| Toughness (16252) | 2% Stamina a point | the same | right |
| Spirit Weapons (16268) | parry, threat 0.7 / 1.302 | above | **fixed** (1.302) |
| Guardian Totems (16258) | +10% a point to Stoneskin and Windwall | the same on Stoneskin; Windwall and Grounding have nothing to stop | right |
| Elemental Weapons (16266) | Rockbiter's attack power +7 / 13 / 20% | the same | right |
| Improved Lightning Shield (16261) | orbs +5% a point | the same | right |
| Shamanistic Focus (1223030) | shocks and Lightning Shield cost 45% less | the same, multiplied with Convection | right |
| Thundering Strikes (16255) | 1% crit a point, melee and spells | the same | right |
| Mental Dexterity, Mental Quickness | client: attack power from 33 / 67 / 100% of Intellect; the November 2 notes raised it to 50 / 100 / 150%, **the November 10 tuning set it to 33 / 66 / 100%** (notes-only); spell power from 15% a point | 33 / 66 / 100%; 15% a point | right |
| Stormstrike, Improved Stormstrike (1223031) | at 2/2 every dodge or parry of ours resets Stormstrike (8 s cooldown since the Nov 10 tuning); 50% regen while casting, 15 s | the same | right (beta, 8.32) |
| Flurry, Maelstrom Weapon, Rage of the Farseer | DPS talents; the notes: Flurry 6 / 12 / 18 / 24 / 30%, Lightning Bolt's top rank +76%; the Nov 10 tuning: Maelstrom Weapon (5 + rank) procs a minute on the main hand's speed, 20% a stack at every rank (was 4 PPM a point); Rage of the Farseer 30% attack speed for 25 s, 3 min | as master's shaman rounds (`round_70753/shaman.md`, `round_notes_2026-11-11/shaman.md`) | right |
| Convection, Concussion, Reverberation | shock costs -2% a point; Earth Shock +1% a point; shock cooldown -0.2 s a point | the same | right |
| Elemental Warding (28996) | -3 / 7 / 10% Fire, Frost and Nature damage taken | the same | right; worth something at Chromaggus, Onyxia, Ragnaros |
| Elemental Devastation, Elemental Focus | spell crits give 3% melee crit a point for 10 s; Clearcasting | as the DPS audit | right |
| Tidal Focus (16179) | 1% hit a point (melee and spells), heals 1% cheaper | the same | right; Earth Shock misses 17% of the time against the boss without hit |
| Improved Reincarnation (16184) | +2% health a point | the same | right |
| Ancestral Healing (16176) | a crit of our direct heals gives the target +8 / 17 / 25% armor for 15 s; its ProcTypeMask (0x15550) leaves out periodic heals | the same: Healing Stream's crits don't give it | right; only a self-heal in the rotation can use it |
| Natural Grace (29187) | -5% threat a point on every school but physical | the same | right; a tank leaves it out |
| Nature's Swiftness, Healing Way, Purification, Healing Focus | emergency and self-heal talents | as the DPS and healing audits | right |
| Earth's Grasp, Improved Ghost Wolf, Earthbound, Elemental Reach, Eye of the Storm | utility, nothing a boss fight scores | not simmed or inert | right |

Not in Forever's tree (the client's skill line still lists them): Shield Specialization (16299 and up),
Improved Weapon Totems, Enhancing Totems, Weapon Mastery. A shaman has no taunt and no shield ability: it
can't be uncrushable (5% block, no Shield Block), so a level 63 boss lands 15% crushing blows on it.

### Health-based buttons

The shaman's emergency buttons are heals: Nature's Swiftness into Healing Wave, or Lesser Healing Wave, at low
health. The scoring replay can't follow a heal that reads health (`docs/tank_sim.md`), so the scoring rotations
leave them out and the UI's rotation has them. What that costs the Warden's score: nothing it would get from
its healers anyway, but a real Warden lives through some bursts the score counts as deaths. The template's
`selfHealMana` knob is a health-blind Lesser Healing Wave (its heals count as fixed heals in the replay).

## Tests

`sim/shaman/warden/forever_test.go` (level 60 Orc Warden, a 2.9 s axe and a shield, against the level 63 boss):

- `TestForeverWardenTankSheet`: the boss's swing table on the Warden: Defense, block (5% less the level gap plus
  Defense), Spirit Weapons' parry (none without it), block value (the shield's + Strength / 20 - 1), no block
  without a shield.
- `TestForeverWardenTalents`: each talent above on against off from the Level 60 build, measuring exactly what
  it moves: Anticipation's 6% dodge on the sheet, Toughness's 10% Stamina, Improved Reincarnation's 4% health,
  Thundering Strikes' 5% crit, Tidal Focus's 5% hit, Elemental Warding's three schools, Elemental Weapons'
  130.6 attack power, Mental Dexterity and Mental Quickness, Spirit Weapons' 1.302, Natural Grace's 15% off
  Earth Shock, Concussion, Reverberation's 1 s, Convection and Shamanistic Focus's costs, Improved Lightning
  Shield's 15%, Guardian Totems' 36 for 30, Ancestral Healing's 25% armor after a heal crit.
- `TestForeverWardenRockbiterThreat`: a white hit's and Stormstrike's threat (damage + 72 x 2.9, times 1.302),
  none on a dodge, Earth Shock's 2 x damage, an orb's 1 x; 0.7 with Windfury Weapon and no Rockbiter threat;
  1 without Spirit Weapons.
- `TestForeverWardenAbilities`: Earth Shock's 5 s cooldown with Reverberation 5/5, shared with Flame Shock, its
  GCD; Stormstrike's 8 s (the Nov 10 tuning; 6 s from hotfix 113380 before it) and 8% of base mana; Lightning Shield's 10 min, 3 orbs, 3.5 s apart, gone after the
  third; Stoneskin Totem's 210 mana, 5 min and 36 off a hit before armor.
- `TestForeverWardenAvoidedHits` (tank round 2): a missed, dodged or parried swing fires no Lightning Shield orb (a
  blocked one or a hit does); a missed, dodged or parried Stormstrike carries none of Rockbiter's threat;
  Stoneskin's 36 come off before armor and before a crit's x2 or a crushing blow's x1.5. The swings roll through
  the engine's own enemy swing table, forced to one outcome.
- `tools/rotopt/shaman_tank_test.go`, `TestShamanTankShockAboveStormstrike`: the template puts the shock above
  Stormstrike by default and below it with `shockFirst=0`.
- `TestForeverWardenImprovedStormstrikeReset`: a dodge or parry of the boss's swing resets Stormstrike, a
  block or a hit doesn't.
- `TestForeverWardenTank`: the bench's Balanced request against the level 63 boss with a stand-in healer: the
  boss attacks the Warden, which blocks and parries, the rotation's abilities go out, Lightning Shield and
  Stoneskin stay up, and the tank score replays every iteration.

Each test fails with the mechanic broken (the old flat Rockbiter threat on Stormstrike, the old 1.3, a shaman
parrying without the talent, Stoneskin after armor, Lightning Shield firing on any swing).

## Gear

`tools/database/forever_synthetic_items_shaman_tank.go`, ids 999201 to 999216, by standards 36 at the Warden row
of `orchestration/gear_targets.md` (from wowsims/sod's phase_4_enh_tank): hybrid budget, 845.1 weighted, Stamina
202, Strength 120, Intellect 131, Spirit 85, attack power 108, spell power 71, hit 55 rating (5.5%), crit 78
rating, Defense 26, block value 50 (all on the shield). A 2.9 s one-hand axe at Deathbringer's 56.4 DPS, the
0.72 x L(71) shield (2,681 armor, 50 block value, 8 Defense), the tank trinket (20 Defense, 40 block value), and
an on-use trinket: the Enhancement set's 335 attack power (990015) or the Warden's 1,665 armor (999216), both 20 s
every 2 min. Hands and legs carry The Spiritcaller's Rage (2 piece: 1% attack and casting speed). Enchants:
on head and legs Lesser Arcanum of Voracity Agility (Max threat, Balanced) or Tenacity (Max survival), Agility on the
neck, Superior Defense on the back (Agility,
1219587, at Max threat and Balanced), Greater Stats on the chest, Superior Deflection (1248665) on the wrists (Max
threat wore Superior Stamina, 1886, until 2026-11-12), Natural Power (Max threat) or Greater Agility on the hands
(Voracity Stamina, Necklace Strength and Superior Strength gloves until 2026-11-12, tank_fairness-20261112-2), Greater
Stamina on the feet
(929, +7, Max survival's; Greater Agility, 1887, at Max threat and Balanced), Enchant Shield - Greater Stamina
on the shield (enchant 1886, +9, spell 20017; tank round 3 moved it from 929, which is no longer a shield enchant in
the DB, so the CLI gave +7 and the web UI loaded no enchant: tank_gear-20261104-3), Crusader on the weapon.

## Rotation

`tools/rotopt/shaman_tank.go` (`-template shaman_tank`): the shield whenever it's down, the earth totem
(Stoneskin by default), the shock with a mana floor (Earth Shock by default, holding for Stormstrike's +20% as a
knob), Stormstrike on cooldown (below the shock since round 2: `shockFirst`), the fire and water totems (none by default; a fire totem's own damage is the
totem's threat, so only Fire Nova feeds ours), Lightning Bolt at Maelstrom Weapon stacks, a health-blind Lesser
Healing Wave, and the defensive trinket on a timer. The tank applies no raid debuff (a shaman has none that the
bench's raid gives: Demoralizing Shout and Thunder Clap stay the raid's), so `tankscore.OwnDebuffs` leaves the
Warden's raid as it is.
