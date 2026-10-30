// The raid sim page's members: the DPS Report's builds (data/builds.json, written by tools/raid_sim_data
// from tools/dps_report/req) and the tank and healer presets (sim/compsearch/presets.go, through
// data/table.json).
import { Blessing, CompSearchResult, Duty, PaladinJudgement, Player, RaidPackage, RaidPreset, RaidRole, RaidTotem, TotemElement as TotemElementProto } from '../core/proto/api';
import { Class, Encounter as EncounterProto, Faction, MobType, Race } from '../core/proto/common';
import { raceNames } from '../core/proto_utils/names';
import buildsUrl from './data/builds.json?url';
import presetsUrl from './data/presets.json?url';
import table from './data/table.json';

// One request of a build: the base build, a duty variant or an encounter variant.
export interface BuildVariant {
	id: string;
	name: string;
	duty?: Duty;
	mobType?: MobType;
	published: number;
	// Changes when the build's request, race files or published DPS change (sim/compsearch, Pool.Stamps).
	stamp: string;
	players: Partial<Record<Faction, Player>>;
}

export interface ReportBuild extends BuildVariant {
	class: Class;
	className: string;
	spec: string;
	summary: string;
	isDefault: boolean;
	gives: string;
	package: RaidPackage;
	races: Partial<Record<Faction, Race>>;
	duties: BuildVariant[];
	encounters: BuildVariant[];
}

export interface RaidSimData {
	// The report's fight: the Generic boss.
	encounter: EncounterProto;
	builds: ReportBuild[];
	// Every build and variant by id.
	byId: Map<string, { build: ReportBuild; variant: BuildVariant }>;
}

type RawVariant = { id: string; name: string; duty?: string; mobType?: string; published: number; stamp?: string; players: Record<string, unknown> };
type RawBuild = RawVariant & {
	class: string;
	className: string;
	spec: string;
	summary?: string;
	isDefault?: boolean;
	gives: string;
	package?: string;
	races: Record<string, string>;
	duties?: RawVariant[];
	encounters?: RawVariant[];
};

const enumValue = <E>(e: Record<string, unknown>, name: string | undefined, fallback: E): E => (name && e[name] !== undefined ? (e[name] as E) : fallback);

const readVariant = (raw: RawVariant): BuildVariant => {
	const players: Partial<Record<Faction, Player>> = {};
	for (const [faction, player] of Object.entries(raw.players)) {
		players[enumValue(Faction, faction, Faction.Unknown)] = Player.fromJson(player as any, { ignoreUnknownFields: true });
	}
	return {
		id: raw.id,
		name: raw.name,
		duty: raw.duty ? enumValue(Duty, raw.duty, Duty.DutyNone) : undefined,
		mobType: raw.mobType ? enumValue(MobType, raw.mobType, MobType.MobTypeUnknown) : undefined,
		published: raw.published,
		stamp: raw.stamp ?? '',
		players,
	};
};

let dataPromise: Promise<RaidSimData> | null = null;

// The data file, fetched once.
export const loadRaidSimData = (): Promise<RaidSimData> => {
	if (!dataPromise) {
		dataPromise = fetch(buildsUrl)
			.then(res => {
				if (!res.ok) throw new Error(`${res.status} ${res.statusText} for ${buildsUrl}`);
				return res.json();
			})
			.then((raw: { encounter: unknown; builds: RawBuild[] }) => {
				const byId: RaidSimData['byId'] = new Map();
				const builds = raw.builds.map(rb => {
					const races: Partial<Record<Faction, Race>> = {};
					for (const [faction, race] of Object.entries(rb.races))
						races[enumValue(Faction, faction, Faction.Unknown)] = enumValue(Race, race, Race.RaceUnknown);
					const build: ReportBuild = {
						...readVariant(rb),
						class: enumValue(Class, rb.class, Class.ClassUnknown),
						className: rb.className,
						spec: rb.spec,
						summary: rb.summary ?? '',
						isDefault: !!rb.isDefault,
						gives: rb.gives,
						package: enumValue(RaidPackage, rb.package, RaidPackage.RaidPackageNone),
						races,
						duties: (rb.duties ?? []).map(readVariant),
						encounters: (rb.encounters ?? []).map(readVariant),
					};
					byId.set(build.id, { build, variant: build });
					[...build.duties, ...build.encounters].forEach(variant => byId.set(variant.id, { build, variant }));
					return build;
				});
				return { encounter: EncounterProto.fromJson(raw.encounter as any, { ignoreUnknownFields: true }), builds, byId };
			});
		dataPromise.catch(() => (dataPromise = null));
	}
	return dataPromise;
};

