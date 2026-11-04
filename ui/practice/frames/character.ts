// The character frame (U), as WoW Forever draws it: its Camelot CharacterFrame, a Mainline style
// frame with Forever's own art. The left pane holds the tab's page and the right pane its details,
// and the mode tabs hang on the frame's right side. The Character tab is the paper doll: your
// model on your race's backdrop with the worn items around it, and the Character Stats pane on
// the right. The Skills tab is in skills.ts. The Reputation tab has no factions, since the room
// has none.
//
// Sizes and anchors follow Forever's Camelot/CharacterFrame.xml and PaperDollFrame.xml. The
// numbers come from the sim's character sheet (link.sheet()). Forever also has PvP, Currency and
// Statistics tabs and the Titles and Equipment Manager panes, which the room has no data for.

import './character.css';
import { atlasTexture, preloadAtlases, setAtlas } from './atlas';
import { glow, SND, sound, TX } from './assets';
import { statAtlases, statCategories, type StatRow } from './character_stats';
import { SidePane, scrollBox, sideTab, sliced, type ScrollBox, type SideTab } from './character_widgets';
import type { RoomLink } from './link';
import { ModelView, portrait } from './model_view';
import { portraitFrame, type PortraitFrame } from './modern';
import { clickable, div, itemButton, type ItemButton, Panel, setTexture, text, texture } from './panels';
import { SkillsPage } from './skills';

// CHARACTER_FRAME_WIDTH, CHARACTER_FRAME_COLLAPSED_WIDTH and CHARACTER_FRAME_HEIGHT. The panel
// manager gives the frame 82 more for its side tabs (CHARACTER_FRAME_WIDTH_PADDING).
const fullWidth = 631;
const collapsedWidth = 398;
const sideTabsPadding = 82;
const collapsedKey = 'bz-char-collapsed';

// Interface\FrameGeneral\UI-Background-Rock, the strip over the model.
const backgroundRock = 374155;
// Interface\DressUpFrame\DressUpBackground-<Race>1 to 4, the backdrop behind the model.
const raceBackgrounds: Record<string, [number, number, number, number]> = {
	RaceHuman: [131093, 131094, 131095, 131096],
	RaceDwarf: [131089, 131090, 131091, 131092],
	RaceNightElf: [131097, 131098, 131099, 131100],
	RaceGnome: [455998, 455999, 456000, 456001],
	RaceOrc: [131101, 131102, 131103, 131104],
	RaceUndead: [131105, 131106, 131107, 131108],
	RaceTauren: [131109, 131110, 131111, 131112],
	RaceTroll: [456006, 456007, 456008, 456009],
};
// How dark the vignette over the backdrop is, by race (PAPERDOLL_BACKGROUND_OVERLAY_ALPHA).
const overlayAlpha: Record<string, number> = { RaceHuman: 1, RaceDwarf: 0.93, RaceNightElf: 0.5, RaceTroll: 0.85, RaceOrc: 0.75, RaceUndead: 0.2 };
// The class colors (RAID_CLASS_COLORS) for the level line.
const classColors: Record<string, string> = {
	ClassWarrior: '#c69b6d',
	ClassPaladin: '#f48cba',
	ClassHunter: '#aad372',
	ClassRogue: '#fff468',
	ClassPriest: '#ffffff',
	ClassShaman: '#0070dd',
	ClassMage: '#3fc7eb',
	ClassWarlock: '#8788ee',
	ClassDruid: '#ff7c0a',
};

// The mode tabs top to bottom: the game's tab number (0 Character, 1 Skills, 2 Reputation), the
// tooltip, the frame title, and the icon (the Character tab shows your portrait).
const modeTabs: { tab: number; tooltip: string; title?: string; icon?: number }[] = [
	{ tab: 0, tooltip: 'Character' },
	{ tab: 2, tooltip: 'Reputation', title: 'Reputation', icon: 8197103 }, // INV_SideTab_Reputation2_c60
	{ tab: 1, tooltip: 'Skills', title: 'Skills', icon: 2447783 }, // Ability_Racial_JackofAllTrades
];

type SlotDef = { name: string; item?: number; texture: number };

