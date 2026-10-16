import { ResourceType } from '../../proto/api.js';
import { DamageDealtLog } from '../../proto_utils/logs_parser.js';
import { SimResult } from '../../proto_utils/sim_result.js';

// One leveling session (sim/core/leveling.go) read back from its log, for the Events view.
//
// The sim writes "Leveling: ..." lines as the session goes: the state it's in ("now
// fighting"), the mobs it finds and pulls, and where we and the mobs are on the 2D plane
// ("we're at (x, y), moving (vx, vy)"). We turn them into the session's states, its pulls,
// and the tracks the replay's map draws.

export type SessionState = 'searching' | 'pulling' | 'fighting' | 'resting' | 'bandaging' | 'dead' | 'ready';

// How the states look on the session strip and the timeline, and what they mean. The colors
// match the session bar of the leveling results (leveling-session-* in _raid_sim_action.scss).
export const SESSION_STATES: Record<SessionState, { label: string; color: string; tooltip: string }> = {
	searching: { label: 'Searching', color: '#7fb3f0', tooltip: 'Running around looking for a mob.' },
	pulling: { label: 'Pulling', color: '#a9cbf5', tooltip: 'A mob found, waiting for us to pull it.' },
	fighting: { label: 'Fighting', color: '#e0584a', tooltip: 'Mobs on us or running at us.' },
	resting: { label: 'Resting', color: '#8fd18f', tooltip: 'Sitting to eat and drink.' },
	bandaging: { label: 'Bandaging', color: '#6fbf9a', tooltip: 'Bandaging.' },
	dead: { label: 'Dead', color: '#8a8fb8', tooltip: 'Running back to the corpse.' },
	ready: { label: 'Ready', color: '#9aa0a6', tooltip: 'Out of combat, deciding what to do next (buffs, heals).' },
};

export interface StateSpan {
	state: SessionState;
	start: number;
	end: number;
}

// A straight line of movement: at time t we were at (x, y), moving (vx, vy) yards a second,
// until the next one.
export interface Motion {
	t: number;
	x: number;
	y: number;
	vx: number;
	vy: number;
}

// One life of a pool mob: from when we found it to when it died, left or the session ended.
// A pool mob comes back in a later pull as a new life.
export interface MobLife {
	// The mob's label in the log, like "Target 11", and its index in the encounter's targets.
	label: string;
	targetIndex: number;
	level: number;
	foundAt: number;
	// When the search that found it started. The mob stood there all along, so the replay's map
	// shows it from then, before we spot it.
	searchedFrom: number;
	foundDistance: number;
	// When it was found while we were fighting another (a chain pull).
	foundInCombat: boolean;
	pulledAt: number | null;
	pulledFrom: number | null;
	// Noticed: it saw us, because we walked up to it (a face pull) or came within its aggro
	// radius. Spell: something of ours landed on it.
	pulledBy: 'spell' | 'noticed' | null;
	reachedAt: number | null;
	diedAt: number | null;
	// It went back to the pool without a kill: we died, or the session ended before we pulled it.
	leftAt: number | null;
	// The end of this life, whichever way it went.
	end: number;
	// Where it stood and how it ran. Once it reached us it stays where we are.
	track: Array<Motion>;
	// The pull it was part of, once it's pulled.
	pull: Pull | null;
}

// A pull: the stretch we fought, from the first mob pulled until the last died or we did. Chain
// pulls (mobs we found and pulled while fighting) are part of it.
export interface Pull {
	// 1 based, in session order.
	number: number;
	start: number;
	end: number;
	// When we started looking for its first mob (the search or the wait before it).
	searchStart: number;
	mobs: Array<MobLife>;
	kills: number;
	died: boolean;
	damageTaken: number;
	damageDone: number;
	// Shares of the max, 0 to 1. Mana is null for a class without mana.
	healthStart: number;
	healthEnd: number;
	lowestHealth: number;
	manaStart: number | null;
	manaEnd: number | null;
	// The rest after it, until the next search or pull: resting and bandaging seconds, and
	// the food, water and bandages used.
	restSeconds: number;
	ate: boolean;
	drank: boolean;
	bandaged: boolean;
}

// A part of the session we can pick on the strip: a pull, a rest, a search or a death.
export type SectionKind = 'pull' | 'rest' | 'search' | 'death';

export interface SessionSection {
	kind: SectionKind;
	// 1 based within its kind.
	number: number;
	start: number;
	end: number;
	pull: Pull | null;
	// For a rest: what it was for.
	ate?: boolean;
	drank?: boolean;
	bandaged?: boolean;
}

