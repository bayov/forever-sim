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

// The shaman imbues a level range can come with, short so they fit under the range.
const SHAMAN_IMBUES: Array<{ name: string; value: WeaponImbue }> = [
	{ name: 'Settings imbue', value: WeaponImbue.WeaponImbueUnknown },
	{ name: 'Rockbiter', value: WeaponImbue.RockbiterWeapon },
	{ name: 'Flametongue', value: WeaponImbue.FlametongueWeapon },
	{ name: 'Frostbrand', value: WeaponImbue.FrostbrandWeapon },
	{ name: 'Windfury', value: WeaponImbue.WindfuryWeapon },
];

// The colors of the plan's rotations, by their place in the library, so a level range and its
// rotation have the same color in the timeline and in the list.
const ROTATION_COLORS = ['#4e9bff', '#f0a030', '#3cc47c', '#c86bff', '#ff6b6b', '#2ec4c4', '#d8c64a', '#ff8fc8'];

// How many of a rotation's spells the list shows.
const LIST_SPELLS = 6;

// The Rotations part of the Leveling Plan tab.
//
// The plan has its own rotations. They start as copies of the spec's rotation presets, and we
// can add more, rename them and change them in the editor on the right. The list on the left is
// the plan's level ranges in order, each with the rotation we play from its level until the next
// range's. A shaman's range also says which weapon imbue we use, because Windfury Weapon only
// comes at level 30. The rotations no range plays come last.
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
					"Each level range plays one of the plan's rotations, until the next range's level. A click opens its rotation in the editor. Changing them here leaves the spec's presets and the Rotation tab alone.",
			},
		});
		const library = ref<HTMLDivElement>();
		const addRange = ref<HTMLButtonElement>();
		const newMenu = ref<HTMLUListElement>();
		libraryBlock.bodyElement.appendChild(
			<>
				<div className="leveling-plan-library" ref={library} />
				<div className="leveling-plan-buttons">
					<button className="btn btn-outline-primary" ref={addRange}>
						<i className="fas fa-plus me-1" />
						Level range
					</button>
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
		tippy(addRange.value!, { content: 'Split the longest range in two, so its second half can play another rotation.' });
		addRange.value!.addEventListener('click', () => this.addRange());

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
		if (name == this.selected) return;
		this.selected = name;
		this.editLevel = 0;
		this.render();
		this.syncEditor();
	}

	private change(change: (plan: LevelingPlanSettings) => void) {
		const plan = this.player.getLevelingPlan();
		plan.rotations = sortedRanges(plan);
		change(plan);
		this.player.setLevelingPlan(TypedEvent.nextEventID(), plan);
	}

	// Adds a rotation to the library and opens it.
	private addRotation(name: string, rotation: APLRotation) {
		const plan = this.plan();
		const free = freeRotationName(plan, name);
		this.change(plan => plan.rotationLibrary.push(LevelingRotationPreset.create({ name: free, rotation: APLRotation.clone(rotation) })));
		this.select(free);
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

	private deleteRotation(name: string) {
		this.change(plan => (plan.rotationLibrary = plan.rotationLibrary.filter(preset => preset.name != name)));
		this.selected = this.plan().rotationLibrary[0]?.name ?? '';
		this.render();
		this.syncEditor();
	}

	// Splits the longest range in two. The new range plays the same rotation, so nothing
	// changes until we pick another one for it.
	private addRange() {
		const plan = this.plan();
		const ranges = sortedRanges(plan);
		if (!ranges.length) {
			this.change(plan => plan.rotations.push(LevelingRotation.create({ level: plan.fromLevel, name: plan.rotationLibrary[0]?.name ?? '' })));
			return;
		}
		let longest = -1;
		let longestLevels = 0;
		ranges.forEach((_, i) => {
			const { from, to } = rangeLevels(plan, i);
			if (to - from + 1 > longestLevels) {
				longest = i;
				longestLevels = to - from + 1;
			}
		});
		if (longest < 0 || longestLevels < 2) return;
		const { from } = rangeLevels(plan, longest);
		const level = from + Math.floor(longestLevels / 2);
		this.change(plan => plan.rotations.push(LevelingRotation.clone({ ...ranges[longest], level })));
	}

	colorOf(plan: LevelingPlanSettings, name: string): string {
		const index = plan.rotationLibrary.findIndex(preset => preset.name == name);
		return index < 0 ? 'var(--bs-danger)' : ROTATION_COLORS[index % ROTATION_COLORS.length];
	}

	private render() {
		const plan = this.plan();
		if (!this.selectedRotation()) this.selected = sortedRanges(plan)[0]?.name ?? plan.rotationLibrary[0]?.name ?? '';
		if (!this.selectedRotation()) this.selected = plan.rotationLibrary[0]?.name ?? '';
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
				<button className={`leveling-plan-library-item ${preset.name == this.selected ? 'active' : ''}`} type="button">
					<i className="leveling-plan-swatch" style={{ backgroundColor: this.colorOf(plan, preset.name) }} />
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

	// A level range: its levels, its rotation and a remove button on top, the rotation's spells
	// and a shaman's imbue under them. A click anywhere else on it opens its rotation.
	private renderRange(plan: LevelingPlanSettings, i: number): HTMLElement {
		const range = sortedRanges(plan)[i];
		const names = plan.rotationLibrary.map(preset => preset.name);
		const { from, to } = rangeLevels(plan, i);
		const level = ref<HTMLInputElement>();
		const select = ref<HTMLSelectElement>();
		const imbue = ref<HTMLSelectElement>();
		const remove = ref<HTMLButtonElement>();
		const spells = ref<HTMLSpanElement>();
		const missing = !names.includes(range.name);
		const row = (
			<div className={`leveling-plan-library-item leveling-plan-range ${range.name == this.selected ? 'active' : ''}`}>
				<div className="leveling-plan-range-head">
					<i className="leveling-plan-swatch" style={{ backgroundColor: this.colorOf(plan, range.name) }} />
					<span className="leveling-plan-range-levels">
						{i == 0 ? (
							<span className="leveling-plan-range-from">{String(plan.fromLevel)}</span>
						) : (
							<input className="form-control leveling-plan-level-input" type="number" value={String(range.level)} ref={level} />
						)}
						<span className="leveling-plan-range-to">{to < from ? 'past the plan' : `to ${to}`}</span>
					</span>
					<select className={`form-select form-select-sm leveling-plan-rotation-select ${missing ? 'is-invalid' : ''}`} ref={select}>
						{missing ? <option value={range.name}>{range.name ? `${range.name} (gone)` : 'No rotation'}</option> : null}
						{names.map(name => (
							<option value={name}>{name}</option>
						))}
					</select>
					{i == 0 ? (
						<span className="leveling-plan-remove-space" />
					) : (
						<button className="btn btn-link link-danger leveling-plan-remove" ref={remove}>
							<i className="fas fa-times" />
						</button>
					)}
				</div>
				<div className="leveling-plan-range-foot">
					<span className="leveling-plan-spells" ref={spells} />
					{this.hasShamanImbue ? (
						<select className="form-select form-select-sm leveling-plan-imbue-select" ref={imbue}>
							{SHAMAN_IMBUES.map(({ name, value }) => (
								<option value={String(value)}>{name}</option>
							))}
						</select>
					) : null}
				</div>
			</div>
		) as HTMLElement;
		fillSpellIcons(spells.value!, libraryRotation(plan, range.name)?.rotation);
		row.addEventListener('click', event => {
			if (!(event.target as HTMLElement).closest('input, select, button')) this.select(range.name);
		});
		select.value!.value = range.name;
		select.value!.addEventListener('change', () => {
			const name = select.value!.value;
			this.change(plan => (plan.rotations[i].name = name));
			this.select(name);
		});
		if (imbue.value) {
			imbue.value.value = String(range.imbue);
			tippy(imbue.value, { content: 'The weapon imbue at these levels. Settings imbue keeps the one in the Settings tab.' });
			imbue.value.addEventListener('change', () => this.change(plan => (plan.rotations[i].imbue = Number(imbue.value!.value) as WeaponImbue)));
		}
		level.value?.addEventListener('change', () => {
			const newLevel = Math.round(Number(level.value!.value));
			if (!(newLevel > plan.fromLevel && newLevel < plan.toLevel)) {
				level.value!.value = String(range.level);
				return;
			}
			this.change(plan => (plan.rotations[i].level = newLevel));
		});
		if (remove.value) {
			tippy(remove.value, { content: 'Drop this range, so the range before it plays on' });
			remove.value.addEventListener('click', () => this.change(plan => plan.rotations.splice(i, 1)));
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
		const revert = ref<HTMLButtonElement>();
		const remove = ref<HTMLButtonElement>();
		const levels: number[] = [];
		for (let l = plan.fromLevel; l < plan.toLevel; l++) levels.push(l);
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
							? `Played at levels ${used.map(({ from, to }) => (from == to ? `${from}` : `${from} to ${to}`)).join(', ')}.`
							: 'No level range plays it yet.'}
					</span>
				</div>
			</>,
		);
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
		remove.value!.disabled = used.length > 0;
		tippy(wrapForTooltip(remove.value!), {
			content: used.length ? 'A level range plays it. Give the range another rotation first.' : 'Take it out of the plan.',
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
