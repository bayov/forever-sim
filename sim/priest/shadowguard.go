package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

const ShadowguardRanks = 6

// Shadowguard is the Troll priest's self buff: 3 charges for 10 min, and a direct hit taken
// (melee, ranged or a harmful spell, AoE too, but not a periodic tick: proc flags 0x222a8)
// strikes the attacker for Shadow damage, at most once every 3.5 sec. A missed strike still
// uses its charge. Forever only lets it be cast in Shadowform (StanceMask 0x8000000).
//
// The buff row states more damage (116 at rank 6) than the tooltip, which prices the strike off
// the triggered row (96 at rank 6, 28382); that row is dealt, at 26.7% of spell power. The strike
// can't crit (Attr2 0x20000000), causes no threat (the tooltip), and flies at speed 20. Rows from
// the 1.60.1.70532 client (wowsims/forever's spell store, shadowguard.go and fe7af8b9b9).
var ShadowguardSpellId = [ShadowguardRanks + 1]int32{0, 18137, 19308, 19309, 19310, 19311, 19312}
var ShadowguardStrikeSpellId = [ShadowguardRanks + 1]int32{0, 28377, 28378, 28379, 28380, 28381, 28382}
var ShadowguardStrikeDamage = [ShadowguardRanks + 1]float64{0, 13, 22, 33, 50, 70, 96}
var ShadowguardManaCost = [ShadowguardRanks + 1]float64{0, 50, 85, 120, 160, 200, 250}
var ShadowguardLevel = [ShadowguardRanks + 1]int{0, 20, 28, 36, 44, 52, 60}

const ShadowguardStrikeCoef = 0.267

func (priest *Priest) registerShadowguardSpell() {
	if priest.Race != proto.Race_RaceTroll || !priest.Env.IsForever() {
		return
	}

	priest.Shadowguard = make([]*core.Spell, ShadowguardRanks+1)
	priest.ShadowguardAuras = make([]*core.Aura, ShadowguardRanks+1)
	icd := core.Cooldown{
		Timer:    priest.NewTimer(),
		Duration: time.Millisecond * 3500,
	}

	for rank := 1; rank <= ShadowguardRanks; rank++ {
		if ShadowguardLevel[rank] <= int(priest.Level) {
			priest.registerShadowguardRank(rank, icd)
		}
	}
}

func (priest *Priest) registerShadowguardRank(rank int, icd core.Cooldown) {
	strikeDamage := ShadowguardStrikeDamage[rank]

	strike := priest.RegisterSpell(core.SpellConfig{
		SpellCode:    SpellCode_PriestShadowguard,
		ActionID:     core.ActionID{SpellID: ShadowguardStrikeSpellId[rank]},
		SpellSchool:  core.SpellSchoolShadow,
		DefenseType:  core.DefenseTypeMagic,
		ProcMask:     core.ProcMaskSpellDamageProc,
		Flags:        SpellFlagPriest | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
		MissileSpeed: 20,

		DamageMultiplier: 1,
		ThreatMultiplier: 0,
		BonusCoefficient: ShadowguardStrikeCoef,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, strikeDamage, spell.OutcomeMagicHit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})

	aura := priest.RegisterAura(core.Aura{
		Label:     fmt.Sprintf("Shadowguard (Rank %d)", rank),
		ActionID:  core.ActionID{SpellID: ShadowguardSpellId[rank]},
		Duration:  time.Minute * 10,
		MaxStacks: 3,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, aura.MaxStacks)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.Unit.IsOpponent(&priest.Unit) || !icd.IsReady(sim) {
				return
			}
			icd.Use(sim)
			strike.Cast(sim, spell.Unit)
			aura.RemoveStack(sim)
		},
	})

	priest.ShadowguardAuras[rank] = aura

	priest.Shadowguard[rank] = priest.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_PriestShadowguard,
		ActionID:    core.ActionID{SpellID: ShadowguardSpellId[rank]},
		SpellSchool: core.SpellSchoolShadow,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       SpellFlagPriest | core.SpellFlagAPL,

		RequiredLevel: ShadowguardLevel[rank],
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: ShadowguardManaCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return priest.ShadowformAura != nil && priest.ShadowformAura.IsActive()
		},
		ExtraCastConditionNeedsEvent: true,

		// A recast replaces any rank's buff and refills the charges.
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			for _, other := range priest.ShadowguardAuras {
				if other != nil && other != aura {
					other.Deactivate(sim)
				}
			}
			aura.Activate(sim)
			aura.SetStacks(sim, aura.MaxStacks)
		},
	})
}
