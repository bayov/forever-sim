# Phase 0 tank: bear druid (2026-11-25)

Balanced, Max threat and Max survival, on the final sim code (origin/phase0-tankheal 1a1c5f0ba, a1904ac61 and
the five sim batches). Scored by the tank score, six bosses, as bis_phase0.sh does. Lists in
`ui/feral_tank_druid/bis/level60_phase0_druid_bear_<preset>.bis.json`, gear sets in `gear_sets/`, both in the
preset's `tools/tankbench/req/druid_bear_<preset>.json`.

## How the lists were made

- Balanced and Max threat: full search on the old code, then RANK_ONLY over all slots on the threat fixes,
  items, procs and caster batches (MIX_DURS=1, before the fight length fix).
- Max survival: full search on the final code.
- All three: RANK_ONLY over the armor slots after the Premier PvP drop (no Premier item was a pick or in a top 5).
- Rotation: re-searched on the threat fixes (docs/tank_druid.md, "Re-searched on the tank threat fixes").
  Max threat Enrage under 80 Rage (+3), Max survival Lacerate refresh at 6s (+19). Talents kept.
- Rotation and talents re-checked on the Phase 0 gear (knob and talent search, paired checks): nothing
  beats noise (best: Balanced Frenzied Regeneration at any Rage +7 ± 11, Max survival Roar at 2s +8 ± 9). No change.
