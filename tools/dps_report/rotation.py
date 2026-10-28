"""rotation.py: each preset's real APL (the rotation in its sim request) as plain-English priority rows."""
import json, os, re

OPS = {'OpLt': 'under', 'OpLe': 'at most', 'OpGt': 'over', 'OpGe': 'at least', 'OpEq': 'exactly', 'OpNe': 'not'}
SYM = {'OpLt': '<', 'OpLe': '≤', 'OpGt': '>', 'OpGe': '≥', 'OpEq': '=', 'OpNe': '≠'}
MATH = {'OpAdd': 'plus', 'OpSub': 'minus', 'OpMul': 'times', 'OpDiv': 'divided by'}
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
        return f'{float(m.group(1)) / 1000:g}s'
    m = re.fullmatch(r'(-?\d+(?:\.\d+)?)s', s)
    if m:
        return f'{float(m.group(1)):g}s'
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


# An aura a talent gives under another name, for "<talent> talent taken".
AURA_TALENT = {'Fire Vulnerability': 'Improved Scorch'}


def load_talents(tree_root):
    """{class: {talent name: (tree, index)}} from the sim UI's talent trees (ui/core/talents/trees/<class>.json)."""
    out = {}
    d = os.path.join(tree_root, 'ui/core/talents/trees')
    for f in sorted(os.listdir(d)):
        cls = f[:-5].capitalize()
        for ti, tr in enumerate(json.load(open(os.path.join(d, f)))):
            for i, tal in enumerate(tr['talents']):
                out.setdefault(cls, {})[tal['name']] = (ti, i)
    return out


def talent_token(cls, talents, aura):
    """A talent in a condition's text: [[tal:<class>:<tree>:<index>|<name>]], which the report draws with the talent's
    icon and tooltip. The aura's own name when no talent of the class has it."""
    name = AURA_TALENT.get(aura, aura)
    at = (talents or {}).get(cls, {}).get(name)
    return f'[[tal:{cls}:{at[0]}:{at[1]}|{name}]]' if at else name


def _once(text, how):
    """Two ranks of one spell read the same, so a condition can name the same part twice ("A or A"): keep the first.
    Splits on the connective outside parentheses, where nested parts of the same kind have been joined in."""
    parts, depth, cur = [], 0, ''
    for tok in re.split(r'(\(|\)| %s )' % how, text):
        if tok == f' {how} ' and depth == 0:
            parts.append(cur)
            cur = ''
            continue
        depth += tok.count('(') - tok.count(')')
        cur += tok
    parts.append(cur)
    return f' {how} '.join(dict.fromkeys(parts))


def _merge(parts, how):
    """Parts that say the same thing of several spells, said once: "a melee weave fits before the next Auto Shot and
    a melee weave fits before Arcane Shot is ready" -> "a melee weave fits before the next Auto Shot and Arcane Shot"."""
    pats = [(r'(a|no) melee weave fits before (?:(the next Auto Shot)|(.+) is ready)', lambda m: (m.group(1), m.group(2) or m.group(3)),
             lambda k, xs: f'{k} melee weave fits before {xs}'),
            (r"(.+) (is|isn't) ready by the time you reach melee", lambda m: (m.group(2), m.group(1)),
             lambda k, xs: f"{xs} {'are' if k == 'is' and how == 'and' else k} ready by the time you reach melee")]
    for rx, key, say in pats:
        groups = {}
        for i, p in enumerate(parts):
            m = p and re.fullmatch(rx, p)
            if m:
                k, item = key(m)
                groups.setdefault(k, []).append((i, item))
        for k, its in groups.items():
            if len(its) < 2:
                continue
            items = [x for _, x in its]
            parts[its[0][0]] = say(k, ', '.join(items[:-1]) + f' {how} ' + items[-1])
            for i, _ in its[1:]:
                parts[i] = None
    return [p for p in parts if p is not None]


def art(word):
    """'a' or 'an' before a name"""
    return ('an ' if word[:1].lower() in 'aeiou' else 'a ') + word