// The paper doll's slots: eight down each side and the weapons at the bottom. Item is the sheet's
// ItemSlot index. The sim has no shirt or tabard.
const leftSlots: SlotDef[] = [
	{ name: 'Head', item: 0, texture: TX.slotHead },
	{ name: 'Neck', item: 1, texture: TX.slotNeck },
	{ name: 'Shoulders', item: 2, texture: TX.slotShoulder },
	{ name: 'Back', item: 3, texture: TX.slotBack },
	{ name: 'Chest', item: 4, texture: TX.slotChest },
	{ name: 'Shirt', texture: TX.slotShirt },
	{ name: 'Tabard', texture: TX.slotTabard },
	{ name: 'Wrist', item: 5, texture: TX.slotWrists },
];
const rightSlots: SlotDef[] = [
	{ name: 'Hands', item: 6, texture: TX.slotHands },
	{ name: 'Waist', item: 7, texture: TX.slotWaist },
	{ name: 'Legs', item: 8, texture: TX.slotLegs },
	{ name: 'Feet', item: 9, texture: TX.slotFeet },
	{ name: 'Finger', item: 10, texture: TX.slotFinger },
	{ name: 'Finger', item: 11, texture: TX.slotFinger },
	{ name: 'Trinket', item: 12, texture: TX.slotTrinket },
	{ name: 'Trinket', item: 13, texture: TX.slotTrinket },
];

// The classes whose ranged slot holds a relic (a totem, libram or idol).
const relicClasses = new Set(['ClassShaman', 'ClassPaladin', 'ClassDruid']);

// The class icons in targetingframe/ui-classes-circles (CLASS_ICON_TCOORDS), the Character tab's
// portrait (SetPortraitToClassIcon).
const classCircle: Record<string, [number, number, number, number]> = {
	ClassWarrior: [0, 0.25, 0, 0.25],
	ClassMage: [0.25, 0.49609375, 0, 0.25],
	ClassRogue: [0.49609375, 0.7421875, 0, 0.25],
	ClassDruid: [0.7421875, 0.98828125, 0, 0.25],
	ClassHunter: [0, 0.25, 0.25, 0.5],
	ClassShaman: [0.25, 0.49609375, 0.25, 0.5],
	ClassPriest: [0.49609375, 0.7421875, 0.25, 0.5],
	ClassWarlock: [0.7421875, 0.98828125, 0.25, 0.5],
	ClassPaladin: [0, 0.25, 0.5, 0.75],
};

// The Character Stats pane's sidebar tabs: Character Stats, Equipment Manager and Titles, with
// their icons in paperdollinfoframe/paperdollsidebartabs. Only the first has data here.
const sidebarTabsFile = 514608;
const sidebarTabs: { name: string; coords?: [number, number, number, number]; off?: string }[] = [
	{ name: 'Character Stats' },
	{ name: 'Equipment Manager', coords: [0.015625, 0.53125, 0.46875, 0.60546875], off: 'Not in the practice room' },
	{ name: 'Titles', coords: [0.015625, 0.53125, 0.32421875, 0.4609375], off: "You don't have any titles." },
];

const toggleTextures = { prevUp: 130869, prevDown: 130868, nextUp: 130866, nextDown: 130865, hilight: 130757 };

function stored(key: string): string | null {
	try {
		return localStorage.getItem(key);
	} catch {
		return null;
	}
}

function store(key: string, value: string) {
	try {
		localStorage.setItem(key, value);
	} catch {
		// Only this visit keeps it.
	}
}

export class CharacterFrame extends Panel {
	width = fullWidth;
	readonly height = 484;
	stackWidth = fullWidth + sideTabsPadding;
	openSound = SND.characterOpen;
	closeSound = SND.characterClose;
	// The open tab: 0 Character, 1 Skills, 2 Reputation (TOGGLECHARACTER0, 1 and 2).
	tab = 0;
	private collapsed = false;
	private root?: HTMLElement;
	private chrome?: PortraitFrame;
	private pages: HTMLElement[] = [];
	private tabs: SideTab[] = [];
	private stone?: HTMLElement;
	private paintToggle?: () => void;
	private slots: { button: ItemButton; def: SlotDef; id?: number }[] = [];
	private stats?: ScrollBox;
	private levelText?: HTMLDivElement;
	private sideIcon?: HTMLElement;
	private model?: ModelView;
	private skills?: SkillsPage;
	private filled = false;

