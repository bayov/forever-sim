package retribution

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestHolyPowerNoDamageCrit checks Holy Power since the October 30 hotfix (spell 5923 in
// 1.60.1.70707+hf: aura 50, crit healing only): 5/5 gives the paladin's damage spells no crit
// and no spell crit stat. Before the hotfix it gave Holy Shock 15% and the judgements 5%.
func TestHolyPowerNoDamageCrit(t *testing.T) {
	const withHolyPower = "0502521122002105-5-05233"
	const withoutHolyPower = "05025211220021-5-05233"

	_, with := retSim(t, withHolyPower, arcaniteReaper)
	_, without := retSim(t, withoutHolyPower, arcaniteReaper)

	if got, want := with.GetStat(stats.SpellCrit), without.GetStat(stats.SpellCrit); got != want {
		t.Errorf("spell crit stat with Holy Power = %.2f, want %.2f (no Holy Power)", got, want)
	}
	for _, c := range []struct {
		name    string
		spellID int32
	}{
		{"Holy Shock", 20930},
		{"Judgement of Righteousness", 20286},
		{"Consecration", 20924},
		{"Holy Strike", 10333},
		{"Seal of Righteousness proc", 25713},
	} {
		spell := with.GetSpell(core.ActionID{SpellID: c.spellID})
		if spell == nil {
			t.Fatalf("%s (%d) is not registered", c.name, c.spellID)
		}
		if spell.BonusCritRating != 0 {
			t.Errorf("%s bonus crit = %.1f, want 0", c.name, spell.BonusCritRating)
		}
	}
}

// TestConsecrationFirstFourTargets checks Consecration rank 5 under Forever against five
// targets: the first four take the extra damage, its 4.2% of spell power and Consecrated
// Ground's 10% (2 points, on the mark of client 1280353), and the fifth takes the flat 12 a tick
// alone (wowsims/forever consecration.go).
func TestConsecrationFirstFourTargets(t *testing.T) {
	tick := func(spellPower float64) (first, fifth float64, firstMult, fifthMult float64) {
		sim, ret := retSimTargets(t, "0502521122002105-5-05233", arcaniteReaper, 5)
		ret.AddStatDynamic(sim, stats.SpellPower, spellPower)
		consecration := ret.GetSpell(core.ActionID{SpellID: 20924})
		consecration.BonusHitRating = 10000
		consecration.BonusCritRating = -10000
		castByID(sim, ret, 20924)
		targets := sim.Encounter.ActiveTargetUnits
		for i, target := range targets {
			mark := target.GetAura("Consecration Mark" + ret.Label)
			if got, want := mark != nil && mark.IsActive(), i < 4; got != want {
				t.Fatalf("target %d marked = %v, want %v", i+1, got, want)
			}
		}
		consecration.AOEDot().TickOnce(sim)
		first = consecration.SpellMetrics[targets[0].UnitIndex].TotalDamage
		fifth = consecration.SpellMetrics[targets[4].UnitIndex].TotalDamage
		firstMult = consecration.TargetDamageMultiplier(ret.AttackTables[targets[0].UnitIndex][consecration.CastType], true)
		fifthMult = consecration.TargetDamageMultiplier(ret.AttackTables[targets[4].UnitIndex][consecration.CastType], true)
		return
	}
	lowFirst, lowFifth, firstMult, fifthMult := tick(0)
	highFirst, highFifth, _, _ := tick(1000)
	if ratio := firstMult / fifthMult; ratio < 1.0999 || ratio > 1.1001 {
		t.Errorf("Consecrated Ground on the first target / the fifth = %.4f, want 1.1000", ratio)
	}
	if lowFifth < 11.5 || lowFifth > 12.5 {
		t.Errorf("the fifth target's tick = %.2f, want 12", lowFifth)
	}
	if highFifth != lowFifth {
		t.Errorf("1000 spell power moved the fifth target's tick from %.2f to %.2f", lowFifth, highFifth)
	}
	// The ticks deal whole points of damage.
	if got, want := highFirst-lowFirst, 1000*0.042*1.1; got < want-1 || got > want+1 {
		t.Errorf("1000 spell power adds %.2f to the first target's tick, want %.2f", got, want)
	}
	if got, want := lowFirst, (12+27)*1.1; got < want-1 || got > want+1 {
		t.Errorf("the first target's tick = %.2f, want %.2f", got, want)
	}
}

// TestJudgementOfRighteousnessMelee checks that Forever's Judgement of Righteousness (client
// 20286, DefenseType 2) is a melee attack: it crits for double and Two-Handed Weapon
// Specialization (3/3: 6%) reaches it, through the Holy school since server hotfix push 113542
// (SpellEffect 697263 school mask 3: the specialization raises all Holy damage).
func TestJudgementOfRighteousnessMelee(t *testing.T) {
	_, with := retSim(t, "--000000000003", arcaniteReaper)
	_, without := retSim(t, "", arcaniteReaper)
	jor := with.GetSpell(core.ActionID{SpellID: 20286})
	if jor.DefenseType != core.DefenseTypeMelee || jor.ProcMask != core.ProcMaskMeleeMHSpecial {
		t.Fatalf("Judgement of Righteousness defense %v, proc mask %v, want melee and main hand special", jor.DefenseType, jor.ProcMask)
	}
	if got := jor.CritMultiplier(with.AttackTables[with.CurrentTarget.UnitIndex][jor.CastType]); got != 2 {
		t.Errorf("Judgement of Righteousness crit multiplier = %.2f, want 2", got)
	}
	holy := stats.SchoolIndexHoly
	ratio := jor.DamageMultiplier * with.PseudoStats.SchoolDamageDealtMultiplier[holy] /
		(without.GetSpell(core.ActionID{SpellID: 20286}).DamageMultiplier * without.PseudoStats.SchoolDamageDealtMultiplier[holy])
	if ratio < 1.0599 || ratio > 1.0601 {
		t.Errorf("Two-Handed Weapon Specialization 3/3 on Judgement of Righteousness = %.4f, want 1.06", ratio)
	}
}

// TestSanctifiedJudgementRefund checks Sanctified Judgement 3/3 under Forever against
// wowsims/forever's measurement (04b40b17b8): the refund is 60% of the seal's base cost with the
// percent cost cut added on top, so Seal of Righteousness rank 8 (200 mana) returns
// 200 x 1.1 x 0.6 = 132 with Benediction 5/5. Unsealed Wrath has no seal cost cut (Twist of
// Light's 20% gave 153.6), so it returns the same.
func TestSanctifiedJudgementRefund(t *testing.T) {
	for _, c := range []struct {
		name    string
		talents string
		want    float64
	}{
		{"Benediction", "--05003", 132},
		{"Benediction and Unsealed Wrath", "--05003000000000001", 132},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, ret := retSim(t, c.talents, arcaniteReaper)
			castByID(sim, ret, 20293)
			ret.SpendMana(sim, 0.5*ret.MaxMana(), ret.NewManaMetrics(core.ActionID{OtherID: 1}))
			before := ret.CurrentMana()
			castByID(sim, ret, 20271)
			if got := ret.CurrentMana() - before; got < c.want-0.01 || got > c.want+0.01 {
				t.Errorf("Judgement refunded %.2f mana, want %.2f", got, c.want)
			}
		})
	}
}

