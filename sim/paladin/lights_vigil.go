package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Light's Vigil (the 31 point Holy talent, needs Holy Shock). From the client (1.60.1.70338):
// a 1.5 sec cast with a 6 sec cooldown that puts Light's Vigil on the target for 30 sec. The
// paladin's next Holy Shock on it triggers no cooldown, and the Vigil detonates for Holy damage
// (42.9% of spell power) and gives back 75% of Light's Vigil's mana. Rank 1 at 40 costs 730 and
// detonates for 182, rank 2 at 50 1000 and 278, rank 3 at 60 1340 and 395 (a 7.5% spread). The
// heal on an ally is not modelled.
//
// We assume a Holy Shock that misses still uses the Vigil up and skips its cooldown, and that
// the detonation and the refund need it to land.
var LightsVigilLevel = [...]int{0, 40, 50, 60}
var LightsVigilSpellId = [...]int32{0, 1310911, 1311590, 1311595}

func (paladin *Paladin) registerLightsVigil() {
	if !paladin.Talents.LightsVigil || !paladin.Talents.HolyShock {
		return
	}

	ranks := []struct {
		level    int32
		spellID  int32
		debuffID int32
		damageID int32
		manaCost float64
		damage   float64
	}{
		{level: 40, spellID: 1310911, debuffID: 1310909, damageID: 1310914, manaCost: 730, damage: 182},
		{level: 50, spellID: 1311590, debuffID: 1311593, damageID: 1311592, manaCost: 1000, damage: 278},
		{level: 60, spellID: 1311595, debuffID: 1311597, damageID: 1311598, manaCost: 1340, damage: 395},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 6,
	}
	refundMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: 1311601})

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		detonate := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.damageID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: 0.429,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, rank.damage*(0.9625+0.075*sim.RandomFloat("Light's Vigil")), spell.OutcomeMagicHitAndCrit)
			},
		})

		debuffs := paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return target.RegisterAura(core.Aura{
				Label:    "Light's Vigil" + paladin.Label + strconv.Itoa(i+1),
				ActionID: core.ActionID{SpellID: rank.debuffID},
				Duration: time.Second * 30,
			})
		})

		var vigil *core.Spell
		vigil = paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagAPL,

			RequiredLevel: int(rank.level),
			Rank:          i + 1,

			ManaCost: core.ManaCostOptions{
				FlatCost: rank.manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: time.Millisecond * 1500,
				},
				CD: cd,
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
				// One Vigil per paladin.
				for _, aura := range debuffs {
					if aura != nil {
						aura.Deactivate(sim)
					}
				}
				debuffs.Get(target).Activate(sim)
			},
		})

		paladin.lightsVigil = func(sim *core.Simulation, target *core.Unit, holyShock *core.Spell, landed bool) {
			aura := debuffs.Get(target)
			if !aura.IsActive() {
				return
			}
			aura.Deactivate(sim)
			holyShock.CD.Reset()
			if landed {
				detonate.Cast(sim, target)
				paladin.AddMana(sim, 0.75*vigil.DefaultCast.Cost, refundMetrics)
			}
		}
	}
}
