//go:build practice

// Package practice runs one sim in real time for the practice room (docs/practice_room.md).
//
// The player presses every cast by hand. The session takes those inputs, moves the player in a
// small world, and drives a normal sim from outside through what the core exports. Nothing in
// the regular sim imports this package.
package practice

// The messages between the page and the session. ui/practice/protocol.ts has the same types.
//
// Every time in a message is sim time in milliseconds, fractional. The sim works in whole
// nanoseconds, and the session converts.

// UnitRef names a unit in the session: you, the dummy, later bosses, bots and other players.
type UnitRef struct {
	Kind  string `json:"kind"` // "player", "target", "enemy" or "ally"
	Index int    `json:"index"`
}

// Input is one thing the player did. The page resolves macros, @mouseover and modifiers at the
// press, and sends plain actions.
type Input struct {
	ID int `json:"id"`
	// Kind is one of cast, useItem, startAttack, stopAttack, target, clearTarget, cancelCast,
	// move and jump.
	Kind    string   `json:"kind"`
	SpellID int32    `json:"spellId,omitempty"`
	ItemID  int32    `json:"itemId,omitempty"`
	Target  *UnitRef `json:"target,omitempty"`

	// A move input says where we are and how we move from its time on. The page moves us at
	// once, and the session takes its word for it, as WoW trusts the client's movement.
	Pos      [2]float64 `json:"pos,omitempty"`
	Velocity [2]float64 `json:"velocity,omitempty"`
	Facing   float64    `json:"facing,omitempty"`
	TurnRate float64    `json:"turnRate,omitempty"`
}

// Answer tells the page whether an input went through, with the game's error text when not.
type Answer struct {
	InputID int    `json:"inputId"`
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
}

// Settings are the fight settings the page starts a session with.
type Settings struct {
	Mode             string  `json:"mode"` // "free" or "encounter"
	LengthMs         float64 `json:"lengthMs"`
	CountdownMs      float64 `json:"countdownMs"`
	PingMs           float64 `json:"pingMs"`
	ExecuteFromStart bool    `json:"executeFromStart"`
	// SpellQueueWindowMs is the SpellQueueWindow CVar, 400 by default.
	SpellQueueWindowMs float64 `json:"spellQueueWindowMs"`
}

// Batch is what the session sends after it ran: everything that happened since the last one,
// and the state the HUD draws.
type Batch struct {
	SimTime  float64      `json:"simTime"`
	LogLines []string     `json:"logLines,omitempty"`
	Answers  []Answer     `json:"answers,omitempty"`
	Hits     []HitEvent   `json:"hits,omitempty"`
	Casts    []CastEvent  `json:"casts,omitempty"`
	Errors   []string     `json:"errors,omitempty"`
	Units    []UnitState  `json:"units"`
	Player   *PlayerState `json:"player,omitempty"`
	Target   *TargetState `json:"target,omitempty"`
	// Rebase says the pull swap moved sim time by this much. The page moves its clock, its HUD
	// times and its inputs in flight by it.
	Rebase *Rebase   `json:"rebase,omitempty"`
	Pulled *Pulled   `json:"pulled,omitempty"`
	Ended  *Ended    `json:"ended,omitempty"`
	Fight  FightInfo `json:"fight"`
}

type Rebase struct {
	ShiftMs float64 `json:"shiftMs"`
}

type Pulled struct {
	Early bool `json:"early"`
}

type Ended struct {
	Summary      *Summary `json:"summary"`
	StoppedEarly bool     `json:"stoppedEarly"`
}

// FightInfo is where the fight stands.
type FightInfo struct {
	Mode string `json:"mode"`
	// Phase is "prepull", "countdown", "fight" or "ended".
	Phase    string  `json:"phase"`
	LengthMs float64 `json:"lengthMs"`
	// CountdownEndsAt is 0, the pull, while the countdown runs.
	EndedAtMs float64 `json:"endedAtMs,omitempty"`
	EarlyPull bool    `json:"earlyPull,omitempty"`
}

// HitEvent is one result against an enemy: a hit, a miss or a tick.
type HitEvent struct {
	At       float64 `json:"at"`
	SpellID  int32   `json:"spellId,omitempty"`
	ItemID   int32   `json:"itemId,omitempty"`
	OtherID  int32   `json:"otherId,omitempty"`
	Tag      int32   `json:"tag,omitempty"`
	Damage   float64 `json:"damage"`
	Outcome  string  `json:"outcome"`
	Periodic bool    `json:"periodic,omitempty"`
	School   int     `json:"school"`
	Target   int     `json:"target"`
	// Source is "player", "pet" or "totem".
	Source string `json:"source"`
}

