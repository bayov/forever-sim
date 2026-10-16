import { ResourceType } from '../../proto/api.js';
import { ActionId } from '../../proto_utils/action_id.js';
import { UnitMetrics } from '../../proto_utils/sim_result.js';
import { EventFilter } from './events_filter.js';
import {
	formatClock,
	LevelingSession,
	mobName,
	MobLife,
	mobPositionAt,
	positionAt,
	SECTION_LABELS,
	SectionKind,
	SESSION_STATES,
	SessionSection,
	SessionState,
	sectionTitle,
} from './leveling_session.js';
import { TotemSlot } from './replay_model.js';

// The map of a leveling session in the middle of the replay: the plane the sim moves us and the
// mobs on (sim/core/leveling_world.go), seen from above, with north up.
//
// Like the rest of the replay, what it shows comes from the time alone, so a seek draws the same
// frame as playing up to it. That holds for the camera too: it follows the average of where it
// would look over a couple of seconds around now, not where it looked on the frame before.

// How long the trails behind us and the running mobs are, in seconds.
const TRAIL_SECONDS = 6;
const TRAIL_STEP = 0.2;
// How long the mark of a kill or of our death stays on the map, in seconds.
const MARK_SECONDS = 8;
// The camera never shows less than this many yards across, so a fight on the spot doesn't fill
// the map with one mob.
const MIN_SPAN_YARDS = 36;
const MAX_SPAN_YARDS = 160;
const PAD_YARDS = 7;
// The camera looks at what matters over this window of the session, so it eases in and out. It
// looks a little ahead too, so a mob we find doesn't show up at the edge.
const CAMERA_WINDOW = 2;
const CAMERA_AHEAD = 0.6;
const CAMERA_SAMPLES = 9;
// Mobs on us stand around us on a ring this many pixels out, so each one shows.
const RING_PX = 30;
// Totems further than this from us don't pull the camera out.
const TOTEM_FOCUS_YARDS = 40;

export interface ReplayMapConfig {
	// The replay's hover tooltip, for the things on the map.
	tip: (elem: Element, content: () => Element | null) => void;
}

export interface ReplayMapData {
	session: LevelingSession;
	playerLevel: number;
	playerName: string;
	classColor: string;
	totems: Array<TotemSlot>;
	// A mob's share of health left, by its target index.
	mobHealth: (targetIndex: number, time: number) => number;
}

interface PlacedTotem {
	actionId: ActionId;
	element: string;
	start: number;
	until: number;
	x: number;
	y: number;
}

interface Camera {
	x: number;
	y: number;
	// Pixels a yard.
	scale: number;
}

interface MobView {
	root: HTMLElement;
	health: HTMLElement;
	label: HTMLElement;
	shown: boolean;
}

const STATE_ICONS: Partial<Record<SessionState, string>> = {
	searching: 'fa-magnifying-glass',
	resting: 'fa-mug-hot',
	bandaging: 'fa-kit-medical',
	dead: 'fa-ghost',
};

// The aggro radius of a mob we haven't pulled (leveling_world.go aggroRadius).
export const aggroRadius = (playerLevel: number, mobLevel: number): number => Math.max(20 - Math.max(playerLevel - mobLevel, -25), 5);

export class ReplayMap {
	readonly root: HTMLElement;
	private readonly canvas: HTMLCanvasElement;
	private readonly layer: HTMLElement;
	private readonly banner: HTMLElement;
	private readonly bannerDot: HTMLElement;
	private readonly bannerText: HTMLElement;
	private readonly bannerSection: HTMLElement;
	private readonly scaleBar: HTMLElement;
	private readonly scaleLabel: HTMLElement;
	private readonly playerElem: HTMLElement;
	private readonly playerArrow: HTMLElement;
	// An icon over us for what we're doing out of combat, like a cup while we eat and drink.
	private readonly playerState: HTMLElement;
	private playerStateShown = '';
	private readonly tip: ReplayMapConfig['tip'];

	private data: ReplayMapData | null = null;
	private totems: Array<PlacedTotem> = [];
	private readonly mobViews = new Map<MobLife, MobView>();
	private readonly totemViews = new Map<PlacedTotem, HTMLElement>();
	private readonly markViews = new Map<string, HTMLElement>();
	private width = 0;
	private height = 0;
	private time = 0;

