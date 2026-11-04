# The Warden shaman (shaman tank)

Forever's shaman tank (`WardenShaman`, `sim/shaman/warden`, `ui/warden_shaman`): a one-hander with Rockbiter
Weapon and a shield, Spirit Weapons for the parry and the threat, Stormstrike, Earth Shock and Lightning
Shield for threat, Stoneskin Totem and Guardian Totems for mitigation. The tank workers' brief
(`orchestration/brief_tanks.md`, T1) and `docs/tank_sim.md` say how tanks are scored.

**Where it stands (2026-11-03, realigned 2026-11-04 to client 70753 and the November 2 notes, below):** the Warden tanks in the sim (the boss attacks it, it blocks with its shield,
parries with Spirit Weapons, and Rockbiter's and Spirit Weapons' threat are on), every tank ability and talent
is audited against client 70753 (the October 30 hotfix's build) and tested, it wears a synthetic tank set by
standards 36, and it has a rotopt template (`tools/rotopt/shaman_tank.go`) and three presets
(`tools/tankbench/req/shaman_warden_*.json`, `ui/warden_shaman`).

## Presets (re-optimized 2026-11-04, tank round 2)

Three presets on the six scoring encounters (Magmadar, Golemagg, Onyxia, Broodlord, Vaelastrasz, Chromaggus), by
the tank objective at w = 1/3, 1 and 3 (`docs/tank_sim.md`): `tools/tankbench/req/shaman_warden_<preset>.json`,
`ui/warden_shaman/apls/level60_<preset>.apl.json`, the synthetic set (`gear_sets/phase1_synthetic_tank.gear.json`,
and `phase1_synthetic_tank_armor_trinket.gear.json` for Max survival) and three builds in `presets.ts`. The spec
page's launch state is unchanged (blank gear, the default APL, the Level 60 talents). Ragnaros is benched, not
scored. Re-searched on the round 2 engine (master 5668ea5: the boss in front parries and blocks, the logs' swing
gap with parry haste off, survival cooldowns that never read health in scoring runs, Elixir of the Phalanx and
Greater Fortitude on every tank).

| Preset | Talents | On-use trinket | Rotation (knobs off the template's defaults) | Race |
|---|---|---|---|---|
| Max threat | `0535-050032303351012201-050002` | attack power (990015) | Mana Spring Totem (`waterTotem=1`), `armorTrinket=0` | Tauren |
| Balanced | `05050002-050032203551012201-050002` | attack power (990015) | `armorTrinket=0` | Tauren |
| Max survival | `05050003-050032113541012201-050002` | armor (999216) | `armorTrinket=0` (the trinket goes out with the other cooldowns) | Tauren |

All three put the shock above Stormstrike (`shockFirst=1`, the template's default since this round).

**Scores** (mean of the six at the preset's own weight; tankbench, 1,000 iterations, `orchestration/audit/ref/tank_bench/summary.tsv`):

| Preset | Before (round 1 build on the round 2 engine) | After |
|---|---|---|
| Max threat | 2,175 | 2,292 |
| Balanced | 414 | 548 |
| Max survival | -4,571 | -4,357 |

**Earth Shock above Stormstrike** (tank_rotations-20261104-4). Earth Shock makes about 2,225 threat a cast against
Stormstrike's 1,500, and with Stormstrike first the shock waited whenever both were ready. Swapping the two lines in
the round 1 requests (paired, `tools/audit/tank_rotations/abtest`, scoring mean): Max threat +91.6 +/- 10.6 (1,000
iterations), Balanced +105.0 +/- 12.5 (3,000), Max survival +129.3 +/- 33.1 (3,000); every scoring boss gains at
every weight (Broodlord +141 / +104 / +107). The template's `shockFirst` knob (default 1) emits it, and the knob
searches kept it at all three weights.

**Talents** (`rotopt -talent-search -require spiritWeapons`, 1,000 iterations a candidate, 3,000 to confirm; paired
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

**Knobs** (600 to 1,000 iterations, 3,000 to confirm): the defaults everywhere (Stoneskin, Lightning Shield, Earth
Shock on cooldown above Stormstrike, no fire totem, no self heal), plus Mana Spring Totem at Max threat. Max
survival's dedicated armor trinket line lost 3.3 +/- 1.2 to the autocast line: since round 2's engine a scoring run
never reads health, so the autocast line fires the defensive trinket on cooldown too.

**Races**: not re-searched this round; round 1 put Tauren first at every weight by a wide margin (Tauren 2,501 /
515 / -5,312 against Orc 2,441 / 123 / -6,830 and the rest lower; its 5% health keeps Broodlord's death floor
lowest).

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

**On-use trinket** (round 1, B, 3,000 iterations): the armor one (999216) against the attack power one (990015): Max
threat 2,395 against 2,476, Balanced 444 against 463, Max survival -5,409 against -5,575. So Max survival wears the
armor trinket and the other two the attack power one. Not re-run this round.

**Gear and mechanics checked for correctness, not for the gap** (tank_fairness-20261104-3, user ruling 2026-11-04,
standards 40). The Warden's 2 to 3 times higher healing need than the plate tanks is a real class gap: a mail class
with much less defense built in, which won't necessarily be a viable raid tank. Nothing here compensates for it. What
was checked: the set sits at its Tanks row (standard 36, `tools/database/forever_synthetic_items_shaman_tank.go`);
the hit over the cap ties (above); the shield block, Spirit Weapons' parry, Lightning Shield, Stoneskin and
Rockbiter have tests that fail when they break (Tests); the abilities and talents audit (Rules audit) has no wrong
row.

### Numbers (1,000 iterations, all seven encounters)

`tankbench -iters 1000` on each encounter. Healing needed is at 5% deaths, in HPS; the three scores are the
objective at w = 1/3, 1 and 3. The mean is over the six scoring encounters (Ragnaros is reported, not scored). The
other tanks' rows are in `orchestration/audit/ref/tank_bench/summary.tsv`.

**Max threat** (`shaman_warden_max_threat.json`)

| Encounter | TPS | SafeDPS | DTPS | Healing needed | Death floor | Max threat | Balanced | Max survival |
|---|---|---|---|---|---|---|---|---|
| Magmadar | 1458 | 1555 | 527 | 471 | 0.0% | 2562 | 1467 | -1819 |
| Golemagg | 1434 | 1523 | 572 | 536 | 0.0% | 2423 | 1177 | -2563 |
| Onyxia | 1567 | 1691 | 657 | 618 | 0.0% | 2318 | 189 | -6200 |
| Broodlord | 1504 | 1612 | 763 | 1346 | 0.2% | 1878 | -813 | -8888 |
| Vaelastrasz | 1559 | 1761 | 1411 | 1658 | 0.0% | 2826 | 1433 | -2746 |
| Chromaggus | 1315 | 1415 | 911 | 1095 | 0.0% | 1747 | -421 | -6925 |
| Ragnaros | 1346 | 1432 | 1426 | 2450 | 0.1% | 286 | -4871 | -20341 |
| Mean of the six | 1473 | 1593 | 807 | 954 | 0.0% | 2292 | 505 | -4857 |

**Balanced** (`shaman_warden_balanced.json`)

| Encounter | TPS | SafeDPS | DTPS | Healing needed | Death floor | Max threat | Balanced | Max survival |
|---|---|---|---|---|---|---|---|---|
| Magmadar | 1417 | 1517 | 521 | 458 | 0.0% | 2502 | 1437 | -1759 |
| Golemagg | 1387 | 1479 | 568 | 532 | 0.0% | 2339 | 1103 | -2607 |
| Onyxia | 1523 | 1649 | 644 | 592 | 0.0% | 2277 | 237 | -5884 |
| Broodlord | 1464 | 1573 | 758 | 1237 | 0.0% | 1910 | -563 | -7983 |
| Vaelastrasz | 1521 | 1723 | 1359 | 1554 | 0.0% | 2793 | 1487 | -2430 |
| Chromaggus | 1261 | 1365 | 900 | 1058 | 0.0% | 1682 | -414 | -6701 |
| Ragnaros | 1302 | 1392 | 1348 | 2046 | 0.0% | 631 | -3675 | -16594 |
| Mean of the six | 1429 | 1551 | 792 | 905 | 0.0% | 2250 | 548 | -4561 |

**Max survival** (`shaman_warden_max_survival.json`)

| Encounter | TPS | SafeDPS | DTPS | Healing needed | Death floor | Max threat | Balanced | Max survival |
|---|---|---|---|---|---|---|---|---|
| Magmadar | 1359 | 1440 | 496 | 434 | 0.0% | 2375 | 1366 | -1660 |
| Golemagg | 1340 | 1407 | 548 | 508 | 0.0% | 2224 | 1044 | -2498 |
| Onyxia | 1424 | 1546 | 611 | 544 | 0.0% | 2155 | 279 | -5347 |
| Broodlord | 1378 | 1477 | 718 | 1232 | 0.0% | 1721 | -744 | -8137 |
| Vaelastrasz | 1439 | 1649 | 1272 | 1385 | 0.0% | 2716 | 1552 | -1940 |
| Chromaggus | 1318 | 1378 | 879 | 1045 | 0.0% | 1721 | -349 | -6558 |
| Ragnaros | 1268 | 1337 | 1301 | 1933 | 0.1% | 640 | -3429 | -15635 |
| Mean of the six | 1376 | 1483 | 754 | 858 | 0.0% | 2152 | 525 | -4357 |

## Calls for the user

- **(a) Earth Shock's threat multiplier.** Default 2 (Season of Discovery's, wowsims/sod, tier 2, and Classic's;
  the tooltip says "causes a high amount of threat"; no Forever source gives a number). wowsims/forever uses 1.
  Measured in round 1 on the then Balanced preset (1,000 iterations, means of the six, older engine): at 1, TPS
  1,589 to 1,399, SafeDPS 1,745 to 1,526, scores 2,493 / 501 / -5,475 to 2,056 / 63 / -5,913. With the shock now
  above Stormstrike more of the Warden's threat rests on it. A /combatlog threat check in game would settle it.
- **(b) The defensive on-use trinket.** Default: the synthetic 1,665 armor for 20 s every 2 min (999216) in Max
  survival, as the defensive counterpart of the Enhancement set's attack power trinket at the same item level. The
  alternative, the attack power trinket everywhere, cost Max survival 166 in round 1 (-5,575 against -5,409).
- **(c) Rockbiter's threat on Stormstrike** (to verify; ruled 2026-11-04, standards 40). The sim keeps 72 x weapon
  speed on specials, as on white swings (wowsims/forever's reading); wowsims/sod gives a special a flat 72, which
  would take about 35 TPS (2.5%) off the Warden. One Stormstrike's threat in a combat log settles it
  (`sim/shaman/rockbiter_weapon.go` says so; the flat case is
  `orchestration/audit/ref/tank_mechanics/patches/shaman_02_rockbiter_special_flat_CONDITIONAL.patch`).
- **(d) Hit above the melee cap.** Default: the set keeps the row's 5.5% gear hit (10.5% with Tidal Focus); moving
  the 2.5% over the cap into crit ties at Balanced on the round 2 engine (415 against 404, unpaired).
- **(e) Boots for Max threat.** Max threat would gain 2.7 +/- 0.5 (0.1%) from Mystic Rugged Armor Kit on the boots,
  but it shares Balanced's gear set, where Greater Stamina wins by 6.6. Default: one set, Greater Stamina; a Max
  threat set of its own is the alternative.
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
   Broodlord therefore decides the Warden's score and its race: Tauren is the only race that survives him.

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
  (517-545 at 60, threat 2 x damage: the biggest threat change), Stormstrike 150% and 6 s, Mental Dexterity
  50 / 100 / 150%, Flurry 6 to 30%, Maelstrom Weapon 4 PPM a point, Lightning Bolt rank 10 at 334-372.
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
- **Tests:** `TestForeverWardenTalents` reads Mental Dexterity at 50 / 100 / 150% and the new talent indices,
  `TestForeverWardenAbilities` Stormstrike's 6 s.

## Rules audit

Client 70753 (`tools/sources/wago.py`, no hotfix rows logged for it: the October 30 hotfix is in its tables,
`orchestration/update_ledger.md`), forever-wiki/classes/shaman.md, `shaman_audit.md` (the shaman's DPS audit,
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
| Earth Shock rank 7 (10414) | client 70753: 301 centre (unchanged); **the notes: Era's base, 517-545 at 60** (notes-only, standards 37); 38.6% of spell power, 450 mana, 6 s shared shock cooldown (category 19), 1.5 s GCD, "Causes a high amount of threat" | 517-545 (master's shaman round, `earth_shock.go`; the x1.08 is 1.0 now); threat 2 x damage | threat multiplier: 2, Season of Discovery's (wowsims/sod earth_shock.go `ThreatMultiplier = 2`), tier 2, and Classic's (wowsims/classic), tier 3; no vmangos; wowsims/forever gives 1 (no source); the tooltip says "high" | right, **call (a)** |
| Flame Shock, Frost Shock | no threat text | threat 1 x damage | client text, tier 3 | right |
| Stormstrike (17364) | client 70753: 100% weapon damage, 8 s (unchanged); **the notes: 150% normalized weapon damage, 6 s** (notes-only); 8% of base mana, +20% to the next Lightning Bolt, Chain Lightning or Earth Shock | 150%, 6 s (master's round, `stormstrike.go`; a flat "+N damage" stays 1 x); threat 1 x damage plus Rockbiter's 72 x speed (not scaled by the 150%) | notes, client, 7.7 | right |
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
| Mental Dexterity, Mental Quickness | client: attack power from 33 / 67 / 100% of Intellect, **the notes: 50 / 100 / 150%** (notes-only); spell power from 15% a point | 50 / 100 / 150%; 15% a point | right |
| Stormstrike, Improved Stormstrike (1223031) | at 2/2 every dodge or parry of ours resets Stormstrike; 50% regen while casting, 15 s | the same | right (beta, 8.32) |
| Flurry, Maelstrom Weapon, Rage of the Farseer | DPS talents; the notes: Flurry 6 / 12 / 18 / 24 / 30%, Maelstrom Weapon's chance doubled (the sim's 4 PPM a point), Lightning Bolt's top rank +76%; Rage of the Farseer 30% attack speed for 25 s, 3 min | as master's shaman round (`round_70753/shaman.md`) | right |
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
  GCD; Stormstrike's 6 s (the notes) and 8% of base mana; Lightning Shield's 10 min, 3 orbs, 3.5 s apart, gone after the
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
Lesser Arcanum of Voracity (Stamina) on head and legs, Strength on the neck, Superior Defense on the back, Greater
Stats on the chest, Superior Stamina on the wrists, Superior Strength on the hands, Greater Stamina on the feet
(it beats the Rugged Armor Kits at Balanced and Max survival, Presets) and the shield, Crusader on the weapon.

## Rotation

`tools/rotopt/shaman_tank.go` (`-template shaman_tank`): the shield whenever it's down, the earth totem
(Stoneskin by default), the shock with a mana floor (Earth Shock by default, holding for Stormstrike's +20% as a
knob), Stormstrike on cooldown (below the shock since round 2: `shockFirst`), the fire and water totems (none by default; a fire totem's own damage is the
totem's threat, so only Fire Nova feeds ours), Lightning Bolt at Maelstrom Weapon stacks, a health-blind Lesser
Healing Wave, and the defensive trinket on a timer. The tank applies no raid debuff (a shaman has none that the
bench's raid gives: Demoralizing Shout and Thunder Clap stay the raid's), so `tankscore.OwnDebuffs` leaves the
Warden's raid as it is.
