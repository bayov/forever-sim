package mage

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Forever's mage-only staff imbue scrolls, from the Comprehension skill (client 1.60.1.70338,
// items 277500 to 277502). They go in the class imbue slot (SpellEffect 360, like poisons and
// warlock stones), so they stack with a weapon oil, and only on a staff, from level 46.
func (mage *Mage) applyStaffImbue() {
	if mage.Level < 46 || mage.Equipment.MainHand().WeaponType != proto.WeaponType_WeaponTypeStaff {
		return
	}
	switch mage.Options.StaffImbue {
	case proto.Mage_Options_ImbueGreaterFlame:
		// 1306581: +20 Fire damage.
		mage.AddStat(stats.FirePower, 20)
	case proto.Mage_Options_ImbueGreaterFrost:
		// 1306582: +20 Frost damage.
		mage.AddStat(stats.FrostPower, 20)
	case proto.Mage_Options_ImbuePrecision:
		// 1306583: +1% spell crit.
		mage.AddStat(stats.SpellCrit, 1*core.SpellCritRatingPerCritChance)
	}
}
