package aplkit

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// WarriorLeveling is a warrior's leveling rotation, for any of the weapon styles. It plays the
// Protection Warrior spec, which is the sim's warrior with Shield Block and Shield Wall, so one
// spec has every ability with a shield and with a two-hander.
//
// This is only a start, so the plan tools run for a warrior: rest, walk up or Charge, Battle
// Shout and Heroic Strike.
type WarriorLeveling struct {
	// The level we build the rotation for, as in ShamanEnhLeveling.
	Level int32
}

// The levels of each ability's ranks, and the ID of its top rank, which the sim casts as the
// highest rank we know. The levels have a 0 for "no rank", as core.HighestRankAt wants.
type warriorAbility struct {
	name   string
	top    int32
	levels []int32
}

var (
	warriorBattleShout  = warriorAbility{"battleShout", core.BattleShoutSpellId[6], []int32{0, 1, 12, 22, 32, 42, 52}}
	warriorHeroicStrike = warriorAbility{"heroicStrike", 11567, []int32{0, 1, 8, 16, 24, 32, 40, 48, 56}}
	warriorCharge       = warriorAbility{"charge", 11578, []int32{0, 4, 26, 46}}
)

var warriorAbilities = []warriorAbility{warriorBattleShout, warriorHeroicStrike, warriorCharge}

type warriorLevelingTemplate struct{}

func init() {
	RegisterLevelingTemplate("warrior_leveling", warriorLevelingTemplate{}, func(player *proto.Player) bool {
		return player.GetTankWarrior() != nil || player.GetWarrior() != nil
	})
}

func (warriorLevelingTemplate) Knobs() []Knob { return WarriorLeveling{}.Knobs() }

func (warriorLevelingTemplate) Build(level int32, k Knobs) *proto.APLRotation {
	return WarriorLeveling{Level: level}.Build(k)
}

func (warriorLevelingTemplate) Abilities(level int32, hasTalent func(string) bool) []string {
	return WarriorLeveling{Level: level}.Abilities(hasTalent)
}

func (warriorLevelingTemplate) NewRank(level int32) bool {
	return WarriorLeveling{Level: level}.NewRank()
}

func (WarriorLeveling) Knobs() []Knob {
	return []Knob{
		{Name: "restHealth", Default: 60, Min: 20, Max: 90, Step: 10, Kind: KnobThreshold},
		{Name: "restUntil", Default: 90, Min: 60, Max: 100, Step: 10, Kind: KnobThreshold},
		{Name: "charge", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"charge"}},
		{Name: "battleShout", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"battleShout"}},
		{Name: "heroicStrikeRage", Default: 40, Min: 10, Max: 80, Step: 10, Needs: []string{"heroicStrike"}},
		{Name: "chainMax", Default: 0, Min: 0, Max: 3, Step: 1},
		{Name: "chainHealth", Default: 70, Min: 40, Max: 90, Step: 10, Kind: KnobThreshold},
	}
}

func (t WarriorLeveling) Build(k Knobs) *proto.APLRotation {
	top := func(a warriorAbility) *proto.ActionID { return topRankID(t.Level, a.levels, []int32{a.top}) }
	noMobs := Eq(MobsPulled(), Num(0))
	inCombat := Ge(MobsPulled(), Num(1))
	needRest := Lt(CurrentHealthPercent(), Pct(k["restHealth"]))
	chainFit := Gt(CurrentHealthPercent(), Pct(k["chainHealth"]))
	pullFit := Or(And(inCombat, chainFit), And(noMobs, Not(needRest)))

	items := []*proto.APLListItem{
		Rest(true, false, k["restUntil"], 0, needRest, "Sit down to eat when health is low."),
	}
	if charge := top(warriorCharge); k["charge"] == 1 && charge.GetSpellId() != 0 {
		items = append(items, CastAt(charge, PullTarget, pullFit, "Charge the mob we found."))
	}
	items = append(items, WalkToPullTarget(0, pullFit, "Walk up to the mob we found until it notices us."))
	if k["chainMax"] > 0 {
		items = append(items, LookForMob(true, And(inCombat, Lt(MobsPulled(), Num(k["chainMax"])), chainFit), "Look for another mob while we fight."))
	}
	if shout := top(warriorBattleShout); k["battleShout"] == 1 {
		items = append(items, Cast(shout, And(inCombat, Not(AuraIsActive(shout))), "Battle Shout when it's down."))
	}
	items = append(items, Cast(top(warriorHeroicStrike), And(inCombat, Ge(CurrentRage(), Num(k["heroicStrikeRage"]))), "Heroic Strike with rage to spare."))
	items = append(items, AutocastOtherCooldowns(inCombat, "Racials and trinkets in combat."))
	return Rotation(items...)
}

// Abilities are what the rotation can use at the level (LevelingTemplate).
func (t WarriorLeveling) Abilities(_ func(name string) bool) []string {
	var out []string
	for _, a := range warriorAbilities {
		if core.HighestRankAt(t.Level, a.levels) > 0 {
			out = append(out, a.name)
		}
	}
	return out
}

// NewRank is whether an ability gets a new rank at the level, past its first.
func (t WarriorLeveling) NewRank() bool {
	for _, a := range warriorAbilities {
		for rank := 2; rank < len(a.levels); rank++ {
			if a.levels[rank] == t.Level {
				return true
			}
		}
	}
	return false
}
