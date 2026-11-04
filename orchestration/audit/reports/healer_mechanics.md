# Report: healer_mechanics, pass 1 (2026-11-04 UTC, master a1b1163d2, client 1.60.1.70755)

Forced first pass after the H3 healers merged (Holy and Discipline priest with the absorb pipeline, Holy paladin,
Restoration druid, Restoration shaman with the November 2 Riptide). Result: **findings** (2 high, 1 medium, 7 low).

## Get current

- Client: 1.60.1.70755 is the newest (the mechanics pass of 06:58Z probed it; its tables are byte-identical to 70753's,
  which the class docs cite). Every heal rank and talent was read from the 70755 tables
  (SpellEffect, SpellMisc, SpellLevels, SpellPower, SpellCastTimes, SpellCooldowns, SpellDuration, SpellClassOptions,
  TraitDefinitionEffectPoints / CurvePoint for one-spell multi-rank talents).
- Blizzard: dev notes up to post 9 (edited 2026-11-04T02:12Z, its Swiftmend line) and the druid / priest articles
  (news 24301515 / 24301514). The Swiftmend line is finding 2.
- Ledger since 36cd929a2: only the paladin hotfix 113542 (Vengeance stacks, Holy two-hand spec), no healing row.
- What changed since the seeded `last.master` 36cd929a2: only the audit loop's own state files (a1b1163); the pass
  therefore read every healer and engine file in full.

## Per spec (full tables: `ref/healer_mechanics/verdicts.md`)

| Spec | Checked | Mismatches against tier 1 |
|---|---|---|
| Restoration shaman | Healing Wave 1-10, Lesser Healing Wave 1-6, Chain Heal 1-3 (3 targets, 50% a jump), Riptide 1-3, Healing Stream, Mana Spring, Mana Tide, Water Shield, 11 healing talents and their masks, doc tables | Ancestral Fortitude's armor (3), no Tier 1 set (6). Riptide's notes-only values stay open as call 1 |
| Holy paladin | Holy Light 1-9, Flash of Light 1-6, Holy Shock heal 1-4 (category 892), Light's Vigil 1-3, Lay on Hands, Healing Light, Spiritual Focus, Illumination, Infusion of Light, Boundless Light, Divine Intellect, Reverence, Conviction, Divine Favor, Oath of Light, Justice Armor 2/4/5, libram masks | Benediction on Divine Favor and the Oath (4), doc's Libram of Holy Alacrity line (9) |
| Restoration druid | Healing Touch, Regrowth (split), Rejuvenation, Tranquility, Wild Growth, Bloom, Swiftmend, Nature's Swiftness, Innervate, 17 talents and their masks, T1 set, Idol of Swiftness | Swiftmend consumes the HoT (2), doc rank numbers and a stale comment (9) |
| Holy / Discipline priest | All 50 ranks: Lesser Heal, Heal, Greater Heal, Flash Heal, Prayer of Healing, Holy Nova, Renew, Power Word: Shield, Prayer of Mending, Penance; Weakened Soul; 21 talents and masks incl. Divine Aegis 5/10/15% capped at 20 a level | Vestments of Conviction 5-piece misses Prayer of Mending (5) |
| Engine | death, targets, heals in flight, cancelling, overheal, absorbs (all schools, after armor and resists, oldest first, no rage), HoT snapshot and refresh, 1.5x crits, HoT tick crits (Attributes_8 0x200), five-second rule, Innervate, party reach of Mana Spring / Mana Tide / Healing Stream / Tranquility, raid damage armor and resists | raid damage share floor (1), school crit on direct heals (7), health above max (8) |

## Open questions (not findings; for the user or the class docs)

- Priest: Prayer of Mending's modifiers. Spiritual Healing and Holy Specialization reach the cast 401859, Twin Disciplines
  only the heal 401863; the sim gives all three to the heal (about 12% of Prayer of Mending if wrong, estimate).
  Inner Focus on Penance (first bolt at most). Holy Nova's "no threat" is tooltip-only. Improved Power Word: Shield before
  or after the 0.1 healing coefficient (<0.1%).
