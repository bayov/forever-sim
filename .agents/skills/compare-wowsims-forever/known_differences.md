# Known differences

A rerun should find these, so it can tell what is new. Keep the list current. Every class was
measured on the report's requests at ours dcc5483e, theirs bef9531580 (2026-11-12, 20,000
iterations, aligned), before our Venom fix (Venom's +20% proc chance now reaches Instant Poison,
as its class mask says: it moves ours about +19 on Mutilate, +16 Seal Fate Daggers, +14 Seal Fate
Swords).

Gaps (theirs - ours) and how far the rows explain them:

- **Hunter:** ranged builds -2.5 to +5.8, within the error once the pet rows cancel. Survival +65
  (and its pet variants): Windfury Totem's mask is 70 of it, then ours sits about 25 above theirs
  (pet melee haste 27, about 10 of the player's unexplained: ours 3% more swings and casts).
- **Rogue:** Combat Daggers +7.9, Combat Swords +19.9, Mutilate +46.5, Seal Fate Daggers -4.4,
  Seal Fate Swords +40.1, Subtlety +12.7. With the Windfury mask, Venom, Shank and Thistle Tea rows,
  all within about 2 but Mutilate (+8 left), Seal Fate Swords (-8) and Seal Fate Daggers (-25 left:
  ours casts 7.2 five-point Eviscerates to 5.0 and fewer Slice and Dice points; Seal Fate's ICD or
  the combo point flow, not isolated).
- **Warrior:** Arms +149, Fury +21, Fury 2H +82. Arms and Fury 2H within about 4 with the Windfury,
  raid damage and Flurry rows; Fury leaves 13 for ours (more Heroic Strikes, 15.1 to 13.8, rage not
  traced).
- **Shaman:** Enhancement +53 to +60: the Nov 10 tuning rows move about 78 together, leaving 25
  for theirs, about 20 of it the older rows. Elemental +8.6 to +9.4: haste GCD 5.6, about 3 left
  (likely the spell queue window).
- **Mage:** Frost -78.6, Frostfire -62.1, Fire -14.9 within about 2. Arcane -26.1: Evocation 21.8,
  then raid damage and the Presence of Mind opener leave about 2.
- **Warlock:** Affliction -2.1 (Wrack rows that cancel, then about 4 to 6 each on Wrack, Corruption
  and Agony, cancelling, engine timing). Affliction Shadow Bolt +17.5, Demonology +14.1, Destruction
  +13.5, Destruction Fire +14.3: raid damage, then within about 4. Curse-duty variants within 2 to 4
  of their presets.
- **Priest:** Shadow -20.9, Shadow/Holy -18.5 within 0.2. Smite -3.9 leaves about 2.
- **Druid:** Balance +8.8 (raid 4.9, Eclipse 3.2), Balance Wrath +2.3, within about 0.7. Feral
  -7.5: Rake's hit 29.9, Shadowmeld 11.1, Thistle Tea; about 6 for ours left (Rip ticks 2.7% higher,
  more Ferocious Bites). Faerie Fire duty variants within about 2 of their presets.
- **Paladin:** presets +4.7 to +8.2. With Chili and Judgement of the Crusader flipped, Retribution
  +2.9, Sub Holy +4.1, Holy Shock Ret +0.5: theirs deals 1.7 to 3.9% more a hit on Consecration
  (20924), 10333 and 20286 at equal spell power, not traced.

Status: **Ruling** (the user's; don't reverse without new evidence), **Assumption** (a report
assumption key), **Evidence** (ours on evidence), **Choice** (engine or fight, not a game fact),
**Open** (needs a log or data). DPS is what the row moves.

