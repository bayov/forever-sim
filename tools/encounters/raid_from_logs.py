"""Put the raid side of Warcraft Logs' boss fights into the encounter files (docs/healing_sim.md).

Reads what `tools/warcraftlogs/wclcli.py raid -o` mined (tools/encounters/logs/<site>_<zone>_raid.json)
and, per boss, replaces the file's raid damage on the players who aren't tanks (the `raidDamage`
events with `who` NotTanks) with the logs' abilities, and writes what real raids had: `raidHealing`
(healers, their healing, deaths), `raidSource` and `raidReports`. The tank side (the boss's melee,
the BossTarget and Tanks events, `healing`, `source`, `reports`) is the tank sim's and stays as it is.

    python3 tools/encounters/raid_from_logs.py --site anniversary tools/encounters/logs/anniversary_*_raid.json
    python3 tools/encounters/raid_from_logs.py --site sod tools/encounters/logs/sod_*_raid.json

Run it after from_logs.py, which writes a site's files from tools/encounters' own (so from the
Season of Discovery raid side until this puts each site's own back).
"""
import argparse, json, os, sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(os.path.dirname(HERE))
sys.path.insert(0, HERE)
from from_logs import BOSSES, SITES, fmt, stretch  # noqa: E402

# An ability goes in when it's at least this share of the non-tanks' damage taken and shows up in
# this share of the fights.
MIN_SHARE = 0.02
MIN_FIGHTS = 0.3
# Hitting at least this share of the non-tanks a cast is hitting them all.
ALL = 0.9

SCHOOLS = {1: 'SpellSchoolPhysical', 2: 'SpellSchoolHoly', 4: 'SpellSchoolFire', 8: 'SpellSchoolNature',
           16: 'SpellSchoolFrost', 32: 'SpellSchoolShadow', 64: 'SpellSchoolArcane'}


def school(mask):
    """A school mask's proto school: the lowest bit of a multi-school ability (Shadowflame is Shadow and Fire)."""
    for bit, name in sorted(SCHOOLS.items(), reverse=True):
        if mask == bit:
            return name
    for bit, name in sorted(SCHOOLS.items()):
        if mask and mask & bit:
            return name
    return 'SpellSchoolPhysical'


def event_of(row, fights):
    """A raidDamage event (protojson) for one mined ability, and a sentence on where it's from."""
    raw = row['raw_p05'] > 0 and row['raw_p95'] > 0
    if raw:
        lo, hi = stretch(row['raw_p05'], row['raw_p95'])
    else:
        lo, hi = 0.9 * row['landed_median'], 1.1 * row['landed_median']
    lo, hi = max(1, round(lo)), max(1, round(hi))
    ev = {'spellId': 0 if row['ability'] == 1 else row['ability'], 'school': school(row['school']),
          'minDamage': lo, 'maxDamage': hi, 'who': 'NotTanks'}
    share = row['hit_fraction']
    if share < ALL:
        ev['share'] = round(share, 3)
    ev['firstAt'] = round(row['first'], 1)
    ev['everySeconds'] = round(1 / row['casts_per_second'], 2)
    ev['randomTiming'] = True
    if row['tick']:
        ticks = max(1, round(row['hits_per_target']))
        ev['ticks'] = ticks
        ev['tickSeconds'] = round(row['tick_seconds'] or 1, 2)
    elif row['ability'] == 1:
        ev['melee'] = True
    what = 'unmitigated' if raw else 'landed (no unmitigated amount in the logs)'
    text = (f'{row["name"]} {row["ability"]}{" ticks" if row["tick"] else ""} from {row["source"]} '
            f'({100 * row["share"]:.0f}% of the non-tanks\' damage, in {row["fights"]} of {fights} kills): '
            f'{fmt(lo)} to {fmt(hi)} {what}, the 5th to 95th percentile stretched to an even roll; '
            f'{100 * share:.0f}% of the non-tanks a cast (10th to 90th percentile {100 * row["hit_fraction_p10"]:.0f} to '
            f'{100 * row["hit_fraction_p90"]:.0f}), {row["casts_per_second"] * 60:.1f} casts a minute at random, '
            f'the first at {row["first"]:.0f} s (median)'
            + (f', {ev["ticks"]} ticks {ev["tickSeconds"]} s apart' if row['tick'] else '') + '.')
    return ev, text


