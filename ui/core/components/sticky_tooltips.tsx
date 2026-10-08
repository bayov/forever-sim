import { Instance } from 'tippy.js';

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
//
// Holding Alt keeps the tooltips that show right now, even when the mouse moves away, so we can
// move to a link inside one or to another item. They go away when we let go of Alt, unless we
// Alt+click to pin them.

let installed = false;

// A mouse with its left button pressed, for 'click' in a hint.
const LEFT_CLICK_ICON = `<svg class="left-click-icon" viewBox="0 0 12 18" aria-label="click" role="img">
	<rect x="0.75" y="0.75" width="10.5" height="16.5" rx="5.25" fill="none" stroke="currentColor" stroke-width="1.5" />
	<path d="M6 1.5 A4.5 4.5 0 0 0 1.5 6 V7.5 H6 Z" class="left-click-icon-button" />
	<path d="M6 0.75 V7.5 M0.75 7.5 H11.25" fill="none" stroke="currentColor" stroke-width="1.2" />
</svg>`;

// A short hint for a key with a click, like 'alt + click - pin tooltip', with the key drawn as
// a key on a keyboard and the click as a mouse.
export function keyClickHint(key: string, action: string, className = ''): HTMLElement {
	const click = document.createElement('span');
	click.innerHTML = LEFT_CLICK_ICON;
	return (
		<p className={`key-hint ${className}`} attributes={{ 'data-key': key }}>
			<kbd>{key}</kbd>
			<span>+</span>
			{click.firstElementChild!}
			<span className="key-hint-action">to {action}</span>
		</p>
	) as HTMLElement;
}

// The menus that open on a click, like the dropdowns, aren't tooltips we'd pin.
const UNPINNABLE_THEMES = ['dropdown-tooltip', 'bonus-stats-popover'];

// Shorter plain text than this isn't worth pinning.
const MIN_PINNABLE_TEXT = 150;

// A tooltip of ours is worth pinning when it holds more than a short line: an item, a talent
// with its rank and description, a list of changes or a table of stats. A short plain one, like
// 'Reset talent points' on a button, says all it has to say while we hover it, so it gets no
// hint and Alt+click doesn't pin it.
//
// We tell them apart by the content. It's worth pinning when it has a picture, a table or a
// list, when it has two or more blocks (like a talent's name, rank and description), or when
// its text is long.
function isWorthPinning(content: Element): boolean {
	const parts = Array.from(content.querySelectorAll('*')).filter(part => !part.closest('.pin-hint'));
	if (parts.some(part => ['IMG', 'TABLE', 'UL', 'OL'].includes(part.tagName))) return true;
	if (parts.filter(part => !getComputedStyle(part).display.startsWith('inline')).length >= 2) return true;
	const text = Array.from(content.childNodes)
		.filter(node => !(node instanceof Element && node.classList.contains('pin-hint')))
		.map(node => node.textContent)
		.join('');
	return text.trim().length >= MIN_PINNABLE_TEXT;
}

// Every tooltip we can pin ends with a hint that says how. tippy and wowhead both fill their
// tooltips when they show, and again when what they show changes, so we watch the page and add
// the hint back at the end each time. A pinned copy hides it, see .sticky-tooltip.
//
// We add the hint right when the tooltip changes, before the browser draws it. When we waited
// for the next frame, a tooltip above its element first showed without the hint, then grew and
// jumped up by a line.
function addPinHints() {
	const ensureHint = (content: Element | null) => {
		if (!content) return;
		const last = content.lastElementChild;
		if (last?.classList.contains('pin-hint')) return;
		const existing = content.querySelector(':scope > .pin-hint');
		content.appendChild(existing ?? keyClickHint('alt', 'pin tooltip', 'pin-hint'));
	};
	const removeHint = (content: Element) => content.querySelector(':scope > .pin-hint')?.remove();

	const update = () => {
		document.querySelectorAll<HTMLElement>('[data-tippy-root] > .tippy-box > .tippy-content').forEach(content => {
			if (isPinnableTippy(content.parentElement!)) ensureHint(content);
			else removeHint(content);
		});
		document.querySelectorAll<HTMLElement>('body > .wowhead-tooltip:not([data-status="error"])').forEach(tooltip => {
			ensureHint(tooltip.querySelector(':scope > table > tbody > tr > td'));
		});
	};
	// Most changes on the page aren't in a tooltip, so we skip those.
	const TOOLTIP = '[data-tippy-root], body > .wowhead-tooltip';
	const touchesTooltip = (record: MutationRecord) =>
		(record.target instanceof Element && !!record.target.closest(TOOLTIP)) ||
		Array.from(record.addedNodes).some(node => node instanceof Element && node.matches(TOOLTIP));
	new MutationObserver(records => {
		if (records.some(touchesTooltip)) update();
	}).observe(document.body, { childList: true, subtree: true });
}

