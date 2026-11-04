# Known differences

A rerun should find these, so it can tell what is new. Keep the list current. Measured on the
report's requests at these commits:

| Rows | Ours | Theirs |
|---|---|---|
| Hunter and rogue | 0e87f3d | b70888a518 |
| Warrior | 7587a1a | 895cd67f99 |
| Shaman | dfa3b4a | 1496ed4db5 |
| Mage | e8d5656 | e2842f0a78 |
| Warlock | b17ce54 | bd8ef24803 |
| Priest | 3e52f94 | fbcd9a3d0f |
| Druid and paladin | 6d9b4d7 | c7eafc85cd |

How far the rows explain each gap:

- **Hunter:** the pet rows add up to the BM cat's whole pet gap (334 against 247).
- **Warrior:** Arms' whole gap; 4 DPS of Fury's and Fury 2H's unexplained.
- **Shaman:** Enhancement (33 to 37) within about 2 DPS, Elemental (5 to 6) within about 1.
- **Mage:** Frost (78.5) within 2, Frostfire (74.1) within about 5, Fire (20.4) with 10 of engine
  timing left (ours fits 0.6 more Fire Blasts, 0.25 more Pyroblasts, per-hit equal). Arcane
  (+5.4) leaves about 9 around Presence of Mind's opener.
- **Warlock:** the four presets (+15 to +19) within about 2; curse-duty variants within the error
  of their presets.
- **Priest:** Shadow -20.8, Shadow/Holy -18.4, Smite -1.7, within 0.1.
- **Druid:** Balance (+4.1) and Balance Wrath (+2.2) within about 0.6, Feral (-29.8) within about
  1; Faerie Fire duty variants within about 2 of their presets.
- **Paladin:** the four presets (-0.5 to +6.2) within about 2, with large rows that cancel.

Status: **Ruling** (the user's; don't reverse without new evidence), **Assumption** (a report
assumption key), **Evidence** (ours on evidence), **Choice** (engine or fight, not a game fact),
**Open** (needs a log or data). DPS is what the row moves.

## Hunter

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Improved Tracking vs the generic boss | Counts the typeless boss as tracked (standards 19) | Only types a Track spell names | Choice. `--their-mob-type MobTypeGiant` aligns it. Survival 41 |
| Pet focus | 10 a second (client RegenCombat, beta logs) | 25 per 5 s x Bestial Discipline | Evidence. -28 pet on BM cat |
| Cobra Reflexes | Hits shrink with the speed (`hunter_pet_speed`) | x0.85 a hit (TBC) | Ruling. +24 pet for theirs on BM cat |
| Hunter's buffs on the pet | Passed live | Not | Wolf and Tallstrider 0 to 10 with Cobra Reflexes |
| Raptor Strike rank 8 | +96 (`hunter_raptor_31plus`) | +70 | Ruling |
| AQ book ranks | Off under Forever | Top ranks only | Worth 30 to 36 to ranged hunters |

