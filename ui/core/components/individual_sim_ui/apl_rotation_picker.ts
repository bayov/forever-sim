import tippy, { Instance as TippyInstance } from 'tippy.js';

import { Player } from '../../player';
import { APLAction, APLListItem, APLPrepullAction, APLValue } from '../../proto/apl';
import { ActionID as ActionIdProto } from '../../proto/common';
import { ActionId } from '../../proto_utils/action_id';
import { SimUI } from '../../sim_ui';
import { EventID, TypedEvent } from '../../typed_event';
import { existsInDOM, randomUUID } from '../../utils';
import { Component } from '../component';
import { dirtySettings } from '../dirty_settings';
import { Input, InputConfig } from '../input';
import { AdaptiveStringPicker } from '../inputs/string_picker';
import { LIST_PICKER_DRAG_HANDLE, ListItemPickerConfig, ListPicker } from '../list_picker';
import { markUnstacked } from '../unstacked_mark';
import { APLActionPicker } from './apl_actions';
import { aplValueSummary } from './apl_summary';
import { APLValueImplStruct } from './apl_values';

export class APLRotationPicker extends Component {
	constructor(
		parent: HTMLElement,
		simUI: SimUI,
		modPlayer: Player<any>,
		prepullActionNote?: (player: Player<any>, action: APLPrepullAction) => string | undefined,
		// The Leveling Plan tab's editor edits a rotation of the plan, not one of a preset, so
		// it leaves the marks of a changed preset to the Rotation tab.
		tracksPresets = true,
	) {
		super(parent, 'apl-rotation-picker-root');

		// We don't mark the editor's rows, see Input.tracksPresetChanges(), so the editor as a
		// whole tells the Rotation tab and the presets' tooltips that the rotation changed.
		if (tracksPresets) {
			dirtySettings.track({
				elem: this.rootElem,
				read: () => ({ prepull: modPlayer.aplRotation.prepullActions, priority: modPlayer.aplRotation.priorityList }),
				name: () => 'Rotation',
				format: () => '',
			});
		}

		new ListPicker<Player<any>, APLPrepullAction>(this.rootElem, modPlayer, {
			extraCssClasses: ['apl-prepull-action-picker'],
			title: 'Prepull Actions',
			titleTooltip: 'Actions to perform before the pull.',
			itemLabel: 'Prepull Action',
			changedEvent: (player: Player<any>) => player.rotationChangeEmitter,
			getValue: (player: Player<any>) => player.aplRotation.prepullActions,
			setValue: (eventID: EventID, player: Player<any>, newValue: Array<APLPrepullAction>) => {
				player.aplRotation.prepullActions = newValue;
				player.rotationChangeEmitter.emit(eventID);
			},
			newItem: () =>
				APLPrepullAction.create({
					action: {},
					doAtValue: {
						value: { oneofKind: 'const', const: { val: '-1s' } },
					},
				}),
			copyItem: (oldItem: APLPrepullAction) => APLPrepullAction.clone(oldItem),
			newItemPicker: (
				parent: HTMLElement,
				listPicker: ListPicker<Player<any>, APLPrepullAction>,
				index: number,
				config: ListItemPickerConfig<Player<any>, APLPrepullAction>,
			) => new APLPrepullActionPicker(parent, modPlayer, config, index, prepullActionNote),
			inlineMenuBar: true,
			dragWholeItem: true,
		});

		new ListPicker<Player<any>, APLListItem>(this.rootElem, modPlayer, {
			extraCssClasses: ['apl-list-item-picker'],
			title: 'Priority List',
			titleTooltip: 'At each decision point, the simulation will perform the first valid action from this list.',
			itemLabel: 'Action',
			changedEvent: (player: Player<any>) => player.rotationChangeEmitter,
			getValue: (player: Player<any>) => player.aplRotation.priorityList,
			setValue: (eventID: EventID, player: Player<any>, newValue: Array<APLListItem>) => {
				player.aplRotation.priorityList = newValue;
				player.rotationChangeEmitter.emit(eventID);
			},
			newItem: () =>
				APLListItem.create({
					action: {},
				}),
			copyItem: (oldItem: APLListItem) => APLListItem.clone(oldItem),
			newItemPicker: (
				parent: HTMLElement,
				listPicker: ListPicker<Player<any>, APLListItem>,
				index: number,
				config: ListItemPickerConfig<Player<any>, APLListItem>,
			) => new APLListItemPicker(parent, modPlayer, config, index),
			inlineMenuBar: true,
			dragWholeItem: true,
		});

		//modPlayer.rotationChangeEmitter.on(() => console.log('APL: ' + APLRotation.toJsonString(modPlayer.aplRotation)))
	}
}

