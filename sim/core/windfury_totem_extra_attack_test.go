package core

import (
	"testing"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

func init() {
	RegisterAgentFactory(
		proto.Player_Warrior{},
		proto.Spec_SpecWarrior,
		NewFakeWindfuryUser,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Warrior)
			if !ok {
				panic("Invalid spec value for Warrior!")
			}
			player.Spec = playerSpec
		},
	)
}

// FakeWindfuryUser carries Windfury Totem (rank 3) and records each main hand white hit's extra
// attack source as it lands.
type FakeWindfuryUser struct {
	Character

	Windfury *Aura
	hits     []windfuryProbeHit
}

type windfuryProbeHit struct {
	at       time.Duration
	source   ActionID
	totem    bool // IsWindfuryTotemAttack when the hit landed
	extraHit bool // IsExtraAttack when the hit landed
}

func (fw *FakeWindfuryUser) GetCharacter() *Character { return &fw.Character }
func (fw *FakeWindfuryUser) ApplyTalents()            {}
func (fw *FakeWindfuryUser) Reset(_ *Simulation)      { fw.hits = nil }
func (fw *FakeWindfuryUser) OnGCDReady(_ *Simulation) {}

func (fw *FakeWindfuryUser) Initialize() {
	fw.Windfury = WindfuryTotemBuffAura(&fw.Character, 3, "Windfury Totem Raid", ActionID{SpellID: WindfuryTotemSpellId[3]})
	fw.RegisterAura(Aura{
		Label:    "Windfury probe",
		Duration: NeverExpires,
		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if spell.ProcMask != ProcMaskMeleeMHAuto { // the white hit lands as a copy of the swing spell
				return
			}
			aa := &fw.AutoAttacks
			fw.hits = append(fw.hits, windfuryProbeHit{sim.CurrentTime, aa.ExtraAttackSource(), aa.IsWindfuryTotemAttack(), aa.IsExtraAttack()})
		},
	})
}

func NewFakeWindfuryUser(char *Character, _ *proto.Player) Agent {
	fw := &FakeWindfuryUser{Character: *char}
	fw.EnableAutoAttacks(fw, AutoAttackOptions{
		MainHand:       Weapon{SwingSpeed: 3.6, BaseDamageMin: 100, BaseDamageMax: 100},
		AutoSwingMelee: true,
	})
	return fw
}

func setupWindfurySim(t *testing.T) (*Simulation, *FakeWindfuryUser) {
	t.Helper()
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 7, Ruleset: proto.Ruleset_RulesetForever},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:      "Warrior",
					Class:     proto.Class_ClassWarrior,
					Level:     60,
					Consumes:  &proto.Consumes{},
					Buffs:     &proto.IndividualBuffs{},
					Spec:      &proto.Player_Warrior{},
					Equipment: &proto.EquipmentSpec{},
					Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
				}},
				Buffs: &proto.PartyBuffs{},
			}},
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60, MobType: proto.MobType_MobTypeHumanoid}},
			Duration: 600,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	fw := sim.Raid.Parties[0].Players[0].(*FakeWindfuryUser)
	stepSwingSim(t, sim, time.Second)
	return sim, fw
}

// rollWindfury hands the totem landed hits of spell, a fifth of a second apart (past its 100 ms
// cooldown), until it procs; it returns false if 200 hits don't proc it. Each hit lands as an
// extra attack from source, unless source is zero.
func rollWindfury(t *testing.T, sim *Simulation, fw *FakeWindfuryUser, spell *Spell, source ActionID) bool {
	t.Helper()
	aa := &fw.AutoAttacks
	for range 200 {
		stepSwingSim(t, sim, sim.CurrentTime+200*time.Millisecond)
		pending, stored := len(aa.mh.extraSourcesPending), len(aa.mh.extraSourcesStored)
		result := spell.NewResult(fw.CurrentTarget)
		result.Outcome = OutcomeHit
		outer := aa.mh.extraSource
		aa.mh.extraSource = source
		fw.Windfury.OnSpellHitDealt(fw.Windfury, sim, spell, result)
		aa.mh.extraSource = outer
		if len(aa.mh.extraSourcesPending) > pending || len(aa.mh.extraSourcesStored) > stored {
			return true
		}
	}
	return false
}

// Every Windfury Totem extra swing reads as one when it lands, whichever hit launched it: a white
// swing, a main hand ability that rolls the totem (standards 17), or one cast with a stop attack
// macro, whose extra swing is stored for the next swing.
func TestWindfuryTotemAttackRecognized(t *testing.T) {
	for _, c := range []struct {
		name  string
		spell func(fw *FakeWindfuryUser) *Spell
	}{
		{"white swing", func(fw *FakeWindfuryUser) *Spell { return fw.AutoAttacks.MHAuto() }},
		{"Heroic Strike", func(fw *FakeWindfuryUser) *Spell {
			return &Spell{ActionID: ActionID{SpellID: 11567}, ProcMask: ProcMaskMeleeMHSpecial, Unit: &fw.Unit}
		}},
		{"Heroic Strike with a stop attack macro", func(fw *FakeWindfuryUser) *Spell {
			return &Spell{ActionID: ActionID{SpellID: 11567}, ProcMask: ProcMaskMeleeMHSpecial, Flags: SpellFlagBatchStopAttackMacro, Unit: &fw.Unit}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, fw := setupWindfurySim(t)
			for proc := 0; proc < 5; proc++ {
				if !rollWindfury(t, sim, fw, c.spell(fw), ActionID{}) {
					t.Fatalf("no Windfury Totem proc in 200 landed hits")
				}
				procAt := sim.CurrentTime
				fw.hits = fw.hits[:0]
				// A stored swing waits for the next swing; the others swing at once.
				stepSwingSim(t, sim, max(procAt, fw.AutoAttacks.MainhandSwingAt())+50*time.Millisecond)
				found := false
				for _, h := range fw.hits {
					if h.source == (ActionID{SpellID: WindfuryBuffSpellId[3]}) {
						found = true
						if !h.totem || !h.extraHit {
							t.Errorf("a Windfury Totem swing at %v reads totem=%v extra=%v, want both", h.at, h.totem, h.extraHit)
						}
					} else if h.totem {
						t.Errorf("a swing from %v at %v reads as a Windfury Totem swing", h.source, h.at)
					}
				}
				if !found {
					t.Fatalf("the Windfury Totem swing from the proc at %v never landed with its source", procAt)
				}
			}
		})
	}
}

// No extra attack rolls Windfury Totem under Forever, whatever proc gave it (the log audit of
// standards 17): Windfury's own swing, Hand of Justice, Sword Specialization, Thrash and a stored
// attack. A normal swing still does.
func TestWindfuryTotemExtraAttacksDontRoll(t *testing.T) {
	sim, fw := setupWindfurySim(t)
	mh := fw.AutoAttacks.MHAuto()
	for _, source := range []ActionID{
		{SpellID: WindfuryBuffSpellId[3]},
		{SpellID: 15601}, // Hand of Justice
		{SpellID: 13964}, // Sword Specialization
		{SpellID: 12815}, // Sword Specialization (warrior)
		{SpellID: 21919}, // Thrash, and the stored attacks' aura
	} {
		if rollWindfury(t, sim, fw, mh, source) {
			t.Errorf("an extra attack from %v rolled Windfury Totem", source)
		}
	}
	if !rollWindfury(t, sim, fw, mh, ActionID{}) {
		t.Errorf("a normal swing never rolled Windfury Totem in 200 hits")
	}
}