	constructor(config: ReplayMapConfig) {
		this.tip = config.tip;
		this.root = document.createElement('div');
		this.root.className = 'replay-map hide';
		this.canvas = document.createElement('canvas');
		this.canvas.className = 'replay-map-canvas';
		this.layer = document.createElement('div');
		this.layer.className = 'replay-map-layer';

		this.banner = div('replay-map-banner');
		this.bannerDot = div('replay-map-banner-dot');
		this.bannerText = div('replay-map-banner-text');
		this.bannerSection = div('replay-map-banner-section');
		this.banner.append(this.bannerDot, this.bannerText, this.bannerSection);

		this.scaleBar = div('replay-map-scale');
		this.scaleLabel = div('replay-map-scale-label');
		this.scaleBar.appendChild(this.scaleLabel);

		const compass = div('replay-map-compass');
		compass.textContent = 'N';

		this.playerElem = div('replay-map-unit player');
		this.playerArrow = div('replay-map-arrow');
		this.playerState = icon('replay-map-state');
		this.playerElem.append(this.playerArrow, div('replay-map-dot'), div('replay-map-label'), this.playerState);
		this.layer.appendChild(this.playerElem);

		this.root.append(this.canvas, this.layer, this.banner, this.scaleBar, compass);

		new ResizeObserver(() => this.measure()).observe(this.root);
	}

	setData(data: ReplayMapData | null) {
		this.data = data;
		this.root.classList.toggle('hide', !data);
		this.mobViews.forEach(view => view.root.remove());
		this.mobViews.clear();
		this.totemViews.forEach(elem => elem.remove());
		this.totemViews.clear();
		this.markViews.forEach(elem => elem.remove());
		this.markViews.clear();
		if (!data) return;

		// A totem stands where we were when we put it down, until it runs out or the next totem of
		// its element takes its place.
		this.totems = data.totems.flatMap(slot =>
			slot.drops.map((drop, i) => {
				const next = slot.drops[i + 1];
				const at = positionAt(data.session.player, drop.start) ?? { x: 0, y: 0 };
				return {
					actionId: drop.actionId,
					element: slot.element,
					start: drop.start,
					until: Math.min(drop.until, next ? next.start : Infinity),
					x: at.x,
					y: at.y,
				};
			}),
		);

		this.playerElem.style.setProperty('--unit-color', data.classColor);
		this.playerElem.querySelector('.replay-map-label')!.textContent = data.playerName;
		this.tip(this.playerElem, () => this.playerTip());
		this.measure();
	}

	private measure() {
		const width = this.root.clientWidth;
		const height = this.root.clientHeight;
		if (width == this.width && height == this.height) return;
		this.width = width;
		this.height = height;
		const ratio = window.devicePixelRatio || 1;
		this.canvas.width = Math.round(width * ratio);
		this.canvas.height = Math.round(height * ratio);
		this.canvas.style.width = `${width}px`;
		this.canvas.style.height = `${height}px`;
		if (this.data) this.draw(this.time);
	}

	draw(time: number) {
		this.time = time;
		const data = this.data;
		if (!data || this.width == 0) return;
		const session = data.session;
		const camera = this.cameraAt(time);
		const toScreen = (x: number, y: number) => ({
			x: this.width / 2 + (x - camera.x) * camera.scale,
			y: this.height / 2 - (y - camera.y) * camera.scale,
		});

		const state = session.stateAt(time);
		this.root.dataset.state = state;
		this.drawCanvas(data, camera, toScreen, time, state);
		this.drawBanner(data, time, state);

		// Us.
		const me = positionAt(session.player, time) ?? { x: 0, y: 0, moving: false };
		const meScreen = toScreen(me.x, me.y);
		place(this.playerElem, meScreen.x, meScreen.y);
		this.playerElem.classList.toggle('ghost', state == 'dead');
		const stateIcon = STATE_ICONS[state] ?? '';
		if (stateIcon != this.playerStateShown) {
			this.playerState.className = `fas replay-map-state ${stateIcon}`;
			this.playerState.classList.toggle('hide', !stateIcon);
			this.playerStateShown = stateIcon;
		}
		this.playerElem.classList.toggle('moving', me.moving);
		const heading = this.headingAt(data, time);
		if (heading != null) this.playerArrow.style.transform = `rotate(${(-heading * 180) / Math.PI}deg)`;
		this.playerArrow.classList.toggle('hide', heading == null);

		// The mobs out of the pool. The ones on us stand on a ring around us, in the order they came.
		const shownMobs = new Set<MobLife>();
		const onUs = this.mobsOnUs(session, time);
		for (const mob of this.activeMobs(session, time)) {
			const view = this.mobView(mob);
			shownMobs.add(mob);
			const ringIndex = onUs.indexOf(mob);
			let screen: { x: number; y: number };
			if (ringIndex >= 0) {
				const angle = -Math.PI / 2 + (ringIndex * 2 * Math.PI) / Math.max(onUs.length, 1) + Math.PI / Math.max(onUs.length, 1);
				screen = { x: meScreen.x + Math.cos(angle) * RING_PX, y: meScreen.y + Math.sin(angle) * RING_PX };
			} else {
				const at = mobPositionAt(session, mob, time);
				if (!at) continue;
				screen = toScreen(at.x, at.y);
			}
			place(view.root, screen.x, screen.y);
			const pulled = mob.pulledAt != null && time >= mob.pulledAt;
			view.root.classList.toggle('found', !pulled);
			view.root.classList.toggle('running', pulled && (mob.reachedAt == null || time < mob.reachedAt));
			view.root.classList.toggle('on-us', ringIndex >= 0);
			const health = pulled ? data.mobHealth(mob.targetIndex, time) : 1;
			view.health.style.width = `${(health * 100).toFixed(1)}%`;
			view.root.classList.toggle('damaged', health < 1);
			if (!view.shown) {
				view.root.classList.remove('hide');
				view.shown = true;
			}
		}
		this.mobViews.forEach((view, mob) => {
			if (!shownMobs.has(mob) && view.shown) {
				view.root.classList.add('hide');
				view.shown = false;
			}
		});

		// Our totems, while they're up.
		for (const totem of this.totems) {
			const up = time >= totem.start && time < totem.until;
			let elem = this.totemViews.get(totem);
			if (!up) {
				elem?.classList.add('hide');
				continue;
			}
			if (!elem) {
				elem = this.totemView(totem);
				this.totemViews.set(totem, elem);
			}
			elem.classList.remove('hide');
			const screen = toScreen(totem.x, totem.y);
			place(elem, screen.x, screen.y);
			const left = totem.until - time;
			const text = elem.querySelector('.replay-map-totem-time')!;
			const shown = Number.isFinite(left) && left < 600 ? (left < 10 ? left.toFixed(0) : formatClock(left)) : '';
			if (text.textContent != shown) text.textContent = shown;
		}

		this.drawMarks(data, time, toScreen, camera);
		this.drawScale(camera);
	}

