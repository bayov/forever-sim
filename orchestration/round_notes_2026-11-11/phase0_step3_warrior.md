# Phase 0 step 3: warrior (2026-11-24, rerun on master d093d6e)

Arms (two-hand), Fury (two-hand, `warrior__fury_2h`) and Fury DW (`warrior__fury`), each a full `bis_phase0.sh`
search at 180s (uses at 90s), rerun per `briefs/phase0_rerun_1123.md` on 5e0a327 with the b1bbe43 content.
Arms and Fury were then ranked again (RANK_ONLY) on d093d6e (Heroic Strike and Cleave spend Flurry and Reckless
Swings charges). Fury DW ran its search on d093d6e. Lists: `ui/warrior/bis/level60_phase0_<spec>.bis.json` and
`.effects.json`.

## What changed from the earlier lists

- **Arms** (967.3 to 1017.7 by the search's score): Warblade of Caer Darrow replaces Destiny. Its Forever on-hit
  Frost proc (27.7 per melee hit, modelled and tested in 5e0a327) is worth +52 over Destiny. Furbolg Berserker's Boots
  (Forever item) replaces Battleboots of Heroism. Watcher's Signet of Ferocity replaces Legionnaire's Band after the
  first pick, Don Julio's Band, became infeasible in b1bbe43.
- **Fury** (991.8 to 993.0): Bracers of Subterfuge for Battleborn Armbraces, Stormshroud Gloves for Timbermaw
  Brawlers, Might of the Timbermaw for Valiant's Waistguard, Furbolg Berserker's Boots for Drudge Boots, Hand of
  Justice for Blackhand's Breadth. Item procs, weapon damage and the weapon skill fix moved these. The two-hand Fury
  build sits over the hit cap with Precision, so Battleborn's hit is worth less: it sims 5.4 behind Subterfuge
  (10000 iterations, about 6 standard errors).
- **Fury DW**: no earlier list (the first run never finished).
- d093d6e moved one pick: Arms takes Hermit's Pauldron over Truestrike Shoulders (+0.5, confirmed at 20000
  iterations). Fury kept every pick.

## PvP reputation check (b1bbe43)

