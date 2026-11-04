// The Spellbook tab of the spells frame: Forever's SpellBookFrame (Blizzard_PlayerSpells
// SpellBook with the Camelot templates). Your spells sit in categories by skill line, one icon
// tab each along the top, and the items you can use get a last Items tab. A category's spells
// fill the parchment pages in two columns under its header, and the book turns pages. You pick
// a spell or item up from it to drop on an action button.
//
// Minimized it is one page (574 wide), maximized two facing pages (1148). Left out: the search
// box and the settings dropdown (Hide Passives, Show All Ranks), since every spell the room
// knows is an active one and we always show all ranks.

import type { SlotContent } from '../actions';
import { iconUrl, setIcon } from '../icons';
import type { ToolkitSpell } from '../protocol';
import { tooltipInfo } from '../tooltip';
import { atlasUrl, setAtlas } from './atlas';
import { glow, sound, tex, TX } from './assets';
import type { RoomLink } from './link';
import { clickable, div, text } from './panels';

// SOUNDKIT.IG_SPELLBOOK_OPEN, the category tabs' select sound.
const tabSound = 567440;
// SOUNDKIT.IG_ABILITY_PAGE_TURN, one of its three files at random.
const pageTurn = [567457, 567472, 567502];
// The Items tab's bag (inv_misc_bag_08).
const itemsIcon = 133633;

// The paged grid (PagedSpellsFrame): views 448 by 520, two columns 15 apart, items 60 high with
// 10 under each, and the header as tall as an item row.
const viewW = 448;
const viewH = 520;
const xPad = 15;
const rowH = 60 + 10;
const colW = (viewW - xPad) / 2;
// The book (SpellBookFrame) sits 14 below the frame's top, and its views 95 below that.
const bookTop = 14;
const viewTop = bookTop + 50 + 45;

type Category = { key: string; name: string; icon: () => Promise<string> };
type Entry = { content: SlotContent; name: string; sub: string; rankNumber: number };
type Placed = { entry?: Entry; header?: string; col: number; row: number };

export class SpellBookView {
	private categories: Category[] = [];
	private entries = new Map<string, Entry[]>();
	private selected = 'general';
	// The open page of each category, from 0.
	private pages = new Map<string, number>();
	private minimized = true;
	private stale = true;
	private bg: HTMLElement[] = [];
	private tabRow: HTMLElement;
	private views: HTMLElement[] = [];
	private paging: HTMLElement;
	private pageText: HTMLElement;
	private prev: HTMLElement;
	private next: HTMLElement;
	// Names the toolkit lacks, from the tooltips, by "spell:id" or "item:id".
	private lookedUp = new Map<string, string>();
	private looking = new Set<string>();

