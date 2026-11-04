package retribution

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The rules of the launch realms' October 30 hotfix trees (1.60.1.70707+hf, wago.tools hotfixes
// region 91; see sim/paladin/talents.go and unsealed_wrath.go).

const unsealedWrath = "--00000001000000001" // with Seal of Command
const sealOfRighteousness9 = 20293
const sealOfCommand5 = 20920

// whiteHit gives the Unsealed Wrath aura one main-hand white swing of outcome on the target.
func whiteHit(sim *core.Simulation, ret *RetributionPaladin, aura *core.Aura, outcome core.HitOutcome) {
	mh := ret.AutoAttacks.MHAuto()
	result := mh.NewResult(ret.CurrentTarget)
	result.Outcome = outcome
	aura.OnSpellHitDealt(aura, sim, mh, result)
}

func procCasts(ret *RetributionPaladin, spellID int32) int32 {
	return ret.GetSpell(core.ActionID{SpellID: spellID}).SpellMetrics[ret.CurrentTarget.UnitIndex].Casts
}

// TestUnsealedWrathStoresAndSpends checks Unsealed Wrath (spell 1310735): a landed white swing
// under a seal stores a copy of it, up to 3 for 15 sec; the first landed swing under another seal
// spends them all, the stored Seal of Righteousness firing once a stack, and stores a copy of the
// new seal. Seal of the Crusader has no Echo, so it stores nothing but spends the others. A miss
// and a Holy Strike do neither.
func TestUnsealedWrathStoresAndSpends(t *testing.T) {
	sim, ret := retSim(t, unsealedWrath, arcaniteReaper)
	uw := ret.GetAura("Unsealed Wrath")
	command, righteousness := ret.GetAura("Echo of Command"), ret.GetAura("Echo of Righteousness")
	if uw == nil || command == nil || righteousness == nil {
		t.Fatalf("no Unsealed Wrath or Echo auras with the talent")
	}
	if !uw.IsActive() {
		t.Fatalf("Unsealed Wrath not active")
	}

	castByID(sim, ret, sealOfRighteousness9)
	whiteHit(sim, ret, uw, core.OutcomeMiss)
	if righteousness.IsActive() {
		t.Fatalf("a missed swing stored a copy")
	}
	for range 4 {
		whiteHit(sim, ret, uw, core.OutcomeHit)
	}
	if got := righteousness.GetStacks(); got != 3 {
		t.Fatalf("Echo of Righteousness stacks after 4 swings = %d, want 3", got)
	}
	if got := righteousness.Duration; got != 15*time.Second {
		t.Errorf("Echo duration %v, want 15s", got)
	}

	castByID(sim, ret, sealOfCommand5)
	if got := righteousness.GetStacks(); got != 3 {
		t.Fatalf("the seal change spent the copies (%d stacks left)", got)
	}
	holyStrike := ret.GetSpell(core.ActionID{SpellID: 10333})
	hsResult := holyStrike.NewResult(ret.CurrentTarget)
	hsResult.Outcome = core.OutcomeHit
	uw.OnSpellHitDealt(uw, sim, holyStrike, hsResult)
	if got := righteousness.GetStacks(); got != 3 {
		t.Fatalf("a Holy Strike spent the copies (%d stacks left)", got)
	}

	before := procCasts(ret, 25713)
	whiteHit(sim, ret, uw, core.OutcomeHit)
	if got := procCasts(ret, 25713) - before; got != 3 {
		t.Errorf("Seal of Righteousness procs from 3 stored copies = %d, want 3", got)
	}
	if righteousness.IsActive() {
		t.Errorf("Echo of Righteousness still up after a swing under Seal of Command")
	}
	if got := command.GetStacks(); got != 1 {
		t.Errorf("Echo of Command stacks after a swing under Seal of Command = %d, want 1", got)
	}

	castByID(sim, ret, 20308)
	whiteHit(sim, ret, uw, core.OutcomeHit)
	if command.IsActive() || righteousness.IsActive() {
		t.Errorf("an Echo is up after a swing under Seal of the Crusader: Command %v, Righteousness %v", command.IsActive(), righteousness.IsActive())
	}
}

