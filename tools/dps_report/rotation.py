"""rotation.py: each preset's real APL (the rotation in its sim request) as plain-English priority rows."""
import json, os, re

OPS = {'OpLt': 'under', 'OpLe': 'at most', 'OpGt': 'over', 'OpGe': 'at least', 'OpEq': 'exactly', 'OpNe': 'not'}
SYM = {'OpLt': '<', 'OpLe': '≤', 'OpGt': '>', 'OpGe': '≥', 'OpEq': '=', 'OpNe': '≠'}
MATH = {'OpAdd': '+', 'OpSub': '−', 'OpMul': '×', 'OpDiv': '÷'}
AUTO = {'MainHand': 'main-hand', 'OffHand': 'off-hand', 'Ranged': 'ranged', 'Any': 'next'}


def key_of(ref):
    """APL spell/item/other ref -> the dashboard's name key."""
    if not isinstance(ref, dict):
        return None
    if ref.get('spellId'):
        return 's%d' % ref['spellId'] + ('t%d' % ref['tag'] if ref.get('tag') else '')
    if ref.get('itemId'):
        return 'i%d' % ref['itemId']
    if ref.get('otherId'):
        return 'o' + str(ref['otherId'])
    return None


def fmt_const(s):
    s = str(s).strip()
    m = re.fullmatch(r'(-?\d+(?:\.\d+)?)ms', s)
    if m:
        return f'{float(m.group(1)) / 1000:g} s'
    m = re.fullmatch(r'(-?\d+(?:\.\d+)?)s', s)
    if m:
        return f'{float(m.group(1)):g} s'
    m = re.fullmatch(r'(-?\d+(?:\.\d+)?)m', s)
    if m:
        return f'{float(m.group(1)):g} min'
    return s


def secs(s):
    s = str(s).strip()
    m = re.fullmatch(r'(-?\d+(?:\.\d+)?)(ms|s|m)?', s)
    if not m:
        return None
    v = float(m.group(1))
    return v / 1000 if m.group(2) == 'ms' else v * 60 if m.group(2) == 'm' else v


