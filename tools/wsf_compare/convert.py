#!/usr/bin/env python3
"""Turns one of our requests (a DPS Report request, tools/dps_report/req/<id>.json) into a matched pair:

- <out>/req_ours.json: ours, with the few things theirs can't model made the same on both sides (see ALIGNED_OURS);
- <out>/req_theirs.json: the same character, gear, buffs, debuffs, encounter and rotation as a wowsims/forever
  RaidSimRequest;
- <out>/items.json: a wowsims/forever SimDatabase of our synthetic items, which theirs doesn't have (cmpsim and
  cmpstats load it).

    convert.py <req.json> <out dir> [--iters N] [--apl ours|theirs:<build>|<apl.json>] [--turret] [--keep-filler]

--apl ours (default): both run our rotation, translated for theirs (APL_TRANSLATION below).
--apl theirs:<name>: theirs runs its own rotation, from $WSF_WORK/theirs_builds/<name>.json, a build or a bare APL
  (setup.sh dumps them: <class>_<spec>_<build> and apl_<class>_<spec>_<apl>).
--apl <file>: both run that APL.
--turret: drop our rotation's lines whose notes start "Weave:" (the hunters' melee weave), so both stand at range.
--their-mob-type <MobType...>: give theirs' boss a creature type. Ours counts the generic boss (no type) as one the
  hunter tracks and theirs doesn't (Improved Tracking, 8aa26ccb4d): MobTypeGiant aligns that and no racial reads it.
--keep-filler: keep our filler explosive. Theirs applies one explosive of a kind, so by default ours drops its
  filler and both use only the Sapper.

Env ALIGN="<their stat index>=<value>,..." adds bonus stats to theirs (align.py finds them).
Env NOSPELLS="<id>,..." names talent spells the build doesn't have, see nospell() below.

Only the classes in CLASSES are mapped. Adding one means its options, its consumables and its spell ids
(SKILL.md, "Adding a class").
"""
import copy, json, os, re, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..'))
WSF = os.environ.get('WSF', os.path.expanduser('~/wowsims-forever'))
WORK = os.environ.get('WSF_WORK', os.path.expanduser('~/wsf_compare'))

args = sys.argv[1:]
def opt(name, default=None):
    return args[args.index(name) + 1] if name in args else default
src, out = args[0], args[1]
APL = opt('--apl', 'ours'); ITERS = opt('--iters')
os.makedirs(out, exist_ok=True)

req = json.load(open(src))
ours_items = {i['id']: i for i in json.load(open(os.path.join(ROOT, 'assets/database/db.json')))['items']}

def enum_names(path, name):
    body = re.search(r'enum %s \{(.*?)\}' % name, open(path).read(), re.S).group(1)
    return {int(v): k for k, v in re.findall(r'(\w+)\s*=\s*(\d+);', body)}
OP = os.path.join(ROOT, 'proto/common.proto'); TP = os.path.join(WSF, 'proto/common.proto')
OS = enum_names(OP, 'Stat'); TS = {v: k for k, v in enum_names(TP, 'Stat').items()}
SMAP = {i: TS[n] for i, n in OS.items() if n in TS}  # our stat index -> theirs, by name
def stats_map(arr, hit=0, crit=0):
    m = {}
    for i, v in enumerate(arr or []):
        if v:
            if i not in SMAP: raise SystemExit(f'unmapped stat {OS.get(i)}={v}')
            m[SMAP[i]] = v
    if hit: m[TS['StatHitRating']] = hit
    if crit: m[TS['StatCritRating']] = crit
    return m
def stats_arr(m, n=len(TS)):
    a = [0.0] * n
    for k, v in m.items(): a[k] += v
    return a
EN = {f: enum_names(OP, e) for f, e in [('type', 'ItemType'), ('armorType', 'ArmorType'), ('weaponType', 'WeaponType'),
      ('handType', 'HandType'), ('rangedWeaponType', 'RangedWeaponType')]}
CLS = enum_names(OP, 'Class')

ours = copy.deepcopy(req)
p = ours['raid']['parties'][0]['players'][0]
klass = p['class']
if ITERS: ours['simOptions']['iterations'] = int(ITERS)
cs = p['consumes']

