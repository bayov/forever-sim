"""Warcraft Logs v2 API client with a disk cache and a points budget.

Credentials come from ~/.config/forever-sim/wcl.json ({"client_id": ..., "client_secret": ...}),
made at https://www.warcraftlogs.com/api/clients/. Every client gets 3600 points an hour, and
the client is shared by everything we run, so we try hard not to spend them:

- Every response is cached under ~/.cache/forever-sim/wcl/. Report data never changes once a
  report is fully uploaded, so we keep it forever. Rankings and zone lists change, so we keep
  them for a day (pass ttl=None for forever).
- We wait at least MIN_GAP seconds between two requests.
- Each run has a points budget (BUDGET, set by the --budget flag of wclcli.py). We read how
  many points the hour has spent before the first request, and we stop with an error when this
  run has spent its budget or the hour is close to the limit.
"""
import hashlib, json, os, sys, time

import requests

CRED_PATH = os.path.expanduser('~/.config/forever-sim/wcl.json')
CACHE_DIR = os.path.expanduser('~/.cache/forever-sim/wcl')
TOKEN_URL = 'https://www.warcraftlogs.com/oauth/token'
UA = 'forever-sim tools/warcraftlogs'

MIN_GAP = 1.0
BUDGET = 300
# We stop when the hour has spent this much of its limit, so other runs still have room.
HOUR_STOP = 0.9

# The site decides which game's data we see: "sod" for Season of Discovery, "vanilla" for
# Anniversary and Era, "fresh" for Fresh, "classic" for the progression servers. WoW Forever
# may get its own subdomain at release, and SITE is just the subdomain.
SITE = 'sod'

DAY = 24 * 3600

_state = {'token': None, 'last': 0.0, 'spent': 0.0, 'start_spent': None, 'limit': None,
          'requests': 0, 'cache_hits': 0}


class BudgetError(RuntimeError):
    pass


def _token():
    if _state['token']:
        return _state['token']
    path = os.path.join(CACHE_DIR, 'token.json')
    try:
        t = json.load(open(path))
        if t['expires_at'] > time.time() + 3600:
            _state['token'] = t['access_token']
            return t['access_token']
    except (OSError, ValueError, KeyError):
        pass
    cred = json.load(open(CRED_PATH))
    r = requests.post(TOKEN_URL, data={'grant_type': 'client_credentials'},
                      auth=(cred['client_id'], cred['client_secret']), timeout=30)
    r.raise_for_status()
    t = r.json()
    t['expires_at'] = time.time() + t['expires_in']
    os.makedirs(CACHE_DIR, exist_ok=True)
    with open(path, 'w') as f:
        json.dump(t, f)
    os.chmod(path, 0o600)
    _state['token'] = t['access_token']
    return t['access_token']


def _post(query, variables):
    wait = _state['last'] + MIN_GAP - time.time()
    if wait > 0:
        time.sleep(wait)
    url = f'https://{SITE}.warcraftlogs.com/api/v2/client'
    headers = {'Authorization': 'Bearer ' + _token(), 'User-Agent': UA}
    for attempt in range(3):
        r = requests.post(url, json={'query': query, 'variables': variables}, headers=headers,
                          timeout=120)
        _state['last'] = time.time()
        _state['requests'] += 1
        if r.status_code == 429:
            # We never retry quickly on a rate limit. The hour is spent, so we stop.
            raise BudgetError('Warcraft Logs says we hit the rate limit (429). Wait for the hour to reset.')
        if r.status_code >= 500 and attempt < 2:
            time.sleep(10 * (attempt + 1))
            continue
        r.raise_for_status()
        body = r.json()
        if body.get('errors'):
            raise RuntimeError(f'GraphQL errors: {body["errors"]}')
        return body['data']


RATE_QUERY = 'rateLimitData { limitPerHour pointsSpentThisHour pointsResetIn }'


def _check_budget():
    """Read the hour's points before the first request, then check our spend before each one."""
    if _state['start_spent'] is None:
        d = _post('{ ' + RATE_QUERY + ' }', {})['rateLimitData']
        _state['start_spent'] = d['pointsSpentThisHour']
        _state['limit'] = d['limitPerHour']
        if d['pointsSpentThisHour'] >= HOUR_STOP * d['limitPerHour']:
            raise BudgetError(f'The hour has spent {d["pointsSpentThisHour"]:.0f} of {d["limitPerHour"]} points. '
                              f'It resets in {d["pointsResetIn"]} sec.')
    if _state['spent'] >= BUDGET:
        raise BudgetError(f'This run spent its budget of {BUDGET} points. Raise --budget to go on '
                          '(the cache keeps everything fetched so far).')


