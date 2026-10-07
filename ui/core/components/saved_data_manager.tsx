import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { EventID, TypedEvent } from '../typed_event';
import { BaseModal } from './base_modal';
import { Component } from './component';
import { ContentBlock, ContentBlockHeaderConfig } from './content_block';
import { ChangeCategory, dirtySettings, PresetSource } from './dirty_settings';
import { PresetTree } from './preset_tree';

export type SavedDataManagerConfig<ModObject, T> = {
	label: string;
	header?: ContentBlockHeaderConfig;
	presetsOnly?: boolean;
	storageKey: string;
	changeEmitters: Array<TypedEvent<any>>;
	equals: (a: T, b: T) => boolean;
	getData: (modObject: ModObject) => T;
	setData: (eventID: EventID, modObject: ModObject, data: T) => void;
	toJson: (a: T) => any;
	fromJson: (obj: any) => T;
	// What the tooltip lists as changed since we loaded a preset. Without it, we list each
	// setting on the tabs that differs from the preset.
	listChanges?: (data: T, changes: ChangeCategory[]) => ChangeCategory[];
};

export type SavedDataConfig<ModObject, T> = {
	name: string;
	data: T;
	tooltip?: string;
	isPreset?: boolean;
	// The folder a preset is listed in, like 'Level 30 PvP'. Left out means the top level.
	group?: string;

	// If set, will automatically hide the saved data when this evaluates to false.
	enableWhen?: (obj: ModObject) => boolean;
	// Will execute when the saved data is loaded.
	onLoad?: (obj: ModObject) => void;
};

// The folder our own saved presets are listed in, after the spec's presets.
const SAVED_GROUP = 'Saved';

type SavedData<ModObject, T> = {
	name: string;
	data: T;
	elem: HTMLElement;
} & Pick<SavedDataConfig<ModObject, T>, 'enableWhen' | 'onLoad'>;

export class SavedDataManager<ModObject, T> extends Component {
	private readonly modObject: ModObject;
	private readonly config: SavedDataManagerConfig<ModObject, T>;

	private readonly userData: Array<SavedData<ModObject, T>>;
	private readonly presets: Array<SavedData<ModObject, T>>;

	private readonly savedDataDiv: HTMLElement;
	private readonly presetTree: PresetTree;
	private readonly dirtySource: PresetSource;

	private frozen: boolean;

	constructor(parent: HTMLElement, modObject: ModObject, config: SavedDataManagerConfig<ModObject, T>) {
		super(parent, 'saved-data-manager-root');
		this.modObject = modObject;
		this.config = config;

		this.userData = [];
		this.presets = [];
		this.frozen = false;

		const contentBlock = new ContentBlock(this.rootElem, 'saved-data', { header: config.header });

		const savedDataRef = ref<HTMLDivElement>();
		const presetDataRef = ref<HTMLDivElement>();
		contentBlock.bodyElement.replaceChildren(
			<div ref={savedDataRef} className="saved-data-container hide">
				<div ref={presetDataRef} className="saved-data-presets" />
			</div>,
		);

		this.savedDataDiv = savedDataRef.value!;
		this.presetTree = new PresetTree(`${config.storageKey}__openFolders__`, `${config.storageKey}__selected__`, SAVED_GROUP);
		presetDataRef.value!.appendChild(this.presetTree.rootElem);

		// The saved sets are presets too, so the settings they control are marked when we
		// change them after loading one.
		this.dirtySource = {
			selected: () => {
				const selected = this.presetTree.selected();
				const savedData = [...this.presets, ...this.userData].find(data => data.elem === selected?.item);
				if (!savedData) return undefined;
				return { apply: () => this.config.setData(TypedEvent.nextEventID(), this.modObject, savedData.data) };
			},
		};
		this.addOnDisposeCallback(dirtySettings.addSource(this.dirtySource));

		this.config.changeEmitters.forEach(emitter => emitter.on(() => this.presetTree.refresh()));

		if (!config.presetsOnly) {
			contentBlock.bodyElement.appendChild(this.buildSaveButton());
		}
	}

