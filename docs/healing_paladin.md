# Holy paladin (H3)

The Holy paladin as a healer of the healing sim ([healing_sim.md](healing_sim.md)): its heals and healing
talents against Forever's newest client, the agent, tests, a synthetic set, a tank healer and a raid healer
rotation tuned on `tools/healbench`, and its numbers against the Restoration shaman's presets. The brief is
`orchestration/brief_healers.md`.

**Where it stands (2026-11-03):** registered (`sim/paladin/holy`), every heal and healing talent in and
tested, a synthetic set (990301-990314), two presets. The UI spec page stays unlaunched (the user's call).

## The rules audit

Sources, best first ([data_sources.md](data_sources.md)): tier 1 the client tables on wago.tools, build
**1.60.1.70753** (2026-11-02), the first client that carries the launch realms' October 30 hotfix (the
paladin tree rework; `orchestration/update_ledger.md`, Build 1.60.1.70753: "the client values now agree with
the sim" for the hotfix's Retribution side). Tables: SpellEffect, SpellMisc, SpellLevels, SpellPower,
SpellCooldowns, SpellCastTimes, SpellDuration, SpellAuraOptions, SpellClassOptions, ItemEffect,
ItemXItemEffect, and the talent tree `tools/forever_talents/client/paladin.json` (TraitTree 1100 from the
hotfix). Tier 2 the BlizzCon level 38 footage (forever-wiki/classes/paladin.md) and the Wowhead Forever
tooltip of Light's Vigil embedded in the cached Holy guide page. A heal's value at a level is the client's
base at its learn level plus its growth per level up to its max level, spread by the variance:
`base * (1 - v/2)` to `base * (1 + v/2)`, as the shaman's heals (`sim/shaman/healing.go`) read the same
tables. The footage agrees to the point (Holy Light 6 at 38: 610 to 682; Flash of Light 3 at 38: 101 to 114).
Forever's downranking (5% of the coefficient a level past learn level + 17, the shaman's and the priest's
rule) applies to every heal.

"Before" is the sim on master 58d6e8814; "now" is this branch.

### Holy Light (2.5 sec, coefficient 0.714 on every rank)

| Rank | Spell | Learned (max) | Base, per level, variance | At 60 | Mana | Before | Now |
|---|---|---|---|---|---|---|---|
| 1 | 635 | 1 (5) | 43, 0.8, 0.186 | 41.9 to 50.5 | 35 | missing | client |
| 2 | 639 | 6 (11) | 83, 1.1, 0.169 | | 60 | missing | client |
| 3 | 647 | 14 (19) | 155, 1.7, 0.162 | | 110 | missing | client |
| 4 | 1026 | 22 (27) | 287, 2.4, 0.138 | | 190 | missing | client |
| 5 | 1042 | 30 (35) | 452, 3.1, 0.119 | 439.7 to 495.3 | 275 | missing | client |
| 6 | 3472 | 38 (43) | 646, 3.8, 0.111 | 628.1 to 701.9 (0.75 of the coefficient) | 365 | missing | client |
| 7 | 10328 | 46 (51) | 899, 4.6, 0.108 | 872.2 to 971.8 | 465 | missing | client |
| 8 | 10329 | 54 (59) | 1217, 5.2, 0.108 | 1176.0 to 1310.0 | 580 | missing | client |
| 9 | 25292 | 60 (65) | 1580, 5.8, 0.107 | 1495.4 to 1664.6 | 660 | missing | client |

### Flash of Light (1.5 sec, coefficient 0.429)

