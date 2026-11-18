# Known differences

A rerun should find these, so it can tell what is new. Keep the list current. Rows as of ours
d5e59ef48, theirs 540ad8fca3 (2026-11-17, aligned). The gaps per class below are the last full
run, on the report's requests at ours dcc5483e, theirs bef9531580 (2026-11-12, 20,000
iterations, aligned), before our Venom fix (Venom's +20% proc chance now reaches Instant Poison,
as its class mask says: it moves ours about +19 on Mutilate, +16 Seal Fate Daggers, +14 Seal Fate
Swords).

Gaps after the 2026-11-17 traces, with every row below flipped (theirs - ours, what no row explains):

- **Warrior:** Arms -10.1; Fury -10.5 (Heroic Strike rage, the rotation re-check timing, Reckless
  Swings not flipped).
- **Hunter:** Survival -14.3 (Immolation Trap timing +3.5, physical hits about 1% bigger in ours
  about 5, Serpent Sting +1.1). The pet's white crit (their 7e0e0c2e46) was a gap too; ours matches
  since 2026-11-17 (Settled).
- **Paladin:** Retribution +0.5, Holy Shock Ret +0.8, Sub Holy +4.1 (Judgement choices).
- **Druid:** Feral -3.8.
- **Rogue:** Combat Daggers -3.4, Combat Swords +1.2, Mutilate -3.2, Seal Fate Daggers -3.0, Seal
  Fate Swords -3.3, Subtlety -5.1 (-10.7 with the 2026-11-17 Death Mark fixes; opener timing).

Raw gaps of the last full run (theirs - ours) and how far the rows explained them then:

- **Hunter:** ranged builds -2.5 to +5.8, within the error once the pet rows cancel. Survival +65
  (and its pet variants): Windfury Totem's mask is 70 of it, then ours sits about 25 above theirs
  (pet melee haste 27, about 10 of the player's unexplained: ours 3% more swings and casts). Pet melee
  haste and autocast are both sims' since 2026-11-12 (Settled list), so the pet's 27 should close.
- **Rogue:** Combat Daggers +7.9, Combat Swords +19.9, Mutilate +46.5, Seal Fate Daggers -4.4,
  Seal Fate Swords +40.1, Subtlety +12.7. With the Windfury mask, Venom, Shank and Thistle Tea rows,
  all within about 2 but Mutilate (+8 left), Seal Fate Swords (-8) and Seal Fate Daggers (-25 left).
  Most of Seal Fate Daggers' was the converter (fixed 2026-11-15, converter.md: theirs read
  `spellUsesRemaining` of untaken Adrenaline Rush as true, so it refreshed Slice and Dice at 5
  points and held Cold Blood); on 2026-11-15 Seal Fate Swords is about -11 and Seal Fate Daggers
  about -8, small rows (Rogue table).
- **Warrior:** Arms +149, Fury +21, Fury 2H +82. Arms and Fury 2H within about 4 with the Windfury,
  raid damage and Flurry rows; Fury leaves 13 for ours (more Heroic Strikes, 15.1 to 13.8, rage not
  traced).
