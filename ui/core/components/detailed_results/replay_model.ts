import { ResourceType } from '../../proto/api.js';
import { OtherAction, Stat } from '../../proto/common.js';
import { ActionId } from '../../proto_utils/action_id.js';
import { AuraUptimeLog, CastLog, DamageDealtLog, Entity, ResourceChangedLog, SimLog } from '../../proto_utils/logs_parser.js';
import { SimResult, SimResultFilter, UnitMetrics } from '../../proto_utils/sim_result.js';
import { formatToNumber } from '../../utils.js';
import { searchText } from './timeline.js';

// Everything the Replay view draws, worked out once per iteration. The view then asks for the
// state at any time of the iteration, so seeking back and forth draws the same frame every time.

// The index of the last time at or before the given one, or -1 when all of them are later.
export const lastIndexAtOrBefore = (times: ArrayLike<number>, time: number): number => {
	let lo = 0;
	let hi = times.length - 1;
	let found = -1;
	while (lo <= hi) {
		const mid = (lo + hi) >> 1;
		if (times[mid] <= time) {
			found = mid;
			lo = mid + 1;
		} else {
			hi = mid - 1;
		}
	}
	return found;
};

// A value that changes at points in time, like the mana bar. Before the first change it has its
// starting value.
export class Steps {
	private readonly times: Array<number> = [];
	private readonly values: Array<number> = [];

	constructor(readonly initial: number) {}

	push(time: number, value: number) {
		this.times.push(time);
		this.values.push(value);
	}

	at(time: number): number {
		const i = lastIndexAtOrBefore(this.times, time);
		return i < 0 ? this.initial : this.values[i];
	}

	get length(): number {
		return this.times.length;
	}
}

export interface ReplayBar {
	type: ResourceType;
	max: number;
	value: Steps;
}

// One stretch of an aura, from when we gained or refreshed it to when it faded or was refreshed.
export interface AuraSpan {
	actionId: ActionId;
	start: number;
	end: number;
	// The aura was refreshed at the end of the span, so it didn't fade then.
	refreshed: boolean;
	// When the aura first came up and when it finally faded, over all its refreshes.
	chainStart: number;
	chainEnd: number;
	stacks: Array<{ time: number; stacks: number }>;
	stackTimes: Array<number>;
}

// All the spans of one aura on one unit.
export interface AuraTrack {
	key: string;
	actionId: ActionId;
	spans: Array<AuraSpan>;
	spanStarts: Array<number>;
	// The longest the aura lasted when nothing cut it short, our best guess for its full
	// duration. Null when it never faded on its own.
	duration: number | null;
	// Up from the pull to the end of the iteration, like Blessing of Kings.
	alwaysUp: boolean;
}

export interface ActiveAura {
	track: AuraTrack;
	span: AuraSpan;
	// Seconds until it fades, or null when we can't tell.
	remaining: number | null;
	// How much of the aura's duration is left, from 1 to 0.
	fraction: number | null;
	stacks: number;
}

// An ability with a cooldown, for the cooldown tracker in the middle of the screen.
export interface TrackedCooldown {
	key: string;
	actionId: ActionId;
	timers: Array<string>;
	castTimes: Array<number>;
	longest: number;
	firstUse: number;
	// The player's buff from this ability, like the buff of Elemental Mastery. While it's up the
	// tracker shows it instead of the cooldown.
	aura: AuraTrack | null;
}

interface TimerUse {
	setAt: number;
	readyAt: number;
}

// Each time a cooldown timer was set, in order.
interface TimerUses {
	uses: Array<TimerUse>;
	setAts: Array<number>;
}

export type SctSide = 'in' | 'out';
export type SctKind = 'damage' | 'crit' | 'tick' | 'avoid' | 'heal' | 'aura' | 'resource' | 'hurt' | 'hurt-crit';

