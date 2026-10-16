import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { setItemQualityCssClass } from '../../css_utils';
import { IndividualSimUI } from '../../individual_sim_ui';
import {
	isPlainItemId,
	LEVELING_GEAR_MAX_LEVEL,
	LEVELING_GEAR_SLOTS,
	LevelingGearLevel,
	LevelingGearPiece,
	LevelingGearPlan,
	planLevelingGear,
} from '../../leveling_gear';
import { Player } from '../../player';
import { ItemSlot, ItemSpec, Spec } from '../../proto/common';
import { LevelingGearPin, LevelingGearPlanSettings } from '../../proto/ui';
import { ActionId } from '../../proto_utils/action_id';
import { EquippedItem } from '../../proto_utils/equipped_item';
import { Gear } from '../../proto_utils/gear';
import { slotNames } from '../../proto_utils/names';
import { Stats } from '../../proto_utils/stats';
import { classToEligibleRangedWeaponTypes, classToEligibleWeaponTypes, dualWieldClasses, maxArmorTypeAt, raceToFaction } from '../../proto_utils/utils';
import { TypedEvent } from '../../typed_event';
import { ContentBlock } from '../content_block';
import GearPicker from '../gear_picker/gear_picker';
import { SelectorModalTabs } from '../gear_picker/selector_modal';
import { NumberPicker } from '../number_picker';
import { LevelingPlanCharacter } from './leveling_plan_character';

// The Gear part of the Leveling Plan tab shows the leveling gear plan (leveling_gear.ts): what
// we wear at each level from 1 to 60 when we gear up as we go.
//
// On top are the plan's settings, the items we put in by hand, and the grid with a row per slot
// and a column per level, in the color of the item we wear then. Under them is the gear of the
// selected level in the Gear tab's gear pane, on the plan's character at that level
// (LevelingPlanCharacter). A click on a slot opens the Gear tab's item list, with the items
// that level can wear sorted by the stat weights of that level, and the item we pick there
// becomes one of ours from that level on.
export class LevelingGearSection {
	readonly rootElem: HTMLElement;
	private readonly simUI: IndividualSimUI<Spec>;
	private readonly player: Player<Spec>;
	private readonly character: LevelingPlanCharacter;
	private readonly characterKey: (level: number) => string;
	private selectedLevel = 20;
	private plan: LevelingGearPlan | null = null;
	private active = false;
	private loadedKey = '';
	private readonly selectedLevelEmitter = new TypedEvent<void>('LevelingGearSelectedLevel');
	// Fires when the plan's gear changes.
	readonly planChangeEmitter = new TypedEvent<void>('LevelingGearPlanned');
	private readonly onGrind: (level: number) => void;

	private readonly gridElem: HTMLElement;
	private readonly paneTitleElem: HTMLElement;
	private readonly paneNoteElem: HTMLElement;
	private readonly pinsElem: HTMLElement;
	private readonly weightsElem: HTMLElement;
	private readonly gearPicker: GearPicker;
	// The line under each slot's item that says since when we wear it.
	private readonly sinceElems = new Map<ItemSlot, HTMLElement>();