# ALIGNED_OURS: what ours changes so both model the same fight.
d = ours['raid']['debuffs']
if d.pop('sunderRamp', None) is not None: d['sunderArmor'] = True  # theirs has no ramp: full Sunder at the pull in both
if cs.get('fillerExplosive') and '--keep-filler' not in args:
    print('ours drops its filler explosive', cs.pop('fillerExplosive'), '(theirs applies one explosive)')

# Our synthetic gear (ids 990000 and up) as theirs' items. Their real items come from their own database.
sim_items = []
for it in p['equipment']['items']:
    if not it or it.get('id', 0) < 990000: continue
    i = ours_items[it['id']]
    so = {'stats': {str(k): v for k, v in stats_map(i.get('stats'), i.get('hitRating', 0), i.get('critRating', 0)).items()}, 'ilvl': i['ilvl']}
    if 'weaponDamageMin' in i:
        so['weaponDamageMin'] = i['weaponDamageMin']; so['weaponDamageMax'] = i['weaponDamageMax']
    item = {'id': i['id'], 'name': i['name'], 'weaponSpeed': i.get('weaponSpeed', 0),
            'classAllowlist': [CLS[c] for c in i.get('classAllowlist', [])], 'scalingOptions': {'0': so}}
    for f in EN:
        if i.get(f): item[f] = EN[f][i[f]]
    sim_items.append({k: v for k, v in item.items() if v})
json.dump({'items': sim_items}, open(os.path.join(out, 'items.json'), 'w'), indent=1)

# Enchants: ours keys Forever's spell ids, theirs the enchant (effect) ids.
ENCH = {1249059: 904, 1248497: 904, 1219587: 7667, 1248640: 2563}
equip = [{k: (ENCH.get(v, v) if k == 'enchant' else v) for k, v in it.items()} for it in p['equipment']['items']]

# Consumables: our enum name -> their item id. A new one is looked up in theirs' database by name.
C = {'FlaskOfNaturalAggression': 274274, 'FlaskOfNaturalPrecision': 274275, 'FlaskOfNaturalAccuracy': 274273,
     'FoodFlankAuPoivre': 250069, 'ElixirOfTheMongoose': 13452, 'GreaterMagebloodElixir': 250341,
     'ElixirOfGreaterIntellect': 9179, 'ElixirOfFerocity': 250350, 'JujuMight': 12460, 'GroundScorpokAssay': 8412,
     'MajorManaPotion': 13444, 'MajorFrenzyPotion': 250943, 'ConjuredDemonicRune': 12662,
     'ConjuredRogueThistleTea': 7676, 'SapperGoblinSapper': 10646, 'ExplosiveEZThroDarkBomb': 260817,
     'ExplosiveEZThroThoriumGrenade': 260816, 'ExplosiveDenseDynamite': 18641, 'ExplosiveThoriumGrenade': 15993,
     'DenseSharpeningStone': 12404, 'DenseWeightstone': 12643,
     'DeadlyPoison': 8985, 'InstantPoison': 8928}  # Deadly Poison IV (theirs has no AQ book rank), Instant VI
def c(k):
    v = cs.get(k)
    if not v: return 0
    if v not in C: raise SystemExit(f'no wowsims/forever id for consumable {k}={v}: add it to C')
    return C[v]
def weapon(slot):
    items = p['equipment']['items']
    return ours_items.get(items[slot]['id'], {}) if len(items) > slot and items[slot] else {}
def stone(k, slot):  # theirs drops a sharpening stone on a blunt weapon (ours applies it anyway): a weightstone there
    v = c(k)
    if v == 12404 and EN['weaponType'].get(weapon(slot).get('weaponType')) in ('WeaponTypeMace', 'WeaponTypeStaff', 'WeaponTypeFist'):
        print(k, 'is on a blunt weapon: Dense Weightstone for theirs')
        return 12643
    return v
