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
| Bear threat (Bear Form (Passive2) 21178) | A_MOD_THREAT 50 on every school: x1.5 | x1.5 | right, tested (Maul, Primal Bite, Roar) |
| Bear swing | SpellShapeshiftForm 8: 2,500 ms, DamageVariance 0.4 | the weapon's DPS at 2.5 sec, 20% either side, the enchant's and stone's flat damage on top (T2: wowsims/forever b5d2e0b992) | right, tested |
| Rage from the swing | not in the client | the two-hander's 4.5 a second of swing whatever the weapon (T2: their log, 11.25 a 2.5 sec swing; standards 21) | right |
| Rage from damage taken | not in the client | standards 21's rule (engine) | engine's |
| The shift into Bear Form | Rage starts at 0 in the form (T2: their forms.go) | **was missing**: now a cast of Bear Form (also from Bear Form) spends the Rage and Furor adds its 10 | fixed, tested |
| Omen of Clarity (16864, baseline from 20) | Clearcasting 16870's class mask takes in Maul and Swipe (0x800), Demoralizing Roar (0x8), Bash (0x2000), Primal Bite and Lacerate (word 1 0x40, 0x100) | **the bear's attacks had no Omen flag, so a proc in Bear Form did nothing**; now they're free under it and use it up (Maul as it lands, since it's the swing) | fixed, tested |
| Omen's rate | the server's | 2 a minute at the swing speed (the cat's rule): 8.3% a 2.5 sec swing, 10 sec cooldown | A (the cat's rule, T2 for the cat) |

### Abilities

| Ability | Client (T1) | Sim | Verdict |
|---|---|---|---|
| Maul rank 7 (9881, 58) | weapon damage + 128, 15 Rage, next swing, 80% back on a miss (Attributes_1 0x8000000), no threat effect | the same; threat its damage x1.5 | right, tested; its threat beyond the form is a call |
| Swipe rank 5 (9908, 54) | 83 to 3 enemies, 20 Rage, no refund, no threat effect | 83 + 3% of attack power (T2: their tester fit, PR #149) | right, tested; the 3% is T2 |
| Demoralizing Roar rank 5 (9898, 52) | -193 - 1.4 a level attack power (204 at 60) for 30 sec, 10 Rage, effect 1 THREAT 42 + 0.8 a level (48.4 at 60) | the same, x1.5 | right, tested |
| Enrage (5229, 12) | effect 1: 10 Rage at once; effect 0: 2 Rage a second for 10 sec; 1 min cooldown, no GCD; the armor cut is a dummy (effect 3) the server runs; the tooltip says 27% in Bear Form, 16% in Dire Bear Form; 25503 (no form) has -26% armor | 10 + 20 Rage, armor from items x0.84 in Dire Bear Form (x0.73 in Bear Form). **It lost its tenth tick** to the aura's expiry (28 Rage, not 30) | fixed, tested; the 16% is a call |
| Frenzied Regeneration (22842, 36; Forever's only rank) | 10 sec, effect 0 periodic 1 sec, "up to 10 Rage a second, each into $s1% health" with s1 1; 3 min cooldown, GCD | 1% of max health a point of Rage (T2 reads it the same) **lost its tenth tick** the same way | fixed, tested |
| Lacerate (talent; 1328617 at 25, 1328616 at 34, 414644 at 42, 1235826 at 50, 1235827 at 58) | ticks of 4 / 6 / 10 / 12 / 15 a stack every 3 sec for 15 sec, 5 stacks, effect 1 "10% weapon damage a stack already on", 15 Rage, refund on a miss, no threat effect | **ranks 1 and 2 were missing** (no Lacerate under 42); Rend and Tear **didn't raise its ticks** (it does Rake's and Rip's, and wowsims/forever's Lacerate's) | fixed, tested |
| Primal Bite (talent; SkillLineAbility: 407995 at 40, 1238069 at 50, 1238070 at 60) | weapon damage + 42 / 62 / 77, 20 Rage, 6 sec cooldown, refund on a miss, no threat effect | **was 26 / 38 / 59 / 77 at 25 / 36 / 48 / 60, and 1238073 at 60**, a row no trainer or talent teaches (the 77 is the same) | fixed, tested |
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
| Rend and Tear 5 | 2% a point on the melee abilities against a bleeding target (a dummy) | Maul, Swipe, Primal Bite, Lacerate's hit and now its ticks; Lacerate's bleed counts | right, tested (the ticks T2) |
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

- **No threat beyond the form's 1.5x** for Maul, Swipe, Lacerate and Primal Bite. No row states one and
  their tooltips' "high amount of threat" isn't a number; wowsims/forever (T2) also gives damage only
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

- **Form**: Dire Bear Form's health, attack power, armor x4.6 from items and threat x1.5 (on against
  cancelled); Bear Form at 30.
- **Every ability**: Maul (queued, struck in the swing's place, damage range +128, threat 1.5x, cost with
  Ferocity and the idol), Swipe (83 + 3% of attack power, 3 of 4 targets, 20 Rage, Feral Instinct and Savage
  Fury, the idol), Demoralizing Roar (204 attack power, 30 sec, both targets, 10 Rage, 48.4 x 1.5 threat),
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

What the searches found (the details per preset below):

- **The opener is most of SafeDPS.** SafeDPS is each fight's lowest moment for a reference DPS, and that is
  the pull, so what the bear has in its first seconds counts far over the rest: a shift into Dire Bear Form
  1.5 sec before the pull for Furor's Rage (+65 at Balanced), Enrage 1 sec before the pull (its Rage at the
  pull, its armor cut over the first 9 sec; earlier, the Mighty Rage Potion at the pull runs into the 100
  Rage cap), the on-use Strength trinket at the pull (+166 over the armor trinket), and the Night Elf's
  Elune's Light (+160 over the Tauren).
