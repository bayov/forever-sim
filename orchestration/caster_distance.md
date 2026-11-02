# Caster distance (standards 33)

User ruling 2026-11-02 (fairness-20261101-1, "Range the rotation needs"): each ranged spec stands at the
longest range all of its rotation's spells reach. Applied 2026-11-02 to the Level 60 report's casters.

How the distance is taken:
- The spells are the ones the build's report APL casts (`tools/dps_report/req/<id>.json`, the same as
  `ui/<class>/apls/<apl>.apl.json`): every `castSpell` / `channelSpell` / `multidot` action, conditions
  left out. Items (trinkets, potions, mana gems, runes) and self spells (Self Only range: buffs, totems,
  Evocation, Life Tap, forms) don't limit it.
- Each spell's range is its client row: SpellMisc `RangeIndex` -> SpellRange `RangeMax_0` (hostile), client
  1.60.1.70707 (`~/.cache/wow-data/wago/1.60.1.70707+hf`, the 70707 tables with the hotfixes).
- Range talents are SpellEffect aura 107 (flat) or 108 (percent) with misc value 5, applied when the
  effect's class mask covers the spell's SpellClassOptions mask and the build takes the talent. Rank values
  are the trait curves' (forever-wiki/classes/*.md): Arctic Reach 10% a rank, Shadow Reach 10% a rank,
  Nature's Reach 10% a rank, Improved Fire Blast 8 / 15 yd, Grim Reach and Destructive Reach 10% a rank,
  Elemental Reach 3 / 6 yd (Lightning Bolt, Chain Lightning, Lava Burst; no longer the shocks), Holy Reach
  10% a rank, Flame Throwing 3 / 6 yd, Arcane Geometry 3 / 6 yd.
- A spell that hits around the caster reaches its radius (SpellRadius): Arcane Explosion 10 yd.
- The distance is the minimum of those ranges. The sim checks a spell's range only in leveling
  (`sim/core/spell.go` `MaxRange`), so in a raid sim a spell out of range still lands; the distance has to
  be right in the request.

| Build (report ids) | Spells the rotation casts, range (yd) | Range talents taken | Distance (was) |
|---|---|---|---|
| Frostfire (`mage__frostfire`) | **Fire Blast 20**, Scorch 30, Arcane Missiles 30, Ice Lance 33, wand 30, Frostfire Bolt 35, Pyroblast 35 | Arctic Reach 2/2 (Ice Lance; not Fire Blast) | **20** (25) |
| Fire (`mage__fire`) | **Scorch 30**, **Arcane Missiles 30**, **wand 30**, Fire Blast 35, Fireball 35, Pyroblast 35 | Improved Fire Blast 2/2 (Fire Blast 20 -> 35) | **30** (25) |
| Frost (`mage__frost`) | **Arcane Missiles 30**, **Arcane Blast 30**, **wand 30**, Frostbolt 33, Ice Lance 33 | Arctic Reach 1/2 (Frostbolt, Ice Lance) | **30** (25) |
| Arcane (`mage__arcane`) | **Fire Blast 20** (finisher), Frostbolt 30, Arcane Blast 30, Arcane Missiles 30, Arcane Wrath 30, wand 30; Arcane Explosion 10 (finisher, dropped) | none | **20** (25) |
| Warlocks, all 12 (`warlock__*`, duty variants included) | Shadow Bolt, Immolate, Corruption, Bane of Agony, Bane of Doom, Searing Pain, Death Coil, Conflagrate, Incinerate, Curse of the Elements / Recklessness: all **30**; Soul Link 100 | none (no Grim Reach, no Destructive Reach) | **30** (25) |
| Shadow, Shadow + Holy (`priest__shadow`, `priest__shadow-holy`) | **Mind Flay 22**, Shadow Word: Pain 33, Mind Blast 33, Shadow Word: Death 33, Devouring Plague 33, wand 30 | Shadow Reach 1/2 (Mind Flay 20 -> 22) | **22** (25) |
| Smite (`priest__smite`) | Smite, Shadow Word: Pain, Shadow Word: Death, Devouring Plague, wand: all **30**; Power Infusion (on self) | none (no Holy Reach) | **30** (25) |
| Elemental, both (`shaman__elemental-level60`, `-mana-tide`) | **Earth Shock 20**, **Flame Shock 20**, Lightning Bolt 30, Lava Burst 30 | none (Elemental Reach no longer reaches the shocks) | **20** (20, unchanged) |
| Balance, all 4 (`druid__balance`, `-wrath`, both `-duty-ff`) | Wrath, Starfire, Moonfire, Insect Swarm, Faerie Fire: all **36** | Nature's Reach 2/2 (30 -> 36) | **36** (30) |