- Paladin: client data for open calls: the Oath / Blessing of Light bonus spells carry bp 1 with coefficients 0.714 / 0.429;
  Light's Vigil casts still store 75 in effect 2 (call 4).
- Shaman: Healing Stream's 5% crit rests on beta logs; the client only says it can crit.
- Engine: spell pushback amounts (1.0 / 0.8 / 0.6 / 0.4 / 0.2 s, `sim/core/cast.go:97-200`) have no tier 1-3 source.
- Druid: Stormrage and Cenarion healing bonuses are stubs; their items aren't in 70755's ItemSparse.

## Findings this pass

| Id | Severity | Title |
|---|---|---|
| healer_mechanics-20261104-1 | high | raid damage share rounded up to one target |
| healer_mechanics-20261104-2 | high | Swiftmend consumes the HoT; Forever's does not |
| healer_mechanics-20261104-3 | medium | Ancestral Fortitude raises only item armor |
| healer_mechanics-20261104-4 | low | Benediction reaches Divine Favor and the Oath |
| healer_mechanics-20261104-5 | low | Vestments of Conviction 5-piece misses Prayer of Mending |
| healer_mechanics-20261104-6 | low | Restoration shaman has no Tier 1 set |
| healer_mechanics-20261104-7 | low | direct heals miss school crit bonuses |
| healer_mechanics-20261104-8 | low | health above max not clamped |
| healer_mechanics-20261104-9 | low | wrong doc numbers (druid ranks, paladin libram) |
| healer_mechanics-20261104-10 | low | untested heals, talents and engine paths |

Measured: finding 1 with healbench raid carry on modified encounter files (shaman and priest presets); finding 2 with
healbench `-carry`, 500 iterations, on a copy with only the cancel removed. The rest are estimates, as each finding says.

## Untested (finding 10)

### Restoration shaman
1. Mid-rank table entries. No test pins Healing Wave 2-9, Lesser Healing Wave 1 and 3-5, or Chain Heal 1-2
   (base, growth, variance, mana) against the client. `TestOrcShamanHealRanks` and `TestOrcShamanHealAmounts`
   check only HW 1/10, LHW 2/6 and Riptide 1-3. A typo in, say, HW 9's mana or CH 2's base would pass.
2. Healing Wave ranks 2-3: cast times of 2 and 2.5 s and coefficients of .571 and .714. Lower-rank Chain Heal
   coefficients are also unpinned.
3. Chain Heal's own amount with healing power and its 0.714 coefficient. It is checked only indirectly, through
   the Riptide test's chain ratio and raid_healing_test's jump ratios, which divide the coefficient out.
4. Riptide ticks critting under Forever (Attributes_8 0x200), and Tidal Mastery reaching the ticks' crit.
5. Ancestral Healing and Water Shield not firing on Riptide tick crits (proc flags without periodic).
6. Nature's Swiftness not being consumed by Riptide (instant) or by a totem.
7. Riptide's 6 s client cooldown, instant heal and full mana outside Forever (the `!notes` branch). Only the
   notes' branch is tested.
8. Healing Stream's heal not getting Tidal Focus, Tidal Mastery or Purification. Purification is tested; Tidal
   Mastery and Focus are not.
9. Ancestral Fortitude not stacking with, or being refreshed by, another shaman's.

### Holy paladin
- Holy Light ranks 1-5, 7, 8 and Flash of Light ranks 1, 2, 4, 5: amounts and mana are only in the tables,
  not asserted. The presets cast Holy Light 8 (10329), and nothing tests its numbers.
- Holy Shock heal ranks 2-3 and the mana of Holy Shock casts 1-3. Light's Vigil ranks 1-2 costs, and rank 2
  amounts.
