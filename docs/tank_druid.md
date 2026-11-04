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
| Rage from the swing | not in the client | the two-hander's 4.5 a second of swing whatever the weapon (T2: their log, 11.25 a 2.5 sec swing; standards 21) | right |
| Rage from damage taken | not in the client | standards 21's rule (engine) | engine's |
| The shift into Bear Form | Rage starts at 0 in the form (T2: their forms.go) | **was missing**: now a cast of Bear Form (also from Bear Form) spends the Rage and Furor adds its 10 | fixed, tested |
| Omen of Clarity (16864, baseline from 20) | Clearcasting 16870's class mask takes in Maul and Swipe (0x800), Demoralizing Roar (0x8), Bash (0x2000), Primal Bite and Lacerate (word 1 0x40, 0x100) | **the bear's attacks had no Omen flag, so a proc in Bear Form did nothing**; now they're free under it and use it up (Maul as it lands, since it's the swing) | fixed, tested |
| Omen's rate | the server's | 2 a minute at the swing speed (the cat's rule): 8.3% a 2.5 sec swing, 10 sec cooldown | A (the cat's rule, T2 for the cat) |

### Abilities

| Ability | Client (T1) | Sim | Verdict |
|---|---|---|---|
| Maul rank 7 (9881, 58) | weapon damage + 128, 15 Rage, next swing, 80% back on a miss (Attributes_1 0x8000000), no threat effect | the same; threat its damage x2.15 (the form) | right, tested; its threat beyond the form is a call |
| Swipe rank 5 (9908, 54) | 83 to 3 enemies, 20 Rage, no refund, no threat effect | 83 + 3% of attack power (T2: their tester fit, PR #149) | right, tested; the 3% is T2 |
| Demoralizing Roar rank 5 (9898, 52) | -193 - 1.4 a level attack power (204 at 60) for 30 sec, 10 Rage, effect 1 THREAT 42 + 0.8 a level (48.4 at 60) | the same, x2.15 | right, tested |
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

- **No threat beyond the form's 2.15x** for Maul, Swipe, Lacerate and Primal Bite. No row states one and
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

What the searches found (the details per preset below):

- **The opener is most of SafeDPS.** SafeDPS is each fight's lowest moment for a reference DPS, and that is
  the pull, so what the bear has in its first seconds counts far over the rest: a shift into Dire Bear Form
  1.5 sec before the pull (+62 at Balanced even with Furor at 0, since the engine counts the mana that
  Judgement of Wisdom gives back as threat; +90 with Furor), Enrage 1 sec before the pull (its Rage at the
  pull, its armor cut over the first 9 sec; earlier, the Mighty Rage Potion at the pull runs into the 100
  Rage cap), the on-use Strength trinket at the pull (+166 over the armor trinket), and the Night Elf's
  Elune's Light (+172 over the Tauren after the update).
- **Barkskin on its timer, as the filler** (after Primal Bite and Lacerate's keep-up): +159 at Balanced.
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

## Presets (2026-11-04, client 70753 and the November 2 notes)

Three presets, all Night Elf on the synthetic set above with one talent build, `-5022233202310212551-055`.
Each was found by a talent search and then a knob search (`tools/rotopt -template druid_tank -tank <w>`) on
the six scoring encounters (Magmadar, Golemagg, Onyxia, Broodlord, Vaelastrasz, Chromaggus): 600 iterations
a candidate, confirmed at 3,000, then a race check at 3,000. The requests are
`tools/tankbench/req/druid_bear_{max_threat,balanced,max_survival}.json` (`druid_bear_balanced.json`
replaces `druid_bear.json`, and `sim/druid/tank/tank_test.go` reads it). The APLs are
`ui/feral_tank_druid/apls/*.apl.json`, and `presets.ts` builds them.

- **Talents.** Every weight moved Furor's 5 points out: to Shredding Attacks 3 and Brutal Impact 2 (+19 at
  Balanced over the old build at 3,000 iterations; Furor 0 is +13.6 over Furor 2). With the prepull shift
  done 1.5 sec early, Furor's 10 Rage a shift is worth nothing, and at Max survival a point scores -0.6.
  Feral Instinct (Swipe, never cast) and Brutal Impact (no Bash in the rotation) are dead points kept only
  because the tier needs them. Lacerate, Primal Bite, Rend and Tear 5, Natural Reaction 5, Blood Frenzy 2,
  Thick Hide 3 and Predatory Strikes 3 all stay at every weight.
- **Balanced (w=1)**: the knobs didn't move from the earlier search. Demoralizing Roar refreshed under 3
  sec, Primal Bite on cooldown, Lacerate to 5 stacks and then the filler, refreshed under 4 sec, Maul at 30
  Rage, Enrage at the pull and 1 sec before it, a shift 1.5 sec before the pull, and Barkskin as the filler.
  What each costs taken out: Barkskin -159, the prepull shift -62, Enrage -24, Maul at 40 Rage -11, at 80
  -144 (25 to 30 is flat). The Swipe threshold has no effect, since Lacerate fills every GCD.
  Frenzied Regeneration at 100 Rage is -3.
- **Max threat (w=1/3)**: Balanced with Maul at 20 Rage (2,910.6 at 3,000 iterations, against about 2,903).
- **Max survival (w=3)**: Balanced with Frenzied Regeneration at 100 Rage and Enrage 3 sec before the pull,
  so its armor cut ends sooner (-857.1 against -1,250.9 at 3,000 iterations).
- **Race.** Night Elf is best at every weight (six-encounter mean, 3,000 iterations):

  | Race | Max threat | Balanced | Max survival |
  |---|---|---|---|
  | Night Elf | 2,912.0 | 1,844.6 | -859.5 |
  | Tauren | 2,729.3 | 1,672.4 | -971.7 |
  | Skyborne (High Order and Windshaper alike) | 2,762.9 | 1,679.9 | -1,100.8 |

  The best Horde race is a Skyborne at Max threat and Balanced, and the Tauren at Max survival.
  `sim.ts` defaults the Horde to Skyborne Windshaper.

### Tankbench (step 7)

`tools/tankbench`, 1,000 iterations, all seven encounters. The score columns are the tank objective at each
weight (2 x SafeDPS - w x 900 x healing needed / one healer's HPS). Mean of the six scoring encounters
(Ragnaros out):

| Preset | TPS | DPS | SafeDPS | DTPS | Healing needed | Max threat | Balanced | Max survival |
|---|---|---|---|---|---|---|---|---|
| Max threat | 1,702 | 756 | 1,715 | 569 | 583 | **2,905** | 1,857 | -1,290 |
| Balanced | 1,700 | 755 | 1,714 | 570 | 582 | 2,904 | **1,855** | -1,292 |
| Max survival | 1,627 | 722 | 1,598 | 568 | 487 | 2,748 | 1,850 | **-844** |

Max threat and Balanced are the same preset within noise (the Maul threshold is flat from 20 to 30 Rage).
Max survival gives up 116 SafeDPS for 95 less healing needed, and wins at w=3 on every encounter except
Chromaggus, where it saves no healing (703 against 705) and only loses SafeDPS. The death floor is 0% everywhere at these healer numbers. The other tanks had no committed
preset numbers when this ran; they go side by side in `docs/tank_sim.md`.

Per encounter:

| Encounter | Preset | TPS | DPS | SafeDPS | DTPS | Healing needed | Death floor | Max threat | Balanced | Max survival |
|---|---|---|---|---|---|---|---|---|---|---|
| Magmadar | max threat | 1659 | 741 | 1683 | 344 | 287 | 0.0% | 3031 | 2363 | 358 |
| Magmadar | balanced | 1651 | 737 | 1679 | 345 | 289 | 0.0% | 3021 | 2348 | 330 |
| Magmadar | max survival | 1589 | 709 | 1565 | 345 | 238 | 0.0% | 2853 | 2299 | 637 |
| Golemagg | max threat | 1650 | 738 | 1680 | 373 | 343 | 0.0% | 2962 | 2166 | -225 |
| Golemagg | balanced | 1651 | 739 | 1681 | 375 | 341 | 0.0% | 2966 | 2174 | -201 |
| Golemagg | max survival | 1599 | 715 | 1575 | 371 | 272 | 0.0% | 2834 | 2202 | 306 |
| Onyxia | max threat | 1743 | 772 | 1737 | 460 | 384 | 0.0% | 2813 | 1488 | -2486 |
| Onyxia | balanced | 1741 | 771 | 1737 | 460 | 384 | 0.0% | 2812 | 1487 | -2490 |
| Onyxia | max survival | 1648 | 729 | 1625 | 455 | 311 | 0.0% | 2714 | 1640 | -1580 |
| Broodlord | max threat | 1703 | 758 | 1704 | 478 | 527 | 0.0% | 2882 | 1829 | -1331 |
| Broodlord | balanced | 1702 | 758 | 1704 | 479 | 529 | 0.0% | 2879 | 1820 | -1357 |
| Broodlord | max survival | 1620 | 720 | 1609 | 476 | 481 | 0.0% | 2736 | 1773 | -1115 |
| Vaelastrasz | max threat | 1794 | 777 | 1840 | 1158 | 1247 | 0.0% | 3155 | 2107 | -1038 |
| Vaelastrasz | balanced | 1794 | 777 | 1840 | 1158 | 1247 | 0.0% | 3155 | 2107 | -1038 |
| Vaelastrasz | max survival | 1673 | 721 | 1673 | 1159 | 919 | 0.0% | 2960 | 2188 | -128 |
| Chromaggus | max threat | 1662 | 750 | 1645 | 600 | 708 | 0.0% | 2589 | 1188 | -3017 |
| Chromaggus | balanced | 1659 | 748 | 1643 | 600 | 705 | 0.0% | 2589 | 1194 | -2993 |
| Chromaggus | max survival | 1633 | 736 | 1543 | 603 | 703 | 0.0% | 2390 | 997 | -3182 |
| Ragnaros | max threat | 1657 | 747 | 1653 | 1307 | 2070 | 0.0% | 1127 | -3232 | -16308 |
| Ragnaros | balanced | 1656 | 746 | 1652 | 1307 | 2070 | 0.0% | 1125 | -3233 | -16309 |
| Ragnaros | max survival | 1614 | 727 | 1561 | 1311 | 2070 | 0.0% | 943 | -3415 | -16487 |

Ragnaros needs 2,070 healing a second at every preset, past what the bench's healers give, so it stays
out of the score.

## Calls for the user

Each with the default the sim uses now.

1. **Bear Form threat at the notes' +115% (x2.15)**, though client 70753's 21178 still says 50%. Notes-only,
   gated under Forever (standards 37). Default: the notes. A bear tank's log would settle it.
2. **Rend and Tear at the notes' 3% a point**, though the tree still says 2%. Default: the notes.
3. **No threat beyond the form's** for Maul, Swipe, Lacerate and Primal Bite (Classic had 1.75x on Maul and
   Swipe). Default: none, as no row states one.
4. **Enrage's armor cut at the tooltip's 16%** of armor from items. Default: 16%.
5. **Omen of Clarity in Bear Form at the cat's 2 a minute**. Default: 2 a minute.
6. **Threat from mana gains.** The engine counts the mana that Judgement of Wisdom gives back as threat
   (about 8 TPS here), which is most of why the prepull shift pays. That is an engine rule for every tank,
   so not mine to change. Default: as it is.
7. **Furor scores nothing** for a bear that shifts before the pull, so the presets drop it. Default: Furor 0.
8. **Max threat and Balanced are one preset in effect.** Default: keep both, since the UI has a slot for
   each and they differ only in Maul's threshold.
