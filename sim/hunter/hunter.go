package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/common/guardians"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Forever tree from the beta client (tools/forever_talents/client/hunter.json, brought up to
// build 1.60.1.70338). The client also parks Improved Serpent Sting and a second Lightning Reflexes
// far off the canvas, where no one can take them, so they are not in the tree.
var TalentTreeSizes = [3]int{16, 16, 18}

const (
	SpellFlagShot   = core.SpellFlagAgentReserved1
	SpellFlagStrike = core.SpellFlagAgentReserved2
	SpellFlagSting  = core.SpellFlagAgentReserved3
	SpellFlagTrap   = core.SpellFlagAgentReserved4
)

const (
	SpellCode_HunterNone int32 = iota

	// Shots
	SpellCode_HunterAimedShot
	SpellCode_HunterArcaneShot
	SpellCode_HunterMultiShot
	SpellCode_HunterSniperShot
	SpellCode_HunterSummonHawk

	// Strikes
	SpellCode_HunterRaptorStrike
	SpellCode_HunterRaptorStrikeHit
	SpellCode_HunterMongooseBite
	SpellCode_HunterStriderKick
	SpellCode_HunterWingClip
	SpellCode_HunterLacerate
	SpellCode_HunterLaceratingStrikes

	// Stings
	SpellCode_HunterSerpentSting

	// Traps
	SpellCode_HunterExplosiveTrap
	SpellCode_HunterFreezingTrap
	SpellCode_HunterImmolationTrap

	// Other
	SpellCode_HunterVolley

	// Pet Spells
	SpellCode_HunterPetClaw
	SpellCode_HunterPetBite
	SpellCode_HunterPetLightningBreath
	SpellCode_HunterPetScreech
	SpellCode_HunterPetScorpidPoison
	SpellCode_HunterPetSavageRend
)

func RegisterHunter() {
	core.RegisterAgentFactory(
		proto.Player_Hunter{},
		proto.Spec_SpecHunter,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewHunter(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Hunter)
			if !ok {
				panic("Invalid spec value for Hunter!")
			}
			player.Spec = playerSpec
		},
	)
}

type Hunter struct {
	core.Character

	Talents *proto.HunterTalents
	Options *proto.Hunter_Options

	pet *HunterPet

	AmmoDPS                   float64
	AmmoDamageBonus           float64
	NormalizedAmmoDamageBonus float64

	// Miscellaneous set bonuses that require extra logic inside of spells
	AspectOfTheHawkAPMultiplier float64

	curQueueAura       *core.Aura
	curQueuedAutoSpell *core.Spell

	AimedShot         *core.Spell
	ArcaneShot        *core.Spell
	ExplosiveTrap     *core.Spell
	ImmolationTrap    *core.Spell
	FreezingTrap      *core.Spell
	Lacerate          *core.Spell
	LaceratingStrikes *core.Spell
	MultiShot         *core.Spell
	RapidFire         *core.Spell
	RaptorStrike      *core.Spell
	RaptorStrikeHit   *core.Spell
	MongooseBite      *core.Spell
	SerpentSting      *core.Spell
	SniperShot        *core.Spell
	StriderKick       *core.Spell
	SummonHawk        *core.Spell
	Volley            *core.Spell
	WingClip          *core.Spell

	Shots       []*core.Spell
	Strikes     []*core.Spell
	MeleeSpells []*core.Spell
	LastShot    *core.Spell

	// The aura that allows you to cast Mongoose Bite
	DefensiveState *core.Aura

	AspectOfTheHawkAura  *core.Aura
	AspectOfTheBeastAura *core.Aura
	RapidFireAura        *core.Aura
	BestialWrathPetAura  *core.Aura
}

func (hunter *Hunter) GetCharacter() *core.Character {
	return &hunter.Character
}

func (hunter *Hunter) GetHunter() *Hunter {
	return hunter
}

// Forever's Trueshot Aura is a party aura the talent teaches, ranged attack power only. The sim only
// has a raid wide switch for it.
func (hunter *Hunter) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	if hunter.Talents.TrueshotAura {
		raidBuffs.TrueshotAura = true
	}
}

