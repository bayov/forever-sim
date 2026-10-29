package compsearch

import (
	"github.com/wowsims/classic/sim/core/proto"
)

// Preset is a tank or healer placeholder (docs/raid_sim.md, Tank and healer presets): a sensible raid
// build with fixed talents, which gives the raid what its build gives and deals no simulated damage.
type Preset struct {
	ID      proto.RaidPreset
	Name    string
	Role    proto.RaidRole
	Class   proto.Class
	Talents string // the sim's per-tree digit format, checked against client 70630's trees
	Gives   Gift
	// A paladin chooses one judgement and gives one Greater Blessing for each class; a shaman chooses a
	// totem per element.
	Paladin bool
	Shaman  bool
	Package proto.RaidPackage
	PI      bool
	// Innervate: the Restoration druid's, and the Feral tank's when switched on.
	Innervate         bool
	InnervateOptional bool
}

var Presets = []*Preset{
	{ID: proto.RaidPreset_RaidPresetProtectionWarrior, Name: "Protection warrior", Role: proto.RaidRole_RaidRoleTank,
		Class: proto.Class_ClassWarrior, Talents: "35--251533120331221351", Gives: GiftBattleShout | GiftSunderTank},
	{ID: proto.RaidPreset_RaidPresetFeralTank, Name: "Feral tank", Role: proto.RaidRole_RaidRoleTank,
		Class: proto.Class_ClassDruid, Talents: "-55220302020132012551-055", Gives: GiftLotP | GiftGotW | GiftMinorArmor,
		InnervateOptional: true},
	{ID: proto.RaidPreset_RaidPresetProtectionPaladin, Name: "Protection paladin", Role: proto.RaidRole_RaidRoleTank,
		Class: proto.Class_ClassPaladin, Talents: "1-5530513321301551-052", Paladin: true},
	{ID: proto.RaidPreset_RaidPresetWardenShaman, Name: "Warden shaman", Role: proto.RaidRole_RaidRoleTank,
		Class: proto.Class_ClassShaman, Talents: "35050013-0502303313551102", Shaman: true, Package: proto.RaidPackage_RaidPackageManaSpring},
	{ID: proto.RaidPreset_RaidPresetHolyPaladin, Name: "Holy paladin", Role: proto.RaidRole_RaidRoleHealer,
		Class: proto.Class_ClassPaladin, Talents: "05320103225101051-500250302003", Paladin: true},
	{ID: proto.RaidPreset_RaidPresetRestorationShaman, Name: "Restoration shaman", Role: proto.RaidRole_RaidRoleHealer,
		Class: proto.Class_ClassShaman, Talents: "-005102-5530500315513151", Shaman: true, Package: proto.RaidPackage_RaidPackageRestoration},
	{ID: proto.RaidPreset_RaidPresetDisciplinePriest, Name: "Discipline priest", Role: proto.RaidRole_RaidRoleHealer,
		Class: proto.Class_ClassPriest, Talents: "005005231330011531-2310510303", Gives: GiftPoS | GiftFort, PI: true},
	{ID: proto.RaidPreset_RaidPresetHolyPriest, Name: "Holy priest", Role: proto.RaidRole_RaidRoleHealer,
		Class: proto.Class_ClassPriest, Talents: "00510223123-03405103030121531", Gives: GiftPoS | GiftFort},
	{ID: proto.RaidPreset_RaidPresetRestorationDruid, Name: "Restoration druid", Role: proto.RaidRole_RaidRoleHealer,
		Class: proto.Class_ClassDruid, Talents: "05302001--5052035153113051", Gives: GiftGotW | GiftMinorArmor, Innervate: true},
}

var presetByID = func() map[proto.RaidPreset]*Preset {
	m := map[proto.RaidPreset]*Preset{}
	for _, p := range Presets {
		m[p.ID] = p
	}
	return m
}()

// PresetByID is a preset.
func PresetByID(id proto.RaidPreset) *Preset { return presetByID[id] }

// PresetsFor is every preset of a role.
func PresetsFor(role proto.RaidRole) []*Preset {
	var out []*Preset
	for _, p := range Presets {
		if p.Role == role {
			out = append(out, p)
		}
	}
	return out
}
