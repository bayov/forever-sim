# Beta verification session: handoff

Written 2026-10-10 by the audit session for the session that runs the beta tests.

## Your job

You go through the items in `notes.md` under `# Need to Verify` that we can test on the WoW
Forever beta, and you settle them with the user. The beta stops at level 30, so only items
marked "Beta" (or that you can show work at level 30) are yours. The user plays the game.
You design each test, tell the user exactly what to do, read what comes back, do the math,
and record the answer.

For each item:

1. Read the bullet in `notes.md` and the audit section it points to (`shaman_audit.md`, item
   numbers like 4.13). Look at the sim code it names, so you know what the sim does today.
2. Write the user a short test plan: character, talents, gear, imbue and totems, where to
   fight, how many hits or procs we need and why that many, and what to screenshot or run.
3. Read the results and work out the answer with the numbers shown. Say how sure the sample
   makes us.
4. Show the user the answer and the sim change it implies. Wait for their OK before changing
   anything in the sim.
5. Record it (see "Recording results") and commit.

The user's standing rules apply to you too:

- Don't take tooltips at face value. The beta's tooltips can have bugs. Ask the user before
  changing anything based on a tooltip.
- Sources, best first: the Forever beta itself, then Season of Discovery (the beta client
  reuses a lot of its code), then Anniversary and 2019 Classic. Avoid citing vmangos, it's a
  private server. https://wago.tools/db2 has the latest Forever client data (build 70291 and
  up).
- Rotation changes go to notes.md, not presets. Rerun knobs or talent searches only when the
  user asks.

## Who else is working here

The audit session (`forever-sim-1c` in ListAgents as of 2026-10-10) keeps going through
`shaman_audit.md` sections 5 to 11 at the same time, in the same jj working copy. It edits
the sim's Go code and tests, `shaman_audit.md`, and adds new items to Need to Verify. Some
of those new items will be marked Beta and become yours. There is also an older idle
`forever-sim-9b` session. Leave it alone.

You can reach the audit session with SendMessage (find its name with ListAgents). Use it
when a beta result changes something in the audit, or when you need to touch a file it holds.

### Who edits what

| Path | Owner | Notes |
|---|---|---|
| `beta_handoff.md` | you | Keep the status list at the end current. |
| The beta results log | you | See "Recording results". |
| `notes.md` `# Need to Verify` | both | The audit session adds bullets, you remove or rewrite the ones you settle. |
| `shaman_audit.md` | audit session | Send it your results and it records them under the item. |
| `sim/**`, `proto/**`, `ui/**`, `*.results` goldens | audit session | You may change them for an OK'd beta result, with the checks below. |

### Rules for sharing the working copy

- Commit only your own paths: `jj commit -m "<msg>" <path> <path>...`. Never run a bare
  `jj commit`, and never use the `/cm` skill's last step that commits everything left,
  because it would sweep up the audit session's unfinished edits.
- Never commit the working copy's `D beta_log.txt`. The user deleted that file in the
  working copy, and its deletion stays out of every commit unless the user says otherwise.
- Never run jj commands that rewrite or drop other people's changes: `jj restore` (without
  the user's go), `jj abandon @`, `jj squash`, `jj edit`, `jj new` with moved changes.
- Edit shared files (`notes.md`) with small Edit tool changes to single bullets. Read the
  file right before you edit it. Never rewrite the whole file with a script or the Write tool.
- Before you change a file under `sim/`, run `jj diff --summary`. When the audit session has
  uncommitted changes in that file, message it and wait, or ask it to make the change.
- Golden refresh (`make update-tests`) copies every `*.results.tmp` in the repo. Run it only
  when no one else has uncommitted `sim/` changes. Otherwise copy only the `.results.tmp`
  files your change produced.
- When a test fails in a file you didn't touch, check `jj diff --summary` first. It's
  probably the audit session in the middle of an edit. Wait a minute and rerun, and don't
  "fix" it.
- jj sometimes says "Could not acquire lock for index file" when the other session runs jj
  at the same time. The commit usually went through. Check `jj log -r 'master..@'` and see
  the `/cm` skill's notes on divergent leftovers.
- Tests need the item database tag: `go test -tags=with_db ./sim/shaman/...`.
- Don't start long rotopt runs unless the user asks. They slow the machine down for both
  sessions.

### Committing

