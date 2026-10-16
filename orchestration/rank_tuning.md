# Level 31+ rank tuning audit

Read-only audit by the rank tuning auditor. It covers the spell ranks and deep talents the Level 60 presets rely on that can't be learned under the beta's level 30 cap, and asks whether each looks tuned for Forever or stale. Pass 1 (2026-10-12 14:30 UTC) covered Demonology, Destruction, Affliction, Retribution, Fury, Rogue, Feral, Elemental and Smite. Pass 2 (15:40 UTC) adds Balance, Shadow, the mages, the hunters, Enhancement, Arms, pet abilities and the shared raid buffs. Between the passes no class branch changed rank data; workers added open-question notes only.

Workers must not change their sims on these verdicts alone; the orchestrator rules on any change.

## Method and sources

- **Forever client (tier 1):** wago.tools db2 for build 1.60.1.70338 (SpellEffect, SpellLevels, SpellPower, SpellMisc, SpellCastTimes, SpellCooldowns, SpellDuration, SkillLineAbility, Trait*, CurvePoint). Also SpellEffect for **every Forever beta build on wago.tools** (69876, 69893, 69913, 69977, 70009, 70058, 70094, 70124, 70170, 70178, 70205, 70235, 70245, 70291, 70334, 70338), plus the trait and curve tables for 69876 / 70009 / 70170 / 70291, so we can see which ranks Blizzard actually edited during the beta.
- **Era and SoD (tiers 2/3):** wago.tools 1.15.9.70003, the Classic Era client. SoD realms run the same client, and the SoD rune spells (Lava Burst 408490, Haunt, Mutilate 399960, Incinerate 412758, Lightning Bolt overloads 408439...) sit in its tables. SoD's server-side baseline-rank changes are not in the client and were not checked this pass; for the baseline ranks below, "Era" means the 1.15.9 client values.
- **Era average damage:** base + (die + 1) / 2. Forever stores the average (EffectBasePointsF) and a Variance (spread / average). For spells Forever left alone the two match exactly, which validates the comparison.
- **DPS sensitivities:** the shipped preset, best race, 180 sec, 20,000 iterations, with a local patch that rescales only the spell's level 31+ ranks (env var hooks in an uncommitted worktree). Baseline numbers are my own runs of the same inputs.
- **No tooltip-only decisions.** Nothing comes from vmangos.

## The main finding (applies to every class)

1. **The 31+ ranks were not left at Era values.** Of 328 class rank chains with a 31+ rank, only a handful have 31+ entries byte-identical to Era while the ≤30 ranks changed (Seal / Judgement of Righteousness, Slam, Swipe, Immolation Trap, Shadowguard, Touch of Weakness, Mana Tide). In those, the ≤30 change is only Forever's removal of the below-level-20 coefficient penalty, which never applied to 31+ ranks anyway. Every caster damage spell's 31+ ranks were rewritten in the same design pass as its ≤30 ranks.
2. **Forever used two curve shapes, both applied to every rank.**
   - **Multiplicative:** about 0.4–0.6× Era at every rank from the high 20s up. Shadow Bolt, Immolate, Corruption, Curse of Agony, Searing Pain, Conflagrate, Fireball, Scorch, Smite, Holy Fire, Wrath, Starfire, Moonfire, Lightning Bolt, Chain Lightning, the shocks.
   - **Subtractive:** Era minus an offset that grows with level and plateaus near level 44. It ends around 0.9× Era at 60. Mind Blast and Exorcism (−32 a hit from L44), Mind Flay (−12 a tick), Shadow Word: Pain (−15), Frostbolt (−60), Fire Blast, the heals.
   - The subtractive plateau sits at level 44, not 30, so it isn't a stale break at the beta cap.
3. **The beta's balancing passes touched only the playable ranks, except for design changes.** Build 70291, the last balance pass, retuned the ≤30 ranks of Shadow Bolt (ranks 1–4), Fireball (1–5), Frostbolt (1–5), Smite (1–4), Lightning Bolt (1–5), Wrath (1–5) and Lesser Heal / Heal, and left every 31+ rank alone. Changes to a spell's design hit all ranks: Wrath ×1.5 everywhere in 70009, Bloodthirst AP 35→45% at all four ranks in 70291, Penance's rework, Holy Strike's weapon % and the Lava Burst raise in 70009, Sunder, Thorns, Retribution Aura, Disengage. **So a 31+ rank is current for the spell's design but has never had the leveling-curve fine-tuning the ≤30 ranks get.** In most spells that fine-tuning moved the top playable rank by 0–3%; Lightning Bolt (+12%) and Frostbolt (+8%) are the exceptions.
4. **Deep talents:** across the five sampled builds, only three 25-point talents changed: Holy Power (crit 2→3% per point, 70009), Champion of the Light (33/66/100 → 20/40/60%, 70170) and Improved Slam. **No 31-point talent (gate 30) changed in any beta build.** The deep-tier values are Forever's own, with rank counts and per-rank values different from Era (e.g. Flurry 5–25%, Elemental Fury 5 × 20%), but untested.
5. **Cross-spell check at level 60 (base damage per cast-second, at the top ≤30 rank → the rank the preset casts):**

   | Spell | Top ≤30 rank | Level 60 rank | Growth | Era growth |
   |---|---|---|---|---|
   | Shadow Bolt | 26.0/s | 83.7/s (r9) | ×3.2 | ×3.2 |
   | Fireball | 43.7/s | 138/s | ×3.2 | ×2.9 |
   | Frostbolt | 34.7/s | 158/s | ×4.6 | ×4.1 |
   | Starfire | 31.4/s | 109/s | ×3.5 | ×3.6 |
   | Scorch | 36.7/s | 119/s | ×3.2 | ×3.0 |
   | **Lightning Bolt** | 32.8/s | **78/s** | **×2.4** | ×3.3 |
   | **Smite** | 32.4/s | **68/s** | **×2.1** | ×2.5 |
   | **Wrath** | 22.5/s | **46/s** | **×2.0** | ×2.3 |
   | **Chain Lightning** (no ≤30 rank) | 44/s at L32 | **62/s** | **×1.4** | ×2.6 |

   Lightning Bolt, Chain Lightning, Smite and Wrath are the spells whose 31+ curve flattens against their own ≤30 trend and against the other Forever nukes. These are the stale-shaped candidates.

