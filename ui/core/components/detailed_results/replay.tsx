import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { ResourceType } from '../../proto/api.js';
import { ActionId } from '../../proto_utils/action_id.js';
import { CastLog } from '../../proto_utils/logs_parser.js';
import { formatToCompactNumber, formatToNumber } from '../../utils.js';
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
} from './replay_model.js';
import { ResultComponent, ResultComponentConfig, SimResultData } from './result_component.js';

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
const HISTORY_SLOT_PX = 48;
// A cooldown icon lights up like a pressed button for this long after we use it.
const PRESS_TIME = 0.25;
// The controls fade out after the mouse rests this long while we play.
const IDLE_MS = 2500;

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
	healthText: HTMLElement;
	healthPercent: HTMLElement;
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
const syncAuras = (host: HTMLElement, pool: Map<string, HTMLElement>, auras: Array<ActiveAura>, className = '') => {
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
			pool.set(aura.track.key, elem);
		}
		if (host.children[i] !== elem) host.insertBefore(elem, host.children[i] ?? null);
		updateIcon(elem, aura.fraction == null ? null : 1 - aura.fraction, aura.remaining == null ? '' : formatCountdown(aura.remaining), aura.stacks);
	});
};

const activeAuras = (tracks: Array<AuraTrack>, time: number, end: number): Array<ActiveAura> =>
	tracks
		.map(track => activeAuraAt(track, time, end))
		.filter((aura): aura is ActiveAura => !!aura)
		.sort((a, b) => a.span.chainStart - b.span.chainStart);