Commit and push without asking once a piece of work is done (the user's standing rule):

```
jj commit -m "<one line in the imperative>

Claude-Session: <your session link from your own system reminder>" <your paths>
jj bookmark set master -r @-
jj git push --remote origin --bookmark master
jj bookmark list --all master
```

The push is done when `master` and `master@origin` show the same commit. Tell the user the
commit ID and message.

## How beta testing has worked so far

- The user plays a level 30 Orc shaman on the beta. They also have a level 20 Undead
  paladin and an Undead rogue around level 15, and can make level 1 characters of any
  class (2026-10-10).
- Before 2026-10-10 we read hits from screenshots of the chat combat log. Now the user types
  `/combatlog` before and after a run, and we read the file it writes in
  `/mnt/d/Gaming/World of Warcraft/_classic_beta_/Logs/WoWCombatLog-*.txt`. It has 0.1 ms
  timestamps, every miss, dodge and parry, and a raw damage field that looks like the hit
  before armor and crit. Count SWING_DAMAGE lines, not their SWING_DAMAGE_LANDED copies.
- For the character's setup, the user types `/bayov export` (their BayovCore addon, since
  2026-10-10). It opens a window with race, level, gear, imbues, talents, skills and sheet
  stats as JSON, and the user pastes it to us. It leaves buffs out on purpose. The memory
  `bayov-export` has the layout.
- Before that, the user ran /run lines and screenshotted what they printed. Ones we've used:
  - `/run for i=1,5 do print(i, UnitStat("player",i)) end`
  - `/run print("ap", UnitAttackPower("player")) print("hp", UnitHealthMax("player")) print("crit", GetCritChance()) print("regen", GetManaRegen())`
  - `/run print(UnitDamage("player"))` for the weapon damage range and the damage percent.
- How we count hits: the whole hit is the shown amount plus the overkill. Crits are listed
  apart (melee crits are 2x a normal hit, spell crits 1.5x). Glancing blows are left out.
  Boars of one level all have the same armor, so a ratio of two attacks on them doesn't
  depend on armor.
- The raw data from every earlier test is in `beta_log.txt`, which is deleted in the working
  copy. Read it with `jj file show -r master beta_log.txt`. It lists the weapons the user
  has (Barbaric Battle Axe of Healing 3.6, Twin-bladed Axe of the Owl 2.7, Bloody Brass
  Knuckles 1.6) and the past runs.

## Recording results

- Raw data (every hit, every timestamp, the /run output) goes in `beta_results.txt`. The
  user chose a new file over bringing `beta_log.txt` back (2026-10-10).
- A settled item comes off the Need to Verify list in `notes.md`. When it settles in the
  sim's favor, remove the bullet. When the sim changes, remove it once the change is in.
  When it only narrows down, rewrite the bullet with what's left.
- When the item has an audit number (4.1, 4.12, 4.13, 7.3 and so on), send the result to the
  audit session so it can record it in `shaman_audit.md`.
- A sim change from a beta result needs a Go test that would have failed before it, like the
  audit's tests in `sim/shaman/enhancement/*_test.go`. Keep the test fast. The user doesn't
  want the suite to grow slow.
- When you learn something future sessions need (a new test method, a beta quirk), save it
  in memory.

## The list

The full text of each item is in `notes.md`. This is the order I'd take them in, by how much
the answer moves the sim's DPS, with what I know beyond the bullet.

1. **Do Windfury Weapon's attacks trigger Flametongue Totem?** The sim procs Flametongue
   Totem on every landed main hand hit, so a Windfury proc can bring 3 Flametongue Totem
   hits (the swing and both attacks). The user remembers seeing only 1. Windfury Weapon and
   Flametongue Totem, then count the Flametongue Totem hits that follow each pair of
   Windfury Weapon hits. Before you plan it, check Flametongue Totem's buff in the client
   (wago SpellAuraOptions: ProcChance, ProcCategoryRecovery and ProcTypeMask) for a cooldown
   that would explain a single proc. Item 4.13, `proc_matrix_test.go`.
2. **Do Flametongue Weapon and Flametongue Totem proc on a miss, dodge, parry or block?** The
   sim procs on every landed hit and counts a block as landed. Fight a mob from the front
   and look for a Flametongue hit right after a Melee or Stormstrike miss, dodge, parry or
   block. Flametongue Weapon blocks Flametongue Totem, so test each one on its own. Item
   4.13.
