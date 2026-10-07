import tippy, { Instance } from 'tippy.js';

import { WOWHEAD_DOMAIN, WOWHEAD_EXPANSION_ENV } from './constants/wowhead';
import { CRIT_RATING_PER_PERCENT, HIT_RATING_PER_PERCENT } from './player';
import { EnchantType, ItemQuality, ItemSlot, ItemType, Stat } from './proto/common';
import { UIEnchant, UIItem } from './proto/ui';
import { Database } from './proto_utils/database';
import { armorTypeNames, classNames, itemTypeNames, rangedWeaponTypeNames, statNames, weaponTypeNames } from './proto_utils/names';

// Tooltips for items wowhead has never heard of, from our own item database.
//
// The tooltips everywhere else come from wowhead's script, which reads the item id off the link.
// For an id Forever's wowhead database does not carry (the made up "Enhancement Synthetic" items
// from 990001 up, or a client item wowhead has not picked up yet) it shows an error and no stats.
// So on the first hover over such a link we ask wowhead's tooltip endpoint about the id. When it
// answers with an error, we take wowhead's tooltip off the link and show ours, with a warning on
// top that wowhead does not know the item.
//
// The Heavy and Thick armor kits get our tooltip too. Forever gave the kits Stamina, and its
// wowhead text for those two lost the numbers: "Permanently increase the Stamina value by 0 and
// armor value by 0". The gear picker links a kit by its spell and the enchant list links it by
// its item, so we check both kinds of link.

const QUALITY_CLASSES: Record<number, string> = {
	[ItemQuality.ItemQualityJunk]: 'item-quality-junk',
	[ItemQuality.ItemQualityCommon]: 'item-quality-common',
	[ItemQuality.ItemQualityUncommon]: 'item-quality-uncommon',
	[ItemQuality.ItemQualityRare]: 'item-quality-rare',
	[ItemQuality.ItemQualityEpic]: 'item-quality-epic',
	[ItemQuality.ItemQualityLegendary]: 'item-quality-legendary',
};

const PRIMARY_STATS = [Stat.StatStrength, Stat.StatAgility, Stat.StatStamina, Stat.StatIntellect, Stat.StatSpirit];
const RESISTANCES = [Stat.StatArcaneResistance, Stat.StatFireResistance, Stat.StatFrostResistance, Stat.StatNatureResistance, Stat.StatShadowResistance];
// Classic items carry these as a percent. Forever's new items carry hit and crit as rating
// (UIItem.hitRating and critRating), which we turn into a percent below.
const PERCENT_STATS = [Stat.StatMeleeHit, Stat.StatMeleeCrit, Stat.StatSpellHit, Stat.StatSpellCrit, Stat.StatDodge, Stat.StatParry, Stat.StatBlock];

// One lookup per item id, shared by every link to it.
const wowheadKnows = new Map<number, Promise<boolean>>();

function isOnWowhead(itemId: number): Promise<boolean> {
	let known = wowheadKnows.get(itemId);
	if (!known) {
		const url = `https://nether.wowhead.com/${WOWHEAD_DOMAIN}/tooltip/item/${itemId}?dataEnv=${WOWHEAD_EXPANSION_ENV}`;
		known = fetch(url)
			.then(response => response.json())
			.then(json => !json['error'])
			// When wowhead can't be reached we leave its tooltip alone.
			.catch(() => true);
		wowheadKnows.set(itemId, known);
	}
	return known;
}

// One lookup per kit spell id. True when wowhead's text for the kit has a 0 in place of its stats.
const wowheadKitBroken = new Map<number, Promise<boolean>>();

function isKitBrokenOnWowhead(spellId: number): Promise<boolean> {
	let broken = wowheadKitBroken.get(spellId);
	if (!broken) {
		const url = `https://nether.wowhead.com/${WOWHEAD_DOMAIN}/tooltip/spell/${spellId}?dataEnv=${WOWHEAD_EXPANSION_ENV}`;
		broken = fetch(url)
			.then(response => response.json())
			.then(json => !!json['error'] || / by 0\b/.test(json['tooltip'] ?? ''))
			.catch(() => false);
		wowheadKitBroken.set(spellId, broken);
	}
	return broken;
}

type LinkKind = 'item' | 'spell';

interface LinkedId {
	kind: LinkKind;
	id: number;
}

