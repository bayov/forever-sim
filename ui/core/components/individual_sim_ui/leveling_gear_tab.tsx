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
import { NumberPicker } from '../number_picker';
import { SimTab } from '../sim_tab';

// The Leveling Gear tab shows the leveling gear plan (leveling_gear.ts): what we wear at each
// level from 1 to 60 when we gear up as we go. It's only there in Leveling mode.
//
// The grid has a row per slot and a column per level, in the color of the item we wear then,
// with a mark where it changed. A click on a column shows that level's gear below it, and
// Equip puts it on in the Gear tab, at that level. Add Item puts an item of our own into the
// plan from a level on, and the plan wears it until it finds a better one.
export class LevelingGearTab extends SimTab {
	protected simUI: IndividualSimUI<Spec>;
	private readonly player: Player<Spec>;
	private selectedLevel = 30;
	private plan: LevelingGearPlan | null = null;
	private readonly selectedLevelEmitter = new TypedEvent<void>('LevelingGearSelectedLevel');

	private readonly gridElem: HTMLElement;
	private readonly levelElem: HTMLElement;
	private readonly pinsElem: HTMLElement;
	private readonly weightsElem: HTMLElement;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		super(parentElem, simUI, { identifier: 'leveling-gear-tab', title: 'Leveling Gear' });
		this.simUI = simUI;
		this.player = simUI.player;

		const controls = ref<HTMLDivElement>();
		const grid = ref<HTMLDivElement>();
		const level = ref<HTMLDivElement>();
		const pins = ref<HTMLDivElement>();
		const weights = ref<HTMLDivElement>();
		const reroll = ref<HTMLButtonElement>();
		const equip = ref<HTMLButtonElement>();
		const addPin = ref<HTMLButtonElement>();
		this.contentContainer.appendChild(
			<div className="leveling-gear">
				<div className="leveling-gear-controls" ref={controls}>
					<button className="btn btn-outline-primary" ref={reroll}>
						<i className="fas fa-dice me-1" />
						Reroll
					</button>
					<button className="btn btn-outline-primary" ref={equip}>
						<i className="fas fa-shirt me-1" />
						Equip
					</button>
					<button className="btn btn-outline-primary" ref={addPin}>
						<i className="fas fa-plus me-1" />
						Add Item
					</button>
					<span className="leveling-gear-weights" ref={weights} />
				</div>
				<div className="leveling-gear-grid" ref={grid} />
				<div className="leveling-gear-bottom">
					<div className="leveling-gear-level" ref={level} />
					<div className="leveling-gear-pins" ref={pins} />
				</div>
			</div>,
		);
		this.gridElem = grid.value!;
		this.levelElem = level.value!;
		this.pinsElem = pins.value!;
		this.weightsElem = weights.value!;

		new NumberPicker<Player<Spec>>(controls.value!, this.player, {
			id: 'leveling-gear-seed',
			label: 'Seed',
			labelTooltip: 'The plan is one random leveling journey. Another seed, or Reroll, plays out another one.',
			inline: true,
			changedEvent: player => player.levelingGearPlanChangeEmitter,
			getValue: player => player.getLevelingGearPlan().seed,
			setValue: (eventID, player, newValue) => {
				const plan = player.getLevelingGearPlan();
				plan.seed = newValue;
				player.setLevelingGearPlan(eventID, plan);
			},
		});
		new NumberPicker<LevelingGearTab>(controls.value!, this, {
			id: 'leveling-gear-level',
			label: 'Level',
			labelTooltip: 'The level whose gear shows below the grid, and that Equip puts on.',
			inline: true,
			positive: true,
			changedEvent: tab => tab.selectedLevelEmitter,
			getValue: tab => tab.selectedLevel,
			setValue: (eventID, tab, newValue) => tab.selectLevel(newValue),
		});
		controls.value!.prepend(reroll.value!, equip.value!, addPin.value!);
		controls.value!.append(this.weightsElem);

