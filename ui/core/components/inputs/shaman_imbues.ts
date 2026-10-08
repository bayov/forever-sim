import { Class, ItemSlot, Spec, WeaponImbue } from '../../proto/common.js';
import { Player } from '../../player';
import { ActionId } from '../../proto_utils/action_id';
import { isWeapon } from '../../proto_utils/utils';
import { ConsumableInputConfig } from './consumables';

// Shaman Imbues
//
// Under Forever a shaman imbue stacks with an oil or a stone, so Enhancement picks its imbue in
// the class settings (see ShamanImbueInput) and the consumables keep the weapon slot for oils
// and stones. The other shaman specs have no such setting yet, so they still pick it here.
const shownForShaman = (player: Player<any>) => player.isClass(Class.ClassShaman) && !player.isSpec(Spec.SpecEnhancementShaman);

export const RockbiterWeaponImbue = (slot: ItemSlot): ConsumableInputConfig<WeaponImbue> => {
	return {
		actionId: () => ActionId.fromSpellId(16316),
		value: WeaponImbue.RockbiterWeapon,
		showWhen: player => {
			if (!shownForShaman(player)) return false;

			const weapon = player.getEquippedItem(slot);
			return !weapon || isWeapon(weapon.item.weaponType);
		},
	};
};

export const FlametongueWeaponImbue = (slot: ItemSlot): ConsumableInputConfig<WeaponImbue> => {
	return {
		actionId: () => ActionId.fromSpellId(16342),
		value: WeaponImbue.FlametongueWeapon,
		showWhen: player => {
			if (!shownForShaman(player)) return false;
			const weapon = player.getEquippedItem(slot);
			return !weapon || isWeapon(weapon.item.weaponType);
		},
	};
};

export const FrostbrandWeaponImbue = (slot: ItemSlot): ConsumableInputConfig<WeaponImbue> => {
	return {
		actionId: () => ActionId.fromSpellId(16356),
		value: WeaponImbue.FrostbrandWeapon,
		showWhen: player => {
			if (!shownForShaman(player)) return false;
			const weapon = player.getEquippedItem(slot);
			return !weapon || isWeapon(weapon.item.weaponType);
		},
	};
};

export const WindfuryWeaponImbue = (slot: ItemSlot): ConsumableInputConfig<WeaponImbue> => {
	return {
		actionId: () => ActionId.fromSpellId(16362),
		value: WeaponImbue.WindfuryWeapon,
		showWhen: player => {
			if (!shownForShaman(player)) return false;
			const weapon = player.getEquippedItem(slot);
			return !weapon || isWeapon(weapon.item.weaponType);
		},
	};
};
