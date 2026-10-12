package warlock

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Forever splits the warlock's curses in two. Bane of Agony, Bane of Doom and Bane of Havoc
// are banes, and only one bane per warlock can be on a target. Curse of the Elements,
// Recklessness, Weakness, Tongues and Exhaustion stay curses, one per warlock. A target can
// have one of each (the client's tooltips: "Only one Bane per Warlock can be active on any
// one target"). Curse of Shadow is gone, Curse of the Elements covers every magic school.

// activateBane puts a bane on the target and takes off the warlock's other bane.
func (warlock *Warlock) activateBane(sim *core.Simulation, target *core.Unit, aura *core.Aura) {
	if active := warlock.ActiveBaneAura.Get(target); active != nil && active != aura {
		active.Deactivate(sim)
	}
	warlock.ActiveBaneAura[target.UnitIndex] = aura
}

// activateCurse does the same for curses.
func (warlock *Warlock) activateCurse(sim *core.Simulation, target *core.Unit, aura *core.Aura) {
	if active := warlock.ActiveCurseAura.Get(target); active != nil && active != aura {
		active.Deactivate(sim)
	}
	warlock.ActiveCurseAura[target.UnitIndex] = aura
}

func (warlock *Warlock) getBaneOfAgonyBaseConfig(rank int) core.SpellConfig {
	data := baneOfAgonyRanks[rank]
	numTicks := int32(12)
	tickLength := time.Second * 2

	// The client's points are the average tick.
	baseDamage := data.damageAt(warlock.Level)
	snapshotBaseDmgNoBonus := 0.0

	return core.SpellConfig{
		SpellCode:     SpellCode_WarlockBaneOfAgony,
		ActionID:      core.ActionID{SpellID: data.SpellID},
		SpellSchool:   core.SpellSchoolShadow,
		DefenseType:   core.DefenseTypeMagic,
		Flags:         core.SpellFlagAPL | core.SpellFlagResetAttackSwing | core.SpellFlagPureDot | WarlockFlagAffliction,
		ProcMask:      core.ProcMaskSpellDamage,
		RequiredLevel: data.Level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: data.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "BaneofAgony-" + warlock.Label + strconv.Itoa(rank),
			},
			NumberOfTicks:    numTicks,
			TickLength:       tickLength,
			BonusCoefficient: data.Coef,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				baseDmg := baseDamage

				if warlock.AmplifyCurseAura != nil && warlock.AmplifyCurseAura.IsActive() {
					baseDmg *= 1.5
					warlock.AmplifyCurseAura.Deactivate(sim)
				}

				// The bane starts at half its average tick, the spell power part is not
				// ramped. Every 4 ticks the base damage is added again, so the last 4 ticks
				// do 150% of it.
				snapshotBaseDmgNoBonus = baseDmg * 0.5

				dot.Snapshot(target, snapshotBaseDmgNoBonus, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				if dot.TickCount%4 == 0 {
					dot.SnapshotBaseDamage += snapshotBaseDmgNoBonus
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				dot := spell.Dot(target)
				warlock.activateBane(sim, target, dot.Aura)
				dot.Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},
	}
}

func (warlock *Warlock) registerBaneOfAgonySpell() {
	warlock.BaneOfAgony = make([]*core.Spell, 0)
	for rank := 1; rank < len(baneOfAgonyRanks); rank++ {
		config := warlock.getBaneOfAgonyBaseConfig(rank)

		if config.RequiredLevel <= int(warlock.Level) {
			warlock.BaneOfAgony = append(warlock.BaneOfAgony, warlock.GetOrRegisterSpell(config))
		}
	}
}

// Curse of Recklessness under Forever only takes armor (505 at rank 4, client 11717) and
// shares the minor armor slot with Faerie Fire, see core.CurseOfRecklessnessAura.
var curseOfRecklessnessRanks = []struct {
	spellID int32
	level   int
	cost    float64
}{{}, {704, 14, 35}, {7658, 28, 60}, {7659, 42, 90}, {11717, 56, 115}}

