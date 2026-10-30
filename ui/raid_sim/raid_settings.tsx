// The raid as the spec pages keep their settings (docs/raid_sim.md, The page): the sim history's runs, the
// sidebar's presets, the JSON import and export, and a shareable link. Each keeps the comp editor's raid
// (RaidState): its members by slot, its assignments, faction, camps and fight. The iterations stay out.
import pako from 'pako';

import { Exporter } from '../core/components/exporter';
import { Importer } from '../core/components/importer';
import { presetListTooltip } from '../core/components/preset_tree';
import { SavedDataManager } from '../core/components/saved_data_manager';
import { SimHistoryHost } from '../core/components/sim_history';
import Toast from '../core/components/toast';
import { CompEditor, RaidState } from './comp_editor';
import { loadRaidSimData } from './data';
import type { RaidSimUI } from './raid_sim_ui';

// The sim history's view of the page: one mode (DPS), and the raid as each run's settings.
export function raidHistoryHost(simUI: RaidSimUI, editor: CompEditor): SimHistoryHost {
	return {
		simUI,
		mode: () => 'dps',
		// The page has no mode to change to.
		onModeChange: () => undefined,
		settings: () => editor.getState(),
		applySettings: settings => editor.applyState(settings, 'the run’s raid'),
		changes: (from, to) => editor.describeChanges(CompEditor.readState(from), to === undefined ? editor.getState() : CompEditor.readState(to)),
	};
}

// The sidebar's Preset Configurations: the built-in raids (each size's new raid and the searches' best comps)
// and the raids we save, in a Saved folder. The selected one shows a dot when the raid changed since.
export function addRaidPresets(parent: HTMLElement, simUI: RaidSimUI, editor: CompEditor) {
	const manager = new SavedDataManager<CompEditor, RaidState>(parent, editor, {
		label: 'Raid',
		header: {
			title: 'Preset Raids',
			tooltip: presetListTooltip('A preset raid sets up the whole raid at once:', ['Members and groups', 'Assignments', 'Faction and camp buffs', 'Encounter']),
		},
		storageKey: simUI.getStorageKey('__savedRaids__'),
		changeEmitters: [editor.changeEmitter],
		equals: (a, b) => CompEditor.sameRaid(a, b, true),
		getData: editor => editor.getState(),
		setData: (_eventID, editor, state) => editor.applyState(state, 'the preset'),
		toJson: state => state,
		fromJson: obj => CompEditor.readState(obj),
		listChanges: state => editor.describeChanges(state, editor.getState()),
	});
	manager.loadUserData();
	editor
		.builtInPresets()
		.then(presets => presets.forEach(p => manager.addSavedData({ name: p.name, group: p.group, tooltip: p.tooltip, isPreset: true, data: p.state })))
		.catch(e => console.warn('Failed to load the preset raids:', e));
	return manager;
}

// What the JSON export writes and the import reads: the raid, marked as this page's.
const KIND = 'forever-raid-sim';

export class RaidJsonExporter extends Exporter {
	protected readonly simUI: RaidSimUI;
	private readonly editor: CompEditor;

	constructor(parent: HTMLElement, simUI: RaidSimUI, editor: CompEditor) {
		super(parent, { title: 'Export Raid JSON', allowDownload: true });
		this.simUI = simUI;
		this.editor = editor;
	}

	getData(): string {
		return JSON.stringify({ kind: KIND, raid: this.editor.getState() }, null, 2);
	}
}

export class RaidJsonImporter extends Importer {
	protected readonly simUI: RaidSimUI;
	private readonly editor: CompEditor;

	constructor(parent: HTMLElement, simUI: RaidSimUI, editor: CompEditor) {
		super(parent, { title: 'Import Raid JSON', allowFileUpload: true });
		this.simUI = simUI;
		this.editor = editor;
		this.descriptionElem.textContent = 'Paste a raid from Export > JSON, or upload its file. Undo on the Comp tab puts back the raid it replaces.';
	}

	onImport(data: string) {
		let json: any;
		try {
			json = JSON.parse(data);
		} catch (e) {
			throw new Error(`It isn’t JSON: ${e instanceof Error ? e.message : e}`);
		}
		this.editor.applyState(json?.kind == KIND ? json.raid : json, 'the import');
		this.close();
		new Toast({ variant: 'success', body: 'Raid imported.' });
	}
}

// The raid goes in the link's hash, deflated, as a spec page's link does.
const LINK_PREFIX = 'raid=';

export class RaidLinkExporter extends Exporter {
	protected readonly simUI: RaidSimUI;
	private readonly editor: CompEditor;

	constructor(parent: HTMLElement, simUI: RaidSimUI, editor: CompEditor) {
		super(parent, { title: 'Shareable Link' });
		this.simUI = simUI;
		this.editor = editor;
	}

	getData(): string {
		return raidLink(this.editor.getState());
	}
}

export function raidLink(state: RaidState): string {
	const deflated = pako.deflate(new TextEncoder().encode(JSON.stringify(state)));
	let binary = '';
	deflated.forEach(byte => (binary += String.fromCharCode(byte)));
	const url = new URL(window.location.href);
	url.search = '';
	url.hash = LINK_PREFIX + btoa(binary).replaceAll('+', '-').replaceAll('/', '_').replace(/=+$/, '');
	return url.toString();
}

// Loads the raid of a link the page opened with, then takes it out of the address, so a reload keeps the
// raid as we change it. Undo puts back the raid it replaced.
export async function loadRaidFromLink(editor: CompEditor) {
	const hash = window.location.hash.slice(1);
	if (!hash.startsWith(LINK_PREFIX)) return;
	try {
		// The members' names come from the data file.
		await loadRaidSimData();
		const base64 = hash.slice(LINK_PREFIX.length).replaceAll('-', '+').replaceAll('_', '/');
		const binary = atob(base64);
		const bytes = Uint8Array.from(binary, c => c.charCodeAt(0));
		editor.applyState(JSON.parse(new TextDecoder().decode(pako.inflate(bytes))), 'the linked raid');
		new Toast({ variant: 'success', body: 'Loaded the linked raid.' });
	} catch (e) {
		new Toast({ variant: 'error', body: `The link’s raid didn’t load: ${e instanceof Error ? e.message : e}` });
	}
	const url = new URL(window.location.href);
	url.hash = '';
	window.history.replaceState(null, '', url.toString());
}
