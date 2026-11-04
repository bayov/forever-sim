# gen.py <request-prefix> <preset> <outdir>: writes <outdir>/<variant>_<preset>.json from
# tools/tankbench/req/<request-prefix>_<preset>.json with one consumable changed.
# Run from the repo root, e.g. python3 orchestration/tank_handoff/shaman/round5/gen.py shaman_warden balanced /tmp/r5/req
import json, os, sys

V = {}
for n, f in [('titans', 'FlaskOfTheTitans'), ('wisdom', 'FlaskOfDistilledWisdom'), ('supreme', 'FlaskOfSupremePower'),
             ('chromatic', 'FlaskOfChromaticResistance'), ('prec', 'FlaskOfNaturalPrecision'), ('acc', 'FlaskOfNaturalAccuracy'),
             ('aggr', 'FlaskOfNaturalAggression'), ('swift', 'FlaskOfNaturalSwiftness')]:
    V['flask_' + n] = [('flask', f)]
V['agi_greater'] = [('agilityElixir', 'ElixirOfGreaterAgility')]; V['agi_none'] = [('agilityElixir', None)]
V['armor_superior'] = [('armorElixir', 'ElixirOfSuperiorDefense')]; V['armor_none'] = [('armorElixir', None)]
V['hp_forever'] = [('healthElixir', 'ElixirOfFortitudeForever')]; V['hp_none'] = [('healthElixir', None)]
V['str_giants'] = [('strengthBuff', 'ElixirOfGiants')]; V['str_none'] = [('strengthBuff', None)]
V['ap_firewater'] = [('attackPowerBuff', 'WinterfallFirewater')]; V['ap_none'] = [('attackPowerBuff', None)]
V['sp_arcane'] = [('spellPowerBuff', 'GreaterArcaneElixir')]
V['fire_greater'] = [('firePowerBuff', 'ElixirOfGreaterFirepower')]
V['nature_power'] = [('naturePowerBuff', 'ElixirOfNaturePower')]
V['mageblood'] = [('manaRegenElixir', 'MagebloodPotion')]
V['int_greater'] = [('intellectElixir', 'ElixirOfGreaterIntellect')]
V['trolls_major'] = [('trollsBloodElixir', 'MajorTrollsBloodElixir')]
V['pot_stoneshield'] = [('defaultPotion', 'GreaterStoneshieldPotion')]; V['pot_healing'] = [('defaultPotion', 'MajorHealingPotion')]
V['pot_none'] = [('defaultPotion', None)]
V['conj_rune'] = [('defaultConjured', 'ConjuredDemonicRune')]
V['chili_off'] = [('dragonBreathChili', False)]
# The Zanza slot holds the Zanza and Blasted Lands buffs (one at a time in the sim); rum is the alcohol slot.
for n, f in [('spirit', 'SpiritOfZanza'), ('sheen', 'SheenOfZanza'), ('swift', 'SwiftnessOfZanza'), ('roids', 'ROIDS'),
             ('scorpok', 'GroundScorpokAssay'), ('lung', 'LungJuiceCocktail'), ('gizzard', 'GizzardGum'),
             ('cortex', 'CerebralCortexCompound')]:
    V['zanza_' + n] = [('zanzaBuff', f)]
V['rum'] = [('alcohol', 'AlcoholRumseyRumBlackLabel')]
V['stout'] = [('alcohol', 'AlcoholKreegsStoutBeatdown')]
# A stone or oil and a shaman imbue are both the main hand's temporary enchant, so these
# variants put the stone or oil on in place of Rockbiter Weapon (shamanImbue removed).
for n, f in [('elemental', 'ElementalSharpeningStone'), ('dense', 'DenseSharpeningStone'),
             ('manaoil', 'BrilliantManaOil'), ('wizardoil', 'BrilliantWizardOil')]:
    V['mh_' + n] = [('mainHandImbue', f), ('-imbue', None)]
# The bear's mainHandImbue is a consumable already (no shaman imbue): the same stones as plain swaps.
V['stone_elemental'] = [('mainHandImbue', 'ElementalSharpeningStone')]

pre, p, out = sys.argv[1:4]
names = sys.argv[4:] or list(V)
os.makedirs(out, exist_ok=True)
for n in names:
    d = json.load(open(f'tools/tankbench/req/{pre}_{p}.json'))
    pl = d['raid']['parties'][0]['players'][0]
    for k, v in V[n]:
        if k == '-imbue':
            pl['wardenShaman']['options'].pop('shamanImbue', None)
        elif v is None:
            pl['consumes'].pop(k, None)
        else:
            pl['consumes'][k] = v
    json.dump(d, open(f'{out}/{n}_{p}.json', 'w'), indent=1)