class APLPrepullActionPicker extends Input<Player<any>, APLPrepullAction> {
	private readonly player: Player<any>;

	private readonly hidePicker: Input<Player<any>, boolean>;
	private readonly doAtPicker: Input<Player<any>, string>;
	private readonly actionPicker: APLActionPicker;

	private getItem(): APLPrepullAction {
		return (
			this.getSourceValue() ||
			APLPrepullAction.create({
				action: {},
			})
		);
	}

	constructor(
		parent: HTMLElement,
		player: Player<any>,
		config: ListItemPickerConfig<Player<any>, APLPrepullAction>,
		index: number,
		prepullActionNote?: (player: Player<any>, action: APLPrepullAction) => string | undefined,
	) {
		config.enableWhen = () => !this.getItem().hide;
		super(parent, 'apl-list-item-picker-root', player, config);
		this.player = player;

		const itemHeaderElem = ListPicker.getItemHeaderElem(this);
		makeListItemWarnings(itemHeaderElem, player, player => player.getCurrentStats().rotationStats?.prepullActions[index]?.warnings || []);
		const iconElem = makeActionIcon(this.rootElem, player, () => this.getItem().action);
		if (prepullActionNote) {
			markUnstacked(iconElem, () => prepullActionNote(player, this.getItem()), player.changeEmitter);
		}

		this.hidePicker = new HidePicker(itemHeaderElem, player, {
			changedEvent: () => this.player.rotationChangeEmitter,
			getValue: () => this.getItem().hide,
			setValue: (eventID: EventID, player: Player<any>, newValue: boolean) => {
				this.getItem().hide = newValue;
				this.player.rotationChangeEmitter.emit(eventID);
			},
		});

		this.doAtPicker = new AdaptiveStringPicker(this.rootElem, this.player, {
			id: randomUUID(),
			label: 'Do At',
			labelTooltip: "Time before pull to do the action. Should be negative, and formatted like, '-1s' or '-2500ms'.",
			extraCssClasses: ['apl-prepull-actions-doat'],
			changedEvent: () => this.player.rotationChangeEmitter,
			getValue: () => (this.getItem().doAtValue?.value as APLValueImplStruct<'const'> | undefined)?.const.val || '',
			setValue: (eventID: EventID, player: Player<any>, newValue: string) => {
				if (newValue) {
					this.getItem().doAtValue = APLValue.create({
						value: { oneofKind: 'const', const: { val: newValue } },
					});
				} else {
					this.getItem().doAtValue = undefined;
				}
				this.player.rotationChangeEmitter.emit(eventID);
			},
			inline: true,
		});
		//this.doAtPicker = new APLValuePicker(this.rootElem, this.player, {
		//	label: 'Do At',
		//	labelTooltip: 'Time before pull to do the action. Should be negative, and formatted like, \'-1s\' or \'-2500ms\'.',
		//	extraCssClasses: ['apl-prepull-actions-doat'],
		//	changedEvent: () => this.player.rotationChangeEmitter,
		//	getValue: () => this.getItem().doAtValue,
		//	setValue: (eventID: EventID, player: Player<any>, newValue: APLValue | undefined) => {
		//		this.getItem().doAtValue = newValue;
		//		this.player.rotationChangeEmitter.emit(eventID);
		//	},
		//	inline: true,
		//});

		this.actionPicker = new APLActionPicker(this.rootElem, this.player, {
			changedEvent: () => this.player.rotationChangeEmitter,
			getValue: () => this.getItem().action!,
			setValue: (eventID: EventID, player: Player<any>, newValue: APLAction) => {
				this.getItem().action = newValue;
				this.player.rotationChangeEmitter.emit(eventID);
			},
		});
		this.init();
	}

	getInputElem(): HTMLElement | null {
		return this.rootElem;
	}

