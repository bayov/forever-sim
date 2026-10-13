// node tips.js <harness prefix> <mode> <scheme>: taps one mark on every chart with a real touch tap; reports tooltip + highlight; screenshots each
const { chromium } = require('playwright');
const path = require('path');
const [, , hp = 'h', mode = 'str', scheme = 'light'] = process.argv;
const CASES = [
  ['rank-row', 'ranking', '#rank-chart .fv-row:nth-child(4)'],
  ['scatter-dot', 'ranking', '#scatter circle:nth-of-type(30)'],
  ['length-line', 'ranking', '#lenchart rect[data-tip]', [0.15, 0.3]],
  ['race-heat-cell', 'races', '#heat .hc:not(.na):nth-of-type(20)'],
  ['race-spread-bar', 'races', '#racespread .tr:nth-of-type(3)'],
  ['race-overview-row', 'races', '#race-overview .fv-orow:nth-of-type(3) .fv-otrack'],
  ['race-spec-segment', 'races', '#race-spec .fv-rrow:nth-of-type(2) .sg:nth-of-type(2)'],
  ['race-spec-base', 'races', '#race-spec .fv-rrow:nth-of-type(4) .sg'],
  ['race-best-tile', 'races', '#ro-tiles .fv-tile'],
  ['dps-spread-bin', 'spec', '#sp-dist rect[data-tip]:nth-of-type(40)'],
  ['ability-bar', 'spec', '#sp-abil .tr'],
  ['stat-weight-bar', 'spec', '#sp-weights .tr'],
  ['talent', 'spec', '#sp-tal .fv-tal:nth-of-type(3)'],
  ['timeline-cast', 'spec', '#tl-scroll rect[data-tt]:nth-of-type(12)'],
  ['timeline-aura-lane', 'spec', '#tl-scroll g[data-tt] rect'],
  ['aura-bar', 'spec', '#sp-auras .tr'],
  ['spec-race-bar', 'spec', '#sp-races g[data-tt] rect:nth-of-type(2)'],
  ['spec-length-chart', 'spec', '#sp-len rect[data-tip]', [0.5, 0.4]],
  ['compare-spread-bin', 'compare', '#cmp-dist rect[data-tip]:nth-of-type(50)'],
  ['compare-length', 'compare', '#cmp-len rect[data-tip]', [0.4, 0.3]],
  ['compare-weight', 'compare', '#cmp-w .tr'],
  ['compare-damage', 'compare', '#cmp-abil .tr'],
];
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.PW_CHROME || '/opt/pw-browsers/chromium-1194/chrome-linux/chrome', args: ['--allow-file-access-from-files'] });
  const ctx = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, colorScheme: scheme });
  const page = await ctx.newPage();
  const errs = [];
  page.on('pageerror', (e) => errs.push(e.message));
  await page.goto('file://' + path.join(__dirname, `${hp}_${mode}.html`));
  await page.waitForTimeout(600);
  const out = [];
  for (const [name, tab, sel, frac] of CASES) {
    await page.evaluate((t) => document.querySelector(`[data-tab="${t}"]`).click(), tab);
    await page.waitForTimeout(250);
    const h = await page.$(sel);
    if (!h) { out.push(`${name}: NO TARGET ${sel}`); continue; }
    await h.evaluate((n) => n.scrollIntoView({ block: 'center' }));
    await page.waitForTimeout(150);
    const bb = await h.boundingBox();
    const fx = frac ? frac[0] : 0.5, fy = frac ? frac[1] : 0.5;
    await page.touchscreen.tap(bb.x + bb.width * fx, bb.y + bb.height * fy);
    await page.waitForTimeout(250);
    const r = await page.evaluate(() => {
      const t = document.querySelector('.fv-tt') || document.querySelector('.fv-tip');
      if (!t) return null;
      const b = t.getBoundingClientRect();
      return { kind: t.className.split(' ')[0], text: t.innerText.replace(/\n+/g, ' | ').slice(0, 150), inView: b.left >= 0 && b.right <= innerWidth, top: Math.round(b.top), bottom: Math.round(b.bottom), hl: document.querySelectorAll('.fv .is-hl, .fv .tt-on').length };
    });
    out.push(`${name}: ${r ? `${r.kind} ${r.inView ? 'in view' : 'OUT OF VIEW'} hl=${r.hl} :: ${r.text}` : 'NO TOOLTIP'}`);
    // screenshot the region around the tooltip and the tapped mark
    const clipTop = Math.max(0, Math.min(bb.y, r ? r.top : bb.y) - 40);
    const clipBot = Math.min(844, Math.max(bb.y + bb.height, r ? r.bottom : 0) + 40);
    await page.screenshot({ path: path.join(__dirname, 'shots', `tip_${scheme}_${name}.png`), clip: { x: 0, y: clipTop, width: 390, height: Math.max(120, clipBot - clipTop) } });
    // tap elsewhere closes it
    await page.touchscreen.tap(5, 5);
    await page.waitForTimeout(120);
    const closed = await page.evaluate(() => !document.querySelector('.fv-tt') && !document.querySelector('.fv-tip') && !document.querySelector('.fv .is-hl, .fv .tt-on'));
    out[out.length - 1] += closed ? ' [closes]' : ' [STAYS OPEN]';
  }
  out.push('errors=' + errs.length + ' ' + errs.slice(0, 3).join(' | '));
  console.log(out.join('\n'));
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });
