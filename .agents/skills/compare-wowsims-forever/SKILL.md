---
name: compare-wowsims-forever
description: Compare our sim with wowsims/forever (github.com/wowsims/forever, the other WoW Forever sim) on one of our builds - the same character, gear, buffs, boss and rotation run in both, the character sheets matched, and every DPS gap traced to a rule the two sims model differently, then sorted into our bugs, their bugs and open questions. Use when the user asks to compare a class, spec, preset or mechanic with wowsims/forever ("compare hunters next", "what does wowsims/forever give for Combat Swords", "why do they get more"), to check a mechanic against theirs, or after wowsims/forever changed a class we sim.
---

# Compare with wowsims/forever

wowsims/forever is the other WoW Forever sim, built on the game files like ours. The comparison runs one of our
requests in both and explains the gap. It is a source of questions, never of answers: a rule only becomes ours on
the evidence behind it (game data, a log, a post), as with any other source.

**Never push to wowsims/forever, comment on it or open anything there.** Its checkout is a reference. Our tools in
it are kept out of its git (`.git/info/exclude`). If a finding looks like their bug, tell the user; reporting it
upstream is the user's call.

The tools are in `tools/wsf_compare/`:

| File | What it does |
|---|---|
| `setup.sh` | Clones wowsims/forever (or moves it to its latest master), adds our two tools to it, dumps their builds and APLs, and builds every binary below. It prints both commits. |
| `convert.py` | One of our requests → `req_ours.json` (ours, aligned), `req_theirs.json` (the same for theirs) and `items.json` (our synthetic gear as their items). |
| `run.py` | Converts, runs both, prints `ours`, `theirs` and `theirs - ours` with its 2σ error. |
| `align.py` | Finds the bonus stats that make their sheet at the pull match ours. Prints the `ALIGN` value for `run.py`. |
| `breakdown.py` | Per ability, side by side: DPS, casts, hits, crits, misses and dodges per iteration, for the player and each pet. |
| `layers.py` | Both sheets built up one layer at a time (naked, + talents, + gear, + consumables, + buffs), so a stat gap shows which layer brings it. |
| `ours_stats/` | Prints our sheet at the pull. Their counterpart, and their runner `cmpsim`, are in `theirs/`, kept as `.txt` so our build doesn't compile them. |

`$WSF` is the checkout (default `~/wowsims-forever`) and `$WSF_WORK` the work folder (default `~/wsf_compare`),
with the binaries in `$WSF_WORK/bin`.

## 1. Set up

```bash
tools/wsf_compare/setup.sh
```

About 3 min the first time, under a minute after that. Rerun it whenever either sim changes, since every binary is
built from a commit. It needs protoc (`/root/tools/protoc/bin`, as in `tools/dps_report/README.md`). Note both
commits it prints: they go in the report. If you patched their checkout to flip a rule (step 4), take the patch out
first (`git -C $WSF diff > $WSF_WORK/<name>.patch && git -C $WSF checkout -- .`), or the update stops.

## 2. Run the builds aligned

Take the builds from the DPS Report's requests (`tools/dps_report/req/<id>.json`), and every variant that differs in
the part being compared (pets, weapons, talents). Screen at 3,000 iterations. Run the final table at 20,000
(standards).

```bash
cd tools/wsf_compare
A=$(./align.py hunter__survival ~/wsf_compare/out/sv)
ALIGN=$A ./run.py hunter__survival ~/wsf_compare/out/sv --iters 3000
```

- **Align the sheets first.** `align.py` closes the stat gaps at the pull (Strength and Agility before Kings, then
  attack power; Intellect, Spirit and spell power for casters). A gap it closes is a difference too: find out what
  it is with `layers.py` (race bases, buff ranks, how a rating converts) and list it.
- **The rotation is ours in both** (`--apl ours`, the default), translated where theirs lacks a value. That
  compares mechanics, not rotations.
- **Ranged hunters:** compare turret (`--turret` drops our `Weave:` lines). Their weave walks differently, which
  hides everything else. Melee specs run as they are.
- `--apl theirs:<name>` runs their own rotation (`$WSF_WORK/theirs_builds`, `<class>_<spec>_<build>` or
  `apl_<class>_<spec>_<apl>`). On our gear their rotations came out 72 to 174 DPS behind ours, out of rotation and
  pull misfit, so it says little about mechanics. Use it only for a question about their rotation.

## 3. Read the gap

1. **Player or pet.** `run.py` prints the pets apart. A gap that is all pet is a pet rule.
2. **Which ability.** `breakdown.py <out dir>`. The two sims label some rows differently: ours shows the main
   hand, off hand and extra attacks as `OtherActionAttack` tags 1, 2 and 3, theirs as tags 14 and 15 and the extra
   attack's own spell (10610 for Windfury, 13960 for Sword Specialization). Look at casts first (the rotation or a resource), then hits and
   crits per cast (the hit table), then damage per hit (scaling, coefficients, ranks).
3. **Which stat.** `layers.py <out dir>` when the sheets disagree before alignment.
4. **Read both codes** for that ability or rule: ours in `sim/`, theirs in `$WSF/sim/` (their spell data is
   generated from the game files, `sim/core/spelldata`, `sim/<class>/spell_data_auto_gen.go`). Their commit
   messages often give the evidence: `git -C $WSF log -S <name>`.

## 4. Isolate each cause

Flip one rule on one side and rerun. The DPS it moves is that rule's worth. Do it at 3,000 iterations, and at
20,000 for anything under 3 DPS.

- **On our side:** edit `req_ours.json` (options, consumables, buffs), or flip the rule in a build of our sim from a
  scratch worktree, never the checkout's working tree.
- **On their side:** edit `req_theirs.json`, or patch their checkout (uncommitted, reverted afterwards, step 1).
- The causes should add up to the gap, within the error. When they don't, there is another cause.

## 5. Sort the causes

For each cause, say which sim is right, on what evidence:

- **Our bug** (theirs has the game data or a log, ours a Classic or Era carryover, or a plain mistake): fix it as
  normal work, with its test, then re-sim the report builds it moves (`regenerate-dps-report`).
- **Their bug:** note it for the user, with the evidence. Nothing in their checkout changes.
- **An open question** (neither side has evidence, or the evidence conflicts): name the report assumption
  (`orchestration/spec_assumptions.md`) or the ruling ours follows, and what would settle it (a log, a sheet). Add
  it to `notes.md`'s To Verify list if it isn't there. The user decides.
- **A choice, not a game fact** (engine timing, the generic boss): keep ours and say so.

Our side holds rulings the user made (`orchestration/`, `notes.md`, the code comments citing them). Don't propose
reversing one without new evidence, and say which ruling it is when theirs differs.

## 6. Report

- A table per build: ours, theirs, `theirs - ours` with its error, at 20,000 iterations, aligned.
- The causes with their DPS, sorted as in step 5.
- Both commits, the iterations, and any alignment it took (`ALIGN`, `--their-mob-type`).
- Ask before acting on a judgement call. Bugs in ours with clear evidence follow the repo's rules (`CLAUDE.md`):
  fix, merge, re-sim, deploy.

## What the converter translates

Each line below cost a wrong result once. The converter does all of them. Check this list when a new class or
rotation gives a result that looks off.

- **Synthetic gear.** Our items 990000 and up don't exist in theirs. `items.json` adds them, and `cmpsim` and
  `cmpstats` load it. `cmpsim` also registers what our synthetic gear does beyond stats, which `items.json` can't
  say: the on-use trinkets (994202, 993016, 993117, 991018, and the shamans' 990015 and 990115) and the hunter, warrior
  and Enhancement sets' 2 piece. A new synthetic effect goes there
  (`theirs/cmpsim/main.go.txt`).
- **A set bonus theirs doesn't register** goes in as bonus stats (`SET_BONUS`): Grimstitch Armor's 1% hit as 10
  hit rating, and the Elemental set's The Spiritcaller's Storm the same.
- **Enchants:** ours keys Forever's spell ids, theirs the enchant ids. The converter matches them by effect id, else
  by name and stats; `ENCH` holds the ones picked by hand. Unmapped, theirs drops the enchant without a word (it cost
  Elemental 105 spell power).
- **Stats:** theirs names the spell damage stats `...Damage` where ours has `...Power`, and its `SpellDamage` is our
  `SpellPower` and our `SpellDamage` (damage only) added up. `layers.py` and `align.py` compare them that way.
- **Races:** theirs spells the Skyborne the other way round (`RaceWindshaperSkyborne`).
- **Consumables:** our enum names → their item ids (`C`). A potion is their `resourcePotId` (mana or rage) or
  `offensivePotId`, and our rotation's `OtherActionPotion` is their `OtherActionResourcePotion` or
  `OtherActionOffensivePotion` to match. Theirs applies one explosive of a kind, so ours drops its filler by default (`--keep-filler`
  keeps it).
- **Stones:** theirs refuses a sharpening stone on a blunt weapon, so the converter swaps in a weightstone, the
  same +8.
- **Buffs:** since their 04899cee95 the group buffs are plain on/off under the single-target names (`arcaneIntellect`,
  `markOfTheWild`, `powerWordFortitude`, `shadowProtection`), and since b6657d3824 a ranked buff is an int, 1 for the
  trainer's rank and 2 for the book's (`strengthOfEarthTotem`, `graceOfAirTotem`, `greaterBlessingOfMight`,
  `greaterBlessingOfWisdom`). Our raid Flametongue Totem is their `flametongueTotem`. Mana Spring Totem is a party
  buff in theirs. `battleShout` is an int there (1 is rank 6, 2 rank 7).
  Their Battle Shout lands just after the pull's sheet, so `align.py` adds its 115 attack power to their sheet.
  Our Improved Battle Shout, Blessing of Might and Gift of the Wild come out the same as their plain ones.
- **Sunder Armor:** theirs ramps the raid's copy itself, a stack every GCD from the pull (5 at 6 s), and our report
  ramp also reaches 5 at 6 s, so ours keeps its ramp. (Full in ours against their ramp was worth about 4 to the hunter
  and 1.5 to the pet.)
- **Rotation values theirs lacks:** `spellUsesRemaining` and `spellUsesLostByDelay` (rewritten from
  `remainingTime`, with the cooldowns in `CD`; `spellUsesRemaining <= k` too), `moveTime` and `inMeleeRange` (from `unitDistance` and
  `unitMovementSpeed`). `autoTimeToNext` and `autoTimeSinceLast` take `MainHandAuto`, `OffHandAuto` and
  `RangedAuto` (`autoSwingTime` keeps `MainHand`).
- **A talent spell the build lacks** (`NOSPELLS=<id>,...`): ours reads a value naming it as false, theirs drops it
  from its `and`, which turns the `and` true.
- **Spell ids and ranks:** theirs registers only some ranks. Rogue: Backstab VIII, Eviscerate VIII, Deadly Poison IV,
  Vanish rank 2, Mutilate's last rank, Berserking 20554. Hunter: only the top ranks, the AQ books' (Aspect of the
  Hawk VII, Serpent Sting IX), so the hunter runs use `ours_aq_cli` with our rotation's ids moved up (`OURS_AQ=0`
  runs ours on its own ranks: the difference is the AQ ranks' worth). Raptor Strike's queue is tag 3 in ours, tag 1
  in theirs. Warrior: Whirlwind is cast as its main hand's copy, tag 14 (untagged it never casts: about 100 DPS on
  Fury), Charge has no rank, Berserking 20554. Shaman: Berserking 20554; theirs registers every rank our rotations cast.
