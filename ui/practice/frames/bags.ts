// The bags (I): the backpack as Forever draws it (ContainerFrame1 of Mainline's
// ContainerFrame.xml, with bags not combined, the default), read only. It holds the items your
// toolkit uses that you don't wear (potions, runes, explosives and the like). We don't know
// their counts, your money or your other bags, so it shows only the backpack's 16 slots, with
// no counts, no money and no search box or sort button.

import './bags.css';
import { nineSlice, setAtlas, type NineSliceLayout } from './atlas';
import { tex } from './assets';
import type { RoomLink } from './link';
import { closeButton } from './modern';
import { div, itemButton, Panel, text, type ItemButton } from './panels';

const slots = 16;
// Interface\Icons\INV_Misc_Bag_08, the backpack's portrait.
const backpackIcon = 133633;
// The sound files of IG_BACKPACK_OPEN and IG_BACKPACK_CLOSE in Forever.
const openSound = 567461;
const closeSound = 567512;

// HeldBagLayout from NineSliceLayouts.lua, with Forever's fixes from
// NineSliceLayoutOverrides.lua (the right corners 2 to the left, the bottom corners at y -8).
const heldBagLayout: NineSliceLayout = {
	TopLeftCorner: { atlas: 'ui-frame-portraitmetal-cornertopleftsmall', x: -13, y: 16 },
	TopRightCorner: { atlas: 'UI-Frame-Metal-CornerTopRight', x: 2, y: 16 },
	BottomLeftCorner: { atlas: 'UI-Frame-Metal-CornerBottomLeft', x: -13, y: -8 },
	BottomRightCorner: { atlas: 'UI-Frame-Metal-CornerBottomRight', x: 2, y: -8 },
	TopEdge: { atlas: '_UI-Frame-Metal-EdgeTop' },
	BottomEdge: { atlas: '_UI-Frame-Metal-EdgeBottom' },
	LeftEdge: { atlas: '!UI-Frame-Metal-EdgeLeft' },
	RightEdge: { atlas: '!UI-Frame-Metal-EdgeRight' },
};

export class BagsFrame extends Panel {
	// The bag frame's width (178), and the backpack's height: 4 rows of 37 with 5 between, the
	// backpack's padding (87, with room for the search box) and the money frame (13).
	readonly width = 178;
	readonly height = 4 * 37 + 3 * 5 + 87 + 13;
	area = 'free' as const;
	openSound = openSound;
	closeSound = closeSound;

	private buttons: ItemButton[] = [];
	private items: number[] = [];

	constructor(readonly link: RoomLink) {
		super();
	}

	protected build(root: HTMLElement) {
		root.classList.add('bz-bags');
		// ContainerFrame1 sits CONTAINER_OFFSET_Y (85) above the screen's bottom and 10 from its
		// right (no right side action bars in the default layout).
		root.style.left = '';
		root.style.right = '10px';
		root.style.bottom = '85px';

		// FlatPanelBackgroundTemplate: PANEL_BACKGROUND_COLOR (31, 30, 33 at 0.8 in the client's
		// GlobalColor table). Its bottom corner pieces only round the last pixel or two.
		div(root, 'bz-bags-bg');
		nineSlice(root, heldBagLayout);

		// The round portrait in the metal ring of the small top left corner.
		const portrait = div(root, 'bz-bags-portrait');
		void tex(backpackIcon).then(u => (portrait.style.backgroundImage = `url(${u})`));

		const title = text(root, 'Backpack', 'GameFontNormal', undefined, 'center');
		title.classList.add('bz-bags-title');
		closeButton(root, { right: 0, y: -3 }, () => this.close());

		// The slots fill from the bottom right, 4 to a row, 5 apart, with the last one 8 from the
		// right and 4 above the money frame. Slot 1 is at the top left.
		const left = this.width - 8 - (4 * 37 + 3 * 5);
		const top = this.height - 8 - 13 - 4 - (4 * 37 + 3 * 5);
		for (let i = 0; i < slots; i++) {
			const col = i % 4;
			const row = Math.floor(i / 4);
			const b = itemButton(root, { x: left + col * 42, y: top + row * 42 }, undefined, false);
			void setAtlas(div(b.el, 'bz-bags-slot'), 'bags-item-slot64');
			const index = i;
			this.link.tooltip.attach(b.el, () => {
				const id = this.items[index];
				return id ? { kind: 'item', id } : undefined;
			}, true);
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
