#!/usr/bin/python

# How dense the mobs we grind are, for the leveling session's camp options
# (docs/leveling_sim.md, Big pulls).
#
# It reads the same files as export_normal_mobs.py (the cmangos classic-db dump and the
# client's FactionTemplate table), and for each level:
#
#   - takes every spawn of a normal melee mob of level ours -3 to +1 (casters, which are mage
#     class mobs, stay out like in the mob table),
#   - ranks them by how many others are within 100 yards, and picks the spots at the 90th, 75th
#     and 50th percentile,
#   - walks from each spot to the nearest mob we haven't pulled yet, again and again, and prints
#     how far each one was (d1 to d12), how many are within 60 and 100 yards, and how many of
#     the mobs there are casters or shooters, flee, or stand within 10 yards of a friend,
#   - plays a session of pulls of N there: run to the nearest mob that's up at 7 yd/s, N times,
#     then fight for T sec. Killed mobs respawn after their respawn time. Prints the average
#     yards we run for a mob.
#
#   python3 tools/leveling_density.py <ClassicDB.sql.gz> <FactionTemplate.csv> [--json]

import bisect
import csv
import gzip
import json
import math
import random
import re
import statistics
import sys

import export_normal_mobs as base

LEVELS = [5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60]
PERCENTILES = [90, 75, 50]
# Spots sampled around each percentile.
SAMPLES = 60
RUN = 7.0
CELL = 100.0

base.TABLES = {"creature_template", "creature", "pool_creature", "pool_template", "game_event_creature",
               "spawn_group", "spawn_group_spawn", "creature_spawn_entry", "creature_spell_list",
               "creature_spell_list_entry", "creature_ai_scripts", "creature_template_spells"}

SHOOTS = re.compile(r"\b(bolt|shoot|throw|fireball|frostbolt|smite|wrath|multi-shot|spit|volley|lightning|arcane missiles|starfire|holy fire|pyroblast|gouge bolt)\b", re.I)


def load(dump, faction_csv):
    factions = {}
    with open(faction_csv) as f:
        for r in csv.DictReader(f):
            factions[r["ID"]] = (int(r["FactionGroup"]), int(r["FriendGroup"]), int(r.get("EnemyGroup", 0)), r["Faction"])
    db = base.read_dump(dump)
    shooters = {r["Id"] for r in db["creature_spell_list"] if SHOOTS.search(r["Comments"])}
    shooters |= {r["Id"] for r in db["creature_spell_list_entry"] if r["ChanceRangedAttack"] not in ("0", "")}
    flees = set()
    for r in db["creature_ai_scripts"]:
        if any(r.get(f"action{i}_type") == "25" for i in (1, 2, 3)):
            flees.add(r["creature_id"])
    mobs = {}
    for r in db["creature_template"]:
        group, friends, enemies, faction = factions.get(r["Faction"], (7, 0, 0, ""))
        if (r["Rank"] != "0" or int(r["CreatureType"]) not in base.FIGHT_TYPES or int(r["NpcFlags"]) != 0
                or int(r["UnitFlags"]) & base.SKIP_UNIT_FLAGS or group & base.PLAYER_GROUPS
                or friends & base.PLAYER_GROUPS or not 1 <= int(r["MinLevel"]) <= int(r["MaxLevel"]) <= 60):
            continue
        mobs[r["Entry"]] = {
            "lo": int(r["MinLevel"]), "hi": int(r["MaxLevel"]), "caster": r["UnitClass"] == "8",
            "shooter": r["UnitClass"] == "8" or r["SpellList"] in shooters and r["SpellList"] != "0",
            "flees": r["Entry"] in flees, "humanoid": r["CreatureType"] == "7", "faction": faction,
            "hostile": bool(enemies & base.PLAYER_GROUPS),
        }
    limits = {r["entry"]: int(r["max_limit"]) for r in db["pool_template"]}
    pools, share = {}, {}
    for r in db["pool_creature"]:
        pools.setdefault(r["pool_entry"], []).append(r["guid"])
    for entry, guids in pools.items():
        for g in guids:
            share[g] = min(1.0, limits.get(entry, 1) / len(guids))
    groups, group_of = {r["Id"]: int(r["MaxCount"]) for r in db["spawn_group"]}, {}
    for r in db["spawn_group_spawn"]:
        group_of.setdefault(r["Id"], []).append(r["Guid"])
    for gid, guids in group_of.items():
        if groups.get(gid, 0) > 0:
            for g in guids:
                share[g] = min(1.0, groups[gid] / len(guids))
    events = {r["guid"] for r in db["game_event_creature"]}
    entries = {}
    for r in db["creature_spawn_entry"]:
        entries.setdefault(r["guid"], []).append(r["entry"])
    rng = random.Random(1)
    spawns = []
    for r in db["creature"]:
        if r["map"] not in ("0", "1") or r["guid"] in events:
            continue
        # Of a pool only a share is up at once.
        if rng.random() >= share.get(r["guid"], 1.0):
            continue
        entry = r["id"] if r["id"] != "0" else rng.choice(entries.get(r["guid"], ["0"]))
        m = mobs.get(entry)
        if m is None:
            continue
        resp = (int(r["spawntimesecsmin"]) + int(r["spawntimesecsmax"])) / 2
        spawns.append((r["map"], float(r["position_x"]), float(r["position_y"]), float(r["position_z"]),
                       rng.randint(m["lo"], m["hi"]), resp, m))
    return spawns


