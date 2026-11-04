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
beta measurement, (2) wowsims/forever's reading, (3) Classic's (vmangos, wowsims/classic).

### Abilities

| Ability | Client (70753) | Sim | Source, tier | Verdict |
|---|---|---|---|---|
| Rockbiter Weapon rank 7 (16316, passive 16313) | 653 attack power at 60 (554 + 16.5 a level to 62), 150 mana, 60 min | 653, x1.07 / 1.13 / 1.2 with Elemental Weapons | client; beta sheet at 30 (7.4), tier 1 | right |
| Rockbiter's threat (proc 20871) | effect 77 (script), 72 points; cast off the passive on done melee auto and special hits (ProcTypeMask 20) | 72 x weapon speed on each landed white hit; **was a flat 72 on Stormstrike, now 72 x speed** | client mask; wowsims/forever's reading of the script effect (threat a second of swing), tier 2 | **fixed** |
| Spirit Weapons' threat (16268) | effect 1 -30% threat, all schools; effect 2 +86 on Rockbiter's passive threat effect (0 alone) | **0.7 x 1.86 = 1.302 under Rockbiter (was 1.3)**, 0.7 otherwise; picked by the imbue in the options or consumes | client, tier 1 | **fixed** |
| Spirit Weapons' parry | not a client effect ("Gives a chance to parry") | can parry: 5% base, plus Defense, less 0.6% for the boss's level | beta (6.4: 18 of 380 below us, low above), tier 1 | right |
| Shield block | Classic's: 5% base with a shield, Defense on top | 5% (less 0.6% vs a level 63), block value = the shield's + Strength / 20 - 1 | Classic, tier 3; no shaman shield on the beta (6.4) | right |
| Earth Shock rank 7 (10414) | client 70753: 301 centre (unchanged); **the notes: Era's base, 517-545 at 60** (notes-only, standards 37); 38.6% of spell power, 450 mana, 6 s shared shock cooldown (category 19), 1.5 s GCD, "Causes a high amount of threat" | 517-545 (master's shaman round, `earth_shock.go`; the x1.08 is 1.0 now); threat 2 x damage | threat multiplier: Classic's 2 (wowsims/classic, vmangos), tier 3; wowsims/forever gives 1 (no source); the tooltip says "high" | right, **call (a)** |
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
- `TestForeverWardenImprovedStormstrikeReset`: a dodge or parry of the boss's swing resets Stormstrike, a
  block or a hit doesn't.
- `TestForeverWardenTank`: the bench's Balanced request against the level 63 boss with a stand-in healer: the
  boss attacks the Warden, which blocks and parries, the rotation's abilities go out, Lightning Shield and
  Stoneskin stay up, and the tank score replays every iteration.

Each test fails with the mechanic broken (the old flat Rockbiter threat on Stormstrike, the old 1.3, a shaman
parrying without the talent, Stoneskin after armor).

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
and the shield, Crusader on the weapon.

## Rotation

`tools/rotopt/shaman_tank.go` (`-template shaman_tank`): the shield whenever it's down, the earth totem
(Stoneskin by default), Stormstrike on cooldown, the shock with a mana floor (Earth Shock by default, holding for
Stormstrike's +20% as a knob), the fire and water totems (none by default; a fire totem's own damage is the
totem's threat, so only Fire Nova feeds ours), Lightning Bolt at Maelstrom Weapon stacks, a health-blind Lesser
Healing Wave, and the defensive trinket on a timer. The tank applies no raid debuff (a shaman has none that the
bench's raid gives: Demoralizing Shout and Thunder Clap stay the raid's), so `tankscore.OwnDebuffs` leaves the
Warden's raid as it is.
