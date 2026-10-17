#!/usr/bin/env python3
"""Spell tooltips, cost lines and icon names from the client's DB2 tables.

    python3 tools/sources/spelltext.py <build> <spell id> [...]

prints each spell's name, icon, cost line and filled description, with the
notes. Other scripts import it:

    spelltext.description(spell_id, build, points=None, notes=None, level=60)
    spelltext.cost(spell_id, build, level=60)
    spelltext.icon(spell_id, build)
    spelltext.icon_file(file_data_id, build)
    spelltext.name(spell_id, build)
    spelltext.passive(spell_id, build)
    spelltext.requirements(spell_id, build)

A spell's Description_lang in the Spell table is a template. "$s1" is the
first effect's value, "$d" the duration, "$12345s1" an effect of another
spell, "${$m1/-1000}.1" an expression rounded to one decimal, and so on. The
game fills it in for the player looking at it. We fill it in the same way
from SpellEffect and the other spell tables, for a character of the given
level with no spell power, attack power or other bonuses.

`points` is the hook for talents whose ranks share one spell. The trait
tables give each rank's value of an effect through a curve, and we take
{effect index: value} for the rank we want in place of SpellEffect's base
points. The racial and other spells pass nothing.

Some things the game knows and we don't: whether the player has a spell or
an aura ("$?s12345[...][...]"), their gear and stats. A conditional takes the
branch for a character who has neither the spell nor the aura, which is what
a tooltip shows before the talent is learned, and stats count as zero. A
construct we can't fill at all shows as "?" in the text, and the notes list
it, so nobody mistakes a guess for the client's number. `notes` collects
those sentences, plus "Values shown at level 60." when a value scales with
level and "Excludes spell power, attack power and other bonuses." when one
has a bonus coefficient.

Every table we read is the 'spelltext' set in wago.py.
"""
import ast
import collections
import functools
import math
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import wago  # noqa: E402

# The tables we read, kept in wago.py so `wago.py fetch <build> --set spelltext` gets them.
TABLES = wago.SETS['spelltext']

LEVEL = 60

LEVEL_NOTE = 'Values shown at level %d.'
BONUS_NOTE = 'Excludes spell power, attack power and other bonuses.'

# PowerType in SpellPower. Rage is stored in tenths.
POWER = {0: 'Mana', 1: 'Rage', 2: 'Focus', 3: 'Energy', 4: 'Combo Points', 5: 'Runes', 6: 'Runic Power'}

# SpellMisc attribute bits we read.
ATTR0_PASSIVE = 0x40
ATTR1_CHANNELED = 0x4 | 0x40
ATTR2_NOT_NEED_SHAPESHIFT = 0x80000

# The effects whose value spell power or attack power adds to, when the effect has a
# bonus coefficient: school damage, health leech, heal, weapon damage (31, 58, 121, 17),
# and auras of periodic damage, periodic heal, periodic leech, damage shield and absorb.
# Many other effects carry a coefficient of 1 that does nothing (energize, dummy auras).
DAMAGE_EFFECTS = {'2', '9', '10', '17', '31', '58', '121'}
AURA_EFFECTS = {'6', '27', '35', '65'}
DAMAGE_AURAS = {'3', '8', '15', '53', '69'}

# The stat placeholders of the expression language. We describe a character
# with no bonuses, so each one counts as zero and the tooltip says so.
STATS = {'sp', 'sps', 'spfi', 'spfr', 'spn', 'sph', 'spa', 'spfrost', 'ap', 'rap', 'bh', 'healpower',
         'damagepower', 'mws', 'mwb', 'mw', 'rwb', 'rws', 'rw'}


def unresolved(note):
	"""Whether a note description() wrote is about a construct it left as "?".

	It writes only those and the two caveats about level and bonuses.
	"""
	return note != BONUS_NOTE and not note.startswith('Values shown at level')


class Unresolved(Exception):
	"""A template construct we can't fill. The message is the note."""


def _num(x):
	try:
		return float(x)
	except (TypeError, ValueError):
		return 0.0


