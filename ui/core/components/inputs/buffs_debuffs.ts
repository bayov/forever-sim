import { Player } from '../../player';
import { Ruleset } from '../../proto/api';
import { Debuffs, Faction, RaidBuffs, SaygesFortune, Stat, TotemWeaponBuff, TristateEffect } from '../../proto/common';
import { ActionId } from '../../proto_utils/action_id';
import { EventID, TypedEvent } from '../../typed_event';
import {
	makeBooleanDebuffInput,
	makeBooleanIndividualBuffInput,
	makeBooleanRaidBuffInput,
	makeEnumIndividualBuffInput,
	makeEnumRaidBuffInput,
	makeMultistateIndividualBuffInput,
	makeMultistatePartyBuffInput,
	makeMultistateRaidBuffInput,
	makeTristateDebuffInput,
	makeTristateIndividualBuffInput,
	makeTristateRaidBuffInput,
	withLabel,
} from '../icon_inputs';
import { IconEnumPicker } from '../icon_enum_picker';
import { ExclusiveDebuffRowConfig, DebuffToggleField, isDebuffOn, setDebuffOn } from '../exclusive_debuff_row';
import { IconPicker, IconPickerConfig, IconPickerDirection } from '../icon_picker';
import * as InputHelpers from '../input_helpers';
import { MultiIconPicker } from '../multi_icon_picker';
import { ItemStatOption, PickerStatOptions } from './stat_options';

///////////////////////////////////////////////////////////////////////////
//                                 RAID BUFFS
///////////////////////////////////////////////////////////////////////////

export const AllStatsBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(9885),
		impId: ActionId.fromSpellId(17055),
		fieldName: 'giftOfTheWild',
	}),
	'Mark of the Wild',
);

// Separate Strength buffs allow us to use a boolean pickers for Horde specifically
export const BlessingOfKings = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(20217),
		fieldName: 'blessingOfKings',
		showWhen: player => player.hasFactionBuffs(Faction.Alliance),
	}),
	'Blessing of Kings',
);

export const ArmorBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10293),
		impId: ActionId.fromSpellId(20142),
		showWhen: player => player.hasFactionBuffs(Faction.Alliance),
		fieldName: 'devotionAura',
	}),
	'Devotion Aura',
);

export const PhysDamReductionBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10408),
		impId: ActionId.fromSpellId(16293),
		showWhen: player => player.hasFactionBuffs(Faction.Horde),
		fieldName: 'stoneskinTotem',
	}),
	'Stoneskin',
);

//export const DamageReductionPercentBuff = withLabel(
//	makeBooleanIndividualBuffInput({
//		actionId: player =>
//			player.getMatchingSpellActionId([
//				{ id: 20911, minLevel: 30, maxLevel: 39 },
//				{ id: 20912, minLevel: 40, maxLevel: 49 },
//				{ id: 20913, minLevel: 50, maxLevel: 59 },
//				{ id: 20914, minLevel: 60 },
//			]),
//		showWhen: player => player.hasFactionBuffs(Faction.Alliance),
//		fieldName: 'blessingOfSanctuary',
//	}),
//	'Blessing of Sanctuary',
//);

export const ResistanceBuff = InputHelpers.makeMultiIconInput({
	values: [
		// Shadow
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(10958),
			fieldName: 'shadowProtection',
		}),
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(19896),
			fieldName: 'shadowResistanceAura',
		}),
		// Nature
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(10601),
			fieldName: 'natureResistanceTotem',
			showWhen: player => player.hasFactionBuffs(Faction.Horde),
		}),
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(20190),
			fieldName: 'aspectOfTheWild',
		}),
		// Fire
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(19900),
			fieldName: 'fireResistanceAura',
			showWhen: player => player.hasFactionBuffs(Faction.Alliance),
		}),
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(10538),
			fieldName: 'fireResistanceTotem',
			showWhen: player => player.hasFactionBuffs(Faction.Horde),
		}),
		// Frost
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(19898),
			fieldName: 'frostResistanceAura',
			showWhen: player => player.hasFactionBuffs(Faction.Alliance),
		}),
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(10479),
			fieldName: 'frostResistanceTotem',
			showWhen: player => player.hasFactionBuffs(Faction.Horde),
		}),
	],
	label: 'Resistances',
});

// The Scrolls of Stamina, Intellect and Spirit are with the consumables, see ScrollOfStamina.
export const StaminaBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10938),
		impId: ActionId.fromSpellId(14767),
		fieldName: 'powerWordFortitude',
	}),
	'Power Word: Fortitude',
);

export const BloodPactBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(11767),
		impId: ActionId.fromSpellId(18696),
		fieldName: 'bloodPact',
	}),
	'Blood Pact',
);

export const BlessingOfMight = withLabel(
	makeTristateIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(25291),
		impId: ActionId.fromSpellId(20048),
		fieldName: 'blessingOfMight',
		showWhen: player => player.hasFactionBuffs(Faction.Alliance),
	}),
	'Blessing of Might',
);

export const StrengthBuffHorde = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(25361),
		impId: ActionId.fromSpellId(16295),
		fieldName: 'strengthOfEarthTotem',
		showWhen: player => player.hasFactionBuffs(Faction.Horde) && player.sim.getRuleset() !== Ruleset.RulesetForever,
	}),
	'Strength',
);

export const GraceOfAir = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(25359),
		impId: ActionId.fromSpellId(16295),
		fieldName: 'graceOfAirTotem',
		showWhen: player => player.hasFactionBuffs(Faction.Horde) && player.sim.getRuleset() !== Ruleset.RulesetForever,
	}),
	'Agility',
);

// The improved Strength of Earth and Grace of Air come from Enhancing Totems, which is not in
// Forever's tree. So under Forever we show these totems as on or off. An improved value from an
// older saved setup counts as on, and the sim treats it as a regular totem.
function makeForeverHordeTotemInput(spellId: number, fieldName: 'strengthOfEarthTotem' | 'graceOfAirTotem') {
	return InputHelpers.makeBooleanIconInput<any, RaidBuffs, Player<any>>(
		{
			getModObject: (player: Player<any>) => player,
			showWhen: (player: Player<any>) => player.hasFactionBuffs(Faction.Horde) && player.sim.getRuleset() === Ruleset.RulesetForever,
			getValue: (player: Player<any>) => player.getRaid()!.getBuffs(),
			setValue: (eventID: EventID, player: Player<any>, newVal: RaidBuffs) => player.getRaid()!.setBuffs(eventID, newVal),
			changeEmitter: (player: Player<any>) =>
				TypedEvent.onAny([player.getRaid()!.buffsChangeEmitter, player.raceChangeEmitter, player.sim.rulesetChangeEmitter]),
			getFieldValue: (player: Player<any>) => player.getRaid()!.getBuffs()[fieldName] !== TristateEffect.TristateEffectMissing,
			setFieldValue: (eventID: EventID, player: Player<any>, newValue: boolean) => {
				const buffs = player.getRaid()!.getBuffs();
				buffs[fieldName] = newValue ? TristateEffect.TristateEffectRegular : TristateEffect.TristateEffectMissing;
				player.getRaid()!.setBuffs(eventID, buffs);
			},
		},
		() => ActionId.fromSpellId(spellId),
		fieldName,
	);
}

export const StrengthBuffHordeForever = withLabel(makeForeverHordeTotemInput(25361, 'strengthOfEarthTotem'), 'Strength');

export const GraceOfAirForever = withLabel(makeForeverHordeTotemInput(25359, 'graceOfAirTotem'), 'Agility');

// Another shaman's Windfury or Flametongue Totem. Under Forever its buff has a weapon slot
// of its own, so it works next to a shaman imbue and an oil or stone. Only one of the two
// totems works at a time.
export const TotemWeaponBuffInput = withLabel(
	makeEnumRaidBuffInput({
		values: [
			{ value: TotemWeaponBuff.TotemWeaponBuffNone },
			{ actionId: () => ActionId.fromSpellId(10614), value: TotemWeaponBuff.TotemWeaponBuffWindfury },
			{ actionId: () => ActionId.fromSpellId(16387), value: TotemWeaponBuff.TotemWeaponBuffFlametongue },
		],
		fieldName: 'totemWeaponBuff',
		showWhen: player => player.hasFactionBuffs(Faction.Horde),
	}),
	'Totem Weapon Buff',
);

export const IntellectBuff = withLabel(
	makeBooleanRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10157),
		fieldName: 'arcaneBrilliance',
	}),
	'Arcane Intellect',
);

export const SpiritBuff = withLabel(
	makeBooleanRaidBuffInput({
		actionId: () => ActionId.fromSpellId(27841),
		fieldName: 'divineSpirit',
	}),
	'Divine Spirit',
);

export const BattleShoutBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(25289),
		impId: ActionId.fromSpellId(12861),
		fieldName: 'battleShout',
	}),
	'Battle Shout',
);

export const TrueshotAuraBuff = withLabel(
	makeBooleanRaidBuffInput({
		actionId: () => ActionId.fromSpellId(20906),
		fieldName: 'trueshotAura',
	}),
	'Trueshot Aura',
);

export const BlessingOfWisdom = withLabel(
	makeTristateIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(25290),
		impId: ActionId.fromSpellId(20245),
		fieldName: 'blessingOfWisdom',
		showWhen: player => player.hasFactionBuffs(Faction.Alliance),
	}),
	'Blessing of Wisdom',
);
export const ManaSpringTotem = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10497),
		impId: ActionId.fromSpellId(16208),
		fieldName: 'manaSpringTotem',
		showWhen: player => player.hasFactionBuffs(Faction.Horde),
	}),
	'Mana Spring Totem',
);

export const MeleeCritBuff = withLabel(
	makeBooleanRaidBuffInput({ actionId: () => ActionId.fromSpellId(24932), fieldName: 'leaderOfThePack' }),
	'Leader of the Pack',
);

export const SpellCritBuff = withLabel(makeBooleanRaidBuffInput({ actionId: () => ActionId.fromSpellId(24907), fieldName: 'moonkinAura' }), 'Moonkin Aura');

// Misc Buffs
export const AtieshMageBuff = makeMultistatePartyBuffInput({
	actionId: () => ActionId.fromSpellId(28142),
	fieldName: 'atieshMage',
	numStates: 5,
});
export const AtieshWarlockBuff = makeMultistatePartyBuffInput({
	actionId: () => ActionId.fromSpellId(28143),
	fieldName: 'atieshWarlock',
	numStates: 5,
});
export const AtieshPriestBuff = makeMultistatePartyBuffInput({
	actionId: () => ActionId.fromSpellId(28144),
	fieldName: 'atieshPriest',
	numStates: 5,
});
export const AtieshDruidBuff = makeMultistatePartyBuffInput({
	actionId: () => ActionId.fromSpellId(28145),
	fieldName: 'atieshDruid',
	numStates: 5,
});

export const RetributionAura = makeTristateRaidBuffInput({
	actionId: () => ActionId.fromSpellId(10301),
	impId: ActionId.fromSpellId(20092),
	fieldName: 'retributionAura',
	showWhen: player => player.hasFactionBuffs(Faction.Alliance),
});

// Sanctity Aura is not in Forever's Retribution tree, the sim ignores it there.
export const SanctityAura = makeBooleanRaidBuffInput({
	actionId: () => ActionId.fromSpellId(20218),
	fieldName: 'sanctityAura',
	showWhen: player => player.hasFactionBuffs(Faction.Alliance) && player.sim.getRuleset() !== Ruleset.RulesetForever,
});

export const Thorns = makeTristateRaidBuffInput({
	actionId: () => ActionId.fromSpellId(9910),
	impId: ActionId.fromSpellId(16840),
	fieldName: 'thorns',
});

export const Innervate = withLabel(
	makeMultistateIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(29166),
		numStates: 11,
		fieldName: 'innervates',
	}),
	'Innervate',
);

export const PowerInfusion = withLabel(
	makeMultistateIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(10060),
		numStates: 11,
		fieldName: 'powerInfusions',
	}),
	'Power Infusion',
);

export const BattleSquawkBuff = makeMultistateRaidBuffInput({
	actionId: () => ActionId.fromSpellId(23060),
	numStates: 6,
	fieldName: 'battleSquawk',
});

///////////////////////////////////////////////////////////////////////////
//                                 WORLD BUFFS
///////////////////////////////////////////////////////////////////////////

export const RallyingCryOfTheDragonslayer = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(22888),
		fieldName: 'rallyingCryOfTheDragonslayer',
	}),
	'Rallying Cry Of The Dragonslayer',
);
export const SpiritOfZandalar = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(24425),
		fieldName: 'spiritOfZandalar',
	}),
	'Spirit of Zandalar',
);
export const SongflowerSerenade = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(15366),
		fieldName: 'songflowerSerenade',
	}),
	'Songflower Serenade',
);
export const WarchiefsBlessing = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(16609),
		fieldName: 'warchiefsBlessing',
		// showWhen: player => player.hasFactionBuffs(Faction.Horde),
	}),
	`Warchief's Blessing`,
);

export const SaygesDarkFortune = (inputs: ItemStatOption<SaygesFortune>[]) =>
	makeEnumIndividualBuffInput({
		direction: IconPickerDirection.Horizontal,
		values: [
			{ iconUrl: 'https://wow.zamimg.com/images/wow/icons/large/inv_misc_orb_02.jpg', value: SaygesFortune.SaygesUnknown, text: `Sayge's Dark Fortune` },
			...inputs.map(input => input.config),
		],
		fieldName: 'saygesFortune',
	});

export const SaygesDamage = { actionId: () => ActionId.fromSpellId(23768), value: SaygesFortune.SaygesDamage, text: `Sayge's Damage` };
export const SaygesAgility = { actionId: () => ActionId.fromSpellId(23736), value: SaygesFortune.SaygesAgility, text: `Sayge's Agility` };
export const SaygesIntellect = { actionId: () => ActionId.fromSpellId(23766), value: SaygesFortune.SaygesIntellect, text: `Sayge's Intellect` };
export const SaygesSpirit = { actionId: () => ActionId.fromSpellId(23738), value: SaygesFortune.SaygesSpirit, text: `Sayge's Spirit` };
export const SaygesStamina = { actionId: () => ActionId.fromSpellId(23737), value: SaygesFortune.SaygesStamina, text: `Sayge's Stamina` };

// Dire Maul Buffs
export const FengusFerocity = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(22817),
		fieldName: 'fengusFerocity',
	}),
	`Fengus' Ferocity`,
);
export const MoldarsMoxie = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(22818),
		fieldName: 'moldarsMoxie',
	}),
	`Moldar's Moxie`,
);
export const SlipKiksSavvy = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(22820),
		fieldName: 'slipkiksSavvy',
	}),
	`Slip'kik's Savvy`,
);

///////////////////////////////////////////////////////////////////////////
//                                 DEBUFFS
///////////////////////////////////////////////////////////////////////////

const notForever = (player: Player<any>) => player.sim.getRuleset() !== Ruleset.RulesetForever;

// A debuff we turn on or off with one icon, even when its field is a tristate. An improved value
// from an older saved setup counts as on, and the sim treats it as the regular debuff.
function makeDebuffToggle(
	spellId: number,
	fieldName: DebuffToggleField,
	options: {
		showWhen?: (player: Player<any>) => boolean;
		// Runs on the debuffs when we turn this one on, before we save them.
		onTurnOn?: (debuffs: Debuffs, player: Player<any>) => void;
	} = {},
) {
	return InputHelpers.makeBooleanIconInput<any, Debuffs, Player<any>>(
		{
			getModObject: (player: Player<any>) => player,
			showWhen: (player: Player<any>) => !options.showWhen || options.showWhen(player),
			getValue: (player: Player<any>) => player.getRaid()!.getDebuffs(),
			setValue: (eventID: EventID, player: Player<any>, newVal: Debuffs) => player.getRaid()!.setDebuffs(eventID, newVal),
			changeEmitter: (player: Player<any>) =>
				TypedEvent.onAny([player.getRaid()!.debuffsChangeEmitter, player.raceChangeEmitter, player.sim.rulesetChangeEmitter]),
			getFieldValue: (player: Player<any>) => isDebuffOn(player.getRaid()!.getDebuffs(), fieldName),
			setFieldValue: (eventID: EventID, player: Player<any>, newValue: boolean) => {
				const debuffs = player.getRaid()!.getDebuffs();
				setDebuffOn(debuffs, fieldName, newValue);
				if (newValue) options.onTurnOn?.(debuffs, player);
				player.getRaid()!.setDebuffs(eventID, debuffs);
			},
		},
		() => ActionId.fromSpellId(spellId),
		fieldName,
	);
}

// A row of debuffs that don't stack, see ExclusiveDebuffRow. The improved versions are left
// out, because Forever has none of them and the sim's Expose Armor is the same either way.
function makeExclusiveDebuffRow(
	options: Array<{ spellId: number; fieldName: DebuffToggleField }>,
	exclusiveWhen: (player: Player<any>) => boolean = () => true,
): ExclusiveDebuffRowConfig {
	const fields = options.map(option => option.fieldName);
	return {
		fields,
		exclusiveWhen,
		options: options.map(option =>
			makeDebuffToggle(option.spellId, option.fieldName, {
				onTurnOn: (debuffs, player) => {
					if (!exclusiveWhen(player)) return;
					fields.filter(field => field !== option.fieldName).forEach(field => setDebuffOn(debuffs, field, false));
				},
			}),
		),
	};
}

export const MajorArmorDebuff = makeExclusiveDebuffRow([
	{ spellId: 11597, fieldName: 'sunderArmor' },
	{ spellId: 11198, fieldName: 'exposeArmor' },
]);

// Faerie Fire and Curse of Recklessness stack in Classic but not in Forever.
export const MinorArmorDebuff = makeExclusiveDebuffRow(
	[
		{ spellId: 11717, fieldName: 'curseOfRecklessness' },
		{ spellId: 9907, fieldName: 'faerieFire' },
	],
	player => !notForever(player),
);

export const CrystalYield = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(15235),
	fieldName: 'crystalYield',
});

export const CurseOfElements = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(11722),
	fieldName: 'curseOfElements',
});

// Forever has no Curse of Shadow. Its Curse of the Elements covers every Magic school.
export const CurseOfShadow = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(17937),
	fieldName: 'curseOfShadow',
	showWhen: notForever,
});

// Under Forever Improved Shadow Bolt, Improved Scorch, Winter's Chill, Shadow Weaving and
// Stormstrike only help the player who applies them, so the sim ignores another player's there.
export const SpellISBDebuff = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(17803),
	fieldName: 'improvedShadowBolt',
	showWhen: notForever,
});

export const SpellScorchDebuff = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(12873),
	fieldName: 'improvedScorch',
	showWhen: notForever,
});

export const SpellWintersChillDebuff = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(28595),
	fieldName: 'wintersChill',
	showWhen: notForever,
});

export const SpellStormstrikeDebuff = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(17364),
	fieldName: 'stormstrike',
	showWhen: notForever,
});

export const SpellShadowWeavingDebuff = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(15334),
	fieldName: 'shadowWeaving',
	showWhen: notForever,
});

// Under Forever Improved Seal of the Crusader and Improved Hunter's Mark are out of the trees,
// so these two are on or off there.
export const JudgementOfTheCrusader = makeTristateDebuffInput({
	actionId: () => ActionId.fromSpellId(20303),
	impId: ActionId.fromSpellId(20337),
	fieldName: 'judgementOfTheCrusader',
	showWhen: player => player.hasFactionBuffs(Faction.Alliance) && notForever(player),
});
export const JudgementOfTheCrusaderForever = makeDebuffToggle(20303, 'judgementOfTheCrusader', {
	showWhen: player => player.hasFactionBuffs(Faction.Alliance) && !notForever(player),
});

export const HuntersMark = makeTristateDebuffInput({
	actionId: () => ActionId.fromSpellId(14325),
	impId: ActionId.fromSpellId(19425),
	fieldName: 'huntersMark',
	showWhen: notForever,
});
export const HuntersMarkForever = makeDebuffToggle(14325, 'huntersMark', { showWhen: player => !notForever(player) });

export const JudgementOfWisdom = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(20355),
	fieldName: 'judgementOfWisdom',
	showWhen: player => player.hasFactionBuffs(Faction.Alliance),
});

export const GiftOfArthas = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(11374),
	fieldName: 'giftOfArthas',
});

// Defensive debuffs

export const AttackPowerDebuff = makeExclusiveDebuffRow([
	{ spellId: 11556, fieldName: 'demoralizingShout' },
	{ spellId: 9898, fieldName: 'demoralizingRoar' },
]);

export const MeleeAttackSpeedDebuff = makeExclusiveDebuffRow([
	{ spellId: 6343, fieldName: 'thunderClap' },
	{ spellId: 21992, fieldName: 'thunderfury' },
]);

export const MeleeHitDebuff = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(24977),
	fieldName: 'insectSwarm',
});

export const ScorpidSting = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(3043),
	fieldName: 'scorpidSting',
});

export const curseOfWeaknessDebuff = makeTristateDebuffInput({
	actionId: () => ActionId.fromSpellId(11708),
	impId: ActionId.fromSpellId(18181),
	fieldName: 'curseOfWeakness',
});

// Judgement of Light heals the players who hit the target, so it only matters for the damage we
// take.
export const JudgementOfLight = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(20346),
	fieldName: 'judgementOfLight',
	showWhen: player => player.hasFactionBuffs(Faction.Alliance),
});

///////////////////////////////////////////////////////////////////////////
//                                 CONFIGS
///////////////////////////////////////////////////////////////////////////

export const RAID_BUFFS_CONFIG = [
	// Core Stat Buffs
	{
		config: AllStatsBuff,
		picker: IconPicker,
		stats: [],
	},
	{
		config: BlessingOfKings,
		picker: IconPicker,
		stats: [],
	},
	{
		config: StaminaBuff,
		picker: IconPicker,
		stats: [],
	},
	{
		config: BloodPactBuff,
		picker: IconPicker,
		stats: [],
	},
	{
		config: IntellectBuff,
		picker: IconPicker,
		stats: [Stat.StatIntellect],
	},
	{
		config: SpiritBuff,
		picker: IconPicker,
		stats: [Stat.StatSpirit],
	},

	// Tank-related Buffs
	{
		config: ArmorBuff,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: PhysDamReductionBuff,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	// {
	// 	config: DamageReductionPercentBuff,
	// 	picker: IconPicker,
	// 	stats: [Stat.StatArmor],
	// },
	{
		config: ResistanceBuff,
		picker: MultiIconPicker,
		stats: [Stat.StatNatureResistance, Stat.StatShadowResistance, Stat.StatFireResistance, Stat.StatFrostResistance],
	},

	// Physical Damage Buffs
	{
		config: BlessingOfMight,
		picker: IconPicker,
		stats: [Stat.StatAttackPower, Stat.StatStrength, Stat.StatAgility],
	},
	{
		config: StrengthBuffHorde,
		picker: IconPicker,
		stats: [Stat.StatStrength],
	},
	{
		config: StrengthBuffHordeForever,
		picker: IconPicker,
		stats: [Stat.StatStrength],
	},
	{
		config: BattleShoutBuff,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: GraceOfAir,
		picker: IconPicker,
		stats: [Stat.StatAgility],
	},
	{
		config: GraceOfAirForever,
		picker: IconPicker,
		stats: [Stat.StatAgility],
	},
	{
		config: TotemWeaponBuffInput,
		picker: IconEnumPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: TrueshotAuraBuff,
		picker: IconPicker,
		stats: [Stat.StatRangedAttackPower, Stat.StatAttackPower],
	},
	{
		config: MeleeCritBuff,
		picker: IconPicker,
		stats: [Stat.StatMeleeCrit],
	},
	// Threat Buffs

	// Spell Damage Buffs
	{
		config: SpellCritBuff,
		picker: IconPicker,
		stats: [Stat.StatSpellCrit],
	},
	{
		config: BlessingOfWisdom,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: ManaSpringTotem,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
] as PickerStatOptions[];

export const MISC_BUFFS_CONFIG = [
	{
		config: AtieshMageBuff,
		picker: IconPicker,
		stats: [Stat.StatSpellCrit],
	},
	{
		config: AtieshWarlockBuff,
		picker: IconPicker,
		stats: [Stat.StatSpellPower, Stat.StatSpellDamage],
	},
	{
		config: AtieshPriestBuff,
		picker: IconPicker,
		stats: [Stat.StatHealingPower],
	},
	{
		config: AtieshDruidBuff,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: Thorns,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: RetributionAura,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: SanctityAura,
		picker: IconPicker,
		stats: [Stat.StatHolyPower],
	},
	{
		config: Innervate,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: PowerInfusion,
		picker: IconPicker,
		stats: [Stat.StatMP5, Stat.StatSpellPower, Stat.StatSpellDamage],
	},
	{
		config: BattleSquawkBuff,
		picker: IconPicker,
		stats: [Stat.StatMeleeHit],
	},
] as PickerStatOptions[];

export const WORLD_BUFFS_CONFIG = [
	{
		config: RallyingCryOfTheDragonslayer,
		picker: IconPicker,
		stats: [Stat.StatMeleeCrit, Stat.StatSpellCrit, Stat.StatAttackPower],
	},
	{
		config: SongflowerSerenade,
		picker: IconPicker,
		stats: [],
	},
	{
		config: SpiritOfZandalar,
		picker: IconPicker,
		stats: [],
	},
	{
		config: WarchiefsBlessing,
		picker: IconPicker,
		stats: [],
	},
	{
		config: FengusFerocity,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: MoldarsMoxie,
		picker: IconPicker,
		stats: [Stat.StatStamina],
	},
	{
		config: SlipKiksSavvy,
		picker: IconPicker,
		stats: [Stat.StatSpellCrit],
	},
] as PickerStatOptions[];

export const SAYGES_CONFIG = [
	{
		config: SaygesDamage,
		stats: [],
	},
	{
		config: SaygesAgility,
		stats: [Stat.StatAgility],
	},
	{
		config: SaygesIntellect,
		stats: [Stat.StatIntellect],
	},
	{
		config: SaygesSpirit,
		stats: [Stat.StatSpirit, Stat.StatMP5],
	},
	{
		config: SaygesStamina,
		stats: [Stat.StatStamina],
	},
] as ItemStatOption<SaygesFortune>[];

// A debuff in a subsection of the Settings tab: an icon, or a row of debuffs that don't stack.
// We leave out the ones that raise no stat the spec cares about, like for the other inputs.
export interface DebuffSubsectionItem {
	config: IconPickerConfig<Player<any>, any> | ExclusiveDebuffRowConfig;
	stats: Array<Stat>;
}

export interface DebuffSubsection {
	label: string;
	items: Array<DebuffSubsectionItem>;
}

// The debuffs that raise our damage.
export const OFFENSIVE_DEBUFF_SUBSECTIONS: Array<DebuffSubsection> = [
	{
		label: 'Armor Penetration',
		items: [
			{ config: MajorArmorDebuff, stats: [Stat.StatAttackPower] },
			{ config: MinorArmorDebuff, stats: [Stat.StatAttackPower] },
			{ config: CrystalYield, stats: [Stat.StatAttackPower, Stat.StatRangedAttackPower] },
		],
	},
	{
		label: 'Spell Damage',
		items: [
			{ config: CurseOfElements, stats: [Stat.StatSpellPower, Stat.StatSpellDamage] },
			{ config: CurseOfShadow, stats: [Stat.StatSpellPower, Stat.StatSpellDamage] },
			{ config: SpellISBDebuff, stats: [Stat.StatShadowPower] },
			{ config: SpellScorchDebuff, stats: [Stat.StatFirePower] },
			{ config: SpellWintersChillDebuff, stats: [Stat.StatFrostPower] },
			{ config: SpellStormstrikeDebuff, stats: [Stat.StatNaturePower] },
			{ config: SpellShadowWeavingDebuff, stats: [Stat.StatShadowPower] },
			{ config: JudgementOfTheCrusader, stats: [Stat.StatHolyPower] },
			{ config: JudgementOfTheCrusaderForever, stats: [Stat.StatHolyPower] },
		],
	},
	{
		label: 'Other',
		items: [
			{ config: GiftOfArthas, stats: [Stat.StatAttackPower, Stat.StatRangedAttackPower] },
			{ config: HuntersMark, stats: [Stat.StatRangedAttackPower] },
			{ config: HuntersMarkForever, stats: [Stat.StatRangedAttackPower] },
			{ config: JudgementOfWisdom, stats: [Stat.StatMP5, Stat.StatIntellect] },
		],
	},
];

// The debuffs that lower the damage the target does. They're off for now, see
// normalizeDebuffs.
export const DEFENSIVE_DEBUFF_SUBSECTIONS: Array<DebuffSubsection> = [
	{ label: 'Attack Power', items: [{ config: AttackPowerDebuff, stats: [Stat.StatArmor] }] },
	{ label: 'Attack Speed', items: [{ config: MeleeAttackSpeedDebuff, stats: [Stat.StatArmor] }] },
	{
		label: 'Chance to Hit',
		items: [
			{ config: MeleeHitDebuff, stats: [Stat.StatDodge] },
			{ config: ScorpidSting, stats: [Stat.StatDodge] },
		],
	},
	{
		label: 'Other',
		items: [
			{ config: curseOfWeaknessDebuff, stats: [Stat.StatArmor] },
			{ config: JudgementOfLight, stats: [Stat.StatStamina] },
		],
	},
];
