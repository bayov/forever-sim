import { ref } from 'tsx-vanilla';

import { EventFilter } from './events_filter.js';
import {
	formatClock,
	formatSpan,
	LevelingSession,
	mobName,
	ResourceSample,
	SECTION_LABELS,
	SectionKind,
	SESSION_STATES,
	SessionSection,
	sectionTitle,
} from './leveling_session.js';

const SECTION_KINDS: Array<SectionKind> = ['pull', 'rest', 'search', 'death'];
const SVG_NS = 'http://www.w3.org/2000/svg';
// A press that moves less than this many pixels is a click, more is a drag that selects a range.
const DRAG_PX = 4;

export interface SessionStripConfig {
	parent: HTMLElement;
	filter: EventFilter;
}

// The whole leveling session in one bar, above every view of the Events tab. It's how we move
// around a 20 min session: the bar shows what we did when (fighting, resting, searching, dead),
// with our health and mana under it and a tick for each kill.
//
// Clicking a part of the bar picks it: a pull, the rest after it, a search or a death. The pick
// is the filter's From and To, so the timeline, the log, the pulls and the replay all show that
// part. Dragging picks any range. The arrows go to the previous and next part of the kind we
// chose (pulls by default), and "Whole session" clears the pick.
export class SessionStrip {
	readonly rootElem: HTMLElement;
	private readonly filter: EventFilter;
	private session: LevelingSession | null = null;
	private kind: SectionKind = 'pull';
	private cursor: number | null = null;
	private view: { start: number; end: number } | null = null;

	private readonly ui: {
		kindButtons: Map<SectionKind, HTMLButtonElement>;
		prev: HTMLButtonElement;
		next: HTMLButtonElement;
		all: HTMLButtonElement;
		title: HTMLElement;
		detail: HTMLElement;
		stats: HTMLElement;
		bar: HTMLElement;
		spans: HTMLElement;
		marks: HTMLElement;
		chart: SVGSVGElement;
		ruler: HTMLElement;
		selection: HTMLElement;
		viewBox: HTMLElement;
		cursor: HTMLElement;
		hover: HTMLElement;
		tooltip: HTMLElement;
	};