| Rank | Spell | Learned (max) | Base, per level, variance | At 60 | Mana | Before | Now |
|---|---|---|---|---|---|---|---|
| 1 | 19750 | 20 (25) | 46, 1.0, 0.149 | | 35 | missing | client |
| 2 | 19939 | 26 (31) | 66, 1.3, 0.136 | | 50 | missing | client |
| 3 | 19940 | 34 (39) | 101, 1.6, 0.117 | 102.6 to 115.4 (0.55 of the coefficient) | 70 | missing | client |
| 4 | 19941 | 42 (47) | 151, 1.9, 0.115 | 151.3 to 169.7 (0.95) | 90 | missing | client |
| 5 | 19942 | 50 (55) | 223, 2.2, 0.113 | 220.8 to 247.2 | 115 | missing | client |
| 6 | 19943 | 58 (63) | 303, 2.6, 0.110 | 291.2 to 325.2 | 140 | missing | client |

### Holy Shock (instant, 10 sec cooldown, 20 Holy points)

The cast spell (1311606, 20473, 20929, 20930) picks its damage or its heal spell by the target: damage at an
enemy (already in the sim, unchanged), heal at an ally. The ranks share the cooldown (one category); the sim
had a timer each.

| Rank | Cast spell | Heal spell | Learned | Heal at 60 | Coefficient | Mana | Before | Now |
|---|---|---|---|---|---|---|---|---|
| 1 | 1311606 | 1311605 | 30 | 109.7 to 118.3 (the tooltip's 110 to 118) | 0.429 (0.35 at 60) | 160 | damage only | client |
| 2 | 20473 | 25914 | 40 | 150.1 to 161.9 | 0.429 (0.85) | 225 | damage only | client |
| 3 | 20929 | 25913 | 48 | 221.3 to 238.7 | 0.429 | 275 | damage only | client |
| 4 | 20930 | 25903 | 56 | 307.4 to 332.6 | 0.429 | 325 | damage only | client |

Benediction ("all instant cast spells") takes its share off the cost, as it did for the damage.

### Light's Vigil (the 31 point Holy talent)

Since the hotfix (1310911, 1311590, 1311595): "Applies Light's Vigil to a friendly target for 30 sec. The
caster's next Holy Shock on the target incurs no cooldown and heals the target and their party for 704. Only
one Vigil per Paladin may be active in each party."

| | Client | Before | Now |
|---|---|---|---|
| Ranks | 40, 50, 60: 730, 1000, 1340 mana | an empty stub | the three ranks |
| Cast | 1.5 sec, a 6 sec cooldown the ranks share (category recovery) | | the same |
| Mark | 1310909 / 1311593 / 1311597, 30 sec, 1 charge | | one per paladin per party |
| Party heal | 1310912 / 1311591 / 1311596: 324 + 1.2 a level (to 49), 487 + 1.5 (to 59), 704 + 1.8 (to 69), variance 0.056, coefficient 0.143, on the target and its party | | the same, every living player of the target's party |
| Holy Shock at the mark | no cooldown | | its cooldown is reset |
| Mana refund | the cast spell's second dummy still stores 75 (the old 75% refund); the hotfix's tooltip has none | | none (Calls for the user) |

Wowhead's tooltip in the cached guide shows rank 3 as 700 to 740, its value at the spell's max level 69
(`ppl60:69:704:180`); at 60 it is 684 to 724.

### Lay on Hands

| | Client | Before | Now |
|---|---|---|---|
| Ranks | 633 (10), 2800 (30, 250 mana back), 10310 (50, 550) | the top rank | the same |
| Target | a friendly target | the paladin only | any player |
| Heal | the paladin's max health; drains all its mana, "does not interrupt Mana regeneration" | the paladin's max health on itself | on the target, no crit, the target's healing taken applies |
| Mana | restores 550 of the target's | 550 to the paladin | 550 to the target |
| Cooldown | 20 min (category recovery 1200000) | 20 min | 20 min |

### The healing talents (the hotfix's tree, client 1.60.1.70753)

| Talent | Client | Before | Now |
|---|---|---|---|
| Spiritual Focus (20205) | 14% a rank of 5: Flash of Light, Holy Light and Light's Vigil don't lose cast time to a hit | not read | PushbackReduction on the three, added to Concentration Aura's 35% (it was the larger of the two) |
| Healing Light (20237) | 4 / 8 / 12% healing, aura 108 on Holy Light, Flash of Light and Holy Shock's heal (and the Oath bonus spells) | not read | the same |
| Divine Favor (20216) | next Flash of Light, Holy Light or Holy Shock 100% crit, 1 charge; 4% of base mana; 2 min; off the global cooldown | Holy Shock's damage only, free | all of them, 60.48 mana (Benediction off it) |
| Illumination (20210) | a crit of Flash of Light, Holy Light, Light's Vigil or Holy Shock's heal: 20% a rank of 5 for mana equal to 50% of the spell's base cost (Classic's 100%) | not read | the same; Light's Vigil once a detonation (Calls for the user) |
| Infusion of Light (426065, buff 437063) | Holy Shock and Flash of Light crits take 0.5 sec a rank of 2 off the next Holy Light within 15 sec (proc flags: helpful and harmful spells) | not read | the same; Holy Shock's damage crits count too |
| Oath of Light (1328548) | on one ally for 5 min: Holy Light up to 400, Flash of Light up to 115, Holy Shock up to 230 more (6.67, 1.92 and 3.83 a level); 8% of base mana; one Oath | not read | the bonus on the heal itself (Calls for the user) |
| Boundless Light (1328550) | +15% a rank of 2 Holy Shock crit at an ally our Oath isn't on | not read | the same |
| Holy Power (5923) | aura 50 (critical healing amount), Holy school, 10 to 50%; no damage crit since the hotfix | no damage crit (right), no heal | a heal's whole crit 10% bigger a rank: 150% to 225% at 5/5 (Calls for the user) |
| Light's Vigil | above | stub | above |
| Divine Intellect (Protection, 20257) | 2 to 10% Intellect | in | unchanged |
| Reverence (1310899) | 15% a rank of 2 of the regen while casting | in | unchanged |
| Conviction (Retribution, 20117) | 2% crit a rank of 3 with all spells and attacks | in (spell crit) | reaches the heals |
| Benediction (Retribution) | 2% a rank off instant spells | in | also Divine Favor and Oath of Light |
| Champion of the Light (Retribution, 1311084) | aura 174: spell damage from Intellect | spell power (heals too) | unchanged: no healer build reaches it (it needs 25 Retribution points); noted below |