	getInputValue(): APLPrepullAction {
		const item = APLPrepullAction.create({
			hide: this.hidePicker.getInputValue(),
			doAtValue: {
				value: { oneofKind: 'const', const: { val: this.doAtPicker.getInputValue() } },
			},
			action: this.actionPicker.getInputValue(),
		});
		return item;
	}

	setInputValue(newValue: APLPrepullAction) {
		if (!newValue) {
			return;
		}
		this.hidePicker.setInputValue(newValue.hide);
		this.doAtPicker.setInputValue((newValue.doAtValue?.value as APLValueImplStruct<'const'> | undefined)?.const.val || '');
		this.actionPicker.setInputValue(newValue.action || APLAction.create());
	}
}

class APLListItemPicker extends Input<Player<any>, APLListItem> {
	private readonly player: Player<any>;

	private readonly hidePicker: Input<Player<any>, boolean>;
	private readonly actionPicker: APLActionPicker;

	private getItem(): APLListItem {
		return (
			this.getSourceValue() ||
			APLListItem.create({
				action: {},
			})
		);
	}

	constructor(parent: HTMLElement, player: Player<any>, config: ListItemPickerConfig<Player<any>, APLListItem>, index: number) {
		config.enableWhen = () => !this.getItem().hide;
		super(parent, 'apl-list-item-picker-root', player, config);
		this.player = player;

		const itemHeaderElem = ListPicker.getItemHeaderElem(this);
		makeListItemWarnings(itemHeaderElem, player, player => player.getCurrentStats().rotationStats?.priorityList[index]?.warnings || []);
		makeActionIcon(this.rootElem, player, () => this.getItem().action);

		this.hidePicker = new HidePicker(itemHeaderElem, player, {
			changedEvent: () => this.player.rotationChangeEmitter,
			getValue: () => this.getItem().hide,
			setValue: (eventID: EventID, player: Player<any>, newValue: boolean) => {
				this.getItem().hide = newValue;
				this.player.rotationChangeEmitter.emit(eventID);
			},
		});

		this.actionPicker = new APLActionPicker(this.rootElem, this.player, {
			changedEvent: () => this.player.rotationChangeEmitter,
			getValue: () => this.getItem().action!,
			setValue: (eventID: EventID, player: Player<any>, newValue: APLAction) => {
				this.getItem().action = newValue;
				this.player.rotationChangeEmitter.emit(eventID);
			},
		});
		makeCollapsible(this.rootElem, this.actionPicker.rootElem, itemHeaderElem, player, () => this.getItem().action?.condition);
		this.init();
	}

	getInputElem(): HTMLElement | null {
		return this.rootElem;
	}

	getInputValue(): APLListItem {
		const item = APLListItem.create({
			hide: this.hidePicker.getInputValue(),
			action: this.actionPicker.getInputValue(),
		});
		return item;
	}

	setInputValue(newValue: APLListItem) {
		if (!newValue) {
			return;
		}
		this.hidePicker.setInputValue(newValue.hide);
		this.actionPicker.setInputValue(newValue.action || APLAction.create());
	}
}

