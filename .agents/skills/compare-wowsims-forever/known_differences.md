# Known differences

A rerun should find these, so it can tell what is new. Keep the list current. Rows and gaps as of
ours 7135be55c, theirs 56473177e7 (2026-11-22, the report's requests, 20,000 iterations, aligned:
mostly attack power +21 and MP5 +4, the Greater Blessing rank row). A size with an older date, or
none, wasn't remeasured on 2026-11-22.

Gaps after the 2026-11-22 traces, with every row below flipped (theirs - ours, what no row explains):

- **Warrior:** Fury -15.0 (Heroic Strike -6.1: ours about 20 more Rage from Anger Management,
  Bloodrage 80.5 to 75.6 and refunds; off-hand white -3.7, Flurry up 154 s to 148.5; not traced
  further). Arms -13.1, Dragonkin -12.0 (Bloodrage 2.8 casts to 2.4, Heroic Strike +0.9). Fury 2H
  +2.4, adding up, but over two untraced parts that offset: Heroic Strike 16.8 for ours (7.3 casts
  to 5.0 at equal Rage) and Windfury Totem's strike 23 for theirs even as a white swing.
- **Hunter:** Survival about -13 (physical hits 1 to 2% bigger in ours, not traced). Beast Mastery:
  the player about -4.1 (Auto Shot 3.6 more casts in ours, hits 0.4% bigger, Aimed Shot's ammo),
  partly unexplained. Marksmanship -5.7 (shot mix, Hunter table).
- **Paladin:** Retribution -0.9, Holy Shock Ret +1.8, Sub Holy -0.1, Undead Sub Holy -0.5. The 1.7 to
  3.9% bigger Holy hits in theirs are gone (Settled: Two-Handed Weapon Specialization).
- **Druid:** Feral -3.8. Balance and Balance Wrath add up (Wrath +1.2, noise).
- **Rogue:** Combat Daggers -0.9 (the 2026-11-17 residual is gone with the Shank row); Combat
  Swords not traced (raw +1.9); Mutilate about -5 (the +246 attack power charges, strike order,
  off-hand Mutilate's half +1.2, about 1 Eviscerate, none flipped); Seal Fate Daggers -3.0 (Instant
  Poison about 1.5 higher in ours, not traced), Seal Fate Swords +0.3; Subtlety about -7, 5.8 of it
  physical hits 0.6% bigger in ours (the same sign as Survival's).
- **Shaman:** Enhancement 2 to 4 (main-hand white swings 64.1 to 63.5, not traced). Elemental:
  Lightning Overload's copies and rank 9 Lightning Bolt 2.1 and 1.0 for ours (mana), not traced.
- **Mage:** Frost +1.6, Fire +1.7, Frostfire -2.6, Arcane adds up.
- **Warlock:** Affliction adds up (Soul Siphon and Wrack's tick timing cancel at 7.7 each).
  Destruction -1.2 with haste's GCD off in both; Destruction Fire -1.9 (the Conflagrate row).
  Demonology -1.3 without raid damage.
- **Priest:** Shadow, Shadow/Holy and Smite add up.

Raw gaps 2026-11-22 (theirs - ours) and the rows that make them:

- **Warrior:** Fury +14.7, Fury 2H +16.6, Arms -14.8 (Dragonkin -16.8). For theirs Windfury Totem's
  strike as a special (Fury 55.5, Fury 2H 41.0, Arms 25.3) and Execute's double roll (Fury 12.7,
  Arms 25.2); for ours raid damage, Bloodthrill on the totem's swing (Arms 16.3), Unbridled Wrath
  on abilities and Flurry. Execute's double roll, Bloodthrill and Flurry on the totem's swing are
  Settled since (2026-11-22), so the next run moves these.
- **Hunter:** Beast Mastery -7.4 to -8.9, Marksmanship +1.7 to +2.2 (Lone Wolf +1.7). Survival -64.6
  (pet variants -63.6 to -64.4), -23.7 against a Giant: Improved Tracking 41.0, Lacerate 24.5, Quick
  Strikes 16.2 for ours, the totem's strike 37.2 for theirs.
- **Rogue:** Combat Daggers +4.5, Combat Swords +1.9, Mutilate +7.5, Seal Fate Daggers -9.2, Seal
  Fate Swords +13.3, Subtlety -2.5: the totem's strike 11.7 to 17.8 for theirs against Thistle Tea,
  Poised and Shadowmeld for ours.
- **Shaman:** Enhancement -17.7 to -24.5 (the Nov 10 rows are Settled): Maelstrom Weapon from the
  totem's strikes 9.0, Stormstrike's speed 4.6, Magma Totem 2.7 to 5.8, all for ours. Elemental
  +4.7, +4.8: raid damage 6.1, Lightning Shield 1.6 and Flame Shock 1.0 (fixed since) for theirs.
- **Mage:** Frost -90.0, Frostfire -68.1 (Ice Lance 111.1, 58.5), Fire -12.6, Arcane -24.8
  (Evocation 22.3).
- **Warlock:** Affliction -9.5, Affliction Shadow Bolt +13.6, Demonology +12.1, Destruction +7.0,
  Destruction Fire +14.3: raid damage, then the rows.
- **Priest:** Shadow -20.9, Shadow/Holy -18.5, Smite -3.9: Shoot and raid damage.
- **Druid:** Balance +4.9, Balance Wrath -2.2. Feral -51.8: Thistle Tea 22.4, Shadowmeld 11.1, Rip
  9.4, Rake 5.7.
- **Paladin:** Retribution +28.7, Holy Shock Ret +22.6 (duty variants +25.0, +26.7, not traced), Sub
  Holy +32.7, Undead +24.3: for theirs the Seal of Command proc rolling the totem (11.0 to 21.2), the
  totem's strike (7.6 to 8.7), Judgement of the Crusader on weapon strikes (4.8 to 8.0) and, on Holy
  Shock Ret, the Judgement wake-up (6.6).

Status: **Ruling** (the user's; don't reverse without new evidence), **Assumption** (a report
assumption key), **Evidence** (ours on evidence), **Choice** (engine or fight, not a game fact),
**Open** (needs a log or data). DPS is what the row moves.

## Melee, every class

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Which hits roll Windfury Totem | Since 2026-11-14 (Blizzard's Nov 13 hotfixes, t/2403040): every landed main-hand melee hit, white or special (Bloodthirst, Whirlwind's main-hand strike, Stormstrike, Holy Strike, Judgement of Command, Windfury Weapon's main-hand attacks), never off-hand strikes, extra attacks or seal procs (`windfuryTotemRollsOnForever`, `sim/core/buffs.go`); from 2026-11-12 to 11-14 a client flag table (Attributes_3 0x400 / Attributes_0 0x4), now gone | White swings plus abilities flagged `SpellFlagRequiresMainHand` or `SpellFlagOnNextSwing` (`sim/core/buffs/drivers.go:~314`), landed hits with no damage too | Blizzard: "Windfury Totem is now correctly activated by all melee abilities ... The main-hand strike from those abilities still does" (standards 17). Before it our log audit of 2026-11-12 (`tools/logcheck/findings.md`) matched their flag rule. Left: they need damage dealt (`wf_totem_nodamage`) |
| Windfury Totem's attack, after the Nov 13-14 hotfixes | A white extra swing: glances, procs Unbridled Wrath, Reckless Swings and Omen of Clarity; since 2026-11-22 it spends no Flurry and opens no Bloodthrill (Settled) | A main-hand special: never glances, spends no Flurry, procs none of those but Unbridled Wrath (their 91aeeeebb3, push 114060; Unbridled Wrath hears it as a swing, EV-0136, `talents_fury.go:75-89`) | Kept, to verify against a log (user ruling 2026-11-15, TASKS.md). For theirs (2026-11-22): Fury 55.5, Fury 2H 41.0, Survival 37.2, Arms 25.3, Seal Fate Swords 17.8, Subtlety 15.7 (with the +246 row), Mutilate 14.2, Seal Fate Daggers 12.7, Combat Daggers 11.7, Sub Holy 8.7 (Undead 8.3), Retribution 8.5, Holy Shock Ret 7.6; Feral's whole bundle (no glance, Flurry, Omen, +246) nets 3.7 |
| Seal of Command, Judgement of Command and the totem | Judgement rolls it, the seal proc never (`SpellFlagNoWindfuryTotem`, `sim/core/buffs.go:~2150-2200`) | Both roll it: triggered hits flagged Not a Proc roll as abilities do (`sim/core/buffs/drivers.go:~307`; their counts before the hotfix were 0/117 and 0/105) | Judgement of Command matched since the Nov 13 rule landed (2026-11-15). The seal proc (20424 and its echo 1311703) is Open. For theirs (2026-11-22, first sizes): Retribution 21.2, Sub Holy 21.1 (Undead 16.5), Holy Shock Ret 11.0 to 11.7 |
| Holy Strike and the totem | Rolls it (since 2026-11-15) | Rolls it (their EV-0008 after the hotfix: 58 of 277) | Matched |
| Windfury Totem's +246 attack power | A buff with 2 charges, so the next swing or ability after the extra attack takes it too (`sim/core/buffs.go:2112-2117`) | On the extra strike only (`sim/core/buffs/drivers.go:253-320`) | Open, needs a log; kept with the user's 2026-11-15 totem ruling. 2026-11-22: Holy Shock Ret 1.8 for ours, Retribution about 1.7, Arms 2.7; rogues 1.0 to 2.1, Survival 2.9, Feral 2.4 (2026-11-15) |
| Raid damage taken | 2 hits a minute of 1,500 (standards 24) | None | Choice. `--no-raid-damage`. Rage (2026-11-22): Fury 20.2, Arms 20.1, Fury 2H 19.2. Casters and hunters in their tables |
| Off-GCD items (Thistle Tea, Cold Blood, Bloodrage) | Fire when the rotation says | Only at the next GCD check; Thistle Tea at energy <= 0 never fires | Choice. Thistle Tea for ours (2026-11-22): Feral 22.4, Seal Fate Daggers 18.9 (theirs casts it 0 times), Seal Fate Swords about 4.5 (timing), Combat Daggers 1.8 net; Combat Swords 14.8, Subtlety 10.3, Mutilate 5.8 (2026-11-12). Cold Blood's 1.0 a fight to 0.35 was the converter (fixed 2026-11-15, now 1.0 both) |
| Judgement of Wisdom on avoided white swings | Missed and dodged swings roll it too (`sim/core/debuffs.go:~578`, a Classic comment; `sim/shaman/enhancement/mana_returns_test.go:13-41` asserts it, 9eef440fb) | Landed swings only (their 9f254e8d87, 2026-11-09, a level 41 log with the miss counts in their DB row; `sim/core/judgement_of_wisdom_test.go`) | Open, likely ours: ask the user, don't fix unasked. Survival 3.8 for ours (2026-11-22, mana) |

## Hunter

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Improved Tracking vs the generic boss | Counts the typeless boss as tracked (standards 19) | Only types a Track spell names | Choice. `--their-mob-type MobTypeGiant` aligns it. Survival 41.0 (2026-11-22) |
| Cobra Reflexes | Hits shrink with the speed (`hunter_pet_speed`) | Same speed normalization since 8223020463 | Ruling. No preset trains it; not remeasured |
| The pet's attack power after the summon | Fixed at the summon since fix-petap (`sim/hunter/pet.go`, keepSummonStats): the pet is summoned at the reset, so the owner's permanent buffs count; since 2026-11-23 the APL's prepull Aspect of the Hawk counts too (`summonedWithHawk`, the summon taken with the usual aspect up, a default), where theirs leaves it out | Fixed at the summon (`pets.go:144-155`, keepSummonStats, 8223020463, EV-0072: in the Nov 7 and 8 logs it holds through 20 Aspect of the Hawk toggles and a gear swap in one summon); health held too, armor follows the owner | Settled (fix-petap, after the user's 2026-11-22 ruling on log-confirmed rules). Ours inherits no health or armor, so only the attack power changes. Golden: Beast Mastery about -5, Marksmanship -2.5, Survival -2 |
| Raptor Strike rank 8 | +96 (`hunter_raptor_31plus`) | +70 | Ruling. Survival 5.2 for ours (2026-11-22) |
| AQ book ranks | Off under Forever | Top ranks only | `ours_aq_cli` aligns it. Worth 30 to 36 to ranged hunters |
| Quick Strikes on an avoided swing | Every main-hand white swing rolls it, misses and dodges too (`sim/hunter/aspects.go:~150`; 19.5 procs from 195 swings) | Landed swings only: their decode takes a landed hit for any mask that names no outcome (`sim/core/proc_types.go`, `ProcTrigger` in `sim/hunter/aspects.go:85-125`) | Open. The client rows (1.60.1.70938) are silent: Aspect of the Beast's SpellAuraOptions has ProcTypeMask_0 4 (deal melee swing) and nothing else, and the client data has no hit mask (that is server side). Survival 16.2 for ours (2026-11-22, `sim/hunter/aspects.go:~160`) |
| Quick Strikes, the pet's part | The extra procs reach the pet, which swings at the hunter's melee haste (Settled) | Fewer procs, so less | Follows the row above. 2.6 for ours on Survival (2026-11-22, new) |
| Ranged Weapon Specialization on Serpent Sting | Left out (`sim/hunter/talents.go:103`, no comment or evidence) | Raises it (`talents_marksmanship.go:243`; their weapon-requirements.md: a per-weapon damage done mod of the mask's schools on ranged attacks) | Open: 19507 is A_MOD_DAMAGE_PERCENT_DONE, Misc 127 (every school), ranged weapons, so the school doesn't leave Nature out; whether a weapon-gated damage done mod reaches a DoT needs a Marksmanship log of Serpent Sting ticks. Marksmanship 7.9 for theirs (2026-11-22) |
| Tallstrider's Dust Cloud | Cast only when no stronger armor cut is up (`sim/hunter/pet_forever.go:400-417`) | An always-on filler that stacks (`pets.go:453-466`), so fewer Bites | Open, no log. Tallstrider Beast Mastery 1.5 for ours (2026-11-22) |
| Raid damage taken | Pushes back shots | None | Choice. Marksmanship 1.9 (2026-11-22) |
| Hunter's Mark on trap ticks | Immolation Trap's tick reads ranged attack power with the target's Hunter's Mark (`sim/hunter/immolation_trap.go`) | Counts it differently (not traced in their code) | Open (notes.md:313: the traps' attack power share is server side). Survival 1.2 |
| Unexplained (2026-11-22, after every flip) | Physical hits 1 to 2% bigger in ours (Subtlety's 0.6% has the same sign) | | Survival about 13 for ours. Beast Mastery's player about 4.1 for ours: Auto Shot 3.6 casts more (clipping timing), hits 0.4% bigger, Aimed Shot's ammo at the item's 2.8 speed to 2.5 (`sim/hunter/aimed_shot.go:78`), partly unexplained. Marksmanship 5.7 for ours: shot mix (ours, theirs: Arcane Shot 34.5, 36.4 casts; Aimed Shot 19.6, 18.5; theirs' Auto Shot has a 0.307 s cast). The 2026-11-17 Immolation Trap timing and Serpent Sting parts not remeasured |

Pet focus now agrees: both regain 10 a second (theirs since 8223020463).

## Rogue

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Poised after the prepull | The rogue starts the prepull at full Energy with Poised, so the opener gets it (`sim/rogue/poised.go:45`) | Not up at the pull | Choice (the fight's start). For ours: Combat Daggers 9.4 (2026-11-22, the whole prepull block), Combat Swords 2.5 |
| Mutilate's strike order | Main hand first, then the off hand (`sim/rogue/mutilate.go:79-82`) | Off hand first | Open, not flipped. About 1.1 |
| Death Mark's attack power and crit | Read at the cast, capped once at the detonation (path not traced) | Read live at each store, clamped each time (`death_mark.go:44,59`) | Open. Subtlety 3.7 for ours (2026-11-22, new) |
| Deadly Poison's damage multipliers | Snapshot at the application (`sim/rogue/poisons.go:~185-199`) | Read at each tick (not traced) | Open. Mutilate 1.4 for ours (2026-11-22, new) |
| Shank's weapon damage | Normalized (`OHNormalizedWeaponDamage`, `sim/rogue/shank.go:~40`, notes.md:360 an assumption) | Not normalized (`OHWeaponDamage`, `sim/rogue/shank.go:40`) | Open. Seal Fate 0.4 to 0.5 for ours (2026-11-22, new) |
| Unexplained (2026-11-22, after every flip) | | | Combat Daggers -0.9; Mutilate about -5; Seal Fate Daggers -3.0, Seal Fate Swords +0.3; Subtlety about -7, 5.8 of it physical hits 0.6% bigger in ours |

## Warrior

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Flurry charges | Two white swings within 500 ms spend one | Every white swing, and a missed Heroic Strike or Cleave | Assumption `flurry_batch`. Fury 9.0 (2026-11-22), Fury 2H 5.4. Their shaman Flurry (b30e3fb1d5, two logs) keeps 0.5 s for Windfury Weapon strikes. Push 114097 (2026-11-17) gave Flurry (SpellAuraOptions 126501) proc type 20, melee abilities too; both sims spend it on Heroic Strike and Cleave only of the abilities (their EV-0138). Windfury Totem's attack spends none in either since 2026-11-22 (Settled) |
| Unbridled Wrath on abilities | Every landed damaging melee ability but Whirlwind, since push 114097 (`applyUnbridledWrath`, `sim/warrior/talents.go:~213-262`, 2026-11-17; SpellAuraOptions 124699, proc type 4 to 20) | White swings, the totem's strike, Heroic Strike and Cleave only (their 91c37d72cf, 2026-11-18, EV-0148: Shield Slam 0 of 977, Revenge 0 of 601, Thunder Clap 0 of 2238, Whirlwind 2 of 425; no Bloodthirst or Arms ability log) | Open. For ours (2026-11-22): Fury 5.8, Arms 4.9 to 6.8 |
| A queued Heroic Strike and the totem's extra attack | The extra attack takes the queued Heroic Strike (`sim/core/attack.go:~397`, replaceSwing) | Not | Engine. Nets about 0 on Fury and Arms, moves DPS between rows (2026-11-22, new) |
| Anger Management | 1 rage every 3 s (tooltip) | 1.5 every 5 s (aura amount) | Open: a log of its rage ticks. Arms 1.6, Fury noise (2026-11-22) |
| Rotation re-checks, Heroic Strike during a cast | Re-checks at once on Rage gain; queued Heroic Strike goes off on a swing during Slam's cast | Re-checks 200 ms after Rage gain or swing; drops it during a cast (`hardcastRuns`); holds major cooldowns during a sequence step | Choice. Moves casts, nets 0 |
| Explosives' hit | 1% miss at most (Era comment, `sim/core/consumes.go`) | Normal spell hit table | Open: client row 13241. Sapper +0.3 |
| Unexplained (2026-11-22, after every flip) | | | Fury -15.0 (Heroic Strike -6.1 on about 20 more Rage in ours; off-hand white -3.7 on Flurry uptime), Arms -13.1, Dragonkin -12.0; Fury 2H +2.4 over Heroic Strike 16.8 for ours and the totem's strike 23 for theirs, offsetting, untraced |
| Touch of the Grave's roll | On any ability cast (a guess, `racials_forever.go`) | On a landed damaging hit | Open. Undead Arms 0.4, Undead Retribution 1.1 |

## Shaman

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Windfury Weapon beside Windfury Totem | Both proc, separate ICDs (`wf_weapon_totem_icd`, Nov 10 notes) | Windfury Weapon shuts the totem out | Open. Ours about 12 totem procs, 51 DPS (before 2026-11-17, not remeasured; the 2026-11-22 Enhancement gap leaves little room for it) |
| Flametongue Totem beside Windfury Weapon | Works (`sham_ftt_stack`) | Nothing while any Windfury Totem stands | Open, notes.md To Verify. Enhancement 16.2 (not remeasured) |
| Stormstrike's weapon speed | Normalized + 0.3 (3.6 two-hander), fitted to beta (shaman_audit.md 4.1) | Normalized | Evidence. Enhancement 4.6 (2026-11-22); Holy Strike the same, Holy Shock Ret 2.8 (2026-11-22), Retribution 2.7 |
| Warchief's Blessing | 15% melee and cast speed, so the spell GCD shortens to 1.304 s (`sim/core/buffs.go` ApplyWarchiefsBlessing) | Melee haste only (A_MOD_MELEE_HASTE_3, `spells_auto_gen.go:23600`) | Evidence: a beta log (Lightning Bolt 1.30 s with it, 1.505 s without, notes.md), so theirs is wrong. No report preset turns it on |
| Magma Totem's last pulse | 20 pulses per 40 s totem | About 19 | Engine timing, no log. Enhancement 2.7 to 5.8 (2026-11-22) |
| Greater Blessing of Might and Wisdom | Rank 2's 133 AP, 40 MP5 (standards 8a) | Rank 1's 112, 36 | Ruling. `align.py` closes it (the AP +21, MP5 +4 of most runs) |
| Blood Fury's weapon damage | +10% of weapon average damage beside +10% AP (beta sheets, logs) | AP only | Evidence (their bug, not reported). Orc Enhancement 2.9 |
| Crusader beside Windfury Weapon's strikes | Holy Strength lands at once, reaching both strikes | 10 ms later | Open, notes.md "Server batch window". Enhancement 0.6 |
| Spell queue window on cooldowns | None | A spell within 400 ms of its cooldown is queued, rotation idles (`spell_queueing.go` `CanQueue`) | Choice. Elemental -0.8 for theirs (2026-11-22); Retribution about +5 for theirs, Holy Shock Ret +4 |
| Raid damage taken | Pushes back casts | None | Choice. Elemental 6.1 (2026-11-22, new) |
| Lightning Shield's orbs | Paced differently (not traced) | | Choice (pacing). Elemental 1.6 for theirs (2026-11-22, new) |
| Totem ranks | Trainer (Strength of Earth 42, Grace of Air 77) | AQ books' (53, 89) | No AQ books at launch (assumed) |

Gone since 2026-11-12: Lightning Bolt rank 10 and Earth Shock rank 7 (our scale is 1 since the
Nov 2 notes; Lightning Bolt equal per cast), Searing Totem's rate (theirs 2.425 s, ours 2.5 s,
equal hits). Since 2026-11-22 the Nov 10 tuning rows (Stormstrike, Maelstrom Weapon's rate,
Mental Dexterity) are Settled.

## Mage

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Frostbolt rank 10 | Base x1.04 (`mage_frostbolt10`) | Client's | Ruling. Frost 10.3 (2026-11-22) |
| Evocation | Rotation line only; the Arcane rotation has none (`orchestration/apl_intent_audit.md:79`: loses at every length) | Autocast below 10% mana (`sim/mage/evocation.go:~31`), 0.7 casts | Choice. Arcane 22.3 for ours (2026-11-22) |
| Presence of Mind opener | Instant Arcane Wrath at 0 | The prepull Blast ends at 0, the GCD frees 1 ns later, Blast queued at 1 ns eats Presence of Mind; first Wrath at 7.45 s | Engine timing. Arcane about 8 for ours (2026-11-22) |
| Raid damage taken | Breaks Ice Barrier, pushes casts back | None | Choice. 2026-11-22: Frost 43 to 49, Fire 8.5, Arcane 6.4, Frostfire 5.8 |
| castSpell Shoot (wand) | Rotation can cut it before its first shot (`sim/core/wand.go`) | Always fires one shot; after Fire Blast it holds the next cast | Choice. Theirs loses (2026-11-22) Fire 23.6, Frostfire 12.7, Frost 12.0; priests' last Shoot line Shadow 24.8, Shadow/Holy 22.5, Smite 6.7 |
| Spell miss floor | 1% (Classic) | None (no source) | Open (orchestration/alignment_round_2026-10-23.md: keep without new evidence). For theirs: Smite 5.8; every mage build and Shadow 0 (hit below the cap) |
| Fire Blast's range | 20 yd, +15 a point Improved Fire Blast, checked only in leveling | Checked in raids | Settled for presets by standards 33. Our raid sims let any spell land from any distance |

Both have Arcane Concentration's split across an Arcane Missiles channel (theirs 5ad9515c15).
Arcane's sheet at the talent layer shows spell crit 3.1 in ours to 6.1; crit per ability agrees
(display only).

## Warlock

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Raid damage taken | Pushes casts back (about 1.5 fewer Shadow Bolts or Incinerates) | None | Choice. 2026-11-22: Affliction Shadow Bolt 16.5, Destruction Fire 16.3, Destruction 13.4, Demonology 13.4, Affliction 4.1 |
| Bane of Agony's ramp under Wrack | Steps by tick number; ticks past the 12th stay at 150% (`sim/warlock/curses.go:~85`, the Nov 10 notes' "advances 2 seconds") | Steps by elapsed thirds of the duration (`sim/warlock/agony.go` `baneOfAgonyRamp`, TODO "tick number or time") | Open: an Agony-under-Wrack log. Theirs 14.1 lower than with ours (2026-11-22) |
| Soul Siphon on Wrack | Counts the target's afflictions once, when the channel lands (`sim/warlock/wrack.go:69`) | Recounts them each tick (`talents_affliction.go:237-261`, TODO in 23b739be95) | Open (notes.md:249 lists it; a Drain Life log settles it). Affliction 7.7 for ours (2026-11-22, new) |
| Haste and the GCD on Destruction | Spellstone's 2% cast speed (`sim/warlock/talents.go:152`) shortens the GCD; turning ours' hastedGCD off costs ours 5.6 | The same formula and 1 s floor, but turning it off costs theirs 0.3 | Open, driver not found: does theirs' Spellstone haste reach the GCD? Dropping theirs' activeImmolate check raised Conflagrate casts (ours 14.6, theirs 13.2) but not DPS. About 5.3 for ours (2026-11-22, new); off in both, Destruction -1.2 |
| Conflagrate right after Immolate | Next GCD | With haste, theirs picks before Immolate lands, Conflagrate's check fails, Shadow Bolt goes first | Choice. Destruction Fire 1.9 (2026-11-22); on Destruction the haste row above holds it. A rotation translation can't reach it |
| Bane of Doom's tick and 2 min cooldowns | Second Doom ticks about 121 s, after Blood Fury and trinket | Recast about 60.2 s, ticks before them | Engine timing. Destruction 3.4, others under 1% of Doom |
| The demon's start | Succubus starts in melee | Walks in, about 1.6 swings fewer | Choice. Pet about 1.2 on Demonology, 1.0 on Destruction (2026-11-22) |
| Demonic Rune's health cost | Not a damage row | A player damage row left out of DPS | Display only. `breakdown.py` lists it as the item's tag 1 |

Both have hotfix 113766 (Wrack, Soul Siphon on three dots); theirs also has DB push 114015
(2026-11-12), newer than ours. No report build casts the Imp's Firebolt.

## Priest

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Raid damage taken | Pushes back Smite and Mind Flay | None | Choice. Shadow 4.1, Shadow/Holy 3.8 (2026-11-22), Smite 3.5 |
| Inner Focus after a DoT (Smite) | Goes off inside Shadow Word: Pain's GCD, so Devouring Plague after it is free | Next GCD check, Shadow Word: Pain takes it | Choice. 2.4 for ours (2026-11-15; mana keeps rank 8 Smite longer). Reordering on purpose gains nothing |
| Smite rank 8 | Base x1.09 (`priest_smite8_tuning`) | Client's | Ruling (2026-10-15). Smite 8.7 for ours |

Shoot and the miss floor are in the Mage table.

## Druid and paladin

| Difference | Ours | Theirs | Status, DPS |
|---|---|---|---|
| Nov 2 bleed scale | Rake, Rip, Ferocious Bite, Pounce base and AP x1.25 (`nov2BleedScale`, `sim/druid/tuning_70753.go:20`) | Client values; the same log fits Rake's tick at 0.0652 AP against our 0.06875 | Ruling (Nov 2 notes, "about 25%"). Worth 98 on our Feral; near-equal ticks either way |
| Rip and Rake ticks' attack power share | Under `nov2BleedScale`: Rip 0.0125 a combo point a tick, Rake 0.06875 | Measured on their logs: Rip 0.0114, Rake 0.065 | Ruling (`nov2BleedScale`), though their logs argue against the x1.25 on the attack power part. For ours (2026-11-22): Rip 9.4, Rake tick 5.7. User ruling 2026-11-15: hold the x1.25 until a level 60 log. Our `sim/druid/rake.go:25-29` comment still cites their PR #149's 0.055; their data is 0.065 now (stale comment) |
| Eclipse charges | Granted when Wrath lands; a Starfire begun before keeps them (`sim/druid/talents.go:278-315`, notes.md:276, proc flag is a magic damage hit) | Granted at Wrath's cast; every Starfire spends one (`talents_balance.go:~165`, 7c70ac16e3) | Open: a log with a Wrath in flight. For theirs (2026-11-22): Balance 2.3, Balance Wrath 0 |
| Raid damage taken | Pushes back Starfire and Wrath | None | Choice. Balance 5.0, Balance Wrath 3.6 (2026-11-22); Feral, Retribution 0 |
| Shadowmeld into Ravage | Night Elf Shadowmeld in combat opens Ravage | Ravage and Pounce need Prowl (no source) | Open, notes.md To Verify (with the rogue's Ambush). Feral 11.1; Shadowmeld into Ambush, Subtlety 7.8 (2026-11-22) |
| Wrath rank 8 | Base x1.085 (`druid_wrath8_tuning`) | Client's | Ruling. Balance Wrath 7.0, Balance 1.8 (2026-11-22) |
| Innervate | Autocast at low mana | Only a rotation line | Choice. Balance about 0.6 for ours (2026-11-22) |
| Shifting Power's casts | 21.6 a fight | 21.3 | Not traced (timing). Feral about 4 for ours (2026-11-22, new) |
| Judgement of the Crusader on weapon strikes | 161 Holy taken through the coefficient (Holy Strike via weapon percent 0.429 x 50%, Seal of Righteousness inside its base) | Full coefficient, added before the done multipliers (unit tests only) | Open, notes.md To Verify. For theirs (2026-11-22): Sub Holy 7.0 (Undead 8.0), Retribution 6.6, Holy Shock Ret 4.8; Smite 5.3 (2026-11-15) |
| Dragonbreath Chili from Seal of Righteousness | Any landed melee hit rolls it | Seal hits are procs; 15852 lacks Can Proc From Procs | Open, notes.md To Verify. For ours (2026-11-22): Retribution 4.8, Sub Holy 4.0 (Undead 6.9); Holy Shock Ret 0.8 |
| Omen of Clarity on Windfury Totem's attack | Procs (a white swing, the 2026-11-15 totem ruling) | None, though their EV-0136 counts it | Open, kept with the totem ruling. Part of Feral's totem bundle, 3.7 net for theirs (2026-11-22, Melee table) |
| Judgement's cooldown and the rotation | Judgement is off the GCD; the rotation sees it ready only at the next GCD or swing, drifting about 0.02 s a cycle and later behind Consecration | Wakes the rotation at Judgement's cooldown end (`sim/paladin/judgement.go:50-56`) | Engine, a candidate fix in ours (any off-GCD spell wakes the rotation at its cooldown end). Holy Shock Ret 6.6 for theirs (2026-11-22, new; Judgement of Command 22.36 casts to 22.99). Retribution's Judgement of Righteousness 5.4 casts to 6.65, offset by Seal of Righteousness procs (timing, not traced) |
| Hammer of Wrath's 0.4 s cast | A swing due during the cast waits | The cast resets the swing | Engine. About 4 for ours on Holy Shock Ret (2026-11-22; 5.39 casts to 4.96) |
| Shadow Oil's procs | 6.69 a fight | 7.20 | Not traced. Holy Shock Ret about 2 for theirs (2026-11-22, new) |

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
| Lacerate | Registered and cast by Survival | Never registered (`spell_data_auto_gen.go:363`) | Survival 24.5 for ours (2026-11-22) |

Unbridled Wrath on Whirlwind left this table on 2026-11-22: theirs no longer procs it (91c37d72cf),
and the difference is now ours procing it on every other ability (Warrior table).

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
  table). Both floor every hasted GCD at 1 s, a base GCD under 1 s included (user ruling 2026-11-18,
  unverified in a log). Ours since 2026-11-18 (`sim/core/cast.go` hastedGCD, a combat log with Warchief's Blessing,
  notes.md). Before, ours kept 1.5 s (`haste_gcd`): Elemental 5.6, Enhancement 0.9, Affliction 0.3,
  Balance 4.0, Balance Wrath 4.4 for theirs. On 2026-11-22 it's worth 4.0 to ours and 4.8 to theirs
  on Elemental, 8.8 and 2.3 to ours on Balance and Balance Wrath; Destruction differs (Warlock table).
- Flame Shock's GCD shortens with spell haste like any spell's. Ours since 2026-11-22 (56e5f070d:
  its `IgnoreHaste` was a Season of Discovery carryover, kept for Classic); before, Elemental 1.0 for
  theirs.
- Explosive reach 10 yd.
- Windfury Totem ICD 100 ms, extra attacks don't reset the swing, Eviscerate 4% AP a combo point
  (on notes.md's To Verify list).
- Mutilate rolls Windfury Totem once, as the user's 2026-11-15 ruling had ours (Blizzard's Nov 13
  note). Theirs since bd717cfa80 (2026-11-21, EV-0155, push 114313: the cast's mask is
  ProcMaskEmpty); before, 8.4 on Mutilate for theirs.
- Execute rolls Windfury Totem twice under Forever: the cast on its own, whatever the damage does,
  then the landed damage (`RollWindfuryTotemOnCast`, sim/warrior/execute.go; report assumption
  `warr_exec_wf2`, Low once). Ours since 2026-11-22 (their 0e8c4a7699; EV-0155 after push 114313,
  1.68 and 2.09 on 89 warriors where two rolls give 1.81; EV-0002 31.6% of landed Executes). It
  flips with pushes: one roll on 2026-11-20 before 114313 (1.01 to 1.05 on 112 warriors). Before,
  for theirs Arms 25.2, Fury 12.7, Fury 2H 24.
- Windfury Totem's attack opens no Bloodthrill. Ours since 2026-11-22, though ours still swings it
  white (their `talents_arms.go:~268-273`; EV-0056, EV-0136: 10 of 1,633 totem hits, 0.61%, no more
  than the 0.74% the ability before gives alone, against 3.5% a white swing). Before, Arms 16.3 for
  ours.
- Windfury Totem's attack spends no warrior Flurry charge; its crit applies or refreshes Flurry.
  Ours since 2026-11-22 (their 20093d47de; EV-0138 and EV-0140, build 70938 to push 114115: 68 of
  74 totem hits with Flurry up removed none, the other 6 explained; 22 of 22 crits gave Flurry).
- Two-hand white crit rage: twice the whole swing.
- Orc shaman base stats 97 Sta, 87 Int, 103 Spi; Undead 79 Str, 128 Agi, priest Spirit 130.
- Ice Lance: every rank 8 s on category 19, shared with Fire Blast (`mage_ice_lance_cd`). Both
  key Fingers of Frost by the buff's id, 400669.
- Ice Lance's spell power coefficient 0.11 on every rank (`mage_ice_lance_coef`). Ours since
  2026-11-22 (user ruling: log-confirmed rules from theirs override ours; their 543cc309aa, EV-0171,
  one level 60 mage's logs on 70959 and 71037: 28 hits at 0 spell power, 255 geared, 0.115 ± 0.007);
  before, ours 0.429 (SoD's server value, the 2026-10-15 ruling): Frost 111.1, Frostfire 58.5 for ours.
  The same logs show Draconic Infused Emblem's +35 spell damage not raising Ice Lance; neither sim
  models that.
- Lightning Overload's copies proc, but Judgement of Wisdom, Elemental Devastation and Water
  Shield don't hear them.
- Stormstrike 100% weapon damage on an 8 s cooldown, and Maelstrom Weapon at 6 to 10 PPM by rank:
  client 1.60.1.70959 carries both (theirs reads it, `talents_enhancement.go:244`; ours cites it
  since 2026-11-17, 30a645072). Mental Dexterity: no gap on the 2026-11-22 sheets (the aligned
  attack power is the Blessing rank's 21 alone); ours keeps the Nov 10 notes' 33/66/100%, theirs
  reads client 415140, whose value isn't traced. These three rows were Enhancement 119, 45 and 29
  for theirs before.
- Maelstrom Weapon: Windfury Totem's strike gives no stack, while white swings, Stormstrike and
  Windfury Weapon's strikes do. Ours since 2026-11-22 under Forever (`applyMaelstromWeapon` in
  `sim/shaman/talents.go` skips a main-hand swing that `IsWindfuryTotemAttack`), from their
  434723c516 (EV-0009: 0 of 140 landed totem strikes at ranks 1 and 2, 0 of 121 at rank 3,
  2026-11-13 and 14 after push 114060, against 0.59 a landed white swing; a Warcraft Logs census
  after the Nov 14 hotfix agrees). Our notes.md 2026-11-09 audit (428 of 430 Windfury extra swings
  gave a stack) predates the strike rework. Before, Enhancement 9.0 for ours.
- Faerie Fire in Cat Form: 1.0 s GCD, 6 s cooldown.
- Ferocious Bite on a miss, dodge or parry: 80% of its 35 Energy back, extra Energy and combo
  points kept. Ours since 2026-11-15 (their 0384d8f98c, 70862 log).
- Sharpened Claws: 3% a point to melee and spell crit in the feral forms (aura 290). Ours since
  2026-11-15 (about 0 DPS).
- Seal Fate: 500 ms ICD in both.
- Shank is an off-hand special: poisons, Dual Wield Specialization and procs reach it. Ours since
  6de6861c5 (`sim/rogue/shank.go:58`), theirs since bbf77b70d3 (ProcMaskMeleeOHSpecial); worth 13.2
  on Combat Daggers, Combat Swords 7.7 and Seal Fate 4.5 for ours before.
- An avoided Execute keeps its extra Rage and refunds none of its base cost. Ours since 2026-11-22
  (aa394d3be; their 7025f28b98: 36 avoided Executes in 1.60.1.70862 logs, 46.9 Rage to 37.9 after a
  parry and a Bloodrage tick at a base cost of 10). From 2026-11-17 ours refunded 80% (Arms 0.7).
- Deep Wounds sizes its bleed from the critting hand's average weapon damage (the off hand without
  its penalty), no attacker damage bonus (Death Wish, Enrage) scales the ticks, and a Thunder Clap
  crit starts none. Ours since 2026-11-22 (990cade4e; their 3af55b0ab8, EV-0139, a level 60 Fury
  log). Before, ours followed their older 3281b13bbe and 36cfc58328: for ours Fury 2H 7.7 to 8.5,
  Arms 4.6, Fury 3.5 (2.9 the bonuses, 0.6 the off hand).
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
  2026-11-17 (their e370277695, f64fef4df4). When it reads attack power and crit is a row (Rogue
  table).
- Rake's direct hit: 6.5% AP at every rank, read at the hit. Ours since 2026-11-12, outside the Nov 2 scale (their 0384d8f98c, 70862 log).
- Dragonfire and Caustic Smog Potions: neither hit nor ticks crit.
- Venom's +20% proc chance on Instant, Deadly and Wound Poison (ours since 2026-11-12).
- Arcane Concentration split across an Arcane Missiles channel; Wrack and Soul Siphon (hotfix
  113766).
- Searing Totem's damage per hit and rate within 3%.
- Judgement of Command is a melee special on the main hand's table, never dodged, parried or
  blocked; Righteousness and Fury are ranged since the Nov 11 hotfix. Ours since 2026-11-15 (their
  `seal_of_command.go:36`, client SpellCategories; their post-hotfix log, 1.062 n=109 under Two-Handed
  Weapon Specialization).
- Two-Handed Weapon Specialization raises all Holy damage with a two-hander, through the Holy school
  (Blizzard's Nov 19 hotfix, "now correctly increases all Holy damage", t/2403040 post 13). Ours
  since 2026-11-22 (`sim/paladin/talents.go` `applyWeaponSpecialization`), theirs since 3c82b227b2
  (2026-11-20). Hits are equal per hit, and theirs' 1.7 to 3.9% bigger Holy hits (Consecration,
  10333, 20286) are gone.
- DoT and channel ticks and haste: cast speed shortens the ticks of Hurricane and Volley only,
  rounded to the ms and read at the cast (a channel through ApplyCastSpeedForSpell). Hurricane keeps
  its 10 s and gains the ticks that fit; Volley keeps its 6 waves and ends sooner, and Auto Shot waits
  out the shorter channel. Every other DoT and channel ticks at its base length in both. Ours since
  2026-11-18, to verify against a log (user ruling; their `dot.go` CalcTickPeriod, `hurricane.go:41`,
  `volley.go:38-39`, b83fe9df5f).
- Wrack's tick timing: while Wrack is channelled, Corruption, Siphon Life and Bane of Agony tick the
  half points of their own grid (set by when they landed) as well as their normal ticks, from the
  channel's start to its end, both included; the channel moves no tick, and the dot ends on a full
  tick at its expiry with nothing dropped or cut short. Ours since 2026-11-22 (`sim/core/dot.go`
  SetTickPeriod, `sim/warlock/talents.go` applyWrack; their 6de4dee7b9, dc66ebf6f3, 999c93cb77 from
  EV-0085's logs). Before, ours scaled the time to the next tick and dropped the last tick: Affliction
  +1.8 to +24 (mean +10.4) on its golden file. Still unmeasured in both: a channel ending on a half
  point, and a dot applied mid-channel.
