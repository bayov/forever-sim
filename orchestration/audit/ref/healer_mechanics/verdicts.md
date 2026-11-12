# Healer mechanics audit: verdicts (as of 2026-11-12 UTC, master cb0d928fd6, client 1.60.1.70888)

Tier: 1 = Forever client (wago.tools) or Blizzard Forever notes, 2 = Season of Discovery, 3 = Era / Classic. Never vmangos.
Pass 10 (2026-11-12, cb0d928fd6): tooltip-only commit; no heal, healing talent or engine change. Riptide overlay text (445 over 15 s, 12 s, 147 mana) and Mana Tide 13 s/88 match the sim; the 147 is finding 11's gap. Finding 13 resolved (3335001a5). Pass 9 (2026-11-11, 707ae92cc5): Nature's Swiftness now takes Wild Growth (9628e85b6, finding 12 resolved); the dot CleanUp change (f20d5938a) drops only ticks due after a HoT's aura ended, so a last tick at expiry still lands; Blizzard's Nov 10 healer lines (Revelation from Vampiric Embrace, Holy Nova unlearn, Lesser Spirit enchants) need no sim change. Pass 8 (2026-11-10, da272241e5): client 70888 touches among healer rows only the Swiftmend 18562 tooltip ("consumed" dropped, now agrees with the Swiftmend row) and Natural Instinct's crit-damage mask (word 3 +0x88, Pounce; healing part 1361117 unchanged); Nov 10 hotfix gives Nature's Swiftness Wild Growth (reopens that row, finding 12). Pass 6 (2026-11-08, 9ba00c2dc4): downranking scope narrowed to heals only (0193721bc; every heal config carries ProcMaskSpellHealing, so no heal left it) and rolls round to whole points (2a60326a1); no healer row changed. Pass 5 (2026-11-07, 12aa2796fc): client 70862 touches only Bloom (1328622 heal now targets the ally, was the caster) among healer spells; Riptide rows unchanged. Pass 3 (2026-11-05, a079415316): no healer spell or engine change; heal threat row added. Pass 2 (2026-11-04, 0facd9428): findings 1-9 fixed on master and re-checked; 70790 touches only Riptide among
healer spells. Pass 1 checked every heal rank and healing talent of the four healers and the engine. A row marked match holds on
70755; re-check it when a newer build or Blizzard post touches the spell. Open rulings stay in each
`docs/healing_<class>.md` "Calls for the user"; open findings in `../../findings/healer_mechanics.md`.

## Settled (don't reopen without new tier 1 evidence)

- The user's healing rulings in `docs/healing_sim.md` (score, assignments, bound ranks, scares reported only).
- Healing on top of spell power is its own stat, priced 0.45 in gear (standards 1b, 35).
- Riptide under Forever: no instant heal, 12 s. Client 70790 now agrees (effect 0 HEAL -> DUMMY base 0, category
  recovery 12000, mana 150/180/230); the sim's notes-scaled mana 147/180/231 is off by 3 and 1 (healer_mechanics-20261104-11).
- Swiftmend: news 24301515 and dev notes post 9 (2026-11-04 edit) say it no longer consumes the HoT; the 18562
  tooltip still said "consumed" until client 70888 dropped it (Blizzard Nov 6 notes: tooltip fix). Fixed on master bc8e5473b (Forever keeps the HoT).

## Restoration shaman

