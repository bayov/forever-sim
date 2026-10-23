package retribution

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// TestHolyPowerOnItsSpellsOnly checks that Holy Power (client 1.60.1.70338, spell 5923) is a
// crit modifier on the paladin's own spells and not a spell crit stat, which would also reach
// Dragonbreath Chili, Shadow Oil and the other consumables. At 5 points Holy Shock gets 15%
// and Judgement of Righteousness 5%.
func TestHolyPowerOnItsSpellsOnly(t *testing.T) {
	const withHolyPower = "55003003201121251--05225230001"
	const withoutHolyPower = "55003003201121201--05225230001"

	_, with := retSim(t, withHolyPower, arcaniteReaper)
	_, without := retSim(t, withoutHolyPower, arcaniteReaper)

	if got, want := with.GetStat(stats.SpellCrit), without.GetStat(stats.SpellCrit); got != want {
		t.Errorf("spell crit stat with Holy Power = %.2f, want %.2f (no Holy Power)", got, want)
	}

	checks := []struct {
		name    string
		spellID int32
		want    float64
	}{
		{"Holy Shock", 20930, 15},
		{"Judgement of Righteousness", 20286, 5},
	}
	for _, c := range checks {
		spell := with.GetSpell(core.ActionID{SpellID: c.spellID})
		if spell == nil {
			t.Fatalf("%s (%d) is not registered", c.name, c.spellID)
		}
		if spell.BonusCritRating != c.want {
			t.Errorf("%s bonus crit = %.1f, want %.1f", c.name, spell.BonusCritRating, c.want)
		}
	}
}

