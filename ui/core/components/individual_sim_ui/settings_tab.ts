import tippy from 'tippy.js';

import * as Tooltips from '../../constants/tooltips';
import { Encounter } from '../../encounter';
import { IndividualSimUI, InputConfig, InputSection } from '../../individual_sim_ui';
import { Player } from '../../player';
import { Consumes, Debuffs, HealingModel, IndividualBuffs, ItemSwap, PartyBuffs, RaidBuffs, Spec } from '../../proto/common';
import { SavedEncounter, SavedSettings } from '../../proto/ui';
import { EventID, TypedEvent } from '../../typed_event';
import { BooleanPicker } from '../boolean_picker';
import { ContentBlock } from '../content_block';
import { EncounterPicker } from '../encounter_picker';
import { EnumPicker } from '../enum_picker';
import { IconEnumPicker } from '../icon_enum_picker';
import * as IconInputs from '../icon_inputs';
import { Input } from '../input';
import * as BuffDebuffInputs from '../inputs/buffs_debuffs';
import { relevantStatOptions } from '../inputs/stat_options';
import { MultiIconPicker, MultiIconPickerItemConfig } from '../multi_icon_picker';
import { NumberPicker } from '../number_picker';
import { SavedDataManager } from '../saved_data_manager';
import { presetListTooltip } from '../preset_tree';
import { SimTab } from '../sim_tab';
import * as OtherInputs from './../other_inputs';
import { IsbConfig, StormstrikeConfig } from './../other_inputs';
import { ConsumesPicker } from './consumes_picker';
import { ItemSwapPicker } from './item_swap_picker';
import { IconEnumRowPicker, LevelPicker, ProfessionsPicker, RacePicker } from './player_pickers';
import { RacialsPicker } from './racials_picker';

export class SettingsTab extends SimTab {
	protected simUI: IndividualSimUI<Spec>;

	readonly leftPanel: HTMLElement;
	readonly rightPanel: HTMLElement;

	// The saved encounters, at the top of the Encounter section. The raid sim's player editor
	// has no Encounter section.
	private encounterPresets?: HTMLElement;
	private encounterPicker?: EncounterPicker;