// The page's copy of the search's table of who gives what (sim/compsearch/table.go), which
// tools/raid_sim_data writes. Its enums are by name, like "RaidPresetHolyPaladin".
const enumOf = <E>(e: object, id: string): E => (e as Record<string, unknown>)[id] as E;

// A tank or healer placeholder (docs/raid_sim.md, Tank and healer presets).
export interface RaidPresetInfo {
	id: RaidPreset;
	name: string;
	role: RaidRole;
	class: Class;
	talents: string;
	// The effects it can give, by the resolver's names, its choices included.
	gives: string[];
	// What it gives in a line.
	summary: string;
	icon: string;
	paladin?: boolean;
	shaman?: boolean;
	package?: RaidPackage;
	powerInfusion?: boolean;
	innervate?: boolean;
	// The Feral tank's Innervate is off by default, since it means leaving Bear Form.
	innervateOptional?: boolean;
}

const icon = (name: string) => `https://wow.zamimg.com/images/wow/icons/large/${name}.jpg`;

export const RAID_PRESETS: RaidPresetInfo[] = table.presets.map(p => ({
	...p,
	id: enumOf<RaidPreset>(RaidPreset, p.id),
	role: enumOf<RaidRole>(RaidRole, p.role),
	class: enumOf<Class>(Class, p.class),
	icon: icon(p.icon),
	package: p.package ? enumOf<RaidPackage>(RaidPackage, p.package) : undefined,
}));

export const presetInfo = (id: RaidPreset): RaidPresetInfo | undefined => RAID_PRESETS.find(p => p.id == id);
export const presetName = (id: RaidPreset): string => presetInfo(id)?.name ?? RaidPreset[id] ?? `${id}`;
export const presetsFor = (role: RaidRole) => RAID_PRESETS.filter(p => p.role == role);

export const CLASS_ICONS: Partial<Record<Class, string>> = Object.fromEntries(table.classes.map(c => [enumOf<Class>(Class, c.id), icon(c.icon)]));

// The classes a Greater Blessing goes to, in the order the page lists them.
export const BLESSING_CLASSES: Class[] = table.classes.map(c => enumOf<Class>(Class, c.id));

// The members table of docs/raid_sim.md: each raid size's default tank and healer slots, and the usual range.
export const RAID_SIZES: Record<number, { tanks: number; healers: number; tankRange: [number, number]; healerRange: [number, number] }> = {
	10: { tanks: 1, healers: 2, tankRange: [1, 2], healerRange: [2, 2] },
	20: { tanks: 2, healers: 4, tankRange: [2, 2], healerRange: [4, 5] },
	40: { tanks: 3, healers: 8, tankRange: [3, 3], healerRange: [7, 9] },
};

// Names for the raid sim's enums, as the page writes them.
const namesOf = <E>(e: object, rows: { id: string; name: string; label?: string }[]) => new Map<E, string>(rows.map(r => [enumOf<E>(e, r.id), r.label || r.name]));
const DUTY_NAMES = namesOf<Duty>(Duty, table.duties);
const BLESSING_NAMES = namesOf<Blessing>(Blessing, table.blessings);
const JUDGEMENT_NAMES = namesOf<PaladinJudgement>(PaladinJudgement, table.judgements);
const TOTEM_NAMES = namesOf<RaidTotem>(RaidTotem, table.totems);
const PACKAGE_NAMES = namesOf<RaidPackage>(RaidPackage, table.packages);

export const dutyName = (d: Duty): string => DUTY_NAMES.get(d) ?? (d == Duty.DutyNone ? 'none' : `${d}`);
export const blessingName = (b: Blessing): string => BLESSING_NAMES.get(b) ?? (b == Blessing.BlessingNone ? 'none' : `${b}`);
// After "Judgement of".
export const judgementName = (j: PaladinJudgement): string =>
	JUDGEMENT_NAMES.get(j) ?? (j == PaladinJudgement.PaladinJudgementNone ? 'none' : `${j}`);
// Before "Totem".
export const totemName = (t: RaidTotem): string => TOTEM_NAMES.get(t) ?? (t == RaidTotem.RaidTotemNone ? 'none' : `${t}`);

