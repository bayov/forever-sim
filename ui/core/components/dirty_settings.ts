import { TypedEvent } from '../typed_event';

// A setting on the page that a preset can change, like the Duration field or the item in
// the head slot. read() returns its value from the settings, not from the page.
export interface TrackedSetting {
	elem: HTMLElement;
	read: () => unknown;
	// Left out means we always track it. The talent and rotation editors turn it off for
	// themselves, because we mark their parts instead.
	isTracked?: () => boolean;
	// The setting's name in the list of changes on a modified preset. Left out means we
	// find it on the page, see nameFromPage().
	name?: () => string;
	// The value in the list of changes, like '60' or 'Crown of Destruction'. Left out means
	// we write simple values as they are, see formatValue(). An empty string means we only
	// list the name, for values like a whole rotation. An element goes in as it is, like an
	// item link we can hover for its tooltip.
	format?: (value: unknown) => string | HTMLElement;
}

// A list of presets where one can be selected, like the gear sets or the preset
// configurations in the sidebar. apply() puts the selected preset's settings in place.
export interface PresetSource {
	selected: () => { apply: () => void } | undefined;
}

// The changes on one tab (or in the stat weights dialog), see changesFor(). The elements of
// a tab carry its name in data-preset-category.
export interface ChangeCategory {
	name: string;
	lines: ChangeLine[];
}

// A line is plain text, or an element when its values are links.
export type ChangeLine = string | HTMLElement;

// Takes a copy of every setting and returns a way to put them back.
export type SettingsSnapshot = () => () => void;

// Marks the settings that differ from the selected presets.
//
// A setting is marked when applying one of the selected presets again would change it. When
// the Level 30 PvP preset is selected and we set the Duration to 60 s, applying the preset
// would put it back to 120 s, so we mark the Duration field. The fields the preset doesn't
// set are never marked, because applying it leaves them alone.
//
// To find them we apply each selected preset to the settings with every event silenced,
// read the tracked settings, and put the old settings back. Nobody sees the settings in
// between, so this costs no sims and no saves.
class DirtySettings {
	private readonly settings = new Set<TrackedSetting>();
	private readonly sources = new Set<PresetSource>();
	private readonly listeners: Array<() => void> = [];
	// For each source with a selected preset, the settings that differ from it. The values
	// are kept as the JSON we compared, so they don't change under us.
	private changes = new Map<PresetSource, Array<{ setting: TrackedSetting; preset: string; current: string }>>();
	private snapshot?: SettingsSnapshot;
	private timer?: number;

	track(setting: TrackedSetting): () => void {
		this.settings.add(setting);
		this.schedule();
		return () => {
			this.settings.delete(setting);
			setting.elem.classList.remove('dirty-setting');
		};
	}

	addSource(source: PresetSource): () => void {
		this.sources.add(source);
		this.schedule();
		return () => {
			this.sources.delete(source);
			this.schedule();
		};
	}

	// Called after every update of the marks, like for the dots on the tabs.
	onUpdate(listener: () => void) {
		this.listeners.push(listener);
	}

	// The settings that differ from the source's selected preset, for its tooltip. They come
	// in lines like 'Duration: 120 → 90', under the tab they're on (Gear, Settings and so
	// on), in the order of the tabs.
	changesFor(source: PresetSource): ChangeCategory[] {
		// The lines of each category by their text, because a setting can show twice on one tab.
		const categories = new Map<string, Map<string, ChangeLine>>();
		(this.changes.get(source) ?? []).forEach(({ setting, preset, current }) => {
			const name = setting.name?.() ?? nameFromPage(setting.elem);
			const format = setting.format ?? formatValue;
			const line = changeLine(name, format(JSON.parse(preset)), format(JSON.parse(current)));
			const category = setting.elem.closest<HTMLElement>('[data-preset-category]')?.dataset.presetCategory ?? 'Other';
			if (!categories.has(category)) categories.set(category, new Map());
			categories.get(category)!.set(lineText(line), line);
		});

		// A setting outside the tabs (in the encounter's Advanced dialog) is listed under
		// Other, unless it also shows on a tab.
		const other = categories.get('Other');
		other?.forEach((_, text) => {
			if (Array.from(categories).some(([name, lines]) => name !== 'Other' && lines.has(text))) other.delete(text);
		});
		if (other?.size === 0) categories.delete('Other');

		const order = Array.from(document.querySelectorAll<HTMLElement>('[data-preset-category]')).map(elem => elem.dataset.presetCategory);
		const rank = (name: string) => (order.includes(name) ? order.indexOf(name) : order.length);
		return Array.from(categories, ([name, lines]) => ({ name, lines: Array.from(lines.values()) })).sort((a, b) => rank(a.name) - rank(b.name));
	}

