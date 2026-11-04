// The Game Menu (Esc) and the Options frame. The options hold the room's settings: graphics,
// sound and the encounter.

import { defaultSoundLevels, setSoundLevels, soundLevels, SoundLevels } from '../glue';
import type { Settings } from '../protocol';
import { Graphics, graphicsPresets } from '../scene';
import { SND, sound, tex, texVars, TX } from './assets';
import { button, checkbox, div, dropdown, Panel, slider, text } from './panels';

// What the menus change in the room.
export interface MenuHost {
	graphics(): Graphics;
	setGraphics(g: Graphics): void;
	fight(): Settings;
	// applyFight saves the encounter settings and restarts the session with them.
	applyFight(s: Settings): void;
	// bindKeys turns on the room's bind mode (hover a button, press a key).
	bindKeys(): void;
	leave(): void;
}

function dialogChrome(root: HTMLElement, title: string) {
	div(root, 'bz-dialog');
	const header = div(root, 'bz-dialog-header');
	text(header, title, 'GameFontNormal');
}

// The menu's buttons in groups, with a gap between groups like the game's.
type MenuItem = { label: string; run: () => void };

export class GameMenu extends Panel {
	readonly width = 195;
	height = 0;
	area = 'dialog' as const;
	openSound = SND.menuOpen;
	closeSound = SND.menuClose;

	constructor(private readonly groups: MenuItem[][]) {
		super();
		this.height = 36 + groups.reduce((h, g) => h + g.length * 22, 0) + (groups.length - 1) * 14 + 24;
	}

	protected build(root: HTMLElement) {
		root.classList.add('bz-gamemenu');
		dialogChrome(root, 'Game Menu');
		let y = 36;
		for (const group of this.groups) {
			for (const item of group) {
				button(root, item.label, { x: (this.width - 144) / 2, y, w: 144, h: 21 }, () => {
					sound(SND.menuOption);
					item.run();
				});
				y += 22;
			}
			y += 14;
		}
	}
}

const categories = ['Graphics', 'Sound', 'Encounter'] as const;
type Category = (typeof categories)[number];

// The Options frame, in the look of the game's old options windows: the dialog box with a
// category list on the left, the category's options on the right, and Defaults, Okay and
// Cancel under them. Graphics and sound apply at once (Cancel puts them back). The encounter
// applies on Okay, with a restart.
export class OptionsFrame extends Panel {
	readonly width = 560;
	readonly height = 400;
	area = 'dialog' as const;
	openSound = SND.menuOpen;
	closeSound = SND.menuClose;
	private category: Category = 'Graphics';
	private list?: HTMLElement;
	private page?: HTMLElement;
	// The values when the frame opened, for Cancel, and the encounter edits waiting for Okay.
	private startGraphics?: Graphics;
	private startSound?: SoundLevels;
	private fight?: Settings;
	onClosed?: () => void;

	constructor(private readonly host: MenuHost) {
		super();
	}

	protected async build(root: HTMLElement) {
		root.classList.add('bz-options');
		await texVars({
			'bz-tooltip-border': () => import('../glue').then(g => g.border9(TX.tooltipBorder)),
			'bz-title-highlight': () => tex(136810),
		});
		dialogChrome(root, 'Options');
		this.list = div(root, 'bz-options-list', { x: 22, y: 40, w: 150, h: 300 });
		this.page = div(root, 'bz-options-page', { x: 184, y: 40, w: 354, h: 300 });
		button(root, 'Defaults', { x: 22, y: 352, w: 96, h: 22 }, () => {
			sound(SND.menuOption);
			this.defaults();
		});
		button(root, 'Okay', { x: 334, y: 352, w: 96, h: 22 }, () => {
			sound(SND.menuOption);
			this.okay();
		});
		button(root, 'Cancel', { x: 438, y: 352, w: 96, h: 22 }, () => {
			sound(SND.menuOption);
			this.cancel();
		});
	}

	protected shown() {
		this.startGraphics = { ...this.host.graphics() };
		this.startSound = soundLevels();
		this.fight = { ...this.host.fight() };
		this.renderList();
		this.renderPage();
	}

	protected hidden() {
		this.onClosed?.();
	}

	private renderList() {
		const list = this.list!;
		list.replaceChildren();
		categories.forEach((c, i) => {
			const row = div(list, `bz-options-cat${c === this.category ? ' selected' : ''}`, { x: 8, y: 8 + i * 18, w: 134, h: 18 });
			text(row, c, c === this.category ? 'GameFontHighlight' : 'GameFontNormal', { x: 6, y: 2 });
			row.addEventListener('click', e => {
				e.stopPropagation();
				if (c === this.category) return;
				sound(SND.menuOption);
				this.category = c;
				this.renderList();
				this.renderPage();
			});
		});
	}

	private renderPage() {
		const page = this.page!;
		page.replaceChildren();
		text(page, this.category, 'GameFontNormalLarge', { x: 16, y: 14 });
		if (this.category === 'Graphics') this.graphicsPage(page);
		else if (this.category === 'Sound') this.soundPage(page);
		else this.encounterPage(page);
	}

