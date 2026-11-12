# Holy paladin (H3)

The Holy paladin as a healer of the healing sim ([healing_sim.md](healing_sim.md)): its heals and healing
talents against Forever's newest client, the agent, tests, a synthetic set, a tank healer and a raid healer
rotation tuned on `tools/healbench`, and its numbers against the Restoration shaman's presets. The brief is
`orchestration/brief_healers.md`.

**Where it stands (2026-11-04):** registered (`sim/paladin/holy`), every heal and healing talent in and
tested, re-checked against client 1.60.1.70753, the November 2 notes and the 11-03 hotfixes (no Holy
change left to take), a synthetic set (990301-990314), two presets tuned on the merged build. Re-benched
2026-11-04 after the engine fixes: tank healer mean carry 1.043, raid healer 1.413 over the 7 boss files (2026-11-12: tank 1.152 on the moved
bench, raid 1.430 as a Flash of Light healer, below)
(the side-by-side with the other healers is H3 in healing_sim.md). The UI spec page stays unlaunched (the user's
call).

A healing paladin never bubbles the tank (standards 40, user ruling 2026-11-04: it costs the tank its threat):
no preset or bench request casts Blessing of Protection, Divine Intervention, Divine Shield or Divine
Protection (`sim/paladin/holy/no_bubble_test.go`).

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
Forever's downranking (5% of the coefficient a level past learn level + 17) applies to every heal, from the
engine since master 8f304b240 and b57f5cc0a (`sim/core/downrank.go`, a user ruling): the heals give their
`Rank` and `RequiredLevel` and `RegisterSpell` scales the coefficient. This branch had the same rule by hand
before the merge, so nothing moved (see "Round 70753" below).

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
| Divine Favor (20216) | next Flash of Light, Holy Light or Holy Shock 100% crit, 1 charge; 4% of base mana; 2 min; off the global cooldown | Holy Shock's damage only, free | all of them, 60.48 mana; no Benediction (its class mask is empty; healer_mechanics-20261104-4) |
| Illumination (20210) | a crit of Flash of Light, Holy Light, Light's Vigil or Holy Shock's heal: 20% a rank of 5 for mana equal to 50% of the spell's base cost (Classic's 100%) | not read | the same; Light's Vigil once a detonation (Calls for the user) |
| Infusion of Light (426065, buff 437063) | Holy Shock and Flash of Light crits take 0.5 sec a rank of 2 off the next Holy Light within 15 sec (proc flags: helpful and harmful spells) | not read | the same; Holy Shock's damage crits count too |
| Oath of Light (1328548) | on one ally for 5 min: Holy Light up to 400, Flash of Light up to 115, Holy Shock up to 230 more (6.67, 1.92 and 3.83 a level); 8% of base mana; one Oath | not read | the bonus on the heal itself (Calls for the user) |
| Boundless Light (1328550) | +15% a rank of 2 Holy Shock crit at an ally our Oath isn't on | not read | the same |
| Holy Power (5923) | aura 50 (critical healing amount), Holy school, 10 to 50%; no damage crit since the hotfix | no damage crit (right), no heal | a heal's whole crit 10% bigger a rank: 150% to 225% at 5/5 (Calls for the user) |
| Light's Vigil | above | stub | above |
| Divine Intellect (Protection, 20257) | 2 to 10% Intellect | in | unchanged |
| Reverence (1310899) | 15% a rank of 2 of the regen while casting | in | unchanged |
| Conviction (Retribution, 20117) | 2% crit a rank of 3 with all spells and attacks | in (spell crit) | reaches the heals |
| Benediction (Retribution) | 2% a rank off instant spells, by class mask (20101 e0 aura 108 misc 14, mask 0x1be11fbf / 0x26018242 / 0x20212100 / 0x8800) | in | Holy Shock; not Divine Favor or Oath of Light, whose class masks are empty (healer_mechanics-20261104-4; it took 10% off both until 2026-11-04) |
| Champion of the Light (Retribution, 1311084) | aura 174: spell damage from Intellect | spell power (heals too) | unchanged: no healer build reaches it (it needs 25 Retribution points); noted below |

