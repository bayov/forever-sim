"""mk.py OUT ROT TALENTS [race=RaceHuman] [knob=value ...]: a healbench priest request with that rotation."""
import json, sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from rots import ROTS
TALENTS = {
    'holy': '00500503113-03505103030121531',
    'disc': '005005031330011531-0350510303',
}
out, rot, talents = sys.argv[1], sys.argv[2], sys.argv[3]
talents = TALENTS.get(talents, talents)
knobs = dict(a.split('=', 1) for a in sys.argv[4:])
race = knobs.pop('race', 'RaceHuman')
base = json.load(open('tools/healbench/req/shaman_restoration.json'))
p = base['raid']['parties'][0]['players'][0]
p['class'] = 'ClassPriest'
p['race'] = race
p['talentsString'] = talents
p.pop('restorationShaman', None)
p['healingPriest'] = {'options': {}}
p['rotation'] = ROTS[rot](knobs)
p['equipment'] = json.load(open('ui/healing_priest/gear_sets/level60_phase1_synthetic.gear.json'))
for key in list(knobs):
    if key.startswith('cons_'):
        p['consumes'][key[5:]] = knobs.pop(key)
json.dump(base, open(out, 'w'), indent=2)