class Humaniser:
    def __init__(self, name, talent=None):
        self.name = name  # key -> display name
        self.talent = talent or (lambda aura: aura)  # an aura's name -> its talent, as talent_token writes it
        self.cur = None  # the key of the line's own spell: "waiting would cost a use" needs no name for it
        self.items = set()  # the item names conditions use (tidy)
        self.fallbacks = []

    def sp(self, ref):
        k = key_of(ref)
        n = self.name(k) if k else '?'
        if k and k.startswith('i'):
            self.items.add(n)
        return n

    def tidy(self, when):
        """A whole condition, once more: a trinket named again and again is "the trinket" after the first time, and
        "no Hawk 1 ... or no Hawk 2 ..." is a count."""
        for n in sorted(self.items, key=len, reverse=True):
            if re.search(r'Trinket|Talisman', n) and when.count(n) > 1:
                i = when.index(n) + len(n)
                when = when[:i] + when[i:].replace(n, 'the trinket').replace('no more the trinket this fight', 'no more trinket uses this fight')
        return re.sub(r'no (\w+) 1 on the boss or no \1 2 on the boss', r'fewer than 2 \1s out', when)

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
            return {'currentMana': 'mana', 'currentRage': 'Rage', 'currentEnergy': 'Energy', 'currentFocus': 'Focus', 'currentComboPoints': 'combo points'}[t]
        if t == 'spellTimeToReady':
            return f"{self.sp(a.get('spellId'))}'s cooldown"
        if t == 'auraRemainingTime':
            return f"{self.sp(a.get('auraId'))}'s time left"
        if t == 'dotRemainingTime':
            return f"{self.sp(a.get('spellId'))}'s time left"
        if t == 'auraNumStacks':
            return f"{self.sp(a.get('auraId'))} stacks"
        if t == 'spellUsesRemaining':
            return f"{self.sp(a.get('spellId'))} uses left"
        if t == 'spellCurrentCost':
            return f"{self.sp(a.get('spellId'))} cost"
        if t == 'spellCastTime':
            return f"{self.sp(a.get('spellId'))} cast time"
        if t == 'autoTimeToNext':
            return f"time to the next {AUTO.get(a.get('autoType'), '')} swing".replace('  ', ' ')
        if t == 'autoSwingTime':
            return f"{AUTO.get(a.get('autoType'), '')} swing time".strip()
        if t == 'math':
            if a.get('op') in ('OpAdd', 'OpSub') and secs((a.get('rhs') or {}).get('const', {}).get('val')) == 0:  # "+ 0s" adds nothing
                return self.noun(a['lhs'])
            return f"{self.noun(a['lhs'])} {MATH.get(a.get('op'), '?')} {self.noun(a['rhs'])}"
        if t == 'max' and a.get('vals'):
            return 'the longer of ' + ' and '.join(self.noun(x) for x in a['vals'])
        if t == 'moveTime':
            return 'the time to reach melee'
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
        if t == 'max' and len(a.get('vals', [])) > 1:
            parts = [re.sub(r'^(for|until) ', '', self.delay(x)) for x in a['vals']]
            return f"for both {' and '.join(parts)}"
        if t == 'const':
            return fmt_const(a.get('val'))
        return ''  # a delay we can't say plainly is left out (the exact condition is in the tooltip)

    def pred(self, v, top=False):
        """a value used as a condition, as a clause"""
        (t, a), = v.items()
        if t in ('and', 'or'):
            how, vals = self.parts(t, a.get('vals', []), False)
            if not vals:
                return 'always' if (vals is None) == (how == 'or') else 'never (this line is switched off)'
            if len(vals) == 1:
                return self.pred(vals[0])
            # wrap a part whose text joins the other way (by node type, since names like Slice and Dice hold "and")
            parts = [f'({self.pred(x)})' if self.joiner(x, False) not in (None, how) else self.pred(x) for x in vals]
            if how == 'and':  # "X is not ready and X is ready within 0.1s" -> "X comes off cooldown within 0.1s"
                for i, p in enumerate(parts):
                    m = re.fullmatch(r'(.+) is not ready', p or '')
                    for j, q in enumerate(parts):
                        m2 = m and re.fullmatch(re.escape(m.group(1)) + r' is ready within (\S+)', q or '')
                        if m2:
                            parts[i], parts[j] = None, f'{m.group(1)} comes off cooldown within {m2.group(1)}'
                parts = [p for p in parts if p is not None]
            out = _once(f' {how} '.join(_merge(parts, how)), how)
            m = re.search(r'in the last (\S+(?: \S+)?) of the fight', out)
            if m:  # "in the last 33s ... or waiting until the last 33s ..." -> "waiting"
                out = out.replace(f' until the last {m.group(1)} of the fight', '')
            return re.sub(r'in the first (\S+) of the fight or in the last (\S+) of the fight', r'in the first \1 or the last \2 of the fight', out)
        if t == 'not':
            return self.neg(a.get('val'))
        if t == 'cmp':
            return self.cmp(a)
        if t == 'const':
            return 'always' if str(a.get('val')).lower() in ('true', '1') else 'never'
        if t == 'auraIsActive':
            on = (a.get('sourceUnit') or {}).get('type') == 'CurrentTarget'
            s = self.sp(a.get('auraId'))
            if s.endswith(' Stance'):
                return f'in {s}'
            return f"{s} {'on the boss' if on else 'is up'}"
        if t == 'auraIsKnown':
            return f"{self.talent(self.sp(a.get('auraId')))} talent taken"
        if t == 'dotIsActive':
            return f"{self.sp(a.get('spellId'))} on the boss"
        if t == 'spellIsReady':
            return f"{self.sp(a.get('spellId'))} is ready"
        if t == 'spellCanCast':
            return f"{self.sp(a.get('spellId'))} can be cast"
        if t == 'spellInFlight':
            return f"{self.sp(a.get('spellId'))} is in flight"
        if t == 'isExecutePhase':
            th = re.sub(r'\D', '', str(a.get('threshold', '20'))) or '20'
            return f"boss under {th}% health"
        if t == 'inMeleeRange':
            return 'in melee range'
        if t == 'gcdIsReady':
            return 'the global cooldown is ready'
        if t == 'frontOfTarget':
            return 'in front of the boss'
        if t == 'spellIsChanneling':
            return f"channeling {self.sp(a.get('spellId'))}"
        if t == 'sequenceIsComplete':
            return f"the {str(a.get('sequenceName') or '').title()} opener is done".replace('the  opener', 'the opener')
        self.fallbacks.append(t)
        return self.noun(v)

    def const_of(self, v):
        """True or False when a condition part has a fixed value, else None"""
        (t, a), = v.items()
        if t == 'const':
            return str(a.get('val')).lower() in ('true', '1')
        # "The Sunder opener is done" on every line after the opener says nothing new, because the opener comes
        # first in the list. So we count it as true and leave it out. "Not done" still shows.
        if t == 'sequenceIsComplete':
            return True
        if t == 'not':
            if 'sequenceIsComplete' in a.get('val', {}):
                return None
            c = self.const_of(a.get('val'))
            return None if c is None else not c
        if t == 'cmp':
            L, R = a.get('lhs', {}), a.get('rhs', {})
            if 'const' in L and 'const' in R:
                x, y = secs(L['const'].get('val')), secs(R['const'].get('val'))
                if x is not None and y is not None:
                    return {'OpLt': x < y, 'OpLe': x <= y, 'OpGt': x > y, 'OpGe': x >= y, 'OpEq': x == y, 'OpNe': x != y}.get(a.get('op'))
            if 'auraRemainingTime' in L and secs(R.get('const', {}).get('val')) == 0 and a.get('op') == 'OpGe':
                return True  # time left is never negative
        return None

    def parts(self, t, vals, neg):
        """The connective an and/or node's text uses (flipped when negated) and the parts it joins.

        A part with a fixed value that changes nothing (true in an and, false in an or) is left out.
        The parts come back as None when one fixes the whole node (false in an and, true in an or).
        """
        how = t if not neg else ('or' if t == 'and' else 'and')
        out = []
        for x in vals:
            c = self.const_of(x)
            if c is not None and neg:
                c = not c
            if c is (how == 'or'):
                return how, None
            if c is not (how == 'and'):
                out.append(x)
        return how, out

    def joiner(self, v, neg):
        """the connective at the top of a condition part's text, or None"""
        (t, a), = v.items()
        if t in ('and', 'or'):
            how, vals = self.parts(t, a.get('vals', []), neg)
            if not vals:
                return None
            return self.joiner(vals[0], neg) if len(vals) == 1 else how
        if t == 'not':
            return self.joiner(a.get('val'), not neg)
        return None

    def neg(self, v):
        (t, a), = v.items()
        if t == 'auraIsActive':
            on = (a.get('sourceUnit') or {}).get('type') == 'CurrentTarget'
            s = self.sp(a.get('auraId'))
            if s.endswith(' Stance'):
                return f'not in {s}'
            return f"{'no ' + s + ' on the boss' if on else s + ' is not up'}"
        if t == 'auraIsKnown':
            return f"{self.talent(self.sp(a.get('auraId')))} talent not taken"
        if t == 'dotIsActive':
            return f"no {self.sp(a.get('spellId'))} on the boss"
        if t == 'spellIsReady':
            return f"{self.sp(a.get('spellId'))} is not ready"
        if t == 'spellCanCast':
            return f"{self.sp(a.get('spellId'))} can't be cast"
        if t == 'spellInFlight':
            return f"{self.sp(a.get('spellId'))} is not in flight"
        if t == 'isExecutePhase':
            th = re.sub(r'\D', '', str(a.get('threshold', '20'))) or '20'
            return f"boss over {th}% health"
        if t == 'inMeleeRange':
            return 'out of melee range'
        if t == 'gcdIsReady':
            return 'the global cooldown is not ready'
        if t == 'frontOfTarget':
            return 'behind the boss'
        if t == 'cmp':  # not (a < b) is a >= b
            inv = {'OpLt': 'OpGe', 'OpLe': 'OpGt', 'OpGt': 'OpLe', 'OpGe': 'OpLt', 'OpEq': 'OpNe', 'OpNe': 'OpEq'}
            return self.cmp(dict(a, op=inv.get(a.get('op'), a.get('op'))))
        if t in ('and', 'or'):  # De Morgan keeps it readable: not (A and B) = A is not ... or B is not ...
            how, vals = self.parts(t, a.get('vals', []), True)
            if not vals:
                return 'always' if (vals is None) == (how == 'or') else 'never (this line is switched off)'
            if len(vals) == 1:
                return self.neg(vals[0])
            parts = [f'({self.neg(x)})' if self.joiner(x, True) not in (None, how) else self.neg(x) for x in vals]
            return _once(f' {how} '.join(_merge(parts, how)), how)
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
        more, less = op in ('OpGe', 'OpGt'), op in ('OpLe', 'OpLt')
        if 'moveTime' in json.dumps(a):
            m = self.weave(more, less, lt, la, R)
            if m:
                return m
        Rm = R.get('math', {})
        if lt == 'remainingTime' and ('spellCastTime' in R or (Rm.get('op') == 'OpAdd' and 'spellCastTime' in Rm.get('lhs', {}) and 'const' in Rm.get('rhs', {}))):
            sc = R.get('spellCastTime') or Rm['lhs']['spellCastTime']
            plus = f" plus {fmt_const(Rm['rhs']['const']['val'])}" if Rm else ''
            return f"{'less' if less else 'more'} than {art(self.sp(sc.get('spellId')))} cast{plus} left in the fight"
        if lt == 'remainingTime' and Rm.get('op') == 'OpSub' and 'const' in Rm.get('lhs', {}) and 'currentTime' in Rm.get('rhs', {}):
            # time left + time so far is the fight's length
            n = fmt_const(Rm['lhs']['const']['val'])
            return f"the fight lasts {n} or more" if more else f"the fight is shorter than {n}"
        if lt == 'math' and la.get('op') == 'OpDiv' and 'currentMana' in la.get('lhs', {}) and 'const' in la.get('rhs', {}) and 'remainingTime' in R:
            k = fmt_const(la['rhs']['const']['val'])
            return f"at least {k} mana per second of fight left" if more else f"under {k} mana per second of fight left"
        if lt == 'math' and la.get('op') == 'OpSub' and 'autoSwingTime' in la.get('lhs', {}) and 'autoTimeToNext' in la.get('rhs', {}) and c:
            hand = AUTO.get(la['lhs']['autoSwingTime'].get('autoType'), '')
            return f"{'over' if more else 'under'} {c} since the last {hand} swing".replace('  ', ' ')
        if lt == 'math' and la.get('op') == 'OpAdd' and 'const' not in la.get('rhs', {}) and c:
            return f"{self.noun(la['lhs'])} and {self.noun(la['rhs'])} add up to {'under' if less else 'at least' if op == 'OpGe' else 'over'} {c}"
        if lt in ('dotRemainingTime', 'auraRemainingTime') and 'spellCastTime' in R:
            s = self.sp(la.get('spellId') or la.get('auraId'))
            cast = art(self.sp(R['spellCastTime'].get('spellId')))
            return f"{s} has less time left than {cast} cast" if less else f"{s} has more time left than {cast} cast"
        if lt == 'spellCastTime' and c:
            s = self.sp(la.get('spellId'))
            return {'OpGe': f'{s} takes {c} or more to cast', 'OpGt': f'{s} takes over {c} to cast', 'OpLt': f'{s} casts in under {c}', 'OpLe': f'{s} casts in {c} or less'}.get(op, f'{s} cast time {w} {c}')
        if lt == 'remainingTime' and c:
            return f"in the last {c} of the fight" if op in ('OpLt', 'OpLe') else f"over {c} of the fight left"
        if lt == 'currentTime' and c:
            return f"in the first {c} of the fight" if op in ('OpLt', 'OpLe') else f"after the first {c} of the fight"
        if lt == 'currentComboPoints' and c:
            n = int(cv) if cv is not None else c
            if (op == 'OpGe' and n == 1) or (op == 'OpGt' and n == 0):
                return 'any combo points'
            if op == 'OpNe' and n == 5:
                return 'fewer than 5 combo points'
            return {'OpGe': f'{n}+ combo points', 'OpGt': f'more than {n} combo points', 'OpEq': f'exactly {n} combo points', 'OpNe': f'not at {n} combo points',
                    'OpLt': f'fewer than {n} combo points', 'OpLe': f'{n} or fewer combo points'}.get(op, f'combo points {w} {n}')
        if lt == 'spellUsesLostByDelay' and c is not None and cv == 0:
            d = self.delay(la.get('delay'))
            of = '' if key_of(la.get('spellId')) == self.cur else f' of {self.sp(la.get("spellId"))}'
            if op in ('OpGt', 'OpNe'):
                return f"waiting {d} would cost a use{of}".replace('  ', ' ')
            if op in ('OpEq', 'OpLe'):
                return f"waiting {d} costs no use{of}".replace('  ', ' ')
        if lt == 'spellUsesRemaining' and c is not None and cv == 0:
            s = self.sp(la.get('spellId'))
            if op in ('OpEq', 'OpLe'):
                return f"no more {s} this fight"
            if op in ('OpGt', 'OpNe'):
                return f"{s} can still be used this fight"
        if lt == 'spellTimeToReady' and c:
            s = self.sp(la.get('spellId'))
            if cv == 0 and op in ('OpEq', 'OpLe'):
                return f"{s} is ready"
            if cv == 0 and op in ('OpGt', 'OpNe'):
                return f"{s} is not ready"
            return f"{s} is ready within {c}" if op in ('OpLt', 'OpLe') else f"over {c} until {s} is ready"
        if lt == 'spellChanneledTicks' and c:
            n = int(cv) if cv is not None else c
            return {'OpGe': f'at least {n} ticks channeled', 'OpGt': f'more than {n} ticks channeled', 'OpLt': f'fewer than {n} ticks channeled'}.get(op, f'ticks channeled {w} {n}')
        if lt == 'autoTimeToNext' and 'math' in R and R['math'].get('op') == 'OpSub' and 'autoSwingTime' in R['math'].get('lhs', {}) and 'const' in R['math'].get('rhs', {}) and op in ('OpGt', 'OpGe'):
            hand = AUTO.get(la.get('autoType'), 'next')
            return f"within {fmt_const(R['math']['rhs']['const']['val'])} after a {hand} swing"
        if lt == 'autoTimeToNext' and c:
            hand = AUTO.get(la.get('autoType'), 'next')
            sw = 'next swing' if hand == 'next' else f'next {hand} swing'
            return f"{sw} in under {c}" if op in ('OpLt', 'OpLe') else f"{sw} in over {c}"
        if lt == 'spellTimeToReady' and 'math' in R and R['math'].get('op') == 'OpSub' and 'remainingTime' in R['math'].get('lhs', {}) and 'const' in R['math'].get('rhs', {}) and op in ('OpGt', 'OpGe'):
            # its cooldown ends after fight time left - c: it won't come back before the last c of the fight
            return f"{self.sp(la.get('spellId'))} won't be ready again before the last {fmt_const(R['math']['rhs']['const']['val'])} of the fight"
        if lt in ('currentMana',) and 'math' in R and R['math'].get('op') == 'OpMul' and 'remainingTime' in R['math'].get('lhs', {}) and 'const' in R['math'].get('rhs', {}):
            k = fmt_const(R['math']['rhs']['const']['val'])
            return f"at least {k} mana per second of fight left" if op in ('OpGe', 'OpGt') else f"under {k} mana per second of fight left"
        if lt == 'auraRemainingTime' and c and cv == 0 and op in ('OpGt', 'OpNe'):
            return f"{self.sp(la.get('auraId'))} is up"
        if lt == 'auraRemainingTime' and c:
            s = self.sp(la.get('auraId'))
            return f"{s} has under {c} left" if op in ('OpLt', 'OpLe') else f"{s} has over {c} left"
        if lt == 'dotRemainingTime' and c:
            s = self.sp(la.get('spellId'))
            return f"{s} has under {c} left" if op in ('OpLt', 'OpLe') else f"{s} has over {c} left"
        if lt == 'auraNumStacks' and c:
            s = self.sp(la.get('auraId'))
            n = int(cv) if cv is not None else c
            if n == 1 and op == 'OpLt':
                return f'no {s} stacks'
            st = 'stack' if n == 1 else 'stacks'
            return {'OpGe': f'{s} at {n}+ stacks', 'OpGt': f'{s} above {n} {st}', 'OpEq': f'{s} at {n} {st}',
                    'OpLt': f'{s} under {n} stacks', 'OpLe': f'{s} at {n} {st} or fewer'}.get(op, f'{s} stacks {w} {n}')
        if lt in ('currentMana', 'currentRage', 'currentEnergy', 'currentFocus') and 'spellCurrentCost' in R:
            s = self.sp(R['spellCurrentCost'].get('spellId'))
            res = self.noun(L)
            return f"enough {res} for {s}" if op in ('OpGe', 'OpGt') else f"not enough {res} for {s}"
        if lt in ('currentMana', 'currentRage', 'currentEnergy', 'currentFocus', 'currentManaPercent', 'currentHealthPercent') and c:
            res = self.noun(L)
            unit = '' if lt.endswith('Percent') else f' {res}'
            if lt.endswith('Percent'):
                pc = f'{cv * 100:g}%' if cv is not None and cv <= 1 else c
                return {'OpGe': f'at least {pc} {res}', 'OpGt': f'over {pc} {res}', 'OpLt': f'under {pc} {res}', 'OpLe': f'{pc} {res} or less'}.get(op, f'{res} {w} {pc}')
            return {'OpGe': f'{c}+{unit}', 'OpGt': f'more than {c}{unit}', 'OpLt': f'under {c}{unit}', 'OpLe': f'{c} or less{unit}', 'OpEq': f'exactly {c}{unit}'}.get(op, f'{res} {w} {c}')
        return f"{self.noun(L)} {w} {self.noun(R)}"

    def weave(self, more, less, lt, la, R):
        """A hunter's melee weave (step into melee, swing, step back out to shoot) in words, or None. The exact
        timings and margins are in the tooltip's exact condition."""
        rj = json.dumps(R)
        if 'max' in rj and 'moveTime' in rj:  # R is the weave: the longer of the next swing and the run in, then out
            if lt == 'math' and 'autoTimeToNext' in la.get('lhs', {}):  # the next ranged swing, plus a margin
                return 'a melee weave fits before the next Auto Shot' if more else 'no melee weave fits before the next Auto Shot'
            if lt == 'spellTimeToReady':
                s = self.sp(la.get('spellId'))
                return f'a melee weave fits before {s} is ready' if more else f'no melee weave fits before {s} is ready'
        if 'moveTime' in R:
            if lt == 'autoTimeToNext':
                return 'the next melee swing is ready by the time you reach melee' if less else "the next melee swing isn't ready by the time you reach melee"
            if lt == 'spellTimeToReady':
                s = self.sp(la.get('spellId'))
                return f'{s} is ready by the time you reach melee' if less else f"{s} isn't ready by the time you reach melee"
        if lt == 'math' and 'spellCastTime' in la.get('lhs', {}) and 'moveTime' in la.get('rhs', {}) and 'autoTimeToNext' in R:
            s = self.sp(la['lhs']['spellCastTime'].get('spellId'))
            return f'time to cast {s} and reach melee before the next melee swing' if less else f'no time to cast {s} and reach melee before the next melee swing'
        return None


