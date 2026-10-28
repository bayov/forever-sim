#!/usr/bin/env python3
"""race_override.py: a build's own rotation and gear for one race (standards 29), read from req/races/.

A build sims each race with the preset's rotation and gear unless `req/races/<id>.<Race>.json` exists (`<Race>` is the
proto race name without "Race": Dwarf, NightElf, SkyborneHighOrder, ...):

  {"rotation": {...an APL rotation, the shape of the request's player.rotation...},   optional
   "bonusStats": {"SpellHit": -1, "SpellPower": 11.6},                                optional, proto.Stat names
   "note": "one line for the Races tab",
   "gain": {"dps": .., "se": .., "iterations": 20000, "seed": 1, "duration": 180},
   "how": "how it was made"}

`rotation` replaces the player's rotation and `bonusStats` is added to the player's bonus stats (on top of the gear).
A duty or encounter variant (a build with `baseBuild`) without a file of its own takes its base build's, but only the
parts that fit it: the rotation when the variant's rotation is the base's, the bonus stats when its gear is the base's
(a duty preset's rotation keeps its duty up, and a variant may wear other gear). The preset's own race is never
overridden: its request is the preset.

  race_override.py [ids...]   print what every build sims at each race
"""
import copy, glob, json, os, re, sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D  # noqa: E402

HERE = os.path.dirname(os.path.abspath(__file__))
DIR = os.path.join(D, 'req', 'races')


def _enum(name):
    src = open(os.path.join(HERE, '..', '..', 'proto', 'common.proto')).read()
    body = src[src.index('enum %s {' % name):]
    body = body[:body.index('}')]
    return {m.group(1): int(m.group(2)) for m in re.finditer(r'^\s*(\w+)\s*=\s*(\d+);', body, re.M)}


_STATS, _PSEUDO = None, None


def stat_index(name):
    """('stats', i) or ('pseudoStats', i) for a proto.Stat name without "Stat" (or a PseudoStat name)."""
    global _STATS, _PSEUDO
    if _STATS is None:
        _STATS, _PSEUDO = _enum('Stat'), _enum('PseudoStat')
    if 'Stat' + name in _STATS:
        return 'stats', _STATS['Stat' + name]
    for k in ('PseudoStat' + name, name):
        if k in _PSEUDO:
            return 'pseudoStats', _PSEUDO[k]
    raise KeyError(f'unknown stat {name!r} in a race override')


def _meta(bid):
    return json.load(open(os.path.join(D, 'req', bid + '.meta.json')))


def _player(bid):
    return json.load(open(os.path.join(D, 'req', bid + '.json')))['raid']['parties'][0]['players'][0]


def tuned_kind(ov):
    """What a race's own setup changes, from its file or the override an out/ file records: 'rotation', 'gear' or
    'rotation+gear' ('' for none). The Races tab names it on the race's badge."""
    if not ov:
        return ''
    return '+'.join(k for k, on in (('rotation', ov.get('rotation')), ('gear', ov.get('bonusStats'))) if on)


# The report's words for each race's own setup: a `note` for the Races tab and, for a rotation, a `why` per line it
# changes ("<section>:<spell key, or the action's name>"). They live apart from the override files, which the sims read
# and the staleness checks hash, so rewording them doesn't make a build stale.
TEXT = os.path.join(HERE, 'race_text.json')
_text = None


def text(ov):
    """The race_text.json entry for an override (its base build's file for a variant), or {}."""
    global _text
    if _text is None:
        _text = json.load(open(TEXT)) if os.path.exists(TEXT) else {}
    return _text.get(re.sub(r'\.json$', '', (ov or {}).get('file', '')), {})


def note(ov):
    """What the report says about a race's own setup: race_text.json's note, else the override file's."""
    return text(ov).get('note') or (ov or {}).get('note', '')


def own_path(bid, race):
    return os.path.join(DIR, f'{bid}.{race}.json')


def files(bid):
    """Every override file a build's sims read: its own, and its base build's for a variant."""
    meta = _meta(bid)
    ids = [bid] + ([meta['baseBuild']] if meta.get('baseBuild') else [])
    return sorted(p for i in ids for p in glob.glob(os.path.join(DIR, i + '.*.json')))


def load(bid, race, meta=None):
    """The override this build sims at this race, or None. For a variant on its base's file, only the parts that fit,
    with 'from' naming the base and 'dropped' the parts left out."""
    meta = meta or _meta(bid)
    if race == meta['race']:
        return None
    p = own_path(bid, race)
    if os.path.exists(p):
        return dict(json.load(open(p)), file=os.path.basename(p))
    base = meta.get('baseBuild')
    if not base or not os.path.exists(own_path(base, race)):
        return None
    ov = dict(json.load(open(own_path(base, race))), file=os.path.basename(own_path(base, race)), **{'from': base})
    mine, theirs = _player(bid), _player(base)
    dropped = []
    if 'rotation' in ov and mine.get('rotation') != theirs.get('rotation'):
        dropped.append('rotation')
        del ov['rotation']
    if 'bonusStats' in ov and mine.get('equipment') != theirs.get('equipment'):
        dropped.append('bonusStats')
        del ov['bonusStats']
    if dropped:
        ov['dropped'] = dropped
    return ov if ('rotation' in ov or 'bonusStats' in ov) else None


def apply(req, bid, race, meta=None):
    """(the request at this race with its override applied, the override or None). Sets the race either way."""
    req = copy.deepcopy(req)
    pl = req['raid']['parties'][0]['players'][0]
    pl['race'] = 'Race' + race
    ov = load(bid, race, meta)
    if not ov:
        return req, None
    if 'rotation' in ov:
        pl['rotation'] = copy.deepcopy(ov['rotation'])
    if ov.get('bonusStats'):
        bs = pl.setdefault('bonusStats', {})
        for name, v in ov['bonusStats'].items():
            field, i = stat_index(name)
            arr = bs.setdefault(field, [])
            arr.extend([0] * (i + 1 - len(arr)))
            arr[i] += v
    return req, ov


def summary(ov):
    """What a race's numbers were simmed with, for out/<id>.json and the Races tab: the note and what it changes."""
    if not ov:
        return None
    s = {'note': ov.get('note', ''), 'file': ov['file'], 'rotation': 'rotation' in ov, 'bonusStats': ov.get('bonusStats') or {}}
    if ov.get('from'):
        s['from'] = ov['from']
    if ov.get('dropped'):
        s['dropped'] = ov['dropped']
    if ov.get('gain'):
        s['gain'] = ov['gain']
    return s


if __name__ == '__main__':
    from run_build import CLASS_RACES
    ids = sys.argv[1:] or sorted(os.path.basename(p)[:-10] for p in glob.glob(os.path.join(D, 'req', '*.meta.json')))
    for bid in ids:
        m = _meta(bid)
        for race in CLASS_RACES[m['class']]:
            ov = load(bid, race, m)
            if ov:
                parts = (['rotation'] if 'rotation' in ov else []) + (['bonusStats ' + json.dumps(ov['bonusStats'])] if ov.get('bonusStats') else [])
                extra = (f" (from {ov['from']}" + (f", left out: {', '.join(ov['dropped'])}" if ov.get('dropped') else '') + ')') if ov.get('from') else ''
                print(f"{bid} {race}: {' + '.join(parts)}{extra}: {ov.get('note', '')}")
