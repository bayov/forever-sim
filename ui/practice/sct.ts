// The scrolling combat text, like the Replay view's (ui/core/components/detailed_results/replay.tsx),
// which copies the user's MSBT. What we do to the dummy scrolls up from the target's frame. What
// happens to us (buffs we gain) scrolls down to our frame. Crits pop up next to the lines and stay
// put for a moment.
//
// Every size here is in UI units (`u`), so the text scales with the HUD like the game's.

// How long a line stays up, in seconds. It fades over the last part of its life.
const SCROLL_LIFE = 2.2;
const STICKY_LIFE = 1.6;
const FADE_SHARE = 0.3;
// How far apart the scrolling lines stay. A new line pushes the lines too close ahead of it further
// along.
const LANE = 30;
// The crits' spots, as many as fit at this spacing, up to 4. The first crit takes a middle spot.
const CRIT_SPOT = 46;
const MAX_CRIT_SPOTS = 4;
// Crits stay this far toward the middle of the screen from the lines that scroll by.
const STICKY_GAP = 36;
// The room we keep from the frames, and the longest way a line scrolls.
const MARGIN = 22;
const MAX_SCROLL = 380;

export type SctSide = 'in' | 'out';
export type SctKind = 'damage' | 'crit' | 'tick' | 'avoid' | 'aura';

export type SctLine = {
	side: SctSide;
	kind: SctKind;
	text: string;
	note?: string;
	color: string;
	icon?: (el: HTMLElement) => void;
};

type Shown = SctLine & { el: HTMLElement; born: number; sticky: boolean; lane: number; fadeAt: number; end: number };

// The colors of the spell schools, by the school's bit in the sim (core.SpellSchool).
const schoolColors: [number, string][] = [
	[1 << 6, '#d1fa99'], // Nature
	[1 << 3, '#ff9b5e'], // Fire
	[1 << 4, '#7fb2ff'], // Frost
	[1 << 7, '#c6b6ff'], // Shadow
	[1 << 5, '#ffff8f'], // Holy
	[1 << 2, '#8ff2ff'], // Arcane
];
export const auraColor = '#ffd75e';
export const avoidColor = '#d9d9d9';

export function schoolColor(school: number): string {
	return schoolColors.find(([bit]) => school & bit)?.[1] ?? '#ffffff';
}

export class CombatText {
	private lines: Shown[] = [];

	constructor(readonly box: HTMLElement) {}

	add(line: SctLine, now = performance.now() / 1000) {
		// A hidden tab draws nothing, so its lines would pile up.
		if (document.hidden) return;
		const sticky = line.kind === 'crit';
		const el = document.createElement('div');
		el.className = `sct sct-${line.side} ${line.kind}${sticky ? ' sticky' : ''}`;
		el.style.color = line.color;
		if (line.icon) {
			const icon = document.createElement('span');
			icon.className = 'sct-icon';
			line.icon(icon);
			el.append(icon);
		}
		const amount = document.createElement('span');
		amount.textContent = line.text;
		el.append(amount);
		if (line.note) {
			const note = document.createElement('span');
			note.className = 'sct-note';
			note.textContent = `(${line.note})`;
			el.append(note);
		}
		el.style.opacity = '0';
		this.box.append(el);
		const life = sticky ? STICKY_LIFE : SCROLL_LIFE;
		const shown: Shown = { ...line, el, born: now, sticky, lane: 0, fadeAt: life * (1 - FADE_SHARE), end: life };
		if (sticky) this.placeCrit(shown, now);
		this.lines.push(shown);
	}

	// placeCrit gives a crit a spot no other crit on its side is on. When every spot is taken, it
	// covers the oldest crit, and that one starts to fade right away.
	private placeCrit(crit: Shown, now: number) {
		const taken = this.lines.filter(l => l.sticky && l.side === crit.side && now - l.born < l.end);
		const order = this.spotOrder(MAX_CRIT_SPOTS);
		const free = order.find(lane => !taken.some(l => l.lane === lane));
		if (free !== undefined) {
			crit.lane = free;
			return;
		}
		const oldest = taken.reduce((a, b) => (b.born < a.born ? b : a));
		crit.lane = oldest.lane;
		const age = now - oldest.born;
		if (age < oldest.fadeAt) {
			oldest.fadeAt = age;
			oldest.end = age + 0.25;
		}
	}

