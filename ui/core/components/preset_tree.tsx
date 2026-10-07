import tippy, { Placement } from 'tippy.js';

import { ChangeCategory, dirtySettings } from './dirty_settings';

// A list of presets laid out like the file explorer of an IDE. A preset with a group goes
// in a folder of that name, and the presets without one sit at the top level under the
// folders. Folders keep the order in which their first preset was added.
//
// A folder starts closed unless it holds the active preset. When we open or close a folder
// by hand, we remember that in local storage under storageKey, and from then on it stays
// the way we left it.
//
// The rows are made by the caller, which tells us through track() when its preset matches
// the settings. We keep one of them selected:
// - The one we clicked last, even after we change a setting it controls. Then it's
//   modified, and we show that with a dot like an IDE does for a modified file.
// - When the selected one no longer matches but another preset does, that one. Loading a
//   preset configuration moves the gear sets to the gear it loaded, for example.
// The selection is kept in local storage under selectionKey, so it survives a reload.
//
// The caller calls refresh() when the settings change, and the rows and folders follow: a
// folder with the selected preset is marked, and a folder where every preset is disabled
// is hidden.
export interface PresetTreeEntry {
	name: string;
	// Whether the settings match the preset right now.
	matches: () => boolean;
	// Whether the preset fits the current settings, see enableWhen. Left out means always.
	enabled?: () => boolean;
}

// A long list on one tab (a whole other gear set) is cut short in the tooltip.
const MAX_LISTED_CHANGES = 8;

export class PresetTree {
	readonly rootElem: HTMLElement;
	private readonly foldersElem: HTMLElement;
	private readonly filesElem: HTMLElement;
	private readonly folders = new Map<string, { elem: HTMLDetailsElement; children: HTMLElement }>();
	private readonly storageKey?: string;
	private readonly selectionKey?: string;
	private readonly entries = new Map<HTMLElement, PresetTreeEntry>();
	private selectedName?: string;

	constructor(storageKey?: string, selectionKey?: string) {
		this.storageKey = storageKey;
		this.selectionKey = selectionKey;
		this.selectedName = this.loadSelection();
		this.foldersElem = (<div className="preset-tree-folders" />) as HTMLElement;
		this.filesElem = (<div className="preset-tree-files" />) as HTMLElement;
		this.rootElem = (
			<div className="preset-tree">
				{this.foldersElem}
				{this.filesElem}
			</div>
		) as HTMLElement;
	}

	// Makes the row for one preset, with a file icon in front of the name and the dot that
	// shows when it's modified after it.
	static makeItem(name: string): HTMLElement {
		return (
			<div className="preset-tree-item" attributes={{ role: 'button' }}>
				<i className="far fa-file preset-tree-icon" />
				<span className="preset-tree-name">{name}</span>
				<span className="preset-tree-dirty" />
			</div>
		) as HTMLElement;
	}

	// Shows the preset's tooltip on hover. When it's the selected preset and we changed it
	// since, the tooltip also lists what changed (from changes(), like 'Duration: 120 → 90')
	// and how to reset it.
	attachTooltip(item: HTMLElement, tooltip: string | undefined, placement: Placement, changes?: () => ChangeCategory[]) {
		tippy(item, {
			placement,
			onShow: instance => {
				const modified = this.selected()?.item === item && item.classList.contains('dirty');
				if (!tooltip && !modified) return false;
				const categories = modified ? changes?.() ?? [] : [];
				instance.setContent(
					<>
						{tooltip ? <p className="mb-0">{tooltip}</p> : undefined}
						{modified ? (
							<div className={`preset-tree-dirty-note ${tooltip ? 'mt-2' : ''}`}>
								<p className="mb-1 preset-tree-changed-title">
									{categories.length ? 'Changed from defaults:' : 'Some of its settings were changed.'}
								</p>
								{categories.map(category => {
									const shown = category.lines.length > MAX_LISTED_CHANGES ? category.lines.slice(0, MAX_LISTED_CHANGES - 1) : category.lines;
									return (
										<>
											<p className="mb-0 fw-bold">{category.name}</p>
											{shown.length ? (
												<ul className="mb-1">
													{shown.map(line => (
														<li>{line}</li>
													))}
													{shown.length < category.lines.length ? (
														<li>{`and ${category.lines.length - shown.length} more`}</li>
													) : undefined}
												</ul>
											) : undefined}
										</>
									);
								})}
								<p className="mb-0 mt-1 fst-italic preset-tree-reset-hint">Ctrl+click to reset to preset defaults.</p>
							</div>
						) : undefined}
					</>,
				);
				return undefined;
			},
		});
	}

	// Loads the preset when it's clicked.
	//
	// A click on the selected preset after we changed it does nothing, so a stray click
	// doesn't throw our changes away. Ctrl+click (Cmd+click on a Mac) loads it again.
	onClick(item: HTMLElement, load: () => void) {
		item.addEventListener('click', event => {
			if (item.classList.contains('dirty') && !(event.ctrlKey || event.metaKey)) return;
			this.select(item);
			load();
		});
	}

	track(item: HTMLElement, entry: PresetTreeEntry) {
		this.entries.set(item, entry);
	}

