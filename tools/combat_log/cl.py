"""Parse a WoW Forever beta combat log into events. Usage: import cl; cl.load(path).

The path may be a plain log or an .xz one, like wow-combat-logs keeps them."""
import csv, datetime, lzma


def is_me(guid, flags):
    """Whether a unit is our own character, from the flags the client writes with it.

    Only our character has a player's flags with the "mine" affiliation (0x511). A party member
    is 0x512 and everyone else 0x518. We never go by the name, because the game changes how it
    writes names ("Bayov-ClassicBetaPvE2-" became "Bayov-Oom-ClassicBetaPvE2-" for the same
    character) and we play more than one character.
    """
    try:
        n = int(flags, 16)
    except ValueError:
        return False
    return guid.startswith('Player-') and n & 0xF == 0x1 and n & 0xFC00 == 0x400


class Ev:
    __slots__ = ('t', 'ts', 'ev', 'src', 'dst', 'srcGUID', 'dstGUID', 'me_src', 'me_dst', 'spellId', 'spell', 'f')
    def __repr__(self):
        return f"{self.ts} {self.ev} {self.src}>{self.dst} {self.spell or ''}"

def load(path):
    out = []
    opener = lzma.open if path.endswith('.xz') else open
    for line in opener(path, 'rt', encoding='utf-8'):
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
        e.me_src = len(f) > 3 and is_me(e.srcGUID, f[3])
        e.me_dst = len(f) > 7 and is_me(e.dstGUID, f[7])
        e.spellId = None
        e.spell = None
        if e.ev.startswith(('SPELL_', 'RANGE_', 'DAMAGE_')) and len(f) > 10:
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
    return [e for e in evs if e.me_src]

def level_of(e):
    """Advanced field 'level' of the info unit, the last field before the damage suffix."""
    try:
        if e.ev.startswith('SWING_DAMAGE'):
            return int(e.f[-11])
        return int(e.f[-12])
    except Exception:
        return None
