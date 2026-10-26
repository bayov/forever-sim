// @ts-expect-error
import debounce from 'lodash/debounce';
import tippy, { Instance } from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { WOWHEAD_DOMAIN, WOWHEAD_EXPANSION_ENV } from '../../constants/wowhead.js';
import { ResourceType } from '../../proto/api.js';
import { ActionId } from '../../proto_utils/action_id.js';
import { CastLog, DamageDealtLog, ResourceChangedLog } from '../../proto_utils/logs_parser.js';
import { resourceNames } from '../../proto_utils/names.js';
import { TypedEvent } from '../../typed_event.js';
import { formatToCompactNumber, formatToNumber } from '../../utils.js';
import { EventFilter } from './events_filter.js';
import { BUFF_GLOW, MISSING_GLOW, setGlow } from './replay_glow.js';
import {
	ActiveAura,
	activeAuraAt,
	AuraTrack,
	buildReplayModel,
	channelAt,
	cooldownAt,
	lastIndexAtOrBefore,
	placeCrits,
	ReplayBar,
	ReplayCast,
	ReplayChannel,
	ReplayEnemy,
	ReplayModel,
	SCROLL_LIFE,
	SctEvent,
	totemAt,
	TotemSlot,
	TrackedCooldown,
} from './replay_model.js';
import { LevelingTransport, ReplayMap } from './replay_map.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';
import { tooltipHit, tooltipRow, tooltipTitle } from './timeline.js';

const SPEEDS = [0.25, 0.5, 0.75, 1, 1.25, 1.5, 2, 3, 4];
// A leveling session is 20 minutes, much of it searching and resting, so it plays faster too.
const LEVELING_SPEEDS = [...SPEEDS, 8, 16, 32];
// How often we tell the Events tab where we are while we play, in milliseconds.
const TIME_EMIT_MS = 100;
const SPEED_STORAGE_KEY = '__detailed_results_replay_speed__';

// How far apart the scrolling lines stay, in pixels. A new line starts at the frame or under
// the cast history, and pushes the lines too close ahead of it further along, like MSBT.
const LANE_PX = 28;
// The crits on each side take spots between the cast history and the procs over the class pack,
// as many as fit at this spacing, up to 4. The first crit takes a middle spot.
const CRIT_SPOT_PX = 42;
const MAX_CRIT_SPOTS = 4;
// What we do to the enemies lines up on the left edge of the target's frame, so the icons
// form a column over it. What happens to us lines up on the right edge of our frame, with the
// icon after the text. Crits stay put in their own column, this many pixels toward the middle
// of the screen, so the lines that scroll by don't run over them.
const STICKY_GAP_PX = 32;
// What happens to us scrolls down from under the cast history to just above our frame. What we
// do scrolls up from just above the target's frame to under our buffs at the top right. This is
// the room we keep from each of them, about half the height of a crit.
const SCT_MARGIN_PX = 20;
// The longest way a line scrolls, in pixels. In full screen the area is much taller, and the lines
// would scroll too far and too fast. Outgoing lines still start at the target's frame and incoming
// lines still end at our frame.
const MAX_SCROLL_PX = 360;
// The narrowest stage that has the frames at its edges and the leveling map between them, in
// pixels. Narrower, the frames stay by the class pack and the combat text runs over the map.
const MAP_WIDE_PX = 1150;

// The cast history is a timeline at the top left that ends at the current time, right under the
// clock. Each cast sits at the moment we made it and slides left as time goes on. Casts too
// close together for their icons to fit side by side (like a racial between two spells) go up
// one row.
const HISTORY_PX_PER_SEC = 40;
const HISTORY_ICON_PX = 36;
const HISTORY_ROWS = 2;
const HISTORY_ROW_PX = 28;
// How far left of the timeline's right end the current time is, so the newest icon fits and
// the axis fades out after it.
const HISTORY_NOW_PX = 26;
// The axis has a tick every second and a time label every few seconds.
const HISTORY_LABEL_EVERY = 5;
// A new cast pops in over this many seconds.
const HISTORY_SHIFT = 0.15;
// A cooldown icon lights up like a pressed button for this long after we use it.
const PRESS_TIME = 0.25;
// An instant pops up where the cast bar is, a bit larger than the buttons, settles a moment, then
// grows and fades over this many seconds, like a cast bar that filled in no time. A few in a row
// ripple out one after another.
const INSTANT_LIFE = 0.9;
const INSTANT_POPS = 4;
// Instants pressed at the same moment, like a trinket and a racial, sit side by side this far
// apart, in pixels.
const INSTANT_SPACING = 64;
// A buff's countdown turns orange in its last seconds, like the Bayov pack's expiring alert.
const EXPIRING = 3;

const powerClass: Partial<Record<ResourceType, string>> = {
	[ResourceType.ResourceTypeMana]: 'mana',
	[ResourceType.ResourceTypeEnergy]: 'energy',
	[ResourceType.ResourceTypeRage]: 'rage',
	[ResourceType.ResourceTypeFocus]: 'focus',
};

// Like a video player clock, "1:05.3". Prepull times are negative, "-0:02.0".
const formatClock = (time: number, digits = 1): string => {
	const sign = time < 0 ? '-' : '';
	const abs = Math.abs(time);
	// We round first, so 119.97 reads 2:00.0 and not 1:60.0.
	const rounded = Math.round(abs * 10 ** digits) / 10 ** digits;
	const minutes = Math.floor(rounded / 60);
	const seconds = rounded - minutes * 60;
	const secondsText = seconds.toFixed(digits).padStart(digits > 0 ? 3 + digits : 2, '0');
	return `${sign}${minutes}:${secondsText}`;
};

// Like OmniCC: minutes for long cooldowns, whole seconds, then tenths in the last 3 seconds.
const formatCountdown = (seconds: number): string => {
	if (seconds >= 60) return `${Math.ceil(seconds / 60)}m`;
	if (seconds >= 3) return `${Math.ceil(seconds)}`;
	return seconds.toFixed(1);
};

const easeOut = (x: number) => 1 - (1 - x) * (1 - x);

// Where each cast goes on the history timeline: its row, and how many pixels left of its time
// we move it.
//
// We place the casts from the newest back. A cast goes in the lowest row where it doesn't cover
// the icon after it. At the pull we press Rage of the Farseer, Blood Fury, a Goblin Sapper Charge
// and a spell all at once, and there are more of them than rows. Then the older ones move left
// to the first free place in a row, so each icon still shows. The newest cast always sits at
// its time.
const historyLayout = (model: ReplayModel): Array<{ row: number; shift: number }> => {
	const step = HISTORY_ICON_PX + 2;
	const rowStarts = new Array<number>(HISTORY_ROWS).fill(Infinity);
	const layout = new Array<{ row: number; shift: number }>(model.actionStarts.length);
	for (let i = model.actionStarts.length - 1; i >= 0; i--) {
		const x = model.actionStarts[i] * HISTORY_PX_PER_SEC;
		let row = rowStarts.findIndex(start => x + step <= start);
		let placed = x;
		if (row < 0) {
			row = rowStarts.indexOf(Math.max(...rowStarts));
			placed = rowStarts[row] - step;
		}
		rowStarts[row] = placed;
		layout[i] = { row, shift: placed - x };
	}
	return layout;
};

const formatSeconds = (seconds: number): string => `${seconds.toFixed(1)}s`;

// The game's own tooltips of spells, buffs and items, from wowhead's Forever data, by
// "spell/17364", "aura/17364" or "item/13444". We ask once per id. Null while we wait, and
// stays null when wowhead doesn't have it.
const ingameTooltips = new Map<string, { content: DocumentFragment | null; loaded: Promise<void> }>();

// Wowhead's tooltip markup, without anything that could run.
const parseTooltip = (html: string): DocumentFragment => {
	const template = document.createElement('template');
	template.innerHTML = html;
	template.content.querySelectorAll('script, iframe, style, link, object, embed').forEach(elem => elem.remove());
	template.content.querySelectorAll('*').forEach(elem => {
		Array.from(elem.attributes).forEach(attribute => {
			if (attribute.name.startsWith('on')) elem.removeAttribute(attribute.name);
		});
	});
	return template.content;
};

// The mana bar's color shifts as the mana runs low, like the user's Bayov resource bar.
const manaColor = (fraction: number): string => (fraction < 0.15 ? '#af313a' : fraction < 0.3 ? '#af46ac' : fraction < 0.5 ? '#6358af' : '#577eaf');

// A tooltip title for a unit, with its portrait, name and level.
const unitTitle = (iconUrl: string, name: string, level: number): Element => (
	<div className="timeline-tooltip-title">
		{iconUrl && <img className="timeline-tooltip-icon" src={iconUrl} />}
		<span className="timeline-tooltip-name">{name}</span>
		<span className="timeline-tooltip-note">{`Level ${level}`}</span>
	</div>
);

const setIcon = (elem: HTMLElement, actionId: ActionId) => {
	elem.style.backgroundImage = actionId.iconUrl ? `url('${actionId.iconUrl}')` : '';
};

// An icon with a cooldown swipe, a countdown and stacks, like the buffs of a unit frame or the
// buttons of a cooldown tracker.
const makeIcon = (actionId: ActionId, className = ''): HTMLElement => {
	const elem = (
		<div className={`replay-icon ${className}`}>
			<div className="replay-icon-img" />
			<div className="replay-icon-swipe" />
			<span className="replay-icon-time" />
			<span className="replay-icon-stacks" />
		</div>
	) as HTMLElement;
	setIcon(elem.querySelector('.replay-icon-img')!, actionId);
	elem.dataset.name = actionId.name;
	return elem;
};

const updateIcon = (elem: HTMLElement, elapsed: number | null, time: string, stacks: number) => {
	elem.style.setProperty('--elapsed', elapsed == null ? '1' : elapsed.toFixed(4));
	elem.classList.toggle('swiping', elapsed != null);
	setText(elem.querySelector('.replay-icon-time')!, time);
	setText(elem.querySelector('.replay-icon-stacks')!, stacks > 1 ? String(stacks) : '');
};

// The size and opacity of an instant's pop, this many seconds after we pressed it.
const instantPop = (age: number): { scale: number; opacity: number } => {
	const p = age / INSTANT_LIFE;
	if (p < 0.12) {
		const k = p / 0.12;
		return { scale: 0.6 + 0.75 * (1 - (1 - k) * (1 - k)), opacity: k };
	}
	if (p < 0.45) return { scale: 1.35 - 0.1 * ((p - 0.12) / 0.33), opacity: 1 };
	const k = (p - 0.45) / 0.55;
	return { scale: 1.25 + 0.6 * k, opacity: 1 - k };
};

const setText = (elem: Element, text: string) => {
	if (elem.textContent != text) elem.textContent = text;
};

const setWidth = (elem: HTMLElement, fraction: number) => {
	elem.style.width = `${(Math.max(0, Math.min(1, fraction)) * 100).toFixed(3)}%`;
};

interface UnitFrameRefs {
	root: HTMLElement;
	health: HTMLElement;
	// The texts on the health bar's left and right, like EllesmereUI draws them.
	leftText: HTMLElement;
	rightText: HTMLElement;
	castBar: CastBarRefs;
	auras: HTMLElement;
	auraElems: Map<string, HTMLElement>;
}

interface CastBarRefs {
	root: HTMLElement;
	fill: HTMLElement;
	ticks: HTMLElement;
	icon: HTMLElement;
	name: HTMLElement;
	time: HTMLElement;
	shown: CastLog | ReplayChannel | null;
}

const makeCastBar = (className: string): CastBarRefs => {
	const root = (
		<div className={`replay-castbar ${className}`}>
			<div className="replay-castbar-icon" />
			<div className="replay-castbar-bar">
				<div className="replay-castbar-fill" />
				<div className="replay-castbar-ticks" />
				<span className="replay-castbar-name" />
				<span className="replay-castbar-time" />
			</div>
		</div>
	) as HTMLElement;
	return {
		root,
		fill: root.querySelector('.replay-castbar-fill')!,
		ticks: root.querySelector('.replay-castbar-ticks')!,
		icon: root.querySelector('.replay-castbar-icon')!,
		name: root.querySelector('.replay-castbar-name')!,
		time: root.querySelector('.replay-castbar-time')!,
		shown: null,
	};
};