	constructor(
		parentElem: HTMLElement,
		simUI: IndividualSimUI<Spec>,
		character: LevelingPlanCharacter,
		characterKey: (level: number) => string,
		onGrind: (level: number) => void,
	) {
		this.simUI = simUI;
		this.player = simUI.player;
		this.character = character;
		this.characterKey = characterKey;
		this.onGrind = onGrind;
		this.rootElem = (<div className="leveling-plan-gear" />) as HTMLElement;
		parentElem.appendChild(this.rootElem);

		const side = (<div className="leveling-gear-side tab-panel-col" />) as HTMLElement;
		const paneCol = (<div className="leveling-gear-pane-col" />) as HTMLElement;
		const gridCol = (<div className="leveling-gear-grid-col" />) as HTMLElement;
		this.rootElem.appendChild(
			<div className="leveling-gear">
				<div className="leveling-gear-top">
					{side}
					{gridCol}
				</div>
				{paneCol}
			</div>,
		);

		// The plan's settings: the stat weights it judges items by and its seed.
		const planBlock = new ContentBlock(side, 'leveling-gear-plan-block', {
			header: {
				title: 'Leveling Gear Plan',
				tooltip:
					'One random leveling journey. Each level, each slot comes across an item now and then, mostly white early and green later with a few blues, and we wear it only when it beats what we have.',
			},
		});
		const reroll = ref<HTMLButtonElement>();
		const seedRow = ref<HTMLDivElement>();
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
		seedRow.value!.append(reroll.value!);
		tippy(reroll.value!, { content: 'Play out another leveling journey, with a new seed.' });
		reroll.value!.addEventListener('click', () => {
			const plan = this.player.getLevelingGearPlan();
			plan.seed = Math.floor(Math.random() * 1_000_000) + 1;
			this.player.setLevelingGearPlan(TypedEvent.nextEventID(), plan);
		});

		// Our own items. We add them in the gear pane below.
		const pinsBlock = new ContentBlock(side, 'leveling-gear-pins-block', {
			header: {
				title: 'Our Items',
				tooltip: 'Items we have, from a level on. The plan wears each one at its level if it beats what it has, until it finds a better one.',
			},
		});
		const pins = ref<HTMLDivElement>();
		pinsBlock.bodyElement.appendChild(<div className="leveling-gear-pins" ref={pins} />);
		this.pinsElem = pins.value!;

		// The selected level's gear, in the Gear tab's pane, with the level and Grind over it.
		const paneTitle = ref<HTMLSpanElement>();
		const paneNote = ref<HTMLSpanElement>();
		const levelPicker = ref<HTMLDivElement>();
		const prev = ref<HTMLButtonElement>();
		const next = ref<HTMLButtonElement>();
		const grind = ref<HTMLButtonElement>();
		const pane = ref<HTMLDivElement>();
		paneCol.appendChild(
			<>
				<div className="leveling-gear-pane-head">
					<button className="btn btn-link leveling-gear-step" ref={prev}>
						<i className="fas fa-chevron-left" />
					</button>
					<span className="leveling-gear-pane-title" ref={paneTitle} />
					<button className="btn btn-link leveling-gear-step" ref={next}>
						<i className="fas fa-chevron-right" />
					</button>
					<div className="leveling-gear-level-picker" ref={levelPicker} />
					<span className="leveling-gear-pane-note" ref={paneNote} />
					<button className="btn btn-primary leveling-gear-grind" ref={grind}>
						<i className="fas fa-person-running me-1" />
						Grind
					</button>
				</div>
				<div className="leveling-gear-pane" ref={pane} />
			</>,
		);
		this.paneTitleElem = paneTitle.value!;
		this.paneNoteElem = paneNote.value!;
		new NumberPicker<LevelingGearSection>(levelPicker.value!, this, {
			id: 'leveling-gear-level',
			label: 'Level',
			labelTooltip: 'The level whose gear the pane shows, and that Grind takes to the Grinding mode. A click on the grid picks one too.',
			positive: true,
			changedEvent: tab => tab.selectedLevelEmitter,
			getValue: tab => tab.selectedLevel,
			setValue: (_eventID, tab, newValue) => tab.selectLevel(newValue),
		});
		prev.value!.addEventListener('click', () => this.selectLevel(this.selectedLevel - 1));
		next.value!.addEventListener('click', () => this.selectLevel(this.selectedLevel + 1));
		tippy(grind.value!, {
			content: () => `Go to the Grinding mode as the plan's level ${this.selectedLevel}, with its gear, talents and rotation.`,
		});
		grind.value!.addEventListener('click', () => this.onGrind(this.selectedLevel));

		// The item list has what the level can wear. An item with no required level in our
		// data, like many quest rewards from the wowhead listing, counts as one of its item level.
		this.gearPicker = new GearPicker(pane.value!, simUI, character.player, {
			disabledTabs: [SelectorModalTabs.Enchants],
			statWeights: false,
			itemFilter: item => (item.requiredLevel || item.ilvl - 5) <= this.selectedLevel,
		});
		for (const picker of this.gearPicker.itemPickers) {
			const since = (<span className="leveling-gear-since" />) as HTMLElement;
			picker.rootElem.querySelector('.item-picker-labels-container')?.appendChild(since);
			this.sinceElems.set(picker.slot, since);
		}
		// An item we pick in the pane is one of ours from this level on.
		character.player.gearChangeEmitter.on(() => {
			if (!character.applying) this.pickedInPane();
		});
		character.loadEmitter.on(() => this.buildLevel());

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

		simUI.sim.waitForInit().then(() => {
			this.rebuild();
			TypedEvent.onAny([this.player.levelingGearPlanChangeEmitter, this.player.raceChangeEmitter, this.player.epWeightsChangeEmitter]).on(() =>
				this.rebuild(),
			);
		});
	}

