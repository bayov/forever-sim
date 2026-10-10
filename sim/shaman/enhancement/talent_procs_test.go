package enhancement

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// Elemental Devastation 3/3, Elemental Focus and Lava Burst, then Flurry 5/5, Stormstrike
// and Maelstrom Weapon 5/5.
const talentProcTalents = "0000031000000001-00000000000510005"

const (
	flurry               = "Flurry"
	maelstromWeapon      = "Maelstrom Weapon"
	elementalDevastation = "Elemental Devastation"
)

// TestOrcShamanTalentProcs checks what triggers Flurry, Maelstrom Weapon and Elemental
// Devastation, which spells Clearcasting makes free, and what uses up Flurry's charges and
// Maelstrom Weapon's stacks, on a level 60 shaman with all four talents.
func TestOrcShamanTalentProcs(t *testing.T) {
	t.Run("triggers", testTalentProcTriggers)
	t.Run("Clearcasting spells", testClearcastingSpells)
	t.Run("Flurry charges", testFlurryCharges)
	t.Run("Maelstrom Weapon stacks", testMaelstromWeaponStacks)
}

// A melee crit (white hit, Stormstrike or Windfury Weapon attack) gives Flurry, and any
// landed melee hit can give a Maelstrom Weapon stack (10 a minute at 5/5, 58% a hit with
// a 3.5 sec weapon). A spell crit gives Elemental Devastation, but not a Fire Nova crit (the
// beta, shaman_audit.md 4.13). The Flametongue damage, Searing Totem attacks and Lightning
// Shield orbs give none of them, and neither does Lightning Bolt without Forever's Totem of
// the Storm. We force each outcome 60 times.
func testTalentProcTriggers(t *testing.T) {
	const rolls = 60
	sim, enh := weaponSim{level: 60, talents: talentProcTalents, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_WindfuryWeapon, totem: proto.TotemWeaponBuff_TotemWeaponBuffFlametongue, seconds: 60}.start()
	procs := map[string]*core.Aura{
		flurry:               enh.GetAura("Flurry Proc (16280)"),
		maelstromWeapon:      enh.MaelstromWeaponAura,
		elementalDevastation: enh.GetAura("Elemental Devastation Proc"),
	}
	for name, aura := range procs {
		if aura == nil {
			t.Fatalf("no %s", name)
		}
	}
	// We roll each attack on its own, without the Windfury and Flametongue Totem procs it
	// would bring.
	enh.GetAura("Windfury Imbue").Deactivate(sim)
	enh.GetAura("Flametongue Totem Raid (Rank 4)").Deactivate(sim)

	white := enh.AutoAttacks.MHAuto()
	melee := []string{flurry, maelstromWeapon}
	type attack struct {
		name    string
		spell   *core.Spell
		outcome core.HitOutcome
		procs   []string
	}
	attacks := []attack{
		{"white hit", white, core.OutcomeHit, []string{maelstromWeapon}},
		{"white crit", white, core.OutcomeCrit, melee},
		{"white miss", white, core.OutcomeMiss, nil},
		{"Stormstrike", enh.Stormstrike, core.OutcomeHit, []string{maelstromWeapon}},
		{"Stormstrike crit", enh.Stormstrike, core.OutcomeCrit, melee},
		{"Windfury Weapon attack", enh.WindfuryWeaponMH, core.OutcomeHit, []string{maelstromWeapon}},
		{"Windfury Weapon attack crit", enh.WindfuryWeaponMH, core.OutcomeCrit, melee},
		{"Flametongue Totem crit", lastSpell(enh, core.FlametongueTotemProcSpellId[:], 1), core.OutcomeCrit, nil},
		{"Searing Totem crit", lastSpell(enh, shaman.SearingTotemAttackSpellId[:], 0), core.OutcomeCrit, nil},
		{"Lightning Shield orb crit", topRank(t, "Lightning Shield", enh.LightningShieldProcs), core.OutcomeCrit, nil},
	}
	for _, spell := range []struct {
		name  string
		ranks []*core.Spell
	}{
		{"Earth Shock", enh.EarthShock},
		{"Flame Shock", enh.FlameShock},
		{"Frost Shock", enh.FrostShock},
		{"Lightning Bolt", enh.LightningBolt},
		{"Chain Lightning", enh.ChainLightning},
		{"Lava Burst", []*core.Spell{enh.LavaBurst}},
	} {
		top := topRank(t, spell.name, spell.ranks)
		attacks = append(attacks,
			attack{spell.name, top, core.OutcomeHit, nil},
			attack{spell.name + " crit", top, core.OutcomeCrit, []string{elementalDevastation}})
	}
	attacks = append(attacks, attack{"Fire Nova crit", topRank(t, "Fire Nova", enh.FireNova), core.OutcomeCrit, nil})

	for _, a := range attacks {
		seen := map[string]bool{}
		for range rolls {
			a.spell.CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(a.outcome))
			for name, aura := range procs {
				if aura.IsActive() {
					seen[name] = true
					aura.Deactivate(sim)
				}
			}
		}
		for _, name := range []string{flurry, maelstromWeapon, elementalDevastation} {
			if want := slices.Contains(a.procs, name); seen[name] != want {
				t.Errorf("%s triggered %s: %t, want %t", a.name, name, seen[name], want)
			}
		}
	}
}