// The cast in progress at this time, from casts in order of their start.
const castAt = (casts: Array<CastLog>, starts: Array<number>, time: number): CastLog | null => {
	const last = lastIndexAtOrBefore(starts, time);
	for (let i = last; i >= 0 && i > last - 8; i--) {
		const cast = casts[i];
		if (cast.castTime > 0 && time < cast.timestamp + cast.castTime) return cast;
	}
	return null;
};

// A cast fills the bar as it goes. A channel empties it, with a marker where each tick lands,
// like EllesmereUI's: tick k lands when the bar is down to 1 - k / ticks. The last one lands as it
// empties, so it has no marker. A pushback empties the bar some at once, past the ticks it cut.
const updateCastBar = (bar: CastBarRefs, shown: CastLog | ReplayChannel | null, time: number) => {
	bar.root.classList.toggle('active', !!shown);
	if (!shown) return;
	const channel = shown instanceof CastLog ? null : shown;
	if (bar.shown != shown) {
		bar.shown = shown;
		const actionId = channel ? channel.actionId : (shown as CastLog).actionId!;
		setIcon(bar.icon, actionId);
		setText(bar.name, actionId.name);
		const duration = channel ? channel.fullEnd - channel.start : 0;
		const ticks = channel ? Array.from({ length: Math.max(0, channel.ticks - 1) }, (_, i) => 1 - ((i + 1) * channel.tickLength) / duration) : [];
		bar.ticks.replaceChildren(
			...ticks
				.filter(at => at > 0.001 && at < 0.999)
				.map(at => {
					const tick = (<div className="replay-castbar-tick" />) as HTMLElement;
					tick.style.left = `${(at * 100).toFixed(3)}%`;
					return tick;
				}),
		);
	}
	if (channel) {
		const duration = channel.fullEnd - channel.start;
		const remaining = Math.max(0, channel.endsAt.at(time) - time);
		setWidth(bar.fill, remaining / duration);
		setText(bar.time, `${remaining.toFixed(1)} / ${duration.toFixed(1)}`);
		return;
	}
	const cast = shown as CastLog;
	const elapsed = time - cast.timestamp;
	setWidth(bar.fill, elapsed / cast.castTime);
	setText(bar.time, `${elapsed.toFixed(1)} / ${cast.castTime.toFixed(1)}`);
};

// Keeps the icons of the auras that are up in the host, in order, and reuses the icon of an
// aura that stays up.
const syncAuras = (
	host: HTMLElement,
	pool: Map<string, HTMLElement>,
	auras: Array<ActiveAura>,
	className: string,
	onCreate: (elem: HTMLElement, track: AuraTrack) => void,
) => {
	const keys = new Set(auras.map(aura => aura.track.key));
	pool.forEach((elem, key) => {
		if (!keys.has(key)) {
			elem.remove();
			pool.delete(key);
		}
	});
	auras.forEach((aura, i) => {
		let elem = pool.get(aura.track.key);
		if (!elem) {
			elem = makeIcon(aura.track.actionId, className);
			onCreate(elem, aura.track);
			pool.set(aura.track.key, elem);
		}
		if (host.children[i] !== elem) host.insertBefore(elem, host.children[i] ?? null);
		elem.classList.toggle('expiring', aura.remaining != null && aura.remaining < EXPIRING);
		updateIcon(elem, aura.fraction == null ? null : 1 - aura.fraction, aura.remaining == null ? '' : formatCountdown(aura.remaining), aura.stacks);
	});
};

const activeAuras = (tracks: Array<AuraTrack>, time: number, end: number): Array<ActiveAura> =>
	tracks
		.map(track => activeAuraAt(track, time, end))
		.filter((aura): aura is ActiveAura => !!aura)
		.sort((a, b) => a.span.chainStart - b.span.chainStart);

export interface ReplayConfig extends ResultComponentConfig {
	// The filter of the Events tab. The replay highlights what its search matches.
	filter: EventFilter;
}

// The Replay view of the Events tab. It plays one iteration back like a video of the fight: our
// unit frame on the left, the enemies on the right, a cooldown tracker in the middle (like a
// class pack WeakAura) and scrolling combat text on both sides (like MSBT).
//
// The state on screen is worked out from the time alone, so seeking draws the same frame as
// playing up to it.
export class Replay extends ResultComponent {
	private model: ReplayModel | null = null;
	private time = 0;
	private playing = false;
	private speed = 1;
	private shown = false;
	private frameRequest = 0;
	private lastFrameAt = 0;
	private wasPlayingBeforeScrub = false;

	// The time we play at, for the session strip of the Events tab. It fires about ten times a
	// second while we play, and on every seek.
	readonly timeEmitter = new TypedEvent<number>('ReplayTime');
	private lastTimeEmit = 0;
	private emittedTime = NaN;
	// A leveling session has a map in the middle and controls of its own. It plays the stretch
	// picked on the session strip (the filter's From and To), or all of it.
	private readonly map: ReplayMap;
	private readonly transport: LevelingTransport;
	private playRange = { start: 0, end: 0 };
	private killTimes: Array<number> = [];

	private readonly ui: {
		player: HTMLElement;
		stage: HTMLElement;
		message: HTMLElement;
		flash: HTMLElement;
		countdown: HTMLElement;
		hudTime: HTMLElement;
		hudDps: HTMLElement;
		hudDamage: HTMLElement;
		hudKills: HTMLElement;
		playerCol: HTMLElement;
		enemyCol: HTMLElement;
		history: HTMLElement;
		historyAxis: HTMLElement;
		procs: HTMLElement;
		coreCooldowns: HTMLElement;
		majorCooldowns: HTMLElement;
		center: HTMLElement;
		alert: HTMLElement;
		resources: HTMLElement;
		maintain: HTMLElement;
		sctZone: HTMLElement;
		sctIn: HTMLElement;
		sctOut: HTMLElement;
		progress: HTMLElement;
		played: HTMLElement;
		hover: HTMLElement;
		thumb: HTMLElement;
		tooltip: HTMLElement;
		markers: HTMLElement;
		playButton: HTMLButtonElement;
		clock: HTMLElement;
		speedButton: HTMLButtonElement;
		speedMenu: HTMLElement;
		fullscreenButton: HTMLButtonElement;
	};

	// The parts of the screen we build for each iteration.
	private playerFrame: UnitFrameRefs | null = null;
	private resourceRefs: {
		powers: Array<{ bar: ReplayBar; fill: HTMLElement; text: HTMLElement }>;
		swings: Array<{ fill: HTMLElement; times: Array<number>; mainHand: boolean }>;
		pips: Array<HTMLElement>;
	} | null = null;
	private playerCastBar = makeCastBar('player');
	private instantPops = Array.from({ length: INSTANT_POPS }, () => {
		const elem = (
			<div className="replay-instant">
				<div className="replay-icon-img" />
			</div>
		) as HTMLElement;
		return { elem, img: elem.firstElementChild as HTMLElement, shown: null as ReplayCast | null };
	});
	private playerCasts: Array<CastLog> = [];
	private enemyFrames: Array<UnitFrameRefs & { enemy: ReplayEnemy; starts: Array<number> }> = [];
	private cooldownViews: Array<{ cooldown: TrackedCooldown; elem: HTMLElement }> = [];
	// The shaman pack's maintain cells: the shield, the four totem slots and Flame Shock.
	private maintainViews: {
		shield: HTMLElement | null;
		totems: Array<{ slot: TotemSlot; elem: HTMLElement; shown: string }>;
		flameShock: HTMLElement | null;
		shieldShown: string;
		flameShockShown: string;
	} | null = null;
	private procElems = new Map<string, HTMLElement>();
	// The buffs that come and go, like procs. A buff of an ability in the cooldown tracker shows
	// on the ability's button instead.
	private procTracks: Array<AuraTrack> = [];
	// The iteration the frames were last lined up with the resource bar for.
	private alignedFor: ReplayModel | null = null;
	private mapWide = false;
	private historyElems = new Map<number, HTMLElement>();
	private historyLabels = new Map<number, HTMLElement>();
	// Where each cast goes on the history timeline, by its index in the model's actions.
	private historyLayout: Array<{ row: number; shift: number }> = [];
	private sctElems = new Map<number, HTMLElement>();
	// How many crit spots the model's crits are placed on.
	private critSpots = 0;

	// The hover tooltip. The parts of the screen move, come and go many times a second, so one
	// tooltip serves all of them. It follows the part under the mouse and shows its state at the
	// time we're at, so it updates while we play.
	private tooltip!: Instance;
	private readonly tips = new WeakMap<Element, () => Element | null>();
	private hovered: Element | null = null;
	private pointer: { x: number; y: number } | null = null;

	// The search highlights what it matches, on screen and on the progress bar. It's the search
	// of the Events tab, so it carries over to the timeline and the log.
	private readonly filter: EventFilter;
	private readonly searchTexts = new WeakMap<Element, string>();
	private matchTimes: Array<number> = [];
	private readonly search: { input: HTMLInputElement; count: HTMLElement; matches: HTMLElement };
	// The buffs that are up the whole fight, at the top right like the game's buff frame.
	private readonly buffsElem: HTMLElement;