// TestUnsealedWrathCommandChance checks that each stored Seal of Command rolls the seal's own
// chance, 7 procs a minute on the 3.8 sec Arcanite Reaper (44.3%), with no 1 sec proc cooldown
// between the stacks, so three stacks can all fire on one swing.
func TestUnsealedWrathCommandChance(t *testing.T) {
	sim, ret := retSim(t, unsealedWrath, arcaniteReaper)
	uw := ret.GetAura("Unsealed Wrath")
	const swings = 3000
	all3 := 0
	var procs int32
	for range swings {
		castByID(sim, ret, sealOfCommand5)
		for range 3 {
			whiteHit(sim, ret, uw, core.OutcomeHit)
		}
		castByID(sim, ret, sealOfRighteousness9)
		before := procCasts(ret, 20947)
		whiteHit(sim, ret, uw, core.OutcomeHit)
		got := procCasts(ret, 20947) - before
		procs += got
		if got == 3 {
			all3++
		}
	}
	chance := 7.0 * 3.8 / 60
	if got := float64(procs) / (3 * swings); got < chance-0.02 || got > chance+0.02 {
		t.Errorf("stored Seal of Command proc rate = %.3f, want %.3f", got, chance)
	}
	if all3 == 0 {
		t.Errorf("no swing fired all three stored Seals of Command")
	}
}

// TestUnsealedWrathTalentNeeded checks that without the talent there are no Echoes, and that the
// seals cost their Benediction price with it (Twist of Light's 20% cut is gone): Seal of
// Righteousness rank 8 is 200 x 0.9 = 180 with Benediction 5/5.
func TestUnsealedWrathTalentNeeded(t *testing.T) {
	_, without := retSim(t, "", arcaniteReaper)
	if without.GetAura("Echo of Command") != nil || without.GetAura("Unsealed Wrath") != nil {
		t.Errorf("Unsealed Wrath auras without the talent")
	}
	_, with := retSim(t, "--05000000000000001", arcaniteReaper)
	if got := with.GetSpell(core.ActionID{SpellID: sealOfRighteousness9}).Cost.GetCurrentCost(); got < 179.99 || got > 180.01 {
		t.Errorf("Seal of Righteousness costs %.2f with Benediction 5/5 and Unsealed Wrath, want 180", got)
	}
}

// TestConvictionAllCrit checks Conviction 3/3 (spell 20117, aura 290): 6% crit with spells and
// attacks alike.
func TestConvictionAllCrit(t *testing.T) {
	_, base := retSim(t, "", arcaniteReaper)
	_, ret := retSim(t, "--00030", arcaniteReaper)
	for _, stat := range []stats.Stat{stats.MeleeCrit, stats.SpellCrit} {
		if got := ret.GetStat(stat) - base.GetStat(stat); got < 5.999 || got > 6.001 {
			t.Errorf("Conviction 3/3 %v = %.3f, want 6", stat, got)
		}
	}
}

// TestRetributionTalentRanks checks the hotfix's new rank values: Improved Seals 3% a rank on the
// seal procs and judgements, Champion of the Light 12% of Intellect a rank as spell power, Divine
// Precision 4% Holy hit a rank, Reverence 15% casting regeneration a rank, and Vindication no
// longer raising the paladin's attack power.
func TestRetributionTalentRanks(t *testing.T) {
	_, base := retSim(t, "", arcaniteReaper)
	ratio := func(ret *RetributionPaladin, id int32) float64 {
		return ret.GetSpell(core.ActionID{SpellID: id}).DamageMultiplier / base.GetSpell(core.ActionID{SpellID: id}).DamageMultiplier
	}
	for _, c := range []struct {
		talents string
		want    float64
	}{{"00002", 1.06}, {"00005", 1.15}} {
		_, ret := retSim(t, c.talents, arcaniteReaper)
		for _, id := range []int32{25713, 20286} {
			if got := ratio(ret, id); got < c.want-1e-9 || got > c.want+1e-9 {
				t.Errorf("Improved Seals %s on %d = %.4f, want %.2f", c.talents, id, got, c.want)
			}
		}
	}

	_, cotl := retSim(t, "--0000000000000005", arcaniteReaper)
	if got, want := cotl.GetStat(stats.SpellPower)-base.GetStat(stats.SpellPower), 0.6*cotl.GetStat(stats.Intellect); got < want-0.01 || got > want+0.01 {
		t.Errorf("Champion of the Light 5/5 spell power = %.2f, want %.2f (60%% of Intellect)", got, want)
	}

	_, dp := retSim(t, "0000000002000100", arcaniteReaper)
	if got := dp.GetSpell(core.ActionID{SpellID: 20930}).BonusHitRating; got != 8*core.SpellHitRatingPerHitChance {
		t.Errorf("Divine Precision 2/2 on Holy Shock = %.1f hit, want 8%%", got)
	}

	_, rev := retSim(t, "0002", arcaniteReaper)
	if got := rev.PseudoStats.SpiritRegenRateCasting - base.PseudoStats.SpiritRegenRateCasting; got < 0.2999 || got > 0.3001 {
		t.Errorf("Reverence 2/2 casting regeneration = %.3f, want 0.3", got)
	}

	_, vind := retSim(t, "--000003", arcaniteReaper)
	if vind.GetAura("Vindication Proc") != nil {
		t.Errorf("Vindication registers an attack power proc")
	}
}