func (hunter *Hunter) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (hunter *Hunter) Initialize() {
	if hunter.Env.IsForever() {
		// Forever's Auto Shot keeps firing while the hunter casts Aimed Shot, Multi-Shot or Sniper Shot
		// (the forever-hunter wiki's beta testing, "Auto Shot can fire while the Hunter is casting
		// another spell with a cast time"). So it takes no aiming time of its own that a cast could
		// clip: it goes off every weapon swing, whatever the hunter is doing.
		ranged := hunter.AutoAttacks.RangedConfig()
		ranged.Flags &^= core.SpellFlagCastTimeNoGCD
		ranged.Cast = core.CastConfig{}
		ranged.ExtraCastCondition = nil

		hunter.removeRacialRangedCrit()
	}

	hunter.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagShot) {
			hunter.Shots = append(hunter.Shots, spell)
		}
	})
	hunter.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagStrike) {
			hunter.Strikes = append(hunter.Strikes, spell)
		}
	})
	hunter.OnSpellRegistered(func(spell *core.Spell) {
		if spell.ProcMask.Matches(core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) {
			hunter.MeleeSpells = append(hunter.MeleeSpells, spell)
		}
	})

	hunter.registerAspectOfTheHawkSpell()
	hunter.registerAspectOfTheBeastSpell()

	// Forever pairs Aimed Shot's cooldown with Multi-Shot ("Aimed Shot shares its cooldown with
	// Multi-Shot", client category 2 on both), and Summon Hawk's with Arcane Shot (category 1173).
	multiShotTimer := hunter.NewTimer()
	arcaneShotTimer := hunter.NewTimer()

	hunter.registerSerpentStingSpell()

	hunter.registerArcaneShotSpell(arcaneShotTimer)
	hunter.registerSummonHawkSpell(arcaneShotTimer)
	hunter.registerAimedShotSpell(multiShotTimer)
	hunter.registerMultiShotSpell(multiShotTimer)
	hunter.registerSniperShotSpell()

	hunter.registerRaptorStrikeSpell()
	hunter.registerMongooseBiteSpell()
	hunter.registerWingClipSpell()
	hunter.registerStriderKickSpell()
	hunter.registerLacerateSpell()
	hunter.registerVolleySpell()

	// Forever gives Fire traps and Frost traps a cooldown each (client categories 411 and 412).
	fireTraps := hunter.NewTimer()
	frostTraps := hunter.NewTimer()

	hunter.registerExplosiveTrapSpell(fireTraps)
	hunter.registerImmolationTrapSpell(fireTraps)
	hunter.registerFreezingTrapSpell(frostTraps)

	hunter.registerRapidFire()
}

func (hunter *Hunter) Reset(sim *core.Simulation) {
}

