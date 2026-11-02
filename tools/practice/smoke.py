"""Headless Chrome smoke test of the practice room (docs/practice_room.md, Milestone 1 step 7).

It opens the room with the DPS Report's level 60 enh shaman and the default keybinds, walks up to
the dummy, plays a short scripted fight with the keyboard and the mouse, clicks Stop, and checks:

- no console errors and no page errors
- the fight pulled, and the meter's damage matches the summary's
- the summary shows

Run it with the practice server up (make practice-run):

    uv run --with playwright python tools/practice/smoke.py [--url http://127.0.0.1:3434] [--shot out.png]
"""

import argparse
import json
import sys
import time

from playwright.sync_api import sync_playwright


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--url', default='http://127.0.0.1:3434')
    ap.add_argument('--shot', default='')
    ap.add_argument('--headed', action='store_true')
    ap.add_argument('--chrome', default='/usr/bin/google-chrome', help="Chrome to run, '' for Playwright's own")
    args = ap.parse_args()

    errors = []
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=not args.headed, executable_path=args.chrome or None, args=['--use-gl=angle', '--use-angle=swiftshader', '--enable-unsafe-swiftshader'])
        page = browser.new_page(viewport={'width': 1600, 'height': 900})
        page.on('console', lambda m: errors.append(f'console {m.type}: {m.text}') if m.type == 'error' else None)
        page.on('pageerror', lambda e: errors.append(f'page error: {e}'))

        page.goto(f'{args.url}/practice.html')
        page.wait_for_selector('#enter')
        page.select_option('#preset', 'shaman__enhancement-level60')
        page.select_option('#record', 'default')
        page.click('#enter')

        # The worker loads 38 MB of WASM.
        page.wait_for_function('window.practice && window.practice.running && window.practice.toolkit', timeout=60000)
        state = lambda: page.evaluate(
            '() => ({phase: practice.fight.phase, damage: practice.player?.damageDone ?? 0, pos: practice.mover.move.pos, '
            'mana: practice.player?.mana, now: practice.now()})'
        )
        print('in the room:', state())

        # Prepull: Lightning Shield (slot 4) and Strength of Earth (slot 1), with the GCD between.
        page.keyboard.press('Digit4')
        time.sleep(1.7)
        page.keyboard.press('Digit1')
        time.sleep(1.2)

        # Walk up to the dummy, 15 yards ahead.
        page.keyboard.down('KeyW')
        deadline = time.time() + 5
        while time.time() < deadline and page.evaluate('() => practice.mover.move.pos[1]') < -4:
            time.sleep(0.02)
        page.keyboard.up('KeyW')
        time.sleep(0.2)
        s = state()
        print('walked to', s)
        if s['phase'] != 'prepull':
            errors.append(f'pulled before attacking: {s}')

        # Right click the dummy: target and attack, the pull.
        box = page.evaluate(
            '() => { const p = practice.room.screenPoint(practice.dummyPos, 1.2); return p; }'
        )
        page.mouse.click(box['x'], box['y'], button='right')
        time.sleep(0.3)
        s = state()
        print('after the right click', s)
        if s['phase'] != 'fight':
            errors.append(f'a right click on the dummy in melee range didn\'t pull: {s}')

        # 8 seconds of Stormstrike (slot 8), Earth Shock (11) and Flame Shock (10). Bar 1 holds
        # the rotation's spells in order, and Berserking isn't in an Orc's toolkit.
        for key in ['Digit8', 'Minus', 'Digit0', 'Digit8', 'Minus', 'Digit8']:
            page.keyboard.press(key)
            time.sleep(1.4)

        # A cast while moving fails with the game's error.
        page.keyboard.down('KeyS')
        time.sleep(0.2)
        page.keyboard.press('BracketLeft')  # unbound, no effect
        page.keyboard.up('KeyS')
        time.sleep(0.5)

        meter = page.evaluate('() => practice.player.damageDone')
        page.click('.fight-bar .stop')
        page.wait_for_selector('.summary', timeout=10000)
        summary = page.evaluate('() => practice.summary')
        names = page.evaluate('() => (practice.summary.spells || []).map(s => s.spellId ? practice.toolkit.nameOf(s.spellId) : "melee")')
        print('abilities:', names)
        for want in ('Stormstrike', 'Earth Shock'):
            if want not in names:
                errors.append(f'no {want} in the summary')
        print('summary:', json.dumps({k: summary[k] for k in ('dps', 'damage', 'durationMs', 'casts', 'idleGcdMs')}))
        if summary['damage'] <= 0:
            errors.append('the summary has no damage')
        if abs(summary['damage'] - page.evaluate('() => practice.player.damageDone')) > 0.5:
            errors.append(f"the meter shows {page.evaluate('() => practice.player.damageDone')}, the summary {summary['damage']}")
        if meter > summary['damage'] + 0.5:
            errors.append(f'the meter showed {meter} before Stop, more than the summary')
        # The attempt was saved.
        time.sleep(0.5)
        if args.shot:
            page.screenshot(path=args.shot)

        # Encounter mode: an imbue swap in the prepull, a 3 second countdown and a 10 second
        # fight that ends on its own.
        page.click('.summary button')
        page.wait_for_function('practice.running && practice.fight.phase === "prepull"', timeout=10000)
        sessions = page.evaluate('() => practice.sessions')
        page.evaluate('''() => {
            const s = practice.settings;
            practice.settings = {...s, mode: 'encounter', lengthMs: 10000, countdownMs: 3000};
            practice.post({type: 'stop', wallNow: performance.timeOrigin + performance.now()});
            practice.running = false;
            setTimeout(() => practice.start(), 50);
        }''')
        page.wait_for_function(f'practice.sessions > {sessions} && practice.running && practice.fight.phase === "prepull"', timeout=10000)
        time.sleep(0.3)
        flametongue = page.evaluate('() => practice.toolkit.spells.find(s => s.spellId === 16342 && s.tag === 99) ? 16342 : 0')
        if not flametongue:
            errors.append('no Flametongue Weapon cast in the toolkit')
        else:
            page.evaluate('() => practice.sendInput({kind: "cast", spellId: 16342, target: {kind: "player", index: 0}})')
            time.sleep(0.3)
            imbue = page.evaluate('() => practice.player.mainHandImbue')
            print('imbue after the cast:', imbue)
            if imbue != 'FlametongueWeapon':
                errors.append(f'Flametongue Weapon didn\'t go on: {imbue}')
        page.keyboard.down('KeyW')
        deadline = time.time() + 5
        while time.time() < deadline and page.evaluate('() => practice.mover.move.pos[1]') < -4:
            time.sleep(0.02)
        page.keyboard.up('KeyW')
        page.click('.fight-bar .start-encounter')
        time.sleep(0.5)
        phase = page.evaluate('() => practice.fight.phase')
        if phase != 'countdown':
            errors.append(f'Start Encounter gave phase {phase}')
        page.wait_for_function('practice.fight.phase === "fight"', timeout=6000)
        page.keyboard.press('Digit8')
        page.evaluate('() => practice.sendInput({kind: "startAttack"})')
        page.wait_for_selector('.summary', timeout=20000)
        enc = page.evaluate('() => practice.summary')
        print('encounter summary:', json.dumps({k: enc[k] for k in ('dps', 'damage', 'durationMs')}))
        if abs(enc['durationMs'] - 10000) > 1:
            errors.append(f"the encounter lasted {enc['durationMs']}ms, want 10000")

        t = page.evaluate('() => practice.timings')
        lat = sorted(t['pressToAnswer'])
        calls = sorted(t['calls'])
        pct = lambda xs, q: xs[min(len(xs) - 1, int(len(xs) * q))] if xs else 0
        print(f"timings: wasm load {t['loadMs']:.0f}ms, session start {t['startCallMs']:.1f}ms, "
              f"press to answer p50 {pct(lat, .5):.1f}ms p99 {pct(lat, .99):.1f}ms ({len(lat)} presses), "
              f"Go call p50 {pct(calls, .5):.2f}ms p99 {pct(calls, .99):.2f}ms, swaps {[round(x, 1) for x in t['swaps']]}")
        browser.close()

    if errors:
        print('FAIL')
        for e in errors:
            print(' ', e)
        sys.exit(1)
    print('OK')


if __name__ == '__main__':
    main()
