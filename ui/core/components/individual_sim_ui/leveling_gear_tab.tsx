import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { setItemQualityCssClass } from '../../css_utils';
import { IndividualSimUI } from '../../individual_sim_ui';
import {
	LEVELING_GEAR_MAX_LEVEL,
	LEVELING_GEAR_SLOTS,
	LevelingGearLevel,
	LevelingGearPiece,
	LevelingGearPlan,
	isPlainItemId,
	planLevelingGear,
} from '../../leveling_gear';
import { Player } from '../../player';
import { ItemQuality, ItemSlot, ItemSpec, Spec } from '../../proto/common';
import { LevelingGearPin, UIItem as Item } from '../../proto/ui';
import { ActionId } from '../../proto_utils/action_id';
import { EquippedItem } from '../../proto_utils/equipped_item';
import { Gear } from '../../proto_utils/gear';
import { slotNames, statNames } from '../../proto_utils/names';
import { Stats } from '../../proto_utils/stats';
import {
	classToEligibleRangedWeaponTypes,
	classToEligibleWeaponTypes,
	dualWieldClasses,
	getEligibleItemSlots,
	maxArmorTypeAt,
	raceToFaction,
} from '../../proto_utils/utils';
import { TypedEvent } from '../../typed_event';
import { BaseModal } from '../base_modal';
import { ContentBlock } from '../content_block';
import { ItemRenderer } from '../gear_picker/gear_picker';
import { getEmptySlotIconUrl } from '../gear_picker/utils';
import { NumberPicker } from '../number_picker';
import { SimTab } from '../sim_tab';

// The Leveling Gear tab shows the leveling gear plan (leveling_gear.ts): what we wear at each
// level from 1 to 60 when we gear up as we go. It's only there in Leveling mode.
//
// It's laid out like the Gear tab. On the left are the plan's settings and the items we put in
// by hand. On the right the selected level's gear is in the same gear pane, and under both the
// grid has a row per slot and a column per level, in the color of the item we wear then. A
// click on a column picks that level, and Equip puts its gear on in the Gear tab.
export class LevelingGearTab extends SimTab {
	protected simUI: IndividualSimUI<Spec>;
	private readonly player: Player<Spec>;
	private selectedLevel = 30;
	private plan: LevelingGearPlan | null = null;
	private readonly selectedLevelEmitter = new TypedEvent<void>('LevelingGearSelectedLevel');

	private readonly gridElem: HTMLElement;
	private readonly paneTitleElem: HTMLElement;
	private readonly pinsElem: HTMLElement;
	private readonly weightsElem: HTMLElement;
	private readonly slotRenderers = new Map<ItemSlot, ItemRenderer>();

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		super(parentElem, simUI, { identifier: 'leveling-gear-tab', title: 'Leveling Gear' });
		this.simUI = simUI;
		this.player = simUI.player;

		const side = (<div className="leveling-gear-side tab-panel-col" />) as HTMLElement;
		const paneCol = (<div className="leveling-gear-pane-col" />) as HTMLElement;
		const gridCol = (<div className="leveling-gear-grid-col" />) as HTMLElement;
		this.contentContainer.appendChild(
			<div className="leveling-gear">
				<div className="leveling-gear-top">
					{side}
					{paneCol}
				</div>
				{gridCol}
			</div>,
		);

