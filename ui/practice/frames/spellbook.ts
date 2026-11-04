// The spellbook (K): the Classic SpellBookFrame. Your spells sit on side tabs by skill line
// (General, then the class's talent schools), twelve to a page, and the items you can use get a
// last Items tab of their own. You pick a spell or item up from it to drop on an action button.
//
// Sizes and anchors are from the vanilla spellbookframe.xml. The game hides the bottom tab row
// unless you have a pet, so we have none.

import './spellbook.css';
import type { SlotContent } from '../actions';
import { setIcon } from '../icons';
import { tooltipInfo } from '../tooltip';
import type { ToolkitSpell } from '../protocol';
import type { RoomLink } from './link';
import { checkbox, clickable, closeButton, div, fourPiece, Panel, text, texture } from './panels';
import { SND, sound, tex, texVars, glow, TX } from './assets';

// The General tab's book and the Items tab's bag.
const generalIcon = 133741; // inv_misc_book_09
const itemsIcon = 133633; // inv_misc_bag_08

const perPage = 12;

type SkillTab = { key: string; name: string; icon: number };

type Entry = { content: SlotContent; name: string; rank: string; rankNumber: number };

type SpellButton = { el: HTMLDivElement; icon: HTMLDivElement; name: HTMLSpanElement; rank: HTMLDivElement; entry?: Entry };

type SideTab = { el: HTMLDivElement; icon: HTMLDivElement; tab?: SkillTab };

export class SpellbookFrame extends Panel {
	readonly width = 384;
	readonly height = 512;
	stackWidth = 353;
	openSound = SND.spellbookOpen;
	closeSound = SND.spellbookClose;

	private tabs: SkillTab[] = [];
	private entries = new Map<string, Entry[]>();
	private selected = 'general';
	// The open page of each tab (SPELLBOOK_PAGENUMBERS), from 0.
	private pages = new Map<string, number>();
	private allRanks = true;
	private buttons: SpellButton[] = [];
	private sideTabs: SideTab[] = [];
	private pageText?: HTMLDivElement;
	private prev?: HTMLDivElement;
	private next?: HTMLDivElement;
	// Whether the toolkit changed since we last drew the pages.
	private stale = true;
	// Names the toolkit lacks, from the tooltips (a totem the server sent no name for, or a
	// synthetic item), by "spell:id" or "item:id".
	private lookedUp = new Map<string, string>();
	private looking = new Set<string>();

	constructor(readonly link: RoomLink) {
		super();
	}

	protected build(root: HTMLElement) {
		root.classList.add('bz-spellbook');
		// The parts' art fills in as it loads, so the book opens at once.
		void texVars({
			'bz-spell-bg': () => tex(TX.spellbookSpellBackground),
			'bz-spell-tab': () => tex(TX.spellbookSkillLineTab),
			'bz-check-hilight': () => glow(TX.checkButtonHilight),
			'bz-prev-up': () => tex(TX.prevPageUp),
			'bz-prev-down': () => tex(TX.prevPageDown),
			'bz-prev-disabled': () => tex(TX.prevPageDisabled),
			'bz-next-up': () => tex(TX.nextPageUp),
			'bz-next-down': () => tex(TX.nextPageDown),
			'bz-next-disabled': () => tex(TX.nextPageDisabled),
			'bz-mouse-hilight': () => glow(TX.mouseHilight),
		});

		// The book takes the mouse only where its art is (the game's HitRectInsets leave out 30 on
		// the right and 70 at the bottom), and the side tabs on their own.
		const hit = div(root, 'bz-spellbook-hit', { x: 0, y: 0, w: 354, h: 442 });
		texture(hit, TX.spellbookIcon, { x: 10, y: 8, w: 58, h: 58 });
		fourPiece(hit, [TX.spellbookTopLeft, TX.spellbookTopRight, TX.spellbookBotLeft, TX.spellbookBotRight]);
		text(hit, 'Spellbook', 'GameFontNormal', { x: 98, y: 19, w: 200 }, 'center');
		closeButton(hit, { x: 324, y: 9 }, () => this.close());

		const ranks = checkbox(hit, 'Show All Spell Ranks', { x: 74, y: 38 }, this.allRanks, on => {
			this.allRanks = on;
			this.stale = true;
			this.update();
		});
		ranks.classList.add('bz-spellbook-ranks');

		for (let i = 0; i < perPage; i++) {
			// The left column holds the page's first six, the right column the next six.
			const x = 34 + (i < 6 ? 0 : 157);
			const y = 85 + (i % 6) * 51;
			this.buttons.push(this.spellButton(hit, x, y));
		}

		this.prev = div(hit, 'bz-page-btn prev', { x: 34, y: 391, w: 32, h: 32 });
		this.next = div(hit, 'bz-page-btn next', { x: 298, y: 391, w: 32, h: 32 });
		clickable(this.prev, () => this.turn(-1));
		clickable(this.next, () => this.turn(1));
		this.pageText = text(hit, '', 'GameFontNormal', { x: 127, y: 402, w: 102 }, 'center');

		for (let i = 0; i < 8; i++) this.sideTabs.push(this.sideTab(root, i));
	}