class Humaniser:
    def __init__(self, name):
        self.name = name  # key -> display name
        self.fallbacks = []

    def sp(self, ref):
        k = key_of(ref)
        return self.name(k) if k else '?'

    # a value as a noun phrase
    def noun(self, v):
        (t, a), = v.items()
        if t == 'const':
            return fmt_const(a.get('val'))
        if t == 'remainingTime':
            return 'fight time left'
        if t == 'currentTime':
            return 'fight time so far'
        if t == 'currentManaPercent':
            return 'mana'
        if t == 'currentHealthPercent':
            return 'health'
        if t in ('currentMana', 'currentRage', 'currentEnergy', 'currentFocus', 'currentComboPoints'):
            return {'currentMana': 'mana', 'currentRage': 'rage', 'currentEnergy': 'energy', 'currentFocus': 'focus', 'currentComboPoints': 'combo points'}[t]
        if t == 'spellTimeToReady':
            return f"{self.sp(a.get('spellId'))} cooldown left"
        if t == 'auraRemainingTime':
            return f"{self.sp(a.get('auraId'))} time left"
        if t == 'dotRemainingTime':
            return f"{self.sp(a.get('spellId'))} time left on the target"
        if t == 'auraNumStacks':
            return f"{self.sp(a.get('auraId'))} stacks"
        if t == 'spellUsesRemaining':
            return f"{self.sp(a.get('spellId'))} uses left in the fight"
        if t == 'spellCurrentCost':
            return f"{self.sp(a.get('spellId'))} cost"
        if t == 'spellCastTime':
            return f"{self.sp(a.get('spellId'))} cast time"
        if t == 'autoTimeToNext':
            return f"time to the {AUTO.get(a.get('autoType'), 'next')} swing"
        if t == 'autoSwingTime':
            return f"{AUTO.get(a.get('autoType'), '')} swing time".strip()
        if t == 'math':
            return f"{self.noun(a['lhs'])} {MATH.get(a.get('op'), '?')} {self.noun(a['rhs'])}"
        if t == 'spellChanneledTicks':
            return 'ticks channeled'
        if t == 'spellUsesLostByDelay':
            return f"{self.sp(a.get('spellId'))} uses lost by waiting {self.delay(a.get('delay'))}"
        self.fallbacks.append(t)
        return t

    def delay(self, d):
        if not d:
            return ''
        (t, a), = d.items()
        if t == 'spellTimeToReady':
            return f"for {self.sp(a.get('spellId'))}"
        if t == 'math' and a.get('op') == 'OpSub' and 'remainingTime' in a.get('lhs', {}) and 'const' in a.get('rhs', {}):
            return f"until the last {fmt_const(a['rhs']['const']['val'])} of the fight"
        if t == 'const':
            return fmt_const(a.get('val'))
        return self.noun(d)

    def pred(self, v, top=False):
        """a value used as a condition, as a clause"""
        (t, a), = v.items()
        if t in ('and', 'or'):
            vals = a.get('vals', [])
            if len(vals) == 1:
                return self.pred(vals[0])
            other = 'or' if t == 'and' else 'and'
            # wrap a child that joins the other way (by its node type, since names like Slice and Dice hold "and")
            parts = [f'({self.pred(x)})' if self.joins(x, other) else self.pred(x) for x in vals]
            out = f' {t} '.join(parts)
            m = re.search(r'in the last (\S+ \S+) of the fight', out)
            if m:  # "in the last 33 s ... or when waiting until the last 33 s ..." -> "until then"
                out = out.replace(f'until the last {m.group(1)} of the fight', 'until then')
            return out
        if t == 'not':
            return self.neg(a.get('val'))
        if t == 'cmp':
            return self.cmp(a)
        if t == 'const':
            return 'always' if str(a.get('val')).lower() in ('true', '1') else 'never'
        if t == 'auraIsActive':
            on = (a.get('sourceUnit') or {}).get('type') == 'CurrentTarget'
            return f"{self.sp(a.get('auraId'))} is {'on the target' if on else 'up'}"
        if t == 'auraIsKnown':
            return f"{self.sp(a.get('auraId'))} is talented"
        if t == 'dotIsActive':
            return f"{self.sp(a.get('spellId'))} is ticking on the target"
        if t == 'spellIsReady':
            return f"{self.sp(a.get('spellId'))} is ready"
        if t == 'spellCanCast':
            return f"{self.sp(a.get('spellId'))} can be cast"
        if t == 'spellInFlight':
            return f"{self.sp(a.get('spellId'))} is in flight"
        if t == 'isExecutePhase':
            th = re.sub(r'\D', '', str(a.get('threshold', '20'))) or '20'
            return f"the target is below {th}% health"
        if t == 'inMeleeRange':
            return 'in melee range'
        if t == 'spellIsChanneling':
            return f"channeling {self.sp(a.get('spellId'))}"
        self.fallbacks.append(t)
        return self.noun(v)

    def joins(self, v, how):
        (t, a), = v.items()
        if t == how and len(a.get('vals', [])) > 1:
            return True
        if t == 'not':
            (t2, a2), = a.get('val', {}).items()
            return t2 in ('and', 'or') and t2 != how and len(a2.get('vals', [])) > 1  # De Morgan flips it
        return False

    def neg(self, v):
        (t, a), = v.items()
        if t == 'auraIsActive':
            on = (a.get('sourceUnit') or {}).get('type') == 'CurrentTarget'
            return f"{self.sp(a.get('auraId'))} is not {'on the target' if on else 'up'}"
        if t == 'auraIsKnown':
            return f"{self.sp(a.get('auraId'))} is not talented"
        if t == 'dotIsActive':
            return f"{self.sp(a.get('spellId'))} is not on the target"
        if t == 'spellIsReady':
            return f"{self.sp(a.get('spellId'))} is not ready"
        if t == 'spellCanCast':
            return f"{self.sp(a.get('spellId'))} can't be cast"
        if t == 'spellInFlight':
            return f"{self.sp(a.get('spellId'))} is not in flight"
        if t == 'isExecutePhase':
            th = re.sub(r'\D', '', str(a.get('threshold', '20'))) or '20'
            return f"the target is above {th}% health"
        if t == 'inMeleeRange':
            return 'out of melee range'
        if t == 'cmp':  # not (a < b) is a >= b
            inv = {'OpLt': 'OpGe', 'OpLe': 'OpGt', 'OpGt': 'OpLe', 'OpGe': 'OpLt', 'OpEq': 'OpNe', 'OpNe': 'OpEq'}
            return self.cmp(dict(a, op=inv.get(a.get('op'), a.get('op'))))
        if t in ('and', 'or'):  # De Morgan keeps it readable: not (A and B) = A is not ... or B is not ...
            how = 'or' if t == 'and' else 'and'
            parts = [f'({self.neg(x)})' if self.joins(x, t) else self.neg(x) for x in a.get('vals', [])]
            return f' {how} '.join(parts)
        return f'not ({self.pred(v)})'

    def cmp(self, a):
        op, L, R = a.get('op'), a.get('lhs', {}), a.get('rhs', {})
        # constant on the left: flip
        if 'const' in L and 'const' not in R:
            flip = {'OpLt': 'OpGt', 'OpLe': 'OpGe', 'OpGt': 'OpLt', 'OpGe': 'OpLe'}
            op, L, R = flip.get(op, op), R, L
        (lt, la), = L.items() if L else (('?', {}),)
        if lt == 'const' and 'const' in R:  # two constants: a line switched on or off
            x, y = secs(la.get('val')), secs(R['const'].get('val'))
            if x is not None and y is not None:
                res = {'OpLt': x < y, 'OpLe': x <= y, 'OpGt': x > y, 'OpGe': x >= y, 'OpEq': x == y, 'OpNe': x != y}.get(op)
                return 'always' if res else 'never (this line is switched off)'
        c = fmt_const(R['const']['val']) if 'const' in R else None
        cv = secs(R['const']['val']) if 'const' in R else None
        w = OPS.get(op, op)
        if lt == 'remainingTime' and c:
            return f"in the last {c} of the fight" if op in ('OpLt', 'OpLe') else f"with more than {c} of the fight left"
        if lt == 'currentTime' and c:
            return f"in the first {c} of the fight" if op in ('OpLt', 'OpLe') else f"after the first {c} of the fight"
        if lt == 'currentComboPoints' and c:
            n = int(cv) if cv is not None else c
            return {'OpGe': f'{n}+ combo points', 'OpGt': f'more than {n} combo points', 'OpEq': f'exactly {n} combo points', 'OpNe': f'not at {n} combo points',
                    'OpLt': f'fewer than {n} combo points', 'OpLe': f'{n} or fewer combo points'}.get(op, f'combo points {w} {n}')
        if lt == 'spellUsesLostByDelay' and c is not None and cv == 0:
            d = self.delay(la.get('delay'))
            s = self.sp(la.get('spellId'))
            if op in ('OpGt', 'OpNe'):
                return f"when waiting {d} would cost a use of {s}"
            if op in ('OpEq', 'OpLe'):
                return f"when waiting {d} costs no use of {s}"
        if lt == 'spellUsesRemaining' and c is not None and cv == 0:
            s = self.sp(la.get('spellId'))
            if op in ('OpEq', 'OpLe'):
                return f"no {s} uses are left in the fight"
            if op in ('OpGt', 'OpNe'):
                return f"{s} has uses left in the fight"
        if lt == 'spellTimeToReady' and c:
            s = self.sp(la.get('spellId'))
            if cv == 0 and op in ('OpEq', 'OpLe'):
                return f"{s} is ready"
            return f"{s} is ready within {c}" if op in ('OpLt', 'OpLe') else f"{s} has more than {c} of cooldown left"
        if lt == 'spellChanneledTicks' and c:
            n = int(cv) if cv is not None else c
            return {'OpGe': f'at least {n} ticks channeled', 'OpGt': f'more than {n} ticks channeled', 'OpLt': f'fewer than {n} ticks channeled'}.get(op, f'ticks channeled {w} {n}')
        if lt == 'autoTimeToNext' and 'math' in R and R['math'].get('op') == 'OpSub' and 'autoSwingTime' in R['math'].get('lhs', {}) and 'const' in R['math'].get('rhs', {}) and op in ('OpGt', 'OpGe'):
            hand = AUTO.get(la.get('autoType'), 'next')
            return f"within {fmt_const(R['math']['rhs']['const']['val'])} after a {hand} swing"
        if lt in ('currentMana',) and 'math' in R and R['math'].get('op') == 'OpMul' and 'remainingTime' in R['math'].get('lhs', {}) and 'const' in R['math'].get('rhs', {}):
            k = fmt_const(R['math']['rhs']['const']['val'])
            return f"enough mana for {k} per second of the fight left" if op in ('OpGe', 'OpGt') else f"less mana than {k} per second of the fight left"
        if lt == 'auraRemainingTime' and c and cv == 0 and op in ('OpGt', 'OpNe'):
            return f"{self.sp(la.get('auraId'))} is up"
        if lt == 'auraRemainingTime' and c:
            s = self.sp(la.get('auraId'))
            return f"{s} has under {c} left" if op in ('OpLt', 'OpLe') else f"{s} has over {c} left"
        if lt == 'dotRemainingTime' and c:
            s = self.sp(la.get('spellId'))
            return f"{s} has under {c} left on the target" if op in ('OpLt', 'OpLe') else f"{s} has over {c} left on the target"
        if lt == 'auraNumStacks' and c:
            s = self.sp(la.get('auraId'))
            n = int(cv) if cv is not None else c
            return {'OpGe': f'{s} at {n}+ stacks', 'OpGt': f'{s} above {n} stacks', 'OpEq': f'{s} at {n} stacks',
                    'OpLt': f'{s} under {n} stacks', 'OpLe': f'{s} at {n} stacks or fewer'}.get(op, f'{s} stacks {w} {n}')
        if lt in ('currentMana', 'currentRage', 'currentEnergy', 'currentFocus') and 'spellCurrentCost' in R:
            s = self.sp(R['spellCurrentCost'].get('spellId'))
            res = self.noun(L)
            return f"enough {res} for {s}" if op in ('OpGe', 'OpGt') else f"not enough {res} for {s}"
        if lt in ('currentMana', 'currentRage', 'currentEnergy', 'currentFocus', 'currentManaPercent', 'currentHealthPercent') and c:
            res = self.noun(L)
            unit = '' if lt.endswith('Percent') else f' {res}'
            if lt.endswith('Percent'):
                pc = f'{cv * 100:g}%' if cv is not None and cv <= 1 else c
                return {'OpGe': f'{res} at least {pc}', 'OpGt': f'{res} above {pc}', 'OpLt': f'{res} under {pc}', 'OpLe': f'{res} at most {pc}'}.get(op, f'{res} {w} {pc}')
            return {'OpGe': f'{c}+{unit}', 'OpGt': f'more than {c}{unit}', 'OpLt': f'under {c}{unit}', 'OpLe': f'{c} or less{unit}', 'OpEq': f'exactly {c}{unit}'}.get(op, f'{res} {w} {c}')
        return f"{self.noun(L)} {w} {self.noun(R)}"


