package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/common/guardians"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

var TalentTreeSizes = [3]int{16, 18, 16}

const (
	SpellFlagShaman    = core.SpellFlagAgentReserved1
	SpellFlagTotem     = core.SpellFlagAgentReserved2
	SpellFlagLightning = core.SpellFlagAgentReserved3
)

// Every rank of an ability is one family, so a rotation written at 60 that names the top
// rank finds the rank a lower level shaman knows.
func init() {
	core.RegisterSpellRanks(LightningBoltSpellId[1:]...)
	core.RegisterSpellRanks(ChainLightningSpellId[1:]...)
	core.RegisterSpellRanks(EarthShockSpellId[1:]...)
	core.RegisterSpellRanks(FlameShockSpellId[1:]...)
	core.RegisterSpellRanks(FrostShockSpellId[1:]...)
	core.RegisterSpellRanks(LightningShieldSpellId[1:]...)
	core.RegisterSpellRanks(SearingTotemSpellId[1:]...)
	core.RegisterSpellRanks(MagmaTotemSpellId[1:]...)
	core.RegisterSpellRanks(FireNovaSpellId[1:]...)
	core.RegisterSpellRanks(FlametongueTotemSpellId[1:]...)
	core.RegisterSpellRanks(StrengthOfEarthTotemSpellId[1:]...)
	core.RegisterSpellRanks(StoneskinTotemSpellId[1:]...)
	core.RegisterSpellRanks(WindfuryTotemSpellId[1:]...)
	core.RegisterSpellRanks(GraceOfAirTotemSpellId[1:]...)
	core.RegisterSpellRanks(WindwallTotemSpellId[1:]...)
	core.RegisterSpellRanks(HealingStreamTotemSpellId[1:]...)
	core.RegisterSpellRanks(ManaSpringTotemSpellId[1:]...)
	core.RegisterSpellRanks(LavaBurstSpellId[1:]...)
}

func NewShaman(character *core.Character, talents string) *Shaman {
	shaman := &Shaman{
		Character: *character,
		Talents:   &proto.ShamanTalents{},
	}

	core.FillTalentsProto(shaman.Talents.ProtoReflect(), talents, TalentTreeSizes)
	shaman.EnableManaBar()

	// Add Shaman stat dependencies
	shaman.AddStatDependency(stats.Strength, stats.AttackPower, core.APPerStrength[character.Class])
	shaman.AddStatDependency(stats.Agility, stats.MeleeCrit, core.CritPerAgi(character.Class, character.Level)*core.CritRatingPerCritChance)
	shaman.AddStatDependency(stats.Agility, stats.Dodge, core.DodgePerAgi(character.Class, character.Level)*core.DodgeRatingPerDodgeChance)
	shaman.AddStatDependency(stats.Intellect, stats.SpellCrit, core.CritPerInt(character.Class, character.Level)*core.SpellCritRatingPerCritChance)
	shaman.AddStatDependency(stats.BonusArmor, stats.Armor, 1)
	shaman.PseudoStats.BlockValuePerStrength = .05 // 20 str = 1 block

	shaman.ApplyRockbiterImbue(shaman.getImbueProcMask(proto.WeaponImbue_RockbiterWeapon))
	shaman.ApplyFlametongueImbue(shaman.getImbueProcMask(proto.WeaponImbue_FlametongueWeapon))
	shaman.ApplyFrostbrandImbue(shaman.getImbueProcMask(proto.WeaponImbue_FrostbrandWeapon))
	shaman.ApplyWindfuryImbue(shaman.getImbueProcMask(proto.WeaponImbue_WindfuryWeapon))

	guardians.ConstructGuardians(&shaman.Character)

	return shaman
}

func (shaman *Shaman) getImbueProcMask(imbue proto.WeaponImbue) core.ProcMask {
	mask := core.ProcMaskUnknown
	if shaman.HasMHWeapon() && (shaman.Consumes.MainHandImbue == imbue || shaman.ShamanImbue == imbue) {
		mask |= core.ProcMaskMeleeMH
	}
	return mask
}

// setTotemWeaponBuff gives the weapon's totem slot to the buff of the totem the shaman
// just placed. The weapon holds one totem buff, so it takes over from the other totem's
// buff and from another shaman's totem in the raid buffs.
//
// sameKindImbue is the imbue that turns this totem's buff off for us (Windfury Weapon for
// Windfury Totem). When our main hand has it, the buff does nothing for us and it doesn't
// take the slot either. That is the raid setup where we keep Windfury Totem down for the
// group and Flametongue Totem down for ourselves, next to our own Windfury Weapon.
func (shaman *Shaman) setTotemWeaponBuff(sim *core.Simulation, aura *core.Aura, sameKindImbue proto.WeaponImbue) {
	if shaman.getImbueProcMask(sameKindImbue) != core.ProcMaskUnknown {
		return
	}
	if shaman.totemWeaponBuff != nil && shaman.totemWeaponBuff != aura {
		shaman.totemWeaponBuff.Deactivate(sim)
	}
	shaman.totemWeaponBuff = aura
	aura.Activate(sim)
	if shaman.RaidTotemWeaponBuff != nil {
		shaman.RaidTotemWeaponBuff.Deactivate(sim)
	}
}

