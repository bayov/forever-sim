import * as BuffDebuffInputs from '../core/components/inputs/buffs_debuffs';
import * as OtherInputs from '../core/components/other_inputs.js';
import { IndividualSimUI, registerSpecConfig } from '../core/individual_sim_ui.js';
import { Player } from '../core/player.js';
import { APLRotation_Type as APLRotationType } from '../core/proto/apl.js';
import { Class, Faction, PartyBuffs, PseudoStat, Race, Spec, Stat } from '../core/proto/common.js';
import { Stats } from '../core/proto_utils/stats.js';
import { getSpecIcon, specNames } from '../core/proto_utils/utils.js';
import * as DruidInputs from './inputs.js';
import * as Presets from './presets.js';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecFeralTankDruid, {
	cssClass: 'feral-tank-druid-sim-ui',
	cssScheme: 'druid',
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStamina,
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatAttackPower,
		Stat.StatFeralAttackPower,
		Stat.StatMeleeHit,
		Stat.StatMeleeCrit,
		Stat.StatArmor,
		Stat.StatBonusArmor,
		Stat.StatDefense,
		Stat.StatDodge,
		Stat.StatHealth,
	],
	epPseudoStats: [PseudoStat.PseudoStatMeleeSpeedMultiplier],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatAttackPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: [
		Stat.StatHealth,
		Stat.StatArmor,
		Stat.StatBonusArmor,
		Stat.StatStamina,
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatAttackPower,
		Stat.StatFeralAttackPower,
		Stat.StatMeleeHit,
		Stat.StatMeleeCrit,
		Stat.StatDefense,
		Stat.StatDodge,
	],
	displayPseudoStats: [PseudoStat.PseudoStatMeleeSpeedMultiplier],

	defaults: {
		race: Presets.OtherDefaults.race,
		// Default equipped gear.
		gear: Presets.DefaultGear.gear,
		// Default EP weights for sorting gear in the gear picker: the tank stat prices of standards 36
		// (1b's weights with Defense 1.5 a point, dodge 12 a 1%, bonus armor 0.1), in attack power.
		epWeights: Stats.fromMap(
			{
				[Stat.StatStamina]: 2,
				[Stat.StatStrength]: 2,
				[Stat.StatAgility]: 2,
				[Stat.StatAttackPower]: 1,
				[Stat.StatFeralAttackPower]: 1,
				[Stat.StatMeleeHit]: 20,
				[Stat.StatMeleeCrit]: 28,
				[Stat.StatBonusArmor]: 0.2,
				[Stat.StatDefense]: 3,
				[Stat.StatDodge]: 24,
			},
			{
				[PseudoStat.PseudoStatMeleeSpeedMultiplier]: 0,
			},
		),
		// Default consumes settings.
		consumes: Presets.DefaultConsumes,
		// Default rotation settings.
		rotationType: APLRotationType.TypeAPL,
		simpleRotation: Presets.DefaultRotation,
		// Default talents.
		talents: Presets.DefaultTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		encounter: Presets.EncounterLevel60Tank.encounter,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: PartyBuffs.create({}),
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Inputs to include in the 'Rotation' section on the settings tab.
	rotationInputs: DruidInputs.FeralTankDruidRotationConfig,
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [BuffDebuffInputs.SpellCritBuff, BuffDebuffInputs.SpellISBDebuff],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			OtherInputs.TankAssignment,
			OtherInputs.IncomingHps,
			OtherInputs.HealingCadence,
			OtherInputs.HealingCadenceVariation,
			OtherInputs.BurstWindow,
			OtherInputs.InspirationUptime,
			OtherInputs.HpPercentForDefensives,
			DruidInputs.StartingRage,
			OtherInputs.InFrontOfTarget,
		],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		builds: [Presets.PresetBuildMaxThreat, Presets.PresetBuildBalanced, Presets.PresetBuildMaxSurvival],
		// Preset talents that the user can quickly select.
		talents: Presets.TalentPresets,
		// Preset rotations that the user can quickly select.
		rotations: Presets.APLPresets,
		// Preset gear configurations that the user can quickly select.
		gear: Presets.GearPresets,
	},

	autoRotation: player => {
		return Presets.DefaultAPL.rotation.rotation!;
	},

	raidSimPresets: [
		{
			spec: Spec.SpecFeralTankDruid,
			tooltip: specNames[Spec.SpecFeralTankDruid],
			defaultName: 'Bear',
			iconUrl: getSpecIcon(Class.ClassDruid, 1),

			talents: Presets.DefaultTalents.data,
			specOptions: Presets.DefaultOptions,
			consumes: Presets.DefaultConsumes,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceNightElf,
				[Faction.Horde]: Race.RaceSkyborneWindshaper,
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

export class FeralTankDruidSimUI extends IndividualSimUI<Spec.SpecFeralTankDruid> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecFeralTankDruid>) {
		super(parentElem, player, SPEC_CONFIG);
	}
}