	// The plan we show, or null before the database has loaded.
	getPlan(): LevelingGearPlan | null {
		return this.plan;
	}

	// The plan's gear at a level, as the Gear tab would have it.
	gearAt(level: number): Gear | null {
		const levelPlan = this.levelPlan(level);
		if (!levelPlan) return null;
		const db = this.simUI.sim.db;
		const gear: Partial<Record<ItemSlot, EquippedItem | null>> = {};
		for (const slot of LEVELING_GEAR_SLOTS) {
			const piece = levelPlan.gear[slot];
			gear[slot] = piece ? db.lookupItemSpec(ItemSpec.create({ id: piece.item.id, randomSuffix: piece.randomSuffix?.id ?? 0 })) : null;
		}
		return new Gear(gear);
	}

	// The spec's stat weights presets for the plan by their level, lowest first.
	//
	// They're the "Leveling Gear Plan" presets, measured on the plan's own gear. A spec without
	// them falls back to its "Leveling" presets.
	private levelingWeights(): Array<{ level: number; weights: Stats }> {
		const presets = this.simUI.individualConfig.presets.epWeights ?? [];
		const planPresets = presets.filter(preset => preset.group == 'Leveling Gear Plan');
		return (planPresets.length ? planPresets : presets.filter(preset => preset.group == 'Leveling'))
			.map(preset => ({ level: Number(/Level (\d+)/.exec(preset.name)?.[1] ?? 60), weights: preset.epWeights }))
			.sort((a, b) => a.level - b.level);
	}

	// The stat weights the plan judges items by at a level.
	//
	// They come from the leveling sim, which measures them at a few levels. Between two of those levels we blend them by how close we are to
	// each, so a level 22 judges items mostly by the level 20 weights and a bit by the level 25
	// ones. Below the lowest and above the highest we take that one. A spec without leveling
	// presets uses the Current EP.
	private weightsAt(level: number): Stats {
		const presets = this.levelingWeights();
		if (!presets.length) return this.player.getEpWeights();
		const above = presets.findIndex(preset => preset.level >= level);
		if (above == -1) return presets[presets.length - 1].weights;
		if (above == 0 || presets[above].level == level) return presets[above].weights;
		const lo = presets[above - 1];
		const hi = presets[above];
		const t = (level - lo.level) / (hi.level - lo.level);
		return lo.weights.scale(1 - t).add(hi.weights.scale(t));
	}

