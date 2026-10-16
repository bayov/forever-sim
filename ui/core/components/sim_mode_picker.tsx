import tippy from 'tippy.js';

import type { IndividualSimUI } from '../individual_sim_ui';
import { Encounter as EncounterProto, LevelingOptions } from '../proto/common';
import { encounterMode, encounterProtoMode, SIM_MODES, SimMode, simModeNames } from '../sim_mode';
import { EventID, TypedEvent } from '../typed_event';

const modeTooltips: Record<SimMode, string> = {
	dps: 'A fight against a boss, scored by DPS. The Encounter settings pick the fight.',
	pvp: 'A fight against an enemy player, scored by DPS. White hits never glance, and Out of Melee (%) in the Encounter settings is the share of the fight we spend out of melee range.',
	leveling:
		'A grinding session against normal mobs around our level, scored by the mobs we kill. The Encounter settings have the options of the session.',
};

// The encounter of each mode, as we last left it.
const MODE_ENCOUNTERS_KEY = '__modeEncounters__';

// DPS, PvP or Leveling, in the sidebar above the iterations. See sim_mode.ts.
//
// Each mode keeps its own encounter. When we go from Leveling to DPS and back, the session's
// options are the ones we had, and the DPS fight is the one we had before we went to Leveling.
// A mode we never used starts from the encounter we have, with the mode turned on.
export class SimModePicker {
	readonly rootElem: HTMLElement;
	private readonly simUI: IndividualSimUI<any>;
	private readonly buttons = new Map<SimMode, HTMLButtonElement>();
	private encounters: Partial<Record<SimMode, unknown>>;

	constructor(parent: HTMLElement, simUI: IndividualSimUI<any>) {
		this.simUI = simUI;
		this.encounters = this.load();

		this.rootElem = (<div className="sim-mode-picker segmented" attributes={{ role: 'group', 'aria-label': 'Mode' }} />) as HTMLElement;
		SIM_MODES.forEach(mode => {
			const button = (
				<button className="segmented-option" type="button" dataset={{ mode }}>
					{simModeNames[mode]}
				</button>
			) as HTMLButtonElement;
			button.addEventListener('click', () => this.setMode(TypedEvent.nextEventID(), mode));
			tippy(button, { content: modeTooltips[mode], placement: 'right' });
			this.buttons.set(mode, button);
			this.rootElem.appendChild(button);
		});
		parent.prepend(this.rootElem);

		const encounter = simUI.sim.encounter;
		const update = () => {
			const mode = encounterMode(encounter);
			this.buttons.forEach((button, buttonMode) => {
				button.classList.toggle('active', buttonMode === mode);
				button.setAttribute('aria-pressed', String(buttonMode === mode));
			});
			this.encounters[mode] = EncounterProto.toJson(encounter.toProto());
			this.store();
		};
		simUI.sim.waitForInit().then(() => {
			update();
			encounter.changeEmitter.on(update);
		});
	}

	private setMode(eventID: EventID, mode: SimMode) {
		const encounter = this.simUI.sim.encounter;
		if (encounterMode(encounter) === mode) return;

		let saved: EncounterProto | null = null;
		try {
			const json = this.encounters[mode];
			if (json) saved = EncounterProto.fromJson(json as any, { ignoreUnknownFields: true });
		} catch (e) {
			console.warn(`Failed to read the encounter of ${simModeNames[mode]} mode: ` + e);
		}

		TypedEvent.freezeAllAndDo(() => {
			if (saved && encounterProtoMode(saved) === mode) {
				encounter.fromProto(eventID, saved);
				return;
			}
			encounter.setPvp(eventID, mode === 'pvp');
			// A new session starts with every option at its default. The Out of Melee field is
			// hidden in Leveling mode, where the session moves us itself, so we clear it rather
			// than keep a downtime nobody sees.
			if (mode === 'leveling') {
				encounter.setLeveling(eventID, LevelingOptions.create());
				encounter.setPvpMeleeDowntime(eventID, 0);
			} else {
				encounter.setLeveling(eventID, undefined);
			}
		});
	}

	private load(): Partial<Record<SimMode, unknown>> {
		try {
			const json = window.localStorage.getItem(this.simUI.getStorageKey(MODE_ENCOUNTERS_KEY));
			const encounters = json ? JSON.parse(json) : {};
			return encounters && typeof encounters === 'object' ? encounters : {};
		} catch (e) {
			console.warn('Failed to load the encounter of each mode: ' + e);
			return {};
		}
	}

	private store() {
		try {
			window.localStorage.setItem(this.simUI.getStorageKey(MODE_ENCOUNTERS_KEY), JSON.stringify(this.encounters));
		} catch (e) {
			console.warn('Failed to save the encounter of each mode: ' + e);
		}
	}
}

