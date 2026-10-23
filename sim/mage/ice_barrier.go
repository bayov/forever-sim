package mage

import (
	"fmt"
	"math"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

const IceBarrierRanks = 4

var IceBarrierSpellId = [IceBarrierRanks + 1]int32{0, 11426, 13031, 13032, 13033}
var IceBarrierManaCost = [IceBarrierRanks + 1]float64{0, 305, 360, 420, 480}
var IceBarrierLevel = [IceBarrierRanks + 1]int{0, 40, 46, 52, 58}

// The absorb's base points, growth per level and the level the growth stops at (effect 1 of each
// rank in the 70450 client).
var IceBarrierAbsorb = [IceBarrierRanks + 1]float64{0, 431, 542, 671, 811}
var IceBarrierAbsorbPerLevel = [IceBarrierRanks + 1]float64{0, 2.8, 3.2, 3.6, 4}
var IceBarrierMaxLevel = [IceBarrierRanks + 1]int{0, 46, 52, 58, 64}

const FrigidPowerSpellId = 1326891

// Ice Barrier since the 2026-10-20 hotfix to 70450 (wago.tools hotfix table, Spell and
// SpellEffect rows of 11426-13033, 1326891 and 11189; wowsims/forever 983ea6fcc1):
//
// "Shield yourself, absorbing ${$s1+($SPFR*0.1)*($11189s3/100+1)} damage and increasing your Frost
// damage by $1326891s1% for $d. While shielded, your Frost damage is increased by an additional $s2%
// and you do not lose casting time when you take damage."
//
// The shield's effects 2 and 3 raise Frost damage by 3% while it holds, and every cast also gives
// Frigid Power, 3% more for the full minute however soon the shield breaks. Both take the mask
// Piercing Ice had, every Frost spell Frostfire Bolt included. Under Forever each is its own
// damage factor, as every percent damage row is (wowsims/forever fef2bcaebd). The absorb takes
// 10% of Frost spell power, and Frost Warding's effect 3 (10% a point) raises the whole of it.
// The only damage a DPS mage takes is the encounter's raid hits (raid_hits.go), so those are what
// break the shield. A hit the shield takes whole does no damage and so pushes back no cast; the
// hit that breaks it pushes back as any other does.
func (mage *Mage) registerIceBarrierSpell() {
	if !mage.Talents.IceBarrier {
		return
	}

	mage.IceBarrier = make([]*core.Spell, IceBarrierRanks+1)
	mage.IceBarrierAuras = make([]*core.Aura, IceBarrierRanks+1)

	frostSpells := []*core.Spell{}
	mage.onMageSpellOfSchool(core.SpellSchoolFrost, func(spell *core.Spell) {
		frostSpells = append(frostSpells, spell)
	})
	frostDamage := func(gain bool) {
		for _, spell := range frostSpells {
			mage.changeDamagePercent(spell, 0.03, gain)
		}
	}

	mage.FrigidPowerAura = mage.RegisterAura(core.Aura{
		ActionID: core.ActionID{SpellID: FrigidPowerSpellId},
		Label:    "Frigid Power",
		Duration: time.Minute,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			frostDamage(true)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			frostDamage(false)
		},
	})

	// What is left of the shield up now.
	var shieldStrength float64

	cdTimer := mage.NewTimer()
	for rank := 1; rank <= IceBarrierRanks; rank++ {
		if IceBarrierLevel[rank] > int(mage.Level) {
			continue
		}
		mage.IceBarrierAuras[rank] = mage.RegisterAura(core.Aura{
			ActionID: core.ActionID{SpellID: IceBarrierSpellId[rank]},
			Label:    fmt.Sprintf("Ice Barrier (Rank %d)", rank),
			Duration: time.Minute,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				frostDamage(true)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				frostDamage(false)
				shieldStrength = 0
				if mage.activeBarrier == aura {
					mage.activeBarrier = nil
				}
			},
		})
		mage.IceBarrier[rank] = mage.GetOrRegisterSpell(mage.newIceBarrierSpellConfig(rank, cdTimer, &shieldStrength))
	}

	mage.AddDynamicDamageTakenModifier(func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		shield := mage.activeBarrier
		if shield == nil || !shield.IsActive() || result.Damage <= 0 {
			return
		}
		absorbed := min(shieldStrength, result.Damage)
		result.Damage -= absorbed
		shieldStrength -= absorbed
		if sim.Log != nil {
			mage.Log(sim, "%s absorbed %.1f damage, %.1f left", shield.Label, absorbed, shieldStrength)
		}
		if shieldStrength <= 0 {
			shield.Deactivate(sim)
		}
	})
}

// IceBarrierAbsorbAt is the rank's absorb for a mage of this level, Frost spell power and Frost
// Warding points.
func IceBarrierAbsorbAt(rank int, level int32, frostSpellPower float64, frostWarding int32) float64 {
	levels := min(int(level), IceBarrierMaxLevel[rank]) - IceBarrierLevel[rank]
	base := math.Floor(IceBarrierAbsorb[rank] + IceBarrierAbsorbPerLevel[rank]*float64(max(0, levels)))
	return (base + 0.1*frostSpellPower) * (1 + 0.1*float64(frostWarding))
}

// The spell power a Frost spell reads ($SPFR in a tooltip).
func (mage *Mage) frostSpellPower() float64 {
	return mage.GetStat(stats.SpellPower) + mage.GetStat(stats.SpellDamage) + mage.GetStat(stats.FrostPower)
}

func (mage *Mage) newIceBarrierSpellConfig(rank int, cdTimer *core.Timer, shieldStrength *float64) core.SpellConfig {
	spellID := IceBarrierSpellId[rank]
	shield := mage.IceBarrierAuras[rank]

	return core.SpellConfig{
		ActionID:      core.ActionID{SpellID: spellID},
		SpellSchool:   core.SpellSchoolFrost,
		Flags:         core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
		Rank:          rank,
		RequiredLevel: IceBarrierLevel[rank],
		ManaCost: core.ManaCostOptions{
			FlatCost: IceBarrierManaCost[rank],
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second * 30,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			// A new shield replaces the one up, whatever its rank.
			if mage.activeBarrier != nil && mage.activeBarrier != shield {
				mage.activeBarrier.Deactivate(sim)
			}
			shield.Activate(sim)
			mage.activeBarrier = shield
			*shieldStrength = IceBarrierAbsorbAt(rank, mage.Level, mage.frostSpellPower(), mage.Talents.FrostWarding)
			mage.FrigidPowerAura.Activate(sim)
		},
	}
}
