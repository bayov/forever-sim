import { setItemQualityCssClass } from '../../css_utils';
import { UIEnchant } from '../../proto/ui';
import { ActionId } from '../../proto_utils/action_id';
import { EquippedItem } from '../../proto_utils/equipped_item';

// Links to an item or an enchant in its quality color, for lists like the changes on a modified
// preset. Hovering one shows wowhead's tooltip for it, or ours for items wowhead doesn't know
// (see item_tooltip_fallback.ts), because both look for item and spell links on the page.
//
// Like in the game, the name goes in brackets: [Crown of Destruction].

export function itemLink(item: EquippedItem): HTMLElement {
	const link = document.createElement('a');
	link.textContent = `[${item.item.name + (item.randomSuffix ? ' ' + item.randomSuffix.name : '')}]`;
	link.href = ActionId.makeItemUrl(item.item.id, item.randomSuffix?.id);
	link.target = '_blank';
	setItemQualityCssClass(link, item.item.quality);
	return link;
}

export function enchantLink(enchant: UIEnchant): HTMLElement {
	const link = document.createElement('a');
	link.textContent = `[${enchant.name}]`;
	link.href = enchant.spellId ? ActionId.makeSpellUrl(enchant.spellId) : ActionId.makeItemUrl(enchant.itemId);
	link.target = '_blank';
	link.dataset.whtticon = 'false';
	setItemQualityCssClass(link, enchant.quality);
	return link;
}

// A link to a buff, a debuff or a consumable in the game's light blue for spells, like
// [Blessing of Might], with wowhead's tooltip on hover.
//
// We leave out the rank, like in 'Sunder Armor (Rank 5)', because we always get the highest
// rank our level has.
export function actionLink(actionId: ActionId): HTMLElement {
	const link = document.createElement('a');
	link.textContent = `[${withoutRank(actionId.name)}]`;
	actionId.setWowheadHref(link);
	link.target = '_blank';
	link.dataset.whtticon = 'false';
	link.classList.add('action-link');
	return link;
}

export function withoutRank(name: string): string {
	return name.replace(/ \(Rank \d+\)$/, '');
}
