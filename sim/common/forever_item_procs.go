package common

import (
	"time"

	"github.com/wowsims/classic/sim/common/itemhelpers"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// Items whose chance on hit, equip proc or use only the Forever client gives the sim. Classic's
// versions of them stay without effects. Every value is from client 1.60.1.71037 (ItemEffect,
// SpellEffect, SpellCategories, SpellMisc and SpellDuration on wago.tools). The client states no
// rate for these chances on hit (ProcChance 101, no SpellProcsPerMinute row), so a rate is the
// one the item's chance-on-hit entry carries in wowsims/forever's db.json where it has one, and
// 1 PPM, the usual rate of a Classic weapon proc, where it has none.
//
// A spell's damage with variance v is its base points times 1 - v/2 to 1 + v/2.
const (
	DwarvenHandCannon           = 2099
	Hurricane                   = 2824
	TaranIcebreaker             = 2915
	SulthrazeTheLasher          = 9372
	ForceOfMagma                = 11803
	AngerforgesBattleAxe        = 11816
	FieryPlateGauntlets         = 12631
	StormGauntlets              = 12632
	ShellLauncherShotgun        = 13146
	TheBlackrockSlicer          = 13285
	Demonshear                  = 13348
	GargoyleShredderTalons      = 13399
	SoulBreaker                 = 13408
	IcebladeHacker              = 13952
	WarbladeOfCaerDarrow        = 13982
	ShimmeringPlatinumWarhammer = 15418
	Windreaper                  = 15853
	DarkIronRifle               = 16004
	IchorSpitter                = 17002
	Waveslicer                  = 18324
	ForceImbuedGauntlets        = 18383
	Hedgecutter                 = 18498
	XorothianFirestick          = 18755
	SerenityField               = 272439
	LightcrestHammer            = 276419
	FlamebreakTalisman          = 280071
)

// foreverProc is a Forever chance on hit: DmgMin to DmgMin + DmgRange when the spell lands, then
// Tick damage every TickLength for Ticks ticks. Against a target that Vs matches, the direct
// damage is VsMultiplier times as much.
type foreverProc struct {
	PPM         float64
	SpellID     int32
	School      core.SpellSchool
	DefenseType core.DefenseType

	DmgMin, DmgRange float64

	Tick       float64
	Ticks      int32
	TickLength time.Duration

	VsMultiplier float64
	Vs           func(target *core.Unit) bool

	// OnEquip adds the item's other effects.
	OnEquip func(character *core.Character)
}

func newForeverWeaponProc(itemID int32, name string, p foreverProc) {
	itemhelpers.CreateWeaponProcSpellByRuleset(itemID, name, 0, p.PPM, func(character *core.Character) *core.Spell {
		if p.OnEquip != nil {
			p.OnEquip(character)
		}
		config := core.SpellConfig{
			ActionID:         core.ActionID{SpellID: p.SpellID},
			SpellSchool:      p.School,
			DefenseType:      p.DefenseType,
			ProcMask:         core.ProcMaskEmpty,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
		}
		if p.DefenseType == core.DefenseTypeMelee {
			// Like the other phantom strikes, it can't proc equip effects.
			config.ProcMask = core.ProcMaskMeleeSpecial
			config.Flags = core.SpellFlagSuppressEquipProcs
		}
		if p.Ticks > 0 {
			config.Dot = core.DotConfig{
				Aura:          core.Aura{Label: name},
				NumberOfTicks: p.Ticks,
				TickLength:    p.TickLength,
				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
					dot.Snapshot(target, p.Tick, isRollover)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				},
			}
		}
		if p.DmgMin == 0 {
			config.Flags |= core.SpellFlagPureDot
		}

		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if p.DmgMin == 0 {
				result := spell.CalcAndDealOutcome(sim, target, foreverProcOutcome(spell, p.DefenseType, false))
				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
				return
			}
			damage := p.DmgMin
			if p.DmgRange > 0 {
				damage += sim.RandomFloat(name) * p.DmgRange
			}
			if p.Vs != nil && p.Vs(target) {
				damage *= p.VsMultiplier
			}
			result := spell.CalcAndDealDamage(sim, target, damage, foreverProcOutcome(spell, p.DefenseType, true))
			if p.Ticks > 0 && result.Landed() {
				spell.Dot(target).Apply(sim)
			}
		}
		return character.RegisterSpell(config)
	})
}