	// The camera at this time: the middle and the size of the box around what matters, averaged
	// over CAMERA_WINDOW seconds that end CAMERA_AHEAD seconds from now.
	private cameraAt(time: number): Camera {
		let x = 0;
		let y = 0;
		let span = 0;
		let weights = 0;
		for (let i = 0; i < CAMERA_SAMPLES; i++) {
			const t = time + CAMERA_AHEAD - (CAMERA_WINDOW * i) / (CAMERA_SAMPLES - 1);
			const box = this.focusBox(Math.max(0, t));
			// The recent samples count more, so the camera keeps up with us.
			const weight = CAMERA_SAMPLES - i;
			x += box.x * weight;
			y += box.y * weight;
			span += box.span * weight;
			weights += weight;
		}
		span /= weights;
		const size = Math.max(1, Math.min(this.width, this.height));
		return { x: x / weights, y: y / weights, scale: size / span };
	}

	// The box around us, the mob we found, the mobs running at us and our totems close by.
	private focusBox(time: number): { x: number; y: number; span: number } {
		const session = this.data!.session;
		const me = positionAt(session.player, time) ?? { x: 0, y: 0 };
		let minX = me.x;
		let maxX = me.x;
		let minY = me.y;
		let maxY = me.y;
		const add = (x: number, y: number) => {
			minX = Math.min(minX, x);
			maxX = Math.max(maxX, x);
			minY = Math.min(minY, y);
			maxY = Math.max(maxY, y);
		};
		for (const mob of this.activeMobs(session, time)) {
			if (mob.reachedAt != null && time >= mob.reachedAt) continue;
			const at = mobPositionAt(session, mob, time);
			if (at) add(at.x, at.y);
		}
		for (const totem of this.totems) {
			if (time >= totem.start && time < totem.until && Math.hypot(totem.x - me.x, totem.y - me.y) < TOTEM_FOCUS_YARDS) add(totem.x, totem.y);
		}
		// The map is wider than it's tall, so the box keeps the map's shape.
		const aspect = this.height > 0 ? this.width / this.height : 1;
		const spanX = (maxX - minX + 2 * PAD_YARDS) / Math.max(aspect, 1);
		const spanY = (maxY - minY + 2 * PAD_YARDS) * Math.min(aspect, 1);
		const span = Math.min(MAX_SPAN_YARDS, Math.max(MIN_SPAN_YARDS, spanX, spanY));
		return { x: (minX + maxX) / 2, y: (minY + maxY) / 2, span };
	}

	// The mob lives out of the pool at this time: found and waiting, running at us or on us.
	private activeMobs(session: LevelingSession, time: number): Array<MobLife> {
		const mobs: Array<MobLife> = [];
		for (const mob of session.mobs) {
			if (mob.foundAt > time) break;
			if (time <= mob.end && mob.diedAt !== time && mob.leftAt !== time) mobs.push(mob);
		}
		return mobs;
	}

