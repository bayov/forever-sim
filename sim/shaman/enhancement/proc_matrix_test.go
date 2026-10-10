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
// Windfury Weapon, Windfury Totem, Flametongue Weapon and Frostbrand all need a landed main
// hand melee hit, white or yellow, and a blocked hit counts as landed. Flametongue Totem needs
// a landed white swing, so Stormstrike and Windfury Weapon's attacks never proc it. Misses,
// dodges, parries, spells, totem attacks, Lightning Shield orbs and the Flametongue and
// Frostbrand damage never trigger them. Flametongue Weapon blocks Flametongue Totem, and a
// Windfury proc's two attacks can't proc Windfury again, because of its 1.5 sec cooldown.
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
		white  []string // what a landed white hit triggers
		strike []string // what a landed Stormstrike triggers
	}{
		{"Windfury Weapon and Flametongue Totem", 30, whirlwind, proto.WeaponImbue_WindfuryWeapon, proto.TotemWeaponBuff_TotemWeaponBuffFlametongue,
			"Windfury Imbue", []string{windfuryWeapon, flametongueTotem}, []string{windfuryWeapon}},
		{"Flametongue Weapon and Flametongue Totem", 30, whirlwind, proto.WeaponImbue_FlametongueWeapon, proto.TotemWeaponBuff_TotemWeaponBuffFlametongue,
			"Flametongue Imbue", []string{flametongueWeapon}, []string{flametongueWeapon}},
		{"Frostbrand", 30, whirlwind, proto.WeaponImbue_FrostbrandWeapon, proto.TotemWeaponBuff_TotemWeaponBuffNone,
			"Frostbrand Imbue", []string{frostbrand}, []string{frostbrand}},
		// Windfury Totem needs level 32.
		{"Flametongue Weapon and Windfury Totem", 60, darkEdge, proto.WeaponImbue_FlametongueWeapon, proto.TotemWeaponBuff_TotemWeaponBuffWindfury,
			"Flametongue Imbue", []string{flametongueWeapon, windfuryTotem}, []string{flametongueWeapon, windfuryTotem}},
	}
	for _, setup := range setups {
		t.Run(setup.name, func(t *testing.T) {
			testImbueProcs(t, setup.level, setup.weapon, setup.imbue, setup.totem, setup.aura, setup.white, setup.strike)
		})
	}
}

func testImbueProcs(t *testing.T, level, weapon int32, imbue proto.WeaponImbue, totem proto.TotemWeaponBuff, auraLabel string, white, strike []string) {
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

	sim, enh := weaponSim{level: level, weapon: &proto.ItemSpec{Id: weapon}, imbue: imbue, totem: totem, seconds: 1}.start()
	swing := enh.AutoAttacks.MHAuto()
	stormstrikeSpell := enh.Stormstrike
	attacks := []attack{
		{"white hit", swing, core.OutcomeHit, white, ""},
		{"white crit", swing, core.OutcomeCrit, white, ""},
		{"glancing blow", swing, core.OutcomeGlance, white, ""},
		{"blocked white hit", swing, core.OutcomeBlock, white, ""},
		{"white miss", swing, core.OutcomeMiss, nil, ""},
		{"dodged white hit", swing, core.OutcomeDodge, nil, ""},
		{"parried white hit", swing, core.OutcomeParry, nil, ""},
		{"Stormstrike", stormstrikeSpell, core.OutcomeHit, strike, ""},
		{"blocked Stormstrike", stormstrikeSpell, core.OutcomeBlock, strike, ""},
		{"missed Stormstrike", stormstrikeSpell, core.OutcomeMiss, nil, ""},
		{"dodged Stormstrike", stormstrikeSpell, core.OutcomeDodge, nil, ""},
		{"parried Stormstrike", stormstrikeSpell, core.OutcomeParry, nil, ""},
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
	}
	sim, enh = weaponSim{level: level, weapon: &proto.ItemSpec{Id: weapon}, imbue: imbue, totem: totem, seconds: every * float64(rolls*len(attacks)+1)}.start()
	sim.PrePull()
	// Combat start turns the swings on, so we turn them off after it.
	at(sim, 0.5, enh.AutoAttacks.CancelAutoSwing)
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

	for i, a := range attacks {
		seen := map[string]bool{}
		for r := range rolls {
			at(sim, every*float64(1+i*rolls+r), func(sim *core.Simulation) {
				counts, rolled = map[string]int{}, a.spell
				a.spell.CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(a.outcome))
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
