// node specswitch.js <ref harness prefix (sync)> <test harness prefix (async)> <mode plain|str> [cards-json]
// Switches the Spec page through presets every way the page allows, then checks every spec card shows the
// current preset: its content must equal a fresh load of that preset (reference), and, where a card sets
// data-build, that attribute must name the current preset.
const { chromium } = require('playwright');
const path = require('path');
const [, , refHp = 'b', hp = 'ba', mode = 'plain', cardsArg] = process.argv;
const CARDS = cardsArg ? JSON.parse(cardsArg) : ['#sp-title', '#sp-meta', '#sp-summary', '#sp-tiles', '#sp-tal', '#sp-dist', '#sp-pct', '#sp-abil', '#sp-weights', '#sp-wtab',
  '#tl-samples', '#tl-labels', '#tl-scroll', '#tl-readout', '#tl-legend', '#tl-log', '#sp-auras', '#sp-races', '#sp-len', '#sp-talents', '#sp-keytal', '#sp-rotation', '#sp-abiltab',
  '#sp-stats', '#sp-gear', '#sp-setup', '#sp-caveats'].concat(process.env.NEWCARDS ? ['#sp-prepull'] : ['#tl-opener']);
const FP = `(() => { const out = {}; for (const s of ${JSON.stringify(CARDS)}) { const n = document.querySelector(s); if (!n) { out[s] = 'MISSING'; continue; }
  const t = (n.textContent || '').replace(/\\s+/g, ' ').trim(); out[s] = { fp: t.length + ':' + n.querySelectorAll('*').length + ':' + t.slice(0, 80) + '…' + t.slice(-60), build: n.closest('[data-build]')?.dataset.build || null, skel: n.classList.contains('dash-skeleton') }; } return out; })()`;
