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
- **Tests**: `sim/raid_healing_druid_test.go`, `sim/raid_healing_druid_talents_test.go` and
  `sim/raid_healing_druid_more_test.go` (plus `sim/druid/healing_ranks_test.go`, every rank against the client), a test per heal and
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

### The November 2 notes (round 70753)

The healing lines of the development notes (post 9, 2026-11-02; `orchestration/round_70753/druid.md` rows 172 to
178 hand them to this worker), against client 1.60.1.70753 (70755's tables are the same). Every one is in the
client, so none is notes-only (standards 37).

| Note line | Client 70753 | Sim |
|---|---|---|
| Bloom: new row 3 talent, an expiring Rejuvenation or Regrowth heals for 100% of Spirit (50% if dispelled) | 1328621 (bp 100 / 50, TraitDefinition 134379 replacing Gift of the Earthmother 414673), heal 1328622 | in (`registerBloomSpell`), `TestDruidBloom` |
| Wild Growth: 3 sec base cast | SpellMisc of 408120 / 1238214 / 1238215 | in, `TestDruidWildGrowth` |
| Naturalist: 2/4/6/8/10% damage, and 0.1 to 0.5 sec off Wild Growth | 17069's description and class mask (hotfix 113283) | the damage in `talents.go` (Feral and Balance worker), the cast time cut here; `TestDruidHealingTalentMultipliers` |
| Swiftmend needs the caster's own Rejuvenation or Regrowth | 18562's description: "with your active Rejuvenation or Regrowth" | in: `swiftmendHot` reads only our ranks; `TestDruidSwiftmend` |
| Gift of the Earthmother removed | gone from the tree | nothing to remove (the sim never had it) |
| Reflection and Subtlety swap rows | TraitDefinition 134378 / 134381 | tree position only (the tree file); talent strings unchanged (`orchestration/talent_remap_70753.md`: druid, nothing to remap) |
| Genesis 2/4/6/8/10% periodic damage and healing | hotfix 113283 | in, on the HoTs and Swiftmend |

Shared lines that reach the healer: Human Spirit is 10% (`sim/core`, no druid race), and the High Order Skyborne's
Read Leyline is 50% regeneration for 15 sec, can't go up under Innervate, and an Innervate ends it (notes-only, in
`sim/core/racials_forever.go`). The rotation generator's `leyline` knob casts it only while Innervate is on
cooldown, so neither wastes the other.

### Healing Touch (Nature, 25297 at 60)

| Rank | Spell | Learned | Heal at 60 | Mana | Cast | Coefficient | Sim before |
|---|---|---|---|---|---|---|---|
| 11 | 25297 | 60 | 2139 to 2525 (2332, +7.3 a level to 65, ±8.3%) | 840 | 3.5 s | 1.0 | missing |
| 10 | 9889 | 56 | 1786 to 2107 (1946) | 755 | 3.5 s | 1.0 | missing |
| 9 | 9888 | 50 | 1415 to 1676 (1546) | 630 | 3.5 s | 1.0 | missing |
| 8 | 9758 | 44 | 1095 to 1303 (1199) | 520 | 3.5 s | 1.0 | missing |
| 7 to 1 | 8903 … 5185 | 38 … 1 | 921 … 47 | 425 … 25 | 3.5 … 1.5 s | 1.0 … 0.429 | missing |

Forever's Healing Touch is below 1.12's (rank 11 2267 to 2677). Downranking: 5% of the coefficient a level past
the rank's learn level plus 17, from the engine (`sim/core/downrank.go`, every ranked heal: user
ruling 2026-11-08, damage is out), so ranks 8 to 11 keep all of it at 60, rank 7 75%, rank 6 45%, rank 4 none. The HoTs take it
the same way on their ticks (Rejuvenation rank 6 55%, Regrowth rank 4 35%); `TestDruidDownrankedHeals`.

### Regrowth (Nature, 9858 at 60)

