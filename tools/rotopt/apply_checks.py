#!/usr/bin/env python3
"""Flip the doubtful items of phase0_content.json from what the game showed us.

    python3 tools/rotopt/apply_checks.py ~/forever-logs/exports/export-<date>.json
    python3 tools/rotopt/apply_checks.py <export.json> --write

The Phase 0 BiS list marks an item "doubtful" when we don't know if its source
works in Forever, and its note says why: "Defilers Honored · Rutherford Twing ·
Arathi Basin in Forever unconfirmed". BayovCore writes down what the game shows
as we play (Checks.lua), and `/bayov export` hands it over as its `checks`
section. This script reads that section and prints each item whose avail
should flip, with the evidence. With --write it saves them: the avail goes
(the item is open), the note stays, and a dated part joins it ("AB queued
Nov 20").

The rules read the notes as they are:

- "<Faction> <Standing>" (a whole note part, like "Watchers Honored"): we've
  reached that standing, so a doubtful or infeasible item is open.
- "Arathi Basin in Forever unconfirmed", "AV queue unconfirmed" and "no rep
  source found yet": we saw the battleground offered, queued or entered, or
  the note's faction gained reputation. A doubtful item is open. An
  infeasible one stays, since its Exalted grind is still too long.
- "Forever card drops unconfirmed": we got any Darkmoon card, so cards drop.
- "Quest: <name>" first in the note: Forever has that quest, because we took
  it or turned it in.

Items marked "out" never flip, since those are a call about the raids. The
Reals we hold are only printed, since no item waits on them.
"""
import argparse
import datetime
import json
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CONTENT = os.path.join(HERE, 'phase0_content.json')

STANDINGS = {'Hated': 1, 'Hostile': 2, 'Unfriendly': 3, 'Neutral': 4, 'Friendly': 5, 'Honored': 6,
             'Revered': 7, 'Exalted': 8}
REP_PART = re.compile(r'^(.+?) (Friendly|Honored|Revered|Exalted)$')

# A note's faction word, where the game's faction name doesn't contain it.
FACTION_NAMES = {'Warsong Gulch': 'Warsong Outriders', 'Frostwolf': 'Frostwolf Clan', 'Timbermaw': 'Timbermaw Hold'}

# A note part that doubts a battleground, and the battleground's name in game.
BG_DOUBTS = {'Arathi Basin in Forever unconfirmed': ('Arathi Basin', 'AB'),
             'AV queue unconfirmed': ('Alterac Valley', 'AV')}


def short_date(day):
    """'2026-11-20' as 'Nov 20', the way the notes date things."""
    try:
        d = datetime.date.fromisoformat(day)
    except (TypeError, ValueError):
        return str(day)
    return f'{d:%b} {d.day}'


def faction(reps, word):
    """The export's faction for a note's faction word, or None."""
    want = FACTION_NAMES.get(word, word).lower()
    for name, rep in reps.items():
        if want in name.lower():
            return name, rep
    return None


def rep_gained(rep):
    """Whether a faction's reputation went up since we first saw it."""
    return rep.get('v', 0) > rep.get('v0', 0)


def bg_seen(bgs, name):
    """How and when we saw a battleground open: ('queued', '2026-11-20'), or None."""
    b = bgs.get(name)
    if not b:
        return None
    for how in ('inside', 'queued'):
        if b.get(how):
            return how, b[how]
    if b.get('listed') and b.get('canEnter'):
        return 'listed', b['listed']
    return None


def quest_seen(quests, name):
    """When we took or turned in a quest by this name, or None."""
    for q in quests.values():
        if (q.get('name') or '').lower() == name.lower():
            return q.get('done') or q.get('took')
    return None


