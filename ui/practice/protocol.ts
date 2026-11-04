// The messages between the page and the practice session. sim/practice/protocol.go has the same
// types. Every time is sim time in milliseconds, unless a name says wall time.

export type UnitRef = { kind: 'player' | 'target' | 'enemy' | 'ally'; index: number };

export type Input = {
	id: number;
	kind: 'cast' | 'useItem' | 'startAttack' | 'stopAttack' | 'target' | 'clearTarget' | 'cancelCast' | 'move' | 'jump';
	spellId?: number;
	itemId?: number;
	target?: UnitRef;
	pos?: [number, number];
	velocity?: [number, number];
	facing?: number;
	turnRate?: number;
};

export type Settings = {
	mode: 'free' | 'encounter';
	lengthMs: number;
	countdownMs: number;
	pingMs: number;
	executeFromStart: boolean;
	spellQueueWindowMs: number;
};

export type Answer = { inputId: number; ok: boolean; error?: string };

export type HitEvent = {
	at: number;
	spellId?: number;
	itemId?: number;
	otherId?: number;
	tag?: number;
	damage: number;
	outcome: 'hit' | 'crit' | 'miss' | 'dodge' | 'parry' | 'glance' | 'block' | 'resist';
	periodic?: boolean;
	school: number;
	target: number;
	source: 'player' | 'pet' | 'totem';
	// The totem's element for a totem's hit (Searing bolt, Magma pulse, Fire Nova).
	totem?: string;
	// The client's spell to show when it differs from spellId (imbue procs, Fire Nova).
	visualSpellId?: number;
	// When a Searing Totem's next bolt hits (sim ms).
	nextAt?: number;
};

export type CastEvent = { at: number; spellId?: number; itemId?: number; totem?: string; visualSpellId?: number };

export type UnitState = {
	ref: UnitRef;
	pos: [number, number];
	facing: number;
	airborneUntil?: number;
	reach: number;
	height: number;
};

export type AuraState = {
	spellId?: number;
	itemId?: number;
	label: string;
	stacks?: number;
	startedAt: number;
	expiresAt?: number;
	duration?: number;
};

export type Cooldown = { spellId?: number; itemId?: number; readyAt: number; length: number };

export type TotemState = { element: string; spellId: number; expiresAt: number; pos: [number, number] };

export type PlayerState = {
	health: number;
	maxHealth: number;
	mana: number;
	maxMana: number;
	gcdStart: number;
	gcdEnd: number;
	cast?: { spellId: number; start: number; end: number };
	attackOn: boolean;
	swinging: boolean;
	nextSwingAt?: number;
	swingMs: number;
	inMeleeRange: boolean;
	moving: boolean;
	hasTarget: boolean;
	auras: AuraState[] | null;
	cooldowns: Cooldown[] | null;
	totems?: TotemState[];
	damageDone: number;
	queuedSpellId?: number;
	mainHandImbue?: string;
	offHandImbue?: string;
	outOfRangeMs: number;
};

export type TargetState = { name: string; level: number; damageTaken: number; debuffs: AuraState[] | null };

export type SpellSummary = {
	spellId?: number;
	itemId?: number;
	otherId?: number;
	tag?: number;
	damage: number;
	casts: number;
	hits: number;
	crits: number;
	misses: number;
	ticks: number;
};

export type Summary = {
	durationMs: number;
	damage: number;
	dps: number;
	spells: SpellSummary[] | null;
	auras: { spellId?: number; itemId?: number; uptimePct: number }[] | null;
	idleGcdMs: number;
	outOfRangeMs: number;
	casts: number;
};

export type FightInfo = {
	mode: 'free' | 'encounter';
	phase: 'prepull' | 'countdown' | 'fight' | 'ended';
	lengthMs: number;
	endedAtMs?: number;
	earlyPull?: boolean;
};

export type Batch = {
	simTime: number;
	logLines?: string[];
	answers?: Answer[];
	hits?: HitEvent[];
	casts?: CastEvent[];
	errors?: string[];
	units: UnitState[];
	player?: PlayerState;
	target?: TargetState;
	rebase?: { shiftMs: number };
	pulled?: { early: boolean };
	ended?: { summary: Summary | null; stoppedEarly: boolean };
	fight: FightInfo;
};

export type ToolkitSpell = {
	spellId?: number;
	itemId?: number;
	tag?: number;
	castMs: number;
	gcdMs: number;
	cdMs: number;
	cost: number;
	maxRange: number;
	melee: boolean;
	harmful: boolean;
	kind: 'spell' | 'item' | 'totem';
};

// Sheet is the character sheet when the session starts (sim/practice/sheet.go). Stats and base
// are by the Stat enum's index (proto/common.proto), pseudo by the PseudoStat enum's.
export type Sheet = {
	level: number;
	stats: number[];
	base: number[];
	pseudo: number[];
	weapons: { slot: 'mainHand' | 'offHand' | 'ranged'; min: number; max: number; speed: number; type?: string }[] | null;
	items: { slot: number; id: number; quality: number; name: string; enchant?: number }[] | null;
	damageMultiplier: number;
};

// Messages between the page and the worker.
export type ToWorker =
	| { type: 'start'; request: unknown; settings: Settings; seed: number; wallNow: number }
	| { type: 'input'; input: Input; pressedAtWall: number }
	| { type: 'startEncounter'; wallNow: number }
	| { type: 'stop'; wallNow: number }
	| { type: 'pause'; wallNow: number }
	| { type: 'resume'; wallNow: number }
	| { type: 'attempt' };

export type FromWorker =
	| { type: 'ready'; loadMs: number }
	| { type: 'started'; toolkit: ToolkitSpell[]; shown?: number[]; sheet?: Sheet; startMs: number; batch: Batch; wallAtStart: number; callMs: number }
	| { type: 'batch'; batch: Batch; simNowMs: number; clock: ClockState; callMs: number }
	| { type: 'attempt'; attempt: unknown }
	| { type: 'error'; message: string };

// How the worker maps wall time to sim time, so the page can draw at the same clock.
export type ClockState = { wallAtStart: number; startMs: number; shiftMs: number; pausedMs: number; pausedAt?: number };

export function simTimeAt(clock: ClockState, wall: number): number {
	const paused = clock.pausedMs + (clock.pausedAt !== undefined ? wall - clock.pausedAt : 0);
	return wall - clock.wallAtStart - paused + clock.startMs + clock.shiftMs;
}

export function wallNow(): number {
	return performance.timeOrigin + performance.now();
}
