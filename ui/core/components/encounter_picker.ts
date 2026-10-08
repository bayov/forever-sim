import tippy from 'tippy.js';

import { BooleanPicker } from '../components/boolean_picker.js';
import { EnumPicker } from '../components/enum_picker.js';
import { ListItemPickerConfig, ListPicker } from '../components/list_picker.js';
import { NumberPicker } from '../components/number_picker.js';
import { Encounter } from '../encounter.js';
import { IndividualSimUI } from '../individual_sim_ui.js';
import { InputType, MobType, SpellSchool, Stat, Target, Target as TargetProto, TargetInput } from '../proto/common.js';
import { statNames } from '../proto_utils/names.js';
import { Stats } from '../proto_utils/stats.js';
import { isHealingSpec } from '../proto_utils/utils.js';
import { Raid } from '../raid.js';
import { SimUI } from '../sim_ui.js';
import { EventID, TypedEvent } from '../typed_event.js';
import { randomUUID } from '../utils.js';
import { BaseModal } from './base_modal.js';
import { Component } from './component.js';
import { dirtySettings } from './dirty_settings.js';
import { Input } from './input.js';

export interface EncounterPickerConfig {
	// We show the execute fields for every spec now, so we no longer read this. The specs still
	// set it.
	showExecuteProportion: boolean;
}

// The fight settings: its length, PvP, the execute phases and the targets.
//
// Each target is a row with its mob type's icon, its name and a short summary. We edit a
// target in a modal with only that target's settings, add one with the button under the list
// and remove one with the button on its row.
export class EncounterPicker extends Component {
	constructor(parent: HTMLElement, modEncounter: Encounter, _config: EncounterPickerConfig, simUI: SimUI) {
		super(parent, 'encounter-picker-root');

		addEncounterFieldPickers(this.rootElem, modEncounter);
		if (!simUI.isIndividualSim()) {
			new BooleanPicker<Encounter>(this.rootElem, modEncounter, {
				id: 'encounter-use-health',
				label: 'Use Health',
				labelTooltip: 'Uses a damage limit in place of a duration limit. Damage limit is equal to sum of all targets health.',
				inline: true,
				changedEvent: (encounter: Encounter) => encounter.changeEmitter,
				getValue: (encounter: Encounter) => encounter.getUseHealth(),
				setValue: (eventID: EventID, encounter: Encounter, newValue: boolean) => {
					encounter.setUseHealth(eventID, newValue);
				},
			});
		}

		// Need to wait so that the encounter and target presets will be loaded.
		modEncounter.sim.waitForInit().then(() => {
			const presetEncounters = modEncounter.sim.db.getAllPresetEncounters();
			if (presetEncounters.length) {
				new EnumPicker<Encounter>(this.rootElem, modEncounter, {
					id: 'encounter-preset-encouter',
					extraCssClasses: ['encounter-preset-picker'],
					label: 'Encounter',
					values: [{ name: 'Custom', value: -1 }].concat(presetEncounters.map((pe, i) => ({ name: pe.path, value: i }))),
					changedEvent: (encounter: Encounter) => encounter.changeEmitter,
					getValue: (encounter: Encounter) => presetEncounters.findIndex(pe => encounter.matchesPreset(pe)),
					setValue: (eventID: EventID, encounter: Encounter, newValue: number) => {
						if (newValue != -1) {
							encounter.applyPreset(eventID, presetEncounters[newValue]);
						}
					},
				});
			}

			if (simUI.isIndividualSim() && isHealingSpec((simUI as IndividualSimUI<any>).player.spec)) {
				new NumberPicker(this.rootElem, simUI.sim.raid, {
					id: 'encounter-num-allies',
					label: 'Num Allies',
					labelTooltip: 'Number of allied players in the raid.',
					changedEvent: (raid: Raid) => raid.targetDummiesChangeEmitter,
					getValue: (raid: Raid) => raid.getTargetDummies(),
					setValue: (eventID: EventID, raid: Raid, newValue: number) => {
						raid.setTargetDummies(eventID, newValue);
					},
				});
			}

			// A target keeps the inputs of the AI it had when we saved it. When its AI has other
			// inputs now (a custom AI, or one that changed), we give the target the AI's inputs.
			const presetTargets = modEncounter.sim.db.getAllPresetTargets();
			let inputsChanged = false;
			modEncounter.targets.forEach(target => {
				const targetInputs = presetTargets.find(pe => target.id == pe.target?.id)?.target?.targetInputs || [];
				if (targetInputs.length != target.targetInputs.length || target.targetInputs.some((ti, i) => ti.label != targetInputs[i].label)) {
					target.targetInputs = targetInputs.map(ti => TargetInput.clone(ti));
					inputsChanged = true;
				}
			});
			if (inputsChanged) {
				modEncounter.targetsChangeEmitter.emit(TypedEvent.nextEventID());
			}

			new TargetList(this.rootElem, modEncounter, simUI);
		});
	}
}