	constructor(readonly link: RoomLink) {
		super();
		this.setCollapsed(stored(collapsedKey) === '1', false);
	}

	protected build(root: HTMLElement) {
		this.root = root;
		root.classList.add('bz-char');
		root.classList.toggle('collapsed', this.collapsed);
		// We ask for the art in one batch up front and draw while it comes in.
		void preloadAtlases([
			'UI-Character-Info-General-BG',
			'UI-Character-Info-Stat-BG',
			'UI-Character-Info-Title',
			'UI-Character-Info-Line-Bounce2',
			'common-insideframe',
			'common-sidetab',
			'UI-Character-Info-GearSlot',
			...statAtlases,
		]);
		this.chrome = portraitFrame(root, this.link.playerName(), () => this.close());
		// The panes go under the portrait and the metal border, which PortraitFrameBaseTemplate draws
		// on top.
		const body = div(null, 'bz-char-body');
		root.insertBefore(body, this.chrome.portrait);
		atlasTexture(body, 'UI-Character-Info-General-BG', { x: 0, y: 20, w: 398, h: 464 });
		const right = div(body, 'bz-right-pane', { x: 398, y: 20, w: 233, h: 464 });
		atlasTexture(right, 'UI-Character-Info-Stat-BG', { x: 0, y: 0, w: 233, h: 464 });
		this.stone = atlasTexture(right, 'UI-Character-Info-Stat-StoneBG3', { x: 0, y: 0, w: 233, h: 60 });

		const paper = div(body, 'bz-char-page');
		const skills = div(body, 'bz-char-page');
		const rep = div(body, 'bz-char-page');
		this.pages = [paper, skills, rep];
		this.buildPaperDoll(paper);
		this.skills = new SkillsPage(skills, this.link, () => this.setCollapsed(false));
		this.buildReputation(rep);

		// The line between the panes, over the border, and the button that folds the right pane.
		const divider = div(root, 'bz-right-pane bz-char-divider', { x: 392, y: 21, w: 11, h: 463 });
		void sliced(divider, 'common-framedivider', [0, 12, 0, 12], true);
		this.buildToggle(root);
		this.buildModeTabs(root);
		this.showPage();

		void this.link.whenReady().then(() => {
			this.fill();
			if (this.isOpen) this.shown();
		});
	}

	// buildToggle is the RightPaneToggleButton: a page arrow at the left pane's top right corner.
	private buildToggle(root: HTMLElement) {
		const b = div(root, 'bz-char-toggle', { x: 364, y: 26, w: 28, h: 28 });
		const up = div(b, 'bz-char-toggle-up');
		const down = div(b, 'bz-char-toggle-down');
		const hi = div(b, 'bz-char-toggle-hi');
		void glow(toggleTextures.hilight).then(u => (hi.style.backgroundImage = `url(${u})`));
		const paint = () => {
			setTexture(up, this.collapsed ? toggleTextures.nextUp : toggleTextures.prevUp);
			setTexture(down, this.collapsed ? toggleTextures.nextDown : toggleTextures.prevDown);
		};
		paint();
		this.paintToggle = paint;
		this.link.tooltip.attach(b, () => ({ kind: 'lines', title: this.collapsed ? 'Show Details' : 'Hide Details' }));
		clickable(b, () => this.setCollapsed(!this.collapsed));
	}

	// buildModeTabs hangs the mode tabs on the frame's right side, 30 down, 2 apart.
	private buildModeTabs(root: HTMLElement) {
		const host = div(root, 'bz-char-modetabs', { y: 30 });
		this.tabs = modeTabs.map((m, i) => {
			const t = sideTab(
				host,
				{ x: 0, y: i * 57 },
				m.tooltip,
				(el, title) => this.link.tooltip.attach(el, () => ({ kind: 'lines', title })),
				() => {
					sound(SND.characterTab);
					if (this.tab !== m.tab) {
						this.tab = m.tab;
						this.showPage();
					}
				},
			);
			if (m.icon) {
				t.icon.classList.add('bz-tex');
				setTexture(t.icon, m.icon, { coords: [0.03125, 0.96875, 0.03125, 0.96875], size: [64, 64] });
			} else this.sideIcon = t.icon;
			return t;
		});
	}

