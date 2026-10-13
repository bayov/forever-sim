# usage: spec_shots.sh <harness prefix> <shot prefix> <scheme> <mode>
cd "$(dirname "$0")"; mkdir -p shots
export NODE_PATH=${NODE_PATH:-/opt/node22/lib/node_modules} HP=$1
P=$2; SC=${3:-light}; M=${4:-plain}
node shoot.js ${P}_tl_lanes $M phone $SC '[["tab","spec"],["click","#tl-zoom [data-z=\"30\"]"],["scroll","#tl"]]' view &
node shoot.js ${P}_tl_single $M phone $SC '[["tab","spec"],["click","#tl-mode [data-v=\"single\"]"],["scroll","#tl"]]' view &
node shoot.js ${P}_tl_single_tip $M phone $SC '[["tab","spec"],["click","#tl-mode [data-v=\"single\"]"],["eval","document.querySelector(\"#tl-scroll\").scrollLeft=0"],["scroll","#tl"],["eval","const g=document.querySelectorAll(\"#tl-scroll svg g[data-tip]\")[6]; g.dispatchEvent(new MouseEvent(\"click\",{bubbles:true}))"]]' view &
node shoot.js ${P}_auras $M phone $SC '[["tab","spec"],["scroll","#sp-auras"]]' view &
node shoot.js ${P}_rot $M phone $SC '[["tab","spec"],["eval","document.querySelector(\"#sp-rot-card\").scrollIntoView()"]]' view &
node shoot.js ${P}_gear $M phone $SC '[["tab","spec"],["eval","document.querySelector(\"#sp-gear-card\").scrollIntoView()"]]' view &
node shoot.js ${P}_setup $M phone $SC '[["tab","spec"],["eval","document.querySelector(\"#sp-setup-card\").scrollIntoView()"]]' view &
wait
