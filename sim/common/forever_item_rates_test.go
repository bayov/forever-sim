//go:build with_db

package common_test

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/common/guardians"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

// A level 60 talentless warrior on ruleset holding weapon in the main hand, against a target of family.
func newWarriorWith(t *testing.T, ruleset proto.Ruleset, weapon int32, family proto.MobFamily) (*core.Simulation, *core.Character) {
	t.Helper()
	specs := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range specs {
		specs[i] = &proto.ItemSpec{}
	}
	specs[proto.ItemSlot_ItemSlotMainHand].Id = weapon
	player := &proto.Player{
		Race:      proto.Race_RaceHuman,
		Class:     proto.Class_ClassWarrior,
		Level:     60,
		Equipment: &proto.EquipmentSpec{Items: specs},
		Spec:      &proto.Player_Warrior{Warrior: &proto.Warrior{Options: &proto.Warrior_Options{}}},
		Rotation:  &proto.APLRotation{},
	}
	target := core.NewDefaultTarget()
	target.MobFamily = family
	s := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{target}},
		SimOptions: &proto.SimOptions{Ruleset: ruleset, RandomSeed: 1},
	}, simsignals.CreateSignals())
	s.Reset()
	s.PrePull()
	return s, s.Raid.Parties[0].Players[0].GetCharacter()
}

// The share of n landed main hand white swings that cast a spell, other than those ignored, or put up
// an aura.
func procShare(s *core.Simulation, c *core.Character, n int, ignore ...int32) float64 {
	mhAuto := c.AutoAttacks.MHAuto()
	casts := func() (total int32) {
		for _, spell := range c.Spellbook {
			if spell == mhAuto || slices.Contains(ignore, spell.ActionID.SpellID) {
				continue
			}
			for _, m := range spell.SpellMetrics() {
				total += m.Casts
			}
		}
		return total
	}
	procs := 0
	for range n {
		for _, aura := range c.GetAuras() {
			if aura.IsActive() && !aura.IsPermanent() {
				aura.Deactivate(s)
			}
		}
		before := casts()
		c.OnSpellHitDealt(s, mhAuto, &core.SpellResult{Target: s.GetTargetUnit(0), Outcome: core.OutcomeHit})
		proc := casts() > before
		for _, aura := range c.GetAuras() {
			if aura.IsActive() && !aura.IsPermanent() {
				proc = true
			}
		}
		if proc {
			procs++
		}
	}
	return float64(procs) / float64(n)
}

// The weapons our level 60 presets wear proc under Forever at the procs per minute wowsims/forever
// measured on Warcraft Logs, priced on the weapon's speed, and at their Classic rate otherwise.
func TestForeverWeaponProcRates(t *testing.T) {
	const n = 40000
	for _, c := range []struct {
		name             string
		item             int32
		forever, classic float64 // procs per minute
		ignore           []int32 // the item's spells that are not its proc
	}{
		{"Felstriker (EV-0136)", common.Felstriker, 0.55, 1, nil},
		{"Brain Hacker (EV-0132)", common.BrainHacker, 0.95, 0, nil},
		{"Destiny (EV-0132)", common.Destiny, 1.05, 1, nil},
		{"Sword of Zeal (EV-0175)", common.SwordOfZeal, 1.95, 1, nil},
		{"Iceblade Hacker (EV-0136)", common.IcebladeHacker, 2, 0, []int32{1298414}},
		{"Heart of Wyrmthalak (EV-0133)", common.HeartOfWyrmthalak, 1.95, 0.4, nil},
		// Weapons the sim had no effect for (forever_item_procs_ev.go).
		{"Ripsaw (EV-0022)", 9478, 2, 0, nil},
		{"Sword of Dismemberment (EV-0024)", 287005, 0.6, 0, nil},
		{"Winter's Bite (EV-0186)", 10623, 2.95, 0, nil},
		{"Night Reaver (EV-0136)", 1318, 1.7, 0, nil},
	} {
		for _, r := range []struct {
			ruleset proto.Ruleset
			ppm     float64
		}{{proto.Ruleset_RulesetForever, c.forever}, {proto.Ruleset_RulesetClassic, c.classic}} {
			s, ch := newWarriorWith(t, r.ruleset, c.item, proto.MobFamily_MobFamilyUnknown)
			want := r.ppm * ch.AutoAttacks.MH().SwingSpeed / 60
			got := procShare(s, ch, n, c.ignore...)
			if math.Abs(got-want)*n > 4*math.Sqrt(want*n)+1 {
				t.Errorf("%s on %v: procs on %.4f of hits, want %.4f (%.2f PPM)", c.name, r.ruleset, got, want, r.ppm)
			}
		}
	}
}