// clearTotemWeaponBuff takes the buff off when its totem goes down. When it still held the
// slot, another shaman's totem buff from the raid buffs comes back.
func (shaman *Shaman) clearTotemWeaponBuff(sim *core.Simulation, aura *core.Aura) {
	aura.Deactivate(sim)
	if shaman.totemWeaponBuff != aura {
		return
	}
	shaman.totemWeaponBuff = nil
	if shaman.RaidTotemWeaponBuff != nil {
		shaman.RaidTotemWeaponBuff.Activate(sim)
	}
}

// ApplyShamanImbue puts ShamanImbue on the main hand, the way NewShaman does for a
// shaman imbue in the consumes. It runs after the spec has set ShamanImbue.
//
// We skip it when the consumes already carry the same imbue, so that Rockbiter's
// attack power is not added twice.
func (shaman *Shaman) ApplyShamanImbue() {
	if !shaman.HasMHWeapon() || shaman.Consumes.MainHandImbue == shaman.ShamanImbue {
		return
	}
	switch shaman.ShamanImbue {
	case proto.WeaponImbue_RockbiterWeapon:
		shaman.ApplyRockbiterImbue(core.ProcMaskMeleeMH)
	case proto.WeaponImbue_FlametongueWeapon:
		shaman.ApplyFlametongueImbue(core.ProcMaskMeleeMH)
	case proto.WeaponImbue_FrostbrandWeapon:
		shaman.ApplyFrostbrandImbue(core.ProcMaskMeleeMH)
	case proto.WeaponImbue_WindfuryWeapon:
		shaman.ApplyWindfuryImbue(core.ProcMaskMeleeMH)
	}
}

// Indexes into NextTotemDrops for self buffs
const (
	AirTotem int = iota
	EarthTotem
	FireTotem
	WaterTotem
)

const (
	SpellCode_ShamanNone int32 = iota

	SpellCode_ShamanChainHeal
	SpellCode_ShamanChainLightning
	SpellCode_ShamanEarthShock
	SpellCode_ShamanFireNova
	SpellCode_ShamanFlameShock
	SpellCode_ShamanFrostShock
	SpellCode_ShamanHealingWave
	SpellCode_ShamanLavaBurst
	SpellCode_ShamanLesserHealingWave
	SpellCode_ShamanLightningBolt
	SpellCode_ShamanLightningShield
	SpellCode_ShamanMagmaTotem
	SpellCode_ShamanSearingTotem
	SpellCode_ShamanStormstrike
	SpellCode_ShamanWaterShield
)