// One line of the scrolling combat text.
export interface SctEvent {
	time: number;
	side: SctSide;
	kind: SctKind;
	text: string;
	note: string;
	actionId: ActionId;
	iconUrl: string;
	// The log lines behind the line, like the hits of a Fire Nova, for its hover tooltip.
	logs: Array<SimLog>;
	// What the search looks in.
	searchText: string;
	// A CSS color, like the color of the spell school.
	color: string;
	// Lines that come close together start a bit apart, so they don't overlap.
	lane: number;
	// Crits pop up in the middle of the area and stay put, like MSBT's sticky crits.
	sticky: boolean;
}

export interface ReplayCast {
	log: CastLog;
	// What happened with its hits worth a badge, like CRIT or MISS.
	badge: string;
	badgeKind: 'crit' | 'avoid' | '';
	// What the search looks in, the cast and its hits.
	searchText: string;
}

export interface ReplayEnemy {
	unit: UnitMetrics;
	name: string;
	level: number;
	// The share of health left, from 1 to 0.
	healthFraction: (time: number) => number;
	damageTaken: Steps;
	debuffs: Array<AuraTrack>;
	casts: Array<CastLog>;
}

export interface ReplayModel {
	start: number;
	end: number;
	player: UnitMetrics;
	playerName: string;
	playerLevel: number;
	health: ReplayBar | null;
	powers: Array<ReplayBar>;
	comboPoints: Steps | null;
	playerAuras: Array<AuraTrack>;
	cooldowns: Array<TrackedCooldown>;
	timerUses: Map<string, TimerUses>;
	casts: Array<ReplayCast>;
	castStarts: Array<number>;
	// The casts of the action history strip, without the auto attacks and procs.
	actions: Array<ReplayCast>;
	actionStarts: Array<number>;
	enemies: Array<ReplayEnemy>;
	// The enemy the player hit last with a single target ability, by time.
	focus: Steps;
	damageDone: Steps;
	sct: Array<SctEvent>;
	sctTimes: Array<number>;
	markers: Array<{ time: number; actionId: ActionId }>;
	executeFraction: number;
	// The fight ends at a set time, so the enemy's health follows the time left.
	timeBasedHealth: boolean;
}

// The colors of the spell schools, by the school's bit in the sim's logs.
const schoolColors: Array<[number, string]> = [
	[1 << 6, '#d1fa99'], // Nature
	[1 << 3, '#ff9b5e'], // Fire
	[1 << 4, '#7fb2ff'], // Frost
	[1 << 7, '#c6b6ff'], // Shadow
	[1 << 5, '#ffff8f'], // Holy
	[1 << 2, '#8ff2ff'], // Arcane
];
const physicalColor = '#ffffff';

const schoolColor = (log: SimLog): string => {
	const match = log.raw.match(/\(SpellSchool: (\d+)\)/);
	const school = match ? parseInt(match[1]) : 0;
	return schoolColors.find(([bit]) => school & bit)?.[1] ?? physicalColor;
};

const hurtColor = '#ff4d4d';
const healColor = '#4dff6a';
const auraColor = '#ffd75e';
const avoidColor = '#d9d9d9';
const resourceColors: Partial<Record<ResourceType, string>> = {
	[ResourceType.ResourceTypeMana]: '#6fb3ff',
	[ResourceType.ResourceTypeEnergy]: '#ffe25e',
	[ResourceType.ResourceTypeRage]: '#ff6a5e',
	[ResourceType.ResourceTypeFocus]: '#ffb36a',
};
const resourceLabels: Partial<Record<ResourceType, string>> = {
	[ResourceType.ResourceTypeMana]: 'Mana',
	[ResourceType.ResourceTypeEnergy]: 'Energy',
	[ResourceType.ResourceTypeRage]: 'Rage',
	[ResourceType.ResourceTypeFocus]: 'Focus',
};

// Regen ticks happen many times a second under Forever. They would bury the other lines.
const regenActions = [OtherAction.OtherActionManaRegen, OtherAction.OtherActionEnergyRegen, OtherAction.OtherActionFocusRegen];
const autoAttacks = [OtherAction.OtherActionAttack, OtherAction.OtherActionShoot];

