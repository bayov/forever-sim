package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

type APLRotation struct {
	unit           *Unit
	prepullActions []*APLAction
	priorityList   []*APLAction

	// Action currently controlling this rotation (only used for certain actions, such as StrictSequence).
	controllingActions []APLActionImpl

	// Value that should evaluate to 'true' if the current channel is to be interrupted.
	// Will be nil when there is no active channel.
	interruptChannelIf APLValue

	// If true, can recast channel when interrupted.
	allowChannelRecastOnInterrupt bool

	// Used inside of actions/value to determine whether they will occur during the prepull or regular rotation.
	parsingPrepull bool

	// Used to avoid recursive APL loops.
	inLoop bool

	// Whether DoNextAction runs for the rotation's own wake-up, see apl_idle.go.
	wakeUp bool
	// When the last check found nothing to do: the sim's event count then, and the earliest
	// time an action could be ready, once we worked it out. See canSkipCheck.
	idle        bool
	idleEvents  uint64
	idleBounded bool
	idleUntil   time.Duration
	// Whether we couldn't tell that time, and nothing but our wake-ups happened since.
	idleHopeless bool
	// Room for readyNotBefore to keep when each action could be ready.
	idleBounds []time.Duration
	// How many times we worked out that time, and how many checks it let us skip, in the current
	// round of aplIdleTries, and how many more wake-ups we don't try for. See aplIdleMinSkips.
	idleTries int
	idleSkips int
	idlePause int

	// The config of each value, until shareConditions. See aplMemo.
	valueConfigs map[APLValue]string
	// Whether values in more than one line keep their results within a check.
	keepsResults bool

	// Validation warnings that occur during proto parsing.
	// We return these back to the user for display in the UI.
	curWarnings          []string
	prepullWarnings      [][]string
	priorityListWarnings [][]string
}

func (rot *APLRotation) ValidationWarning(message string, vals ...interface{}) {
	warning := fmt.Sprintf(message, vals...)
	rot.curWarnings = append(rot.curWarnings, warning)
}

// Invokes the fn function, and attributes all warnings generated during its invocation
// to the provided warningsList.
func (rot *APLRotation) doAndRecordWarnings(warningsList *[]string, isPrepull bool, fn func()) {
	rot.parsingPrepull = isPrepull
	fn()
	if warningsList != nil {
		*warningsList = append(*warningsList, rot.curWarnings...)
	}
	rot.curWarnings = nil
	rot.parsingPrepull = false
}

func (unit *Unit) newCustomRotation() *APLRotation {
	return unit.newAPLRotation(&proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PriorityList: []*proto.APLListItem{
			{
				Action: &proto.APLAction{
					Action: &proto.APLAction_CustomRotation{},
				},
			},
		},
	})
}

func (unit *Unit) newAPLRotation(config *proto.APLRotation) *APLRotation {
	if config == nil {
		return nil
	}

	rotation := &APLRotation{
		unit:                 unit,
		prepullWarnings:      make([][]string, len(config.PrepullActions)),
		priorityListWarnings: make([][]string, len(config.PriorityList)),
	}

	// Parse prepull actions
	for i, prepullItem := range config.PrepullActions {
		prepullIdx := i // Save to local variable for correct lambda capture behavior
		rotation.doAndRecordWarnings(&rotation.prepullWarnings[prepullIdx], true, func() {
			if !prepullItem.Hide {
				doAtVal := rotation.newAPLValue(prepullItem.DoAtValue)
				if doAtVal != nil {
					doAt := doAtVal.GetDuration(nil)
					if doAt > 0 {
						rotation.ValidationWarning("Invalid time for 'Do At', ignoring this Prepull Action")
					} else {
						action := rotation.newAPLAction(prepullItem.Action)
						if action != nil {
							rotation.prepullActions = append(rotation.prepullActions, action)
							unit.RegisterPrepullAction(doAt, func(sim *Simulation) {
								// Warnings for prepull cast failure are detected by running a fake prepull,
								// so this action.Execute needs to record warnings.
								rotation.doAndRecordWarnings(&rotation.prepullWarnings[prepullIdx], true, func() {
									action.Execute(sim)
								})
							})
						}
					}
				}
			}
		})
	}

	// Parse priority list
	var configIdxs []int
	for i, aplItem := range config.PriorityList {
		rotation.doAndRecordWarnings(&rotation.priorityListWarnings[i], false, func() {
			if !aplItem.Hide {
				action := rotation.newAPLAction(aplItem.Action)
				if action != nil {
					rotation.priorityList = append(rotation.priorityList, action)
					configIdxs = append(configIdxs, i)
				}
			}
		})
	}

	rotation.priorityList, configIdxs = rotation.limitHealRanks(rotation.priorityList, configIdxs)

	// Finalize
	for i, action := range rotation.prepullActions {
		rotation.doAndRecordWarnings(&rotation.prepullWarnings[i], true, func() {
			action.Finalize(rotation)
		})
	}
	for i, action := range rotation.priorityList {
		rotation.doAndRecordWarnings(&rotation.priorityListWarnings[configIdxs[i]], false, func() {
			action.Finalize(rotation)
		})
	}
	// After the warnings, which go by the user's lines.
	if defaults := rotation.levelingDefaultRests(config); len(defaults) > 0 {
		rotation.priorityList = append(defaults, rotation.priorityList...)
	}
	rotation.shareConditions()

	// Remove MCDs that are referenced by APL actions, so that the Autocast Other Cooldowns
	// action does not include them.
	agent := unit.Env.GetAgentFromUnit(unit)
	if agent != nil {
		character := agent.GetCharacter()
		for _, action := range rotation.allAPLActions() {
			if castSpellAction, ok := action.impl.(*APLActionCastSpell); ok {
				character.removeInitialMajorCooldown(castSpellAction.spell.ActionID)
			}
		}
	}

	return rotation
}
func (rot *APLRotation) getStats() *proto.APLStats {
	return &proto.APLStats{
		PrepullActions: MapSlice(rot.prepullWarnings, func(warnings []string) *proto.APLActionStats { return &proto.APLActionStats{Warnings: warnings} }),
		PriorityList:   MapSlice(rot.priorityListWarnings, func(warnings []string) *proto.APLActionStats { return &proto.APLActionStats{Warnings: warnings} }),
	}
}

