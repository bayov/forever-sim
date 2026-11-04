# gen.py <preset> : writes req/<variant>_<preset>.json from tools/tankbench/req/warrior_protection_<preset>.json (one consumable changed)
import json,sys,copy
V={}
def c(k,v): return (k,v)
V['flask_aggr']=('flask','FlaskOfNaturalAggression');V['flask_acc']=('flask','FlaskOfNaturalAccuracy');V['flask_none']=('flask',None)
V['flask_prec']=('flask','FlaskOfNaturalPrecision');V['flask_swift']=('flask','FlaskOfNaturalSwiftness')
for n in ['Dirges:FoodDirgesKickChimaerokChops','Squid:FoodGrilledSquid','Nightfin:FoodNightfinSoup','Runn:FoodRunnTumTuberSurprise','Sunfruit:FoodBlessSunfruit','SunfruitJuice:FoodBlessedSunfruitJuice','WolfRibs:FoodHotWolfRibs','WolfSteak:FoodTenderWolfSteak','Sagefish:FoodSmokedSagefish','SagefishD:FoodSagefishDelight','None:None']:
    a,b=n.split(':');V['food_'+a]=('food',None if b=='None' else b)
V['agi_greater']=('agilityElixir','ElixirOfGreaterAgility');V['agi_none']=('agilityElixir',None)
V['armor_superior']=('armorElixir','ElixirOfSuperiorDefense');V['armor_greater']=('armorElixir','ElixirOfGreaterDefense');V['armor_none']=('armorElixir',None)
V['hp_forever']=('healthElixir','ElixirOfFortitudeForever');V['hp_none']=('healthElixir',None)
V['str_giants']=('strengthBuff','ElixirOfGiants');V['str_none']=('strengthBuff',None)
V['str_growth']=('strengthBuff','ElixirOfGiantGrowth');V['str_grizzly']=('strengthBuff','ElixirOfTheGrizzly')
V['ap_firewater']=('attackPowerBuff','WinterfallFirewater');V['ap_none']=('attackPowerBuff',None)
V['pot_stoneshield']=('defaultPotion','GreaterStoneshieldPotion');V['pot_great_rage']=('defaultPotion','GreatRagePotion');V['pot_none']=('defaultPotion',None);V['pot_healing']=('defaultPotion','MajorHealingPotion')
V['stone_none']=('mainHandImbue',None);V['stone_elemental']=('mainHandImbue','ElementalSharpeningStone');V['stone_dense_weight']=('mainHandImbue','DenseWeightstone')
V['chili_off']=('dragonBreathChili',False)
V['zanza_spirit']=('zanzaBuff','SpiritOfZanza');V['zanza_sheen']=('zanzaBuff','SheenOfZanza');V['zanza_swift']=('zanzaBuff','SwiftnessOfZanza');V['zanza_roids']=('zanzaBuff','ROIDS');V['zanza_scorpok']=('zanzaBuff','GroundScorpokAssay');V['zanza_lung']=('zanzaBuff','LungJuiceCocktail');V['zanza_gizzard']=('zanzaBuff','GizzardGum')
V['rum']=('alcohol','AlcoholRumseyRumBlackLabel');V['grog']=('alcohol','AlcoholGordokGreenGrog');V['stout']=('alcohol','AlcoholKreegsStoutBeatdown')
V['grenade']=('fillerExplosive','ExplosiveThoriumGrenade');V['dynamite']=('fillerExplosive','ExplosiveDenseDynamite');V['sapper']=('sapperExplosive','SapperGoblinSapper')
V['conj_rune']=('defaultConjured','ConjuredDemonicRune')
p=sys.argv[1]
for n,(k,v) in V.items():
    d=json.load(open(f'/home/user/forever-sim/tools/tankbench/req/warrior_protection_{p}.json'));pl=d['raid']['parties'][0]['players'][0]
    if v is None: pl['consumes'].pop(k,None)
    else: pl['consumes'][k]=v
    json.dump(d,open(f'req/{n}_{p}.json','w'),indent=1)