| Rank | Learned | Direct heal at 60 | HoT | Mana | Cast | Coefficients | Sim before |
|---|---|---|---|---|---|---|---|
| 9 | 60 | 965 to 1077 (1021 ±5.5%) | 142 every 3 s, 7 ticks (994) | 525 | 2 s | 0.286 direct, 0.071 a tick | missing |
| 8 | 54 | 800 to 894 | 113 × 7 | 445 | 2 s | same | missing |
| 7 | 48 | 633 to 710 | 88 × 7 | 370 | 2 s | same | missing |
| 6 to 1 | 42 … 12 | 527 … 96 | 68 … 13 × 7 | 305 … 70 | 2 s | same | missing |

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
| 3 | 50 | 201 | 5 | 695 | 5 min | 0.067 | missing |
| 2 | 40 | 133 | 5 | 505 | 5 min | 0.067 | missing |
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
| Takes the HoT | no: "Redesigned to not consume the Rejuvenation or Regrowth effect on the target" (news 24301515; development notes post 9, 2026-11-04); client 70755's tooltip still says "consumed", and the notes win (standards 37) | the HoT stays and ticks on, and still Blooms when it ends (gated under Forever) | consumed it (healer_mechanics-20261104-2) |

Gift of Nature and Genesis name Swiftmend in their class masks (Gift of Nature's first effect, Genesis's second, an
all-of-the-spell modifier), so the sim gives Swiftmend those and not the HoT's own (Improved Rejuvenation, the
periodic ones). That's a call below.

### Bloom (1328621, the October 30 hotfix's Restoration row 3 talent)

An expiring Rejuvenation or Regrowth heals its target for 100% of our Spirit (50% if dispelled, which doesn't
come up). The heal (1328622) has no class mask, so no healing talent reaches it. The sim heals at the HoT's last
tick, not when it's recast or ends with its target. Swiftmend no longer takes it, so a HoT Swiftmend read still Blooms. It can crit (no "can't crit" bit).

### Nature's Swiftness (17116)

The next Healing Touch, Regrowth, Wild Growth or Wrath (its class mask; Wild Growth since hotfix push 113885 of
2026-11-10) is instant. 3 min cooldown,
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
2. **Swiftmend's heal.** Default: every tick of the HoT it reads (which it no longer consumes) including Nature's Splendor's (5 Rejuvenation
   ticks, 9 Regrowth ticks), at the HoT's snapshot base and healing power, with Swiftmend's own modifiers (Gift of
   Nature, Genesis) in place of the HoT's. Alternatives: 1.12's fixed 12 sec of Rejuvenation or 18 of Regrowth (4
   and 6 ticks), and/or the HoT's own modifiers carried over (Improved Rejuvenation's 15% on a consumed
   Rejuvenation).
3. **Which HoT Swiftmend reads** with both up: the one with less time left (1.12 and later;
   `TestDruidSwiftmendPicksLeastTimeLeft`). Untested in Forever. Now that the HoT stays, the pick only sets the size of the heal.
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
8. **Downranking** is no longer a call: the engine applies it to every ranked heal (user ruling 2026-11-08,
   `sim/core/downrank.go`). What it moved on the presets is under "Downranking on the presets" below.
9. **The raid sim's Restoration druid placeholder** (`sim/compsearch/presets.go`) keeps its old talents; whether
   it takes the bench's build is the raid sim's (H6).
10. **Launching the spec page** (`ui/core/launched_sims.ts`) is the user's call; it stays unlaunched.
11. **Swiftmend keeps the HoT** (healer_mechanics-20261104-2). Default: the notes' "does not consume" (news
    24301515 and development notes post 9), though client 70755's tooltip still says "consumed"; standards 37 puts
    the notes first. If a later client or a log shows the HoT taken, `hot.Cancel` comes back under Forever.
12. **The tank preset's 85% line.** The best tank preset now casts Swiftmend and then Regrowth at the tank
    whenever it is under 85% health, so Regrowth is close to its main heal and Healing Touch fills behind it.
    That's what the bench rewards (+0.07 mean carry over the 40% line); a player would call it a Regrowth-heavy
    tank rotation rather than an emergency line. Default: ship it.

