// The Blizzard frames of the room: the character frame, spellbook, talents, bags, the Game Menu
// and Options, and the chat frame. main.ts makes one Frames and sends it the keys' commands.

import { loadFonts, sharedVars } from './assets';
import { BagsFrame } from './bags';
import { CharacterFrame } from './character';
import { ChatFrame, ChatLayout } from './chat';
import { GameMenu, MenuHost, OptionsFrame } from './gamemenu';
import type { RoomLink } from './link';
import { UIPanels } from './panels';
import { SpellbookFrame } from './spellbook';
import { TalentFrame } from './talents';

export type { MenuHost } from './gamemenu';
export type { RoomLink } from './link';

export class Frames {
	readonly panels = new UIPanels();
	readonly character: CharacterFrame;
	readonly spellbook: SpellbookFrame;
	readonly talents: TalentFrame;
	readonly bags: BagsFrame;
	readonly menu: GameMenu;
	readonly options: OptionsFrame;
	readonly chat: ChatFrame;

	constructor(
		readonly link: RoomLink,
		host: MenuHost,
		chatLayout: ChatLayout | undefined,
	) {
		const p = this.panels;
		this.character = p.add(new CharacterFrame(link));
		this.spellbook = p.add(new SpellbookFrame(link));
		this.talents = p.add(new TalentFrame(link));
		this.bags = p.add(new BagsFrame(link));
		this.options = p.add(new OptionsFrame(host));
		this.menu = p.add(
			new GameMenu([
				[
					{ label: 'Options', run: () => this.openOptions() },
					{
						label: 'Key Bindings',
						run: () => {
							p.close(this.menu, true);
							host.bindKeys();
						},
					},
				],
				[
					{
						label: 'Leave Room',
						run: () => {
							p.close(this.menu, true);
							host.leave();
						},
					},
				],
				[{ label: 'Return to Game', run: () => p.close(this.menu) }],
			]),
		);
		// Closing Options brings the Game Menu back, like the game.
		this.options.onClosed = () => {
			if (this.reopenMenu) {
				this.reopenMenu = false;
				void p.open(this.menu);
			}
		};
		this.chat = new ChatFrame(link, chatLayout);
		this.panels.layer.append(this.chat.el);
	}

	private reopenMenu = false;

	private openOptions() {
		this.reopenMenu = true;
		this.panels.close(this.menu, true);
		void this.panels.open(this.options);
	}

	// attach puts the frames' layer in the HUD, before the element given (the HUD's pieces that
	// must stay on top, like the cursor icon and the tooltip).
	attach(hud: HTMLElement, before: Element | null) {
		hud.insertBefore(this.panels.layer, before);
	}

	// preload loads the textures and fonts every frame shares.
	preload() {
		void Promise.all([sharedVars(), loadFonts()]);
	}

	setScale(u: number) {
		this.panels.setScale(u);
	}

	// command runs a key binding's command and says whether it was one of ours.
	command(command: string): boolean {
		const p = this.panels;
		switch (command) {
			case 'TOGGLECHARACTER0':
				this.toggleCharacter(0);
				return true;
			case 'TOGGLECHARACTER1':
				this.toggleCharacter(1);
				return true;
			case 'TOGGLECHARACTER2':
				this.toggleCharacter(2);
				return true;
			case 'TOGGLESPELLBOOK':
				p.toggle(this.spellbook);
				return true;
			case 'TOGGLETALENTS':
				p.toggle(this.talents);
				return true;
			case 'OPENALLBAGS':
			case 'TOGGLEBACKPACK':
				p.toggle(this.bags);
				return true;
			case 'TOGGLEGAMEMENU':
				this.toggleMenu();
				return true;
			case 'OPENCHAT':
				this.chat.openEditBox();
				return true;
		}
		return false;
	}

	// toggleCharacter opens the character frame on a tab, switches tabs when it's open on
	// another, and closes it when it's open on that one.
	toggleCharacter(tab: number) {
		const c = this.character;
		if (c.isOpen && c.tab === tab) this.panels.close(c);
		else {
			c.showTab(tab);
			void this.panels.open(c);
		}
	}

	toggleMenu() {
		if (this.options.isOpen) {
			this.panels.close(this.options);
			return;
		}
		this.panels.toggle(this.menu);
	}

	// escapeDialog closes the Game Menu or Options, or the chat's edit box: the first steps of
	// the game's Esc. It says whether it did.
	escapeDialog(): boolean {
		if (this.chat.closeEditBox()) return true;
		if (this.options.isOpen) {
			this.panels.close(this.options);
			return true;
		}
		if (this.menu.isOpen) {
			this.panels.close(this.menu);
			return true;
		}
		return false;
	}

	// closeWindows closes the open panels (CloseAllWindows) and says whether there was one.
	closeWindows(): boolean {
		return this.panels.closeAll();
	}

	openMenu() {
		void this.panels.open(this.menu);
	}
}