Bold: the spells that set the distance.

Notes:
- Arcane: its finisher cast Fire Blast and Arcane Explosion when a Frostbolt would not land before the end.
  Arcane Explosion reaches only 10 yd around the mage, so with it the rule would put Arcane at 10 yd, inside
  the Goblin Sapper Charge's reach. The finisher drops it instead (`tools/rotopt/mage_arcane.go`), and Arcane
  stands at Fire Blast's 20. At 180 sec, seed 1, 20,000 iterations: 939.8 at 20 yd with Arcane Explosion
  (which the sim lets land from any distance), 937.5 at 20 without it, 941.6 at 10 with it (no Sapper),
  935.4 at 30 without it, 929.1 at 30 with neither Fire Blast nor Arcane Explosion. Its High Order race
  rotation (`req/races/mage__arcane.SkyborneHighOrder.json`) drops the line too.
- Fire: the finisher pad was the Fireball's flight time from 25 yd (cast plus 1.04 sec). From 30 it is 1.25
  sec: +0.6 / +0.4 / +0.3 at 120 / 180 / 300 sec at 30 yd, so `finishPad=1.25` (`mage_presets.sh`, the
  template default, and the High Order race rotation).
- Frost's pad (cast plus 1 sec) is the Frostbolt's flight from 28 yd; 1.25 at 30 yd loses 0.9 / 0.6 / 0.3,
  so it stays. Arcane's (0.75) is 0.71 sec of flight from 20 yd and Frostfire's (1.5, searched) is above
  its 0.83, so neither moved.
- Standards 33's "24 with Shadow Reach 3/3" is from the old tree: Shadow Reach has 2 ranks under Forever
  (10% a rank) and the Shadow builds take 1, so 22 yd.
- The wand (Shoot, 5019) is 30 yd; a Fire or Frost build without it would still stand at 30 for Scorch or
  Arcane Missiles.
- Weaving hunters (user ruling 2026-11-02, standards 33): they stand where weaving does best, so their
  distance is searched. Their shots reach 35 yd (Auto Shot 8 to 35 in SpellRange, Arcane Shot, Aimed Shot,
  Multi-Shot, Serpent Sting; Sniper Shot 45), Raptor Strike 5. The sim holds Auto Shot and the shots to
  Forever's deadzone, 11.67 yd (Auto Shot's 8 past both combat reaches, `core.ForeverMinRangedAttackDistance`),
  and every ranged rotation walks to 12 yd 6 sec before the pull and weaves between 12 and melee, so the
  request's distance is only where the hunter and its pet start. Swept 12, 13, 15, 18, 21, 25, 30, 35 yd at
  120 / 180 / 300 sec (seed 1, 20,000 iterations), then 12 against 25 on seed 1000003 and on the pet
  variants. Sum of the three lengths, 12 yd against 25:
  - Beast Mastery +4.3 / +3.7 (seeds 1 / 1000003), Wolf +3.0 / +3.4, Tallstrider +3.7 / +4.0; 12 to 18 tie,
    21 is +2.5, 30 -0.9, 35 -3.6. The pet runs in from the hunter's side. Beast Mastery and its variants
    start at 12: 1045.4 -> 1046.9, 1001.0 -> 1002.1, 987.6 -> 988.9 at 180 sec (120 s 1048.3 -> 1050.4,
    300 s 1032.9 -> 1033.7 for the base).
  - Marksmanship +1.1 / +1.6, Wolf +1.3 / +1.2, Tallstrider +1.7 / +0.3: inside 2 standard errors, so they
    stay at 25.
  - Lone Wolf (no pet): bit-identical from 12 to 35, stays at 25.
  - Survival is melee and stays at 5.

