import { SpecOptions } from '../../../core/proto_utils/utils';
import { IndividualSimUI } from '../../individual_sim_ui';
import { PresetBuild } from '../../preset_utils';
import { APLRotation, APLRotation_Type } from '../../proto/apl';
import { Consumes, Debuffs, Encounter, EquipmentSpec, HealingModel, IndividualBuffs, RaidBuffs, Spec } from '../../proto/common';
import { SavedTalents } from '../../proto/ui';
import { statNames } from '../../proto_utils/names';
import { TypedEvent } from '../../typed_event';
import { Component } from '../component';
import { ContentBlock } from '../content_block';
import { ChangeCategory, dirtySettings, PresetSource } from '../dirty_settings';
import { presetListTooltip, PresetTree } from '../preset_tree';

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

// The preset configurations, in the sidebar above the Simulate button. Each one sets up a
// whole character at once (gear, talents, rotation, encounter and so on), so we show them
// on every tab.
export class PresetConfigurationPicker extends Component {
	readonly simUI: IndividualSimUI<Spec>;
	readonly builds: Array<PresetBuild>;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		super(parentElem, 'preset-configuration-picker-root');
		this.rootElem.classList.add('saved-data-manager-root');

		this.simUI = simUI;
		this.builds = this.simUI.individualConfig.presets.builds ?? [];

		if (!this.builds.length) {
			this.rootElem.classList.add('hide');
			return;
		}

		const parts = BUILD_PARTS.filter(part => this.builds.some(part.has));
		const contentBlock = new ContentBlock(this.rootElem, 'saved-data', {
			header: {
				title: 'Preset Configurations',
				tooltip: presetListTooltip(
					'A preset configuration sets up the whole character at once:',
					parts.map(part => part.name),
				),
			},
		});

		const tree = new PresetTree(
			this.simUI.getStorageKey('__presetConfigurationsOpenFolders__'),
			this.simUI.getStorageKey('__selectedPresetConfiguration__'),
		);
		const container = (
			<div className="saved-data-container">
				<div className="saved-data-presets">{tree.rootElem}</div>
			</div>
		);

		this.simUI.sim.waitForInit().then(() => {
			const builds = new Map<HTMLElement, PresetBuild>();
			const source: PresetSource = {
				selected: () => {
					const build = builds.get(tree.selected()?.item as HTMLElement);
					return build && { apply: () => this.applyBuild(build) };
				},
			};
			this.builds.forEach(build => {
				const item = PresetTree.makeItem(build.name);
				builds.set(item, build);
				tree.onClick(item, () => this.applyBuild(build));
				tree.attachTooltip(item, build.tooltip, 'right', () => this.listChanges(build, source));
				tree.track(item, { name: build.name, matches: () => !this.changedParts(build).length });
				tree.add(item, build.group);
			});

			TypedEvent.onAny([
				this.simUI.player.changeEmitter,
				this.simUI.sim.settingsChangeEmitter,
				this.simUI.sim.raid.changeEmitter,
				this.simUI.sim.encounter.changeEmitter,
			]).on(() => tree.refresh());
			tree.refresh();

			dirtySettings.addSource(source);
			contentBlock.bodyElement.replaceChildren(container);
		});
	}

	private applyBuild({ gear, rotation, rotationType, talents, epWeights, encounter, race, level, bonusTalentPoints, professions, options }: PresetBuild) {
		const eventID = TypedEvent.nextEventID();
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
	private listChanges(build: PresetBuild, source: PresetSource): ChangeCategory[] {
		const parts = this.changedParts(build);
		const categories = dirtySettings.changesFor(source).filter(category => category.name !== 'Stat Weights');
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
		const hasDebuffs = encounter?.debuffs ? Debuffs.equals(encounter.debuffs, raid.getDebuffs()) : true;
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

	private containsAllFields<T extends Spec>(full: SpecOptions<T>, partial: Partial<SpecOptions<T>>): boolean {
		return Object.keys(partial).every(key => key in full && full[key as keyof SpecOptions<T>] === partial[key as keyof SpecOptions<T>]);
	}
}
