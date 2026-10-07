import tippy from 'tippy.js';

import { IndividualSimUI } from '../../individual_sim_ui';
import { Player } from '../../player';
import { Spec } from '../../proto/common';
import { SavedTalents } from '../../proto/ui';
import { classTalentsConfig } from '../../talents/factory';
import { TalentsPicker } from '../../talents/talents_picker';
import { EventID, TypedEvent } from '../../typed_event';
import { dirtySettings } from '../dirty_settings';
import { SavedDataManager } from '../saved_data_manager';
import { presetListTooltip } from '../preset_tree';
import { SimTab } from '../sim_tab';

// The Legacy perk Talented gives one talent point per rank.
const TALENTED_MAX_RANK = 5;

export class TalentsTab extends SimTab {
	protected simUI: IndividualSimUI<Spec>;

	readonly leftPanel: HTMLElement;
	readonly rightPanel: HTMLElement;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		super(parentElem, simUI, { identifier: 'talents-tab', title: 'Talents' });
		this.simUI = simUI;

		this.leftPanel = document.createElement('div');
		this.leftPanel.classList.add('talents-tab-left', 'tab-panel-left');
		this.rightPanel = document.createElement('div');
		this.rightPanel.classList.add('talents-tab-right', 'tab-panel-right', 'within-raid-sim-hide');

		// The saved talents come first, to the left of the trees, like the gear sets in the Gear
		// tab.
		this.contentContainer.appendChild(this.rightPanel);
		this.contentContainer.appendChild(this.leftPanel);

		this.buildTabContent();
	}

	protected buildTabContent() {
		const talentsPicker = this.buildTalentsPicker(this.leftPanel);

		this.buildTalentedPerk(talentsPicker.rootElem.querySelector('.talents-picker-header') as HTMLElement);
		this.buildSavedTalentsPicker();
	}

	// The extra talent points, as the Legacy perk Talented that gives them, next to the points
	// left. Like on foreverchanges.pro, it's the perk's icon and its ranks from 0 to 5.
	private buildTalentedPerk(header: HTMLElement) {
		const player = this.simUI.player;
		const perk = document.createElement('div');
		perk.classList.add('talented-perk');

		const icon = document.createElement('img');
		icon.classList.add('talented-perk-icon');
		icon.src = 'https://wow.zamimg.com/images/wow/icons/large/ability_marksmanship.jpg';
		icon.alt = 'Talented';
		// The tooltip reads like the perk's in the game. Each rank starts the talent points one
		// level earlier than level 10. At rank 0 it describes rank 1, like the game does for a
		// perk we don't have yet.
		tippy(icon, {
			onShow: instance => {
				const rank = player.getBonusTalentPoints();
				const startLevel = 10 - Math.max(rank, 1);
				instance.setContent(
					`<div class="talented-perk-tooltip">` +
						`<div class="talented-perk-tooltip-name">Talented</div>` +
						`<div>Rank <b>${rank}/${TALENTED_MAX_RANK}</b></div>` +
						`<div class="talented-perk-tooltip-kind">Passive</div>` +
						`<div class="talented-perk-tooltip-text">You gain talent points every level starting at level <b>${startLevel}</b> instead of starting at level <b>10</b>, but you still may not have more than <b>51</b> total talent points.</div>` +
						`</div>`,
				);
			},
		});

		const ranks = document.createElement('span');
		ranks.classList.add('talented-perk-ranks');
		ranks.setAttribute('role', 'group');
		ranks.setAttribute('aria-label', 'Ranks of the Legacy perk Talented');
		const buttons = Array.from({ length: TALENTED_MAX_RANK + 1 }, (_, rank) => {
			const button = document.createElement('button');
			button.type = 'button';
			button.textContent = `${rank}`;
			button.addEventListener('click', () => player.setBonusTalentPoints(TypedEvent.nextEventID(), rank));
			ranks.appendChild(button);
			return button;
		});

		const update = () => buttons.forEach((button, rank) => button.setAttribute('aria-pressed', `${rank === player.getBonusTalentPoints()}`));
		update();
		player.miscOptionsChangeEmitter.on(update);

		perk.append(icon, ranks);
		header.insertBefore(perk, header.children[1] ?? null);

		// A preset configuration can set the extra points, so the perk is marked when we change
		// them after loading one.
		this.addOnDisposeCallback(
			dirtySettings.track({
				elem: perk,
				read: () => player.getBonusTalentPoints(),
				name: () => 'Talented',
			}),
		);
	}

	private buildTalentsPicker(parentElem: HTMLElement): TalentsPicker<any> {
		return new TalentsPicker(parentElem, this.simUI.player, {
			klass: this.simUI.player.getClass(),
			trees: classTalentsConfig[this.simUI.player.getClass()],
			changedEvent: (player: Player<any>) => player.talentsChangeEmitter,
			getValue: (player: Player<any>) => player.getTalentsString(),
			setValue: (eventID: EventID, player: Player<any>, newValue: string) => {
				player.setTalentsString(eventID, newValue);
			},
			pointsPerRow: 5,
		});
	}

	private buildSavedTalentsPicker() {
		const savedTalentsManager = new SavedDataManager<Player<any>, SavedTalents>(this.rightPanel, this.simUI.player, {
			label: 'Talents',
			header: {
				title: 'Saved Talents',
				tooltip: presetListTooltip('Loading a talent build changes:', ['The talent points in all three trees']),
			},
			storageKey: this.simUI.getSavedTalentsStorageKey(),
			getData: (player: Player<any>) =>
				SavedTalents.create({
					talentsString: player.getTalentsString(),
				}),
			setData: (eventID: EventID, player: Player<any>, newTalents: SavedTalents) => {
				TypedEvent.freezeAllAndDo(() => {
					player.setTalentsString(eventID, newTalents.talentsString);
				});
			},
			changeEmitters: [this.simUI.player.talentsChangeEmitter],
			equals: (a: SavedTalents, b: SavedTalents) => SavedTalents.equals(a, b),
			toJson: (a: SavedTalents) => SavedTalents.toJson(a),
			fromJson: (obj: any) => SavedTalents.fromJson(obj),
		});

		this.simUI.sim.waitForInit().then(() => {
			savedTalentsManager.loadUserData();
			this.simUI.individualConfig.presets.talents.forEach(config => {
				config.isPreset = true;
				savedTalentsManager.addSavedData({
					name: config.name,
					group: config.group,
					tooltip: config.tooltip,
					isPreset: true,
					data: config.data,
					enableWhen: config.enableWhen,
				});
			});
		});
	}
}
