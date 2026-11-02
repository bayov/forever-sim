// The game tooltip: spells, items and auras with Wowhead's Forever tooltip text (through the
// practice server's cache), units and plain one-liners. It shows at the user's fixed tooltip
// spot, like EllesmereUI's.

type TooltipInfo = { name?: string; icon?: string; tooltip?: string; buff?: string };

const infos = new Map<string, Promise<TooltipInfo>>();

export function tooltipInfo(kind: 'spell' | 'item', id: number): Promise<TooltipInfo> {
	const key = `${kind}:${id}`;
	let p = infos.get(key);
	if (!p) {
		p = fetch(`/api/tooltip/${kind}/${id}`)
			.then(r => (r.ok ? r.json() : {}))
			.catch(() => ({}));
		infos.set(key, p);
	}
	return p;
}

// Wowhead's HTML, with only the tags and classes a tooltip needs. Links become plain text.
const allowed = new Set(['TABLE', 'TBODY', 'TR', 'TD', 'TH', 'DIV', 'SPAN', 'B', 'BR', 'A', 'SMALL']);

function sanitize(html: string): DocumentFragment {
	const doc = new DOMParser().parseFromString(`<body>${html}</body>`, 'text/html');
	const out = document.createDocumentFragment();
	const copy = (node: Node, into: Node) => {
		for (const child of [...node.childNodes]) {
			if (child.nodeType === Node.TEXT_NODE) {
				into.appendChild(document.createTextNode(child.textContent ?? ''));
			} else if (child.nodeType === Node.ELEMENT_NODE) {
				const el = child as Element;
				if (!allowed.has(el.tagName)) {
					copy(el, into);
					continue;
				}
				const tag = el.tagName === 'A' ? 'span' : el.tagName.toLowerCase();
				const made = document.createElement(tag);
				const cls = el.getAttribute('class');
				if (cls) made.className = cls;
				// The "N seconds remaining" line of an aura's tooltip is ours to fill.
				if (el.tagName === 'SPAN' && /remaining/.test(el.textContent ?? '')) continue;
				copy(el, made);
				into.appendChild(made);
			}
		}
	};
	copy(doc.body, out);
	return out;
}

export type TooltipContent =
	| { kind: 'spell' | 'item'; id: number; extra?: string[] }
	| { kind: 'aura'; spellId?: number; itemId?: number; label: string; remaining?: () => string }
	| { kind: 'lines'; title: string; lines?: string[]; titleClass?: string };

export class Tooltip {
	readonly el = document.createElement('div');
	private shown?: TooltipContent;
	private timer?: ReturnType<typeof setInterval>;

	constructor(
		parent: HTMLElement,
		private at: () => { left: number; top: number },
	) {
		this.el.className = 'tooltip hidden';
		parent.append(this.el);
	}

	// attach shows content while the mouse is over an element.
	attach(el: HTMLElement, content: () => TooltipContent | undefined) {
		el.addEventListener('mouseenter', () => {
			const c = content();
			if (c) this.show(c);
		});
		el.addEventListener('mouseleave', () => this.hide());
	}

	async show(content: TooltipContent) {
		this.shown = content;
		clearInterval(this.timer);
		const body = document.createElement('div');
		if (content.kind === 'lines') {
			const title = document.createElement('div');
			title.className = `tt-title ${content.titleClass ?? ''}`;
			title.textContent = content.title;
			body.append(title);
			for (const line of content.lines ?? []) {
				const d = document.createElement('div');
				d.className = 'tt-line';
				d.textContent = line;
				body.append(d);
			}
		} else if (content.kind === 'aura') {
			const info = content.spellId ? await tooltipInfo('spell', content.spellId) : content.itemId ? await tooltipInfo('item', content.itemId) : {};
			if (this.shown !== content) return;
			if (info.buff) {
				body.append(sanitize(info.buff));
			} else {
				const title = document.createElement('div');
				title.className = 'tt-title';
				title.textContent = info.name ?? content.label;
				body.append(title);
			}
			if (content.remaining) {
				const left = document.createElement('div');
				left.className = 'tt-remaining';
				body.append(left);
				const tick = () => (left.textContent = content.remaining!());
				tick();
				this.timer = setInterval(tick, 200);
			}
		} else {
			const info = await tooltipInfo(content.kind, content.id);
			if (this.shown !== content) return;
			if (info.tooltip) body.append(sanitize(info.tooltip));
			else {
				const title = document.createElement('div');
				title.className = 'tt-title';
				title.textContent = info.name ?? `${content.kind} ${content.id}`;
				body.append(title);
			}
			for (const line of content.extra ?? []) {
				const d = document.createElement('div');
				d.className = 'tt-extra';
				d.textContent = line;
				body.append(d);
			}
		}
		this.el.replaceChildren(body);
		this.el.classList.remove('hidden');
		const { left, top } = this.at();
		// The fixed spot is the tooltip's centre, kept on screen.
		const w = this.el.offsetWidth;
		const h = this.el.offsetHeight;
		this.el.style.left = `${Math.max(4, Math.min(innerWidth - w - 4, left - w / 2))}px`;
		this.el.style.top = `${Math.max(4, Math.min(innerHeight - h - 4, top - h / 2))}px`;
	}

	hide() {
		this.shown = undefined;
		clearInterval(this.timer);
		this.el.classList.add('hidden');
	}
}
