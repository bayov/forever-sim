import clsx from 'clsx';
import tippy, { Instance as TippyInstance } from 'tippy.js';

import { EventID, TypedEvent } from '../typed_event.js';
import { Input, InputConfig } from './input.js';

export type ListItemAction = 'create' | 'delete' | 'move' | 'copy';

export interface ListPickerActionsConfig {
	create?: {
		// Whether or not to use an icon for the create action button
		// defaults to FALSE
		useIcon?: boolean;
	};
}

export interface ListPickerConfig<ModObject, ItemType> extends Omit<InputConfig<ModObject, Array<ItemType>>, 'id'> {
	itemLabel: string;
	newItem: () => ItemType;
	copyItem: (oldItem: ItemType) => ItemType;
	newItemPicker: (
		parent: HTMLElement,
		listPicker: ListPicker<ModObject, ItemType>,
		index: number,
		config: ListItemPickerConfig<ModObject, ItemType>,
	) => Input<ModObject, ItemType>;
	actions?: ListPickerActionsConfig;
	title?: string;
	titleTooltip?: string;
	inlineMenuBar?: boolean;
	hideUi?: boolean;
	horizontalLayout?: boolean;
	// if set, will remove the border and padding of the list items
	isCompact?: boolean;
	// If set, will disable the delete button if the list is at the minimum.
	minimumItems?: number;
	// If set, only actions included in the list are allowed. Otherwise, all actions are allowed.
	allowedActions?: Array<ListItemAction>;
	// If set, we drag an item from anywhere on it, except its inputs, dropdowns and buttons, and
	// the items around it slide away to make room, see ListPicker.makeSortable. Its move handle
	// goes first in the item, on its left. A link in the item, like the rotation editor's action
	// icon, starts a drag too when it has the LIST_PICKER_DRAG_HANDLE class.
	dragWholeItem?: boolean;
}

export const LIST_PICKER_DRAG_HANDLE = 'list-picker-drag-handle';

// What we press on to use it rather than to drag the item it's in.
const NOT_DRAGGABLE = `input, textarea, select, button, [contenteditable], .dropdown-menu, a:not(.${LIST_PICKER_DRAG_HANDLE}):not(.list-picker-item-move)`;

// How far we move the pointer before a press turns into a drag, so a click stays a click.
const DRAG_THRESHOLD_PX = 4;

const DEFAULT_CONFIG = {
	actions: {
		create: {
			useIcon: false,
		},
	},
};

export interface ListItemPickerConfig<ModObject, ItemType> extends InputConfig<ModObject, ItemType> {}

interface ItemPickerPair<ItemType> {
	elem: HTMLElement;
	picker: Input<any, ItemType>;
	idx: number;
}

interface ListDragData<ModObject, ItemType> {
	listPicker: ListPicker<ModObject, ItemType>;
	item: ItemPickerPair<ItemType>;
}

let curDragData: ListDragData<any, any> | null = null;

export class ListPicker<ModObject, ItemType> extends Input<ModObject, Array<ItemType>> {
	readonly config: ListPickerConfig<ModObject, ItemType>;
	private readonly itemsDiv: HTMLElement;

	private itemPickerPairs: Array<ItemPickerPair<ItemType>>;

