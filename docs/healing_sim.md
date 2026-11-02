# Healing sim

Healers that heal: healer presets that cast real heals and spend real mana, damage on the tanks and the
raid for them to heal, and a score that ranks healer builds and tells the raid sim how many healers a comp
needs and which. The user, 2026-11-02: "We also want to simulate proper healing soon."

It meets the tank work ([tank_sim.md](tank_sim.md)) in the middle. Tanks need someone healing them for the
chance of death to mean anything, which today is the engine's stand-in heal stream. Healers need damage to
heal, which is the boss hitting the tanks plus the encounter's damage on the raid. Both use the same health
and damage groundwork, which this doc's first milestone builds.

**Where it stands (2026-11-02):** H1, the engine groundwork, is in. Heals are aimed at the hurt and their
overhealing is booked, a healing fight tracks every player's health and kills, the boss has raid damage
events and tank swaps, and the shaman heals others. `tools/healbench` sims a healer in a raid against made-up
profiles. No healer spec is registered yet: the bench's shaman is the Elemental agent with the Restoration
preset's talents and a heal rotation. Next is H2, the Restoration shaman.

## The user's rulings (2026-11-02)

- **What a healer is scored on:** effective healing per second, efficient use of mana, keeping the tanks and
  the raid alive, and as few "scares" as possible (a tank or member brought close to death who lives).
- **How heals pick targets:** assignments plus smart targets in the rotation. Each healer has an assignment
  (a tank, several tanks, or the raid), and the rotation can aim a heal at "the most hurt in my assignment"
  and similar. The raid sim fills assignments on its own, and the user can set any of them by hand.
- **Who takes damage:** the main tank, several tanks in some encounters, and per encounter other raid members
  (the user: "Healers need to simulate main tanks (or multiple tanks in certain encounters) and sometimes
  various other raid members taking damage (depending on encounter)").
- **The damage until Forever's bosses exist** (no boss data before 2026-12-09): configurable damage profiles,
  with some representative Classic and Season of Discovery bosses as stand-ins.
- **Order:** the shared groundwork now. Nobody else is building it (the tank session stopped after its tank
  sheet and bench, waiting on its own open decisions).

## Shared with the tank sim

The user's rulings for tanks ([tank_sim.md](tank_sim.md), 2026-11-02) that apply here too:

- **No death chance floor** as a hard rule. A score is found by bisection at 5% deaths instead, as the tank
  sim's "healing needed" is. The proposed healer score below is built the same way.
- **Some deaths no healing stops**: a hit that kills from full, or a burst between two heals. A healer set
  isn't blamed for those, so the bench reports them apart, as the tank bench's biggest hit and death floor do.
- **Forever's boss crit** (0.6% at level 63, `classic_crit` for Classic's 5.6%) holds for the boss's melee in a
  healing fight, since it's the same engine.
