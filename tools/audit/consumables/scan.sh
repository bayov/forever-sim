#!/bin/bash
# usage: AUDIT_TREE=<checkout> scan.sh [ref, default origin/master]
# The standing checks on the eleven Level 60 DPS specs. Prints only violations, then a few counts to eyeball. Silence above "== counts" = clean.
set -e
cd "${AUDIT_TREE:?}"; b=${1:-origin/master}
SPECS="warrior retribution_paladin hunter rogue shadow_priest elemental_shaman enhancement_shaman mage warlock balance_druid feral_druid"
paths=""; for d in $SPECS; do paths="$paths ui/$d/presets.ts $(git ls-tree -r --name-only $b ui/$d/gear_sets | grep level60 | tr '\n' ' ')"; done
paths="$paths $(git ls-tree --name-only $b tools/rotopt/mksettings/ | grep -E 'spec_|main.go' | tr '\n' ' ')"
echo "== banned (raid-only sources, Revelation item, ruled-out buffs, Expose Armor on)"
git grep -n -i -E "BrilliantWizardOil|ElementalSharpening|MagebloodPotion|\"enchant\": *(2504|2544)\b|1248805|EphemeralPower|FireToastedBun|ExposeArmor: *proto\.TristateEffect_TristateEffect(Regular|Improved)|exposeArmor: *TristateEffect\.TristateEffect(Regular|Improved)|sanctityAura: *true|curseOfShadow: *true" $b -- $paths | grep -v -E ':\s*//' || true
echo "== Revelation outside the shamans"; git grep -n Revelation $b -- $paths | grep -v -E ':\s*//|_shaman' || true
echo "== non-synthetic trinkets (slots 12/13 must be 99xxxx or empty)"
for g in $(echo $paths | tr ' ' '\n' | grep gear.json); do git show $b:$g | python3 -c "
import json,sys
it=json.load(sys.stdin).get('items',[])
for i in (12,13):
  if i<len(it) and it[i] and it[i].get('id') and not str(it[i]['id']).startswith('99'): print('NONSYN',sys.argv[1],it[i]['id'])
" "$g"; done
echo "== more than one +2% crit elixir in a Consumes preset"
for d in $SPECS; do git show $b:ui/$d/presets.ts | python3 -c "
import re,sys
t=sys.stdin.read()
for m in re.finditer(r'export const (\w+) = Consumes\.create\(\{(.*?)\}\);',t,re.S):
  if len(re.findall(r'Mongoose|Grizzly|OfTheOwl|Sages|Whale',m.group(2)))>1: print('2crit',sys.argv[1],m.group(1))
" "$d"; done
echo "== counts"
echo "minor-armor exclusive effects in debuffs.go (expect 2): $(git show $b:sim/core/debuffs.go | grep -c 'NewExclusiveEffect(minorArmorReductionEffectCategory')"
echo "mksettings main.go: $(git show $b:tools/rotopt/mksettings/main.go | grep -o -E 'MobType[A-Za-z]+|SunderArmor: *true|Salvation: *true' | sort | uniq -c | tr '\n' ' ')  (expect MobTypeUnknown default; Humanoid only as the alternate)"
for f in warrior retribution_paladin rogue feral_druid enhancement_shaman hunter; do echo "$f raid air/weapon totems: $(git show $b:ui/$f/presets.ts | awk '/RaidBuffs.create\(\{/,/\}\);/' | grep -o -E 'graceOfAirTotem: *[A-Za-z.]+|totemWeaponBuff: *[A-Za-z.]+' | sort -u | tr '\n' ' ')"; done
