// The Game Menu (Esc) and the Settings panel, as Forever's Mainline style UI draws them. From the
// Gethe/wow-ui-source `forever` branch: Blizzard_GameMenu (GameMenuFrame on MainMenuFrameTemplate,
// with the DialogBorderTemplate and DialogHeaderTemplate) and Blizzard_Settings_Shared
// (SettingsPanel, its category list and settings list, and the setting controls), with the
// Camelot art from the client's atlases.
//
// The Settings hold the room's settings: graphics, audio and the encounter. Graphics and audio
// apply at once. The encounter waits for Apply or Close, then restarts the fight.

import './gamemenu.css';
import { defaultSoundLevels, setSoundLevels, soundLevels, SoundLevels } from '../glue';
import type { Settings } from '../protocol';
import { Graphics, graphicsPresets } from '../scene';
import type { Tooltip } from '../tooltip';
import { atlasInfo, atlasTexture, atlasUrl, NineSliceLayout, nineSlice, setAtlas } from './atlas';
import { sound, tex } from './assets';
import { closeButton } from './modern';
import { Box, button, clickable, div, Panel, text } from './panels';

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

// The sound files of the kits the frames play (SoundKitConstants, files from the client).
const SOUND = {
	// IG_MAINMENU_OPEN (850).
	menuOpen: 567490,
	// IG_MAINMENU_OPTION (852). IG_MAINMENU_QUIT, _CONTINUE and _LOGOUT play the same file.
	menuOption: 567481,
	// IG_MAINMENU_OPTION_CHECKBOX_ON and _OFF (856, 857): checkboxes, categories, steppers, menus.
	checkbox: 567407,
	// IG_CHARACTER_INFO_TAB (841): the Game and AddOns tabs.
	tab: 567422,
};

// A FontString in one of the fonts frames.css doesn't have (Blizzard_Fonts_Shared).
type GmFont = 'HighlightLarge' | 'NormalMed1' | 'HighlightHuge' | 'HighlightMedium' | 'White';
function label(parent: HTMLElement, s: string, font: GmFont, box?: Box): HTMLDivElement {
	return div(parent, `bz-text gm-f-${font}`, box, s);
}

// fillAtlas is a texture with an atlas stretched over it, even one that tiles in the game, for
// art that is the same all along (button middles).
function fillAtlas(parent: HTMLElement, name: string, box: Box): HTMLDivElement {
	const e = div(parent, 'bz-tex', box);
	void setAtlas(e, name).then(a => {
		if (!a) return;
		e.style.backgroundRepeat = 'no-repeat';
		e.style.backgroundSize = '100% 100%';
	});
	return e;
}

// ---- Game Menu ----

// NineSliceLayouts.Dialog: the diamond metal border of dialogs.
const dialogLayout: NineSliceLayout = {
	TopLeftCorner: { atlas: 'UI-Frame-DiamondMetal-CornerTopLeft' },
	TopRightCorner: { atlas: 'UI-Frame-DiamondMetal-CornerTopRight' },
	BottomLeftCorner: { atlas: 'UI-Frame-DiamondMetal-CornerBottomLeft' },
	BottomRightCorner: { atlas: 'UI-Frame-DiamondMetal-CornerBottomRight' },
	TopEdge: { atlas: '_UI-Frame-DiamondMetal-EdgeTop' },
	BottomEdge: { atlas: '_UI-Frame-DiamondMetal-EdgeBottom' },
	LeftEdge: { atlas: '!UI-Frame-DiamondMetal-EdgeLeft' },
	RightEdge: { atlas: '!UI-Frame-DiamondMetal-EdgeRight' },
};

// Interface\DialogFrame\UI-DialogBox-Background, tiled inside the dialog border.
const dialogBackground = 131071;

// dialogBorder is DialogBorderTemplate: the tiled dark background and the Dialog NineSlice.
function dialogBorder(root: HTMLElement) {
	const bg = div(root, 'bz-tex gm-dialog-bg', { x: 7, y: 7, right: 7, bottom: 7 });
	void tex(dialogBackground).then(u => (bg.style.backgroundImage = `url(${u})`));
	nineSlice(root, dialogLayout);
}