	constructor(config: SessionStripConfig) {
		this.filter = config.filter;
		const refs = {
			prev: ref<HTMLButtonElement>(),
			next: ref<HTMLButtonElement>(),
			all: ref<HTMLButtonElement>(),
			title: ref<HTMLSpanElement>(),
			detail: ref<HTMLSpanElement>(),
			stats: ref<HTMLSpanElement>(),
			bar: ref<HTMLDivElement>(),
			spans: ref<HTMLDivElement>(),
			marks: ref<HTMLDivElement>(),
			ruler: ref<HTMLDivElement>(),
			selection: ref<HTMLDivElement>(),
			viewBox: ref<HTMLDivElement>(),
			cursor: ref<HTMLDivElement>(),
			hover: ref<HTMLDivElement>(),
			tooltip: ref<HTMLDivElement>(),
		};
		const kindButtons = new Map<SectionKind, HTMLButtonElement>();
		const chart = document.createElementNS(SVG_NS, 'svg') as SVGSVGElement;
		chart.classList.add('session-strip-chart');
		chart.setAttribute('preserveAspectRatio', 'none');

		this.rootElem = (
			<div className="session-strip hide">
				<div className="session-strip-nav">
					<div className="segmented session-strip-kinds" attributes={{ role: 'group', 'aria-label': 'Go through' }}>
						{SECTION_KINDS.map(kind => {
							const button = (
								<button className="segmented-option" type="button">
									<span className={`session-strip-swatch kind-${kind}`} />
									<span className="session-strip-kind-label">{SECTION_LABELS[kind].many}</span>
									<span className="session-strip-kind-count"></span>
								</button>
							) as HTMLButtonElement;
							button.addEventListener('click', () => this.setKind(kind));
							kindButtons.set(kind, button);
							return button;
						})}
					</div>
					<div className="session-strip-step">
						<button
							ref={refs.prev}
							className="btn btn-outline-secondary session-strip-arrow"
							type="button"
							attributes={{ 'aria-label': 'Previous' }}>
							<i className="fas fa-chevron-left" />
						</button>
						<div className="session-strip-current">
							<span ref={refs.title} className="session-strip-title"></span>
							<span ref={refs.detail} className="session-strip-detail"></span>
						</div>
						<button ref={refs.next} className="btn btn-outline-secondary session-strip-arrow" type="button" attributes={{ 'aria-label': 'Next' }}>
							<i className="fas fa-chevron-right" />
						</button>
					</div>
					<button ref={refs.all} className="btn btn-link session-strip-all" type="button">
						<i className="fas fa-expand me-1" />
						Whole session
					</button>
					<span ref={refs.stats} className="session-strip-stats"></span>
				</div>
				<div ref={refs.bar} className="session-strip-bar" attributes={{ tabindex: '0', role: 'slider', 'aria-label': 'Session' }}>
					<div ref={refs.spans} className="session-strip-spans" />
					{chart}
					<div ref={refs.marks} className="session-strip-marks" />
					<div ref={refs.viewBox} className="session-strip-view hide" />
					<div ref={refs.selection} className="session-strip-selection hide" />
					<div ref={refs.hover} className="session-strip-hover hide" />
					<div ref={refs.cursor} className="session-strip-cursor hide" />
					<div ref={refs.tooltip} className="session-strip-tooltip hide" />
				</div>
				<div ref={refs.ruler} className="session-strip-ruler" />
			</div>
		) as HTMLElement;
		config.parent.appendChild(this.rootElem);

		this.ui = {
			kindButtons,
			prev: refs.prev.value!,
			next: refs.next.value!,
			all: refs.all.value!,
			title: refs.title.value!,
			detail: refs.detail.value!,
			stats: refs.stats.value!,
			bar: refs.bar.value!,
			spans: refs.spans.value!,
			marks: refs.marks.value!,
			chart,
			ruler: refs.ruler.value!,
			selection: refs.selection.value!,
			viewBox: refs.viewBox.value!,
			cursor: refs.cursor.value!,
			hover: refs.hover.value!,
			tooltip: refs.tooltip.value!,
		};

		this.ui.prev.addEventListener('click', () => this.step(-1));
		this.ui.next.addEventListener('click', () => this.step(1));
		this.ui.all.addEventListener('click', () => this.filter.setRange(null, null));
		this.ui.bar.addEventListener('keydown', event => {
			if (event.key == 'ArrowLeft' || event.key == '[') this.step(-1);
			else if (event.key == 'ArrowRight' || event.key == ']') this.step(1);
			else if (event.key == 'Escape') this.filter.setRange(null, null);
			else return;
			event.preventDefault();
		});
		this.bindPointer();
		this.filter.changeEmitter.on(() => this.updateSelection());
	}

	setSession(session: LevelingSession | null, averageKills?: number) {
		this.session = session;
		this.rootElem.classList.toggle('hide', !session);
		if (!session) return;
		this.drawSpans(session);
		this.drawChart(session);
		this.drawMarks(session);
		this.drawRuler(session);

		const kills = session.kills.length;
		const deaths = session.deaths.length;
		this.ui.stats.textContent = [
			`${kills} kills${averageKills != null ? ` (average ${averageKills.toFixed(1)})` : ''}`,
			`${session.pulls.length} pulls`,
			`${deaths} ${deaths == 1 ? 'death' : 'deaths'}`,
		].join(' · ');
		this.ui.kindButtons.forEach((button, kind) => {
			const count = session.sections.filter(s => s.kind == kind).length;
			button.querySelector('.session-strip-kind-count')!.textContent = String(count);
			button.disabled = count == 0;
		});
		if (this.ui.kindButtons.get(this.kind)!.disabled) this.kind = 'pull';
		this.updateSelection();
	}

	// The replay's time, as a line on the bar. Null hides it.
	setCursor(time: number | null) {
		this.cursor = time;
		this.placeCursor();
	}

	// The part of the session the timeline shows, as a box on the bar. Null hides it.
	setView(view: { start: number; end: number } | null) {
		this.view = view;
		const session = this.session;
		const box = this.ui.viewBox;
		if (!session || !view || (view.start <= 0 && view.end >= session.length)) {
			box.classList.add('hide');
			return;
		}
		box.classList.remove('hide');
		box.style.left = this.percent(Math.max(0, view.start));
		box.style.width = this.percent(Math.min(session.length, view.end) - Math.max(0, view.start));
	}

