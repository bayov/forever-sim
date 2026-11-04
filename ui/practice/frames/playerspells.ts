// The spells frame: Forever's PlayerSpellsFrame (Blizzard_PlayerSpells with its Camelot files).
// Forever keeps Mainline's one frame for the spellbook and the talents but hides its tab row, so
// they act like two frames, K for the spellbook and Shift+K for the talents.
//
// The spellbook opens as a single page (577 wide, a "left" panel) the first time, and as the
// whole book (1154, centered when alone) after that, unless you minimized it or other panels are
// open. The talents are 1218 wide. Both scale down to fit the screen (checkFit), like the game.

import './playerspells.css';
import { iconUrl } from '../icons';
import { setAtlas } from './atlas';
import { SND, sound, TX } from './assets';
import type { RoomLink } from './link';
import { portraitFrame, PortraitFrame } from './modern';
import { clickable, div, Panel, PanelArea, setTexture } from './panels';
import { SpellBookView } from './playerspells_book';
import { TalentsView } from './playerspells_talents';
import { classCircleCoords } from './talent_trees';

export type SpellsTab = 'spellbook' | 'talents';

// SOUNDKIT.UI_CLASS_TALENT_OPEN_WINDOW and UI_CLASS_TALENT_CLOSE_WINDOW, for both tabs.
const openWindow = 567507;
const closeWindow = 567433;

// The spellBookMinimize CVar: whether you minimized the spellbook with its button.
const minimizeKey = 'practice.spellBookMinimize';

export class PlayerSpellsFrame extends Panel {
	width = 577;
	height = 720;
	area: PanelArea = 'left';
	checkFit = true;
	openSound = openWindow;
	closeSound = closeWindow;
	tab: SpellsTab = 'spellbook';

	private minimized = false;
	private manualMinimize = readMinimize();
	// The first time the spellbook opens it opens minimized (Camelot's SetTab).
	private spellBookOpenedOnce = false;
	private minimizedOnNextShow = false;
	private pf?: PortraitFrame;
	private maxMin?: HTMLElement;
	private book?: SpellBookView;
	private talents?: TalentsView;
	private bookRoot?: HTMLElement;
	private talentsRoot?: HTMLElement;

	constructor(readonly link: RoomLink) {
		super();
	}

	protected build(root: HTMLElement) {
		root.classList.add('bz-ps');
		this.pf = portraitFrame(root, 'Spellbook', () => this.close());
		// The tabs' content sits under the metal border and the portrait, like the game's frame
		// levels (content at 100, the NineSlice at 5500).
		this.bookRoot = div(null, 'bz-ps-content');
		this.talentsRoot = div(null, 'bz-ps-content');
		root.insertBefore(this.bookRoot, this.pf.portrait);
		root.insertBefore(this.talentsRoot, this.pf.portrait);
		this.maxMin = this.maxMinButton(root);
		this.apply();
	}

	protected shown() {
		this.minimizedOnNextShow = false;
		this.apply();
		if (this.tab === 'spellbook') this.book?.shown();
	}

	// showTab opens a tab (the frame is open already or opens next).
	showTab(tab: SpellsTab) {
		if (!this.isOpen && tab === 'spellbook' && !this.spellBookOpenedOnce) {
			this.spellBookOpenedOnce = true;
			this.minimizedOnNextShow = true;
		}
		const changed = tab !== this.tab;
		this.tab = tab;
		if (changed && this.isOpen) {
			this.apply();
			this.manager?.layout();
			if (tab === 'spellbook') this.book?.shown();
		}
	}

	// refresh redraws the spellbook when the toolkit changes.
	refresh() {
		this.book?.refresh();
	}

	// relayout is the game's EvaluateAutoMinimize: the spellbook minimizes while other panels are
	// open, after you minimized it, or on its first open, and is the whole book otherwise.
	relayout(othersOpen: boolean) {
		const min = this.tab === 'spellbook' && (othersOpen || this.manualMinimize || this.minimizedOnNextShow);
		if (min !== this.minimized) {
			this.minimized = min;
			if (this.el) this.apply();
		}
		this.sizeFor();
	}

	// sizeFor sets the frame's size and area for the open tab (UpdateSize, AreaAttributeForTab).
	private sizeFor() {
		const book = this.tab === 'spellbook';
		this.width = book ? (this.minimized ? 577 : 1154) : 1218;
		this.height = book ? 720 : 708;
		this.area = book && this.minimized ? 'left' : 'centerOrLeft';
		if (this.el) {
			this.el.style.width = `${this.width}px`;
			this.el.style.height = `${this.height}px`;
		}
	}

	// apply shows the open tab's content, title and portrait, building it on its first show.
	private apply() {
		this.sizeFor();
		if (!this.pf || !this.bookRoot || !this.talentsRoot) return;
		const book = this.tab === 'spellbook';
		this.pf.title.textContent = book ? 'Spellbook' : 'Talents';
		this.bookRoot.style.display = book ? '' : 'none';
		this.talentsRoot.style.display = book ? 'none' : '';
		if (this.maxMin) {
			this.maxMin.style.display = book ? '' : 'none';
			this.maxMin.classList.toggle('minimized', this.minimized);
		}
		const p = this.pf.portrait;
		p.style.backgroundImage = '';
		if (book) {
			// The General skill line's icon (SetSpellBookPortrait).
			p.style.backgroundImage = `url(${iconUrl('inv_misc_book_09')})`;
			p.style.backgroundSize = '100% 100%';
			this.book ??= new SpellBookView(this.link, this.bookRoot);
			this.book.setMinimized(this.minimized);
		} else {
			// The class icon (SetTalentPortrait).
			const circle = classCircleCoords[this.link.loaded.class];
			if (circle) setTexture(p, TX.classCircles, { coords: circle, size: [256, 256] });
			this.talents ??= new TalentsView(this.link, this.talentsRoot);
		}
	}

	// maxMinButton is MaximizeMinimizeButtonFrameTemplate (28 by 28 in Camelot), left of the
	// close button.
	private maxMinButton(root: HTMLElement): HTMLElement {
		const e = div(root, 'bz-ps-maxmin', { right: 28, y: -3, w: 28, h: 28 });
		const max = div(e, 'bz-ps-max');
		const min = div(e, 'bz-ps-min');
		for (const [el, up, down] of [
			[max, 'RedButton-Expand', 'RedButton-Expand-Pressed'],
			[min, 'RedButton-Condense', 'RedButton-Condense-Pressed'],
		] as const) {
			void setAtlas(div(el, 'bz-ps-mm-up'), up);
			void setAtlas(div(el, 'bz-ps-mm-down'), down);
			void setAtlas(div(el, 'bz-ps-mm-hi'), 'RedButton-Highlight');
		}
		clickable(max, () => this.setManualMinimize(false));
		clickable(min, () => this.setManualMinimize(true));
		return e;
	}

	private setManualMinimize(on: boolean) {
		sound(SND.checkboxOn);
		this.manualMinimize = on;
		writeMinimize(on);
		this.manager?.layout();
	}
}

function readMinimize(): boolean {
	try {
		return localStorage.getItem(minimizeKey) === '1';
	} catch {
		return false;
	}
}

function writeMinimize(on: boolean) {
	try {
		localStorage.setItem(minimizeKey, on ? '1' : '0');
	} catch {
		// The setting just doesn't last past this visit.
	}
}