def clean_note(t):
    """APL notes: drop sentences about the sim UI's warnings; keep the explanation."""
    if not t:
        return ''
    sents = re.split(r'(?<=[.!?])\s+', t.strip())
    sents = [s for s in sents if not re.search(r'warning', s, re.I)]
    return ' '.join(sents).strip()


def apl_rows(bid, rot, race, nm, names, casts, racial_race, H):
    """One APL's prepull and priority rows, as the preset's rotation dataset shows them. race is the race it is played
    at (its racials are the ones that fire); casts maps (build, key) to casts per fight, empty for none known."""
    rows = []

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
                extra = 'Stops channeling once ' + H.pred(body['interruptIf'])
        elif kind == 'move':
            r = body.get('rangeFromTarget', {}).get('const', {}).get('val', '0')
            extra = 'Into melee range' if str(r) in ('0', '0s') else f'To {r} yd from the boss'
        # A sequence (an opener like Sunder Armor, or Shadowmeld into Ambush) is named by its spells, and its first
        # spell gives the icon, the tooltip and the racial it belongs to.
        steps = [key_of(a['castSpell'].get('spellId')) for a in (body.get('actions') or []) if 'castSpell' in a] if kind in ('sequence', 'strictSequence') else []
        if steps:
            ref = (body['actions'][0].get('castSpell') or {}).get('spellId')
        key = key_of(ref) if ref else None
        H.cur = key
        name = ', then '.join(nm(k) for k in steps) if steps else nm(key) if key else {'move': 'Move', 'autocastOtherCooldowns': 'Other cooldowns', 'wait': 'Wait'}.get(kind, kind)
        when = H.tidy(H.pred(cond, top=True)) if cond else ''
        if not cond or when == 'always':  # no condition, or only one we leave out (see Humaniser.const_of)
            when = '' if section == 'prepull' else 'when ready' if kind in ('castSpell', 'channelSpell') else ''
        if kind == 'autocastOtherCooldowns':
            when = 'trinkets, racials and potions not listed here, as soon as they are ready'
        rr = (racial_race.get(int(ref['spellId'])) if ref and ref.get('spellId') else None) or (racial_race.get(name) if key else None)
        mr = re.match(r'^(Human|Dwarf|Night Elf|Gnome|High Order|Orc|Undead|Tauren|Troll|Windshaper)[.:] ', notes or '')
        if not rr and mr:  # a racial line the notes name ("Undead. Dark Sacrifice when ...")
            rr = mr.group(1)
        rows.append({
            'build': bid, 'section': section, 'order': order,
            'kind': {'castSpell': 'cast', 'channelSpell': 'channel', 'move': 'move', 'autocastOtherCooldowns': 'auto', 'strictSequence': 'sequence'}.get(kind, kind),
            'key': key, 'name': name, 'icon': names[key]['icon'] if key and key in names else None,
            'at': at, 'when': when[:1].upper() + when[1:] if when else '', 'extra': extra, 'notes': clean_note(notes),
            'race': rr, 'otherRace': bool(rr and race and rr != race and not (isinstance(rr, list))),
            # A tagged press (a queued Heroic Strike or Raptor Strike) counts on its spell's row, where
            # build_data folds it.
            'casts': round(casts.get((bid, key), casts.get((bid, re.sub(r't\d+$', '', key)), 0)), 1) if key else None,
            'cond': json.dumps(cond, separators=(',', ':')) if cond else None,
        })

    for i, p in enumerate(prepull_sorted(rot)):
        at = secs(p.get('doAtValue', {}).get('const', {}).get('val', '0'))
        row('prepull', i, p['action'], at=at, notes=p.get('notes', ''))
    for i, p in enumerate(rot.get('priorityList') or []):
        if p.get('hide'):
            continue
        row('priority', i + 1, p['action'], notes=p.get('notes', ''))
    return rows