	constructor(config: ReplayConfig) {
		config.rootCssClass = 'replay-root';
		super(config);
		this.filter = config.filter;

		try {
			const stored = parseFloat(window.localStorage.getItem(SPEED_STORAGE_KEY) ?? '');
			if (LEVELING_SPEEDS.includes(stored)) this.speed = stored;
		} catch {
			// The speed is a convenience, the replay works without it.
		}

		const refs = {
			player: ref<HTMLDivElement>(),
			stage: ref<HTMLDivElement>(),
			message: ref<HTMLDivElement>(),
			flash: ref<HTMLDivElement>(),
			countdown: ref<HTMLDivElement>(),
			hudTime: ref<HTMLSpanElement>(),
			hudDps: ref<HTMLSpanElement>(),
			hudDamage: ref<HTMLSpanElement>(),
			hudKills: ref<HTMLSpanElement>(),
			playerCol: ref<HTMLDivElement>(),
			enemyCol: ref<HTMLDivElement>(),
			history: ref<HTMLDivElement>(),
			historyAxis: ref<HTMLDivElement>(),
			procs: ref<HTMLDivElement>(),
			coreCooldowns: ref<HTMLDivElement>(),
			majorCooldowns: ref<HTMLDivElement>(),
			center: ref<HTMLDivElement>(),
			alert: ref<HTMLDivElement>(),
			resources: ref<HTMLDivElement>(),
			maintain: ref<HTMLDivElement>(),
			sctZone: ref<HTMLDivElement>(),
			sctIn: ref<HTMLDivElement>(),
			sctOut: ref<HTMLDivElement>(),
			progress: ref<HTMLDivElement>(),
			played: ref<HTMLDivElement>(),
			hover: ref<HTMLDivElement>(),
			thumb: ref<HTMLDivElement>(),
			tooltip: ref<HTMLDivElement>(),
			markers: ref<HTMLDivElement>(),
			playButton: ref<HTMLButtonElement>(),
			back: ref<HTMLButtonElement>(),
			forward: ref<HTMLButtonElement>(),
			clock: ref<HTMLSpanElement>(),
			speedButton: ref<HTMLButtonElement>(),
			speedMenu: ref<HTMLDivElement>(),
			fullscreenButton: ref<HTMLButtonElement>(),
		};

		this.rootElem.tabIndex = 0;
		this.rootElem.appendChild(
			<div ref={refs.player} className="replay-player">
				<div ref={refs.stage} className="replay-stage">
					<div className="replay-hud">
						<span ref={refs.hudTime} className="replay-hud-time" />
						<span className="replay-hud-stat">
							<span className="replay-hud-label">DPS</span>
							<span ref={refs.hudDps} className="replay-hud-value" />
						</span>
						<span className="replay-hud-stat damage">
							<span className="replay-hud-label">Damage</span>
							<span ref={refs.hudDamage} className="replay-hud-value" />
						</span>
						<span className="replay-hud-stat kills">
							<span className="replay-hud-label">Kills</span>
							<span ref={refs.hudKills} className="replay-hud-value" />
						</span>
					</div>
					<div className="replay-history">
						<div className="replay-history-now" />
						<div ref={refs.history} className="replay-history-casts" />
						<div ref={refs.historyAxis} className="replay-history-axis" />
					</div>
					<div ref={refs.sctZone} className="replay-sct-zone">
						<div ref={refs.sctIn} className="replay-sct replay-sct-in" />
						<div ref={refs.sctOut} className="replay-sct replay-sct-out" />
					</div>
					<div className="replay-frames">
						<div ref={refs.playerCol} className="replay-col replay-col-player" />
						<div ref={refs.center} className="replay-center">
							<div ref={refs.alert} className="replay-alert">
								<div className="replay-alert-arc left" />
								<div className="replay-alert-arc right" />
							</div>
							<div ref={refs.procs} className="replay-procs" />
							<div ref={refs.maintain} className="replay-maintain" />
							<div ref={refs.resources} className="replay-resources" />
							<div ref={refs.coreCooldowns} className="replay-cooldowns core" />
							<div ref={refs.majorCooldowns} className="replay-cooldowns major" />
							<div className="replay-cast-slot">
								{this.playerCastBar.root}
								<div className="replay-instants">{this.instantPops.map(pop => pop.elem)}</div>
							</div>
						</div>
						<div ref={refs.enemyCol} className="replay-col replay-col-enemies" />
					</div>
					<div ref={refs.message} className="replay-message hide" />
					<div ref={refs.countdown} className="replay-countdown" />
					<div className="replay-big-play">
						<i className="fas fa-play" />
					</div>
					<div ref={refs.flash} className="replay-flash">
						<i className="fas fa-play" />
					</div>
				</div>
				<div className="replay-controls">
					<div ref={refs.progress} className="replay-progress">
						<div className="replay-progress-track">
							<div ref={refs.hover} className="replay-progress-hover" />
							<div ref={refs.played} className="replay-progress-played" />
							<div ref={refs.markers} className="replay-progress-markers" />
						</div>
						<div ref={refs.thumb} className="replay-progress-thumb" />
						<div ref={refs.tooltip} className="replay-progress-tooltip" />
					</div>
					<div className="replay-buttons">
						<button ref={refs.playButton} className="replay-button" type="button" attributes={{ 'aria-label': 'Play (k)' }}>
							<i className="fas fa-play" />
						</button>
						<button ref={refs.back} className="replay-button small" type="button" attributes={{ 'aria-label': 'Back 5 seconds (←)' }}>
							<i className="fas fa-rotate-left" />
						</button>
						<button ref={refs.forward} className="replay-button small" type="button" attributes={{ 'aria-label': 'Forward 5 seconds (→)' }}>
							<i className="fas fa-rotate-right" />
						</button>
						<span ref={refs.clock} className="replay-clock" />
						<span className="replay-buttons-filler" />
						<span className="replay-keys">Space play · ← → 5s · , . step · &lt; &gt; speed · F full screen</span>
						<div className="replay-speed">
							<button ref={refs.speedButton} className="replay-button replay-speed-button" type="button" attributes={{ 'aria-label': 'Speed' }} />
							<div ref={refs.speedMenu} className="replay-speed-menu hide">
								<div className="replay-speed-title">Speed</div>
							</div>
						</div>
						<button ref={refs.fullscreenButton} className="replay-button" type="button" attributes={{ 'aria-label': 'Full screen (f)' }}>
							<i className="fas fa-expand" />
						</button>
					</div>
				</div>
			</div>,
		);

		this.ui = {
			player: refs.player.value!,
			stage: refs.stage.value!,
			message: refs.message.value!,
			flash: refs.flash.value!,
			countdown: refs.countdown.value!,
			hudTime: refs.hudTime.value!,
			hudDps: refs.hudDps.value!,
			hudDamage: refs.hudDamage.value!,
			hudKills: refs.hudKills.value!,
			playerCol: refs.playerCol.value!,
			enemyCol: refs.enemyCol.value!,
			history: refs.history.value!,
			historyAxis: refs.historyAxis.value!,
			procs: refs.procs.value!,
			coreCooldowns: refs.coreCooldowns.value!,
			majorCooldowns: refs.majorCooldowns.value!,
			center: refs.center.value!,
			alert: refs.alert.value!,
			resources: refs.resources.value!,
			maintain: refs.maintain.value!,
			sctZone: refs.sctZone.value!,
			sctIn: refs.sctIn.value!,
			sctOut: refs.sctOut.value!,
			progress: refs.progress.value!,
			played: refs.played.value!,
			hover: refs.hover.value!,
			thumb: refs.thumb.value!,
			tooltip: refs.tooltip.value!,
			markers: refs.markers.value!,
			playButton: refs.playButton.value!,
			clock: refs.clock.value!,
			speedButton: refs.speedButton.value!,
			speedMenu: refs.speedMenu.value!,
			fullscreenButton: refs.fullscreenButton.value!,
		};

		LEVELING_SPEEDS.forEach(speed => {
			const option = (
				<button className="replay-speed-option" type="button">
					{speed == 1 ? 'Normal' : `${speed}x`}
				</button>
			) as HTMLButtonElement;
			option.dataset.speed = String(speed);
			option.addEventListener('click', event => {
				event.stopPropagation();
				this.setSpeed(speed);
				this.ui.speedMenu.classList.add('hide');
			});
			this.ui.speedMenu.appendChild(option);
		});
		this.updateSpeed();

		this.ui.playButton.addEventListener('click', () => this.togglePlay());
		refs.back.value!.addEventListener('click', () => this.seek(this.time - 5));
		refs.forward.value!.addEventListener('click', () => this.seek(this.time + 5));
		this.transport = new LevelingTransport({ time: () => this.time, seek: time => this.seek(time), tip: (elem, content) => this.tip(elem, content) }, this.filter);
		refs.forward.value!.after(this.transport.root);
		this.map = new ReplayMap({ tip: (elem, content) => this.tip(elem, content) });
		this.ui.sctZone.prepend(this.map.root);
		this.ui.speedButton.addEventListener('click', event => {
			event.stopPropagation();
			this.ui.speedMenu.classList.toggle('hide');
		});
		// A leveling session on a wide enough stage has its frames at the edges, with the map
		// between them (_replay_map.scss). The combat text follows the frames.
		new ResizeObserver(() => this.updateMapWide()).observe(this.ui.player);
		this.ui.fullscreenButton.addEventListener('click', () => this.toggleFullscreen());
		document.addEventListener('fullscreenchange', () => {
			const full = document.fullscreenElement == this.ui.player;
			this.ui.player.classList.toggle('fullscreen', full);
			this.ui.fullscreenButton.querySelector('i')!.className = `fas ${full ? 'fa-compress' : 'fa-expand'}`;
		});
		[
			[this.ui.playButton, 'Play or pause (Space or K)'],
			[refs.back.value!, 'Back 5 seconds (←)'],
			[refs.forward.value!, 'Forward 5 seconds (→)'],
			[this.ui.speedButton, 'Playback speed (< and >)'],
			[this.ui.fullscreenButton, 'Full screen (F)'],
		].forEach(([elem, content]) => tippy(elem as HTMLElement, { content: content as string }));

		// A click on the screen plays or pauses, like a video. A double click goes full screen.
		this.ui.stage.addEventListener('click', () => {
			this.ui.speedMenu.classList.add('hide');
			this.togglePlay(true);
		});
		this.ui.stage.addEventListener('dblclick', () => this.toggleFullscreen());
		// The keys work while the replay has focus. A click on a button would move the focus to
		// the button, and then Space would press the button as well as play or pause. A click in
		// the search box keeps the focus there, so we can type in it.
		this.ui.player.addEventListener('pointerup', event => {
			if ((event.target as HTMLElement).closest('input')) return;
			this.rootElem.focus({ preventScroll: true });
		});
		document.addEventListener('click', () => this.ui.speedMenu.classList.add('hide'));

		this.bindProgress();
		this.bindTooltip();

		this.buffsElem = (<div className="replay-buffs" />) as HTMLElement;
		this.ui.stage.appendChild(this.buffsElem);
		this.search = this.buildSearch();
		this.filter.changeEmitter.on(() => {
			if (document.activeElement != this.search.input && this.search.input.value != this.filter.search) {
				this.search.input.value = this.filter.search;
			}
			this.updateSearch();
			this.applyRange();
		});
		this.rootElem.addEventListener('keydown', event => this.onKey(event));
	}

	onSimResult(resultData: SimResultData) {
		this.pause();
		const model = buildReplayModel(resultData.result, resultData.filter);
		if (typeof model == 'string') {
			this.model = null;
			this.ui.message.textContent = model;
			this.ui.message.classList.remove('hide');
			this.ui.player.classList.add('empty');
			return;
		}
		// An old result plays without its cooldowns. We say why, under the clock.
		this.ui.message.textContent = 'This result has no cooldown data. Sim again to see the cooldowns.';
		this.ui.message.classList.toggle('hide', model.hasCooldownData);
		this.ui.message.classList.toggle('notice', !model.hasCooldownData);
		this.ui.player.classList.remove('empty');
		this.model = model;
		this.setLeveling(model);
		this.updateMapWide();
		this.playRange = this.rangeFor(model);
		this.time = this.playRange.start;
		this.build(model);
		this.updateSearch();
		this.draw();
	}

	// The Events tab tells us when the Replay view opens and closes. We stop playing when it
	// closes, like a video in a background tab.
	setShown(shown: boolean) {
		this.shown = shown;
		if (!shown) {
			this.pause();
		} else {
			this.draw();
		}
	}

	private isVisible(): boolean {
		return this.shown && this.rootElem.offsetParent != null;
	}

	private togglePlay(withFlash = false) {
		if (!this.model) return;
		if (this.playing) {
			this.pause();
		} else {
			this.play();
		}
		if (withFlash) this.flash(this.playing ? 'fa-play' : 'fa-pause');
	}

	// Sets the replay up for a leveling session, or back for a fight: the map, the session's
	// controls and its faster speeds.
	private updateMapWide() {
		const wide = !!this.model?.leveling && this.ui.player.clientWidth >= MAP_WIDE_PX;
		if (wide == this.mapWide) return;
		this.mapWide = wide;
		this.ui.player.classList.toggle('map-wide', wide);
		// The frames moved, so they line up with the resource bar again.
		this.alignedFor = null;
	}

	private setLeveling(model: ReplayModel) {
		const leveling = model.leveling;
		this.ui.player.classList.toggle('leveling', !!leveling);
		this.map.setData(
			leveling
				? {
						session: leveling,
						playerLevel: model.playerLevel,
						playerName: model.playerName,
						classColor: `var(--bs-${model.player.classColor})`,
						totems: model.shaman?.totems ?? [],
						mobHealth: (targetIndex, time) => model.enemies.find(enemy => enemy.unit.index == targetIndex)?.healthFraction(time) ?? 1,
						mobIcon: model.levelingMobIcon,
						focusAt: time => model.focus.at(time),
					}
				: null,
		);
		this.transport.setSession(leveling);
		this.killTimes = leveling ? leveling.kills.map(kill => kill.t) : [];
		const speeds = this.speeds();
		this.ui.speedMenu.querySelectorAll<HTMLElement>('.replay-speed-option').forEach(option => {
			option.classList.toggle('hide', !speeds.includes(parseFloat(option.dataset.speed!)));
		});
		if (!speeds.includes(this.speed)) {
			this.speed = speeds[speeds.length - 1];
			this.updateSpeed();
		}
	}

