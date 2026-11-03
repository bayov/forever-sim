package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

var HolyShieldValues = []struct {
	level    int32
	spellID  int32
	procID   int32
	manaCost float64
	damage   float64
}{
	// Forever: the block bonus is 30% (the 69876 beta client had 20%, build 70245 is back at
	// Classic's 30%). The launch realms' October 30 hotfix (1.60.1.70707+hf) moved every rank down:
	// a new rank 1 at level 30 for 100 mana and 75 a block (20925), and the old ranks 1 and 2 at 40
	// and 50 (20927: 150 mana, 110; 20928: 195 mana, 153). Before it the ranks were at 40 / 50 / 60
	// for 150 / 195 / 240 mana and 110 / 153 / 221 a block (Classic 65 / 95 / 130). The talent no
	// longer needs Templar's Bulwark (tree data). The hotfix also added a rank 4 at 60
	// (1239570: 240 mana, 221 a block, 8% of spell power, 30% block, 4 charges, 10 sec on the
	// same 10 sec category cooldown; client 70753 has the row). Its block damage has no spell of
	// its own (effect 1 is aura 43, the aura deals it), so the sim tags the aura's ID for it.
	{level: 30, spellID: 20925, procID: 20955, manaCost: 100, damage: 75},
	{level: 40, spellID: 20927, procID: 20956, manaCost: 150, damage: 110},
	{level: 50, spellID: 20928, procID: 20957, manaCost: 195, damage: 153},
	{level: 60, spellID: 1239570, procID: 0, manaCost: 240, damage: 221},
}

var HolyShieldLevel = [...]int{0, 30, 40, 50, 60}
var HolyShieldSpellId = [...]int32{0, 20925, 20927, 20928, 1239570}

func (paladin *Paladin) registerHolyShield() {
	if !paladin.Talents.HolyShield {
		return
	}

	numCharges := int32(4)
	blockBonus := 30.0 * core.BlockRatingPerBlockChance
	// Steadfast Libram's 30% block value while Holy Shield is up, added to the other block value
	// percents as Shield Specialization's is (both aura 272).
	steadfast := core.TernaryFloat64(paladin.Ranged().ID == SteadfastLibram, 0.3, 0)
	// The ranks share one category cooldown (931).
	cdTimer := paladin.NewTimer()

	for i, values := range HolyShieldValues {
		rank := i + 1
		level := values.level
		spellID := values.spellID
		procID := values.procID
		manaCost := values.manaCost
		damage := values.damage

		if paladin.Level < level {
			break
		}

		procActionID := core.ActionID{SpellID: procID}
		if procID == 0 {
			procActionID = core.ActionID{SpellID: spellID, Tag: 1}
		}
		paladin.holyShieldProc[i] = paladin.RegisterSpell(core.SpellConfig{
			ActionID:    procActionID,
			SpellCode:   SpellCode_PaladinHolyShieldProc,
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,

			RequiredLevel: int(level),
			Rank:          rank,

			DamageMultiplier: 1,
			ThreatMultiplier: 1.2,
			// Forever's rows give the damage 8% of spell power (client 20925, 20927 and 20928,
			// effect 1; Classic's 5%).
			BonusCoefficient: core.TernaryFloat64(paladin.Env.IsForever(), 0.08, 0.05),

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				// Spell damage from Holy Shield can crit, but does not miss.
				spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicCrit)
			},
		})

		paladin.holyShieldAura[i] = paladin.RegisterAura(core.Aura{
			Label:     "Holy Shield" + paladin.Label + strconv.Itoa(rank),
			ActionID:  core.ActionID{SpellID: spellID},
			Duration:  time.Second * 10,
			MaxStacks: numCharges,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				aura.SetStacks(sim, numCharges)
				paladin.AddStatDynamic(sim, stats.Block, blockBonus)
				paladin.PseudoStats.BlockValueMultiplier += steadfast
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				paladin.AddStatDynamic(sim, stats.Block, -blockBonus)
				paladin.PseudoStats.BlockValueMultiplier -= steadfast
			},
			OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.DidBlock() {
					paladin.holyShieldProc[i].Cast(sim, spell.Unit)
					aura.RemoveStack(sim)
				}
			},
		})

		paladin.RegisterSpell(core.SpellConfig{
			ActionID:      core.ActionID{SpellID: spellID},
			SpellCode:     SpellCode_PaladinHolyShield,
			Flags:         core.SpellFlagAPL,
			RequiredLevel: int(level),
			Rank:          rank,
			ManaCost: core.ManaCostOptions{
				FlatCost:   manaCost,
				Multiplier: paladin.benediction(),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    cdTimer,
					Duration: time.Second * 10,
				},
			},
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				// One Holy Shield at a time: a rank replaces another's.
				for j, other := range paladin.holyShieldAura {
					if j != i && other != nil {
						other.Deactivate(sim)
					}
				}
				paladin.holyShieldAura[i].Activate(sim)
			},
		})
	}
}
