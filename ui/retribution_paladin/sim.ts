import * as BuffDebuffInputs from '../core/components/inputs/buffs_debuffs';
import * as OtherInputs from '../core/components/other_inputs.js';
import { Phase } from '../core/constants/other.js';
import { IndividualSimUI, registerSpecConfig } from '../core/individual_sim_ui.js';
import { Player } from '../core/player.js';
import { Class, Faction, PartyBuffs, PseudoStat, Race, Spec, Stat } from '../core/proto/common.js';
import { Stats } from '../core/proto_utils/stats.js';
import { getSpecIcon } from '../core/proto_utils/utils.js';
import * as RetributionPaladinInputs from './inputs.js';
import * as Presets from './presets.js';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecRetributionPaladin, {
	cssClass: 'retribution-paladin-sim-ui',
	cssScheme: 'paladin',
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [
		`Forever rules: a Judgement does not consume the seal, Judgement of the Crusader lasts 40 sec and melee hits refresh it,
		Holy Strike is baseline, Consecration is baseline and reworked, and the talent trees are the beta client's. Twist of Light,
		Seal of Fury and Light's Vigil are not simulated.`,
		`Be aware that not all item and weapon enchants are currently implemented in the sim, which make some notable Retribution
		weapons like Pendulum of Doom and The Jackhammer undervalued.`,
	],
	warnings: [
		(simUI: IndividualSimUI<Spec.SpecRetributionPaladin>) => {
			return {
				updateOn: simUI.player.changeEmitter,
				getContent: () => {
					if (simUI.player.getSpecOptions().primarySeal == 0) {
						return `Your previously selected seal is no longer available because of a talent or rune change.
							No seal will be cast with this configuration. Please select an available seal in the Settings>Player menu.`;
					} else {
						return '';
					}
				},
			};
		},
	],
	// All stats for which EP should be calculated.
	epStats: [
		// Attributes
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatIntellect,
		// Physical
		Stat.StatAttackPower,
		Stat.StatMeleeHit,
		Stat.StatMeleeCrit,
		Stat.StatExpertise,
		// Spell
		Stat.StatSpellPower,
		Stat.StatHolyPower,
		Stat.StatSpellCrit,
		Stat.StatSpellHit,
		Stat.StatMP5,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatMeleeSpeedMultiplier],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatAttackPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: [
		// Primary
		Stat.StatMana,
		// Attributes
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatIntellect,
		// Physical
		Stat.StatAttackPower,
		Stat.StatMeleeHit,
		Stat.StatMeleeCrit,
		Stat.StatExpertise,
		// Spell
		Stat.StatSpellPower,
		Stat.StatHolyPower,
		Stat.StatSpellCrit,
		Stat.StatSpellHit,
		Stat.StatMP5,
	],
	displayPseudoStats: [PseudoStat.PseudoStatMeleeSpeedMultiplier],

	defaults: {
		// Default equipped gear.
		gear: Presets.DefaultGear.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Stats.fromMap(
			{
				[Stat.StatStrength]: 2.53,
				[Stat.StatAgility]: 1.13,
				[Stat.StatIntellect]: 0.15,
				[Stat.StatSpellPower]: 0.32,
				[Stat.StatSpellHit]: 0.41,
				[Stat.StatSpellCrit]: 0.01,
				[Stat.StatMP5]: 0.05,
				[Stat.StatAttackPower]: 1,
				[Stat.StatMeleeHit]: 1.96,
				[Stat.StatMeleeCrit]: 1.16,
				[Stat.StatFireResistance]: 0.5,
			},
			{
				[PseudoStat.PseudoStatMainHandDps]: 7.33,
				[PseudoStat.PseudoStatMeleeSpeedMultiplier]: 7.33,
			},
		),
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
		race: Race.RaceUndead,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [RetributionPaladinInputs.PrimarySealSelection, RetributionPaladinInputs.AuraSelection],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [BuffDebuffInputs.SpellScorchDebuff],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.TankAssignment, OtherInputs.InFrontOfTarget, RetributionPaladinInputs.JudgementStopAttack],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		builds: [Presets.PresetBuildLevel20, Presets.PresetBuildLevel20x2, Presets.PresetBuildLevel20x3, Presets.PresetBuildLevel20x4],
		rotations: [
			...Presets.APLPresets[Phase.Phase4],
			...Presets.APLPresets[Phase.Phase3],
			...Presets.APLPresets[Phase.Phase2],
			...Presets.APLPresets[Phase.Phase1],
		],
		// Preset talents that the user can quickly select.
		talents: [
			...Presets.TalentPresets[Phase.Phase4],
			...Presets.TalentPresets[Phase.Phase3],
			...Presets.TalentPresets[Phase.Phase2],
			...Presets.TalentPresets[Phase.Phase1],
		],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.DefaultGear, Presets.GearLevel20, Presets.GearLevel20AoE],
	},

	autoRotation: player => {
		return Presets.DefaultAPL.rotation.rotation!;
	},

	raidSimPresets: [
		{
			spec: Spec.SpecRetributionPaladin,
			tooltip: 'Retribution Paladin',
			defaultName: 'Retribution',
			iconUrl: getSpecIcon(Class.ClassPaladin, 2),

			talents: Presets.DefaultTalents.data,
			specOptions: Presets.DefaultOptions,
			consumes: Presets.DefaultConsumes,
			// Forever gives the Forsaken paladins, so the Horde has one now.
			defaultFactionRaces: {
				[Faction.Alliance]: Race.RaceHuman,
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Horde]: Race.RaceUndead,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.DefaultGear.gear,
					2: Presets.DefaultGear.gear,
				},
				[Faction.Horde]: {
					1: Presets.DefaultGear.gear,
					2: Presets.DefaultGear.gear,
				},
			},
		},
	],
});

export class RetributionPaladinSimUI extends IndividualSimUI<Spec.SpecRetributionPaladin> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecRetributionPaladin>) {
		super(parentElem, player, SPEC_CONFIG);
	}
}
