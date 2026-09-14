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
| Undead | Touch of the Grave: 5% chance on spells and attacks to deal Shadow damage and heal for it | Modelled at the 5% max health cap, so this is an upper bound |
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
- Shaman: Three enhancement rotations ship. "Grace of Air" (the default) holds Grace of Air, "Windfury Totem" holds Windfury Totem for the melee group, and "Twist WF + GoA" twists both the way a raid shaman would. The sim gives the shaman nothing for its own Windfury Totem (Windfury Weapon does not stack with it), so against a Humanoid the three measure about 768, 740 and 645 DPS. Twisting also spends two thirds of the mana pool on totems and runs the shaman dry by 85 sec of a 2 min fight. Maelstrom Weapon has no published proc rate, modelled as 2 PPM per point. Rage of the Farseer assumes a 3 min cooldown. Lava Burst cast time and cooldown are from the WotLK spell. Enhancing Totems is gone so Strength of Earth and Grace of Air always land improved.

## Gear and buffs

All gear is Classic gear with Classic stats, so Classic phase presets are what you pick from. Blizzard has not published any Forever item. World buffs are all off in the presets, since Forever does not have them at launch. They can still be turned on in the Settings tab. Raid buff assumptions are Classic 40-man ones, Forever's launch raids are 10 and 20 man plus Onyxia at 40.

## Comparing races

Racials are the only difference between races, since both factions have blessings and totems. The rogue and shaman APLs fire Blood Fury and Berserking as soon as they are ready. The warrior APLs hold them for the last 20 to 30 sec so they stack with Death Wish and Recklessness, which measures the same or slightly better than firing them at the pull on a 2 min fight.
