# Leveling warrior

What the leveling session (`docs/leveling_sim.md`) models of a warrior from level 1 to 60. Values
are from client 71091 unless a line says otherwise.

## Abilities

A rotation can name any rank of an ability. The sim casts the highest rank known at the level. A
spell the level hasn't learned doesn't exist, so a rotation line for it never fires. The ranks of the
Ahn'Qiraj books (Heroic Strike, Battle Shout and Revenge at 60) aren't in Forever.

Levels are the levels the ranks are learned at. D is Defensive Stance, B Battle Stance, Z Berserker
Stance. The cap is how many mobs an area ability hits. In a session it only hits mobs we pulled that
stand within its radius.

| Ability | Spell id | Levels | Stance | Rage, cooldown | Notes |
| --- | --- | --- | --- | --- | --- |
| Battle Stance, Defensive Stance, Berserker Stance | 2457, 71, 2458 | 1, 10, 30 | | | Swaps keep 10 rage from level 14 (Tactical Mastery), 5 more a point of Improved Tactical Mastery. A warrior starts in the stance its options ask for, or in Battle Stance before it's learned. |
| Heroic Strike | 78, tag 1 | 1, 8, 16, 24, 32, 40, 48, 56 | any | 15 | On the next swing. In a session only tag 1 (the queue) works. |
| Battle Shout | 6673 | 1, 12, 22, 32, 42, 52 | any | 10 | |
| Charge | 100 | 4, 26, 46 | B (D with Vanguard) | 15 sec | 8 to 25 yards, out of combat only. Gives 9, 12 or 15 rage, 3 more a point of Improved Charge, and stuns for 1 sec. Aim it at the pull target. |
| Rend | 772 | 4, 10, 20, 30, 40, 50, 60 | B, D | 10 | One on each mob works. |
| Thunder Clap | 6343 | 6, 18, 28, 38, 48, 58 | B, D | 20, 6 sec | Cap 4, 8 yards. |
| Hamstring | 1715 | 8, 32, 54 | B, Z | 10 | |
| Bloodrage | 2687 | 10 | any | 1 min | 20% of base health. Usable out of combat. |
| Sunder Armor | 7386 | 10, 22, 34, 46, 58 | any | 15 | Every rank has its own armor value. |
| Overpower | 7384 | 12, 28, 44, 60 | B | 5, 5 sec | After a dodge. |
| Demoralizing Shout | 1160 | 14, 24, 34, 44, 54 | any | 10 | No cap, 10 yards. |
| Revenge | 6572 | 14, 24, 34, 44, 54 | D | 5, 5 sec | After a dodge, parry or block. |
| Shield Block | 2565 | 16 | D, shield | 10 | Tank wrapper only. |
| Cleave | 845, tag 1 | 20, 30, 40, 50, 60 | any | 20 | Cap 2, 5 yards. On the next swing. |
| Slam | 1464 | 20, 30, 38, 46, 54 | any | 15, 15 sec | |
| Retaliation | 20230 | 20 | B | 15 min | 15 sec, 30 counterattacks. Session and raid. |
| Victory Rush | 402927 | 20 | any | free, 30 sec | For 20 sec after a mob that gives experience dies. 1 + 15% of attack power, heals 10% of max health. Session only. |
| Execute | 5308 | 24, 32, 40, 48, 56 | B, Z | 15 | Reads the target's own health in a session. So does the execute phase in a rotation. |
| Shield Wall | 871 | 28 | D, shield | | Tank wrapper only. |
| Berserker Rage | 18499 | 30 | Z | | |
| Intercept | 20252 | 30, 42, 52 | Z | 10 | 8 to 25 yards, in combat too. 3 sec stun. Session only. |
| Whirlwind | 1680 | 36 | Z | 25 | Cap 4, 8 yards. |
| Pummel | 6552 | 38, 58 | Z | 10 | |
| Recklessness | 1719 | 50 | Z | | |
| Mortal Strike, Bloodthirst, Shield Slam | 12294, 23881, 23922 | 40, 48, 54, 60 | any | 30, 30, 20 | Talents. Shield Slam needs a shield. |
| Sweeping Strikes | 12292 | talent | B | 30 | 5 charges, one more target in reach each. |
| Spearing Strike | 1310222 | talent | B | 15, 20 sec | |
| Death Wish | 12328 | talent | any | 10 | |
| Reckless Swings | 1329555 | talent | any | 15 sec | 25% of base health. |
| Last Stand | 12975 | talent | any | | |
| Concussion Blow | 12809 | talent | any | 10, 45 sec | 5 sec stun, off the global cooldown, no damage. Session only. |
| Cannibalize | 20577 | Undead racial | any | 2 min | 35% of max health over 10 sec. Humanoid or Undead mobs. Out of combat, at the last corpse. Session only. |

