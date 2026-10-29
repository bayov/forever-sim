# Raid sim

A sim mode for a whole raid. Every member is simmed together in one fight, and the result is the raid's
total damage. The DPS and PvP modes each sim one player against an assumed raid. Here the other members are
real players.

The point is finding the best comp: which specs to bring, and how many of each, for a raid size with a set
number of tanks and healers.

**How to read this doc:**

- The decisions marked "(the user, 2026-10-29)" are settled, and [Decisions](#decisions) lists them.
- Everything else is a proposal.
- The measurements behind the design are in the [Appendix](#appendix-measurements-2026-10-29).
- An implementer starts at [Milestones](#milestones), which says which sections each step needs.

**History:** first drafted on 2026-10-29 from TASKS.md's 2026-10-14 idea. It was reviewed three times the
same day:

- Forever's buffs against client 70685
- the engine
- the search, against real raids
- the page
- the doc as a spec
- scope
- the bug fixes, done in a scratch copy

## What it answers

- **The best comp** (the main question). Given a raid size and the number of tank and healer slots, it finds
  the DPS builds that give the most raid DPS. The total counts pets.
- **The best tank and healer presets** for the raid's damage, leaving out utility, threat and healing.
- **What each buff role is worth.** For example, what the second shaman's air totem adds to the raid, and
  what it costs the shaman who drops it.
- **What a spec is worth to a raid.** That is its own DPS plus what it gives everyone else. The DPS Report
  ranks specs on their own DPS, in a raid that already has every buff. It can't show that the first two
  shamans bring Windfury and Grace of Air, or that a second warlock adds nothing while Faerie Fire is up.
- **Assignments.** Who keeps up each debuff, who gives which blessing, which totems each shaman drops, and
  what each assignment costs the player who holds it.
- **Each member's DPS in a real raid**, next to their DPS Report number.

## Forever changes the question

In Classic, raid comps were mostly about groups. Battle Shout, the totems, Trueshot Aura, Leader of the
Pack and Moonkin Aura reached only their own party, so the work was building a melee group around a shaman.

**Client 70425 made almost all of those raid-wide.** This was checked against the client tables of 70338
and 70685:

- Strength of Earth, Grace of Air, Windfury and Flametongue Totem, Moonkin Aura, Leader of the Pack and
  Trueshot Aura went from a party area aura to a raid area aura.
- Battle Shout and the hunter wolf's Furious Howl went from the caster's party to the caster's raid. The
  patch notes don't mention Furious Howl.
- The resistance totems and auras were raid-wide already.

**What is still party-wide** is mostly mana and survival:

- Mana Spring, Mana Tide, Healing Stream, Stoneskin, Windwall and Tranquil Air Totem
- Devotion, Retribution and Concentration Aura
- Blood Pact
- Dark Inspiration (a Shadow talent: +15% Spirit to the party)
- Vampiric Embrace and Aspect of the Pack

**Forever made the old shared damage debuffs personal.** Improved Scorch, Winter's Chill, Shadow Weaving,
Improved Shadow Bolt and Stormstrike each count only their caster's damage (aura 270 or 271), and so do the
new Haunt and Wrack.

**Curse of the Elements covers all six magic schools,** Holy included, and Curse of Shadow is gone. So one
warlock's curse raises:

- the casters: mages, shadow priests, Ret paladins, Elemental shamans and Balance druids
- the demons
- the hunters' Arcane Shot and Serpent Sting
- the rogues' poisons
- Enhancement's shocks and Flametongue
- Wind Serpent pets

**Forever has no limit** on how many buffs or debuffs a target can carry (the user, 2026-10-29).

**Faction.** Both sides have paladins and shamans, so faction doesn't limit the classes. But Forever has no
cross-faction grouping (`forever-wiki/overview.md`), so a raid's races come from one faction. The search
runs once for each faction. Each build takes its best race in that faction, from the report's race tables
under the full buff set. That's an assumption: in a raid short on mana, another race could overtake.

**So a Forever comp is mostly about which specs, and how many of each.** What is left that is shared:

- raid-wide buffs that stop adding once someone gives them (one Battle Shout, one Moonkin Aura, one Leader
  of the Pack)
- buffs capped per provider (one air totem per shaman, one blessing per paladin for each player, one
  judgement per paladin)
- the shared debuff slots (the minor armor slot holds Faerie Fire, Curse of Recklessness or a Tallstrider's
  Dust Cloud)

**Group placement** changes mana: Mana Spring, Mana Tide and Dark Inspiration. It changes damage only
through Retribution Aura and Thorns in the tank's party, which do nothing until tanks are simmed.

**Range.** "Raid-wide" still means within 15 yards (Furious Howl), 20 yards (Battle Shout), 30 yards (the
totems) or 45 yards (Moonkin, Leader of the Pack, Trueshot Aura). The raid sim assumes everyone is in range
of everything (the user, 2026-10-29).

## The score

**Raid DPS: all damage the raid deals to the bosses, divided by the fight length.** Bigger is better.

The result also shows:

- each party's DPS and each member's DPS
- for each DPS member, the difference from their DPS Report build
- each buff and debuff's uptime, and who provided it

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

**DPS members are report builds** (see [The build pool](#the-build-pool)).

**Tanks and healers are placeholders.** Simming them is out of scope for now: no damage, no threat, no
healing (the user, 2026-10-29). Their damage is left out with no stand-in. That's the same for every comp
with the same slot counts, so comps still rank the same way.

But a placeholder is a full provider (the user, 2026-10-29). Whatever its preset gives goes to the raid,
and the search counts it as covered at no DPS cost.

**A placeholder is not a unit in the engine.** A unit would join every list sized by the raid (attack
tables, raid damage hits, presims) and do nothing for DPS. The resolver writes what a placeholder gives
into the members' settings, and that is its only effect.

With no tank unit, the boss attacks nobody, and every DPS member takes the raid damage hits, as in the
report.

### Tank and healer presets

**A placeholder is a preset: a spec with a fixed talent build.** What it gives the raid follows from that
build, so Power Infusion comes only with a Discipline priest.

**Every preset is a sensible raid build** (the user, 2026-10-29). It takes the core tanking or healing
talents a real raid tank or healer would. Threat and healing aren't scored yet, so nothing else stops a
build from trading its role talents for raid buffs. The presets are what keep the sim realistic.

So the presets are a short, hand-picked list. The search chooses among them for each slot and never builds
a tank's or healer's talents itself. When threat and healing are simmed, their scores keep builds realistic
on their own, and the list stays the starting point.

The presets were checked against client 70630's trees (unchanged in 70685) and 70685's spells. Every string
is valid: 51 points, with rows and prerequisites right. The talent strings are in the sim's per-tree digit
format:

| Preset | Talents | Gives |
|---|---|---|
| Protection warrior | 8/0/43, `35--251533120331221351` (`ui/tank_warrior/presets.ts`) | Sunder Armor upkeep after the pull (see [Sunder Armor](#sunder-armor)); Battle Shout |
| Feral tank | 0/41/10, `-55220302020132012551-055` (`ui/feral_tank_druid/presets.ts`) | Leader of the Pack; Gift of the Wild; Faerie Fire as an assignment. Innervate is off by default, since it means leaving Bear Form. |
| Protection paladin | 1/43/7, `1-5530513321301551-052` (`ui/protection_paladin/presets.ts`) | one Greater Blessing for each class; one aura; one judgement (Wisdom or the Crusader) |
| Holy paladin | 31/20/0, `05320103225101051-500250302003` | the same as the Protection paladin |
| Restoration shaman | 0/8/43, `-005102-5530500315513151` (`ui/restoration_shaman/presets.ts`) | one totem per element (earth, air, water, fire); Mana Spring with Restorative Totems (×1.25); Mana Tide to its party |
| Discipline priest | 33/18/0, `005005231330011531-2310510303` (`ui/healing_priest/presets.ts`) | Fortitude, Prayer of Spirit, Power Infusion |
| Holy priest | 19/32/0, `00510223123-03405103030121531` (`ui/healing_priest/presets.ts`) | Fortitude, Prayer of Spirit |
| Restoration druid | 11/0/40, `05302001--5052035153113051` | Gift of the Wild; Innervate; Faerie Fire as an assignment |

**Only the priest's talents change what a tank or healer gives.** Discipline reaches Power Infusion (31
points) and deep Holy can't. No talent in any tree changes a blessing, an aura, a judgement, Battle Shout,
Sunder Armor or a weapon totem. Every "Improved" talent for them is gone:

- Improved Battle Shout. Booming Voice only widens shouts now, and Improved Sunder only cuts its Rage cost.
- Improved Seal of the Crusader, Improved Blessing of Might and Wisdom, Lasting Judgement.
- Improved Devotion, Concentration and Retribution Aura, and Sanctity Aura.
- Enhancing Totems and Improved Weapon Totems.
- Improved Mark of the Wild, which is baseline now.

So the two paladin presets give the same set, and the two priests differ by Power Infusion only.

**Not in a sensible build**, so no preset has them:

- Dark Inspiration, which needs 20 Shadow points and procs only from the priest's Mind Blast or a fully
  channeled Mind Flay
- Moonkin Aura from a healer, since a Moonkin can't heal
- Leader of the Pack from a Restoration druid, since it needs Cat or Bear Form

**Other facts:**

- Kings, Innervate, Faerie Fire, Gift of the Wild, Fortitude and Prayer of Spirit are trainer spells, so
  every priest, paladin or druid has them.
- The tanks' Demoralizing Shout, Demoralizing Roar and Thunder Clap change no DPS, so they're left out.
- **Faerie Fire (Feral) is gone from Forever.** Spells 16857 and 17390 to 17392 have no rows in 70338,
  70425, 70630 or 70685, and foreverchanges.pro says "Not in Forever". The caster Faerie Fire can be cast
  in Cat, Bear and Dire Bear Form, and costs nothing there (dev notes). forever-wiki/classes/druid.md:54
  and notes.md misread this.
- **The sim is wrong about it.** It still gives Feral the free Faerie Fire (Feral) with a 6-second
  cooldown (`sim/druid/faerie_fire.go`). Only the Feral Faerie Fire duty build uses it, so that build's
  duty cost is too low until the TASKS.md item is done.
- **The UI's talent strings for the Holy paladin and both Restoration druid presets are invalid in
  Forever.** The table has valid ones: the paladin's 3 points in Consecrated Ground moved to Unyielding
  Faith and Purifying Power, and the druid's build was rebuilt on the Thicc Resto layout.
- **Small realism nits** that don't change what the presets give: the Holy paladin takes Improved Righteous
  Fury, the paladins skip Pursuit of Justice, and the Feral tank skips Feral Charge.

**A slot's choices** are assignments, not talents: which blessings, which aura, which judgement, which
totems, and whether it holds Faerie Fire.

**Tank and healer presets are searched too** (the user, 2026-10-29); see
[Tank and healer combos](#tank-and-healer-combos). The user can also fix any slot's preset.

## Buffs and debuffs come from the members

Nothing is assumed. **The resolver fills the settings with what the members give.** Each buff and debuff
arrives in one of two ways, never both:

- **Given.** The resolver writes it into the settings, so it's up for the whole fight. This covers:
  - the long buffs, which last 60 minutes in Forever and go on before the pull (Fortitude, Arcane
    Brilliance, Gift of the Wild, Prayer of Spirit, the blessings, the auras)
  - everything a placeholder gives
- **Cast.** A DPS member's rotation casts it, and the engine sends it to its scope. A dropped totem or a late
  debuff then costs the whole raid, as it would in game. A cast comes from one of two places:
  - **The build's own provisions.** Some base rotations already keep up a raid effect, the way Moonkin Aura
    comes with Balance:
    - every warrior casts one Sunder Armor on its first GCD
    - Enhancement drops its totems before the pull (Strength of Earth, Windfury, Mana Spring)
    - Elemental drops Mana Spring
    - every hunter keeps Hunter's Mark up
    - Balance holds Moonkin Aura, Feral holds Leader of the Pack, and Marksmanship holds Trueshot Aura

    The resolver records these as cast by that member. They need no assignment.
  - **An assignment** (see [Assignments](#assignments)).

**Why never both:**

- **Doubled effects.** These would double the effect:
  - a given and a cast Windfury Totem, which are separate auras with separate cooldowns
  - a given Mana Spring (a plain MP5 stat) and a cast one (mana ticks)
  - the Windfury path in a member's consumes (`Consumes.WeaponImbue_Windfury`), which the resolver strips
- **Skipped casts.** A given buff or debuff makes the caster's rotation see it already up and skip its
  cast. That happens for Battle Shout, Faerie Fire, Curse of the Elements, Judgement of the Crusader,
  Hunter's Mark and Strength of Earth.

Tests prove the resolver keeps given and cast apart.

**Members get their own buff set** (milestone M1b). Today one `RaidBuffs` covers the whole raid, and
`PartyBuffs` holds only Atiesh and Mana Tide. So `Player` gets a new field, `RaidBuffs raid_buffs`
(`proto/api.proto`):

- When set, it replaces `Raid.buffs` for that player wholesale. It isn't merged.
- The class `AddRaidBuffs` hooks are skipped for that player, so the resolver owns every buff. A Balance
  druid's own Moonkin Aura then comes from the resolver too.
- Party buffs and debuffs stay where they are.

From M1b the same field serves the search's single-player sims, the report raid check and the page. Until
then, a single-player request carries its buffs in `Raid.buffs`, which works the same for one player.

**Pets.** Player buffs flagged players-only never reach pets, and the server blocks the rest too (standard
14). The engine already skips them (`applyPetBuffEffects`). Target debuffs do reach pets: armor debuffs
raise every pet's damage, and Curse of the Elements raises the demons' spells.

When cast buffs reach the raid (M1b), they need a players-only raid aura array, because `NewRaidAuraArray`
today loops every unit, pets included. The exception is Furious Howl, which has no players-only flag (see
[Hunter pets](#hunter-pets)).

### The provider table

The resolver works from a table that says, for each buff and debuff:

- who gives it
- its scope (raid, party, a class across the raid, one target, or the boss)
- whether it reaches pets
- whether several providers stack
- what it excludes
- how long it lasts and whether it costs GCDs in the fight
- the settings field that gives it, and its spell ids

**It's written by hand**, about 25 rows, checked against client 70685. The client's scope rule is in the
appendix, for checking a row. `tools/sources/` (update-sources) flags any of the table's spell ids that
change in a new build's wago.tools diff.

**In Go** it's a slice of rows:

```
{Effect, Providers (class, spec or preset, talent), Scope (Raid, Party, ClassInRaid, OneTarget, Boss),
 ReachesPets, Stacking (Strongest, One, Stacks n), Slot ("minor armor", "air totem", "judgement",
 "curse", ...), Excludes, InFightCost, SpellIDs,
 Write func(*proto.RaidBuffs, *proto.PartyBuffs, *proto.IndividualBuffs, *proto.Debuffs)}
```

DPS-relevant buffs. "Unverified" means the client can't say and the server decides. "No field" means there
is no given form yet:

| Buff | Given by | Scope | Several providers | Excludes | Lasts / in-fight cost | Settings field |
|---|---|---|---|---|---|---|
| Battle Shout (melee AP, rank 6, 115 at 60) | any warrior | raid, 20 yd, players | strongest counts (unverified) | Furious Howl (sim rule) | 3 min; refreshed once in fights over about 177 s | `battle_shout` |
| Furious Howl (melee AP, 83 at 60) | hunter with a wolf | raid, 15 yd; no players-only flag, so probably pets too | unverified | Battle Shout (sim rule) | the pet's: 60 s, 30 s cooldown | `furious_howl`, never read today (fixed in B) |
| Strength of Earth Totem (rank 4) | any shaman, earth | raid, 30 yd, players | doesn't stack (sim); unverified (shaman_audit 8.46) | the shaman's other earth totems | 5 min, 1 s GCD | `strength_of_earth_totem` |
| Grace of Air Totem (rank 2) | any shaman, air | raid, 30 yd, players | doesn't stack (settled, dev notes) | Windfury on the same shaman (air slot) | 5 min | `grace_of_air_totem` |
| Windfury Totem | any shaman, air | raid, 30 yd, players, main hand only; not for a shaman with Windfury Weapon | doesn't stack (settled, dev notes) | Flametongue Totem; Grace of Air on the same shaman | 5 min; 1.5 s cooldown per receiver | `totem_weapon_buff` |
| Flametongue Totem | any shaman, fire | raid, 30 yd, players | doesn't stack | Windfury Totem; Flametongue Weapon; the shaman's other fire totems (so Enhancement loses Magma and Elemental loses Searing) | 5 min; only helps Enhancement (standard 25) | `totem_weapon_buff` |
| Trueshot Aura (50 ranged AP since the 10-28 hotfix) | Marksmanship | raid, 45 yd; only hunters gain | doesn't stack | | 30 min, before the pull | `trueshot_aura` |
| Moonkin Aura (3% spell crit) | Balance | raid, 45 yd, players | doesn't stack | (stacks with Leader of the Pack now) | permanent in form | `moonkin_aura` |
| Leader of the Pack (3% physical crit) | Feral | raid, 45 yd, players | doesn't stack | | permanent in form | `leader_of_the_pack` |
| Greater Blessing of Kings, Might, Wisdom | any paladin (Kings is baseline) | one class across the raid, players | one blessing per paladin for each player (settled, tooltip; the resolver enforces it, the engine doesn't) | the paladin's other blessings on that class | 60 min | `IndividualBuffs.blessing_of_*` |
| Gift of the Wild | any druid | raid, players | doesn't stack | | 60 min | `gift_of_the_wild` |
| Arcane Brilliance | any mage | raid | doesn't stack | | 60 min; mana | `arcane_brilliance` |
| Prayer of Spirit (Divine Spirit is one target) | any priest | raid | doesn't stack | | 60 min; mana | `divine_spirit` |
| Mana Spring Totem | any shaman, water; ×1.25 with Restorative Totems | party, 30 yd | open (shaman_audit 8.45) | the shaman's other water totems, Mana Tide included | 5 min | `mana_spring_totem` |
| Mana Tide Totem | shaman talent | party | unverified | the shaman's own Mana Spring (settled: the same water slot). Dropping Mana Tide destroys it, and it must be dropped again. Another shaman's Mana Spring stays. | 12 s by the tooltip, 13 s in the sim; 5 min cooldown, at min(fight/2, 60 s) today | `PartyBuffs.mana_tide_totems` (pays a party of 5 five times today, fixed in B) |
| Dark Inspiration (+15% Spirit) | Shadow talent, 20 points | party | unverified | | 10 s, from the priest's Mind Blast or a fully channeled Mind Flay; only a Shadow priest has it | no field; the priest's own aura only |
| Power Infusion (+20% magic damage) | Discipline, 31 points | one target | unverified | Arcane Power on the target (an Era server rule, unverified) | 15 s, 3 min cooldown, off the GCD; fires as soon as ready | `IndividualBuffs.power_infusions` |
| Innervate | any druid | one target | unverified | | 20 s, 6 min cooldown; below a mana threshold | `IndividualBuffs.innervates` |
| Retribution Aura, Thorns | paladin, druid | the tank's party, the tank | | | only when the boss hits a member: ignored until tanks are simmed | |

Salvation and Mass Salvation are threat only, so they add no DPS while threat is out of scope. The
following add no DPS either:

- Fortitude, Shadow Protection and Blood Pact. Forever's Blood Pact is Stamina only, and the raid damage
  hits ignore resistances.
- Stoneskin, Windwall, Tranquil Air, Healing Stream, the resistance totems and auras, Devotion and
  Concentration Aura.

**Camp buffs** may change this (see [Decisions](#decisions)).

The debuffs:

| Debuff | Given by | Several providers | Excludes | Lasts / in-fight cost | Settings field |
|---|---|---|---|---|---|
| Sunder Armor (450 a stack) | any warrior | stacks to 5 from any warrior | Expose Armor (standard 18) | 30 s, a GCD each | `sunder_armor`, full from the pull today |
| Faerie Fire (505 armor) | any druid, in caster, Moonkin, Cat, Bear or Dire Bear Form | one per target, from any druid (settled, dev notes) | the minor armor slot | 40 s; 115 mana in caster or Moonkin Form, free in Cat and Bear; 1.5 s GCD, no cooldown | `faerie_fire` |
| Dust Cloud (505 armor) | hunter with a Tallstrider | one | the minor armor slot | the pet's, 30 s | no field |
| Curse of Recklessness | warlock | doesn't stack (70425) | the minor armor slot; the warlock's other curse | 2 min | `curse_of_recklessness` |
| Curse of the Elements (+10%, all magic schools) | warlock | doesn't stack | the warlock's other curse, not its bane | 5 min | `curse_of_elements` |
| Judgement of the Crusader (+161 Holy) | paladin | one (client-consistent, server-side) | the paladin's other judgement | 40 s; about 2 GCDs to seal and re-seal | `judgement_of_the_crusader` |
| Judgement of Wisdom | paladin | one | the paladin's other judgement | 40 s; mana for every attacker with mana | `judgement_of_wisdom` |
| Hunter's Mark (ranged AP for hunters, not pets) | hunter | one per target, the strongest counts (client-consistent, server-side) | | 2 min, a GCD | `hunters_mark` |

**Judgement of Wisdom matters.** It is worth 56 to 73 DPS to each mana user at 180 seconds in the report's
one-buff-at-a-time runs: Beast Mastery 56, fire 59, Ret Undead 73. That makes it the third-largest effect,
after Sunder and the weapon totem.

A paladin keeps up one judgement. So a raid with one paladin chooses between Wisdom and the Crusader, and
the search treats them as one choice with two options.

**Warlock curses.** Forever splits a warlock's curses from its banes (Agony, Doom, Havoc are banes;
warlock.md). So a curse assignment costs GCDs, not the warlock's Agony.

Insect Swarm, Scorpid Sting, Demoralizing Shout and Roar, Thunder Clap and Curse of Weakness only lower the
boss's damage. They add no DPS and are left out.

**Writing a role into a request.** The resolver starts from empty `RaidBuffs`, `Debuffs` and `PartyBuffs`.
It writes only the rows the buff set has, so the report's own "both judgements" and "Faerie Fire and Curse
of Recklessness" don't leak through. It writes:

- `TristateEffectRegular` wherever the old talent is gone (Battle Shout, the blessings, the totems)
- `TristateEffectImproved` for Mana Spring with Restorative Totems (×1.25 in `buffs.go`)
- each Greater Blessing into `IndividualBuffs`, on every player of that class

Thorns, Retribution Aura, Salvation and Judgement of Light are left out. M1 measures once whether dropping
them moves any base build.

**Rules the resolver enforces:**

- **One per provider.**
  - A paladin gives each player one blessing. Greater Blessings cover a class, so the choice is a paladin
    and a class.
  - A paladin keeps up one judgement and holds one aura.
  - A shaman has one totem per element. Grace of Air and Windfury together need two shamans.
  - A warlock has one curse.
- **Shared slots.**
  - The minor armor slot takes one of Faerie Fire, Curse of Recklessness and Dust Cloud (standard 20).
  - Windfury and Flametongue Totem exclude each other.
- **Nobody provides it.** If no member gives a buff, nobody has it. When three paladins can't cover Kings,
  Might and Wisdom for every class, the result shows who misses out.
- **What stacks from several providers** follows the tables above. Where the client can't say, the raid sim
  uses the sim's current rule and lists it as an assumption: Mana Spring, the other totems, Hunter's Mark,
  Judgement of the Crusader, the pets' effects, and Power Infusion against Arcane Power. notes.md (Need to
  Verify, At 60) has the launch checks.

### Sunder Armor

Every warrior's base rotation casts one Sunder on its first GCD. After that, the warrior tank keeps it up
(the user, 2026-10-29). In a raid with a warrior tank, the DPS warriors never sunder again.

The engine doesn't model that today:

- The given Sunder puts all 5 stacks on in the first 0.8 seconds (`debuffs.go:157-170`) and never lets them
  drop.
- So the DPS warriors' pull Sunders always land on a full stack and do nothing.

That overstates physical builds against casters by about 0.4 to 1% of their DPS. A ramp is worth 20 to 35
DPS in a 5-member physical raid, and Sunder is the largest single effect.

**The model** (milestone M1b) is a given Sunder with a ramp:

- Each DPS warrior's pull Sunder adds a stack when it lands.
- The tank adds one stack per GCD from the pull, until there are 5.
- From then on the tank keeps it at 5.

The ramp is a setting on the given debuff, so the raid sim and the single-player sims use the same one. The
number of DPS warriors' pull Sunders is part of the buff set. A raid without a warrior tank but with DPS
warriors needs a Sunder upkeep build (see [Assignments](#assignments)).

### Hunter pets

Two pet families give the raid something. Until the launch checks settle it, the raid sim assumes neither
stacks with its player counterpart (the user, 2026-10-29).

- **A Tallstrider's Dust Cloud** takes 505 armor off the boss for 30 seconds. It sits in the minor armor
  slot with Faerie Fire (505) and Curse of Recklessness, and only the strongest counts. So with a druid
  keeping Faerie Fire up, it adds nothing. Its value is in a raid without one: the slot gets filled without
  a druid, or without a warlock spending its curse on Recklessness.
- **A wolf's Furious Howl** shares Battle Shout's slot (`battleShoutCategory`), so with a warrior shouting
  it adds nothing to players. But it has no players-only flag, and pets never get Battle Shout. If it
  reaches other pets, a wolf gives 83 attack power to every hunter pet and demon nearby even with a warrior
  shouting. The sim today gives it only to the wolf and the players in its party.

The launch checks are in notes.md (Need to Verify, At 60): what stacks, and whether the pets' buffs reach
other parties and other pets.

Both pets cost the hunter some of their own DPS, since the usual pet is picked for it. Each hunter build
with a pet (Beast Mastery, Marksmanship, Survival) gets variants with a Tallstrider and with a wolf (the
user, 2026-10-29), with their own ids. The variants carry that cost, and the search weighs it against what
the pet gives. Lone Wolf has no pet.

## Assignments

An assignment is a job one member does for the raid. Each one is **automatic or set by hand** (the user,
2026-10-29).

There are three kinds:

- **Upkeep in the fight**, which costs the holder GCDs:
  - Sunder Armor after the pull, usually the tank
  - the minor armor slot: Faerie Fire, Curse of Recklessness or Dust Cloud
  - Curse of the Elements
  - a judgement: the Crusader or Wisdom
  - Battle Shout
- **Before the pull**, which costs nothing in the fight:
  - which Greater Blessing each paladin gives each class
  - which aura each paladin holds
  - which totem each shaman holds in each element
- **Cooldowns given to someone else:**
  - Power Infusion: which caster
  - Innervate: which mana user
  - Mana Tide, which reaches only its own party

  Today they fire on fixed rules: Power Infusion as soon as it's ready, Innervate below a mana threshold,
  Mana Tide at min(fight/2, 60 s). Choosing when needs new settings, so the first version only chooses
  who.

**Automatic assignment** is for raids built by hand, and for re-resolving a comp when a member is taken
out. It gives each job to a placeholder first, since that costs no DPS. When no placeholder can, it goes to
the eligible DPS member whose raid DPS drops least, by their single-player sims under the set at the raid's
length, then to the lower slot. The search assigns its own way (see [The search](#the-search)).

**An assignment's cost** is the holder's base DPS minus their DPS on duty, under the same set at the raid's
length. Duty runs the build's duty request, with the effect switched off in the holder's own settings, so
it casts the effect itself and pays for it. The report's `builds.json` has the numbers at 180 seconds:

- Faerie Fire costs Balance 31.8 (33.5 at 120 s, 29.1 at 300 s).
- Curse of the Elements costs Destruction 8.9.
- Battle Shout costs Fury 1.8.

**A DPS member can only hold an upkeep assignment its build has a duty variant for.** In `meta.json`, a
duty variant has `duty` (one of "Battle Shout", "Faerie Fire", "Curse of the Elements", "Judgement of the
Crusader") and `baseBuild`, the base build's id. Today there are 12:

- Battle Shout: the 3 warrior builds
- Faerie Fire: Balance, Balance Wrath and Feral
- Curse of the Elements: the 4 warlock builds
- Judgement of the Crusader: `paladin__level_60_sub_holy` and `paladin__level_60_shock`

**Builds still missing** (milestone M1b):

- **A Sunder upkeep warrior**, for a raid with no warrior tank.
- **A Curse of Recklessness warlock.**
- **Judgement of the Crusader** for `paladin__level_60` and `paladin__level_60_shock_ret`. The UI already
  has `level60_crusader_duty.apl.json`.
- **A Judgement of Wisdom paladin**, for the Ret builds.
- **A hunter without Hunter's Mark.** Today every hunter casts it, so "only one hunter marks" can't be
  expressed. The other hunters' casts are wasted GCDs, which is small (Survival gains 2 to 4 from a
  stronger mark next to it).
- **Assignable shaman totems.** Each rotation's totems are fixed by spell id:
  - Enhancement drops Strength of Earth, Windfury, Mana Spring and Magma.
  - Elemental drops Mana Spring and Searing.

  So two Enhancement shamans both drop Windfury, and nobody drops Grace of Air. This needs a per-player
  "assigned totems" option that an APL action reads ("cast the assigned totem for this element"), not a
  build for each totem mix.

Until those exist, those jobs can only be held by placeholders, or given.

## Groups

On the page, the user places groups by dragging members. A new raid starts from the rule below.

**The search places groups by the same rule, and it's exact** (the user, 2026-10-29: "whatever maximizes
DPS"). Only Mana Spring, Mana Tide and Dark Inspiration depend on groups, and each member's gain from them
is its own.

1. **A package** is whatever party effects one party has. A Restoration shaman gives Mana Spring ×1.25 plus
   Mana Tide, an Enhancement or Elemental shaman gives plain Mana Spring, and two shamans in one party give
   their stacking rule's result. Each build is simmed once per distinct package in the set, and once
   without.
2. **A party's free seats** are 5 minus its placeholders and its provider members.
3. **Each packaged seat takes the best build under its package**, repeats allowed. So choosing the builds
   and placing them is one step, and exact.
4. The rest fill on their DPS without a package.

Until Dark Inspiration has a given form (M1b), it isn't a package.

## The fight

- **The bosses.** Each launch raid gets an encounter preset:
  - **Barrow Deeps (10) and Hyjal Summit (20):** the report's generic level 63 boss (3731 armor, Unknown
    creature type) in the natural flask zone, as in every Level 60 preset (standard 11).
  - **Onyxia's Lair (40):** a Dragonkin boss outside the flask zone. Off the zone, the natural flasks give
    only their Stamina: that costs a Beast Mastery hunter about 38 DPS and a Fury warrior about 64.

  Forever's raid bosses aren't known yet. When they are, each one is its own encounter.
- **The creature type** picks the encounter variants in the pool (vs Humanoid, Undead, Dragonkin): a variant
  joins only when the boss's type matches.
- **Length.** It can be set two ways (the user, 2026-10-29):
  - **A fixed duration**, with the report's 180 ± 20 seconds as the default.
  - **The bosses' health.** The fight ends when every boss is dead, each one dying on its own
    (`sim/core/target_death.go`). A health fight has no duration variation (fixed in B). Today the
    variation still moves the end of fight every rotation sees, and with ±20 seconds a 10-member raid lost
    0.8%.
- **Length matters a lot.** From 150 to 210 seconds, member DPS falls by different amounts:
  - Fury 6.0%
  - rogue 3.7%
  - Destruction 2.8%
  - the hunters 2.3%
  - fire 1.5%
  - frost and Enhancement 0.8%

  The ranking flips: Fury beats Destruction at 150 seconds and loses at 210. So the search always sims at
  the raid's own length (see [Fight length](#fight-length)), and the result is checked at 120 and 300
  seconds (see [Sensitivity](#sensitivity)).
- **Several bosses come later** (the user, 2026-10-29: "or multiple if multiple exist"). They need engine
  work (see [Later](#later)). Until then, the raid sim fights one boss.
- **Raid damage.** Every DPS member takes the encounter's raid damage hits, as in the report (2 a minute at
  1500, standard 24). They feed procs and Rage, since nobody heals.
- **Threat is out of scope for now** (the user, 2026-10-29). When it's added, it is reported but not
  modelled:
  - a warning whenever the tank couldn't hold threat, saying when and against whom
  - a results page that drills down into each member's threat on each boss
  - nothing holds a member's damage back

  Until then the result prints each member's threat per second, which is already measured, so a raid of 14
  Fury warriors is at least visible as one.

## The build pool

**The search's DPS pool is the presets marked for raid sims, for now the DPS Report's builds** (the user,
2026-10-29).

For a generic boss that's 37 builds: the report's 58 requests, less 12 duty variants and 9 encounter
variants. The 6 hunter pet variants (3 builds × Tallstrider and wolf) will make it 43.

- **The duty variants** are the base build holding an assignment, not entries of their own (`duty` and
  `baseBuild` in the variant's `meta.json`).
- **The encounter variants** have `baseBuild` and `encounter` (Humanoid, Undead, Dragonkin) instead, and
  join only when the boss's creature type matches.
- **The "+ Mana Tide", Water Shield and Imp GW builds** are different talent builds, so they are entries of
  their own.
- **New requests** (the pet variants, the missing duty builds) go into `tools/dps_report/req/`; see
  [Decisions](#decisions) for whether the report publishes them.

**Where the pool comes from:**

- **The command line search and the tests** read `tools/dps_report/req/*.json` and its `meta.json`
  directly. The report's regenerations then flow into the search on their own.
- **The page** builds the same players from the UI's presets. This needs work, because today nothing links
  the UI presets to the report's builds:
  - the report's requests come from `tools/rotopt/mksettings`, with its own consumes and hand edits listed
    in each `meta.json`
  - the names differ ("Fury Dual Wield" against "Level 60 Fury")
  - 12 builds already differ: the 8 warlocks' and 3 shadow priests' distance and channel clip, and the
    Survival hunter's distance
  - Enhancement's Imp GW build has no UI preset

  So the page's first version loads the report's players as a data file. A `raidSim` mark on `PresetBuild`
  and a parity check against `req/` come only when the page takes members from the spec pages
  ([Later](#later)).

## Comp search

The main thing the raid sim is for (the user, 2026-10-29).

**The inputs:**

- the raid size
- the tank and healer slots, each with a fixed preset or "any" (the default)
- the encounter (boss, creature type, length)
- the faction

**There are no limits per spec or class** (the user, 2026-10-29). The goal is the most raid DPS and
nothing else. The search may fill the raid with one spec, and may drop the only provider of a buff when a
stronger spec in its place gives more. The output says so when the answer is "29 Beast Mastery hunters".

### Why it doesn't try combinations

The number of distinct comps runs from 3e7 for 7 DPS to 2.5e18 for 29, so trying them is out.

It's also unnecessary (the user, 2026-10-29). **Each build's DPS can be judged on its own, under a given
set of buffs and debuffs**, because members affect each other almost only through what they give. This was
measured with raids of report builds, with everything given and a fixed 180 ± 20 second fight:

- Each member landed within ±3 DPS of their single-player sim. Casters were within ±1.5. Survival and
  Enhancement were up to +4 next to many physical members, partly from the shared Hunter's Mark.
- The raid landed within 0.1% of the sum: −0.05% to +0.04% for 10 members, +0.06% for 20, and +0.1% for
  12 physical members, reproduced on a second seed.
- Cast Battle Shout, Faerie Fire and Curse of the Elements against given ones differed by 0.05% at most.

So a raid is two things:

- **The buff set:** the placeholders' choices and the provider members (each a build plus what it holds).
  The buffs and debuffs that are up follow from it through the resolver.
- **The fill:** given the buff set, each other DPS slot's DPS is that build's own DPS under it, which a
  single-player sim measures. One player costs about 0.4 ms an iteration, against 19 ms for a raid of 20.

### Roles

A role is one thing the raid can have or not, with the ways it can be covered:

| Role | Options |
|---|---|
| Battle Shout | a warrior tank; a DPS warrior on duty; a wolf; none |
| Sunder upkeep | a warrior tank; a Sunder upkeep warrior; none (pull Sunders only) |
| Minor armor slot | a Feral tank; a Restoration druid; a DPS druid on Faerie Fire; a Curse of Recklessness warlock; a Tallstrider; none |
| Curse of the Elements | a warlock on duty; none |
| Judgement, per paladin | the Crusader; Wisdom |
| Blessings | Kings, Might, Wisdom per class, from the paladins present (placeholders or Ret) |
| Earth, air, water and fire totems | each shaman's choice per element (placeholders, Enhancement, Elemental) |
| Moonkin Aura | a Balance druid; none |
| Leader of the Pack | a Feral druid (cat or tank); none |
| Trueshot Aura | a Marksmanship hunter; none |
| Arcane Brilliance | a mage; none |
| Gift of the Wild | a Feral tank; a Restoration druid; any DPS druid; none |
| Prayer of Spirit | a priest (placeholder or Shadow); none |
| Power Infusion, Innervate, Mana Tide | from the placeholders that have them; only "who" is chosen |

Roles are coupled through their providers:

- A shaman covers all four elements.
- A warlock's one curse is either Curse of the Elements or Curse of Recklessness.
- A Balance druid on Faerie Fire covers Moonkin Aura, Gift of the Wild and the minor armor slot.

So a search move adds, removes or swaps one provider member, or changes one placeholder's choice, not one
role.

### Tank and healer combos

The search covers every combination of tank and healer presets that fits the slots (the user, 2026-10-29).
The presets (see [Tank and healer presets](#tank-and-healer-presets)):

- **Tanks:** Protection warrior, Feral tank, Protection paladin.
- **Healers:** Holy paladin, Restoration shaman, Discipline priest, Holy priest, Restoration druid.

Combinations are counted as multisets, since two warrior tanks in either order are one combo:

| Raid | Tanks | Healers | Combos | Distinct signatures |
|---|---|---|---|---|
| 10 | 1 or 2 | 2 | 3 × 15 + 6 × 15 = 135 | 58 |
| 20 | 2 | 4 or 5 | 6 × 70 + 6 × 126 = 1,176 | 110 |
| 40 | 3 | 7 to 9 | 10 × (330 + 495 + 715) = 15,400 | 102 |

**A combo's signature** is what its placeholders can give:

- whether there's a warrior tank
- whether there's a Feral tank
- whether a druid gives Faerie Fire and Gift of the Wild
- the number of paladins, capped at 3
- the number of Restoration shamans, capped at 2
- whether there's a priest

The search runs once per signature, and every combo with that signature takes its result. Power Infusion,
Innervate and Mana Tide counts only decide how many members get those cooldowns inside the fill.

**What it reports:**

- **The combos ranked by raid DPS**, each with its best DPS comp, and how far it is behind the best.
- **Each tank and healer preset's DPS value** at the best combo: the best combo minus the best combo with
  that slot's preset changed.
- **Each preset's best-with minus best-without:** the best combo with at least one of the preset, minus the
  best combo with none. This is how the presets are ranked against each other.

  An average over the combos would mislead. Combos are weighted by how often a multiset occurs, not by how
  realistic it is, and most of the 40-player combos have 3 or more paladins, so a paladin's average value
  is about 0.
- **A note on the 40-player ranking:** it's decided by the cooldown counts (Power Infusion, Innervate),
  since every signature there covers every buff. Those cooldowns fire on fixed rules (see
  [Assignments](#assignments)).
- **Their own damage isn't in it** while tanks and healers aren't simmed. When they are, it adds to the
  same table.

### The search

For each signature:

1. **Prune the fillers.** A build a dominates b if a's DPS minus b's is at least the sum, over the roles, of
   how much more b gains from each role than a. Only the builds no other build dominates fill slots (at a
   margin of 1.5 DPS, since mana effects aren't additive). That leaves 14 of the 37 builds at 180 seconds.
   Providers need only one duty variant per role.
2. **Rank every buff set with the additive model.** The model is each build's DPS under the full set,
   minus the report's one-buff-at-a-time ablations, plus pair terms where two roles interact:
   - the minor armor slot (Faerie Fire and Curse of Recklessness each show 0 alone for Beast Mastery, but
     64.4 together)
   - the two air totems
   - the mana sources

   Every combination of placeholder choices and provider members is scored this way: about 50,000
   arithmetic evaluations, no sims.
3. **Sim the top 50 buff sets exactly.** Each build's DPS under a set comes from a single-player sim, with
   the set written into its request. The model's ranking is checked against the exact numbers, and
   reported.
4. **Fill a buff set exactly:**
   - The provider members are in. Each one runs its duty variant for what it holds, simmed with that effect
     off in its own settings, so it pays the assignment's cost. In the search, only provider members the
     search chose hold duties. Fillers hold none.
   - Per-member choices are separable, so they're solved exactly inside the fill:
     - each class's blessing, picked by which gives that class's best build the most
     - the packages (see [Groups](#groups))
     - Power Infusion and Innervate targets
   - The remaining DPS slots take the best builds under the set, repeats allowed.
   - A build that gives anything (the "Given by" column: a mage's Arcane Brilliance, a priest's Prayer of
     Spirit, a druid's Gift of the Wild, a Balance druid's Moonkin Aura, a shaman's totems) may only fill a
     slot when the set already has what it gives. So the comp stays the set it was scored under.
   - The Sunder ramp's stack count follows the number of DPS warriors. The fill is run again once with the
     new count.
   - The score is the sum of the members' DPS.
5. **Polish with local moves** from the best exact sets:
   - add, remove or swap one provider member
   - change one placeholder's choice: an air or fire totem, a judgement, or a class's blessing

   A swap re-resolves the duties, so one move can turn Balance on Faerie Fire into a Feral cat on Faerie
   Fire. Moves that raise the score are kept, until none helps.
6. **Confirm with raid sims** (see [Confirmation](#confirmation)).

### Fight length

The single-player sims run at the raid's fight length:

- **A fixed fight:** the same duration and variation as the raid. Iteration seeding pairs the durations
  between the raid and its single-player sims.
- **A health fight:** the length is health ÷ raid DPS, and raid DPS depends on the length. Iterating that
  converges in 1 to 2 steps, since DPS changes little with length (the factor is about 0.18 for Fury and
  0.04 to 0.07 for hunters and casters).

  So each pass of the search holds one length for all sets. The single-player sims run on a 5-second grid,
  and a build's DPS is interpolated linearly between grid points, so the cache hits across sets and combos.
  The pass runs again when the winner's length moves by more than 1 second. A 3-second error would bias a
  melee-heavy comp against a caster-heavy one by about 0.26%, almost the whole confirmation window. The
  confirmation runs real health fights.

  A rotation that saves cooldowns for the end (Fury's Death Wish) knows the exact end in a fixed
  single-player sim, but only an estimate in a health fight. The confirmation shows that gap.

### Confirmation

Raid sims confirm **distinct buff sets, not comps**:

- Fill variants under one set differ exactly by their members' single-player sims, which check 1 already
  covers.
- Comps that resolve to the same request are one: the two paladin presets, and party permutations.
- At most 8 sets are raid-simmed, paired, at 5,000 iterations. The top 2 go on to 20,000 (standard 15).
- If a raid sim disagrees with its score by more than noise, the result shows the gap.

The raid sim catches what the single-player sims can't:

- a member's first curse or Faerie Fire landing after the pull
- the shared Hunter's Mark
- Mana Tide's timing
- the engine's interactions

**The tie rule:**

- Comps whose paired gap is under the larger of 2 standard errors and 0.1% are one tie set, and they're
  published as a tie.
- The comp shown is the one with the fewest distinct builds, then the fewest DPS members on duty.
- Fillers within 2 standard errors are listed as interchangeable. Fire and frostfire are 0.6 DPS apart, with
  standard errors of 1.86 and 2.57 at 1,000 iterations, so picking between them would be a coin flip.

### The cache

**A spec simmed under a buff set it was already simmed under is never simmed again** (the user,
2026-10-29).

- **The key** is the SHA-256 of the deterministic proto marshal of the resolved single-player
  `RaidSimRequest`, plus the sim's version:
  - The request covers the build's whole setup (gear, talents, base or duty rotation, consumes, race), its
    raid buffs, individual and party buffs, debuffs, the encounter, the duration (on the length grid) and
    variation, the iterations and the seed.
  - The version is a hash of the inputs that can change a number: `sim/`, the item database and the spec's
    folders. These are the same hashes the report's `tools/dps_report/stale.mjs` uses.
- **Buffs a build doesn't read are dropped from its key.** Without this, Moonkin Aura on or off would be a
  new key for every warrior, and the cache would barely hit. A buff is dropped only when removing it gives
  a bit-identical result for that build (the report's ablation runs record this as `same: true`). A
  bit-identical result proves the build doesn't read it, so dropping it can't be wrong. Anything else
  stays in the key.
- **The value** is the DPS, its standard error and standard deviation, and the mana stats.
- **Where it lives:**
  - In memory during a search.
  - On disk next to the native server (`compsearch/` in its data folder).
  - For the command line, in `~/.cache/forever-sim/compsearch/`.
- **What gets reused:** two Fury warriors in one comp, or one comp reached two ways, are one sim. Other raid
  sizes, other slot counts, other signatures and later searches reuse what's there, since a sim is keyed by
  the buff set a build sees, not by who gives it.
- **A code or database change starts a fresh cache.** Old entries are pruned.
- **In a Claude Code cloud environment** the cache starts empty each session, and that's fine. The cache
  isn't committed.

### Cost

The exact sims are what cost. They're keyed by the buff set a build sees, so the signatures share them.
Expect a few hundred distinct sets × about 14 builds × one or two packages, at 1,000 iterations:

- **Cold:** minutes on 16 cores, about an hour on one.
- **Warm:** seconds.

The model ranking and the fills are arithmetic.

The confirmation is at most 8 raid sims at 5,000 iterations plus 2 at 20,000:

- **20 members:** about 6 seconds a set on 16 cores.
- **40 members:** about 1.5 minutes for the final run with today's engine.

M1 measures all of this.

### Noise

- At 5,000 iterations a 10-member raid's standard error is 0.021%. With today's correlation of 0.22 between
  two raids on the same seed, their difference has a standard error of about 0.026%, so a 0.1% gap is about
  4σ. For 20 members it's about 5σ.
- The real limit is not noise but how well the single-player sims add up, about 0.1%. That is the same size
  as the gaps between the top comps: swapping one build in a 10-member raid is worth 0.02% to 0.16%. That's
  why there's a tie rule.
- Seeds s and s + 1 are the same sample shifted by one iteration (iteration i uses seed + i). Replicates use
  seeds further apart than the iteration count.
- Per-member random streams would make whole-raid comparisons sharper. The search doesn't need them, so
  they come later.

### Where it runs

On the user's machine or in Claude Code cloud environments, natively; nothing is hosted (the user,
2026-10-29).

- **The command line** (`tools/compsearch`, then `wowsimcli`) runs the search on goroutines across every
  core. It writes the result as JSON and markdown.
- **The native server** (`sim/web/main.go`, run by the downloadable binary or `make host`) gets the same
  search as an async endpoint, `/compSearchAsync`. It streams progress with the best comp so far, and Abort
  stops it. The pages already send their sims to the native server when it serves them (`ui/core/sim.ts`,
  net workers).
- **A page served as static files** can still show results and sim a single raid in its WebAssembly
  workers, but searching needs the native server.

### The result

It leads with what could change the answer:

1. **The tie set**, and every sensitivity scenario that changes the winner (see
   [Sensitivity](#sensitivity)).
2. **For each raid size and faction:** the best comp and 2 to 4 runners-up, with the difference and paired
   95% intervals.
3. **The tank and healer combos ranked**, and each preset's DPS value (see
   [Tank and healer combos](#tank-and-healer-combos)).
4. **The role table:** each role's gain, its provider's cost and the net, at 10, 20 and 40.
   - Each role's gain is measured at the winning set, with that role alone removed, then re-filled and
     re-resolved.
   - Pairs that interact strongly are shown together: Grace of Air with Trueshot Aura, Grace of Air with
     Moonkin Aura, and the mana sources. Trueshot Aura, for example, is worth 10.8 a hunter with Grace of
     Air up, and 0 under Windfury only, because the fill turns to mages.
5. **The filler ranking** under the winning buff set, each build with the spread of the report's
   assumptions (`assumption_dps.json`).
6. **Each member's contribution and replacement margin** (see [Member value](#member-value)).
7. **Each member's threat per second.**
8. **What isn't modelled:** healing, tanking, threat, pet deaths, movement, boss mechanics.

## Sensitivity

The top comps are close. A rough estimate from the report's one-buff-at-a-time runs puts a hunter-heavy raid
and a caster-heavy raid a few percent apart, about 43 DPS a slot. That is inside the spread of the report's
own open assumptions, up to ±60 to 120 DPS a build.

So the result says whether the winner holds under each of these:

| Scenario | Why | Cost |
|---|---|---|
| Each build shifted by its assumption combos (`assumption_dps.json`) | the report's open rules move builds against each other | no sims |
| 120 and 300 seconds | Fury, rogues and the hunters fall faster than casters with length; hunters run out of mana by 300 | a re-run at each length |
| Each raid's boss: generic in the flask zone; Dragonkin outside it | Onyxia, where Beast Mastery (1093.3) and fire (1089.0) nearly tie | a re-run |
| Raid damage at 0 and 10 hits a minute | Demonology loses 63 DPS at 10 a minute; Fury 2H gains 56 | from the assumption data |
| Light consumes (no flask) | the flask is worth 112 to a fire mage and 38 to a Beast Mastery hunter | from the one-buff-at-a-time runs |
| Pet uptime about 80%, and Lone Wolf | the pet is 332.7 of a Beast Mastery hunter's DPS and never dies in the sim. At 80% Beast Mastery falls to 1065.3, below Lone Wolf (1094.3) and fire (1089.0), and "29 Beast Mastery" becomes "29 Lone Wolf" | a re-run of the hunter builds |
| Each unverified stacking rule on and off | Mana Spring from two shamans, Furious Howl against Battle Shout and on pets, Hunter's Mark, the Crusader | a re-run |
| Camp buffs on (if the user wants them modelled, see [Decisions](#decisions)) | they replace most of Kings, Might and Strength of Earth | a re-run |

When the winner changes under a scenario, the result says so and shows the comps that win as a tie set, not
one answer.

## Member value

All of these come from the single-player sims under buff sets. So they need no raid sims once the cache has
them.

- **Contribution of a DPS member:** raid DPS minus raid DPS without the member. Without them, the comp is
  resolved again: automatic assignments move to the next cheapest holder, and assignments set by hand are
  dropped. It's shown as the member's own DPS plus what they add to everyone else. This is "what a spec
  brings".
- **Contribution of a tank or healer:** the slot count is fixed, so taking one out leaves a hole. It's the
  combo's raid DPS minus the best raid DPS with that slot's preset changed.
- **Identical members share one value.** Two copies of one build get the same contribution, whichever one
  the search calls the provider.
- **Replacement margin:** raid DPS minus the best raid DPS with another build in that slot. The neighbours
  are evaluated on demand from the cache, not taken from whichever sets the search happened to visit. At
  the best comp it is zero or more, and a margin within noise means two builds are interchangeable.
- **Value of adding** a spec that isn't in the comp: the best raid DPS from swapping it in for any member,
  minus raid DPS.
- **Assignment cost:** the holder's base DPS minus their DPS with the assignment, under the same set.

A search returns all of them with its result. For a raid built by hand, a button runs the sims the cache
lacks.

## The page

The raid sim is **a page of its own** (the user, 2026-10-29). It is not part of a spec's page or of the DPS
Report.

It's linked from a "Raid sim" card in the landing page's Reports section, next to the DPS Report
(`ui/index.html`), and launched through `raidSimStatus` in `launched_sims.ts`. The sidebar's mode picker
doesn't fit: its modes are views of one page.

**The page sets up only the raid:**

- the raid size and the tank and healer slots
- which member goes in each slot, and which group they're in
- assignments
- the encounter
- the comp search

**It doesn't set up a member's gear, talents, rotation or consumes.** Those come whole from the spec (the
user, 2026-10-29).

Since nothing is hosted, a raid's result is shared as a static page. So the page comes in steps:

1. **A results viewer** (M2a) for the search's JSON, deployable as static files. This is what a guild can
   open.
2. **A comp editor with Run** (M2b):
   - members are the report's players, loaded as a data file
   - placeholders, groups, assignments and the encounter
   - Run through `Sim.runRaidSimRequest`, with `EmbeddedDetailedResults` and its player filter
3. **The search button** (M2c), through the native server's `/compSearchAsync`.

**Later**, members from the spec pages:

- **A member from a spec page** is a reference plus a snapshot:
  - a preset marked for raid sims
  - a saved setup (a Preset Configuration, in `<spec prefix>__savedPresetConfigurations__`)
  - the spec page's current settings, taken once as a snapshot

  All pages are under `/forever/` on one host, so the raid page can read the spec pages' storage. The
  prefixes come from `specToLocalStorageKey`, typo included (`__classic_enhacement_shaman`). A headless
  builder turns a source into a `Player`: the spec's `migrateSettings`, then `applyDefaults` and
  `applyBuild` (moved into shared functions), then `toProto(false, true)`.
- **"Open in spec page"** opens a member in a raid member mode of its spec page:
  - It doesn't overwrite the page's own settings.
  - Its assumed raid is everything this raid gives the member, at full uptime from the pull, with the
    member's own assignments switched off. It runs the member's base or duty rotation, as in the raid.
  - Its fight has a fixed length (the raid's average) with no duration variation.
  - A "Save to raid" button writes back to a key the raid page owns. The raid page listens for the
    browser's `storage` event.
- **"Add to raid sim"** on a spec page adds a reference to a raid.
- **Raid history, saved raids and share links.**
- **Retiring `ui/raid`.** It's the upstream raid page: hidden, stale, and built around editing members
  inside it. `raidSimPresets` and the `isWithinRaidSim` branches go with it.

**What the page reuses:**

- the `SimUI` base and `addRaidSimAction`, which already shows raid DPS for more than one player
- `EmbeddedDetailedResults`, with its player filter
- `EncounterPicker`
- `Sim.runRaidSimRequest` and the worker pool
- `Raid` and `Party` as the request builder. Set `numActiveParties` to 2, 4 or 8, and never send empty
  players for empty slots.
- the old raid picker's drag and drop

**Results on the page:**

- raid DPS with a 95% interval
- a party table
- a member table with each member's difference from the report, shown when the member's request equals the
  report build's
- each buff and debuff with its provider, scope and uptime, from the resolver's record
- members who miss a buff the raid could have given them
- assignment cost and member value

If a change touches shared result views (the events view, the replay), the DPS Report is deployed after it
(CLAUDE.md).

## Checking it

1. **The report raid.** A raid of base report builds, with no duty or encounter variants:
   - Each member gets their own request's raid buffs (the per-player field) and the base requests' common
     debuffs (all 46 requests without a duty carry the same ones).
   - The fight is a fixed 180 ± 20 seconds.
   - Each member lands within 4 × √(se₁² + se₂²) of a single-player sim of the same request, on the same
     commit, at the same seed and iterations. That's 4σ, since 3σ across 37 members fails by chance about 1
     time in 10.
   - Survival and Enhancement next to many physical members are allowed ±5 DPS until their small coupling
     is found.

   It runs on a 10-member raid, a 20-member raid and a physical-heavy raid. It compares with the
   single-player sim, not the published numbers, which come from the report's merged tree with its own
   patches.
2. **Doubles.** Every base build twice in one raid: the second member lands within noise of the build alone.
3. **Size.** Raids of 20 and 40 build and run without a panic, a 40-player health fight included. A party
   with gaps in slots 1 and 3 gives everyone the right setup.
4. **Scope.**
   - A member gets a party buff only from their own party, and a raid buff from anyone.
   - A cast Battle Shout reaches the raid, and two shouts give the strongest.
   - A dropped totem reaches its scope, and the buff stays while a second identical totem stands.
   - Given and cast never both apply.
5. **The search against brute force.** On a pool of 6 builds and 4 slots, the search finds the same best comp
   as trying every one under the summed single-player model.
6. **After 2026-12-09**, the user's raid logs (`tools/warcraftlogs`, `logcheck`) and public logs are checked
   against the sim: raid DPS by comp and buff uptimes.

## Milestones

Forever launches on 2026-11-04, and the classes will change with it. The raids open on 2026-12-09.

**Implementation starts with the bug fixes** (the user, 2026-10-29). M1 doesn't depend on them, since its
single-player sims hit none of them, so this is an order, not a dependency.

**The plan:**

1. B, the bug fixes, first.
2. M1 working by about 11-07, then run against the launch builds.
3. M1b by mid-November.
4. M2a by late November.
5. A final run with the sensitivity matrix in the week of 12-01.
6. Check 6 against real logs after 12-09.

Each step says what it needs from this doc and when it's done.

### B: the bug fixes

Every bug in the Appendix's "What breaks" list, except a buff that must reach a new scope (Battle Shout, the
totems, Furious Howl, Blood Pact to the party), which is M1b.

**None of them may move a single-player number.** Every one of the 58 report requests gives bit-identical
results at a fixed seed before and after.

**Already done once:** all of B was fixed in a scratch copy on 2026-10-29. Its patch and test file are in
this session's scratchpad (`milestone_b/milestone_b.patch`, `milestone_b/raid_forever_test.go`). The 58
requests' result hashes came out identical, and each new test failed before the fix and passed after.
Implementation can start from that patch. The scratchpad goes with the session, but this section lists every fix
without it.

**The fixes, in commit order:**

0. **The test file and the baseline.** `sim/raid_forever_test.go`, with the helpers and
   `TestRaidForeverReportHashes`. That test is opt-in (`RAID_FOREVER_HASHES=file`, about 21 s): it runs
   the 58 requests at 1,000 iterations, seed 1, with first-iteration logs on, and hashes each whole
   canonicalized result. Metric lists come out in map order, so they're sorted before hashing. Take the
   baseline before the first fix.
1. **The aura cap and the presim list.** These unblock raids of 20 and 40.
   - The aura cap (`aura.go:461`) becomes 200 × the number of players. Each caster registers every rank's
     DoT on each target. Not registering ranks that can't be cast would save the most, but it changes
     registration order, so it belongs in Speed.
   - The presim list (`presim.go:38`) becomes a map. Any health fight with 26 or more players panics today.
   - Tests: `AuraCap` (12 fire mages) and `Size`.
2. **Positions.**
   - Configs are indexed by `party.Index` and `char.PartyIndex`, not by position: `raid.go:270,282`,
     `environment.go:206,212`, `presim.go:83,85,106`.
   - `GetUnit` matches `party.Index` (`environment.go:392`).
   - The results keep one slot per config party, with an empty party built like a real one, since
     `sim_concurrent` reads their DPS.
   - The presim's target-dummy lookup is skipped by type.
   - A nil party or player doesn't panic.
   - Test: `Gaps`. A raid with gaps or a missing party gives exactly the compact raid's DPS, and `GetUnit`
     finds index 6.
3. **Own auras.**
   - The personal debuffs' labels are already per caster: Fire Vulnerability, Winter's Chill, Shadow
     Weaving, Light's Vigil, Shadow Vulnerability, Hemorrhage.
   - The fix is the lookup. A new `Aura.Owner` is set in `NewEnemyAuraArray` and for per-target DoTs.
     `NewAuraReference` then takes the first aura whose owner is nil or the rotation's own unit, then the
     own rank family, then the others. "Nil or own" keeps single-player results identical. A strict "own
     first" rule would flip the given-then-own pairs (Demoralizing Shout, Scorpid Sting, Insect Swarm).
   - Stormstrike's label becomes `"Stormstrike - " + shaman.Label`.
   - Test: `Doubles`. It takes about 40 seconds, so it's skipped under `-short`.
4. **Judgement of the Crusader.** Today `GetOrRegisterAura` keeps the first paladin's `OnGain` but the
   last one's hooks (`aura.go:493-510`).
   - Fix: one shared aura per target, with the last judge, the bonus and the applied bonus kept on the
     environment. A new `core.JudgeTheCrusader(sim, aura, caster, bonus)` lets another paladin's judgement
     replace it, and a given permanent one stays. It keeps a `judges` list, so the report paladins' log
     lines stay identical.
   - Per-caster auras in one exclusive category would be cleaner, but they would move the 49 requests that
     carry a given and a cast Crusader together.
   - Test: `JudgementOfTheCrusader` (two duty paladins at 194 and 161).
5. **Buff plumbing.**
   - `GetRaidBuffs` clones the proto, and calls the hooks once (`raid.go:220,225`).
   - `AddPartyBuffs` runs once: `Character.AddPartyBuffs` becomes empty, and Atiesh moves to
     `addItemPartyBuffs`.
   - Mana Tide pays only the member it's registered on (`buffs.go:1472,1491`). Today a party of 5 gets
     5800 mana each, against 1160.
   - Tests: `RequestUnchanged`, `Atiesh`, `ManaTide`.
6. **Blood Pact to the party.**
   - A new `PartyBuffs.blood_pact` field (`common.proto:829`) is set in a warlock `AddPartyBuffs`, and
     `applyBuffEffects` takes the higher of the raid's and the party's.
   - It needs `make proto` for the UI's TypeScript.
   - Dropping the hook instead would lose a lone warlock its own Stamina.
   - Test: `BloodPact`.
7. **Battle Shout and the exclusive effects.**
   - `NewExclusiveEffect` keeps the first caster's values (`exclusive_effect.go:110-115`). The live case is
     Battle Shout with the 3-piece Wrath bonus.
   - Fix: labels keyed by value at the call sites (`"Battle Shout (145)"`), as Hunter's Mark already is,
     plus a guard that panics when the priority differs. The guard fires in none of the 58 requests, the
     raids of 20 and 40, or the doubles.
   - The given `furious_howl` is read: a permanent Furious Howl at its rank by level, with the rank table
     moved into core.
   - Tests: `BattleShout`, `FuriousHowl`.
8. **The health fight's duration variation.** `DurationVariation` is set to 0 when the fight uses health,
   as the leveling sim already does (`leveling.go:226`). That also drops one random draw an iteration, so
   single-player health fights move. None are in the report.
   - Test: `HealthFightVariation`.

Commits 1 to 3 can go in any order.

**After B:**

- **The report's stale check.** `stale.mjs` will flag every report build, because the `sim/core` and
  `proto` hashes change. The numbers don't move, so refresh `provenance.json`, or re-sim as a confirmation.
- **Two tests already fail on master:** `TestP1Mage` and `TestElementalLevel60` have stale `.results`
  files. They're not B's to fix, but they'll show in the run.

- **Reads:** [Appendix](#appendix-measurements-2026-10-29), [Checking it](#checking-it).
- **Done when:**
  - the Doubles and Size checks pass, with a test for each bug
  - the 58 report requests' hashes equal the baseline

### M0: which roles matter (half a day, a script)

An additive estimate from data the report already has:

- each build's one-buff-at-a-time runs (`abl.json` in `tools/dps_report/raw/<id>.tgz`)
- new single-player sims for the pairs that interact (the minor armor slot, the two air totems)
- the published DPS and the duty variants' cost

It gives each role's value at 10, 20 and 40 and the ranking of the fillers, before any search code is
written. It's also the first version of the search's model.

- **Reads:** [Roles](#roles), [The search](#the-search).
- **Done when:** a table of role gain, provider cost and net, and the filler ranking, is in TASKS.md or a
  note.

### M1: the first answer (a command line search)

Single-player sims of the report's requests, with a buff set written into each one, already work today. So
M1 needs none of B's fixes and no new engine piece.

With one player per request, the resolver writes `Raid.buffs`, `Raid.debuffs`, the party's `buffs` and
the player's `buffs` directly. The per-player `raid_buffs` comes in M1b. The class hooks still run, which
is right for the player's own Moonkin Aura, Leader of the Pack, Trueshot Aura and Blood Pact.

**M1 searches only the role options that exist today.** It leaves out:

- the Tallstrider and the wolf
- the Sunder upkeep warrior, the Curse of Recklessness warlock, and Judgement of Wisdom from a Ret paladin
- a DPS shaman's totem choice: Enhancement and Elemental drop their fixed totems
- Dark Inspiration

A warrior tank gives `sunder_armor` full from the pull, and "none" gives no Sunder. The search stops before
the confirmation, which joins in M1b.

**What M1 builds:**

- **The pool** from `tools/dps_report/req` (see [The build pool](#the-build-pool)).
- **The provider table** as Go, with Judgement of Wisdom and one judgement per paladin.
- **A single-player resolver:** copy the request, then write in its buffs and debuffs per
  [Writing a role into a request](#the-provider-table), with the provider's own effect left off.
- **The search, the cache and the result**, on goroutines, in `tools/compsearch/`.
- **Fight length** from the encounter, at the raid's length.
- **The sensitivity scenarios** that need no sims.

- **Reads:** [Members](#members), [Tank and healer presets](#tank-and-healer-presets),
  [Buffs and debuffs come from the members](#buffs-and-debuffs-come-from-the-members),
  [Assignments](#assignments), [Groups](#groups), [The fight](#the-fight), [The build pool](#the-build-pool),
  [Comp search](#comp-search), [Sensitivity](#sensitivity), [Member value](#member-value).
- **Done when:**
  - check 5 passes
  - a second run makes zero sims
  - changing the encounter, the length, `sim/` or the item database misses the cache
  - given and cast are never both, per the resolver's tests
  - the 20-player result for each faction is written up, with its tie set, role table, filler ranking,
    tank and healer combo ranking and sensitivity

### M1b: whole raids that run right (engine)

This is what the confirmation raid sims need, on top of B.

**The new pieces:**

- `raid_buffs` per player, skipping the class hooks
- the Sunder ramp on the given debuff
- the given forms that are missing: `Debuffs.dust_cloud`, and Dark Inspiration reaching the priest's party
  when it procs, with a given form at the priest's measured uptime for the single-player sims
- a players-only raid aura array
- cast buffs reaching their scope:
  - Battle Shout to the raid
  - the shaman's totems to the raid or party, with a count of standing totems per receiver
  - the weapon totem rules per receiver
  - Blood Pact to the party, through B's field
  - Furious Howl to the raid, and to pets once the launch check says so
- Faerie Fire in Cat and Bear Form as the caster spell, free, with a 1.5 s GCD and no cooldown, replacing
  Faerie Fire (Feral) (the TASKS.md item). Then re-measure the Feral duty cost.
- the missing builds in [Assignments](#assignments), and the per-player assigned-totems option
- the hunter pet variants, as requests in `tools/dps_report/req/`

**Then** the search's confirmation runs whole raids, through the comp proto (`proto/raid_comp.proto`,
below) and one Go call, `RunComp`, which resolves a comp into a `RaidSimRequest` and runs it.

- **Reads:** [Buffs and debuffs come from the members](#buffs-and-debuffs-come-from-the-members),
  [Sunder Armor](#sunder-armor), [Assignments](#assignments), [Confirmation](#confirmation),
  [Checking it](#checking-it), [Appendix](#appendix-measurements-2026-10-29).
- **Done when:**
  - the report raid and Scope checks (1 and 4) pass in `sim/raid_forever_test.go`
  - the 58 single-player requests still give bit-identical results at a fixed seed
  - a comp round-trips through JSON
  - `RunComp` on 7 builds plus 3 placeholders equals a hand-built request, with every DPS member taking
    the raid damage hits

**The comp proto:**

```
RaidComp { size; repeated CompSlot slots; Encounter encounter; repeated Assignment assignments; Faction faction; }
CompSlot { party; index; target_index (0 until several bosses); oneof member { Player player; Placeholder placeholder; } MemberRef ref; }
MemberRef { report_id; source (report, preset, saved, snapshot); name; }   // display and the report comparison only
Placeholder { preset; blessings per class; aura; judgement; totems (earth, air, water, fire); faerie_fire; }
Assignment { duty; holder slot; class (for a blessing); target slot (Power Infusion, Innervate); set_by_hand; }
```

Go never looks up a reference. The page fills `player`, and the tests and the command line fill it from
`req/`.

**The resolver**, in Go:

- **Input:** a `RaidComp`.
- **Output:**
  - the `RaidSimRequest`: each player's `raid_buffs` and `buffs`, the party buffs, the debuffs, and each
    member's base or duty rotation
  - a record of each effect: scope, provider slot, given or cast, receivers
  - the list of who misses which buff

The page reaches it through a new WebAssembly worker call, `resolveComp`.

### M2: the page

- **M2a, the results viewer.** Done when a search's JSON shows as a static page.
- **M2b, the comp editor with Run.** Done when a 20-player raid of report players runs in the WebAssembly
  workers and shows every item under results.
- **M2c, the search button through the native server.** Done when the search streams progress and stops on
  Abort, and the button is off when the page is served as static files.

**Reads:** [The page](#the-page).

### Later

- **Members from the spec pages:**
  - the `raidSim` mark on `PresetBuild`, with `makePresetBuild` passing it through
  - the parity check against `req/`
  - the headless builder
  - "Open in spec page" and "Save to raid", and "Add to raid sim"
  - raid history, saved raids and share links
  - retiring `ui/raid`
- **Speed**, in the profile's order (Appendix), and not registering DoT ranks that can't be cast. The
  search runs single-player sims, so this matters only for big raid sims on the page.
- **Several bosses:**
  - which boss each member attacks
  - per-boss debuffs and execute phases
  - raid damage only from living bosses
  - the Vaelastrasz encounter's `Parties[0].Players[0]`

  No Forever boss data exists before 2026-12-09.
- **Choosing when** for the cooldowns (Power Infusion, Innervate, Mana Tide).
- **Per-member random streams**, behind a `SimOptions` flag.
- **Tanks and healers simmed for real**, and threat.

**Maintenance:**

- The search reads `req/`, so the report's regenerations flow in on their own.
- The regenerate-dps-report and dps-report skills get a step: re-run the default searches when a build is
  re-simmed.
- A committed result is stamped with the report's `datasets.json` stamp and the builds' fingerprints. A
  stale check flags it when they move.

## Decisions

The user's rulings (2026-10-29):

- **The raid**
  - The raid sim is a page of its own, and takes each member's setup whole from the spec.
  - Tank and healer slot counts can be set for each raid size.
  - Everyone is in range of every raid-wide buff.
  - There's no limit on buffs or debuffs per target.
- **Tanks and healers**
  - They aren't simmed and their damage is left out. They give their buffs and debuffs, and the warrior
    tank maintains Sunder after every warrior's pull Sunder.
  - They are a short list of presets, each a sensible raid build with its core role talents. The search
    never builds their talents itself.
  - Every tank and healer combo is simmed and ranked by raid DPS, with each preset's DPS value.
- **The search**
  - The DPS pool is the presets marked for raid sims, for now the DPS Report's builds.
  - There are no limits per spec or class.
  - Groups and assignments are whatever gives the most raid DPS, and every assignment can be automatic or
    set by hand.
  - It runs natively on the user's machine or in Claude Code cloud environments. It judges each build on
    its own under a buff set, and caches every single-player result.
  - The hunter builds get Tallstrider and wolf variants. The pets' effects don't stack with Battle Shout,
    Faerie Fire or Curse of Recklessness until the launch checks.
- **The fight**
  - The length is a fixed duration or the bosses' health, and several bosses come later.
  - Threat is out of scope. When it's added it is reported, not modelled.
- **Scope and order**
  - The DPS Report isn't touched for now.
  - Implementation starts with the bug fixes (B).

**Open:**

1. **New requests and the report.** The pet variants and the missing duty builds are new requests in
   `tools/dps_report/req/`, where the report reads its builds. Proposed: they go there, but the report
   doesn't publish them, so its builds and numbers stay as they are.
2. **Camp buffs.** Forever's campsite features give everyone sitting nearby a 1-hour buff, each one
   "mutually exclusive with" a class buff:

   | Feature | Buff at 60 | Exclusive with |
   |---|---|---|
   | Fish Bowl | 8% all stats | Kings (10%) |
   | Lodestone | 90 AP | Might (133) |
   | Sharpening Wheel | 34 Strength | Strength of Earth (42) |
   | Camp Chair | 2% melee and spell crit | Moonkin Aura |
   | Incense Candle | 25 Intellect | Arcane Intellect (31) |
   | Faction Banner | 32 Spirit | Divine Spirit (40) |
   | Mana Well | 29 mana every 5 s | Wisdom (40) |

   - If a raid has them, a raid with no paladin or shaman still gets about 75 to 80% of Kings, Might and
     Strength of Earth, so those providers are worth much less.
   - The Camp Chair's 2% physical crit is excluded by Moonkin Aura, not by Leader of the Pack, so a Balance
     druid might take 2% crit away from the physical members.
   - Neither the sim nor the report models camp buffs (`settings_tab.ts`: "Off for now"). How the exclusion
     works, and whether a raid can use them before the pull, is unverified.

   Proposed: model them as weaker substitutes, off by default, and run them as a sensitivity scenario. Or
   leave them out entirely?

## Appendix: measurements (2026-10-29)

From the reviews, on builds outside the repo with the protos generated. Nothing here was committed.

### The client's scope rule

For checking a row of the provider table against a new client:

| How the spell is built | Scope |
|---|---|
| Persistent area aura (totem buffs, paladin auras, Trueshot Aura, Moonkin Aura, Leader of the Pack, Blood Pact) | SpellEffect's Effect: 35 is the party, 65 is the raid |
| Cast group buff (Battle Shout, Furious Howl, Gift of the Wild, Arcane Brilliance, Prayer of Fortitude, Prayer of Spirit, the Greater Blessings, Dark Inspiration); Effect is 6, apply aura | `ImplicitTarget_0`: 20 is the caster's party, 56 is the caster's raid, 153 is the raid members of the target's class, 21 or 57 is one target |

- **Pets:** SpellMisc `Attributes_3 & 0x100` marks players only.
- **Several casters:** `Attributes_3 & 0x80` lets copies from different casters stack. Forever added it to
  the debuffs it made per caster (Fire Vulnerability, Winter's Chill, Shadow Weaving, Moonfire, Bane of
  Agony). Hunter's Mark, the Crusader, Mana Spring and Strength of Earth lack it, which is consistent with
  one per target, but the server decides.
- **Totem slots:** a totem summon (Effect 28) names a creature through SummonProperties, whose slot is its
  element (fire 1, water 3, air 4). The client has no link from a totem to the buff spell it casts.
- The client has no stacking or spell-group table.

### What the engine has today

A raid sim already exists, from upstream wowsims: the engine, which this fork uses for every sim, and the
`ui/raid` page. Every sim here is a raid sim of one player plus the assumed buffs, but this fork had never
simmed more than one player.

The 58 report requests are all 1 party × 1 player, with no party buffs, no health fight, and 180 ± 20 s.
Every one gives Blood Pact, all 8 warlocks use the Succubus, and none uses Atiesh or the Wrath set. So
neither Blood Pact nor the double Atiesh moves a report number. 49 carry a given Judgement of the Crusader,
and the paladins in them also cast it.

**What works:**

- `RunRaidSim` builds and runs a raid in parties of 5.
- Metrics add up per member, per party and for the raid; pet damage goes to its owner.
- Each caster's DoTs are its own.
- Shared debuffs dedupe by label, as in game.
- Windfury Totem's cooldown is per receiver.
- Healing Stream already reaches the party.
- Members who don't interact add up (see [Why it doesn't try combinations](#why-it-doesnt-try-combinations)).

**What breaks:**

- **More than 200 auras on the boss panics** (`aura.go:461`). Every caster registers every rank's DoT on
  each target: a fire mage brings 33 auras of its own. A mixed 20 puts 308 on the boss, a mixed 40 546, and
  40 Beast Mastery hunters 296. 6 fire mages fail.
- **Presims for 25 players.** `presim.go:38` indexes a 25-slot list by player index, so any health fight
  with 26 or more players panics.
- **Gaps in a party.**
  - `applyCharacterEffects` and the rotation setup look players up by position in the shortened list.
    With a fire mage in slot 1 and a Fury warrior in slot 3, the mage did 0 DPS and the warrior ran the
    mage's rotation.
  - A missing party shifts everything after it, in `GetUnit` and in the result's parties.
  - A nil party or player panics.
- **Two of one build.**
  - APL references to target auras find the first aura with that spell id (`apl_helpers.go:116-123`). So a
    second fire mage reads the first one's Fire Vulnerability. The same happens with Winter's Chill,
    Shadow Weaving and Light's Vigil.
  - Doubled into one raid, the second member lost DPS: fire mage −28.5, frostfire −29.6, Shock paladin
    −8.9 (from Light's Vigil). Every other base build was within ±2.5.
  - Two Enhancement shamans panic on Stormstrike's fixed label (`shaman/stormstrike.go:47`).
  - The last paladin to register Judgement of the Crusader overwrites the others' refresh hooks, while the
    first one's bonus stays (`aura.go:493-510`).
  - `NewExclusiveEffect` keeps the first caster's values (`exclusive_effect.go:110-115`).
- **Buffs.**
  - `GetRaidBuffs` writes into the request proto.
  - `AddPartyBuffs` runs twice, so Atiesh counts double.
  - Blood Pact goes raid-wide through its hook.
  - A warrior's Battle Shout reaches only their party. When two warriors shout, each receiver keeps the
    first one's value.
  - A shaman's Strength of Earth, Grace of Air, Windfury and Mana Spring reach only the shaman, and the
    weapon totem rules live on the shaman.
  - When one of two shamans' identical totems ends, the buff comes off everyone.
  - Furious Howl reaches only the hunter's party, and its given field is never read.
  - A given Mana Tide pays a party of 5 five times (`buffs.go:1472,1491`).
- **Health fights.** The duration variation still moves the end of fight that rotations see, and draws one
  extra random number an iteration: −0.8% for a 10-member raid, and −1.6% for Destruction.
- **One-player assumptions.**
  - Stat weights, the bulk sim, `sim/lib`, the test suite and the Vaelastrasz encounter all read
    `Parties[0].Players[0]`.
  - So do the external Improved Shadow Bolt and Stormstrike debuffs, which Forever skips anyway, and which
    the resolver never gives.
  - Many settings debuffs go on boss 0 only.

**Known and harmless:**

- `spell_result.go:110,117,197` indexes attack tables by `target.Index`, which for a friendly target is the
  raid index. It makes no DPS difference, and fixing it could move single-player numbers.
- Vampiric Embrace's aura is keyed by target, not caster. It's healing only.

### Speed

On one thread, native, with the aura cap raised:

| Raid | Per iteration | Against the same builds simmed alone |
|---|---|---|
| 1 player | 0.10 to 0.93 ms (mean 0.37) | |
| 10 | 7.0 ms | 1.9× |
| 20 | 19.2 ms | 2.6× |
| 40 | 63.5 ms | 4.3× |

It grows as about 0.34·n + 0.031·n² ms. At 40 the n² part is 80% of the cost.

The 40-player profile, biggest first:

1. **Spell metrics** sized by every unit, reset and scanned every iteration (`spell.go:449`): about 25%.
2. **The idle-rotation shortcut is dead in a raid.** It needs no event anywhere in the sim since the last
   check (`apl_idle.go`), so every idle rotation re-runs its whole priority list every 50 ms: about 18%.
3. **Every mana bar settled** on every time step (`sim.go:807`): 15%.
4. **The pending-action list** is a linear sorted insert (`sim.go:858`): 12%.
5. **Weapon attacks** are all looped on each swing: 11%.

Removing the n² parts should bring 40 players to about 14 ms an iteration and 20 to about 7 ms.

### Random streams

Every member draws from one random stream (`sim.go`), so swapping one member changes every other member's
draws. Measured on a 20-DPS raid, 1000 iterations, one member swapped:

| Random streams | How closely the two raids' DPS move together | Spread of the difference |
|---|---|---|
| independent seeds | 0.02 | 285 DPS |
| the same seed, today | 0.22 | 255 |
| per call site (the labeled streams) | 0.55 | 202 |
| per member, a rough prototype | 0.77 to 0.80 | 135 to 148 |
