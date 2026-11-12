#!/usr/bin/env python3
"""tank_fairness: write one-field variants of each tank preset request (consumables, flask, enchants, race).
usage: variants.py <outdir> <request.json>...   -> <outdir>/<preset>@<variant>.json
Each variant changes one thing; bench it with bench.sh at the base's seed and compare with paired.py."""
import copy,json,os,sys
FLASKS=['FlaskOfTheTitans','FlaskOfSupremePower','FlaskOfNaturalPrecision','FlaskOfNaturalAccuracy','FlaskOfNaturalAggression','FlaskOfNaturalSwiftness']
FOODS=['FoodSmokedDesertDumpling','FoodDirgesKickChimaerokChops','FoodNightfinSoup','FoodTenderWolfSteak','FoodGrilledSquid']
ZANZA=['SpiritOfZanza','ROIDS','GroundScorpokAssay']
# enchant effect ids by slot index (items list): head 0, neck 1, back 3, chest 4, wrist 5, hands 6, legs 8, feet 9, MH 14, OH 15
ENCH={0:[1503,1506,1507,1508],8:[1503,1506,1507,1508],1:[1249019,1249059],3:[1219587,1889],4:[1891,1892],
      5:[1885,1248665,1886],6:[1248640,1887],9:[1887,929]}
POTION_ACTION={"action":{"castSpell":{"spellId":{"otherId":"OtherActionPotion"}}},"notes":"audit: potion"}
def write(out,name,d):
    json.dump(d,open(os.path.join(out,name+'.json'),'w'),indent=1)
def main(out,reqs):
    os.makedirs(out,exist_ok=True)
    for r in reqs:
        base=json.load(open(r)); n=os.path.basename(r)[:-5]; p0=base['raid']['parties'][0]['players'][0]
        c0=p0['consumes']; items=p0['equipment']['items']; cls=n.split('_')[0]
        def var(tag,fn):
            d=copy.deepcopy(base); p=d['raid']['parties'][0]['players'][0]
            if fn(p) is False: return
            write(out,n+'@'+tag,d)
        for f in FLASKS:
            if f!=c0.get('flask'): var('flask='+f,lambda p,f=f:p['consumes'].__setitem__('flask',f))
        for f in FOODS:
            if f!=c0.get('food'): var('food='+f,lambda p,f=f:p['consumes'].__setitem__('food',f))
        for z in ZANZA:
            if z!=c0.get('zanzaBuff'): var('zanza='+z,lambda p,z=z:p['consumes'].__setitem__('zanzaBuff',z))
        if cls=='shaman':
            if c0.get('manaRegenElixir'): var('mageblood=none',lambda p:p['consumes'].pop('manaRegenElixir'))
            else: var('mageblood=Greater',lambda p:p['consumes'].__setitem__('manaRegenElixir','GreaterMagebloodElixir'))
        for slot,opts in ENCH.items():
            if slot>=len(items) or not items[slot]: continue
            cur=items[slot].get('enchant')
            for e in opts:
                if e!=cur: var('ench%d=%d'%(slot,e),lambda p,s=slot,e=e:p['equipment']['items'][s].__setitem__('enchant',e))
        mh=items[14] if len(items)>14 else {}
        if mh.get('enchant') in (2646,1248668):
            e=1248668 if mh['enchant']==2646 else 2646
            var('ench14=%d'%e,lambda p,e=e:p['equipment']['items'][14].__setitem__('enchant',e))
        oh=items[15] if len(items)>15 else {}
        if oh and oh.get('enchant') in (863,1886):
            e=1886 if oh['enchant']==863 else 863
            var('ench15=%d'%e,lambda p,e=e:p['equipment']['items'][15].__setitem__('enchant',e))
        if cls=='paladin':
            var('imbue=DenseSharpeningStone',lambda p:p['consumes'].__setitem__('mainHandImbue','DenseSharpeningStone'))
            def stone(p):
                p['consumes']['defaultPotion']='GreaterStoneshieldPotion'
                pl=p['rotation']['priorityList']
                i=next(k for k,a in enumerate(pl) if 'autocastOtherCooldowns' in json.dumps(a))
                pl.insert(i,copy.deepcopy(POTION_ACTION))
            var('potion=StoneshieldAction',stone)
        if cls in('warrior','druid'):
            alt='GreaterStoneshieldPotion' if c0.get('defaultPotion')=='MightyRagePotion' else 'MightyRagePotion'
            var('potion='+alt,lambda p:p['consumes'].__setitem__('defaultPotion',alt))
        if cls=='shaman':
            for race in ('RaceOrc','RaceTroll','RaceTauren','RaceDwarf'):
                var('race='+race,lambda p,race=race:p.__setitem__('race',race))
main(sys.argv[1],sys.argv[2:])