	// Only the individual sims set this. Without it we mark nothing.
	setSnapshot(snapshot: SettingsSnapshot) {
		this.snapshot = snapshot;
		this.schedule();
	}

	// One change often comes with many events (a preset sets gear, talents and buffs at
	// once), so we wait for them to settle and update the marks once.
	schedule() {
		if (this.timer !== undefined) return;
		this.timer = window.setTimeout(() => {
			this.timer = undefined;
			this.update();
		}, 50);
	}

	private update() {
		const settings = Array.from(this.settings).filter(setting => setting.elem.isConnected && (setting.isTracked?.() ?? true));
		const dirty = new Set<TrackedSetting>();
		this.changes = new Map();

		if (this.snapshot) {
			const current = settings.map(readKey);
			const presets = Array.from(this.sources)
				.map(source => ({ source, preset: source.selected() }))
				.filter(({ preset }) => !!preset);
			if (presets.length) {
				TypedEvent.silentlyDo(() => {
					const restore = this.snapshot!();
					presets.forEach(({ source, preset }) => {
						const changes: Array<{ setting: TrackedSetting; preset: string; current: string }> = [];
						try {
							preset!.apply();
							settings.forEach((setting, i) => {
								const presetKey = readKey(setting);
								if (presetKey === current[i]) return;
								dirty.add(setting);
								changes.push({ setting, preset: presetKey, current: current[i] });
							});
						} catch (e) {
							console.warn('Failed to compare settings with a preset: ' + e);
						} finally {
							restore();
						}
						this.changes.set(source, changes);
					});
				});
			}
		}

		// Some elements hold more than one setting, like a gear slot with its item and its
		// enchant. They're marked when any of them changed.
		const dirtyElems = new Set(Array.from(dirty, setting => setting.elem));
		settings.forEach(setting => setting.elem.classList.toggle('dirty-setting', dirtyElems.has(setting.elem)));
		this.listeners.forEach(listener => listener());
	}
}

// Settings come as numbers, strings, arrays and protos, so we compare them as JSON. A
// setting that fails to read (an input built for settings that are gone) reads as the same
// thing every time, so we don't mark it.
function readKey(setting: TrackedSetting): string {
	try {
		return JSON.stringify(setting.read() ?? null);
	} catch {
		return 'null';
	}
}

// A line like 'Duration: 120 → 90'. With an element for a value, the line is an element too.
function changeLine(name: string, from: string | HTMLElement, to: string | HTMLElement): ChangeLine {
	const isEmpty = (value: string | HTMLElement) => typeof value === 'string' && !value;
	if (isEmpty(from) && isEmpty(to)) return name;
	const value = (value: string | HTMLElement) => (isEmpty(value) ? 'none' : value);
	if (typeof from === 'string' && typeof to === 'string') return `${name}: ${value(from)} → ${value(to)}`;
	const line = document.createElement('span');
	line.append(`${name}: `, value(from), ' → ', value(to));
	return line;
}

function lineText(line: ChangeLine): string {
	return typeof line === 'string' ? line : line.textContent ?? '';
}

function formatValue(value: unknown): string {
	if (typeof value === 'boolean') return value ? 'on' : 'off';
	if (typeof value === 'number') return String(Math.round(value * 100) / 100);
	if (typeof value === 'string') return value;
	return '';
}

// A name for a setting from what the page shows around it: the field's label, the first cell
// of its table row (the stat weights), or the title of its section (the consumables, which
// show only icons).
function nameFromPage(elem: HTMLElement): string {
	const label = elem.querySelector('label, .form-label')?.textContent?.trim();
	if (label) return label;
	const row = elem.closest('tr')?.querySelector('td, th')?.textContent?.trim();
	if (row) return row;
	const rowLabel = elem.closest('.picker-group, .consumes-row')?.querySelector('label, .form-label')?.textContent?.trim();
	if (rowLabel) return rowLabel;
	const section = elem.closest('.content-block')?.querySelector('.content-block-title')?.textContent?.trim();
	return section ? `${section} setting` : 'A setting';
}

export const dirtySettings = new DirtySettings();