export const SECTION_LABELS: Record<SectionKind, { one: string; many: string }> = {
	pull: { one: 'Pull', many: 'Pulls' },
	rest: { one: 'Rest', many: 'Rests' },
	search: { one: 'Search', many: 'Searches' },
	death: { one: 'Death', many: 'Deaths' },
};

export interface ResourceSample {
	t: number;
	// A share of the max, 0 to 1.
	value: number;
}

export interface TotemSpot {
	t: number;
	spellId: number;
	x: number;
	y: number;
}

export interface LevelingSession {
	// Seconds, the session's length.
	length: number;
	spans: Array<StateSpan>;
	pulls: Array<Pull>;
	sections: Array<SessionSection>;
	mobs: Array<MobLife>;
	kills: Array<{ t: number; mob: MobLife }>;
	deaths: Array<{ t: number; backAt: number; pull: Pull | null }>;
	// Our position. Empty for results simmed before the position lines.
	player: Array<Motion>;
	// Where each totem landed, a few yards from us, or moved to with Totemic Projection. Empty
	// for results simmed before the totem position lines.
	totemSpots: Array<TotemSpot>;
	health: Array<ResourceSample>;
	mana: Array<ResourceSample>;
	stateAt(t: number): SessionState;
	sectionAt(t: number, kind?: SectionKind): SessionSection | null;
}

const cache = new WeakMap<SimResult, LevelingSession | null>();

// Whether the run was in the leveling mode.
export const isLevelingResult = (result: SimResult | null | undefined): boolean => !!result?.request.encounter?.leveling;

// The session of this iteration, or null when it isn't a leveling run. We read it once per
// result.
export function getLevelingSession(result: SimResult | null | undefined): LevelingSession | null {
	if (!result || !isLevelingResult(result)) return null;
	let session = cache.get(result);
	if (session === undefined) {
		session = buildSession(result);
		cache.set(result, session);
	}
	return session;
}

const reNow = /Leveling: now (\w+)/;
const reTotemAt = /Leveling: \{SpellID: (\d+)[^}]*\} at \((-?[\d.]+), (-?[\d.]+)\)/;
const reWe = /Leveling: we're at \((-?[\d.]+), (-?[\d.]+)\), moving \((-?[\d.]+), (-?[\d.]+)\)/;
const reMobAt = /Leveling: (Target (\d+)) at \((-?[\d.]+), (-?[\d.]+)\), moving \((-?[\d.]+), (-?[\d.]+)\)/;
const reSearch = /Leveling: (searching for a mob|looking for another mob)/;
const reFound = /Leveling: found (Target (\d+)) \(level (\d+)\) (\d+) yards away/;
const reNoticed = /Leveling: (Target \d+) noticed us/;
const rePulled = /Leveling: pulled (Target \d+) \(level \d+\) from (\d+) yards/;
const reReached = /Leveling: (Target \d+) reached us/;
const reLeft = /Leveling: (Target \d+) left/;
const reDied = /Leveling: died, back in/;
const reSitting = /Leveling: sitting down \(eat: (\w+), drink: (\w+)\)/;
const reBandaging = /Leveling: bandaging/;
const reTargetDied = /^\[-?[\d.]+\] \[(Target \d+)\] Died/;