	protected shown() {
		if (this.stale) this.update();
	}

	// refresh redraws the pages, when the toolkit changes.
	refresh() {
		this.stale = true;
		if (this.isOpen) this.update();
	}

	private spellButton(parent: HTMLElement, x: number, y: number): SpellButton {
		const el = div(parent, 'bz-spell', { x, y, w: 37, h: 37 });
		const icon = div(el, 'bz-spell-icon');
		div(el, 'bz-spell-frame');
		// The name is centered beside the icon, 3 above its middle, and the rank hangs under it.
		const label = text(el, '', 'GameFontNormal');
		label.classList.add('bz-spell-name');
		const name = document.createElement('span');
		label.append(name);
		const rank = text(label, '', 'SubSpellFont');
		rank.classList.add('bz-spell-rank');
		const b: SpellButton = { el, icon, name, rank };
		el.addEventListener('mousedown', e => {
			if (e.button !== 0 || !b.entry) return;
			e.stopPropagation();
			el.classList.add('pushed');
			const up = () => {
				el.classList.remove('pushed');
				removeEventListener('mouseup', up, true);
			};
			addEventListener('mouseup', up, true);
			this.link.pickUp(b.entry.content);
		});
		el.addEventListener('click', e => e.stopPropagation());
		this.link.tooltip.attach(el, () => {
			const c = b.entry?.content;
			return c && c.kind !== 'macro' ? { kind: c.kind, id: c.id } : undefined;
		});
		return b;
	}

	private sideTab(parent: HTMLElement, i: number): SideTab {
		const el = div(parent, 'bz-skilltab', { x: 352, y: 65 + i * 49, w: 32, h: 32 });
		const icon = div(el, 'bz-skilltab-icon');
		const t: SideTab = { el, icon };
		clickable(el, () => {
			if (!t.tab || t.tab.key === this.selected) return;
			this.selected = t.tab.key;
			sound(SND.characterTab);
			this.update();
		});
		this.link.tooltip.attach(el, () => (t.tab ? { kind: 'lines', title: t.tab.name } : undefined));
		return t;
	}

	private turn(by: number) {
		const page = (this.pages.get(this.selected) ?? 0) + by;
		if (page < 0 || page >= this.pageCount()) return;
		this.pages.set(this.selected, page);
		sound(SND.pageTurn);
		this.update();
	}

	private pageCount(): number {
		return Math.max(1, Math.ceil((this.entries.get(this.selected)?.length ?? 0) / perPage));
	}

