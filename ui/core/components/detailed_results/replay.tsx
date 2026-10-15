// @ts-expect-error
import debounce from 'lodash/debounce';
import tippy, { Instance } from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { ResourceType } from '../../proto/api.js';
import { ActionId } from '../../proto_utils/action_id.js';
import { CastLog, DamageDealtLog } from '../../proto_utils/logs_parser.js';
import { resourceNames } from '../../proto_utils/names.js';
import { formatToCompactNumber, formatToNumber } from '../../utils.js';
import { EventFilter } from './events_filter.js';
import {
	ActiveAura,
	activeAuraAt,
	AuraTrack,
	buildReplayModel,
	cooldownAt,
	lastIndexAtOrBefore,
	ReplayBar,
	ReplayCast,
	ReplayEnemy,
	ReplayModel,
	SctEvent,
	TrackedCooldown,
} from './replay_model.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';
import { tooltipHit, tooltipRow, tooltipTitle } from './timeline.js';

const SPEEDS = [0.25, 0.5, 0.75, 1, 1.25, 1.5, 2, 3, 4];
const SPEED_STORAGE_KEY = '__detailed_results_replay_speed__';

// How long a line of the combat text stays up, in seconds of the iteration. At 2x it scrolls
// by twice as fast, like everything else.
const SCROLL_LIFE = 2.2;
const STICKY_LIFE = 1.6;
// How far apart lines that start close together are, in pixels.
const LANE_PX = 28;
const STICKY_LANE_PX = 42;
// Where the lines sit across their area, as a share of its width. Crits stay put in their own
// column toward the middle of the screen, so the lines that scroll by don't run over them.
const SCROLL_X: Record<SctEvent['side'], number> = { in: 0.36, out: 0.64 };
const STICKY_X: Record<SctEvent['side'], number> = { in: 0.8, out: 0.2 };

// The action history strip shows this many of our latest casts. A new cast pushes the others
// one slot to the left over this many seconds.
const HISTORY_SLOTS = 6;
const HISTORY_SHIFT = 0.15;
const HISTORY_SLOT_PX = 50;
// A cooldown icon lights up like a pressed button for this long after we use it.
const PRESS_TIME = 0.25;
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
	const minutes = Math.floor(abs / 60);
	const seconds = abs - minutes * 60;
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

const formatSeconds = (seconds: number): string => `${seconds.toFixed(1)}s`;

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
	icon: HTMLElement;
	name: HTMLElement;
	time: HTMLElement;
	shown: CastLog | null;
}

