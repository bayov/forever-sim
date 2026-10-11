package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// manaCostCheck is a spell, its cost and the share of it we should pay.
type manaCostCheck struct {
	name  string
	spell *core.Spell
	cost  float64
	pays  float64
}

// rankCost checks the top rank of a spell against its cost table.
func rankCost(t *testing.T, name string, ranks []*core.Spell, costs []float64, pays float64) manaCostCheck {
	spell := topRank(t, name, ranks)
	return manaCostCheck{name, spell, costs[spell.Rank], pays}
}

func checkManaCosts(t *testing.T, label string, checks []manaCostCheck) {
	for _, c := range checks {
		want := c.cost * c.pays
		if got := c.spell.Cost.GetCurrentCost(); math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: %s costs %.2f, want %.2f", label, c.name, got, want)
		}
	}
}

// TestOrcShamanManaCostTalents checks which spells each cost talent reaches at level 60, and
// that the talents add up.
//
// Convection takes 2% a point off shocks, Lightning Bolt, Chain Lightning and Lava Burst.
// Shamanistic Focus takes 45% off shocks and Lightning Shield. Totemic Focus takes 5% a point
// off totems. That's what the beta client texts say. They multiply, as the beta showed
// (shaman_audit.md 8.1), so with Convection 5/5 and Shamanistic Focus we pay 0.9 * 0.55 of a
// shock's cost, 49.5%. Added up, it would be 45%.
//
// Fire Nova is a spell and not a totem under Forever, so none of them reach it.
func TestOrcShamanManaCostTalents(t *testing.T) {
	cases := []struct {
		name                       string
		talents                    string
		shock, bolt, shield, totem float64
	}{
		// Every case has Lava Burst.
		{"no cost talents", "0000000000000001", 1, 1, 1, 1},
		{"Convection 2/5", "2000000000000001", .96, .96, 1, 1},
		{"Convection 5/5", "5000000000000001", .9, .9, 1, 1},
		{"Shamanistic Focus", "0000000000000001-000000001", .55, 1, .55, 1},
		{"Totemic Focus 5/5", "0000000000000001--05", 1, 1, 1, .75},
		{"all three", "5000000000000001-000000001-05", .9 * .55, .9, .55, .75},
	}
	for _, c := range cases {
		_, enh := newShamanSim(newOrcShaman(60, c.talents, &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
		checkManaCosts(t, c.name, []manaCostCheck{
			rankCost(t, "Earth Shock", enh.EarthShock, shaman.EarthShockManaCost[:], c.shock),
			rankCost(t, "Flame Shock", enh.FlameShock, shaman.FlameShockManaCost[:], c.shock),
			rankCost(t, "Frost Shock", enh.FrostShock, shaman.FrostShockManaCost[:], c.shock),
			rankCost(t, "Lightning Bolt", enh.LightningBolt, shaman.LightningBoltManaCost[:], c.bolt),
			rankCost(t, "Chain Lightning", enh.ChainLightning, shaman.ChainLightningManaCost[:], c.bolt),
			rankCost(t, "Lava Burst", []*core.Spell{enh.LavaBurst}, shaman.LavaBurstManaCost, c.bolt),
			rankCost(t, "Lightning Shield", enh.LightningShield, shaman.LightningShieldManaCost[:], c.shield),
			rankCost(t, "Searing Totem", enh.SearingTotem, shaman.SearingTotemManaCost[:], c.totem),
			rankCost(t, "Magma Totem", enh.MagmaTotem, shaman.MagmaTotemManaCost[:], c.totem),
			rankCost(t, "Flametongue Totem", enh.FlametongueTotem, shaman.FlametongueTotemManaCost[:], c.totem),
			rankCost(t, "Strength of Earth Totem", enh.StrengthOfEarthTotem, shaman.StrengthOfEarthTotemManaCost[:], c.totem),
			rankCost(t, "Stoneskin Totem", enh.StoneskinTotem, shaman.StoneskinTotemManaCost[:], c.totem),
			rankCost(t, "Windfury Totem", enh.WindfuryTotem, shaman.WindfuryTotemManaCost[:], c.totem),
			rankCost(t, "Grace of Air Totem", enh.GraceOfAirTotem, shaman.GraceOfAirTotemManaCost[:], c.totem),
			rankCost(t, "Windwall Totem", enh.WindwallTotem, shaman.WindwallTotemManaCost[:], c.totem),
			rankCost(t, "Mana Spring Totem", enh.ManaSpringTotem, shaman.ManaSpringTotemManaCost[:], c.totem),
			rankCost(t, "Healing Stream Totem", enh.HealingStreamTotem, shaman.HealingStreamTotemManaCost[:], c.totem),
			{"Tremor Totem", enh.TremorTotem, 60, c.totem},
			rankCost(t, "Fire Nova", enh.FireNova, shaman.FireNovaManaCost[:], 1),
		})
	}
}

// TestOrcShamanCostTalentsOnTheBeta checks the costs our beta character paid at level 30
// (shaman_audit.md 8.1).
//
// The log logs each cast's base cost, so the cost is how much our mana dropped by our next
// event. With Shamanistic Focus, Earth Shock rank 4 dropped 80, Flame Shock rank 3 88 and
// Lightning Shield rank 3 69. With Convection 5, Earth Shock dropped 130 and Flame Shock 144.
// With both, Earth Shock dropped 72 and Flame Shock 79, so the two talents multiply. Added up
// they would cost 65.25 and 72. The sim doesn't round costs.
func TestOrcShamanCostTalentsOnTheBeta(t *testing.T) {
	for _, c := range []struct {
		name, talents        string
		earth, flame, shield float64
	}{
		{"Shamanistic Focus", "-000000001", 145 * .55, 160 * .55, 125 * .55},
		{"Convection 5", "5", 145 * .9, 160 * .9, 125},
		{"both", "5-000000001", 145 * .9 * .55, 160 * .9 * .55, 125 * .55},
	} {
		_, enh := newShamanSim(newOrcShaman(30, c.talents, &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
		for _, check := range []struct {
			name  string
			spell *core.Spell
			rank  int
			want  float64
		}{
			{"Earth Shock", topRank(t, "Earth Shock", enh.EarthShock), 4, c.earth},
			{"Flame Shock", topRank(t, "Flame Shock", enh.FlameShock), 3, c.flame},
			{"Lightning Shield", topRank(t, "Lightning Shield", enh.LightningShield), 3, c.shield},
		} {
			if check.spell.Rank != check.rank {
				t.Fatalf("level 30 casts %s rank %d, want %d", check.name, check.spell.Rank, check.rank)
			}
			if got := check.spell.Cost.GetCurrentCost(); math.Abs(got-check.want) > 1e-9 {
				t.Errorf("%s: %s costs %.3f, want %.3f", c.name, check.name, got, check.want)
			}
		}
	}
}

// TestOrcShamanMaelstromWeaponCost checks Maelstrom Weapon 5/5 with Convection 5/5 at level
// 60.
//
// Each stack takes 20% off the next Lightning Bolt's cost and cast time, and the cost multiplies
// with Convection's 0.9. So 5 stacks make the bolt instant and free. The bolt uses up the stacks,
// and Chain Lightning doesn't get them.
func TestOrcShamanMaelstromWeaponCost(t *testing.T) {
	sim, enh := newShamanSim(newOrcShaman(60, "5-00000000000000005", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 30)
	full := enh.ManaRegenPerSecondWhileNotCasting()
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
	lbCost := shaman.LightningBoltManaCost[lb.Rank]
	cl := topRank(t, "Chain Lightning", enh.ChainLightning)
	clCost := shaman.ChainLightningManaCost[cl.Rank]
	mw := enh.MaelstromWeaponAura

	checkLightningBolt := func(sim *core.Simulation, stacks int) {
		want := lbCost * 0.9 * max(0, 1-0.2*float64(stacks))
		if got := lb.Cost.GetCurrentCost(); math.Abs(got-want) > 1e-9 {
			t.Errorf("%d stacks: Lightning Bolt costs %.2f, want %.2f", stacks, got, want)
		}
		wantCast := time.Duration(float64(lb.DefaultCast.CastTime) * (1 - 0.2*float64(stacks)))
		if got := lb.CastTime(); (got - wantCast).Abs() > time.Microsecond {
			t.Errorf("%d stacks: Lightning Bolt casts in %s, want %s", stacks, got, wantCast)
		}
		if got := cl.Cost.GetCurrentCost(); math.Abs(got-clCost*0.9) > 1e-9 {
			t.Errorf("%d stacks: Chain Lightning costs %.2f, want %.2f", stacks, got, clCost*0.9)
		}
	}

	at(sim, 0.5, func(sim *core.Simulation) {
		enh.SpendMana(sim, enh.MaxMana()-1000, enh.GetManaNotCastingMetrics())
	})
	at(sim, 1, func(sim *core.Simulation) {
		for stacks := 1; stacks <= 5; stacks++ {
			mw.Activate(sim)
			mw.SetStacks(sim, int32(stacks))
			checkLightningBolt(sim, stacks)
		}
		// The stacks run out after 30 sec, and the bolt goes back to its cost.
		mw.Deactivate(sim)
		checkLightningBolt(sim, 0)
	})

	// 2 stacks: a 1.5 sec cast for 0.9 * 0.6 of the cost.
	var mana float64
	at(sim, 2, func(sim *core.Simulation) {
		mw.Activate(sim)
		mw.SetStacks(sim, 2)
		mana = enh.CurrentMana()
		castNow(t, sim, enh, lb)
	})
	at(sim, 3.6, func(sim *core.Simulation) {
		want := mana + 1.5*full - lbCost*0.54
		if got := enh.CurrentMana(); math.Abs(got-want) > 1e-6 {
			t.Errorf("after a 2 stack bolt: got %.3f mana, want %.3f", got, want)
		}
		if mw.IsActive() {
			t.Errorf("Maelstrom Weapon is still up after Lightning Bolt")
		}
		checkLightningBolt(sim, 0)
	})

	// 5 stacks: instant and free.
	at(sim, 10, func(sim *core.Simulation) {
		mw.Activate(sim)
		mw.SetStacks(sim, 5)
		mana = enh.CurrentMana()
		castNow(t, sim, enh, lb)
		if enh.IsCasting(sim) {
			t.Errorf("a 5 stack bolt has a cast time")
		}
		if got := enh.CurrentMana(); got != mana {
			t.Errorf("a 5 stack bolt cost %.2f mana", mana-got)
		}
		if mw.IsActive() {
			t.Errorf("Maelstrom Weapon is still up after Lightning Bolt")
		}
		checkLightningBolt(sim, 0)
	})
	runSim(sim)
}

// TestOrcShamanClearcasting checks Elemental Focus at level 60.
//
// Each damage spell we cast has a 10% chance to give Clearcasting, which makes our next
// damage spell free. Fire Nova counts as one. Lightning Shield and totems still cost mana
// under it, and putting down a totem doesn't use it up. Our Searing Totem's attacks don't
// give it or use it up either.
func TestOrcShamanClearcasting(t *testing.T) {
	// Elemental Focus and Lava Burst.
	sim, enh := newShamanSim(newOrcShaman(60, "0000001000000001", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 3400)
	cc := enh.ClearcastingAura
	es := topRank(t, "Earth Shock", enh.EarthShock)
	fireNova := topRank(t, "Fire Nova", enh.FireNova)
	searing := topRank(t, "Searing Totem", enh.SearingTotem)
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
	metrics := enh.GetManaNotCastingMetrics()
	// The way Elemental Focus gives it, with the one stack a damage spell uses up.
	giveClearcasting := func(sim *core.Simulation) {
		cc.Activate(sim)
		cc.SetStacks(sim, cc.MaxStacks)
	}
	putDownSearing := func(sim *core.Simulation) {
		enh.AddMana(sim, enh.MaxMana(), metrics)
		castNow(t, sim, enh, searing)
	}

	// Searing Totem alone for 300 sec, about 120 attacks. We check every 5 sec that none of
	// them gave Clearcasting (it lasts 15 sec).
	for s := 1.0; s < 301; s += 50 {
		at(sim, s, putDownSearing)
	}
	for s := 2.0; s <= 300; s += 5 {
		at(sim, s, func(sim *core.Simulation) {
			if cc.IsActive() {
				t.Fatalf("at %s: Searing Totem's attacks gave Clearcasting", sim.CurrentTime)
			}
		})
	}
	at(sim, 310, giveClearcasting)
	at(sim, 320, func(sim *core.Simulation) {
		if !cc.IsActive() || cc.RemainingDuration(sim) != 5*time.Second {
			t.Errorf("Searing Totem's attacks used up Clearcasting")
		}
	})
	at(sim, 326, func(sim *core.Simulation) {
		if cc.IsActive() {
			t.Errorf("Clearcasting is still up after 16 sec")
		}
		attacks := enh.GetSpell(core.ActionID{SpellID: shaman.SearingTotemAttackSpellId[searing.Rank]})
		if casts := attacks.SpellMetrics[enh.CurrentTarget.UnitIndex].Casts; casts < 120 {
			t.Errorf("Searing Totem attacked only %d times", casts)
		}
	})

	// Earth Shock and Fire Nova in turn, 1000 casts each. A cast under Clearcasting must be
	// free and use it up, unless the cast gives it again.
	casts := map[*core.Spell]int{}
	procs := map[*core.Spell]int{}
	for i := range 2000 {
		s := 330 + 1.5*float64(i)
		if i%30 == 0 {
			// Searing Totem lasts 55 sec, and Fire Nova needs it up.
			at(sim, s-0.75, func(sim *core.Simulation) {
				wasActive := cc.IsActive()
				remaining := cc.RemainingDuration(sim)
				putDownSearing(sim)
				if wasActive && cc.RemainingDuration(sim) != remaining {
					t.Errorf("at %s: Searing Totem used up Clearcasting", sim.CurrentTime)
				}
			})
		}
		spell := []*core.Spell{es, fireNova}[i%2]
		at(sim, s, func(sim *core.Simulation) {
			wasActive := cc.IsActive()
			mana := enh.CurrentMana()
			spell.CD.Reset()
			if spell.SharedCD.Timer != nil {
				spell.SharedCD.Reset()
			}
			castNow(t, sim, enh, spell)
			cost, base := mana-enh.CurrentMana(), spell.Cost.BaseCost
			if wasActive && cost != 0 || !wasActive && cost != base {
				t.Errorf("at %s: %s cost %.2f with Clearcasting %t", sim.CurrentTime, spell.ActionID, cost, wasActive)
			}
			gained := cc.IsActive() && cc.RemainingDuration(sim) == cc.Duration
			if wasActive && !gained && cc.IsActive() {
				t.Errorf("at %s: %s didn't use up Clearcasting", sim.CurrentTime, spell.ActionID)
			}
			casts[spell]++
			if gained {
				procs[spell]++
			}
			enh.AddMana(sim, enh.MaxMana(), metrics)
		})
	}
	at(sim, 3330, func(sim *core.Simulation) {
		for _, spell := range []*core.Spell{es, fireNova} {
			if rate := float64(procs[spell]) / float64(casts[spell]); rate < 0.075 || rate > 0.125 {
				t.Errorf("%s gave Clearcasting on %.1f%% of %d casts, want 10%%", spell.ActionID, rate*100, casts[spell])
			}
		}
	})

	// Under Clearcasting the damage spells cost nothing and the rest cost what they always do.
	at(sim, 3360, func(sim *core.Simulation) {
		giveClearcasting(sim)
		checkManaCosts(t, "Clearcasting", []manaCostCheck{
			rankCost(t, "Earth Shock", enh.EarthShock, shaman.EarthShockManaCost[:], 0),
			rankCost(t, "Flame Shock", enh.FlameShock, shaman.FlameShockManaCost[:], 0),
			rankCost(t, "Frost Shock", enh.FrostShock, shaman.FrostShockManaCost[:], 0),
			rankCost(t, "Lightning Bolt", enh.LightningBolt, shaman.LightningBoltManaCost[:], 0),
			rankCost(t, "Chain Lightning", enh.ChainLightning, shaman.ChainLightningManaCost[:], 0),
			rankCost(t, "Lava Burst", []*core.Spell{enh.LavaBurst}, shaman.LavaBurstManaCost, 0),
			rankCost(t, "Fire Nova", enh.FireNova, shaman.FireNovaManaCost[:], 0),
			rankCost(t, "Lightning Shield", enh.LightningShield, shaman.LightningShieldManaCost[:], 1),
			rankCost(t, "Searing Totem", enh.SearingTotem, shaman.SearingTotemManaCost[:], 1),
			rankCost(t, "Magma Totem", enh.MagmaTotem, shaman.MagmaTotemManaCost[:], 1),
			rankCost(t, "Windfury Totem", enh.WindfuryTotem, shaman.WindfuryTotemManaCost[:], 1),
			rankCost(t, "Mana Spring Totem", enh.ManaSpringTotem, shaman.ManaSpringTotemManaCost[:], 1),
		})
	})

	// Lightning Bolt has a cast time. It pays at the end of the cast, so that's when it uses up
	// Clearcasting.
	var mana float64
	at(sim, 3370, func(sim *core.Simulation) {
		giveClearcasting(sim)
		mana = enh.CurrentMana()
		castNow(t, sim, enh, lb)
	})
	at(sim, 3371, func(sim *core.Simulation) {
		if !cc.IsActive() {
			t.Errorf("Clearcasting is gone during the Lightning Bolt cast")
		}
	})
	at(sim, 3372.6, func(sim *core.Simulation) {
		if cc.IsActive() && cc.RemainingDuration(sim) != cc.Duration-100*time.Millisecond {
			t.Errorf("Lightning Bolt didn't use up Clearcasting")
		}
		// We're at full mana, so we only check that the bolt paid nothing.
		if lb.CurCast.Cost != 0 || enh.CurrentMana() < mana {
			t.Errorf("Lightning Bolt cost %.2f under Clearcasting", lb.CurCast.Cost)
		}
	})
	runSim(sim)
}

// TestOrcShamanNotEnoughMana checks a spell we can't pay for at level 60.
//
// We can't cast it, it costs nothing, and the sim marks us out of mana until we can pay. The
// rotation then moves on to the next spell in its list. Here that's Earth Shock rank 1 after
// Lightning Bolt.
func TestOrcShamanNotEnoughMana(t *testing.T) {
	sim, enh := newShamanSim(newOrcShaman(60, "", &proto.EnhancementShaman_Options{}), &proto.Debuffs{}, 10)
	lb := topRank(t, "Lightning Bolt", enh.LightningBolt)
	at(sim, 1, func(sim *core.Simulation) {
		enh.SpendMana(sim, enh.CurrentMana()-lb.Cost.GetCurrentCost()+1, enh.GetManaNotCastingMetrics())
		mana := enh.CurrentMana()
		enh.SetGCDTimer(sim, sim.CurrentTime)
		if lb.Cast(sim, enh.CurrentTarget) {
			t.Errorf("cast Lightning Bolt with %.0f mana", mana)
		}
		if enh.CurrentMana() != mana || enh.IsCasting(sim) {
			t.Errorf("a failed Lightning Bolt still started or paid")
		}
		if !enh.IsOOM() {
			t.Errorf("not out of mana after a failed Lightning Bolt")
		}
		enh.AddMana(sim, 1, enh.GetManaNotCastingMetrics())
		castNow(t, sim, enh, lb)
		if enh.IsOOM() {
			t.Errorf("still out of mana after paying for Lightning Bolt")
		}
	})
	runSim(sim)

	spellID := func(id int32) *proto.ActionID { return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}} }
	castSpell := func(id int32) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spellID(id)}}}}
	}
	player := newOrcShaman(60, "", &proto.EnhancementShaman_Options{})
	player.Rotation = &proto.APLRotation{PriorityList: []*proto.APLListItem{
		castSpell(shaman.LightningBoltSpellId[shaman.LightningBoltRanks]),
		castSpell(shaman.EarthShockSpellId[1]),
	}}
	sim, enh = newShamanSim(player, &proto.Debuffs{}, 60)
	runSim(sim)
	target := enh.CurrentTarget.UnitIndex
	lbCasts := topRank(t, "Lightning Bolt", enh.LightningBolt).SpellMetrics[target].Casts
	esCasts := enh.EarthShock[1].SpellMetrics[target].Casts
	if lbCasts == 0 || esCasts < 5 {
		t.Errorf("cast Lightning Bolt %d times and Earth Shock %d times, want Earth Shock once out of mana", lbCasts, esCasts)
	}
}