// dialogHeader is DialogHeaderTemplate: the metal plate on the top edge, as wide as its text plus
// 64, with the text 13 down.
function dialogHeader(root: HTMLElement, title: string) {
	const h = div(root, 'gm-header');
	atlasTexture(h, 'UI-Frame-DiamondMetal-Header-CornerLeft', { x: 0, y: 0, w: 32, h: 39 });
	atlasTexture(h, 'UI-Frame-DiamondMetal-Header-CornerRight', { right: 0, y: 0, w: 32, h: 39 });
	fillAtlas(h, '_UI-Frame-DiamondMetal-Header-Tile', { x: 32, right: 32, y: 0, h: 39 });
	label(h, title, 'NormalMed1').classList.add('gm-header-text');
}

// redButton is BigRedThreeSliceButtonTemplate (MainMenuFrameButtonTemplate): the 128-RedButton
// art scaled to the button's height, with the red glow on hover.
function redButton(parent: HTMLElement, text: string, box: Box & { w: number; h: number }, onClick: () => void): HTMLDivElement {
	const e = div(parent, 'gm-red', box);
	const s = box.h / 128;
	const lw = 114 * s;
	const rw = 292 * s;
	for (const [kind, suffix] of [
		['n', ''],
		['p', '-Pressed'],
	]) {
		const set = div(e, `gm-red-${kind}`);
		// The middle reaches half a unit under each end, so no seam shows at fractional sizes.
		fillAtlas(set, `_128-RedButton-Center${suffix}`, { x: lw - 0.5, right: rw - 0.5, y: 0, h: box.h });
		atlasTexture(set, `128-RedButton-Left${suffix}`, { x: 0, y: 0, w: lw, h: box.h });
		atlasTexture(set, `128-RedButton-Right${suffix}`, { right: 0, y: 0, w: rw, h: box.h });
	}
	atlasTexture(e, '128-RedButton-Highlight', { x: 0, y: 0, w: box.w, h: box.h }, { cls: 'gm-red-hi' });
	label(e, text, 'HighlightLarge').classList.add('gm-red-text');
	clickable(e, onClick);
	return e;
}

// The menu's buttons in sections, with a gap between sections like the game's.
type MenuItem = { label: string; run: () => void };

// GameMenu is GameMenuFrame: a dialog with "Game Menu" on its header and the big red buttons in
// a column (MainMenuFrameTemplate: 48 above, 34 below, 28 at the sides, 20 between sections).
export class GameMenu extends Panel {
	readonly width = 256;
	height = 0;
	area = 'dialog' as const;
	openSound = SOUND.menuOpen;
	// Esc plays IG_MAINMENU_QUIT.
	closeSound = SOUND.menuOption;

	constructor(private readonly groups: MenuItem[][]) {
		super();
		const buttons = groups.reduce((n, g) => n + g.length, 0);
		this.height = 48 + buttons * 36 + (groups.length - 1) * 20 + 34;
	}

	protected build(root: HTMLElement) {
		root.classList.add('gm-menu');
		dialogBorder(root);
		dialogHeader(root, 'Game Menu');
		let y = 48;
		this.groups.forEach((group, i) => {
			if (i) y += 20;
			for (const item of group) {
				redButton(root, item.label, { x: 28, y, w: 200, h: 36 }, () => {
					sound(SOUND.menuOption);
					item.run();
				});
				y += 36;
			}
		});
	}
}

// ---- Settings controls ----

// NineSliceLayouts.ButtonFrameTemplateNoPortrait with the Camelot offsets (the right corners 2
// to the left, the bottom ones at y -8).
const settingsLayout: NineSliceLayout = {
	TopLeftCorner: { atlas: 'UI-Frame-Metal-CornerTopLeft', x: -8, y: 16 },
	TopRightCorner: { atlas: 'UI-Frame-Metal-CornerTopRight', x: 2, y: 16 },
	BottomLeftCorner: { atlas: 'UI-Frame-Metal-CornerBottomLeft', x: -8, y: -8 },
	BottomRightCorner: { atlas: 'UI-Frame-Metal-CornerBottomRight', x: 2, y: -8 },
	TopEdge: { atlas: '_UI-Frame-Metal-EdgeTop' },
	BottomEdge: { atlas: '_UI-Frame-Metal-EdgeBottom' },
	LeftEdge: { atlas: '!UI-Frame-Metal-EdgeLeft' },
	RightEdge: { atlas: '!UI-Frame-Metal-EdgeRight' },
};

// A control on a settings row. refresh reads its value again, after any change on the page.
type Control = { refresh(): void };