// Clearcasting makes the next damage spell free: the shocks, Lightning Bolt, Chain
// Lightning, Lava Burst and Fire Nova. Those are the spells in its spell family mask in
// client 70291, and the same spells give it when we cast them. Totems, Lightning Shield and
// Stormstrike still cost mana.
func testClearcastingSpells(t *testing.T) {
	sim, enh := weaponSim{level: 60, talents: talentProcTalents, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_WindfuryWeapon, seconds: 60}.start()
	enh.ClearcastingAura.Activate(sim)
	for _, spell := range []struct {
		name  string
		ranks []*core.Spell
		free  bool
	}{
		{"Earth Shock", enh.EarthShock, true},
		{"Flame Shock", enh.FlameShock, true},
		{"Frost Shock", enh.FrostShock, true},
		{"Lightning Bolt", enh.LightningBolt, true},
		{"Chain Lightning", enh.ChainLightning, true},
		{"Lava Burst", []*core.Spell{enh.LavaBurst}, true},
		{"Fire Nova", enh.FireNova, true},
		{"Lightning Shield", enh.LightningShield, false},
		{"Searing Totem", enh.SearingTotem, false},
		{"Stormstrike", []*core.Spell{enh.Stormstrike}, false},
	} {
		cost := topRank(t, spell.name, spell.ranks).Cost.GetCurrentCost()
		if free := cost == 0; free != spell.free {
			t.Errorf("%s costs %.0f under Clearcasting, want free: %t", spell.name, cost, spell.free)
		}
	}
}

// White swings use Flurry's charges, at most one per 0.5 sec. Stormstrike and Windfury
// Weapon's attacks don't. A white crit uses a charge and then gives all 3 back. The buff
// runs out after 15 sec when we don't swing (clients 70291 and 1.15.9).
func testFlurryCharges(t *testing.T) {
	sim, enh := weaponSim{level: 60, talents: talentProcTalents, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_WindfuryWeapon, seconds: 25}.start()
	sim.PrePull()
	// Combat start turns the swings on, so we turn them off after it.
	at(sim, 0.5, enh.AutoAttacks.CancelAutoSwing)
	enh.GetAura("Windfury Imbue").Deactivate(sim)
	flurryAura := enh.GetAura("Flurry Proc (16280)")
	white := enh.AutoAttacks.MHAuto()

	roll := func(spell *core.Spell, outcome core.HitOutcome) {
		spell.CalcAndDealDamage(sim, enh.CurrentTarget, 100, forcedOutcome(outcome))
	}
	check := func(seconds float64, fn func(), want int32) {
		at(sim, seconds, func(sim *core.Simulation) {
			fn()
			if got := flurryAura.GetStacks(); got != want {
				t.Errorf("at %s: %d Flurry charges, want %d", sim.CurrentTime, got, want)
			}
		})
	}
	check(1, func() {
		flurryAura.Activate(sim)
		flurryAura.SetStacks(sim, 3)
		roll(enh.Stormstrike, core.OutcomeHit)
		roll(enh.WindfuryWeaponMH, core.OutcomeHit)
	}, 3)
	check(1.1, func() { roll(white, core.OutcomeHit) }, 2)
	check(1.4, func() { roll(white, core.OutcomeHit) }, 2)
	check(2, func() { roll(white, core.OutcomeCrit) }, 3)
	check(3, func() { roll(white, core.OutcomeHit) }, 2)
	check(4, func() { roll(white, core.OutcomeHit) }, 1)
	check(5, func() { roll(white, core.OutcomeHit) }, 0)

	check(6, func() {
		flurryAura.Activate(sim)
		flurryAura.SetStacks(sim, 3)
	}, 3)
	check(20.9, func() {}, 3)
	check(21.1, func() {}, 0)
	runSim(sim)
}

// Lightning Bolt uses all the Maelstrom Weapon stacks, and Chain Lightning uses none.
func testMaelstromWeaponStacks(t *testing.T) {
	sim, enh := weaponSim{level: 60, talents: talentProcTalents, weapon: &proto.ItemSpec{Id: darkEdge}, imbue: proto.WeaponImbue_WindfuryWeapon, seconds: 10}.start()
	sim.PrePull()
	// Combat start turns the swings on, so we turn them off after it.
	at(sim, 0.5, enh.AutoAttacks.CancelAutoSwing)
	mw := enh.MaelstromWeaponAura
	at(sim, 1, func(sim *core.Simulation) {
		mw.Activate(sim)
		mw.SetStacks(sim, 5)
		castNow(t, sim, enh, topRank(t, "Chain Lightning", enh.ChainLightning))
	})
	// 5 stacks at 5/5 make Lightning Bolt instant.
	at(sim, 5, func(sim *core.Simulation) {
		if got := mw.GetStacks(); got != 5 {
			t.Errorf("Chain Lightning left %d Maelstrom Weapon stacks, want 5", got)
		}
		castNow(t, sim, enh, topRank(t, "Lightning Bolt", enh.LightningBolt))
	})
	at(sim, 5.1, func(sim *core.Simulation) {
		if mw.IsActive() {
			t.Errorf("Lightning Bolt left %d Maelstrom Weapon stacks, want 0", mw.GetStacks())
		}
	})
	runSim(sim)
}

// forcedOutcome gives every roll the outcome, and no damage when it doesn't land.
func forcedOutcome(outcome core.HitOutcome) core.OutcomeApplier {
	return func(sim *core.Simulation, result *core.SpellResult, attackTable *core.AttackTable) {
		result.Outcome = outcome
		if !result.Landed() {
			result.Damage = 0
		}
	}
}