		// The plan's settings: the seed, the level we look at and Equip.
		const planBlock = new ContentBlock(side, 'leveling-gear-plan-block', {
			header: {
				title: 'Leveling Gear Plan',
				tooltip:
					'One random leveling journey. Each level, each slot comes across an item now and then, mostly white early and green later with a few blues, and we wear it only when it beats what we have.',
			},
		});
		const reroll = ref<HTMLButtonElement>();
		const equip = ref<HTMLButtonElement>();
		const seedRow = ref<HTMLDivElement>();
		const levelRow = ref<HTMLDivElement>();
		const weights = ref<HTMLParagraphElement>();
		planBlock.bodyElement.appendChild(
			<>
				<p className="leveling-gear-weights" ref={weights} />
				<div className="leveling-gear-row" ref={seedRow}>
					<button className="btn btn-outline-primary" ref={reroll}>
						<i className="fas fa-dice me-1" />
						Reroll
					</button>
				</div>
				<div className="leveling-gear-row" ref={levelRow}>
					<button className="btn btn-primary" ref={equip}>
						<i className="fas fa-shirt me-1" />
						Equip
					</button>
				</div>
			</>,
		);
		this.weightsElem = weights.value!;
		new NumberPicker<Player<Spec>>(seedRow.value!, this.player, {
			id: 'leveling-gear-seed',
			label: 'Seed',
			labelTooltip: 'The plan is one random leveling journey. Another seed, or Reroll, plays out another one.',
			changedEvent: player => player.levelingGearPlanChangeEmitter,
			getValue: player => player.getLevelingGearPlan().seed,
			setValue: (eventID, player, newValue) => {
				const plan = player.getLevelingGearPlan();
				plan.seed = newValue;
				player.setLevelingGearPlan(eventID, plan);
			},
		});
		new NumberPicker<LevelingGearTab>(levelRow.value!, this, {
			id: 'leveling-gear-level',
			label: 'Level',
			labelTooltip: 'The level whose gear the pane shows, and that Equip puts on. A click on the grid picks one too.',
			positive: true,
			changedEvent: tab => tab.selectedLevelEmitter,
			getValue: tab => tab.selectedLevel,
			setValue: (eventID, tab, newValue) => tab.selectLevel(newValue),
		});
		seedRow.value!.append(reroll.value!);
		levelRow.value!.append(equip.value!);
		tippy(reroll.value!, { content: 'Play out another leveling journey, with a new seed.' });
		tippy(equip.value!, { content: () => `Put on the plan's level ${this.selectedLevel} gear in the Gear tab, and make us level ${this.selectedLevel}.` });
		reroll.value!.addEventListener('click', () => {
			const plan = this.player.getLevelingGearPlan();
			plan.seed = Math.floor(Math.random() * 1_000_000) + 1;
			this.player.setLevelingGearPlan(TypedEvent.nextEventID(), plan);
		});
		equip.value!.addEventListener('click', () => this.equipSelectedLevel());

		// Our own items, with Add Item under them like Save Preset under a preset list.
		const pinsBlock = new ContentBlock(side, 'leveling-gear-pins-block', {
			header: {
				title: 'Our Items',
				tooltip: 'Items we have, from a level on. The plan wears each one at its level if it beats what it has, until it finds a better one.',
			},
		});
		const addPin = ref<HTMLButtonElement>();
		const pins = ref<HTMLDivElement>();
		pinsBlock.bodyElement.appendChild(
			<>
				<div className="leveling-gear-pins" ref={pins} />
				<button className="btn btn-outline-primary leveling-gear-add-button" ref={addPin}>
					<i className="fas fa-plus me-1" />
					Add item
				</button>
			</>,
		);
		this.pinsElem = pins.value!;
		addPin.value!.addEventListener('click', () => new AddPinModal(this.simUI.rootElem, this, this.selectedLevel));

