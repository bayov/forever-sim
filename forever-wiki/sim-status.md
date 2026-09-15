# Sim status

What this repo's sim models of Forever as of 2026-09-14, and what it assumes where nothing is published. Most of the engine work was ported from [ElliotWood/Forever](https://github.com/ElliotWood/Forever) (see [references.md](references.md)), then checked against this wiki and adjusted. Only Warrior, Rogue and Shaman have Forever talent trees so far. The other six classes still run their Classic trees.

## Ruleset switch

`SimOptions.ruleset` picks `RulesetForever` (the wire default, so anything that never sets it runs Forever) or `RulesetClassic`. The web UI has a "Forever Rules" checkbox in the options menu (cog icon in the top bar). The unconverted classes pin their tests to Classic. Everything below is gated on Forever, so Classic results are unchanged. The Warrior, Rogue and Shaman trees are Forever only either way, because the talent proto is positional and cannot hold two trees.

## Core mechanics

| Rule | Modelled as | Source |
| --- | --- | --- |
| Unified hit | Hit printed on gear counts for melee, ranged and spells. Hit from talents and buffs is untouched. | Deep Dive panel |
| Unified crit | Crit printed on gear counts for melee, ranged and spells. Agility and Intellect conversions are untouched. | Deep Dive panel |
| Bonus healing | Healing on gear grants one third as spell damage. | Deep Dive panel |
| Weapon skill | Still Classic. Forever thins it out on gear, but no item numbers exist, so gear is Classic gear. | Deep Dive panel, no numbers |
| Avoidance reduction stat | Not modelled. No item carries it. | Deep Dive panel |
| Periodic crits | Dot ticks roll against the caster's crit chance by school. | ElliotWood/Forever's reading, not confirmed by Blizzard |
| Stormstrike | Not consumed by Nature hits, lasts its full duration. | Spellbook transcription |
| Personal debuffs | Improved Shadow Bolt, Shadow Weaving, Improved Scorch and Winter's Chill are no longer raid debuffs in the test suites and presets. | Talent transcriptions |
| Windfury | The sim already models Windfury Totem as a main hand only proc that cannot be combined with a shaman's own main hand Windfury Weapon, which is what Forever says. | Spellbook transcription |
| Faction buffs | Paladin blessings and Shaman totems (including Windfury Totem) apply to every race. Classic gated blessings to Alliance and totems to Horde. | Dwarf Shaman and Undead Paladin exist, so both factions have both classes |

## Racials

`sim/core/racials_forever.go`, written against [racials.md](racials.md). Cooldowns of the new actives are unpublished. Blood Fury and Berserking keep their Classic cooldowns (2 min and 3 min), Elune's Light and Eureka! assume 3 min.

| Race | Modelled | Assumption |
| --- | --- | --- |
| Human | 2% crit with swords (melee and spell), Spirit +5% | Weapon crit applies to auto attacks too, not only abilities |
| Dwarf | 1% crit with maces, +5% vs Beasts | Same. Creature type bonuses follow the upstream Classic modelling and apply again on crits, so they measure about +8% at 50% crit |
| Night Elf | Elune's Light: +10% crit for 15 sec, Dodge +1% | 3 min cooldown |
| Gnome | Eureka!: next 3 abilities +10% damage and cheaper. Max mana, rage and energy +5% | Cost reduction is not on the tooltip, assumed 50%. 3 min cooldown |
| Orc | 1% crit with axes, Blood Fury +10% AP and SP for 15 sec | 10% of total AP and SP at activation. Command (pet damage) is gone |
| Undead | Touch of the Grave: 5% chance on spells and attacks to deal Shadow damage and heal for it | Modelled at the 5% max health cap with no internal cooldown, so this is an upper bound. `core.TouchOfTheGraveICD` sets a cooldown for comparisons. On the rogue and warrior presets the racial is 1.7 to 3.0% of damage with no cooldown, 1.3 to 2.2% at 5 sec, 1.0 to 1.7% at 10 sec and 0.8 to 1.4% at 15 sec |
| Tauren | Health +5%, hit +1% (melee and spell) | |
| Troll | Berserking flat +10% for 10 sec, +5% vs Beasts | Bow and Throwing Specialization removed |
| Skyborne | +1% melee, ranged and cast speed, +5% vs Elementals | No offensive active on either half, so High Order and Windshaper sim the same. Base stats sit at the class baseline, none published |

