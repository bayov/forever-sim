import * as OtherInputs from '../core/components/other_inputs';
import { IndividualSimUI, registerSpecConfig } from '../core/individual_sim_ui';
import { Player } from '../core/player';
import { Class, Faction, PartyBuffs, Race, Spec, Stat } from '../core/proto/common';
import { Stats } from '../core/proto_utils/stats';
import { getSpecIcon } from '../core/proto_utils/utils';
import * as Presets from './presets';

// The Holy and Discipline priest of the healing sim (docs/healing_priest.md), set up as the Restoration
// shaman's page is: its numbers come from tools/healbench.
const SPEC_CONFIG = registerSpecConfig(Spec.SpecHealingPriest, {
	cssClass: 'healing-priest-sim-ui',
	cssScheme: 'priest',
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
			[Stat.StatIntellect]: 0.3,
			[Stat.StatSpirit]: 0.2,
			[Stat.StatSpellPower]: 1,
			[Stat.StatHealingPower]: 1,
			[Stat.StatSpellCrit]: 0.5,
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
	// IconInputs to include in the 'Player' section on the settings tab. Power Infusion and Inner Focus are
	// cast by the rotation; Inner Fire isn't simulated.
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.DistanceFromTarget],
	},
	customSections: [],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		builds: [Presets.PresetBuildHolyTankHealer, Presets.PresetBuildHolyRaidHealer, Presets.PresetBuildDisciplineTankHealer],
		// Preset talents that the user can quickly select.
		talents: Presets.TalentPresets,
		// Preset rotations that the user can quickly select.
		rotations: Presets.APLPresets,
		// Preset gear configurations that the user can quickly select.
		gear: Presets.GearPresets,
	},

	autoRotation: (player: Player<Spec.SpecHealingPriest>) => {
		// Discipline (the first tree deepest) shields the tank; Holy heals it.
		return (player.getTalentTree() == 0 ? Presets.APLDisciplineTankHealer : Presets.DefaultAPL).rotation.rotation!;
	},

	raidSimPresets: [
		{
			spec: Spec.SpecHealingPriest,
			tooltip: 'Holy Priest',
			defaultName: 'Holy',
			iconUrl: getSpecIcon(Class.ClassPriest, 1),

			talents: Presets.TalentsHoly.data,
			specOptions: Presets.DefaultOptions,
			consumes: Presets.DefaultConsumes,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Presets.AllianceRace,
				[Faction.Horde]: Presets.HordeRace,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.DefaultGear.gear,
				},
				[Faction.Horde]: {
					1: Presets.DefaultGear.gear,
				},
			},
		},
		{
			spec: Spec.SpecHealingPriest,
			tooltip: 'Discipline Priest',
			defaultName: 'Discipline',
			iconUrl: getSpecIcon(Class.ClassPriest, 0),

			talents: Presets.TalentsDiscipline.data,
			specOptions: Presets.DefaultOptions,
			consumes: Presets.DefaultConsumes,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Presets.AllianceRace,
				[Faction.Horde]: Presets.HordeRace,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.DefaultGear.gear,
				},
				[Faction.Horde]: {
					1: Presets.DefaultGear.gear,
				},
			},
		},
	],
});

export class HealingPriestSimUI extends IndividualSimUI<Spec.SpecHealingPriest> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecHealingPriest>) {
		super(parentElem, player, SPEC_CONFIG);
	}
}