// The targets, one row each, and a button to add one. There's always at least one target.
class TargetList extends Component {
	constructor(parent: HTMLElement, encounter: Encounter, simUI: SimUI) {
		super(parent, 'encounter-targets');

		this.rootElem.innerHTML = `
			<label class="form-label">Targets</label>
			<div class="encounter-target-list"></div>
			<button class="encounter-target-add btn btn-outline-primary">
				<i class="fas fa-plus me-1"></i>Add target
			</button>
		`;
		const listElem = this.rootElem.querySelector('.encounter-target-list') as HTMLElement;

		// We mark the Targets label when the preset has another number of targets, and a target's
		// row when the preset has another target in its place, see dirtySettings.
		const untrackList = dirtySettings.track({ elem: this.rootElem, read: () => encounter.targets.length, name: () => 'Targets' });
		let untrackRows: Array<() => void> = [];
		this.addOnDisposeCallback(() => {
			untrackList();
			untrackRows.forEach(untrack => untrack());
		});

		const presetTargets = encounter.sim.db.getAllPresetTargets();
		const render = () => {
			untrackRows.forEach(untrack => untrack());
			const rows = encounter.targets.map((target, index) => buildTargetRow(target, index));
			untrackRows = rows.map((row, index) =>
				dirtySettings.track({ elem: row, read: () => encounter.targets[index], name: () => `Target ${index + 1}`, format: () => '' }),
			);
			listElem.replaceChildren(...rows);
		};

		const buildTargetRow = (target: TargetProto, index: number): HTMLElement => {
			const preset = presetTargets.find(pe => equalTargetsIgnoreInputs(target, pe.target));
			const mobType = mobTypeEnumValues.find(mt => mt.value == target.mobType) ?? mobTypeEnumValues[0];
			const summary = [`Level ${target.level}`, mobType.name, `${target.stats[Stat.StatArmor]} Armor`];
			if (target.stats[Stat.StatHealth]) summary.push(`${target.stats[Stat.StatHealth]} Health`);

			const row = document.createElement('div');
			row.classList.add('encounter-target-row');
			row.innerHTML = `
				<img class="encounter-target-icon" src="https://wow.zamimg.com/images/wow/icons/large/${mobType.icon}.jpg" />
				<div class="encounter-target-info">
					<div class="encounter-target-name"></div>
					<div class="encounter-target-summary"></div>
				</div>
				<button class="encounter-target-edit btn btn-link" aria-label="Edit target"><i class="fas fa-pen"></i></button>
				<button class="encounter-target-remove btn btn-link link-danger" aria-label="Remove target"><i class="fas fa-times"></i></button>
			`;
			(row.querySelector('.encounter-target-name') as HTMLElement).textContent = preset?.path.split('/').pop() ?? 'Custom';
			(row.querySelector('.encounter-target-summary') as HTMLElement).textContent = summary.join(' \u00b7 ');

			const icon = row.querySelector('.encounter-target-icon') as HTMLElement;
			tippy(icon, { content: mobType.name });
			const editButton = row.querySelector('.encounter-target-edit') as HTMLElement;
			tippy(editButton, { content: 'Edit target' });
			editButton.addEventListener('click', () => new TargetModal(simUI.rootElem, encounter, index).open());

			const removeButton = row.querySelector('.encounter-target-remove') as HTMLElement;
			if (encounter.targets.length > 1) {
				tippy(removeButton, { content: 'Remove target' });
				removeButton.addEventListener('click', () => {
					encounter.targets = encounter.targets.filter((_, i) => i != index);
					encounter.targetsChangeEmitter.emit(TypedEvent.nextEventID());
				});
			} else {
				removeButton.remove();
			}
			return row;
		};

		// A new target starts as a copy of the last one, since we usually add more of the same mob.
		this.rootElem.querySelector('.encounter-target-add')!.addEventListener('click', () => {
			const last = encounter.targets[encounter.targets.length - 1];
			encounter.targets = [...encounter.targets, TargetProto.clone(last ?? Encounter.getDefaultTarget(simUI.sim).target!)];
			encounter.targetsChangeEmitter.emit(TypedEvent.nextEventID());
		});

		render();
		const event = encounter.targetsChangeEmitter.on(render);
		this.addOnDisposeCallback(() => event.dispose());
	}
}

