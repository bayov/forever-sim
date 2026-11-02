//go:build practice

package practice

import (
	"math"
	"sort"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// run is one sim of a session, from the start of its prepull window. The session swaps in a new
// run at the pull and when the prepull window runs out.
type run struct {
	s     *Session
	sim   *core.Simulation
	start time.Duration

	char   *core.Character
	shaman *shaman.Shaman
	dummy  *core.Unit

	// toolkit holds your castable spells by spell id, and items your usable items by item id.
	toolkit map[int32]*core.Spell
	items   map[int32]*core.Spell
	// externals are the party cooldowns other players cast on you (Power Infusion, Innervate,
	// Mana Tide), which the session fires on the sim's own rule.
	externals []*core.MajorCooldown

	me, target body

	queue    []*action
	queuePA  *core.PendingAction
	worldSeq int
	ticking  bool

	// quiet is on while a swap replays inputs the page already heard about.
	quiet   bool
	fullLog []string

	attackOn  bool
	hasTarget bool
	swinging  bool
	inMelee   bool

	// queued is the cast waiting for the GCD's or the cast's end.
	queued      *action
	queuedInput int
	queuedSpell *core.Spell
	// castCheck is the range and facing check at the current cast's end.
	castCheck *core.PendingAction
	// ghostWolf is the Ghost Wolf form aura, for a shaman who knows it.
	ghostWolf *core.Aura

	// pullAt is set when a harmful result against an enemy came before 0.
	pullAt *time.Duration

	gcdStart     time.Duration
	busyUntil    time.Duration
	idleGCD      time.Duration
	outOfRange   time.Duration
	lastWorldAt  time.Duration
	damage       float64
	hits         map[hitKey]*SpellSummary
	casts        map[hitKey]int32
	imbueCastAt  time.Duration
	totemPos     [4]vec2
	totemPosSet  [4]bool
}

type hitKey struct {
	spellID, itemID, otherID, tag int32
}

// action is one entry of the session queue: an input or a world change. The queue runs them in
// order of time, then world changes before inputs, then by arrival.
type action struct {
	at   time.Duration
	kind int // 0 world, 1 input
	seq  int
	do   func(sim *core.Simulation)
	done bool
}

// onEnvironment runs while the sim is built, before it's finalized: the watcher aura on each
// enemy, and the class hooks.
func (r *run) onEnvironment(env *core.Environment) {
	for _, target := range env.Encounter.TargetUnits {
		target := target
		target.RegisterAura(core.Aura{
			Label:    "Practice Watcher",
			Duration: core.NeverExpires,
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				aura.Activate(sim)
			},
			OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				r.onHarmful(sim, spell, result, false)
			},
			OnPeriodicDamageTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				r.onHarmful(sim, spell, result, true)
			},
		})
	}

	player := env.Raid.Parties[0].Players[0]
	if sh, ok := player.(shaman.ShamanAgent); ok {
		sh.GetShaman().EnableImbueSwaps()
		r.registerUtility(sh.GetShaman())
		sh.GetShaman().FireTotemInReach = func(sim *core.Simulation, mob *core.Unit, yards float64) bool {
			if !r.totemPosSet[shaman.FireTotem] {
				return true
			}
			return r.totemPos[shaman.FireTotem].dist(r.targetPos(sim)) <= yards+r.target.reach
		}
	}
}

// onHarmful sees each result against an enemy: before 0 it's the pull, from 0 on it's a hit for
// the meter and the combat text.
func (r *run) onHarmful(sim *core.Simulation, spell *core.Spell, result *core.SpellResult, periodic bool) {
	if !r.isOurs(spell.Unit) {
		return
	}
	if sim.CurrentTime < 0 {
		if r.pullAt == nil {
			at := sim.CurrentTime
			r.pullAt = &at
		}
		return
	}
	r.damage += result.Damage
	key := hitKey{spell.ActionID.SpellID, spell.ActionID.ItemID, int32(spell.ActionID.OtherID), spell.ActionID.Tag}
	sum := r.hits[key]
	if sum == nil {
		sum = &SpellSummary{SpellID: key.spellID, ItemID: key.itemID, OtherID: key.otherID, Tag: key.tag}
		r.hits[key] = sum
	}
	sum.Damage += result.Damage
	outcome := outcomeName(result)
	switch outcome {
	case "miss", "dodge", "parry":
		sum.Misses++
	default:
		sum.Hits++
		if outcome == "crit" {
			sum.Crits++
		}
	}
	if periodic {
		sum.Ticks++
	}
	if r.quiet {
		return
	}
	source := "player"
	switch spell.Unit.Type {
	case core.PetUnit:
		source = "pet"
	}
	if spell.Flags.Matches(core.SpellFlagCastFromTotem) {
		source = "totem"
	}
	r.s.batch.Hits = append(r.s.batch.Hits, HitEvent{
		At:       toMs(sim.CurrentTime),
		SpellID:  spell.ActionID.SpellID,
		ItemID:   spell.ActionID.ItemID,
		OtherID:  int32(spell.ActionID.OtherID),
		Tag:      spell.ActionID.Tag,
		Damage:   result.Damage,
		Outcome:  outcome,
		Periodic: periodic,
		School:   int(spell.SpellSchool),
		Target:   int(result.Target.Index),
		Source:   source,
	})
}

