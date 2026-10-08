import tippy from 'tippy.js';

import { MAX_CHARACTER_LEVEL } from '../../constants/mechanics';
import { Player } from '../../player';
import { Profession, Race, Spec } from '../../proto/common';
import { professionNames, raceNames } from '../../proto_utils/names';
import { specToEligibleRaces } from '../../proto_utils/utils';
import { TypedEvent } from '../../typed_event';
import { Input } from '../input';

const iconUrl = (icon: string) => `https://wow.zamimg.com/images/wow/icons/large/${icon}.jpg`;

// The level as a slider from 1 to 60, with a button on each side to step it by one and a
// field to type it in.
//
// A level of 0 means max level, so we show it as 60. The slider moves the field while we drag
// it, but we only set the level when we let go, because each change of level reloads the
// gear picker and runs a new sim.
export class LevelPicker extends Input<Player<Spec>, number> {
	private readonly slider: HTMLInputElement;
	private readonly field: HTMLInputElement;
	private readonly down: HTMLButtonElement;
	private readonly up: HTMLButtonElement;

	constructor(parent: HTMLElement, player: Player<Spec>) {
		super(parent, 'level-picker-root', player, {
			id: 'level',
			label: 'Level',
			labelTooltip:
				'Character level. Base stats, ability ranks, the crit each point of Agility gives, the talent points and the items offered in the gear picker follow it.',
			changedEvent: player => player.miscOptionsChangeEmitter,
			getValue: player => player.getLevel(),
			setValue: (eventID, player, newValue) => player.setLevel(eventID, newValue),
		});

		this.down = (
			<button type="button" className="level-picker-step" aria-label="One level down">
				<i className="fas fa-minus"></i>
			</button>
		) as HTMLButtonElement;
		this.slider = (<input type="range" className="level-picker-slider" min="1" max={String(MAX_CHARACTER_LEVEL)} step="1" />) as HTMLInputElement;
		this.up = (
			<button type="button" className="level-picker-step" aria-label="One level up">
				<i className="fas fa-plus"></i>
			</button>
		) as HTMLButtonElement;
		this.field = (<input type="text" id="level" inputMode="numeric" className="form-control level-picker-field" />) as HTMLInputElement;

		this.rootElem.appendChild(
			<div className="level-picker-controls">
				{this.down}
				{this.slider}
				{this.up}
				{this.field}
			</div>,
		);

		this.slider.addEventListener('input', () => this.show(this.slider.valueAsNumber), { signal: this.signal });
		this.slider.addEventListener('change', () => this.commit(this.slider.valueAsNumber), { signal: this.signal });
		this.field.addEventListener('change', () => this.commit(parseInt(this.field.value)), { signal: this.signal });
		this.down.addEventListener('click', () => this.commit(this.getInputValue() - 1), { signal: this.signal });
		this.up.addEventListener('click', () => this.commit(this.getInputValue() + 1), { signal: this.signal });

		this.init();
	}

	getInputElem(): HTMLElement {
		return this.slider;
	}

	getInputValue(): number {
		return this.slider.valueAsNumber;
	}

	setInputValue(newValue: number) {
		this.show(newValue > 0 && newValue < MAX_CHARACTER_LEVEL ? newValue : MAX_CHARACTER_LEVEL);
	}

	// A typed level outside 1 to 60 goes to the nearest end, and anything that isn't a number
	// puts back the level we had.
	private commit(level: number) {
		if (isNaN(level)) {
			this.setInputValue(this.getSourceValue());
			return;
		}
		this.show(Math.min(MAX_CHARACTER_LEVEL, Math.max(1, Math.round(level))));
		this.inputChanged(TypedEvent.nextEventID());
	}

	private show(level: number) {
		this.slider.value = String(level);
		this.field.value = String(level);
		this.down.disabled = level <= 1;
		this.up.disabled = level >= MAX_CHARACTER_LEVEL;
		// The part of the track left of the thumb is filled, see .level-picker-slider.
		this.slider.style.setProperty('--fill', `${((level - 1) / (MAX_CHARACTER_LEVEL - 1)) * 100}%`);
	}
}

// The Skyborne are new in Forever and wowhead has no race icon for them, so they get the
// icon of their Walk on Air racial.
const raceIcons: Record<Race, string> = {
	[Race.RaceUnknown]: 'inv_misc_questionmark',
	[Race.RaceDwarf]: 'race_dwarf_male',
	[Race.RaceGnome]: 'race_gnome_male',
	[Race.RaceHuman]: 'race_human_male',
	[Race.RaceNightElf]: 'race_nightelf_male',
	[Race.RaceOrc]: 'race_orc_male',
	[Race.RaceTauren]: 'race_tauren_male',
	[Race.RaceTroll]: 'race_troll_male',
	[Race.RaceUndead]: 'race_scourge_male',
	[Race.RaceSkyborneHighOrder]: 'inv_elemental_primal_air',
	[Race.RaceSkyborneWindshaper]: 'inv_elemental_primal_air',
};

// The races our class can be, as icons. The selected one is lit and the others are dimmed,
// and the label says which one it is.
export class RacePicker extends Input<Player<Spec>, Race> {
	private readonly value: HTMLElement;
	private readonly options = new Map<Race, HTMLButtonElement>();
	private race = Race.RaceUnknown;