pot = c('defaultPotion')
consumables = {
    'flaskId': c('flask'), 'foodId': c('food'),
    'elixirIds': [x for x in (c('agilityElixir'), c('strengthBuff'), c('manaRegenElixir'), c('intellectElixir')) if x],
    'extraIds': [x for x in (c('attackPowerBuff'), c('zanzaBuff'), c('fillerExplosive')) if x] + ([12217] if cs.get('dragonBreathChili') else []),
    'mhImbueId': stone('mainHandImbue', 14), 'ohImbueId': stone('offHandImbue', 15),
    'potions': [pot] if pot else [], 'conjuredId': c('defaultConjured'), 'conjuredItems': [c('defaultConjured')] if c('defaultConjured') else [],
    'explosiveId': c('sapperExplosive'),
}
if pot == 13444: consumables['resourcePotId'] = pot
elif pot: consumables['offensivePotId'] = pot

# Raid and party buffs, blessings, debuffs.
b = ours['raid']['buffs']; ib = p['buffs']
raid_buffs = {
    'arcaneBrilliance': b.get('arcaneBrilliance', False), 'prayerOfSpirit': b.get('divineSpirit', False),
    'giftOfTheWild': bool(b.get('giftOfTheWild')), 'prayerOfFortitude': bool(b.get('powerWordFortitude')),
    'prayerOfShadowProtection': b.get('shadowProtection', False), 'fireResistanceAura': b.get('fireResistanceAura', False),
    'frostResistanceAura': b.get('frostResistanceAura', False),
    'strengthOfEarthTotem': bool(b.get('strengthOfEarthTotem')), 'graceOfAirTotem': bool(b.get('graceOfAirTotem')),
    'windfuryTotem': b.get('totemWeaponBuff') == 'TotemWeaponBuffWindfury',
    'battleShout': 1 if b.get('battleShout') else 0,  # 1 is rank 6 (2 rank 7); theirs has no Improved Battle Shout state
    'trueshotAura': b.get('trueshotAura', False), 'leaderOfThePack': b.get('leaderOfThePack', False),
    'moonkinAura': b.get('moonkinAura', False),
}
party_buffs = {'manaSpringTotem': b.get('manaSpringTotem', 'TristateEffectMissing'),  # a party buff in theirs
               'bloodPact': bool(b.get('bloodPact')), 'devotionAura': bool(b.get('devotionAura')),
               'retributionAura': bool(b.get('retributionAura'))}
ind_buffs = {'greaterBlessingOfKings': ib.get('blessingOfKings', False), 'greaterBlessingOfMight': bool(ib.get('blessingOfMight')),
             'greaterBlessingOfWisdom': bool(ib.get('blessingOfWisdom')), 'greaterBlessingOfSalvation': b.get('salvation', False),
             'thorns': bool(b.get('thorns'))}
debuffs = {k: bool(d.get(k)) for k in ['judgementOfWisdom', 'judgementOfLight', 'judgementOfTheCrusader', 'faerieFire',
           'curseOfElements', 'sunderArmor', 'curseOfWeakness', 'curseOfRecklessness', 'demoralizingRoar', 'demoralizingShout',
           'thunderClap', 'insectSwarm', 'scorpidSting', 'exposeArmor']}

# Cooldowns of the spells spellUsesLostByDelay names, which theirs lacks (it is rewritten from remainingTime).
CD = {20572: '120s', 1259799: '180s', ('item', 993016): '120s', ('item', 993117): '120s'}
REMAP = {}  # (spellId, frozenset of the other keys) handled per class in remap()
SET_BONUS = {}  # theirs' stat index -> value for a synthetic set bonus theirs doesn't register