const makeCastBar = (className: string): CastBarRefs => {
	const root = (
		<div className={`replay-castbar ${className}`}>
			<div className="replay-castbar-icon" />
			<div className="replay-castbar-bar">
				<div className="replay-castbar-fill" />
				<span className="replay-castbar-name" />
				<span className="replay-castbar-time" />
			</div>
		</div>
	) as HTMLElement;
	return {
		root,
		fill: root.querySelector('.replay-castbar-fill')!,
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

const updateCastBar = (bar: CastBarRefs, cast: CastLog | null, time: number) => {
	bar.root.classList.toggle('active', !!cast);
	if (!cast) return;
	if (bar.shown != cast) {
		bar.shown = cast;
		setIcon(bar.icon, cast.actionId!);
		setText(bar.name, cast.actionId!.name);
	}
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

	private readonly ui: {
		player: HTMLElement;
		stage: HTMLElement;
		message: HTMLElement;
		flash: HTMLElement;
		hudTime: HTMLElement;
		hudDps: HTMLElement;
		hudDamage: HTMLElement;
		playerCol: HTMLElement;
		enemyCol: HTMLElement;
		history: HTMLElement;
		procs: HTMLElement;
		coreCooldowns: HTMLElement;
		majorCooldowns: HTMLElement;
		center: HTMLElement;
		resources: HTMLElement;
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
		swingFill: HTMLElement;
		pips: Array<HTMLElement>;
	} | null = null;
	private playerCastBar = makeCastBar('player');
	private playerCasts: Array<CastLog> = [];
	private enemyFrames: Array<UnitFrameRefs & { enemy: ReplayEnemy; starts: Array<number> }> = [];
	private cooldownElems: Array<HTMLElement> = [];
	private procElems = new Map<string, HTMLElement>();
	// The buffs that come and go, like procs. A buff of an ability in the cooldown tracker shows
	// on the ability's button instead.
	private procTracks: Array<AuraTrack> = [];
	private historyElems = new Map<number, HTMLElement>();
	private sctElems = new Map<number, HTMLElement>();

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
			if (SPEEDS.includes(stored)) this.speed = stored;
		} catch {
			// The speed is a convenience, the replay works without it.
		}

		const refs = {
			player: ref<HTMLDivElement>(),
			stage: ref<HTMLDivElement>(),
			message: ref<HTMLDivElement>(),
			flash: ref<HTMLDivElement>(),
			hudTime: ref<HTMLSpanElement>(),
			hudDps: ref<HTMLSpanElement>(),
			hudDamage: ref<HTMLSpanElement>(),
			playerCol: ref<HTMLDivElement>(),
			enemyCol: ref<HTMLDivElement>(),
			history: ref<HTMLDivElement>(),
			procs: ref<HTMLDivElement>(),
			coreCooldowns: ref<HTMLDivElement>(),
			majorCooldowns: ref<HTMLDivElement>(),
			center: ref<HTMLDivElement>(),
			resources: ref<HTMLDivElement>(),
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
						<span className="replay-hud-stat">
							<span className="replay-hud-label">Damage</span>
							<span ref={refs.hudDamage} className="replay-hud-value" />
						</span>
					</div>
					<div className="replay-sct-zone">
						<div ref={refs.sctIn} className="replay-sct replay-sct-in" />
						<div ref={refs.sctOut} className="replay-sct replay-sct-out" />
					</div>
					<div className="replay-frames">
						<div ref={refs.playerCol} className="replay-col replay-col-player" />
						<div ref={refs.center} className="replay-center">
							<div ref={refs.history} className="replay-history" />
							<div ref={refs.procs} className="replay-procs" />
							<div ref={refs.resources} className="replay-resources" />
							<div ref={refs.coreCooldowns} className="replay-cooldowns core" />
							<div ref={refs.majorCooldowns} className="replay-cooldowns major" />
							{this.playerCastBar.root}
						</div>
						<div ref={refs.enemyCol} className="replay-col replay-col-enemies" />
					</div>
					<div ref={refs.message} className="replay-message hide" />
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
			hudTime: refs.hudTime.value!,
			hudDps: refs.hudDps.value!,
			hudDamage: refs.hudDamage.value!,
			playerCol: refs.playerCol.value!,
			enemyCol: refs.enemyCol.value!,
			history: refs.history.value!,
			procs: refs.procs.value!,
			coreCooldowns: refs.coreCooldowns.value!,
			majorCooldowns: refs.majorCooldowns.value!,
			center: refs.center.value!,
			resources: refs.resources.value!,
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

		SPEEDS.forEach(speed => {
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
		this.ui.speedButton.addEventListener('click', event => {
			event.stopPropagation();
			this.ui.speedMenu.classList.toggle('hide');
		});
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
		// the button, and then Space would press the button as well as play or pause.
		this.ui.player.addEventListener('pointerup', () => this.rootElem.focus({ preventScroll: true }));
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
		this.ui.message.classList.add('hide');
		this.ui.player.classList.remove('empty');
		this.model = model;
		this.time = model.start;
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

	private play() {
		if (!this.model || this.playing) return;
		if (this.time >= this.model.end) this.time = this.model.start;
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
		this.time = Math.min(this.model.end, this.time + elapsed * this.speed);
		this.draw();
		if (this.time >= this.model.end) {
			this.pause();
			return;
		}
		this.frameRequest = requestAnimationFrame(next => this.tick(next));
	}

	private seek(time: number) {
		if (!this.model) return;
		this.time = Math.max(this.model.start, Math.min(this.model.end, time));
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
		const i = SPEEDS.indexOf(this.speed);
		const next = SPEEDS[Math.max(0, Math.min(SPEEDS.length - 1, i + direction))];
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
		const span = model.end - model.start;
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
				this.seek(model.start);
				break;
			case 'End':
				this.seek(model.end);
				break;
			default:
				if (/^[0-9]$/.test(event.key)) {
					this.seek(model.start + (span * parseInt(event.key)) / 10);
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
			return this.model ? this.model.start + fraction * (this.model.end - this.model.start) : 0;
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
		const span = model.end - model.start;
		const slots = new Set(this.matchTimes.map(time => Math.round(((time - model.start) / span) * 400)));
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
			</div>
		) as Element;
	}

	// A line of the combat text: the hits behind it, or the buff or resource we gained.
	private sctTip(event: SctEvent): Element {
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
			</div>
		) as Element;
	}

	private updateButtons() {
		const ended = !!this.model && this.time >= this.model.end;
		const icon = this.playing ? 'fa-pause' : ended ? 'fa-rotate-right' : 'fa-play';
		this.ui.playButton.querySelector('i')!.className = `fas ${icon}`;
		this.ui.player.classList.toggle('paused', !this.playing);
		// Like a video before we start it or after it ends, a big button in the middle asks us to
		// play.
		const atStart = !!this.model && this.time <= this.model.start;
		this.ui.player.classList.toggle('show-big-play', !this.playing && (atStart || ended));
		this.ui.player.querySelector('.replay-big-play i')!.className = `fas ${ended ? 'fa-rotate-right' : 'fa-play'}`;
	}

	// Builds the frames, the cooldown tracker and the progress bar markers for a new iteration.
	private build(model: ReplayModel) {
		this.sctElems.forEach(elem => elem.remove());
		this.sctElems.clear();
		this.historyElems.forEach(elem => elem.remove());
		this.historyElems.clear();
		this.procElems.forEach(elem => elem.remove());
		this.procElems.clear();

		this.playerCasts = model.casts.map(cast => cast.log);
		const trackedAuras = new Set(model.cooldowns.map(cooldown => cooldown.aura?.key));
		this.procTracks = model.playerAuras.filter(track => !track.alwaysUp && !trackedAuras.has(track.key));
		this.buildPlayerFrame(model);
		this.buildEnemyFrames(model);

		// The cooldown tracker.
		this.ui.coreCooldowns.replaceChildren();
		this.ui.majorCooldowns.replaceChildren();
		this.cooldownElems = model.cooldowns.map(cooldown => {
			const elem = makeIcon(cooldown.actionId, 'cooldown');
			(cooldown.longest > 30 ? this.ui.majorCooldowns : this.ui.coreCooldowns).appendChild(elem);
			this.tip(elem, this.cooldownTip(cooldown));
			this.searchable(elem, cooldown.actionId.name);
			return elem;
		});

		// The progress bar marks where we used a major cooldown, like chapters on a video.
		const span = model.end - model.start;
		this.ui.markers.replaceChildren(
			...model.markers.map(marker => {
				const elem = (<div className="replay-progress-marker" />) as HTMLElement;
				elem.style.left = `${((marker.time - model.start) / span) * 100}%`;
				return elem;
			}),
		);
		this.ui.player.style.setProperty('--class-color', `var(--bs-${model.player.classColor})`);
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
		const swing = (
			<div className="replay-swing">
				<div className="replay-swing-fill" />
			</div>
		) as HTMLElement;
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
		this.ui.resources.replaceChildren(
			...powers.map(power => power.elem),
			...(model.mainHandSwings.length > 1 ? [swing] : []),
			...(pipKind ? [pipRow] : []),
		);
		this.tip(this.ui.resources, () => {
			const time = this.time;
			return (
				<div className="timeline-tooltip">
					<div className="timeline-tooltip-rows">
						{model.powers.map(bar =>
							tooltipRow(resourceNames.get(bar.type)!, `${formatToNumber(Math.floor(bar.value.at(time)))} / ${formatToNumber(bar.max)}`),
						)}
						{model.comboPoints && tooltipRow('Combo points', String(model.comboPoints.at(time)))}
						{model.maelstrom && tooltipRow('Maelstrom Weapon', String(this.pipsAt(model, time)))}
						{this.swingAt(model, time) != null && tooltipRow('Next main hand swing', formatSeconds(this.swingAt(model, time)!.next - time))}
					</div>
				</div>
			) as Element;
		});
		this.resourceRefs = { powers, swingFill: swing.firstElementChild as HTMLElement, pips };
	}

	private pipsAt(model: ReplayModel, time: number): number {
		if (model.comboPoints) return model.comboPoints.at(time);
		if (!model.maelstrom) return 0;
		const aura = activeAuraAt(model.maelstrom, time, model.end);
		return aura ? Math.max(1, aura.stacks) : 0;
	}

	// The last and the next main hand swing around this time. A long gap, like while we were out
	// of range, has no swing timer.
	private swingAt(model: ReplayModel, time: number): { last: number; next: number } | null {
		const swings = model.mainHandSwings;
		const i = lastIndexAtOrBefore(swings, time);
		if (i < 0 || i + 1 >= swings.length || swings[i + 1] - swings[i] > 5) return null;
		return { last: swings[i], next: swings[i + 1] };
	}

	// A cast bar's tooltip is the cast's, while the bar shows one.
	private tipCastBar(bar: CastBarRefs) {
		this.tip(bar.root, () => (bar.root.classList.contains('active') && bar.shown ? this.castTip(bar.shown) : null));
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
			if (model.executeFraction > 0) {
				const mark = (<div className="replay-bar-execute" />) as HTMLElement;
				// The enemy's bar empties toward the right, so the mark is measured from the right.
				mark.style.right = `${model.executeFraction * 100}%`;
				frame.root.querySelector('.health')!.appendChild(mark);
			}
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

	// Draws the screen at the current time.
	private draw() {
		const model = this.model;
		if (!model || !this.isVisible()) return;
		const time = this.time;

		this.drawHud(model, time);
		this.drawPlayer(model, time);
		this.drawEnemies(model, time);
		this.drawCooldowns(model, time);
		this.drawHistory(model, time);
		this.drawSct(model, time);
		this.drawControls(model, time);
		this.updateHover();
	}

	private drawHud(model: ReplayModel, time: number) {
		setText(this.ui.hudTime, formatClock(time));
		const damage = model.damageDone.at(time);
		setText(this.ui.hudDps, time >= 1 ? formatToNumber(damage / time, { maximumFractionDigits: 1, minimumFractionDigits: 1 }) : '–');
		setText(this.ui.hudDamage, formatToCompactNumber(damage, { maximumFractionDigits: 1 }));
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
		const swing = this.swingAt(model, time);
		setWidth(resources.swingFill, swing ? (time - swing.last) / (swing.next - swing.last) : 0);
		const pips = this.pipsAt(model, time);
		resources.pips.forEach((pip, i) => pip.classList.toggle('on', i < pips));
		// At five Maelstrom Weapon stacks the pack lights up around the middle, like the
		// PowerAuras textures of the user's WeakAura.
		this.ui.center.classList.toggle('maelstrom-full', !!model.maelstrom && pips >= 5);

		// The procs and other buffs that come and go, in the middle above the cooldowns.
		syncAuras(this.ui.procs, this.procElems, activeAuras(this.procTracks, time, model.end), 'proc', (elem, track) => {
			this.tip(elem, this.auraTip(track, 'us'));
			this.searchable(elem, track.actionId.name);
		});

		updateCastBar(this.playerCastBar, castAt(this.playerCasts, model.castStarts, time), time);
	}

	private drawEnemies(model: ReplayModel, time: number) {
		const focus = model.focus.at(time);
		this.enemyFrames.forEach(frame => {
			const { enemy } = frame;
			const fraction = enemy.healthFraction(time);
			setWidth(frame.health, fraction);
			setText(frame.leftText, `${(fraction * 100).toFixed(fraction < 0.1 ? 1 : 0)}%`);
			frame.root.classList.toggle('focus', enemy.unit.index == focus && this.enemyFrames.length > 1);
			frame.root.classList.toggle('execute', model.executeFraction > 0 && fraction <= model.executeFraction);
			syncAuras(frame.auras, frame.auraElems, activeAuras(enemy.debuffs, time, model.end), 'debuff', (elem, track) => {
				this.tip(elem, this.auraTip(track, enemy.name));
				this.searchable(elem, track.actionId.name);
			});
			updateCastBar(frame.castBar, castAt(enemy.casts, frame.starts, time), time);
		});
	}

	private drawCooldowns(model: ReplayModel, time: number) {
		model.cooldowns.forEach((cooldown, i) => {
			const elem = this.cooldownElems[i];
			const aura = cooldown.aura ? activeAuraAt(cooldown.aura, time, model.end) : null;
			const cd = cooldownAt(model, cooldown, time);
			const pressIndex = lastIndexAtOrBefore(cooldown.castTimes, time);
			const pressed = pressIndex >= 0 && time - cooldown.castTimes[pressIndex] < PRESS_TIME;

			elem.classList.toggle('pressed', pressed);
			elem.classList.toggle('buff-up', !!aura);
			elem.classList.toggle('on-cooldown', !aura && !!cd);
			elem.classList.toggle('ready', !aura && !cd);
			elem.classList.toggle('expiring', !!aura && aura.remaining != null && aura.remaining < EXPIRING);
			if (aura) {
				updateIcon(elem, aura.fraction == null ? null : 1 - aura.fraction, aura.remaining == null ? '' : formatCountdown(aura.remaining), aura.stacks);
			} else if (cd) {
				updateIcon(elem, 1 - cd.remaining / cd.total, formatCountdown(cd.remaining), 0);
			} else {
				updateIcon(elem, null, '', 0);
			}
		});
	}

	// The action history strip: our latest casts, newest on the right, sliding left as we cast
	// more, like the TrufiGCD addon.
	private drawHistory(model: ReplayModel, time: number) {
		const last = lastIndexAtOrBefore(model.actionStarts, time);
		const shown = new Set<number>();
		if (last >= 0) {
			const newest = model.actions[last].log.timestamp;
			const shift = easeOut(Math.min(1, (time - newest) / HISTORY_SHIFT));
			for (let k = 0; k < HISTORY_SLOTS + 1 && last - k >= 0; k++) {
				const index = last - k;
				const position = k - 1 + shift;
				if (position > HISTORY_SLOTS - 1) break;
				shown.add(index);
				let elem = this.historyElems.get(index);
				if (!elem) {
					elem = this.makeHistoryIcon(model.actions[index]);
					this.historyElems.set(index, elem);
					this.ui.history.appendChild(elem);
				}
				const x = -(position + 1) * HISTORY_SLOT_PX;
				const fadeOut = Math.max(0, position - (HISTORY_SLOTS - 2));
				const opacity = (k == 0 ? shift : 1) * (1 - fadeOut) * (1 - 0.1 * Math.max(0, position));
				const scale = k == 0 ? 1.12 - 0.12 * (1 - shift) : 1;
				elem.style.transform = `translateX(${x.toFixed(2)}px) scale(${scale.toFixed(3)})`;
				elem.style.opacity = opacity.toFixed(3);
				elem.classList.toggle('newest', k == 0);
			}
		}
		this.historyElems.forEach((elem, index) => {
			if (!shown.has(index)) {
				elem.remove();
				this.historyElems.delete(index);
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
		const first = lastIndexAtOrBefore(model.sctTimes, time - SCROLL_LIFE) + 1;
		const last = lastIndexAtOrBefore(model.sctTimes, time);
		const shown = new Set<number>();
		for (let i = first; i <= last; i++) {
			const event = model.sct[i];
			const age = time - event.time;
			const life = event.sticky ? STICKY_LIFE : SCROLL_LIFE;
			if (age > life) continue;
			shown.add(i);
			let elem = this.sctElems.get(i);
			if (!elem) {
				elem = this.makeSctText(event);
				this.sctElems.set(i, elem);
				(event.side == 'in' ? this.ui.sctIn : this.ui.sctOut).appendChild(elem);
			}
			const progress = age / life;
			// Where the line is, as a share of the area's height and a number of pixels from there.
			// A line that starts close after another starts a few lanes behind it.
			let y: number;
			let offset: number;
			let scale = 1;
			if (event.sticky) {
				y = 0.45;
				offset = (event.side == 'in' ? 1 : -1) * event.lane * STICKY_LANE_PX;
				scale = 1 + 0.55 * Math.max(0, 1 - age / 0.12);
			} else if (event.side == 'out') {
				y = 1 - progress;
				offset = event.lane * LANE_PX;
			} else {
				y = progress;
				offset = -event.lane * LANE_PX;
			}
			const x = (event.lane % 2 ? 1 : -1) * event.lane * 10;
			const opacity = Math.min(1, age / 0.06) * Math.min(1, (1 - progress) / 0.3);
			elem.style.top = `calc(${(y * 100).toFixed(2)}% + ${offset}px)`;
			elem.style.left = `${(event.sticky ? STICKY_X : SCROLL_X)[event.side] * 100}%`;
			elem.style.transform = `translate(calc(-50% + ${x}px), -50%) scale(${scale.toFixed(3)})`;
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
		const fraction = (time - model.start) / (model.end - model.start);
		setWidth(this.ui.played, fraction);
		this.ui.thumb.style.left = `${(fraction * 100).toFixed(3)}%`;
		setText(this.ui.clock, `${formatClock(time)} / ${formatClock(model.end)}`);
		this.updateButtons();
	}
}