// Lines this close together on one side start apart from each other.
const LANE_WINDOW = 0.45;
const LANES = 6;
const STICKY_WINDOW = 0.7;
const STICKY_LANES = 3;
const EPSILON = 1e-6;
// The smallest mana gain the combat text shows, as a share of the mana bar.
const MIN_MANA_GAIN = 0.01;

const formatAmount = (amount: number) => formatToNumber(Math.round(amount), { maximumFractionDigits: 0 });

export const buildReplayModel = (result: SimResult, filter: SimResultFilter): ReplayModel | string => {
	const players = result.getRaidIndexedPlayers(filter);
	if (players.length != 1) {
		return 'The replay follows one player. Pick a player in the results filter above.';
	}
	const player = players[0];
	const targets = result.getTargets();
	const end = result.firstIterationDuration;
	const firstEvent = result.logs.reduce((first, log) => Math.min(first, log.timestamp), 0);
	const start = Math.floor(firstEvent);

	const playerProto = result.request.raid?.parties[Math.floor(player.index / 5)]?.players[player.index % 5];
	const encounter = result.request.encounter;

	const maxLog = player.logs.find(log => log.isMaxResources());
	const maxValues = maxLog?.isMaxResources() ? maxLog.maxValues : {};
	const resourceLogs = player.logs.filter((log): log is ResourceChangedLog => log.isResourceChanged());

	const barFor = (type: ResourceType): ReplayBar | null => {
		const logs = resourceLogs.filter(log => log.resourceType == type);
		const max = maxValues[type];
		if (max === undefined && logs.length == 0) return null;
		const observedMax = logs.reduce((most, log) => Math.max(most, log.valueBefore, log.valueAfter), 0);
		const value = new Steps(logs.length > 0 && logs[0].timestamp > start ? logs[0].valueBefore : max ?? observedMax);
		logs.forEach(log => value.push(log.timestamp, log.valueAfter));
		return { type, max: Math.max(max ?? 0, observedMax, 1), value };
	};

	const health = barFor(ResourceType.ResourceTypeHealth);
	const powers = [ResourceType.ResourceTypeEnergy, ResourceType.ResourceTypeRage, ResourceType.ResourceTypeFocus, ResourceType.ResourceTypeMana]
		.map(barFor)
		.filter((bar): bar is ReplayBar => !!bar);

	const comboLogs = resourceLogs.filter(log => log.resourceType == ResourceType.ResourceTypeComboPoints);
	let comboPoints: Steps | null = null;
	if (comboLogs.length > 0) {
		comboPoints = new Steps(0);
		comboLogs.forEach(log => comboPoints!.push(log.timestamp, log.valueAfter));
	}

	const playerAuras = buildAuraTracks(player.auraUptimeLogs, start, end);

	// The cooldowns, from the lines the sim writes when an ability goes on cooldown.
	const timerUseLists = new Map<string, Array<TimerUse>>();
	const cooldownsByKey = new Map<string, TrackedCooldown>();
	player.logs.forEach(log => {
		if (!log.isCooldown()) return;
		const actionId = log.actionId!;
		const key = actionId.toStringIgnoringTag();
		const timer = log.timer >= 0 ? `timer ${log.timer}` : `ability ${key}`;
		let cooldown = cooldownsByKey.get(key);
		if (!cooldown) {
			cooldown = { key, actionId, timers: [], castTimes: [], longest: 0, firstUse: log.timestamp, aura: null };
			cooldownsByKey.set(key, cooldown);
		}
		if (!cooldown.timers.includes(timer)) cooldown.timers.push(timer);
		cooldown.longest = Math.max(cooldown.longest, log.readyAt - log.timestamp);
		if (!timerUseLists.has(timer)) timerUseLists.set(timer, []);
		timerUseLists.get(timer)!.push({ setAt: log.timestamp, readyAt: log.readyAt });
	});
	const timerUses = new Map<string, TimerUses>();
	timerUseLists.forEach((uses, timer) => {
		uses.sort((a, b) => a.setAt - b.setAt);
		timerUses.set(timer, { uses, setAts: uses.map(use => use.setAt) });
	});

	const castLogs = player.castLogs.slice().sort((a, b) => a.timestamp - b.timestamp);
	const casts = castLogs.map(toReplayCast);
	const majorCooldownIds = player.majorCooldownLogs.map(log => log.actionId!);

	// Abilities with a cooldown shorter than the GCD, like an auto attack, aren't worth tracking.
	const cooldowns = Array.from(cooldownsByKey.values()).filter(cooldown => cooldown.longest >= 1.5);
	cooldowns.forEach(cooldown => {
		cooldown.castTimes = castLogs.filter(log => log.actionId!.toStringIgnoringTag() == cooldown.key).map(log => log.timestamp);
		cooldown.aura = playerAuras.find(track => track.actionId.equalsIgnoringTag(cooldown.actionId)) ?? null;
	});
	// The short rotation cooldowns first, then the long ones, each in the order we first used them.
	cooldowns.sort((a, b) => Number(a.longest > 30) - Number(b.longest > 30) || a.firstUse - b.firstUse);

	const trackedKeys = new Set(cooldowns.map(cooldown => cooldown.key));
	const actions = casts.filter(cast => {
		const actionId = cast.log.actionId!;
		if (autoAttacks.includes(actionId.otherId)) return false;
		// Procs, like a Windfury attack, are casts too. Only what we pressed takes the GCD or has
		// a cooldown of its own.
		return cast.log.effectiveTime > 0 || trackedKeys.has(actionId.toStringIgnoringTag()) || majorCooldownIds.some(id => id.equalsIgnoringTag(actionId));
	});

	// The enemies. The sim ends the fight at a set time unless the encounter uses health, and
	// then the execute phases follow the time left. So the health bars follow it too.
	const timeBasedHealth = !encounter?.useHealth;
	const playerDamage = (log: DamageDealtLog) => log.isDamage() && !!log.target?.isTarget;
	const allDamage = result.logs.filter((log): log is DamageDealtLog => log.isDamageDealt() && playerDamage(log));

	const enemies: Array<ReplayEnemy> = targets.map((target, i) => {
		const damageTaken = new Steps(0);
		let total = 0;
		allDamage
			.filter(log => log.target!.index == target.index)
			.forEach(log => {
				total += log.amount;
				damageTaken.push(log.timestamp, total);
			});
		const maxHealth = encounter?.targets[i]?.stats[Stat.StatHealth] || 0;
		const healthFraction = timeBasedHealth
			? (time: number) => clamp01((end - Math.max(0, time)) / end)
			: (time: number) => (maxHealth > 0 ? clamp01(1 - damageTaken.at(time) / maxHealth) : 1);
		return {
			unit: target,
			name: target.name || `Target ${i + 1}`,
			level: encounter?.targets[i]?.level || 63,
			healthFraction,
			damageTaken,
			debuffs: buildAuraTracks(target.auraUptimeLogs, start, end),
			casts: target.castLogs.filter(log => log.castTime > 0).sort((a, b) => a.timestamp - b.timestamp),
		};
	});

	// The focus is the enemy we hit last with an ability that hits one enemy.
	const focus = new Steps(enemies[0]?.unit.index ?? 0);
	castLogs.forEach(log => {
		const hitTargets = distinctTargets(log.damageDealtLogs);
		if (hitTargets.length == 1) focus.push(log.timestamp, hitTargets[0].index);
	});

	const ownDamage = [player, ...player.pets]
		.flatMap(unit => unit.damageDealtLogs)
		.filter(playerDamage)
		.sort((a, b) => a.timestamp - b.timestamp);
	const damageDone = new Steps(0);
	let totalDone = 0;
	ownDamage.forEach(log => {
		totalDone += log.amount;
		damageDone.push(log.timestamp, totalDone);
	});

	const sct = buildSct(player, targets, playerAuras, resourceLogs, powers);
	const markers = player.majorCooldownLogs.map(log => ({ time: log.timestamp, actionId: log.actionId! }));

	return {
		start,
		end,
		player,
		playerName: player.name.trim() || 'Player',
		playerLevel: playerProto?.level || 60,
		health,
		powers,
		comboPoints,
		playerAuras,
		cooldowns,
		timerUses,
		casts,
		castStarts: casts.map(cast => cast.log.timestamp),
		actions,
		actionStarts: actions.map(cast => cast.log.timestamp),
		enemies,
		focus,
		damageDone,
		sct,
		sctTimes: sct.map(event => event.time),
		markers,
		executeFraction: encounter?.executeProportion20 ?? 0,
		timeBasedHealth,
	};
};

