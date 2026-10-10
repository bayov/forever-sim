package enhancement

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

const (
	windfuryWeapon    = "Windfury Weapon"
	windfuryTotem     = "Windfury Totem"
	flametongueWeapon = "Flametongue Weapon"
	flametongueTotem  = "Flametongue Totem"
	frostbrand        = "Frostbrand"
)

// TestOrcShamanImbueProcs checks which of our attacks trigger the weapon imbues and the
// weapon totems.
//
// Windfury Weapon, Windfury Totem, Flametongue Weapon, Flametongue Totem and Frostbrand all
// need a landed main hand melee hit, white or yellow, and a blocked hit counts as landed.
// Misses, dodges, parries, spells, totem attacks, Lightning Shield orbs and the Flametongue
// and Frostbrand damage never trigger them. Flametongue Weapon blocks Flametongue Totem, and
// a Windfury proc's two attacks can't proc Windfury again, because of its 1.5 sec cooldown.
//
// We force each outcome and roll each attack 60 times, 2 sec apart so that the cooldowns are
// always ready. A proc that can trigger shows up in 60 rolls (Windfury's 20% misses all 60
// one time in 600000), and one that can't never does.
func TestOrcShamanImbueProcs(t *testing.T) {
	setups := []struct {
		name   string
		level  int32
		weapon int32
		imbue  proto.WeaponImbue
		totem  proto.TotemWeaponBuff
		aura   string
		landed []string // what a landed white hit or Stormstrike triggers
	}{
		{"Windfury Weapon and Flametongue Totem", 30, whirlwind, proto.WeaponImbue_WindfuryWeapon, proto.TotemWeaponBuff_TotemWeaponBuffFlametongue,
			"Windfury Imbue", []string{windfuryWeapon, flametongueTotem}},
		{"Flametongue Weapon and Flametongue Totem", 30, whirlwind, proto.WeaponImbue_FlametongueWeapon, proto.TotemWeaponBuff_TotemWeaponBuffFlametongue,
			"Flametongue Imbue", []string{flametongueWeapon}},
		{"Frostbrand", 30, whirlwind, proto.WeaponImbue_FrostbrandWeapon, proto.TotemWeaponBuff_TotemWeaponBuffNone,
			"Frostbrand Imbue", []string{frostbrand}},
		// Windfury Totem needs level 32.
		{"Flametongue Weapon and Windfury Totem", 60, darkEdge, proto.WeaponImbue_FlametongueWeapon, proto.TotemWeaponBuff_TotemWeaponBuffWindfury,
			"Flametongue Imbue", []string{flametongueWeapon, windfuryTotem}},
	}
	for _, setup := range setups {
		t.Run(setup.name, func(t *testing.T) {
			testImbueProcs(t, setup.level, setup.weapon, setup.imbue, setup.totem, setup.aura, setup.landed)
		})
	}
}

