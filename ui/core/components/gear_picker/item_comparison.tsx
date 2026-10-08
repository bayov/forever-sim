import { WOWHEAD_DOMAIN, WOWHEAD_EXPANSION_ENV } from '../../constants/wowhead';
import { isOnWowhead, tooltipContent } from '../../item_tooltip_fallback';
import { CRIT_RATING_PER_PERCENT, HIT_RATING_PER_PERCENT } from '../../player';
import { ItemRandomSuffix, Stat } from '../../proto/common';
import { UIItem } from '../../proto/ui';
import { EquippedItem } from '../../proto_utils/equipped_item';
import { statNames } from '../../proto_utils/names';

// Shows the equipped item next to the tooltip of an item we hover in the item list, like the
// game does when we hover gear in our bags.
//
// The equipped item's tooltip comes from wowhead's tooltip endpoint, with its enchant and
// random suffix, in the same frame wowhead's script draws around the hovered one. Items wowhead
// doesn't know get the tooltip from our own item database (see item_tooltip_fallback.ts).
// Under it we list the stats we'd gain or lose by swapping, like "+4 Armor" and "-13 Stamina".
//
// wowhead's tooltip for the hovered item stays where wowhead opens it. We float ours next to it,
// on the side away from the mouse, so neither tooltip covers the item list where we point. We
// only take the side with the mouse when the other side has no room. We follow wowhead's tooltip
// while we hover, because wowhead moves it with the mouse.

// The space between the two tooltips.
const GAP = 8;

// The item list, the gear pane and the consumable slots already show each item's icon right where
// we point, so while we hover them the tooltips leave theirs out (wowhead's and ours). Other links,
// like the items in a modified preset's tooltip, keep the icon.
export function hideTooltipIconsWhileHovered(elem: HTMLElement) {
	elem.addEventListener('mouseenter', () => document.body.classList.add('hide-wowhead-tooltip-icons'));
	elem.addEventListener('mouseleave', () => document.body.classList.remove('hide-wowhead-tooltip-icons'));
}

// One request per item, enchant and suffix, shared by every hover.
const wowheadTooltips = new Map<string, Promise<{ tooltip: string; icon?: string } | null>>();

function fetchWowheadTooltip(item: EquippedItem): Promise<{ tooltip: string; icon?: string } | null> {
	const params = new URLSearchParams({ dataEnv: String(WOWHEAD_EXPANSION_ENV) });
	if (item.randomSuffix) params.set('rand', String(item.randomSuffix.id));
	if (item.enchant) params.set('ench', String(item.enchant.effectId));
	const url = `https://nether.wowhead.com/${WOWHEAD_DOMAIN}/tooltip/item/${item.item.id}?${params}`;
	let tooltip = wowheadTooltips.get(url);
	if (!tooltip) {
		tooltip = fetch(url)
			.then(response => response.json())
			.then(json => (json['error'] || !json['tooltip'] ? null : { tooltip: json['tooltip'], icon: json['icon'] }))
			.catch(() => null);
		wowheadTooltips.set(url, tooltip);
	}
	return tooltip;
}

// The equipped item's tooltip in wowhead's frame, with the stat changes under it.
async function buildTooltip(equipped: EquippedItem, hovered: UIItem): Promise<HTMLElement> {
	const fromWowhead = (await isOnWowhead(equipped.item.id)) ? await fetchWowheadTooltip(equipped) : null;
	const body = document.createElement('div');
	body.innerHTML = fromWowhead?.tooltip ?? tooltipContent(equipped.item.id, equipped.item);

	return (
		<div
			className="wowhead-tooltip wowhead-tooltip-width-restriction wowhead-tooltip-width-320 item-comparison-tooltip"
			dataset={{ visible: 'yes', game: 'wow', tree: 'classicplus', env: 'classicplus', type: 'item' }}>
			<table>
				<tbody>
					<tr>
						<td>
							<div className="item-comparison-label">Currently Equipped</div>
							{body}
							{statChanges(equipped, hovered)}
						</td>
						<th style={{ backgroundPosition: 'right top' }} />
					</tr>
					<tr>
						<th style={{ backgroundPosition: 'left bottom' }} />
						<th style={{ backgroundPosition: 'right bottom' }} />
					</tr>
				</tbody>
			</table>
		</div>
	) as HTMLElement;
}

