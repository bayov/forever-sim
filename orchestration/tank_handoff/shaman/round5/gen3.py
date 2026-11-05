# gen3.py <reqdir>: round 2 of the round 5 candidates (every single-slot winner of round 5's first pass, out/) put together,
# <reqdir>/final/<request>.json, and the checks on them, <reqdir>/loo/<request-prefix>/<check>_<preset>.json: each
# winner taken back out (or swapped for the runner-up in its slot), and the Greater Stoneshield Potion with an
# explicit potion action (autocast never uses a defensive potion: HpPercentForDefensives is 0 in a tank score).
# Run from the repo root.
import copy, json, os, sys

out = sys.argv[1]
POT = {'action': {'castSpell': {'spellId': {'otherId': 'OtherActionPotion'}}}, 'notes': 'Greater Stoneshield Potion on cooldown.'}
SPELL_ELIXIRS = {'spellPowerBuff': 'GreaterArcaneElixir', 'naturePowerBuff': 'ElixirOfNaturePower', 'intellectElixir': 'ElixirOfGreaterIntellect'}
STONE = {'defaultPotion': 'GreaterStoneshieldPotion', '+pot': True}
# Round 2 of the combination (gen3.py, run4.sh, out4/): gen2's checks (out3/) moved the Warden to the Greater
# Stoneshield Potion with the potion action, Max threat to the Ground Scorpok Assay and Max survival to Natural Precision.
FINAL = {
    'shaman_warden_max_threat': dict(flask='FlaskOfNaturalPrecision', zanzaBuff='GroundScorpokAssay', alcohol='AlcoholRumseyRumBlackLabel',
                                     defaultConjured='ConjuredDemonicRune', **SPELL_ELIXIRS, **STONE),
    'shaman_warden_balanced': dict(flask='FlaskOfNaturalPrecision', zanzaBuff='SpiritOfZanza', alcohol='AlcoholRumseyRumBlackLabel',
                                   defaultConjured='ConjuredDemonicRune', **SPELL_ELIXIRS, **STONE),
    'shaman_warden_max_survival': dict(flask='FlaskOfNaturalPrecision', zanzaBuff='SpiritOfZanza', alcohol='AlcoholRumseyRumBlackLabel',
                                       **SPELL_ELIXIRS, **STONE),
    'druid_bear_max_threat': dict(flask='FlaskOfNaturalPrecision', zanzaBuff='ROIDS', alcohol='AlcoholRumseyRumBlackLabel'),
    'druid_bear_balanced': dict(flask='FlaskOfNaturalPrecision', zanzaBuff='SpiritOfZanza', alcohol='AlcoholRumseyRumBlackLabel'),
    'druid_bear_max_survival': dict(flask='FlaskOfNaturalPrecision', zanzaBuff='SpiritOfZanza', alcohol='AlcoholRumseyRumBlackLabel'),
}
# Checks on the final: name -> {consumes key: value or None}; '+pot' adds the potion action.
CHECKS = {
    'shaman_warden': {
        'manapot': {'defaultPotion': 'MajorManaPotion', '-pot': True}, 'no_rune': {'defaultConjured': None},
        'rune': {'defaultConjured': 'ConjuredDemonicRune'}, 'no_int': {'intellectElixir': None},
        'mageblood': {'manaRegenElixir': 'MagebloodPotion'}, 'titans': {'flask': 'FlaskOfTheTitans'},
        'spirit': {'zanzaBuff': 'SpiritOfZanza'}, 'scorpok': {'zanzaBuff': 'GroundScorpokAssay'},
    },
    # The bear's round 3 Stoneshield A/B left it to autocast, which never drinks a defensive potion in a tank score.
    'druid_bear': {'stoneshield': STONE},
}


def load(req):
    return json.load(open(f'tools/tankbench/req/{req}.json'))


def apply(d, changes):
    pl = d['raid']['parties'][0]['players'][0]
    for k, v in changes.items():
        if k == '+pot':
            rot = pl['rotation']
            isstr = isinstance(rot, str)
            r = json.loads(rot) if isstr else rot
            i = next(i for i, a in enumerate(r['priorityList']) if 'autocastOtherCooldowns' in a.get('action', {}))
            r['priorityList'].insert(i, copy.deepcopy(POT))
            pl['rotation'] = json.dumps(r) if isstr else r
        elif k == '-pot':
            rot = pl['rotation']
            isstr = isinstance(rot, str)
            r = json.loads(rot) if isstr else rot
            r['priorityList'] = [a for a in r['priorityList'] if a != POT]
            pl['rotation'] = json.dumps(r) if isstr else r
        elif v is None:
            pl['consumes'].pop(k, None)
        else:
            pl['consumes'][k] = v
    return d


os.makedirs(f'{out}/final', exist_ok=True)
for req, ch in FINAL.items():
    json.dump(apply(load(req), ch), open(f'{out}/final/{req}.json', 'w'), indent=1)
    pre, p = ('shaman_warden', req[len('shaman_warden_'):]) if req.startswith('shaman') else ('druid_bear', req[len('druid_bear_'):])
    os.makedirs(f'{out}/loo/{pre}', exist_ok=True)
    for n, c in CHECKS[pre].items():
        json.dump(apply(json.load(open(f'{out}/final/{req}.json')), c), open(f'{out}/loo/{pre}/{n}_{p}.json', 'w'), indent=1)