- Balance Wrath: from 36 yd a Wrath flies 1.8 sec, longer than its 1.5 sec cast, and the build loses 7 to 8
  DPS (29 yd 855.8, 31 yd 848.9, 36 yd 848.0 at 180 sec: the drop is at 30 yd, where the flight passes the
  cast). Its end Moonfire slack was the Wrath's flight from 30 yd (1.5 sec); 1.8 is +0.5 / +0.4 / +0.2, so
  `endMoonfireSlack=1.8` (`druid_presets.sh`, both Wrath rotations). The Starfire build barely moves (+1.2).
- Race check (standards 29; only the Wrath builds moved more than 2 at 180 sec; the duty variant follows its
  base): `druid__balance-wrath` at 180 ± 20 sec, 20,000 iterations, seed 1, each race's own rotation where it
  has one. Before (30 yd): High Order 855.7, Windshaper 855.7 (the preset), Night Elf 853.2, Tauren 844.5.
  After (36 yd): Night Elf 852.4, High Order 848.3, Windshaper 848.3, Tauren 844.2. The hasted Skyborne lose
  most, since haste shortens the cast but not the flight.
- Race re-pick (standards 29 and 16): Night Elf over Windshaper is +6.3 / +4.1 / +3.4 at 120 / 180 / 300 sec,
  the same on seed 1000003, so both Wrath builds are Night Elf now. Windshaper stays the best Horde race
  (Tauren 4.1 behind it). Knob searches at Night Elf and at Windshaper (180 sec, Wrath kept as the filler; with
  the filler free the search weaves Starfire, +33, which is the Balance build) kept every knob, so no race
  rotation. As Night Elf the rotation holds Elune's Light for the trinket like the Balance build
  (`elune=1`, +0.9 / +0.6 / +1.1, and nothing changes for races without it). At 180 sec: Balance Wrath
  853.0 (848.4 as Windshaper), its Faerie Fire duty 820.8 (816.4), at 120 / 300 sec 849.8 / 850.8 and
  815.7 / 821.1.