	// The spots in the order the crits take them, from the middle out.
	private spotOrder(spots: number): number[] {
		return Array.from({ length: spots }, (_, i) => i).sort((a, b) => Math.abs(a - (spots - 1) / 2) - Math.abs(b - (spots - 1) / 2) || a - b);
	}

	// draw places the lines for this frame. What we do scrolls from over the target's frame up to
	// `outTopAt` (under your buffs). What happens to us scrolls from `inTopAt` (under the cast
	// history) down to over our frame. Both are pixels from the top of the window.
	draw(u: number, player: DOMRect, target: DOMRect | undefined, inTopAt: number, outTopAt: number, now = performance.now() / 1000) {
		const zone = this.box.getBoundingClientRect();
		const W = zone.width;
		// Without a target, the lines scroll where its frame would be, across from ours.
		const targetLeft = target && target.width > 0 ? target.left - zone.left : W - (player.right - zone.left);
		const targetTop = target && target.height > 0 ? target.top - zone.top : player.top - zone.top;
		const ourRight = W - (player.right - zone.left);
		const outBottom = targetTop - MARGIN * u;
		const outTop = Math.max(outTopAt - zone.top + MARGIN * u, outBottom - MAX_SCROLL * u);
		const inBottom = player.top - zone.top - MARGIN * u;
		const inTop = Math.max(inTopAt - zone.top + MARGIN * u, inBottom - MAX_SCROLL * u);
		const spots = Math.max(1, Math.min(MAX_CRIT_SPOTS, Math.floor((outBottom - outTop) / (CRIT_SPOT * u))));
		const rows = this.spotOrder(spots);

		this.lines = this.lines.filter(l => {
			if (now - l.born <= l.end) return true;
			l.el.remove();
			return false;
		});
		// We place the lines from the newest back, so an older line too close to a newer one gets
		// pushed ahead of it.
		let newerIn = -Infinity;
		let newerOut = Infinity;
		for (let i = this.lines.length - 1; i >= 0; i--) {
			const l = this.lines[i];
			const age = now - l.born;
			const progress = age / l.end;
			const el = l.el;
			let y: number;
			let scale = 1;
			let pushedOut = false;
			if (l.sticky) {
				const bottom = l.side === 'out' ? outBottom : inBottom;
				const span = bottom - (l.side === 'out' ? outTop : inTop);
				const row = rows.indexOf(l.lane % spots);
				y = bottom - span + ((row < 0 ? 0 : row) + 0.5) * (span / spots);
				scale = 1 + 0.55 * Math.max(0, 1 - age / 0.12);
			} else if (l.side === 'out') {
				y = newerOut = Math.min(outBottom - progress * (outBottom - outTop), newerOut - LANE * u);
				pushedOut = y < outTop;
			} else {
				y = newerIn = Math.max(inTop + progress * (inBottom - inTop), newerIn + LANE * u);
				pushedOut = y > inBottom;
			}
			const opacity = pushedOut ? 0 : Math.min(1, age / 0.06) * Math.min(1, (l.end - age) / (l.end - l.fadeAt));
			el.style.top = `${y.toFixed(1)}px`;
			if (l.side === 'out' && !l.sticky) el.style.left = `${Math.max(0, targetLeft).toFixed(1)}px`;
			else if (l.side === 'out') el.style.right = `${(W - targetLeft + STICKY_GAP * u).toFixed(1)}px`;
			else if (!l.sticky) el.style.right = `${Math.max(0, ourRight).toFixed(1)}px`;
			else el.style.left = `${(W - ourRight + STICKY_GAP * u).toFixed(1)}px`;
			el.style.transform = `translateY(-50%) scale(${scale.toFixed(3)})`;
			el.style.opacity = opacity.toFixed(3);
		}
	}

	clear() {
		for (const l of this.lines) l.el.remove();
		this.lines = [];
	}
}
