#!/usr/bin/python

# Generates sim/core/base_stats_levels_auto_gen.go: base attributes, health and mana for
# every race/class at levels 1-60, and the level curve of crit per agility and intellect.
#
# The attribute rows and base health come from the 1.12 world database (player_levelstats
# and player_classlevelstats, exported to CSV from the cmangos classic-db dump). Base mana
# and the crit curves come from the client's gtOCTBaseMPByClass, gtChanceToMeleeCrit and
# gtChanceToSpellCrit tables.
#
#   python3 tools/gen_level_base_stats.py

import csv

DIR = "assets/db_inputs/basestats/"
OUT = "sim/core/base_stats_levels_auto_gen.go"

RACES = {1: "Human", 2: "Orc", 3: "Dwarf", 4: "NightElf", 5: "Undead", 6: "Tauren", 7: "Gnome", 8: "Troll"}
CLASSES = {1: "Warrior", 2: "Paladin", 3: "Hunter", 4: "Rogue", 5: "Priest", 7: "Shaman", 8: "Mage", 9: "Warlock", 11: "Druid"}
# Column of each class in the gt tables.
GT_COLUMNS = {"Warrior": 1, "Paladin": 2, "Hunter": 3, "Rogue": 4, "Priest": 5, "Shaman": 7, "Mage": 8, "Warlock": 9, "Druid": 11}


def read_gt(name):
    rows = {}
    with open(DIR + name) as f:
        for line in f.read().splitlines()[1:]:
            parts = line.split("\t")
            if len(parts) < 12:
                continue
            rows[int(parts[0])] = [float(x) for x in parts[1:12]]
    return rows


levelstats = {}
with open(DIR + "player_levelstats.csv") as f:
    for r in csv.DictReader(f):
        levelstats.setdefault((int(r["race"]), int(r["class"])), {})[int(r["level"])] = (
            int(r["str"]), int(r["agi"]), int(r["sta"]), int(r["int"]), int(r["spi"]))

# cmangos has the level 1 Undead mage's Intellect and Spirit swapped (27 and 21). At every other
# level, and for the other Undead classes, the Undead has 2 less Intellect and more Spirit than
# the Human. The beta's level 1 Undead mage fits the swapped back row.
s, a, st, i, sp = levelstats[(5, 8)][1]
levelstats[(5, 8)][1] = (s, a, st, sp, i)

# Forever's races take Classic's attributes: "All Races now have the correct amount of base
# stats. For example, Tauren Warriors now start with the correct +5 strength relative to Human
# Warriors (was incorrectly +2)" (development notes 2026-11-02, post 9, build 1.60.1.70753). So
# every 1.12 row stands as it is. Until then the beta's races were 20 in every attribute plus a
# small offset (our naked level 1 sheets of 2026-10-10 and wowsims/forever 6296e8546d); that
# shift is gone.
#
# The Forever Orc shaman's own rows at levels 2 to 43.
#
# These are the base values the in-game character sheet shows (UnitStat's base, with no gear,
# buffs or talents counted), which BayovCore saved for every level we played through in the
# live logs of Nov 5-8 2026. Stamina and Spirit match 1.12 at every level. Strength runs a point
# or two above 1.12, and Agility and Intellect a point or two below, and the gap grows slowly
# with level. Levels 2, 4, 6, 7 and 11 happen to round to 1.12's row.
#
# We have no sheet for level 1 or for levels 44 to 60, so those stay on 1.12's rows. Level 60
# doesn't read this table at all (it uses ClassBaseStats in base_stats.go).
FOREVER_ATTRIBUTES = {
    (2, 7): {  # Orc shaman
        2: (25, 17, 24, 19, 26),
        3: (26, 18, 25, 20, 27),
        4: (26, 18, 26, 20, 28),
        5: (27, 18, 26, 21, 29),
        6: (28, 19, 27, 22, 30),
        7: (29, 19, 28, 23, 31),
        8: (30, 20, 29, 23, 31),
        9: (30, 20, 30, 24, 32),
        10: (31, 20, 31, 25, 33),
        11: (32, 21, 32, 26, 34),
        12: (33, 21, 33, 27, 35),
        13: (34, 22, 34, 28, 36),
        14: (35, 22, 35, 29, 37),
        15: (36, 23, 36, 29, 39),
        16: (37, 23, 37, 30, 40),
        17: (37, 23, 38, 31, 41),
        18: (38, 24, 39, 32, 42),
        19: (39, 24, 40, 33, 43),
        20: (40, 25, 41, 34, 44),
        21: (41, 25, 42, 35, 45),
        22: (42, 26, 43, 36, 46),
        23: (43, 26, 44, 37, 47),
        24: (44, 27, 45, 38, 49),
        25: (45, 27, 47, 39, 50),
        26: (46, 28, 48, 40, 51),
        27: (47, 28, 49, 41, 52),
        28: (48, 29, 50, 42, 53),
        29: (50, 29, 51, 43, 55),
        30: (51, 30, 52, 44, 56),
        31: (52, 30, 54, 45, 57),
        32: (53, 31, 55, 46, 59),
        33: (54, 31, 56, 48, 60),
        34: (55, 32, 57, 49, 61),
        35: (56, 32, 59, 50, 63),
        36: (58, 33, 60, 51, 64),
        37: (59, 34, 61, 52, 65),
        38: (60, 34, 63, 53, 67),
        39: (61, 35, 64, 55, 68),
        40: (62, 35, 65, 56, 70),
        41: (64, 36, 67, 57, 71),
        42: (65, 37, 68, 58, 73),
        43: (66, 37, 70, 60, 74),
    },
}
for (race, cls), levels in FOREVER_ATTRIBUTES.items():
    levelstats[(race, cls)].update(levels)