// TestTwoHandedSpecializationThreat checks the hotfix's threat cut on Two-Handed Weapon
// Specialization (spell 20111, effect 1): 20% less threat at 3/3 with a two-hander, none with a
// one-hander, which also gets no damage from it.
func TestTwoHandedSpecializationThreat(t *testing.T) {
	_, base := retSim(t, "", arcaniteReaper)
	_, twoHand := retSim(t, "--000000000003", arcaniteReaper)
	_, oneHand := retSim(t, "--000000000003", thunderfury)
	if got := twoHand.PseudoStats.ThreatMultiplier / base.PseudoStats.ThreatMultiplier; got < 0.7999 || got > 0.8001 {
		t.Errorf("threat with a two-hander = %.4f, want 0.80", got)
	}
	if got := oneHand.PseudoStats.ThreatMultiplier / base.PseudoStats.ThreatMultiplier; got != 1 {
		t.Errorf("threat with a one-hander = %.4f, want 1", got)
	}
	if got := oneHand.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]; got != base.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] {
		t.Errorf("physical damage with a one-hander = %.4f, want unchanged", got)
	}
}

// TestProtectionTalentRanks checks the Protection side: Redoubt 6% block a rank, Improved
// Righteous Fury 3% less damage taken a rank, Sacred Duty 5 sec off Forbearance a rank and no
// Stamina, Iron Creed 10% Holy Strike threat a rank and, with Righteous Fury, 2% less damage taken
// a rank for 6 sec after a Holy Strike.
func TestProtectionTalentRanks(t *testing.T) {
	_, base := retSim(t, "", arcaniteReaper)

	sim, redoubt := retSim(t, "-05", arcaniteReaper)
	block := redoubt.GetStat(stats.Block)
	redoubt.GetAura("Redoubt").Activate(sim)
	if got := redoubt.GetStat(stats.Block) - block; got < 29.999 || got > 30.001 {
		t.Errorf("Redoubt 5/5 block = %.2f, want 30", got)
	}

	rf := &proto.PaladinOptions{PrimarySeal: proto.PaladinSeal_Righteousness, RighteousFury: true}
	_, rfBase := retSimOptions(t, "", withWeapon(nil), 1, proto.MobType_MobTypeUnknown, rf)
	_, irf := retSimOptions(t, "-00000000003", withWeapon(nil), 1, proto.MobType_MobTypeUnknown, rf)
	if got := irf.PseudoStats.DamageTakenMultiplier / rfBase.PseudoStats.DamageTakenMultiplier; got < 0.9099 || got > 0.9101 {
		t.Errorf("Improved Righteous Fury 3/3 damage taken = %.4f, want 0.91", got)
	}

	_, sd := retSim(t, "-00002", arcaniteReaper)
	if got := sd.GetAura("Forbearance").Duration; got != 50*time.Second {
		t.Errorf("Forbearance %v with Sacred Duty 2/2, want 50s", got)
	}
	if got, want := sd.GetStat(stats.Stamina), base.GetStat(stats.Stamina); got != want {
		t.Errorf("Stamina %.1f with Sacred Duty 2/2, want %.1f", got, want)
	}

	sim, ic := retSimOptions(t, "-0000000000000003", withWeapon(nil), 1, proto.MobType_MobTypeUnknown, rf)
	holyStrike := ic.GetSpell(core.ActionID{SpellID: 10333})
	if got := holyStrike.ThreatMultiplier / rfBase.GetSpell(core.ActionID{SpellID: 10333}).ThreatMultiplier; got < 1.2999 || got > 1.3001 {
		t.Errorf("Iron Creed 3/3 Holy Strike threat = %.4f, want 1.3", got)
	}
	holyStrike.BonusHitRating = 10000
	taken := ic.PseudoStats.DamageTakenMultiplier
	holyStrike.ApplyEffects(sim, ic.CurrentTarget, holyStrike)
	aura := ic.GetAura("Iron Creed")
	if aura == nil || !aura.IsActive() || aura.Duration != 6*time.Second {
		t.Fatalf("no 6 sec Iron Creed aura after a Holy Strike")
	}
	if got := ic.PseudoStats.DamageTakenMultiplier / taken; got < 0.9399 || got > 0.9401 {
		t.Errorf("Iron Creed 3/3 damage taken = %.4f, want 0.94", got)
	}
}