Not modelled, no healing effect in a raid fight: Purifying Power's Cleanse cost, Unyielding Faith,
Guardian's Favor, Stoicism, Divine Precision (Holy hit), Consecrated Ground, Holy Conduit.

### Items

- **Libram of Economy** (272436, spell 1291089, a Stranglethorn drop with Libram of Law): Holy Light 5% less
  mana. Added (`sim/paladin/healing.go`).
- **Libram of Infusion** (279248): Holy Shock 6% crit, the heal too (its mask covers 1311605's family). Was
  already in for the damage.
- **Libram of Holy Alacrity** (228175): Holy Shock takes time off the next Holy Light, but its buff spell
  (449982) has no effect rows in 70753, so its value is unknown. Not added.
- Libram of Divinity (23201) heals the party for 25 on a Judgement now (1302540), and Libram of Light is
  Naxxramas. Not used.
- Justice Armor (2107), the Tier 1 Holy set: the synthetic set carries its 2-piece (+10 Spirit) on two
  pieces, as every class's preset does (standards 6). 4-piece +26 healing, 5-piece 1 sec off Holy Shock.

### Fixed elsewhere on the way

- Light's Vigil's test on the Retribution side (`TestLightsVigilHealsOnly`) now checks the detonation spells
  are gone, since the Vigil itself is a heal.
- The Ret Shock report hashes moved: Divine Favor's 4% of base mana (-0.01 DPS at 180 s) and Holy Shock's one
  shared cooldown timer, which renumbers the timers in the first iteration's log.
