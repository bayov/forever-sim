<!--
The Healing mode's player-facing text, read by tankheal.py build. Written for players, not converted from the docs:
docs/healing_*.md are the facts behind it, for developers. Short sentences, everyday words, normal WoW terms only. No
tool, file or process names: the build's leak check fails on them (tankheal.py LEAK).

Sections (# level), each a list of ## entries:
  Presets   ## <preset id>: one short paragraph, the banner line of its Specs page (what it's for, what it trades)
  Notes     ## <Class or preset id, comma separated>: <title>, then Markdown; shown under "Good to know" on the
            Specs page of every preset it names (a class: all of its presets)
  Bosses    ## <boss file name>: a line or two on what the fight has and what it leaves out
  Method    ## <card title>, then Markdown; a line {bosses} becomes the list of bosses from the Bosses section
  Rotation  from: <regex> / to: <replacement> pairs, applied in order to the rotation's notes (the requests' own)
Every preset and boss needs its entry: the build stops when one is missing.
-->

# Presets

## paladin__holy
Holy Light at the tank, with Flash of Light between casts at whoever is most hurt. Oath of Light kept on the tank,
Divine Favor before a big Holy Light, Holy Shock and Flash of Light when someone drops low. Holy Light only while
you're above half mana, so you don't run dry in a long fight.

## paladin__holy_raid
The tank healer's setup without Oath of Light, every heal going to whoever is most hurt.

## druid__restoration
Rejuvenation kept on the tank, Healing Touch in the rank that fits the damage, Swiftmend and Regrowth when the tank
drops low, and Rejuvenation on the rest of the raid when it fits.

## druid__restoration_raid
Wild Growth and Rejuvenation on whoever is hurt, Swiftmend and Regrowth for anyone who drops low, and Healing Touch
for big gaps.

## shaman__restoration
Riptide kept on the tank, Lesser Healing Wave while you're above half mana and Healing Wave below it, Nature's
Swiftness when the tank drops low. Chain Heal on the raid when the tank is topped. Water Shield, Mana Spring and Mana
Tide Totem for mana.

## shaman__restoration_raid
Riptide on whoever is hurt and Chain Heal as the main heal, Healing Wave for one big gap. Lesser Healing Wave when
someone drops low.

## priest__discipline
Power Word: Shield on the tank whenever you can, Penance, Renew, and Flash Heal or Greater Heal in the rank that fits
the damage. Flash Heal first while you have lots of mana. Power Infusion and Inner Focus for the big heals.

## priest__holy
Renew and Prayer of Mending kept on the tank, Flash Heal or Greater Heal in the rank that fits the damage, Flash Heal
first while you have lots of mana. Power Infusion and Inner Focus for the big heals.

## priest__holy_raid
Prayer of Mending, Renew and Power Word: Shield on whoever is hurt, Flash Heal or Greater Heal in the rank that fits,
and Prayer of Healing when a group is hurt.

# Notes

## Paladin: Why Human
Human is the best race here, a little ahead of Dwarf and Undead: The Human Spirit's extra Spirit keeps your mana up
in long fights.

## Paladin: Free talent points
Every Holy talent that helps healing is taken, so 7 points sit in Retribution talents that don't heal. Spend them as
you like.

## Druid, Shaman: Why Windshaper
Windshaper is the best race for both. Its edge over the other races is small.

## Druid: Swiftmend keeps your HoT
On Forever, Swiftmend doesn't use up the Rejuvenation or Regrowth it eats. That's why Swiftmend comes up so often in
these rotations.

## druid__restoration_raid: Overhealing
The raid druid overheals a lot, about 60% of its healing, as HoTs land on people who are already full. It doesn't
run out of mana, so it doesn't matter much.

## Shaman: Riptide
Riptide here only puts the heal over time on the target, not the instant heal: Forever's patch notes don't mention
one. If Riptide does heal up front in game, the raid shaman would score a bit higher (about 0.07).

## Priest: What's left out
Shadowfiend and Lightwell aren't used.

## priest__discipline, priest__holy: Holy or Discipline
On the tank, Holy and Discipline heal about the same. Discipline runs tighter on mana in long fights.

## priest__holy_raid: Prayer of Healing
Prayer of Healing is weak here: these groups rarely have five hurt players at once. Most of the healing is Flash
Heal, Greater Heal and Prayer of Mending.

