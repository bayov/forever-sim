#!/usr/bin/python

# Writes assets/db_inputs/mobs/normal_mobs.csv: every normal (not elite or rare) mob we can
# grind in the open world of the 1.12 world database, with its spawn count, for
# tools/gen_mob_stats.py.
#
# It reads the cmangos classic-db dump (Full_DB/ClassicDB_1_12_1_*.sql.gz from
# https://github.com/cmangos/classic-db) and the client's FactionTemplate table
# (https://wago.tools/db2/FactionTemplate/csv?build=1.15.7.61582):
#
#   python3 tools/export_normal_mobs.py <ClassicDB_1_12_1_z2815.sql.gz> <FactionTemplate.csv>
#
# A mob is kept when it:
#   - has spawns in Eastern Kingdoms or Kalimdor (map 0 or 1), so no dungeon mobs,
#   - is rank 0 (normal), levels 1 to 60,
#   - is a creature type we fight (no critters, totems or pets),
#   - has no NPC flags (vendors, quest givers, trainers),
#   - can be attacked (no non-attackable, not-selectable or immune-to-player flags),
#   - belongs to no player faction and is friendly to none, so hostile and neutral mobs
#     stay in and Alliance and Horde NPCs go.

import csv
import gzip
import re
import sys

OUT = "assets/db_inputs/mobs/normal_mobs.csv"

TABLES = {"creature_template", "creature"}
# Beast, dragonkin, demon, elemental, giant, undead, humanoid, mechanical and "not specified".
FIGHT_TYPES = {1, 2, 3, 4, 5, 6, 7, 9, 10}
# UNIT_FLAG_NON_ATTACKABLE, UNIT_FLAG_IMMUNE_TO_PLAYER, UNIT_FLAG_IMMUNE_TO_NPC and
# UNIT_FLAG_NOT_SELECTABLE.
SKIP_UNIT_FLAGS = 0x2 | 0x100 | 0x200 | 0x02000000
# FactionGroup and FriendGroup bits for players, the Alliance and the Horde.
PLAYER_GROUPS = 0x7


def parse_values(s):
    """Splits the VALUES part of an INSERT into rows of strings."""
    rows, i, n = [], 0, len(s)
    while i < n:
        if s[i] != "(":
            i += 1
            continue
        i += 1
        vals, buf, quoted = [], [], False
        while True:
            c = s[i]
            if quoted:
                if c == "\\":
                    buf.append(s[i + 1])
                    i += 2
                    continue
                if c == "'":
                    if i + 1 < n and s[i + 1] == "'":
                        buf.append("'")
                        i += 2
                        continue
                    quoted = False
                    i += 1
                    continue
                buf.append(c)
                i += 1
                continue
            if c == "'":
                quoted = True
            elif c == ",":
                vals.append("".join(buf))
                buf = []
            elif c == ")":
                vals.append("".join(buf))
                i += 1
                break
            else:
                buf.append(c)
            i += 1
        rows.append(vals)
    return rows


def read_dump(path):
    cols, rows, cur = {}, {t: [] for t in TABLES}, None
    with gzip.open(path, "rt", encoding="utf-8", errors="replace") as f:
        for line in f:
            m = re.match(r"CREATE TABLE `(\w+)`", line)
            if m:
                cur = m.group(1) if m.group(1) in TABLES else None
                if cur:
                    cols[cur] = []
                continue
            if cur and line.startswith("  `"):
                cols[cur].append(line.split("`")[1])
                continue
            m = re.match(r"INSERT INTO `(\w+)` (?:\([^)]*\) )?VALUES\s*(.*)", line, re.S)
            if m and m.group(1) in TABLES:
                rows[m.group(1)].extend(parse_values(m.group(2)))
    return {t: [dict(zip(cols[t], r)) for r in rows[t]] for t in TABLES}


def main():
    dump, faction_csv = sys.argv[1], sys.argv[2]
    factions = {}
    with open(faction_csv) as f:
        for r in csv.DictReader(f):
            factions[r["ID"]] = (int(r["FactionGroup"]), int(r["FriendGroup"]))

    db = read_dump(dump)
    spawns = {}
    for r in db["creature"]:
        if r["map"] in ("0", "1"):
            spawns[r["id"]] = spawns.get(r["id"], 0) + 1

    out = []
    for r in db["creature_template"]:
        n = spawns.get(r["Entry"], 0)
        group, friends = factions.get(r["Faction"], (PLAYER_GROUPS, 0))
        if (n == 0 or r["Rank"] != "0" or int(r["CreatureType"]) not in FIGHT_TYPES
                or int(r["NpcFlags"]) != 0 or int(r["UnitFlags"]) & SKIP_UNIT_FLAGS
                or group & PLAYER_GROUPS or friends & PLAYER_GROUPS
                or not 1 <= int(r["MinLevel"]) <= int(r["MaxLevel"]) <= 60):
            continue
        out.append([r["Entry"], r["Name"], r["MinLevel"], r["MaxLevel"], r["CreatureType"], r["UnitClass"],
                    r["MinLevelHealth"], r["MaxLevelHealth"], r["Armor"], r["MinMeleeDmg"], r["MaxMeleeDmg"],
                    r["MeleeAttackPower"], r["MeleeBaseAttackTime"], r["SpeedRun"], n])

    out.sort(key=lambda row: int(row[0]))
    with open(OUT, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["entry", "name", "min_level", "max_level", "creature_type", "unit_class", "min_level_health",
                    "max_level_health", "armor", "min_melee_dmg", "max_melee_dmg", "melee_attack_power",
                    "melee_base_attack_time", "speed_run", "spawns"])
        w.writerows(out)
    print(f"{len(out)} mobs written to {OUT}")


if __name__ == "__main__":
    main()
