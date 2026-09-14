package core

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

///////////////////////////////////////////////////////////////////////////
//                            Weapon Specialization Auras
///////////////////////////////////////////////////////////////////////////

func (character *Character) SwordSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Sword Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.SwordsSkill += 5
			character.PseudoStats.TwoHandedSwordsSkill += 5
		},
	})
}

func (character *Character) AxeSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Axe Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.AxesSkill += 5
			character.PseudoStats.TwoHandedAxesSkill += 5
		},
	})
}

func (character *Character) MaceSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Mace Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.MacesSkill += 5
			character.PseudoStats.TwoHandedMacesSkill += 5
		},
	})
}

func (character *Character) DaggerSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Dagger Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.DaggersSkill += 5
		},
	})
}

func (character *Character) FistWeaponSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Fist Weapon Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.UnarmedSkill += 5
		},
	})
}

func (character *Character) PoleWeaponSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Pole Weapon Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.StavesSkill += 5
			character.PseudoStats.PolearmsSkill += 5
		},
	})
}

func (character *Character) GunSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Gun Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.GunsSkill += 5
		},
	})
}

func (character *Character) BowSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Bow Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.BowsSkill += 5
		},
	})
}

func (character *Character) CrossbowSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Crossbow Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.CrossbowsSkill += 5
		},
	})
}

func (character *Character) ThrownSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Thrown Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.ThrownSkill += 5
		},
	})
}

func (character *Character) FeralCombatSpecializationAura() *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:      "Feral Combat Skill Specialization",
		BuildPhase: CharacterBuildPhaseGear,
		Duration:   NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.FeralCombatSkill += 5
		},
	})
}

// Under Forever the weapon skill racials pay critical strike instead, and only while the
// matching weapon is held. The tooltips read "spell and ability critical strike chance",
// so it lands on both pools. Equipment cannot change mid-sim, so this is a flat add at
// build time rather than an aura.
func (character *Character) AddWeaponSpecializationCrit(critPercent float64, types ...proto.WeaponType) {
	if !character.hasWeaponOfType(types...) {
		return
	}

	character.AddStat(stats.MeleeCrit, critPercent*CritRatingPerCritChance)
	character.AddStat(stats.SpellCrit, critPercent*SpellCritRatingPerCritChance)
}

func (character *Character) hasWeaponOfType(types ...proto.WeaponType) bool {
	for _, weapon := range []*Item{character.MainHand(), character.OffHand()} {
		for _, weaponType := range types {
			if weapon.WeaponType == weaponType {
				return true
			}
		}
	}
	return false
}