- Benediction on Holy Shock, Divine Favor and Oath costs: no healing-side test, so F1 would pass unnoticed either way.
- Divine Intellect: no test anywhere.
- Holy Power on the Holy Shock heal and the Light's Vigil heal (only Flash of Light 5 is tested), and its
  absence from Holy Shock damage crits.
- Illumination on Light's Vigil, and its use of the base cost under Libram of Economy (330, not 313.5).
- Infusion of Light at 1/2 (0.5 s), from a Holy Shock damage crit, and its 15 s expiry.
- Divine Favor on Holy Shock damage, and not being consumed by the Vigil heal.
- The Oath bonus on a downranked rank (the `downrankFactor` path).
- Lay on Hands ranks 633/2800 (the level gating, 0/250 mana), and not resetting the five-second rule.
- Healing Light and Divine Favor not reaching Light's Vigil (Healing Light's "not Light's Vigil" is in a
  comment only: the test loops over 25292, 19943 and 20930 only).
- Justice Armor 2/4/5-piece bonuses (the Holy Shock 9 s cooldown, +26 healing) and the Libram of Infusion heal crit.
- Spiritual Focus is checked only as the PushbackReduction field, not as pushback in a cast, nor stacked
  with Concentration Aura.

### Restoration druid
There is no test file in sim/druid/restoration; the druid healing tests are sim/raid_healing_druid_test.go and
sim/raid_healing_druid_talents_test.go. Tested: HT r11 (base, cost, cast), downranked HT r6/r8, Rejuv r6,
Regrowth r4 coefficients, Regrowth r9, Rejuv r11 (+Splendor), Tranq r4 (+Imp Tranq CD, party), Swiftmend
(Rejuv only, cost, CD, idol), WG r3 (+Naturalist), Bloom (+Living Spirit), NS on HT, GoN/Genesis/Imp Rejuv on
HT/RG/RJ, Tranquil Spirit (HT, Tranq), Naturalist on HT, Imp Regrowth (direct crit), Nature's Majesty, Natural
Instinct, Heart of the Wild (Int), OoC, NG, Innervate (loose), Reflection, Nature's Focus, Subtlety.
Untested (would not fail if broken):
1. Base values / costs of every rank other than HT 11, Regrowth 9, Rejuv 11, Tranq 4, WG 3 (lower-rank tables are
   only checked against themselves in the downrank test; no check against client numbers).
2. Swiftmend on a Regrowth (9 ticks with Splendor), the least-time-left pick with both HoTs up, and
   Gift of Nature / Genesis on Swiftmend; the T1 5-piece's 3 s.
3. Swiftmend non-consumption (after F1 the test must assert the HoT survives).
4. GoN on Tranquility and on Wild Growth; Genesis on Wild Growth and not on
   Tranquility.