def query(q, variables=None, ttl=None):
    """Run a GraphQL query and return its data, from the cache when we have it.

    ttl is the cache lifetime in seconds, None for forever. We add rateLimitData to every
    query we send, so we always know how many points the hour has spent.
    """
    variables = variables or {}
    key = hashlib.sha1(json.dumps([SITE, q, variables], sort_keys=True).encode()).hexdigest()
    path = os.path.join(CACHE_DIR, key[:2], key + '.json')
    try:
        if ttl is None or time.time() - os.path.getmtime(path) < ttl:
            data = json.load(open(path))
            _state['cache_hits'] += 1
            return data
    except OSError:
        pass
    _check_budget()
    body = q.strip()
    assert body.endswith('}'), q
    data = _post(body[:-1] + ' ' + RATE_QUERY + ' }', variables)
    rate = data.pop('rateLimitData')
    _state['spent'] = rate['pointsSpentThisHour'] - _state['start_spent']
    if rate['pointsSpentThisHour'] >= HOUR_STOP * rate['limitPerHour']:
        # We still save this answer, but the next uncached query will stop.
        _state['spent'] = max(_state['spent'], BUDGET)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path + '.tmp', 'w') as f:
        json.dump(data, f)
    os.replace(path + '.tmp', path)
    return data


def summary():
    """One line about what this run cost, for the end of each command."""
    return (f'[wcl] {_state["requests"]} requests, {_state["cache_hits"]} from cache, '
            f'{_state["spent"]:.1f} points spent this run')


# Report queries. Reports are immutable once uploaded, so they are cached forever.

def report_meta(code):
    """Fights, actors and abilities of a report."""
    return query('''query($code: String!) { reportData { report(code: $code) {
        title startTime endTime zone { id name }
        fights { id encounterID name startTime endTime kill difficulty size friendlyPlayers
                 enemyNPCs { id gameID } }
        masterData { gameVersion
            actors { id gameID name type subType petOwner server }
            abilities { gameID name type } }
    } } }''', {'code': code})['reportData']['report']


def events(code, fight=None, data_type='All', source_id=None, target_id=None, ability_id=None):
    """Every event of one fight (or of the whole report when fight is None), page by page.

    Each page is one cached query. With source_id we also get the events of that player's
    pets, so callers check sourceID themselves.
    """
    meta = report_meta(code)
    if fight is None:
        start, end, fights = 0, meta['endTime'] - meta['startTime'], None
    else:
        f = next(x for x in meta['fights'] if x['id'] == fight)
        start, end, fights = f['startTime'], f['endTime'], [fight]
    out = []
    args = {'code': code, 'fights': fights, 'end': end, 'dt': data_type, 'src': source_id,
            'dst': target_id, 'ab': ability_id}
    while start is not None:
        args['start'] = start
        d = query('''query($code: String!, $fights: [Int], $start: Float, $end: Float,
                           $dt: EventDataType, $src: Int, $dst: Int, $ab: Float) {
            reportData { report(code: $code) {
                events(fightIDs: $fights, startTime: $start, endTime: $end, dataType: $dt,
                       sourceID: $src, targetID: $dst, abilityID: $ab, limit: 10000) {
                    data nextPageTimestamp }
            } } }''', args)['reportData']['report']['events']
        out += d['data']
        start = d['nextPageTimestamp']
    return out


def player_events(code, fight, player):
    """What one player did, plus the buffs others put on them, in time order.

    Two queries: the player's own events, and the buffs on the player (a Windfury Totem proc
    can show up as a buff from the shaman on a warrior).
    """
    own = [e for e in events(code, fight, source_id=player) if e.get('sourceID') == player]
    # With the Buffs data type, Warcraft Logs reads sourceID as the unit that holds the aura.
    buffs = [e for e in events(code, fight, data_type='Buffs', source_id=player)
             if e.get('targetID') == player and e.get('sourceID') != player]
    return sorted(own + buffs, key=lambda e: e['timestamp'])


def die(msg):
    print(msg, file=sys.stderr)
    sys.exit(1)
