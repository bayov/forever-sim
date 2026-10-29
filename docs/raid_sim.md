# Raid sim

A sim mode for a whole raid. Every member is simmed together in one fight against the boss, and the result
is the raid's total damage. The DPS and PvP modes each sim one player against an assumed raid. Here the
other members are real players.

This is a first draft (2026-10-29), written from the user's idea in TASKS.md (2026-10-14). Nothing in it is
decided yet. Each section gives a proposal, and [Open decisions](#open-decisions) lists the questions the
user needs to settle first.

## What it answers

- **Total raid DPS for a comp.** For example: 2 tanks, 5 healers and 13 DPS in Hyjal Summit, or 4 tanks,
  11 healers and 25 DPS for Onyxia. The total counts pets.
- **What a class is worth to a raid.** That is its own DPS plus what it does for everyone else. The DPS
  Report ranks specs on their own DPS, in a raid that already has every buff. It can't show that the first
  warrior brings Battle Shout and Sunder Armor, that the first shaman brings Windfury Totem, or that a third
  paladin adds less than the second. The raid sim can: a member's value is the raid DPS with that member
  minus the raid DPS with a filler in their place.
- **Who holds which duty.** Which warrior keeps up Battle Shout, which warlock uses which curse, which
  druid uses Faerie Fire, and what each duty costs the player who holds it. Today the DPS Report models this
  with its duty presets (standard 23), one player at a time.
- **Each member's DPS in a real raid, next to their DPS Report number.** The difference is what the
  report's shared raid assumes and this raid doesn't have, or the other way round.
- **Later: the best comp.** That means which classes and specs, how many of each, and who goes in which
  group (see [Comp search](#comp-search)).

## Forever changes the question

In Classic, raid comps were mostly about groups. Battle Shout, the totems, Trueshot Aura, Leader of the
Pack and Moonkin Aura reached only their own party, so the work was building a melee group around a shaman.

Client 70425 made all of those raid-wide (`fc_patch-notes.txt`: Battle Shout, Trueshot Aura, Grace of Air,
Strength of Earth, Windfury and Flametongue Totem, Moonkin Form, Leader of the Pack). What is still
party-wide is mostly mana and survival:

- Mana Spring, Stoneskin, Healing Stream, Mana Tide and Tranquil Air Totem
- the paladin auras
- Blood Pact

`ui/core/components/inputs/buffs_debuffs.ts` describes the party buffs the same way.

Forever also made the old shared damage debuffs personal: Improved Scorch, Winter's Chill, Shadow Weaving,
Improved Shadow Bolt and Stormstrike (`sim/core/debuffs.go`). Faction doesn't limit the comp either,
because both sides have paladins and shamans (`sim/core/buffs.go`).

So a Forever comp is mostly about **which classes and specs, and how many of each**. The shared effects
that are left are the raid-wide buffs, which are capped per provider (one air totem per shaman, one
blessing per paladin per target), and the shared debuff slots. Group placement only changes mana. Over a
180-second fight that is worth little, but over a long boss it can matter. The design puts its effort in
that order.

## The score

**Raid DPS: all damage the raid deals to the boss, divided by the fight length.** Bigger is better.

The result also shows:

- each party's DPS and each member's DPS
- for each DPS member, the difference from their DPS Report build
- each buff and debuff's uptime, and who provided it

Healing and threat are not scored (see [Members](#members) and [The fight](#the-fight)).

## Members

A raid of 10, 20 or 40 members in groups of 5. Forever's launch raids are the Barrow Deeps (10), Hyjal
Summit (20) and Onyxia's Lair (40).

**DPS members are DPS Report builds.** Each one is a Level 60 `PresetBuild`: gear, talents, rotation, race,
professions and consumes. Its request in `tools/dps_report/req` is already verified to reproduce its
published DPS. Duty presets become a choice on the member rather than separate entries in the list. For
example, a Fury warrior is either "on Battle Shout" or not.

**Tanks and healers are support members (proposal).** A support member fills a slot, has a class and a
spec, and gives the raid what that spec gives. It runs no rotation, deals no damage and spends no mana.
The engine can't run them as real players today:

- The healing specs are unregistered (`sim/register_all.go`), and some are left out of the build entirely.
- Nobody takes damage that needs healing.
- Nothing models aggro, so a simmed tank doesn't do a tank's job.

What a support member gives, with a proposed default for each spec:

| Support | Gives (default) |
|---|---|
| Tank warrior | Sunder Armor (5 stacks on the existing schedule), Demoralizing Shout, Thunder Clap; Battle Shout when no DPS warrior holds it |
| Feral tank | Faerie Fire (Feral), Leader of the Pack, Demoralizing Roar |
| Protection paladin | one blessing, one aura |
| Holy paladin | one blessing, one aura |
| Healing priest | Fortitude, Divine Spirit (if talented), Shadow Protection |
| Restoration shaman | one totem per element (earth, air, water, fire), Mana Tide to its party |
| Restoration druid | Gift of the Wild |

Each of these can be changed per member: which blessing, which totems, whether a talent is "improved", and
whether it has a duty at all.

Simming real tanks and healers can come later, member by member, once the engine has damage intake,
healing on other players and threat. The rest of the design doesn't change when that happens.

## Buffs and debuffs come from the members

A raid sim has **no assumed raid**. `Raid.buffs` and `Raid.debuffs` stay empty, and every buff and debuff
comes from a member. Each one arrives in one of two ways, never both:

- **Given.** It goes on before the fight as a permanent aura over its scope (the party or the raid). This
  covers:
  - the long buffs from anyone (Fortitude, Arcane Intellect, Gift of the Wild, blessings, auras), which
    cost no GCD in the sim (standard 23)
  - everything a support member provides
- **Cast.** A DPS member's rotation casts it on its duty, and the engine sends it to its scope. This covers
  Battle Shout, the Enhancement and Elemental shamans' own totems, Sunder Armor from a DPS warrior, Faerie
  Fire, the curses, Judgement of the Crusader and Hunter's Mark. A dropped totem or a late debuff then
  costs the whole raid, as it would in game.

**The provider table** says, for each buff and debuff:

- which classes and specs provide it, and which talents improve it
- its scope: party or raid
- whether it costs a GCD
- what it excludes, and what limit it has per provider

The scope comes from client data, generated rather than written by hand: in the SpellEffect table,
aura effect 35 is a party aura and effect 65 is a raid aura. The patch notes are a cross-check.

The rules the table has to cover:

- **One per provider.**
  - A paladin gives each player one blessing. Salvation doesn't count against that: since client 70425
    it's no longer a blessing.
  - A shaman has one totem per element. Grace of Air and Windfury Totem together need two shamans.
  - A warlock casts one curse.
- **Shared slots.**
  - Faerie Fire and Curse of Recklessness share the minor armor slot (standard 20).
  - Insect Swarm and Scorpid Sting share a slot, and several Insect Swarms don't stack their hit reduction.
  - Windfury Totem and Flametongue Totem exclude each other.
- **Stacking.**
  - Sunder Armor stacks to 5 from any warrior.
  - Two shamans' Mana Springs: an open question (notes.md, item 8.45).
  - Two shamans' other totems: an open question (notes.md, item 8.46).
  - Until those are settled, the raid sim uses the sim's current rules and lists them as assumptions.
- **Nobody provides it.** If no member gives a buff, nobody has it. When three paladins can't cover Kings,
  Might, Wisdom and Light for everyone, the raid sim shows which members miss out.

**Duties are assignments.** The raid assigns each GCD duty to one member: Battle Shout, each curse, Faerie
Fire, Judgement of the Crusader, Sunder Armor (usually a tank) and each element's totem. A DPS member with a
duty runs their build's duty rotation. Every other member runs the base rotation. When nobody holds a
duty, the raid goes without it.

By default, the raid sim fills duties from the members it has, cheapest first, and shows the result. The
user can change any assignment.

## Groups

Groups only matter for party-scoped buffs. These are:

- Mana Spring, Mana Tide and Healing Stream from a shaman
- Blood Pact from a warlock's Imp
- the paladin auras, of which none raise DPS in Forever (it has no Sanctity Aura)
- Stoneskin and Tranquil Air Totem, which change no DPS

**Proposal:** place groups by rule, and let the user drag members to change it. The rule puts casters with
the shamans that keep up Mana Spring and Mana Tide, and leaves everyone else where the comp put them. The
groups only get a real search if long fights show they matter (see
[Comp search](#comp-search)).

## The fight

- **The boss.** It's the same generic level 63 boss as the DPS Report (3731 armor, Unknown creature type):
  one target, tanked by the first tank. Forever's raid bosses aren't known yet. When they are, each one is
  an encounter, like the Vaelastrasz encounter in `sim/encounters`.
- **Length.** It defaults to a fixed duration, the report's 180 seconds, so member numbers compare with the
  DPS Report. A "kill the boss" option uses the boss's health, which needs a health per raid size. That
  waits for real Forever boss health, or the user's estimate.
- **Raid damage.** Every member the boss isn't attacking takes the encounter's raid damage hits, as in the
  report (2 a minute at 1500, standard 24). That damage only feeds procs and Rage, since nobody heals.
- **Threat.** It's reported, not modelled (proposal). Every member's threat is already measured
  (`UnitMetrics.threat`). The result warns when a DPS member's threat passes the tank's, and says when.
  Nothing stops that member's damage. A threat cap that holds members back could come later.
- **Debuff limit.** Classic had a cap on the number of debuffs a boss could carry. A full raid's DoTs,
  stings and curses can reach it, and the sim doesn't model it. Forever's limit isn't known. Proposal: add
  the cap as an encounter setting, off until the limit is measured. The game drops the oldest debuff, and
  the result shows when the cap was reached.

## Results

- Raid DPS (mean, standard deviation, 95% interval), each party's DPS, and each member's DPS with their
  difference from the report.
- For each buff and debuff: its provider, its scope and its uptime.
- Each duty and what it cost the member who held it, which is their DPS against their base build in the
  same raid.
- Members who miss a buff the raid could have given them.
- Threat warnings.
- A member's value (raid DPS minus raid DPS with a filler in their place), when asked for. This takes one
  sim per member, so it's a button rather than part of every run.

## Checking it

1. **The report raid.** Put each DPS Report build in a raid, with the report's shared raid buffs and
   debuffs switched back on as assumed. Every member should match their published DPS within noise. This
   proves the multi-player engine changes no one's numbers. It is a test, run on the 10 and 20 reference
   comps.
2. **Scope tests.** A member gets a party buff only from their own party, and a raid buff from anyone. A
   shaman's dropped totem reaches the right members. Battle Shout cast by a warrior reaches the whole raid.
3. **After 2026-12-09**, the user's own raid logs (`tools/warcraftlogs`, `logcheck`) and public logs are
   checked against the sim: raid DPS by comp and buff uptimes.

## Comp search

This comes after the sim itself. The inputs are:

- the raid size
- the number of tanks and healers (or a fixed list of them)
- the builds to choose DPS members from
- optionally, limits per class

The search:

1. **Fill greedily.** Start from the tanks and healers. Add the DPS build that raises raid DPS the most,
   and repeat until the raid is full.
2. **Swap.** Try replacing each member with each other build, and keep any swap that improves raid DPS.
   Repeat until no swap helps.
3. **Assign.** For each comp, fill duties and groups by the rules above. A real search of groups and duties
   is only added if it changes the answer.

Candidates are compared at a few thousand iterations with the same seeds (common random numbers), so a
small difference between two comps isn't noise. The best few are then confirmed at 20,000 iterations
(standard 15).

What this costs depends on how fast a raid sims. That is unmeasured; see groundwork step 5.

## What the engine has today

The engine was written to sim raids (it's upstream wowsims' design), but this fork has only ever simmed
one player.

**What works:**

- `RunRaidSim` builds and runs a raid of any size, in parties of 5 (`sim/core/raid.go`,
  `environment.go`).
- Metrics add up per member, per party and for the raid, pets included (`raid.go`,
  `metrics_aggregator.go`).
- Iterations are split across CPUs (`sim_concurrent.go`).

**What is broken in a real raid:**

- **Presims.** `presim.go` allocates options for 25 players and indexes them by player index, which runs to
  39. A 40-player raid panics on any fight that runs presims (health-based fights, tanks with a healing
  model).
- **Gaps in a party.** `NewParty` drops empty slots, but `applyCharacterEffects` and the rotation setup look
  players up by their position in the shortened list instead of `PartyIndex`. With a gap, members get the
  wrong buffs and the wrong rotation.
- **Raid buffs.** `GetRaidBuffs` writes into the request proto without copying it.
- **Atiesh.** `AddPartyBuffs` runs twice per player, so Atiesh counts double for the classes that don't
  override it.
- **One-player assumptions.**
  - The external Improved Shadow Bolt and Stormstrike debuffs, which Forever skips anyway (`debuffs.go`).
  - Stat weights, the bulk sim, `sim/lib` and the test suite (`Parties[0].Players[0]`).
  - `spell_result.go` reads attack tables by `target.Index`. That is wrong for spells cast on a friendly
    player.

**Buffs:**

- Every buff in the settings is a raid buff, with no party scope (`buffs.go`).
- Members' passive buffs reach the raid through `AddRaidBuffs` hooks: Arcane Intellect, Fortitude and
  Divine Spirit, Trueshot Aura, Moonkin Aura and Leader of the Pack, Blood Pact.
- Cast buffs don't reach their scope:
  - A warrior's Battle Shout reaches only their own party (`warrior/shouts.go`), though Forever's is
    raid-wide.
  - A shaman's Strength of Earth, Grace of Air and Windfury Totem reach only the shaman
    (`shaman/earth_totems.go`, `air_totems.go`).
- Settings and casts share aura labels, so they mostly dedupe rather than stack. Nothing switches a
  settings buff off when a member provides it.

**Everything else:**

- **Debuffs.** They come from settings, or from casts through the same core auras. There is no debuff cap.
- **Healers, tanks, threat.**
  - The healing specs are unregistered.
  - A tank's health doesn't drop without a healing model.
  - Threat is measured but never used.
- **Speed.** Unmeasured; `sim/raid_bench_test.go` is a dead WotLK fixture with no benchmarks. Several costs
  grow with the number of players:
  - Pending actions are inserted into a sorted list (`sim.go`).
  - Every time step settles every mana bar.
  - Every unit's event wakes every rotation, because idle APLs check one global event counter
    (`apl_idle.go`).
  - Attack tables are all units by all units.
  - A raid of 20 could cost far more than 20 single sims.
- **The UI.** `ui/raid` is upstream's raid sim page. It still compiles and is built, but it is unlaunched
  (`launched_sims.ts`) and stale:
  - WotLK buff and scroll ids in `raid_stats.tsx` and `settings_tab.ts`.
  - Raid sizes of 5, 10, 25 and 40.
  - Its member list is the specs' `raidSimPresets`, which are old Classic data. The Enhancement and Warden
    entries are copies of Balance, and the hunters are commented out.
  - It works: drag members into slots, edit a member in the full single-player UI, saved raids, JSON and
    Warcraft Logs import.

## Groundwork

In order. Each step is usable on its own.

1. **Multi-player correctness.**
   - Fix the presim array, the slot index (`PartyIndex`), the proto copy in `GetRaidBuffs` and the double
     `AddPartyBuffs`.
   - Replace the dead raid fixtures with a Forever one: a raid of DPS Report builds, plus a test with gaps
     in its parties.
   - Run check 1 (the report raid).
2. **Cast buffs reach their scope.**
   - Battle Shout covers the raid under Forever.
   - A shaman's totems reach their scope: the raid for Strength of Earth, Grace of Air, Windfury and
     Flametongue; the party for Mana Spring, Stoneskin, Healing Stream and Mana Tide.
   - Windfury Totem's 1.5 second cooldown applies to each member who receives it.
   - Run the scope tests (check 2).
3. **The provider table and the resolver.**
   - Generate the scope column from client SpellEffect data.
   - The resolver turns a comp (members, support members and duties) into the given buffs per party and
     the given debuffs, and turns off assumed settings.
   - It enforces the rules in [Buffs and debuffs come from the members](#buffs-and-debuffs-come-from-the-members),
     and records which member provided what, for the results.
   - Because the settings don't scope buffs today, it gives each party its own buff set.
4. **Support members.**
   - Add a slot kind to the raid proto that has a class, a spec and what it gives, and runs no rotation.
   - Add a default for each spec, using the table in [Members](#members).
5. **Speed.**
   - Add a benchmark for raids of 10, 20 and 40 report builds.
   - Fix whatever dominates. The suspects are listed under What the engine has today.
   - This decides how many comps the search can afford.
6. **A command-line raid sim (`tools/raidsim`).**
   - It takes a comp file: the size, and the parties as report build ids, support members and duties.
   - It builds the request from `tools/dps_report/req`, runs it and prints the results above.
   - This produces numbers before any UI.
7. **The raid page.** Bring `ui/raid` up to Forever:
   - Raid sizes of 10, 20 and 40.
   - Members from the Level 60 `PresetBuild`s and the support catalog.
   - Duty and blessing assignments.
   - A results view, the Forever buff and debuff lists, and a launch.
   - Remove the WotLK leftovers.
8. **Member value and the comp search** (see [Comp search](#comp-search)).

Steps 1 to 6 are Go and tools. Step 7 is UI. None of it changes a DPS Report number. If the report later
gets a raid page, the deploy rule in CLAUDE.md applies.

## Open decisions

1. **Tanks and healers.** Should they be support members (buff providers with no rotation) for now?
   Proposed: yes. Simming them for real needs healing, damage intake and threat, each a project of its own.
2. **Where the mode lives.**
   - Option A (proposed): the `ui/raid` page, brought up to Forever, with a "Raid" entry in the
     sidebar's mode picker that opens it.
   - Option B: a fifth mode inside the single-player page. That page is built around one player: its
     tabs, its history and its last run are all per player.
3. **Command line first.** Should numbers come from `tools/raidsim` before the page is built? Proposed: yes.
4. **Fight length.** Proposed: a fixed 180 seconds by default, with "kill the boss" as an option once
   there's a boss health per raid size.
5. **Threat.** Proposed: report it and warn, without holding anyone back.
6. **Debuff limit.** Should there be a cap setting, off until Forever's limit is measured? And who measures
   it, and how? A beta test needs a boss with enough debuffs on it.
7. **Reference comps.** Proposed defaults, as tanks / healers / DPS:
   - Barrow Deeps (10): 2 / 2 / 6
   - Hyjal Summit (20): 2 / 5 / 13
   - Onyxia (40): 4 / 11 / 25

   Which DPS fill them first: one of each class, or what the user's guild runs?
8. **Duties by default.** Should the sim fill them automatically, cheapest first, with the user able to
   change them? Proposed: yes.