class _Tables:
	"""Every spell table of one build, read once and indexed by spell."""

	def __init__(self, build):
		self.build = build
		self.version = wago.version(build)

	def _rows(self, table):
		try:
			return wago.rows(table, self.build)
		except OSError:
			return []

	def _by_spell(self, table):
		# Some tables have a row per difficulty. DifficultyID 0 is the normal one.
		d = {}
		for r in self._rows(table):
			if r.get('DifficultyID', '0') != '0' and r['SpellID'] in d:
				continue
			d[r['SpellID']] = r
		return d

	@functools.cached_property
	def names(self):
		return {r['ID']: r['Name_lang'] for r in self._rows('SpellName')}

	@functools.cached_property
	def spells(self):
		return {r['ID']: r for r in self._rows('Spell')}

	@functools.cached_property
	def effects(self):
		d = collections.defaultdict(dict)
		for r in self._rows('SpellEffect'):
			i = int(r['EffectIndex'])
			if r['DifficultyID'] != '0' and i in d[r['SpellID']]:
				continue
			d[r['SpellID']][i] = r
		return d

	@functools.cached_property
	def misc(self):
		return self._by_spell('SpellMisc')

	@functools.cached_property
	def levels(self):
		return self._by_spell('SpellLevels')

	@functools.cached_property
	def cooldowns(self):
		return self._by_spell('SpellCooldowns')

	@functools.cached_property
	def aura_options(self):
		return self._by_spell('SpellAuraOptions')

	@functools.cached_property
	def target_restrictions(self):
		return self._by_spell('SpellTargetRestrictions')

	@functools.cached_property
	def shapeshift(self):
		return self._by_spell('SpellShapeshift')

	@functools.cached_property
	def equipped(self):
		return self._by_spell('SpellEquippedItems')

	@functools.cached_property
	def power(self):
		d = collections.defaultdict(list)
		for r in self._rows('SpellPower'):
			d[r['SpellID']].append(r)
		for v in d.values():
			v.sort(key=lambda r: int(r['OrderIndex']))
		return d

	@functools.cached_property
	def durations(self):
		return {r['ID']: r for r in self._rows('SpellDuration')}

	@functools.cached_property
	def ranges(self):
		return {r['ID']: r for r in self._rows('SpellRange')}

	@functools.cached_property
	def radii(self):
		return {r['ID']: r for r in self._rows('SpellRadius')}

	@functools.cached_property
	def cast_times(self):
		return {r['ID']: r for r in self._rows('SpellCastTimes')}

	@functools.cached_property
	def forms(self):
		return {r['ID']: r for r in self._rows('SpellShapeshiftForm')}

	@functools.cached_property
	def subclasses(self):
		return {(r['ClassID'], r['SubClassID']): r for r in self._rows('ItemSubClass')}

	@functools.cached_property
	def variables(self):
		text = {r['ID']: r['Variables'] for r in self._rows('SpellDescriptionVariables')}
		return {r['SpellID']: text.get(r['SpellDescriptionVariablesID'], '')
		        for r in self._rows('SpellXDescriptionVariables')}

	@functools.cached_property
	def icons(self):
		return {r['ID']: r['FileName'] for r in self._rows('ManifestInterfaceData')
		        if r['FilePath'].lower().rstrip('\\') == 'interface\\icons'}


@functools.lru_cache(maxsize=None)
def tables(build):
	return _Tables(build)


def fmt(x, decimals=None):
	"""A number the way a tooltip prints it: whole when it is, else up to two decimals."""
	if decimals is not None:
		s = '%.*f' % (decimals, x + (5e-9 if x >= 0 else -5e-9))
		return s
	if abs(x - round(x)) < 1e-6:
		return str(int(round(x)))
	return ('%.2f' % x).rstrip('0').rstrip('.')


def round_half_up(x):
	return math.floor(x + 0.5) if x >= 0 else -math.floor(-x + 0.5)