melee_crit = read_gt("chancetomeleecrit.txt")
spell_crit = read_gt("chancetospellcrit.txt")
base_mana = read_gt("octbasempbyclass.txt")

# We take base mana from the client's table rather than cmangos. The cmangos table has wrong
# rows, for example 718 for a level 30 shaman where the client and the Forever beta have 665.
classstats = {}
with open(DIR + "player_classlevelstats.csv") as f:
    for r in csv.DictReader(f):
        cls, level = int(r["class"]), int(r["level"])
        mana = int(r["basemana"])
        if cls in CLASSES and level <= 60:
            mana = round(base_mana[level][GT_COLUMNS[CLASSES[cls]] - 1])
        classstats.setdefault(cls, {})[level] = (int(r["basehp"]), mana)

out = []
out.append("package core\n")
out.append("// **************************************")
out.append("// AUTO GENERATED BY GEN_LEVEL_BASE_STATS.PY")
out.append("// **************************************\n")
out.append('import "github.com/wowsims/classic/sim/core/proto"\n')

out.append("// Strength, Agility, Stamina, Intellect and Spirit at each level, index 0 is level 1.")
out.append("var levelAttributes = map[RaceClass][60][5]int16{")
for (race, cls), levels in sorted(levelstats.items()):
    if race not in RACES or cls not in CLASSES:
        continue
    out.append(f"\t{{proto.Race_Race{RACES[race]}, proto.Class_Class{CLASSES[cls]}}}: {{")
    for start in range(1, 61, 10):
        row = ", ".join("{%d, %d, %d, %d, %d}" % levels[l] for l in range(start, start + 10))
        out.append(f"\t\t{row},")
    out.append("\t},")
out.append("}\n")

out.append("// Base health and mana at each level, index 0 is level 1.")
out.append("var levelHealthMana = map[proto.Class][60][2]int16{")
for cls, levels in sorted(classstats.items()):
    if cls not in CLASSES:
        continue
    out.append(f"\tproto.Class_Class{CLASSES[cls]}: {{")
    for start in range(1, 61, 10):
        row = ", ".join("{%d, %d}" % levels[l] for l in range(start, start + 10))
        out.append(f"\t\t{row},")
    out.append("\t},")
out.append("}\n")


def curve(name, table, comment):
    out.append(comment)
    out.append(f"var {name} = map[proto.Class][60]float64{{")
    for cls, col in GT_COLUMNS.items():
        out.append(f"\tproto.Class_Class{cls}: {{")
        for start in range(1, 61, 10):
            out.append("\t\t" + ", ".join("%.6f" % table[l][col - 1] for l in range(start, start + 10)) + ",")
        out.append("\t},")
    out.append("}\n")


curve("meleeCritPerAgiCurve", melee_crit, "// Melee crit chance per point of Agility at each level (gtChanceToMeleeCrit), index 0 is level 1.")
curve("spellCritPerIntCurve", spell_crit, "// Spell crit chance per point of Intellect at each level (gtChanceToSpellCrit), index 0 is level 1.")

with open(OUT, "w") as f:
    f.write("\n".join(out))