// Each pinned tooltip we bring to the front goes above the ones before it.
let topZIndex = 10000;

// Puts the alt-held and ctrl-held classes on the page while we hold Alt or Ctrl (or Cmd on a
// Mac). The hints for that key light up then, and a run in the sim history shows the hand
// cursor, because a click loads its settings.
//
// We also check the keys on each mouse move, for when we pressed one while the page didn't
// have focus. When the window loses focus we clear them. Otherwise Alt+Tab to another window
// would leave alt-held on, because we never see the Alt key go up.
function trackModifierKeys() {
	const update = (event: KeyboardEvent | MouseEvent) => {
		document.documentElement.classList.toggle('alt-held', event.altKey);
		document.documentElement.classList.toggle('ctrl-held', event.ctrlKey || event.metaKey);
	};
	window.addEventListener('keydown', update, true);
	window.addEventListener('keyup', update, true);
	window.addEventListener('mousemove', update, { capture: true, passive: true });
	window.addEventListener('blur', () => document.documentElement.classList.remove('alt-held', 'ctrl-held'));
}

// The copies of the tooltips we keep while we hold Alt, and how to bring back the tooltips
// they stand for when we let go.
let held: { copies: HTMLElement[]; showSources: () => void } | null = null;

// We copy the tooltips like a pin does, but the copies let the mouse through and go away when
// we let go of Alt. The tooltip under the mouse shows again then, when the mouse is still on
// the element we pointed at.
function holdTooltips(mouseTarget: Element | null) {
	const tooltips = visibleTooltips();
	if (!tooltips.length) return;
	held = { copies: tooltips.map(tooltip => pin(tooltip, true)), showSources: hidePinnedSources(mouseTarget) };
}

// We show the tooltips again before we remove the copies, and the copies go only after the
// browser drew the tooltips. Otherwise there's a frame with neither, and the tooltip blinks.
function releaseTooltips() {
	if (!held) return;
	const { copies, showSources } = held;
	held = null;
	showSources();
	afterNextPaint(() => copies.forEach(copy => copy.remove()));
}

function afterNextPaint(callback: () => void) {
	requestAnimationFrame(() => requestAnimationFrame(callback));
}

// wowhead's tooltips link their set pieces and item effects with paths of its own site, like
// '/forever/item=11729/savage-gladiator-helm'. On our page that path points to us, so a click
// on one would open our own page. When the mouse comes onto such a link (in a tooltip, or in a
// pinned copy of one), we point it at wowhead and open it in a new tab, like a click on an icon.
function fixWowheadTooltipLinks() {
	window.addEventListener(
		'mouseover',
		event => {
			const link = (event.target as Element | null)?.closest<HTMLAnchorElement>('.wowhead-tooltip a[href^="/"]');
			if (!link) return;
			link.href = `https://www.wowhead.com${link.getAttribute('href')}`;
			link.target = '_blank';
		},
		{ capture: true, passive: true },
	);
}