// The rows of a settings page. A child row is indented 15 in a smaller font, under the row it
// depends on (SetParentInitializer).
type Row =
	| { kind: 'header'; title: string }
	| { kind: 'check'; label: string; tip?: string; child?: boolean; get(): boolean; set(v: boolean): void }
	| {
			kind: 'slider';
			label: string;
			tip?: string;
			child?: boolean;
			min: number;
			max: number;
			step: number;
			format(v: number): string;
			get(): number;
			set(v: number): void;
	  }
	| { kind: 'dropdown'; label: string; tip?: string; child?: boolean; options(): [string, string][]; get(): string; set(v: string): void };

// checkboxControl is SettingsCheckboxTemplate: the minimal box with the yellow check.
function checkboxControl(row: HTMLElement, x: number, get: () => boolean, set: (v: boolean) => void): Control {
	const e = div(row, 'gm-check', { x, y: -1.5, w: 30, h: 29 });
	atlasTexture(e, 'checkbox-minimal', { x: 0, y: 0, w: 30, h: 29 });
	const mark = atlasTexture(e, 'checkmark-minimal', { x: 0, y: 0, w: 30, h: 29 });
	const refresh = () => (mark.style.display = get() ? '' : 'none');
	clickable(e, () => {
		sound(SOUND.checkbox);
		set(!get());
	});
	refresh();
	return { refresh };
}

// stepper is a slider's arrow: dimmed at its end (STEPPER_DISABLED_ALPHA).
function stepper(parent: HTMLElement, atlas: string, box: Box & { w: number; h: number }, onClick: () => void): HTMLDivElement {
	const e = div(parent, 'gm-step', box);
	atlasTexture(e, atlas, { x: 0, y: 0, w: box.w, h: box.h });
	clickable(e, () => {
		if (e.classList.contains('disabled')) return;
		sound(SOUND.checkbox);
		onClick();
	});
	return e;
}

// sliderControl is MinimalSliderWithSteppersTemplate (250 by 40): the thin bar with its thumb,
// an arrow at each end and the value on the right in yellow.
function sliderControl(row: HTMLElement, x: number, o: Extract<Row, { kind: 'slider' }>): Control {
	const e = div(row, 'gm-slider', { x, y: -10, w: 250, h: 40 });
	const clamp = (v: number) => Math.max(o.min, Math.min(o.max, +(Math.round(v / o.step) * o.step).toFixed(6)));
	const change = (v: number) => {
		const next = clamp(v);
		if (next !== o.get()) o.set(next);
	};
	const back = stepper(e, 'Minimal_SliderBar_Button_Left', { x: 4, y: 10.5, w: 11, h: 19 }, () => change(o.get() - o.step));
	const forward = stepper(e, 'Minimal_SliderBar_Button_Right', { x: 235, y: 11, w: 9, h: 18 }, () => change(o.get() + o.step));
	const track = div(e, 'gm-slider-track', { x: 19, y: 0, w: 212, h: 40 });
	atlasTexture(track, 'Minimal_SliderBar_Left', { x: 0, y: 11.5, w: 11, h: 17 });
	atlasTexture(track, 'Minimal_SliderBar_Right', { x: 201, y: 11.5, w: 11, h: 17 });
	fillAtlas(track, '_Minimal_SliderBar_Middle', { x: 11, y: 11.5, w: 190, h: 17 });
	const thumb = atlasTexture(track, 'Minimal_SliderBar_Button', { y: 10.5, w: 20, h: 19 });
	const value = text(e, '', 'GameFontNormal', { x: 256, y: 13 });
	const refresh = () => {
		const v = o.get();
		thumb.style.left = `${((v - o.min) / (o.max - o.min || 1)) * 192}px`;
		value.textContent = o.format(v);
		back.classList.toggle('disabled', v <= o.min);
		forward.classList.toggle('disabled', v >= o.max);
	};
	// The thumb's center follows the mouse along the bar, in steps.
	const drag = (clientX: number) => {
		const r = track.getBoundingClientRect();
		const zoom = r.width / 212;
		const f = (clientX - r.left - 10 * zoom) / (192 * zoom);
		change(o.min + Math.max(0, Math.min(1, f)) * (o.max - o.min));
	};
	track.addEventListener('mousedown', ev => {
		if (ev.button !== 0) return;
		ev.stopPropagation();
		drag(ev.clientX);
		const move = (m: MouseEvent) => drag(m.clientX);
		const up = () => {
			removeEventListener('mousemove', move);
			removeEventListener('mouseup', up);
		};
		addEventListener('mousemove', move);
		addEventListener('mouseup', up);
	});
	refresh();
	return { refresh };
}