	constructor(parent: HTMLElement, modObject: ModObject, config: ListPickerConfig<ModObject, ItemType>) {
		if (config.isCompact) config.extraCssClasses = [...(config.extraCssClasses || []), 'list-picker-compact'];

		super(parent, 'list-picker-root', modObject, config);
		this.config = { ...DEFAULT_CONFIG, ...config };
		this.itemPickerPairs = [];

		this.rootElem.appendChild(
			<>
				{config.title && <label className="list-picker-title form-label">{config.title}</label>}
				<div className="list-picker-items"></div>
			</>,
		);

		if (this.config.hideUi) {
			this.rootElem.classList.add('d-none');
		}
		if (this.config.horizontalLayout) {
			this.config.inlineMenuBar = true;
			this.rootElem.classList.add('horizontal');
		}

		if (this.config.titleTooltip) {
			const titleTooltip = tippy(this.rootElem.querySelector('.list-picker-title') as HTMLElement, {
				content: this.config.titleTooltip,
			});
			this.addOnDisposeCallback(() => titleTooltip?.destroy());
		}

		this.itemsDiv = this.rootElem.getElementsByClassName('list-picker-items')[0] as HTMLElement;

		if (this.actionEnabled('create')) {
			let newItemButton: HTMLElement | null = null;
			let newButtonTooltip: TippyInstance | null = null;
			if (this.config.actions?.create?.useIcon) {
				newItemButton = ListPicker.makeActionElem('link-success', 'fa-plus');
				newButtonTooltip = tippy(newItemButton, {
					allowHTML: false,
					content: `New ${config.itemLabel}`,
				});
				this.addOnDisposeCallback(() => newButtonTooltip?.destroy());
			} else {
				newItemButton = (<button className="btn btn-primary">New {config.itemLabel}</button>) as HTMLButtonElement;
			}
			newItemButton.classList.add('list-picker-new-button');
			newItemButton.addEventListener(
				'click',
				() => {
					const newItem = this.config.newItem();
					const newList = this.config.getValue(this.modObject).concat([newItem]);
					this.config.setValue(TypedEvent.nextEventID(), this.modObject, newList);
					if (newButtonTooltip) {
						newButtonTooltip.hide();
					}
				},
				{ signal: this.signal },
			);

			this.rootElem.appendChild(newItemButton);
		}

		this.init();
	}

	getInputElem(): HTMLElement {
		return this.rootElem;
	}

	getInputValue(): Array<ItemType> {
		return this.itemPickerPairs.map(pair => pair.picker.getInputValue());
	}

	setInputValue(newValue: Array<ItemType>): void {
		// Add/remove pickers to make the lengths match.
		if (newValue.length < this.itemPickerPairs.length) {
			this.itemPickerPairs.slice(newValue.length).forEach(ipp => ipp.elem.remove());
			this.itemPickerPairs = this.itemPickerPairs.slice(0, newValue.length);
		} else if (newValue.length > this.itemPickerPairs.length) {
			const numToAdd = newValue.length - this.itemPickerPairs.length;
			for (let i = 0; i < numToAdd; i++) {
				this.addNewPicker();
			}
		}

		// Set all the values.
		newValue.forEach((val, i) => this.itemPickerPairs[i].picker.setInputValue(val));
	}

	private actionEnabled(action: ListItemAction): boolean {
		return !this.config.allowedActions || this.config.allowedActions.includes(action);
	}

