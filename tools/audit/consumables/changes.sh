#!/bin/bash
# usage: AUDIT_TREE=<checkout> changes.sh <last audited master sha>
# Prints what moved in the files this audit owns: on origin/master since the sha, and on any class/consumes branch ahead of master.
# A pass is quiet when both lists are empty and wago has no newer 1.60 build.
set -e
cd "${AUDIT_TREE:?}"; git fetch -q origin
P="ui/*/presets.ts tools/rotopt/mksettings/*.go ui/*/gear_sets/level60* sim/core/consumes.go sim/core/buffs.go sim/core/debuffs.go sim/core/totem_weapon_buffs.go sim/common tools/database/forever_enchants.go tools/dps_report/assumptions_def.py sim/mage/forever_staff_imbue.go"
echo "master ${1:?}..$(git rev-parse --short origin/master)"; git diff --stat "$1" origin/master -- $P | tail -40
for c in warrior paladin hunter rogue priest shaman mage warlock druid consumes; do
  b=origin/claude/wow-forever-dps-presets-m9u15h-$c; git rev-parse -q --verify $b >/dev/null || continue
  n=$(git rev-list --count origin/master..$b)
  [ "$n" -gt 0 ] && echo "$c $(git rev-parse --short $b) ahead $n: $(git diff --name-only $(git merge-base origin/master $b) $b -- $P | tr '\n' ' ')"
done; true