	// update draws the selected tab's open page, and the side tabs.
	private update() {
		if (!this.pageText) return;
		if (this.stale) {
			this.stale = false;
			this.collect();
		}
		if (!this.tabs.some(t => t.key === this.selected)) this.selected = this.tabs[0]?.key ?? 'general';

		this.sideTabs.forEach((s, i) => {
			const t = this.tabs[i];
			s.tab = t;
			s.el.style.display = t ? '' : 'none';
			s.el.classList.toggle('selected', t?.key === this.selected);
			if (t && s.icon.dataset.icon !== String(t.icon)) {
				s.icon.dataset.icon = String(t.icon);
				void tex(t.icon).then(u => {
					if (s.icon.dataset.icon === String(t.icon)) s.icon.style.backgroundImage = `url(${u})`;
				});
			}
		});

		const list = this.entries.get(this.selected) ?? [];
		const count = this.pageCount();
		const page = Math.min(this.pages.get(this.selected) ?? 0, count - 1);
		this.pages.set(this.selected, page);
		this.buttons.forEach((b, i) => this.fill(b, list[page * perPage + i]));
		this.pageText.textContent = `Page ${page + 1}`;
		this.prev!.classList.toggle('disabled', page === 0);
		this.next!.classList.toggle('disabled', page >= count - 1);
	}

	private fill(b: SpellButton, entry: Entry | undefined) {
		b.entry = entry;
		b.el.classList.toggle('filled', !!entry);
		if (!entry || entry.content.kind === 'macro') {
			b.icon.style.backgroundImage = '';
			delete b.icon.dataset.icon;
			b.name.textContent = '';
			b.rank.textContent = '';
			return;
		}
		setIcon(b.icon, entry.content.kind, entry.content.id);
		b.name.textContent = entry.name;
		b.rank.textContent = entry.rank;
	}

	// collect sorts the toolkit into the side tabs, like the game: General first, then the
	// class's skill lines, each by name and then rank.
	private collect() {
		const toolkit = this.link.toolkit();
		const tabs = new Map<string, SkillTab>();
		const byTab = new Map<string, Entry[]>();
		const add = (key: string, e: Entry) => {
			const list = byTab.get(key) ?? [];
			list.push(e);
			byTab.set(key, list);
		};
		const items: Entry[] = [];
		for (const s of toolkit?.spells ?? []) {
			if (s.itemId) {
				const worn = this.link.sheet()?.items?.find(i => i.id === s.itemId)?.name;
				const name = this.link.itemName(s.itemId) ?? worn ?? this.lookUp('item', s.itemId) ?? this.link.labelOf(s);
				items.push({ content: { kind: 'item', id: s.itemId }, name, rank: '', rankNumber: 0 });
				continue;
			}
			if (!s.spellId) continue;
			const n = toolkit!.names[s.spellId] as { name: string; rank?: string; tab?: { id: number; name: string; icon: number } } | undefined;
			const entry: Entry = {
				content: { kind: 'spell', id: s.spellId },
				name: n?.name ?? this.lookUp('spell', s.spellId) ?? nameOf(this.link, s),
				rank: n?.rank ?? '',
				rankNumber: Number(/Rank (\d+)/.exec(n?.rank ?? '')?.[1] ?? 0),
			};
			const t = n?.tab;
			if (t) {
				const key = `skill:${t.id}`;
				if (!tabs.has(key)) tabs.set(key, { key, name: t.name, icon: t.icon });
				add(key, entry);
			} else add('general', entry);
		}

		const order = (a: Entry, b: Entry) => a.name.localeCompare(b.name) || a.rankNumber - b.rankNumber || a.rank.localeCompare(b.rank);
		this.entries.clear();
		const general: SkillTab = { key: 'general', name: 'General', icon: generalIcon };
		this.tabs = [general, ...[...tabs.values()].sort((a, b) => a.name.localeCompare(b.name))];
		for (const t of this.tabs) this.entries.set(t.key, this.ranked((byTab.get(t.key) ?? []).sort(order)));
		if (items.length) {
			this.tabs.push({ key: 'items', name: 'Items', icon: itemsIcon });
			this.entries.set('items', items.sort(order));
		}
		// The game has eight side tabs at most.
		this.tabs = this.tabs.slice(0, 8);
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

	// ranked keeps only each spell's highest rank, unless Show All Spell Ranks is on.
	private ranked(list: Entry[]): Entry[] {
		if (this.allRanks) return list;
		return list.filter((e, i) => list[i + 1]?.name !== e.name);
	}
}

// nameOf is a spell's name when the toolkit has none for it: the room's label without the rank.
function nameOf(link: RoomLink, s: ToolkitSpell): string {
	return link.labelOf(s).replace(/ \(Rank \d+\)$/, '');
}