Talents that only matter alone: Gore Drinker heals 0.5 or 1% of max health for each of the 3 melee
hits after a trigger. Lingering Rage delays the rage decay by 2 sec a point, and Anger Management
slows it by 30% and gives its rage only in combat. Vanguard lets Charge work in Defensive Stance.

Gore Drinker, Victory Rush, Shield Specialization and Master of Defense match our logs (checked on
2026-11-25). Each target a Whirlwind or Cleave lands on uses one of Gore Drinker's 3 hits, and Thunder
Clap uses none. Victory Rush heals on a miss too. Shield Specialization 5/5 gives 5 Rage on every block
and Master of Defense 2/2 on every dodge and parry, with no cooldown. One small gap is left: Victory
Rush's heal crits about 5% of the time for half more, and the sim's doesn't.

A stunned mob can't dodge, parry or block, and doesn't swing. Its swing timer goes on, so a swing
that came due lands when the stun ends.

### Surviving a big pull

The Recovery enchant, Thrill of Adventure, the Krol'dok set, shield spikes and health per 5 sec gear
are in `docs/leveling_sim.md`, "Healing while we fight". All of them are off by default.

We measured them on the Undead protection warrior with 600 sessions each, a pack of up to 12 and a
camp (2026-11-25). The base rotation is the one tuned without them, so these are low estimates. The
numbers are XP an hour.

| | Level 40 | Level 58 |
|---|---|---|
| None | 43,950 ± 90 | 50,650 ± 90 |
| Recovery | +840 | +1,900 |
| Mithril Spike at 40, Thorium Spike at 58 | +1,640 | +1,850 |
| Thrill of Adventure 5/5 | +950 | +1,830 |
| Krol'dok set | +880 (a test, since it needs 41) | +1,140 |
| 10 health per 5 sec | +550 | +500 |
| All five | +2,570 | +5,410 |
| All five, gathering 6 mobs | 48,560 against 38,160 without | 56,050 (the tuned rotation gathers 6) |
| Thorns | +1,810 | +2,120 |

With all five the best of 4, 6, 8, 10 and 12 mobs to gather is 6 at level 40, where it's 4 without
them. Gathering 10 or 12 still dies too often.

### Left out

- Disarm (676, 18), Intimidating Shout (5246, 22), Challenging Shout (1161, 26), Taunt (355),
  Mocking Blow (694, 16), Shield Bash (72, 12, 32, 52) and Piercing Howl (12323). A solo warrior
  holds every mob already, and the session has no casters and no fleeing mobs.
- Iron Will, Improved Hamstring, Improved Disarm and Improved Shield Bash.

## Rage

- A pull starts with the rage the last one left, less 1.25 a second out of combat. 247 gaps between
  rage readings in our logs drop off sharply above 1.3 a second, with the most common at 1.2 to 1.3.
- The rage bar empties when we die.
- A swing gives 3.46 times the weapon's speed for a one-hander and 4.5 times for a two-hander, at any
  level. A crit doubles it. Low-level warriors in our logs (13 of them, Nov 5 to 22) match: 9.3 from
  a 2.7 sec one-hander, 17.3 to a crit, and 15.2 from a 3.4 sec two-hander.