		// The selected level's gear, in the Gear tab's pane.
		const paneTitle = ref<HTMLDivElement>();
		const left = ref<HTMLDivElement>();
		const right = ref<HTMLDivElement>();
		paneCol.appendChild(
			<>
				<div className="leveling-gear-pane-title" ref={paneTitle} />
				<div className="gear-picker-root leveling-gear-pane">
					<div className="gear-picker-left tab-panel-col" ref={left} />
					<div className="gear-picker-right tab-panel-col" ref={right} />
				</div>
			</>,
		);
		this.paneTitleElem = paneTitle.value!;
		const leftSlots = [
			ItemSlot.ItemSlotHead,
			ItemSlot.ItemSlotNeck,
			ItemSlot.ItemSlotShoulder,
			ItemSlot.ItemSlotBack,
			ItemSlot.ItemSlotChest,
			ItemSlot.ItemSlotWrist,
			ItemSlot.ItemSlotMainHand,
			ItemSlot.ItemSlotOffHand,
			ItemSlot.ItemSlotRanged,
		];
		const rightSlots = [
			ItemSlot.ItemSlotHands,
			ItemSlot.ItemSlotWaist,
			ItemSlot.ItemSlotLegs,
			ItemSlot.ItemSlotFeet,
			ItemSlot.ItemSlotFinger1,
			ItemSlot.ItemSlotFinger2,
			ItemSlot.ItemSlotTrinket1,
			ItemSlot.ItemSlotTrinket2,
		];
		for (const [column, slots] of [
			[left.value!, leftSlots],
			[right.value!, rightSlots],
		] as Array<[HTMLElement, Array<ItemSlot>]>) {
			for (const slot of slots) {
				const root = (<div className="item-picker-root" />) as HTMLElement;
				column.appendChild(root);
				this.slotRenderers.set(slot, new ItemRenderer(null as unknown as HTMLElement, root, this.player));
			}
		}

		// Every level's gear, a column each.
		const gridBlock = new ContentBlock(gridCol, 'leveling-gear-grid-block', {
			header: {
				title: 'Gear by Level',
				tooltip:
					'A row per slot and a column per level, in the color of what we wear then. A bright cell is a new item at that level, and a gold edge is one of ours.',
			},
		});
		const grid = ref<HTMLDivElement>();
		gridBlock.bodyElement.appendChild(
			<>
				<div className="leveling-gear-legend">
					<span>
						<i className="leveling-gear-swatch leveling-gear-q1" />
						White
					</span>
					<span>
						<i className="leveling-gear-swatch leveling-gear-q2" />
						Green
					</span>
					<span>
						<i className="leveling-gear-swatch leveling-gear-q3" />
						Blue
					</span>
					<span>
						<i className="leveling-gear-swatch leveling-gear-q2 leveling-gear-pinned" />
						Ours
					</span>
					<span className="leveling-gear-legend-note">Bright: a new item that level</span>
				</div>
				<div className="leveling-gear-grid" ref={grid} />
			</>,
		);
		this.gridElem = grid.value!;

		// The tab is only there in Leveling mode.
		const updateMode = () => {
			this.navLink.hidden = !simUI.sim.encounter.isLeveling();
		};
		simUI.sim.encounter.changeEmitter.on(updateMode);
		updateMode();

