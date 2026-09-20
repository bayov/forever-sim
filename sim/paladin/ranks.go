package paladin

import "github.com/wowsims/classic/sim/core"

// The learn level and spell ID of every rank of the abilities a rotation names, so a
// rotation written at level 60 finds the rank a lower level paladin knows, and the
// rotation optimizer can name the level's rank in the file it writes.
var (
	SealOfRighteousnessLevel   = [...]int{0, 1, 10, 18, 26, 34, 42, 50, 58}
	SealOfRighteousnessSpellId = [...]int32{0, 20154, 20287, 20288, 20289, 20290, 20291, 20292, 20293}
	SealOfCommandLevel         = [...]int{0, 20, 30, 40, 50, 60}
	SealOfCommandSpellId       = [...]int32{0, 20375, 20915, 20918, 20919, 20920}
	SealOfTheCrusaderLevel     = [...]int{0, 6, 12, 22, 32, 42, 52}
	SealOfTheCrusaderSpellId   = [...]int32{0, 21082, 20162, 20305, 20306, 20307, 20308}
	ConsecrationLevel          = [...]int{0, 20, 30, 40, 50, 60}
	ConsecrationSpellId        = [...]int32{0, 26573, 20116, 20922, 20923, 20924}
	ExorcismLevel              = [...]int{0, 20, 28, 36, 44, 52, 60}
	ExorcismSpellId            = [...]int32{0, 879, 5614, 5615, 10312, 10313, 10314}
	HolyShockLevel             = [...]int{0, 30, 40, 48, 56}
	HolyShockSpellId           = [...]int32{0, 1311606, 20473, 20929, 20930}
	HammerOfWrathLevel         = [...]int{0, 44, 52, 60}
	HammerOfWrathSpellId       = [...]int32{0, 24275, 24274, 24239}

	JudgementSpellId              int32 = 20271
	JudgementOfTheCrusaderSpellId int32 = 20303
)

func init() {
	core.RegisterSpellRanks(SealOfRighteousnessSpellId[1:]...)
	core.RegisterSpellRanks(SealOfCommandSpellId[1:]...)
	core.RegisterSpellRanks(SealOfTheCrusaderSpellId[1:]...)
	core.RegisterSpellRanks(ConsecrationSpellId[1:]...)
	core.RegisterSpellRanks(ExorcismSpellId[1:]...)
	core.RegisterSpellRanks(HolyShockSpellId[1:]...)
	core.RegisterSpellRanks(HammerOfWrathSpellId[1:]...)
	core.RegisterSpellRanks(HolyStrikeSpellId[1:]...)
}