	private mobsOnUs(session: LevelingSession, time: number): Array<MobLife> {
		return this.activeMobs(session, time)
			.filter(mob => mob.reachedAt != null && time >= mob.reachedAt)
			.sort((a, b) => a.reachedAt! - b.reachedAt!);
	}

	// Which way we face: the way we move, or toward the mob we fight or found.
	private headingAt(data: ReplayMapData, time: number): number | null {
		const session = data.session;
		const i = lastIndex(session.player, time);
		const motion = i >= 0 ? session.player[i] : null;
		if (motion && (motion.vx != 0 || motion.vy != 0)) return Math.atan2(motion.vy, motion.vx);
		const me = positionAt(session.player, time);
		if (!me) return null;
		const mobs = this.activeMobs(session, time).filter(mob => mob.reachedAt == null || time < mob.reachedAt);
		const mob = mobs[0];
		const at = mob && mobPositionAt(session, mob, time);
		if (at && (at.x != me.x || at.y != me.y)) return Math.atan2(at.y - me.y, at.x - me.x);
		// Without a better guess, we keep facing the way we last moved.
		for (let j = i; j >= 0; j--) {
			const m = session.player[j];
			if (m.vx != 0 || m.vy != 0) return Math.atan2(m.vy, m.vx);
		}
		return null;
	}

	private drawCanvas(data: ReplayMapData, camera: Camera, toScreen: (x: number, y: number) => { x: number; y: number }, time: number, state: SessionState) {
		const ctx = this.canvas.getContext('2d')!;
		const ratio = this.canvas.width / Math.max(this.width, 1);
		ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
		ctx.clearRect(0, 0, this.width, this.height);

		// The grid, a line every 5 yards and a stronger one every 25. When we zoom out far, the
		// thin lines would be too close, so we only draw the strong ones.
		const left = camera.x - this.width / 2 / camera.scale;
		const right = camera.x + this.width / 2 / camera.scale;
		const bottom = camera.y - this.height / 2 / camera.scale;
		const top = camera.y + this.height / 2 / camera.scale;
		const step = camera.scale * 5 < 9 ? 25 : 5;
		ctx.lineWidth = 1;
		for (let gx = Math.floor(left / step) * step; gx <= right; gx += step) {
			const sx = Math.round(toScreen(gx, 0).x) + 0.5;
			ctx.strokeStyle = gx % 25 == 0 ? 'rgba(255, 255, 255, 0.13)' : 'rgba(255, 255, 255, 0.05)';
			ctx.beginPath();
			ctx.moveTo(sx, 0);
			ctx.lineTo(sx, this.height);
			ctx.stroke();
		}
		for (let gy = Math.floor(bottom / step) * step; gy <= top; gy += step) {
			const sy = Math.round(toScreen(0, gy).y) + 0.5;
			ctx.strokeStyle = gy % 25 == 0 ? 'rgba(255, 255, 255, 0.13)' : 'rgba(255, 255, 255, 0.05)';
			ctx.beginPath();
			ctx.moveTo(0, sy);
			ctx.lineTo(this.width, sy);
			ctx.stroke();
		}

		const session = data.session;
		const me = positionAt(session.player, time) ?? { x: 0, y: 0, moving: false };
		const meScreen = toScreen(me.x, me.y);

		// The aggro radius of the mob we found, and a dashed line to it.
		for (const mob of this.activeMobs(session, time)) {
			const pulled = mob.pulledAt != null && time >= mob.pulledAt;
			const at = mobPositionAt(session, mob, time);
			if (!at) continue;
			const screen = toScreen(at.x, at.y);
			if (!pulled) {
				const radius = aggroRadius(data.playerLevel, mob.level) * camera.scale;
				ctx.fillStyle = 'rgba(255, 90, 70, 0.045)';
				ctx.strokeStyle = 'rgba(255, 110, 90, 0.5)';
				ctx.setLineDash([6, 4]);
				ctx.beginPath();
				ctx.arc(screen.x, screen.y, radius, 0, 2 * Math.PI);
				ctx.fill();
				ctx.stroke();
				ctx.strokeStyle = 'rgba(255, 220, 120, 0.45)';
				ctx.setLineDash([4, 5]);
				ctx.beginPath();
				ctx.moveTo(meScreen.x, meScreen.y);
				ctx.lineTo(screen.x, screen.y);
				ctx.stroke();
				ctx.setLineDash([]);
			} else if (mob.reachedAt == null || time < mob.reachedAt) {
				// A mob running at us leaves a red trail.
				this.trail(ctx, toScreen, t => (t >= (mob.pulledAt ?? 0) ? mobPositionAt(session, mob, t) : null), time, 'rgba(255, 90, 70, ');
				ctx.strokeStyle = 'rgba(255, 90, 70, 0.35)';
				ctx.setLineDash([2, 4]);
				ctx.beginPath();
				ctx.moveTo(screen.x, screen.y);
				ctx.lineTo(meScreen.x, meScreen.y);
				ctx.stroke();
				ctx.setLineDash([]);
			}
		}

		// Our trail, in our class color.
		this.trail(ctx, toScreen, t => positionAt(session.player, t), time, 'rgba(255, 230, 170, ');

		// While we're dead the world goes gray, like the game's ghost view.
		if (state == 'dead') {
			ctx.fillStyle = 'rgba(20, 30, 40, 0.35)';
			ctx.fillRect(0, 0, this.width, this.height);
		}
	}

