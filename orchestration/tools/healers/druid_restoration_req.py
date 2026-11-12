"""Write tools/healbench/req/druid_restoration*.json from the shaman's request: the same raid, buffs and
consumes, a Restoration druid player with the gear, talents and rotation given.

    python3 mkreq.py OUT.json APL.json TALENTS GEAR.json|balance [RACE]
"""
import json, sys

out, apl_path, talents, gear = sys.argv[1:5]
race = sys.argv[5] if len(sys.argv) > 5 else 'RaceNightElf'
req = json.load(open('tools/healbench/req/shaman_restoration.json'))
p = req['raid']['parties'][0]['players'][0]
if gear == 'balance':
    bal = json.load(open('tools/dps_report/req/druid__balance.json'))
    equipment = bal['raid']['parties'][0]['players'][0]['equipment']
else:
    equipment = json.load(open(gear))
p.pop('restorationShaman', None)
p['class'] = 'ClassDruid'
p['race'] = race
p['talentsString'] = talents
p['equipment'] = equipment
p['restorationDruid'] = {'options': {'innervateTarget': {'type': 'Self'}}}
p['rotation'] = json.load(open(apl_path))
c = p['consumes']
c.pop('mainHandImbue', None)
c['mainHandImbue'] = 'WizardOil'
c['food'] = 'FoodWickedSmoothie'  # the druid's best phase 1 food (docs/healing_sim.md, Consumables)
json.dump(req, open(out, 'w'), indent=2)
open(out, 'a').write('\n')
