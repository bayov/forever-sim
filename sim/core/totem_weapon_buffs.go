package core

import (
	"fmt"
	"slices"

	"github.com/wowsims/classic/sim/core/proto"
)

// Under Forever a main hand holds three weapon buffs at once:
//   - the buff of a Windfury or Flametongue Totem (only one of the two at a time),
//   - the shaman's own imbue (Windfury, Rockbiter, Flametongue or Frostbrand Weapon),
//   - an oil or a stone.
//
// They work together, except that the shaman's imbue turns off the totem buff of the same
// kind. Windfury Weapon's tooltip says "When applied to main hand, disables any benefit
// you personally receive from Windfury Totem", and Flametongue Weapon says the same about
// Flametongue Totem. So Windfury Weapon works next to Flametongue Totem, and Flametongue
// or Rockbiter Weapon next to Windfury Totem, but Windfury Weapon with Windfury Totem only
// gives the weapon's procs.
//
// This file has the totem buffs. The raid buffs give them for another shaman's totem, and
// the shaman package uses the same auras for the shaman's own totems.

const FlametongueTotemRanks = 4

var FlametongueTotemSpellId = [FlametongueTotemRanks + 1]int32{0, 8227, 8249, 10526, 16387}
var FlametongueTotemProcSpellId = [FlametongueTotemRanks + 1]int32{0, 8253, 8248, 10523, 16389}

// Fire damage per 4 sec of weapon speed, N / 25 on wowhead's Forever tooltips (548, 781,
// 1061 and 1363). They do not grow with level.
var FlametongueTotemMaxDamage = [FlametongueTotemRanks + 1]float64{0, 21.92, 31.24, 42.44, 54.52}
var FlametongueTotemLevel = [FlametongueTotemRanks + 1]int{0, 28, 38, 48, 58}

var WindfuryTotemLevel = [WindfuryRanks + 1]int{0, 32, 42, 52}

// The temporary enchants of every Windfury Weapon and Flametongue Weapon rank, the same
// as WindfuryWeaponEnchantId and FlametongueWeaponEnchantId in the shaman package.
var windfuryWeaponEnchantIds = []int32{283, 284, 525, 1669}
var flametongueWeaponEnchantIds = []int32{5, 4, 3, 523, 1665, 1666}

// mainHandHasImbue tells whether the main hand carries one of these imbues right now. We
// look at it on every hit instead of once, because the shaman's imbue goes on the weapon
// after the raid buffs are applied, and an item swap can change it.
func mainHandHasImbue(character *Character, enchantIds []int32) bool {
	mh := character.MainHand()
	return mh != nil && slices.Contains(enchantIds, mh.TempEnchant)
}

// FlametongueTotemBuffAura gives our landed white swings extra Fire damage, by weapon speed.
//
// Under Forever only white swings proc it, extra swings included. The Forever client
// (70291) makes the totem's buff a proc aura that triggers on melee swings alone
// (ProcTypeMask 0x4), where Windfury Totem and the Forever imbues also take melee abilities
// (0x14). On the beta Stormstrike never procced it (the user, 2026-10-10), and the user
// remembers one proc for each Windfury Weapon proc, from the swing and not the two attacks.
// In Era the totem put an enchant on the main hand, which procced on every main hand hit.
//
// Nothing in the Forever client ties the aura to a weapon, so off-hand swings proc it too,
// as the user expects (2026-10-10). We assume an off-hand proc goes by the off hand's speed.
//
// Spell power adds nothing to it, unlike Flametongue Weapon. On the beta a level 30 shaman
// with 55 spell power hit 20 with rank 1 on a 3.6 speed axe (2026-10-08), the 19.73 of the
// base damage, where 10% of spell power would have made it 25. cmangos doesn't give it
// spell power either.
//
// The proc crits at our spell crit, for 1.5 times a normal hit. On the beta at level 30 it crit
// on 46 of 400 procs with 11.81% spell crit on the sheet, and on 31 of 417 with about 7.3% after
// an Elemental respec (2026-10-10 and 11). A flat 8% would have given about 32 of the 400.
//
// The aura is not active on its own. The caller makes it permanent or turns it on and off
// with the totem. The label and the tag keep another shaman's totem apart from the
// shaman's own.
func FlametongueTotemBuffAura(character *Character, rank int, label string, tag int32) *Aura {
	damagePerSecond := FlametongueTotemMaxDamage[rank] / 4
	var weaponSpeed float64

	procSpell := character.RegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: FlametongueTotemProcSpellId[rank], Tag: tag},
		SpellSchool: SpellSchoolFire,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskSpellDamageProc,
		Flags:       SpellFlagNoOnCastComplete | SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			damage := damagePerSecond * weaponSpeed
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
		},
	})

	procMask := ProcMaskMeleeMH
	if character.Env.IsForever() {
		procMask = ProcMaskMeleeWhiteHit
	}

	return character.RegisterAura(Aura{
		Label:    fmt.Sprintf("%s (Rank %d)", label, rank),
		ActionID: ActionID{SpellID: FlametongueTotemSpellId[rank]},
		Duration: NeverExpires,
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(procMask) || mainHandHasImbue(character, flametongueWeaponEnchantIds) {
				return
			}
			weaponSpeed = character.MainHand().SwingSpeed
			if spell.ProcMask.Matches(ProcMaskMeleeOH) {
				weaponSpeed = character.OffHand().SwingSpeed
			}
			procSpell.Cast(sim, result.Target)
		},
	})
}

// applyRaidTotemWeaponBuff puts another shaman's totem buff on the character's main hand,
// at the highest rank of that totem the character's level allows. Only the first one
// applied counts, because the weapon has one slot for it.
func applyRaidTotemWeaponBuff(character *Character, buff proto.TotemWeaponBuff) {
	if character.RaidTotemWeaponBuff != nil || !character.HasMHWeapon() {
		return
	}

	switch buff {
	case proto.TotemWeaponBuff_TotemWeaponBuffWindfury:
		if rank := HighestRankAt(character.Level, WindfuryTotemLevel[:]); rank > 0 {
			character.RaidTotemWeaponBuff = MakePermanent(WindfuryTotemBuffAura(character, int32(rank), "Windfury Totem Raid"))
		}
	case proto.TotemWeaponBuff_TotemWeaponBuffFlametongue:
		if rank := HighestRankAt(character.Level, FlametongueTotemLevel[:]); rank > 0 {
			character.RaidTotemWeaponBuff = MakePermanent(FlametongueTotemBuffAura(character, rank, "Flametongue Totem Raid", 1))
		}
	}
}