def clean_note(t):
    """APL notes: drop sentences about the sim UI's warnings; keep the explanation."""
    if not t:
        return ''
    sents = re.split(r'(?<=[.!?])\s+', t.strip())
    sents = [s for s in sents if not re.search(r'warning', s, re.I)]
    return ' '.join(sents).strip()


def build(req_dir, outs, nm, names, abilities, racial_race):
    rows, fallbacks = [], []
    casts = {(a['build'], a['key']): a['casts'] for a in abilities}
    for o in outs:
        bid, m = o['id'], o['meta']
        rq = json.load(open(os.path.join(req_dir, bid + '.json')))
        rot = rq['raid']['parties'][0]['players'][0].get('rotation') or {}
        H = Humaniser(lambda k: 'the potion' if k == 'oOtherActionPotion' else nm(k))
        race = m.get('race')

        def row(section, order, act, at=None, notes=''):
            cond = act.get('condition')
            kind, ref, extra = None, None, ''
            for k in act:
                if k == 'condition':
                    continue
                kind, body = k, act[k]
            if kind == 'castSpell':
                ref = body.get('spellId')
            elif kind == 'channelSpell':
                ref = body.get('spellId')
                if body.get('interruptIf'):
                    extra = 'Stops channeling when ' + H.pred(body['interruptIf'])
            elif kind == 'move':
                r = body.get('rangeFromTarget', {}).get('const', {}).get('val', '0')
                extra = 'Into melee range' if str(r) in ('0', '0s') else f'To {r} yd from the target'
            key = key_of(ref) if ref else None
            name = nm(key) if key else {'move': 'Move', 'autocastOtherCooldowns': 'Other cooldowns', 'wait': 'Wait'}.get(kind, kind)
            when = H.pred(cond, top=True) if cond else ('' if section == 'prepull' else 'whenever it is ready' if kind in ('castSpell', 'channelSpell') else '')
            if kind == 'autocastOtherCooldowns':
                when = 'any major cooldown not named in this list (trinkets, racials, potions) is used as soon as it is ready'
            rr = (racial_race.get(int(ref['spellId'])) if ref and ref.get('spellId') else None) or (racial_race.get(name) if key else None)
            mr = re.match(r'^(Human|Dwarf|Night Elf|Gnome|High Order|Orc|Undead|Tauren|Troll|Windshaper)\. ', notes or '')
            if not rr and mr:  # a racial line the notes name ("Undead. Dark Sacrifice when ...")
                rr = mr.group(1)
            rows.append({
                'build': bid, 'section': section, 'order': order,
                'kind': {'castSpell': 'cast', 'channelSpell': 'channel', 'move': 'move', 'autocastOtherCooldowns': 'auto'}.get(kind, kind),
                'key': key, 'name': name, 'icon': names[key]['icon'] if key and key in names else None,
                'at': at, 'when': when[:1].upper() + when[1:] if when else '', 'extra': extra, 'notes': clean_note(notes),
                'race': rr, 'otherRace': bool(rr and race and rr != race and not (isinstance(rr, list))),
                'casts': round(casts.get((bid, key), 0), 1) if key else None,
                'cond': json.dumps(cond, separators=(',', ':')) if cond else None,
            })

        for i, p in enumerate(sorted(rot.get('prepullActions') or [], key=lambda p: secs(p.get('doAtValue', {}).get('const', {}).get('val', '0')) or 0)):
            at = secs(p.get('doAtValue', {}).get('const', {}).get('val', '0'))
            row('prepull', i, p['action'], at=at, notes=p.get('notes', ''))
        for i, p in enumerate(rot.get('priorityList') or []):
            if p.get('hide'):
                continue
            row('priority', i + 1, p['action'], notes=p.get('notes', ''))
        fallbacks += H.fallbacks
    return rows, sorted(set(fallbacks))