	addSavedData(config: SavedDataConfig<ModObject, T>) {
		this.savedDataDiv.classList.remove('hide');

		const newData = this.makeSavedData(config);
		const dataArr = config.isPreset ? this.presets : this.userData;
		const oldIdx = dataArr.findIndex(data => data.name == config.name);

		if (oldIdx == -1) {
			this.presetTree.add(newData.elem, config.isPreset ? config.group : SAVED_GROUP);
			dataArr.push(newData);
		} else {
			this.presetTree.untrack(dataArr[oldIdx].elem);
			dataArr[oldIdx].elem.replaceWith(newData.elem);
			dataArr[oldIdx] = newData;
		}
		this.presetTree.refresh();
	}

	private makeSavedData(config: SavedDataConfig<ModObject, T>): SavedData<ModObject, T> {
		const overrideButtonRef = ref<HTMLAnchorElement>();
		const deleteButtonRef = ref<HTMLAnchorElement>();
		const dataElem = PresetTree.makeItem(config.name);
		if (!config.isPreset) {
			dataElem.append(
				<a ref={overrideButtonRef} href="javascript:void(0)" className="saved-data-set-action saved-data-set-override" attributes={{ role: 'button' }}>
					<i className="fas fa-floppy-disk"></i>
				</a>,
				<a ref={deleteButtonRef} href="javascript:void(0)" className="saved-data-set-action saved-data-set-delete" attributes={{ role: 'button' }}>
					<i className="fa fa-times"></i>
				</a>,
			);
		}

		this.presetTree.track(dataElem, {
			name: config.name,
			matches: () => this.config.equals(config.data, this.config.getData(this.modObject)),
			enabled: config.enableWhen ? () => config.enableWhen!(this.modObject) : undefined,
		});

		this.presetTree.onClick(dataElem, () => {
			this.config.setData(TypedEvent.nextEventID(), this.modObject, config.data);

			config.onLoad?.(this.modObject);
		});

		if (!config.isPreset && overrideButtonRef.value) {
			tippy(overrideButtonRef.value, { content: 'Override with the current settings' });
			overrideButtonRef.value.addEventListener('click', event => {
				event.stopPropagation();
				if (this.frozen || !this.confirmOverride(config.name)) return;
				this.saveCustom(config.name);
			});
		}

		if (!config.isPreset && deleteButtonRef.value) {
			const tooltip = tippy(deleteButtonRef.value, { content: `Delete saved ${this.config.label}` });
			deleteButtonRef.value.addEventListener('click', event => {
				event.stopPropagation();
				const shouldDelete = confirm(`Delete saved ${this.config.label} '${config.name}'?`);
				if (!shouldDelete) return;

				tooltip.destroy();

				const idx = this.userData.findIndex(data => data.name == config.name);
				this.presetTree.untrack(this.userData[idx].elem);
				this.userData[idx].elem.remove();
				this.userData.splice(idx, 1);
				this.saveUserData();
			});
		}

		this.presetTree.attachTooltip(dataElem, config.tooltip, () => {
			const changes = dirtySettings.changesFor(this.dirtySource);
			return this.config.listChanges ? this.config.listChanges(config.data, changes) : changes;
		});

		return {
			name: config.name,
			data: config.data,
			elem: dataElem,
			enableWhen: config.enableWhen,
			onLoad: config.onLoad,
		};
	}

	// Save data to window.localStorage.
	private saveUserData() {
		const userData: Record<string, unknown> = {};
		this.userData.forEach(savedData => {
			userData[savedData.name] = this.config.toJson(savedData.data);
		});

		if (this.userData.length == 0 && this.presets.length == 0) this.savedDataDiv.classList.add('hide');

		window.localStorage.setItem(this.config.storageKey, JSON.stringify(userData));
	}

	// Load data from window.localStorage.
	loadUserData() {
		const dataStr = window.localStorage.getItem(this.config.storageKey);
		if (!dataStr) return;

		let jsonData;
		try {
			jsonData = JSON.parse(dataStr);
		} catch (e) {
			console.warn('Invalid json for local storage value: ' + dataStr);
		}

		for (const name in jsonData) {
			try {
				this.addSavedData({
					name: name,
					data: this.config.fromJson(jsonData[name]),
				});
			} catch (e) {
				console.warn('Failed parsing saved data: ' + jsonData[name]);
			}
		}
	}

	// Prevent user input from creating / deleting saved data.
	freeze() {
		this.frozen = true;
		this.rootElem.classList.add('frozen');
	}