- Damage taken by mobs gives what the formula says. In our logs it is 1.00 of it in a
  stance that changes nothing. Defensive Stance's 10% less damage lowers the rage by as much, and Berserker's 10% more
  raises it (0.85 to 0.91 and 1.05 to 1.13 of the formula without the stance). So the rage follows
  the stance's modifier only, in a session as in the raid sim (`core.ForeverDamageTakenRageModsStance`,
  orchestration/rage_taken_by_stance.md).
- Bloodrage costs 20% and Reckless Swings 25% of base health. Base health is 20 at level 1, 101 at
  10, 199 at 20, 374 at 30, 649 at 40, 1079 at 50 and 1689 at 60.

Health regen out of combat is the session's, for every class the same way.

## Base stats

The sim has Undead Warrior base stats for all 60 levels, rising smoothly (level 1: 22 Strength, 18
Agility, 23 Stamina, 18 Intellect, 25 Spirit. Level 60: 119, 78, 111, 28, 50). Naked health is 70 at
level 1 and 89 at level 2, as our logs show.

## Leveling 1-60

Our plan for an Undead warrior is the "Leveling: Prot" build of the Protection Warrior sim
(`ui/tank_warrior`, since every leveling session plays that spec). The config
`tools/levelingplan/configs/undead_warrior_leveling_prot.json` makes the same request for
`tools/levelingplan`. Measured 2026-11-26.

The builds we compared it with are presets next to it: "Leveling: Arms then DW" (the config
`undead_warrior_leveling_1_60.json`), "Leveling: Arms", "Leveling: Dual Wield", "Leveling: Arms,
Prot, DW" and "Leveling: Fury Prot". Their configs are `undead_warrior_leveling_<build>.json` in
the same folder. Arms then DW had the talent and rotation search to 39. The others have hand-tuned
rotations. The Fury Prot plan judges shield gear by its own
weights ("Leveling Gear Plan Shield (Fury Prot)"), which the plan's `weights_build` picks.

What every number below assumes: solo grinding, mobs 1 level below us in a camp, pulls of up to 12,
1 point of Thrill of Adventure, Talented (the first talent at 5), the leveling gear plan's gear
judged by our measured weights for the weapon style, food and bandages, and no potions or outside
buffs. The hours compare builds. They aren't a forecast, since quests, travel and dungeons are left
out.

### The plan

- Protection all the way. A two-hander until 15, and a one-hander and shield from Shield Block at
  16. With the shield from 1 it takes 0.5h longer, all of it before 20.
- Talents, from the talent search: Shield Specialization 5, Anticipation 2, Improved Revenge 3, Last
  Stand, Master of Defense 2, Improved Thunder Clap 1, Anticipation 4, Improved Thunder Clap 2,
  Improved Bloodrage 2, Vanguard, Focused Rage 3, Improved Thunder Clap 3, Concussion Blow, Bastion
  5, Shield Slam at 35. Then Cruelty 4, Improved Rend 1, Cruelty 5, Unbridled Wrath 5, Enrage 5,
  Improved Rend 3, Improved Execute 2.
- No Defiance. It only adds threat, which does nothing alone.
- Rotations from the `warrior_leveling` template, all the talent and rotation search's ("Searched
  1" and so on).
- Enchant Weapon - Recovery on the main hand from 20 (`recoveryFromLevel`).

The weights are `ui/tank_warrior/leveling_weights_*.json` (two-hand, dual wield and shield), from
`tools/levelingplan -weights`.

### Hours

1 to 60 with Recovery from 20, 8 gear seeds and 400 sessions a level: 96.6h on seeds the search
never saw (gear seeds 41 to 48), and 96.6h on the seeds of the table below (28 deaths).

