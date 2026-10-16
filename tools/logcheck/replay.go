package main

import (
	"runtime"
	"strings"
	"sync"
	"time"

	googleproto "google.golang.org/protobuf/proto"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// replayCast is one of our casts from the log, to repeat in the sim at the same time.
type replayCast struct {
	at      time.Duration // from the start of the window
	spellID int32
	name    string
}

// plan is what we replay against one mob level.
//
// The sim covers the whole window, from its start to the end of the last fight at this
// level. Our white swings are on only during those fights. Casts that dealt damage are
// replayed only when they went at a mob of this level, but buffs and totems are replayed every
// time, so the sim has them up when the log did.
type plan struct {
	level    int
	fights   []*fight
	duration time.Duration
	windows  [][2]time.Duration
	casts    []replayCast
	// hitsPerMinute is how often mobs hit us during these fights, spread over the whole sim.
	// It sets off our Lightning Shield the way the mobs did.
	hitsPerMinute float64
}

// replayStats counts what happened to the replayed casts in all iterations.
type replayStats struct {
	cast    int
	late    int              // cast, but only after a retry
	failed  map[string]int   // by spell, never cast in 2 sec of retries
	missing map[string]int32 // spells the sim doesn't know, by name
}

// recorder watches every spell result of one sim.
type recorder struct {
	char        *core.Character
	manaMetrics *core.ResourceMetrics
	tallies     *tallies
}

var (
	recorders sync.Map // *core.Environment to *recorder
	hookOnce  sync.Once
)

// installHook makes every sim we build record its spell results. The aura has to go on before
// the sim is finalized, so we add it from core.EnvHooks.
func installHook() {
	hookOnce.Do(func() {
		core.EnvHooks = append(core.EnvHooks, func(env *core.Environment) {
			char := env.Raid.Parties[0].Players[0].GetCharacter()
			rec := &recorder{char: char, tallies: newTallies()}
			rec.manaMetrics = char.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaGain, Tag: 9001})
			rec.watch(&char.Unit)
			for _, pet := range char.Pets {
				rec.watch(&pet.Unit)
			}
			recorders.Store(env, rec)
		})
	})
}

func (rec *recorder) watch(unit *core.Unit) {
	core.MakePermanent(unit.RegisterAura(core.Aura{
		Label: "logcheck recorder",
		OnSpellHitDealt: func(_ *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			rec.record(spell, result, false)
		},
		OnPeriodicDamageDealt: func(_ *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			rec.record(spell, result, true)
		},
	}))
}

func (rec *recorder) record(spell *core.Spell, result *core.SpellResult, periodic bool) {
	// A landed result with no damage is an aura going on (a debuff, a DoT's application),
	// which the client logs as an aura and not as damage.
	if result.Landed() && result.Damage == 0 {
		return
	}
	key := abilityKey{spell: spell.ActionID.SpellID, periodic: periodic}
	if spell.ActionID.OtherID == proto.OtherAction_OtherActionAttack {
		key.spell = 0
	}
	outcome := simOutcome(result.Outcome)
	physical := spell.SpellSchool == core.SpellSchoolPhysical
	raw, resisted := 0.0, 0.0
	if landed(outcome) {
		raw = result.RawDamage()
		if !physical {
			resisted = 1 - result.ResistanceMultiplier
		}
	}
	rec.tallies.add(int(result.Target.Level), key, outcome, raw, physical, resisted)
	if (outcome == "hit" || outcome == "crit") && !result.DidResist() {
		t := rec.tallies.get(int(result.Target.Level), key)
		t.hits = append(t.hits, landedHit{raw: raw, amount: result.Damage, crit: outcome == "crit"})
	}
}

func simOutcome(outcome core.HitOutcome) string {
	switch {
	case outcome.Matches(core.OutcomeCrit):
		return "crit"
	case outcome.Matches(core.OutcomeGlance):
		return "glance"
	case outcome.Matches(core.OutcomeCrush):
		return "crush"
	case outcome.Matches(core.OutcomeBlock):
		return "block"
	case outcome.Matches(core.OutcomeHit):
		return "hit"
	case outcome.Matches(core.OutcomeMiss):
		return "miss"
	case outcome.Matches(core.OutcomeDodge):
		return "dodge"
	case outcome.Matches(core.OutcomeParry):
		return "parry"
	}
	return "other"
}

