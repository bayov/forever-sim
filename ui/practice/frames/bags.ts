// The bags (I): the backpack's ContainerFrame, read only. It holds the items your toolkit uses
// that you don't wear (potions, runes, explosives and the like). We don't know their counts or
// your other bags, so it shows only the backpack and no counts.

import './bags.css';
import type { RoomLink } from './link';
import { closeButton, itemButton, Panel, text, texture, type ItemButton } from './panels';
import { SND, TX } from './assets';

const slots = 16;

export class BagsFrame extends Panel {
	// CONTAINER_WIDTH, and the backpack's art height.
	readonly width = 192;
	readonly height = 256;
	area = 'free' as const;
	openSound = SND.bagOpen;
	closeSound = SND.bagClose;

	private buttons: ItemButton[] = [];
	private items: number[] = [];

	constructor(readonly link: RoomLink) {
		super();
	}

	protected build(root: HTMLElement) {
		root.classList.add('bz-bags');
		// ContainerFrame1 sits at the screen's bottom right, CONTAINER_OFFSET_Y above the bottom.
		root.style.right = '0px';
		root.style.bottom = '70px';
		// The 256 wide art hangs from the frame's top right, so its see-through left part is
		// outside the frame.
		texture(root, TX.backpackBackground, { x: -64, y: 0, w: 256, h: 256 });
		text(root, 'Backpack', 'GameFontHighlight', { x: 47, y: 10, w: 112, h: 12 });
		closeButton(root, { x: 160, y: 0 }, () => this.close());
		// Slot 1 is top left. The game lays the buttons from the last slot at the bottom right,
		// 5 apart across and 4 apart up.
		for (let i = 0; i < slots; i++) {
			const col = i % 4;
			const row = Math.floor(i / 4);
			const b = itemButton(root, { x: 17 + col * 42, y: 63 + row * 41 });
			const index = i;
			this.link.tooltip.attach(b.el, () => {
				const id = this.items[index];
				return id ? { kind: 'item', id } : undefined;
			});
			b.el.addEventListener('mousedown', e => e.stopPropagation());
			this.buttons.push(b);
		}
	}

	protected shown() {
		const worn = new Set((this.link.sheet()?.items ?? []).map(i => i?.id).filter(Boolean));
		this.items = (this.link.toolkit()?.spells ?? [])
			.map(s => s.itemId)
			.filter((id): id is number => !!id && !worn.has(id))
			.filter((id, i, all) => all.indexOf(id) === i)
			.slice(0, slots);
		this.buttons.forEach((b, i) => {
			const id = this.items[i];
			b.set(id ? { kind: 'item', id } : undefined);
		});
	}
}