// The open dropdown list, one at a time.
let openMenu: { close(): void } | undefined;

function closeOpenMenu(): boolean {
	if (!openMenu) return false;
	openMenu.close();
	return true;
}

// dropdownControl is SettingsDropdownWithButtonsTemplate: a WowStyle2Dropdown (220 wide) with a
// back and a next arrow on its sides that step through the choices. A click opens the list
// (MenuStyle2, its choices as highlight radios). A value no choice has reads "Custom", like the
// game's SetDefaultText(CUSTOM).
function dropdownControl(row: HTMLElement, x: number, o: Extract<Row, { kind: 'dropdown' }>): Control {
	const control = div(row, 'gm-dd-control', { x, y: -9, w: 223, h: 38 });
	const dd = div(control, 'gm-dd', { x: 1.5, y: 6.5, w: 220, h: 25 });
	const bg = div(dd, 'bz-tex gm-dd-bg');
	const arrow = atlasTexture(dd, 'common-dropdown-c-button-hover-arrow', { y: 25, w: 12, h: 5 }, { cls: 'gm-dd-arrow' });
	arrow.style.left = '104px';
	const t = text(dd, '', 'GameFontNormal', undefined, 'center');
	t.classList.add('gm-dd-text');

	let isOpen = false;
	// WowStyle2DropdownMixin:GetBackgroundAtlas, its border cut 12 from each side
	// (UiTextureAtlasElementSliceData).
	const paint = () => {
		const over = dd.matches(':hover, .v-hover');
		const down = dd.classList.contains('pushed');
		const name = down && over ? 'pressedhover-1' : over ? 'hover-1' : down ? 'pressed-1' : isOpen ? 'open' : '';
		void atlasUrl(name ? `common-dropdown-c-button-${name}` : 'common-dropdown-c-button').then(u => u && (bg.style.borderImageSource = `url(${u})`));
	};
	for (const ev of ['mouseenter', 'mouseleave', 'mousedown', 'mouseup']) dd.addEventListener(ev, () => requestAnimationFrame(paint));

	const index = () => o.options().findIndex(([v]) => v === o.get());
	const stepTo = (d: number) => {
		const opts = o.options();
		const i = index() + d;
		if (i >= 0 && i < opts.length) o.set(opts[i][0]);
	};
	// WowStyle2IconButtonTemplate: the arrow icon over the dropdown's button art.
	const arrowButton = (left: boolean) => {
		const b = div(control, 'gm-dd-step', { x: left ? -29.5 : 225.5, y: 6.5, w: 26, h: 25 });
		const art = atlasTexture(b, 'common-dropdown-c-button', { x: -6.5, y: -7, w: 39, h: 39 });
		const icon = div(b, 'bz-tex gm-dd-icon', { x: 4.5, y: 4, w: 17, h: 17 });
		const name = left ? 'common-dropdown-icon-back' : 'common-dropdown-icon-next';
		let state = '';
		const paintStep = () => {
			const disabled = b.classList.contains('disabled');
			const over = b.matches(':hover, .v-hover');
			const down = b.classList.contains('pushed');
			const s = disabled ? 'd' : down && over ? 'ph' : over ? 'h' : down ? 'p' : 'n';
			if (s === state) return;
			state = s;
			const bgName = { d: 'common-dropdown-c-button-disabled', ph: 'common-dropdown-c-button-pressedhover-2', h: 'common-dropdown-c-button-hover-2', p: 'common-dropdown-c-button-pressed-2', n: 'common-dropdown-c-button' }[s];
			void setAtlas(art, bgName);
			void setAtlas(icon, disabled ? `${name}-disabled` : name);
		};
		for (const ev of ['mouseenter', 'mouseleave', 'mousedown', 'mouseup']) b.addEventListener(ev, () => requestAnimationFrame(paintStep));
		clickable(b, () => {
			if (b.classList.contains('disabled')) return;
			sound(SOUND.checkbox);
			stepTo(left ? -1 : 1);
		});
		return { el: b, paint: paintStep };
	};
	const prev = arrowButton(true);
	const next = arrowButton(false);

	// The list: MenuStyle2Mixin's background (out 17 at the sides, 12 above, 22 below, cut
	// 23/18/23/28) and the choices 20 high inside its inset (3, 6, 3, 7).
	const showMenu = () => {
		closeOpenMenu();
		const opts = o.options();
		const menu = div(control, 'gm-menu-list', { x: 1.5, y: 33, w: 220, h: 6 + opts.length * 20 + 7 });
		div(menu, 'bz-tex gm-menu-bg');
		opts.forEach(([v, l], i) => {
			const item = div(menu, `gm-menu-item${v === o.get() ? ' selected' : ''}`, { x: 3, y: 6 + i * 20, w: 214, h: 20 });
			div(item, 'bz-tex gm-menu-hi');
			text(item, l, 'GameFontNormal', { x: 14, y: 3 });
			clickable(item, () => {
				sound(SOUND.checkbox);
				close();
				if (v !== o.get()) o.set(v);
			});
		});
		void Promise.all([atlasUrl('common-dropdown-c-bg'), atlasUrl('common-dropdown-customize-mouseover')]).then(([b, h]) => {
			if (b) menu.style.setProperty('--gm-menu-bg', `url(${b})`);
			if (h) menu.style.setProperty('--gm-menu-hi', `url(${h})`);
		});
		// A press anywhere else closes the list, like the game's menus.
		const outside = (ev: MouseEvent) => {
			if (!menu.contains(ev.target as Node) && !dd.contains(ev.target as Node)) close();
		};
		const close = () => {
			if (openMenu !== handle) return;
			openMenu = undefined;
			isOpen = false;
			menu.remove();
			removeEventListener('mousedown', outside, true);
			paint();
		};
		const handle = { close };
		addEventListener('mousedown', outside, true);
		openMenu = handle;
		isOpen = true;
		paint();
	};
	clickable(dd, () => {
		sound(SOUND.checkbox);
		if (isOpen) closeOpenMenu();
		else showMenu();
	});

	const refresh = () => {
		const opts = o.options();
		const i = index();
		t.textContent = i >= 0 ? opts[i][1] : 'Custom';
		prev.el.classList.toggle('disabled', i <= 0);
		next.el.classList.toggle('disabled', i < 0 || i >= opts.length - 1);
		prev.paint();
		next.paint();
	};
	paint();
	refresh();
	return { refresh };
}