## The healer audit fixes (2026-11-04)

From the healer auditors' first pass (`orchestration/audit/findings/healer_*.md`), on master after the
orchestrator's engine fixes (raid damage share, school crit on direct heals, health clamp, the tank role's idle
fillers), which moved every number below:

- **healer_mechanics-20261104-2, fixed**: Swiftmend no longer consumes the Rejuvenation or Regrowth it reads
  under Forever (news 24301515; development notes post 9). The HoT ticks on and still Blooms. `TestDruidSwiftmend`
  now fails if the HoT is taken (every Rejuvenation must run its 5 ticks and tick after each Swiftmend).
  With the HoT surviving, both rotations still want Swiftmend where it was (the emergency line first, then
  Swiftmend at the tank when the hole covers it); a lower gate on the tank's Swiftmend line loses (0.6x -0.003,
  0.3x -0.023 mean carry), a higher one is level (1.3x -0.002 to +0.002); the raid healer's gains from 0.6x.
- **healer_mechanics-20261104-9, fixed**: the rank tables above had Healing Touch 8, Regrowth 7 and 8 and
  Tranquility 2 and 3 grown past their max level; corrected to the client's (70755). The Genesis comment in
  `sim/druid/talents.go` and the mechanics verdict that said the sim has no Swiftmend are corrected.
- **healer_mechanics-20261104-10, druid part fixed**: new tests `sim/druid/healing_ranks_test.go` (every rank of
  every heal against client 70755: ranges at 60, ticks, mana, cast times, coefficients) and
  `sim/raid_healing_druid_more_test.go` (Swiftmend on Regrowth with Nature's Splendor's 9 ticks; the least-time-left
  pick both ways; Gift of Nature and Genesis on Swiftmend; the Tier 1 set's 2, 4 and 5 pieces; Gift of Nature on
  Tranquility and Wild Growth, Genesis on Wild Growth and not Tranquility; Nature's Splendor's 2 Regrowth ticks;
  Rejuvenation and Tranquility tick crits and Improved Regrowth on Regrowth's ticks; Nature's Swiftness on
  Regrowth, and on Wild Growth since hotfix 113885; Moonglow leaving heal costs alone; Bloom on Regrowth). Still untested from that
  list: Improved Tranquility's threat cut, Innervate's exact regen, Bloom crit.
- **healer_fairness-20261104-2, fixed for the druid**: the tank preset's Healing Touch gate goes from 0.9x to 0.8x
  of the heal's Expected Healing (0.6x, 0.7x, 0.75x, 0.85x and 1.0x all lower), its Regrowth gate from 1.0x to
  0.7x, and its emergency line (Swiftmend, then Regrowth, at the tank) from 40% to 85% health. The line was the
  big lever once the HoT survives Swiftmend: 40% 1.087, 50% 1.097, 60% 1.108, 80% 1.130, 85% 1.140, 90% 1.132,
  95% 1.116 (with the 0.8x / 0.7x gates). The raid preset: emergency line 50% → 60% and Swiftmend's gate 1.0x →
  0.6x at the most hurt (70% loses on Vaelastrasz).