	private addNewPicker() {
		const index = this.itemPickerPairs.length;
		const itemContainer = document.createElement('div');
		itemContainer.classList.add('list-picker-item-container');
		if (this.config.inlineMenuBar) {
			itemContainer.classList.add('inline');
		}
		this.itemsDiv.appendChild(itemContainer);

		const itemElem = document.createElement('div');
		itemElem.classList.add('list-picker-item');

		const itemHeader = document.createElement('div');
		itemHeader.classList.add('list-picker-item-header');

		if (this.config.inlineMenuBar) {
			itemContainer.appendChild(itemElem);
			itemContainer.appendChild(itemHeader);
		} else {
			itemContainer.appendChild(itemHeader);
			itemContainer.appendChild(itemElem);
			if (this.config.itemLabel) {
				const itemLabel = document.createElement('h6');
				itemLabel.classList.add('list-picker-item-title');
				itemLabel.textContent = `${this.config.itemLabel} ${this.itemPickerPairs.length + 1}`;
				itemHeader.appendChild(itemLabel);
			}
		}

		const itemPicker = this.config.newItemPicker(itemElem, this, index, {
			changedEvent: this.config.changedEvent,
			getValue: () => this.getSourceValue()[index],
			setValue: (eventID: EventID, modObj: ModObject, newValue: ItemType) => {
				const newList = this.getSourceValue();
				newList[index] = newValue;
				this.config.setValue(eventID, modObj, newList);
			},
		});

		const item: ItemPickerPair<ItemType> = { elem: itemContainer, picker: itemPicker, idx: index };

		if (this.actionEnabled('move')) {
			const moveButton = ListPicker.makeActionElem('list-picker-item-move', this.config.dragWholeItem ? 'fa-grip-vertical' : 'fa-arrows-up-down');
			if (this.config.dragWholeItem) {
				itemContainer.prepend(moveButton);
			} else {
				itemHeader.appendChild(moveButton);
			}

			const moveButtonTooltip = tippy(moveButton, {
				allowHTML: false,
				content: 'Move (Drag+Drop)',
			});

			moveButton.addEventListener(
				'click',
				() => {
					moveButtonTooltip.hide();
				},
				{ signal: this.signal },
			);
			this.addOnDisposeCallback(() => {
				moveButtonTooltip?.destroy();
			});

			if (this.config.dragWholeItem) {
				this.makeSortable(itemContainer, index);
			}

			moveButton.draggable = !this.config.dragWholeItem;
			moveButton.addEventListener(
				'dragstart',
				event => {
					if (event.target == moveButton) {
						event.dataTransfer!.dropEffect = 'move';
						event.dataTransfer!.effectAllowed = 'move';
						itemContainer.classList.add('dragfrom');
						curDragData = {
							listPicker: this,
							item: item,
						};
					}
				},
				{ signal: this.signal },
			);

			let dragEnterCounter = 0;
			itemContainer.addEventListener(
				'dragenter',
				event => {
					if (!curDragData || curDragData.listPicker != this) {
						return;
					}
					event.preventDefault();
					dragEnterCounter++;
					itemContainer.classList.add('dragto');
				},
				{ signal: this.signal },
			);

			itemContainer.addEventListener(
				'dragleave',
				event => {
					if (!curDragData || curDragData.listPicker != this) {
						return;
					}
					event.preventDefault();
					dragEnterCounter--;
					if (dragEnterCounter <= 0) {
						itemContainer.classList.remove('dragto');
					}
				},
				{ signal: this.signal },
			);

			itemContainer.addEventListener(
				'dragover',
				event => {
					if (!curDragData || curDragData.listPicker != this) {
						return;
					}
					event.preventDefault();
				},
				{ signal: this.signal },
			);

			itemContainer.addEventListener(
				'drop',
				event => {
					if (!curDragData || curDragData.listPicker != this) {
						return;
					}
					event.preventDefault();
					dragEnterCounter = 0;
					itemContainer.classList.remove('dragto');
					curDragData.item.elem.classList.remove('dragfrom');

					const srcIdx = curDragData.item.idx;
					const dstIdx = index;
					const newList = this.config.getValue(this.modObject);
					const arrElem = newList[srcIdx];
					newList.splice(srcIdx, 1);
					newList.splice(dstIdx, 0, arrElem);
					this.config.setValue(TypedEvent.nextEventID(), this.modObject, newList);

					curDragData = null;
				},
				{ signal: this.signal },
			);
		}

		if (this.actionEnabled('copy')) {
			const copyButton = ListPicker.makeActionElem('list-picker-item-copy', 'fa-copy');
			itemHeader.appendChild(copyButton);
			const copyButtonTooltip = tippy(copyButton, {
				allowHTML: false,
				content: `Copy to New ${this.config.itemLabel}`,
			});

			copyButton.addEventListener(
				'click',
				() => {
					const newList = this.config.getValue(this.modObject).slice();
					newList.splice(index, 0, this.config.copyItem(newList[index]));
					this.config.setValue(TypedEvent.nextEventID(), this.modObject, newList);
					copyButtonTooltip.hide();
				},
				{ signal: this.signal },
			);
			this.addOnDisposeCallback(() => copyButtonTooltip?.destroy());
		}

		if (this.actionEnabled('delete')) {
			if (!this.config.minimumItems || index + 1 > this.config.minimumItems) {
				const deleteButton = ListPicker.makeActionElem('list-picker-item-delete', 'fa-times');
				deleteButton.classList.add('link-danger');
				itemHeader.appendChild(deleteButton);

				const deleteButtonTooltip = tippy(deleteButton, {
					allowHTML: false,
					content: `Delete ${this.config.itemLabel}`,
				});

				deleteButton.addEventListener(
					'click',
					() => {
						const newList = this.config.getValue(this.modObject);
						newList.splice(index, 1);
						this.config.setValue(TypedEvent.nextEventID(), this.modObject, newList);
						deleteButtonTooltip.hide();
					},
					{ signal: this.signal },
				);
				this.addOnDisposeCallback(() => deleteButtonTooltip?.destroy());
			}
		}

		this.itemPickerPairs.push(item);
	}