	private buildPaperDoll(page: HTMLElement) {
		const { loaded } = this.link;
		// The rock strip over the model, with the level line in it.
		texture(page, backgroundRock, { x: 0, y: 20, w: 398, h: 36 }, { cls: 'bz-char-rock' });
		atlasTexture(page, '_UI-Frame-InnerTopTile', { x: 0, y: 53, w: 398, h: 3 });
		this.levelText = text(page, '', 'GameFontNormal', { x: 89, y: 28, w: 220, h: 20 }, 'center');
		this.levelText.classList.add('GameFontNormalMed2Outline');

		// The model scene: the race backdrop in four pieces, a vignette over it, and the model.
		const race = raceBackgrounds[loaded.race];
		if (race) {
			const [tl, tr, bl, br] = race;
			texture(page, tl, { x: 0, y: 56, w: 319, h: 299 }, { coords: [0.171875, 1, 0.0392156862745098, 1] });
			texture(page, tr, { x: 318, y: 56, w: 80, h: 299 }, { coords: [0, 0.296875, 0.0392156862745098, 1] });
			texture(page, bl, { x: 0, y: 354, w: 319, h: 130 }, { coords: [0.171875, 1, 0, 0.62] });
			texture(page, br, { x: 318, y: 354, w: 80, h: 130 }, { coords: [0, 0.296875, 0, 0.62] });
		}
		const overlay = atlasTexture(page, 'UI-Character-Info-RaceBG-Overlay', { x: 0, y: 56, w: 398, h: 428 });
		overlay.style.opacity = String(overlayAlpha[loaded.race] ?? 0.85);
		// The model stands between the slots with its feet over the weapons.
		this.model = new ModelView(page, { x: 0, y: 56, w: 398, h: 428 }, this.link, { fraction: 0.74, lift: 0.04 });

		const ranged: SlotDef = relicClasses.has(loaded.class) ? { name: 'Relic', item: 16, texture: TX.slotRelic } : { name: 'Ranged', item: 16, texture: TX.slotRanged };
		const bottom: SlotDef[] = [{ name: 'Main Hand', item: 14, texture: TX.slotMainHand }, { name: 'Off Hand', item: 15, texture: TX.slotSecondaryHand }, ranged];
		const place = (defs: SlotDef[], x0: number, y0: number, dx: number, dy: number) =>
			defs.forEach((def, i) => {
				const x = x0 + dx * i;
				const y = y0 + dy * i;
				// The bronze slot frame sits behind the button, centered on it.
				atlasTexture(page, 'UI-Character-Info-GearSlot', { x: x - 9, y: y - 9, w: 55, h: 55 });
				const button = itemButton(page, { x, y }, def.texture, true);
				button.el.classList.add('bz-char-slot');
				const slot = { button, def, id: undefined as number | undefined };
				this.slots.push(slot);
				this.link.tooltip.attach(button.el, () => (slot.id ? { kind: 'item', id: slot.id } : { kind: 'lines', title: def.name }));
				// The doll is read only here, but a press must not reach the world behind.
				button.el.addEventListener('mousedown', e => e.stopPropagation());
			});
		// Head down the left at 24, 60 into the left pane, Hands down the right at 20 from its right
		// edge, each 6 under the last. The weapons sit 30 over the bottom, the main hand 40 left of
		// the middle (no ammo slot), each 6 right of the last.
		place(leftSlots, 24, 80, 0, 43);
		place(rightSlots, 398 - 20 - 37, 80, 0, 43);
		place(bottom, 199 - 40 - 18.5, 484 - 30 - 37, 43, 0);

		this.buildSidebarTabs(page);
		this.buildStatsPane(page);
	}

