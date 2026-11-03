# The Restoration druid (H3)

The healing sim's druid healer ([healing_sim.md](healing_sim.md), H3; the workers' brief
`orchestration/brief_healers.md`). Before this the sim had no druid heal at all: Healing Touch, Regrowth,
Rejuvenation, Swiftmend, Tranquility, Wild Growth and Bloom were missing, and the old wowsims package
(`sim/druid/_restoration`) never built. Now:

- **The heals** are in `sim/druid/healing.go`, every value from the Forever client, build 1.60.1.70753 (the
  November 2 build, which wrote the launch realms' October 30 hotfix into the tables). HoTs go through the dot
  pipeline (`Spell.Hot`), direct heals through `CalcAndDealHealing` with `OutcomeHealingCrit`, and every heal
  has `SpellConfig.ExpectedHealing` for heals in flight and bound ranks.
- **The agent** is `sim/druid/restoration` (registered in `sim/register_all.go`): a druid in caster form with no
  auto attacks, Innervate on itself unless its options name someone else.
- **Tests**: `sim/raid_healing_druid_test.go` and `sim/raid_healing_druid_talents_test.go`, a test per heal and
  per healing talent on small healing fights, most of them reading one iteration's debug log (each heal's base,
  healing power and caster multiplier, each cast's cost and cast time).
- **Gear**: the Restoration druid synthetic set (`tools/database/forever_synthetic_items_druid_restoration.go`,
  ids 990401 to 990413) on the Restoration druid row of `orchestration/gear_targets.md`.
- **Presets**: a tank healer and a raid healer rotation (`ui/restoration_druid/apls`,
  `tools/healbench/req/druid_restoration*.json`), tuned on `tools/healbench`. The spec page stays unlaunched.

## Rules audit

Sources, best first ([data_sources.md](data_sources.md)): the Forever client's tables on wago.tools (1.60.1.70753:
SpellEffect, SpellMisc, SpellLevels, SpellPower, SpellCastTimes, SpellDuration, SpellCooldowns, SpellClassOptions,
SpellAuraOptions, SkillLineAbility, TraitDefinitionEffectPoints through `tools/forever_talents/data/druid.json`),
then the development notes and foreverchanges.pro, then Season of Discovery, then later expansions' behaviour
(only for what the client leaves to the server). The 70707 → 70753 diff (`orchestration/update_ledger.md`, build
70753) holds the October 30 hotfix for the druid: Bloom in Gift of the Earthmother's place, Naturalist's cast time
cut on Wild Growth, Wild Growth's 3 sec cast, Reflection and Subtlety swapping rows, Heart of the Wild and Furor
swapping trees. The 70374 → 70425 diff took the "can't be overwritten by a weaker rank" bit off every Rejuvenation
and Regrowth (development notes, 70425: one druid's high rank no longer blocks another's).

A rank's heal at level 60 is its base plus its growth a level up to its max level, spread evenly by the variance
(the shaman's reading of the same tables, `sim/shaman/healing.go`, which gives the DPS spells' tables to the
point). The sim's value is the client's in every row below; "sim before" is what the sim had before H3.

### Healing Touch (Nature, 25297 at 60)

| Rank | Spell | Learned | Heal at 60 | Mana | Cast | Coefficient | Sim before |
|---|---|---|---|---|---|---|---|
| 11 | 25297 | 60 | 2139 to 2525 (2332, +7.3 a level to 65, ±8.3%) | 840 | 3.5 s | 1.0 | missing |
| 10 | 9889 | 56 | 1786 to 2107 (1946) | 755 | 3.5 s | 1.0 | missing |
| 9 | 9888 | 50 | 1415 to 1676 (1546) | 630 | 3.5 s | 1.0 | missing |
| 8 | 9758 | 44 | 1100 to 1311 (1205) | 520 | 3.5 s | 1.0 | missing |
| 7 to 1 | 8903 … 5185 | 38 … 1 | 921 … 45 | 425 … 25 | 3.5 … 1.5 s | 1.0 … 0.429 | missing |

Forever's Healing Touch is below 1.12's (rank 11 2267 to 2677). Downranking: 5% of the coefficient a level past
the rank's learn level plus 17 (Forever's rule, the shaman's and priest's `downrankCoefficient`), so rank 9 keeps
all of it at 60, rank 7 75%, rank 4 none.

### Regrowth (Nature, 9858 at 60)