function buildSession(result: SimResult): LevelingSession {
	const length = result.result.firstIterationDuration || result.request.encounter?.leveling?.sessionLength || 1200;
	const logs = result.logs;

	const spans: Array<StateSpan> = [];
	const mobs: Array<MobLife> = [];
	const alive = new Map<string, MobLife>();
	const kills: Array<{ t: number; mob: MobLife }> = [];
	const player: Array<Motion> = [];
	const totemSpots: Array<TotemSpot> = [];
	const health: Array<ResourceSample> = [];
	const mana: Array<ResourceSample> = [];
	const deathTimes: Array<number> = [];
	const rests: Array<{ t: number; ate: boolean; drank: boolean; bandaged: boolean }> = [];
	const noticed = new Set<string>();
	// A mob's first position line comes just before the line that says we found it.
	const pendingAt = new Map<string, Motion>();

	let maxHealth = 0;
	let maxMana = 0;
	let state: SessionState | null = null;
	// When the search we're in started. A search that finds nothing searches again, and that
	// is still the same search.
	let searchStart: number | null = null;

	const endLife = (mob: MobLife, t: number) => {
		mob.end = t;
		alive.delete(mob.label);
	};

	for (const log of logs) {
		const t = log.timestamp;
		if (log.isMaxResources()) {
			maxHealth = log.maxValues[ResourceType.ResourceTypeHealth] ?? maxHealth;
			maxMana = log.maxValues[ResourceType.ResourceTypeMana] ?? maxMana;
			continue;
		}
		if (log.isResourceChanged() && log.source?.isPlayer && !log.target) {
			if (log.resourceType == ResourceType.ResourceTypeHealth && maxHealth > 0) {
				health.push({ t, value: Math.min(1, log.valueAfter / maxHealth) });
			} else if (log.resourceType == ResourceType.ResourceTypeMana && maxMana > 0) {
				mana.push({ t, value: Math.min(1, log.valueAfter / maxMana) });
			}
			continue;
		}
		const raw = log.raw;
		const targetDied = raw.match(reTargetDied);
		if (targetDied) {
			const mob = alive.get(targetDied[1]);
			if (mob) {
				mob.diedAt = t;
				kills.push({ t, mob });
				endLife(mob, t);
			}
			continue;
		}
		const at = raw.indexOf('Leveling: ');
		if (at < 0) continue;

		let m: RegExpMatchArray | null;
		if ((m = raw.match(reNow))) {
			const next = m[1] as SessionState;
			if (!(next in SESSION_STATES)) continue;
			let last = spans[spans.length - 1];
			if (last) {
				last.end = t;
				// A state we left at once, like ready when we stand up and search right away.
				if (last.start == t) {
					spans.pop();
					last = spans[spans.length - 1];
				}
			}
			state = next;
			// A chain search ends with the fight, and we only search again once we're out.
			if (next != 'searching' && next != 'pulling' && next != 'fighting') searchStart = null;
			if (last && last.state == next) {
				last.end = length;
				continue;
			}
			spans.push({ state: next, start: t, end: length });
		} else if ((m = raw.match(reWe))) {
			player.push({ t, x: +m[1], y: +m[2], vx: +m[3], vy: +m[4] });
		} else if ((m = raw.match(reTotemAt))) {
			totemSpots.push({ t, spellId: parseInt(m[1]), x: +m[2], y: +m[3] });
		} else if ((m = raw.match(reMobAt))) {
			const motion = { t, x: +m[3], y: +m[4], vx: +m[5], vy: +m[6] };
			const mob = alive.get(m[1]);
			if (mob) mob.track.push(motion);
			else pendingAt.set(m[1], motion);
		} else if ((m = raw.match(reSearch))) {
			if (searchStart == null || m[1] == 'looking for another mob') searchStart = t;
		} else if ((m = raw.match(reFound))) {
			const mob: MobLife = {
				label: m[1],
				targetIndex: parseInt(m[2]) - 1,
				level: parseInt(m[3]),
				foundAt: t,
				searchedFrom: searchStart ?? t,
				foundDistance: parseInt(m[4]),
				foundInCombat: state == 'fighting',
				pulledAt: null,
				pulledFrom: null,
				pulledBy: null,
				reachedAt: null,
				diedAt: null,
				leftAt: null,
				end: length,
				track: [],
				pull: null,
			};
			const first = pendingAt.get(m[1]);
			if (first) mob.track.push(first);
			pendingAt.delete(m[1]);
			alive.set(mob.label, mob);
			mobs.push(mob);
			searchStart = null;
		} else if ((m = raw.match(reNoticed))) {
			noticed.add(m[1]);
		} else if ((m = raw.match(rePulled))) {
			const mob = alive.get(m[1]);
			if (mob) {
				mob.pulledAt = t;
				mob.pulledFrom = parseInt(m[2]);
				mob.pulledBy = noticed.has(m[1]) ? 'noticed' : 'spell';
			}
			noticed.delete(m[1]);
		} else if ((m = raw.match(reReached))) {
			const mob = alive.get(m[1]);
			if (mob) mob.reachedAt = t;
		} else if ((m = raw.match(reLeft))) {
			const mob = alive.get(m[1]);
			if (mob) {
				mob.leftAt = t;
				endLife(mob, t);
			}
		} else if (reDied.test(raw)) {
			deathTimes.push(t);
		} else if ((m = raw.match(reSitting))) {
			rests.push({ t, ate: m[1] == 'true', drank: m[2] == 'true', bandaged: false });
		} else if (reBandaging.test(raw)) {
			rests.push({ t, ate: false, drank: false, bandaged: true });
		}
	}

	// Older results have no state lines. Then we make do without the strip's states.
	if (!spans.length) spans.push({ state: 'fighting', start: 0, end: length });
	spans[spans.length - 1].end = length;

	// Pulls: each stretch of fighting.
	const pulls: Array<Pull> = [];
	spans.forEach((span, i) => {
		if (span.state != 'fighting') return;
		let searchStart = span.start;
		for (let j = i - 1; j >= 0 && (spans[j].state == 'pulling' || spans[j].state == 'searching'); j--) searchStart = spans[j].start;
		pulls.push({
			number: pulls.length + 1,
			start: span.start,
			end: span.end,
			searchStart,
			mobs: [],
			kills: 0,
			died: false,
			damageTaken: 0,
			damageDone: 0,
			healthStart: valueAt(health, span.start, 1),
			healthEnd: valueAt(health, span.end, 1),
			lowestHealth: lowestBetween(health, span.start, span.end),
			manaStart: maxMana > 0 ? valueAt(mana, span.start, 1) : null,
			manaEnd: maxMana > 0 ? valueAt(mana, span.end, 1) : null,
			restSeconds: 0,
			ate: false,
			drank: false,
			bandaged: false,
		});
	});
	const pullAt = (t: number): Pull | null => {
		let lo = 0;
		let hi = pulls.length - 1;
		while (lo <= hi) {
			const mid = (lo + hi) >> 1;
			if (pulls[mid].end < t) lo = mid + 1;
			else if (pulls[mid].start > t) hi = mid - 1;
			else return pulls[mid];
		}
		return null;
	};
	for (const mob of mobs) {
		if (mob.pulledAt == null) continue;
		mob.pull = pullAt(mob.pulledAt);
		if (mob.pull) {
			mob.pull.mobs.push(mob);
			if (mob.diedAt != null) mob.pull.kills++;
		}
	}
	for (const t of deathTimes) {
		const pull = pullAt(t);
		if (pull) pull.died = true;
	}
	for (const log of logs) {
		if (!log.isDamageDealt()) continue;
		const dmg = log as DamageDealtLog;
		if (dmg.amount <= 0) continue;
		const pull = pullAt(log.timestamp);
		if (!pull) continue;
		if (log.target?.isPlayer && log.source?.isTarget) pull.damageTaken += dmg.amount;
		else if (log.target?.isTarget && log.source && !log.source.isTarget) pull.damageDone += dmg.amount;
	}

	// The rest after each pull: the time resting and bandaging until the next pull's search.
	pulls.forEach((pull, i) => {
		const until = i + 1 < pulls.length ? pulls[i + 1].searchStart : length;
		for (const span of spans) {
			if (span.start >= pull.end && span.start < until && (span.state == 'resting' || span.state == 'bandaging'))
				pull.restSeconds += span.end - span.start;
		}
		for (const rest of rests) {
			if (rest.t >= pull.end && rest.t < until) {
				pull.ate ||= rest.ate;
				pull.drank ||= rest.drank;
				pull.bandaged ||= rest.bandaged;
			}
		}
	});

	const deaths = deathTimes.map(t => {
		const span = spans.find(s => s.state == 'dead' && s.start == t);
		return { t, backAt: span ? span.end : Math.min(length, t), pull: pullAt(t) };
	});

	const sections = buildSections(spans, pulls, rests);

	const session: LevelingSession = {
		length,
		spans,
		pulls,
		sections,
		mobs,
		kills,
		deaths,
		player,
		totemSpots,
		health,
		mana,
		stateAt(t: number) {
			const i = lastAtOrBefore(spans, t, s => s.start);
			return spans[Math.max(0, i)].state;
		},
		sectionAt(t: number, kind?: SectionKind) {
			return sections.find(s => (!kind || s.kind == kind) && s.start <= t && t < s.end) ?? null;
		},
	};
	return session;
}