	// The individual sim has three lanes. The left one is about us: the saved settings, the
	// Player, the Class Settings and the consumables. The middle one is about the target: the
	// Encounter and the debuffs. The right one has the buffs the party and the raid give us.
	readonly column1: HTMLElement = this.buildColumn(1, 'settings-left-col');
	readonly column2: HTMLElement = this.buildColumn(2, 'settings-left-col');
	readonly column3: HTMLElement = this.buildColumn(3, 'settings-left-col');
	readonly column4?: HTMLElement;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>) {
		super(parentElem, simUI, { identifier: 'settings-tab', title: 'Settings' });
		this.simUI = simUI;

		this.leftPanel = document.createElement('div');
		this.leftPanel.classList.add('settings-tab-left', 'tab-panel-left');

		this.leftPanel.appendChild(this.column1);
		this.leftPanel.appendChild(this.column2);
		this.leftPanel.appendChild(this.column3);

		// The 4th column is only used in the raid sim player editor to spread out player settings
		if (this.simUI.isWithinRaidSim) {
			this.column4 = this.buildColumn(4, 'settings-left-col');
			this.leftPanel.appendChild(this.column4);
		}

		// The saved settings are at the top of the left lane. The raid sim's player editor has
		// none.
		this.rightPanel = document.createElement('div');
		this.rightPanel.classList.add('settings-tab-saved', 'within-raid-sim-hide');
		this.column1.appendChild(this.rightPanel);
		this.contentContainer.appendChild(this.leftPanel);

		this.buildTabContent();
	}

	protected buildTabContent() {
		if (!this.simUI.isWithinRaidSim) {
			this.buildEncounterSettings();
		}

		// We build the sections of each lane from the top down.
		this.simUI.sim.waitForInit().then(() => {
			this.buildPlayerSettings();
			this.buildCustomSettingsSections();
			this.buildOtherSettings();
			this.buildConsumesSection();

			if (!this.simUI.isWithinRaidSim) {
				this.buildEncounterInputs();
				this.buildDebuffsSettings();
				this.buildIsbSettings();
				this.buildStormstrikeSettings();

				this.buildBuffsSection(this.column3, 'party-buffs-settings', 'Party Buffs', PARTY_BUFFS_TOOLTIP, 'party');
				this.buildBuffsSection(this.column3, 'buffs-settings', 'Raid Buffs', RAID_BUFFS_TOOLTIP, 'raid');
				this.buildWorldBuffsSettings();
				this.buildSavedDataPickers();
			}
		});
	}

	private buildEncounterSettings() {
		const contentBlock = new ContentBlock(this.column2, 'encounter-settings', {
			header: { title: 'Encounter' },
		});

		this.encounterPresets = document.createElement('div');
		this.encounterPresets.classList.add('encounter-presets');
		this.encounterPresets.innerHTML = '<span>Presets:</span><div class="saved-data-presets"></div>';
		contentBlock.bodyElement.appendChild(this.encounterPresets);

		this.encounterPicker = new EncounterPicker(contentBlock.bodyElement, this.simUI.sim.encounter, this.simUI.individualConfig.encounterPicker, this.simUI);
	}

	// The player's inputs that describe the fight go in the Encounter section, above the
	// Advanced button: In Front of Target, and the spec's encounterInputs (like how often raid
	// damage hits an Enhancement shaman). The encounter picker adds the Advanced button once the
	// sim loads, which is before we get here.
	private buildEncounterInputs() {
		const inputs = [
			...(this.simUI.individualConfig.otherInputs.inputs.includes(OtherInputs.InFrontOfTarget) ? [OtherInputs.InFrontOfTarget] : []),
			...(this.simUI.individualConfig.encounterInputs?.inputs ?? []),
		];
		if (!inputs.length || !this.encounterPicker) return;

		const container = document.createElement('div');
		container.classList.add('encounter-player-inputs');
		this.configureInputSection(container, { inputs });
		container.querySelectorAll('.input-root').forEach(elem => elem.classList.add('input-inline'));
		const advancedButton = this.encounterPicker.rootElem.querySelector(':scope > .advanced-button');
		this.encounterPicker.rootElem.insertBefore(container, advancedButton);
	}

	private buildPlayerSettings() {
		const contentBlock = new ContentBlock(this.column1, 'player-settings', {
			header: { title: 'Player' },
		});

		const playerIconGroup = Input.newGroupContainer();
		playerIconGroup.classList.add('player-icon-group', 'icon-group');
		contentBlock.bodyElement.appendChild(playerIconGroup);

		this.configureIconSection(
			playerIconGroup,
			this.simUI.individualConfig.playerIconInputs.map(iconInput => IconInputs.buildIconInput(playerIconGroup, this.simUI.player, iconInput)),
			true,
		);

		// Level is global: every spec has it, whatever its own inputs are. The extra talent points
		// are in the Talents tab.
		new LevelPicker(contentBlock.bodyElement, this.simUI.player);
		new RacePicker(contentBlock.bodyElement, this.simUI.player);
		new RacialsPicker(contentBlock.bodyElement, this.simUI.player);

		if (this.simUI.individualConfig.playerInputs?.inputs.length) {
			this.configureInputSection(contentBlock.bodyElement, this.simUI.individualConfig.playerInputs);
		}

		new ProfessionsPicker(contentBlock.bodyElement, this.simUI.player);
	}

	private buildCustomSettingsSections() {
		(this.simUI.individualConfig.customSections || []).forEach(customSection => {
			const section = customSection(this.simUI.isWithinRaidSim ? this.column2 : this.column1, this.simUI);
			section.rootElem.classList.add('custom-section');
		});
	}

	private buildConsumesSection() {
		const column = this.simUI.isWithinRaidSim ? this.column3 : this.column1;
		const contentBlock = new ContentBlock(column, 'consumes-settings', {
			header: { title: 'Consumables' },
		});

		new ConsumesPicker(contentBlock.bodyElement, this.simUI);
	}

	private buildOtherSettings() {
		// Level is in the Player section and the extra talent points are in the Talents tab. In the
		// individual sim, In Front of Target is in the Encounter section.
		const elsewhere: InputConfig<Player<any>>[] = [OtherInputs.Level, OtherInputs.BonusTalentPoints];
		if (!this.simUI.isWithinRaidSim) elsewhere.push(OtherInputs.InFrontOfTarget);
		const otherInputs: InputSection = {
			...this.simUI.individualConfig.otherInputs,
			inputs: this.simUI.individualConfig.otherInputs.inputs.filter(input => !elsewhere.includes(input)),
		};
		const settings = otherInputs.inputs.filter(inputs => !inputs.extraCssClasses || !inputs.extraCssClasses?.includes('within-raid-sim-hide'));

		const itemSwapConfig = this.simUI.individualConfig.itemSwapConfig;

		if (settings.length || itemSwapConfig?.itemSlots.length) {
			const contentBlock = new ContentBlock(this.simUI.isWithinRaidSim ? this.column2 : this.column1, 'other-settings', {
				header: { title: 'Class Settings' },
			});

			if (settings.length) {
				this.configureInputSection(contentBlock.bodyElement, otherInputs);
				contentBlock.bodyElement.querySelectorAll('.input-root').forEach(elem => {
					elem.classList.add('input-inline');
				});
			}

			if (itemSwapConfig?.itemSlots.length) {
				new ItemSwapPicker(contentBlock.bodyElement, this.simUI, this.simUI.player, itemSwapConfig);
			}
		}
	}

	private buildIsbSettings() {
		if (!this.simUI.isWithinRaidSim) {
			const contentBlock = new ContentBlock(this.column2, 'other-settings', {
				header: { title: 'Improved Shadow Bolt' },
			});

			this.configureInputSection(contentBlock.bodyElement, IsbConfig);

			// It shows only with the debuff. We check right away too, because the debuffs are
			// built before it.
			const update = () => {
				const isWlAndIsb = (this.simUI.player as Player<Spec.SpecWarlock>)?.getTalents().improvedShadowBolt > 0;
				const externalIsb = this.simUI.player.getRaid()?.getDebuffs()?.improvedShadowBolt == true;
				contentBlock.rootElem.classList.toggle('hide', !(externalIsb || isWlAndIsb));
			};
			TypedEvent.onAny([this.simUI.player.talentsChangeEmitter, this.simUI.player.getRaid()!.debuffsChangeEmitter]).on(update);
			update();
		}
	}

	private buildStormstrikeSettings() {
		if (!this.simUI.isWithinRaidSim) {
			const contentBlock = new ContentBlock(this.column2, 'other-settings', {
				header: { title: 'Stormstrike' },
			});

			this.configureInputSection(contentBlock.bodyElement, StormstrikeConfig);

			// It shows only with the debuff. We check right away too, because the debuffs are
			// built before it.
			const update = () => contentBlock.rootElem.classList.toggle('hide', !this.simUI.player.getRaid()?.getDebuffs()?.stormstrike);
			this.simUI.player.getRaid()!.debuffsChangeEmitter.on(update);
			update();
		}
	}

	// One of the two buff sections, with the buffs that the party or the raid gives us, see
	// buffSource.
	//
	// The misc buffs go in a dropdown under the others, except for Innervate and Power Infusion.
	// They are buffs of their own, so they get an icon like the rest of the raid buffs.
	private buildBuffsSection(column: HTMLElement, cssClass: string, title: string, tooltip: string, source: BuffSource) {
		const fromSource = (options: { config: unknown }) => buffSource(options.config) === source;
		const miscOptions = relevantStatOptions(BuffDebuffInputs.MISC_BUFFS_CONFIG, this.simUI).filter(fromSource);
		const buffOptions = [
			...relevantStatOptions(BuffDebuffInputs.RAID_BUFFS_CONFIG, this.simUI).filter(fromSource),
			...miscOptions.filter(options => RAID_BUFFS.includes(options.config)),
		];
		const miscBuffOptions = miscOptions.filter(options => !RAID_BUFFS.includes(options.config));
		if (!buffOptions.length && !miscBuffOptions.length) return;

		const contentBlock = new ContentBlock(column, cssClass, {
			header: { title, tooltip },
		});
		contentBlock.rootElem.classList.add('buffs-section');

		this.configureIconSection(
			contentBlock.bodyElement,
			buffOptions.map(options => options.picker && new options.picker(contentBlock.bodyElement, this.simUI.player, options.config as any, this.simUI)),
		);

		if (miscBuffOptions.length) {
			new MultiIconPicker(
				contentBlock.bodyElement,
				this.simUI.player,
				{
					values: miscBuffOptions.map(options => options.config) as Array<MultiIconPickerItemConfig<Player<Spec>>>,
					label: 'Misc Buffs',
				},
				this.simUI,
			);
		}
	}

	// The world buffs are off for now. Forever has no world buffs that we know of, and it may have
	// something else in their place, like buffs from a camp. We show the section dimmed, with a
	// note, and Player.setBuffs turns them off, so saved settings or an import with them don't
	// count them either.
	private buildWorldBuffsSettings() {
		const contentBlock = new ContentBlock(this.column3, 'world-buffs-settings', {
			header: { title: 'World Buffs / Camp', tooltip: Tooltips.WORLD_BUFFS_SECTION },
		});
		contentBlock.rootElem.classList.add('buffs-section', 'world-buffs-off');

		const note = document.createElement('p');
		note.classList.add('world-buffs-note');
		note.textContent =
			"Off for now. Forever has no world buffs that we know of. It may have camp buffs or something like them in their place, and they'd go here.";
		contentBlock.rootElem.insertBefore(note, contentBlock.bodyElement);
		contentBlock.bodyElement.inert = true;

		const saygesOptions = relevantStatOptions(BuffDebuffInputs.SAYGES_CONFIG, this.simUI);
		new IconEnumPicker(contentBlock.bodyElement, this.simUI.player, BuffDebuffInputs.SaygesDarkFortune(saygesOptions));

		const worldBuffOptions = relevantStatOptions(BuffDebuffInputs.WORLD_BUFFS_CONFIG, this.simUI);
		this.configureIconSection(
			contentBlock.bodyElement,
			worldBuffOptions.map(
				options => options.picker && new options.picker(contentBlock.bodyElement, this.simUI.player, options.config as any, this.simUI),
			),
		);
	}

	private buildDebuffsSettings() {
		const debuffOptions = relevantStatOptions(BuffDebuffInputs.DEBUFFS_CONFIG, this.simUI);
		const miscDebuffOptions = relevantStatOptions(BuffDebuffInputs.MISC_DEBUFFS_CONFIG, this.simUI);

		if (!debuffOptions.length && !miscDebuffOptions.length) return;

		const contentBlock = new ContentBlock(this.column2, 'debuffs-settings', {
			header: { title: 'Debuffs', tooltip: Tooltips.DEBUFFS_SECTION },
		});

		this.configureIconSection(
			contentBlock.bodyElement,
			debuffOptions.map(options => options.picker && new options.picker(contentBlock.bodyElement, this.simUI.player, options.config as any, this.simUI)),
		);

		if (miscDebuffOptions.length) {
			new MultiIconPicker(
				contentBlock.bodyElement,
				this.simUI.player,
				{
					values: miscDebuffOptions.map(options => options.config) as Array<MultiIconPickerItemConfig<Player<Spec>>>,
					label: 'Misc Debuffs',
				},
				this.simUI,
			);
		}

		// In case no debuffs are active, this will fire a change event to update the pickers
		this.simUI.player.getRaid()?.debuffsChangeEmitter.emit(TypedEvent.nextEventID());
	}

	private buildSavedDataPickers() {
		// The saved encounters are part of the Encounter section, like the stat weight presets
		// in the Stat Weights section.
		const encounterPresets = this.encounterPresets!;
		tippy(encounterPresets.querySelector('span')!, {
			content: presetListTooltip('Loading an encounter changes:', ['Duration and its variation', 'Execute phases', 'PvP options', 'Targets']),
		});
		const savedEncounterManager = new SavedDataManager<Encounter, SavedEncounter>(
			encounterPresets.querySelector('.saved-data-presets') as HTMLElement,
			this.simUI.sim.encounter,
			{
				label: 'Encounter',
				storageKey: this.simUI.getSavedEncounterStorageKey(),
				getData: (encounter: Encounter) => SavedEncounter.create({ encounter: encounter.toProto() }),
				setData: (eventID: EventID, encounter: Encounter, newEncounter: SavedEncounter) => encounter.fromProto(eventID, newEncounter.encounter!),
				changeEmitters: [this.simUI.sim.encounter.changeEmitter],
				equals: (a: SavedEncounter, b: SavedEncounter) => SavedEncounter.equals(a, b),
				toJson: (a: SavedEncounter) => SavedEncounter.toJson(a),
				fromJson: (obj: any) => SavedEncounter.fromJson(obj),
			},
		);

		const savedSettingsManager = new SavedDataManager<IndividualSimUI<any>, SavedSettings>(this.rightPanel, this.simUI, {
			label: 'Settings',
			header: {
				title: 'Saved Settings',
				tooltip: presetListTooltip('Loading saved settings changes:', [
					'Raid, party and world buffs',
					'Debuffs',
					'Consumables',
					'Race, level, extra talent points and professions',
					'Item swap',
					'Reaction time, channel clip delay, position and distance from the target',
					'Healing model',
				]),
			},
			storageKey: this.simUI.getSavedSettingsStorageKey(),
			getData: (simUI: IndividualSimUI<any>) => {
				const player = simUI.player;
				return SavedSettings.create({
					raidBuffs: simUI.sim.raid.getBuffs(),
					partyBuffs: player.getParty()?.getBuffs() || PartyBuffs.create(),
					playerBuffs: player.getBuffs(),
					debuffs: simUI.sim.raid.getDebuffs(),
					consumes: player.getConsumes(),
					race: player.getRace(),
					level: player.getLevel(),
					bonusTalentPoints: player.getBonusTalentPoints(),
					professions: player.getProfessions(),
					enableItemSwap: player.getEnableItemSwap(),
					itemSwap: player.getItemSwapGear().toProto(),
					reactionTimeMs: player.getReactionTime(),
					channelClipDelayMs: player.getChannelClipDelay(),
					inFrontOfTarget: player.getInFrontOfTarget(),
					distanceFromTarget: player.getDistanceFromTarget(),
					healingModel: player.getHealingModel(),
				});
			},
			setData: (eventID: EventID, simUI: IndividualSimUI<any>, newSettings: SavedSettings) => {
				TypedEvent.freezeAllAndDo(() => {
					simUI.sim.raid.setBuffs(eventID, newSettings.raidBuffs || RaidBuffs.create());
					simUI.sim.raid.setDebuffs(eventID, newSettings.debuffs || Debuffs.create());
					const party = simUI.player.getParty();
					if (party) {
						party.setBuffs(eventID, newSettings.partyBuffs || PartyBuffs.create());
					}
					simUI.player.setBuffs(eventID, newSettings.playerBuffs || IndividualBuffs.create());
					simUI.player.setConsumes(eventID, newSettings.consumes || Consumes.create());
					simUI.player.setRace(eventID, newSettings.race);
					simUI.player.setLevel(eventID, newSettings.level);
					simUI.player.setBonusTalentPoints(eventID, newSettings.bonusTalentPoints);
					simUI.player.setProfessions(eventID, newSettings.professions);
					simUI.player.setEnableItemSwap(eventID, newSettings.enableItemSwap);
					simUI.player.setItemSwapGear(eventID, simUI.sim.db.lookupItemSwap(newSettings.itemSwap || ItemSwap.create()));
					simUI.player.setReactionTime(eventID, newSettings.reactionTimeMs);
					simUI.player.setChannelClipDelay(eventID, newSettings.channelClipDelayMs);
					simUI.player.setInFrontOfTarget(eventID, newSettings.inFrontOfTarget);
					simUI.player.setDistanceFromTarget(eventID, newSettings.distanceFromTarget);
					simUI.player.setHealingModel(eventID, newSettings.healingModel || HealingModel.create());
				});
			},
			changeEmitters: [
				this.simUI.sim.raid.buffsChangeEmitter,
				this.simUI.sim.raid.debuffsChangeEmitter,
				this.simUI.player.getParty()!.buffsChangeEmitter,
				this.simUI.player.buffsChangeEmitter,
				this.simUI.player.consumesChangeEmitter,
				this.simUI.player.raceChangeEmitter,
				this.simUI.player.professionChangeEmitter,
				this.simUI.player.itemSwapChangeEmitter,
				this.simUI.player.miscOptionsChangeEmitter,
				this.simUI.player.inFrontOfTargetChangeEmitter,
				this.simUI.player.distanceFromTargetChangeEmitter,
				this.simUI.player.healingModelChangeEmitter,
			],
			equals: (a: SavedSettings, b: SavedSettings) => SavedSettings.equals(a, b),
			toJson: (a: SavedSettings) => SavedSettings.toJson(a),
			fromJson: (obj: any) => SavedSettings.fromJson(obj),
		});

		this.simUI.sim.waitForInit().then(() => {
			savedEncounterManager.loadUserData();
			savedSettingsManager.loadUserData();
		});
	}

	private configureInputSection(sectionElem: HTMLElement, sectionConfig: InputSection) {
		sectionConfig.inputs.forEach(inputConfig => {
			if (inputConfig.type == 'number') {
				new NumberPicker(sectionElem, this.simUI.player, inputConfig);
			} else if (inputConfig.type == 'boolean') {
				new BooleanPicker(sectionElem, this.simUI.player, { ...inputConfig, reverse: true });
			} else if (inputConfig.type == 'enum' && inputConfig.values.some(value => value.icon)) {
				new IconEnumRowPicker(sectionElem, this.simUI.player, inputConfig);
			} else if (inputConfig.type == 'enum') {
				new EnumPicker(sectionElem, this.simUI.player, inputConfig);
			}
		});
	}

	private configureIconSection(sectionElem: HTMLElement, iconPickers: Array<any>, adjustColumns?: boolean) {
		if (iconPickers.length == 0) {
			sectionElem.classList.add('hide');
		} else if (adjustColumns) {
			if (iconPickers.length <= 4) {
				sectionElem.style.gridTemplateColumns = `repeat(${iconPickers.length}, 1fr)`;
			} else if (iconPickers.length > 4 && iconPickers.length < 8) {
				sectionElem.style.gridTemplateColumns = `repeat(${Math.ceil(iconPickers.length / 2)}, 1fr)`;
			}
		}
	}
}