	constructor(parent: HTMLElement, player: Player<Spec>) {
		super(parent, 'race-picker-root', player, {
			label: 'Race',
			changedEvent: player => player.raceChangeEmitter,
			getValue: player => player.getRace(),
			setValue: (eventID, player, newValue) => player.setRace(eventID, newValue),
		});
		this.rootElem.classList.add('player-icon-picker');

		this.value = (<span className="player-icon-picker-value"></span>) as HTMLElement;
		const options = (<div className="player-icon-picker-options"></div>) as HTMLElement;
		this.rootElem.append(this.value, options);

		specToEligibleRaces[player.spec].forEach(race => {
			const option = iconOption(raceIcons[race], raceNames.get(race)!);
			option.addEventListener(
				'click',
				() => {
					if (race === this.race) return;
					this.setInputValue(race);
					this.inputChanged(TypedEvent.nextEventID());
				},
				{ signal: this.signal },
			);
			const tooltip = tippy(option, { content: raceNames.get(race) });
			this.addOnDisposeCallback(() => tooltip.destroy());
			this.options.set(race, option);
			options.appendChild(option);
		});

		this.init();
	}

	getInputElem(): HTMLElement {
		return this.rootElem;
	}

	getInputValue(): Race {
		return this.race;
	}

	setInputValue(newValue: Race) {
		this.race = newValue;
		this.value.textContent = raceNames.get(newValue) ?? '';
		this.options.forEach((option, race) => option.classList.toggle('active', race === newValue));
	}

	protected formatPresetValue(value: Race): string {
		return raceNames.get(value) ?? '';
	}
}

const professionIcons: Record<Profession, string> = {
	[Profession.ProfessionUnknown]: 'inv_misc_questionmark',
	[Profession.Alchemy]: 'trade_alchemy',
	[Profession.Blacksmithing]: 'trade_blacksmithing',
	[Profession.Enchanting]: 'trade_engraving',
	[Profession.Engineering]: 'trade_engineering',
	[Profession.Herbalism]: 'trade_herbalism',
	[Profession.Leatherworking]: 'trade_leatherworking',
	[Profession.Mining]: 'trade_mining',
	[Profession.Skinning]: 'inv_misc_pelt_wolf_01',
	[Profession.Tailoring]: 'trade_tailoring',
};

const MAX_PROFESSIONS = 2;

// The professions as icons, where we pick up to two.
//
// When we already have two, the others are dimmed and a click on them does nothing. To switch
// one, we first click one of our two to drop it. The value is sorted, so picking the same two
// in another order doesn't count as a change from the preset.
export class ProfessionsPicker extends Input<Player<Spec>, Array<Profession>> {
	private readonly value: HTMLElement;
	private readonly options = new Map<Profession, HTMLButtonElement>();
	private professions: Array<Profession> = [];

	constructor(parent: HTMLElement, player: Player<Spec>) {
		super(parent, 'professions-picker-root', player, {
			label: 'Professions',
			changedEvent: player => player.professionChangeEmitter,
			getValue: player => [...player.getProfessions()].sort((a, b) => a - b),
			setValue: (eventID, player, newValue) => player.setProfessions(eventID, newValue),
		});
		this.rootElem.classList.add('player-icon-picker');

		this.value = (<span className="player-icon-picker-value"></span>) as HTMLElement;
		const options = (<div className="player-icon-picker-options"></div>) as HTMLElement;
		this.rootElem.append(this.value, options);

		Object.values(Profession)
			.filter((profession): profession is Profession => typeof profession === 'number' && profession !== Profession.ProfessionUnknown)
			.forEach(profession => {
				const option = iconOption(professionIcons[profession], professionNames.get(profession)!);
				option.addEventListener('click', () => this.toggle(profession), { signal: this.signal });
				const tooltip = tippy(option, {
					content: professionNames.get(profession),
					onShow: instance => {
						const full = this.professions.length >= MAX_PROFESSIONS && !this.professions.includes(profession);
						instance.setContent(
							full ? `${professionNames.get(profession)} (drop one of your two professions first)` : professionNames.get(profession)!,
						);
					},
				});
				this.addOnDisposeCallback(() => tooltip.destroy());
				this.options.set(profession, option);
				options.appendChild(option);
			});

		this.init();
	}

	getInputElem(): HTMLElement {
		return this.rootElem;
	}

	getInputValue(): Array<Profession> {
		return [...this.professions];
	}

	setInputValue(newValue: Array<Profession>) {
		this.professions = [...newValue].sort((a, b) => a - b);
		this.value.textContent = this.professions.map(profession => professionNames.get(profession)).join(', ') || 'None';
		const full = this.professions.length >= MAX_PROFESSIONS;
		this.options.forEach((option, profession) => {
			const active = this.professions.includes(profession);
			option.classList.toggle('active', active);
			option.classList.toggle('locked', full && !active);
		});
	}

	protected formatPresetValue(value: Array<Profession>): string {
		return value.map(profession => professionNames.get(profession)).join(', ') || 'none';
	}

	private toggle(profession: Profession) {
		if (this.professions.includes(profession)) {
			this.setInputValue(this.professions.filter(p => p !== profession));
		} else if (this.professions.length < MAX_PROFESSIONS) {
			this.setInputValue([...this.professions, profession]);
		} else {
			return;
		}
		this.inputChanged(TypedEvent.nextEventID());
	}
}

function iconOption(icon: string, name: string): HTMLButtonElement {
	return (
		<button type="button" className="player-icon-picker-option" aria-label={name} style={{ backgroundImage: `url('${iconUrl(icon)}')` }}></button>
	) as HTMLButtonElement;
}
