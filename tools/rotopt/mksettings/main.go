// mksettings writes the RaidSimRequest JSON the rotation optimizer sims against, built
// from the UI's shipped pieces instead of a hand export: a gear set and a rotation from
// the spec's ui folder, a talent string, and the level 60 raid the enhancement presets
// document (both blessings, Judgement of Wisdom, every debuff, the Phase 1 consumes).
// The rogue gets the raid's Strength of Earth and Grace of Air on top. The target is the
// UI's Level 60 boss (level 63, 3731 armor, Humanoid). At -level 20 it is instead the UI's
// Level 20 build: the level 22 boss soloed with no buffs, Thistle Tea and Instant Poison
// for the rogue, Rockbiter and the boss on the shaman so Lightning Shield fires.
//
//	go run --tags=with_db ./tools/rotopt/mksettings -spec rogue -gear combat_sinister_strike_p2_bis \
//	    -apl combat_sinister_strike -talents 00530310501-32003311201515231 -race Human -out ss60.json
//	go run --tags=with_db ./tools/rotopt/mksettings -spec enh -gear phase_2 -apl optimized \
//	    -talents 05033305-053030031005112251 -race Orc -out enh60.json
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	spec := flag.String("spec", "rogue", "rogue or enh")
	gear := flag.String("gear", "", "gear set name under ui/<spec>/gear_sets")
	apl := flag.String("apl", "", "rotation name under ui/<spec>/apls")
	talents := flag.String("talents", "", "talent string")
	race := flag.String("race", "Human", "race name as in the proto enum, without the Race prefix")
	level := flag.Int("level", 60, "player level, 60 or 20")
	bonusTalents := flag.Int("bonus-talents", 0, "extra talent points beyond the level's, 5 for the Forever beta at level 20")
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

	var uiDir string
	var consumes *proto.Consumes
	var withSpec func(*proto.Player)
	switch *spec {
	case "rogue":
		uiDir = "ui/rogue"
		// The rogue is in a melee group with a shaman.
		raidBuffs.StrengthOfEarthTotem = improved
		raidBuffs.GraceOfAirTotem = improved
		consumes = &proto.Consumes{
			AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
			AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
			DefaultConjured:   proto.Conjured_ConjuredRogueThistleTea,
			DragonBreathChili: true,
			Flask:             proto.Flask_FlaskOfSupremePower,
			Food:              proto.Food_FoodGrilledSquid,
			MainHandImbue:     proto.WeaponImbue_InstantPoison,
			OffHandImbue:      proto.WeaponImbue_DeadlyPoison,
			SpellPowerBuff:    proto.SpellPowerBuff_GreaterArcaneElixir,
			StrengthBuff:      proto.StrengthBuff_JujuPower,
			ZanzaBuff:         proto.ZanzaBuff_GroundScorpokAssay,
			SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
		}
		withSpec = func(player *proto.Player) {
			player.Class = proto.Class_ClassRogue
			player.Spec = &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.RogueOptions{}}}
		}
	case "enh":
		uiDir = "ui/enhancement_shaman"
		consumes = &proto.Consumes{
			AgilityElixir:     proto.AgilityElixir_ElixirOfTheMongoose,
			AttackPowerBuff:   proto.AttackPowerBuff_JujuMight,
			DefaultPotion:     proto.Potions_MajorManaPotion,
			DefaultConjured:   proto.Conjured_ConjuredDemonicRune,
			DragonBreathChili: true,
			FirePowerBuff:     proto.FirePowerBuff_ElixirOfFirepower,
			Flask:             proto.Flask_FlaskOfSupremePower,
			Food:              proto.Food_FoodBlessSunfruit,
			MainHandImbue:     proto.WeaponImbue_WindfuryWeapon,
			ManaRegenElixir:   proto.ManaRegenElixir_MagebloodPotion,
			OffHandImbue:      proto.WeaponImbue_WindfuryWeapon,
			SapperExplosive:   proto.SapperExplosive_SapperGoblinSapper,
			SpellPowerBuff:    proto.SpellPowerBuff_GreaterArcaneElixir,
			StrengthBuff:      proto.StrengthBuff_JujuPower,
			ZanzaBuff:         proto.ZanzaBuff_ROIDS,
		}
		withSpec = func(player *proto.Player) {
			player.Class = proto.Class_ClassShaman
			player.Spec = &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{SyncType: proto.ShamanSyncType_Auto},
			}}
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown spec %q\n", *spec)
		os.Exit(1)
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
	if *level == 20 {
		raidBuffs = &proto.RaidBuffs{}
		individualBuffs = &proto.IndividualBuffs{}
		debuffs = &proto.Debuffs{}
		encounter = &proto.Encounter{
			Duration:             60,
			DurationVariation:    5,
			ExecuteProportion_20: 0.2,
			ExecuteProportion_25: 0.25,
			ExecuteProportion_35: 0.35,
			Targets: []*proto.Target{{
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
			}},
		}
		switch *spec {
		case "rogue":
			consumes = &proto.Consumes{
				DefaultConjured: proto.Conjured_ConjuredRogueThistleTea,
				MainHandImbue:   proto.WeaponImbue_InstantPoison,
				OffHandImbue:    proto.WeaponImbue_InstantPoison,
			}
		case "enh":
			consumes = &proto.Consumes{MainHandImbue: proto.WeaponImbue_RockbiterWeapon}
			tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
		}
	}
	if *duration > 0 {
		encounter.Duration = *duration
	}

	player := &proto.Player{
		Race:               proto.Race(raceValue),
		Level:              int32(*level),
		BonusTalentPoints:  int32(*bonusTalents),
		Equipment:          core.GetGearSet(uiDir+"/gear_sets", *gear).GearSet,
		TalentsString:      *talents,
		Consumes:           consumes,
		Buffs:              individualBuffs,
		Profession1:        proto.Profession_Engineering,
		Profession2:        proto.Profession_Alchemy,
		Rotation:           core.GetAplRotation(uiDir+"/apls", *apl).Rotation,
		DistanceFromTarget: 5,
		ReactionTimeMs:     150,
		ChannelClipDelayMs: 50,
	}
	withSpec(player)

	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, debuffs)
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