# ---- classes ----------------------------------------------------------------------------------------------------
talents = p['talentsString']
extra_lines = []
if klass == 'ClassRogue':
    ro = p['rogue']['options']
    consumables['mhWeaponImbueId'] = C[ro['mainHandPoison']]; consumables['ohWeaponImbueId'] = C[ro['offHandPoison']]
    class_opts = {'rogue': {'options': {'classOptions': {}}}}
    SET_BONUS[TS['StatHitRating']] = 10  # Grimstitch Armor's 2 piece, 1% hit, on our synthetic set
    def remap(n):
        if n.get('spellId') == 26297:  # Berserking: ours keys the rogue's 26297 (tagged), theirs Forever's 20554
            n = {k: v for k, v in n.items() if k != 'tag'}; n['spellId'] = 20554
        if n.get('spellId') == 1856 and set(n) <= {'spellId', 'rank'}:  # Vanish: ours keys rank 1, theirs rank 2
            n = {'spellId': 1857, 'rank': 2}
        if n.get('spellId') == 1310707: n = {**n, 'spellId': 1241584}  # Mutilate: the cast, theirs its ladder's last rank
        if n.get('spellId') == 25300: n = {**n, 'spellId': 11281}  # Backstab IX is an AQ book: theirs keys rank 8
        if n.get('spellId') == 31016 and len(n) == 1: n = {'spellId': 11300}  # Eviscerate: theirs keys rank VIII
        return n
    # Theirs registers Vanish as a DPS cooldown, so autocastOtherCooldowns would Vanish at the pull; a never-true
    # line naming it keeps it out of the autocast, as ours has no Vanish cooldown.
    extra_lines.append({'action': {'condition': {'cmp': {'op': 'OpLt', 'lhs': {'const': {'val': '1'}}, 'rhs': {'const': {'val': '0'}}}},
                                   'castSpell': {'spellId': {'spellId': 1857, 'rank': 2}}}})
elif klass == 'ClassHunter':
    ho = p['hunter']['options']
    ho.setdefault('petAttackSpeed', 'One')  # the proto default: Faster Attack VII, a 1.0 s swing before Cobra Reflexes
    QUIVER = {'Speed10': 'HuntingQuiver', 'Speed11': 'QuiverOfTheNightWatch', 'Speed12': 'HeavyQuiver', 'Speed13': 'QuickdrawQuiver',
              'Speed15': 'AncientSinewWrappedLamina'}
    SPEED = {'One': 'FasterAttackVII', 'OneTwo': 'FasterAttackVI', 'OneThree': 'FasterAttackV', 'OneFour': 'FasterAttackIV',
             'OneFive': 'FasterAttackIII', 'OneSix': 'FasterAttackII', 'OneSeven': 'FasterAttackI', 'Two': 'PetAttackSpeedNone',
             'TwoFive': 'SlowerAttackII', 'TwoSeven': 'SlowerAttackIII'}
    copt = {'quiverBonus': QUIVER[ho['quiverBonus']], 'ammo': ho['ammo'], 'petUptime': ho.get('petUptime', 0),
            'cobraReflexes': ho.get('cobraReflexes', False), 'petAttackSpeed': SPEED[ho['petAttackSpeed']]}
    # (Theirs dropped Pet Aggression on 2026-11-01, 8aa26ccb4d; ours keeps rank 5, a difference to report.)
    if ho.get('petType') and ho['petType'] != 'PetNone':
        copt['petType'] = {'Crocolisk': 'Crocilisk'}.get(ho['petType'], ho['petType'])
    class_opts = {'hunter': {'options': {'classOptions': copt}}}
    # Their Beast Mastery tree has Pathfinding (5th) and Improved Revive Pet (6th) the other way round.
    trees = talents.split('-'); bm = list(trees[0].ljust(6, '0')); bm[4], bm[5] = bm[5], bm[4]
    trees[0] = ''.join(bm).rstrip('0'); talents = '-'.join(trees)
    mm = (p['talentsString'].split('-') + ['', ''])[1]
    CD[3045] = '%ds' % (300 - 60 * (int(mm[6]) if len(mm) > 6 else 0))  # Rapid Fire, 1 min less a Rapid Killing rank
    # Theirs registers only each spell's top rank, the AQ book's (Aspect of the Hawk VII, Serpent Sting IX). Ours runs
    # them too, with the ours_aq binary (setup.sh), and our rotation's ids are moved up to match (OURS_AQ=0 keeps ours
    # on its own ranks, which gives the AQ ranks' worth).
    AQ = {14322: 25296, 13555: 25295}
    def remap(n):
        if isinstance(n.get('spellId'), int) and n['spellId'] in AQ and set(n) <= {'spellId', 'tag'}:
            n = {**n, 'spellId': AQ[n['spellId']]}
        if n.get('spellId') == 14266 and n.get('tag') == 3:  # ours queues Raptor Strike as its tag 3, theirs as tag 1
            n = {'spellId': 14266, 'tag': 1}
        return n
    if os.environ.get('OURS_AQ', '1') == '1':
        s_ = json.dumps(p['rotation'])
        for a_, b_ in AQ.items(): s_ = s_.replace('"spellId": %d' % a_, '"spellId": %d' % b_)
        p['rotation'] = json.loads(s_)
