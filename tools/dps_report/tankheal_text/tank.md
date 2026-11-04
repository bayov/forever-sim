<!--
The Tanking mode's player-facing text, read by tankheal.py build. Written for players, not converted from the docs:
docs/tank_*.md are the facts behind it, for developers. Short sentences, everyday words, normal WoW terms only. No
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

## warrior__protection_max_threat
The most threat a Protection Warrior makes. It moves a few talent points from defense to threat, skips Thunder Clap
and spends Rage on Heroic Strike sooner. Your DPS can push harder, but you take more damage and need more healing.

## warrior__protection_balanced
The all-round Protection Warrior: Deflection, 31 points in Protection with Shield Slam, Shield Block whenever it's
down and Thunder Clap kept up on the boss. Strong threat and the least damage taken.

## warrior__protection_max_survival
Built on Balanced's talents and rotation. Nothing we tried cut the healing the warrior needs by enough to be worth the
threat it costs, so this plays almost exactly like Balanced.

## paladin__protection_max_threat
More threat for a Protection Paladin: Improved Seals in place of Anticipation and a point of Toughness. The rotation
is the same as Balanced.

## paladin__protection_balanced
Seal of Fury and Righteous Fury, Holy Shield on cooldown (and once before the pull), Judgement and Holy Strike on
cooldown, and Consecration while you have mana to spare. Templar's Bulwark on cooldown.

## paladin__protection_max_survival
Balanced's talents and rotation. Nothing we tried cut the healing needed by enough to be worth the threat it costs.

## druid__bear_max_threat
Balanced, but Maul from 20 Rage instead of 30. It plays and scores almost the same as Balanced.

## druid__bear_balanced
Lacerate to 5 stacks, Primal Bite on cooldown, Maul from 30 Rage, Demoralizing Roar kept up and Barkskin on cooldown.
No Furor: shift into Bear Form just before the pull instead.

## druid__bear_max_survival
Balanced plus Frenzied Regeneration at 100 Rage, and Enrage a bit earlier before the pull so its armor loss ends
sooner. A little less threat for less healing needed.

## shaman__warden_max_threat
Elemental Weapons in place of Shamanistic Focus, the attack power trinket and Mana Spring Totem. A bit more threat,
but you run low on mana sooner.

## shaman__warden_balanced
Tidal Focus for hit and Improved Reincarnation for 4% more health. Stoneskin Totem and Lightning Shield kept up,
Stormstrike and Earth Shock on cooldown, the attack power trinket.

## shaman__warden_max_survival
Balanced's talents and rotation with the armor on-use trinket in place of the attack power one.

# Notes

## Warrior, Paladin: Why Undead
Undead comes out on top because of Touch of the Grave: your attacks sometimes drain the boss and heal you. Blizzard
only says it drains "up to" 5% of your health, and we count the full amount. If it heals for less in game, the gap
to the other races shrinks, and another race may be as good.

## Warrior: Low-health buttons
The sim doesn't wait for you to get low. Shield Wall is used near the end of the fight, and Last Stand and health
potions aren't used at all. In game, use them when you need them: that's on top of what's shown here.

## Paladin: No bubbles
Divine Shield, Divine Protection and Lay on Hands aren't used: a tank under a bubble loses the boss to someone
else. Templar's Bulwark is used on cooldown instead, and it shares Forbearance with them.

## Druid: Bear Form threat
Bear Form's threat bonus follows Blizzard's patch notes (+115%), not the older tooltip (+50%). If the notes are
wrong in game, bears make a lot less threat than shown here.

## Druid: Furor
Furor's Rage is wasted when you shift into Bear Form before the pull, so none of the setups take it.

## druid__bear_max_threat, druid__bear_balanced: Max threat and Balanced
These two are the same setup in practice: they only differ in when Maul is queued.

## Shaman: Earth Shock threat
Earth Shock counts as twice its damage in threat, as in Classic. Forever's tooltip only says it causes "a high amount
of threat". If it's less in game, the Warden loses most of its lead in Safe DPS.

## Shaman: More damage taken
The Warden has no plate and no Shield Block, so it takes more damage than the other tanks and needs the most
healing. Its threat comes early instead: Earth Shock with a full mana bar, Stormstrike and totems before the pull.

# Bosses

## magmadar
Fast swings, with his Frenzy counted as often as real raids let it through. Lava Breath is in. His fear, Magma Spit
and Lava Bomb are left out.

