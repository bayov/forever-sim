package aplkit

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

// WarriorLeveling is a warrior's leveling rotation, for any of the weapon styles. It plays the
// Protection Warrior spec, which is the DPS spec's warrior with Shield Block and Shield Wall on
// top. So one spec has every ability with a shield and with a two-hander, and the builds compare.
//
// The rules (docs/leveling_warrior.md has what each ability does in a session):
//
//   - Out of combat: Cannibalize, a bandage, or sitting down to eat.
//   - Pulls: Charge, then the stance we fight in. With gather we first run around to collect a
//     pack, with the mobs we have following, and fight them all where the last one joins. After
//     that we can chain pull while fighting, as the shaman does.
//   - The stance: one to fight few mobs in and one for packs. With Sweeping Strikes ready we open a
//     pack in Battle Stance.
//   - The pack: Demoralizing Shout, Thunder Clap, Cleave, Whirlwind, Rend on every mob, Shield
//     Block, Retaliation, and the cooldowns below some health.
//   - The fight: Victory Rush, Execute, Overpower, Revenge, the main strike, Slam, Sunder Armor,
//     Heroic Strike with rage to spare.
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
	warriorBattleShout   = warriorAbility{"battleShout", core.BattleShoutSpellId[6], []int32{0, 1, 12, 22, 32, 42, 52}}
	warriorHeroicStrike  = warriorAbility{"heroicStrike", 11567, []int32{0, 1, 8, 16, 24, 32, 40, 48, 56}}
	warriorCharge        = warriorAbility{"charge", 11578, []int32{0, 4, 26, 46}}
	warriorRend          = warriorAbility{"rend", 11574, []int32{0, 4, 10, 20, 30, 40, 50, 60}}
	warriorThunderClap   = warriorAbility{"thunderClap", 11581, []int32{0, 6, 18, 28, 38, 48, 58}}
	warriorBloodrage     = warriorAbility{"bloodrage", 2687, []int32{0, 10}}
	warriorDefStance     = warriorAbility{"defensiveStance", 71, []int32{0, 10}}
	warriorSunder        = warriorAbility{"sunderArmor", 11597, []int32{0, 10, 22, 34, 46, 58}}
	warriorOverpower     = warriorAbility{"overpower", 11585, []int32{0, 12, 28, 44, 60}}
	warriorDemoShout     = warriorAbility{"demoralizingShout", 11556, []int32{0, 14, 24, 34, 44, 54}}
	warriorRevenge       = warriorAbility{"revenge", 11601, []int32{0, 14, 24, 34, 44, 54}}
	warriorShieldBlock   = warriorAbility{"shieldBlock", 2565, []int32{0, 16}}
	warriorCleave        = warriorAbility{"cleave", 20569, []int32{0, 20, 30, 40, 50, 60}}
	warriorSlam          = warriorAbility{"slam", 11605, []int32{0, 20, 30, 38, 46, 54}}
	warriorRetaliation   = warriorAbility{"retaliation", 20230, []int32{0, 20}}
	warriorVictoryRush   = warriorAbility{"victoryRush", 402927, []int32{0, 20}}
	warriorExecute       = warriorAbility{"execute", 20662, []int32{0, 24, 32, 40, 48, 56}}
	warriorShieldWall    = warriorAbility{"shieldWall", 871, []int32{0, 28}}
	warriorBerserker     = warriorAbility{"berserkerStance", 2458, []int32{0, 30}}
	warriorBerserkerRage = warriorAbility{"berserkerRage", 18499, []int32{0, 30}}
	warriorIntercept     = warriorAbility{"intercept", 20617, []int32{0, 30, 42, 52}}
	warriorWhirlwind     = warriorAbility{"whirlwind", 1680, []int32{0, 36}}
	warriorRecklessness  = warriorAbility{"recklessness", 1719, []int32{0, 50}}
)

