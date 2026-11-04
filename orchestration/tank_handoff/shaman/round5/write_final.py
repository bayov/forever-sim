# write_final.py: writes the round 5 consumables (and the potion action) into tools/tankbench/req/{shaman_warden,
# druid_bear}_*.json; the UI APLs got the same potion line by hand. gen3.py's FINAL plus out4/'s winners: the Mageblood Potion at
# Warden Max threat and Balanced, the Demonic Rune at Warden Max survival, the Greater Stoneshield Potion (with the
# potion action) at bear Max survival. Run once from the repo root, on the round 4 requests.
import copy, json

POT = {'action': {'castSpell': {'spellId': {'otherId': 'OtherActionPotion'}}}, 'notes': 'Greater Stoneshield Potion on cooldown.'}
SPELL = {'spellPowerBuff': 'GreaterArcaneElixir', 'naturePowerBuff': 'ElixirOfNaturePower', 'intellectElixir': 'ElixirOfGreaterIntellect'}
COMMON = {'flask': 'FlaskOfNaturalPrecision', 'alcohol': 'AlcoholRumseyRumBlackLabel'}
STONE = {'defaultPotion': 'GreaterStoneshieldPotion'}
FINAL = {
    'shaman_warden_max_threat': ({**COMMON, **SPELL, **STONE, 'zanzaBuff': 'GroundScorpokAssay', 'defaultConjured': 'ConjuredDemonicRune',
                                  'manaRegenElixir': 'MagebloodPotion'}, True),
    'shaman_warden_balanced': ({**COMMON, **SPELL, **STONE, 'zanzaBuff': 'SpiritOfZanza', 'defaultConjured': 'ConjuredDemonicRune',
                                'manaRegenElixir': 'MagebloodPotion'}, True),
    'shaman_warden_max_survival': ({**COMMON, **SPELL, **STONE, 'zanzaBuff': 'SpiritOfZanza', 'defaultConjured': 'ConjuredDemonicRune'}, True),
    'druid_bear_max_threat': ({**COMMON, 'zanzaBuff': 'ROIDS'}, False),
    'druid_bear_balanced': ({**COMMON, 'zanzaBuff': 'SpiritOfZanza'}, False),
    'druid_bear_max_survival': ({**COMMON, **STONE, 'zanzaBuff': 'SpiritOfZanza'}, True),
}


def add_pot(rot):
    i = next(i for i, a in enumerate(rot['priorityList']) if 'autocastOtherCooldowns' in a.get('action', {}))
    rot['priorityList'].insert(i, copy.deepcopy(POT))


for req, (cons, pot) in FINAL.items():
    f = f'tools/tankbench/req/{req}.json'
    d = json.load(open(f))
    pl = d['raid']['parties'][0]['players'][0]
    pl['consumes'].update(cons)
    if pot:
        add_pot(pl['rotation'])
    ind = 2 if open(f).read().startswith('{\n  "') else 1
    json.dump(d, open(f, 'w'), indent=ind)
    open(f, 'a').write('\n') if open(f).read()[-1] != '\n' else None