5. Nature's Splendor on Regrowth (+2 ticks).
6. Improved Regrowth on the HoT ticks (critRate counts the action's hits; ticks not checked separately).
7. Improved Tranquility's threat cut.
8. Nature's Swiftness on Regrowth; NS not affecting Rejuv/WG/Tranq.
9. Moonglow not reducing heal costs (no test pairs Moonglow with a heal).
10. Innervate's exact effect (400% + full regen while casting; the test only wants >= 1500 mana in 20 s).
11. Tranquility and HoT tick crits (periodic-can-crit) on heals.
12. Forever T1 Restoration set 2/4-piece (+10 Spirit, +26 healing/+9 damage).
13. Bloom on Regrowth expiry (only Rejuvenation tested); Bloom crit.

### Priest
- Per-rank table values except the ranks TestHealingPriestRanks checks (LH3, Heal4, GH1, GH5, FH7, PoH5, HN6, Renew10 cost, PW:S10,
  PoM3 cost, Penance4 cost; coefs of Penance1 bolt, PoM1 heal, PW:S5, Renew5). LH1-2, Heal1-3, GH2-4, FH1-6, PoH1-4, HN1-5 base values,
  Renew1-9 and PW:S1-9 bases are guarded only by orchestration/tools/healers/priest_audit.py, not a go test.
- Heal variance (min/max spread): only averages are checked (DirectHeals within 2%).
- Spiritual Healing on FH, LH, Heal, PoH, Holy Nova, PoM (only GH, Renew, Penance, PW:S-negative tested).
- Holy Specialization on FH, PoH, HN, PoM, Penance bolts (only GH, Renew, PW:S-negative).
- Improved Healing on LH and Heal; Mental Agility on PoM and its absence on Penance/FH.
- Twin Disciplines' absence on FH and PW:S (only GH negative).
- Divine Fury on Heal ranks 1-3 / Lesser Heal negative.
- Inner Focus: no crit on Renew, free PW:S/Renew, 3 min cooldown, interaction with Penance (A2).
- Power Infusion: cost and 3 min cooldown; PI on Renew's snapshot.
- Renewed Hope: its spell list (Renew/PoH/PoM must not cut Weakened Soul or gain crit).
- Litany of Light: PW:S excluded; Holy Nova/Penance counted.
- Holy Binding: no heal on a Flash Heal at oneself.
- Divine Aegis / Inspiration from PoM, Penance and Holy Nova crits (only GH crits tested); no trigger from Renew crit ticks.
- PoM: trigger on a non-periodic heal taken, "one per caster" re-cast moving the charge, jump skips holders.
- Renew: a new rank replacing our other rank.
- Blessed Recovery carry-over of the remaining heal.
- Twilight Focus pushback behaviour (only the field value).
- Vestments of Conviction 4/5pc and the Era sets' heal bonuses (none tested; 5pc PoM part missing, F1).
- sim/priest/healing/healing_priest_test.go: character suites TestDisc/TestHoly are commented out (no regression suite for the spec).

### Engine
1. `max(count, 1)` / share below one target (TestHealingFightRaidDamageShare tests 0.375 x 4 = 1.5 only).
2. Heal assignments (`Player.heal_assignment`, heal_targets.go:9-31): no test sets one.
3. Lowest Health picking by percent rather than amount (only used as a cast target in TestHealingFightEmptyHealTarget).
4. Scares (health.go:125-127, 169-172, 256-260): no test.
5. Reaction time's health history (`healthAt`, health.go:62-72) on its own; only indirectly via TestHealingFightHealComm.
6. Raid damage events: DoT (`Ticks`), healing-taken debuff (`HealingTakenMultiplier`/`DebuffSeconds`), `Melee`
   avoidance path, physical armor path, `RandomTiming`, `BossTarget`/`Tanks` who: no test.
7. Heal crit multiplier (1.5x) and heal crit chance source (`HealingCritChance`), incl. school bonus; HoT tick crits
   under Forever (only per-class talent tests read crit chance deltas).
8. HoT snapshot rules in core (`SnapshotHeal`: healing multipliers not damage ones; healing taken applied per tick).
   dot_test.go covers damage dots only; Power Infusion's heal test is per-class.
9. Max health changes during a healing fight (finding 3).
10. Cancel's GCD handling (gcd.go:129-131) and that a cancelled heal costs nothing and doesn't start the five
    second rule (TestHealingFightHealComm checks overheal, not cost/FSR).
11. Mana Tide external aura (buffs.go:1528) and Innervate's full regen while casting under Forever (druid
    self-Innervate test exists; the core aura's casting regen is not asserted).
12. Absorb ordering with two shields / Ice Barrier on one unit; absorb on a periodic (DoT) tick.
Tested: overheal, death and its effects, tank swaps, Chain Heal jumps, dead regen, wipe end, empty heal
target, HealComm, bound ranks, share 0.375, mana tools reaching the party, PW:S absorb and unused booking,
Divine Aegis, shield and tank score, Forever Spirit regen, Forever resists.

## Next pass

Start from `last.master` a1b1163d2: re-check what the fixes for findings 1-9 change, and re-read Swiftmend, Riptide and the
other notes-only rows when a client build past 70755 lands.
