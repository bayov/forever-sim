package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Light's Vigil (the 31 point Holy talent, needs Holy Shock) heals only since the launch realms'
// October 30 hotfix (spell 1310911 in 1.60.1.70753): "Applies Light's Vigil to a friendly target
// for 30 sec. The caster's next Holy Shock on the target incurs no cooldown and heals the target and
// their party for 704. Only one Vigil per Paladin may be active in each party." 1.5 sec cast, 730 /
// 1000 / 1340 mana at levels 40 / 50 / 60, a 6 sec cooldown the ranks share (category recovery).
// The mark (1310909, 1311593, 1311597) is one charge for 30 sec; the party heal (1310912, 1311591,
// 1311596) is LightsVigilHealRanks, 14.3% of healing power, on the target and every living player
// in its party (everyone is in reach, docs/healing_sim.md). The cast spell's second dummy effect
// still stores 75, the old 75% mana refund, but the hotfix took it out of the tooltip, so the sim
// gives nothing back (Calls for the user, docs/healing_paladin.md). Before the hotfix the Vigil
// went on an enemy and the next Holy Shock detonated it for Holy damage.
var LightsVigilLevel = [...]int{0, 40, 50, 60}
var LightsVigilSpellId = [...]int32{0, 1310911, 1311590, 1311595}

// LightsVigilDebuffId is the mark each rank puts on its target.
var LightsVigilDebuffId = [...]int32{0, 1310909, 1311593, 1311597}

func (paladin *Paladin) registerLightsVigil() {
	if !paladin.Talents.LightsVigil || !paladin.Talents.HolyShock {
		return
	}

	heals := make([]*core.Spell, len(LightsVigilHealRanks))
	for rank := 1; rank < len(LightsVigilHealRanks); rank++ {
		r := LightsVigilHealRanks[rank]
		if !paladin.KnowsRank(LightsVigilSpellId[rank], r.learnLevel) {
			continue
		}
		low, high := r.At(paladin.Level)
		heals[rank] = paladin.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_PaladinLightsVigilHeal,
			ActionID:    core.ActionID{SpellID: r.spellID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellHealing,
			Flags:       core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete,

			RequiredLevel: int(r.learnLevel),
			Rank:          rank,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: paladin.downrankCoefficient(r.coefficient, r.learnLevel),
			CritDamageBonus:  paladin.holyPowerCritBonus(),

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				crit := false
				for _, unit := range paladin.vigilTargets(target) {
					if unit.IsDead() {
						continue
					}
					crit = spell.CalcAndDealHealing(sim, unit, sim.Roll(low, high), spell.OutcomeHealingCrit).DidCrit() || crit
				}
				// Illumination once a detonation, on the Vigil's base cost (Calls for the user).
				if crit {
					paladin.illuminate(sim, r.manaCost)
				}
			},
		})
	}

	paladin.vigilRank = make([]int, len(paladin.Env.AllUnits))
	paladin.vigilMarks = paladin.NewRaidPlayerAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(core.Aura{
			Label:    "Light's Vigil-" + paladin.Label,
			ActionID: core.ActionID{SpellID: LightsVigilDebuffId[3]},
			Duration: time.Second * 30,
		})
	})
	paladin.vigilHeals = heals

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 6,
	}
	for rank := 1; rank < len(LightsVigilSpellId); rank++ {
		rank := rank
		r := LightsVigilHealRanks[rank]
		if heals[rank] == nil {
			continue
		}
		paladin.RegisterSpell(core.SpellConfig{
			SpellCode:   SpellCode_PaladinLightsVigil,
			ActionID:    core.ActionID{SpellID: LightsVigilSpellId[rank]},
			SpellSchool: core.SpellSchoolHoly,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagHelpful | core.SpellFlagAPL,

			RequiredLevel: int(r.learnLevel),
			Rank:          rank,

			ManaCost: core.ManaCostOptions{
				FlatCost: r.manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: r.castTime,
				},
				CD: cd,
			},
			PushbackReduction: paladin.spiritualFocusChance(),

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				target = paladin.healTarget(target)
				if target.IsDead() {
					return
				}
				// One Vigil a paladin in each party: ours on another member of the target's party goes.
				for _, unit := range paladin.vigilTargets(target) {
					paladin.vigilMarks.Get(unit).Deactivate(sim)
				}
				paladin.vigilRank[target.UnitIndex] = rank
				paladin.vigilMarks.Get(target).Activate(sim)
			},
		})
	}
}

// vigilTargets is the target of a Light's Vigil and the players of its party.
func (paladin *Paladin) vigilTargets(target *core.Unit) []*core.Unit {
	party := paladin.Env.Raid.GetPlayerParty(target)
	units := []*core.Unit{target}
	for _, agent := range party.Players {
		if unit := &agent.GetCharacter().Unit; unit != target {
			units = append(units, unit)
		}
	}
	return units
}

// HasLightsVigil is whether our Light's Vigil is on the unit.
func (paladin *Paladin) HasLightsVigil(unit *core.Unit) bool {
	return paladin.vigilMarks != nil && unit.Type != core.EnemyUnit && paladin.vigilMarks.Get(unit).IsActive()
}

// detonateLightsVigil is a Holy Shock (spell) landing on an ally with our Vigil on it: the mark
// goes, the Holy Shock incurs no cooldown, and the Vigil heals the target and its party.
func (paladin *Paladin) detonateLightsVigil(sim *core.Simulation, target *core.Unit, holyShock *core.Spell) {
	if !paladin.HasLightsVigil(target) {
		return
	}
	paladin.vigilMarks.Get(target).Deactivate(sim)
	holyShock.CD.Reset()
	heal := paladin.vigilHeals[paladin.vigilRank[target.UnitIndex]]
	heal.Cast(sim, target)
}