3. **Do Windfury Weapon's attacks restart the swing timer?** The sim says no (two yellow
   attacks, as in the SoD sim). Era made them extra attacks, which do restart it. The swing
   times around a Windfury proc show it: the next Melee hit comes either on the old schedule
   or a full swing after the Windfury hits. This one needs timestamps finer than a second,
   so it depends on the combat log file or many samples. Item 4.12.
4. **Stormstrike's damage model.** We assume the normalized speed plus 0.3 (3.6 for
   two-handers, 2.7 for one-handers). Two other models fit the data so far. The bullet says
   which runs tell them apart. The data so far is in shaman_audit.md 4.1 and
   `beta_log.txt`. Then **Stormstrike with a dagger** and **Rage of the Storm's +10%
   Stormstrike** use the same method.
5. **Do the Flametongue Weapon and Flametongue Totem procs trigger Elemental Devastation?**
   The sim says no. Elemental Devastation is in row 2 of the elemental tree, so a level 30
   build can take it. A Flametongue crit that brings the buff with no spell crit around it
   answers it. Spell crit is a few percent, so it takes a couple of hundred procs.
6. **Weapon Mastery as a baseline +10%.** The sim doesn't give it, and the level 30 sheet
   showed no bonus. White hits on critters above the UnitDamage range would show it.
7. **Spirit regen of the classes we haven't measured** (druid, hunter, mage, priest,
   warlock). Needs those characters. One GetManaRegen reading below 50 Spirit and one above
   settles a class.
8. **Undead paladin base stats** (paladin section of the list). Needs an Undead paladin.
   Settled, see the status.

Not marked Beta, but maybe testable at level 30. Ask the user whether they're worth it:

- Revelation (the weapon enchant): which spells trigger it, its cooldown, its chance.
- The server batch window (the sim lands white hits and their procs 10 ms after the swing).
- Stacking of the Forever-only elixirs.
- Skipped earlier: the miss rate test (shaman_audit.md 4.4), whether GetManaRegen counts
  MP5 (2.4), Toughness (1.8, item 8.27).

Not yours: anything that needs level 31 or more (the Windfury Totem swing reset needs 32,
Maelstrom Weapon 34, everything under "At 60").

## Status

Keep this current as you go.

- [x] 1. Windfury Weapon attacks and Flametongue Totem. Settled 2026-10-10: Flametongue Totem
  procs only on white swings. The client (70291 and 70338) gives its buff ProcTypeMask 0x4,
  and the user saw no proc from Stormstrike on the beta. The audit session made the sim
  change. Windfury Weapon's attacks weren't counted on their own, so the item 2 run checks
  them on the side (it uses Windfury Weapon with Flametongue Totem).
- [x] 2. Flametongue procs on a miss, dodge, parry or block. Settled 2026-10-10 in the sim's
  favor: Flametongue Totem and Flametongue Weapon proc on every landed hit (blocks included)
  and never on a miss, dodge or parry.
- [x] 3. Windfury Weapon attacks and the swing timer. Settled 2026-10-10 in the sim's favor:
  after 3 Windfury procs from Stormstrike, the next swing came on the old schedule (within
  36 ms), so the attacks don't restart it.
- [ ] 4. Stormstrike damage model, dagger. The model is settled (2026-10-10): 30 Stormstrikes
  with Rage of the Storm fit the sim's normalized 3.6, and the other two models each had hits
  outside their range. Only the dagger is left. Rage of the Storm's +10% is off the list: the user
  said to assume it works (2026-10-10), because it only matters at level 30 and the sims we
  care about most are level 60. The client agrees (item spell 1309422, +10% to Stormstrike
  only). We count it in when we read the user's Stormstrikes with Rage of the Storm.
- [ ] 5. Flametongue procs and Elemental Devastation
- [x] 6. Weapon Mastery baseline. Settled 2026-10-10 in the sim's favor: 128 white hits in the
  15:07 log had raw 191 to 238, inside the sheet's 191.7 to 238.7, and +10% would put 55 of
  them too low. The sheet's damage percent is 1. The user said to assume Weapon Mastery no
  longer exists, at any level.
- [ ] 7. Spirit regen of the other classes. Every class with mana matches under 50 Spirit
  (0.25 x Spirit at level 1, 2026-10-10). Over 50 is only checked for the shaman and
  paladin, and it needs a caster with 50+ Spirit.