## golemagg
Golemagg only: his two Core Ragers are another tank's job. Magma Splash stacks Fire damage on you (a steady 4 stacks
here); its armor loss is left out, and so are Earthquake and Pyroblast.

## onyxia
The ground phases only: her air phase is left out. Knock Away's threat loss and Bellowing Roar's fear aren't in.

## broodlord
The hardest boss to keep a tank alive on: big hits, and Mortal Strike halves your healing for 5 seconds. Blast Wave
is in. Knock Away's threat loss isn't.

## vaelastrasz
One tank's stretch before Burning Adrenaline forces a swap, so a short fight. Flame Breath's burn stacks up to 5
times, and fire resistance helps.

## chromaggus
Incinerate and Corrosive Acid as his breaths, and his Frenzy counted as often as raids let it through. Corrosive
Acid's armor loss is left out.

## ragnaros
Shown, not scored. Elemental Fire burns you on most of his hits, and it's one tank's time between his knockbacks.
Real raids tank him in fire resistance gear, which these setups don't wear, so every tank needs far more healing here
than raids give.

# Method

## How tanks are scored
A tank has two jobs: hold aggro so your DPS can go all out, and stay alive. We sim each tank on six raid bosses and
measure both:

- **Safe DPS**: how much DPS a damage dealer can do behind this tank without pulling aggro.
- **Healing needed**: how much healing per second the tank needs to stay alive.

The score turns both into one number, in raid DPS. More threat lets the raid do more damage. Needing more healing
costs the raid: each extra healer on the tank takes a slot that could hold a DPS player.

The three setups weigh these differently. **Max threat** cares most about aggro, **Max survival** most about needing
less healing, and **Balanced** weighs them evenly. A score can go below zero, mostly at Max
survival, where healing counts three times as much.

## Safe DPS
Aggro trouble mostly comes in the first seconds of a pull, before the tank has built a lead. So Safe DPS isn't the
tank's average threat: it's the steady DPS a player can do, starting 2 seconds after the pull, without ever going
over the tank's threat, even at the worst moment of the fight. You pull aggro at 110% of the tank's threat, as in
melee range. Forever has no Blessing of Salvation, so every point of your damage is a point of threat.

## Healing needed
On each boss the tank gets a steady stream of heals, as often as real raids heal their main tank. We turn that
healing down until the tank dies in 1 fight out of 20. That amount, in HPS, is the healing it needs.

It counts everything that keeps a tank alive: armor, avoidance, health and blocks. Spiky damage counts too: a tank
that takes big hits needs more healing than one that takes the same damage in small ones.

Sometimes no amount of healing is enough, when the boss kills the tank between two heals too often. Then healing
needed shows **none** and the score drops far below the rest.

## The bosses
Six bosses count toward the score, with the damage and healing real 40-player raids had on them (Warcraft Logs of
the Anniversary realms). Forever's own raids have no logs yet, so these Classic bosses stand in. Each fight is the
time the main tank spends on the boss.

{bosses}

## What the sim leaves out
- **Threat drops**, like Onyxia's and Broodlord's Knock Away.
- **Buttons for when you're low**: Last Stand, Shield Wall when you're low, health potions, Lay on Hands. Defensive
  cooldowns are used on cooldown or at a set time instead.
- **Healers reacting to you**: the heals come at a steady pace whatever your health.
- **Adds and tank swaps**: one tank on the boss.

## The exact math
On each boss:

**Score = 2 × Safe DPS − weight × 900 × healing needed ÷ one healer's HPS**

- **2 × Safe DPS**: the two DPS players closest to the tank's threat can both go that fast.
- **Healing needed ÷ one healer's HPS**: how many healers the tank needs. One healer's HPS is what each healer on
  the main tank healed in real kills of that boss, from about 170 to about 710.
- **900**: the DPS a player would add in that healer's slot.
- **Weight**: ⅓ for Max threat, 1 for Balanced, 3 for Max survival.

The score is the average over the six bosses. The ± next to a number is how far it could move by luck: two setups
closer than that are about even.

# Rotation

from: \s*\(on a timer: a scoring run reads no health\)
to:

from: ,? on a timer \(a scoring run reads no health\)
to:

from: : the tank applies its own \(the bench drops the raid's Roar and Shout\)
to: . Only your own counts here, not another player's Roar or Shout

from: ^The defensive on-use trinket on cooldown, on a timer\.
to: The defensive on-use trinket on cooldown.

from: \bour dodges and parries\b
to: your dodges and parries

from: : the tank applies its own slow\.
to: : apply your own slow.