// TestHammerOfWrathHitAndCrit checks Precision and Conviction on Hammer of Wrath under Forever:
// Precision (client 20189, aura 54) is unscoped hit, so the ranged Hammer keeps it, and since the
// October 30 hotfix Conviction (20117, aura 290) is crit with all spells and attacks, so the Hammer
// gets its 6% through the crit stat with nothing taken back off. Champion of the Light takes 0.2
// sec a rank off its 1 sec cast.
func TestHammerOfWrathHitAndCrit(t *testing.T) {
	_, ret := retSim(t, "-003-0003", arcaniteReaper)
	_, base := retSim(t, "-003", arcaniteReaper)
	hammer := ret.GetSpell(core.ActionID{SpellID: 24239})
	if hammer.BonusHitRating != 0 {
		t.Errorf("Hammer of Wrath bonus hit = %.1f, want 0 (Precision kept)", hammer.BonusHitRating)
	}
	if got := hammer.BonusCritRating; got != 0 {
		t.Errorf("Hammer of Wrath bonus crit = %.1f, want 0 (Conviction is in the crit stat)", got)
	}
	if got, want := ret.GetStat(stats.MeleeCrit)-base.GetStat(stats.MeleeCrit), 6*float64(core.CritRatingPerCritChance); got < want-1e-9 || got > want+1e-9 {
		t.Errorf("Conviction 3/3 melee and ranged crit = %.2f, want %.2f", got, want)
	}
	if got := hammer.DefaultCast.CastTime; got != time.Second {
		t.Errorf("Hammer of Wrath cast %v without Champion of the Light, want 1s", got)
	}
	for _, c := range []struct {
		talents string
		want    time.Duration
	}{
		{"--0000000000000003", 400 * time.Millisecond},
		{"--0000000000000005", 0},
	} {
		_, ret := retSim(t, c.talents, arcaniteReaper)
		if got := ret.GetSpell(core.ActionID{SpellID: 24239}).DefaultCast.CastTime; got != c.want {
			t.Errorf("%s: Hammer of Wrath cast %v, want %v", c.talents, got, c.want)
		}
	}
}