- **Barkskin on its timer, as the filler** (after Primal Bite and Lacerate's keep-up): +108 at Balanced.
  Ahead of the attacks it costs the opener's GCD (-145).
- **Lacerate is the filler** (to 5 stacks, refreshed under 4 sec, then cast whenever the GCD is free) and
  Primal Bite goes on cooldown; Maul at 30 Rage. Swipe never beats Lacerate on one target, and Frenzied
  Regeneration on a timer loses at every weight (its Rage is threat).

## Gear (`tools/database/forever_synthetic_items_druid_tank.go`)

Standards 36's bear row: physical archetype (3 lines a piece, the cat's budgets), the feral two-hander at
0.90 x L(76), 729.5 weighted points against 729.3. The row's small shares (Intellect, crit, attack and spell
power, 4.8 points) went to the rest in proportion. Hit by the tank objective: the Flask of Natural Accuracy
(5% hit) is the best flask (Balanced: +105 over the Titans; with the gear's hit at 8%, +35 over Natural
Aggression; Max survival likewise, Max threat a tie with Aggression), so the set carries 3% and its 1.6%
over the 8% cap went into Agility (+43, against +27 for Defense and +18 for Stamina). 256 Stamina, 164
Agility, 122 Strength, 105 Defense, 3% hit.

- **Trinkets**: the tank stat trinket (23 Defense and 1% dodge, 46.5 weighted, 1.3% over 1b's 45.9 since no
  whole Defense and dodge fall in the 1% band) and, of standards 36's two on-use kinds, the cat's Strength
  on-use (165 Strength for 20 sec every 2 min). The defensive kind (990815, 1,665 bonus armor, made for the
  comparison) scores 166 lower at Balanced and lower at every weight: its armor never touches healing needed
  (the bad moments come later in a fight), while the Strength lifts the opener's threat.
- **Idol**: Idol of the Ursine Twins, then Idol of Brutality (-19) and Enraged Idol (-36).
- **Enchants** (each against its tank alternative at Balanced, 2,000 iterations, the same seeds): Agility on
  the head and legs (Lesser Arcanum of Voracity, +23 over Stamina), Superior Strength on the gloves (+12 over
  Greater Agility), Greater Agility on the boots (+12 over Greater Stamina), Agility on the back (+8 over
  Superior Defense), Greater Stats on the chest (+13 over Major Health), Agility on the neck (+3 over
  Strength), Superior Stamina on the wrist (a tie with Superior Strength, so the tank one), +25 Agility on
  the weapon.
- **Consumables**: the Flask of Natural Accuracy, Elixir of the Phalanx (+112), Elixir of Greater Fortitude
  (+46), the Mongoose (taking it out costs 124), Juju Power and Might, Smoked Desert Dumpling (Dirge's
  Stamina chops -5, Grilled Squid -15), Dragonbreath Chili (+28), a Dense Weightstone (the mace's stone; the
  bench's old request had a sharpening stone on it) and the Mighty Rage Potion at the pull (Greater
  Stoneshield loses 650: no Rage at the pull).

PRESETS_PLACEHOLDER
