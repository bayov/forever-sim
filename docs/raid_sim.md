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

**History:** first drafted on 2026-10-29 from TASKS.md's 2026-10-14 idea. It was reviewed four times the
same day, for:

- Forever's buffs against client 70685
- the engine
- the search, against real raids
- the page
- the doc as a spec
- scope
- the bug fixes, done in a scratch copy
- a last round with every reviewer starting cold: the buffs against the client, the search's statistics,
  the plan against the code, the presets' talents, and the doc read whole

**Terms:**

- **Standard N:** a rule in `orchestration/standards.md`, which every Level 60 preset follows.
- **The buff set:** what the raid has up. It's the placeholders' choices plus the provider members and what
  they hold (see [Why it doesn't try combinations](#why-it-doesnt-try-combinations)).
- **A provider member:** a DPS member in the comp for what it gives. **A filler:** one there for its own
  damage only.
- **A duty variant:** a build's request that holds an assignment (see [Assignments](#assignments)).
- **A package:** the party effects one party has (see [Groups](#groups)).
- **A signature:** what a tank and healer combo's placeholders can give (see
  [Tank and healer combos](#tank-and-healer-combos)).
- **The Crusader:** Judgement of the Crusader.
- **The flask zone:** where Forever's natural flasks work in full (standard 11).

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
cross-faction grouping (`forever-wiki/overview.md`), so a raid's races come from one faction. A search run
covers both factions (see [Comp search](#comp-search)). Each build takes its best race in that faction,
from the report's race tables under the full buff set (see [The build pool](#the-build-pool)). That's an
assumption: in a raid short on mana, another race could overtake.

**So a Forever comp is mostly about which specs, and how many of each.** What is left that is shared:

- raid-wide buffs that stop adding once someone gives them (one Battle Shout, one Moonkin Aura, one Leader
  of the Pack)
- buffs capped per provider (one air totem per shaman, one blessing per paladin for each player, one
  judgement per paladin)
- the shared debuff slots (the minor armor slot holds Faerie Fire, Curse of Recklessness or a Tallstrider's
  Dust Cloud)

**Group placement** changes mana: Mana Spring, Mana Tide and Dark Inspiration. It changes damage directly
only through Retribution Aura and Thorns in the tank's party, which do nothing until tanks are simmed.

**Range.** "Raid-wide" still means within 15 yards (Furious Howl), 20 yards (Battle Shout), 30 yards (the
totems) or 45 yards (Moonkin, Leader of the Pack, Trueshot Aura). The raid sim assumes everyone is in range
of everything (the user, 2026-10-29).

## The score

**Raid DPS: all damage the raid deals to the bosses, divided by the fight length.** Bigger is better.

Everything else a result holds is listed in [The result](#the-result).

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
healing (the user, 2026-10-29). Their damage is left out with no stand-in. So comps are only compared at
the same slot counts: fewer tanks or healers means more DPS slots, which would always win. The slot counts
must leave at least one DPS slot. The presets' own damage differs too (a Protection warrior hits harder
than a Restoration druid), so tanks and healers are ranked by what they give, not by what they deal.

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
| Warden shaman (tank) | 17/34/0, `35050013-0502303313551102` (`ui/warden_shaman/presets.ts`) | one totem per element (earth, air, water, fire); plain Mana Spring to its party |
| Holy paladin | 31/20/0, `05320103225101051-500250302003` | the same as the Protection paladin |
| Restoration shaman | 0/8/43, `-005102-5530500315513151` (`ui/restoration_shaman/presets.ts`) | one totem per element (earth, air, water, fire); Mana Spring with Restorative Totems (×1.25); Mana Tide to its party |
| Discipline priest | 33/18/0, `005005231330011531-2310510303` (`ui/healing_priest/presets.ts`) | Fortitude, Prayer of Spirit, Power Infusion |
| Holy priest | 19/32/0, `00510223123-03405103030121531` (`ui/healing_priest/presets.ts`) | Fortitude, Prayer of Spirit |
| Restoration druid | 11/0/40, `05302001--5052035153113051` | Gift of the Wild; Innervate; Faerie Fire as an assignment |

**Within a role, only the priest has a raid-relevant talent choice.** Leader of the Pack, Restorative
Totems and Mana Tide are talents too, but every sensible build of their role takes them. Discipline reaches
Power Infusion (31 points) and deep Holy can't. No talent in any tree changes a blessing, an aura, a judgement, Battle Shout,
Sunder Armor or a weapon totem. Every "Improved" talent for them is gone:

- Improved Battle Shout. Booming Voice only widens shouts now, and Improved Sunder only cuts its Rage cost.
- Improved Seal of the Crusader, Improved Blessing of Might and Wisdom, Lasting Judgement.
- Improved Devotion, Concentration and Retribution Aura, and Sanctity Aura.
- Enhancing Totems and Improved Weapon Totems.
- Improved Mark of the Wild, which is baseline now.

So the two paladin presets give the same set, and the two priests differ by Power Infusion only. The Holy
priest gives the Discipline priest's set less Power Infusion, so it never ranks ahead; it stays for a user
who fixes a slot. The two shamans give the same totems and differ in their package: Mana Spring ×1.25 and
Mana Tide from the Restoration shaman, plain Mana Spring from the Warden.

**Not in a sensible build**, so no preset has them:

- Dark Inspiration, which sits in Shadow's fourth row (16 points for one rank, 20 for the full 15%) and procs only from the priest's Mind Blast or a fully
  channeled Mind Flay
- Moonkin Aura from a healer, since a Moonkin can't heal
- Leader of the Pack from a Restoration druid, since it needs Cat or Bear Form

**Other facts:**

- Kings, Innervate, Faerie Fire and Prayer of Spirit are trainer spells. Gift of the Wild, Prayer of
  Fortitude and Arcane Brilliance are taught by books (Book: Gift of the Wild, Codex: Prayer of Fortitude,
  Tome of Arcane Brilliance), and the raid sim assumes every member has them. Without the book, the
  single-target spell gives the same values for 60 minutes, cast on each player.
- The tanks' Demoralizing Shout, Demoralizing Roar and Thunder Clap change no DPS, so they're left out.
- **Faerie Fire (Feral) is gone from Forever.** Spells 16857 and 17390 to 17392 have no rows in 70338,
  70425, 70630 or 70685, and foreverchanges.pro says "Not in Forever". The caster Faerie Fire can be cast
  in Cat, Bear and Dire Bear Form, and costs nothing there (dev notes). forever-wiki/classes/druid.md:54
  and notes.md misread this.
- **Fixed in M1b:** under Forever every form casts Faerie Fire 9907, free in Cat and Bear Form, on a
  1.5 s GCD with no cooldown (`sim/druid/faerie_fire.go`). The Feral duty build's cost went from 4.5 to
  8.3 at 180 s (10.2 at 120 s, 4.3 at 300 s).
- **The UI's Holy paladin and Restoration druid presets were invalid in Forever**, and now have the
  table's strings (2026-10-30): the paladin's 3 points in Consecrated Ground moved to Unyielding Faith and
  Purifying Power, the druid's build was rebuilt on the Thicc Resto layout, and its second preset, built
  on Celestial Focus (gone from Forever), is now Furor Resto.
- **Small realism nits** that don't change what the presets give: the Holy paladin takes Improved Righteous
  Fury, the paladins skip Pursuit of Justice, the Feral tank skips Feral Charge and Shredding Attacks, and a
  few points are filler (the warrior's Iron Will, the Protection paladin's Holy point, the druid's Moonglow,
  the Restoration shaman's Guardian Totems over Ancestral Healing).

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
  - **a DPS member's held auras:** Balance's Moonkin Aura, Feral's Leader of the Pack and Marksmanship's
    Trueshot Aura. The engine gives these as static buffs from the class hooks (`sim/druid/druid.go:138`,
    `sim/hunter/hunter.go:149`), not as casts. The resolver gives them, credited to that member.
- **Cast.** A DPS member's rotation casts it, and the engine sends it to its scope. A dropped totem or a late
  debuff then costs the whole raid, as it would in game. A cast comes from one of two places:
  - **The build's own provisions.** Some base rotations already keep up a raid effect:
    - every warrior casts one Sunder Armor on its first GCD
    - Enhancement drops its totems before the pull (Strength of Earth, Windfury, Mana Spring)
    - Elemental drops Mana Spring
    - every hunter keeps Hunter's Mark up
    - the "+ Mana Tide" shaman builds drop Mana Tide, which already pays the shaman's party
      (`sim/shaman/water_totems.go`)

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

**Members get their own buff set** (milestone M1b, done). One `RaidBuffs` used to cover the whole raid, so
`Player` has `RaidBuffs raid_buffs` (`proto/api.proto`):

- When set, it replaces `Raid.buffs` for that player wholesale. It isn't merged.
- The class `AddRaidBuffs` hooks are skipped for that player, so the resolver owns every buff. That's why
  the held auras are given.
- Party buffs and debuffs stay where they are.

The resolver's raids (the confirmation, RunComp, the page) and the report raid check use it. The search's
single-player sims write `Raid.buffs`, which works the same for one player, and there the class hooks still
run (a Balance druid's own Moonkin Aura comes from its hook).

**Pets.** Standard 14 blocks every player buff on pets, and the engine skips them (`applyPetBuffEffects`).
Some buffs lack the client's players-only flag: Furious Howl, Trueshot Aura, Mana Spring, Mana Tide, Blood
Pact and Dark Inspiration. Of those, only Furious Howl and the mana totems (a demon's mana) could change
DPS. They're unverified exceptions, and until the launch checks they don't reach pets (see
[Hunter pets](#hunter-pets)). Target debuffs do reach pets: armor debuffs raise every pet's damage, and
Curse of the Elements raises the demons' spells.

Cast buffs reach the raid through a players-only raid aura array (M1b), since `NewRaidAuraArray` loops every
unit, pets included.

### The provider table

The resolver works from a table that says, for each buff and debuff:

- who gives it
- its scope (raid, party, a class across the raid, one target, or the boss)
- whether it reaches pets
- whether several providers stack
- what it excludes
- how long it lasts and whether it costs GCDs in the fight
- the settings field that gives it, and its spell ids

**It's written by hand**, checked against client 70685. The client's scope rule is in the appendix, for
checking a row.

**In Go** (as built) the table is spread over `sim/compsearch`: a `Gift` bit for each raid-wide effect
(`effects.go`), what each build gives and casts itself (`PlayerGifts`: its class, its held auras from the
class hooks, its duty, its pet), what each preset gives (`presets.go`), and `Seat.Write`, which writes a
member's effects into its request's settings fields. The scopes, the stacking and the exclusions live in
the engine (M1b). Nothing yet flags a spell id that changes in a new client build.

DPS-relevant buffs. "Unverified" means the client can't say and the server decides. "No field" means there
is no given form yet:

| Buff | Given by | Scope | Several providers | Excludes | Lasts / in-fight cost | Settings field |
|---|---|---|---|---|---|---|
| Battle Shout (melee AP, rank 6, 115 at 60) | any warrior | raid, 20 yd, players | strongest counts (unverified) | Furious Howl (sim rule) | 3 min; refreshed once in fights over about 177 s | `battle_shout` |
| Furious Howl (melee AP, 83 at 60) | hunter with a wolf | raid, 15 yd; no players-only flag, so probably pets too | unverified | Battle Shout (sim rule) | the pet's: 60 s, 30 s cooldown | `furious_howl` (read since B) |
| Strength of Earth Totem (rank 4) | any shaman, earth | raid, 30 yd, players | doesn't stack (sim); unverified (shaman_audit 8.46) | the shaman's other earth totems | 5 min, 1 s GCD | `strength_of_earth_totem` |
| Grace of Air Totem (rank 2) | any shaman, air | raid, 30 yd, players | doesn't stack (settled, dev notes) | Windfury on the same shaman (air slot) | 5 min | `grace_of_air_totem` |
| Windfury Totem | any shaman, air | raid, 30 yd, players, main hand only; not for a shaman with Windfury Weapon | doesn't stack (settled, dev notes) | Flametongue Totem; Grace of Air on the same shaman | 5 min; 1.5 s cooldown per receiver | `totem_weapon_buff` |
| Flametongue Totem | any shaman, fire | raid, 30 yd, players | doesn't stack | Windfury Totem on the same receiver; Flametongue Weapon; the shaman's other fire totems (so Enhancement loses Magma and Elemental loses Searing) | 5 min; only helps Enhancement (standard 25) | `totem_weapon_buff` |
| Trueshot Aura (50 ranged AP since the 10-28 hotfix) | Marksmanship | raid, 45 yd; only hunters gain | doesn't stack | | 30 min, before the pull | `trueshot_aura` |
| Moonkin Aura (3% spell crit) | Balance | raid, 45 yd, players | doesn't stack | (stacks with Leader of the Pack now) | permanent in form | `moonkin_aura` |
| Leader of the Pack (3% physical crit) | Feral | raid, 45 yd, players | doesn't stack | | permanent in form | `leader_of_the_pack` |
| Greater Blessing of Kings, Might, Wisdom | any paladin (Kings is baseline) | one class across the raid, players | one blessing per paladin for each player (settled, tooltip; the resolver enforces it, the engine doesn't) | the paladin's other blessings on that class | 60 min | `IndividualBuffs.blessing_of_*` |
| Gift of the Wild | any druid | raid, players | doesn't stack | | 60 min | `gift_of_the_wild` |
| Arcane Brilliance | any mage | raid | doesn't stack | | 60 min; mana | `arcane_brilliance` |
| Prayer of Spirit (Divine Spirit is one target) | any priest | raid | doesn't stack | | 60 min; mana | `divine_spirit` |
| Mana Spring Totem | any shaman, water; ×1.25 with Restorative Totems | party, 30 yd | open (shaman_audit 8.45) | the shaman's other water totems, Mana Tide included | 5 min | `mana_spring_totem` |
| Mana Tide Totem | Restoration shaman; the "+ Mana Tide" builds | party | unverified | the shaman's own Mana Spring (settled: the same water slot). Dropping Mana Tide destroys it, and it must be dropped again. Another shaman's Mana Spring stays. | 12 s by the tooltip, 13 s in the sim; 5 min cooldown, at min(fight/2, 60 s) today | `PartyBuffs.mana_tide_totems` (it paid a party of 5 five times before B) |
| Dark Inspiration (+15% Spirit) | Shadow talent, row 4 (16 points; 20 for 15%) | party | unverified | | 10 s, from the priest's Mind Blast or a fully channeled Mind Flay; only a Shadow priest has it | `PartyBuffs.dark_inspiration` (M1b; a given form at 99.9% uptime), not yet a package in the search |
| Power Infusion (+20% magic damage) | Discipline, 31 points | one target | unverified | Arcane Power on the target (an Era server rule, unverified) | 15 s, 3 min cooldown, off the GCD; fires as soon as ready | `IndividualBuffs.power_infusions` |
| Innervate | any druid | one target | unverified | | 20 s, 6 min cooldown; below a mana threshold | `IndividualBuffs.innervates` |
| Retribution Aura, Thorns | paladin, druid | the tank's party, the tank | | | only when the boss hits a member: ignored until tanks are simmed | |

Salvation and Mass Salvation are threat only, so they add no DPS while threat is out of scope. The
following add no DPS either:

- Fortitude, Shadow Protection and Blood Pact. Forever's Blood Pact is Stamina only, and the raid damage
  hits ignore resistances.
- Stoneskin, Windwall, Tranquil Air, Healing Stream, the resistance totems and auras, Devotion and
  Concentration Aura.

**Camp buffs** stand in for Kings, Might, Strength of Earth, Moonkin Aura, Arcane Brilliance, Divine Spirit,
Wisdom, Gift of the Wild and Fortitude when the raid lacks them (see [Camp buffs](#camp-buffs)).

The debuffs:

| Debuff | Given by | Several providers | Excludes | Lasts / in-fight cost | Settings field |
|---|---|---|---|---|---|
| Sunder Armor (450 a stack) | any warrior | stacks to 5 from any warrior | Expose Armor (standard 18) | 30 s, a GCD each | `sunder_armor`, full from the pull until the ramp (M1) |
| Faerie Fire (505 armor) | any druid, in caster, Moonkin, Cat, Bear or Dire Bear Form | one per target, from any druid (settled, dev notes) | the minor armor slot | 40 s; 115 mana in caster or Moonkin Form, free in Cat and Bear; 1.5 s GCD, no cooldown | `faerie_fire` |
| Dust Cloud (505 armor) | hunter with a Tallstrider | one | the minor armor slot | the pet's, 30 s | `Debuffs.dust_cloud` (M1b) |
| Curse of Recklessness | warlock | doesn't stack (70425) | the minor armor slot; the warlock's other curse | 2 min | `curse_of_recklessness` |
| Curse of the Elements (+10%, all magic schools) | warlock | doesn't stack | the warlock's other curse, not its bane | 5 min | `curse_of_elements` |
| Judgement of the Crusader (+161 Holy) | paladin | one (client-consistent, server-side) | the paladin's other judgement | 40 s; about 2 GCDs to seal and re-seal | `judgement_of_the_crusader` |
| Judgement of Wisdom | paladin | one | the paladin's other judgement | 40 s; mana for every attacker with mana | `judgement_of_wisdom` |
| Hunter's Mark (ranged AP for hunters, not pets) | hunter | one per target, the strongest counts (client-consistent, server-side) | | 2 min, a GCD | `hunters_mark` |

**Judgement of Wisdom matters.** It is worth 56 to 73 DPS to each mana user at 180 seconds in the report's
one-buff-at-a-time runs: Beast Mastery 56, fire 59, Ret Undead 73. That makes it the third-largest effect,
after Sunder and the weapon totem. A paladin placeholder or a Ret member on its Wisdom duty (Seal of Wisdom,
M1b) gives it.

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
them moves any base build (a done-when of M1).

**Rules the resolver enforces:**

- **One per provider.**
  - A paladin gives each player one blessing. Greater Blessings cover a class, so the choice is a paladin
    and a class.
  - A paladin keeps up one judgement and holds one aura.
  - A shaman has one totem per element. Grace of Air and Windfury together need two shamans.
  - A warlock has one curse.
- **Shared slots.**
  - The minor armor slot takes one of Faerie Fire, Curse of Recklessness and Dust Cloud (standard 20).
  - A receiver takes one weapon totem: Flametongue for Enhancement, whose Windfury Weapon makes Windfury
    Totem useless to it, and Windfury for the others. So two shamans can have both up (standard 25).
- **Nobody provides it.** If no member gives a buff, nobody has it. When two paladins can't cover Kings,
  Might and Wisdom for every class, the result shows who misses out.
- **What stacks from several providers** follows the tables above. Where the client can't say, the raid sim
  uses the sim's current rule and lists it as an assumption: Mana Spring, the other totems, Hunter's Mark,
  Judgement of the Crusader, the pets' effects, and Power Infusion against Arcane Power. notes.md (Need to
  Verify, At 60) has the launch checks.

### Sunder Armor

Every warrior's base rotation casts one Sunder on its first GCD. After that, the warrior tank keeps it up
(the user, 2026-10-29). In a raid with a warrior tank, the DPS warriors never sunder again.

The old given Sunder didn't model that:

- It puts all 5 stacks on in the first 0.8 seconds and never lets them drop (it's still what `sunder_armor`
  does, for the report's requests).
- So the DPS warriors' pull Sunders always land on a full stack and do nothing.

That overstates physical builds against casters by about 0.4 to 1% of their DPS. A ramp is worth 20 to 35
DPS in a 5-member physical raid, and Sunder is the largest single effect.

**The model** (milestone M1) is a given Sunder with a ramp:

- Each DPS warrior still casts its own pull Sunder. In the raid sim that cast adds its stack. In a warrior's
  single-player sim its own cast adds its stack, and the ramp adds the other warriors'.
- The tank adds one stack per GCD from the pull, until there are 5, and keeps it at 5.
- With no warrior tank, the pull stacks fall off 30 seconds after the last one, unless a Sunder upkeep
  warrior holds it.

The ramp is a setting on the given debuff (the other warriors' pull Sunders, and whether a tank keeps it
up), so the raid sim and the single-player sims use the same one. The number of DPS warriors' pull Sunders
is part of the buff set. A raid without a warrior tank but with DPS
warriors needs a Sunder upkeep build (see [Assignments](#assignments)).

A Sunder upkeep warrior's own stacks are the ramp, so it isn't counted among the pull Sunders too. Its
ramp is still the tank's, one stack per GCD, though a DPS warrior with little rage at the pull stacks
slower. The second audit measured the other members 4.8 DPS high each under a Sunder upkeep warrior (a
Beast Mastery hunter 897.5 alone against 892.7 in the raid). Counting it twice was 1.8 of that and is
fixed; the other 3.0 (about 0.3%) is the slower ramp. The confirmation's raid sims carry it, and a ramp
measured from the duty build would remove it.

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
  shouting. The cast one reaches the raid's players (M1b), not other pets.

The launch checks are in notes.md (Need to Verify, At 60): what stacks, and whether the pets' buffs reach
other parties and other pets.

Both pets cost the hunter some of their own DPS, since the usual pet is picked for it. Each hunter build
with a pet (Beast Mastery, Marksmanship, Survival) gets variants with a Tallstrider and with a wolf (the
user, 2026-10-29), with their own ids. The variants carry that cost, and the search weighs it against what
the pet gives. Lone Wolf has no pet.

The variants (M1b, `hunter__<build>-tallstrider` and `-wolf`, with a `pet` field in their meta) cost:

| Build | Tallstrider | Wolf |
|---|---|---|
| Beast Mastery (1131.8) | 71.0 | 57.4 |
| Marksmanship (1115.7) | 40.2 | 32.6 |
| Survival (1028.9) | 41.1 | 33.2 |

Only the pet type changes; the pet keeps the cat's 1.0 s swing and Cobra Reflexes, an assumption. In the
search a Tallstrider hunter gives the minor armor slot, which the others get as a given 505 armor in that
slot (`Debuffs.dust_cloud` exists too, for a single-player sim of a hunter's own Dust Cloud), and a wolf
hunter gives Furious Howl, a role of its own in Battle Shout's slot.

### Camp buffs

Forever's camps give everyone who sits at the campfire a 1-hour buff for each feature placed there. There
are 9 features with a buff, one per profession. Each one is "mutually exclusive with" a class buff.

| Profession | Feature | Buff at 60 | Exclusive with (class value at 60) |
|---|---|---|---|
| Fishing | Fish Bowl | 8% all stats | Kings (10%) |
| Mining | Lodestone | 90 melee AP | Might (133) |
| Blacksmithing | Sharpening Wheel | 34 Strength | Strength of Earth (42, rank 4) |
| Skinning | Camp Chair | 2% melee and ranged crit and 2% spell crit, one aura | Moonkin Aura |
| Herbalism | Incense Candle | 25 Intellect | Arcane Brilliance (31) |
| Tailoring | Faction Banner | 32 Spirit | Divine Spirit (40) |
| Alchemy | Mana Well | 29 mana every 5 s | Blessing of Wisdom (40) |
| Enchanting | Enchanted Lute | 308 armor, 13 all stats, 22 resistances | Gift of the Wild (385 / 16 / 27) |
| First Aid | First Aid Kit | 56 Stamina | Fortitude (70) |

These are the client's values in each buff spell (1.60.1.70685). Leatherworking, Engineering and Cooking
features give no stat buff. No camp buff gives hit, haste, spell power or a damage percentage.

**Facts** (client and guides, checked 2026-10-29):

- **A feature's higher tiers don't change its buff.** The tier 2 and 3 versions "provide all of the
  benefits of" the tier 1 one, with one buff spell per feature.
- **The same feature twice adds nothing.**
- **How many features a camp holds:** 3, 5 or 10, by campfire. The Basic Campfire holds 3 and comes with
  Cooking 1. The Journeyman holds 5 and needs Cooking 140 and a Blueprint from Wailing Caverns. The Expert
  holds 10 and needs Cooking 220; its Blueprint has no known source yet.
- **Each player places one feature an hour,** so each feature needs a different player with that
  profession at 20.
- **There's no per-character cap.** A character can hold all 9 at once. Guides say sitting at a second camp
  adds its buffs without losing the first's. That's server-side and unverified, but if it holds, three
  Basic campfires give all 9.
- **No camp inside a raid.** Campfires are outdoors only, and their spells carry the "not in raid
  instances" flag. The raid sim assumes the buffs taken outside carry in. The client doesn't settle it: the
  world buffs have no zone flag either, yet their tooltips say "No effect while in a raid instance" (22888,
  16609, 15366). They last 1 hour (1.5 with the Legacy perk Permanence), can't be refreshed inside, and are
  probably lost on death.
- **"Mutually exclusive" is the server's.** No client table links a camp buff to its class buff. At 60
  every class buff is the stronger one, so "the stronger wins" and "the class buff wins" give the same
  result. The exception is the Camp Chair against Moonkin Aura.
- **The tooltips name the single-target spells:** Blessing of Kings, Blessing of Might, Arcane Intellect,
  Divine Spirit, Blessing of Wisdom, Mark of the Wild and Power Word: Fortitude. A raid uses the group
  versions (the Greater Blessings, Arcane Brilliance, Prayer of Spirit, Gift of the Wild, Prayer of
  Fortitude), which have other spell ids. The raid sim assumes they exclude the same way.

**The rule** (the user, 2026-10-29):

- A camp buff doesn't stack with its class alternative.
- When the raid lacks a class buff, its camp version is used instead, up to the limit of camp buffs a
  character can have.

**A toggle decides whether the raid relies on camp buffs** (the user, 2026-10-29). It's a setting of a raid,
and a search run covers both settings: the best comps with camps on and with camps off. With
camps on, the sim gives the raid all 9, since there's no per-character cap and 9 is the whole list. Each
player gets a camp buff only when they lack its class buff:

- A class buff on that player means the camp buff adds nothing, and the class buff counts in full.
- This is decided per player. Blessings are given per class, so it can differ within a raid. Every other
  class buff here is raid-wide in Forever.
- Lodestone is melee attack power only.
- In the engine, each camp buff shares an exclusive-effect category with its class buff, given or cast, and
  the class buff always wins. A resolver rule alone can't do it, since Enhancement's own Strength of Earth
  comes from its cast totem.

**Two consequences:**

- **Paladins, shamans, druids, mages and priests are worth less.** A raid without paladins still gets 80% of
  Kings and two thirds of Might from the camp. A raid without shamans gets 81% of Strength of Earth. The raid
  sim counts only what each provider adds over the camp.
- **A Balance druid costs everyone the Camp Chair's physical crit.** The Camp Chair is one aura, excluded
  whole by Moonkin Aura, and Moonkin Aura is raid-wide (client effect 65). So with a Balance druid, every
  member gains 3% spell crit and loses the chair's 2% melee and ranged crit. Leader of the Pack isn't
  excluded by the chair and stacks with it. Unverified in game.

The launch checks go into notes.md (Need to Verify, At 60):

- sit at a camp with and without Kings
- cast Kings after the camp buffs land
- do each of these with the group versions too (Greater Blessing, Arcane Brilliance, Prayer of Spirit, Gift
  of the Wild, Prayer of Fortitude)
- the Camp Chair with a Balance druid in the raid
- sitting at two camps in turn
- entering a raid instance with the camp buffs, and dying inside

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
out. **Not built yet as planned below:** the resolver gives each duty no placeholder covers to the first DPS
member with that duty's build, in a fixed order (Sunder, Battle Shout, Faerie Fire, Curse of the Elements,
the Crusader), blessings by what each class's members want, a placeholder's air totem by the raid's melee
share, and Power Infusion and Innervate to the first casters and mana users. A placeholder's own effects
(its Faerie Fire, Battle Shout, Sunder) can't be switched off. The search's fill is where assignments are
chosen by raid DPS. The plan: jobs go one at a time, the biggest gain first. A job goes to a placeholder first, since that costs no
DPS. When no placeholder can, it goes to the eligible DPS member whose raid DPS drops least, by their
single-player sims under the set at the raid's length, then to the lower slot. But a DPS member takes it
only when the job's gain to the raid is more than that cost. Otherwise it stays uncovered, and the result
says so. One job at a time can fall short of the best assignment; the search assigns exactly within its
sets (see [The search](#the-search)).

**An assignment's cost** is the holder's base DPS minus their DPS on duty, under the same set at the raid's
length. Duty runs the build's duty request, with the effect switched off in the holder's own settings, so
it casts the effect itself and pays for it. The report's `builds.json` has the numbers at 180 seconds:

- Faerie Fire costs Balance 31.8 (33.5 at 120 s, 29.1 at 300 s).
- Curse of the Elements costs Destruction 8.9.
- Battle Shout costs Fury 1.8.

**A DPS member can only hold an upkeep assignment its build has a duty variant for.** In `meta.json`, a
duty variant has `duty` (one of "Battle Shout", "Sunder Armor", "Faerie Fire", "Curse of the Elements",
"Curse of Recklessness", "Judgement of the Crusader", "Judgement of Wisdom") and `baseBuild`, the base
build's id. There are 25:

- Battle Shout: the 3 warrior builds
- Faerie Fire: Balance, Balance Wrath and Feral
- Curse of the Elements: the 4 warlock builds
- Judgement of the Crusader: `paladin__level_60_sub_holy` and `paladin__level_60_shock` (and 2 more below)
- the 13 M1b added, below

**The duty builds M1b added** (2026-10-29; DPS at 20,000 iterations, 180 ± 20 s, against the base):

| Duty | Builds | Cost |
|---|---|---|
| Sunder upkeep (5 stacks from the pull, refreshed in their last 3 s) | `warrior__arms-duty-sunder`, `warrior__fury-duty-sunder`, `warrior__fury_2h-duty-sunder` | 42.9, 34.1, 35.8 |
| Curse of Recklessness (in place of the Curse of the Elements) | `warlock__<build>-duty-recklessness`, all 4 | 14.2 to 15.3 |
| Judgement of the Crusader | `paladin__level_60-duty-crusader`, `paladin__level_60_shock_ret-duty-crusader` | 5.9, 0.7 |
| Judgement of Wisdom (new Seal of Wisdom, `sim/paladin/sow.go`) | `paladin__<build>-duty-wisdom`, all 4 base Ret builds | 3.6 to 14.7 |

- **Seal of Wisdom** has Forever's ranks and mana. Its own 90-mana proc chance is an assumption, 50%, like
  Judgement of Wisdom's server rule (the client only says "a chance"). A paladin gets the seal only if its
  rotation names it, so no other paladin's result changes.
- **A hunter can skip Hunter's Mark** (`Hunter.Options.skip_hunters_mark`); the base builds keep marking.
- **Assignable shaman totems:** a per-player `Player.assigned_totems` and an APL action, Cast Assigned
  Totem, which drops the assigned totem for an element. The report's shaman rotations don't use it yet,
  so the search still takes their fixed totems (Enhancement: Strength of Earth, Windfury, Mana Spring and
  Magma; Elemental: Mana Spring and Searing).

**Still missing:** duty variants of the encounter variants (a Battle Shout Arms warrior against
Dragonkin). Until they exist, a member on duty runs the base build's duty request.

## Groups

On the page, the user places groups by dragging members. A new raid starts from the rule below.

**The search places groups by the same rule** (the user, 2026-10-29: "whatever maximizes DPS"). Only Mana
Spring, Mana Tide and Dark Inspiration depend on groups, and each member's gain from them is its own.

1. **A package** is whatever party effects one party has. A Restoration shaman gives Mana Spring ×1.25 plus
   Mana Tide, a "+ Mana Tide" shaman gives Mana Spring plus Mana Tide, an Enhancement, Elemental or Warden
   shaman gives plain Mana Spring, and two shamans in one party give their stacking rule's result. Each
   build is simmed once per distinct package in the set, and once without.
2. **The package givers go one per party**, the richest package first. Two in one party isn't tried: under
   the sim's current rule two Mana Springs don't stack, so a second one in a party adds nothing (the rule is
   open, shaman_audit 8.45). A DPS shaman casts its own Mana Spring, so it's never given a package.
3. **The other placeholders go to parties without a package first**, so the packaged seats go to members
   who gain from them.
4. **Each packaged seat takes the member who gains most from its package.** The provider members compete
   for those seats as mana users, and the fillers take the best build under the package, repeats allowed.
   So choosing the builds and placing them is one step.
5. The rest fill on their DPS without a package.

With one package giver per party, this is exact, since each member's gain is its own and the seats of a
party are interchangeable. Dark Inspiration has a given form since M1b but isn't a package yet.

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
  - **The bosses' health**, set on the encounter, since no preset has Forever's numbers yet. The fight
    ends when every boss is dead, each one dying on its own (`sim/core/target_death.go`). A health fight
    has no duration variation. Before B the variation still moved the end of fight every rotation sees, and
    with ±20 seconds a 10-member raid lost 0.8%.
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

For a generic boss that's 43 builds: the report's 77 requests are 37 base builds, 25 duty variants, 9
encounter variants and 6 hunter pet variants (3 builds × Tallstrider and wolf), and the pet variants are
builds of their own.

- **The duty variants** are the base build holding an assignment, not entries of their own (`duty` and
  `baseBuild` in the variant's `meta.json`).
- **The encounter variants** have `baseBuild` and `encounter` (Humanoid, Undead, Dragonkin) instead, and
  join only when the boss's creature type matches.
- **The "+ Mana Tide", Water Shield and Imp GW builds** are different talent builds, so they are entries of
  their own. The "+ Mana Tide" builds also give their party Mana Tide (see [Groups](#groups)).
- **Each faction** needs each build's request at its best race in that faction, and 31 of the 58 requests
  have none for the other faction. The report makes race variants with `tools/dps_report/race_override.py`,
  which swaps in a race's own rotation and bonus stats from `req/races/`, with its own rules for the duty
  and encounter variants. The search's `Pool.Player` applies the same `req/races/` overrides the same
  way; a race with no file sims the preset's own rotation and gear.
- **New requests** (the pet variants, the missing duty builds) go into `tools/dps_report/req/`, and the
  report publishes them like any other build (the user, 2026-10-29).

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

**A run's inputs:**

- the raid size
- the tank and healer slot counts, and each slot's fixed preset or "any" (the default)
- optionally, fixed DPS members (a guild's roster), with the rest filled by the search. A fixed member
  may hold a duty in its own slot (a fixed Fury warrior on Sunder upkeep), and the fill tries each way.
- the encounter (boss, creature type, length)

**A run covers both factions, with camps on and off: four results** (see [Camp buffs](#camp-buffs)). Each
is its own search, and they share no sims, since the race and the camps are in every request.

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
| Earth, air, water and fire totems | a shaman placeholder's air totem (Grace of Air or Windfury); its earth, water and fire are fixed (Strength of Earth, Mana Spring, Flametongue cost a placeholder nothing); Enhancement and Elemental drop their builds' totems |
| Packages (Mana Spring, Mana Tide) | from the shamans present; which party gets each (see [Groups](#groups)) |
| Moonkin Aura | a Balance druid; none |
| Leader of the Pack | a Feral druid (cat or tank); none |
| Trueshot Aura | a Marksmanship hunter; none |
| Arcane Brilliance | a mage; none |
| Gift of the Wild | a Feral tank; a Restoration druid; any DPS druid; none |
| Prayer of Spirit | a priest (placeholder or Shadow); none |
| Power Infusion, Innervate | from the placeholders that have them; only "who" is chosen |

**Hunter's Mark isn't a role.** Only hunters gain from it (ranged attack power), each hunter marks for
itself and the strongest counts, so it's the same for everyone whether another hunter is there or not.
The shared mark (Survival gains 2 to 4 from a stronger one) is left to the raid sims.

Roles are coupled through their providers:

- A shaman covers all four elements.
- A warlock's one curse is either Curse of the Elements or Curse of Recklessness.
- A Balance druid on Faerie Fire covers Moonkin Aura, Gift of the Wild and the minor armor slot.

So a search move adds, removes or swaps one provider member, or changes one placeholder's choice, not one
role.

### Tank and healer combos

The search covers every combination of tank and healer presets that fits the slots (the user, 2026-10-29).
The presets (see [Tank and healer presets](#tank-and-healer-presets)):

- **Tanks:** Protection warrior, Feral tank, Protection paladin, Warden shaman.
- **Healers:** Holy paladin, Restoration shaman, Discipline priest, Holy priest, Restoration druid.

Combinations are counted as multisets, since two warrior tanks in either order are one combo. A run takes
one slot count, and combos with different counts are never ranked together:

| Raid | Tanks | Healers | Combos | Distinct signatures |
|---|---|---|---|---|
| 10 | 1 | 2 | 4 × 15 = 60 | 35 |
| 10 | 2 | 2 | 10 × 15 = 150 | 65 |
| 20 | 2 | 4 | 10 × 70 = 700 | 112 |
| 20 | 2 | 5 | 10 × 126 = 1,260 | 121 |
| 40 | 3 | 7 | 20 × 330 = 6,600 | 117 |
| 40 | 3 | 8 | 20 × 495 = 9,900 | 117 |
| 40 | 3 | 9 | 20 × 715 = 14,300 | 117 |

**A combo's signature** is what its placeholders can give:

- whether there's a warrior tank
- whether there's a Feral tank
- whether a druid gives Faerie Fire and Gift of the Wild
- the number of paladins, capped at 3
- the number of shamans (Restoration or Warden), capped at 2, for the totems
- whether there's a priest

The buff sets are ranked and simmed once per signature. Every combo with that signature then re-fills the
signature's best sets with its own counts: Power Infusion, Innervate, and the packages (each shaman's kind,
and the shamans beyond two). That's arithmetic on the cached sims.

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
  since almost every signature there covers every buff (three tanks without a warrior give no Sunder
  upkeep). Those cooldowns fire on fixed rules (see
  [Assignments](#assignments)).
- **Their own damage isn't in it** while tanks and healers aren't simmed. When they are, it adds to the
  same table.

### The search

For each signature:

1. **Prune the fillers.** Under any set, a build's DPS is its DPS under the full set less what each missing
   role costs it. So a dominates b when a's DPS under the full set, minus b's, is at least the sum over the
   roles of how much more a loses than b when that role is missing (only the roles where a loses more
   count). The roles include the packages, Power Infusion and Innervate.
   - a must also be able to fill wherever b can. A build that gives something can't fill a set without it
     (step 4), so a Balance druid never dominates a build that fills a camps-on set without Moonkin Aura.
   - Only the builds no other build dominates fill slots, at a margin of 1.5 DPS, since mana effects
     aren't additive.
   - It's recomputed for each scenario (faction, camps, length, boss). M0's count on the report's
     ablations kept 6 of the 37 builds (`docs/raid_sim_m0.md`). The search's own role sims keep more,
     since they price every role at its real cost: 9 of the 37 base builds for 10 and 20 players, Horde,
     camps off, before the pet variants joined (11 of 43 with them) (Beast
     Mastery, Lone Wolf, Marksmanship, fire, frostfire, sub Holy Ret, Shadow, Demonology, fire
     Destruction).
   - The best shaman of each package kind fills too, for its party's mana.
   - A provider's duty variants are providers only, never fillers.
2. **Rank every buff set with a model.** The model is each build's DPS under the full set, less what each
   missing role costs it. Those costs come from the search's own single-player sims: one per build with
   each role removed from the full set, at the run's length, race and camp setting (29 sims for each of
   68 builds, the 43 and the 25 duty variants, plus 64 mana mixes for each of 51 mana users: about 5,200;
   builds). The report's one-buff-at-a-time runs can't serve. They run 300 iterations at 180 seconds with
   no camps, and they skip the buffs a request never had (no Moonkin Aura on Feral, no totems on the
   warlocks). On top of the costs:
   - **The roles multiply:** each missing role leaves its fraction of the full-set DPS.
   - **Three pairs are simmed together**, since they don't multiply: Battle Shout with Furious Howl (one
     slot), Sunder upkeep with the minor armor slot
     (armor: without a warrior tank, Faerie Fire is worth 36 to Beast Mastery, against 64.4 with Sunder
     up), and Grace of Air with Windfury Totem.
   - **The mana sources are simmed as a mix.** A caster loses little from any one while the others
     suffice, and runs dry without them all, so multiplying them missed by up to 120 DPS. Each mana user
     is simmed under all 64 mixes of Judgement of Wisdom, Arcane Brilliance, Prayer of Spirit, Blessing of
     Wisdom and the four packages, and the model reads its mix. That brought the worst miss on a sparse
     set to about 45.

   Every combination of placeholder choices and provider members is scored this way, with no sims. The
   second audit counted 0.75 to 1.1 million set fills (10 to 40 players), not the 50,000 first guessed:
   90 s to 20 minutes of one core per signature group, the same warm or cold (see [Cost](#cost)).
3. **Sim the best sets exactly, until the model can't change the answer.** Each build's DPS under a set
   comes from a single-player sim, with the set written into its request.
   - Every signature's 4 best sets by the model are filled exactly.
   - The 3 signatures with the best exact sets go on in the model's order while a set's model score is
     within the model's worst miss on that signature's simmed sets of its best exact score, up to 200
     sets.
   - The model's misses are reported (its worst and mean miss over every set simmed).
   - An exact fill runs the fill with the model first, sims what it asked for in one batch, and fills
     again, until it asks for nothing new. It sims only the 3 providers the model makes cheapest for each
     role, and a best filler only for the packages the raid's parties have.
4. **Fill a buff set exactly:**
   - The provider members are in. Each one runs its duty variant for what it holds, simmed with that effect
     off in its own settings, so it pays the assignment's cost. In the search, only provider members the
     search chose hold duties. Fillers hold none.
   - Per-member choices are separable, so they're solved exactly inside the fill:
     - each class's blessings, picked by what they add summed over that class's members, providers
       included
     - the packages (see [Groups](#groups))
     - Power Infusion and Innervate targets
   - The remaining DPS slots take the best builds under the set, repeats allowed.
   - A build that gives anything (the "Given by" column: a mage's Arcane Brilliance, a priest's Prayer of
     Spirit, a druid's Gift of the Wild, a Balance druid's Moonkin Aura, a shaman's totems) may only fill a
     slot when the set already has what it gives. So the comp stays the set it was scored under.
   - The provider members are the cheapest cover of the roles the placeholders don't give, found exactly
     over the 4 strongest givers of each role by full-set DPS (3 per role, the model's cheapest, in an exact
     fill), a small search over the roles left to cover. Shaman
     fillers are tried at every count from 0 to the number of parties, each bringing its party's
     package. A class's blessings are solved by turns, each class taking the ones its members gain most
     from, until none changes.
   - The Sunder ramp's stack count follows the number of DPS warriors. So the fill is scored twice: under
     zero warriors first, and under the count that fill came out with, each re-filled under its own count.
     The better one stands.
   - The score is the sum of the members' DPS. Its standard error comes from their sims, and k copies of
     one build share one sim, so they carry k times its error, not √k times.
5. **Polish with local moves** from the best exact sets of the 3 leading signatures: each move toggles one
   role the DPS members bring, one paladin more or fewer, or one placeholder's choice (the air totem, the
   judgement), and the fill picks the members again, so one move can turn Balance on Faerie Fire into a
   Feral cat on Faerie Fire. A move is kept only when it raises the score by more than 2 standard errors of the difference,
   until none does.
6. **Confirm with raid sims** (see [Confirmation](#confirmation)).

**Iterations.** The single-player sims screen at 1,000 iterations with `SimOptions.use_labeled_rands`,
which pairs a build's sims under two sets more closely (0.55 against 0.22 in the Appendix's raid
measurement). The builds of the sets still in contention after step 5 are re-simmed at 10,000. Whether the
labeled streams still pair the fight durations hasn't been checked.

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

- Fill variants under one set differ mostly by their members' single-player sims, which check 1 covers.
  Not exactly: Survival and Enhancement gain up to 4 DPS next to many physical members, and the Hunter's
  Mark is shared, which across 13 copies is about 0.3%. The plan was to raid-sim a set's top 2 or 3 fills
  when they're within 0.3% of each other; that isn't built yet, so each set's best fill is the one
  confirmed.
- Comps that resolve to the same request are one: the two paladin presets, and party permutations.
- Every set within 0.1% plus 2.5 standard errors of the best score is raid-simmed, paired, at 5,000
  iterations. The top 2 go on to 20,000 (standard 15), on a fresh seed range, since seed s at 20,000
  iterations repeats the 5,000 run's iterations.
- The raid sims rank the confirmed sets, and their numbers are the answer. If a raid sim disagrees with its
  score by more than noise, the result notes the gap. (Widening the confirmation to the next sets of a type
  whose gaps lean one way isn't built; the notes show it.)

The raid sim catches what the single-player sims can't:

- a member's first curse or Faerie Fire landing after the pull
- the shared Hunter's Mark
- Mana Tide's timing
- the engine's interactions

**The tie rule:**

- Comps whose paired gap is under the larger of 2 standard errors and 0.1% are one tie set, and they're
  published as a tie. Before the confirmation (M1), the same rule runs on the scores and the standard error
  of their difference.
- The comp shown is the one with the fewest distinct builds, then the fewest DPS members on duty.
- Fillers within 2 standard errors of the best filler (one build against another) are listed as
  interchangeable. Fire and
  frostfire are 0.6 DPS apart, with standard errors of 1.86 and 2.57 at 1,000 iterations, so picking
  between them would be a coin flip.

### The cache

**A spec simmed under a buff set it was already simmed under is never simmed again** (the user,
2026-10-29).

- **The key** is the SHA-256 of the deterministic proto marshal of the resolved, normalized single-player
  `RaidSimRequest`, plus the sim's version:
  - The request covers the build's whole setup (gear, talents, base or duty rotation, consumes, race), its
    raid buffs, individual and party buffs, debuffs, the encounter, the duration (on the length grid) and
    variation, the iterations, the seed and the random stream option.
  - **Normalized** means the effects, not the fields that give them: the minor armor slot is "505 armor"
    whether Faerie Fire, Curse of Recklessness or Dust Cloud fills it. The resolver settles every exclusion
    first (camp buffs against class buffs, the weapon totem per receiver), so the key holds what the player
    gets.
  - **Effects a build can't read are dropped**, by a static table, `sim/compsearch/drops.json`, one for
    each build and camp setting. `compsearch -drops` writes it: an effect goes in only when the build's
    result is the same with it set and cleared, to within a hundred-thousandth, under the full set and the
    empty set, at 120, 180, 300
    and 420 s, in both factions and at two seeds, for each effect alone and then all together. Arcane
    Brilliance is dropped for a warrior, so it's one key whether a mage is there or not. With camps on,
    Moonkin Aura stays in a warrior's key, since it takes the Camp Chair.
  - The weapon totem is two entries, Windfury and Flametongue, so Enhancement (whose Windfury Weapon makes
    Windfury Totem useless to it) keeps its Flametongue.
  - A test checks the whole table again at another seed. Not bit for bit: Mana Tide does nothing for an
    Arms warrior, but with the other mana fields it moves the order of its float sums, 904.4142 to 904.4167
    at that seed, so the rule allows a hundred-thousandth.
  - The report's ablations don't decide this. Their `same: true` means only equal DPS at 300 iterations
    under the full set: Faerie Fire and Curse of Recklessness both show it for every build, because each
    fills the slot without the other.
  - **The version** is the SHA-256 of the running binary, which embeds the item database. A git hash would
    miss an uncommitted edit, and the downloadable server has no checkout.
- **The value** is the DPS, its standard error and standard deviation, and the mana stats.
- **Where it lives:**
  - In memory during a search.
  - On disk in `forever-sim/compsearch/` under `os.UserCacheDir()` (`~/.cache/forever-sim/compsearch/` on
    Linux), a folder per binary. The command line, the native server and the tests are different binaries,
    so they share the place but not the entries.
  - One small file per sim, written to a temp file and renamed, so a reader never sees half of one. That's
    about 4 KB of disk for 125 bytes; a packed store would fix it if the cache grows large.
- **What gets reused:** two Fury warriors in one comp, or one comp reached two ways, are one sim. Other raid
  sizes, other slot counts, other signatures and later searches reuse what's there, since a sim is keyed by
  the buff set a build sees, not by who gives it.
- **A code or database change starts a fresh cache.** Another binary's cache is pruned once it has gone
  unused for a week (a run marks its own as used, hourly while it stores), so the CLI and the native server
  built from different code don't wipe each other's. The binary is hashed at startup through
  `/proc/self/exe`, so one rebuilt in place under a running server doesn't lend its hash to old code.
- **In a Claude Code cloud environment** the cache starts empty each session, and that's fine. The cache
  isn't committed.

### Cost

The exact sims are what cost. They're keyed by the buff set a build sees, so the signatures share them. But
nothing is shared between the factions, the camp settings, the lengths or the bosses. One run of one
faction and camp setting, roughly:

- the role sims for the model: about 5,200 a scenario (see [The search](#the-search), step 2)
- a few hundred distinct sets × the surviving builds × each package, plus the providers' duty variants, the
  blessing alternatives, and the Power Infusion and Innervate targets
- the local moves, the role table's removals and the member values' neighbours
- two grid points per length, in a health fight

That was guessed at about 24,000 sims. The second audit measured one scenario at small iterations and
scaled it to the defaults:

| Size | Screening and contender sims | Ranking and combos (arithmetic) | Core-hours before the confirmation |
|---|---|---|---|
| 10 | about 9,800 | about 50 core-seconds | about 2.2 |
| 20 | about 12,800 | about 420 core-seconds | about 2.9 |
| 40 | about 10,800 | about 1,560 core-seconds | about 2.4 |

- About half the guessed sims, but at about 0.6 ms a single-player iteration (a Beast Mastery hunter
  1.2 ms, a mage 0.22), not 0.37.
- The contenders' re-sims at 10,000 iterations are 20 to 35% of the screening's time.
- Warm, the sims are free but the arithmetic isn't: a 40-player scenario still takes about 25 core-minutes,
  most of it the ranking (one goroutine per signature) and the combos' re-fills. The model is indexed by
  arrays and the fill's sort is typed since; an upper bound to prune plans, and a default combo limit for
  40 players, would cut it further.
- A search keeps every sim it has run for its scenario. Each kept its 37 KB request until the second audit
  (a 10-player search peaked at 744 MB); now only the value and the key stay.

M1's write-up is 4 of those (2 factions, camps on and off) under about 8 scenarios: about 80 core-hours, or
5 hours on 16 cores.

The confirmation, with today's engine:

- **20 members:** about 6 seconds a set at 5,000 iterations on 16 cores.
- **40 members:** 8 sets at 5,000 iterations and 2 at 20,000, at 63.5 ms an iteration, is about 5,100
  core-seconds: 5 minutes on 16 cores per scenario.

M1 measures all of this.

### Noise

- At 5,000 iterations a 10-member raid's standard error is 0.021%. With today's correlation of 0.22 between
  two raids on the same seed, their difference has a standard error of about 0.026%, so a 0.1% gap is about
  4σ. For 20 members it's about 5σ.
- **The screening is noisier than the confirmation.** A fill is k copies of one cached sim, so its error
  grows k times. At 1,000 iterations with k of 12 to 25, two sets' score difference has a standard error of
  about 0.12 to 0.15% of raid DPS, the size of the tie window. That's why local moves need 2 standard
  errors, the contenders are re-simmed at 10,000, and the confirmation takes every set near the top rather
  than a fixed number.
- The other limit is how well the single-player sims add up, about 0.1%. That is the same size as the gaps
  between the top comps: swapping one build in a 10-member raid is worth 0.02% to 0.16%. That's why there's
  a tie rule.
- Seeds s and s + 1 are the same sample shifted by one iteration (iteration i uses seed + i). Replicates use
  seeds further apart than the iteration count.
- Per-member random streams would make whole-raid comparisons sharper. The search doesn't need them, so
  they come later.

### Where it runs

On the user's machine or in Claude Code cloud environments, natively; nothing is hosted (the user,
2026-10-29).

- **The search is a Go package,** `sim/compsearch`, so the native server imports the same code.
- **The command line** (`tools/compsearch`, a thin wrapper; later perhaps a `wowsimcli` command) runs the
  search on goroutines across every core. It writes the result as JSON (the `CompSearchResult` proto) and
  markdown.
- **The native server** (`sim/web/main.go`, run by the downloadable binary or `make host`) gets the same
  search as an async endpoint, `/compSearchAsync`. It streams progress with the best comp so far, through a
  new comp search result in `ProgressMetrics` (`api.proto`). Progress goes out at least every few seconds,
  since the server drops an async call after 10 minutes without any. Abort stops it. The pages already send
  their sims to the native server when it serves them (`ui/core/sim.ts`, net workers).
- **A page served as static files** can still show results and sim a single raid in its WebAssembly
  workers, but searching needs the native server.

### The result

One `CompSearchResult` (defined in M1, read by the M2a viewer) holds everything below, and the page shows
the same items. It leads with what could change the answer:

1. **The tie set**, and every sensitivity scenario that changes the winner (see
   [Sensitivity](#sensitivity)).
2. **For each faction and camp setting:** the best comp and 2 to 4 runners-up, with the difference and
   paired 95% intervals (from the scores, before the confirmation). Each comp lists:
   - every slot's build or preset, its party and its duties, with each duty's cost
   - the placeholders' choices
   - raid DPS, each party's DPS and each member's DPS
3. **The tank and healer combos ranked**, and each preset's DPS value (see
   [Tank and healer combos](#tank-and-healer-combos)).
4. **The role table:** each role's gain, its provider's cost and the net.
   - Each role a DPS member brings is measured at the winning set, with that role alone removed, then
     re-filled and re-resolved. A role the winning set lacks gets its value of adding instead (see
     [Member value](#member-value)).
   - A role whose givers all bring another role too (a Balance druid's Moonkin Aura with its Gift of the
     Wild) can't be removed or added alone, so it goes or comes with the giver's other roles, and the row
     says which; the Crusader comes with one more paladin. A provider holding two roles is charged its
     cost under each, and the row says so.
   - A shaman placeholder's air totem is valued against the other one, re-filled. Other roles the
     placeholders give are listed without a value: they cost nothing and can't be removed.
   - Not built yet: pairs that interact shown together (Grace of Air with Trueshot Aura or Moonkin Aura,
     the mana sources). Trueshot Aura, for example, is worth 10.8 a hunter with Grace of Air up, and under
     Windfury alone no hunter is in the fill, so it's worth 0.
5. **Each buff and debuff** with its provider, scope and receivers, from the resolver's record, and the
   members who miss a buff the raid could have given them. Uptimes come from the raid sims, so only
   confirmed comps have them.
6. **The filler ranking** under the winning buff set, each build with the spread of the report's
   assumptions (`assumption_dps.json`).
7. **Each member's contribution, replacement margin and difference from the DPS Report** (see
   [Member value](#member-value)).
8. **Each member's threat per second.**
9. **The assumptions it ran under** (the unverified stacking rules, the Camp Chair's exclusion), the role
   options the run left out (M1's list), and what isn't modelled: healing, tanking, threat, pet deaths,
   movement, boss mechanics.

## Sensitivity

The top comps are close. A rough estimate from the report's one-buff-at-a-time runs puts a hunter-heavy raid
and a caster-heavy raid a few percent apart, about 43 DPS a slot. That is inside the spread of the report's
own open assumptions, up to ±60 to 120 DPS a build.

So the result says whether the winner holds under each of these:

| Scenario | Why | Cost |
|---|---|---|
| Each build shifted by its assumption combos (`assumption_dps.json`) | the report's open rules move builds against each other | no sims |
| 120 and 300 seconds | Fury, rogues and the hunters fall faster than casters with length; hunters run out of mana by 300 | a re-run at each length |
| Each raid's boss: generic in the flask zone; Dragonkin outside it | off the zone the flasks give only Stamina, which costs builds unevenly: fire about 112, Fury 64, Beast Mastery 38 | a re-run |
| Raid damage at 0 and 10 hits a minute | Demonology loses 63 DPS at 10 a minute; Fury 2H gains 56 | from the assumption data |
| Light consumes (no flask) | the flask is worth 112 to a fire mage and 38 to a Beast Mastery hunter | from the one-buff-at-a-time runs |
| Pet uptime about 80%, and Lone Wolf | the pet is 332.7 of a Beast Mastery hunter's DPS and never dies in the sim. At 80% Beast Mastery falls to 1065.3, below Lone Wolf (1094.3) and fire (1089.0), and "29 Beast Mastery" becomes "29 Lone Wolf" | a re-run of the hunter builds |
| Each unverified stacking rule on and off | Mana Spring from two shamans, Furious Howl against Battle Shout and on pets, Hunter's Mark, the Crusader | a re-run |
| The Camp Chair excluded only in its spell crit part | how the Chair's exclusion works is unverified | a re-run with camps on |
| Camp buffs stacking with the group versions of their class buffs | the tooltips name only the single-target spells | a re-run with camps on |
| Camps on with one Journeyman camp: the 5 features worth most to the winning comp | all 9 needs three Basic camps (the unverified second-camp rule) or an Expert camp, whose Blueprint has no known source | a re-run with camps on |

When the winner changes under a scenario, the result says so and shows the comp that wins.

**Built** (2026-10-30): the assumption combos (every search, no sims), and `-matrix` (the command line;
`RunMatrix`), which re-runs the search under 120 and 300 s, the other boss, raid damage at 0 and 10 a
minute, no flask, and, with camps on, the Camp Chair keeping its physical crit and camp buffs stacking
with the group versions. Not built: the pet uptime, the unverified stacking rules and the Journeyman camp
(the engine can't express them yet; `RaidAssumptions.journeyman_camp` is read nowhere).

## Member value

All of these come from the single-player sims under buff sets. So they need no raid sims once the cache has
them.

- **Contribution of a DPS member:** raid DPS minus raid DPS without the member. Without them, the comp is
  resolved again: the parties' packages are recomputed from the givers left, and a duty the member held
  goes with it unless another member gives that effect (another member doesn't take it over). It's shown as the member's own DPS plus what they add to everyone else. This is "what a spec
  brings".
- **Contribution of a tank or healer:** the slot count is fixed, so taking one out leaves a hole. It's the
  combo's raid DPS minus the best raid DPS with that slot's preset changed.
- **Identical members share one value.** Two copies of one build get the same contribution, whichever one
  the search calls the provider.
- **Replacement margin:** raid DPS minus the best raid DPS with another build in that slot. The neighbours
  are evaluated on demand from the cache, not taken from whichever sets the search happened to visit. At
  the best comp it should be zero or more, within noise. A margin below that means the search missed a
  better comp: the result's notes say so (the search doesn't yet go on from it). A margin within noise
  means two builds
  are interchangeable.
- **Value of adding** a spec that isn't in the comp: the best raid DPS from swapping it in for any member,
  minus raid DPS.
- **Assignment cost:** the holder's base DPS minus their DPS with the assignment, under the same set.
- **Difference from the DPS Report:** the member's raid DPS minus their report build's published DPS,
  broken down by single-player sims, one change at a time: the report's request in this sim, the race, the
  buff set, then the duty, pet or encounter variant (the length is in the buff set step). A
  member whose setup isn't a report build's shows none.

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
- whether the raid relies on camp buffs
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

**Results on the page:** every item of [The result](#the-result) for a search. For a single raid run:

- raid DPS with a 95% interval
- the party and member tables, with each member's difference from the DPS Report
- each buff and debuff with its provider, scope and uptime, from the resolver's record and the sim
- the members who miss a buff the raid could have given them
- assignment cost and member value

If a change touches shared result views (the events view, the replay), the DPS Report is deployed after it
(CLAUDE.md).

## Checking it

1. **The report raid.** A raid of base report builds, with no duty or encounter variants:
   - Each member's set (the per-player field) is their own request's raid buffs, less what the raid's
     members cast (a cast Windfury Totem replaces each request's given one), plus their own held auras,
     which the resolver gives. A Balance druid's request leaves out its own Moonkin Aura, since the class
     hook gave it.
   - The debuffs are the base requests' common ones (every base request carries the same ones).
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
   as trying every one, scored by exact cached single-player sims, not only by the model. The pool has a
   pair that doesn't add: physical builds with and without a warrior tank and the minor armor slot.
6. **After 2026-12-09**, the user's raid logs (`tools/warcraftlogs`, `logcheck`) and public logs are checked
   against the sim: raid DPS by comp and buff uptimes.
7. **The resolver**, as unit tests with no sims:
   - a Greater Blessing lands only on its class, and on its pets or not as the table says
   - one blessing per paladin for each player
   - the minor armor slot, and the weapon totem per receiver
   - "nobody provides it" shows in the miss list
   - camp buffs per player, the Camp Chair against Moonkin Aura included
   - with placeholders leaving gaps in the parties, each result slot maps back to its comp slot
8. **Metrics.** Raid DPS equals the sum of the members' DPS, pets included.
9. **Determinism.** One search gives the same result across runs and core counts (its JSON, less the seconds
   and the timings by phase).
10. **The cache's dropped effects** (see [The cache](#the-cache)).

## Milestones

Forever launches on 2026-11-04, and the classes will change with it. The raids open on 2026-12-09.

**Implementation starts with the bug fixes** (the user, 2026-10-29). M1 doesn't depend on them, since its
single-player sims hit none of them, so this is an order, not a dependency.

**The plan:**

1. B, the bug fixes (done 2026-10-29).
2. M0, the role-value estimate, half a day.
3. M1 working by about 11-07, then run against the launch builds.
4. M1b by mid-November.
5. M2a by late November, then M2b and M2c.
6. A final run with the sensitivity matrix in the week of 12-01.
7. Check 6 against real logs after 12-09.

Each step says what it needs from this doc and when it's done.

### B: the bug fixes

Every bug in the Appendix's "What breaks" list, except a cast buff that must reach a new scope (Battle
Shout, the totems, Furious Howl), which is M1b. Blood Pact moved to the party here (fix 6).

**None of them may move a single-player number.** Every one of the 58 report requests gives bit-identical
results at a fixed seed before and after.

**Done** (2026-10-29), in one commit, from the scratch patch reviewed again before applying:

- The 12 new tests in `sim/raid_forever_test.go` pass. On the commit before they failed, and the empty-slot
  test crashed.
- The 58 report requests' result hashes (1,000 iterations, seed 1, first-iteration logs) equal the baseline
  taken before the fix.
- The whole `./sim/...` suite gives the same outcomes before and after. `TestP1Mage` and
  `TestElementalLevel60` fail on both, with stale `.results` files, and `sim/web` doesn't set up without
  `binary_dist`.
- The given Furious Howl was the rank 4 value at 60 (83). Since the third audit it's the player's level's
  rank (`core.ForeverFuriousHowlAt`), with a test that it matches the wolf's own ranks.

**The fixes** (planned as separate commits, landed as one):

0. **The test file and the baseline.** `sim/raid_forever_test.go`, with the helpers and
   `TestRaidForeverReportHashes`. That test is opt-in (`RAID_FOREVER_HASHES=file`, about 21 s): it runs
   the 58 requests at 1,000 iterations, seed 1, with first-iteration logs on, and hashes each whole
   canonicalized result. Metric lists come out in map order, so they're sorted before hashing. Take the
   baseline before the first fix.
1. **The aura cap and the presim list.** These unblock raids of 20 and 40.
   - The aura cap (`aura.go:461`) becomes 200 × the number of players. Each caster registers every rank's
     DoT on each target. Not registering ranks that can't be cast would save the most, but it changes
     registration order, so it belongs in Speed.
   - The presim list (`presim.go:38`) becomes a map. Any health fight with 26 or more players panicked.
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
   - Per-caster auras in one exclusive category would be cleaner, but they would move the report's 6
     paladin requests, which carry a given Crusader and cast their own.
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
     raids of 20 and 40, or the doubles. It did fire below level 56, which no report request is at: a
     warlock's own Curse of Recklessness and a rogue's Expose Armor registered at the top rank under the
     label the given one, at its own rank, used too. The cast curse now ranks by the warlock's level, a
     given debuff below the top rank has a label of its own (as Faerie Fire's does), and a test runs every
     report build at 30, 50 and 55 with the armor debuffs given.
   - The given `furious_howl` is read: a permanent Furious Howl at the player's level's rank (83 at 60).
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

It gives each role's value at 10, 20 and 40, the ranking of the fillers and the dominance count, before
any search code is written. It's rough: the runs are 300 iterations at 180 seconds with no camps, and they
skip the buffs a request never had. M1's own role sims replace it as the search's model.

- **Reads:** [Roles](#roles), [The search](#the-search).
- **Done when:** a table of role gain, provider cost and net, the filler ranking and the builds left after
  pruning (by the rule in [The search](#the-search)) are in TASKS.md or a note.

**Done** (2026-10-29): `tools/compsearch/m0.py` writes `docs/raid_sim_m0.md`. 6 of the 37 builds survive its
pruning.

### M1: the first answer (a command line search)

Single-player sims of the report's requests, with a buff set written into each one, already work today. So
M1 needs none of B's fixes. Its engine pieces:

- **The camp buffs:** 9 settings fields, each one's effect, and an exclusive-effect category shared with
  its class buff, the class buff winning (see [Camp buffs](#camp-buffs)). The `NextIndex` comments in
  `RaidBuffs` and `Player` are stale (41 is `salvation`, and `Player` runs to 52), so check the real next
  free number.
- **The Sunder ramp** on the given debuff (see [Sunder Armor](#sunder-armor)).

With one player per request, the resolver writes `Raid.buffs`, `Raid.debuffs`, the party's `buffs` and
the player's `buffs` directly. The per-player `raid_buffs` comes in M1b. The class hooks still run, which
is right for the player's own Moonkin Aura, Leader of the Pack and Trueshot Aura. Blood Pact is a party
hook since B.

**M1 searches only the role options that exist.** With M1b's builds in the pool, it leaves out:

- a DPS shaman's totem choice: Enhancement and Elemental drop their builds' fixed totems (the engine's
  assigned totems aren't in their rotations yet)
- Dark Inspiration as a party package

The search stops before the confirmation, which joins in M1b, so M1's tie set and intervals come from the
scores (see [Confirmation](#confirmation)).

**What M1 builds:**

- **The pool** from `tools/dps_report/req`, with the missing best-race requests for each faction made by
  `race_override.py` (see [The build pool](#the-build-pool)).
- **The preset table:** each tank and healer preset's spec, talent string and what it gives, from
  [Tank and healer presets](#tank-and-healer-presets), in `sim/compsearch/presets.go`.
- **The protos:** `CompSearchRequest` and `CompSearchResult` (see [The result](#the-result)), and the comp
  proto, at the end of `api.proto`. A file of their own can't be imported from `api.proto`, since they
  use its `Player` and `SimOptions` and `ProgressMetrics` carries the search's result; and the makefile
  generates TypeScript only for api, test and ui (`makefile:93-96`). The M2a viewer reads the result.
- **The provider table** as Go, with Judgement of Wisdom and one judgement per paladin.
- **A single-player resolver:** copy the request, then write in its buffs and debuffs per
  [Writing a role into a request](#the-provider-table), with the provider's own effect left off.
- **The search, the cache and the result**, on goroutines, in `sim/compsearch/`, with the command line in
  `tools/compsearch/`.
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
  - the dropped-effects test passes (see [The cache](#the-cache)), and the output is byte-identical
    across runs and core counts
  - Thorns, Retribution Aura, Salvation and Judgement of Light are measured once against every base
    build
  - the 20-player result for each faction, with camps on and off, is written up, with its tie set, role
    table, filler ranking, tank and healer combo ranking and sensitivity

**Done** (2026-10-30), apart from the write-up, which is in [M1's 20-player result](#m1s-20-player-result):

- `sim/compsearch` and `tools/compsearch` (`go run -tags with_db ./tools/compsearch -size 20 -out x` writes
  `x.json` and `x.md`; `-health`, `-boss Onyxia`, `-fixed-*`, `-confirm`, `-drops` and
  `-measure-left-out` are the other flags).
- Check 5 (`TestSearchMatchesBruteForce`, any tank, the pet variants and every placeholder choice), the
  cache (`TestCacheReuse`: a second run makes zero sims, the key changes with the encounter and the
  length; a code or database change is a new binary, so a new version), the resolver's given-or-cast tests,
  the drop table (`TestDropsAreUnread`) and determinism (`TestSearchDeterministic`) pass.
- `-measure-left-out`: every base build's result is bit-identical with and without Thorns, Retribution
  Aura, Salvation and Judgement of Light, so leaving them out moves nothing.

### M1b: whole raids that run right (engine)

This is what the confirmation raid sims need, on top of B.

**The new pieces:**

- `raid_buffs` per player, skipping the class hooks
- the given forms that are missing: `Debuffs.dust_cloud`, and Dark Inspiration reaching the priest's party
  when it procs, with a given form at the priest's measured uptime for the single-player sims
- a players-only raid aura array
- cast buffs reaching their scope:
  - Battle Shout to the raid. `newShoutSpellConfig` (`shouts.go`) needs to skip nil entries and count
    party threat per receiver, or it panics on a raid aura array.
  - the shaman's totems to the raid or party, with a count of standing totems per receiver
  - the weapon totem rules per receiver
  - Furious Howl to the raid, and to pets once the launch check says so
- Faerie Fire in Cat and Bear Form as the caster spell, free, with a 1.5 s GCD and no cooldown, replacing
  Faerie Fire (Feral) (the TASKS.md item). Then re-measure the Feral duty cost.
- Seal of Wisdom and its judgement in `sim/paladin`
- the missing builds in [Assignments](#assignments), and the per-player assigned-totems option
- the hunter pet variants, as requests in `tools/dps_report/req/`

**Then** the search's confirmation runs whole raids, through the comp proto (at the end of `api.proto`,
below) and one Go call, `RunComp`, which resolves a comp into a `RaidSimRequest` and runs it.

- **Reads:** [Buffs and debuffs come from the members](#buffs-and-debuffs-come-from-the-members),
  [Sunder Armor](#sunder-armor), [Assignments](#assignments), [Confirmation](#confirmation),
  [Checking it](#checking-it), [Appendix](#appendix-measurements-2026-10-29).
- **Done when:**
  - the report raid and Scope checks (1 and 4) pass in `sim/raid_forever_test.go`
  - the requests other than the Feral Faerie Fire duty build still give bit-identical results at a fixed
    seed (56 did; `druid__feral-cat` moved only in an unused aura's name, with every number the same); that
    one and the new requests are re-simmed into the report
  - checks 7 and 8 pass
  - a comp round-trips through JSON
  - `RunComp` on 7 builds plus 3 placeholders equals a hand-built request, with every DPS member taking
    the raid damage hits

**Done** (2026-10-30):

- **The engine:** `Player.raid_buffs` (53), `Player.assigned_totems` (54) with the Cast Assigned Totem APL
  action, `PartyBuffs.dark_inspiration` (its given form at 99.9% uptime, measured on both Shadow builds),
  `Debuffs.dust_cloud` (27), a players-only raid aura array, and Battle Shout, the totems, the weapon totem
  rules and Furious Howl reaching their scope (tests: `TestRaidForeverBuffScope`, `CastBattleShout`,
  `TotemHolds`, `WeaponTotems`, `OwnRaidBuffs`, `AssignedTotems`, `CastFuriousHowl`).
- **Choices made there:**
  - A player with `raid_buffs` neither gets the other players' hook buffs nor gives its own, so the
    resolver gives every held aura to everyone.
  - Every receiver of a raid-wide shout adds its threat, so a shout makes far more threat than the old
    party shout.
  - A receiver of both weapon totems takes the one it dropped itself, then another shaman's Windfury over
    Flametongue, then the last dropped (the resolver gives Enhancement Flametongue and the others
    Windfury anyway).
  - Identical totems from different shamans share one aura on each receiver, and each totem keeps it up
    until that totem ends, so the buff stays while a second totem stands.
- **The builds:** the duty builds and pet variants in [Assignments](#assignments) and
  [Hunter pets](#hunter-pets), Seal of Wisdom, and Faerie Fire in Cat and Bear Form. Their metas hold
  their own 20,000-iteration sims, and the DPS Report re-simmed and published them (as variants under its
  Duty chip; deployed 2026-10-30).
- **The hashes:** the 56 other requests are bit-identical. `druid__feral-cat`'s hash moved too, with its
  DPS and every number the same: it used to register a never-active Faerie Fire (Feral) aura on the
  target, which no longer exists.
- **The resolver and RunComp** in `sim/compsearch/resolve.go`, `resolveComp` in WebAssembly and
  `/resolveComp` on the native server. Automatic assignments give a job to a placeholder first, then to the
  first DPS member with that duty's build; the gain-over-cost check needs the member value sims, so a
  hand-built raid shows each assignment's cost instead.
- **Checks:** 1 (`TestReportRaid`: a 10-member, a 20-member (with an Enhancement shaman) and a 12-member physical raid within 4σ of each
  member's single-player sim), 4 (the engine tests above), 7 (`TestResolve`), 8 (`TestRaidMetricsAddUp`),
  the round trip and `TestRunCompMatchesHandBuilt` pass.
- **The confirmation** (`-confirm`) raid-sims every distinct set near the top, at least 2. On a 10-player
  search the raid sims landed 1.5 to 11 DPS from the scores (0.15% at most).

**The comp proto** (as built, at the end of `api.proto`):

```
RaidComp        { size; repeated CompSlot slots; Encounter encounter; repeated Assignment assignments;
                  Faction faction; bool camp_buffs; RaidAssumptions assumptions; SimOptions sim_options; }
CompSlot        { id (stable); party; index; RaidRole role;
                  oneof member { CompMemberBuild build; RaidPreset placeholder; }
                  report_id; name; bool preset_innervate; }   // the Feral tank's Innervate
CompMemberBuild { Player base; repeated CompDutyPlayer duties; }   // CompDutyPlayer { Duty duty; Player player; }
Assignment      { holder (slot id); set_by_hand;
                  oneof job { Duty duty; BlessingAssignment blessing (class, blessing); PaladinJudgement judgement;
                              TotemAssignment totem; CooldownTarget cooldown (Power Infusion or Innervate, target slot id); } }
RaidAssumptions { camp_chair_spell_crit_only; camp_buffs_stack_with_group_buffs; journeyman_camp; }
Duty            enum { BattleShout, SunderUpkeep, FaerieFire, CurseOfTheElements, CurseOfRecklessness,
                       JudgementOfTheCrusader, JudgementOfWisdom }
RaidCompResolution { RaidSimRequest request; repeated slot_of_player; repeated CompBuffRecord buffs;
                     repeated string misses; repeated Assignment assignments; ErrorOutcome error; }
```

- **A duty is a different `Player`, not a setting.** The duty variants differ from their base in the
  rotation, some Crusader paladins in their gear too, and the Feral one in its consumes. So a member
  carries its duty players, and the resolver picks one by its assignments. Moving a duty to another member
  picks that member's.
- **Slot ids are stable**, so dragging a member to another group keeps its assignments.
- **Every choice is an assignment**, a placeholder's included, so each one can be automatic or set by hand.
- **Go never looks up a reference.** The page fills the players, and the tests and the command line fill
  them from `req/`.

**The resolver**, in Go:

- **Input:** a `RaidComp`.
- **Output:**
  - the `RaidSimRequest`: each player's `raid_buffs` and `buffs`, the party buffs, the debuffs, and each
    member's base or duty rotation
  - a record of each effect: scope, provider slot, given or cast, receivers
  - the list of who misses which buff

The page reaches it through a new worker call, `resolveComp`: a WebAssembly function, and a `/resolveComp`
endpoint on the native server, with its `SimRequest` entry (`ui/worker/types.ts`, `worker_http.ts`).

### M2: the page

- **M2a, the results viewer.** Done when a search's JSON shows as a static page.
- **M2b, the comp editor with Run.** Done when a 20-player raid of report players runs in the WebAssembly
  workers and shows every item under results.
- **M2c, the search button through the native server.** Done when the search streams progress and stops on
  Abort, and the button is off when the page is served as static files.

**Reads:** [The page](#the-page).

**Done** (2026-10-29): `ui/raid_sim/` (`/forever/raid_sim/`), linked from a "Raid sim" card next to the DPS
Report, with four tabs: Search result, Search, Comp and Raid results.

- **M2a:** a CompSearchResult JSON from a file, a drop or `?result=<url>`, every item of
  [The result](#the-result), one tab per faction and camp setting. "Edit" on a comp loads it into the editor.
- **M2c:** the search form (size, slots, fixed presets and DPS members, the boss, factions, camps,
  iterations) streams the phase, the sims done and the best comp from `/compSearchAsync`, and Abort
  stops it. Served as static files, the page has WebAssembly workers and the button is off with a line
  saying why.
- **M2b:** members from `ui/raid_sim/data/builds.json` (written by `go run -tags with_db
  ./tools/raid_sim_data` through the search's own pool: each base build's Player per faction at its best
  race, its duty and encounter variants), placeholders, drag and drop (or pick and click on a phone),
  assignments automatic or by hand, the boss, faction and camps. Run resolves the comp (`resolveComp`)
  and runs it in the workers; the results are raid DPS, the parties, the members, each buff with its
  provider and uptime, the misses and the resolved assignments, with `EmbeddedDetailedResults` in Raid
  results. Member values (`/compValues`) gives each member's contribution, replacement margin and duty
  cost from the native server, which runs the single-player sims its cache lacks.
- **The native server is built with the item database** (`make host`, the release binaries), since the
  search and the resolver build each member's environment there.
- **Different from the plan:** the page is a `SimUI` (for `EmbeddedDetailedResults`, the sidebar and the
  tabs) without `addRaidSimAction`, and the resolver, not `Raid`/`Party`, builds the request. A placeholder
  can't hold a duty (the druids' Faerie Fire is automatic), and a DPS shaman has no totem choice, as in
  the resolver. Uptime matches effect names to the sim's aura names, so a given buff the sim tracks no
  aura for shows none.

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

## M1's 20-player result

The write-up of M1's done-when: 20 players, 2 tanks and 4 healers, the generic boss at 180 ± 20 s, each
faction with camps on and off. It's being run (a few hours on this machine) and lands here.

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
  - The DPS Report gets the new variants (the hunter pet variants, the missing duty builds) and publishes
    them. Nothing else in it changes for the raid sim.
  - Implementation starts with the bug fixes (B).
- **Camp buffs**
  - The sim models Forever's camp buffs. A camp buff doesn't stack with its class alternative.
  - When the raid lacks a class buff, its camp version is used instead, if a camp feature gives one, up to
    the limit of camp buffs a character can have. There's no per-character limit and 9 features in all, so
    the raid gets all 9 (see [Camp buffs](#camp-buffs)).
  - A toggle decides whether the raid relies on camp buffs, and the search finds the best comps with
    camps on and off.

None open. The launch checks are in notes.md (Need to Verify, At 60).

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
neither Blood Pact nor the double Atiesh moves a report number. 56 carry a given Judgement of the Crusader
(all but the 2 Crusader duty requests), and the 6 paladins among them also cast it.

**What works:**

- `RunRaidSim` builds and runs a raid in parties of 5.
- Metrics add up per member, per party and for the raid; pet damage goes to its owner.
- Each caster's DoTs are its own.
- Shared debuffs dedupe by label, as in game.
- Windfury Totem's cooldown is per receiver.
- Healing Stream already reaches the party.
- Members who don't interact add up (see [Why it doesn't try combinations](#why-it-doesnt-try-combinations)).

**What breaks** (before B; each item was fixed in B unless it says otherwise):

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
  - A warrior's Battle Shout reaches only their party (M1b). When two warriors shout, each receiver keeps
    the first one's value.
  - A shaman's Strength of Earth, Grace of Air, Windfury and Mana Spring reach only the shaman, and the
    weapon totem rules live on the shaman (M1b).
  - When one of two shamans' identical totems ends, the buff comes off everyone (M1b).
  - Furious Howl reaches only the hunter's party (M1b), and its given field was never read.
  - A given Mana Tide pays a party of 5 five times (`buffs.go:1472,1491`).
- **Health fights.** The duration variation still moves the end of fight that rotations see, and draws one
  extra random number an iteration: −0.8% for a 10-member raid, and −1.6% for Destruction.
- **One-player assumptions** (left as they are: these paths stay single-player).
  - Stat weights, the bulk sim, `sim/lib`, the test suite and the Vaelastrasz encounter all read
    `Parties[0].Players[0]`.
  - So do the external Improved Shadow Bolt and Stormstrike debuffs, which Forever skips anyway, and which
    the resolver never gives.
  - Many settings debuffs go on boss 0 only.

**Known and harmless:**

- Fixed since (fc03f362c): `spell_result.go` indexed attack tables by `target.Index`, the raid index for a
  friendly target, which panicked on self-damage in a raid with gaps. It reads them by unit index now, and
  no single-player number moved.
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