		simUI.sim.waitForInit().then(() => {
			this.rebuild();
			TypedEvent.onAny([this.player.levelingGearPlanChangeEmitter, this.player.raceChangeEmitter, this.player.epWeightsChangeEmitter]).on(() =>
				this.rebuild(),
			);
		});
	}

	protected buildTabContent() {}

	// The stat weights the plan judges items by at a level: the spec's leveling stat weights
	// preset closest to that level, or the Current EP when the spec has none.
	private weightsAt(level: number): { weights: Stats; name: string } {
		const presets = (this.simUI.individualConfig.presets.epWeights ?? []).filter(preset => preset.group == 'Leveling');
		let best: { weights: Stats; name: string } | undefined;
		let bestGap = Infinity;
		for (const preset of presets) {
			const presetLevel = Number(/Level (\d+)/.exec(preset.name)?.[1] ?? 60);
			if (Math.abs(presetLevel - level) < bestGap) {
				bestGap = Math.abs(presetLevel - level);
				best = { weights: preset.epWeights, name: preset.name };
			}
		}
		return best ?? { weights: this.player.getEpWeights(), name: 'the Current EP' };
	}

	private rebuild() {
		const db = this.simUI.sim.db;
		const playerClass = this.player.getClass();
		const pins = this.player
			.getLevelingGearPlan()
			.pins.map(pin => {
				const item = pin.item && db.getItemById(pin.item.id);
				return item
					? [
							{
								level: pin.level,
								slot: pin.slot,
								item,
								randomSuffix: pin.item?.randomSuffix ? db.getRandomSuffixById(pin.item.randomSuffix) : undefined,
							},
					  ]
					: [];
			})
			.flat();
		this.plan = planLevelingGear({
			items: db.getAllItems(),
			randomSuffix: id => db.getRandomSuffixById(id),
			playerClass,
			classRules: {
				maxArmorTypeAt: level => maxArmorTypeAt(playerClass, level),
				weaponTypes: classToEligibleWeaponTypes[playerClass],
				rangedWeaponTypes: classToEligibleRangedWeaponTypes[playerClass],
				canDualWield: dualWieldClasses.includes(playerClass),
			},
			faction: raceToFaction[this.player.getRace()],
			seed: this.player.getLevelingGearPlan().seed,
			pins,
			weightsAt: level => this.weightsAt(level).weights,
		});
		const names = [...new Set(Array.from({ length: LEVELING_GEAR_MAX_LEVEL }, (_, i) => this.weightsAt(i + 1).name))];
		this.weightsElem.textContent = `We judge items by the ${names.join(', ')} stat weights.`;
		this.buildGrid();
		this.buildLevel();
		this.buildPins();
	}

	selectLevel(level: number) {
		this.selectedLevel = Math.min(Math.max(Math.round(level), 1), LEVELING_GEAR_MAX_LEVEL);
		this.selectedLevelEmitter.emit(TypedEvent.nextEventID());
		this.gridElem.querySelectorAll('.leveling-gear-col-selected').forEach(elem => elem.classList.remove('leveling-gear-col-selected'));
		this.gridElem.querySelectorAll(`[data-level="${this.selectedLevel}"]`).forEach(elem => elem.classList.add('leveling-gear-col-selected'));
		this.buildLevel();
	}

	private levelPlan(level: number): LevelingGearLevel | undefined {
		return this.plan?.levels[level - 1];
	}

	private buildGrid() {
		const plan = this.plan;
		if (!plan) return;
		const table = (<table className="leveling-gear-table" />) as HTMLTableElement;
		const header = document.createElement('tr');
		header.appendChild(document.createElement('th'));
		for (const level of plan.levels) {
			const th = document.createElement('th');
			th.dataset.level = String(level.level);
			th.classList.add('leveling-gear-level-head');
			th.textContent = level.level % 5 == 0 || level.level == 1 ? String(level.level) : '';
			th.addEventListener('click', () => this.selectLevel(level.level));
			header.appendChild(th);
		}
		table.appendChild(header);
		for (const slot of LEVELING_GEAR_SLOTS) {
			const row = document.createElement('tr');
			const label = document.createElement('th');
			label.classList.add('leveling-gear-slot-head');
			label.textContent = slotNames.get(slot) ?? '';
			row.appendChild(label);
			for (const level of plan.levels) {
				const piece = level.gear[slot];
				const cell = document.createElement('td');
				cell.dataset.level = String(level.level);
				cell.classList.add('leveling-gear-cell', `leveling-gear-q${piece?.item.quality ?? 0}`);
				if (level.changed.includes(slot)) cell.classList.add('leveling-gear-changed');
				if (piece?.source == 'pinned') cell.classList.add('leveling-gear-pinned');
				cell.addEventListener('click', () => this.selectLevel(level.level));
				if (piece) {
					tippy(cell, {
						content: `${slotNames.get(slot)} at level ${level.level}: ${pieceName(piece)}, since ${piece.since}${
							piece.source == 'pinned' ? ' (ours)' : ''
						}`,
					});
				}
				row.appendChild(cell);
			}
			table.appendChild(row);
		}
		this.gridElem.replaceChildren(table);
		this.selectLevel(this.selectedLevel);
	}

	private buildLevel() {
		const level = this.levelPlan(this.selectedLevel);
		if (!level) return;
		this.paneTitleElem.textContent = `Level ${this.selectedLevel}`;
		for (const [slot, renderer] of this.slotRenderers) {
			const piece = level.gear[slot];
			renderer.clear();
			renderer.nameElem.textContent = slotNames.get(slot) ?? '';
			setItemQualityCssClass(renderer.nameElem, null);
			if (!piece) {
				renderer.iconElem.style.backgroundImage = `url('${getEmptySlotIconUrl(slot)}')`;
				continue;
			}
			renderer.update(new EquippedItem({ item: piece.item, randomSuffix: piece.randomSuffix ?? null }));
			// The line under the name says since when we wear it and where it came from.
			renderer.enchantElem.textContent = `since ${piece.since}, ${sourceText(piece)}`;
			renderer.enchantElem.classList.remove('hide');
			renderer.enchantElem.classList.add('leveling-gear-since');
			renderer.enchantElem.classList.toggle('leveling-gear-since-new', piece.since == this.selectedLevel);
		}
	}

	private buildPins() {
		const db = this.simUI.sim.db;
		const pins = this.player.getLevelingGearPlan().pins;
		const rows = pins.map((pin, i) => {
			const item = pin.item ? db.getItemById(pin.item.id) : undefined;
			const suffix = pin.item?.randomSuffix ? db.getRandomSuffixById(pin.item.randomSuffix) : undefined;
			const icon = ref<HTMLAnchorElement>();
			const name = ref<HTMLAnchorElement>();
			const remove = ref<HTMLButtonElement>();
			const row = (
				<div className="leveling-gear-pin">
					<a className="leveling-gear-pin-icon" ref={icon} />
					<div className="leveling-gear-pin-labels">
						<a className="leveling-gear-pin-name" ref={name}>
							{item ? `${item.name}${suffix ? ` ${suffix.name}` : ''}` : `Item ${pin.item?.id ?? '?'}`}
						</a>
						<span className="leveling-gear-pin-where">
							{slotNames.get(pin.slot) ?? ''}, from level {String(pin.level)}
						</span>
					</div>
					<button className="btn btn-link link-danger leveling-gear-pin-remove" ref={remove}>
						<i className="fas fa-times" />
					</button>
				</div>
			) as HTMLElement;
			if (item) {
				setItemQualityCssClass(name.value!, item.quality);
				const id = ActionId.fromItemId(item.id, undefined, suffix?.id);
				id.fillAndSet(icon.value!, true, true);
				id.fillAndSet(name.value!, true, false);
			}
			tippy(remove.value!, { content: 'Take it out of the plan' });
			remove.value!.addEventListener('click', () => {
				const plan = this.player.getLevelingGearPlan();
				plan.pins.splice(i, 1);
				this.player.setLevelingGearPlan(TypedEvent.nextEventID(), plan);
			});
			return row;
		});
		this.pinsElem.replaceChildren(...(rows.length > 0 ? rows : [<p className="leveling-gear-none">None yet.</p>]));
	}

	addPin(pin: LevelingGearPin) {
		const plan = this.player.getLevelingGearPlan();
		plan.pins.push(pin);
		plan.pins.sort((a, b) => a.level - b.level);
		this.player.setLevelingGearPlan(TypedEvent.nextEventID(), plan);
	}

	private equipSelectedLevel() {
		const level = this.levelPlan(this.selectedLevel);
		if (!level) return;
		const db = this.simUI.sim.db;
		const gear: Partial<Record<ItemSlot, EquippedItem | null>> = {};
		for (const slot of LEVELING_GEAR_SLOTS) {
			const piece = level.gear[slot];
			gear[slot] = piece ? db.lookupItemSpec(ItemSpec.create({ id: piece.item.id, randomSuffix: piece.randomSuffix?.id ?? 0 })) : null;
		}
		const eventID = TypedEvent.nextEventID();
		TypedEvent.freezeAllAndDo(() => {
			this.player.setLevel(eventID, this.selectedLevel);
			this.player.setGear(eventID, new Gear(gear));
		});
	}

	get db() {
		return this.simUI.sim.db;
	}
}

