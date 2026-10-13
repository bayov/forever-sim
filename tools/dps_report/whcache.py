"""whcache.py: wowhead tooltip JSON for WoW Forever (and Classic as the fallback), cached on disk.

Forever:  https://nether.wowhead.com/forever/tooltip/<item|spell>/<id>?dataEnv=16&locale=0
Classic:  https://nether.wowhead.com/classic/tooltip/<item|spell>/<id>?dataEnv=4&locale=0
The /forever/ path picks Forever's database whatever dataEnv says (16, 17 or none answer the same). Without the path, dataEnv 16 is
Forever and dataEnv 17 is retail's squished data (Earthstrike item level 13, Forever-only ids not found).
Every answer (an error too) is kept under forever_tt_cache/<domain>/<kind>/<id>.json, so a rebuild reads the
disk and never the network. `python3 whcache.py ids.json` fetches what is missing (ids.json: [[kind, id], ...]).
"""
import json, os, ssl, sys, time, urllib.error, urllib.request

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D
CACHE = os.environ.get('DASH_TT_CACHE') or os.path.join(D, 'forever_tt_cache')
ENV = {'forever': 16, 'classic': 4}
FALLBACK = 'classic'


def url(domain, kind, i):
    return f'https://nether.wowhead.com/{domain}/tooltip/{kind}/{i}?dataEnv={ENV[domain]}&locale=0'


def page(domain, kind, i):
    return f'https://www.wowhead.com/{domain}/{kind}={i}'


def path(domain, kind, i):
    return os.path.join(CACHE, domain, kind, f'{int(i)}.json')


def cached(domain, kind, i):
    p = path(domain, kind, i)
    if not os.path.exists(p):
        return None
    with open(p) as f:
        return json.load(f)


def ok(j):
    return bool(j) and not j.get('error') and bool(j.get('name'))


_ctx = None


def fetch(domain, kind, i, pause=0.35):
    global _ctx
    j = cached(domain, kind, i)
    if j is not None:
        return j
    if _ctx is None:
        ca = '/root/.ccr/ca-bundle.crt'
        _ctx = ssl.create_default_context(cafile=ca if os.path.exists(ca) else None)
    req = urllib.request.Request(url(domain, kind, i), headers={'User-Agent': 'forever-sim-dashboard/1.0', 'Accept': 'application/json'})
    delay = 2.0
    for attempt in range(6):
        try:
            with urllib.request.urlopen(req, timeout=30, context=_ctx) as r:
                body = r.read().decode('utf-8')
            j = json.loads(body)
            break
        except urllib.error.HTTPError as e:
            if e.code == 404:
                try:
                    j = json.loads(e.read().decode('utf-8'))
                except Exception:
                    j = {'error': 'HTTP 404'}
                break
            if e.code in (429, 500, 502, 503, 504):
                time.sleep(delay)
                delay *= 2
                continue
            raise
        except (urllib.error.URLError, TimeoutError, ConnectionError):
            time.sleep(delay)
            delay *= 2
    else:
        raise RuntimeError(f'wowhead unreachable for {domain} {kind} {i}')
    os.makedirs(os.path.dirname(path(domain, kind, i)), exist_ok=True)
    with open(path(domain, kind, i), 'w') as f:
        json.dump(j, f)
    time.sleep(pause)
    return j


def lookup(kind, i):
    """(domain, json) for the Forever answer, else Classic's, else (None, None). Reads the cache only."""
    j = cached('forever', kind, i)
    if ok(j):
        return 'forever', j
    c = cached(FALLBACK, kind, i)
    if ok(c):
        return FALLBACK, c
    return None, None


if __name__ == '__main__':
    ids = json.load(open(sys.argv[1]))
    n = 0
    for kind, i in ids:
        had = cached('forever', kind, i) is not None
        j = fetch('forever', kind, i)
        if not ok(j):
            had = had and cached(FALLBACK, kind, i) is not None
            fetch(FALLBACK, kind, i)
        if not had:
            n += 1
            if n % 50 == 0:
                print('fetched', n, file=sys.stderr, flush=True)
    print('done; network fetches:', n)