// EV-0119: Ironfoe procs on 3% of landed hits under Forever, 6% against an Orc, and at 0.8 PPM on
// Classic.
func TestForeverIronfoeChance(t *testing.T) {
	const n = 40000
	for _, c := range []struct {
		name    string
		ruleset proto.Ruleset
		family  proto.MobFamily
		want    float64
	}{
		{"Forever", proto.Ruleset_RulesetForever, proto.MobFamily_MobFamilyUnknown, 0.03},
		{"Forever vs an Orc", proto.Ruleset_RulesetForever, proto.MobFamily_MobFamilyOrc, 0.06},
		{"Classic", proto.Ruleset_RulesetClassic, proto.MobFamily_MobFamilyUnknown, 0.8 * 2.4 / 60},
	} {
		s, ch := newWarriorWith(t, c.ruleset, common.Ironfoe, c.family)
		got := procShare(s, ch, n)
		if math.Abs(got-c.want)*n > 4*math.Sqrt(c.want*n) {
			t.Errorf("Ironfoe on %s: procs on %.4f of hits, want %.4f", c.name, got, c.want)
		}
	}
}

// EV-0021: under Forever Dragon's Call procs 4.5 times a minute behind a 45 s cooldown, and its whelp
// is its owner's level.
func TestForeverDragonsCall(t *testing.T) {
	if common.DragonsCallForeverPPM != 4.5 || common.DragonsCallForeverICD != 45*time.Second {
		t.Errorf("Dragon's Call: %v PPM, %v cooldown, want 4.5 and 45s", common.DragonsCallForeverPPM, common.DragonsCallForeverICD)
	}
	for _, c := range []struct {
		ruleset proto.Ruleset
		level   int32
	}{{proto.Ruleset_RulesetForever, 60}, {proto.Ruleset_RulesetClassic, 55}} {
		_, ch := newWarriorWith(t, c.ruleset, common.DragonsCall, proto.MobFamily_MobFamilyUnknown)
		found := false
		for _, pet := range ch.PetAgents {
			if whelp, ok := pet.(*guardians.EmeraldDragonWhelp); ok {
				found = true
				if whelp.Level != c.level {
					t.Errorf("whelp on %v is level %d, want %d", c.ruleset, whelp.Level, c.level)
				}
			}
		}
		if !found {
			t.Errorf("no whelp on %v", c.ruleset)
		}
	}
}

// EV-0025: Blade of Eternal Darkness procs 1.8 times a minute on the cast time, at least 1.5 s.
func TestForeverBladeOfEternalDarknessChance(t *testing.T) {
	for _, c := range []struct {
		cast time.Duration
		want float64
	}{{0, 0.045}, {time.Second, 0.045}, {3 * time.Second, 0.09}} {
		if got := common.BladeOfEternalDarknessChance(c.cast); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%v cast: %v, want %v", c.cast, got, c.want)
		}
	}
}

// The Forever item details wowsims/forever read off Warcraft Logs: Flame Wrath's ring (EV-0151),
// Masterwork Stormhammer's jumps (EV-0184), Barovian Family Sword's heal (EV-0179) and Shahram's
// spell weights (EV-0125).
func TestForeverItemDetails(t *testing.T) {
	if math.Abs(common.FlameWrathForeverMin-133.47) > 0.01 || math.Abs(common.FlameWrathForeverMax-174.53) > 0.01 {
		t.Errorf("Flame Wrath: %v to %v, want 133.47 to 174.53", common.FlameWrathForeverMin, common.FlameWrathForeverMax)
	}
	if common.MasterworkStormhammerForeverJump != 0.7 {
		t.Errorf("Masterwork Stormhammer jump: %v, want 0.7", common.MasterworkStormhammerForeverJump)
	}
	if common.BarovianFamilySwordForeverTickHeal != 49 {
		t.Errorf("Barovian Family Sword heal: %v, want 49", common.BarovianFamilySwordForeverTickHeal)
	}
	// Curse, Might, Fist, Blessing, Will, Flames.
	want := []float64{2, 2, 2, 2, 5, 7}
	got := make([]float64, 6)
	const n = 2000
	for i := range n {
		got[common.ShahramForeverSpell((float64(i)+0.5)/n)]++
	}
	for i := range want {
		if math.Abs(got[i]/n-want[i]/20) > 1e-9 {
			t.Errorf("Shahram spell %d: %v, want %v", i, got[i]/n, want[i]/20)
		}
	}
}