export type TotemElement = 'earth' | 'air' | 'water' | 'fire';
export const TOTEM_ELEMENTS: { element: TotemElement; proto: TotemElementProto; label: string; totems: RaidTotem[] }[] = table.elements.map(e => ({
	element: e.name.toLowerCase() as TotemElement,
	proto: enumOf<TotemElementProto>(TotemElementProto, e.id),
	label: e.name,
	totems: table.totems.filter(t => t.of == e.id).map(t => enumOf<RaidTotem>(RaidTotem, t.id)),
}));
export const totemElement = (t: RaidTotem): TotemElement | undefined => TOTEM_ELEMENTS.find(e => e.totems.includes(t))?.element;
// A totem assignment's element: its own, or its totem's (RaidTotemNone names its element).
export const assignedElement = (a: { totem: RaidTotem; element: TotemElementProto }): TotemElement | undefined =>
	TOTEM_ELEMENTS.find(e => e.proto == a.element)?.element ?? totemElement(a.totem);

export const packageName = (p: RaidPackage): string => PACKAGE_NAMES.get(p) ?? 'none';

export const roleName = (r: RaidRole): string =>
	({ [RaidRole.RaidRoleDps]: 'DPS', [RaidRole.RaidRoleTank]: 'Tank', [RaidRole.RaidRoleHealer]: 'Healer' })[r] ?? `${r}`;

export const factionName = (f: Faction): string => ({ [Faction.Horde]: 'Horde', [Faction.Alliance]: 'Alliance' })[f as number] ?? 'Unknown';

// The search writes a race as its enum name without the prefix, like "NightElf" or "SkyborneWindshaper".
export const raceLabel = (race: string | Race): string => {
	const value = typeof race == 'string' ? enumValue(Race, race.startsWith('Race') ? race : `Race${race}`, Race.RaceUnknown) : race;
	if (value == Race.RaceUnknown) return typeof race == 'string' ? race : '';
	return raceNames.get(value) ?? `${race}`;
};

export const mobTypeName = (m: MobType): string => (MobType[m] ?? `${m}`).replace(/^MobType/, '');

// A build's name as the page writes it, like "Mage: Frostfire". The report names some builds only by their
// level ("Level 60"), and those get their spec.
export const buildLabel = (build: ReportBuild, variant: BuildVariant = build): string => {
	const generic = !variant.name.includes(build.spec) && (variant.name.startsWith('Level') || build.class == Class.ClassShaman);
	return `${build.className}: ${generic ? `${build.spec} ${variant.name}` : variant.name}`;
};

// How the builds a search chose from differ from the page's: the ones whose stamp changed, the ones gone
// from the pool, and the ones it didn't have. Undefined for a result from before the stamps.
export interface Staleness {
	changed: string[];
	gone: string[];
	added: string[];
}

export const stalenessOf = (stamps: { [id: string]: string }, data: RaidSimData): Staleness | undefined => {
	if (!Object.keys(stamps).length) return undefined;
	const s: Staleness = { changed: [], gone: [], added: [] };
	for (const [id, stamp] of Object.entries(stamps)) {
		const found = data.byId.get(id);
		if (!found) s.gone.push(id);
		else if (found.variant.stamp != stamp) s.changed.push(id);
	}
	data.byId.forEach((_, id) => {
		if (!(id in stamps)) s.added.push(id);
	});
	return s.changed.length || s.gone.length || s.added.length ? s : undefined;
};

// A staleness in a line, like "Builds changed since: Mage: Fire, Hunter: Beast Mastery · 2 new".
export const stalenessText = (s: Staleness, data: RaidSimData): string => {
	const names = [...new Set(s.changed.map(id => data.byId.get(id)).map(f => (f ? buildLabel(f.build, f.variant) : '')))].filter(n => n);
	const shown = names.slice(0, 4).join(', ') + (names.length > 4 ? ` +${names.length - 4}` : '');
	const parts = names.length ? [`Builds changed since: ${shown}`] : [];
	if (s.added.length) parts.push(`${s.added.length} new`);
	if (s.gone.length) parts.push(`${s.gone.length} gone`);
	return parts.join(' · ');
};

// The built-in raids' searches (data/presets.json, written by tools/raid_sim_data from the committed
// write-ups): each search's request and each scenario's best comp. Fetched once.
let presetsPromise: Promise<CompSearchResult[]> | null = null;
export const loadRaidSimPresets = (): Promise<CompSearchResult[]> => {
	if (!presetsPromise) {
		presetsPromise = fetch(presetsUrl)
			.then(res => {
				if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
				return res.json();
			})
			.then((results: unknown[]) => results.map(r => CompSearchResult.fromJson(r as any, { ignoreUnknownFields: true })));
		presetsPromise.catch(() => (presetsPromise = null));
	}
	return presetsPromise;
};