// foreverProcOutcome is the attack table a proc spell rolls on, by its defense type.
func foreverProcOutcome(spell *core.Spell, defenseType core.DefenseType, canCrit bool) core.OutcomeApplier {
	switch defenseType {
	case core.DefenseTypeMelee:
		return core.Ternary(canCrit, spell.OutcomeMeleeSpecialHitAndCrit, spell.OutcomeMeleeSpecialHit)
	case core.DefenseTypeRanged:
		return core.Ternary(canCrit, spell.OutcomeRangedHitAndCrit, spell.OutcomeRangedHit)
	default:
		return core.Ternary(canCrit, spell.OutcomeMagicHitAndCrit, spell.OutcomeMagicHit)
	}
}

// newForeverOnHitDamage is a Forever equip that deals DmgMin to DmgMin + DmgRange with every
// landed attack in procMask (a weapon's own mask for "attacks with this weapon").
func newForeverOnHitDamage(itemID int32, name string, spellID int32, school core.SpellSchool, defenseType core.DefenseType,
	dmgMin, dmgRange float64, procMask func(character *core.Character) core.ProcMask) {

	core.NewItemEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		if character.Env.IsForever() {
			registerForeverOnHitDamage(character, name, spellID, school, defenseType, dmgMin, dmgRange, procMask(character))
		}
	})
}

func registerForeverOnHitDamage(character *core.Character, name string, spellID int32, school core.SpellSchool, defenseType core.DefenseType,
	dmgMin, dmgRange float64, mask core.ProcMask) {

	if mask == core.ProcMaskUnknown {
		return
	}
	flags := core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagSuppressEquipProcs | core.SpellFlagSuppressWeaponProcs
	// EV-0165: Warblade of Caer Darrow's 1298499 states Suppress Caster Procs, and none of its
	// 3,045 lone crits refreshed the wearer's Vengeance, where 36 to 53% of the other crits did.
	if spellID == 1298499 {
		flags |= core.SpellFlagNoProcs
	}
	spell := character.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: spellID},
		SpellSchool:      school,
		DefenseType:      defenseType,
		ProcMask:         core.Ternary(defenseType == core.DefenseTypeMelee, core.ProcMaskMeleeSpecial, core.ProcMaskEmpty),
		Flags:            flags,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := dmgMin
			if dmgRange > 0 {
				damage += sim.RandomFloat(name) * dmgRange
			}
			outcome := spell.OutcomeMagicCrit
			if defenseType == core.DefenseTypeMelee {
				outcome = spell.OutcomeMeleeSpecialHitAndCrit
			}
			spell.CalcAndDealDamage(sim, target, damage, outcome)
		},
	})
	core.MakeProcTriggerAura(&character.Unit, core.ProcTrigger{
		Name:              name + " Trigger",
		Callback:          core.CallbackOnSpellHitDealt,
		Outcome:           core.OutcomeLanded,
		ProcMask:          mask,
		SpellFlagsExclude: core.SpellFlagSuppressEquipProcs,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			spell.Cast(sim, result.Target)
		},
	})
}

func itemMeleeMask(itemID int32) func(*core.Character) core.ProcMask {
	return func(character *core.Character) core.ProcMask {
		return character.GetProcMaskForItem(itemID) & core.ProcMaskMelee
	}
}

func allMeleeMask(*core.Character) core.ProcMask { return core.ProcMaskMelee }

func isMobType(mobType proto.MobType) func(*core.Unit) bool {
	return func(target *core.Unit) bool { return target.MobType == mobType }
}

func isMobFamily(families ...proto.MobFamily) func(*core.Unit) bool {
	return func(target *core.Unit) bool {
		for _, family := range families {
			if target.MobFamily == family {
				return true
			}
		}
		return false
	}
}

