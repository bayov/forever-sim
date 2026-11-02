package core

import (
	"time"

	goproto "google.golang.org/protobuf/proto"

	"github.com/wowsims/classic/sim/core/proto"
)

// A rotation check looks at the same values many times. The level 30 leveling rotation has 182
// values in its conditions, but only 64 different ones: "Leveling Mobs On Us + Leveling Mobs
// Running In" alone is in 15 lines. So we work out such a value once per check:
//   - The lines share one value object for each value they have in common (shareConditions).
//   - A shared Compare, Math, Max or Min keeps its result until the check ends (aplMemo).
//
// Nothing in the sim changes during a check, so the results are the same. We only keep results
// within a check: while the rotation looks through its lines for one to do (getNextAction), and
// while it works out when an action could be ready (readyNotBefore). A value that can change
// the sim when we look at it (aplHasSideEffects) is neither shared nor kept.
//
// Keeping a result costs a little, and so does a check for one. So only the values in more
// than one place keep their results, in a wrapper (like aplMemoCompare) that takes the place
// of the value there. The other values don't pay for it.

// aplMemo holds a value's result within one check (Simulation.aplCheck).
type aplMemo struct {
	check uint64
	kind  aplMemoKind
	b     bool
	i     int32
	f     float64
	d     time.Duration
}

type aplMemoKind uint8

const (
	aplMemoBool aplMemoKind = iota + 1
	aplMemoInt
	aplMemoFloat
	aplMemoDuration
)

// has is whether the memo holds this check's result of this kind.
func (m *aplMemo) has(sim *Simulation, kind aplMemoKind) bool {
	return sim != nil && sim.aplCheck != 0 && m.check == sim.aplCheck && m.kind == kind
}

// keeps is whether we keep a result now.
func (m *aplMemo) keeps(sim *Simulation, kind aplMemoKind) bool {
	if sim == nil || sim.aplCheck == 0 {
		return false
	}
	m.check, m.kind = sim.aplCheck, kind
	return true
}

// The wrappers that keep a shared value's result. Everything else, like steadyUntil and trend
// in apl_idle.go, comes from the value they wrap.
type aplMemoCompare struct {
	*APLValueCompare
	memo aplMemo
}

func (v *aplMemoCompare) GetBool(sim *Simulation) bool {
	if v.memo.has(sim, aplMemoBool) {
		return v.memo.b
	}
	result := v.APLValueCompare.GetBool(sim)
	if v.memo.keeps(sim, aplMemoBool) {
		v.memo.b = result
	}
	return result
}

type aplMemoMath struct {
	*APLValueMath
	memo aplMemo
}

func (v *aplMemoMath) GetInt(sim *Simulation) int32 {
	if v.memo.has(sim, aplMemoInt) {
		return v.memo.i
	}
	result := v.APLValueMath.GetInt(sim)
	if v.memo.keeps(sim, aplMemoInt) {
		v.memo.i = result
	}
	return result
}
func (v *aplMemoMath) GetFloat(sim *Simulation) float64 {
	if v.memo.has(sim, aplMemoFloat) {
		return v.memo.f
	}
	result := v.APLValueMath.GetFloat(sim)
	if v.memo.keeps(sim, aplMemoFloat) {
		v.memo.f = result
	}
	return result
}
func (v *aplMemoMath) GetDuration(sim *Simulation) time.Duration {
	if v.memo.has(sim, aplMemoDuration) {
		return v.memo.d
	}
	result := v.APLValueMath.GetDuration(sim)
	if v.memo.keeps(sim, aplMemoDuration) {
		v.memo.d = result
	}
	return result
}

type aplMemoMax struct {
	*APLValueMax
	memo aplMemo
}

func (v *aplMemoMax) GetInt(sim *Simulation) int32 {
	if v.memo.has(sim, aplMemoInt) {
		return v.memo.i
	}
	result := v.APLValueMax.GetInt(sim)
	if v.memo.keeps(sim, aplMemoInt) {
		v.memo.i = result
	}
	return result
}
func (v *aplMemoMax) GetFloat(sim *Simulation) float64 {
	if v.memo.has(sim, aplMemoFloat) {
		return v.memo.f
	}
	result := v.APLValueMax.GetFloat(sim)
	if v.memo.keeps(sim, aplMemoFloat) {
		v.memo.f = result
	}
	return result
}
func (v *aplMemoMax) GetDuration(sim *Simulation) time.Duration {
	if v.memo.has(sim, aplMemoDuration) {
		return v.memo.d
	}
	result := v.APLValueMax.GetDuration(sim)
	if v.memo.keeps(sim, aplMemoDuration) {
		v.memo.d = result
	}
	return result
}

type aplMemoMin struct {
	*APLValueMin
	memo aplMemo
}