def prepull_sorted(rot):
    return sorted(rot.get('prepullActions') or [], key=lambda p: secs(p.get('doAtValue', {}).get('const', {}).get('val', '0')) or 0)


def _humaniser(nm, cls=None, talents=None):
    return Humaniser(lambda k: 'the potion' if k == 'oOtherActionPotion' else nm(k), lambda aura: talent_token(cls, talents, aura))


def _rotation(req_dir, bid):
    rq = json.load(open(os.path.join(req_dir, bid + '.json')))
    return rq['raid']['parties'][0]['players'][0].get('rotation') or {}


def build(req_dir, outs, nm, names, abilities, racial_race, talents=None):
    """talents: load_talents(), to name the talents conditions check."""
    rows, fallbacks = [], []
    casts = {(a['build'], a['key']): a['casts'] for a in abilities}
    for o in outs:
        bid = o['id']
        H = _humaniser(nm, o['meta'].get('class'), talents)
        rows += apl_rows(bid, _rotation(req_dir, bid), o['meta'].get('race'), nm, names, casts, racial_race, H)
        fallbacks += H.fallbacks
    return rows, sorted(set(fallbacks))


def _sig(p):
    """An APL line as the sim reads it (its notes and the sim UI's hide flag change nothing)."""
    return json.dumps({k: v for k, v in p.items() if k not in ('notes', 'hide')}, sort_keys=True)


