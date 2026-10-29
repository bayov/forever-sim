# Raid sim

A sim mode for a whole raid. Every member is simmed together in one fight, and the result is the raid's
total damage. The DPS and PvP modes each sim one player against an assumed raid. Here the other members are
real players.

The point is finding the best comp: which specs to bring, and how many of each, for a raid size with a set
number of tanks and healers.

**How to read this doc:**

- The decisions marked "(the user, 2026-10-29)" are settled. Everything else is a proposal.
- The decisions so far are listed at the end.
- The measurements behind the design are in the [Appendix](#appendix-measurements-2026-10-29).
- An implementer starts at [Milestones](#milestones), which says which sections each step needs.

**History:** the first draft was written on 2026-10-29 from the user's idea in TASKS.md (2026-10-14). Two
rounds of reviews followed the same day:

- Forever's buffs and debuffs, checked against client 70685
- the engine
- the search, with a working build, raids of report builds and measurements
- the page
- whether the engine can express the design
- the doc as a spec
- scope and risk

## What it answers

- **The best comp** (the main question). Given a raid size and the number of tank and healer slots, it finds
  the DPS builds that give the most raid DPS. The total counts pets.
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

Forever has no limit on how many buffs or debuffs a target can carry (the user, 2026-10-29).

**Faction.** Both sides have paladins and shamans, so faction doesn't limit the classes. But Forever has no
cross-faction grouping (`forever-wiki/overview.md`), so a raid's races come from one faction. The search
runs once for each faction. Each build takes its best race in that faction, from the report's race tables.

So a Forever comp is mostly about **which specs, and how many of each**. What is left that is shared:

- raid-wide buffs that stop adding once someone gives them (one Battle Shout, one Moonkin Aura, one Leader
  of the Pack)
- buffs capped per provider (one air totem per shaman, one blessing per paladin for each class, one
  judgement per paladin)
- the shared debuff slots (the minor armor slot holds Faerie Fire, Curse of Recklessness or a Tallstrider's
  Dust Cloud)

Group placement changes mana (Mana Spring, Mana Tide, Blood Pact, Dark Inspiration). It changes damage
only through Retribution Aura and Thorns in the tank's party, which do nothing until tanks are simmed.

**Range.** "Raid-wide" still means within 20 yards (Battle Shout), 30 yards (the totems) or 45 yards
(Moonkin, Leader of the Pack, Trueshot Aura). The raid sim assumes everyone is in range of everything (the
user, 2026-10-29).

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

But a placeholder is a full provider (the user, 2026-10-29). Whatever its spec gives goes to the raid, and
the search counts it as covered at no DPS cost.

- **Paladin tanks and healers** give Greater Blessings and keep a judgement up.
- **Shaman healers** drop a totem for each element and give Mana Tide to their party.
- **A warrior tank** maintains Sunder Armor after the pull.

A placeholder is not a unit in the engine. A unit would join every list sized by the raid (attack tables,
raid damage hits, presims) and do nothing for DPS. The resolver writes what a placeholder gives into the
members' settings, and that is its only effect.

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

The presets, checked against client 70630's trees and 70685's spells. The talent strings are in the sim's
per-tree digit format:

| Preset | Talents | Gives |
|---|---|---|
| Protection warrior | 8/0/43, `35--251533120331221351` (`ui/tank_warrior/presets.ts`) | Sunder Armor upkeep after the pull (see [Sunder Armor](#sunder-armor)); Battle Shout when no DPS warrior holds it |
| Feral tank | 0/41/10, `-55220302020132012551-055` (`ui/feral_tank_druid/presets.ts`) | Leader of the Pack; Faerie Fire; Gift of the Wild. Innervate is off by default, since it means leaving Bear Form. |
| Protection paladin | 1/43/7, `1-5530513321301551-052` (`ui/protection_paladin/presets.ts`) | a Greater Blessing for each class; one aura; one judgement (Wisdom or the Crusader) |
| Holy paladin | 31/20/0, `05320103225101051-500250302003` | the same as the Protection paladin |
| Restoration shaman | 0/8/43, `-005102-5530500315513151` (`ui/restoration_shaman/presets.ts`) | one totem per element (earth, air, water, fire); Mana Spring with Restorative Totems (×1.25); Mana Tide to its party |
| Discipline priest | 33/18/0, `005005231330011531-2310510303` (`ui/healing_priest/presets.ts`) | Fortitude, Divine Spirit, Power Infusion |
| Holy priest | 19/32/0, `00510223123-03405103030121531` (`ui/healing_priest/presets.ts`) | Fortitude, Divine Spirit |
| Restoration druid | 11/0/40, `05302001--5052035153113051` | Gift of the Wild; Innervate; Faerie Fire as an assignment |

**Only the priest's talents change what a tank or healer gives.** Discipline reaches Power Infusion (31
points) and deep Holy can't. No talent in any tree changes a blessing, an aura, a judgement, Battle Shout,
Sunder Armor or a weapon totem. Every "Improved" talent for them is gone:

- Improved Battle Shout, Booming Voice's duration, Improved Sunder's armor
- Improved Seal of the Crusader, Improved Blessing of Might and Wisdom, Lasting Judgement
- Improved Devotion, Concentration and Retribution Aura
- Enhancing Totems and Improved Weapon Totems
- Improved Mark of the Wild, which is baseline now

Restorative Totems and Mana Tide are in every sensible Restoration shaman build. So the two paladin presets
give the same set, and the two priests differ by Power Infusion only.

**Not in a sensible build**, so no preset has them:

- Dark Inspiration, which needs 20 Shadow points and a priest casting Mind Blast
- Moonkin Aura from a healer, since a Moonkin can't heal
- Leader of the Pack from a Restoration druid, since it needs Cat or Bear Form

**Other facts:**

- Divine Spirit, Kings, Innervate, Faerie Fire and Gift of the Wild are trainer spells, so every priest,
  paladin or druid has them.
- The tanks' Demoralizing Shout, Demoralizing Roar and Thunder Clap change no DPS, so they're left out.
- **Faerie Fire (Feral) isn't in client 70685** (spells 16857 and 17390 to 17392 have no rows). Instead,
  the caster Faerie Fire can be cast in Cat and Bear Form, for mana. The raid effect is the same 505 armor,
  and any druid can hold the minor armor slot. The sim still has the free Faerie Fire (Feral) with a
  6-second cooldown (`sim/druid/faerie_fire.go`), which the Feral DPS builds use. That's a TASKS.md item.
- The UI's Holy paladin and both Restoration druid talent strings are invalid in Forever. The table has
  valid ones: the paladin's 3 points in Consecrated Ground moved to Unyielding Faith and Purifying Power,
  and the druid's build was rebuilt on the Thicc Resto layout.

**A slot's choices** are assignments, not talents: which blessings, which judgement, which totems, and
whether it holds Faerie Fire.

**Tank and healer presets are searched too** (the user, 2026-10-29). For the slot counts set, the search
runs every combination of tank and healer presets and ranks them by raid DPS (see
[Tank and healer combos](#tank-and-healer-combos)). The user can also fix any slot's preset. The ranking
counts only what each preset gives the DPS. Nothing checks that a combo can tank or heal the fight, but
every preset is a real tank or healer build.

## Buffs and debuffs come from the members

Nothing is assumed. **The resolver fills the settings with what the members give.** Each buff and debuff
arrives in one of two ways, never both:

- **Given.** The resolver writes it into the settings, so it's up for the whole fight. This covers:
  - the long buffs, which last 60 minutes in Forever and go on before the pull (Fortitude, Arcane
    Brilliance, Gift of the Wild, Divine Spirit, the blessings, the auras)
  - everything a placeholder gives
- **Cast.** A DPS member's rotation casts it, and the engine sends it to its scope. A dropped totem or a late
  debuff then costs the whole raid, as it would in game. A cast comes from one of two places:
  - **The build's own provisions.** Some base rotations already keep up a raid effect, like Moonkin Aura
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

**Members get their own buff set.** Today one `RaidBuffs` covers the whole raid. `PartyBuffs` holds only
Atiesh and Mana Tide. So `Player` gets a new field, `RaidBuffs raid_buffs` (`proto/api.proto`):

- When set, it replaces `Raid.buffs` for that player wholesale. It isn't merged.
- The class `AddRaidBuffs` hooks are skipped for that player, so the resolver owns every buff. A Balance
  druid's own Moonkin Aura then comes from the resolver too.
- Party buffs and debuffs stay where they are.

The same field serves the single-player sims of the search, the report raid check and the page.

**Pets.** Player buffs never reach pets in Forever (standard 14). The engine already skips them
(`applyPetBuffEffects`). Target debuffs do reach pets: armor debuffs raise every pet's damage, and Curse of
the Elements raises the demons' spells. When cast buffs reach the raid (milestone M1b), they need a
players-only raid aura array. `NewRaidAuraArray` today loops every unit, pets included.

### The provider table

The resolver works from a table that says, for each buff and debuff:

- who gives it
- its scope (raid, party, a class across the raid, one target, or the boss)
- whether it reaches pets
- whether several providers stack
- what it excludes
- how long it lasts and whether it costs GCDs in the fight
- the settings field that gives it

**It's written by hand**, about 25 rows, checked against client 70685. The client's scope rule is in the
appendix, for checking a row. `tools/sources/` (update-sources) flags any of the table's spell ids that
change in a new build's wago.tools diff.

DPS-relevant buffs. "Unverified" means the client can't say and the server decides. "No field" means there
is no given form yet:

| Buff | Given by | Scope | Several providers | Excludes | Lasts / in-fight cost | Settings field |
|---|---|---|---|---|---|---|
| Battle Shout (melee AP, rank 6, 115 at 60) | any warrior | raid, 20 yd, players | strongest counts (unverified) | Furious Howl (sim rule) | 3 min; refreshed once in fights over about 177 s | `battle_shout` |
| Furious Howl (melee AP, 83 at 60) | hunter with a wolf | raid, 15 yd | unverified | Battle Shout (sim rule) | the pet's: 60 s, 30 s cooldown | `furious_howl`, never read today |
| Strength of Earth Totem (rank 4) | any shaman, earth | raid, 30 yd, players | doesn't stack (sim); unverified (shaman_audit 8.46) | the shaman's other earth totems | 5 min, 1 s GCD | `strength_of_earth_totem` |
| Grace of Air Totem (rank 2) | any shaman, air | raid, 30 yd, players | doesn't stack | Windfury on the same shaman | 5 min | `grace_of_air_totem` |
| Windfury Totem | any shaman, air | raid, 30 yd, players, main hand only; not for a shaman with Windfury Weapon | doesn't stack | Flametongue Totem; Grace of Air on the same shaman | 5 min; 1.5 s cooldown per receiver | `totem_weapon_buff` |
| Flametongue Totem | any shaman, fire | raid, 30 yd, players | doesn't stack | Windfury Totem; Flametongue Weapon | 5 min; only helps Enhancement (standard 25) | `totem_weapon_buff` |
| Trueshot Aura (50 ranged AP since the 10-28 hotfix) | Marksmanship | raid, 45 yd; only hunters gain | doesn't stack | | permanent | `trueshot_aura` |
| Moonkin Aura (3% spell crit) | Balance | raid, 45 yd, players | doesn't stack | (stacks with Leader of the Pack now) | permanent in form | `moonkin_aura` |
| Leader of the Pack (3% physical crit) | Feral | raid, 45 yd, players | doesn't stack | | permanent in form | `leader_of_the_pack` |
| Greater Blessing of Kings, Might, Wisdom | any paladin (Kings is baseline) | one class across the raid, players | one blessing per paladin for each player (the resolver enforces it; the engine doesn't) | the paladin's other blessings on that class | 60 min | `IndividualBuffs.blessing_of_*` |
| Gift of the Wild | any druid | raid, players | doesn't stack | | 60 min | `gift_of_the_wild` |
| Arcane Brilliance, Divine Spirit (mana) | any mage, any priest | raid | doesn't stack | | 60 min | `arcane_brilliance`, `divine_spirit` |
| Mana Spring Totem | any shaman, water (Restorative Totems) | party, 30 yd | open (shaman_audit 8.45) | the shaman's other water totems | 5 min | `mana_spring_totem` |
| Mana Tide Totem | shaman talent | party | unverified | probably Mana Spring, for its duration: both are water totems (unverified) | 12 s by the tooltip, 13 s in the sim; 5 min cooldown, at min(fight/2, 60 s) today | `PartyBuffs.mana_tide_totems`, which pays a party of 5 five times today |
| Dark Inspiration (+15% Spirit) | Shadow talent, 20 points | party | unverified | | 10 s, set off by the priest's Mind Blast or Mind Flay; only a Shadow priest has it | no field; the priest's own aura only |
| Power Infusion (+20% spell damage) | Discipline, 31 points | one target | unverified | | 15 s, 3 min cooldown, off the GCD; fires as soon as ready | `IndividualBuffs.power_infusions` |
| Innervate | any druid | one target | unverified | | 20 s, 6 min cooldown; below a mana threshold | `IndividualBuffs.innervates` |
| Retribution Aura, Thorns | paladin, druid | the tank's party, the tank | | | only when the boss hits a member: ignored until tanks are simmed | |

Salvation and Mass Salvation are threat only, so they add no DPS while threat is out of scope. The
following add no DPS either:

- Fortitude, Shadow Protection and Blood Pact. The raid damage hits ignore resistances.
- Stoneskin, Windwall, Tranquil Air, Healing Stream, the resistance totems and auras, Devotion and
  Concentration Aura.

The debuffs:

| Debuff | Given by | Several providers | Excludes | Lasts / in-fight cost | Settings field |
|---|---|---|---|---|---|
| Sunder Armor (450 a stack) | any warrior | stacks to 5 from any warrior | Expose Armor (standard 18) | 30 s, a GCD each | `sunder_armor`, full from the pull today |
| Faerie Fire (505 armor) | any druid, in caster, Moonkin, Cat or Bear Form | one per target, from any druid | the minor armor slot | 40 s, for mana | `faerie_fire` |
| Dust Cloud (505 armor) | hunter with a Tallstrider | one | the minor armor slot | the pet's, 30 s | no field |
| Curse of Recklessness | warlock | doesn't stack (70425) | the minor armor slot; the warlock's other curse | 2 min | `curse_of_recklessness` |
| Curse of the Elements (+10%, all magic schools) | warlock | doesn't stack | the warlock's other curse, not its bane | 5 min | `curse_of_elements` |
| Judgement of the Crusader (+161 Holy) | paladin | one (unverified) | the paladin's other judgement | 40 s; about 2 GCDs to seal and re-seal | `judgement_of_the_crusader` |
| Judgement of Wisdom | paladin | one | the paladin's other judgement | 40 s; mana for every attacker with mana | `judgement_of_wisdom` |
| Hunter's Mark (ranged AP for hunters, not pets) | hunter | one per target; the strongest counts | | 2 min, a GCD | `hunters_mark` |

**Judgement of Wisdom matters.** It is worth 56 to 73 DPS to each mana user at 180 seconds in the report's
one-buff-at-a-time runs: Beast Mastery 56, fire 59, Ret Undead 73. That makes it the third-largest effect,
after Sunder and the weapon totem.

A paladin keeps up one judgement. So a raid with one paladin chooses between Wisdom and the Crusader, and
the search treats them as one role with two options.

**Warlock curses.** Forever splits a warlock's curses from its banes (Agony, Doom, Havoc are banes;
warlock.md). So a curse assignment costs GCDs, not the warlock's Agony.

Insect Swarm, Scorpid Sting, Demoralizing Shout and Roar, Thunder Clap and Curse of Weakness only lower the
boss's damage. They add no DPS and are left out.

**Rules the resolver enforces:**

- **One per provider.**
  - A paladin gives each player one blessing. Greater Blessings cover a class, so the assignment is a
    paladin and a class.
  - A paladin keeps up one judgement.
  - A shaman has one totem per element. Grace of Air and Windfury together need two shamans.
  - A warlock has one curse.
- **Shared slots.**
  - The minor armor slot takes one of Faerie Fire, Curse of Recklessness and Dust Cloud (standard 20).
  - Windfury and Flametongue Totem exclude each other.
- **Nobody provides it.** If no member gives a buff, nobody has it. When three paladins can't cover Kings,
  Might and Wisdom for every class, the result shows who misses out.
- **What stacks from several providers** follows the tables above. Where the client can't say (Mana
  Spring, the other totems, Hunter's Mark, Judgement of the Crusader, the pets' effects), the raid sim uses
  the sim's current rule and lists it as an assumption. notes.md (Need to Verify, At 60) has the launch
  checks.

### Sunder Armor

Every warrior's base rotation casts one Sunder on its first GCD. After that, the warrior tank keeps it up
(the user, 2026-10-29). In a raid with a warrior tank, the DPS warriors never sunder again.

The engine doesn't model that today:

- The given Sunder puts all 5 stacks on in the first 0.8 seconds (`debuffs.go:157-170`) and never lets them
  drop.
- So the DPS warriors' pull Sunders always land on a full stack and do nothing.

That overstates physical builds against casters by about 0.4 to 1% of their DPS. A ramp is worth 20 to 35
DPS in a 5-member physical raid, and Sunder is the largest single effect.

**The model** (milestone M1b): a given Sunder with a ramp.

- Each DPS warrior's pull Sunder adds a stack when it lands.
- The tank adds one stack per GCD from the pull, until there are 5.
- From then on the tank keeps it at 5.

The ramp is a setting on the given debuff, so the raid sim and the single-player sims use the same one. A
raid without a warrior tank but with DPS warriors needs a Sunder upkeep build (see
[Assignments](#assignments)).

### Hunter pets

Two pet families give the raid something, and both are substitutes, not extras. Until the launch checks
settle it, the raid sim assumes neither stacks with its player counterpart (the user, 2026-10-29).

- **A Tallstrider's Dust Cloud** takes 505 armor off the boss for 30 seconds. It sits in the minor armor
  slot with Faerie Fire (505) and Curse of Recklessness, and only the strongest counts. So with a druid
  keeping Faerie Fire up, it adds nothing. Its value is in a raid without one: the slot gets filled without
  a druid, or without a warlock spending its curse on Recklessness.
- **A wolf's Furious Howl** gives 83 melee attack power at 60 for 60 seconds, raid-wide within 15 yards,
  on a 30-second cooldown.
  - It shares Battle Shout's slot, and only the stronger counts (`battleShoutCategory`,
    `sim/core/buffs.go`). Battle Shout is 115, so with a warrior shouting, the howl adds nothing to
    players.
  - With warrior tank placeholders, Battle Shout is always covered, so the wolf is worth taking only in a
    raid with no warrior at all.
  - The howl reaches the wolf itself. Whether it reaches other hunters' pets is unverified.

The launch checks are in notes.md (Need to Verify, At 60): what stacks, and whether the pets' buffs reach
other parties and other pets.

Both pets cost the hunter some of their own DPS, since the usual pet is picked for it. Each hunter build
gets marked variants with a Tallstrider and with a wolf (the user, 2026-10-29), with their own report ids.
The variants carry that cost, and the search weighs it against what the pet gives.

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

**Automatic assignment** gives each job to a placeholder first, since that costs no DPS. A tank or healer
who can do it does it. When no placeholder can, the job goes to the eligible DPS member it costs least,
then to the lower slot.

**An assignment's cost** is its build's base DPS minus its duty variant's DPS, both at 180 seconds against
the generic boss. The variant is the build's duty request in the report, with the effect switched off in
its own settings, so it casts the effect itself and pays for it. The numbers come from `builds.json`, and
the raid sim measures them too. Measured: Faerie Fire costs Balance 31.8, Curse of the Elements costs
Destruction 8.9, and Battle Shout costs Fury 1.8.

**A DPS member can only hold an upkeep assignment its build has a duty variant for.** Today there are 12:

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

**The search places groups by the same rule, and the rule is exact** (the user, 2026-10-29: "whatever
maximizes DPS"). Only Mana Spring, Mana Tide, Blood Pact and Dark Inspiration depend on groups, and each
member's gain from them is its own.

1. Each build is simmed twice under the buff set: with its party's package (whichever of those effects the
   set has) and without.
2. Each party with a package has 4 seats beside its provider.
3. The seats go to the members who gain most from the package.
4. The rest fill on their DPS without it.

## The fight

- **The bosses.** Each launch raid gets an encounter preset:
  - **Barrow Deeps (10) and Hyjal Summit (20):** the report's generic level 63 boss (3731 armor, Unknown
    creature type) in the natural flask zone, as in every Level 60 preset (standard 11).
  - **Onyxia's Lair (40):** a Dragonkin boss outside the flask zone. Off the zone, the natural flasks give
    only their Stamina: that costs a Beast Mastery hunter about 38 DPS and a Fury warrior about 64.

  Forever's raid bosses aren't known yet. When they are, each one is its own encounter.
- **The creature type** picks the encounter variants in the pool (vs Humanoid, Undead, Dragonkin): a
  variant joins the pool only when the boss's type matches.
- **Length** (the user, 2026-10-29). It can be set two ways:
  - **A fixed duration**, with the report's 180 ± 20 seconds as the default.
  - **The bosses' health.** The fight ends when every boss is dead, each one dying on its own
    (`sim/core/target_death.go`). A health fight sets the duration variation to 0. Today the variation
    still moves the end of fight every rotation sees (`GetRemainingDuration`), and with ±20 seconds a
    10-member raid lost 0.8%.
- **Length matters a lot.** From 150 to 210 seconds, member DPS falls by different amounts:
  - Fury 6.0%
  - rogue 3.7%
  - Destruction 2.8%
  - the hunters 2.3%
  - fire 1.5%
  - frost and Enhancement 0.8%

  The ranking flips: Fury beats Destruction at 150 seconds and loses at 210. So the search always sims at
  the raid's own length, and the result is checked at 120 and 300 seconds (see
  [Sensitivity](#sensitivity)).
- **Several bosses** (the user, 2026-10-29: "or multiple if multiple exist"). They need engine work (see
  [Milestones](#milestones), Later). Until then, the raid sim fights one boss.
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
variants. The 8 hunter pet variants (4 builds × Tallstrider and wolf) will make it 45.

- **The duty variants** are the base build holding an assignment, not entries of their own (`duty` and
  `baseBuild` in each request's `meta.json`).
- **The encounter variants** join only when the boss's creature type matches.
- **The "+ Mana Tide", Water Shield and Imp GW builds** are different talent builds, so they are entries
  of their own.

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
  and a parity check against `req/` come only when the page takes members from the spec pages (milestone
  M2, Later).

## Comp search

The main thing the raid sim is for (the user, 2026-10-29).

**The inputs:**

- the raid size
- the tank and healer slots, each with a fixed spec or "any" (the default)
- the encounter (boss, creature type, length)
- the faction

**There are no limits per spec or class** (the user, 2026-10-29). The goal is the most raid DPS and
nothing else. The search may fill the raid with one spec, and may drop the only provider of a buff when a
stronger spec in its place gives more. The output says so when the answer is "13 Beast Mastery hunters".

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

- **The buff set:** the provider members (each a build plus what it holds) and the placeholders' choices.
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
| Power Infusion, Innervate, Mana Tide | from the placeholders that have them, fixed, not searched |

Roles are coupled through their providers:

- A shaman covers all four elements.
- A warlock's one curse is either Curse of the Elements or Curse of Recklessness.
- A Balance druid on Faerie Fire covers Moonkin Aura and the minor armor slot.

So a search move adds, removes or swaps one provider member, not one role.

### The search

1. **Start** with each role covered by a placeholder when one can. Otherwise use the provider build with
   the highest DPS under the full set.
2. **Score a buff set by filling it:**
   - The provider members are in. Each one runs its duty variant for what it holds, simmed with that
     effect off in its own settings, so it pays the assignment's cost.
   - The remaining DPS slots take the best builds under the set, repeats allowed. A build that gives a
     role passively (Balance's Moonkin Aura, Feral's Leader of the Pack, Marksmanship's Trueshot Aura, a
     shaman's totems) may only fill a slot when the set has that role, so the comp stays the set it was
     scored under.
   - Groups are placed by the rule in [Groups](#groups).
   - The score is the sum of the members' DPS.
3. **Move.** Add, remove or swap one provider member. Keep any move that raises the score. Stop when no
   move helps. The set of roles is small enough (about 2^10 subsets) that the search can also try them
   all.
4. **Confirm with raid sims.** Every comp whose score is within 0.3% of the best is simmed as a whole raid,
   paired, at 5,000 iterations, and ranked by the raid sim. The winner is published at 20,000 iterations
   (standard 15). The raid sim catches what the single-player sims can't:
   - a member's first curse or Faerie Fire landing after the pull
   - the shared Hunter's Mark
   - Mana Tide's timing
   - the engine's interactions

   If a raid sim disagrees with its score by more than noise, the result shows the gap.

**Fight length in the single-player sims:**

- A fixed fight: the same duration and variation as the raid. Iteration seeding pairs the durations between
  the raid and its single-player sims.
- A health fight: the raid's mean length, with no variation. Simming the length distribution instead added
  only 0.07%. The search starts from health ÷ the predicted raid DPS. It re-sims a buff set's builds when
  the predicted length moves by more than about 3 seconds.

### The cache

**A spec simmed under a buff set it was already simmed under is never simmed again** (the user,
2026-10-29).

- **The key** is the SHA-256 of the deterministic proto marshal of the resolved single-player
  `RaidSimRequest`, plus the sim's version:
  - The request covers the build's whole setup (gear, talents, base or duty rotation, consumes, race), its
    raid buffs, individual and party buffs, debuffs, the encounter, the duration and variation, the
    iterations and the seed.
  - The version is a hash of the inputs that can change a number: `sim/`, the item database and the
    spec's folders. These are the same hashes the report's `tools/dps_report/stale.mjs` uses.

  Two buff sets that give a build the same settings share an entry. Nothing is dropped from the key, such
  as Mana Spring for a warrior. Deciding what can't matter is easy to get subtly wrong. If hit rates turn
  out poor, that can be added later.
- **The value** is the DPS, its standard error and standard deviation, and the mana stats.
- **The cache is kept in memory during a search, and on disk next to the native server** (`compsearch/`
  in its data folder). Two Fury warriors in one comp, or one comp reached two ways, are one sim. Other raid
  sizes, other slot counts and later searches reuse what's there.
- **A code or database change starts a fresh cache.** Old entries are pruned.
- **In a Claude Code cloud environment** the cache starts empty each session, and that's fine. A cold
  search is a few hundred single-player sims at 1,000 iterations: minutes on any machine. The cache isn't
  committed.

### Tank and healer combos

The search runs once for each combination of tank and healer presets that fits the slots (the user,
2026-10-29). The presets (see [Tank and healer presets](#tank-and-healer-presets)):

- **Tanks:** Protection warrior, Feral tank, Protection paladin.
- **Healers:** Holy paladin, Restoration shaman, Discipline priest, Holy priest, Restoration druid.

Combinations are counted as multisets, since two warrior tanks in either order are one combo:

| Raid | Tanks | Healers | Combos |
|---|---|---|---|
| 10 | 1 or 2 | 2 | 3 × 15 + 6 × 15 = 135 |
| 20 | 2 | 4 or 5 | 6 × 70 + 6 × 126 = 1,176 |
| 40 | 3 | 7 to 9 | 10 × (330 + 495 + 715) = 15,400 |

That's cheap, because a combo only changes the buff set. Each combo's placeholder choices (blessings,
judgements, totems) are part of its search. Many combos give the same buffs: the two paladins give the same set,
the two priests differ only by Power Infusion, and a second Restoration druid adds only another
Innervate. So most combos are scored from
the cache without a new single-player sim. At 40, the search drops a combo early when even its best
possible buff set can't reach the best combo found so far.

**What it reports:**

- **The combos ranked by raid DPS**, each with its best DPS comp, and how far it is behind the best.
- **Each tank and healer preset's DPS value:** the raid DPS it adds through what it gives, as the
  contribution of that slot (see [Member value](#member-value)) in each combo's best comp, and on average
  over the combos. This answers which tanks and healers are best for the raid's damage, leaving out
  utility, threat and healing.
- Their own damage isn't in it while tanks and healers aren't simmed. When they are, it adds to the same
  table.

### Cost

- A buff set needs at most one single-player sim per build it hasn't seen, two with a group package: 37
  sims at 1,000 iterations is about 15 seconds on one core, or about 1 second on 16.
- A search visits tens of buff sets, and the cache shrinks that quickly.
- The confirmation is a handful of raid sims: a 20-member raid at 5,000 iterations is about 6 seconds on
  16 cores, and a 40-member raid at 20,000 about 1.5 minutes with today's engine.

### Noise

- At 5,000 iterations a 10-member raid's standard error is 0.021%. With today's correlation of 0.22 between
  two raids on the same seed, their difference has a standard error of about 0.026%, so a 0.1% gap is about
  4σ. For 20 members it's about 5σ.
- The real limit is not noise but how well the single-player sims add up, about 0.1%. That is the same size
  as the gaps between the top comps: swapping one build in a 10-member raid is worth 0.02% to 0.16%. That's
  why the confirmation step sims every comp within 0.3%.
- Seeds s and s + 1 are the same sample shifted by one iteration (iteration i uses seed + i). Replicates
  use seeds further apart than the iteration count.
- Per-member random streams would make whole-raid comparisons sharper. The search doesn't need them; they
  come later.

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

- **For each raid size and faction:** the best comp and 2 to 4 runners-up, with the difference and paired
  95% intervals.
- **The tank and healer combos ranked**, and each tank and healer preset's DPS value (see
  [Tank and healer combos](#tank-and-healer-combos)).
- **The role table:** each role's gain, its provider's cost and the net, at 10, 20 and 40.
- **The filler ranking** under the winning buff set, each build with the spread of the report's
  assumptions (`assumption_dps.json`).
- **Each member's contribution and replacement margin** (see [Member value](#member-value)).
- **The sensitivity matrix** (see [Sensitivity](#sensitivity)).
- **Each member's threat per second.**
- **What isn't modelled:** healing, tanking, threat, pet deaths, movement, boss mechanics.

## Sensitivity

The top comps are close. A rough estimate from the report's one-buff-at-a-time runs puts a hunter-heavy raid
and a caster-heavy raid a few percent apart, about 43 DPS a slot. That is inside the spread of the report's
own open assumptions, up to ±60 to 120 DPS a build. So the result says whether the winner holds under each
of these:

| Scenario | Why | Cost |
|---|---|---|
| Each build shifted by its assumption combos (`assumption_dps.json`) | the report's open rules move builds against each other | no sims |
| 120 and 300 seconds | Fury, rogues and the hunters fall faster than casters with length; hunters run out of mana by 300 | a re-run at each length |
| Each raid's boss: generic in the flask zone; Dragonkin outside it | Onyxia | a re-run |
| Raid damage at 0 and 10 hits a minute | Demonology loses 63 DPS at 10 a minute; Fury 2H gains 56 | from the assumption data |
| Light consumes (no flask) | the flask is worth 112 to a fire mage and 38 to a Beast Mastery hunter | from the one-buff-at-a-time runs |
| Pet uptime about 80%, and Lone Wolf | the pet is 28% of a Beast Mastery hunter's damage and never dies in the sim | a re-run of the hunter builds |
| Each unverified stacking rule on and off | Mana Spring from two shamans, Furious Howl against Battle Shout, Hunter's Mark, the Crusader | a re-run |

When the winner changes under a scenario, the result says so and shows the comps that win as a tie set,
not one answer.

## Member value

All of these come from the single-player sims under buff sets. So they need no raid sims once the cache
has them.

- **Contribution:** raid DPS minus raid DPS without the member. Without them, the comp is resolved again:
  automatic assignments move to the next cheapest holder, and assignments set by hand are dropped. It's
  shown as the member's own DPS plus what they add to everyone else. This is "what a spec brings".
- **Replacement margin:** raid DPS minus the best raid DPS with another build in that slot.
  - For a member that provides nothing, it's their DPS minus the next build's under the same set.
  - For a provider, it comes from the best neighbouring set the search scored with that provider removed
    or swapped.
  - At the best comp it is zero or more, and a margin within noise means two builds are interchangeable.
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

**It doesn't set up a member's gear, talents, rotation or consumes.** Those come whole from the spec
(the user, 2026-10-29).

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
- a member table with each member's difference from the report, shown when the member's request equals
  the report build's
- each buff and debuff with its provider, scope and uptime, from the resolver's record
- members who miss a buff the raid could have given them
- assignment cost and member value

If a change touches shared result views (the events view, the replay), the DPS Report is deployed after it
(CLAUDE.md).

## Checking it

1. **The report raid.** A raid of base report builds, with no duty or encounter variants:
   - Each member gets their own request's raid buffs (the per-player field) and the base requests' common
     debuffs (all 49 non-duty requests carry the same ones).
   - The fight is a fixed 180 ± 20 seconds.
   - Each member lands within 4 × √(se₁² + se₂²) of a single-player sim of the same request, on the same
     commit, at the same seed and iterations. That's 4σ, since 3σ across 37 members fails by chance about
     1 time in 10.
   - Survival and Enhancement next to many physical members are allowed ±5 DPS until their small coupling
     is found.

   It runs on a 10-member raid, a 20-member raid and a physical-heavy raid. It compares with the
   single-player sim, not the published numbers, which come from the report's merged tree with its own
   patches.
2. **Doubles.** Every base build twice in one raid: the second member within noise of the build alone.
3. **Size.** Raids of 20 and 40 build and run without a panic, a 40-player health fight included. A party
   with gaps in slots 1 and 3 gives everyone the right setup.
4. **Scope.**
   - A member gets a party buff only from their own party, and a raid buff from anyone.
   - A cast Battle Shout reaches the raid, and two shouts give the strongest.
   - A dropped totem reaches its scope, and the buff stays while a second identical totem stands.
   - Given and cast never both apply.
5. **The search against brute force.** On a pool of 6 builds and 4 slots, the search finds the same best
   comp as trying every one under the summed single-player model.
6. **After 2026-12-09**, the user's raid logs (`tools/warcraftlogs`, `logcheck`) and public logs are
   checked against the sim: raid DPS by comp and buff uptimes.

## Milestones

Forever launches on 2026-11-04, and the classes will change with it. The raids open on 2026-12-09.

**The plan:**

- M1 working by about 11-07, then run against the launch builds
- M1b by mid-November
- M2a by late November
- a final run with the sensitivity matrix in the week of 12-01
- check 6 against real logs after 12-09

Each step says what it needs from this doc and when it's done.

### M0: which roles matter (half a day, a script)

An additive estimate from data the report already has: each build's one-buff-at-a-time runs (`abl.json` in
`tools/dps_report/raw/<id>.tgz`), the published DPS and the duty variants' cost. It gives each role's value
at 10, 20 and 40 and the ranking of the fillers, before any code is written.

- **Reads:** [Roles](#roles).
- **Done when:** a table of role gain, provider cost and net, and the filler ranking, is in TASKS.md or a
  note.

### M1: the first answer (a command line search, no engine changes)

Single-player sims of the report's requests, with a buff set written into each one, already work today. So
the first answer needs no engine fixes.

- **The pool** from `tools/dps_report/req` (see [The build pool](#the-build-pool)).
- **The provider table** as Go, with Judgement of Wisdom and one judgement per paladin.
- **A single-player resolver**: copy the request, then write in its raid buffs, debuffs and per-class
  blessings, with the provider's own effect left off.
- **The search, the cache and the result**, on goroutines, in `tools/compsearch/`.
- **Fight length** from the encounter, at the raid's length.
- **The sensitivity scenarios** that need no sims.
- **Reads:** [Buffs and debuffs come from the members](#buffs-and-debuffs-come-from-the-members),
  [Assignments](#assignments), [Groups](#groups), [Comp search](#comp-search), [Sensitivity](#sensitivity),
  [Member value](#member-value).
- **Done when:**
  - check 5 passes
  - a second run makes zero sims
  - changing the encounter, the length or the commit misses the cache
  - given and cast are never both, per the resolver's tests
  - the 20-player result for each faction is written up, with its role table, filler ranking, tank and
    healer combo ranking and sensitivity

### M1b: whole raids that run right (engine)

This is what the confirmation raid sims need.

**The engine fixes** (the bugs are in the [Appendix](#appendix-measurements-2026-10-29)):

- the aura cap
- a member's own auras found by owner: APL references prefer the rotation's own unit's aura, personal
  debuffs are keyed by caster, Stormstrike's label, and the Crusader's hook isn't overwritten
- copying in `GetRaidBuffs`, and one `AddPartyBuffs`
- `PartyIndex` and the missing-party index
- the presim list
- the health fight's duration variation

**The new pieces:**

- `raid_buffs` per player, skipping the class hooks
- the Sunder ramp on the given debuff
- the given forms that are missing: read `furious_howl`, add `Debuffs.dust_cloud` and a given Dark
  Inspiration, and give Mana Tide once per party
- a players-only raid aura array
- cast buffs reaching their scope: Battle Shout, the shaman's totems with a count of standing totems per
  receiver, the weapon totem rules per receiver, Blood Pact to the party, and Furious Howl to the raid
- the missing builds in [Assignments](#assignments), and the per-player assigned-totems option
- the hunter pet variants, in the report

**Then** the search's confirmation step runs whole raids, and the comp proto
(`proto/raid_comp.proto`, below) with one Go call, `RunComp`, which resolves a comp into a
`RaidSimRequest` and runs it.

- **Reads:** [Buffs and debuffs come from the members](#buffs-and-debuffs-come-from-the-members),
  [Sunder Armor](#sunder-armor), [Assignments](#assignments), [Checking it](#checking-it),
  [Appendix](#appendix-measurements-2026-10-29).
- **Done when:**
  - checks 1 to 4 pass in `sim/raid_forever_test.go`
  - every one of the 58 single-player requests gives bit-identical results at a fixed seed, so the report
    doesn't move
  - a comp round-trips through JSON
  - `RunComp` on 7 builds plus 3 placeholders equals a hand-built request, with every DPS member taking the
    raid damage hits

**The comp proto:**

```
RaidComp { size; repeated CompSlot slots; Encounter encounter; repeated Assignment assignments; Faction faction; }
CompSlot { party; index; target_index (0 until several bosses); oneof member { Player player; Placeholder placeholder; } MemberRef ref; }
MemberRef { report_id; source (report, preset, saved, snapshot); name; }   // display and the report comparison only
Placeholder { spec; blessings per class; aura; judgement; totems (earth, air, water, fire); restorative_totems; guardian_totems; }
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

- **Members from the spec pages:** the `raidSim` mark on `PresetBuild` with `makePresetBuild` passing it
  through, the parity check against `req/`, the headless builder, "Open in spec page" and "Save to raid",
  "Add to raid sim", raid history, saved raids, share links, and retiring `ui/raid`.
- **Speed**, in the profile's order (Appendix). The search runs single-player sims, so this matters only
  for big raid sims on the page.
- **Several bosses:** which boss each member attacks, per-boss debuffs and execute phases, raid damage only
  from living bosses, and the Vaelastrasz encounter's `Parties[0].Players[0]`. No Forever boss data exists
  before 2026-12-09.
- **Choosing when for the cooldowns** (Power Infusion, Innervate, Mana Tide).
- **Per-member random streams**, behind a `SimOptions` flag.
- **Tanks and healers simmed for real**, and threat.

**Maintenance:**

- The search reads `req/`, so the report's regenerations flow in on their own.
- The regenerate-dps-report and dps-report skills get a step: re-run the default searches when a build is
  re-simmed.
- A committed result is stamped with the report's `datasets.json` stamp and the builds' fingerprints, and a
  stale check flags it when they move.

## Decisions

The user's rulings so far (2026-10-29):

- Tanks and healers aren't simmed and their damage is left out, but they give their buffs and debuffs, and
  the warrior tank maintains Sunder after every warrior's pull Sunder.
- The DPS pool is the presets marked for raid sims, for now the DPS Report's builds.
- There are no limits per spec or class.
- Groups and assignments are whatever gives the most raid DPS.
- Everyone is in range of every raid-wide buff.
- The hunter builds get Tallstrider and wolf variants, and the pets' effects don't stack with Battle Shout,
  Faerie Fire or Curse of Recklessness until the launch checks.
- The search runs natively on the user's machine or in Claude Code cloud environments, judges each build
  on its own under a buff set, and caches every single-player result.
- Every tank and healer combo is simmed and ranked by raid DPS, with each tank and healer preset's DPS
  value.
- Tanks and healers are a short list of presets, each a sensible raid build with its core role talents,
  and the search never builds their talents itself.
- The DPS Report isn't touched for now.

None open.

## Appendix: measurements (2026-10-29)

From the reviews, on builds outside the repo with the protos generated. Nothing here was committed.

### The client's scope rule

For checking a row of the provider table against a new client:

| How the spell is built | Scope |
|---|---|
| Persistent area aura (totem buffs, paladin auras, Trueshot Aura, Moonkin Aura, Leader of the Pack, Blood Pact) | SpellEffect's Effect: 35 is the party, 65 is the raid |
| Cast group buff (Battle Shout, Furious Howl, Gift of the Wild, Arcane Brilliance, Prayer of Fortitude, the Greater Blessings, Dark Inspiration); Effect is 6, apply aura | `ImplicitTarget_0`: 20 is the caster's party, 56 is the caster's raid, 153 is the raid members of the target's class, 21 or 57 is one target |

- **Pets:** SpellMisc `Attributes_3 & 0x100` marks players only.
- **Radius:** SpellRadius.
- **Totems:** a totem summon (Effect 28) names a creature, and the client has no link from it to the buff
  spell the totem casts.

### What the engine has today

A raid sim already exists, from upstream wowsims: the engine, which this fork uses for every sim, and the
`ui/raid` page. Every sim here is a raid sim of one player plus the assumed buffs, but this fork had never
simmed more than one player.

**What works:**

- `RunRaidSim` builds and runs a raid in parties of 5.
- Metrics add up per member, per party and for the raid; pet damage goes to its owner.
- Each caster's DoTs are its own.
- Shared debuffs dedupe by label as in game.
- Windfury Totem's cooldown is per receiver.
- Healing Stream already reaches the party.
- Members who don't interact add up (see [Comp search](#why-it-doesnt-try-combinations)).

**What breaks:**

- **More than 200 auras on the boss panics** (`aura.go:461`). Every caster registers its own per-rank DoT
  and debuff auras on the boss, about 33 for a fire mage. 6 fire mages fail, and so do a mixed 20-player
  raid of report builds and 40 Beast Mastery hunters.
- **Presims for 25 players.** `presim.go` indexes a 25-slot list by player index, so a 40-player health
  fight panics.
- **Gaps in a party.**
  - `applyCharacterEffects` and the rotation setup look players up by position in the shortened list.
    With a fire mage in slot 1 and a Fury warrior in slot 3, the mage did 0 DPS and the warrior ran the
    mage's rotation.
  - A missing party shifts everything after it.
  - A nil player panics.
- **Two of one build.**
  - APL references to target auras find the first aura with that spell id (`apl_helpers.go:118`). So a
    second fire mage reads the first one's Fire Vulnerability. The same happens with Winter's Chill,
    Shadow Weaving and Light's Vigil.
  - Doubled into one raid, the second member lost DPS: fire mage −27, frostfire −26, Shock paladin −7 (−14
    in its duty and Undead variants). Every other base build was within ±2.5.
  - Two Enhancement shamans panic on Stormstrike's fixed label (`shaman/stormstrike.go:46`).
  - The last paladin to register Judgement of the Crusader overwrites the others' refresh hooks
    (`aura.go:493-510`). `NewExclusiveEffect` keeps the first caster's values.
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
  - A given Mana Tide pays a party of 5 five times (`buffs.go:484`).
- **Health fights.** The duration variation still moves the end of fight that rotations see: −0.8% for a
  10-member raid, and −1.6% for Destruction.
- **One-player assumptions.**
  - Stat weights, the bulk sim, `sim/lib`, the test suite, the Vaelastrasz encounter and the external
    Improved Shadow Bolt and Stormstrike debuffs all read `Parties[0].Players[0]`.
  - `spell_result.go` reads attack tables by target index.
  - Many settings debuffs go on boss 0 only.

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
