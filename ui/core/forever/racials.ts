import { Class, Race } from '../proto/common';
import { IconData } from '../proto/ui';
import RacialsJson from './racials.json';

// The Forever racials, two actives and two passives per race, as the beta client shows
// them (build 1.60.1.69876, read through hyjal.cc). A racial with class variants, like
// Eureka! or Touch of the Grave, lists one entry per variant with the classes it applies
// to. Berserking's duration is not in the client data, the 10 sec is Classic's.

export interface RacialVariant {
	spellId: number;
	classes: Array<keyof typeof Class>;
	description: string;
	icon: string;
	passive: boolean;
	cost: string | null;
}

export interface RacialData {
	name: string;
	variants: Array<RacialVariant>;
}

// The racial as one class sees it.
export interface Racial {
	name: string;
	spellId: number;
	description: string;
	icon: string;
	passive: boolean;
	cost: string | null;
}

const racialsByRace = RacialsJson as Record<keyof typeof Race, Array<RacialData>>;

export function getRacials(race: Race, klass: Class): Array<Racial> {
	const raceName = Race[race] as keyof typeof Race;
	const className = Class[klass] as keyof typeof Class;
	return (racialsByRace[raceName] ?? []).flatMap(racial => {
		const variant = racial.variants.find(v => v.classes.includes(className)) ?? racial.variants[0];
		if (!variant) return [];
		return [
			{
				name: racial.name,
				spellId: variant.spellId,
				description: variant.description,
				icon: variant.icon,
				passive: variant.passive,
				cost: variant.cost,
			},
		];
	});
}

export function racialIconUrl(icon: string): string {
	return `https://wow.zamimg.com/images/wow/icons/large/${icon}.jpg`;
}

// Icons and names for the racial spell ids Wowhead does not know (the new Forever
// spells), so an APL action or a timeline entry for Eureka! or Elune's Light still gets
// its icon.
export const foreverSpellIcons: Array<IconData> = Object.values(racialsByRace)
	.flat()
	.flatMap(racial =>
		racial.variants.map(variant =>
			IconData.create({
				id: variant.spellId,
				name: racial.name,
				icon: variant.icon,
				hasBuff: !variant.passive,
			}),
		),
	);