- **Vanish and Sweeping Strikes** are DPS cooldowns in theirs, so `autocastOtherCooldowns` would Vanish at the pull
  and spend 30 rage on Sweeping Strikes every 30 s against one target (their own rotations cast it only on 2+
  targets and don't autocast: 32 DPS on Arms). The converter adds a never-true line naming each.
- **Charge:** ours is an opener for its 15 rage at any range; theirs moves the warrior in and needs 8 yd. A rotation
  that opens with it starts theirs at 20 yd. Their Charge also holds the off-GCD items after the pull's first GCD to
  3 s (Blood Fury, the potion and the trinket at 3.0 s, not 0.0 s); that cost nothing measurable on Fury.
- **Raid damage taken** (our encounter's `raidDamageHitsPerMinute`): theirs has no such model. `--no-raid-damage`
  turns ours off to compare the rest. For shamans the converter sets their `shieldProcrate` to the same hits a
  minute, which spend Lightning Shield's and Water Shield's charges as ours does.
- **Totems:** our rotations ask whether a totem stands by its buff on the shaman (`auraIsActive`) or as a DoT for a
  fire totem (`dotIsActive`). Theirs keys neither by the totem's spell, so the converter asks its element's slot
  (`totemRemainingTime`). Untranslated, theirs recast Strength of Earth 50 times a fight and ran out of mana.
- **Their sheet leaves out talent crit** they apply per attack (Cruelty, Weaponmaster's axe crit), so `layers.py`
  shows ours 5 to 10% crit higher on warriors. Compare crits per hit in `breakdown.py` instead.
- **Hunter options:** the quiver and pet speed enums, the Crocolisk spelling, and their Beast Mastery tree with
  Pathfinding and Improved Revive Pet swapped (5th and 6th).
- **Off-GCD timing:** their engine fires an off-GCD item only at the next GCD check. A rotation that uses one in
  between (Thistle Tea at 10 energy) loses it there: compare that with care.

## Known differences

As of 2026-11-02, on the report's requests: the hunter and rogue rows with ours at 0e87f3d and theirs at b70888a518, the
warrior rows with ours at 7587a1a and theirs at 895cd67f99, the shaman rows with ours at dfa3b4a and theirs at 1496ed4db5. The four pet rows (two since settled) add up to the BM cat's whole
pet gap (334 against 247 at 20,000 iterations, measured by flipping each in ours). The warrior rows add up to Arms' whole
gap and leave 4 DPS of Fury's and Fury 2H's unexplained. The shaman rows add up to Enhancement's whole gap (33 to 37 DPS
on the four presets) within about 2 DPS and Elemental's (5 to 6) within about 1. Engine timing (when each engine re-checks the rotation, Heroic Strike
1.4 casts more in ours on Arms) moves casts but nets about 0 DPS once the Windfury rule below is matched. A rerun should find
these, so it can tell what is new. Keep the list current when one is settled.

| Difference | Ours | Theirs | Status |
|---|---|---|---|
| Pet swing damage | (creature DPS − the pet's own attack power / 14) × 2 a 2.0 s swing since 2026-11-01 (user ruling) | The same (71d4f46012) | Settled on their level 30 cat log; level 60 unmeasured (notes.md To Verify). It was −48 pet DPS on the BM cat |
| Pet Aggression | Dropped 2026-11-01 (user ruling) | Dropped (8aa26ccb4d) | Settled: no trainer teaches it (its teaching spells are on no skill line, Beast Training's has none). It was −34 pet DPS on the BM cat |
| Improved Tracking vs the generic boss | Counts the typeless boss as tracked (standards 19) | Only the types a Track spell names (8aa26ccb4d) | A choice. `--their-mob-type MobTypeGiant` aligns it (Survival 41 DPS) |
| Pet focus | 10 a second (client RegenCombat, beta logs) | Classic's 25 per 5 sec × Bestial Discipline | Ours on evidence. −28 pet DPS on the BM cat (half the Claws) |
| Cobra Reflexes | Hits shrink with the speed (`hunter_pet_speed`, user ruling) | ×0.85 a hit (TBC) | Ruling. +24 pet DPS for theirs on the BM cat |
| The hunter's buffs on the pet | Passed live | Not | Wolf and Tallstrider 0 to 10 DPS with Cobra Reflexes |
| Raptor Strike rank 8 | +96 (`hunter_raptor_31plus`) | +70 | Ruling |
| AQ book ranks | Off under Forever | Hunter's top ranks only | Worth 30 to 36 DPS to ranged hunters |
| Explosive reach | 10 yd, since 2026-11-01 | 10 yd | Settled (ours was a bug) |
| Windfury Totem on a landed hit that deals no damage (Rend's application, Sunder Armor, Rupture's and Garrote's) | Rolls it (any landed main-hand hit, `sim/core/buffs.go` `CreateExtraAttackAuraCommon`) | Doesn't (`RequireDamageDealt`, the decoder's default, with a TODO to check in game; their own weapon-proc rule says the opposite) | Report assumption `wf_totem_nodamage` (one value, ours: Sunder Armor procs it in Classic; user ruling 2026-11-02). Arms 11.9 DPS (all of its gap: about 6 Rend applications a fight), Subtlety 2.5 |
| Windfury Totem ICD, extra attacks and the swing, Eviscerate's attack power | Theirs' rules since 2026-10-31 | 100 ms, no swing reset, 4% a combo point | Settled, on `notes.md`'s To Verify list |
| Off-GCD items (Thistle Tea) | Fire when the rotation says | Only at the next GCD check | A choice (engine). With Charge in the opener theirs walks in and starts its GCDs at 0.21 s, which shifts the trinket windows (Arms Slam +0.7, white hits −0.9 in ours: they cancel). Ours drinks Thistle Tea at 10 energy, which theirs never does: about 13.5 DPS on Combat Swords. Bloodrage at under 20 rage: 2.6 casts against 1.9 on Fury, worth nothing |
| Raid damage taken | 2 hits a minute of 1,500 (standards 24) | None | A choice (the fight). The warrior's rage from it: Arms 8.5, Fury 15.7, Fury 2H 19.5 DPS. `--no-raid-damage` |
| Flurry charges | Two white swings within 500 ms spend one charge (the Classic sims') | Every white swing spends one, and a missed Heroic Strike or Cleave (0183e67d2f, no log cited) | Report assumption `flurry_batch` (one value, user ruling 2026-11-02). Fury 11.4, Fury 2H 3.4 DPS |
| Anger Management | 1 rage every 3 s (Forever's tooltip) | 1.5 every 5 s (the aura's amount, a power regen per 5 s) | Open: a log of its rage ticks settles it. Under 2.1 DPS |
| Unbridled Wrath on Heroic Strike | Procs (client proc flag 4, notes.md) | White swings only | Report assumption `warr_uw_queued` (one value, user ruling 2026-11-02). Within 1 DPS |
| Two-hand white crit rage | Twice the whole swing since 2026-11-01 | The same (7c28cfa731, Forever logs) | Settled (ours was a bug): it was −12.4 on Arms, −13 on Fury 2H |
| Rotation re-checks and Heroic Strike during a cast | Re-checks at once on a Rage gain; a queued Heroic Strike goes off on a swing that lands during Slam's cast | Re-checks 200 ms after a Rage gain or a swing; drops the queued Heroic Strike during a cast (`hardcastRuns`); holds major cooldowns while a sequence step runs | Choices (engine). They move Arms' Heroic Strike (+0.8 casts in ours) and Bloodrage, but net 0 DPS: theirs spends the Rage on Execute |
| Explosives' hit | 1% miss at most (an Era comment, `sim/core/consumes.go`) | The normal spell hit table | Open: the client row 13241 would say. Sapper +0.3 DPS |
| Touch of the Grave's roll | On any ability cast (a guess, `racials_forever.go`) | On a landed hit that deals damage | Open. +0.4 DPS on Undead Arms |
| Flametongue Totem beside Windfury Weapon (Enhancement) | Works (report assumption `sham_ftt_stack`, user ruling 2026-10-16) | Does nothing while any Windfury Totem stands, the shaman's own included, though its Windfury Weapon turns that totem off for it | Open, on notes.md's To Verify list. 16.2 DPS on Enhancement |
| Lightning Bolt rank 10 | Base damage ×1.12 (`sham_lb_31plus`, user ruling 2026-10-15) | The client's | Ruling. Elemental 14.9, Enhancement 2.1 DPS |
| Earth Shock rank 7 | Base damage ×1.08 (`sham_es7_tuning`, user ruling 2026-10-15) | The client's | Ruling. Enhancement 4.0 DPS |
| Stormstrike's weapon speed | The weapon type's normalized speed + 0.3 (3.6 for a two-hander), fitted to beta Stormstrikes (shaman_audit.md 4.1) | The normalized speed | Ours on evidence. 4.6 DPS on Enhancement |
| Searing Totem's attack rate | Every 2.5 s: 72 beta gaps averaged 2.53 s (the 2.2 s bolt and a pause) | Every 2.2 s, the bolt's cast time | Ours on evidence. +3.7 DPS for theirs on Elemental |
| Haste and the GCD | 1.5 s (`haste_gcd`) | Spell haste shortens it (their TBC code) | Report assumption. Elemental 2.6, Enhancement 0.9 DPS for theirs |
| Magma Totem's last pulse | A 40 s totem pulses 20 times | About 19, by the counts (84 pulses a fight against our 89) | Engine timing, no log either way. About 3 DPS on Enhancement |
| Greater Blessing of Might and Wisdom | Rank 2's 133 attack power and 40 MP5, a trainer rank (standards 8a) | Rank 1's 112 and 36, rank 2 counted as a book rank (b6657d3824, no item teaches it) | Ruling. `align.py` closes it (21 attack power, 4 MP5) |
| Blood Fury's weapon damage | +10% of the weapon's average damage beside the +10% attack power (beta sheets and logs, `racials_forever.go`) | Attack power only, as the client's three auras read | Ours on evidence (their bug, not reported). 2.9 DPS on Orc Enhancement |
| Crusader beside Windfury Weapon's strikes | Holy Strength lands at once, so the triggering hit's proc reaches both strikes | 10 ms later (no `TriggerImmediately`), after the strikes | Open, notes.md's "Server batch window". 0.6 DPS on Enhancement |
| Lightning Overload's copies | Procs since 2026-11-02 (Judgement of Wisdom, Elemental Devastation and Water Shield don't hear them) | The same | Settled (ours was a bug, the client's proc attributes). About 0.1 DPS on Elemental |
| Spell queue window on cooldowns | No spell queue: the rotation casts what is ready now | A spell within 400 ms of its own cooldown counts as ready and is queued, the rotation idling till then (`spell_queueing.go` `CanQueue`, wowsims' generic core, no Forever evidence) | Open (`docs/practice_room.md`'s spell queue check). Elemental casts 14.5 Lava Bursts against 13.7, but the idle costs theirs 0.8 DPS. It likely moves other cooldown spells in theirs too, unmeasured |
| Orc shaman base stats | 96 Stamina, 89 Intellect, 100 Spirit | 97, 87, 103 | Race offset ruling. `align.py` closes it |
| Synthetic sets' 2 piece (1% attack speed) | Live | Registered by `cmpsim` (warrior head 991001, 991019, Enhancement hands 990007) | A translation: without it theirs was 5 to 8 DPS low |
| Undead base stats | 82 Str, 129 Agi | 79, 128 | Race offset ruling. `align.py` closes it |
| Totem ranks | Trainer (Strength of Earth 42, Grace of Air 77) | AQ books' (53, 89) | No AQ books at launch (assumed) |

## Adding a class

`convert.py` maps rogues, hunters, warriors and shamans, and stops on any other class. For a new one:

1. Its options block (`class_opts`): their class proto (`$WSF/proto/<class>.proto`) against ours.
2. Its consumables in `C`: weapon oils, poisons, imbues, class-only items.
3. Its spell ids in `remap()`: which ranks theirs registers (their `spell_data_auto_gen.go`) against what our
   rotation names, and any tag that differs. Check every cast in `breakdown.py`: a spell with 0 casts in theirs is
   an id that didn't map.
4. Its cooldowns in `CD`, for `spellUsesLostByDelay`.
5. A talent tree whose order differs (compare their talent tree file with ours, as the hunter's Beast Mastery tree).
6. Synthetic gear effects in `cmpsim`, and the class's caster or melee stats in `align.py`.

Then run one build, check the sheets match (`layers.py`), check every ability casts in both, and add what bit to the
translations above.