def _act_sig(p):
    """What an APL line does, without when: the action kind and its spell, item or other ref."""
    act = p.get('action') or {}
    kind = next((k for k in act if k != 'condition'), None)
    body = act.get(kind) or {}
    return json.dumps([kind, body.get('spellId') if isinstance(body, dict) else None], sort_keys=True)


def diff_lines(base, mine):
    """Pairs a race's APL lines with the preset's: [(base index or None, race index or None, 'same' | 'changed' |
    'added' | 'removed')]. Lines that read the same pair up in order; in a stretch that differs, a race line pairs
    with the preset line doing the same thing (the same spell, a new condition or pre-pull time), and the rest are
    added or removed. A line that only moved is removed at its old place and added at its new one."""
    import difflib
    a, b = [_sig(p) for p in base], [_sig(p) for p in mine]
    out = []
    for op, a0, a1, b0, b1 in difflib.SequenceMatcher(None, a, b, autojunk=False).get_opcodes():
        if op == 'equal':
            out += [(a0 + k, b0 + k, 'same') for k in range(a1 - a0)]
            continue
        left = list(range(a0, a1))
        for j in range(b0, b1):
            i = next((i for i in left if _act_sig(base[i]) == _act_sig(mine[j])), None)
            if i is None:
                out.append((None, j, 'added'))
                continue
            # the preset lines before the pair that nothing pairs with are removed there
            while left and left[0] != i:
                out.append((left.pop(0), None, 'removed'))
            left.pop(0)
            out.append((i, j, 'changed'))
        out += [(i, None, 'removed') for i in left]
    return out