// MinimalTabTemplate: the Game and AddOns tabs over the inner frame, sized to their text.
function minimalTab(parent: HTMLElement, text: string): HTMLDivElement {
	const e = div(parent, 'gm-tab');
	for (const [kind, suffix] of [
		['u', ''],
		['a', 'Active_'],
	]) {
		const set = div(e, `gm-tab-${kind}`);
		const h = kind === 'a' ? 26 : 23;
		atlasTexture(set, `Options_Tab_${suffix}Left`, { x: 0, bottom: 0, w: 7, h });
		atlasTexture(set, `Options_Tab_${suffix}Right`, { right: 0, bottom: 0, w: 7, h });
		fillAtlas(set, `Options_Tab_${suffix}Middle`, { x: 7, right: 7, bottom: 0, h });
	}
	div(e, 'bz-text gm-tab-text', undefined, text);
	return e;
}

// ---- Settings ----

type Category = 'Encounter' | 'Graphics' | 'Audio';

// The groups of the category list (Groups.lua: Gameplay, then System), each under its header art
// (Options_CategoryHeader_1, _2).
const groups: { title: string; categories: Category[] }[] = [
	{ title: 'Gameplay', categories: ['Encounter'] },
	{ title: 'System', categories: ['Graphics', 'Audio'] },
];

const fightDefaults = { mode: 'free', lengthMs: 180000, countdownMs: 5000, pingMs: 60, executeFromStart: false, spellQueueWindowMs: 400 } as const;
const fightKeys = ['mode', 'lengthMs', 'countdownMs', 'pingMs', 'executeFromStart', 'spellQueueWindowMs'] as const;

// The settings list's geometry (SettingsListTemplate in the Container): the rows start 62 down,
// 10 in, 635 wide, 26 high with 9 between, and their controls start at their middle less 80
// (checkbox, slider) or less 48 (dropdown).
const rowX = 10;
const rowW = 635;
const rowTop = 62;
const rowGap = 9;
const controlX = rowW / 2 - 80;
const dropdownX = rowW / 2 - 48;

// OptionsFrame is SettingsPanel (920 by 724): the metal frame titled Options, the Game and AddOns
// tabs, the category list on the left, the category's settings on the right under its title and
// Defaults, and Apply and Close at the bottom right. Esc or Close keeps the changes and goes back
// to the Game Menu, like the game.
export class OptionsFrame extends Panel {
	readonly width = 920;
	readonly height = 724;
	area = 'dialog' as const;
	private category: Category = 'Encounter';
	private list?: HTMLElement;
	private page?: HTMLElement;
	private title?: HTMLElement;
	private apply?: HTMLElement;
	private output?: HTMLElement;
	private controls: Control[] = [];
	// The encounter edits, waiting for Apply or Close.
	private fight?: Settings;
	onClosed?: () => void;

