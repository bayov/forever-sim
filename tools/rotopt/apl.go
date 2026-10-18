package main

import "github.com/wowsims/classic/tools/rotopt/aplkit"

// The APL builders live in aplkit, which tools/levelingplan shares. These short names keep
// the templates here as they were.

type value = aplkit.Value

var (
	constant               = aplkit.Constant
	num                    = aplkit.Num
	seconds                = aplkit.Seconds
	and                    = aplkit.And
	or                     = aplkit.Or
	nonNil                 = aplkit.NonNil
	not                    = aplkit.Not
	cmp                    = aplkit.Cmp
	ge                     = aplkit.Ge
	gt                     = aplkit.Gt
	le                     = aplkit.Le
	lt                     = aplkit.Lt
	eq                     = aplkit.Eq
	add                    = aplkit.Add
	sub                    = aplkit.Sub
	mainHandTimeToNext     = aplkit.MainHandTimeToNext
	mainHandSwingTime      = aplkit.MainHandSwingTime
	comboPoints            = aplkit.ComboPoints
	energy                 = aplkit.Energy
	remainingTime          = aplkit.RemainingTime
	auraIsActive           = aplkit.AuraIsActive
	auraRemainingTime      = aplkit.AuraRemainingTime
	spellIsReady           = aplkit.SpellIsReady
	spellTimeToReady       = aplkit.SpellTimeToReady
	spellUsesRemaining     = aplkit.SpellUsesRemaining
	spellUsesLostByDelay   = aplkit.SpellUsesLostByDelay
	spellID                = aplkit.SpellID
	spellIDRank            = aplkit.SpellIDRank
	spellIDTag             = aplkit.SpellIDTag
	itemID                 = aplkit.ItemID
	cast                   = aplkit.Cast
	autocastOtherCooldowns = aplkit.AutocastOtherCooldowns
	rotation               = aplkit.Rotation
	prepull                = aplkit.Prepull
	spellCurrentCost       = aplkit.SpellCurrentCost
	targetAuraIsActive     = aplkit.TargetAuraIsActive
	dotIsActive            = aplkit.DotIsActive
	dotRemainingTime       = aplkit.DotRemainingTime
	auraNumStacks          = aplkit.AuraNumStacks
	inMeleeRange           = aplkit.InMeleeRange
	currentManaPercent     = aplkit.CurrentManaPercent
	currentHealthPercent   = aplkit.CurrentHealthPercent
	targetHealthPercent    = aplkit.TargetHealthPercent
	pct                    = aplkit.Pct
	pullTargetDistance     = aplkit.PullTargetDistance
	mobsOnUs               = aplkit.MobsOnUs
	mobsRunningIn          = aplkit.MobsRunningIn
	mobsPulled             = aplkit.MobsPulled
	isMoving               = aplkit.IsMoving
	bandageReady           = aplkit.BandageReady
	totemDistance          = aplkit.TotemDistance
	totemRemainingTime     = aplkit.TotemRemainingTime
	castAt                 = aplkit.CastAt
	action                 = aplkit.Action
	lookForMob             = aplkit.LookForMob
	walkToPullTarget       = aplkit.WalkToPullTarget
	walkToTotem            = aplkit.WalkToTotem
	rest                   = aplkit.Rest
	bandage                = aplkit.Bandage
	pullTarget             = aplkit.PullTarget
)
