# Level 31+ rank tuning audit

Read-only audit by the rank tuning auditor. It covers the spell ranks and deep talents the Level 60 presets rely on that can't be learned under the beta's level 30 cap, and asks whether each looks tuned for Forever or stale. Pass 1, 2026-10-12, covering Demonology, Destruction (and Affliction), Retribution, Fury, Rogue, Feral, Elemental and Smite. The other specs come next pass.

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
| Smite IV (984), downranked | 21% | **Tuned** | A ≤30 rank, retuned in 70291. | – |
| Holy Fire VIII (15261) | 18% | **Tuned** | 0.54× at L30, 0.46–0.52× at 31+. | none |
| Shadow Word: Pain VIII (10894) | 16% | **Tuned** | Subtractive family (−15 a tick, plateau from L50); coefficient ×1.2 at every rank. | none |

Shadow priest (next pass in detail): Mind Flay, Mind Blast and Shadow Word: Pain are all in the subtractive family, which sits at about 0.9× Era at 60. That's tuned in shape and, if anything, generous.

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

**Bottom line for pass 1:** the 31+ ranks the presets use are mostly Forever-edited, not stale Era copies. The ranking's top and bottom don't come from stale data, with one exception that matters: Bane of Doom's 4.0 coefficient is untested and worth about 4% to every Shadow Bolt warlock. It's the difference between warlocks holding places 1–4 and Retribution taking 2–3. Lightning Bolt is the clearest stale-shaped curve: up to +6–7% for Elemental, which would still leave it in the bottom third.

## Not covered yet

- Next pass: Balance (Wrath and Starfire), Shadow in detail (Shadow Word: Death, Devouring Plague), the mage specs (Frostfire Bolt, Frostbolt's +8% top-rank retune), hunters (Multi-Shot, Arcane Shot, Aimed Shot, Serpent Sting, Lone Wolf / Sniper Shot), Enhancement (Stormstrike, Windfury), Arms (Mortal Strike), and pet abilities (Lash of Pain, the cat's Claw / Bite).
- Raid buffs and debuffs, for the cross-check. Seen so far: Battle Shout 0.60×, Blessing of Might 0.72×, Blessing of Wisdom 1.2×, Mark / Gift of the Wild 1.35×, Fortitude 1.30×, all at every rank (tuned). Curse of Recklessness 1.25× down to 0.79× by rank (to check).
- SoD server-side baseline-rank values (not in the client).