	private trail(
		ctx: CanvasRenderingContext2D,
		toScreen: (x: number, y: number) => { x: number; y: number },
		at: (t: number) => { x: number; y: number } | null,
		time: number,
		rgbaPrefix: string,
	) {
		let last: { x: number; y: number } | null = null;
		ctx.lineWidth = 2.5;
		ctx.lineCap = 'round';
		for (let back = TRAIL_SECONDS; back >= 0; back -= TRAIL_STEP) {
			const t = time - back;
			if (t < 0) continue;
			const p = at(t);
			if (!p) {
				last = null;
				continue;
			}
			const screen = toScreen(p.x, p.y);
			if (last && (Math.abs(last.x - screen.x) > 0.3 || Math.abs(last.y - screen.y) > 0.3)) {
				ctx.strokeStyle = `${rgbaPrefix}${(0.55 * (1 - back / TRAIL_SECONDS)).toFixed(3)})`;
				ctx.beginPath();
				ctx.moveTo(last.x, last.y);
				ctx.lineTo(screen.x, screen.y);
				ctx.stroke();
			}
			last = screen;
		}
		ctx.lineWidth = 1;
	}

	// The marks of the kills and of our death, where they happened, fading over MARK_SECONDS.
	private drawMarks(data: ReplayMapData, time: number, toScreen: (x: number, y: number) => { x: number; y: number }, camera: Camera) {
		const session = data.session;
		const shown = new Set<string>();
		const mark = (key: string, kind: 'kill' | 'death', t: number, x: number, y: number, tip: () => Element) => {
			const age = time - t;
			if (age < 0 || age > MARK_SECONDS) return;
			let elem = this.markViews.get(key);
			if (!elem) {
				elem = div(`replay-map-mark ${kind}`);
				elem.appendChild(icon(kind == 'kill' ? 'fa-xmark' : 'fa-skull'));
				this.tip(elem, tip);
				this.layer.appendChild(elem);
				this.markViews.set(key, elem);
			}
			const screen = toScreen(x, y);
			place(elem, screen.x, screen.y);
			elem.style.opacity = (1 - age / MARK_SECONDS).toFixed(3);
			elem.classList.remove('hide');
			shown.add(key);
		};
		for (const kill of session.kills) {
			if (kill.t > time) break;
			if (time - kill.t > MARK_SECONDS) continue;
			const at = mobPositionAt(session, kill.mob, kill.t) ?? positionAt(session.player, kill.t);
			if (!at) continue;
			// A mob that died on us lay on the ring around us. We put its mark a little off our spot.
			const offset = kill.mob.reachedAt != null && kill.t >= kill.mob.reachedAt ? RING_PX / camera.scale : 0;
			mark(`kill ${kill.t} ${kill.mob.label}`, 'kill', kill.t, at.x, at.y + offset, () =>
				tipBox(`${mobName(kill.mob)} (level ${kill.mob.level})`, [`Killed at ${formatClock(kill.t, true)}`]),
			);
		}
		for (const death of session.deaths) {
			const at = positionAt(session.player, death.t);
			// Our corpse stays until we're back.
			if (!at || death.t > time || time > death.backAt + MARK_SECONDS) continue;
			mark(`death ${death.t}`, 'death', Math.max(death.t, Math.min(time, death.backAt)), at.x, at.y, () =>
				tipBox('We died', [`At ${formatClock(death.t, true)}`, `Back at ${formatClock(death.backAt, true)}`]),
			);
		}
		this.markViews.forEach((elem, key) => {
			if (!shown.has(key)) elem.classList.add('hide');
		});
	}