func outcomeName(result *core.SpellResult) string {
	switch {
	case result.Outcome.Matches(core.OutcomeMiss):
		return "miss"
	case result.Outcome.Matches(core.OutcomeDodge):
		return "dodge"
	case result.Outcome.Matches(core.OutcomeParry):
		return "parry"
	case result.DidCrit():
		return "crit"
	case result.DidGlance():
		return "glance"
	case result.DidBlock():
		return "block"
	case result.DidResist():
		return "resist"
	}
	return "hit"
}

// isOurs is whether a unit is you or one of your pets or guardians.
func (r *run) isOurs(unit *core.Unit) bool {
	if unit == &r.char.Unit {
		return true
	}
	for _, pet := range r.char.PetAgents {
		if unit == &pet.GetPet().Unit {
			return true
		}
	}
	return false
}

// setup follows the core's own session steps (sim/lib/library.go) with the room's changes.
func (r *run) setup() {
	sim := r.sim
	agent := sim.Raid.Parties[0].Players[0]
	r.char = agent.GetCharacter()
	if sh, ok := agent.(shaman.ShamanAgent); ok {
		r.shaman = sh.GetShaman()
	}
	r.dummy = sim.Encounter.TargetUnits[0]
	r.hits = map[hitKey]*SpellSummary{}
	r.casts = map[hitKey]int32{}
	r.imbueCastAt = core.NeverExpires

	// Your unit never acts on its own, and its auto attacks start off.
	r.char.ManualControl = true
	r.char.AutoAttacks.SkipPullStart = true

	for _, spell := range r.char.Spellbook {
		if !spell.Flags.Matches(core.SpellFlagAPL) {
			continue
		}
		id := spell.ActionID
		switch {
		case id.Tag < 0:
			continue
		case id.ItemID != 0:
			if old := r.items[id.ItemID]; old == nil || id.Tag < old.ActionID.Tag {
				r.items[id.ItemID] = spell
			}
		case id.SpellID != 0:
			if old := r.toolkit[id.SpellID]; old == nil || id.Tag < old.ActionID.Tag {
				r.toolkit[id.SpellID] = spell
			}
		}
	}
	sim.Reset()
	if r.shaman != nil {
		// Each imbue aura turns itself on at Reset: only the settings' imbue stays.
		r.shaman.ResetImbueSwaps(sim)
	}
	for _, mcd := range r.char.GetMajorCooldowns() {
		if mcd != nil && mcd.Spell != nil && mcd.Spell.ActionID.Tag == -1 {
			r.externals = append(r.externals, mcd)
		}
	}

	sim.PrePull()
	sim.CurrentTime = min(sim.CurrentTime, r.start)
	r.lastWorldAt = sim.CurrentTime
	r.busyUntil = sim.CurrentTime
	r.char.PseudoStats.FiveSecondRuleRefreshTime = sim.CurrentTime - 5*time.Second
	sim.Log = r.log

	// You stand 15 yards south of the dummy, facing it. The dummy faces you and never turns.
	r.target = body{pos: vec2{0, 0}, facing: math.Pi / 2 * 3, reach: dummyReach, height: dummyHeight, at: sim.CurrentTime}
	r.me = body{pos: vec2{0, -15}, facing: math.Pi / 2, reach: playerReach, height: playerHeight, at: sim.CurrentTime, airborneUntil: math.MinInt64}
	r.hasTarget = true
	r.char.CurrentTarget = r.dummy
	r.updateWorld(sim)
}

func (r *run) log(message string, vals ...interface{}) {
	if r.s.keepLog {
		r.fullLog = append(r.fullLog, formatLog(r.sim.CurrentTime, message, vals...))
	}
	if r.quiet {
		return
	}
	r.s.logLines = append(r.s.logLines, formatLog(r.sim.CurrentTime, message, vals...))
}

func (r *run) onPulled() {
	r.startExternals()
	r.keepTicking(r.sim)
}

// runUntil runs every event before until, with no fight logic: a swap replaying inputs.
func (r *run) runUntil(until time.Duration) {
	for r.sim.NextEventTime() < until {
		r.sim.Step()
	}
}

// The session queue.

func (r *run) enqueueInput(rec recordedInput) {
	in := rec.Input
	r.enqueue(&action{at: rec.At, kind: 1, seq: rec.Seq, do: func(sim *core.Simulation) {
		r.doInput(sim, in)
	}})
}

func (r *run) enqueueWorld(at time.Duration, do func(sim *core.Simulation)) *action {
	a := &action{at: at, kind: 0, seq: r.worldSeq, do: do}
	r.worldSeq++
	r.enqueue(a)
	return a
}

func (r *run) enqueue(a *action) {
	i := sort.Search(len(r.queue), func(i int) bool { return less(a, r.queue[i]) })
	r.queue = append(r.queue, nil)
	copy(r.queue[i+1:], r.queue[i:])
	r.queue[i] = a
	r.schedule()
}

