//go:build practice

package practice

import (
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The sheet has the character's stats with the setup's buffs, the weapons' damage lines and the
// worn items.
func TestSheet(t *testing.T) {
	s := newEnhSession(t, Settings{Mode: "free", LengthMs: 60000, PingMs: 0})
	sheet := s.Sheet()
	if sheet.Stats[stats.Strength] <= sheet.Base[stats.Strength] {
		t.Errorf("strength %v is not above the base %v", sheet.Stats[stats.Strength], sheet.Base[stats.Strength])
	}
	if len(sheet.Weapons) == 0 || sheet.Weapons[0].Slot != "mainHand" || sheet.Weapons[0].Min <= 0 || sheet.Weapons[0].Max < sheet.Weapons[0].Min {
		t.Errorf("weapons %+v", sheet.Weapons)
	}
	if len(sheet.Items) < 10 {
		t.Errorf("only %d items", len(sheet.Items))
	}
	if len(sheet.Pseudo) <= int(proto.PseudoStat_PseudoStatAxesSkill) {
		t.Errorf("pseudo stats %v", sheet.Pseudo)
	}
	t.Logf("str %.0f agi %.0f ap %.0f crit %.2f hit %.2f weapons %+v", sheet.Stats[stats.Strength], sheet.Stats[stats.Agility], sheet.Stats[stats.AttackPower], sheet.Stats[stats.MeleeCrit], sheet.Stats[stats.MeleeHit], sheet.Weapons)
}
