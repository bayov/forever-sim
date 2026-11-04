import * as OtherInputs from '../core/components/other_inputs.js';
import { IndividualSimUI, registerSpecConfig } from '../core/individual_sim_ui.js';
import { Player } from '../core/player.js';
import { Class, Faction, PartyBuffs, Race, Spec, Stat } from '../core/proto/common.js';
import { Stats } from '../core/proto_utils/stats.js';
import { getSpecIcon, specNames } from '../core/proto_utils/utils.js';
import * as HolyPaladinInputs from './inputs.js';
import * as Presets from './presets.js';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecHolyPaladin, {
	cssClass: 'holy-paladin-sim-ui',
	cssScheme: 'paladin',
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [],

	// All stats for which EP should be calculated.
	epStats: [Stat.StatIntellect, Stat.StatSpirit, Stat.StatSpellPower, Stat.StatHealingPower, Stat.StatSpellCrit, Stat.StatSpellHaste, Stat.StatMP5],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatSpellPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: [
		Stat.StatHealth,
		Stat.StatMana,
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatSpellPower,
		Stat.StatHealingPower,
		Stat.StatSpellCrit,
		Stat.StatSpellHaste,
		Stat.StatMP5,
	],
	displayPseudoStats: [],

	defaults: {
		race: Presets.OtherDefaults.race,
		// Default equipped gear.
		gear: Presets.DefaultGear.gear,
		// Default EP weights for sorting gear in the gear picker. Not measured: a healer's EP needs a
		// healing score, which the individual sim doesn't have yet (docs/healing_sim.md).
		epWeights: Stats.fromMap({
			[Stat.StatIntellect]: 0.4,
			[Stat.StatSpirit]: 0.1,
			[Stat.StatSpellPower]: 1,
			[Stat.StatHealingPower]: 1,
			[Stat.StatSpellCrit]: 0.6,
			[Stat.StatSpellHaste]: 0.5,
			[Stat.StatMP5]: 1,
		}),
		// Default consumes settings.
		consumes: Presets.DefaultConsumes,
		// Default talents.
		talents: Presets.DefaultTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: PartyBuffs.create({}),
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [HolyPaladinInputs.AuraSelection, OtherInputs.DistanceFromTarget],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		builds: [Presets.PresetBuildTankHealer, Presets.PresetBuildRaidHealer],
		// Preset talents that the user can quickly select.
		talents: Presets.TalentPresets,
		// Preset rotations that the user can quickly select.
		rotations: Presets.APLPresets,
		// Preset gear configurations that the user can quickly select.
		gear: Presets.GearPresets,
	},

	autoRotation: () => {
		return Presets.DefaultAPL.rotation.rotation!;
	},

	// The raid sim's placeholder Holy paladin keeps its talents and blank gear (sim/compsearch/presets.go
	// and ui/raid_sim/data mirror them); the healers there come with H6 (docs/healing_sim.md).
	raidSimPresets: [
		{
			spec: Spec.SpecHolyPaladin,
			tooltip: specNames[Spec.SpecHolyPaladin],
			defaultName: 'Holy',
			iconUrl: getSpecIcon(Class.ClassPaladin, 0),

			talents: Presets.StandardTalents.data,
			specOptions: Presets.RaidSimOptions,
			consumes: Presets.RaidSimConsumes,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceHuman,
				[Faction.Horde]: Race.RaceUnknown,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.GearBlank.gear,
				},
				[Faction.Horde]: {
					1: Presets.GearBlank.gear,
				},
			},
		},
	],
});

export class HolyPaladinSimUI extends IndividualSimUI<Spec.SpecHolyPaladin> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecHolyPaladin>) {
		super(parentElem, player, SPEC_CONFIG);
	}
}
