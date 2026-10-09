package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestOrcShamanAttackTable checks how our melee attacks roll against a level 63 boss.
//
// White hits roll once on one table: miss, dodge, parry, glancing, block, crit, hit. Parry
// and block only happen when we stand in front of the target. Stormstrike and Windfury
// Weapon's attacks are yellow (the user, 2026-10-09). They roll miss, dodge, parry and block
// first, never glance, and a second roll decides the crit. A blocked yellow attack can't
// crit.
//
// With 300 weapon skill against 315 defense the 1.12 numbers are 8% miss, 6.5% dodge, 14%
// parry, 5% block and 40% glancing. We add 20% crit, so that the crit roll is big enough to
// tell one roll from two. Then we add 60% from behind, which pushes the white table past
// the crit cap. Only there does the order matter: glancing comes first and takes its full
// 40%, and crits fill what is left.
func TestOrcShamanAttackTable(t *testing.T) {
	cases := []struct {
		name      string
		front     bool
		addedCrit float64
	}{
		{"behind", false, 20},
		{"in front", true, 20},
		{"crit cap", false, 60},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			testAttackTable(t, c.front, c.addedCrit)
		})
	}
}

// outcomeCounts holds how many attacks of one kind ended in each outcome.
type outcomeCounts struct {
	total                                        int
	miss, dodge, parry, glance, block, crit, hit int
}

func (c *outcomeCounts) add(outcome core.HitOutcome) {
	c.total++
	switch {
	case outcome.Matches(core.OutcomeMiss):
		c.miss++
	case outcome.Matches(core.OutcomeDodge):
		c.dodge++
	case outcome.Matches(core.OutcomeParry):
		c.parry++
	case outcome.Matches(core.OutcomeGlance):
		c.glance++
	case outcome.Matches(core.OutcomeBlock):
		c.block++
	case outcome.Matches(core.OutcomeCrit):
		c.crit++
	case outcome.Matches(core.OutcomeHit):
		c.hit++
	}
}

func testAttackTable(t *testing.T, front bool, addedCrit float64) {
	const (
		miss   = 0.08
		dodge  = 0.065
		parry  = 0.14
		block  = 0.05
		glance = 0.4
	)
	player := newOrcShaman(60, "-0000000000001", &proto.EnhancementShaman_Options{ShamanImbue: proto.WeaponImbue_WindfuryWeapon})
	player.InFrontOfTarget = front
	player.Equipment = &proto.EquipmentSpec{}
	for range proto.ItemSlot_ItemSlotMainHand {
		player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{})
	}
	// Dark Edge of Insanity.
	player.Equipment.Items = append(player.Equipment.Items, &proto.ItemSpec{Id: 21134})
	player.Rotation.PriorityList = []*proto.APLListItem{{Action: &proto.APLAction{
		Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
			SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 17364}},
		}},
	}}}
	// We fight for 20 hours, for about 20,000 white hits and 8,000 of each yellow attack.
	sim, enh := newShamanSim(player, &proto.Debuffs{}, 72000)
	enh.AddStatDynamic(sim, stats.MeleeCrit, addedCrit*core.CritRatingPerCritChance)
	metrics := enh.GetManaNotCastingMetrics()

	if enh.Stormstrike == nil || enh.WindfuryWeaponMH == nil {
		t.Fatalf("missing Stormstrike or Windfury Weapon")
	}
	white := enh.AutoAttacks.MHAuto()
	attackTable := enh.AttackTables[enh.CurrentTarget.UnitIndex][white.CastType]
	crit := white.PhysicalCritChance(attackTable)
	for _, spell := range []*core.Spell{enh.Stormstrike, enh.WindfuryWeaponMH} {
		if got := spell.PhysicalCritChance(attackTable); math.Abs(got-crit) > 1e-9 {
			t.Fatalf("%s has %.4f crit chance, white hits %.4f", spell.ActionID, got, crit)
		}
	}

	var whites, stormstrikes, windfuries outcomeCounts
	imbue := enh.GetAura("Windfury Imbue")
	onHit := imbue.OnSpellHitDealt
	imbue.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		onHit(aura, sim, spell, result)
		switch {
		// A white hit is a copy of the auto attack spell, so we match on the action.
		case spell.ActionID == white.ActionID:
			whites.add(result.Outcome)
		case spell == enh.Stormstrike:
			stormstrikes.add(result.Outcome)
			// We never run out of mana for the next Stormstrike.
			enh.AddMana(sim, enh.MaxMana(), metrics)
		case spell == enh.WindfuryWeaponMH:
			windfuries.add(result.Outcome)
		}
	}
	runSim(sim)

	// check fails when got of n attacks is more than 4 standard errors away from the chance
	// want.
	check := func(attack, outcome string, got, n int, want float64) {
		if n < 3000 {
			t.Fatalf("only %d %s attacks to count %s on", n, attack, outcome)
		}
		rate := float64(got) / float64(n)
		if want == 0 {
			if got != 0 {
				t.Errorf("%s: %d of %d %s, want none", attack, got, n, outcome)
			}
			return
		}
		if se := math.Sqrt(want * (1 - want) / float64(n)); math.Abs(rate-want) > 4*se {
			t.Errorf("%s: %.4f %s (%d of %d), want %.4f +- %.4f", attack, rate, outcome, got, n, want, 4*se)
		}
	}
	frontOnly := func(chance float64) float64 {
		if front {
			return chance
		}
		return 0
	}

	// White hits take every outcome from the same roll, so each share is its chance. Crits
	// only get what the outcomes before them leave.
	whiteCrit := min(crit, 1-miss-dodge-frontOnly(parry)-glance-frontOnly(block))
	check("white", "miss", whites.miss, whites.total, miss)
	check("white", "dodge", whites.dodge, whites.total, dodge)
	check("white", "parry", whites.parry, whites.total, frontOnly(parry))
	check("white", "glancing", whites.glance, whites.total, glance)
	check("white", "block", whites.block, whites.total, frontOnly(block))
	check("white", "crit", whites.crit, whites.total, whiteCrit)
	if whiteCrit < crit {
		check("white", "normal hit", whites.hit, whites.total, 0)
	}

	for _, y := range []struct {
		name   string
		counts outcomeCounts
	}{{"Stormstrike", stormstrikes}, {"Windfury", windfuries}} {
		c := y.counts
		check(y.name, "miss", c.miss, c.total, miss)
		check(y.name, "dodge", c.dodge, c.total, dodge)
		check(y.name, "parry", c.parry, c.total, frontOnly(parry))
		check(y.name, "glancing", c.glance, c.total, 0)
		check(y.name, "block", c.block, c.total, frontOnly(block))
		// The crit comes from a second roll on the attacks that hit and weren't blocked.
		check(y.name, "crit", c.crit, c.crit+c.hit, crit)
	}
}