## Warlock: Demonology, Destruction, Affliction

| Spell (rank cast at 60) | Share of damage | Ranks checked | Verdict | Evidence | DPS sensitivity |
|---|---|---|---|---|---|
| Shadow Bolt IX (11661) | 55% Demo, 44% Destro, 53% Aff | 1–10 | **Tuned** | Every rank ≈0.5× Era (0.47–0.53 at 31+, 0.51 at L28); growth L28→L60 ×3.2, matching Era's ×3.2. 70291 retuned only ranks 1–4 and moved the L28 rank by nothing. | none needed |
| Corruption VI (11672) | 11–13% | 1–7 | **Tuned** | 0.49–0.53× Era at 31+ (0.59 at L24); coefficient ×1.2 over Era at every rank; growth 14% flatter than Era. | <1% |
| Immolate VII (11668) | 9–10% | 1–8 | **Tuned** | 0.54–0.57× at 31+ against 0.62 at L30; same coefficients at every rank. | <1% |
| Curse of Agony / Bane of Agony VI (11713) | 2% | 1–6 | **Tuned** | 0.50–0.53× at every rank from L28; coefficient ×1.6 at every rank. | negligible |
| Conflagrate (18932) | 9% Destro | 1–6 | **Tuned** | Two new Forever ranks (L25 95, L32 122) lead smoothly into the old ranks at 0.53–0.56× Era. | negligible |
| **Bane of Doom (603)** | 12% | L60 only | **Unclear (Forever design, untested)** | Base 1742 = 0.54× Era's 3200, in line with the warlock pattern. Spell power coefficient **4.0 against Era's client 1.0** (other Forever dots got ×1.2–1.6). Unchanged in all 16 beta builds. | **coefficient 2.0: −4.1% on every Shadow Bolt build** (Demo 1087→1043, Destro 985→945, Aff 976→935); coefficient 1.0: −6.1% |
| Incinerate, Wrack, Demonic Pact (31-point capstones) | Destro Fire / Aff / Demo | – | **Unclear (Forever design, untested)** | Incinerate 97/145/217 at L40/50/60 (SoD rune: 240 flat, 2.25 s); Wrack 36 a sec + 14.3%. No beta build changed them. | not run this pass |

## Paladin: Retribution (twist, sub-Holy, Shock)

| Spell | Share | Ranks | Verdict | Evidence | DPS sensitivity |
|---|---|---|---|---|---|
| Seal of Righteousness IX proc (25713) / Judgement of Righteousness (20286) | 15% / 6–11% | 1–9 | **Unchanged everywhere** | 31+ ranks byte-identical to Era. Forever's only ≤30 change floors the low ranks' coefficient at 10% / 50%, which 31+ ranks already had. | none |
| Seal / Judgement of Command (20920, 20966) | 12% / 6% | all | **Unchanged everywhere** | 70% weapon proc identical at every rank. | none |
| Consecration V (20924) | 15% | 1–5 | **Unclear** | Client stores a flat dummy (30) at every rank (server script). The wowhead Forever totals are 48 / 80 / 136 / 224 / 312 against Era 64 / 120 / 192 / 280 / 384: 0.67× at L30 rising to 0.81× at L60. Mana costs are still Era's at every rank. | holding the L30 ratio (×0.82 at 31+): −0.8% |
| Holy Strike VIII (10333), new Forever strike | 9% | 1–8 | **Stale-shaped flat part** | 70009 rewrote the weapon % at every rank (25→50%, smooth). The flat Holy part goes 12 / 17 / 19 / 24 at ≤30 (+0.1–0.3 a level), then 31 / 53 / 68 / 93 with per-level growth 1.0–3.2: a jump right past the beta cap. | flat part held to its ≤30 slope (×0.52): −0.4% |
| Light's Vigil (1311598) | 9% Shock | L60 | **Unclear (31-point, untested)** | 395 + 42.9%; the 40/50 ranks were deleted in 70009. | not run |
| Holy Power / Champion of the Light (25-point talents) | – | – | **Tuned** | The only deep talents Blizzard has retuned in the beta. | – |
| Seal of Fury (talent, not used by the presets) | – | 1–7 | **Data errors at 31+** | Rank 4 (20419) coefficient 0.900 against 0.090 on the other ranks; rank 7 (20423) 0; rank 5 (20421) has no trigger spell. Judgement of Fury rank 2 is 0.200 against 0.180. | – |

Retribution with both rescales: twist 962→951, sub-Holy 957→946, Shock 936→925 (−1.2%).

## Warrior: Fury

| Spell | Share | Verdict | Evidence |
|---|---|---|---|
| Execute (20662), Heroic Strike VIII (11567), Whirlwind | 24% / 24% / 6% | **Unchanged everywhere** | Identical to Era at every rank, ≤30 included. |
| Bloodthirst (23894), 31-point | 12% | **Tuned** | Forever rework (flat 30/37/43/48 plus % AP). 70291 raised the AP part 35→45% on all four ranks: Blizzard is editing it in the beta. |
| Battle Shout (raid) | – | **Tuned** | 0.60× Era at every rank. |
| Flurry, Bloodthrill (deep talents) | – | **Tuned / design** | Bloodthrill doubled in 70009; Flurry is a Forever 5–25% curve, untouched since. |

No sensitivity needed; Fury doesn't rest on stale-looking data.

## Rogue

| Spell | Share | Verdict | Evidence |
|---|---|---|---|
| Backstab, Sinister Strike, Eviscerate VIII (11300) | 8–30% | **Unchanged everywhere** | Identical to Era at every rank. |
| Rupture | – | **Tuned** | 0.58–0.62× at every rank. |
| Mutilate | 21% of the Mutilate build | **Tuned** | Forever ranks 23 / 33 / 48 / 67 at L30 / 40 / 50 / 60 (a smooth ×1.42 a rank), 75% weapon at every rank. |
| Thousand Cuts, Venom, Adrenaline Rush (31-point) | – | **Unclear (untested)** | No beta build changed them. |

No sensitivity needed; the rogue's place at the bottom doesn't come from stale 31+ data.

## Druid: Feral cat