| Spell / talent | Client (70755) | Sim | Verdict | Tier |
|---|---|---|---|---|
| Healing Wave 1-10: levels, base, growth, variance | 331 1/5 39 +0.7 var .2564 … 25357 60/65 1615 +5.5 var .1326 | `HealingWaveRanks` identical, every rank | match | 1 |
| Healing Wave: mana, cast | 25/45/80/155/200/265/340/440/560/620; 1.5/2/2.5 then 3 s | same | match | 1 |
| Healing Wave: coefficient | .429/.571/.714 then .857 (EffectBonusCoefficient) | same; × core.DownrankFactor | match | 1 |
| Lesser Healing Wave 1-6 | 8004 20/153/1.7/.1379/105 … 10468 60/820/4.2/.1091/380; 1.5 s; .429 | same | match | 1 |
| Chain Heal 1-3 | 1064 40/251/2.5/.1395/260, 10622 46/335/2.8/.1379/315, 10623 54/490/3.3/.1322/405; 2.5 s; .714 | same | match | 1 |
| Chain Heal jumps / falloff | EffectChainTargets 3, EffectChainAmplitude 0.5; party-only when cast on party member (tooltip) | 3 targets, ×0.5 a jump, most hurt, party-only rule | match | 1 |
| GCD (HW, LHW, CH, Riptide) | 1500 (SpellCooldowns StartRecoveryTime) | GCDDefault | match | 1 |
| Riptide ranks | 408521 40 (max 48), 1239242 50 (58), 1239243 60 (68) | 3 ranks, same levels | match | 1 |
| Riptide instant heal | 70790: effect 0 is DUMMY (3), base 0 (70755 had HEAL 480/612/841) | none under Forever | match (client 70790) | 1 |
| Riptide ticks | 89 / 115 / 161 (+3.7 a level past 60 on r3), every 3 s, 15 s, coef .1 | 5 ticks, same values, coef .1, downranked | match | 1 |
| Riptide tick crit | Attributes_8 0x200 (periodic can crit) on all ranks | Forever ruleset: dot ticks roll crit (core/ruleset.go) | match (untested for Riptide) | 1 |
| Riptide mana / cooldown | 70790: 150/180/230; category recovery 12000 (70755: 245/300/385, 6 s) | 147/180/231 (0.6 x 70755); 12 s shared timer; T1 5P -1 s | cooldown match; mana off by 3/0/1 (finding 11) | 1 |
| Riptide Chain Heal bonus | eff 2 dummy aura 4, 25%; tooltip "casts directly on that target" | +25% on the whole chain when first target has our Riptide | open call 2 (not a finding) | 1 |
| Healing Stream Totem 1-5 | 5394…10463: lvl 20/30/40/50/60, mana 40-80, heal 5/6/7/9/11 per 2 s (5672…10461), coef .022, 300 s, GCD 1 s | same | match | 1 |
| Healing Stream crit | heal 10461 has Attributes_8 0x200 (periodic can crit) | flat 5% (beta log ruling), not spell crit | ruling; client says "can crit", chance unknown | 1 + logs |
| Mana Spring Totem 1-4 | 26/36/46/56, 40/60/80/100 mana, 4/6/8/10 per 2 s | same | match | 1 |
| Mana Tide Totem 1-4 | 16190 1, 17354 38, 17359 48, 1325256 58; 10/20/30/60 mana; 88/133/197/290 per 3 s; 13 s; 300 s category; GCD 1 s | same; 4 ticks | match | 1 |
| Water Shield | 408510: 3 charges, 3.5 s proc ICD, 600 s, 1.5 s GCD, ProcTypeMask 0x362a8 (no periodic); 408511: 2% max mana | same | match | 1 |
| Nature's Swiftness | 16188: -100% cast, mask HW/LHW/CH/LB/CL/Ghost Wolf/Healing Rain, 1 charge, 180 s, CD starts when used up | Nature spells with a cast time; 3 min from consumption | match | 1 |
| Improved Healing Wave | -100 ms/pt, aura 107 misc 10, mask HW | flat -0.1 s/pt on HW | match | 1 |
| Tidal Focus | -1%/pt cost (misc 14), mask HW/LHW/CH/Riptide; +1% melee+spell hit/pt | cost factor on all shaman healing spells (= those 4); hit | match | 1 |
| Tidal Mastery | +1%/pt crit (misc 7), mask HW/LHW/CH/Riptide (+LS, Rolling Thunder) | +crit on the 4 heals | match | 1 |
| Purification | +2%/pt eff 0 (HW/LHW/CH/Riptide) and eff 1 misc 22 (Riptide periodic) | ×(1+.02 pts) on the 4 heals incl. ticks; not Healing Stream | match | 1 |
| Healing Way | flat +8/17/25% (aura 108 misc 0, mask HW) | flat, multiplies with Purification | match | 1 |
| Healing Focus | 23/47/70% no pushback, mask HW/LHW/CH/Healing Rain | PushbackReduction on HW/LHW/CH | match | 1 |
| Restorative Totems | +5%/pt Mana Spring (misc 8), +10%/pt Healing Stream (misc 22 periodic) | same; whole heal under Forever | match | 1 |
| Totemic Focus | -5%/pt every totem's cost | same | match | 1 |
| Mindfulness / Natural Grace | 17/33/50% casting regen; -5/10/15% threat all magic schools | same | match | 1 |
| Ancestral Healing → Ancestral Fortitude | 16176 proc 0x15550 (no periodic), 16177 **aura 101** (MOD_RESISTANCE_PCT, armor) 8/17/25% | PseudoStats.ArmorMultiplier on all armor (4648ff419) | match | 1 |
| Forever Tier 1 Restoration set | 1300960 2P +10 Spirit; 1301059 3P totem radius; 1301093 4P +26 healing/+9 dmg; 1301713 5P Riptide -1 s | synthetic set on head and chest, 2P live; 4P 17 healing + 9 SP; 5P riptideCooldownCut (c938028ee) | match | 1 |
| Relics: Tidal Totem, Totem of Urgency | 1291076 HW -5% cost; 1306448 LHW +4% crit | modeled | match | 1 |
| Tidebringer's Claw (280605, ilvl 40) | 1309539 +2% crit HW/LHW | not modeled | not modeled (leveling only) | 1 |
| Totem of Life / Flowing Water | Naxx (ilvl 78/83) | not modeled | not modeled (standards 8a) | 1 |
| Earthfury / Ten Storms bonuses | items not in Forever ItemSparse | "Nothing to do" stubs | n/a (sets not in game) | 1 |
| Earth Shield, Healing Rain | not shaman-trainable in Forever data seen | absent | n/a | 1 |

