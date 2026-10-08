import { TypedEvent } from '../typed_event';
import { badgeTooltip } from './badge_tooltip';

// Marks a buff that doesn't stack with another one that's on, like the consumables mark a
// scroll next to the raid buff of the same stat (see ConsumesPicker.markUnstacked).
//
// The icon gets a red edge and a red tint, and a "!" badge on its corner whose tooltip is the
// note. note() returns nothing while the buff is fine, and we check it again on each change.
export function markUnstacked(elem: HTMLElement, note: () => string | undefined, changed: TypedEvent<any>) {
	const badge = document.createElement('span');
	badge.classList.add('unstacked-badge');
	badge.textContent = '!';
	elem.appendChild(badge);
	const tooltip = badgeTooltip(badge, { theme: 'consumes-unstacked' });

	const update = () => {
		const text = note();
		elem.classList.toggle('buff-unstacked', !!text);
		badge.classList.toggle('hide', !text);
		tooltip.setContent(text ?? '');
	};
	update();
	changed.on(update);
}