	private speeds(): Array<number> {
		return this.model?.leveling ? LEVELING_SPEEDS : SPEEDS;
	}

	// What we play: in a leveling session the stretch picked on the session strip, and otherwise
	// the whole iteration.
	private rangeFor(model: ReplayModel): { start: number; end: number } {
		if (!model.leveling) return { start: model.start, end: model.end };
		const clamp = (t: number) => Math.max(model.start, Math.min(model.end, t));
		const start = this.filter.from == null ? model.start : clamp(this.filter.from);
		const end = this.filter.to == null ? model.end : clamp(this.filter.to);
		return end > start ? { start, end } : { start: model.start, end: model.end };
	}

	// A new stretch picked on the session strip: we play that, from its start.
	private applyRange() {
		const model = this.model;
		if (!model?.leveling) return;
		const range = this.rangeFor(model);
		if (range.start == this.playRange.start && range.end == this.playRange.end) return;
		this.playRange = range;
		this.drawProgressMarks(model);
		this.updateSearch();
		this.seek(range.start);
	}

	private play() {
		if (!this.model || this.playing) return;
		if (this.time >= this.playRange.end) this.time = this.playRange.start;
		this.playing = true;
		this.lastFrameAt = performance.now();
		this.frameRequest = requestAnimationFrame(now => this.tick(now));
		this.rootElem.focus({ preventScroll: true });
		this.updateButtons();
	}

	private pause() {
		this.playing = false;
		cancelAnimationFrame(this.frameRequest);
		if (this.ui) this.updateButtons();
	}

	private tick(now: number) {
		if (!this.playing || !this.model) return;
		if (!this.isVisible()) {
			this.pause();
			return;
		}
		// A long gap means the browser held the frames back, like in a background tab. We don't
		// jump ahead by all of it.
		const elapsed = Math.min(0.25, (now - this.lastFrameAt) / 1000);
		this.lastFrameAt = now;
		this.time = Math.min(this.playRange.end, this.time + elapsed * this.speed);
		this.draw(true);
		if (this.time >= this.playRange.end) {
			this.pause();
			return;
		}
		this.frameRequest = requestAnimationFrame(next => this.tick(next));
	}

	// Moves the replay to a time, kept within what we play.
	seek(time: number) {
		if (!this.model) return;
		this.time = Math.max(this.playRange.start, Math.min(this.playRange.end, time));
		this.draw();
	}

	private setSpeed(speed: number) {
		this.speed = speed;
		try {
			window.localStorage.setItem(SPEED_STORAGE_KEY, String(speed));
		} catch {
			// See the constructor.
		}
		this.updateSpeed();
	}

	private updateSpeed() {
		this.ui.speedButton.textContent = `${this.speed}x`;
		this.ui.speedMenu.querySelectorAll<HTMLElement>('.replay-speed-option').forEach(option => {
			option.classList.toggle('active', parseFloat(option.dataset.speed!) == this.speed);
		});
	}

	private stepSpeed(direction: number) {
		const speeds = this.speeds();
		const i = speeds.indexOf(this.speed);
		const next = speeds[Math.max(0, Math.min(speeds.length - 1, i + direction))];
		this.setSpeed(next);
		this.flash(null, `${next}x`);
	}

	private toggleFullscreen() {
		if (document.fullscreenElement == this.ui.player) {
			document.exitFullscreen();
		} else {
			this.ui.player.requestFullscreen?.();
		}
	}

	// The big icon that pops up in the middle and fades, when we play, pause or change speed.
	private flash(icon: string | null, text = '') {
		const flash = this.ui.flash;
		flash.replaceChildren(icon ? <i className={`fas ${icon}`} /> : <span>{text}</span>);
		flash.classList.remove('show');
		void flash.offsetWidth;
		flash.classList.add('show');
	}

	private onKey(event: KeyboardEvent) {
		if (!this.model || event.ctrlKey || event.metaKey || event.altKey) return;
		const target = event.target as HTMLElement;
		if (target.closest('input, textarea, select')) return;
		if (target.closest('button') && (event.key == ' ' || event.key == 'Enter')) return;
		const model = this.model;
		const { start, end } = this.playRange;
		const span = end - start;
		switch (event.key) {
			case ' ':
			case 'k':
				this.togglePlay(true);
				break;
			case 'ArrowLeft':
				this.seek(this.time - 5);
				break;
			case 'ArrowRight':
				this.seek(this.time + 5);
				break;
			case 'j':
				this.seek(this.time - 10);
				break;
			case 'l':
				this.seek(this.time + 10);
				break;
			case ',':
				this.pause();
				this.seek(this.time - 0.1);
				break;
			case '.':
				this.pause();
				this.seek(this.time + 0.1);
				break;
			case '<':
				this.stepSpeed(-1);
				break;
			case '>':
				this.stepSpeed(1);
				break;
			case 'f':
				this.toggleFullscreen();
				break;
			case 'Home':
				this.seek(start);
				break;
			case 'End':
				this.seek(end);
				break;
			// A leveling session's sections: the previous and the next one, and on to the next pull.
			case '[':
				if (!model.leveling) return;
				this.transport.step(-1);
				break;
			case ']':
				if (!model.leveling) return;
				this.transport.step(1);
				break;
			case 's':
				if (!model.leveling) return;
				this.transport.skip();
				break;
			default:
				if (/^[0-9]$/.test(event.key)) {
					this.seek(start + (span * parseInt(event.key)) / 10);
					break;
				}
				return;
		}
		event.preventDefault();
	}

	private bindProgress() {
		const progress = this.ui.progress;
		const timeAt = (clientX: number) => {
			const rect = progress.getBoundingClientRect();
			const fraction = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
			const { start, end } = this.playRange;
			return this.model ? start + fraction * (end - start) : 0;
		};
		const showHover = (clientX: number) => {
			if (!this.model) return;
			const rect = progress.getBoundingClientRect();
			const x = Math.max(0, Math.min(rect.width, clientX - rect.left));
			this.ui.hover.style.width = `${x}px`;
			this.ui.tooltip.style.left = `${x}px`;
			this.ui.tooltip.textContent = formatClock(timeAt(clientX));
		};

		let scrubbing = false;
		progress.addEventListener('pointerdown', event => {
			if (!this.model || event.button != 0) return;
			event.stopPropagation();
			scrubbing = true;
			progress.setPointerCapture(event.pointerId);
			progress.classList.add('scrubbing');
			this.wasPlayingBeforeScrub = this.playing;
			this.pause();
			this.seek(timeAt(event.clientX));
			showHover(event.clientX);
		});
		progress.addEventListener('pointermove', event => {
			showHover(event.clientX);
			if (scrubbing) this.seek(timeAt(event.clientX));
		});
		const stopScrub = (event: PointerEvent) => {
			if (!scrubbing) return;
			scrubbing = false;
			progress.releasePointerCapture(event.pointerId);
			progress.classList.remove('scrubbing');
			if (this.wasPlayingBeforeScrub) this.play();
		};
		progress.addEventListener('pointerup', stopScrub);
		progress.addEventListener('pointercancel', stopScrub);
		progress.addEventListener('click', event => event.stopPropagation());
		progress.addEventListener('pointerenter', event => showHover(event.clientX));
	}

	private buildSearch() {
		const input = (<input type="search" className="replay-search-input" placeholder="Highlight abilities and events" />) as HTMLInputElement;
		const count = (<span className="replay-search-count" />) as HTMLElement;
		const prev = (
			<button className="replay-search-step" type="button" attributes={{ 'aria-label': 'Previous match (Shift + Enter)' }}>
				<i className="fas fa-chevron-left" />
			</button>
		) as HTMLButtonElement;
		const next = (
			<button className="replay-search-step" type="button" attributes={{ 'aria-label': 'Next match (Enter)' }}>
				<i className="fas fa-chevron-right" />
			</button>
		) as HTMLButtonElement;
		const box = (
			<div className="replay-search">
				<i className="fas fa-magnifying-glass replay-search-icon" />
				{input}
				{count}
				{prev}
				{next}
			</div>
		) as HTMLElement;
		// Clicks in the search don't play or pause, like clicks on the screen do.
		['click', 'dblclick', 'pointerdown'].forEach(type => box.addEventListener(type, event => event.stopPropagation()));
		input.value = this.filter.search;
		input.addEventListener(
			'input',
			debounce(() => this.filter.setSearch(input.value), 150),
		);
		input.addEventListener('keydown', event => {
			if (event.key == 'Enter') {
				this.filter.setSearch(input.value);
				this.stepMatch(event.shiftKey ? -1 : 1);
				event.preventDefault();
			} else if (event.key == 'Escape') {
				input.blur();
			}
		});
		prev.addEventListener('click', () => this.stepMatch(-1));
		next.addEventListener('click', () => this.stepMatch(1));
		tippy(prev, { content: 'Previous match (Shift + Enter)' });
		tippy(next, { content: 'Next match (Enter)' });
		this.ui.stage.appendChild(box);

		const matches = (<div className="replay-progress-matches" />) as HTMLElement;
		this.ui.markers.before(matches);
		return { input, count, matches };
	}

	// Marks a part of the screen the search can match, with the text it looks in.
	private searchable(elem: Element, text: string) {
		elem.classList.add('replay-searchable');
		this.searchTexts.set(elem, text.toLowerCase());
		this.applyMatch(elem);
	}

	private applyMatch(elem: Element) {
		const text = this.searchTexts.get(elem);
		elem.classList.toggle('match', !!text && this.filter.search.trim() != '' && this.filter.matchesText(text));
	}

	// Finds the times of the matches for the progress bar, and highlights what's on screen.
	private updateSearch() {
		const model = this.model;
		const searching = this.filter.search.trim() != '';
		this.ui.player.classList.toggle('searching', searching && !!model);
		this.ui.stage.querySelectorAll('.replay-searchable').forEach(elem => this.applyMatch(elem));

		const times: Array<number> = [];
		if (model && searching) {
			model.casts.forEach(cast => this.filter.matchesText(cast.searchText) && times.push(cast.log.timestamp));
			model.sct.forEach(event => this.filter.matchesText(event.searchText) && times.push(event.time));
			[...model.playerAuras, ...model.enemies.flatMap(enemy => enemy.debuffs)].forEach(track => {
				if (!track.alwaysUp && this.filter.matchesText(track.actionId.name.toLowerCase())) {
					track.spans.forEach(span => span.start == span.chainStart && times.push(span.start));
				}
			});
		}
		times.sort((a, b) => a - b);
		// The same moment can match a cast, its hit and its buff. We count it once.
		this.matchTimes = times.filter((time, i) => i == 0 || time - times[i - 1] > 0.005);

		this.search.count.textContent = searching ? `${formatToNumber(this.matchTimes.length)} ${this.matchTimes.length == 1 ? 'match' : 'matches'}` : '';
		this.search.count.classList.toggle('none', searching && this.matchTimes.length == 0);

		// A tick on the progress bar for each match. Many matches close together share a tick,
		// one for each of the bar's 400 slots.
		if (!model) return;
		const { start, end } = this.playRange;
		const span = end - start;
		const slots = new Set(this.matchTimes.filter(time => time >= start && time <= end).map(time => Math.round(((time - start) / span) * 400)));
		this.search.matches.replaceChildren(
			...Array.from(slots).map(slot => {
				const tick = (<div className="replay-progress-match" />) as HTMLElement;
				tick.style.left = `${slot / 4}%`;
				return tick;
			}),
		);
	}

	// Jumps to the next or the previous match. We land just after it, so its combat text is
	// already up.
	private stepMatch(direction: number) {
		if (this.matchTimes.length == 0) return;
		const landing = 0.08;
		const target =
			direction > 0
				? this.matchTimes.find(time => time + landing > this.time + 0.01)
				: [...this.matchTimes].reverse().find(time => time + landing < this.time - 0.01);
		if (target === undefined) return;
		this.seek(target + landing);
	}