def per_second(ev, others):
    """What an event deals a second to each non-tank on average, before mitigation."""
    hits = (ev.get('share') or 1) * (ev.get('ticks') or 1)
    return (ev['minDamage'] + ev['maxDamage']) / 2 * hits / ev['everySeconds']


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument('--site', choices=SITES, required=True)
    ap.add_argument('--out', help='where the files are (default: by site, tools/encounters for the Anniversary realms, '
                    'whose files score tanks (the user, 2026-11-02), tools/encounters/<site> for the others)')
    ap.add_argument('--boss', action='append', help='only this boss (repeat for more)')
    ap.add_argument('mined', nargs='+', help="wclcli.py raid -o files of this site")
    a = ap.parse_args()
    if not a.out:
        a.out = HERE if a.site == 'anniversary' else os.path.join(HERE, a.site)
    site = SITES[a.site]
    mined = {}
    for path in a.mined:
        for enc, m in json.load(open(path)).items():
            if not enc.startswith('_'):
                mined[int(enc)] = (m, os.path.relpath(path, ROOT))
    for boss, cfg in BOSSES.items():
        if a.boss and boss not in a.boss:
            continue
        enc_id = cfg['ids'][a.site]
        if enc_id not in mined:
            print(f'{boss}: no raid numbers for encounter {enc_id}, left as it is', file=sys.stderr)
            continue
        m, mined_path = mined[enc_id]
        path = os.path.join(a.out, boss + '.json')
        d = json.load(open(path))
        rows = [r for r in m['abilities'] if r['share'] >= MIN_SHARE and r['fights'] >= MIN_FIGHTS * m['fights']
                and r['source'] not in ('friendly', '?') and r['casts_per_second'] > 0]
        events, texts = [], []
        for r in rows:
            ev, text = event_of(r, m['fights'])
            events.append(ev)
            texts.append(text)
        covered = sum(r['share'] for r in rows)
        sim_dtps = sum(per_second(ev, 1) for ev in events)
        # An Everyone event of the tank side's (Vaelastrasz's Fire Nova) that the logs now give for
        # the non-tanks keeps its numbers on the tanks only, so nobody takes it twice.
        mined_ids = {ev['spellId'] for ev in events if ev['spellId']}
        kept = []
        for e in d['encounter'].get('raidDamage', []):
            if e.get('who') == 'NotTanks':
                continue
            if e.get('who', 'Everyone') == 'Everyone' and e.get('spellId') in mined_ids:
                e = dict(e, who='Tanks')
                texts.append(f'{e["spellId"]}\'s Everyone event keeps its own numbers on the tanks only.')
            kept.append(e)
        d['encounter']['raidDamage'] = kept + events

        h = m
        d['raidHealing'] = {
            'raidSize': round(h['players_median']), 'tanks': round(h['tanks_median']),
            'healers': round(h['healers_median']), 'healersRange': [h['healers_min'], h['healers_max']],
            'healerClasses': {k: round(v, 2) for k, v in sorted(h['healer_classes'].items(), key=lambda x: -x[1])},
            'hpsEachHealer': round(h['hps_each_median']),
            'raidHps': round(h['hps_effective']), 'raidHpsWithOverheal': round(h['hps_total']),
            'hpsOnTanks': round(h['hps_tanks_effective']), 'hpsOnOthers': round(h['hps_others_effective']),
            'hpsOnOthersWithOverheal': round(h['hps_others_total']),
            'dtpsEachOther': round(h['dtps_others_each'], 1), 'dtpsTanks': round(h['dtps_tanks_total']),
            'deathsOthersPerKill': round(h['deaths_others_mean'], 2), 'deathsTanksPerKill': round(h['deaths_tanks_mean'], 2),
            'killsWithoutDeath': f'{h["deathless_fights"]}/{h["fights"]}',
            'fightSeconds': [round(h['seconds_p10']), round(h['seconds_median']), round(h['seconds_p90'])],
            'notes': (f'What {site["label"]} raids had on {m["name"]} ({m["fights"]} kills, Warcraft Logs): the median kill\'s '
                      f'raid of {h["players_median"]:.0f} with {h["tanks_median"]:.0f} tanks the boss swung at, '
                      f'{h["healers_median"]:.0f} healers by Warcraft Logs\' healer role ({100 * h["healers_share"]:.0f}% of the healing), '
                      f'each {fmt(h["hps_each_median"])} HPS with overheal. The raid healed {fmt(h["hps_effective"])} a second '
                      f'({fmt(h["hps_tanks_effective"])} on the tanks, {fmt(h["hps_others_effective"])} on the rest, '
                      f'{100 * (1 - h["hps_others_effective"] / max(h["hps_others_total"], 1)):.0f}% overheal on the rest). '
                      f'Each non-tank took {h["dtps_others_each"]:.1f} a second after mitigation. Fight length is the 10th '
                      f'percentile, median and 90th percentile kill. For the healing sim to check its healer counts against.'),
        }
        d['raidSource'] = (f'[logs: {site["label"]}] The raidDamage events with who NotTanks: Warcraft Logs, {m["fights"]} kills '
                           f'(tools/warcraftlogs/wclcli.py raid, {mined_path}): the abilities that dealt {MIN_SHARE * 100:.0f}% or more '
                           f'of the non-tanks\' damage taken in {MIN_FIGHTS * 100:.0f}% or more of the kills, '
                           f'{100 * covered:.0f}% of it together. Each is a share of the non-tanks, so the file fits any raid size. '
                           + ' '.join(texts)
                           + f' Together {sim_dtps:.0f} a second on each non-tank before mitigation, against '
                             f'{h["dtps_others_each"]:.1f} landed in the logs.')
        d['raidReports'] = m['reports']
        with open(path, 'w') as f:
            json.dump(d, f, indent=1, ensure_ascii=False)
            f.write('\n')
        print(f'{boss}: {len(events)} raid events, {100 * covered:.0f}% of the non-tanks\' damage, '
              f'{sim_dtps:.0f} a second each before mitigation (logs {h["dtps_others_each"]:.0f} landed); '
              f'{h["healers_median"]:.0f} healers of {h["players_median"]:.0f} -> {os.path.relpath(path, ROOT)}')


if __name__ == '__main__':
    main()