| Rank | Learned | Direct heal at 60 | HoT | Mana | Cast | Coefficients | Sim before |
|---|---|---|---|---|---|---|---|
| 9 | 60 | 965 to 1077 (1021 ±5.5%) | 142 every 3 s, 7 ticks (994) | 525 | 2 s | 0.286 direct, 0.071 a tick | missing |
| 8 | 54 | 815 to 911 | 113 × 7 | 445 | 2 s | same | missing |
| 7 | 48 | 659 to 738 | 88 × 7 | 370 | 2 s | same | missing |
| 6 to 1 | 42 … 12 | 527 … 92 | 68 … 13 × 7 | 305 … 70 | 2 s | same | missing |

Forever made Regrowth much cheaper than 1.12 (rank 9 880 mana there) and its HoT smaller (152 a tick). The
Season of Discovery rows (436937 and on, 1.12's values) aren't trained in Forever (no SkillLineAbility row).

### Rejuvenation (Nature, 25299 at 60)

| Rank | Learned | Tick | Ticks | Mana | Coefficient a tick | Sim before |
|---|---|---|---|---|---|---|
| 11 | 60 | 194 | 4, every 3 s (12 s) | 360 | 0.2 | missing |
| 10 | 58 | 161 | 4 | 335 | 0.2 | missing |
| 9 | 52 | 124 | 4 | 280 | 0.2 | missing |
| 8 | 46 | 94 | 4 | 235 | 0.2 | missing |
| 7 to 1 | 40 … 4 | 71 … 8 | 4 | 195 … 25 | 0.2 | missing |

Instant. Forever cut it (1.12's rank 11 222 a tick). Rank 11 is a trainer rank at 60 in Forever (1.12's was an
Ahn'Qiraj book; SkillLineAbility has it with AcquireMethod 0).

### Tranquility (Nature, 9863 at 60)

| Rank | Learned | Tick on each party member | Ticks | Mana | Cooldown | Coefficient a tick | Sim before |
|---|---|---|---|---|---|---|---|
| 4 | 60 | 285 (+1 a level to 66) | 5, every 2 s, channeled (10 s) | 925 | 5 min | 0.067 | missing |
| 3 | 50 | 205 | 5 | 695 | 5 min | 0.067 | missing |
| 2 | 40 | 136 | 5 | 505 | 5 min | 0.067 | missing |
| 1 | 30 | 91 | 5 | 375 | 5 min | 0.067 | missing |

An area aura on the caster's party (effect 35). 1.12's rank 4 was 294 a tick for 1375 mana.

### Wild Growth (Nature; the 31 point talent, rank 1, and trainer ranks 2 and 3)

| Rank | Spell | Learned | Tick | Ticks | Mana | Cast | Cooldown | Coefficient a tick | Sim before |
|---|---|---|---|---|---|---|---|---|---|
| 3 | 1238215 | 60 | 97 (+1.6 a level to 68) | 7, a second apart | 1050 | 3 s | 6 s | 0.033 | missing |
| 2 | 1238214 | 50 | 73 | 7 | 755 | 3 s | 6 s | 0.033 | missing |
| 1 | 408120 | 40 | 48 at 48 | 7 | 550 | 3 s | 6 s | 0.033 | missing |

Heals the target and its party, 5 at most (effect 2, a dummy of 5). The 3 sec cast is the October 30 hotfix's
(Season of Discovery's is instant). "The amount healed is applied quickly at first, and slows down": the client
leaves the shape to the server, so the sim takes Wrath of the Lich King's (the first tick 6% over the average, 2%
less each tick, the last 6% under), which keeps the total. That's a call for the user below.

### Swiftmend (18562, the Restoration row 4 talent)

| Value | Client | Sim | Sim before |
|---|---|---|---|
| Cost | 20% of base mana | same | missing |
| Cooldown | 15 s (Idol of Swiftness and the Tier 1 5 piece 3 s each off) | same | missing |
| Heal | "an amount equal to the full duration of the periodic effect of the consumed spell" | every tick of the HoT (Nature's Splendor's too) at its snapshot (base and healing power), with Swiftmend's own modifiers | missing |
| Which HoT | not in the client | the one with the least time left (1.12 and later) | |

Gift of Nature and Genesis name Swiftmend in their class masks (Gift of Nature's first effect, Genesis's second, an
all-of-the-spell modifier), so the sim gives Swiftmend those and not the HoT's own (Improved Rejuvenation, the
periodic ones). That's a call below.

### Bloom (1328621, the October 30 hotfix's Restoration row 3 talent)

An expiring Rejuvenation or Regrowth heals its target for 100% of our Spirit (50% if dispelled, which doesn't
come up). The heal (1328622) has no class mask, so no healing talent reaches it. The sim heals at the HoT's last
tick, not when it's recast, consumed by Swiftmend or ends with its target. It can crit (no "can't crit" bit).

### Nature's Swiftness (17116)

The next Healing Touch, Regrowth or Wrath (its class mask; Wild Growth isn't in it) is instant. 3 min cooldown,
from the buff's use; off the global cooldown. A rotation that casts it controls it; one that doesn't has it used
with the other cooldowns.

### Innervate on ourselves (29166)

400% more Spirit regen and all of it while casting, 20 s, 6 min cooldown, 5% of base mana: the client's, already
the sim's (`sim/core/buffs.go`). The Restoration druid casts it on itself unless its options name another
target. A rotation names it by spell ID alone: an Innervate on oneself no longer carries the druid's raid index as
its tag (no DPS request changes: the DPS report's druids are each player 0).

### Healing talents

| Talent | Client (rank values from TraitDefinitionEffectPoints) | Reaches (class mask) | Sim before | Sim now |
|---|---|---|---|---|
| Gift of Nature (17104) | 2% a point, 10% at 5 | all of Healing Touch, Regrowth, Tranquility, Swiftmend; the periodic part of Rejuvenation, Regrowth, Wild Growth | nothing | each heal once, Regrowth's HoT included |
| Improved Rejuvenation (17111) | 5% a point, 15% at 3 | Rejuvenation | nothing | as the client |
| Genesis (1223081, Balance) | 2% a point periodic damage and healing, 10% at 5 | periodic: Rejuvenation, Regrowth, Wild Growth (and Moonfire, Insect Swarm, Rake, Rip); all of Swiftmend | damage only | heals too; not Tranquility (not in its mask) |
| Tranquil Spirit (24968) | 2% a point off the cost, 10% at 5 | Healing Touch, Tranquility | nothing | as the client |
| Naturalist (17069) | 0.1 s a point off the cast, 0.5 at 5; 2% a point damage | Healing Touch, Wild Growth (the hotfix) | damage only | the cast time cut too |
| Improved Regrowth (17074) | 10% a point crit, 50% at 5 | Regrowth | nothing | the spell's crit, so its ticks too |
| Nature's Splendor (1223083, Balance) | Rejuvenation 3 s, Regrowth 6 s more | Moonfire, Rejuvenation, Regrowth, Insect Swarm | Moonfire and Insect Swarm only | one more Rejuvenation tick, two more Regrowth ticks |
| Improved Tranquility (17123) | 30% a point off the cooldown, 50% a point off the threat | Tranquility | nothing | as the client |
| Nature's Focus (17063) | 14% a point, 70% at 5, not to lose cast time to a hit | Healing Touch, Regrowth, Tranquility (and Wrath, Starfire, Hurricane …) | damage spells | Healing Touch and Regrowth too (the sim pushes back hardcasts only) |
| Reflection (17106) | 17 / 33 / 50% of regen while casting | | in | unchanged |
| Living Spirit (1309631) | 5% Spirit a point | | in | unchanged |
| Heart of the Wild (17003) | 2% Intellect a point | | in | unchanged |
| Subtlety (17118) | 10% a point less threat from Nature and Arcane spells | | in | heals take it (they're Nature) |
| Nature's Majesty (1223082, Balance) | 2% crit a point with spells and attacks | | in | heals take it (spell crit) |
| Nature's Grace (16880, Balance) | a non-periodic spell crit: 10% cast speed and 10% off the global cooldown for 3 s | its proc flags take helpful spells (0x4000); its global cooldown cut names the heals but Swiftmend | damage spells only | heal crits proc it, and it cuts the heals' global cooldown |
| Omen of Clarity (16864, baseline) | spells and attacks proc Clearcasting, which makes the next damage or healing spell free | its proc flags take helpful spells; Clearcasting's mask names every heal | damage spells only | heals roll at the sim's spell rate (2 a minute over the cast time) and Clearcasting makes them free |
| Natural Instinct (1223242, Feral) | 12 / 25% of Intellect as spell healing (and the melee crit damage) | | crit damage only | the healing too, on the Restoration druid |
| Natural Shapeshifter, Improved Mark of the Wild … | no healing | | | |

HoT ticks crit under Forever: every druid HoT has SPELL_ATTR8_PERIODIC_CAN_CRIT (Attributes_8 0x200) since
build 70009, as every Forever dot rolls its ticks (`sim/core/ruleset.go`).

### Items

- **Idol of Swiftness** (279250, the Forever guide's BiS idol): 3 s off Swiftmend's cooldown (1306488).
- **Forever's Tier 1 Restoration set** (the bonus spells 1300943, 1301006, 1301077, 1301250): +10 Spirit, Barkskin
  10 s, +26 healing (9 of it damage), Swiftmend 3 s. The synthetic set carries it on two pieces (standards 6).
- Not simulated: Idol of Synthesis (each HoT on the target takes 1 s off Swiftmend's remaining cooldown) and Idol
  of Rejuvenation.

## Calls for the user

Each is what the client leaves to the server or reads two ways. The default is what the sim does; the measured
effect is on the presets where it was measured.

1. **Wild Growth's shape.** Default: Wrath of the Lich King's front-loaded ticks (+6% to −6%, the same total).
   The alternative is 7 even ticks. It only moves overhealing.
2. **Swiftmend's heal.** Default: every tick of the consumed HoT including Nature's Splendor's (5 Rejuvenation
   ticks, 9 Regrowth ticks), at the HoT's snapshot base and healing power, with Swiftmend's own modifiers (Gift of
   Nature, Genesis) in place of the HoT's. Alternatives: 1.12's fixed 12 sec of Rejuvenation or 18 of Regrowth (4
   and 6 ticks), and/or the HoT's own modifiers carried over (Improved Rejuvenation's 15% on a consumed
   Rejuvenation).
3. **Which HoT Swiftmend takes** with both up: the one with less time left (1.12 and later). Untested in Forever.
4. **Improved Regrowth on the HoT's ticks.** Default: yes, the 50% crit is the spell's, and Forever's HoT ticks
   crit on the spell's crit chance (every druid HoT has the "periodic can crit" bit). The alternative is the direct
   heal only.
5. **Gift of Nature on Tranquility.** Default: yes, 10%: its first effect names Tranquility, though Tranquility is
   periodic and that effect is the "all of the spell" kind. A strict reading (that kind of modifier reaches direct
   heals only) leaves Tranquility without it.
6. **Omen of Clarity on heals.** Default: a heal rolls at the sim's spell rate (2 procs a minute over its cast
   time, the 1.12 melee rate the Balance druid's spells use too), when it's cast. The client leaves the chance to
   the server.
7. **Bloom can crit** (the heal has no "can't crit" bit). Default yes.
8. **Downranking.** Default: Forever's rule as the shaman's and priest's heals have it (5% of the coefficient a
   level past the rank's learn level plus 17). The October 1 notes give no numbers; a rank 7 Healing Touch at 60
   keeps 75%.
9. **The raid sim's Restoration druid placeholder** (`sim/compsearch/presets.go`) keeps its old talents; whether
   it takes the bench's build is the raid sim's (H6).
10. **Launching the spec page** (`ui/core/launched_sims.ts`) is the user's call; it stays unlaunched.

## Tuning so far (WIP, before the game update)

`orchestration/tools/healers/druid_restoration_apl.py` writes the two rotations from knobs and
`druid_restoration_req.py` the bench requests (Thicc Resto `05302001--5053025153113051`, Windshaper Skyborne,
the synthetic set). Tank healer, 500 iterations, carry: Healing Touch bound at ranks 11, 8 and 6 beats 11, 10, 9
on every file (Onyxia 1.165 against 1.094, Broodlord 0.730 against 0.701, Vaelastrasz 0.695 against 0.617,
Chromaggus 0.859 against 0.790, Onyxia at 180 / 300 s 1.062 / 1.042 against 0.959 / 0.911); Regrowth's HoT and
Swiftmend on the tank both pay; the Innervate threshold doesn't matter (5,000 mana left at 300 s). Raid healer,
400 iterations: Rejuvenation at the most hurt once it misses 30% of the HoT is best of the first round (Onyxia
0.684, Broodlord 0.821, Vaelastrasz 1.098, Chromaggus 0.917, Ragnaros 1.196). The first baseline against the
shaman's presets (1,000 iterations, tank role): Magmadar 1.541 / 1.536, Golemagg 1.503 / 1.440, Ragnaros 0.479 /
0.467, Onyxia 1.087 / 1.016, Broodlord 0.710 / 0.618, Vaelastrasz 0.620 / 0.639, Chromaggus 0.783 / 0.764.
Next: round 3 (talent builds, races, more raid knobs), then the presets and the full numbers.