- Lacerate first-hit fix (95962a7d3): Max survival now refreshes Lacerate under 10s, not 6 (+69 ± 11, +88 ± 12
  on seeds 1 and 2; 10 is past rotopt's knob range of 1 to 9). Balanced and Max threat keep 4. RANK_ONLY over all
  slots on the fix: no pick changed.
- No Demoralizing Roar for Balanced and Max threat (audit tank_rotations-20261125-1, roarRefresh 0, 09709d8):
  paired +46 ± 6 and +52 ± 6 (Balanced), +143 ± 3 and +150 ± 3 (Max threat) on seeds 1 and 2. Max survival keeps
  Roar (-120 ± 22 without).
- Re-rank on the no-Roar APL and the prices (fc39e50): Balanced over all slots, no pick changed. Max threat over all
  slots: Hammer of Thundering Hooves for Manual Crowd Pummeler (+21.3 ± 6.4 at 5000) and Dawn Armor for Timbermaw
  Tunic (+15.8 ± 3.6). Feralheart Girdle (+7.6 ± 3.4) didn't add on top of Dawn Armor. Max survival chest, trinket
  and ranged: no change. Tauren row 3494.0 against 3683.5.
- phase0_check.py flags that stand, each item failed its 5000-iteration confirmation over the pick: Balanced neck
  Horn of Ve'han (-0.6 ± 8.5), shoulder Truestrike Shoulders (-3.9 ± 8.9), chest Dawn Armor (+7.0 ± 8.5); Max
  survival chest Shadowcraft Tunic (-50.0 ± 14.7), trinket Adaptive Combat Assistant (-6.5 ± 21.8), ranged Idol of
  Brutality (-17.7 ± 21.6). The flags compare the 1000-iteration ranking.
- Balanced re-searched from the Max threat list (audit 849024cbb: the Max threat set beat the Balanced list by 193 at
  the Balanced weight): Horn of Ve'han for Beads of Ogre Might (+12.2 ± 5.3), the rest kept from Max threat. 3767.9
  against 3683.5 for the old list re-ranked on the same code. Theramore Pauldrons are out, so Balanced has no faction
  pick now. The Tauren row is the same set at Tauren, 3516.8. Flag that stands: ranged Idol of Brutality (-22.0 ± 10.2
  at 5000).

## Tags added

- phase0_content.json: 32 items. Out: Grovekeeper Gauntlets, Leggings, Spaulders, Treads (Malleable Essence
  of Nature, from Hyjal and Barrow Deeps), Warbear Harness (pattern from Lethon, Emeriss or Meilosh for an
  Earthfury Helmet, which looks like a data oddity, Woolies 15065 is left open). Doubtful: Idol of the Ursine
  Twins (formula 279300 has no known source), Vosh'gajin's Strand, Dinosaur Hide Belt, Predator's Shroud,
  Yeti Hide Bracers. The rest open, dungeon drops and crafts.
- phase0_enchants.json: 16 enchants, all open (trainer, vendor, Merchant's Favor, Dire Maul librams).
  Superior Deflection (+9 Defense) is 240 Merchant's Favor at Honored.

## Set trap

`rotopt -eval` at 10000 iterations, seeds 1 to 4, sets swapped into the list:
- Balanced: Devilsaur 2pc +52 (the trap). A full search from it kept Devilsaur Gauntlets and Leggings,
  +48 on the mean of four seeds. List replaced. Feralheart 2 to 8pc, Cadaverous 2 to 5pc, Shadowcraft 3pc,
  Wildheart and Stormshroud all lose.
- Max threat: already wears Devilsaur 2pc. Feralheart 2pc -44, everything else further down. List kept.
- Max survival: Feralheart 6pc within noise of the 3pc list. A full search from it kept six Feralheart
  pieces, Dawn Armor and Idol of the Ursine Twins, +32 on four seeds. List replaced.
- Feralheart Armor 226804 is a stub with no stats, so the chest is Dawn Armor.

## Hand check

- Defense 315 (Max threat), 330 (Balanced), 370 (Max survival). Crit immunity is at 315 with talents, all
  three take 0% crits.
- Doubtful picks: Watcher's and Tracker's Signet of Vigilance in all three (Watchers rep, no source found),
  Drudge Boots in all three, Bracers of Subterfuge (Balanced), Idol of the Ursine Twins (Max threat, Max
  survival), Feralheart Faceguard and Darkmoon Card: Heroism (Max survival).
- Manual Crowd Pummeler (Balanced, Max threat): 3 charges that don't come back, modelled as the sim does.
- No hidden-stat or unmodelled-effect item is a pick.
- Public list: Icy Veins refused the fetch (403), so no comparison.

## Faction row

- Balanced wore Theramore Pauldrons (Alliance only) until the 2026-11-26 re-search. Horde row (Tauren, RANK_ONLY SLOTS=shoulder):
  Truestrike Shoulders, 3494.0 against 3683.5 (no-Roar APL). `level60_phase0_druid_bear_balanced_horde.bis.json` and
  `gear_sets/level60_phase0_druid_bear_balanced_horde.gear.json`.
- Max threat and Max survival: no faction pick, skipped. (Max survival's top mainhand is The Unstoppable
  Force, infeasible, so not picked.)

## Score

tankbench, 1000 iterations, six bosses, the preset's own column. Synthetic is the same request (same APL)
with origin/master's gear.

| Preset | Phase 0 | Synthetic | Gap | TPS Phase 0 / synthetic |
|---|---|---|---|---|
| Balanced | 3771.5 | 3701.2 | +70.3 | 2822 / 2519 |
| Max threat | 5606.7 | 4959.7 | +647.0 | 2931 / 2645 |
| Max survival | 1048.0 | 1618.3 | -570.3 | 1905 / 1972 |

Max survival loses 570 because the synthetic set has 10159 armor and 437 defense against 9473 and 370.
Chromaggus drives most of it: -2127 against -1470.

## Picks

### Balanced (Night Elf)

| Slot | Pick | Source | Flag |
|---|---|---|---|
| head | Defias Enforcer's Mask | Trash, likely Alcaz Prison like its set (zone 16620, no NPC known) |  |
| neck | Horn of Ve'han | Quest: Work Well Done | avail doubtful |
| shoulder | Truestrike Shoulders | Pyroguard Emberseer · Blackrock Spire · 15g per point of score |  |
| back | Howler's Furs | Pix Xizzix · 75 Reals |  |
| chest | Dawn Armor | On AH ~350g · or Leatherworking craft |  |
| wrist | Blackmist Armguards | The Beast · Blackrock Spire · 6g per point of score |  |
| hands | Devilsaur Gauntlets | 0g per point of score |  |
| waist | Shifter's Belt | Pix Xizzix · Reals (price not in data) · Druid only |  |
| legs | Devilsaur Leggings | On AH ~56g · or Leatherworking 290 · pattern on AH ~22g · 1g per point of score |  |
| feet | Drudge Boots | Quest: The Nightmare Manifests | avail doubtful |
| finger | Watcher's Signet of Vigilance | Watchers Revered · no rep source found yet | avail doubtful |
| finger | Tracker's Signet of Vigilance | Watchers Honored · no rep source found yet | avail doubtful |
| trinket | Chained Essence of Eranikus | Quest: The Essence of Eranikus, Sunken Temple |  |
| trinket | Adaptive Combat Assistant | Pix Xizzix · 125 Reals |  |
| mainhand | Hammer of Thundering Hooves |  |  |
| ranged | Idol of the Ursine Twins | Enchanting 300 craft · Formula: Idol of the Ursine Twins (279300) has no known source | avail doubtful |

### Max threat (Night Elf)

| Slot | Pick | Source | Flag |
|---|---|---|---|
| head | Defias Enforcer's Mask | Trash, likely Alcaz Prison like its set (zone 16620, no NPC known) |  |
| neck | Beads of Ogre Might | Quest: Falrin's Vendetta (Dire Maul) · 25 Ogre Warbeads from Dire Maul or Blackrock Spire  |  |
| shoulder | Truestrike Shoulders | Pyroguard Emberseer · Blackrock Spire · 7g per point of score |  |
| back | Howler's Furs | Pix Xizzix · 75 Reals |  |
| chest | Dawn Armor | On AH ~350g · or Leatherworking craft |  |
| wrist | Blackmist Armguards | The Beast · Blackrock Spire · 2g per point of score |  |
| hands | Devilsaur Gauntlets | 0g per point of score |  |
| waist | Shifter's Belt | Pix Xizzix · Reals (price not in data) · Druid only |  |
| legs | Devilsaur Leggings | On AH ~56g · or Leatherworking 290 · pattern on AH ~22g · 0g per point of score |  |
| feet | Drudge Boots | Quest: The Nightmare Manifests | avail doubtful |
| finger | Watcher's Signet of Vigilance | Watchers Revered · no rep source found yet | avail doubtful |
| finger | Tracker's Signet of Vigilance | Watchers Honored · no rep source found yet | avail doubtful |
| trinket | Chained Essence of Eranikus | Quest: The Essence of Eranikus, Sunken Temple |  |
| trinket | Adaptive Combat Assistant | Pix Xizzix · 125 Reals |  |
| mainhand | Hammer of Thundering Hooves | 30g per point of score |  |
| ranged | Idol of the Ursine Twins | Enchanting 300 craft · Formula: Idol of the Ursine Twins (279300) has no known source | avail doubtful |

### Max survival (Tauren)

| Slot | Pick | Source | Flag |
|---|---|---|---|
| head | Feralheart Faceguard | Forever's Dungeon Set 2 copy (set 1667) · quest, Dungeon Set 1 pieces + gold, quest giver  | avail doubtful |
| neck | Amulet of the Darkmoon | Quest: 1200 Tickets - Amulet of the Darkmoon · Faire weeks Nov 9-16, Dec 7-14 (Mulgore), J |  |
| shoulder | Feralheart Pauldrons | Forever's Dungeon Set 2 copy (set 1667) · quest, Dungeon Set 1 pieces + gold, quest giver  |  |
| back | Shifting Cloak | Leatherworking 300, pattern from Knot Thimblejack's Cache, Dire Maul |  |
| chest | Dawn Armor | On AH ~350g · or Leatherworking craft · 2g per point of score |  |
| wrist | Feralheart Wristguards | Forever's Dungeon Set 2 copy (set 1667) · quest, Dungeon Set 1 pieces + gold, quest giver  |  |
| hands | Feralheart Grips | Forever's Dungeon Set 2 copy (set 1667) · quest, Dungeon Set 1 pieces + gold, quest giver  |  |
| waist | Feralheart Waistguard | Forever's Dungeon Set 2 copy (set 1667) · quest, Dungeon Set 1 pieces + gold, quest giver  |  |
| legs | Feralheart Legguards | Forever's Dungeon Set 2 copy (set 1667) · quest, Dungeon Set 1 pieces + gold, quest giver  |  |
| feet | Drudge Boots | Quest: The Nightmare Manifests | avail doubtful |
| finger | Watcher's Signet of Vigilance | Watchers Revered · no rep source found yet | avail doubtful |
| finger | Tracker's Signet of Vigilance | Watchers Honored · no rep source found yet | avail doubtful |
| trinket | Chained Essence of Eranikus | Quest: The Essence of Eranikus, Sunken Temple |  |
| trinket | Darkmoon Card: Heroism | Darkmoon Faire deck · Forever card drops unconfirmed | avail doubtful |
| mainhand | Unyielding Maul | Tribute run, Dire Maul · 5g per point of score |  |
| ranged | Idol of the Ursine Twins | Enchanting 300 craft · Formula: Idol of the Ursine Twins (279300) has no known source | avail doubtful |