func testImbueProcs(t *testing.T, level, weapon int32, imbue proto.WeaponImbue, totem proto.TotemWeaponBuff, auraLabel string, landed []string) {
	const (
		rolls = 60
		every = 2.0
	)
	type attack struct {
		name    string
		spell   *core.Spell
		outcome core.HitOutcome
		procs   []string
		ignore  string // a proc this attack never meets in a fight, see below
	}

	sim, enh := newWeaponSim(level, &proto.ItemSpec{Id: weapon}, imbue, totem, 0, 1)
	white := enh.AutoAttacks.MHAuto()
	stormstrike := enh.Stormstrike
	attacks := []attack{
		{"white hit", white, core.OutcomeHit, landed, ""},
		{"white crit", white, core.OutcomeCrit, landed, ""},
		{"glancing blow", white, core.OutcomeGlance, landed, ""},
		{"blocked white hit", white, core.OutcomeBlock, landed, ""},
		{"white miss", white, core.OutcomeMiss, nil, ""},
		{"dodged white hit", white, core.OutcomeDodge, nil, ""},
		{"parried white hit", white, core.OutcomeParry, nil, ""},
		{"Stormstrike", stormstrike, core.OutcomeHit, landed, ""},
		{"blocked Stormstrike", stormstrike, core.OutcomeBlock, landed, ""},
		{"missed Stormstrike", stormstrike, core.OutcomeMiss, nil, ""},
		{"dodged Stormstrike", stormstrike, core.OutcomeDodge, nil, ""},
		{"parried Stormstrike", stormstrike, core.OutcomeParry, nil, ""},
		{"Earth Shock", topRank(t, "Earth Shock", enh.EarthShock), core.OutcomeHit, nil, ""},
		{"Lightning Bolt", topRank(t, "Lightning Bolt", enh.LightningBolt), core.OutcomeHit, nil, ""},
		{"Fire Nova", topRank(t, "Fire Nova", enh.FireNova), core.OutcomeHit, nil, ""},
		{"Searing Totem attack", lastSpell(enh, shaman.SearingTotemAttackSpellId[:], 0), core.OutcomeHit, nil, ""},
		{"Lightning Shield orb", topRank(t, "Lightning Shield", enh.LightningShieldProcs), core.OutcomeHit, nil, ""},
	}
	for _, proc := range []struct {
		name string
		ids  []int32
		tag  int32
	}{
		{flametongueWeapon, shaman.FlametongueWeaponSpellId[:], 0},
		// Another shaman's Flametongue Totem.
		{flametongueTotem, core.FlametongueTotemProcSpellId[:], 1},
		{frostbrand, shaman.FrostbrandWeaponSpellId[:], 0},
	} {
		if spell := lastSpell(enh, proc.ids, proc.tag); spell != nil {
			attacks = append(attacks, attack{proc.name + " damage", spell, core.OutcomeHit, nil, ""})
		}
	}
	if enh.WindfuryWeaponMH != nil {
		// Its attacks only come right after a proc, inside the cooldown. We check that with
		// the white hits and Stormstrikes, which always bring both attacks and never more.
		attacks = append(attacks, attack{"Windfury Weapon attack", enh.WindfuryWeaponMH, core.OutcomeHit, nil, windfuryWeapon})
		if slices.Contains(landed, flametongueTotem) {
			attacks[len(attacks)-1].procs = []string{flametongueTotem}
		}
	}
	sim, enh = newWeaponSim(level, &proto.ItemSpec{Id: weapon}, imbue, totem, 0, every*float64(rolls*len(attacks)+1))
	sim.PrePull()
	enh.AutoAttacks.CancelAutoSwing(sim)
	// We rebuilt the sim to make it long enough, so we take the spells again.
	for i := range attacks {
		attacks[i].spell = enh.GetSpell(attacks[i].spell.ActionID)
	}

	classify := func(spell *core.Spell) string {
		id := spell.ActionID.SpellID
		switch {
		case spell == enh.WindfuryWeaponMH:
			return windfuryWeapon
		case id == 0:
			return ""
		case slices.Contains(core.FlametongueTotemProcSpellId[:], id):
			return flametongueTotem
		case slices.Contains(shaman.FlametongueWeaponSpellId[:], id):
			return flametongueWeapon
		case slices.Contains(shaman.FrostbrandWeaponSpellId[:], id):
			return frostbrand
		}
		return ""
	}

	// We see every hit through the imbue's callback, which runs for each of our hits. The
	// attack we roll is the first one, and we leave it out.
	var counts map[string]int
	var rolled *core.Spell
	hook := enh.GetAura(auraLabel)
	trigger := hook.OnSpellHitDealt
	hook.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if counts != nil {
			if spell == rolled {
				rolled = nil
			} else if name := classify(spell); name != "" {
				counts[name]++
			}
		}
		trigger(aura, sim, spell, result)
	}

	forced := func(outcome core.HitOutcome) core.OutcomeApplier {
		return func(sim *core.Simulation, result *core.SpellResult, attackTable *core.AttackTable) {
			result.Outcome = outcome
			if !result.Landed() {
				result.Damage = 0
			}
		}
	}
	for i, a := range attacks {
		seen := map[string]bool{}
		for r := range rolls {
			at(sim, every*float64(1+i*rolls+r), func(sim *core.Simulation) {
				counts, rolled = map[string]int{}, a.spell
				a.spell.CalcAndDealDamage(sim, enh.CurrentTarget, 100, forced(a.outcome))
				// A Windfury Totem proc moves the next swing to right now.
				if enh.AutoAttacks.MainhandSwingAt() == sim.CurrentTime {
					counts[windfuryTotem]++
				}
				if n := counts[windfuryWeapon]; a.ignore == "" && n != 0 && n != 2 {
					t.Errorf("%s brought %d Windfury Weapon attacks, want 2", a.name, n)
				}
				for name := range counts {
					seen[name] = true
				}
				counts = nil
			})
		}
		at(sim, every*float64(1+(i+1)*rolls)-1, func(sim *core.Simulation) {
			for _, name := range []string{windfuryWeapon, windfuryTotem, flametongueWeapon, flametongueTotem, frostbrand} {
				if name == a.ignore {
					continue
				}
				if want := slices.Contains(a.procs, name); seen[name] != want {
					t.Errorf("%s triggered %s: %t, want %t", a.name, name, seen[name], want)
				}
			}
		})
	}
	runSim(sim)
}

// lastSpell is our highest rank among the spells with ids, nil when we have none.
func lastSpell(enh *EnhancementShaman, ids []int32, tag int32) *core.Spell {
	for i := len(ids) - 1; i > 0; i-- {
		if spell := enh.GetSpell(core.ActionID{SpellID: ids[i], Tag: tag}); spell != nil {
			return spell
		}
	}
	return nil
}