func NewHunter(character *core.Character, options *proto.Player) *Hunter {
	hunterOptions := options.GetHunter()

	hunter := &Hunter{
		Character: *character,
		Talents:   &proto.HunterTalents{},
		Options:   hunterOptions.Options,
	}
	core.FillTalentsProto(hunter.Talents.ProtoReflect(), options.TalentsString, TalentTreeSizes)
	hunter.EnableManaBar()

	hunter.PseudoStats.CanParry = true

	rangedWeapon := hunter.WeaponFromRanged()

	if hunter.HasRangedWeapon() {
		// Ammo
		switch hunter.Options.Ammo {
		case proto.Hunter_Options_RazorArrow:
			hunter.AmmoDPS = 7.5
		case proto.Hunter_Options_SolidShot:
			hunter.AmmoDPS = 7.5
		case proto.Hunter_Options_JaggedArrow:
			hunter.AmmoDPS = 13
		case proto.Hunter_Options_AccurateSlugs:
			hunter.AmmoDPS = 13
		case proto.Hunter_Options_MithrilGyroShot:
			hunter.AmmoDPS = 15
		case proto.Hunter_Options_IceThreadedArrow:
			hunter.AmmoDPS = 16.5
		case proto.Hunter_Options_IceThreadedBullet:
			hunter.AmmoDPS = 16.5
		case proto.Hunter_Options_ThoriumHeadedArrow:
			hunter.AmmoDPS = 17.5
		case proto.Hunter_Options_ThoriumShells:
			hunter.AmmoDPS = 17.5
		case proto.Hunter_Options_RockshardPellets:
			hunter.AmmoDPS = 18
		case proto.Hunter_Options_Doomshot:
			hunter.AmmoDPS = 20
		case proto.Hunter_Options_MiniatureCannonBalls:
			hunter.AmmoDPS = 20.5
		}
		hunter.AmmoDamageBonus = hunter.AmmoDPS * rangedWeapon.SwingSpeed
		hunter.NormalizedAmmoDamageBonus = hunter.AmmoDPS * 2.8

		// Quiver
		switch hunter.Options.QuiverBonus {
		case proto.Hunter_Options_Speed10:
			hunter.PseudoStats.RangedSpeedMultiplier *= 1.1
		case proto.Hunter_Options_Speed11:
			hunter.PseudoStats.RangedSpeedMultiplier *= 1.11
		case proto.Hunter_Options_Speed12:
			hunter.PseudoStats.RangedSpeedMultiplier *= 1.12
		case proto.Hunter_Options_Speed13:
			hunter.PseudoStats.RangedSpeedMultiplier *= 1.13
		case proto.Hunter_Options_Speed14:
			hunter.PseudoStats.RangedSpeedMultiplier *= 1.14
		case proto.Hunter_Options_Speed15:
			hunter.PseudoStats.RangedSpeedMultiplier *= 1.15
		}
	}

	hunter.EnableAutoAttacks(hunter, core.AutoAttackOptions{
		MainHand:        hunter.WeaponFromMainHand(),
		OffHand:         hunter.WeaponFromOffHand(),
		Ranged:          rangedWeapon,
		ReplaceMHSwing:  hunter.TryRaptorStrike,
		AutoSwingRanged: true,
		AutoSwingMelee:  true,
	})

	hunter.AutoAttacks.RangedConfig().Flags |= core.SpellFlagCastTimeNoGCD
	hunter.AutoAttacks.RangedConfig().Cast = core.CastConfig{
		DefaultCast: core.Cast{
			CastTime: time.Millisecond * 500,
		},
		ModifyCast: func(_ *core.Simulation, spell *core.Spell, cast *core.Cast) {
			cast.CastTime = spell.CastTime()
		},
		IgnoreHaste: true, // Hunter GCD is locked at 1.5s
		CastTime: func(spell *core.Spell) time.Duration {
			return time.Duration(float64(spell.DefaultCast.CastTime) / hunter.RangedSwingSpeed())
		},
	}
	hunter.AutoAttacks.RangedConfig().ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return !hunter.IsCasting(sim)
	}
	hunter.AutoAttacks.RangedConfig().CritDamageBonus = hunter.mortalShots()
	hunter.AutoAttacks.RangedConfig().BonusCoefficient = 1
	hunter.AutoAttacks.RangedConfig().ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := hunter.RangedWeaponDamage(sim, spell.RangedAttackPower(target, false)) +
			hunter.AmmoDamageBonus
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	hunter.pet = hunter.NewHunterPet()

	hunter.AddStatDependency(stats.Strength, stats.AttackPower, core.APPerStrength[character.Class])
	hunter.AddStatDependency(stats.Agility, stats.AttackPower, 1)
	hunter.AddStatDependency(stats.Agility, stats.RangedAttackPower, 2)
	hunter.AddStatDependency(stats.Agility, stats.MeleeCrit, core.CritPerAgi(character.Class, character.Level)*core.CritRatingPerCritChance)
	hunter.AddStatDependency(stats.Intellect, stats.SpellCrit, core.CritPerInt(character.Class, character.Level)*core.SpellCritRatingPerCritChance)
	hunter.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	guardians.ConstructGuardians(&hunter.Character)

	return hunter
}

func (hunter *Hunter) OnGCDReady(_ *core.Simulation) {
}

// Forever's weapon specialization racials give melee and spell crit with a sword (Human, 2%), mace
// (Dwarf, 1%) or axe (Orc, 1%), sim/core/racials_forever.go. The sim keeps one crit stat for melee
// and ranged attacks, but the forever-hunter wiki's beta testing finds the racial gives no ranged
// crit, so the ranged attacks take it back off. The pet inherits it, as the wiki also finds.
func (hunter *Hunter) removeRacialRangedCrit() {
	weaponCrit := map[proto.Race]struct {
		weaponType proto.WeaponType
		crit       float64
	}{
		proto.Race_RaceHuman: {proto.WeaponType_WeaponTypeSword, 2},
		proto.Race_RaceDwarf: {proto.WeaponType_WeaponTypeMace, 1},
		proto.Race_RaceOrc:   {proto.WeaponType_WeaponTypeAxe, 1},
	}
	racial, ok := weaponCrit[hunter.Race]
	if !ok {
		return
	}
	for _, weapon := range []*core.Item{hunter.MainHand(), hunter.OffHand()} {
		if weapon.WeaponType == racial.weaponType {
			bonus := racial.crit * core.CritRatingPerCritChance
			hunter.OnSpellRegistered(func(spell *core.Spell) {
				if spell.ProcMask.Matches(core.ProcMaskRanged) {
					spell.BonusCritRating -= bonus
				}
			})
			return
		}
	}
}

// The hunter's ranged spells need the target out of melee range, its melee spells need it in.
func (hunter *Hunter) inRangedRange() bool {
	return hunter.DistanceFromTarget >= core.MinRangedAttackDistance
}

func (hunter *Hunter) inMeleeRange() bool {
	return hunter.DistanceFromTarget <= core.MaxMeleeAttackDistance
}

// Agent is a generic way to access underlying hunter on any of the agents.
type HunterAgent interface {
	GetHunter() *Hunter
}
