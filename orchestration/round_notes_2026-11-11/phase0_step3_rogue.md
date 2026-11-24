# Phase 0 step 3: rogue (2026-11-23)

Lists: `ui/rogue/bis/level60_phase0_rogue_{combat_swords,assassination,subtlety}.bis.json`. Full searches on
master 5e0a327 (item procs, weapon damage, weapon skill fixes), then the finger slot re-ranked (`RANK_ONLY=1
SLOTS=finger`) after b1bbe43 made the Exalted rewards infeasible. The pools were checked with `DRY_RUN` after
each merge: the AV and AB items b1bbe43 included were already in the rogue pools, so only the ring picks moved.

Combat Swords and Assassination are Undead (Horde), Subtlety Night Elf (Alliance), all Engineering and
Leatherworking, as in the report builds.

## Calibration (report build, 180s, 20000 iterations, seed 1)

| spec | synthetic | Phase 0 | gap | earlier Phase 0 list |
|---|---|---|---|---|
| Combat Swords | 948.4 | 843.2 | -105.2 (-11.1%) | 845.7 |
| Assassination | 870.3 | 821.1 | -49.2 (-5.7%) | 854.1 |
| Subtlety | 875.4 | 816.1 | -59.3 (-6.8%) | none (stopped for the fixes) |

Combat Swords' synthetic gear sits well above anything Phase 0 reaches. Its gap is twice the others', so its
synthetic set looks too strong for the report.

Set check (every Darkmantle subset of 2, 4 and 6 pieces against the picks): no subset beats the picks.
Combat Swords keeps 4 pieces (6 loses 5.7), Assassination and Subtlety wear all 6.

## Combat Swords

| slot | pick | source |
|---|---|---|
| head | Outlaw's Collar | Leatherworking craft, pattern 279302 has no known source |
| neck | Mark of Fordring | Quest: In Dreams, Plaguelands |
| shoulder | Darkspear Pauldrons | Creeg Bothunk, the Barrens, Darkspear Raiders Revered, Horde only |
| back | Cape of the Black Baron | Baron Rivendare, Stratholme |
| chest | Dawn Armor | AH ~350g or Leatherworking |
| wrist, hands, legs, feet | Darkmantle Bracers, Grips, Pants, Footpads | Quests at Deliana or Mokvar from Shadowcraft pieces, gold step unconfirmed |
| waist | Assassin's Waistguard | Pix Xizzix, Reals, price not in data |
| finger | Watcher's Signet of Ferocity, Cutthroat's Signet | Watchers Revered (no rep source found), Pix Xizzix 75 Reals |
| trinket | Chained Essence of Eranikus, Darkmoon Card: Maelstrom | Quest, Sunken Temple · Darkmoon deck |
| mainhand | Sword of Zeal | World drop |
| offhand | Greenhammer | Blacksmithing, plans have no known source |
| ranged | Deep Jungle Rifle | Thunderhoof (rare, zone not in data), seen in game 2026-11-22 |

Changed from the earlier list: Theramore Pauldrons (Alliance only, now out) became Darkspear Pauldrons, and the
rings Don Julio's Band and Warden's Signet (now infeasible) became Watcher's and Cutthroat's Signet.

## Assassination

Same as Combat Swords except: shoulder Darkmantle Spaulders, waist Darkmantle Belt (6 piece set), mainhand and
offhand Felstriker (Rend, Blackrock Spire, not unique), ranged Blackcrow (Vosh'gajin, Blackrock Spire).

Changed from the earlier list: Distracting Dagger fell to daggers +1, so Felstriker takes both hands. Dawn Armor,
Cape of the Black Baron and Darkmantle Grips and Pants replace the Stormshroud set and Howler's Furs. Rings as in
Combat Swords.

## Subtlety

Same as Assassination except ranged Deep Jungle Rifle. First list for this spec.

## Flags

- Doubtful, kept per standards 42: Greenhammer (Combat), Darkmoon Card: Maelstrom (all, card drops
  unconfirmed), Watcher's Signet (no rep source found).
- No tag but unsure: the Outlaw's Collar pattern has no known source, Thunderhoof's zone is unknown, Darkspear
  Pauldrons need Revered, Darkmantle and Waistguard prices are unconfirmed.
- Chained Essence of Eranikus leads the 180s trinket ranking but ranks 26th in the cooldown ranking (-16.8).
- Darkmantle Cap and Tunic have hidden stats, so the set's head and chest aren't in the pools.
- Don Julio's Band (Frostwolf) still ranks in the Night Elf list. It's infeasible, so no pick depends on it,
  but the faction filter misses it.
- No pick has hidden stats or an unmodelled effect, and no pick is from the other faction.

## Public list (timesaver.gg WoW Forever, checked 2026-11-20)

Back and chest match. Every other public pick is in our pools and ranks below ours (loss in Combat /
Assassination / Subtlety):

- Mask of the Unforgiven -3.5 / -4.3 / -4.2, Beads of Ogre Might -0.9 / -0.5 / -1.7, Truestrike Shoulders
  -6.1 / -9.9 / -11.8, Satyr's Bow -0.5 / +0.0 / -1.3: close, our picks win on hit and the Darkmantle set.
- Blackmist Armguards, Devilsaur Gauntlets and Leggings, Might of the Timbermaw, Shadowcraft Boots: -8 to -25,
  the Darkmantle 4 or 6 piece energy bonus beats them.
- Royal Seal of Eldre'Thalas and Smolderweb's Eye: about -10, Maelstrom and Eranikus are stronger on our fight.
- Persuader and Black Amnesty are out: their crafts need raid materials (`tools/rotopt/phase0.go`).
- Blackstone Ring and The Postmaster's Seal rank below the Signets.