// One target's settings. We throw the modal away when it closes, so its index never goes
// stale after we remove a target.
class TargetModal extends BaseModal {
	constructor(parent: HTMLElement, encounter: Encounter, index: number) {
		super(parent, 'target-picker-modal', { title: `Target ${index + 1}`, disposeOnClose: true });
		new TargetPicker(this.body, encounter, index, {
			id: `encounter-target-${index}`,
			changedEvent: (encounter: Encounter) => encounter.targetsChangeEmitter,
			getValue: (encounter: Encounter) => encounter.targets[index],
			setValue: (eventID: EventID, encounter: Encounter, newValue: TargetProto) => {
				encounter.targets[index] = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});
	}
}

class TargetPicker extends Input<Encounter, TargetProto> {
	private readonly encounter: Encounter;
	private readonly targetIndex: number;

	private readonly aiPicker: Input<null, number>;
	private readonly levelPicker: Input<null, number>;
	private readonly mobTypePicker: Input<null, number>;
	private readonly tankIndexPicker: Input<null, number>;
	private readonly statPickers: Array<Input<null, number>>;
	private readonly swingSpeedPicker: Input<null, number>;
	private readonly minBaseDamagePicker: Input<null, number>;
	private readonly dualWieldPicker: Input<null, boolean>;
	private readonly dwMissPenaltyPicker: Input<null, boolean>;
	private readonly parryHastePicker: Input<null, boolean>;
	private readonly spellSchoolPicker: Input<null, number>;
	private readonly damageSpreadPicker: Input<null, number>;
	private readonly targetInputPickers: ListPicker<Encounter, TargetInput>;

	private getTarget(): TargetProto {
		return this.encounter.targets[this.targetIndex] || Target.create();
	}