class Field:
    def __init__(self, spawns):
        self.s = spawns
        self.cells = {}
        for i, s in enumerate(spawns):
            self.cells.setdefault((s[0], int(s[1] // CELL), int(s[2] // CELL)), []).append(i)

    def near(self, m, x, y, r):
        n = int(r // CELL) + 1
        cx, cy = int(x // CELL), int(y // CELL)
        for dx in range(-n, n + 1):
            for dy in range(-n, n + 1):
                yield from self.cells.get((m, cx + dx, cy + dy), ())

    def within(self, i, r):
        s = self.s[i]
        return [j for j in self.near(s[0], s[1], s[2], r) if j != i and dist(self.s[j], s) <= r]

    def nearest(self, m, p, skip, horizon=600):
        best, bd = None, None
        for r in (60, 150, 300, horizon):
            for j in self.near(m, p[0], p[1], r):
                if j in skip:
                    continue
                d = math.dist(self.s[j][1:4], p)
                if bd is None or d < bd:
                    best, bd = j, d
            if bd is not None and bd <= r:
                return best, bd
        return best, bd


def dist(a, b):
    return math.dist(a[1:4], b[1:4])


def walk(field, start, k):
    """The distance to each next mob of the greedy path."""
    p, skip, out = field.s[start][1:4], {start}, [0.0]
    for _ in range(k - 1):
        j, d = field.nearest(field.s[start][0], p, skip)
        if j is None:
            break
        out.append(d)
        skip.add(j)
        p = field.s[j][1:4]
    return out


def session(field, start, n, t, length=1200):
    """Yards run for a mob (after the first), and mobs killed, in a session of n mob pulls that
    each fight for t sec."""
    m, p, now, dead = field.s[start][0], field.s[start][1:4], 0.0, {}
    yards, kills = 0.0, 0
    while now < length:
        for j in [j for j, at in dead.items() if at <= now]:
            del dead[j]
        pull = []
        for _ in range(n):
            j, d = field.nearest(m, p, dead.keys() | set(pull))
            if j is None:
                break
            yards += d
            now += d / RUN
            p = field.s[j][1:4]
            pull.append(j)
        if not pull:
            now += 10
            continue
        now += t
        for j in pull:
            dead[j] = now + field.s[j][5]
        kills += len(pull)
    return yards / max(kills, 1), kills


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    spawns = load(args[0], args[1])
    out = {}
    for level in LEVELS:
        near = [s for s in spawns if level - 3 <= s[4] <= level + 1]
        everyone = Field(near)
        field = Field([s for s in near if not s[6]["caster"]])
        counts = sorted((len(field.within(i, 100)), i) for i in range(len(field.s)))
        keys = [c for c, _ in counts]
        resp = statistics.median(s[5] for s in field.s)
        out[level] = {"spawns": len(field.s), "respawn": resp}
        for pct in PERCENTILES:
            want = keys[min(len(keys) - 1, int(len(keys) * pct / 100))]
            at = bisect.bisect_left(keys, want)
            pool = sorted(counts, key=lambda c: abs(c[0] - want))[:SAMPLES * 4]
            spots = [i for _, i in random.Random(level).sample(pool, min(SAMPLES, len(pool)))]
            med = statistics.mean
            walks = [walk(field, i, 12) for i in spots]
            row = {"in100": med(len(field.within(i, 100)) for i in spots),
                   "in60": med(len(field.within(i, 60)) for i in spots),
                   "d": [round(med(w[k] for w in walks if len(w) > k), 1) for k in range(1, 12)]}
            local = [s for i in spots for s in [everyone.s[j] for j in everyone.near(*field.s[i][:3], 100)]
                     if dist(s, field.s[i]) <= 100]
            row["casters"] = med(s[6]["caster"] for s in local)
            row["shooters"] = med(s[6]["shooter"] for s in local)
            row["flees"] = med(s[6]["flees"] for s in local)
            row["hostile"] = med(s[6]["hostile"] for s in local)
            row["friends10"] = med(sum(field.s[j][6]["faction"] == field.s[i][6]["faction"] for j in field.within(i, 10)) for i in spots)
            row["session"] = {}
            for n, t in ((1, 10), (1, 20), (3, 30), (5, 30), (5, 60), (10, 45), (10, 60), (10, 90)):
                res = [session(field, i, n, t) for i in spots[:30]]
                row["session"][f"{n}x{t}"] = (round(med(r[0] for r in res), 1), round(med(r[1] for r in res), 1))
            out[level][pct] = row
    if "--json" in sys.argv:
        json.dump(out, sys.stdout)
        return
    for level, o in out.items():
        print(f"level {level}: {o['spawns']} spawns, respawn {o['respawn']:.0f} s")
        for pct in PERCENTILES:
            r = o[pct]
            print(f"  p{pct}: within 60 yd {r['in60']:.1f}, 100 yd {r['in100']:.1f}, casters {r['casters']:.0%}, shooters {r['shooters']:.0%}, "
                  f"flee {r['flees']:.0%}, hostile {r['hostile']:.0%}, friends in 10 yd {r['friends10']:.2f}")
            print("    d1..d11 " + " ".join(f"{d:.0f}" for d in r["d"]))
            print("    yd/mob, kills " + "  ".join(f"{k}: {v[0]:.0f}, {v[1]:.0f}" for k, v in r["session"].items()))


if __name__ == "__main__":
    main()