	// The scale bar: the roundest length that takes about a sixth of the map.
	private drawScale(camera: Camera) {
		const target = this.width / 6 / camera.scale;
		const yards = [5, 10, 20, 25, 50, 100].find(y => y >= target) ?? 100;
		this.scaleBar.style.width = `${Math.round(yards * camera.scale)}px`;
		const text = `${yards} yd`;
		if (this.scaleLabel.textContent != text) this.scaleLabel.textContent = text;
	}

	// What we're doing, at the top of the map, like "Searching 6.2 s" or "Pull 7: 3 mobs".
	private drawBanner(data: ReplayMapData, time: number, state: SessionState) {
		const session = data.session;
		const span =
			session.spans[
				Math.max(
					0,
					lastIndexBy(session.spans, time, s => s.start),
				)
			];
		const inState = time - span.start;
		const section = session.sectionAt(time);
		let text = SESSION_STATES[state].label;
		switch (state) {
			case 'searching':
				text = `Searching ${inState.toFixed(1)} s`;
				break;
			case 'pulling': {
				const mob = this.activeMobs(session, time).find(m => m.pulledAt == null || time < m.pulledAt);
				text = mob ? `Pulling ${mobName(mob)} (level ${mob.level})` : 'Pulling';
				break;
			}
			case 'fighting': {
				const pull = section?.pull;
				const pulledSoFar = pull ? pull.mobs.filter(mob => mob.pulledAt! <= time).length : 0;
				text = pull ? `${sectionTitle(section!)}: ${pulledSoFar} mob${pulledSoFar == 1 ? '' : 's'} · ${formatClock(inState)}` : 'Fighting';
				break;
			}
			case 'resting': {
				const what = section?.ate && section?.drank ? 'Eating and drinking' : section?.drank ? 'Drinking' : section?.ate ? 'Eating' : 'Resting';
				const share = Math.min(1, inState / Math.max(span.end - span.start, 0.001));
				text = `${what} ${Math.round(share * 100)}%`;
				break;
			}
			case 'bandaging':
				text = `Bandaging ${inState.toFixed(1)} s`;
				break;
			case 'dead':
				text = `Corpse run ${formatClock(Math.max(0, span.end - time))}`;
				break;
			case 'ready':
				text = 'Ready';
				break;
		}
		if (this.bannerText.textContent != text) this.bannerText.textContent = text;
		this.bannerDot.style.background = SESSION_STATES[state].color;
		const sectionText = section && state != 'fighting' ? sectionTitle(section) : '';
		if (this.bannerSection.textContent != sectionText) this.bannerSection.textContent = sectionText;
	}

	private mobView(mob: MobLife): MobView {
		const existing = this.mobViews.get(mob);
		if (existing) return existing;
		const root = div('replay-map-unit mob hide');
		const health = div('replay-map-health-fill');
		const bar = div('replay-map-health');
		bar.appendChild(health);
		const label = div('replay-map-label');
		label.textContent = `${mobName(mob)} · ${mob.level}`;
		root.append(div('replay-map-dot'), bar, label);
		this.layer.appendChild(root);
		this.tip(root, () => this.mobTip(mob));
		const view = { root, health, label, shown: false };
		this.mobViews.set(mob, view);
		return view;
	}

	private totemView(totem: PlacedTotem): HTMLElement {
		const elem = div(`replay-map-totem ${totem.element.toLowerCase()}`);
		const img = div('replay-map-totem-icon');
		totem.actionId.setBackground(img as unknown as HTMLAnchorElement);
		const time = div('replay-map-totem-time');
		elem.append(img, time);
		this.layer.appendChild(elem);
		this.tip(elem, () =>
			tipBox(totem.actionId.name, [
				`${totem.element} totem, down at ${formatClock(totem.start, true)}`,
				Number.isFinite(totem.until) ? `Gone at ${formatClock(totem.until, true)}` : '',
			]),
		);
		return elem;
	}

	private mobTip(mob: MobLife): Element {
		const time = this.time;
		const lines: Array<string> = [`Found ${mob.foundDistance} yd away at ${formatClock(mob.foundAt, true)}${mob.foundInCombat ? ', while fighting' : ''}`];
		if (mob.pulledAt != null && time >= mob.pulledAt) {
			lines.push(
				`Pulled ${mob.pulledBy == 'noticed' ? 'when it noticed us' : 'by a spell'} from ${mob.pulledFrom} yd at ${formatClock(mob.pulledAt, true)}`,
			);
			lines.push(`Health ${Math.round(this.data!.mobHealth(mob.targetIndex, time) * 100)}%`);
		} else {
			lines.push(`Aggro radius ${aggroRadius(this.data!.playerLevel, mob.level)} yd`);
		}
		if (mob.pull) lines.push(`Part of Pull ${mob.pull.number}`);
		return tipBox(`${mobName(mob)} (level ${mob.level})`, lines);
	}