	constructor(
		readonly link: RoomLink,
		readonly root: HTMLElement,
	) {
		root.classList.add('bz-sb');
		void atlasVars(root, {
			'--sb-tab': 'spellbook-Tab-Frame-C60',
			'--sb-tab-glow': 'spellbook-Tab-Frame-Glow-C60',
			'--sb-tab-gradient': 'spellbook-Tab-Frame-glow-gradient-C60',
			'--sb-square-mask': 'SquareMask',
			'--sb-backplate': 'spellbook-item-backplate',
			'--sb-iconframe': 'spellbook-item-iconframe',
			'--sb-iconframe-shadow': 'spellbook-item-iconframe-shadow',
			'--sb-iconframe-hover': 'spellbook-item-iconframe-hover',
			'--sb-icon-mask': 'spellbook-item-spellicon-mask',
			'--sb-divider': 'spellbook-divider',
			'--sb-list-backplate': 'spellbook-list-backplate',
		});
		void texVarsOn(root, {
			'--sb-prev-up': () => tex(TX.prevPageUp),
			'--sb-prev-down': () => tex(TX.prevPageDown),
			'--sb-prev-disabled': () => tex(TX.prevPageDisabled),
			'--sb-next-up': () => tex(TX.nextPageUp),
			'--sb-next-down': () => tex(TX.nextPageDown),
			'--sb-next-disabled': () => tex(TX.nextPageDisabled),
			'--sb-mouse-hilight': () => glow(TX.mouseHilight),
		});

		// BookBGHalved (minimized) or BookBGLeft and BookBGRight (maximized).
		// The pages' art is too big for a CSS variable, so we set it on each page.
		this.bg = [div(root, 'bz-sb-page'), div(root, 'bz-sb-page'), div(root, 'bz-sb-page')];
		this.bg.forEach((e, i) => void setAtlas(e, i === 1 ? 'spellbook-Page-Left-C60' : 'spellbook-Page-Right-C60'));
		this.tabRow = div(root, 'bz-sb-tabs', { x: 70, y: bookTop + 26 });
		// The wheel turns the pages anywhere over the book (PagedContentFrame's OnMouseWheel).
		root.addEventListener(
			'wheel',
			e => {
				e.preventDefault();
				e.stopPropagation();
				this.turn(e.deltaY > 0 ? 1 : -1);
			},
			{ passive: false },
		);
		this.views = [
			div(root, 'bz-sb-view', { x: 85, y: viewTop, w: viewW, h: viewH }),
			div(root, 'bz-sb-view', { x: 1148 - 50 - viewW, y: viewTop, w: viewW, h: viewH }),
		];
		// PagingControlsHorizontalTemplate at the book's BOTTOMRIGHT -75, 40: the page text, then
		// the two buttons, 8 apart.
		this.paging = div(root, 'bz-sb-paging', { y: bookTop + 702 - 40 - 32, h: 32 });
		this.pageText = text(this.paging, '', 'GameFontHighlight');
		this.pageText.className = 'bz-sb-pagetext';
		this.prev = div(this.paging, 'bz-sb-pagebtn prev');
		this.next = div(this.paging, 'bz-sb-pagebtn next');
		clickable(this.prev, () => this.turn(-1));
		clickable(this.next, () => this.turn(1));
	}

	setMinimized(min: boolean) {
		if (min === this.minimized && !this.stale) return;
		this.minimized = min;
		this.update();
	}

	shown() {
		if (this.stale) this.update();
	}

	refresh() {
		this.stale = true;
		if (this.root.isConnected && this.root.style.display !== 'none') this.update();
	}

	private get viewsPerPage(): number {
		return this.minimized ? 1 : 2;
	}

	// update draws the book: its art for the size, the category tabs and the open page.
	private update() {
		if (this.stale) {
			this.stale = false;
			this.collect();
		}
		if (!this.categories.some(c => c.key === this.selected)) this.selected = this.categories[0]?.key ?? 'general';
		const w = this.minimized ? 574 : 1148;
		const [halved, left, right] = this.bg;
		place(halved, 0, bookTop, w, 702, this.minimized);
		place(left, 0, bookTop, 574, 702, !this.minimized);
		place(right, 574, bookTop, 574, 702, !this.minimized);
		this.views[1].style.display = this.minimized ? 'none' : '';
		// The controls end 75 from the book's right edge (the frame is 3 or 6 wider than the book).
		this.paging.style.right = `${(this.minimized ? 3 : 6) + 75}px`;
		this.drawTabs();
		this.drawPage();
	}

	private drawTabs() {
		this.tabRow.replaceChildren();
		this.categories.forEach((c, i) => {
			const t = div(this.tabRow, 'bz-sb-tab', { x: i * 45, y: 0, w: 44, h: 32 });
			div(t, 'bz-sb-tab-bg');
			const icon = div(t, 'bz-sb-tab-icon');
			void c.icon().then(u => (icon.style.backgroundImage = `url(${u})`));
			div(t, 'bz-sb-tab-glow');
			t.classList.toggle('selected', c.key === this.selected);
			clickable(t, () => {
				if (c.key === this.selected) return;
				this.selected = c.key;
				sound(tabSound);
				this.update();
			});
			this.link.tooltip.attach(t, () => ({ kind: 'lines', title: c.name }));
		});
	}