// The stats an item gives, by the name we show them under. Forever pays out the hit and crit
// from gear against every kind of attack, so we add up the melee, spell and rating parts into
// one Hit and one Critical Strike, the way the sim does.
function itemStats(item: UIItem, randomSuffix: ItemRandomSuffix | null): Map<string, { value: number; percent: boolean }> {
	const stats = item.stats.slice();
	randomSuffix?.stats.forEach((value, stat) => (stats[stat] = (stats[stat] ?? 0) + value));
	const result = new Map<string, { value: number; percent: boolean }>();
	const add = (name: string, value: number, percent = false) => {
		if (value) result.set(name, { value: (result.get(name)?.value ?? 0) + value, percent });
	};

	if (item.weaponSpeed) add('Damage Per Second', (item.weaponDamageMin + item.weaponDamageMax) / 2 / item.weaponSpeed);
	// Bonus armor counts as armor too.
	add('Armor', (stats[Stat.StatArmor] ?? 0) + (stats[Stat.StatBonusArmor] ?? 0));
	add('Hit', (stats[Stat.StatMeleeHit] ?? 0) + (stats[Stat.StatSpellHit] ?? 0) + item.hitRating / HIT_RATING_PER_PERCENT, true);
	add('Critical Strike', (stats[Stat.StatMeleeCrit] ?? 0) + (stats[Stat.StatSpellCrit] ?? 0) + item.critRating / CRIT_RATING_PER_PERCENT, true);
	const handled = [Stat.StatArmor, Stat.StatBonusArmor, Stat.StatMeleeHit, Stat.StatSpellHit, Stat.StatMeleeCrit, Stat.StatSpellCrit];
	stats.forEach((value, stat) => {
		if (handled.includes(stat)) return;
		// Classic items list their attack power twice, once for melee and once for ranged.
		if (stat === Stat.StatRangedAttackPower && value === stats[Stat.StatAttackPower]) return;
		add(statNames.get(stat) ?? `Stat ${stat}`, value, [Stat.StatDodge, Stat.StatParry, Stat.StatBlock].includes(stat));
	});
	return result;
}

// The gains come first and the losses after, like in the game.
function statChanges(equipped: EquippedItem, hovered: UIItem): HTMLElement | undefined {
	const before = itemStats(equipped.item, equipped.randomSuffix);
	const after = itemStats(hovered, null);
	const changes: Array<{ name: string; delta: number; percent: boolean }> = [];
	new Set([...after.keys(), ...before.keys()]).forEach(name => {
		const delta = (after.get(name)?.value ?? 0) - (before.get(name)?.value ?? 0);
		if (Math.abs(delta) >= 0.005) changes.push({ name, delta, percent: (after.get(name) ?? before.get(name))!.percent });
	});
	if (!changes.length) return undefined;
	changes.sort((a, b) => Number(b.delta > 0) - Number(a.delta > 0));

	const format = (value: number) => String(Math.round(Math.abs(value) * 100) / 100);
	return (
		<div className="item-comparison-changes">
			<div className="item-comparison-changes-title">If you replace this item, the following stat changes will occur:</div>
			{changes.map(({ name, delta, percent }) => (
				<div>
					<span className={delta > 0 ? 'item-comparison-gain' : 'item-comparison-loss'}>{`${delta > 0 ? '+' : '-'}${format(delta)}${
						percent ? '%' : ''
					}`}</span>
					{` ${name}`}
				</div>
			))}
		</div>
	) as HTMLElement;
}