## Melee, every class

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Which hits roll Windfury Totem | White main-hand swings only (`sim/core/buffs.go` `ProcMaskMeleeMHAuto`) | White swings plus abilities flagged `SpellFlagRequiresMainHand` or `SpellFlagOnNextSwing` (`sim/core/buffs/drivers.go:~314`), landed hits with no damage too; they cite Seal of Command 0/117 | Ruling 2026-11-08 and 2026-11-12 (our logs: 0 procs from about 2,770 Stormstrikes). Our log audit of 2026-11-12 (`tools/logcheck/findings.md`, 894b0596) sides with theirs: on-next-swing attacks and main-hand abilities with Attributes_3 0x400 (Sunder, Revenge, Execute, Sinister Strike, Eviscerate) roll it; Stormstrike, Holy Strike, Bloodthirst, Whirlwind, Claw, Swipe never. Waits for the user's go. For theirs: Arms 173, Fury 2H 109, Fury 62, Survival 70, Combat Swords 43, Seal Fate Swords 44, Mutilate 26, Subtlety 25, Combat Daggers 21, Seal Fate Daggers 19. Ours flipped to every main-hand hit overshoots theirs (Survival +90). Replaces `wf_totem_nodamage` |
| Raid damage taken | 2 hits a minute of 1,500 (standards 24) | None | Choice. `--no-raid-damage`. Rage: Arms 17.5, Fury 18.6, Fury 2H 19.5. Casters in their tables |
| Off-GCD items (Thistle Tea, Cold Blood, Bloodrage) | Fire when the rotation says | Only at the next GCD check; Thistle Tea at energy <= 0 never fires | Choice. Thistle Tea for ours: Combat Swords 14.8, Subtlety 10.3, Seal Fate Daggers 7.3, Combat Daggers 6.7, Mutilate 5.8, Seal Fate Swords 4.8, Feral 16.2 (not remeasured). Cold Blood 1.0 a fight to 0.35, small |

## Hunter

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Pet melee haste from the owner | None (`Pet.EnableDynamicMeleeSpeed` unused; `orchestration/pet_buffs.md:249` lists it unknown) | Pet swings at the owner's melee speed: Rapid Fire, Quick Strikes, Wind Blessed; not ranged haste (`sim/hunter/pets.go:100`, d729b13992, from a Nov 8 log: 1.965 s, 1.40 with Rapid Fire, 1.08 with Frenzy) | Open, their log. Pet for theirs: Survival 24 to 27, BM cat 9 to 10.5 |
| Pet autocast on a cooldown short of focus | Waits (`sim/hunter/pet_forever.go:185`, their old code) | Passes to the next ability (8223020463, Nov 12, logs) | Open, their log. BM Claw 66 / Bite 16.5 casts, theirs 86 / 2. For ours: pet 2.6 Survival, 3.2 BM |
| Improved Tracking vs the generic boss | Counts the typeless boss as tracked (standards 19) | Only types a Track spell names | Choice. `--their-mob-type MobTypeGiant` aligns it. Survival 41 |
| Cobra Reflexes | Hits shrink with the speed (`hunter_pet_speed`) | Same speed normalization since 8223020463 | Ruling. No preset trains it; not remeasured |
| Hunter's buffs on the pet | Passed live | Not | Not remeasured. Their pet keeps inherited AP and health from the summon (8223020463) |
| Raptor Strike rank 8 | +96 (`hunter_raptor_31plus`) | +70 | Ruling. Survival 9.1 for ours |
| AQ book ranks | Off under Forever | Top ranks only | `ours_aq_cli` aligns it. Worth 30 to 36 to ranged hunters |

Pet focus now agrees: both regain 10 a second (theirs since 8223020463).

## Rogue

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Shank and Dual Wield Specialization | Off-hand multiplier, so DWS reaches it (`sim/rogue/shank.go:58`) | A modifier scoped to the off hand doesn't reach a spell with no hand in its mask | Open (notes.md Rogue, assumed). For ours: Combat Swords 7.7, Seal Fate 4.2 |
| Seal Fate, Seal Fate Daggers | 500 ms ICD, Mutilate rolls once (`sim/rogue/talents.go:163-212`) | ICD from spell data | Candidate for Seal Fate Daggers' 25 for ours, not isolated |

