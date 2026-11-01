package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Unsealed Wrath (the 31 point Retribution talent, needs Repentance) took Twist of Light's place
// with the launch realms' October 30 hotfix (spell 1310735 in 1.60.1.70707+hf, wago.tools hotfixes
// region 91): "Your melee attacks store a copy of your active Seal for 15 sec, stacking up to 3
// times. Landing an attack with a different Seal consumes all stacks, giving the stored Seal its
// normal chance to activate once per stack."
//
// The copies are the Echo auras Twist of Light had: Echo of Command (1311703), of Righteousness
// (1311704), of Fury (1311701) and of Justice (1311705), each now 15 sec (duration index 8),
// CumulativeAura 3 and no charges or proc flags of their own, with the tooltip "Empowers your next
// melee attack without Seal of Command active with $s3 instances of Seal of Command" ($s3 is 1, so
// one a stack). There is no Echo of the Crusader or of Wisdom, so those two seals store nothing.
// The talent keeps Twist of Light's one proc entry, 100% on done melee auto attacks (proc type 4),
// and its 20% seal cost cut is gone (the same cost effect at 0, see sealCostMultiplier).
//
// We read it literally (report assumption pal_unsealed_wrath):
//   - Only the paladin's landed white swings (extra attacks included) store and spend copies, the
//     proc type the talent has. Holy Strike and the other abilities do neither.
//   - A landed swing first spends the stacks of every Echo whose seal is not the active one, if a
//     seal is up ("with a different Seal"): one roll of that seal's own chance a stack, with the
//     rank of the seal that stored the last copy. Then, if the active seal has an Echo, it stores a
//     copy: a stack more, up to 3, and 15 sec again.
//   - Seal of Righteousness' normal chance is every swing, so its stacks always fire. Seal of
//     Command's is its procs a minute on the main hand's speed; each stack rolls it, and its 1 sec
//     proc cooldown, a cooldown of the seal's own aura, neither stops nor starts the rolls.
//   - A swing that misses, or is dodged or parried, stores and spends nothing, as a seal proc needs
//     a hit. A swing with no seal up spends nothing either.
//
// The sim has Seal of Command and Seal of Righteousness, so those two Echoes are modelled.
const unsealedWrathDuration = time.Second * 15
const unsealedWrathMaxStacks = 3

var EchoOfCommandSpellId int32 = 1311703
var EchoOfRighteousnessSpellId int32 = 1311704

func (paladin *Paladin) registerUnsealedWrath() {
	if !paladin.Talents.UnsealedWrath {
		return
	}
	paladin.echoOfCommand = paladin.registerEchoAura("Echo of Command", EchoOfCommandSpellId)
	paladin.echoOfRighteousness = paladin.registerEchoAura("Echo of Righteousness", EchoOfRighteousnessSpellId)
	paladin.sealEchoes = map[*core.Aura]sealEcho{}
	paladin.echoApply = map[*core.Aura]func(sim *core.Simulation, target *core.Unit){}

	echoes := []*core.Aura{paladin.echoOfCommand, paladin.echoOfRighteousness}
	paladin.RegisterAura(core.Aura{
		Label:    "Unsealed Wrath",
		ActionID: core.ActionID{SpellID: 1310735},
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
				return
			}
			seal := paladin.currentSeal
			if seal == nil || !seal.IsActive() {
				return
			}
			own, hasOwn := paladin.sealEchoes[seal]
			for _, echo := range echoes {
				if !echo.IsActive() || hasOwn && own.aura == echo {
					continue
				}
				stacks := echo.GetStacks()
				apply := paladin.echoApply[echo]
				echo.Deactivate(sim)
				for i := int32(0); i < stacks; i++ {
					apply(sim, result.Target)
				}
			}
			if hasOwn {
				paladin.echoApply[own.aura] = own.apply
				own.aura.Activate(sim)
				own.aura.AddStack(sim)
			}
		},
	})
}

// sealEcho is the Echo a seal stores its copies in, and the seal's effect on a hit.
type sealEcho struct {
	aura  *core.Aura
	apply func(sim *core.Simulation, target *core.Unit)
}

func (paladin *Paladin) registerEchoAura(label string, spellID int32) *core.Aura {
	return paladin.RegisterAura(core.Aura{
		Label:     label,
		ActionID:  core.ActionID{SpellID: spellID},
		Duration:  unsealedWrathDuration,
		MaxStacks: unsealedWrathMaxStacks,
	})
}

// registerSealEcho ties a seal's aura to its Echo. All ranks of a seal share the Echo aura, and
// the Echo applies the rank that stored the last copy.
func (paladin *Paladin) registerSealEcho(sealAura *core.Aura, echoAura *core.Aura, apply func(sim *core.Simulation, target *core.Unit)) {
	if echoAura == nil {
		return
	}
	paladin.sealEchoes[sealAura] = sealEcho{aura: echoAura, apply: apply}
}