## Holy paladin

| Spell / talent | Client (70755, tier 1) | Sim | Verdict | Tier |
|---|---|---|---|---|
| Holy Light 1-9 (635…25292) | base/rpl/variance/learn/max exactly as `healing.go:39-49`; coef 0.714; 2.5 s; mana 35…660; no CD | same | match | 1 |
| Flash of Light 1-6 (19750…19943) | base/rpl/var as table; coef 0.429; 1.5 s; mana 35…140 | same | match | 1 |
| Holy Shock heal 1-4 (1311605, 25914, 25913, 25903) | 114/156/230/320, rpl 0, var 0.0755-0.0789, coef 0.429 | same | match | 1 |
| Holy Shock cast 1-4 (1311606, 20473, 20929, 20930) | 160/225/275/325 mana; category 892, category recovery 10 s (shared); GCD 1.5 | same, one shared timer | match | 1 |
| Holy Shock damage (1311604, 25912, 25911, 25902) | 134/182/258/348 ±var, coef 0.429 | 129-139 … 334-362, 0.429 | match | 1 |
| Light's Vigil cast 1-3 (1310911/1311590/1311595) | 730/1000/1340 mana, 1.5 s, category recovery 6 s (category 2576); eff#1 dummy bp 75 | same; no refund | match (refund = open Call 4) | 1 |
| Light's Vigil mark (1310909/1311593/1311597) | 30 s, 1 charge | same | match | 1 |
| Light's Vigil heal (1310912/1311591/1311596) | 324+1.2/lvl (≤49), 487+1.5 (≤59), 704+1.8 (≤69), var 0.0562, coef 0.143, target+party | same | match | 1 |
| Lay on Hands 633/2800/10310 | learned 10/30/50; eff 67 max health; 0/250/550 mana to target; no power row; 20 min category CD | same, top rank only | match | 1 |
| Downranking (heals, Oath bonus) | engine rule (user ruling, foreverchanges) | `core.DownrankFactor` on coef and Oath bonus | match / Oath = open Call 2 | ruling |
| Healing Light (20237) | aura 108 misc 0, curve 4/8/12; mask 0x6000,0x10000 → HL, FoL, HS heal (+Oath/BoL bonus spells), not Vigil | ×1.04/rank on HL, FoL, HS heal (Oath included) | match | 1 |
| Spiritual Focus (20205) | curve 14…70; aura 108 misc 9; mask → HL, FoL, Light's Vigil | PushbackReduction 0.14/rank on all three | match | 1 |
| Divine Favor (20216) | +100% crit (aura 107 misc 7), mask → HL, FoL, Holy Shock (cast, dmg, heal); 1 charge; 4% base mana; CD 120 s; GCD 0; **own class mask 0** | same; no Benediction (d64804b9f) | match | 1 |
| Illumination (20210) | "$m1% chance… $m3% of base cost": eff0 curve 20…100, eff2 50 | 20%/rank, 50% of the rank's base cost | match | 1 |
| Illumination on Vigil | tooltip lists Light's Vigil | once per detonation, Vigil base cost | open Call 3 | 1 |
| Infusion of Light (426065 → 437063) | curve −500/−1000 ms; buff 15 s, 1 charge, mask HL; procs on helpful and harmful | 0.5 s/rank, 15 s, 1 charge, FoL/HS (dmg+heal) crits | match | 1 |
| Holy Power (5923) | aura 50 misc 2 (Holy), curve 10…50 | CritDamageBonus 0.3/rank (225% at 5/5) | reading = open Call 1 | 1 |
| Oath of Light (1328548) | 8% base mana, GCD 1.5, 5 min; rpl 6.6667/1.9167/3.8333 → 400/115/230 at 60; **class mask 0** | same; no Benediction (d64804b9f) | match | 1 |
| Boundless Light (1328550) | curve 15/30, dummy | +15%/rank HS heal crit when not on Oath target | match | 1 |
| Divine Intellect (20257) | aura 137, Int (bitmask 8), curve 2…10 | ×(1+0.02/rank) Int | match | 1 |
| Reverence (1310899) | aura 134, curve 15/30 | +0.15/rank casting regen | match | 1 |
| Conviction (20117) | aura 290, curve 2/4/6 | +2%/rank spell and melee crit | match | 1 |
| Benediction (20101) | aura 108 misc 14, curve −2…−10; mask covers Holy Shock (cast and heal) and seals etc.; **not** Divine Favor or Oath (mask 0) | Holy Shock, not DF/Oath; `TestBenediction` | match | 1 |
| Libram of Economy (272436 → 1291089) | −5% cost, mask HL only | ×0.95 HL cost; Illumination stays on 660 base | match | 1 |
| Libram of Infusion (279248 → 1306429) | +6% crit, mask HS cast + HS heal | +6% crit on both | match | 1 |
| Libram of Holy Alacrity (228175 → 449980 → 449982) | 449982: aura 107 misc 10 −200 ms, mask HL, 10 s, 1 charge | doc answered from the client (call 9) | doc match; trigger server side | 1 |
| Justice Armor 2pc (1300948) | +10 Spirit | +10 Spirit | match | 1 |
| Justice Armor 4pc (1301080) | +26 healing (aura 135), +9 damage (aura 13) | +26 HP, +9 SD | match | 1 |
| Justice Armor 5pc (1301694) | aura 107 misc 11 −1000 ms, mask Holy Shock | Holy Shock CD −1 s | match | 1 |
| Holy Power on HS damage | aura 50 is healing crit only | no damage crit bonus | match | 1 |