| Levels | 1-9 | 10-19 | 20-29 | 30-39 | 40-49 | 50-59 |
|---|---|---|---|---|---|---|
| Hours, with Recovery | 2.0 | 8.0 | 11.3 | 16.3 | 26.0 | 32.9 |

Every build's preset, on the same 8 gear seeds (21 to 28) and 400 sessions a level, with Recovery
from 20.

| Preset | Build | Hours | Deaths |
|---|---|---|---|
| Leveling: Prot | Protection, two-hander until 16, then a shield | 96.6 | 28 |
| Leveling: Arms then DW | Arms with a two-hander, Gore Drinker dual wield from 44 | 99.5 | 21 |
| Leveling: Arms, Prot, DW | Arms, Protection with a shield at 30, dual wield at 40 | 101.2 | 20 |
| Leveling: Dual Wield | Gore Drinker dual wield all the way | 102.5 | 10 |
| Leveling: Arms | Arms with a two-hander all the way | 103.0 | 11 |
| Leveling: Fury Prot | Fury Protection 4/31/16 with a shield from 5 | 112.7 | 30 |

How Prot got there, on gear seeds 41 to 48: Prot with hand-tuned rotations 102.4h, with the
search's rotations 99.5h, with its talents too 97.1h, and with the two-hander until 16 96.6h.
Arms then DW took 99.2h on those seeds.

Arms and Dual Wield with no respec die about half as often as the rest, but take 3h longer and
more. Earlier runs without Recovery took 3 to 4h longer on every build.

Fury Protection is a known endgame build, not one people level with.

### Leveling guides

The Forever guides we read (2026-11) all level deep Protection with Shield Slam at 40 (35 with
Talented) and Fury after it: wowforeverinfo.com, kami-labs.fr, method.gg, warcrafttavern.com. Our
talents share their core. Theirs differ in a few points each: Improved Thunder Clap 3 early,
Defiance, Improved Sunder Armor, Lingering Rage, Deflection. Wowforeverinfo's two-hander until
Revenge and Shield Block is what our plan does to 15.

### What we found

- When to go dual wield is a tie from 42 to 48. The plan takes 44.
- Protection with a shield at 30 to 39 dies more and gains nothing.
- Mobs a pull counts every mob chained into one fight, so it overstates the danger. Mobs on us at
  once is the better measure. At 58 Fury Protection has the most, 4.4 on average and 6.4 at the
  peak, with no deaths and 54.0k XP an hour. That's about even with Protection (2.8 and 4.7) and
  Gore Drinker dual wield (1.7 and 3.3, 54.7k XP an hour).
- Pulls of 10 or more mobs lose, even with every heal on.
- Protection's safe pull shrinks in the mid 40s. Starting with 5 mobs is best at 40, but it dies
  at 42 and later. The most we can start with is 4 at 41 and 42, and 3 from 44 to 48. One mob more
  than that costs about 1 death every 20 minutes. XP an hour dips from 46k at 40 to 41k at 46 and
  comes back at 48. Hand-tuned at each of those levels, the rules for 41 to 48 saved 1.2h over
  single pulls there. The search's rotations that the Prot plan has now go further. They bring up
  to 5 mobs into a pull and keep up to 3 on us at 40 to 48, and take about 0.4 to 1 death every 20
  minutes there for it, which still comes out ahead.
- Recovery is worth 3 to 4 hours over the whole plan.
- Mobs 1 level below us are best or tied from 16 to 58. At 10, 2 below is better.
- The talent search's own order didn't beat the hand order above.

### Doubts

- The gear weights come from the sim's own gear plan, not real drops.
- Our logs show more attack power at levels 1 and 2 than the sim has (notes.md).
- Whirlwind's weapon timing.
- Dual wield at 13 is a guess.
- No gold for the respec and no second weapon set are counted.
- Victory Rush's heal doesn't crit in the sim.
- We only tried Beasts, not Humanoids.

## Open guesses

Each is a line in `notes.md`, Need to Verify, section Warrior.
