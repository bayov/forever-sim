package retribution

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Five pieces each of Justice Armor and Justice Battleplate.
var justiceArmorPieces = map[proto.ItemSlot]int32{
	proto.ItemSlot_ItemSlotHead:     280936,
	proto.ItemSlot_ItemSlotShoulder: 280935,
	proto.ItemSlot_ItemSlotChest:    280934,
	proto.ItemSlot_ItemSlotHands:    280931,
	proto.ItemSlot_ItemSlotLegs:     280933,
}
var justiceBattleplatePieces = map[proto.ItemSlot]int32{
	proto.ItemSlot_ItemSlotHead:     280937,
	proto.ItemSlot_ItemSlotShoulder: 280942,
	proto.ItemSlot_ItemSlotChest:    280941,
	proto.ItemSlot_ItemSlotHands:    280938,
	proto.ItemSlot_ItemSlotLegs:     280940,
}

func withWeapon(items map[proto.ItemSlot]int32) map[proto.ItemSlot]int32 {
	gear := map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: arcaniteReaper}
	for slot, id := range items {
		gear[slot] = id
	}
	return gear
}

// TestTenetsOfTheSilverHand checks the libram's spell 1302545: 1% more damage against an Undead
// target and nothing against the default boss.
func TestTenetsOfTheSilverHand(t *testing.T) {
	for _, c := range []struct {
		mobType proto.MobType
		want    float64
	}{
		{proto.MobType_MobTypeUndead, 1.01},
		{proto.MobType_MobTypeUnknown, 1},
	} {
		multiplier := func(ranged int32) float64 {
			_, ret := retSimGear(t, "", withWeapon(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: ranged}), 1, c.mobType)
			return ret.AttackTables[ret.CurrentTarget.UnitIndex][proto.CastType_CastTypeMainHand].DamageDealtMultiplier
		}
		if got := multiplier(249397) / multiplier(0); got < c.want-1e-9 || got > c.want+1e-9 {
			t.Errorf("%v: Tenets of the Silver Hand damage multiplier = %.4f, want %.2f", c.mobType, got, c.want)
		}
	}
}

// TestLibramOfInvocation checks the libram's 5% off the seals with Benediction 5/5 and Twist of
// Light: Seal of Righteousness rank 8 costs 200 x 0.9 x 0.8 x 0.95 = 136.8, and Sanctified
// Judgement 3/3 returns 60% of 200 x (2 - 0.684) = 157.92.
func TestLibramOfInvocation(t *testing.T) {
	sim, ret := retSimGear(t, "--05000030000000001", withWeapon(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: 249442}), 1, proto.MobType_MobTypeUnknown)
	seal := ret.GetSpell(core.ActionID{SpellID: 20293})
	if got := seal.Cost.GetCurrentCost(); got < 136.79 || got > 136.81 {
		t.Errorf("Seal of Righteousness costs %.2f, want 136.8", got)
	}
	castByID(sim, ret, 20293)
	ret.SpendMana(sim, 0.5*ret.MaxMana(), ret.NewManaMetrics(core.ActionID{OtherID: 1}))
	before := ret.CurrentMana()
	castByID(sim, ret, 20271)
	if got := ret.CurrentMana() - before; got < 157.91 || got > 157.93 {
		t.Errorf("Judgement refunded %.2f mana, want 157.92", got)
	}
}

// TestJudgementOfTheCrusaderGloves checks spell 23300 on the PvP gloves: 20 more Holy damage
// taken on Judgement of the Crusader, 181 at rank 6.
func TestJudgementOfTheCrusaderGloves(t *testing.T) {
	for _, gloves := range []int32{16410, 16471, 23274} {
		sim, ret := retSimGear(t, "52003-503-05225331001330311", withWeapon(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotHands: gloves}), 1, proto.MobType_MobTypeUnknown)
		castByID(sim, ret, 20308)
		castByID(sim, ret, 20271)
		if got := ret.CurrentTarget.PseudoStats.SchoolBonusDamageTaken[stats.SchoolIndexHoly]; got < 180.99 || got > 181.01 {
			t.Errorf("%d: Holy damage taken after Judgement of the Crusader = %.2f, want 181", gloves, got)
		}
	}
}

