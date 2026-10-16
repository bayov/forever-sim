"""Spell pushback from combat logs: every player's casts and channels that took hits.

Usage: python3 -I tools/combat_log/pushback.py LOG [LOG ...]   (plain or .gz)

A cast runs from SPELL_CAST_START to the caster's SPELL_CAST_SUCCESS. A channel runs from its
SPELL_AURA_APPLIED to its SPELL_AURA_REMOVED. A hit is a SWING_DAMAGE, SPELL_DAMAGE or
RANGE_DAMAGE on the caster with an amount over 0, from someone else. DoT ticks don't count,
as in the sim (sim/core/cast.go).

We take each caster's own time for a spell when nothing hit them (the median for casts, the
longest for channels, because players stop channels early), so talents and haste don't
matter. Then we compare the time of each cast that took hits with a few rules:

  cmangos   1 sec for the first hit, then 0.8, 0.6, 0.4 and 0.2 for every hit after that.
            A cast's bar can't go back past its start. A channel ends sooner, not before now.
            This is what the sim does.
  flat0.5   0.5 sec a hit (TBC and later).
  flat1     1 sec a hit.
  none      no pushback.

2026-10-16, the beta logs from 10-10 to 10-16: 285 casts with hits. The cmangos rule
predicts 243 within 0.1 sec (median error 0.042 sec), flat 1 sec 215, flat 0.5 sec 68 and no
pushback 23. Only 3 channels took hits, too few to tell the channel rule.
"""
import collections, gzip, os, shutil, statistics, sys, tempfile

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl

CHANNELS = {'Drain Life', 'Mind Flay', 'Drain Soul', 'Drain Mana', 'Health Funnel', 'Arcane Missiles',
            'Evocation', 'Hellfire', 'Tranquility', 'Hurricane', 'Blizzard', 'Volley', 'Rain of Fire'}
MODELS = ('cmangos', 'flat0.5', 'flat1', 'none')
HIT_EVENTS = ('SWING_DAMAGE', 'SPELL_DAMAGE', 'RANGE_DAMAGE')


def load(path):
    if not path.endswith('.gz'):
        return cl.load(path)
    tmp = tempfile.NamedTemporaryFile(delete=False, suffix='.txt')
    with gzip.open(path, 'rb') as f:
        shutil.copyfileobj(f, tmp)
    tmp.close()
    try:
        return cl.load(tmp.name)
    finally:
        os.unlink(tmp.name)


def delay(model, n):
    return {'cmangos': max(0.2, 1.0 - 0.2 * n), 'flat0.5': 0.5, 'flat1': 1.0, 'none': 0.0}[model]


def cast_end(base, hits, model):
    """When a cast of base seconds lands, with hits at these times from its start."""
    end = base
    for n, h in enumerate(hits):
        if h >= end - 1e-4:
            break
        progress = base - (end - h)
        end += min(delay(model, n), progress)
    return end


def channel_end(base, hits, model):
    end = base
    for n, h in enumerate(hits):
        if h >= end - 1e-4:
            break
        end = max(h, end - delay(model, n))
    return end


def is_hit(e, caster):
    if e.ev not in HIT_EVENTS or e.dstGUID != caster or e.srcGUID == caster:
        return False
    try:
        return (cl.dmg(e)['amount'] or 0) > 0
    except Exception:
        return False


def collect(paths):
    casts, channels = [], []  # (caster, spell, seconds, [hit times from the start])
    for path in paths:
        open_casts, open_channels = {}, {}
        for e in load(path):
            if e.ev == 'SPELL_CAST_START' and e.srcGUID.startswith('Player-'):
                open_casts[e.srcGUID] = (e, [])
            elif e.ev in ('SPELL_CAST_SUCCESS', 'SPELL_CAST_FAILED', 'SPELL_INTERRUPT') and e.srcGUID in open_casts:
                start, hits = open_casts.pop(e.srcGUID)
                if e.ev == 'SPELL_CAST_SUCCESS' and e.spellId == start.spellId:
                    casts.append((start.src, start.spell, e.t - start.t, [h - start.t for h in hits]))
            if e.ev == 'SPELL_AURA_APPLIED' and e.spell in CHANNELS and e.srcGUID.startswith('Player-'):
                open_channels[e.srcGUID] = (e, [])
            elif e.ev == 'SPELL_AURA_REMOVED' and e.srcGUID in open_channels and open_channels[e.srcGUID][0].spell == e.spell:
                start, hits = open_channels.pop(e.srcGUID)
                channels.append((start.src, start.spell, e.t - start.t, [h - start.t for h in hits]))
            for caster, (_, hits) in list(open_casts.items()) + list(open_channels.items()):
                if is_hit(e, caster):
                    hits.append(e.t)
    return casts, channels


def report(title, records, base_of, end_of, within):
    by = collections.defaultdict(list)
    for r in records:
        by[(r[0], r[1])].append(r)
    errors = collections.defaultdict(list)
    for (caster, spell), rs in sorted(by.items()):
        clean = [r[2] for r in rs if not r[3]]
        if len(clean) < 2 or base_of(clean) < 1.0:
            continue
        base = base_of(clean)
        for r in rs:
            if r[3]:
                for m in MODELS:
                    errors[m].append(abs(r[2] - end_of(base, r[3], m)))
    n = len(errors['none'])
    print(f'{title}: {n} with hits')
    for m in MODELS:
        if n:
            e = errors[m]
            print(f'  {m:8s} median error {statistics.median(e):.3f} sec, within {within} sec: {sum(x < within for x in e)}/{n}')


def main():
    casts, channels = collect(sys.argv[1:])
    report('Casts', casts, statistics.median, cast_end, 0.1)
    report('Channels', channels, max, channel_end, 0.15)


if __name__ == '__main__':
    main()
