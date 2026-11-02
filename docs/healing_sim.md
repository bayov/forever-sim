# Healing sim

Healers that heal: healer presets that cast real heals and spend real mana, damage on the tanks and the
raid for them to heal, and a score that ranks healer builds and tells the raid sim how many healers a comp
needs and which. The user, 2026-11-02: "We also want to simulate proper healing soon."

It meets the tank work ([tank_sim.md](tank_sim.md)) in the middle. Tanks need someone healing them for the
chance of death to mean anything, which today is the engine's stand-in heal stream. Healers need damage to
heal, which is the boss hitting the tanks plus the encounter's damage on the raid. Both use the same health
and damage groundwork, which this doc's first milestone builds.

**Where it stands (2026-11-02):** the design and the user's first rulings. Nothing simmed yet.

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

The numbers come from the bosses' scripts in the open Classic server cores (vmangos, cmangos) and Wowhead's
Classic and SoD spell pages, with sources on each profile. Forever's raids are 10, 20 and 40 players, so each
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

## Milestones

**H1: the groundwork** (engine, no healer spec)

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
