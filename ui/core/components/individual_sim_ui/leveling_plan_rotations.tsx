import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import { Player } from '../../player';
import { Spec, WeaponImbue } from '../../proto/common';
import { APLRotation } from '../../proto/apl';
import { LevelingPlanSettings, LevelingRotation } from '../../proto/ui';
import { TypedEvent } from '../../typed_event';
import { ContentBlock } from '../content_block';

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

// The Rotation part of the Leveling Plan tab: which rotation we play at which levels.
//
// Each rotation is played from its level until the next one's. It's one of the spec's rotation
// presets, or a copy of the Rotation tab's rotation. A shaman's rotation also says which
// weapon imbue we use, because Windfury Weapon only comes at level 30.
export class LevelingRotationsSection {
	readonly rootElem: HTMLElement;
	private readonly simUI: IndividualSimUI<Spec>;
	private readonly player: Player<Spec>;
	private readonly listElem: HTMLElement;
	private readonly hasShamanImbue: boolean;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		this.simUI = simUI;
		this.player = simUI.player;
		this.hasShamanImbue = 'shamanImbue' in (this.player.getSpecOptions() as object);
		this.rootElem = (<div className="leveling-plan-rotations" />) as HTMLElement;
		parentElem.appendChild(this.rootElem);

		const block = new ContentBlock(this.rootElem, 'leveling-plan-rotations-block', {
			header: {
				title: 'Rotations',
				tooltip:
					"Each rotation is played from its level until the next one's. Pick one of the presets, or copy the Rotation tab's rotation as it is now.",
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

		this.player.levelingPlanChangeEmitter.on(() => this.render());
		this.render();
	}

	private plan(): LevelingPlanSettings {
		return this.player.getLevelingPlan();
	}

	private sorted(plan: LevelingPlanSettings): LevelingRotation[] {
		return [...plan.rotations].sort((a, b) => a.level - b.level);
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
		this.change(rotations => {
			const plan = this.plan();
			const last = rotations[rotations.length - 1];
			const level = last ? Math.min(last.level + 5, plan.toLevel - 1) : plan.fromLevel;
			rotations.push(
				last
					? LevelingRotation.clone({ ...last, level })
					: LevelingRotation.create({ level, name: '', rotation: APLRotation.clone(this.player.aplRotation) }),
			);
		});
	}

	private render() {
		const plan = this.plan();
		const rotations = this.sorted(plan);
		const rows = rotations.map((rotation, i) => {
			const to = i + 1 < rotations.length ? rotations[i + 1].level - 1 : plan.toLevel - 1;
			const level = ref<HTMLInputElement>();
			const select = ref<HTMLSelectElement>();
			const imbue = ref<HTMLSelectElement>();
			const remove = ref<HTMLButtonElement>();
			const presetNames = this.presets().map(preset => preset.name);
			const row = (
				<div className="leveling-plan-rotation">
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

			select.value!.value = rotation.name && presetNames.includes(rotation.name) ? rotation.name : '';
			select.value!.addEventListener('change', () => {
				const value = select.value!.value;
				this.change(rotations => {
					if (value == COPY || value == '') {
						rotations[i].name = '';
						rotations[i].rotation = APLRotation.clone(this.player.aplRotation);
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
				remove.value.addEventListener('click', () => this.change(rotations => rotations.splice(i, 1)));
			}
			return row;
		});
		this.listElem.replaceChildren(
			...(rows.length ? rows : [<p className="leveling-plan-note">No rotations yet, so we play the Rotation tab's at every level.</p>]),
		);
	}
}
