"""Parse a WoW Forever beta combat log into events. Usage: import cl; cl.load(path)."""
import csv, datetime

ME = 'Bayov-'

class Ev:
    __slots__ = ('t', 'ts', 'ev', 'src', 'dst', 'srcGUID', 'dstGUID', 'spellId', 'spell', 'f')
    def __repr__(self):
        return f"{self.ts} {self.ev} {self.src}>{self.dst} {self.spell or ''}"

def load(path):
    out = []
    for line in open(path, encoding='utf-8'):
        line = line.rstrip('\n')
        if '  ' not in line:
            continue
        ts, rest = line.split('  ', 1)
        f = next(csv.reader([rest]))
        e = Ev()
        d = datetime.datetime.strptime(ts, '%m/%d/%Y %H:%M:%S.%f')
        e.t = d.timestamp()
        e.ts = ts.split(' ')[1]
        e.ev = f[0]
        e.f = f
        e.src = f[2] if len(f) > 2 else ''
        e.dst = f[6] if len(f) > 6 else ''
        e.srcGUID = f[1] if len(f) > 1 else ''
        e.dstGUID = f[5] if len(f) > 5 else ''
        e.spellId = None
        e.spell = None
        if e.ev.startswith(('SPELL_', 'RANGE_')) and len(f) > 10:
            e.spellId = int(f[9]) if f[9].isdigit() else f[9]
            e.spell = f[10]
        out.append(e)
    return out

def dmg(e):
    """amount, raw, overkill, school, resisted, blocked, absorbed, crit, glancing, crushing"""
    if e.ev.startswith('SWING_DAMAGE'):
        t = e.f[-10:]
    else:  # SPELL_DAMAGE has one more trailing field (ST/AOE)
        t = e.f[-11:-1]
    def n(x):
        return None if x in ('nil', '') else int(float(x))
    return dict(amount=n(t[0]), raw=n(t[1]), overkill=n(t[2]), school=n(t[3]), resisted=n(t[4]),
                blocked=n(t[5]), absorbed=n(t[6]), crit=t[7] == '1', glancing=t[8] == '1', crushing=t[9] == '1')

def miss(e):
    """SWING_MISSED: missType at f[9]. SPELL_MISSED: missType at f[12]."""
    if e.ev == 'SWING_MISSED':
        return e.f[9]
    return e.f[12]

def mine(evs):
    return [e for e in evs if e.src.startswith(ME)]

def level_of(e):
    """Advanced field 'level' of the info unit, the last field before the damage suffix."""
    try:
        if e.ev.startswith('SWING_DAMAGE'):
            return int(e.f[-11])
        return int(e.f[-12])
    except Exception:
        return None