	constructor(
		private readonly host: MenuHost,
		private readonly tooltip?: Tooltip,
	) {
		super();
	}

	protected async build(root: HTMLElement) {
		root.classList.add('gm-settings');
		// SettingsFrameTemplate: the flat background (PANEL_BACKGROUND_COLOR, its bottom corners
		// rounded by the uiframebackground corner atlases) under the metal NineSlice.
		const bg = div(root, 'gm-flat-bg', { x: 7, y: 18, right: 3, bottom: 3 });
		div(bg, 'gm-flat-top');
		const corners = [div(bg, 'gm-flat-bl'), div(bg, 'gm-flat-edge'), div(bg, 'gm-flat-br')];
		void Promise.all([atlasUrl('uiframebackground-nineslice-cornerbottomleft'), atlasUrl('uiframebackground-nineslice-cornerbottomright')]).then(([l, r]) => {
			if (l) corners[0].style.maskImage = `url(${l})`;
			if (r) corners[2].style.maskImage = `url(${r})`;
		});
		atlasTexture(root, 'Options_InnerFrame', { x: 17, y: 64, w: 886, h: 618 });
		nineSlice(root, settingsLayout);
		const t = text(root, 'Options', 'GameFontNormal', { x: 60, right: 60, y: 5 }, 'center');
		t.classList.add('gm-title');
		closeButton(root, { right: 0, y: -3 }, () => this.close());

		// The tabs: the room has no addons, so the AddOns tab has nothing to show.
		const tabs = div(root, 'gm-tabs', { x: 32, y: 27, h: 37 });
		minimalTab(tabs, 'Game').classList.add('selected');
		const addons = minimalTab(tabs, 'AddOns');
		addons.classList.add('disabled');
		this.tooltip?.attach(addons, () => ({ kind: 'lines', title: 'AddOns', lines: ['No addons in the room'] }));

		this.list = div(root, 'gm-cat-list', { x: 18, y: 76, w: 199, h: 602 });
		const container = div(root, 'gm-container', { x: 233, y: 76, w: 665, h: 603 });
		this.title = label(container, '', 'HighlightHuge', { x: 7, y: 22 });
		button(container, 'Defaults', { x: 533, y: 16, w: 96, h: 22 }, () => {
			sound(SOUND.menuOption);
			this.defaults();
		});
		atlasTexture(container, 'Options_HorizontalDivider', { x: 17.5, y: 50, w: 630, h: 1 });
		this.page = div(container, 'gm-page', { x: 0, y: 0, w: 665, h: 603 });

		button(root, 'Close', { x: 808, y: 686, w: 96, h: 22 }, () => this.close());
		this.apply = button(root, 'Apply', { x: 710, y: 686, w: 96, h: 22 }, () => this.applyFight());
		// OutputText, the panel's message line at the bottom.
		this.output = div(root, 'bz-text gm-f-White gm-output', { bottom: 24 });
		await atlasInfo('Options_InnerFrame');
	}

	protected shown() {
		this.fight = { ...this.host.fight() };
		// The game opens on the first category every time.
		this.category = groups[0].categories[0];
		this.renderList();
		this.renderPage();
	}

	protected hidden() {
		closeOpenMenu();
		this.tooltip?.hide();
		this.onClosed?.();
		// Closing keeps the changes, like the game's ExitWithCommit.
		if (this.pending()) this.applyFight();
	}

	// closeMenus closes an open dropdown list (the first thing Esc does) and says whether there
	// was one.
	closeMenus(): boolean {
		return closeOpenMenu();
	}

