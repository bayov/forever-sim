import { ResourceType } from '../../proto/api.js';
import { OtherAction, Stat } from '../../proto/common.js';
import { ActionId, resourceTypeToIcon } from '../../proto_utils/action_id.js';
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
	// When the aura was set to run out at the start of the span, from the log. Null in results
	// simmed before the sim wrote it, where we guess from the spans instead.
	expiresAt: number | null;
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
	// A crit's spot in the middle of the area. The scrolling lines don't use it.
	lane: number;
	// Crits pop up in the middle of the area and stay put, like MSBT's sticky crits.
	sticky: boolean;
	// How long after its time the line starts to fade, and when it's gone, in seconds.
	fadeAt: number;
	end: number;
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

// A totem in one of the shaman's four slots, from when it went down until it runs out or the
// next totem of its element takes its place.
export interface TotemDrop {
	actionId: ActionId;
	start: number;
	until: number;
}

export interface TotemSlot {
	element: string;
	drops: Array<TotemDrop>;
	starts: Array<number>;
}

// The shaman class pack, laid out like the user's Bayov WeakAura: the bars of cooldowns in its
// order, and the maintain cells over the resource bars (the shield, the four totems and Flame
// Shock on the target).
export interface ShamanPack {
	shields: Array<AuraTrack>;
	totems: Array<TotemSlot>;
	flameShock: string;
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
	// The two rows under the resource bars: the main bar and the utility row.
	mainBar: Array<TrackedCooldown>;
	utilityBar: Array<TrackedCooldown>;
	shaman: ShamanPack | null;
	timerUses: Map<string, TimerUses>;
	casts: Array<ReplayCast>;
	castStarts: Array<number>;
	// The casts of the action history strip, without the auto attacks and procs.
	actions: Array<ReplayCast>;
	actionStarts: Array<number>;
	enemies: Array<ReplayEnemy>;
	// The enemy the player hit last with a single target ability, by time.
	focus: Steps;
	// When each hand swung, for the swing timers. A Heroic Strike, Cleave or Maul takes the main
	// hand's swing, so it counts as one.
	mainHandSwings: Array<number>;
	offHandSwings: Array<number>;
	// The auras of a queued Heroic Strike, Cleave, Maul or Raptor Strike, while it waits for the
	// next main hand swing.
	swingQueue: Array<AuraTrack>;
	// The Maelstrom Weapon buff, which the shaman pack shows as five pips like combo points.
	maelstrom: AuraTrack | null;
	damageDone: Steps;
	sct: Array<SctEvent>;
	sctTimes: Array<number>;
	markers: Array<{ time: number; actionId: ActionId }>;
	// Results simmed before the sim wrote the cooldown lines have no cooldown tracker.
	hasCooldownData: boolean;
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

// How long a line of the combat text stays up, in seconds of the iteration. At 2x it scrolls
// by twice as fast, like everything else. A line fades out over the last part of its life.
export const SCROLL_LIFE = 2.2;
export const STICKY_LIFE = 1.6;
const FADE_SHARE = 0.3;
const EPSILON = 1e-6;
// The smallest mana gain the combat text shows, as a share of the mana bar.
const MIN_MANA_GAIN = 0.01;
// Small gains of rage or energy, like the +1 Rage of each Unbridled Wrath proc, come many times a
// second. Gains under this share of the bar that come within the window of the first one go on
// one line, like MSBT's merged lines. The line shows at the last gain, so it never tells of a
// gain before it happens.
const SMALL_GAIN = 0.05;
const MERGE_WINDOW = 2;

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

	// The sim keeps an aura on us while a totem of ours is down, like Searing Totem. In game it's
	// only a totem, so we show it in its totem cell and not as a buff too.
	const totemSlots = buildTotemSlots(player.logs);
	const totemIds = new Set(totemSlots.flatMap(slot => slot.drops.map(drop => drop.actionId.toString())));
	const totemNames = new Set(totemSlots.flatMap(slot => slot.drops.map(drop => baseName(drop.actionId)).filter(Boolean)));
	const playerAuras = buildAuraTracks(player.auraUptimeLogs, start, end).filter(
		track => !totemNames.has(baseName(track.actionId)) && !totemIds.has(track.actionId.toString()),
	);

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
	const isShaman = player.classColor == 'shaman';
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
	// The shaman pack's bars always hold the spells we cast, like the WeakAura's icons that show
	// once a spell is learned. A spell we cast without a cooldown line (like Chain Lightning in a
	// result simmed before the sim wrote them) shows ready and lights up when we press it.
	if (isShaman) {
		[...SHAMAN_MAIN, ...SHAMAN_UTILITY].forEach(name => {
			if (cooldowns.some(cooldown => named(cooldown.actionId, name))) return;
			const cast = castLogs.find(log => named(log.actionId!, name));
			if (!cast) return;
			const key = cast.actionId!.toStringIgnoringTag();
			cooldowns.push({ key, actionId: cast.actionId!, timers: [], castTimes: [], longest: 0, firstUse: cast.timestamp, aura: null });
		});
	}
	// The same goes for the major cooldowns of every class (racials, trinkets, engineering
	// explosives, potions), so they keep their button when the result has no cooldown lines.
	majorCooldownIds.forEach(actionId => {
		const key = actionId.toStringIgnoringTag();
		if (cooldowns.some(cooldown => cooldown.key == key)) return;
		const cast = castLogs.find(log => log.actionId!.toStringIgnoringTag() == key);
		if (!cast) return;
		cooldowns.push({ key, actionId: cast.actionId!, timers: [], castTimes: [], longest: 0, firstUse: cast.timestamp, aura: null });
	});
	cooldowns.forEach(cooldown => {
		cooldown.castTimes = castLogs.filter(log => log.actionId!.toStringIgnoringTag() == cooldown.key).map(log => log.timestamp);
		cooldown.aura = playerAuras.find(track => track.actionId.equalsIgnoringTag(cooldown.actionId)) ?? null;
	});
	cooldowns.sort((a, b) => a.firstUse - b.firstUse);