def fmt_time(ms):
	"""A duration the way a tooltip prints it: 12 sec, 1.5 sec, 2 min, 1 hour."""
	sec = ms / 1000.0
	if sec >= 3600 and sec % 3600 == 0:
		n = int(sec // 3600)
		return '%d hour%s' % (n, '' if n == 1 else 's')
	if sec >= 60 and sec % 60 == 0:
		return '%d min' % (sec // 60)
	if sec > 60 and sec % 30 == 0:
		return '%s min' % fmt(sec / 60)
	return '%s sec' % fmt(sec)


def name(spell_id, build):
	return tables(build).names.get(str(spell_id), '')


def passive(spell_id, build):
	m = tables(build).misc.get(str(spell_id))
	return bool(m and int(m['Attributes_0']) & ATTR0_PASSIVE)


def icon_file(file_data_id, build):
	"""The lowercase icon name of a FileDataID, like spell_nature_wispsplode, or None."""
	f = tables(build).icons.get(str(file_data_id))
	if not f:
		return None
	f = f.lower()
	f = f[:-4] if f.endswith('.blp') else f
	# Spell_Shadow_DevouringPlague..blp has a stray dot.
	return f.rstrip('.')


def icon(spell_id, build):
	m = tables(build).misc.get(str(spell_id))
	return icon_file(m['SpellIconFileDataID'], build) if m else None


class _Context:
	"""The state of one description being filled: which spell, the rank's points, the notes."""

	def __init__(self, build, spell_id, points, level, notes, depth=0):
		self.t = tables(build)
		self.build = build
		self.spell = str(spell_id)
		self.points = points or {}
		self.level = level
		self.notes = notes
		self.depth = depth
		self.last_number = None
		self._vars = None

	def note(self, text):
		if text not in self.notes:
			self.notes.append(text)

	# Effect values.

	def effect(self, spell, index):
		e = self.t.effects.get(spell, {}).get(index)
		if e is None:
			has = sorted(self.t.effects.get(spell, {}))
			raise Unresolved('The tooltip reads effect %d of spell %s (%s), which has %s in the client tables.' % (
				index + 1, spell, self.t.names.get(spell, '?'),
				'effect%s %s' % ('' if len(has) == 1 else 's', ', '.join(str(i + 1) for i in has)) if has else 'no effects'))
		return e

	def base(self, spell, index):
		"""An effect's base points, or the rank's value when a talent rank sets it."""
		if spell == self.spell and index in self.points:
			return float(self.points[index])
		return _num(self.effect(spell, index)['EffectBasePointsF'])

	def value(self, spell, index):
		"""An effect's value at our level, signed, before any variance."""
		e = self.effect(spell, index)
		v = self.base(spell, index)
		per = _num(e['EffectRealPointsPerLevel'])
		if per:
			lv = self.t.levels.get(spell)
			if lv:
				level = self.level
				top = int(lv['MaxLevel'])
				if top and level > top:
					level = top
				level = max(level, int(lv['BaseLevel']))
				gain = per * (level - int(lv['SpellLevel']))
				if gain:
					v += gain
					self.note(LEVEL_NOTE % self.level)
		if _num(e['BonusCoefficientFromAP']) or (_num(e['EffectBonusCoefficient']) and (
				e['Effect'] in DAMAGE_EFFECTS or (e['Effect'] in AURA_EFFECTS and e['EffectAura'] in DAMAGE_AURAS))):
			self.note(BONUS_NOTE)
		return v

	def spread(self, spell, index):
		"""(min, max) of an effect's value, rounded the way the tooltip shows them.

		The variance is a share of the base points, not of the value at our level. The
		classic spells it replaced rolled a die of fixed size on top of a value that grew
		with level, so the width of the range doesn't grow. Riptide is 480 base points,
		3.7 more a level for 8 levels and a variance of 0.1: 509.6 plus or minus 24, which
		shows as 486 to 534.
		"""
		v = self.value(spell, index)
		var = _num(self.effect(spell, index)['Variance'])
		if var and v:
			half = self.base(spell, index) * var / 2
			return round_half_up(v - half), round_half_up(v + half)
		return v, v

	def duration(self, spell):
		m = self.t.misc.get(spell)
		d = self.t.durations.get(m['DurationIndex']) if m else None
		if not d:
			raise Unresolved('The tooltip reads the duration of spell %s (%s), which has none in the client tables.' % (
				spell, self.t.names.get(spell, '?')))
		return int(d['Duration'])

	def period(self, spell, index):
		p = int(_num(self.effect(spell, index)['EffectAuraPeriod']))
		if not p:
			raise Unresolved('Effect %d of spell %s has no period.' % (index + 1, spell))
		return p

	def radius(self, spell, index, which=0):
		e = self.effect(spell, index)
		# Many effects have only their second radius set (Blast Wave, War Stomp).
		r = self.t.radii.get(e['EffectRadiusIndex_%d' % which]) or self.t.radii.get(e['EffectRadiusIndex_%d' % (1 - which)])
		if not r:
			raise Unresolved('Effect %d of spell %s has no radius.' % (index + 1, spell))
		return _num(r['Radius'])

	def aura_option(self, spell, field):
		a = self.t.aura_options.get(spell)
		if not a:
			raise Unresolved('Spell %s has no %s.' % (spell, field))
		return int(a[field])

	def ref(self, spell, letter, index):
		"""A $<spell><letter><index> reference as a signed number, for expressions."""
		spell = spell or self.spell
		i = (index or 1) - 1
		if letter in ('s', 'm', 'S', 'w'):
			return self.value(spell, i)
		if letter == 'M':
			return self.spread(spell, i)[1]
		if letter == 'o':
			return self.value(spell, i) * (self.duration(spell) // self.period(spell, i))
		if letter == 't':
			return self.period(spell, i) / 1000.0
		if letter == 'd' or letter == 'D':
			return self.duration(spell) / 1000.0
		if letter == 'a':
			return self.radius(spell, i)
		if letter == 'A':
			return self.radius(spell, i, 1)
		if letter == 'x':
			return int(self.effect(spell, i)['EffectChainTargets'])
		if letter == 'e':
			return _num(self.effect(spell, i)['EffectAmplitude'])
		if letter == 'b':
			return _num(self.effect(spell, i)['EffectPointsPerResource'])
		if letter == 'q':
			return int(self.effect(spell, i)['EffectMiscValue_0'])
		if letter in ('h', 'H'):
			return self.aura_option(spell, 'ProcChance')
		if letter == 'n':
			return self.aura_option(spell, 'ProcCharges')
		if letter in ('u', 'U'):
			return self.aura_option(spell, 'CumulativeAura')
		if letter == 'proccooldown':
			return self.aura_option(spell, 'ProcCategoryRecovery') / 1000.0
		if letter == 'i':
			r = self.t.target_restrictions.get(spell)
			if r and int(r['MaxTargets']):
				return int(r['MaxTargets'])
			raise Unresolved('Spell %s has no target limit.' % spell)
		if letter in ('r', 'R'):
			m = self.t.misc.get(spell)
			r = self.t.ranges.get(m['RangeIndex']) if m else None
			if r:
				return _num(r['RangeMax_0'])
			raise Unresolved('Spell %s has no range.' % spell)
		if letter in ('pl', 'PL'):
			self.note(LEVEL_NOTE % self.level)
			return self.level
		if letter.lower() in STATS:
			self.note(BONUS_NOTE)
			return 0
		if letter == 'bc':
			return _num(self.effect(spell, 0)['EffectBonusCoefficient'])
		raise Unresolved('"$%s%s" (spell %s) is a placeholder we can\'t fill.' % (letter, index or '', spell))

	def text_ref(self, spell, letter, index):
		"""A $<spell><letter><index> reference as the tooltip prints it."""
		spell = spell or self.spell
		i = (index or 1) - 1
		if letter in ('s', 'S'):
			lo, hi = self.spread(spell, i)
			if lo != hi:
				self.last_number = hi
				return '%s to %s' % (fmt(abs(lo)), fmt(abs(hi)))
			v = self.shown(spell, i, abs(lo))
			self.last_number = v
			return fmt(v)
		if letter == 'c':
			# The power cost, as the cost line gives it: "55% of base" or "120" before the
			# power's name, which the tooltip text says itself.
			for p in self.t.power.get(spell, []):
				amount, pct = int(p['ManaCost']), _num(p['PowerCostPct'])
				if amount:
					self.last_number = amount // 10 if p['PowerType'] == '1' else amount
					return fmt(self.last_number)
				if pct:
					self.last_number = pct
					return '%s%% of base' % fmt(pct)
			raise Unresolved('Spell %s has no cost for "$c".' % spell)
		if letter in ('d', 'D'):
			ms = self.duration(spell)
			# A duration of -1 is an aura that stays until it's cancelled (Find Treasure).
			if ms < 0:
				return 'until cancelled'
			self.last_number = ms / 1000.0
			return fmt_time(ms)
		v = self.ref(spell, letter, index)
		if letter in ('m', 'M', 'o', 'w'):
			v = self.shown(spell, i, abs(v))
		self.last_number = v
		return fmt(v)

	def shown(self, spell, index, v):
		"""An effect value as the tooltip prints it.

		Whole base points stay whole when the level adds a fraction (Penance's 42 base
		points plus 0.18 a level show as 45, not 44.61). Base points that have a fraction
		themselves keep it.
		"""
		base = self.base(spell, index)
		return round_half_up(v) if abs(base - round(base)) < 1e-6 else v

	# Description variables ($<name>), from SpellDescriptionVariables.

	def var(self, key):
		if self._vars is None:
			self._vars = parse_variables(self.t.variables.get(self.spell, ''))
		body = self._vars.get(key.lower())
		if body is None:
			raise Unresolved('"$<%s>" isn\'t a description variable of spell %s.' % (key, self.spell))
		text = render(self, body)
		try:
			return float(text)
		except ValueError:
			raise Unresolved('"$<%s>" of spell %s doesn\'t come out as a number.' % (key, self.spell))


def parse_variables(text):
	"""{name: body} from a SpellDescriptionVariables text.

	Most put one "$name=body" on each line, but some run them together on one line
	("$sk=${...}$sc=${$PL*5}"), so a new variable starts at any "$name=" outside brackets.
	"""
	found = {}
	depth = 0
	name, start = None, 0
	i = 0
	while i < len(text):
		c = text[i]
		if c in '{[(':
			depth += 1
		elif c in '}])':
			depth -= 1
		elif depth <= 0 and c == '$':
			m = re.match(r'\$(\w+)\s*=', text[i:])
			if m:
				if name:
					found[name] = text[start:i].strip()
				name, start = m.group(1).lower(), i + m.end()
				i += m.end()
				continue
		i += 1
	if name:
		found[name] = text[start:].strip()
	return found


# The expression language inside ${...}: arithmetic plus a few functions. We
# turn it into a Python expression of plain numbers and evaluate that tree by
# hand, so nothing in a downloaded table is ever run.

FUNCS = {
	'cond': lambda c, a, b: a if c else b,
	'gt': lambda a, b: a > b, 'lt': lambda a, b: a < b, 'gte': lambda a, b: a >= b,
	'lte': lambda a, b: a <= b, 'eq': lambda a, b: a == b,
	'max': max, 'min': min, 'floor': math.floor, 'ceil': math.ceil, 'abs': abs,
	'round': round_half_up, 'clamp': lambda x, lo, hi: max(lo, min(hi, x)),
}


def _eval(node):
	if isinstance(node, ast.Expression):
		return _eval(node.body)
	if isinstance(node, ast.Constant) and isinstance(node.value, (int, float)):
		return node.value
	if isinstance(node, ast.UnaryOp) and isinstance(node.op, (ast.USub, ast.UAdd)):
		v = _eval(node.operand)
		return -v if isinstance(node.op, ast.USub) else v
	if isinstance(node, ast.BinOp):
		a, b = _eval(node.left), _eval(node.right)
		if isinstance(node.op, ast.Add):
			return a + b
		if isinstance(node.op, ast.Sub):
			return a - b
		if isinstance(node.op, ast.Mult):
			return a * b
		if isinstance(node.op, ast.Div):
			return a / b if b else 0
	if isinstance(node, ast.Call) and isinstance(node.func, ast.Name) and node.func.id in FUNCS and not node.keywords:
		return FUNCS[node.func.id](*[_eval(a) for a in node.args])
	raise ValueError('unsupported')


REF = re.compile(r'\$(\d*)(proccooldown|[a-zA-Z]{2,}|[a-zA-Z])(\d?)')
NAMES = ('proccooldown', 'pl', 'PL', 'bc')


def read_ref(s):
	"""The reference at the start of s, like $s1, $12345d or $proccooldown.

	Answers (spell, letter, index, length), or None. A one letter reference can run into
	the word after it ("$ssec"), so a run of letters that isn't one of our names is read as
	its first letter.
	"""
	m = REF.match(s)
	if not m:
		return None
	letter = m.group(2)
	if len(letter) > 1 and letter not in NAMES and letter.lower() not in STATS:
		m = re.match(r'\$(\d*)([a-zA-Z])(\d?)', s)
	return m.group(1), m.group(2), int(m.group(3)) if m.group(3) else None, m.end()


def expression(ctx, body):
	"""The value of an expression body (the text inside ${...})."""
	out = []
	i = 0
	while i < len(body):
		c = body[i]
		if c == '$':
			if body.startswith('$?', i):
				j, text = conditional(ctx, body, i)
				out.append('(%s)' % expression(ctx, text) if text.strip() else '0')
				i = j
				continue
			if body.startswith('${', i):
				j = matching(body, i + 1, '{', '}')
				out.append('(%r)' % expression(ctx, body[i + 2:j]))
				i = j + 1
				continue
			if body.startswith('$<', i):
				j = body.index('>', i)
				out.append('(%r)' % ctx.var(body[i + 2:j]))
				i = j + 1
				continue
			m = re.match(r'\$(cond|gt|lt|gte|lte|eq|max|min|floor|ceil|abs|round|clamp)\(', body[i:])
			if m:
				out.append(m.group(1) + '(')
				i += m.end()
				continue
			r = read_ref(body[i:])
			if not r:
				raise Unresolved('"${%s}" isn\'t an expression we can read.' % body)
			spell, letter, index, length = r
			i += length
			out.append('(%r)' % float(ctx.ref(spell, letter, index)))
			continue
		out.append(c)
		i += 1
	src = ''.join(out).strip()
	if not src:
		return 0.0
	# A few expressions close one parenthesis more than they open (Horn of Lordaeron's
	# chain of $cond), which the game reads past.
	while src.count(')') > src.count('(') and src.endswith(')'):
		src = src[:-1]
	try:
		return float(_eval(ast.parse(src, mode='eval')))
	except (SyntaxError, ValueError, TypeError, ZeroDivisionError, RecursionError):
		raise Unresolved('"${%s}" isn\'t an expression we can read.' % body)


def matching(s, i, open_, close):
	"""The index of the bracket that closes the one at s[i]."""
	depth = 0
	for j in range(i, len(s)):
		if s[j] == open_:
			depth += 1
		elif s[j] == close:
			depth -= 1
			if depth == 0:
				return j
	raise Unresolved('An unclosed "%s" in "%s".' % (open_, s[i:i + 60]))


def conditional(ctx, s, i):
	"""Read $?cond[a]?cond[b][c] at s[i]. Answers (end, the branch taken).

	Every condition asks whether the player has a spell or an aura (or is a
	class). We describe a character who has none of them yet, so each one is
	false and we take the last branch.
	"""
	j = i + 1
	branches = []
	while j < len(s) and s[j] == '?':
		k = s.index('[', j)
		end = matching(s, k, '[', ']')
		branches.append(s[k + 1:end])
		j = end + 1
		if j < len(s) and s[j] == '[':
			end = matching(s, j, '[', ']')
			branches.append(s[j + 1:end])
			return end + 1, branches[-1]
	return j, ''


def render(ctx, text):
	"""Fill a template, as text."""
	out = []
	i = 0
	n = len(text)
	while i < n:
		c = text[i]
		if c != '$':
			if c == '|' and i + 1 < n and text[i + 1] in 'cC' and re.match(r'\|[cC][0-9a-fA-F]{8}', text[i:]):
				i += 10
				continue
			if c == '|' and i + 1 < n and text[i + 1] in 'rR':
				i += 2
				continue
			out.append(c)
			i += 1
			continue
		try:
			i = _construct(ctx, text, i, out)
		except Unresolved as e:
			ctx.note(str(e))
			out.append('?')
			i = _skip(text, i)
	return ''.join(out)


def _skip(text, i):
	"""Past the construct at text[i] that we couldn't fill."""
	if text.startswith('${', i):
		try:
			j = matching(text, i + 1, '{', '}') + 1
		except Unresolved:
			return len(text)
		m = re.match(r'\.\d', text[j:])
		return j + (2 if m else 0)
	if text.startswith('$<', i):
		return text.index('>', i) + 1 if '>' in text[i:] else len(text)
	m = re.match(r'\$(/|\*)[\d.]+;\d*[a-zA-Z]\d?', text[i:]) or re.match(r'\$@\w+', text[i:])
	if m:
		return i + m.end()
	r = read_ref(text[i:])
	return i + r[3] if r else i + 1


def _construct(ctx, text, i, out):
	"""Fill the construct that starts with the $ at text[i]. Answers where it ends."""
	rest = text[i:]
	if rest.startswith('$$'):
		out.append('$')
		return i + 2
	if rest.startswith('$?'):
		j, branch = conditional(ctx, text, i)
		out.append(render(ctx, branch))
		return j
	if rest.startswith('${'):
		j = matching(text, i + 1, '{', '}')
		v = expression(ctx, text[i + 2:j])
		j += 1
		m = re.match(r'\.(\d)', text[j:])
		if m:
			v = round(v, int(m.group(1)))
			out.append(fmt(v, int(m.group(1))))
			j += 2
		else:
			v = round_half_up(v)
			out.append(fmt(v))
		ctx.last_number = v
		return j
	if rest.startswith('$<'):
		j = text.index('>', i)
		v = ctx.var(text[i + 2:j])
		out.append(fmt(round_half_up(v)))
		ctx.last_number = v
		return j + 1
	m = re.match(r'\$([/*])([\d.]+);(\d*)([a-zA-Z])(\d?)', rest)
	if m:
		op, k, spell, letter, index = m.groups()
		v = abs(ctx.ref(spell, letter, int(index) if index else None))
		if letter in ('s', 'S'):
			lo, hi = ctx.spread(spell or ctx.spell, (int(index) if index else 1) - 1)
			if lo != hi:
				f = (lambda x: abs(x) / float(k)) if op == '/' else (lambda x: abs(x) * float(k))
				out.append('%s to %s' % (fmt(round_half_up(f(lo))), fmt(round_half_up(f(hi)))))
				ctx.last_number = f(hi)
				return i + m.end()
		v = v / float(k) if op == '/' else v * float(k)
		out.append(fmt(v))
		ctx.last_number = v
		return i + m.end()
	m = re.match(r'\$@(spelldesc|spelltooltip|spellname|spellaura|spellicon)(\d*)', rest)
	if m:
		kind, spell = m.group(1), m.group(2) or ctx.spell
		if kind == 'spellname':
			out.append(ctx.t.names.get(spell, '?'))
		elif kind == 'spellicon':
			pass
		elif ctx.depth > 4:
			raise Unresolved('"$@%s%s" nests too deep.' % (kind, spell))
		else:
			field = 'AuraDescription_lang' if kind == 'spellaura' else 'Description_lang'
			body = ctx.t.spells.get(spell, {}).get(field, '')
			sub = _Context(ctx.build, spell, None, ctx.level, ctx.notes, ctx.depth + 1)
			out.append(render(sub, body))
		return i + m.end()
	m = re.match(r'\$([lLgG])([^:;]*):([^;]*);', rest)
	if m:
		kind, a, b = m.groups()
		if kind in 'gG':
			out.append(a)
		else:
			out.append(a if ctx.last_number is not None and abs(ctx.last_number) == 1 else b)
		return i + m.end()
	r = read_ref(rest)
	if r:
		spell, letter, index, length = r
		out.append(ctx.text_ref(spell, letter, index))
		return i + length
	out.append('$')
	return i + 1


def tidy(text):
	"""One space between words, the client's line breaks kept."""
	text = text.replace('\r\n', '\n').replace('\r', '\n')
	text = re.sub(r'[ \t]+', ' ', text)
	text = re.sub(r' *\n *', '\n', text)
	return text.strip()


def description(spell_id, build, points=None, notes=None, level=LEVEL):
	"""A spell's tooltip text with every number filled in.

	points is {effect index: value} for the effects whose value a talent rank
	sets (the trait tables' curves). notes, when given, collects a sentence for
	every construct left as "?" and for the level and bonus caveats.
	"""
	notes = notes if notes is not None else []
	t = tables(build)
	body = t.spells.get(str(spell_id), {}).get('Description_lang', '')
	ctx = _Context(build, spell_id, points, level, notes)
	return tidy(render(ctx, body))


def cost(spell_id, build, level=LEVEL):
	"""The line under an active spell's name: '135 Mana; 30 yd range; 2.5 sec cast; 10 sec cooldown'.

	Empty for a passive spell. A cost given as a share of base mana stays a
	share, since base mana isn't in the client tables.
	"""
	s = str(spell_id)
	t = tables(build)
	if passive(s, build):
		return ''
	parts = []
	for p in t.power.get(s, []):
		kind = int(p['PowerType'])
		amount = int(p['ManaCost'])
		pct = _num(p['PowerCostPct'])
		label = POWER.get(kind, 'Power')
		if kind == 1:
			amount //= 10
		if kind == 4 and amount:
			extra = int(p['OptionalCost'])
			parts.append('%d–%d %s' % (amount, amount + extra, label) if extra else '%d %s' % (amount, label))
		elif amount:
			parts.append('%d %s' % (amount, label))
		elif pct:
			parts.append('%s%% base %s' % (fmt(pct), label))
	m = t.misc.get(s)
	if m:
		r = t.ranges.get(m['RangeIndex'])
		if r:
			lo, hi, friend = _num(r['RangeMin_0']), _num(r['RangeMax_0']), _num(r['RangeMax_1'])
			if int(r['Flags']) & 1:
				parts.append('Melee Range')
			elif hi and friend and hi != friend:
				parts.append('%s yd hostile / %s yd friendly' % (fmt(hi), fmt(friend)))
			elif hi or friend:
				parts.append(('%s–%s yd range' % (fmt(lo), fmt(hi or friend))) if lo else '%s yd range' % fmt(hi or friend))
		if int(m['Attributes_1']) & ATTR1_CHANNELED:
			parts.append('Channeled')
		else:
			ct = t.cast_times.get(m['CastingTimeIndex'])
			ms = int(ct['Base']) if ct else 0
			parts.append('%s cast' % fmt_time(ms) if ms > 0 else 'Instant')
	cd = t.cooldowns.get(s)
	if cd:
		ms = max(int(cd['RecoveryTime']), int(cd['CategoryRecoveryTime']))
		if ms:
			parts.append('%s cooldown' % fmt_time(ms))
	return '; '.join(parts)


def requirements(spell_id, build):
	"""What the spell needs equipped or which form, like 'Requires Cat Form', or None."""
	s = str(spell_id)
	t = tables(build)
	needs = []
	sh = t.shapeshift.get(s)
	m = t.misc.get(s)
	# A spell that can also be cast out of the forms in its mask doesn't say "Requires".
	if sh and not (m and int(m['Attributes_2']) & ATTR2_NOT_NEED_SHAPESHIFT):
		mask = int(sh['ShapeshiftMask_0']) | (int(sh['ShapeshiftMask_1']) << 32)
		for bit in range(64):
			if mask & (1 << bit):
				f = t.forms.get(str(bit + 1))
				if f:
					needs.append(f['Name_lang'])
	eq = t.equipped.get(s)
	if eq and eq['EquippedItemClass'] in ('2', '4'):
		cls, mask = eq['EquippedItemClass'], int(eq['EquippedItemSubclass'])
		if cls == '2' and mask and (mask & MELEE) == MELEE:
			needs.append('Melee Weapon')
		else:
			subs = [t.subclasses.get((cls, str(bit))) for bit in range(32) if mask & (1 << bit)]
			subs = [x for x in subs if x]
			# One kind of weapon reads in the plural ("Requires Daggers"), a list in the
			# singular ("Requires Bow, Gun, Crossbow"), and armor always in the singular
			# ("Requires Shield").
			if len(subs) == 1 and cls == '2':
				needs.append(subs[0]['VerboseName_lang'] or subs[0]['DisplayName_lang'])
			else:
				needs.extend(x['DisplayName_lang'] for x in subs)
	return 'Requires ' + ', '.join(needs) if needs else None


# Every melee weapon subclass: axes, maces, swords (both kinds), polearm,
# staff, fist weapon, dagger. A spell that takes all of them says "Melee Weapon".
MELEE = (1 << 0) | (1 << 1) | (1 << 4) | (1 << 5) | (1 << 6) | (1 << 7) | (1 << 8) | (1 << 10) | (1 << 13) | (1 << 15)


def main():
	if len(sys.argv) < 3:
		print(__doc__)
		return 1
	build = sys.argv[1]
	for s in sys.argv[2:]:
		notes = []
		print('%s %s [%s]' % (s, name(s, build), icon(s, build)))
		c = cost(s, build)
		if c:
			print('  ' + c)
		r = requirements(s, build)
		if r:
			print('  ' + r)
		print('  ' + description(s, build, notes=notes).replace('\n', '\n  '))
		for n in notes:
			print('  note: ' + n)
	return 0


if __name__ == '__main__':
	sys.exit(main())