## Warrior

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Flurry charges | Two white swings within 500 ms spend one | Every white swing, and a missed Heroic Strike or Cleave | Assumption `flurry_batch`. Fury 10.4, Fury 2H 5.4. Their shaman Flurry (b30e3fb1d5, two logs) keeps 0.5 s for Windfury Weapon strikes |
| Anger Management | 1 rage every 3 s (tooltip) | 1.5 every 5 s (aura amount) | Open: a log of its rage ticks. Under 2.1 |
| Unbridled Wrath on Heroic Strike | Procs (client proc flag 4) | White swings only | Assumption `warr_uw_queued`. Within 1 |
| Rotation re-checks, Heroic Strike during a cast | Re-checks at once on Rage gain; queued Heroic Strike goes off on a swing during Slam's cast | Re-checks 200 ms after Rage gain or swing; drops it during a cast (`hardcastRuns`); holds major cooldowns during a sequence step | Choice. Moves casts, nets 0 |
| Explosives' hit | 1% miss at most (Era comment, `sim/core/consumes.go`) | Normal spell hit table | Open: client row 13241. Sapper +0.3 |
| Touch of the Grave's roll | On any ability cast (a guess, `racials_forever.go`) | On a landed damaging hit | Open. Undead Arms 0.4, Undead Retribution 1.1 |

## Shaman

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Stormstrike after the Nov 10 notes | 100% weapon, 8 s cooldown (`sim/shaman/stormstrike.go:29`, `:78`) | Client rows: 150%, 6 s (SpellEffect 1368308, SpellCooldowns 55241 at push 113380) | Open (standards 37, notes-only values). 29.7 casts to 21.7; Enhancement 119 for theirs alone. Settled by a client row or a log |
| Maelstrom Weapon's rate | 5 + rank PPM, 10 at 5/5 (`sim/shaman/talents.go:565`) | 20 PPM (`talents_enhancement.go:244`, the notes' "doubling") | Open (Nov 10 notes). Enhancement 45 for theirs alone |
| Windfury Weapon beside Windfury Totem | Both proc, separate ICDs (`wf_weapon_totem_icd`, Nov 10 notes) | Windfury Weapon shuts the totem out | Open. Ours about 12 totem procs, 51 DPS. With the two rows above, 87 together (they overlap) |
| Mental Dexterity | 33/66/100% (`sim/shaman/talents.go:39`, Nov 10 notes) | Client row 415140: 50/100/150% | Open. +144.5 AP for theirs at 3/3; `align.py` closes it. 29 unaligned |
| Flametongue Totem beside Windfury Weapon | Works (`sham_ftt_stack`) | Nothing while any Windfury Totem stands | Open, notes.md To Verify. Enhancement 16.2 (not remeasured) |
| Stormstrike's weapon speed | Normalized + 0.3 (3.6 two-hander), fitted to beta (shaman_audit.md 4.1) | Normalized | Evidence. Enhancement 4.6; Holy Strike the same, Retribution 2.7 |
| Haste and the GCD | 1.5 s (`haste_gcd`) | Spell haste shortens it (`sim/core/cast.go:95`) | Assumption. For theirs: Elemental 5.6, Enhancement 0.9, Affliction 0.3; Balance 4.0, Balance Wrath 4.4; Retribution none |
| Magma Totem's last pulse | 20 pulses per 40 s totem | About 19 | Engine timing, no log. Enhancement 3 |
| Greater Blessing of Might and Wisdom | Rank 2's 133 AP, 40 MP5 (standards 8a) | Rank 1's 112, 36 | Ruling. `align.py` closes it |
| Blood Fury's weapon damage | +10% of weapon average damage beside +10% AP (beta sheets, logs) | AP only | Evidence (their bug, not reported). Orc Enhancement 2.9 |
| Crusader beside Windfury Weapon's strikes | Holy Strength lands at once, reaching both strikes | 10 ms later | Open, notes.md "Server batch window". Enhancement 0.6 |
| Spell queue window on cooldowns | None | A spell within 400 ms of its cooldown is queued, rotation idles (`spell_queueing.go` `CanQueue`) | Choice. Elemental -0.8 for theirs; Retribution about +5 for theirs, Holy Shock Ret +4 |
| Totem ranks | Trainer (Strength of Earth 42, Grace of Air 77) | AQ books' (53, 89) | No AQ books at launch (assumed) |

Gone since 2026-11-12: Lightning Bolt rank 10 and Earth Shock rank 7 (our scale is 1 since the
Nov 2 notes; Lightning Bolt equal per cast), Searing Totem's rate (theirs 2.425 s, ours 2.5 s,
equal hits).