	// layoutViews splits a category into views, as PagedCondensedVerticalGridContentFrame does:
	// the header takes the first row, and each view fills its left column before its right one,
	// with as few rows as the remaining spells need.
	private layoutViews(key: string): Placed[][] {
		const list = this.entries.get(key) ?? [];
		const name = this.categories.find(c => c.key === key)?.name ?? '';
		const views: Placed[][] = [];
		let i = 0;
		let first = true;
		do {
			const view: Placed[] = [];
			let rows = Math.floor(viewH / rowH);
			let top = 0;
			if (first) {
				view.push({ header: name, col: 0, row: 0 });
				rows = Math.floor((viewH - rowH) / rowH);
				top = 1;
				first = false;
			}
			const left = list.length - i;
			const n = Math.min(rows, Math.ceil(left / 2));
			for (let c = 0; c < 2; c++) {
				for (let r = 0; r < n && i < list.length; r++) view.push({ entry: list[i++], col: c, row: top + r });
			}
			views.push(view);
		} while (i < list.length);
		return views;
	}

	private pageCount(views: Placed[][]): number {
		return Math.max(1, Math.ceil(views.length / this.viewsPerPage));
	}

	private turn(by: number) {
		const views = this.layoutViews(this.selected);
		const page = (this.pages.get(this.selected) ?? 0) + by;
		if (page < 0 || page >= this.pageCount(views)) return;
		this.pages.set(this.selected, page);
		sound(pageTurn[Math.floor(Math.random() * pageTurn.length)]);
		this.drawPage();
	}

	private drawPage() {
		const views = this.layoutViews(this.selected);
		const count = this.pageCount(views);
		const page = Math.min(this.pages.get(this.selected) ?? 0, count - 1);
		this.pages.set(this.selected, page);
		this.views.forEach((v, k) => {
			v.replaceChildren();
			const view = views[page * this.viewsPerPage + k];
			if (!view || (k > 0 && this.minimized)) return;
			for (const p of view) {
				const x = p.col * (colW + xPad);
				const y = p.row * rowH;
				if (p.header !== undefined) this.header(v, p.header, y);
				else if (p.entry) this.item(v, p.entry, x, y);
			}
		});
		this.pageText.textContent = `Page ${page + 1}/${count}`;
		this.prev.classList.toggle('disabled', page === 0);
		this.next.classList.toggle('disabled', page >= count - 1);
	}

	// header is SpellBookHeaderTemplate: the category's name over the divider, on a soft backplate.
	private header(parent: HTMLElement, name: string, y: number) {
		const h = div(parent, 'bz-sb-header', { x: 0, y, w: viewW, h: 60 });
		div(h, 'bz-sb-header-plate', { x: -85, y: 16 + 4 - 53, w: 416, h: 106 });
		div(h, 'bz-sb-header-text', { x: -8, y: 16 - 15, w: viewW - 60 + 8, h: 30 }, name);
		div(h, 'bz-sb-header-divider', { x: -32, y: 49, w: viewW - 60 + 32, h: 11 });
	}

	// item is SpellBookItemTemplate: the icon in its bronze frame on the left, the name and its
	// rank beside it, and the backplate that brightens on hover.
	private item(parent: HTMLElement, entry: Entry, x: number, y: number) {
		const e = div(parent, 'bz-sb-item', { x, y, w: colW, h: 60 });
		div(e, 'bz-sb-backplate');
		const b = div(e, 'bz-sb-button', { x: 0, y: 10, w: 40, h: 40 });
		const icon = div(b, 'bz-sb-icon');
		if (entry.content.kind !== 'macro') setIcon(icon, entry.content.kind, entry.content.id);
		div(b, 'bz-sb-shadow');
		div(b, 'bz-sb-border');
		div(b, 'bz-sb-hi');
		const txt = div(e, 'bz-sb-text');
		div(txt, 'bz-sb-name', undefined, entry.name);
		if (entry.sub) div(txt, 'bz-sb-sub', undefined, entry.sub);
		// A press picks the spell up (OnDragStart and OnClick with the cursor both pick it up).
		b.addEventListener('mousedown', ev => {
			if (ev.button !== 0) return;
			ev.stopPropagation();
			b.classList.add('pushed');
			const up = () => {
				b.classList.remove('pushed');
				removeEventListener('mouseup', up, true);
			};
			addEventListener('mouseup', up, true);
			this.link.pickUp(entry.content);
		});
		b.addEventListener('click', ev => ev.stopPropagation());
		this.link.tooltip.attach(b, () => {
			const c = entry.content;
			return c.kind !== 'macro' ? { kind: c.kind, id: c.id } : undefined;
		});
	}

