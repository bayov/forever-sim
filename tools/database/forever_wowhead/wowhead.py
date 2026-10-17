# Fetch pages from www.wowhead.com, and read the data blocks embedded in them.
#
# wowhead sits behind AWS WAF bot control. curl gets a flat 403 whatever user agent it
# sends, even for robots.txt. A request with a real browser's TLS fingerprint gets a
# JavaScript challenge instead (status 202, x-amzn-waf-action: challenge), and a browser
# that runs the challenge gets an aws-waf-token cookie that is good for about four days.
#
# So we open one page in headless Chrome to earn the cookie, and then fetch everything
# with curl_cffi, which sends Chrome's TLS fingerprint, plus that cookie. That's much
# faster than rendering every page in Chrome. The cookie is kept in
# ~/.cache/wow-data/wowhead/token.json, so a second run within a few days never starts
# Chrome at all. When wowhead challenges us again (a 202), we earn a new one.
#
# wowhead also limits how fast one address may ask. On 2026-10-17 about 110 requests in
# a minute or two got our address a flat 403 for every page, browser or not, token or
# not. A new token doesn't help with that, only waiting does. So we ask once every few
# seconds, and when a 403 comes anyway we wait it out (five minutes, then longer) and
# slow down for the rest of the run.
#
# That makes a full scrape slow, so every page we fetch is kept, gzipped, in
# ~/.cache/wow-data/wowhead/pages (the cache bayovui and forever-sim share, which
# WOW_DATA_CACHE moves). A run that stops halfway (a block that never lifts,
# a closed terminal) picks up where it stopped, and a page younger than
# WOWHEAD_CACHE_DAYS (an environment variable, 7 by default) isn't fetched again. Set it
# to 0 to fetch everything fresh.
#
# The scripts that use this declare their dependencies inline, so run them with uv:
#
#   uv run tools/database/forever_wowhead/scrape.py
import gzip, json, os, re, subprocess, sys, threading, time, urllib.parse, urllib.request

from curl_cffi import requests

BASE = 'https://www.wowhead.com/'
CACHE = os.path.join(os.environ.get('WOW_DATA_CACHE') or os.path.expanduser('~/.cache/wow-data'), 'wowhead')
TOKEN_FILE = os.path.join(CACHE, 'token.json')
PAGES = os.path.join(CACHE, 'pages')
# Seconds between requests, and the most we slow down to after blocks.
DELAY, MAX_DELAY = 3.5, 20.0
# How long we wait out each block in a row, in seconds. After the last we give up.
BLOCK_WAITS = [300, 600, 900, 1800, 1800]
PORT = 9334
# CloudFront answers headless Chrome's own user agent with a 403, a desktop one gets the
# challenge, which headless Chrome then solves like any browser.
USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36'