	private bindTooltip() {
		this.tooltip = tippy(this.ui.stage, {
			trigger: 'manual',
			placement: 'top',
			content: '',
			maxWidth: 360,
			duration: [100, 0],
			// In full screen only what's inside the player shows, so the tooltip lives there too.
			appendTo: () => this.ui.player,
		});
		const track = (event: PointerEvent) => {
			this.pointer = { x: event.clientX, y: event.clientY };
			this.hover((event.target as Element).closest('.replay-tip'));
		};
		this.ui.stage.addEventListener('pointerover', track);
		this.ui.stage.addEventListener('pointermove', track);
		this.ui.stage.addEventListener('pointerleave', () => {
			this.pointer = null;
			this.hover(null);
		});
	}

	// Gives a part of the screen a hover tooltip.
	private tip(elem: Element, content: () => Element | null) {
		elem.classList.add('replay-tip');
		this.tips.set(elem, content);
	}

	private hover(elem: Element | null) {
		if (elem == this.hovered) return;
		this.hovered = elem;
		if (!elem) {
			this.tooltip.hide();
			return;
		}
		this.tooltip.setProps({ getReferenceClientRect: () => elem.getBoundingClientRect() });
		this.refreshTooltip();
	}

	private refreshTooltip() {
		const content = this.hovered ? this.tips.get(this.hovered)?.() : null;
		if (!content) {
			this.tooltip.hide();
			return;
		}
		this.tooltip.setContent(content);
		if (!this.tooltip.state.isShown) this.tooltip.show();
		this.tooltip.popperInstance?.update();
	}

	// While we play, the parts move under a mouse that stands still, like the combat text that
	// scrolls by. So after each frame we look again at what's under the mouse.
	private updateHover() {
		if (this.pointer) {
			const under = document.elementFromPoint(this.pointer.x, this.pointer.y);
			this.hover(under?.closest('.replay-tip') ?? null);
		}
		if (this.hovered && !this.hovered.isConnected) this.hover(null);
		if (this.hovered) this.refreshTooltip();
	}

	// A buff or debuff: since when it's up, how long it has left, its stacks and its uptime.
	private auraTip(track: AuraTrack, where: string): () => Element {
		const model = this.model!;
		const fightStart = Math.max(0, model.start);
		const uptime =
			track.spans.reduce((total, span) => total + Math.max(0, Math.min(span.end, model.end) - Math.max(span.start, fightStart)), 0) /
			(model.end - fightStart);
		const gains = track.spans.filter(span => span.start == span.chainStart).length;
		return () => {
			const aura = activeAuraAt(track, this.time, model.end);
			const meta = track.alwaysUp
				? `On ${where} the whole fight`
				: aura
				? `On ${where} since ${formatSeconds(aura.span.chainStart)}`
				: `Not on ${where} now`;
			return (
				<div className="timeline-tooltip">
					{tooltipTitle(track.actionId, aura && aura.stacks > 1 ? `${aura.stacks} stacks` : undefined)}
					<div className="timeline-tooltip-meta">{meta}</div>
					{!track.alwaysUp && (
						<div className="timeline-tooltip-rows">
							{aura?.remaining != null && tooltipRow('Time left', formatSeconds(aura.remaining))}
							{track.duration != null && tooltipRow('Duration', formatSeconds(track.duration))}
							{tooltipRow('Uptime', `${(uptime * 100).toFixed(1)}%`)}
							{tooltipRow('Times gained', String(gains))}
						</div>
					)}
					{this.ingameTooltip(track.actionId, true)}
				</div>
			) as Element;
		};
	}

	private cooldownTip(cooldown: TrackedCooldown): () => Element {
		return () => {
			const model = this.model!;
			const aura = cooldown.aura ? activeAuraAt(cooldown.aura, this.time, model.end) : null;
			const cd = cooldownAt(model, cooldown, this.time);
			const used = lastIndexAtOrBefore(cooldown.castTimes, this.time) + 1;
			const status = aura
				? aura.remaining != null
					? `Buff up, ${formatSeconds(aura.remaining)} left`
					: 'Buff up'
				: cd
				? `On cooldown, ${formatSeconds(cd.remaining)} left`
				: 'Ready';
			return (
				<div className="timeline-tooltip">
					{tooltipTitle(cooldown.actionId)}
					<div className="timeline-tooltip-meta">{status}</div>
					<div className="timeline-tooltip-rows">
						{tooltipRow('Cooldown', formatSeconds(cooldown.longest))}
						{tooltipRow('Used so far', `${used} of ${cooldown.castTimes.length}`)}
						{used > 0 && tooltipRow('Last used', formatSeconds(cooldown.castTimes[used - 1]))}
					</div>
					{this.ingameTooltip(cooldown.actionId)}
				</div>
			) as Element;
		};
	}

	// A cast: when, how long it took, and how its hits landed.
	private castTip(log: CastLog): Element {
		const timing = [formatSeconds(log.timestamp)];
		if (log.castTime > 0) timing.push(`${formatSeconds(log.castTime)} cast`);
		if (log.effectiveTime > 0) timing.push(`${formatSeconds(log.effectiveTime)} GCD`);
		const hits = log.damageDealtLogs;
		const total = hits.filter(hit => !hit.miss && !hit.dodge && !hit.parry).reduce((sum, hit) => sum + hit.amount, 0);
		return (
			<div className="timeline-tooltip">
				{tooltipTitle(log.actionId!)}
				<div className="timeline-tooltip-meta">{timing.join(' · ')}</div>
				{hits.length > 0 && <ul className="timeline-tooltip-hits">{hits.map(hit => tooltipHit(hit))}</ul>}
				{hits.length > 1 && <div className="timeline-tooltip-rows">{tooltipRow('Total', formatToNumber(Math.round(total)))}</div>}
				{this.ingameTooltip(log.actionId!)}
			</div>
		) as Element;
	}

	// A line of the combat text: the hits behind it, or the buff or resource we gained.
	private sctTip(event: SctEvent): Element {
		// A line of merged resource gains lists each gain.
		const gains = event.logs.filter((log): log is ResourceChangedLog => log.isResourceChanged());
		if (gains.length > 1) {
			const sources = new Set(gains.map(log => log.actionId!.name));
			return (
				<div className="timeline-tooltip">
					<div className="timeline-tooltip-title">
						<img className="timeline-tooltip-icon" src={event.iconUrl} />
						<span className="timeline-tooltip-name">{event.text}</span>
					</div>
					<div className="timeline-tooltip-meta">
						{`${gains.length} small gains from ${formatSeconds(gains[0].timestamp)} to ${formatSeconds(event.time)}`}
					</div>
					<div className="timeline-tooltip-rows">
						{gains.map(log =>
							tooltipRow(
								`${formatSeconds(log.timestamp)}${sources.size > 1 ? ` ${log.actionId!.name}` : ''}`,
								`+${formatToNumber(log.valueAfter - log.valueBefore, { maximumFractionDigits: 1 })}`,
							),
						)}
					</div>
					{sources.size == 1 && this.ingameTooltip(event.actionId)}
				</div>
			) as Element;
		}
		const hits = event.logs.filter((log): log is DamageDealtLog => log.isDamageDealt());
		const targets = Array.from(new Set(hits.map(hit => hit.target?.name).filter(Boolean)));
		const what = event.kind == 'aura' ? 'Gained' : event.kind == 'resource' ? event.text : '';
		return (
			<div className="timeline-tooltip">
				{tooltipTitle(event.actionId)}
				<div className="timeline-tooltip-meta">{[formatSeconds(event.time), what].filter(Boolean).join(' · ')}</div>
				{hits.length > 0 && <ul className="timeline-tooltip-hits">{hits.map(hit => tooltipHit(hit))}</ul>}
				{(hits.length > 1 || targets.length > 1) && (
					<div className="timeline-tooltip-rows">
						{hits.length > 1 &&
							tooltipRow(
								'Total',
								formatToNumber(Math.round(hits.filter(hit => !hit.miss && !hit.dodge && !hit.parry).reduce((sum, hit) => sum + hit.amount, 0))),
							)}
						{targets.length > 1 && tooltipRow('Targets', targets.join(', '))}
					</div>
				)}
				{this.ingameTooltip(event.actionId, event.kind == 'aura')}
			</div>
		) as Element;
	}

	private updateButtons() {
		const ended = !!this.model && this.time >= this.playRange.end;
		const icon = this.playing ? 'fa-pause' : ended ? 'fa-rotate-right' : 'fa-play';
		this.ui.playButton.querySelector('i')!.className = `fas ${icon}`;
		this.ui.player.classList.toggle('paused', !this.playing);
		// Like a video before we start it or after it ends, a big button in the middle asks us to
		// play.
		// A leveling session's map is what we look at, so we don't cover it.
		const atStart = !!this.model && this.time <= this.playRange.start;
		this.ui.player.classList.toggle('show-big-play', !this.playing && (atStart || ended) && !this.model?.leveling);
		this.ui.player.querySelector('.replay-big-play i')!.className = `fas ${ended ? 'fa-rotate-right' : 'fa-play'}`;
	}

	// Builds the frames, the cooldown tracker and the progress bar markers for a new iteration.
	private build(model: ReplayModel) {
		this.critSpots = 0;
		this.sctElems.forEach(elem => elem.remove());
		this.sctElems.clear();
		this.historyElems.forEach(elem => elem.remove());
		this.historyElems.clear();
		this.historyLabels.forEach(elem => elem.remove());
		this.historyLabels.clear();
		this.historyLayout = historyLayout(model);
		this.procElems.forEach(elem => elem.remove());
		this.procElems.clear();

		this.playerCasts = model.casts.map(cast => cast.log);
		// The buffs with a place of their own (a cooldown's button, a maintain cell or the pips)
		// aren't procs.
		const placed = new Set([
			...model.cooldowns.map(cooldown => cooldown.aura?.key),
			...(model.shaman?.shields ?? []).map(track => track.key),
			model.maelstrom?.key,
		]);
		this.procTracks = model.playerAuras.filter(track => !track.alwaysUp && !placed.has(track.key));
		this.buildPlayerFrame(model);
		this.buildEnemyFrames(model);

		// The cooldown tracker.
		this.ui.coreCooldowns.replaceChildren();
		this.ui.majorCooldowns.replaceChildren();
		this.cooldownViews = [
			...model.mainBar.map(cooldown => ({ cooldown, row: this.ui.coreCooldowns })),
			...model.utilityBar.map(cooldown => ({ cooldown, row: this.ui.majorCooldowns })),
		].map(({ cooldown, row }) => {
			const elem = makeIcon(cooldown.actionId, 'cooldown');
			row.appendChild(elem);
			this.tip(elem, this.cooldownTip(cooldown));
			this.searchable(elem, cooldown.actionId.name);
			return { cooldown, elem };
		});
		this.buildMaintain(model);

		this.drawProgressMarks(model);
		this.ui.player.style.setProperty('--class-color', `var(--bs-${model.player.classColor})`);
	}

	// The progress bar marks where we used a major cooldown, like chapters on a video. A leveling
	// session draws its states, kills and deaths along it instead, over the stretch we play.
	private drawProgressMarks(model: ReplayModel) {
		const { start, end } = this.playRange;
		const span = Math.max(end - start, 0.001);
		this.ui.markers.replaceChildren(
			...(model.leveling ? [] : model.markers).map(marker => {
				const elem = (<div className="replay-progress-marker" />) as HTMLElement;
				elem.style.left = `${((marker.time - start) / span) * 100}%`;
				return elem;
			}),
		);
		this.transport.drawProgress(this.ui.markers.parentElement!, start, end);
	}