	private playerTip(): Element {
		const data = this.data!;
		const time = this.time;
		const at = positionAt(data.session.player, time);
		const state = data.session.stateAt(time);
		return tipBox(data.playerName, [SESSION_STATES[state].label, at ? `At (${at.x.toFixed(0)}, ${at.y.toFixed(0)}) yd` : '']);
	}
}

// A pool mob's share of health left over the session. It comes back at full health each time we
// find it again, and the log has no line for that, so we start each of its lives at full.
export function levelingMobHealth(session: LevelingSession, target: UnitMetrics): (time: number) => number {
	const events: Array<{ t: number; value: number | null; order: number }> = [];
	let max = 0;
	target.logs.forEach(log => {
		if (!log.isResourceChanged() || log.resourceType != ResourceType.ResourceTypeHealth) return;
		max = Math.max(max, log.valueBefore, log.valueAfter);
		events.push({ t: log.timestamp, value: log.valueAfter, order: 1 });
	});
	session.mobs.forEach(mob => {
		if (mob.targetIndex == target.index) events.push({ t: mob.foundAt, value: null, order: 0 });
	});
	events.sort((a, b) => a.t - b.t || a.order - b.order);
	const times = events.map(e => e.t);
	const values = events.map(e => (e.value == null || max <= 0 ? 1 : Math.max(0, Math.min(1, e.value / max))));
	return (time: number) => {
		const i = lastIndexBy(times, time, t => t);
		return i < 0 ? 1 : values[i];
	};
}

// What the replay lends the leveling transport: its clock, a seek, and the stretch it plays.
export interface LevelingTransportHost {
	time(): number;
	seek(time: number): void;
	tip: ReplayMapConfig['tip'];
}

const NAV_KINDS: Array<SectionKind> = ['pull', 'rest', 'search', 'death'];

// The leveling session's own controls under the replay: previous and next section (of the kind
// we pick, pulls by default), skipping ahead to the next pull while we're not fighting, and the
// session drawn along the progress bar.
//
// With a stretch picked on the session strip (the Events tab's From and To), the replay plays
// only that stretch. Then the previous and next buttons pick the section before or after it on
// the strip, so the other views follow. Without one they only move the replay.
export class LevelingTransport {
	readonly root: HTMLElement;
	private readonly kindSelect: HTMLSelectElement;
	private readonly prevButton: HTMLButtonElement;
	private readonly nextButton: HTMLButtonElement;
	private readonly skipButton: HTMLButtonElement;
	private session: LevelingSession | null = null;
	private kind: SectionKind = 'pull';

	constructor(
		private readonly host: LevelingTransportHost,
		private readonly filter: EventFilter,
	) {
		this.root = div('replay-section-buttons hide');
		this.prevButton = button('replay-button small', 'fa-backward-step');
		this.nextButton = button('replay-button small', 'fa-forward-step');
		this.kindSelect = document.createElement('select');
		this.kindSelect.className = 'replay-section-kind';
		NAV_KINDS.forEach(kind => {
			const option = document.createElement('option');
			option.value = kind;
			option.textContent = SECTION_LABELS[kind].many;
			this.kindSelect.appendChild(option);
		});
		this.kindSelect.addEventListener('change', () => (this.kind = this.kindSelect.value as SectionKind));
		this.skipButton = document.createElement('button');
		this.skipButton.type = 'button';
		this.skipButton.className = 'replay-skip-rest';
		this.skipButton.textContent = 'Skip to pull';
		this.root.append(this.prevButton, this.kindSelect, this.nextButton, this.skipButton);

		// The controls sit on the stage's controls, and a click there would play or pause.
		this.root.addEventListener('click', event => event.stopPropagation());
		this.prevButton.addEventListener('click', () => this.step(-1));
		this.nextButton.addEventListener('click', () => this.step(1));
		this.skipButton.addEventListener('click', () => this.skip());
		host.tip(this.prevButton, () => tipBox(`Previous ${SECTION_LABELS[this.kind].one.toLowerCase()} ([)`, []));
		host.tip(this.nextButton, () => tipBox(`Next ${SECTION_LABELS[this.kind].one.toLowerCase()} (])`, []));
		host.tip(this.skipButton, () =>
			tipBox('Skip to the next pull (S)', ['Jumps past the search, the rest or the corpse run we are in, to when we pull the next mob.']),
		);
	}

	setSession(session: LevelingSession | null) {
		this.session = session;
		this.root.classList.toggle('hide', !session);
	}

