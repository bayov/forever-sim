#!/bin/bash
# bench.sh ROLE ITERS "boss1 boss2" req1.json req2.json ...  (extra healbench flags in $HB)
role=$1; iters=$2; bosses=$3; shift 3
go build -tags=with_db -o .ps/healbench ./tools/healbench || exit 1
for b in $bosses; do
  ./.ps/healbench -role $role -encounter tools/encounters/$b.json -iters $iters -carry $HB "$@" 2>/dev/null | grep '^| [a-z]' | grep -v '^| Healer' | sed "s/^/$b /"
done