	private weightsNote(): string {
		const levels = this.levelingWeights().map(preset => preset.level);
		if (!levels.length) return 'We judge items by the Current EP.';
		if (levels.length == 1) return `We judge items by the Level ${levels[0]} Leveling stat weights.`;
		return `We judge items by the leveling sim's stat weights at levels ${levels.slice(0, -1).join(', ')} and ${
			levels[levels.length - 1]
		}, blended between them by level.`;
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
			weightsAt: level => this.weightsAt(level),
			zoneName: zoneId => db.getZone(zoneId)?.name,
		});
		this.weightsElem.textContent = this.weightsNote();
		this.buildGrid();
		this.buildPins();
		this.planChangeEmitter.emit(TypedEvent.nextEventID());
	}

	// The tab calls it when this part opens or closes, and when the talents or rotations change.
	setActive(active: boolean) {
		this.active = active;
		this.syncCharacter();
	}

	refresh() {
		this.syncCharacter();
	}

	// Makes the pane's character the plan's at the selected level, unless it already is.
	private syncCharacter() {
		this.buildLevel();
		if (!this.active || !this.plan) return;
		const key = this.characterKey(this.selectedLevel);
		if (key == this.loadedKey) return;
		this.loadedKey = key;
		this.character.load(this.selectedLevel, { weights: this.weightsAt(this.selectedLevel) });
	}

	selectLevel(level: number) {
		this.selectedLevel = Math.min(Math.max(Math.round(level), 1), LEVELING_GEAR_MAX_LEVEL);
		this.selectedLevelEmitter.emit(TypedEvent.nextEventID());
		this.gridElem.querySelectorAll('.leveling-gear-col-selected').forEach(elem => elem.classList.remove('leveling-gear-col-selected'));
		this.gridElem.querySelectorAll(`[data-level="${this.selectedLevel}"]`).forEach(elem => elem.classList.add('leveling-gear-col-selected'));
		this.syncCharacter();
	}

	// An item we picked in the pane's item list for a slot becomes ours from the selected
	// level, in place of one of ours we had put at that level and slot. Taking an item out
	// takes ours out, and otherwise the plan's item comes back, because the plan always wears
	// what it found.
	private pickedInPane() {
		const level = this.levelPlan(this.selectedLevel);
		if (!level) return;
		const gearPlan = this.player.getLevelingGearPlan();
		let changed = false;
		for (const slot of LEVELING_GEAR_SLOTS) {
			const picked = this.character.player.getEquippedItem(slot);
			const piece = level.gear[slot];
			if ((picked?.item.id ?? 0) == (piece?.item.id ?? 0) && (picked?.randomSuffix?.id ?? 0) == (piece?.randomSuffix?.id ?? 0)) continue;
			gearPlan.pins = gearPlan.pins.filter(pin => !(pin.level == this.selectedLevel && pin.slot == slot));
			if (picked) {
				gearPlan.pins.push(
					LevelingGearPin.create({
						level: this.selectedLevel,
						slot,
						item: ItemSpec.create({ id: picked.item.id, randomSuffix: picked.randomSuffix?.id ?? 0 }),
					}),
				);
			}
			changed = true;
		}
		if (!changed) return;
		gearPlan.pins.sort((a, b) => a.level - b.level);
		// When nothing in the plan changes, like taking out an item the plan found, we put the
		// plan's gear back in the pane.
		this.loadedKey = '';
		if (!LevelingGearPlanSettings.equals(gearPlan, this.player.getLevelingGearPlan())) {
			this.player.setLevelingGearPlan(TypedEvent.nextEventID(), gearPlan);
		} else {
			this.syncCharacter();
		}
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
		this.paneTitleElem.textContent = `Level ${this.selectedLevel}`;
		const changed = level?.changed.filter(slot => LEVELING_GEAR_SLOTS.includes(slot)).length ?? 0;
		this.paneNoteElem.textContent = level
			? `${changed ? `${changed} new item${changed == 1 ? '' : 's'} at this level. ` : ''}A click on a slot picks an item of ours from this level on.`
			: '';
		for (const [slot, since] of this.sinceElems) {
			const piece = level?.gear[slot];
			// The line under the name says since when we wear it and where it came from.
			since.textContent = piece ? `since ${piece.since}, ${sourceText(piece)}` : '';
			since.classList.toggle('leveling-gear-since-new', piece?.since == this.selectedLevel);
			since.classList.toggle('leveling-gear-since-ours', piece?.source == 'pinned');
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
		this.pinsElem.replaceChildren(
			...(rows.length > 0 ? rows : [<p className="leveling-gear-none">None yet. A click on a slot of the gear pane below adds one at its level.</p>]),
		);
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
		case 'quest':
			return 'quest reward';
		case 'dungeon':
			return piece.dungeon ?? 'dungeon';
		default:
			return isPlainItemId(piece.item.id) ? 'plain drop' : 'drop';
	}
}