// A priority list row as a card that we expand to edit its condition.
//
// A long condition, like a leveling rotation's checks on the mobs and our health and mana, is a
// tall tree of dropdowns. So a row starts collapsed and shows its condition as one line of text
// next to the action. A click on that line, or on the chevron with the row's buttons, expands the
// row, and the condition's dropdowns go under the action across the whole row. The row's picker
// stays with its place in the list, so a row we expanded stays open while we edit it.
function makeCollapsible(
	rowElem: HTMLElement,
	actionPickerElem: HTMLElement,
	itemHeaderElem: HTMLElement,
	player: Player<any>,
	getCondition: () => APLValue | undefined,
) {
	const summaryElem = document.createElement('a');
	summaryElem.classList.add('apl-condition-summary');
	summaryElem.href = 'javascript:void(0)';
	summaryElem.setAttribute('role', 'button');
	actionPickerElem.appendChild(summaryElem);
	const summaryTooltip = tippy(summaryElem, { content: '', maxWidth: '40rem' });

	const toggleElem = ListPicker.makeActionElem('apl-expand-toggle', 'fa-chevron-down');
	itemHeaderElem.prepend(toggleElem);
	const toggleTooltip = tippy(toggleElem, { content: 'Edit condition' });

	const setExpanded = (expanded: boolean) => {
		rowElem.classList.toggle('apl-collapsed', !expanded);
		toggleElem.querySelector('i')?.classList.toggle('fa-chevron-up', expanded);
		toggleElem.querySelector('i')?.classList.toggle('fa-chevron-down', !expanded);
		toggleTooltip.setContent(expanded ? 'Collapse' : 'Edit condition');
	};
	setExpanded(false);
	toggleElem.addEventListener('click', () => setExpanded(rowElem.classList.contains('apl-collapsed')));
	summaryElem.addEventListener('click', () => setExpanded(true));

	// The names of spells and auras come in later, so a newer change can finish first.
	let shown = 0;
	const update = async () => {
		if (!existsInDOM(summaryElem)) {
			player.rotationChangeEmitter.off(update);
			summaryTooltip.destroy();
			toggleTooltip.destroy();
			return;
		}
		const call = ++shown;
		const text = await aplValueSummary(getCondition());
		if (call !== shown) return;
		summaryElem.textContent = text ? `If ${text}` : 'Always';
		summaryElem.classList.toggle('apl-condition-summary-empty', !text);
		summaryTooltip.setContent(text ? `If ${text}` : 'No condition. Click to add one.');
	};
	update();
	player.rotationChangeEmitter.on(update);
}

function makeListItemWarnings(itemHeaderElem: HTMLElement, player: Player<any>, getWarnings: (player: Player<any>) => Array<string>) {
	const warningsElem = ListPicker.makeActionElem('apl-warnings', 'fa-exclamation-triangle');
	warningsElem.classList.add('warning', 'link-warning');
	warningsElem.setAttribute('data-bs-html', 'true');
	const warningsTooltip = tippy(warningsElem, {
		theme: 'dropdown-tooltip',
		content: 'Warnings',
	});
	itemHeaderElem.appendChild(warningsElem);

	const updateWarnings = async () => {
		if (!existsInDOM(warningsElem)) {
			warningsTooltip?.destroy();
			warningsElem?.remove();
			player.currentStatsEmitter.off(updateWarnings);
			return;
		}
		warningsTooltip.setContent('');
		const warnings = getWarnings(player);
		if (!warnings.length) {
			warningsElem.style.visibility = 'hidden';
		} else {
			warningsElem.style.visibility = 'visible';
			const formattedWarnings = await Promise.all(warnings.map(w => ActionId.replaceAllInString(w)));
			warningsTooltip.setContent(
				`
				<p>This action has warnings, and might not behave as expected.</p>
				<ul>
					${formattedWarnings.map(w => `<li>${w}</li>`).join('')}
				</ul>
			`,
			);
		}
	};
	updateWarnings();
	player.currentStatsEmitter.on(updateWarnings);
}

// The large icon on the left of a row, for what the row does.
//
// A row that casts a spell or uses an aura shows that spell's icon, with its wowhead tooltip. A
// Scheduled Action shows the icon of the action it schedules. Other actions, like Wait or
// Autocast Other Cooldowns, show a plain icon for their kind. We can drag the icon to move the
// row, like the handle on its left, and a click opens the spell on wowhead.
function makeActionIcon(parent: HTMLElement, player: Player<any>, getAction: () => APLAction | undefined): HTMLElement {
	const iconElem = document.createElement('a');
	iconElem.classList.add('apl-row-icon', LIST_PICKER_DRAG_HANDLE);
	iconElem.dataset.whtticon = 'false';
	iconElem.target = '_blank';
	parent.prepend(iconElem);

	let shownKey = '';
	const update = () => {
		if (!existsInDOM(iconElem)) {
			player.rotationChangeEmitter.off(update);
			return;
		}
		let action = getAction();
		while (action?.action.oneofKind === 'schedule' && action.action.schedule.innerAction) {
			action = action.action.schedule.innerAction;
		}
		const kind = action?.action.oneofKind;
		const idProto = actionIdOf(action);
		const key = `${kind}:${idProto ? ActionIdProto.toJsonString(idProto) : ''}`;
		if (key === shownKey) {
			return;
		}
		shownKey = key;

		iconElem.querySelector(':scope > i.fa')?.remove();
		iconElem.style.backgroundImage = '';
		iconElem.removeAttribute('href');
		delete iconElem.dataset.wowhead;
		const actionId = idProto ? ActionId.fromProto(idProto) : undefined;
		if (actionId && actionId.anyId()) {
			iconElem.classList.remove('apl-row-icon-kind');
			actionId.fillAndSet(iconElem, true, true).then(filled => {
				if (shownKey === key) {
					filled.setWowheadDataset(iconElem, { useBuffAura: kind !== 'castSpell' && kind !== 'channelSpell' });
				}
			});
		} else {
			iconElem.classList.add('apl-row-icon-kind');
			const fontIcon = document.createElement('i');
			fontIcon.classList.add('fa', ACTION_KIND_ICONS[kind ?? ''] ?? 'fa-question');
			iconElem.appendChild(fontIcon);
		}
	};
	update();
	player.rotationChangeEmitter.on(update);
	return iconElem;
}