func (warlock *Warlock) registerCurseOfRecklessnessSpell() {
	rank := 0
	for r := 1; r < len(curseOfRecklessnessRanks); r++ {
		if curseOfRecklessnessRanks[r].level <= int(warlock.Level) {
			rank = r
		}
	}
	if rank == 0 {
		return
	}
	data := curseOfRecklessnessRanks[rank]

	warlock.CurseOfRecklessnessAuras = warlock.NewEnemyAuraArray(core.CurseOfRecklessnessAura)

	warlock.CurseOfRecklessness = warlock.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: data.spellID},
		SpellSchool: core.SpellSchoolShadow,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL | WarlockFlagAffliction,
		Rank:        rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: data.cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  156,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				aura := warlock.CurseOfRecklessnessAuras.Get(target)
				warlock.activateCurse(sim, target, aura)
				aura.Activate(sim)
			}
		},

		RelatedAuras: []core.AuraArray{warlock.CurseOfRecklessnessAuras},
	})
}

// Curse of the Elements is a trainer spell from 20 under Forever, 4 ranks (440892, 1311676,
// 1311677, 1311680 in the client): 30 / 45 / 60 / 75 resistance and 4 / 6 / 8 / 10% more
// damage taken from every magic school, for 50 / 100 / 150 / 200 mana.
var curseOfElementsSpellIDs = []int32{0, 440892, 1311676, 1311677, 1311680}
var curseOfElementsCost = []float64{0, 50, 100, 150, 200}

func (warlock *Warlock) registerCurseOfElementsSpell() {
	rank := core.ForeverCurseOfElementsRankAt(warlock.Level)
	if rank == 0 {
		return
	}

	warlock.CurseOfElementsAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ForeverCurseOfElementsAura(target, rank)
	})

	warlock.CurseOfElements = warlock.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: curseOfElementsSpellIDs[rank]},
		SpellSchool: core.SpellSchoolShadow,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL | WarlockFlagAffliction,
		Rank:        rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: curseOfElementsCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  156,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				aura := warlock.CurseOfElementsAuras.Get(target)
				warlock.activateCurse(sim, target, aura)
				aura.Activate(sim)
			}
		},

		RelatedAuras: []core.AuraArray{warlock.CurseOfElementsAuras},
	})
}

func (warlock *Warlock) registerAmplifyCurseSpell() {
	if !warlock.Talents.AmplifyCurse {
		return
	}

	actionID := core.ActionID{SpellID: 18288}

	warlock.AmplifyCurseAura = warlock.GetOrRegisterAura(core.Aura{
		Label:    "Amplify Curse",
		ActionID: actionID,
		Duration: time.Second * 30,
	})

	warlock.AmplifyCurse = warlock.GetOrRegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolShadow,
		Flags:       core.SpellFlagAPL | WarlockFlagAffliction,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: 3 * time.Minute,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			warlock.AmplifyCurseAura.Activate(sim)
		},
	})
}

// Bane of Doom (603): 1742 Shadow damage after 1 min, with a 400% spell power coefficient
// (Classic's Curse of Doom was 3200 and 100%). 300 mana, 1 min cooldown, learned at 60.
func (warlock *Warlock) registerBaneOfDoomSpell() {
	if warlock.Level < 60 {
		return
	}

	warlock.BaneOfDoom = warlock.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_WarlockBaneOfDoom,
		ActionID:    core.ActionID{SpellID: 603},
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagAPL | core.SpellFlagPureDot | WarlockFlagAffliction,

		RequiredLevel: 60,

		ManaCost: core.ManaCostOptions{
			FlatCost: 300,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: time.Second * 60,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		FlatThreatBonus:  160,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "BaneofDoom",
			},
			NumberOfTicks:    1,
			TickLength:       time.Minute,
			BonusCoefficient: 4,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, 1742, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				dot := spell.Dot(target)
				warlock.activateBane(sim, target, dot.Aura)
				dot.Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})
}