	// Lets us drag the item to a new place in the list, with the items around it sliding away.
	//
	// A press on the item that moves the pointer a few pixels starts the drag, unless it's on an
	// input, a dropdown or a button. The item follows the pointer, and each item it passes slides
	// up or down by the item's height to open a gap. When we let go, the item slides into the gap
	// and only then we save the new order, so the list doesn't jump. Near the top or the bottom
	// of the scrolling area we scroll it, for lists taller than the screen.
	private makeSortable(itemContainer: HTMLElement, index: number) {
		itemContainer.classList.add('list-picker-sortable');
		itemContainer.addEventListener(
			'dragstart',
			event => {
				// We drag with the pointer events below. A link, like the rotation editor's icon,
				// would start the browser's own drag.
				if ((event.target as HTMLElement).closest?.('.list-picker-sortable') == itemContainer) {
					event.preventDefault();
				}
			},
			{ signal: this.signal },
		);

		itemContainer.addEventListener(
			'pointerdown',
			down => {
				const target = down.target as HTMLElement;
				if (down.button != 0 || target.closest(NOT_DRAGGABLE) || target.closest('.list-picker-sortable') != itemContainer) {
					return;
				}

				const items = Array.from(this.itemsDiv.children) as Array<HTMLElement>;
				const scroller = scrollParent(this.itemsDiv);
				const startScroll = scroller.scrollTop;
				let started = false;
				let pointerY = down.clientY;
				let dropIndex = index;
				let rects: Array<DOMRect> = [];
				let gap = 0;
				let shift = 0;
				let frame = 0;

				const place = () => {
					const dy = pointerY - down.clientY + scroller.scrollTop - startScroll;
					itemContainer.style.transform = `translateY(${dy}px)`;

					// The middle of the dragged item, in the list's coordinates before the drag.
					const middle = rects[index].top + rects[index].height / 2 + dy;
					dropIndex = index;
					items.forEach((other, i) => {
						if (i == index) return;
						const otherMiddle = rects[i].top + rects[i].height / 2;
						let offset = 0;
						if (i > index && middle > otherMiddle) {
							offset = -shift;
							dropIndex = Math.max(dropIndex, i);
						} else if (i < index && middle < otherMiddle) {
							offset = shift;
							dropIndex = Math.min(dropIndex, i);
						}
						other.style.transform = offset ? `translateY(${offset}px)` : '';
					});
				};

				const autoScroll = () => {
					const bounds = scroller == document.scrollingElement ? { top: 0, bottom: window.innerHeight } : scroller.getBoundingClientRect();
					const edge = 60;
					if (pointerY < bounds.top + edge) {
						scroller.scrollTop -= Math.ceil((bounds.top + edge - pointerY) / 4);
					} else if (pointerY > bounds.bottom - edge) {
						scroller.scrollTop += Math.ceil((pointerY - bounds.bottom + edge) / 4);
					}
					place();
					frame = requestAnimationFrame(autoScroll);
				};

				const start = () => {
					started = true;
					rects = items.map(item => item.getBoundingClientRect());
					// Measured with the scroll at the start, like the pointer's movement.
					gap = items.length > 1 ? rects[1].top - rects[0].bottom : 0;
					shift = rects[index].height + gap;
					itemContainer.setPointerCapture(down.pointerId);
					this.itemsDiv.classList.add('list-picker-sorting');
					itemContainer.classList.add('list-picker-dragging');
					frame = requestAnimationFrame(autoScroll);
				};

				const move = (event: PointerEvent) => {
					pointerY = event.clientY;
					if (!started && Math.abs(pointerY - down.clientY) >= DRAG_THRESHOLD_PX) {
						start();
					}
					if (started) {
						event.preventDefault();
					}
				};

				const finish = () => {
					window.removeEventListener('pointermove', move);
					window.removeEventListener('pointerup', finish);
					window.removeEventListener('pointercancel', finish);
					if (!started) {
						return;
					}
					cancelAnimationFrame(frame);

					// A drag isn't a click, so the icon we pressed on doesn't open its link.
					itemContainer.addEventListener(
						'click',
						click => {
							click.preventDefault();
							click.stopPropagation();
						},
						{ capture: true, once: true },
					);

					// Where the item's top ends up: the top of the item it takes the place of when it
					// moves up, or the bottom of that item less its own height when it moves down.
					const dyNow = scroller.scrollTop - startScroll;
					const finalTop = dropIndex < index ? rects[dropIndex].top : rects[dropIndex].bottom - rects[index].height;
					itemContainer.classList.remove('list-picker-dragging');
					itemContainer.classList.add('list-picker-settling');
					itemContainer.style.transform = `translateY(${finalTop - rects[index].top + dyNow}px)`;

					window.setTimeout(() => {
						items.forEach(item => {
							item.style.transform = '';
							item.classList.remove('list-picker-settling');
						});
						this.itemsDiv.classList.remove('list-picker-sorting');
						if (dropIndex != index) {
							const newList = this.config.getValue(this.modObject);
							const [moved] = newList.splice(index, 1);
							newList.splice(dropIndex, 0, moved);
							this.config.setValue(TypedEvent.nextEventID(), this.modObject, newList);
						}
					}, SETTLE_MS);
				};

				// On the window, since a quick first move can leave the item before the drag starts.
				window.addEventListener('pointermove', move);
				window.addEventListener('pointerup', finish);
				window.addEventListener('pointercancel', finish);
			},
			{ signal: this.signal },
		);
	}