var warriorAbilities = []warriorAbility{
	warriorBattleShout, warriorHeroicStrike, warriorCharge, warriorRend, warriorThunderClap, warriorBloodrage, warriorDefStance,
	warriorSunder, warriorOverpower, warriorDemoShout, warriorRevenge, warriorShieldBlock, warriorCleave, warriorSlam,
	warriorRetaliation, warriorVictoryRush, warriorExecute, warriorShieldWall, warriorBerserker, warriorBerserkerRage,
	warriorIntercept, warriorWhirlwind, warriorRecklessness,
}

// The abilities that are talents, by their field names in the talent tree.
var warriorTalentAbilities = []string{"deathWish", "recklessSwings", "bloodthirst", "mortalStrike", "shieldSlam", "concussionBlow",
	"lastStand", "sweepingStrikes", "spearingStrike", "vanguard", "improvedSlam", "goreDrinker"}

type warriorLevelingTemplate struct{}

func init() {
	RegisterLevelingTemplate("warrior_leveling", warriorLevelingTemplate{}, func(player *proto.Player) bool {
		return player.GetTankWarrior() != nil
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
		// Out of combat.
		{Name: "restHealth", Default: 60, Min: 20, Max: 90, Step: 10, Kind: KnobThreshold},
		{Name: "restUntil", Default: 90, Min: 60, Max: 100, Step: 10, Kind: KnobThreshold},
		{Name: "bandageBelow", Default: 60, Min: 0, Max: 80, Step: 20, Kind: KnobThreshold},
		{Name: "cannibalizeBelow", Default: 60, Min: 0, Max: 80, Step: 20, Kind: KnobSteady},
		// Pulling. Intercept pulls in combat from Berserker Stance.
		{Name: "charge", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"charge", "vanguard"}},
		{Name: "intercept", Default: 0, Min: 0, Max: 1, Step: 1, Needs: []string{"intercept"}},
		// Gathering: run around for mobs until the pull has taken this many (0 and 1 don't),
		// above gatherHealth. Then we fight.
		{Name: "gather", Default: 0, Min: 0, Max: 10, Step: 1},
		{Name: "gatherHealth", Default: 70, Min: 30, Max: 90, Step: 10, Kind: KnobThreshold},
		// Chain pulls while fighting: look while fewer than chainMax mobs are pulled (0 off) and
		// health is above lookAheadHealth, and pull above chainHealth. chainStand 1 stands and
		// fights while we look, 0 runs around.
		{Name: "chainMax", Default: 2, Min: 0, Max: 10, Step: 1},
		{Name: "chainHealth", Default: 70, Min: 30, Max: 90, Step: 10, Kind: KnobThreshold},
		{Name: "lookAheadHealth", Default: 40, Min: 0, Max: 70, Step: 10, Kind: KnobThreshold},
		{Name: "chainStand", Default: 1, Min: 0, Max: 1, Step: 1, Kind: KnobSteady},
		// The furthest mob we walk to with mobs on us, in yards, chain pulling and gathering. In
		// a camp the next mob can be up to 100 yards away.
		{Name: "chainDistance", Default: 30, Min: 10, Max: 100, Step: 10, Kind: KnobThreshold},
		{Name: "gatherDistance", Default: 60, Min: 30, Max: 100, Step: 10, Kind: KnobThreshold},
		// Pull the next mob while the only one on us is under this share of its health, above
		// chainDyingHealth. 0 off.
		{Name: "chainDyingBelow", Default: 20, Min: 0, Max: 40, Step: 10},
		{Name: "chainDyingHealth", Default: 50, Min: 30, Max: 90, Step: 10, Kind: KnobThreshold},
		// Stances: 0 Battle, 1 Defensive, 2 Berserker, for few mobs and for packMobs or more.
		{Name: "stance", Default: 0, Min: 0, Max: 2, Step: 1, Needs: []string{"defensiveStance", "berserkerStance"}},
		{Name: "packStance", Default: 0, Min: 0, Max: 2, Step: 1, Needs: []string{"defensiveStance", "berserkerStance"}},
		{Name: "packMobs", Default: 2, Min: 2, Max: 5, Step: 1, Needs: []string{"defensiveStance", "berserkerStance"}},
		// Swap only with no more rage than this, since a swap loses what Tactical Mastery
		// doesn't keep.
		{Name: "swapRage", Default: 30, Min: 10, Max: 100, Step: 10, Needs: []string{"defensiveStance", "berserkerStance"}, Kind: KnobThreshold},
		// Shouts and debuffs, from this many mobs on us (0 off). Thunder Clap on cooldown (2)
		// or only to keep its slow up (1).
		{Name: "battleShout", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"battleShout"}, Kind: KnobSteady},
		{Name: "demoShoutMobs", Default: 0, Min: 0, Max: 4, Step: 1, Needs: []string{"demoralizingShout"}},
		{Name: "thunderClap", Default: 1, Min: 0, Max: 2, Step: 1, Needs: []string{"thunderClap"}},
		{Name: "thunderClapMobs", Default: 3, Min: 1, Max: 4, Step: 1, Needs: []string{"thunderClap"}},
		{Name: "sunderStacks", Default: 0, Min: 0, Max: 5, Step: 1, Needs: []string{"sunderArmor"}},
		// Rend on a target above rendAbove of its health: 1 our target, 2 every mob on us.
		{Name: "rend", Default: 1, Min: 0, Max: 2, Step: 1, Needs: []string{"rend"}},
		{Name: "rendAbove", Default: 50, Min: 10, Max: 90, Step: 20, Needs: []string{"rend"}, Kind: KnobThreshold},
		// Main strikes.
		{Name: "overpower", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"overpower"}},
		{Name: "revenge", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"revenge"}},
		{Name: "execute", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"execute"}},
		{Name: "executeMaxMobs", Default: 5, Min: 1, Max: 5, Step: 1, Needs: []string{"execute"}},
		{Name: "strike", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"mortalStrike", "bloodthirst", "shieldSlam"}},
		{Name: "whirlwindMobs", Default: 2, Min: 0, Max: 4, Step: 1, Needs: []string{"whirlwind"}},
		{Name: "spearingStrike", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"spearingStrike"}},
		{Name: "slam", Default: 0, Min: 0, Max: 1, Step: 1, Needs: []string{"slam", "improvedSlam"}},
		{Name: "victoryRushBelow", Default: 100, Min: 0, Max: 100, Step: 20, Needs: []string{"victoryRush"}, Kind: KnobThreshold},
		// Rage dumps. Cleave takes Heroic Strike's place with two mobs or more on us.
		{Name: "heroicStrikeRage", Default: 40, Min: 10, Max: 100, Step: 10, Needs: []string{"heroicStrike"}, Kind: KnobThreshold},
		{Name: "cleaveRage", Default: 40, Min: 0, Max: 100, Step: 20, Needs: []string{"cleave"}, Kind: KnobThreshold},
		// Cooldowns. The health ones are a share of health, and 0 is off.
		{Name: "bloodrageAbove", Default: 50, Min: 0, Max: 100, Step: 10, Needs: []string{"bloodrage"}, Kind: KnobThreshold},
		{Name: "bloodrageEarly", Default: 0, Min: 0, Max: 1, Step: 1, Needs: []string{"bloodrage"}},
		{Name: "berserkerRage", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"berserkerRage"}, Kind: KnobSteady},
		{Name: "deathWish", Default: 1, Min: 0, Max: 3, Step: 1, Needs: []string{"deathWish"}},
		{Name: "recklessSwingsAbove", Default: 0, Min: 0, Max: 90, Step: 30, Needs: []string{"recklessSwings"}, Kind: KnobThreshold},
		// Gore Drinker heals with our next three hits after Bloodrage, Reckless Swings or Berserker Rage.
		// Below this health we use them for that at any rage, when its heals have run out.
		{Name: "goreFeedBelow", Default: 80, Min: 0, Max: 100, Step: 20, Needs: []string{"goreDrinker"}, Kind: KnobThreshold},
		{Name: "goreDance", Default: 0, Min: 0, Max: 1, Step: 1, Needs: []string{"goreDrinker", "berserkerRage"}},
		{Name: "recklessness", Default: 0, Min: 0, Max: 1, Step: 1, Needs: []string{"recklessness"}, Kind: KnobSteady},
		{Name: "sweepingStrikes", Default: 1, Min: 0, Max: 1, Step: 1, Needs: []string{"sweepingStrikes"}},
		{Name: "concussionBlowMobs", Default: 0, Min: 0, Max: 3, Step: 1, Needs: []string{"concussionBlow"}},
		{Name: "shieldBlockMobs", Default: 2, Min: 0, Max: 4, Step: 1, Needs: []string{"shieldBlock"}},
		{Name: "retaliationMobs", Default: 3, Min: 0, Max: 6, Step: 1, Needs: []string{"retaliation"}},
		{Name: "lastStandBelow", Default: 30, Min: 0, Max: 90, Step: 10, Needs: []string{"lastStand"}, Kind: KnobThreshold},
		{Name: "shieldWallBelow", Default: 30, Min: 0, Max: 60, Step: 20, Needs: []string{"shieldWall"}, Kind: KnobThreshold},
	}
}