// The tooltip of the hovered item: wowhead's, or ours for items wowhead doesn't know.
function hoveredTooltip(): HTMLElement | null {
	const isShown = (elem: Element) => getComputedStyle(elem).visibility === 'visible' && elem.getBoundingClientRect().width > 0;
	const wowhead = Array.from(document.querySelectorAll<HTMLElement>('body > .wowhead-tooltip:not(.item-comparison-tooltip)')).find(isShown);
	if (wowhead) return wowhead;
	return (
		Array.from(document.querySelectorAll<HTMLElement>('[data-tippy-root] > .tippy-box[data-state="visible"]')).find(box => isShown(box.parentElement!)) ??
		null
	);
}

// The left edge of a tooltip, counting the icon wowhead puts out to its left.
function leftEdge(tooltip: HTMLElement): number {
	// A hidden icon (see hideTooltipIconsWhileHovered) has no size, and doesn't count.
	const icon = tooltip.querySelector('.whtt-tooltip-icon')?.getBoundingClientRect();
	return Math.min(tooltip.getBoundingClientRect().left, icon?.width ? icon.left : Infinity);
}

function place(panel: HTMLElement, mouseX: number) {
	const reference = hoveredTooltip();
	if (!reference) {
		panel.style.visibility = 'hidden';
		return;
	}
	const rect = reference.getBoundingClientRect();
	const width = panel.getBoundingClientRect().width;
	// How far the panel's own icon sticks out to its left.
	const iconOutset = panel.getBoundingClientRect().left - leftEdge(panel);
	const right = rect.right + GAP + iconOutset;
	const left = leftEdge(reference) - GAP - width;
	const fitsRight = right + width <= window.innerWidth;
	const fitsLeft = left - iconOutset >= 0;
	// wowhead opens its tooltip beside the mouse, so the side away from the mouse is the one
	// past the far edge of the tooltip.
	const mouseOnLeft = mouseX < (rect.left + rect.right) / 2;
	const awayFromMouse = mouseOnLeft ? (fitsRight ? right : undefined) : fitsLeft ? left : undefined;
	const towardMouse = mouseOnLeft ? (fitsLeft ? left : undefined) : fitsRight ? right : undefined;
	const x = awayFromMouse ?? towardMouse ?? Math.max(iconOutset, window.innerWidth - width);
	const top = Math.max(0, Math.min(rect.top, window.innerHeight - panel.getBoundingClientRect().height));
	panel.style.left = `${x}px`;
	panel.style.top = `${top}px`;
	panel.style.visibility = 'visible';
}

// Shows the comparison while we hover the item's row. Nothing shows when the slot is empty or
// already holds the item.
export function attachItemComparison(elem: HTMLElement, hovered: () => UIItem, equipped: () => EquippedItem | null) {
	let panel: HTMLElement | null = null;
	let frame = 0;
	let hovering = false;
	let mouseX = 0;
	const trackMouse = (event: MouseEvent) => (mouseX = event.clientX);

	const stop = () => {
		hovering = false;
		elem.removeEventListener('mousemove', trackMouse);
		cancelAnimationFrame(frame);
		panel?.remove();
		panel = null;
	};

	elem.addEventListener('mouseenter', async event => {
		mouseX = event.clientX;
		elem.addEventListener('mousemove', trackMouse);
		const equippedItem = equipped();
		const item = hovered();
		if (!equippedItem || equippedItem.item.id === item.id) return;
		hovering = true;
		const built = await buildTooltip(equippedItem, item);
		// We may have left the row while the tooltip loaded.
		if (!hovering || panel) return;
		panel = built;
		panel.style.position = 'fixed';
		// wowhead's script sets this width on its item tooltips, and its styles need it.
		panel.style.width = '320px';
		panel.style.visibility = 'hidden';
		document.body.appendChild(panel);
		const follow = () => {
			if (!panel) return;
			place(panel, mouseX);
			frame = requestAnimationFrame(follow);
		};
		follow();
	});
	elem.addEventListener('mouseleave', stop);
	// The list rebuilds its rows, so a row can go away under the mouse.
	elem.addEventListener('click', stop);
}