func (v *aplMemoMin) GetInt(sim *Simulation) int32 {
	if v.memo.has(sim, aplMemoInt) {
		return v.memo.i
	}
	result := v.APLValueMin.GetInt(sim)
	if v.memo.keeps(sim, aplMemoInt) {
		v.memo.i = result
	}
	return result
}
func (v *aplMemoMin) GetFloat(sim *Simulation) float64 {
	if v.memo.has(sim, aplMemoFloat) {
		return v.memo.f
	}
	result := v.APLValueMin.GetFloat(sim)
	if v.memo.keeps(sim, aplMemoFloat) {
		v.memo.f = result
	}
	return result
}
func (v *aplMemoMin) GetDuration(sim *Simulation) time.Duration {
	if v.memo.has(sim, aplMemoDuration) {
		return v.memo.d
	}
	result := v.APLValueMin.GetDuration(sim)
	if v.memo.keeps(sim, aplMemoDuration) {
		v.memo.d = result
	}
	return result
}

// aplMemoWrapper gives the wrapper that keeps the value's results, or nil when we don't keep
// that kind of value's results.
func aplMemoWrapper(value APLValue) APLValue {
	switch v := value.(type) {
	case *APLValueCompare:
		return &aplMemoCompare{APLValueCompare: v}
	case *APLValueMath:
		return &aplMemoMath{APLValueMath: v}
	case *APLValueMax:
		return &aplMemoMax{APLValueMax: v}
	case *APLValueMin:
		return &aplMemoMin{APLValueMin: v}
	}
	return nil
}

// aplUnmemo is the value a wrapper that keeps its results stands for, or the value itself.
func aplUnmemo(value APLValue) APLValue {
	switch v := value.(type) {
	case *aplMemoCompare:
		return v.APLValueCompare
	case *aplMemoMath:
		return v.APLValueMath
	case *aplMemoMax:
		return v.APLValueMax
	case *aplMemoMin:
		return v.APLValueMin
	}
	return value
}

// startAPLCheck starts a check, in which the shared values keep their results. It returns the
// check it interrupts, for endAPLCheck.
func (sim *Simulation) startAPLCheck() uint64 {
	outer := sim.aplCheck
	sim.aplChecks++
	sim.aplCheck = sim.aplChecks
	return outer
}

func (sim *Simulation) endAPLCheck(outer uint64) {
	sim.aplCheck = outer
}

// noteValueConfig remembers the config a value came from, so shareConditions can tell which
// values are the same.
func (rot *APLRotation) noteValueConfig(value APLValue, config *proto.APLValue) {
	if rot.valueConfigs == nil {
		rot.valueConfigs = map[APLValue]string{}
	}
	// A value can stand for more than one config, like an And with one value, which is that
	// value. Its own config is enough.
	if _, ok := rot.valueConfigs[value]; ok {
		return
	}
	data, err := goproto.MarshalOptions{Deterministic: true}.Marshal(config)
	if err != nil {
		return
	}
	rot.valueConfigs[value] = string(data)
}

// shareConditions makes the priority list's lines share one object for each value they have in
// common, and makes the values in more than one place keep their results within a check. See
// aplMemo.
func (rot *APLRotation) shareConditions() {
	// First the lines share the values they have in common, and we count each value's places.
	shared := map[string]APLValue{}
	places := map[APLValue]int{}
	var share func(value APLValue) APLValue
	share = func(value APLValue) APLValue {
		if value == nil {
			return nil
		}
		replaceInnerValues(value, share)
		if aplHasSideEffects(value) {
			return value
		}
		if config, ok := rot.valueConfigs[value]; ok {
			if first, ok := shared[config]; ok {
				value = first
			} else {
				shared[config] = value
			}
		}
		places[value]++
		return value
	}
	// Then each value in more than one place takes its wrapper there.
	wrappers := map[APLValue]APLValue{}
	seen := map[APLValue]bool{}
	var wrap func(value APLValue) APLValue
	wrap = func(value APLValue) APLValue {
		if value == nil {
			return nil
		}
		if !seen[value] {
			seen[value] = true
			replaceInnerValues(value, wrap)
		}
		if places[value] < 2 {
			return value
		}
		wrapper, ok := wrappers[value]
		if !ok {
			wrapper = aplMemoWrapper(value)
			if wrapper == nil {
				wrapper = value
			} else {
				rot.keepsResults = true
			}
			wrappers[value] = wrapper
		}
		return wrapper
	}
	var actions []*APLAction
	for _, line := range rot.priorityList {
		actions = append(actions, line.GetAllActions()...)
	}
	for _, action := range actions {
		action.condition = share(action.condition)
	}
	for _, action := range actions {
		action.condition = wrap(action.condition)
	}
	rot.valueConfigs = nil
}

// replaceInnerValues replaces each inner value of an operator with what replace gives for it.
func replaceInnerValues(value APLValue, replace func(APLValue) APLValue) {
	switch v := value.(type) {
	case *APLValueAnd:
		replaceAll(v.vals, replace)
	case *APLValueOr:
		replaceAll(v.vals, replace)
	case *APLValueMax:
		replaceAll(v.vals, replace)
	case *APLValueMin:
		replaceAll(v.vals, replace)
	case *APLValueNot:
		v.val = replace(v.val)
	case *APLValueCompare:
		v.lhs, v.rhs = replace(v.lhs), replace(v.rhs)
	case *APLValueMath:
		v.lhs, v.rhs = replace(v.lhs), replace(v.rhs)
	case *APLValueCoerced:
		v.inner = replace(v.inner)
	}
}

func replaceAll(values []APLValue, replace func(APLValue) APLValue) {
	for i, value := range values {
		values[i] = replace(value)
	}
}
