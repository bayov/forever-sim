"""Write the tank side of the encounter files from Warcraft Logs numbers.

The boss's melee, its weapon damage abilities (Cleave, Mortal Strike, Knock Away), the fight's
length and the healing a raid puts on its main tank, from what `wclcli.py tank -o` mined
(tools/encounters/logs/*.json). The other abilities keep their client spell data and script
timers. docs/tank_sim.md, Encounters, says how the numbers were picked.

    python3 tools/encounters/from_logs.py --site anniversary tools/encounters/logs/anniversary_*.json
    python3 tools/encounters/from_logs.py --site sod tools/encounters/logs/sod_*.json

The template (the bosses, their abilities and abilitiesSource) is the top-level files, whatever
logs they took last: a weapon damage ability with too few logged hits scales with the melee.

A log's swings had a raid's debuffs on the boss (Demoralizing Shout, Curse of Weakness,
Thunder Clap), and the bench puts the requests' own on him again, so the weapon damage and swing
speed are taken back through them (`tankbench -boss-stats`): under the bench's debuffs the sim's
boss swings as the log's did. That assumes the logged raids had about the bench's debuffs.
"""
import argparse, copy, json, os, subprocess, sys, time

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.normpath(os.path.join(HERE, '../..'))
REQUEST = 'tools/tankbench/req/warrior_protection.json'

SITES = {
    'sod': {'label': 'Season of Discovery', 'tier': 'logs: Season of Discovery', 'subdomain': 'sod'},
    # The Anniversary realms' Hardcore realms (Doomhowl, Soulseeker), which share the vanilla
    # site with Classic Era: fresh.warcraftlogs.com's API serves the Classic progression realms.
    'anniversary': {'label': 'the Anniversary Hardcore realms', 'tier': 'logs: Anniversary', 'subdomain': 'vanilla'},
    'era': {'label': 'Classic Era', 'tier': 'logs: Era', 'subdomain': 'vanilla'},
}
DIFFICULTY = {3: 'Normal', 4: 'Heroic', 5: 'Mythic'}

# Each boss: its Warcraft Logs encounter on each site, what its fight length is (the kill's
# length, the main tank's time on him, or his longest stretch on one tank), and its weapon damage
# abilities: the file's spell ID and the IDs the logs show it as.
BOSSES = {
    'magmadar': {'ids': {'sod': 100664, 'anniversary': 50664, 'era': 50664}, 'length': 'fight', 'weapon': {}},
    'golemagg': {'ids': {'sod': 100670, 'anniversary': 50670, 'era': 50670}, 'length': 'fight',
                 # Double Attack's extra swings log as white swings, so they're in the swing rate.
                 'weapon': {18943: 'in swings'}},
    # Elemental Fire procs on his landed hits: about 3 in 4 swings on the Anniversary realms, far
    # fewer in Season of Discovery, so its rate is the logs' (its hits a minute of tanking).
    'ragnaros': {'ids': {'sod': 100672, 'anniversary': 50672, 'era': 50672}, 'length': 'tanking', 'weapon': {},
                 'log_rate': {20564: '20564'}},
    'onyxia': {'ids': {'sod': 151084, 'anniversary': 51084, 'era': 51084}, 'length': 'tanking',
               'weapon': {19983: [19983], 19633: [19633]}},
    'broodlord': {'ids': {'sod': 100612, 'anniversary': 50612, 'era': 50612}, 'length': 'fight',
                  'weapon': {24573: [24573], 18670: [18670], 15284: [15284, 15754]}},
    'vaelastrasz': {'ids': {'sod': 100611, 'anniversary': 50611, 'era': 50611}, 'length': 'stint',
                    'weapon': {19983: [19983]}},
    'chromaggus': {'ids': {'sod': 100616, 'anniversary': 50616, 'era': 50616}, 'length': 'fight', 'weapon': {}},
}
LENGTHS = {'fight': ('lengths', 'the kills\' length'),
           'tanking': ('tankings', "the main tank's time on him a kill"),
           'stint': ('stints', "his longest stretch on one tank a kill")}

# Fewer normal hits than this and a range isn't worth fitting.
MIN_HITS = 20
MIN_ABILITY_HITS = 8


def pctl(xs, p):
    xs = sorted(xs)
    return xs[min(len(xs) - 1, max(0, round(p / 100 * (len(xs) - 1))))] if xs else 0


def median(xs):
    return pctl(xs, 50)


def stretch(p05, p95):
    """The range of an even roll whose 5th and 95th percentiles are these."""
    w = (p95 - p05) / 0.9
    return p05 - 0.05 * w, p95 + 0.05 * w