The six new race and class pairings are selectable. Skyborne is two entries (High Order for Alliance, Windshaper for Horde) because faction is derived from race throughout the sim.

## Talents

Trees for Warrior, Rogue and Shaman come from `tools/forever_talents/data/` (the nikftw transcription vendored by ElliotWood/Forever). Cross-checked against this wiki's `data/talents.json` (wowforevertalents.com): rows, columns and rank counts agree for all three classes, and rank 1 text agrees except for wording. The one name difference is Blood Craze (wiki) vs Blood Crazed (sim).

Regenerate a tree and its proto with `tools/forever_talents/import_talents.py <class> --write`. The tree json and the proto message have to stay in the same order, so always regenerate both.

The talent picker shows the Forever tooltip for every talent and marks two things:

- Ranks 2 and up that were never on screen and are extrapolated.
- Talents the sim does not read, outlined in the picker. Points spent there change nothing. As of now: Warrior (Improved Charge, Spearing Strike, Improved Hamstring, Booming Voice, Iron Will, Piercing Howl, Blood Crazed, Improved Intercept, Improved Disarm, Vanguard, Improved Shield Bash), Rogue (Improved Gouge, Remorseless Attacks, Improved Kidney Shot, Improved Sprint, Improved Kick, Camouflage, Master of Deception, Setup, Dirty Tricks, Improved Distract, Heightened Senses), Shaman (Eye of the Storm, Elemental Reach, Earthbound, Earth's Grasp, Improved Ghost Wolf, Improved Healing Wave, Ancestral Healing, Healing Focus, Healing Way, Riptide).

Per-class assumptions carried over from the fork are marked `TODO` in `sim/warrior`, `sim/rogue` and `sim/shaman`. The ones that move DPS:

- Warrior: Weaponmaster scales 1% crit, 3% armor ignore, 1% extra attack per point. Improved Battle Shout and Improved Demoralizing Shout are treated as baseline.
- Rogue: Hack and Slash scales linearly. Mutilate costs 60 Energy and needs daggers. Venom costs 25 Energy, and as transcribed (30% poison damage, no damage of its own) it is a 3 to 5% DPS loss at any combo point count because the points are worth more as Eviscerate. Improved Expose Armor's Classic 2/2 armor value is treated as baseline.
- Shaman: Four enhancement rotations ship. "Optimized" (the default) is the tools/rotopt search result: Flame Shock when its DoT is down and Earth Shock otherwise, Lightning Bolt only at 5 Maelstrom Weapon stacks, Rage of the Farseer on cooldown with Blood Fury and Berserking waiting for it when the wait costs no use (it never does on a 2 min Blood Fury against a 3 min Rage in a 4 to 5 min fight, so the holds change nothing). "Grace of Air" holds Grace of Air, "Windfury Totem" holds Windfury Totem for the melee group, and "Twist WF + GoA" twists both the way a raid shaman would. The sim gives the shaman nothing for its own Windfury Totem (Windfury Weapon does not stack with it), so with no second shaman in the group the player's own Strength of Earth and Grace of Air are each worth about 30 DPS, Windfury Totem in place of Grace of Air costs the same 30, and twisting costs about 160. On the Phase 2 gear with the searched talents the Optimized rotation measures about 860 DPS against a Humanoid in the default raid (both blessings, Judgement of Wisdom, every debuff, no second shaman). Judgement of Wisdom is the shaman's mana: without it the pool is dry two minutes in and the rotation loses 60 DPS. With no Mana Spring the pool is dry at five minutes anyway, so the Level 60 talents take Improved Stormstrike over the last two points of Ancestral Knowledge (2 DPS behind in a 4 minute fight, 6 ahead at 8 minutes, 12 at 10). Improved Stormstrike's regen was a no-op in the fork (the mana tick was never recomputed) and is fixed. Windfury Weapon is 100 DPS ahead of Rockbiter and 140 ahead of Flametongue on Sulfuras. A two hander beats any one hander plus shield by 60 or more. Twisting also spends two thirds of the mana pool on totems and runs the shaman dry by 85 sec of a 2 min fight. Maelstrom Weapon has no published proc rate, modelled as 2 PPM per point. Rage of the Farseer assumes a 3 min cooldown. Lava Burst cast time and cooldown are from the WotLK spell. Enhancing Totems is gone so Strength of Earth and Grace of Air always land improved.

## Gear and buffs

All gear is Classic gear with Classic stats, so Classic phase presets are what you pick from. Blizzard has not published any Forever item. World buffs are all off in the presets, since Forever does not have them at launch. They can still be turned on in the Settings tab. Raid buff assumptions are Classic 40-man ones, Forever's launch raids are 10 and 20 man plus Onyxia at 40.

## Comparing races

Racials are the only difference between races, since both factions have blessings and totems. The rogue and shaman APLs hold Blood Fury and Berserking for their burst anchor (Adrenaline Rush, Rage of the Farseer) when the wait costs no use, and fire them on cooldown otherwise. The warrior APLs hold them for the last 20 to 30 sec so they stack with Death Wish and Recklessness, which measures the same or slightly better than firing them at the pull on a 2 min fight.

## Low levels

The sim takes a player level (Other settings) with 1.12 base stats and spell ranks per level, and an Extra Talent Points setting next to it for the Forever beta, which caps at level 20 but is expected to give 16 talent points there instead of 11. The Rogue and Enhancement Shaman UIs ship "Level 20" and "Level 20 + 5" builds (gear, talents, rotation, race, level, extra points and encounter together) from the tools/rotopt searches: every item the level can equip, every build of the damage talents, then the rotation knobs. The encounter is Mutanus the Devourer (level 22, 922 armor), 60 sec, no raid buffs, the boss hitting the shaman so Lightning Shield fires.

- Level 20 (11 points): Sinister Strike 64 DPS, Backstab 62, Enhancement 62 (Rockbiter, Frost Shock, Searing Totem, Lightning Shield).
- Level 20 + 5 (16 points): Sinister Strike 69, Backstab 67, Enhancement 71. Stormstrike sits behind 15 points under Forever, so the 16th point buys it, and it is worth 5 DPS over the best build without it (Flurry 1 in its place). Frost Shock still beats Earth Shock rank 3 by 1.6 (at 25 Earth Shock rank 4 wins by 7) and Strength of Earth is worth its mana. The rogue takes Assassination rows 1 and 2 (Malice, Ruthlessness, Murder, Imp SnD, Relentless Strikes) plus Imp SS, or the dagger talents with Malice for Backstab. Lethality, Cold Blood, Dual Wield Specialization and the poison talents all score lower. Rupture at 3 points beats Eviscerate. A Backstab is 60 energy, so a point comes every 6 sec and a 1 point Slice and Dice (9 sec) is gone before the next point: the Backstab rotations wait for 4 points before putting a fresh one up (+1.4 DPS), or the rogue would refresh it forever and never reach a Rupture. The Sinister Strike build has Improved Slice and Dice, whose 13 sec 1 point one is worth putting up at once. The gear search with 16 points settled on the same level 20 sets.
- Races at 20 + 5: Gnome leads the rogue (Eureka! on cooldown makes three abilities half price, SS 71.0 / BS 70.7), then Night Elf (Elune's Light, 70.6 / 68.4), Human with swords (69.2 / 67.1), Troll and Undead (68.5 / 66.3), Skyborne (68.1 / 66.1), Dwarf (67.1 / 65.2), Orc (66.8 / 64.6). The shaman races are within 1 DPS: Orc 71.2 (Axe Specialization on The Axe of Severing), Skyborne 70.5, Tauren 70.2, Dwarf 70.2, Troll 70.1. Blood Fury is a GCD for 10% of a few hundred attack power at this level.
- The shaman's mana is gone after a minute of Frost Shocks and Stormstrikes: 75 DPS in a 30 sec fight, 71 at 60, 60 at 120, 54 at 300. The rogue loses less over time (SS 73, 69, 66, 64.5, Backstab 75, 67, 64, 61.5: its Thistle Tea and the opening Rupture weigh more in a short fight).