	private renderList() {
		const list = this.list!;
		list.replaceChildren();
		let y = 0;
		groups.forEach((g, gi) => {
			if (gi) y += 18 + 2;
			// SettingsCategoryListHeaderTemplate: 175 by 30, its art at the top left at its own size.
			const header = div(list, 'gm-cat-header', { x: 0, y, w: 175, h: 30 });
			atlasTexture(header, `Options_CategoryHeader_${(gi % 3) + 1}`, { x: 0, y: 0, w: 199, h: 144 });
			label(header, g.title, 'HighlightMedium', { x: 20, y: 8 });
			y += 30 + 2;
			for (const c of g.categories) {
				// SettingsCategoryListButtonTemplate: 175 by 20, the hover or active bar (187 by 21)
				// centered behind the name.
				const selected = c === this.category;
				const b = div(list, `gm-cat${selected ? ' selected' : ''}`, { x: 0, y, w: 175, h: 20 });
				atlasTexture(b, selected ? 'Options_List_Active' : 'Options_List_Hover', { x: -6, y: -0.5, w: 187, h: 21 }, { cls: 'gm-cat-bar' });
				text(b, c, selected ? 'GameFontHighlight' : 'GameFontNormal', { x: 36, y: 3 });
				clickable(b, () => {
					if (c === this.category) return;
					sound(SOUND.checkbox);
					closeOpenMenu();
					this.category = c;
					this.renderList();
					this.renderPage();
				});
				y += 20 + 2;
			}
		});
	}

	private rows(): Row[] {
		if (this.category === 'Graphics') return this.graphicsRows();
		if (this.category === 'Audio') return this.audioRows();
		return this.encounterRows();
	}

	private renderPage() {
		const page = this.page!;
		page.replaceChildren();
		this.tooltip?.hide();
		this.title!.textContent = this.category;
		this.controls = [];
		let y = rowTop;
		for (const r of this.rows()) {
			if (r.kind === 'header') {
				// SettingsListSectionHeaderTemplate: 45 high, the title 16 down.
				label(page, r.title, 'HighlightLarge', { x: rowX + 7, y: y + 16 });
				y += 45 + rowGap;
				continue;
			}
			const row = div(page, 'gm-row', { x: rowX, y, w: rowW, h: 26 });
			const indent = r.child ? 15 : 0;
			const name = text(row, r.label, r.child ? 'GameFontNormalSmall' : 'GameFontNormal', { x: indent + 37, y: r.child ? 7 : 6, w: controlX - 5 - indent - 37 });
			name.classList.add('gm-row-label');
			if (r.tip) this.tooltip?.attach(row, () => ({ kind: 'lines', title: r.label, lines: [r.tip!] }));
			if (r.kind === 'check') this.controls.push(checkboxControl(row, controlX, r.get, r.set));
			else if (r.kind === 'slider') this.controls.push(sliderControl(row, controlX, r));
			else this.controls.push(dropdownControl(row, dropdownX, r));
			y += 26 + rowGap;
		}
		this.refresh();
	}

	// refresh brings every control and the Apply button up to date after a change.
	private refresh() {
		for (const c of this.controls) c.refresh();
		const pending = this.pending();
		this.apply!.style.display = pending ? '' : 'none';
		this.output!.textContent = pending ? 'Apply restarts the fight' : '';
	}

	private graphicsRows(): Row[] {
		const g = () => this.host.graphics();
		const set = (next: Partial<Graphics>) => {
			this.host.setGraphics({ ...g(), ...next });
			this.refresh();
		};
		// The screen's own size in pixels (the first choice), and the common heights under it.
		const nativeW = Math.round(screen.width * devicePixelRatio);
		const nativeH = Math.round(screen.height * devicePixelRatio);
		const heights = [2160, 1440, 1080, 720].filter(v => v < nativeH);
		const size = (h: number) => `${Math.round((h * nativeW) / nativeH)}x${h}`;
		return [
			{
				kind: 'dropdown',
				label: 'Resolution',
				tip: 'The most pixels the room draws',
				options: () => [['0', size(nativeH)], ...heights.map(h => [String(h), size(h)] as [string, string])],
				get: () => (heights.includes(g().maxHeight) ? String(g().maxHeight) : '0'),
				set: v => set({ maxHeight: +v }),
			},
			{
				kind: 'slider',
				label: 'Render Scale',
				tip: 'Lower draws fewer pixels, for more FPS',
				child: true,
				min: 0.5,
				max: 1,
				step: 0.25,
				format: v => `${Math.round(v * 100)}%`,
				get: () => g().renderScale,
				set: v => set({ preset: 'custom', renderScale: v }),
			},
			{ kind: 'header', title: 'Graphics Quality' },
			{
				kind: 'dropdown',
				label: 'Graphics Quality',
				tip: 'Sets shadows and render scale at once',
				options: () => [
					['low', 'Low'],
					['medium', 'Good'],
					['high', 'High'],
				],
				get: () => g().preset,
				set: v => set({ ...graphicsPresets[v as 'low' | 'medium' | 'high'], maxHeight: g().maxHeight }),
			},
			{
				kind: 'dropdown',
				label: 'Shadow Quality',
				tip: 'Lower shadows give more FPS',
				child: true,
				options: () => [
					['off', 'Disabled'],
					['low', 'Low'],
					['medium', 'Good'],
					['high', 'High'],
				],
				get: () => g().shadows,
				set: v => set({ preset: 'custom', shadows: v as Graphics['shadows'] }),
			},
		];
	}