// TestHolyShieldAndShieldsworn checks Holy Shield's ranks since the hotfix (20925 at 30 for 100
// mana and 75 a block, 20927 at 40, 20928 at 50 for 195 mana and 153) and Shieldsworn (1328544):
// Holy Strike's cooldown is 6 sec, and a Holy Strike that lands deals Holy Shield's damage while
// it is up without taking a charge.
func TestHolyShieldAndShieldsworn(t *testing.T) {
	sim, ret := retSim(t, "-00000000000001001", arcaniteReaper)
	for _, c := range []struct {
		id    int32
		level int
		cost  float64
	}{{20925, 30, 100}, {20927, 40, 150}, {20928, 50, 195}} {
		spell := ret.GetSpell(core.ActionID{SpellID: c.id})
		if spell == nil {
			t.Fatalf("Holy Shield %d not registered", c.id)
		}
		if spell.RequiredLevel != c.level || spell.Cost.GetCurrentCost() != c.cost {
			t.Errorf("Holy Shield %d: level %d, cost %.0f, want %d and %.0f", c.id, spell.RequiredLevel, spell.Cost.GetCurrentCost(), c.level, c.cost)
		}
	}
	holyStrike := ret.GetSpell(core.ActionID{SpellID: 10333})
	if got := holyStrike.CD.Duration; got != 6*time.Second {
		t.Errorf("Holy Strike cooldown %v with Shieldsworn, want 6s", got)
	}
	castByID(sim, ret, 20928)
	shield := ret.GetAuraByID(core.ActionID{SpellID: 20928})
	proc := ret.GetSpell(core.ActionID{SpellID: 20957})
	holyStrike.BonusHitRating = 10000
	before := proc.SpellMetrics[ret.CurrentTarget.UnitIndex].Casts
	holyStrike.ApplyEffects(sim, ret.CurrentTarget, holyStrike)
	if got := proc.SpellMetrics[ret.CurrentTarget.UnitIndex].Casts - before; got != 1 {
		t.Errorf("Holy Shield procs from a Shieldsworn Holy Strike = %d, want 1", got)
	}
	if got := shield.GetStacks(); got != 4 {
		t.Errorf("Holy Shield charges after the Holy Strike = %d, want 4", got)
	}
}

// TestHotfix113542 checks the paladin rows of server hotfix push 113542 (region 91, build 70707,
// 2026-11-03 18:29Z, with the rogue push 113483; orchestration/round_70753/hotfix_113483_paladin.md),
// the November 2 development notes' values that client 1.60.1.70755 still lacks: Vengeance stacks
// up to 5 times (SpellAuraOptions 247421 CumulativeAura 5), and Two-Handed Weapon Specialization
// raises Holy damage too by 2% a rank with a two-hander (SpellEffect 697263 school mask 3), counted
// once on the seals and Holy Strike (the Holy school carries it, not their own multipliers).
func TestHotfix113542(t *testing.T) {
	_, veng := retSim(t, "--000000000000003", arcaniteReaper)
	if got := veng.GetAura("Vengeance Proc").MaxStacks; got != 5 {
		t.Errorf("Vengeance max stacks = %d, want 5", got)
	}

	_, base := retSim(t, "--00000001", arcaniteReaper)
	_, twoHand := retSim(t, "--000000010003", arcaniteReaper)
	_, oneHand := retSim(t, "--000000010003", thunderfury)
	holy := func(ret *RetributionPaladin) float64 {
		return ret.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly]
	}
	if got := holy(twoHand) / holy(base); got < 1.0599 || got > 1.0601 {
		t.Errorf("Holy damage with 3/3 and a two-hander = %.4f, want 1.06", got)
	}
	if got := holy(oneHand) / holy(base); got != 1 {
		t.Errorf("Holy damage with 3/3 and a one-hander = %.4f, want 1", got)
	}
	for _, id := range []int32{20424, 25742} { // Seal of Command's proc, Seal of Righteousness rank 9's
		b, s := base.GetSpell(core.ActionID{SpellID: id}), twoHand.GetSpell(core.ActionID{SpellID: id})
		if b == nil || s == nil {
			continue
		}
		if s.DamageMultiplier != b.DamageMultiplier {
			t.Errorf("spell %d's own multiplier %.4f with 3/3, want %.4f (the school carries it)", id, s.DamageMultiplier, b.DamageMultiplier)
		}
	}
}