# Bosses

## magmadar
Fast swings on the tank, with his Frenzy counted as often as real raids let it through, and Lava Breath. His fear,
Magma Spit and Lava Bomb are left out.

## golemagg
Golemagg on the tank, with Magma Splash stacking Fire damage on them. The Core Ragers, Earthquake and Pyroblast are
left out.

## onyxia
The ground phases only: the air phase is left out. The raid healer's fight is the whole kill.

## broodlord
Big hits on the tank, and Mortal Strike halves the tank's healing for 5 seconds. Blast Wave hits the melee too.

## vaelastrasz
The most damage of any boss here, in a short fight: Flame Breath stacking on the tank and fire on the whole raid.
The tank's fight is one tank's stretch before Burning Adrenaline forces a swap.

## chromaggus
Incinerate and Corrosive Acid as his breaths and his Frenzy on the tank, and the Brood Afflictions on the raid. The
longest fight here, so mana matters most.

## ragnaros
Elemental Fire burns the tank on most of his hits, and the tank's fight is one tank's time between his knockbacks.
Real raids tank him in fire resistance gear, which this tank doesn't wear, so he's the hardest boss here.

# Method

## How healers are scored
We turn the boss's damage up until the healer can't keep everyone alive any more. The score is how far it got:
**1.20** means the healer still kept up with 20% more damage than real raids took on that boss, and **0.80** means
it fell behind at 80% of it. Higher is better.

This counts everything that matters in a fight: how much you heal, how well your heals land where they're needed,
and whether your mana lasts.

## Tank healer and raid healer
Healers do two different jobs, so they're ranked apart.

- **Tank healer**: one healer alone on a Protection Warrior tank, for the time a main tank spends on the boss. It
  fails when the tank dies in more than 1 fight out of 50.
- **Raid healer**: two of the same healer on a group of seven mages, through the boss's damage to the raid. It fails
  when anyone dies in more than 1 fight out of 20. The fight is the length of a real kill.

## The bosses
Seven bosses, all scored, with the damage real 40-player raids took on them (Warcraft Logs of the Anniversary realms).
Forever's own raids have no logs yet, so these Classic bosses stand in.

{bosses}

## How the sim heals
- Healers react to damage in about 0.3 seconds, a bit faster or slower each time, like a player.
- A heal that would mostly overheal (more than half) is cancelled before it lands, as a good healer would.
- A healer picks the biggest rank of a heal that fits the missing health, so it doesn't waste mana.
- The tables per boss show HPS, overhealing and mana at the boss's normal damage, not at the turned-up damage.

## The exact math
On each boss, the score is the biggest multiple of the boss's damage at which:

- **Tank healer**: the tank dies in no more than 2% of fights.
- **Raid healer**: nobody dies in more than 5% of fights.

A healer's score is the average over the seven bosses. The ± next to a number is how far it could move by luck: two
healers closer than that are about even.

# Rotation

from: \bwhoever in my assignment is
to: anyone you're healing who is

from: ^Someone in my assignment is under (\d+)% health: the fast heal\.
to: Anyone you're healing under \1% health: the fast heal.

from: \bhas my HoT\b
to: has your HoT

from: all my mana
to: all your mana

from: \bon ourselves\b
to: on yourself

from: \bon myself\b
to: on yourself

from: ,? the biggest (?:that|whose heal \d+% of) what (?:it|they|is) miss(?:es|ing)? covers(?: \(\d+% of the heal\))?
to: , the biggest rank that fits what's missing

from: : each rank, biggest first, once what is missing covers (?:\d+% of )?its Expected Healing
to: : the biggest rank that fits what's missing

from: once what is missing covers \d+% of its Expected Healing
to: once enough health is missing

from: when what it misses covers (?:100% of )?(?:the heal|it)
to: when enough health is missing

from: and what it misses covers the heal
to: and enough health is missing

from: only once what is missing covers 100% of a Greater Heal 5
to: only once the tank is missing a full Greater Heal

from: ^Penance once what is missing covers its 3 bolts\.
to: Penance once the tank is missing enough health for all 3 bolts.

from: , whatever it misses
to:

from: \(below 50% mana first\)
to: (first once you're below 50% mana)

from: \bmisses 20% of the HoT\b
to: is missing at least a fifth of the HoT

from: is under (\d+)%:
to: is under \1% health:
