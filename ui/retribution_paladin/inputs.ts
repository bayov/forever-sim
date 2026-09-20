import * as InputHelpers from '../core/components/input_helpers.js';
import { Player } from '../core/player.js';
import { Spec } from '../core/proto/common.js';
import { PaladinAura,PaladinSeal } from '../core/proto/paladin.js';
import { ActionId } from '../core/proto_utils/action_id.js';
import { TypedEvent } from '../core/typed_event.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

// Sanctity Aura is not in Forever's Retribution tree, so the paladin's own damage aura is
// Retribution Aura, which only does anything while something hits the paladin.
export const AuraSelection = InputHelpers.makeSpecOptionsEnumIconInput<Spec.SpecRetributionPaladin, PaladinAura>({
	fieldName: 'aura',
	values: [
		{ value: PaladinAura.NoPaladinAura, tooltip: 'No Aura' },
		{ actionId: () => ActionId.fromSpellId(10301), value: PaladinAura.RetributionAura },
	],
});

// The below is used in the custom APL action "Cast Primary Seal".
// Only shows SoC if it's talented, only shows SoM if the relevant rune is equipped.
export const PrimarySealSelection = InputHelpers.makeSpecOptionsEnumIconInput<Spec.SpecRetributionPaladin, PaladinSeal>({
	fieldName: 'primarySeal',
	values: [
		{
			actionId: () => ActionId.fromSpellId(20293),
			value: PaladinSeal.Righteousness,
		},
		{
			actionId: () => ActionId.fromSpellId(20920),
			value: PaladinSeal.Command,
			showWhen: (player: Player<Spec.SpecRetributionPaladin>) => player.getTalents().sealOfCommand,
		},
	],
	// changeEmitter: (player: Player<Spec.SpecRetributionPaladin>) => player.changeEmitter,
	changeEmitter: (player: Player<Spec.SpecRetributionPaladin>) =>
		TypedEvent.onAny([player.gearChangeEmitter, player.talentsChangeEmitter, player.specOptionsChangeEmitter]),
});

export const JudgementStopAttack = InputHelpers.makeSpecOptionsBooleanInput<Spec.SpecRetributionPaladin>({
	fieldName: 'isUsingJudgementStopAttack',
	label: 'Using Judgement StopAttack Macro',
	labelTooltip: 'Allows saving of extra attacks',
});
