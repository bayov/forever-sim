# Phase 0 tank and healer round: shaman_restoration

Branch `phase0-tankheal-shaman_restoration`. List `ui/restoration_shaman/bis/level60_phase0_shaman_restoration.bis.json`,
gear set `ui/restoration_shaman/gear_sets/level60_phase0_shaman_restoration.gear.json`, worn by both healbench reqs.

## Score

healbench `-carry -iters 1000` on the seven bosses, tank req `-role tank`, raid req `-role raid`, the score is the
mean of the two role means.

| Set | Tank | Raid | Score |
|---|---|---|---|
| Phase 0 | 120.1 | 134.8 | 127.5 |
| Synthetic (origin/master reqs) | 127.6 | 133.1 | 130.3 |
| Gap | -7.5 | +1.7 | -2.9 |

The synthetic set gives 130.3 here, against scoreRef 131.2. The search's own figure for the list is 127.6.
The loss is all in the tank role. The Phase 0 set has less mp5 and healing than the synthetic, so the tank
healer runs out sooner (Chromaggus OOM at 129s vs 147s).

| Boss | P0 tank | P0 raid | Syn tank | Syn raid |
|---|---|---|---|---|
| broodlord | 92.7 | 89.1 | 96.0 | 88.1 |
| chromaggus | 92.4 | 113.2 | 100.5 | 105.2 |
| golemagg | 183.1 | 243.8 | 197.5 | 246.9 |
| magmadar | 198.0 | 175.0 | 209.1 | 177.5 |
| onyxia | 132.8 | 95.4 | 139.5 | 90.3 |
| ragnaros | 55.8 | 121.0 | 57.8 | 117.7 |
| vaelastrasz | 85.9 | 106.2 | 92.8 | 105.7 |

## Picks

| Slot | Pick | Source | Flag |
|---|---|---|---|
| head | Crown of The Five Thunders (227013) | Mokvar quest, end of the Valthalak chain | doubtful |
| neck | Orb of the Darkmoon (19426) | Darkmoon Faire, 1200 tickets | |
| shoulder | Pauldrons of The Five Thunders (227003) | Mokvar quest, Anthion chain | |
| back | Battle Healer's Cloak (19526) | Warsong Outriders Honored | Horde |
| chest | Chain of The Five Thunders (227024) | Mokvar quest, end of the Valthalak chain | doubtful |
| wrist | Bracers of The Five Thunders (227009) | Mokvar quest, Bindings of Elements + 20g | |
| hands | Raider Handwraps (272097) | Darkspear Raiders Revered | Horde |
| waist | Sash of The Five Thunders (227010) | Mokvar quest | |
| legs | Ironfeather Leggings (252486) | AH or Leatherworking | |
| feet | Slippers of The Five Thunders (227007) | Mokvar quest, Anthion chain | |
| finger | Watcher's Signet of Brilliance (275977) | Watchers Revered, no rep source found | doubtful |
| finger | Band of Piety (22681) | Argent Dawn armaments quest | doubtful |
| trinket | Burst of Knowledge (11832) | Ambassador Flamelash, BRD | |
| trinket | Smokey's Lighter (13171) | Strat quest | |
| mainhand | Masterwork Crimson Hammer (276404) | Crimson Hammersmith, Strat | |
| offhand | Evergreen Shield (279262) | Blacksmithing, plans have no source | doubtful |
| ranged | Totem of Urgency (279249) | Enchanting, formula has no source | doubtful |

Enchants are the list's. Enchant Weapon - Healing Power is out (Molten Core), so the hammer took Mighty Intellect (2568).

Process: full search, then RANK_ONLY on weapons, trinkets and head from the ranked tops at 2000 iterations
(the full search made few swaps). No swap held at 5000 iterations: Weakness Analyzer +0.2, Coif +0.1 and
Lorespinner +0.1, all inside noise. The re-rank's set score is 127.6, down 1.7 from the old list, mostly the
weapon enchant going out. The trinket and mainhand tops are all within 0.5.

Tags added to phase0_content.json: Raider Handguards, Totem of Urgency (doubtful), Charstone Dirk, and the tag
agent's batch (Battle Healer's Cloak, Darkspear Raider's Cloak, Tidal Totem, Barkskin Baubles, High Clerist's
Seal and others). Enchants added: Light Armor Kit, Shield - Superior Spirit. The Premier PvP tags were dropped
after the gear.go fix.

## Set trap

The set wears 6 Five Thunders pieces (head, shoulder, chest, wrist, waist, feet), and 6pc is the top bonus
(+23 spell power). The trap does not bite. The Premier Linkmail PvP set leaked in once (4pc, +1.4), fixed in
gear.go by the parent.

## Hand check

- Five Thunders head and chest need the end of the Valthalak chain, step gold unknown. The 95g Mokvar buy-back
  after the quest is in the notes.
- Watcher's Signet: no Watchers rep source in our data. Next best ring is close.
- Evergreen Shield and Totem of Urgency: the plans and formula have no known source.
- Public list: none reachable. Icy Veins 403, wowisclassic 404, Wowhead served the guide index, not the gear
  page (no bot-check workaround tried).

## Rotation and mana

No change. The list's set has 5752 mana and 920s to OOM in the search. On healbench the tank role goes OOM
first on Chromaggus (129s), Onyxia (252s) and Vaelastrasz (255s), a bit earlier than the synthetic set. That is
the gear's mp5, not the rotation.

## Faction row

FACTION_ROW