	// buildSidebarTabs are the three square tabs on the stone strip over the stats.
	private buildSidebarTabs(page: HTMLElement) {
		sidebarTabs.forEach((t, i) => {
			const el = div(page, 'bz-right-pane bz-sidebar-tab', { x: 398 + 116.5 - 21 + (i - 1) * 42, y: 29, w: 42, h: 42 });
			const icon = div(el, 'bz-sidebar-icon');
			if (t.coords) setTexture(icon, sidebarTabsFile, { coords: t.coords, size: [64, 256] });
			else icon.classList.add('bz-sidebar-portrait');
			atlasTexture(el, 'UI-Character-Info-StatTab', { x: 0, y: 0, w: 42, h: 42 });
			if (i === 0) atlasTexture(el, 'UI-Character-Info-StatTab-Selected', { x: 0, y: 0, w: 42, h: 42 });
			el.classList.toggle('disabled', !!t.off);
			this.link.tooltip.attach(el, () => ({ kind: 'lines', title: t.name, lines: t.off ? [t.off] : undefined }));
			el.addEventListener('mousedown', e => e.stopPropagation());
		});
	}

	// buildStatsPane is the CharacterStatsPaneScrollBox under the stone strip: the class art, the
	// inner frame, and the scrolling stat list with its line under it.
	private buildStatsPane(page: HTMLElement) {
		const cls = this.link.loaded.class.replace('Class', '').toUpperCase();
		atlasTexture(page, `UI-Character-Info-${cls}-BG`, { x: 398, y: 75, w: 233, h: 383 }, { cls: 'bz-right-pane' });
		const border = div(page, 'bz-right-pane bz-stats-border', { x: 399, y: 79, w: 228, h: 403 });
		void sliced(border, 'common-insideframe', [53, 53, 53, 53]);
		this.stats = scrollBox(page, { x: 408, y: 88, w: 193, h: 366 }, { x: 606, y: 100, h: 354 }, { fade: 45, step: 46 });
		this.stats.el.classList.add('bz-right-pane');
		this.stats.el.nextElementSibling?.classList.add('bz-right-pane');
		atlasTexture(page, 'UI-Character-Info-ScrollLine', { x: 408 + 96.5 - 105, y: 454, w: 210, h: 7 }, { cls: 'bz-right-pane' });
	}

	private buildReputation(page: HTMLElement) {
		scrollBox(page, { x: 10, y: 60, w: 363, h: 409 }, { x: 378, y: 62, h: 403 }).update();
		atlasTexture(page, 'UI-Character-Info-ScrollLine-Long', { x: 191.5 - 192, y: 56, w: 384, h: 8 });
		atlasTexture(page, 'UI-Character-Info-ScrollLine-Long', { x: 191.5 - 192, y: 465, w: 384, h: 8 });
		const pane = new SidePane(page, { x: 414, y: 34, w: 205, h: 436 });
		pane.el.classList.add('bz-right-pane');
		pane.setEmpty('Select a faction to view its details.');
	}

	// fill puts the sheet's items and stats in, once the session has started.
	private fill() {
		const s = this.link.sheet();
		if (!s) return;
		this.filled = true;
		const { loaded } = this.link;
		const lt = this.levelText!;
		lt.textContent = `Level ${s.level} `;
		const c = document.createElement('span');
		c.style.color = classColors[loaded.class] ?? '#fff';
		c.textContent = loaded.class.replace('Class', '');
		lt.append(c);
		const worn = new Map((s.items ?? []).map(it => [it.slot, it]));
		for (const slot of this.slots) {
			const it = slot.def.item !== undefined ? worn.get(slot.def.item) : undefined;
			slot.id = it?.id;
			slot.button.set(it ? { kind: 'item', id: it.id, quality: it.quality } : undefined);
		}
		this.fillStats();
		this.skills?.refresh();
	}

	private fillStats() {
		const s = this.link.sheet();
		if (!s || !this.stats) return;
		const c = this.stats.content;
		c.textContent = '';
		for (const cat of statCategories(s, { cls: this.link.loaded.class, level: s.level })) {
			const h = div(c, 'bz-stat-cat');
			atlasTexture(h, 'UI-Character-Info-Title', { x: 0, y: 0, w: 193, h: 40 });
			text(h, cat.name, 'GameFontHighlight', undefined, 'center').classList.add('bz-stat-cat-title');
			cat.rows.forEach((r, i) => this.statRow(c, r, i));
		}
		div(c, 'bz-stat-pad');
		this.stats.update();
	}