def earn_token():
    """Open one wowhead page in headless Chrome and return its aws-waf-token cookie."""
    import websocket
    profile = os.path.join(CACHE, 'chrome')
    chrome = subprocess.Popen(['google-chrome', '--headless=new', '--no-sandbox', '--disable-gpu',
        f'--remote-debugging-port={PORT}', '--remote-allow-origins=*', '--window-size=1600,1200',
        f'--user-data-dir={profile}', f'--user-agent={USER_AGENT}', 'about:blank'],
        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        tabs = None
        for _ in range(100):
            try:
                tabs = json.load(urllib.request.urlopen(f'http://127.0.0.1:{PORT}/json'))
                break
            except Exception:
                time.sleep(0.2)
        if tabs is None:
            raise RuntimeError('headless Chrome did not start (is google-chrome installed?)')
        ws = websocket.create_connection([t for t in tabs if t['type'] == 'page'][0]['webSocketDebuggerUrl'])
        ws.settimeout(60)
        seq = [0]
        def send(method, **params):
            seq[0] += 1
            ws.send(json.dumps({'id': seq[0], 'method': method, 'params': params}))
            while True:
                m = json.loads(ws.recv())
                if m.get('id') == seq[0]:
                    return m.get('result', {})
        send('Network.enable')
        send('Page.navigate', url=BASE + 'forever/item=2589')
        for _ in range(60):
            time.sleep(1)
            for c in send('Network.getCookies').get('cookies', []):
                if c['name'] == 'aws-waf-token':
                    return {'value': c['value'], 'expires': c.get('expires', time.time() + 3600)}
        raise RuntimeError('wowhead never handed headless Chrome an aws-waf-token cookie')
    finally:
        chrome.terminate()


class Wowhead:
    """Fetches wowhead pages one at a time, politely, with a shared WAF token."""

    def __init__(self, cache_days=None):
        if cache_days is None:
            cache_days = float(os.environ.get('WOWHEAD_CACHE_DAYS', 7))
        self.cache_days = cache_days
        self.delay = DELAY
        self.lock = threading.Lock()
        self.session = None
        self.token = None
        self.last = 0.0
        self.fetched = self.cached = 0
        try:
            with open(TOKEN_FILE) as f:
                token = json.load(f)
            # An hour of slack, so a long scrape doesn't run out halfway.
            if token['expires'] > time.time() + 3600:
                self.token = token
        except (OSError, ValueError, KeyError):
            pass

    def _renew(self):
        print('earning a wowhead token in headless Chrome', file=sys.stderr)
        self.token = earn_token()
        self.session = None
        os.makedirs(os.path.dirname(TOKEN_FILE), exist_ok=True)
        with open(TOKEN_FILE, 'w') as f:
            json.dump(self.token, f)

    def _get(self, url):
        if not self.token:
            self._renew()
        if self.session is None:
            self.session = requests.Session(impersonate='chrome')
            self.session.cookies.set('aws-waf-token', self.token['value'], domain='.wowhead.com')
        wait = self.last + self.delay - time.time()
        if wait > 0:
            time.sleep(wait)
        try:
            return self.session.get(url, timeout=60)
        finally:
            self.last = time.time()

    @staticmethod
    def _cache_path(url):
        return os.path.join(PAGES, urllib.parse.quote(url[len(BASE):], safe='') + '.html.gz')

    def page(self, path):
        """The HTML of a wowhead page, like page('forever/item=6270').

        None when wowhead says the page doesn't exist. Raises when wowhead keeps
        refusing us after every wait in BLOCK_WAITS.
        """
        url = path if path.startswith('http') else BASE + path
        cache = self._cache_path(url)
        with self.lock:
            try:
                if time.time() - os.path.getmtime(cache) < self.cache_days * 86400:
                    self.cached += 1
                    with gzip.open(cache, 'rt', encoding='utf-8') as f:
                        return f.read() or None
            except OSError:
                pass
            blocks = challenges = errors = 0
            while True:
                try:
                    r = self._get(url)
                except Exception as e:
                    errors += 1
                    if errors > 5:
                        raise
                    print(f'retrying {url}: {e}', file=sys.stderr)
                    time.sleep(10 * errors)
                    continue
                waf = r.headers.get('x-amzn-waf-action')
                if r.status_code in (200, 404) and not waf:
                    text = r.text if r.status_code == 200 else ''
                    os.makedirs(PAGES, exist_ok=True)
                    with gzip.open(cache, 'wt', encoding='utf-8') as f:
                        f.write(text)
                    self.fetched += 1
                    return text or None
                if r.status_code == 202 or waf == 'challenge':
                    # The token is spent. A fresh one fixes this, twice is suspicious.
                    challenges += 1
                    if challenges > 2:
                        raise RuntimeError(f'wowhead keeps challenging us at {url}')
                    self._renew()
                    continue
                if r.status_code in (403, 429) or r.status_code >= 500:
                    if blocks >= len(BLOCK_WAITS):
                        raise RuntimeError(f'wowhead still refuses us ({r.status_code}) at {url}, '
                                           f'after {self.fetched} pages fetched this run')
                    self.delay = min(MAX_DELAY, self.delay * 1.5)
                    print(f'wowhead answered {r.status_code} after {self.fetched} pages, waiting '
                          f'{BLOCK_WAITS[blocks] // 60} minutes, then one request every '
                          f'{self.delay:.1f}s', file=sys.stderr)
                    time.sleep(BLOCK_WAITS[blocks])
                    blocks += 1
                    continue
                raise RuntimeError(f'wowhead answered {r.status_code} at {url}')


# The tokens of a JavaScript literal: a double or single quoted string, a word, a
# number, or one punctuation character.
_TOKEN = re.compile(r'''"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|[A-Za-z_$][\w$]*|-?[\d.]+(?:[eE][+-]?\d+)?|\S''', re.S)


def _js_value(html, start):
    """The JavaScript object or array literal that begins at html[start], as Python.

    wowhead's data is nearly JSON, but some keys aren't quoted (firstseenpatch: 0) and a
    few strings use single quotes. We rewrite it token by token into JSON and stop at the
    bracket that closes the literal.
    """
    out = []
    depth = 0
    tokens = _TOKEN.finditer(html, start)
    for m in tokens:
        t = m.group(0)
        if t[0] == "'":
            t = json.dumps(re.sub(r"\\(.)", r"\1", t[1:-1]))
        elif t[0].isalpha() or t[0] in '_$':
            if t not in ('true', 'false', 'null'):
                # A bare word followed by a colon is a key. Anything else (undefined, a
                # variable) has no JSON value, so it reads as null.
                rest = html[m.end():m.end() + 20].lstrip()
                t = json.dumps(t) if rest.startswith(':') else 'null'
        elif t in '[{':
            depth += 1
        elif t in ']}':
            depth -= 1
            # JSON has no trailing commas.
            if out and out[-1] == ',':
                out.pop()
        out.append(t)
        if depth == 0:
            break
    return json.loads(' '.join(out))


def listview(html, id):
    """The rows of the Listview tab with this id ('sold-by', 'dropped-by', ...), or []."""
    m = re.search(r"new Listview\(\{[^{}]*?id: *'%s'" % re.escape(id), html)
    if not m:
        return []
    i = html.find('data:', m.start())
    return _js_value(html, i + len('data:'))


def listview_template(html, template):
    """The rows of the first Listview of this template ('quest', ...), or None."""
    i = html.find("template: '%s'" % template)
    if i < 0:
        return None
    j = html.find('data:', i)
    return _js_value(html, j + len('data:'))


def listview_items(html):
    """The rows of an item listing page (its `listviewitems` variable), or None."""
    i = html.find('listviewitems = ')
    if i < 0:
        return None
    return _js_value(html, i + len('listviewitems = '))


def gatherer(html, type):
    """Every WH.Gatherer.addData block of this type merged by id (3 is items)."""
    out = {}
    for m in re.finditer(r'WH\.Gatherer\.addData\(%d, \d+, ' % type, html):
        out.update(_js_value(html, m.end()))
    return out


def mapper(html):
    """The page's map pins by zone id: {zone: [{coords: [[x, y]], uiMapName, ...}]}."""
    i = html.find('g_mapperData = ')
    if i < 0:
        return {}
    pins = _js_value(html, i + len('g_mapperData = '))
    # An NPC wowhead has no position for (Plugger Spazzring in Blackrock Depths) has [].
    return pins if isinstance(pins, dict) else {}