// Who gives us a buff in the game: someone in our party, or anyone in the raid. It's about who
// casts it, not who it reaches, so Innervate is a raid buff even though it lands on us alone.
// The protos don't tell: they keep the totems and the auras with the raid buffs, but only our
// party's shaman or paladin gives them to us.
type BuffSource = 'party' | 'raid';

// The buffs anyone in the raid can give us, like Mark of the Wild or Fortitude. The Blessings
// go here too, even for a paladin, because another paladin in the raid gives them.
const RAID_BUFFS: unknown[] = [
	BuffDebuffInputs.BlessingOfKings,
	BuffDebuffInputs.BlessingOfMight,
	BuffDebuffInputs.BlessingOfWisdom,
	BuffDebuffInputs.AllStatsBuff,
	BuffDebuffInputs.StaminaBuff,
	BuffDebuffInputs.IntellectBuff,
	BuffDebuffInputs.SpiritBuff,
	BuffDebuffInputs.Thorns,
	BuffDebuffInputs.Innervate,
	BuffDebuffInputs.PowerInfusion,
];

// The rest come from our party: the totems, the auras, Battle Shout, Blood Pact and Atiesh.
function buffSource(config: unknown): BuffSource {
	return RAID_BUFFS.includes(config) ? 'raid' : 'party';
}

const PARTY_BUFFS_TOOLTIP = 'Buffs the members of our party give us, like totems, auras and Battle Shout.';
const RAID_BUFFS_TOOLTIP = 'Buffs anyone in the raid can give us, like Mark of the Wild, Fortitude, a Blessing or Innervate.';
