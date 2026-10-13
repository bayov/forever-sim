import * as InputHelpers from '../core/components/input_helpers.js';
import { Spec } from '../core/proto/common.js';
import { Mage_Options_ArmorType as ArmorType, Mage_Options_StaffImbue as StaffImbue } from '../core/proto/mage.js';
import { ActionId } from '../core/proto_utils/action_id.js';
import { TypedEvent } from '../core/typed_event.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const Armor = InputHelpers.makeSpecOptionsEnumIconInput<Spec.SpecMage, ArmorType>({
	fieldName: 'armor',
	values: [
		{ value: ArmorType.NoArmor, tooltip: 'No Armor' },
		{
			actionId: () => ActionId.fromSpellId(22783),
			value: ArmorType.MageArmor,
		},
		{
			actionId: () => ActionId.fromSpellId(10220),
			value: ArmorType.IceArmor,
		},
	],
	changeEmitter: player => TypedEvent.onAny([player.gearChangeEmitter, player.specOptionsChangeEmitter]),
});

// Forever's mage-only staff imbue scrolls, in the class imbue slot (they stack with a weapon oil).
export const StaffImbueInput = InputHelpers.makeSpecOptionsEnumIconInput<Spec.SpecMage, StaffImbue>({
	fieldName: 'staffImbue',
	values: [
		{ value: StaffImbue.NoStaffImbue, tooltip: 'No Staff Imbue' },
		{ actionId: () => ActionId.fromItemId(277500), value: StaffImbue.ImbueGreaterFlame, tooltip: 'Scroll of Imbue Greater Flame: +20 Fire damage' },
		{ actionId: () => ActionId.fromItemId(277501), value: StaffImbue.ImbueGreaterFrost, tooltip: 'Scroll of Imbue Greater Frost: +20 Frost damage' },
		{ actionId: () => ActionId.fromItemId(277502), value: StaffImbue.ImbuePrecision, tooltip: 'Scroll of Imbue Precision: +1% spell crit' },
	],
	changeEmitter: player => TypedEvent.onAny([player.gearChangeEmitter, player.specOptionsChangeEmitter]),
});

export const MageRotationConfig = {
	inputs: [],
};