- [x] 8. Undead base stats. Settled 2026-10-10: Forever's Undead race is 1.12's + 3 Str + 1
  Agi - 5 Spi for every class. The level 1 rogue and priest match exactly against their 1.12
  rows, the paladin at levels 1 and 20 against the row the sim builds, and the level 1 mage
  once cmangos' swapped Int and Spi are put back. The user OK'd the sim change (generator
  shift, level 60 RaceOffsets, TestForeverUndeadBaseAttributes). Level 1 base health is off
  (paladin 10 under, mage and priest 1 over), and no preset uses it.
- [x] 10. Forever's other races. Settled 2026-10-10: at level 1 the Orc is 0 / 0 / -1 / +2 / -3
  off 1.12 (shaman and warlock), the Tauren -3 / +3 / 0 / +3 / -2 (hunter and druid), the
  Troll 0 / 0 / -1 / +1 / -1 (priest), and the Windshaper -1 / +1 / -1 / +1 / -1 off the
  sim's Human baseline (hunter). Each Forever race is 20 in every attribute plus an offset
  that adds up to 0 and leaves Spirit alone. The user OK'd the sim change (all levels, 60
  included, and the druid's flat attack power below 60). The warrior, hunter and warlock crit
  per Agility at level 1 are off (notes.md), not our sims.
- [x] 9. Can Lightning Shield's orbs miss (audit 5.1, added by the audit session 2026-10-10).
  It rides along with item 2's run: Lightning Shield up, part of the fights on level 33 mobs.
  Answered 2026-10-10: 14 of 71 orbs missed and none of 57 hits crit. The user OK'd it, and
  under Forever an orb now rolls spell hit and never crits (`lightning_shield.go`,
  `TestOrcShamanLightningShieldOrbOutcome`).
- [x] 11. Fire Nova's damage spell (audit 5.4, from notes.md). Settled 2026-10-10: the beta's
  log names SoD's 408424 on every rank 2 hit, not the old totem's 8502. The user OK'd the sim
  change: SoD's base damage and 21.4% at every rank (`fire_totems.go`, a Fire Nova case in
  `TestOrcShamanSpellPowerBeta`).
