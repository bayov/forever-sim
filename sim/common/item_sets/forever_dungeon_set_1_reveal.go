package item_sets

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// The Dungeon Set 1 sets Forever revealed on 2026-11-16 and 17 (PR #134). They carry new set IDs and
// Classic Dungeon Set 2 names, and their bonuses are the client's ItemSetSpell rows (1.60.1.71037).
// Each goes +8 all resistances at 2 pieces, 8 mana per 5 sec or 15 health per 5 sec at 3, the old
// Dungeon Set 1 5-piece proc at 4, a defensive proc when struck at 5 and a stat row at 6. The 5-piece
// rows (escape a root or snare, freeze, silence, absorb, fear or disarm when struck, 60s cooldown)
// need an attacker that a boss fight does not model, so they are left out. The pieces exist only
// under Forever, so the sets have no Classic bonuses. The Five Thunders (set 1679) is in sim/shaman.

// 14049: +40 attack power.
func foreverDungeonAttackPower40(agent core.Agent) {
	agent.GetCharacter().AddStats(stats.Stats{stats.AttackPower: 40, stats.RangedAttackPower: 40})
}

// 14047: increases damage and healing done by magical spells and effects by up to 23.
func foreverDungeonSpellPower23(agent core.Agent) {
	agent.GetCharacter().AddStat(stats.SpellPower, 23)
}

// 21347 (Vitality): restores 15 health per 5 sec. The sim has no health regeneration stat, so it
// is left out.
func foreverDungeonVitality(_ core.Agent) {}

// https://www.wowhead.com/forever/item-set=1667/feralheart-raiment
//
// 4 pieces (450608) is Wildheart's 5-piece row: we model the spell cast's 200 mana and, since the
// November 23 hotfix, a Cat Form swing's 4 Energy a sec for 5 sec (foreverDruidEnergizeTrigger). 6 pieces has two rows, 40 attack power and 23 spell power.
var ItemSetFeralheartRaimentForever = core.NewItemSet(core.ItemSet{
	ID:   1667,
	Name: "Feralheart Raiment",
	Bonuses: core.RulesetSetBonuses(nil, map[int32]core.ApplyEffect{
		2: foreverDungeonResistances,
		3: foreverDungeonManaRegen,
		4: foreverDruidEnergizeTrigger(450608, "Feralheart Raiment - 4pc"),
		6: func(agent core.Agent) {
			foreverDungeonAttackPower40(agent)
			foreverDungeonSpellPower23(agent)
		},
	}),
})

// https://www.wowhead.com/forever/item-set=1669/beastmaster-armor
var ItemSetBeastmasterArmorForever = core.NewItemSet(core.ItemSet{
	ID:   1669,
	Name: "Beastmaster Armor",
	Bonuses: core.RulesetSetBonuses(nil, map[int32]core.ApplyEffect{
		2: foreverDungeonResistances,
		3: foreverDungeonManaRegen,
		4: foreverAutoAttackMana("Beastmaster Armor - 4pc"),
		6: foreverDungeonAttackPower40,
	}),
})

// https://www.wowhead.com/forever/item-set=1671/sorcerers-regalia
var ItemSetSorcerersRegaliaForever = core.NewItemSet(core.ItemSet{
	ID:   1671,
	Name: "Sorcerer's Regalia",
	Bonuses: core.RulesetSetBonuses(nil, map[int32]core.ApplyEffect{
		2: foreverDungeonResistances,
		3: foreverDungeonManaRegen,
		4: foreverSpellCastMana(450527, 450554, "Sorcerer's Regalia - 4pc"),
		6: foreverDungeonSpellPower23,
	}),
})

// https://www.wowhead.com/forever/item-set=1673/soulforge-armor
//
// Both the Retribution pieces (226993-227000) and the tank pieces (226985-226992) are in it.
var ItemSetSoulforgeArmorForever = core.NewItemSet(core.ItemSet{
	ID:   1673,
	Name: "Soulforge Armor",
	Bonuses: core.RulesetSetBonuses(nil, map[int32]core.ApplyEffect{
		2: foreverDungeonResistances,
		3: foreverDungeonManaRegen,
		4: foreverSpellPowerProc(450625, 27499, "Crusader's Wrath"),
		6: foreverDungeonSpellPower23,
	}),
})

// https://www.wowhead.com/forever/item-set=1675/vestments-of-the-virtuous
var ItemSetVestmentsOfTheVirtuousForever = core.NewItemSet(core.ItemSet{
	ID:   1675,
	Name: "Vestments of the Virtuous",
	Bonuses: core.RulesetSetBonuses(nil, map[int32]core.ApplyEffect{
		2: foreverDungeonResistances,
		3: foreverDungeonManaRegen,
		4: foreverSpellCastMana(450576, 450554, "Vestments of the Virtuous - 4pc"),
		6: foreverDungeonSpellPower23,
	}),
})

// https://www.wowhead.com/forever/item-set=1676/darkmantle-armor
var ItemSetDarkmantleArmorForever = core.NewItemSet(core.ItemSet{
	ID:   1676,
	Name: "Darkmantle Armor",
	Bonuses: core.RulesetSetBonuses(nil, map[int32]core.ApplyEffect{
		2: foreverDungeonResistances,
		3: foreverDungeonVitality,
		4: foreverRogueArmorEnergize("Darkmantle Armor - 4pc"),
		6: foreverDungeonAttackPower40,
	}),
})

// https://www.wowhead.com/forever/item-set=1681/deathmist-raiment
//
// 4 pieces (450585): a 5% chance on a spell cast to heal the wearer for 300 (Dark Reward, 450583),
// which changes no damage, so it is left out.
var ItemSetDeathmistRaimentForever = core.NewItemSet(core.ItemSet{
	ID:   1681,
	Name: "Deathmist Raiment",
	Bonuses: core.RulesetSetBonuses(nil, map[int32]core.ApplyEffect{
		2: foreverDungeonResistances,
		3: foreverDungeonManaRegen,
		6: foreverDungeonSpellPower23,
	}),
})

// https://www.wowhead.com/forever/item-set=1778/battlegear-of-heroism
var ItemSetBattlegearOfHeroismForever = core.NewItemSet(core.ItemSet{
	ID:   1778,
	Name: "Battlegear of Heroism",
	Bonuses: core.RulesetSetBonuses(nil, map[int32]core.ApplyEffect{
		2: foreverDungeonResistances,
		3: foreverDungeonVitality,
		4: foreverWarriorsResolve("Battlegear of Heroism - 4pc"),
		// 14492: +20 Strength.
		6: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.Strength, 20)
		},
	}),
})