	constructor(parent: HTMLElement, encounter: Encounter, targetIndex: number, config: ListItemPickerConfig<Encounter, TargetProto>) {
		super(parent, 'target-picker-root', encounter, config);
		this.encounter = encounter;
		this.targetIndex = targetIndex;

		this.rootElem.innerHTML = `
			<div class="picker-group target-picker-section target-picker-section1"></div>
			<div class="picker-group target-picker-section target-picker-section2"></div>
			<div class="picker-group target-picker-section target-picker-section3 threat-metrics"></div>
		`;

		const section1 = this.rootElem.getElementsByClassName('target-picker-section1')[0] as HTMLElement;
		const section2 = this.rootElem.getElementsByClassName('target-picker-section2')[0] as HTMLElement;
		const section3 = this.rootElem.getElementsByClassName('target-picker-section3')[0] as HTMLElement;

		const presetTargets = encounter.sim.db.getAllPresetTargets();
		new EnumPicker<null>(section1, null, {
			id: 'encounter-npc-picker',
			extraCssClasses: ['npc-picker'],
			label: 'NPC',
			labelTooltip: 'Selects a preset NPC configuration.',
			values: [{ name: 'Custom', value: -1 }].concat(
				presetTargets.map((pe, i) => {
					return {
						name: pe.path,
						value: i,
					};
				}),
			),
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => presetTargets.findIndex(pe => equalTargetsIgnoreInputs(this.getTarget(), pe.target)),
			setValue: (eventID: EventID, _: null, newValue: number) => {
				if (newValue != -1) {
					encounter.applyPresetTarget(eventID, presetTargets[newValue], this.targetIndex);
					encounter.targetsChangeEmitter.emit(eventID);
				}
			},
		});

		this.aiPicker = new EnumPicker<null>(section1, null, {
			id: 'encounter-ai-picker',
			extraCssClasses: ['ai-picker'],
			label: 'AI',
			labelTooltip: `
				<p>Determines the target\'s ability rotation.</p>
				<p>Note that most rotations are not yet implemented.</p>
			`,
			values: [{ name: 'None', value: 0 }].concat(
				presetTargets.map(pe => {
					return {
						name: pe.path,
						value: pe.target!.id,
					};
				}),
			),
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().id,
			setValue: (eventID: EventID, _: null, newValue: number) => {
				const target = this.getTarget();
				target.id = newValue;

				// Transfer Target Inputs from the AI of the selected target
				target.targetInputs = (presetTargets.find(pe => target.id == pe.target?.id)?.target?.targetInputs || []).map(ti => TargetInput.clone(ti));

				encounter.targetsChangeEmitter.emit(eventID);
			},
		});

		this.levelPicker = new EnumPicker<null>(section1, null, {
			id: 'encounter-level-picker',
			label: 'Level',
			values: [
				{ name: '63', value: 63 },
				{ name: '62', value: 62 },
				{ name: '61', value: 61 },
				{ name: '60', value: 60 },
				{ name: '53', value: 53 },
				{ name: '52', value: 52 },
				{ name: '51', value: 51 },
				{ name: '50', value: 50 },
				{ name: '43', value: 43 },
				{ name: '42', value: 42 },
				{ name: '41', value: 41 },
				{ name: '40', value: 40 },
				{ name: '28', value: 28 },
				{ name: '27', value: 27 },
				{ name: '26', value: 26 },
				{ name: '25', value: 25 },
			],
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().level,
			setValue: (eventID: EventID, _: null, newValue: number) => {
				this.getTarget().level = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});
		this.mobTypePicker = new EnumPicker(section1, null, {
			id: 'encounter-mob-type',
			label: 'Mob Type',
			values: mobTypeEnumValues,
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().mobType,
			setValue: (eventID: EventID, _: null, newValue: number) => {
				this.getTarget().mobType = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});
		this.tankIndexPicker = new EnumPicker<null>(section1, null, {
			id: 'target-picker-tanked-by',
			extraCssClasses: ['threat-metrics'],
			label: 'Tanked By',
			labelTooltip:
				'Determines which player in the raid this enemy will attack. If no player is assigned to the specified tank slot, this enemy will not attack.',
			values: [
				{ name: 'None', value: -1 },
				{ name: 'Main Tank', value: 0 },
				{ name: 'Tank 2', value: 1 },
				{ name: 'Tank 3', value: 2 },
				{ name: 'Tank 4', value: 3 },
			],
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().tankIndex,
			setValue: (eventID: EventID, _: null, newValue: number) => {
				this.getTarget().tankIndex = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});

		this.targetInputPickers = makeTargetInputsPicker(section1, encounter, this.targetIndex);

		this.statPickers = ALL_TARGET_STATS.map(statData => {
			const stat = statData.stat;
			return new NumberPicker(section2, null, {
				id: `target-picker-stats-${statData.stat}`,
				inline: true,
				extraCssClasses: statData.extraCssClasses,
				label: statNames.get(stat),
				labelTooltip: statData.tooltip,
				changedEvent: () => encounter.targetsChangeEmitter,
				getValue: () => this.getTarget().stats[stat],
				setValue: (eventID: EventID, _: null, newValue: number) => {
					this.getTarget().stats[stat] = newValue;
					encounter.targetsChangeEmitter.emit(eventID);
				},
			});
		});

		this.swingSpeedPicker = new NumberPicker(section3, null, {
			id: 'target-picker-swing-speed',
			label: 'Swing Speed',
			labelTooltip: 'Time in seconds between auto attacks. Set to 0 to disable auto attacks.',
			float: true,
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().swingSpeed,
			setValue: (eventID: EventID, _: null, newValue: number) => {
				this.getTarget().swingSpeed = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});
		this.minBaseDamagePicker = new NumberPicker(section3, null, {
			id: 'target-picker-min-base-damage',
			label: 'Min Base Damage',
			labelTooltip: 'Base damage for auto attacks, i.e. lowest roll with 0 AP against a 0-armor Player.',
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().minBaseDamage,
			setValue: (eventID: EventID, _: null, newValue: number) => {
				this.getTarget().minBaseDamage = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});
		this.damageSpreadPicker = new NumberPicker(section3, null, {
			id: 'target-picker-damage-spread',
			label: 'Damage Spread',
			labelTooltip: 'Fractional spread between the minimum and maximum auto-attack damage from this enemy at 0 Attack Power.',
			float: true,
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().damageSpread,
			setValue: (eventID: EventID, _: null, newValue: number) => {
				this.getTarget().damageSpread = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});
		this.dualWieldPicker = new BooleanPicker(section3, null, {
			id: 'target-picker-dual-wield',
			label: 'Dual Wield',
			labelTooltip: 'Uses 2 separate weapons to attack.',
			inline: true,
			reverse: true,
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().dualWield,
			setValue: (eventID: EventID, _: null, newValue: boolean) => {
				this.getTarget().dualWield = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});
		this.dwMissPenaltyPicker = new BooleanPicker(section3, null, {
			id: 'target-picker-dw-miss-penalty',
			label: 'DW Miss Penalty',
			labelTooltip:
				'Enables the Dual Wield Miss Penalty (+19% chance to miss) if dual wielding. Bosses in Hyjal/BT/SWP usually have this disabled to stop tanks from avoidance stacking.',
			inline: true,
			reverse: true,
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().dualWieldPenalty,
			setValue: (eventID: EventID, _: null, newValue: boolean) => {
				this.getTarget().dualWieldPenalty = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
			enableWhen: () => this.getTarget().dualWield,
		});
		this.parryHastePicker = new BooleanPicker(section3, null, {
			id: 'target-picker-parry-haste',
			label: 'Parry Haste',
			labelTooltip: 'Whether this enemy will gain parry haste when parrying attacks.',
			inline: true,
			reverse: true,
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().parryHaste,
			setValue: (eventID: EventID, _: null, newValue: boolean) => {
				this.getTarget().parryHaste = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});
		this.spellSchoolPicker = new EnumPicker<null>(section3, null, {
			id: 'target-picker-spell-school',
			label: 'Spell School',
			labelTooltip: 'Type of damage caused by auto attacks. This is usually Physical, but some enemies have elemental attacks.',
			values: [
				{ name: 'Physical', value: SpellSchool.SpellSchoolPhysical },
				{ name: 'Arcane', value: SpellSchool.SpellSchoolArcane },
				{ name: 'Fire', value: SpellSchool.SpellSchoolFire },
				{ name: 'Frost', value: SpellSchool.SpellSchoolFrost },
				{ name: 'Holy', value: SpellSchool.SpellSchoolHoly },
				{ name: 'Nature', value: SpellSchool.SpellSchoolNature },
				{ name: 'Shadow', value: SpellSchool.SpellSchoolShadow },
			],
			changedEvent: () => encounter.targetsChangeEmitter,
			getValue: () => this.getTarget().spellSchool,
			setValue: (eventID: EventID, _: null, newValue: number) => {
				this.getTarget().spellSchool = newValue;
				encounter.targetsChangeEmitter.emit(eventID);
			},
		});

		this.init();
	}

	// We return the root, not null. With null the first change would dispose the picker, as if it
	// had left the page, and that skips the NPC picker's update for that change.
	getInputElem(): HTMLElement | null {
		return this.rootElem;
	}
	getInputValue(): TargetProto {
		return TargetProto.create({
			id: this.aiPicker.getInputValue(),
			level: this.levelPicker.getInputValue(),
			mobType: this.mobTypePicker.getInputValue(),
			tankIndex: this.tankIndexPicker.getInputValue(),
			swingSpeed: this.swingSpeedPicker.getInputValue(),
			minBaseDamage: this.minBaseDamagePicker.getInputValue(),
			dualWield: this.dualWieldPicker.getInputValue(),
			dualWieldPenalty: this.dwMissPenaltyPicker.getInputValue(),
			parryHaste: this.parryHastePicker.getInputValue(),
			spellSchool: this.spellSchoolPicker.getInputValue(),
			damageSpread: this.damageSpreadPicker.getInputValue(),
			stats: this.statPickers
				.map(picker => picker.getInputValue())
				.map((statValue, i) => new Stats().withStat(ALL_TARGET_STATS[i].stat, statValue))
				.reduce((totalStats, curStats) => totalStats.add(curStats))
				.asArray(),
			targetInputs: this.targetInputPickers.getInputValue(),
		});
	}
	setInputValue(newValue: TargetProto) {
		if (!newValue) {
			return;
		}
		this.aiPicker.setInputValue(newValue.id);
		this.levelPicker.setInputValue(newValue.level);
		this.mobTypePicker.setInputValue(newValue.mobType);
		this.tankIndexPicker.setInputValue(newValue.tankIndex);
		this.swingSpeedPicker.setInputValue(newValue.swingSpeed);
		this.minBaseDamagePicker.setInputValue(newValue.minBaseDamage);
		this.dualWieldPicker.setInputValue(newValue.dualWield);
		this.dwMissPenaltyPicker.setInputValue(newValue.dualWieldPenalty);
		this.parryHastePicker.setInputValue(newValue.parryHaste);
		this.spellSchoolPicker.setInputValue(newValue.spellSchool);
		this.damageSpreadPicker.setInputValue(newValue.damageSpread);
		ALL_TARGET_STATS.forEach((statData, i) => this.statPickers[i].setInputValue(newValue.stats[statData.stat]));
		this.targetInputPickers.setInputValue(newValue.targetInputs);
	}
}

class TargetInputPicker extends Input<Encounter, TargetInput> {
	private readonly encounter: Encounter;
	private readonly targetIndex: number;
	private readonly targetInputIndex: number;

	private boolPicker: Input<null, boolean> | null;
	private numberPicker: Input<null, number> | null;
	private enumPicker: EnumPicker<null> | null;

	private getTargetInput(): TargetInput {
		return this.encounter.targets[this.targetIndex].targetInputs[this.targetInputIndex] || TargetInput.create();
	}

	private clearPickers() {
		if (this.boolPicker) {
			this.boolPicker.rootElem.remove();
			this.boolPicker = null;
		}
		if (this.numberPicker) {
			this.numberPicker.rootElem.remove();
			this.numberPicker = null;
		}
		if (this.enumPicker) {
			this.enumPicker.rootElem.remove();
			this.enumPicker = null;
		}
	}

	constructor(
		parent: HTMLElement,
		encounter: Encounter,
		targetIndex: number,
		targetInputIndex: number,
		config: ListItemPickerConfig<Encounter, TargetInput>,
	) {
		super(parent, 'target-input-picker-root', encounter, config);
		this.encounter = encounter;
		this.targetIndex = targetIndex;
		this.targetInputIndex = targetInputIndex;

		this.boolPicker = null;
		this.numberPicker = null;
		this.enumPicker = null;
		this.init();
	}

	getInputElem(): HTMLElement | null {
		return this.rootElem;
	}
	getInputValue(): TargetInput {
		return TargetInput.create({
			boolValue: this.boolPicker ? this.boolPicker.getInputValue() : undefined,
			numberValue: this.numberPicker ? this.numberPicker.getInputValue() : undefined,
			enumValue: this.enumPicker ? this.enumPicker.getInputValue() : undefined,
		});
	}
	setInputValue(newValue: TargetInput) {
		if (!newValue) {
			return;
		}
		if (newValue.inputType == InputType.Number) {
			if (this.numberPicker && this.numberPicker.inputConfig.label === newValue.label) {
				return;
			}

			this.clearPickers();
			this.numberPicker = new NumberPicker(this.rootElem, null, {
				id: randomUUID(),
				label: newValue.label,
				labelTooltip: newValue.tooltip,
				changedEvent: () => this.encounter.targetsChangeEmitter,
				getValue: () => this.getTargetInput().numberValue,
				setValue: (eventID: EventID, _: null, newValue: number) => {
					this.getTargetInput().numberValue = newValue;
					this.encounter.targetsChangeEmitter.emit(eventID);
				},
			});
		} else if (newValue.inputType == InputType.Bool) {
			if (this.boolPicker && this.boolPicker.inputConfig.label === newValue.label) {
				return;
			}

			this.clearPickers();
			this.boolPicker = new BooleanPicker(this.rootElem, null, {
				id: randomUUID(),
				label: newValue.label,
				labelTooltip: newValue.tooltip,
				extraCssClasses: ['input-inline'],
				changedEvent: () => this.encounter.targetsChangeEmitter,
				getValue: () => this.getTargetInput().boolValue,
				setValue: (eventID: EventID, _: null, newValue: boolean) => {
					this.getTargetInput().boolValue = newValue;
					this.encounter.targetsChangeEmitter.emit(eventID);
				},
			});
		} else if (newValue.inputType == InputType.Enum) {
			this.clearPickers();
			this.enumPicker = new EnumPicker<null>(this.rootElem, null, {
				id: randomUUID(),
				label: newValue.label,
				labelTooltip: newValue.tooltip,
				values: newValue.enumOptions.map((option, index) => {
					return { value: index, name: option };
				}),
				changedEvent: () => this.encounter.targetsChangeEmitter,
				getValue: () => this.getTargetInput().enumValue,
				setValue: (eventID: EventID, _: null, newValue: number) => {
					this.getTargetInput().enumValue = newValue;
					this.encounter.targetsChangeEmitter.emit(eventID);
				},
			});
		}
	}
}

function addEncounterFieldPickers(rootElem: HTMLElement, encounter: Encounter) {
	const durationGroup = Input.newGroupContainer();
	rootElem.appendChild(durationGroup);

	new NumberPicker(durationGroup, encounter, {
		id: 'encounter-duration',
		label: 'Duration',
		labelTooltip: 'The fight length for each sim iteration, in seconds.',
		changedEvent: (encounter: Encounter) => encounter.changeEmitter,
		getValue: (encounter: Encounter) => encounter.getDuration(),
		setValue: (eventID: EventID, encounter: Encounter, newValue: number) => {
			encounter.setDuration(eventID, newValue);
		},
		enableWhen: _ => {
			return !encounter.getUseHealth();
		},
	});
	new NumberPicker(durationGroup, encounter, {
		id: 'encounter-duration-variation',
		label: 'Duration +/-',
		labelTooltip:
			'Adds a random amount of time, in seconds, between [value, -1 * value] to each sim iteration. For example, setting Duration to 180 and Duration +/- to 10 will result in random durations between 170s and 190s.',
		changedEvent: (encounter: Encounter) => encounter.changeEmitter,
		getValue: (encounter: Encounter) => encounter.getDurationVariation(),
		setValue: (eventID: EventID, encounter: Encounter, newValue: number) => {
			encounter.setDurationVariation(eventID, newValue);
		},
		enableWhen: _ => {
			return !encounter.getUseHealth();
		},
	});

	const pvpGroup = Input.newGroupContainer();
	rootElem.appendChild(pvpGroup);

	new BooleanPicker<Encounter>(pvpGroup, encounter, {
		id: 'encounter-pvp',
		label: 'PvP',
		labelTooltip:
			'The target is an enemy player. White hits never glance, and for random stretches of 1 to 10 sec you are out of melee range: no white hits, no melee abilities and no Fire Nova, but spells still go out.',
		inline: true,
		changedEvent: (encounter: Encounter) => encounter.changeEmitter,
		getValue: (encounter: Encounter) => encounter.getPvp(),
		setValue: (eventID: EventID, encounter: Encounter, newValue: boolean) => {
			encounter.setPvp(eventID, newValue);
		},
	});
	new NumberPicker(pvpGroup, encounter, {
		id: 'encounter-pvp-melee-downtime',
		label: 'Out of Melee (%)',
		labelTooltip: 'In PvP, the share of the fight spent out of melee range.',
		changedEvent: (encounter: Encounter) => encounter.changeEmitter,
		getValue: (encounter: Encounter) => encounter.getPvpMeleeDowntime() * 100,
		setValue: (eventID: EventID, encounter: Encounter, newValue: number) => {
			encounter.setPvpMeleeDowntime(eventID, newValue / 100);
		},
		showWhen: _ => encounter.getPvp(),
	});

	const executeGroup = Input.newGroupContainer();
	executeGroup.classList.add('execute-group');
	rootElem.appendChild(executeGroup);

	new NumberPicker(executeGroup, encounter, {
		id: 'encounter-execute-proportion',
		label: 'Execute Duration 20 (%)',
		labelTooltip:
			'Percentage of the total encounter duration, for which the targets will be considered to be in execute range (< 20% HP) for the purpose of effects like Warrior Execute or Mage Molten Fury.',
		changedEvent: (encounter: Encounter) => encounter.changeEmitter,
		getValue: (encounter: Encounter) => encounter.getExecuteProportion20() * 100,
		setValue: (eventID: EventID, encounter: Encounter, newValue: number) => {
			encounter.setExecuteProportion20(eventID, newValue / 100);
		},
		enableWhen: _ => {
			return !encounter.getUseHealth();
		},
	});
	new NumberPicker(executeGroup, encounter, {
		id: 'encounter-execute-proportion-25',
		label: 'Execute Duration 25 (%)',
		labelTooltip:
			"Percentage of the total encounter duration, for which the targets will be considered to be in execute range (< 25% HP) for the purpose of effects like Warlock's Drain Soul.",
		changedEvent: (encounter: Encounter) => encounter.changeEmitter,
		getValue: (encounter: Encounter) => encounter.getExecuteProportion25() * 100,
		setValue: (eventID: EventID, encounter: Encounter, newValue: number) => {
			encounter.setExecuteProportion25(eventID, newValue / 100);
		},
		enableWhen: _ => {
			return !encounter.getUseHealth();
		},
	});
	new NumberPicker(executeGroup, encounter, {
		id: 'encounter-execute-proportion-35',
		label: 'Execute Duration 35 (%)',
		labelTooltip:
			'Percentage of the total encounter duration, for which the targets will be considered to be in execute range (< 35% HP) for the purpose of effects like Warrior Execute or Mage Molten Fury.',
		changedEvent: (encounter: Encounter) => encounter.changeEmitter,
		getValue: (encounter: Encounter) => encounter.getExecuteProportion35() * 100,
		setValue: (eventID: EventID, encounter: Encounter, newValue: number) => {
			encounter.setExecuteProportion35(eventID, newValue / 100);
		},
		enableWhen: _ => {
			return !encounter.getUseHealth();
		},
	});
}

function makeTargetInputsPicker(parent: HTMLElement, encounter: Encounter, targetIndex: number) {
	return new ListPicker<Encounter, TargetInput>(parent, encounter, {
		allowedActions: [],
		itemLabel: 'Target Input',
		extraCssClasses: ['mb-0', 'w-100'],
		isCompact: true,
		horizontalLayout: true,
		changedEvent: (encounter: Encounter) => encounter.targetsChangeEmitter,
		getValue: (encounter: Encounter) => encounter.targets[targetIndex].targetInputs,
		setValue: (eventID: EventID, encounter: Encounter, newValue: Array<TargetInput>) => {
			encounter.targets[targetIndex].targetInputs = newValue;
			encounter.targetsChangeEmitter.emit(eventID);
		},
		newItem: () => TargetInput.create(),
		copyItem: (oldItem: TargetInput) => TargetInput.clone(oldItem),
		newItemPicker: (
			parent: HTMLElement,
			listPicker: ListPicker<Encounter, TargetInput>,
			index: number,
			config: ListItemPickerConfig<Encounter, TargetInput>,
		) => new TargetInputPicker(parent, encounter, targetIndex, index, config),
	});
}

function equalTargetsIgnoreInputs(target1: TargetProto | undefined, target2: TargetProto | undefined): boolean {
	if ((target1 == null) != (target2 == null)) {
		return false;
	}
	if (target1 == null) {
		return true;
	}
	const modTarget2 = TargetProto.clone(target2!);
	modTarget2.targetInputs = target1.targetInputs;
	return TargetProto.equals(target1, modTarget2);
}

const ALL_TARGET_STATS: Array<{ stat: Stat; tooltip: string; extraCssClasses: Array<string> }> = [
	{ stat: Stat.StatHealth, tooltip: '', extraCssClasses: [] },
	{ stat: Stat.StatArmor, tooltip: '', extraCssClasses: [] },
	{ stat: Stat.StatArcaneResistance, tooltip: '', extraCssClasses: [] },
	{ stat: Stat.StatFireResistance, tooltip: '', extraCssClasses: [] },
	{ stat: Stat.StatFrostResistance, tooltip: '', extraCssClasses: [] },
	{ stat: Stat.StatNatureResistance, tooltip: '', extraCssClasses: [] },
	{ stat: Stat.StatShadowResistance, tooltip: '', extraCssClasses: [] },
	{ stat: Stat.StatAttackPower, tooltip: '', extraCssClasses: ['threat-metrics'] },
	{ stat: Stat.StatBlockValue, tooltip: 'Damage a block takes off our hit, for a target with a shield.', extraCssClasses: [] },
	{
		stat: Stat.StatDodge,
		tooltip: "PvP only: the enemy player's dodge chance in percent. When Dodge, Parry or Block is set, the three replace the 5% each a boss has.",
		extraCssClasses: [],
	},
	{
		stat: Stat.StatParry,
		tooltip: "PvP only: the enemy player's parry chance in percent, when we attack from the front. 0 for a class that can't parry.",
		extraCssClasses: [],
	},
	{
		stat: Stat.StatBlock,
		tooltip: "PvP only: the enemy player's block chance in percent, when we attack from the front. 0 without a shield.",
		extraCssClasses: [],
	},
];

// The icons are the hunter's tracking spells for each mob type. There's none for Mechanical, so
// it gets a gizmo.
const mobTypeEnumValues = [
	{ name: 'None', value: MobType.MobTypeUnknown, icon: 'inv_misc_questionmark' },
	{ name: 'Beast', value: MobType.MobTypeBeast, icon: 'ability_tracking' },
	{ name: 'Demon', value: MobType.MobTypeDemon, icon: 'spell_shadow_summonfelhunter' },
	{ name: 'Dragonkin', value: MobType.MobTypeDragonkin, icon: 'inv_misc_head_dragon_01' },
	{ name: 'Elemental', value: MobType.MobTypeElemental, icon: 'spell_frost_summonwaterelemental' },
	{ name: 'Giant', value: MobType.MobTypeGiant, icon: 'ability_racial_avatar' },
	{ name: 'Humanoid', value: MobType.MobTypeHumanoid, icon: 'spell_holy_prayerofhealing' },
	{ name: 'Mechanical', value: MobType.MobTypeMechanical, icon: 'inv_gizmo_02' },
	{ name: 'Undead', value: MobType.MobTypeUndead, icon: 'spell_shadow_darksummoning' },
];