else:
    raise SystemExit(f'{klass} is not mapped yet (SKILL.md, "Adding a class")')

# ---- rotation ---------------------------------------------------------------------------------------------------
# APL_TRANSLATION: values theirs lacks, rewritten from ones it has.
NOSPELLS = {int(x) for x in os.environ.get('NOSPELLS', '').split(',') if x}
def add(a, b_): return {'math': {'op': 'OpAdd', 'lhs': a, 'rhs': b_}}
def sid_key(s): return ('item', s['itemId']) if 'itemId' in s else s.get('spellId')
DIST = {'unitDistance': {}}
def nospell(n):
    # A value naming a talent spell the build lacks: ours reads it as false, theirs drops it from its and/or, which
    # turns an `and` true. NOSPELLS makes it false (and never ready) for theirs too.
    if len(n) != 1: return None
    k, v = next(iter(n.items()))
    if isinstance(v, dict) and isinstance(v.get('spellId'), dict) and v['spellId'].get('spellId') in NOSPELLS:
        return k
def walk(n):
    if isinstance(n, dict):
        k = nospell(n)
        if k in ('spellIsReady', 'spellCanCast', 'auraIsActive'): return {'const': {'val': 'false'}}
        if k == 'spellTimeToReady': return {'const': {'val': '1000s'}}
        if k == 'spellUsesRemaining': return {'const': {'val': '0'}}
        if 'inMeleeRange' in n and len(n) == 1:
            return {'cmp': {'op': 'OpLe', 'lhs': DIST, 'rhs': {'const': {'val': '5'}}}}
        if 'moveTime' in n and len(n) == 1:  # the walk from where it stands to the range, at its run speed
            r = walk(n['moveTime']['rangeFromTarget'])
            return {'math': {'op': 'OpMul', 'lhs': {'const': {'val': '1s'}},
                             'rhs': {'math': {'op': 'OpDiv', 'lhs': {'math': {'op': 'OpSub', 'lhs': DIST, 'rhs': r}}, 'rhs': {'unitMovementSpeed': {}}}}}}
        if 'cmp' in n and 'spellUsesRemaining' in n['cmp']['lhs']:
            c_ = n['cmp']; u = c_['lhs']['spellUsesRemaining']; t = u['minTimeLeft']
            assert c_['rhs'] == {'const': {'val': '0'}}
            if u['spellId'].get('spellId') in NOSPELLS:
                return {'const': {'val': 'false' if c_['op'] in ('OpGt', 'OpNe') else 'true'}}
            # spellUsesRemaining(X, minTimeLeft T) > 0, used beside spellIsReady(X): one use left now
            if c_['op'] == 'OpGt':
                return {'cmp': {'op': 'OpGe', 'lhs': {'remainingTime': {}}, 'rhs': t}}
            if c_['op'] == 'OpEq':  # no use left with T to go: fewer than T left, or not ready in time
                return {'cmp': {'op': 'OpLt', 'lhs': {'remainingTime': {}}, 'rhs': add({'spellTimeToReady': {'spellId': u['spellId']}}, t)}}
        if 'cmp' in n and 'spellUsesLostByDelay' in n['cmp']['lhs']:
            # spellUsesLostByDelay(X, delay D, minTimeLeft T) > 0: delaying X by D costs a use. With R the time left
            # and fewer than two cooldowns of X in R - T: R < T + D when R < T + CD, else R < T + D + CD.
            c_ = n['cmp']; u = c_['lhs']['spellUsesLostByDelay']
            assert c_['op'] == 'OpGt' and c_['rhs'] == {'const': {'val': '0'}}
            key = sid_key(u['spellId'])
            if key not in CD: raise SystemExit(f'spellUsesLostByDelay on {key}: add its cooldown to CD')
            cd = {'const': {'val': CD[key]}}
            R = {'remainingTime': {}}; T = u['minTimeLeft']; Dl = walk(u['delay'])
            return {'or': {'vals': [
                {'and': {'vals': [{'cmp': {'op': 'OpLt', 'lhs': R, 'rhs': add(T, cd)}}, {'cmp': {'op': 'OpLt', 'lhs': R, 'rhs': add(T, Dl)}}]}},
                {'and': {'vals': [{'cmp': {'op': 'OpGe', 'lhs': R, 'rhs': add(T, cd)}}, {'cmp': {'op': 'OpLt', 'lhs': R, 'rhs': add(add(T, Dl), cd)}}]}}]}}
        for k2 in ('autoTimeToNext', 'autoTimeSinceLast'):  # theirs: MainHandAuto, OffHandAuto, RangedAuto
            if k2 in n and n[k2].get('autoType') in ('MainHand', 'Ranged', 'OffHand'):
                n = {**n, k2: {**n[k2], 'autoType': n[k2]['autoType'] + 'Auto'}}  # (autoSwingTime keeps MainHand)
        n = remap(n)
        return {k3: walk(v) for k3, v in n.items()}
    if isinstance(n, list): return [walk(x) for x in n]
    return n