const clamp01 = (value: number) => Math.max(0, Math.min(1, value));

const distinctTargets = (logs: Array<SimLog>): Array<Entity> => {
	const found: Array<Entity> = [];
	logs.forEach(log => {
		if (log.target?.isTarget && !found.some(target => target.equals(log.target!))) found.push(log.target);
	});
	return found;
};

const toReplayCast = (log: CastLog): ReplayCast => {
	const hits = log.damageDealtLogs;
	const text = searchText([log, ...hits]);
	if (hits.some(hit => hit.crit)) return { log, badge: 'CRIT', badgeKind: 'crit', searchText: text };
	const avoided = hits.find(hit => hit.miss || hit.dodge || hit.parry);
	if (avoided && hits.every(hit => hit.miss || hit.dodge || hit.parry)) {
		return { log, badge: avoided.miss ? 'MISS' : avoided.dodge ? 'DODGE' : 'PARRY', badgeKind: 'avoid', searchText: text };
	}
	return { log, badge: '', badgeKind: '', searchText: text };
};

const buildAuraTracks = (uptimeLogs: Array<AuraUptimeLog>, start: number, end: number): Array<AuraTrack> => {
	const byKey = new Map<string, Array<AuraUptimeLog>>();
	uptimeLogs.forEach(log => {
		const key = log.actionId!.toString();
		if (!byKey.has(key)) byKey.set(key, []);
		byKey.get(key)!.push(log);
	});

	return Array.from(byKey.entries()).map(([key, logs]) => {
		logs.sort((a, b) => a.gainedAt - b.gainedAt);
		const spans: Array<AuraSpan> = logs.map((log, i) => ({
			actionId: log.actionId!,
			start: log.gainedAt,
			end: log.fadedAt,
			refreshed: i + 1 < logs.length && Math.abs(logs[i + 1].gainedAt - log.fadedAt) < EPSILON,
			chainStart: log.gainedAt,
			chainEnd: log.fadedAt,
			stacks: log.stacksChange.map(change => ({ time: change.timestamp, stacks: change.newStacks })),
			stackTimes: log.stacksChange.map(change => change.timestamp),
		}));
		for (let i = 1; i < spans.length; i++) {
			if (spans[i - 1].refreshed) spans[i].chainStart = spans[i - 1].chainStart;
		}
		for (let i = spans.length - 2; i >= 0; i--) {
			if (spans[i].refreshed) spans[i].chainEnd = spans[i + 1].chainEnd;
		}

		let duration: number | null = null;
		spans.forEach(span => {
			if (!span.refreshed && span.end < end - EPSILON) duration = Math.max(duration ?? 0, span.end - span.start);
		});

		const alwaysUp = spans.some(span => span.chainStart <= Math.max(0, start) + 0.01 && span.chainEnd >= end - 0.01);
		return { key, actionId: logs[0].actionId!, spans, spanStarts: spans.map(span => span.start), duration, alwaysUp };
	});
};