- **The stand-in healer** (the healing model) is what the tank sim scores with until real healers are
  simmed. The healing sim replaces it, and its cadence question (2 ± 1 s or the Molten Core logs' 0.3 s)
  goes away once real heals land.

**Boss encounters are one set for both** (agreed 2026-11-02 with the tank session): a boss is an encounter
with its melee, tank swaps and raid damage events. The tank session sources the bosses' melee and tank
abilities, and the healing side adds their raid damage to the same files.

## What the sim has today

Checked 2026-11-02.

**The engine can already aim a spell at another player.** An APL cast takes a target, and
`UnitReference{type: Player, index: N}` resolves to raid member N (`sim/core/environment.go`, `GetUnit`).
Attack tables are built for friendly pairs too, and a raid run puts every member in one sim. What's missing:

- **No smart target.** Nothing picks the lowest or most hurt player. The APL editor can't pick a player at
  all (`ui/core/components/individual_sim_ui/apl_helpers.tsx` lists Self, pets, targets and the pull target).
- **No overheal.** `dealHealingInternal` (`sim/core/spell_result.go`) books the whole heal as healing, and
  HPS includes overheal. Only healing threat leaves overheal out.
- **Only a tanked player loses health**, and only with a healing model, a health fight or a leveling session
  (`trackChanceOfDeath`, `sim/core/health.go`). The raid damage hits (`sim/core/raid_hits.go`) take no health
  on purpose: they feed procs and Rage in the DPS Report.
- **Death only books a flag.** A dead player keeps casting, and a heal brings it back.
- **Shields absorb nothing.** `Shield.Apply` (`sim/core/shield.go`) books shielding and an aura.
- **The boss's target is fixed at the start** (`Raid.tanks` and `Target.tank_index`). There's no tank swap.
  The only encounters are the generic Level 60 boss and Vaelastrasz, whose AI casts only Essence of the Red.

**Only the shaman can heal, and only itself.** By class:

| | Heals in the sim | Healing talents | Agent |
|---|---|---|---|
| Shaman | Healing Wave 1–10, Lesser Healing Wave 1–6, Chain Heal 1–3, Riptide, Healing Stream Totem, all with Forever rank data (`sim/shaman/healing.go`, `water_totems.go`). Each heals the shaman (`&shaman.Unit`), and Chain Heal never jumps: they were built for the leveling sim. Healing Stream heals the party. | all of them, tested against the client's curves | `sim/shaman/_restoration` is parked and doesn't compile |
| Priest | none: Flash Heal, Greater Heal, Renew, Prayer of Healing and Power Word: Shield are commented out and hold WotLK data. Lesser Heal, Heal, Holy Nova, Prayer of Mending and Penance's heal are missing. | Inner Focus and Holy Specialization only | `sim/priest/healing` compiles but isn't registered |
| Druid | none: Healing Touch, Regrowth, Rejuvenation, Swiftmend, Tranquility, Wild Growth and Bloom are missing | none of the healing ones | `sim/druid/_restoration` is parked and doesn't compile |
| Paladin | Lay on Hands on itself. Holy Shock deals damage only, and Light's Vigil is an empty stub. Holy Light and Flash of Light are missing. | Divine Favor on Holy Shock damage only | none (`register_all.go` points at a missing `paladin/holy`) |

None of the four healer specs is registered (`sim/register_all.go`), and their UIs hold talent presets with
blank gear and no working rotation. The priest's APLs use WotLK spell IDs. The presets are used today only as
the raid sim's placeholders, which give their buffs ([raid_sim.md](raid_sim.md), Tank and healer presets).

**Mana is simmed well**, for the DPS casters: regen, Spirit, the five second rule, consumables, time to OOM.
Innervate, Power Infusion and Mana Tide are in the raid sim. All of it carries over to healers.

## The design

### Health, damage and death

**A healing fight tracks every player's health.** A new encounter switch (`track_health`) turns it on: every
player loses health to every hit it takes, tank or not, and can die. The DPS Report doesn't set it, so its
numbers don't move.

**A dead player stops.** It casts nothing, the boss leaves it, and heals don't reach it. The fight goes on.
Death counts against the healers' score, and the raid sim also loses the dead member's damage from then on.

**The stand-in heal stream stays** for a tank sim with no healers, and is off for a player a simmed healer
heals.

### What deals the damage

**The boss's melee on its tank**, as today. A boss can have its own tank, and an encounter can swap tanks on a
timer or on a debuff's stacks, so a two-tank fight has both tanks taking hits in turn.

**Raid damage events**, a list on the encounter. Each one is an ability of the boss:

- its school and its damage range, before the player's resistances and damage taken modifiers
- when: every N seconds with a spread, or at random times with an average rate, from a start time
- who: every player, N random players, the tanks, the melee, or the players in one group (a cleave, a
  breath, a bolt volley on the nearest)
- once or over time (a DoT on its targets)

The DPS Report's raid damage hits stay as they are: a DPS sim doesn't heal, so they take no health there.

**Damage profiles** are named sets of those settings. They come from Classic and Season of Discovery bosses
whose numbers are on record, picked to cover the shapes healers meet:

| Shape | Stand-in | What it tests |
|---|---|---|
| One tank, steady | Patchwerk: heavy melee, Hateful Strike on the healthiest melee in reach | tank healing, the off-tanks' spikes |
| Two tanks, swapping | Broodlord Lashlayer or Ragnaros: a knockback that hands the boss to the second tank | healing two tanks, switching targets |
| Raid damage | Vaelastrasz or Sapphiron: steady damage on everyone | raid heals, mana over a long fight |
| Spiky random targets | Onyxia: Fireball on random members, Deep Breath, a phase of raid damage | reaction, scares |
| A 10-player raid | Season of Discovery's Blackfathom Deeps (Aku'mai or Kelris) | the small raid, two healers |