Not modelled, no healing effect in a raid fight: Purifying Power's Cleanse cost, Unyielding Faith,
Guardian's Favor, Stoicism, Divine Precision (Holy hit), Consecrated Ground, Holy Conduit.

### Items

- **Libram of Economy** (272436, spell 1291089, a Stranglethorn drop with Libram of Law): Holy Light 5% less
  mana. Added (`sim/paladin/healing.go`).
- **Libram of Infusion** (279248): Holy Shock 6% crit, the heal too (its mask covers 1311605's family). Was
  already in for the damage.
- **Libram of Holy Alacrity** (228175): its proc 449980 triggers buff 449982, which client 70755 (tables
  byte-identical to 70753's) gives e0 aura 107 misc 10 bp -200: the next Holy Light casts 0.2 sec faster, 10 sec,
  1 charge. The trigger condition is server side (the tooltip: casting Holy Shock). Not added (call 9);
  healer_mechanics-20261104-9 corrected this line, which said 449982 had no effect rows.
- Libram of Divinity (23201) heals the party for 25 on a Judgement now (1302540), and Libram of Light is
  Naxxramas. Not used.
- Justice Armor (2107), the Tier 1 Holy set: the synthetic set carries its 2-piece (+10 Spirit) on two
  pieces, as every class's preset does (standards 6). 4-piece +26 healing, 5-piece 1 sec off Holy Shock.

### Fixed elsewhere on the way

- Light's Vigil's test on the Retribution side (`TestLightsVigilHealsOnly`) now checks the detonation spells
  are gone, since the Vigil itself is a heal.
- The Ret Shock report hashes moved: Divine Favor's 4% of base mana (-0.01 DPS at 180 s) and Holy Shock's one
  shared cooldown timer, which renumbers the timers in the first iteration's log.

## Round 70753 (the November 2 notes, re-checked 2026-11-04)

Re-checked after master took client 1.60.1.70753, the November 2 development notes (post 9) and the
server hotfixes of 2026-11-03 (`orchestration/round_70753/`, standards 37 to 39).

| Note line (Holy healing) | Source that settles it | Sim |
|---|---|---|
| Oath of Light, row 4: 40 yd, 8% of base mana, no cooldown, 5 min, one Oath, bonus grows with level | client 70753 1328548 (6.67 / 1.92 / 3.83 a level, SpellPower 8%, duration 5 min); hotfix rows only push 113139 (10-30), the same values | in (above) |
| Boundless Light, row 5: Holy Shock +15 / 30% crit at an ally without our Oath | client 70753 1328550; hotfix rows only push 113139 | in |
| Holy Power: critical strike healing bonus of Holy spells +10 to 50% (was crit chance) | client 70753 5923-25829 aura 50, Holy | in (the reading is a call below) |
| Light's Vigil friendly only, offensive part removed | client 70753 1310911 / 1311590 / 1311595 (SpellMisc re-sent in push 113341 with the client's values: 1.5 sec, range 5) | in (heals only) |
| Divine Precision 2 points, 4 / 8% Holy hit, Silence 10 / 20% | client 70753 1310904 | Retribution's (the Holy builds take none) |
| Spiritual Focus 5 points, row 1 | client 70753 20205-20209 | in (14% a rank) |
| Divine Favor, Illumination row 3 | client 70753 | in; the trees' strings didn't move |

- **Hotfixes**: wago.tools' hotfix log (region 91) has no row for a Holy heal or healing talent after the
  October 30 push 113139 except push 113341's SpellMisc rows for Light's Vigil and Flash of Light 6, which
  equal the client's (`tools/sources/hotfixes.py fetch 1.60.1.70707 --search <id>`). The searches by a long
  spell id are complete; those by a five-digit id (Holy Power, Divine Favor, Illumination, Healing Light,
  Holy Light 9, Flash of Light 6, Holy Shock 4) match many unrelated rows and got 30 to 70% of them through
  the site's paging, and none of the rows found that names one of these spells differs from the client.
  The 11-03 paladin push 113542 is Vengeance and Two-Handed Weapon Specialization only
  (`orchestration/round_70753/hotfix_113483_paladin.md`). Searched: Oath of Light, Boundless Light, the
  Vigil's cast, mark and heal spells, Holy Shock's heal, Infusion of Light and its buff, Holy Power, Divine
  Favor, Illumination, Healing Light, Holy Light 9, Flash of Light 6 and Holy Shock 4.
- **Notes-only values (standards 37)**: none for the Holy healer. Every Holy line of the notes is in the
  client.
- **The talent remap** (`orchestration/talent_remap_70753.md`): the paladin trees are the same in 70753 as
  the 70707+hf trees this branch was written on, so no talent string moved.
- **Downranking from the engine**: the heals give `Rank` and `RequiredLevel` and core scales their
  coefficients; the hand-written penalty this branch had is gone (it would have applied twice after the
  merge). The Oath bonus takes `core.DownrankFactor` (my reading, a call below). The presets cast Holy
  Light 9 and 8, Flash of Light 6 and Holy Shock 4, all learned within 17 levels of 60, so the penalty is
  1.0 on every cast they make: nothing moved on them (table below).
- **Human Spirit 10%** was already in before this round; **Read Leyline** (50% for 15 sec, not under
  Innervate) is the Skyborne's, not a preset race. A2's race base stats (Classic's offsets) are the only
  other shared change that reaches the presets.

What the merge moved on the presets as they were committed before it (625db99b0: Holy Light 9, 8 and 7),
the branch's build before the merge against after it, healbench, 500 iterations, the files' own lengths. HPS
is the healer's effective healing a second, carry the damage carried at the death line:

| Boss and role | HPS before | HPS after | Carry before | Carry after | Moved |
|---|---|---|---|---|---|
| onyxia tank | 686 | 686 | 1.096 | 1.096 | +0.000 |
| broodlord tank | 590 | 592 | 0.723 | 0.725 | +0.002 |
| chromaggus tank | 776 | 775 | 0.828 | 0.830 | +0.002 |
| golemagg tank | 566 | 566 | 1.695 | 1.698 | +0.003 |
| magmadar tank | 518 | 518 | 1.635 | 1.635 | +0.000 |
| ragnaros tank | 336 | 339 | 0.523 | 0.523 | +0.000 |
| vaelastrasz tank | 739 | 742 | 0.709 | 0.709 | +0.000 |
| onyxia raid | 420 | 420 | 0.714 | 0.714 | +0.000 |
| broodlord raid | 523 | 523 | 0.815 | 0.815 | +0.000 |
| chromaggus raid | 228 | 228 | 0.979 | 0.979 | +0.000 |
| golemagg raid | 260 | 260 | 1.994 | 1.994 | +0.000 |
| magmadar raid | 446 | 446 | 1.490 | 1.490 | +0.000 |
| ragnaros raid | 450 | 450 | 1.254 | 1.254 | +0.000 |
| vaelastrasz raid | 913 | 913 | 0.972 | 0.972 | +0.000 |

Nothing moved past noise (at most +0.003, the tank healer on Golemagg): the engine's downranking gives the same
factors this branch had by hand, and those are 1.0 on every rank the presets cast.


## The rotations (tuned on tools/healbench)

Consumables (2026-11-12): Greater Cleric's Elixir for the Greater Arcane Elixir, Runn Tum Tuber Surprise, Flask of Supreme Power; the tank healer the Wizard Oil, the raid healer the Lesser Mana Oil (the Brilliant Mana Oil is a Zandalar formula); the rest as every healer (`docs/healing_sim.md`, Consumables).

`orchestration/tools/healers/paladin_holy_requests.py` writes both presets from one set of knobs, and a
variant with `--out FILE --set key=value`; each variant was benched against the default, 500 iterations, the
`-carry` score (the damage carried at the death line), the bench's healer reaction time and cancelling. A
carry moves about ±0.01 to 0.02 between seeds at 500 iterations; a difference under that is noise.

### Tank healer (one healer on the tanks, the files' own tank time, and Onyxia at 180 and 300 sec)

The first rotation: Oath of Light on the tank, Divine Favor before a Holy Light 9 the tank misses, Holy Light
9, 8 and 7 at the tank, Flash of Light 6 at the most hurt, Holy Shock under 50% health, Flash of Light under
35%, Lay on Hands under 10%. Carry, against it (each variant changes one knob):

| Variant | Onyxia | Broodlord | Magmadar | Ragnaros | Onyxia 180 s | Onyxia 300 s |
|---|---|---|---|---|---|---|
| the first rotation | 1.096 | 0.723 | 1.635 | 0.523 | 1.011 | 0.898 |
| Holy Light 9 and 8 | 1.090 | 0.734 | 1.641 | 0.510 | **1.031** | **0.922** |
| Holy Light 9, 8 and 6 | 1.066 | 0.669 | 1.600 | 0.503 | 0.971 | 0.860 |
| Holy Light 9 and 7 | 1.094 | 0.721 | 1.632 | 0.522 | | |
| no Holy Shock under 50% | 1.079 | 0.725 | 1.646 | 0.511 | | |
| Holy Shock under 35% | 1.094 | 0.725 | 1.692 | 0.508 | 1.019 | |
| Flash of Light under 25% | 1.094 | 0.719 | 1.635 | 0.521 | | |
| Flash of Light under 50% | 1.094 | 0.722 | 1.690 | 0.516 | 1.020 | |
| no Divine Favor | 1.083 | 0.730 | 1.641 | 0.513 | 1.025 | 0.892 |
| no Oath of Light | 1.070 | 0.742 | 1.669 | 0.484 | 0.944 | 0.779 |
| no Lay on Hands | 1.094 | 0.717 | 1.635 | 0.523 | 1.007 | 0.889 |
| Libram of Infusion (Holy Shock crit) | 1.096 | 0.723 | 1.625 | 0.512 | 1.010 | 0.875 |
| Dwarf | 1.094 | 0.723 | 1.635 | 0.505 | 1.009 | 0.873 |
| Undead | 1.085 | 0.719 | 1.650 | 0.490 | 1.000 | 0.880 |
| no Flash of Light at the most hurt | 0.840 | 0.660 | 1.348 | 0.430 | | |
| Holy Shock at the tank on cooldown | 1.073 | 0.684 | 1.641 | 0.509 | | |
| Concentration Aura | | | | | 1.011 | 0.898 |

On Holy Light 9 and 8 (the files' tank times never run the healer short of mana, so the long fight decides):

| Variant | Onyxia | Broodlord | Magmadar | Ragnaros | Onyxia 300 s |
|---|---|---|---|---|---|
| Holy Light 9 and 8 | 1.090 | 0.734 | 1.641 | 0.510 | 0.922 |
| Holy Light only above 30% mana, Flash of Light below | 1.090 | 0.734 | 1.641 | 0.510 | 0.960 |
| Holy Light only above 50% mana | 1.090 | 0.734 | 1.641 | 0.510 | **1.006** |
| Flash of Light 6 and 5 | 1.085 | 0.719 | 1.641 | 0.513 | 0.922 |
| Flash of Light under 50% | 1.087 | 0.724 | 1.630 | 0.507 | 0.906 |
| Holy Shock under 35% | 1.082 | 0.753 | 1.644 | 0.516 | 0.914 |
| ranks cast at 90% of their size | 1.090 | 0.752 | 1.629 | 0.507 | 0.908 |

- **Flash of Light at the most hurt between Holy Lights is most of the rotation**: without it the tank healer
  carries 0.04 to 0.29 less. Flash of Light 6 with the Oath's 115 is a cheap heal for the tank's small holes.
- **Oath of Light** is worth 0.03 to 0.04 on Onyxia and Ragnaros at their lengths and 0.07 to 0.12 at 180 and
  300 sec; on the short Broodlord and Magmadar stints its global cooldown at the pull costs a little.
- Concentration Aura changes nothing in the tank role: the boss never hits the healer.
- Human is the race: 0.02 to 0.05 ahead of Dwarf and Undead at 300 sec (The Human Spirit's 10% Spirit).

The mana line, on Holy Light 9 and 8, on the merged build (round 70753's client and master's engine):

| Holy Light only above | Onyxia 180 s | Onyxia 300 s |
|---|---|---|
| any mana | 1.031 | 0.924 |
| 40% mana | 1.039 | 1.001 |
| **50% mana** | 1.035 | **1.005** |
| 60% mana | 1.038 | 1.012 |

40 to 60% are the same within noise and all 0.08 ahead of no line at 300 sec; the files' own lengths never
reach it. The preset takes 50%.

### Raid healer (one healer on the raid's assignment, the files' raid time)

The raid healer is the tank healer without the Oath, with every heal at the most hurt. Each variant against
it, 500 iterations: Holy Shock at the most hurt whenever it misses what Holy Shock heals, the Vigil on our
party, ranks out at 80% of their size, Flash of Light 6, 5 and 4, Holy Light 9, 7 and 5. None was ahead past
noise (Onyxia 0.714 the base and the best; Broodlord 0.831 for Flash of Light 6, 5 and 4 and 0.822 for the
Vigil against 0.815; Vaelastrasz 1.008 for the Vigil against 0.996). On the merged build:

| Variant | Onyxia | Vaelastrasz | Onyxia 300 s |
|---|---|---|---|
| Holy Light 9, 8 and 7 (the first rotation) | 0.714 | 0.972 | 0.580 |
| **Holy Light 9 and 8** | 0.714 | **0.983** | **0.597** |
| and Holy Light only above 40, 50 or 60% mana | 0.714 | 0.983 | 0.596 to 0.597 |

Both presets take Holy Light 9 and 8 and the 50% mana line (one set of knobs: it costs the raid healer
nothing).

**Re-tuned 2026-11-12 (healer_fairness-20261104-5): the raid healer is a Flash of Light healer.** With no tank
in the raid role, the tank healer's Holy Light lines (target the tank) never fire, and its rotation in the raid
role out-carried the raid preset, whose Holy Light 9 and 8 and Divine Favor at the most hurt are too slow and
too big for the raid's spread damage. The raid preset now has no Holy Light or Divine Favor line: Lay on Hands
under 10%, Holy Shock under 50%, Flash of Light 6 under 35%, then Flash of Light 6 and 4 at the most hurt, each
once what's missing covers 70% of its Expected Healing. Healbench -role raid -carry, 1,000 iterations, 7 boss
files, mean carry on seed 101 (seed 202 where run), the raid preset's consumables (Lesser Mana Oil):

| Variant | Seed 101 | Seed 202 |
|---|---|---|
| The old raid preset (Holy Light 9 and 8, Divine Favor, Flash of Light 6 at 1.0x) | 1.400 | 1.402 |
| No Holy Light or Divine Favor, Flash of Light 6 at 0.7x (the tank preset's lines) | 1.424 | 1.430 |
| the same at 0.6x / 0.8x | 1.423 / 1.410 | |
| **Flash of Light 6 and 4 at 0.7x (the preset)** | **1.430** | **1.429** |
| Flash of Light 6, 5 and 4 / 6 and 5 / 6 and 3 / 4 alone | 1.430 / 1.422 / 1.417 / 1.392 | 1.430 / - / - / - |
| Flash of Light 6 and 4 at 0.6x / 0.8x | 1.420 / 1.423 | |
| and Holy Shock under 60% / 70% | 1.418 / 1.419 | |
| and Holy Shock at the most hurt when it misses what Holy Shock heals (on 6 at 0.7x) | 1.404 | |
| and no Holy Shock / no emergency Flash of Light / Flash of Light under 50% (on 6 at 0.7x) | 1.385 / 1.423 / 1.424 | |
| and Divine Favor once someone is under 35% / 50% | 1.424 / 1.423 | |
| and Concentration Aura (on 6 at 0.7x) / no Lay on Hands | 1.425 / 1.412 | |

Flash of Light 6, 5 and 4 ties 6 and 4 on both seeds; the preset binds the two. Seed 101 by file, old ->
new: Onyxia 0.977 -> 1.018, Broodlord 0.900 -> 0.912, Chromaggus 1.087 -> 1.146, Golemagg 2.684 -> 2.702,
Magmadar 1.903 -> 1.914, Ragnaros 1.290 -> 1.283, Vaelastrasz 0.957 -> 1.035. Mana is no limit: Onyxia at
300 s, 7,559 mana left at carry (7,330 before), carry 0.784 -> 0.799.

### Talents

The preset build, `53320015002101251-5-052330002`: Holy 31 (Spiritual Focus 5, Divine Strength 3, Healing
Light 3, Reverence 2, Divine Favor, Illumination 5, Infusion of Light 2, Oath of Light, Holy Shock, Boundless
Light 2, Holy Power 5, Light's Vigil), Divine Intellect 5, and Retribution 15 for Benediction 5 and
Conviction 3 (and 7 points that heal nothing: Improved Judgement 2, Sanctified Judgement 3 and Pursuit of
Justice 2). Each
other build moves the points to Divine Strength and Improved Seals, which heal nothing. Carry, 500
iterations:

| Build | Onyxia | Broodlord | Magmadar | Ragnaros | Onyxia 300 s | Onyxia raid | Vaelastrasz raid |
|---|---|---|---|---|---|---|---|
| **the preset** | 1.090 | 0.750 | 1.641 | 0.509 | 1.005 | 0.714 | 0.983 |
| no Infusion of Light | 1.083 | 0.710 | 1.604 | 0.484 | 0.987 | 0.700 | 0.986 |
| no Holy Power (and so no Light's Vigil) | 1.080 | 0.729 | 1.634 | 0.504 | 0.960 | 0.695 | 0.984 |
| no Illumination | 1.088 | 0.750 | 1.668 | 0.511 | 0.938 | 0.715 | 0.982 |
| no Spiritual Focus | 1.090 | 0.750 | 1.641 | 0.509 | 1.005 | 0.697 | 0.898 |

- Every Holy healing talent pays somewhere: Infusion of Light on the short tank stints (Holy Light 2 sec),
  Holy Power and Illumination once the fight is long enough for mana to count, Spiritual Focus in the raid
  role, where the healer takes hits (Vaelastrasz -0.085; the tank role never hits the healer, so it reads
  0 there).
- The Retribution side is for Benediction (10% off Holy Shock; Divine Favor and the Oath are outside its class mask) and Conviction's 3%
  crit, which feeds Illumination, Infusion of Light and Holy Power. A deeper Holy build has no healing
  talent left to buy: 31 points take every Holy healing talent, and Light's Vigil (the 31st) is only the
  capstone the presets don't cast (their Vigil variant was noise).

## The numbers (re-benched 2026-11-04)

The presets (`tools/healbench/req/paladin_holy.json`, `paladin_holy_raid.json`, the UI's
`ui/holy_paladin/apls/`), healbench `-carry`, 1,000 iterations, every boss file at its own length and
default carry, on branch `healers/fix-paladin` (master 0606686d3 plus this round). That master already
has the orchestrator's engine fixes that moved every healer: the raid damage share under one player
(healer_mechanics-20261104-1), school crit on direct heals (-7), the health clamp (-8) and the healbench
tank role keeping the raid as idle fillers (healer_fairness-20261104-1), so the earlier table against
the shaman is gone; the side-by-side is the H3 table in [healing_sim.md](healing_sim.md). EHPS is
effective healing a second; carry is the damage scale at the death line.

| Boss | Role | HPS | EHPS | Overheal | EH per mana | Mana left | Carry | Carry at a 1.0x gate |
|---|---|---|---|---|---|---|---|---|
| broodlord | tank | 640 | 487 | 23.9% | 5.13 | 6776 | **0.769** | 0.754 |
| chromaggus | tank | 783 | 628 | 19.9% | 5.86 | 3512 | **0.828** | 0.835 |
| golemagg | tank | 626 | 463 | 25.9% | 6.11 | 6993 | **1.701** | 1.648 |
| magmadar | tank | 588 | 427 | 27.5% | 6.23 | 7092 | **1.661** | 1.638 |
| onyxia | tank | 747 | 542 | 27.5% | 5.94 | 7043 | **1.106** | 1.097 |
| ragnaros | tank | 442 | 350 | 20.7% | 5.54 | 6630 | **0.517** | 0.510 |
| vaelastrasz | tank | 845 | 645 | 23.7% | 5.50 | 7305 | **0.721** | 0.711 |
| broodlord | raid | 396 | 313 | 21.0% | 4.30 | 7493 | **0.905** | |
| chromaggus | raid | 108 | 83 | 22.8% | 4.84 | 7884 | **1.081** | |
| golemagg | raid | 57 | 44 | 22.6% | 6.12 | 7971 | **2.699** | |
| magmadar | raid | 216 | 179 | 16.9% | 6.62 | 7885 | **1.918** | |
| onyxia | raid | 154 | 122 | 20.9% | 5.26 | 7667 | **0.984** | |
| ragnaros | raid | 371 | 296 | 20.2% | 6.02 | 7742 | **1.301** | |
| vaelastrasz | raid | 879 | 706 | 19.7% | 5.15 | 7079 | **1.002** | |

- **Tank healer: mean carry 1.043** (1.028 at the old 1.0x gate). No deaths below the carry on any file;
  out of mana only past 398 s, so it is never mana-bound at its carry.
- **Raid healer: mean carry 1.413.** Its floor on the others is 0 to 3.6%, none on the tanks. (2026-11-12, after the
  Flash of Light re-tune and the new consumables: 1.430, seed 101; the old rotation 1.400. The rows above are
  the old rotation's.)

### The gate (healer_fairness-20261104-2)

Every main heal at the tank (Holy Light 9 and 8, Divine Favor before Holy Light 9, Flash of Light 6 at the
most hurt) fired only when the tank missed the heal's whole Expected Healing. Scaling those gates, tank
role, 1,000 iterations, the 7 files' mean carry: 1.0x 1.028, 0.6x 1.037, **0.7x 1.043** (up on 6 of 7,
Chromaggus -0.007), 0.8x 1.038; Holy Light 9's gate alone at 0.4-0.8x 1.000-1.031 (nothing over noise);
Holy Light 8 and Flash of Light only at 0.6x 1.044. The tank preset takes 0.7x on all four. The raid
preset at 0.6 / 0.7 / 0.8x: 1.407 / 1.417 / 1.411 against 1.413, up on 4 of 7 at best, noise; it stays at
1.0x. Health-percent gates (Holy Shock under 50%, Flash of Light 6 under 35%, Lay on Hands under 10%) and
the mana gates are unchanged.

### This round's fixes (2026-11-04)

- healer_mechanics-20261104-4: Benediction no longer takes 10% off Divine Favor (60.48 mana) or Oath of
  Light (120.96); their class masks are empty. Holy Shock keeps it. Test `TestBenediction`.
- healer_mechanics-20261104-9: Libram of Holy Alacrity's buff read from the client (Items, call 9).
- healer_mechanics-20261104-10: the untested list is tested in `sim/paladin/holy/healing_more_test.go`
  (lower Holy Light / Flash of Light ranks incl. Holy Light 8, Holy Shock and Light's Vigil ranks and
  costs, Benediction, Divine Intellect, Holy Power on Holy Shock / Vigil heals and not on Holy Shock damage,
  Illumination on the Vigil and under Libram of Economy, Infusion of Light, Divine Favor on Holy Shock
  damage and off the Vigil, the Oath on downranked ranks, Lay on Hands ranks, Healing Light off the Vigil,
  Justice Armor, Libram of Infusion, Spiritual Focus pushback with Concentration Aura). No sim bug found.
  The holy package's tests need `-tags=with_db`.
- healer_fairness-20261104-2: the tank gate above.

## Calls for the user

Each is the default I used, with the alternatives' measured effect (tank healer, 500 iterations).

1. **Holy Power's reading.** The tooltip says "Increases the critical strike healing bonus of your Holy
   spells by 10/20/30/40/50%"; the client's aura is 50 (critical healing amount), which in the 1.12 and
   TBC servers multiplies the whole critical heal. Default: a crit heals 150% x 1.5 = 225% at 5/5 (the
   aura's way, and the racials' precedent in notes.md). Alternatives: the bonus +50 points (200%) or the
   bonus x1.5 (175%): Onyxia 300 s 0.897 and 0.896 against 0.922 (Holy Light 9 and 8, no mana line), the files' own
   lengths within 0.02.
2. **Oath of Light's bonus.** "Up to" 400 / 115 / 230 at 60, a flat amount on the heal: it crits with it,
   takes no healing power, and a rank downranking cuts loses the same share of it (`core.DownrankFactor`).
   Alternative: the bonus scaled by the heal's coefficient, like healing power: Onyxia 300 s 0.869 against
   0.922, -0.01 at the files' own lengths.
3. **Illumination on Light's Vigil**: once a detonation (not once a healed player), 50% of the Vigil's base
   cost. No preset casts the Vigil.
4. **Light's Vigil's refund**: the cast spell's second effect still stores 75 (the old 75% refund) but the
   hotfix's tooltip has none. Default: no refund. No preset casts the Vigil.
5. **The gear lists**: the cached Forever Holy guide's gear planner holds beta item ids no database has, so
   the synthetic set is priced from the Classic pre-raid list with the shaman's spell power and healing
   split (`orchestration/gear_targets.md`).
6. **The 7 free points.** Holy 31, Divine Intellect 5, Benediction 5 and Conviction 3 leave 7 points, which
   the preset puts in Improved Judgement, Sanctified Judgement and Pursuit of Justice. They heal nothing, and
   neither does any other place for them, so they're the user's to move (run speed, say).
7. **Champion of the Light** adds spell power (heals too) although its aura (174) reads as spell damage
   only; no healer build reaches it (25 Retribution points), so no number moves.
8. **The raid sim's placeholder talents** (`StandardTalents` in presets.ts, used by the raid sim's paladin
   slot) are unchanged: whether the raid sim's paladin heals is the raid sim's call.
9. **Libram of Holy Alacrity** (228175): answered from the client (healer_mechanics-20261104-9): buff 449982
   takes 0.2 sec off the next Holy Light within 10 sec, 1 charge, triggered (server side) by Holy Shock. At one
   Holy Shock every 15 sec that is at most 0.2 sec a 15 sec, about 1.3% more casting time, against the 6% Holy
   Shock crit of Libram of Infusion or the mana of Libram of Economy; not added. Whether to add it is open.
