//go:build practice

package practice

import (
	"fmt"
	"sort"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/shaman"
)

// Summary is what the room shows when the fight ends. It covers the time from the pull to the
// end, from the one fight you played.
type Summary struct {
	DurationMs   float64        `json:"durationMs"`
	Damage       float64        `json:"damage"`
	DPS          float64        `json:"dps"`
	Spells       []SpellSummary `json:"spells"`
	Auras        []AuraUptime   `json:"auras"`
	IdleGCDMs    float64        `json:"idleGcdMs"`
	OutOfRangeMs float64        `json:"outOfRangeMs"`
	Casts        int32          `json:"casts"`
}

type SpellSummary struct {
	SpellID int32   `json:"spellId,omitempty"`
	ItemID  int32   `json:"itemId,omitempty"`
	OtherID int32   `json:"otherId,omitempty"`
	Tag     int32   `json:"tag,omitempty"`
	Damage  float64 `json:"damage"`
	Casts   int32   `json:"casts"`
	Hits    int32   `json:"hits"`
	Crits   int32   `json:"crits"`
	Misses  int32   `json:"misses"`
	Ticks   int32   `json:"ticks"`
}

type AuraUptime struct {
	SpellID   int32   `json:"spellId,omitempty"`
	ItemID    int32   `json:"itemId,omitempty"`
	UptimePct float64 `json:"uptimePct"`
}

// summary adds up the fight once the sim's Cleanup ran.
func (r *run) summary(end time.Duration) *Summary {
	duration := max(end, 0)
	s := &Summary{
		DurationMs:   toMs(duration),
		Damage:       r.damage,
		OutOfRangeMs: toMs(r.outOfRange),
	}
	if duration > 0 {
		s.DPS = r.damage / duration.Seconds()
	}
	idle := r.idleGCD
	if duration > max(r.busyUntil, 0) {
		idle += duration - max(r.busyUntil, 0)
	}
	s.IdleGCDMs = toMs(idle)

	keys := map[hitKey]bool{}
	for k := range r.hits {
		keys[k] = true
	}
	for k := range r.casts {
		keys[k] = true
	}
	for k := range keys {
		sum := SpellSummary{SpellID: k.spellID, ItemID: k.itemID, OtherID: k.otherID, Tag: k.tag}
		if h := r.hits[k]; h != nil {
			sum = *h
		}
		sum.Casts = r.casts[k]
		s.Casts += sum.Casts
		s.Spells = append(s.Spells, sum)
	}
	sort.Slice(s.Spells, func(i, j int) bool {
		if s.Spells[i].Damage != s.Spells[j].Damage {
			return s.Spells[i].Damage > s.Spells[j].Damage
		}
		return s.Spells[i].SpellID < s.Spells[j].SpellID
	})

	metrics := r.char.GetMetricsProto()
	for _, aura := range metrics.Auras {
		if aura.UptimeSecondsAvg <= 0 || duration <= 0 {
			continue
		}
		id := aura.Id
		up := AuraUptime{UptimePct: 100 * aura.UptimeSecondsAvg / duration.Seconds()}
		switch raw := id.GetRawId().(type) {
		case *proto.ActionID_SpellId:
			up.SpellID = raw.SpellId
		case *proto.ActionID_ItemId:
			up.ItemID = raw.ItemId
		default:
			continue
		}
		if up.UptimePct >= 99.99 {
			continue // buffs from the settings, up the whole fight
		}
		s.Auras = append(s.Auras, up)
	}
	sort.Slice(s.Auras, func(i, j int) bool { return s.Auras[i].UptimePct > s.Auras[j].UptimePct })
	return s
}

func formatLog(at time.Duration, message string, vals ...interface{}) string {
	return fmt.Sprintf("[%0.2f] "+message, append([]interface{}{at.Seconds()}, vals...)...)
}

// Toolkit lists your castable spells and usable items, for the spellbook and the bars.
func (s *Session) Toolkit() []Spell {
	r := s.run
	var out []Spell
	add := func(spell *core.Spell, kind string) {
		entry := Spell{
			SpellID:  spell.ActionID.SpellID,
			ItemID:   spell.ActionID.ItemID,
			Tag:      spell.ActionID.Tag,
			CastMs:   toMs(spell.DefaultCast.CastTime),
			GCDMs:    toMs(spell.DefaultCast.GCD),
			CDMs:     toMs(spell.CD.Duration),
			MaxRange: spell.MaxRange,
			Melee:    spell.ProcMask.Matches(core.ProcMaskMelee),
			Harmful:  harmful(spell),
			Kind:     kind,
		}
		if spell.Cost != nil {
			entry.Cost = spell.Cost.GetCurrentCost()
		}
		if spell.Flags.Matches(shaman.SpellFlagTotem) {
			entry.Kind = "totem"
		}
		out = append(out, entry)
	}
	for _, spell := range r.toolkit {
		add(spell, "spell")
	}
	for _, spell := range r.items {
		add(spell, "item")
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		if out[i].SpellID != out[j].SpellID {
			return out[i].SpellID < out[j].SpellID
		}
		return out[i].ItemID < out[j].ItemID
	})
	return out
}