## Mage

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Ice Lance's spell power | 0.429, SoD's (`mage_ice_lance_coef`) | 1/7, TBC's | Ruling: client has none. Frost 99.6, Frostfire 52.5 |
| Frostbolt rank 10 | Base x1.04 (`mage_frostbolt10`) | Client's | Ruling. Frost 10.7 |
| Evocation | Rotation line only; the Arcane rotation has none (`orchestration/apl_intent_audit.md:79`: loses at every length) | Autocast below 10% mana (`sim/mage/evocation.go:~31`), 0.7 casts | Choice. Arcane 21.8 for ours |
| Presence of Mind opener | Instant Arcane Wrath at 0 | The prepull Blast ends at 0, the GCD frees 1 ns later, Blast queued at 1 ns eats Presence of Mind; first Wrath at 7.45 s | Engine timing. Arcane 8.4 for ours |
| Raid damage taken | Breaks Ice Barrier, pushes casts back | None | Choice. Frost 48.5, Fire 8.4, Arcane 6.4, Frostfire 5.8 |
| castSpell Shoot (wand) | Rotation can cut it before its first shot (`sim/core/wand.go`) | Always fires one shot; after Fire Blast it holds the next cast | Choice. Theirs loses Fire 21 to 25, Frostfire 14, Frost 13; priests' last Shoot line Shadow 24.8, Shadow/Holy 21.9, Smite 6.7 |
| Spell miss floor | 1% (Classic) | None (no source) | Open (orchestration/alignment_round_2026-10-23.md: keep without new evidence). For theirs: Smite 5.8; every mage build and Shadow 0 (hit below the cap) |
| Fire Blast's range | 20 yd, +15 a point Improved Fire Blast, checked only in leveling | Checked in raids | Settled for presets by standards 33. Our raid sims let any spell land from any distance |

Both have Arcane Concentration's split across an Arcane Missiles channel (theirs 5ad9515c15).

## Warlock

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Raid damage taken | Pushes casts back (about 1.5 fewer Shadow Bolts or Incinerates) | None | Choice. Destruction 17.8, Destruction Fire 17.3, Demonology 16.3, Affliction Shadow Bolt 14.8, Affliction 4.8 |
| Bane of Agony's ramp under Wrack | Steps by tick number; ticks past the 12th stay at 150% (`sim/warlock/curses.go:~85`, the Nov 10 notes' "advances 2 seconds") | Steps by elapsed thirds of the duration (`sim/warlock/agony.go` `baneOfAgonyRamp`, TODO "tick number or time") | Open: an Agony-under-Wrack log. Theirs 12.2 lower than with ours |
| Last tick of a Wrack-retimed dot | Dropped (`sim/warlock/talents.go:~311`, f20d5938) | A truncated partial tick at expiry (`sim/core/dot.go`) | Open, no evidence. Theirs 6.8 higher |
| Conflagrate right after Immolate | Next GCD | With haste, theirs picks before Immolate lands, Conflagrate's check fails, Shadow Bolt goes first | Choice. Destruction 0.3, Destruction Fire 1.9. A rotation translation can't reach it |
| Bane of Doom's tick and 2 min cooldowns | Second Doom ticks about 121 s, after Blood Fury and trinket | Recast about 60.2 s, ticks before them | Engine timing. Destruction 3.4, others under 1% of Doom |
| The demon's start | Succubus starts in melee | Walks in, about 1.6 swings fewer | Choice. About 1 pet on Succubus builds |
| Demonic Rune's health cost | Not a damage row | A player damage row left out of DPS | Display only. `breakdown.py` lists it as the item's tag 1 |

Both have hotfix 113766 (Wrack, Soul Siphon on three dots); theirs also has DB push 114015
(2026-11-12), newer than ours. No report build casts the Imp's Firebolt.

## Priest

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Raid damage taken | Pushes back Smite and Mind Flay | None | Choice. Shadow 4.0, Shadow/Holy 3.6, Smite 3.5 |
| Inner Focus after a DoT (Smite) | Goes off inside Shadow Word: Pain's GCD, so Devouring Plague after it is free | Next GCD check, Shadow Word: Pain takes it | Choice. About 4 for ours (mana keeps rank 8 Smite longer). Reordering on purpose gains nothing |

Shoot and the miss floor are in the Mage table.