The numbers come from Warcraft Logs, which shows what Blizzard's servers did (the user, 2026-11-02): Season
of Discovery logs first, then Anniversary (Vanilla, then TBC), then Era, with vmangos only as a last resort
where no log shows a value ([data_sources.md](data_sources.md), Which source wins). Each profile names its
reports. The logs also show how real raids healed each boss: healer counts, HPS and overhealing, which the
sim's healers can be checked against. Forever's raids are 10, 20 and 40 players, so each
profile scales its damage to the raid. When Forever's bosses are known, each gets its own profile.

### Who heals whom

**Each healer has an assignment**: a set of players it heals, which the rotation's smart targets choose
from. An assignment is one or more tanks, a group, or the raid. With none, it's the whole raid.

**New targets for the rotation** (`UnitReference` types):

- **Most hurt**: the player in the assignment missing the most health (for a big heal)
- **Lowest health**: the lowest health percent in the assignment (for a fast heal)
- **Tank N**: the raid's Nth tank
- each of them only among living players, and none when nobody is hurt, so a heal on it waits

With these, an APL reads "Flash Heal on Lowest health when it's under 50%" or "Renew on Tank 1 when it's
missing". New values go with them: a unit's health missing, and its max health.

**Several healers on the same target**: in the game, a healer sees the other healers' casts, and in the sim
two healers can both start a big heal on the same tank, and one heal lands as overheal. That's real, it's
overhealing. A first version doesn't model healers watching each other's casts, and the overheal it causes is
part of the score.

