import { SpecOptions } from '../../../core/proto_utils/utils';
import { IndividualSimUI } from '../../individual_sim_ui';
import { PresetBuild } from '../../preset_utils';
import { APLRotation, APLRotation_Type } from '../../proto/apl';
import { Consumes, Debuffs, Encounter, EquipmentSpec, HealingModel, IndividualBuffs, RaidBuffs, Spec } from '../../proto/common';
import { IndividualSimSettings, SavedTalents } from '../../proto/ui';
import { statNames } from '../../proto_utils/names';
import { normalizeDebuffs } from '../../raid';
import { EventID, TypedEvent } from '../../typed_event';
import { Component } from '../component';
import { ChangeCategory } from '../dirty_settings';
import { presetListTooltip } from '../preset_tree';
import { SavedDataManager } from '../saved_data_manager';

// What a preset configuration can set, in the order we list it.
const BUILD_PARTS: Array<{ name: string; has: (build: PresetBuild) => boolean }> = [
	{ name: 'Gear', has: build => !!build.gear },
	{ name: 'Talents', has: build => !!build.talents },
	{ name: 'Rotation', has: build => !!build.rotation || !!build.rotationType },
	{ name: 'Stat weights', has: build => !!build.epWeights },
	{ name: 'Encounter', has: build => !!build.encounter?.encounter || !!build.encounter?.healingModel || build.encounter?.inFrontOfTarget !== undefined },
	{
		name: 'Buffs, debuffs and consumes',
		has: build => !!build.encounter?.buffs || !!build.encounter?.raidBuffs || !!build.encounter?.debuffs || !!build.encounter?.consumes,
	},
	{ name: 'Race', has: build => build.race !== undefined },
	{ name: 'Level and extra talent points', has: build => build.level !== undefined || build.bonusTalentPoints !== undefined },
	{ name: 'Professions', has: build => !!build.professions },
	{ name: 'Class options', has: build => !!build.options },
];

// A configuration in the list: one of the spec's preset builds, which sets the parts it
// has, or one we saved, which holds the whole character.
type SavedConfiguration = { build: PresetBuild; settings?: undefined } | { build?: undefined; settings: IndividualSimSettings };

// The preset configurations, in the sidebar under the Simulate button. Each one sets up a
// whole character at once (gear, talents, rotation, encounter and so on), so we show them
// on every tab.
export class PresetConfigurationPicker extends Component {
	readonly simUI: IndividualSimUI<Spec>;
	readonly builds: Array<PresetBuild>;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		super(parentElem, 'preset-configuration-picker-root');

		this.simUI = simUI;
		this.builds = this.simUI.individualConfig.presets.builds ?? [];

		// A configuration we save holds every part, so with no builds we list them all.
		const parts = this.builds.length ? BUILD_PARTS.filter(part => this.builds.some(part.has)) : BUILD_PARTS;
		const manager = new SavedDataManager<IndividualSimUI<Spec>, SavedConfiguration>(this.rootElem, simUI, {
			label: 'Configuration',
			header: {
				title: 'Preset Configurations',
				tooltip: presetListTooltip(
					'A preset configuration sets up the whole character at once:',
					parts.map(part => part.name),
				),
			},
			storageKey: simUI.getStorageKey('__savedPresetConfigurations__'),
			changeEmitters: [simUI.player.changeEmitter, simUI.sim.settingsChangeEmitter, simUI.sim.raid.changeEmitter, simUI.sim.encounter.changeEmitter],
			equals: (a, b) => (a.build ? !this.changedParts(a.build).length : IndividualSimSettings.equals(a.settings!, b.settings!)),
			getData: () => ({ settings: this.currentSettings() }),
			setData: (eventID, simUI, data) => {
				if (data.build) {
					this.applyBuild(eventID, data.build);
				} else {
					// We keep the sim settings we have now, like the iterations and the phase.
					const settings = IndividualSimSettings.clone(data.settings!);
					settings.settings = simUI.sim.toProto();
					simUI.fromProto(eventID, settings);
				}
			},
			toJson: data => IndividualSimSettings.toJson(data.settings!),
			fromJson: obj => ({ settings: IndividualSimSettings.fromJson(obj) }),
			listChanges: (data, changes) => (data.build ? this.listChanges(data.build, changes) : changes),
		});
		this.addOnDisposeCallback(() => manager.dispose());