func init() {
	// Weapons, chance on hit

	// Angerforge's Battle Axe: Slaughter (1300783), 109 Physical on the melee table. Trolls and
	// Orcs take twice as much (effect 2) and are horrified, which a boss shrugs off.
	core.RegisterItemConditions(AngerforgesBattleAxe, core.ConditionVsTrolls, core.ConditionVsOrcsAndOgres)
	newForeverWeaponProc(AngerforgesBattleAxe, "Angerforge's Battle Axe", foreverProc{PPM: 1, SpellID: 1300783,
		School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, DmgMin: 109,
		VsMultiplier: 2, Vs: isMobFamily(proto.MobFamily_MobFamilyTroll, proto.MobFamily_MobFamilyOrc)})

	// Demonshear: Shadow Bolt (17483), 150 Shadow, then 34 every 2 sec for 6 sec. 1.1 PPM.
	newForeverWeaponProc(Demonshear, "Demonshear", foreverProc{PPM: 1.1, SpellID: 17483,
		School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, DmgMin: 150,
		Tick: 34, Ticks: 3, TickLength: time.Second * 2})

	// Force of Magma: Firebolt (18086), 182 Fire. 2 PPM.
	newForeverWeaponProc(ForceOfMagma, "Force of Magma", foreverProc{PPM: 2, SpellID: 18086,
		School: core.SpellSchoolFire, DefenseType: core.DefenseTypeMagic, DmgMin: 182})

	// Gargoyle Shredder Talons: Rend (1320824), a bleed of 20 every 2 sec for 14 sec.
	newForeverWeaponProc(GargoyleShredderTalons, "Gargoyle Shredder Talons", foreverProc{PPM: 1, SpellID: 1320824,
		School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee,
		Tick: 20, Ticks: 7, TickLength: time.Second * 2})

	// Hedgecutter (1302344): 42 Arcane, variance 0.42 (33.2 to 50.8). Plants take three times as
	// much (effect 2).
	core.RegisterItemConditions(Hedgecutter, core.ConditionVsPlants)
	newForeverWeaponProc(Hedgecutter, "Hedgecutter", foreverProc{PPM: 1, SpellID: 1302344,
		School: core.SpellSchoolArcane, DefenseType: core.DefenseTypeMagic, DmgMin: 42 * 0.79, DmgRange: 42 * 0.42,
		VsMultiplier: 3, Vs: isMobFamily(proto.MobFamily_MobFamilyPlant)})

	// Iceblade Hacker: its Frostbolt (1298416), 105 Frost, variance 0.6 (73.5 to 136.5), and a 50%
	// slow. Its equip (1298414) deals 40.7 Frost with each melee attack of the axe, on the melee
	// table.
	newForeverWeaponProc(IcebladeHacker, "Iceblade Hacker", foreverProc{PPM: 1, SpellID: 1298416,
		School: core.SpellSchoolFrost, DefenseType: core.DefenseTypeMagic, DmgMin: 73.5, DmgRange: 63,
		OnEquip: func(character *core.Character) {
			registerForeverOnHitDamage(character, "Iceblade Hacker Frost", 1298414, core.SpellSchoolFrost, core.DefenseTypeMelee,
				40.7, 0, itemMeleeMask(IcebladeHacker)(character))
		}})

	// Ichor Spitter: Poison (17511), 8 Nature every 2 sec for 20 sec.
	newForeverWeaponProc(IchorSpitter, "Ichor Spitter", foreverProc{PPM: 1, SpellID: 17511,
		School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic,
		Tick: 8, Ticks: 10, TickLength: time.Second * 2})

	// Lightcrest Hammer: Smite (1299631), 105 Holy, twice as much to Undead (effect 2). 3.6 PPM.
	core.RegisterItemConditions(LightcrestHammer, core.MobTypeCondition(proto.MobType_MobTypeUndead))
	newForeverWeaponProc(LightcrestHammer, "Lightcrest Hammer", foreverProc{PPM: 3.6, SpellID: 1299631,
		School: core.SpellSchoolHoly, DefenseType: core.DefenseTypeMagic, DmgMin: 105,
		VsMultiplier: 2, Vs: isMobType(proto.MobType_MobTypeUndead)})

	// Shimmering Platinum Warhammer: Lightning Bolt (19874), 252 Nature, variance 0.3256 (211 to
	// 293).
	newForeverWeaponProc(ShimmeringPlatinumWarhammer, "Shimmering Platinum Warhammer", foreverProc{PPM: 1, SpellID: 19874,
		School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, DmgMin: 252 * (1 - 0.32558140159/2), DmgRange: 252 * 0.32558140159})

	// Soul Breaker (17506): 10 Shadow every sec for 10 sec, and as much mana.
	newForeverWeaponProc(SoulBreaker, "Soul Breaker", foreverProc{PPM: 1, SpellID: 17506,
		School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic,
		Tick: 10, Ticks: 10, TickLength: time.Second})

	// Sul'thraze the Lasher: Rage of Sul'thraze (1294430), 159 Shadow, variance 0.8 (95.4 to
	// 222.6), and 180 less attack power on the target for 15 sec. Shadow Lash (1294431) follows 5
	// sec later with 159 more. Trolls take twice the second part (effect 5), which we leave out.
	newForeverWeaponProc(SulthrazeTheLasher, "Sul'thraze the Lasher", foreverProc{PPM: 1, SpellID: 1294430,
		School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeMagic, DmgMin: 95.4, DmgRange: 127.2,
		Tick: 159, Ticks: 1, TickLength: time.Second * 5})

	// Taran Icebreaker: Fireball (16415), 300 Fire, variance 0.2 (270 to 330), then 25 every 2 sec
	// for 8 sec.
	newForeverWeaponProc(TaranIcebreaker, "Taran Icebreaker", foreverProc{PPM: 1, SpellID: 16415,
		School: core.SpellSchoolFire, DefenseType: core.DefenseTypeMagic, DmgMin: 270, DmgRange: 60,
		Tick: 25, Ticks: 4, TickLength: time.Second * 2})

	// The Blackrock Slicer: Wound (17407), 100 Physical, variance 1 (50 to 150), on the melee
	// table, then 13 every sec for 10 sec.
	newForeverWeaponProc(TheBlackrockSlicer, "The Blackrock Slicer", foreverProc{PPM: 1, SpellID: 17407,
		School: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMelee, DmgMin: 50, DmgRange: 100,
		Tick: 13, Ticks: 10, TickLength: time.Second})

	// Waveslicer (1302166): 70 Nature. Aquatic enemies take three times as much, and no boss we
	// sim is Aquatic.
	newForeverWeaponProc(Waveslicer, "Waveslicer", foreverProc{PPM: 1, SpellID: 1302166,
		School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic, DmgMin: 70})

	// Windreaper (20586): 21 Nature every 2 sec for 10 sec.
	newForeverWeaponProc(Windreaper, "Windreaper", foreverProc{PPM: 1, SpellID: 20586,
		School: core.SpellSchoolNature, DefenseType: core.DefenseTypeMagic,
		Tick: 21, Ticks: 5, TickLength: time.Second * 2})

	// Ranged weapons, chance on hit. Each spell rolls on the ranged table.

	// Dark Iron Rifle: Shadow Shot (29641), 26 Shadow, variance 0.364 (21.3 to 30.7).
	newForeverWeaponProc(DarkIronRifle, "Dark Iron Rifle", foreverProc{PPM: 1, SpellID: 29641,
		School: core.SpellSchoolShadow, DefenseType: core.DefenseTypeRanged, DmgMin: 26 * (1 - 0.36363637447/2), DmgRange: 26 * 0.36363637447})

	// Dwarven Hand Cannon: Flaming Cannonball (29639), 49 Fire, variance 0.39 (39.4 to 58.6).
	newForeverWeaponProc(DwarvenHandCannon, "Dwarven Hand Cannon", foreverProc{PPM: 1, SpellID: 29639,
		School: core.SpellSchoolFire, DefenseType: core.DefenseTypeRanged, DmgMin: 49 * (1 - 0.3902438879/2), DmgRange: 49 * 0.3902438879})

	// Shell Launcher Shotgun: Flaming Shell (29647), 38.5 Fire, variance 0.364 (31.5 to 45.5).
	newForeverWeaponProc(ShellLauncherShotgun, "Shell Launcher Shotgun", foreverProc{PPM: 1, SpellID: 29647,
		School: core.SpellSchoolFire, DefenseType: core.DefenseTypeRanged, DmgMin: 31.5, DmgRange: 14})

	// Xorothian Firestick: Xorothian Fire (1302343), 14 Fire every sec for 3 sec.
	newForeverWeaponProc(XorothianFirestick, "Xorothian Firestick", foreverProc{PPM: 1, SpellID: 1302343,
		School: core.SpellSchoolFire, DefenseType: core.DefenseTypeRanged,
		Tick: 14, Ticks: 3, TickLength: time.Second})

	// Equips that deal damage with every landed attack (ProcChance 100)

	// Fiery Plate Gauntlets: Fire Strike (7714), 4 Fire with each melee attack.
	newForeverOnHitDamage(FieryPlateGauntlets, "Fiery Plate Gauntlets", 7714, core.SpellSchoolFire, core.DefenseTypeMagic, 4, 0, allMeleeMask)

	// Force Imbued Gauntlets: Force Imbued (1302247), 2 Arcane with each melee attack.
	newForeverOnHitDamage(ForceImbuedGauntlets, "Force Imbued Gauntlets", 1302247, core.SpellSchoolArcane, core.DefenseTypeMagic, 2, 0, allMeleeMask)

	// Hurricane: Frost Arrow (29502), 8 Frost, variance 0.5 (6 to 10), with each ranged auto shot.
	newForeverOnHitDamage(Hurricane, "Hurricane", 29502, core.SpellSchoolFrost, core.DefenseTypeMagic, 6, 4, func(character *core.Character) core.ProcMask {
		return character.GetProcMaskForItem(Hurricane) & core.ProcMaskRangedAuto
	})

	// Storm Gauntlets: Lightning Strike (16614), 3 Nature with each melee attack.
	newForeverOnHitDamage(StormGauntlets, "Storm Gauntlets", 16614, core.SpellSchoolNature, core.DefenseTypeMagic, 3, 0, allMeleeMask)

	// Warblade of Caer Darrow (1298499): 27.72 Frost, variance 0.4 (22.2 to 33.3), with each melee
	// attack of the sword, on the melee table.
	newForeverOnHitDamage(WarbladeOfCaerDarrow, "Warblade of Caer Darrow", 1298499, core.SpellSchoolFrost, core.DefenseTypeMelee,
		27.72*0.8, 27.72*0.4, itemMeleeMask(WarbladeOfCaerDarrow))

	// Trinkets

	// Flamebreak Talisman: Use: 33 Agility for 30 sec, and 33 more with an elemental nearby
	// (1315878). 5 min cooldown, and 30 sec on the burst trinket category.
	core.RegisterItemConditions(FlamebreakTalisman, core.MobTypeCondition(proto.MobType_MobTypeElemental))
	core.NewItemEffect(FlamebreakTalisman, func(agent core.Agent) {
		character := agent.GetCharacter()
		if !character.Env.IsForever() {
			return
		}
		agility := 33.0
		for _, target := range character.Env.Encounter.TargetUnits {
			if target.MobType == proto.MobType_MobTypeElemental {
				agility = 66
			}
		}
		registerForeverStatTrinket(character, FlamebreakTalisman, stats.Stats{stats.Agility: agility},
			time.Second*30, time.Minute*5, time.Second*30)
	})

	// Serenity Field: Use: 72 Spirit for 15 sec (1291103). 6 min cooldown, and 15 sec on the burst
	// trinket category.
	core.NewItemEffect(SerenityField, func(agent core.Agent) {
		character := agent.GetCharacter()
		if !character.Env.IsForever() {
			return
		}
		registerForeverStatTrinket(character, SerenityField, stats.Stats{stats.Spirit: 72},
			time.Second*15, time.Minute*6, time.Second*15)
	})
}