	// A unit frame like the user's EllesmereUI ones: a flat health bar with a text on each side,
	// no portrait. The name and level are in the tooltip.
	private buildUnitFrame(side: 'player' | 'enemy'): UnitFrameRefs {
		const castBar = makeCastBar(side);
		const root = (
			<div className={`replay-unit ${side}`}>
				<div className="replay-unit-auras" />
				<div className="replay-unit-main">
					<div className="replay-bar health">
						<div className="replay-bar-fill" />
						<span className="replay-bar-left" />
						<span className="replay-bar-right" />
					</div>
				</div>
			</div>
		) as HTMLElement;
		root.appendChild(castBar.root);
		return {
			root,
			health: root.querySelector('.health .replay-bar-fill')!,
			leftText: root.querySelector('.health .replay-bar-left')!,
			rightText: root.querySelector('.health .replay-bar-right')!,
			castBar,
			auras: root.querySelector('.replay-unit-auras')!,
			auraElems: new Map(),
		};
	}

	// The resource block of the class pack, between the procs and the cooldowns: the power bar,
	// the swing timer and the pips (combo points, or the shaman's Maelstrom Weapon stacks).
	private buildResources(model: ReplayModel) {
		const powers = model.powers.map(bar => {
			const elem = (
				<div className={`replay-bar power ${powerClass[bar.type] ?? ''}`}>
					<div className="replay-bar-fill" />
					<span className="replay-bar-center" />
				</div>
			) as HTMLElement;
			return { bar, elem, fill: elem.querySelector<HTMLElement>('.replay-bar-fill')!, text: elem.querySelector<HTMLElement>('.replay-bar-center')! };
		});
		// A swing timer for each hand we swing with, the main hand above the off hand.
		const swings = [
			{ times: model.mainHandSwings, mainHand: true },
			{ times: model.offHandSwings, mainHand: false },
		]
			.filter(({ times }) => times.length > 1)
			.map(({ times, mainHand }) => {
				const elem = (
					<div className="replay-swing">
						<div className="replay-swing-fill" />
					</div>
				) as HTMLElement;
				return { elem, fill: elem.firstElementChild as HTMLElement, times, mainHand };
			});
		const pips: Array<HTMLElement> = [];
		const pipKind = model.comboPoints ? 'combo' : model.maelstrom ? 'maelstrom' : '';
		const pipRow = (<div className={`replay-pips ${pipKind}`} />) as HTMLElement;
		if (pipKind) {
			for (let i = 0; i < 5; i++) {
				const pip = (<div className="replay-pip" />) as HTMLElement;
				pipRow.appendChild(pip);
				pips.push(pip);
			}
		}
		// From the bottom, like the user's Bayov resource bars: power, the swings, then the pips.
		this.ui.resources.replaceChildren(...(pipKind ? [pipRow] : []), ...swings.map(swing => swing.elem), ...powers.map(power => power.elem));
		// Each bar has its own tooltip, so we see the one resource under the mouse.
		const rowsTip = (rows: () => Array<Element | null>) => () =>
			(
				<div className="timeline-tooltip">
					<div className="timeline-tooltip-rows">{rows()}</div>
				</div>
			) as Element;
		powers.forEach(({ bar, elem }) =>
			this.tip(
				elem,
				rowsTip(() => [
					tooltipRow(resourceNames.get(bar.type)!, `${formatToNumber(Math.floor(bar.value.at(this.time)))} / ${formatToNumber(bar.max)}`),
				]),
			),
		);
		swings.forEach(({ elem, times, mainHand }) =>
			this.tip(
				elem,
				rowsTip(() => {
					const next = this.swingAt(times, this.time);
					const queued = mainHand ? this.swingQueuedAt(model, this.time) : null;
					return [
						tooltipRow(`Next ${mainHand ? 'main' : 'off'} hand swing`, next ? formatSeconds(next.next - this.time) : 'None'),
						queued && tooltipRow('Queued', queued.name.replace(/ \(Queue\)/, '')),
					];
				}),
			),
		);
		if (model.maelstrom) {
			this.tip(pipRow, this.auraTip(model.maelstrom, 'us'));
			const maelstromTip = this.auraTip(model.maelstrom, 'us');
			this.tip(this.ui.alert, () => {
				const tip = maelstromTip();
				tip.insertBefore(
					<div className="timeline-tooltip-meta">
						These arcs light up at 5 stacks of Maelstrom Weapon, like a WeakAura spell alert. Each stack makes the next Lightning Bolt 4% faster and
						cheaper for each talent point, so with 5 points it's instant and free.
					</div>,
					tip.children[1],
				);
				return tip;
			});
		} else if (model.comboPoints) {
			this.tip(
				pipRow,
				rowsTip(() => [tooltipRow('Combo points', String(model.comboPoints!.at(this.time)))]),
			);
		}
		this.resourceRefs = { powers, swings, pips };
	}

	private pipsAt(model: ReplayModel, time: number): number {
		if (model.comboPoints) return model.comboPoints.at(time);
		if (!model.maelstrom) return 0;
		const aura = activeAuraAt(model.maelstrom, time, model.end);
		return aura ? Math.max(1, aura.stacks) : 0;
	}

	// The Heroic Strike (or Cleave, Maul, Raptor Strike) waiting for the next main hand swing.
	private swingQueuedAt(model: ReplayModel, time: number): ActionId | null {
		return model.swingQueue.find(track => activeAuraAt(track, time, model.end))?.actionId ?? null;
	}

	// The last and the next swing of a hand around this time. A long gap, like while we were out
	// of range, has no swing timer.
	private swingAt(swings: Array<number>, time: number): { last: number; next: number } | null {
		const i = lastIndexAtOrBefore(swings, time);
		if (i < 0 || i + 1 >= swings.length || swings[i + 1] - swings[i] > 5) return null;
		return { last: swings[i], next: swings[i + 1] };
	}

	// A cast bar's tooltip is the cast's, while the bar shows one. A channel's is the cast that
	// started it.
	private tipCastBar(bar: CastBarRefs) {
		this.tip(bar.root, () => {
			if (!bar.root.classList.contains('active') || !bar.shown) return null;
			const cast = bar.shown instanceof CastLog ? bar.shown : bar.shown.cast;
			return cast ? this.castTip(cast) : null;
		});
	}

	private buildPlayerFrame(model: ReplayModel) {
		const frame = this.buildUnitFrame('player');
		this.buildResources(model);

		// The buffs that are up the whole fight never change, so we draw them once, at the top
		// right like the game's buff frame.
		const buffs: Array<HTMLElement> = [];
		model.playerAuras
			.filter(track => track.alwaysUp)
			.forEach(track => {
				const elem = makeIcon(track.actionId, 'buff');
				buffs.push(elem);
				this.tip(elem, this.auraTip(track, 'us'));
				this.searchable(elem, track.actionId.name);
			});

		this.tip(frame.root.querySelector('.replay-unit-main')!, () => {
			const time = this.time;
			const damage = model.damageDone.at(time);
			return (
				<div className="timeline-tooltip">
					{unitTitle(model.player.iconUrl, model.playerName, model.playerLevel)}
					<div className="timeline-tooltip-rows">
						{model.health &&
							tooltipRow('Health', `${formatToNumber(Math.round(model.health.value.at(time)))} / ${formatToNumber(model.health.max)}`)}
						{model.powers.map(bar =>
							tooltipRow(resourceNames.get(bar.type)!, `${formatToNumber(Math.floor(bar.value.at(time)))} / ${formatToNumber(bar.max)}`),
						)}
						{model.comboPoints && tooltipRow('Combo points', String(model.comboPoints.at(time)))}
						{tooltipRow('Damage done', formatToNumber(Math.round(damage)))}
						{time >= 1 && tooltipRow('DPS so far', formatToNumber(damage / time, { maximumFractionDigits: 1 }))}
					</div>
				</div>
			) as Element;
		});
		this.tipCastBar(this.playerCastBar);

		this.buffsElem.replaceChildren(...buffs);
		this.ui.playerCol.replaceChildren(frame.root);
		this.playerFrame = frame;
	}

	private buildEnemyFrames(model: ReplayModel) {
		this.enemyFrames = model.enemies.map(enemy => {
			const frame = this.buildUnitFrame('enemy');
			frame.rightText.textContent = enemy.name;
			this.tip(frame.root.querySelector('.replay-unit-main')!, () => {
				const time = this.time;
				const damage = enemy.damageTaken.at(time);
				return (
					<div className="timeline-tooltip">
						{unitTitle(enemy.unit.iconUrl, enemy.name, enemy.level)}
						<div className="timeline-tooltip-rows">
							{tooltipRow('Health', `${(enemy.healthFraction(time) * 100).toFixed(1)}%`)}
							{tooltipRow('Damage taken', formatToNumber(Math.round(damage)))}
							{time >= 1 && tooltipRow('DPS taken so far', formatToNumber(damage / time, { maximumFractionDigits: 1 }))}
						</div>
						{model.timeBasedHealth && (
							<div className="timeline-tooltip-meta">
								The fight ends at a set time, so the health bar follows the time left, like the execute phases of the sim do.
							</div>
						)}
					</div>
				) as Element;
			});
			this.tipCastBar(frame.castBar);
			return { ...frame, enemy, starts: enemy.casts.map(cast => cast.timestamp) };
		});
		this.ui.enemyCol.replaceChildren(...this.enemyFrames.map(frame => frame.root));
		this.ui.enemyCol.classList.toggle('many', this.enemyFrames.length > 2);
	}

	// Draws the screen at the current time. A frame of play lets the map ease from the frame before,
	// see ReplayMap.draw.
	private draw(playing = false) {
		const model = this.model;
		if (!model || !this.isVisible()) return;
		const time = this.time;

		this.alignFrames(model);
		this.drawHud(model, time);
		this.drawPlayer(model, time);
		this.drawEnemies(model, time);
		this.drawCooldowns(model, time);
		this.drawMaintain(model, time);
		this.drawHistory(model, time);
		this.drawSct(model, time);
		this.drawCountdown(model, time);
		this.drawControls(model, time);
		if (model.leveling) {
			this.map.draw(time, playing ? this.speed : null);
			this.transport.update(time);
		}
		this.updateHover();
		this.emitTime(time);
	}

	// Tells the Events tab where we are, at most every TIME_EMIT_MS while we play.
	private emitTime(time: number) {
		if (time == this.emittedTime) return;
		const now = performance.now();
		if (this.playing && now - this.lastTimeEmit < TIME_EMIT_MS) return;
		this.lastTimeEmit = now;
		this.emittedTime = time;
		this.timeEmitter.emit(TypedEvent.nextEventID(), time);
	}

	// Puts the unit frames level with the resource bar of the class pack. Everything above the
	// bar keeps its size while we play, so we only measure once per iteration. With three or more
	// enemies their frames start at the top, so they fit.
	private alignFrames(model: ReplayModel) {
		if (this.alignedFor == model) return;
		const top = this.ui.resources.offsetTop;
		if (top <= 0) return;
		this.ui.playerCol.style.paddingTop = `${top}px`;
		// A leveling session has a frame for each mob of its pool, but only the few we pulled show.
		this.ui.enemyCol.style.paddingTop = this.enemyFrames.length > 2 && !model.leveling ? '0' : `${top}px`;
		this.alignedFor = model;
	}

	private drawHud(model: ReplayModel, time: number) {
		setText(this.ui.hudTime, formatClock(time));
		const damage = model.damageDone.at(time);
		setText(this.ui.hudDps, time >= 1 ? formatToNumber(damage / time, { maximumFractionDigits: 1, minimumFractionDigits: 1 }) : '–');
		setText(this.ui.hudDamage, formatToCompactNumber(damage, { maximumFractionDigits: 1 }));
		if (model.leveling) {
			const kills = lastIndexAtOrBefore(this.killTimes, time) + 1;
			setText(this.ui.hudKills, String(kills));
		}
	}