// Gore Drinker's buff, the three hits that heal.
const warriorGoreDrinkerAura = 1323969

const (
	warriorStanceBattle = iota
	warriorStanceDefensive
	warriorStanceBerserker
)

var warriorStanceIDs = []int32{2457, 71, 2458}

func (t WarriorLeveling) Build(k Knobs) *proto.APLRotation {
	knows := func(a warriorAbility) bool { return core.HighestRankAt(t.Level, a.levels) > 0 }
	top := func(a warriorAbility) *proto.ActionID { return topRankID(t.Level, a.levels, []int32{a.top}) }
	tagged := func(a warriorAbility) *proto.ActionID {
		id := top(a)
		if id.GetSpellId() != 0 {
			id.Tag = 1 // Tag 1 is the press that queues it for the next swing.
		}
		return id
	}
	noMobs := Eq(MobsPulled(), Num(0))
	inCombat := Ge(MobsPulled(), Num(1))
	onUs := func(n float64) Value { return Ge(MobsOnUs(), Num(n)) }
	fighting := onUs(1)
	needRest := Lt(CurrentHealthPercent(), Pct(k["restHealth"]))
	inStance := func(s int) Value { return AuraIsActive(SpellID(warriorStanceIDs[s])) }
	// A stance we don't know yet is Battle Stance.
	stance := func(s float64) int {
		switch {
		case s == warriorStanceDefensive && !knows(warriorDefStance):
			return warriorStanceBattle
		case s == warriorStanceBerserker && !knows(warriorBerserker):
			return warriorStanceBattle
		}
		return int(s)
	}

	var items []*proto.APLListItem
	add := func(list ...*proto.APLListItem) { items = append(items, list...) }

	// Between pulls.
	if k["cannibalizeBelow"] > 0 {
		add(Cast(SpellID(20577), And(noMobs, Lt(CurrentHealthPercent(), Pct(k["cannibalizeBelow"]))), "Cannibalize the last corpse."))
	}
	if k["bandageBelow"] > 0 {
		add(Bandage(Lt(CurrentHealthPercent(), Pct(k["bandageBelow"])), "A bandage between pulls."))
	}
	add(Rest(true, false, k["restUntil"], 0, needRest, "Sit down to eat when health is low."))

	// Staying alive comes before pulling more.
	if k["lastStandBelow"] > 0 {
		add(Cast(SpellID(12975), And(fighting, Lt(CurrentHealthPercent(), Pct(k["lastStandBelow"]))), "Last Stand when health runs low."))
	}
	if k["shieldWallBelow"] > 0 && knows(warriorShieldWall) {
		add(Cast(top(warriorShieldWall), And(fighting, Lt(CurrentHealthPercent(), Pct(k["shieldWallBelow"]))), "Shield Wall when health runs low."))
	}
	if k["victoryRushBelow"] > 0 {
		add(Cast(top(warriorVictoryRush), Le(CurrentHealthPercent(), Pct(k["victoryRushBelow"])), "Victory Rush, free after a kill, heals us."))
	}

	// Pulling the mob we found.
	gatherMore := And(Lt(LevelingMobsThisPull(), Num(k["gather"])), Gt(CurrentHealthPercent(), Pct(k["gatherHealth"])))
	chainFit := Gt(CurrentHealthPercent(), Pct(k["chainHealth"]))
	if k["chainMax"] > 0 {
		chainFit = And(chainFit, Lt(MobsPulled(), Num(k["chainMax"])))
	} else {
		chainFit = Eq(Num(0), Num(1))
	}
	near := func(knob string) Value { return Le(PullTargetDistance(), Num(k[knob])) }
	fits := []Value{And(inCombat, chainFit, near("chainDistance")), And(noMobs, Not(needRest))}
	if k["gather"] >= 2 {
		fits = append(fits, And(inCombat, gatherMore, near("gatherDistance")))
	}
	if k["chainDyingBelow"] > 0 {
		fits = append(fits, And(near("chainDistance"), Eq(MobsPulled(), Num(1)), Lt(TargetHealthPercent(), Pct(k["chainDyingBelow"])), Gt(CurrentHealthPercent(), Pct(k["chainDyingHealth"]))))
	}
	pullFit := Or(fits...)

	useCharge := k["charge"] == 1 && knows(warriorCharge)
	main, pack := stance(k["stance"]), stance(k["packStance"])
	if useCharge {
		// Vanguard charges from Defensive Stance, so there we don't swap.
		canCharge := Or(inStance(warriorStanceBattle), And(inStance(warriorStanceDefensive), Ge(TalentPoints("vanguard"), Num(1))))
		if main != warriorStanceBattle || pack != warriorStanceBattle {
			add(Cast(SpellID(warriorStanceIDs[warriorStanceBattle]), And(noMobs, Not(needRest), Gt(PullTargetDistance(), Num(8)), Not(canCharge), SpellIsReady(top(warriorCharge))),
				"Battle Stance for the Charge."))
		}
		add(CastAt(top(warriorCharge), PullTarget, pullFit, "Charge the mob we found."))
		add(WalkToPullTarget(25, And(pullFit, noMobs, Gt(PullTargetDistance(), Num(25)), SpellIsReady(top(warriorCharge))), "Walk into Charge range."))
	}
	if k["intercept"] == 1 && knows(warriorIntercept) {
		add(CastAt(top(warriorIntercept), PullTarget, pullFit, "Intercept the mob we found."))
	}
	add(WalkToPullTarget(0, pullFit, "Walk up to the mob we found until it notices us."))

	// Looking for more.
	if k["gather"] >= 2 {
		add(LookForMob(false, And(inCombat, gatherMore), "Run around to gather the pack."))
	}
	if k["chainMax"] > 0 {
		look := Gt(CurrentHealthPercent(), Pct(max(k["lookAheadHealth"], 1)))
		if k["lookAheadHealth"] == 0 {
			look = Gt(CurrentHealthPercent(), Pct(k["chainHealth"]))
		}
		add(LookForMob(k["chainStand"] == 1, And(inCombat, Lt(MobsPulled(), Num(k["chainMax"])), look), "Look for another mob while we fight."))
	}

	// The stance we fight in.
	useSweeping := k["sweepingStrikes"] == 1
	swapRage := Le(CurrentRage(), Num(k["swapRage"]))
	sweeping := SpellID(12292)
	opening := And(Ge(TalentPoints("sweepingStrikes"), Num(1)), onUs(2), Le(SpellTimeToReady(sweeping), Seconds(2)))
	if !useSweeping {
		opening = Eq(Num(0), Num(1))
	}
	if useSweeping && (main != warriorStanceBattle || pack != warriorStanceBattle) {
		add(Cast(SpellID(warriorStanceIDs[warriorStanceBattle]), And(opening, swapRage, Not(inStance(warriorStanceBattle))), "Battle Stance for Sweeping Strikes."))
	}
	packed := onUs(k["packMobs"])
	if main == pack {
		if main != warriorStanceBattle || useCharge {
			add(Cast(SpellID(warriorStanceIDs[main]), And(fighting, Not(opening), swapRage, Not(inStance(main))), "The stance we fight in."))
		}
	} else {
		add(
			Cast(SpellID(warriorStanceIDs[pack]), And(packed, Not(opening), swapRage, Not(inStance(pack))), "The stance for a pack."),
			Cast(SpellID(warriorStanceIDs[main]), And(fighting, Not(packed), Not(opening), swapRage, Not(inStance(main))), "The stance for few mobs."))
	}

	// Cooldowns.
	if k["bloodrageAbove"] > 0 && k["bloodrageAbove"] < 100 {
		when := fighting
		if k["bloodrageEarly"] == 1 {
			when = inCombat
		}
		add(Cast(top(warriorBloodrage), And(when, Ge(CurrentHealthPercent(), Pct(k["bloodrageAbove"])), Le(CurrentRage(), Num(40))), "Bloodrage for rage while healthy."))
	}
	if k["goreFeedBelow"] > 0 {
		low := And(Ge(TalentPoints("goreDrinker"), Num(1)), fighting, Lt(CurrentHealthPercent(), Pct(k["goreFeedBelow"])), Not(AuraIsActive(SpellID(warriorGoreDrinkerAura))))
		// Both cost health, so we stop short of dying to them.
		safe := And(low, Gt(CurrentHealthPercent(), Pct(30)))
		add(Cast(SpellID(1329555), safe, "Reckless Swings for Gore Drinker, without the talent ignore the warning on this line."))
		if knows(warriorBloodrage) {
			add(Cast(top(warriorBloodrage), safe, "Bloodrage for Gore Drinker."))
		}
		if k["goreDance"] == 1 && knows(warriorBerserkerRage) {
			ready := SpellIsReady(top(warriorBerserkerRage))
			add(Cast(SpellID(warriorStanceIDs[warriorStanceBerserker]), And(low, ready, swapRage, Not(inStance(warriorStanceBerserker))), "Berserker Stance for Berserker Rage."),
				Cast(top(warriorBerserkerRage), low, "Berserker Rage for Gore Drinker."))
		}
	}
	if k["retaliationMobs"] > 0 && knows(warriorRetaliation) {
		n := onUs(k["retaliationMobs"])
		ready := SpellIsReady(top(warriorRetaliation))
		add(Cast(SpellID(warriorStanceIDs[warriorStanceBattle]), And(n, ready, Not(inStance(warriorStanceBattle))), "Battle Stance for Retaliation."),
			Cast(top(warriorRetaliation), n, "Retaliation against a pack."))
	}
	if k["sweepingStrikes"] == 1 {
		add(Cast(sweeping, onUs(2), "Sweeping Strikes with two mobs on us."))
	}
	if k["deathWish"] > 0 {
		add(Cast(SpellID(12328), And(onUs(k["deathWish"]), Gt(CurrentHealthPercent(), Pct(40))), "Death Wish."))
	}
	if k["recklessness"] == 1 {
		add(Cast(top(warriorRecklessness), fighting, "Recklessness."))
	}
	if k["berserkerRage"] == 1 {
		add(Cast(top(warriorBerserkerRage), fighting, "Berserker Rage."))
	}
	if k["concussionBlowMobs"] > 0 {
		add(Cast(SpellID(12809), And(onUs(k["concussionBlowMobs"]), Gt(TargetHealthPercent(), Pct(20))), "Concussion Blow to stun our target."))
	}
	if k["shieldBlockMobs"] > 0 && knows(warriorShieldBlock) {
		add(Cast(top(warriorShieldBlock), And(onUs(k["shieldBlockMobs"]), Not(AuraIsActive(top(warriorShieldBlock)))), "Shield Block."))
	}

	// Shouts and debuffs.
	if shout := top(warriorBattleShout); k["battleShout"] == 1 {
		add(Cast(shout, And(inCombat, Not(AuraIsActive(shout))), "Battle Shout when it's down."))
	}
	if k["demoShoutMobs"] > 0 && knows(warriorDemoShout) {
		add(Cast(top(warriorDemoShout), And(onUs(k["demoShoutMobs"]), Not(TargetAuraIsActive(top(warriorDemoShout))), Gt(TargetHealthPercent(), Pct(30))), "Demoralizing Shout."))
	}
	if k["thunderClap"] > 0 && knows(warriorThunderClap) {
		cond := onUs(k["thunderClapMobs"])
		if k["thunderClap"] == 1 {
			cond = And(cond, Not(TargetAuraIsActive(top(warriorThunderClap))))
		}
		add(Cast(top(warriorThunderClap), cond, "Thunder Clap."))
	}

	// The strikes.
	if k["execute"] == 1 {
		add(Cast(top(warriorExecute), And(Le(MobsOnUs(), Num(k["executeMaxMobs"])), Lt(TargetHealthPercent(), Pct(20))), "Execute."))
	}
	if k["revenge"] == 1 {
		add(Cast(top(warriorRevenge), nil, "Revenge after a dodge, parry or block."))
	}
	if k["overpower"] == 1 {
		add(Cast(top(warriorOverpower), nil, "Overpower after a dodge."))
	}
	if k["whirlwindMobs"] > 0 && knows(warriorWhirlwind) {
		add(Cast(top(warriorWhirlwind), onUs(k["whirlwindMobs"]), "Whirlwind."))
	}
	if k["strike"] == 1 {
		add(Cast(SpellID(21553), fighting, "Mortal Strike, without the talent ignore the warning on this line."),
			Cast(SpellID(23894), fighting, "Bloodthirst, without the talent ignore the warning on this line."),
			Cast(SpellID(23925), fighting, "Shield Slam, without the talent ignore the warning on this line."))
	}
	if k["spearingStrike"] == 1 {
		add(Cast(SpellID(1310222), fighting, "Spearing Strike, without the talent ignore the warning on this line."))
	}
	if k["rend"] > 0 && knows(warriorRend) {
		rend := top(warriorRend)
		above := Gt(TargetHealthPercent(), Pct(k["rendAbove"]))
		add(Cast(rend, And(fighting, Not(DotIsActive(rend)), above), "Rend on a target with health left for it."))
		if k["rend"] == 2 {
			add(Action(onUs(2), "Rend on every mob on us.", &proto.APLAction{Action: &proto.APLAction_Multidot{Multidot: &proto.APLActionMultidot{
				SpellId: rend, MaxDots: 60, MaxOverlap: Seconds(0)}}}))
		}
	}
	if k["slam"] == 1 {
		add(Cast(top(warriorSlam), fighting, "Slam."))
	}
	if k["recklessSwingsAbove"] > 0 {
		add(Cast(SpellID(1329555), And(fighting, Gt(CurrentHealthPercent(), Pct(k["recklessSwingsAbove"])), Le(CurrentRage(), Num(40))), "Reckless Swings for rage while healthy."))
	}
	if k["sunderStacks"] > 0 {
		sunder := top(warriorSunder)
		add(Cast(sunder, And(fighting, Gt(TargetHealthPercent(), Pct(50)), Lt(TargetAuraNumStacks(sunder), Num(k["sunderStacks"]))), "Sunder Armor on a fresh target."))
	}

	// Rage to spare.
	if k["cleaveRage"] > 0 && knows(warriorCleave) {
		add(Cast(tagged(warriorCleave), And(onUs(2), Ge(CurrentRage(), Num(k["cleaveRage"]))), "Cleave with two mobs on us."))
		add(Cast(tagged(warriorHeroicStrike), And(Eq(MobsOnUs(), Num(1)), Ge(CurrentRage(), Num(k["heroicStrikeRage"]))), "Heroic Strike with rage to spare."))
	} else {
		add(Cast(tagged(warriorHeroicStrike), And(fighting, Ge(CurrentRage(), Num(k["heroicStrikeRage"]))), "Heroic Strike with rage to spare."))
	}
	return Rotation(items...)
}

// Abilities are what the rotation can use at the level (LevelingTemplate).
func (t WarriorLeveling) Abilities(hasTalent func(name string) bool) []string {
	var out []string
	for _, a := range warriorAbilities {
		if core.HighestRankAt(t.Level, a.levels) > 0 {
			out = append(out, a.name)
		}
	}
	for _, talent := range warriorTalentAbilities {
		if hasTalent(talent) {
			out = append(out, talent)
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