def flips(content, checks):
    """The items whose avail should flip: [(id, item, evidence, note part)], by item ID."""
    reps = checks.get('reps') or {}
    bgs = checks.get('bgs') or {}
    cards = checks.get('cards') or {}
    quests = checks.get('quests') or {}
    first_card = min(cards.items(), key=lambda kv: kv[1].get('d') or '') if cards else None
    out = []
    for iid, item in content.items():
        avail = item.get('avail')
        if avail not in ('doubtful', 'infeasible'):
            continue
        parts = [p.strip() for p in (item.get('note') or '').split('·')]
        found = None
        # A standing we've reached settles the item, whatever else the note doubts.
        for p in parts:
            m = REP_PART.match(p)
            got = m and faction(reps, m.group(1))
            if got and got[1].get('top', got[1].get('r', 0)) >= STANDINGS[m.group(2)]:
                day = got[1].get('d') or got[1].get('d0')
                found = (f'{got[0]} {m.group(2)} reached ({day})', f'{m.group(2)} reached {short_date(day)}')
                break
        if not found and avail == 'doubtful':
            rep_word = next((REP_PART.match(p).group(1) for p in parts if REP_PART.match(p)), None)
            rep = rep_word and faction(reps, rep_word)
            for p in parts:
                if p in BG_DOUBTS:
                    name, abbr = BG_DOUBTS[p]
                    seen = bg_seen(bgs, name)
                    if seen:
                        found = (f'{name} {seen[0]} ({seen[1]})', f'{abbr} {seen[0]} {short_date(seen[1])}')
                    elif rep and rep_gained(rep[1]):
                        day = rep[1].get('d')
                        found = (f'{rep[0]} reputation gained ({day})', f'{rep_word} rep gained {short_date(day)}')
                elif p == 'no rep source found yet' and rep and rep_gained(rep[1]):
                    day = rep[1].get('d')
                    found = (f'{rep[0]} reputation gained ({day})', f'{rep_word} rep gained {short_date(day)}')
                elif p == 'Forever card drops unconfirmed' and first_card:
                    cid, card = first_card
                    what = card.get('name') or f'item {cid}'
                    found = (f'Darkmoon card {what} looted ({card.get("d")})', f'card looted {short_date(card.get("d"))}')
                if found:
                    break
            if not found and parts[0].startswith('Quest: '):
                day = quest_seen(quests, parts[0][len('Quest: '):])
                if day:
                    found = (f'{parts[0]} seen in game ({day})', f'quest seen {short_date(day)}')
        if found:
            out.append((iid, item, found[0], found[1]))
    return out


def apply(content, found):
    """Opens each found item: its avail goes and the dated part joins its note."""
    for iid, _, _, part in found:
        item = content[iid]
        item.pop('avail', None)
        item['note'] = f"{item['note']} · {part}" if item.get('note') else part


def dump(content):
    """phase0_content.json's own layout: one item a line, in the file's order."""
    lines = [f'  {json.dumps(k)}: {json.dumps(v, ensure_ascii=False)}' for k, v in content.items()]
    return '{\n' + ',\n'.join(lines) + '\n}\n'


def reals_line(checks):
    r = checks.get('reals')
    if not r:
        return None
    return f"Reals: {r.get('n')} held, {r.get('top')} at most · first seen {r.get('d0')}"


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split('\n')[0])
    ap.add_argument('export', help='a /bayov export JSON file')
    ap.add_argument('--write', action='store_true', help='apply the flips to phase0_content.json')
    ap.add_argument('--content', default=CONTENT, help=argparse.SUPPRESS)
    args = ap.parse_args(argv)
    with open(args.export, encoding='utf-8') as f:
        export = json.load(f)
    checks = export.get('checks')
    if not checks:
        print(f'{args.export} has no checks section (export from a BayovCore with Checks.lua)')
        return 1
    with open(args.content, encoding='utf-8') as f:
        content = json.load(f)
    who = export.get('who') or {}
    print(f"Checks of {who.get('name', '?')} on {checks.get('date')} (since {checks.get('since')})")
    found = flips(content, checks)
    for iid, item, evidence, _ in found:
        print(f"  {iid} {item['name']}: {item['avail']} -> open · {evidence}")
    if not found:
        print('  nothing to flip')
    line = reals_line(checks)
    if line:
        print(line)
    if args.write and found:
        apply(content, found)
        with open(args.content, 'w', encoding='utf-8') as f:
            f.write(dump(content))
        print(f'Wrote {len(found)} items to {os.path.relpath(args.content)}. Rerun tools/rotopt/bis_phase0.sh to rank them.')
    return 0


if __name__ == '__main__':
    sys.exit(main())