	const { mainBar, utilityBar } = isShaman ? shamanBars(cooldowns) : splitBars(cooldowns);
	const shaman: ShamanPack | null = isShaman
		? {
				shields: playerAuras.filter(track => SHAMAN_SHIELDS.some(name => named(track.actionId, name))),
				totems: totemSlots,
				flameShock: 'Flame Shock',
		  }
		: null;

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
			// The sim keeps Searing Totem and Magma Totem going as a dot on the first enemy, but in game
			// it's a totem that casts at the enemy, and the enemy has no debuff for it.
			debuffs: buildAuraTracks(
				target.auraUptimeLogs.filter(log => !totemIds.has(log.actionId!.toString())),
				start,
				end,
			),
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
		mainBar,
		utilityBar,
		shaman,
		timerUses,
		casts,
		castStarts: casts.map(cast => cast.log.timestamp),
		actions,
		actionStarts: actions.map(cast => cast.log.timestamp),
		enemies,
		focus,
		mainHandSwings: castLogs.filter(log => isAutoAttack(log.actionId!, 1) || isNextSwingHit(log.actionId!)).map(log => log.timestamp),
		offHandSwings: castLogs.filter(log => isAutoAttack(log.actionId!, 2)).map(log => log.timestamp),
		swingQueue: playerAuras.filter(track => isSwingQueue(track.actionId)),
		maelstrom: playerAuras.find(track => track.actionId.name.startsWith('Maelstrom Weapon')) ?? null,
		damageDone,
		sct,
		sctTimes: sct.map(event => event.time),
		markers,
		hasCooldownData: player.logs.some(log => log.isCooldown() || log.isMaxResources()),
		timeBasedHealth,
	};
};

const clamp01 = (value: number) => Math.max(0, Math.min(1, value));

// The rows of the user's Bayov Shaman pack (bayovui/auras/forever/shaman/group.py), in its
// order. The main bar is the rotation and the big cooldowns, then the racial and the items we
// carry. Utility is the situational cooldowns. Flame Shock isn't on a bar, it has its own
// maintain cell.
const SHAMAN_MAIN = ['Stormstrike', 'Earth Shock', 'Fire Nova', 'Chain Lightning', 'Lava Burst', 'Rage of the Farseer', "Nature's Swiftness"];
const SHAMAN_UTILITY = ['Grounding Totem', 'Earthbind Totem', 'Stoneclaw Totem', 'Mana Tide Totem', 'Totemic Projection'];
const SHAMAN_UTILITY_ITEMS = ['Healthstone', 'Bandage', 'Target Dummy'];
const SHAMAN_MAINTAIN = ['Flame Shock'];
const SHAMAN_SHIELDS = ['Lightning Shield', 'Water Shield'];
// The totem slots in the order of the pack's cells.
const TOTEM_ELEMENTS = ['Air', 'Earth', 'Fire', 'Water'];

