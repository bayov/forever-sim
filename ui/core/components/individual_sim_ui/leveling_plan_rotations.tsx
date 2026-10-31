import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import { freeRotationName, libraryRotation, rangeLevels, sortedRanges } from '../../leveling_plan';
import { Player } from '../../player';
import { APLAction, APLRotation } from '../../proto/apl';
import { Spec, WeaponImbue } from '../../proto/common';
import { LevelingPlanSettings, LevelingRotation, LevelingRotationPreset } from '../../proto/ui';
import { ActionId } from '../../proto_utils/action_id';
import { TypedEvent } from '../../typed_event';
import { ContentBlock } from '../content_block';
import { APLRotationPicker } from './apl_rotation_picker';
import { LevelingPlanCharacter } from './leveling_plan_character';

// The shaman imbues a level range can come with.
const SHAMAN_IMBUES: Array<{ name: string; value: WeaponImbue }> = [
	{ name: "Settings tab's imbue", value: WeaponImbue.WeaponImbueUnknown },
	{ name: 'Rockbiter Weapon', value: WeaponImbue.RockbiterWeapon },
	{ name: 'Flametongue Weapon', value: WeaponImbue.FlametongueWeapon },
	{ name: 'Frostbrand Weapon', value: WeaponImbue.FrostbrandWeapon },
	{ name: 'Windfury Weapon', value: WeaponImbue.WindfuryWeapon },
];

// The colors of the plan's rotations, by their place in the library, so a level range and its
// rotation have the same color in the timeline and in the list.
const ROTATION_COLORS = ['#4e9bff', '#f0a030', '#3cc47c', '#c86bff', '#ff6b6b', '#2ec4c4', '#d8c64a', '#ff8fc8'];

// How many of a rotation's spells the list shows.
const LIST_SPELLS = 6;

// The Rotations part of the Leveling Plan tab.
//
// The plan has its own rotations. They start as copies of the spec's rotation presets, and we
// can add more, rename them and change them in the editor on the right. The list on the left has
// a card per rotation in the order we play them, each from its level until the next one's. Under
// the hood each card is a level range of the plan, and the rotation it plays. A shaman's range
// also says which weapon imbue we use, because Windfury Weapon only comes at level 30. We set it in
// the editor. Rotations that no range plays, like ones from an older plan, come last.
//
// The editor is the Rotation tab's, on the plan's character at one level of the plan
// (LevelingPlanCharacter), because its spell lists come from the character's level and talents.
// The character of the settings doesn't change.
export class LevelingRotationsSection {
	readonly rootElem: HTMLElement;
	private readonly simUI: IndividualSimUI<Spec>;
	private readonly player: Player<Spec>;
	private readonly character: LevelingPlanCharacter;
	private readonly characterKey: (level: number) => string;
	private readonly hasShamanImbue: boolean;

	private readonly libraryElem: HTMLElement;
	private readonly newMenuElem: HTMLElement;
	private readonly editorHeadElem: HTMLElement;
	private readonly editorElem: HTMLElement;
	private readonly editorEmptyElem: HTMLElement;

	// The name of the rotation in the editor.
	private selected = '';
	// The level range we opened it from, by its place in sortedRanges, or -1 for none.
	private selectedRange = -1;
	// The level of the plan the editor's character is at. 0 means the first level we play the
	// selected rotation at.
	private editLevel = 0;
	private active = false;
	// What the editor's character was loaded with, see characterKey.
	private loadedKey = '';

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>, character: LevelingPlanCharacter, characterKey: (level: number) => string) {
		this.simUI = simUI;
		this.player = simUI.player;
		this.character = character;
		this.characterKey = characterKey;
		this.hasShamanImbue = 'shamanImbue' in (this.player.getSpecOptions() as object);
		this.rootElem = (<div className="leveling-plan-rotations" />) as HTMLElement;
		parentElem.appendChild(this.rootElem);

		const main = (<div className="leveling-plan-rotations-main" />) as HTMLElement;
		this.rootElem.appendChild(main);
		const libraryCol = (<div className="leveling-plan-library-col" />) as HTMLElement;
		const editorCol = (<div className="leveling-plan-editor-col" />) as HTMLElement;
		main.append(libraryCol, editorCol);

		// The level ranges and the plan's rotations, with the buttons to add them under the list
		// like Save Preset under a preset list.
		const libraryBlock = new ContentBlock(libraryCol, 'leveling-plan-library-block', {
			header: {
				title: 'Rotations',
				tooltip:
					"The plan's rotations, each played from its level until the next one's. A click opens one in the editor. Changing them here leaves the spec's presets and the Rotation tab alone.",
			},
		});
		const library = ref<HTMLDivElement>();
		const newMenu = ref<HTMLUListElement>();
		libraryBlock.bodyElement.appendChild(
			<>
				<div className="leveling-plan-library" ref={library} />
				<div className="leveling-plan-buttons">
					<div className="dropdown leveling-plan-new-rotation">
						<button
							className="btn btn-outline-primary dropdown-toggle"
							type="button"
							dataset={{ bsToggle: 'dropdown' }}
							attributes={{ 'aria-expanded': 'false' }}>
							<i className="fas fa-plus me-1" />
							Rotation
						</button>
						<ul className="dropdown-menu" ref={newMenu} />
					</div>
				</div>
			</>,
		);
		this.libraryElem = library.value!;
		this.newMenuElem = newMenu.value!;

		const editorBlock = new ContentBlock(editorCol, 'leveling-plan-editor-block', {
			header: { title: 'Editor' },
		});
		const editorHead = ref<HTMLDivElement>();
		const editor = ref<HTMLDivElement>();
		const editorEmpty = ref<HTMLParagraphElement>();
		editorBlock.bodyElement.appendChild(
			<>
				<div className="leveling-plan-editor-head" ref={editorHead} />
				<p className="leveling-plan-note hide" ref={editorEmpty}>
					Pick a rotation on the left to change it here.
				</p>
				<div className="leveling-plan-rotation-editor" ref={editor} />
			</>,
		);
		this.editorHeadElem = editorHead.value!;
		this.editorElem = editor.value!;
		this.editorEmptyElem = editorEmpty.value!;
		new APLRotationPicker(this.editorElem, simUI, character.player, simUI.individualConfig.prepullActionNote, false);

		// A change in the editor goes into the plan's rotation of that name.
		character.player.rotationChangeEmitter.on(() => {
			if (character.applying) return;
			const preset = this.selectedRotation();
			if (!preset || APLRotation.equals(preset.rotation ?? APLRotation.create(), character.player.aplRotation)) return;
			this.change(plan => (libraryRotation(plan, preset.name)!.rotation = APLRotation.clone(character.player.aplRotation)));
		});
		this.player.levelingPlanChangeEmitter.on(() => {
			this.render();
			this.syncEditor();
		});
		this.render();
	}

	// The tab calls it when this part opens or closes, and when the plan's gear changes.
	setActive(active: boolean) {
		this.active = active;
		this.syncEditor();
	}

	refresh() {
		this.syncEditor();
	}

	// The plan to read, see readLevelingPlan. A change starts from a copy, see change().
	private plan(): LevelingPlanSettings {
		return this.player.readLevelingPlan();
	}

	private selectedRotation(): LevelingRotationPreset | undefined {
		return libraryRotation(this.plan(), this.selected);
	}

	private specPresets() {
		return this.simUI.individualConfig.presets.rotations.filter(preset => preset.rotation.rotation);
	}

	private specPreset(name: string) {
		return this.specPresets().find(preset => preset.name == name);
	}

	// The levels that play a rotation, like "10-14, 25-29".
	private usedAt(plan: LevelingPlanSettings, name: string): Array<{ from: number; to: number }> {
		return sortedRanges(plan)
			.map((range, i) => ({ range, levels: rangeLevels(plan, i) }))
			.filter(({ range, levels }) => range.name == name && levels.to >= levels.from)
			.map(({ levels }) => levels);
	}

	// The level the editor's character is at.
	private characterLevel(plan: LevelingPlanSettings): number {
		const first = this.usedAt(plan, this.selected)[0]?.from ?? plan.fromLevel;
		const level = this.editLevel || first;
		return Math.min(Math.max(level, plan.fromLevel), plan.toLevel - 1);
	}

	// Puts the selected rotation in the editor, on the plan's character at the level we edit it
	// at. When only the rotation changed, like after an edit or a revert, we keep the character
	// and give it the rotation.
	private syncEditor() {
		const preset = this.selectedRotation();
		this.editorElem.classList.toggle('hide', !preset);
		this.editorEmptyElem.classList.toggle('hide', !!preset);
		if (!this.active || !preset) return;
		const rotation = preset.rotation ?? APLRotation.create();
		const level = this.characterLevel(this.plan());
		const key = this.characterKey(level);
		if (key != this.loadedKey) {
			this.loadedKey = key;
			this.character.load(level, { rotation });
		} else if (!APLRotation.equals(rotation, this.character.player.aplRotation)) {
			this.character.player.setAplRotation(TypedEvent.nextEventID(), rotation);
		}
	}

	private select(name: string) {
		this.selectedRange = -1;
		if (name == this.selected) {
			this.render();
			return;
		}
		this.selected = name;
		this.editLevel = 0;
		this.render();
		this.syncEditor();
	}

	// Opens a range's rotation in the editor, on the character at the range's first level.
	private selectRange(i: number) {
		const plan = this.plan();
		const range = sortedRanges(plan)[i];
		if (!range) return;
		this.selectedRange = i;
		this.selected = range.name;
		this.editLevel = rangeLevels(plan, i).from;
		this.render();
		this.syncEditor();
	}

	private change(change: (plan: LevelingPlanSettings) => void) {
		const plan = this.player.getLevelingPlan();
		plan.rotations = sortedRanges(plan);
		change(plan);
		this.player.setLevelingPlan(TypedEvent.nextEventID(), plan);
	}

	// Adds a rotation to the plan and opens it. It plays the second half of the open rotation's
	// levels, or of the longest one's when the open one plays a single level.
	private addRotation(name: string, rotation: APLRotation) {
		const plan = this.plan();
		const free = freeRotationName(plan, name);
		const ranges = sortedRanges(plan);
		const split = this.rangeToSplit(plan);
		this.change(plan => {
			plan.rotationLibrary.push(LevelingRotationPreset.create({ name: free, rotation: APLRotation.clone(rotation) }));
			if (!ranges.length) {
				plan.rotations.push(LevelingRotation.create({ level: plan.fromLevel, name: free }));
			} else if (split) {
				plan.rotations.push(LevelingRotation.clone({ ...ranges[split.index], level: split.level, name: free }));
			}
		});
		const index = sortedRanges(this.plan()).findIndex(range => range.name == free);
		if (index >= 0) this.selectRange(index);
		else this.select(free);
	}

	private rangeToSplit(plan: LevelingPlanSettings): { index: number; level: number } | undefined {
		const length = (i: number) => {
			const { from, to } = rangeLevels(plan, i);
			return to - from + 1;
		};
		let index = this.selectedRange;
		if (!(index >= 0 && length(index) >= 2)) {
			index = -1;
			let longest = 1;
			sortedRanges(plan).forEach((_, i) => {
				if (length(i) > longest) {
					index = i;
					longest = length(i);
				}
			});
		}
		if (index < 0) return undefined;
		return { index, level: rangeLevels(plan, index).from + Math.floor(length(index) / 2) };
	}

	private renameRotation(from: string, to: string): boolean {
		const name = to.trim();
		if (!name || name == from) return false;
		if (libraryRotation(this.plan(), name)) return false;
		this.change(plan => {
			libraryRotation(plan, from)!.name = name;
			plan.rotations.forEach(range => {
				if (range.name == from) range.name = name;
			});
		});
		this.selected = name;
		this.render();
		return true;
	}

	// Takes a rotation out of the plan. The rotation before it plays on at its levels.
	private deleteRotation(name: string) {
		const preset = libraryRotation(this.plan(), name);
		const spec = this.specPreset(name);
		const lost = !spec || !APLRotation.equals(spec.rotation.rotation!, preset?.rotation ?? APLRotation.create());
		if (lost && !window.confirm(`Delete ${name}? Its changes can't come back.`)) return;
		this.change(plan => {
			plan.rotationLibrary = plan.rotationLibrary.filter(preset => preset.name != name);
			plan.rotations = plan.rotations.filter(range => range.name != name);
		});
		this.selected = this.plan().rotationLibrary[0]?.name ?? '';
		this.render();
		this.syncEditor();
	}

	colorOf(plan: LevelingPlanSettings, name: string): string {
		const index = plan.rotationLibrary.findIndex(preset => preset.name == name);
		return index < 0 ? 'var(--bs-danger)' : ROTATION_COLORS[index % ROTATION_COLORS.length];
	}

	private render() {
		const plan = this.plan();
		if (!this.selectedRotation()) this.selected = sortedRanges(plan)[0]?.name ?? plan.rotationLibrary[0]?.name ?? '';
		if (!this.selectedRotation()) this.selected = plan.rotationLibrary[0]?.name ?? '';
		const ranges = sortedRanges(plan);
		if (ranges[this.selectedRange]?.name != this.selected) {
			this.selectedRange = ranges.findIndex(range => range.name == this.selected);
		}
		this.renderLibrary(plan);
		this.renderNewMenu();
		this.renderEditorHead(plan);
	}

	// The level ranges in order, then the rotations no range plays.
	private renderLibrary(plan: LevelingPlanSettings) {
		const ranges = sortedRanges(plan);
		const rows: Array<Element> = ranges.map((_, i) => this.renderRange(plan, i));
		if (!ranges.length) rows.push(<p className="leveling-plan-note">No level ranges yet, so we play the Rotation tab's at every level.</p>);
		const unplayed = plan.rotationLibrary.filter(preset => !ranges.some(range => range.name == preset.name));
		if (unplayed.length) rows.push(<div className="leveling-plan-library-heading">Not played</div>);
		unplayed.forEach(preset => {
			const spec = this.specPreset(preset.name);
			const edited = !!spec && !APLRotation.equals(spec.rotation.rotation!, preset.rotation ?? APLRotation.create());
			const spells = ref<HTMLDivElement>();
			const row = (
				<button
					className={`leveling-plan-library-item ${preset.name == this.selected && this.selectedRange < 0 ? 'active' : ''}`}
					type="button"
					style={{ '--rotation-color': this.colorOf(plan, preset.name) } as any}>
					<span className="leveling-plan-library-labels">
						<span className="leveling-plan-library-name">
							{preset.name}
							{edited ? <span className="leveling-plan-badge">edited</span> : null}
						</span>
						<span className="leveling-plan-spells" ref={spells} />
					</span>
				</button>
			) as HTMLButtonElement;
			fillSpellIcons(spells.value!, preset.rotation);
			row.addEventListener('click', () => this.select(preset.name));
			rows.push(row);
		});
		this.libraryElem.replaceChildren(...rows);
	}

	// A rotation's card: its name and a remove button on top, its levels and spells under them. A
	// click anywhere else on it opens it in the editor.
	private renderRange(plan: LevelingPlanSettings, i: number): HTMLElement {
		const ranges = sortedRanges(plan);
		const range = ranges[i];
		const preset = libraryRotation(plan, range.name);
		const spec = this.specPreset(range.name);
		const edited = !!preset && !!spec && !APLRotation.equals(spec.rotation.rotation!, preset.rotation ?? APLRotation.create());
		const { from, to } = rangeLevels(plan, i);
		const level = ref<HTMLInputElement>();
		const remove = ref<HTMLButtonElement>();
		const spells = ref<HTMLSpanElement>();
		const row = (
			<div
				className={`leveling-plan-library-item leveling-plan-range ${i == this.selectedRange ? 'active' : ''}`}
				style={{ '--rotation-color': this.colorOf(plan, range.name) } as any}>
				<div className="leveling-plan-range-head">
					<span className={`leveling-plan-library-name ${preset ? '' : 'text-danger'}`}>
						{preset ? range.name : `${range.name || 'No rotation'} (gone)`}
						{edited ? <span className="leveling-plan-badge">edited</span> : null}
					</span>
					{ranges.length > 1 ? (
						<button className="btn btn-link link-danger leveling-plan-remove" ref={remove}>
							<i className="fas fa-times" />
						</button>
					) : null}
				</div>
				<div className="leveling-plan-range-foot">
					<span className="leveling-plan-range-levels">
						{i == 0 ? (
							<span className="leveling-plan-range-from">{String(plan.fromLevel)}</span>
						) : (
							<input className="form-control leveling-plan-level-input" type="number" value={String(range.level)} ref={level} />
						)}
						<span className="leveling-plan-range-to">{to < from ? 'past the plan' : `to ${to}`}</span>
					</span>
					<span className="leveling-plan-spells" ref={spells} />
				</div>
			</div>
		) as HTMLElement;
		fillSpellIcons(spells.value!, preset?.rotation);
		row.addEventListener('click', event => {
			if (!(event.target as HTMLElement).closest('input, button')) this.selectRange(i);
		});
		if (level.value) tippy(level.value, { content: 'The level we start playing it at.' });
		level.value?.addEventListener('change', () => {
			const newLevel = Math.round(Number(level.value!.value));
			if (!(newLevel > plan.fromLevel && newLevel < plan.toLevel)) {
				level.value!.value = String(range.level);
				return;
			}
			this.change(plan => (plan.rotations[i].level = newLevel));
		});
		if (remove.value) {
			tippy(remove.value, { content: 'Take it out of the plan. The rotation before it plays on at its levels.' });
			remove.value.addEventListener('click', () => {
				// Another card can play the same rotation. Then we only drop this card's levels.
				if (ranges.filter(other => other.name == range.name).length > 1) this.change(plan => plan.rotations.splice(i, 1));
				else this.deleteRotation(range.name);
			});
		}
		return row;
	}

	private renderNewMenu() {
		const item = (text: string, tooltip: string | null, onClick: () => void) => {
			const elem = (
				<li>
					<button className="dropdown-item" type="button">
						{text}
					</button>
				</li>
			) as HTMLElement;
			if (tooltip) tippy(elem, { content: tooltip, placement: 'right' });
			elem.querySelector('button')!.addEventListener('click', onClick);
			return elem;
		};
		const groups = new Map<string, ReturnType<LevelingRotationsSection['specPresets']>>();
		this.specPresets().forEach(preset => {
			const group = preset.group ?? 'Presets';
			groups.set(group, [...(groups.get(group) ?? []), preset]);
		});
		const elems: HTMLElement[] = [];
		const selected = this.selectedRotation();
		if (selected) elems.push(item(`Copy of ${selected.name}`, null, () => this.addRotation(selected.name, selected.rotation ?? APLRotation.create())));
		elems.push(
			item("Copy of the Rotation tab's", "The rotation of the settings' character.", () => this.addRotation("Rotation tab's", this.player.aplRotation)),
		);
		groups.forEach((presets, group) => {
			elems.push((<li className="dropdown-header">{group}</li>) as HTMLElement);
			presets.forEach(preset => elems.push(item(preset.name, preset.tooltip ?? null, () => this.addRotation(preset.name, preset.rotation.rotation!))));
		});
		this.newMenuElem.replaceChildren(...elems);
	}

	private renderEditorHead(plan: LevelingPlanSettings) {
		const preset = this.selectedRotation();
		if (!preset) {
			this.editorHeadElem.replaceChildren();
			return;
		}
		const used = this.usedAt(plan, preset.name);
		const spec = this.specPreset(preset.name);
		const edited = !!spec && !APLRotation.equals(spec.rotation.rotation!, preset.rotation ?? APLRotation.create());
		const name = ref<HTMLInputElement>();
		const level = ref<HTMLSelectElement>();
		const imbue = ref<HTMLSelectElement>();
		const revert = ref<HTMLButtonElement>();
		const remove = ref<HTMLButtonElement>();
		const levels: number[] = [];
		for (let l = plan.fromLevel; l < plan.toLevel; l++) levels.push(l);
		const range = sortedRanges(plan)[this.selectedRange];
		const { from: rangeFrom, to: rangeTo } = range ? rangeLevels(plan, this.selectedRange) : { from: 0, to: 0 };
		this.editorHeadElem.replaceChildren(
			<>
				<div className="leveling-plan-editor-title">
					<i className="leveling-plan-swatch" style={{ backgroundColor: this.colorOf(plan, preset.name) }} />
					<input className="form-control leveling-plan-name-input" type="text" value={preset.name} ref={name} />
					<button className="btn btn-outline-primary" ref={revert}>
						<i className="fas fa-rotate-left me-1" />
						Spec's version
					</button>
					<button className="btn btn-outline-danger" ref={remove}>
						<i className="fas fa-trash me-1" />
						Delete
					</button>
				</div>
				<div className="leveling-plan-editor-level">
					<span>Spells and stats of the plan's level</span>
					<select className="form-select leveling-plan-level-select" ref={level}>
						{levels.map(l => (
							<option value={String(l)}>{String(l)}</option>
						))}
					</select>
					<span className="leveling-plan-note">
						{used.length
							? `Played at ${used.length == 1 && used[0].from == used[0].to ? 'level' : 'levels'} ${used
									.map(({ from, to }) => (from == to ? `${from}` : `${from} to ${to}`))
									.join(', ')}.`
							: 'No level range plays it yet.'}
					</span>
				</div>
				{this.hasShamanImbue && range ? (
					<div className="leveling-plan-editor-level">
						<span>{rangeFrom == rangeTo ? `Weapon imbue at level ${rangeFrom}` : `Weapon imbue at levels ${rangeFrom} to ${rangeTo}`}</span>
						<select className="form-select leveling-plan-imbue-select" ref={imbue}>
							{SHAMAN_IMBUES.map(({ name, value }) => (
								<option value={String(value)}>{name}</option>
							))}
						</select>
					</div>
				) : null}
			</>,
		);
		if (imbue.value) {
			const i = this.selectedRange;
			imbue.value.value = String(range!.imbue);
			tippy(imbue.value, { content: "Settings tab's imbue keeps the one in the Settings tab." });
			imbue.value.addEventListener('change', () => this.change(plan => (plan.rotations[i].imbue = Number(imbue.value!.value) as WeaponImbue)));
		}
		level.value!.value = String(this.characterLevel(plan));
		level.value!.addEventListener('change', () => {
			this.editLevel = Number(level.value!.value);
			this.syncEditor();
		});
		tippy(level.value!, { content: "The editor lists the spells the plan's character has at this level, with its talents and gear there." });
		name.value!.addEventListener('change', () => {
			if (!this.renameRotation(preset.name, name.value!.value)) name.value!.value = preset.name;
		});
		revert.value!.classList.toggle('hide', !edited);
		tippy(revert.value!, { content: `Undo our changes, back to the spec's ${preset.name}.` });
		revert.value!.addEventListener('click', () =>
			this.change(plan => (libraryRotation(plan, preset.name)!.rotation = APLRotation.clone(spec!.rotation.rotation!))),
		);
		const onlyOne = sortedRanges(plan).every(range => range.name == preset.name) && used.length > 0;
		remove.value!.disabled = onlyOne;
		tippy(wrapForTooltip(remove.value!), {
			content: onlyOne ? 'The plan needs a rotation.' : 'Take it out of the plan. The rotation before it plays on at its levels.',
		});
		remove.value!.addEventListener('click', () => this.deleteRotation(preset.name));
	}
}

// A disabled button gets no mouse events, so its tooltip goes on a wrapper around it.
function wrapForTooltip(button: HTMLElement): HTMLElement {
	const wrapper = (<span className="leveling-plan-tooltip-wrap" />) as HTMLElement;
	button.replaceWith(wrapper);
	wrapper.appendChild(button);
	return wrapper;
}

// The spells a rotation casts, in its order, without the same one twice.
function rotationSpells(rotation: APLRotation | undefined): ActionId[] {
	const ids: ActionId[] = [];
	const seen = new Set<string>();
	const visit = (action: APLAction | undefined) => {
		if (!action) return;
		const kind = action.action;
		let spellId = undefined;
		if (kind.oneofKind == 'castSpell') spellId = kind.castSpell.spellId;
		else if (kind.oneofKind == 'channelSpell') spellId = kind.channelSpell.spellId;
		else if (kind.oneofKind == 'multidot') spellId = kind.multidot.spellId;
		else if (kind.oneofKind == 'sequence') kind.sequence.actions.forEach(visit);
		else if (kind.oneofKind == 'strictSequence') kind.strictSequence.actions.forEach(visit);
		if (!spellId) return;
		const id = ActionId.fromProto(spellId);
		if (!id.spellId && !id.itemId && !id.otherId) return;
		const key = id.toString();
		if (!seen.has(key)) {
			seen.add(key);
			ids.push(id);
		}
	};
	rotation?.priorityList.filter(item => !item.hide).forEach(item => visit(item.action));
	return ids;
}

function fillSpellIcons(elem: HTMLElement, rotation: APLRotation | undefined) {
	const ids = rotationSpells(rotation);
	ids.slice(0, LIST_SPELLS).forEach(id => {
		const icon = (<span className="leveling-plan-spell" />) as HTMLElement;
		elem.appendChild(icon);
		id.fill().then(filled => {
			icon.style.backgroundImage = `url('${filled.iconUrl}')`;
			tippy(icon, { content: filled.name });
		});
	});
	if (ids.length > LIST_SPELLS) elem.appendChild(<span className="leveling-plan-spells-more">{`+${ids.length - LIST_SPELLS}`}</span>);
}