// Shaman represents a shaman character.
type Shaman struct {
	core.Character

	Talents *proto.ShamanTalents

	// Spells
	ChainHeal              []*core.Spell
	ChainLightning         []*core.Spell
	ChainLightningOverload []*core.Spell
	EarthShield            *core.Spell
	EarthShock             []*core.Spell
	FireNova               []*core.Spell
	FlametongueTotem       []*core.Spell
	FlameShock             []*core.Spell
	FrostShock             []*core.Spell
	GraceOfAirTotem        []*core.Spell
	HealingStreamTotem     []*core.Spell
	HealingWave            []*core.Spell
	LavaBurst              *core.Spell
	LesserHealingWave      []*core.Spell
	LightningBolt          []*core.Spell
	LightningBoltOverload  []*core.Spell
	LightningShield        []*core.Spell
	LightningShieldProcs   []*core.Spell // The damage component of lightning shield is a separate spell
	MagmaTotem             []*core.Spell
	ManaSpringTotem        []*core.Spell
	ManaTideTotem          []*core.Spell
	SearingTotem           []*core.Spell
	StoneskinTotem         []*core.Spell
	Stormstrike            *core.Spell
	// The Forever Stormstrike mark on each target, nil under Classic.
	stormstrikeMarks     core.AuraArray
	StrengthOfEarthTotem []*core.Spell
	TremorTotem          *core.Spell
	WaterShield          *core.Spell
	WindfuryTotem        []*core.Spell
	WindfuryWeaponMH     *core.Spell
	WindfuryWeaponOH     *core.Spell
	WindwallTotem        []*core.Spell

	// Auras
	ClearcastingAura     *core.Aura
	LightningShieldAuras []*core.Aura
	MaelstromWeaponAura  *core.Aura
	WaterShieldAura      *core.Aura
	// One per Mana Spring rank, so Mana Tide Totem can take its place.
	manaSpringAuras []*core.Aura

	// Set by Forever's Totem of the Storm, Lightning Bolt then also triggers Maelstrom Weapon at half the chance.
	lightningBoltTriggersMaelstrom bool

	// Totems
	ActiveTotems     [4]*core.Spell
	ActiveTotemBuffs [4]*core.Aura
	TotemExpirations [4]time.Duration // The expiration time of each totem (earth, air, fire, water).

	EarthTotems []*core.Spell
	FireTotems  []*core.Spell
	WaterTotems []*core.Spell
	AirTotems   []*core.Spell
	Totems      *proto.ShamanTotems

	// The buff of the shaman's own Windfury or Flametongue Totem that holds the weapon's
	// totem slot, nil without one.
	totemWeaponBuff *core.Aura

	// Shield
	ActiveShield     *core.Spell // Tracks the Shaman's active shield spell
	ActiveShieldAura *core.Aura
	// What each shield's aura does when an enemy lands a direct hit on us.
	shieldHitTaken map[*core.Aura]func(sim *core.Simulation, attacker *core.Unit)
	// Direct hits taken from raid damage per minute, each one reaching our shield. Set from
	// the enhancement options before Initialize.
	RaidDamageHitsPerMinute float64
	// Each iteration's rate is up to this many hits per minute above or below
	// RaidDamageHitsPerMinute, like Duration +/- for the fight length.
	RaidDamageHitsPerMinuteVariation float64
	// The shaman's own main hand imbue in Forever, beside the oil or stone in the
	// consumes. Set from the enhancement options before Initialize.
	ShamanImbue proto.WeaponImbue
	// The totems we have standing when the fight starts. Set from the enhancement options
	// before Initialize.
	StartingTotems *proto.StartingTotems

	ChainLightningBounceCoefficient float64
}

// Implemented by each Shaman spec.
type ShamanAgent interface {
	core.Agent

	// The Shaman controlled by this Agent.
	GetShaman() *Shaman
}

func (shaman *Shaman) GetCharacter() *core.Character {
	return &shaman.Character
}

func (shaman *Shaman) AddRaidBuffs(_ *proto.RaidBuffs) {
	// Buffs are handled explicitly through APLs now
}

func (shaman *Shaman) Initialize() {
	// Core abilities
	shaman.registerChainLightningSpell()
	shaman.registerLavaBurstSpell()
	shaman.registerLightningBoltSpell()
	shaman.shieldHitTaken = map[*core.Aura]func(*core.Simulation, *core.Unit){}
	shaman.registerLightningShieldSpell()
	shaman.registerShocks()
	shaman.registerStormstrikeSpell()
	shaman.registerWaterShieldSpell()
	shaman.registerRaidDamageHits()

	// Imbues
	// In the Initialize due to frost brand adding the aura to the enemy
	shaman.RegisterRockbiterImbue(shaman.getImbueProcMask(proto.WeaponImbue_RockbiterWeapon))
	shaman.RegisterFlametongueImbue(shaman.getImbueProcMask(proto.WeaponImbue_FlametongueWeapon))
	shaman.RegisterWindfuryImbue(shaman.getImbueProcMask(proto.WeaponImbue_WindfuryWeapon))
	shaman.RegisterFrostbrandImbue(shaman.getImbueProcMask(proto.WeaponImbue_FrostbrandWeapon))

	// Totems
	shaman.registerStrengthOfEarthTotemSpell()
	shaman.registerStoneskinTotemSpell()
	shaman.registerTremorTotemSpell()
	shaman.registerSearingTotemSpell()
	shaman.registerMagmaTotemSpell()
	shaman.registerFireNovaSpell()
	shaman.registerFlametongueTotemSpell()
	shaman.registerHealingStreamTotemSpell()
	shaman.registerManaSpringTotemSpell()
	shaman.registerManaTideTotemSpell()
	shaman.registerWindfuryTotemSpell()
	shaman.registerGraceOfAirTotemSpell()
	shaman.registerWindwallTotemSpell()
	shaman.registerStartingTotems()
}

func (shaman *Shaman) Reset(_ *core.Simulation) {
	shaman.ActiveShield = nil
	shaman.ActiveShieldAura = nil
	shaman.totemWeaponBuff = nil

	for i := range []int{EarthTotem, FireTotem, WaterTotem, AirTotem} {
		shaman.ActiveTotems[i] = nil
		shaman.TotemExpirations[i] = 0
		shaman.ActiveTotemBuffs[i] = nil
	}
}
