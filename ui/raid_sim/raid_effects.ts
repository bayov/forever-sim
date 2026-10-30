// What each member of the comp editor's raid can give, and the icons the page draws for it.
//
// The raid's effects are the resolver's gifts, by the names its buff records use. They, and everything
// else here, come from the search's table of who gives what (sim/compsearch/table.go) through
// data/table.json. Before a run we only know who could give each one: a build's `gives` and its duty
// variants, or a placeholder's preset. After a run the resolver's record says what was up.
import { Blessing, Duty, PaladinJudgement, RaidPackage, RaidPreset, RaidTotem } from '../core/proto/api';
import { Class } from '../core/proto/common';
import { CLASS_ICONS, presetInfo, RaidPresetInfo, ReportBuild } from './data';
import table from './data/table.json';

export const wowIcon = (name: string) => `https://wow.zamimg.com/images/wow/icons/large/${name}.jpg`;

export interface RaidEffect {
	// The resolver's name for it, as its buff records and the builds' `gives` write it.
	name: string;
	// The name the page shows, when it differs.
	label?: string;
	icon: string;
	debuff?: boolean;
	// Who can give it, for the Who gives what table.
	from: string;
}

// The raid's effects in the order the page's strip shows them: the long buffs, the held auras and totems,
// then the boss's debuffs.
export const RAID_EFFECTS: RaidEffect[] = table.effects.map(e => ({ ...e, icon: wowIcon(e.icon) }));

// An enum value by its name in the table, like "DutyBattleShout".
const enumOf = <E>(e: object, id: string): E => (e as Record<string, unknown>)[id] as E;

// What a duty gives the raid.
export const DUTY_EFFECT: Partial<Record<Duty, string>> = Object.fromEntries(table.duties.map(d => [enumOf<Duty>(Duty, d.id), d.effect]));

// The effects a member could give: a placeholder's, or a build's own and those of its duty variants.
export const effectsOf = (preset: RaidPreset | undefined, build: ReportBuild | undefined): string[] => {
	if (preset) return presetInfo(preset)?.gives ?? [];
	if (!build) return [];
	const own = build.gives == 'nothing' ? [] : build.gives.split(',').map(s => s.trim());
	// A paladin's Salvation comes with its Blessings.
	if (own.includes('Blessings')) own.push('Salvation');
	const duties = build.duties.map(d => (d.duty !== undefined ? DUTY_EFFECT[d.duty] : undefined)).filter((e): e is string => !!e);
	return [...new Set([...own, ...duties])];
};

// A party's mana package (Mana Spring, and Mana Tide), as the resolver records it.
export const PACKAGE_ICON = wowIcon(table.totems.find(t => t.id == 'RaidTotemManaSpring')?.icon ?? '');
export const MANA_TIDE_ICON = wowIcon(table.icons.manaTide);

export const BLESSING_ICONS: Partial<Record<Blessing, string>> = Object.fromEntries(table.blessings.map(b => [enumOf<Blessing>(Blessing, b.id), wowIcon(b.icon)]));

export const JUDGEMENT_ICONS: Partial<Record<PaladinJudgement, string>> = Object.fromEntries(
	table.judgements.map(j => [enumOf<PaladinJudgement>(PaladinJudgement, j.id), wowIcon(j.icon)]),
);

export const TOTEM_ICONS: Partial<Record<RaidTotem, string>> = Object.fromEntries(table.totems.map(t => [enumOf<RaidTotem>(RaidTotem, t.id), wowIcon(t.icon)]));

export const DUTY_ICONS: Partial<Record<Duty, string>> = Object.fromEntries(table.duties.map(d => [enumOf<Duty>(Duty, d.id), wowIcon(d.icon)]));

export const POWER_INFUSION_ICON = wowIcon(table.icons.powerInfusion);
export const INNERVATE_ICON = wowIcon(table.icons.innervate);

// Each build's spec icon, by its class and the report's spec name. A spec the table lacks takes its
// class's icon.
const SPEC_ICONS: Record<string, string> = table.specIcons;

export const buildIcon = (build: ReportBuild): string => {
	const name = SPEC_ICONS[`${build.className}:${build.spec}`];
	return name ? wowIcon(name) : CLASS_ICONS[build.class] ?? '';
};

// The icon of a member: its preset's, or its build's spec.
export const memberIcon = (info: RaidPresetInfo | undefined, build: ReportBuild | undefined, cls: Class): string =>
	info?.icon ?? (build ? buildIcon(build) : CLASS_ICONS[cls] ?? '');

// The party package's words and whether it has Mana Tide.
export const packageHasTide = (p: RaidPackage) => p == RaidPackage.RaidPackageManaSpringTide || p == RaidPackage.RaidPackageRestoration;