def race_build(req_dir, outs, nm, names, racial_race, preset_rows, race_name, load_override, talents=None):
    """The races that play their own rotation on a preset (race_override.py), as rows of the preset's shape plus
    forRace (the race), diff ('same', 'changed', 'added' or 'removed', against the preset's line) and base (the
    preset line's order, None for an added line) and, on a line that differs, why (race_text.json). A removed row
    is the preset's line, with order None. Lines the race plays as the preset does are the preset's rows (its notes,
    potion name and icon). Casts are left out: the race sweep records DPS only."""
    from race_override import text
    rows, fallbacks = [], []
    pre = {}
    for r in preset_rows:
        pre.setdefault(r['build'], []).append(r)
    for o in outs:
        bid = o['id']
        base = _rotation(req_dir, bid)
        for rc in o.get('races') or []:
            ov = load_override(bid, rc['race'])
            if not ov or 'rotation' not in ov:
                continue
            race = race_name(rc['race'])
            mine = ov['rotation']
            whys = text(ov).get('why', {})
            H = _humaniser(nm, o['meta'].get('class'), talents)
            new = {(r['section'], r['order']): r for r in apl_rows(bid, mine, race, nm, names, {}, racial_race, H)}
            fallbacks += H.fallbacks
            old = {(r['section'], r['order']): r for r in pre.get(bid, [])}
            for section, lb, lm, num in (('prepull', prepull_sorted(base), prepull_sorted(mine), lambda i: i),
                                         ('priority', base.get('priorityList') or [], mine.get('priorityList') or [], lambda i: i + 1)):
                for i, j, d in diff_lines(lb, lm):
                    if (j is not None and lm[j].get('hide')) or (j is None and lb[i].get('hide')):
                        continue
                    if d == 'removed':
                        src = old.get((section, num(i)))
                    elif d == 'same' and old.get((section, num(i))):
                        src = dict(old[(section, num(i))], order=num(j))
                    else:
                        src = new.get((section, num(j)))
                    if not src:
                        continue
                    r = dict(src, forRace=race, diff=d, base=num(i) if i is not None else None, casts=None,
                             why=whys.get(f"{section}:{src['key'] or src['name']}", '') if d != 'same' else '')
                    if d == 'removed':
                        r['order'] = None
                    r['otherRace'] = bool(r['race'] and r['race'] != race)
                    rows.append(r)
    return rows, sorted(set(fallbacks))
