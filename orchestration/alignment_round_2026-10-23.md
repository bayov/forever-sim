# Alignment round with wowsims/forever (2026-10-23)

A record for later sessions and for the class workers who regear and re-search rotations next. Branch
`claude/keen-brown-o2g32n`, commits `27f916a179..HEAD` (`git log --format='%h %s' 27f916a179..HEAD`). Each commit
body has the evidence (wowsims/forever hashes, client rows) and the golden and preset deltas. The comparison that
started it is in the orchestrator's scratch (`forever_sim_vs_wowsims_forever.md`, `audit2_findings.md`).

## What changed in the sim, by area

Everything is under the Forever ruleset; Classic keeps its behaviour.

**Core combat**
- 338eaf8fc1: no melee or ranged hit suppression (their 3af105fe16: 4,259 swings at 1% hit missed 7.04%). The caps are
  8% yellow and 27% dual-wield white; the physical presets built to 9% sit 1% over.
- b3f15c1d13: Rage. No Rage on misses, dodges and parries; damage taken gives 2.5 × (1 − A) × D / c(attacker level)
  (9.2 Rage for a 1500 boss hit, was 3.4); a two-hand crit adds the one-hand rate. k's level scaling is unchanged.
- e04953fb98: spell resists in 10% steps around R / (R + 9L), Spell Penetration adding damage, no level term.
- 216d430cbf, c4ee6d3e9d: any cast with a cast time resets the swing when it completes (Do Not Reset Combat Timers
  spells, channels and Auto Shot keep it); a cast at its end completes before anything else at that instant.
- 90f6f97192: an extra attack swings through a resetting cast; a missed ability refunds 80% of its Rage.
- d9549de245: pets work their inherited stats out from the owner's totals and take no consumables.
- 0354001610: caster and Moonkin Faerie Fire keep the swing; Hurricane's ticks roll hit and never crit.

**Items, sets, consumables**
- da4f2018a7: item procs and imbues on their Forever rows; Elemental Sharpening Stone crits only the stoned hand;
  Nightfall is a real 15% magic damage taken.
- d86436d1d7: the dungeon, crafted and PvP sets Forever reworked, counted by client set ID.
- f8557de17a: natural flasks (zone bonus), explosives and damage potions anyone can use, one crit elixir's crit,
  Thunderstrike and the Stormhammer fixed.
- 1bd41fea27, b0f368e6ed: the Level 60 presets fight in a natural flask zone and take the best flask (natural where it
  beats Supreme Power); Arms has its own consumes; Retribution the Dragonfire Potion, Elemental the Caustic Smog Potion.
- e5cbda3cc6: new item effects with a stated rate, Read Leyline castable from an APL, Felheart 8 piece a cost factor.
- 1cc6e540e2, c43c22730c: wand damage school from ItemSparse; wand Shoot and Wand Specialization; Smite, Shadow and
  Frostfire wand when out of mana.

**Warrior**: 26c0101fa1 (Deep Wounds, Bloodthrill procs, live Rend, Sweeping Strikes 20 sec, Charge before the pull in
every Level 60 rotation; Arms −4.2%), 7afa40fdba (Protection rows), c54f103bfb (Rage threat, Cleave and Sweeping
Strikes copies, Shield Block refill, Shield Slam from the shield).

**Paladin**: 9b7228699b (judgements on the melee table, Sanctified Judgement, one-hand Seal of Righteousness, Holy
Power's mask, Consecration's extra to four enemies, Champion of the Light 20/40/60; presets +47 to +66), 8e7a7c9639
(Tenets of the Silver Hand, Libram of Invocation, Judgement of the Crusader gloves, Justice sets by ID, one Twist of
Light Echo), a7f6cf12ac (Tenets in the vs Undead presets).