| Spell | Share | Verdict | Evidence | DPS sensitivity |
|---|---|---|---|---|
| Shred V (9830) | 39% | **Tuned** | 155% weapon (Era 225%) at every rank, flat part unchanged. | none |
| Rip VI (9896) | 25% | **Unclear (31+ diverges from ≤30)** | Per-combo-point 4.4 / 7.2 at L20 / 28 (+10% / +3% on Era) but 8.5 / 12.7 / 18.2 / 25.5 at 31+ (−6 to −9% on Era). Fractional, Forever-authored values, but the 31+ ranks move against the ≤30 change. | back to Era at 31+: **+1.6%**; following the ≤30 trend (×1.13): **+2.1%** (648→662) |
| Ferocious Bite V | 2% | **Unchanged everywhere** | Average equals Era's. | none |
| Tiger's Fury | – | **Tuned (redesign)** | One rank (+15, 30 s cooldown); the 36 / 48 / 60 ranks are deleted. | – |

## Shaman: Elemental

| Spell | Share | Verdict | Evidence | DPS sensitivity |
|---|---|---|---|---|
| **Lightning Bolt X (15208) + overload** | 65% | **Stale-shaped** | 0.61× Era at L26, L32 and L38, then 0.54 / 0.48 / 0.44 at L44 / 50 / 56. Rank increments 21 / 29 / 36, then 15 / 16 / 18 (Era 45–75). 70291 raised the top playable rank 12% at its cap and touched no 31+ rank. At 60 it does 78 base a cast-second against 84–158 for the other nukes, with which it was even at level 26–30. | 70291's +12% applied to rank 10: **+1.8%**; held at the 0.61 ratio (×1.38): **+5.8%** (690→730) |
| Lava Burst III (1238300), 31-point | 14% | **Tuned relative to Lightning Bolt** | 70009 raised ranks 1–2 (113→164, 179→196) to about 10% above the Lightning Bolt of their level. Rank 3 (220) is 12% above Lightning Bolt X. It inherits Lightning Bolt's flat curve. | with Lightning Bolt ×1.38 and Lava Burst ×1.38 together: **+7.2%** (690→740) |
| Flame Shock VI (29228) | 14% | **Tuned** | 0.54–0.57× at 31+ against 0.50 at L28. | none |
| Chain Lightning (not in the preset) | – | **Stale-shaped, plus a data error** | 88 / 100 / 112 / 123 (+12 a rank, Era +90–120); rank 3 (2860) coefficient 0.517 against 0.571 on the other ranks (the sim copies it). | – |
| Searing / Magma Totem | 5% | **Tuned** | Base unchanged; coefficient cut at every rank. | none |

## Priest: Smite

| Spell | Share | Verdict | Evidence | DPS sensitivity |
|---|---|---|---|---|
| **Smite VIII (10934)** | 39% | **Stale-shaped (mild)** | 0.51× Era at L30, then 0.42 / 0.41 / 0.43. Growth L30→L54 ×2.1 against Era ×2.5. | held at 0.51 (×1.18): **+1.3%** |
| Smite IV (984), downranked | 21% before 2026-10-13 | **Tuned**, but no longer cast | A ≤30 rank, retuned in 70291. Under Forever's downranking penalty it keeps no spell power at 60, so the preset falls back to Smite VII (a 31+ rank, stale-shaped like VIII). | see the 18:25 check |
| Holy Fire VIII (15261) | 18% | **Tuned** | 0.54× at L30, 0.46–0.52× at 31+. | none |
| Shadow Word: Pain VIII (10894) | 16% | **Tuned** | Subtractive family (−15 a tick, plateau from L50); coefficient ×1.2 at every rank. | none |

Shadow priest (next pass in detail): Mind Flay, Mind Blast and Shadow Word: Pain are all in the subtractive family, which sits at about 0.9× Era at 60. That's tuned in shape and, if anything, generous.

## Pass 2: the other specs

### Druid: Balance

| Spell (rank) | Share | Verdict | Evidence | DPS sensitivity |
|---|---|---|---|---|
| Starfire VI (9876) | 56% Balance | **Tuned** | 0.72× Era at L26, 0.64–0.71 at 31+; growth ×3.5 against Era's ×3.6. | none |
| **Wrath VIII (9912)** | 14% Balance, 59% Balance Wrath | **Stale-shaped (mild)** | 70009 multiplied every rank ×1.5; 70291 then retuned only ranks 1–5. 0.42× Era at L30, then 0.36 / 0.34 / 0.36. Growth L30→L54 ×2.0 against Era's ×2.3. | held at 0.42 (×1.17): Balance Wrath **+1.6%** (760→772), Balance +0.4% |
| Moonfire X (9835) | 14% | **Tuned** | 0.70 at L28, 0.63–0.66 at 31+. | none |
| Insect Swarm V (24977) | 16% | **Tuned (slight flattening)** | 0.65 at L30, then 0.69 / 0.57 / 0.57. | <0.5% |
| Moonkin Form, Wild Growth (31-point) | – | **Unclear (untested)** | Never changed in a beta build. | – |

### Priest: Shadow

| Spell (rank) | Share | Verdict | Evidence |
|---|---|---|---|
| Mind Flay VI (18807) | 44% | **Tuned** | Subtractive family: −12 a tick from L44, 0.81× at L28 rising to 0.92× at L60; coefficient ×1.11 at every rank. |
| Mind Blast IX (10947) | 20% | **Tuned** | −32 a hit from L46; 0.88× at L28 rising to 0.94×. |
| Shadow Word: Pain VIII (10894) | 20% | **Tuned** | see pass 1. |
| Shadow Word: Death (1309636), new | 9% | **Unclear (31+ only, untested)** | One rank at L56: 448 + 42.9%, 15 sec cooldown. In line with Mind Blast IX (485 + 42.9%, 8 sec). |
| Devouring Plague VI (19280) | 5% | **Tuned** | 0.85× at L28 rising to 0.94×; coefficient ×1.59 and a 60 sec cooldown (was 180) at every rank. |

### Mage: Frostfire, Fire, Arcane, Frost