// CastEvent is one cast that went off, for the cast sounds and animations.
type CastEvent struct {
	At      float64 `json:"at"`
	SpellID int32   `json:"spellId,omitempty"`
	ItemID  int32   `json:"itemId,omitempty"`
}

type UnitState struct {
	Ref           UnitRef    `json:"ref"`
	Pos           [2]float64 `json:"pos"`
	Facing        float64    `json:"facing"`
	AirborneUntil float64    `json:"airborneUntil,omitempty"`
	Reach         float64    `json:"reach"`
	Height        float64    `json:"height"`
}

// PlayerState is your unit at the batch's sim time.
type PlayerState struct {
	Health    float64 `json:"health"`
	MaxHealth float64 `json:"maxHealth"`
	Mana      float64 `json:"mana"`
	MaxMana   float64 `json:"maxMana"`
	// ManaPerSecond is the regen rate now, so the HUD can count mana between batches.
	GCDStart float64 `json:"gcdStart"`
	GCDEnd   float64 `json:"gcdEnd"`
	Cast     *Cast   `json:"cast,omitempty"`
	// AttackOn is whether auto attack is on, and Swinging whether it swings now (in range and
	// facing the target).
	AttackOn       bool          `json:"attackOn"`
	Swinging       bool          `json:"swinging"`
	NextSwingAt    float64       `json:"nextSwingAt,omitempty"`
	SwingMs        float64       `json:"swingMs"`
	InMeleeRange   bool          `json:"inMeleeRange"`
	Moving         bool          `json:"moving"`
	HasTarget      bool          `json:"hasTarget"`
	Auras          []AuraState   `json:"auras"`
	Cooldowns      []Cooldown    `json:"cooldowns"`
	Totems         []TotemState  `json:"totems,omitempty"`
	DamageDone     float64       `json:"damageDone"`
	QueuedSpellID  int32         `json:"queuedSpellId,omitempty"`
	Resources      []ResourceBar `json:"resources,omitempty"`
	MainHandImbue  string        `json:"mainHandImbue,omitempty"`
	ImbueCastAtMs  float64       `json:"imbueCastAtMs,omitempty"`
	OutOfRangeTime float64       `json:"outOfRangeMs"`
}

type ResourceBar struct {
	Kind    string  `json:"kind"`
	Current float64 `json:"current"`
	Max     float64 `json:"max"`
}

type Cast struct {
	SpellID int32   `json:"spellId"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
}

type AuraState struct {
	SpellID   int32   `json:"spellId,omitempty"`
	ItemID    int32   `json:"itemId,omitempty"`
	Label     string  `json:"label"`
	Stacks    int32   `json:"stacks,omitempty"`
	StartedAt float64 `json:"startedAt"`
	// ExpiresAt is left out for an aura that never expires.
	ExpiresAt float64 `json:"expiresAt,omitempty"`
	Duration  float64 `json:"duration,omitempty"`
}

type Cooldown struct {
	SpellID int32   `json:"spellId,omitempty"`
	ItemID  int32   `json:"itemId,omitempty"`
	ReadyAt float64 `json:"readyAt"`
	Length  float64 `json:"length"`
}

type TotemState struct {
	Element   string     `json:"element"`
	SpellID   int32      `json:"spellId"`
	ExpiresAt float64    `json:"expiresAt"`
	Pos       [2]float64 `json:"pos"`
}

// TargetState is the dummy at the batch's sim time.
type TargetState struct {
	Name        string      `json:"name"`
	Level       int32       `json:"level"`
	DamageTaken float64     `json:"damageTaken"`
	Debuffs     []AuraState `json:"debuffs"`
}

// Spell is one entry of your toolkit, for the spellbook and the bars.
type Spell struct {
	SpellID  int32   `json:"spellId,omitempty"`
	ItemID   int32   `json:"itemId,omitempty"`
	Tag      int32   `json:"tag,omitempty"`
	CastMs   float64 `json:"castMs"`
	GCDMs    float64 `json:"gcdMs"`
	CDMs     float64 `json:"cdMs"`
	Cost     float64 `json:"cost"`
	MaxRange float64 `json:"maxRange"`
	Melee    bool    `json:"melee"`
	Harmful  bool    `json:"harmful"`
	// Imbue is the main hand imbue the spell puts on, for the imbue spells.
	Imbue string `json:"imbue,omitempty"`
	// Kind is "spell", "item" or "racial".
	Kind string `json:"kind"`
}