// replay runs the plan iterations times, each with its own seed, and returns what the sim's
// attacks did in all of them together.
func replay(player *proto.Player, p *plan, iterations int) (*tallies, *replayStats) {
	installHook()
	total := newTallies()
	stats := &replayStats{failed: map[string]int{}, missing: map[string]int32{}}
	var mu sync.Mutex

	jobs := make(chan int)
	var wg sync.WaitGroup
	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for seed := range jobs {
				t, s := replayOnce(player, p, int64(seed))
				mu.Lock()
				total.merge(t)
				stats.cast += s.cast
				stats.late += s.late
				for name, n := range s.failed {
					stats.failed[name] += n
				}
				for name, id := range s.missing {
					stats.missing[name] = id
				}
				mu.Unlock()
			}
		}()
	}
	for i := range iterations {
		jobs <- i + 1
	}
	close(jobs)
	wg.Wait()
	return total, stats
}

func replayOnce(base *proto.Player, p *plan, seed int64) (*tallies, *replayStats) {
	player := googleproto.Clone(base).(*proto.Player)
	target := &proto.Target{
		Level:     int32(p.level),
		MobType:   proto.MobType_MobTypeUnknown,
		Stats:     make([]float64, len(core.DefaultTargetProtoLvl60.Stats)),
		TankIndex: -1,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		// The mobs' hits on us, at the logged rate, for the shield's procs. Their damage is a token
		// one: we only replay what we deal.
		Encounter: &proto.Encounter{Duration: p.duration.Seconds(), Targets: []*proto.Target{target},
			RaidDamageHitsPerMinute: p.hitsPerMinute, RaidDamagePerHit: 1},
		// The amount check needs each hit's unrounded damage.
		SimOptions: &proto.SimOptions{Ruleset: proto.Ruleset_RulesetForever, RandomSeed: seed, Iterations: 1, ExactDamage: true},
	}, simsignals.CreateSignals())
	value, _ := recorders.LoadAndDelete(sim.Environment)
	rec := value.(*recorder)
	char := rec.char
	stats := &replayStats{failed: map[string]int{}, missing: map[string]int32{}}

	sim.Reset()
	sim.PrePull()
	at := func(d time.Duration, fn func(*core.Simulation)) {
		sim.AddPendingAction(&core.PendingAction{NextActionAt: d, OnAction: fn})
	}

	// Combat start turns the swings on, so we turn them off right after it and back on for
	// each fight.
	at(time.Millisecond, func(sim *core.Simulation) { char.AutoAttacks.CancelAutoSwing(sim) })
	for _, w := range p.windows {
		at(max(w[0], 2*time.Millisecond), func(sim *core.Simulation) { char.AutoAttacks.EnableAutoSwing(sim) })
		at(w[1], func(sim *core.Simulation) { char.AutoAttacks.CancelAutoSwing(sim) })
	}

	for _, c := range p.casts {
		spell := findSpell(char, c.spellID)
		if spell == nil {
			stats.missing[c.name] = c.spellID
			continue
		}
		tries := 0
		var attempt func(sim *core.Simulation)
		attempt = func(sim *core.Simulation) {
			// The sim's GCD and mana don't follow the log's exactly, and neither is what we
			// check here. The empty rotation keeps the GCD busy in 50 ms steps, so we free
			// it, and we fill the mana so the cast never fails for lack of it.
			if !char.IsCasting(sim) {
				char.SetGCDTimer(sim, sim.CurrentTime)
				if char.MaxMana() > 0 && char.CurrentMana() < char.MaxMana() {
					char.AddMana(sim, char.MaxMana()-char.CurrentMana(), rec.manaMetrics)
				}
				if spell.Cast(sim, char.CurrentTarget) {
					stats.cast++
					if tries > 0 {
						stats.late++
					}
					return
				}
			}
			tries++
			if tries > 20 {
				stats.failed[c.name]++
				return
			}
			at(sim.CurrentTime+100*time.Millisecond, attempt)
		}
		at(max(c.at, 3*time.Millisecond), attempt)
	}

	for !sim.Step() {
	}
	rec.tallies.iterations = 1
	return rec.tallies, stats
}

// findSpell finds the sim's spell for a log spell ID, any tag.
func findSpell(char *core.Character, id int32) *core.Spell {
	for _, spell := range char.Spellbook {
		if spell.ActionID.SpellID == id {
			return spell
		}
	}
	return nil
}

// imbueFromName maps a weapon imbue's spell name to the sim's, "Flametongue Weapon" to
// FlametongueWeapon.
func imbueFromName(name string) (proto.WeaponImbue, bool) {
	value, ok := proto.WeaponImbue_value[strings.ReplaceAll(name, " ", "")]
	return proto.WeaponImbue(value), ok && value != 0
}