- [ ] 12. Does Flametongue Totem's proc crit chance follow our crit (audit 5.3, notes.md)? The
  user wants it settled eventually (2026-10-10). Its rate fits our crit (12 of 127 at 12.3%
  spell crit) and a flat 5% is unlikely, but that doesn't show the rate moves with our crit.
  Count it again at a second crit chance, about 500 procs each. The Elemental respec without
  Thundering Strikes (5% less crit) is the cheapest way, together with item 5 (Elemental
  Devastation) and audit 5.6 (Call of Flame with Improved Fire Nova).
  The respec (given to the user 2026-10-10), 26 points: Elemental Convection 5, Elemental
  Warding 3, Call of Flame 3, Elemental Devastation 3, Elemental Focus 1 and Improved Fire
  Nova 2 (17, Improved Fire Nova's row needs 15), then Enhancement Ancestral Knowledge 5,
  Mental Dexterity 3 and Improved Ghost Wolf 1 (9, Mental Dexterity's row needs 5 in
  Enhancement, the user). The filler talents don't change damage. With no Thundering Strikes (all crit, 1% a point in the client), spell crit drops
  from 12.3% to about 7.3%. No Stormstrike or Flurry, because Enhancement's row 4 needs 15
  points. The same run shows whether Call of Flame touches Flametongue Totem (sim: no, raw
  18 would become 20) and Searing Totem (sim: yes). The 20:42 export after the respec has
  7.41% spell crit and 49 spell power (no Mental Quickness), and beta_results.txt has the
  expected Fire Nova, Searing and Flametongue Totem numbers at 49. The
  steps: a `/bayov export`, 5 to 10 Fire Novas into packs of 3 or more mobs, then about 500
  Flametongue Totem procs with Windfury Weapon (shocks are fine, Searing Totem is the same
  element), then optionally about 150 landed swings with Flametongue Weapon for item 5.
- [x] 13. Glancing by mob level (audit 4.6). Parked 2026-10-10: the user called it good enough
  for now. Up to 20:25, +2 (25 of 72) and +3 (26 of 66) fit the sim's 30% and 40%, and also
  the rule on our real skill (32% and 42%), which would take about 2000 swings a level to
  tell apart. The -1 glance, the low +1 and +4 and +5 over the cap don't touch any sim we
  run. No sim change. Keep counting it on new logs (the recheck list).

## Recheck with every new log

The user wants the settled answers checked again as more data comes in, now and after
release (2026-10-10). Each one is a hypothesis that held on one run. When you read a new
combat log, run these checks on it too, and tell the user when one breaks.

- Lightning Shield orbs miss at the spell miss rate for the mob's level and never crit
  (item 9). A crit, or a miss rate far from the spell table, breaks it. Count only MISS. An
  IMMUNE is a mob immune to Nature, like Thundering Boulderkin.
- Stormstrike's raw damage stays inside (weapon roll + AP / 14 x 3.6) x the percent bonuses,
  with 2.7 for one-handers (item 4). Rage of the Storm's +10% is assumed. Read AP from the
  white hits in the same log, not the sheet. A hit outside the range breaks it.
- Flametongue Totem procs on every landed white swing (blocks and glancing blows too) and
  never on a miss, dodge, parry, Stormstrike or Windfury Weapon attack (items 1 and 2).
  Leave out swings where the totem wasn't up or in range, and swings where a Windfury Weapon
  attack or an orb killed the mob within 130 ms (the proc comes 0 to 10 or 90 to 127 ms
  after the swing). The buff stays a few seconds after we run 40 or more yards from the
  totem. scratchpad cl/ftt.py sorts every landed swing by the totem's state. Up to 19:52 all
  67 swings in range of a live totem on a mob that lived procced, and 63 of 63 from 20:09 to
  20:25.
- Flametongue Weapon procs on every landed swing and Stormstrike (blocks too) and never on a
  miss, dodge or parry (item 2). Leave out a hit that kills the mob, and the swing that
  breaks Ghost Wolf (it lands in the form, without the weapon).
- Fire Nova hits name damage spell 408424 at rank 2 and do 102.94 to 117.06 + 21.4% of spell
  power (item 11). The raw field drops the fraction.
- Searing Totem's bolts do 19 to 25 + 1.7% of spell power at rank 3 and use our spell hit
  and crit (audit 7.22). The log's raw damage drops the fraction.
- Searing Totem's bolt gap (audit 7.22, the user waits for more data). The cast takes 2.21
  sec, and the next cast starts about 0.22 sec after it, or about 0.6 sec in a quarter of
  the gaps. It doesn't wait for the bolt to land, which flies 19 yards a second: the gap is
  the same with the totem 2 or 19 yards away. Over 72 gaps up to 16:16 the mean is 2.527 sec,
  plus 21 gaps from 18:27 to 18:42 at 2.480, so 2.516 over 93 (sim 2.5), and a 40 sec totem fires 16 bolts. scratchpad cl/searing.py leaves out gaps
  where the mob moved or the totem switched mobs.
- Partial resists on our spells, by mob level (audit 5.2, the audit session proposed to drop
  the sim's level based partial resists under Forever). Count the landed hits of spells
  that can partially resist (Flametongue Weapon, Lightning Shield orbs, Searing Totem, Flame
  Shock) and how many had a resisted amount. On SPELL_DAMAGE the resisted field is row[-7],
  because spell lines end with an extra "ST" field. Base total (the audit session, up to
  16:05): 0 of 257 on level 33 to 35 mobs, where the sim expects about 57. Plus 16:05 to
  16:16: 0 of 38 on level 33. Count dungeon bosses and elites apart from regular mobs
  (the audit session, notes.md Need to Verify), because Forever could keep the rule for
  them only. Under Classic rules about 24% of hits would be partly resisted at +4 and 12% at
  +2. The log has no elite flag, so tell them apart by name (Bloodmage Thalnos 34, Vishas 32)
  or by the max HP in the advanced fields. Leave out mobs with a school resistance: the
  level 29 Elder Cloud Serpent resisted 20% or 30% of every Nature hit and no Fire (19:30).
  Up to 19:32 no other mob of level 27 to 35 resisted any part of a hit.
- Spell crits by spell and by mob level (audit 5.3, the audit session). Count crits out of
  landed hits, Lightning Shield left out because it never crits. On spell lines the crit flag
  is row[-4], because of the trailing "ST". Leave Windfury Weapon's attacks out, they're
  melee. scratchpad cl/crit.py counts by spell group and level. Base totals, the whole
  WoWCombatLog-101026_150733.txt up to 19:32:39 (12.3% spell crit on the sheet): Searing
  27/215, shocks 8/82, Fire Nova 4/23 (together 39/320, 12.2%), Flametongue Weapon 24/198
  (12.1%), Flametongue Totem 5/82 (6.1%), Flame Shock ticks 15/122 (at 1.5x, audit 5.7).
  Plus 19:32 to 19:52: Flametongue Totem 7/45, so 12/127 (9.4%). Both Flametongue procs
  started low by chance and came up, so the sim keeps the full chance for them (2026-10-10).
  Whether the totem's rate moves with our crit is item 12. A flat 5% for the totem is 1 time in 38, but 12.3% and 7.3% need about 500
  procs to tell apart. The sim lets every player DoT tick crit under Forever, at the crit
  chance of the moment (ruleset.go canCrit).
- Improved Stormstrike's regen buff (1238931) comes on every Stormstrike cast, missed,
  dodged and parried ones too (the audit session, 93 of 93 in the Shimmering Flats log, 17 of
  17 from 18:27). It comes with the cast, so a hit that lands late (105 ms once) shows it
  before the damage.
- The Stormstrike mark (aura 17364 on the mob) comes only with a landed Stormstrike, blocks
  too, and not with a killing blow, a miss, a dodge or a parry (55 of 55 landed non-killing
  ones up to 18:42, and none of the 30 misses, dodges and parries or the 8 killing blows).
- Healing Stream crits under Forever (the audit session's sim change). Our own totems' heals
  crit on 29 of 605 up to 19:52 (4.8%), well under the sheet's 12.3% spell crit, while
  Searing Totem's bolts crit at the full chance. Count only heals from our own totems: the
  log has other shamans' totems too.
- After a Windfury Weapon proc from a Stormstrike, the next white swing comes on the old
  schedule, not a full swing after the proc (item 3). Skip procs with a Flurry change in
  between.
- Parry haste follows Classic's rule, for mobs and for us (audit 6.2, the audit session's sim
  change, 5740215a). With more than 60% of the swing left, the next swing comes 40% of the
  swing sooner. With 20 to 60% left, it comes 20% of the swing after the parry. With 20% or
  less left, nothing changes. Base counts up to the audit session's read: 58 of 58 mob
  parries within 0.1 sec, and 23 of our 25 within 0.16 sec (the other 2 waited on a cast).
  scratchpad cl/parry.py (from the audit session's enemy/fit.py) takes each mob's speed from
  its most common swing interval, so the 3.05 sec Boulderkin and the 1.2 sec Needles Cougar
  fit too. It breaks ties in the same ms by file order. The whole log up to 20:25: mobs 72
  of 73 within 0.1 sec (largest 0.11), us 23 of 29 within 0.1 sec, 3 more within 0.16
  (Flurry changed our speed), and 3 swings that were already late (19:49:02, 19:52:01 and
  20:21:01, the swing was held).
- Glancing by mob level (audit 4.6, parked as good enough with the sim's rule, 2026-10-10).
  Count our white swings that glanced out of all our white swings (misses, dodges and
  parries included), per target level. The mob's level is the field just before the damage
  suffix on the mob's own SWING_DAMAGE lines (row[-11]), matched to our target by GUID.
  Note our weapon skill from the latest `/bayov export` too (Two-Handed Maces was 149 of 150
  at 14:50 and still at 18:21 on 2026-10-10). Base totals (WoWCombatLog-101026_150733.txt up to 15:51, the
  audit session's count), plus 16:04 to 16:16: level 33 23/55 (42%), level 34 66/111 (59%),
  level 35 6/8. Plus 18:27 to 18:42: levels 27 to 29 0/61, where the sim gives 0. Plus 18:57
  to 19:06: level 34 10/21, so level 34 is 76/132 (58%). Plus 19:06 to 19:32: level 34
  26/55 and levels 27 to 29 0/68. So level 34 is 102/187 (55%) and levels 27 to 29 0/129.
  Whole log up to 20:09 by mob level: 0/175 at levels 24 to 28, 1/68 at 29 (a Thundering
  Boulderkin at 19:58, the first glance below our level), 0/7 at 30, 3/11 at 32, 23/57 at 33,
  102/187 at 34, 6/8 at 35. Plus 20:09 to 20:18: 1/9 at 31 and 20/50 at 32. Whole log up
  to 20:25: 3/33 at 31, 25/72 at 32 (35%), 26/66 at 33 (39%). scratchpad cl/glance.py counts by level, with the mob's level from its spell
  lines too. The level field on our own lines is our item level (33), not our level. Add
  each new log to these. Without the cap the sim would give 50% at +4 and 60% at
  +5. The glance at -1 fits glancing on our real weapon skill (149): 10% plus 2% a point of
  the mob's defense over 149 gives 2% at -1.
