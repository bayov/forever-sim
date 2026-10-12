package main

import (
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	templates["hunter_ranged"] = hunterRanged{}
}

// The level 60 hunter spells the templates name.
var (
	hunterAimedShot        = spellID(20904)
	hunterArcaneShot       = spellID(14287)
	hunterMultiShot        = spellID(2643)
	hunterSerpentSting     = spellID(25295)
	hunterSummonHawk       = spellID(1293527)
	hunterSniperShot       = spellID(1310786)
	hunterRapidFire        = spellID(3045)
	hunterIntimidation     = spellID(19577)
	hunterBestialWrath     = spellID(19574)
	hunterAspectOfTheHawk  = spellID(25296)
	hunterAspectOfTheBeast = spellID(1299447)
	hunterRaptorStrikeQ    = spellIDTag(14266, 3)
	hunterMongooseBite     = spellID(14271)
	hunterStriderKick      = spellID(1317257)
	hunterLacerate         = spellID(1299332)
	hunterWingClip         = spellID(14268)
	hunterImmolationTrap   = spellID(14305)
)

// Beast Mastery and Marksmanship under Forever rules. Auto Shot keeps firing through casts
// (sim/hunter/hunter.go), so the shots only compete for the GCD and the mana. Aimed Shot and
// Multi-Shot share a 6 sec cooldown, Arcane Shot and Summon Hawk another, Sniper Shot has its
// own 15 sec one. The lines that need a talent the build lacks simply never fire.
type hunterRanged struct{}

func (hunterRanged) Knobs() []Knob {
	return []Knob{
		// 1 Aimed Shot on the shared cooldown, 0 Multi-Shot.
		{Name: "aimed", Default: 1, Min: 0, Max: 1, Step: 1},
		// What goes on the Arcane Shot cooldown: 0 Arcane Shot, 1 Summon Hawk, 2 Summon Hawk while
		// fewer than 2 hawks are out and Arcane Shot otherwise.
		{Name: "hawk", Default: 2, Min: 0, Max: 2, Step: 1},
		{Name: "serpent", Default: 1, Min: 0, Max: 1, Step: 1},
		{Name: "sniper", Default: 1, Min: 0, Max: 1, Step: 1},
		// Sniper Shot above Aimed Shot when both are ready.
		{Name: "sniperFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		// The Arcane Shot cooldown above Aimed Shot when both are ready.
		{Name: "arcaneFirst", Default: 0, Min: 0, Max: 1, Step: 1},
		{Name: "intimidation", Default: 0, Min: 0, Max: 1, Step: 1},
		// Mana floors, in percent: Serpent Sting and the Arcane Shot cooldown only above them.
		{Name: "serpentMana", Default: 30, Min: 0, Max: 60, Step: 10},
		{Name: "arcaneMana", Default: 20, Min: 0, Max: 60, Step: 10},
		// Aimed Shot only above this much mana (Multi-Shot is a third of the cost).
		{Name: "aimedMana", Default: 0, Min: 0, Max: 40, Step: 10},
	}
}

func (hunterRanged) Build(k Knobs) *proto.APLRotation {
	floor := func(name string) value {
		if k[name] <= 0 {
			return nil
		}
		return ge(currentManaPercent(), num(k[name]/100))
	}

	items := []*proto.APLListItem{
		{
			Notes: "Step back to 25 yards when the fight starts in melee range, inside the 8 yard minimum.",
			Action: &proto.APLAction{
				Condition: inMeleeRange(),
				Action:    &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: num(25)}},
			},
		},
		cast(hunterAspectOfTheHawk, not(auraIsActive(hunterAspectOfTheHawk)), "Aspect of the Hawk if it ever drops."),
		autocastOtherCooldowns(nil, "Rapid Fire, Bestial Wrath, racials, trinkets and potions on cooldown."),
	}
	if k["intimidation"] == 1 {
		items = append(items, cast(hunterIntimidation, nil, "Intimidation on cooldown for the pet's crit, with the talent."))
	}
	if k["serpent"] == 1 {
		items = append(items, cast(hunterSerpentSting, and(not(dotIsActive(hunterSerpentSting)), floor("serpentMana")), "Serpent Sting when it is down."))
	}

	arcaneFloor := floor("arcaneMana")
	var arcane []*proto.APLListItem
	switch k["hawk"] {
	case 1:
		arcane = append(arcane, cast(hunterSummonHawk, arcaneFloor, "Summon Hawk on the cooldown it shares with Arcane Shot, with the talent."))
	case 2:
		hawksOut := and(dotIsActive(spellIDTag(1293527, 1)), dotIsActive(spellIDTag(1293527, 2)))
		arcane = append(arcane, cast(hunterSummonHawk, and(not(hawksOut), arcaneFloor), "Summon Hawk while fewer than 2 hawks are out, with the talent."))
	}
	arcane = append(arcane, cast(hunterArcaneShot, arcaneFloor, "Arcane Shot on its cooldown."))

	if k["arcaneFirst"] == 1 {
		items = append(items, arcane...)
	}
	sniper := cast(hunterSniperShot, nil, "Sniper Shot on cooldown, with the talent.")
	if k["sniper"] == 1 && k["sniperFirst"] == 1 {
		items = append(items, sniper)
	}
	if k["aimed"] == 1 {
		items = append(items, cast(hunterAimedShot, floor("aimedMana"), "Aimed Shot on the cooldown it shares with Multi-Shot."))
	}
	items = append(items, cast(hunterMultiShot, nil, "Multi-Shot when Aimed Shot is not taken."))
	if k["sniper"] == 1 && k["sniperFirst"] == 0 {
		items = append(items, sniper)
	}
	if k["arcaneFirst"] == 0 {
		items = append(items, arcane...)
	}

	rot := rotation(items...)
	rot.PrepullActions = append(rot.PrepullActions, prepull(hunterAspectOfTheHawk, 3))
	return rot
}
