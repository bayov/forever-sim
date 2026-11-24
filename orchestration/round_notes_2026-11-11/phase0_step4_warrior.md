# Phase 0 step 4: warrior DPS (2026-11-24)

The ten warrior DPS report builds now wear their Phase 0 sets: Arms and its duties and Dragonkin wear
`level60_phase0_arms`, Fury 2H and its duties `level60_phase0_fury`, Fury DW and its duties `level60_phase0_fury_dw`.
Requests come from mksettings with `-gear` set to those. Protection is another session's.

## Per build

verify.py, 180s ± 20s, 20000 iterations, seed 1. Race Orc for all.

| Build | Phase 0 | Synthetic it replaces |
|---|---|---|
| Arms | 1014.81 | 1078.7 |
| Arms, Battle Shout duty | 1012.31 | 1074.7 |
| Arms, Sunder duty | 964.87 | 1032.2 |
| Arms, Dragonkin | 1042.04 | 1135.7 |
| Fury 2H | 990.34 | 1112.4 |
| Fury 2H, Battle Shout duty | 986.58 | 1109.1 |
| Fury 2H, Sunder duty | 956.12 | 1074.4 |
| Fury DW | 1007.12 | 1102.1 |
| Fury DW, Sunder duty | 973.03 | 1063.1 |
| Fury DW, Battle Shout duty | 1001.30 | 1097.2 |

The requests match their lists: Arms eval 1015.5, Fury 2H list 993.0 (step 3 calibration 991.3), Fury DW eval 1007.0.

## List changes

- **Hands:** Raider Gauntlets (272095) now rank first. Arms: replace Gauntlets of Heroism, +2.3 (list 1019.79). Fury DW:
  replace Timbermaw Brawlers, +1.8 (list 998.29).
- **Fury DW weapons and the imbue:** the old Fury DW request had a Sharpening Stone on the main hand. The sim drops a
  stone that doesn't fit the weapon, so Ironfoe (a mace) ran with no imbue, and the list ranked it so. The request now
  has Dense Weightstone on the main hand (mksettings picks it by weapon type), Dense Sharpening Stone on the off hand.
  RANK_ONLY SLOTS=mainhand,offhand on the corrected request keeps Ironfoe and Felstriker. The set scores 1012.40,
  +14.11 over the old list (eval 1007.0 at 180s).
- Effects files refreshed for Arms and Fury DW.

## Hit, talents and rotations

Melee hit on the Phase 0 sets at Orc: Arms 8.3%, Fury 2H 8.0%, Fury DW 8.0%. The synthetic Fury DW set was at 12% (its
hit+2 step), the other two at 8%.

- **Talents:** the Fury DW search at the new hit (-combined 120,180,300) keeps `30313113002-050530055120100231-1`. Arms
  and Fury 2H hit didn't move, so no talent change, and no RANK_ONLY rerun for talents.
- **Fury DW, Execute first:** the knob search at 8% hit found executeFirst=1. Confirmed at 20000 iterations:

  | Preset | 120s | 180s | 300s |
  |---|---|---|---|
  | Fury DW | +5.6 | +4.1 | +4.5 |
  | Sunder duty | +4.1 | +3.1 | +3.8 |
  | Battle Shout duty | +5.6 | +4.3 | +4.2 |

  On in `tools/rotopt/warrior_presets.sh` for the three Fury DW lines, APLs regenerated. recklessSwingsRage=70 also
  came up but is noise, not kept.
- Execute first still loses on Fury 2H (-3.4, -3.4, -3.5) and Arms (+0.1, -3.1, -6.6), so it stays off there. Fury 2H's
  loss was 22.7 on the synthetic set, the metas now say 3.4.

## Races

- Alliance sets swap the Horde rewards:
  - `level60_phase0_arms_alliance`: Gauntlets of Heroism for Raider Gauntlets. 1013.88 at Human.
  - `level60_phase0_fury_dw_alliance`: Theramore Epaulets (271956) for Darkspear Epaulets, Gauntlets of Heroism for
    Raider Gauntlets. 998.77 at Human.
  - Lists in `phase0_step4_warrior/`. Fury 2H has no Horde-only pick, so no Alliance set.
- race_gear: Human, Dwarf, Night Elf, Gnome and High Order Arms and Fury DW rows wear the Alliance sets. Fury 2H Dwarf
  gets the bonus-stats trade, +5.50 ± 0.10 (1% hit re-spent). The old synthetic hit-1 Dwarf gear is gone.

## Problems

1. race_gear.py returns early for a faction set, so an Alliance Dwarf gets no hit trade. Arms Dwarf sits at 9.3% hit,
   probably about 5 DPS left (Fury 2H's trade was worth 5.5).
2. The Phase 0 search scores every weapon candidate with the request's imbue. A dual-wield list ranking maces against
   swords misjudges them, since a stone that doesn't fit is dropped. Any spec that dual-wields mixed weapon types has it.
3. Dragonkin drops more than Arms (-94 vs -64), since Spearing Strike scales with weapon damage. It keeps the Arms set.