	private placeCursor() {
		const session = this.session;
		const visible = !!session && this.cursor != null;
		this.ui.cursor.classList.toggle('hide', !visible);
		if (visible) this.ui.cursor.style.left = this.percent(this.cursor!);
	}

	private percent(t: number): string {
		return `${((100 * t) / this.session!.length).toFixed(4)}%`;
	}

	private setKind(kind: SectionKind) {
		this.kind = kind;
		this.updateSelection();
		this.ui.bar.focus({ preventScroll: true });
	}

	private sectionsOfKind(): Array<SessionSection> {
		return this.session?.sections.filter(s => s.kind == this.kind) ?? [];
	}

	// The section the filter range is, if it is one.
	private selectedSection(): SessionSection | null {
		const { from, to } = this.filter;
		if (!this.session || from == null || to == null) return null;
		return this.session.sections.find(s => Math.abs(s.start - from) < 0.005 && Math.abs(s.end - to) < 0.005) ?? null;
	}

	// The previous or next section of our kind, from the one we picked, or from the middle of the
	// range we look at.
	private step(direction: 1 | -1) {
		const sections = this.sectionsOfKind();
		if (!sections.length) return;
		const selected = this.selectedSection();
		let index: number;
		if (selected && selected.kind == this.kind) {
			index = sections.indexOf(selected) + direction;
		} else {
			const { from, to } = this.filter;
			const at = from != null || to != null ? ((from ?? 0) + (to ?? this.session!.length)) / 2 : this.cursor ?? -1;
			index = direction > 0 ? sections.findIndex(s => s.start > at) : findLastIndex(sections, s => s.end <= at || s.start < at);
			if (direction < 0 && index >= 0 && sections[index].start <= at && at < sections[index].end && sections[index] === selected) index--;
		}
		index = Math.max(0, Math.min(sections.length - 1, index < 0 ? (direction > 0 ? sections.length - 1 : 0) : index));
		this.pick(sections[index]);
	}

	private pick(section: SessionSection) {
		this.kind = section.kind;
		this.filter.setRange(round2(section.start), round2(section.end));
	}

	private updateSelection() {
		const session = this.session;
		if (!session) return;
		const { from, to } = this.filter;
		const has = from != null || to != null;
		const start = Math.max(0, from ?? 0);
		const end = Math.min(session.length, to ?? session.length);
		this.ui.selection.classList.toggle('hide', !has);
		this.ui.bar.classList.toggle('has-selection', has);
		if (has) {
			this.ui.selection.style.left = this.percent(start);
			this.ui.selection.style.width = this.percent(Math.max(0, end - start));
		}
		this.ui.all.classList.toggle('invisible', !has);

		const selected = this.selectedSection();
		if (selected) this.kind = selected.kind;
		this.ui.kindButtons.forEach((button, kind) => {
			button.classList.toggle('active', kind == this.kind);
			button.setAttribute('aria-pressed', String(kind == this.kind));
		});
		const sections = this.sectionsOfKind();
		if (selected) {
			this.ui.title.textContent = `${sectionTitle(selected)} of ${sections.length}`;
			this.ui.detail.textContent = `${formatClock(selected.start)} to ${formatClock(selected.end)} · ${sectionSummary(selected)}`;
		} else if (has) {
			this.ui.title.textContent = 'Range';
			this.ui.detail.textContent = `${formatClock(start)} to ${formatClock(end)} · ${formatSpan(end - start)}`;
		} else {
			this.ui.title.textContent = 'Whole session';
			this.ui.detail.textContent = `${formatClock(session.length)} · pick a part of the bar, or step through the ${SECTION_LABELS[
				this.kind
			].many.toLowerCase()}`;
		}
		this.ui.prev.disabled = !sections.length || (selected?.kind == this.kind && sections.indexOf(selected) == 0);
		this.ui.next.disabled = !sections.length || (selected?.kind == this.kind && sections.indexOf(selected) == sections.length - 1);
	}

	private drawSpans(session: LevelingSession) {
		this.ui.spans.replaceChildren(
			...session.spans.map(span => {
				const elem = (<div className={`session-strip-span state-${span.state}`} />) as HTMLElement;
				elem.style.left = this.percent(span.start);
				elem.style.width = this.percent(span.end - span.start);
				elem.style.background = SESSION_STATES[span.state].color;
				return elem;
			}),
		);
	}

