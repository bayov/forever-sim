import { IndividualLinkImporter } from './components/individual_sim_ui/importers';
import Toast, { ToastOptions } from './components/toast';
import * as Tooltips from './constants/tooltips.js';
import { Player } from './player.js';
import { APLRotation, APLRotation_Type as APLRotationType } from './proto/apl.js';
import {
	Consumes,
	Debuffs,
	Encounter as EncounterProto,
	EquipmentSpec,
	Faction,
	HealingModel,
	IndividualBuffs,
	Profession,
	Race,
	RaidBuffs,
	Spec,
	UnitReference,
} from './proto/common.js';
import { SavedRotation, SavedTalents } from './proto/ui.js';
import { Stats } from './proto_utils/stats.js';
import { SpecOptions, SpecRotation, specTypeFunctions } from './proto_utils/utils.js';

interface PresetBase {
	name: string;
	tooltip?: string;
	// The folder the preset is listed in, like 'Level 30 PvP'. Left out means the top level.
	group?: string;
	enableWhen?: (obj: Player<any>) => boolean;
	onLoad?: (player: Player<any>) => void;
}

interface PresetOptionsBase extends Pick<PresetBase, 'onLoad' | 'group'> {
	customCondition?: (player: Player<any>) => boolean;
}

export interface PresetGear extends PresetBase {
	name: string;
	gear: EquipmentSpec;
	tooltip?: string;
	enableWhen?: (obj: Player<any>) => boolean;
}
export interface PresetGearOptions extends PresetOptionsBase, Pick<PresetBase, 'tooltip'> {
	talentTree?: number;
	talentTrees?: Array<number>;
	faction?: Faction;
	customCondition?: (player: Player<any>) => boolean;
}

export interface PresetTalents extends Pick<PresetBase, 'group' | 'tooltip'> {
	name: string;
	data: SavedTalents;
	enableWhen?: (obj: Player<any>) => boolean;
}
export interface PresetTalentsOptions extends Pick<PresetBase, 'group' | 'tooltip'> {
	customCondition?: (player: Player<any>) => boolean;
}

export interface PresetRotation extends PresetBase {
	name: string;
	rotation: SavedRotation;
	tooltip?: string;
	enableWhen?: (obj: Player<any>) => boolean;
}
export interface PresetRotationOptions extends Pick<PresetOptionsBase, 'onLoad' | 'group'>, Pick<PresetBase, 'tooltip'> {
	talentTree?: number;
	customCondition?: (player: Player<any>) => boolean;
}

export interface PresetEpWeights extends PresetBase {
	epWeights: Stats;
}
export interface PresetEpWeightsOptions extends PresetOptionsBase, Pick<PresetBase, 'tooltip'> {}

export interface PresetEncounter extends PresetBase {
	encounter?: EncounterProto;
	healingModel?: HealingModel;
	tanks?: UnitReference[];
	raidBuffs?: RaidBuffs;
	debuffs?: Debuffs;
	buffs?: IndividualBuffs;
	consumes?: Consumes;
	// Whether we attack from in front of the target, where it can parry and block. Left
	// out means the encounter does not touch the setting.
	inFrontOfTarget?: boolean;
}
// The buffs and consumes can be given directly instead of through an exported link.
export interface PresetEncounterOptions
	extends PresetOptionsBase,
		Pick<PresetEncounter, 'healingModel' | 'tanks' | 'raidBuffs' | 'debuffs' | 'buffs' | 'consumes' | 'inFrontOfTarget'> {}

export interface PresetBuild {
	name: string;
	group?: string;
	// A short line about the build, shown when we hover it.
	tooltip?: string;
	gear?: PresetGear;
	talents?: PresetTalents;
	rotation?: PresetRotation;
	rotationType?: APLRotationType;
	epWeights?: PresetEpWeights;
	encounter?: PresetEncounter;
	race?: Race;
	// Character level for a low level build (the level 20 rogue sets for example). Left
	// out for a level 60 build so applying it does not touch the level.
	level?: number;
	// Talent points beyond what the level grants, see Player.bonus_talent_points. Left out
	// means the build does not touch the setting.
	bonusTalentPoints?: number;
	// The two professions, because some of the gear and consumes need them (Engineering
	// goggles, Goblin Sapper Charge). Left out means the build does not touch them.
	professions?: [Profession, Profession];
	options?: Partial<SpecOptions<any>>;
}

export interface PresetBuildOptions extends Omit<PresetBuild, 'name'> {}

export function makePresetGear(name: string, gearJson: any, options?: PresetGearOptions): PresetGear {
	const gear = EquipmentSpec.fromJson(gearJson);
	return makePresetGearHelper(name, gear, options || {});
}

function makePresetGearHelper(name: string, gear: EquipmentSpec, options: PresetGearOptions): PresetGear {
	const conditions: Array<(player: Player<any>) => boolean> = [];
	if (options.talentTree != undefined) {
		conditions.push((player: Player<any>) => player.getTalentTree() == options.talentTree);
	}
	if (options.talentTrees != undefined) {
		conditions.push((player: Player<any>) => (options.talentTrees || []).includes(player.getTalentTree()));
	}
	if (options.faction != undefined) {
		conditions.push((player: Player<any>) => player.getFaction() == options.faction);
	}
	if (options.customCondition != undefined) {
		conditions.push(options.customCondition);
	}

	return {
		name: name,
		tooltip: options.tooltip || Tooltips.BASIC_BIS_DISCLAIMER,
		group: options.group,
		gear: gear,
		enableWhen: conditions.length > 0 ? (player: Player<any>) => conditions.every(cond => cond(player)) : undefined,
		onLoad: options?.onLoad,
	};
}