	static makeActionElem(cssClass: string, iconCssClass: string): HTMLAnchorElement {
		return (
			<a href="javascript:void(0)" className={clsx('list-picker-item-action', cssClass)} attributes={{ role: 'button' }}>
				<i className={clsx('fa', 'fa-xl', iconCssClass)}></i>
			</a>
		) as HTMLAnchorElement;
	}

	static getItemHeaderElem(itemPicker: Input<any, any>): HTMLElement {
		const itemElem = itemPicker.rootElem.parentElement!;
		const headerElem = itemElem.nextElementSibling || itemElem.previousElementSibling;
		if (!headerElem?.classList.contains('list-picker-item-header')) {
			throw new Error('Could not find list item header');
		}
		return headerElem as HTMLElement;
	}
}

// How long the dragged item takes to slide into its place, in ms. Keep in sync with the
// list-picker-settling transition in _list_picker.scss.
const SETTLE_MS = 150;

// The element that scrolls the list, or the page when nothing in between does.
function scrollParent(elem: HTMLElement): HTMLElement {
	for (let parent = elem.parentElement; parent; parent = parent.parentElement) {
		const overflow = getComputedStyle(parent).overflowY;
		if ((overflow == 'auto' || overflow == 'scroll') && parent.scrollHeight > parent.clientHeight) {
			return parent;
		}
	}
	return document.scrollingElement as HTMLElement;
}