- **Nature's Swiftness on Wild Growth (healer_rotations-20261111-1, healer_fairness-20261111-1)**: since hotfix
  113885 Nature's Swiftness also makes Wild Growth instant, so the raid preset's autocast NS (raised at 0 s) went
  on its first Wild Growth instead of the emergency Regrowth (mean raid carry 1.437 -> 1.391). Both presets now
  cast NS on their own line (generator knobs `ns_on` = rg, ht or wg, the heal it is spent on, and `ns_at`, the
  lowest health in the assignment that raises it; line 1's autocast no longer takes it). Healbench, own lengths,
  paired seeds, 300 iterations a screen (raid mean carry): NS on the emergency Regrowth at 30 / 35 / 40 / 45 / 50 /
  60% 1.453 / 1.465 / 1.465 / 1.461 / 1.472 / 1.448, on Healing Touch 11 at 35 / 40 / 45 / 60% 1.457 / 1.456 /
  1.445 / 1.409, on Wild Growth at 45 / 60 / 80% 1.419 / 1.418 / 1.406, the autocast 1.409. At 1,000: Regrowth at
  45% 1.438, 50% 1.437 (60%, the audit's variant, 1.439); at 180 / 300 s (Onyxia, Magmadar, Ragnaros) 45% beats
  60%: 1.241 / 1.158 against 1.215 / 1.148 (autocast 1.193 / 1.096). Tank (300 screen, mean): an instant Healing
  Touch 11 at 50 / 60 / 70 / 80% 1.275 / 1.282 / 1.281 / 1.253, Regrowth at 50 / 85% 1.270 / 1.247, Wild Growth at
  50% 1.223, the autocast 1.230; at 1,000 HT at 60% 1.272 and 70% 1.265 against 1.235, and at 180 / 300 s (Onyxia,
  Golemagg, Chromaggus) 1.264 / 1.016 against 1.252 / 1.008. Shipped: Regrowth at 45% (raid), Healing Touch 11 at
  60% (tank). Talents: Subtlety's 2 points (threat only, which the bench doesn't score) go to Heart of the Wild
  (4% Intellect): raid level at own lengths (1.438 both), +0.010 / +0.009 at 180 / 300 s with 160-240 more mana
  left; tank 1.273 against 1.272. Heart of the Wild 5 for Moonglow, Nature's Splendor and Subtlety (Genesis 5,
  Nature's Majesty 2, Improved Tranquility 1) loses on the screen (raid 1.453 against 1.464, tank 1.267 against
  1.282), as does Heart of the Wild 3 for Subtlety and Nature's Swiftness (raid 1.390).

Healbench, 1,000 iterations, all seven boss files, default carry and fight length, built from this branch
(`healers/fix-druid`) on master 0606686d3, carry:

| Boss file | Tank, old preset | Tank, new | Raid, old preset | Raid, new |
|---|---|---|---|---|
| Magmadar | 1.731 | 1.867 | 1.922 | 2.002 |
| Golemagg | 1.644 | 1.812 | 2.660 | 2.660 |
| Ragnaros | 0.534 | 0.574 | 1.289 | 1.270 |
| Onyxia | 1.183 | 1.195 | 0.984 | 0.997 |
| Broodlord | 0.756 | 0.793 | 0.909 | 0.920 |
| Vaelastrasz | 0.742 | 0.793 | 1.099 | 1.102 |
| Chromaggus | 0.893 | 0.946 | 1.113 | 1.109 |
| Mean | 1.069 | 1.140 | 1.425 | 1.437 |

"Old preset" is the shipped rotation on this branch's sim (Swiftmend no longer consuming); both sides leave
5,900 to 6,300 mana at the end on Onyxia, so neither is mana bound. The tables below are from before these fixes
and the engine's, and are kept for the record.

## Tuning and the presets

Consumables (2026-11-12): Greater Cleric's Elixir for the Greater Arcane Elixir, the Wicked Smoothie, the Wizard Oil (the Brilliant Mana Oil is a Zandalar formula), Flask of Supreme Power, in both roles; the rest as every healer (`docs/healing_sim.md`, Consumables).