// Sections: a pull is its fighting stretch. The resting and bandaging stretches between two
// pulls make one rest, a search is a searching and pulling stretch, and a death is the corpse
// run.
function buildSections(
	spans: Array<StateSpan>,
	pulls: Array<Pull>,
	rests: Array<{ t: number; ate: boolean; drank: boolean; bandaged: boolean }>,
): Array<SessionSection> {
	const sections: Array<SessionSection> = [];
	const counts: Record<SectionKind, number> = { pull: 0, rest: 0, search: 0, death: 0 };
	const add = (kind: SectionKind, start: number, end: number, pull: Pull | null = null) => {
		const last = sections[sections.length - 1];
		// Stretches of the same kind next to each other, like resting then bandaging, make one.
		if (last && last.kind == kind && kind != 'pull' && Math.abs(last.end - start) < 1e-6) {
			last.end = end;
			return last;
		}
		const section: SessionSection = { kind, number: ++counts[kind], start, end, pull };
		sections.push(section);
		return section;
	};
	let pullIndex = 0;
	for (const span of spans) {
		switch (span.state) {
			case 'fighting':
				add('pull', span.start, span.end, pulls[pullIndex++] ?? null);
				break;
			case 'searching':
			case 'pulling':
				add('search', span.start, span.end);
				break;
			case 'resting':
			case 'bandaging': {
				const section = add('rest', span.start, span.end);
				for (const rest of rests) {
					if (rest.t >= section.start && rest.t < section.end) {
						section.ate ||= rest.ate;
						section.drank ||= rest.drank;
						section.bandaged ||= rest.bandaged;
					}
				}
				break;
			}
			case 'dead':
				add('death', span.start, span.end);
				break;
			case 'ready':
				// A short wait to decide; it belongs to no section of its own.
				break;
		}
	}
	return sections;
}

