# Mage

Source: wowforevertalents.com dataset (BlizzCon 2026 footage overlaid on Classic Era 1.15.9 client tables), data version 2026-09-13/14. See [../README.md](../README.md) for caveats. Numbers on new talents were read off level 38 demo characters, so absolute damage and mana values are not level 60 values.

Talents: 54 total, 6 new, 32 changed, 16 unchanged (verified).

Row N needs 5*(N-1) points spent in that tree. Row 1 is the top row. Rows 3, 4, 5 and 7 hold the 11, 16, 21 and 31 point talents.

Warning on multi-rank NEW and CHANGED talents: the footage showed rank 1 only. The dataset scaled the other ranks linearly from rank 1, which sometimes produces nonsense (for example Maelstrom Weapon rank 5 reading 'stacks up to 25 times'). Treat rank 1 as observed and the rest as a guess until the beta client is out. The [ranks 2+ settled by the 70338 trait curves] tag marks every multi-rank new or changed talent. Read the Note line, it says when a higher rank was actually seen on a BlizzCon slide.

## Client build 1.60.1.70707 (2026-10-30)

70707 puts the 10-27 to 10-29 hotfixes' mage rework into the client tables (foreverchanges.pro "Build 1.60.1.70707"; ledger MAGE-U-20 to 27, finding client_update-20261030-1). It supersedes the sections below wherever they disagree.