// Returns all action objects as an unstructured list. Used for easily finding specific actions.
// MaxBoundHealRanks is how many ranks of one heal a rotation may cast. A healer keeps 2 to 4 ranks
// of a heal on their bars and picks one by what's missing (Spell Expected Healing), never the
// exact rank out of all of them (docs/healing_sim.md).
const MaxBoundHealRanks = 4

// limitHealRanks drops each priority list line that would cast a rank of a heal past the first
// MaxBoundHealRanks ranks of it the list names, with a warning on the line.
func (rot *APLRotation) limitHealRanks(list []*APLAction, configIdxs []int) ([]*APLAction, []int) {
	bound := map[int32]map[int32]bool{} // a heal's top rank: the ranks the list casts
	var keptList []*APLAction
	var keptIdxs []int
	for i, action := range list {
		var extra []*Spell
		ranks := map[int32][]int32{}
		for _, inner := range action.GetAllActions() {
			cast, ok := inner.impl.(*APLActionCastSpell)
			if !ok || !cast.spell.ProcMask.Matches(ProcMaskSpellHealing) {
				continue
			}
			id := cast.spell.ActionID.SpellID
			top := TopRankID(id)
			if top == 0 || bound[top][id] {
				continue
			}
			if len(bound[top])+len(ranks[top]) >= MaxBoundHealRanks && !slices.Contains(ranks[top], id) {
				extra = append(extra, cast.spell)
				continue
			}
			if !slices.Contains(ranks[top], id) {
				ranks[top] = append(ranks[top], id)
			}
		}
		if len(extra) > 0 {
			rot.doAndRecordWarnings(&rot.priorityListWarnings[configIdxs[i]], false, func() {
				rot.ValidationWarning("%s would be a %dth rank of its heal: a healer binds at most %d, so this line is off", extra[0].ActionID, MaxBoundHealRanks+1, MaxBoundHealRanks)
			})
			continue
		}
		for top, ids := range ranks {
			if bound[top] == nil {
				bound[top] = map[int32]bool{}
			}
			for _, id := range ids {
				bound[top][id] = true
			}
		}
		keptList = append(keptList, action)
		keptIdxs = append(keptIdxs, configIdxs[i])
	}
	return keptList, keptIdxs
}

func (rot *APLRotation) allAPLActions() []*APLAction {
	return Flatten(MapSlice(rot.priorityList, func(action *APLAction) []*APLAction { return action.GetAllActions() }))
}

// Returns all action objects from the prepull as an unstructured list. Used for easily finding specific actions.
func (rot *APLRotation) allPrepullActions() []*APLAction {
	return Flatten(MapSlice(rot.prepullActions, func(action *APLAction) []*APLAction { return action.GetAllActions() }))
}

func (rot *APLRotation) reset(sim *Simulation) {
	rot.controllingActions = nil
	rot.inLoop = false
	rot.idle = false
	rot.idleHopeless = false
	rot.interruptChannelIf = nil
	rot.allowChannelRecastOnInterrupt = false
	for _, action := range rot.allAPLActions() {
		action.impl.Reset(sim)
	}
}