## Restoration druid

| Spell / talent | Client (70755) | Sim | Verdict | Tier |
|---|---|---|---|---|
| Healing Touch r1-11 (5185...25297): base, growth, variance | bp 44...2332, RealPointsPerLevel 0.8...7.3, BaseLevel..MaxLevel (r11 60..65), Variance 0.318...0.166 (full width) | `HealingTouchRanks` healing.go:89-102, same numbers; `At()` = bp + rppl*(min(lvl,max)-base), ±var/2 | match | 1 |
| Healing Touch mana / cast / coef / GCD | 25...840; 1.5/2/2.5/3/3.5 s (r5+ 3.5); coef 0.429/0.571/0.714/0.857/1.0; GCD 1.5 s | same | match | 1 |
| Regrowth r1-9 (8936...9858) direct | bp 87...1021, rppl 1.8...6.8, var 0.154...0.109, coef 0.286, 2 s, 70...525 mana | same (healing.go:107-117, consts :119-124) | match | 1 |
| Regrowth HoT | eff 6 aura 8, bp 13...142, rppl 0, period 3 s, dur 21 s (7 ticks), coef 0.071/tick | same; tick has no growth | match | 1 |
| Rejuvenation r1-11 (774...25299) | aura 8 bp 8...194, 3 s, 12 s (4 ticks), coef 0.2/tick, 25...360 mana, instant | same (healing.go:128-141) | match | 1 |
| Tranquility r1-4 (740...9863) | eff 35 aura 8 bp 87/129/196/285, rppl 0.6/0.7/0.9/1.0, MaxLevel 36/46/56/66, 2 s x5 (10 s), coef 0.067, 375...925 mana, category CD 300 s | same (healing.go:151-157) | match | 1 |
| Swiftmend 18562 cost/CD/GCD | PowerCostPct 20, CD 15 s, GCD 1.5 s | same | match | 1 |
| Swiftmend amount | tooltip: "full duration of the periodic effect" | all ticks (Splendor's too) at the HoT's snapshot, Swiftmend's own mods | match (amount reading is doc call 2) | 1 |
| Swiftmend keeps the HoT | news 24301515 ("Redesigned to not consume the Rejuvenation or Regrowth effect"); dev notes post 9, 11-04 edit | healing.go:698 cancels only outside Forever (bc8e5473b); tests in sim/raid_healing_druid*_test.go | match | 1 |
| Swiftmend which HoT | not in data | least time left | not in data (doc call 3) | - |
| Wild Growth r1-3 | bp 40/60/97, rppl 1/1.3/1.6, 3 s cast, 6 s CD, 7x1 s, coef 0.033, 550/755/1050, 5 targets | same | match (shape = doc call 1) | 1 |
| Bloom 1328621/1328622 | 100% Spirit (50% dispel), heal has no mask; 70862: heal ImplicitTarget 1 (caster) -> 21 (target ally) | same; heals the HoT's target | match (70862) | 1 |
| HoT / Tranquility tick crits | Attributes_8 0x200 on Rejuv, Regrowth, Tranq, WG | ticks roll crit (ruleset.go canCrit) | match | 1 |
| Nature's Swiftness 17116 | aura 108 misc 10 -100%; mask: HT, Regrowth, Wrath, Roots, Hibernate, Rebirth, Soothe, Nourish, Starsurge; CD 180 s, GCD 0. Nov 10 hotfix (Blizzard's Nov 10 hotfix post, tier 1; push 113885): also Wild Growth | HT, Regrowth, Wrath, Wild Growth (9628e85b6); off GCD; 3 min | match | 1 |
| Innervate 29166 | aura 110 bp 400 (power regen %), aura 134 bp 100 (regen while casting); 20 s; CD 360 s; 5% base mana | SpiritRegenMultiplier += 4, ForceFullSpiritRegen; 20 s; 6 min; 5% (buffs.go:1455, innervate.go) | match | 1 |
| Gift of Nature 17104 | #0 misc 0 (all) 2%/pt: HT, Regrowth, Swiftmend, Tranq; #1 misc 22 (periodic): Regrowth, Rejuv, WG | DamageMultiplier on HT/RG/Tranq/SM; periodic on Rejuv/WG; Regrowth HoT once | match (Tranq = doc call 5) | 1 |
| Improved Rejuvenation 17111 | misc 22 5%/pt (15%), Rejuvenation | Rejuv periodic x1.15 | match | 1 |
| Genesis 1223081 | #0 misc 22 2%/pt: Rejuv, Regrowth, WG (+ dots); #1 misc 0: Swiftmend | same; not Tranq | match | 1 |
| Tranquil Spirit 24968 | misc 14 -2%/pt: HT, Tranquility | cost factor on HT, Tranq | match | 1 |
| Moonglow 16845 | misc 14 8/17/25%: damage spells only (no heal in mask) | `damageSpells()` only | match | 1 |
| Improved Regrowth 17074 | aura 107 misc 7 +10%/pt crit, Regrowth | BonusCritRating on Regrowth (ticks too) | match (ticks = doc call 4) | 1 |
| Naturalist 17069 | misc 10 -0.1 s/pt: HT, WG; +2%/pt damage | same; damage multiplier not healing | match | 1 |
| Nature's Splendor 1223083 | +3 s Rejuv (+1 tick), +6 s Regrowth (+2) | same | match | 1 |
| Improved Tranquility 17123 | 50%/pt threat, 30%/pt CD | same | match | 1 |
| Nature's Focus 17063 | 14%/pt pushback; HT, Regrowth, Tranq, ... | HT, Regrowth (hardcasts; channel pushback not modeled) | match for what's modeled | 1 |
| Reflection 17106 | aura 134, 17/33/50% | SpiritRegenRateCasting += 0.17/0.33/0.5 | match | 1 |
| Subtlety 17118 | aura 10 misc 72 (Nature+Arcane) -10%/pt | Nature/Arcane spells x(1-0.1/pt) | match | 1 |
| Living Spirit / Heart of the Wild | 5%/pt Spirit; 2%/pt Int | same | match | 1 |
| Natural Instinct 1223242 | 12/25% Int as spell healing | restoration.go:61-63, Restoration agent only | match | 1 |
| Nature's Grace / Omen of Clarity on heals | 16886 GCD mask has HT/RG/Rejuv/Tranq/WG (no Swiftmend); 16870 Clearcasting mask all heals | same | match (OoC rate = doc call 6) | 1 |
| Forever T1 Resto set 1300943/1301006/1301077/1301250 | +10 Spi; Barkskin -10 s; +26 heal/+9 dmg; Swiftmend -3 s | same (HP 17 + SP 9) | match | 1 |
| Idol of Swiftness 1306488 | Swiftmend -3 s | same | match | 1 |
| Stormrage 5/8, Cenarion 8 | (Era sets; not in the 70755 ItemSparse export) | "Nothing to do" | not modeled (likely moot) | 3 |
| Downranking | ranks' SpellLevel/MaxLevel | engine rule, heals only (user ruling 2026-11-08) | ruling, not audited | - |
| Doc tables, lower ranks | formula above | doc now 1095-1303, 800-894, 633-710, Tranq from the client | match (fixed) | 1 |

