// The comp editor's small controls: an icon that opens a menu of choices (an assignment, a member's swap),
// and a segmented control (the raid size, the faction, the boss).

export interface Choice {
	value: string;
	label: string;
	icon?: string;
	// A Font Awesome icon, for a choice with no game icon (none, remove).
	faIcon?: string;
	// Muted words after the label, like a build's DPS.
	hint?: string;
	// A CSS class for the label, like a class's text color.
	labelClass?: string;
	danger?: boolean;
}

// The one menu open at a time.
let openMenu: { elem: HTMLElement; owner: HTMLElement; close: () => void } | null = null;

const closeOpenMenu = () => openMenu?.close();

document.addEventListener('mousedown', e => {
	if (openMenu && !openMenu.elem.contains(e.target as Node) && !openMenu.owner.contains(e.target as Node)) closeOpenMenu();
});
window.addEventListener('resize', closeOpenMenu);
document.addEventListener('scroll', e => openMenu && !openMenu.elem.contains(e.target as Node) && closeOpenMenu(), true);

// Opens a menu of `choices` under `owner`. The current one is marked. A click or Enter picks, Escape or a
// click elsewhere closes it, and the arrow keys move.
export function openChoiceMenu(owner: HTMLElement, opts: { title?: string; choices: Choice[]; value?: string; onPick: (value: string) => void }) {
	const wasOwn = openMenu?.owner == owner;
	closeOpenMenu();
	if (wasOwn) return;
	const items = opts.choices.map(c => {
		const item = (
			<button
				type="button"
				className={`rs-menu-item ${c.value == opts.value ? 'active' : ''} ${c.danger ? 'rs-menu-danger' : ''}`}
				attributes={{ role: 'menuitemradio', 'aria-checked': c.value == opts.value ? 'true' : 'false' }}>
				{choiceIcon(c)}
				<span className={`rs-menu-label ${c.labelClass ?? ''}`}>{c.label}</span>
				{c.hint ? <span className="rs-menu-hint">{c.hint}</span> : null}
			</button>
		) as HTMLButtonElement;
		item.addEventListener('click', () => {
			close();
			opts.onPick(c.value);
		});
		return item;
	});
	const elem = (
		<div className="rs-menu" attributes={{ role: 'menu' }}>
			{opts.title ? <div className="rs-menu-title">{opts.title}</div> : null}
			{items}
		</div>
	) as HTMLElement;
	elem.addEventListener('keydown', e => {
		const at = items.indexOf(document.activeElement as HTMLButtonElement);
		if (e.key == 'Escape') {
			e.preventDefault();
			close();
			owner.focus();
		} else if (e.key == 'ArrowDown' || e.key == 'ArrowUp') {
			e.preventDefault();
			const next = (at + (e.key == 'ArrowDown' ? 1 : -1) + items.length) % items.length;
			items[next]?.focus();
		} else if (e.key == 'Tab') {
			close();
		}
	});
	document.body.appendChild(elem);
	owner.setAttribute('aria-expanded', 'true');

	// Under the owner, or over it when there's no room below, and never past the window's right edge.
	const r = owner.getBoundingClientRect();
	const m = elem.getBoundingClientRect();
	const below = r.bottom + 4 + m.height <= window.innerHeight || r.top - 4 - m.height < 0;
	const top = below ? r.bottom + 4 : r.top - 4 - m.height;
	const left = Math.max(8, Math.min(r.left, window.innerWidth - m.width - 8));
	elem.style.top = `${top + window.scrollY}px`;
	elem.style.left = `${left + window.scrollX}px`;

	function close() {
		elem.remove();
		owner.setAttribute('aria-expanded', 'false');
		if (openMenu?.elem == elem) openMenu = null;
	}
	openMenu = { elem, owner, close };
	(items.find(i => i.classList.contains('active')) ?? items[0])?.focus();
}

export const choiceIcon = (c: { icon?: string; faIcon?: string }) =>
	c.icon ? (
		<img className="rs-icon" src={c.icon} alt="" draggable={false} />
	) : (
		<span className="rs-icon rs-icon-fa">
			<i className={`fas ${c.faIcon ?? 'fa-wand-magic-sparkles'}`} />
		</span>
	);

// An assignment as an icon: what's set by hand, with a gold ring, or Auto, dashed, with what the last run's
// resolver picked when we know it. A click opens the choices. `choices` holds Auto first, as value ''.
export function assignmentPicker(opts: {
	label: string;
	choices: Choice[];
	value: string;
	// What Auto came to in the last run, when it's of this raid.
	resolved?: Choice | null;
	onChange: (value: string) => void;
}): HTMLButtonElement {
	const byHand = opts.value !== '';
	const shown = byHand ? opts.choices.find(c => c.value == opts.value) : opts.resolved ?? undefined;
	const words = byHand ? `${shown?.label ?? opts.value}, set by hand` : shown ? `Auto: ${shown.label}` : 'Auto';
	const button = (
		<button
			type="button"
			className={`rs-pick ${byHand ? 'rs-pick-hand' : 'rs-pick-auto'} ${shown ? '' : 'rs-pick-empty'}`}
			title={`${opts.label}: ${words}`}
			attributes={{ 'aria-label': `${opts.label}: ${words}` }}>
			{shown ? choiceIcon(shown) : <i className="fas fa-wand-magic-sparkles rs-pick-auto-icon" />}
		</button>
	) as HTMLButtonElement;
	button.setAttribute('aria-haspopup', 'menu');
	button.setAttribute('aria-expanded', 'false');
	button.dataset.pick = opts.label;
	button.addEventListener('click', () =>
		openChoiceMenu(button, {
			title: opts.label,
			choices: opts.choices,
			value: opts.value,
			onPick: v => {
				opts.onChange(v);
				// The page draws its pickers again on a change, so the keyboard goes back to this one's new self.
				document.querySelector<HTMLElement>(`.rs-pick[data-pick="${CSS.escape(opts.label)}"]`)?.focus();
			},
		}),
	);
	return button;
}

// A segmented control: one button per option, the current one lit.
export function segmented(options: { value: string; label: string; title?: string }[], value: string, onChange: (value: string) => void, ariaLabel: string) {
	const group = (<div className="rs-segmented" attributes={{ role: 'radiogroup', 'aria-label': ariaLabel }} />) as HTMLElement;
	options.forEach(o => {
		const b = (
			<button
				type="button"
				className={`rs-segment ${o.value == value ? 'active' : ''}`}
				title={o.title ?? ''}
				attributes={{ role: 'radio', 'aria-checked': o.value == value ? 'true' : 'false' }}>
				{o.label}
			</button>
		) as HTMLButtonElement;
		b.addEventListener('click', () => o.value != value && onChange(o.value));
		group.appendChild(b);
	});
	return group;
}

// A labelled control in the toolbar.
export const toolbarItem = (label: string, control: Node) => (
	<div className="rs-toolbar-item">
		<span className="rs-toolbar-label">{label}</span>
		{control}
	</div>
);