// TestJusticeSets checks Justice Armor (2107) and Justice Battleplate (2108) at 5 pieces, counted
// by the client's set ID: +10 Spirit, +26 healing and +9 spell damage, Holy Shock 1 sec faster;
// +7 Defense, 1.2% expertise and Forbearance 10 sec shorter.
func TestJusticeSets(t *testing.T) {
	const shock = "55003003201121251--05225230001"
	_, base := retSimGear(t, shock, withWeapon(nil), 1, proto.MobType_MobTypeUnknown)
	_, armor := retSimGear(t, shock, withWeapon(justiceArmorPieces), 1, proto.MobType_MobTypeUnknown)
	_, plate := retSimGear(t, shock, withWeapon(justiceBattleplatePieces), 1, proto.MobType_MobTypeUnknown)

	diff := func(ret *RetributionPaladin, stat stats.Stat) float64 {
		return ret.GetStat(stat) - base.GetStat(stat)
	}
	if got := diff(armor, stats.Spirit); got < 9.9 || got > 11.1 { // Divine Intellect and Human Spirit raise it a little
		t.Errorf("Justice Armor Spirit +%.2f, want 10", got)
	}
	if got := diff(armor, stats.HealingPower); got != 26 {
		t.Errorf("Justice Armor healing +%.1f, want 26", got)
	}
	if got := diff(armor, stats.SpellDamage); got != 9 {
		t.Errorf("Justice Armor spell damage +%.1f, want 9", got)
	}
	if got := armor.GetSpell(core.ActionID{SpellID: 20930}).CD.Duration; got != 9*time.Second {
		t.Errorf("Holy Shock cooldown %v with Justice Armor, want 9s", got)
	}
	if got := diff(plate, stats.Defense); got != 7 {
		t.Errorf("Justice Battleplate Defense +%.1f, want 7", got)
	}
	if got := diff(plate, stats.Expertise); got < 1.19 || got > 1.21 {
		t.Errorf("Justice Battleplate expertise +%.2f, want 1.2", got)
	}
	if got := plate.GetAura("Forbearance").Duration; got != 50*time.Second {
		t.Errorf("Forbearance lasts %v with Justice Battleplate, want 50s", got)
	}
	if got := base.GetAura("Forbearance").Duration; got != time.Minute {
		t.Errorf("Forbearance lasts %v, want 1m", got)
	}
}

// TestTwistOfLightOneEcho checks that only one Echo is up and that a seal drops its own Echo:
// Seal of Command to Righteousness leaves an Echo of Command, and back to Command swaps it for an
// Echo of Righteousness. Seal of Command to the Crusader and back to Command leaves no Echo.
func TestTwistOfLightOneEcho(t *testing.T) {
	sim, ret := retSim(t, "--25225331001330301", arcaniteReaper)
	command, righteousness := ret.GetAura("Echo of Command"), ret.GetAura("Echo of Righteousness")
	castByID(sim, ret, 20920)
	castByID(sim, ret, 20293)
	if !command.IsActive() || righteousness.IsActive() {
		t.Fatalf("after Command then Righteousness: Echo of Command %v, Echo of Righteousness %v", command.IsActive(), righteousness.IsActive())
	}
	castByID(sim, ret, 20920)
	if command.IsActive() || !righteousness.IsActive() {
		t.Errorf("back to Command: Echo of Command %v, Echo of Righteousness %v", command.IsActive(), righteousness.IsActive())
	}
	righteousness.Deactivate(sim)
	castByID(sim, ret, 20308)
	if !command.IsActive() {
		t.Fatalf("no Echo of Command after Command then the Crusader")
	}
	castByID(sim, ret, 20920)
	if command.IsActive() {
		t.Errorf("Echo of Command still up under Seal of Command")
	}
}