	// Our health and mana under the states, one point a pixel at most.
	private drawChart(session: LevelingSession) {
		const width = 1000;
		const height = 100;
		const line = (samples: Array<ResourceSample>, className: string) => {
			if (!samples.length) return null;
			const buckets = new Map<number, number>();
			let last = samples[0].value;
			const points: Array<string> = [`0,${(height * (1 - last)).toFixed(1)}`];
			for (const sample of samples) {
				const x = Math.round((width * sample.t) / session.length);
				// The lowest value in a pixel, so a near death shows.
				buckets.set(x, Math.min(buckets.get(x) ?? Infinity, sample.value));
				last = sample.value;
			}
			// Steps, not slopes: a value holds until the next change, like our health through a
			// corpse run with no samples in it.
			let y = points[0].split(',')[1];
			[...buckets.entries()].forEach(([x, value]) => {
				points.push(`${x},${y}`);
				y = (height * (1 - value)).toFixed(1);
				points.push(`${x},${y}`);
			});
			points.push(`${width},${(height * (1 - last)).toFixed(1)}`);
			const polyline = document.createElementNS(SVG_NS, 'polyline');
			polyline.setAttribute('points', points.join(' '));
			polyline.setAttribute('class', className);
			polyline.setAttribute('vector-effect', 'non-scaling-stroke');
			return polyline;
		};
		this.ui.chart.setAttribute('viewBox', `0 0 ${width} ${height}`);
		this.ui.chart.replaceChildren(
			...[line(session.mana, 'session-strip-mana'), line(session.health, 'session-strip-health')].filter((e): e is SVGPolylineElement => !!e),
		);
	}

	private drawMarks(session: LevelingSession) {
		this.ui.marks.replaceChildren(
			...session.kills.map(kill => {
				const elem = (<div className="session-strip-kill" />) as HTMLElement;
				elem.style.left = this.percent(kill.t);
				return elem;
			}),
			...session.deaths.map(death => {
				const elem = (
					<div className="session-strip-death">
						<i className="fas fa-skull" />
					</div>
				) as HTMLElement;
				elem.style.left = this.percent(death.t);
				return elem;
			}),
		);
	}

	private drawRuler(session: LevelingSession) {
		const step = session.length > 1800 ? 300 : session.length > 600 ? 120 : 60;
		const labels: Array<HTMLElement> = [];
		for (let t = 0; t <= session.length + 1e-6; t += step) {
			const elem = (<span className="session-strip-tick">{formatClock(t)}</span>) as HTMLElement;
			elem.style.left = this.percent(t);
			labels.push(elem);
		}
		this.ui.ruler.replaceChildren(...labels);
	}

	// Hover shows what's under the mouse, a click picks it, and a drag picks a range.
	private bindPointer() {
		const bar = this.ui.bar;
		let down: { x: number; t: number } | null = null;
		let dragging = false;

		const timeAt = (clientX: number) => {
			const rect = bar.getBoundingClientRect();
			return Math.max(0, Math.min(1, (clientX - rect.left) / rect.width)) * (this.session?.length ?? 0);
		};

		bar.addEventListener('pointerdown', event => {
			if (!this.session || event.button != 0) return;
			down = { x: event.clientX, t: timeAt(event.clientX) };
			dragging = false;
			try {
				bar.setPointerCapture(event.pointerId);
			} catch {
				// A synthetic press has no pointer to capture, and the drag still works inside the bar.
			}
		});
		bar.addEventListener('pointermove', event => {
			if (!this.session) return;
			const t = timeAt(event.clientX);
			if (down && !dragging && Math.abs(event.clientX - down.x) > DRAG_PX) dragging = true;
			if (down && dragging) {
				const start = Math.min(down.t, t);
				const end = Math.max(down.t, t);
				this.ui.hover.classList.remove('hide');
				this.ui.hover.style.left = this.percent(start);
				this.ui.hover.style.width = this.percent(end - start);
				this.showTooltip(event.clientX, <>{`${formatClock(start)} to ${formatClock(end)} · ${formatSpan(end - start)}`}</>);
				return;
			}
			const section = this.sectionNear(t);
			if (section) {
				this.ui.hover.classList.remove('hide');
				this.ui.hover.style.left = this.percent(section.start);
				this.ui.hover.style.width = this.percent(section.end - section.start);
			} else {
				this.ui.hover.classList.add('hide');
			}
			this.showTooltip(event.clientX, this.tooltipFor(t, section));
		});
		const end = (event: PointerEvent) => {
			if (!down || !this.session) return;
			const t = timeAt(event.clientX);
			if (dragging) {
				const start = Math.min(down.t, t);
				const stop = Math.max(down.t, t);
				if (stop - start > 0.5) this.filter.setRange(round2(start), round2(stop));
			} else {
				const section = this.sectionNear(t);
				if (section) {
					const selected = this.selectedSection();
					// A second click on the picked part goes back to the whole session.
					if (selected === section) this.filter.setRange(null, null);
					else this.pick(section);
				}
			}
			down = null;
			dragging = false;
			this.ui.hover.classList.add('hide');
		};
		bar.addEventListener('pointerup', end);
		bar.addEventListener('pointercancel', () => {
			down = null;
			dragging = false;
		});
		bar.addEventListener('pointerleave', () => {
			if (down) return;
			this.ui.hover.classList.add('hide');
			this.ui.tooltip.classList.add('hide');
		});
	}