function actionIdOf(action: APLAction | undefined): ActionIdProto | undefined {
	const impl = action?.action;
	switch (impl?.oneofKind) {
		case 'castSpell':
			return impl.castSpell.spellId;
		case 'channelSpell':
			return impl.channelSpell.spellId;
		case 'multidot':
			return impl.multidot.spellId;
		case 'multishield':
			return impl.multishield.spellId;
		case 'activateAura':
			return impl.activateAura.auraId;
		case 'activateAuraWithStacks':
			return impl.activateAuraWithStacks.auraId;
		case 'cancelAura':
			return impl.cancelAura.auraId;
		case 'triggerIcd':
			return impl.triggerIcd.auraId;
	}
	return undefined;
}

const ACTION_KIND_ICONS: Record<string, string> = {
	autocastOtherCooldowns: 'fa-bolt',
	wait: 'fa-hourglass-half',
	waitUntil: 'fa-hourglass-half',
	schedule: 'fa-clock',
	sequence: 'fa-list-ol',
	strictSequence: 'fa-list-ol',
	resetSequence: 'fa-rotate-left',
	changeTarget: 'fa-crosshairs',
	itemSwap: 'fa-right-left',
	move: 'fa-person-running',
	addComboPoints: 'fa-plus',
	catOptimalRotationAction: 'fa-gears',
	customRotation: 'fa-gears',
	castPaladinPrimarySeal: 'fa-certificate',
	levelingLookForMob: 'fa-binoculars',
	levelingWalkToPullTarget: 'fa-person-walking',
	levelingRest: 'fa-utensils',
	levelingBandage: 'fa-bandage',
	shamanWalkToTotem: 'fa-monument',
	castAssignedTotem: 'fa-monument',
};

class HidePicker extends Input<Player<any>, boolean> {
	private readonly inputElem: HTMLElement;
	private readonly iconElem: HTMLElement;
	private tooltip: TippyInstance;

	constructor(parent: HTMLElement, modObject: Player<any>, config: InputConfig<Player<any>, boolean>) {
		super(parent, 'hide-picker-root', modObject, config);

		this.inputElem = ListPicker.makeActionElem('hide-picker-button', 'fa-eye');
		this.iconElem = this.inputElem.childNodes[0] as HTMLElement;
		this.rootElem.appendChild(this.inputElem);
		this.tooltip = tippy(this.inputElem, { content: 'Enable/Disable' });

		this.init();

		this.inputElem.addEventListener('click', () => {
			this.setInputValue(!this.getInputValue());
			this.inputChanged(TypedEvent.nextEventID());
		});
	}

	getInputElem(): HTMLElement {
		return this.inputElem;
	}

	getInputValue(): boolean {
		return this.iconElem.classList.contains('fa-eye-slash');
	}

	setInputValue(newValue: boolean) {
		if (newValue) {
			this.iconElem.classList.add('fa-eye-slash');
			this.iconElem.classList.remove('fa-eye');
			this.tooltip.setContent('Enable Action');
		} else {
			this.iconElem.classList.add('fa-eye');
			this.iconElem.classList.remove('fa-eye-slash');
			this.tooltip.setContent('Disable Action');
		}
	}
}