	// statRow is a CharacterStatFrameScrollBoxLabelElementTemplate (or the Icon one for the
	// resistances): a gold label, a white value, and the band on every other row.
	private statRow(parent: HTMLElement, r: StatRow, i: number) {
		const row = div(parent, 'bz-stat-row');
		if (i % 2 === 0) atlasTexture(row, 'UI-Character-Info-Line-Bounce2', { x: -10, y: 0, w: 213, h: 23 });
		let x = 11;
		if (r.icon) {
			const icon = div(row, 'bz-stat-icon', { x: 5, y: -5 });
			void setAtlas(icon, r.icon, { useAtlasSize: true });
			x = 5 + 33 + 5;
		}
		text(row, `${r.label}:`, 'GameFontNormal', { x }).classList.add('bz-stat-label');
		const v = text(row, r.value, 'GameFontHighlight');
		v.classList.add('bz-stat-value');
		if (r.color) v.classList.add(r.color === 'up' ? 'bz-green' : 'bz-red');
		this.link.tooltip.attach(row, () => ({ kind: 'lines', title: r.tip.title, lines: r.tip.lines.map(l => l || ' ') }));
	}

	// setCollapsed folds or opens the right pane (the stats and the detail panes), like the
	// RightPaneToggleButton, and keeps the choice for later visits (the characterFrameCollapsed
	// CVar).
	private setCollapsed(on: boolean, save = true) {
		if (on === this.collapsed && save) return;
		this.collapsed = on;
		this.width = on ? collapsedWidth : fullWidth;
		this.stackWidth = this.width + sideTabsPadding;
		if (!save) return;
		store(collapsedKey, on ? '1' : '0');
		sound(on ? SND.characterClose : SND.characterOpen);
		if (this.el) this.el.style.width = `${this.width}px`;
		this.root?.classList.toggle('collapsed', on);
		this.paintToggle?.();
		this.manager?.layout();
		if (!on) this.stats?.update();
	}

	private showPage() {
		this.pages.forEach((p, i) => (p.style.display = i === this.tab ? '' : 'none'));
		const m = modeTabs.find(t => t.tab === this.tab);
		this.tabs.forEach((t, i) => t.select(modeTabs[i].tab === this.tab));
		if (this.chrome) {
			const t = this.chrome.title;
			t.textContent = m?.title ?? this.link.playerName();
			t.classList.toggle('GameFontHighlight', !m?.title);
		}
		if (this.stone) this.stone.style.display = this.tab === 0 ? '' : 'none';
		this.drawPortrait();
		if (this.isOpen && this.tab === 0) this.model?.show();
		if (this.tab === 0) this.stats?.update();
		if (this.tab === 1) this.skills?.update();
	}

	// drawPortrait is the frame's round portrait: your class icon on the Character tab, your face on
	// the others (UpdatePortrait).
	private drawPortrait() {
		const el = this.chrome?.portrait;
		if (!el) return;
		const circle = classCircle[this.link.loaded.class];
		const face = this.tab !== 0 || !circle;
		el.classList.toggle('bz-tex', !face);
		if (!face) {
			el.querySelector('canvas')?.remove();
			setTexture(el, TX.classCircles, { coords: circle, size: [256, 256] });
		} else if (this.isOpen) {
			el.style.backgroundImage = '';
			portrait(el, this.link);
		}
	}

	protected shown() {
		if (!this.filled) this.fill();
		this.drawPortrait();
		if (this.sideIcon) portrait(this.sideIcon, this.link);
		const tab1 = this.root?.querySelector<HTMLElement>('.bz-sidebar-portrait');
		if (tab1) portrait(tab1, this.link);
		if (this.tab === 0) {
			this.model?.show();
			this.stats?.update();
		}
		if (this.tab === 1) this.skills?.update();
	}

	// showTab opens a tab (the frame is open already or opens next).
	showTab(tab: number) {
		if (tab === this.tab) return;
		this.tab = tab;
		if (!this.pages.length) return;
		if (this.isOpen) sound(SND.characterTab);
		this.showPage();
	}
}