// TestConsecrationFirstFourTargets checks Consecration rank 5 under Forever against five
// targets: the first four take the extra damage, its 4.2% of spell power and Consecrated
// Ground's 10% (2 points, on the mark of client 1280353), and the fifth takes the flat 12 a tick
// alone (wowsims/forever consecration.go).
func TestConsecrationFirstFourTargets(t *testing.T) {
	tick := func(spellPower float64) (first, fifth float64, firstMult, fifthMult float64) {
		sim, ret := retSimTargets(t, "55003003201121251--05225230001", arcaniteReaper, 5)
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

// TestHolyPowerOnConsecration checks that Holy Power's list (client 5923) takes in Consecration
// and leaves out the Seal of Righteousness proc.
func TestHolyPowerOnConsecration(t *testing.T) {
	_, ret := retSim(t, "55003003201121251--05225230001", arcaniteReaper)
	for _, c := range []struct {
		name    string
		spellID int32
		want    float64
	}{
		{"Consecration", 20924, 5},
		{"Seal of Righteousness proc", 25713, 0},
	} {
		if got := ret.GetSpell(core.ActionID{SpellID: c.spellID}).BonusCritRating; got != c.want {
			t.Errorf("%s bonus crit = %.1f, want %.1f", c.name, got, c.want)
		}
	}
}

// TestJudgementOfRighteousnessMelee checks that Forever's Judgement of Righteousness (client
// 20286, DefenseType 2) is a melee attack: it crits for double and Two-Handed Weapon
// Specialization (3/3: 6%) reaches it.
func TestJudgementOfRighteousnessMelee(t *testing.T) {
	_, with := retSim(t, "--00000000000300", arcaniteReaper)
	_, without := retSim(t, "", arcaniteReaper)
	jor := with.GetSpell(core.ActionID{SpellID: 20286})
	if jor.DefenseType != core.DefenseTypeMelee || jor.ProcMask != core.ProcMaskMeleeMHSpecial {
		t.Fatalf("Judgement of Righteousness defense %v, proc mask %v, want melee and main hand special", jor.DefenseType, jor.ProcMask)
	}
	if got := jor.CritMultiplier(with.AttackTables[with.CurrentTarget.UnitIndex][jor.CastType]); got != 2 {
		t.Errorf("Judgement of Righteousness crit multiplier = %.2f, want 2", got)
	}
	ratio := jor.DamageMultiplier / without.GetSpell(core.ActionID{SpellID: 20286}).DamageMultiplier
	if ratio < 1.0599 || ratio > 1.0601 {
		t.Errorf("Two-Handed Weapon Specialization 3/3 on Judgement of Righteousness = %.4f, want 1.06", ratio)
	}
}

// TestSanctifiedJudgementRefund checks Sanctified Judgement 3/3 under Forever against
// wowsims/forever's measurement (04b40b17b8): the refund is 60% of the seal's base cost with the
// percent cost cut added on top, so Seal of Righteousness rank 8 (200 mana) returns
// 200 x 1.1 x 0.6 = 132 with Benediction 5/5, and 200 x 1.28 x 0.6 = 153.6 with Twist of Light's
// 20% multiplied in.
func TestSanctifiedJudgementRefund(t *testing.T) {
	for _, c := range []struct {
		name    string
		talents string
		want    float64
	}{
		{"Benediction", "--0500003", 132},
		{"Benediction and Twist of Light", "--05000030000000001", 153.6},
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
// Precision (client 20189, aura 54) is unscoped hit, so the ranged Hammer keeps it, and
// Conviction (20117, weapon crit) only reaches the melee weapon's attacks (wowsims/forever
// conviction_test.go).
func TestHammerOfWrathHitAndCrit(t *testing.T) {
	_, ret := retSim(t, "-003-00005", arcaniteReaper)
	hammer := ret.GetSpell(core.ActionID{SpellID: 24239})
	if hammer.BonusHitRating != 0 {
		t.Errorf("Hammer of Wrath bonus hit = %.1f, want 0 (Precision kept)", hammer.BonusHitRating)
	}
	if got, want := hammer.BonusCritRating, -5.0; got != want {
		t.Errorf("Hammer of Wrath bonus crit = %.1f, want %.1f (Conviction 5/5 taken back off)", got, want)
	}
}

// TestSwiftJudgement checks Swift Judgement (client 1310994): it finishes Judgement's cooldown
// and makes the next Judgement free, and the Judgement after that costs mana again.
func TestSwiftJudgement(t *testing.T) {
	sim, ret := retSim(t, "02-5032210021-05225331001330311", arcaniteReaper)
	judgement := ret.GetSpell(core.ActionID{SpellID: 20271})
	swift := ret.GetSpell(core.ActionID{SpellID: 1310994})
	seal := ret.GetSpell(core.ActionID{SpellID: 20293})
	if judgement == nil || swift == nil || seal == nil {
		t.Fatalf("missing Judgement, Swift Judgement or Seal of Righteousness")
	}
	target := ret.CurrentTarget
	seal.Cast(sim, target)
	judgement.CD.Use(sim)
	if judgement.CD.IsReady(sim) {
		t.Fatalf("Judgement ready right after use")
	}
	swift.Cast(sim, target)
	if !judgement.CD.IsReady(sim) {
		t.Errorf("Judgement still on cooldown after Swift Judgement")
	}
	if cost := judgement.Cost.GetCurrentCost(); cost != 0 {
		t.Errorf("Judgement after Swift Judgement costs %.1f, want 0", cost)
	}
	judgement.Cast(sim, target)
	if cost := judgement.Cost.GetCurrentCost(); cost <= 0 {
		t.Errorf("the Judgement after the free one costs %.1f, want more than 0", cost)
	}
}

// TestJudgementOfTheCrusaderRank checks that a level 60 paladin's own Judgement of the Crusader
// puts the rank 6 debuff on the target: 161 Holy damage taken under Forever (140 x 1.15). Every
// rank used to share one debuff aura, so rank 1's 23 won.
func TestJudgementOfTheCrusaderRank(t *testing.T) {
	sim, ret := retSim(t, "52003-503-05225331001330311", arcaniteReaper)
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
// charge (5/5: always, at most once every 1.5 sec) without touching the swing timer, and the next
// main-hand swing spends it.
func TestReckoningCharges(t *testing.T) {
	sim, ret := retSim(t, "-00000000000005", arcaniteReaper)
	trigger := ret.GetAura("Reckoning Trigger")
	if trigger == nil {
		t.Fatalf("no Reckoning trigger with Reckoning 5/5")
	}
	mh := ret.AutoAttacks.MHAuto()
	crit := mh.NewResult(&ret.Unit)
	crit.Outcome = core.OutcomeCrit
	swingAt := ret.AutoAttacks.MainhandSwingAt()
	trigger.OnSpellHitTaken(trigger, sim, mh, crit)
	trigger.OnSpellHitTaken(trigger, sim, mh, crit) // inside the 1.5 sec cooldown
	stored := ret.GetAuraByID(core.ActionID{SpellID: 21919})
	if stored == nil || stored.GetStacks() != 1 {
		t.Fatalf("stored extra attacks after two crits inside 1.5 sec = %v, want 1", stored)
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

// TestShieldSpecializationMana checks Shield Specialization 3/3's mana since client 70425
// (spell 1310925: 4%, 6% before): a block restores 4% of maximum mana.
func TestShieldSpecializationMana(t *testing.T) {
	sim, ret := retSim(t, "-00000003", arcaniteReaper)
	aura := ret.GetAura("Shield Specialization Mana")
	if aura == nil {
		t.Fatalf("no Shield Specialization mana aura with 3/3")
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
