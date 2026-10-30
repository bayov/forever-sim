// The raid sim page's members: the DPS Report's builds (data/builds.json, written by tools/raid_sim_data
// from tools/dps_report/req) and the tank and healer presets (sim/compsearch/presets.go).
import { Blessing, Duty, PaladinJudgement, Player, RaidPackage, RaidPreset, RaidRole, RaidTotem, TotemElement as TotemElementProto } from '../core/proto/api';
import { Class, Encounter as EncounterProto, Faction, MobType, Race } from '../core/proto/common';
import { raceNames } from '../core/proto_utils/names';
import buildsUrl from './data/builds.json?url';

// One request of a build: the base build, a duty variant or an encounter variant.
export interface BuildVariant {
	id: string;
	name: string;
	duty?: Duty;
	mobType?: MobType;
	published: number;
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

type RawVariant = { id: string; name: string; duty?: string; mobType?: string; published: number; players: Record<string, unknown> };
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

// A tank or healer placeholder (docs/raid_sim.md, Tank and healer presets), as sim/compsearch/presets.go
// has it.
export interface RaidPresetInfo {
	id: RaidPreset;
	name: string;
	role: RaidRole;
	class: Class;
	talents: string;
	gives: string;
	icon: string;
	paladin?: boolean;
	shaman?: boolean;
	package?: RaidPackage;
	powerInfusion?: boolean;
	innervate?: boolean;
	// The Feral tank's Innervate is off by default, since it means leaving Bear Form.
	innervateOptional?: boolean;
	// Holds Faerie Fire as an assignment.
	faerieFire?: boolean;
}

const icon = (name: string) => `https://wow.zamimg.com/images/wow/icons/large/${name}.jpg`;

export const RAID_PRESETS: RaidPresetInfo[] = [
	{
		id: RaidPreset.RaidPresetProtectionWarrior,
		name: 'Protection warrior',
		role: RaidRole.RaidRoleTank,
		class: Class.ClassWarrior,
		talents: '35--251533120331221351',
		gives: 'Sunder Armor upkeep after the pull; Battle Shout',
		icon: icon('ability_warrior_defensivestance'),
	},
	{
		id: RaidPreset.RaidPresetFeralTank,
		name: 'Feral tank',
		role: RaidRole.RaidRoleTank,
		class: Class.ClassDruid,
		talents: '-55220302020132012551-055',
		gives: 'Leader of the Pack; Gift of the Wild; Faerie Fire as an assignment; Innervate when switched on',
		icon: icon('ability_racial_bearform'),
		innervateOptional: true,
		faerieFire: true,
	},
	{
		id: RaidPreset.RaidPresetProtectionPaladin,
		name: 'Protection paladin',
		role: RaidRole.RaidRoleTank,
		class: Class.ClassPaladin,
		talents: '1-5530513321301551-052',
		gives: 'one Greater Blessing for each class; one judgement (Wisdom or the Crusader)',
		icon: icon('spell_holy_devotionaura'),
		paladin: true,
	},
	{
		id: RaidPreset.RaidPresetWardenShaman,
		name: 'Warden shaman',
		role: RaidRole.RaidRoleTank,
		class: Class.ClassShaman,
		talents: '35050013-0502303313551102',
		gives: 'one totem per element; plain Mana Spring to its party',
		icon: icon('spell_shaman_lavaflow'),
		shaman: true,
		package: RaidPackage.RaidPackageManaSpring,
	},
	{
		id: RaidPreset.RaidPresetHolyPaladin,
		name: 'Holy paladin',
		role: RaidRole.RaidRoleHealer,
		class: Class.ClassPaladin,
		talents: '05320103225101051-500250302003',
		gives: 'one Greater Blessing for each class; one judgement (Wisdom or the Crusader)',
		icon: icon('spell_holy_holybolt'),
		paladin: true,
	},
	{
		id: RaidPreset.RaidPresetRestorationShaman,
		name: 'Restoration shaman',
		role: RaidRole.RaidRoleHealer,
		class: Class.ClassShaman,
		talents: '-005102-5530500315513151',
		gives: 'one totem per element; Mana Spring ×1.25 and Mana Tide to its party',
		icon: icon('spell_nature_magicimmunity'),
		shaman: true,
		package: RaidPackage.RaidPackageRestoration,
	},
	{
		id: RaidPreset.RaidPresetDisciplinePriest,
		name: 'Discipline priest',
		role: RaidRole.RaidRoleHealer,
		class: Class.ClassPriest,
		talents: '005005231330011531-2310510303',
		gives: 'Fortitude, Prayer of Spirit, Power Infusion',
		icon: icon('spell_holy_powerinfusion'),
		powerInfusion: true,
	},
	{
		id: RaidPreset.RaidPresetHolyPriest,
		name: 'Holy priest',
		role: RaidRole.RaidRoleHealer,
		class: Class.ClassPriest,
		talents: '00510223123-03405103030121531',
		gives: 'Fortitude, Prayer of Spirit',
		icon: icon('spell_holy_guardianspirit'),
	},
	{
		id: RaidPreset.RaidPresetRestorationDruid,
		name: 'Restoration druid',
		role: RaidRole.RaidRoleHealer,
		class: Class.ClassDruid,
		talents: '05302001--5052035153113051',
		gives: 'Gift of the Wild; Innervate; Faerie Fire as an assignment',
		icon: icon('spell_nature_healingtouch'),
		innervate: true,
		faerieFire: true,
	},
];

export const presetInfo = (id: RaidPreset): RaidPresetInfo | undefined => RAID_PRESETS.find(p => p.id == id);
export const presetName = (id: RaidPreset): string => presetInfo(id)?.name ?? RaidPreset[id] ?? `${id}`;
export const presetsFor = (role: RaidRole) => RAID_PRESETS.filter(p => p.role == role);

export const CLASS_ICONS: Partial<Record<Class, string>> = {
	[Class.ClassDruid]: icon('class_druid'),
	[Class.ClassHunter]: icon('class_hunter'),
	[Class.ClassMage]: icon('class_mage'),
	[Class.ClassPaladin]: icon('class_paladin'),
	[Class.ClassPriest]: icon('class_priest'),
	[Class.ClassRogue]: icon('class_rogue'),
	[Class.ClassShaman]: icon('class_shaman'),
	[Class.ClassWarlock]: icon('class_warlock'),
	[Class.ClassWarrior]: icon('class_warrior'),
};

// The classes a Greater Blessing goes to, in the order the page lists them.
export const BLESSING_CLASSES: Class[] = [
	Class.ClassWarrior,
	Class.ClassPaladin,
	Class.ClassHunter,
	Class.ClassRogue,
	Class.ClassPriest,
	Class.ClassShaman,
	Class.ClassMage,
	Class.ClassWarlock,
	Class.ClassDruid,
];

// The members table of docs/raid_sim.md: each raid size's default tank and healer slots, and the usual range.
export const RAID_SIZES: Record<number, { tanks: number; healers: number; tankRange: [number, number]; healerRange: [number, number] }> = {
	10: { tanks: 1, healers: 2, tankRange: [1, 2], healerRange: [2, 2] },
	20: { tanks: 2, healers: 4, tankRange: [2, 2], healerRange: [4, 5] },
	40: { tanks: 3, healers: 8, tankRange: [3, 3], healerRange: [7, 9] },
};

// Names for the raid sim's enums, as the page writes them.
export const dutyName = (d: Duty): string =>
	({
		[Duty.DutyNone]: 'none',
		[Duty.DutyBattleShout]: 'Battle Shout',
		[Duty.DutySunderUpkeep]: 'Sunder Armor upkeep',
		[Duty.DutyFaerieFire]: 'Faerie Fire',
		[Duty.DutyCurseOfTheElements]: 'Curse of the Elements',
		[Duty.DutyCurseOfRecklessness]: 'Curse of Recklessness',
		[Duty.DutyJudgementOfTheCrusader]: 'Judgement of the Crusader',
		[Duty.DutyJudgementOfWisdom]: 'Judgement of Wisdom',
	})[d] ?? `${d}`;

export const blessingName = (b: Blessing): string =>
	({ [Blessing.BlessingNone]: 'none', [Blessing.BlessingKings]: 'Kings', [Blessing.BlessingMight]: 'Might', [Blessing.BlessingWisdom]: 'Wisdom' })[b] ??
	`${b}`;

export const judgementName = (j: PaladinJudgement): string =>
	({
		[PaladinJudgement.PaladinJudgementNone]: 'none',
		[PaladinJudgement.PaladinJudgementCrusader]: 'the Crusader',
		[PaladinJudgement.PaladinJudgementWisdom]: 'Wisdom',
	})[j] ?? `${j}`;

export const totemName = (t: RaidTotem): string =>
	({
		[RaidTotem.RaidTotemNone]: 'none',
		[RaidTotem.RaidTotemStrengthOfEarth]: 'Strength of Earth',
		[RaidTotem.RaidTotemGraceOfAir]: 'Grace of Air',
		[RaidTotem.RaidTotemWindfury]: 'Windfury',
		[RaidTotem.RaidTotemManaSpring]: 'Mana Spring',
		[RaidTotem.RaidTotemFlametongue]: 'Flametongue',
	})[t] ?? `${t}`;

export type TotemElement = 'earth' | 'air' | 'water' | 'fire';
export const TOTEM_ELEMENTS: { element: TotemElement; proto: TotemElementProto; label: string; totems: RaidTotem[] }[] = [
	{ element: 'earth', proto: TotemElementProto.TotemElementEarth, label: 'Earth', totems: [RaidTotem.RaidTotemStrengthOfEarth] },
	{ element: 'air', proto: TotemElementProto.TotemElementAir, label: 'Air', totems: [RaidTotem.RaidTotemGraceOfAir, RaidTotem.RaidTotemWindfury] },
	{ element: 'water', proto: TotemElementProto.TotemElementWater, label: 'Water', totems: [RaidTotem.RaidTotemManaSpring] },
	{ element: 'fire', proto: TotemElementProto.TotemElementFire, label: 'Fire', totems: [RaidTotem.RaidTotemFlametongue] },
];
export const totemElement = (t: RaidTotem): TotemElement | undefined => TOTEM_ELEMENTS.find(e => e.totems.includes(t))?.element;
// A totem assignment's element: its own, or its totem's (RaidTotemNone names its element).
export const assignedElement = (a: { totem: RaidTotem; element: TotemElementProto }): TotemElement | undefined =>
	TOTEM_ELEMENTS.find(e => e.proto == a.element)?.element ?? totemElement(a.totem);

export const packageName = (p: RaidPackage): string =>
	({
		[RaidPackage.RaidPackageNone]: 'none',
		[RaidPackage.RaidPackageManaSpring]: 'Mana Spring',
		[RaidPackage.RaidPackageManaSpringTide]: 'Mana Spring and Mana Tide',
		[RaidPackage.RaidPackageRestoration]: 'Mana Spring ×1.25 and Mana Tide',
	})[p] ?? `${p}`;

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
