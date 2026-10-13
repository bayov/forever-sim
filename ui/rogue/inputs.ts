import * as InputHelpers from '../core/components/input_helpers.js';
import { Player } from '../core/player.js';
import { ItemSlot, Spec, WeaponImbue } from '../core/proto/common.js';
import { ActionId } from '../core/proto_utils/action_id.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

// Forever's poisons are class imbues (client 70338: SpellEffect 360) that sit beside a sharpening stone,
// so each hand's poison is a rogue option and the consumable imbue slot holds the stone.
const poisonValues = [
	{ value: WeaponImbue.WeaponImbueUnknown, tooltip: 'No Poison' },
	{ actionId: () => ActionId.fromItemId(20844), value: WeaponImbue.DeadlyPoison },
	{ actionId: () => ActionId.fromItemId(8928), value: WeaponImbue.InstantPoison },
	{ actionId: () => ActionId.fromItemId(10922), value: WeaponImbue.WoundPoison },
];

export const MainHandPoison = InputHelpers.makeSpecOptionsEnumIconInput<Spec.SpecRogue, WeaponImbue>({
	fieldName: 'mainHandPoison',
	values: poisonValues,
	tooltip: 'Main Hand Poison',
});

export const OffHandPoison = InputHelpers.makeSpecOptionsEnumIconInput<Spec.SpecRogue, WeaponImbue>({
	fieldName: 'offHandPoison',
	values: poisonValues,
	tooltip: 'Off Hand Poison',
	showWhen: (player: Player<Spec.SpecRogue>) => player.getEquippedItem(ItemSlot.ItemSlotOffHand) != null,
	changeEmitter: (player: Player<Spec.SpecRogue>) => player.changeEmitter,
});
