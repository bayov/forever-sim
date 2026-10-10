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
- [ ] 2. Flametongue procs on a miss, dodge, parry or block
- [ ] 3. Windfury Weapon attacks and the swing timer
- [ ] 4. Stormstrike damage model, dagger. Rage of the Storm's +10% is off the list: the user
  said to assume it works (2026-10-10), because it only matters at level 30 and the sims we
  care about most are level 60. The client agrees (item spell 1309422, +10% to Stormstrike
  only). We count it in when we read the user's Stormstrikes with Rage of the Storm.
- [ ] 5. Flametongue procs and Elemental Devastation
- [ ] 6. Weapon Mastery baseline
- [ ] 7. Spirit regen of the other classes
- [ ] 8. Undead paladin base stats
- [ ] 9. Can Lightning Shield's orbs miss (audit 5.1, added by the audit session 2026-10-10).
  It rides along with item 2's run: Lightning Shield up, part of the fights on level 33 mobs.