	// Moves to the section before or after the one we're in, of the kind we picked.
	step(direction: number) {
		const session = this.session;
		if (!session) return;
		const time = this.host.time();
		const sections = session.sections.filter(section => section.kind == this.kind);
		const range = this.pickedRange();
		let target: SessionSection | undefined;
		if (direction > 0) {
			// The one after the picked stretch, or after now.
			const after = range ? range.start + 0.001 : time + 0.001;
			target = sections.find(section => section.start >= after);
		} else {
			// Back to the start of the one we're in when we're well into it, like a video's chapters.
			const before = range ? range.start - 0.001 : time - 1;
			target = [...sections].reverse().find(section => section.start <= before);
		}
		if (target) this.goTo(target);
	}

	// Jumps to when the next pull starts.
	skip() {
		const session = this.session;
		if (!session) return;
		const time = this.host.time();
		const next = session.sections.find(section => section.kind == 'pull' && section.start > time + 0.001);
		if (!next) return;
		const range = this.pickedRange();
		if (range && next.start > range.end) {
			this.goTo(next);
		} else {
			this.host.seek(next.start);
		}
	}

	private goTo(section: SessionSection) {
		if (this.pickedRange()) {
			// The Events tab tells the replay about the new stretch, and it seeks to its start.
			this.filter.setRange(section.start, section.end);
		} else {
			this.host.seek(section.start);
		}
	}

	private pickedRange(): { start: number; end: number } | null {
		if (this.filter.from == null && this.filter.to == null) return null;
		return { start: this.filter.from ?? 0, end: this.filter.to ?? Infinity };
	}

	update(time: number) {
		const session = this.session;
		if (!session) return;
		const fighting = session.stateAt(time) == 'fighting';
		const hasNext = session.sections.some(section => section.kind == 'pull' && section.start > time + 0.001);
		this.skipButton.disabled = fighting || !hasNext;
	}

	// The session along the progress bar from start to end: the states as colors, a tick for each
	// kill and a red one for each death.
	drawProgress(track: HTMLElement, start: number, end: number) {
		track.querySelectorAll('.replay-progress-states, .replay-progress-kill, .replay-progress-death').forEach(elem => elem.remove());
		const session = this.session;
		if (!session) return;
		const span = Math.max(end - start, 0.001);
		const percent = (t: number) => `${(((Math.max(start, Math.min(end, t)) - start) / span) * 100).toFixed(3)}%`;
		const states = div('replay-progress-states');
		for (const s of session.spans) {
			if (s.end <= start || s.start >= end) continue;
			const elem = div('');
			elem.style.left = percent(s.start);
			elem.style.right = `calc(100% - ${percent(s.end)})`;
			elem.style.background = SESSION_STATES[s.state].color;
			states.appendChild(elem);
		}
		const ticks: Array<HTMLElement> = [];
		for (const kill of session.kills) {
			if (kill.t < start || kill.t > end) continue;
			const elem = div('replay-progress-kill');
			elem.style.left = percent(kill.t);
			ticks.push(elem);
		}
		for (const death of session.deaths) {
			if (death.t < start || death.t > end) continue;
			const elem = div('replay-progress-death');
			elem.style.left = percent(death.t);
			ticks.push(elem);
		}
		track.prepend(states);
		track.append(...ticks);
	}
}

const button = (className: string, iconName: string): HTMLButtonElement => {
	const elem = document.createElement('button');
	elem.type = 'button';
	elem.className = className;
	elem.appendChild(icon(iconName));
	return elem;
};

const div = (className: string): HTMLElement => {
	const elem = document.createElement('div');
	elem.className = className;
	return elem;
};

const icon = (name: string): HTMLElement => {
	const elem = document.createElement('i');
	elem.className = `fas ${name}`;
	return elem;
};

const place = (elem: HTMLElement, x: number, y: number) => {
	elem.style.transform = `translate(${x.toFixed(1)}px, ${y.toFixed(1)}px)`;
};

const tipBox = (title: string, lines: Array<string>): Element => {
	const root = div('timeline-tooltip');
	const head = div('timeline-tooltip-title');
	head.textContent = title;
	root.appendChild(head);
	const rows = div('timeline-tooltip-meta');
	lines
		.filter(line => line)
		.forEach(line => {
			const row = document.createElement('div');
			row.textContent = line;
			rows.appendChild(row);
		});
	root.appendChild(rows);
	return root;
};

const lastIndexBy = <T>(items: Array<T>, t: number, key: (item: T) => number): number => {
	let lo = 0;
	let hi = items.length - 1;
	let found = -1;
	while (lo <= hi) {
		const mid = (lo + hi) >> 1;
		if (key(items[mid]) <= t) {
			found = mid;
			lo = mid + 1;
		} else {
			hi = mid - 1;
		}
	}
	return found;
};

const lastIndex = (track: Array<{ t: number }>, t: number): number => lastIndexBy(track, t, m => m.t);