// The aura's state at this time, or null when it's down.
export const activeAuraAt = (track: AuraTrack, time: number, end: number): ActiveAura | null => {
	const i = lastIndexAtOrBefore(track.spanStarts, time);
	if (i < 0) return null;
	const span = track.spans[i];
	if (time >= span.end) return null;

	let remaining: number | null = null;
	if (!track.alwaysUp) {
		if (!span.refreshed && span.end < end - EPSILON) {
			remaining = span.end - time;
		} else if (track.duration != null) {
			remaining = Math.max(0, span.start + track.duration - time);
		}
	}
	const full = track.duration ?? span.end - span.start;
	const stackIndex = lastIndexAtOrBefore(span.stackTimes, time);
	return {
		track,
		span,
		remaining,
		fraction: remaining != null && full > 0 ? clamp01(remaining / full) : null,
		stacks: stackIndex >= 0 ? span.stacks[stackIndex].stacks : 0,
	};
};

// The cooldown left on the ability at this time, with the length of that cooldown.
export const cooldownAt = (model: ReplayModel, cooldown: TrackedCooldown, time: number): { remaining: number; total: number } | null => {
	let best: { remaining: number; total: number } | null = null;
	cooldown.timers.forEach(timer => {
		const { uses, setAts } = model.timerUses.get(timer)!;
		const i = lastIndexAtOrBefore(setAts, time);
		if (i < 0) return;
		const use = uses[i];
		if (use.readyAt > time && (!best || use.readyAt - time > best.remaining)) {
			best = { remaining: use.readyAt - time, total: use.readyAt - use.setAt };
		}
	});
	return best;
};