const LINK_REGEX: Record<LinkKind, RegExp> = {
	item: /[?&/]item=(\d+)/,
	spell: /[?&/]spell=(\d+)/,
};

function linkedId(elem: HTMLElement): LinkedId | null {
	for (const kind of ['item', 'spell'] as const) {
		const regex = LINK_REGEX[kind];
		const match = (elem.getAttribute('href') ?? '').match(regex) ?? ('?' + (elem.dataset.wowhead ?? '')).match(regex);
		if (match) return { kind, id: parseInt(match[1]) };
	}
	return null;
}

const sameLink = (a: LinkedId | null, b: LinkedId) => a?.kind === b.kind && a.id === b.id;

// Every armor kit fits the chest, so the chest list has them all.
function findKit(db: Database, link: LinkedId): UIEnchant | undefined {
	return db
		.getEnchants(ItemSlot.ItemSlotChest)
		.find(enchant => enchant.enchantType === EnchantType.EnchantTypeKit && (link.kind === 'item' ? enchant.itemId : enchant.spellId) === link.id);
}

function line(text: string, cssClass = ''): string {
	const div = document.createElement('div');
	div.textContent = text;
	if (cssClass) div.className = cssClass;
	return div.outerHTML;
}

function tooltipContent(itemId: number, item: UIItem | undefined): string {
	const lines = [line('Not found on wowhead. Stats from the sim’s item database.', 'text-warning small mb-1')];
	if (!item) {
		lines.push(line(`Item ${itemId}`));
		return lines.join('');
	}

	lines.push(line(item.name, `fw-bold ${QUALITY_CLASSES[item.quality] ?? ''}`));
	if (item.ilvl) lines.push(line(`Item Level ${item.ilvl}`, 'text-warning'));
	if (item.unique) lines.push(line('Unique'));

	let kind = '';
	if (item.type === ItemType.ItemTypeWeapon) kind = weaponTypeNames.get(item.weaponType) ?? '';
	else if (item.type === ItemType.ItemTypeRanged) kind = rangedWeaponTypeNames.get(item.rangedWeaponType) ?? '';
	else kind = armorTypeNames.get(item.armorType) ?? '';
	lines.push(line([itemTypeNames.get(item.type), kind].filter(Boolean).join(', ')));

	if (item.weaponSpeed) {
		const dps = (item.weaponDamageMin + item.weaponDamageMax) / 2 / item.weaponSpeed;
		lines.push(line(`${item.weaponDamageMin} - ${item.weaponDamageMax} Damage, Speed ${item.weaponSpeed.toFixed(2)}`));
		lines.push(line(`(${dps.toFixed(1)} damage per second)`));
	}

	const stats = item.stats;
	if (stats[Stat.StatArmor]) lines.push(line(`${stats[Stat.StatArmor]} Armor`));
	for (const stat of [...PRIMARY_STATS, ...RESISTANCES]) {
		if (stats[stat]) lines.push(line(`${stats[stat] > 0 ? '+' : ''}${stats[stat]} ${statNames.get(stat)}`));
	}
	// Bonus armor goes on its own line after the stats, the way the game lists it ("43 Armor",
	// "+7 Stamina", "+20 Armor" on Black Wolf Bracers).
	if (stats[Stat.StatBonusArmor]) lines.push(line(`+${stats[Stat.StatBonusArmor]} Armor`));
	if (item.requiredLevel) lines.push(line(`Requires Level ${item.requiredLevel}`));
	if (item.classAllowlist.length) lines.push(line(`Classes: ${item.classAllowlist.map(c => classNames.get(c)).join(', ')}`));

	const equips: Array<string> = [];
	stats.forEach((value, stat) => {
		if (!value || stat === Stat.StatArmor || stat === Stat.StatBonusArmor || PRIMARY_STATS.includes(stat) || RESISTANCES.includes(stat)) return;
		// Classic items list their attack power twice, once for melee and once for ranged.
		if (stat === Stat.StatRangedAttackPower && value === stats[Stat.StatAttackPower]) return;
		const percent = PERCENT_STATS.includes(stat) ? '%' : '';
		equips.push(`Equip: +${value}${percent} ${statNames.get(stat) ?? `stat ${stat}`}`);
	});
	// The client stores hit and crit as rating, but the game shows them as a flat percent at every
	// level (9 hit rating on Frozen Heart of the Mountain reads "by 0.9%"), so we print them the same way.
	if (item.hitRating) equips.push(`Equip: Improves your chance to hit by ${(item.hitRating / HIT_RATING_PER_PERCENT).toFixed(1)}%.`);
	if (item.critRating) equips.push(`Equip: Improves your chance to get a critical strike by ${(item.critRating / CRIT_RATING_PER_PERCENT).toFixed(1)}%.`);
	for (const equip of equips) lines.push(line(equip, 'item-quality-uncommon'));

	if (item.setName) lines.push(line(item.setName, 'text-warning mt-1'));
	return lines.join('');
}