func less(a, b *action) bool {
	if a.at != b.at {
		return a.at < b.at
	}
	if a.kind != b.kind {
		return a.kind < b.kind
	}
	return a.seq < b.seq
}

// schedule keeps one pending action in the sim, at the head of the queue.
func (r *run) schedule() {
	for len(r.queue) > 0 && r.queue[0].done {
		r.queue = r.queue[1:]
	}
	if len(r.queue) == 0 {
		if r.queuePA != nil {
			r.queuePA.Cancel(r.sim)
			r.queuePA = nil
		}
		return
	}
	head := r.queue[0]
	at := max(head.at, r.sim.CurrentTime)
	if r.queuePA != nil {
		if r.queuePA.NextActionAt == at {
			return
		}
		r.queuePA.Cancel(r.sim)
	}
	r.queuePA = &core.PendingAction{
		NextActionAt: at,
		Priority:     sessionPriority,
		OnAction: func(sim *core.Simulation) {
			r.queuePA = nil
			if len(r.queue) > 0 && r.queue[0].at <= sim.CurrentTime {
				a := r.queue[0]
				r.queue = r.queue[1:]
				if !a.done {
					a.done = true
					a.do(sim)
				}
			}
			r.schedule()
		},
	}
	r.sim.AddPendingAction(r.queuePA)
}

// The party cooldowns of other players.
func (r *run) startExternals() {
	if len(r.externals) == 0 {
		return
	}
	var tick func(sim *core.Simulation)
	tick = func(sim *core.Simulation) {
		for _, mcd := range r.externals {
			if !mcd.IsReady(sim) || !mcd.Spell.CanCast(sim, &r.char.Unit) {
				continue
			}
			if mcd.ShouldActivate != nil && !mcd.ShouldActivate(sim, r.char) {
				continue
			}
			mcd.TryActivate(sim, r.char)
		}
		r.enqueueWorld(sim.CurrentTime+externalsPeriod, tick)
	}
	r.enqueueWorld(max(r.sim.CurrentTime, 0), tick)
}

// Inputs.

func (r *run) doInput(sim *core.Simulation, in Input) {
	switch in.Kind {
	case "cast":
		r.pressCast(sim, in)
	case "useItem":
		r.pressItem(sim, in)
	case "startAttack":
		if !r.hasTarget {
			r.s.answer(in.ID, false, errNothingAttack)
			return
		}
		r.attackOn = true
		r.updateWorld(sim)
		r.s.answer(in.ID, true, "")
	case "stopAttack":
		r.attackOn = false
		r.updateWorld(sim)
		r.s.answer(in.ID, true, "")
	case "target":
		r.hasTarget = in.Target != nil && in.Target.Kind != "player"
		if r.hasTarget {
			r.char.CurrentTarget = r.dummy
		}
		r.updateWorld(sim)
		r.s.answer(in.ID, true, "")
	case "clearTarget":
		r.hasTarget = false
		r.attackOn = false
		r.updateWorld(sim)
		r.s.answer(in.ID, true, "")
	case "cancelCast":
		r.cancelQueued()
		if r.char.IsCasting(sim) {
			r.char.InterruptHardcast(sim)
			r.s.answer(in.ID, true, "")
		} else {
			r.s.answer(in.ID, false, "")
		}
	case "move":
		r.move(sim, in)
		r.s.answer(in.ID, true, "")
	case "setImbue":
		// The imbue an input older than the prepull window left behind (Session.rewindow).
		if r.shaman != nil {
			for _, imbue := range shamanImbues {
				if r.shaman.ImbueCastSpellID(imbue) == in.SpellID {
					r.shaman.SwapMainHandImbue(sim, imbue)
				}
			}
		}
	case "jump":
		r.updateWorld(sim)
		if !r.me.airborne(sim.CurrentTime) {
			r.me.airborneUntil = sim.CurrentTime + jumpTime
			r.interruptCast(sim)
		}
		r.s.answer(in.ID, true, "")
	default:
		r.s.answer(in.ID, false, "Unknown input")
	}
}

var shamanImbues = []proto.WeaponImbue{proto.WeaponImbue_RockbiterWeapon, proto.WeaponImbue_FlametongueWeapon, proto.WeaponImbue_FrostbrandWeapon, proto.WeaponImbue_WindfuryWeapon}

// isImbueCast is whether a spell id is one of the shaman's imbue casts.
func (r *run) isImbueCast(spellID int32) bool {
	if r.shaman == nil {
		return false
	}
	for _, imbue := range shamanImbues {
		if r.shaman.ImbueCastSpellID(imbue) == spellID {
			return true
		}
	}
	return false
}

// cancelQueued drops the queued cast. It never happens, with no error, as in game.
func (r *run) cancelQueued() {
	if r.queued != nil {
		r.queued.done = true
		r.queued = nil
		r.s.answer(r.queuedInput, false, "")
	}
}

func (r *run) interruptCast(sim *core.Simulation) {
	if r.char.IsCasting(sim) {
		r.char.InterruptHardcast(sim)
		if !r.quiet {
			r.s.batch.Errors = append(r.s.batch.Errors, "Interrupted")
		}
	}
}
