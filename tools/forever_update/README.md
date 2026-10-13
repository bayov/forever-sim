# Forever Update: the daily report page

The daily "WoW Forever Daily Update" routine writes the day's findings as `report.json`. `build.py` turns that file into
one self-contained HTML page, which the routine publishes as an artifact. The page uses the same style as the Forever DPS
report: a dark stone header, a sticky tab bar, cards, class colours, and game-style tooltips (tap on a phone, hover with
a mouse).

```bash
python3 tools/forever_update/build.py report.json -o forever-update.html          # add --strict to fail on any problem
python3 tools/forever_update/build.py tools/forever_update/example.json -o /tmp/example.html   # the Oct 13 report
```

`build.py` needs only Python 3 and curl, plus network access to `nether.wowhead.com` (Wowhead Forever tooltips) and
`wow.zamimg.com` (icons). For every item in `items.list` and every `"ref"`, it fetches the Forever tooltip (name, quality,
icon, every stat line) and inlines the icons, so the report only needs the IDs. At the end it lists problems (unknown
labels, items Wowhead doesn't have yet). Fix them before publishing. Tooltips and icons are cached in
`~/.cache/forever_update`.

## Pages

| Tab | Shows |
|---|---|
| Today | Big changes, then one card each for Shaman, Paladin and Rogue: **class changes only**. Each card links to its items. |
| Items | Revealed/hidden counters, then every newly revealed or changed item as a game tooltip. You can filter by class (S/P/R by default). |
| Sim | PR status and link, the newest build vs the sim's build, PR changes, a per-spec DPS impact chart, "needs your call" items and TASKS.md items |
| Others | Class changes for the other six classes |
| News | Everything else, then the sources |

Items never appear in the class cards. They go in `items.list`, with `who` naming the classes that want them.

## report.json

All text is plain, and `**bold**` is the only markup. Leave out any key you don't need.

```jsonc
{
  "date": "2026-10-13",                       // required
  "pageTitle": "Forever Update Oct 13",       // <title>; publish the artifact with the same title
  "title": "Forever Update · Oct 13",         // big header line (default: built from date)
  "headline": "**Quiet day:** ...",           // one or two sentences under the title
  "phase": "Beta", "phaseNote": "ends Oct 21–22 · live Nov 4",
  "window": "Since Mon 12 Oct 20:00 IDT",
  "build": { "newest": "1.60.1.70338", "newestDate": "Oct 10", "sim": "70338", "isNew": false },
  "alerts": ["**No push access** ..."],       // red lines in the header (no push access, network refused, ...)
  "kpis": [],                                 // optional override of the four tiles: {k, v, s, tab}

  "big": [ { "title": "...", "text": "...", "ref": "i:18671", "old": "35", "new": "90", "kind": "buff", "status": "client" } ],

  "classes": {                                // Shaman, Paladin, Rogue: always all three, even if empty
    "Shaman": { "summary": "optional one line", "changes": [ CHANGE, ... ] },
    "Paladin": { "changes": [] },
    "Rogue": { "changes": [] }
  },
  "other": { "Warrior": { "changes": [ CHANGE ] }, ... },    // only classes with changes
  "otherNote": "optional line above the other classes",

  "items": {
    "counters": [ { "label": "Overall", "revealed": 4731, "total": 12599, "prev": 4720 } ],  // prev = yesterday's count, when known
    "note": "Newest reveal group: **October 11 to 12** ...",
    "list": [ ITEM, ... ],
    "skippedTitle": "Skipped: 23 more cosmetic or low-level items",
    "skipped": ["one line per group of skipped items"]
  },

  "sim": {
    "pr": { "state": "open|draft|none|blocked", "number": 12, "url": "https://github.com/bayov/forever-sim/pull/12", "title": "..." },
    "reason": "why there's no PR, why it's a draft, or a one-line summary",
    "changes": [ { "what": "...", "ref": "i:18671", "file": "tools/database/forever_items.go", "old": "...", "new": "..." } ],
    "impact": [ { "spec": "Shaman Enhancement L60", "metric": "DPS", "before": 705.31, "after": 712.04 } ],
    "impactNote": "...",
    "notIncluded": ["each 'needs your call' item"],
    "tasks": [ { "text": "TASKS.md item", "sub": "detail", "done": false } ]
  },

  "misc": [ { "cat": "Systems", "status": "blue", "date": "Oct 12", "source": "Wowhead", "url": "...", "text": "...", "details": ["..."], "repeat": false } ],
  "sources": [ { "site": "Wowhead", "label": "what it is", "url": "..." } ]
}
```

`CHANGE`, one class change (an ability, talent, racial, rune or mechanic, never an item):

```jsonc
{ "name": "Stormstrike", "ref": "s:17364",          // ref gives the icon and the tooltip; name is what the row shows
  "old": "20%", "new": "25%", "dir": "up",          // value pill; dir up/down/none (default from kind)
  "value": "text pill when there's no old/new",
  "pills": ["extra pill", "..."],
  "kind": "buff|nerf|change|fix|new",
  "status": "official|hotfix|blue|interview|client|datamined|tracker|tooltip|unclear",
  "date": "Oct 13", "repeat": false,                // repeat: "may repeat yesterday"
  "take": "one line of min-max impact, **bold** allowed" }
```

`ITEM`, one newly revealed or changed item:

```jsonc
{ "id": 18671,                                      // required; the tooltip comes from Wowhead Forever
  "who": ["Shaman", "Paladin", "Rogue", "Ele Shaman"],   // required; class names, optionally with a spec in front
  "kind": "new|buff|nerf|change",                   // default: new, or change when "changed" is set
  "status": "client",                               // default client
  "changed": [["Fire proc", "35", "90"]],          // Classic → Forever differences, shown in the tooltip
  "group": "Oct 11–12", "source": "Molten Core", "set": "...", "inSim": false,   // true | false | "added"
  "repeat": false,
  "take": "one line: who wants it and why",
  // only when Wowhead has no tooltip yet (build.py says so): "name", "quality" (0-5), "icon", "lines": [["t","+12 Strength"],["g","Equip: ..."]]
}
```

Status labels: `official` (patch notes), `hotfix` and `blue` are green. `interview` is green too: it's official, but it
isn't patch notes. `client` (client data, revealed items) is blue. `datamined` and `tracker` (the testers' bug tracker)
are amber. `tooltip` and `unclear` are gray.