## Warrior and rogue

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Windfury Totem on a landed hit with no damage (Rend's application, Sunder Armor, Rupture's, Garrote's, Rip's) | Rolls it (`CreateExtraAttackAuraCommon`, `sim/core/buffs.go`) | Doesn't (`RequireDamageDealt`, with a TODO) | Assumption `wf_totem_nodamage`. Arms 11.9 (all its gap), Subtlety 2.5, Feral 3.4 |
| Off-GCD items (Thistle Tea, Bloodrage) | Fire when the rotation says | Only at the next GCD check | Choice. Thistle Tea at 10 energy: Combat Swords 13.5, Feral 16.2. Charge opener shifts trinket windows, cancels out. Bloodrage worth nothing |
| Raid damage taken | 2 hits a minute of 1,500 (standards 24) | None | Choice. Rage: Arms 8.5, Fury 15.7, Fury 2H 19.5. `--no-raid-damage` |
| Flurry charges | Two white swings within 500 ms spend one | Every white swing, and a missed Heroic Strike or Cleave | Assumption `flurry_batch`. Fury 11.4, Fury 2H 3.4 |
| Anger Management | 1 rage every 3 s (tooltip) | 1.5 every 5 s (aura amount) | Open: a log of its rage ticks. Under 2.1 |
| Unbridled Wrath on Heroic Strike | Procs (client proc flag 4) | White swings only | Assumption `warr_uw_queued`. Within 1 |
| Rotation re-checks, Heroic Strike during a cast | Re-checks at once on Rage gain; queued Heroic Strike goes off on a swing during Slam's cast | Re-checks 200 ms after Rage gain or swing; drops it during a cast (`hardcastRuns`); holds major cooldowns during a sequence step | Choice. Moves casts, nets 0 |
| Explosives' hit | 1% miss at most (Era comment, `sim/core/consumes.go`) | Normal spell hit table | Open: client row 13241. Sapper +0.3 |
| Touch of the Grave's roll | On any ability cast (a guess, `racials_forever.go`) | On a landed damaging hit | Open. Undead Arms 0.4, Undead Retribution 1.1 |

## Shaman

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Flametongue Totem beside Windfury Weapon | Works (`sham_ftt_stack`) | Nothing while any Windfury Totem stands | Open, notes.md To Verify. Enhancement 16.2 |
| Lightning Bolt rank 10 | Base x1.12 (`sham_lb_31plus`) | Client's | Ruling. Elemental 14.9, Enhancement 2.1 |
| Earth Shock rank 7 | Base x1.08 (`sham_es7_tuning`) | Client's | Ruling. Enhancement 4.0 |
| Stormstrike's weapon speed | Normalized + 0.3 (3.6 two-hander), fitted to beta (shaman_audit.md 4.1) | Normalized | Evidence. Enhancement 4.6; Holy Strike the same, Retribution 2.7 |
| Searing Totem's attack rate | Every 2.5 s (72 beta gaps averaged 2.53 s) | Every 2.2 s | Evidence. +3.7 for theirs on Elemental |
| Haste and the GCD | 1.5 s (`haste_gcd`) | Spell haste shortens it | Assumption. For theirs: Elemental 2.6, Enhancement 0.9; Spellstone haste Destruction 8.8, Affliction 7.5, Demonology -1.3; Balance 4.0, Balance Wrath 4.4; Retribution none |
| Magma Totem's last pulse | 20 pulses per 40 s totem | About 19 | Engine timing, no log. Enhancement 3 |
| Greater Blessing of Might and Wisdom | Rank 2's 133 AP, 40 MP5 (standards 8a) | Rank 1's 112, 36 | Ruling. `align.py` closes it |
| Blood Fury's weapon damage | +10% of weapon average damage beside +10% AP (beta sheets, logs) | AP only | Evidence (their bug, not reported). Orc Enhancement 2.9 |
| Crusader beside Windfury Weapon's strikes | Holy Strength lands at once, reaching both strikes | 10 ms later | Open, notes.md "Server batch window". Enhancement 0.6 |
| Spell queue window on cooldowns | None | A spell within 400 ms of its cooldown is queued, rotation idles (`spell_queueing.go` `CanQueue`) | Choice. Elemental -0.8 for theirs (14.5 Lava Bursts vs 13.7); Retribution about +5 for theirs, Holy Shock Ret +4. Likely moves other cooldown spells, unmeasured |
| Totem ranks | Trainer (Strength of Earth 42, Grace of Air 77) | AQ books' (53, 89) | No AQ books at launch (assumed) |