	private audioRows(): Row[] {
		const set = (next: Partial<SoundLevels>) => {
			setSoundLevels({ ...soundLevels(), ...next });
			this.refresh();
		};
		const pct = (v: number) => `${Math.round(v * 100)}%`;
		const volume = (label: string, key: 'master' | 'effects' | 'music' | 'ambience', tip: string, child = true): Row => ({
			kind: 'slider',
			label,
			tip,
			child,
			min: 0,
			max: 1,
			step: 0.05,
			format: pct,
			get: () => soundLevels()[key],
			set: v => set({ [key]: v }),
		});
		return [
			{ kind: 'check', label: 'Enable Sound', tip: 'Turns all sound on or off', get: () => soundLevels().enabled, set: on => set({ enabled: on }) },
			volume('Master Volume', 'master', 'The volume of all sound', false),
			volume('Music', 'music', 'Plays on the setup page only'),
			volume('Effects', 'effects', 'Spells, hits and the interface'),
			volume('Ambience', 'ambience', 'The sounds of the zone around you'),
		];
	}

	private encounterRows(): Row[] {
		const f = () => this.fight!;
		const set = (next: Partial<Settings>) => {
			this.fight = { ...f(), ...next };
			this.refresh();
		};
		const ms = (v: number) => `${v}ms`;
		return [
			{
				kind: 'dropdown',
				label: 'Mode',
				tip: 'Encounter is a timed fight with a countdown',
				options: () => [
					['free', 'Free'],
					['encounter', 'Encounter'],
				],
				get: () => f().mode,
				set: v => set({ mode: v as Settings['mode'] }),
			},
			{
				kind: 'slider',
				label: 'Length',
				tip: 'How long an encounter lasts',
				min: 30,
				max: 600,
				step: 10,
				format: fmtLength,
				get: () => f().lengthMs / 1000,
				set: v => set({ lengthMs: v * 1000 }),
			},
			{
				kind: 'slider',
				label: 'Countdown',
				tip: 'The wait before an encounter starts',
				min: 0,
				max: 15,
				step: 1,
				format: v => `${v}s`,
				get: () => f().countdownMs / 1000,
				set: v => set({ countdownMs: v * 1000 }),
			},
			{ kind: 'check', label: 'Start in Execute Range', tip: 'The target starts under 20% health', get: () => f().executeFromStart, set: on => set({ executeFromStart: on }) },
			{ kind: 'header', title: 'Latency' },
			{ kind: 'slider', label: 'Ping', tip: 'The delay to the server and back', min: 0, max: 300, step: 10, format: ms, get: () => f().pingMs, set: v => set({ pingMs: v }) },
			{
				kind: 'slider',
				label: 'Spell Queue Window',
				tip: 'How early a cast can queue before the GCD ends',
				min: 0,
				max: 400,
				step: 10,
				format: ms,
				get: () => f().spellQueueWindowMs,
				set: v => set({ spellQueueWindowMs: v }),
			},
		];
	}

	// defaults puts the open category back to its defaults. The game asks first, for this
	// category or all of them.
	private defaults() {
		closeOpenMenu();
		if (this.category === 'Graphics') this.host.setGraphics({ ...graphicsPresets.high, maxHeight: this.host.graphics().maxHeight });
		else if (this.category === 'Audio') setSoundLevels(defaultSoundLevels());
		else this.fight = { ...this.fight!, ...fightDefaults };
		this.refresh();
	}

	private pending(): boolean {
		const f = this.fight;
		if (!f) return false;
		const was = this.host.fight();
		return fightKeys.some(k => f[k] !== was[k]);
	}

	private applyFight() {
		if (!this.fight || !this.pending()) return;
		sound(SOUND.menuOption);
		this.host.applyFight({ ...this.fight });
		if (this.isOpen) this.refresh();
	}
}

function fmtLength(s: number): string {
	const m = Math.floor(s / 60);
	const r = s % 60;
	return m ? (r ? `${m}m ${r}s` : `${m}m`) : `${r}s`;
}
