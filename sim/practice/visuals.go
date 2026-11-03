//go:build practice

package practice

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// What the page needs to show a hit or a cast the way the client does: which totem it came from,
// and the client's spell id when the sim reports another one.
//
// The sim reports the weapon imbues' procs under the imbue's own spell (Windfury Weapon 16362,
// say), which the client shows as putting the imbue on. The client's proc spells have their own
// visuals: Windfury's extra swings are 439440 on Forever, Flametongue's hits Flametongue Attack
// and Frostbrand's Frostbrand Attack (by rank, from SpellItemEnchantment's EffectArg). Fire Nova's
// casts (408341 to 408345) have no visual at all, and the client shows its damage spells'
// (408423 and up, visual 131764).

// flametongueAttack is the client's Flametongue proc. The server picks it (Forever's enchants
// all point at the 436519 aura), and the beta logs show 10444, 16368, 29469 and 29470 across
// ranks and builds. 29469 is the only one in the release client's logs, and 10444, 29469 and
// 29470 share visual 140476.
const flametongueAttack = 29469

// windfuryAttack is Windfury Weapon's extra attack on Forever (439440 in every log).
const windfuryAttack = 439440

// frostbrandAttack is Frostbrand Attack by Frostbrand Weapon's rank.
var frostbrandAttack = [shaman.FrostbrandWeaponRanks + 1]int32{0, 8034, 8037, 10458, 16352, 16353}

// fireNovaDamage is Fire Nova's damage spell by rank, which has the nova's visual.
var fireNovaDamage = [shaman.FireNovaRanks + 1]int32{0, 408423, 408424, 408426, 408427, 408428}

// hitVisual says which totem a hit or cast came from ("fire", or "" for one of yours), the
// client's spell for its visual when it differs from the sim's (0 when it doesn't), and for a
// Searing Totem bolt when the next one hits (0 for none).
func (r *run) hitVisual(sim *core.Simulation, spell *core.Spell) (totem string, visual int32, nextAt float64) {
	if r.shaman == nil {
		return "", 0, 0
	}
	id := spell.ActionID.SpellID
	tag := spell.ActionID.Tag
	for rank := 1; rank <= shaman.SearingTotemRanks; rank++ {
		if id == shaman.SearingTotemAttackSpellId[rank] {
			// The totem casts its next bolt right away, and it lands at the dot's next tick.
			if totemSpell := r.shaman.SearingTotem[rank]; totemSpell != nil {
				if dot := totemSpell.AOEDot(); dot != nil && dot.IsActive() && dot.MaxTicksRemaining() > 0 {
					nextAt = toMs(dot.NextTickAt())
				}
			}
			return "fire", 0, nextAt
		}
	}
	for rank := 1; rank <= shaman.MagmaTotemRanks; rank++ {
		if id == shaman.MagmaTotemAoeSpellId[rank] {
			return "fire", 0, 0
		}
	}
	for rank := 1; rank <= shaman.FireNovaRanks; rank++ {
		if id == shaman.FireNovaSpellId[rank] {
			return "fire", fireNovaDamage[rank], 0
		}
	}
	if tag == shaman.ImbueCastTag {
		return "", 0, 0
	}
	for rank := 1; rank <= shaman.WindfuryWeaponRanks; rank++ {
		if id == shaman.WindfuryWeaponSpellId[rank] {
			return "", windfuryAttack, 0
		}
	}
	for rank := 1; rank <= shaman.FlametongueWeaponRanks; rank++ {
		if id == shaman.FlametongueWeaponSpellId[rank] {
			return "", flametongueAttack, 0
		}
	}
	for rank := 1; rank <= shaman.FrostbrandWeaponRanks; rank++ {
		if id == shaman.FrostbrandWeaponSpellId[rank] {
			return "", frostbrandAttack[rank], 0
		}
	}
	return "", 0, 0
}

// offHandImbue is the shaman imbue on the off hand, by its enchant ("WindfuryWeapon"), or "".
func offHandImbue(char *core.Character) string {
	oh := char.OffHand()
	if oh == nil || oh.ID == 0 || oh.TempEnchant == 0 {
		return ""
	}
	byEnchant := []struct {
		imbue proto.WeaponImbue
		ids   []int32
	}{
		{proto.WeaponImbue_WindfuryWeapon, shaman.WindfuryWeaponEnchantId[1:]},
		{proto.WeaponImbue_FlametongueWeapon, shaman.FlametongueWeaponEnchantId[1:]},
		{proto.WeaponImbue_FrostbrandWeapon, shaman.FrostbrandWeaponEnchantId[1:]},
		{proto.WeaponImbue_RockbiterWeapon, shaman.RockbiterWeaponEnchantId[1:]},
	}
	for _, b := range byEnchant {
		for _, id := range b.ids {
			if id == oh.TempEnchant {
				return b.imbue.String()
			}
		}
	}
	return ""
}