	// The pull timer in the middle of the screen while we're in the prepull, like the game's
	// countdown. Each number pops in big, shrinks to its size and fades out before the next one.
	private drawCountdown(model: ReplayModel, time: number) {
		const elem = this.ui.countdown;
		if (model.leveling || time >= 0) {
			elem.style.opacity = '0';
			return;
		}
		const number = Math.ceil(-time);
		// How far we are into this number's second, from 0 to 1.
		const progress = number + time;
		const scale = 1.6 - 0.6 * easeOut(Math.min(1, progress / 0.25));
		const opacity = Math.min(1, progress / 0.1, (1 - progress) / 0.25);
		setText(elem, String(number));
		elem.style.opacity = String(opacity);
		elem.style.transform = `translate(-50%, -50%) scale(${scale})`;
	}

	private drawPlayer(model: ReplayModel, time: number) {
		const frame = this.playerFrame!;
		if (model.health) {
			const health = model.health.value.at(time);
			setWidth(frame.health, health / model.health.max);
			// Current health and percent on the right, like EllesmereUI's "both" text.
			setText(frame.rightText, `${formatToCompactNumber(health, { maximumFractionDigits: 1 })} | ${Math.round((health / model.health.max) * 100)}%`);
		} else {
			setWidth(frame.health, 1);
		}

		const resources = this.resourceRefs!;
		resources.powers.forEach(({ bar, fill, text }) => {
			const value = bar.value.at(time);
			const fraction = value / bar.max;
			setWidth(fill, fraction);
			if (bar.type == ResourceType.ResourceTypeMana) fill.style.background = manaColor(fraction);
			setText(text, String(Math.floor(value)));
		});
		resources.swings.forEach(({ fill, times, mainHand }) => {
			const swing = this.swingAt(times, time);
			setWidth(fill, swing ? (time - swing.last) / (swing.next - swing.last) : 0);
			// The main hand's timer turns yellow while a Heroic Strike waits for its swing.
			if (mainHand) fill.classList.toggle('queued', !!this.swingQueuedAt(model, time));
		});
		const pips = this.pipsAt(model, time);
		resources.pips.forEach((pip, i) => pip.classList.toggle('on', i < pips));
		// At five Maelstrom Weapon stacks arcs light up on both sides of the pack, like the
		// PowerAuras textures of the user's WeakAura.
		this.ui.alert.classList.toggle('on', !!model.maelstrom && pips >= 5);

		// The procs and other buffs that come and go, in the middle above the cooldowns.
		syncAuras(this.ui.procs, this.procElems, activeAuras(this.procTracks, time, model.end), 'proc', (elem, track) => {
			this.tip(elem, this.auraTip(track, 'us'));
			this.searchable(elem, track.actionId.name);
		});

		// The cast we started when we died in a leveling session never ends in the log, so the
		// corpse run has no cast bar.
		const dead = model.leveling?.stateAt(time) == 'dead';
		updateCastBar(this.playerCastBar, dead ? null : castAt(this.playerCasts, model.castStarts, time) ?? channelAt(model.channels, time), time);
		this.drawInstants(model, time);
	}

	// The instants pressed in the last moment, oldest first so the newest pops up on top.
	private drawInstants(model: ReplayModel, time: number) {
		const shown: Array<ReplayCast> = [];
		const last = lastIndexAtOrBefore(model.instantStarts, time);
		for (let i = last; i >= 0 && shown.length < INSTANT_POPS && time - model.instantStarts[i] < INSTANT_LIFE; i--) {
			shown.unshift(model.instants[i]);
		}
		this.instantPops.forEach((pop, i) => {
			const cast = shown[i];
			if (!cast) {
				pop.elem.style.opacity = '0';
				pop.shown = null;
				return;
			}
			if (pop.shown != cast) {
				pop.shown = cast;
				setIcon(pop.img, cast.log.actionId!);
			}
			const together = shown.filter(other => Math.abs(other.log.timestamp - cast.log.timestamp) < 0.001);
			const offset = (together.indexOf(cast) - (together.length - 1) / 2) * INSTANT_SPACING;
			const { scale, opacity } = instantPop(time - cast.log.timestamp);
			pop.elem.style.opacity = opacity.toFixed(3);
			pop.elem.style.transform = `translateX(${offset}px) scale(${scale.toFixed(3)})`;
		});
	}

	private drawEnemies(model: ReplayModel, time: number) {
		const focus = model.focus.at(time);
		// In a leveling session only the mobs we pulled and that are still alive have a frame.
		const pulled = model.leveling ? this.pulledEnemies(model, time) : null;
		if (pulled) this.ui.enemyCol.classList.toggle('many', pulled.size > 2);
		this.enemyFrames.forEach(frame => {
			const { enemy } = frame;
			if (pulled) {
				const shown = pulled.has(enemy.unit.index);
				frame.root.classList.toggle('hide', !shown);
				if (!shown) return;
			}
			const fraction = enemy.healthFraction(time);
			setWidth(frame.health, fraction);
			setText(frame.leftText, `${(fraction * 100).toFixed(fraction < 0.1 ? 1 : 0)}%`);
			frame.root.classList.toggle('focus', enemy.unit.index == focus && (pulled ? pulled.size : this.enemyFrames.length) > 1);
			syncAuras(frame.auras, frame.auraElems, activeAuras(enemy.debuffs, time, model.end), 'debuff', (elem, track) => {
				this.tip(elem, this.auraTip(track, enemy.name));
				this.searchable(elem, track.actionId.name);
			});
			updateCastBar(frame.castBar, castAt(enemy.casts, frame.starts, time) ?? channelAt(enemy.channels, time), time);
		});
	}

	private drawCooldowns(model: ReplayModel, time: number) {
		this.cooldownViews.forEach(({ cooldown, elem }) => {
			const aura = cooldown.aura ? activeAuraAt(cooldown.aura, time, model.end) : null;
			const cd = cooldownAt(model, cooldown, time);
			const pressIndex = lastIndexAtOrBefore(cooldown.castTimes, time);
			const pressed = pressIndex >= 0 && time - cooldown.castTimes[pressIndex] < PRESS_TIME;

			elem.classList.toggle('pressed', pressed);
			elem.classList.toggle('buff-up', !!aura);
			elem.classList.toggle('on-cooldown', !aura && !!cd);
			elem.classList.toggle('ready', !aura && !cd);
			elem.classList.toggle('expiring', !!aura && aura.remaining != null && aura.remaining < EXPIRING);
			setGlow(elem, aura ? BUFF_GLOW : null);
			if (aura) {
				updateIcon(elem, aura.fraction == null ? null : 1 - aura.fraction, aura.remaining == null ? '' : formatCountdown(aura.remaining), aura.stacks);
			} else if (cd) {
				updateIcon(elem, 1 - cd.remaining / cd.total, formatCountdown(cd.remaining), 0);
			} else {
				updateIcon(elem, null, '', 0);
			}
		});
	}

	// The maintain cells, nine to a row as wide as the resource bars. The shield is in cell 1, the
	// totems in cells 3 to 6 shifted half a cell right so the four are centred, and Flame Shock in
	// cell 9, like the user's Bayov Shaman pack.
	private buildMaintain(model: ReplayModel) {
		const pack = model.shaman;
		this.ui.maintain.replaceChildren();
		this.ui.maintain.classList.toggle('hide', !pack);
		if (!pack) {
			this.maintainViews = null;
			return;
		}
		const cell = (index: number, shift = 0): HTMLElement => {
			const elem = makeIcon(ActionId.fromEmpty(), 'cooldown maintain');
			elem.style.left = `calc(${index} * (var(--cell) + var(--gap)) + ${shift} * (var(--cell) + var(--gap)))`;
			this.ui.maintain.appendChild(elem);
			return elem;
		};
		const shield = pack.shields.length > 0 ? cell(0) : null;
		const totems = pack.totems.map((slot, i) => {
			const elem = cell(2 + i, 0.5);
			this.tip(elem, this.totemTip(slot));
			return { slot, elem, shown: '' };
		});
		const flameShock = model.enemies.some(enemy => enemy.debuffs.some(track => track.actionId.name.startsWith(pack.flameShock))) ? cell(8) : null;
		this.maintainViews = { shield, totems, flameShock, shieldShown: '', flameShockShown: '' };
	}

	private drawMaintain(model: ReplayModel, time: number) {
		const views = this.maintainViews;
		const pack = model.shaman;
		if (!views || !pack) return;
		const show = (elem: HTMLElement, actionId: ActionId, shownKey: string): string => {
			if (shownKey != actionId.toString()) {
				setIcon(elem.querySelector('.replay-icon-img')!, actionId);
				elem.dataset.name = actionId.name;
			}
			return actionId.toString();
		};

		// The shield. While it's down it stays in its cell, grey with a red glow.
		if (views.shield) {
			const up = pack.shields.map(track => activeAuraAt(track, time, model.end)).find(aura => !!aura) ?? null;
			const track = up?.track ?? pack.shields.find(track => track.spanStarts[0] <= time) ?? pack.shields[0];
			views.shieldShown = show(views.shield, track.actionId, views.shieldShown);
			views.shield.classList.toggle('missing', !up);
			setGlow(views.shield, up ? null : MISSING_GLOW);
			updateIcon(views.shield, null, '', up?.stacks ?? 0);
			if (!this.tips.has(views.shield)) {
				this.tip(views.shield, () => this.auraTip(pack.shields.find(track => activeAuraAt(track, this.time, model.end)) ?? pack.shields[0], 'us')());
			}
		}

		// The totems. A countdown turns orange, then red, near the totem's end. An empty slot
		// is grey with a red glow.
		views.totems.forEach(view => {
			const totem = totemAt(view.slot, time);
			view.elem.classList.toggle('hide', !totem);
			if (!totem) return;
			view.shown = show(view.elem, totem.drop.actionId, view.shown);
			view.elem.classList.toggle('missing', !totem.up);
			setGlow(view.elem, totem.up ? null : MISSING_GLOW);
			const left = totem.drop.until - time;
			view.elem.classList.toggle('expiring', totem.up && left < 10 && left >= 5);
			view.elem.classList.toggle('ending', totem.up && left < 5);
			updateIcon(
				view.elem,
				totem.up ? (time - totem.drop.start) / (totem.drop.until - totem.drop.start) : null,
				totem.up ? formatCountdown(left) : '',
				0,
			);
		});

		// Flame Shock on the enemy we're hitting. After it fades it stays in its cell, grey and
		// faded back.
		if (views.flameShock) {
			const enemy = model.enemies.find(enemy => enemy.unit.index == model.focus.at(time)) ?? model.enemies[0];
			const track = enemy.debuffs.find(track => track.actionId.name.startsWith(pack.flameShock));
			const aura = track ? activeAuraAt(track, time, model.end) : null;
			const applied = !!track && track.spanStarts[0] <= time;
			views.flameShock.classList.toggle('hide', !applied);
			if (track && applied) {
				views.flameShockShown = show(views.flameShock, track.actionId, views.flameShockShown);
				views.flameShock.classList.toggle('faded', !aura);
				updateIcon(
					views.flameShock,
					aura?.fraction == null ? null : 1 - aura.fraction,
					aura?.remaining == null ? '' : formatCountdown(aura.remaining),
					aura?.stacks ?? 0,
				);
				this.tip(views.flameShock, this.auraTip(track, enemy.name));
			}
		}
	}