- The High Order race rotations, re-measured at 180 sec after their edits: Arcane +4.9 (937.2 against 932.3
  with the preset's rotation; was +5.0), Fire +13.7 (1047.2 against 1033.5; was +14.0). Both still beat the
  preset's rotation at High Order by far more than 2 SE, so they stay; the same rotations at the preset races
  are -0.3 (Gnome) and 0.0 (Human).

DPS, seed 1, 20,000 iterations, 180 ± 20 sec fights at each length (old -> new):

| Build | Distance | 120 s | 180 s | 300 s |
|---|---|---|---|---|
| druid__balance | 30 -> 36 | 893.3 -> 894.7 (+1.4) | 891.1 -> 892.3 (+1.2) | 882.2 -> 883.4 (+1.3) |
| druid__balance-duty-ff | 30 -> 36 | 857.5 -> 858.5 (+1.1) | 857.7 -> 858.6 (+1.0) | 851.6 -> 852.4 (+0.9) |
| druid__balance-wrath | 30 -> 36 | 850.4 -> 842.6 (-7.8) | 855.7 -> 848.4 (-7.3) | 853.2 -> 846.3 (-6.9) |
| druid__balance-wrath-duty-ff | 30 -> 36 | 816.8 -> 809.7 (-7.1) | 823.2 -> 816.4 (-6.8) | 822.9 -> 816.4 (-6.5) |
| mage__arcane | 25 -> 20 | 947.3 -> 944.8 (-2.5) | 939.0 -> 937.5 (-1.5) | 931.1 -> 930.6 (-0.5) |
| mage__fire | 25 -> 30 | 1065.6 -> 1065.4 (-0.2) | 1056.4 -> 1057.1 (+0.7) | 1039.9 -> 1040.1 (+0.2) |
| mage__frost | 25 -> 30 | 945.3 -> 944.5 (-0.8) | 944.9 -> 944.7 (-0.3) | 937.6 -> 937.7 (+0.1) |
| mage__frostfire | 25 -> 20 | 1078.6 -> 1076.9 (-1.7) | 1076.5 -> 1076.0 (-0.5) | 1055.7 -> 1054.2 (-1.5) |
| priest__shadow | 25 -> 22 | 944.7 -> 944.7 (0.0) | 948.7 -> 948.7 (0.0) | 946.0 -> 946.0 (0.0) |
| priest__shadow-holy | 25 -> 22 | 902.3 -> 902.3 (0.0) | 906.6 -> 906.6 (0.0) | 873.6 -> 873.7 (+0.1) |
| priest__smite | 25 -> 30 | 777.7 -> 777.7 (0.0) | 761.4 -> 761.4 (0.0) | 686.6 -> 686.6 (0.0) |
| warlock__affliction | 25 -> 30 | 1016.3 -> 1015.2 (-1.1) | 1012.1 -> 1012.6 (+0.5) | 994.9 -> 996.2 (+1.2) |
| warlock__affliction-duty-curse | 25 -> 30 | 1000.5 -> 999.9 (-0.7) | 1003.9 -> 1003.7 (-0.2) | 988.5 -> 989.5 (+1.0) |
| warlock__affliction-duty-recklessness | 25 -> 30 | 999.1 -> 998.3 (-0.8) | 998.2 -> 997.4 (-0.9) | 982.8 -> 983.1 (+0.3) |
| warlock__demonology (its own set, standards 34) | 25 -> 30 | 1059.4 -> 1058.2 (-1.3) | 1052.7 -> 1053.3 (+0.6) | 1027.5 -> 1029.0 (+1.5) |
| warlock__demonology-duty-curse | 25 -> 30 | 1042.8 -> 1042.7 (-0.1) | 1043.3 -> 1043.8 (+0.6) | 1020.3 -> 1022.0 (+1.7) |
| warlock__demonology-duty-recklessness | 25 -> 30 | 1041.5 -> 1041.0 (-0.5) | 1037.0 -> 1037.4 (+0.4) | 1014.0 -> 1015.2 (+1.2) |
| warlock__destruction | 25 -> 30 | 997.5 -> 997.3 (-0.1) | 995.3 -> 995.9 (+0.6) | 978.2 -> 979.4 (+1.2) |
| warlock__destruction-duty-curse | 25 -> 30 | 981.9 -> 981.3 (-0.5) | 986.8 -> 987.2 (+0.4) | 971.8 -> 972.7 (+0.8) |
| warlock__destruction-duty-recklessness | 25 -> 30 | 980.3 -> 979.8 (-0.5) | 981.6 -> 981.7 (+0.1) | 965.7 -> 966.6 (+0.9) |
| warlock__destruction_fire | 25 -> 30 | 985.2 -> 983.3 (-1.9) | 977.7 -> 976.5 (-1.2) | 956.3 -> 955.8 (-0.5) |
| warlock__destruction_fire-duty-curse | 25 -> 30 | 970.2 -> 968.3 (-1.9) | 968.9 -> 967.5 (-1.3) | 949.8 -> 949.1 (-0.7) |
| warlock__destruction_fire-duty-recklessness | 25 -> 30 | 967.4 -> 965.6 (-1.8) | 962.2 -> 961.2 (-1.0) | 943.2 -> 942.8 (-0.5) |

The Shadow builds are bit-identical (no spell of theirs flies); Elemental is unchanged at 20.

Where it is set: `tools/rotopt/mksettings/spec_*.go` (the report requests), `ui/<class>/presets.ts`
`OtherDefaults` (each page's default, the report default build's) and, where a page's builds differ, the
build's own `distanceFromTarget` (mage, Shadow priest page's Smite).