export function makePresetTalents(name: string, data: SavedTalents, options?: PresetTalentsOptions): PresetTalents {
	const conditions: Array<(player: Player<any>) => boolean> = [];
	if (options && options.customCondition) {
		conditions.push(options.customCondition);
	}

	return {
		name,
		group: options?.group,
		tooltip: options?.tooltip,
		data,
		enableWhen: conditions.length > 0 ? (player: Player<any>) => conditions.every(cond => cond(player)) : undefined,
	};
}

export const makePresetEpWeights = (name: string, epWeights: Stats, options?: PresetEpWeightsOptions): PresetEpWeights => {
	return makePresetEpWeightHelper(name, epWeights, options || {});
};

const makePresetEpWeightHelper = (name: string, epWeights: Stats, options?: PresetEpWeightsOptions): PresetEpWeights => {
	const conditions: Array<(player: Player<any>) => boolean> = [];
	if (options?.customCondition !== undefined) {
		conditions.push(options.customCondition);
	}

	return {
		name,
		tooltip: options?.tooltip,
		group: options?.group,
		epWeights,
		enableWhen: !!conditions.length ? (player: Player<any>) => conditions.every(cond => cond(player)) : undefined,
		onLoad: options?.onLoad,
	};
};

export function makePresetAPLRotation(name: string, rotationJson: any, options?: PresetRotationOptions): PresetRotation {
	const rotation = SavedRotation.create({
		rotation: APLRotation.fromJson(rotationJson),
	});
	return makePresetRotationHelper(name, rotation, options);
}

export function makePresetSimpleRotation<SpecType extends Spec>(
	name: string,
	spec: SpecType,
	simpleRotation: SpecRotation<SpecType>,
	options?: PresetRotationOptions,
): PresetRotation {
	const rotation = SavedRotation.create({
		rotation: {
			type: APLRotationType.TypeSimple,
			simple: {
				specRotationJson: JSON.stringify(specTypeFunctions[spec].rotationToJson(simpleRotation)),
			},
		},
	});
	return makePresetRotationHelper(name, rotation, options);
}

function makePresetRotationHelper(name: string, rotation: SavedRotation, options?: PresetRotationOptions): PresetRotation {
	const conditions: Array<(player: Player<any>) => boolean> = [];
	if (options?.talentTree != undefined) {
		conditions.push((player: Player<any>) => player.getTalentTree() == options.talentTree);
	}
	if (options?.customCondition != undefined) {
		conditions.push(options.customCondition);
	}

	return {
		name: name,
		group: options?.group,
		tooltip: options?.tooltip,
		rotation: rotation,
		enableWhen: conditions.length > 0 ? (player: Player<any>) => conditions.every(cond => cond(player)) : undefined,
		onLoad: options?.onLoad,
	};
}

export const makePresetEncounter = (name: string, encounter?: PresetEncounter['encounter'] | string, options?: PresetEncounterOptions): PresetEncounter => {
	let healingModel: PresetEncounter['healingModel'] = undefined;
	let tanks: PresetEncounter['tanks'] = undefined;
	let raidBuffs: PresetEncounter['raidBuffs'] = undefined;
	let debuffs: PresetEncounter['debuffs'] = undefined;
	let buffs: PresetEncounter['buffs'] = undefined;
	let consumes: PresetEncounter['consumes'] = undefined;
	if (typeof encounter === 'string') {
		const parsedUrl = IndividualLinkImporter.tryParseUrlLocation(new URL(encounter));
		const settings = parsedUrl?.settings;
		encounter = settings?.encounter;
		healingModel = settings?.player?.healingModel;
		tanks = settings?.tanks;
		raidBuffs = settings?.raidBuffs;
		debuffs = settings?.debuffs;
		buffs = settings?.player?.buffs;
		consumes = settings?.player?.consumes;
	}

	return {
		name,
		encounter,
		tanks,
		healingModel,
		raidBuffs,
		debuffs,
		buffs,
		consumes,
		...options,
	};
};

export const makePresetBuild = (
	name: string,
	{ group, tooltip, gear, talents, rotation, epWeights, encounter, race, level, bonusTalentPoints, professions, options }: PresetBuildOptions,
): PresetBuild => {
	return { name, group, tooltip, gear, talents, rotation, epWeights, encounter, race, level, bonusTalentPoints, professions, options };
};

export type SpecCheckWarning = {
	condition: (player: Player<any>) => boolean;
	message: string;
};

export const makeSpecChangeWarningToast = (checks: SpecCheckWarning[], player: Player<any>, options?: Partial<ToastOptions>) => {
	const messages: string[] = checks.map(({ condition, message }) => condition(player) && message).filter((m): m is string => !!m);
	if (messages.length)
		new Toast({
			variant: 'warning',
			body: (
				<>
					{messages.map(message => (
						<p>{message}</p>
					))}
				</>
			),
			delay: 5000 * messages.length,
			...options,
		});
};