function pieceName(piece: LevelingGearPiece): string {
	return piece.randomSuffix ? `${piece.item.name} ${piece.randomSuffix.name}` : piece.item.name;
}

function sourceText(piece: LevelingGearPiece): string {
	switch (piece.source) {
		case 'start':
			return 'starting gear';
		case 'pinned':
			return 'ours';
		default:
			return isPlainItemId(piece.item.id) ? 'found, plain' : 'found';
	}
}

// A random suffix's stats, like "+8 Agility, +8 Stamina", so two suffixes of one name tell
// apart.
function suffixStats(stats: Array<number>): string {
	return stats.flatMap((value, stat) => (value ? [`+${value} ${statNames.get(stat) ?? ''}`] : [])).join(', ');
}

// The Add Item window: a level, a slot and an item of ours for the plan.
class AddPinModal extends BaseModal {
	constructor(parent: HTMLElement, tab: LevelingGearTab, startLevel: number) {
		super(parent, 'leveling-gear-add-modal', { title: 'Add Item', disposeOnClose: true });
		const level = ref<HTMLInputElement>();
		const slot = ref<HTMLSelectElement>();
		const search = ref<HTMLInputElement>();
		const results = ref<HTMLDivElement>();
		this.body.appendChild(
			<div className="leveling-gear-add">
				<div className="leveling-gear-add-row">
					<label className="form-label">Level</label>
					<input className="form-control" type="number" min="1" max={String(LEVELING_GEAR_MAX_LEVEL)} value={String(startLevel)} ref={level} />
					<label className="form-label">Slot</label>
					<select className="form-select" ref={slot}>
						{LEVELING_GEAR_SLOTS.map(s => (
							<option value={String(s)}>{slotNames.get(s) ?? ''}</option>
						))}
					</select>
				</div>
				<input className="form-control" type="text" placeholder="Search items by name" ref={search} />
				<div className="leveling-gear-add-results" ref={results} />
			</div>,
		);

		const update = () => {
			const lvl = Number(level.value!.value) || 1;
			const itemSlot = Number(slot.value!.value) as ItemSlot;
			const text = search.value!.value.trim().toLowerCase();
			// A random suffix item is one entry per suffix, like "of the Monkey", since the suffix
			// is part of the item we have.
			const choices = tab.db
				.getItems(itemSlot)
				.filter(item => item.requiredLevel <= lvl && getEligibleItemSlots(item).includes(itemSlot))
				.flatMap(item =>
					item.randomSuffixOptions.length == 0
						? [{ item, suffixId: 0, name: item.name, stats: '' }]
						: item.randomSuffixOptions.map(id => {
								const suffix = tab.db.getRandomSuffixById(id);
								return { item, suffixId: id, name: `${item.name} ${suffix?.name ?? ''}`, stats: suffix ? suffixStats(suffix.stats) : '' };
						  }),
				)
				.filter(choice => !text || choice.name.toLowerCase().includes(text))
				.sort((a, b) => b.item.ilvl - a.item.ilvl)
				.slice(0, 60);
			results.value!.replaceChildren(
				...choices.map(({ item, suffixId, name, stats }) => {
					const elem = (
						<a href="javascript:void(0)" className="leveling-gear-add-item">
							{name}{' '}
							<span className="text-muted">
								({stats ? `${stats}, ` : ''}ilvl {String(item.ilvl)}, needs {String(item.requiredLevel)})
							</span>
						</a>
					) as HTMLAnchorElement;
					setItemQualityCssClass(elem, item.quality ?? ItemQuality.ItemQualityCommon);
					elem.addEventListener('click', () => {
						tab.addPin(LevelingGearPin.create({ level: lvl, slot: itemSlot, item: ItemSpec.create({ id: item.id, randomSuffix: suffixId }) }));
						this.close();
					});
					return elem;
				}),
			);
		};
		[level, slot, search].forEach(input => input.value!.addEventListener('input', update));
		update();
		this.open();
		search.value!.focus();
	}
}