	// The section at t. The short waits that belong to no section count toward the next one.
	private sectionNear(t: number): SessionSection | null {
		const sections = this.session?.sections ?? [];
		return sections.find(s => s.start <= t && t < s.end) ?? sections.find(s => s.start > t) ?? null;
	}

	private tooltipFor(t: number, section: SessionSection | null): Element {
		const state = this.session!.stateAt(t);
		return (
			<>
				<div className="session-strip-tooltip-time">
					{formatClock(t, true)} · <span style={{ color: SESSION_STATES[state].color }}>{SESSION_STATES[state].label}</span>
				</div>
				{section && (
					<>
						<div className="session-strip-tooltip-title">
							{sectionTitle(section)} · {formatSpan(section.end - section.start)}
						</div>
						<div className="session-strip-tooltip-body">{sectionSummary(section, true)}</div>
					</>
				)}
			</>
		) as Element;
	}

	private showTooltip(clientX: number, content: Element) {
		const tooltip = this.ui.tooltip;
		tooltip.replaceChildren(content);
		tooltip.classList.remove('hide');
		const rect = this.ui.bar.getBoundingClientRect();
		const width = tooltip.offsetWidth;
		const x = Math.max(0, Math.min(rect.width - width, clientX - rect.left - width / 2));
		tooltip.style.left = `${x}px`;
	}
}

// What a section was, in a few words.
function sectionSummary(section: SessionSection, long = false): string {
	const pull = section.pull;
	switch (section.kind) {
		case 'pull': {
			if (!pull) return formatSpan(section.end - section.start);
			const mobs = pull.mobs.length;
			const parts = [`${mobs} ${mobs == 1 ? 'mob' : 'mobs'}`, `${pull.kills} killed`];
			if (long) {
				parts.push(
					pull.mobs.map(m => `${mobName(m)} (${m.level})`).join(', '),
					`health ${pct(pull.healthStart)} to ${pct(pull.healthEnd)}, lowest ${pct(pull.lowestHealth)}`,
				);
				if (pull.manaStart != null && pull.manaEnd != null) parts.push(`mana ${pct(pull.manaStart)} to ${pct(pull.manaEnd)}`);
			}
			if (pull.died) parts.push('died');
			return parts.join(' · ');
		}
		case 'rest': {
			const what = [section.ate && 'ate', section.drank && 'drank', section.bandaged && 'bandaged'].filter(Boolean).join(', ');
			return `${formatSpan(section.end - section.start)}${what ? ` · ${what}` : ''}`;
		}
		case 'search':
			return `${formatSpan(section.end - section.start)} looking for a mob`;
		case 'death':
			return `corpse run, ${formatSpan(section.end - section.start)}`;
	}
}

const pct = (value: number) => `${Math.round(100 * value)}%`;
const round2 = (t: number) => Math.round(t * 100) / 100;

function findLastIndex<T>(items: Array<T>, predicate: (item: T) => boolean): number {
	for (let i = items.length - 1; i >= 0; i--) if (predicate(items[i])) return i;
	return -1;
}
