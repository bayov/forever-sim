// mksettings writes the RaidSimRequest JSON the rotation optimizer sims against, built
// from the UI's shipped pieces instead of a hand export: a gear set and a rotation from
// the spec's ui folder, a talent string, and the level 60 raid the enhancement presets
// document (both blessings, Judgement of Wisdom, every debuff, the Phase 1 consumes).
// The rogue gets the raid's Strength of Earth and Grace of Air on top. The target is the
// UI's Level 60 boss (level 63, 3731 armor, Humanoid). Below 60 it is instead the UI's
// Level 20 build: a boss two levels up (see lowLevelTarget) soloed with no raid buffs or
// debuffs, Thistle Tea and Instant Poison for the rogue, Rockbiter and the boss on the
// shaman so Lightning Shield fires, and the boss on the paladin too (Retribution Aura).
// Everyone gets Blessing of Kings, because a group at that level almost always has a
// paladin in it.
//
//	go run --tags=with_db ./tools/rotopt/mksettings -spec rogue -gear combat_sinister_strike_p2_bis \
//	    -apl combat_sinister_strike -talents 00530310501-32003311201515231 -race Human -out ss60.json
//	go run --tags=with_db ./tools/rotopt/mksettings -spec enh -gear phase_2 -apl optimized \
//	    -talents 0505331-055030031005112251 -race Orc -out enh60.json
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