		tippy(reroll.value!, { content: 'Play out another leveling journey, with a new seed.' });
		tippy(equip.value!, { content: () => `Put on the plan's level ${this.selectedLevel} gear in the Gear tab, and make us level ${this.selectedLevel}.` });
		tippy(addPin.value!, { content: 'Put an item of our own into the plan, from a level on. The plan wears it until it finds a better one.' });
		reroll.value!.addEventListener('click', () => {
			const plan = this.player.getLevelingGearPlan();
			plan.seed = Math.floor(Math.random() * 1_000_000) + 1;
			this.player.setLevelingGearPlan(TypedEvent.nextEventID(), plan);
		});
		equip.value!.addEventListener('click', () => this.equipSelectedLevel());
		addPin.value!.addEventListener('click', () => new AddPinModal(this.simUI.rootElem, this, this.selectedLevel));

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
		this.weightsElem.textContent = `Items are judged by ${names.join(', ')}.`;
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
			th.textContent = level.level % 5 == 0 || level.level == 1 ? String(level.level) : '';
			th.addEventListener('click', () => this.selectLevel(level.level));
			header.appendChild(th);
		}
		table.appendChild(header);
		for (const slot of LEVELING_GEAR_SLOTS) {
			const row = document.createElement('tr');
			const label = document.createElement('th');
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
				if (piece && level.changed.includes(slot)) {
					tippy(cell, { content: `Level ${level.level}: ${pieceName(piece)}${piece.source == 'pinned' ? ' (ours)' : ''}` });
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
		const rows = LEVELING_GEAR_SLOTS.map(slot => {
			const piece = level.gear[slot];
			const icon = ref<HTMLAnchorElement>();
			const name = ref<HTMLAnchorElement>();
			const row = (
				<tr>
					<td className="leveling-gear-slot">{slotNames.get(slot) ?? ''}</td>
					<td>
						<a className="leveling-gear-icon" ref={icon} />
						<a className="leveling-gear-name" ref={name}>
							{piece ? pieceName(piece) : '-'}
						</a>
					</td>
					<td className="leveling-gear-since">{piece ? `since ${piece.since}` : ''}</td>
					<td className="leveling-gear-source">{piece ? sourceText(piece) : ''}</td>
				</tr>
			) as HTMLElement;
			if (piece) {
				setItemQualityCssClass(name.value!, piece.item.quality);
				if (!isPlainItemId(piece.item.id)) {
					const id = ActionId.fromItemId(piece.item.id, undefined, piece.randomSuffix?.id);
					id.fillAndSet(icon.value!, true, true);
					id.fillAndSet(name.value!, true, false);
				} else {
					icon.value!.style.backgroundImage = `url('https://wow.zamimg.com/images/wow/icons/large/inv_misc_questionmark.jpg')`;
					tippy(name.value!, { content: `A plain white item we made up: ${plainStats(piece.item)}.` });
				}
			}
			return row;
		});
		this.levelElem.replaceChildren(
			<div className="leveling-gear-level-title">
				Level {String(this.selectedLevel)}, {String(level.finds)} {level.finds == 1 ? 'find' : 'finds'} this level
			</div>,
			<table className="leveling-gear-level-table">{rows}</table>,
		);
	}

	private buildPins() {
		const db = this.simUI.sim.db;
		const pins = this.player.getLevelingGearPlan().pins;
		const rows = pins.map((pin, i) => {
			const item = pin.item ? db.getItemById(pin.item.id) : undefined;
			const suffix = pin.item?.randomSuffix ? db.getRandomSuffixById(pin.item.randomSuffix) : undefined;
			const remove = ref<HTMLButtonElement>();
			const row = (
				<tr>
					<td>Level {String(pin.level)}</td>
					<td>{slotNames.get(pin.slot) ?? ''}</td>
					<td>{item ? `${item.name}${suffix ? ` ${suffix.name}` : ''}` : `item ${pin.item?.id ?? '?'}`}</td>
					<td>
						<button className="btn btn-link link-danger p-0" ref={remove}>
							<i className="fas fa-times" />
						</button>
					</td>
				</tr>
			) as HTMLElement;
			remove.value!.addEventListener('click', () => {
				const plan = this.player.getLevelingGearPlan();
				plan.pins.splice(i, 1);
				this.player.setLevelingGearPlan(TypedEvent.nextEventID(), plan);
			});
			return row;
		});
		this.pinsElem.replaceChildren(
			<div className="leveling-gear-level-title">Our items</div>,
			pins.length > 0 ? (
				<table className="leveling-gear-pins-table">{rows}</table>
			) : (
				<div className="leveling-gear-none">None yet. Add Item puts one in.</div>
			),
		);
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

function plainStats(item: Item): string {
	if (item.weaponSpeed > 0) {
		return `${item.weaponDamageMin} - ${item.weaponDamageMax} damage, ${item.weaponSpeed.toFixed(2)} speed`;
	}
	return `${item.stats[26]} armor`;
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