	// collect sorts the toolkit into categories like the game: General first, then the class's
	// skill lines, each by name and then rank, and the usable items last.
	private collect() {
		const toolkit = this.link.toolkit();
		const cats = new Map<string, Category>();
		const byCat = new Map<string, Entry[]>();
		const add = (key: string, e: Entry) => {
			const list = byCat.get(key) ?? [];
			list.push(e);
			byCat.set(key, list);
		};
		const items: Entry[] = [];
		for (const s of toolkit?.spells ?? []) {
			if (s.itemId) {
				const worn = this.link.sheet()?.items?.find(i => i.id === s.itemId)?.name;
				const name = this.link.itemName(s.itemId) ?? worn ?? this.lookUp('item', s.itemId) ?? this.link.labelOf(s);
				items.push({ content: { kind: 'item', id: s.itemId }, name, sub: '', rankNumber: 0 });
				continue;
			}
			if (!s.spellId) continue;
			const n = toolkit!.names[s.spellId] as { name: string; rank?: string; tab?: { id: number; name: string; icon: number } } | undefined;
			const entry: Entry = {
				content: { kind: 'spell', id: s.spellId },
				name: n?.name ?? this.lookUp('spell', s.spellId) ?? nameOf(this.link, s),
				sub: n?.rank ?? '',
				rankNumber: Number(/Rank (\d+)/.exec(n?.rank ?? '')?.[1] ?? 0),
			};
			const t = n?.tab;
			if (t) {
				const key = `skill:${t.id}`;
				if (!cats.has(key)) cats.set(key, { key, name: t.name, icon: () => tex(t.icon) });
				add(key, entry);
			} else add('general', entry);
		}

		const order = (a: Entry, b: Entry) => a.name.localeCompare(b.name) || a.rankNumber - b.rankNumber || a.sub.localeCompare(b.sub);
		// The General tab shows the class's icon (SpellBookSingleSkillLineCategoryMixin).
		const cls = this.link.loaded.class.replace(/^Class/, '').toLowerCase();
		const general: Category = { key: 'general', name: 'General', icon: async () => iconUrl(`classicon_${cls}`) };
		this.categories = [general, ...[...cats.values()].sort((a, b) => a.name.localeCompare(b.name))];
		this.entries.clear();
		for (const c of this.categories) this.entries.set(c.key, (byCat.get(c.key) ?? []).sort(order));
		if (items.length) {
			this.categories.push({ key: 'items', name: 'Items', icon: () => tex(itemsIcon) });
			this.entries.set('items', items.sort(order));
		}
	}

	// lookUp gives a name we found in its tooltip, or asks for it and redraws once it's there.
	private lookUp(kind: 'spell' | 'item', id: number): string | undefined {
		const key = `${kind}:${id}`;
		const name = this.lookedUp.get(key);
		if (name || this.looking.has(key)) return name;
		this.looking.add(key);
		void tooltipInfo(kind, id).then(info => {
			if (!info.name) return;
			this.lookedUp.set(key, info.name);
			this.refresh();
		});
		return undefined;
	}
}

function place(e: HTMLElement, x: number, y: number, w: number, h: number, shown: boolean) {
	e.style.left = `${x}px`;
	e.style.top = `${y}px`;
	e.style.width = `${w}px`;
	e.style.height = `${h}px`;
	e.style.display = shown ? '' : 'none';
}

// nameOf is a spell's name when the toolkit has none for it: the room's label without the rank.
function nameOf(link: RoomLink, s: ToolkitSpell): string {
	return link.labelOf(s).replace(/ \(Rank \d+\)$/, '');
}

// atlasVars sets CSS variables on an element to atlases' art, so many elements share one image.
export async function atlasVars(el: HTMLElement, vars: Record<string, string>) {
	await Promise.all(
		Object.entries(vars).map(async ([v, name]) => {
			const u = await atlasUrl(name).catch(e => console.warn('atlas', name, e));
			if (u) el.style.setProperty(v, `url(${u})`);
		}),
	);
}

async function texVarsOn(el: HTMLElement, vars: Record<string, () => Promise<string>>) {
	await Promise.all(Object.entries(vars).map(async ([v, f]) => el.style.setProperty(v, `url(${await f()})`)));
}