	private graphicsPage(page: HTMLElement) {
		const g = this.host.graphics();
		const set = (next: Graphics) => {
			this.host.setGraphics(next);
			this.renderPage();
		};
		text(page, 'How the room draws', 'GameFontHighlightSmall', { x: 16, y: 36 });
		const row = (y: number, label: string) => text(page, label, 'GameFontNormalSmall', { x: 16, y: y + 5 });
		// The screen's own height in pixels (the first choice), and the common heights under it.
		const nativeH = Math.round(screen.height * devicePixelRatio);
		const heights = [2160, 1440, 1080, 720].filter(v => v < nativeH);
		row(64, 'Resolution');
		dropdown(
			page,
			{ x: 150, y: 64, w: 150 },
			[['0', `${nativeH}p`], ...heights.map(h => [String(h), `${h}p`] as [string, string])],
			heights.includes(g.maxHeight) ? String(g.maxHeight) : '0',
			v => set({ ...this.host.graphics(), maxHeight: +v }),
		);
		row(96, 'Graphics Quality');
		dropdown(
			page,
			{ x: 150, y: 96, w: 150 },
			[
				['low', 'Low'],
				['medium', 'Medium'],
				['high', 'High'],
				['custom', 'Custom'],
			],
			g.preset,
			v => v !== 'custom' && set({ ...graphicsPresets[v as 'low' | 'medium' | 'high'], maxHeight: this.host.graphics().maxHeight }),
		);
		row(128, 'Shadow Quality');
		dropdown(
			page,
			{ x: 150, y: 128, w: 150 },
			[
				['off', 'Off'],
				['low', 'Low'],
				['medium', 'Medium'],
				['high', 'High'],
			],
			g.shadows,
			v => set({ ...this.host.graphics(), preset: 'custom', shadows: v as Graphics['shadows'] }),
		);
		slider(page, 'Render Scale', { x: 150, y: 188, w: 150 }, { min: 0.5, max: 1, step: 0.25, value: g.renderScale, format: v => `${Math.round(v * 100)}%` }, v =>
			set({ ...this.host.graphics(), preset: 'custom', renderScale: v }),
		);
		text(page, 'Lower draws fewer pixels, for more FPS', 'GameFontDisableSmall', { x: 16, y: 230 });
	}

	private soundPage(page: HTMLElement) {
		const l = soundLevels();
		const set = (next: Partial<SoundLevels>) => setSoundLevels({ ...soundLevels(), ...next });
		checkbox(page, 'Enable Sound', { x: 14, y: 40 }, l.enabled, on => set({ enabled: on }));
		const pct = (v: number) => `${Math.round(v * 100)}%`;
		const volume = (y: number, label: string, key: 'master' | 'effects' | 'music' | 'ambience') =>
			slider(page, label, { x: 30, y, w: 180 }, { min: 0, max: 1, step: 0.05, value: l[key], format: pct, low: '0%', high: '100%' }, v => set({ [key]: v }));
		volume(100, 'Master Volume', 'master');
		volume(150, 'Sound Volume', 'effects');
		volume(200, 'Ambience Volume', 'ambience');
		volume(250, 'Music Volume', 'music');
		text(page, 'Music plays on the setup page only', 'GameFontDisableSmall', { x: 30, y: 280 });
	}

	private encounterPage(page: HTMLElement) {
		const f = this.fight!;
		text(page, 'Applies on Okay, with a restart', 'GameFontHighlightSmall', { x: 16, y: 36 });
		const row = (y: number, label: string) => text(page, label, 'GameFontNormalSmall', { x: 16, y: y + 5 });
		row(64, 'Mode');
		dropdown(
			page,
			{ x: 150, y: 64, w: 150 },
			[
				['free', 'Free'],
				['encounter', 'Encounter'],
			],
			f.mode,
			v => (f.mode = v as Settings['mode']),
		);
		slider(page, 'Length', { x: 150, y: 114, w: 170 }, { min: 30, max: 600, step: 10, value: f.lengthMs / 1000, format: v => fmtLength(v) }, v => (f.lengthMs = v * 1000));
		slider(page, 'Countdown', { x: 150, y: 164, w: 170 }, { min: 0, max: 15, step: 1, value: f.countdownMs / 1000, format: v => `${v}s` }, v => (f.countdownMs = v * 1000));
		slider(page, 'Ping', { x: 150, y: 214, w: 170 }, { min: 0, max: 300, step: 10, value: f.pingMs, format: v => `${v}ms` }, v => (f.pingMs = v));
		slider(page, 'Spell Queue Window', { x: 150, y: 254, w: 170 }, { min: 0, max: 400, step: 10, value: f.spellQueueWindowMs, format: v => `${v}ms` }, v => (f.spellQueueWindowMs = v));
		checkbox(page, 'Execute range from the start', { x: 14, y: 272 }, f.executeFromStart, on => (f.executeFromStart = on));
	}

	private defaults() {
		if (this.category === 'Graphics') this.host.setGraphics({ ...graphicsPresets.high, maxHeight: this.host.graphics().maxHeight });
		else if (this.category === 'Sound') setSoundLevels(defaultSoundLevels());
		else this.fight = { ...this.fight!, mode: 'free', lengthMs: 180000, countdownMs: 5000, pingMs: 60, executeFromStart: false, spellQueueWindowMs: 400 };
		this.renderPage();
	}

	private okay() {
		const f = this.fight!;
		const was = this.host.fight();
		const changed = (['mode', 'lengthMs', 'countdownMs', 'pingMs', 'executeFromStart', 'spellQueueWindowMs'] as const).some(k => f[k] !== was[k]);
		this.close();
		if (changed) this.host.applyFight(f);
	}

	private cancel() {
		if (this.startGraphics) this.host.setGraphics(this.startGraphics);
		if (this.startSound) setSoundLevels(this.startSound);
		this.close();
	}
}

function fmtLength(s: number): string {
	const m = Math.floor(s / 60);
	const r = s % 60;
	return m ? (r ? `${m}m ${r}s` : `${m}m`) : `${r}s`;
}
