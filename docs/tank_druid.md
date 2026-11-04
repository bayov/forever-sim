# Feral tank (bear)

The Forever bear as the tank sim runs it: the rules audit of every tank ability and talent, what was fixed
and built, the tests, the rotation template, the synthetic set, the three presets and their numbers, and the
calls for the user. Part of the tank workers' brief (`orchestration/brief_tanks.md`, T1); the tank sim itself
is [tank_sim.md](tank_sim.md).

Code: `sim/druid/bear.go`, `bear_attacks.go`, `barkskin.go`, `swipe.go` and the talents in `talents.go`;
the spec `sim/druid/tank`; tests `sim/druid/tank/bear_rules_test.go` (a test per ability and talent) and
`tank_test.go` (the bench's bear as a whole); the template `tools/rotopt/druid_tank.go`; the set
`tools/database/forever_synthetic_items_druid_tank.go`; the requests `tools/tankbench/req/druid_bear*.json`;
the UI `ui/feral_tank_druid`.

## After the 70753 update (2026-11-04)

The audit below was made on client 1.60.1.70753 already (its talent trees, Lacerate a row 4 talent with
its two new ranks, Primal Bite row 7 with three ranks and Leader of the Pack before it, came in with the
10-30 hotfix). Client 70755's spell tables are 70753's, and no hotfix push since (through 113483, all
rogue) touches a druid row. What moved for the bear is the November 2 notes, modelled under Forever at the
notes' value (standards 37, notes-only, `orchestration/round_70753/druid.md`):

- **Bear and Dire Bear Form threat 115%** (was 50%; 21178 still says 50): x2.15 (`bearFormThreatMultiplierNov2`,
  the orchestrator's change). It is applied once, on the form's aura: the threat tests read x2.15 on the
  form, Maul, Primal Bite and Demoralizing Roar.
- **Rend and Tear 3% a point** (was 2%; master's druid round): it raises the bear's Maul, Swipe, Primal
  Bite, Lacerate's hit and its ticks too. The tests read 1.15 at 5/5.
- **SafeDPS's reference member** makes 1 threat a point of damage (Salvation is gone; it was 0.7). Against
  the form's x2.15 / 1.5 = x1.43 that is a wash by design: the notes raised the form "to compensate for
  Blessing of Salvation being removed", and 1 / 0.7 is x1.43 too.
- The races' base stats (A2): the Night Elf bear has 3 Agility more and 1 Stamina less.
- The merge kept nothing twice: the bear's files took master's threat and Rend and Tear lines and nothing
  else (`git diff` of `sim/druid` since the pause). Two golden lines per DPS spec (the tank set's AllItems
  lines) kept their pre-round values and moved by the round's own factor (Feral x1.20, Balance x1.07).

**What it moved** (Balanced, the pre-update Balanced rotation and build, mean of the six scoring encounters,
3,000 iterations, the same seeds): **1,736.2 before, 1,833.6 after (+97.4)**, the 3% Rend and Tear and
the race stats; the threat change and the reference member's cancel.

## Rules audit (2026-11-03)

Against Forever's newest client, **1.60.1.70753** with its hotfix rows (`tools/sources/wago.py fetch
1.60.1.70753+hf`; 70753 writes the October 30 hotfix into the tables, and no hotfix row since touches a bear
spell), the talent rows and their trait curves (TraitDefinitionEffectPoints, CurvePoint), SkillLineAbility
for what a druid can learn, and the patch notes in `orchestration/update_ledger.md`. Tiers as the brief:
**T1** the client, **T2** wowsims/forever's in-game measurements and logs (github.com/wowsims/forever,
read at c7eafc85cd), **T3** Classic's (Era client, vmangos), **A** an assumption no source confirms (it's in
the calls at the end). The tooltip alone is never a source for a number the effect rows don't carry.

### Form and passives

| What | Client (T1) | Sim | Verdict |
|---|---|---|---|
| Dire Bear Form (9634, passive 9635), from 40 | armor from items +360% (aura 142 misc 1, also an aura 466 row of 360), health 600 + 32 a level (1,240 at 60), attack power 120 + 3 a level (180), 55% of base mana, GCD | the same; health takes the Tauren's 5% | right, tested |
| Bear Form (5487, passive 1178), 10 to 39 | +180%, 20 + 18 a level, 30 + 3 a level | the same | right, tested |
| Bear threat (Bear Form (Passive2) 21178) | A_MOD_THREAT 50 on every school: x1.5 (70753 and 70755 unchanged). The November 2 notes: "Threat bonus increased from 50% to 115%", for Salvation's removal | x2.15 under Forever (`bearFormThreatMultiplierNov2`) | notes-only (standards 37), tested (the form, Maul, Primal Bite, Roar; `TestForeverBearFormThreat`) |
| Bear swing | SpellShapeshiftForm 8: 2,500 ms, DamageVariance 0.4 | the weapon's DPS at 2.5 sec, 20% either side, the enchant's and stone's flat damage on top (T2: wowsims/forever b5d2e0b992) | right, tested |
| Rage from the swing | not in the client | the two-hander's 4.5 a second of swing whatever the weapon (T2: their log, 11.25 a 2.5 sec swing; standards 21), a crit twice that (T1: the October 8 notes, Bear and Dire Bear Form crit Rage 100%) | **was Classic's Rage from damage (13.2 for two swings at +0 attack power, 48.0 at +3,000), crits not doubled** (tank_mechanics-20261104-7) | fixed 2026-11-04, tested (TestBearWhiteHitRage) |
| Rage from damage taken | not in the client | standards 21's rule (engine) | engine's |
| The shift into Bear Form | Rage starts at 0 in the form (T2: their forms.go) | **was missing**: now a cast of Bear Form (also from Bear Form) spends the Rage and Furor adds its 10 | fixed, tested |
| Omen of Clarity (16864, baseline from 20) | Clearcasting 16870's class mask takes in Maul and Swipe (0x800), Demoralizing Roar (0x8), Bash (0x2000), Primal Bite and Lacerate (word 1 0x40, 0x100) | **the bear's attacks had no Omen flag, so a proc in Bear Form did nothing**; now they're free under it and use it up (Maul as it lands, since it's the swing) | fixed, tested |
| Omen's rate | the server's | 2 a minute at the swing speed (the cat's rule): 8.3% a 2.5 sec swing, 10 sec cooldown | A (the cat's rule, T2 for the cat) |

### Abilities

| Ability | Client (T1) | Sim | Verdict |
|---|---|---|---|
| Maul rank 7 (9881, 58) | weapon damage + 128, 15 Rage, next swing, 80% back on a miss (Attributes_1 0x8000000), no threat effect | the same; threat its damage x2.15 (the form) | right, tested; its threat beyond the form is a call |
| Swipe rank 5 (9908, 54) | 83 to 3 enemies, 20 Rage, no refund, no threat effect | 83 + 3% of attack power (T1: the post 4 development notes; wowsims/forever's tester fit, PR #149, agrees) | right, tested; the 3% is T1, notes-only |
| Demoralizing Roar rank 5 (9898, 52) | -193 - 1.4 a level attack power (204 at 60) for 30 sec, 10 Rage, effect 1 THREAT 42 + 0.8 a level (48.4 at 60) | the same, x2.15 | right, tested |
| Enrage (5229, 12) | effect 1: 10 Rage at once; effect 0: 2 Rage a second for 10 sec; 1 min cooldown, no GCD; the armor cut is a dummy (effect 3) the server runs; the tooltip says 27% in Bear Form, 16% in Dire Bear Form; 25503 (no form) has -26% armor | 10 + 20 Rage, armor from items x0.84 in Dire Bear Form (x0.73 in Bear Form). **It lost its tenth tick** to the aura's expiry (28 Rage, not 30) | fixed, tested; the 16% is a call |
| Frenzied Regeneration (22842, 36; Forever's only rank) | 10 sec, effect 0 periodic 1 sec, "up to 10 Rage a second, each into $s1% health" with s1 1; 3 min cooldown, GCD | 1% of max health a point of Rage (T2 reads it the same) **lost its tenth tick** the same way | fixed, tested |
| Lacerate (talent; 1328617 at 25, 1328616 at 34, 414644 at 42, 1235826 at 50, 1235827 at 58) | ticks of 4 / 6 / 10 / 12 / 15 a stack every 3 sec for 15 sec, 5 stacks, effect 1 "10% weapon damage a stack already on", 15 Rage, refund on a miss, no threat effect | **ranks 1 and 2 were missing** (no Lacerate under 42); Rend and Tear **didn't raise its ticks** (it does Rake's and Rip's, and wowsims/forever's Lacerate's) | fixed, tested |
| Primal Bite (talent; SkillLineAbility: 407995 at 40, 1238069 at 50, 1238070 at 60) | weapon damage + 42 / 62 / 77, 20 Rage, 6 sec cooldown, refund on a miss, no threat effect; the development notes (post 4, client 70170): "Threat generated by this ability has been roughly doubled" | **was 26 / 38 / 59 / 77 at 25 / 36 / 48 / 60, and 1238073 at 60**, a row no trainer or talent teaches (the 77 is the same); **its threat was the damage's only** until 2026-11-04 (tank_mechanics-20261104-2); now 2x the damage, then the form's 2.15x, under Forever (T1, notes-only, standards 37) | fixed, tested |
| Barkskin (22812, 44) | 20% less **Physical** damage taken (aura 87, school 1) for **15 sec**, 1 min cooldown, the GCD, no cost, no shapeshift row (cast in any form) | was WotLK's: 20% of all damage for 12 sec, off the GCD, stopping auto attack, and never registered | fixed (rewritten), tested |
| Bash (5211 at 14, 6798 at 30, 8983 at 46) | a 2 / 3 / 4 sec stun, 10 Rage, 1 min cooldown, GCD, a melee attack with a refund, no damage or threat | **was missing**: added (raid bosses can't be stunned, so it scores nothing) | added, tested |
| Growl (6795), Challenging Roar (5209) | taunts: Growl's effect 114 (attack me) on an 8 sec cooldown; Challenging Roar 6 sec, 10 min, 15 Rage | not simmed | rightly nothing: the scoring fights have one tank on the boss (tank swaps are the engine's, `Target.tank_swap_seconds`), and a taunt adds no threat when the tank is already on top |
| Feral Charge (Bear) (16979 / talent 1238122) | charge, 5 Rage, 15 sec | not simmed | rightly nothing: the bench's bear starts in melee range |
| Faerie Fire (Feral) | none under Forever (16857, 17390 to 17392 have no client rows; the cat casts the caster one, free in form) | not cast by the bear | the raid's Faerie Fire is on the boss; the bear's threat from it is 0 (its E_THREAT has base 0, T2 51f0e9e08f) |
| Leader of the Pack (17007, aura 24932) | 3% crit of every kind (aura 290) for the party or raid in the feral forms | the raid buff, set by the bear in Bear Form | right, tested (it doesn't stack with Moonkin Aura, the user's ruling) |

### Talents a bear can reach

| Talent | Client (T1, the trait curves) | Sim | Verdict |
|---|---|---|---|
| Ferocity 5 | -1 Rage a point off Maul, Swipe (0x800) and Primal Bite (word 1 0x40) | the same | right, tested |
| Furor 5 | 20% a point to gain 10 Rage (17057) on a shift into the bear forms | **the bear half was missing** | added, tested (0, 40% and 100%) |
| Feral Swiftness 2 | 2% dodge a point (aura 49, no form row: always) | in the bear forms (the bear never leaves them) | right, tested |
| Feral Instinct 3 | 10% Swipe damage a point (word 1 0x100000) | the same | right, tested |
| Brutal Impact 2 | 0.5 sec more stun and 15 sec less Bash cooldown a point | with Bash | added, tested |
| Thick Hide 3 | dummies: 1 base armor a level a point (curve 1 / 2 / 3, measured 30 / 60 / 90 at 30, T2), and 0.67 / 1.33 / 2.00 base armor a point of Defense past five times the level (effect 0's curve 67 / 133 / 200, over 100), both raised by the form | the same, the Defense part read when the form comes up | right, tested; the Defense part is the client's curve on a dummy the server runs, no log |
| Shredding Attacks 3 | -1 Rage a point off Lacerate (effect 1, -10 tenths) | the same | right, tested |
| Savage Fury 2 | 5% a point on Claw, Rake, Shred, Maul and Swipe (mask 0x9800) | the same | right, tested |
| Sharpened Claws 2 | 3% crit a point in the feral forms | the same | right, tested |
| Predatory Strikes 3 | 50% of the level as attack power a point in the feral forms | the same | right, tested |
| Lacerate, Primal Bite | the spells above | | |
| Blood Frenzy 2 | 50% a point to gain 5 Rage (16959) on a crit in the bear forms (ProcTypeMask melee) | the same | right, tested |
| Leader of the Pack | above | | |
| Natural Instinct 2 | 10% a point more crit damage bonus on 1223242's mask: word 0 0x839800 (Maul and Swipe's 0x800 among them), word 1 0x540 (Primal Bite, Lacerate, Rake) | **covered only the cat's attacks** | fixed, tested |
| Rend and Tear 5 (column 1 since 70753) | 2% a point on the melee abilities against a bleeding target (a dummy; the tree still says 2%). The November 2 notes: 3 / 6 / 9 / 12 / 15% | 3% a point under Forever, on Maul, Swipe, Primal Bite, Lacerate's hit and its ticks; Lacerate's bleed counts | the percent notes-only (standards 37); tested (the ticks T2) |
| Natural Reaction 5 | 1% dodge a point; 20% a point to gain 5 Rage (417053) on a dodge | the same | right, tested in a whole fight (5 Rage a dodge at 5/5) |
| Heart of the Wild 5 (Restoration) | 4% Stamina a point in the bear forms | the same | right, tested |
| Naturalist 5 (Restoration) | 2% damage a point | the same | right, tested |
| Natural Shapeshifter 3 (Restoration) | 10% a point off the shapeshifts' mana | the same | right (the prepull shift's mana) |
| Nature's Majesty 2, Nature's Reach 2 (Balance) | 2% crit / hit a point | the same (the cat's audit) | right |

### Idols

| Idol | Client (T1) | Sim | Verdict |
|---|---|---|---|
| Idol of Brutality (23198, Stratholme) | 28855: -20 tenths (2 Rage) off Maul and Swipe; 1.12's tooltip said 3 | **was a stub ("implemented in maul.go")** | added, tested |
| Enraged Idol (272428, Forever) | 1291060: +10 tenths on Enrage's effect 0, so 3 Rage a tick, 10 more in all | missing | added, tested |
| Idol of the Ursine Twins (279251, Forever, crafted) | 1306483: a Lacerate that lands resets Primal Bite 10% of the time | missing | added, tested |

### Assumptions that stay (calls at the end)

- **No threat beyond the form's 2.15x** for Maul, Swipe and Lacerate, and Primal Bite's 2x from the notes
  ("roughly doubled", post 4, client 70170; the first audit missed it, tank_mechanics-20261104-2). No row
  states any of them and their tooltips' "high amount of threat" isn't a number; wowsims/forever (T2) also gives damage only
  (their TODOs say "in game research needed"). Classic's Maul and Swipe had 1.75x on top (T3, a server
  table, not the client).
- **Enrage's armor cut** at the tooltip's 16% of armor from items in Dire Bear Form: the client's only
  number is 25503's 26% (on no form), and wowsims/forever reads the tooltip too.
- **Omen of Clarity's rate in Bear Form**, the cat's 2 a minute at the swing speed.

## What the tests check

`sim/druid/tank/bear_rules_test.go` builds a level 60 bear with only what each test checks (a chest and the
two-hander, a level 63 target with no armor, exact damage), casts by hand and reads every result; its chest's
item effect records the results, since an aura can't be added once the sim is built. Each test fails when its
mechanic breaks: each compares the sim's result with the client's number, not with the code's (the tenth
Enrage and Frenzied Regeneration tick were found by them).

- **Form**: Dire Bear Form's health, attack power, armor x4.6 from items and threat x2.15 (on against
  cancelled); Bear Form at 30.
- **Every ability**: Maul (queued, struck in the swing's place, damage range +128, threat 2.15x, cost with
  Ferocity and the idol), Swipe (83 + 3% of attack power, 3 of 4 targets, 20 Rage, Feral Instinct and Savage
  Fury, the idol), Demoralizing Roar (204 attack power, 30 sec, both targets, 10 Rage, 48.4 x 2.15 threat),
  Lacerate (ranks by level, 15 Rage, stacks to 5, the hit by stacks, 5 ticks of 75 at 5 stacks, Shredding
  Attacks, Rend and Tear on the ticks), Primal Bite (ranks by level, damage +77, 20 Rage, 6 sec, threat,
  the refund on a miss, Ferocity), Enrage (10 at once, 30 in all, 40 with the idol, armor x0.84, no GCD, 1
  min), Frenzied Regeneration (10 Rage a second into 1% of health each), Barkskin (Physical only, 15 sec,
  1 min, GCD, in form), Bash (rank, cost, cooldown, stun, Brutal Impact), the shift (Rage to 0, Furor's 10
  at 0 / 2 / 5 points, its 40% at 2/5), Clearcasting on every bear attack and Maul, Omen procs in form.
- **Every talent's stats**: Feral Swiftness, Natural Reaction, Sharpened Claws, Predatory Strikes, Leader of
  the Pack, Heart of the Wild, Thick Hide with Defense; **damage**: Savage Fury, Naturalist, Rend and Tear,
  Natural Instinct; **Rage**: Blood Frenzy (every crit), Natural Reaction (every dodge, a whole fight).
- **Items**: the three idols and the armor trinket.
- `TestForeverBear` runs the bench's bear: the sheet, its rotation's casts, its own Roar, a tank score.

## Rotation (`tools/rotopt/druid_tank.go`)

The knobs are what a bear decides: Demoralizing Roar's refresh, Primal Bite on cooldown, Lacerate (kept at 5,
or the filler too) and its refresh, Swipe as a filler, Maul's Rage threshold and whether it leaves Primal
Bite its Rage, Enrage (before the pull at a chosen time, in the fight under a Rage, both), a shift before
the pull for Furor, Barkskin on its timer (ahead of the attacks or as the filler, from a time), Frenzied
Regeneration at a Rage, and the armor trinket. Nothing reads the bear's health. The bear keeps its own
Demoralizing Roar up (`tankscore.OwnDebuffs` drops the raid's Roar and Shout).

What round 1's searches found (round 2's changes are under "Presets"; the details per preset there):

- **The opener is most of SafeDPS.** SafeDPS is each fight's lowest moment for a reference DPS, and that is
  the pull, so what the bear has in its first seconds counts far over the rest: a shift into Dire Bear Form
  1.5 sec before the pull (+62 at Balanced even with Furor at 0, since the engine counts the mana that
  Judgement of Wisdom gives back as threat; +90 with Furor; measured when that threat was multiplied by the bear's
  2.15. Round 3, at 0.5 a mana flat: +25.5 at Balanced, +25.6 at Max threat, +22.0 at Max survival, all of it
  through Judgement of Wisdom, since without the debuff the shift is worth exactly 0), Enrage 1 sec before the pull in round 1 (its Rage at the
  pull, its armor cut over the first 9 sec; round 2's presets cast it in the fight instead; earlier, the Mighty Rage Potion at the pull runs into the 100
  Rage cap), the on-use Strength trinket at the pull (+166 over the armor trinket), and the Night Elf's
  Elune's Light (+172 over the Tauren after the update).
- **Barkskin on its timer, as the filler** (after Primal Bite and Lacerate's keep-up): +159 at Balanced.
  Ahead of the attacks it costs the opener's GCD (-145).
- **Lacerate is the filler** (to 5 stacks, refreshed under 4 sec, then cast whenever the GCD is free) and
  Primal Bite goes on cooldown; Maul at 30 Rage (20, and 10 at Max survival, in round 2). Swipe never beats
  Lacerate on one target. Frenzied Regeneration on a timer lost at every weight in round 1; at 100 Rage it
  fires once, at the pull's Rage peak, and round 2's Balanced and Max survival keep it there.
- **Demoralizing Roar below Lacerate's keep-up line** (round 2, `roarAfterLacerate`): ahead of the attacks it
  took the pull's first GCD and 10 Rage, about +200 at every weight (tank_rotations-20261104-1).

## Gear (`tools/database/forever_synthetic_items_druid_tank.go`)

Standards 36's bear row: physical archetype (3 lines a piece, the cat's budgets), the feral two-hander at
0.90 x L(76), 729.5 weighted points against 729.3. The row's small shares (Intellect, crit, attack and spell
power, 4.8 points) went to the rest in proportion. Hit by the tank objective: the Flask of Natural Accuracy
(5% hit) is the best flask (Balanced: +105 over the Titans; with the gear's hit at 8%, +35 over Natural
Aggression; Max survival likewise, Max threat a tie with Aggression), so the set carries 3% and its 1.6%
over the 8% cap went into Agility (+43, against +27 for Defense and +18 for Stamina). 254 Stamina, 164
Agility, 125 Strength, 105 Defense, 3% hit (after tank_gear-20261104-4: the chest +1 Stamina, 3 Stamina to Strength).

- **Trinkets**: the tank stat trinket (23 Defense and 1% dodge, 46.5 weighted, 1.3% over 1b's 45.9 since no
  whole Defense and dodge fall in the 1% band) and, of standards 36's two on-use kinds, the cat's Strength
  on-use (165 Strength for 20 sec every 2 min). The defensive kind (990815, 1,665 bonus armor, made for the
  comparison) scores 166 lower at Balanced and lower at every weight: its armor never touches healing needed
  (the bad moments come later in a fight), while the Strength lifts the opener's threat.
- **Idol**: Idol of the Ursine Twins, then Idol of Brutality (-19) and Enraged Idol (-36).
- **Enchants, round 3** (tank_fairness-20261104-8, tank_gear-20261104-3 (b) and -7; every slot's alternatives
  at each preset's weight, paired abtest, 3,000 iterations on seed 1, each gain confirmed on the disjoint seed 2
  and the close ones at 10,000 on seeds 3 and 4; raw output in `orchestration/tank_handoff/druid/round3/`).
  Diffs are the alternative against the round-2 enchant, Max threat / Balanced / Max survival:

  | Slot (worn) | Alternative | w=1/3 | w=1 | w=3 | Verdict |
  |---|---|---|---|---|---|
  | Wrist (Superior Stamina 1886) | **Superior Deflection 1248665** (Defense +9) | +5.2 +- 1.9 | +4.3 +- 4.6; seed 2 +8.6 +- 5.2; **10k +6.1 +- 2.7** | +12.4 +- 13.9; seed 2 +20.3 +- 14.7; **10k +23.9 +- 7.3** | **Balanced and Max survival** |
  | | Superior Strength 1885 | **+15.1 +- 0.4** (seed 2 +15.1 +- 0.3) | +4.5 +- 1.1; 10k +4.3 +- 0.7 | -35.0 +- 3.7 | **Max threat** |
  | | Superior Agility 1248599 | +11.5 +- 1.6 (-4.9, -3.0 against Strength) | +7.1 +- 4.2; 10k +10.0 +- 2.4 (+2.8 +- 1.8 over Deflection, a tie) | -5.1 +- 12.6 | tie with Deflection at Balanced, so the tank one |
  | Back (Agility 1219587) | Superior Defense 1889 (+70 armor) | -4.6 +- 1.1; seed 2 -6.9 +- 1.2 | -9.3 +- 3.0; seed 2 -10.8 +- 3.5 | -1.8 +- 8.8; seed 2 +5.2 +- 9.6 | Agility stays at every weight |
  | Head and legs (Agility 1508) | Strength 1506 | +4.2 +- 1.4 (head); both, on Max threat's wrist and weapon: **+7.2 +- 1.9, seed 2 +7.9 +- 1.9** | -9.1 +- 4.1 | -24.1 +- 10.8 | **Max threat** |
  | | Stamina 1507 / 100 health 1503 / 125 armor 1504 | -9.1 / -9.1 / -7.9 | -13.1 / -13.3 / -11.2 | +4.8 +- 11.0 (seed 2 -7.1) / +0.9 / -0.2 | Agility stays |
  | Weapon (2H Agility 2646) | 2H Strength 1248668 | **+11.2 +- 2.3, seed 2 +6.3 +- 2.2** | -23.1 +- 5.8 | -77.9 +- 16.8 | **Max threat** |
  | Neck (Agility 1249059) | Strength 1249019 | +3.5 +- 1.1; seed 2 +1.0 +- 1.2; on the new set +1.7 / +1.4 +- 1.2 | -7.7 +- 3.1 | -12.9 +- 8.7 | Agility stays (a tie at Max threat) |
  | Chest (Greater Stats 1891) | Major Health 1892 / Superior Stamina 13858 / Rugged Armor Kit 1843 | -11.7 / -11.9 / -11.8 | -13.4 / -13.6 / -14.2 | -7.5 / -8.0 / -12.4 | Greater Stats stays |
  | Gloves (Superior Strength 1248640) | Greater Agility 1887 / Strength 13887 | -18.2 / -15.4 | -14.4 / -14.7 | +5.8 +- 11.8 (seed 2 +7.0 +- 12.0) / -13.9 | stays |
  | Boots (Greater Agility 1887) | Greater Stamina 929 | -7.8 +- 1.3 | -14.1 +- 3.7 | +2.2 +- 10.3 (seed 2 +4.7 +- 10.8) | stays |

  Max threat's whole Strength set (head, legs, wrist, weapon) is +30.1 +- 2.9 on seed 2. Enchant ids now, each
  found for its slot in `assets/database/db.json`: Balanced and Max survival head and legs 1508, neck
  1249059, back 1219587, chest 1891, wrist 1248665, gloves 1248640, boots 1887, weapon 2646; Max threat the
  same with head and legs 1506, wrist 1885, weapon 1248668 (its own UI gear set, "Bear (Max threat)"). A bear
  has no shield, so the dead shield 929 (tank_gear-20261104-3) never reached it.
- **Consumables, round 3** (the same runs): every slot holds but Max survival's food. The Flask of Natural
  Accuracy (Titans -301 / -246 / -126, Natural Aggression -237 / -224 / -225), Elixir of the Phalanx and
  Greater Fortitude (every tank's), the Mongoose (Greater Agility -46 / -48 / -42), Juju Power (Giants -9.8 /
  -9.3 / -8.9), Juju Might (Winterfall Firewater -4.7 / -4.5 / -4.2), Dragonbreath Chili (without it -22 / -32
  / -77), the Dense Weightstone (none -42 / -40 / -38) and the Mighty Rage Potion (Greater Stoneshield -618 /
  -550 / -471). Food: the Smoked Desert Dumpling at Max threat and Balanced (Dirge's Kickin' Chimaerok Chops,
  25 Stamina, -31.8 +- 0.5 / -3.1 +- 1.9; Grilled Squid and Blessed Sunfruit -17 to -21), **Dirge's chops at
  Max survival (+89.9 +- 6.9, seed 2 +79.2 +- 6.1)**. Explosives score 0: an item can't be used in a form (the
  sim's explosives check `IsShapeshifted`), so a bear never throws them. Three Phase 1 consumables no tank
  request carries would gain (call 10): a Blasted Lands buff (Ground Scorpok Assay +42 / +53 / +112, R.O.I.D.S.
  +52 / +50 / +47, Lung Juice Cocktail +9 / +36 / +127) and Rumsey Rum Black Label (+6 / +21 / +75).

## Presets (round 3, 2026-11-04: unmultiplied mana-gain threat, disjoint seeds)

Round 3 re-searched the three presets on master 95bd28c's engine: mana-gain threat at 0.5 a mana with no threat
multiplier (tank_mechanics-20261104-12; the bear's 2.15 had multiplied Judgement of Wisdom's mana), and `-seed s`
of tankbench, abtest and rotopt's tank runs on iterations disjoint from seed 1's (tank_rotations-20261104-13).
The search is round 2's (rotopt knobs and `-talent-search -require feralSwiftness` at 1,000 iterations a
candidate, 3,000 to confirm, the six scoring bosses at the preset's weight), then every gear, enchant and
consumable choice A/B'd with abtest (Gear above), and every choice resting on a small paired difference
confirmed on a disjoint seed. Raw output: `orchestration/tank_handoff/druid/round3/` (`ab_summary.txt`
indexes the abtest files; `rotopt/` the knob and talent searches; `tankbench/` the bench).

- **Talents**: `-5022233202310212551-055` at every weight, again (searched before and after the gear changes).
- **Max threat**: Enrage's Rage knob widened to 100 (tank_rotations-20261104-12). Enrage under 70 Rage is
  +4.0 +- 0.5 over 60 on seed 2 (80 and 100 the same, +4.0 and +3.9), so it takes 70, the search's pick.
  Strength enchants on the head, legs, wrist and weapon (+30.1 +- 2.9). The rest as round 2.
- **Balanced**: the rotation holds (Roar refreshed under 5 sec looked +5.7 at 3,000 on seed 1 and was -1.6 +- 5.4
  and -1.4 +- 4.9 on seeds 2 and 3). Superior Deflection on the wrist.
- **Max survival**: the rotation holds. Superior Deflection on the wrist, Dirge's chops for the Dumpling.
- **Judgement of Wisdom and the shift** (tank_mechanics-20261104-12's follow-up): with mana threat unmultiplied,
  the raid's Judgement of Wisdom is still worth 34 / 38 / 53 (Max threat / Balanced / Max survival, 3,000,
  without it against with it), and the pre-pull shift 25.6 / 25.5 / 22.0, every point of it through Judgement of
  Wisdom's mana (without the debuff, dropping the shift changes nothing). Both stay; the shift's value rests on
  the to-verify mana rule (call 6).
- **Race** (seed 3, 10,000, on the final Max survival): the Tauren is -19.8 +- 11.3 behind the Night Elf, a tie
  that leans Night Elf (seed 2 at 3,000 on round 2's build: +0.8 +- 22.4). At Max threat and Balanced the Tauren
  loses 161 +- 3 and 143 +- 7 (seed 2). Night Elf at every weight.
- **Before and after** (own weight; tankbench 1,000 iterations on seed 1, so the "before" is master's re-bench
  after the engine fix; the gain is paired, abtest at 3,000 on seed 2):

  | Preset | Round 2 build | Round 3 | Paired gain | 3,000 (rotopt confirm, seed 1) |
  |---|---|---|---|---|
  | Max threat | 2,808.5 | 2,842.5 | +33.1 +- 2.8 | 2,844.6 |
  | Balanced | 1,839.0 | 1,851.8 | +8.6 +- 5.2 (the wrist: +6.1 +- 2.7 at 10,000) | 1,849.1 |
  | Max survival | -775.0 | -633.8 | +102.8 +- 15.2 | -655.6 |

- **Each preset wins its own weight** (the other two builds against it, paired, 3,000, seed 2), so no two are the
  same build:

  | Weight | Max threat build | Balanced build | Max survival build |
  |---|---|---|---|
  | w=1/3 | **own** | -95.0 +- 5.2 | -259.5 +- 5.3 |
  | w=1 | -127.3 +- 9.4 | **own** | -91.9 +- 9.0 |
  | w=3 | -920.1 +- 26.3 | -125.8 +- 26.5 | **own** |

### Tankbench (round 3)

`tools/tankbench`, 1,000 iterations, all seven encounters (`orchestration/audit/ref/tank_bench/summary.tsv`).
Mean of the six scoring encounters:

| Preset | TPS | DPS | SafeDPS | DTPS | Healing needed | Score at its own weight |
|---|---|---|---|---|---|---|
| Max threat | 1,688 | 644 | 1,699 | 595 | 618 | **2,842.5** |
| Balanced | 1,608 | 610 | 1,589 | 569 | 483 | **1,851.8** |
| Max survival | 1,453 | 543 | 1,491 | 570 | 442 | **-633.8** |

The bear's DPS is about 610 at Balanced, down from about 755 in round 1: the white hits' flat 11.25 Rage is less
than Classic's Rage from damage gave at this gear, and the boss parries and blocks (round 2). Per encounter:

| Encounter | Preset | Score (own weight) | TPS | DPS | SafeDPS | DTPS | Healing needed | Death floor |
|---|---|---|---|---|---|---|---|---|
| Magmadar | max threat | 2936 | 1645 | 632 | 1650 | 363 | 313 | 0.0% |
| Magmadar | balanced | 2261 | 1577 | 603 | 1560 | 346 | 247 | 0.0% |
| Magmadar | max survival | 788 | 1405 | 528 | 1446 | 346 | 201 | 0.0% |
| Golemagg | max threat | 2889 | 1642 | 633 | 1652 | 390 | 358 | 0.0% |
| Golemagg | balanced | 2111 | 1571 | 601 | 1549 | 370 | 283 | 0.0% |
| Golemagg | max survival | 383 | 1396 | 526 | 1432 | 376 | 237 | 0.0% |
| Onyxia | max threat | 2773 | 1728 | 656 | 1737 | 482 | 407 | 0.0% |
| Onyxia | balanced | 1701 | 1643 | 621 | 1643 | 462 | 307 | 0.0% |
| Onyxia | max survival | -905 | 1497 | 558 | 1554 | 462 | 259 | 0.0% |
| Broodlord | max threat | 2829 | 1692 | 647 | 1691 | 496 | 553 | 0.0% |
| Broodlord | balanced | 1912 | 1610 | 613 | 1587 | 471 | 421 | 0.0% |
| Broodlord | max survival | -703 | 1456 | 546 | 1495 | 471 | 410 | 0.0% |
| Vaelastrasz | max threat | 3161 | 1793 | 658 | 1859 | 1203 | 1327 | 0.0% |
| Vaelastrasz | balanced | 2210 | 1671 | 609 | 1697 | 1164 | 939 | 0.0% |
| Vaelastrasz | max survival | 95 | 1619 | 588 | 1666 | 1163 | 856 | 0.0% |
| Chromaggus | max threat | 2467 | 1630 | 635 | 1604 | 636 | 747 | 0.0% |
| Chromaggus | balanced | 916 | 1578 | 613 | 1500 | 601 | 702 | 0.0% |
| Chromaggus | max survival | -3461 | 1345 | 512 | 1351 | 599 | 692 | 0.0% |
| Ragnaros | max threat | 946 | 1628 | 632 | 1610 | 1323 | 2160 | 0.0% |
| Ragnaros | balanced | -3541 | 1583 | 613 | 1529 | 1306 | 2090 | 0.1% |
| Ragnaros | max survival | -15634 | 1361 | 517 | 1386 | 1305 | 1943 | 0.0% |

Ragnaros needs about 1,950 to 2,150 healing a second at every preset, past what the bench's healers give, so
it stays out of the score.

## Presets (round 2, 2026-11-04: the tank-facing engine and the audit's fixes)

Round 2 re-searched the three presets on master 5668ea5's engine (the boss faces the tank and parries and
blocks, the logs' swing gap with parry haste off, no health read in a scoring run) with this round's bear
fixes: the white hit's flat 11.25 Rage (22.5 on a crit), Primal Bite's 2x threat, Demoralizing Roar below
Lacerate's keep-up line (`roarAfterLacerate`), and the chest's budget. Scores are the mean of the six scoring
encounters (Magmadar, Golemagg, Onyxia, Broodlord, Vaelastrasz, Chromaggus; Ragnaros is benched, not
scored) at the preset's weight, 1,000 iterations a candidate, 3,000 to confirm; the +- is the paired error.

- **Talents.** Unchanged at every weight: `-5022233202310212551-055` (Feral Swiftness 2 pinned, standards
  26). The talent searches, rerun with the final knobs, moved nothing.
- **Shared knobs.** Roar below Lacerate's keep-up line, refreshed under 3 sec; Primal Bite on cooldown;
  Lacerate refreshed under 4 sec; a shift into Dire Bear Form 1.5 sec before the pull; Barkskin as the
  filler from 15 sec in; and **Enrage in the fight, not before the pull** (under a Rage threshold). The
  audit's pre-pull Enrage at -1 sec does beat -3 sec (+50.7 +- 16.0 at 3,000, tank_rotations-20261104-5), but
  Enrage in the fight beats both at every weight.
- **Max threat (w=1/3)**: Lacerate as the filler, Maul at 20 Rage, Enrage under 60 Rage, no Frenzied
  Regeneration. 2,836.6 at 3,000 iterations.
- **Balanced (w=1)**: Lacerate as the filler, Maul at 20 Rage, Enrage under 40 Rage, Frenzied Regeneration
  at 100 Rage (at the pull's Rage peak, about once a fight). 1,873.5 at 3,000. Maul at 10 with Roar refreshed
  under 8 looked +7.6 at 3,000 and was -1.6 +- 2.9 at 10,000, so it stays out.
- **Max survival (w=3)**: Lacerate kept at 5 stacks but not the filler, Maul at 10 Rage, Enrage under 40,
  Frenzied Regeneration at 100 Rage. -736.1 at 3,000.
- **Before and after** (own weight, 1,000 iterations; "before" is round 1's build on this code with the old
  Roar order, then with Roar moved):

  | Preset | Round 1 build | + Roar moved | Round 2 | Gain |
  |---|---|---|---|---|
  | Max threat | 2,586.8 | 2,804.6 | 2,837.3 | +250.6 +- 14.1 |
  | Balanced | 1,538.3 | 1,729.5 | 1,867.6 | +329.2 +- 18.7 |
  | Max survival | -1,166.4 | -896.0 | -750.2 | +416.2 +- 48.5 |

- **Each preset wins its own weight** (3,000 iterations; the gap to that weight's preset in brackets), so no
  two are the same build (tank_fairness-20261104-4, and the core worker's Max survival over Balanced at w=1
  is gone):

  | Build | w=1/3 | w=1 | w=3 |
  |---|---|---|---|
  | Max threat | **2,836.6** | 1,763.3 (-110 +- 9) | -1,456.8 (-721 +- 27) |
  | Balanced | 2,769.1 (-68 +- 5) | **1,873.5** | -813.4 (-77 +- 25) |
  | Max survival | 2,625.2 (-211 +- 5) | 1,784.9 (-89 +- 9) | **-736.1** |

- **Race** (3,000): Night Elf, +137 over the next race at Max threat and +134 at Balanced. At Max survival the
  Tauren ties it (-1.8 +- 11.8 at 10,000); Night Elf is kept. `sim.ts` still defaults the Horde to Skyborne
  Windshaper.
- **Gear choices.** The armor trinket (990815) still loses at every weight (-238 / -207 / -94); Idol of the
  Ursine Twins stays (Brutality ties at Balanced, +0.8 unpaired, and loses at the others; Enraged Idol loses
  everywhere). Enchants and consumables were re-checked in round 3 (Gear).

## Calls for the user

Each with the default the sim uses now.

1. **Bear Form threat at the notes' +115% (x2.15)**, though client 70753's 21178 still says 50%. Notes-only,
   gated under Forever (standards 37). Default: the notes. A bear tank's log would settle it.
2. **Rend and Tear at the notes' 3% a point**, though the tree still says 2%. Default: the notes.
3. **No threat beyond the form's** for Maul, Swipe and Lacerate (Classic had 1.75x on Maul and Swipe).
   Default: none, as no row states one. Primal Bite takes the notes' "roughly doubled" as exactly 2x (notes
   only, no client row; standards 37). Default: 2x; a bear tank's threat log would settle it.
4. **Enrage's armor cut at the tooltip's 16%** of armor from items. Default: 16%.
5. **Omen of Clarity in Bear Form at the cat's 2 a minute**. Default: 2 a minute.
6. **Threat from mana gains.** The engine counts the mana that Judgement of Wisdom gives back as threat,
   0.5 a mana with no threat multiplier (it was multiplied by the bear's 2.15 until 2026-11-04; about 8 TPS then, about
   2 now; standards 40 keeps it unmultiplied until a Forever log shows otherwise). Judgement of Wisdom is still
   worth 34 / 38 / 53 to the presets, and the pre-pull shift's whole 22 to 26 rests on it (round 3). That is an
   engine rule for every tank, so not mine to change. Default: as it is.
7. **Furor scores nothing** for a bear that shifts before the pull, so the presets drop it. Default: Furor 0.
8. **Enrage in the fight, not before the pull**, at every weight (the round-1 presets cast it before the
   pull). Its armor cut lands mid-fight under a Rage threshold. Default: as the search found.
9. **Max survival's race**: the Tauren ties the Night Elf (round 3: -19.8 +- 11.3 at 10,000 on a disjoint seed;
   round 2: -1.8 +- 11.8). Default: Night Elf, as the other two presets.
10. **Consumables no tank carries.** A Blasted Lands buff (Phase 1 quest items, standards 8) would add to every
   preset: Ground Scorpok Assay +42 / +53 / +112, R.O.I.D.S. +52 / +50 / +47, Lung Juice Cocktail +9 / +36 / +127
   (Max threat / Balanced / Max survival, 3,000), and Rumsey Rum Black Label +6 / +21 / +75. The DPS presets
   take Blasted Lands buffs (standards 11, the full raid treatment); none of the four tank requests does, so
   the bear's presets leave them out to stay comparable with the other tanks. Default: none, for every tank
   alike; taking them is a call for all four at once.