export interface ReplayConfig extends ResultComponentConfig {}

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
	private idleTimer = 0;
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
		gcd: HTMLElement;
		gcdFill: HTMLElement;
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
	private playerFrame: (UnitFrameRefs & { powers: Array<{ bar: ReplayBar; fill: HTMLElement; text: HTMLElement }>; combo: Array<HTMLElement> }) | null = null;
	private playerCastBar = makeCastBar('player');
	private playerCasts: Array<CastLog> = [];
	private enemyFrames: Array<UnitFrameRefs & { enemy: ReplayEnemy; damage: HTMLElement; starts: Array<number> }> = [];
	private cooldownElems: Array<HTMLElement> = [];
	private procElems = new Map<string, HTMLElement>();
	// The buffs that come and go, like procs. A buff of an ability in the cooldown tracker shows
	// on the ability's button instead.
	private procTracks: Array<AuraTrack> = [];
	private historyElems = new Map<number, HTMLElement>();
	private sctElems = new Map<number, HTMLElement>();

	constructor(config: ReplayConfig) {
		config.rootCssClass = 'replay-root';
		super(config);

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
			gcd: ref<HTMLDivElement>(),
			gcdFill: ref<HTMLDivElement>(),
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
						<div className="replay-center">
							<div ref={refs.history} className="replay-history" />
							<div ref={refs.procs} className="replay-procs" />
							<div ref={refs.coreCooldowns} className="replay-cooldowns core" />
							<div ref={refs.majorCooldowns} className="replay-cooldowns major" />
							<div ref={refs.gcd} className="replay-gcd">
								<div ref={refs.gcdFill} className="replay-gcd-fill" />
							</div>
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
			gcd: refs.gcd.value!,
			gcdFill: refs.gcdFill.value!,
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
		this.ui.player.addEventListener('pointermove', () => this.wake());
		// The keys work while the replay has focus. A click on a button would move the focus to
		// the button, and then Space would press the button as well as play or pause.
		this.ui.player.addEventListener('pointerup', () => this.rootElem.focus({ preventScroll: true }));
		this.ui.player.addEventListener('pointerleave', () => this.ui.player.classList.toggle('idle', this.playing));
		document.addEventListener('click', () => this.ui.speedMenu.classList.add('hide'));

		this.bindProgress();
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
		this.wake();
		this.updateButtons();
	}

	private pause() {
		this.playing = false;
		cancelAnimationFrame(this.frameRequest);
		this.ui?.player.classList.remove('idle');
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

	// The controls fade out while we play and the mouse rests, like on a video.
	private wake() {
		this.ui.player.classList.remove('idle');
		window.clearTimeout(this.idleTimer);
		if (this.playing) {
			this.idleTimer = window.setTimeout(() => this.ui.player.classList.toggle('idle', this.playing), IDLE_MS);
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
		this.wake();
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
			tippy(elem, { content: cooldown.actionId.name });
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

	private buildUnitFrame(name: string, level: number, iconUrl: string, side: 'player' | 'enemy'): UnitFrameRefs & { extra: HTMLElement } {
		const castBar = makeCastBar(side);
		const root = (
			<div className={`replay-unit ${side}`}>
				<div className="replay-unit-auras" />
				<div className="replay-unit-main">
					<div className="replay-unit-portrait" />
					<div className="replay-unit-body">
						<div className="replay-unit-head">
							<span className="replay-unit-name">{name}</span>
							<span className="replay-unit-level">{String(level)}</span>
						</div>
						<div className="replay-bar health">
							<div className="replay-bar-fill" />
							<span className="replay-bar-text" />
							<span className="replay-bar-percent" />
						</div>
						<div className="replay-unit-extra" />
					</div>
				</div>
			</div>
		) as HTMLElement;
		(root.querySelector('.replay-unit-portrait') as HTMLElement).style.backgroundImage = iconUrl ? `url('${iconUrl}')` : '';
		root.appendChild(castBar.root);
		return {
			root,
			health: root.querySelector('.health .replay-bar-fill')!,
			healthText: root.querySelector('.health .replay-bar-text')!,
			healthPercent: root.querySelector('.health .replay-bar-percent')!,
			castBar,
			auras: root.querySelector('.replay-unit-auras')!,
			auraElems: new Map(),
			extra: root.querySelector('.replay-unit-extra')!,
		};
	}

	private buildPlayerFrame(model: ReplayModel) {
		const frame = this.buildUnitFrame(model.playerName, model.playerLevel, model.player.iconUrl, 'player');
		const powers = model.powers.map(bar => {
			const elem = (
				<div className={`replay-bar power ${powerClass[bar.type] ?? ''}`}>
					<div className="replay-bar-fill" />
					<span className="replay-bar-text" />
				</div>
			) as HTMLElement;
			frame.extra.appendChild(elem);
			return { bar, fill: elem.querySelector<HTMLElement>('.replay-bar-fill')!, text: elem.querySelector<HTMLElement>('.replay-bar-text')! };
		});
		const combo: Array<HTMLElement> = [];
		if (model.comboPoints) {
			const pips = (<div className="replay-combo" />) as HTMLElement;
			for (let i = 0; i < 5; i++) {
				const pip = (<div className="replay-combo-pip" />) as HTMLElement;
				pips.appendChild(pip);
				combo.push(pip);
			}
			frame.extra.appendChild(pips);
		}

		// The buffs that are up the whole fight never change, so we draw them once, small.
		model.playerAuras
			.filter(track => track.alwaysUp)
			.forEach(track => {
				const elem = makeIcon(track.actionId, 'buff');
				frame.auras.appendChild(elem);
				tippy(elem, { content: track.actionId.name });
			});

		this.ui.playerCol.replaceChildren(frame.root);
		this.playerFrame = { ...frame, powers, combo };
	}

	private buildEnemyFrames(model: ReplayModel) {
		this.enemyFrames = model.enemies.map(enemy => {
			const frame = this.buildUnitFrame(enemy.name, enemy.level, enemy.unit.iconUrl, 'enemy');
			const damage = (<span className="replay-unit-damage" />) as HTMLElement;
			frame.extra.appendChild(damage);
			if (model.executeFraction > 0) {
				const mark = (<div className="replay-bar-execute" />) as HTMLElement;
				// The enemy's bar empties toward the right, so the mark is measured from the right.
				mark.style.right = `${model.executeFraction * 100}%`;
				frame.root.querySelector('.health')!.appendChild(mark);
			}
			if (model.timeBasedHealth) {
				tippy(frame.root.querySelector('.health') as HTMLElement, {
					content: 'The fight ends at a set time, so the health bar follows the time left, like the execute phases of the sim do.',
				});
			}
			return { ...frame, enemy, damage, starts: enemy.casts.map(cast => cast.timestamp) };
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
			setText(frame.healthText, `${formatToNumber(Math.round(health))} / ${formatToNumber(model.health.max)}`);
			setText(frame.healthPercent, `${Math.round((health / model.health.max) * 100)}%`);
		} else {
			setWidth(frame.health, 1);
		}
		frame.powers.forEach(({ bar, fill, text }) => {
			const value = bar.value.at(time);
			setWidth(fill, value / bar.max);
			setText(
				text,
				bar.type == ResourceType.ResourceTypeMana ? `${formatToNumber(Math.floor(value))} / ${formatToNumber(bar.max)}` : String(Math.floor(value)),
			);
		});
		if (model.comboPoints) {
			const points = model.comboPoints.at(time);
			frame.combo.forEach((pip, i) => pip.classList.toggle('on', i < points));
		}

		// The procs and other buffs that come and go, in the middle above the cooldowns.
		syncAuras(this.ui.procs, this.procElems, activeAuras(this.procTracks, time, model.end), 'proc');

		updateCastBar(this.playerCastBar, castAt(this.playerCasts, model.castStarts, time), time);

		// The GCD of the last thing we pressed, as a thin line under the cooldowns.
		const last = lastIndexAtOrBefore(model.actionStarts, time);
		let gcd = 0;
		for (let i = last; i >= 0 && i > last - 4; i--) {
			const log = model.actions[i].log;
			if (log.castTime == 0 && log.effectiveTime > 0 && time < log.timestamp + log.effectiveTime) {
				gcd = 1 - (time - log.timestamp) / log.effectiveTime;
				break;
			}
		}
		this.ui.gcd.classList.toggle('active', gcd > 0);
		setWidth(this.ui.gcdFill, gcd);
	}

	private drawEnemies(model: ReplayModel, time: number) {
		const focus = model.focus.at(time);
		this.enemyFrames.forEach(frame => {
			const { enemy } = frame;
			const fraction = enemy.healthFraction(time);
			setWidth(frame.health, fraction);
			setText(frame.healthPercent, `${(fraction * 100).toFixed(fraction < 0.1 ? 1 : 0)}%`);
			setText(frame.damage, `${formatToCompactNumber(enemy.damageTaken.at(time), { maximumFractionDigits: 1 })} damage taken`);
			frame.root.classList.toggle('focus', enemy.unit.index == focus && this.enemyFrames.length > 1);
			frame.root.classList.toggle('execute', model.executeFraction > 0 && fraction <= model.executeFraction);
			syncAuras(frame.auras, frame.auraElems, activeAuras(enemy.debuffs, time, model.end), 'debuff');
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
