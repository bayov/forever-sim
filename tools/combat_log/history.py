#!/usr/bin/env python3
"""Copy BayovCore's character history out of the game's SavedVariables into ~/forever-logs.

BayovCore's export module (bayovui/addons/BayovCore/Modules/Export.lua in the WoW folder) also
records the character's setup as it changes: a full snapshot at login, then each gear swap,
talent change, level up and skill up, and the sheet after a change. Each record is a compact JSON
string. Sheets and talent lists are stored once and records point at them by id.

The addon keeps only the last 7 days, and WoW writes the file only on /reload, logout or exit.
So archive.sh runs this every 5 minutes, and we keep every record we've ever seen:

    ~/forever-logs/history/<Name>-<Realm>.jsonl        one record per line, in time order
    ~/forever-logs/history/<Name>-<Realm>.blobs.json   the sheets and talent lists, by id

A record we already have is skipped, so running this again changes nothing. logcheck reads these
files and rebuilds an export for each moment a record describes.

    python3 tools/combat_log/history.py [--game DIR] [--archive DIR] [--flavors "_classic_beta_"]
"""
import argparse
import glob
import json
import os
import re
import sys


class LuaParser:
    """Reads the Lua table constructors WoW writes into a SavedVariables file.

    Only what the client writes: `name = value` at the top, tables with `["key"] = value`,
    `[1] = value` and bare list values, strings, numbers, booleans and nil, and `-- [n]`
    comments after list values.
    """

    TOKEN = re.compile(r'\s+|--[^\n]*|"(?:[^"\\]|\\.)*"|[{}\[\]=,;]|[^\s{}\[\]=,;"]+', re.S)

    def __init__(self, text):
        self.tokens = [t for t in self.TOKEN.findall(text) if t.strip() and not t.startswith('--')]
        self.i = 0

    def peek(self):
        return self.tokens[self.i] if self.i < len(self.tokens) else None

    def take(self, want=None):
        t = self.peek()
        if want is not None and t != want:
            raise ValueError('expected %r at token %d, got %r' % (want, self.i, t))
        self.i += 1
        return t

    def globals(self):
        out = {}
        while self.peek() is not None:
            name = self.take()
            self.take('=')
            out[name] = self.value()
        return out

    def value(self):
        t = self.peek()
        if t == '{':
            return self.table()
        self.take()
        if t.startswith('"'):
            return unescape(t[1:-1])
        if t == 'true':
            return True
        if t == 'false':
            return False
        if t == 'nil':
            return None
        try:
            return int(t)
        except ValueError:
            return float(t)

    def table(self):
        self.take('{')
        keyed, listed = {}, []
        while self.peek() != '}':
            if self.peek() == '[':
                self.take('[')
                key = self.value()
                self.take(']')
                self.take('=')
                keyed[key] = self.value()
            elif self.i + 1 < len(self.tokens) and self.tokens[self.i + 1] == '=':
                key = self.take()
                self.take('=')
                keyed[key] = self.value()
            else:
                listed.append(self.value())
            if self.peek() in (',', ';'):
                self.take()
        self.take('}')
        if listed and not keyed:
            return listed
        for n, v in enumerate(listed, 1):
            keyed[n] = v
        # A list WoW wrote as [1] = ..., [2] = ... is a list too.
        if keyed and all(isinstance(k, int) for k in keyed) and sorted(keyed) == list(range(1, len(keyed) + 1)):
            return [keyed[k] for k in sorted(keyed)]
        return keyed


ESCAPES = {'n': '\n', 'r': '\r', 't': '\t', '\\': '\\', '"': '"', "'": "'", '\n': '\n', 'a': '\a', 'b': '\b',
           'f': '\f', 'v': '\v'}


