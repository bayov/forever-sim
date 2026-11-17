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
// Forever (standards 17): Judgement of Command, melee again since the November 13 hotfixes.
func TestUnsealedWrathWindfuryTotemSwingFromAnyTrigger(t *testing.T) {
	for _, c := range []struct {
		name  string
		spell func(ret *RetributionPaladin) *core.Spell
	}{
		{"white swing", func(ret *RetributionPaladin) *core.Spell { return ret.AutoAttacks.MHAuto() }},
		{"Judgement of Command", func(ret *RetributionPaladin) *core.Spell {
			return ret.GetSpell(core.ActionID{SpellID: 20966})
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

// Which of the paladin's hits roll Windfury Totem since Blizzard's hotfixes of November 13, 2026
// (t/2403040): "Windfury Totem is now correctly activated by all melee abilities" and "Judgement
// of Command now activates on-hit effects and counts as a melee attack. All other judgements will
// continue to be considered Ranged attacks." So Judgement of Command (a main hand melee special
// again) and Holy Strike roll it at 20%, as a white swing does, while Judgement of Righteousness
// and of Fury (ranged attacks since the November 11 hotfixes) and the seals' on-hit procs
// (Righteousness, Command, Fury: SpellFlagNoWindfuryTotem) never do. The paladin has Unsealed
// Wrath and Seal of Command, so every damaging Judgement is registered.
func TestJudgementsAndWindfuryTotem(t *testing.T) {
	const n = 150
	byID := func(id int32) func(ret *RetributionPaladin) *core.Spell {
		return func(ret *RetributionPaladin) *core.Spell { return ret.GetSpell(core.ActionID{SpellID: id}) }
	}
	holyStrike := func(ret *RetributionPaladin) *core.Spell {
		var top *core.Spell
		for _, spell := range ret.Spellbook {
			if spell.SpellCode == paladin.SpellCode_PaladinHolyStrike && (top == nil || spell.Rank > top.Rank) {
				top = spell
			}
		}
		return top
	}

	for _, c := range []struct {
		name     string
		spell    func(ret *RetributionPaladin) *core.Spell
		procMask core.ProcMask
		rolls    bool
	}{
		{"white swing", func(ret *RetributionPaladin) *core.Spell { return ret.AutoAttacks.MHAuto() }, core.ProcMaskMeleeMHAuto, true},
		{"Judgement of Command", byID(20966), core.ProcMaskMeleeMHSpecial, true},
		{"Holy Strike", holyStrike, core.ProcMaskMeleeMHSpecial, true},
		{"Judgement of Righteousness", byID(20286), core.ProcMaskRangedSpecial, false},
		{"Judgement of Fury", byID(20414), core.ProcMaskRangedSpecial, false},
		{"Seal of Righteousness proc", byID(25713), 0, false},
		{"Seal of Command proc", byID(20947), 0, false},
		{"Seal of Fury proc", byID(20418), 0, false},
	} {
		// A fresh sim for each, as the fight is 60 sec.
		sim, ret, totem, attack := retSimWindfuryTotem(t, unsealedWrath)
		spell := c.spell(ret)
		if spell == nil {
			t.Fatalf("%s is not registered", c.name)
		}
		if c.procMask != 0 && spell.ProcMask != c.procMask {
			t.Errorf("%s proc mask %v, want %v", c.name, spell.ProcMask, c.procMask)
		}
		procs := 0
		for range n {
			if rollTotem(t, sim, ret, totem, attack, spell, nil) {
				procs++
			}
		}
		if c.rolls && (procs < n/10 || procs > n/3) {
			t.Errorf("%s procced Windfury Totem %d times in %d landed hits, want about a fifth", c.name, procs, n)
		}
		if !c.rolls && procs != 0 {
			t.Errorf("%s procced Windfury Totem %d times in %d landed hits, want none", c.name, procs, n)
		}
	}
}

// A Shieldsworn Holy Strike strikes with the shield and never rolls Windfury Totem (Blizzard's
// hotfixes of 2026-11-16, t/2403040/11), where one without the talent rolls it at 20%.
func TestShieldswornHolyStrikeNoWindfuryTotem(t *testing.T) {
	const n = 150
	for _, c := range []struct {
		name    string
		talents string
		rolls   bool
	}{
		{"without Shieldsworn", "-00000000000001030", true},
		{"with Shieldsworn", "-00000000000001031", false},
	} {
		sim, ret, totem, attack := retSimWindfuryTotem(t, c.talents)
		var strike *core.Spell
		for _, spell := range ret.Spellbook {
			if spell.SpellCode == paladin.SpellCode_PaladinHolyStrike && (strike == nil || spell.Rank > strike.Rank) {
				strike = spell
			}
		}
		if strike == nil {
			t.Fatalf("%s: no Holy Strike", c.name)
		}
		procs := 0
		for range n {
			if rollTotem(t, sim, ret, totem, attack, strike, nil) {
				procs++
			}
		}
		if c.rolls && (procs < n/10 || procs > n/3) {
			t.Errorf("Holy Strike %s procced Windfury Totem %d times in %d landed hits, want about a fifth", c.name, procs, n)
		}
		if !c.rolls && procs != 0 {
			t.Errorf("Holy Strike %s procced Windfury Totem %d times in %d landed hits, want none", c.name, procs, n)
		}
	}
}
