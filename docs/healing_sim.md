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
profiles and the seven boss files, whose raid damage now comes from Warcraft Logs ([Raid damage from the
logs](#raid-damage-from-the-logs-2026-11-02)). H2, the Restoration shaman, is registered with a tank healer
and a raid healer rotation ([H2](#h2-the-restoration-shaman-2026-11-02)), on stand-in gear. Next is H3, the
other healers.

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
- **The score** (2026-11-02, after the scoring agents' report): **damage carried**, the highest scale of a
  profile's damage the healers hold, with **separate lines**: tank deaths in under 2% of fights, the others'
  deaths in under 5%. Judged at **the encounter's own fight length**. **Separate tank healer and raid healer
  scores** per build, and presets per role. In the raid sim healing is a requirement with a **20% margin**
  (healers carry 1.2 times the encounter's damage), **checked against Warcraft Logs**: the healer counts,
  healing and deaths real raids had on each boss.
- **Downranking** (2026-11-02): no action that picks the exact rank out of all of a heal's ranks. Players
  bind 2 to 4 ranks of a heal and choose one of those by the situation, so a rotation does the same: it casts
  at most 4 ranks of a heal, each on its own condition.
- **Order:** the shared groundwork now. Nobody else is building it (the tank session stopped after its tank
  sheet and bench, waiting on its own open decisions).

## Shared with the tank sim

The user's rulings for tanks ([tank_sim.md](tank_sim.md), 2026-11-02) that apply here too:

- **No death chance floor** as a hard rule. A score is found by bisection at 5% deaths instead, as the tank
  sim's "healing needed" is. The proposed healer score below is built the same way.
- **Some deaths no healing stops**: a hit that kills from full, or a burst between two heals. A healer set
  isn't blamed for those, so the bench is to report them apart, as the tank bench's biggest hit and death floor
  do (not in `tools/healbench` yet).
- **Forever's boss crit** (0.6% at level 63, `classic_crit` for Classic's 5.6%) holds for the boss's melee in a
  healing fight, since it's the same engine.
- **The stand-in healer** (the healing model) is what the tank sim scores with until real healers are
  simmed. The healing sim replaces it, and its cadence question (2 ± 1 s or the Molten Core logs' 0.3 s)
  goes away once real heals land.

**Boss encounters are one set for both** (agreed 2026-11-02 with the tank session): a boss is an encounter
with its melee, tank swaps and raid damage events. The tank session sources the bosses' melee and tank
abilities, and the healing side adds their raid damage to the same files.

## What the sim had before H1

Checked 2026-11-02, before H1. H1 ([below](#h1-whats-in-2026-11-02)) has since added the engine pieces this
section lists as missing; the class table still holds, except for the shaman's heals on others.

**The engine could already aim a spell at another player.** An APL cast takes a target, and
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
| Shaman | Healing Wave 1–10, Lesser Healing Wave 1–6, Chain Heal 1–3, Riptide, Healing Stream Totem, all with Forever rank data (`sim/shaman/healing.go`, `water_totems.go`). Each healed the shaman (`&shaman.Unit`), and Chain Heal never jumped: they were built for the leveling sim. Healing Stream heals the party. | all of them, tested against the client's curves, but written for self heals: Ancestral Healing buffed only the shaman, and Healing Stream's heal was left out of the metrics (both fixed after H1's audit) | `sim/shaman/_restoration` is parked and doesn't compile |
| Priest | none: Flash Heal, Greater Heal, Renew, Prayer of Healing and Power Word: Shield are commented out and hold WotLK data. Lesser Heal, Heal, Holy Nova, Prayer of Mending and Penance's heal are missing. | Inner Focus, and the mana talents (Meditation, Mental Strength, Silent Resolve). Holy Specialization reaches only Smite, Holy Fire and Penance, and Spiritual Guidance gives spell damage only. | `sim/priest/healing` compiles but isn't registered |
| Druid | none: Healing Touch, Regrowth, Rejuvenation, Swiftmend, Tranquility, Wild Growth and Bloom are missing | the mana ones (Reflection, Furor), none of the healing ones | `sim/druid/_restoration` is parked and doesn't compile |
| Paladin | Lay on Hands on itself. Holy Shock deals damage only, and Light's Vigil is an empty stub. Holy Light and Flash of Light are missing. | Divine Favor on Holy Shock damage only | none (`register_all.go` points at a missing `paladin/holy`) |

None of the four healer specs is registered (`sim/register_all.go`), and their UIs hold talent presets with
blank gear and no working rotation. The priest's APLs use WotLK spell IDs. The presets are used today only as
the raid sim's placeholders, which give their buffs ([raid_sim.md](raid_sim.md), Tank and healer presets).

**Mana is simmed well**, for the DPS casters: regen, Spirit, the five second rule, consumables, time to OOM.
Healers get all of it. The raid sim's mana support isn't healer-aware yet: Innervate and Power Infusion go to
the first casters, Innervate waits for a mage's Evocation, and Mana Tide fires at a fixed time to its own party
([raid_sim.md](raid_sim.md), Assignments). Forever's Spirit regen tables are unverified
(`orchestration/forever_class_changes.md`), and the healer scores lean on them more than any DPS number does.

## The design

### Health, damage and death

**A healing fight tracks every player's health.** A new encounter switch (`track_health`) turns it on: every
player loses health to every hit it takes, tank or not, and can die. The DPS Report doesn't set it, so its
numbers don't move.

**A dead player stops.** It casts nothing, the boss leaves it, and heals don't reach it. The fight goes on.
Death counts against the healers' score, and the raid sim also loses the dead member's damage from then on.

**The stand-in heal stream stays** for a tank sim with no healers. The engine applies it whenever a player has
one (`Player.healing_model`), so a healing fight leaves it off: `tools/healbench` clears it.

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
| Spiky random targets | Onyxia: Fireball on random members, Deep Breath, a phase of raid damage. Onyxia's Lair is one of Forever's launch raids, so Forever's own numbers replace Classic's when they're known. | reaction, scares |
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

**Chain Heal** jumps to the most hurt players, within the target's party when it's in ours (the tooltip).
**The group heals** (Prayer of Healing, Tranquility, Holy Nova, Healing Stream, Light's Vigil's party heal)
reach a party and don't jump. Everyone is in reach of everyone for now (the raid sim's ruling for buffs), which
flatters Chain Heal's jumps between melee and ranged.

### What a healer scores

**Per heal**: healing, effective healing (what was missing), overhealing.

**Per healer**:

- **Effective HPS**: the healing that was needed, per second
- **Overheal %**
- **Mana**: healing per mana, time to OOM, mana left at the end
- **Deaths**: the chance that a player in its assignment dies
- **Scares**: how often a player in its assignment drops under 20% health and lives, a minute

**The score** (worked out 2026-11-02 by two independent agents, one from how healing is judged and simmed
elsewhere, one from experiments on `tools/healbench`; they agree). Proposed, for the user's decisions below:

- **Primary: damage carried.** The highest scale of a profile's damage at which fights end with a death in
  under 5% of them, found by bisection with interpolation between the two closest scales, paired seeds and
  about 1,600 iterations a step (about 8 sims a score). It weighs throughput against mana on its own, and it
  resolves a 3% gear change (+150 healing or +33 mp5 both moved the raid profile's carry from 0.666 to 0.688).
  It's the same shape as the tank sim's "healing needed".
- **Scored per role**, since the two rank builds differently: a **tank healer** (one healer, the tank's damage,
  tank deaths) and a **raid healer** (raid damage, the others' deaths). In the experiments the efficient
  rotation won the tank role (1.17 against 0.96) and the mid-rank rotation the raid role.
- **At a set fight length**, since length flips rankings: at 90 s throughput builds win, at 300 s efficient
  ones (raid profile: the top-rank rotation 0.80 against the efficient one's 0.71 at 90 s, 0.55 against 0.59 at
  300 s).
- **Next to it**: the profile's death floor (the deaths no healing stops, as the tank bench reports), mana left
  and time to OOM at a standard load, effective healing per mana, scares at the carry, overhealing as a
  diagnostic.
- **Not EHPS.** At a load everyone survives, every build's EHPS is the damage taken (seven tank rotations spent
  10.7k to 18.3k mana for the same 458 to 464 EHPS), and at an overload deaths cut the damage there is to heal.
- **Scares** mostly echo deaths: about none at a standard load, and they show near the cliff how a build fails.
  Reported, not scored.
- **The comp search uses healing as a constraint**, not a DPS term: a set of healers (with its tanks)
  qualifies when its carry at the encounter's damage is at least 1 plus a margin, for tank and raid deaths both.
  The healer count is the smallest that qualifies, and qualifying sets rank by raid DPS as today, so Mana Tide,
  Power Infusion and Innervate still count. Carry doesn't add up across healers (two healers on one tank carry
  far less than twice one), so sets are simmed whole: each tank's healers and the raid's healers on their own,
  carry tabulated per preset set, with fixed stand-in fillers.

**Engine metrics the score needs** (H2): per-iteration raid outcomes (anyone died, a tank died, first death,
dead player-seconds; "Most deaths (one player)" undercounts the chance anyone dies, 11.4% against 14.8%), a
healer's real OOM time, each healer's deaths and scares within its assignment, effective healing per mana, a
profile's death floor, and heals in flight so healers stop sniping each other.

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

`tools/healbench`, 2 bench shamans (synthetic level 60 caster gear; a Major Mana Potion, a Demonic Rune,
Mageblood and the raid's Wisdom, but no Mana Spring, Mana Tide or Innervate), a Protection warrior (two in
the swap profile) and 7 Frost mages, 180 s, 500 iterations, after the audit's fixes. HPS and EHPS are the two
healers summed, mana left is each healer's. The profiles are made up to test the engine:

| Profile | Scale | HPS | EHPS | Overheal | Mana left | Raid DTPS | Most deaths (one player) | Raid scares a fight |
|---|---|---|---|---|---|---|---|---|
| tank | 1 | 809 | 461 | 43.0% | 3820 | 470 | 0% | 0 |
| raid | 0.5 | 974 | 674 | 30.8% | 5287 | 703 | 0% | 0.01 |
| raid | 1 | 947 | 776 | 18.0% | 622 | 1104 | 99.8% | 3.1 |
| swap | 1 | 1032 | 633 | 38.6% | 1482 | 649 | 2.2% | 0.11 |

The raid profile at full damage is a cliff: the healers run dry and the raid dies. The bench rotation now keeps
Riptide on the tank, and with both healers doing it the overhealing on the tank profiles rose; the rotation is a
test of the engine, not a tuned one (H2).

### The audit (2026-11-02)

Three auditors read H1 cold (the engine, the shaman and the APL editor and the bench, and this doc against the
code). Fixed:

- **A dead player kept regenerating mana**, which put 4,600 mana back into dead healers and into the first
  numbers' raid ×1 row. A dead player's mana now stops (`mana.go`).
- **A wiped health fight never ended**: with nobody left to kill the boss, the fight ran on. It ends now.
- **A dead player's channel kept ticking** on the boss (a channel isn't a hardcast), and some auras turned
  its swings back on as they ended. Both stop now, and a dead player's totems go (its death effects run).
- **A heal at a player who died mid-cast** landed, cost its mana and still jumped. The cast stops unpaid now,
  as the game's fails.
- **Change Target to Most Hurt with nobody hurt crashed**, and Channel Spell at it fell back to the boss. Both
  wait now.
- **A tank swap leaked between iterations** for a boss with no tank of its own. Swaps need a starting tank.
- **Ancestral Healing buffed only the shaman.** Ancestral Fortitude goes on the healed target now (one buff
  per player, shared by every shaman's talent).
- **Healing Stream Totem's heal was left out of the metrics.** It counts now. No DPS Report number moved.
- **Health and Max Health of Most Hurt or Lowest Health read 0 when nobody was hurt**, while Health % read
  100%. All three read a full player's now, so "Lowest Health's Health < 1500" is false at full health.
- **The APL editor's health values lost the pets and Target N** when the heal targets came in. They're back.
- **The tests could pass with Chain Heal's jumps broken.** `TestHealingFightChainHealJumps` counts the jumps and
  their falloff, and new tests cover the dead player's regen, the wiped health fight and an empty heal target.
- Smaller: scares count for players only, and a periodic tick that kills skips the auras it just ended.

**Not fixed yet** (H2 and later), ranked by how much they would bend healer rankings:

1. **Done 2026-11-02** (What healers see, below the audit): ~~Healers see health perfectly, but not each other's heals.~~ No reaction delay, no heals in flight, no
   cancelling a heal that would overheal, and identical healers act in lockstep. The game had HealComm (heals in
   flight shown to other healers) and stop-casting. This penalises slow big heals (Greater Heal, Healing Touch,
   Holy Light, Healing Wave) and favours instants, HoTs and shields. Needed: an incoming heals value or a Most
   Hurt net of incoming heals, a stop-cast rule with a reaction delay, and per-healer jitter.
2. **The mana model decides the score**: healer-aware Innervate and Mana Tide, Mana Spring by group, Water
   Shield in the bench's rotation, Forever's Spirit regen, and the fight length each profile is judged at.
3. **Done 2026-11-02** (Bound ranks, below): ~~**Rank choice.** Classic healers' efficiency is mostly
   downranking. The APL can name a rank but not "the smallest rank that covers the deficit".~~ Not that
   action, by the user's ruling: a rotation binds up to 4 ranks and picks among them.
4. **Absorbs and armor from crit heals**: Power Word: Shield, Divine Aegis and Renewed Hope have no absorb
   pipeline, and Inspiration is missing.
5. **Range and groups**: "everyone in reach" flatters Chain Heal. Group heals and totems reach a party, and
   Prayer of Mending needs a damage taken hook.
6. **The damage model**: "the healthiest melee in reach" (Hateful Strike), players saving themselves
   (healthstones, potions, tank cooldowns), and Judgement of Light, which does nothing yet.
7. **Aura values on Most Hurt or Lowest Health** ("Riptide not on Lowest Health"): the rotation drops them when
   it's built, since nobody is hurt then. On a fixed tank they work.
8. **Time to OOM reads wrong for a healer**: with no heal it can pay for, its rotation finds nothing to cast
   rather than waiting for mana, so the engine never marks it OOM. The bench shows mana left.
9. **An idle healer checks its rotation every 50 ms**, which costs some speed in a long fight.

## The score in the bench (2026-11-02)

**The engine books how a healing fight went** (`RaidMetrics.healing_fight`): fights, fights with any death, a
tank's death and another's, the first death's time and the seconds players spent dead, as counts and sums so
results combine. Each unit has its mana spent a fight (`UnitMetrics.mana_spent_avg`), for effective healing per
mana. A healer that wants a heal it can't pay for is marked out of mana by the engine as any caster is; the
"time to OOM wrong" finding was the dead healers' regen, fixed with the audit.

**`tools/healbench`** scores by role (`-role tank|raid|both`), against a made-up `-profile` or a boss from
`tools/encounters` (`-encounter`, at the boss's own fight length):

- **tank**: the boss's melee and its abilities on the tanks, one healer by default, assigned the tanks; fails
  when a tank dies in 2% of fights.
- **raid**: the damage on the raid only, no tanks; fails when anyone dies in 5%.
- **both**: everything; either line fails.

`-carry` finds the damage carried (the scale of all damage at the line: bracketed, bisected five times on the
same seed, interpolated) and the death floor there (every player topped off every 1.5 ± 0.5 s, so only a
one-shot or a burst between heals kills; `-floor` prints it at `-scale`). About 10 sims, 2.5 s for a tank role
score at 800 iterations.

First carries, the bench shaman (a test rotation, not a tuned one), 1,600 iterations:

| Fight | Role | Healers | Carry | Death floor there | At ×1: deaths (tank / other), mana left |
|---|---|---|---|---|---|
| tank profile | tank | 1 | 0.956 | 0% | 5.6% / 4.7%, 1611 |
| raid profile | raid | 2 | 0.824 | 0% | - / 36.6%, 2062 |
| raid profile | both | 2 | 0.683 | 0% | 99.7% / 99.9%, 623 |
| swap profile | tank | 1 | 0.676 | 0% | 100% / 94%, 900 |
| Onyxia (tank side) | tank | 1 | 0.907 | 0% | 37.4% / 29.9%, 1381 |
| Broodlord (tank side) | tank | 1 | 0.818 | 0% | 38.6% / 36.5%, 3990 |

In the tank role "other" deaths are the healer's: once the tank dies the boss turns to it. Broodlord kills the
tank with 3,990 mana left in the healer: his Mortal Strike's healing cut and spikes, not mana, are the limit.

## What healers see (2026-11-02)

The audit's top gap, now in (`sim/core/heal_comm.go`), for healing fights only:

- **Heals in flight**, as Classic's HealComm addon showed them. A heal with a cast time declares its average
  size (`SpellConfig.ExpectedHealing`, the shaman's three heals so far; Chain Heal counts its target only, as
  HealComm did). Most Hurt, Lowest Health and Health Missing take the other healers' heals in flight off
  what's missing, so a second healer doesn't start the same big heal on the same hit.
- **A reaction time**: a healer (`Player.reaction_time_ms`) reads other players' health that long ago, from
  each player's last 32 health changes.
- **Cancelling** (`Player.stop_cast_overheal`): a reaction time before its heal lands, a healer whose heal
  would overheal by more than this share stops it, unpaid, and casts again at once. With no reaction time
  it never gets the chance.

`tools/healbench` gives its healers 0.3 s ± 25% (spread between them so identical healers don't act in
lockstep) and cancels over 50% overheal; `-reaction` and `-stopcast` change them. What each piece does, 800
iterations:

| Healers see | Tank profile, 2 healers: overheal, mana left, carry | Raid profile, both roles: carry |
|---|---|---|
| everything at once, no heals in flight (before) | 43%, 3,820, - | 0.683 |
| heals in flight, no reaction time | 26.2%, 6,685, 1.607 | 0.741 |
| heals in flight, 0.3 s late | 44.2%, 4,480, 1.564 | 0.686 |
| heals in flight, 0.3 s late, cancelling over 50% (the bench's default) | 22.6%, 6,323, 1.538 | 0.722 |

A late healer double-heals again (it sees the tank hurt after the other heal landed), and cancelling takes most
of that back. `TestHealingFightHealComm` checks it: without the heals in flight, 42 of 47 Healing Waves in a
fight started in lockstep with the other healer's.

## Bound ranks (2026-11-02)

Downranking the way players do it (the user's ruling): a healer binds 2 to 4 ranks of a heal and picks one by
what's missing, never the exact rank out of all ten.

- **Expected Healing** (`APLValueSpellExpectedHealing`, `sim/core/apl_values_spell.go`): what a heal (one
  rank) heals on average at a unit before crits, as its tooltip says with healing power, healing bonuses and
  the target's healing taken (a Mortal Strike shows). So a line reads "Healing Wave rank 8 at the tank when
  Health Missing (Tank) >= Expected Healing (Healing Wave rank 8, Tank)", with no magic numbers to retune for
  gear. With no player to heal, it's the heal on the caster.
- **At most 4 ranks of a heal** a rotation (`MaxBoundHealRanks`, `sim/core/apl.go`): a priority list line
  that would cast a 5th rank of a heal is off, with a warning on it in the editor.
  `TestHealingFightBoundRanks` checks both.

The bench's Restoration shaman binds 3 ranks of each heal now: Healing Wave 10, 8 and 6 at the tank, and
Chain Heal 3, 2 and 1 at the most hurt, each the biggest that what's missing covers. Against the old rotation
(Healing Wave 9 past 1,800 missing, Chain Heal 3 past 1,200) and other bind sets, carry with one tank
healer (2,000 iterations), and the raid profile's raid healer (500):

| Bind set | Onyxia | Broodlord | Magmadar | Raid profile |
|---|---|---|---|---|
| old: HW 9, CH 3 on fixed lines | 0.895 | 0.795 | 1.661 | 0.833 |
| HW 10, 7; CH 3, 1 | 0.917 | 0.817 | 1.737 | 0.856 |
| **HW 10, 8, 6; CH 3, 2, 1** (the preset) | **0.930** | 0.820 | **1.766** | **0.876** |
| HW 10, 8; CH 3, 1, cast at 80% of their size | 0.871 | 0.835 | 1.574 | 0.868 |

A rank cast at 80% of its size overheals more and runs out of mana sooner on the long fights. One rank
(HW 10, CH 3 alone) carries less than the old rotation on Onyxia (0.863).

Still open from the audit's list: absorbs, healer-aware mana support, range and groups, the damage model's
gaps.

## Raid damage from the logs (2026-11-02)

The seven boss files (`tools/encounters`, shared with the tank sim) have their raid side from Warcraft Logs: the
boss's damage on the players who aren't tanks as `raidDamage` events with `who` NotTanks, and what real raids
had in `raidHealing`. The tank sim scores on the Anniversary realms' set (`tools/encounters/*.json`, the user's
ruling to the tank session), so that's the one benched here; Season of Discovery's is in `tools/encounters/sod/`.
How they're mined and written: `tools/warcraftlogs/wclcli.py raid` (`wclraid.py`) and
`tools/encounters/raid_from_logs.py`, README in both folders.

- **The tanks** in a log are the players the bosses swung at (5% or more of the swings) and the players who
  tanked an add (30% or more of its swings: Golemagg's two Core Rager tanks, Broodlord's whelps). The rest are
  the non-tanks, and their damage taken is the raid side. The players' own damage (runes, sapper charges,
  Hellfire, Seal of Martyrdom) is left out.
- **An event** is an ability with 2% or more of the non-tanks' damage in 10% or more of the kills: its
  unmitigated hit, the share of the non-tanks a cast hits (the new `RaidDamageEvent.share`, so the file fits a
  10, 20 or 40 player raid), a DoT's ticks, and the rate that lands the logs' damage a second on each non-tank
  after its first cast. Every file's events land what the logs did, within 0.1 a second.
- **`healbench -role raid` and `both`** run the logs' median kill (Onyxia 137 s, Chromaggus 115 s), since the
  file's `duration` is the main tank's time on the boss.

What the raids had (Anniversary, 21 to 30 kills a boss; effective healing a second, the healers' HPS with
overheal; Warcraft Logs' healer role):

| Boss | Raid | Tanks | Healers | HPS each | Raid HPS | On tanks | On the rest | Each non-tank's DTPS | Deaths a kill (others / tanks) | Kill |
|---|---|---|---|---|---|---|---|---|---|---|
| Magmadar | 40 | 1 | 11 | 328 | 1,760 | 599 | 1,058 | 23.6 | 0 / 0 | 54 s |
| Golemagg | 40 | 3 | 11 | 401 | 2,069 | 1,539 | 496 | 11.9 | 0 / 0 | 56 s |
| Ragnaros | 40 | 2 | 11 | 322 | 2,737 | 821 | 1,823 | 47.5 | 0.07 / 0.04 | 97 s |
| Onyxia | 40 | 3 | 11 | 182 | 1,335 | 399 | 895 | 22.6 | 0.05 / 0 | 137 s |
| Broodlord | 40 | 4 | 11 | 439 | 2,071 | 869 | 1,109 | 29.0 | 0 / 0 | 44 s |
| Vaelastrasz | 40 | 1 | 11 | 790 | 4,900 | 1,187 | 3,608 | 106.9 | 0 / 0 | 31 s |
| Chromaggus | 40 | 1 | 11 | 267 | 1,171 | 724 | 417 | 8.6 | 0.05 / 0 | 115 s |

Season of Discovery's Normal raids are 17 to 24 players with 4 or 5 healers, each non-tank taking 43 to 237 a
second (two to four times the Anniversary realms'), and up to 2.5 deaths a kill.

**The bench against them** (the bench's shaman, 1,000 iterations; `raid` is 2 healers and 7 Frost mages,
`both` adds the tank, and 3 healers is the logs' 11 to 40):

| Boss | raid: carry, deaths at ×1 | both: carry, deaths (tank / other) | both, 3 healers |
|---|---|---|---|
| Magmadar | 1.54, 0% | 1.20, 0% / 0.2% | 1.46 |
| Golemagg | 1.92, 0% | 1.35, 0% / 0% | 1.69 |
| Ragnaros | 1.20, 1.1% | 0.52, 98% / 97% | 0.56 |
| Onyxia | **0.67, 56%** | 0.63, 10% / 82% | 0.71 |
| Broodlord | **0.80, 20%** | 0.68, 48% / 61% | 0.74 |
| Vaelastrasz | 1.07, 3% | 0.57, 99% / 98% | 0.62 |
| Chromaggus | 0.93, 8% | 0.82, 17% / 30% | 0.91 |

- **`both` doesn't compare with the logs as it stands.** The tank's damage doesn't shrink with the raid: a
  10-player bench raid puts 2 or 3 healers on a tank that 40-player raids gave 6 to 8 (the tank session's
  `healing.healers`). Ragnaros and Vaelastrasz kill the tank whatever the raid healers do. The check against the
  logs is per role: `tank` at the file's tank healers, `raid` at the logs' raid healers a non-tank.
- **`raid` carries 1 or more on four bosses**, as the logs' deathless kills say it should. Onyxia, Broodlord and
  Chromaggus fall short with mana to spare (6,458, 7,145 and 7,441 left), so it's spikes, not throughput:
  Broodlord's Blast Wave (2,678 to 3,336 Fire on half the raid at once), Onyxia's Flame Breath and Fireballs on
  a Frost mage's health. A 40-player raid answers a spike with any of its 11 healers; the bench's 2 at the same
  ratio react slower. The death floor is under 2% on all three, so a heal could have saved them.
- Not yet in the files: the logs' cast timing and phases (each event is at random times after its first cast),
  players' fire resistance gear and self-saves (the logs' damage is after them; the sim's mages take it
  unmitigated by gear), and who stands where (each cast hits random non-tanks).

## H2: the Restoration shaman (2026-11-02)

**The spec is registered** (`sim/shaman/restoration`): a shaman agent with no auto attacks and the shaman
imbue option of the other specs, in place of the old wowsims package (`_restoration`, never built). Its
heals, Water Shield, Mana Spring Totem (its party, `totem_receivers.go`) and Mana Tide Totem (its party, the
5 min cooldown) were already the shaman's. `tools/healbench`'s requests are the spec now: the bench ran the
same numbers on it as on the Elemental agent it used before. `TestHealingFightRestorationManaTools` checks
the mana tools are cast and reach the party.

**Two rotations, one per role** (`ui/restoration_shaman/apls`, the bench's `shaman_restoration.json` and
`shaman_restoration_raid.json`). Both put Water Shield up at -3 s and Mana Spring Totem at -1.5 s, cast
Lesser Healing Wave at whoever in the assignment is under 35% health, Mana Tide Totem at 35% mana and Mana
Spring again after it, and bind at most 4 ranks of a heal.

- **Tank healer:** Riptide on the tank whenever it's off, Healing Wave 10, 8 and 6 at the tank, Chain Heal 3,
  2 and 1 at the most hurt. Water Shield goes back up as soon as its globes are spent (a tank healer is seldom
  hit; its globes go on its heals' crits).
- **Raid healer:** Riptide at the most hurt when it fits, Chain Heal 3, 2 and 1 at the most hurt, Healing Wave
  10 and 8 for one big hole. Raid damage spends Water Shield's globes every few hits, and a recast is a global
  cooldown, so it goes back up only below 60% mana (0.899 against 0.874 recasting at once on the raid profile,
  1.140 against 1.056 on Vaelastrasz, 0.865 against 0.843 at 300 s).
- Mana Tide at 20, 35, 50 and 70% mana carried the same within the bench's ±0.01 at 120, 180 and 300 s on
  Onyxia's tank side; 90% was 0.01 to 0.02 behind.

Against the bench's earlier rotation (the same heals, no mana tools; 1,000 iterations; the tank role at each
file's tank time, the raid role at the logs' median kill, 2 healers and 7 Frost mages):

| Boss | tank: old → tank healer | raid: old → raid healer |
|---|---|---|
| Magmadar | 1.548 → 1.518 | 1.539 → 1.583 |
| Golemagg | 1.462 → 1.402 | 1.913 → 1.990 |
| Ragnaros | 0.458 → 0.459 | 1.197 → 1.230 |
| Onyxia | 1.031 → 1.031 | 0.672 → 0.695 |
| Broodlord | 0.656 → 0.643 | 0.796 → 0.844 |
| Vaelastrasz | 0.661 → 0.659 | 1.066 → 1.140 |
| Chromaggus | 0.778 → 0.764 | 0.928 → 0.990 |
| Onyxia's tank side at 180 s / 300 s | 0.836 → 0.959 / 0.677 → 0.830 | |

- **The raid healer carries more on every boss.** Its overheal is 21 to 56% against the old rotation's 10 to
  25%: Riptide's HoT and Chain Heal's jumps land on players who are full again.
- **The tank healer is a long-fight rotation.** At the files' tank times (17 to 71 s, the main tank's stint)
  mana never runs short, and Water Shield, Mana Spring and Mana Tide only cost globals: up to 0.06 behind
  (Golemagg). From 180 s it carries a good deal more, since the old rotation runs dry at 178 s on Onyxia.
- **Gear is a stand-in**, the Elemental synthetic set, which carries spell hit a healer doesn't use. A healer
  set needs a healer stat mix in `orchestration/gear_targets.md`, and that table's source (wowsims/sod's level
  60 BiS lists) has none for any healer (the user's decision, open).
- The talents are the bench's (31 Restoration to Riptide, Ancestral Knowledge, Mental Dexterity and
  Shamanistic Focus), not searched: `tools/rotopt` scores DPS.
- The UI presets (builds, rotations, talents, consumes, a raid sim preset) are in; the spec stays unlaunched.

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
  absorbs (a new absorb pipeline), Holy Nova, Prayer of Mending, Penance's heal, Binding Heal (trained at 25
  since the October 30 hotfix); the Holy and Discipline talents, with Holy Binding, Divine Aegis, Renewed Hope,
  Litany of Light, Power in Light and Inspiration
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

1. **The scare line**: 20% health, or another; reported only (proposed).
2. **Absorbs** (Power Word: Shield): effective as long as they absorb, and the unused part as overheal
   (proposed).