function kitTooltipContent(kit: UIEnchant): string {
	const lines = [line('wowhead shows this kit without its stats. Stats from the sim’s enchant database.', 'text-warning small mb-1')];
	lines.push(line(kit.name, `fw-bold ${QUALITY_CLASSES[kit.quality] ?? ''}`));
	if (kit.requiredLevel) lines.push(line(`Requires Level ${kit.requiredLevel}`));
	const stats: Array<string> = [];
	kit.stats.forEach((value, stat) => {
		if (!value) return;
		// Attack power kits list it twice, once for melee and once for ranged.
		if (stat === Stat.StatRangedAttackPower && value === kit.stats[Stat.StatAttackPower]) return;
		const name = stat === Stat.StatBonusArmor ? 'Armor' : statNames.get(stat) ?? `stat ${stat}`;
		stats.push(`+${value} ${name}`);
	});
	lines.push(line(`Use: Permanently adds ${stats.join(', ')} to an item worn on the chest, legs, hands or feet.`, 'item-quality-uncommon'));
	return lines.join('');
}

// A link we took over gets this href. wowhead's script ignores it, and it tells us on every show
// whether the link still belongs to the item, because the page rewrites or clears the href when the
// item changes.
const fallbackHref = (link: LinkedId) => `#fallback-${link.kind}-${link.id}`;

interface FallbackLink extends HTMLElement {
	_fallbackTooltip?: Instance;
}

// Returns our tooltip for the link, or nothing when wowhead's is fine.
async function fallbackContent(target: LinkedId): Promise<string | null> {
	const db = await Database.get();
	const kit = findKit(db, target);
	if (kit) return (await isKitBrokenOnWowhead(kit.spellId)) ? kitTooltipContent(kit) : null;
	if (target.kind !== 'item' || (await isOnWowhead(target.id))) return null;
	return tooltipContent(target.id, db.getItemById(target.id));
}

async function onHover(link: FallbackLink) {
	const target = linkedId(link);
	if (!target) return;
	const content = await fallbackContent(target);
	// The link may point at another item by the time wowhead answers.
	if (!content || !sameLink(linkedId(link), target)) return;

	// Without the item id on the link wowhead's script has nothing to show, so only ours appears.
	link.setAttribute('href', fallbackHref(target));
	link.removeAttribute('data-wowhead');
	const whWindow = window as any;
	whWindow.WH?.Tooltip?.hide?.();
	whWindow.$WowheadPower?.hideTooltip?.();

	link._fallbackTooltip?.destroy();
	const tip = tippy(link, {
		content,
		allowHTML: true,
		placement: 'right',
		maxWidth: 320,
		onShow: instance => {
			if (link.getAttribute('href') !== fallbackHref(target)) {
				instance.destroy();
				if (link._fallbackTooltip === instance) delete link._fallbackTooltip;
				return false;
			}
		},
	});
	link._fallbackTooltip = tip;
	if (link.matches(':hover')) tip.show();
}

let installed = false;

// installItemTooltipFallback watches every item and spell link on the page, once per page load.
export function installItemTooltipFallback() {
	if (installed) return;
	installed = true;
	// The capture phase runs before wowhead's own handler, so the href is gone before it looks.
	document.addEventListener(
		'mouseover',
		event => {
			const target = event.target as HTMLElement | null;
			const link = target?.closest?.('[href*="item="], [data-wowhead*="item="], [href*="spell="], [data-wowhead*="spell="]') as FallbackLink | null;
			if (link) onHover(link);
		},
		true,
	);
	// The made up href must not land in the page address, the sim keeps its settings there.
	document.addEventListener(
		'click',
		event => {
			const link = (event.target as HTMLElement | null)?.closest?.('[href^="#fallback-"]');
			if (link) event.preventDefault();
		},
		true,
	);
}