## Druid and paladin

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Rake's direct hit | No attack power (`sim/druid/rake.go`, from their PR #149 tester) | 6.5% AP, as its tick (0384d8f98c: a 1.60.1.70862 log, rank 1 read 45 on 26 of 26 at 221 AP; ranks 2 to 4 assumed equal) | Open, their log, newer than our source. Feral 29.9 for theirs, 29.4 for ours when added |
| Nov 2 bleed scale | Rake, Rip, Ferocious Bite, Pounce base and AP x1.25 (`nov2BleedScale`, `sim/druid/tuning_70753.go:20`) | Client values; the same log fits Rake's tick at 0.0652 AP against our 0.06875 | Ruling (Nov 2 notes, "about 25%"). Worth 98 on our Feral; near-equal ticks either way |
| Eclipse charges | Granted when Wrath lands; a Starfire begun before keeps them (`sim/druid/talents.go:278-315`, notes.md:276, proc flag is a magic damage hit) | Granted at Wrath's cast; every Starfire spends one (`talents_balance.go:~165`, 7c70ac16e3) | Open: a log with a Wrath in flight. Balance 3.2 for theirs (Wrath 26.0 casts to 24.2) |
| Raid damage taken | Pushes back Starfire and Wrath | None | Choice. Balance 4.9, Balance Wrath 4.1; Feral, Retribution 0 |
| Shadowmeld into Ravage | Night Elf Shadowmeld in combat opens Ravage | Ravage and Pounce need Prowl (no source) | Open, notes.md To Verify (with the rogue's Ambush). Feral 11.1 |
| Wrath rank 8 | Base x1.085 (`druid_wrath8_tuning`) | Client's | Ruling. Balance Wrath 6.9, Balance 1.8 |
| Innervate | Autocast at low mana | Only a rotation line | Choice. Balance 3.2 for ours |
| Judgement of the Crusader on weapon strikes | 161 Holy taken through the coefficient (Holy Strike via weapon percent 0.429 x 50%, Seal of Righteousness inside its base) | Full coefficient after done multipliers (unit tests only) | Open, notes.md To Verify. For theirs: Retribution and Sub Holy 9.0, Holy Shock Ret 6.9 (2 more than at the last run, not traced) |
| Dragonbreath Chili from Seal of Righteousness | Any landed melee hit rolls it | Seal hits are procs; 15852 lacks Can Proc From Procs | Open, notes.md To Verify. For ours: Retribution and Sub Holy 7.2, Holy Shock Ret 0.8 |
| Hammer of Wrath's 0.4 s cast | A swing due during the cast waits | The cast resets the swing | Engine. About 2 for ours on Holy Shock Ret |

Sheets before alignment (`layers.py`, not traced): AP 1285 ours to 1170 theirs after buffs, 701
to 722 at gear; talents give ours 6 more spell and melee crit, equal by the gear layer.

## Settled: both sims agree

A gap on one of these is a regression.

- Pet swing damage: (creature DPS - pet AP / 14) x 2 per 2.0 s swing. Level 60 unmeasured
  (notes.md To Verify).
- Pet focus: 10 a second.
- Pet Aggression: dropped (no trainer teaches it).
- Explosive reach 10 yd.
- Windfury Totem ICD 100 ms, extra attacks don't reset the swing, Eviscerate 4% AP a combo point
  (on notes.md's To Verify list).
- Two-hand white crit rage: twice the whole swing.
- Orc shaman base stats 97 Sta, 87 Int, 103 Spi; Undead 79 Str, 128 Agi, priest Spirit 130.
- Ice Lance: every rank 8 s on category 19, shared with Fire Blast (`mage_ice_lance_cd`). Both
  key Fingers of Frost by the buff's id, 400669.
- Lightning Overload's copies proc, but Judgement of Wisdom, Elemental Devastation and Water
  Shield don't hear them.
- Faerie Fire in Cat Form: 1.0 s GCD, 6 s cooldown.
- Dragonfire and Caustic Smog Potions: neither hit nor ticks crit.
- Venom's +20% proc chance on Instant, Deadly and Wound Poison (ours since 2026-11-12).
- Arcane Concentration split across an Arcane Missiles channel; Wrack and Soul Siphon (hotfix
  113766).
- Searing Totem's damage per hit and rate within 3%.