def unescape(s):
    """A Lua string literal's body as text. The client writes UTF-8 bytes as they are."""
    out, i = bytearray(), 0
    raw = s.encode('utf-8')
    while i < len(raw):
        c = raw[i]
        if c != 0x5c:
            out.append(c)
            i += 1
            continue
        nxt = chr(raw[i + 1])
        if nxt.isdigit():
            m = re.match(rb'\d{1,3}', raw[i + 1:i + 4])
            out.append(int(m.group()))
            i += 1 + len(m.group())
        else:
            out += ESCAPES.get(nxt, nxt).encode('utf-8')
            i += 2
    return out.decode('utf-8', errors='replace')


def history_chars(path):
    """The export module's history, by character, in one SavedVariables file, or {} when it has none."""
    with open(path, encoding='utf-8', errors='replace') as f:
        text = f.read()
    if '"chars"' not in text:
        return {}
    db = LuaParser(text).globals().get('BayovCoreDB') or {}
    export = ((db.get('modules') or {}).get('export') or {})
    return export.get('chars') or {}


def merge(archive, key, char):
    """Adds the character's records and blobs we don't have yet. Answers how many records were
    new."""
    os.makedirs(archive, exist_ok=True)
    safe = re.sub(r'[^\w.-]', '_', key)
    records_path = os.path.join(archive, safe + '.jsonl')
    blobs_path = os.path.join(archive, safe + '.blobs.json')

    lines = []
    if os.path.exists(records_path):
        with open(records_path, encoding='utf-8') as f:
            lines = [l.rstrip('\n') for l in f if l.strip()]
    seen = set(lines)

    new = []
    for r in char.get('records') or []:
        j = r.get('j') if isinstance(r, dict) else None
        if not j or j in seen:
            continue
        try:
            json.loads(j)
        except ValueError as e:
            print('%s: skipping a record that is not JSON (%s)' % (key, e), file=sys.stderr)
            continue
        seen.add(j)
        new.append(j)

    blobs = {}
    if os.path.exists(blobs_path):
        with open(blobs_path, encoding='utf-8') as f:
            blobs = json.load(f)
    added_blobs = 0
    for blob_id, b in (char.get('blobs') or {}).items():
        if blob_id in blobs or not isinstance(b, dict) or not b.get('j'):
            continue
        blobs[blob_id] = json.loads(b['j'])
        added_blobs += 1

    if added_blobs:
        tmp = blobs_path + '.tmp'
        with open(tmp, 'w', encoding='utf-8') as f:
            json.dump(blobs, f, separators=(',', ':'), sort_keys=True)
        os.replace(tmp, blobs_path)
    if new:
        # Records come in time order within one file, but two accounts or a late flush can
        # interleave, so we sort the whole file by time. The sort is stable, so records from
        # the same second keep the order the addon wrote them in.
        both = lines + new
        both.sort(key=lambda j: json.loads(j).get('t', 0))
        tmp = records_path + '.tmp'
        with open(tmp, 'w', encoding='utf-8') as f:
            f.write('\n'.join(both) + '\n')
        os.replace(tmp, records_path)
    return len(new)


def main():
    p = argparse.ArgumentParser(description=__doc__.split('\n')[0])
    p.add_argument('--game', default=os.environ.get('WOW_DIR', '/mnt/d/Gaming/World of Warcraft'))
    p.add_argument('--archive', default=os.path.join(os.environ.get('FOREVER_LOGS', os.path.expanduser('~/forever-logs')), 'history'))
    p.add_argument('--flavors', default=os.environ.get('FLAVORS', '_classic_beta_'))
    args = p.parse_args()

    for flavor in args.flavors.split():
        pattern = os.path.join(args.game, flavor, 'WTF', 'Account', '*', 'SavedVariables', 'BayovCore.lua')
        for path in sorted(glob.glob(pattern)):
            try:
                chars = history_chars(path)
            except (ValueError, IndexError) as e:
                print('%s: %s' % (path, e), file=sys.stderr)
                continue
            for key, char in chars.items():
                n = merge(args.archive, key, char)
                if n:
                    print('%s: %d new records' % (key, n))


if __name__ == '__main__':
    main()
