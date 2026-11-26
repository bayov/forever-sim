# Phase 0 re-rank: druid and priest (2026-11-25)

Balance, Feral, Shadow and Smite after the sim batches db86a18dc, aed8856d1, efeb06b95 and 22d2cf952.
`RANK_ONLY=1 SLOTS=mainhand,offhand,ranged,trinket`, 180s. Request dps: verify.py, 180s ± 20s, 20,000 iterations, seed 1,
"before" is the master 34fe148 request on the current sim.

| Build | Slot change | List before -> after | Request before -> after |
|---|---|---|---|
| druid__balance | trinket Chained Essence of Eranikus -> Talisman of Ascendance | 813.66 -> 814.43 | 796.08 -> 802.06 |
| druid__balance-duty-ff | same set | | 762.38 -> 766.72 |
| druid__balance-wrath | same set | | 711.65 -> 713.65 |
| druid__balance-wrath-duty-ff | same set | | 684.24 -> 685.22 |
| druid__feral-cat | trinket Darkmoon Card: Maelstrom -> Heart of Wyrmthalak, Land Mine APL line | 1068.16 -> 1073.52 | 1058.41 -> 1060.26 |
| druid__feral-cat-duty-ff | same set and APL | | 1053.39 -> 1055.04 |
| priest__shadow | none | 865.71 -> 878.33 | 861.73 (unchanged) |
| priest__shadow-holy | none | | 819.34 (unchanged) |
| priest__smite | none | 686.81 -> 688.61 | 679.02 (unchanged) |

Each request matches its list's plain sim of the final set (Balance 802.1, Feral 1060.2, Shadow 861.8, Smite 679.2).
Weapons, off-hands, idols and wands keep their picks. The Balance run exits 1 on Swarming Idol (272430), unlisted in
phase0_content.json and not a pick (1.6 under Idol of the Moon), so it was left out, not shipped with ALLOW_UNLISTED.

## Faction copies and race rows

- Balance Horde (Tauren, SLOTS=finger): Watcher's Signet of Brilliance and Eye of Orgrimmar tie (806.56 vs 806.28).
  Eye of Orgrimmar kept, as master 6a6ea27 chose it. The copy changes only its trinket. Tauren request 793.81.
- Feral Horde (Tauren, SLOTS=shoulder,finger): Darkspear Pauldrons and Legionnaire's Band again, the trinket follows
  the list. List 1046.51, Tauren request 1036.78 (plain 1036.7).
- race_gear.py --write on druid__balance, druid__balance-wrath, druid__feral-cat: the Tauren and Skyborne Windshaper
  rows take the new trinket, nothing else moves. Shadow and Smite gear is unchanged, so their alliance copies stay.

## Set trap

10,000 iterations, set pieces swapped in at each bonus point (Feralheart 1667 and Virtuous 1675 have bonuses at 2, 3,
4 and 6 pieces). Balance and Smite already wear 6 pieces, not 1.

- Balance (6 pieces, 796.1): every subset of the 8 slots tried, 7 pieces 795.8, 8 pieces 779.8. Keeps 6.
- Feral (1 piece, 1059.1): best 4 pieces 1053.7, 6 pieces 1045.9, 8 pieces 1018.3. Keeps 1.
- Shadow (1 piece, 862.0): best 2 pieces 860.2 (the 2 piece bonus is resistances, so a single swap), 6 pieces 852.8.
- Smite (6 pieces, 679.1): the 7th piece (waist) 670.5. Keeps 6.

Nothing within 5 of the list, so no full search from a set start.

## Open items

- **Feral Dark Bomb:** the APL line cast the EZ-Thro Dark Bomb (260817), a stun explosive that does nothing to a level
  63 boss. It read exactly as no explosive (1058.41 both). The line now casts the Goblin Land Mine (4395), the
  request's fillerExplosive: 1058.89, +0.5 (same on a second seed), 2.4 mines a fight. Changed in
  `tools/rotopt/druid_feral.go`, both feral APLs and both requests. The APL note still says the cast restarts the swing
  timer, which 22d2cf952 changed (throws keep the swing).
- **Virtuous Cowl (Shadow):** included under standard 42. Its only doubt is availability (end of the Valthalak chain,
  quest giver not in our data), which 42 says is no reason to leave a good item out. It is the head pick on Shadow
  (+0.38 over Spellweaver's Turban, noise) and Smite (+7.5 over Virtuous Crown). phase0_content.json already has it as
  include, doubtful, with that note.
