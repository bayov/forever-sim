import tippy, { Instance, Props } from 'tippy.js';

// badgeTooltip gives a "!" badge on an icon its own tooltip, and hides the icon's tooltips
// while the mouse is on the badge.
//
// The badge sits inside the icon's element, so the mouse on the badge is also on the icon.
// Without this, wowhead's spell or item tooltip and the icon's own tippy tooltip show next to
// the badge's note, like on a rotation row's icon with a starting totem's warning. We hide
// them while the mouse is on the badge, and show the tippy one again when the mouse goes back
// to the icon. wowhead shows its own again by itself.
export function badgeTooltip(badge: HTMLElement, props: Partial<Props>): Instance {
	watchBadgeHover();
	badge.classList.add('tooltip-badge');

	const paused: Instance[] = [];
	badge.addEventListener('mouseenter', () => {
		for (let elem = badge.parentElement; elem; elem = elem.parentElement) {
			const instance = (elem as { _tippy?: Instance })._tippy;
			if (instance?.state.isEnabled) {
				instance.disable();
				paused.push(instance);
			}
		}
	});
	badge.addEventListener('mouseleave', () => {
		paused.splice(0).forEach(instance => {
			instance.enable();
			if (instance.reference.matches(':hover')) instance.show();
		});
	});

	return tippy(badge, props);
}

// wowhead shows its tooltip on any mouse move over a link, and the badge is inside the link
// on a rotation row. So we hide wowhead's tooltips with a class on the body while the mouse is
// on a badge. We set the class on every mouse move instead of on the badge's mouseleave,
// because a badge that goes away under the mouse gets no mouseleave, and the class would
// stay.
let watchingBadgeHover = false;
function watchBadgeHover() {
	if (watchingBadgeHover) return;
	watchingBadgeHover = true;
	window.addEventListener(
		'mouseover',
		event => {
			const onBadge = event.target instanceof Element && !!event.target.closest('.tooltip-badge');
			document.body.classList.toggle('tooltip-badge-hovered', onBadge);
		},
		{ capture: true, passive: true },
	);
}
