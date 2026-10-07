// Alt+click pins the tooltips that show right now, so we can keep one around while we look at
// something else. Comparing two items is the usual case: we pin the tooltip of the equipped
// item and hover the other one.
//
// A pinned tooltip is a copy of the one on the page, at the same spot. We can drag it anywhere
// and close it with the button in its top right corner. The copy doesn't change after we pin
// it, so a tooltip that lists changes (like a modified preset's) shows them as they were.
//
// Tooltips come from two places, and we pin both kinds:
// - tippy, for our own tooltips. tippy puts each one in a [data-tippy-root] element in body.
// - wowhead's script, for items and spells. It puts them in .wowhead-tooltip elements in body.
//
// The tooltips usually show over something we can click, like an item in the gear picker. The
// Alt+click only pins, so we stop the click from reaching the page. Without that, Chrome would
// also download the wowhead link of the item, which is what it does on Alt+click.

let installed = false;

// Each pinned tooltip we bring to the front goes above the ones before it.
let topZIndex = 10000;

export function installStickyTooltips() {
	if (installed) return;
	installed = true;

	// The mousedown pins the tooltips, and we drop the click that follows it. tippy hides its
	// tooltip on mousedown, so we have to copy it before that, in the capture phase.
	let dropClick = false;
	window.addEventListener(
		'mousedown',
		event => {
			dropClick = false;
			if (!event.altKey || event.button !== 0) return;
			if ((event.target as Element | null)?.closest('.sticky-tooltip')) return;
			const tooltips = visibleTooltips();
			if (!tooltips.length) return;
			event.preventDefault();
			event.stopPropagation();
			dropClick = true;
			tooltips.forEach(pin);
		},
		true,
	);
	window.addEventListener(
		'click',
		event => {
			if (!dropClick) return;
			dropClick = false;
			event.preventDefault();
			event.stopPropagation();
		},
		true,
	);
}

function visibleTooltips(): HTMLElement[] {
	const isShown = (elem: HTMLElement) => getComputedStyle(elem).visibility === 'visible' && elem.getBoundingClientRect().width > 0;
	const tippyBoxes = Array.from(document.querySelectorAll<HTMLElement>('[data-tippy-root] > .tippy-box[data-state="visible"]')).filter(box =>
		isShown(box.parentElement!),
	);
	// wowhead's "Item Not Found" isn't worth pinning. For our made up items the tooltip from
	// our own item database shows next to it, and we pin that one.
	const wowheadTooltips = Array.from(document.querySelectorAll<HTMLElement>('body > .wowhead-tooltip:not([data-status="error"])')).filter(isShown);
	return [...tippyBoxes, ...wowheadTooltips];
}

function pin(tooltip: HTMLElement) {
	const rect = tooltip.getBoundingClientRect();
	const copy = tooltip.cloneNode(true) as HTMLElement;
	copy.removeAttribute('id');
	copy.querySelector('.tippy-arrow')?.remove();
	// wowhead places its tooltip with an absolute position on the page. In the pinned copy it
	// sits where the wrapper is.
	if (copy.classList.contains('wowhead-tooltip')) {
		copy.style.position = 'relative';
		copy.style.top = '';
		copy.style.left = '';
	}

	const closeButton = (
		<button className="sticky-tooltip-close" attributes={{ 'aria-label': 'Close' }}>
			<i className="fas fa-times" />
		</button>
	) as HTMLButtonElement;
	const sticky = (
		<div className="sticky-tooltip">
			{copy}
			{closeButton}
		</div>
	) as HTMLElement;
	sticky.style.left = `${rect.left}px`;
	sticky.style.top = `${rect.top}px`;
	sticky.style.zIndex = String(++topZIndex);
	closeButton.addEventListener('click', () => sticky.remove());
	makeDraggable(sticky);
	document.body.appendChild(sticky);
}

// Drags the pinned tooltip by any part of it except the close button. We keep at least a corner
// of it on the screen, so we can always grab it again.
function makeDraggable(sticky: HTMLElement) {
	const MIN_VISIBLE = 24;
	sticky.addEventListener('pointerdown', event => {
		if (event.button !== 0 || (event.target as Element).closest('.sticky-tooltip-close')) return;
		event.preventDefault();
		sticky.style.zIndex = String(++topZIndex);
		sticky.classList.add('dragging');
		sticky.setPointerCapture(event.pointerId);
		const startX = event.clientX - sticky.offsetLeft;
		const startY = event.clientY - sticky.offsetTop;

		const move = (moveEvent: PointerEvent) => {
			const left = Math.min(Math.max(moveEvent.clientX - startX, MIN_VISIBLE - sticky.offsetWidth), window.innerWidth - MIN_VISIBLE);
			const top = Math.min(Math.max(moveEvent.clientY - startY, 0), window.innerHeight - MIN_VISIBLE);
			sticky.style.left = `${left}px`;
			sticky.style.top = `${top}px`;
		};
		const stop = () => {
			sticky.classList.remove('dragging');
			sticky.removeEventListener('pointermove', move);
			sticky.removeEventListener('pointerup', stop);
			sticky.removeEventListener('pointercancel', stop);
		};
		sticky.addEventListener('pointermove', move);
		sticky.addEventListener('pointerup', stop);
		sticky.addEventListener('pointercancel', stop);
	});
}