// The ability's name without its rank, like "Earth Shock" for "Earth Shock (Rank 7)".
const named = (actionId: ActionId, name: string) => actionId.name == name || actionId.name.startsWith(`${name} (`);
// The name without its rank or tag, like "Searing Totem" for "Searing Totem (Rank 5)".
const baseName = (actionId: ActionId) => actionId.name.replace(/ \(.*$/, '');

const isAutoAttack = (actionId: ActionId, tag: number) => actionId.otherId == OtherAction.OtherActionAttack && actionId.tag == tag;
// The abilities that take the next main hand swing. The sim tags the queue spell and its aura
// 1, and the hit 0. Raptor Strike's hit is tagged 1 and its queue 3.
const NEXT_SWING = ['Heroic Strike', 'Cleave', 'Maul'];
const isNextSwingHit = (actionId: ActionId) =>
	(actionId.tag == 0 && NEXT_SWING.some(name => named(actionId, name))) || (actionId.tag == 1 && named(actionId, 'Raptor Strike'));
const isSwingQueue = (actionId: ActionId) =>
	(actionId.tag == 1 && NEXT_SWING.some(name => named(actionId, name))) || (actionId.tag == 3 && named(actionId, 'Raptor Strike'));

const shamanBars = (cooldowns: Array<TrackedCooldown>) => {
	const indexIn = (list: Array<string>, cooldown: TrackedCooldown) => list.findIndex(name => named(cooldown.actionId, name));
	const rest = cooldowns.filter(cooldown => indexIn([...SHAMAN_MAIN, ...SHAMAN_UTILITY, ...SHAMAN_MAINTAIN], cooldown) < 0);
	const utilityItems = rest.filter(cooldown => SHAMAN_UTILITY_ITEMS.some(name => cooldown.actionId.name.includes(name)));
	const mainBar = [
		...cooldowns.filter(cooldown => indexIn(SHAMAN_MAIN, cooldown) >= 0).sort((a, b) => indexIn(SHAMAN_MAIN, a) - indexIn(SHAMAN_MAIN, b)),
		// The racial, the explosive, the potion and the rune, and anything else we used.
		...rest.filter(cooldown => !utilityItems.includes(cooldown)).sort((a, b) => a.firstUse - b.firstUse),
	];
	const utilityBar = [
		...cooldowns.filter(cooldown => indexIn(SHAMAN_UTILITY, cooldown) >= 0).sort((a, b) => indexIn(SHAMAN_UTILITY, a) - indexIn(SHAMAN_UTILITY, b)),
		...utilityItems,
	];
	return { mainBar, utilityBar };
};

// The other classes have no pack of the user's to follow, so we do what the shaman pack does.
// The main bar holds what we press for damage: the rotation, the big cooldowns, the racial and
// the items we use (trinkets, explosives, potions, runes). The row under it holds the situational
// cooldowns, like defensives, crowd control and movement.
//
// A row wraps after 9 icons. The main bar has our spells first, then the items, each in the
// order we first used them.
const OTHER_UTILITY = [
	'Shield Wall',
	'Last Stand',
	'Shield Block',
	'Berserker Rage',
	'Intimidating Shout',
	'Pummel',
	'Shield Bash',
	'Charge',
	'Intercept',
	'Intervene',
	'Evasion',
	'Sprint',
	'Kick',
	'Gouge',
	'Blind',
	'Kidney Shot',
	'Distract',
	'Divine Shield',
	'Divine Protection',
	'Blessing of Protection',
	'Blessing of Freedom',
	'Lay on Hands',
	'Hammer of Justice',
	'Ice Block',
	'Blink',
	'Counterspell',
	'Frost Nova',
	'Psychic Scream',
	'Fade',
	'Power Word: Shield',
	'Silence',
	'Feign Death',
	'Freezing Trap',
	'Frost Trap',
	'Scatter Shot',
	'Disengage',
	'Wing Clip',
	'Intimidation',
	'Barkskin',
	'Dash',
	'Bash',
	"Nature's Grasp",
	'Frenzied Regeneration',
	'Howl of Terror',
	'Spell Lock',
	'Shadow Ward',
	'War Stomp',
	'Will of the Forsaken',
	'Stoneform',
	'Escape Artist',
	'Perception',
];

const splitBars = (cooldowns: Array<TrackedCooldown>) => {
	const isUtility = (cooldown: TrackedCooldown) =>
		OTHER_UTILITY.some(name => named(cooldown.actionId, name)) || SHAMAN_UTILITY_ITEMS.some(name => cooldown.actionId.name.includes(name));
	const main = cooldowns.filter(cooldown => !isUtility(cooldown));
	return {
		mainBar: [...main.filter(cooldown => !cooldown.actionId.itemId), ...main.filter(cooldown => cooldown.actionId.itemId)],
		utilityBar: cooldowns.filter(isUtility),
	};
};

const buildTotemSlots = (logs: Array<SimLog>): Array<TotemSlot> => {
	const slots = new Map<string, Array<TotemDrop>>();
	logs.forEach(log => {
		if (!log.isTotem()) return;
		if (!slots.has(log.element)) slots.set(log.element, []);
		const drops = slots.get(log.element)!;
		// A totem we put down before the pull logs its shorter time left right after.
		const last = drops[drops.length - 1];
		if (last && last.start == log.timestamp && last.actionId.equals(log.actionId!)) {
			last.until = log.until;
		} else {
			drops.push({ actionId: log.actionId!, start: log.timestamp, until: log.until });
		}
	});
	return TOTEM_ELEMENTS.filter(element => slots.has(element)).map(element => {
		const drops = slots.get(element)!;
		return { element, drops, starts: drops.map(drop => drop.start) };
	});
};

// The totem down in the slot at this time, or the last one when the slot is empty.
export const totemAt = (slot: TotemSlot, time: number): { drop: TotemDrop; up: boolean } | null => {
	const i = lastIndexAtOrBefore(slot.starts, time);
	if (i < 0) return null;
	const drop = slot.drops[i];
	return { drop, up: time < drop.until };
};

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
			expiresAt: log.expiresAt,
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
	let full = track.duration ?? span.end - span.start;
	if (!track.alwaysUp) {
		if (span.expiresAt != null) {
			remaining = Math.max(0, span.expiresAt - time);
			full = span.expiresAt - span.start;
		} else if (!span.refreshed && span.end < end - EPSILON) {
			remaining = span.end - time;
		} else if (track.duration != null) {
			remaining = Math.max(0, span.start + track.duration - time);
		}
	}
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
	const events: Array<Omit<SctEvent, 'lane' | 'searchText' | 'fadeAt' | 'end'>> = [];
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

	const gained = (log: ResourceChangedLog) => log.valueAfter - log.valueBefore;
	const resourceLine = (logs: Array<ResourceChangedLog>) => {
		const sources = new Set(logs.map(log => log.actionId!.toStringIgnoringTag()));
		const type = logs[0].resourceType;
		events.push({
			time: logs[logs.length - 1].timestamp,
			side: 'in',
			kind: 'resource',
			text: `+${formatAmount(logs.reduce((sum, log) => sum + gained(log), 0))} ${resourceLabels[type]}`,
			note: logs.length > 1 ? `×${logs.length}` : '',
			actionId: logs[0].actionId!,
			// Gains from several sources show the resource's icon.
			iconUrl: sources.size > 1 ? resourceTypeToIcon[type] : logs[0].actionId!.iconUrl,
			logs,
			color: resourceColors[type]!,
			sticky: false,
		});
	};
	const merging = new Map<ResourceType, Array<ResourceChangedLog>>();
	resourceLogs.forEach(log => {
		if (log.isSpend || log.timestamp < 0) return;
		const amount = gained(log);
		if (!resourceLabels[log.resourceType] || amount < 1 || regenActions.includes(log.actionId!.otherId)) return;
		const max = bars.find(bar => bar.type == log.resourceType)?.max ?? 0;
		// Small mana gains that come all the time, like Blessing of Wisdom, would bury the rest.
		if (log.resourceType == ResourceType.ResourceTypeMana) {
			if (amount >= max * MIN_MANA_GAIN) resourceLine([log]);
			return;
		}
		if (amount >= max * SMALL_GAIN) {
			resourceLine([log]);
			return;
		}
		const batch = merging.get(log.resourceType);
		if (batch && log.timestamp - batch[0].timestamp <= MERGE_WINDOW) {
			batch.push(log);
		} else {
			if (batch) resourceLine(batch);
			merging.set(log.resourceType, [log]);
		}
	});
	merging.forEach(resourceLine);

	events.sort((a, b) => a.time - b.time);

	const sct = events.map(event => {
		const life = event.sticky ? STICKY_LIFE : SCROLL_LIFE;
		const search = `${searchText(event.logs)} ${event.actionId.name} ${event.text} ${event.note}`.toLowerCase();
		return { ...event, lane: 0, fadeAt: life * (1 - FADE_SHARE), end: life, searchText: search };
	});
	placeCrits(sct, 3);
	return sct;
};

// Gives each crit one of the spots on its side. The replay calls this again when its height fits
// a different number of spots.
//
// A crit takes a spot that no other crit is on, also one that is still fading. When every spot
// is taken, it covers the oldest crit, and that one starts to fade right away.
export const placeCrits = (sct: Array<SctEvent>, spotCount: number) => {
	const spots: Record<SctSide, Array<SctEvent | null>> = {
		in: new Array(spotCount).fill(null),
		out: new Array(spotCount).fill(null),
	};
	sct.forEach(event => {
		if (!event.sticky) return;
		event.fadeAt = STICKY_LIFE * (1 - FADE_SHARE);
		event.end = STICKY_LIFE;
		const side = spots[event.side];
		let lane = side.findIndex(other => !other || other.time + other.end <= event.time);
		if (lane < 0) {
			lane = side.reduce((oldest, other, i) => (other!.time < side[oldest]!.time ? i : oldest), 0);
			const covered = side[lane]!;
			const age = event.time - covered.time;
			if (age < covered.fadeAt) {
				covered.fadeAt = age;
				covered.end = age + STICKY_LIFE * FADE_SHARE;
			}
		}
		event.lane = lane;
		side[lane] = event;
	});
};