const buildSct = (
	player: UnitMetrics,
	targets: Array<UnitMetrics>,
	playerAuras: Array<AuraTrack>,
	resourceLogs: Array<ResourceChangedLog>,
	bars: Array<ReplayBar>,
): Array<SctEvent> => {
	const events: Array<Omit<SctEvent, 'lane' | 'searchText'>> = [];
	const isPlayer = (entity: Entity | null) => !!entity && entity.isPlayer && entity.index == player.index;

	// What we did to the enemies. Hits of one ability at the same moment, like Fire Nova on three
	// enemies, show as one line, like MSBT merges them.
	const outgoing = [player, ...player.pets].flatMap(unit => unit.damageDealtLogs).filter(log => log.target?.isTarget);
	const groups = new Map<string, Array<DamageDealtLog>>();
	outgoing.forEach(log => {
		const key = `${log.timestamp.toFixed(2)} ${log.actionId!.toString()} ${log.isHealing()}`;
		if (!groups.has(key)) groups.set(key, []);
		groups.get(key)!.push(log);
	});
	groups.forEach(logs => {
		const first = logs[0];
		const landed = logs.filter(log => !log.miss && !log.dodge && !log.parry);
		const iconUrl = first.actionId!.iconUrl;
		if (landed.length == 0) {
			events.push({
				time: first.timestamp,
				side: 'out',
				kind: 'avoid',
				text: first.miss ? 'Miss' : first.dodge ? 'Dodge' : 'Parry',
				note: logs.length > 1 ? `×${logs.length}` : '',
				actionId: first.actionId!,
				iconUrl,
				logs,
				color: avoidColor,
				sticky: false,
			});
			return;
		}
		const amount = landed.reduce((total, log) => total + log.amount, 0);
		const crit = landed.some(log => log.crit);
		const notes: Array<string> = [];
		if (logs.length > 1) notes.push(`×${logs.length}`);
		if (landed.some(log => log.glance)) notes.push('glancing');
		if (landed.some(log => log.block || log.blockedCrit)) notes.push('blocked');
		const resist = landed.find(log => log.partialResist1_4 || log.partialResist2_4 || log.partialResist3_4);
		if (resist) notes.push(`${resist.partialResist1_4 ? 25 : resist.partialResist2_4 ? 50 : 75}% resisted`);
		events.push({
			time: first.timestamp,
			side: 'out',
			kind: crit ? 'crit' : first.tick ? 'tick' : 'damage',
			text: crit ? `${formatAmount(amount)}!` : formatAmount(amount),
			note: notes.join(', '),
			actionId: first.actionId!,
			iconUrl,
			logs,
			color: schoolColor(first),
			sticky: crit,
		});
	});

	// What happened to us: hits from the enemies, heals and shields, buffs and resources.
	const incoming = targets.flatMap(target => target.damageDealtLogs).filter(log => isPlayer(log.target));
	incoming.forEach(log => {
		const avoided = log.miss || log.dodge || log.parry;
		events.push({
			time: log.timestamp,
			side: 'in',
			kind: avoided ? 'avoid' : log.crit || log.crush ? 'hurt-crit' : 'hurt',
			text: avoided ? (log.miss ? 'Miss' : log.dodge ? 'Dodge' : 'Parry') : `-${formatAmount(log.amount)}`,
			note: log.crush ? 'crushing' : log.block ? 'blocked' : '',
			actionId: log.actionId!,
			iconUrl: log.actionId!.iconUrl,
			logs: [log],
			color: avoided ? avoidColor : hurtColor,
			sticky: !avoided && (log.crit || log.crush),
		});
	});

	[player, ...player.pets]
		.flatMap(unit => unit.damageDealtLogs)
		.filter(log => isPlayer(log.target) && (log.isHealing() || log.isShielding()) && log.amount > 0)
		.forEach(log => {
			events.push({
				time: log.timestamp,
				side: 'in',
				kind: 'heal',
				text: `+${formatAmount(log.amount)}`,
				note: log.isShielding() ? 'absorb' : log.crit ? 'crit' : '',
				actionId: log.actionId!,
				iconUrl: log.actionId!.iconUrl,
				logs: [log],
				color: healColor,
				sticky: false,
			});
		});

	playerAuras
		.filter(track => !track.alwaysUp)
		.forEach(track => {
			track.spans
				.filter(span => span.start == span.chainStart && span.start >= 0)
				.forEach(span => {
					events.push({
						time: span.start,
						side: 'in',
						kind: 'aura',
						text: `+${track.actionId.name}`,
						note: '',
						actionId: track.actionId,
						iconUrl: track.actionId.iconUrl,
						logs: [],
						color: auraColor,
						sticky: false,
					});
				});
		});

	resourceLogs.forEach(log => {
		if (log.isSpend || log.timestamp < 0) return;
		const label = resourceLabels[log.resourceType];
		const amount = log.valueAfter - log.valueBefore;
		if (!label || amount < 1 || regenActions.includes(log.actionId!.otherId)) return;
		// Small mana gains that come all the time, like Blessing of Wisdom, would bury the rest.
		const max = bars.find(bar => bar.type == log.resourceType)?.max ?? 0;
		if (log.resourceType == ResourceType.ResourceTypeMana && amount < max * MIN_MANA_GAIN) return;
		events.push({
			time: log.timestamp,
			side: 'in',
			kind: 'resource',
			text: `+${formatAmount(amount)} ${label}`,
			note: '',
			actionId: log.actionId!,
			iconUrl: log.actionId!.iconUrl,
			logs: [log],
			color: resourceColors[log.resourceType]!,
			sticky: false,
		});
	});

	events.sort((a, b) => a.time - b.time);

	// Lines close together start apart, so a burst of hits doesn't print on top of itself.
	const recent: Record<string, Array<number>> = {};
	return events.map(event => {
		const group = `${event.side} ${event.sticky}`;
		const window = event.sticky ? STICKY_WINDOW : LANE_WINDOW;
		const times = (recent[group] = (recent[group] ?? []).filter(time => event.time - time < window));
		const lane = times.length % (event.sticky ? STICKY_LANES : LANES);
		times.push(event.time);
		return { ...event, lane, searchText: `${searchText(event.logs)} ${event.actionId.name} ${event.text} ${event.note}`.toLowerCase() };
	});
};