func main() {
	spec := flag.String("spec", "rogue", "one of the specs registered in spec_*.go: "+strings.Join(specNames(), ", "))
	gear := flag.String("gear", "", "gear set name under ui/<spec>/gear_sets")
	apl := flag.String("apl", "", "rotation name under ui/<spec>/apls")
	talents := flag.String("talents", "", "talent string")
	race := flag.String("race", "Human", "race name as in the proto enum, without the Race prefix")
	level := flag.Int("level", 60, "player level, 60, 20 or 30")
	bonusTalents := flag.Int("bonus-talents", 0, "extra talent points beyond the level's, 5 for the Forever beta at level 20")
	blessing := flag.String("blessing", "kings", "the ret paladin's own blessing below level 60: kings, might or wisdom")
	targets := flag.Int("targets", 1, "number of enemies, copies of the level's target")
	duration := flag.Float64("duration", 0, "fight length in seconds, 300 at level 60 and 60 at level 20 when left out")
	out := flag.String("out", "", "file to write, stdout when empty")
	flag.Parse()

	raceValue, ok := proto.Race_value["Race"+*race]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown race %q\n", *race)
		os.Exit(1)
	}

	improved := proto.TristateEffect_TristateEffectImproved
	raidBuffs := &proto.RaidBuffs{
		ArcaneBrilliance:    true,
		BattleShout:         improved,
		BloodPact:           improved,
		DevotionAura:        improved,
		DivineSpirit:        true,
		FireResistanceAura:  true,
		FrostResistanceAura: true,
		GiftOfTheWild:       improved,
		LeaderOfThePack:     true,
		MoonkinAura:         true,
		PowerWordFortitude:  improved,
		RetributionAura:     improved,
		SanctityAura:        true,
		ShadowProtection:    true,
		Thorns:              improved,
		TrueshotAura:        true,
	}
	individualBuffs := &proto.IndividualBuffs{
		BlessingOfKings:  true,
		BlessingOfMight:  improved,
		BlessingOfWisdom: improved,
	}
	debuffs := &proto.Debuffs{
		CurseOfElements:        true,
		CurseOfRecklessness:    true,
		CurseOfShadow:          true,
		CurseOfWeakness:        improved,
		DemoralizingRoar:       improved,
		DemoralizingShout:      improved,
		ExposeArmor:            improved,
		FaerieFire:             true,
		InsectSwarm:            true,
		JudgementOfLight:       true,
		JudgementOfTheCrusader: improved,
		JudgementOfWisdom:      true,
		ScorpidSting:           true,
		SunderArmor:            true,
		ThunderClap:            improved,
	}

	setup, ok := specs[*spec]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown spec %q, want one of %s\n", *spec, strings.Join(specNames(), ", "))
		os.Exit(1)
	}
	if setup.raid != nil {
		setup.raid(raidBuffs, individualBuffs, debuffs)
	}
	consumes := setup.consumes()
	partyBuffs := &proto.PartyBuffs{}
	if setup.partyBuffs != nil {
		partyBuffs = setup.partyBuffs()
	}

	encounter := &proto.Encounter{
		Duration:             300,
		DurationVariation:    20,
		ExecuteProportion_20: 0.2,
		ExecuteProportion_25: 0.25,
		ExecuteProportion_35: 0.35,
		Targets: []*proto.Target{{
			Id:      213336,
			Name:    "Level 60",
			Level:   63,
			MobType: proto.MobType_MobTypeHumanoid,
			Stats: stats.Stats{
				stats.AttackPower: 805,
				stats.Armor:       3731,
				stats.Health:      127393,
			}.ToFloatArray(),
			MinBaseDamage: 3000,
			DamageSpread:  0.3333,
			SwingSpeed:    2,
			ParryHaste:    true,
		}},
	}
	var tanks []*proto.UnitReference
	if *level < 60 {
		raidBuffs = &proto.RaidBuffs{}
		individualBuffs = &proto.IndividualBuffs{BlessingOfKings: true}
		debuffs = &proto.Debuffs{}
		encounter = &proto.Encounter{
			Duration:             60,
			DurationVariation:    5,
			ExecuteProportion_20: 0.2,
			ExecuteProportion_25: 0.25,
			ExecuteProportion_35: 0.35,
			Targets:              []*proto.Target{lowLevelTarget(int32(*level))},
		}
		consumes = &proto.Consumes{}
		if setup.lowLevel != nil {
			low, err := setup.lowLevel(*blessing)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if low.consumes != nil {
				consumes = low.consumes
			}
			if low.buffs != nil {
				individualBuffs = low.buffs
			}
			if low.tankSelf {
				tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
			}
		}
	}
	if *duration > 0 {
		encounter.Duration = *duration
	}
	for len(encounter.Targets) < *targets {
		encounter.Targets = append(encounter.Targets, goproto.Clone(encounter.Targets[0]).(*proto.Target))
	}

	player := &proto.Player{
		Race:               proto.Race(raceValue),
		Level:              int32(*level),
		BonusTalentPoints:  int32(*bonusTalents),
		Equipment:          core.GetGearSet(setup.uiDir+"/gear_sets", *gear).GearSet,
		TalentsString:      *talents,
		Consumes:           consumes,
		Buffs:              individualBuffs,
		Profession1:        proto.Profession_Engineering,
		Profession2:        proto.Profession_Alchemy,
		Rotation:           core.GetAplRotation(setup.uiDir+"/apls", *apl).Rotation,
		DistanceFromTarget: 5,
		ReactionTimeMs:     200,
		ChannelClipDelayMs: 50,
	}
	if setup.professions[0] != proto.Profession_ProfessionUnknown {
		player.Profession1, player.Profession2 = setup.professions[0], setup.professions[1]
	}
	setup.withSpec(player)

	raid := core.SinglePlayerRaidProto(player, partyBuffs, raidBuffs, debuffs)
	raid.Tanks = tanks
	request := &proto.RaidSimRequest{
		Raid:      raid,
		Encounter: encounter,
		SimOptions: &proto.SimOptions{
			Iterations: 1000,
			RandomSeed: 1,
			Ruleset:    proto.Ruleset_RulesetForever,
		},
	}

	data, err := protojson.MarshalOptions{Multiline: true}.Marshal(request)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(1)
	}
	if *out == "" {
		os.Stdout.Write(data)
		return
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
}

// The boss a low level character fights: a dungeon boss two levels above them, with its
// armor from the 1.12 creature data. Mutanus the Devourer (Wailing Caverns) up to level
// 25, and Interrogator Vishas (Scarlet Monastery Graveyard) from level 30. Health only
// sets when the execute phases start, so it is a round number. The melee damage is what
// Lightning Shield, Retribution Aura and the paladin's Deflection see.
func lowLevelTarget(level int32) *proto.Target {
	target := &proto.Target{
		Id:      3654,
		Name:    "Mutanus the Devourer",
		Level:   22,
		MobType: proto.MobType_MobTypeHumanoid,
		Stats: stats.Stats{
			stats.Armor:  922,
			stats.Health: 20000,
		}.ToFloatArray(),
		MinBaseDamage: 40,
		DamageSpread:  0.3333,
		SwingSpeed:    2,
		ParryHaste:    true,
	}
	if level >= 30 {
		target.Id = 3983
		target.Name = "Interrogator Vishas"
		target.Level = 32
		target.Stats = stats.Stats{
			stats.Armor:  1063,
			stats.Health: 30000,
		}.ToFloatArray()
		target.MinBaseDamage = 52
	}
	return target
}
