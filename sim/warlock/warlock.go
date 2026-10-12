package warlock

import (
	"github.com/wowsims/classic/sim/common/guardians"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

var TalentTreeSizes = [3]int{17, 19, 16}

const (
	WarlockFlagAffliction  = core.SpellFlagAgentReserved1
	WarlockFlagDemonology  = core.SpellFlagAgentReserved2
	WarlockFlagDestruction = core.SpellFlagAgentReserved3

	SpellFlagWarlock = WarlockFlagAffliction | WarlockFlagDemonology | WarlockFlagDestruction
)

const (
	SpellCode_WarlockNone int32 = iota

	SpellCode_WarlockConflagrate
	SpellCode_WarlockCorruption
	SpellCode_WarlockBaneOfAgony
	SpellCode_WarlockBaneOfDoom
	SpellCode_WarlockDeathCoil
	SpellCode_WarlockDemonicSacrifice
	SpellCode_WarlockDrainLife
	SpellCode_WarlockDrainSoul
	SpellCode_WarlockImmolate
	SpellCode_WarlockLifeTap
	SpellCode_WarlockSearingPain
	SpellCode_WarlockShadowBolt
	SpellCode_WarlockShadowburn
	SpellCode_WarlockSoulFire
	SpellCode_WarlockSiphonLife
	SpellCode_WarlockWrack
	SpellCode_WarlockIncinerate
)

type Warlock struct {
	core.Character
	Talents *proto.WarlockTalents
	Options *proto.WarlockOptions

	BasePets   []*WarlockPet
	ActivePet  *WarlockPet
	Felhunter  *WarlockPet
	Imp        *WarlockPet
	Succubus   *WarlockPet
	Voidwalker *WarlockPet

	// Doomguard *DoomguardPet
	// Infernal  *InfernalPet

	Conflagrate []*core.Spell
	Corruption  []*core.Spell
	Incinerate  []*core.Spell
	Wrack       *core.Spell
	DrainSoul   []*core.Spell
	Immolate    []*core.Spell
	LifeTap     []*core.Spell
	SearingPain []*core.Spell
	ShadowBolt  []*core.Spell
	Shadowburn  []*core.Spell
	SoulFire    []*core.Spell
	DrainLife   []*core.Spell
	RainOfFire  []*core.Spell
	SiphonLife  []*core.Spell
	DeathCoil   []*core.Spell

	ActiveCurseAura          core.AuraArray
	ActiveBaneAura           core.AuraArray
	CurseOfElements          *core.Spell
	CurseOfElementsAuras     core.AuraArray
	CurseOfRecklessness      *core.Spell
	CurseOfRecklessnessAuras core.AuraArray
	CurseOfWeakness          *core.Spell
	CurseOfWeaknessAuras     core.AuraArray
	CurseOfTongues           *core.Spell
	CurseOfTonguesAuras      core.AuraArray
	BaneOfAgony              []*core.Spell
	BaneOfDoom               *core.Spell
	AmplifyCurse             *core.Spell

	soulFireCD    *core.Timer
	shadowburnCD  *core.Timer
	conflagrateCD *core.Timer

	// Track all DoT spells for effecrs that add multipliers based on active effects
	DoTSpells         []*core.Spell
	DebuffSpells      []*core.Spell
	SummonDemonSpells []*core.Spell

	AmplifyCurseAura        *core.Aura
	ImprovedShadowBoltAuras core.AuraArray
	SoulLinkAura            *core.Aura
	MasterDemonologistAura  *core.Aura
	DemonicSacrificeAuras   map[*WarlockPet]*core.Aura
	DemonicBrandAuras       core.AuraArray
	DecimationAura          *core.Aura

	// Talent effects that scale the warlock's damage against one target, from the target's
	// side (Improved Shadow Bolt, Wrack). See applyTargetDamageModifiers.
	targetDamageModifiers []func(spell *core.Spell, target *core.Unit) float64

	sim *core.Simulation

	// Extra Life Tap mana from items (the Demonheart Raiment 5 piece bonus), on top of Improved Life Tap.
	lifeTapBonus float64
}

func (warlock *Warlock) GetCharacter() *core.Character {
	return &warlock.Character
}

func (warlock *Warlock) GetWarlock() *Warlock {
	return warlock
}

func (warlock *Warlock) Initialize() {
	warlock.registerCorruptionSpell()
	warlock.registerImmolateSpell()
	warlock.registerShadowBoltSpell()
	warlock.registerLifeTapSpell()
	warlock.registerSoulFireSpell()
	warlock.registerShadowBurnSpell()
	// warlock.registerSeedSpell()
	warlock.registerDrainSoulSpell()
	warlock.registerConflagrateSpell()
	warlock.registerIncinerateSpell()
	warlock.registerSiphonLifeSpell()
	warlock.registerWrackSpell()
	warlock.registerSearingPainSpell()
	// warlock.registerInfernoSpell()
	// warlock.registerBlackBook()
	warlock.registerDrainLifeSpell()
	warlock.registerRainOfFireSpell()
	warlock.registerDeathCoilSpell()

	warlock.registerCurseOfElementsSpell()
	warlock.registerCurseOfRecklessnessSpell()
	warlock.registerBaneOfAgonySpell()
	warlock.registerAmplifyCurseSpell()
	warlock.registerBaneOfDoomSpell()
	warlock.registerSummonDemon()

	warlock.registerPetAbilities()

	warlock.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Flags.Matches(SpellFlagWarlock) {
			return
		}

		if !spell.Flags.Matches(core.SpellFlagChanneled) && len(spell.Dots()) > 0 {
			warlock.DoTSpells = append(warlock.DoTSpells, spell)
		} else if len(spell.RelatedAuras) > 0 {
			warlock.DebuffSpells = append(warlock.DebuffSpells, spell)
		}
	})
}