Merged before the Fury and Fury DW searches started, so both ran on it. Arms started before it: its DRY_RUN pool is
unchanged (2646 items), and its one Exalted pick (Don Julio's Band) was reranked with SLOTS=finger. Each newly
included AV and AB item that fits a warrior was in the pool already.

## Tags added to `phase0_content.json`

- Warblade of Caer Darrow (quest, Scholomance), Demonshear (Stratholme drop), Furbolg Berserker's Boots and Spaulders
  of Containment (Forever items seen in game, source not in our data), Handcrafted Mastersmith Leggings (seen in game).
- Darkspear Epaulets and Darkspear Pauldrons: doubtful. Darkspear Raiders rep, level not in our data. They are item
  level 65 like the Raider gloves the file takes as Exalted, so the two tags disagree. Standards 42 keeps them. If they
  are Exalted, Fury DW loses about 2.6 at shoulder (Defiler's Spaulders is Exalted too).
- Battlegear of Heroism (set 1778, rows 226857 to 226872, head and chest doubtful), the Dungeon Set 2 flip by
  standards 42 (757a9e9), and source notes on Engraved Band, Timbermaw Brawlers and Might of the Timbermaw.

## Hand check

Every pick's flag matches `phase0_content.json`. No pick is infeasible, Alliance only or from a closed source. Negative
losses are infeasible Exalted items (Raider Gauntlets, Don Julio's Band, Warden's Signet, Forest Stalker's and
Berserker Bracers) or noise that failed the 20000 iteration confirm (Fury DW waist: Valiant's +0.3 ± 0.3).

Doubtful picks kept by standards 42: Lionheart Helm (plans 0.2% drop), Titanic Leggings (plans no source), Darkmoon
Card: Maelstrom, Bloodseeker (AV), Watcher's Signet (Arms, Fury DW), Hermit's Pauldron (Arms), Bracers of Subterfuge
(Fury, Naxx-era Insignia quest), Signet of the Soulless Rider (Fury, no known source), Darkspear Epaulets (Fury DW).

Bashguuder (MythicSim's Fury DW off hand) adds 0: its Puncture Armor sits in the minor armor category, which the
report's debuffs already fill.

## Compared with MythicSim (Forever provisional lists, Arms 2H and Fury DW)

- Same: Lionheart Helm, Mark of Fordring, Howler's Furs, Timbermaw Tunic, Arms Warblade and Valiant's, Fury DW
  Ironfoe, Battleborn and Blackhand's Breadth.
- Ours wins in the sim: Bloodseeker over Blackcrow (Arms, Blackcrow ties in Fury DW), Howler's over Cape of the Black
  Baron (5.5), Titanic over Devilsaur Leggings (31) and the plate legguards, Furbolg Berserker's Boots over
  Battleboots of Heroism (5.1), Darkmoon Maelstrom over Heart of Wyrmthalak (12.6 to 15), Timbermaw Brawlers over
  Devilsaur Gauntlets (4.8), Felstriker over Bashguuder.
- Infeasible or Alliance for us: Forest Stalker's Bracers and the plate legguards (WSG Exalted), Raider Gauntlets and
  Don Julio's Band (Exalted), Theramore Gauntlets and Epaulets (Alliance).
- Ties: Truestrike against Hermit's Pauldron (Arms), Darkspear Pauldrons against Epaulets (Fury DW).
- MythicSim has no Fury two-hand list.

## Battlegear of Heroism set test

Best Heroism set at each size against the list, 180s, 10000 iterations, randomSuffix dropped on swapped slots:

| Spec | 2pc | 4pc | 6pc |
|---|---|---|---|
| Arms | +1.0 (1.2 SE, noise) | -2.4 | -8.1 |
| Fury | -2.6 | -10.0 | -16.6 |
| Fury DW | -6.1 | -11.4 | -19.7 |

No set wins, so no "start" in `phase0_specs.json`. Arms wears the Heroism gauntlets alone.

## Calibration

Report build, 180s, 10000 iterations, `rotopt -eval`, Phase 0 list against the synthetic gear:

| Spec | Phase 0 | Synthetic | Gap |
|---|---|---|---|
| Arms | 1013.2 | 1080.0 | -66.8 (-6.2%) |
| Fury | 991.3 | 1130.3 | -139.0 (-12.3%) |
| Fury DW | 990.1 | 1109.7 | -119.6 (-10.8%) |

## Picks

### Arms

| Slot | Pick | Source | Flag |
|---|---|---|---|
| head | Lionheart Helm | Blacksmithing 300 | doubtful |
| neck | Mark of Fordring | Quest: In Dreams (Tirion Fordring chain), Plaguelands |  |
| shoulder | Hermit's Pauldron | Quest: Work Well Done | doubtful |
| back | Howler's Furs | Pix Xizzix, Reals |  |
| chest | Timbermaw Tunic | AH ~251g or Leatherworking 300 |  |
| wrist | Battleborn Armbraces | Warchief Rend Blackhand |  |
| hands | Gauntlets of Heroism | Quest (NPC not in our data yet) |  |
| waist | Valiant's Waistguard | Pix Xizzix, Reals |  |
| legs | Titanic Leggings | Blacksmithing 300 | doubtful |
| feet | Furbolg Berserker's Boots | Forever item, seen in game 2026-11-21 |  |
| finger | Engraved Band | Quest: Blackmaw Reclamation, Blackmaw Hold (as Eagle's Head Signet) |  |
| finger | Watcher's Signet of Ferocity | Watchers Revered | doubtful |
| trinket | Blackhand's Breadth | Quest reward, Blackrock Spire |  |
| trinket | Darkmoon Card: Maelstrom | Darkmoon Faire deck | doubtful |
| mainhand | Warblade of Caer Darrow | Quest: The Lich, Ras Frostwhisper, Scholomance |  |
| ranged | Bloodseeker | Alterac Valley quest | doubtful |

### Fury

| Slot | Pick | Source | Flag |
|---|---|---|---|
| head | Lionheart Helm | Blacksmithing 300 | doubtful |
| neck | Mark of Fordring | Quest: In Dreams (Tirion Fordring chain), Plaguelands |  |
| shoulder | Truestrike Shoulders | Pyroguard Emberseer |  |
| back | Howler's Furs | Pix Xizzix, Reals |  |
| chest | Timbermaw Tunic | AH ~251g or Leatherworking 300 |  |
| wrist | Bracers of Subterfuge | Quest: Epic Armaments of Battle - Friend of the Dawn | doubtful |
| hands | Stormshroud Gloves | Leatherworking 290 |  |
| waist | Might of the Timbermaw | Leatherworking 300 |  |
| legs | Titanic Leggings | Blacksmithing 300 | doubtful |
| feet | Furbolg Berserker's Boots | Forever item, seen in game 2026-11-21 |  |
| finger | Engraved Band | Quest: Blackmaw Reclamation, Blackmaw Hold (as Eagle's Head Signet) |  |
| finger | Signet of the Soulless Rider | No known source | doubtful |
| trinket | Hand of Justice | Drop: Emperor Dagran Thaurissan, Blackrock Depths |  |
| trinket | Darkmoon Card: Maelstrom | Darkmoon Faire deck | doubtful |
| mainhand | Destiny | Random world drop |  |
| ranged | Bloodseeker | Alterac Valley quest | doubtful |

### Fury DW

| Slot | Pick | Source | Flag |
|---|---|---|---|
| head | Lionheart Helm | Blacksmithing 300 | doubtful |
| neck | Mark of Fordring | Quest: In Dreams (Tirion Fordring chain), Plaguelands |  |
| shoulder | Darkspear Epaulets | Darkspear Raiders rep, level not in our data | doubtful |
| back | Howler's Furs | Pix Xizzix, Reals |  |
| chest | Timbermaw Tunic | AH ~251g or Leatherworking 300 |  |
| wrist | Battleborn Armbraces | Warchief Rend Blackhand |  |
| hands | Timbermaw Brawlers | Leatherworking 300 |  |
| waist | Might of the Timbermaw | Leatherworking 300 |  |
| legs | Titanic Leggings | Blacksmithing 300 | doubtful |
| feet | Furbolg Berserker's Boots | Forever item, seen in game 2026-11-21 |  |
| finger | Engraved Band | Quest: Blackmaw Reclamation, Blackmaw Hold (as Eagle's Head Signet) |  |
| finger | Watcher's Signet of Ferocity | Watchers Revered | doubtful |
| trinket | Blackhand's Breadth | Quest reward, Blackrock Spire |  |
| trinket | Darkmoon Card: Maelstrom | Darkmoon Faire deck | doubtful |
| mainhand | Ironfoe | Emperor Dagran Thaurissan |  |
| offhand | Felstriker | Warchief Rend Blackhand |  |
| ranged | Satyr's Bow | Zevrim Thornhoof |  |

## On the way

- Container fixes, not in the repo: a `jj` shim, protoc and protoc-gen-go. `ui/warrior/bis/` didn't exist, so the
  first Arms run failed at write.
- A container restart killed the Fury DW search and both rank runs. They were run again from the start.
