import { TypedEvent } from '../typed_event.js';
import { Input, InputConfig } from './input.js';

export interface EnumValueConfig {
	name: string;
	value: number;
	tooltip?: string;
	// A wowhead icon name, like 'spell_nature_cyclone'. When the values have icons, the Settings
	// tab shows the input as a row of icons, see IconEnumRowPicker.
	icon?: string;
	// A tooltip for the icon in place of the one with the name, like the game's tooltip of a
	// shaman weapon imbue. We build it each time it shows, so it can follow the settings.
	richTooltip?: (modObject: any) => HTMLElement | undefined;
}

export interface EnumPickerConfig<ModObject> extends InputConfig<ModObject, number> {
	id: string;
	values: Array<EnumValueConfig>;
}

export class EnumPicker<ModObject> extends Input<ModObject, number> {
	private readonly selectElem: HTMLSelectElement;

	constructor(parent: HTMLElement, modObject: ModObject, config: EnumPickerConfig<ModObject>) {
		super(parent, 'enum-picker-root', modObject, config);

		this.selectElem = document.createElement('select');
		this.selectElem.id = config.id;
		this.selectElem.classList.add('enum-picker-selector', 'form-select');

		config.values.forEach(value => {
			const option = document.createElement('option');
			option.value = String(value.value);
			option.textContent = value.name;
			this.selectElem.appendChild(option);

			if (value.tooltip) {
				option.title = value.tooltip;
			}
		});
		this.rootElem.appendChild(this.selectElem);

		this.init();

		this.selectElem.addEventListener(
			'change',
			() => {
				this.inputChanged(TypedEvent.nextEventID());
			},
			{ signal: this.signal },
		);
	}

	getInputElem(): HTMLElement {
		return this.selectElem;
	}

	protected formatPresetValue(value: number): string {
		return (this.inputConfig as EnumPickerConfig<ModObject>).values.find(v => v.value === value)?.name ?? String(value);
	}

	getInputValue(): number {
		return parseInt(this.selectElem.value);
	}

	setInputValue(newValue: number) {
		this.selectElem.value = String(newValue);
	}
}
