# Raid sim

A sim mode for a whole raid. Every member is simmed together in one fight, and the result is the raid's
total damage. The DPS and PvP modes each sim one player against an assumed raid. Here the other members are
real players.

The point is finding the best comp: which specs to bring, and how many of each, for a raid size with a set
number of tanks and healers.

The decisions marked "(the user, 2026-10-29)" are settled. Everything else is a proposal, and the
questions still open are at the end. The first draft was written on 2026-10-29 from the user's idea in
TASKS.md (2026-10-14).

## What it answers

- **The best comp** (the main question). Given a raid size and the number of tank and healer slots, it finds
  the DPS specs and builds that give the most raid DPS (see [Comp search](#comp-search)).
- **Total raid DPS for a given comp.** The total counts pets.
- **What a spec is worth to a raid.** That is its own DPS plus what it does for everyone else. The DPS
  Report ranks specs on their own DPS, in a raid that already has every buff. It can't show that the first
  warrior brings Battle Shout and Sunder Armor, that the first shaman brings Windfury Totem, or that a third
  paladin adds less than the second. The raid sim can: a member's value is the raid DPS with that member
  minus the raid DPS with a filler in their place.
- **Assignments.** Which warrior keeps up Battle Shout, which warlock uses which curse, which druid uses
  Faerie Fire, and what each assignment costs the player who holds it. The DPS Report calls these duties
  and models them one player at a time (standard 23).
- **Each member's DPS in a real raid, next to their DPS Report number.** The difference is what the
  report's shared raid assumes and this raid doesn't have, or the other way round.

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
Improved Shadow Bolt and Stormstrike (`sim/core/debuffs.go`). Faction doesn't limit the comp, because both
sides have paladins and shamans (`sim/core/buffs.go`). And Forever has no limit on how many buffs or
debuffs a target can carry (the user, 2026-10-29).

So a Forever comp is mostly about **which specs, and how many of each**. The shared effects that are left
are:

- the raid-wide buffs, which are capped per provider (one air totem per shaman, one blessing per paladin
  per target)
- the shared debuff slots

Group placement only changes mana. Over a 180-second fight that is worth little, but over a long boss it
can matter.

## The score

**Raid DPS: all damage the raid deals to the bosses, divided by the fight length.** Bigger is better.

The result also shows:

- each party's DPS and each member's DPS
- for each DPS member, the difference from their DPS Report build
- each buff and debuff's uptime, and who provided it

## The page

The raid sim is **a page of its own** (the user, 2026-10-29). It is not part of a spec's page or of the DPS
Report.

**The page sets up only the raid:**

- the raid size and how many slots are kept for tanks and healers
- which member goes in each slot, and which group they're in
- assignments
- the encounter
- the comp search

**It doesn't set up a member's gear, talents, rotation or consumes.** Each member is taken whole from a
spec's page, in one of three ways:

- one of the spec's presets (the Level 60 `PresetBuild`s, the same builds as the DPS Report)
- a saved setup from that spec's page
- that page's current settings

The spec pages already keep these in the browser's storage (`IndividualSimUI`'s saved gear, talents,
rotation and settings), so the raid page can read them.

**From any member, "Open in spec page"** opens that spec's page with the member's setup, and with the raid
buffs and debuffs that this raid gives that member as its assumed raid. That is where the user changes the
member's gear, talents, rotation or consumes. A sim there shows the member as they'd do in this raid. The
change comes back to the raid page as a saved setup on the spec's page.

**The upstream raid page is replaced (proposal).** `ui/raid` (see
[What the engine has today](#what-the-engine-has-today)) edits each member inside the raid page, takes its
members from the old `raidSimPresets` and is full of WotLK leftovers. The new page starts fresh. It keeps
what still fits: `ui/core/raid.ts` and `party.ts`, and the raid picker's drag and drop of members between
slots and groups.

## Members

A raid of 10, 20 or 40 members in groups of 5. Forever's launch raids are the Barrow Deeps (10), Hyjal
Summit (20) and Onyxia's Lair (40).

**The number of tank and healer slots can be set for each raid size** (the user, 2026-10-29). The defaults
and the usual range:

| Raid | Tanks | Healers | DPS |
|---|---|---|---|
| 10 | 1 (1 to 2) | 2 | 7 (6 to 7) |
| 20 | 2 | 4 (4 to 5) | 14 (13 to 14) |
| 40 | 3 | 8 (7 to 9) | 29 (28 to 30) |

**DPS members are DPS spec builds** (see [The page](#the-page)). An assignment makes a member run their
build's assignment rotation, for example the Fury warrior's Battle Shout preset.

**Simming tanks and healers is out of scope for now** (the user, 2026-10-29). Until then, a tank or healer
slot holds a placeholder: a spec that gives the raid its buffs and debuffs, runs no rotation, and deals no
damage.

Tanks do deal damage, even when holding threat comes first. Healers can too, when they have time to weave
damage spells between heals. So placeholders leave the raid's DPS somewhat short. That's the same for
every comp with the same slot counts, so comps still rank the same way. When tanks and healers are simmed
for real, their DPS joins the total, and the rest of the design doesn't change.

What each placeholder gives, with a proposed default for each spec:

| Placeholder | Gives (default) |
|---|---|
| Tank warrior | Sunder Armor (5 stacks on the existing schedule), Demoralizing Shout, Thunder Clap; Battle Shout when no DPS warrior has it |
| Feral tank | Faerie Fire (Feral), Leader of the Pack, Demoralizing Roar |
| Protection paladin | one blessing, one aura |
| Holy paladin | one blessing, one aura |
| Healing priest | Fortitude, Divine Spirit (if talented), Shadow Protection |
| Restoration shaman | one totem per element (earth, air, water, fire), Mana Tide to its party |
| Restoration druid | Gift of the Wild |

Each of these can be changed per slot: which blessing, which totems, whether a talent is "improved", and
which assignments it holds.

## Buffs and debuffs come from the members

A raid sim has **no assumed raid**. `Raid.buffs` and `Raid.debuffs` stay empty, and every buff and debuff
comes from a member. Each one arrives in one of two ways, never both:

- **Given.** It goes on before the fight as a permanent aura over its scope (the party or the raid). This
  covers:
  - the long buffs from anyone (Fortitude, Arcane Intellect, Gift of the Wild, blessings, auras), which
    cost no GCD in the sim (standard 23)
  - everything a tank or healer placeholder provides
- **Cast.** A DPS member's rotation casts it as their assignment, and the engine sends it to its scope.
  This covers Battle Shout, the Enhancement and Elemental shamans' own totems, Sunder Armor from a DPS
  warrior, Faerie Fire, the curses, Judgement of the Crusader and Hunter's Mark. A dropped totem or a late
  debuff then costs the whole raid, as it would in game.

**The provider table** says, for each buff and debuff:

- which specs provide it, and which talents improve it
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

## Assignments

An assignment is a GCD-costing buff or debuff that one member keeps up for the raid:

- Battle Shout
- each curse
- Faerie Fire
- Judgement of the Crusader
- Sunder Armor, usually a tank
- each element's totem
- each paladin's blessing

**Each assignment is either automatic or set by hand** (the user, 2026-10-29). Automatic assignment gives
it to the member it costs least, and the page shows who that is.

A DPS member with an assignment runs their build's assignment rotation. Every other member runs the base
rotation. When nobody holds an assignment, the raid goes without it.

The comp search assigns automatically too, so every comp it tries has its assignments filled.

## Groups

Groups only matter for party-scoped buffs. These are:

- Mana Spring, Mana Tide and Healing Stream from a shaman
- Blood Pact from a warlock's Imp
- the paladin auras, of which none raise DPS in Forever (it has no Sanctity Aura)
- Stoneskin and Tranquil Air Totem, which change no DPS

**Proposal:** place groups by rule, and let the user drag members to change it. The rule puts casters with
the shamans that keep up Mana Spring and Mana Tide, and leaves everyone else where the comp put them. The
comp search uses the same rule. It only searches groups if long fights show they matter.

## The fight

- **The bosses.** By default it's the DPS Report's generic level 63 boss (3731 armor, Unknown creature
  type), tanked by the first tank. An encounter can have several bosses, each tanked by its own tank.
  Forever's raid bosses aren't known yet. When they are, each one is an encounter, like the Vaelastrasz
  encounter in `sim/encounters`.
- **Length** (the user, 2026-10-29). It can be set two ways:
  - A fixed duration, with the report's 180 seconds as the default, so member numbers compare with the
    DPS Report.
  - The boss's health. The fight ends when every boss is dead, each one dying on its own, through the
    separate mob death the leveling sim uses (`sim/core/target_death.go`).
- **Raid damage.** Every member the boss isn't attacking takes the encounter's raid damage hits, as in the
  report (2 a minute at 1500, standard 24). That damage only feeds procs and Rage, since nobody heals.
- **Threat is out of scope for now** (the user, 2026-10-29). When it's added, it is reported but not
  modelled:
  - a warning whenever the tank couldn't hold threat, saying when and against whom
  - a results page that drills down into each member's threat on each boss
  - nothing holds a member's damage back

  Every member's threat is already measured (`UnitMetrics.threat`).

## Results

- Raid DPS (mean, standard deviation, 95% interval), each party's DPS, and each member's DPS with their
  difference from the report.
- For each buff and debuff: its provider, its scope and its uptime.
- Each assignment and what it cost the member who held it, which is their DPS against their base build in
  the same raid.
- Members who miss a buff the raid could have given them.
- A member's value (raid DPS minus raid DPS with a filler in their place), when asked for. This takes one
  sim per member, so it's a button rather than part of every run.
- Later, threat (see [The fight](#the-fight)).

## Checking it

1. **The report raid.** Put each DPS Report build in a raid, with the report's shared raid buffs and
   debuffs switched back on as assumed. Every member should match their published DPS within noise. This
   proves the multi-player engine changes no one's numbers. It is a Go test, run on the 10 and 20 default
   comps.
2. **Scope tests.** A member gets a party buff only from their own party, and a raid buff from anyone. A
   shaman's dropped totem reaches the right members. Battle Shout cast by a warrior reaches the whole raid.
3. **After 2026-12-09**, the user's own raid logs (`tools/warcraftlogs`, `logcheck`) and public logs are
   checked against the sim: raid DPS by comp and buff uptimes.

## Comp search

The main thing the raid sim is for (the user, 2026-10-29).

**The inputs:**

- the raid size, and the number of tank and healer slots
- the tank and healer specs, or "any"
- the DPS builds to choose from: every Level 60 preset by default, or the user's pick
- optionally, limits per spec or class
- the encounter

**The search:**

1. **Fill greedily.** Start from the tanks and healers. Add the DPS build that raises raid DPS the most,
   and repeat until the raid is full.
2. **Swap.** Try replacing each member with each other build, and keep any swap that improves raid DPS.
   Repeat until no swap helps. The healer and tank specs are swapped the same way when they're set to
   "any", since their buffs differ.
3. **Assign.** Each comp it tries gets automatic assignments and rule-placed groups.

Candidates are compared at a few thousand iterations with the same seeds (common random numbers), so a
small difference between two comps isn't noise. The best few are then confirmed at 20,000 iterations
(standard 15).

**The results:**

- the best comp, with its raid DPS
- the runners-up, and how far each one is behind the best
- each member's value in the best comp

The search runs in Go, as one request, so the page runs it in its workers the same way the tests do.
What it costs depends on how fast a raid sims. That is unmeasured; see groundwork step 5. A big search
(a 40-player raid over every build) may need to run outside the browser, through the same request.

## What the engine has today

A raid sim already exists, in two parts, both from upstream wowsims:

- **The engine**, which this fork uses for every sim.
- **The `ui/raid` page**, which is hidden and stale.

### The engine

The engine sims raids. Every sim here is a raid sim: the spec pages and the DPS Report each send a raid of
one player, plus the assumed buffs. But this fork has never simmed more than one player.

**What works:**

- `RunRaidSim` builds and runs a raid of any size, in parties of 5 (`sim/core/raid.go`,
  `environment.go`).
- Metrics add up per member, per party and for the raid, pets included (`raid.go`,
  `metrics_aggregator.go`).
- Iterations are split across CPUs (`sim_concurrent.go`).

**What is broken in a real raid:**

- **Presims.** `presim.go` allocates options for 25 players and indexes them by player index, which runs to
  39. A 40-player raid panics on any fight that runs presims, and that includes health-based fights.
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

- **Debuffs.** They come from settings, or from casts through the same core auras.
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

**So the engine stays.** The groundwork below fixes it for real raids.

### The upstream page

`ui/raid` is upstream's raid sim page. It still compiles and is built, but it is unlaunched
(`launched_sims.ts`) and stale:

- WotLK buff and scroll ids in `raid_stats.tsx` and `settings_tab.ts`.
- Raid sizes of 5, 10, 25 and 40.
- Its members come from the specs' `raidSimPresets`, which are old Classic data. The Enhancement and
  Warden entries are copies of Balance, and the hunters are commented out.
- It configures each member inside the raid page, in the full single-player UI.

What works: dragging members into slots, saved raids, and JSON and Warcraft Logs import.

The new page replaces it (see [The page](#the-page)).

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
   - The resolver turns a comp (members, placeholders and assignments) into the given buffs per party and
     the given debuffs, and turns off assumed settings.
   - It enforces the rules in [Buffs and debuffs come from the members](#buffs-and-debuffs-come-from-the-members),
     makes the automatic assignments, and records which member provided what, for the results.
   - Because the settings don't scope buffs today, it gives each party its own buff set.
   - The same resolver gives "Open in spec page" its assumed raid for one member.
4. **Tank and healer placeholders.**
   - Add a slot kind to the raid proto that has a class, a spec and what it gives, and runs no rotation.
   - Add a default for each spec, using the table in [Members](#members).
5. **Speed.**
   - Add a benchmark for raids of 10, 20 and 40 report builds.
   - Fix whatever dominates. The suspects are listed under [The engine](#the-engine).
   - This decides how big a comp search the browser can run.
6. **The raid page.**
   - Raid size and tank and healer slots.
   - Members from presets, the spec pages' saved setups or their current settings.
   - Groups, assignments, the encounter (fixed length or boss health) and the results.
   - "Open in spec page".
   - Retire `ui/raid`.
7. **Member value and the comp search** (see [Comp search](#comp-search)).

None of this changes a DPS Report number.

## Open decisions

1. **Healers' and tanks' damage.** When they're simmed for real, does the page also need a simple
   stand-in until then, such as a set DPS per tank and healer slot? The stand-in would make the raid total
   closer to real, though it doesn't change which comp wins.
2. **The DPS choices for the comp search.** Should it choose among every Level 60 preset, the assignment
   presets included? Or only among builds the user picks?
3. **Limits per class.** Should the search respect something like "at most N of a class"? Or should it be
   free to fill a raid with the single best spec, if the buffs don't stop it?
4. **Rule-placed groups.** Should the search keep them fixed, or try other groups too?