## Holy and Discipline priest

| spell / talent | client 70755 | sim | verdict | tier |
|---|---|---|---|---|
| Lesser Heal r1-3 (2050/2052/2053) | bp 51/94/163, rppl .3/.7/1.2, var .196/.179/.151, coef .429/.571/.714, lv 1-3/4-9/10-15, cast 1.5/2.0/2.5, mana 30/45/75 | identical (healing.go:57-62) | match | 1 |
| Heal r1-4 (2054..6064) | bp 270/375/523/651, rppl 2/2.9/4/4.5, coef .857, cast 3.0, mana 155-305 | identical | match | 1 |
| Greater Heal r1-5 (2060..25314) | bp 844..1960, rppl 5.1..8.1, var .119..0.110, coef .857, cast 3.0, mana 370..710, r5 lv 60-65 | identical | match | 1 |
| Flash Heal r1-7 (2061..10917) | bp 194..825, rppl 1.9..4.2, coef .429, cast 1.5, mana 125..380 | identical | match | 1 |
| Prayer of Healing r1-5 (596..25316) | bp 178..649, var ~.054-.064, coef .286, cast 3.0, mana 410..1070, target 132/31 (target's party) | identical, partyOf(target) | match | 1 |
| Holy Nova r1-6 heal (23455..27805) | heal bp 51..311, rppl .4..0.9, var, coef .107, target 20 (caster party, 10 yd); cast 15237.. mana 185..750, instant | identical; party of priest; damage part not modeled (doc call 9) | match | 1 |
| Renew r1-10 (139..25315) | aura 8 bp 9..166, coef .2/tick, period 3000, dur 15000 (5 ticks), mana 30..410, Attr8 0x200 (ticks crit) | identical; ticks crit live (ruleset canCrit) | match | 1 |
| Power Word: Shield r1-10 (17..10901) | aura 69 misc 127, bp 44..928, rppl .8..4.3, coef .1, dur 30 s, category CD 4 s, mana 45..500 | identical; shared 4 s timer | match | 1 |
| Weakened Soul 6788 | 15 s | 15 s, blocks PW:S | match | 1 |
| PW:S replace by >= rank | no client field (notes 2026-11-02) | ExtraCastCondition | match (notes) | 1 (notes) |
| Prayer of Mending r1-3 (401859/1240826/1240827) | dummy bp 172/298/413, coef .429, charges eff#1 = 5, aura 401877 30 s, category CD 10 s, mana 210/305/390; heal 401863 | identical; 5 heals, 30 s per hop, 10 s CD, triggers on damage or non-periodic heal taken | match | 1 |
| Penance heal bolts (402289/1240723/1240724/1316991) | bp 126/200/268/425, rppl .87/.98/1.27/0, coef .19, cast mana 150/220/270/385, CD 12 s | identical (penance.go:31-33) | match | 1 |
| Downranking | engine (sim/core/downrank.go), user ruling 2026-11-08 (heals only) | heals set Rank/RequiredLevel; PoM and Penance heals DownrankLevel | not re-audited (engine) | - |
| Spiritual Healing 14898 | 3/7/10%; eff0 mask: LH,Heal,GH,FH,PoH,HN heal 23455,PoM cast 401859,Penance heal bolts; eff1 (DOT) Renew; not PW:S | all healCodes x mult, not PW:S | match (see A1 for PoM) | 1 |
| Improved Renew 14908 | 5/10/15%, DOT mod, Renew | x(1+.05r) Renew | match | 1 |
| Holy Specialization 14889 | +1%/r crit; mask: all heals incl. Renew, HN heal, PoM cast, Penance bolts, Smite, HF; not PW:S | same set | match (see A1) | 1 |
| Divine Fury 18530 | -100 ms/r cast; Heal, GH, Smite, HF | Heal, GH (+dmg spells) | match | 1 |
| Improved Healing 14912 | -5/10/15% cost; LH, Heal, GH, Penance cast, PoM cast | same | match | 1 |
| Spiritual Guidance 14901 | a174 1/3/5/6/8% Spirit->SP; a175 5..25% Spirit->healing | same | match | 1 |
| Meditation 14521 | a134 17/33/50 | SpiritRegenRateCasting += .17/.33/.5 (859bdef1a: adds, no longer overwrites) | match | 1 |
| Mental Agility 14520 | -3/7/10% cost; heals: Renew, PW:S, HN cast, PoM cast (+PI, SW:P, DP, Smite, HF...) | same | match | 1 |
| Mental Strength 18551 | 2..10% Int (a137 stat mask 8 = Int) | 2%/r Int | match | 1 |
| Improved PW:S 14748 | +7/14/20% PointsIndex0 on PW:S | x1.07/1.14/1.20 on (base + .1 SP) | match (order of SP vs mod not in client data) | 1 |
| Soul Warding 402000 | PW:S CD -4000, cost -15% | CD 0, cost x0.85 | match | 1 |
| Twin Disciplines 1225132 | 1%/r; eff0 (direct) HN cast+heal, PoM heal 401863, Penance bolts (not FH 0x800, not PW:S); eff1 (DOT) Renew | HN, PoM, Penance, Renew | match | 1 |
| Twilight Focus 14913 | 23/47/70 pushback, LH/Heal/GH/FH/PoH/Penance | same | match | 1 |
| Inner Focus 14751 | eff0 -100% cost (all incl. PW:S, Renew); eff1 +25% crit non-periodic (heals, HN heal, PoM heal, Penance bolts; not Renew); 3 min; 1 charge | free any priest spell with cost; +25 crit on non-PureDot non-channel spells; 3 min | match | 1 |
| Power Infusion 10060 | a136 + a79 20%, 15 s, 3 min, 20% base mana | HealingDealtMultiplier x1.2 + magic dmg x1.2, 15 s, 3 min, 20% | match (self only) | 1 |
| Renewed Hope 425280 | +2%/r crit, -1 s/r Weakened Soul; FH, LH, Heal, GH, Penance (tooltip; aura 4 no mask) | same | match | 1 |
| Divine Aegis 431622/431624 | 5/10/15% of crit heal, cap 20*level, 12 s, absorb all schools | same; stacking to cap = doc call 1 | match | 1 |
| Inspiration 14892/14893 | non-periodic crit heal, +8/17/25% armor (a101 misc 1), 15 s | same | match | 1 |
| Litany of Light 1317006 | 5/10% of base cost when previous heal differs | same; PW:S excluded, same-name ranks = same (doc call 2) | match | 1 |
| Holy Binding 1328569/1328570 | 50% of FH on other, no threat | same | match | 1 |
| Blessed Recovery 27811/27813 | 8/17/25% over 6 s (2 s period), melee/ranged crit or >30% hit (proc 0x2A8) | same, carry-over | match | 1 |
| Spirit Tap 15270, Martyrdom, Spirit of Redemption, Holy Reach | non-healing / range | not modeled in healer | n/a | 1 |
| Vestments of Conviction 5pc 1301703 | CD -1000 ms on Penance (0x800000) AND PoM (0x20) | both (dda14db65) | match | 1 |
| Vestments of Conviction 2/4pc | +10 Spi; +26 heal / +9 dmg | same | match | 1 |
| Era sets (Prophecy 3/8pc, Transcendence 8pc, Oracle, Faith 2/4pc) | heal bonuses | "Nothing to do" | not modeled (no such items in Forever ItemSparse 70755; not on presets) | 3 |
| docs/healing_priest.md audit tables | say client 70753 | every value equals 70755 | match (label only) | 1 |

## Engine (sim/core)

| Check | Rule (source, tier) | Sim (file:line) | Verdict |
|---|---|---|---|
| Death at 0 health, dead player stops, heals and casts can't reach the dead | game rule; H1 design | health.go:215-232, player_death.go:10-54, spell.go:560-565, apl_values_resources.go:147 | OK |
| Max health falls (Last Stand ends): current health clamps to the new max | game rule (current health never above max) | unit.go processDynamicBonus clamps on a negative Health bonus (0aed93989) | match (no engine test) |
| Max health rises (buff): current health stays | game rule (a buff adds max, not current) | health.go: nothing on stat change | OK |
| Most Hurt / Lowest Health: living, hurt, in the assignment, seen health less others' heals in flight | H1 design (user ruling) | heal_targets.go:36-57, heal_comm.go:48-50 | OK |
| Tank N; dead tank reads 0 missing, heal on it refused | design | environment.go:442-446, apl_values_resources.go:147, spell.go:563 | OK |
| Chain Heal jumps to the most hurt on real health, party-limited when target in our party | client 10623 ImplicitTarget 45, 3 chain targets; 1.12 tooltip (tier 3, tooltip) | heal_targets.go:82-100 (MostHurtAmong) | OK (server picks at landing, so real health is right) |
| Heals in flight (HealComm) and reaction time | design (healing_sim.md, What healers see) | heal_comm.go:20-44, health.go:52-72 | OK as designed |
| Cancelling: stop unpaid, cooldown reset, GCD held to cast start + 1.5 s, no five-second rule | game: mana paid at cast end (tier 3) | heal_comm.go:53-80, gcd.go:109-131, cast.go:347-360 (cost at completion) | OK |
| Overheal booking: part above missing; target without health bar all effective | design | spell_result.go:700-704, health.go:88-96 | OK |
| Heal threat: 0.5 per point of effective healing x spell and healer threat mods; whole heal on a unit with no health bar or a tank in a tank-score replay | Classic (tier 3); replay case is tank-sim design (tank_mechanics-20261104-16) | spell_result.go:137-152 (464deda7b) | OK (replay branch never hit in healing fights: tankScore nil) |
| Absorbs: all schools | client 10901 aura 69 misc 127; Divine Aegis 431624 misc 127 (tier 1) | absorb.go:120-137 (no school check) | OK |
| Absorbs: after armor, resists, damage-taken mods, the outcome; before health, procs, Rage | game (absorbed damage is not damage taken; tier 3) | spell_result.go:315-323 | OK |
| Absorbs: order among shields (oldest first), after Ice Barrier-like dynamic absorbs | game: aura order; not sourced | absorb.go:124, spell_result.go:317-322 | OK enough (only matters with two absorbs on one unit) |
| Absorb booking (absorbed = effective healing; rest overheal; fight end books nothing) | Open decisions 2 default; priest Call 6 | absorb.go:102-117, 129 | as ruled default |
| Absorb `Add` gives the whole shield to the newest spell | n/a | absorb.go:93 | harmless (only Divine Aegis adds, same spell) |
| HoT snapshot: healing power and caster healing multipliers at cast; target's healing taken (Mortal Strike) at each tick | Classic (tier 3) | spell_result.go:681-698, druid/healing.go:276-279, spell_result.go:645-656 | OK |
| HoT refresh: recast replaces the HoT with a new snapshot and ticks | Classic (tier 3) | dot.go:139-145 (`Apply`) | OK |
| HoT ticks crit under Forever | client SpellMisc Attributes_8 0x200 on Renew, Rejuvenation, Regrowth, Tranquility, Wild Growth, Riptide 1239242, Healing Stream 5672-10461 (tier 1); verdicts "Periodic crits" | ruleset.go:15-28, spell_outcome.go:31-49 | OK. Bloom (1328622) has no 0x200 but is a direct heal in the sim |
| Crit heal size 1.5x (talents aside) | Classic (tier 3); client has no general change | spell_outcome.go:1063-1068 (DefenseTypeMagic: 1 + (1.5*at.CritMultiplier-1)*CritDamageBonus) | OK |
| Heal crit chance source | spell crit with all bonuses, school ones too (tier 3: Vestments of Prophecy 5pc "chance to crit with Holy spells") | HealingCritChance adds GetSchoolBonusCritChance (0aed93989), as HoT ticks | match (no engine test) |
| Five second rule, Spirit regen under Forever, continuous settle | verdicts (verified) | mana.go:195-270, sim.go:861-866 | not redone; settle on every advance, so mid-interval rate changes are right |
| Innervate: Spirit regen x5, full regen while casting, 20 s, 6 min | client 29166: aura 110 bp 400, aura 134 bp 100, dur 20000, cd 360000 (tier 1) | buffs.go:1396-1397, 1465-1474 | OK |
| Mana Spring, Mana Tide, Healing Stream, Tranquility reach: party | client 5677/10494, 16191, 5672/10461, 9863: effect 35 (area aura, party), radius index 10 (30 yd) / 9 (tier 1) | shaman water_totems.go, totem_receivers.go; external Mana Tide buffs.go:1528-1556 pays only the member it's on | OK |
| External Mana Tide (raid sim placeholder) 290 a tick, 4 ticks | rank 4 1325256 (290); 16191 period 3000 | buffs.go:1538 | OK for rank 4 |
| Raid damage: physical through armor (attacker 63: armor/(armor+400+85*63)), magic through Forever resists, DoT ticks a tenth of resistance, no armor on physical ticks | shared verdicts; Classic armor formula (tier 3) | raid_damage.go:40-57, spell_resistances.go:43-57, 108-114, 187-209 | OK |
| Raid damage: who and how many | encounter files: `share` = the logs' share of the non-tanks a cast hits, rate fitted on share x non-tanks (tools/encounters/raid_from_logs.py:66-79) | raid_damage.go:188-200: share x N < 1 hits one player that often, nobody otherwise (b2c16f671) | match (no engine test) |
| Raid damage DoT ticks can crit (dot.canCrit under Forever) | n/a | raid_damage.go:107 | harmless: bosses have no SpellCrit stat, so 0% |
| Heal roll in whole points (low to high, rounded to nearest; average unchanged for whole ends) | Nov 5-8 live logs, damage (Searing Totem 9/10/11 at 1/4, 1/2, 1/4); heals assumed alike | sim.go Roll (2a60326a1); every heal's base roll goes through sim.Roll | OK (heals not seen in a log yet) |
| Spell pushback on a healer hit while casting: 1.0, 0.8 ... 0.2 s | source is cmangos code (cast.go:97-107), not tier 1-3 | cast.go:108-200 | source gap, not a finding (verdicts only checks "not past its start") |