export function lastAtOrBefore<T>(items: Array<T>, t: number, key: (item: T) => number): number {
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
}

// The value of a resource at t: the last sample at or before it.
export function valueAt(samples: Array<ResourceSample>, t: number, fallback: number): number {
	const i = lastAtOrBefore(samples, t, s => s.t);
	return i < 0 ? fallback : samples[i].value;
}

function lowestBetween(samples: Array<ResourceSample>, start: number, end: number): number {
	let lowest = valueAt(samples, start, 1);
	for (
		let i = Math.max(
			0,
			lastAtOrBefore(samples, start, s => s.t),
		);
		i < samples.length && samples[i].t <= end;
		i++
	) {
		lowest = Math.min(lowest, samples[i].value);
	}
	return lowest;
}

// Where a track is at t: the last motion at or before it, carried on in a straight line.
export function positionAt(track: Array<Motion>, t: number): { x: number; y: number; moving: boolean } | null {
	const i = lastAtOrBefore(track, t, m => m.t);
	if (i < 0) return track.length ? { x: track[0].x, y: track[0].y, moving: false } : null;
	const m = track[i];
	const dt = t - m.t;
	return { x: m.x + m.vx * dt, y: m.y + m.vy * dt, moving: m.vx != 0 || m.vy != 0 };
}

// Where a mob is at t: on its track while it stands or runs, and where we are once it reached
// us. Null when it isn't out of the pool at t.
export function mobPositionAt(session: LevelingSession, mob: MobLife, t: number): { x: number; y: number; moving: boolean } | null {
	if (t < mob.searchedFrom || t > mob.end) return null;
	// Before we found it, it stood where we found it.
	if (t < mob.foundAt) return mob.track[0] ? { x: mob.track[0].x, y: mob.track[0].y, moving: false } : null;
	if (mob.reachedAt != null && t >= mob.reachedAt) {
		return positionAt(session.player, t);
	}
	return positionAt(mob.track, t);
}

// The mob's name for people: "Mob 11 (level 31)" reads better than the log's "Target 11".
export const mobName = (mob: MobLife): string => `Mob ${mob.targetIndex + 1}`;

// Like 12:05, or 12:05.3 with tenths.
export function formatClock(seconds: number, tenths = false): string {
	const sign = seconds < 0 ? '-' : '';
	// We round first, so 59.97 reads 1:00.0 and not 0:60.0.
	const abs = tenths ? Math.round(Math.abs(seconds) * 10) / 10 : Math.floor(Math.abs(seconds));
	const minutes = Math.floor(abs / 60);
	const rest = abs - minutes * 60;
	const secs = tenths ? rest.toFixed(1).padStart(4, '0') : String(Math.floor(rest)).padStart(2, '0');
	return `${sign}${minutes}:${secs}`;
}

// Like 42 s, or 3:05 over a minute.
export function formatSpan(seconds: number): string {
	return seconds < 60 ? `${seconds.toFixed(seconds < 10 ? 1 : 0)} s` : formatClock(seconds);
}

export const sectionTitle = (section: SessionSection): string => `${SECTION_LABELS[section.kind].one} ${section.number}`;