**Hunter**: 972736c67b (pet creature weapon and 254 attack power, owner's crit, 17/8 spell hit, client ability rows,
Lone Wolf divides the pet, Hunter's Mark 143), 9a508cae47 (patch 0004 pet hunks), e5739fed31 (Summon Hawk guardian),
22eb80d0a4 (Pet Aggression, Cobra Reflexes on, passive swing rows, run-in, Screech on every enemy, Explosive Trap
attack power, Swiftfeather Arrow in the presets), 1ca4850e1e (every family ability, the Fox, their pet AI, Furious
Howl), e0227b6824 (Auto Shot windup, movement, deadzone, Strider Kick speed, hunter_ranged weave knobs), 800b1a3530
(ranged presets melee weave; turret rotations kept as bm_turret, mm_turret, lw_turret; Survival in melee).

**Rogue**: d0dd3a58a9 (Mutilate needs two daggers, poisons get attack power, missed finishers refund, Blade Flurry copy,
client spell ids), 1ff61694dc (live Rupture and Garrote, client Mutilate id, Grimstitch by set ID, Ambush behind),
84523db102 (Night Elf Shadowmeld into Ambush, shared with the druid).

**Priest**: 5d38d2ca10 (cost cuts multiply, Shadow Weaving live on the target, no Chastise in Shadowform), f8317e9ab7
(Troll Shadowguard in the level60 rotation, Penance and Vampiric Embrace on-cast procs, Penance travel), c43c22730c
(wand).

**Shaman**: 99e7e3aa9f (Fire Nova's hits a proc that item procs don't hear, Frost Shock binary; Enhancement −0.2 to
−0.7%).

**Mage**: 24af9d158f (the 2026-10-20 Frost hotfix: Piercing Ice out, Elemental Precision, Ice Barrier and Frigid Power;
talents re-placed), 91387ae545 (percent modifiers multiply, Frost Channeling on Frostfire Bolt, Fingers of Frost spent a
batch late, Master of Elements), 8a2a950446 (gen_db's Frost tree), 6bfbc6d1bd (Cone of Cold, Winter's Chill, Improved
Cone of Cold, Ice Lance not binary), c43c22730c (wand).

**Warlock**: a3eab2e207 (demons on their measured shares, melee and regen; −3.7 to −4.2%), 9d8d888a8d (percent rows
multiply, Spellstone Fire, Demonic Brand, Conflagrate, Decimation, Rain of Fire), 15d8d1bd16 (Demonic Brand never
misses, summon cost cuts multiply, Imp Firebolt travel, demon stunned until a summon ends), e5cbda3cc6 (Felheart 8).

**Druid**: f2e59548b4 (Nature's Grace a batch late, caster Faerie Fire rolls Omen; Balance −10), 25747a3950 (Rake
attack power, Omen at the swing speed, Rend and Tear on ticks, Rip refund, enchant and stone on the cat swing),
20e6ce9f0f (Natural Instinct on Rip, live Rip, failed Bite spends all Energy, Feralheart only when struck, Idol of
the Dream, Talons of Wrath, Mystic Mushroom, Haruspex's Garb, prepull Clearcasting option, real Excess Energy),
84523db102 (Ravage, Prowl, Shadowmeld into Ravage in the Level 60 Feral APLs).

**Report inputs (Phase A, this file's commit)**: assumptions resolved or added (below), patches 0002 and 0004 rebased
onto this branch, `run_assumptions.py` edits that set several request fields, the seed check.

## User decisions (2026-10-23)

- Earlier rulings that a fix superseded are fine: they were assumptions.
- Natural flasks: modelled, presets on, the fight in Hyjal Summit or the Barrow Deeps, best natural flask in the presets.
- Cone of Cold: modelled; the Frost preset stays at range; the melee-range variant is a report assumption.
- Cobra Reflexes: hits shrink with the speed, like `hunter_pet_speed`.
- Swiftfeather Arrow: modelled, presets on.
- Goblin Land Mine resets the swing: a single-value assumption.
- Next: fix gear (drop the 1% hit over the cap and reallocate), resolve assumptions from their logs, tank stone off the
  shield, rotations and talents, regenerate the report.

## Report assumptions after Phase A

`tools/dps_report/assumptions_def.py` has 41 rows, none pending; details in `orchestration/spec_assumptions.md`.

- **Resolved to one value** (row kept for reference, knob out of patch 0004 unless noted):
  - `lock_pet_scaling`: measured shares (wowsims/forever c3d9ea9516, 9ce48b6aba, 19922b122e, 7097a63bc9, f5b2f7ef33).
  - `lock_demon_melee`: the Succubus 66 to 103 (b207517048, level 20 and 30 sheets).
  - `hunter_pet_scaling`: 10% attack power read on a level 17 pet sheet, hit from Blizzard (643d636345, 9ce48b6aba).
  - `hunter_trap_ap`: 10% of ranged attack power over a trap (3993d79540, their reading of Blizzard's notes, no log;
    the knob stays in 0004).
  - `pal_2h_spec_holy`: the specializations raise the melee Holy spells in game (d35afca4da, followed in 9b7228699b).
  - `warr_dt_rage_mods`: none of our modifiers (ee2d69b16a, 73b317c7cf: the hit before mitigation, geared and naked).
  - Already one value, rechecked: `level_partial_resist` (ed4b4e93e5, e04953fb98), `druid_ooc_rate` (ca7b3ab462,
    af6df6c0e4; its knob is out of 0004).
- **Kept open, judgement call:** `warr_dt_rage_k` keeps its 3.0 High. Their logs fit melee hits only, and their own TODO
  has spells landing at 1 to 2 times the rate (Fireball 1.2); the report's raid hits are Fire spells.
- **New:** `mage_cone_of_cold` (wired: High sets the APL, wowsims/forever's talents and 10 yards; +10.9%),
  `hunter_cobra_damage` (High: wowsims/forever's TBC 15% cut, +1.4% Beast Mastery), `natural_flask_zone` (Low: out of
  the zone, −3 to −6% on the melee, hunter and cat builds), `land_mine_swing` (one value).
- **Left out as below the threshold:** Ferocious Bite on a miss (0.06%), the one Twist of Light Echo (duty presets
  −0.3%). Shadow Weaving is not an assumption (5d38d2ca10 put it on the target as theirs).
- **Still open, unchanged:** `warr_rage_k` (level scaling), `warr_rend_ap` (measured on rank 2 only), `hunter_pet_speed`,
  `hunter_tsa_hm_rank` (Trueshot Aura only), `hunter_crit_per_agi`, `hunter_raptor_31plus`, `mage_ice_lance_coef`,
  `mage_frostbolt10`, `mage_imbue_scrolls`, `priest_sf_tick_crit`, `priest_smite8_tuning`, `lock_bod_coef`,
  `lock_stone_with_oil`, `pal_31plus_ranks`, `pal_soc_wf`, `pal_soc_ppm`, `sham_mw_ppm`, `sham_lb_31plus`,
  `sham_es7_tuning`, `druid_rip_31plus`, `druid_wrath8_tuning`, `bleed_tick_crit`, `raid_hits`.

Patch 0004's `pal_soc_wf` knob did nothing from 2026-10-16 (its check in `CreateExtraAttackAuraCommon` was lost in
c3c8167f20's rebuild); it is back, and its Low measures −1.4% on twist. Patches 0001 to 0004 apply in order on this
branch, the patched `wowsimcli` builds, and with `DASH_ASSUME` unset it is bit-identical to the plain one on every
non-duty request (`run_assumptions.py --identical`, 300 iterations). Patch 0002 gains the parts `shadowmeld` and
`leyline` (run_races.py and build_races.py list them).

## Seeds

The CLI honours `simOptions.randomSeed`, and iteration i runs on seed + i (sim/core/sim.go; the concurrent split
starts each part where the previous one ends). Seeds 1 and 2 at N iterations share N − 1 iterations, so their averages
are nearly identical by construction (twist paladin, 2,000 iterations: 935.18, 935.14, 935.19 on seeds 1, 2, 3; 935.01
on 1000003). A second seed set must start more than the iteration count away: keep 1 and 1000003
(orchestration/handoff_orchestrator.md says so now).

## Still open (needs a log or a ruling)

- From the audit's open list, not to fix without new evidence: the 1% spell miss floor; extra attacks and the swing
  timer; Heroic Strike and Cleave proc masks and Flurry; Eviscerate's attack power share; Arcane Missiles proccing
  Missile Barrage and Clearcasting; hunter crit per Agility; Earth Fury's mace gate; Inner Focus on channels; Eclipse
  charge timing; Anger Management; Amplify Curse on Bane of Agony; Lightning Overload details; Maelstrom Weapon's rate;
  Lacerating Strikes' refresh; Mind Flay binary; totems reaching pets.
- Raised this round: race stat offsets (their 6fc22306b1 reverted to 1.12; needs a sheet); the paid-channel regen hold;
  whether the Land Mine has a cast at all (client 4100 is instant); the Cobra Reflexes cut; the pet AI's ability order
  and Charge's speed; Training-point budgets (Pet Aggression 5 with Cobra Reflexes); demon melee at 60 (read at 20 and
  30 only); the pet's hit share on a sheet; Seal of Command on white hits only (theirs) or every melee hit (ours);
  Totem of the Storm's Forever rate; Flurry spent on a miss; Stormstrike's charge timing; which explosives and
  elixirs stack (notes.md).

## Per-class follow-ups for the regear and rotation round

Every class: regenerate the `tools/dps_report/req` requests from the presets (many still carry pre-round APLs, flasks,
no natural flask zone and old pet options), then the finals at 120/180/300 s on seeds 1 and 1000003, the race sweep,
presets.ts, goldens and the wiki. Physical builds: drop the 1% hit over the new 8% / 27% caps and reallocate it by
standards 1b.

- **Warrior.** Arms lost 4.2% to the Deep Wounds and Bloodthrill fixes: re-search its rotation, including the Hamstring
  filler (Bloodthrill no longer procs from Hamstring, 0 in 99 in their log), Slam and Heroic Strike thresholds with the
  Charge opener. Tank warrior: the stone goes on the main hand, not the shield (Elemental Sharpening Stone crits only
  its own hand since da4f2018a7). Check Fury's Precision against Aggression after the hit change.
- **Paladin.** Give the Shock builds (and Shock vs Undead) their own consumes: with the shared Dragonfire Potion they
  lose 12.8 and 18.5 to the Major Mana Potion. Try Libram of Invocation (−5% seal cost, the audit guessed +1 to +3 on twist) against the
  current relics. The Crusader duty rotations lose 2.7 to 3.0 to the one Echo: check their seal swap.
- **Hunter.** The weave rotations, Cobra Reflexes and Swiftfeather Arrows are in the presets: re-search talents and the
  weave knobs with the hit change; try Slower Attack III and the raptor's Savage Rend (+0.25% over the cat); the hawk
  build is about even with the preset.
- **Rogue.** The Mutilate build needs two daggers (its sets changed in d0dd3a58a9); the sets sit 1% over the cap with
  Precision 3/3 and Grimstitch's 2 piece. The Night Elf now leads Subtlety (Shadowmeld into Ambush, 2.7 ahead on both
  seed sets): race pass.
- **Priest.** The rotopt templates and request files carry the old APLs: regenerate with the wand lines (Smite, Shadow)
  and the Troll's Shadowguard. The Troll is now the best Horde Shadow race; Human stays first.
- **Shaman.** Elemental Mana Tide: give it the Major Mana Potion (Caustic Smog is 1.8 behind there). Enhancement lost
  0.2 to 0.7% to Fire Nova's proc fix: recheck the totem lines.
- **Mage.** Frost talents: wowsims/forever's string at range (832.3) beats the preset's (820.7); re-search. Rotation:
  Frostbolt then Ice Lance on one Fingers of Frost charge (both read Frozen, the spend comes a batch late): +1 to +3%
  for Frost and Frostfire. Frostfire's rotopt template needs the wand line. A High Order's Read Leyline costs 17.4
  cast on cooldown; only a mana condition is free. `mage_cone_of_cold` stays a report row, not the preset.
- **Warlock.** The demons lost about 4%: Destruction fire's old Succubus sacrifice build
  (`252-0050003201-0050345003101351`) is now 6.4 ahead of the preset (984.8); re-search the sacrifice choices.
- **Druid.** Feral: turn on prepull Clearcasting (+0.9, +2.4 with the idol), take Idol of the Dream (+3.1 over Howling
  Idol), and add Rake maintenance ("Rake if combo points < 5, Rake down and more than 6 sec left", above Shred: 771.7 to
  804.2, +4.2%). The Night Elf is now the best Feral race (3.4 ahead of the Tauren on both seed sets): race pass.
  Balance lost 10 to Nature's Grace's batch delay; Talons of Wrath and Mystic Mushroom are 3.3 to 3.7 behind the
  Swarming Idol.