func (warlock *Warlock) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	raidBuffs.BloodPact = max(raidBuffs.BloodPact, core.MakeTristateValue(
		warlock.Options.Summon == proto.WarlockOptions_Imp,
		// Forever's Improved Imp no longer improves Blood Pact.
		false,
	))
}

func (warlock *Warlock) Reset(sim *core.Simulation) {
	warlock.sim = sim
	warlock.applyTargetDamageModifiers()
	warlock.setDefaultActivePet()
	warlock.ActiveCurseAura = make([]*core.Aura, len(sim.Environment.AllUnits))
	warlock.ActiveBaneAura = make([]*core.Aura, len(sim.Environment.AllUnits))

	// warlock.ItemSwap.SwapItems(sim, []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand,
	// 	proto.ItemSlot_ItemSlotOffHand, proto.ItemSlot_ItemSlotRanged}, false)
	// warlock.setupCooldowns(sim)
}

func NewWarlock(character *core.Character, options *proto.Player, warlockOptions *proto.WarlockOptions) *Warlock {
	warlock := &Warlock{
		Character: *character,
		Talents:   &proto.WarlockTalents{},
		Options:   warlockOptions,
	}
	core.FillTalentsProto(warlock.Talents.ProtoReflect(), options.TalentsString, TalentTreeSizes)
	warlock.EnableManaBar()

	warlock.AddStatDependency(stats.Strength, stats.AttackPower, core.APPerStrength[character.Class])
	warlock.AddStatDependency(stats.Agility, stats.MeleeCrit, core.CritPerAgi(character.Class, character.Level)*core.CritRatingPerCritChance)
	warlock.AddStatDependency(stats.Agility, stats.Dodge, core.DodgePerAgi(character.Class, character.Level)*core.DodgeRatingPerDodgeChance)
	warlock.AddStatDependency(stats.Intellect, stats.SpellCrit, core.CritPerInt(character.Class, character.Level)*core.SpellCritRatingPerCritChance)
	warlock.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	switch warlock.Options.Armor {
	case proto.WarlockOptions_DemonArmor:
		warlock.applyDemonArmor()
	}

	warlock.registerPets()
	warlock.setDefaultActivePet()

	guardians.ConstructGuardians(&warlock.Character)

	return warlock
}

func (warlock *Warlock) OnGCDReady(_ *core.Simulation) {
}

// Agent is a generic way to access underlying warlock on any of the agents.
type WarlockAgent interface {
	GetWarlock() *Warlock
}

func isWarlockSpell(spell *core.Spell) bool {
	return spell.Flags.Matches(WarlockFlagAffliction) || spell.Flags.Matches(WarlockFlagDemonology) || spell.Flags.Matches(WarlockFlagDestruction)
}
