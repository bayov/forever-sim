package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
)

const PenanceRanks = 4

// Penance is Discipline's new 21 point talent. It fires a Holy bolt at once and one a
// second for 2 sec, 3 bolts with a hit roll each, on a 12 sec cooldown. The bolt damage at
// each rank's cap, growth per level and the 19% coefficient are from the client
// (1.60.1.70338, wago.tools: 402284, 1240727, 1240730 and 1316993). Each bolt flies at the
// rows' speed 40 (wowsims/forever fe7af8b9b9).
//
// The channel is the cast, so it is what rolls on-cast procs such as Vestments of the Devout's
// 5 pieces; the bolts don't, or one Penance would roll three times (wowsims/forever gives the
// channel the spell damage proc mask, penance.go).
var PenanceSpellId = [PenanceRanks + 1]int32{0, 402174, 1240720, 1240721, 1316995}
var PenanceBoltSpellId = [PenanceRanks + 1]int32{0, 402284, 1240727, 1240730, 1316993}
var PenanceBoltDamage = [PenanceRanks + 1]float64{0, 44.61, 52.97, 71.87, 92}
var PenanceScaling = [PenanceRanks + 1]core.RankScaling{{}, {39, 0.29}, {49, 0.33}, {59, 0.43}, {}}
var PenanceManaCost = [PenanceRanks + 1]float64{0, 150, 220, 270, 385}
var PenanceLevel = [PenanceRanks + 1]int{0, 30, 40, 50, 60}

func (priest *Priest) registerPenanceSpell() {
	if !priest.Talents.Penance {
		return
	}

	priest.Penance = make([]*core.Spell, PenanceRanks+1)
	cdTimer := priest.NewTimer()

	for rank := 1; rank <= PenanceRanks; rank++ {
		if PenanceLevel[rank] <= int(priest.Level) {
			priest.Penance[rank] = priest.registerPenanceRank(rank, cdTimer)
		}
	}
}

func (priest *Priest) registerPenanceRank(rank int, cdTimer *core.Timer) *core.Spell {
	boltDamage := PenanceScaling[rank].At(PenanceBoltDamage[rank], priest.Level)

	bolt := priest.RegisterSpell(core.SpellConfig{
		SpellCode:    SpellCode_PriestPenance,
		ActionID:     core.ActionID{SpellID: PenanceBoltSpellId[rank]},
		SpellSchool:  core.SpellSchoolHoly,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagPriest | core.SpellFlagNoOnCastComplete,
		MissileSpeed: 40,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.19,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			multiplier := priest.PowerInLightMultiplier(target)
			spell.DamageMultiplier *= multiplier
			result := spell.CalcDamage(sim, target, boltDamage, spell.OutcomeMagicHitAndCrit)
			spell.DamageMultiplier /= multiplier
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})

	return priest.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: PenanceSpellId[rank]},
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagPriest | core.SpellFlagChanneled | core.SpellFlagAPL,

		RequiredLevel: PenanceLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: PenanceManaCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second * 12,
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("Penance-%d", rank),
			},
			NumberOfTicks: 2,
			TickLength:    time.Second,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				bolt.Cast(sim, target)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			bolt.Cast(sim, target)
			spell.Dot(target).Apply(sim)
		},
	})
}

// PowerInLightMultiplier is Power in Light's 2% a rank on Smite and Penance against a
// target with the priest's Holy Fire or Shadow Word: Pain on it (15 October beta notes: "Power in
// Light now increases Smite and Penance if the target is afflicted by Holy Fire or Shadow Word
// Pain"; client 1.60.1.70425 tooltip of 1309969: "your Holy Fire or Shadow Word: Pain").
func (priest *Priest) PowerInLightMultiplier(target *core.Unit) float64 {
	if priest.Talents.PowerInLight == 0 {
		return 1
	}
	multiplier := 1 + 0.02*float64(priest.Talents.PowerInLight)
	for _, spells := range [][]*core.Spell{priest.HolyFire, priest.ShadowWordPain} {
		for _, spell := range spells {
			if spell != nil && spell.Dot(target) != nil && spell.Dot(target).IsActive() {
				return multiplier
			}
		}
	}
	return 1
}
