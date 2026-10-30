package core

import (
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's camp buffs (docs/raid_sim.md, Camp buffs). Sitting at a campfire gives a 1-hour buff for
// each camp feature placed there, at the values of client 1.60.1.70685 (the camp benefits tooltip,
// 1229741). Each one is "mutually exclusive with" a class buff, and at 60 the class buff is the
// stronger, so a player who has the class buff gets nothing from the camp's.
//
// Most class buffs are given for the whole fight, so the check is made once here. Strength of Earth
// can also come from a cast totem, so the Sharpening Wheel shares an exclusive effect with the totem's
// aura instead, and the totem wins while it stands.

const sharpeningWheelCategory = "SharpeningWheel"

func applyCampBuffs(character *Character, raidBuffs *proto.RaidBuffs, individualBuffs *proto.IndividualBuffs) {
	camp := raidBuffs.CampBuffs
	if camp == nil {
		return
	}
	// The unverified rule that the camp buffs stack with the group versions of their class buffs (the
	// tooltips name only the single-target spells): then the Greater Blessings, Arcane Brilliance, Prayer of
	// Spirit, Gift of the Wild and Prayer of Fortitude turn nothing off. Moonkin Aura and Strength of Earth
	// have no group version, so they still do.
	if camp.StackWithGroupBuffs {
		raidBuffs = &proto.RaidBuffs{MoonkinAura: raidBuffs.MoonkinAura, StrengthOfEarthTotem: raidBuffs.StrengthOfEarthTotem}
		individualBuffs = &proto.IndividualBuffs{}
	}

	if camp.FishBowl && !individualBuffs.BlessingOfKings {
		campMultiplyStatsAura(character, "Fish Bowl", 1230098, 1.08)
	}
	if camp.Lodestone && individualBuffs.BlessingOfMight == proto.TristateEffect_TristateEffectMissing {
		campStatsAura(character, "Lodestone", 1230164, stats.Stats{stats.AttackPower: 90})
	}
	if camp.SharpeningWheel && raidBuffs.StrengthOfEarthTotem == proto.TristateEffect_TristateEffectMissing {
		sharpeningWheelAura(character)
	}
	if camp.CampChair && !raidBuffs.MoonkinAura {
		campStatsAura(character, "Camp Chair", 1229519, stats.Stats{
			stats.MeleeCrit: 2 * CritRatingPerCritChance,
			stats.SpellCrit: 2 * SpellCritRatingPerCritChance,
		})
	} else if camp.CampChair && camp.CampChairSpellCritOnly {
		// The unverified reading where Moonkin Aura takes only the chair's spell crit.
		campStatsAura(character, "Camp Chair", 1229519, stats.Stats{stats.MeleeCrit: 2 * CritRatingPerCritChance})
	}
	if camp.IncenseCandle && !raidBuffs.ArcaneBrilliance {
		campStatsAura(character, "Incense Candle", 1229513, stats.Stats{stats.Intellect: 25})
	}
	if camp.FactionBanner && !raidBuffs.DivineSpirit {
		campStatsAura(character, "Faction Banner", 1229718, stats.Stats{stats.Spirit: 32})
	}
	if camp.ManaWell && individualBuffs.BlessingOfWisdom == proto.TristateEffect_TristateEffectMissing {
		campStatsAura(character, "Mana Well", 1230587, stats.Stats{stats.MP5: 29})
	}
	if camp.EnchantedLute && raidBuffs.GiftOfTheWild == proto.TristateEffect_TristateEffectMissing {
		campStatsAura(character, "Enchanted Lute", 1230653, stats.Stats{
			stats.BonusArmor:       308,
			stats.Stamina:          13,
			stats.Agility:          13,
			stats.Strength:         13,
			stats.Intellect:        13,
			stats.Spirit:           13,
			stats.ArcaneResistance: 22,
			stats.FireResistance:   22,
			stats.FrostResistance:  22,
			stats.NatureResistance: 22,
			stats.ShadowResistance: 22,
		})
	}
	if camp.FirstAidKit && raidBuffs.PowerWordFortitude == proto.TristateEffect_TristateEffectMissing {
		campStatsAura(character, "First Aid Kit", 1230124, stats.Stats{stats.Stamina: 56})
	}
}

func campStatsAura(character *Character, label string, spellID int32, bonus stats.Stats) *Aura {
	return MakePermanent(character.RegisterAura(Aura{
		Label:      label,
		ActionID:   ActionID{SpellID: spellID},
		BuildPhase: CharacterBuildPhaseBuffs,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.AddBuildPhaseStatsDynamic(sim, bonus)
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.AddBuildPhaseStatsDynamic(sim, bonus.Multiply(-1))
		},
	}))
}

// The Fish Bowl's 8% to every attribute, as Blessing of Kings gives its 10%.
func campMultiplyStatsAura(character *Character, label string, spellID int32, multiplier float64) *Aura {
	statDeps := []*stats.StatDependency{
		character.NewDynamicMultiplyStat(stats.Stamina, multiplier),
		character.NewDynamicMultiplyStat(stats.Agility, multiplier),
		character.NewDynamicMultiplyStat(stats.Strength, multiplier),
		character.NewDynamicMultiplyStat(stats.Intellect, multiplier),
		character.NewDynamicMultiplyStat(stats.Spirit, multiplier),
	}
	return MakePermanent(character.RegisterAura(Aura{
		Label:      label,
		ActionID:   ActionID{SpellID: spellID},
		BuildPhase: CharacterBuildPhaseBuffs,
		OnGain: func(aura *Aura, sim *Simulation) {
			for _, dep := range statDeps {
				if aura.Unit.Env.MeasuringStats && aura.Unit.Env.State != Finalized {
					aura.Unit.StatDependencyManager.EnableDynamicStatDep(dep)
				} else {
					aura.Unit.EnableDynamicStatDep(sim, dep)
				}
			}
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, dep := range statDeps {
				if aura.Unit.Env.MeasuringStats && aura.Unit.Env.State != Finalized {
					aura.Unit.StatDependencyManager.DisableDynamicStatDep(dep)
				} else {
					aura.Unit.DisableDynamicStatDep(sim, dep)
				}
			}
		},
	}))
}

// The Sharpening Wheel's Strength counts only while no Strength of Earth Totem reaches the player.
func sharpeningWheelAura(character *Character) {
	bonus := stats.Stats{stats.Strength: 34}
	wheel := MakePermanent(character.RegisterAura(Aura{
		Label:      "Sharpening Wheel",
		ActionID:   ActionID{SpellID: 1230172},
		BuildPhase: CharacterBuildPhaseBuffs,
	}))
	wheel.NewExclusiveEffect(sharpeningWheelCategory, false, ExclusiveEffect{
		Priority: 1,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddBuildPhaseStatsDynamic(sim, bonus)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddBuildPhaseStatsDynamic(sim, bonus.Multiply(-1))
		},
	})
	// The totem's own aura holds the stronger effect in the same category, with nothing in it: while it's
	// up, the wheel's Strength is off.
	soe := StrengthOfEarthTotemAura(&character.Unit, 1)
	soe.NewExclusiveEffect(sharpeningWheelCategory, false, ExclusiveEffect{Priority: 2})
	// Nobody gives this player Strength of Earth (else the wheel isn't here), so the aura is only a cast
	// totem's: keep it out of the stats panel's build phase, which would turn it on.
	soe.BuildPhase = CharacterBuildPhaseNone
}