// We intentionally try to mimic the behavior of simc APL to avoid confusion
// and leverage the community's existing familiarity.
// https://github.com/simulationcraft/simc/wiki/ActionLists
func (apl *APLRotation) DoNextAction(sim *Simulation) {
	if sim.CurrentTime < 0 {
		return
	}
	// A player that died in a healing fight (player_death.go).
	if apl.unit.dead && apl.unit.Type != EnemyUnit {
		return
	}

	// A cast ending now lands before we pick what comes after it. Its completion runs this
	// check itself, or the GCD action due with it does when the cast fails at the end.
	if hc := &apl.unit.Hardcast; hc.Expires == sim.CurrentTime && hc.OnComplete != nil {
		apl.unit.completeDueHardcast(sim)
		return
	}

	if apl.unit.ManualControl {
		return
	}

	if apl.inLoop {
		return
	}

	if apl.shouldInterruptChannel(sim) {
		apl.unit.ChanneledDot.Cancel(sim)
		apl.idle = false
	}

	if apl.unit.IsChanneling(sim) {
		// While Shoot runs the rotation keeps looking as it does with nothing to do, so the
		// next cast it can make ends Shoot when it can be made.
		if wand := apl.unit.wand; wand != nil && wand.wanding() && apl.unit.GCD.IsReady(sim) {
			apl.unit.WaitUntil(sim, sim.CurrentTime+time.Millisecond*50)
		}
		return
	}

	if apl.canSkipCheck(sim) {
		// This check would find nothing to do, like the one before it. We take back the event
		// this wake-up counted, so the next one can skip its check too.
		sim.events--
		if sim.Log != nil {
			apl.unit.Log(sim, "No available actions!")
		}
		// The leveling session hears about it like about any check that finds nothing.
		if sim.Encounter.leveling != nil {
			sim.Encounter.leveling.onRotationIdle(sim, apl.unit)
		}
		if apl.unit.GCD.IsReady(sim) {
			apl.unit.WaitUntil(sim, sim.CurrentTime+time.Millisecond*50)
		}
		return
	}
	apl.idle = false

	i := 0
	apl.inLoop = true

	for nextAction := apl.getNextAction(sim); nextAction != nil; i, nextAction = i+1, apl.getNextAction(sim) {
		if i > 1000 {
			panic(fmt.Sprintf("[USER_ERROR] Infinite loop detected, current action:\n%s", nextAction))
		}

		nextAction.Execute(sim)
	}
	apl.inLoop = false

	if sim.Log != nil && i == 0 {
		apl.unit.Log(sim, "No available actions!")
	}
	if i == 0 && sim.Encounter.leveling != nil {
		sim.Encounter.leveling.onRotationIdle(sim, apl.unit)
	}
	if i == 0 {
		apl.noteIdle(sim)
	}

	gcdReady := apl.unit.GCD.IsReady(sim)
	if gcdReady {
		apl.unit.WaitUntil(sim, sim.CurrentTime+time.Millisecond*50)
	}
}

func (apl *APLRotation) getNextAction(sim *Simulation) *APLAction {
	if len(apl.controllingActions) != 0 {
		return apl.controllingActions[len(apl.controllingActions)-1].GetNextAction(sim)
	}

	if !apl.keepsResults {
		return apl.firstReady(sim)
	}
	outer := sim.startAPLCheck()
	action := apl.firstReady(sim)
	sim.endAPLCheck(outer)
	return action
}

func (apl *APLRotation) firstReady(sim *Simulation) *APLAction {
	for _, action := range apl.priorityList {
		if action.IsReady(sim) {
			return action
		}
	}
	return nil
}

func (apl *APLRotation) pushControllingAction(ca APLActionImpl) {
	apl.controllingActions = append(apl.controllingActions, ca)
}

func (apl *APLRotation) popControllingAction(ca APLActionImpl) {
	if len(apl.controllingActions) == 0 || apl.controllingActions[len(apl.controllingActions)-1] != ca {
		panic("Wrong APL controllingAction in pop()")
	}
	apl.controllingActions = apl.controllingActions[:len(apl.controllingActions)-1]
}

func (apl *APLRotation) shouldInterruptChannel(sim *Simulation) bool {
	channeledDot := apl.unit.ChanneledDot

	if channeledDot == nil {
		return false
	}

	// Shoot gives way to the next cast whenever the rotation looks, as at each shot.
	if apl.unit.wand != nil && channeledDot.Spell == apl.unit.wand.Shoot {
		return apl.unit.wand.givesWay(sim)
	}

	if channeledDot.MaxTicksRemaining() == 0 {
		// Channel has ended, but apl.unit.ChanneledDot hasn't been cleared yet meaning the aura is still active.
		return false
	}

	if apl.interruptChannelIf == nil || !apl.interruptChannelIf.GetBool(sim) {
		// Continue the channel.
		return false
	}

	// Allow next action to interrupt the channel, but if the action is the same action then it still needs to continue.
	// CanCast refuses every spell while the unit channels, so the channel is set aside while
	// the next action is picked. Without that no action is ever ready and interruptIf never fires.
	apl.unit.ChanneledDot = nil
	nextAction := apl.getNextAction(sim)
	apl.unit.ChanneledDot = channeledDot
	if nextAction == nil {
		return false
	}

	// Wanding is what the rotation does with nothing else to do, so it doesn't cut a channel short.
	if apl.unit.wand.isShootAction(nextAction) {
		return false
	}

	if channelAction, ok := nextAction.impl.(*APLActionChannelSpell); ok && channelAction.spell == channeledDot.Spell {
		// Newly selected action is channeling the same spell, so continue the channel unless recast is allowed.
		return apl.allowChannelRecastOnInterrupt
	}

	return true
}

func APLRotationFromJsonString(jsonString string) *proto.APLRotation {
	apl := &proto.APLRotation{}
	data := []byte(jsonString)
	if err := protojson.Unmarshal(data, apl); err != nil {
		panic(err)
	}
	return apl
}
