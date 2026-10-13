# usage: suite.sh <harness prefix> <shot prefix>
cd "$(dirname "$0")"; mkdir -p shots
export NODE_PATH=${NODE_PATH:-/opt/node22/lib/node_modules} HP=$1
P=$2
OPEN='document.querySelectorAll(".fv details").forEach(d=>d.open=true)'
for sch in light dark; do
node shoot.js ${P}_rank_$sch plain phone $sch '[]' slice &
node shoot.js ${P}_races_$sch plain phone $sch '[["tab","races"]]' slice &
node shoot.js ${P}_spec_$sch plain phone $sch "[[\"tab\",\"spec\"],[\"eval\",\"$(echo $OPEN | sed 's/"/\\"/g')\"]]" slice &
node shoot.js ${P}_cmp_$sch plain phone $sch '[["tab","compare"]]' slice &
node shoot.js ${P}_method_$sch plain phone $sch '[["tab","method"]]' slice &
wait
done
node shoot.js ${P}_x_tt str phone dark '[["tab","spec"],["scroll","#sp-abil"],["tap","#sp-abil .nm"]]' view &
node shoot.js ${P}_x_tl str phone light '[["tab","spec"],["click","#tl-zoom [data-z=\"15\"]"],["scroll","#tl"],["tap","#tl-scroll"],["scroll","#tl-readout"]]' view &
node shoot.js ${P}_x_tal str phone dark '[["tab","spec"],["eval","[...document.querySelectorAll(\"#sp-tal-seg button\")].pop().click()"],["scroll","#sp-tal"]]' view &
node shoot.js ${P}_x_pet str phone light '[["tab","spec"],["eval","const s=document.querySelector(\"#sp-pick\");s.value=\"warlock__demonology\";s.dispatchEvent(new Event(\"change\"))"],["scroll","#tl"]]' view &
node shoot.js ${P}_x_gear str phone light '[["tab","spec"],["eval","document.querySelector(\"#sp-gear-card details\").open=true"],["scroll","#sp-gear"],["tap","#sp-gear tbody tr"]]' view &
node shoot.js ${P}_d_rank plain desk light '[]' &
node shoot.js ${P}_d_spec plain desk dark '[["tab","spec"]]' &
wait