	untrack(item: HTMLElement) {
		this.entries.delete(item);
		this.refresh();
		dirtySettings.schedule();
	}

	// Called when the preset is clicked, before its settings are applied.
	select(item: HTMLElement) {
		const entry = this.entries.get(item);
		if (!entry) return;
		this.selectedName = entry.name;
		this.saveSelection();
		this.refresh();
		dirtySettings.schedule();
	}

	// The selected preset and whether we changed it since, or undefined when nothing is
	// selected.
	selected(): { item: HTMLElement; entry: PresetTreeEntry; modified: boolean } | undefined {
		for (const [item, entry] of this.entries) {
			if (entry.name === this.selectedName) return { item, entry, modified: !entry.matches() };
		}
		return undefined;
	}

	add(item: HTMLElement, group?: string) {
		if (!group) {
			this.filesElem.appendChild(item);
		} else {
			this.folder(group).children.appendChild(item);
		}
		this.refresh();
	}

	refresh() {
		this.updateSelection();

		const stored = this.loadOpenFolders();
		this.folders.forEach((folder, group) => {
			const items = Array.from(folder.children.children);
			const hasActive = items.some(item => item.classList.contains('active'));
			folder.elem.classList.toggle('has-active', hasActive);
			folder.elem.classList.toggle(
				'has-dirty',
				items.some(item => item.classList.contains('dirty')),
			);
			folder.elem.classList.toggle(
				'hide',
				items.every(item => item.classList.contains('disabled')),
			);
			const open = stored[group] ?? hasActive;
			if (folder.elem.open != open) folder.elem.open = open;
		});
	}

	// Opens the folder and remembers it, like when we save a custom preset into it.
	openFolder(group: string) {
		const folder = this.folders.get(group);
		if (!folder) return;
		const stored = this.loadOpenFolders();
		stored[group] = true;
		this.saveOpenFolders(stored);
		folder.elem.open = true;
	}

	private folder(group: string) {
		const existing = this.folders.get(group);
		if (existing) return existing;

		const children = (<div className="preset-tree-children" />) as HTMLElement;
		const summary = (
			<summary className="preset-tree-folder-row">
				<i className="fas fa-chevron-right preset-tree-chevron" />
				<i className="fas fa-folder preset-tree-icon preset-tree-closed-icon" />
				<i className="fas fa-folder-open preset-tree-icon preset-tree-open-icon" />
				<span className="preset-tree-name">{group}</span>
			</summary>
		) as HTMLElement;
		const elem = (
			<details className="preset-tree-folder">
				{summary}
				{children}
			</details>
		) as HTMLDetailsElement;

		// The click toggles the folder after this handler, so the new state is the opposite
		// of the current one.
		summary.addEventListener('click', () => {
			const stored = this.loadOpenFolders();
			stored[group] = !elem.open;
			this.saveOpenFolders(stored);
		});

		const folder = { elem, children };
		this.folders.set(group, folder);
		this.foldersElem.appendChild(elem);
		return folder;
	}

	private updateSelection() {
		const entries = Array.from(this.entries);
		const isEnabled = (entry: PresetTreeEntry) => entry.enabled?.() ?? true;
		const current = entries.find(([_, entry]) => entry.name === this.selectedName);
		const currentMatches = !!current && current[1].matches();
		if (!currentMatches) {
			const match = entries.find(([_, entry]) => isEnabled(entry) && entry.matches());
			if (match) {
				this.selectedName = match[1].name;
				this.saveSelection();
			}
		}

		entries.forEach(([item, entry]) => {
			const isSelected = entry.name === this.selectedName;
			item.classList.toggle('active', isSelected);
			item.classList.toggle('dirty', isSelected && !entry.matches());
			item.classList.toggle('disabled', !isEnabled(entry));
		});
	}

	private loadSelection(): string | undefined {
		if (!this.selectionKey) return undefined;
		try {
			return window.localStorage.getItem(this.selectionKey) ?? undefined;
		} catch {
			return undefined;
		}
	}

	private saveSelection() {
		if (!this.selectionKey || this.selectedName === undefined) return;
		try {
			window.localStorage.setItem(this.selectionKey, this.selectedName);
		} catch {
			// Without storage we pick the selection again from the settings on the next visit.
		}
	}

	private loadOpenFolders(): Record<string, boolean> {
		if (!this.storageKey) return {};
		try {
			return JSON.parse(window.localStorage.getItem(this.storageKey) ?? '{}') ?? {};
		} catch {
			return {};
		}
	}

	private saveOpenFolders(open: Record<string, boolean>) {
		if (!this.storageKey) return;
		try {
			window.localStorage.setItem(this.storageKey, JSON.stringify(open));
		} catch {
			// Without storage the folders still work, they just start over on the next visit.
		}
	}
}

// The tooltip on the title of a list of presets. It says what loading one of them changes,
// and that we mark them when we change it afterwards.
export function presetListTooltip(intro: string, parts: string[]): HTMLElement {
	return (
		<div>
			<p className="mb-1">{intro}</p>
			<ul className="mb-1">
				{parts.map(part => (
					<li>{part}</li>
				))}
			</ul>
			<p className="mb-0">When we change one of these after loading a preset, the preset and the setting are marked as modified.</p>
		</div>
	) as HTMLElement;
}