if '--turret' in args:  # both stand at range: our rotation without its weave lines
    p['rotation']['priorityList'] = [a for a in p['rotation']['priorityList'] if not a.get('notes', '').startswith('Weave:')]
if APL == 'ours':
    apl = walk(copy.deepcopy(p['rotation']))
elif APL.startswith('theirs:'):
    tb = json.load(open(os.path.join(WORK, 'theirs_builds', APL[7:] + '.json')))
    apl = tb['player']['rotation'] if 'player' in tb else tb  # a build's rotation, or a bare APL
else:
    p['rotation'] = json.load(open(APL))  # a common APL: ours runs it too
    apl = walk(copy.deepcopy(p['rotation']))
if not APL.startswith('theirs:'):
    apl['priorityList'] = extra_lines + apl['priorityList']
    for k in ('spellUsesRemaining', 'spellUsesLostByDelay', 'moveTime', 'inMeleeRange'):
        assert k not in json.dumps(apl), f'{k} left in theirs\' rotation'

ALIGN = {int(k): float(v) for k, v in (x.split('=') for x in os.environ.get('ALIGN', '').split(',') if x)}
bonus = dict(SET_BONUS)
for k, v in ALIGN.items(): bonus[k] = bonus.get(k, 0) + v
e = ours['encounter']; t = e['targets'][0]
player = {
    'name': 'ours', 'race': p['race'], 'class': klass, 'equipment': {'items': equip},
    'consumables': {k: v for k, v in consumables.items() if v}, 'buffs': ind_buffs, **class_opts,
    'talentsString': talents, 'profession1': p.get('profession1'), 'profession2': p.get('profession2'),
    'prof1Rank': 300, 'prof2Rank': 300, 'cooldowns': p.get('cooldowns', {}), 'rotation': apl,
    'reactionTimeMs': p['reactionTimeMs'], 'channelClipDelayMs': p['channelClipDelayMs'],
    'distanceFromTarget': p['distanceFromTarget'], 'bonusStats': {'stats': stats_arr(bonus)},
}
theirs = {
    'raid': {'parties': [{'players': [player], 'buffs': party_buffs}], 'buffs': raid_buffs, 'debuffs': debuffs},
    'encounter': {'duration': e['duration'], 'durationVariation': e['durationVariation'],
                  'executeProportion20': e['executeProportion20'], 'executeProportion35': e['executeProportion35'],
                  'targets': [{'id': t['id'], 'name': t['name'], 'level': t['level'], 'stats': stats_arr(stats_map(t['stats'])),
                               'minBaseDamage': t['minBaseDamage'], 'damageSpread': t['damageSpread'],
                               'swingSpeed': t['swingSpeed'], 'parryHaste': t['parryHaste'],
                               **({'mobType': t['mobType']} if t.get('mobType') else {}),
                               **({'mobType': opt('--their-mob-type')} if opt('--their-mob-type') else {})}]},
    'simOptions': ours['simOptions'],
}
json.dump(theirs, open(os.path.join(out, 'req_theirs.json'), 'w'), indent=1)
json.dump(ours, open(os.path.join(out, 'req_ours.json'), 'w'), indent=1)
print(f'{klass}: {len(sim_items)} synthetic items, talents {p["talentsString"]} -> {talents}, apl {APL}')