const SEQ = [
  ['tab', 'spec'],
  ['pick', 'mage__fire'],
  ['next'],
  ['prev'],
  ['swipe', -1],
  ['pick', 'warlock__demonology'],
  ['rankrow', 'rogue__mutilate'],
  ['heat', '20'],
  ['pick', 'hunter__survival'],
  ['crest', 'Priest'],
  ['swipe', 1],
  ['pick', 'paladin__level_60_shock'],
  ['next'],
  ['pick', 'druid__feral-cat'],
  ['pick', 'shaman__enhancement-level60'],
  ['pick', 'warrior__fury'],
  ['rankrow', 'mage__frost'],
];
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.PW_CHROME || '/opt/pw-browsers/chromium-1194/chrome-linux/chrome', args: ['--allow-file-access-from-files'] });
  const mk = async () => {
    const ctx = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 1, isMobile: true, hasTouch: true });
    const page = await ctx.newPage();
    page.__errs = [];
    page.on('pageerror', (e) => page.__errs.push(e.message));
    return page;
  };
  const openAll = `document.querySelectorAll('.fv details').forEach(d=>d.open=true)`;
  // references: a fresh, fully loaded page per preset
  const refs = {};
  const refPage = await mk();
  const refOf = async (id) => {
    if (refs[id]) return refs[id];
    await refPage.goto('about:blank');
    await refPage.goto('file://' + path.join(__dirname, `${refHp}_${mode}.html`) + '#' + id);
    await refPage.waitForTimeout(500);
    await refPage.evaluate(openAll); await refPage.waitForTimeout(150);
    await refPage.evaluate(() => window.dispatchEvent(new Event('resize'))); await refPage.waitForTimeout(250);
    refs[id] = await refPage.evaluate(FP);
    return refs[id];
  };
  const page = await mk();
  await page.goto('file://' + path.join(__dirname, `${hp}_${mode}.html`));
  await page.waitForTimeout(300);
  const cur = () => page.evaluate(() => document.querySelector('#sp-pick').value);
  const report = [];
  let prev = null;
  for (const [k, a] of SEQ) {
    await page.evaluate(() => { const sv = document.querySelector('#tl-scroll svg'); if (sv && document.querySelector('#spec:not([hidden])')) { const r = sv.getBoundingClientRect(); sv.dispatchEvent(new MouseEvent('click', { bubbles: true, clientX: r.left + 60, clientY: r.top + 30 })); } });
    if (k === 'tab') await page.click(`[data-tab="${a}"]`);
    else if (k === 'pick') await page.evaluate((v) => { const s = document.querySelector('#sp-pick'); s.value = v; s.dispatchEvent(new Event('change')); }, a);
    else if (k === 'next' || k === 'prev') await page.evaluate((k) => document.querySelector('#sp-' + k).click(), k);
    else if (k === 'swipe') await page.evaluate((dir) => {
      const h = document.querySelector('#sp-head'); const r = h.getBoundingClientRect();
      const mkT = (x) => new Touch({ identifier: 1, target: h, clientX: x, clientY: r.top + 40 });
      h.dispatchEvent(new TouchEvent('touchstart', { touches: [mkT(200)], changedTouches: [mkT(200)], bubbles: true }));
      h.dispatchEvent(new TouchEvent('touchend', { touches: [], changedTouches: [mkT(200 + dir * 120)], bubbles: true }));
    }, a);
    else if (k === 'rankrow') {
      await page.click('[data-tab="ranking"]'); await page.waitForTimeout(300);
      await page.evaluate(() => document.querySelector('#rank-classes button[data-c="All"]')?.click());
      await page.waitForSelector(`#rank-chart .fv-row[data-id="${a}"]`, { timeout: 5000 });
      await page.evaluate((id) => { const r = document.querySelector(`#rank-chart .fv-row[data-id="${id}"]`); r.scrollIntoView({ block: 'center' }); r.click(); r.click(); }, a);
    } else if (k === 'heat') {
      await page.click('[data-tab="races"]'); await page.waitForSelector('#heat .hl', { timeout: 5000 }); await page.waitForTimeout(200);
      // a preset's row in the heat grid (any preset: the check follows whichever one opens): tap its first cell, then the tip's open button
      await page.evaluate((n) => { const all = document.querySelectorAll('#heat .hl'); const hl = all[n % all.length]; let c = hl.nextElementSibling; while (c && c.classList.contains('na')) c = c.nextElementSibling; c.scrollIntoView({ block: 'center' }); c.click(); }, +a);
      await page.waitForTimeout(150);
      await page.evaluate(() => document.querySelector('.fv-tip-open')?.click());
    } else if (k === 'crest') {
      await page.evaluate((c) => document.querySelector(`#hero [data-tt="c:${c}"]`).click(), a); await page.waitForTimeout(700);
      await page.evaluate(() => { const r = document.querySelector('#rank-chart .fv-row'); r.click(); r.click(); });
    }
    await page.waitForTimeout(60);
    await page.evaluate(openAll);
    const early = await page.evaluate(FP);
    await page.waitForTimeout(2200);
    await page.evaluate(openAll); await page.waitForTimeout(150);
    await page.evaluate(() => window.dispatchEvent(new Event('resize'))); await page.waitForTimeout(250);
    const late = await page.evaluate(FP);
    const id = await cur();
    const ref = await refOf(id), pref = prev ? await refOf(prev) : null;
    const stale = [], oldShown = [], wrongAttr = [];
    for (const s of CARDS) {
      if (late[s] === 'MISSING' || ref[s] === 'MISSING') { if (late[s] !== ref[s]) stale.push(s + ' missing'); continue; }
      if (late[s].fp !== ref[s].fp && !(s === '#tl-readout' && prev === id)) stale.push(s + (process.env.DBG ? ' GOT ' + late[s].fp + ' WANT ' + ref[s].fp : ''));
      if (late[s].build && late[s].build !== id) wrongAttr.push(`${s}=${late[s].build}`);
      if (pref && prev !== id && early[s] !== 'MISSING' && pref[s] !== 'MISSING' && early[s].fp === pref[s].fp && pref[s].fp !== ref[s].fp && !early[s].skel) oldShown.push(s);
    }
    const unstamped = CARDS.filter((c) => late[c] !== 'MISSING' && !late[c].build);
    report.push({ unstamped, step: `${k} ${a ?? ''}`.trim(), build: id, staleAfterLoad: stale, oldContentWhileLoading: oldShown, wrongDataBuild: wrongAttr });
    prev = id;
  }
  const bad = report.filter((r) => r.staleAfterLoad.length || r.wrongDataBuild.length || (process.env.NEWCARDS && r.unstamped.length));
  console.log(JSON.stringify({ mode, steps: report.length, presetsVisited: new Set(report.map((r) => r.build)).size, classesVisited: new Set(report.map((r) => String(r.build).split('__')[0])).size, failingSteps: bad.length, errors: page.__errs.concat(refPage.__errs).slice(0, 5), report }, null, 1));
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });
