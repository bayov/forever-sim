import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import { Player } from '../../player';
import { APLRotation } from '../../proto/apl';
import { Spec, WeaponImbue } from '../../proto/common';
import { LevelingPlanSettings, LevelingRotation } from '../../proto/ui';
import { TypedEvent } from '../../typed_event';
import { ContentBlock } from '../content_block';
import { APLRotationPicker } from './apl_rotation_picker';

// The shaman imbues a rotation can come with, by the names the class options give them.
const SHAMAN_IMBUES: Array<{ name: string; value: WeaponImbue }> = [
	{ name: 'Class options', value: WeaponImbue.WeaponImbueUnknown },
	{ name: 'Rockbiter Weapon', value: WeaponImbue.RockbiterWeapon },
	{ name: 'Flametongue Weapon', value: WeaponImbue.FlametongueWeapon },
	{ name: 'Frostbrand Weapon', value: WeaponImbue.FrostbrandWeapon },
	{ name: 'Windfury Weapon', value: WeaponImbue.WindfuryWeapon },
];

// The value of the select for a rotation copied from the Rotation tab.
const COPY = '__copy__';

// The Rotation part of the Leveling Plan tab: which rotation we play at which levels, and the
// selected one in the rotation editor.
//
// Each rotation is played from its level until the next one's. It starts as one of the spec's
// rotation presets, or a copy of the Rotation tab's rotation, and we can change it in the editor
// under the list. A shaman's rotation also says which weapon imbue we use, because Windfury
// Weapon only comes at level 30.
//
// The editor is the Rotation tab's, and it edits the character's own rotation. Its spell lists
// come from the character's level and talents too. So while this part is open, the character
// is the plan's at the first level of the selected rotation (loadLevel), and each change in the
// editor goes back into the plan.
export class LevelingRotationsSection {
	readonly rootElem: HTMLElement;
	private readonly simUI: IndividualSimUI<Spec>;
	private readonly player: Player<Spec>;
	private readonly listElem: HTMLElement;
	private readonly editorTitleElem: HTMLElement;
	private readonly hasShamanImbue: boolean;
	private readonly loadLevel: (level: number) => void;
	private readonly rotationTabRotation: () => APLRotation;
	private selected = 0;
	private active = false;

	// rotationTabRotation is the Rotation tab's rotation, the one the character had before this
	// part made it the plan's.
	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>, loadLevel: (level: number) => void, rotationTabRotation: () => APLRotation) {
		this.simUI = simUI;
		this.player = simUI.player;
		this.loadLevel = loadLevel;
		this.rotationTabRotation = rotationTabRotation;
		this.hasShamanImbue = 'shamanImbue' in (this.player.getSpecOptions() as object);
		this.rootElem = (<div className="leveling-plan-rotations" />) as HTMLElement;
		parentElem.appendChild(this.rootElem);

		const block = new ContentBlock(this.rootElem, 'leveling-plan-rotations-block', {
			header: {
				title: 'Rotations',
				tooltip:
					"Each rotation is played from its level until the next one's. Pick one of the presets, or copy the Rotation tab's rotation, and change it in the editor below. A click on a row opens it in the editor.",
			},
		});
		const list = ref<HTMLDivElement>();
		const add = ref<HTMLButtonElement>();
		block.bodyElement.appendChild(
			<>
				<div className="leveling-plan-rotation-list" ref={list} />
				<div className="leveling-plan-buttons">
					<button className="btn btn-outline-primary" ref={add}>
						<i className="fas fa-plus me-1" />
						Add rotation
					</button>
				</div>
			</>,
		);
		this.listElem = list.value!;
		tippy(add.value!, { content: 'Play another rotation from 5 levels after the last one.' });
		add.value!.addEventListener('click', () => this.addRotation());

		const editorBlock = new ContentBlock(this.rootElem, 'leveling-plan-rotation-editor-block', {
			header: {
				title: 'Rotation',
				tooltip:
					"The selected rotation. While it's open, the character is the plan's at the rotation's first level, so the spells are the ones it has there.",
			},
		});
		const editorTitle = ref<HTMLParagraphElement>();
		const editor = ref<HTMLDivElement>();
		editorBlock.bodyElement.appendChild(
			<>
				<p className="leveling-plan-note" ref={editorTitle} />
				<div className="leveling-plan-rotation-editor" ref={editor} />
			</>,
		);
		this.editorTitleElem = editorTitle.value!;
		new APLRotationPicker(editor.value!, simUI, this.player, simUI.individualConfig.prepullActionNote, false);

		// A change in the editor goes into the selected rotation of the plan.
		this.player.rotationChangeEmitter.on(() => {
			const rotation = this.selectedRotation();
			if (!this.active || !rotation || APLRotation.equals(rotation.rotation ?? APLRotation.create(), this.player.aplRotation)) return;
			this.change(rotations => (rotations[this.selected].rotation = APLRotation.clone(this.player.aplRotation)));
		});
		// A rotation we picked in the list goes into the editor.
		this.player.levelingPlanChangeEmitter.on(() => {
			this.render();
			const rotation = this.selectedRotation();
			if (this.active && rotation?.rotation && !APLRotation.equals(rotation.rotation, this.player.aplRotation)) {
				this.player.setAplRotation(TypedEvent.nextEventID(), rotation.rotation);
			}
		});
		this.render();
	}

	// The tab calls it when this part opens or closes.
	setActive(active: boolean) {
		this.active = active;
		if (active) this.load();
	}

	private plan(): LevelingPlanSettings {
		return this.player.getLevelingPlan();
	}

	private sorted(plan: LevelingPlanSettings): LevelingRotation[] {
		return [...plan.rotations].sort((a, b) => a.level - b.level);
	}

	private selectedRotation(): LevelingRotation | undefined {
		return this.sorted(this.plan())[this.selected];
	}

	// The first level the rotation is played at.
	private startLevel(index: number): number {
		const plan = this.plan();
		return index == 0 ? plan.fromLevel : this.sorted(plan)[index]?.level ?? plan.fromLevel;
	}

	private load() {
		if (!this.selectedRotation()) return;
		this.loadLevel(Math.min(this.startLevel(this.selected), this.plan().toLevel - 1));
	}

	private select(index: number) {
		this.selected = index;
		this.render();
		if (this.active) this.load();
	}

	private presets() {
		return this.simUI.individualConfig.presets.rotations;
	}

	private change(change: (rotations: LevelingRotation[]) => void) {
		const plan = this.plan();
		plan.rotations = this.sorted(plan);
		change(plan.rotations);
		this.player.setLevelingPlan(TypedEvent.nextEventID(), plan);
	}

	private addRotation() {
		const plan = this.plan();
		const rotations = this.sorted(plan);
		const last = rotations[rotations.length - 1];
		const level = last ? Math.min(last.level + 5, plan.toLevel - 1) : plan.fromLevel;
		this.change(rotations =>
			rotations.push(
				last
					? LevelingRotation.clone({ ...last, level })
					: LevelingRotation.create({ level, name: '', rotation: APLRotation.clone(this.rotationTabRotation()) }),
			),
		);
		this.select(rotations.length);
	}

	// The levels a rotation is played at, like "Levels 10 to 14".
	private levelsText(index: number, rotations: LevelingRotation[]): string {
		const plan = this.plan();
		const from = this.startLevel(index);
		const to = index + 1 < rotations.length ? rotations[index + 1].level - 1 : plan.toLevel - 1;
		return to < from ? `From level ${from}, past the plan` : `Levels ${from} to ${to}`;
	}

	private render() {
		const plan = this.plan();
		const rotations = this.sorted(plan);
		if (this.selected >= rotations.length) this.selected = Math.max(rotations.length - 1, 0);
		const rows = rotations.map((rotation, i) => {
			const to = i + 1 < rotations.length ? rotations[i + 1].level - 1 : plan.toLevel - 1;
			const level = ref<HTMLInputElement>();
			const select = ref<HTMLSelectElement>();
			const imbue = ref<HTMLSelectElement>();
			const remove = ref<HTMLButtonElement>();
			const changed = ref<HTMLSpanElement>();
			const presetNames = this.presets().map(preset => preset.name);
			const preset = this.presets().find(preset => preset.name == rotation.name);
			const row = (
				<div className={`leveling-plan-rotation ${i == this.selected ? 'active' : ''}`}>
					<span className="leveling-plan-rotation-levels">
						{i == 0 ? (
							`From ${plan.fromLevel}`
						) : (
							<>
								From <input className="form-control leveling-plan-level-input" type="number" value={String(rotation.level)} ref={level} />
							</>
						)}
						{to < rotation.level ? ', past the plan' : ` to ${to}`}
					</span>
					<select className="form-select leveling-plan-rotation-select" ref={select}>
						{presetNames.map(name => (
							<option value={name}>{name}</option>
						))}
						<option value="">Copy of the Rotation tab</option>
						<option value={COPY}>Copy the Rotation tab again</option>
					</select>
					<span className="leveling-plan-rotation-changed" ref={changed}>
						changed
					</span>
					{this.hasShamanImbue ? (
						<select className="form-select leveling-plan-imbue-select" ref={imbue}>
							{SHAMAN_IMBUES.map(({ name, value }) => (
								<option value={String(value)}>{name}</option>
							))}
						</select>
					) : null}
					{i == 0 ? null : (
						<button className="btn btn-link link-danger leveling-plan-remove" ref={remove}>
							<i className="fas fa-times" />
						</button>
					)}
				</div>
			) as HTMLElement;
			row.addEventListener('click', event => {
				if ((event.target as HTMLElement).closest('input, select, button')) return;
				this.select(i);
			});

			const isChanged = !!preset?.rotation.rotation && !APLRotation.equals(preset.rotation.rotation, rotation.rotation ?? APLRotation.create());
			changed.value!.classList.toggle('hide', !isChanged);
			tippy(changed.value!, { content: 'Changed in the editor since we picked the preset. Pick the preset again to undo.' });
			select.value!.value = preset ? rotation.name : '';
			select.value!.addEventListener('change', () => {
				const value = select.value!.value;
				this.selected = i;
				this.change(rotations => {
					if (value == COPY || value == '') {
						rotations[i].name = '';
						rotations[i].rotation = APLRotation.clone(this.rotationTabRotation());
						return;
					}
					const preset = this.presets().find(preset => preset.name == value);
					if (!preset?.rotation.rotation) return;
					rotations[i].name = preset.name;
					rotations[i].rotation = APLRotation.clone(preset.rotation.rotation);
				});
			});
			if (imbue.value) {
				imbue.value.value = String(rotation.imbue);
				tippy(imbue.value, { content: 'The weapon imbue at these levels. Class options keeps the one in the Settings tab.' });
				imbue.value.addEventListener('change', () => this.change(rotations => (rotations[i].imbue = Number(imbue.value!.value) as WeaponImbue)));
			}
			level.value?.addEventListener('change', () => {
				const newLevel = Math.round(Number(level.value!.value));
				if (!(newLevel > plan.fromLevel && newLevel < plan.toLevel)) {
					level.value!.value = String(rotation.level);
					return;
				}
				this.change(rotations => (rotations[i].level = newLevel));
			});
			if (remove.value) {
				tippy(remove.value, { content: 'Drop this rotation' });
				remove.value.addEventListener('click', () => {
					if (this.selected >= i) this.selected = Math.max(this.selected - 1, 0);
					this.change(rotations => rotations.splice(i, 1));
				});
			}
			return row;
		});
		this.listElem.replaceChildren(
			...(rows.length ? rows : [<p className="leveling-plan-note">No rotations yet, so we play the Rotation tab's at every level.</p>]),
		);

		const rotation = rotations[this.selected];
		this.editorTitleElem.textContent = rotation
			? `${this.levelsText(this.selected, rotations)}: ${rotation.name || 'Copy of the Rotation tab'}. The character is level ${Math.min(
					this.startLevel(this.selected),
					plan.toLevel - 1,
			  )} of the plan while this is open.`
			: 'Add a rotation to edit it here.';
	}
}