// TestSwiftJudgement checks Swift Judgement since the October 30 hotfix (spell 1310994 in
// 1.60.1.70707+hf): a Judgement of its own, free, off the global cooldown, on a 1 min cooldown
// that Judgement's own doesn't share. It judges the active seal and leaves it up, and it leaves
// Judgement's cooldown and cost alone.
func TestSwiftJudgement(t *testing.T) {
	sim, ret := retSim(t, "-5032212001-05203", arcaniteReaper)
	judgement := ret.GetSpell(core.ActionID{SpellID: 20271})
	swift := ret.GetSpell(core.ActionID{SpellID: 1310994})
	seal := ret.GetSpell(core.ActionID{SpellID: 20293})
	jor := ret.GetSpell(core.ActionID{SpellID: 20286})
	if judgement == nil || swift == nil || seal == nil || jor == nil {
		t.Fatalf("missing Judgement, Swift Judgement, Seal of Righteousness or its judgement")
	}
	target := ret.CurrentTarget
	seal.Cast(sim, target)
	cost := judgement.Cost.GetCurrentCost()
	judgement.CD.Use(sim)
	casts := jor.SpellMetrics[target.UnitIndex].Casts
	mana := ret.CurrentMana()
	if !swift.Cast(sim, target) {
		t.Fatalf("Swift Judgement not cast while Judgement is on cooldown")
	}
	if got := jor.SpellMetrics[target.UnitIndex].Casts - casts; got != 1 {
		t.Errorf("Swift Judgement cast Judgement of Righteousness %d times, want 1", got)
	}
	if got := mana - ret.CurrentMana(); got != 0 {
		t.Errorf("Swift Judgement cost %.1f mana, want 0", got)
	}
	if judgement.CD.IsReady(sim) {
		t.Errorf("Judgement ready after Swift Judgement, want its cooldown left alone")
	}
	if got := judgement.Cost.GetCurrentCost(); got != cost {
		t.Errorf("Judgement costs %.1f after Swift Judgement, want %.1f", got, cost)
	}
	if got := swift.CD.Duration; got != time.Minute {
		t.Errorf("Swift Judgement cooldown %v, want 1m", got)
	}
	if swift.DefaultCast.GCD != 0 {
		t.Errorf("Swift Judgement GCD %v, want none", swift.DefaultCast.GCD)
	}
	if !ret.GetAuraByID(core.ActionID{SpellID: 20293}).IsActive() {
		t.Errorf("Swift Judgement consumed the seal")
	}
}

// TestJudgementOfTheCrusaderRank checks that a level 60 paladin's own Judgement of the Crusader
// puts the rank 6 debuff on the target: 161 Holy damage taken under Forever (140 x 1.15). Every
// rank used to share one debuff aura, so rank 1's 23 won.
func TestJudgementOfTheCrusaderRank(t *testing.T) {
	sim, ret := retSim(t, "--05233001001301351", arcaniteReaper)
	castByID(sim, ret, 20308)
	castByID(sim, ret, 20271)
	if got := ret.CurrentTarget.PseudoStats.SchoolBonusDamageTaken[stats.SchoolIndexHoly]; got < 160.99 || got > 161.01 {
		t.Errorf("Holy damage taken after Judgement of the Crusader = %.2f, want 161", got)
	}
}

// TestSealOfRighteousnessSpellPowerOnce checks that a Seal of Righteousness proc under Forever
// takes spell power once (Blizzard's 15 October beta notes fixed it gaining "almost twice as
// much benefit from Spell Damage"): 12% with a two-hander and 10% with a one-hander, as
// wowsims/forever measured (49f94c9326), and 2000 more spell power adds as much to a proc as to
// the spell power part alone. The one-hander's weapon factor is 0.85 less 0.02 a tenth of a
// second under 2.0 speed, so 0.83 for Thunderfury's 1.9.
func TestSealOfRighteousnessSpellPowerOnce(t *testing.T) {
	cases := []struct {
		name   string
		weapon int32
		coeff  float64
		hand   float64
	}{
		{"two-hander", arcaniteReaper, 0.12, 1.2},
		{"one-hander", thunderfury, 0.1, 0.83},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, ret := retSim(t, "", c.weapon)
			target := ret.CurrentTarget
			proc := ret.GetSpell(core.ActionID{SpellID: 25713})
			if proc == nil {
				t.Fatalf("no Seal of Righteousness rank 9 proc")
			}
			if proc.BonusCoefficient != c.coeff {
				t.Errorf("proc spell power coefficient = %.4f, want %.4f", proc.BonusCoefficient, c.coeff)
			}
			proc.BonusCritRating = -10000 // no crits, so every proc is a plain hit
			metrics := &proc.SpellMetrics[target.UnitIndex]
			hit := func() float64 {
				before := metrics.TotalDamage
				proc.ApplyEffects(sim, target, proc)
				return metrics.TotalDamage - before
			}
			spellPowerPart := func() float64 {
				return proc.CalcDamage(sim, target, 0, proc.OutcomeAlwaysHit).Damage
			}
			lowHit, lowPart := hit(), spellPowerPart()
			ret.AddStatDynamic(sim, stats.SpellPower, 2000)
			highHit, highPart := hit(), spellPowerPart()
			if highPart-lowPart < 1 {
				t.Fatalf("2000 spell power adds %.2f to the spell power part", highPart-lowPart)
			}
			// The procs deal whole points of damage, so allow 1 either way.
			if got, want := highHit-lowHit, highPart-lowPart; got < want-1 || got > want+1 {
				t.Errorf("2000 spell power adds %.2f to a proc, want %.2f (counted once)", got, want)
			}
			// Rank 8 at level 60 is 18.8 per hundred of swing speed, times the hand's factor.
			speed := ret.MainHand().SwingSpeed
			if got, want := lowHit-lowPart, 18.8*speed*c.hand; got < want-1 || got > want+1 {
				t.Errorf("the proc's weapon part = %.2f, want %.2f (%.2f at %.1f speed)", got, want, c.hand, speed)
			}
		})
	}
}