`orchestration/tools/healers/druid_restoration_apl.py` writes the two rotations from knobs and
`druid_restoration_req.py` the bench requests from the shaman's (the same raid, buffs and consumes): Thicc Resto
`05302001--5253005153113051` (since the Nature's Swiftness round below; `05302001--5053025153113051` before),
Windshaper Skyborne, the synthetic set. Carry on `tools/healbench`, screened at 400 to
500 iterations, the finalists and the table at 1,000, on master with client 70753 and the November 2 notes.

- **Tank healer** (`apls/tank_healer.apl.json`, `req/druid_restoration.json`: now `ht=11,8,6 ht_scale=0.8 rg_scale=0.7
  low=0.85 ns_on=ht ns_at=0.6`, see above and the Nature's Swiftness round; the round below shipped `ht=11,8,6 low=0.4 ht_scale=0.9`):
  - Healing Touch bound at ranks 11, 8 and 6 beats 11, 10, 9 on every file (Onyxia 1.165 against 1.094,
    Vaelastrasz 0.695 against 0.617, 500 iterations) and 11, 8, 7 or 11, 8, 6, 5 (500: 1.159 and 1.147 on Onyxia).
  - The emergency line at 40% and Healing Touch at 90% of its heal against the 35% line at the full heal, at
    1,000: Golemagg 1.648 / 1.618, Onyxia at 180 / 300 sec 1.087 / 1.065 and 1.063 / 1.046, Chromaggus 0.870 /
    0.867, Onyxia 1.165 / 1.164; Ragnaros 0.517 / 0.526, Broodlord 0.731 / 0.740 and Vaelastrasz 0.688 / 0.693 the
    other way. Shipped for the long fights.
  - Lost or level: Nature's Swiftness at 25% (+0.02 Broodlord, -0.015 Vaelastrasz), Regrowth at half its heal,
    Swiftmend at half, the Innervate threshold (6,000 mana left at the end: the tank side isn't mana bound).
  - Talents (500): Heart of the Wild for Nature's Focus level (1.177 / 1.172 on Onyxia, 1.095 / 1.087 at 180
    sec), all 51 in Restoration -0.08 on Onyxia, 46 with Genesis -0.06. Races: Night Elf and Tauren within 0.01
    of the Windshaper, the High Order the same as it (its Read Leyline never goes up, since Innervate is never
    on cooldown at 60% mana; the generator's `leyline` knob is there for a build that is mana bound).
- **Raid healer** (`apls/raid_healer.apl.json`, `req/druid_restoration_raid.json`: now `rj_scale=0.2 low=0.6
  sm_scale=0.6 ns_on=rg ns_at=0.45`, see above and the Nature's Swiftness round; the round below shipped `rj_scale=0.2 low=0.5`):
  - Rejuvenation at the most hurt once it misses 20% of the HoT, with a 50% emergency line, against 30% and 35%,
    at 1,000: ahead on every file (Onyxia 0.711 / 0.679, Chromaggus 0.953 / 0.887, Ragnaros 1.240 / 1.181,
    Magmadar 1.605 / 1.456, Golemagg 2.021 / 1.877).
  - Lost or mixed (500): Rejuvenation at 10% (Ragnaros +0.03, Broodlord and Chromaggus -0.01), a second rank of
    Rejuvenation (Vaelastrasz +0.07, Chromaggus -0.05), Wild Growth at half or 1.5 times its HoT, no Regrowth,
    Tranquility at 40%, Nature's Swiftness; Heart of the Wild or all 51 in Restoration lose up to 0.05.
  - It overheals about 60% (HPS 950, effective 370 on Onyxia): Wild Growth's other four targets and HoTs on
    players the next heal tops up. It still isn't mana bound (3,700 to 6,000 left).

### Against the Restoration shaman (1,000 iterations, carry)

On master 43a70ec18, the re-bench after the 2026-11-04 audit fixes and the tank bench's change of that day
(tanks face the boss), each healer's request from `tools/healbench/req`, one healbench for both (the 7 files'
rows are healing_sim.md's H3 table):

| Boss file | Tank: druid | Tank: shaman | Raid: druid | Raid: shaman |
|---|---|---|---|---|
| Magmadar | 2.030 | 2.019 | 2.002 | 1.809 |
| Golemagg | 1.904 | 1.915 | 2.660 | 2.662 |
| Ragnaros | 0.627 | 0.577 | 1.270 | 1.235 |
| Onyxia | 1.327 | 1.346 | 0.997 | 0.919 |
| Broodlord | 0.875 | 0.974 | 0.920 | 0.902 |
| Vaelastrasz | 0.852 | 0.894 | 1.102 | 1.152 |
| Chromaggus | 1.029 | 0.969 | 1.109 | 1.090 |
| Mean, 7 files | 1.235 | 1.242 | 1.437 | 1.396 |
| Onyxia, 180 sec | 1.228 | 1.138 | 0.865 | 0.825 |
| Onyxia, 300 sec | 1.017 | 0.980 | 0.785 | 0.738 |

The druid raid healer is ahead on every file but Vaelastrasz (-0.050) and Golemagg (-0.002, level), by up to
0.193 (Magmadar), and on both long Onyxia rows. The tank healers are level on the mean (-0.007): the shaman is
ahead on Broodlord (+0.099), Vaelastrasz (+0.042) and Onyxia (+0.019), the druid on Chromaggus (+0.060), Ragnaros
(+0.050) and the long fights (+0.090 and +0.037 at 180 and 300 sec). The druid's own tank numbers above (1.140)
are from before the tank bench's change. On master 41aa10700, before the fixes, the shaman's new Riptide had put
it ahead on the short tank fights (Broodlord +0.106, Magmadar +0.099); before that change (master 3d1cb9701) the
druid was ahead on every tank row and behind on four raid rows.

### Downranking on the presets

The engine applies Forever's downranking to the druid's heals now (8f304b240, b57f5cc0a). This branch had the same
rule in the druid package before (so the tuning above already had it); measured against the heals at their full
coefficients (the scope without heals, `ForeverDownrankScope = DownrankDamage`, the round 3 rotations), 1,000
iterations:

| Boss file | Tank, with | Tank, without | Raid, with | Raid, without |
|---|---|---|---|---|
| Magmadar | 1.699 | 1.684 | 1.456 | 1.456 |
| Golemagg | 1.618 | 1.613 | 1.877 | 1.877 |
| Ragnaros | 0.526 | 0.504 | 1.181 | 1.181 |
| Onyxia | 1.164 | 1.089 | 0.679 | 0.679 |
| Broodlord | 0.740 | 0.715 | 0.815 | 0.815 |
| Vaelastrasz | 0.693 | 0.652 | 1.086 | 1.086 |
| Chromaggus | 0.867 | 0.837 | 0.887 | 0.887 |
| Onyxia, 180 / 300 sec | 1.065 / 1.046 | 1.033 / 1.008 | | |

- The raid healer casts only ranks learned at 50 or later (Rejuvenation 11, Regrowth 9, Healing Touch 11 and 9,
  Wild Growth 3), which keep their whole coefficient, so nothing moves.
- The tank healer's Healing Touch 6 (learned at 32) keeps 45% of its coefficient. Without the penalty it heals
  about 650 more, close to rank 8, so the ladder of 11, 8 and 6 loses its small step: the rank 6 line fires
  less, rank 11 more, and the carry drops by 0.005 to 0.075. A rotation tuned without the penalty would bind other
  ranks; this one was tuned with it, so the "without" column is a lower bound on what the rule costs.

### The shared change

`sim/core/apl_helpers.go` (`GetAPLDot`): a dot or HoT check aimed at Most Hurt or Lowest Health finds the HoT
through `HealthUnit()` instead of `Get()`, which is nil while the rotation is built (nobody is hurt yet). For a
spell whose HoT is a `Hot` (no `Dot`), `CurDot()` was nil too, so the check was dropped from the condition
(`TestDruidHotAtMostHurt`). The Restoration shaman's raid rotation has one such check (Riptide 1239243 not active
at the most hurt). Its raid request at 1,000 iterations on all seven boss files, healbench built from this branch
with and without the change: the same to the third decimal on every file (Magmadar 1.540, Golemagg 1.913,
Ragnaros 1.171, Onyxia 0.678, Broodlord 0.817, Vaelastrasz 1.110, Chromaggus 0.906; mean 1.162 both ways). The
same held on the shaman's rotation before 41aa10700.