		this.simUI.sim.waitForInit().then(() => {
			manager.loadUserData();
			this.builds.forEach(build =>
				manager.addSavedData({
					name: build.name,
					tooltip: build.tooltip,
					group: build.group,
					isPreset: true,
					data: { build },
				}),
			);
		});
	}

	// The whole character as we save it in a custom configuration. The sim settings, like the
	// iterations and the phase, stay out, so changing them doesn't change the configuration.
	private currentSettings(): IndividualSimSettings {
		const settings = this.simUI.toProto();
		settings.settings = undefined;
		return settings;
	}

	private applyBuild(
		eventID: EventID,
		{ gear, rotation, rotationType, talents, epWeights, encounter, race, level, bonusTalentPoints, professions, options }: PresetBuild,
	) {
		TypedEvent.freezeAllAndDo(() => {
			if (gear) this.simUI.player.setGear(eventID, this.simUI.sim.db.lookupEquipmentSpec(gear.gear));
			if (race) this.simUI.player.setRace(eventID, race);
			if (level !== undefined) this.simUI.player.setLevel(eventID, level);
			if (bonusTalentPoints !== undefined) this.simUI.player.setBonusTalentPoints(eventID, bonusTalentPoints);
			if (professions) this.simUI.player.setProfessions(eventID, professions);
			if (talents) this.simUI.player.setTalentsString(eventID, talents.data.talentsString);
			if (rotationType) {
				this.simUI.player.aplRotation.type = rotationType;
				this.simUI.player.rotationChangeEmitter.emit(eventID);
			} else if (rotation?.rotation.rotation) {
				this.simUI.player.setAplRotation(eventID, rotation.rotation.rotation);
			}
			if (epWeights) this.simUI.player.setEpWeights(eventID, epWeights.epWeights);
			if (encounter) {
				if (encounter.encounter) this.simUI.sim.encounter.fromProto(eventID, encounter.encounter);
				if (encounter.healingModel) this.simUI.player.setHealingModel(eventID, encounter.healingModel);
				if (encounter.tanks) this.simUI.sim.raid.setTanks(eventID, encounter.tanks);
				if (encounter.buffs) this.simUI.player.setBuffs(eventID, encounter.buffs);
				if (encounter.debuffs) this.simUI.sim.raid.setDebuffs(eventID, encounter.debuffs);
				if (encounter.raidBuffs) this.simUI.sim.raid.setBuffs(eventID, encounter.raidBuffs);
				if (encounter.consumes) this.simUI.player.setConsumes(eventID, encounter.consumes);
				if (encounter.inFrontOfTarget !== undefined) this.simUI.player.setInFrontOfTarget(eventID, encounter.inFrontOfTarget);
			}
			if (options) {
				this.simUI.player.setSpecOptions(eventID, {
					...this.simUI.player.getSpecOptions(),
					...options,
				});
			}
		});
	}

	// What differs from the build, for its tooltip: each setting on the tabs that changed,
	// and the stat weights, which we compare here because their dialog is closed. When none
	// of that changed (a setting without a field, like the tanks) we name the parts that
	// differ instead.
	private listChanges(build: PresetBuild, changes: ChangeCategory[]): ChangeCategory[] {
		const parts = this.changedParts(build);
		const categories = changes.filter(category => category.name !== 'Stat Weights');
		if (parts.includes('stat weights') && build.epWeights) {
			const preset = build.epWeights.epWeights;
			const current = this.simUI.player.getEpWeights();
			const format = (value: number) => String(Math.round(value * 100) / 100);
			categories.push({
				name: 'Stat Weights',
				lines: Array.from(statNames)
					.filter(([stat]) => preset.getStat(stat) !== current.getStat(stat))
					.map(([stat, name]) => `${name}: ${format(preset.getStat(stat))} → ${format(current.getStat(stat))}`),
			});
		}
		return categories.length ? categories : parts.map(part => ({ name: part.charAt(0).toUpperCase() + part.slice(1), lines: [] }));
	}

	// The parts of the build that differ from the current settings, in lower case. An empty
	// list means the build is loaded as is.
	private changedParts({
		gear,
		rotation,
		rotationType,
		talents,
		epWeights,
		encounter,
		race,
		level,
		bonusTalentPoints,
		professions,
		options,
	}: PresetBuild): string[] {
		const player = this.simUI.player;
		const hasGear = gear ? EquipmentSpec.equals(gear.gear, player.getGear().asSpec()) : true;
		const hasRace = typeof race === 'number' ? race === player.getRace() : true;
		const hasLevel = level !== undefined ? level === player.getLevel() : true;
		const hasBonusTalentPoints = bonusTalentPoints !== undefined ? bonusTalentPoints === player.getBonusTalentPoints() : true;
		const hasProfessions = professions ? professions[0] === player.getProfession1() && professions[1] === player.getProfession2() : true;
		const hasTalents = talents
			? SavedTalents.equals(
					talents.data,
					SavedTalents.create({
						talentsString: player.getTalentsString(),
					}),
			  )
			: true;
		let hasRotation = true;
		if (rotationType) {
			hasRotation = rotationType === player.getRotationType();
		} else if (rotation) {
			const activeRotation = player.getResolvedAplRotation();
			// Ensure that the auto rotation can be matched with a preset
			if (activeRotation.type === APLRotation_Type.TypeAuto) activeRotation.type = APLRotation_Type.TypeAPL;
			if (rotation.rotation?.rotation?.type === APLRotation_Type.TypeSimple && rotation.rotation.rotation?.simple?.specRotationJson) {
				hasRotation = player.specTypeFunctions.rotationEquals(
					player.specTypeFunctions.rotationFromJson(JSON.parse(rotation.rotation.rotation.simple.specRotationJson)),
					player.getSimpleRotation(),
				);
			} else {
				hasRotation = APLRotation.equals(rotation.rotation.rotation, activeRotation);
			}
		}
		const hasEpWeights = epWeights ? player.getEpWeights().equals(epWeights.epWeights) : true;
		const hasEncounter = encounter?.encounter ? Encounter.equals(encounter.encounter, this.simUI.sim.encounter.toProto()) : true;
		const hasHealingModel = encounter?.healingModel ? HealingModel.equals(encounter.healingModel, player.getHealingModel()) : true;
		const hasInFrontOfTarget = encounter?.inFrontOfTarget !== undefined ? encounter.inFrontOfTarget === player.getInFrontOfTarget() : true;
		const hasOptions = options ? this.containsAllFields(player.getSpecOptions(), options) : true;
		const raid = this.simUI.sim.raid;
		const hasTanks = encounter?.tanks ? JSON.stringify(encounter.tanks) === JSON.stringify(raid.getTanks()) : true;
		const hasBuffs =
			(encounter?.buffs ? IndividualBuffs.equals(encounter.buffs, player.getBuffs()) : true) &&
			(encounter?.raidBuffs ? RaidBuffs.equals(encounter.raidBuffs, raid.getBuffs()) : true);
		// The raid turns some debuffs off, see normalizeDebuffs, so a preset that has them still matches.
		const hasDebuffs = encounter?.debuffs ? Debuffs.equals(normalizeDebuffs(encounter.debuffs, this.simUI.sim.getRuleset()), raid.getDebuffs()) : true;
		const hasConsumes = encounter?.consumes ? Consumes.equals(encounter.consumes, player.getConsumes()) : true;

		const changed: Array<[boolean, string]> = [
			[hasGear, 'gear'],
			[hasTalents, 'talents'],
			[hasRotation, 'rotation'],
			[hasEpWeights, 'stat weights'],
			[hasEncounter && hasHealingModel && hasInFrontOfTarget && hasTanks, 'encounter'],
			[hasBuffs, 'buffs'],
			[hasDebuffs, 'debuffs'],
			[hasConsumes, 'consumes'],
			[hasRace, 'race'],
			[hasLevel && hasBonusTalentPoints, 'level'],
			[hasProfessions, 'professions'],
			[hasOptions, 'class options'],
		];
		return changed.filter(([matches]) => !matches).map(([_, name]) => name);
	}

	// A field can be a message of its own, like an Enhancement shaman's starting totems, so we
	// compare those by their fields.
	private containsAllFields<T extends Spec>(full: SpecOptions<T>, partial: Partial<SpecOptions<T>>): boolean {
		return Object.keys(partial).every(key => key in full && sameValue(full[key as keyof SpecOptions<T>], partial[key as keyof SpecOptions<T>]));
	}
}

function sameValue(a: unknown, b: unknown): boolean {
	if (a === b) {
		return true;
	}
	if (typeof a !== 'object' || typeof b !== 'object' || a === null || b === null) {
		return false;
	}
	const keys = new Set([...Object.keys(a), ...Object.keys(b)]);
	return [...keys].every(key => sameValue((a as Record<string, unknown>)[key], (b as Record<string, unknown>)[key]));
}
