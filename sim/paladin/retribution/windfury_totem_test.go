package retribution

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/paladin"
)

// retSimWindfuryTotem is retSim with a raid shaman's Windfury Totem on the paladin.
func retSimWindfuryTotem(t *testing.T, talents string) (*core.Simulation, *RetributionPaladin, *core.Aura, *core.Aura) {
	t.Helper()
	sim, ret := retSimRaidBuffs(t, talents, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: arcaniteReaper}, 1,
		proto.MobType_MobTypeUnknown, &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness},
		&proto.RaidBuffs{TotemWeaponBuff: proto.TotemWeaponBuff_TotemWeaponBuffWindfury})
	totem := ret.GetAura("Windfury Totem Raid")
	attack := ret.GetAura("Windfury Attack")
	if totem == nil || attack == nil {
		t.Fatalf("no Windfury Totem on the paladin")
	}
	return sim, ret, totem, attack
}

// stepTo runs the sim to at.
func stepTo(t *testing.T, sim *core.Simulation, at time.Duration) {
	t.Helper()
	sim.AddPendingAction(&core.PendingAction{NextActionAt: at, OnAction: func(*core.Simulation) {}})
	for steps := 0; sim.CurrentTime < at; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", at)
		}
		sim.Step()
	}
}

// rollTotem hands Windfury Totem one landed hit of spell, past its 100 ms cooldown and with no
// swing of the paladin's own due within 100 ms, and says whether it procced. It drops the last
// proc's attack power first, and runs prepare, when set, right before the hit.
func rollTotem(t *testing.T, sim *core.Simulation, ret *RetributionPaladin, totem, attack *core.Aura, spell *core.Spell, prepare func()) bool {
	t.Helper()
	stepTo(t, sim, sim.CurrentTime+101*time.Millisecond)
	for ret.AutoAttacks.MainhandSwingAt() < sim.CurrentTime+100*time.Millisecond {
		stepTo(t, sim, sim.CurrentTime+50*time.Millisecond)
	}
	attack.Deactivate(sim)
	if prepare != nil {
		prepare()
	}
	result := spell.NewResult(ret.CurrentTarget)
	result.Outcome = core.OutcomeHit
	totem.OnSpellHitDealt(totem, sim, spell, result)
	return attack.IsActive()
}

// Unsealed Wrath stores no echo from a Windfury Totem swing (the November 11 hotfix), whichever
// hit launched the swing: a white swing, or a main hand ability that rolls the totem under
// Forever (standards 17; Heroic Strike stands in, the paladin having none).
func TestUnsealedWrathWindfuryTotemSwingFromAnyTrigger(t *testing.T) {
	for _, c := range []struct {
		name  string
		spell func(ret *RetributionPaladin) *core.Spell
	}{
		{"white swing", func(ret *RetributionPaladin) *core.Spell { return ret.AutoAttacks.MHAuto() }},
		{"main hand ability", func(ret *RetributionPaladin) *core.Spell {
			return &core.Spell{ActionID: core.ActionID{SpellID: 11567}, ProcMask: core.ProcMaskMeleeMHSpecial, Unit: &ret.Unit}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, ret, totem, attack := retSimWindfuryTotem(t, unsealedWrath)
			righteousness := ret.GetAura("Echo of Righteousness")
			spell := c.spell(ret)
			landed := 0
			for tries := 0; landed < 5; tries++ {
				if tries > 150 {
					t.Fatalf("%d of 5 landed Windfury Totem swings in 150 tries", landed)
				}
				prepare := func() {
					castByID(sim, ret, sealOfRighteousness9)
					righteousness.Deactivate(sim)
				}
				if !rollTotem(t, sim, ret, totem, attack, spell, prepare) {
					continue
				}
				// The extra swing's metrics (split 1), which the proc switched to.
				metrics := &ret.AutoAttacks.MHAuto().SpellMetrics[ret.CurrentTarget.UnitIndex]
				before := *metrics
				stepTo(t, sim, sim.CurrentTime+50*time.Millisecond)
				if metrics.Casts != before.Casts+1 {
					t.Fatalf("%d extra swings after a proc, want 1", metrics.Casts-before.Casts)
				}
				if metrics.Misses != before.Misses || metrics.Dodges != before.Dodges || metrics.Parries != before.Parries {
					continue
				}
				landed++
				if righteousness.IsActive() {
					t.Errorf("a landed Windfury Totem swing stored an Echo of Righteousness")
				}
			}
		})
	}
}

// No Judgement rolls Windfury Totem: the damaging ones are ranged attacks since the November 11
// hotfix, and the Crusader's and Wisdom's have no proc mask. A white swing still does.
func TestJudgementsDontRollWindfuryTotem(t *testing.T) {
	sim, ret, totem, attack := retSimWindfuryTotem(t, "")
	judgements := 0
	for _, spell := range ret.Spellbook {
		switch spell.SpellCode {
		case paladin.SpellCode_PaladinJudgementOfRighteousness, paladin.SpellCode_PaladinJudgementOfCommand, paladin.SpellCode_PaladinJudgementOfFury:
		default:
			continue
		}
		judgements++
		if judgements > 15 {
			t.Fatalf("more Judgement spells than the 60 sec fight has time for")
		}
		if spell.ProcMask.Matches(core.ProcMaskMelee) {
			t.Errorf("Judgement %v has a melee proc mask %v", spell.ActionID, spell.ProcMask)
		}
		for range 10 {
			if rollTotem(t, sim, ret, totem, attack, spell, nil) {
				t.Errorf("Judgement %v rolled Windfury Totem", spell.ActionID)
				break
			}
		}
	}
	if judgements == 0 {
		t.Fatalf("no damaging Judgement registered")
	}
	procced := false
	for range 100 {
		if procced = rollTotem(t, sim, ret, totem, attack, ret.AutoAttacks.MHAuto(), nil); procced {
			break
		}
	}
	if !procced {
		t.Errorf("a white swing never rolled Windfury Totem in 100 hits")
	}
}
