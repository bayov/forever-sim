import { gameSpellTooltip } from '../components/game_spell_tooltip';
import { WeaponImbue } from '../proto/common';

// The shaman weapon imbues' ranks under Forever, for their tooltips on the Settings tab. Keep
// the levels and the values in sync with sim/shaman/*_weapon.go.
//
// The mana costs and the values come from wowhead's Forever tooltips. A rank's value grows
// from the level we learn it up to its cap level (wowhead's <!--ppl<learn>:<cap>:<base>:<growth
// * 100>--> marker). The top ranks have a cap past 60, so at 60 they show less than wowhead.
//
// In the game, Flametongue and Frostbrand also add some of our spell power to the damage they
// show. We leave it out, because the sim adds it on top.
interface ImbueRank {
	level: number;
	mana: number;
	base: number;
	capLevel: number;
	growth: number;
}

interface ImbueData {
	name: string;
	ranks: Array<ImbueRank>;
	description: (value: number) => string;
}

const imbues: Partial<Record<WeaponImbue, ImbueData>> = {
	[WeaponImbue.WindfuryWeapon]: {
		name: 'Windfury Weapon',
		ranks: [
			{ level: 30, mana: 90, base: 46, capLevel: 38, growth: 7.2 },
			{ level: 40, mana: 115, base: 119, capLevel: 48, growth: 12.8 },
			{ level: 50, mana: 140, base: 249, capLevel: 58, growth: 8.3 },
			{ level: 60, mana: 165, base: 333, capLevel: 68, growth: 12.5 },
		],
		description: ap =>
			`Imbue the Shaman's weapon with wind. Each hit has a 20% chance of granting you 2 extra attacks with ${ap} extra melee attack power. ` +
			'When applied to main hand, disables any benefit you personally receive from Windfury Totem. Lasts for 60 minutes.',
	},
	[WeaponImbue.RockbiterWeapon]: {
		name: 'Rockbiter Weapon',
		ranks: [
			{ level: 1, mana: 15, base: 29, capLevel: 6, growth: 4.1 },
			{ level: 8, mana: 25, base: 58, capLevel: 14, growth: 3.5 },
			{ level: 16, mana: 50, base: 88, capLevel: 22, growth: 5 },
			{ level: 24, mana: 75, base: 129, capLevel: 32, growth: 8.1 },
			{ level: 34, mana: 100, base: 211, capLevel: 42, growth: 18 },
			{ level: 44, mana: 125, base: 393, capLevel: 52, growth: 16.1 },
			{ level: 54, mana: 150, base: 554, capLevel: 62, growth: 16.5 },
		],
		description: ap =>
			`Imbue the Shaman's weapon, increasing melee attack power by ${ap} and allowing melee attacks to cause additional threat when using that weapon. ` +
			'Lasts for 60 minutes.',
	},
	// The value is the fire damage per 100 seconds of weapon speed. The game shows it as the
	// damage on a fast weapon (value / 77 - 1) to the damage on a slow one (value / 25).
	[WeaponImbue.FlametongueWeapon]: {
		name: 'Flametongue Weapon',
		ranks: [
			{ level: 10, mana: 30, base: 326, capLevel: 16, growth: 19 },
			{ level: 18, mana: 55, base: 479, capLevel: 24, growth: 29 },
			{ level: 26, mana: 80, base: 716, capLevel: 34, growth: 42 },
			{ level: 36, mana: 105, base: 1144, capLevel: 44, growth: 73 },
			{ level: 46, mana: 130, base: 1876, capLevel: 54, growth: 62 },
			{ level: 56, mana: 155, base: 2498, capLevel: 64, growth: 78 },
		],
		description: value =>
			`Imbue the Shaman's weapon with fire. Each hit causes ${Math.floor(value / 77 - 1)} to ${Math.floor(value / 25)} additional Fire damage, ` +
			'based on the speed of the weapon. Slower weapons cause more fire damage per swing. ' +
			'When applied to main hand, disables any benefit you personally receive from Flametongue Totem. Lasts for 60 minutes.',
	},
	[WeaponImbue.FrostbrandWeapon]: {
		name: 'Frostbrand Weapon',
		ranks: [
			{ level: 20, mana: 60, base: 32, capLevel: 26, growth: 2.1 },
			{ level: 28, mana: 85, base: 48, capLevel: 36, growth: 3 },
			{ level: 38, mana: 110, base: 77, capLevel: 46, growth: 5 },
			{ level: 48, mana: 135, base: 127, capLevel: 56, growth: 4 },
			{ level: 58, mana: 160, base: 158, capLevel: 66, growth: 5.6 },
		],
		description: damage =>
			`Imbue the Shaman's weapon with frost. Each hit has a chance of causing ${damage} additional Frost damage ` +
			"and slowing the target's movement speed by 25% for 8 sec. Lasts for 60 minutes.",
	},
};

// The tooltip of the highest rank we have at our level, with its values at that level. Below
// the first rank's level, it shows the first rank and the level it needs, in red.
export function shamanImbueTooltip(imbue: WeaponImbue, level: number): HTMLElement | undefined {
	const data = imbues[imbue];
	if (!data) return undefined;

	const known = data.ranks.filter(rank => rank.level <= level);
	const rank = known.length ? known[known.length - 1] : data.ranks[0];
	const atLevel = Math.max(rank.level, Math.min(level, rank.capLevel));
	const value = Math.floor(rank.base + (atLevel - rank.level) * rank.growth);

	return gameSpellTooltip({
		name: data.name,
		rank: `Rank ${data.ranks.indexOf(rank) + 1}`,
		lines: [[`${rank.mana} Mana`], ['Instant']],
		unmet: known.length ? undefined : `Requires level ${rank.level}`,
		description: data.description(value),
	});
}