export function installStickyTooltips() {
	if (installed) return;
	installed = true;
	addPinHints();
	trackModifierKeys();
	fixWowheadTooltipLinks();

	let mouseTarget: Element | null = null;
	window.addEventListener('mouseover', event => (mouseTarget = event.target as Element | null), { capture: true, passive: true });
	window.addEventListener(
		'keydown',
		event => {
			if (event.key !== 'Alt') return;
			// Firefox opens its menu bar when we let go of Alt, unless we stop the key.
			event.preventDefault();
			if (!event.repeat && !held) holdTooltips(mouseTarget);
		},
		true,
	);
	window.addEventListener(
		'keyup',
		event => {
			if (event.key !== 'Alt') return;
			event.preventDefault();
			releaseTooltips();
		},
		true,
	);
	window.addEventListener('blur', releaseTooltips);

	// The mousedown pins the tooltips, and we drop the click that follows it. tippy hides its
	// tooltip on mousedown, so we have to copy it before that, in the capture phase.
	let dropClick = false;
	window.addEventListener(
		'mousedown',
		event => {
			dropClick = false;
			if (!event.altKey || event.button !== 0) return;
			if ((event.target as Element | null)?.closest('.sticky-tooltip')) return;
			// The tooltips we keep while we hold Alt become pins, next to the ones that show now.
			const kept = held?.copies ?? [];
			held = null;
			kept.forEach(copy => {
				copy.classList.remove('sticky-tooltip-held');
				copy.style.zIndex = String(++topZIndex);
			});
			const tooltips = visibleTooltips();
			if (!tooltips.length && !kept.length) return;
			event.preventDefault();
			event.stopPropagation();
			dropClick = true;
			tooltips.forEach(tooltip => pin(tooltip));
			hidePinnedSources(event.target as Element | null);
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

const isShown = (elem: HTMLElement) => getComputedStyle(elem).visibility === 'visible' && elem.getBoundingClientRect().width > 0;

const visibleTippyBoxes = () =>
	Array.from(document.querySelectorAll<HTMLElement>('[data-tippy-root] > .tippy-box[data-state="visible"]')).filter(box => isShown(box.parentElement!));

function isPinnableTippy(box: HTMLElement): boolean {
	const theme = box.dataset.theme ?? '';
	if (UNPINNABLE_THEMES.some(unpinnable => theme.split(' ').includes(unpinnable))) return false;
	const content = box.querySelector(':scope > .tippy-content');
	return !!content && isWorthPinning(content);
}

function visibleTooltips(): HTMLElement[] {
	// wowhead's "Item Not Found" isn't worth pinning. For our made up items the tooltip from
	// our own item database shows next to it, and we pin that one.
	const wowheadTooltips = Array.from(document.querySelectorAll<HTMLElement>('body > .wowhead-tooltip:not([data-status="error"])')).filter(isShown);
	return [...visibleTippyBoxes().filter(isPinnableTippy), ...wowheadTooltips];
}

// Makes the pinned copy of a tooltip, or the copy we keep while we hold Alt.
function pin(tooltip: HTMLElement, whileAltHeld = false): HTMLElement {
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
		<div className={`sticky-tooltip ${whileAltHeld ? 'sticky-tooltip-held' : ''}`}>
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
	return sticky;
}

// The pinned copy takes the place of the tooltip we hovered, so we hide that one. It shows
// again once the mouse leaves the element we pointed at and comes back to it.
//
// tippy does this for us when we hide its tooltip. wowhead shows its tooltip again on any mouse
// move over the link, so we keep it hidden with a class until the mouse leaves the link. That
// includes wowhead's "Item Not Found", which we don't pin but which shows next to the tooltip
// from our own item database.
//
// It returns a function that shows them right away, for when we let go of Alt. It shows a tippy
// tooltip again only when the mouse is still on its element.
function hidePinnedSources(target: Element | null): () => void {
	const instances = visibleTippyBoxes()
		.filter(isPinnableTippy)
		.map(box => (box.parentElement as { _tippy?: Instance })._tippy)
		.filter((instance): instance is Instance => !!instance);
	instances.forEach(instance => instance.hide());

	const wowheadTooltips = Array.from(document.querySelectorAll<HTMLElement>('body > .wowhead-tooltip')).filter(isShown);
	const showWowhead = () => wowheadTooltips.forEach(tooltip => tooltip.classList.remove('sticky-tooltip-source-hidden'));
	if (wowheadTooltips.length) {
		wowheadTooltips.forEach(tooltip => tooltip.classList.add('sticky-tooltip-source-hidden'));
		const link = target?.closest('a, [data-wowhead]') ?? target;
		if (link) link.addEventListener('mouseleave', showWowhead, { once: true });
		else showWowhead();
	}
	// tippy fades its tooltips in. The copy is already there, so we show the tooltip right away,
	// without the fade.
	return () => {
		instances
			.filter(instance => instance.reference.matches(':hover'))
			.forEach(instance => {
				const duration = instance.props.duration;
				instance.setProps({ duration: 0 });
				instance.show();
				afterNextPaint(() => instance.setProps({ duration }));
			});
		showWowhead();
	};
}

// Drags the pinned tooltip by any part of it except the close button and its links, which we
// click to open. We keep at least a corner of it on the screen, so we can always grab it again.
function makeDraggable(sticky: HTMLElement) {
	const MIN_VISIBLE = 24;
	sticky.addEventListener('pointerdown', event => {
		if (event.button !== 0 || (event.target as Element).closest('.sticky-tooltip-close, a[href]')) return;
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