## Mage

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Ice Lance's spell power | 0.429, SoD's (`mage_ice_lance_coef`) | 1/7, TBC's | Ruling: client has none. Frost 104, Frostfire 59 |
| Frostbolt rank 10 | Base x1.04 (`mage_frostbolt10`) | Client's | Ruling. Frost 10.8 |
| Raid damage taken | Breaks Ice Barrier, pushes casts back | None | Choice. Frost 48.5, Fire 11, Arcane 9, Frostfire 6 |
| castSpell Shoot (wand) | Rotation can cut it before its first shot (`sim/core/wand.go`) | Always fires one shot; after Fire Blast it holds the next cast | Choice. Theirs loses Fire 21, Frostfire 14, Frost 13, Arcane 1; priests' last Shoot line Shadow 24.8, Shadow/Holy 21.9, Smite 6.7 |
| Spell miss floor | 1% (Classic) | None (no source) | Open (orchestration/alignment_round_2026-10-23.md: keep without new evidence). For theirs: Arcane 7.8, Smite 5.5, Shadow 0 |
| Fire Blast's range | 20 yd, +15 a point Improved Fire Blast, checked only in leveling | Checked in raids | Settled for presets by standards 33. Our raid sims let any spell land from any distance |

## Warlock

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Raid damage taken | Pushes casts back (about 1.5 fewer Shadow Bolts or Incinerates) | None | Choice. Demonology 17.2, Destruction Fire 16.4, Affliction 13.8, Destruction 10.8 |
| Conflagrate right after Immolate | Next GCD | With haste, theirs picks before Immolate lands, Conflagrate's check fails, Shadow Bolt goes first | Choice. Destruction 0.3, Destruction Fire 1.9. A rotation translation can't reach it |
| Bane of Doom's tick and 2 min cooldowns | Second Doom ticks about 121 s, after Blood Fury and trinket | Recast about 60.2 s, ticks before them | Engine timing. Destruction 3.4, others under 1% of Doom |
| The demon's start | Succubus starts in melee | Walks in, about 1.6 swings fewer | Choice. About 1 pet on Succubus builds |
| Demonic Rune's health cost | Not a damage row | A player damage row left out of DPS | Display only. `breakdown.py` lists it as the item's tag 1 |

## Priest

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Raid damage taken | Pushes back Smite and Mind Flay | None | Choice. Shadow 4.0, Shadow/Holy 3.6, Smite 3.5 |
| Inner Focus after a DoT (Smite) | Goes off inside Shadow Word: Pain's GCD, so Devouring Plague after it is free | Next GCD check, Shadow Word: Pain takes it | Choice. About 4 for ours (mana keeps rank 8 Smite longer). Reordering on purpose gains nothing |

Shoot and the miss floor are in the Mage table.

## Druid and paladin

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Raid damage taken | Pushes back Starfire and Wrath | None | Choice. Balance 4.5, Balance Wrath 4.1; Feral, Retribution 0 |
| Shadowmeld into Ravage | Night Elf Shadowmeld in combat opens Ravage | Ravage and Pounce need Prowl (no source) | Open, notes.md To Verify (with the rogue's Ambush). Feral 9.3 |
| Wrath rank 8 | Base x1.085 (`druid_wrath8_tuning`) | Client's | Ruling. Balance Wrath 6.9, Balance 1.8 |
| Innervate | Autocast at low mana | Only a rotation line | Choice. Balance 3.2 for ours |
| Judgement of the Crusader on weapon strikes | 161 Holy taken through the coefficient (Holy Strike via weapon percent 0.429 x 50%, Seal of Righteousness inside its base) | Full coefficient after done multipliers (unit tests only) | Open, notes.md To Verify. About 7 for theirs on Retribution |
| Dragonbreath Chili from Seal of Righteousness | Any landed melee hit rolls it | Seal hits are procs; 15852 lacks Can Proc From Procs | Open, notes.md To Verify. 7.3 for ours on Retribution |
| Hammer of Wrath's 0.4 s cast | A swing due during the cast waits | The cast resets the swing | Engine. About 2 for ours on Holy Shock Ret |

## Settled: both sims agree

A gap on one of these is a regression.

- Pet swing damage: (creature DPS - pet AP / 14) x 2 per 2.0 s swing. Level 60 unmeasured
  (notes.md To Verify).
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
