# Raid sim

A sim mode for a whole raid. Every member is simmed together in one fight, and the result is the raid's
total damage. The DPS and PvP modes each sim one player against an assumed raid. Here the other members are
real players.

The point is finding the best comp: which specs to bring, and how many of each, for a raid size with a set
number of tanks and healers.

The decisions marked "(the user, 2026-10-29)" are settled. Everything else is a proposal, and the
questions still open are at the end.

**History:**

- The first draft was written on 2026-10-29 from the user's idea in TASKS.md (2026-10-14).
- Four reviews followed the same day: Forever's buffs and debuffs, checked against the client tables; the
  engine; the comp search, with a working build and measurements; and the page.

## What it answers

- **The best comp** (the main question). Given a raid size and the number of tank and healer slots, it finds
  the DPS builds that give the most raid DPS (see [Comp search](#comp-search)).
- **Total raid DPS for a given comp.** The total counts pets.
- **What a spec is worth to a raid.** That is its own DPS plus what it does for everyone else (see
  [Member value](#member-value)). The DPS Report ranks specs on their own DPS, in a raid that already has
  every buff. It can't show that the first warrior brings Battle Shout and Sunder Armor, that the first
  two shamans bring Windfury and Grace of Air, or that a second warlock adds nothing while Faerie Fire is
  up.
- **Assignments.** Who keeps up each debuff, who gives which blessing, which caster gets Power Infusion,
  and what each assignment costs the player who holds it.
- **Each member's DPS in a real raid, next to their DPS Report number.**

## Forever changes the question

In Classic, raid comps were mostly about groups. Battle Shout, the totems, Trueshot Aura, Leader of the
Pack and Moonkin Aura reached only their own party, so the work was building a melee group around a shaman.

Client 70425 made almost all of those raid-wide. This was checked against the client tables of 70338 and
70685:

- Strength of Earth, Grace of Air, Windfury and Flametongue Totem, Moonkin Aura, Leader of the Pack and
  Trueshot Aura went from a party area aura to a raid area aura.
- Battle Shout and the hunter wolf's Furious Howl went from the caster's party to the caster's raid. The
  patch notes don't mention Furious Howl.
- The resistance totems and auras were raid-wide already.

What is still party-wide is mostly mana and survival:

- Mana Spring, Mana Tide, Healing Stream, Stoneskin, Windwall and Tranquil Air Totem
- Devotion, Retribution and Concentration Aura
- Blood Pact
- Dark Inspiration (a Shadow talent: +15% Spirit to the party)
- Vampiric Embrace and Aspect of the Pack

Forever made the old shared damage debuffs personal. Improved Scorch, Winter's Chill, Shadow Weaving,
Improved Shadow Bolt and Stormstrike each count only their caster's damage (aura 270 or 271), and so do the
new Haunt and Wrack.

Curse of the Elements now covers all six magic schools, Holy included, and Curse of Shadow is gone. So one
warlock's curse raises the damage of mages, shadow priests, Ret paladins, Elemental shamans, Balance druids
and the demons.

Faction doesn't limit the comp, because both sides have paladins and shamans (`sim/core/buffs.go`). And
Forever has no limit on how many buffs or debuffs a target can carry (the user, 2026-10-29).

So a Forever comp is mostly about **which specs, and how many of each**. What is left that is shared:

- raid-wide buffs that stop adding once someone gives them (one Battle Shout, one Moonkin Aura, one Leader
  of the Pack)
- buffs capped per provider (one air totem per shaman, one blessing per paladin for each class)
- the shared debuff slots (the minor armor slot holds Faerie Fire, Curse of Recklessness or a Tallstrider's
  Dust Cloud)

Group placement changes mana (Mana Spring, Mana Tide, Blood Pact, Dark Inspiration). It changes damage
only through Retribution Aura and Thorns in the tank's party, which do nothing until tanks are simmed.

**Range.** "Raid-wide" still means within 20 yards (Battle Shout), 30 yards (the totems) or 45 yards
(Moonkin, Leader of the Pack, Trueshot Aura). The raid sim assumes everyone is in range of everything (see
[Open decisions](#open-decisions)).

## The score

**Raid DPS: all damage the raid deals to the bosses, divided by the fight length.** Bigger is better.

The result also shows:

- each party's DPS and each member's DPS
- for each DPS member, the difference from their DPS Report build
- each buff and debuff's uptime, and who provided it

## The page

The raid sim is **a page of its own** (the user, 2026-10-29). It is not part of a spec's page or of the DPS
Report. It's linked from a "Raid sim" card in the landing page's Reports section, next to the DPS Report
(`ui/index.html`), and launched through `raidSimStatus` in `launched_sims.ts`. The sidebar's mode picker
doesn't fit: its modes are views of one page.

**The page sets up only the raid:**

- the raid size and how many slots are kept for tanks and healers
- which member goes in each slot, and which group they're in
- assignments
- the encounter
- the comp search

**It doesn't set up a member's gear, talents, rotation or consumes.** Each member comes whole from a spec,
and the raid stores a reference to it plus a snapshot to fall back on. There are three sources:

- **A preset marked for raid sims.** The reference is the build's id.
- **A saved setup on the spec's page.** This is a Preset Configuration, stored in
  `<spec prefix>__savedPresetConfigurations__`, and the reference is its name. When the name is renamed or
  deleted, the raid keeps the snapshot and says so.
- **The spec page's current settings** (`<spec prefix>__currentSettings__`). These are taken once, as a
  snapshot, since they change with every tweak on that page.

All spec pages are under `/forever/` on one host, so the raid page can read their storage. The prefixes
come from `specToLocalStorageKey`, typo included (`__classic_enhacement_shaman`).

**Turning a source into a member** happens without the spec's page:

1. Load the source with the spec's `migrateSettings`.
2. Build a headless `Player` from the spec's defaults plus the build: `applyDefaults` and `applyBuild`,
   moved out of `IndividualSimUI` and the preset picker into shared functions.
3. `toProto(false, true)`, which resolves Auto and Simple rotations through the spec's config.
4. Drop the source's own raid buffs, party buffs, debuffs, encounter and tanks.
5. In the player's own buffs, keep the world buffs and replace whatever members give (blessings,
   Innervate, Power Infusion) with what the resolver gives.

A preset's consumes and blessings live in its encounter preset (`PresetEncounter`). Take those, and
`inFrontOfTarget`, and drop the rest of that encounter.

**"Open in spec page"** opens the member in a new tab, in a raid member mode (for example
`<spec>/?raidMember=<raid>.<slot>#<link>`). In that mode the spec page:

- doesn't overwrite its own current settings
- shows a banner with "Save to raid"
- gets as its assumed raid everything this raid gives the member, at full uptime from the pull, with the
  member's own assignments switched off
- runs a fixed-length fight, the raid's average length when the raid fight is set by boss health

The spec page's settings are all on or off. So the spec page's sim is close to the member's raid number,
not equal to it: it can't show late debuffs, dropped totems or when Mana Tide goes down.

"Save to raid" writes the member's settings to a key the raid page owns. The raid page listens for the
browser's `storage` event and re-reads.

The reverse works too: an "Add to raid sim" action in a spec page's Preset Configurations adds a reference
to the raid.

**The upstream raid page is replaced.** `ui/raid` edits each member inside the raid page, takes its members
from the old `raidSimPresets` and is full of WotLK leftovers. The new page is a new folder, with a new
storage prefix.

**What it reuses:**

- the `SimUI` base and `addRaidSimAction`, which already shows raid DPS for more than one player
- `EmbeddedDetailedResults`, with its player filter
- `EncounterPicker`
- `Sim.runRaidSimRequest` and the worker pool
- `Raid` and `Party` as the request builder: set `numActiveParties` to 2, 4 or 8, and never send empty
  players for empty slots
- `SavedDataManager` for saved raids
- the old raid picker's drag and drop

`raidSimPresets` and the `isWithinRaidSim` branches go when `ui/raid` does.

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

**DPS members are builds marked for raid sims** (see [The page](#the-page) and
[The build pool](#the-build-pool)).

**Simming tanks and healers is out of scope for now** (the user, 2026-10-29). Their damage is left out,
with no stand-in. Tanks do deal damage, even when holding threat comes first, and healers can weave in
damage spells, so the total is somewhat short. That's the same for every comp with the same slot counts,
so comps still rank the same way.

Until then, a tank or healer slot holds a **placeholder**: a spec and what it gives. A placeholder is not a
unit in the engine, because a unit would join every list sized by the raid (attack tables, raid damage
hits, presims) and do nothing for DPS. The resolver writes what a placeholder gives into the members'
settings, and that is the placeholder's only effect.

With no tank unit, the boss attacks nobody, and every DPS member takes the raid damage hits, as in the
report. A test pins this.

What each placeholder gives, with a proposed default for each spec. Only effects that change DPS or mana
are listed:

| Placeholder | Gives (default) |
|---|---|
| Tank warrior | Sunder Armor; Battle Shout when no DPS warrior has it |
| Feral tank | Faerie Fire (Feral), Leader of the Pack |
| Protection paladin | one Greater Blessing for each class, one aura, optionally Judgement of the Crusader |
| Holy paladin | one Greater Blessing for each class, one aura, optionally Judgement of the Crusader |
| Healing priest (Discipline) | Fortitude, Divine Spirit, Power Infusion |
| Healing priest (Holy) | Fortitude, Divine Spirit |
| Restoration shaman | one totem per element (earth, air, water, fire), Mana Tide to its party |
| Restoration druid | Gift of the Wild, Innervate |

Divine Spirit is a trainer spell in Forever, not a talent.

The tanks' Demoralizing Shout, Demoralizing Roar and Thunder Clap change no DPS, so they're left out.

Each placeholder can be changed per slot: which blessings, which totems, whether it holds an assignment,
and Restorative Totems (Mana Spring) or Guardian Totems. Those two are the only "improved" talents Forever
still has for these buffs. Improved Battle Shout, Enhancing Totems, Improved Blessing of Wisdom, Improved
Devotion Aura and Improved Mark of the Wild are gone, and Improved Imp no longer raises Blood Pact.

## Buffs and debuffs come from the members

Nothing is assumed. **The resolver fills the settings with what the members give.** Each buff and debuff
arrives in one of two ways, never both:

- **Given.** The resolver writes it into the settings, so it's up for the whole fight. This covers:
  - the long buffs, which last 60 minutes in Forever and go on before the pull (Fortitude, Arcane
    Brilliance, Gift of the Wild, Divine Spirit, the blessings, the auras)
  - the permanent ones (Moonkin Aura, Leader of the Pack, Trueshot Aura)
  - everything a placeholder gives
- **Cast.** A DPS member's rotation casts it as their assignment, and the engine sends it to its scope.
  This covers Battle Shout, Sunder Armor, Faerie Fire, the curses, Judgement of the Crusader, Hunter's
  Mark and the shamans' own totems. A dropped totem or a late debuff then costs the whole raid, as it
  would in game.

Never both matters for more than tidiness:

- A given Battle Shout or curse would make the assigned member's rotation see it already up and skip its
  cast.
- A given Windfury Totem and a cast one are separate auras with separate cooldowns, so together they would
  double the procs.

A test proves the resolver keeps them apart.

**Members get their own buff set.** Today one `RaidBuffs` covers the whole raid, and `PartyBuffs` holds
only Atiesh and Mana Tide. The report's requests already differ per member, not just per party:

- The casters and hunters have no Strength of Earth or Windfury.
- The melee have no Mana Spring.
- Enhancement gets Flametongue in place of Windfury.

So the engine gets an optional `RaidBuffs` per player, which `applyCharacterEffects` uses in place of the
raid's. The resolver does all the scoping, and the engine stays small. The same field serves the report
raid check and "Open in spec page".

**Pets.** Player buffs never reach pets in Forever (standard 14), but target debuffs do. Armor debuffs
raise every pet's damage, and Curse of the Elements raises the demons' spells. Since pet damage counts
in raid DPS, the debuffs are valued with the pets included.

### The provider table

The resolver works from a table that says, for each buff and debuff:

- who gives it
- its scope (raid, party, a class across the raid, one target, or the boss)
- whether it reaches pets
- its radius
- whether several providers stack
- what it excludes
- how long it lasts and whether it costs GCDs in the fight

**Scope comes from the client tables, generated, with a hand-written part:**

| How the spell is built | Scope |
|---|---|
| Persistent area aura (totem buffs, paladin auras, Trueshot Aura, Moonkin Aura, Leader of the Pack, Blood Pact) | SpellEffect's Effect: 35 is the party, 65 is the raid |
| Cast group buff (Battle Shout, Furious Howl, Gift of the Wild, Arcane Brilliance, Prayer of Fortitude, the Greater Blessings, Dark Inspiration); Effect is 6, apply aura | `ImplicitTarget_0`: 20 is the caster's party, 56 is the caster's raid, 153 is the raid members of the target's class, 21 or 57 is one target |

- **Pets:** SpellMisc `Attributes_3 & 0x100` marks players only.
- **Radius:** SpellRadius.
- **Totems (written by hand):** a totem summon (Effect 28) names a creature, and the client has no link
  from it to the buff spell the totem casts. So the map from totem to buff is written by hand.

DPS-relevant buffs, checked against client 70685 (rows from
`wago.tools/db2/SpellEffect?build=1.60.1.70685`; "unverified" means the client can't say and the server
decides):

| Buff | Given by | Scope | Several providers | Excludes | Lasts / in-fight cost |
|---|---|---|---|---|---|
| Battle Shout (melee AP, rank 6 at launch) | any warrior | raid, 20 yd, players | strongest counts (unverified) | Furious Howl (sim rule) | 3 min: refreshed once in fights over about 177 s |
| Furious Howl (82 melee AP) | hunter with a wolf | raid, 15 yd | unverified | Battle Shout (sim rule) | the pet's: 60 s, 30 s cooldown |
| Strength of Earth Totem (rank 4) | any shaman, earth | raid, 30 yd, players | doesn't stack (sim); unverified (shaman_audit 8.46) | the shaman's other earth totems | 5 min, 1 s GCD |
| Grace of Air Totem (rank 2) | any shaman, air | raid, 30 yd, players | doesn't stack | Windfury on the same shaman | 5 min |
| Windfury Totem | any shaman, air | raid, 30 yd, players, main hand only; not for a shaman with Windfury Weapon | doesn't stack | Flametongue Totem; Grace of Air on the same shaman | 5 min; 1.5 s cooldown per receiver |
| Flametongue Totem | any shaman, fire | raid, 30 yd, players | doesn't stack | Windfury Totem; Flametongue Weapon | 5 min; only helps Enhancement (standard 25) |
| Trueshot Aura (ranged AP) | Marksmanship | raid, 45 yd; only hunters gain | doesn't stack | | permanent |
| Moonkin Aura (3% spell crit) | Balance | raid, 45 yd, players | doesn't stack | (stacks with Leader of the Pack now) | permanent in form |
| Leader of the Pack (3% physical crit) | Feral | raid, 45 yd, players | doesn't stack | | permanent in form |
| Greater Blessing of Kings, Might, Wisdom | any paladin (Kings is baseline) | one class across the raid, players | one blessing per paladin for each player | the paladin's other blessings on that class | 60 min |
| Gift of the Wild | any druid | raid, players | doesn't stack | | 60 min |
| Arcane Brilliance, Divine Spirit (mana) | any mage, any priest | raid | doesn't stack | | 60 min |
| Mana Spring Totem | any shaman, water (Restorative Totems) | party, 30 yd | open (shaman_audit 8.45); the sim adds both today, which likely overstates a second shaman | the shaman's other water totems | 5 min |
| Mana Tide Totem | shaman talent | party | unverified | | 13 s, 5 min cooldown: an assignment of when |
| Dark Inspiration (+15% Spirit) | Shadow talent | party | unverified | | procs on its own, 10 s |
| Power Infusion (+20% spell damage) | Discipline, 31 points | one target | unverified | | 15 s, 3 min cooldown, off the GCD: an assignment of who |
| Innervate | any druid | one target | unverified | | 20 s, 6 min cooldown: an assignment of who |
| Retribution Aura, Thorns | paladin, druid | the tank's party, the tank | | | only when the boss hits a member: ignored until tanks are simmed |

Salvation and Mass Salvation are threat only, so they add no DPS while threat is out of scope. The
following add no DPS either:

- Fortitude, Shadow Protection and Blood Pact. The raid damage hits ignore resistances.
- Stoneskin, Windwall, Tranquil Air, Healing Stream, the resistance totems and auras, Devotion and
  Concentration Aura.

The debuffs:

| Debuff | Given by | Several providers | Excludes | Lasts / in-fight cost |
|---|---|---|---|---|
| Sunder Armor (450 a stack) | any warrior | stacks to 5 from any warrior | Expose Armor (standard 18) | 30 s, a GCD each |
| Faerie Fire / Faerie Fire (Feral) (505 armor) | any druid | one per target, from any druid | the minor armor slot | 40 s |
| Dust Cloud (505 armor) | hunter with a Tallstrider | one | the minor armor slot | the pet's, 30 s |
| Curse of Recklessness | warlock | doesn't stack (70425) | the minor armor slot; the warlock's other curse | 2 min |
| Curse of the Elements (+10%, all magic schools) | warlock | doesn't stack | the warlock's other curse, not its bane | 5 min |
| Judgement of the Crusader (+161 Holy) | paladin | one (unverified) | | 40 s; about 2 GCDs to seal and re-seal |
| Judgement of Wisdom | paladin | one | | 40 s; mana for every attacker with mana |
| Hunter's Mark (ranged AP for hunters, not pets) | hunter | one per target (unverified) | | 2 min, a GCD |

Forever splits a warlock's curses from its banes (Agony, Doom, Havoc are banes; warlock.md). So a curse
assignment costs GCDs, not the warlock's Agony.

Insect Swarm, Scorpid Sting, Demoralizing Shout and Roar, Thunder Clap and Curse of Weakness only lower the
boss's damage. They add no DPS and are left out.

**Rules the resolver enforces:**

- **One per provider.**
  - A paladin gives each player one blessing. Greater Blessings cover a class, so the assignment is a
    paladin and a class.
  - A shaman has one totem per element. Grace of Air and Windfury together need two shamans.
  - A warlock has one curse.
- **Shared slots.**
  - The minor armor slot takes one of Faerie Fire, Curse of Recklessness and Dust Cloud (standard 20).
  - Windfury and Flametongue Totem exclude each other.
- **Nobody provides it.** If no member gives a buff, nobody has it. When three paladins can't cover Kings,
  Might and Wisdom for every class, the result shows who misses out.
- **What stacks from several providers** follows the tables above. Where the client can't say (Mana
  Spring, the other totems, Hunter's Mark, Judgement of the Crusader, Furious Howl against Battle Shout),
  the raid sim uses the sim's current rule and lists it as an assumption. The beta tests that settle
  these go to notes.md.

## Assignments

Each assignment is **automatic, or set by hand** (the user, 2026-10-29). Automatic assignment gives it to
the member it costs least, and the page shows who that is. There are three kinds:

- **Upkeep in the fight**, which costs the holder GCDs:
  - Sunder Armor (usually a tank placeholder)
  - Faerie Fire, Curse of Recklessness or Dust Cloud for the minor armor slot
  - Curse of the Elements
  - Judgement of the Crusader
  - Hunter's Mark
  - Battle Shout
  - the totems a DPS shaman drops

  The 5-minute totems and Curse of the Elements cost almost nothing until fights pass 300 seconds.
- **Before the pull**, which costs nothing in the fight:
  - which Greater Blessing each paladin gives each class
  - which aura each paladin holds
  - which totem each shaman holds in each element
- **Cooldowns given to someone else:**
  - Power Infusion: which caster, and when
  - Innervate: which mana user
  - Mana Tide: when, and it only reaches its party

A DPS member holding an upkeep assignment runs their build's assignment rotation, which is the report's
duty preset for that build (Battle Shout, the curse, Faerie Fire, Judgement of the Crusader). Every other
member runs the base rotation. When nobody holds an assignment, the raid goes without it.

Two members of the same build therefore run different rotations: the warrior on Battle Shout shouts, and
the others never do.

## Groups

On the page, the user places groups by dragging members. A new raid starts from a rule: casters go with
the shamans that keep up Mana Spring and Mana Tide.

The comp search tries whatever groups give the most raid DPS (the user, 2026-10-29). In Forever only
Mana Spring, Mana Tide, Blood Pact and Dark Inspiration depend on groups, so the search only moves the
members those touch.

## The fight

- **The bosses.** By default it's the DPS Report's generic level 63 boss (3731 armor, Unknown creature
  type). Forever's raid bosses aren't known yet. When they are, each one is an encounter.
- **Several bosses** (the user, 2026-10-29: "or multiple if multiple exist"). The engine needs work for
  this (groundwork step 7):
  - Every member starts on boss 0.
  - Many settings debuffs go on boss 0 only.
  - A health fight ends on the bosses' summed health.
  - The execute phases follow the raid's total damage.
  - The raid damage hits come from boss 0 even after it dies.

  The first version has one boss.
- **Length** (the user, 2026-10-29). It can be set two ways:
  - A fixed duration, with the report's 180 seconds as the default.
  - The bosses' health. The fight ends when every boss is dead, each one dying on its own
    (`sim/core/target_death.go`, which works for any health fight).
- **Long fights.** Past about 177 seconds Battle Shout needs a refresh, and past 300 seconds so do the
  totems and Curse of the Elements. The 30 to 40 second debuffs are the real upkeep at any length.
- **Raid damage.** Every DPS member takes the encounter's raid damage hits, as in the report (2 a minute at
  1500, standard 24). They feed procs and Rage, since nobody heals.
- **Threat is out of scope for now** (the user, 2026-10-29). When it's added, it is reported but not
  modelled:
  - a warning whenever the tank couldn't hold threat, saying when and against whom
  - a results page that drills down into each member's threat on each boss
  - nothing holds a member's damage back

  Every member's threat is already measured (`UnitMetrics.threat`).

## Results

**Already there:**

- `addRaidSimAction` shows raid DPS and its standard deviation.
- `EmbeddedDetailedResults` has a player filter, and its aura, cast, damage, timeline and replay views
  work once a member is picked.
- `PlayerDamageMetricsTable` shows each member's share.
- The typical, lucky and unlucky iterations are already picked by raid DPS.

**New:**

- **A 95% interval** for raid DPS.
- **A party table.** `PartyMetrics` has the numbers, but nothing shows them.
- **A member table** with each member's difference from their report build. The report's numbers come from
  `ui/dps_report/data/builds.json`, by build id. The difference is shown only for an unchanged marked
  preset, at 120, 180 or 300 seconds, against a boss the report also sims. It is tested against the
  report's standard error.
- **Each buff and debuff** with its provider, its scope and its uptime. `AuraMetrics` has no provider, so
  the provider comes from the resolver's record.
- **Members who miss a buff** the raid could have given them.
- **Assignment cost and member value**, as buttons, since each takes extra sims.
- **A raid history and saved raids of their own.** `SimHistory` and `CompareResults` are built around one
  player.

If a change touches shared result views (events view, replay), the DPS Report is deployed after it
(CLAUDE.md).

## Member value

"A filler in their place" was undefined. There are three numbers, each simmed with common random numbers,
once per distinct build:

- **Contribution:** raid DPS minus raid DPS with the member's slot left empty. It's shown as the member's
  own DPS plus what they add to everyone else. This is "what a spec brings".
- **Replacement margin:** raid DPS minus the best raid DPS with any other build in that slot. The comp
  search computes it as it goes. At the best comp it is zero or more, and a margin within noise means two
  builds are interchangeable. It's shown with the runner-up builds.
- **Value of adding** a spec that isn't in the comp: the best raid DPS from swapping it in for any member,
  minus raid DPS.

**Assignment cost** is the holder's DPS with the assignment minus their DPS without it, in the same raid
with the effect given. The raid DPS the assignment adds is part of the holder's contribution.

## Checking it

1. **The report raid.** A raid of base report builds, with no duty or encounter variants:
   - each member gets their own request's raid buffs (the per-player field) and the base requests' common
     debuffs
   - a fixed 180 seconds
   - every member within 3 standard errors of a single-player sim of the same request, on the same commit,
     at the same seed and iterations

   It compares with the single-player sim, not the published numbers, which come from the report's merged
   tree with its own patches.
2. **Doubles.** Every base build twice in one raid: the second member within noise of the build alone.
3. **Size.** Raids of 20 and 40 build and run without a panic, a 40-player health fight included, and a
   party with gaps in slots 1 and 3 gives everyone the right setup.
4. **Scope.** A member gets a party buff only from their own party and a raid buff from anyone. A cast
   Battle Shout reaches the raid, a dropped totem reaches its scope, and given and cast never both apply.
5. **After 2026-12-09**, the user's raid logs (`tools/warcraftlogs`, `logcheck`) and public logs are
   checked against the sim: raid DPS by comp and buff uptimes.

## The build pool

**The comp search's DPS pool is the presets marked for raid sims, for now the DPS Report's builds** (the
user, 2026-10-29).

Today nothing links the UI's presets to the report's builds:

- The report's requests come from the Go tool `tools/rotopt/mksettings`, with its own copy of the consumes
  and hand edits listed in each `meta.json`.
- The names differ ("Fury Dual Wield" against "Level 60 Fury").
- 12 builds already differ: the warlocks' and shadow priests' distance and channel clip, and the Survival
  hunter's distance.
- Enhancement's Imp GW build has no UI preset.

So:

- **The mark lives on `PresetBuild`:** `raidSim: { reportId, base?, assignment?, mobType? }`.
  `makePresetBuild` is fixed to pass it through. Today it drops fields it doesn't know, the leveling plan
  among them.
- **A parity check:** each marked build, on its spec's defaults, equals the player in
  `tools/dps_report/req/<reportId>.json`, apart from a short allowed list. The 12 differences are settled
  or allowed, and Imp GW gets a preset.
- **Assignment presets are not entries of their own.** `assignment` and `base` (the report's `duty` and
  `baseBuild`) make them the base build holding an assignment.
- **The encounter variants** (vs Humanoid, Undead, Dragonkin) join the pool only when the boss's creature
  type matches.
- **The "+ Mana Tide", Water Shield and Imp GW builds** are different talent builds, so they are entries
  of their own.
- **Pet families are a comp choice.** A Tallstrider fills the minor armor slot and a wolf gives melee
  attack power raid-wide. So these may need marked variants of the hunter builds (see
  [Open decisions](#open-decisions)).

The page builds the pool's players from the marked presets. The Go tests read the same builds from
`tools/dps_report/req`, and the parity check keeps the two equal.

## Comp search

The main thing the raid sim is for (the user, 2026-10-29).

**The inputs:**

- the raid size, and the number of tank and healer slots
- the tank and healer specs, or "any"
- the encounter

**There are no limits per spec or class** (the user, 2026-10-29). The goal is the most raid DPS and
nothing else. The search may fill the raid with one spec, and may drop the only provider of a buff when a
stronger spec in its place gives more.

### The size of it

For a generic boss the pool has 37 builds: 58 requests, less 12 duty variants and 9 encounter variants.
About 20 are left after dropping builds that give the same buffs as a better one.

The number of distinct comps:

| Builds | 7 DPS | 14 DPS | 29 DPS |
|---|---|---|---|
| 37 | 3.2e7 | 9.4e11 | 2.5e18 |
| 20 | 6.6e5 | 8.2e8 | 1.2e13 |

The first draft's greedy fill and swaps take about 1,100, 2,000 and 3,500 raid sims for 10, 20 and 40.
At today's speed that is 5 hours for 10 and 6 days for 40 on one thread. The plan below takes about
150 to 200.

### The search

1. **A model of each build's DPS** as a function of which buffs and debuffs are up. Each buff is a role,
   for example "the minor armor slot", "the first and second air totems", or blessings by count. The
   model is log-linear with pairwise terms between roles. It's fitted from about 32 designed
   single-player sims a build, at 2,000 iterations, in the report's pipeline: about 16 minutes native for
   the whole pool. It ships as data next to `builds.json`, and the duty variants give each assignment's
   cost.

   A one-buff-at-a-time model isn't enough. Removing Faerie Fire alone costs a Fury warrior nothing,
   because Curse of Recklessness takes its place, and removing everything is 21% off in the additive
   model.
2. **A shortlist from the model** (milliseconds):
   - Count how many of each provider role are worth having, up to where one more adds nothing.
   - Fill the rest of the raid with the best builds under that buff set.
   - Keep the top 20 distinct comps.
3. **Confirm with raid sims.** Sim the 20 at 500 iterations. Correct the model with each build's
   difference (raid against model). Repeat 2 to 3 times: about 50 raid sims.
4. **Polish with raid sims** at 300 to 500 iterations:
   - swap each distinct build in the comp for each of the 5 best replacements (about 50)
   - swap provider pairs, such as two shamans in or out (about 10)
   - move assignments (about 15)
   - move members between groups, only those Mana Spring, Mana Tide, Blood Pact and Dark Inspiration touch
     (about 10)

   Every comp tried is remembered, so none is simmed twice.
5. **Decide.** Sim the top 3 paired, at 5,000 iterations each. The winner is published at 20,000
   (standard 15).

**The result:**

- the best comp, with its raid DPS and interval
- the runners-up, and how far each one is behind
- each member's contribution and replacement margin

### Noise

Every member draws from one random stream (`sim.go`), so swapping one member changes every other
member's draws. The same seed then helps little.

Measured on a 20-DPS raid, 1000 iterations, one member swapped:

| Random streams | How closely the two raids' DPS move together | Spread of the difference |
|---|---|---|
| independent seeds | 0.02 | 285 DPS |
| the same seed, today | 0.22 | 255 |
| per call site (the labeled streams) | 0.55 | 202 |
| per member, a rough prototype | 0.77 to 0.80 | 135 to 148 |
| only the swapped members' own noise (the ideal) | | about 70 |

**Per-member random streams** (groundwork step 6):

- Each member's spells, dots and swings draw from a stream seeded by the iteration, the build and the
  copy, not the slot, so moving a member between groups keeps their stream.
- The raid damage hits are keyed by the player who takes them.
- They are behind a `SimOptions` flag that only raid sims and the search turn on, so no report number
  moves.
- They cost 1 to 2% in speed once the streams are set up ahead.

Iterations needed to tell two comps apart, at 2σ, paired, on a raid of 18,500 DPS:

| Difference | Today | Per member (prototype) | Per member (ideal) |
|---|---|---|---|
| 0.1% (18 DPS) | 760 | 229 | 57 |
| 0.5% | 31 | 9 | 3 |
| 10 DPS | 2,600 | 780 | 200 |

### Where it runs

One request uses one core: the worker pool sends a whole request to one worker, and Go's WebAssembly is
single threaded. So the search is a **stateless Go step function**, `compSearchStep(state, results) →
(state, requests)`.

- **In the browser:** the page fans each batch of requests out over the workers, and reports progress
  (`ProgressMetrics`) with the best comp so far.
- **Abort** cancels every request in flight.
- **Natively:** a command-line run or the local server loops the same function with goroutines.

Wall clock for the whole search, about 100,000 raid iterations:

| Engine | 10 | 20 | 40 |
|---|---|---|---|
| Today, native, one thread | about 9 min | about 25 min | about 90 min |
| After the speed fixes | about 5 min | about 10 min | about 20 min |
| After the fixes, 4 browser workers (WebAssembly about 2× slower, estimated) | about 2.5 min | about 5 min | about 10 min |

## What the engine has today

A raid sim already exists, in two parts, both from upstream wowsims: the engine, which this fork uses for
every sim, and the `ui/raid` page, which is hidden and stale (see [The page](#the-page)). Every sim here
is a raid sim of one player plus the assumed buffs, but this fork has never simmed more than one player.
The reviews built it and ran raids of report builds.

**What works:**

- `RunRaidSim` builds and runs a raid in parties of 5.
- Metrics add up per member, per party and for the raid, pets included (pet damage goes to its owner).
- Iterations are split across CPUs.
- Each caster's DoTs are its own.
- Shared debuffs dedupe by label as in game: Sunder, Faerie Fire and Curse of Recklessness, Curse of the
  Elements.
- Windfury Totem's cooldown is per receiver.
- Healing Stream and Mana Tide already reach the party.
- Members who don't interact land within ±2.3 DPS of their build alone.

**What breaks** (all measured):

- **More than 200 auras on the boss panics** (`aura.go:461`). Every caster registers its own per-rank DoT
  and debuff auras on the boss, about 33 for a fire mage. 6 fire mages fail, a mixed 20-player raid of
  report builds fails, and 40 Beast Mastery hunters fail. A 10-player raid works.
- **Presims for 25 players.** `presim.go` indexes a 25-slot list by player index, so a 40-player health
  fight panics.
- **Gaps in a party.**
  - `applyCharacterEffects` and the rotation setup look players up by position in the shortened list.
    With a fire mage in slot 1 and a Fury warrior in slot 3, the mage did 0 DPS and the warrior ran the
    mage's rotation for 358.
  - A missing party shifts everything after it.
  - A nil player panics.
- **Two of one build don't work.**
  - APL references to target auras find the first aura with that spell id (`apl_helpers.go:118`). Fire
    Vulnerability, Winter's Chill, Shadow Weaving and Light's Vigil have a label per caster but share an
    id, so a second caster reads the first one's.
  - Doubled into one raid, the second member loses DPS: fire mage −27, frostfire mage −26, Shock paladin
    −7 (−14 in its duty and Undead variants). Every other base build is within ±2.5.
  - Two Enhancement shamans panic: Stormstrike's personal aura has a fixed label
    (`shaman/stormstrike.go:46`).
  - The last paladin to register Judgement of the Crusader overwrites the others' refresh hooks
    (`aura.go:493-510`). `NewExclusiveEffect` keeps the first caster's values.
- **Raid buffs and cast buffs.**
  - `GetRaidBuffs` writes into the request proto, so a search must copy every candidate.
  - `AddPartyBuffs` runs twice, so Atiesh counts double.
  - Blood Pact goes raid-wide through its hook, though Forever's is a party buff.
  - Divine Spirit is given whether or not it's talented. Either is fine in Forever, but it shows the hooks
    aren't checked.
  - A warrior's Battle Shout reaches only their own party. When two warriors shout, each receiver keeps the
    first one's value.
  - A shaman's Strength of Earth, Grace of Air, Windfury Totem and Mana Spring reach only the shaman. The
    weapon totem rules (Windfury or Flametongue, the imbue check, a totem coming back) live on the shaman,
    for the shaman alone.
  - When one of two shamans' identical totems ends, the buff comes off everyone, though the other totem
    still stands.
- **One-player assumptions.**
  - Stat weights, the bulk sim, `sim/lib`, the test suite, the Vaelastrasz encounter and the external
    Improved Shadow Bolt and Stormstrike debuffs all read `Parties[0].Players[0]`.
  - `spell_result.go` reads attack tables by target index.
- **Shared buff sets.** With one raid buff set for everyone, members were off their solo DPS well beyond
  noise:
  - Enhancement −36
  - feral −26
  - Seal Fate rogue −24
  - Survival −23
  - Ret −20
  - the Faerie Fire duty Balance druid +30

  This is partly because the report gives each build its own buffs (the per-player field fixes that) and
  partly the problems above.

**Speed.** On one thread, native, with the aura cap raised:

| Raid | Per iteration | Against the same builds simmed alone |
|---|---|---|
| 1 player | 0.10 to 0.93 ms (mean 0.37) | |
| 10 | 7.0 ms | 1.9× |
| 20 | 19.2 ms | 2.6× |
| 40 | 63.5 ms | 4.3× |

It grows as about 0.34·n + 0.031·n² ms. At 40 the n² part is 80% of the cost.

The 40-player profile, biggest first:

1. **Spell metrics** sized by every unit, reset and scanned every iteration (`spell.go:449`): about 25%.
2. **Every mana bar settled** on every time step (`sim.go:807`): 15%.
3. **The idle-rotation shortcut is dead in a raid.** It needs no event anywhere in the sim since the last
   check (`apl_idle.go`), so every idle rotation re-runs its whole priority list every 50 ms: about 18%.
4. **The pending-action list** is a linear sorted insert (`sim.go:858`): 12%.
5. **Weapon attacks** are all looped on each swing: 11%.

Removing the n² parts should bring 40 players to about 14 ms an iteration and 20 to about 7.5 ms.

## Groundwork

In order. Each step is usable on its own.

1. **A raid that builds and runs right.** A Go test, `sim/raid_forever_test.go`, loads the base requests
   from `tools/dps_report/req` and runs checks 1 to 3. The fixes:
   - the aura cap
   - the presim list
   - `PartyIndex` and the missing-party index
   - copying in `GetRaidBuffs`, and one `AddPartyBuffs`
   - a member's own auras found by owner: APL references prefer the rotation's own unit's aura, personal
     debuffs are keyed by caster, Stormstrike's label, and shared debuffs don't capture their caster
   - `RaidBuffs` per player
2. **Speed.** A benchmark for raids of 10, 20 and 40, then the fixes in profile order.
3. **Cast buffs reach their scope.**
   - Battle Shout to the raid.
   - The shaman's totems to the raid or party.
   - The weapon totem rules moved into core, per receiver.
   - A count of standing totems per receiver.
   - Blood Pact to the party.
   - Check 4.
4. **The provider table and the resolver.**
   - The generated scope table from SpellEffect, ImplicitTarget, SpellMisc and SpellRadius, plus the
     hand-written totem map.
   - The resolver turns a comp into per-player raid buffs and the debuffs, enforces the rules, makes
     automatic assignments, and records who gave what.
5. **The comp proto and its Go entry point.**
   - Members as build references, placeholders, groups, assignments and the encounter.
   - One Go call that resolves a comp and runs it, used by the tests, the page and the search.
6. **Per-member random streams**, behind a `SimOptions` flag.
7. **Several bosses and tanks.**
   - Which boss each member attacks.
   - Per-boss debuffs and execute phases.
   - Raid damage from living bosses.
8. **The build pool.** The `raidSim` mark, `makePresetBuild` fixed, the parity check against `req/`, and
   the Imp GW preset.
9. **The raid page**, in order:
   1. A headless member builder.
   2. The page shell: raid size and slot counts.
   3. Slots and groups with drag and drop.
   4. The member picker and placeholders.
   5. Assignments and the encounter.
   6. Run and results.
   7. Raid history, saved raids and share links.
   8. "Open in spec page" with "Save to raid", and "Add to raid sim".
   9. The landing card and launch.
   10. Retire `ui/raid` and `raidSimPresets`.
10. **The search model.** The designed single-player sims in the report pipeline, and the fitted model as
    data.
11. **The comp search and member value.** The step function, the browser fan-out and the native loop.

None of steps 1 to 8 changes a DPS Report number. Steps 10 and 11 add data to the report pipeline
without changing its numbers.

## Open decisions

The four from the first round are settled (the user, 2026-10-29):

- Tank and healer damage is left out for now.
- The search's DPS pool is the presets marked for raid sims, for now the DPS Report's builds.
- There are no limits per spec or class.
- The search tries whatever groups and assignments give the most raid DPS.

New, from the reviews:

1. **Range.** Should the sim assume everyone is within reach of every raid-wide buff? Battle Shout reaches
   20 yards, the totems 30. Proposed: yes. The alternative is a melee group and a ranged group, with each
   buff reaching the groups it would in game.
2. **Hunter pet families.** A Tallstrider's Dust Cloud fills the minor armor slot, and a wolf's Furious
   Howl gives raid-wide melee attack power. Should the report get marked hunter variants with those pets,
   so the search can use them? Today the hunter builds use their best pet for their own DPS.
3. **Where big searches run.** After the speed fixes, a 40-player search takes about 10 minutes in the
   browser. Is that fine, or should the 10, 20 and 40 default searches run natively and ship with the page
   as data, like the report's numbers, with the page running only custom searches?