def boss_stats(path):
    out = subprocess.run(['go', 'run', '-tags=with_db', './tools/tankbench', '-boss-stats', '-encounter', path, REQUEST],
                         cwd=ROOT, capture_output=True, text=True, check=True).stdout
    return json.loads(next(l for l in out.splitlines() if l.startswith('{')))


def fmt(x):
    return f'{x:,.0f}'


def melee_from(m, st):
    """The file's minBaseDamage, damageSpread and swingSpeed from the log's normal hits."""
    # The roll within a kill, around the median over all kills: the spread between raids is
    # their debuffs, which the bench sets itself.
    lo, hi = stretch(m['unmitigated_median'] * m['unmitigated_rel_p05'], m['unmitigated_median'] * m['unmitigated_rel_p95'])
    k = 1 + st['attackPower'] * st['apCoefficient']
    dd = st['damageDone']
    base = (lo / dd - st['flatDamage']) / k
    spread = (hi / dd - st['flatDamage']) / base - k
    # The mean gap between swings while he's on the tank (parry haste in), back through Thunder Clap.
    swings = sum(v for k2, v in m['swings'].items())
    gap = m['tanking_seconds'] / swings
    return round(base, 1), round(spread, 4), round(gap * st['attackSpeed'], 2), lo, hi, gap


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument('--site', choices=SITES, required=True)
    ap.add_argument('--out', help='where the files are (default: by site, tools/encounters for the Anniversary realms, '
                    'whose files score tanks (the user, 2026-11-02), tools/encounters/<site> for the others)')
    ap.add_argument('--template', default=HERE, help='the files whose bosses and abilities are kept (default tools/encounters)')
    ap.add_argument('--boss', action='append', help='only this boss (repeat for more)')
    ap.add_argument('mined', nargs='+', help="wclcli.py tank -o files of this site")
    a = ap.parse_args()
    if not a.out:
        a.out = HERE if a.site == 'anniversary' else os.path.join(HERE, a.site)
    site = SITES[a.site]
    mined = {}
    for path in a.mined:
        data = json.load(open(path))
        if isinstance(data, list):
            continue  # a sample file (the fights read), not what tank mined
        for enc, m in data.items():
            mined[int(enc)] = (m, os.path.relpath(path, ROOT))
    os.makedirs(a.out, exist_ok=True)
    for boss, cfg in BOSSES.items():
        if a.boss and boss not in a.boss:
            continue
        enc_id = cfg['ids'][a.site]
        if enc_id not in mined:
            print(f'{boss}: no log numbers for encounter {enc_id}, left as it is', file=sys.stderr)
            continue
        m, mined_path = mined[enc_id]
        tpl_path = os.path.join(a.template, boss + '.json')
        d = json.load(open(tpl_path))
        st = boss_stats(os.path.relpath(tpl_path, ROOT))
        target = d['encounter']['targets'][0]
        if m['unmitigated_hits'] < MIN_HITS:
            print(f'{boss}: only {m["unmitigated_hits"]} normal hits, melee left as it is', file=sys.stderr)
            continue
        k0 = 1 + st['attackPower'] * st['apCoefficient']
        old_mid = target['minBaseDamage'] * (k0 + target['damageSpread'] / 2)
        base, spread, swing, lo, hi, gap = melee_from(m, st)
        target['minBaseDamage'], target['damageSpread'], target['swingSpeed'] = base, spread, swing
        scale = base * (k0 + spread / 2) / old_mid

        # The weapon damage abilities: the log's hits, or the file's scaled with the melee.
        ab_text = []
        events = d['encounter'].get('raidDamage', [])
        for ev in list(events):
            if ev['spellId'] in cfg.get('log_rate', {}):
                row = m['abilities'].get(cfg['log_rate'][ev['spellId']])
                if row:
                    ev['everySeconds'] = round(60 / row['per_minute'], 2)
                    if ev is events[0] or not ev.get('ticks'):
                        ab_text.append(f'{row["name"]} {ev["spellId"]} {row["per_minute"]:.1f} a minute of tanking')
            ids = cfg['weapon'].get(ev['spellId'])
            if ids is None:
                continue
            if ids == 'in swings':
                events.remove(ev)
                continue
            rows = [m['abilities'][str(i)] for i in ids if str(i) in m['abilities']]
            hits = sum(r['outcomes'].get('hit', 0) for r in rows)
            if rows and hits >= MIN_ABILITY_HITS:
                r = max(rows, key=lambda r: r['outcomes'].get('hit', 0))
                alo, ahi = stretch(r['unmitigated_p05'], r['unmitigated_p95'])
                ev['minDamage'], ev['maxDamage'] = round(alo), round(ahi)
                ab_text.append(f'{r["name"]} {ev["spellId"]} {fmt(alo)} to {fmt(ahi)} from {hits} hits')
            else:
                ev['minDamage'], ev['maxDamage'] = round(ev['minDamage'] * scale), round(ev['maxDamage'] * scale)
                ab_text.append(f'{ev["spellId"]} scaled with the melee ({hits} hits in the logs, too few)')

        # The fight's length.
        key, what = LENGTHS[cfg['length']]
        xs = m[key]
        dur = median(xs)
        var = min((pctl(xs, 90) - pctl(xs, 10)) / 2, dur / 2)
        d['encounter']['duration'] = round(dur)
        d['encounter']['durationVariation'] = round(var)

        # The healing on the main tank.
        h = m['file_healing']
        healers = max(1, round(h['healers']))
        hps_each = m['healers_hps'] / healers
        cast = healers / m['healers_heals_ps']
        d['healing'] = {
            'healers': healers, 'hpsEach': round(hps_each), 'castSeconds': round(cast, 2),
            'notes': (f'From {site["label"]}\'s logs ({m["fights"]} kills): while the boss was on the main tank, '
                      f'{healers} healers had 5% or more of its healing (the median kill), and the healing it took '
                      f'from others was {fmt(m["healers_hps"])} a second with overheal and absorbs '
                      f'({m["hps_effective"] / max(m["hps_total"], 1) * 100:.0f}% of the heals effective), '
                      f'{m["healers_heals_ps"]:.2f} heals and HoT ticks a second, against {fmt(m["dtps"])} damage taken '
                      f'a second. The tank\'s own heals ({m["own_share"] * 100:.0f}% of its healing) are left out. '
                      f'Main tanks: {", ".join(f"{k} {v}" for k, v in sorted(m["tank_classes"].items(), key=lambda x: -x[1]))}.'),
        }

        diff = ', '.join(DIFFICULTY.get(x, str(x)) for x in m['difficulty'])
        dates = [time.strftime('%Y-%m-%d', time.gmtime(t / 1000)) for t in m.get('dates', [])]
        when = f', {min(dates)} to {max(dates)}' if dates else ''
        melee = (f'[{site["tier"]}] Melee {fmt(lo * 1)} to {fmt(hi)} before armor under the bench\'s raid debuffs, '
                 f'a swing every {gap:.2f} s on the tank: Warcraft Logs, {m["fights"]} kills on {site["label"]} '
                 f'({diff}{when}): the median of {m["unmitigated_hits"]} normal hits\' unmitigatedAmount '
                 f'({fmt(m["unmitigated_median"])}), the 5th to 95th percentile of each hit over its own kill\'s median '
                 f'({m["unmitigated_rel_hits"]} hits) stretched to an even roll, and the swings a second while he was on the main tank. '
                 f'Taken back through the debuffs the bench puts on him (attack power {st["baseAttackPower"]:.0f} to '
                 f'{st["attackPower"]:.0f}, {st["flatDamage"]:+.0f} a swing, attack speed {st["attackSpeed"]:.3f}): '
                 f'{fmt(base)} weapon damage, spread {spread:.3f}, a {swing:.2f} s swing. '
                 f'Length: {what}, median {dur:.0f} s (10th to 90th percentile {pctl(xs, 10):.0f} to {pctl(xs, 90):.0f}).')
        if ab_text:
            melee += ' Weapon damage abilities: ' + '; '.join(ab_text) + '.'
        d['source'] = melee + ' Abilities: ' + d['abilitiesSource']
        d['reports'] = m['reports']
        d['warcraftLogs'] = {'site': site['subdomain'], 'encounter': enc_id, 'zone': d['warcraftLogs']['zone'],
                             'difficulty': diff, 'mined': mined_path}
        out = os.path.join(a.out, boss + '.json')
        with open(out, 'w') as f:
            json.dump(d, f, indent=1, ensure_ascii=False)
            f.write('\n')
        print(f'{boss}: melee {fmt(lo)} to {fmt(hi)} every {gap:.2f} s (was weapon {old_mid:,.0f} mid, now {base * (k0 + spread / 2):,.0f}), '
              f'{dur:.0f} ± {var:.0f} s, {healers} healers at {hps_each:.0f} HPS in {cast:.2f} s heals -> {os.path.relpath(out, ROOT)}')


if __name__ == '__main__':
    main()