**Chain Heal and the group heals** jump to the most hurt players in reach. Everyone is in reach of everyone
for now (the raid sim's ruling for buffs).

### What a healer scores

**Per heal**: healing, effective healing (what was missing), overhealing.

**Per healer**:

- **Effective HPS**: the healing that was needed, per second
- **Overheal %**
- **Mana**: healing per mana, time to OOM, mana left at the end
- **Deaths**: the chance that a player in its assignment dies
- **Scares**: how often a player in its assignment drops under 20% health and lives, a minute

**Proposed score** (an open decision): a healer set has to keep everyone alive, so the score is the **damage
it can carry**, the highest scale of the profile's damage at which the raid has a death in under 5% of fights,
found by bisection. At the profile's own damage it reports effective HPS, overheal, scares and mana left. It
mirrors the tank sim's "healing needed", and it's how the raid sim learns how many healers a comp needs.

## H1: what's in (2026-11-02)

- **Overhealing.** Every heal books the part that found its target at full health
  (`TargetedActionMetrics.overhealing`, `sim/core/spell_result.go`), and each unit has an effective HPS
  (`UnitMetrics.ehps`: healing less overhealing, no shielding) next to HPS. A target dummy has no health
  bar, so a heal on it is all effective. The healing table has an Overheal column.
- **The healing fight** (`Encounter.track_health`). Every player loses health to every hit it takes and
  dies when it runs out (`sim/core/player_death.go`): its cast stops, its rotation, swings and auras end,
  the mobs on it turn to the next living tank (or any living player once no tank is left), heals don't
  reach it and nothing it would cast goes. Its pets and dots carry on. A fight without the switch is
  unchanged, so the DPS Report's numbers didn't move.
- **Scares** (`UnitMetrics.scares_avg`): a dip under 20% health that we live through, counted once until
  we're back over 50%, per fight. A dip that ends in death is a death, not a scare.
- **Raid damage events** (`Encounter.raid_damage`, `sim/core/raid_damage.go`): school, damage range, who
  (everyone, the tanks, the others, or the boss's target, all of them or N random), timing (every N ± a
  spread, or random with an average) and an optional DoT. Physical goes through armor, the rest through
  resistances.
- **Tank swaps** (`Target.tank_swap_seconds`): the boss turns to the next living tank on a timer. Each
  iteration starts on the first tank again.
- **Heal targets in the rotation**: Most Hurt, Lowest Health and Tank N (`UnitReference`), the Health
  Missing value, and Max Health on any unit. Health % of Most Hurt or Lowest Health reads 100% when nobody
  is hurt, and a heal at them waits. The APL editor offers them, and every player, as cast targets.
- **Assignments** (`Player.heal_assignment`): the players and tanks a healer's Most Hurt and Lowest Health
  pick from. Proto only for now; the raid sim fills it in H6.
- **The shaman's heals** land on the player they're cast at, and on the shaman for a cast at a mob, so the
  leveling sim is unchanged. Chain Heal jumps to the 2 most hurt (within the target's party when it's in
  ours), each jump 50% of the one before. Riptide can go on anyone.
- **Tests**: `sim/raid_healing_test.go` (heals on others and overhealing, death stopping a player, tank
  swaps).

### First numbers

`tools/healbench`, 2 healbench shamans (synthetic level 60 caster gear, no mana support), a Protection
warrior and 7 Frost mages, 180 s, 500 iterations. The profiles are made up to test the engine:

| Profile | Scale | HPS | EHPS | Overheal | Mana left | Raid DTPS | Most deaths | Raid scares a fight |
|---|---|---|---|---|---|---|---|---|
| tank | 1 | 691 | 462 | 33.1% | 4456 | 472 | 0% | 0 |
| raid | 0.5 | 879 | 672 | 23.6% | 5055 | 703 | 0% | 0 |
| raid | 1 | 848 | 707 | 16.7% | 4641 | 1041 | 100% | 3.1 |
| swap | 1 | 872 | 629 | 27.9% | 2214 | 645 | 0.4% | 0.03 |

The raid profile at full damage is a cliff: the healers run dry around 85 seconds and the raid dies after.
That is the shape the proposed score measures (the damage a healer set can carry).

### Found on the way

- **Time to OOM reads wrong for a healer.** With no heal it can pay for, its rotation finds nothing to cast
  rather than waiting for mana, so the engine never marks it OOM. The estimate from the mana left at the end
  is then off, more so once the raid dies. The bench shows mana left instead. To fix in H2.
- **Two healers start the same big heal on the same tank**, and one lands as overhealing (the bench's Healing
  Wave on the tank). It's real, and part of the score, but it's also what makes the tank profile's
  overhealing 33%. A healer's rotation can wait for a bigger deficit.
- **An idle healer checks its rotation every 50 ms**, which costs some speed in a long fight.

## Milestones

**H1: the groundwork** (engine, no healer spec). Done 2026-11-02, see above.

1. Overheal: effective and overheal booked on every heal, in the metrics and the results UI.
2. `track_health`: every player loses health to every hit, death stops the player, heals skip the dead.
3. Raid damage events on the encounter.
4. Assignments and the new targets and values in the APL, and the APL editor offers them.
5. The shaman's heals aim at their target, and Chain Heal jumps.
6. A heal bench (`tools/healbench`), like the tank bench: a healer, a tank, a few DPS players and a
   profile, printing the scores above.

**H2: the Restoration shaman**, the first healer: its agent registered, gear, a rotation, a preset. Its spells
and talents are already in.

**H3: the other healers**, each with a rules audit against Forever's client data and foreverchanges (heal
ranks, costs, coefficients, healing talents), like the DPS class audits:

- **Priest**: Lesser Heal, Heal, Greater Heal, Flash Heal, Renew, Prayer of Healing, Power Word: Shield with
  absorbs (a new absorb pipeline), Holy Nova, Prayer of Mending, Penance; the Holy and Discipline talents
- **Druid**: Healing Touch, Regrowth, Rejuvenation, Swiftmend, Tranquility, Wild Growth, Bloom; the
  Restoration talents
- **Paladin**: Holy Light, Flash of Light, Holy Shock's heal, Light's Vigil; Illumination, Healing Light, Holy
  Power and the October 30 hotfix talents

**H4: the damage profiles**, sourced, and the score. Then healer gear (synthetic, like the DPS sets), talent
and rotation searches.

**H5: the DPS Report's healer section.**

**H6: the raid sim**: healers heal the simmed tanks and raid, the comp search counts healing, and later
chooses how many healers.

## Open decisions

1. **The score**: the damage a healer set can carry (proposed), or a weighted mix of effective HPS, mana,
   deaths and scares at a fixed damage.
2. **The scare line**: 20% health, or another.
3. **Absorbs** (Power Word: Shield): effective as long as they absorb, and the unused part as overheal
   (proposed).