- **Shaman:** Enhancement +53 to +60: the Nov 10 tuning rows move about 78 together, leaving 25
  for theirs, about 20 of it the older rows. Elemental +8.6 to +9.4: haste GCD 5.6 (both sims' rule since 2026-11-18, Settled), about 3
  left (likely the spell queue window).
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
| Which hits roll Windfury Totem | Since 2026-11-14 (Blizzard's Nov 13 hotfixes, t/2403040): every landed main-hand melee hit, white or special (Bloodthirst, Whirlwind's main-hand strike, Stormstrike, Holy Strike, Judgement of Command, Windfury Weapon's main-hand attacks), never off-hand strikes, extra attacks or seal procs (`windfuryTotemRollsOnForever`, `sim/core/buffs.go`); from 2026-11-12 to 11-14 a client flag table (Attributes_3 0x400 / Attributes_0 0x4), now gone | White swings plus abilities flagged `SpellFlagRequiresMainHand` or `SpellFlagOnNextSwing` (`sim/core/buffs/drivers.go:~314`), landed hits with no damage too | Blizzard: "Windfury Totem is now correctly activated by all melee abilities ... The main-hand strike from those abilities still does" (standards 17). Before it our log audit of 2026-11-12 (`tools/logcheck/findings.md`) matched their flag rule. Left: they need damage dealt (`wf_totem_nodamage`) |
| Windfury Totem's attack, after the Nov 13-14 hotfixes | A white extra swing: glances, spends Flurry, procs Unbridled Wrath, Reckless Swings and Omen of Clarity | A main-hand special: never glances, spends no Flurry, procs none of those (their 91aeeeebb3, push 114060) | Kept, to verify against a log (user ruling 2026-11-15, TASKS.md). Rogues 14 to 20, Feral about 5, large on warriors |
| Execute and the totem | Rolls once | Rolls twice (0e8c4a7699, 31.6% of landed Executes) | Probably a game bug: ours rolls once by default, with the double roll as a report assumption tagged "probably a bug" (user ruling 2026-11-15, TASKS.md). For theirs: Fury 11 to 18, Fury 2H 24, Arms 22 |
| Mutilate and the totem | Rolls once | Rolls twice | Ours (user ruling 2026-11-15, TASKS.md): Blizzard's Nov 13 note says Mutilate's off-hand strike can't roll the totem, and no log shows two rolls. Only a log showing two would change it. Mutilate 8.4 for theirs |
| Seal of Command, Judgement of Command and the totem | Judgement rolls it, the seal proc never | Roll it (their counts before the hotfix were 0/117 and 0/105) | Judgement of Command matched since the Nov 13 rule landed (2026-11-15, it rolls); the Seal of Command proc still never rolls on ours (`SpellFlagNoWindfuryTotem`), theirs rolls it. Open |
| Holy Strike and the totem | Rolls it (since 2026-11-15) | Rolls it (their EV-0008 after the hotfix: 58 of 277) | Matched |
| Windfury Totem's +246 attack power | A buff with 2 charges, so the next swing or ability after the extra attack takes it too (`sim/core/buffs.go:2112-2117`) | On the extra strike only (`sim/core/buffs/drivers.go:253-320`) | Open, needs a log; kept with the user's 2026-11-15 totem ruling. For ours: paladins 3.4 to 3.9, rogues 1.0 to 2.1, Survival 2.9, Feral 2.4 |
| Raid damage taken | 2 hits a minute of 1,500 (standards 24) | None | Choice. `--no-raid-damage`. Rage: Arms 17.5, Fury 18.6, Fury 2H 19.5. Casters in their tables |
| Off-GCD items (Thistle Tea, Cold Blood, Bloodrage) | Fire when the rotation says | Only at the next GCD check; Thistle Tea at energy <= 0 never fires | Choice. Thistle Tea for ours: Seal Fate Daggers 18.8 (theirs' autocast fires it 0.6 times even without the line), Feral 20.8 (2026-11-15); Combat Swords 14.8, Subtlety 10.3, Combat Daggers 6.7, Mutilate 5.8, Seal Fate Swords 4.8 (2026-11-12). Cold Blood's 1.0 a fight to 0.35 was the converter (fixed 2026-11-15, now 1.0 both) |

## Hunter

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Improved Tracking vs the generic boss | Counts the typeless boss as tracked (standards 19) | Only types a Track spell names | Choice. `--their-mob-type MobTypeGiant` aligns it. Survival 41 |
| Cobra Reflexes | Hits shrink with the speed (`hunter_pet_speed`) | Same speed normalization since 8223020463 | Ruling. No preset trains it; not remeasured |
| Hunter's buffs on the pet | Passed live | Not | Not remeasured. Their pet keeps inherited AP and health from the summon (8223020463) |
| Raptor Strike rank 8 | +96 (`hunter_raptor_31plus`) | +70 | Ruling. Survival 4.8 for ours (2026-11-17) |
| AQ book ranks | Off under Forever | Top ranks only | `ours_aq_cli` aligns it. Worth 30 to 36 to ranged hunters |
| Quick Strikes on an avoided swing | Every main-hand white swing rolls it, misses and dodges too (`sim/hunter/aspects.go:~150`; 19.5 procs from 195 swings) | Landed swings only: their decode takes a landed hit for any mask that names no outcome (`sim/core/proc_types.go`, `ProcTrigger` in `sim/hunter/aspects.go:85-125`) | Open. The client rows (1.60.1.70938) are silent: Aspect of the Beast's SpellAuraOptions has ProcTypeMask_0 4 (deal melee swing) and nothing else, and the client data has no hit mask (that is server side). Survival 15.3 for ours (2026-11-17, `sim/hunter/aspects.go:~160`) |
| Hunter's Mark on trap ticks | Immolation Trap's tick reads ranged attack power with the target's Hunter's Mark (`sim/hunter/immolation_trap.go`) | Counts it differently (not traced in their code) | Open (notes.md:313: the traps' attack power share is server side). Survival 1.2 |
| Unexplained (2026-11-17, Survival after every flip) | Immolation Trap timing; physical hits about 1% bigger | | -14.3 for ours: the trap's timing +3.5 for theirs, the physical hits about 5 for ours, Serpent Sting +1.1 for theirs; not traced further |

Pet focus now agrees: both regain 10 a second (theirs since 8223020463).

## Rogue

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Shank and Dual Wield Specialization | Off-hand multiplier, so DWS reaches it (`sim/rogue/shank.go:58`) | A modifier scoped to the off hand doesn't reach a spell with no hand in its mask | Open (notes.md Rogue, assumed). For ours: Combat Swords 7.7, Seal Fate 4.5 |
| Poised after the prepull | The rogue starts the prepull at full Energy with Poised, so the opener gets it (`sim/rogue/poised.go:45`) | Not up at the pull | Choice (the fight's start). For ours: Combat Daggers 5.2, Combat Swords 2.5 |
| Mutilate's strike order | Main hand first, then the off hand (`sim/rogue/mutilate.go:79-82`) | Off hand first | Open, not flipped. About 1.1 |
| Unexplained (2026-11-17, after every flip) | | | Combat Daggers -3.4, Combat Swords +1.2, Mutilate -3.2, Seal Fate Daggers -3.0, Seal Fate Swords -3.3; Subtlety -5.1 (opener timing; -10.7 with the Death Mark fixes) |

## Warrior

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Flurry charges | Two white swings within 500 ms spend one | Every white swing, and a missed Heroic Strike or Cleave | Assumption `flurry_batch`. Fury 10.4, Fury 2H 5.4. Their shaman Flurry (b30e3fb1d5, two logs) keeps 0.5 s for Windfury Weapon strikes. Push 114097 (2026-11-17) gave Flurry (SpellAuraOptions 126501) proc type 20, melee abilities too; both sims spend it on Heroic Strike and Cleave only of the abilities (their EV-0138). Windfury Totem's attack spends a charge in ours (a white swing, the 2026-11-15 totem ruling), none in theirs (20093d47de) |
| Anger Management | 1 rage every 3 s (tooltip) | 1.5 every 5 s (aura amount) | Open: a log of its rage ticks. Arms 0.8 (2026-11-17) |
| Rotation re-checks, Heroic Strike during a cast | Re-checks at once on Rage gain; queued Heroic Strike goes off on a swing during Slam's cast | Re-checks 200 ms after Rage gain or swing; drops it during a cast (`hardcastRuns`); holds major cooldowns during a sequence step | Choice. Moves casts, nets 0 |
| Explosives' hit | 1% miss at most (Era comment, `sim/core/consumes.go`) | Normal spell hit table | Open: client row 13241. Sapper +0.3 |
| Unexplained (2026-11-17, after every flip) | | | Arms -10.1; Fury -10.5 (Heroic Strike rage, the rotation re-check timing, Reckless Swings not flipped) |
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
| Warchief's Blessing | 15% melee and cast speed, so the spell GCD shortens to 1.304 s (`sim/core/buffs.go` ApplyWarchiefsBlessing) | Melee haste only (A_MOD_MELEE_HASTE_3, `spells_auto_gen.go:23600`) | Evidence: a beta log (Lightning Bolt 1.30 s with it, 1.505 s without, notes.md), so theirs is wrong. No report preset turns it on |
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
| Inner Focus after a DoT (Smite) | Goes off inside Shadow Word: Pain's GCD, so Devouring Plague after it is free | Next GCD check, Shadow Word: Pain takes it | Choice. 2.4 for ours (2026-11-15; mana keeps rank 8 Smite longer). Reordering on purpose gains nothing |
| Smite rank 8 | Base x1.09 (`priest_smite8_tuning`) | Client's | Ruling (2026-10-15). Smite 8.7 for ours |

Shoot and the miss floor are in the Mage table.

## Druid and paladin

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Nov 2 bleed scale | Rake, Rip, Ferocious Bite, Pounce base and AP x1.25 (`nov2BleedScale`, `sim/druid/tuning_70753.go:20`) | Client values; the same log fits Rake's tick at 0.0652 AP against our 0.06875 | Ruling (Nov 2 notes, "about 25%"). Worth 98 on our Feral; near-equal ticks either way |
| Rip and Rake ticks' attack power share | Under `nov2BleedScale`: Rip 0.0125 a combo point a tick, Rake 0.06875 | Measured on their logs: Rip 0.0114, Rake 0.065 | Ruling (`nov2BleedScale`), though their logs argue against the x1.25 on the attack power part. For ours: Rip 8.9, Rake tick 5.6. User ruling 2026-11-15: hold the x1.25 until a level 60 log |
| Eclipse charges | Granted when Wrath lands; a Starfire begun before keeps them (`sim/druid/talents.go:278-315`, notes.md:276, proc flag is a magic damage hit) | Granted at Wrath's cast; every Starfire spends one (`talents_balance.go:~165`, 7c70ac16e3) | Open: a log with a Wrath in flight. Balance 3.2 for theirs (Wrath 26.0 casts to 24.2) |
| Raid damage taken | Pushes back Starfire and Wrath | None | Choice. Balance 4.9, Balance Wrath 4.1; Feral, Retribution 0 |
| Shadowmeld into Ravage | Night Elf Shadowmeld in combat opens Ravage | Ravage and Pounce need Prowl (no source) | Open, notes.md To Verify (with the rogue's Ambush). Feral 11.1; Shadowmeld into Ambush, Subtlety 5.5 |
| Wrath rank 8 | Base x1.085 (`druid_wrath8_tuning`) | Client's | Ruling. Balance Wrath 6.9, Balance 1.8 |
| Innervate | Autocast at low mana | Only a rotation line | Choice. Balance 3.2 for ours |
| Judgement of the Crusader on weapon strikes | 161 Holy taken through the coefficient (Holy Strike via weapon percent 0.429 x 50%, Seal of Righteousness inside its base) | Full coefficient, added before the done multipliers (unit tests only) | Open, notes.md To Verify. For theirs (2026-11-15): paladins 10.1 to 10.9, Smite 5.3 |
| Dragonbreath Chili from Seal of Righteousness | Any landed melee hit rolls it | Seal hits are procs; 15852 lacks Can Proc From Procs | Open, notes.md To Verify. For ours: Retribution and Sub Holy 7.2, Holy Shock Ret 0.8 |
| Omen of Clarity on Windfury Totem's attack | Procs (a white swing, the 2026-11-15 totem ruling) | None, though their EV-0136 counts it | Open, kept with the totem ruling. Feral 3.8 for ours |
| Hammer of Wrath's 0.4 s cast | A swing due during the cast waits | The cast resets the swing | Engine. About 2 for ours on Holy Shock Ret |

Sheets before alignment (`layers.py`, explained 2026-11-15): theirs' +21 AP at the gear layer was
alignment's own `bonusStats` (layers.py keeps them from the gear layer on, and now warns when
they're there); the 136 AP for ours after buffs is Battle Shout on their pull (align.py adds it)
and the Greater Blessing of Might rank row; the paladin's 6 crit at talents is Conviction, gated
on a weapon in theirs (no DPS).

## Their bugs

Theirs differs from its own evidence or data here; ours follows the evidence. Not reported (their
checkout is read-only to us).

| Difference | Ours | Theirs | DPS |
|---|---|---|---|
| Unbridled Wrath on Whirlwind | Never: every other landed damaging melee ability procs it since push 114097 (`applyUnbridledWrath`, `sim/warrior/talents.go`) | Procs (`talents_fury.go:62-80`), against their own EV-0138 log (2 of 425 Whirlwind hits) | Fury 2.4 for theirs |
| Lacerate | Registered and cast by Survival | Never registered (`spell_data_auto_gen.go:363`) | Survival 23.8 for ours |

## Settled: both sims agree

A gap on one of these is a regression.

- Pet swing damage: (creature DPS - pet AP / 14) x 2 per 2.0 s swing. Level 60 unmeasured
  (notes.md To Verify).
- Pet focus: 10 a second.
- Pet melee haste: the pet swings at its owner's melee and attack speed (Rapid Fire, Quick Strikes,
  Wind Blessed), times its own (Frenzy, Cobra Reflexes); ranged haste (Quick Shots) doesn't reach
  it. Ours since 2026-11-12 (their d729b13992, Nov 8 log).
- Pet autocast: an ability short of focus passes to the next one on the bar. Ours since 2026-11-12
  (their 8223020463, Nov 12, logs).
- Pet Aggression: dropped (no trainer teaches it).
- Haste and the GCD: spell haste shortens a spell's GCD down to 1 s, and physical abilities keep
  1.5 s, Nature's Grace composing to 1.227 s. Warchief's Blessing's cast speed is a separate row (Shaman
  table). Theirs floors any hasted GCD at 1 s and ours keeps a base GCD already under 1 s, but no
  magic spell of ours has one (the shortest is 1 s: a mage set bonus's 0.5 s cut, the totems). Ours since 2026-11-18 (`sim/core/cast.go` hastedGCD, a combat log with Warchief's Blessing,
  notes.md). Before, ours kept 1.5 s (`haste_gcd`): Elemental 5.6, Enhancement 0.9, Affliction 0.3,
  Balance 4.0, Balance Wrath 4.4 for theirs.
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
- Ferocious Bite on a miss, dodge or parry: 80% of its 35 Energy back, extra Energy and combo
  points kept. Ours since 2026-11-15 (their 0384d8f98c, 70862 log).
- Sharpened Claws: 3% a point to melee and spell crit in the feral forms (aura 290). Ours since
  2026-11-15 (about 0 DPS).
- Seal Fate: 500 ms ICD in both.
- An avoided Execute keeps its extra Rage and refunds 80% of the base cost; the extra Rage goes only
  when it lands. Ours since 2026-11-17 (their 7025f28b98, 1.60.1.70862 logs).
- Unbridled Wrath procs on landed damaging melee abilities as well as white swings (push 114097,
  SpellAuraOptions 124699, proc type 4 to 20), Whirlwind excepted in ours (Their bugs). Ours since
  2026-11-17 (their a566a4179e, EV-0138).
- Reckless Swings: of the abilities only Heroic Strike and Cleave spend a charge, once a cast however
  many targets Cleave hits (push 114097, 249144). Ours since 2026-11-17 (their EV-0138; the push's CooldownSetLinkedSpell rows 23196 and 23197 only put Reckless Swings in the warrior's
  Cooldown Manager, UI with nothing to sim).
- The hunter pet's white swing takes none of the hunter's crit; its abilities take it. Ours since
  2026-11-17 (their 7e0e0c2e46, EV-0134).
- Predator's Edge counts twice, compounded, on Raptor Strike (x2.69 at 5/5), once on Mongoose Bite
  and Strider Kick (x2.30). Ours since 2026-11-17 (their 258ef0d164, EV-0134: 2.690 over 497 crits).
- A queued Raptor Strike keeps the dual-wield off-hand miss penalty. Ours since 2026-11-17 (their
  engine has no queued-swing exception for it; EV-0134's off-hand misses).
- Immolation and Explosive Trap roll with the hunter's spell hit, Survival Tactics on top. Ours since
  2026-11-17 (their `sim/hunter/traps.go`).
- Death Mark always hits (Attributes_3 0x40000); its cap is (156 + 245 a combo point + 0.04 AP a
  combo point) x (1 + sheet melee crit/100) x Coup de Grace, and the payout never crits. Ours since
  2026-11-17 (their e370277695, f64fef4df4).
- Rake's direct hit: 6.5% AP at every rank, read at the hit. Ours since 2026-11-12, outside the Nov 2 scale (their 0384d8f98c, 70862 log).
- Dragonfire and Caustic Smog Potions: neither hit nor ticks crit.
- Venom's +20% proc chance on Instant, Deadly and Wound Poison (ours since 2026-11-12).
- Arcane Concentration split across an Arcane Missiles channel; Wrack and Soul Siphon (hotfix
  113766).
- Searing Totem's damage per hit and rate within 3%.
- Two-Handed Weapon Specialization's Holy part follows the defense type: the melee-defense Holy hits
  (seal hits, Holy Strike, Judgement of Command) and Consecration take it; the ranged judgements,
  Exorcism, Hammer of Wrath, Holy Wrath and Holy Shock don't. Ours since 2026-11-15 (their fca3d825d9,
  EV-0064/EV-0022 logs).
- Judgement of Command is a melee special on the main hand's table, never dodged, parried or
  blocked; Righteousness and Fury are ranged since the Nov 11 hotfix. Ours since 2026-11-15 (their
  `seal_of_command.go:36`, client SpellCategories; their post-hotfix log, 1.062 n=109 under Two-Handed
  Weapon Specialization).