	// The tooltip the game shows for the spell, buff or item, under ours. A buff shows the
	// buff's text, like hovering a buff in the game, and falls back to the spell's.
	private ingameTooltip(actionId: ActionId, isAura = false): Element | null {
		const id = actionId.spellTooltipOverride ?? actionId;
		const type = id.itemId ? 'item' : id.spellId ? 'spell' : null;
		if (!type) return null;
		const key = `${isAura && type == 'spell' ? 'aura' : type}/${id.itemId || id.spellId}`;
		let entry = ingameTooltips.get(key);
		if (!entry) {
			const url = `https://nether.wowhead.com/${WOWHEAD_DOMAIN}/tooltip/${type}/${id.itemId || id.spellId}?dataEnv=${WOWHEAD_EXPANSION_ENV}`;
			const newEntry: { content: DocumentFragment | null; loaded: Promise<void> } = { content: null, loaded: Promise.resolve() };
			newEntry.loaded = fetch(url)
				.then(response => response.json())
				.then(json => {
					const html = (key.startsWith('aura/') && json['buff']) || json['tooltip'];
					if (html && !json['error']) newEntry.content = parseTooltip(html);
				})
				.catch(() => undefined)
				.then(() => this.refreshTooltip());
			ingameTooltips.set(key, newEntry);
			entry = newEntry;
		}
		if (!entry.content) return null;
		const box = (<div className="replay-ingame-tooltip" />) as HTMLElement;
		box.appendChild(entry.content.cloneNode(true));
		return box;
	}

	private totemTip(slot: TotemSlot): () => Element | null {
		return () => {
			const totem = totemAt(slot, this.time);
			if (!totem) return null;
			const left = totem.drop.until - this.time;
			return (
				<div className="timeline-tooltip">
					{tooltipTitle(totem.drop.actionId)}
					<div className="timeline-tooltip-meta">
						{totem.up ? `${slot.element} totem, down since ${formatSeconds(totem.drop.start)}` : `No ${slot.element.toLowerCase()} totem down`}
					</div>
					{totem.up && <div className="timeline-tooltip-rows">{tooltipRow('Time left', formatSeconds(left))}</div>}
					{this.ingameTooltip(totem.drop.actionId)}
				</div>
			) as Element;
		};
	}

	// The cast history timeline: our casts where we made them, newest on the right at the
	// current time, sliding left as time goes on, like the TrufiGCD addon on a time axis.
	private drawHistory(model: ReplayModel, time: number) {
		const nowX = this.ui.history.clientWidth - HISTORY_NOW_PX;
		const xAt = (at: number) => nowX - (time - at) * HISTORY_PX_PER_SEC;
		const oldest = time - (nowX + HISTORY_ICON_PX) / HISTORY_PX_PER_SEC;

		const shown = new Set<number>();
		for (let index = lastIndexAtOrBefore(model.actionStarts, time); index >= 0; index--) {
			const at = model.actionStarts[index];
			if (at < oldest) break;
			shown.add(index);
			let elem = this.historyElems.get(index);
			if (!elem) {
				elem = this.makeHistoryIcon(model.actions[index]);
				elem.style.top = `${(HISTORY_ROWS - 1 - this.historyLayout[index].row) * HISTORY_ROW_PX}px`;
				this.historyElems.set(index, elem);
				this.ui.history.appendChild(elem);
			}
			const pop = easeOut(Math.min(1, (time - at) / HISTORY_SHIFT));
			const scale = 1.12 - 0.12 * pop;
			const x = xAt(at) + this.historyLayout[index].shift - HISTORY_ICON_PX / 2;
			elem.style.transform = `translateX(${x.toFixed(2)}px) scale(${scale.toFixed(3)})`;
			elem.style.opacity = pop.toFixed(3);
		}
		this.historyElems.forEach((elem, index) => {
			if (!shown.has(index)) {
				elem.remove();
				this.historyElems.delete(index);
			}
		});

		// The ticks are a repeating background that we slide with the time. The labels are the
		// fight's time every few seconds.
		const axis = this.ui.historyAxis;
		axis.style.setProperty('--offset', `${xAt(0).toFixed(2)}px`);
		const labels = new Set<number>();
		for (let at = Math.ceil(oldest / HISTORY_LABEL_EVERY) * HISTORY_LABEL_EVERY; at <= time; at += HISTORY_LABEL_EVERY) {
			labels.add(at);
			let elem = this.historyLabels.get(at);
			if (!elem) {
				elem = (<span className="replay-history-label">{formatClock(at, 0)}</span>) as HTMLElement;
				this.historyLabels.set(at, elem);
				axis.appendChild(elem);
			}
			elem.style.transform = `translateX(${xAt(at).toFixed(2)}px) translateX(-50%)`;
		}
		this.historyLabels.forEach((elem, at) => {
			if (!labels.has(at)) {
				elem.remove();
				this.historyLabels.delete(at);
			}
		});
	}

	private makeHistoryIcon(cast: ReplayCast): HTMLElement {
		const elem = makeIcon(cast.log.actionId!, 'history');
		this.tip(elem, () => this.castTip(cast.log));
		this.searchable(elem, cast.searchText);
		if (cast.badge) {
			elem.appendChild(<span className={`replay-icon-badge ${cast.badgeKind}`}>{cast.badge}</span>);
		}
		return elem;
	}

	// The scrolling combat text. What happens to us scrolls down on the left, what we do to the
	// enemies scrolls up on the right, and crits pop up and stay put for a moment.
	private drawSct(model: ReplayModel, time: number) {
		// Where the frames are, in pixels from the left and the right of the combat text's area.
		const zone = this.ui.sctZone.getBoundingClientRect();
		const playerFrame = this.playerFrame!.root.querySelector('.replay-unit-main')!.getBoundingClientRect();
		const focus = model.focus.at(time);
		const shownFrames = this.enemyFrames.filter(frame => !frame.root.classList.contains('hide'));
		const targetFrame = (shownFrames.find(frame => frame.enemy.unit.index == focus) ?? shownFrames[0])?.root
			.querySelector('.replay-unit-main')!
			.getBoundingClientRect();
		const ourRight = zone.right - playerFrame.right;
		// With the leveling map on a wide stage, the frames sit at its edges (_replay_map.scss). Without
		// a target, the target's frame would be where ours is, from the other side.
		const withMap = this.mapWide;
		const targetLeft = targetFrame ? targetFrame.left - zone.left : withMap ? zone.width - (playerFrame.right - zone.left) : zone.width * 0.7;

		// Where each side's lines start and end, in pixels from the top of the area.
		const history = this.ui.history.parentElement!.getBoundingClientRect();
		const inBottom = playerFrame.top - zone.top - SCT_MARGIN_PX;
		const inTop = Math.max(history.bottom - zone.top + SCT_MARGIN_PX, inBottom - MAX_SCROLL_PX);
		const outBottom = (targetFrame ? targetFrame.top - zone.top : zone.height) - SCT_MARGIN_PX;
		const outTop = Math.max(this.buffsElem.getBoundingClientRect().bottom - zone.top + SCT_MARGIN_PX, outBottom - MAX_SCROLL_PX);
		// Where the newer line on each side is. We place the lines from the newest back, so an older
		// line too close to a newer one gets pushed ahead of it.
		let newerIn = -Infinity;
		let newerOut = Infinity;

		// The crits' area has the same cap, so in full screen they stay near the scrolling lines.
		let critBottom = this.ui.procs.getBoundingClientRect().top - zone.top;
		let critTop = Math.max(history.bottom - zone.top, critBottom - MAX_SCROLL_PX);
		// The map takes the middle, so there the crits stay over the frames like the lines that
		// scroll, above where those scroll. When that room is too small, they share the lines' room.
		const mapCritBottom = Math.min(inTop, outTop) - SCT_MARGIN_PX;
		if (withMap && mapCritBottom - (history.bottom - zone.top) >= CRIT_SPOT_PX) {
			critTop = history.bottom - zone.top;
			critBottom = mapCritBottom;
		}
		const critSpan = critBottom - critTop;
		const spots = Math.max(1, Math.min(MAX_CRIT_SPOTS, Math.floor(critSpan / CRIT_SPOT_PX)));
		if (spots != this.critSpots) {
			placeCrits(model.sct, spots);
			this.critSpots = spots;
		}
		// The rows of the spots in the order the crits take them, from the middle out.
		const critRows = Array.from({ length: spots }, (_, row) => row).sort((a, b) => Math.abs(a - (spots - 1) / 2) - Math.abs(b - (spots - 1) / 2) || a - b);

		const first = lastIndexAtOrBefore(model.sctTimes, time - SCROLL_LIFE) + 1;
		const last = lastIndexAtOrBefore(model.sctTimes, time);
		const shown = new Set<number>();
		for (let i = last; i >= first; i--) {
			const event = model.sct[i];
			const age = time - event.time;
			if (age > event.end) continue;
			shown.add(i);
			let elem = this.sctElems.get(i);
			if (!elem) {
				elem = this.makeSctText(event);
				this.sctElems.set(i, elem);
				(event.side == 'in' ? this.ui.sctIn : this.ui.sctOut).appendChild(elem);
			}
			const progress = age / event.end;
			let y: number;
			let scale = 1;
			// A line pushed past the end of its area is gone.
			let pushedOut = false;
			if (event.sticky) {
				y = critTop + ((critRows[event.lane] ?? 0) + 0.5) * (critSpan / spots);
				scale = 1 + 0.55 * Math.max(0, 1 - age / 0.12);
			} else if (event.side == 'out') {
				y = newerOut = Math.min(outBottom - progress * (outBottom - outTop), newerOut - LANE_PX);
				pushedOut = y < outTop;
			} else {
				y = newerIn = Math.max(inTop + progress * (inBottom - inTop), newerIn + LANE_PX);
				pushedOut = y > inBottom;
			}
			const opacity = pushedOut ? 0 : Math.min(1, age / 0.06) * Math.min(1, (event.end - age) / (event.end - event.fadeAt));
			elem.style.top = `${y.toFixed(1)}px`;
			if (event.side == 'out' && (!event.sticky || withMap)) {
				elem.style.left = `${targetLeft}px`;
			} else if (event.side == 'out') {
				elem.style.right = `${zone.width - targetLeft + STICKY_GAP_PX}px`;
			} else if (!event.sticky || withMap) {
				elem.style.right = `${ourRight}px`;
			} else {
				elem.style.left = `${zone.width - ourRight + STICKY_GAP_PX}px`;
			}
			elem.style.transform = `translateY(-50%) scale(${scale.toFixed(3)})`;
			elem.style.opacity = opacity.toFixed(3);
		}
		this.sctElems.forEach((elem, index) => {
			if (!shown.has(index)) {
				elem.remove();
				this.sctElems.delete(index);
			}
		});
	}

	private makeSctText(event: SctEvent): HTMLElement {
		const elem = (
			<div className={`replay-sct-text ${event.kind}${event.sticky ? ' sticky' : ''}`}>
				{event.iconUrl ? <span className="replay-sct-icon" /> : null}
				<span className="replay-sct-amount">{event.text}</span>
				{event.note ? <span className="replay-sct-note">{`(${event.note})`}</span> : null}
			</div>
		) as HTMLElement;
		elem.style.color = event.color;
		this.tip(elem, () => this.sctTip(event));
		this.searchable(elem, event.searchText);
		const icon = elem.querySelector<HTMLElement>('.replay-sct-icon');
		if (icon) icon.style.backgroundImage = `url('${event.iconUrl}')`;
		return elem;
	}

	private drawControls(model: ReplayModel, time: number) {
		const { start, end } = this.playRange;
		const fraction = (time - start) / Math.max(end - start, 0.001);
		setWidth(this.ui.played, fraction);
		this.ui.thumb.style.left = `${(fraction * 100).toFixed(3)}%`;
		// A picked stretch of a session shows where it starts and ends in the session.
		const picked = end < model.end || start > model.start;
		setText(this.ui.clock, picked ? `${formatClock(time)} · ${formatClock(start)} to ${formatClock(end)}` : `${formatClock(time)} / ${formatClock(end)}`);
		this.updateButtons();
	}

	// The target indexes of the mobs out of the pool that we pulled and that are alive at this time.
	private pulledEnemies(model: ReplayModel, time: number): Set<number> {
		const pulled = new Set<number>();
		for (const mob of model.leveling!.mobs) {
			if (mob.foundAt > time) break;
			if (mob.pulledAt != null && mob.pulledAt <= time && time < mob.end) pulled.add(mob.targetIndex);
		}
		return pulled;
	}
}