- Arcane Focus is gone (TraitNode 105814). The Arcane tree has 18 talents: Wand Specialization, Improved Channeling, Arcane Resilience, Arcane Subtlety, Magic Absorption, Arcane Concentration, Missile Barrage, Arcane Geometry, Arcane Impact, Arcane Blast, Arcane Shielding, Improved Counterspell, Arcane Meditation, Arcane Wrath, Presence of Mind, Arcane Mind, Arcane Instability, Arcane Power. Missile Barrage and Arcane Resilience swapped spots, Improved Fire Blast went from tier 1 to tier 3 and Improved Channeling moved left, so every mage talent string reads differently.
- Elemental Precision is Precision: 2 / 4 / 6% hit with all spells (29438 effect 0, aura 107 on a mask became aura 55). The main set's 11% makes 17% for every school, the Arcane set's 10% makes 16%, the cap.
- Improved Fireball is Accelerant: 0.1 sec a point off Fireball and Frostfire Bolt as before, and now off Arcane Wrath too (11069 effect 0 gains Arcane Wrath's class mask bit).
- Incineration has 5 points (2% a point, 10% crit on Fire Blast, Scorch, Ice Lance and Arcane Blast) and Critical Mass 3 (2 / 4 / 6% Fire crit, was 10%).
- Ice Lance has an 8 sec cooldown (category 19). The client puts it on rank 1 only (SpellCooldowns 101119, 1312002); ranks 2 to 6 have no cooldown row. The sim gives every rank the cooldown, as the tooltip and foreverchanges.pro have it (user ruling 2026-10-30, report assumption `mage_ice_lance_cd`). Check rank 6 in game at launch.
- Missile Barrage's "missiles fire every 0.5 sec" effect (400589 effect 1) became a dummy aura; the -50% duration and -100% cost effects stay and the tooltip is unchanged. The sim keeps five missiles in half the time, since a dummy usually means a server script does the job (finding client_update-20261030-2: check an Arcane combat log).
- Frost Blast left the spellbook (the sim never cast it).

The presets, re-picked by talent searches at 180 sec, confirmed at 20,000 iterations on seeds 1 and 1000003 (`ui/mage/presets.ts` has the alternatives):

| Build | Talents | 120 s | 180 s | 300 s | 180 s before |
|---|---|---|---|---|---|
| Fire (default) | `0502052-55500201230133031-003` | 1071.7 | 1061.0 | 1044.0 | 1101.0 |
| Frostfire | `-2550030103013303-0235000311203002` | 1063.5 | 1051.0 | 1030.8 | 1096.4 |
| Arcane (Gnome) | `050205203100301531-5-05350002` | 926.1 | 915.5 | 895.4 | 901.6 |
| Frost | `05020520310011053-1-0535000301003002` | 911.2 | 908.3 | 886.6 | 946.6 |

- Fire: Critical Mass's two points are Incineration 5. With Arcane Focus gone, Arcane row 1 needs five other points to reach Arcane Concentration: Improved Channeling 5, one of them Burning Soul's third point.
- Frostfire: Critical Mass's two points are Incineration 2 and Master of Elements 3. The loss is Critical Mass and Ice Lance's cooldown; the rotation knob search kept every knob.
- Arcane gains: Precision's 6% covers Arcane now (Arcane Focus gave 5%), and the five points are Incineration 5 (Arcane Blast) and Frost Channeling 2. Its rotation no longer drops to one Arcane Blast stack below 20% mana (+1.6 at 120 / 180 / 300 sec together).
- Frost: Arcane Focus's five points and Presence of Mind's are Arcane Concentration 5, Arcane Meditation, Arcane Wrath and Incineration. Ice Lance on every Fingers of Frost charge was this build's damage, so the cooldown costs it most. With mana to spare the rotation keeps three Arcane Blast stacks and drops Evocation (+4.8 at 120 / 180 / 300 sec together).

## Regear and rotation round (2026-10-23)

After the wowsims/forever alignment round (orchestration/alignment_round_2026-10-23.md). Numbers at 180 sec, 20,000 iterations, seed 1 (seed 1000003 within 1.1); the full tables are in [../sim-status.md](../sim-status.md#mage).

- Spell hit: the 16% cap against a level 63 boss is unchanged (the 1% spell miss floor). The main set (Fire, Frostfire, Frost) sat a point over it since the hotfix made Elemental Precision 6%; its head gave 10 hit rating to 3 Intellect and 8 spell power (Fire +5.3, Frostfire +7.3, Frost +2.1). The Arcane set keeps 10% for Arcane Focus.
- Frost talents: `055220200310000153--0535000301003002`, Frost to Fingers of Frost (22) and 29 Arcane with Arcane Mind and Arcane Instability, two Arcane Blast stacks in the rotation: 946.6, against 822.8 for the old deep-Frost preset on the same gear. Every seed that kept Ice Barrier stopped near 845, wowsims/forever's string at range included.
- Ice Lance with the charge spent a spell batch late: a bolt that completes with a Fingers of Frost charge up and the Ice Lance right after it both read Frozen. Frost's Ice Lance on every charge already does that when the charge comes during a Frostbolt's cast; waiting for a bolt before every Ice Lance lost 11 to 23. Frostfire's knob search put Ice Lance on the last charge only, Evocation below 10% with 90 sec left and rank 2 below 30 mana a second: 1052.2 to 1096.4.
- Every rotation ends with the wand when out of mana. Fire (1101.0) is the mage's default build, 3.4 to 4.6 ahead of Frostfire on both seed sets.
- Races and consumables are unchanged: Human (High Order on Frost), Orc the best Horde race; Flask of Supreme Power beats every natural flask. Since the 2026-10-27 race re-pick (orchestration/standards.md items 29 and 16) the Arcane preset is Gnome (2.6 ahead of the Human at 180 sec).

## Server hotfix to 1.60.1.70450 (2026-10-20)

A Frost hotfix the client tables don't carry (70532 still has the old tree). It is in wago.tools' hotfix table (push of 2026-10-20 09:51, build 70450) and on foreverchanges.pro's 2026-10-20 hotfix list, and wowsims/forever applied it in 983ea6fcc1 ("mage hotfix and talent shift"). `tools/forever_talents/client/mage.json` carries it by hand until a client build does. It supersedes the notes below wherever they disagree.

- Piercing Ice (11151, 2 / 4 / 6% Frost damage) is gone (TraitNode 105773 deleted). The Frost tree has 18 talents.
- The tree moved: Frostbite is now row 2 column 2 and Permafrost row 2 column 4 (they swapped), Frost Channeling row 3 column 1 and Improved Blizzard row 3 column 2 (TraitNode 105774, 105776, 105772, 105771). Ice Lance stays at row 3 column 3.
- Elemental Precision: 3 ranks of 2% (TraitNodeEntry 130507, curve 2 / 4 / 6), 6% hit at most (5 ranks of 1% before). With the synthetic gear's 11% that is 17%, a point over the 16% cap against a level 63 boss; 2 ranks are 1% short.
- Ice Barrier (11426 to 13033): "Shield yourself, absorbing $s1 + 10% of your Frost spell power (raised by Frost Warding) damage and increasing your Frost damage by 3% for 1 min. While shielded, your Frost damage is increased by an additional 3% and you do not lose casting time when you take damage." Every cast gives Frigid Power (1326891), 3% more Frost damage for 60 sec however soon the shield breaks, and the shield's own effects 2 and 3 add 3% while it holds. Both take Piercing Ice's old mask, every Frost spell Frostfire Bolt included. The absorb's base points did not change (rank 4 is 819 at 60), so it holds a 1500 raid hit for 819 + 10% of Frost spell power, times 1.2 with Frost Warding 2: the first raid hit breaks it.
- Frost Warding (11189): its Frost Ward chance negates the spell instead of reflecting it, and a new effect 3 raises the Ice Barrier and Frost Ward absorbs by 10 / 20%. Frost Ward's absorb gained 10% of Frost spell power (the sim has no Frost Ward).
- Also in the same push (push 112694): Winter's Chill's tooltip (11180) now reads "Gives your Frostbolt, Frost Blast, and Fingers of Frost empowered Ice Lance spells a $m2% chance to apply Winter's Chill, increasing the target's damage taken from your next Cone of Cold by $12579s1% and it gains an additional $s3% of your spell damage, stacking up to $s1 times", with new effects on 11180 and 12579: SpellEffect 688253 (aura 271, Cone of Cold mask 0x200) is 5, and the dummies 1364854 and 1364855 are 7. So each stack is 5% more damage taken from the mage's Cone of Cold and 7% of spell power on its coefficient, up to 5 stacks. Frost Blast's coefficient became 27.1% in the same push. The sim models both Winter's Chill and Cone of Cold since 2026-10-23 (below).

The sim (`sim/mage/ice_barrier.go`, `talents.go`): Ice Barrier is a real absorb on the mage now, and the encounter's raid hits are the only damage a DPS mage takes, so a hit breaks it and takes its 3% with it, while Frigid Power stays. A hit the shield takes whole does no damage and pushes back nothing; the hit that breaks it pushes back as any other. The two 3% bonuses are each their own damage factor (they added to the others, as Piercing Ice did, until 2026-10-23). The Frost rotation casts Ice Barrier before the pull and again when Frigid Power has under 2 sec left (821.7 at 180 sec, against 813.4 recasting whenever the shield is down and 812.4 never casting it).

Talent strings: the Frost tree reads Frost Warding, Improved Frostbolt, Elemental Precision, Ice Shards, Frostbite, Improved Frost Nova, Permafrost, Frost Channeling, Improved Blizzard, Ice Lance, Arctic Reach, Ice Block, Shatter, Improved Cone of Cold, Cold Snap, Fingers of Frost, Winter's Chill, Ice Barrier, so every saved string with Frost points past Ice Shards reads differently. The presets moved to Frostfire `-0155003013012503-0235000311203002` (1081.9 → 1041.5 at 180 sec), Frost `05020050031001-03-253500133101301201` (843.9 → 820.8), Fire `0402205-23550030130135031-003` (1067.6 → 1094.3) and Arcane `0552205003100301531-02-0535` (964.7 → 981.9). Frostfire and Frost need the five freed points in Frost to keep Fingers of Frost and Ice Barrier, on talents that do no damage; Fire and Arcane put them in Arcane Subtlety (and Improved Channeling), which is why they gained. See `ui/mage/presets.ts` for the alternatives simmed.

## Sim changes of 2026-10-23 (wowsims/forever evidence)

- Percent damage and cost modifiers multiply under Forever, each its own factor, as every aura 108 SPELLMOD_DAMAGE, SPELLMOD_DOT and SPELLMOD_COST row does in game (wowsims/forever fef2bcaebd, 810101380a: Improved Wrath with Moonglow leaves a 120-mana Wrath at 45, not 30). Arcane Power with Arcane Instability 3 is 1.339, four Arcane Blast stacks under Arcane Power cost 10.4 times the base, and a cost row of -100% or lower (Missile Barrage) makes the spell free whatever else is up. Frost Channeling, Ice Barrier, Frigid Power and the Gnome's Eureka! multiply too.
- Frost Channeling's cost row (11160 effect 0, mask 0xA0000 and word 1 0x1000) covers Frostfire Bolt (word 1 bit 12) and Ice Barrier.
- Fingers of Frost (wowsims/forever 1373a0b96f, ea90ab9635, from their logs): Shatter's crit reaches every mage damage spell, Blizzard's ticks included. A damage spell the rotation casts spends a charge 10 ms (a spell batch) after it finishes, so a spell in the same batch still reads Frozen; a channel spends none at its cast, and each Arcane Missiles missile that lands spends one.
- Master of Elements refunds its share of the base cost on any Fire or Frost crit, a Clearcasting cast's too (wowsims/forever c999a66b3f).
- Cone of Cold is simulated (`sim/mage/cone_of_cold.go`): ranks 1 to 5 from the client (rank 5 is 328 to 358 at 60, 12.9% coefficient, 555 mana, instant, 10 sec cooldown), at every target in front within its 10 yards (Arctic Reach adds 10% a point), and the mage can only cast it with the boss that close. It is binary and a Chill, so each hit rolls Fingers of Frost, and it takes Improved Cone of Cold (12 / 23 / 35%), Ice Shards, Elemental Precision, Frost Channeling, Arcane Power, Arcane Instability, Clearcasting, Ice Barrier and Frigid Power by their masks (not Combustion); Cold Snap resets it. Winter's Chill stacks come from landed Frostbolts and from Ice Lances cast at a Frozen target, and each Cone of Cold spends a target's stacks (wowsims/forever talents_frost.go, cone_of_cold.go). The Frost preset stands at 25 yards and never casts it; a Frost mage at 10 yards casting it on cooldown is the report assumption `mage_cone_of_cold` (orchestration/spec_assumptions.md).
- Ice Lance is not binary under Forever: the 1240047 row has no slow, and wowsims/forever's rolls partial resists.
- Fingers of Frost: a cast is exempt from spending the charges only when it granted them from inside its own effects (a Cone of Cold hit), not by timestamp (wowsims/forever e875ae5bd3).

## Beta client 1.60.1.70425 (2026-10-16)

The 2026-10-15 mage rework, read from the 70425 client tables (wago.tools SpellEffect, SpellLevels, SpellMisc, SpellPower, SpellCooldowns, SpellAuraOptions and the Trait tables of tree 1112) and wowhead's Forever tooltips. It supersedes the 70338 notes below wherever they disagree. `tools/forever_talents/client/mage.json` carries the new tree, and the talent strings changed with it (see the end of this section).

- Arcane Blast: 1.5 sec cast (2.5 before), 42.9% coefficient (71.4%). Rank 5 deals 218 to 254 at 60 (base 236, 15.1% variance, +1.6 a level to 68). Ranks 1 to 4 run 33 to 39, 78 to 91, 101 to 117 and 165 to 192 at their caps. It still costs 15% of base mana, plus 175% per stack. Each stack (400573, up to 4) is now 15% of Intellect as spell damage on every school (aura 174), for 15 sec from the first stack. More stacks don't refresh it, and other spells no longer end it (no charges, no proc flags). The 10% damage bonus on the other spells is gone.
- Arcane Wrath (new talent, Arcane row 4 column 4, needs Arcane Blast; ranks 1326079 to 1326082 at 25, 36, 48 and 60): rank 4 is 451 base (20.1% variance, so 406 to 496 at 60, +3 a level to 71), 100% coefficient, plus 5% of maximum mana as spell damage. 3.5 sec cast, 6 sec cooldown, 395 mana. The masks of Arcane Focus, Impact, Mind, Power, Instability, Improved Channeling and Geometry now include it.
- Missile Barrage: 2 ranks, moved to Arcane row 1 column 4. The trait curve is 20 / 40. At rank 2, Arcane Blast has a 40% chance, each Arcane Missiles hit 10%, and Arcane Wrath, Fireball, Frostbolt and Frostfire Bolt 20%. Rank 1 is half of each. ProcChance is 100 now, so the tooltip's numbers stand.
- Improved Fire Blast (11078, was Wake of Fire): 1 / 2 sec off Fire Blast's cooldown and 8 / 15 yds of range. The crit after a kill is gone.
- Critical Mass: 5 ranks, 2% Fire crit each (3 before).
- Fire Power: 3 ranks. Each rank gives 1% more Fire damage per Fire effect of the mage's on the target, up to 4 effects (3% each and at most 12% at rank 3). Effects that mix schools don't count. The sim counts Fire Vulnerability and the DoTs of Fireball, Pyroblast, Flamestrike and Ignite, and multiplies the bonus with Fire Vulnerability on the mage's attack table. Ignite takes no bonus itself, because its share of the crit already has it. Frostfire Bolt, being a Fire spell, does take the bonus. Its DoT doesn't count as an effect. Whether Frostfire Bolt's own damage qualifies is an assumption: the tooltip doesn't say.
- Winter's Chill (12579): a Cone of Cold debuff now, up to 5 stacks, with a 20% chance per rank, and no crit. The 70425 client had 40% Cone of Cold damage taken a stack; the 2026-10-20 hotfix made it 5% and 7% of spell power a stack (above). Frostbite now freezes on 15% of chills, which a raid boss is immune to.
- Ice Lance: 200% more damage on a Frozen target at every rank (300% before), so 3 times the damage.
- Frost Channeling: the threat effect is now −10 / −20 / −30 (the mana part is unchanged).
- Heating Up: every Fire Blast that deals damage adds a stack, two on a crit (the 400624 tooltip: "Fire Blast grants an additional stack every time it deals damage"), and a Pyroblast that Presence of Mind made instant no longer uses the stacks (development notes). Both are server side; the buff 400625 only changed its proc flags.
- Pushback, now that the presets take 2 raid hits a minute: Burning Soul (11083 effect 0) gives Fireball, Scorch, Pyroblast, Flamestrike and Frostfire Bolt a 23 / 47 / 70% chance not to lose casting time, and Improved Channeling (11237) 20% a point on Arcane Missiles (effect 0) and 14% a point on Arcane Blast and Arcane Wrath (effect 1, whose mask gained Arcane Wrath). Frost has no pushback talent.
- Frost Blast (new trainer spell, 1326053 to 1326055 at 24, 36 and 48): rank 3 is 80 to 89 at 60, 27.1% coefficient since the 2026-10-20 hotfix (push 112694; 34.3% in 70425), 1.5 sec cast, 90 mana, and a 40% slow for 9 sec. Its damage per second of cast is under half of Frostbolt's, so the sim leaves it out.

Talent strings: Missile Barrage now sits at Arcane index 3 and Arcane Wrath at index 14, so the Arcane tree has 19 talents and any saved string with Arcane points reads differently. Fire Power caps at 3, so the old Fire and Frostfire strings (Fire Power 5 and 4) are invalid. Combustion still needs Critical Mass maxed, now at 5 points.

## Beta client 1.60.1.69876 to 1.60.1.70338 (2026-10-12)

The beta client's mage tree, read through [hyjal.cc](https://hyjal.cc/updates/1.60.1.69876) (1.60.1.69876), with the talents later builds changed brought up to 1.60.1.70338 from [foreverchanges.pro](https://foreverchanges.pro/talents/mage). Positions, rank counts and prerequisites match the BlizzCon data below. Ranks 2 and up are real numbers now, and `tools/forever_talents/data/mage.json` in the sim repo is regenerated from the client tree (`client/mage.json`), so the per-talent text below is superseded where the two differ. The spells are from the client's spell tables (wago.tools DB2, 1.60.1.70338).

Every multi-rank talent's per-rank values were checked on 2026-10-12 against the 70338 client's trait curves (wago.tools TraitDefinition, TraitDefinitionEffectPoints and CurvePoint), which override the rank 1 spell's base points. They all match the client tree's text and the sim, so the [ranks 2+ unverified] tags below are settled. Some rank 1 spells keep Classic leftovers in SpellEffect (Arcane Subtlety stores -30 / -15, Arcane Concentration 100, Wake of Fire -2000), which the curves replace.

Talents (client ranks where the BlizzCon data guessed):

- Arcane Focus and Elemental Precision are 1% hit per rank to 5% (Elemental Precision is 3 ranks of 2% since the 2026-10-20 hotfix). Arcane Impact is 2 / 4 / 6% crit on all Arcane spells. Arcane Subtlety is 8 / 15 resistance. Magic Absorption is 5 / 10 resistance.
- Arcane Meditation is 17 / 33 / 50% casting regen (Classic 5 / 10 / 15%). Arcane Resilience is 25 / 50% of Intellect as armor.
- Arcane Mind is 2% Intellect and 20% Arcane crit damage bonus per rank (Classic gave max mana). Arcane Instability is 1% damage and 1% crit per rank on every spell.
- Arcane Concentration has a 1 sec internal cooldown in the client (11213), so one Arcane Missiles can't roll it five times.
- Improved Channeling (new) is 20% Arcane Missiles and 14% Arcane Blast pushback protection per rank. Arcane Geometry (new) is 3 / 6 yds of Arcane range.
- Arcane Blast (talent, then trainer ranks at 30, 40, 50 and 60): 364 to 424 at 60 with a 71.4% coefficient, 2.5 sec, 15% of base mana. Each cast is a stack (8 sec, 4 stacks) worth 10% more damage on the mage's other spells and 175% more mana on Arcane Blast. Any other damage spell ends them.
- Missile Barrage (new): Arcane Blast 40%, Fireball, Frostbolt and Frostfire Bolt 20%, to make the next Arcane Missiles within 15 sec free and twice as fast.
- Wake of Fire (new): Fire Blast 1 / 2 sec shorter, and 25 / 50% crit on the next Fire Blast after a kill (30 sec since 2026-10-08, 20 before).
- Incineration has 3 ranks (2 / 4 / 6%) and covers Fire Blast, Scorch, Ice Lance and Arcane Blast. Improved Fireball covers Frostfire Bolt. Impact is 3 / 7 / 10%, Burning Soul 23 / 47 / 70% with 10% less threat per rank.
- Improved Scorch is 33 / 67 / 100%, and the vulnerability (22959) counts only for the mage who put it up ("all Fire damage you deal to your target").
- Heating Up (Hot Streak until 2026-10-08, new): non-periodic crits with Fireball, Frostfire Bolt, Fire Blast and Scorch take 25% off the next Pyroblast's cast within 20 sec (15 sec before 2026-10-08), stacking to 3. The buff (400625) has one charge and a -25% cast time modifier per stack.
- Combustion: 3 non-periodic Fire crits again since 2026-10-08 (4 on the beta before).
- Ignite: "no longer counts damage bonuses twice" since 2026-09-24.
- Permafrost is 11 / 22 / 33% longer chills, Improved Blizzard 15 / 25 / 40% slower, Improved Cone of Cold 12 / 23 / 35%, Frost Channeling 5% cheaper Frost spells and 10% less threat per rank.
- Shatter has 3 ranks: 17 / 33 / 50% crit against Frozen targets.
- Ice Lance (talent, then trainer ranks at 28, 34, 42, 48 and 56): 136 to 161 at 60, instant, 160 mana, 300% more damage to Frozen targets. The client has no spell power coefficient on it (nor did SoD's, whose server gave it 42.9%). The sim uses SoD's 42.9% by the user's ruling on the report assumptions (2026-10-15).
- Fingers of Frost (new): chills have a 15% chance to make the next spell (two at rank 2) act as if the target were Frozen, for 15 sec.
- Winter's Chill: 20% per rank, at most one stack per rank, and each stack is 2% crit for the mage's own Frostbolt and Ice Lance only (12579's crit aura is from the caster).
- Ice Barrier rank 1 is 448 at 46 and rank 4 819 at 60.

Spells at level 60 (client, top rank):

- Fireball 425 to 541 plus 60 over 8 sec (Classic 596 to 760 plus 76), 100% coefficient, 410 mana.
- Frostbolt 457 to 493 (Classic 515 to 555), 81.4%, 290 mana.
- Frostfire Bolt (new trainer spell at 40, 50 and 60): Fire and Frost, 270 to 314 plus 57 over 9 sec, 81.4%, 3 sec, 370 mana, chills.
- Arcane Missiles 209 a missile (Classic 230), 28.6% a missile (143% for the channel, Classic 120%), 655 mana.
- Those three are the Ahn'Qiraj book ranks (Fireball XII, Frostbolt XI, Arcane Missiles VIII). The 70338 client has the books with no source and Forever starts before Molten Core, so the sim's Forever ruleset stops at the ranks before them: Fireball XI (397 to 505), Frostbolt X (382 to 413) and Arcane Missiles VII (175 a missile), at their caps. The sim deals 4% more with Frostbolt X (398 to 429), half of the 8% beta build 70291 gave the level 20 and 26 ranks (the user's ruling, 2026-10-15).
- Scorch 166 to 196 (Classic 237 to 280), Fire Blast 415 to 491 (Classic 446 to 524), both 42.9% on every rank.
- Pyroblast 520 to 646 plus 212 over 12 sec (Classic 716 to 890 plus 268). Rank 1 comes with the talent at 20, ranks 2 to 8 from the trainer.
- Arcane Explosion 239 to 258, Flamestrike 381 to 466 plus 332, Blizzard 1168 over 8 sec (3% a tick), Blast Wave 453 to 533, Cone of Cold 328 to 358.
- Mage Armor gives 50% casting regen at every rank (Classic 30%). Evocation and the mana gems are unchanged.
- The low ranks of the damage spells have close to full coefficients, not the 1.12 penalty for spells learned below 20.

Manaflare Regalia (280450 to 280455, item level 65, no stats in the client yet) is Forever's level 60 mage set: 2 pieces 1% hit, 3 Counterspell 5 sec shorter, 4 up to 21 spell damage against Elementals, 5 Frostfire Bolt gets 10% more chance at Missile Barrage and Fingers of Frost and 10% crit during Combustion.

## Mechanics check against the sim (2026-10-13)

Every mage spell and talent the sim models was compared again with the 70338 client (wago.tools DB2: SpellEffect, SpellMisc, SpellPower, SpellCooldowns, SpellCategories, SpellAuraOptions, Trait tables), wowhead's Forever tooltips (`nether.wowhead.com/forever/tooltip/spell/<id>?dataEnv=16`), Blizzard's Forever development notes (us.forums.blizzard.com/en/wow/t/2360696) and mage deep dive (worldofwarcraft.blizzard.com/news/24302097), and foreverchanges.pro/class/mage.

Fixed in the sim (no change to the single target presets):
- Blizzard's ticks are the area trigger's own spells (1279976 to 1279980, 1279949), each with a 4.2% coefficient (SpellEffect EffectBonusCoefficient 0.042). The 3% the sim used is Blizzard's dummy effect (10187 effect 0).
- Flamestrike's DoT ticks (1279983 to 1279990) are 3.2% each on every rank (the sim had 1.7 to 2%).
- Combustion (28682) stacks to 10 (SpellAuraOptions CumulativeAura), not 20.

Confirmed as the sim has them:
- Evocation 12051: 8 sec, 1500% regen, 8 min cooldown, the same as Era.
- Mage Armor 6117 / 22782 / 22783: 50% casting regen at every rank (Era 30%), resistances 5 / 10 / 15. The deep dive: "Mana regeneration bonus from Spirit increased".
- Mana gems (Agate 375 to 425, Jade 550 to 650, Citrine 775 to 925, Ruby 1000 to 1200) and the Demonic and Dark Runes (900 to 1500) share one 2 min cooldown: ItemEffect SpellCategoryID 1153, 120000 ms on all six.
- Arcane Intellect and Arcane Brilliance are still +31 Intellect. They now last an hour, and Brilliance reaches the raid (ImplicitTarget 56).
- No Forever source changes the Spirit regen formula, the five second rule, base mana or the Intellect conversions, so the sim keeps Era's.
- Frostfire Bolt "always uses the higher of the Mage's spell damage bonuses to Fire or Frost damage" (deep dive). The sim's multi-school spells already take the higher school power.
- Every damage rank, coefficient, cost and cast time of Fireball, Frostbolt, Frostfire Bolt, Arcane Blast, Arcane Missiles, Ice Lance, Pyroblast, Scorch, Fire Blast, Arcane Explosion and Blast Wave, and every talent value, matches the client, but for two report assumptions the user ruled on (2026-10-15): Ice Lance's 42.9% coefficient and Frostbolt X's 4%.

Wands: the 2026-09-24 development notes say "Wands no longer gain any additional damage from the user's spell damage". Shoot (5019) and Wand Specialization (13 / 25%, now 2 points) are otherwise Era's, and wand damage still comes from ItemDamageWand. Since 2026-10-23 the sim models Shoot and Wand Specialization (+13 / 25% on magic wands, the rows' school mask 126, so a physical wand gets nothing), as wowsims/forever has them (sim-status.md, Wand Shoot). The Frostfire preset wands when nothing else can be cast. Bonecreeper Stylus (13938) has no ItemSparse row in 70338, so the synthetic wand borrows its damage, speed and school (Arcane, Classic's DamageType 6).

Open, with what they are worth (notes.md):
- Arcane Blast's stacks on Arcane Missiles: gone with 70425's redesign (the stacks are spell damage now).
- 400573's periodic mask: gone in 70425 (the stacks are spell damage). Arcane Power's still leaves Frostfire Bolt's DoT out; the sim gives it to every DoT, and no preset has both.
- Presence of Mind and Combustion share the client's category 1151 cooldown (3 min). A level 60 mage can't have both, so the sim keeps them apart.
- Missile Barrage's talent spell (400588): ProcChance 100 since 70425 (50 before), so the tooltip's numbers stand.

## Spellbook changes (trainer abilities)

0 of 48 trainer abilities confirmed from footage. Abilities not listed below are either verified unchanged or still Classic placeholders (unverified).

### Not yet verified (Classic text assumed)

Frost Armor, Frostbolt, Frost Nova, Blizzard, Frost Ward, Cone of Cold, Ice Armor, Ice Barrier, Fireball, Fire Blast, Flamestrike, Fire Ward, Scorch, Pyroblast, Blast Wave, Arcane Intellect, Conjure Water, Conjure Food, Arcane Missiles, Polymorph, Dampen Magic, Slow Fall, Arcane Explosion, Detect Magic, Amplify Magic, Remove Lesser Curse, Blink, Evocation, Mana Shield, Teleport: Ironforge, Teleport: Orgrimmar, Teleport: Stormwind, Teleport: Undercity, Counterspell, Conjure Mana Agate, Teleport: Darnassus, Teleport: Thunder Bluff, Mage Armor, Conjure Mana Jade, Portal: Ironforge, Portal: Orgrimmar, Portal: Stormwind, Portal: Undercity, Conjure Mana Citrine, Portal: Darnassus, Portal: Thunder Bluff, Arcane Brilliance, Conjure Mana Ruby

## Talents

### Arcane

18 talents, 3 new. 31-point talent: Arcane Power.

#### Row 1 (0 points)

- **Wand Specialization** (2 ranks, column 1) [unchanged]
  - Rank 1: Increases your damage with Wands by 13%.
  - Rank 2: Increases your damage with Wands by 25%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Arcane Focus** (5 ranks, column 2) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases your chance to hit with your Arcane spells by 1%.
  - Rank 2: Increases your chance to hit with your Arcane spells by 2%.
  - Rank 3: Increases your chance to hit with your Arcane spells by 3%.
  - Rank 4: Increases your chance to hit with your Arcane spells by 4%.
  - Rank 5: Increases your chance to hit with your Arcane spells by 5%.
  - Classic (5 ranks):
    - Rank 1: Reduces the chance that the opponent can resist your Arcane spells by 2%.
    - Rank 2: Reduces the chance that the opponent can resist your Arcane spells by 4%.
    - Rank 3: Reduces the chance that the opponent can resist your Arcane spells by 6%.
    - Rank 4: Reduces the chance that the opponent can resist your Arcane spells by 8%.
    - Rank 5: Reduces the chance that the opponent can resist your Arcane spells by 10%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Improved Channeling** (5 ranks, column 3) [NEW] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Gives you a 20% chance to avoid interruption caused by damage while channeling Arcane Missiles and a 14% chance while casting Arcane Blast.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Rank 3: Rank 3 values not yet captured from footage.
  - Rank 4: Rank 4 values not yet captured from footage.
  - Rank 5: Rank 5 values not yet captured from footage.

#### Row 2 (5 points)

- **Arcane Subtlety** (2 ranks, column 1) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Reduces your target's resistance to all your spells by 8 and reduces the threat caused by your Arcane spells by 15%.
  - Rank 2: Reduces your target's resistance to all your spells by 15 and reduces the threat caused by your Arcane spells by 30%.
  - Classic (2 ranks):
    - Rank 1: Reduces your target's resistance to all your spells by 5 and reduces the threat caused by your Arcane spells by 20%.
    - Rank 2: Reduces your target's resistance to all your spells by 10 and reduces the threat caused by your Arcane spells by 40%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Magic Absorption** (2 ranks, column 2) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases all your resistances by 5 and causes all spells you fully resist to restore 1% of your total mana. Cannot trigger more often than 1 time per sec.
  - Rank 2: Increases all your resistances by 10 and causes all spells you fully resist to restore 2% of your total mana. Cannot trigger more often than 1 time per sec.
  - Classic (5 ranks):
    - Rank 1: Increases all resistances by 2 and causes all spells you fully resist to restore 1% of your total mana. 1 sec. cooldown.
    - Rank 2: Increases all resistances by 4 and causes all spells you fully resist to restore 2% of your total mana. 1 sec. cooldown.
    - Rank 3: Increases all resistances by 6 and causes all spells you fully resist to restore 3% of your total mana. 1 sec. cooldown.
    - Rank 4: Increases all resistances by 8 and causes all spells you fully resist to restore 4% of your total mana. 1 sec. cooldown.
    - Rank 5: Increases all resistances by 10 and causes all spells you fully resist to restore 5% of your total mana. 1 sec. cooldown.
- **Arcane Concentration** (5 ranks, column 3) [unchanged]
  - Rank 1: Gives you a 2% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
  - Rank 2: Gives you a 4% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
  - Rank 3: Gives you a 6% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
  - Rank 4: Gives you a 8% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
  - Rank 5: Gives you a 10% chance of entering a Clearcasting state after any damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage spell by 100%.
- **Arcane Resilience** (2 ranks, column 4) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases your Armor by an amount equal to 25% of your Intellect.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Classic: Increases your armor by an amount equal to 50% of your Intellect.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Arcane Geometry** (2 ranks, column 1) [NEW] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases the range of your Arcane spells by 3 yards.
  - Rank 2: Rank 2 values not yet captured from footage.
- **Arcane Impact** (3 ranks, column 2) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases the critical strike chance of your Arcane spells by 2%.
  - Rank 2: Increases the critical strike chance of your Arcane spells by 4%.
  - Rank 3: Increases the critical strike chance of your Arcane spells by 6%.
  - Classic (3 ranks):
    - Rank 1: Increases the critical strike chance of your Arcane Explosion spell by an additional 2%.
    - Rank 2: Increases the critical strike chance of your Arcane Explosion spell by an additional 4%.
    - Rank 3: Increases the critical strike chance of your Arcane Explosion spell by an additional 6%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Arcane Blast** (1 rank, column 4) [NEW]
  - Cost: 122 Mana, 30 yd range, 2.5 sec cast
  - Blasts the target with energy, dealing 95 to 104 Arcane damage. Each time you cast Arcane Blast, the damage of all your other spells is increased by 10% and the mana cost of Arcane Blast is increased by 175%. Effect stacks up to 4 times and lasts 8 sec or until any other damage spell is cast.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Arcane Shielding** (2 ranks, column 1) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Decreases the Mana lost per point of damage taken when your Mana Shield spell is active by 17% and increases the resistances granted by your Mage Armor spell by 25%.
  - Rank 2: Decreases the Mana lost per point of damage taken when your Mana Shield spell is active by 34% and increases the resistances granted by your Mage Armor spell by 50%.
  - Classic (2 ranks):
    - Rank 1: Decreases the mana lost per point of damage taken when Mana Shield is active by 10%.
    - Rank 2: Decreases the mana lost per point of damage taken when Mana Shield is active by 20%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Improved Counterspell** (2 ranks, column 2) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Your Counterspell also Silences the target for 2 sec.
  - Rank 2: Rank 2 values not yet captured from footage.
  - Classic (2 ranks):
    - Rank 1: Gives your Counterspell a 50% chance to silence the target for 4 sec.
    - Rank 2: Gives your Counterspell a 100% chance to silence the target for 4 sec.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Arcane Meditation** (3 ranks, column 3) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Requires: Arcane Concentration (5/5)
  - Rank 1: Allows 17% of your Mana regeneration to continue while casting.
  - Rank 2: Allows 34% of your Mana regeneration to continue while casting.
  - Rank 3: Allows 51% of your Mana regeneration to continue while casting.
  - Classic (3 ranks):
    - Rank 1: Allows 5% of your Mana regeneration to continue while casting.
    - Rank 2: Allows 10% of your Mana regeneration to continue while casting.
    - Rank 3: Allows 15% of your Mana regeneration to continue while casting.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Missile Barrage** (1 rank, column 4) [CHANGED]
  - Gives your Arcane Blast spell a 40% chance, and your Fireball, Frostbolt, and Frostfire Bolt spells a 20% chance to reduce the channeled duration of your next Arcane Missiles spell by 50%, reduce the Mana cost by 100%, and missiles will fire every 0.5 sec.
  - Note: Replaced ability_mage_missiles (404 on zamimg) with the WotLK Missile Barrage icon.

#### Row 5 (20 points)

- **Presence of Mind** (1 rank, column 2) [unchanged]
  - Cost: Instant, 3 min cooldown
  - When activated, your next Mage spell with a casting time less than 10 sec becomes an instant cast spell.
- **Arcane Mind** (5 ranks, column 3) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases your Intellect by 2% and increases the critical strike damage bonus of your Arcane spells by 20%.
  - Rank 2: Increases your Intellect by 4% and increases the critical strike damage bonus of your Arcane spells by 40%.
  - Rank 3: Increases your Intellect by 6% and increases the critical strike damage bonus of your Arcane spells by 60%.
  - Rank 4: Increases your Intellect by 8% and increases the critical strike damage bonus of your Arcane spells by 80%.
  - Rank 5: Increases your Intellect by 10% and increases the critical strike damage bonus of your Arcane spells by 100%.
  - Classic (5 ranks):
    - Rank 1: Increases your maximum Mana by 2%.
    - Rank 2: Increases your maximum Mana by 4%.
    - Rank 3: Increases your maximum Mana by 6%.
    - Rank 4: Increases your maximum Mana by 8%.
    - Rank 5: Increases your maximum Mana by 10%.

#### Row 6 (25 points)

- **Arcane Instability** (3 ranks, column 3) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Requires: Arcane Mind (5/5)
  - Rank 1: Increases the damage done by all your spells by 1% and your critical strike chance with all attacks by 1%.
  - Rank 2: Increases the damage done by all your spells by 2% and your critical strike chance with all attacks by 2%.
  - Rank 3: Increases the damage done by all your spells by 3% and your critical strike chance with all attacks by 3%.
  - Classic (3 ranks):
    - Rank 1: Increases your spell damage and critical strike chance by 1%.
    - Rank 2: Increases your spell damage and critical strike chance by 2%.
    - Rank 3: Increases your spell damage and critical strike chance by 3%.
  - Note: Rank 1 from footage. Rank 2 was still Classic text, rewritten from rank 1. Rank 3 already matched the Forever wording.

#### Row 7 (30 points)

- **Arcane Power** (1 rank, column 2) [CHANGED]
  - Requires: Presence of Mind (1/1)
  - Cost: Instant, 3 min cooldown
  - For the next 15 sec, your spells deal 30% more damage while costing 30% more mana to cast.
  - Classic: When activated, your spells deal 30% more damage while costing 30% more mana to cast. This effect lasts 15 sec.

### Fire

17 talents, 2 new. 31-point talent: Combustion.

#### Row 1 (0 points)

- **Wake of Fire** (2 ranks, column 1) [NEW] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Reduces the cooldown of your Fire Blast spell by 1 sec. Killing a non-trivial target increases the critical strike chance of your next Fire Blast cast within 20 sec by 25%.
  - Rank 2: Reduces the cooldown of your Fire Blast spell by 2 sec. Killing a non-trivial target increases the critical strike chance of your next Fire Blast cast within 20 sec by 50%.
  - Note: Rank 2 confirmed on screen: 2 sec and 50%.
- **Incineration** (3 ranks, column 2) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases the critical strike chance of your Fire Blast and Scorch spells by 2%.
  - Rank 2: Increases the critical strike chance of your Fire Blast and Scorch spells by 4%.
  - Rank 3: Increases the critical strike chance of your Fire Blast and Scorch spells by 4%.
  - Classic (2 ranks):
    - Rank 1: Increases the critical strike chance of your Fire Blast and Scorch spells by 2%.
    - Rank 2: Increases the critical strike chance of your Fire Blast and Scorch spells by 4%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Improved Fireball** (5 ranks, column 3) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Reduces the casting time of your Fireball spell by 0.1 sec.
  - Rank 2: Reduces the casting time of your Fireball spell by 0.2 sec.
  - Rank 3: Reduces the casting time of your Fireball spell by 0.3 sec.
  - Rank 4: Reduces the casting time of your Fireball spell by 0.4 sec.
  - Rank 5: Reduces the casting time of your Fireball spell by 0.5 sec.

#### Row 2 (5 points)

- **Ignite** (5 ranks, column 1) [unchanged]
  - Rank 1: Your critical strikes from Fire damage spells cause the target to burn for an additional 8% of your spell's damage over 4 sec.
  - Rank 2: Your critical strikes from Fire damage spells cause the target to burn for an additional 16% of your spell's damage over 4 sec.
  - Rank 3: Your critical strikes from Fire damage spells cause the target to burn for an additional 24% of your spell's damage over 4 sec.
  - Rank 4: Your critical strikes from Fire damage spells cause the target to burn for an additional 32% of your spell's damage over 4 sec.
  - Rank 5: Your critical strikes from Fire damage spells cause the target to burn for an additional 40% of your spell's damage over 4 sec.
  - Note: Text and slot from the third-party dataset, not yet verified in footage by us.
- **Flame Throwing** (2 ranks, column 2) [unchanged]
  - Rank 1: Increases the range of your Fire spells by 3 yards.
  - Rank 2: Increases the range of your Fire spells by 6 yards.
  - Note: Placed from the third-party dataset, not yet verified in footage by us.
- **Impact** (3 ranks, column 3) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Gives your Fire spells a 2% chance to stun the target for 2 sec.
  - Rank 2: Gives your Fire spells a 4% chance to stun the target for 2 sec.
  - Rank 3: Gives your Fire spells a 6% chance to stun the target for 2 sec.
  - Classic (5 ranks):
    - Rank 1: Gives your Fire spells a 2% chance to stun the target for 2 sec.
    - Rank 2: Gives your Fire spells a 4% chance to stun the target for 2 sec.
    - Rank 3: Gives your Fire spells a 6% chance to stun the target for 2 sec.
    - Rank 4: Gives your Fire spells a 8% chance to stun the target for 2 sec.
    - Rank 5: Gives your Fire spells a 10% chance to stun the target for 2 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 3 (10 points)

- **Burning Soul** (3 ranks, column 1) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Gives your Fire spells a 23% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 10%.
  - Rank 2: Gives your Fire spells a 70% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 30%.
  - Rank 3: Gives your Fire spells a 70% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 30%.
  - Classic (2 ranks):
    - Rank 1: Gives your Fire spells a 35% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 15%.
    - Rank 2: Gives your Fire spells a 70% chance to not lose casting time when you take damage and reduces the threat caused by your Fire spells by 30%.
- **Improved Flamestrike** (3 ranks, column 2) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases the critical strike chance of your Flamestrike spell by 5%.
  - Rank 2: Increases the critical strike chance of your Flamestrike spell by 10%.
  - Rank 3: Increases the critical strike chance of your Flamestrike spell by 15%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Pyroblast** (1 rank, column 3) [CHANGED]
  - Cost: 125 Mana, 35 yd range, 6 sec cast
  - Hurls an immense fiery boulder that causes 155 to 185 Fire damage and an additional 76 Fire damage over 12 sec.
  - Classic: Hurls an immense fiery boulder that causes 149 to 195 Fire damage and an additional 56 Fire damage over 12 sec.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.

#### Row 4 (15 points)

- **Improved Scorch** (3 ranks, column 1) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Your Scorch spell has a 33% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases all Fire damage you deal to your target by 3% and lasts 30 sec, stacking up to 5 times.
  - Rank 2: Your Scorch spells have a 66% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
  - Rank 3: Your Scorch spells have a 100% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
  - Classic (3 ranks):
    - Rank 1: Your Scorch spells have a 33% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
    - Rank 2: Your Scorch spells have a 66% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
    - Rank 3: Your Scorch spells have a 100% chance to cause your target to be vulnerable to Fire damage. This vulnerability increases the Fire damage dealt to your target by 3% and lasts 30 sec. Stacks up to 5 times.
- **Improved Fire Ward** (2 ranks, column 2) [unchanged]
  - Rank 1: Causes your Fire Ward to have a 10% chance to reflect Fire spells while active.
  - Rank 2: Causes your Fire Ward to have a 20% chance to reflect Fire spells while active.
- **Hot Streak** (1 rank, column 3) [NEW]
  - Requires: Pyroblast (1/1)
  - Your non-periodic critical strikes with Fireball, Frostfire Bolt, Fire Blast, and Scorch grant Hot Streak for 15 sec. Hot Streak reduces the cast time of Pyroblast by 25%, stacking up to 3 times.
- **Master of Elements** (3 ranks, column 4) [unchanged]
  - Rank 1: Your Fire and Frost critical strikes will refund 10% of their base mana cost.
  - Rank 2: Your Fire and Frost spell criticals will refund 20% of their base mana cost.
  - Rank 3: Your Fire and Frost spell criticals will refund 30% of their base mana cost.

#### Row 5 (20 points)

- **Critical Mass** (3 ranks, column 2) [unchanged]
  - Rank 1: Increases the critical strike chance of your Fire spells by 2%.
  - Rank 2: Increases the critical strike chance of your Fire spells by 4%.
  - Rank 3: Increases the critical strike chance of your Fire spells by 6%.
- **Blast Wave** (1 rank, column 3) [unchanged]
  - Cost: 215 Mana, Instant, 45 sec cooldown
  - A wave of flame radiates outward from the caster, damaging all enemies caught within the blast for 160 to 191 Fire damage, and Dazing them for 50% reduced movement speed for 6 sec.
  - Note: Placed from the third-party dataset, not yet verified in footage by us.

#### Row 6 (25 points)

- **Fire Power** (5 ranks, column 3) [unchanged]
  - Rank 1: Increases the damage done by your Fire spells by 2%.
  - Rank 2: Increases the damage done by your Fire spells by 4%.
  - Rank 3: Increases the damage done by your Fire spells by 6%.
  - Rank 4: Increases the damage done by your Fire spells by 8%.
  - Rank 5: Increases the damage done by your Fire spells by 10%.

#### Row 7 (30 points)

- **Combustion** (1 rank, column 2) [CHANGED]
  - Requires: Critical Mass (3/3)
  - Cost: Instant, 3 min cooldown
  - When activated, this spell causes each of your Fire damage spell hits to increase your critical strike chance with Fire damage spells by 10%. This effect lasts until you have caused 4 non-periodic critical strikes with Fire spells.
  - Classic: When activated, this spell causes each of your Fire damage spell hits to increase your critical strike chance with Fire damage spells by 10%. This effect lasts until you have caused 3 critical strikes with Fire spells.

### Frost

19 talents, 1 new. 31-point talent: Ice Barrier.

#### Row 1 (0 points)

- **Frost Warding** (2 ranks, column 1) [unchanged]
  - Rank 1: Increases the Armor and resistance given by your Frost Armor and Ice Armor spells by 15%. In addition, gives your Frost Ward a 10% chance to reflect Frost spells and effects while active.
  - Rank 2: Increases the armor and resistances given by your Frost Armor and Ice Armor spells by 30%. In addition, gives your Frost Ward a 20% chance to reflect Frost spells and effects while active.
- **Improved Frostbolt** (5 ranks, column 2) [unchanged]
  - Rank 1: Reduces the casting time of your Frostbolt spell by 0.1 sec.
  - Rank 2: Reduces the casting time of your Frostbolt spell by 0.2 sec.
  - Rank 3: Reduces the casting time of your Frostbolt spell by 0.3 sec.
  - Rank 4: Reduces the casting time of your Frostbolt spell by 0.4 sec.
  - Rank 5: Reduces the casting time of your Frostbolt spell by 0.5 sec.
- **Elemental Precision** (5 ranks, column 3) [CHANGED] [ranks 2+ settled by the 70338 trait curves] [3 ranks of 2% since the 2026-10-20 hotfix, see the top]
  - Rank 1: Increases your chance to hit with Frost and Fire spells by 1%.
  - Rank 2: Increases your chance to hit with Frost and Fire spells by 2%.
  - Rank 3: Increases your chance to hit with Frost and Fire spells by 3%.
  - Rank 4: Increases your chance to hit with Frost and Fire spells by 4%.
  - Rank 5: Increases your chance to hit with Frost and Fire spells by 5%.
  - Classic (3 ranks):
    - Rank 1: Reduces the chance that the opponent can resist your Frost and Fire spells by 2%.
    - Rank 2: Reduces the chance that the opponent can resist your Frost and Fire spells by 4%.
    - Rank 3: Reduces the chance that the opponent can resist your Frost and Fire spells by 6%.
  - Note: Rank 1 from footage (hit, not resist). Ranks 2+ were still Classic resist text, rewritten from rank 1 with linear scaling.

#### Row 2 (5 points)

- **Ice Shards** (5 ranks, column 1) [unchanged]
  - Rank 1: Increases the critical strike damage bonus of your Frost spells by 20%.
  - Rank 2: Increases the critical strike damage bonus of your Frost spells by 40%.
  - Rank 3: Increases the critical strike damage bonus of your Frost spells by 60%.
  - Rank 4: Increases the critical strike damage bonus of your Frost spells by 80%.
  - Rank 5: Increases the critical strike damage bonus of your Frost spells by 100%.
- **Permafrost** (3 ranks, column 2) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases the duration of your Chill effects by 11% and reduces the target's speed by an additional 3%.
  - Rank 2: Increases the duration of your Chill effects by 22% and reduces the target's speed by an additional 6%.
  - Rank 3: Increases the duration of your Chill effects by 33% and reduces the target's speed by an additional 9%.
  - Classic (3 ranks):
    - Rank 1: Increases the duration of your Chill effects by 1 sec and reduces the target's speed by an additional 4%.
    - Rank 2: Increases the duration of your Chill effects by 2 secs and reduces the target's speed by an additional 7%.
    - Rank 3: Increases the duration of your Chill effects by 3 secs and reduces the target's speed by an additional 10%.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.
- **Improved Frost Nova** (2 ranks, column 3) [unchanged]
  - Rank 1: Reduces the cooldown of your Frost Nova spell by 2 sec.
  - Rank 2: Reduces the cooldown of your Frost Nova spell by 4 sec.
- **Frostbite** (3 ranks, column 4) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Gives your Chill effects a 5% chance to freeze the target for 5 sec.
  - Rank 2: Gives your Chill effects a 10% chance to freeze the target for 5 sec.
  - Rank 3: Gives your Chill effects a 15% chance to freeze the target for 5 sec.

#### Row 3 (10 points)

- **Piercing Ice** (3 ranks, column 1) [unchanged] [removed by the 2026-10-20 hotfix, see the top]
  - Rank 1: Increases the damage done by your Frost spells by 2%.
  - Rank 2: Increases the damage done by your Frost spells by 4%.
  - Rank 3: Increases the damage done by your Frost spells by 6%.
- **Frost Channeling** (3 ranks, column 2) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Reduces the mana cost of your Frost spells by 5% and reduces the threat caused by your Frost spells by 10%.
  - Rank 2: Reduces the mana cost of your Frost spells by 10% and reduces the threat caused by your Frost spells by 20%.
  - Rank 3: Reduces the mana cost of your Frost spells by 15% and reduces the threat caused by your Frost spells by 30%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Ice Lance** (1 rank, column 3) [CHANGED]
  - Requires: Improved Frost Nova (2/2)
  - Cost: 45 Mana, 30 yd range, Instant
  - Deals 28 to 33 Frost damage to an enemy target. Deals 300% increased damage to Frozen targets.
  - Note: Filled the question-mark placeholder with the WotLK Ice Lance icon.
- **Improved Blizzard** (3 ranks, column 4) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Adds a Chill effect to your Blizzard spell. This effect lowers the target's movement speed by 15% for 1.5 sec.
  - Rank 2: Adds a Chill effect to your Blizzard spell. This effect lowers the target's movement speed by 30% for 1.5 sec.
  - Rank 3: Adds a Chill effect to your Blizzard spell. This effect lowers the target's movement speed by 45% for 1.5 sec.
  - Classic (3 ranks):
    - Rank 1: Adds a chill effect to your Blizzard spell. This effect lowers the target's movement speed by 30%. Lasts 1.5 sec.
    - Rank 2: Adds a chill effect to your Blizzard spell. This effect lowers the target's movement speed by 50%. Lasts 1.5 sec.
    - Rank 3: Adds a chill effect to your Blizzard spell. This effect lowers the target's movement speed by 65%. Lasts 1.5 sec.
  - Note: Rank 1 from footage. Ranks 2+ were still Classic text, rewritten from rank 1 with linear scaling. Rank 2+ values not yet seen on camera.

#### Row 4 (15 points)

- **Arctic Reach** (2 ranks, column 1) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases the range of your Frostbolt and Blizzard spells and the radius of your Frost Nova and Cone of Cold spells by 10%.
  - Rank 2: Increases the range of your Frostbolt and Blizzard spells and the radius of your Frost Nova and Cone of Cold spells by 20%.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Ice Block** (1 rank, column 2) [unchanged]
  - Cost: 15 Mana, Instant, 5 min cooldown
  - You become encased in a block of ice, protecting you from all physical attacks and spells for 10 sec, but during that time you cannot attack, move or cast spells.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Shatter** (3 ranks, column 4) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases the critical strike chance of all your spells against frozen targets by 10%.
  - Rank 2: Increases the critical strike chance of all your spells against frozen targets by 20%.
  - Rank 3: Increases the critical strike chance of all your spells against frozen targets by 30%.
  - Classic (5 ranks):
    - Rank 1: Increases the critical strike chance of all your spells against frozen targets by 10%.
    - Rank 2: Increases the critical strike chance of all your spells against frozen targets by 20%.
    - Rank 3: Increases the critical strike chance of all your spells against frozen targets by 30%.
    - Rank 4: Increases the critical strike chance of all your spells against frozen targets by 40%.
    - Rank 5: Increases the critical strike chance of all your spells against frozen targets by 50%.

#### Row 5 (20 points)

- **Improved Cone of Cold** (3 ranks, column 1) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Increases the damage dealt by your Cone of Cold spell by 12%.
  - Rank 2: Increases the damage dealt by your Cone of Cold spell by 25%.
  - Rank 3: Increases the damage dealt by your Cone of Cold spell by 35%.
  - Classic (3 ranks):
    - Rank 1: Increases the damage dealt by your Cone of Cold spell by 15%.
    - Rank 2: Increases the damage dealt by your Cone of Cold spell by 25%.
    - Rank 3: Increases the damage dealt by your Cone of Cold spell by 35%.
- **Cold Snap** (1 rank, column 2) [CHANGED]
  - Cost: Instant, 10 min cooldown
  - Finishes the remaining cooldown on all your other Frost spells.
  - Classic: When activated, this spell finishes the cooldown on all of your Frost spells.
  - Note: Slot corrected from the third-party dataset, not yet verified in footage by us.
- **Fingers of Frost** (2 ranks, column 3) [NEW] [ranks 2+ settled by the 70338 trait curves]
  - Requires: Ice Lance (1/1)
  - Rank 1: Gives your Chill effects a 15% chance to grant you the Fingers of Frost effect, which treats your next 1 spell cast as if the target were Frozen. Lasts 15 sec.
  - Rank 2: Gives your Chill effects a 30% chance to grant you the Fingers of Frost effect, which treats your next 1 spell cast as if the target were Frozen. Lasts 15 sec.

#### Row 6 (25 points)

- **Winter's Chill** (5 ranks, column 3) [CHANGED] [ranks 2+ settled by the 70338 trait curves]
  - Rank 1: Gives your Frost damage spells a 20% chance to apply the Winter's Chill effect, which increases the chance your Ice Lance and Frostbolt spells will critically hit the target by 2% for 15 sec. Stacks up to 1 times.
  - Rank 2: Gives your Frost damage spells a 40% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
  - Rank 3: Gives your Frost damage spells a 60% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
  - Rank 4: Gives your Frost damage spells a 80% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
  - Rank 5: Gives your Frost damage spells a 100% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
  - Classic (5 ranks):
    - Rank 1: Gives your Frost damage spells a 20% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
    - Rank 2: Gives your Frost damage spells a 40% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
    - Rank 3: Gives your Frost damage spells a 60% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
    - Rank 4: Gives your Frost damage spells a 80% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.
    - Rank 5: Gives your Frost damage spells a 100% chance to apply the Winter's Chill effect, which increases the chance a Frost spell will critically hit the target by 2% for 15 sec. Stacks up to 5 times.

#### Row 7 (30 points)

- **Ice Barrier** (1 rank, column 2) [CHANGED] [3% Frost damage while shielded and Frigid Power since the 2026-10-20 hotfix, see the top]
  - Requires: Cold Snap (1/1)
  - Cost: 305 Mana, Instant, 30 sec cooldown
  - Instantly shields you, absorbing 431 damage. Lasts 1 min. While the shield holds, your spellcasts will not be interrupted or delayed from taking damage.
  - Classic: Instantly shields you, absorbing 455 damage. Lasts 1 min. While the shield holds, spells will not be interrupted.