// TestReckoningCharges checks Reckoning since client 70425 (spell 1326084): a crit taken stores a
// charge (5/5: always) without touching the swing timer, a block too (5/5: 40%) but at most once
// every 1.5 sec ("Cannot occur from Blocks more than once every 1.5 sec", 20177), and the next
// main-hand swing spends them.
func TestReckoningCharges(t *testing.T) {
	sim, ret := retSim(t, "-000000000000005", arcaniteReaper)
	trigger := ret.GetAura("Reckoning Trigger")
	if trigger == nil {
		t.Fatalf("no Reckoning trigger with Reckoning 5/5")
	}
	mh := ret.AutoAttacks.MHAuto()
	crit := mh.NewResult(&ret.Unit)
	crit.Outcome = core.OutcomeCrit
	swingAt := ret.AutoAttacks.MainhandSwingAt()
	trigger.OnSpellHitTaken(trigger, sim, mh, crit)
	trigger.OnSpellHitTaken(trigger, sim, mh, crit) // the blocks' cooldown doesn't hold crits back
	stored := ret.GetAuraByID(core.ActionID{SpellID: 21919})
	if stored == nil || stored.GetStacks() != 2 {
		t.Fatalf("stored extra attacks after two crits = %v, want 2", stored)
	}
	block := mh.NewResult(&ret.Unit)
	block.Outcome = core.OutcomeBlock
	for range 50 { // 40% a block, so one charge comes in 50 tries, and only one inside 1.5 sec
		trigger.OnSpellHitTaken(trigger, sim, mh, block)
	}
	if got := stored.GetStacks(); got != 3 {
		t.Fatalf("stored extra attacks after 50 blocks inside 1.5 sec = %d, want 3", got)
	}
	if got := ret.AutoAttacks.MainhandSwingAt(); got != swingAt {
		t.Errorf("a stored charge moved the next swing from %s to %s", swingAt, got)
	}
	for sim.CurrentTime <= swingAt {
		sim.Step()
	}
	if got := stored.GetStacks(); got != 0 {
		t.Errorf("stored extra attacks after the next swing = %d, want 0", got)
	}
}

// TestShieldSpecializationMana checks Shield Specialization 2/2's mana: since client 70425 (spell
// 1310925) 4% of maximum mana (6% before), and since the October 30 hotfix on every block at 2/2
// (50% a rank of 2, it was 33% a rank of 3).
func TestShieldSpecializationMana(t *testing.T) {
	sim, ret := retSim(t, "-00000002", arcaniteReaper)
	aura := ret.GetAura("Shield Specialization Mana")
	if aura == nil {
		t.Fatalf("no Shield Specialization mana aura with 2/2")
	}
	ret.SpendMana(sim, 0.5*ret.MaxMana(), ret.NewManaMetrics(core.ActionID{OtherID: 1}))
	before := ret.CurrentMana()
	mh := ret.AutoAttacks.MHAuto()
	block := mh.NewResult(&ret.Unit)
	block.Outcome = core.OutcomeBlock
	aura.OnSpellHitTaken(aura, sim, mh, block)
	if got, want := ret.CurrentMana()-before, 0.04*ret.MaxMana(); got < want-0.01 || got > want+0.01 {
		t.Errorf("a block restored %.1f mana, want %.1f (4%% of maximum)", got, want)
	}
}