	// Asks before we replace a custom preset with the current settings.
	private confirmOverride(name: string): boolean {
		return confirm(`Override custom ${this.config.label} '${name}' with the current settings?`);
	}

	// Saves the current settings as a custom preset, or over the one with that name, and
	// selects it.
	private saveCustom(name: string) {
		this.addSavedData({
			name: name,
			data: this.config.getData(this.modObject),
		});
		this.saveUserData();
		const saved = this.userData.find(data => data.name == name);
		if (saved) this.presetTree.select(saved.elem);
		this.presetTree.openFolder(SAVED_GROUP);
	}

	private buildSaveButton(): HTMLElement {
		const button = (
			<button className="saved-data-save-button btn btn-sm btn-outline-primary">
				<i className="fas fa-plus me-1"></i>
				Save preset
			</button>
		) as HTMLButtonElement;
		tippy(button, { content: `Save the current ${this.config.label.toLowerCase()} as a custom preset` });

		button.addEventListener('click', () => {
			if (this.frozen) return;
			const selected = this.userData.find(data => data.elem === this.presetTree.selected()?.item);
			new SavePresetModal(this.rootElem.closest<HTMLElement>('.sim-ui') ?? document.body, {
				label: this.config.label,
				// When one of our own presets is selected, we start from its name, so saving
				// our changes over it is one click.
				initialName: selected?.name ?? '',
				customNames: this.userData.map(data => data.name),
				presetNames: this.presets.map(data => data.name),
				confirmOverride: name => this.confirmOverride(name),
				save: name => this.saveCustom(name),
			}).open();
		});

		return button;
	}
}

type SavePresetModalConfig = {
	label: string;
	initialName: string;
	customNames: string[];
	presetNames: string[];
	confirmOverride: (name: string) => boolean;
	save: (name: string) => void;
};

// Asks for a name for a new custom preset.
//
// A name of one of our custom presets overrides that preset, after we confirm. A name of a
// built-in preset isn't allowed, because the list selects presets by name and we couldn't
// tell the two apart.
class SavePresetModal extends BaseModal {
	constructor(parent: HTMLElement, config: SavePresetModalConfig) {
		super(parent, 'save-preset-modal', { title: `Save ${config.label} preset`, footer: true, size: 'md', disposeOnClose: true });

		const inputRef = ref<HTMLInputElement>();
		const hintRef = ref<HTMLParagraphElement>();
		const saveButtonRef = ref<HTMLButtonElement>();
		const listId = `save-preset-names-${Math.random().toString(36).slice(2)}`;
		this.body.append(
			<label className="form-label">Name</label>,
			<input ref={inputRef} className="form-control" type="text" placeholder="Name" attributes={{ list: listId }} />,
			<datalist id={listId}>
				{config.customNames.map(name => (
					<option value={name} />
				))}
			</datalist>,
			<p ref={hintRef} className="save-preset-hint mb-0 mt-2" />,
		);
		this.footer!.append(
			<button className="btn btn-outline-primary me-2" onclick={() => this.close()}>
				Cancel
			</button>,
			<button ref={saveButtonRef} className="btn btn-primary">
				Save
			</button>,
		);

		const input = inputRef.value!;
		const hint = hintRef.value!;
		const saveButton = saveButtonRef.value!;
		input.value = config.initialName;

		const update = () => {
			const name = input.value.trim();
			const isPreset = config.presetNames.includes(name);
			const isCustom = config.customNames.includes(name);
			hint.textContent = isPreset
				? 'A built-in preset has this name. Choose another one.'
				: isCustom
				? 'A custom preset has this name. Saving overrides it.'
				: '';
			hint.classList.toggle('text-danger', isPreset);
			hint.classList.toggle('save-preset-override', isCustom);
			saveButton.disabled = !name || isPreset;
		};
		update();
		input.addEventListener('input', update);

		const save = () => {
			const name = input.value.trim();
			if (!name || config.presetNames.includes(name)) return;
			if (config.customNames.includes(name) && !config.confirmOverride(name)) return;
			config.save(name);
			this.close();
		};
		saveButton.addEventListener('click', save);
		input.addEventListener('keydown', event => {
			if (event.key == 'Enter') save();
		});

		this.rootElem.addEventListener('shown.bs.modal', () => input.select(), { once: true });
	}
}