| Spell (rank) | Share | Verdict | Evidence | DPS sensitivity |
|---|---|---|---|---|
| Frostfire Bolt III (1237313), SoD rune | 63% Frostfire | **Tuned, relative to SoD** | 100 / 182 / 292 at L40 / 50 / 60, 3 s cast, coefficient 0.814 (SoD rune: 1.0). At 60 it's 0.76× Frostbolt X's base per cast; SoD's rune was 0.66× Frostbolt XI. | none |
| **Frostbolt X (10181)** | 98% Frost | **Unclear (mild lag)** | Subtractive family: 0.79× at L26, 0.74× at L32 rising to 0.89× at L56. 70291 raised the L20/L26 ranks (+8% at L26's cap) but no 31+ rank. The 31+ ratios already sit above the playable one, so it's unclear whether the raise should carry over. | if it does (×1.08): Frost **+2.8%** (744→764, past Enhancement + Water Shield and Survival) |
| Fireball XI (10151), Pyroblast, Scorch, Fire Blast | 46% / 10% / 6% / 16% Fire | **Tuned** | Fireball 0.66× at L30, 0.63–0.71× at 31+; Pyroblast 0.66→0.73; Scorch 0.65→0.70; Fire Blast subtractive (−32 from L46). | none |
| Arcane Blast V (1239700), SoD rune | 52% Arcane | **Tuned at 60** | 54 / 130 / 169 / 274 / 394 at L20–60. The L40 rank grows little (+39), but at 60 it's 158 base per cast-second against Fireball XI's 155. | none |
| Arcane Missiles VII (10212) | 46% Arcane | **Tuned** | Bolts are Era minus 21 at 31+ (subtractive); coefficient 0.24 → 0.286 at every rank. | none |

### Hunter: Lone Wolf, Marksmanship, Beast Mastery, Survival

| Spell (rank) | Share | Verdict | Evidence | DPS sensitivity |
|---|---|---|---|---|
| Arcane Shot VIII (14287) | 13–18% | **Tuned** | 1.10× Era at L28, 1.13–1.19× at 31+. | none |
| Aimed Shot VI (20904) | 13–16% | **Tuned** | 0.27–0.29× Era at every rank, with a 2 s cast (Era 3 s). | none |
| Serpent Sting VIII (13555) | 9–10% | **Tuned** | 0.79 at L26, 0.81–0.85 at 31+. | none |
| Multi-Shot | – | **Tuned (redesign)** | One rank (L18); the 30–60 ranks are deleted. | – |
| Sniper Shot III (1310786), 31-point | 6–8% | **Unclear (untested)** | 160 / 225 / 295, smooth; never changed in a beta build. | – |
| **Raptor Strike VIII (14266)** | 11% Survival | **Stale-shaped** | Flat bonus 30 at L24 (0.88× Era), then 35 / 40 / 55 / 70 (0.50–0.70×). The playable ratio isn't kept at 31+. | held at 0.88 (×1.76 flat): Survival **+1.2%** (761→770) |
| Mongoose Bite IV | 8% Survival | **Tuned** | 0.49–0.50× at L30 and 31+. | none |
| Pet Claw / Bite | 3–6% | **Unchanged everywhere** | Era values at every rank. | none |

### Shaman: Enhancement

| Spell (rank) | Share | Verdict | Evidence | DPS sensitivity |
|---|---|---|---|---|
| Windfury Weapon IV (16362) | 17% | **Unchanged everywhere** | Era values at every rank (the sim's 333 AP is deliberate). | none |
| Stormstrike (17364) | 10% | **Tuned (redesign)** | Forever version: 8 sec cooldown, flat mana cost; one rank. | – |
| Earth Shock VII (10414) | 10% | **Stale-shaped (mild)** | 0.66× at L24, then 0.57 / 0.55 / 0.57. Growth L24→L60 ×3.7 against Era's ×4.3. | held at 0.66 (×1.16): **+0.8%** |
| Flame Shock, Lightning Bolt X | 8% / 7% | see Elemental (pass 1) | | |

### Warrior: Arms (and Fury 2H)

Mortal Strike IV (21553), Overpower IV (11585), Slam IV (11605), Execute and Heroic Strike are **unchanged everywhere**: Era values at every rank. Forever only added a 20-level Slam rank and Slam's 18 sec category cooldown. No stale-data risk.

### Shared raid buffs and debuffs

All **tuned**, with the same Forever factor at every rank: Battle Shout 0.60×, Blessing of Might 0.72×, Blessing of Wisdom 1.2×, Mark / Gift of the Wild 1.35× (plus the 70291 aura change on every rank), Power Word / Prayer of Fortitude 1.30×, Demoralizing Shout / Roar 1.40–1.48×. Arcane Intellect, Divine Spirit, Faerie Fire and Sunder Armor (reworked in 70009 at every rank) are Era-identical or changed at every rank. Curse of Recklessness is a new linear curve, 175 / 285 / 395 / 505 armor (+110 a rank; Era 140 / 290 / 465 / 640), the same step at every rank. The totem values sit in triggered aura spells; the shared branch caps their ranks already (Forever mechanics auditor's area).

## Rank inversion scan (orchestrator request, 16:10 UTC; ruling f8ae7cdc: presets use the strongest learned rank)

Scanned client 70338 for every rank chain on the nine classes' skill lines, talent-taught ranks included and the Ahn'Qiraj book ranks excluded. Each effect's value is taken at level 60, with per-level growth capped at the rank's max level. Then the same with 600 spell power added through each rank's coefficient. Then the triggered spells (seal procs, aura triggers), the totem aura spells (Strength of Earth, Grace of Air, Windfury Totem, Mana Spring, Flametongue Totem, Stoneskin) and the weapon imbue and poison enchants (SpellItemEnchantment: Windfury, Flametongue, Rockbiter, Frostbrand, Instant, Deadly, Wound and Crippling Poison).

**The only inversions in anything a preset or the shared raid uses are the three the hunter found:**

| Spell | Rank the sim used | Stronger lower rank |
|---|---|---|
| Aspect of the Hawk | VI (14322, L58): 55 ranged AP | V (14321, L48): 90 |
| Trueshot Aura | V (20906, L60): +50 | IV (20905, L50): +75 |
| Hunter's Mark | IV (14325, L58): 71 | III (14324, L40): 98 |

Everything else rises monotonically with rank: every damage and heal chain (also with 600 spell power), every raid buff and debuff (Battle Shout, Might, Wisdom, Kings, Mark / Gift of the Wild, Fortitude, Arcane Intellect, Divine Spirit, Sunder, Faerie Fire, Curse of Recklessness, Demoralizing Shout / Roar), the totem auras (Strength of Earth 7/14/25/42, Grace of Air 49/77, Windfury Totem 95/179/246, Mana Spring 4/6/8/10) and the imbues and poisons (Instant 15→88, Deadly 6→18, Rockbiter, Frostbrand). Forever's Windfury and Flametongue Weapon use one level-scaled spell for every rank, so they can't invert.

Hits that are not real inversions: the Lightning Bolt / Chain Lightning / Lava Burst overload copies and the Penance / Holy Shock / Flash of Light / Holy Light helper spells (same-level companions), the SoD rune entries Lacerate 414647 and Tiger's Fury 417045 (not learnable), and the scripted dummy values on Blizzard, Hurricane, Volley and Rain of Fire.

One data-error case worth knowing under the new ruling: Seal of Fury rank 4 (20419, L34) procs 74 with a 0.900 coefficient, against 35 and 0.000 for rank 7 (20423). That's the client error noted in pass 1. Retribution doesn't use Seal of Fury, but if a build ever did, "strongest learned rank" would pick the broken rank 4.

## Pass 3 (16:15 UTC)

- No Forever build newer than 1.60.1.70338 on wago.tools. The class branch commits since 15:35 are talent-verification and rotation fixes (warlock pre-pull Demonic Sacrifice, paladin Holy Power / Consecrated Ground, rogue Murder, hunter Expose Prey and Efficiency, shaman and warrior talent checks). None touches rank data, so the sensitivities above stand as percentages.
- Succubus Lash of Pain (11780, 2–2.5% of the warlock builds): **tuned**, 0.48–0.51× Era at every rank (16 / 22 / 30 / 36 / 43 / 50), coefficient unchanged.
- Venom (1310703: +30% via two damage-percent auras, plus a 10 flat modifier), Thousand Cuts (1310723: −3 flat modifier) and Lacerating Strikes (1310533: 40, with a 1-a-tick periodic): one rank each, never changed in any beta build. **Untested**, with no Era or SoD counterpart to compare against.
- The orchestrator's rule (89729b6) excludes client data errors such as Seal of Fury rank 4 from the strongest-rank ruling.

## Check 2026-10-13 12:20 UTC: the standards 1b gear round

- No Forever build newer than 1.60.1.70338. wago.tools was down (Cloudflare 503) from 22:15 Oct 12 to about 12:00 Oct 13 and is back.
- Every class rebuilt its gear on the Forever budget (standards 1a / 1b / 5a) and added synthetic trinkets. No commit touched rank data; the warlock's curses.go change only names the banes.
- Re-run on the new heads (20k iterations, same local patches):
  - Bane of Doom at a 2.0 coefficient: Demonology 1109.8 → 1062.9 (−4.2%), Destruction 985.4 → 944.0 (−4.2%), Affliction 977.4 → 934.4 (−4.4%). With Fury now 1020 and Retribution twist 967 on the new gear, Destruction and Affliction would still drop below Retribution; Demonology stays first.
  - Elemental, Lightning Bolt X: base 673.9; ×1.12 → 687.1 (+1.9%), ×1.38 → 715.5 (+6.2%), ×1.38 with Lava Burst too → 725.4 (+7.6%).
- The other sensitivities are small and scale the same way with spell power and attack power; not re-run.

## Check 2026-10-13 18:25 UTC: Forever's downranking penalty

- Still no Forever build after 1.60.1.70338.
- The priest (6ee9d02) and shaman (651dcb6) branches now apply Forever's downranking penalty (Blizzard's 2026-10-01 notes; foreverchanges.pro's formula: a rank loses 5% of its coefficient for each level past learn level + 17). That changes what "the rank the preset casts" means for this audit:
  - Smite: the low-mana fallback moved from Smite IV (L22, now no spell power at 60) to Smite VII (L46), a stale-shaped 31+ rank like Smite VIII (0.41–0.43× Era). Smite VII and VIII are now most of the damage. Holding every 31+ Smite rank at the L30 ratio (×1.18): **+1.8%** (634.0 → 645.5, with Shadow's consumes, since mksettings has no Smite spec; only the percentage counts). It was +1.3% before.
  - Elemental: the fallback is now Lightning Bolt VIII (L44), also stale-shaped. The 20k-iteration numbers on the new head match the 12:20 check (673.9 base; +1.9% / +6.2% / +7.6%), because the fallback only plays below 20% mana.
- The druid branch now treats Tiger's Fury as removed under Forever; this audit had it as a single rank (5217 at L24) in client 70338. That's the druid worker's call, and it isn't a 31+ rank question.
- The hunter applied the strongest-rank ruling (Aspect of the Hawk V, Trueshot Aura IV, Hunter's Mark III). Since the user's ruling on the report assumptions (2026-10-15), the sim gives Trueshot Aura V 100 and Hunter's Mark IV 140 (fixed onto the curve), so the top ranks are cast, and Raptor Strike VIII adds 96, halfway from the client's 70 to the 0.88 of Era (123).

## Check 2026-10-14 00:30 UTC: client 1.60.1.70374

- wago.tools lists a new Forever build, 1.60.1.70374. Its SpellEffect, SpellPower, SpellMisc, SpellLevels, SpellCooldowns, SpellName, SkillLineAbility, TraitDefinition, TraitDefinitionEffectPoints and CurvePoint tables are byte-identical to 70338's, so nothing in this report changes.
- Branch commits since 18:25 Oct 13 are UI, merges and a Retribution Shock rotation knob (Light's Vigil mana floor); none touch rank data.

## Check 2026-10-16 06:50 UTC: client 1.60.1.70425, a real balance build

wago.tools lists 1.60.1.70425. Unlike 70374 it changes a lot: 575 SpellEffect rows (82 on class spells), SpellLevels, SpellPower, SpellCooldowns, the talent trees (TraitNode, TraitNodeEntry, node groups) and 2,900 CurvePoint rows. **For the first time in the beta, Blizzard edited 31+ ranks and deep-tier talents on purpose:**

| Change (70374 → 70425) | Ranks / tier | Presets affected |
|---|---|---|
| Arcane Blast: base ×0.6 and coefficient 0.714 → 0.429 at every rank (rank 5: 394 → 236) | all, 31+ included | Arcane mage (Arcane Blast ~52% of its damage): a large nerf |
| Lava Burst: ranks moved 10 levels earlier (30 / 40 / 50), the talent moved from the 30-point gate to the 20-point gate, costs cut; per-level values kept, so the L60 rank (220) is gone and the best at 60 is rank 3 (196 + 1.1 a level to 58 ≈ 205, −7%) | 31-point talent → 21-point; Elemental Mastery (16166) is the new 30-point capstone | Elemental: Lava Burst −7%, and the talent string needs checking against the new tree |
| Chain Lightning: costs −35% at every rank (485 → 315 at rank 4) | all, 31+ included | not cast by the presets |
| Judgement of Righteousness coefficient 0.5 → 0.429, Judgement of Fury 0.45 → 0.429, every rank | all | Retribution (JoR 6–11% of its damage) |
| Devouring Plague cooldown 60 → 24 sec at every rank; Vampiric Embrace 60 → 30 sec | all | Shadow, Smite |
| Holy Fire rank 5 (L42) periodic 13 → 9: an out-of-line 31+ entry fixed (4/5/6/7/**13**/10/13/15) | one 31+ rank | none (rank 8 is cast) |
| Shadow Mastery (25-point gate) 1–5% → 2–10% | deep tier | warlocks (Affliction / Demonology builds that take it) |
| Champion of the Light (25-point) 20/40/60% → 10/20/30% of Intellect | deep tier | Retribution builds that take it |
| Mage tree: Critical Mass 3 → 5 ranks, Fire Power 5 → 3 ranks (1/2/3%), Winter's Chill 5% at every rank, Missile Barrage, Frost Channeling, Improved Fire Blast reworked, Arcane Wrath added | several tiers | all mage presets |
| Priest tree: Shadow Affinity and Improved Fade removed, Mental Strength and Improved Mind Flay reworked, Dark Inspiration added | several tiers | Shadow, Smite |
| Warlock tree: Malediction removed, Grim Reach added; Curse of Recklessness reworked (armor via a new aura) | – | warlock talent strings; the raid's CoR |
| Ice Lance 300 → 200 at every rank; Shred rank 1 moved to L20 (24 → 22); Moonkin Aura effect changed; Righteous Fury 60 → 90 | – | Frost / Frostfire, Balance |

Also changed in 70425, on proc / aura spells (found by the paladin, the druid and a second diff):

| Change | Presets affected |
|---|---|
| Seal of Righteousness proc spells 25713–25742: coefficient 0.1 → 0 | Retribution (the paladin left the sim as is, pending a ruling) |
| Consecration tick spells 1280345–1280353: coefficient 0.095 → 0.042 | Retribution |
| Shaman Fire Nova damage 408423–408428: base about ×0.6, coefficient 0.214 → 0.143 | Enhancement |
| Mana Tide restore 197 → 133 and 290 → 197 | Elemental + Mana Tide |
| Strength of Earth, Grace of Air, Windfury Totem, Flametongue Totem, Trueshot Aura, Leader of the Pack: party area aura → raid area aura (Effect 35 → 65); Leader of the Pack is now 3% physical crit, Moonkin Aura 3% spell crit | shared raid layout |
| Howling Idol's Shifting Power cooldown cut −1000 → −500 ms (1291059) | Feral (about −7.5 DPS) |
| Wind Blessed (Skyborne) 1 → 3; Earth Fury (1259719) changed | racials |

The class ports so far: paladin PR #54 (twist and sub-Holy about −102 DPS, Shock about −64; Shock now leads up to 180 sec) and druid PR #52.

**What it means for this audit:**
- The 31+ ranks I flagged as stale-shaped were **not** touched: Lightning Bolt, Smite, Wrath, Raptor Strike, Earth Shock, Rip, Holy Strike, Consecration and Bane of Doom's 4.0 coefficient are byte-identical. So are Shadow Bolt and the other tuned curves.
- Blizzard now edits 31+ data, though: design changes at every rank (Arcane Blast, Judgements, Devouring Plague, Chain Lightning costs), a fix for one out-of-line 31+ rank (Holy Fire V), two 25-point talents retuned, and a 31-point talent pulled down into the playable tiers (Lava Burst). The verdicts above still describe 70425's 31+ data, but "never changed in a beta build" no longer holds for the deep tier as a whole.
- Every class's sims are on 70338 values and need the 70425 deltas (the workers' and the Forever mechanics auditor's job). Talent strings may no longer fit the changed trees (Elemental, mage, priest, warlock). This report's sensitivities are percentages on 70338 data and will be re-run once the branches move to 70425.

## Check 2026-10-16 08:55 UTC: sensitivities on master with the 70425 ports

Master d511c09 has the 70425 racials (#50, #53, #56), raid hits (#55), the damage-taken Rage change (#57) and the warlock (#59), priest (#58, #60), warrior (#62), druid (#61, #64), hunter (#65) and rogue (#66) ports. Mage, paladin (#54) and shaman are still pending. No client build is newer than 70425.

Re-run on the master request files (tools/dps_report/req, 20k iterations, same local patches):

| Preset | Base | 31+ rescale | Change |
|---|---|---|---|
| Demonology, Bane of Doom coefficient 2.0 | 1099.2 | 1053.1 | −4.2% |
| Destruction, Bane coefficient 2.0 | 980.7 | 941.4 | −4.0% |
| Affliction, Bane coefficient 2.0 | 998.0 | 956.4 | −4.2% |
| Destruction Fire, Bane coefficient 2.0 | 953.3 | 917.7 | −3.7% |
| Smite, 31+ Smite ranks ×1.18 | 710.6 | 725.4 | +2.1% |
| Feral, Rip 31+ ×1.13 | 713.7 | 726.3 | +1.8% |

**What the Bane of Doom risk now means for the ranking:** dual-wield Fury on Orc is 1055.1 after #62. At coefficient 2.0, Demonology (1053.1) would fall just below Fury. Once the paladin's #54 lands (about −100 for twist and sub-Holy), Destruction and Affliction stay above Retribution even at coefficient 2.0. So the swing now decides first place (Demonology vs Fury), not warlocks vs Retribution.

Still to re-run when their ports merge: Elemental (Lightning Bolt / Lava Burst on the new Lava Burst ranks), Retribution (Holy Strike / Consecration with Consecration's new 0.042 ticks), Balance Wrath.

## Check 2026-10-16 11:00 UTC: paladin and mage ports on master

Master 4477bcb now has the paladin (#54, #81) and mage (#89, #90) ports, the off-hand Rage ruling (#67, #75), demons inheriting spell crit (#85), Smite without Holy Fire (#83) and the hunter (#76, #82) and rogue (#73, #80) follow-ups. The shaman 70425 port hasn't merged yet. No client build is newer than 70425.

Re-run on the master request files (20k iterations, same local patches; I re-simmed every preset's base for the ranking):

| Preset | Base | 31+ rescale | Change |
|---|---|---|---|
| Demonology, Bane coefficient 2.0 | 1118.3 | 1074.1 | −4.0% |
| Affliction, Bane coefficient 2.0 | 1013.3 | 971.6 | −4.1% |
| Destruction, Bane coefficient 2.0 | 995.9 | 956.6 | −3.9% |
| Destruction Fire, Bane coefficient 2.0 | 970.2 | 936.1 | −3.5% |
| Ret twist (level_60), Holy Strike ×0.52 / Consecration ×0.82 / both | 852.7 | 849.1 / 844.8 / 841.2 | −0.4% / −0.9% / −1.3% |
| Ret Shock, both | 865.2 | 852.9 | −1.4% |
| Ret sub-Holy, both | 846.5 | 835.2 | −1.3% |
| Frost, Frostbolt 31+ ×1.08 | 827.3 | 839.4 | +1.5% |
| Frostfire, Frostbolt 31+ ×1.08 | 1010.8 | 1010.8 | 0 (casts no Frostbolt) |
| Smite, 31+ Smite ×1.18 | 723.7 | 743.3 | +2.7% |
| Feral, Rip 31+ ×1.13 | 752.7 | 765.7 | +1.7% |

Consecration's share went up with 70425's 0.042 tick coefficient on master, so its ratio risk is now the larger half of Ret's (−0.9% of −1.3%). Frost's swing is smaller than on 70338 (+2.8%) because Frost now leans less on Frostbolt.

**Ranking effects (master bases):**
- **First place:** Demonology 1118.3, Lone Wolf hunter 1075.7. Fury dropped to 970.8 after the off-hand Rage ruling. At coefficient 2.0, Demonology (1074.1) lands 1.6 below Lone Wolf, so the Bane of Doom risk still decides first place, now against Lone Wolf instead of Fury.
- **Affliction and Destruction:** at coefficient 2.0, Affliction (971.6) falls below Arms on dragonkin (1015.4), Frostfire (1010.8), Fire (993.2) and Arms (976.4). Destruction (956.6) also falls below Fury (970.8) and Fury 2H (969.3).
- **Frost:** +1.5% lifts it from 827.3 to 839.4, past Balance (828.2). It stays below Ret twist (852.7).
- **Smite:** at 743.3 it passes both Combat rogue presets on humanoids (740.5, 734.1), and stays below Feral (752.7).
- **Feral:** at 765.7 it passes Balance Wrath duty-FF (755.8), and stays below Balance Wrath (784.4).
- **Ret:** −1.3% doesn't change the order: Ret twist at 841.2 is still above Balance (828.2), and Enhancement (852.9) was already above it.

Balance (druid ports #61 and #64 are already on master), with Wrath 31+ ×1.17: Balance Wrath goes 784.4 → 798.1 (+1.7%), still below Enhancement + Water Shield (804.2). Balance goes 828.2 → 831.8 (+0.4%). Neither changes the order.

## Check 2026-10-16 15:15 UTC: the shaman port on master

The shaman 70425 port (#91) merged at 10:59 UTC. The 13:05 check missed it because it only searched from 11:00. Re-run on master 0bf4de5 with the shaman patches added, 20k iterations, and every preset's base re-simmed for the ranking. No client build is newer than 70425.

| Preset | Base | 31+ rescale | Change |
|---|---|---|---|
| Elemental, Lightning Bolt X ×1.12 (70291's ≤30 raise) | 724.9 | 740.4 | +2.1% |
| Elemental, Lightning Bolt X ×1.38 (held at the 0.61 ratio) | 724.9 | 774.0 | +6.8% |
| Elemental, Lightning Bolt ×1.38 and Lava Burst ×1.38 | 724.9 | 785.7 | +8.4% |
| Enhancement, Earth Shock 31+ ×1.16 | 818.5 | 827.0 | +1.0% |

The 70425 port lifted Elemental from 684.9 to 724.9, with Lava Burst now a 20-point talent. It sits 52nd of 64 presets, just above Smite (723.7).

**Ranking effects:**
- **Lightning Bolt ×1.12:** Elemental at 740.4 passes nothing new and stays below Combat Daggers on humanoids (744.6).
- **Lightning Bolt ×1.38:** at 774.0 it passes both Combat rogue presets on humanoids, Feral (752.7), Enhancement + Mana Tide (752.6) and Balance Wrath duty-FF (755.8). It stays below Balance Wrath (784.4).
- **Lightning Bolt and Lava Burst ×1.38:** at 785.7 it passes Balance Wrath too and lands 1.8 below Enhancement + Water Shield (787.5). That's still the largest upward swing in the audit, about 13 places at most.
- **Enhancement:** with Earth Shock ×1.16 it goes 818.5 → 827.0, 1.0 above Frost (826.0) and 1.2 below Balance (828.2). That margin is within noise, so call it no change in order.

The other current bases are unchanged since the 11:00 check, apart from small moves from the new ports: Fury 989.0, Frostfire 1018.1, Arcane 922.9. Demonology (1118.3) and Lone Wolf (1075.7) are still first and second, so the Bane of Doom finding at 11:00 stands: coefficient 2.0 gives 1074.1, 1.6 below Lone Wolf.

Every rank-tuning sensitivity has now been re-run on master with the 70425 ports.


## Check 2026-10-16 15:30 UTC: re-baseline on master ddb095a (no level-based partial resists)

This section replaces the ranking effects in the 11:00 and 15:15 sections. Since then master has merged #102 (no level-based partial resist under Forever, a user ruling), #98 (Arcane), #99 (Survival), #100 (Ret) and #101 (Fury). I re-simmed all 66 presets' bases and every sensitivity on ddb095a: same local patches, 20k iterations, 180 s.

| Preset | Base | 31+ rescale | Change | Place (of 66): before → after |
|---|---|---|---|---|
| Demonology, Bane coefficient 2.0 | 1176.3 | 1129.3 | −4.0% | 1 → 2, behind only its own duty-curse preset (1164.7) |
| Affliction, Bane coefficient 2.0 | 1066.0 | 1021.7 | −4.2% | 7 → 12 |
| Destruction, Bane coefficient 2.0 | 1047.4 | 1005.6 | −4.0% | 9 → 15 |
| Destruction Fire, Bane coefficient 2.0 | 1018.5 | 982.2 | −3.6% | 13 → 18 |
| Ret twist, Holy Strike ×0.52 and Consecration ×0.82 | 889.2 | 876.9 | −1.4% | 36 → 39 |
| Ret Shock, same | 901.7 | 888.7 | −1.4% | 34 → 36 |
| Ret sub-Holy, same | 883.1 | 871.1 | −1.4% | 37 → 41 |
| Frost, Frostbolt 31+ ×1.08 | 843.3 | 855.6 | +1.5% | 42 → 42 |
| Enhancement, Earth Shock 31+ ×1.16 | 841.4 | 850.5 | +1.1% | 43 → 42 |
| Balance, Wrath 31+ ×1.17 | 872.8 | 876.6 | +0.4% | 41 → 40 |
| Balance Wrath-only, Wrath 31+ ×1.17 | 825.9 | 840.5 | +1.8% | 46 → 44 |
| Elemental, Lightning Bolt X ×1.12 | 771.1 | 787.6 | +2.1% | 50 → 49 |
| Elemental, Lightning Bolt X ×1.38 | 771.1 | 823.3 | +6.8% | 50 → 47 |
| Elemental, Lightning Bolt and Lava Burst ×1.38 | 771.1 | 835.8 | +8.4% | 50 → 45 |
| Smite, 31+ ×1.18 | 765.4 | 786.3 | +2.7% | 51 → 49 |
| Feral, Rip 31+ ×1.13 | 752.9 | 765.9 | +1.7% | 53 → 51 |

**What changed:** #102 lifted the warlocks more than the hunters. Demonology is now 82 above Lone Wolf (1094.1), so Bane of Doom at coefficient 2.0 (1129.3) **no longer decides first place**. It still costs every Shadow Bolt warlock about 4%:
- Affliction falls below Fire, Marksmanship and Destruction.
- Destruction falls below Destruction Fire and Arms on dragonkin.
- Destruction Fire falls below Fury.

**Other specs:**
- Ret's −1.4% puts sub-Holy just below Balance.
- Elemental's biggest case (+8.4%) passes Water Shield Enhancement, Balance Wrath-only and Imp GW Enhancement. It stays below Balance with Faerie Fire duty.
- Smite's +2.7% passes Elemental and Mana Tide Enhancement.
- Balance (+0.4%), Frost and Enhancement move within noise.

## Cross-class summary

**Specs whose standing rests on likely-stale or untested 31+ data:**

| Spec | Rank now (of 37) | Main 31+ risk | Plausible swing | Rank after swing |
|---|---|---|---|---|
| Demonology | 1 (1087) | Bane of Doom coefficient 4.0 | −4.1% (coefficient 2.0) | still 1 (1043) |
| Destruction / Affliction | 2–4 (985 / 976) | Bane of Doom coefficient 4.0 | −4.1% | **fall below Ret twist 962 and sub-Holy 957, level with Fury 944** (945 / 935) |
| Retribution (all three) | 5, 6, 8 | Holy Strike flat jump, Consecration ratio | −1.2% | unchanged order |
| Fury | 7 | none (Era-identical abilities) | 0 | – |
| Elemental | 29 (690) | Lightning Bolt 31+ flattening | +1.8% to +7.2% | up to 26–27 (730–740), passing Smite and the best rogue, still below Frost mage (744) |
| Smite | 27 (717 documented) | Smite VIII flattening | +1.3% | unchanged |
| Feral | 34 (648) | Rip 31+ nerf against the ≤30 buff | +1.6% to +2.1% | 32–33, past Mutilate and Combat Swords |
| Rogue | 28–37 | none | 0 | – |
| Frost mage | 25 (744) | Frostbolt X missing 70291's +8% | +2.8% | 22–23 (764) |
| Balance Wrath-only | 23 (760) | Wrath VIII flattening | +1.6% | 22 (772) |
| Survival | 22 (760) | Raptor Strike flat part | +1.2% | 21 (770) |
| Enhancement | 20 (800) | Earth Shock VII flattening | +0.8% | unchanged |
| Frostfire, Fire, Arcane, Shadow, Balance, hunters, Arms | 9–19 | none beyond untested 31-point spells | <0.5% | unchanged |

**Bottom line after pass 2:** nothing in pass 2 moves a spec more than 3%. The pass 1 conclusions stand.

**Bottom line for pass 1:** the 31+ ranks the presets use are mostly Forever-edited, not stale Era copies. The ranking's top and bottom don't come from stale data, with one exception that matters: Bane of Doom's 4.0 coefficient is untested and worth about 4% to every Shadow Bolt warlock. It's the difference between warlocks holding places 1–4 and Retribution taking 2–3. Lightning Bolt is the clearest stale-shaped curve: up to +6–7% for Elemental, which would still leave it in the bottom third.

## Not covered yet

- SoD's server-side baseline-rank values (not in the client).
- Pet ability ranks other than the hunter cat's (Succubus Lash of Pain), and Forever's new rotational talents' proc values (Venom, Thousand Cuts, Lacerating Strikes). No beta build changed any of these.
- Later passes re-check the branches and the beta builds on wago.tools for new data (a new build that touches 31+ ranks would settle several "unclear" verdicts).
